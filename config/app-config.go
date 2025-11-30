// Package config holds the application's configuration settings.
package config

// AppConfig defines environment-based configuration for the application.
type AppConfig struct {
	AuthDBURI                string `env:"DINE_AUTH_DB_URI"`
	ManagementDBURI          string `env:"DINE_MANAGEMENT_DB_URI"`
	HTTPAddress              string `env:"DINE_HTTP_ADDRESS"`
	AuthSecret               string `env:"DINE_AUTH_SECRET"`
	TokenValidSeconds        int    `env:"DINE_TOKEN_VALID_SECONDS"`
	RefreshTokenValidSeconds int    `env:"DINE_REFRESH_TOKEN_VALID_SECONDS"`
	AuthorizeEndpoint        string `env:"DINE_AUTHORIZE_ENDPOINT"`
	MaxImageSizeBytes        int64  `env:"DINE_MAX_IMAGE_SIZE_BYTES"`
	UploadsDirectory         string `env:"DINE_UPLOADS_DIRECTORY"`
	StripeSecretKey          string `env:"STRIPE_SECRET_KEY"`
}
