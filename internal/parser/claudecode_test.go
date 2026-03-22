// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package parser

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Bash tool
// ---------------------------------------------------------------------------

func TestClaudeCode_BashTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Bash",
		"tool_input": { "command": "npm test", "description": "Run test suite" },
		"session_id": "abc123",
		"cwd": "/home/user/project"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tc.Tool != "Bash" {
		t.Errorf("Tool = %q, want %q", tc.Tool, "Bash")
	}
	if tc.RawInput != "npm test" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "npm test")
	}
	if tc.CWD != "/home/user/project" {
		t.Errorf("CWD = %q, want %q", tc.CWD, "/home/user/project")
	}
	if tc.Session != "abc123" {
		t.Errorf("Session = %q, want %q", tc.Session, "abc123")
	}
	if tc.Worktree != "" {
		t.Errorf("Worktree = %q, want empty for Claude Code", tc.Worktree)
	}
}

func TestClaudeCode_BashMissingCommand(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Bash",
		"tool_input": { "description": "oops no command" },
		"session_id": "abc123",
		"cwd": "/home/user/project"
	}`)

	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for missing command, got nil")
	}
	if !strings.Contains(err.Error(), "missing 'command'") {
		t.Errorf("error = %q, want mention of missing command", err.Error())
	}
}

// ---------------------------------------------------------------------------
// File tools (Read, Write, Edit, MultiEdit)
// ---------------------------------------------------------------------------

func TestClaudeCode_ReadTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Read",
		"tool_input": { "file_path": "/etc/passwd", "offset": 1, "limit": 50 },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "/etc/passwd" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "/etc/passwd")
	}
}

func TestClaudeCode_WriteTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Write",
		"tool_input": { "file_path": "/home/user/.env", "content": "SECRET=123" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "/home/user/.env" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "/home/user/.env")
	}
}

func TestClaudeCode_EditTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Edit",
		"tool_input": { "file_path": "/home/user/app.go", "old_string": "foo", "new_string": "bar" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "/home/user/app.go" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "/home/user/app.go")
	}
}

func TestClaudeCode_MultiEditTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "MultiEdit",
		"tool_input": { "file_path": "/home/user/app.go", "edits": [] },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "/home/user/app.go" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "/home/user/app.go")
	}
}

func TestClaudeCode_FileToolMissingPath(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Write",
		"tool_input": { "content": "hello" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for missing file_path, got nil")
	}
	if !strings.Contains(err.Error(), "missing 'file_path'") {
		t.Errorf("error = %q, want mention of missing file_path", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Glob and Grep tools
// ---------------------------------------------------------------------------

func TestClaudeCode_GlobToolPatternAndPath(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Glob",
		"tool_input": { "pattern": "**/*.pem", "path": "/home/user/.ssh" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both pattern and path should be concatenated.
	if tc.RawInput != "**/*.pem /home/user/.ssh" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "**/*.pem /home/user/.ssh")
	}
}

func TestClaudeCode_GlobToolPatternOnly(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Glob",
		"tool_input": { "pattern": "**/*.key" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "**/*.key" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "**/*.key")
	}
}

func TestClaudeCode_GrepTool(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Grep",
		"tool_input": { "pattern": "password", "path": "/etc/shadow", "include": "*.conf" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "password /etc/shadow" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "password /etc/shadow")
	}
}

func TestClaudeCode_GlobMissingPattern(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Glob",
		"tool_input": {},
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for missing pattern, got nil")
	}
	if !strings.Contains(err.Error(), "missing 'pattern'") {
		t.Errorf("error = %q, want mention of missing pattern", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Unknown / MCP tools (default fallback)
// ---------------------------------------------------------------------------

func TestClaudeCode_UnknownToolStringifiesFallback(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "mcp__memory__create_entities",
		"tool_input": { "entities": [{"name": "test"}] },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be the raw JSON string of tool_input.
	if !strings.Contains(tc.RawInput, `"entities"`) {
		t.Errorf("RawInput = %q, want stringified tool_input containing 'entities'", tc.RawInput)
	}
	if tc.Tool != "mcp__memory__create_entities" {
		t.Errorf("Tool = %q, want %q", tc.Tool, "mcp__memory__create_entities")
	}
}

// ---------------------------------------------------------------------------
// Error cases
// ---------------------------------------------------------------------------

func TestClaudeCode_InvalidJSON(t *testing.T) {
	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize([]byte(`{not json}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "claudecode: parsing payload") {
		t.Errorf("error = %q, want wrapped parse error", err.Error())
	}
}

func TestClaudeCode_EmptyInput(t *testing.T) {
	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestClaudeCode_EmptyToolInput(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Bash",
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for missing tool_input, got nil")
	}
	if !strings.Contains(err.Error(), "tool_input is empty") {
		t.Errorf("error = %q, want mention of empty tool_input", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Case insensitivity
// ---------------------------------------------------------------------------

func TestClaudeCode_ToolNameCaseInsensitive(t *testing.T) {
	// Claude Code sends "Bash" (capitalized). Verify lowercase also works
	// in case payload shape changes.
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "bash",
		"tool_input": { "command": "echo hi" },
		"session_id": "s1",
		"cwd": "/tmp"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "echo hi" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "echo hi")
	}
}

// ---------------------------------------------------------------------------
// Extra fields ignored
// ---------------------------------------------------------------------------

func TestClaudeCode_ExtraFieldsIgnored(t *testing.T) {
	// Claude Code docs show transcript_path, permission_mode, etc. in the
	// payload. These should be silently ignored by the normalizer.
	raw := []byte(`{
		"hook_event_name": "PreToolUse",
		"tool_name": "Bash",
		"tool_input": { "command": "ls" },
		"session_id": "s1",
		"cwd": "/tmp",
		"transcript_path": "/path/to/transcript.jsonl",
		"permission_mode": "default",
		"tool_use_id": "tu_123"
	}`)

	n := &ClaudeCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "ls" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "ls")
	}
}
