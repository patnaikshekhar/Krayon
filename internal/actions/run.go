package actions

import (
	"krayon/internal/config"
	"krayon/internal/llm"
	"krayon/internal/plugins"
	"krayon/internal/ui"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func Run(ctx *cli.Context) error {

	selectedProfile := ctx.String("profile")

	pluginStartChannel := make(chan plugins.RequestInfo)
	pluginRespChannel := make(chan string)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if selectedProfile == "" {
		selectedProfile = cfg.DefaultProfile
	}

	profile := cfg.GetProfile(selectedProfile)

	provider, err := llm.GetProvider(profile.Provider, profile.ApiKey, profile.Stream)
	if err != nil {
		return err
	}
	socketPath := "/tmp/krayon.sock"

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Printf("fatal: %s", err)
		return err
	}
	defer f.Close()
	defer os.Remove(socketPath)

	go func() {
		err := plugins.RunPluginServer(pluginStartChannel, provider, profile, socketPath, pluginRespChannel)
		if err != nil {
			log.Printf("fatal: running unix socket %s", err)
		}
	}()

	model, err := ui.NewModel(provider, profile, pluginStartChannel, socketPath, pluginRespChannel)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}

	return nil
}
