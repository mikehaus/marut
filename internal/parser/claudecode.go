package parser

// Payload shape and tool_input schemas verified against Claude Code hooks docs:
// https://code.claude.com/docs/en/hooks (2026-03-15)
//
// SIM MODE VERIFICATION (Block 9):
// - Run marut --sim --platform claudecode against a live Claude Code session
//   and compare SIM_RAW output to the struct below.
// - Confirm tool_input shapes for each tool type match the documented schemas.
// - Check for undocumented fields or wrapping in the real payload.
//
// KNOWN GAPS:
// - transcript_path and permission_mode are in the payload (per docs) but not
//   captured. Revisit if audit logging needs them.
// - Glob/Grep: raw_input is "pattern path" concatenated with a space. This is
//   synthetic (not a real command). Fine for substring matching but worth noting.
// - MCP tools (mcp__*) and any future tools fall to the default stringify path.
//   Monitor SIM logs to see if any need explicit handling.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// claudeCodePayload is the JSON shape sent by Claude Code's PreToolUse hook.
// tool_input is kept as raw JSON because its internal structure varies by tool.
type claudeCodePayload struct {
	HookEventName string          `json:"hook_event_name"`
	ToolName      string          `json:"tool_name"`
	ToolInput     json.RawMessage `json:"tool_input"`
	SessionID     string          `json:"session_id"`
	CWD           string          `json:"cwd"`
}

// ClaudeCodeNormalizer converts a Claude Code PreToolUse payload into a ToolCall.
type ClaudeCodeNormalizer struct{}

func (n *ClaudeCodeNormalizer) Normalize(raw []byte) (ToolCall, error) {
	var p claudeCodePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return ToolCall{}, fmt.Errorf("claudecode: parsing payload: %w", err)
	}

	rawInput, err := extractInput(p.ToolName, p.ToolInput)
	if err != nil {
		return ToolCall{}, fmt.Errorf("claudecode: %w", err)
	}

	return ToolCall{
		Tool:     p.ToolName,
		RawInput: rawInput,
		CWD:      p.CWD,
		Session:  p.SessionID,
	}, nil
}

// extractInput pulls the relevant string from tool_input based on tool type.
//
// Known tool_input shapes (from Claude Code docs):
//
//	Bash:      { "command": "...", "description": "...", "timeout": N }
//	Read:      { "file_path": "...", "offset": N, "limit": N }
//	Write:     { "file_path": "...", "content": "..." }
//	Edit:      { "file_path": "...", "old_string": "...", "new_string": "..." }
//	MultiEdit: { "file_path": "...", "edits": [...] }
//	Glob:      { "pattern": "...", "path": "..." }
//	Grep:      { "pattern": "...", "path": "...", "include": "..." }
//
// Unknown tools (including MCP tools like mcp__*) fall back to the raw JSON
// string of tool_input so the matcher still has something to check against.
func extractInput(toolName string, toolInput json.RawMessage) (string, error) {
	if len(toolInput) == 0 {
		return "", fmt.Errorf("tool_input is empty for tool %q", toolName)
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(toolInput, &fields); err != nil {
		return "", fmt.Errorf("parsing tool_input for tool %q: %w", toolName, err)
	}

	switch strings.ToLower(toolName) {
	case "bash":
		if cmd, ok := fields["command"].(string); ok && cmd != "" {
			return cmd, nil
		}
		return "", fmt.Errorf("tool_input missing 'command' for Bash tool")

	case "read", "write", "edit", "multiedit":
		if fp, ok := fields["file_path"].(string); ok && fp != "" {
			return fp, nil
		}
		return "", fmt.Errorf("tool_input missing 'file_path' for %s tool", toolName)

	case "glob", "grep":
		// Both have a "pattern" field; glob also has "path".
		// Concatenate what's available so the matcher can scan all of it.
		var parts []string
		if p, ok := fields["pattern"].(string); ok && p != "" {
			parts = append(parts, p)
		}
		if p, ok := fields["path"].(string); ok && p != "" {
			parts = append(parts, p)
		}
		if len(parts) == 0 {
			return "", fmt.Errorf("tool_input missing 'pattern' for %s tool", toolName)
		}
		return strings.Join(parts, " "), nil

	default:
		// Unknown tool (including MCP tools) -- stringify the entire
		// tool_input as fallback.
		return string(toolInput), nil
	}
}
