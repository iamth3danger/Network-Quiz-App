package ai

import (
	"context"
	"encoding/json"
	"fmt"
	anthropic "github.com/anthropics/anthropic-sdk-go"
	"strings"
)

type AIRequest struct {
	ExamObjective float64
	ReviewTerms   []string
	Questions     []string
}

type AIQuestion struct {
	Question      string  `json:"question"`
	A             string  `json:"a"`
	B             string  `json:"b"`
	C             string  `json:"c"`
	D             string  `json:"d"`
	Answer        string  `json:"answer"`
	Term          string  `json:"term"`
	ExamObjective float64 `json:"exam_objective"`
}

type TermRequest struct {
	ExamObjective float64
	Term          []string
}

type AITerms struct {
	Objective string   `json:"objective"`
	Terms     []string `json:"terms"`
}

func GenerateTerms(ctx context.Context, client *anthropic.Client, request TermRequest) (AITerms, error) {
	prompt := fmt.Sprintf(
		"List 20 key terms for CompTIA N10-009 Network+ exam objective %.1f. "+
			"Focus on terms a student would need to memorize, not exhaustive technical detail. "+
			"Exclude terms already in this list: %s. "+
			"Return only standardized networking terms as they appear in official documentation. "+
			"Avoid variants or subcategories — use 'VPN' not 'site-to-site VPN', 'STP' not 'rapid spanning tree'. "+
			"I want max 50 terms per category, so make sure each term is important"+
			`Return only a raw JSON object. No markdown, no backticks, no explanation, no leading characters. Format: {"objective":"","terms":[]}`,
		request.ExamObjective,
		strings.Join(request.Term, ", "),
	)

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return AITerms{}, fmt.Errorf("anthropic error: %w", err)
	}

	raw := strings.TrimSpace(msg.Content[0].Text)
	raw = strings.TrimPrefix(raw, ": ")

	var terms AITerms
	if err := json.Unmarshal([]byte(raw), &terms); err != nil {
		return AITerms{}, fmt.Errorf("failed to parse response: %w", err)
	}
	return terms, nil
}

func GenerateQuestions(ctx context.Context, client *anthropic.Client, request AIRequest) ([]AIQuestion, error) {
	prompt := "You are a CompTIA n10-009 Network+ question maker. \n"
	prompt += fmt.Sprintf("Make questions covering exam objective %.1f .\n", request.ExamObjective)

	if len(request.Questions) > 0 {
		prompt += fmt.Sprintf("Do not return any of these questions %s\n", strings.Join(request.Questions, ", "))
	}

	if len(request.ReviewTerms) > 0 {
		prompt += fmt.Sprintf("Create one question for each of these terms %s\n",
			strings.Join(request.ReviewTerms, ", "))
	} else {
		return nil, fmt.Errorf("review terms empty")
	}
	prompt += `Return only a raw JSON array. No markdown, no backticks, no explanation. Answer column should be "a", "b", "c"  or "d". The term value must be copied character-for-character from the provided term list — do not alter spelling, capitalization, punctuation, or spacing.
	Format: [{"question":"","a":"","b":"","c":"","d":"","answer":"","term":""}]`

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic error: %w", err)
	}

	var questions []AIQuestion
	if err := json.Unmarshal([]byte(msg.Content[0].Text), &questions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	for i := range questions {
		questions[i].ExamObjective = request.ExamObjective
	}

	return questions, nil
}
