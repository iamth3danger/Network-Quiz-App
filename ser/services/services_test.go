package services

import (
	"reflect"
	"testing"
)

type mockScorer struct {
	id     int32
	right  int32
	wrong  int32
	streak int32
}

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
