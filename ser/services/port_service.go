package services

import (
	"cmp"
	"context"
	"dorzi/server/db/quiz"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"math/rand"
	"slices"
)

type PortService struct {
	queries *quiz.Queries
}

func (s *PortService) UpdatePorts(ctx context.Context, ports []MatchingRequest) error {
	for _, port := range ports {
		updateFunc := s.queries.UpdatePortScoreRight
		if !port.Correct {
			updateFunc = s.queries.UpdatePortScoreWrong
		}

		err := updateFunc(ctx, pgtype.Int4{Int32: int32(port.PortID), Valid: true})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PortService) generateMatching(ctx context.Context) error {
	const MaxQuestions = 10

	portRow, err := s.queries.GetPorts(ctx)
	if err != nil {
		return err
	}

	var newPorts []quiz.GetPortsRow
	var reviewedPorts []quiz.GetPortsRow

	for _, port := range portRow {
		if port.PortID.Int32 == 0 {
			newPorts = append(newPorts, port)
		} else {
			reviewedPorts = append(reviewedPorts, port)
		}
	}

	questionSize := min(len(portRow), MaxQuestions)
	var quizIds []int32

	slices.SortFunc(reviewedPorts, func(a, b quiz.GetPortsRow) int {
		if n := cmp.Compare(a.Streak.Int32, b.Streak.Int32); n != 0 {
			return n
		}
		aRatio := float64(a.RightCount.Int32) / max(float64(a.WrongCount.Int32), 1)
		bRatio := float64(b.RightCount.Int32) / max(float64(b.WrongCount.Int32), 1)
		return cmp.Compare(aRatio, bRatio)
	})

	rand.Shuffle(len(newPorts), func(i, j int) {
		newPorts[i], newPorts[j] = newPorts[j], newPorts[i]
	})

	if len(reviewedPorts) == 0 {

		for i := range questionSize {
			quizIds = append(quizIds, newPorts[i].ID)
		}
	} else if len(newPorts) == 0 {
		for i := range questionSize {
			quizIds = append(quizIds, reviewedPorts[i].ID)
		}
	} else {
		numReviewPorts := min(len(reviewedPorts), 5)
		wrongCount := 0
		for i := range numReviewPorts {
			if reviewedPorts[i].WrongCount.Int32 > reviewedPorts[i].RightCount.Int32 {
				wrongCount += 1
			}
		}

		numQuestions := min(len(reviewedPorts), wrongCount)

		for i := range numQuestions {
			quizIds = append(quizIds, reviewedPorts[i].ID)
		}

		numNewQuestions := min(len(newPorts), MaxQuestions-numQuestions)

		for i := range numNewQuestions {
			quizIds = append(quizIds, newPorts[i].ID)
		}
	}

	return s.queries.CreateMatchingQuiz(ctx, quizIds)

}

func (s *PortService) CreatePorts(ctx context.Context, ports []Port) error {
	params := PortAsCreatePortsParam(ports)
	return s.queries.CreatePorts(ctx, params)
}

func (s *PortService) GetMatchingQuiz(ctx context.Context) ([]quiz.GetMatchingQuizRow, error) {
	avg, err := s.queries.GetAverageQuizStreak(ctx)

	if err != nil {
		return nil, err
	}

	if avg >= 3 {
		fmt.Print("creating new quiz \n")
		err := s.queries.DeleteMatchingQuiz(ctx)
		if err != nil {
			return nil, err
		}
		err = s.generateMatching(ctx)
	}

	return s.queries.GetMatchingQuiz(ctx)
}
