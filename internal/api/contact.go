package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) ListContacts(ctx context.Context, page, size int) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if size > 0 {
		query.Set("size", strconv.Itoa(size))
	}
	err := c.Get(ctx, "/v1/contacts", query, &result)
	return result, err
}

func (c *Client) GetContact(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, fmt.Sprintf("/v1/contacts/%s", id), nil, &result)
	return result, err
}

func (c *Client) CreateContact(ctx context.Context, req models.ContactCreateRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Post(ctx, "/v1/contacts", req, &result)
	return result, err
}

func (c *Client) UpdateContact(ctx context.Context, id string, req models.ContactCreateRequest) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Put(ctx, fmt.Sprintf("/v1/contacts/%s", id), req, &result)
	return result, err
}

func (c *Client) DeleteContact(ctx context.Context, id string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Delete(ctx, fmt.Sprintf("/v1/contacts/%s", id), &result)
	return result, err
}

func (c *Client) GetDomainContacts(ctx context.Context, domain string) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, fmt.Sprintf("/v1/contacts/for-domain/%s/all", domain), nil, &result)
	return result, err
}
