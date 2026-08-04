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
	subjectIdText, subjectTextId, err := s.fetchSubjects(ctx)
	if err != nil {
		return &Store{}, err
	}
	questions, questionExamId, questionSubjectId, err := s.fetchQuestions(ctx)
	if err != nil {
		return &Store{}, err
	}
	examQuiz, subjectQuiz, err := s.fetchQuiz(ctx)
	if err != nil {
		return &Store{}, err
	}
	termExamId, terms, err := s.fetchTerms(ctx)
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
	autoGenScores, err := s.fetchAutogenScores(ctx)
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
		examIdCode:        examIdCode,
		examCodeId:        examCodeId,
		subjectTextId:     subjectTextId,
		subjectIdText:     subjectIdText,
		questions:         questions,
		questionExamId:    questionExamId,
		questionSubjectId: questionSubjectId,
		examQuiz:          examQuiz,
		subjectQuiz:       subjectQuiz,
		terms:             terms,
		termExamId:        termExamId,
		ports:             ports,
		portScores:        portScores,
		objectiveScores:   objectiveScores,
		autogenScores:     autoGenScores,
		scores:            scores,
		matchingQuiz:      matchingQuiz,
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

	examQuiz, subQuiz, err := s.fetchQuiz(ctx)

	if err != nil {
		return nil, err
	}

	store.scores = scores
	store.objectiveScores = objectiveScores
	store.examQuiz = examQuiz
	store.subjectQuiz = subQuiz
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

func (s *StoreService) UpdateStoreOnQuiz(ctx context.Context, store *Store) (*Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	examQuiz, subQuiz, err := s.fetchQuiz(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w : %w", ErrDBFetch, err)
	}

	store.examQuiz = examQuiz
	store.subjectQuiz = subQuiz
	return store, nil
}

func (s *StoreService) UpdateStoreOnMatching(ctx context.Context, store *Store) (*Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	matching, err := s.fetchMatchingQuiz(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w : %w", ErrDBFetch, err)
	}

	store.matchingQuiz = matching
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

func (s *StoreService) fetchSubjects(ctx context.Context) (map[int32]string, map[string]int32, error) {
	subjectIdText := make(map[int32]string)
	subjectTextId := make(map[string]int32)

	subjectDb, err := s.queries.GetAllSubjects(ctx)

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, sub := range subjectDb {
		subjectIdText[sub.ID] = sub.Title
		subjectTextId[sub.Title] = sub.ID
	}

	return subjectIdText, subjectTextId, nil
}

func (s *StoreService) fetchQuestions(ctx context.Context) (map[int32]Question, map[int32][]Question, map[int32][]Question, error) {
	questions := make(map[int32]Question)
	questionExamId := make(map[int32][]Question)
	questionSubjectId := make(map[int32][]Question)
	questionDb, err := s.queries.GetAllQuestions(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, q := range questionDb {
		qe := questionFromDB(q)
		questions[qe.ID] = qe
		if qe.ObjectiveID != 0 {
			questionExamId[qe.ObjectiveID] = append(questionExamId[qe.ObjectiveID], qe)
		} else {
			questionSubjectId[qe.SubjectID] = append(questionSubjectId[qe.SubjectID], qe)
		}

	}

	return questions, questionExamId, questionSubjectId, nil
}

func (s *StoreService) fetchQuiz(ctx context.Context) (map[int32][]Quiz, map[int32][]Quiz, error) {
	examQuiz := make(map[int32][]Quiz)
	subQuiz := make(map[int32][]Quiz)
	quizDb, err := s.queries.GetAllQuiz(ctx)

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, q := range quizDb {
		qi := quizFromDB(q)
		if qi.ObjectiveID != 0 {
			examQuiz[qi.ObjectiveID] = append(examQuiz[qi.ObjectiveID], qi)
		} else {
			subQuiz[qi.SubjectID] = append(subQuiz[qi.SubjectID], qi)
		}
	}

	return examQuiz, subQuiz, nil
}

func (s *StoreService) fetchTerms(ctx context.Context) (map[string]int32, map[string]int32, error) {
	termExamId := make(map[string]int32)
	terms := make(map[string]int32)
	termDb, err := s.queries.GetAllTerms(ctx)

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, t := range termDb {
		termExamId[t.TermName] = t.ObjectiveID.Int32
		terms[t.TermName] = t.ID
	}

	return termExamId, terms, nil
}

func (s *StoreService) fetchPorts(ctx context.Context) ([]Port, error) {
	var ports []Port
	portDb, err := s.queries.GetAllPorts(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, p := range portDb {
		ports = append(ports, portFromDB(p))
	}

	return ports, nil
}

func (s *StoreService) fetchPortScores(ctx context.Context) (map[int32]Scorer, error) {
	portScores := make(map[int32]Scorer)
	portScoreDb, err := s.queries.GetAllPortScores(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, p := range portScoreDb {
		ps := portScoreFromDB(p)
		portScores[ps.PortID] = &ps
	}

	return portScores, nil
}

func (s *StoreService) fetchAutogenScores(ctx context.Context) (map[Generator]Scorer, error) {
	autogenScores := make(map[Generator]Scorer)
	autogenDb, err := s.queries.GetAutoGenScores(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, score := range autogenDb {
		autoScoreFromDB(score, autogenScores)
	}

	for _, gen := range Generators {
		_, ok := autogenScores[gen]

		if !ok {
			score := Score{
				RightCount: 0,
				WrongCount: 0,
				Streak:     0,
			}
			var scorer Scorer = &score
			autogenScores[gen] = scorer
		}
	}

	return autogenScores, nil
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

func (s *StoreService) fetchScores(ctx context.Context) (map[int32]Scorer, error) {
	scores := make(map[int32]Scorer)
	scoreDb, err := s.queries.GetAllScores(ctx)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDBFetch, err)
	}

	for _, s := range scoreDb {
		sc := scoreFromDB(s)
		scores[sc.QuestionID] = &sc
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
