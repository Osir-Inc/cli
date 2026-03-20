package auth

import (
	"testing"
	"time"
)

func TestIsExpired_FutureToken(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() + 3600000}
	if cred.IsExpired() {
		t.Error("token should not be expired")
	}
}

func TestIsExpired_PastToken(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() - 1000}
	if !cred.IsExpired() {
		t.Error("token should be expired")
	}
}

func TestExpiresInSeconds(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() + 120000}
	secs := cred.ExpiresInSeconds()
	if secs < 118 || secs > 121 {
		t.Errorf("ExpiresInSeconds = %d, expected ~120", secs)
	}
}

func TestExpiresInSeconds_Expired(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() - 5000}
	if cred.ExpiresInSeconds() != 0 {
		t.Error("expected 0 for expired token")
	}
}

func TestNeedsRefresh_SoonToExpire(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() + 30000} // 30s left
	if !cred.NeedsRefresh() {
		t.Error("token expiring in 30s should need refresh")
	}
}

func TestNeedsRefresh_FarFromExpiry(t *testing.T) {
	cred := &StoredCredential{ExpiresAt: time.Now().UnixMilli() + 3600000} // 1h left
	if cred.NeedsRefresh() {
		t.Error("token with 1h left should not need refresh")
	}
}

func TestBearerToken(t *testing.T) {
	cred := &StoredCredential{TokenType: "Bearer", AccessToken: "abc123"}
	expected := "Bearer abc123"
	if got := cred.BearerToken(); got != expected {
		t.Errorf("BearerToken = %q, want %q", got, expected)
	}
}
