package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var credentialsDir = filepath.Join(homeDir(), ".osir")
var credentialsFile = filepath.Join(credentialsDir, "credentials.json")

func SaveCredentials(cred *StoredCredential) error {
	if err := os.MkdirAll(credentialsDir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return err
	}
	// Atomic write: write to temp file then rename to avoid corruption on interrupt
	tmp, err := os.CreateTemp(credentialsDir, ".creds.tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, credentialsFile)
}

func LoadCredentials() *StoredCredential {
	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil
	}
	var cred StoredCredential
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil
	}
	return &cred
}

func ClearCredentials() error {
	return os.Remove(credentialsFile)
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
