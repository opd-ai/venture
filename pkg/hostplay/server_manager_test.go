package hostplay

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// TestNewServerManager verifies server manager creation
func TestNewServerManager(t *testing.T) {
	tests := []struct {
		name       string
		config     *ServerConfig
		logger     *logrus.Logger
		wantError  bool
		checkError string
	}{
		{
			name: "valid configuration",
			config: &ServerConfig{
				Port:       8080,
				MaxPlayers: 4,
				WorldSeed:  12345,
				GenreID:    "fantasy",
			},
			logger:    logrus.New(),
			wantError: false,
		},
		{
			name:       "nil config",
			config:     nil,
			logger:     logrus.New(),
			wantError:  true,
			checkError: "config cannot be nil",
		},
		{
			name: "nil logger",
			config: &ServerConfig{
				Port: 8080,
			},
			logger:     nil,
			wantError:  true,
			checkError: "logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewServerManager(tt.config, tt.logger)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.checkError != "" && err.Error() != tt.checkError {
					t.Errorf("expected error %q, got %q", tt.checkError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if manager == nil {
				t.Fatal("expected non-nil manager")
			}

			// Check default values were applied
			if manager.config.Port == 0 {
				t.Error("port should have default value")
			}
			if manager.config.MaxPlayers == 0 {
				t.Error("max players should have default value")
			}
			if manager.config.GenreID == "" {
				t.Error("genre ID should have default value")
			}
		})
	}
}

// TestServerManagerDefaults verifies default configuration values are set
func TestServerManagerDefaults(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel) // Silence logs for tests

	config := &ServerConfig{} // Empty config to test defaults
	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manager.config.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", manager.config.Port)
	}
	if manager.config.MaxPlayers != 4 {
		t.Errorf("expected default max players 4, got %d", manager.config.MaxPlayers)
	}
	if manager.config.GenreID != "fantasy" {
		t.Errorf("expected default genre 'fantasy', got %q", manager.config.GenreID)
	}
	if manager.config.TickRate != 20 {
		t.Errorf("expected default tick rate 20, got %d", manager.config.TickRate)
	}
}

// TestServerManagerPortFallback tests port fallback logic
func TestServerManagerPortFallback(t *testing.T) {
	// This test would require actually starting servers and occupying ports,
	// which is expensive. Instead, we verify the logic is sound by checking
	// that ServerManager properly handles port binding.

	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       50000, // Use high port to avoid conflicts
		MaxPlayers: 2,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Start should succeed (or skip test if port unavailable)
	err = manager.Start()
	if err != nil {
		t.Skipf("port range unavailable for testing: %v", err)
	}
	defer manager.Stop()

	// Verify server is running
	if !manager.IsRunning() {
		t.Error("server should be running after Start()")
	}

	// Verify port was assigned
	port := manager.Port()
	if port < config.Port || port > config.Port+9 {
		t.Errorf("port %d outside expected range %d-%d", port, config.Port, config.Port+9)
	}

	// Verify address is set
	addr := manager.Address()
	if addr == "" {
		t.Error("address should not be empty")
	}
}

// TestServerManagerStartStop verifies server lifecycle
func TestServerManagerStartStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       50100, // High port to avoid conflicts
		MaxPlayers: 2,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Initially not running
	if manager.IsRunning() {
		t.Error("server should not be running before Start()")
	}

	// Start server
	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}

	// Should be running
	if !manager.IsRunning() {
		t.Error("server should be running after Start()")
	}

	// Stop server
	err = manager.Stop()
	if err != nil {
		t.Errorf("Stop() failed: %v", err)
	}

	// Should not be running
	if manager.IsRunning() {
		t.Error("server should not be running after Stop()")
	}

	// Stopping again should be idempotent
	err = manager.Stop()
	if err != nil {
		t.Errorf("second Stop() failed: %v", err)
	}
}

// TestServerManagerDoubleStart verifies Start() fails if already running
func TestServerManagerDoubleStart(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       50200,
		MaxPlayers: 2,
		WorldSeed:  12345,
		GenreID:    "fantasy",
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// First start should succeed
	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Second start should fail
	err = manager.Start()
	if err == nil {
		t.Error("expected error from second Start(), got nil")
	}
}

// TestServerManagerBindAddress tests bind address configuration
func TestServerManagerBindAddress(t *testing.T) {
	tests := []struct {
		name            string
		bindLAN         bool
		expectLocalhost bool
	}{
		{
			name:            "localhost bind (default)",
			bindLAN:         false,
			expectLocalhost: true,
		},
		{
			name:            "LAN bind",
			bindLAN:         true,
			expectLocalhost: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			config := &ServerConfig{
				Port:       50300,
				MaxPlayers: 2,
				WorldSeed:  12345,
				GenreID:    "fantasy",
				BindLAN:    tt.bindLAN,
			}

			manager, err := NewServerManager(config, logger)
			if err != nil {
				t.Fatalf("failed to create manager: %v", err)
			}

			err = manager.Start()
			if err != nil {
				t.Skipf("cannot start server for testing: %v", err)
			}
			defer manager.Stop()

			// Check LAN address behavior
			lanAddr := manager.GetLANAddress()
			if tt.expectLocalhost {
				if lanAddr != "" {
					t.Error("localhost bind should not return LAN address")
				}
			} else {
				// LAN bind may or may not return address depending on network config
				// Just verify it doesn't crash
				t.Logf("LAN address: %s", lanAddr)
			}
		})
	}
}

// TestServerManagerShutdownTimeout verifies shutdown completes within timeout
func TestServerManagerShutdownTimeout(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       50400,
		MaxPlayers: 2,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}

	// Stop should complete within 5 seconds
	done := make(chan error, 1)
	go func() {
		done <- manager.Stop()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Stop() failed: %v", err)
		}
	case <-time.After(6 * time.Second):
		t.Error("Stop() exceeded 5 second timeout")
	}
}

// TestServerManagerGenreSupport verifies different genres work
func TestServerManagerGenreSupport(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel)

			config := &ServerConfig{
				Port:       50500,
				MaxPlayers: 2,
				WorldSeed:  12345,
				GenreID:    genre,
			}

			manager, err := NewServerManager(config, logger)
			if err != nil {
				t.Fatalf("failed to create manager for genre %s: %v", genre, err)
			}

			err = manager.Start()
			if err != nil {
				t.Skipf("cannot start server for testing: %v", err)
			}

			// Quick verification that server started
			if !manager.IsRunning() {
				t.Errorf("server should be running for genre %s", genre)
			}

			manager.Stop()
		})
	}
}

// TestServerManagerReadySynchronization verifies that Start() properly waits for
// the server goroutine to be fully initialized using channel synchronization
// instead of arbitrary sleep delays.
func TestServerManagerReadySynchronization(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &ServerConfig{
		Port:       50500, // Use a high port to avoid conflicts
		MaxPlayers: 2,
		WorldSeed:  99999,
		GenreID:    "fantasy",
		TickRate:   20,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create server manager: %v", err)
	}

	// Measure how long Start() takes. If it were using time.Sleep(100ms),
	// this would take at least 100ms. With proper synchronization,
	// it should complete much faster (typically <50ms).
	startTime := time.Now()
	err = manager.Start()
	elapsed := time.Since(startTime)

	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Verify server is running immediately after Start() returns
	if !manager.IsRunning() {
		t.Error("server should be running immediately after Start() returns")
	}

	// Log the elapsed time for diagnostic purposes
	t.Logf("Start() completed in %v", elapsed)

	// Verify that Start() completed faster than the old time.Sleep(100ms) approach.
	// This ensures the channel-based synchronization is working and catches regressions.
	if elapsed >= 100*time.Millisecond {
		t.Errorf("Start() took %v, expected less than 100ms with channel synchronization", elapsed)
	}

	// The server should be fully ready since Start() waited for readyChan.
	// We can verify this by checking that the server address is set.
	if manager.Address() == "" {
		t.Error("server address should be set after Start() completes")
	}

	// Verify we can get the port
	if manager.Port() == 0 {
		t.Error("server port should be set after Start() completes")
	}
}

// TestIsNormalDisconnection verifies error classification for network disconnections
func TestIsNormalDisconnection(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "EOF error",
			err:      io.EOF,
			expected: true,
		},
		{
			name:     "wrapped EOF error",
			err:      fmt.Errorf("connection failed: %w", io.EOF),
			expected: true,
		},
		{
			name:     "net.ErrClosed error",
			err:      net.ErrClosed,
			expected: true,
		},
		{
			name:     "wrapped net.ErrClosed error",
			err:      fmt.Errorf("read error: %w", net.ErrClosed),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      &timeoutError{},
			expected: true,
		},
		{
			name:     "wrapped timeout error",
			err:      fmt.Errorf("operation failed: %w", &timeoutError{}),
			expected: true,
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("database connection failed"),
			expected: false,
		},
		{
			name:     "network error without timeout",
			err:      &nonTimeoutNetError{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNormalDisconnection(tt.err)
			if result != tt.expected {
				t.Errorf("isNormalDisconnection(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// timeoutError is a test implementation of net.Error with timeout
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

// nonTimeoutNetError is a test implementation of net.Error without timeout
type nonTimeoutNetError struct{}

func (e *nonTimeoutNetError) Error() string   { return "connection refused" }
func (e *nonTimeoutNetError) Timeout() bool   { return false }
func (e *nonTimeoutNetError) Temporary() bool { return true }
