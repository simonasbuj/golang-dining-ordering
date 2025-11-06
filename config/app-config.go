// Package config holds the application's configuration settings.
package config

// AppConfig defines environment-based configuration for the application.
type AppConfig struct {
	DineDBURI                  string `env:"DINE_DB_URI"`
	DineHTTPAddress            string `env:"DINE_HTTP_ADDRESS"`
	DineAuthSecret             string `env:"DINE_AUTH_SECRET"`
	DineTokenValidHours        int    `env:"DINE_TOKEN_VALID_HOURS"`
	DineRefreshTokenValidHours int    `env:"DINE_REFRESH_TOKEN_VALID_HOURS"`
}
