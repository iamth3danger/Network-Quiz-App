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
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5"
)

type ErrEmptyParam struct{ Param string }
type ErrInvalidArg struct{ Param, Value string }

var ErrTooManyParam = errors.New("too many params!")

func (e ErrEmptyParam) Error() string { return fmt.Sprintf("missing required params: %s", e.Param) }
func (e ErrInvalidArg) Error() string {
	return fmt.Sprintf("invalid value for %s: %q", e.Param, e.Value)
}

type HttpParams struct {
	examObjective float64
	subjectTitle  string
}

type App struct {
	quizService  *services.QuizService
	portService  *services.PortService
	store        *services.Store
	storeService *services.StoreService
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
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, services.ErrStoreNotExist) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, services.ErrInvalidObjective) {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if errors.Is(err, services.ErrStoreNotExist) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeError(w, http.StatusInternalServerError, err)
}

func (a *App) GetStarRating(w http.ResponseWriter, r *http.Request) {
	ratings, err := a.store.GetObjectiveRating()
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
	a.store, err = a.storeService.UpdateStoreOnQuestion(r.Context(), a.store)

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

	err = a.portService.UpdatePorts(r.Context(), req)
	if err != nil {
		handleError(w, err)
		return
	}

	a.store, err = a.storeService.UpdateStoreOnPort(r.Context(), a.store)

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

func parseParams(r *http.Request) (HttpParams, error) {

	examObjective, err := parseExamObjective(r)

	if err != nil {
		var e ErrInvalidArg
		if errors.As(err, &e) {
			return HttpParams{}, err
		}
	}

	subject, err := parseSubject(r)
	fmt.Printf("subject: %s\n", subject)

	if err != nil {
		var e ErrInvalidArg
		if errors.As(err, &e) {
			return HttpParams{}, err
		}
	}

	if subject != "" && examObjective != 0 {
		return HttpParams{}, fmt.Errorf("%w", ErrTooManyParam)
	}

	if subject == "" && examObjective == 0 {
		return HttpParams{}, ErrEmptyParam{Param: "subject or exam_obj"}
	}

	return HttpParams{
		examObjective: examObjective,
		subjectTitle:  subject,
	}, nil
}

func parseSubject(r *http.Request) (string, error) {
	subject := r.URL.Query().Get("subject")

	if subject == "" {
		return "", ErrEmptyParam{Param: "subject"}
	}

	subject = strings.ReplaceAll(subject, "_", " ")

	return subject, nil
}

func (a *App) GetQuestionNumber(w http.ResponseWriter, r *http.Request) {

	params, err := parseParams(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	examObj := params.examObjective
	subject := params.subjectTitle

	var quizInfo services.QuizStatus

	if examObj != 0 {
		quizInfo = a.store.GetQuizAnsweredByExam(examObj)
	} else {
		quizInfo = a.store.GetQuizAnsweredBySubject(subject)
	}

	writeJSON(w, http.StatusOK, quizInfo)
}

func (a *App) GetQuestion(w http.ResponseWriter, r *http.Request) {

	params, err := parseParams(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	examObj := params.examObjective
	subject := params.subjectTitle

	var question services.Question

	if examObj != 0 {
		question, err = a.store.FetchQuestionExam(examObj)
		if err != nil {
			handleError(w, err)
			return
		}
	} else {
		question, err = a.store.FetchQuestionSubject(subject)
		if err != nil {
			handleError(w, err)
			return
		}
	}

	go func() {
		err := a.quizService.SetQuizStatusReviewing(context.Background(), question.ID)
		if err != nil {
			log.Printf("failed to set quiz status: %v", err)
		}
	}()
	writeJSON(w, http.StatusOK, question)
}

func (a *App) GenerateQuiz(w http.ResponseWriter, r *http.Request) {
	params, err := parseParams(r)

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("request requires exam_objective"))
		return
	}

	examObj := params.examObjective
	now := time.Now().UnixNano()
	subject := params.subjectTitle
	var questions []services.Question

	if examObj != 0 {
		questions, err = a.store.GenerateQuizFromExam(examObj)
		log.Printf("time taken to generate quiz %d", time.Now().UnixNano()-now)

		if err != nil {
			handleError(w, err)
		}
	} else {
		questions, err = a.store.GenerateQuizFromSubject(subject)
		log.Printf("time taken to generate quiz %d", time.Now().UnixNano()-now)

		if err != nil {
			handleError(w, err)
		}
	}

	retQuestion := questions[0]
	questions = questions[1:]
	go func() {
		ctx := context.Background()

		var id int32
		if examObj != 0 {
			id, err = a.store.GetExamCodeById(examObj)
			if err != nil {
				log.Printf("%v", err)
			}
		} else {
			id, err = a.store.GetSubjectIdByTitle(subject)
			if err != nil {
				log.Printf("%v", err)
			}
		}

		if examObj != 0 {
			err = a.quizService.DeleteQuizExam(ctx, id)
			if err != nil {
				log.Printf("cooked %v", err)
			}
		} else {
			err = a.quizService.DeleteQuizSubject(ctx, id)
			if err != nil {
				log.Printf("cooked %v", err)
			}
		}

		err = a.quizService.SaveQuizQuestions(ctx, retQuestion, questions)
		if err != nil {
			log.Printf("cooked %v", err)
		}
		a.store, err = a.storeService.UpdateStoreOnQuiz(ctx, a.store)
		if err != nil {
			log.Printf("%v", err)
		}
	}()

	writeJSON(w, http.StatusOK, retQuestion)
}

func (a *App) GenerateMatching(w http.ResponseWriter, r *http.Request) {
	matching := a.store.GenerateMatching()

	go func() {
		ctx := context.Background()
		err := a.portService.DeleteMatchingQuiz(ctx)
		if err != nil {
			log.Printf("deleting from db %v", err)
		}

		err = a.portService.SaveMatchingQuiz(ctx, matching)
		if err != nil {
			log.Printf("error saving to db %v", err)
		}

		a.store, err = a.storeService.UpdateStoreOnMatching(ctx, a.store)
	}()

	writeJSON(w, http.StatusOK, matching)
}

func (a *App) GetTerms(w http.ResponseWriter, r *http.Request) {
	termStats, err := a.quizService.FetchTermStats(r.Context())
	if err != nil {
		handleError(w, err)
	}

	writeJSON(w, http.StatusOK, termStats)
}

func main() {

	ctx := context.Background()
	db.Init()
	quizDB := quiz.New(db.DB)
	storeService := services.NewStoreService(quizDB)
	portService := services.NewPortService(quizDB)
	quizService := services.NewQuizService(quizDB)
	store, err := storeService.NewStore(ctx)
	if err != nil {
		log.Fatalf("unable to generate store %v", err)
		return
	}

	app := App{quizService: &quizService, store: store, portService: &portService, storeService: &storeService}

	fmt.Println("Starting server")

	http.HandleFunc("/quiz/count", app.GetQuestionNumber)
	http.HandleFunc("/terms", app.GetTerms)
	http.HandleFunc("/rating", app.GetStarRating)
	http.HandleFunc("/questions/generate", app.GenerateQuiz)
	http.HandleFunc("/questions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.GetQuestion(w, r)
		case http.MethodPost:
			app.UpdateQuestion(w, r)
		}
	})

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
