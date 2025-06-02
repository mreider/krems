package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config matches the structure of config.yaml
type QuackerConfig struct {
	Domain    string `yaml:"domain"`
	SiteOwner string `yaml:"site_owner"`
	Target    string `yaml:"target"`
}

type Config struct {
	Website struct {
		URL      string `yaml:"url"`
		Name     string `yaml:"name"`
		BasePath string `yaml:"basePath,omitempty"` // Added BasePath
		DevPath  string `yaml:"devPath,omitempty"`
	} `yaml:"website"`
	Menu []struct {
		Title string `yaml:"title"`
		Path  string `yaml:"path"`
	} `yaml:"menu"`

	Quacker *QuackerConfig `yaml:"quacker,omitempty"`
}

func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
