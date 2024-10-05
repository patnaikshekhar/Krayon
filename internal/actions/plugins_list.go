package actions

import (
	"encoding/json"
	"fmt"
	"io"
	"krayon/internal/config"
	"krayon/internal/db"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

func ListPlugins(ctx *cli.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.PluginsServer == "" {
		return fmt.Errorf("Plugins server connection string not found in Krayon Config")
	}

	pluginsEndpoint := fmt.Sprintf("%s/api/v1/plugins", cfg.PluginsServer)

	resp, err := http.Get(pluginsEndpoint)
	if err != nil {
		return err
	}

	plugins := []db.Plugin{}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBytes, &plugins)
	if err != nil {
		return err
	}

	pluginsWithVersions := map[string][]string{}

	for _, pl := range plugins {
		pluginsWithVersions[pl.Name], err = fetchPluginVersions(pl.ID, cfg.PluginsServer)
		if err != nil {
			return err
		}
	}

	generatePluginsTable(plugins, pluginsWithVersions)
	return nil

}

func fetchPluginVersions(pluginId int, server string) ([]string, error) {

	versionsEndpoint := fmt.Sprintf("%s/api/v1/plugins/%d/versions", server, pluginId)

	resp, err := http.Get(versionsEndpoint)
	if err != nil {
		return nil, err
	}

	pluginVersions := []db.PluginVersion{}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &pluginVersions)
	if err != nil {
		return nil, err
	}

	var versions []string

	for _, ver := range pluginVersions {
		versions = append(versions, ver.Version)
	}

	return versions, nil

}

func generatePluginsTable(plugins []db.Plugin, pluginsWithVersions map[string][]string) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("ID", "Name", "Description", "Versions")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, p := range plugins {
		tbl.AddRow(p.ID, p.Name, p.Description, strings.Join(pluginsWithVersions[p.Name], ", "))
	}

	tbl.Print()
}
