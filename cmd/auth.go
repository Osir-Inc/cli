package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func addAuthCommands(parent *cobra.Command) {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication management",
	}

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to OSIR platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			device, _ := cmd.Flags().GetBool("device")
			username, _ := cmd.Flags().GetString("username")

			if device {
				return loginDevice(ctx, app)
			}
			return loginPassword(ctx, app, username)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)

			cred := app.Session.GetCredential()
			if cred == nil || !app.Session.IsAuthenticated() {
				app.Output.Println("Not authenticated. Run 'osir auth login' to log in.")
				return nil
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(map[string]any{
					"authenticated": true,
					"username":      cred.Username,
					"expiresIn":     cred.ExpiresInSeconds(),
					"loginMethod":   cred.LoginMethod,
				})
			} else {
				app.Output.Println("Authenticated")
				app.Output.PrintKeyValue("Username", cred.Username)
				app.Output.PrintKeyValue("Method", cred.LoginMethod)
				secs := cred.ExpiresInSeconds()
				app.Output.PrintKeyValue("Token expires in", fmt.Sprintf("%dm %ds", secs/60, secs%60))
			}
			return nil
		},
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if err := app.Session.Logout(ctx); err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintSuccess("Logged out successfully")
			return nil
		},
	}

	loginCmd.Flags().BoolP("device", "d", false, "Use OAuth 2.0 device authorization flow (browser-based)")
	loginCmd.Flags().StringP("username", "u", "", "Username for password login")

	authCmd.AddCommand(loginCmd, statusCmd, logoutCmd)
	parent.AddCommand(authCmd)
}

func loginPassword(ctx context.Context, app *App, username string) error {
	if username == "" {
		fmt.Print("Username: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		username = strings.TrimSpace(input)
	}

	fmt.Print("Password: ")
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passBytes)

	if err := app.Session.LoginWithPassword(ctx, username, password); err != nil {
		app.Output.PrintError(err.Error())
		return err
	}

	app.Output.PrintSuccess(fmt.Sprintf("Logged in as %s", username))
	return nil
}

func loginDevice(ctx context.Context, app *App) error {
	deviceResp, err := app.Session.StartDeviceLogin(ctx)
	if err != nil {
		app.Output.PrintError(err.Error())
		return err
	}

	app.Output.Println("To sign in, open the following URL in a browser:")
	app.Output.Println("")
	if deviceResp.VerificationURIComplete != "" {
		app.Output.Println("  " + deviceResp.VerificationURIComplete)
	} else {
		app.Output.Println("  " + deviceResp.VerificationURI)
		app.Output.Println("")
		app.Output.Printf("  And enter code: %s\n", deviceResp.UserCode)
	}
	app.Output.Println("")
	app.Output.Println("Waiting for authentication...")

	if err := app.Session.PollDeviceToken(ctx, deviceResp.DeviceCode, deviceResp.Interval, deviceResp.ExpiresIn); err != nil {
		app.Output.PrintError(err.Error())
		return err
	}

	cred := app.Session.GetCredential()
	username := "user"
	if cred != nil {
		username = cred.Username
	}
	app.Output.PrintSuccess(fmt.Sprintf("Logged in as %s", username))
	return nil
}
