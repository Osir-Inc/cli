package api

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type SSEEvent struct {
	Event string
	Data  string
}

// ListenSSE opens an SSE connection and sends events to the channel.
// Blocks until context is cancelled or connection drops.
func (c *Client) ListenSSE(ctx context.Context, path string, events chan<- SSEEvent) error {
	fullURL := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	if c.Session != nil {
		token, err := c.Session.GetToken(ctx)
		if err != nil {
			return err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	// SSE needs a client with no timeout (the connection stays open)
	sseClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := &net.Dialer{Timeout: 10 * time.Second}
				return d.DialContext(ctx, "tcp4", addr)
			},
			IdleConnTimeout: 0,
		},
		Timeout: 0, // No timeout for SSE
	}

	resp, err := sseClient.Do(req)
	if err != nil {
		return fmt.Errorf("SSE connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE connection returned HTTP %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	var currentEvent SSEEvent

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event, dispatch it
			if currentEvent.Data != "" {
				select {
				case events <- currentEvent:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			currentEvent = SSEEvent{}
			continue
		}

		// Handle colon-prefixed fields (with or without space after colon)
		if strings.HasPrefix(line, "event:") {
			currentEvent.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if currentEvent.Data == "" {
				currentEvent.Data = data
			} else {
				currentEvent.Data += "\n" + data
			}
		} else if strings.HasPrefix(line, ":") {
			// Comment line, ignore (keep-alive pings)
			continue
		}
	}

	return scanner.Err()
}
