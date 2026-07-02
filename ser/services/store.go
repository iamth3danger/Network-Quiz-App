package services

import "sync"

type Store struct {
	examIdCode   map[int32]float64
	examCodeId   map[float64]int
	questions    map[int32]Question
	scores       map[int32]Scores
	quiz         map[int32]Quiz
	ports        map[int32]Port
	portscores   map[int32]PortScore
	terms        map[int32]Term
	matchingQuiz map[int32]MatchingQuiz
	mu           sync.RWMutex
}
