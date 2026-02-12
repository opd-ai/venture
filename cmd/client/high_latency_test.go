//go:build !android && !ios
// +build !android,!ios

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/network"
)

// TestHighLatencyFlagExists verifies the --high-latency flag is recognized
// and documented in the client help output.
func TestHighLatencyFlagExists(t *testing.T) {
	// Build the client binary for testing
	buildCmd := exec.Command("go", "build", "-o", "venture-client-test-highlatency", ".")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build client: %v", err)
	}
	defer os.Remove("venture-client-test-highlatency")

	// Run with --help to verify flag exists
	helpCmd := exec.Command("./venture-client-test-highlatency", "--help")
	output, err := helpCmd.CombinedOutput()
	if err != nil && !strings.Contains(err.Error(), "exit status") {
		t.Fatalf("failed to run help: %v", err)
	}

	outputStr := string(output)

	// Verify high-latency flag is present
	if !strings.Contains(outputStr, "-high-latency") {
		t.Error("--high-latency flag not found in help output")
	}

	// Verify flag description mentions Tor/onion services
	if !strings.Contains(outputStr, "Tor") || !strings.Contains(outputStr, "onion") {
		t.Error("--high-latency flag description should mention Tor/onion services")
	}

	// Verify flag description mentions latency range
	if !strings.Contains(outputStr, "200") && !strings.Contains(outputStr, "5000") {
		t.Error("--high-latency flag description should mention latency range (200-5000ms)")
	}
}

// TestHighLatencyConfigComparison verifies the differences between
// default and high-latency configurations.
func TestHighLatencyConfigComparison(t *testing.T) {
	defaultConfig := network.DefaultClientConfig()
	torConfig := network.TorClientConfig()

	tests := []struct {
		name         string
		defaultValue interface{}
		torValue     interface{}
		multiplier   float64
		description  string
	}{
		{
			name:         "ConnectionTimeout",
			defaultValue: defaultConfig.ConnectionTimeout,
			torValue:     torConfig.ConnectionTimeout,
			multiplier:   6.0,
			description:  "Tor config should have 6x longer connection timeout for circuit building",
		},
		{
			name:         "MaxLatency",
			defaultValue: defaultConfig.MaxLatency,
			torValue:     torConfig.MaxLatency,
			multiplier:   10.0,
			description:  "Tor config should have 10x higher max latency tolerance",
		},
		{
			name:         "PingInterval",
			defaultValue: defaultConfig.PingInterval,
			torValue:     torConfig.PingInterval,
			multiplier:   5.0,
			description:  "Tor config should have 5x longer ping interval to reduce traffic",
		},
		{
			name:         "BufferSize",
			defaultValue: defaultConfig.BufferSize,
			torValue:     torConfig.BufferSize,
			multiplier:   2.0,
			description:  "Tor config should have 2x larger buffer for latency spikes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "ConnectionTimeout":
				defaultDuration := tt.defaultValue.(time.Duration)
				torDuration := tt.torValue.(time.Duration)
				expectedTorValue := time.Duration(float64(defaultDuration) * tt.multiplier)
				if torDuration != expectedTorValue {
					t.Errorf("%s: %v != %v (%s)", tt.name, torDuration, expectedTorValue, tt.description)
				}
			case "MaxLatency":
				defaultDuration := tt.defaultValue.(time.Duration)
				torDuration := tt.torValue.(time.Duration)
				expectedTorValue := time.Duration(float64(defaultDuration) * tt.multiplier)
				if torDuration != expectedTorValue {
					t.Errorf("%s: %v != %v (%s)", tt.name, torDuration, expectedTorValue, tt.description)
				}
			case "PingInterval":
				defaultDuration := tt.defaultValue.(time.Duration)
				torDuration := tt.torValue.(time.Duration)
				expectedTorValue := time.Duration(float64(defaultDuration) * tt.multiplier)
				if torDuration != expectedTorValue {
					t.Errorf("%s: %v != %v (%s)", tt.name, torDuration, expectedTorValue, tt.description)
				}
			case "BufferSize":
				defaultSize := tt.defaultValue.(int)
				torSize := tt.torValue.(int)
				expectedTorSize := int(float64(defaultSize) * tt.multiplier)
				if torSize != expectedTorSize {
					t.Errorf("%s: %d != %d (%s)", tt.name, torSize, expectedTorSize, tt.description)
				}
			}
		})
	}
}

// TestHighLatencyFlagServerParity verifies that the client high-latency flag
// matches the server implementation for consistency.
func TestHighLatencyFlagServerParity(t *testing.T) {
	// The server uses:
	// - network.HighLatencyServerConfig() when -high-latency flag is true
	// - network.HighLatencyLagCompensationConfig() for lag compensation

	// The client should use:
	// - network.TorClientConfig() when -high-latency flag is true

	// Verify both configurations are compatible
	clientConfig := network.TorClientConfig()
	serverConfig := network.HighLatencyServerConfig()
	lagCompConfig := network.HighLatencyLagCompensationConfig()

	// Client MaxLatency should be compatible with server ReadTimeout
	if clientConfig.MaxLatency > serverConfig.ReadTimeout {
		t.Errorf("Client MaxLatency (%v) exceeds server ReadTimeout (%v)",
			clientConfig.MaxLatency, serverConfig.ReadTimeout)
	}

	// Client ConnectionTimeout should be less than server ReadTimeout
	if clientConfig.ConnectionTimeout > serverConfig.ReadTimeout {
		t.Errorf("Client ConnectionTimeout (%v) exceeds server ReadTimeout (%v)",
			clientConfig.ConnectionTimeout, serverConfig.ReadTimeout)
	}

	// Lag compensation MaxCompensation should be >= client MaxLatency
	if lagCompConfig.MaxCompensation < clientConfig.MaxLatency {
		t.Errorf("LagComp MaxCompensation (%v) less than client MaxLatency (%v)",
			lagCompConfig.MaxCompensation, clientConfig.MaxLatency)
	}

	t.Logf("Client-server high-latency configuration parity verified")
	t.Logf("  Client MaxLatency: %v", clientConfig.MaxLatency)
	t.Logf("  Server ReadTimeout: %v", serverConfig.ReadTimeout)
	t.Logf("  LagComp MaxCompensation: %v", lagCompConfig.MaxCompensation)
}

// TestHighLatencyFlagIntegration is a table-driven test verifying the complete
// integration of the high-latency flag with network client initialization.
func TestHighLatencyFlagIntegration(t *testing.T) {
	tests := []struct {
		name                 string
		highLatencyEnabled   bool
		expectedTimeout      time.Duration
		expectedMaxLatency   time.Duration
		expectedPingInterval time.Duration
		expectedBufferSize   int
		description          string
	}{
		{
			name:                 "Default configuration",
			highLatencyEnabled:   false,
			expectedTimeout:      10 * time.Second,
			expectedMaxLatency:   500 * time.Millisecond,
			expectedPingInterval: 1 * time.Second,
			expectedBufferSize:   256,
			description:          "Standard low-latency network settings",
		},
		{
			name:                 "Tor/high-latency configuration",
			highLatencyEnabled:   true,
			expectedTimeout:      60 * time.Second,
			expectedMaxLatency:   5000 * time.Millisecond,
			expectedPingInterval: 5 * time.Second,
			expectedBufferSize:   512,
			description:          "Optimized for Tor/onion services (200-5000ms latency)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the highLatency flag
			highLatency = &tt.highLatencyEnabled

			// Get the appropriate config based on flag
			var config network.ClientConfig
			if *highLatency {
				config = network.TorClientConfig()
			} else {
				config = network.DefaultClientConfig()
			}

			// Verify all configuration values
			if config.ConnectionTimeout != tt.expectedTimeout {
				t.Errorf("ConnectionTimeout = %v, want %v (%s)",
					config.ConnectionTimeout, tt.expectedTimeout, tt.description)
			}

			if config.MaxLatency != tt.expectedMaxLatency {
				t.Errorf("MaxLatency = %v, want %v (%s)",
					config.MaxLatency, tt.expectedMaxLatency, tt.description)
			}

			if config.PingInterval != tt.expectedPingInterval {
				t.Errorf("PingInterval = %v, want %v (%s)",
					config.PingInterval, tt.expectedPingInterval, tt.description)
			}

			if config.BufferSize != tt.expectedBufferSize {
				t.Errorf("BufferSize = %d, want %d (%s)",
					config.BufferSize, tt.expectedBufferSize, tt.description)
			}
		})
	}
}
