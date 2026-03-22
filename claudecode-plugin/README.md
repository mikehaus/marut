# Marut Claude Code Plugin

This plugin integrates Marut with Claude Code's hook system to validate and control tool usage.

## Installation

### Method 1: Using `--plugin-dir` (Development/Testing)

For testing or development, load the plugin directly without installation:

```bash
# From the marut project root
make build  # Build marut binary
cd claudecode-plugin && npm install && npm run build && cd ..

# Run Claude Code with the plugin
claude --plugin-dir ./claudecode-plugin
```

Set environment variables before running:

```bash
export MARUT_BIN="$(pwd)/marut"
export MARUT_ARGS="--config $(pwd)/config/default.yaml --log $(pwd)/audit.log"
```

### Method 2: Install via Plugin Manager

For permanent installation, use Claude Code's plugin system:

```bash
# Build the plugin first
make build
cd claudecode-plugin && npm install && npm run build && cd ..

# Install via CLI
claude plugin install ./claudecode-plugin --scope user
# Or for project-level: --scope project
# Or for local (gitignored): --scope local
```

Set environment variables in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export MARUT_BIN="marut"  # If marut is in PATH
export MARUT_ARGS="--config /path/to/config.yaml --log /path/to/audit.log"
```

## Configuration

The plugin uses a **wrapper script** (`marut-wrapper.sh`) that automatically configures paths. You can optionally override these with environment variables:

- `MARUT_BIN`: Path to the marut binary (default: `$HOME/Developer/projects/go/marut/marut`)
- `MARUT_CONFIG`: Path to config file (default: `$HOME/Developer/projects/go/marut/config/default.yaml`)
- `MARUT_LOG`: Path to audit log (default: `$HOME/Developer/projects/go/marut/audit.log`)

**Note:** The wrapper script has sensible defaults, so you typically don't need to set these variables unless you've installed marut in a custom location.

## How It Works

The plugin uses Claude Code's **command hook** system:

1. **Hook Registration**: The plugin registers a `PreToolUse` hook via `hooks/hooks.json`
2. **Event Trigger**: Before Claude Code executes any tool, the hook fires
3. **Wrapper Script**: Claude Code runs `marut-wrapper.sh` with tool context as JSON on stdin
4. **Validation**: The wrapper script invokes marut with the correct arguments
5. **Decision**: Based on marut's exit code:
   - `0` = Allow (tool executes normally)
   - `2` = Block (tool is prevented, Claude sees the error)
   - `1` = Internal error (logged, execution may continue)

### Hook Configuration

The plugin's `hooks/hooks.json` configures the PreToolUse hook:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "\"${CLAUDE_PLUGIN_ROOT}/marut-wrapper.sh\"",
            "timeout": 5,
            "statusMessage": "Validating command with Marut..."
          }
        ]
      }
    ]
  }
}
```

The `matcher: "*"` means the hook fires for **all tools**. You can restrict it to specific tools:
- `"matcher": "Bash"` - Only Bash commands
- `"matcher": "Edit|Write"` - File modifications
- `"matcher": "mcp__.*"` - All MCP tool calls

## Debugging

Run Claude Code with `--debug` to see plugin loading and hook execution details:

```bash
claude --debug --plugin-dir ./claudecode-plugin
```

This shows:
- Plugin loading status
- Hook registration
- Hook execution and exit codes

## Payload Format

The plugin sends payloads in this format to marut:

```json
{
  "hook_event_name": "PreToolUse",
  "tool_name": "Bash",
  "tool_input": { "command": "...", "description": "..." },
  "session_id": "abc123",
  "cwd": "/path/to/project",
  "worktree": "/path/to/worktree"
}
```

This matches the format expected by the `ClaudeCodeNormalizer` in the Go binary.
