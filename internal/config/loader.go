package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ModelCost holds per-million-token pricing for a model. Used for savings
// estimation in audit entries. Selected via --model flag in main.
// If the key is not found, savings calculation is skipped — never an error.
type ModelCost struct {
	Input  float64 // cost per million input tokens (USD)
	Output float64 // cost per million output tokens (USD)
}

// ModelCosts is the registry of known model pricing.
// Extend as new models are deployed.
var ModelCosts = map[string]ModelCost{
	"claude-haiku-4-5":  {Input: 1.00, Output: 5.00},
	"claude-sonnet-4-6": {Input: 3.00, Output: 15.00},
	"claude-opus-4-6":   {Input: 5.00, Output: 25.00},
	"gemini-2.0-flash":  {Input: 0.10, Output: 0.40},
	"gpt-4o-mini":       {Input: 0.25, Output: 2.00},
}

// Config holds pattern configuration loaded from YAML plus runtime fields
// injected by main from CLI flags. The YAML file is purely pattern
// configuration — all operational concerns (log path, mode, platform)
// are CLI flags and never appear here.
type Config struct {
	// From YAML
	Patterns       []string
	MonitorPhrases []string

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
//	monitor_phrases:
//	  - "oops"
//	  - ...
type configFile struct {
	Patterns       []string `yaml:"patterns"`
	MonitorPhrases []string `yaml:"monitor_phrases"`
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
		Patterns:       raw.Patterns,
		MonitorPhrases: raw.MonitorPhrases,
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
