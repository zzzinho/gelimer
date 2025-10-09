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

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gelimer/pkg/apierr"
	"gelimer/pkg/rest"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Port to run the MCP Mart server on")
	flag.Parse()

	// Create REST server
	config := &rest.Config{
		Port:         port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server := rest.New(config)

	// Register routes
	registerRoutes(server)

	// Start server
	httpServer, err := server.StartWithShutdown()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("MCP Mart server starting on port %d", port)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down MCP Mart server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("MCP Mart server stopped")
}

func registerRoutes(server *rest.Server) {
	// Health check endpoint
	server.GET("/health", func(w http.ResponseWriter, r *http.Request, params rest.Params) {
		resp := rest.NewResponse(w)
		resp.Success(200, map[string]string{"status": "ok"}, "MCP Mart is healthy")
	})

	// API v1 routes
	api := server
	api.GET("/api/v1/mcp-servers", listMCPServers)
	api.POST("/api/v1/mcp-servers", createMCPServer)
	api.GET("/api/v1/mcp-servers/:id", getMCPServer)
	api.PUT("/api/v1/mcp-servers/:id", updateMCPServer)
	api.DELETE("/api/v1/mcp-servers/:id", deleteMCPServer)
}

func listMCPServers(w http.ResponseWriter, r *http.Request, params rest.Params) {
	resp := rest.NewResponse(w)

	// Placeholder data
	servers := []map[string]any{
		{"id": "1", "name": "example-server", "status": "running"},
		{"id": "2", "name": "test-server", "status": "stopped"},
	}

	resp.Success(200, servers, "MCP servers retrieved successfully")
}

func createMCPServer(w http.ResponseWriter, r *http.Request, params rest.Params) {
	resp := rest.NewResponse(w)

	// Placeholder implementation
	newServer := map[string]any{
		"id":     "3",
		"name":   "new-server",
		"status": "creating",
	}

	resp.Success(201, newServer, "MCP server created successfully")
}

func getMCPServer(w http.ResponseWriter, r *http.Request, params rest.Params) {
	resp := rest.NewResponse(w)

	id := params.ByName("id")
	if id == "" {
		apiErr := apierr.InvalidRequest.WithMessage("Missing server ID")
		resp.Error(apiErr.StatusCode(), apiErr.Code(), apiErr.Message())
		return
	}

	// Placeholder implementation
	if id == "999" {
		apiErr := apierr.NotFound.WithMessage("MCP server not found")
		resp.Error(apiErr.StatusCode(), apiErr.Code(), apiErr.Message())
		return
	}

	server := map[string]any{
		"id":     id,
		"name":   "example-server",
		"status": "running",
	}

	resp.Success(200, server, "MCP server retrieved successfully")
}

func updateMCPServer(w http.ResponseWriter, r *http.Request, params rest.Params) {
	resp := rest.NewResponse(w)

	id := params.ByName("id")
	if id == "" {
		apiErr := apierr.InvalidRequest.WithMessage("Missing server ID")
		resp.Error(apiErr.StatusCode(), apiErr.Code(), apiErr.Message())
		return
	}

	// Placeholder implementation
	updatedServer := map[string]any{
		"id":     id,
		"name":   "updated-server",
		"status": "running",
	}

	resp.Success(200, updatedServer, "MCP server updated successfully")
}

func deleteMCPServer(w http.ResponseWriter, r *http.Request, params rest.Params) {
	resp := rest.NewResponse(w)

	id := params.ByName("id")
	if id == "" {
		apiErr := apierr.InvalidRequest.WithMessage("Missing server ID")
		resp.Error(apiErr.StatusCode(), apiErr.Code(), apiErr.Message())
		return
	}

	// Placeholder implementation
	resp.Success(204, nil, "MCP server deleted successfully")
}
