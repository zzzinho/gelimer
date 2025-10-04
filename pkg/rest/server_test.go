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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_RegisterRoutes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "GET route",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST route",
			method:         "POST",
			path:           "/test",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "PUT route",
			method:         "PUT",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE route",
			method:         "DELETE",
			path:           "/test",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(nil)

			// Register routes based on method
			switch tt.method {
			case "GET":
				server.GET("/test", func(w http.ResponseWriter, r *http.Request, params Params) {
					w.WriteHeader(http.StatusOK)
				})
			case "POST":
				server.POST("/test", func(w http.ResponseWriter, r *http.Request, params Params) {
					w.WriteHeader(http.StatusCreated)
				})
			case "PUT":
				server.PUT("/test", func(w http.ResponseWriter, r *http.Request, params Params) {
					w.WriteHeader(http.StatusOK)
				})
			case "DELETE":
				server.DELETE("/test", func(w http.ResponseWriter, r *http.Request, params Params) {
					w.WriteHeader(http.StatusNoContent)
				})
			}

			// Create test request
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute request
			server.router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestParams_ByName(t *testing.T) {
	server := New(nil)

	server.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, params Params) {
		id := params.ByName("id")
		if id != "123" {
			t.Errorf("Expected id '123', got '%s'", id)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Port)
	}

	if config.ReadTimeout.Seconds() != 15 {
		t.Errorf("Expected read timeout 15s, got %v", config.ReadTimeout)
	}
}
