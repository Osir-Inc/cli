package cmd

import (
	"testing"
)

func TestNewRootCmd_NilApp(t *testing.T) {
	root := NewRootCmd(nil)
	if root == nil {
		t.Fatal("NewRootCmd(nil) returned nil")
	}
	if root.Use != "osir" {
		t.Errorf("expected Use='osir', got '%s'", root.Use)
	}
}

func TestNewRootCmd_AllCommandGroupsRegistered(t *testing.T) {
	root := NewRootCmd(nil)
	expected := []string{
		"auth", "domain", "dns", "billing", "contact",
		"audit", "account", "catalog", "suggest", "vps", "completion",
	}
	cmds := root.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}
	for _, e := range expected {
		if !names[e] {
			t.Errorf("missing command group: %s", e)
		}
	}
}

func TestNewRootCmd_WithApp(t *testing.T) {
	app := &App{}
	root := NewRootCmd(app)
	if root == nil {
		t.Fatal("NewRootCmd(app) returned nil")
	}
	if root.Use != "osir" {
		t.Errorf("expected Use='osir', got '%s'", root.Use)
	}
}

func TestNewRootCmd_NoShellCommandInTree(t *testing.T) {
	root := NewRootCmd(nil)
	for _, c := range root.Commands() {
		if c.Name() == "shell" {
			t.Error("shell command should not be in the factory-built tree")
		}
	}
}

func TestNewRootCmd_FreshTreeEachCall(t *testing.T) {
	root1 := NewRootCmd(nil)
	root2 := NewRootCmd(nil)
	if root1 == root2 {
		t.Error("NewRootCmd should return a new instance each call")
	}
}

func TestNewRootCmd_SubcommandCounts(t *testing.T) {
	root := NewRootCmd(nil)

	tests := []struct {
		name     string
		expected int
	}{
		{"domain", 12},
		{"dns", 11},
		{"billing", 12},
		{"contact", 6},
		{"audit", 3},
		{"account", 2},
		{"catalog", 2},
		{"suggest", 7},
		{"vps", 10},
		{"auth", 3},
	}

	cmds := make(map[string]int)
	for _, c := range root.Commands() {
		cmds[c.Name()] = len(c.Commands())
	}

	for _, tt := range tests {
		got, ok := cmds[tt.name]
		if !ok {
			t.Errorf("command group %s not found", tt.name)
			continue
		}
		if got != tt.expected {
			t.Errorf("%s: expected %d subcommands, got %d", tt.name, tt.expected, got)
		}
	}
}
