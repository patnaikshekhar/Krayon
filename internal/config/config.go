package config

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type config struct {
	DefaultProfile string    `yaml:"default_profile"`
	Profiles       []Profile `yaml:"credentials"`
	PluginsServer  string    `yaml:"plugins_server"`
}

type Profile struct {
	Name     string `yaml:"profile"`
	ApiKey   string `yaml:"api_key"`
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Stream   bool   `yaml:"stream"`
}

func (c *config) GetProfile(name string) *Profile {
	for _, p := range c.Profiles {
		if p.Name == name {
			return &p
		}
	}

	return nil
}

func (c *config) AddProfile(name, provider, apiKey, model string, stream bool) {
	for i, p := range c.Profiles {
		if p.Name == name {
			c.Profiles[i] = Profile{name, apiKey, provider, model, stream}
			return
		}
	}

	c.Profiles = append(c.Profiles, Profile{name, apiKey, provider, model, stream})

	if c.DefaultProfile == "" {
		c.DefaultProfile = name
	}

	Save(c)
}

func GetConfigBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".krayon"), nil
}

func getConfigPath() (string, error) {
	krayonDirectory, err := GetConfigBasePath()
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(krayonDirectory, os.ModePerm)
	if err != nil {
		return "", err
	}

	return path.Join(krayonDirectory, "config.yaml"), nil
}

func Load() (*config, error) {
	configLocation, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	fileBytes, err := os.ReadFile(configLocation)
	if err != nil {
		return &config{}, nil
	}

	var cfg *config
	err = yaml.Unmarshal(fileBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *config) error {
	configLocation, err := getConfigPath()
	if err != nil {
		return err
	}

	dataBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configLocation, dataBytes, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
