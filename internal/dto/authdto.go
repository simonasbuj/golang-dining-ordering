// Package dto contains data transfer objects used for API requests/responses.
package dto

// SignUpRequestDto represents the payload sent when signing up.
type SignUpRequestDto struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name"     validate:"required"`
	Lastname string `json:"lastname" validate:"required"`
	Role     string `json:"role"     validate:"required,oneof=manager waiter"`
}

// SignInRequestDto represents the payload sent when signing in.
type SignInRequestDto struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// TokenResponseDto represents the payload when a new access token is issued.
type TokenResponseDto struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenRequestDto represents the payload required to refresh an authentication token.
type RefreshTokenRequestDto struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
