package commands

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"krayon/internal/llm"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dslipak/pdf"
	"github.com/playwright-community/playwright-go"
)

func Include(userInput string) (string, []llm.Source, string, error) {

	context := ""
	var sources []llm.Source

	userInputParts := strings.Split(userInput, " ")
	if len(userInputParts) < 2 {
		return "", nil, "", fmt.Errorf("A file name, directory or url must be provided\n")
	}

	if strings.HasPrefix(userInputParts[1], "http") {
		// Download content
		pageContents, err := getPageContents(userInputParts[1])
		if err != nil {
			return "", nil, "", err
		}

		return fmt.Sprintf("```%s\n%s\n```", userInputParts[1], string(pageContents)), nil, userInputParts[1], nil
	}

	path := userInputParts[1]
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, "", err
	}

	if info.IsDir() {
		filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			ct, source, err := readContent(p)
			if err == nil {
				if ct != "" {
					context += ct
				} else if source != nil {
					sources = append(sources, *source)
				}
			}

			return nil
		})
	} else {
		ct, source, err := readContent(path)
		if err == nil {
			if ct != "" {
				context += ct
			} else if source != nil {
				sources = append(sources, *source)
			}
		}
	}

	return context, sources, path, nil
}

func readContent(fileName string) (string, *llm.Source, error) {
	extn := filepath.Ext(fileName)
	if extn == ".pdf" {
		pdfReader, err := pdf.Open(fileName)
		if err != nil {
			return "", nil, err
		}

		reader, err := pdfReader.GetPlainText()
		if err != nil {
			return "", nil, err
		}

		b := bytes.NewBuffer([]byte{})
		_, err = io.Copy(b, reader)
		if err != nil {
			return "", nil, err
		}

		return b.String(), nil, nil
	}

	contents, err := os.ReadFile(fileName)
	if err != nil {
		return "", nil, err
	}

	if extn == ".jpeg" || extn == ".jpg" || extn == ".png" {
		mediaType := map[string]string{
			".jpeg": "image/jpeg",
			".jpg":  "image/jpeg",
			".png":  "image/png",
		}
		// Convert to base64 format and return
		return "", &llm.Source{
			Type:      "base64",
			Data:      base64.StdEncoding.EncodeToString(contents),
			MediaType: mediaType[extn],
		}, nil
	}

	return fmt.Sprintf("```%s\n%s\n```", fileName, string(contents)), nil, nil
}

func getPageContents(path string) (string, error) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install Playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return "", err
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return "", err
	}
	page, err := browser.NewPage()
	if err != nil {
		return "", err
	}
	if _, err = page.Goto(path); err != nil {
		return "", err
	}

	result := ""
	allText, err := page.Locator("body").AllInnerTexts()
	if err != nil {
		return "", err
	}
	for _, text := range allText {
		result += text
	}

	return result, nil
}
