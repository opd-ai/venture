// Package main demonstrates voice chat integration with the Venture game engine.
//
// # Overview
//
// This example shows how to integrate the voice codec system with the engine's
// voice channel and spatial audio systems. It demonstrates:
//
//   - Audio manager setup with voice codec initialization
//   - Voice system creation (VoiceChannelSystem, SpatialVoiceSystem, VoiceAudioSystem)
//   - Entity setup with voice components
//   - Voice channel management (joining party channels)
//   - Voice input simulation and transmission
//   - Spatial audio calculations (distance-based volume and panning)
//
// # What This Example Demonstrates
//
//   - Voice codec API usage (encoding/decoding)
//   - Voice system registration pattern
//   - Component wiring (VoiceAudioComponent, SpatialVoiceComponent)
//   - Voice channel lifecycle management
//   - Voice input processing (ProcessInput) and output retrieval (ProcessOutput)
//   - Spatial audio parameters (volume, pan, distance)
//   - Voice activity detection simulation
//
// # What Real Implementation Needs
//
// This is a simplified demonstration. A production voice chat implementation requires:
//
//   - Actual microphone input capture (not simulated audio)
//   - Real network transport (UDP/WebRTC) instead of in-memory queue
//   - Audio playback for received voice samples
//   - Push-to-talk or voice activity detection from real audio
//   - Voice activity visualization in UI
//   - Echo cancellation and noise suppression
//   - Mobile permissions handling (iOS/Android microphone access)
//   - WASM browser API integration (getUserMedia, Web Audio API)
//
// # Usage
//
// Run the example:
//
//	go run examples/voice_integration_demo/main.go
//
// Expected output shows:
//  1. Audio manager initialization with sample rate and bitrate
//  2. Voice system creation and configuration
//  3. Player entity setup with voice components
//  4. Voice channel join confirmation
//  5. Voice input simulation and encoding (ProcessInput)
//  6. Voice output retrieval and decoding (ProcessOutput)
//  7. Spatial audio calculations (distance, volume, pan)
//  8. Channel information display
//
// # Prerequisites
//
// No special dependencies required. The example uses:
//
//   - pkg/audio - Audio manager and voice codec
//   - pkg/engine - ECS world, voice systems, and components
//
// No microphone access or network connectivity needed for this demo.
//
// # Platform Compatibility
//
// The example builds on all platforms (Linux, macOS, Windows, WASM, mobile)
// but voice chat requires platform-specific implementations:
//
//   - Desktop: Microphone access via OS audio APIs
//   - WASM: getUserMedia API + WebRTC DataChannels
//   - Mobile: iOS/Android microphone permissions + background audio
//
// # Related Documentation
//
// See docs/VOICE_CHAT.md for comprehensive voice chat architecture documentation.
//
// # Note on Logging
//
// This example uses simple log.Fatal() for clarity. Production code should use
// structured logging with logrus.WithFields() as demonstrated throughout the
// codebase.
package main
