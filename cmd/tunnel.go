package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/osir/cli/internal/tunnel"
	"github.com/spf13/cobra"
)

func addTunnelCommands(parent *cobra.Command) {
	var server, auth string

	tunnelCmd := &cobra.Command{
		Use:   "tunnel [target]",
		Short: "Expose a local web server on a public HTTPS URL",
		Long: `Forward a public HTTPS address to a web server running on this machine, so you
can share work in progress, test webhooks, or demo from a laptop.

The URL is random and temporary: it lives as long as the command runs. Opening a
tunnel requires an OSIR account ('osir auth login'), and which account opened
which address is recorded, so abuse can be traced. Anyone with the link can reach
your local server, so only expose what you mean to share.

The target can be a port, a host:port, or a full URL:
  3000                     localhost:3000 over http
  localhost:8080           the same, written out
  https://localhost:8443   a local server that speaks TLS

Requests reach your app unchanged, including the public Host header
(e.g. Host: crisp-willow-knix.osir.run). Dev servers that check the host
reject it until you allow the domain, for example:
  Vite                     server.allowedHosts: ['.osir.run']
  webpack-dev-server       allowedHosts: ['.osir.run']`,
		Example: `  osir tunnel 3000
  osir tunnel http://localhost:8080
  osir tunnel 3000 --server https://osir.run
  osir tunnel 3000 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			edge := server
			if edge == "" {
				edge = app.Config.TunnelURL
			}
			// Read here, not as the flag default, so --help never prints the token.
			if auth == "" {
				auth = os.Getenv("OSIR_TUNNEL_AUTH")
			}

			target, err := tunnel.ParseTarget(args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			// Ctrl+C should close the tunnel, not kill the process mid-write.
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			// A tunnel is opened by an account, so abuse can be traced to one; the edge
			// rejects the connection when this token is missing or invalid.
			if !app.Session.IsAuthenticated() {
				err := fmt.Errorf("you need to sign in first: run 'osir auth login'")
				app.Output.PrintError(err.Error())
				return err
			}

			opts := tunnel.Options{
				Server: edge,
				Target: target,
				Auth:   auth,
				Token:  app.Session.GetToken,
				OnReady: func(url string) {
					if app.Output.IsJSON() {
						app.Output.PrintResult(map[string]string{
							"status": "ready",
							"url":    url,
							"target": target.String(),
						})
						return
					}
					line := strings.Repeat("-", len(url)+4)
					app.Output.Println("")
					app.Output.Println("+" + line + "+")
					app.Output.Println("|  " + url + "  |")
					app.Output.Println("+" + line + "+")
					app.Output.Println(fmt.Sprintf("Forwarding %s -> %s", url, target))
					app.Output.Println("Anyone with this link can reach your app. Press Ctrl+C to stop.")
					app.Output.Println("")
				},
				// Status lines are news (a restart, a reconnect), not failures, so they go to
				// stderr without the [ERROR] prefix. Real failures come back from Run.
				OnStatus: func(msg string) {
					if !app.Output.IsJSON() {
						fmt.Fprintln(cmd.ErrOrStderr(), msg)
					}
				},
			}

			if err := tunnel.Run(ctx, opts); err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if !app.Output.IsJSON() {
				app.Output.Println("Tunnel closed.")
			}
			return nil
		},
	}

	tunnelCmd.Flags().StringVar(&server, "server", "", "Tunnel edge to use (default https://osir.run, env OSIR_TUNNEL_SERVER)")
	tunnelCmd.Flags().StringVar(&auth, "auth", "", "Auth token, for private tunnel edges (env OSIR_TUNNEL_AUTH)")

	parent.AddCommand(tunnelCmd)
}
