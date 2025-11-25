// Package dto contains data transfer objects used for API requests/responses.
package dto

import "github.com/google/uuid"

// Role represents the type of user in the system.
type Role int

const (
	// RoleManager is the manager role.
	RoleManager Role = 1
	// RoleWaiter is the waiter role.
	RoleWaiter Role = 2
)

// SignUpRequestDto represents the payload sent when signing up.
type SignUpRequestDto struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name"     validate:"required"`
	Lastname string `json:"lastname" validate:"required"`
	Role     Role   `json:"role"     validate:"required,oneof=1 2"`
}

// SignInRequestDto represents the payload sent when signing in.
type SignInRequestDto struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// TokenResponseDto represents the payload when a new access token is issued.
type TokenResponseDto struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenRequestDto represents the payload required to refresh an authentication token.
type RefreshTokenRequestDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequestDto represents the payload required to logout user by making their tokens invalid in database.
type LogoutRequestDto struct {
	Token string `json:"token" validate:"required"`
}

// TokenClaimsDto represents the claims stored in a JWT for a user.
type TokenClaimsDto struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	TokenType string    `json:"token_type"`
	Role      Role      `json:"role"`
	Exp       int64     `json:"exp"`
}
