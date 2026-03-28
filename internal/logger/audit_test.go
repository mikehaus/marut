package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mikehaus/marut/schema"
)

// makeEntry builds a minimal AuditEntry for testing. UID and Timestamp
// are intentionally left blank — the logger assigns them at write time.
func makeEntry(level schema.Level, action schema.Action, tool, rawInput string) schema.AuditEntry {
	return schema.AuditEntry{
		Level:   level,
		AgentID: "test-agent",
		PID:     12345,
		SID:     "test-session",
		Context: schema.Context{
			CWD:      "/tmp/test",
			AgentSeq: 1,
		},
		Event: schema.Event{
			Type:     schema.EventShellCommand,
			Tool:     tool,
			RawInput: rawInput,
		},
		Outcome: schema.Outcome{
			Action:    action,
			ExitCode:  0,
			LatencyMs: 0.5,
			Message:   "test",
		},
	}
}

// readLines reads all non-empty lines from a file.
func readLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open log file: %v", err)
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := sc.Text(); line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan log file: %v", err)
	}
	return lines
}

// TestWrite_ThreeEntries writes three entries, reads the file back, and
// asserts three valid JSON lines each with a non-empty UID and Timestamp.
func TestWrite_ThreeEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := New(path, false)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	entries := []schema.AuditEntry{
		makeEntry(schema.LevelPass, schema.ActionPass, "bash", "go test ./..."),
		makeEntry(schema.LevelDeny, schema.ActionBlock, "bash", "rm -rf /"),
		makeEntry(schema.LevelSim, schema.ActionPass, "bash", "ls -la"),
	}
	for _, e := range entries {
		if err := l.Write(e); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	lines := readLines(t, path)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var got schema.AuditEntry
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}
		if got.UID == "" {
			t.Errorf("line %d: UID is empty", i)
		}
		if got.Timestamp == "" {
			t.Errorf("line %d: Timestamp is empty", i)
		}
		if got.Level != entries[i].Level {
			t.Errorf("line %d: level: want %q, got %q", i, entries[i].Level, got.Level)
		}
	}
}

// TestWrite_AppendsNotTruncates verifies that a second Logger opened on
// the same path appends rather than overwrites existing content.
func TestWrite_AppendsNotTruncates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")

	l1, _ := New(path, false)
	if err := l1.Write(makeEntry(schema.LevelPass, schema.ActionPass, "bash", "first")); err != nil {
		t.Fatal(err)
	}

	l2, _ := New(path, false)
	if err := l2.Write(makeEntry(schema.LevelDeny, schema.ActionBlock, "bash", "second")); err != nil {
		t.Fatal(err)
	}

	lines := readLines(t, path)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines after two separate loggers, got %d", len(lines))
	}
}

// TestWrite_CreatesFileIfMissing verifies that the log file is created on
// first write when it does not yet exist.
func TestWrite_CreatesFileIfMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subdir", "audit.log")
	// Create the parent directory but not the file.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	l, _ := New(path, false)
	if err := l.Write(makeEntry(schema.LevelPass, schema.ActionPass, "bash", "ls")); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

// TestWrite_UIDIsUnique checks that each entry receives a distinct ULID.
func TestWrite_UIDIsUnique(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, _ := New(path, false)

	for i := 0; i < 5; i++ {
		if err := l.Write(makeEntry(schema.LevelPass, schema.ActionPass, "bash", "ls")); err != nil {
			t.Fatal(err)
		}
	}

	lines := readLines(t, path)
	seen := make(map[string]bool)
	for i, line := range lines {
		var got schema.AuditEntry
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d: invalid JSON: %v", i, err)
		}
		if seen[got.UID] {
			t.Errorf("duplicate UID %q on line %d", got.UID, i)
		}
		seen[got.UID] = true
	}
}

// TestWrite_TimestampIsISO8601 verifies the timestamp parses as RFC3339Nano.
func TestWrite_TimestampIsISO8601(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, _ := New(path, false)

	if err := l.Write(makeEntry(schema.LevelPass, schema.ActionPass, "bash", "ls")); err != nil {
		t.Fatal(err)
	}

	lines := readLines(t, path)
	var got schema.AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatal(err)
	}

	if _, err := time.Parse(time.RFC3339Nano, got.Timestamp); err != nil {
		t.Errorf("Timestamp %q does not parse as RFC3339Nano: %v", got.Timestamp, err)
	}
}

// TestWriteSIMRaw verifies the SIM_RAW line is written with the correct prefix.
func TestWriteSIMRaw(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, _ := New(path, true)

	raw := []byte(`{"tool":"bash","raw_input":"ls -la"}`)
	if err := l.WriteSIMRaw(raw); err != nil {
		t.Fatalf("WriteSIMRaw: %v", err)
	}

	lines := readLines(t, path)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "SIM_RAW:") {
		t.Errorf("expected SIM_RAW: prefix, got %q", lines[0])
	}
	if !strings.Contains(lines[0], "ls -la") {
		t.Errorf("expected raw payload in SIM_RAW line, got %q", lines[0])
	}
}

// TestNew_EmptyPathReturnsError verifies that an empty path is rejected
// at construction time rather than silently failing at write time.
func TestNew_EmptyPathReturnsError(t *testing.T) {
	_, err := New("", false)
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}
