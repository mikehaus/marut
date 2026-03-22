// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/mikehaus/marut/internal/cli"
	"github.com/mikehaus/marut/internal/config"
	"github.com/mikehaus/marut/internal/logger"
	"github.com/mikehaus/marut/internal/matcher"
	"github.com/mikehaus/marut/internal/parser"
	"github.com/mikehaus/marut/schema"
)

// Exit codes
const (
	exitPass    = 0
	exitError   = 1
	exitBlocked = 2
)

// create max 10MB threshold to prevent arbitrarily large toolcall from exhausting memory
const maxInputBytes = 10 * 1024 * 1024

func main() {
	os.Exit(run())
}

func run() int {
	// --- Flags ---
	configPath := flag.String("config", "", "path to patterns yaml (required)")
	logPath := flag.String("log", "./audit.log", "path to audit log")
	platform := flag.String("platform", "opencode", "platform: opencode | claudecode")
	mode := flag.String("mode", "validate", "mode: validate | monitor")
	sim := flag.Bool("sim", false, "SIM mode: log everything, never block")
	killAgent := flag.Bool("kill-agent", false, "on match, SIGTERM the parent process in addition to blocking")
	agentID := flag.String("agent-id", "default", "agent identifier")
	sid := flag.String("sid", "none", "tmux session id")
	agentSeq := flag.Int("agent-seq", 1, "agent attempt number on this worktree")
	model := flag.String("model", "claude-haiku-4-5", "agent model key for cost estimation")
	flag.Parse()

	// --- Validate flags ---
	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "error: --config is required; provide a path to a patterns yaml file")
		return exitError
	}

	if *mode != "validate" && *mode != "monitor" {
		fmt.Fprintf(os.Stderr, "error: --mode must be \"validate\" or \"monitor\", got %q\n", *mode)
		return exitError
	}

	if *platform != "opencode" && *platform != "claudecode" {
		fmt.Fprintf(os.Stderr, "error: --platform must be \"opencode\" or \"claudecode\", got %q\n", *platform)
		return exitError
	}

	// --- Load config ---
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}

	// Warn if model key is unrecognised — savings calc will be skipped,
	// but this is never a fatal error.
	if _, ok := config.ModelCosts[*model]; !ok {
		fmt.Fprintf(os.Stderr, "warning: unknown --model %q; savings estimation will be skipped\n", *model)
	}
	// model is intentionally unused beyond validation until savings is wired (stretch goal).
	_ = model

	// Guard: monitor mode with no phrases is a silent no-op.
	if *mode == "monitor" && !*sim && len(cfg.MonitorPhrases) == 0 {
		fmt.Fprintln(os.Stderr, "error: --mode monitor requires monitor_phrases in config; none found")
		return exitError
	}

	// Inject runtime fields from flags.
	cfg.KillAgent = *killAgent
	cfg.AgentID = *agentID
	cfg.SID = *sid
	cfg.AgentSeq = *agentSeq

	// --- Select normalizer ---
	normalizers := map[string]parser.Normalizer{
		"opencode":   &parser.OpenCodeNormalizer{},
		"claudecode": &parser.ClaudeCodeNormalizer{},
	}
	norm := normalizers[*platform]

	log, err := logger.New(*logPath, *sim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return exitError
	}

	rawBytes, err := io.ReadAll(io.LimitReader(os.Stdin, maxInputBytes+1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: reading stdin: %v\n", err)
		return exitError
	}
	if len(rawBytes) > maxInputBytes {
		fmt.Fprintf(os.Stderr, "error: input exceeds maximum size of %d bytes\n", maxInputBytes)
		return exitBlocked
	}

	// --- SIM mode: log raw bytes and a SIM entry, then exit 0 ---
	if *sim {
		_ = log.WriteSIMRaw(rawBytes)
		tc, _ := norm.Normalize(rawBytes) // best-effort; ignore error in SIM
		entry := cli.BuildEntry(cfg, tc, *mode, schema.LevelSim, schema.ActionPass, "", exitPass, 0, "sim mode")
		_ = log.Write(entry)
		return exitPass
	}

	// --- Normalize ---
	tc, err := norm.Normalize(rawBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: normalizing payload: %v\n", err)
		return exitError
	}

	// --- Match ---
	start := time.Now()
	m := matcher.New(cfg.Patterns)
	matched, pattern := m.Match(tc.RawInput)
	latencyMs := float64(time.Since(start).Nanoseconds()) / 1e6

	// --- Determine outcome ---
	var (
		level    schema.Level
		action   schema.Action
		exitCode int
		message  string
	)

	if matched {
		level = schema.LevelDeny
		action = schema.ActionBlock
		exitCode = exitBlocked
		message = cli.BlockMessage(pattern)
	} else {
		level = schema.LevelPass
		action = schema.ActionPass
		exitCode = exitPass
		message = "allowed"
	}

	entry := cli.BuildEntry(cfg, tc, *mode, level, action, pattern, exitCode, latencyMs, message)

	// --- Write audit log ---
	if err := log.Write(entry); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing audit log: %v\n", err)
		return exitError
	}

	// --- Block path ---
	if matched {
		fmt.Fprintf(os.Stderr, "marut: blocked — %s\n", message)
		if cfg.KillAgent {
			_ = syscall.Kill(os.Getppid(), syscall.SIGTERM)
		}
		return exitBlocked
	}

	return exitPass
}
