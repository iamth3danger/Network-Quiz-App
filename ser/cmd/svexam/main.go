package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/services"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	content, err := os.ReadFile("n10-009-objectives.json")

	if err != nil {
		log.Fatalf("FILE NOT FOUND")
	}

	var exams []services.ExamObject

	err = json.Unmarshal([]byte(content), &exams)

	if err != nil {
		log.Fatalf("UNABLE TO TURN INTO JSON")
	}

	db.Init("../../.env")

	quizService := services.NewQuizService()

	for _, exam := range exams {
		err := quizService.SaveExamObj(context.Background(), exam)
		if err != nil {
			fmt.Errorf("COULD NOT SAVE exam!!")
		}
	}

}
