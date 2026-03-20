package config

import "os"

type Config struct {
	BackendURL   string
	KeycloakURL  string
	KeycloakRealm string
	ClientID     string
}

func Load() *Config {
	return &Config{
		BackendURL:    envOrDefault("OSIR_BACKEND_URL", "https://be.osir.com"),
		KeycloakURL:   envOrDefault("KEYCLOAK_URL", "https://auth.osir.com"),
		KeycloakRealm: envOrDefault("KEYCLOAK_REALM", "osir"),
		ClientID:      envOrDefault("KEYCLOAK_CLIENT_ID", "osir-cli"),
	}
}

func (c *Config) KeycloakTokenURL() string {
	return c.KeycloakURL + "/realms/" + c.KeycloakRealm + "/protocol/openid-connect/token"
}

func (c *Config) KeycloakDeviceURL() string {
	return c.KeycloakURL + "/realms/" + c.KeycloakRealm + "/protocol/openid-connect/auth/device"
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
