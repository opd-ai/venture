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
// # Image Sharing System (V5.0 Phase 23)
//
// The image sharing system enables players to upload and share images with:
//
// - Chunked transfer (64KB chunks, resume on disconnect)
// - Automatic thumbnail generation (128×128 JPEG, quality 75)
// - Size/type validation (<500KB, PNG/JPEG/GIF only, <2048×2048 pixels)
// - Rate limiting (1 image per 60 seconds per player)
// - Moderation hooks for server-side content filtering
// - Automatic expiry (10 minutes timeout OR sender disconnect)
// - Channel-based distribution (global, local, party, whisper)
//
// Example image sharing usage:
//
//	// Create image manager
//	im := network.NewImageManager()
//
//	// Set moderation hook (optional)
//	im.SetModerationHook(func(metadata *network.ImageMetadata, data []byte) error {
//		// Return error to reject image
//		return nil
//	})
//
//	// Set thumbnail relay callback
//	im.SetThumbnailRelayCallback(func(metadata *network.ImageMetadata, thumbnail []byte) {
//		// Relay thumbnail to recipients
//	})
//
//	// Upload image
//	req := &network.ImageUploadRequest{
//		SenderID: playerID,
//		Channel:  0,
//		Format:   "png",
//		Data:     imageData,
//	}
//	metadata, err := im.UploadImage(req)
//
//	// Download image (chunked)
//	totalChunks, _ := im.StartChunkedDownload(requesterID, metadata.ImageID)
//	for i := 0; i < totalChunks; i++ {
//		chunk, _ := im.GetNextChunk(requesterID, metadata.ImageID)
//		// Process chunk
//	}
//
// Constraints (Phase 23 spec):
//
// - Max size: 500KB, Max dimensions: 2048×2048
// - Supported formats: PNG, JPEG, GIF (non-animated)
// - Rate limit: 1 upload per 60 seconds per player
// - Expiry: 10 minutes OR sender disconnect
// - Bandwidth: <2MB/s upload, <25KB/s overhead per player
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
// - E2E encrypted chat with ACK/NACK reliability (Phase 21)
// - Profanity filtering - client-side, opt-in (Phase 21)
// - Image sharing with chunked transfer and moderation (Phase 23)
package network
