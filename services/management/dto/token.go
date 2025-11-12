package dto

// TokenClaimsDto represents the claims extracted from a JWT token.
type TokenClaimsDto struct {
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	Role         int    `json:"role"`
	TokenVersion int64  `json:"tokenVersion"`
}
