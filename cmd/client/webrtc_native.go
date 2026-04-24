//go:build !js || !wasm

// Package main — native (non-WASM) WebRTC federation stub.
// wireWebRTCFederation is a no-op on native builds because browsers are not
// available; the TCP-based FederationProtocol is used instead.
package main

import "github.com/opd-ai/venture/pkg/network/federation"

// wireWebRTCFederation returns nil on native builds.
// WASM builds (webrtc_wasm.go) return a real WebRTCTransport backed by pion.
func wireWebRTCFederation(_ string) *federation.WebRTCTransport {
	return nil
}
