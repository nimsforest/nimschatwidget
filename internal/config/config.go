package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig `yaml:"server"`
	WebchatURL string       `yaml:"webchat_url"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

func Load(path string) (Config, error) {
	var cfg Config

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return cfg, fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, ".nimschatwidget", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
