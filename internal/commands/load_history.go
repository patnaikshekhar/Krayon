package commands

import (
	"encoding/json"
	"krayon/internal/config"
	"krayon/internal/llm"
	"os"
	"path"
	"strings"
)

func LoadHistory(userInput string) ([]llm.Message, string, error) {
	historyPath := ""
	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		// Save to default Krayon directory
		basePath, err := config.GetConfigBasePath()
		if err != nil {
			return nil, "", err
		}

		historyPath = path.Join(basePath, "history.json")

	} else {
		historyPath = userInputParts[1]
	}

	content, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, "", err
	}

	hf := historyFormat{}
	err = json.Unmarshal(content, &hf)
	if err != nil {
		return nil, "", err
	}

	return hf.History, hf.Context, nil

}
