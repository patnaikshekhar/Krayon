package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Include(userInput string) (string, error) {

	context := ""
	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		return "", fmt.Errorf("A file name or directory must be provided\n")
	}

	path := userInputParts[1]
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			ct, err := readContent(p)
			if err == nil {
				context += ct
			}

			return nil
		})
	} else {
		ct, err := readContent(path)
		if err == nil {
			context += ct
		}
	}

	return context, nil
}

func readContent(fileName string) (string, error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("```%s\n%s\n```", fileName, string(contents)), nil
}
