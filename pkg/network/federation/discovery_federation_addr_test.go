package federation

import (
	"testing"
)

// TestNewDiscoverySystem_FederationAddr verifies that the federationAddr parameter
// is properly stored and returned by getLocalAddress.
func TestNewDiscoverySystem_FederationAddr(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	tests := []struct {
		name           string
		federationAddr string
		wantAddr       string
	}{
		{
			name:           "custom federation address",
			federationAddr: "192.168.1.100:9090",
			wantAddr:       "192.168.1.100:9090",
		},
		{
			name:           "empty uses default",
			federationAddr: "",
			wantAddr:       "localhost:8080",
		},
		{
			name:           "localhost with custom port",
			federationAddr: "localhost:7777",
			wantAddr:       "localhost:7777",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := NewDiscoverySystem(identity, ":0", tt.federationAddr)
			if err != nil {
				t.Fatalf("NewDiscoverySystem() error = %v", err)
			}

			// Test that getLocalAddress returns the expected address
			gotAddr := ds.getLocalAddress()
			if gotAddr != tt.wantAddr {
				t.Errorf("getLocalAddress() = %v, want %v", gotAddr, tt.wantAddr)
			}

			// Verify the field is set correctly
			if ds.federationAddr != tt.wantAddr {
				t.Errorf("ds.federationAddr = %v, want %v", ds.federationAddr, tt.wantAddr)
			}
		})
	}
}

// TestNewDiscoverySystem_BackwardCompatibility ensures that the new parameter
// works correctly with the default value when empty string is passed.
func TestNewDiscoverySystem_BackwardCompatibility(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("Failed to create identity: %v", err)
	}

	// Create discovery system with empty federation address
	ds, err := NewDiscoverySystem(identity, ":0", "")
	if err != nil {
		t.Fatalf("NewDiscoverySystem() error = %v", err)
	}

	// Should use default localhost:8080
	expectedAddr := "localhost:8080"
	if ds.getLocalAddress() != expectedAddr {
		t.Errorf("getLocalAddress() = %v, want %v", ds.getLocalAddress(), expectedAddr)
	}
}
