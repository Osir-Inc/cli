package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/osir/cli/internal/api/models"
	"github.com/spf13/cobra"
)

// resolveInstanceID resolves a short or full instance ID by prefix-matching
// against the user's instance list. Returns the full UUID.
func resolveInstanceID(ctx context.Context, app *App, idPrefix string) (string, error) {
	// If it looks like a full UUID, use it directly
	if len(idPrefix) == 36 && strings.Count(idPrefix, "-") == 4 {
		return idPrefix, nil
	}

	instances, err := app.Client.ListVpsInstances(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to list instances: %w", err)
	}

	var matches []models.VpsInstance
	for _, inst := range instances {
		if strings.HasPrefix(inst.ID, idPrefix) || inst.Hostname == idPrefix {
			matches = append(matches, inst)
		}
	}

	switch len(matches) {
	case 0:
		app.Output.PrintError(fmt.Sprintf("No instance matching '%s'", idPrefix))
		app.Output.Println("Run 'vps list' to see your instances.")
		return "", fmt.Errorf("instance '%s' not found", idPrefix)
	case 1:
		return matches[0].ID, nil
	default:
		app.Output.PrintError(fmt.Sprintf("'%s' matches %d instances — be more specific:", idPrefix, len(matches)))
		for _, m := range matches {
			app.Output.Println(fmt.Sprintf("  %s  %s  %s", m.ID, m.Hostname, m.IPAddress))
		}
		return "", fmt.Errorf("ambiguous instance ID '%s'", idPrefix)
	}
}

// resolvePackage looks up a package by name or ID from the catalog.
func resolvePackage(ctx context.Context, app *App, nameOrID string) (*models.VpsPackageDetail, error) {
	catalog, err := app.Client.ListVpsCatalog(ctx, true, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch packages: %w", err)
	}

	// Try exact name match (case-insensitive)
	for i, pkg := range catalog.Packages {
		if strings.EqualFold(pkg.Name, nameOrID) {
			return &catalog.Packages[i], nil
		}
	}

	// Try ID match
	for i, pkg := range catalog.Packages {
		if pkg.ID == nameOrID {
			return &catalog.Packages[i], nil
		}
	}

	// Not found — show available packages
	app.Output.PrintError(fmt.Sprintf("Package '%s' not found", nameOrID))
	app.Output.Println("")
	app.Output.Println("Available packages:")
	for _, pkg := range catalog.Packages {
		location := ""
		if pkg.Location != nil {
			location = " (" + pkg.Location.DisplayName + ")"
		}
		app.Output.Println(fmt.Sprintf("  %-10s %d cores, %s RAM, %s/mo%s",
			pkg.Name, pkg.CpuCores, formatMemory(pkg.MemoryMb), formatCents(pkg.PriceMonthly), location))
	}
	return nil, fmt.Errorf("package '%s' not found", nameOrID)
}

// resolveLocation looks up a location by name or ID from the catalog.
func resolveLocation(ctx context.Context, app *App, nameOrID string) (*models.VpsLocationDetail, error) {
	locations, err := app.Client.ListVpsCatalogLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}

	for i, loc := range locations.Locations {
		if strings.EqualFold(loc.DisplayName, nameOrID) ||
			strings.EqualFold(loc.City, nameOrID) ||
			loc.ID == nameOrID {
			return &locations.Locations[i], nil
		}
	}

	app.Output.PrintError(fmt.Sprintf("Location '%s' not found", nameOrID))
	app.Output.Println("")
	app.Output.Println("Available locations:")
	for _, loc := range locations.Locations {
		app.Output.Println(fmt.Sprintf("  %-25s %s", loc.DisplayName, loc.ID))
	}
	return nil, fmt.Errorf("location '%s' not found", nameOrID)
}

func addVpsCommands(parent *cobra.Command) {
	vpsCmd := &cobra.Command{
		Use:   "vps",
		Short: "VPS hosting management",
		Long: `Manage VPS instances: browse packages and locations, order new servers,
list and inspect your instances, change payment terms, and access the control panel.

Quick start:
  osir vps packages                              Browse available plans
  osir vps order --package ZANA-S --hostname web1 Order a server
  osir vps list                                   See your instances`,
	}

	vpsPackagesCmd := &cobra.Command{
		Use:   "packages",
		Short: "List available VPS packages",
		Long: `List all VPS packages available for ordering, showing CPU cores, RAM,
disk, traffic, storage type, monthly price, and datacenter location.

Use the package NAME when ordering: osir vps order --package ZANA-S --hostname myserver`,
		Example: `  osir vps packages
  osir vps packages -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.ListVpsCatalog(ctx, true, "")
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			if len(result.Packages) == 0 {
				app.Output.Println("No VPS packages available")
				return nil
			}

			headers := []string{"NAME", "CPU", "RAM", "DISK", "TRAFFIC", "STORAGE", "MONTHLY", "ANNUAL", "LOCATION"}
			rows := make([][]string, len(result.Packages))
			for i, pkg := range result.Packages {
				location := ""
				if pkg.Location != nil {
					location = pkg.Location.DisplayName
				}
				annual := ""
				if pkg.PriceAnnual > 0 {
					annual = formatCents(pkg.PriceAnnual)
				}
				rows[i] = []string{
					pkg.Name,
					strconv.Itoa(pkg.CpuCores) + " cores",
					formatMemory(pkg.MemoryMb),
					strconv.Itoa(pkg.StorageGb) + " GB",
					strconv.Itoa(pkg.TrafficGb) + " GB",
					pkg.StorageProfile,
					formatCents(pkg.PriceMonthly),
					annual,
					location,
				}
			}
			app.Output.PrintTable(headers, rows)
			app.Output.Println("")
			app.Output.Println("Order with: osir vps order --package <NAME> --hostname <hostname>")
			return nil
		},
	}

	vpsLocationsCmd := &cobra.Command{
		Use:   "locations",
		Short: "List available VPS datacenter locations",
		Long: `List all datacenter locations where VPS instances can be deployed.

Use the location name or city when ordering:
  osir vps order --package ZANA-S --hostname web1 --location Nueremberg`,
		Example: `  osir vps locations
  osir vps locations -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.ListVpsCatalogLocations(ctx)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			if len(result.Locations) == 0 {
				app.Output.Println("No VPS locations available")
				return nil
			}

			headers := []string{"LOCATION", "COUNTRY", "CODE"}
			rows := make([][]string, len(result.Locations))
			for i, loc := range result.Locations {
				rows[i] = []string{
					loc.DisplayName,
					loc.CountryName,
					loc.CountryCode,
				}
			}
			app.Output.PrintTable(headers, rows)
			return nil
		},
	}

	vpsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your VPS instances",
		Long: `List all VPS instances in your account. Optionally filter by status.

Statuses: ACTIVE, SUSPENDED, TERMINATED, PENDING`,
		Example: `  osir vps list
  osir vps list --status ACTIVE
  osir vps list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			status, _ := cmd.Flags().GetString("status")

			instances, err := app.Client.ListVpsInstances(ctx, status)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(instances)
				return nil
			}

			printVpsInstanceTable(app, instances)
			return nil
		},
	}

	vpsActiveCmd := &cobra.Command{
		Use:     "active",
		Short:   "List active VPS instances only",
		Long:    "List only VPS instances with ACTIVE status. Shorthand for 'vps list --status ACTIVE'.",
		Example: `  osir vps active`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			instances, err := app.Client.ListActiveVpsInstances(ctx)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(instances)
				return nil
			}

			printVpsInstanceTable(app, instances)
			return nil
		},
	}

	vpsCountCmd := &cobra.Command{
		Use:   "count",
		Short: "Count your VPS instances",
		Long:  "Get the total count of VPS instances in your account. Use --active-only to count only active ones.",
		Example: `  osir vps count
  osir vps count --active-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			activeOnly, _ := cmd.Flags().GetBool("active-only")

			result, err := app.Client.CountVpsInstances(ctx, activeOnly)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			app.Output.PrintResult(result)
			return nil
		},
	}

	vpsInfoCmd := &cobra.Command{
		Use:   "info <instanceId>",
		Short: "Get VPS instance details",
		Long: `Get detailed information about a specific VPS instance.

Find instance IDs with 'vps list'.`,
		Example: `  osir vps info 7088d366-1f10-48ec-9e24-30a349dfdc5b`,
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			instanceID, err := resolveInstanceID(ctx, app, args[0])
			if err != nil {
				return err
			}

			inst, err := app.Client.GetVpsInstance(ctx, instanceID)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(inst)
				return nil
			}

			printVpsInstanceDetail(app, inst)

			// Generate a fresh control panel login URL
			login, err := app.Client.GetVpsPanelLogin(ctx, instanceID)
			if err == nil && login.LoginURL != "" {
				app.Output.Println("")
				app.Output.PrintKeyValue("Control Panel", login.LoginURL)
			}

			return nil
		},
	}

	vpsOrderCmd := &cobra.Command{
		Use:   "order --package <name> --hostname <hostname>",
		Short: "Order a new VPS instance",
		Long: `Order a new VPS instance by package name and hostname.

The --package flag accepts the package NAME (e.g. ZANA-S, ZANA-M, ZANA-L) or UUID.
The --location flag accepts the location name (e.g. Nueremberg), city, or UUID.

Payment terms: MONTHLY (default), QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL

Workflow:
  1. Browse plans:     osir vps packages
  2. Browse locations: osir vps locations
  3. Order:            osir vps order --package ZANA-S --hostname myserver`,
		Example: `  osir vps order --package ZANA-S --hostname myserver
  osir vps order --package ZANA-M --hostname web01 --payment-term ANNUAL
  osir vps order --package ZANA-L --hostname db01 --location Nueremberg --root-password "S3cur3!"`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{
				"--package\tPackage name, e.g. ZANA-S (required)",
				"--hostname\tHostname for the VPS (required)",
				"--payment-term\tMONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL",
				"--location\tLocation name, e.g. Nueremberg",
				"--root-password\tRoot password for the VPS",
			}, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			packageRef, _ := cmd.Flags().GetString("package")
			hostname, _ := cmd.Flags().GetString("hostname")
			paymentTerm, _ := cmd.Flags().GetString("payment-term")
			locationRef, _ := cmd.Flags().GetString("location")
			rootPassword, _ := cmd.Flags().GetString("root-password")

			// Resolve package name → ID
			pkg, err := resolvePackage(ctx, app, packageRef)
			if err != nil {
				return err
			}

			// Resolve location name → ID (if provided)
			locationID := ""
			if locationRef != "" {
				loc, err := resolveLocation(ctx, app, locationRef)
				if err != nil {
					return err
				}
				locationID = loc.ID
			} else if pkg.Location != nil {
				// Default to the package's location
				locationID = pkg.Location.ID
			}

			req := models.VpsOrderRequest{
				PackageID:    pkg.ID,
				PaymentTerm:  strings.ToUpper(paymentTerm),
				Hostname:     hostname,
				LocationID:   locationID,
				RootPassword: rootPassword,
			}

			app.Output.Println(fmt.Sprintf("Ordering %s (%d cores, %s RAM, %s/mo) as '%s'...",
				pkg.Name, pkg.CpuCores, formatMemory(pkg.MemoryMb), formatCents(pkg.PriceMonthly), hostname))

			result, err := app.Client.OrderVps(ctx, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("VPS '%s' ordered with package %s (%s)", hostname, pkg.Name, strings.ToUpper(paymentTerm)))
				app.Output.Println("")
				app.Output.PrintKeyValue("Order", fmt.Sprintf("%s (ID: %d)", result.OrderNumber, result.OrderID))
				app.Output.PrintKeyValue("Status", result.OrderStatus)
				app.Output.PrintKeyValue("Invoice", fmt.Sprintf("%s (ID: %d)", result.InvoiceNumber, result.InvoiceID))
				if result.TotalAmount > 0 {
					app.Output.PrintKeyValue("Amount", fmt.Sprintf("%s %s", formatCents(result.TotalAmount), result.Currency))
				}
				if result.DueDate != "" {
					app.Output.PrintKeyValue("Due Date", result.DueDate)
				}
				if result.Instance != nil {
					inst := result.Instance
					app.Output.Println("")
					app.Output.PrintKeyValue("Instance ID", inst.ID)
					app.Output.PrintKeyValue("Hostname", inst.Hostname)
					app.Output.PrintKeyValue("Package", inst.PackageName)
					app.Output.PrintKeyValue("Status", inst.Status)
					app.Output.PrintKeyValue("Provisioning", inst.ProvisioningStatus)
					if inst.IPAddress != "" {
						app.Output.PrintKeyValue("IPv4", inst.IPAddress)
					}
					if inst.IPv6Addresses != "" {
						app.Output.PrintKeyValue("IPv6", inst.IPv6Addresses)
					}
					if inst.Message != "" {
						app.Output.PrintKeyValue("Message", inst.Message)
					}
					if inst.ControlPanelUrl != "" {
						app.Output.Println("")
						app.Output.PrintKeyValue("Control Panel", inst.ControlPanelUrl)
						app.Output.Println("(this link expires — regenerate with: osir vps login " + inst.ID + ")")
					}
				}
			}
			return nil
		},
	}

	vpsDeleteCmd := &cobra.Command{
		Use:   "delete <instanceId>",
		Short: "Delete a VPS instance",
		Long: `Delete (cancel) one of your VPS instances. This action cannot be undone.

Find instance IDs with 'vps list'.`,
		Example: `  osir vps delete 7088d366-1f10-48ec-9e24-30a349dfdc5b`,
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			instanceID, err := resolveInstanceID(ctx, app, args[0])
			if err != nil {
				return err
			}

			result, err := app.Client.DeleteVpsInstance(ctx, instanceID)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("VPS instance %s deleted", instanceID))
			}
			return nil
		},
	}

	vpsChangeTermCmd := &cobra.Command{
		Use:   "change-term <instanceId> --term <MONTHLY|QUARTERLY|SEMI_ANNUAL|ANNUAL|BIENNIAL|TRIENNIAL>",
		Short: "Change payment term for a VPS instance",
		Long: `Change the billing payment term for a VPS instance.

Payment terms: MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL
Longer terms typically offer discounts — run 'vps packages' to see pricing.`,
		Example: `  osir vps change-term 7088d366-1f10-48ec-9e24-30a349dfdc5b --term ANNUAL
  osir vps change-term 7088d366-... --term SEMI_ANNUAL`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			instanceID, err := resolveInstanceID(ctx, app, args[0])
			if err != nil {
				return err
			}

			term, _ := cmd.Flags().GetString("term")

			req := models.VpsPaymentTermChangeRequest{
				NewPaymentTerm: strings.ToUpper(term),
			}

			result, err := app.Client.ChangeVpsPaymentTerm(ctx, instanceID, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("Payment term changed to %s for %s", strings.ToUpper(term), instanceID))
			}
			return nil
		},
	}

	vpsLoginCmd := &cobra.Command{
		Use:   "login <instanceId>",
		Short: "Get VPS control panel login URL",
		Long: `Generate a one-time login URL for the VPS control panel (VirtFusion).
Gives you access to the server console, power controls, OS reinstall, and more.`,
		Example: `  osir vps login 7088d366-1f10-48ec-9e24-30a349dfdc5b`,
		Args:    cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			instanceID, err := resolveInstanceID(ctx, app, args[0])
			if err != nil {
				return err
			}

			result, err := app.Client.GetVpsPanelLogin(ctx, instanceID)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else if result.LoginURL != "" {
				app.Output.PrintSuccess("Control panel login URL generated")
				app.Output.PrintKeyValue("URL", result.LoginURL)
			} else {
				app.Output.PrintResult(result)
			}
			return nil
		},
	}

	// Flags
	vpsListCmd.Flags().String("status", "", "Filter: ACTIVE, SUSPENDED, TERMINATED, PENDING")
	vpsCountCmd.Flags().Bool("active-only", false, "Count only active instances")

	vpsOrderCmd.Flags().String("package", "", "Package name (e.g. ZANA-S) or ID (required)")
	vpsOrderCmd.Flags().String("hostname", "", "Hostname for the VPS (required)")
	vpsOrderCmd.Flags().String("payment-term", "MONTHLY", "MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL")
	vpsOrderCmd.Flags().String("location", "", "Location name (e.g. Nueremberg) or ID")
	vpsOrderCmd.Flags().String("root-password", "", "Root password for the VPS")
	_ = vpsOrderCmd.MarkFlagRequired("package")
	_ = vpsOrderCmd.MarkFlagRequired("hostname")

	vpsChangeTermCmd.Flags().String("term", "", "MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL, BIENNIAL, TRIENNIAL (required)")
	_ = vpsChangeTermCmd.MarkFlagRequired("term")

	vpsCmd.AddCommand(
		vpsPackagesCmd,
		vpsLocationsCmd,
		vpsListCmd,
		vpsActiveCmd,
		vpsCountCmd,
		vpsInfoCmd,
		vpsOrderCmd,
		vpsDeleteCmd,
		vpsChangeTermCmd,
		vpsLoginCmd,
	)

	parent.AddCommand(vpsCmd)
}

func printVpsInstanceTable(app *App, instances []models.VpsInstance) {
	if len(instances) == 0 {
		app.Output.Println("No VPS instances found")
		return
	}

	headers := []string{"ID", "HOSTNAME", "STATUS", "PACKAGE", "IPv4", "TERM", "RENEWAL", "LOCATION"}
	rows := make([][]string, len(instances))
	for i, inst := range instances {
		pkg := ""
		if inst.VpsPackage != nil {
			pkg = inst.VpsPackage.Name
		}
		location := ""
		if inst.HypervisorGroup != nil {
			location = inst.HypervisorGroup.DisplayName
		}
		rows[i] = []string{
			inst.ID,
			inst.Hostname,
			inst.Status,
			pkg,
			inst.IPAddress,
			inst.PaymentTerm,
			inst.NextRenewalDate,
			location,
		}
	}
	app.Output.PrintTable(headers, rows)
}

func printVpsInstanceDetail(app *App, inst *models.VpsInstance) {
	app.Output.PrintKeyValue("Instance ID", inst.ID)
	app.Output.PrintKeyValue("Hostname", inst.Hostname)
	app.Output.PrintKeyValue("Status", inst.Status)
	app.Output.PrintKeyValue("Provisioning", inst.ProvisioningStatus)
	if inst.VpsPackage != nil {
		p := inst.VpsPackage
		app.Output.PrintKeyValue("Package", fmt.Sprintf("%s (%d cores, %s RAM, %d GB disk)",
			p.Name, p.CpuCores, formatMemory(p.MemoryMb), p.StorageGb))
	}
	app.Output.PrintKeyValue("IPv4", inst.IPAddress)
	if inst.IPv6Addresses != "" {
		app.Output.PrintKeyValue("IPv6", inst.IPv6Addresses)
	}
	app.Output.PrintKeyValue("Payment Term", inst.PaymentTerm)
	if inst.NextRenewalDate != "" {
		app.Output.PrintKeyValue("Next Renewal", inst.NextRenewalDate)
	}
	if inst.HypervisorGroup != nil {
		app.Output.PrintKeyValue("Location", inst.HypervisorGroup.DisplayName)
	}
	if inst.CreatedAt != "" {
		app.Output.PrintKeyValue("Created", inst.CreatedAt)
	}
}

// formatCents converts a price in cents to a dollar string (e.g. 299 -> "$2.99").
func formatCents(cents int) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100.0)
}

// formatMemory formats megabytes into a human-readable string.
func formatMemory(mb int) string {
	if mb >= 1024 && mb%1024 == 0 {
		return strconv.Itoa(mb/1024) + " GB"
	}
	return strconv.Itoa(mb) + " MB"
}
