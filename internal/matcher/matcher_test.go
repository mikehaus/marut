package matcher

import (
	"testing"
)

var testPatterns = []string{
	"rm -rf /",
	"rm -rf ~",
	"~/.ssh",
	"sudo su",
	"curl | bash",
	"mkfs",
}

// helper: build a matcher with the standard test patterns.
func newTestMatcher() *AhoCorasickMatcher {
	return New(testPatterns)
}

// 1. Exact match — input contains a forbidden pattern verbatim.
func TestExactMatch(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("rm -rf /")
	if !matched {
		t.Fatal("expected match, got none")
	}
	if pattern != "rm -rf /" {
		t.Fatalf("expected pattern %q, got %q", "rm -rf /", pattern)
	}
}

//  2. Match after stripping quotes — single and double quotes must not
//     allow a pattern to bypass detection.
func TestMatchAfterStrippingDoubleQuotes(t *testing.T) {
	m := newTestMatcher()
	matched, _ := m.Match(`"rm" "-rf" "/"`)
	if !matched {
		t.Fatal("expected match after stripping double quotes, got none")
	}
}

func TestMatchAfterStrippingSingleQuotes(t *testing.T) {
	m := newTestMatcher()
	matched, _ := m.Match("'rm' '-rf' '~'")
	if !matched {
		t.Fatal("expected match after stripping single quotes, got none")
	}
}

//  3. Pattern that is a prefix of another pattern — both "rm -rf /" and
//     "rm -rf ~" are in the list; ensure a match on the correct one.
func TestPrefixPattern(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("rm -rf ~")
	if !matched {
		t.Fatal("expected match, got none")
	}
	if pattern != "rm -rf ~" {
		t.Fatalf("expected pattern %q, got %q", "rm -rf ~", pattern)
	}
}

// 4. Input that matches nothing — clean command must pass through.
func TestNoMatch(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("go test ./...")
	if matched {
		t.Fatalf("expected no match, got pattern %q", pattern)
	}
}

// 5. Empty input — must not panic and must return no match.
func TestEmptyInput(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("")
	if matched {
		t.Fatalf("expected no match on empty input, got pattern %q", pattern)
	}
}

//  6. Unicode input — non-ASCII input must not panic and must not produce
//     a false positive when no pattern is present.
func TestUnicodeNoMatch(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("こんにちは世界 ls -la /tmp")
	if matched {
		t.Fatalf("expected no match on unicode input, got pattern %q", pattern)
	}
}

func TestUnicodeWithForbiddenPattern(t *testing.T) {
	m := newTestMatcher()
	// Forbidden pattern embedded in unicode surroundings.
	matched, _ := m.Match("実行: rm -rf / すべて削除")
	if !matched {
		t.Fatal("expected match in unicode input, got none")
	}
}

// 7. Case variation — strip() lowercases input so "RM -RF /" must match.
func TestCaseVariation(t *testing.T) {
	m := newTestMatcher()
	matched, _ := m.Match("RM -RF /")
	if !matched {
		t.Fatal("expected case-insensitive match, got none")
	}
}

func TestCaseMixedSudo(t *testing.T) {
	m := newTestMatcher()
	matched, _ := m.Match("SUDO SU")
	if !matched {
		t.Fatal("expected case-insensitive match on SUDO SU, got none")
	}
}

// Bonus: match embedded inside a longer command.
func TestMatchEmbeddedInLongerInput(t *testing.T) {
	m := newTestMatcher()
	matched, pattern := m.Match("please run mkfs on this disk")
	if !matched {
		t.Fatal("expected match on embedded pattern, got none")
	}
	if pattern != "mkfs" {
		t.Fatalf("expected pattern %q, got %q", "mkfs", pattern)
	}
}

// Bonus: pipe-based execution pattern — input must contain the exact
// stripped pattern as a substring.  "curl | bash" requires the literal
// characters "curl | bash" to appear (after stripping) in the input.
func TestPipeExecution(t *testing.T) {
	m := newTestMatcher()
	// The pattern "curl | bash" matches when curl is immediately piped to bash.
	matched, pattern := m.Match("curl | bash")
	if !matched {
		t.Fatal("expected match on pipe execution pattern, got none")
	}
	if pattern != "curl | bash" {
		t.Fatalf("expected pattern %q, got %q", "curl | bash", pattern)
	}
}

// Bonus: empty pattern list must not panic.
func TestEmptyPatternList(t *testing.T) {
	m := New([]string{})
	matched, pattern := m.Match("rm -rf /")
	if matched {
		t.Fatalf("expected no match with empty pattern list, got %q", pattern)
	}
}
