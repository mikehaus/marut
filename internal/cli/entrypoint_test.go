// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package cli

import (
	"strings"
	"testing"

	"github.com/mikehaus/marut/internal/config"
	"github.com/mikehaus/marut/internal/parser"
	"github.com/mikehaus/marut/schema"
)

// --- EventType ---

func TestEventType_BashIsShellCommand(t *testing.T) {
	if got := EventType("bash"); got != schema.EventShellCommand {
		t.Errorf("expected %q, got %q", schema.EventShellCommand, got)
	}
}

func TestEventType_UnknownToolIsShellCommand(t *testing.T) {
	// MCP tools, custom tools, and future tool types all fall through to
	// EventShellCommand — the safest catch-all so they are still evaluated
	// and logged even without an explicit mapping.
	if got := EventType("mcp__some_tool"); got != schema.EventShellCommand {
		t.Errorf("expected %q, got %q", schema.EventShellCommand, got)
	}
}

func TestEventType_EmptyToolIsShellCommand(t *testing.T) {
	if got := EventType(""); got != schema.EventShellCommand {
		t.Errorf("expected %q, got %q", schema.EventShellCommand, got)
	}
}

func TestEventType_FileToolsAreFileAccess(t *testing.T) {
	fileTools := []string{"read", "write", "edit", "multiedit", "glob", "grep"}
	for _, tool := range fileTools {
		if got := EventType(tool); got != schema.EventFileAccess {
			t.Errorf("tool %q: expected %q, got %q", tool, schema.EventFileAccess, got)
		}
	}
}

// --- BlockMessage ---

func TestBlockMessage_ContainsPattern(t *testing.T) {
	msg := BlockMessage("~/.ssh")
	if !strings.Contains(msg, "~/.ssh") {
		t.Errorf("expected message to contain pattern, got %q", msg)
	}
	if !strings.Contains(msg, "forbidden") {
		t.Errorf("expected message to contain 'forbidden', got %q", msg)
	}
}

// --- BuildEntry ---

func baseCfg() *config.Config {
	return &config.Config{
		Patterns: []string{"rm -rf /"},
		AgentID:  "test-agent",
		SID:      "test-sid",
		AgentSeq: 2,
	}
}

func baseToolCall() parser.ToolCall {
	return parser.ToolCall{
		Tool:     "bash",
		RawInput: "go test ./...",
		CWD:      "/tmp/worktree",
	}
}

func TestBuildEntry_PassOutcome(t *testing.T) {
	entry := BuildEntry(baseCfg(), baseToolCall(), "validate",
		schema.LevelPass, schema.ActionPass, "", 0, 0.5, "allowed")

	if entry.Level != schema.LevelPass {
		t.Errorf("level: want %q, got %q", schema.LevelPass, entry.Level)
	}
	if entry.Outcome.Action != schema.ActionPass {
		t.Errorf("action: want %q, got %q", schema.ActionPass, entry.Outcome.Action)
	}
	if entry.Outcome.ExitCode != 0 {
		t.Errorf("exit_code: want 0, got %d", entry.Outcome.ExitCode)
	}
	if entry.Outcome.Message != "allowed" {
		t.Errorf("message: want %q, got %q", "allowed", entry.Outcome.Message)
	}
}

func TestBuildEntry_DenyOutcome(t *testing.T) {
	tc := baseToolCall()
	tc.RawInput = "rm -rf /"
	msg := BlockMessage("rm -rf /")

	entry := BuildEntry(baseCfg(), tc, "validate",
		schema.LevelDeny, schema.ActionBlock, "rm -rf /", 2, 1.2, msg)

	if entry.Level != schema.LevelDeny {
		t.Errorf("level: want %q, got %q", schema.LevelDeny, entry.Level)
	}
	if entry.Outcome.Action != schema.ActionBlock {
		t.Errorf("action: want %q, got %q", schema.ActionBlock, entry.Outcome.Action)
	}
	if entry.Outcome.ExitCode != 2 {
		t.Errorf("exit_code: want 2, got %d", entry.Outcome.ExitCode)
	}
	if entry.Event.MatchPattern != "rm -rf /" {
		t.Errorf("match_pattern: want %q, got %q", "rm -rf /", entry.Event.MatchPattern)
	}
	if entry.Outcome.LatencyMs != 1.2 {
		t.Errorf("latency_ms: want 1.2, got %f", entry.Outcome.LatencyMs)
	}
}

func TestBuildEntry_ContextPopulated(t *testing.T) {
	entry := BuildEntry(baseCfg(), baseToolCall(), "validate",
		schema.LevelPass, schema.ActionPass, "", 0, 0, "allowed")

	if entry.Context.CWD != "/tmp/worktree" {
		t.Errorf("cwd: want %q, got %q", "/tmp/worktree", entry.Context.CWD)
	}
	if entry.Context.AgentSeq != 2 {
		t.Errorf("agent_seq: want 2, got %d", entry.Context.AgentSeq)
	}
	if entry.AgentID != "test-agent" {
		t.Errorf("agent_id: want %q, got %q", "test-agent", entry.AgentID)
	}
	if entry.SID != "test-sid" {
		t.Errorf("sid: want %q, got %q", "test-sid", entry.SID)
	}
}

func TestBuildEntry_EventPopulated(t *testing.T) {
	entry := BuildEntry(baseCfg(), baseToolCall(), "monitor",
		schema.LevelPass, schema.ActionPass, "", 0, 0, "allowed")

	if entry.Event.Tool != "bash" {
		t.Errorf("tool: want %q, got %q", "bash", entry.Event.Tool)
	}
	if entry.Event.RawInput != "go test ./..." {
		t.Errorf("raw_input: want %q, got %q", "go test ./...", entry.Event.RawInput)
	}
	if entry.Event.Mode != "monitor" {
		t.Errorf("mode: want %q, got %q", "monitor", entry.Event.Mode)
	}
	if entry.Event.Type != schema.EventShellCommand {
		t.Errorf("type: want %q, got %q", schema.EventShellCommand, entry.Event.Type)
	}
}

func TestBuildEntry_FileToolEventType(t *testing.T) {
	tc := baseToolCall()
	tc.Tool = "read"
	tc.RawInput = "/some/file.go"

	entry := BuildEntry(baseCfg(), tc, "validate",
		schema.LevelPass, schema.ActionPass, "", 0, 0, "allowed")

	if entry.Event.Type != schema.EventFileAccess {
		t.Errorf("type: want %q, got %q", schema.EventFileAccess, entry.Event.Type)
	}
}

func TestBuildEntry_UIDAndTimestampBlank(t *testing.T) {
	// UID and Timestamp must be blank — the logger stamps them at write time.
	entry := BuildEntry(baseCfg(), baseToolCall(), "validate",
		schema.LevelPass, schema.ActionPass, "", 0, 0, "allowed")

	if entry.UID != "" {
		t.Errorf("expected UID to be blank, got %q", entry.UID)
	}
	if entry.Timestamp != "" {
		t.Errorf("expected Timestamp to be blank, got %q", entry.Timestamp)
	}
}

func TestBuildEntry_SIMEntry(t *testing.T) {
	entry := BuildEntry(baseCfg(), baseToolCall(), "validate",
		schema.LevelSim, schema.ActionPass, "", 0, 0, "sim mode")

	if entry.Level != schema.LevelSim {
		t.Errorf("level: want %q, got %q", schema.LevelSim, entry.Level)
	}
	if entry.Outcome.Action != schema.ActionPass {
		t.Errorf("action: want %q, got %q", schema.ActionPass, entry.Outcome.Action)
	}
	if entry.Outcome.ExitCode != 0 {
		t.Errorf("exit_code: want 0 for SIM, got %d", entry.Outcome.ExitCode)
	}
}
