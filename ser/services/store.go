package services

import (
	"cmp"
	"fmt"
	"math/rand"
	"slices"
	"sync"
)

type Store struct {
	examIdCode      map[int32]float64
	examCodeId      map[float64]int32
	questions       map[int32]Question
	questionExamId  map[int32][]Question
	scores          map[int32]Score
	objectiveScores map[int32]ObjectiveScore
	quiz            map[int32][]Quiz
	ports           map[int32]Port
	portScores      map[int32]PortScore
	terms           map[int32]Term
	matchingQuiz    []MatchingQuiz
	mu              sync.RWMutex
}

func (s *Store) FetchQuestion(objective float64) (Question, error) {
	examId, err := s.getExamCodeById(objective)

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

	rand.Shuffle(len(pending), func(i, j int) {
		pending[i], pending[j] = pending[j], pending[i]
	})

	err = s.setQuizQuestionReviewed(examId, pending[0].ID)

	if err != nil {
		return Question{}, err
	}

	return s.getQuestionById(pending[0].QuestionID)
}

func (s *Store) GetQuizAnswered(objective float64) (QuizInfo, error) {
	examId, err := s.getExamCodeById(objective)

	if err != nil {
		return QuizInfo{}, err
	}

	quizQuestions, err := s.getQuizByExamId(examId)
	count := 0
	for _, q := range quizQuestions {
		if q.Status == "answered" {
			count += 1
		}
	}

	return QuizInfo{Count: count}, nil
}

func (s *Store) GenerateQuiz(objective float64) ([]Question, error) {
	const MaxQuestions = 10
	examId := s.examCodeId[objective]
	questions := s.questionExamId[examId]

	var newQuestions []Question
	var reviewedQuestions []Question
	var scores []Score
	scoreQuestion := make(map[int32]Question)

	for _, q := range questions {
		score, ok := s.scores[q.ID]
		if ok {
			reviewedQuestions = append(reviewedQuestions, q)
			scores = append(scores, score)
			scoreQuestion[score.ID] = q
		} else {
			newQuestions = append(newQuestions, q)
		}
	}

	slices.SortFunc(scores, func(a, b Score) int {
		if n := cmp.Compare(a.Streak, b.Streak); n != 0 {
			return n
		}
		aRatio := float64(a.RightCount) / max(float64(a.WrongCount), 1)
		bRatio := float64(b.RightCount) / max(float64(b.WrongCount), 1)
		return cmp.Compare(aRatio, bRatio)
	})

	rand.Shuffle(len(newQuestions), func(i, j int) {
		newQuestions[i], newQuestions[j] = newQuestions[j], newQuestions[i]
	})

	if len(reviewedQuestions) == 0 {
		return newQuestions[0:10], nil

	} else if len(newQuestions) == 0 {
		return reviewedQuestions[0:10], nil
	} else {
		var quizVal []Question
		numReviewQuestions := min(len(reviewedQuestions), 5)
		streakCount := 0
		for _, s := range scores[0:5] {
			streakCount += int(s.Streak)
		}

		streakRatio := float64(streakCount) / float64(numReviewQuestions)
		var numQuestions int
		switch {
		case streakCount == 0:
			numQuestions = min(len(reviewedQuestions), 5)
		case streakRatio <= 1.0:
			numQuestions = min(len(reviewedQuestions), 3)
		case streakRatio <= 1.5:
			numQuestions = min(len(reviewedQuestions), 2)
		default:
			numQuestions = 1
		}

		quizVal = append(quizVal, reviewedQuestions[0:numQuestions]...)
		numNewQuestions := min(len(newQuestions), MaxQuestions-numQuestions)
		quizVal = append(quizVal, newQuestions[0:numNewQuestions]...)

		return quizVal, nil
	}

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
			return nil, fmt.Errorf("fetching exam id %w", err)
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

func (s *Store) UpdateQuestion(question QuestionRequest) error {
	s.updateScore(question)
	err := s.updateObjectiveScore(question)
	if err != nil {
		return err
	}
	return s.setQuizQuestionAnswered(int32(question.QuestionID), int32(question.ExamObjectiveID))
}

func (s *Store) getQuizByExamId(examId int32) ([]Quiz, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	quiz, ok := s.quiz[examId]
	if !ok {
		return nil, fmt.Errorf("invalid exam id: %d", examId)
	}

	return quiz, nil
}

func (s *Store) setQuizQuestionStatus(examId int32, questionID int32, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	quiz, ok := s.quiz[examId]
	if !ok {
		return fmt.Errorf("invalid exam id: %d", examId)
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

func (s *Store) getExamCodeById(objective float64) (int32, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	examId, ok := s.examCodeId[objective]
	if !ok {
		return 0, fmt.Errorf("invalid exam objective: %.1f", objective)
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

func (s *Store) getScoreById(questionId int32) (Score, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	score, ok := s.scores[questionId]
	if !ok {
		return Score{}, fmt.Errorf("invalid question id: %d", questionId)
	}
	return score, nil
}

func (s *Store) updateScore(question QuestionRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	score, ok := s.scores[int32(question.QuestionID)]
	if !ok {
		if question.Correct {
			s.scores[int32(question.QuestionID)] = Score{
				QuestionID: int32(question.QuestionID),
				RightCount: 1,
				WrongCount: 0,
				Streak:     1,
			}
		} else {
			s.scores[int32(question.QuestionID)] = Score{
				QuestionID: int32(question.QuestionID),
				RightCount: 0,
				WrongCount: 1,
				Streak:     0,
			}
		}
		return
	}

	if question.Correct {
		score.RightCount += 1
		score.Streak += 1
	} else {
		score.WrongCount += 1
		score.Streak = 0
	}

	s.scores[int32(question.QuestionID)] = score
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
