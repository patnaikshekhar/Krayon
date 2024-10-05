package ui

import tea "github.com/charmbracelet/bubbletea"

type PluginDelta string

func (m model) pluginResponseHandler() tea.Cmd {
	return func() tea.Msg {
		delta := <-m.pluginResponseCh
		return PluginDelta(delta)
	}
}
