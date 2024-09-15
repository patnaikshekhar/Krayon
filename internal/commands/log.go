package commands

import (
	"encoding/json"
	"krayon/internal/config"
	"log"
	"os"
	"path"
)

func GetUserLog() []string {
	basePath, err := config.GetConfigBasePath()
	if err != nil {
		return []string{}
	}

	userLogPath := path.Join(basePath, "user_log.json")

	userLogBytes, err := os.ReadFile(userLogPath)
	if err != nil {
		return []string{}
	}

	var userLog []string
	err = json.Unmarshal(userLogBytes, &userLog)
	if err != nil {
		return []string{}
	}

	return userLog
}

func LogUserInput(userInput string) {
	basePath, err := config.GetConfigBasePath()
	if err != nil {
		log.Printf("Error trying to get config base path: %s", err)
		return
	}

	userLogPath := path.Join(basePath, "user_log.json")

	userLogBytes, err := os.ReadFile(userLogPath)
	var userLog []string
	if err == nil {
		log.Printf("Error trying to read user log: %s", err)
		err = json.Unmarshal(userLogBytes, &userLog)
		if err != nil {
			log.Printf("Error trying to unmarshal user log: %s", err)
			return
		}
	}

	userLog = append(userLog, userInput)
	userLogBytes, err = json.Marshal(userLog)
	if err != nil {
		log.Printf("Error trying to marshal user log: %s", err)
		return
	}

	err = os.WriteFile(userLogPath, userLogBytes, os.ModePerm)
	if err != nil {
		log.Printf("Error trying to write user log: %s", err)
		return
	}
}
