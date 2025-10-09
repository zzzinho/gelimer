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
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Handler defines a function that handles HTTP requests
type Handler func(w http.ResponseWriter, r *http.Request, params Params)

// Params wraps httprouter.Params for easier access to path parameters
type Params httprouter.Params

// ByName returns the value of the named parameter
func (p Params) ByName(name string) string {
	return httprouter.Params(p).ByName(name)
}

// Server provides a REST API server with route registration
type Server struct {
	router *httprouter.Router
	config *Config
}

// Config holds server configuration
type Config struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:         8080,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// New creates a new REST server
func New(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	return &Server{
		router: httprouter.New(),
		config: config,
	}
}

// GET registers a GET route
func (s *Server) GET(path string, handler Handler) {
	s.router.GET(path, s.wrapHandler(handler))
}

// POST registers a POST route
func (s *Server) POST(path string, handler Handler) {
	s.router.POST(path, s.wrapHandler(handler))
}

// PUT registers a PUT route
func (s *Server) PUT(path string, handler Handler) {
	s.router.PUT(path, s.wrapHandler(handler))
}

// DELETE registers a DELETE route
func (s *Server) DELETE(path string, handler Handler) {
	s.router.DELETE(path, s.wrapHandler(handler))
}

// PATCH registers a PATCH route
func (s *Server) PATCH(path string, handler Handler) {
	s.router.PATCH(path, s.wrapHandler(handler))
}

// HEAD registers a HEAD route
func (s *Server) HEAD(path string, handler Handler) {
	s.router.HEAD(path, s.wrapHandler(handler))
}

// OPTIONS registers an OPTIONS route
func (s *Server) OPTIONS(path string, handler Handler) {
	s.router.OPTIONS(path, s.wrapHandler(handler))
}

// ServeFiles serves static files from a directory
func (s *Server) ServeFiles(path string, root http.FileSystem) {
	s.router.ServeFiles(path, root)
}

// wrapHandler converts our Handler to httprouter.Handle
func (s *Server) wrapHandler(handler Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		handler(w, r, Params(params))
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	return server.ListenAndServe()
}

// StartWithShutdown starts the server and returns a shutdown function
func (s *Server) StartWithShutdown() (*http.Server, error) {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("Server failed to start: %v", err))
		}
	}()

	return server, nil
}

// Router returns the underlying httprouter.Router for advanced usage
func (s *Server) Router() *httprouter.Router {
	return s.router
}
