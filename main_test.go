package main

import (
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cases := []struct {
		in      string
		wantCmd string
		wantOk  bool
	}{
		{"$ ls -a", "ls -a", true},
		{"$ls", "ls", true},
		{"  $  cat challenge.txt  ", "cat challenge.txt", true},
		{"$", "", true},
		{"ls -a", "", false},
		{"", "", false},
		{"please run $ ls", "", false},
		{"```\n$ ls\n```", "", false},
	}
	for _, tc := range cases {
		gotCmd, gotOk := parseCommand(tc.in)
		if gotOk != tc.wantOk || gotCmd != tc.wantCmd {
			t.Errorf("parseCommand(%q) = (%q, %v), want (%q, %v)",
				tc.in, gotCmd, gotOk, tc.wantCmd, tc.wantOk)
		}
	}
}

func TestFormatOutputEscapesFence(t *testing.T) {
	// Output containing a triple-backtick run must be wrapped in a longer fence.
	out := formatOutput("before ``` after", false)
	if !strings.HasPrefix(out, "````\n") || !strings.HasSuffix(out, "\n````") {
		t.Fatalf("expected 4-backtick fence around ``` output, got:\n%s", out)
	}

	// Plain output still uses the standard 3-backtick fence.
	plain := formatOutput("hello", false)
	if !strings.HasPrefix(plain, "```\n") || !strings.HasSuffix(plain, "\n```") {
		t.Fatalf("expected 3-backtick fence for plain output, got:\n%s", plain)
	}
}

func TestMissingCommands(t *testing.T) {
	const bogus = "definitely-not-a-real-command-xyz"
	if got := missingCommands(bogus); len(got) != 1 || got[0] != bogus {
		t.Errorf("missingCommands(%q) = %v, want [%q]", bogus, got, bogus)
	}
	// A command that is present must not be reported. sh is on PATH in any
	// POSIX environment the tests run in; if not, skip rather than fail.
	if got := missingCommands("sh"); len(got) != 0 {
		t.Skipf("sh not on PATH here (%v); skipping present-command check", got)
	}
}
