/*
Copyright 2025 zzzinho.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apierr

import (
	"fmt"
	"net/http"
)

// ErrorCode represents predefined API error codes
type ErrorCode string

// Predefined error codes
const (
	ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
	ErrUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrForbidden         ErrorCode = "FORBIDDEN"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrConflict          ErrorCode = "CONFLICT"
	ErrInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrBadGateway        ErrorCode = "BAD_GATEWAY"
	ErrTimeout           ErrorCode = "TIMEOUT"
	ErrTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"
)

// APIError represents a structured API error
type APIError struct {
	code       ErrorCode `json:"code"`
	message    string    `json:"message"`
	statusCode int       `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// StatusCode returns the HTTP status code for this error
func (e *APIError) StatusCode() int {
	return e.statusCode
}

// Code returns the error code
func (e *APIError) Code() string {
	return string(e.code)
}

// Message returns the error message
func (e *APIError) Message() string {
	return e.message
}


// Predefined API errors - users cannot create new ones
var (
	InvalidRequest    = &APIError{ErrInvalidRequest, "Invalid request", http.StatusBadRequest}
	Unauthorized      = &APIError{ErrUnauthorized, "Unauthorized access", http.StatusUnauthorized}
	Forbidden         = &APIError{ErrForbidden, "Access forbidden", http.StatusForbidden}
	NotFound          = &APIError{ErrNotFound, "Resource not found", http.StatusNotFound}
	Conflict          = &APIError{ErrConflict, "Resource conflict", http.StatusConflict}
	InternalError     = &APIError{ErrInternalError, "Internal server error", http.StatusInternalServerError}
	ServiceUnavailable = &APIError{ErrServiceUnavailable, "Service unavailable", http.StatusServiceUnavailable}
	BadGateway        = &APIError{ErrBadGateway, "Bad gateway", http.StatusBadGateway}
	Timeout           = &APIError{ErrTimeout, "Request timeout", http.StatusRequestTimeout}
	TooManyRequests   = &APIError{ErrTooManyRequests, "Too many requests", http.StatusTooManyRequests}
)

// WithMessage returns a copy of the error with a custom message
func (e *APIError) WithMessage(message string) *APIError {
	return &APIError{
		code:       e.code,
		message:    message,
		statusCode: e.statusCode,
	}
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) (*APIError, bool) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	return nil, false
}