package main

import (
	"context"
	"dorzi/server/ai"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"dorzi/server/services"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go"
)

// BuildQuestionSet is probably buggy because store isn't updating terms on save or questions
// If running again pls fix

func BuildQuestionSet(ctx context.Context, client anthropic.Client, service *services.AIGenService, store *services.Store) error {
	const MaxTermEmpty = 3

	for k, code := range store.GetExamIdCodes() {
		termEmptyCount := 0

		for {
			req, err := service.GenerateRequest(ctx, code, int(k), 10)

			for _, r := range req.ReviewTerms {
				fmt.Printf("%v\n", r)
			}

			if err != nil {
				return err
			}

			if len(req.ReviewTerms) == 0 {
				break
			}

			questions, err := ai.GenerateQuestions(ctx, &client, req)
			if err != nil {
				return err
			}

			for _, question := range questions {
				if !store.ValidateTerm(question.Term, k) {
					return fmt.Errorf("term not in question examId")
				}
				_, err := service.SaveQuestion(ctx, question, k, store.GetTermId(question.Term))
				if err != nil {
					return err
				}
			}

			if termEmptyCount < MaxTermEmpty {
				terms, err := service.GetAllTerms(ctx, k)
				if len(terms) >= 50 {
					termEmptyCount = 5
					continue
				}

				if err != nil {
					return err
				}
				termReq := ai.TermRequest{
					ExamObjective: code,
					Term:          terms,
				}
				newTerms, err := ai.GenerateTerms(ctx, &client, termReq)
				if err != nil {
					return err
				}

				for _, nt := range newTerms.Terms {
					service.SaveTerm(ctx, nt, k)
				}

				if len(newTerms.Terms) == 0 {
					termEmptyCount += 1
				}
			}
		}
	}

	return nil
}

func main() {
	ctx := context.Background()
	db.Init("../../.env")
	quizDB := quiz.New(db.DB)
	aiService := services.NewAIGenService(quizDB)
	storeService := services.NewStoreService(quizDB)
	store, err := storeService.NewStore(ctx)

	if err != nil {
		fmt.Printf("%v\n", err)
	}

	client := anthropic.NewClient()

	err = BuildQuestionSet(ctx, client, &aiService, store)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
