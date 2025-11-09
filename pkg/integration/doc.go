// Package integration provides integration tests for Venture's multiplayer and cross-system functionality.
// These tests verify that different subsystems work correctly together, particularly for multiplayer scenarios.
//
// Test coverage includes:
// - Deterministic content generation across multiple clients
// - Cross-genre content generation consistency
// - Save/load compatibility across versions
// - Network latency simulation and lag compensation
//
// Integration tests use real generators and systems, unlike unit tests which may use mocks.
// These tests are critical for ensuring multiplayer synchronization and backward compatibility.
package integration
