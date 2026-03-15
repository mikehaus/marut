package schema

type Level string

const (
	LevelDeny Level = "DENY"
	LevelWarn Level = "WARN"
	LevelPass Level = "PASS"
	LevelSim  Level = "SIM"
)

type Action string

const (
	ActionBlock Action = "block"
	ActionWarn  Action = "warn"
	ActionKill  Action = "kill_agent"
	ActionPass  Action = "pass"
)

type EventType string

const (
	EventShellCommand  EventType = "shell_command"
	EventFileAccess    EventType = "file_access"
	EventNetworkAccess EventType = "network_access"
)

type Context struct {
	CWD          string  `json:"cwd"`
	GitBranch    string  `json:"git_branch"`
	RatchetClick string  `json:"ratchet_click,omitempty"`
	AgentSeq     int     `json:"agent_seq"`
	Temp         float64 `json:"temp,omitempty"`
}

type Event struct {
	Type         EventType `json:"type"`
	Tool         string    `json:"tool"`
	RawInput     string    `json:"raw_input"`
	MatchPattern string    `json:"match_pattern,omitempty"`
	Mode         string    `json:"mode"`
}

type Outcome struct {
	Action    Action  `json:"action"`
	ExitCode  int     `json:"exit_code"`
	LatencyMs float64 `json:"latency_ms"`
	Message   string  `json:"message"`
}

type AuditEntry struct {
	UID       string  `json:"uid"`       // ULID — sortable, unique, generated at write time
	Timestamp string  `json:"timestamp"` // ISO8601 — human readable
	Level     Level   `json:"level"`
	AgentID   string  `json:"agent_id"`
	PID       int     `json:"pid"`
	SID       string  `json:"sid"`
	Context   Context `json:"context"`
	Event     Event   `json:"event"`
	Outcome   Outcome `json:"outcome"`
}
