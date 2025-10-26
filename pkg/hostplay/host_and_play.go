// Package hostplay provides host-and-play functionality for LAN party mode.
// This allows a single client to host an embedded server and automatically connect to it.
//
// Design Philosophy:
// - Simple: Single function to start server in background goroutine
// - Safe: Proper cleanup with context cancellation and timeouts
// - User-friendly: Automatic port fallback if default port is occupied
package hostplay

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Config contains configuration for host-and-play mode
type Config struct {
	// StartPort is the first port to try (default: 8080)
	StartPort int
	// PortRange is the number of ports to try (default: 10, tries 8080-8089)
	PortRange int
	// BindLAN controls whether to bind to 0.0.0.0 (LAN) or 127.0.0.1 (localhost only)
	BindLAN bool
}

// DefaultConfig returns default configuration for host-and-play mode
func DefaultConfig() Config {
	return Config{
		StartPort: 8080,
		PortRange: 10,
		BindLAN:   false,
	}
}

// Server manages the embedded server lifecycle for host-and-play mode
type Server struct {
	config Config
	ctx    context.Context
	cancel context.CancelFunc
	port   int
	addr   string
}

// New creates a new host-and-play server manager
func New(config Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// FindAvailablePort finds the first available port in the configured range
// Returns (port, bindAddress, error)
func (s *Server) FindAvailablePort() (int, string, error) {
	bindAddr := "127.0.0.1" // Default to localhost
	if s.config.BindLAN {
		bindAddr = "0.0.0.0"
	}

	// Try each port in range
	for i := 0; i < s.config.PortRange; i++ {
		port := s.config.StartPort + i
		address := fmt.Sprintf("%s:%d", bindAddr, port)

		// Try to listen on this port
		listener, err := net.Listen("tcp", address)
		if err == nil {
			// Port is available, close the test listener
			listener.Close()
			return port, bindAddr, nil
		}
	}

	return 0, "", fmt.Errorf("no available ports in range %d-%d",
		s.config.StartPort, s.config.StartPort+s.config.PortRange-1)
}

// GetContext returns the server's cancellation context
func (s *Server) GetContext() context.Context {
	return s.ctx
}

// Shutdown gracefully stops the server
// Blocks until shutdown completes or timeout (5 seconds)
func (s *Server) Shutdown() error {
	s.cancel()

	// Wait for cleanup with timeout
	select {
	case <-time.After(5 * time.Second):
		return fmt.Errorf("shutdown timeout after 5 seconds")
	case <-s.ctx.Done():
		// Additional small delay to ensure goroutine cleanup completes
		time.Sleep(100 * time.Millisecond)
		return nil
	}
}

// GetAddress returns the actual server address (e.g., "localhost:8080")
func (s *Server) GetAddress() string {
	if s.port == 0 {
		return ""
	}
	return fmt.Sprintf("localhost:%d", s.port)
}

// SetPort stores the actual port and address used
func (s *Server) SetPort(port int, addr string) {
	s.port = port
	s.addr = addr
}
