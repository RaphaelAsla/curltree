package utils

import (
	"errors"
	"fmt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidSSHKey      = errors.New("invalid SSH key")
	ErrInvalidInput       = errors.New("invalid input")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrUnauthorized       = errors.New("unauthorized")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewAppError(code int, message string, err error) AppError {
	return AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}