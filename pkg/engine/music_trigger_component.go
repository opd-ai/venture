package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/audio"
)

// MusicTriggerComponent tracks events that should trigger music context changes.
// This component enables adaptive music that responds to gameplay events.
type MusicTriggerComponent struct {
	// CurrentContext is the active music context
	CurrentContext audio.MusicContext

	// PendingContext is a context waiting to be applied
	PendingContext *audio.MusicContext

	// TransitionTime is how long to wait before applying pending context
	TransitionTime float64

	// LastCombatTime tracks when combat last occurred
	LastCombatTime time.Time

	// LastBossTime tracks when a boss was last encountered
	LastBossTime time.Time

	// LastQuestComplete tracks the last quest completion
	LastQuestComplete time.Time

	// LastMilestone tracks the last exploration milestone
	LastMilestone time.Time

	// CombatActive indicates if currently in combat
	CombatActive bool

	// BossNearby indicates if a boss is nearby
	BossNearby bool

	// ExplorationMilestones counts discovered areas
	ExplorationMilestones int

	// ReputationTier tracks the player's overall reputation level
	ReputationTier string // "hated", "neutral", "honored", "revered"
}

// Type returns the component type identifier.
func (m *MusicTriggerComponent) Type() string {
	return "music_trigger"
}

// TriggerType represents different types of music triggers.
type TriggerType int

const (
	// TriggerCombatStart signals the beginning of combat
	TriggerCombatStart TriggerType = iota
	// TriggerCombatEnd signals the end of combat
	TriggerCombatEnd
	// TriggerBossAppear signals a boss has appeared
	TriggerBossAppear
	// TriggerBossDefeated signals a boss has been defeated
	TriggerBossDefeated
	// TriggerQuestComplete signals quest completion
	TriggerQuestComplete
	// TriggerExplorationMilestone signals discovering a new area
	TriggerExplorationMilestone
	// TriggerReputationChange signals significant reputation shift
	TriggerReputationChange
)

// String returns the string representation of a TriggerType.
func (t TriggerType) String() string {
	switch t {
	case TriggerCombatStart:
		return "combat_start"
	case TriggerCombatEnd:
		return "combat_end"
	case TriggerBossAppear:
		return "boss_appear"
	case TriggerBossDefeated:
		return "boss_defeated"
	case TriggerQuestComplete:
		return "quest_complete"
	case TriggerExplorationMilestone:
		return "exploration_milestone"
	case TriggerReputationChange:
		return "reputation_change"
	default:
		return "unknown"
	}
}

// MusicTriggerEvent represents a music trigger event with context.
type MusicTriggerEvent struct {
	Type      TriggerType
	Timestamp time.Time
	EntityID  uint64 // Entity that triggered the event
	Data      map[string]interface{}
}

// NewMusicTriggerComponent creates a new music trigger component.
func NewMusicTriggerComponent() *MusicTriggerComponent {
	return &MusicTriggerComponent{
		CurrentContext: audio.MusicContext{
			Location:  "exploration",
			Combat:    false,
			Danger:    0.0,
			TimeOfDay: "day",
		},
		CombatActive:          false,
		BossNearby:            false,
		ExplorationMilestones: 0,
		ReputationTier:        "neutral",
	}
}

// TriggerCombat updates the music context for combat.
func (m *MusicTriggerComponent) TriggerCombat(active bool) {
	m.CombatActive = active
	m.CurrentContext.Combat = active

	if active {
		m.LastCombatTime = time.Now()
		m.CurrentContext.Danger = 0.6 // Moderate danger during normal combat
		if m.BossNearby {
			m.CurrentContext.Danger = 1.0 // Maximum danger with boss
		}
	} else {
		m.CurrentContext.Danger = 0.2 // Slight tension after combat
	}
}

// TriggerBoss updates the music context for boss encounters.
func (m *MusicTriggerComponent) TriggerBoss(present bool) {
	m.BossNearby = present
	m.CurrentContext.BossNearby = present

	if present {
		m.LastBossTime = time.Now()
		m.CurrentContext.Danger = 1.0
		m.CurrentContext.Combat = true
		m.CombatActive = true
	}
}

// TriggerQuestCompletion creates a victory context.
func (m *MusicTriggerComponent) TriggerQuestCompletion() {
	m.LastQuestComplete = time.Now()

	// Temporarily set to victory context
	m.PendingContext = &audio.MusicContext{
		Location:  "victory",
		Combat:    false,
		Danger:    0.0,
		TimeOfDay: m.CurrentContext.TimeOfDay,
	}
	m.TransitionTime = 5.0 // Play victory music for 5 seconds
}

// TriggerExploration updates context for exploration milestones.
func (m *MusicTriggerComponent) TriggerExploration(newArea bool) {
	if newArea {
		m.ExplorationMilestones++
		m.LastMilestone = time.Now()
	}

	// Exploration context
	if !m.CombatActive && !m.BossNearby {
		m.CurrentContext.Location = "exploration"
		m.CurrentContext.Combat = false
		m.CurrentContext.Danger = 0.1
	}
}

// TriggerReputationChange updates context based on reputation.
func (m *MusicTriggerComponent) TriggerReputationChange(tier string) {
	m.ReputationTier = tier

	// Reputation affects ambient danger level when not in combat
	if !m.CombatActive {
		switch tier {
		case "hated":
			m.CurrentContext.Danger = 0.5 // Higher tension when hated
		case "hostile":
			m.CurrentContext.Danger = 0.4
		case "unfriendly":
			m.CurrentContext.Danger = 0.3
		case "neutral":
			m.CurrentContext.Danger = 0.1
		case "friendly":
			m.CurrentContext.Danger = 0.05
		case "honored":
			m.CurrentContext.Danger = 0.0
		case "revered":
			m.CurrentContext.Danger = 0.0
		}
	}
}

// GetMusicContext returns the current music context.
func (m *MusicTriggerComponent) GetMusicContext() audio.MusicContext {
	return m.CurrentContext
}

// UpdatePendingTransition handles pending context transitions.
func (m *MusicTriggerComponent) UpdatePendingTransition(deltaTime float64) {
	if m.PendingContext != nil {
		m.TransitionTime -= deltaTime
		if m.TransitionTime <= 0 {
			// Restore previous context
			m.PendingContext = nil
			// Return to exploration if not in combat
			if !m.CombatActive {
				m.TriggerExploration(false)
			}
		} else {
			// Apply pending context
			m.CurrentContext = *m.PendingContext
		}
	}
}
