package commands

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"krayon/internal/llm"
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
		pageContents, sources, err := getPageContents(userInputParts[1])
		if err != nil {
			return "", nil, "", err
		}

		return fmt.Sprintf("```%s\n%s\n```\n Screenshot taken.\n", userInputParts[1], string(pageContents)), sources, userInputParts[1], nil
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

func getPageContents(path string) (string, []llm.Source, error) {
	err := playwright.Install()
	if err != nil {
		return "", nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		return "", nil, err
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return "", nil, err
	}
	page, err := browser.NewPage()
	if err != nil {
		return "", nil, err
	}
	if _, err = page.Goto(path); err != nil {
		return "", nil, err
	}

	result := ""
	allText, err := page.Locator("body").AllInnerTexts()
	if err != nil {
		return "", nil, err
	}
	for _, text := range allText {
		result += text
	}

	fullpage := true
	screenShot, _ := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: &fullpage,
		Clip: &playwright.Rect{
			X:      0,
			Y:      0,
			Width:  1920,
			Height: 7000,
		},
	})

	sources := []llm.Source{}
	sources = append(sources, llm.Source{
		Type:      "base64",
		Data:      base64.StdEncoding.EncodeToString(screenShot),
		MediaType: "image/png",
	})

	if err = browser.Close(); err != nil {
		return "", nil, err
	}
	if err = pw.Stop(); err != nil {
		return "", nil, err
	}

	return result, sources, nil
}
