package parser

import (
	"strings"
	"testing"
)

func TestOpenCode_ValidPayload(t *testing.T) {
	raw := []byte(`{
		"tool": "bash",
		"raw_input": "go test ./...",
		"cwd": "/home/user/project",
		"worktree": "feat-auth"
	}`)

	n := &OpenCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tc.Tool != "bash" {
		t.Errorf("Tool = %q, want %q", tc.Tool, "bash")
	}
	if tc.RawInput != "go test ./..." {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "go test ./...")
	}
	if tc.CWD != "/home/user/project" {
		t.Errorf("CWD = %q, want %q", tc.CWD, "/home/user/project")
	}
	if tc.Worktree != "feat-auth" {
		t.Errorf("Worktree = %q, want %q", tc.Worktree, "feat-auth")
	}
	if tc.Session != "" {
		t.Errorf("Session = %q, want empty (OpenCode has no session field)", tc.Session)
	}
}

func TestOpenCode_MissingRawInput(t *testing.T) {
	raw := []byte(`{
		"tool": "bash",
		"cwd": "/home/user/project"
	}`)

	n := &OpenCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for missing raw_input, got nil")
	}
	if !strings.Contains(err.Error(), "missing raw_input") {
		t.Errorf("error = %q, want mention of missing raw_input", err.Error())
	}
}

func TestOpenCode_EmptyRawInput(t *testing.T) {
	raw := []byte(`{
		"tool": "bash",
		"raw_input": "",
		"cwd": "/home/user/project"
	}`)

	n := &OpenCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for empty raw_input, got nil")
	}
}

func TestOpenCode_InvalidJSON(t *testing.T) {
	raw := []byte(`{not valid json}`)

	n := &OpenCodeNormalizer{}
	_, err := n.Normalize(raw)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "opencode: parsing payload") {
		t.Errorf("error = %q, want wrapped parse error", err.Error())
	}
}

func TestOpenCode_EmptyInput(t *testing.T) {
	n := &OpenCodeNormalizer{}
	_, err := n.Normalize([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestOpenCode_OptionalWorktreeEmpty(t *testing.T) {
	raw := []byte(`{
		"tool": "read",
		"raw_input": "/path/to/file.go",
		"cwd": "/home/user/project"
	}`)

	n := &OpenCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.Worktree != "" {
		t.Errorf("Worktree = %q, want empty when not provided", tc.Worktree)
	}
}

func TestOpenCode_ExtraFieldsIgnored(t *testing.T) {
	// If the shim sends fields we don't know about, they should be silently ignored.
	raw := []byte(`{
		"tool": "bash",
		"raw_input": "ls",
		"cwd": "/tmp",
		"worktree": "main",
		"some_future_field": "value"
	}`)

	n := &OpenCodeNormalizer{}
	tc, err := n.Normalize(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.RawInput != "ls" {
		t.Errorf("RawInput = %q, want %q", tc.RawInput, "ls")
	}
}
