package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing temp yaml: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	yaml := `
patterns:
  forbidden:
    - "rm -rf"
    - "sudo"
mode: enforcement
log_path: /tmp/audit.log
`
	cfg, err := Load(writeTempYAML(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(cfg.Patterns))
	}
	if cfg.Patterns[0] != "rm -rf" {
		t.Errorf("expected first pattern %q, got %q", "rm -rf", cfg.Patterns[0])
	}
	if cfg.Mode != ModeDefault {
		t.Errorf("expected mode %q, got %q", ModeEnforcement, cfg.Mode)
	}
	if cfg.LogPath != "/tmp/audit.log" {
		t.Errorf("expected log_path %q, got %q", "/tmp/audit.log", cfg.LogPath)
	}
}

func TestLoad_DefaultsEmptyModeToEnforcement(t *testing.T) {
	yaml := `
patterns:
  forbidden:
    - "rm -rf"
`
	cfg, err := Load(writeTempYAML(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != ModeEnforcement {
		t.Errorf("expected mode to default to %q, got %q", ModeEnforcement, cfg.Mode)
	}
}

func TestLoad_SimMode(t *testing.T) {
	yaml := `
patterns:
  forbidden:
    - "sudo"
mode: sim
`
	cfg, err := Load(writeTempYAML(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Mode != ModeSim {
		t.Errorf("expected mode %q, got %q", ModeSim, cfg.Mode)
	}
}

func TestLoad_InvalidMode(t *testing.T) {
	yaml := `
patterns:
  forbidden:
    - "rm -rf"
mode: yolo
`
	_, err := Load(writeTempYAML(t, yaml))
	if err == nil {
		t.Fatal("expected validation error for invalid mode, got nil")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("expected error to mention invalid mode, got: %v", err)
	}
}

func TestLoad_EmptyPatterns(t *testing.T) {
	yaml := `
patterns:
  forbidden: []
`
	_, err := Load(writeTempYAML(t, yaml))
	if err == nil {
		t.Fatal("expected validation error for empty patterns, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error to mention empty patterns, got: %v", err)
	}
}

func TestLoad_MissingPatternsKey(t *testing.T) {
	yaml := `
mode: enforcement
`
	_, err := Load(writeTempYAML(t, yaml))
	if err == nil {
		t.Fatal("expected validation error for missing patterns, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "reading config file") {
		t.Errorf("expected wrapped read error, got: %v", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTempYAML(t, `{{{not valid yaml`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid yaml, got nil")
	}
	if !strings.Contains(err.Error(), "parsing config file") {
		t.Errorf("expected wrapped parse error, got: %v", err)
	}
}

func TestLoad_RuntimeFieldsNotSetFromYAML(t *testing.T) {
	yaml := `
patterns:
  forbidden:
    - "rm -rf"
agent_id: "should-be-ignored"
sid: "should-be-ignored"
agent_seq: 99
kill_agent: true
`
	cfg, err := Load(writeTempYAML(t, yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Runtime fields should remain at zero values since they have yaml:"-".
	if cfg.AgentID != "" {
		t.Errorf("expected AgentID to be empty, got %q", cfg.AgentID)
	}
	if cfg.SID != "" {
		t.Errorf("expected SID to be empty, got %q", cfg.SID)
	}
	if cfg.AgentSeq != 0 {
		t.Errorf("expected AgentSeq to be 0, got %d", cfg.AgentSeq)
	}
	if cfg.KillAgent {
		t.Error("expected KillAgent to be false")
	}
}

func TestLoadDefault_WithDefaultYAML(t *testing.T) {
	// LoadDefault uses a relative path, so we need to run from project root.
	// Change to project root for this test.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting working dir: %v", err)
	}

	// Walk up from internal/config to project root.
	projectRoot := filepath.Join(origDir, "..", "..")
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("changing to project root: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Patterns) == 0 {
		t.Fatal("expected patterns to be populated from default.yaml")
	}

	// Check a known pattern from config/default.yaml.
	found := false
	for _, p := range cfg.Patterns {
		if p == "rm -rf" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'rm -rf' in patterns, got: %v", cfg.Patterns)
	}
}
