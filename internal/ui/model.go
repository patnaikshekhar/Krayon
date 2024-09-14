package ui

import (
	"krayon/internal/config"
	"krayon/internal/llm"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

type model struct {
	history   []llm.Message
	context   string
	userInput textinput.Model
	provider  llm.Provider
	profile   *config.Profile

	chatRequestCh  chan []llm.Message
	chatResponseCh chan string

	viewport   viewport.Model
	focusIndex int // 0: viewport, 1: userInput
}

func NewModel() (*model, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	profile := cfg.GetProfile(cfg.DefaultProfile)
	provider, err := llm.GetProvider(profile.Provider, profile.ApiKey)
	if err != nil {
		return nil, err
	}

	ta := textinput.New()
	ta.Prompt = "┃ "
	ta.Placeholder = "Your question here..."
	ta.Focus()
	ta.CharLimit = 280
	ta.ShowSuggestions = true
	ta.SetSuggestions([]string{"/include", "/exit", "/explain", "/clear", "/save-history", "/load-history", "/quit", "/save"})
	ta.KeyMap.AcceptSuggestion.SetEnabled(true)
	ta.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("right"))
	ta.KeyMap.PrevSuggestion = key.NewBinding(key.WithKeys("up"))
	ta.KeyMap.NextSuggestion = key.NewBinding(key.WithKeys("down"))

	vp := viewport.New(80, 20)

	return &model{
		userInput:      ta,
		provider:       provider,
		profile:        profile,
		chatRequestCh:  make(chan []llm.Message),
		chatResponseCh: make(chan string),
		viewport:       vp,
		focusIndex:     1,
	}, nil
}
