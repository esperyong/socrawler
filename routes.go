package main

import (
	"github.com/gin-gonic/gin"
)

// setupRoutes configures all routes for the application
func setupRoutes(appServer *AppServer) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add middleware
	router.Use(errorHandlingMiddleware())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", healthHandler)

	// MCP endpoint - using Streamable HTTP protocol
	mcpHandler := appServer.StreamableHTTPHandler()
	router.Any("/mcp", gin.WrapH(mcpHandler))
	router.Any("/mcp/*path", gin.WrapH(mcpHandler))

	return router
}

// healthHandler handles health check requests
func healthHandler(c *gin.Context) {
	respondSuccess(c, map[string]any{
		"status":    "healthy",
		"service":   "gomcp-scaffold",
		"version":   "1.0.0",
		"timestamp": "now",
	}, "Service is healthy")
}
