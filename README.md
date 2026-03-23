# Marut: Deterministic Circuit Breaker for Autonomous Coding Agents

A sub-millisecond security layer for AI agents. No LLM in the loop.

**Marut** intercepts tool calls from autonomous coding agents (OpenCode, Claude Code), runs Aho-Corasick pattern matching against a configurable forbidden-words list, emits structured audit logs, and blocks dangerous commands before they execute.

**TL;DR:** Marut is a pre-call hook that stops agents before they can run `rm -rf ~/.ssh`.

---

## What It Does

Marut reads a tool call from stdin, checks it against a pattern list, writes an audit log entry, and exits with a code:

- **Exit 0:** Command is clean, allow execution
- **Exit 1:** Internal error (bad config, malformed payload)
- **Exit 2:** Command is forbidden, block execution

Latency: **~0.2ms** on pattern sets with 20-30 entries.

---

## Architecture

```
┌─────────────┐
│ OpenCode/   │
│ Claude Code │
└──────┬──────┘
       │ tool call
       ▼
┌──────────────────┐
│ Plugin Shim      │  Normalizes payload, spawns marut
│ (TypeScript)     │
└──────┬───────────┘
       │ JSON over stdin
       ▼
┌──────────────────┐
│ Marut Binary     │  Parse → Match → Log → Exit 0/1/2
│ (Go)             │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ audit.log        │  NDJSON append-only
└──────────────────┘
```

Marut is stateless. No daemon. No persistent state. Each invocation is independent.

---

## Privacy

Marut is local-only by design.

- **Zero network calls:** No telemetry, no analytics, no update checks, no cloud dependency
- **Single static binary:** No runtime dependencies
- **All data on-disk:** Audit logs and config are local files you control

---

## Installation

### Prerequisites

- Go 1.23+
- Node.js (for plugins)
- OpenCode or Claude Code

### Quick Start

```bash
# Clone the repo
git clone https://github.com/mikehaus/marut.git
cd marut

# Build marut binary
make build

# Install for your tool of choice
make install-opencode-global
# OR
make install-claude-global
```

### Install Targets

| Target | Platform | Scope |
|--------|----------|-------|
| `make install-opencode-global` | OpenCode | All sessions (recommended) |
| `make install-opencode-local [DIR=~/myproject]` | OpenCode | One project directory |
| `make install-claude-global` | Claude Code | All sessions (recommended) |
| `make install-claude-local [DIR=~/myproject]` | Claude Code | One project directory |
| `make uninstall-opencode-global` | OpenCode | Remove global |
| `make uninstall-opencode-local [DIR=~/myproject]` | OpenCode | Remove from project |
| `make uninstall-claude-global` | Claude Code | Remove global |
| `make uninstall-claude-local [DIR=~/myproject]` | Claude Code | Remove from project |

Local installs default to the current directory if `DIR` is not specified.

### Environment Variables

Both platforms use the same env vars. Add to your `~/.zshenv` (or `~/.bashrc`/`~/.zshrc`):

```bash
export MARUT_BIN="/path/to/marut/marut"
export MARUT_CONFIG="/path/to/marut/config/default.yaml"
export MARUT_LOG="/path/to/marut/audit.log"

# Optional — for orchestrators passing per-session flags.
# Keep to short flags only (--sim, --agent-id, --sid, etc.).
# Do NOT put --config, --log, or --platform here.
# export MARUT_ARGS="--agent-id myagent --sid mysession"
```

> **Note:** Both OpenCode and Claude Code launch as subprocesses and do not inherit env vars set inline in the terminal. These must be in your shell profile.

Then reload your shell and install:

```bash
source ~/.zshrc
cd /path/to/marut
```

#### OpenCode

```bash
make install-opencode-global              # all sessions
make install-opencode-local DIR=~/myproject  # one project only
```

#### Claude Code

```bash
make install-claude-global                # all sessions
make install-claude-local DIR=~/myproject # one project only
```

---

## Configuration

Marut uses a YAML file to define forbidden patterns and monitor phrases.

**The defaults are a starting point.** Extend the pattern list for your 
specific codebase before running agents autonomously.

**`config/default.yaml`:**

```yaml
patterns:
  # Destructive filesystem
  - "rm -rf /"
  - "rm -rf ~"
  - "~/.ssh"
  # Privilege escalation
  - "sudo su"
  - "chmod 777 /"
  # Pipe execution
  - "curl | bash"
  - "wget | sh"
```

Pattern matching is **case-insensitive** and **strips quotes/whitespace**. See `internal/matcher/matcher.go` for normalization logic.

---

## CLI Flags

```
--config       path to patterns yaml (required)
--log          path to audit.log (default: ./audit.log)
--platform     opencode | claudecode (default: opencode)
--mode         validate | monitor (default: validate, monitor is unimplemented)
--sim          SIM mode: log everything, never block. Run this first to tune your pattern list before enforcing. 
--kill-agent   on match, SIGTERM the parent process
--agent-id     agent identifier (default: "default")
--sid          tmux session id (default: "none")
--agent-seq    attempt number on this worktree (default: 1)
```

**Example:**

```bash
echo '{"tool":"bash","raw_input":"rm -rf ~/.ssh"}' | \
  marut --config config/default.yaml --platform opencode
```

Exit code: `2` (blocked)

---

## Audit Log Schema

Marut writes one NDJSON entry per tool call:

```json
{
  "uid": "01KM8ZRB2ZNF4Q1W7GJJJF78PQ",
  "timestamp": "2026-03-21T20:00:32.351813Z",
  "level": "DENY",
  "agent_id": "default",
  "pid": 4893,
  "sid": "none",
  "context": {
    "cwd": "/Users/you/project",
    "agent_seq": 1
  },
  "event": {
    "type": "shell_command",
    "tool": "bash",
    "raw_input": "rm -rf /Users/you/project",
    "match_pattern": "rm -rf /",
    "mode": "validate"
  },
  "outcome": {
    "action": "block",
    "exit_code": 2,
    "latency_ms": 0.203917,
    "message": "Deterministic block: pattern \"rm -rf /\" is forbidden."
  }
}
```

**Levels:** `DENY`, `WARN`, `PASS`, `SIM`  
**Actions:** `block`, `warn`, `kill_agent`, `pass`

---

## Development

```bash
# Run tests
make test

# Build binary
make build

# Clean artifacts
make clean
```

**Repo structure:**

```
marut/
  cmd/marut/main.go          # CLI entrypoint
  internal/
    config/loader.go         # YAML config loader
    parser/                  # Platform-specific normalizers
      opencode.go            # OpenCode payload parser
      claudecode.go          # Claude Code payload parser
    matcher/matcher.go       # Aho-Corasick wrapper
    logger/audit.go          # NDJSON audit logger
    cli/entrypoint.go        # Shared logic (BuildEntry, EventType)
  schema/event.go            # AuditEntry types
  opencode-plugin/           # OpenCode TypeScript shim
  claudecode-plugin/         # Claude Code plugin shim
  config/default.yaml        # Default forbidden patterns
```

---

## How It Works

### 1. Pattern Matching (Aho-Corasick)

Marut uses [Cloudflare's Aho-Corasick](https://github.com/cloudflare/ahocorasick) implementation for multi-pattern matching in O(n) time. All inputs are normalized before matching:

- Lowercased
- Quotes stripped (`"` and `'`)
- Whitespace collapsed

This prevents trivial bypasses like `"rm" '-rf' '~/.ssh'`.

### 2. Plugin Shims

Marut uses platform-specific plugins to intercept tool calls:

#### OpenCode Plugin

The TypeScript shim (`opencode-plugin/index.ts`) hooks into OpenCode's `tool.execute.before` event:

```typescript
"tool.execute.before": async (input, output) => {
  const normalized = {
    tool: input.tool,
    raw_input: output.args.command ?? output.args.filePath ?? JSON.stringify(output.args),
    cwd: directory,
    worktree: worktree,
  };
  const code = await invokeMarut(normalized);
  if (code === 2) {
    throw new Error("Blocked by Marut: forbidden command");
  }
}
```

#### Claude Code Plugin

The Claude Code plugin (`claudecode-plugin/`) uses command hooks defined in `hooks/hooks.json`:

```bash
# Claude Code sends JSON to stdin
PreToolUse event → marut-wrapper.sh → marut binary → exit code
```

The wrapper script (`marut-wrapper.sh`) pipes Claude Code's hook payload to marut and maps exit codes:
- Exit 0 → Allow execution
- Exit 2 → Block (shows error to Claude)
- Exit 1 → Internal error

### 3. Audit Trail

Every tool call is logged to `audit.log` as NDJSON. Use this for:

- Debugging blocks
- Observing agent behavior
- Cost/latency tracking
- Post-session analysis

---

## Known Limitations

**Marut is a first line of defense against "dumb" autonomous agents making accidental mistakes.** It is not designed to stop sophisticated adversaries.

### What Marut Blocks

- Trivial bypasses: quoted commands (`"rm" '-rf' ~/.ssh`), case variations (`RM -RF /`)
- Accidental damage: agents hallucinating destructive commands
- Known-bad patterns: `rm -rf /`, `curl | bash`, `~/.ssh` access

Update the policy `yaml` file to add your own patterns.

### What Marut Does Not Block

- **Advanced obfuscation:** Base64 encoding, command substitution (`$(echo rm)`, backticks), hex escapes
- **Indirect execution:** Scripts written to disk then executed, `eval`, `source`, aliases
- **Network exfiltration:** Data sent over HTTP/DNS if not explicitly forbidden
- **Privilege escalation via exploits:** Container escapes, kernel vulnerabilities
- **Intentional adversaries:** A motivated attacker with knowledge of the pattern list can bypass detection

### Specific Limitations

- **Pattern specificity:** `"curl | bash"` only matches when `curl` is directly adjacent to `|`. Use `"| bash"` for broader coverage.
- **Monitor mode unimplemented:** Flag exists but phrase matching and rolling count logic are not wired (Block 11). Use `--mode validate` only.

### Defense in Depth

Marut is one layer. For production deployments or defense against adversaries, add:

- **Containers:** Docker/Podman with restricted capabilities
- **VMs:** Firecracker, gVisor for stronger isolation
- **seccomp/AppArmor:** Kernel-level syscall filtering
- **Network policies:** Firewall rules, DNS filtering
- **Human review:** Quality gates before auto-merge

---

## Roadmap

- [ ] **Monitor mode:** Hallucination phrase tracking with rolling count and threshold-based warnings (flag exists, logic unimplemented)
- [ ] **Cost estimation:** Populate `savings` field in audit log with estimated tokens/cost saved per block (flag and temp pricing map exist, calculation not wired)

---

## License

**MIT License** - See [LICENSE](LICENSE)

*Marut was built by a solo developer to make autonomous coding agents safer to run while grabbing a coffee.*

---

**In the cosmology of Mechanus, a Marut does not negotiate; it simply executes the terms of the contract.**
