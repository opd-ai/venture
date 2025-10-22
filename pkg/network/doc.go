// Package network provides multiplayer networking functionality including
// client-server communication, state synchronization, and lag compensation.
//
// The network package supports high-latency connections (200-5000ms) through
// client-side prediction, server reconciliation, entity interpolation, and
// server-side lag compensation for fair hit detection.
//
// Key features:
// - Binary protocol serialization
// - Client/server networking layers
// - Client-side prediction for responsive controls
// - Entity interpolation for smooth movement
// - Lag compensation for fair hit detection
// - Delta compression for bandwidth efficiency
package network
