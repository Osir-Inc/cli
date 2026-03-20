package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/osir/cli/internal/config"
)

// AuthTokenResponse matches the OAuth token response from Keycloak / backend.
type AuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

// AuthRequest is sent to POST /api/auth for password login.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Session struct {
	mu         sync.Mutex
	cred       *StoredCredential
	cfg        *config.Config
	httpClient *http.Client
	Verbose    bool
}

func NewSession(cfg *config.Config) *Session {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := &net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, "tcp4", addr)
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &Session{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second, Transport: transport},
	}
}

func (s *Session) Restore(cred *StoredCredential) {
	s.cred = cred
}

func (s *Session) IsAuthenticated() bool {
	return s.cred != nil && s.cred.AccessToken != ""
}

func (s *Session) GetCredential() *StoredCredential {
	return s.cred
}

// GetToken returns the current access token, refreshing if needed.
func (s *Session) GetToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cred == nil {
		if s.Verbose {
			log.Printf("[AUTH] No credentials found, skipping token")
		}
		return "", nil
	}
	if s.Verbose {
		log.Printf("[AUTH] Token expires in %ds, needs refresh: %v", s.cred.ExpiresInSeconds(), s.cred.NeedsRefresh())
	}
	if s.cred.NeedsRefresh() && s.cred.RefreshToken != "" {
		if s.Verbose {
			log.Printf("[AUTH] Refreshing token via %s ...", s.cfg.KeycloakTokenURL())
		}
		refreshStart := time.Now()
		if err := s.refreshToken(ctx); err != nil {
			if s.Verbose {
				log.Printf("[AUTH] Refresh failed after %v: %v", time.Since(refreshStart), err)
			}
			// If refresh fails but token isn't fully expired, use it anyway
			if !s.cred.IsExpired() {
				return s.cred.AccessToken, nil
			}
			// Clear stale credentials so the user can log in fresh
			_ = ClearCredentials()
			s.cred = nil
			return "", fmt.Errorf("your session has expired, please log in again with: login")
		}
		if s.Verbose {
			log.Printf("[AUTH] Refresh completed in %v", time.Since(refreshStart))
		}
	}
	return s.cred.AccessToken, nil
}

// LoginWithPassword authenticates via POST /api/auth on the backend.
func (s *Session) LoginWithPassword(ctx context.Context, username, password string) error {
	reqBody := AuthRequest{Username: username, Password: password}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.BackendURL+"/api/auth",
		strings.NewReader(string(data)),
	)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp AuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse login response: %w", err)
	}

	s.cred = &StoredCredential{
		Username:     username,
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		ExpiresAt:    time.Now().UnixMilli() + tokenResp.ExpiresIn*1000,
		RefreshToken: tokenResp.RefreshToken,
		LoginMethod:  "password",
	}

	return SaveCredentials(s.cred)
}

// Logout clears the session and credential file.
func (s *Session) Logout(ctx context.Context) error {
	// Try to call backend logout if we have a token
	if s.cred != nil && s.cred.AccessToken != "" {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BackendURL+"/auth/logout", nil)
		req.Header.Set("Authorization", s.cred.BearerToken())
		s.httpClient.Do(req) // best-effort, ignore errors
	}
	s.cred = nil
	_ = ClearCredentials()
	return nil
}

func (s *Session) refreshToken(ctx context.Context) error {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {s.cred.RefreshToken},
		"client_id":     {s.cfg.ClientID},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.cfg.KeycloakTokenURL(),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return fmt.Errorf("failed to read refresh response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp AuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	s.cred.AccessToken = tokenResp.AccessToken
	s.cred.RefreshToken = tokenResp.RefreshToken
	s.cred.ExpiresAt = time.Now().UnixMilli() + tokenResp.ExpiresIn*1000
	if tokenResp.TokenType != "" {
		s.cred.TokenType = tokenResp.TokenType
	}

	return SaveCredentials(s.cred)
}
