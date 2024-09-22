package actions

import (
	"fmt"
	"krayon/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func Run(ctx *cli.Context) error {

	profile := ctx.String("profile")

	model, err := ui.NewModel(profile)
	if err != nil {
		return err
	}

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		return err
	}
	defer f.Close()

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}

	return nil
}
