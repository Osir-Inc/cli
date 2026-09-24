package cmd

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

// A failed command must always tell the user why, exactly once. Cobra's own errors (wrong
// argument count, unknown flag or command) and command paths that return an error without
// printing it used to exit 1 silently, because SilenceErrors is on.
func TestFailuresArePrintedOnce(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		usageHint bool
	}{
		{[]string{"tunnel"}, true},
		{[]string{"tunnel", "3000", "--bogus"}, true},
		{[]string{"bogus"}, true},
		{[]string{"tunnel", "99999"}, false},          // the command prints this one itself
		{[]string{"auth", "login", "-u", "x"}, false}, // no terminal: password read fails unprinted
	} {
		r, w, _ := os.Pipe()
		stderr, stdin, argv := os.Stderr, os.Stdin, os.Args
		devnull, _ := os.Open(os.DevNull)
		os.Stderr, os.Stdin, os.Args = w, devnull, append([]string{"osir"}, tc.args...)
		err := ExecuteContext(context.Background())
		w.Close()
		devnull.Close()
		os.Stderr, os.Stdin, os.Args = stderr, stdin, argv
		out, _ := io.ReadAll(r)

		if err == nil {
			t.Errorf("%v: expected an error", tc.args)
		}
		if n := strings.Count(string(out), "[ERROR]"); n != 1 {
			t.Errorf("%v: printed %d errors, want exactly 1: %q", tc.args, n, out)
		}
		if tc.usageHint != strings.Contains(string(out), "--help' for usage") {
			t.Errorf("%v: usage hint shown = %v, want %v: %q", tc.args, !tc.usageHint, tc.usageHint, out)
		}
	}
}
