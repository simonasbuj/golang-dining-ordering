package customerrors

import "errors"

// UniqueConstraintError represents a unique constraint violation error.
type UniqueConstraintError struct {
	CustomError
}

// ErrUnauthorized is returned when an operation is attempted without proper authorization.
var ErrUnauthorized = errors.New("unauthorized")

// ErrParseClaims is returned when parsing jwt token claims fails.
var ErrParseClaims = errors.New("failed to parse claims")

// ErrParseToken is returned when parsing jwt token fails.
var ErrParseToken = errors.New("failed to parse JWT token")

// ErrInvalidToken is returned when jwt token is not valid.
var ErrInvalidToken = errors.New("invalid token")

// ErrUnexpectedSigninMethod is returned when provided jwt token was signed in unexpected method.
var ErrUnexpectedSigninMethod = errors.New("unexpected signing method")

// ErrInvalidTokenData is returned when required values are missing to create jwt token.
var ErrInvalidTokenData = errors.New(
	"generating token failed because one of rquired fields are invalid",
)

// ErrMissingClaims is returned when jwt token is missing required claims.
var ErrMissingClaims = errors.New("refresh token is missing required claims")
