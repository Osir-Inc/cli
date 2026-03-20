package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) GetBalance(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, "/v1/payment/balance", nil, &result)
	return result, err
}

func (c *Client) ListInvoices(ctx context.Context, status, invoiceType string, page, size int) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	if invoiceType != "" {
		query.Set("type", invoiceType)
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	err := c.Get(ctx, "/v1/billing/invoices", query, &result)
	return result, err
}

func (c *Client) GetInvoice(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, fmt.Sprintf("/v1/billing/invoices/%s", id), nil, &result)
	return result, err
}

func (c *Client) GetInvoiceByNumber(ctx context.Context, number string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, fmt.Sprintf("/v1/billing/invoices/number/%s", number), nil, &result)
	return result, err
}

func (c *Client) GetInvoiceStatistics(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, "/v1/billing/invoices/statistics", nil, &result)
	return result, err
}

func (c *Client) PayInvoice(ctx context.Context, id string, amount float64) (json.RawMessage, error) {
	var result json.RawMessage
	req := models.PayInvoiceRequest{Amount: amount}
	err := c.Post(ctx, fmt.Sprintf("/v1/billing/invoices/%s/pay", id), req, &result)
	return result, err
}

func (c *Client) CreateCheckoutSession(ctx context.Context, req models.CheckoutSessionRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, "/v1/payment/checkout-session", req, &result)
	return result, err
}

func (c *Client) PreviewFees(ctx context.Context, amount float64, currency, processor string) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	query.Set("amount", strconv.FormatFloat(amount, 'f', 2, 64))
	if currency != "" {
		query.Set("currency", currency)
	}
	if processor != "" {
		query.Set("processor", processor)
	}
	err := c.Get(ctx, "/v1/payment/fee-preview", query, &result)
	return result, err
}

func (c *Client) GetPaymentSession(ctx context.Context, sessionID string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, fmt.Sprintf("/v1/payment/session/%s", sessionID), nil, &result)
	return result, err
}

func (c *Client) ListTransactions(ctx context.Context, page, size int) (json.RawMessage, error) {
	// Old /v1/payment/transactions was a stub; now use balance history
	return c.GetBalanceHistory(ctx, page, size)
}

func (c *Client) GetBalanceHistory(ctx context.Context, page, size int) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	err := c.Get(ctx, "/v1/payment/balance/history", query, &result)
	return result, err
}

func (c *Client) GetDomainPricing(ctx context.Context, extension string) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if extension != "" {
		query.Set("extension", extension)
	}
	err := c.Get(ctx, "/v1/public/catalog/domains", query, &result)
	return result, err
}
