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
					&cli.BoolFlag{
						Name:    "stream",
						Usage:   "Should stream",
						Aliases: []string{"s"},
					},
				},
				Action: actions.Init,
			},
			{
				Name:        "plugins",
				Aliases:     []string{"p"},
				Description: "Manage plugins in Krayon",
				Subcommands: []*cli.Command{
					{
						Name:        "server",
						Description: "Manage the plugins server",
						Aliases:     []string{"s"},
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:    "port",
								Usage:   "Server port",
								Aliases: []string{"p"},
								Value:   8000,
							},
							&cli.StringFlag{
								Name:    "driver",
								Usage:   "Database driver to use, could be 'postgres', 'sqlite3",
								Aliases: []string{"d"},
								Value:   "sqlite3",
							},
							&cli.StringFlag{
								Name:    "connection-string",
								Usage:   "Connection String to the database",
								Aliases: []string{"cs"},
								Value:   "krayon_plugins.db",
							},
							&cli.StringFlag{
								Name:  "storage-type",
								Usage: "Type of storage to use - can be gcs or s3",
								Value: "gcs",
							},
							&cli.StringFlag{
								Name:  "bucket",
								Usage: "The storage bucket name",
							},
						},
						Action: actions.PluginsServe,
					},
					{
						Name:        "list",
						Description: "List plugins",
						Aliases:     []string{"ls"},
						Action:      actions.ListPlugins,
					},
					{
						Name:        "install",
						Description: "Install a plugin",
						Aliases:     []string{"i"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "plugin",
								Usage:    "Name of the plugin",
								Aliases:  []string{"p"},
								Required: true,
							},
							&cli.StringFlag{
								Name:     "version",
								Usage:    "Plugin version to download in Semver",
								Aliases:  []string{"ver"},
								Required: true,
							},
						},
						Action: actions.DownloadPlugin,
					},
					{
						Name:        "register",
						Description: "Register a plugin",
						Aliases:     []string{"rg"},
					},
				},
			},
		},
		Action: actions.Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "profile",
				Usage:   "Select a profile",
				Aliases: []string{"p"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
