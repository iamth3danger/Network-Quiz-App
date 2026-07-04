package services

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

type Store struct {
	examIdCode      map[int32]float64
	examCodeId      map[float64]int32
	questions       map[int32]Question
	questionExamId  map[int32][]Question
	scores          map[int32]Scorer
	objectiveScores map[int32]ObjectiveScore
	quiz            map[int32][]Quiz
	ports           []Port
	portScores      map[int32]Scorer
	termExamId      map[string]int32
	terms           map[string]int32
	matchingQuiz    []MatchingQuiz
	mu              sync.RWMutex
}

func (s *Store) FetchQuestion(objective float64) (Question, error) {
	examId, err := s.GetExamCodeById(objective)

	if err != nil {
		return Question{}, err
	}

	var pending []Quiz
	quiz, err := s.getQuizByExamId(examId)

	if err != nil {
		return Question{}, err
	}

	for _, q := range quiz {
		switch q.Status {
		case "reviewing":
			return s.questions[q.QuestionID], nil
		case "pending":
			pending = append(pending, q)
		default:
		}

	}

	if len(pending) == 0 {
		return Question{}, ErrStoreNotExist
	}

	rand.Shuffle(len(pending), func(i, j int) {
		pending[i], pending[j] = pending[j], pending[i]
	})

	err = s.setQuizQuestionReviewed(examId, pending[0].ID)

	if err != nil {
		return Question{}, err
	}

	return s.getQuestionById(pending[0].QuestionID)
}

func (s *Store) GetQuizAnswered(objective float64) (QuizStatus, error) {
	examId, err := s.GetExamCodeById(objective)

	if err != nil {
		return QuizStatus{}, err
	}

	quizQuestions, err := s.getQuizByExamId(examId)
	count := 0
	for _, q := range quizQuestions {
		if q.Status == "answered" {
			count += 1
		}
	}

	return QuizStatus{Count: count}, nil
}

func (s *Store) GenerateQuiz(objective float64) ([]Question, error) {
	examId, ok := s.examCodeId[objective]
	if !ok {
		return nil, fmt.Errorf("%w", ErrInvalidObjective)
	}
	questions := questToQuizer(s.questionExamId[examId])

	info := createQuizerInfo(questions, s.scores)
	info.rearrangeArrs()

	return toQuestions(info.makeQuiz()), nil
}

func (s *Store) GenerateMatching() []Port {
	ports := portToQuizer(s.ports)
	info := createQuizerInfo(ports, s.portScores)
	info.rearrangeArrs()
	matchingPorts := toPorts(info.makeMatching())

	for _, p := range matchingPorts {
		log.Printf("%v", p)
	}

	return matchingPorts
}

func (s *Store) GetObjectiveRating() ([]ObjectiveRating, error) {
	objStats, err := s.createObjectiveStats()
	var ratings []ObjectiveRating
	if err != nil {
		return nil, err
	}

	for _, s := range objStats {

		stat := Stat{
			total:   int(s.Right + s.Wrong),
			correct: float64(s.Right) / float64(s.Right+s.Wrong),
			term:    float64(s.AttemptedTerms) / float64(s.Terms),
		}

		ratings = append(ratings, ObjectiveRating{
			Code:  s.Code,
			Stars: getStarRating(stat),
		})

	}

	return ratings, nil
}

func (s *Store) createObjectiveStats() ([]ObjectiveStat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var objectiveStats []ObjectiveStat

	rightmap := make(map[float64]int32)
	wrongmap := make(map[float64]int32)
	terms := make(map[int32]struct{})
	seenTermMap := make(map[float64]int32)
	totalTermMap := make(map[float64]int32)

	for k, v := range s.examCodeId {
		objectiveScore, ok := s.objectiveScores[v]
		if !ok {
			rightmap[k] = 0
			wrongmap[k] = 0
		} else {
			rightmap[k] = objectiveScore.RightCount
			wrongmap[k] = objectiveScore.WrongCount
		}
	}

	for _, v := range s.questions {
		_, ok := terms[v.TermID]
		examId, ex := s.examIdCode[v.ObjectiveID]
		if !ex {
			return nil, fmt.Errorf("fetching exam id %d", v.ObjectiveID)
		}

		if !ok {
			terms[v.TermID] = struct{}{}
			totalTermMap[examId] = totalTermMap[examId] + 1
			_, k := s.scores[v.ID]
			if k {
				seenTermMap[examId] = seenTermMap[examId] + 1
			}
		}
	}

	for k := range s.examCodeId {
		stat := ObjectiveStat{
			Code:           k,
			Right:          rightmap[k],
			Wrong:          wrongmap[k],
			Terms:          totalTermMap[k],
			AttemptedTerms: seenTermMap[k],
		}
		objectiveStats = append(objectiveStats, stat)
	}

	return objectiveStats, nil
}

func (s *Store) getQuizByExamId(examId int32) ([]Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	quiz, ok := s.quiz[examId]
	if !ok {
		return nil, fmt.Errorf("%w: exam id %d", ErrStoreNotExist, examId)
	}

	return quiz, nil
}

func (s *Store) setQuizQuestionStatus(examId int32, questionID int32, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quiz, ok := s.quiz[examId]
	if !ok {
		return fmt.Errorf("%w: quiz", ErrStoreNotExist)
	}
	for i, q := range quiz {
		if q.ID == questionID {
			s.quiz[examId][i].Status = status
		}
	}
	return nil
}

func (s *Store) setQuizQuestionReviewed(examId int32, questionID int32) error {
	return s.setQuizQuestionStatus(examId, questionID, "reviewing")
}

func (s *Store) setQuizQuestionAnswered(examId int32, questionID int32) error {
	return s.setQuizQuestionStatus(examId, questionID, "answered")
}

func (s *Store) GetExamCodeById(objective float64) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	examId, ok := s.examCodeId[objective]
	if !ok {
		return 0, fmt.Errorf("%w: %.1f", ErrInvalidObjective, objective)
	}
	return examId, nil
}

func (s *Store) getQuestionById(questionId int32) (Question, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	question, ok := s.questions[questionId]
	if !ok {
		return Question{}, fmt.Errorf("invalid question id: %d", questionId)
	}
	return question, nil
}

func (s *Store) updateObjectiveScore(req QuestionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	question, err := s.getQuestionById(int32(req.QuestionID))
	if err != nil {
		return err
	}

	score, ok := s.objectiveScores[question.ID]
	if !ok {
		if req.Correct {
			s.objectiveScores[question.ID] = ObjectiveScore{
				ObjectiveID: question.ObjectiveID,
				RightCount:  1,
				WrongCount:  0,
			}
		} else {
			s.objectiveScores[question.ID] = ObjectiveScore{
				ObjectiveID: question.ObjectiveID,
				RightCount:  0,
				WrongCount:  1,
			}
		}
		return nil
	}

	if req.Correct {
		score.RightCount += 1
	} else {
		score.WrongCount += 1
	}

	s.objectiveScores[question.ID] = score
	return nil
}

func (s *Store) GetExamIdCodes() map[int32]float64 {
	return s.examIdCode
}

func (s *Store) ValidateTerm(termName string, examId int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.termExamId[termName] == examId
}

func (s *Store) GetTermId(termName string) int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.terms[termName]
}
