package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/osir/cli/internal/api"
	"github.com/osir/cli/internal/api/models"
	"github.com/osir/cli/internal/auth"
	"github.com/osir/cli/internal/output"
)

// mockBackend implements api.Backend for testing.
type mockBackend struct {
	api.Backend // embed to satisfy interface; panics on unimplemented methods

	checkAvailability func(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error)
	listDomains       func(ctx context.Context, page, size int, sortBy, sortDir string) (*models.DomainListResponse, error)
	getDomainInfo     func(ctx context.Context, domain string) (*models.DomainInfoResponse, error)
	validateDomain    func(domain string) *models.ValidationResult
	listDnsRecords    func(ctx context.Context, domain string) ([]models.DnsRecord, error)
	checkZoneExists   func(ctx context.Context, domain string) (*models.ZoneExistsResponse, error)
	getProfile        func(ctx context.Context) (*models.UserProfile, error)
}

func (m *mockBackend) CheckDomainAvailability(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error) {
	return m.checkAvailability(ctx, domain)
}

func (m *mockBackend) ListDomains(ctx context.Context, page, size int, sortBy, sortDir string) (*models.DomainListResponse, error) {
	return m.listDomains(ctx, page, size, sortBy, sortDir)
}

func (m *mockBackend) GetDomainInfo(ctx context.Context, domain string) (*models.DomainInfoResponse, error) {
	return m.getDomainInfo(ctx, domain)
}

func (m *mockBackend) ValidateDomainName(domain string) *models.ValidationResult {
	if m.validateDomain != nil {
		return m.validateDomain(domain)
	}
	return &models.ValidationResult{Valid: true, Message: "valid"}
}

func (m *mockBackend) ListDnsRecords(ctx context.Context, domain string) ([]models.DnsRecord, error) {
	return m.listDnsRecords(ctx, domain)
}

func (m *mockBackend) CheckZoneExists(ctx context.Context, domain string) (*models.ZoneExistsResponse, error) {
	return m.checkZoneExists(ctx, domain)
}

func (m *mockBackend) GetProfile(ctx context.Context) (*models.UserProfile, error) {
	return m.getProfile(ctx)
}

// mockSession implements auth.SessionManager for testing.
type mockSession struct {
	authenticated bool
	cred          *auth.StoredCredential
}

func (m *mockSession) GetToken(ctx context.Context) (string, error) { return "test-token", nil }
func (m *mockSession) LoginWithPassword(ctx context.Context, username, password string) error {
	return nil
}
func (m *mockSession) StartDeviceLogin(ctx context.Context) (*auth.DeviceCodeResponse, error) {
	return nil, nil
}
func (m *mockSession) PollDeviceToken(ctx context.Context, deviceCode string, interval int, expiresIn int) error {
	return nil
}
func (m *mockSession) Logout(ctx context.Context) error          { return nil }
func (m *mockSession) IsAuthenticated() bool                     { return m.authenticated }
func (m *mockSession) GetCredential() *auth.StoredCredential     { return m.cred }
func (m *mockSession) Restore(cred *auth.StoredCredential)       { m.cred = cred }

// executeCommand runs a command with a mock App and returns stdout.
func executeCommand(app *App, args ...string) (string, error) {
	var buf bytes.Buffer
	app.Output.SetOut(&buf)
	app.Output.SetErr(&buf)

	root := NewRootCmd(app)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return buf.String(), err
}

func newTestApp(backend api.Backend) *App {
	return &App{
		Session: &mockSession{authenticated: true},
		Client:  backend,
		Output:  output.New(false),
	}
}

// --- Tests ---

func TestDomainCheck_Available(t *testing.T) {
	mock := &mockBackend{
		checkAvailability: func(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error) {
			return &models.DomainAvailabilityResponse{
				Domain:    domain,
				Available: true,
				Price:     12.99,
				Currency:  "USD",
			}, nil
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "domain", "check", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "example.com") {
		t.Errorf("output should contain domain name, got: %s", out)
	}
	if !strings.Contains(out, "Available") {
		t.Errorf("output should show availability, got: %s", out)
	}
	if !strings.Contains(out, "12.99") {
		t.Errorf("output should show price, got: %s", out)
	}
}

func TestDomainCheck_NotAvailable(t *testing.T) {
	mock := &mockBackend{
		checkAvailability: func(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error) {
			return &models.DomainAvailabilityResponse{
				Domain:    domain,
				Available: false,
				Status:    "registered",
			}, nil
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "domain", "check", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Not available") {
		t.Errorf("output should show not available, got: %s", out)
	}
}

func TestDomainCheck_InvalidDomain(t *testing.T) {
	mock := &mockBackend{
		validateDomain: func(domain string) *models.ValidationResult {
			return &models.ValidationResult{Valid: false, Message: "invalid domain"}
		},
	}
	app := newTestApp(mock)
	_, err := executeCommand(app, "domain", "check", "not-a-domain")
	if err == nil {
		t.Fatal("expected error for invalid domain")
	}
}

func TestDomainCheck_JSON(t *testing.T) {
	mock := &mockBackend{
		checkAvailability: func(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error) {
			return &models.DomainAvailabilityResponse{
				Domain:    domain,
				Available: true,
			}, nil
		},
	}
	app := newTestApp(mock)
	app.Output.SetJSON(true)
	out, err := executeCommand(app, "-o", "json", "domain", "check", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
		t.Errorf("output should be valid JSON, got: %s", out)
	}
}

func TestDomainValidate_Valid(t *testing.T) {
	mock := &mockBackend{
		validateDomain: func(domain string) *models.ValidationResult {
			return &models.ValidationResult{Valid: true, Message: "'example.com' is valid"}
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "domain", "validate", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "valid") {
		t.Errorf("output should confirm validity, got: %s", out)
	}
}

func TestDomainValidate_Invalid(t *testing.T) {
	mock := &mockBackend{
		validateDomain: func(domain string) *models.ValidationResult {
			return &models.ValidationResult{Valid: false, Message: "invalid format"}
		},
	}
	app := newTestApp(mock)
	out, _ := executeCommand(app, "domain", "validate", "bad!")
	if !strings.Contains(out, "invalid") {
		t.Errorf("output should show invalid, got: %s", out)
	}
}

func TestDomainList_Empty(t *testing.T) {
	mock := &mockBackend{
		listDomains: func(ctx context.Context, page, size int, sortBy, sortDir string) (*models.DomainListResponse, error) {
			return &models.DomainListResponse{
				Success: true,
				Data: models.DomainListData{
					Domains: []models.DomainSummary{},
				},
			}, nil
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "domain", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No domains found") {
		t.Errorf("should show no domains message, got: %s", out)
	}
}

func TestDomainList_WithDomains(t *testing.T) {
	mock := &mockBackend{
		listDomains: func(ctx context.Context, page, size int, sortBy, sortDir string) (*models.DomainListResponse, error) {
			return &models.DomainListResponse{
				Success: true,
				Data: models.DomainListData{
					Domains: []models.DomainSummary{
						{Domain: "foo.com", Status: "active", AutoRenew: true},
						{Domain: "bar.net", Status: "active", Privacy: true},
					},
					TotalElements: 2,
					TotalPages:    1,
				},
			}, nil
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "domain", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "foo.com") || !strings.Contains(out, "bar.net") {
		t.Errorf("should list both domains, got: %s", out)
	}
}

func TestDnsList_WithRecords(t *testing.T) {
	mock := &mockBackend{
		checkZoneExists: func(ctx context.Context, domain string) (*models.ZoneExistsResponse, error) {
			return &models.ZoneExistsResponse{Exists: true}, nil
		},
		listDnsRecords: func(ctx context.Context, domain string) ([]models.DnsRecord, error) {
			return []models.DnsRecord{
				{ID: "rec-1", Type: "A", Name: "example.com.", Content: "192.0.2.1", TTL: 3600},
				{ID: "rec-2", Type: "CNAME", Name: "www.example.com.", Content: "example.com.", TTL: 3600},
			}, nil
		},
	}
	app := newTestApp(mock)
	out, err := executeCommand(app, "dns", "list", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "192.0.2.1") {
		t.Errorf("should show record content, got: %s", out)
	}
	if !strings.Contains(out, "CNAME") {
		t.Errorf("should show record types, got: %s", out)
	}
}

func TestDnsList_NoZone(t *testing.T) {
	mock := &mockBackend{
		checkZoneExists: func(ctx context.Context, domain string) (*models.ZoneExistsResponse, error) {
			return &models.ZoneExistsResponse{Exists: false}, nil
		},
	}
	app := newTestApp(mock)
	_, err := executeCommand(app, "dns", "list", "nozone.com")
	if err == nil {
		t.Fatal("expected error for non-existent zone")
	}
}

func TestDomainRegister_YearsValidation(t *testing.T) {
	mock := &mockBackend{}
	app := newTestApp(mock)
	_, err := executeCommand(app, "domain", "register", "example.com", "--nameservers", "ns1.test.com", "--years", "0")
	if err == nil {
		t.Fatal("expected error for years=0")
	}

	_, err = executeCommand(app, "domain", "register", "example.com", "--nameservers", "ns1.test.com", "--years", "11")
	if err == nil {
		t.Fatal("expected error for years=11")
	}
}
