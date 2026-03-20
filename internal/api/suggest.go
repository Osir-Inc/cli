package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/osir/cli/internal/api/models"
)

func (c *Client) GenerateSuggestions(ctx context.Context, name, tlds, lang string, useNumbers bool, maxResults int) (*models.SuggestResponse, error) {
	var result models.SuggestResponse
	query := url.Values{}
	query.Set("name", name)
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	if lang != "" {
		query.Set("lang", lang)
	}
	query.Set("use-numbers", strconv.FormatBool(useNumbers))
	if maxResults > 0 {
		query.Set("max-results", strconv.Itoa(maxResults))
	}
	err := c.Get(ctx, "/namesuggestions/suggest", query, &result)
	return &result, err
}

func (c *Client) SpinWords(ctx context.Context, name string, position int, similarity float64, tlds, lang string, maxResults int) (*models.SuggestResponse, error) {
	var result models.SuggestResponse
	query := url.Values{}
	query.Set("name", name)
	query.Set("position", strconv.Itoa(position))
	query.Set("similarity", fmt.Sprintf("%.2f", similarity))
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	if lang != "" {
		query.Set("lang", lang)
	}
	if maxResults > 0 {
		query.Set("max-results", strconv.Itoa(maxResults))
	}
	err := c.Get(ctx, "/namesuggestions/spin-word", query, &result)
	return &result, err
}

func (c *Client) AddPrefix(ctx context.Context, name, vocabulary, tlds, lang string, maxResults int) (*models.SuggestResponse, error) {
	var result models.SuggestResponse
	query := url.Values{}
	query.Set("name", name)
	if vocabulary != "" {
		query.Set("vocabulary", vocabulary)
	}
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	if lang != "" {
		query.Set("lang", lang)
	}
	if maxResults > 0 {
		query.Set("max-results", strconv.Itoa(maxResults))
	}
	err := c.Get(ctx, "/namesuggestions/add-prefix", query, &result)
	return &result, err
}

func (c *Client) AddSuffix(ctx context.Context, name, vocabulary, tlds, lang string, maxResults int) (*models.SuggestResponse, error) {
	var result models.SuggestResponse
	query := url.Values{}
	query.Set("name", name)
	if vocabulary != "" {
		query.Set("vocabulary", vocabulary)
	}
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	if lang != "" {
		query.Set("lang", lang)
	}
	if maxResults > 0 {
		query.Set("max-results", strconv.Itoa(maxResults))
	}
	err := c.Get(ctx, "/namesuggestions/add-suffix", query, &result)
	return &result, err
}

func (c *Client) BulkSuggest(ctx context.Context, req models.BulkSuggestRequest) (*models.BulkSuggestResponse, error) {
	var result models.BulkSuggestResponse
	err := c.Post(ctx, "/namesuggestions/bulk-suggest", req, &result)
	return &result, err
}

func (c *Client) CheckKeywordAvailability(ctx context.Context, keyword, registries, tlds string) (*models.BulkAvailabilityResponse, error) {
	var result models.BulkAvailabilityResponse
	query := url.Values{}
	if registries != "" {
		query.Set("registries", registries)
	}
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	path := fmt.Sprintf("/namesuggestions/keyword-availability/%s", keyword)
	err := c.Get(ctx, path, query, &result)
	return &result, err
}

func (c *Client) CheckKeywordSummary(ctx context.Context, keyword, registries, tlds string) (*models.BulkAvailabilityResponse, error) {
	var result models.BulkAvailabilityResponse
	query := url.Values{}
	if registries != "" {
		query.Set("registries", registries)
	}
	if tlds != "" {
		query.Set("tlds", tlds)
	}
	path := fmt.Sprintf("/namesuggestions/keyword-availability/%s/summary", keyword)
	err := c.Get(ctx, path, query, &result)
	return &result, err
}
