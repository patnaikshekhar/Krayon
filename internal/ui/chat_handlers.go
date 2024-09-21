package ui

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) chatRequestHandler() tea.Cmd {
	return func() tea.Msg {
		for {
			history := <-m.chatRequestCh
			modelCtx := context.Background()
			finalResponse, deltaCh, err := m.provider.Chat(modelCtx, m.profile.Model, 0, history, nil)
			if err != nil {
				log.Printf("error: %v", err)
				return nil
			}

			for delta := range deltaCh {
				m.chatResponseCh <- delta
			}

			m.chatResponseCh <- "<done>"

			<-finalResponse
		}
	}
}

func (m model) chatResponseHandler() tea.Cmd {
	return func() tea.Msg {
		delta := <-m.chatResponseCh
		return ChatDelta(delta)
	}
}
