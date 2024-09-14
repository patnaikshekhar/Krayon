package ui

import (
	"fmt"
)

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.viewport.View(),
		m.userInput.View(),
		"(esc to quit)",
	)
}
