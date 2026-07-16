package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	cases := []struct {
		in   string
		want string
	}{
		{"~", home},
		{"~/.ssh/id_ed25519.pub", filepath.Join(home, ".ssh/id_ed25519.pub")},
		// Another user's home: not ours to resolve, so it must pass through untouched rather than
		// becoming $HOME/otheruser/.ssh/id.pub.
		{"~otheruser/.ssh/id.pub", "~otheruser/.ssh/id.pub"},
		{"/abs/path.pub", "/abs/path.pub"},
		{"relative.pub", "relative.pub"},
	}
	for _, c := range cases {
		if got := expandHome(c.in); got != c.want {
			t.Errorf("expandHome(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// A private key passed to 'ssh-keys add' would be stored and shipped to the server, so the guard
// that rejects it is worth pinning down. The mock backend has no VPS methods: reaching the client at
// all would panic, which is the point — the guard must reject before any upload.
func TestSshKeyAdd_RejectsPrivateKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519")
	priv := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXk=\n-----END OPENSSH PRIVATE KEY-----\n"
	if err := os.WriteFile(keyPath, []byte(priv), 0o600); err != nil {
		t.Fatal(err)
	}

	app := newTestApp(&mockBackend{})
	out, err := executeCommand(app, "vps", "ssh-keys", "add", "--name", "laptop", "--key-file", keyPath)
	if err == nil {
		t.Fatalf("expected an error for a private key, got nil (output: %s)", out)
	}
	if !strings.Contains(err.Error(), "PRIVATE key") {
		t.Errorf("expected the private-key guard to fire, got: %v", err)
	}
}
