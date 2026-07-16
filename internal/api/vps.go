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

// GetVpsOsTemplates lists the operating systems installable on one instance. Templates are resolved
// per server and their ids change on re-import, so they must be looked up fresh — never cache an id.
func (c *Client) GetVpsOsTemplates(ctx context.Context, instanceID string, includeEol bool) ([]models.VpsOsTemplate, error) {
	q := url.Values{}
	q.Set("instanceId", instanceID)
	if includeEol {
		q.Set("includeEol", "true")
	}
	var result models.VpsOsTemplateListResponse
	if err := c.Get(ctx, "/v1/hosting/vps/os-templates", q, &result); err != nil {
		return nil, err
	}
	return result.Templates, nil
}

// BuildVpsInstance installs an OS. On a server that already has one this ERASES ALL DATA — callers
// must confirm first. A 409 means a build is already running; poll the instance instead of retrying.
func (c *Client) BuildVpsInstance(ctx context.Context, id string, req models.VpsBuildRequest) (*models.VpsBuildStatus, error) {
	var result models.VpsBuildStatus
	err := c.Post(ctx, fmt.Sprintf("/v1/hosting/vps/instances/%s/build", id), req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// StoreVpsSshKey is idempotent: storing a key you already have returns the existing one.
func (c *Client) StoreVpsSshKey(ctx context.Context, req models.VpsSshKeyCreateRequest) (*models.VpsSshKey, error) {
	var result models.VpsSshKey
	err := c.Post(ctx, "/v1/hosting/vps/ssh-keys", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListVpsSshKeys(ctx context.Context) ([]models.VpsSshKey, error) {
	var result models.VpsSshKeyListResponse
	if err := c.Get(ctx, "/v1/hosting/vps/ssh-keys", nil, &result); err != nil {
		return nil, err
	}
	return result.Keys, nil
}

func (c *Client) DeleteVpsSshKey(ctx context.Context, keyID int) error {
	return c.Delete(ctx, fmt.Sprintf("/v1/hosting/vps/ssh-keys/%d", keyID), nil)
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
