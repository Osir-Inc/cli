package cmd

import (
	"fmt"
	"strconv"

	"github.com/osir/cli/internal/api/models"
	"github.com/spf13/cobra"
)

func addAuditCommands(parent *cobra.Command) {
	auditCmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit log management",
		Long:  "View audit trails: recent activity, domain-specific logs, and failed operations.",
	}

	auditRecentCmd := &cobra.Command{
		Use:   "recent",
		Short: "Get recent audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			limit, _ := cmd.Flags().GetInt("limit")

			entries, err := app.Client.GetRecentActivity(ctx, limit)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get recent activity: %s", err))
				return err
			}

			if len(entries) == 0 {
				app.Output.Println("No recent activity found")
				return nil
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(entries)
			} else {
				printAuditTable(app.Output, entries)
			}

			return nil
		},
	}

	auditDomainCmd := &cobra.Command{
		Use:   "domain <domain>",
		Short: "Get audit trail for a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			result, err := app.Client.GetDomainAudit(ctx, args[0], page, size)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get domain audit: %s", err))
				return err
			}

			if len(result.Data) == 0 {
				app.Output.Println(fmt.Sprintf("No audit entries found for %s", args[0]))
				return nil
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.Println(fmt.Sprintf("Total entries: %d", result.Total))
				printAuditTable(app.Output, result.Data)
			}

			return nil
		},
	}

	auditFailuresCmd := &cobra.Command{
		Use:   "failures",
		Short: "Get failed operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			result, err := app.Client.GetFailedOperations(ctx, page, size)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get failed operations: %s", err))
				return err
			}

			if len(result.Data) == 0 {
				app.Output.Println("No failed operations found")
				return nil
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				app.Output.Println(fmt.Sprintf("Total entries: %d", result.Total))
				printAuditTable(app.Output, result.Data)
			}

			return nil
		},
	}

	auditRecentCmd.Flags().Int("limit", 50, "Number of entries to retrieve")
	auditDomainCmd.Flags().Int("page", 0, "Page number")
	auditDomainCmd.Flags().Int("size", 20, "Page size")
	auditFailuresCmd.Flags().Int("page", 0, "Page number")
	auditFailuresCmd.Flags().Int("size", 20, "Page size")

	auditCmd.AddCommand(auditRecentCmd, auditDomainCmd, auditFailuresCmd)
	parent.AddCommand(auditCmd)
}

func printAuditTable(out interface {
	PrintTable(headers []string, rows [][]string)
}, entries []models.AuditEntry) {
	headers := []string{"ID", "ACTION", "DOMAIN", "SUCCESS", "IP", "DATE"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		domain := "-"
		if e.Domain != nil {
			domain = *e.Domain
		}
		ip := "-"
		if e.ClientIP != nil {
			ip = *e.ClientIP
		}
		date := ""
		if len(e.CreatedAt) >= 16 {
			date = e.CreatedAt[:16]
		} else {
			date = e.CreatedAt
		}
		rows[i] = []string{
			strconv.Itoa(e.ID),
			e.Action,
			domain,
			strconv.FormatBool(e.Success),
			ip,
			date,
		}
	}
	out.PrintTable(headers, rows)
}
