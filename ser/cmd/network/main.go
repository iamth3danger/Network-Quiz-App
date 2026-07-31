package main

import (
	"dorzi/server/services"
	"fmt"
)

func main() {

	question, err := services.IPCidrToFirstLastHostQuestion()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("question: %v\n", question)
}
