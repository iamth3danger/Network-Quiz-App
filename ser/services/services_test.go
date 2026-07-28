package services

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"strconv"
	"testing"
)

type mockScorer struct {
	id     int32
	right  int32
	wrong  int32
	streak int32
}

type mockQuizer struct {
	id int32
}

func (m mockQuizer) getId() int32 { return m.id }

func (m mockScorer) getId() int32     { return m.id }
func (m mockScorer) getRight() int32  { return m.right }
func (m mockScorer) getWrong() int32  { return m.wrong }
func (m mockScorer) getStreak() int32 { return m.streak }

// pad fills up to MaxReviewSample with right-majority scorers (never counted)
func pad(lead []mockScorer) []Scorer {
	out := make([]Scorer, 0, MaxReviewSample)
	for _, s := range lead {
		out = append(out, s)
	}
	for i := len(lead); i < MaxReviewSample; i++ {
		out = append(out, mockScorer{id: int32(1000 + i), right: 1})
	}
	return out
}

func TestGetMatchingReviewableScores(t *testing.T) {
	tests := []struct {
		name    string
		scores  []Scorer
		wantIDs []int32
	}{
		{
			name:    "none wrong",
			scores:  pad([]mockScorer{{id: 1, right: 2}, {id: 2, right: 1}}),
			wantIDs: []int32{},
		},
		{
			name:    "equal wrong and right not counted",
			scores:  pad([]mockScorer{{id: 1, wrong: 2, right: 2}}),
			wantIDs: []int32{},
		},
		{
			name:    "leading wrong majority",
			scores:  pad([]mockScorer{{id: 1, wrong: 2}, {id: 2, wrong: 3, right: 1}}),
			wantIDs: []int32{1, 2},
		},
		{
			name:    "single wrong at front",
			scores:  pad([]mockScorer{{id: 1, wrong: 1}, {id: 2, right: 1}}),
			wantIDs: []int32{1},
		},
		{
			name: "all five wrong",
			scores: []Scorer{
				mockScorer{id: 1, wrong: 1}, mockScorer{id: 2, wrong: 1},
				mockScorer{id: 3, wrong: 1}, mockScorer{id: 4, wrong: 1},
				mockScorer{id: 5, wrong: 1},
			},
			wantIDs: []int32{1, 2, 3, 4, 5},
		},
		{
			// count=2 but the wrong ones are at idx 1,2 → returns [1,2] by position
			name:    "unsorted returns by position not by which are wrong",
			scores:  pad([]mockScorer{{id: 1, right: 1}, {id: 2, wrong: 1}, {id: 3, wrong: 1}}),
			wantIDs: []int32{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &QuizInfo{scores: tt.scores}
			got := q.getMatchingReviewableScores()

			gotIDs := make([]int32, len(got))
			for i, s := range got {
				gotIDs[i] = s.getId()
			}
			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Errorf("got ids %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}

}

func mockToQuizer(mocks []mockQuizer) []Quizer {
	var quizers []Quizer
	for _, m := range mocks {
		quizers = append(quizers, &m)
	}

	return quizers
}

func mockToScore(mocks []mockScorer) []Scorer {
	var scorers []Scorer
	for _, m := range mocks {
		scorers = append(scorers, &m)
	}

	return scorers
}

func TestMakeMatching(t *testing.T) {
	reviewMock := mockToQuizer([]mockQuizer{{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}})
	newMock := mockToQuizer([]mockQuizer{{id: 6}, {id: 7}, {id: 8}, {id: 9}, {id: 10}})
	scoreMock := []mockScorer{{id: 1}, {id: 2}, {id: 3}, {id: 4}, {id: 5}}
	for i := range scoreMock {
		scoreMock[i].wrong = 2
	}

	scoreQuizer := map[int32]Quizer{
		1: reviewMock[4],
		2: reviewMock[2],
		3: reviewMock[1],
		4: reviewMock[3],
		5: reviewMock[0],
	}

	item := QuizInfo{
		reviewedQuizer: reviewMock,
		newQuizer:      newMock,
		scores:         mockToScore(scoreMock),
		scoreQuizer:    scoreQuizer,
	}

	item.rearrangeArrs()
	matchingQuiz := item.makeMatching()

	assert.Equal(t, 10, len(matchingQuiz), "quiz length should be max length")

}

func getAnswerStr(question Question) (string, error) {
	switch question.Answer {
	case "a":
		return question.A, nil
	case "b":
		return question.B, nil
	case "c":
		return question.C, nil
	case "d":
		return question.D, nil
	default:
		return "", fmt.Errorf("Question in wrong format!")
	}
}

func extractIP(addr string) (string, error) {
	re := regexp.MustCompile(`(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})`)

	matches := re.FindAllString(addr, -1)
	fmt.Println(matches)

	if len(matches) < 1 {
		return "", fmt.Errorf("no address found in %s", addr)
	}

	return matches[0], nil
}

func TestExtractIP(t *testing.T) {

	str := "what is 192.55.255.0/25 as a network address?"

	ip, err := extractIP(str)

	if err != nil {
		t.Logf("%v\n", err)
	}

	assertSubnet("192.55.255.0", ip, t)
}

func extractCIDRPrefix(cidr string) (int, error) {
	re := regexp.MustCompile(`/(\d+)`)
	matches := re.FindStringSubmatch(cidr)

	if len(matches) < 2 {
		return 0, fmt.Errorf("no CIDR prefix found in %s", cidr)
	}

	prefix, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid prefix: %v", err)
	}

	return prefix, nil
}

func extractSubnetMask(s string) string {
	re := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func assertSubnet(expected string, actual string, t *testing.T) {
	if expected != actual {
		t.Fatalf("expected subnet %s actual subnet %s", expected, actual)
	}
}

func TestCidrToSubnet(t *testing.T) {
	assertSubnet("0.0.0.0", CidrToSubnet(0), t)
	assertSubnet("224.0.0.0", CidrToSubnet(3), t)
	assertSubnet("255.224.0.0", CidrToSubnet(11), t)
	assertSubnet("255.252.0.0", CidrToSubnet(14), t)
	assertSubnet("255.255.192.0", CidrToSubnet(18), t)
	assertSubnet("255.255.252.0", CidrToSubnet(22), t)
	assertSubnet("255.255.255.240", CidrToSubnet(28), t)
}

func TestIPConversionQuestion(t *testing.T) {
	question, err := iPCidrConversionQuestion(IPCidrToNetwork, "network")
	if err != nil {
		t.Fatalf("%v", err)
	}

	answerStr, err := getAnswerStr(question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	ipStr, err := extractIP(question.Question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	cidr, err := extractCIDRPrefix(question.Question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	ip, err := addressToBits(ipStr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	net, err := IPCidrToNetwork(ip, cidr)

	assertSubnet(net, answerStr, t)
}

func TestCidrToSubnetQuestion(t *testing.T) {
	question, err := CidrToSubnetQuestion()
	if err != nil {
		t.Fatalf("%v", err)
	}
	answerStr, err := getAnswerStr(question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	cidr, err := extractCIDRPrefix(question.Question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	subnet := CidrToSubnet(int32(cidr))

	if answerStr != subnet {
		t.Fatalf("answer is %s while computed subnet %s", answerStr, subnet)
	}
}

func TestSubnetToCidrQuestion(t *testing.T) {
	question := SubnetToCidrQuestion()
	answerStr, err := getAnswerStr(question)
	if err != nil {
		t.Fatalf("%v", err)
	}

	cidr, err := extractCIDRPrefix(answerStr)

	if err != nil {
		t.Fatalf("%v", err)
	}

	mask := extractSubnetMask(question.Question)
	if mask == "" {
		t.Fatalf("subnet mask not extracted!")
	}

	answerMask := CidrToSubnet(int32(cidr))

	if answerMask != mask {
		t.Fatalf("answer is %s while computed subnet %s", answerMask, mask)
	}
}

func TestIPCidrToNetwork(t *testing.T) {
	ip, err := addressToBits("192.68.44.37")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err := IPCidrToNetwork(ip, 28)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("192.68.44.32", network, t)

	ip, err = addressToBits("72.102.249.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 23)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("72.102.248.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 21)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.48.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 5)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("144.0.0.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 9)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.0.0.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 12)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.64.0.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 18)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.0.0", network, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	network, err = IPCidrToNetwork(ip, 24)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.48.0", network, t)

}

func TestIPCidrToBroadcast(t *testing.T) {
	ip, err := addressToBits("192.68.44.37")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err := IPCidrToBroadcast(ip, 28)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("192.68.44.47", broadcast, t)
	ip, err = addressToBits("72.102.249.1")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 23)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("72.102.249.255", broadcast, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 21)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.55.255", broadcast, t)

	ip, err = addressToBits("151.72.55.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 9)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.127.255.255", broadcast, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 12)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.79.255.255", broadcast, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 18)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.63.255", broadcast, t)

	ip, err = addressToBits("151.72.48.210")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	broadcast, err = IPCidrToBroadcast(ip, 24)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	assertSubnet("151.72.48.255", broadcast, t)

}

func checkUnique(cidrs []int32) bool {
	cidrSet := make(map[int32]struct{})

	for _, cidr := range cidrs {
		_, ok := cidrSet[cidr]
		if ok {
			return false
		}
		cidrSet[cidr] = struct{}{}
	}

	return true
}

func checkUniqueDotted(decimals []string) bool {
	decimealSet := make(map[string]struct{})

	for _, d := range decimals {
		_, ok := decimealSet[d]
		if ok {
			return false
		}
		decimealSet[d] = struct{}{}
	}

	return true
}

func TestGenOptions(t *testing.T) {
	for range 10000 {
		cidrs := generateCidrs(3)
		if !checkUnique(cidrs) {
			t.Fatalf("duplicate values!")
		}
	}
}

func TestGetConvChoices(t *testing.T) {
	for range 10000 {
		ch, err := getConvChoices(IPCidrToNetwork)
		if err != nil {
			t.Fatalf("%v\n", err)
		}
		if !checkUniqueDotted(ch.addresses) {
			t.Fatalf("duplicate values!")
		}

	}
}
