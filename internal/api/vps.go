package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) ListVpsCatalog(ctx context.Context, activeOnly bool, locationId string) (*models.VpsCatalogResponse, error) {
	var result models.VpsCatalogResponse
	query := url.Values{}
	if activeOnly {
		query.Set("activeOnly", "true")
	}
	if locationId != "" {
		query.Set("locationId", locationId)
	}
	err := c.Get(ctx, "/v1/public/catalog/vps", query, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListVpsCatalogLocations(ctx context.Context) (*models.VpsLocationListResponse, error) {
	var result models.VpsLocationListResponse
	err := c.Get(ctx, "/v1/public/catalog/vps/locations", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListVpsPackages(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, "/v1/hosting/vps/packages", nil, &result)
	return result, err
}

func (c *Client) ListVpsLocations(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, "/v1/hosting/vps/locations", nil, &result)
	return result, err
}

func (c *Client) ListVpsInstances(ctx context.Context, status string) ([]models.VpsInstance, error) {
	var result []models.VpsInstance
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}
	err := c.Get(ctx, "/v1/hosting/vps/instances", query, &result)
	return result, err
}

func (c *Client) ListActiveVpsInstances(ctx context.Context) ([]models.VpsInstance, error) {
	var result []models.VpsInstance
	err := c.Get(ctx, "/v1/hosting/vps/instances/active", nil, &result)
	return result, err
}

func (c *Client) CountVpsInstances(ctx context.Context, activeOnly bool) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if activeOnly {
		query.Set("activeOnly", strconv.FormatBool(activeOnly))
	}
	err := c.Get(ctx, "/v1/hosting/vps/instances/count", query, &result)
	return result, err
}

func (c *Client) GetVpsInstance(ctx context.Context, id string) (*models.VpsInstance, error) {
	var result models.VpsInstance
	err := c.Get(ctx, fmt.Sprintf("/v1/hosting/vps/instances/%s", id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) OrderVps(ctx context.Context, req models.VpsOrderRequest) (*models.VpsOrderResponse, error) {
	var result models.VpsOrderResponse
	err := c.Post(ctx, "/v1/hosting/vps/order", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteVpsInstance(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, fmt.Sprintf("/v1/hosting/vps/instances/%s/delete", id), nil, &result)
	return result, err
}

func (c *Client) ChangeVpsPaymentTerm(ctx context.Context, id string, req models.VpsPaymentTermChangeRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, fmt.Sprintf("/v1/hosting/vps/instances/%s/change-payment-term", id), req, &result)
	return result, err
}

func (c *Client) GetVpsPanelLogin(ctx context.Context, id string) (*models.VpsPanelLoginResponse, error) {
	var result models.VpsPanelLoginResponse
	err := c.Post(ctx, fmt.Sprintf("/v1/hosting/vps/instances/%s/login", id), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
