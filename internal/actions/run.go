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
