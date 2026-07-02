package services

import (
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

type QuizInfo struct {
	Count int
}

type QuizRow struct {
	questionId  int32
	objectiveId int32
}

type PortRead struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
}

type ObjectiveRating struct {
	Code  string `json:"code"`
	Stars int    `json:"stars"`
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

type MatchingQuiz struct {
	ID     int32 `json:"id"`
	PortID int32 `json:"port_id"`
}

type Score struct {
	ID         int32 `json:"id"`
	QuestionID int32 `json:"question_id"`
	RightCount int32 `json:"right_count"`
	WrongCount int32 `json:"wrong_count"`
	Streak     int32 `json:"streak"`
}

type ObjectiveScore struct {
	ID          int32 `json:"id"`
	ObjectiveID int32 `json:"objective_id"`
	RightCount  int32 `json:"right_count"`
	WrongCount  int32 `json:"wrong_count"`
}

type Port struct {
	ID           int32  `json:"id"`
	Abbreviation string `json:"abbreviation"`
	Port         int32  `json:"port"`
	Protocol     string `json:"protocol"`
}

type PortScore struct {
	ID         int32 `json:"id"`
	PortID     int32 `json:"port_id"`
	RightCount int32 `json:"right_count"`
	WrongCount int32 `json:"wrong_count"`
	Streak     int32 `json:"streak"`
}

type Question struct {
	ID          int32  `json:"id"`
	ObjectiveID int32  `json:"objective_id"`
	Question    string `json:"question"`
	A           string `json:"a"`
	B           string `json:"b"`
	C           string `json:"c"`
	D           string `json:"d"`
	Answer      string `json:"answer"`
	TermID      int32  `json:"term_id"`
}

type Quiz struct {
	ID          int32  `json:"id"`
	QuestionID  int32  `json:"question_id"`
	ObjectiveID int32  `json:"objective_id"`
	Status      string `json:"status"`
}

type Term struct {
	ID          int32       `json:"id"`
	ObjectiveID pgtype.Int4 `json:"objective_id"`
	TermName    string      `json:"term_name"`
}

var ErrInvalidObjective = errors.New("invalid error objective")
