package actions

import (
	"fmt"
	"io"
	"krayon/internal/config"
	"krayon/internal/plugins"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/urfave/cli/v2"
)

func DownloadPlugin(ctx *cli.Context) error {
	userOS := runtime.GOOS
	platform := "mac"
	switch userOS {
	case "windows":
		platform = "windows"
	case "darwin":
		platform = "mac"
	case "linux":
		platform = "linux"
	default:
		platform = "mac"
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.PluginsServer == "" {
		return fmt.Errorf("Plugins server connection string not found in Krayon Config")
	}

	downloadEndpoint := fmt.Sprintf("%s/api/v1/plugins/%s/versions/%s/platforms/%s", cfg.PluginsServer, ctx.String("plugin"), ctx.String("version"), platform)
	resp, err := http.Get(downloadEndpoint)
	if err != nil {
		return err
	}

	kbp, err := config.GetConfigBasePath()
	if err != nil {
		return err
	}

	pluginsDir := path.Join(kbp, "plugins")
	err = os.MkdirAll(pluginsDir, os.ModePerm)
	if err != nil {
		return err
	}

	baseFileName := fmt.Sprintf("%s_%s", ctx.String("plugin"), ctx.String("version"))
	filePath := path.Join(pluginsDir, baseFileName)
	binary, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(binary, resp.Body)
	if err != nil {
		return err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	newMode := info.Mode() | 0111
	err = os.Chmod(filePath, newMode)
	if err != nil {
		return err
	}

	err = plugins.AddPluginToManifest(ctx.String("plugin"), ctx.String("version"))
	if err != nil {
		return err
	}

	return nil
}
