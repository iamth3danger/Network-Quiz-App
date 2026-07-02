package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/services"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Objective struct {
	Objectives []Term `json:"objectives"`
}

type Term struct {
	Objective string   `json:"objective"`
	Terms     []string `json:"terms"`
}

func main() {
	content, err := os.ReadFile("n10-009-terms.json")

	if err != nil {
		log.Fatalf("FILE NOT FOUND")
	}

	var terms Objective

	err = json.Unmarshal([]byte(content), &terms)

	if err != nil {
		log.Fatalf("UNABLE TO TURN INTO JSON")
	}

	db.Init("../../.env")

	ctx := context.Background()
	quizService := services.NewQuizService()
	err = quizService.Init(ctx)
	if err != nil {
		fmt.Errorf("Issue initing the db")
	}

	for _, term := range terms.Objectives {
		obj, err := strconv.ParseFloat(term.Objective, 64)
		if err != nil {
			fmt.Errorf("ERROR PARSING FLOAT %v", err)
		}
		for _, t := range term.Terms {
			err = quizService.SaveTerm(ctx, t, obj)
			if err != nil {
				fmt.Errorf("COULD NOT SAVE term!! %v", err)
			}
		}
	}

}
