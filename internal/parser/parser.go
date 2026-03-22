// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package parser

// ToolCall is the platform-agnostic representation of an intercepted tool call.
// Everything downstream of the normalizer (matcher, logger, main) works with
// this struct exclusively.
type ToolCall struct {
	Tool     string
	RawInput string
	CWD      string
	Session  string
	Worktree string // populated by OpenCode normalizer, empty for Claude Code
}

// Normalizer converts a raw platform-specific JSON payload into a ToolCall.
// Each supported platform (OpenCode, Claude Code) has its own implementation.
type Normalizer interface {
	Normalize(raw []byte) (ToolCall, error)
}
