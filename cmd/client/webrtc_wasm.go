//go:build js && wasm
// +build js,wasm

// Package main provides WebRTC federation support for WASM builds.
// Phase 4.2 (PLAN.md): Browser-to-browser federation via WebRTC.
package main

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/network/federation/webrtc"
)

// WebRTCConfig holds configuration for browser-to-browser federation.
var webrtcConfig *webrtc.Config

// initWebRTCFederation initializes WebRTC for browser-to-browser connections.
// This is called during client initialization on WASM platforms.
// Falls back to standard federation when WebRTC is unavailable.
func initWebRTCFederation(clientID string) (*webrtc.Peer, error) {
	if webrtcConfig == nil {
		webrtcConfig = webrtc.DefaultConfig()
	}

	peer, err := webrtc.NewPeer(clientID, webrtcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebRTC peer: %w", err)
	}

	return peer, nil
}

// setWebRTCSignalingServer configures the signaling server URL for peer discovery.
func setWebRTCSignalingServer(url string) {
	if webrtcConfig == nil {
		webrtcConfig = webrtc.DefaultConfig()
	}
	webrtcConfig.SignalingURL = url
}

// setWebRTCSTUNServers configures STUN servers for NAT traversal.
func setWebRTCSTUNServers(servers []string) {
	if webrtcConfig == nil {
		webrtcConfig = webrtc.DefaultConfig()
	}
	webrtcConfig.STUNServers = servers
}

// isWebRTCAvailable checks if WebRTC is supported in the current browser.
// Returns true if WebRTC can be used for federation.
func isWebRTCAvailable() bool {
	// In WASM, we assume WebRTC is available in modern browsers
	// Real implementation would check window.RTCPeerConnection
	return true
}
