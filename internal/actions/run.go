package actions

import (
	"context"
	"fmt"
	"krayon/internal/config"
	"krayon/internal/llm"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

type model struct {
	history   []llm.Message
	userInput textinput.Model
	provider  llm.Provider
	profile   *config.Profile
}

func initialModel() (*model, error) {

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	profile := cfg.GetProfile(cfg.DefaultProfile)
	provider, err := llm.GetProvider(profile.Provider, profile.ApiKey)
	if err != nil {
		return nil, err
	}

	ti := textinput.New()
	ti.Placeholder = "Your question here..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return &model{
		userInput: ti,
		provider:  provider,
		profile:   profile,
	}, nil
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.history = append(m.history, llm.Message{
				Role: "user",
				Content: []llm.Content{
					{
						Text:        m.userInput.Value(),
						ContentType: "text",
					},
				},
			})

			return m, m.chat
		}
	}

	m.userInput, cmd = m.userInput.Update(msg)
	return m, cmd
}

func (m model) chat() tea.Msg {
	modelCtx := context.Background()
	messageChan, deltaChan, err := m.provider.Chat(modelCtx, m.profile.Model, 0, m.history, nil)
	if err != nil {
		return err
	}
}

func (m model) View() string {
	return fmt.Sprintf(
		"You%s\n\n%s",
		m.userInput.View(),
		"(esc to quit)",
	) + "\n"
}

func Run(ctx *cli.Context) error {

	model, err := initialModel()
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(model).Run()
	if err != nil {
		return err
	}

	// history := []llm.Message{}
	// context := ""

	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Printf("You: ")
	// 	userInput, _ := reader.ReadString('\n')

	// 	userInput = strings.Trim(userInput, "\n")
	// 	if userInput == "/exit" || userInput == "/quit" {
	// 		break
	// 	}

	// 	if strings.HasPrefix(userInput, "/include") {
	// 		newContext, err := commands.Include(userInput)
	// 		if err != nil {
	// 			fmt.Printf("Error occured: %s", err)
	// 			continue
	// 		}

	// 		context += newContext
	// 		continue
	// 	}

	// 	if strings.HasPrefix(userInput, "/clear") {
	// 		history = []llm.Message{}
	// 		context = ""
	// 		continue
	// 	}

	// 	if strings.HasPrefix(userInput, "/save") {
	// 		err := commands.Save(userInput, history, context)
	// 		if err != nil {
	// 			fmt.Printf("Error saving history: %s", err)
	// 			continue
	// 		}

	// 		fmt.Println("History saved")
	// 		continue
	// 	}

	// 	if strings.HasPrefix(userInput, "/load") {
	// 		history, context, err = commands.Load(userInput)
	// 		if err != nil {
	// 			fmt.Printf("Error loading from history: %s", err)
	// 			continue
	// 		}

	// 		fmt.Println("History loaded")
	// 		continue
	// 	}

	// 	if context != "" {
	// 		userInput = fmt.Sprintf("%s\n---Context---\n%s", userInput, context)
	// 		context = ""
	// 	}

	// 	history = append(history, llm.Message{
	// 		Role: "user",
	// 		Content: []llm.Content{
	// 			{
	// 				Text:        userInput,
	// 				ContentType: "text",
	// 			},
	// 		},
	// 	})

	// 	messageChan, deltaChan, err := provider.Chat(modelCtx, profile.Model, 0, history, nil)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	fmt.Printf("Assistant says:")
	// 	for delta := range deltaChan {
	// 		fmt.Printf("%s", delta)
	// 	}

	// 	fmt.Println("")

	// 	response := <-messageChan
	// 	history = append(history, response)

	// }

	return nil
}
