package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/osir/cli/internal/api"
	"github.com/osir/cli/internal/auth"
	"github.com/osir/cli/internal/config"
	"github.com/osir/cli/internal/output"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

// addShellCommand registers the shell command on the given parent.
// Only called for non-interactive mode (so "shell" doesn't appear inside the shell).
func addShellCommand(parent *cobra.Command) {
	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Launch interactive shell with tab completion and ? help",
		Long: `Launch an interactive shell with tab completion, command history,
and context-sensitive help. Press Tab to complete commands, ? to show
available options. Type 'exit' or 'quit' to leave the shell.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShell()
		},
	}
	parent.AddCommand(shellCmd)
}

func runShell() error {
	// Create App ONCE for the lifetime of the shell session
	cfg := config.Load()
	out := output.New(outputFormat == "json")
	session := auth.NewSession(cfg)
	client := api.NewClient(cfg, session)
	client.Version = appVersion
	if cred := auth.LoadCredentials(); cred != nil {
		session.Restore(cred)
	}
	app := &App{Config: cfg, Session: session, Client: client, Output: out}

	// Create the console instance
	c := console.New("osir")
	c.NewlineBefore = true
	c.NewlineAfter = true

	// Configure the active menu
	menu := c.ActiveMenu()

	// Bind the command tree factory
	menu.SetCommands(shellCommandFactory(app))

	// Prompt
	p := menu.Prompt()
	p.Primary = func() string { return "\x1b[36mosir\x1b[0m> " }
	p.Right = func() string {
		if session.IsAuthenticated() {
			cred := session.GetCredential()
			if cred != nil {
				return fmt.Sprintf("\x1b[32m%s\x1b[0m", cred.Username)
			}
		}
		return "\x1b[33mnot authenticated\x1b[0m"
	}

	// Persistent history at ~/.osir/shell_history
	home, _ := os.UserHomeDir()
	historyDir := filepath.Join(home, ".osir")
	os.MkdirAll(historyDir, 0700)
	menu.AddHistorySourceFile("osir history", filepath.Join(historyDir, "shell_history"))

	// Ctrl+D exits the shell
	menu.AddInterrupt(io.EOF, func(c *console.Console) {
		fmt.Println("Goodbye!")
		os.Exit(0)
	})

	// Bind ? to show completions (same as Tab but triggered by ?)
	shell := c.Shell()
	shell.Config.Bind("emacs", "?", "possible-completions", false)
	shell.Config.Bind("vi-insert", "?", "possible-completions", false)

	// Welcome banner
	c.SetPrintLogo(func(_ *console.Console) {
		fmt.Println("OSIR Interactive Shell v" + appVersion)
		fmt.Println("Type 'help' for commands, Tab or '?' for completions, 'exit' to quit.")
		fmt.Println()
	})

	return c.Start()
}

// shellCommandFactory returns the console.Commands factory that produces a
// fresh Cobra tree on each REPL iteration, with the pre-built App injected.
func shellCommandFactory(app *App) console.Commands {
	return func() *cobra.Command {
		rootCmd := NewRootCmd(app)

		// Add shell-only commands
		rootCmd.AddCommand(&cobra.Command{
			Use:     "exit",
			Aliases: []string{"quit"},
			Short:   "Exit the interactive shell",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Goodbye!")
				os.Exit(0)
				return nil
			},
		})

		rootCmd.AddCommand(&cobra.Command{
			Use:   "clear",
			Short: "Clear the screen",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Print("\033[H\033[2J")
				return nil
			},
		})

		return rootCmd
	}
}
