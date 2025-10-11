package customerrors

import "errors"

type UniqueConstraintError struct {
	CustomError
}

var UnauthorizedError = errors.New("unauthorized")
