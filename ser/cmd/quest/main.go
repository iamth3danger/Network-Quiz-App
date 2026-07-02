package main

import (
	"context"
	"dorzi/server/ai"
	"dorzi/server/db"
	"dorzi/server/services"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go"
	"log"
	"strconv"
	"time"
)

func BuildQuestionSet(ctx context.Context, client anthropic.Client, service *services.QuizService) error {

	const MaxTermEmpty = 3
	codes, err := service.GetExamCodes(ctx)
	codeNum := make([]float64, len(codes))
	if err != nil {
		return err
	}

	for i := range codes {
		cf, err := strconv.ParseFloat(codes[i], 64)
		if err != nil {
			return err
		}
		codeNum[i] = cf
	}

	for _, code := range codeNum {
		termEmptyCount := 0

		for {
			req, err := service.GenerateRequest(ctx, code, 10)

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
				start := time.Now()
				_, err := service.SaveQuestion(ctx, question)
				if err != nil {
					return err
				}
				log.Printf("Save Question took %v", time.Since(start))
			}

			if termEmptyCount < MaxTermEmpty {
				terms, err := service.GetAllTerms(ctx, code)
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
					service.SaveTerm(ctx, nt, code)
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
	db.Init("../../.env")

	quizService := services.NewQuizService()
	client := anthropic.NewClient()
	ctx := context.Background()

	quizService.Init(ctx)
	err := BuildQuestionSet(ctx, client, quizService)
	if err != nil {
		fmt.Printf("%w\n", err)
	}
}
