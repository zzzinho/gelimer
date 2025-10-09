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
	"errors"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name:     "not found error",
			apiErr:   NotFound,
			expected: "NOT_FOUND: Resource not found",
		},
		{
			name:     "unauthorized error",
			apiErr:   Unauthorized,
			expected: "UNAUTHORIZED: Unauthorized access",
		},
		{
			name:     "custom message",
			apiErr:   NotFound.WithMessage("User not found"),
			expected: "NOT_FOUND: User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiErr.Error()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAPIError_StatusCode(t *testing.T) {
	tests := []struct {
		name           string
		apiErr         *APIError
		expectedStatus int
	}{
		{
			name:           "not found",
			apiErr:         NotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "bad request",
			apiErr:         InvalidRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal error",
			apiErr:         InternalError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiErr.StatusCode()
			if result != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, result)
			}
		})
	}
}

func TestAPIError_Code(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name:     "not found code",
			apiErr:   NotFound,
			expected: "NOT_FOUND",
		},
		{
			name:     "unauthorized code",
			apiErr:   Unauthorized,
			expected: "UNAUTHORIZED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiErr.Code()
			if result != tt.expected {
				t.Errorf("Expected code '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAPIError_Message(t *testing.T) {
	tests := []struct {
		name     string
		apiErr   *APIError
		expected string
	}{
		{
			name:     "default message",
			apiErr:   NotFound,
			expected: "Resource not found",
		},
		{
			name:     "custom message",
			apiErr:   NotFound.WithMessage("Custom not found message"),
			expected: "Custom not found message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiErr.Message()
			if result != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestAPIError_WithMessage(t *testing.T) {
	original := NotFound
	customMessage := "User with ID 123 not found"

	modified := original.WithMessage(customMessage)

	// Original should be unchanged
	if original.Message() != "Resource not found" {
		t.Errorf("Original error message changed: expected 'Resource not found', got '%s'", original.Message())
	}

	// Modified should have new message but same code and status
	if modified.Message() != customMessage {
		t.Errorf("Expected custom message '%s', got '%s'", customMessage, modified.Message())
	}

	if modified.Code() != original.Code() {
		t.Errorf("Code should remain the same: expected '%s', got '%s'", original.Code(), modified.Code())
	}

	if modified.StatusCode() != original.StatusCode() {
		t.Errorf("Status code should remain the same: expected %d, got %d", original.StatusCode(), modified.StatusCode())
	}
}

func TestIsAPIError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectAPIErr bool
		expectedErr  *APIError
	}{
		{
			name:         "valid API error",
			err:          NotFound,
			expectAPIErr: true,
			expectedErr:  NotFound,
		},
		{
			name:         "custom message API error",
			err:          NotFound.WithMessage("Custom"),
			expectAPIErr: true,
			expectedErr:  NotFound.WithMessage("Custom"),
		},
		{
			name:         "standard error",
			err:          errors.New("standard error"),
			expectAPIErr: false,
			expectedErr:  nil,
		},
		{
			name:         "nil error",
			err:          nil,
			expectAPIErr: false,
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr, isAPI := IsAPIError(tt.err)

			if isAPI != tt.expectAPIErr {
				t.Errorf("Expected IsAPIError to return %v, got %v", tt.expectAPIErr, isAPI)
			}

			if tt.expectAPIErr && apiErr.Code() != tt.expectedErr.Code() {
				t.Errorf("Expected API error code '%s', got '%s'", tt.expectedErr.Code(), apiErr.Code())
			}

			if !tt.expectAPIErr && apiErr != nil {
				t.Errorf("Expected nil API error, got %v", apiErr)
			}
		})
	}
}
