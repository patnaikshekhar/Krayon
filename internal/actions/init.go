package actions

import (
	"fmt"
	"krayon/internal/config"

	"github.com/urfave/cli/v2"
)

func Init(ctx *cli.Context) error {
	key := ctx.String("key")
	provider := ctx.String("provider")
	name := ctx.String("name")
	model := ctx.String("model")
	stream := ctx.Bool("stream")

	if name == "" {
		fmt.Println("Please enter the Name to use for this profile: ")
		fmt.Scanln(&name)
	}

	if key == "" {
		fmt.Println("Please enter the API key to use for this profile: ")
		fmt.Scanln(&key)
	}

	if model == "" {
		fmt.Println("Please enter the model to use for this provider: ")
		fmt.Scanln(&model)
	}

	// Load existing config.yaml from the directory
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Update Config
	cfg.AddProfile(name, provider, key, model, stream)

	return nil
}
