package services

import (
	"cmp"
	"context"
	"dorzi/server/ai"
	"dorzi/server/db/quiz"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"slices"
	"strings"
)

type AIGenService struct {
	queries *quiz.Queries
}

func NewAIGenService(q *quiz.Queries) AIGenService {
	return AIGenService{
		queries: q,
	}
}

func (s *AIGenService) GenerateRequest(ctx context.Context, objective float64, examId int, numTerms int) (ai.AIRequest, error) {

	terms, err := s.getTerms(ctx, examId, numTerms)
	if err != nil {
		return ai.AIRequest{}, err
	}

	terms = slices.DeleteFunc(terms, func(term quiz.GetTermQuestionCountRow) bool {
		return term.Total > 2
	})

	termIds := make([]int32, len(terms))
	termNames := make([]string, len(terms))
	for i, term := range terms {
		termIds[i] = term.ID
		termNames[i] = term.TermName
	}

	questions, err := s.getQuestionsByIds(ctx, termIds)

	if err != nil {
		return ai.AIRequest{}, err
	}

	return ai.AIRequest{
		ExamObjective: objective,
		ReviewTerms:   termNames,
		Questions:     questions,
	}, nil
}

func (s *AIGenService) SaveQuestion(ctx context.Context, question ai.AIQuestion, examId int32, termId int32) (quiz.Question, error) {

	fmt.Printf("answer: %v\n", question.Answer)
	params := quiz.CreateQuestionParams{
		Question:    question.Question,
		A:           question.A,
		B:           question.B,
		C:           question.C,
		D:           question.D,
		ObjectiveID: pgtype.Int4{Int32: examId, Valid: true},
		Answer:      question.Answer,
		TermID:      pgtype.Int4{Int32: int32(termId), Valid: true},
	}

	return s.queries.CreateQuestion(ctx, params)
}

func (s *AIGenService) getTerms(ctx context.Context, examId int, numTerms int) ([]quiz.GetTermQuestionCountRow, error) {
	terms, err := s.queries.GetTermQuestionCount(ctx, pgtype.Int4{Int32: int32(examId), Valid: true})
	if err != nil {
		return nil, fmt.Errorf("fetching terms %v", err)
	}

	slices.SortFunc(terms, func(a, b quiz.GetTermQuestionCountRow) int {
		return cmp.Compare(a.Total, b.Total)
	})

	filtered := slices.DeleteFunc(terms, func(t quiz.GetTermQuestionCountRow) bool {
		return t.Total > 2
	})

	return filtered[:min(numTerms, len(filtered))], nil
}

func (s *AIGenService) createQuizQuestion(ctx context.Context, questionID int, examID int) error {

	quizQuestion := quiz.CreateQuizQuestionParams{
		QuestionID:  pgtype.Int4{Int32: int32(questionID), Valid: true},
		ObjectiveID: pgtype.Int4{Int32: int32(examID), Valid: true},
		Status:      quiz.NullQuestionStatus{QuestionStatus: quiz.QuestionStatusPending, Valid: true},
	}
	_, err := s.queries.CreateQuizQuestion(ctx, quizQuestion)

	return err
}

func (s *AIGenService) getQuestionsByIds(ctx context.Context, termIds []int32) ([]string, error) {
	return s.queries.GetQuestionsByIds(ctx, termIds)
}

func (s *AIGenService) SaveTerm(ctx context.Context, termName string, examId int32) error {
	_, err := s.queries.CreateTerm(ctx, quiz.CreateTermParams{
		ObjectiveID: pgtype.Int4{Int32: int32(examId), Valid: true},
		TermName:    strings.ToLower(termName),
	})
	return err
}

func (s *AIGenService) GetAllTerms(ctx context.Context, examId int32) ([]string, error) {
	return s.queries.GetTermsByObjectiveId(ctx, pgtype.Int4{Int32: examId, Valid: true})
}
