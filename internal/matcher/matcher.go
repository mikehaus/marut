// Copyright (c) 2026 Mike Hollingshaus
// Licensed under the MIT License
// See https://github.com/mikehollingshaus/marut/blob/main/LICENSE

package matcher

import (
	"strings"

	"github.com/cloudflare/ahocorasick"
)

// Matcher is the interface exposed to the rest of the binary.
// Match returns the first forbidden pattern found in input, or
// matched=false if the input is clean.
type Matcher interface {
	Match(input string) (matched bool, pattern string)
}

// AhoCorasickMatcher is a thin wrapper around cloudflare/ahocorasick.
// All matching runs on the stripped (normalized) form of the input so
// that trivial bypasses like quote injection or extra whitespace do not
// evade detection.
type AhoCorasickMatcher struct {
	ac       *ahocorasick.Matcher
	patterns []string
}

// New constructs an AhoCorasickMatcher from the provided pattern list.
// Patterns are stored verbatim for reporting; the trie is built from
// their stripped forms so matching is consistent with what strip()
// produces at call time.
func New(patterns []string) *AhoCorasickMatcher {
	stripped := make([]string, len(patterns))
	for i, p := range patterns {
		stripped[i] = strip(p)
	}
	return &AhoCorasickMatcher{
		ac:       ahocorasick.NewStringMatcher(stripped),
		patterns: patterns,
	}
}

// Match normalizes input and runs it through the Aho-Corasick automaton.
// Returns (true, first-matched-pattern) or (false, "") if nothing matched.
func (m *AhoCorasickMatcher) Match(input string) (bool, string) {
	hits := m.ac.Match([]byte(strip(input)))
	if len(hits) == 0 {
		return false, ""
	}
	// Any match is sufficient to trigger a block. When multiple patterns
	// hit, pick the lowest pattern-list index for determinism. Note: hits
	// contains dictionary indexes, not input positions — lowest index is
	// not the earliest match in the string.
	first := hits[0]
	for _, idx := range hits[1:] {
		if idx < first {
			first = idx
		}
	}
	return true, m.patterns[first]
}

// strip normalizes an input string before matching to reduce the surface
// area for trivial bypass techniques.  Conservative on day one — harden
// based on bypass patterns observed in audit logs over time.
func strip(input string) string {
	input = strings.ToLower(input)
	input = strings.ReplaceAll(input, `"`, "")
	input = strings.ReplaceAll(input, `'`, "")
	input = strings.Join(strings.Fields(input), " ")
	return input
}
