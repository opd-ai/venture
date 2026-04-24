//go:build !js || !wasm

// Package webrtc — native (non-WASM) connection implementation.
// This file provides a lightweight in-process simulation of WebRTC connection
// establishment used on native platforms and during unit tests.  Real
// browser-to-browser P2P is provided by peer_wasm.go on WASM builds.
package webrtc

import (
	"context"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
)

// launchConnectionAttempt starts the simulated connection goroutine.
// This is the native/test implementation; the WASM equivalent uses pion.
func (p *Peer) launchConnectionAttempt(ctx context.Context, _ string) {
	go func(ctx context.Context) {
		defer recovery.RecoverPanicWithLogger("webrtc_peer", "simulate connection", nil)()
		p.simulateConnection(ctx)
	}(ctx)
}

// simulateConnection simulates WebRTC connection establishment for native
// builds and tests.  It completes after a minimal delay so tests run fast,
// then starts the processMessages loop.
func (p *Peer) simulateConnection(ctx context.Context) {
	// Simulate minimal signaling + ICE gathering delay.
	time.Sleep(10 * time.Millisecond)

	select {
	case <-ctx.Done():
		p.stateChangeChan <- StateFailed
		return
	default:
	}

	p.mu.Lock()
	p.state = StateConnected
	p.stats.State = StateConnected
	p.mu.Unlock()

	p.stateChangeChan <- StateConnected

	go func() {
		defer recovery.RecoverPanicWithLogger("webrtc_peer", "process messages", nil)()
		p.processMessages()
	}()
}

// trySend is a no-op on native builds: the simulation discards outgoing
// messages because there is no real transport.  Stats are already updated
// in Send() before the message is enqueued.
func (p *Peer) trySend(_ []byte) {}
