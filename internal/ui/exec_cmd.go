package ui

import (
	"fmt"
	"krayon/internal/config"
	"krayon/internal/plugins"
	"log"
	"os"
	"os/exec"
	"path"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) ExecutePlugin(plugin plugins.ManifestPlugin, userPrompt string) tea.Cmd {
	return func() tea.Msg {
		krayonHome, _ := config.GetConfigBasePath()
		// workingDir, _ := os.Getwd()
		pluginBinaryName := fmt.Sprintf("%s_%s", plugin.Name, plugin.Version)
		pluginPath := path.Join(krayonHome, "plugins", pluginBinaryName)
		execCmd := exec.Command(pluginPath, "--socket", m.socketPath)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		// execCmd.Path = workingDir
		err := execCmd.Start()
		if err != nil {
			log.Printf("Error executing plugin %s: %s", plugin.Name, err)
			return nil
		}

		m.pluginStartChannel <- plugins.RequestInfo{
			Question: userPrompt,
			History:  m.history,
			Context:  m.context,
		}

		err = execCmd.Wait()
		if err != nil {
			log.Printf("Error executing plugin %s: %s", plugin.Name, err)
			return nil
		}

		m.pluginResponseCh <- "<done>"

		return nil
	}
}
