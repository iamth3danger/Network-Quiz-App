package services

import (
	"cmp"
	"context"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"math/rand"
	"slices"
	"strconv"
	"strings"
)

type QuizService struct {
	queries     *quiz.Queries
	examIdMap   map[float64]int
	termIdMap   map[string]int
	termExamMap map[int]int
}

func NewQuizService() *QuizService {
	return &QuizService{
		queries:     quiz.New(db.DB),
		examIdMap:   make(map[float64]int),
		termIdMap:   make(map[string]int),
		termExamMap: make(map[int]int),
	}
}

func (s *QuizService) Init(ctx context.Context) error {
	terms, err := s.queries.GetTerms(ctx)
	if err != nil {
		return err
	}

	s.fillTerms(terms)
	s.fillExamTermMap(terms)

	return s.fillMap(ctx)
}

func (s *QuizService) fillMap(ctx context.Context) error {
	examCodes, err := s.queries.GetExamIdsCodes(ctx)
	if err != nil {
		return err
	}
	for _, exam := range examCodes {
		code, err := strconv.ParseFloat(exam.Code, 64)
		if err != nil {
			return fmt.Errorf("issue parsing float %w", err)
		}
		s.examIdMap[code] = int(exam.ID)
	}
	return nil
}

func (s *QuizService) fillTerms(terms []quiz.Term) {
	for _, term := range terms {
		s.termIdMap[strings.ToLower(term.TermName)] = int(term.ID)
	}
}

func (s *QuizService) fillExamTermMap(terms []quiz.Term) {
	for _, term := range terms {
		s.termExamMap[int(term.ID)] = int(term.ObjectiveID.Int32)
	}
}

func (s *QuizService) DeleteQuiz(ctx context.Context, objective float64) error {
	examId, ok := s.examIdMap[objective]
	if !ok {
		return fmt.Errorf("%w: %v", ErrInvalidObjective, objective)
	}

	return s.queries.DeleteQuizWithExamId(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
}

func (s *QuizService) FetchQuestion(ctx context.Context, objective float64) (quiz.Question, error) {
	examId, ok := s.examIdMap[objective]
	if !ok {
		return quiz.Question{}, fmt.Errorf("invalid exam objective: %w", objective)
	}

	question, err := s.queries.FetchReviewQuestion(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s.FetchNewQuestion(ctx, examId)
		}

		return quiz.Question{}, fmt.Errorf("error fetching review question %w", err)
	}

	return quiz.Question{ID: question.ID, ObjectiveID: question.ObjectiveID, Question: question.Question, A: question.A, B: question.B, C: question.C, D: question.D, Answer: question.Answer, TermID: question.TermID}, nil
}

func (s *QuizService) FetchNewQuestion(ctx context.Context, examId int) (quiz.Question, error) {
	question, err := s.queries.GetRandomQuestion(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
	if err != nil {
		return quiz.Question{}, err
	}
	_, err = s.queries.SetQuizStatusReviewing(ctx, pgtype.Int4{Int32: question.ID, Valid: true})
	if err != nil {
		return quiz.Question{}, err
	}

	return quiz.Question{ID: question.ID, ObjectiveID: question.ObjectiveID, Question: question.Question, A: question.A, B: question.B, C: question.C, D: question.D, Answer: question.Answer, TermID: question.TermID}, nil
}

func (s *QuizService) SetQuizStatusReviewing(ctx context.Context, questionId int32) error {
	_, err := s.queries.SetQuizStatusReviewing(ctx, pgtype.Int4{Int32: questionId, Valid: true})
	return err
}

func (s *QuizService) GetQuizAnswered(ctx context.Context, objective float64) (QuizInfo, error) {
	examId := s.examIdMap[objective]
	count, err := s.queries.GetNumberAnswered(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
	if err != nil {
		return QuizInfo{}, err
	}
	return QuizInfo{Count: int(count) + 1}, err
}

func (s *QuizService) GetQuestionCount(ctx context.Context, objective float64) (int64, error) {
	examId, ok := s.examIdMap[objective]
	if !ok {
		return 0, fmt.Errorf("%w: %v", ErrInvalidObjective, objective)
	}
	return s.queries.GetQuestionCount(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
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

func (s *QuizService) GetWrongQuestionsByExam(ctx context.Context, objective float64) ([]quiz.Question, error) {
	examId := s.examIdMap[objective]
	return s.queries.GetWrongQuestionsByExam(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
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

func (s *QuizService) GenerateQuiz(ctx context.Context, objective float64) error {
	const MaxQuestions = 10
	examId := s.examIdMap[objective]

	questionRow, err := s.queries.GetQuestions(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})

	if err != nil {
		return err
	}

	var newQuestions []quiz.GetQuestionsRow
	var reviewedQuestions []quiz.GetQuestionsRow

	for _, question := range questionRow {
		if question.QuestionID.Int32 == 0 {
			newQuestions = append(newQuestions, question)
		} else {
			reviewedQuestions = append(reviewedQuestions, question)
		}
	}

	questionSize := min(len(questionRow), MaxQuestions)
	var quizVal []QuizRow

	slices.SortFunc(reviewedQuestions, func(a, b quiz.GetQuestionsRow) int {
		if n := cmp.Compare(a.Streak.Int32, b.Streak.Int32); n != 0 {
			return n
		}
		aRatio := float64(a.RightCount.Int32) / max(float64(a.WrongCount.Int32), 1)
		bRatio := float64(b.RightCount.Int32) / max(float64(b.WrongCount.Int32), 1)
		return cmp.Compare(aRatio, bRatio)
	})

	rand.Shuffle(len(newQuestions), func(i, j int) {
		newQuestions[i], newQuestions[j] = newQuestions[j], newQuestions[i]
	})

	if len(reviewedQuestions) == 0 {

		for i := range questionSize {
			quizVal = append(quizVal, makeQuizRow(newQuestions[i]))
		}
	} else if len(newQuestions) == 0 {
		for i := range questionSize {
			quizVal = append(quizVal, makeQuizRow(reviewedQuestions[i]))
		}
	} else {
		numReviewQuestions := min(len(reviewedQuestions), 5)
		streakCount := 0
		for i := range numReviewQuestions {
			streakCount += int(reviewedQuestions[i].Streak.Int32)
		}

		streakRatio := float64(streakCount) / float64(numReviewQuestions)
		var numQuestions int
		switch {
		case streakCount == 0:
			numQuestions = min(len(reviewedQuestions), 5)
		case streakRatio <= 1.0:
			numQuestions = min(len(reviewedQuestions), 3)
		case streakRatio <= 1.5:
			numQuestions = min(len(reviewedQuestions), 2)
		default:
			numQuestions = 1
		}

		for i := range numQuestions {
			quizVal = append(quizVal, makeQuizRow(reviewedQuestions[i]))
		}

		numNewQuestions := min(len(newQuestions), MaxQuestions-numQuestions)

		for i := range numNewQuestions {
			quizVal = append(quizVal, makeQuizRow(newQuestions[i]))
		}
	}

	return s.batchQuizQuestion(ctx, quizVal)
}

func (s *QuizService) batchQuizQuestion(ctx context.Context, quizVal []QuizRow) error {
	batch := makeBatchQuizQuestion(quizVal)
	log.Printf("batch %v\n", batch)
	return s.queries.BatchQuizQuestion(ctx, batch)
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

func (s *QuizService) GetQuestionSetByExamId(ctx context.Context, objectiveId float64) ([]string, error) {
	examId := s.examIdMap[objectiveId]
	return s.queries.GetQuestionSetByExamId(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
}

// func (s *QuizService) GetObjectiveRating(ctx context.Context) ([]ObjectiveRating, error) {
// 	objectiveStats, err := s.queries.StatsByObjective(ctx)
// 	var ratings []ObjectiveRating
//
// 	if err != nil {
// 		return []ObjectiveRating{}, err
// 	}
//
// 	for _, ob := range objectiveStats {
//
// 		stat := Stat{
// 			total:   int(ob.NumRight + ob.NumWrong),
// 			correct: float64(ob.NumRight) / float64(ob.NumRight+ob.NumWrong),
// 			term:    float64(ob.NumTermsAttempted) / float64(ob.NumTerms),
// 		}
//
// 		ratings = append(ratings, ObjectiveRating{
// 			Code:  ob.Code,
// 			Stars: getStarRating(stat),
// 		})
// 	}
//
// 	return ratings, nil
// }

func (s *QuizService) SaveQuizQuestions(ctx context.Context, reviewed Question, questions []Question) error {
	_, err := s.queries.CreateQuizQuestion(ctx, quiz.CreateQuizQuestionParams{
		QuestionID:  pgtype.Int4{Int32: reviewed.ID, Valid: true},
		ObjectiveID: pgtype.Int4{Int32: reviewed.ObjectiveID, Valid: true},
		Status:      quiz.NullQuestionStatus{QuestionStatus: quiz.QuestionStatusReviewing, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	quizQuestion := makeBatchQuizQuestion(questions)
	return s.queries.BatchQuizQuestion(ctx, quizQuestion)
}
