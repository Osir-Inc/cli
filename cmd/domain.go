package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/osir/cli/internal/api/models"
	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

func validateDomainArg(app *App, domain string) error {
	result := app.Client.ValidateDomainName(domain)
	if !result.Valid {
		app.Output.PrintError(result.Message)
		return fmt.Errorf("%s", result.Message)
	}
	return nil
}

func addDomainCommands(parent *cobra.Command) {
	domainCmd := &cobra.Command{
		Use:   "domain",
		Short: "Domain management commands",
		Long:  "Manage domains: check availability, register, renew, lock/unlock, configure privacy, nameservers, and more.",
	}

	domainCheckCmd := &cobra.Command{
		Use:   "check <domain>",
		Short: "Check domain availability",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain name to check (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if err := validateDomainArg(app, args[0]); err != nil {
				return err
			}

			result, err := app.Client.CheckDomainAvailability(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintKeyValue("Domain", result.Domain)
				if result.Available {
					app.Output.PrintSuccess("Available")
				} else {
					app.Output.PrintError("Not available")
				}
				if result.Status != "" {
					app.Output.PrintKeyValue("Status", result.Status)
				}
				if result.Price > 0 {
					app.Output.PrintKeyValue("Price", fmt.Sprintf("%.2f %s", result.Price, result.Currency))
				}
			}
			return nil
		},
	}

	domainRegisterCmd := &cobra.Command{
		Use:   "register <domain>",
		Short: "Register a new domain",
		Long:  "Register a new domain. Requires at least one nameserver via --nameservers.",
		Example: `  osir domain register osir.click --nameservers ns1.osir.com,ns2.osir.com
  osir domain register osir.click --nameservers ns1.osir.com,ns2.osir.com --years 2 --privacy --auto-renew`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain name to register (e.g. osir.click)",
					"--nameservers\tRequired: comma-separated nameservers (e.g. ns1.osir.com,ns2.osir.com)",
					"--years\tRegistration period in years (default: 1)",
					"--privacy\tEnable WHOIS privacy protection",
					"--auto-renew\tEnable auto-renewal",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if err := validateDomainArg(app, args[0]); err != nil {
				return err
			}

			years, _ := cmd.Flags().GetInt("years")
			nameservers, _ := cmd.Flags().GetStringSlice("nameservers")
			privacy, _ := cmd.Flags().GetBool("privacy")
			autoRenew, _ := cmd.Flags().GetBool("auto-renew")

			if years < 1 || years > 10 {
				app.Output.PrintError("Registration period must be between 1 and 10 years")
				return fmt.Errorf("invalid registration period: %d", years)
			}

			if len(nameservers) == 0 {
				app.Output.PrintError("At least one nameserver is required (use --nameservers ns1.example.com,ns2.example.com)")
				return fmt.Errorf("at least one nameserver is required")
			}

			req := &models.DomainRegistrationRequest{
				Domain:      args[0],
				Period:      years,
				Nameservers: nameservers,
				Privacy:     privacy,
				AutoRenew:   autoRenew,
			}

			result, err := app.Client.RegisterDomain(ctx, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if result.Success {
					app.Output.PrintSuccess(fmt.Sprintf("Domain %s registered", result.Domain))
				} else {
					app.Output.PrintError(result.Message)
				}
				if result.Status != "" {
					app.Output.PrintKeyValue("Status", result.Status)
				}
				if result.Price > 0 {
					app.Output.PrintKeyValue("Price", fmt.Sprintf("%.2f %s", result.Price, result.Currency))
				}
				if result.TransactionID != "" {
					app.Output.PrintKeyValue("Transaction", result.TransactionID)
				}
			}
			return nil
		},
	}

	domainInfoCmd := &cobra.Command{
		Use:   "info <domain>",
		Short: "Get detailed domain information",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain name to query (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if err := validateDomainArg(app, args[0]); err != nil {
				return err
			}

			result, err := app.Client.GetDomainInfo(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			if !result.Success || result.Data == nil {
				app.Output.PrintError("Failed to retrieve domain info")
				return fmt.Errorf("domain info unavailable")
			}

			d := result.Data
			app.Output.PrintKeyValue("Domain", d.Domain)
			app.Output.PrintKeyValue("Status", d.Status)
			if len(d.Statuses) > 0 {
				app.Output.PrintKeyValue("Statuses", strings.Join(d.Statuses, ", "))
			}
			app.Output.PrintKeyValue("Created", d.CreationDate)
			app.Output.PrintKeyValue("Expires", d.ExpiryDate)
			app.Output.PrintKeyValue("Expired", strconv.FormatBool(d.Expired))
			app.Output.PrintKeyValue("Locked", strconv.FormatBool(d.Locked))
			app.Output.PrintKeyValue("Auto-Renew", strconv.FormatBool(d.AutoRenew))
			app.Output.PrintKeyValue("Privacy", strconv.FormatBool(d.Privacy))
			app.Output.PrintKeyValue("Premium", strconv.FormatBool(d.Premium))
			if len(d.Nameservers) > 0 {
				app.Output.PrintKeyValue("Nameservers", strings.Join(d.Nameservers, ", "))
			}
			if d.RegistrantEmail != "" {
				app.Output.PrintKeyValue("Registrant Email", d.RegistrantEmail)
			}
			if d.Registrar != "" {
				app.Output.PrintKeyValue("Registrar", d.Registrar)
			}
			app.Output.PrintKeyValue("DNSSEC Enabled", strconv.FormatBool(d.DnssecEnabled))
			app.Output.PrintKeyValue("DNSSEC Supported", strconv.FormatBool(d.DnssecSupported))
			if d.InRedemptionPeriod {
				app.Output.PrintKeyValue("Redemption Period", "true")
				if d.RedemptionEndDate != nil {
					app.Output.PrintKeyValue("Redemption Ends", *d.RedemptionEndDate)
				}
			}
			if d.InAutoRenewGracePeriod {
				app.Output.PrintKeyValue("Auto-Renew Grace", "true")
				if d.AutoRenewGracePeriodEndDate != nil {
					app.Output.PrintKeyValue("Grace Period Ends", *d.AutoRenewGracePeriodEndDate)
				}
			}

			return nil
		},
	}

	domainListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all domains in your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")
			sortBy, _ := cmd.Flags().GetString("sort-by")
			sortDir, _ := cmd.Flags().GetString("sort-dir")

			result, err := app.Client.ListDomains(ctx, page, size, sortBy, sortDir)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			printDomainList(app.Output, result)
			return nil
		},
	}

	domainRenewCmd := &cobra.Command{
		Use:   "renew <domain>",
		Short: "Renew a domain registration",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain to renew (e.g. osir.click)",
						"--years\tNumber of years to renew (default: 1)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if err := validateDomainArg(app, args[0]); err != nil {
				return err
			}

			years, _ := cmd.Flags().GetInt("years")
			if years < 1 || years > 10 {
				app.Output.PrintError("Renewal period must be between 1 and 10 years")
				return fmt.Errorf("invalid renewal period: %d", years)
			}

			req := &models.DomainRenewalRequest{
				Period: years,
			}

			result, err := app.Client.RenewDomain(ctx, args[0], req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if result.Success {
					app.Output.PrintSuccess(fmt.Sprintf("Domain %s renewed", result.Domain))
				} else {
					app.Output.PrintError(result.Message)
				}
				if result.ExpirationDate != "" {
					app.Output.PrintKeyValue("New Expiry", result.ExpirationDate)
				}
				if result.Price > 0 {
					app.Output.PrintKeyValue("Price", fmt.Sprintf("%.2f", result.Price))
				}
			}
			return nil
		},
	}

	domainLockCmd := &cobra.Command{
		Use:   "lock <domain>",
		Short: "Lock a domain to prevent unauthorized transfers",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain to lock (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.LockDomain(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("Domain locked: %s", args[0]))
			}
			return nil
		},
	}

	domainUnlockCmd := &cobra.Command{
		Use:   "unlock <domain>",
		Short: "Unlock a domain to allow transfers",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain to unlock (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.UnlockDomain(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("Domain unlocked: %s", args[0]))
			}
			return nil
		},
	}

	domainAutoRenewCmd := &cobra.Command{
		Use:   "auto-renew <domain> <true|false>",
		Short: "Enable or disable auto-renewal for a domain",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain name (e.g. osir.click)",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return []string{"true\tEnable auto-renew", "false\tDisable auto-renew"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Invalid value '%s': must be 'true' or 'false'", args[1]))
				return fmt.Errorf("invalid value '%s': must be 'true' or 'false'", args[1])
			}

			result, err := app.Client.SetAutoRenew(ctx, args[0], enabled)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				state := "disabled"
				if enabled {
					state = "enabled"
				}
				app.Output.PrintSuccess(fmt.Sprintf("Auto-renew %s for %s", state, args[0]))
			}
			return nil
		},
	}

	domainPrivacyCmd := &cobra.Command{
		Use:   "privacy <domain> <true|false>",
		Short: "Enable or disable WHOIS privacy protection",
		Args:  cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain name (e.g. osir.click)",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			if len(args) == 1 {
				return []string{"true\tEnable WHOIS privacy", "false\tDisable WHOIS privacy"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Invalid value '%s': must be 'true' or 'false'", args[1]))
				return fmt.Errorf("invalid value '%s': must be 'true' or 'false'", args[1])
			}

			result, err := app.Client.SetPrivacy(ctx, args[0], enabled)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				state := "disabled"
				if enabled {
					state = "enabled"
				}
				app.Output.PrintSuccess(fmt.Sprintf("WHOIS privacy %s for %s", state, args[0]))
			}
			return nil
		},
	}

	domainValidateCmd := &cobra.Command{
		Use:   "validate <domain>",
		Short: "Validate a domain name format",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain name to validate (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)

			result := app.Client.ValidateDomainName(args[0])
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if result.Valid {
					app.Output.PrintSuccess(result.Message)
				} else {
					app.Output.PrintError(result.Message)
				}
			}
			return nil
		},
	}

	domainSuggestCmd := &cobra.Command{
		Use:   "suggest <keyword>",
		Short: "Suggest alternative domain names",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<keyword>\tKeyword to find domain suggestions for",
						"--limit\tMax number of suggestions (default: 10)",
						"--tlds\tComma-separated TLDs (e.g. com,net,org)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			limit, _ := cmd.Flags().GetInt("limit")
			tlds, _ := cmd.Flags().GetString("tlds")

			result, err := app.Client.SuggestAlternatives(ctx, args[0], limit, tlds)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if len(result.Suggestions) == 0 {
					app.Output.Println("No suggestions found")
				} else {
					headers := []string{"DOMAIN"}
					rows := make([][]string, len(result.Suggestions))
					for i, s := range result.Suggestions {
						rows[i] = []string{s}
					}
					app.Output.PrintTable(headers, rows)
				}
			}
			return nil
		},
	}

	domainNameserversCmd := &cobra.Command{
		Use:   "nameservers <domain> <ns1> [ns2...]",
		Short: "Update nameservers for a domain",
		Args:  cobra.MinimumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
						"<domain>\tDomain to update nameservers for (e.g. osir.click)",
						"--help\tShow full help for this command",
					}, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{
					"<nameserver>\tNameserver hostname (e.g. ns1.osir.com)",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]
			nameservers := args[1:]

			req := &models.NameserverUpdateRequest{
				Nameservers: nameservers,
			}

			result, err := app.Client.UpdateNameservers(ctx, domain, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("Nameservers updated for %s", domain))
			}
			return nil
		},
	}

	// register command flags
	domainRegisterCmd.Flags().Int("years", 1, "Number of years to register")
	domainRegisterCmd.Flags().StringSlice("nameservers", nil, "Nameservers (comma-separated, at least 1 required)")
	domainRegisterCmd.Flags().Bool("privacy", false, "Enable WHOIS privacy protection")
	domainRegisterCmd.Flags().Bool("auto-renew", false, "Enable auto-renewal")
	_ = domainRegisterCmd.MarkFlagRequired("nameservers")

	// list command flags
	domainListCmd.Flags().Int("page", 0, "Page number")
	domainListCmd.Flags().Int("size", 50, "Page size")
	domainListCmd.Flags().String("sort-by", "", "Sort by field (e.g. expirationDate, domain)")
	domainListCmd.Flags().String("sort-dir", "", "Sort direction (asc or desc)")

	// renew command flags
	domainRenewCmd.Flags().Int("years", 1, "Number of years to renew")

	// suggest command flags
	domainSuggestCmd.Flags().Int("limit", 10, "Maximum number of suggestions")
	domainSuggestCmd.Flags().String("tlds", "", "Comma-separated TLDs to search (e.g. com,net,org)")

	domainCmd.AddCommand(
		domainCheckCmd,
		domainRegisterCmd,
		domainInfoCmd,
		domainListCmd,
		domainRenewCmd,
		domainLockCmd,
		domainUnlockCmd,
		domainAutoRenewCmd,
		domainPrivacyCmd,
		domainValidateCmd,
		domainSuggestCmd,
		domainNameserversCmd,
	)

	parent.AddCommand(domainCmd)
}

func printDomainList(out *output.Formatter, result *models.DomainListResponse) {
	d := result.Data
	if len(d.Domains) == 0 {
		out.Println("No domains found")
		return
	}

	out.Println(fmt.Sprintf("Total: %d (page %d of %d)", d.TotalElements, d.Page+1, d.TotalPages))
	out.Println("")

	headers := []string{"DOMAIN", "STATUS", "EXPIRES", "AUTO-RENEW", "PRIVACY", "NAMESERVERS"}
	rows := make([][]string, len(d.Domains))
	for i, dom := range d.Domains {
		ns := ""
		if len(dom.Nameservers) > 0 {
			ns = strings.Join(dom.Nameservers, ", ")
		}
		rows[i] = []string{
			dom.Domain,
			dom.Status,
			formatDate(dom.ExpirationDate),
			strconv.FormatBool(dom.AutoRenew),
			strconv.FormatBool(dom.Privacy),
			ns,
		}
	}
	out.PrintTable(headers, rows)
}
