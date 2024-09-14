package commands

import (
	"fmt"
	"krayon/internal/llm"
	"os"
	"strings"
)

func Save(userInput string, history []llm.Message) error {
	filePath := ""
	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		// Save to default Krayon directory
		return fmt.Errorf("Please provide a path to save history to")
	}

	filePath = userInputParts[1]

	aiResponse := history[len(history)-1].Content

	err := os.WriteFile(filePath, []byte(aiResponse[0].Text), os.ModePerm)
	if err != nil {
		return err
	}

	return nil

}
