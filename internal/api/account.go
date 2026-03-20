package api

import (
	"context"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) GetProfile(ctx context.Context) (*models.UserProfile, error) {
	var result models.UserProfile
	err := c.Get(ctx, "/v1/customers/me", nil, &result)
	return &result, err
}

func (c *Client) GetAccountSummary(ctx context.Context) (*models.AccountSummary, error) {
	var result models.AccountSummary
	err := c.Get(ctx, "/v1/customers/summary", nil, &result)
	return &result, err
}
