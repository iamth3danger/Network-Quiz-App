package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/db/quiz"
	"dorzi/server/services"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"strconv"
	"time"
)

type ErrEmptyParam struct{ Param string }
type ErrInvalidArg struct{ Param, Value string }

func (e ErrEmptyParam) Error() string { return fmt.Sprintf("missing required params: %s", e.Param) }
func (e ErrInvalidArg) Error() string {
	return fmt.Sprintf("invalid value for %s: %q", e.Param, e.Value)
}

type App struct {
	quizService *services.QuizService
}

const portNum string = ":8080"

var client = anthropic.NewClient()

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WriteOk(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"data":   data,
	})
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"error":  err.Error(),
	})
}

func handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, services.ErrInvalidObjective) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeError(w, http.StatusInternalServerError, err)
}

func (a *App) GetStarRating(w http.ResponseWriter, r *http.Request) {
	ratings, err := a.quizService.GetObjectiveRating(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ratings)
}

func (a *App) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	var req services.QuestionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}
	err = a.quizService.UpdateQuestion(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteOk(w, http.StatusOK)
}

func (a *App) UpdatePort(w http.ResponseWriter, r *http.Request) {
	var req []services.MatchingRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		handleError(w, err)
		return
	}

	err = a.quizService.UpdatePorts(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}
	WriteOk(w, http.StatusOK)
}

func parseExamObjective(r *http.Request) (float64, error) {

	examObjective := r.URL.Query().Get("exam_objective")

	if examObjective == "" {
		return 0, ErrEmptyParam{Param: "exam_objective"}
	}

	examObj, err := strconv.ParseFloat(examObjective, 64)
	if err != nil {
		return 0, ErrInvalidArg{Param: "exam_objective", Value: examObjective}
	}

	return examObj, nil
}

func (a *App) GetQuestionNumber(w http.ResponseWriter, r *http.Request) {

	examObj, err := parseExamObjective(r)
	if err != nil {
		handleError(w, err)
		return
	}

	quizInfo, err := a.quizService.GetQuizAnswered(r.Context(), examObj)

	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, quizInfo)
}

func (a *App) GetQuestion(w http.ResponseWriter, r *http.Request) {

	examObj, err := parseExamObjective(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("request requires exam_objective"))
		return
	}

	question, err := a.quizService.FetchQuestion(r.Context(), examObj)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, question)
}

func (a *App) GenerateQuiz(w http.ResponseWriter, r *http.Request) {
	examObj, err := parseExamObjective(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("request requires exam_objective"))
		return
	}

	err = a.quizService.DeleteQuiz(r.Context(), examObj)
	if err != nil {
		handleError(w, err)
	}

	now := time.Now().UnixMilli()
	err = a.quizService.GenerateQuiz(r.Context(), examObj)
	log.Printf("time taken to generate quiz %d", time.Now().UnixMilli()-now)

	if err != nil {
		handleError(w, err)
	}

	question, err := a.quizService.FetchQuestion(r.Context(), examObj)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, question)
}

func (a *App) GenerateMatching(w http.ResponseWriter, r *http.Request) {
	matching, err := a.quizService.GetMatchingQuiz(r.Context())
	if err != nil {
		handleError(w, err)
	}

	writeJSON(w, http.StatusOK, matching)
}

func (a *App) GetTerms(w http.ResponseWriter, r *http.Request) {
	termStats, err := a.quizService.FetchTermStats(r.Context())
	if err != nil {
		handleError(w, err)
	}

	writeJSON(w, http.StatusOK, termStats)
}

func (a *App) WrongInfo(w http.ResponseWriter, r *http.Request) {
	var questions []quiz.Question
	examObj, err := parseExamObjective(r)

	switch {
	case err == nil:
		questions, err = a.quizService.GetWrongQuestionsByExam(r.Context(), examObj)
		if err != nil {
			handleError(w, err)
			return
		}
	case errors.As(err, &ErrEmptyParam{}):
		questions, err = a.quizService.GetAllWrongQuestions(r.Context())
		if err != nil {
			handleError(w, err)
			return
		}
	default:
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, q := range questions {
		fmt.Fprintf(w, "A. %s\n", q.A)
		fmt.Fprintf(w, "B. %s\n", q.B)
		fmt.Fprintf(w, "C. %s\n", q.C)
		fmt.Fprintf(w, "D. %s\n", q.D)
	}
}

func (a *App) QuestionSetByExamId(w http.ResponseWriter, r *http.Request) {
	examObj, err := parseExamObjective(r)
	questions, err := a.quizService.GetQuestionSetByExamId(r.Context(), examObj)
	if err != nil {
		handleError(w, err)
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, q := range questions {
		fmt.Fprintf(w, "%s\n", q)
	}
}

func main() {

	ctx := context.Background()
	db.Init("../.env")
	app := App{quizService: services.NewQuizService()}
	err := app.quizService.Init(ctx)
	if err != nil {
		log.Fatalf("Error creating map %v", err)
	}

	fmt.Println("Starting server")

	http.HandleFunc("/quiz/count", app.GetQuestionNumber)
	http.HandleFunc("/terms", app.GetTerms)
	http.HandleFunc("/rating", app.GetStarRating)
	http.HandleFunc("/questions/generate", app.GenerateQuiz)
	http.HandleFunc("/questions/set", app.QuestionSetByExamId)
	http.HandleFunc("/questions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetQuestion(w, r)
		case http.MethodPost:
			app.UpdateQuestion(w, r)
		}
	})
	http.HandleFunc("/wrong", app.WrongInfo)
	http.HandleFunc("/matching", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GenerateMatching(w, r)
		case http.MethodPost:
			app.UpdatePort(w, r)
		}
	})
	http.HandleFunc("/matching/generate", app.GenerateMatching)

	log.Println("Started on port", portNum)

	http_err := http.ListenAndServe(portNum, corsMiddleware(http.DefaultServeMux))
	if http_err != nil {
		log.Fatal(http_err)
	}

}
