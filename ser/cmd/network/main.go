// Command network generates multiple choice questions from the Ethernet and
// Wi-Fi standards tables and saves them to the questions table under the
// matching subject.
//
// Every question asks "which standard ...?" and describes the standard by its
// attributes, so the answer is always a standard name. A question is only
// generated when the attributes it names identify exactly one standard, and the
// distractors are only drawn from standards the question is definitely false
// for, so no question has two right answers.
//
// Usage (from the ser/ directory, so the default -env path finds .env):
//
//	go run ./cmd/network -dry-run          # print what would be saved
//	go run ./cmd/network                   # save both sets to the database
//	go run ./cmd/network -set wifi         # save the Wi-Fi questions only
package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

// item is one row of a standards table: the name that answers the question plus
// its attribute values, in the same order as the set's attributes.
type item struct {
	name   string
	values []string
}

// attribute is one column of a standards table that a question can describe.
type attribute struct {
	// phrase turns a value into a clause reading after the set's stem.
	phrase func(value string) string
	// satisfies reports whether a standard holding itemValue could also be a
	// correct answer to a question claiming claim. Equality when nil — set it
	// where a value can cover another, as a tri-band radio covers a question
	// about dual-band operation.
	satisfies func(itemValue, claim string) bool
	// tiebreak marks an attribute that is only worth asking about when the rest
	// leave two standards indistinguishable.
	tiebreak bool
}

// standardSet is a table plus the subject its questions are saved under.
type standardSet struct {
	key        string
	subject    string
	stem       string
	attributes []attribute
	items      []item
}

var ethernet = standardSet{
	key:     "ethernet",
	subject: "ethernet standards",
	stem:    "Which Ethernet standard",
	attributes: []attribute{
		{phrase: func(v string) string { return fmt.Sprintf("provides a bandwidth of %s", v) }},
		{phrase: func(v string) string { return fmt.Sprintf("runs over %s", v) }},
		{phrase: func(v string) string { return fmt.Sprintf("operates %s", v) }},
		{phrase: func(v string) string { return fmt.Sprintf("has a maximum segment length of %s", v) }, tiebreak: true},
	},
	// speed, medium, duplex, distance
	items: []item{
		{"10BASE5", []string{"10 Mbps", "thicknet coaxial cable", "half duplex only", "500 m"}},
		{"10BASE2", []string{"10 Mbps", "thinnet coaxial cable", "half duplex only", "185 m"}},
		{"10BASE-T", []string{"10 Mbps", "Cat3 or better twisted-pair copper", "half or full duplex", "100 m"}},
		{"100BASE-TX", []string{"100 Mbps", "Cat5 or better twisted-pair copper", "half or full duplex", "100 m"}},
		{"100BASE-FX", []string{"100 Mbps", "multimode fiber", "half or full duplex", "2 km"}},
		{"1000BASE-T", []string{"1 Gbps", "Cat5e or better twisted-pair copper", "full duplex only", "100 m"}},
		{"1000BASE-SX", []string{"1 Gbps", "multimode fiber", "full duplex only", "550 m"}},
		{"1000BASE-LX", []string{"1 Gbps", "single-mode fiber", "full duplex only", "5 km"}},
		{"10GBASE-T", []string{"10 Gbps", "Cat6a or better twisted-pair copper", "full duplex only", "100 m"}},
		{"10GBASE-SR", []string{"10 Gbps", "multimode fiber", "full duplex only", "400 m"}},
		{"10GBASE-LR", []string{"10 Gbps", "single-mode fiber", "full duplex only", "10 km"}},
		{"10GBASE-ER", []string{"10 Gbps", "single-mode fiber", "full duplex only", "40 km"}},
		{"40GBASE-SR4", []string{"40 Gbps", "multimode fiber", "full duplex only", "150 m"}},
		{"40GBASE-LR4", []string{"40 Gbps", "single-mode fiber", "full duplex only", "10 km"}},
		{"100GBASE-SR4", []string{"100 Gbps", "multimode fiber", "full duplex only", "100 m"}},
		{"100GBASE-LR4", []string{"100 Gbps", "single-mode fiber", "full duplex only", "10 km"}},
	},
}

var wifi = standardSet{
	key:     "wifi",
	subject: "wifi standards",
	stem:    "Which Wi-Fi standard",
	attributes: []attribute{
		{phrase: bandPhrase, satisfies: bandSatisfies},
		{phrase: func(v string) string { return fmt.Sprintf("has a maximum speed of %s", v) }},
		{phrase: func(v string) string { return fmt.Sprintf("has a typical range of %s", v) }},
		{phrase: func(v string) string { return fmt.Sprintf("was released in %s", v) }, tiebreak: true},
	},
	// bands (comma separated GHz), max speed, typical range, year
	items: []item{
		{"802.11a", []string{"5", "54 Mbps", "about 35 m", "1999"}},
		{"802.11b", []string{"2.4", "11 Mbps", "about 35 to 40 m", "1999"}},
		{"802.11g", []string{"2.4", "54 Mbps", "about 38 m", "2003"}},
		{"802.11n", []string{"2.4,5", "600 Mbps", "about 70 m", "2009"}},
		{"802.11ac", []string{"5", "3.5 Gbps", "about 35 m", "2013"}},
		{"802.11ax", []string{"2.4,5,6", "9.6 Gbps", "about 30 m", "2019"}},
		{"802.11be", []string{"2.4,5,6", "46 Gbps", "about 30 m", "2024"}},
	},
}

// maxVariants caps how many questions a single standard contributes.
const maxVariants = 3

type question struct {
	text   string
	a      string
	b      string
	c      string
	d      string
	answer string
}

func main() {
	dryRun := flag.Bool("dry-run", false, "print the questions instead of saving them")
	which := flag.String("set", "all", "which standards to generate: ethernet, wifi or all")
	envPath := flag.String("env", ".env", "path to the .env file holding the database credentials")
	seed := flag.Uint64("seed", 1, "seed for answer and distractor shuffling")
	flag.Parse()

	sets, err := selectSets(*which)
	if err != nil {
		log.Fatal(err)
	}

	rng := rand.New(rand.NewPCG(*seed, *seed))
	generated := make(map[string][]question, len(sets))
	for _, set := range sets {
		if err := set.validate(); err != nil {
			log.Fatalf("%s: %v", set.key, err)
		}
		generated[set.subject] = generateQuestions(set, rng)
	}

	if *dryRun {
		for _, set := range sets {
			questions := generated[set.subject]
			fmt.Printf("--- %s (%d questions) ---\n\n", set.subject, len(questions))
			for _, q := range questions {
				printQuestion(q)
			}
		}
		return
	}

	if err := save(context.Background(), *envPath, sets, generated); err != nil {
		log.Fatalf("saving questions: %v", err)
	}
}

func selectSets(which string) ([]standardSet, error) {
	switch which {
	case "all":
		return []standardSet{ethernet, wifi}, nil
	case ethernet.key:
		return []standardSet{ethernet}, nil
	case wifi.key:
		return []standardSet{wifi}, nil
	default:
		return nil, fmt.Errorf("unknown set %q: use ethernet, wifi or all", which)
	}
}

// validate catches a row whose values fell out of step with the attributes.
func (s standardSet) validate() error {
	for _, it := range s.items {
		if len(it.values) != len(s.attributes) {
			return fmt.Errorf("%s has %d values, want %d", it.name, len(it.values), len(s.attributes))
		}
	}
	if len(s.items) < 4 {
		return fmt.Errorf("need at least 4 standards to fill four choices, have %d", len(s.items))
	}
	return nil
}

func save(ctx context.Context, envPath string, sets []standardSet, generated map[string][]question) error {
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return fmt.Errorf("loading %s: %w", envPath, err)
		}
	}

	db.Init()
	queries := quiz.New(db.DB)

	for _, set := range sets {
		if err := saveSet(ctx, queries, set, generated[set.subject]); err != nil {
			return err
		}
	}
	return nil
}

func saveSet(ctx context.Context, queries *quiz.Queries, set standardSet, questions []question) error {
	subjectID, err := queries.GetSubjectIdByTitle(ctx, set.subject)
	if err != nil {
		return fmt.Errorf("looking up subject %q: %w", set.subject, err)
	}

	// The questions table has no unique constraint on the text, so skip anything
	// already saved to keep re-runs from piling up duplicates.
	existingRows, err := queries.GetQuestionsBySubjectId(ctx, pgtype.Int4{Int32: subjectID, Valid: true})
	if err != nil {
		return fmt.Errorf("fetching existing questions: %w", err)
	}
	existing := make(map[string]struct{}, len(existingRows))
	for _, text := range existingRows {
		existing[text] = struct{}{}
	}

	saved, skipped := 0, 0
	for _, q := range questions {
		if _, ok := existing[q.text]; ok {
			skipped++
			continue
		}

		_, err := queries.CreateSubjectQuestion(ctx, quiz.CreateSubjectQuestionParams{
			SubjectID: pgtype.Int4{Int32: subjectID, Valid: true},
			Question:  q.text,
			A:         q.a,
			B:         q.b,
			C:         q.c,
			D:         q.d,
			Answer:    q.answer,
		})
		if err != nil {
			return fmt.Errorf("saving %q: %w", q.text, err)
		}
		saved++
	}

	fmt.Printf("saved %d questions under subject %q (%d already present)\n", saved, set.subject, skipped)
	return nil
}

// generateQuestions builds every question whose attributes single out one
// standard, preferring the ones that name the fewest attributes.
func generateQuestions(set standardSet, rng *rand.Rand) []question {
	var questions []question

	for _, answer := range set.items {
		var used [][]int
		needsTiebreak := !identifiableWithoutTiebreak(set, answer)

		for _, combo := range combos(set) {
			if len(used) == maxVariants {
				break
			}
			if set.hasTiebreak(combo) && !needsTiebreak {
				continue
			}
			// A superset of a combo already used asks the same question with a
			// redundant clause bolted on, so it is not a new variant.
			if !identifies(set, combo, answer) || isSuperset(combo, used) {
				continue
			}
			q, ok := buildQuestion(set, combo, answer, rng)
			if !ok {
				continue
			}
			questions = append(questions, q)
			used = append(used, combo)
		}
	}

	return questions
}

// combos returns the attribute index sets a question can be built from, ordered
// so a question names no more attributes than it needs to be unambiguous and
// only falls back to a tiebreak attribute when the others cannot separate the
// standard from the rest of the table.
func combos(set standardSet) [][]int {
	var all [][]int
	for mask := 1; mask < 1<<len(set.attributes); mask++ {
		var combo []int
		for i := range set.attributes {
			if mask&(1<<i) != 0 {
				combo = append(combo, i)
			}
		}
		all = append(all, combo)
	}

	slices.SortStableFunc(all, func(a, b []int) int { return set.cost(a) - set.cost(b) })
	return all
}

func (s standardSet) cost(combo []int) int {
	if !s.hasTiebreak(combo) {
		return len(combo)
	}
	if len(combo) == 1 {
		// A tiebreak attribute on its own teaches none of the main columns.
		return len(combo) + 2*len(s.attributes)
	}
	return len(combo) + len(s.attributes)
}

func (s standardSet) hasTiebreak(combo []int) bool {
	return slices.ContainsFunc(combo, func(i int) bool { return s.attributes[i].tiebreak })
}

// identifiableWithoutTiebreak reports whether the main attributes are enough to
// tell the standard apart from every other row in the table.
func identifiableWithoutTiebreak(set standardSet, answer item) bool {
	var combo []int
	for i, attr := range set.attributes {
		if !attr.tiebreak {
			combo = append(combo, i)
		}
	}
	return identifies(set, combo, answer)
}

// identifies reports whether the combo's attribute values leave the answer as
// the only standard a question could be asking about.
func identifies(set standardSet, combo []int, answer item) bool {
	for _, it := range set.items {
		if it.name != answer.name && matches(set, combo, answer, it) {
			return false
		}
	}
	return true
}

// matches reports whether it also satisfies every claim the combo makes about
// answer, which is what makes it either an alternative right answer or an
// unusable distractor.
func matches(set standardSet, combo []int, answer item, it item) bool {
	for _, i := range combo {
		if !set.attributes[i].holds(it.values[i], answer.values[i]) {
			return false
		}
	}
	return true
}

func (a attribute) holds(itemValue, claim string) bool {
	if a.satisfies == nil {
		return itemValue == claim
	}
	return a.satisfies(itemValue, claim)
}

func buildQuestion(set standardSet, combo []int, answer item, rng *rand.Rand) (question, bool) {
	wrong := distractors(set, combo, answer, rng)
	if len(wrong) < 3 {
		return question{}, false
	}

	choices := append([]string{answer.name}, wrong[:3]...)
	correct := rng.IntN(len(choices))
	choices[0], choices[correct] = choices[correct], choices[0]

	return question{
		text:   questionText(set, combo, answer),
		a:      choices[0],
		b:      choices[1],
		c:      choices[2],
		d:      choices[3],
		answer: string(rune('a' + correct)),
	}, true
}

// isSuperset reports whether combo contains every index of any set in used.
func isSuperset(combo []int, used [][]int) bool {
	for _, u := range used {
		contained := true
		for _, i := range u {
			if !slices.Contains(combo, i) {
				contained = false
				break
			}
		}
		if contained {
			return true
		}
	}
	return false
}

func questionText(set standardSet, combo []int, answer item) string {
	clauses := make([]string, len(combo))
	for i, attr := range combo {
		clauses[i] = set.attributes[attr].phrase(answer.values[attr])
	}

	last := len(clauses) - 1
	joined := strings.Join(clauses[:last], ", ")
	if last > 0 {
		joined += " and "
	}
	joined += clauses[last]

	return fmt.Sprintf("%s %s?", set.stem, joined)
}

// distractors picks standards the question is false for, favouring the ones
// that still share some of its attributes so the wrong answers stay plausible.
func distractors(set standardSet, combo []int, answer item, rng *rand.Rand) []string {
	candidates := make([]item, 0, len(set.items)-1)
	for _, it := range set.items {
		if it.name != answer.name && !matches(set, combo, answer, it) {
			candidates = append(candidates, it)
		}
	}

	rng.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	slices.SortStableFunc(candidates, func(a, b item) int {
		return shared(set, combo, answer, b) - shared(set, combo, answer, a)
	})

	names := make([]string, len(candidates))
	for i, c := range candidates {
		names[i] = c.name
	}
	return names
}

// shared counts how many of the combo's claims it happens to meet.
func shared(set standardSet, combo []int, answer item, it item) int {
	count := 0
	for _, i := range combo {
		if set.attributes[i].holds(it.values[i], answer.values[i]) {
			count++
		}
	}
	return count
}

// bandPhrase reads a comma separated band list as a clause.
func bandPhrase(value string) string {
	bands := strings.Split(value, ",")
	for i, b := range bands {
		bands[i] = b + " GHz"
	}

	last := len(bands) - 1
	joined := strings.Join(bands[:last], ", ")
	if last > 0 {
		joined += " and "
	}
	joined += bands[last]

	if len(bands) == 1 {
		return fmt.Sprintf("operates on the %s band", joined)
	}
	return fmt.Sprintf("operates on the %s bands", joined)
}

// bandSatisfies reports whether a radio covering itemValue also answers a
// question about claim: 802.11ax covers 2.4 GHz, so it cannot be a distractor
// for a question about a 2.4 GHz standard.
func bandSatisfies(itemValue, claim string) bool {
	have := strings.Split(itemValue, ",")
	for _, want := range strings.Split(claim, ",") {
		if !slices.Contains(have, want) {
			return false
		}
	}
	return true
}

func printQuestion(q question) {
	fmt.Printf("%s\n  a) %s\n  b) %s\n  c) %s\n  d) %s\n  answer: %s\n\n", q.text, q.a, q.b, q.c, q.d, q.answer)
}
