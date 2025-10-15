package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AppServer represents the main application server
type AppServer struct {
	router     *gin.Engine
	httpServer *http.Server
}

// NewAppServer creates a new application server instance
func NewAppServer() *AppServer {
	return &AppServer{}
}

// Start starts the HTTP server
func (s *AppServer) Start(addr string) error {
	s.router = setupRoutes(s)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting HTTP server on %s", addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("Server startup failed: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("Server shutdown failed: %v", err)
		return err
	}

	logrus.Info("Server stopped")
	return nil
}
