package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
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
	if len(m.contextItems) > 0 {
		contextItemsText = contextItemsStyle.Render("Included:")
	}
	for _, item := range m.contextItems {
		contextItemsText += fmt.Sprintf(" %s", item)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n(esc to quit) %s\n%s",
		m.viewport.View(),
		m.userInput.View(),
		errorMessageStyle.Render(em),
		contextItemsText,
	)
}
