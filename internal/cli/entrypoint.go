package cli

import (
	"fmt"
	"os"

	"github.com/mikehaus/marut/internal/config"
	"github.com/mikehaus/marut/internal/parser"
	"github.com/mikehaus/marut/schema"
)

// BuildEntry constructs an AuditEntry from a normalized ToolCall and match
// outcome. UID and Timestamp are left blank — the logger assigns them at
// write time. exitCode should reflect the actual process exit code:
// 0 for pass, 1 for internal error, 2 for blocked.
func BuildEntry(
	cfg *config.Config,
	tc parser.ToolCall,
	mode string,
	level schema.Level,
	action schema.Action,
	pattern string,
	exitCode int,
	latencyMs float64,
	message string,
) schema.AuditEntry {
	return schema.AuditEntry{
		Level:   level,
		AgentID: cfg.AgentID,
		PID:     os.Getpid(),
		SID:     cfg.SID,
		Context: schema.Context{
			CWD:      tc.CWD,
			AgentSeq: cfg.AgentSeq,
		},
		Event: schema.Event{
			Type:         EventType(tc.Tool),
			Tool:         tc.Tool,
			RawInput:     tc.RawInput,
			MatchPattern: pattern,
			Mode:         mode,
		},
		Outcome: schema.Outcome{
			Action:    action,
			ExitCode:  exitCode,
			LatencyMs: latencyMs,
			Message:   message,
		},
	}
}

// BlockMessage returns the stderr message printed when a tool call is blocked.
func BlockMessage(pattern string) string {
	return fmt.Sprintf("Deterministic block: pattern %q is forbidden.", pattern)
}

// EventType maps a tool name to the appropriate schema.EventType.
// File-access tools → EventFileAccess; everything else → EventShellCommand.
// The default is EventShellCommand because it is the safest catch-all: MCP
// tools, custom tools, and any future tool types that haven't been explicitly
// mapped will still be evaluated and logged, just under the shell_command type.
// TODO: add network tool mapping when network tools are observed in audit logs.
// EventNetworkAccess exists in the schema but nothing maps to it yet — curl
// invoked as a tool (rather than inside a bash command) would be the first case.
func EventType(tool string) schema.EventType {
	switch tool {
	case "read", "write", "edit", "multiedit", "glob", "grep":
		return schema.EventFileAccess
	default:
		return schema.EventShellCommand
	}
}
