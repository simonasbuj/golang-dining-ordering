package services

import "errors"

// ErrUserIsNotManager is returned when a user attempts an action that requires restaurant manager privileges.
var ErrUserIsNotManager = errors.New("user is not a manager")
