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

func TestCidrToSubnetQuestion(t *testing.T) {
	question := CidrToSubnetQuestion()
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
