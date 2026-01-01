package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/soumadeep-sarkar/multi-threaded-proxy-server/config"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create proxy server instance
	server := NewProxyServer(cfg)

	// Start server
	if err := server.Start(); err != nil {
		server.errorLogger.Fatalf("Failed to start server: %v", err)
	}
}

// Start begins listening for incoming connections
func (s *ProxyServer) Start() error {
	address := fmt.Sprintf(":%s", s.config.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to create listener on %s: %w", address, err)
	}
	defer listener.Close()

	s.infoLogger.Printf("🚀 Multi-threaded Proxy Server started on port %s", s.config.Port)
	s.infoLogger.Printf("📊 Configuration: Max Clients=%d, Cache Size=%d", s.config.MaxClients, s.config.CacheSize)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		s.infoLogger.Println("🛑 Shutting down server...")
		listener.Close()
		os.Exit(0)
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if error is due to listener being closed
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.errorLogger.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each connection in a separate goroutine
		go s.handleConnection(conn)
	}
}
