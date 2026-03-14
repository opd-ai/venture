package persistence

import (
	"time"
)

// Type Definitions
// This file consolidates all type definitions for the persistence package

// TimeProvider is an interface for obtaining the current time.
// This enables deterministic timestamps for testing and reproducible ID generation.
// In production, use RealTimeProvider (via DefaultTimeProvider()); in tests, use
// a mock implementation that returns a fixed time.
//
// The TimeProvider pattern is used consistently across all managers in this package:
// - ImageGallery: Uses TimeProvider for image timestamp and ID generation
// - TrustManager: Uses TimeProvider in automatic decay loop
// - ReputationManager: Supports TimeProvider for consistency (via constructor injection)
// - ChatHistory: Supports TimeProvider for deterministic testing (via constructor injection)
//
// IMPORTANT: TimeProvider exists SOLELY FOR TEST DETERMINISM, not for procedural content
// generation (Coding Guideline #2). This package manages social metadata (trust scores,
// timestamps, chat history IDs, image upload times) which are server-side operational
// data that do not affect procedurally generated game content (terrain, items, quests).
//
// Production behavior: All managers use RealTimeProvider (real wall-clock time) by default
// for operational timestamps and decay scheduling.
//
// Test behavior: Tests inject a fixed or mock TimeProvider to eliminate time-dependent
// test flakiness and enable reproducible test execution.
//
// This is an intentional exception to strict determinism guidelines because social
// metadata timestamps are not part of the seed-based procedural generation system.
type TimeProvider interface {
	// Now returns the current time
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using the actual system clock.
//
// This is the ONLY implementation used in production. The time.Now() call here
// is an intentional exception to Coding Guideline #2 because social metadata
// (trust decay, timestamps, IDs) are server-side operational data, not procedural
// content generation.
type RealTimeProvider struct{}

// Now returns the current system time for production use.
//
// This method intentionally calls time.Now() for server-side operational timestamps.
// See TimeProvider interface godoc for rationale on determinism exception.
func (RealTimeProvider) Now() time.Time {
	return time.Now()
}

// DefaultTimeProvider returns the default TimeProvider (real system time).
func DefaultTimeProvider() TimeProvider {
	return RealTimeProvider{}
}

// ImageFormat represents supported image formats
// Originally from: image_gallery.go
type ImageFormat string

// ReputationCategory represents different types of reputation
// Originally from: reputation_manager.go
type ReputationCategory string

// TrustLevel represents a tier of trust between players
type TrustLevel int

// String returns the human-readable name of a TrustLevel
func (t TrustLevel) String() string {
	switch t {
	case TrustLevelStranger:
		return "Stranger"
	case TrustLevelAcquaintance:
		return "Acquaintance"
	case TrustLevelFriend:
		return "Friend"
	case TrustLevelTrusted:
		return "Trusted"
	default:
		return "Unknown"
	}
}

// TrustRecord stores trust information between two players
type TrustRecord struct {
	// PlayerA is the first player ID (lexicographically sorted)
	PlayerA string
	// PlayerB is the second player ID (lexicographically sorted)
	PlayerB string
	// Score is the trust value (0.0-1.0)
	Score float64
	// LastUpdate is when the trust was last modified
	LastUpdate time.Time
	// Interactions is the number of positive interactions
	Interactions int
}

// GetTrustLevel returns the trust tier for a given score
func GetTrustLevel(score float64) TrustLevel {
	if score < 0.3 {
		return TrustLevelStranger
	} else if score < 0.6 {
		return TrustLevelAcquaintance
	} else if score < 0.8 {
		return TrustLevelFriend
	}
	return TrustLevelTrusted
}

// CanTradeRarity returns true if the trust level allows trading items of the given rarity
func CanTradeRarity(level TrustLevel, rarity string) bool {
	rarityLevel := map[string]int{
		"common":    0,
		"uncommon":  1,
		"rare":      2,
		"epic":      3,
		"legendary": 4,
	}

	maxRarity := map[TrustLevel]int{
		TrustLevelStranger:     0, // common only
		TrustLevelAcquaintance: 1, // common + uncommon
		TrustLevelFriend:       2, // up to rare
		TrustLevelTrusted:      4, // all items
	}

	itemLevel, exists := rarityLevel[rarity]
	if !exists {
		return false
	}

	maxLevel, exists := maxRarity[level]
	if !exists {
		return false
	}

	return itemLevel <= maxLevel
}

// ImageThumbnail is lightweight metadata about a stored image, without the image data itself.
// Returned by ImageGallery.GetThumbnails for efficient listing without memory overhead.
type ImageThumbnail struct {
	ID        string      `json:"id"`
	OwnerID   string      `json:"owner_id"`
	Title     string      `json:"title,omitempty"`
	Format    ImageFormat `json:"format"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
	SizeBytes int         `json:"size_bytes"`
	Timestamp time.Time   `json:"timestamp"`
	Tags      []string    `json:"tags,omitempty"`
}
