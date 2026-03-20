package cmd

import (
	"context"
	"time"

	"github.com/osir/cli/internal/api"
	"github.com/osir/cli/internal/auth"
	"github.com/osir/cli/internal/config"
	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

// App holds all dependencies for command execution.
type App struct {
	Config  *config.Config
	Session auth.SessionManager
	Client  api.Backend
	Output  *output.Formatter
}

type contextKey string

const appKey contextKey = "app"

func getApp(cmd *cobra.Command) *App {
	return cmd.Context().Value(appKey).(*App)
}

var (
	outputFormat string
	appVersion   string
	verbose      bool
	timeout      time.Duration
)

func SetVersion(v string) {
	appVersion = v
}

// NewRootCmd creates a fresh command tree. If app is non-nil (shell mode),
// it is injected into the context directly; otherwise PersistentPreRunE
// creates the App on-demand (normal CLI mode).
func NewRootCmd(app *App) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "osir",
		Short:         "OSIR Domain Registrar CLI",
		Long:          "A command-line tool for managing domains, DNS, billing, contacts, and more via the OSIR platform.",
		Version:       appVersion,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	if app != nil {
		// Shell mode: inject pre-existing App into every command's context
		rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			// Sync output mode from flag on each command invocation
			app.Output.SetJSON(outputFormat == "json")
			cmd.SetContext(context.WithValue(cmd.Context(), appKey, app))
			return nil
		}
	} else {
		// Normal CLI mode: create App fresh
		rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			out := output.New(outputFormat == "json")
			session := auth.NewSession(cfg)
			session.Verbose = verbose
			client := api.NewClient(cfg, session)
			client.Verbose = verbose
			client.Version = appVersion
			if timeout > 0 {
				client.HTTPClient.Timeout = timeout
			}

			if cred := auth.LoadCredentials(); cred != nil {
				session.Restore(cred)
			}

			a := &App{
				Config:  cfg,
				Session: session,
				Client:  client,
				Output:  out,
			}
			cmd.SetContext(context.WithValue(cmd.Context(), appKey, a))
			return nil
		}
	}

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text or json")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show timing and debug info")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP request timeout (e.g. 10s, 1m)")
	rootCmd.SetVersionTemplate("osir {{.Version}}\n")

	// Register all command groups
	addAuthCommands(rootCmd)
	addDomainCommands(rootCmd)
	addDnsCommands(rootCmd)
	addBillingCommands(rootCmd)
	addContactCommands(rootCmd)
	addAuditCommands(rootCmd)
	addAccountCommands(rootCmd)
	addCatalogCommands(rootCmd)
	addSuggestCommands(rootCmd)
	addVpsCommands(rootCmd)
	addCompletionCommands(rootCmd)

	return rootCmd
}

// Execute runs the CLI in normal (non-interactive) mode.
func Execute() error {
	return ExecuteContext(context.Background())
}

// ExecuteContext runs the CLI with the given context for signal handling.
func ExecuteContext(ctx context.Context) error {
	rootCmd := NewRootCmd(nil)
	// Add shell command only in non-interactive mode
	addShellCommand(rootCmd)
	return rootCmd.ExecuteContext(ctx)
}
