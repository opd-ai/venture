package persistence

import "time"

// Chat History Configuration
// Originally defined in: chat_history.go

// MaxMessagesPerPlayer is the maximum chat history size per player
const MaxMessagesPerPlayer = 1000

// MaxMessageAge is the maximum age for messages (30 days)
const MaxMessageAge = 30 * 24 * time.Hour

// MaxChangelogSize is the maximum number of changelog entries to retain.
// The changelog is used for efficient delta synchronization and operates as
// a circular buffer. When exceeded, oldest entries are discarded.
const MaxChangelogSize = 1000

// Image Gallery Configuration
// Originally defined in: image_gallery.go

// MaxImagesPerPlayer is the maximum number of images per player
const MaxImagesPerPlayer = 100

// MaxImageSizeBytes is the maximum size per image (500KB)
const MaxImageSizeBytes = 500 * 1024

// Image Format Constants
// Originally defined in: image_gallery.go

const (
	// ImageFormatPNG is the PNG format
	ImageFormatPNG ImageFormat = "png"
	// ImageFormatJPEG is the JPEG format
	ImageFormatJPEG ImageFormat = "jpeg"
)

// Reputation Category Constants
// Originally defined in: reputation_manager.go

const (
	// ReputationTrade is reputation from trading activities
	ReputationTrade ReputationCategory = "trade"
	// ReputationCombat is reputation from combat/PvP activities
	ReputationCombat ReputationCategory = "combat"
	// ReputationSocial is reputation from social interactions
	ReputationSocial ReputationCategory = "social"
	// ReputationQuest is reputation from quest completions
	ReputationQuest ReputationCategory = "quest"
)

// Trust Level Constants
// Originally defined in: types.go

const (
	// TrustLevelStranger is the lowest trust level (0.0-0.3)
	TrustLevelStranger TrustLevel = iota
	// TrustLevelAcquaintance is basic familiarity (0.3-0.6)
	TrustLevelAcquaintance
	// TrustLevelFriend is established friendship (0.6-0.8)
	TrustLevelFriend
	// TrustLevelTrusted is highest trust (0.8-1.0)
	TrustLevelTrusted
)

// Trust Score Limits
// Originally defined in: types.go

const (
	// DecayRatePerDay is the trust decay rate (0.01 per day)
	DecayRatePerDay = 0.01
	// MinTrustScore is the minimum trust value
	MinTrustScore = 0.0
	// MaxTrustScore is the maximum trust value
	MaxTrustScore = 1.0
)
