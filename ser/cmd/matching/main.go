package main

import (
	"context"
	"dorzi/server/db"
	"dorzi/server/services"
	"encoding/json"
	"log"
	"os"
)

func main() {
	content, err := os.ReadFile("network_plus_ports.json")

	if err != nil {
		log.Fatalf("FILE NOT FOUND")
	}

	var ports []services.Port

	err = json.Unmarshal([]byte(content), &ports)

	db.Init("../../.env")

	quizService := services.NewQuizService()
	err = quizService.CreatePorts(context.Background(), ports)
	if err != nil {
		log.Fatalf("saving to db %v", err)
	}

}
