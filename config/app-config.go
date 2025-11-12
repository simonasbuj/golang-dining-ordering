// Package config holds the application's configuration settings.
package config

// AppConfig defines environment-based configuration for the application.
type AppConfig struct {
	DineAuthDBURI                string `env:"DINE_AUTH_DB_URI"`
	DineManagementDBURI          string `env:"DINE_MANAGEMENT_DB_URI"`
	DineHTTPAddress              string `env:"DINE_HTTP_ADDRESS"`
	DineAuthSecret               string `env:"DINE_AUTH_SECRET"`
	DineTokenValidSeconds        int    `env:"DINE_TOKEN_VALID_SECONDS"`
	DineRefreshTokenValidSeconds int    `env:"DINE_REFRESH_TOKEN_VALID_SECONDS"`
	DineAuthorizeEndpoint        string `env:"DINE_AUTHORIZE_ENDPOINT"`
}
