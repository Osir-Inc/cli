package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/osir/cli/internal/api"
	"github.com/osir/cli/internal/api/models"
	"github.com/osir/cli/internal/output"
	"github.com/spf13/cobra"
)

func addBillingCommands(parent *cobra.Command) {
	billingCmd := &cobra.Command{
		Use:   "billing",
		Short: "Billing and payment commands",
		Long:  "Manage billing: check balance, invoices, payments, transactions, fees, and pricing.",
	}

	billingBalanceCmd := &cobra.Command{
		Use:   "balance",
		Short: "Get account balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetBalance(ctx)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printBalance(app.Output, result)
			return nil
		},
	}

	billingInvoicesCmd := &cobra.Command{
		Use:   "invoices",
		Short: "List invoices",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			status, _ := cmd.Flags().GetString("status")
			invoiceType, _ := cmd.Flags().GetString("type")
			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			result, err := app.Client.ListInvoices(ctx, status, invoiceType, page, size)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printInvoiceList(app.Output, result)
			return nil
		},
	}

	billingInvoiceCmd := &cobra.Command{
		Use:   "invoice <invoiceId>",
		Short: "Get invoice details by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetInvoice(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printInvoiceDetail(app.Output, result)
			return nil
		},
	}

	billingInvoiceByNumberCmd := &cobra.Command{
		Use:   "invoice-number <invoiceNumber>",
		Short: "Get invoice by invoice number",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetInvoiceByNumber(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printInvoiceDetail(app.Output, result)
			return nil
		},
	}

	billingPayCmd := &cobra.Command{
		Use:   "pay <invoiceId> <amount>",
		Short: "Pay an invoice from account balance",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			var amount float64
			if _, err := fmt.Sscanf(args[1], "%f", &amount); err != nil {
				app.Output.PrintError(fmt.Sprintf("Invalid amount: %s", args[1]))
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			result, err := app.Client.PayInvoice(ctx, args[0], amount)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	billingStatsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Get invoice statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetInvoiceStatistics(ctx)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printInvoiceStats(app.Output, result)
			return nil
		},
	}

	billingCheckoutCmd := &cobra.Command{
		Use:   "checkout <amount>",
		Short: "Add funds via Stripe checkout",
		Long:  "Creates a Stripe checkout session to add funds to your account and opens it in your browser.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			var amount float64
			if _, err := fmt.Sscanf(args[0], "%f", &amount); err != nil {
				app.Output.PrintError(fmt.Sprintf("Invalid amount: %s", args[0]))
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			currency, _ := cmd.Flags().GetString("currency")
			description, _ := cmd.Flags().GetString("description")

			req := models.CheckoutSessionRequest{
				Processor:   "stripe",
				Amount:      amount,
				Currency:    currency,
				Description: description,
				SuccessURL:  "https://osir.com/payment/success",
				CancelURL:   "https://osir.com/payment/cancel",
			}

			result, err := app.Client.CreateCheckoutSession(ctx, req)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}

			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}

			var resp struct {
				Data struct {
					CheckoutURL string  `json:"checkoutUrl"`
					Amount      float64 `json:"amount"`
					Currency    string  `json:"currency"`
					SessionID   string  `json:"sessionId"`
					ExpiresAt   string  `json:"expiresAt"`
					Processor   string  `json:"processor"`
				} `json:"data"`
			}
			if err := json.Unmarshal(result, &resp); err != nil {
				app.Output.PrintResult(result)
				return nil
			}

			d := resp.Data
			app.Output.PrintKeyValue("Amount", fmt.Sprintf("$%.2f %s", d.Amount, d.Currency))
			app.Output.PrintKeyValue("Processor", d.Processor)
			app.Output.PrintKeyValue("Session", d.SessionID)
			app.Output.PrintKeyValue("Expires", formatDate(d.ExpiresAt))
			app.Output.Println("")

			if d.CheckoutURL != "" {
				app.Output.Println("Opening checkout in browser...")
				_ = openBrowser(d.CheckoutURL)
				app.Output.Println("Waiting for payment confirmation... (Ctrl+C to cancel)")
				app.Output.Println("")
				waitForPayment(ctx, app.Client, app.Output)
			}

			return nil
		},
	}

	billingTransactionsCmd := &cobra.Command{
		Use:   "history",
		Short: "Show balance history",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			page, _ := cmd.Flags().GetInt("page")
			size, _ := cmd.Flags().GetInt("size")

			result, err := app.Client.ListTransactions(ctx, page, size)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printBalanceHistory(app.Output, result)
			return nil
		},
	}

	billingFeesCmd := &cobra.Command{
		Use:   "fees <amount>",
		Short: "Preview payment processing fees",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			var amount float64
			if _, err := fmt.Sscanf(args[0], "%f", &amount); err != nil {
				app.Output.PrintError(fmt.Sprintf("Invalid amount: %s", args[0]))
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			currency, _ := cmd.Flags().GetString("currency")
			processor, _ := cmd.Flags().GetString("processor")

			result, err := app.Client.PreviewFees(ctx, amount, currency, processor)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printFeePreview(app.Output, result)
			return nil
		},
	}

	billingPricingCmd := &cobra.Command{
		Use:   "pricing [extension]",
		Short: "Get domain pricing from catalog",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			extension := ""
			if len(args) > 0 {
				extension = args[0]
			}

			result, err := app.Client.GetDomainPricing(ctx, extension)
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			if app.Output.IsJSON() {
				app.Output.PrintResult(result)
				return nil
			}
			printDomainPricing(app.Output, result)
			return nil
		},
	}

	billingListenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for real-time payment events",
		Long:  "Opens an SSE connection to receive live payment notifications (balance changes, refunds, invoice payments).",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			app.Output.Println("Listening for payment events... (Ctrl+C to stop)")
			app.Output.Println("")

			events := make(chan api.SSEEvent, 10)
			go func() {
				err := app.Client.ListenSSE(ctx, "/v1/payment/events/stream", events)
				if err != nil && ctx.Err() == nil {
					app.Output.PrintError(fmt.Sprintf("SSE connection lost: %s", err))
				}
				close(events)
			}()

			for evt := range events {
				printSSEEvent(app.Output, evt)
			}

			return nil
		},
	}

	billingSessionCmd := &cobra.Command{
		Use:   "session <sessionId>",
		Short: "Get payment session status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := getApp(cmd)
			ctx := cmd.Context()

			result, err := app.Client.GetPaymentSession(ctx, args[0])
			if err != nil {
				app.Output.PrintError(err.Error())
				return err
			}
			app.Output.PrintResult(result)
			return nil
		},
	}

	// invoices command flags
	billingInvoicesCmd.Flags().String("status", "", "Filter by status (e.g. PAID, UNPAID, CANCELLED)")
	billingInvoicesCmd.Flags().String("type", "", "Filter by type")
	billingInvoicesCmd.Flags().Int("page", 0, "Page number")
	billingInvoicesCmd.Flags().Int("size", 50, "Page size")

	// checkout command flags
	billingCheckoutCmd.Flags().String("currency", "USD", "Payment currency")
	billingCheckoutCmd.Flags().String("description", "", "Payment description")

	// transactions command flags
	billingTransactionsCmd.Flags().Int("page", 0, "Page number")
	billingTransactionsCmd.Flags().Int("size", 20, "Page size")

	// fees command flags
	billingFeesCmd.Flags().String("currency", "USD", "Currency code")
	billingFeesCmd.Flags().String("processor", "stripe", "Payment processor")

	billingCmd.AddCommand(
		billingBalanceCmd,
		billingInvoicesCmd,
		billingInvoiceCmd,
		billingInvoiceByNumberCmd,
		billingPayCmd,
		billingStatsCmd,
		billingCheckoutCmd,
		billingTransactionsCmd,
		billingFeesCmd,
		billingPricingCmd,
		billingSessionCmd,
		billingListenCmd,
	)

	parent.AddCommand(billingCmd)
}

func waitForPayment(ctx context.Context, client api.Backend, out *output.Formatter) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	events := make(chan api.SSEEvent, 10)
	go func() {
		_ = client.ListenSSE(ctx, "/v1/payment/events/stream", events)
		close(events)
	}()

	for {
		select {
		case evt, ok := <-events:
			if !ok {
				return
			}
				eventType := getSSEEventType(evt)
			if eventType == "heartbeat" {
				continue
			}
			printSSEEvent(out, evt)
			if eventType == "balance.added" {
				return
			}
		case <-ctx.Done():
			out.Println("Timed out waiting for payment.")
			return
		}
	}
}

func getSSEEventType(evt api.SSEEvent) string {
	// Event type is now inside the JSON data payload
	var data map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &data); err != nil {
		return evt.Event
	}
	if eventType, ok := data["event"].(string); ok {
		return eventType
	}
	// Fallback to SSE event: header
	if evt.Event != "" {
		return evt.Event
	}
	if t, ok := data["type"].(string); ok {
		return t
	}
	return ""
}

func printSSEEvent(out *output.Formatter, evt api.SSEEvent) {
	eventType := getSSEEventType(evt)
	if eventType == "heartbeat" || eventType == "connected" {
		return
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(evt.Data), &data); err != nil {
		out.Println(fmt.Sprintf("[%s] %s", eventType, evt.Data))
		return
	}

	switch eventType {
	case "balance.added":
		out.Println("[OK] Payment received!")
		out.PrintKeyValue("Added", fmt.Sprintf("+$%s %s", data["amountAdded"], data["currency"]))
		out.PrintKeyValue("New Balance", fmt.Sprintf("$%s %s", data["newBalance"], data["currency"]))
		if desc, ok := data["description"]; ok && desc != nil {
			out.PrintKeyValue("Description", fmt.Sprintf("%s", desc))
		}
	case "balance.refunded":
		out.Println("[OK] Refund processed")
		out.PrintKeyValue("Refunded", fmt.Sprintf("-$%s %s", data["amountRefunded"], data["currency"]))
		out.PrintKeyValue("New Balance", fmt.Sprintf("$%s %s", data["newBalance"], data["currency"]))
		if reason, ok := data["reason"]; ok && reason != nil {
			out.PrintKeyValue("Reason", fmt.Sprintf("%s", reason))
		}
	case "invoice.paid":
		out.Println("[OK] Invoice paid")
		out.PrintKeyValue("Invoice", fmt.Sprintf("%s", data["invoiceNumber"]))
		out.PrintKeyValue("Amount", fmt.Sprintf("$%s %s", data["amountPaid"], data["currency"]))
	default:
		out.Println(fmt.Sprintf("[%s] %s", eventType, evt.Data))
	}
	out.Println("")
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func printBalance(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Data struct {
			Balance    float64 `json:"balance"`
			Currency   string  `json:"currency"`
			CustomerID string  `json:"customerId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}
	out.PrintKeyValue("Balance", fmt.Sprintf("$%.2f %s", resp.Data.Balance, resp.Data.Currency))
}

func printFeePreview(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Data struct {
			NetAmount     float64 `json:"netAmount"`
			ProcessingFee float64 `json:"processingFee"`
			AmountToPay   float64 `json:"amountToPay"`
			Currency      string  `json:"currency"`
			Processor     string  `json:"processor"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}
	d := resp.Data
	out.PrintKeyValue("Amount", fmt.Sprintf("$%.2f %s", d.NetAmount, d.Currency))
	out.PrintKeyValue("Processing Fee", fmt.Sprintf("$%.2f", d.ProcessingFee))
	out.PrintKeyValue("Total to Pay", fmt.Sprintf("$%.2f %s", d.AmountToPay, d.Currency))
	out.PrintKeyValue("Processor", d.Processor)
}


func printInvoiceList(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Data struct {
			Invoices []struct {
				ID            int     `json:"id"`
				InvoiceNumber string  `json:"invoiceNumber"`
				InvoiceType   string  `json:"invoiceType"`
				Status        string  `json:"status"`
				TotalAmount   float64 `json:"totalAmount"`
				Currency      string  `json:"currency"`
				IssueDate     string  `json:"issueDate"`
				DueDate       string  `json:"dueDate"`
				PaidDate      *string `json:"paidDate"`
			} `json:"invoices"`
			TotalCount int `json:"totalCount"`
			Page       int `json:"page"`
			TotalPages int `json:"totalPages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	out.Println(fmt.Sprintf("Total: %d (page %d of %d)", resp.Data.TotalCount, resp.Data.Page+1, resp.Data.TotalPages))
	out.Println("")

	headers := []string{"ID", "NUMBER", "TYPE", "STATUS", "AMOUNT", "ISSUED", "DUE"}
	rows := make([][]string, len(resp.Data.Invoices))
	for i, inv := range resp.Data.Invoices {
		rows[i] = []string{
			strconv.Itoa(inv.ID),
			inv.InvoiceNumber,
			inv.InvoiceType,
			inv.Status,
			fmt.Sprintf("$%.2f %s", inv.TotalAmount, inv.Currency),
			inv.IssueDate,
			inv.DueDate,
		}
	}
	out.PrintTable(headers, rows)
}

func printInvoiceDetail(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Data struct {
			Invoice struct {
				ID            int     `json:"id"`
				InvoiceNumber string  `json:"invoiceNumber"`
				InvoiceType   string  `json:"invoiceType"`
				Status        string  `json:"status"`
				TotalAmount   float64 `json:"totalAmount"`
				Subtotal      float64 `json:"subtotal"`
				TaxAmount     float64 `json:"taxAmount"`
				PaidAmount    float64 `json:"paidAmount"`
				Remaining     float64 `json:"remainingAmount"`
				Currency      string  `json:"currency"`
				IssueDate     string  `json:"issueDate"`
				DueDate       string  `json:"dueDate"`
				PaidDate      *string `json:"paidDate"`
				PaymentMethod *string `json:"paymentMethod"`
				Items         []struct {
					ID          int     `json:"id"`
					Description string  `json:"description"`
					Quantity    int     `json:"quantity"`
					UnitPrice   float64 `json:"unitPrice"`
					TotalPrice  float64 `json:"totalPrice"`
				} `json:"items"`
			} `json:"invoice"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	inv := resp.Data.Invoice
	out.PrintKeyValue("Invoice", inv.InvoiceNumber)
	out.PrintKeyValue("ID", strconv.Itoa(inv.ID))
	out.PrintKeyValue("Type", inv.InvoiceType)
	out.PrintKeyValue("Status", inv.Status)
	out.PrintKeyValue("Issue Date", inv.IssueDate)
	out.PrintKeyValue("Due Date", inv.DueDate)
	if inv.PaidDate != nil {
		out.PrintKeyValue("Paid Date", formatDate(*inv.PaidDate))
	}
	if inv.PaymentMethod != nil {
		out.PrintKeyValue("Payment Method", *inv.PaymentMethod)
	}
	out.Println("")

	if len(inv.Items) > 0 {
		out.Println("--- Items ---")
		headers := []string{"DESCRIPTION", "QTY", "UNIT PRICE", "TOTAL"}
		rows := make([][]string, len(inv.Items))
		for i, item := range inv.Items {
			rows[i] = []string{
				item.Description,
				strconv.Itoa(item.Quantity),
				fmt.Sprintf("$%.2f", item.UnitPrice),
				fmt.Sprintf("$%.2f", item.TotalPrice),
			}
		}
		out.PrintTable(headers, rows)
		out.Println("")
	}

	out.PrintKeyValue("Subtotal", fmt.Sprintf("$%.2f", inv.Subtotal))
	if inv.TaxAmount > 0 {
		out.PrintKeyValue("Tax", fmt.Sprintf("$%.2f", inv.TaxAmount))
	}
	out.PrintKeyValue("Total", fmt.Sprintf("$%.2f %s", inv.TotalAmount, inv.Currency))
	out.PrintKeyValue("Paid", fmt.Sprintf("$%.2f", inv.PaidAmount))
	if inv.Remaining > 0 {
		out.PrintKeyValue("Remaining", fmt.Sprintf("$%.2f", inv.Remaining))
	}
}

func printInvoiceStats(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Data struct {
			Statistics struct {
				TotalInvoices      int `json:"totalInvoices"`
				PendingCount       int `json:"pendingCount"`
				PaidCount          int `json:"paidCount"`
				OverdueCount       int `json:"overdueCount"`
				CancelledCount     int `json:"cancelledCount"`
				TotalAmountPending int `json:"totalAmountPending"`
				TotalAmountPaid    int `json:"totalAmountPaid"`
				TotalAmountOverdue int `json:"totalAmountOverdue"`
			} `json:"statistics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}
	s := resp.Data.Statistics
	out.PrintKeyValue("Total Invoices", strconv.Itoa(s.TotalInvoices))
	out.PrintKeyValue("Paid", fmt.Sprintf("%d ($%.2f)", s.PaidCount, float64(s.TotalAmountPaid)/100.0))
	out.PrintKeyValue("Pending", fmt.Sprintf("%d ($%.2f)", s.PendingCount, float64(s.TotalAmountPending)/100.0))
	out.PrintKeyValue("Overdue", fmt.Sprintf("%d ($%.2f)", s.OverdueCount, float64(s.TotalAmountOverdue)/100.0))
	if s.CancelledCount > 0 {
		out.PrintKeyValue("Cancelled", strconv.Itoa(s.CancelledCount))
	}
}

func printBalanceHistory(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
		Total    int `json:"total"`
		Data     []struct {
			ID                  int     `json:"id"`
			Amount              int     `json:"amount"`
			Source              string  `json:"source"`
			Operator            *string `json:"operator"`
			OrderID             *int    `json:"orderId"`
			LastChangedDateTime string  `json:"lastChangedDateTime"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	if len(resp.Data) == 0 {
		out.Println("No balance history found")
		return
	}

	totalPages := (resp.Total + resp.PageSize - 1) / resp.PageSize
	out.Println(fmt.Sprintf("Total: %d (page %d of %d)", resp.Total, resp.Page+1, totalPages))
	out.Println("")

	headers := []string{"ID", "BALANCE", "SOURCE", "DATE"}
	rows := make([][]string, len(resp.Data))
	for i, e := range resp.Data {
		date := ""
		if len(e.LastChangedDateTime) >= 16 {
			date = e.LastChangedDateTime[:16]
		}
		rows[i] = []string{
			strconv.Itoa(e.ID),
			fmt.Sprintf("$%.2f", float64(e.Amount)/100.0),
			e.Source,
			date,
		}
	}
	out.PrintTable(headers, rows)
}

func printDomainPricing(out *output.Formatter, raw json.RawMessage) {
	var resp struct {
		Extensions []struct {
			Extension             string  `json:"extension"`
			ExtensionType         string  `json:"extensionType"`
			Registrar             string  `json:"registrar"`
			RegistrationPrice     float64 `json:"registrationPrice"`
			RenewalPrice          float64 `json:"renewalPrice"`
			TransferPrice         float64 `json:"transferPrice"`
			RestorePrice          float64 `json:"restorePrice"`
			MinRegistrationPeriod int     `json:"minRegistrationPeriod"`
			MaxRegistrationPeriod int     `json:"maxRegistrationPeriod"`
			MinCharacters         int     `json:"minCharacters"`
			MaxCharacters         int     `json:"maxCharacters"`
			SupportsWhoisPrivacy  bool    `json:"supportsWhoisPrivacy"`
			SupportsRegistrarLock bool    `json:"supportsRegistrarLock"`
			HasDNSSEC             bool    `json:"hasDNSSEC"`
		} `json:"extensions"`
		TotalExtensions int `json:"totalExtensions"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		out.PrintResult(raw)
		return
	}

	out.Println(fmt.Sprintf("Total extensions: %d", resp.TotalExtensions))
	out.Println("")

	headers := []string{"EXTENSION", "TYPE", "REGISTER", "RENEW", "TRANSFER", "RESTORE", "REGISTRAR"}
	rows := make([][]string, len(resp.Extensions))
	for i, ext := range resp.Extensions {
		rows[i] = []string{
			ext.Extension,
			ext.ExtensionType,
			fmt.Sprintf("$%.2f", ext.RegistrationPrice/100),
			fmt.Sprintf("$%.2f", ext.RenewalPrice/100),
			fmt.Sprintf("$%.2f", ext.TransferPrice/100),
			fmt.Sprintf("$%.2f", ext.RestorePrice/100),
			ext.Registrar,
		}
	}
	out.PrintTable(headers, rows)
}
