package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/osir/cli/internal/api/models"
	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

func addContactCommands(parent *cobra.Command) {
	contactCmd := &cobra.Command{
		Use:   "contact",
		Short: "Contact management commands",
		Long:  "Manage contacts: list, create, update, delete, and view domain contacts.",
	}

	contactListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all contacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			result, err := app.Client.ListContacts(ctx, page, size)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printContactList(app.Output, result)
			return nil
		},
	}

	contactGetCmd := &cobra.Command{
		Use:   "get <contactId>",
		Short: "Get contact details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetContact(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printContactDetail(app.Output, result)
			return nil
		},
	}

	contactCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contact",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")
			email, _ := cmd.Flags().GetString("email")
			phone, _ := cmd.Flags().GetString("phone")
			organization, _ := cmd.Flags().GetString("organization")
			street1, _ := cmd.Flags().GetString("street1")
			street2, _ := cmd.Flags().GetString("street2")
			city, _ := cmd.Flags().GetString("city")
			state, _ := cmd.Flags().GetString("state")
			postalCode, _ := cmd.Flags().GetString("postal-code")
			country, _ := cmd.Flags().GetString("country")

			if firstName == "" {
				app.Output.PrintError("First name is required (use --first-name)")
				return fmt.Errorf("first name is required")
			}
			if lastName == "" {
				app.Output.PrintError("Last name is required (use --last-name)")
				return fmt.Errorf("last name is required")
			}
			if email == "" || !strings.Contains(email, "@") {
				app.Output.PrintError("A valid email is required (use --email)")
				return fmt.Errorf("valid email is required")
			}

			req := models.ContactCreateRequest{
				FirstName:    firstName,
				LastName:     lastName,
				Email:        email,
				Phone:        phone,
				Organization: organization,
				Street1:      street1,
				Street2:      street2,
				City:         city,
				State:        state,
				PostalCode:   postalCode,
				Country:      country,
			}

			result, err := app.Client.CreateContact(ctx, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	contactUpdateCmd := &cobra.Command{
		Use:   "update <contactId>",
		Short: "Update an existing contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			firstName, _ := cmd.Flags().GetString("first-name")
			lastName, _ := cmd.Flags().GetString("last-name")
			email, _ := cmd.Flags().GetString("email")
			phone, _ := cmd.Flags().GetString("phone")
			organization, _ := cmd.Flags().GetString("organization")
			street1, _ := cmd.Flags().GetString("street1")
			street2, _ := cmd.Flags().GetString("street2")
			city, _ := cmd.Flags().GetString("city")
			state, _ := cmd.Flags().GetString("state")
			postalCode, _ := cmd.Flags().GetString("postal-code")
			country, _ := cmd.Flags().GetString("country")

			req := models.ContactCreateRequest{
				FirstName:    firstName,
				LastName:     lastName,
				Email:        email,
				Phone:        phone,
				Organization: organization,
				Street1:      street1,
				Street2:      street2,
				City:         city,
				State:        state,
				PostalCode:   postalCode,
				Country:      country,
			}

			result, err := app.Client.UpdateContact(ctx, args[0], req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	contactDeleteCmd := &cobra.Command{
		Use:   "delete <contactId>",
		Short: "Delete a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.DeleteContact(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	contactForDomainCmd := &cobra.Command{
		Use:   "for-domain <domain>",
		Short: "Get contacts associated with a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetDomainContacts(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	// list command flags
	contactListCmd.Flags().Int("page", 0, "Page number")
	contactListCmd.Flags().Int("size", 20, "Page size")

	// create command flags
	contactCreateCmd.Flags().String("first-name", "", "First name")
	contactCreateCmd.Flags().String("last-name", "", "Last name")
	contactCreateCmd.Flags().String("email", "", "Email address")
	contactCreateCmd.Flags().String("phone", "", "Phone number")
	contactCreateCmd.Flags().String("organization", "", "Organization name")
	contactCreateCmd.Flags().String("street1", "", "Street address line 1")
	contactCreateCmd.Flags().String("street2", "", "Street address line 2")
	contactCreateCmd.Flags().String("city", "", "City")
	contactCreateCmd.Flags().String("state", "", "State or province")
	contactCreateCmd.Flags().String("postal-code", "", "Postal code")
	contactCreateCmd.Flags().String("country", "", "Country code")
	_ = contactCreateCmd.MarkFlagRequired("first-name")
	_ = contactCreateCmd.MarkFlagRequired("last-name")
	_ = contactCreateCmd.MarkFlagRequired("email")

	// update command flags
	contactUpdateCmd.Flags().String("first-name", "", "First name")
	contactUpdateCmd.Flags().String("last-name", "", "Last name")
	contactUpdateCmd.Flags().String("email", "", "Email address")
	contactUpdateCmd.Flags().String("phone", "", "Phone number")
	contactUpdateCmd.Flags().String("organization", "", "Organization name")
	contactUpdateCmd.Flags().String("street1", "", "Street address line 1")
	contactUpdateCmd.Flags().String("street2", "", "Street address line 2")
	contactUpdateCmd.Flags().String("city", "", "City")
	contactUpdateCmd.Flags().String("state", "", "State or province")
	contactUpdateCmd.Flags().String("postal-code", "", "Postal code")
	contactUpdateCmd.Flags().String("country", "", "Country code")

	contactCmd.AddCommand(
		contactListCmd,
		contactGetCmd,
		contactCreateCmd,
		contactUpdateCmd,
		contactDeleteCmd,
		contactForDomainCmd,
	)

	parent.AddCommand(contactCmd)
}

func printContactList(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
		Data     []struct {
			ID           int    `json:"id"`
			FirstName    string `json:"firstName"`
			LastName     string `json:"lastName"`
			Email        string `json:"email"`
			Organization string `json:"organization"`
			Country      string `json:"country"`
			DomainCount  int    `json:"domainCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	totalPages := 1
	if resp.PageSize > 0 {
		totalPages = (resp.Total + resp.PageSize - 1) / resp.PageSize
	}
	out.Println(fmt.Sprintf("Total: %d (page %d of %d)", resp.Total, resp.Page+1, totalPages))
	out.Println("")

	headers := []string{"ID", "NAME", "EMAIL", "ORGANIZATION", "COUNTRY", "DOMAINS"}
	rows := make([][]string, len(resp.Data))
	for i, c := range resp.Data {
		rows[i] = []string{
			strconv.Itoa(c.ID),
			c.FirstName + " " + c.LastName,
			c.Email,
			c.Organization,
			c.Country,
			strconv.Itoa(c.DomainCount),
		}
	}
	out.PrintTable(headers, rows)
}

func printContactDetail(out *output.Formatter, raw json.RawMessage) {
	var c struct {
		ID           int    `json:"id"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		Email        string `json:"email"`
		Phone        string `json:"phone"`
		Organization string `json:"organization"`
		Address      string `json:"address"`
		Country      string `json:"country"`
		DomainCount  int    `json:"domainCount"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		out.PrintResult(raw)
		return
	}

	out.PrintKeyValue("ID", strconv.Itoa(c.ID))
	out.PrintKeyValue("Name", c.FirstName+" "+c.LastName)
	out.PrintKeyValue("Email", c.Email)
	if c.Phone != "" {
		out.PrintKeyValue("Phone", c.Phone)
	}
	if c.Organization != "" {
		out.PrintKeyValue("Organization", c.Organization)
	}
	if c.Address != "" {
		out.PrintKeyValue("Address", c.Address)
	}
	out.PrintKeyValue("Domains", strconv.Itoa(c.DomainCount))
}
