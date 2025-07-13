package errors

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func NewAPIError(code int, message string, details ...string) *APIError {
	err := &APIError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

var (
	ErrBadRequest          = &APIError{Code: http.StatusBadRequest, Message: "Bad request"}
	ErrUnauthorized        = &APIError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrNotFound            = &APIError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrInternalServerError = &APIError{Code: http.StatusInternalServerError, Message: "Internal server error"}
	ErrDatabaseError       = &APIError{Code: http.StatusInternalServerError, Message: "Database error"}
	ErrValidationError     = &APIError{Code: http.StatusBadRequest, Message: "Validation error"}
)

func BadRequest(details string) *APIError {
	return NewAPIError(http.StatusBadRequest, "Bad request", details)
}

func NotFound(details string) *APIError {
	return NewAPIError(http.StatusNotFound, "Resource not found", details)
}

func InternalError(details string) *APIError {
	return NewAPIError(http.StatusInternalServerError, "Internal server error", details)
}

func DatabaseError(details string) *APIError {
	return NewAPIError(http.StatusInternalServerError, "Database error", details)
}

func ValidationError(details string) *APIError {
	return NewAPIError(http.StatusBadRequest, "Validation error", details)
}