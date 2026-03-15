package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TODO: Add validation on config forbidden patterns

type Config struct {
	Patterns PatternsConfig `yaml:"patterns"`
}

type PatternsConfig struct {
	Forbidden []string `yaml:"forbidden"`
}

func Load(path string) (*Config, error) {
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

func LoadDefault() (*Config, error) {
	return Load("config/default.yaml")
}
