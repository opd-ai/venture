//go:build !js || !wasm

// Package webrtc — native (non-WASM) signaling transport stub.
// connectTransport, sendViaTransport, and closeTransport are no-ops on native
// builds.  The in-process message loop in processMessages() still runs, but
// outgoing messages are dropped because there is no real WebSocket to send
// them to.  The stub is sufficient for unit tests that exercise the signaling
// API without a live server.
package webrtc

// connectTransport is a no-op on native/test builds.
func (s *SignalingClient) connectTransport() error { return nil }

// sendViaTransport drops the message on native builds.
// Unit tests that need delivery should use SignalingServer.RelayMessage directly.
func (s *SignalingClient) sendViaTransport(_ *SignalingMessage) {}

// closeTransport is a no-op on native builds (no resources to release).
func (s *SignalingClient) closeTransport() {}
