package ui

import (
	"fmt"
	"krayon/internal/commands"
	"krayon/internal/llm"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/philistino/teacup/markdown"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.userInput.Width = msg.Width
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.focusIndex == 1 {

				// Check user action
				userInput := strings.Trim(m.userInput.Value(), "\n")
				if userInput == "/exit" || userInput == "/quit" {
					return m, tea.Quit
				}

				if strings.HasPrefix(userInput, "/include") {
					newContext, err := commands.Include(userInput)
					if err != nil {
						fmt.Printf("Error occured: %s", err)
						return m, nil
					}

					m.context += newContext
					m.userInput.Reset()
					return m, nil
				}

				if strings.HasPrefix(userInput, "/clear") {
					m.history = []llm.Message{}
					m.context = ""
					m.viewport.SetContent(m.renderHistory())
					m.userInput.Reset()
					return m, nil
				}

				if strings.HasPrefix(userInput, "/save-history") {
					err := commands.SaveHistory(userInput, m.history, m.context)
					if err != nil {
						fmt.Printf("Error saving history: %s", err)
						m.viewport.SetContent(m.renderHistory())
						m.userInput.Reset()
						return m, nil
					}

					m.viewport.SetContent(m.renderHistory())
					m.userInput.Reset()
					return m, nil
				}

				if strings.HasPrefix(userInput, "/save") {
					err := commands.Save(userInput, m.history)
					if err != nil {
						fmt.Printf("Error saving response from AI: %s", err)
						m.userInput.Reset()
						return m, nil
					}

					m.userInput.Reset()
					return m, nil
				}

				if strings.HasPrefix(userInput, "/load-history") {

					var err error
					m.history, m.context, err = commands.LoadHistory(userInput)
					if err != nil {
						fmt.Printf("Error loading from history: %s", err)
						m.viewport.SetContent(m.renderHistory())
						m.userInput.Reset()
						return m, nil
					}
					log.Printf("History loaded")

					m.viewport.SetContent(m.renderHistory())
					m.userInput.Reset()
					return m, nil
				}

				if m.context != "" {
					userInput = fmt.Sprintf("%s\n---Context---\n%s", userInput, m.context)
					m.context = ""
				}
				m.history = append(m.history, llm.Message{
					Role: "user",
					Content: []llm.Content{
						{
							Text:        userInput,
							ContentType: "text",
						},
					},
				})

				m.userInput.Reset()
				m.chatRequestCh <- m.history

				m.history = append(m.history, llm.Message{
					Role: "assistant",
					Content: []llm.Content{
						{
							Text:        "",
							ContentType: "text",
						},
					},
				})

				m.viewport.SetContent(m.renderHistory())
				m.viewport.GotoBottom()
				return m, m.chatResponseHandler()
			}
		case tea.KeyTab:
			if m.focusIndex == 0 {
				m.focusIndex = 1
				m.userInput.Focus()
			} else {
				m.focusIndex = 0
				m.userInput.Blur()
			}
			return m, nil
		}
	case ChatDelta:
		if msg == "<done>" {
			return m, nil
		}

		m.history[len(m.history)-1].Content[0].Text += string(msg)
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()
		return m, m.chatResponseHandler()
	}

	if m.focusIndex == 0 {
		m.viewport, cmd = m.viewport.Update(msg)
	} else {
		m.userInput, cmd = m.userInput.Update(msg)
	}
	return m, cmd
}

func (m *model) renderHistory() string {
	var userStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	var aiStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5"))

	history := ""
	for _, h := range m.history {
		role := ""
		if h.Role == "assistant" {
			role = aiStyle.Render("  ֍ AI")
		} else if h.Role == "user" {
			role = userStyle.Render("  YOU")
		}

		contentMarkdown, err := markdown.RenderMarkdown(80, h.Content[0].Text)
		if err != nil {
			log.Printf("Error rendering markdown: %s", err)
		}
		history += fmt.Sprintf("%s %s\n", role, contentMarkdown)
	}

	if history == "" {
		history = "Welcome to Krayon!\n\n"
	}

	return history
}
