package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.[a-zA-Z]{2,}$`)

func (c *Client) CheckDomainAvailability(ctx context.Context, domain string) (*models.DomainAvailabilityResponse, error) {
	var result models.DomainAvailabilityResponse
	err := c.Get(ctx, fmt.Sprintf("/v2/domains/%s/available", domain), nil, &result)
	return &result, err
}

func (c *Client) RegisterDomain(ctx context.Context, req *models.DomainRegistrationRequest) (*models.DomainRegistrationResponse, error) {
	var result models.DomainRegistrationResponse
	err := c.Post(ctx, "/v2/domains/register", req, &result)
	return &result, err
}

func (c *Client) GetDomainInfo(ctx context.Context, domain string) (*models.DomainInfoResponse, error) {
	var result models.DomainInfoResponse
	err := c.Get(ctx, fmt.Sprintf("/v2/domains/%s/info", domain), nil, &result)
	return &result, err
}

func (c *Client) ListDomains(ctx context.Context, page, size int, sortBy, sortDirection string) (*models.DomainListResponse, error) {
	var result models.DomainListResponse
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	if sortBy != "" {
		query.Set("sortBy", sortBy)
	}
	if sortDirection != "" {
		query.Set("sortDirection", sortDirection)
	}
	err := c.Get(ctx, "/v2/domains", query, &result)
	return &result, err
}

func (c *Client) RenewDomain(ctx context.Context, domain string, req *models.DomainRenewalRequest) (*models.DomainRenewalResponse, error) {
	var result models.DomainRenewalResponse
	err := c.Post(ctx, fmt.Sprintf("/v2/domains/%s/renew", domain), req, &result)
	return &result, err
}

func (c *Client) LockDomain(ctx context.Context, domain string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, fmt.Sprintf("/v2/domains/%s/lock", domain), nil, &result)
	return result, err
}

func (c *Client) UnlockDomain(ctx context.Context, domain string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, fmt.Sprintf("/v2/domains/%s/unlock", domain), nil, &result)
	return result, err
}

func (c *Client) SetAutoRenew(ctx context.Context, domain string, enable bool) (json.RawMessage, error) {
	var result json.RawMessage
	action := "disable"
	if enable {
		action = "enable"
	}
	err := c.Post(ctx, fmt.Sprintf("/v2/domains/%s/autorenew/%s", domain, action), nil, &result)
	return result, err
}

func (c *Client) SetPrivacy(ctx context.Context, domain string, enable bool) (json.RawMessage, error) {
	var result json.RawMessage
	action := "disable"
	if enable {
		action = "enable"
	}
	err := c.Post(ctx, fmt.Sprintf("/v2/domains/%s/privacy/%s", domain, action), nil, &result)
	return result, err
}

func (c *Client) UpdateNameservers(ctx context.Context, domain string, req *models.NameserverUpdateRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Put(ctx, fmt.Sprintf("/v2/domains/%s/nameservers", domain), req, &result)
	return result, err
}

func (c *Client) SuggestAlternatives(ctx context.Context, keyword string, limit int, tlds string) (*models.DomainSuggestionsResponse, error) {
	var result models.DomainSuggestionsResponse
	query := url.Values{}
	query.Set("keyword", keyword)
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	err := c.Get(ctx, "/v2/domains/suggestions", query, &result)
	return &result, err
}

func (c *Client) ValidateDomainName(domain string) *models.ValidationResult {
	if domain == "" {
		return &models.ValidationResult{
			Valid:   false,
			Message: "Domain name cannot be empty",
		}
	}
	if !domainRegex.MatchString(domain) {
		return &models.ValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("'%s' is not a valid domain name format", domain),
		}
	}
	return &models.ValidationResult{
		Valid:   true,
		Message: fmt.Sprintf("'%s' is a valid domain name", domain),
	}
}
