// Command json imports a JSON file of multiple choice questions into the
// questions table under a subject, creating the subject if it does not exist.
//
// The file is an array of objects. Fields beyond these are ignored, so files
// carrying extras such as exam_objective import fine:
//
//	[{"question":"...","a":"...","b":"...","c":"...","d":"...","answer":"b"}]
//
// Usage (from the ser/ directory, so the default -env path finds .env):
//
//	go run ./cmd/json questions.json "tcp flags"
//	go run ./cmd/json -dry-run questions.json "tcp flags"
package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

type question struct {
	Question string `json:"question"`
	A        string `json:"a"`
	B        string `json:"b"`
	C        string `json:"c"`
	D        string `json:"d"`
	Answer   string `json:"answer"`
}

func main() {
	dryRun := flag.Bool("dry-run", false, "validate the file and report what would be saved, without writing")
	envPath := flag.String("env", ".env", "path to the .env file holding the database credentials")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [flags] <file.json> <subject>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}
	path, subject := flag.Arg(0), strings.TrimSpace(flag.Arg(1))
	if subject == "" {
		log.Fatal("subject must not be empty")
	}

	questions, err := load(path)
	if err != nil {
		log.Fatalf("reading %s: %v", path, err)
	}

	if *dryRun {
		for _, q := range questions {
			fmt.Printf("%s\n  a) %s\n  b) %s\n  c) %s\n  d) %s\n  answer: %s\n\n", q.Question, q.A, q.B, q.C, q.D, q.Answer)
		}
		fmt.Printf("%d questions ready for subject %q\n", len(questions), subject)
		return
	}

	if err := save(context.Background(), *envPath, subject, questions); err != nil {
		log.Fatalf("saving questions: %v", err)
	}
}

// load reads the file and returns its questions, rejecting the whole file if any
// entry is unusable so an import is all or nothing.
func load(path string) ([]question, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var questions []question
	if err := json.Unmarshal(data, &questions); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}
	if len(questions) == 0 {
		return nil, errors.New("file contains no questions")
	}

	var problems []string
	seen := make(map[string]int, len(questions))
	cleaned := make([]question, 0, len(questions))

	for i, q := range questions {
		q, err := clean(q)
		if err != nil {
			problems = append(problems, fmt.Sprintf("entry %d: %v", i, err))
			continue
		}
		if first, ok := seen[q.Question]; ok {
			problems = append(problems, fmt.Sprintf("entry %d: duplicates entry %d: %q", i, first, q.Question))
			continue
		}
		seen[q.Question] = i
		cleaned = append(cleaned, q)
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("%d invalid entries:\n  %s", len(problems), strings.Join(problems, "\n  "))
	}
	return cleaned, nil
}

// clean trims a question and checks it can stand as a four choice item with a
// real answer letter, since the questions table takes any text it is given.
func clean(q question) (question, error) {
	q.Question = strings.TrimSpace(q.Question)
	q.A = strings.TrimSpace(q.A)
	q.B = strings.TrimSpace(q.B)
	q.C = strings.TrimSpace(q.C)
	q.D = strings.TrimSpace(q.D)
	q.Answer = strings.ToLower(strings.TrimSpace(q.Answer))

	fields := map[string]string{"question": q.Question, "a": q.A, "b": q.B, "c": q.C, "d": q.D}
	for _, name := range []string{"question", "a", "b", "c", "d"} {
		if fields[name] == "" {
			return q, fmt.Errorf("%s is empty", name)
		}
	}

	switch q.Answer {
	case "a", "b", "c", "d":
	case "":
		return q, errors.New("answer is empty")
	default:
		return q, fmt.Errorf("answer %q is not a, b, c or d", q.Answer)
	}

	return q, nil
}

func save(ctx context.Context, envPath string, subject string, questions []question) error {
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return fmt.Errorf("loading %s: %w", envPath, err)
		}
	}

	db.Init()
	queries := quiz.New(db.DB)

	subjectID, err := queries.CreateOrGetSubject(ctx, subject)
	if err != nil {
		return fmt.Errorf("creating subject %q: %w", subject, err)
	}

	// The questions table has no unique constraint on the text, so skip anything
	// already saved to keep re-imports from piling up duplicates.
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
		if _, ok := existing[q.Question]; ok {
			skipped++
			continue
		}

		_, err := queries.CreateSubjectQuestion(ctx, quiz.CreateSubjectQuestionParams{
			SubjectID: pgtype.Int4{Int32: subjectID, Valid: true},
			Question:  q.Question,
			A:         q.A,
			B:         q.B,
			C:         q.C,
			D:         q.D,
			Answer:    q.Answer,
		})
		if err != nil {
			return fmt.Errorf("saving %q: %w", q.Question, err)
		}
		saved++
	}

	fmt.Printf("saved %d questions under subject %q (id %d, %d already present)\n", saved, subject, subjectID, skipped)
	return nil
}
