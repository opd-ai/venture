// constants.go defines configuration validation constants.
// This file centralizes magic numbers used for validation bounds,
// making the codebase more maintainable and self-documenting.
//
// Package config provides configuration validation utilities for server and client.
package config

// Port validation constants define the valid range for network ports.
// Ports below MinPort require root privileges on Unix systems.
const (
	// MinPort is the minimum allowed port number (1024).
	// Ports 1-1023 are privileged and require root/admin access.
	MinPort = 1024

	// MaxPort is the maximum valid port number (65535).
	// This is the highest valid TCP/UDP port number.
	MaxPort = 65535
)

// Player limit constants define the valid range for MaxPlayers setting.
const (
	// MinPlayers is the minimum allowed player count (1).
	MinPlayers = 1

	// MaxPlayers is the maximum recommended player count (100).
	// Performance degrades significantly above this threshold.
	MaxPlayersLimit = 100
)

// Tick rate constants define the valid range for server tick rate in Hz.
const (
	// MinTickRate is the minimum allowed tick rate in Hz (1).
	MinTickRate = 1

	// MaxTickRate is the maximum recommended tick rate in Hz (60).
	// Higher values provide diminishing returns for increased CPU usage.
	MaxTickRate = 60
)
