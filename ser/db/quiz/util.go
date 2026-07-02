package quiz

import (
	"github.com/jackc/pgx/v5/pgtype"
)

func MakeQuizParam(question GetQuestionsRow) CreateQuizQuestionParams {
	return CreateQuizQuestionParams{
		QuestionID:  pgtype.Int4{Int32: question.ID, Valid: true},
		ObjectiveID: question.ObjectiveID,
		Status:      NullQuestionStatus{QuestionStatus: QuestionStatusPending, Valid: true},
	}
}
