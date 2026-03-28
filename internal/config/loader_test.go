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
	path := writeTempYAML(t, `
patterns:
  - "rm -rf /"
  - "sudo su"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(cfg.Patterns))
	}
	if cfg.Patterns[0] != "rm -rf /" {
		t.Errorf("expected first pattern %q, got %q", "rm -rf /", cfg.Patterns[0])
	}
}

func TestLoad_EmptyPatterns(t *testing.T) {
	path := writeTempYAML(t, `
patterns: []
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty patterns, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error to mention empty, got: %v", err)
	}
}

func TestLoad_MissingPatternsKey(t *testing.T) {
	path := writeTempYAML(t, `
`)
	_, err := Load(path)
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
	// Runtime fields have no yaml tags — confirm they are never populated
	// from the file regardless of what keys appear in it.
	path := writeTempYAML(t, `
patterns:
  - "rm -rf /"
agent_id: "should-be-ignored"
sid: "should-be-ignored"
agent_seq: 99
kill_agent: true
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestLoad_ModeAndLogPathIgnoredIfPresent(t *testing.T) {
	// Confirm that stale YAML files containing unknown fields like log_path
	// load without error — those keys are simply ignored by the parser.
	path := writeTempYAML(t, `
patterns:
  - "rm -rf /"
log_path: /tmp/audit.log
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error loading yaml with legacy fields: %v", err)
	}
	if len(cfg.Patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(cfg.Patterns))
	}
}
