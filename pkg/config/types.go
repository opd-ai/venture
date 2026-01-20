// types.go defines configuration data structures.
// This file contains the Config type used for validation input.
//
// Package config provides configuration validation utilities for server and client.
package config

// Config holds configuration values to validate.
// Originally from: validator.go
type Config struct {
	// Port is the server port number in string format.
	// Valid range: "1024" to "65535" (ports < 1024 require root privileges)
	Port string

	// MaxPlayers is the maximum number of concurrent players.
	// Valid range: 1 to 100 (performance degrades above 100)
	MaxPlayers int

	// ValidateMaxPlayers indicates whether MaxPlayers field should be validated.
	ValidateMaxPlayers bool

	// TickRate is the server tick rate in Hz (updates per second).
	// Valid range: 1 to 60 Hz (diminishing returns above 60 Hz)
	TickRate int

	// ValidateTickRate indicates whether TickRate field should be validated.
	ValidateTickRate bool

	// Genre is the game genre identifier.
	// Valid values can be retrieved via Validator.GetAvailableGenres()
	Genre string

	// SaveDir is the directory path for game save files.
	// Will be created if CreateDirs is true and directory doesn't exist.
	SaveDir string

	// LogDir is the directory path for log files.
	// Will be created if CreateDirs is true and directory doesn't exist.
	LogDir string

	// ModsDir is the directory path for game modifications.
	// Will be created if CreateDirs is true and directory doesn't exist.
	ModsDir string

	// CreateDirs indicates whether to create missing directories during validation.
	CreateDirs bool
}
