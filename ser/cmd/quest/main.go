// Command quest is a one-off script that fills the db with AI generated
// questions. It talks to the db directly and never builds a Store — the
// running server owns that.
package main

import (
	"context"
	"dorzi/server/ai"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"dorzi/server/services"
	"fmt"
	"os"
	"strconv"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
)

// examCodes maps exam objective id -> objective code (1.1, 1.2, ...).
func examCodes(ctx context.Context, queries *quiz.Queries) (map[int32]float64, error) {
	objectives, err := queries.GetAllExamObjectives(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching exam objectives: %w", err)
	}

	codes := make(map[int32]float64, len(objectives))
	for _, o := range objectives {
		code, err := strconv.ParseFloat(o.Code, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing objective code %q: %w", o.Code, err)
		}
		codes[o.ID] = code
	}

	return codes, nil
}

// termIndex maps term name -> term row, so a generated question can be matched
// back to the term it belongs to. Re-read after saving terms.
func termIndex(ctx context.Context, queries *quiz.Queries) (map[string]quiz.Term, error) {
	terms, err := queries.GetAllTerms(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching terms: %w", err)
	}

	index := make(map[string]quiz.Term, len(terms))
	for _, t := range terms {
		index[t.TermName] = t
	}

	return index, nil
}

func BuildQuestionSet(ctx context.Context, client anthropic.Client, queries *quiz.Queries, service *services.AIGenService) error {
	const MaxTermEmpty = 3

	codes, err := examCodes(ctx, queries)
	if err != nil {
		return err
	}

	terms, err := termIndex(ctx, queries)
	if err != nil {
		return err
	}

	for examId, code := range codes {
		termEmptyCount := 0

		for {
			req, err := service.GenerateRequest(ctx, code, int(examId), 10)
			if err != nil {
				return err
			}

			for _, r := range req.ReviewTerms {
				fmt.Printf("%v\n", r)
			}

			if len(req.ReviewTerms) == 0 {
				break
			}

			questions, err := ai.GenerateQuestions(ctx, &client, req)
			if err != nil {
				return err
			}

			for _, question := range questions {
				term, ok := terms[question.Term]
				if !ok || term.ObjectiveID.Int32 != examId {
					fmt.Printf("skipping question: term %q not on objective %.1f\n", question.Term, code)
					continue
				}
				if _, err := service.SaveQuestion(ctx, question, examId, term.ID); err != nil {
					return err
				}
			}

			if termEmptyCount < MaxTermEmpty {
				termNames, err := service.GetAllTerms(ctx, examId)
				if err != nil {
					return err
				}

				if len(termNames) >= 50 {
					termEmptyCount = MaxTermEmpty
					continue
				}

				termReq := ai.TermRequest{
					ExamObjective: code,
					Term:          termNames,
				}
				newTerms, err := ai.GenerateTerms(ctx, &client, termReq)
				if err != nil {
					return err
				}

				for _, nt := range newTerms.Terms {
					if err := service.SaveTerm(ctx, nt, examId); err != nil {
						return err
					}
				}

				if len(newTerms.Terms) == 0 {
					termEmptyCount += 1
				}

				// pick up the terms just written so the next round can match
				// its questions against them
				terms, err = termIndex(ctx, queries)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	ctx := context.Background()

	// run from ser/ or from ser/cmd/quest, either works
	for _, envPath := range []string{".env", "../../.env"} {
		if _, err := os.Stat(envPath); err != nil {
			continue
		}
		if err := godotenv.Load(envPath); err != nil {
			fmt.Printf("loading %s: %v\n", envPath, err)
			return
		}
		break
	}

	db.Init()
	queries := quiz.New(db.DB)
	aiService := services.NewAIGenService(queries)

	client := anthropic.NewClient()

	if err := BuildQuestionSet(ctx, client, queries, &aiService); err != nil {
		fmt.Printf("%v\n", err)
	}
}
