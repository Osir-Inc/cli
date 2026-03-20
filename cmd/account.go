package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func addAccountCommands(parent *cobra.Command) {
	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Account management",
		Long:  "View your account profile and summary information.",
	}

	accountProfileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Show user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetProfile(ctx)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get profile: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
			} else {
				if result.Details != nil {
					d := result.Details
					app.Output.PrintKeyValue("Name", d.Name+" "+d.Surname)
					app.Output.PrintKeyValue("Email", d.Email)
					if d.Organization != "" {
						app.Output.PrintKeyValue("Organization", d.Organization)
					}
					if d.Phone != "" {
						app.Output.PrintKeyValue("Phone", d.Phone)
					}
					address := d.Street
					if d.Street2 != "" {
						address += " " + d.Street2
					}
					if d.City != "" {
						address += ", " + d.City
					}
					if d.PostalCode != "" {
						address += " " + d.PostalCode
					}
					if d.Country != "" {
						address += ", " + d.Country
					}
					app.Output.PrintKeyValue("Address", address)
				}
				app.Output.PrintKeyValue("Active", fmt.Sprintf("%t", result.Active))
				app.Output.PrintKeyValue("Domains", fmt.Sprintf("%d", result.TotalDomains))
				app.Output.PrintKeyValue("Orders", fmt.Sprintf("%d", result.OrderCount))
				if result.Balance != nil {
					currency := "USD"
					if result.Balance.Currency != nil {
						currency = *result.Balance.Currency
					}
					app.Output.PrintKeyValue("Balance", fmt.Sprintf("$%.2f %s", float64(result.Balance.Amount)/100.0, currency))
				}
				if result.LastLogin != "" {
					app.Output.PrintKeyValue("Last Login", formatDate(result.LastLogin))
				}
				if result.CreatedAt != "" {
					app.Output.PrintKeyValue("Member Since", formatDate(result.CreatedAt))
				}
			}

			return nil
		},
	}

	accountSummaryCmd := &cobra.Command{
		Use:   "summary",
		Short: "Show account summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetAccountSummary(ctx)
			if err != nil {
				app.Output.PrintError(fmt.Sprintf("Failed to get account summary: %s", err))
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			// Overview
			app.Output.Println("=== Account Summary ===")
			app.Output.Println("")

			// Balance
			if len(result.Balances) > 0 {
				bal := result.Balances[0]
				currency := "USD"
				if bal.Currency != nil {
					currency = *bal.Currency
				}
				app.Output.PrintKeyValue("Balance", fmt.Sprintf("$%.2f %s", float64(bal.Amount)/100.0, currency))
			}

			// Counts
			app.Output.PrintKeyValue("Domains", fmt.Sprintf("%d", result.DomainCount))
			app.Output.PrintKeyValue("Contacts", fmt.Sprintf("%d", result.ContactCount))
			app.Output.PrintKeyValue("VPS Instances", fmt.Sprintf("%d", result.VpsCount))
			app.Output.PrintKeyValue("Talents", fmt.Sprintf("%d", result.TalentCount))
			app.Output.Println("")

			// Recent domains (last 5)
			if len(result.DomainRegistry) > 0 {
				app.Output.Println("--- Recent Domains ---")
				count := 5
				if len(result.DomainRegistry) < count {
					count = len(result.DomainRegistry)
				}
				for i := 0; i < count; i++ {
					d := result.DomainRegistry[len(result.DomainRegistry)-1-i]
					statuses := ""
					if len(d.Statuses) > 0 {
						statuses = " [" + strings.Join(d.Statuses, ", ") + "]"
					}
					app.Output.Printf("  %-30s expires: %s%s\n", d.Domain, formatDate(d.ExpirationDate), statuses)
				}
				app.Output.Println("")
			}

			// Last order
			if len(result.Orders) > 0 {
				o := result.Orders[0]
				app.Output.Println("--- Last Order ---")
				app.Output.PrintKeyValue("Domain", o.Domain)
				app.Output.PrintKeyValue("Type", o.Type)
				app.Output.PrintKeyValue("Status", o.Status)
				app.Output.PrintKeyValue("Amount", fmt.Sprintf("$%.2f", float64(o.Amount)/100.0))
				app.Output.PrintKeyValue("Date", formatDate(o.OrderDate))
				if o.Message != "" {
					app.Output.PrintKeyValue("Message", o.Message)
				}
				app.Output.Println("")
			}

			// Last message
			if len(result.Messages) > 0 {
				m := result.Messages[0]
				app.Output.Println("--- Last Message ---")
				app.Output.PrintKeyValue("Subject", m.Subject)
				app.Output.Println("")
			}

			// Last balance change
			if len(result.Balances) > 0 {
				b := result.Balances[0]
				app.Output.Println("--- Last Balance Change ---")
				app.Output.PrintKeyValue("Source", b.Source)
				app.Output.PrintKeyValue("Date", formatDate(b.LastChangedDateTime))
			}

			return nil
		},
	}

	accountCmd.AddCommand(accountProfileCmd, accountSummaryCmd)
	parent.AddCommand(accountCmd)
}

func formatDate(datetime string) string {
	if len(datetime) >= 10 {
		return datetime[:10]
	}
	return datetime
}
