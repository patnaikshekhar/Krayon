package plugins

import (
	"krayon/internal/config"
	"os"
	"path"
	"time"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Plugins []ManifestPlugin `yaml:"plugins"`
}

type ManifestPlugin struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	InstalledAt time.Time `yaml:"installed_at"`
}

func LoadManifest() (*Manifest, error) {
	configBasePath, err := config.GetConfigBasePath()
	if err != nil {
		return nil, err
	}

	manifestBasePath := path.Join(configBasePath, "plugins")
	err = os.MkdirAll(manifestBasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	manifestPath := path.Join(configBasePath, "plugins", "manifest.yaml")

	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return &Manifest{Plugins: []ManifestPlugin{}}, nil
	}

	manifest := Manifest{}
	err = yaml.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return &Manifest{Plugins: []ManifestPlugin{}}, nil
	}

	return &manifest, nil
}

func SaveManifest(manifest *Manifest) error {
	configBasePath, err := config.GetConfigBasePath()
	if err != nil {
		return err
	}

	manifestPath := path.Join(configBasePath, "plugins", "manifest.yaml")

	err = os.MkdirAll(path.Join(configBasePath, "plugins"), os.ModePerm)
	if err != nil {
		return err
	}

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, manifestBytes, os.ModePerm)
}

func AddPluginToManifest(pluginName string, pluginVersion string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	bFound := false
	for i, plugin := range manifest.Plugins {
		if plugin.Name == pluginName {
			manifest.Plugins[i].Version = pluginVersion
			bFound = true
		}
	}

	if !bFound {
		manifest.Plugins = append(manifest.Plugins, ManifestPlugin{
			Name:        pluginName,
			Version:     pluginVersion,
			InstalledAt: time.Now(),
		})
	}

	err = SaveManifest(manifest)
	if err != nil {
		return err
	}

	return nil
}
