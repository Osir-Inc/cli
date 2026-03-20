package auth

import "context"

// SessionManager defines the interface for authentication operations.
// Commands accept SessionManager instead of *Session, enabling mock-based unit tests.
type SessionManager interface {
	GetToken(ctx context.Context) (string, error)
	LoginWithPassword(ctx context.Context, username, password string) error
	StartDeviceLogin(ctx context.Context) (*DeviceCodeResponse, error)
	PollDeviceToken(ctx context.Context, deviceCode string, interval int, expiresIn int) error
	Logout(ctx context.Context) error
	IsAuthenticated() bool
	GetCredential() *StoredCredential
	Restore(cred *StoredCredential)
}

// Verify Session implements SessionManager at compile time.
var _ SessionManager = (*Session)(nil)
