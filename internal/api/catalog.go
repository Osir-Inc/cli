package api

import (
	"context"
	"encoding/json"
	"net/url"
)

func (c *Client) ListDomainExtensions(ctx context.Context, extension string) (json.RawMessage, error) {
	var result json.RawMessage
	query := url.Values{}
	if extension != "" {
		query.Set("extension", extension)
	}
	err := c.Get(ctx, "/v1/public/catalog/domains", query, &result)
	return result, err
}

func (c *Client) ListDedicatedServers(ctx context.Context) (json.RawMessage, error) {
	var result json.RawMessage
	err := c.Get(ctx, "/v1/public/catalog/dedicated", nil, &result)
	return result, err
}
