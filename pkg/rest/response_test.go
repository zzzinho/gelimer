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
	"net/http/httptest"
	"testing"
)

func TestResponse_JSON(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		data     interface{}
		expected string
	}{
		{
			name:     "success response",
			status:   200,
			data:     map[string]string{"message": "success"},
			expected: `{"message":"success"}`,
		},
		{
			name:     "created response",
			status:   201,
			data:     map[string]int{"id": 123},
			expected: `{"id":123}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			resp := NewResponse(w)

			err := resp.JSON(tt.status, tt.data)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if w.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}

			body := w.Body.String()
			// Remove trailing newline from JSON encoder
			if len(body) > 0 && body[len(body)-1] == '\n' {
				body = body[:len(body)-1]
			}

			if body != tt.expected {
				t.Errorf("Expected body '%s', got '%s'", tt.expected, body)
			}
		})
	}
}

func TestResponse_Success(t *testing.T) {
	w := httptest.NewRecorder()
	resp := NewResponse(w)

	data := map[string]string{"key": "value"}
	err := resp.Success(200, data, "Test message")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result SuccessResponse
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", result.Message)
	}
}

func TestResponse_Text(t *testing.T) {
	w := httptest.NewRecorder()
	resp := NewResponse(w)

	text := "Hello, World!"
	err := resp.Text(200, text)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", contentType)
	}

	body := w.Body.String()
	if body != text {
		t.Errorf("Expected body '%s', got '%s'", text, body)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"test": "data"}

	err := WriteJSON(w, 200, data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}
