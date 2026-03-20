package auth

import "time"

type StoredCredential struct {
	Username     string `json:"username"`
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType"`
	ExpiresAt    int64  `json:"expiresAt"`
	RefreshToken string `json:"refreshToken,omitempty"`
	LoginMethod  string `json:"loginMethod"`
}

func (c *StoredCredential) IsExpired() bool {
	return time.Now().UnixMilli() >= c.ExpiresAt
}

func (c *StoredCredential) ExpiresInSeconds() int64 {
	remaining := c.ExpiresAt - time.Now().UnixMilli()
	if remaining < 0 {
		return 0
	}
	return remaining / 1000
}

func (c *StoredCredential) NeedsRefresh() bool {
	return c.ExpiresInSeconds() < 60
}

func (c *StoredCredential) BearerToken() string {
	return c.TokenType + " " + c.AccessToken
}
