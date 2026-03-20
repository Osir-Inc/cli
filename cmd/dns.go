package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/osir/cli/internal/api"
	"github.com/osir/cli/internal/api/models"
	"github.com/spf13/cobra"
)

// Known DNS record types for smart selector detection.
var knownRecordTypes = map[string]bool{
	"A": true, "AAAA": true, "CNAME": true, "MX": true, "TXT": true,
	"NS": true, "SRV": true, "CAA": true, "PTR": true, "SOA": true,
}

// isRecordType returns true if s is a known DNS record type (case-insensitive).
func isRecordType(s string) bool {
	return knownRecordTypes[strings.ToUpper(s)]
}

// normalizeRecordName strips a trailing dot for comparison purposes.
func normalizeRecordName(name string) string {
	return strings.TrimSuffix(strings.ToLower(name), ".")
}

// resolveDnsRecord finds a single DNS record by type and optional filters.
// Use nameFilter/contentFilter for exact field matching, or nameOrContent
// to match against name first then fall back to content — all in one API call.
func resolveDnsRecord(client api.Backend, cmd *cobra.Command, domain, recordType string, nameFilter, contentFilter, nameOrContent string) (*models.DnsRecord, error) {
	ctx := cmd.Context()
	app := getApp(cmd)

	records, err := client.ListDnsRecords(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	recordType = strings.ToUpper(recordType)

	// Filter by type
	var typed []models.DnsRecord
	for _, r := range records {
		if strings.ToUpper(r.Type) == recordType {
			typed = append(typed, r)
		}
	}

	if len(typed) == 0 {
		app.Output.PrintError(fmt.Sprintf("No %s record found for %s", recordType, domain))
		app.Output.Println("Hint: Run 'dns list " + domain + "' to see all records.")
		return nil, fmt.Errorf("no %s record found for %s", recordType, domain)
	}

	candidates := typed

	// nameOrContent mode: try name match first, fall back to content — single pass
	if nameOrContent != "" {
		norm := normalizeRecordName(nameOrContent)
		var byName, byContent []models.DnsRecord
		for _, r := range candidates {
			if normalizeRecordName(r.Name) == norm {
				byName = append(byName, r)
			}
			if normalizeRecordName(r.Content) == norm {
				byContent = append(byContent, r)
			}
		}
		if len(byName) > 0 {
			candidates = byName
		} else if len(byContent) > 0 {
			candidates = byContent
		} else {
			app.Output.PrintError(fmt.Sprintf("No %s record matching '%s' found for %s", recordType, nameOrContent, domain))
			app.Output.Println("Hint: Run 'dns list " + domain + "' to see all records.")
			return nil, fmt.Errorf("no %s record matching '%s' found for %s", recordType, nameOrContent, domain)
		}
	}

	// Apply exact name filter
	if nameFilter != "" {
		var filtered []models.DnsRecord
		for _, r := range candidates {
			if normalizeRecordName(r.Name) == normalizeRecordName(nameFilter) {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			app.Output.PrintError(fmt.Sprintf("No %s record with name '%s' found for %s", recordType, nameFilter, domain))
			app.Output.Println("Hint: Run 'dns list " + domain + "' to see all records.")
			return nil, fmt.Errorf("no %s record with name '%s' found for %s", recordType, nameFilter, domain)
		}
		candidates = filtered
	}

	// Apply exact content filter
	if contentFilter != "" {
		var filtered []models.DnsRecord
		for _, r := range candidates {
			if normalizeRecordName(r.Content) == normalizeRecordName(contentFilter) {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) == 0 {
			app.Output.PrintError(fmt.Sprintf("No %s record with content '%s' found for %s", recordType, contentFilter, domain))
			app.Output.Println("Hint: Run 'dns list " + domain + "' to see all records.")
			return nil, fmt.Errorf("no %s record with content '%s' found for %s", recordType, contentFilter, domain)
		}
		candidates = filtered
	}

	if len(candidates) == 1 {
		return &candidates[0], nil
	}

	// Ambiguous — show candidates
	app.Output.PrintError(fmt.Sprintf("Found %d %s records for %s. Be more specific:", len(candidates), recordType, domain))
	app.Output.Println("")
	formatMatchingRecords(app, candidates)
	app.Output.Println("")
	if len(candidates) > 0 {
		r := candidates[0]
		app.Output.Println(fmt.Sprintf("Tip: dns <command> %s %s %s %s", domain, recordType, r.Name, r.Content))
	}
	return nil, fmt.Errorf("ambiguous: found %d %s records for %s", len(candidates), recordType, domain)
}

// formatMatchingRecords prints a table of candidate records.
func formatMatchingRecords(app *App, records []models.DnsRecord) {
	headers := []string{"NAME", "CONTENT", "TTL"}
	rows := make([][]string, len(records))
	for i, r := range records {
		rows[i] = []string{r.Name, r.Content, strconv.Itoa(r.TTL)}
	}
	app.Output.PrintTable(headers, rows)
}

// printRegistryResult handles the registryPublished/registryPublishError pattern
// shared by DNSSEC enable and disable responses.
func printRegistryResult(app *App, domain string, action string, registryPublished *bool, registryPublishErr *string) {
	switch {
	case registryPublished != nil && *registryPublished:
		if action == "enabled" {
			app.Output.PrintSuccess(fmt.Sprintf("DNSSEC enabled and DS records published to registry for %s", domain))
		} else {
			app.Output.PrintSuccess(fmt.Sprintf("DNSSEC disabled and DS records removed from registry for %s", domain))
		}
	case registryPublished != nil && !*registryPublished:
		verb := "publish to"
		if action == "disabled" {
			verb = "remove from"
		}
		app.Output.PrintError(fmt.Sprintf("DNSSEC %s for %s, but DS records failed to %s registry", action, domain, verb))
		if registryPublishErr != nil {
			app.Output.PrintKeyValue("Error", *registryPublishErr)
		}
		app.Output.Println(fmt.Sprintf("Fix manually via: PUT /v2/domains/%s/dnssec", domain))
	default:
		app.Output.PrintSuccess(fmt.Sprintf("DNSSEC %s for %s", action, domain))
	}
}

// ensureTrailingDot appends a trailing dot to content for record types
// that require a FQDN (CNAME, MX, NS, SRV, PTR).
func ensureTrailingDot(recordType, content string) string {
	switch strings.ToUpper(recordType) {
	case "CNAME", "MX", "NS", "SRV", "PTR":
		if content != "" && !strings.HasSuffix(content, ".") {
			return content + "."
		}
	}
	return content
}

func addDnsCommands(parent *cobra.Command) {
	dnsCmd := &cobra.Command{
		Use:   "dns",
		Short: "DNS record management",
		Long:  "Manage DNS records for your domains, including listing, creating, updating, and deleting records.",
	}

	dnsListCmd := &cobra.Command{
		Use:   "list <domain>",
		Short: "List all DNS records for a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]

			// Check if zone exists first
			zoneCheck, err := app.Client.CheckZoneExists(ctx, domain)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to check zone: %s", err))
				return err
			}
			if !zoneCheck.Exists {
				app.Output.PrintError(fmt.Sprintf("No DNS zone found for %s. Create it first with: dns zone-init %s", domain, domain))
				return fmt.Errorf("zone does not exist for %s", domain)
			}

			records, err := app.Client.ListDnsRecords(ctx, domain)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to list DNS records: %s", err))
				return err
			}

			if len(records) == 0 {
				app.Output.Println(fmt.Sprintf("No DNS records found for %s", domain))
				return nil
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(records)
			} else {
				headers := []string{"ID", "TYPE", "NAME", "CONTENT", "TTL", "PRIORITY"}
				rows := make([][]string, len(records))
				for i, r := range records {
					priority := ""
					if r.Priority > 0 {
						priority = strconv.Itoa(r.Priority)
					}
					rows[i] = []string{r.ID, r.Type, r.Name, r.Content, strconv.Itoa(r.TTL), priority}
				}
				app.Output.PrintTable(headers, rows)
			}

			return nil
		},
	}

	dnsGetCmd := &cobra.Command{
		Use:   "get <domain> <recordId|TYPE> [nameOrContent]",
		Short: "Get a specific DNS record",
		Long: `Get a DNS record by ID or by type selector.

Smart selectors (when second arg is a DNS type like A, CNAME, MX):
  dns get <domain> <TYPE>                  Find unique record of that type
  dns get <domain> <TYPE> <nameOrContent>  Disambiguate by name or content

Legacy:
  dns get <domain> <recordId>              Direct lookup by record ID`,
		Example: `  osir dns get wxuh.com wxuh_com__A_989519897
  osir dns get wxuh.com CNAME
  osir dns get wxuh.com A mail.wxuh.com`,
		Args: cobra.RangeArgs(2, 3),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return []string{"<domain>\tDomain name (e.g. wxuh.com)"}, cobra.ShellCompDirectiveNoFileComp
			case 1:
				return []string{
					"A\tIPv4 address record",
					"AAAA\tIPv6 address record",
					"CNAME\tCanonical name (alias)",
					"MX\tMail exchange",
					"TXT\tText record",
					"NS\tNameserver",
					"<recordId>\tRecord ID from 'dns list'",
				}, cobra.ShellCompDirectiveNoFileComp
			case 2:
				return []string{"<nameOrContent>\tFilter by name or content"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]
			selector := args[1]

			var record *models.DnsRecord

			if isRecordType(selector) {
				// Smart mode: resolve by type + optional filter
				nameOrContent := ""
				if len(args) == 3 {
					nameOrContent = args[2]
				}

				resolved, err := resolveDnsRecord(app.Client, cmd, domain, selector, "", "", nameOrContent)
				if err != nil {
					return err
				}
				record = resolved
			} else {
				// Legacy: direct ID lookup
				var err error
				record, err = app.Client.GetDnsRecord(ctx, domain, selector)
				if err != nil {
					app.Output.PrintError(fmt.Sprintf("Failed to get DNS record: %s", err))
					return err
				}
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(record)
			} else {
				app.Output.PrintKeyValue("ID", record.ID)
				app.Output.PrintKeyValue("Type", record.Type)
				app.Output.PrintKeyValue("Name", record.Name)
				app.Output.PrintKeyValue("Content", record.Content)
				app.Output.PrintKeyValue("TTL", strconv.Itoa(record.TTL))
				if record.Priority > 0 {
					app.Output.PrintKeyValue("Priority", strconv.Itoa(record.Priority))
				}
			}

			return nil
		},
	}

	dnsCreateCmd := &cobra.Command{
		Use:   "create <domain> <type> <name> <content>",
		Short: "Create a new DNS record",
		Long:  "Create a new DNS record. Trailing dot is auto-added for CNAME, MX, NS, SRV, PTR.",
		Example: `  osir dns create wxuh.com A mail.wxuh.com 91.239.7.58
  osir dns create wxuh.com CNAME www1.wxuh.com wxuh.com
  osir dns create wxuh.com MX wxuh.com mail.wxuh.com --priority 10 --ttl 600`,
		Args: cobra.RangeArgs(1, 4),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return []string{"<domain>\tDomain name (e.g. wxuh.com)"}, cobra.ShellCompDirectiveNoFileComp
			case 1:
				return []string{
					"A\tIPv4 address record",
					"AAAA\tIPv6 address record",
					"CNAME\tCanonical name (alias)",
					"MX\tMail exchange",
					"TXT\tText record",
					"NS\tNameserver",
					"SRV\tService record",
					"CAA\tCertificate authority",
					"PTR\tPointer record",
				}, cobra.ShellCompDirectiveNoFileComp
			case 2:
				return []string{"<name>\tRecord name (e.g. www.wxuh.com)"}, cobra.ShellCompDirectiveNoFileComp
			case 3:
				return []string{"<content>\tRecord value (e.g. IP or hostname)"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			if len(args) != 4 {
				app.Output.PrintError("Usage: dns create <domain> <TYPE> <name> <content>")
				app.Output.Println("")
				app.Output.Println("Examples:")
				app.Output.Println("  dns create wxuh.com A mail.wxuh.com 91.239.7.58")
				app.Output.Println("  dns create wxuh.com CNAME www.wxuh.com target.com")
				app.Output.Println("  dns create wxuh.com MX wxuh.com mail.wxuh.com --priority 10")
				if len(args) == 3 && !isRecordType(args[1]) {
					app.Output.Println("")
					app.Output.Println(fmt.Sprintf("Did you forget the record type? Try:"))
					app.Output.Println(fmt.Sprintf("  dns create %s A %s %s", args[0], args[1], args[2]))
				}
				return fmt.Errorf("expected 4 arguments, got %d", len(args))
			}

			domain := args[0]
			recordType := strings.ToUpper(args[1])
			name := args[2]
			content := args[3]
			ttl, _ := cmd.Flags().GetInt("ttl")
			priority, _ := cmd.Flags().GetInt("priority")

			// CNAME, MX, NS, SRV, PTR content must be a FQDN ending with a dot
			content = ensureTrailingDot(recordType, content)

			req := models.DnsRecordRequest{
				Type:     recordType,
				Name:     name,
				Content:  content,
				TTL:      ttl,
				Priority: priority,
			}

			record, err := app.Client.CreateDnsRecord(ctx, domain, req)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to create DNS record: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(record)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("DNS record created (ID: %s)", record.ID))
				app.Output.PrintKeyValue("Type", record.Type)
				app.Output.PrintKeyValue("Name", record.Name)
				app.Output.PrintKeyValue("Content", record.Content)
				app.Output.PrintKeyValue("TTL", strconv.Itoa(record.TTL))
				if record.Priority > 0 {
					app.Output.PrintKeyValue("Priority", strconv.Itoa(record.Priority))
				}
			}

			return nil
		},
	}

	dnsUpdateCmd := &cobra.Command{
		Use:   "update <domain> <recordId|TYPE> [args...]",
		Short: "Update an existing DNS record",
		Long: `Update a DNS record by ID or by smart type selector.

Smart selectors (when second arg is a DNS type like A, CNAME, MX):
  dns update <domain> <TYPE> <newContent>                     Unique record of that type
  dns update <domain> <TYPE> <nameOrOldContent> <newContent>  Disambiguate by name or old value
  dns update <domain> <TYPE> <name> <oldContent> <newContent> Full disambiguation

Legacy (unchanged):
  dns update <domain> <recordId> [content]                    Direct update by record ID

Use flags (--content, --ttl, --name, --type, --priority) for more control.`,
		Example: `  osir dns update wxuh.com A 91.239.7.59
  osir dns update wxuh.com A mail.wxuh.com 91.239.7.59
  osir dns update wxuh.com A 91.239.7.58 91.239.7.59
  osir dns update wxuh.com A mail.wxuh.com 91.239.7.58 91.239.7.59
  osir dns update wxuh.com wxuh_com__A_989519897 91.239.7.58
  osir dns update wxuh.com wxuh_com__A_989519897 --content 91.239.7.58 --ttl 600`,
		Args: cobra.RangeArgs(2, 5),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return []string{
					"<domain>\tDomain name (e.g. wxuh.com)",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			case 1:
				return []string{
					"A\tIPv4 address record",
					"AAAA\tIPv6 address record",
					"CNAME\tCanonical name (alias)",
					"MX\tMail exchange",
					"TXT\tText record",
					"NS\tNameserver",
					"<recordId>\tRecord ID from 'dns list'",
				}, cobra.ShellCompDirectiveNoFileComp
			case 2:
				return []string{
					"<content>\tNew record value (e.g. IP address)",
					"--ttl\tNew TTL in seconds",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]
			selector := args[1]

			if isRecordType(selector) {
				// Smart mode
				recordType := strings.ToUpper(selector)
				var resolved *models.DnsRecord
				var newContent string

				switch len(args) {
				case 3:
					// dns update <domain> <TYPE> <newContent> — unique record
					newContent = args[2]
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, "", "", "")
					if err != nil {
						return err
					}
					resolved = r

				case 4:
					// dns update <domain> <TYPE> <nameOrOldContent> <newContent>
					newContent = args[3]
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, "", "", args[2])
					if err != nil {
						return err
					}
					resolved = r

				case 5:
					// dns update <domain> <TYPE> <name> <oldContent> <newContent>
					newContent = args[4]
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, args[2], args[3], "")
					if err != nil {
						return err
					}
					resolved = r

				default:
					app.Output.PrintError("Smart update requires: <TYPE> <newContent>, <TYPE> <nameOrOld> <newContent>, or <TYPE> <name> <old> <new>")
					return fmt.Errorf("invalid number of arguments for smart update")
				}

				// Build update request from existing record
				req := models.DnsRecordUpdateRequest{
					Type:     resolved.Type,
					Name:     resolved.Name,
					Content:  newContent,
					TTL:      resolved.TTL,
					Priority: resolved.Priority,
				}

				// Allow flags to override
				if cmd.Flags().Changed("content") {
					content, _ := cmd.Flags().GetString("content")
					req.Content = content
				}
				if cmd.Flags().Changed("type") {
					rt, _ := cmd.Flags().GetString("type")
					req.Type = rt
				}
				if cmd.Flags().Changed("name") {
					name, _ := cmd.Flags().GetString("name")
					req.Name = name
				}
				if cmd.Flags().Changed("ttl") {
					ttl, _ := cmd.Flags().GetInt("ttl")
					req.TTL = ttl
				}
				if cmd.Flags().Changed("priority") {
					priority, _ := cmd.Flags().GetInt("priority")
					req.Priority = priority
				}

				req.Content = ensureTrailingDot(req.Type, req.Content)

				record, err := app.Client.UpdateDnsRecord(ctx, domain, resolved.ID, req)
				if err != nil {
					app.Output.PrintError(fmt.Sprintf("Failed to update DNS record: %s", err))
					return err
				}

				if app.Output.IsJSON() {
					app.Output.PrintResult(record)
				} else {
					app.Output.PrintSuccess(fmt.Sprintf("DNS record %s updated", resolved.ID))
					app.Output.PrintKeyValue("Type", record.Type)
					app.Output.PrintKeyValue("Name", record.Name)
					app.Output.PrintKeyValue("Content", record.Content)
					app.Output.PrintKeyValue("TTL", strconv.Itoa(record.TTL))
					if record.Priority > 0 {
						app.Output.PrintKeyValue("Priority", strconv.Itoa(record.Priority))
					}
				}

				return nil
			}

			// Legacy mode: selector is a record ID
			recordId := selector

			// Fetch existing record to use as base
			existing, err := app.Client.GetDnsRecord(ctx, domain, recordId)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to fetch existing record: %s", err))
				return err
			}

			// Start from existing values, override with user-specified flags
			req := models.DnsRecordUpdateRequest{
				Type:     existing.Type,
				Name:     existing.Name,
				Content:  existing.Content,
				TTL:      existing.TTL,
				Priority: existing.Priority,
			}

			// 3rd positional arg is shorthand for --content
			if len(args) == 3 {
				req.Content = args[2]
			}
			if cmd.Flags().Changed("content") {
				content, _ := cmd.Flags().GetString("content")
				req.Content = content
			}
			if cmd.Flags().Changed("type") {
				recordType, _ := cmd.Flags().GetString("type")
				req.Type = recordType
			}
			if cmd.Flags().Changed("name") {
				name, _ := cmd.Flags().GetString("name")
				req.Name = name
			}
			if cmd.Flags().Changed("ttl") {
				ttl, _ := cmd.Flags().GetInt("ttl")
				req.TTL = ttl
			}
			if cmd.Flags().Changed("priority") {
				priority, _ := cmd.Flags().GetInt("priority")
				req.Priority = priority
			}

			// CNAME, MX, NS, SRV, PTR content must be a FQDN ending with a dot
			req.Content = ensureTrailingDot(req.Type, req.Content)

			record, err := app.Client.UpdateDnsRecord(ctx, domain, recordId, req)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to update DNS record: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(record)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("DNS record %s updated", recordId))
				app.Output.PrintKeyValue("Type", record.Type)
				app.Output.PrintKeyValue("Name", record.Name)
				app.Output.PrintKeyValue("Content", record.Content)
				app.Output.PrintKeyValue("TTL", strconv.Itoa(record.TTL))
				if record.Priority > 0 {
					app.Output.PrintKeyValue("Priority", strconv.Itoa(record.Priority))
				}
			}

			return nil
		},
	}

	dnsDeleteCmd := &cobra.Command{
		Use:   "delete <domain> <recordId|TYPE> [nameOrContent] [content]",
		Short: "Delete a DNS record",
		Long: `Delete a DNS record by ID or by smart type selector.

Smart selectors (when second arg is a DNS type like A, CNAME, MX):
  dns delete <domain> <TYPE>                          Unique record of that type
  dns delete <domain> <TYPE> <nameOrContent>          Disambiguate by name or content
  dns delete <domain> <TYPE> <name> <content>         Full disambiguation

Legacy (unchanged):
  dns delete <domain> <recordId>                      Direct delete by record ID

Safety: SOA records cannot be deleted (use 'dns fix-soa' instead).
        NS records at zone apex require --force.`,
		Example: `  osir dns delete wxuh.com wxuh_com__A_989519897
  osir dns delete wxuh.com CNAME
  osir dns delete wxuh.com CNAME www3.wxuh.com
  osir dns delete wxuh.com A mail.wxuh.com 91.239.7.58`,
		Args: cobra.RangeArgs(2, 4),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return []string{
					"<domain>\tDomain name (e.g. wxuh.com)",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			case 1:
				return []string{
					"A\tIPv4 address record",
					"AAAA\tIPv6 address record",
					"CNAME\tCanonical name (alias)",
					"MX\tMail exchange",
					"TXT\tText record",
					"NS\tNameserver",
					"<recordId>\tRecord ID from 'dns list'",
				}, cobra.ShellCompDirectiveNoFileComp
			case 2:
				return []string{"<nameOrContent>\tFilter by name or content"}, cobra.ShellCompDirectiveNoFileComp
			case 3:
				return []string{"<content>\tFilter by content"}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]
			selector := args[1]
			force, _ := cmd.Flags().GetBool("force")

			if isRecordType(selector) {
				// Smart mode
				recordType := strings.ToUpper(selector)

				// SOA protection
				if recordType == "SOA" {
					app.Output.PrintError("Cannot delete SOA records. Use 'dns fix-soa " + domain + "' to repair the SOA record instead.")
					return fmt.Errorf("SOA records cannot be deleted")
				}

				var resolved *models.DnsRecord

				switch len(args) {
				case 2:
					// dns delete <domain> <TYPE>
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, "", "", "")
					if err != nil {
						return err
					}
					resolved = r

				case 3:
					// dns delete <domain> <TYPE> <nameOrContent>
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, "", "", args[2])
					if err != nil {
						return err
					}
					resolved = r

				case 4:
					// dns delete <domain> <TYPE> <name> <content>
					r, err := resolveDnsRecord(app.Client, cmd, domain, recordType, args[2], args[3], "")
					if err != nil {
						return err
					}
					resolved = r
				}

				// NS at zone apex protection
				if recordType == "NS" && !force {
					if normalizeRecordName(resolved.Name) == normalizeRecordName(domain) {
						app.Output.PrintError(fmt.Sprintf("Refusing to delete NS record at zone apex (%s). Use --force to override.", resolved.Name))
						return fmt.Errorf("NS record at zone apex requires --force")
					}
				}

				_, err := app.Client.DeleteDnsRecord(ctx, domain, resolved.ID)
				if err != nil {
					app.Output.PrintError(fmt.Sprintf("Failed to delete DNS record: %s", err))
					return err
				}

				app.Output.PrintSuccess(fmt.Sprintf("DNS record %s deleted from %s (%s %s → %s)", resolved.ID, domain, resolved.Type, resolved.Name, resolved.Content))
				return nil
			}

			// Legacy mode: selector is a record ID
			recordId := selector

			_, err := app.Client.DeleteDnsRecord(ctx, domain, recordId)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to delete DNS record: %s", err))
				return err
			}

			app.Output.PrintSuccess(fmt.Sprintf("DNS record %s deleted from %s", recordId, domain))
			return nil
		},
	}

	dnsZoneInitCmd := &cobra.Command{
		Use:   "zone-init <domain>",
		Short: "Create a DNS zone with OSIR default records (NS, A, CNAME, SOA)",
		Long: `Create a DNS zone with OSIR defaults for domains using ns1.osir.com and ns3.osir.com.
Sets up NS, A, CNAME, and SOA records automatically.`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain pointed to ns1/ns3.osir.com",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			domain := args[0]

			result, err := app.Client.CreateZoneWithOsirDefaults(ctx, domain)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to initialize zone: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("DNS zone initialized with OSIR defaults for %s", domain))
			}
			return nil
		},
	}

	dnsZoneExistsCmd := &cobra.Command{
		Use:   "zone-exists <domain>",
		Short: "Check if a DNS zone exists",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain to check",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.CheckZoneExists(ctx, args[0])
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to check zone: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if result.Exists {
					app.Output.PrintSuccess(fmt.Sprintf("Zone exists for %s", args[0]))
				} else {
					app.Output.Println(fmt.Sprintf("No zone found for %s", args[0]))
				}
			}
			return nil
		},
	}

	dnsFixSoaCmd := &cobra.Command{
		Use:   "fix-soa <domain>",
		Short: "Fix the SOA record for a zone",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain to fix SOA for",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.FixSOARecord(ctx, args[0])
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to fix SOA record: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.PrintSuccess(fmt.Sprintf("SOA record fixed for %s", args[0]))
			}
			return nil
		},
	}

	dnsDnssecStatusCmd := &cobra.Command{
		Use:   "dnssec-status <domain>",
		Short: "Get DNSSEC status for a zone",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain to check DNSSEC for",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetDnssecStatus(ctx, args[0])
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get DNSSEC status: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				status := "disabled"
				if result.Enabled {
					status = "enabled"
				}
				app.Output.PrintKeyValue("DNSSEC", status)
			}
			return nil
		},
	}

	dnsDnssecEnableCmd := &cobra.Command{
		Use:   "dnssec-enable <domain>",
		Short: "Enable DNSSEC for a zone",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain to enable DNSSEC for",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()
			domain := args[0]

			result, err := app.Client.EnableDnssec(ctx, domain)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to enable DNSSEC: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				printRegistryResult(app, domain, "enabled", result.RegistryPublished, result.RegistryPublishErr)
				// If not on OSIR nameservers, show DS records for manual publishing
				if result.RegistryPublished == nil && result.DsRecords != nil {
					app.Output.Println("Your domain does not use OSIR nameservers. Publish the DS records at your DNS provider:")
					app.Output.PrintKeyValue("DS Records", fmt.Sprintf("%v", result.DsRecords))
				}
			}
			return nil
		},
	}

	dnsDnssecDisableCmd := &cobra.Command{
		Use:   "dnssec-disable <domain>",
		Short: "Disable DNSSEC for a zone",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{
					"<domain>\tDomain to disable DNSSEC for",
					"--help\tShow full help for this command",
				}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()
			domain := args[0]

			result, err := app.Client.DisableDnssec(ctx, domain)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to disable DNSSEC: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				printRegistryResult(app, domain, "disabled", result.RegistryPublished, result.RegistryPublishErr)
			}
			return nil
		},
	}

	dnsCreateCmd.Flags().Int("ttl", 3600, "Time to live in seconds (default: 3600)")
	dnsCreateCmd.Flags().Int("priority", 0, "Record priority (for MX, SRV records)")

	dnsUpdateCmd.Flags().String("type", "", "Record type (A, AAAA, CNAME, MX, TXT, etc.)")
	dnsUpdateCmd.Flags().String("name", "", "Record name (e.g., www, mail, @)")
	dnsUpdateCmd.Flags().String("content", "", "Record content (e.g., IP address, hostname)")
	dnsUpdateCmd.Flags().Int("ttl", 0, "Time to live in seconds")
	dnsUpdateCmd.Flags().Int("priority", 0, "Record priority (for MX, SRV records)")

	dnsDeleteCmd.Flags().Bool("force", false, "Force deletion of protected records (e.g., NS at zone apex)")

	dnsCmd.AddCommand(dnsListCmd, dnsGetCmd, dnsCreateCmd, dnsUpdateCmd, dnsDeleteCmd,
		dnsZoneInitCmd, dnsZoneExistsCmd, dnsFixSoaCmd,
		dnsDnssecStatusCmd, dnsDnssecEnableCmd, dnsDnssecDisableCmd)
	parent.AddCommand(dnsCmd)
}
