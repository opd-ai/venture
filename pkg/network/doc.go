// Package network provides multiplayer networking functionality including
// client-server communication, state synchronization, lag compensation, and
// player-to-player chat with end-to-end encryption.
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
// # Chat System (V5.0 Phase 21)
//
// The chat system provides encrypted player-to-player messaging with:
//
// - Four channels: Global, Local (range-limited), Party, Whisper
// - End-to-end encryption using Diffie-Hellman key exchange and AES-256-GCM
// - ACK/NACK protocol with automatic retries for reliability
// - Rate limiting and mute enforcement (30s base, doubles per violation)
// - Client-side profanity filter (opt-in, user-configurable)
// - Range validation for local chat (default 10 tiles, extendable with items)
//
// Example chat usage:
//
//	// Create chat manager
//	cm := network.NewChatManager()
//
//	// Register player with encryption key
//	params := network.DefaultDHParams()
//	keyPair, _ := network.GenerateKeyPair(params)
//	secret, _ := network.ComputeSharedSecret(keyPair.PrivateKey, peerPublicKey, params)
//	encKey := network.DeriveAESKey(secret)
//	cm.AddPlayer(playerID, position, encKey)
//
//	// Send message
//	packet, err := cm.SendMessage(playerID, 0, "Hello!", recipientID, -1)
//
//	// Process ACK
//	ack := &network.MessageACK{MessageID: packet.MessageID, Success: true}
//	cm.ProcessACK(ack)
//
// Profanity filter:
//
//	pf := network.NewProfanityFilter()
//	pf.Enable()
//	filtered := pf.Filter("Message text")
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
// - E2E encrypted chat with ACK/NACK reliability
// - Profanity filtering (client-side, opt-in)
package network
