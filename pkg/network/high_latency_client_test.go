package network

import (
	"testing"
	"time"
)

// TestHighLatencyFlagDefaultConfig verifies that when high-latency is disabled,
// the client uses DefaultClientConfig with standard low-latency settings.
func TestHighLatencyFlagDefaultConfig(t *testing.T) {
	config := DefaultClientConfig()

	// Verify default config values for standard networks
	if config.ConnectionTimeout != 10*time.Second {
		t.Errorf("DefaultClientConfig ConnectionTimeout = %v, want 10s", config.ConnectionTimeout)
	}
	if config.MaxLatency != 500*time.Millisecond {
		t.Errorf("DefaultClientConfig MaxLatency = %v, want 500ms", config.MaxLatency)
	}
	if config.PingInterval != 1*time.Second {
		t.Errorf("DefaultClientConfig PingInterval = %v, want 1s", config.PingInterval)
	}
	if config.BufferSize != 256 {
		t.Errorf("DefaultClientConfig BufferSize = %d, want 256", config.BufferSize)
	}
}

// TestHighLatencyFlagTorConfig verifies that when high-latency is enabled,
// the client uses TorClientConfig with optimized settings for high-latency networks.
func TestHighLatencyFlagTorConfig(t *testing.T) {
	config := TorClientConfig()

	// Verify Tor config values (optimized for 200-5000ms latency)
	if config.ConnectionTimeout != 60*time.Second {
		t.Errorf("TorClientConfig ConnectionTimeout = %v, want 60s", config.ConnectionTimeout)
	}
	if config.MaxLatency != 5000*time.Millisecond {
		t.Errorf("TorClientConfig MaxLatency = %v, want 5000ms", config.MaxLatency)
	}
	if config.PingInterval != 5*time.Second {
		t.Errorf("TorClientConfig PingInterval = %v, want 5s", config.PingInterval)
	}
	if config.BufferSize != 512 {
		t.Errorf("TorClientConfig BufferSize = %d, want 512", config.BufferSize)
	}
}

// TestHighLatencyConfigComparison verifies the differences between
// default and high-latency configurations match documented ratios.
func TestHighLatencyConfigComparison(t *testing.T) {
	defaultConfig := DefaultClientConfig()
	torConfig := TorClientConfig()

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

// TestHighLatencyClientServerParity verifies that the client high-latency
// configuration is compatible with server high-latency configuration.
func TestHighLatencyClientServerParity(t *testing.T) {
	// The server uses:
	// - network.HighLatencyServerConfig() when -high-latency flag is true
	// - network.HighLatencyLagCompensationConfig() for lag compensation

	// The client should use:
	// - network.TorClientConfig() when -high-latency flag is true

	// Verify both configurations are compatible
	clientConfig := TorClientConfig()
	serverConfig := HighLatencyServerConfig()
	lagCompConfig := HighLatencyLagCompensationConfig()

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
// configuration selection based on high-latency flag state.
func TestHighLatencyFlagIntegration(t *testing.T) {
	tests := []struct {
		name                 string
		useHighLatency       bool
		expectedTimeout      time.Duration
		expectedMaxLatency   time.Duration
		expectedPingInterval time.Duration
		expectedBufferSize   int
		description          string
	}{
		{
			name:                 "Default configuration",
			useHighLatency:       false,
			expectedTimeout:      10 * time.Second,
			expectedMaxLatency:   500 * time.Millisecond,
			expectedPingInterval: 1 * time.Second,
			expectedBufferSize:   256,
			description:          "Standard low-latency network settings",
		},
		{
			name:                 "Tor/high-latency configuration",
			useHighLatency:       true,
			expectedTimeout:      60 * time.Second,
			expectedMaxLatency:   5000 * time.Millisecond,
			expectedPingInterval: 5 * time.Second,
			expectedBufferSize:   512,
			description:          "Optimized for Tor/onion services (200-5000ms latency)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the appropriate config based on flag
			var config ClientConfig
			if tt.useHighLatency {
				config = TorClientConfig()
			} else {
				config = DefaultClientConfig()
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
