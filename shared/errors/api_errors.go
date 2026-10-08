package errors

import (
	"fmt"
)

// InvalidAccessToken returns the shared unauthorized AppError instead of a
// generic error so gateway middleware can map it to HTTP 401 consistently.
func InvalidAccessToken() error {
	return ErrUnauthorized.WithMessage("invalid access token")
}

func NewBadRequestError(message string) *AppError {
	return ErrBadRequest.WithMessage(message)
}

func NewNotFoundError(resource string) *AppError {
	return ErrNotFound.WithMessage(fmt.Sprintf("%s not found", resource))
}

func NewConflictError(message string) *AppError {
	return ErrConflict.WithMessage(message)
}

func NewInternalError(err error) *AppError {
	return ErrInternal.WithInternal(err)
}

func NewServiceUnavailableError(service string) *AppError {
	return ErrServiceUnavailable.WithMessage(fmt.Sprintf("%s is temporarily unavailable", service))
}
