package choice_consequences

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// types.go defines the core domain types for choice tracking and consequences.
// This includes player choices, NPC relationships, content locks, quest branches,
// companion reactions, and the ECS component for integrating with the game engine.

// PlayerChoice represents a single decision made by a player.
type PlayerChoice struct {
	ChoiceID       string          // Unique identifier for this choice
	StoryNodeID    string          // Story node where choice was made
	Timestamp      int64           // Unix timestamp of choice
	MoralAlignment *AlignmentShift // Alignment impact
	Irreversible   bool            // If true, choice locks content permanently
	Consequences   []string        // List of consequence IDs triggered
	NPCsAffected   []string        // NPC IDs whose attitude changed
	ClassSpecific  bool            // If true, only available to specific classes
	Classes        []string        // Required classes if ClassSpecific is true
}

// NPCRelationship tracks a player's relationship with a specific NPC.
type NPCRelationship struct {
	NPCID           string           // NPC identifier
	Attitude        float64          // -1.0 (hostile) to +1.0 (friendly)
	TrustLevel      float64          // 0.0 (stranger) to 1.0 (trusted ally)
	LastUpdate      int64            // Unix timestamp of last interaction
	MemorableEvents []MemorableEvent // Events NPC remembers
	DialogueUnlocks []string         // Unlocked dialogue options
	QuestsAffected  []string         // Quests impacted by relationship
}

// MemorableEvent represents a significant player action remembered by an NPC.
type MemorableEvent struct {
	EventID     string  // Unique event identifier
	ChoiceID    string  // Related player choice
	Timestamp   int64   // When event occurred
	Impact      float64 // -1.0 (very negative) to +1.0 (very positive)
	Description string  // Human-readable description
}

// ContentLock represents content that has been locked due to player choices.
type ContentLock struct {
	ContentID        string   // Locked content identifier (quest, NPC, area)
	LockedBy         string   // Choice ID that caused lock
	LockType         LockType // Type of lock
	Timestamp        int64    // When content was locked
	Permanent        bool     // If true, cannot be unlocked
	UnlockConditions []string // Conditions needed to unlock (if not permanent)
}

// LockType represents different types of content locks.
type LockType int

const (
	LockTypeQuest     LockType = iota // Quest line locked
	LockTypeNPC                       // NPC unavailable
	LockTypeArea                      // Location inaccessible
	LockTypeDialogue                  // Dialogue option removed
	LockTypeReward                    // Reward path closed
	LockTypeCompanion                 // Companion recruitment blocked
)

// String returns the string representation of LockType.
func (lt LockType) String() string {
	switch lt {
	case LockTypeQuest:
		return "Quest"
	case LockTypeNPC:
		return "NPC"
	case LockTypeArea:
		return "Area"
	case LockTypeDialogue:
		return "Dialogue"
	case LockTypeReward:
		return "Reward"
	case LockTypeCompanion:
		return "Companion"
	default:
		return "Unknown"
	}
}

// QuestBranch represents a branching quest outcome.
type QuestBranch struct {
	QuestID         string   // Quest identifier
	BranchID        string   // Specific branch identifier
	Prerequisites   []string // Choice IDs required for this branch
	Outcomes        []string // Consequence IDs from completing this branch
	ClassRestricted bool     // If true, only specific classes can access
	AllowedClasses  []string // Classes allowed to take this branch
}

// ClassSpecificQuest represents a quest available only to certain classes.
type ClassSpecificQuest struct {
	QuestID       string                // Quest identifier
	RequiredClass string                // Class required to accept quest
	MinLevel      int                   // Minimum level required
	AlignmentReq  *AlignmentRequirement // Optional alignment requirement
	Prerequisites []string              // Previous choices required
}

// CompanionReaction represents how a companion reacts to player choices.
type CompanionReaction struct {
	CompanionID    string  // Companion identifier
	ChoiceID       string  // Choice that triggered reaction
	LoyaltyDelta   float64 // Change in loyalty (-1.0 to +1.0)
	Approval       bool    // True if companion approves, false if disapproves
	Comment        string  // Companion's verbal reaction
	LeaveThreshold float64 // Loyalty level at which companion leaves (if crossed)
}

// ChoiceTrackerComponent is an ECS component for tracking player choices.
type ChoiceTrackerComponent struct {
	PlayerID           string                      // Player entity ID
	ChoiceHistory      []*PlayerChoice             // All choices made
	NPCRelationships   map[string]*NPCRelationship // NPC ID -> relationship
	ContentLocks       map[string]*ContentLock     // Content ID -> lock
	Alignment          *PlayerAlignment            // Current alignment
	CompanionReactions []*CompanionReaction        // Recent companion reactions
}

// Type returns the component type identifier.
func (c ChoiceTrackerComponent) Type() string {
	return "choice_tracker"
}

// choiceTrackerLogger provides structured logging for component operations.
var choiceTrackerLogger = log.WithField("component_type", "choice_tracker")

// Serialize encodes the component to bytes for persistence.
func (c *ChoiceTrackerComponent) Serialize() ([]byte, error) {
	choiceTrackerLogger.WithFields(log.Fields{
		"player_id":      c.PlayerID,
		"choice_count":   len(c.ChoiceHistory),
		"npc_count":      len(c.NPCRelationships),
		"lock_count":     len(c.ContentLocks),
		"reaction_count": len(c.CompanionReactions),
	}).Debug("Serializing choice tracker component")

	data, err := json.Marshal(c)
	if err != nil {
		choiceTrackerLogger.WithFields(log.Fields{
			"player_id": c.PlayerID,
			"error":     err.Error(),
		}).Error("Failed to serialize choice tracker component")
		return nil, fmt.Errorf("failed to serialize choice tracker: %w", err)
	}

	choiceTrackerLogger.WithFields(log.Fields{
		"player_id": c.PlayerID,
		"bytes":     len(data),
	}).Debug("Choice tracker component serialized successfully")

	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *ChoiceTrackerComponent) Deserialize(data []byte) error {
	choiceTrackerLogger.WithFields(log.Fields{
		"bytes": len(data),
	}).Debug("Deserializing choice tracker component")

	if len(data) == 0 {
		choiceTrackerLogger.WithFields(log.Fields{
			"expected_bytes": ">0",
			"received_bytes": 0,
		}).Error("Empty data for choice tracker component deserialization")
		return fmt.Errorf("empty data for choice tracker deserialization")
	}

	if err := json.Unmarshal(data, c); err != nil {
		choiceTrackerLogger.WithFields(log.Fields{
			"bytes": len(data),
			"error": err.Error(),
		}).Error("Failed to deserialize choice tracker component")
		return fmt.Errorf("failed to deserialize choice tracker: %w", err)
	}

	choiceTrackerLogger.WithFields(log.Fields{
		"player_id":      c.PlayerID,
		"choice_count":   len(c.ChoiceHistory),
		"npc_count":      len(c.NPCRelationships),
		"lock_count":     len(c.ContentLocks),
		"reaction_count": len(c.CompanionReactions),
	}).Debug("Choice tracker component deserialized successfully")

	return nil
}
