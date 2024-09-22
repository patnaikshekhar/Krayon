package ui

import (
	"io/fs"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func getAllFiles() []list.Item {
	var result []list.Item

	filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == "." || d.Name() == ".." {
			return nil
		}

		if d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == ".venv") {
			return fs.SkipDir
		}

		var description string
		if d.IsDir() {
			description = "Directory"
		} else {
			description = "File"
		}

		result = append(result, item{
			title: p,
			desc:  description,
		})
		return nil
	})

	return result
}
