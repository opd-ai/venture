// Package chat provides player-to-player chat functionality with validation
// and rate limiting.
//
// This package wraps the engine chat system with additional network security
// features including message sanitization via pkg/validation and rate limiting
// to prevent DoS attacks.
//
// Note: This package uses time.Now() for message timestamps, which is intentional
// and exempt from the deterministic-procgen rule. Network chat messages inherently
// require real timestamps for multiplayer synchronization.
//
// For enhanced encryption support, see pkg/engine/chat_system.go which provides
// the full chat system with optional encryption capabilities.
package chat
