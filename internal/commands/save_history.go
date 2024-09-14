package commands

import (
	"encoding/json"
	"krayon/internal/config"
	"krayon/internal/llm"
	"os"
	"path"
	"strings"
)

func SaveHistory(userInput string, history []llm.Message, context string) error {
	historyPath := ""
	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		// Save to default Krayon directory
		basePath, err := config.GetConfigBasePath()
		if err != nil {
			return err
		}

		historyPath = path.Join(basePath, "history.json")

	} else {
		historyPath = userInputParts[1]
	}

	h := historyFormat{
		Context: context,
		History: history,
	}

	historyBytes, err := json.Marshal(h)
	if err != nil {
		return err
	}

	err = os.WriteFile(historyPath, historyBytes, os.ModePerm)
	if err != nil {
		return err
	}

	return nil

}

type historyFormat struct {
	Context string        `json:"context"`
	History []llm.Message `json:"history"`
}
