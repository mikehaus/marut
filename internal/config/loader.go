package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ModeEnforcement = "enforcement"
	ModeSim         = "sim"
	ModeDefault     = ModeEnforcement
)

// Config holds both file-sourced configuration (patterns, mode, log path)
// and runtime fields set by the orchestrator (agent identity, kill switch).
type Config struct {
	// YAML config Values
	Patterns []string `yaml:"-"`
	Mode     string   `yaml:"mode"`
	LogPath  string   `yaml:"log_path"`

	// Runtime orchestratory values
	KillAgent bool   `yaml:"-"`
	AgentID   string `yaml:"-"`
	SID       string `yaml:"-"`
	AgentSeq  int    `yaml:"-"`
}

// configFile mirrors the nested YAML structure in config/default.yaml:
//
//	patterns:
//	  forbidden:
//	    - "rm -rf"
//	    - ...
type configFile struct {
	Patterns struct {
		Forbidden []string `yaml:"forbidden"`
	} `yaml:"patterns"`
	Mode    string `yaml:"mode"`
	LogPath string `yaml:"log_path"`
}

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
		Patterns: raw.Patterns.Forbidden,
		Mode:     raw.Mode,
		LogPath:  raw.LogPath,
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config file %s: %w", path, err)
	}

	return cfg, nil
}

func LoadDefault() (*Config, error) {
	return Load("config/default.yaml")
}

func (c *Config) Validate() error {
	if len(c.Patterns) == 0 {
		return errors.New("config: forbidden patterns list is empty")
	}

	if c.Mode == "" {
		c.Mode = ModeDefault
	}

	if c.Mode != ModeDefault && c.Mode != ModeSim {
		return fmt.Errorf("config: invalid mode %q (must be %q or %q)", c.Mode, ModeEnforcement, ModeSim)
	}

	return nil
}
