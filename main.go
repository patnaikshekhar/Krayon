package main

import (
	"log"
	"os"

	"krayon/internal/actions"

	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "Krayon",
		Usage: "Makes interacting with AI intuitive",
		Commands: []*cli.Command{
			{
				Name:        "init",
				Aliases:     []string{"i"},
				Description: "Setup the Krayon CLI",
				Usage:       "krayon init",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "key",
						Usage:   "Enter your AI provider's API Key",
						Aliases: []string{"k"},
					},
					&cli.StringFlag{
						Name:    "provider",
						Usage:   "Enter your AI provider's name",
						Value:   "anthropic",
						Aliases: []string{"p"},
					},
					&cli.StringFlag{
						Name:    "model",
						Usage:   "Enter the name of the model",
						Aliases: []string{"m"},
					},
					&cli.StringFlag{
						Name:    "name",
						Usage:   "Enter the name of the profile",
						Aliases: []string{"n"},
					},
				},
				Action: actions.Init,
			},
		},
		Action: actions.Run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
