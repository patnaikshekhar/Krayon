package ui

import (
	"krayon/internal/commands"
	"krayon/internal/config"
	"krayon/internal/llm"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	history []llm.Message

	context      string
	contextItems []string
	imageContext []llm.Source

	userInput textinput.Model

	provider llm.Provider
	profile  *config.Profile

	errorMessage error

	chatRequestCh  chan []llm.Message
	chatResponseCh chan string

	viewport   viewport.Model
	focusIndex int // 0: viewport, 1: userInput

	questionHistory      []string
	questionHistoryIndex int

	readingContextSpinner spinner.Model
	readingContext        bool

	modePickfile bool
	fileList     list.Model
}

func NewModel(selectedProfile string) (*model, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if selectedProfile == "" {
		selectedProfile = cfg.DefaultProfile
	}

	profile := cfg.GetProfile(selectedProfile)

	provider, err := llm.GetProvider(profile.Provider, profile.ApiKey, profile.Stream)
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

	includeContextSpinner := spinner.New()
	includeContextSpinner.Spinner = spinner.Points
	includeContextSpinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#0000FF"))

	fileList := list.New(getAllFiles(), list.NewDefaultDelegate(), 0, 0)
	fileList.Title = "Select a file"

	vp := viewport.New(80, 20)

	userLog := commands.GetUserLog()

	return &model{
		userInput:             ta,
		provider:              provider,
		profile:               profile,
		chatRequestCh:         make(chan []llm.Message),
		chatResponseCh:        make(chan string),
		viewport:              vp,
		focusIndex:            1,
		questionHistory:       userLog,
		questionHistoryIndex:  len(userLog),
		readingContextSpinner: includeContextSpinner,
		readingContext:        false,
		fileList:              fileList,
	}, nil
}
