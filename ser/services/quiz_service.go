package services

import (
	"context"
	"dorzi/server/db/quiz"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Question struct {
	ID          int32  `json:"id"`
	ObjectiveID int32  `json:"objective_id"`
	SubjectID   int32  `json:"subject_id"`
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
	SubjectID   int32  `json:"subject_id"`
	Status      string `json:"status"`
}

func quizFromDB(quizDb quiz.Quiz) Quiz {
	return Quiz{
		ID:          quizDb.ID,
		QuestionID:  quizDb.QuestionID.Int32,
		ObjectiveID: quizDb.ObjectiveID.Int32,
		SubjectID:   quizDb.SubjectID.Int32,
		Status:      string(quizDb.Status.QuestionStatus),
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

func autoScoreFromDB(autoDb quiz.AutogenScore, genMap map[Generator]Scorer) {
	gen := Generator(autoDb.Question)
	score := Score{
		RightCount: autoDb.RightCount.Int32,
		WrongCount: autoDb.WrongCount.Int32,
		Streak:     autoDb.Streak.Int32,
	}

	var scorer Scorer = &score

	genMap[gen] = Scorer(scorer)
}

type ObjectiveScore struct {
	ID          int32 `json:"id"`
	ObjectiveID int32 `json:"objective_id"`
	RightCount  int32 `json:"right_count"`
	WrongCount  int32 `json:"wrong_count"`
}

func objScoreFromDB(objScoreDb quiz.ObjectiveScore) ObjectiveScore {
	return ObjectiveScore{
		ID:          objScoreDb.ID,
		ObjectiveID: objScoreDb.ObjectiveID.Int32,
		RightCount:  objScoreDb.RightCount.Int32,
		WrongCount:  objScoreDb.WrongCount.Int32,
	}
}

type QuizService struct {
	queries *quiz.Queries
}

func NewQuizService(q *quiz.Queries) QuizService {
	return QuizService{
		queries: q,
	}
}

func (s *QuizService) DeleteQuiz(ctx context.Context, examId int32) error {
	return s.queries.DeleteQuizWithExamId(ctx, pgtype.Int4{Int32: examId, Valid: true})
}

func (s *QuizService) SetQuizStatusReviewing(ctx context.Context, questionId int32) error {
	_, err := s.queries.SetQuizStatusReviewing(ctx, pgtype.Int4{Int32: questionId, Valid: true})
	return err
}

func (s *QuizService) GetQuestionCount(ctx context.Context, examId int32) (int64, error) {
	return s.queries.GetQuestionCount(ctx, pgtype.Int4{Int32: examId, Valid: true})
}

func (s *QuizService) SaveExamObj(ctx context.Context, params ExamObject) error {
	examParams := quiz.CreateExamObjParams{
		Code:        params.Code,
		Description: params.Description,
	}

	err := s.queries.CreateExamObj(ctx, examParams)
	return err
}

func (s *QuizService) createAndUpdateScore(ctx context.Context, questionID int, correct bool) error {
	_, err := s.findOrCreateScore(ctx, questionID)
	if err != nil {
		return err
	}

	updateFn := s.queries.UpdateRightScore

	if !correct {
		updateFn = s.queries.UpdateWrongScore
	}

	_, err = updateFn(ctx, pgtype.Int4{Int32: int32(questionID), Valid: true})
	return err
}

func (s *QuizService) findOrCreateScore(ctx context.Context, qID int) (quiz.Score, error) {
	score, err := s.queries.GetScoreByQuestionId(ctx, pgtype.Int4{Int32: int32(qID), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.createScore(ctx, qID)
		} else {
			return quiz.Score{}, err
		}
	}
	return score, nil
}

func (s *QuizService) createScore(ctx context.Context, qID int) (quiz.Score, error) {
	params := quiz.CreateScoreParams{
		QuestionID: pgtype.Int4{Int32: int32(qID), Valid: true},
		RightCount: pgtype.Int4{Int32: 0, Valid: true},
		WrongCount: pgtype.Int4{Int32: 0, Valid: true},
		Streak:     pgtype.Int4{Int32: 0, Valid: true},
	}
	return s.queries.CreateScore(ctx, params)
}

func (s *QuizService) GetAllWrongQuestions(ctx context.Context) ([]quiz.Question, error) {
	return s.queries.GetAllWrongQuestions(ctx)
}

func (s *QuizService) GetWrongQuestionsByExam(ctx context.Context, examId int32) ([]quiz.Question, error) {
	return s.queries.GetWrongQuestionsByExam(ctx, pgtype.Int4{Int32: examId, Valid: true})
}

func (s *QuizService) createOrUpdateObjectiveScore(ctx context.Context, examID int, correct bool) (quiz.ObjectiveScore, error) {
	var oScore quiz.ObjectiveScore
	var err error

	updateFn := s.queries.UpdateObjectiveScoreRight
	if !correct {
		updateFn = s.queries.UpdateObjectiveScoreWrong
	}

	oScore, err = updateFn(ctx, pgtype.Int4{Int32: int32(examID), Valid: true})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = s.queries.CreateObjectiveScore(ctx, pgtype.Int4{Int32: int32(examID), Valid: true})
			if err != nil {
				return quiz.ObjectiveScore{}, err
			}
			return s.createOrUpdateObjectiveScore(ctx, examID, correct)
		} else {
			return quiz.ObjectiveScore{}, err
		}
	}
	return oScore, nil
}

func (s *QuizService) UpdateQuestion(ctx context.Context, question QuestionRequest) error {
	qID := question.QuestionID
	correct := question.Correct
	examID := question.ExamObjectiveID

	err := s.createAndUpdateScore(ctx, qID, correct)
	if err != nil {
		return fmt.Errorf("saving / updating score %w", err)
	}

	if examID != 0 {
		_, err = s.createOrUpdateObjectiveScore(ctx, examID, correct)

		if err != nil {
			return fmt.Errorf("error with objective score %w", err)
		}
	}

	_, err = s.queries.SetQuizStatusAnswered(ctx, pgtype.Int4{Int32: int32(qID), Valid: true})

	return err
}

func (s *QuizService) GetExamCodes(ctx context.Context) ([]string, error) {
	return s.queries.GetExamCodes(ctx)
}

func (s *QuizService) FetchTermStats(ctx context.Context) ([]quiz.TermStatsRow, error) {
	return s.queries.TermStats(ctx)
}

func (s *QuizService) GetQuestionSetByExamId(ctx context.Context, examId int32) ([]string, error) {
	return s.queries.GetQuestionSetByExamId(ctx, pgtype.Int4{Int32: examId, Valid: true})
}

func (s *QuizService) SaveQuizQuestions(ctx context.Context, reviewed Question, questions []Question) error {

	if reviewed.ObjectiveID != 0 {
		_, err := s.queries.CreateQuizQuestion(ctx, quiz.CreateQuizQuestionParams{
			QuestionID:  pgtype.Int4{Int32: reviewed.ID, Valid: true},
			ObjectiveID: pgtype.Int4{Int32: reviewed.ObjectiveID, Valid: true},
			Status:      quiz.NullQuestionStatus{QuestionStatus: quiz.QuestionStatusReviewing, Valid: true},
		})

		if err != nil {
			return fmt.Errorf("%w: %w", ErrDBFetch, err)
		}

		quizQuestion := makeBatchQuizQuestionExam(questions)
		return s.queries.BatchQuizQuestionExam(ctx, quizQuestion)

	} else {
		_, err := s.queries.CreateQuizQuestion(ctx, quiz.CreateQuizQuestionParams{
			QuestionID: pgtype.Int4{Int32: reviewed.ID, Valid: true},
			SubjectID:  pgtype.Int4{Int32: reviewed.SubjectID, Valid: true},
			Status:     quiz.NullQuestionStatus{QuestionStatus: quiz.QuestionStatusReviewing, Valid: true},
		})

		if err != nil {
			return fmt.Errorf("%w: %w", ErrDBFetch, err)
		}

		quizQuestion := makeBatchQuizQuestionSubject(questions)
		return s.queries.BatchQuizQuestionSubject(ctx, quizQuestion)
	}

}
