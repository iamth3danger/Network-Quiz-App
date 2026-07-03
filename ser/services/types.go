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

type PortRead struct {
	Abbreviation string `json:"abbreviation"`
	Name         string `json:"name"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
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

type Score struct {
	ID         int32 `json:"id"`
	QuestionID int32 `json:"question_id"`
	RightCount int32 `json:"right_count"`
	WrongCount int32 `json:"wrong_count"`
	Streak     int32 `json:"streak"`
}

func (s *Score) getId() int32 {
	return s.ID
}

func (s *Score) getRight() int32 {
	return s.RightCount
}

func (s *Score) getWrong() int32 {
	return s.WrongCount
}

func (s *Score) getStreak() int32 {
	return s.Streak
}

func scoreFromDB(scoreDb quiz.Score) Score {
	return Score{
		ID:         scoreDb.ID,
		QuestionID: scoreDb.QuestionID.Int32,
		RightCount: scoreDb.RightCount.Int32,
		WrongCount: scoreDb.WrongCount.Int32,
		Streak:     scoreDb.Streak.Int32,
	}
}

type ObjectiveScore struct {
	ID          int32 `json:"id"`
	ObjectiveID int32 `json:"objective_id"`
	RightCount  int32 `json:"right_count"`
	WrongCount  int32 `json:"wrong_count"`
}

type ObjectiveStat struct {
	Code           float64
	Right          int32
	Wrong          int32
	Terms          int32
	AttemptedTerms int32
}

func objScoreFromDB(objScoreDb quiz.ObjectiveScore) ObjectiveScore {
	return ObjectiveScore{
		ID:          objScoreDb.ID,
		ObjectiveID: objScoreDb.ObjectiveID.Int32,
		RightCount:  objScoreDb.RightCount.Int32,
		WrongCount:  objScoreDb.WrongCount.Int32,
	}
}

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

func (q *Question) getId() int32 {
	return q.ID
}

func questionFromDB(questionDb quiz.Question) Question {
	return Question{
		ID:          questionDb.ID,
		ObjectiveID: questionDb.ObjectiveID.Int32,
		Question:    questionDb.Question,
		A:           questionDb.A,
		B:           questionDb.B,
		C:           questionDb.C,
		D:           questionDb.D,
		Answer:      questionDb.Answer,
		TermID:      questionDb.TermID.Int32,
	}
}

type Quiz struct {
	ID          int32  `json:"id"`
	QuestionID  int32  `json:"question_id"`
	ObjectiveID int32  `json:"objective_id"`
	Status      string `json:"status"`
}

func quizFromDB(quizDb quiz.Quiz) Quiz {
	return Quiz{
		ID:          quizDb.ID,
		QuestionID:  quizDb.QuestionID.Int32,
		ObjectiveID: quizDb.ObjectiveID.Int32,
		Status:      string(quizDb.Status.QuestionStatus),
	}
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
