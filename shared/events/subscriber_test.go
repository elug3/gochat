package events

import (
	"context"
	"testing"
)

func TestWildcardMatch(t *testing.T) {
	testCases := map[string]struct {
		pattern string
		subject string
		match   bool
	}{
		"exact match":    {"a.b.c", "a.b.c", true},
		"wildcard match": {"a.*.c", "a.b.c", true},

		"'>' matches all parts":       {">", "a.b.c.d.e", true},
		"'>' matches remaining parts": {"a.>", "a.b.c.d", true},

		"subject cannot have wildcards matching pattern": {"a.b.c", "a.*.c", false},
		"'>' must match at least one part":               {"a.b.>", "a.b", false},
		"extra subject parts without '>'":                {"a.*", "a.b.c", false},
		"pattern longer than subject":                    {"a.b.c", "a.b", false},
		"'>' not last does not match":                    {"a.>.c", "a.b.c", false},
		"subject shorter than pattern with '*'":          {"a.*.c", "a.c", false},
		"empty subject part":                             {"a.*.c", "a..c", false},

		"empty strings":               {"", "", false},
		"empty pattern":               {"", "a.b.c", false},
		"empty subject":               {"a.b.c", "", false},
		"wildcard with empty subject": {"*", "", false},

		"empty pattern parts":          {"...", "...", false},
		"empty pattern parts with '>'": {">", "...", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			match, _ := wildcardMatch(tc.subject, tc.pattern)
			if match != tc.match {
				t.Errorf("expected pattern %q to match subject %q = %v, got %v",
					tc.pattern, tc.subject, tc.match, match)
			}
		})
	}
}

type testHandler struct {
	id string
}

func (h *testHandler) Handle(_ context.Context, _ string, _ []byte) error {
	return nil
}

func TestGetHandlerSelection(t *testing.T) {
	s := &Subscriber{
		handlers: make(map[string]Handler),
	}

	exact := &testHandler{id: "exact"}
	wildcard := &testHandler{id: "wildcard"}
	greater := &testHandler{id: "greater"}
	earlierWildcard := &testHandler{id: "earlierWildcard"}

	s.Handle("a.b.c", exact)
	s.Handle("a.*.c", wildcard)
	s.Handle("a.b.>", greater)
	s.Handle("*.b.c", earlierWildcard)

	h, ok := s.getHandler("a.b.c")
	if !ok || h != exact {
		t.Fatalf("expected exact handler to be selected, got %v (ok=%v)", h, ok)
	}

	h, ok = s.getHandler("a.b.d")
	if !ok || h != greater {
		t.Fatalf("expected '>' handler to be selected, got %v (ok=%v)", h, ok)
	}

	h, ok = s.getHandler("a.x.c")
	if !ok || h != wildcard {
		t.Fatalf("expected '*' handler to be selected, got %v (ok=%v)", h, ok)
	}

	h, ok = s.getHandler("x.b.c")
	if !ok || h != earlierWildcard {
		t.Fatalf("expected earlier wildcard handler to be selected, got %v (ok=%v)", h, ok)
	}
}

func TestWildcardMatchPriority(t *testing.T) {
	// lower priority value means more specific match
	// if all parts exact matchs, prio is zero

	testCases := map[string]struct {
		pattern string
		subject string
		prio    int
	}{
		"exact match":         {"a.b.c", "a.b.c", 0},
		"one '*' wildcard":    {"a.b.*", "a.b.c", 1},
		"one '>' wildcard":    {"a.b.>", "a.b.c", 2},
		"0 1 0":               {"a.*.c", "a.b.c", 3},
		"1 0 0":               {"*.b.c", "a.b.c", 9},
		"1 1 1":               {"*.*.*", "a.b.c", 13},
		"'>' matches 3 parts": {">", "a.b.c", 26},
		"'>' matches 5 parts": {">", "a.b.c.d.e", 242}, // 2*3^4 + 2*3^3 + 2*3^2 + 2*3^1 + 2*3^0 = 242

		"no match priority":     {"a.b", "a.b.c", -1},
		"'>' not last priority": {"a.>.c", "a.b.c", -1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			match, prio := wildcardMatch(tc.subject, tc.pattern)
			if !match && tc.prio != -1 {
				t.Errorf("expected pattern %q to match subject %q with priority %d, but did not match",
					tc.pattern, tc.subject, tc.prio)
			} else if prio != tc.prio {
				t.Errorf("expected pattern %q to match subject %q with priority %d, got %d",
					tc.pattern, tc.subject, tc.prio, prio)
			} else if match && tc.prio == -1 {
				t.Errorf("expected pattern %q to not match subject %q, but it matched with priority %d",
					tc.pattern, tc.subject, prio)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	testCases := map[string]struct {
		pattern string
		valid   bool
	}{
		"valid exact pattern":    {"a.b.c", true},
		"valid wildcard pattern": {"a.*.c", true},
		"valid '>' pattern":      {"a.>", true},
		"valid only '>' pattern": {">", true},

		"invalid empty pattern":      {"", false},
		"invalid empty pattern part": {"a..c", false},
		"invalid '>' not last":       {"a.>.c", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validatePattern(tc.pattern)
			if (err == nil) != tc.valid {
				t.Errorf("expected pattern %q valid=%v, got error: %v",
					tc.pattern, tc.valid, err)
			}
		})
	}
}
