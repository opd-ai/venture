// Package network provides multiplayer networking functionality including
// client-server communication, state synchronization, and lag compensation.
//
// The network package supports high-latency connections (200-5000ms) through
// client-side prediction, server reconciliation, entity interpolation, and
// server-side lag compensation for fair hit detection.
//
// # Configuration Presets
//
// Two configuration presets are available for different network environments:
//
// DefaultServerConfig/DefaultClientConfig:
//   - Optimized for low-latency LAN/internet connections (<100ms)
//   - Read timeout: 10s, Write timeout: 5s
//   - Buffer size: 256 messages
//   - Suitable for most gameplay scenarios
//
// HighLatencyServerConfig/TorClientConfig:
//   - Optimized for high-latency connections (200-5000ms)
//   - Read timeout: 60s, Write timeout: 30s
//   - Buffer size: 512 messages
//   - Supports Tor/onion services and slow networks
//   - Includes TCP keepalive configuration (30s period)
//
// # Usage
//
// Start a server with high-latency support:
//
//	./venture-server --high-latency --port 8080
//
// Connect a client through Tor:
//
//	config := network.TorClientConfig()
//	config.ServerAddress = "example.onion:8080"
//	client := network.NewClient(config)
//
// # Key Features
//
// - Binary protocol serialization
// - Client/server networking layers
// - Client-side prediction for responsive controls
// - Entity interpolation for smooth movement
// - Lag compensation for fair hit detection
// - Delta compression for bandwidth efficiency
// - TCP keepalive for long-duration connections
package network
