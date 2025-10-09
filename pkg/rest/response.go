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

package rest

import (
	"encoding/json"
	"net/http"
)

// Response provides utility functions for writing HTTP responses
type Response struct {
	writer http.ResponseWriter
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewResponse creates a new Response wrapper
func NewResponse(w http.ResponseWriter) *Response {
	return &Response{writer: w}
}

// JSON writes a JSON response with the given status code
func (r *Response) JSON(status int, data any) error {
	r.writer.Header().Set("Content-Type", "application/json")
	r.writer.WriteHeader(status)
	return json.NewEncoder(r.writer).Encode(data)
}

// Success writes a successful JSON response with custom status code
func (r *Response) Success(status int, data any, message ...string) error {
	response := SuccessResponse{Data: data}
	if len(message) > 0 {
		response.Message = message[0]
	}
	return r.JSON(status, response)
}

// Error writes an error JSON response with status code, code and message
func (r *Response) Error(status int, code string, message string) error {
	response := ErrorResponse{
		Code:    code,
		Message: message,
	}
	return r.JSON(status, response)
}

// Text writes a plain text response with custom status code
func (r *Response) Text(status int, text string) error {
	r.writer.Header().Set("Content-Type", "text/plain")
	r.writer.WriteHeader(status)
	_, err := r.writer.Write([]byte(text))
	return err
}

// Redirect sends a redirect response
func (r *Response) Redirect(url string, status int) {
	http.Redirect(r.writer, nil, url, status)
}

// Helper functions for common response patterns

// WriteJSON is a standalone function for writing JSON responses
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	return NewResponse(w).JSON(status, data)
}

// WriteError is a standalone function for writing error responses
func WriteError(w http.ResponseWriter, status int, code string, message string) error {
	return NewResponse(w).Error(status, code, message)
}

// WriteSuccess is a standalone function for writing success responses
func WriteSuccess(w http.ResponseWriter, status int, data any, message ...string) error {
	return NewResponse(w).Success(status, data, message...)
}
