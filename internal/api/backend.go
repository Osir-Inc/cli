package api

import (
	"context"
	"encoding/json"

	"github.com/osir/cli/internal/api/models"
)

// Backend defines the interface for all API operations.
// Commands accept Backend instead of *Client, enabling mock-based unit tests.
type Backend interface {
	// Account
	GetProfile(ctx context.Context) (*models.UserProfile, error)
	GetAccountSummary(ctx context.Context) (*models.AccountSummary, error)

	// Audit
	GetRecentActivity(ctx context.Context, limit int) ([]models.AuditEntry, error)
	GetDomainAudit(ctx context.Context, domain string, page, size int) (*models.AuditPagedResponse, error)
	GetFailedOperations(ctx context.Context, page, size int) (*models.AuditPagedResponse, error)

	// Billing
	GetBalance(ctx context.Context) (json.RawMessage, error)
	ListInvoices(ctx context.Context, status, invoiceType string, page, size int) (json.RawMessage, error)
	GetInvoice(ctx context.Context, id string) (json.RawMessage, error)
	GetInvoiceByNumber(ctx context.Context, number string) (json.RawMessage, error)
	GetInvoiceStatistics(ctx context.Context) (json.RawMessage, error)
	PayInvoice(ctx context.Context, id string, amount float64) (json.RawMessage, error)
	CreateCheckoutSession(ctx context.Context, req models.CheckoutSessionRequest) (json.RawMessage, error)
	PreviewFees(ctx context.Context, amount float64, currency, processor string) (json.RawMessage, error)
	GetPaymentSession(ctx context.Context, sessionID string) (json.RawMessage, error)
	ListTransactions(ctx context.Context, page, size int) (json.RawMessage, error)
	GetBalanceHistory(ctx context.Context, page, size int) (json.RawMessage, error)
	GetDomainPricing(ctx context.Context, extension string) (json.RawMessage, error)

	// Catalog
	ListDomainExtensions(ctx context.Context, extension string) (json.RawMessage, error)
	ListDedicatedServers(ctx context.Context) (json.RawMessage, error)

	// Contact
	ListContacts(ctx context.Context, page, size int) (json.RawMessage, error)
	GetContact(ctx context.Context, id string) (json.RawMessage, error)
	CreateContact(ctx context.Context, req models.ContactCreateRequest) (json.RawMessage, error)
	UpdateContact(ctx context.Context, id string, req models.ContactCreateRequest) (json.RawMessage, error)
	DeleteContact(ctx context.Context, id string) (json.RawMessage, error)
	GetDomainContacts(ctx context.Context, domain string) (json.RawMessage, error)

	// DNS
	ListDnsRecords(ctx context.Context, domain string) ([]models.DnsRecord, error)
	GetDnsRecord(ctx context.Context, domain, recordId string) (*models.DnsRecord, error)
	CreateDnsRecord(ctx context.Context, domain string, req models.DnsRecordRequest) (*models.DnsRecord, error)
	UpdateDnsRecord(ctx context.Context, domain, recordId string, req models.DnsRecordUpdateRequest) (*models.DnsRecord, error)
	DeleteDnsRecord(ctx context.Context, domain, recordId string) (*models.DnsActionResponse, error)
	CreateZoneWithOsirDefaults(ctx context.Context, domain string) (*models.ZoneResponse, error)
	CheckZoneExists(ctx context.Context, domain string) (*models.ZoneExistsResponse, error)
	FixSOARecord(ctx context.Context, domain string) (*models.ZoneResponse, error)
	GetDnssecStatus(ctx context.Context, domain string) (*models.DnssecStatusResponse, error)
	EnableDnssec(ctx context.Context, domain string) (*models.DnssecEnableResponse, error)
	DisableDnssec(ctx context.Context, domain string) (*models.DnssecDisableResponse, error)

	// Domain
	CheckDomainAvailability(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error)
	RegisterDomain(ctx context.Context, req *models.DomainRegistrationRequest) (*models.DomainRegistrationResponse, error)
	GetDomainInfo(ctx context.Context, domain string) (*models.DomainInfoResponse, error)
	ListDomains(ctx context.Context, page, size int, sortBy, sortDirection string) (*models.DomainListResponse, error)
	RenewDomain(ctx context.Context, domain string, req *models.DomainRenewalRequest) (*models.DomainRenewalResponse, error)
	LockDomain(ctx context.Context, domain string) (json.RawMessage, error)
	UnlockDomain(ctx context.Context, domain string) (json.RawMessage, error)
	SetAutoRenew(ctx context.Context, domain string, enable bool) (json.RawMessage, error)
	SetPrivacy(ctx context.Context, domain string, enable bool) (json.RawMessage, error)
	UpdateNameservers(ctx context.Context, domain string, req *models.NameserverUpdateRequest) (json.RawMessage, error)
	SuggestAlternatives(ctx context.Context, keyword string, limit int, tlds string) (*models.DomainSuggestionsResponse, error)
	ValidateDomainName(domain string) *models.ValidationResult

	// Suggestions
	GenerateSuggestions(ctx context.Context, name, tlds, lang string, useNumbers bool, maxResults int) (*models.SuggestResponse, error)
	SpinWords(ctx context.Context, name string, position int, similarity float64, tlds, lang string, maxResults int) (*models.SuggestResponse, error)
	AddPrefix(ctx context.Context, name, vocabulary, tlds, lang string, maxResults int) (*models.SuggestResponse, error)
	AddSuffix(ctx context.Context, name, vocabulary, tlds, lang string, maxResults int) (*models.SuggestResponse, error)
	BulkSuggest(ctx context.Context, req models.BulkSuggestRequest) (*models.BulkSuggestResponse, error)
	CheckKeywordAvailability(ctx context.Context, keyword, registries, tlds string) (*models.BulkAvailabilityResponse, error)
	CheckKeywordSummary(ctx context.Context, keyword, registries, tlds string) (*models.BulkAvailabilityResponse, error)

	// VPS
	ListVpsCatalog(ctx context.Context, activeOnly bool, locationId string) (*models.VpsCatalogResponse, error)
	ListVpsCatalogLocations(ctx context.Context) (*models.VpsLocationListResponse, error)
	ListVpsPackages(ctx context.Context) (json.RawMessage, error)
	ListVpsLocations(ctx context.Context) (json.RawMessage, error)
	ListVpsInstances(ctx context.Context, status string) ([]models.VpsInstance, error)
	ListActiveVpsInstances(ctx context.Context) ([]models.VpsInstance, error)
	CountVpsInstances(ctx context.Context, activeOnly bool) (json.RawMessage, error)
	GetVpsInstance(ctx context.Context, id string) (*models.VpsInstance, error)
	OrderVps(ctx context.Context, req models.VpsOrderRequest) (*models.VpsOrderResponse, error)
	DeleteVpsInstance(ctx context.Context, id string) (json.RawMessage, error)
	ChangeVpsPaymentTerm(ctx context.Context, id string, req models.VpsPaymentTermChangeRequest) (json.RawMessage, error)
	GetVpsPanelLogin(ctx context.Context, id string) (*models.VpsPanelLoginResponse, error)

	// SSE
	ListenSSE(ctx context.Context, path string, events chan<- SSEEvent) error
}

// Verify Client implements Backend at compile time.
var _ Backend = (*Client)(nil)
