package services

import (
	"dorzi/server/db/quiz"
	"errors"
)

type ExamObject struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type QuestionRequest struct {
	QuestionID      int  `json:"question_id"`
	Correct         bool `json:"correct"`
	ExamObjectiveID int  `json:"exam_objective_id"`
}

type MatchingRequest struct {
	PortID  int  `json:"port_id"`
	Correct bool `json:"correct"`
}

type QuizStatus struct {
	Count int
}

type QuizRow struct {
	questionId  int32
	objectiveId int32
}

type ObjectiveRating struct {
	Code  float64 `json:"code"`
	Stars int     `json:"stars"`
}

type StarRating struct {
	rating  int
	total   int
	correct float64
	term    float64
}

func (s *StarRating) isRating(stat Stat) bool {
	return stat.total >= s.total && stat.correct >= s.correct && stat.term >= s.term
}

type Stat struct {
	total   int
	correct float64
	term    float64
}

type ExamObjective struct {
	ID          int32  `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func examFromDB(examDb quiz.ExamObjective) ExamObjective {
	return ExamObjective{
		ID:          examDb.ID,
		Code:        examDb.Code,
		Description: examDb.Description,
	}
}

type ObjectiveStat struct {
	Code           float64
	Right          int32
	Wrong          int32
	Terms          int32
	AttemptedTerms int32
}

type Term struct {
	ID          int32  `json:"id"`
	ObjectiveID int32  `json:"objective_id"`
	TermName    string `json:"term_name"`
}

func termFromDB(termDb quiz.Term) Term {
	return Term{
		ID:          termDb.ID,
		ObjectiveID: termDb.ObjectiveID.Int32,
		TermName:    termDb.TermName,
	}
}

var ErrInvalidObjective = errors.New("invalid error objective")
var ErrDBFetch = errors.New("fetching from db")
