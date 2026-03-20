package models

// GET /v1/payment/balance - response is a balance object (same as SummaryBalance)

// GET /v1/billing/invoices - paginated
// GET /v1/billing/invoices/{id}
// GET /v1/billing/invoices/number/{invoiceNumber}
// GET /v1/billing/invoices/statistics
// POST /v1/billing/invoices/{id}/pay

// POST /v1/payment/checkout-session
type CheckoutSessionRequest struct {
	Processor  string  `json:"processor"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Description string `json:"description,omitempty"`
	SuccessURL string  `json:"successUrl"`
	CancelURL  string  `json:"cancelUrl"`
	OrderID    *int64  `json:"orderId,omitempty"`
}

// GET /v1/payment/fee-preview?amount=X&currency=USD&processor=stripe
// GET /v1/payment/transactions?page=0&size=20
// GET /v1/payment/session/{sessionId}
// GET /v1/public/catalog/domains?extension=com

type PayInvoiceRequest struct {
	Amount float64 `json:"amount"`
}
