// Package engine provides reputation-based quest gating components.
package engine

// ReputationTier represents faction standing levels that gate quest access.
type ReputationTier int

const (
	// TierHated requires -75 or below reputation
	TierHated ReputationTier = iota
	// TierHostile requires -50 to -74 reputation
	TierHostile
	// TierUnfriendly requires -25 to -49 reputation
	TierUnfriendly
	// TierNeutral requires -24 to +24 reputation
	TierNeutral
	// TierFriendly requires +25 to +49 reputation
	TierFriendly
	// TierHonored requires +50 to +74 reputation
	TierHonored
	// TierRevered requires +75 or above reputation
	TierRevered
)

// String returns the tier name.
func (t ReputationTier) String() string {
	switch t {
	case TierHated:
		return "Hated"
	case TierHostile:
		return "Hostile"
	case TierUnfriendly:
		return "Unfriendly"
	case TierNeutral:
		return "Neutral"
	case TierFriendly:
		return "Friendly"
	case TierHonored:
		return "Honored"
	case TierRevered:
		return "Revered"
	default:
		return "Unknown"
	}
}

// MinReputation returns the minimum reputation value for this tier.
func (t ReputationTier) MinReputation() float64 {
	switch t {
	case TierHated:
		return -100.0
	case TierHostile:
		return -75.0
	case TierUnfriendly:
		return -50.0
	case TierNeutral:
		return -25.0
	case TierFriendly:
		return 25.0
	case TierHonored:
		return 50.0
	case TierRevered:
		return 75.0
	default:
		return 0.0
	}
}

// ReputationFromValue returns the tier for a given reputation value.
func ReputationFromValue(rep float64) ReputationTier {
	switch {
	case rep >= 75.0:
		return TierRevered
	case rep >= 50.0:
		return TierHonored
	case rep >= 25.0:
		return TierFriendly
	case rep > -25.0:
		return TierNeutral
	case rep > -50.0:
		return TierUnfriendly
	case rep > -75.0:
		return TierHostile
	default:
		return TierHated
	}
}

// QuestReputationRequirement defines the reputation requirements for a quest.
type QuestReputationRequirement struct {
	// FactionID is the faction whose reputation is checked
	FactionID string

	// MinTier is the minimum tier required (inclusive)
	MinTier ReputationTier

	// MaxTier is the maximum tier allowed (inclusive, optional - use TierRevered for no max)
	MaxTier ReputationTier

	// FailureMessage is shown when requirements aren't met
	FailureMessage string
}

// MeetsRequirement checks if a reputation value meets this requirement.
func (r *QuestReputationRequirement) MeetsRequirement(reputation float64) bool {
	tier := ReputationFromValue(reputation)
	return tier >= r.MinTier && tier <= r.MaxTier
}

// GatedQuest represents a quest with reputation requirements.
type GatedQuest struct {
	// QuestID is the unique identifier for this quest
	QuestID string

	// FactionID is the faction offering the quest
	FactionID string

	// Requirements are the reputation conditions that must be met
	Requirements []QuestReputationRequirement

	// UnlockMessage is shown when the quest becomes available
	UnlockMessage string

	// RewardReputationBonus is additional reputation gained on completion
	RewardReputationBonus float64

	// IsExclusive means completing this quest locks out opposing faction quests
	IsExclusive bool

	// ExcludesFactions lists factions whose quests become locked if this is completed
	ExcludesFactions []string
}

// ReputationQuestGatingComponent tracks faction-gated quests for an entity.
type ReputationQuestGatingComponent struct {
	// GatedQuests maps quest IDs to their reputation requirements
	GatedQuests map[string]*GatedQuest

	// UnlockedQuests tracks quests that have become available due to reputation
	UnlockedQuests map[string]bool

	// LockedOutQuests tracks quests that are permanently locked due to faction choices
	LockedOutQuests map[string]string // questID -> reason

	// FactionQuestProgress tracks how many faction-specific quests completed per faction
	FactionQuestProgress map[string]int

	// RecentUnlocks tracks recently unlocked quests for UI notification
	RecentUnlocks []string

	// RecentLockouts tracks recently locked quests for UI notification
	RecentLockouts []string
}

// Type returns the component type identifier.
func (r *ReputationQuestGatingComponent) Type() string {
	return "reputation_quest_gating"
}

// NewReputationQuestGatingComponent creates a new component.
func NewReputationQuestGatingComponent() *ReputationQuestGatingComponent {
	return &ReputationQuestGatingComponent{
		GatedQuests:          make(map[string]*GatedQuest),
		UnlockedQuests:       make(map[string]bool),
		LockedOutQuests:      make(map[string]string),
		FactionQuestProgress: make(map[string]int),
		RecentUnlocks:        make([]string, 0),
		RecentLockouts:       make([]string, 0),
	}
}

// RegisterGatedQuest adds a quest with reputation requirements.
func (r *ReputationQuestGatingComponent) RegisterGatedQuest(quest *GatedQuest) {
	if quest != nil && quest.QuestID != "" {
		r.GatedQuests[quest.QuestID] = quest
	}
}

// IsQuestAvailable checks if a quest can be offered based on current reputation.
func (r *ReputationQuestGatingComponent) IsQuestAvailable(questID string, factionReps map[string]float64) bool {
	// Check if permanently locked out
	if _, locked := r.LockedOutQuests[questID]; locked {
		return false
	}

	quest, exists := r.GatedQuests[questID]
	if !exists {
		// Quest not registered as gated, assume available
		return true
	}

	// Check all requirements
	for _, req := range quest.Requirements {
		rep, hasRep := factionReps[req.FactionID]
		if !hasRep {
			rep = 0.0 // Neutral by default
		}
		if !req.MeetsRequirement(rep) {
			return false
		}
	}

	return true
}

// GetQuestBlockReason returns why a quest is unavailable.
func (r *ReputationQuestGatingComponent) GetQuestBlockReason(questID string, factionReps map[string]float64) string {
	// Check lockout first
	if reason, locked := r.LockedOutQuests[questID]; locked {
		return reason
	}

	quest, exists := r.GatedQuests[questID]
	if !exists {
		return ""
	}

	// Find first unmet requirement
	for _, req := range quest.Requirements {
		rep, hasRep := factionReps[req.FactionID]
		if !hasRep {
			rep = 0.0
		}
		if !req.MeetsRequirement(rep) {
			if req.FailureMessage != "" {
				return req.FailureMessage
			}
			currentTier := ReputationFromValue(rep)
			return "Requires " + req.MinTier.String() + " standing with " + req.FactionID +
				" (current: " + currentTier.String() + ")"
		}
	}

	return ""
}

// MarkQuestUnlocked records that a quest has become available.
func (r *ReputationQuestGatingComponent) MarkQuestUnlocked(questID string) {
	if !r.UnlockedQuests[questID] {
		r.UnlockedQuests[questID] = true
		r.RecentUnlocks = append(r.RecentUnlocks, questID)
	}
}

// MarkQuestLockedOut permanently locks a quest with a reason.
func (r *ReputationQuestGatingComponent) MarkQuestLockedOut(questID, reason string) {
	if _, exists := r.LockedOutQuests[questID]; !exists {
		r.LockedOutQuests[questID] = reason
		r.RecentLockouts = append(r.RecentLockouts, questID)
	}
}

// RecordFactionQuestCompletion tracks completion and handles exclusivity.
func (r *ReputationQuestGatingComponent) RecordFactionQuestCompletion(questID string) []string {
	quest, exists := r.GatedQuests[questID]
	if !exists {
		return nil
	}

	// Track progress
	r.FactionQuestProgress[quest.FactionID]++

	// Handle exclusive quests
	var lockedOut []string
	if quest.IsExclusive {
		for _, excludedFaction := range quest.ExcludesFactions {
			// Lock out all quests from excluded factions
			for qid, gq := range r.GatedQuests {
				if gq.FactionID == excludedFaction {
					reason := "Locked due to completing " + quest.QuestID + " for " + quest.FactionID
					r.MarkQuestLockedOut(qid, reason)
					lockedOut = append(lockedOut, qid)
				}
			}
		}
	}

	return lockedOut
}

// GetRecentUnlocks returns and clears recent unlock notifications.
func (r *ReputationQuestGatingComponent) GetRecentUnlocks() []string {
	unlocks := r.RecentUnlocks
	r.RecentUnlocks = make([]string, 0)
	return unlocks
}

// GetRecentLockouts returns and clears recent lockout notifications.
func (r *ReputationQuestGatingComponent) GetRecentLockouts() []string {
	lockouts := r.RecentLockouts
	r.RecentLockouts = make([]string, 0)
	return lockouts
}

// GetFactionQuestCount returns completed quests for a faction.
func (r *ReputationQuestGatingComponent) GetFactionQuestCount(factionID string) int {
	return r.FactionQuestProgress[factionID]
}

// GetAllUnlockedQuests returns all currently unlocked quest IDs.
func (r *ReputationQuestGatingComponent) GetAllUnlockedQuests() []string {
	var unlocked []string
	for questID := range r.UnlockedQuests {
		unlocked = append(unlocked, questID)
	}
	return unlocked
}

// GetAllLockedOutQuests returns all permanently locked quest IDs with reasons.
func (r *ReputationQuestGatingComponent) GetAllLockedOutQuests() map[string]string {
	result := make(map[string]string)
	for k, v := range r.LockedOutQuests {
		result[k] = v
	}
	return result
}
