// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package parser

// Payload shape verified against OpenCode plugin docs:
// https://opencode.ai/docs/plugins/ (2026-03-15)
//
// SIM MODE VERIFICATION (Block 9):
// - Confirm output.args shape matches what the TS shim sends. The shim builds
//   raw_input from output.args.command ?? output.args.filePath ?? JSON.stringify(output.args).
//   If output.args has a different structure in practice, raw_input may be wrong or empty.
// - Check whether additional context fields (e.g. project, client) are available
//   in the hook and worth forwarding to the binary.

import (
	"encoding/json"
	"fmt"
)

// openCodePayload is the JSON shape sent by the OpenCode plugin shim.
// The shim pre-normalizes the payload before piping it to marut, so this
// is a flat struct with the input already resolved to raw_input.
type openCodePayload struct {
	Tool     string `json:"tool"`
	RawInput string `json:"raw_input"`
	CWD      string `json:"cwd"`
	Worktree string `json:"worktree"`
}

// OpenCodeNormalizer converts an OpenCode plugin payload into a ToolCall.
type OpenCodeNormalizer struct{}

func (n *OpenCodeNormalizer) Normalize(raw []byte) (ToolCall, error) {
	var p openCodePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return ToolCall{}, fmt.Errorf("opencode: parsing payload: %w", err)
	}

	if p.RawInput == "" {
		return ToolCall{}, fmt.Errorf("opencode: payload missing raw_input")
	}

	return ToolCall{
		Tool:     p.Tool,
		RawInput: p.RawInput,
		CWD:      p.CWD,
		Worktree: p.Worktree,
	}, nil
}
