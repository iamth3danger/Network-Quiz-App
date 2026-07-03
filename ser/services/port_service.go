package services

import (
	"context"
	"dorzi/server/db/quiz"
	"github.com/jackc/pgx/v5/pgtype"
)

type Port struct {
	ID           int32  `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Port         int32  `json:"port"`
	Protocol     string `json:"protocol"`
}

func (p *Port) getId() int32 {
	return p.ID
}

func portFromDB(portDb quiz.Port) Port {
	return Port{
		ID:           portDb.ID,
		Abbreviation: portDb.Abbreviation,
		Port:         portDb.Port,
		Protocol:     portDb.Protocol,
	}
}

type PortScore struct {
	ID         int32 `json:"id"`
	PortID     int32 `json:"port_id"`
	RightCount int32 `json:"right_count"`
	WrongCount int32 `json:"wrong_count"`
	Streak     int32 `json:"streak"`
}

func (s *PortScore) getId() int32 {
	return s.ID
}

func (s *PortScore) getRight() int32 {
	return s.RightCount
}

func (s *PortScore) getWrong() int32 {
	return s.WrongCount
}

func (s *PortScore) getStreak() int32 {
	return s.Streak
}

func portScoreFromDB(portScoreDb quiz.PortScore) PortScore {
	return PortScore{
		ID:         portScoreDb.ID,
		PortID:     portScoreDb.PortID.Int32,
		RightCount: portScoreDb.RightCount.Int32,
		WrongCount: portScoreDb.WrongCount.Int32,
		Streak:     portScoreDb.Streak.Int32,
	}
}

type MatchingQuiz struct {
	ID     int32 `json:"id"`
	PortID int32 `json:"port_id"`
}

func matchingFromDB(matchingDb quiz.MatchingQuiz) MatchingQuiz {
	return MatchingQuiz{
		ID:     matchingDb.ID,
		PortID: matchingDb.PortID.Int32,
	}
}

type PortRead struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
}

type PortService struct {
	queries *quiz.Queries
}

func NewPortService(q *quiz.Queries) PortService {
	return PortService{queries: q}
}

func (s *PortService) SaveMatchingQuiz(ctx context.Context, ports []Port) error {
	var portIds []int32
	for _, p := range ports {
		portIds = append(portIds, p.ID)
	}
	return s.queries.CreateMatchingQuiz(ctx, portIds)
}

func (s *PortService) DeleteMatchingQuiz(ctx context.Context) error {
	return s.queries.DeleteMatchingQuiz(ctx)
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

func (s *PortService) CreatePorts(ctx context.Context, ports []Port) error {
	params := PortAsCreatePortsParam(ports)
	return s.queries.CreatePorts(ctx, params)
}
