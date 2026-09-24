package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/osir/cli/internal/config"
)

// An expired access token whose refresh fails must only log the user out when Keycloak
// rejected the refresh token; an outage (or no network) keeps the saved login.
func TestGetTokenKeepsLoginWhenRefreshFailsTransiently(t *testing.T) {
	for _, tc := range []struct {
		status   int
		wantKept bool
	}{
		{http.StatusServiceUnavailable, true},
		{http.StatusBadRequest, false},
	} {
		tmp := t.TempDir()
		credentialsDir = tmp
		credentialsFile = filepath.Join(tmp, "credentials.json")

		kc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
		}))
		s := &Session{cfg: &config.Config{KeycloakURL: kc.URL, KeycloakRealm: "osir"}, httpClient: kc.Client()}
		s.Restore(&StoredCredential{AccessToken: "old", RefreshToken: "r", ExpiresAt: 1})
		if err := SaveCredentials(s.cred); err != nil {
			t.Fatal(err)
		}

		if _, err := s.GetToken(context.Background()); err == nil {
			t.Errorf("HTTP %d: expected an error for an expired token that cannot refresh", tc.status)
		}
		_, statErr := os.Stat(credentialsFile)
		if kept := statErr == nil; kept != tc.wantKept {
			t.Errorf("HTTP %d: credentials kept = %v, want %v", tc.status, kept, tc.wantKept)
		}
		kc.Close()
	}
}
