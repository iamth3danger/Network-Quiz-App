package services

import (
	"context"
	"dorzi/server/db/quiz"
	"fmt"
	"strconv"
	"sync"
)

type StoreService struct {
	queries *quiz.Queries
	mu      sync.RWMutex
}

func NewStoreService(q *quiz.Queries) StoreService {
	return StoreService{queries: q}
}

func (s *StoreService) NewStore(ctx context.Context) (*Store, error) {
	examIdCode, examCodeId, err := s.fetchExamObjectives(ctx)
	if err != nil {
		return &Store{}, err
	}
	questions, questionExamId, err := s.fetchQuestions(ctx)
	if err != nil {
		return &Store{}, err
	}
	quiz, err := s.fetchQuiz(ctx)
	if err != nil {
		return &Store{}, err
	}
	terms, err := s.fetchTerms(ctx)
	if err != nil {
		return &Store{}, err
	}
	ports, err := s.fetchPorts(ctx)
	if err != nil {
		return &Store{}, err
	}
	portScores, err := s.fetchPortScores(ctx)
	if err != nil {
		return &Store{}, err
	}
	objectiveScores, err := s.fetchObjectiveScores(ctx)
	if err != nil {
		return &Store{}, err
	}
	scores, err := s.fetchScores(ctx)
	if err != nil {
		return &Store{}, err
	}
	matchingQuiz, err := s.fetchMatchingQuiz(ctx)
	if err != nil {
		return &Store{}, err
	}

	return &Store{
		examIdCode:      examIdCode,
		examCodeId:      examCodeId,
		questions:       questions,
		questionExamId:  questionExamId,
		quiz:            quiz,
		terms:           terms,
		ports:           ports,
		portScores:      portScores,
		objectiveScores: objectiveScores,
		scores:          scores,
		matchingQuiz:    matchingQuiz,
	}, nil

}

func (s *StoreService) UpdateStoreOnQuestion(ctx context.Context, store *Store) (*Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	scores, err := s.fetchScores(ctx)
	if err != nil {
		return nil, err
	}
	objectiveScores, err := s.fetchObjectiveScores(ctx)
	if err != nil {
		return nil, err
	}

	quiz, err := s.fetchQuiz(ctx)

	if err != nil {
		return nil, err
	}

	store.scores = scores
	store.objectiveScores = objectiveScores
	store.quiz = quiz
	return store, nil
}

func (s *StoreService) UpdateStoreOnPort(ctx context.Context, store *Store) (*Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	portScores, err := s.fetchPortScores(ctx)

	if err != nil {
		return nil, err
	}

	store.portScores = portScores
	return store, nil
}

func (s *StoreService) fetchExamObjectives(ctx context.Context) (map[int32]float64, map[float64]int32, error) {
	var examObjectives []ExamObjective
	examIdCode := make(map[int32]float64)
	examCodeId := make(map[float64]int32)

	examObjDb, err := s.queries.GetAllExamObjectives(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, ex := range examObjDb {
		examObjectives = append(examObjectives, examFromDB(ex))
	}

	for _, ex := range examObjectives {
		code, err := strconv.ParseFloat(ex.Code, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing float %w", err)
		}
		examIdCode[ex.ID] = code
		examCodeId[code] = ex.ID
	}

	return examIdCode, examCodeId, nil
}

func (s *StoreService) fetchQuestions(ctx context.Context) (map[int32]Question, map[int32][]Question, error) {
	questions := make(map[int32]Question)
	questionExamId := make(map[int32][]Question)
	questionDb, err := s.queries.GetAllQuestions(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, q := range questionDb {
		qe := questionFromDB(q)
		questions[qe.ID] = qe
		questionExamId[qe.ObjectiveID] = append(questionExamId[qe.ObjectiveID], qe)
	}

	return questions, questionExamId, nil
}

func (s *StoreService) fetchQuiz(ctx context.Context) (map[int32][]Quiz, error) {
	quizQuestions := make(map[int32][]Quiz)
	quizDb, err := s.queries.GetAllQuiz(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, q := range quizDb {
		qi := quizFromDB(q)
		quizQuestions[qi.ObjectiveID] = append(quizQuestions[qi.ObjectiveID], qi)
	}

	return quizQuestions, nil
}

func (s *StoreService) fetchTerms(ctx context.Context) (map[int32]Term, error) {
	terms := make(map[int32]Term)
	termDb, err := s.queries.GetAllTerms(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, t := range termDb {
		terms[t.ID] = termFromDB(t)
	}

	return terms, nil
}

func (s *StoreService) fetchPorts(ctx context.Context) (map[int32]Port, error) {
	ports := make(map[int32]Port)
	portDb, err := s.queries.GetAllPorts(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, p := range portDb {
		ports[p.ID] = portFromDB(p)
	}

	return ports, nil
}

func (s *StoreService) fetchPortScores(ctx context.Context) (map[int32]PortScore, error) {
	portScores := make(map[int32]PortScore)
	portScoreDb, err := s.queries.GetAllPortScores(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, p := range portScoreDb {
		ps := portScoreFromDB(p)
		portScores[ps.PortID] = ps
	}

	return portScores, nil
}

func (s *StoreService) fetchObjectiveScores(ctx context.Context) (map[int32]ObjectiveScore, error) {
	objectiveScores := make(map[int32]ObjectiveScore)
	objectiveScoreDb, err := s.queries.GetAllObjectiveScore(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, o := range objectiveScoreDb {
		os := objScoreFromDB(o)
		objectiveScores[os.ObjectiveID] = os
	}

	return objectiveScores, nil
}

func (s *StoreService) fetchScores(ctx context.Context) (map[int32]Score, error) {
	scores := make(map[int32]Score)
	scoreDb, err := s.queries.GetAllScores(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, s := range scoreDb {
		sc := scoreFromDB(s)
		scores[sc.QuestionID] = sc
	}

	return scores, nil
}

func (s *StoreService) fetchMatchingQuiz(ctx context.Context) ([]MatchingQuiz, error) {
	var matchingQuiz []MatchingQuiz
	matchingDb, err := s.queries.GetAllMatching(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, m := range matchingDb {
		matchingQuiz = append(matchingQuiz, matchingFromDB(m))
	}

	return matchingQuiz, nil
}
