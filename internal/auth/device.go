package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DeviceCodeResponse is returned when initiating device authorization.
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// TokenErrorResponse is returned on OAuth errors during polling.
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// StartDeviceLogin initiates the OAuth 2.0 Device Authorization flow.
func (s *Session) StartDeviceLogin(ctx context.Context) (*DeviceCodeResponse, error) {
	form := url.Values{
		"client_id": {s.cfg.ClientID},
		"scope":     {"openid"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.KeycloakDeviceURL(),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("device auth request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("device auth failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, fmt.Errorf("failed to parse device auth response: %w", err)
	}

	if deviceResp.Interval == 0 {
		deviceResp.Interval = 5
	}

	return &deviceResp, nil
}

// PollDeviceToken polls the token endpoint until the user completes browser auth.
func (s *Session) PollDeviceToken(ctx context.Context, deviceCode string, interval int, expiresIn int) error {
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)
	pollInterval := time.Duration(interval) * time.Second

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}

		form := url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"client_id":   {s.cfg.ClientID},
			"device_code": {deviceCode},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			s.cfg.KeycloakTokenURL(),
			strings.NewReader(form.Encode()),
		)
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 200 {
			var tokenResp AuthTokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return fmt.Errorf("failed to parse token response: %w", err)
			}

			username := extractUsernameFromJWT(tokenResp.AccessToken)

			s.cred = &StoredCredential{
				Username:     username,
				AccessToken:  tokenResp.AccessToken,
				TokenType:    tokenResp.TokenType,
				ExpiresAt:    time.Now().UnixMilli() + tokenResp.ExpiresIn*1000,
				RefreshToken: tokenResp.RefreshToken,
				LoginMethod:  "device",
			}

			return SaveCredentials(s.cred)
		}

		var tokenErr TokenErrorResponse
		if err := json.Unmarshal(body, &tokenErr); err != nil {
			continue
		}

		switch tokenErr.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			pollInterval += 5 * time.Second
		case "expired_token":
			return fmt.Errorf("device code expired - please try again")
		case "access_denied":
			return fmt.Errorf("access denied by user")
		default:
			return fmt.Errorf("device auth error: %s - %s", tokenErr.Error, tokenErr.ErrorDescription)
		}
	}

	return fmt.Errorf("device authorization timed out after %d seconds", expiresIn)
}

func extractUsernameFromJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "unknown"
	}

	// Add padding if needed
	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	data, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "unknown"
	}

	var claims map[string]any
	if err := json.Unmarshal(data, &claims); err != nil {
		return "unknown"
	}

	if username, ok := claims["preferred_username"].(string); ok {
		return username
	}
	return "unknown"
}
