package services

import (
	"dorzi/server/db/quiz"
)

const OneStar = 1
const TwoStar = 2
const ThreeStar = 3

const OneStarTotal int = 50
const TwoStarTotal int = 75
const ThreeStarTotal int = 100

const OneStarPerc float64 = 0.7
const TwoStarPerc float64 = 0.8
const ThreeStarPerc float64 = 0.9

const OneStarTerm float64 = 0.5
const TwoStarTerm float64 = 0.75
const ThreeStarTerm float64 = 1.00

var ratings = []StarRating{
	{ThreeStar, ThreeStarTotal, ThreeStarPerc, ThreeStarTerm},
	{TwoStar, TwoStarTotal, TwoStarPerc, TwoStarTerm},
	{OneStar, OneStarTotal, OneStarPerc, OneStarTerm},
}

func PortAsCreatePortsParam(ports []Port) quiz.CreatePortsParams {
	params := quiz.CreatePortsParams{}

	for _, port := range ports {
		params.Abbreviations = append(params.Abbreviations, port.Abbreviation)
		params.Ports = append(params.Ports, int32(port.Port))
		params.Protocols = append(params.Protocols, port.Protocol)
	}

	return params
}

func makeQuizRow(row quiz.GetQuestionsRow) QuizRow {
	return QuizRow{
		questionId:  row.ID,
		objectiveId: row.ObjectiveID.Int32,
	}
}

func makeBatchQuizQuestion(vals []QuizRow) quiz.BatchQuizQuestionParams {
	params := quiz.BatchQuizQuestionParams{}

	for _, val := range vals {
		// log.Printf("objective id %d", val.questionId)
		params.ObjectiveIds = append(params.ObjectiveIds, val.objectiveId)
		params.QuestionIds = append(params.QuestionIds, val.questionId)
		params.Status = append(params.Status, "pending")
	}
	return params
}

func getStarRating(stat Stat) int {
	for _, rating := range ratings {
		if rating.isRating(stat) {
			return rating.rating
		}
	}
	return 0
}
