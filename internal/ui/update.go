package ui

import (
	"fmt"
	"krayon/internal/commands"
	"krayon/internal/llm"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/philistino/teacup/markdown"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.fileList.SetSize(msg.Width-h, msg.Height-v)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.userInput.Width = msg.Width
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			if m.focusIndex == 1 {
				if m.questionHistoryIndex > 0 {
					m.questionHistoryIndex--
					m.userInput.SetValue(m.questionHistory[m.questionHistoryIndex])
				}
			}
		case tea.KeyDown:
			if m.focusIndex == 1 {
				if m.questionHistoryIndex < len(m.questionHistory)-1 {
					m.questionHistoryIndex++
					m.userInput.SetValue(m.questionHistory[m.questionHistoryIndex])
				} else {
					m.userInput.Reset()
				}
			}
		case tea.KeyEnter:
			if m.modePickfile {
				m.modePickfile = false
				selectedItem := m.fileList.SelectedItem().(item)
				m.readingContext = true
				return m, tea.Batch(
					m.readingContextSpinner.Tick,
					handleInclude("/include "+selectedItem.title),
					m.userInput.Focus(),
				)
			}

			if m.focusIndex == 1 {
				if m.errorMessage != nil {
					m.errorMessage = nil
				}
				// Check user action
				userInput := strings.Trim(m.userInput.Value(), "\n")

				// Check if plugin mentioned?
				for _, plugin := range m.plugins {
					if strings.Contains(userInput, "@"+plugin.Name) {
						// Call plugin
						m.history = append(m.history, llm.Message{
							Role:       "plugin",
							PluginName: "@" + plugin.Name,
							Content: []llm.Content{
								{
									Text:        "",
									ContentType: "text",
								},
							},
						})
						m.userInput.Reset()
						return m, tea.Batch(
							m.ExecutePlugin(plugin, userInput),
							m.pluginResponseHandler(),
						)
					}
				}

				if userInput == "/exit" || userInput == "/quit" {
					return m, tea.Quit
				}

				if strings.HasPrefix(userInput, "/include") {
					m.userInput.Reset()

					components := strings.Split(userInput, " ")
					if len(components) < 2 {
						m.modePickfile = true
						m.fileList, cmd = m.fileList.Update(msg)
						return m, cmd
					}

					m.readingContext = true
					return m, tea.Batch(
						m.readingContextSpinner.Tick,
						handleInclude(userInput),
					)
				}

				if strings.HasPrefix(userInput, "/clear") {
					m.history = []llm.Message{}
					m.context = ""
					m.contextItems = []string{}
					m.imageContext = []llm.Source{}
					m.viewport.SetContent(m.renderHistory())
					m.userInput.Reset()
					return m, nil
				}

				if strings.HasPrefix(userInput, "/save-history") {
					err := commands.SaveHistory(userInput, m.history, m.context)
					if err != nil {
						m.errorMessage = err
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
						m.errorMessage = err
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
						m.errorMessage = err
						m.viewport.SetContent(m.renderHistory())
						m.userInput.Reset()
						return m, nil
					}
					log.Printf("History loaded")

					m.viewport.SetContent(m.renderHistory())
					m.userInput.Reset()
					return m, nil
				}

				commands.LogUserInput(userInput)
				m.questionHistory = append(m.questionHistory, userInput)

				if m.context != "" {
					userInput = fmt.Sprintf("%s\n---Context---\n%s", userInput, m.context)
					m.context = ""
					m.contextItems = []string{}
				}

				content := []llm.Content{
					{
						Text:        userInput,
						ContentType: "text",
					},
				}

				if m.imageContext != nil {
					for _, m := range m.imageContext {
						content = append(content, llm.Content{
							ContentType: "image",
							Source:      &m,
						})
					}
					m.imageContext = []llm.Source{}
				}

				m.history = append(m.history, llm.Message{
					Role:    "user",
					Content: content,
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
	case PluginDelta:
		if msg == "<done>" {
			return m, nil
		}
		m.history[len(m.history)-1].Content[0].Text += string(msg)
		m.viewport.SetContent(m.renderHistory())
		m.viewport.GotoBottom()
		return m, m.pluginResponseHandler()

	case includeResultMsg:
		if msg.err != nil {
			m.errorMessage = msg.err
			return m, nil
		}

		m.context += msg.newContext
		m.imageContext = append(m.imageContext, msg.newSources...)
		m.contextItems = append(m.contextItems, msg.path)
		m.readingContext = false
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.readingContext {
			m.readingContextSpinner, cmd = m.readingContextSpinner.Update(msg)
		}
		return m, cmd
	}

	if m.modePickfile {
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, cmd
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

	var pluginStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("9"))

	history := ""
	for _, h := range m.history {
		role := ""
		if h.Role == "assistant" {
			role = aiStyle.Render("  ֍ AI")
		} else if h.Role == "user" {
			role = userStyle.Render("  YOU")
		} else if h.Role == "plugin" {
			role = pluginStyle.Render(fmt.Sprintf("  %s says:", strings.ToUpper(h.PluginName)))
		}

		// Strip content of --context---
		contentParts := strings.Split(h.Content[0].Text, "\n---Context---\n")

		contentMarkdown, err := markdown.RenderMarkdown(80, contentParts[0])
		if err != nil {
			log.Printf("Error rendering markdown: %s", err)
		}
		history += fmt.Sprintf("%s %s\n", role, contentMarkdown)
	}

	if history == "" {
		history = "Welcome to Krayon!\n\n"
	}

	log.Printf("Rendering history %+v", m.history)

	return history
}

func handleInclude(userInput string) tea.Cmd {
	return func() tea.Msg {
		newContext, newSources, path, err := commands.Include(userInput)
		if err != nil {
			return includeResultMsg{
				err: err,
			}
		}

		return includeResultMsg{
			newContext: newContext,
			newSources: newSources,
			path:       path,
		}
	}
}
