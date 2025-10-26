package hostplay

import (
	"fmt"
	"net"
	"testing"
	"time"
)

// TestDefaultConfig verifies default configuration values
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.StartPort != 8080 {
		t.Errorf("expected StartPort 8080, got %d", config.StartPort)
	}
	if config.PortRange != 10 {
		t.Errorf("expected PortRange 10, got %d", config.PortRange)
	}
	if config.BindLAN != false {
		t.Errorf("expected BindLAN false, got %v", config.BindLAN)
	}
}

// TestNew verifies server creation
func TestNew(t *testing.T) {
	config := DefaultConfig()
	server := New(config)

	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.ctx == nil {
		t.Error("expected non-nil context")
	}
	if server.cancel == nil {
		t.Error("expected non-nil cancel function")
	}
}

// TestFindAvailablePort tests port discovery logic
func TestFindAvailablePort(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "default config should find port",
			config: Config{
				StartPort: 8080,
				PortRange: 10,
				BindLAN:   false,
			},
			wantError: false,
		},
		{
			name: "localhost binding",
			config: Config{
				StartPort: 9000,
				PortRange: 5,
				BindLAN:   false,
			},
			wantError: false,
		},
		{
			name: "LAN binding",
			config: Config{
				StartPort: 9100,
				PortRange: 5,
				BindLAN:   true,
			},
			wantError: false,
		},
		{
			name: "high port range",
			config: Config{
				StartPort: 50000,
				PortRange: 10,
				BindLAN:   false,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(tt.config)
			port, bindAddr, err := server.FindAvailablePort()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantError {
				if port < tt.config.StartPort || port >= tt.config.StartPort+tt.config.PortRange {
					t.Errorf("port %d out of range %d-%d",
						port, tt.config.StartPort, tt.config.StartPort+tt.config.PortRange-1)
				}

				expectedAddr := "127.0.0.1"
				if tt.config.BindLAN {
					expectedAddr = "0.0.0.0"
				}
				if bindAddr != expectedAddr {
					t.Errorf("expected bindAddr %s, got %s", expectedAddr, bindAddr)
				}
			}
		})
	}
}

// TestFindAvailablePort_PortConflict tests behavior when ports are occupied
func TestFindAvailablePort_PortConflict(t *testing.T) {
	// Occupy ports 8080-8082
	listeners := make([]net.Listener, 0, 3)
	for i := 8080; i <= 8082; i++ {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", i))
		if err != nil {
			t.Skipf("cannot occupy port %d for testing: %v", i, err)
		}
		listeners = append(listeners, listener)
		defer listener.Close()
	}

	// Try to find port in range 8080-8084 (should find 8083 or 8084)
	config := Config{
		StartPort: 8080,
		PortRange: 5,
		BindLAN:   false,
	}
	server := New(config)
	port, _, err := server.FindAvailablePort()
	if err != nil {
		t.Fatalf("expected to find available port, got error: %v", err)
	}

	// Should get port >= 8083 (first available after occupied ports)
	if port < 8083 {
		t.Errorf("expected port >= 8083, got %d", port)
	}
}

// TestFindAvailablePort_AllPortsOccupied tests error when no ports are available
func TestFindAvailablePort_AllPortsOccupied(t *testing.T) {
	// Use a narrow range and occupy all ports
	startPort := 9200
	portRange := 3

	listeners := make([]net.Listener, 0, portRange)
	for i := 0; i < portRange; i++ {
		port := startPort + i
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			t.Skipf("cannot occupy port %d for testing: %v", port, err)
		}
		listeners = append(listeners, listener)
		defer listener.Close()
	}

	// Try to find port in occupied range
	config := Config{
		StartPort: startPort,
		PortRange: portRange,
		BindLAN:   false,
	}
	server := New(config)
	_, _, err := server.FindAvailablePort()

	if err == nil {
		t.Error("expected error when all ports occupied, got nil")
	}
}

// TestGetContext verifies context access
func TestGetContext(t *testing.T) {
	server := New(DefaultConfig())
	ctx := server.GetContext()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	select {
	case <-ctx.Done():
		t.Error("context should not be cancelled initially")
	default:
		// Expected: context not cancelled
	}
}

// TestShutdown verifies graceful shutdown
func TestShutdown(t *testing.T) {
	server := New(DefaultConfig())

	// Shutdown should complete quickly
	done := make(chan error, 1)
	go func() {
		done <- server.Shutdown()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("shutdown failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("shutdown took too long")
	}

	// Context should be cancelled after shutdown
	select {
	case <-server.GetContext().Done():
		// Expected: context cancelled
	default:
		t.Error("context should be cancelled after shutdown")
	}
}

// TestGetAddress tests address formatting
func TestGetAddress(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected string
	}{
		{"zero port", 0, ""},
		{"standard port", 8080, "localhost:8080"},
		{"high port", 50000, "localhost:50000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New(DefaultConfig())
			if tt.port > 0 {
				server.SetPort(tt.port, "127.0.0.1")
			}

			addr := server.GetAddress()
			if addr != tt.expected {
				t.Errorf("expected address %q, got %q", tt.expected, addr)
			}
		})
	}
}

// TestSetPort verifies port storage
func TestSetPort(t *testing.T) {
	server := New(DefaultConfig())

	port := 8080
	addr := "127.0.0.1"
	server.SetPort(port, addr)

	if server.port != port {
		t.Errorf("expected port %d, got %d", port, server.port)
	}
	if server.addr != addr {
		t.Errorf("expected addr %s, got %s", addr, server.addr)
	}
}
