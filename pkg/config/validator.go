// validator.go implements configuration validation logic.
// This file contains the Validator type and all validation methods
// for server and client configuration parameters. It validates ports,
// player limits, tick rates, genres, and directory paths.
//
// Package config provides configuration validation utilities.
package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/opd-ai/venture/pkg/procgen/dialog"
)

// Validator provides configuration validation for server and client settings.
type Validator struct {
	validGenres map[string]bool
}

// NewValidator creates a new configuration validator with default settings.
func NewValidator() *Validator {
	// Get valid genres from dialog package (centralized genre list)
	genres := dialog.GetAvailableGenres()
	genreMap := make(map[string]bool, len(genres))
	for _, genre := range genres {
		genreMap[genre] = true
	}

	return &Validator{
		validGenres: genreMap,
	}
}

// ValidatePort validates that a port string is within valid range (1024-65535).
// Ports below 1024 require root privileges and are not recommended.
func (v *Validator) ValidatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be a number: %w", err)
	}

	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535, got %d (ports < 1024 require root privileges)", port)
	}

	return nil
}

// ValidateMaxPlayers validates that max players is within reasonable range (1-100).
func (v *Validator) ValidateMaxPlayers(maxPlayers int) error {
	if maxPlayers < 1 {
		return fmt.Errorf("max-players must be at least 1, got %d", maxPlayers)
	}

	if maxPlayers > 100 {
		return fmt.Errorf("max-players must be at most 100, got %d (performance degrades with >100 players)", maxPlayers)
	}

	return nil
}

// ValidateTickRate validates that tick rate is within reasonable range (1-60 Hz).
func (v *Validator) ValidateTickRate(tickRate int) error {
	if tickRate < 1 {
		return fmt.Errorf("tick-rate must be at least 1 Hz, got %d", tickRate)
	}

	if tickRate > 60 {
		return fmt.Errorf("tick-rate must be at most 60 Hz, got %d (diminishing returns above 60 Hz)", tickRate)
	}

	return nil
}

// ValidateGenre validates that the genre ID is supported.
// Returns available genres in error message if invalid.
func (v *Validator) ValidateGenre(genreID string) error {
	if genreID == "" {
		return fmt.Errorf("genre cannot be empty")
	}

	if !v.validGenres[genreID] {
		available := v.GetAvailableGenres()
		return fmt.Errorf("invalid genre '%s', available genres: %s", genreID, strings.Join(available, ", "))
	}

	return nil
}

// ValidateDirectory validates that a directory path exists and is accessible.
// If create is true, attempts to create the directory if it doesn't exist.
func (v *Validator) ValidateDirectory(path string, create bool) error {
	if path == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) && create {
			// Attempt to create directory
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			return nil
		}
		return fmt.Errorf("directory %s is not accessible: %w", path, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path %s exists but is not a directory", path)
	}

	return nil
}

// GetAvailableGenres returns a sorted list of valid genre IDs.
func (v *Validator) GetAvailableGenres() []string {
	genres := make([]string, 0, len(v.validGenres))
	for genre := range v.validGenres {
		genres = append(genres, genre)
	}
	sort.Strings(genres)
	return genres
}

// ValidateAll performs validation on all common server/client configuration.
// Returns the first validation error encountered, or nil if all valid.
func (v *Validator) ValidateAll(cfg *Config) error {
	if err := v.validateServerSettings(cfg); err != nil {
		return err
	}
	if err := v.validateDirectories(cfg); err != nil {
		return err
	}
	return nil
}

// validateServerSettings validates port, max players, tick rate, and genre.
func (v *Validator) validateServerSettings(cfg *Config) error {
	if cfg.Port != "" {
		if err := v.ValidatePort(cfg.Port); err != nil {
			return err
		}
	}

	if cfg.ValidateMaxPlayers {
		if err := v.ValidateMaxPlayers(cfg.MaxPlayers); err != nil {
			return err
		}
	}

	if cfg.ValidateTickRate {
		if err := v.ValidateTickRate(cfg.TickRate); err != nil {
			return err
		}
	}

	if cfg.Genre != "" {
		if err := v.ValidateGenre(cfg.Genre); err != nil {
			return err
		}
	}

	return nil
}

// validateDirectories validates save, log, and mods directories.
func (v *Validator) validateDirectories(cfg *Config) error {
	if cfg.SaveDir != "" {
		if err := v.ValidateDirectory(cfg.SaveDir, cfg.CreateDirs); err != nil {
			return fmt.Errorf("save directory: %w", err)
		}
	}

	if cfg.LogDir != "" {
		if err := v.ValidateDirectory(cfg.LogDir, cfg.CreateDirs); err != nil {
			return fmt.Errorf("log directory: %w", err)
		}
	}

	if cfg.ModsDir != "" {
		if err := v.ValidateDirectory(cfg.ModsDir, cfg.CreateDirs); err != nil {
			return fmt.Errorf("mods directory: %w", err)
		}
	}

	return nil
}
