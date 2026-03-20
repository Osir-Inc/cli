package api

import (
	"context"
	"fmt"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) ListDnsRecords(ctx context.Context, domain string) ([]models.DnsRecord, error) {
	var records []models.DnsRecord
	path := fmt.Sprintf("/dns/domains/%s/records", domain)
	if err := c.Get(ctx, path, nil, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (c *Client) GetDnsRecord(ctx context.Context, domain, recordId string) (*models.DnsRecord, error) {
	var resp models.DnsRecord
	path := fmt.Sprintf("/dns/domains/%s/records/%s", domain, recordId)
	if err := c.Get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateDnsRecord(ctx context.Context, domain string, req models.DnsRecordRequest) (*models.DnsRecord, error) {
	var resp models.DnsRecord
	path := fmt.Sprintf("/dns/domains/%s/records", domain)
	if err := c.Post(ctx, path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateDnsRecord(ctx context.Context, domain, recordId string, req models.DnsRecordUpdateRequest) (*models.DnsRecord, error) {
	var resp models.DnsRecord
	path := fmt.Sprintf("/dns/domains/%s/records/%s", domain, recordId)
	if err := c.Put(ctx, path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteDnsRecord(ctx context.Context, domain, recordId string) (*models.DnsActionResponse, error) {
	var resp models.DnsActionResponse
	path := fmt.Sprintf("/dns/domains/%s/records/%s", domain, recordId)
	if err := c.Delete(ctx, path, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Zone management

func (c *Client) CreateZoneWithOsirDefaults(ctx context.Context, domain string) (*models.ZoneResponse, error) {
	var resp models.ZoneResponse
	path := fmt.Sprintf("/dns/zones/%s/osir-defaults", domain)
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CheckZoneExists(ctx context.Context, domain string) (*models.ZoneExistsResponse, error) {
	var resp models.ZoneExistsResponse
	path := fmt.Sprintf("/dns/zones/%s/exists", domain)
	if err := c.Get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) FixSOARecord(ctx context.Context, domain string) (*models.ZoneResponse, error) {
	var resp models.ZoneResponse
	path := fmt.Sprintf("/dns/zones/%s/fix-soa", domain)
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetDnssecStatus(ctx context.Context, domain string) (*models.DnssecStatusResponse, error) {
	var resp models.DnssecStatusResponse
	path := fmt.Sprintf("/dns/zones/%s/dnssec/status", domain)
	if err := c.Get(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) EnableDnssec(ctx context.Context, domain string) (*models.DnssecEnableResponse, error) {
	var resp models.DnssecEnableResponse
	path := fmt.Sprintf("/dns/zones/%s/dnssec/enable", domain)
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DisableDnssec(ctx context.Context, domain string) (*models.DnssecDisableResponse, error) {
	var resp models.DnssecDisableResponse
	path := fmt.Sprintf("/dns/zones/%s/dnssec/disable", domain)
	if err := c.Post(ctx, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
