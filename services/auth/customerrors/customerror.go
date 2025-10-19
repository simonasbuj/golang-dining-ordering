// Package customerrors provides application-specific error types.
package customerrors

// CustomError serves as a base type for custom application errors menat to be 'inherited'.
type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}
