package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds pattern configuration loaded from YAML plus runtime fields
// injected by main from CLI flags. The YAML file is purely pattern
// configuration — all operational concerns (log path, platform) are CLI
// flags and never appear here.
type Config struct {
	// From YAML
	Patterns []string

	// Runtime — set by main from CLI flags, never from YAML
	KillAgent bool
	AgentID   string
	SID       string
	AgentSeq  int
}

// configFile mirrors the YAML structure on disk:
//
//	patterns:
//	  - "rm -rf /"
//	  - ...
type configFile struct {
	Patterns []string `yaml:"patterns"`
}

// Load reads and parses the YAML file at path, returning a validated Config.
// There is no default path — if --config is not provided, main should exit 1
// with a message directing the user to supply --config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	var raw configFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	cfg := &Config{
		Patterns: raw.Patterns,
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config file %s: %w", path, err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.Patterns) == 0 {
		return errors.New("config: patterns list is empty")
	}
	return nil
}
