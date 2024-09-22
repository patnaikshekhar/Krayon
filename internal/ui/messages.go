package ui

import "krayon/internal/llm"

type ChatDelta string

type includeResultMsg struct {
	err        error
	newContext string
	newSources []llm.Source
	path       string
}

type FoldersAndFiles []string
