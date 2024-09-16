package commands

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dslipak/pdf"
)

func Include(userInput string) (string, string, error) {

	context := ""
	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		return "", "", fmt.Errorf("A file name or directory must be provided\n")
	}

	path := userInputParts[1]
	info, err := os.Stat(path)
	if err != nil {
		return "", "", err
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

	return context, path, nil
}

func readContent(fileName string) (string, error) {
	extn := filepath.Ext(fileName)
	if extn == ".pdf" {
		pdfReader, err := pdf.Open(fileName)
		if err != nil {
			return "", err
		}

		reader, err := pdfReader.GetPlainText()
		if err != nil {
			return "", err
		}

		b := bytes.NewBuffer([]byte{})
		_, err = io.Copy(b, reader)
		if err != nil {
			return "", err
		}

		return b.String(), nil
	}

	contents, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("```%s\n%s\n```", fileName, string(contents)), nil
}
