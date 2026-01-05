// Package config provides configuration validation utilities for server and client.
//
// This package validates command-line flags and environment variables to ensure
// they meet requirements before starting the game server or client. Invalid
// configurations are reported with helpful error messages.
//
// Example usage:
//
//	validator := config.NewValidator()
//	if err := validator.ValidatePort("8080"); err != nil {
//	    log.Fatalf("Invalid port: %v", err)
//	}
//
//	if err := validator.ValidateGenre("fantasy"); err != nil {
//	    log.Fatalf("Invalid genre: %v", err)
//	}
package config
