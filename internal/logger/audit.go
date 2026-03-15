package logger

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/mikehaus/marut/schema"
)

// Logger appends structured audit entries to an NDJSON file.
// It never truncates the file. Safe for sequential use within a single
// process; each invocation of the marut binary opens, writes, and closes.
type Logger struct {
	path string
	sim  bool
}

// New creates a Logger targeting path. If sim is true the logger is in
// SIM mode (callers may choose to annotate entries differently, but the
// logger itself behaves identically — that distinction belongs in main).
func New(path string, sim bool) (*Logger, error) {
	if path == "" {
		return nil, fmt.Errorf("logger: path must not be empty")
	}
	return &Logger{path: path, sim: sim}, nil
}

// Write assigns a ULID and ISO8601 timestamp to entry, serializes it as
// a single JSON line, and appends it to the log file.
// The file is created if it does not exist.
func (l *Logger) Write(entry schema.AuditEntry) error {
	now := time.Now().UTC()

	entry.UID = newULID(now)
	entry.Timestamp = now.Format(time.RFC3339Nano)

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("logger: marshal entry: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("logger: open %s: %w", l.path, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
		return fmt.Errorf("logger: write %s: %w", l.path, err)
	}
	return nil
}

// WriteSIMRaw appends the raw stdin bytes as an unstructured line prefixed
// with "SIM_RAW:" so payload shapes can be inspected during SIM mode runs.
func (l *Logger) WriteSIMRaw(raw []byte) error {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("logger: open %s: %w", l.path, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "SIM_RAW: %s\n", raw); err != nil {
		return fmt.Errorf("logger: write sim raw %s: %w", l.path, err)
	}
	return nil
}

// newULID generates a ULID using the provided time and a cryptographically
// random entropy source read directly — no monotonic wrapper. Marut exits
// after each write so sub-millisecond ordering within a single process is
// not needed, and skipping Monotonic avoids its panic-on-overflow edge case
// (2^80 increments within one millisecond). MustNew panics only if the
// system entropy source fails, which is unrecoverable in any case.
func newULID(t time.Time) string {
	return ulid.MustNew(ulid.Timestamp(t), rand.Reader).String()
}
