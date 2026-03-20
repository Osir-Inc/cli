package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) GetRecentActivity(ctx context.Context, limit int) ([]models.AuditEntry, error) {
	var result []models.AuditEntry
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	err := c.Get(ctx, "/v1/audit/recent", query, &result)
	return result, err
}

func (c *Client) GetDomainAudit(ctx context.Context, domain string, page, size int) (*models.AuditPagedResponse, error) {
	var result models.AuditPagedResponse
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	path := fmt.Sprintf("/v1/audit/domain/%s", domain)
	err := c.Get(ctx, path, query, &result)
	return &result, err
}

func (c *Client) GetFailedOperations(ctx context.Context, page, size int) (*models.AuditPagedResponse, error) {
	var result models.AuditPagedResponse
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	err := c.Get(ctx, "/v1/audit/failures", query, &result)
	return &result, err
}
