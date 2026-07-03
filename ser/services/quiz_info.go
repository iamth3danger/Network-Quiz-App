package services

import (
	"cmp"
	"math/rand"
	"slices"
)

const MaxQuestions = 10
const MaxReviewSample = 10

type Quizer interface {
	getId() int32
}

type Scorer interface {
	getRight() int32
	getWrong() int32
	getStreak() int32
	getId() int32
}

func toQuestions(quizables []Quizer) []Question {
	var questions []Question
	for _, q := range quizables {
		questions = append(questions, *(q.(*Question)))
	}
	return questions
}

func toPorts(quizables []Quizer) []Port {
	var questions []Port
	for _, q := range quizables {
		questions = append(questions, *(q.(*Port)))
	}
	return questions
}

func questToQuizer(questions []Question) []Quizer {
	var quizables []Quizer
	for _, q := range questions {
		quizables = append(quizables, &q)
	}
	return quizables
}

func portToQuizer(ports []Port) []Quizer {
	var quizables []Quizer
	for _, p := range ports {
		quizables = append(quizables, &p)
	}
	return quizables
}

type QuizInfo struct {
	newQuizer      []Quizer
	reviewedQuizer []Quizer
	scores         []Scorer
	scoreQuizer    map[int32]Quizer
}

func (q *QuizInfo) rearrangeArrs() {
	q.sortScorer()
	q.shuffleQuizer()
}

func (q *QuizInfo) sortScorer() {
	slices.SortFunc(q.scores, func(a, b Scorer) int {
		if n := cmp.Compare(a.getStreak(), b.getStreak()); n != 0 {
			return n
		}
		aRatio := float64(a.getRight()) / max(float64(a.getWrong()), 1)
		bRatio := float64(b.getRight()) / max(float64(b.getWrong()), 1)
		return cmp.Compare(aRatio, bRatio)
	})
}

func (q *QuizInfo) shuffleQuizer() {
	rand.Shuffle(len(q.newQuizer), func(i, j int) {
		q.newQuizer[i], q.newQuizer[j] = q.newQuizer[j], q.newQuizer[i]
	})
}

func (q *QuizInfo) make(getReviewableScore func() []Scorer) []Quizer {
	if len(q.reviewedQuizer) == 0 {
		return q.newQuizer[0:10]

	} else if len(q.newQuizer) == 0 {
		return q.reviewedQuizer[0:10]
	} else {
		var quizVal []Quizer

		scores := getReviewableScore()
		for _, s := range scores {
			quizVal = append(quizVal, q.scoreQuizer[s.getId()])
		}
		numNewQuestions := min(len(q.newQuizer), MaxQuestions-len(scores))
		quizVal = append(quizVal, q.newQuizer[0:numNewQuestions]...)
		return quizVal
	}
}

func (q *QuizInfo) getQuizReviewableScores() []Scorer {
	numReviewQuestions := min(len(q.reviewedQuizer), MaxReviewSample)
	streakCount := 0
	for _, s := range q.scores[0:MaxReviewSample] {
		streakCount += int(s.getStreak())
	}

	streakRatio := float64(streakCount) / float64(numReviewQuestions)
	var reviewNum int
	switch {
	case streakCount == 0:
		reviewNum = min(len(q.reviewedQuizer), MaxReviewSample)
	case streakRatio <= 1.0:
		reviewNum = min(len(q.reviewedQuizer), 3)
	case streakRatio <= 1.5:
		reviewNum = min(len(q.reviewedQuizer), 2)
	default:
		reviewNum = 1
	}

	return q.scores[0:reviewNum]
}

func (q *QuizInfo) getMatchingReviewableScores() []Scorer {
	wrongCount := 0
	for _, s := range q.scores[0:MaxReviewSample] {
		if s.getWrong() > s.getRight() {
			wrongCount += 1
		}
	}

	return q.scores[0:wrongCount]
}

func (q *QuizInfo) makeQuiz() []Quizer {
	return q.make(q.getQuizReviewableScores)
}

func (q *QuizInfo) makeMatching() []Quizer {
	return q.make(q.getMatchingReviewableScores)
}

func createQuizerInfo(quizables []Quizer, scoreMap map[int32]Scorer) QuizInfo {

	info := QuizInfo{
		scoreQuizer: make(map[int32]Quizer),
	}

	for _, q := range quizables {
		score, ok := scoreMap[q.getId()]
		if ok {
			info.reviewedQuizer = append(info.reviewedQuizer, q)
			info.scores = append(info.scores, score)
			info.scoreQuizer[score.getId()] = q
		} else {
			info.newQuizer = append(info.newQuizer, q)
		}
	}

	return info
}
