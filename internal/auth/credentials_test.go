package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	credentialsDir = tmpDir
	credentialsFile = filepath.Join(tmpDir, "credentials.json")

	cred := &StoredCredential{
		Username:     "testuser",
		AccessToken:  "token123",
		TokenType:    "Bearer",
		ExpiresAt:    9999999999999,
		RefreshToken: "refresh456",
		LoginMethod:  "password",
	}

	if err := SaveCredentials(cred); err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	loaded := LoadCredentials()
	if loaded == nil {
		t.Fatal("LoadCredentials returned nil")
	}
	if loaded.Username != "testuser" {
		t.Errorf("Username = %q, want %q", loaded.Username, "testuser")
	}
	if loaded.AccessToken != "token123" {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, "token123")
	}
	if loaded.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", loaded.TokenType, "Bearer")
	}
	if loaded.RefreshToken != "refresh456" {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, "refresh456")
	}
	if loaded.LoginMethod != "password" {
		t.Errorf("LoginMethod = %q, want %q", loaded.LoginMethod, "password")
	}
}

func TestLoad_NoFile_ReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()
	credentialsDir = tmpDir
	credentialsFile = filepath.Join(tmpDir, "credentials.json")

	loaded := LoadCredentials()
	if loaded != nil {
		t.Error("expected nil when no file exists")
	}
}

func TestClear_DeletesFile(t *testing.T) {
	tmpDir := t.TempDir()
	credentialsDir = tmpDir
	credentialsFile = filepath.Join(tmpDir, "credentials.json")

	cred := &StoredCredential{Username: "user", AccessToken: "tok", TokenType: "Bearer", ExpiresAt: 0, LoginMethod: "device"}
	_ = SaveCredentials(cred)

	if LoadCredentials() == nil {
		t.Fatal("expected credentials to exist before clear")
	}

	_ = ClearCredentials()

	if LoadCredentials() != nil {
		t.Error("expected nil after clear")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "dir")
	credentialsDir = subDir
	credentialsFile = filepath.Join(subDir, "credentials.json")

	cred := &StoredCredential{Username: "u", AccessToken: "t", TokenType: "B", ExpiresAt: 0, LoginMethod: "p"}
	if err := SaveCredentials(cred); err != nil {
		t.Fatalf("SaveCredentials failed: %v", err)
	}

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}

func TestLoad_CorruptFile_ReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()
	credentialsDir = tmpDir
	credentialsFile = filepath.Join(tmpDir, "credentials.json")

	_ = os.WriteFile(credentialsFile, []byte("not valid json {{{"), 0600)

	loaded := LoadCredentials()
	if loaded != nil {
		t.Error("expected nil for corrupt JSON file")
	}
}
