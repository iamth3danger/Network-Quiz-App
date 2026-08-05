package services

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

type QuizMap struct {
	mu   sync.RWMutex
	data map[int32][]Quiz
}

func (q *QuizMap) getQuizById(id int32) ([]Quiz, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	quiz, ok := q.data[id]
	if !ok {
		return nil, fmt.Errorf("%w: exam id %d", ErrStoreNotExist, id)
	}

	return quiz, nil
}

func (q *QuizMap) setQuizQuestionReviewed(id int32, questionID int32) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.setQuizQuestionStatus(id, questionID, "reviewing")
}

func (q *QuizMap) setQuizQuestionAnswered(id int32, questionID int32) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.setQuizQuestionStatus(id, questionID, "answered")
}

func (q *QuizMap) getQuizAnswered(id int32) QuizStatus {
	quiz := q.data[id]
	count := 0
	for _, q := range quiz {
		if q.Status == "answered" {
			count++
		}
	}

	return QuizStatus{
		Count: count,
	}
}

func (q *QuizMap) setQuizQuestionStatus(id int32, questionID int32, status string) error {

	quiz, ok := q.data[id]
	if !ok {
		return fmt.Errorf("%w: Quiz", ErrStoreNotExist)
	}
	for i, qe := range quiz {
		if qe.ID == questionID {
			q.data[id][i].Status = status
		}
	}
	return nil
}

type Store struct {
	examIdCode        map[int32]float64
	examCodeId        map[float64]int32
	subjectTextId     map[string]int32
	subjectIdText     map[int32]string
	questions         map[int32]Question
	questionExamId    map[int32][]Question
	questionSubjectId map[int32][]Question
	scores            map[int32]Scorer
	objectiveScores   map[int32]ObjectiveScore
	cidrGenerator     *CidrGenerator
	examQuiz          *QuizMap
	subjectQuiz       *QuizMap
	ports             []Port
	portScores        map[int32]Scorer
	termExamId        map[string]int32
	terms             map[string]int32
	matchingQuiz      []MatchingQuiz
	mu                sync.RWMutex
}

func (s *Store) FetchQuestionExam(objective float64) (Question, error) {
	examId, err := s.GetExamCodeById(objective)

	if err != nil {
		return Question{}, err
	}

	return s.fetchQuestion(examId, s.examQuiz)
}

func (s *Store) FetchQuestionSubject(subject string) (Question, error) {
	subId, err := s.GetSubjectIdByTitle(subject)

	if err != nil {
		return Question{}, err
	}

	return s.fetchQuestion(subId, s.subjectQuiz)
}

func (s *Store) fetchQuestion(id int32, quizMap *QuizMap) (Question, error) {
	var pending []Quiz
	// quiz, err := s.getQuizByExamId(id)
	quiz, err := quizMap.getQuizById(id)

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

	pick := pending[rand.Intn(len(pending))]
	err = quizMap.setQuizQuestionReviewed(id, pick.ID)

	if err != nil {
		return Question{}, err
	}

	return s.getQuestionById(pick.QuestionID)
}

func (s *Store) GetQuizAnsweredBySubject(title string) QuizStatus {
	id := s.subjectTextId[title]
	return s.subjectQuiz.getQuizAnswered(id)
}

func (s *Store) GetQuizAnsweredByExam(examObj float64) QuizStatus {
	id := s.examCodeId[examObj]
	return s.examQuiz.getQuizAnswered(id)
}

func (s *Store) GenerateQuizFromSubject(title string) ([]Question, error) {
	subjectId, ok := s.subjectTextId[title]

	if !ok {
		return nil, fmt.Errorf("%w", ErrInvalidObjective)
	}

	return s.generateQuiz(s.questionSubjectId[subjectId])
}

func (s *Store) GenerateQuizFromExam(objective float64) ([]Question, error) {
	examId, ok := s.examCodeId[objective]
	if !ok {
		return nil, fmt.Errorf("%w", ErrInvalidObjective)
	}

	return s.generateQuiz(s.questionExamId[examId])
}

func (s *Store) GenerateQuizFromAuto() ([]Question, error) {
	return s.cidrGenerator.createQuiz()
}

func (s *Store) generateQuiz(questions []Question) ([]Question, error) {
	q := questToQuizer(questions)

	info := createQuizerInfo(q, s.scores)
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

func (s *Store) GetExamCodeById(objective float64) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	examId, ok := s.examCodeId[objective]
	if !ok {
		return 0, fmt.Errorf("%w: %.1f", ErrInvalidObjective, objective)
	}
	return examId, nil
}

func (s *Store) GetSubjectIdByTitle(title string) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subId, ok := s.subjectTextId[title]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrInvalidObjective, title)
	}
	return subId, nil
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
