package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var listStyle = lipgloss.NewStyle().Margin(1, 2)

func (m model) View() string {
	if m.modePickfile {
		return listStyle.Render(m.fileList.View())
	}

	var errorMessageStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9"))

	var contextItemsStyle = lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("13"))

	em := ""
	if m.errorMessage != nil {
		em = m.errorMessage.Error()
	}

	contextItemsText := ""
	if m.readingContext {
		contextItemsText = m.readingContextSpinner.View()
	} else {
		if len(m.contextItems) > 0 {
			contextItemsText = contextItemsStyle.Render("Included:")
		}
		for _, item := range m.contextItems {
			contextItemsText += fmt.Sprintf(" %s", item)
		}
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n(esc to quit) %s\n%s",
		m.viewport.View(),
		m.userInput.View(),
		errorMessageStyle.Render(em),
		contextItemsText,
	)
}
