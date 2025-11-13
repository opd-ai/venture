package engine

import "time"

// MoralChoiceComponent tracks pending and historical moral choices for an entity.
// It manages decision points in quests, faction conflicts, and other scenarios where
// player choices have alignment or reputation consequences.
type MoralChoiceComponent struct {
	// PendingChoices are decisions awaiting player input
	PendingChoices []MoralChoice

	// ChoiceHistory records all completed moral choices
	ChoiceHistory []CompletedChoice

	// ActiveRedemptions tracks ongoing redemption arcs for regaining lost reputation
	ActiveRedemptions []RedemptionArc
}

// Type returns the component type identifier.
func (m MoralChoiceComponent) Type() string {
	return "moral_choice"
}

// MoralChoice represents a decision point with multiple options and consequences.
type MoralChoice struct {
	// ID uniquely identifies this choice instance
	ID string

	// Description is the situation or dilemma presented
	Description string

	// Context provides additional information (quest name, location, NPCs involved)
	Context string

	// Options are the available choices
	Options []ChoiceOption

	// QuestID links this choice to a quest (optional)
	QuestID string

	// TimeOffered records when this choice became available
	TimeOffered time.Time

	// ExpiresAt is when this choice expires (optional, zero value means no expiry)
	ExpiresAt time.Time
}

// ChoiceOption represents one possible decision with its consequences.
type ChoiceOption struct {
	// Label is the short description shown to the player
	Label string

	// Description provides more detail about this option
	Description string

	// AlignmentImpact affects the character's law/good axes
	AlignmentImpact AlignmentDelta

	// ReputationImpact affects faction standings
	ReputationImpact map[string]float64

	// Rewards are optional benefits for choosing this option
	Rewards *ChoiceRewards

	// Consequences are optional penalties for choosing this option
	Consequences *ChoiceConsequences
}

// AlignmentDelta represents changes to alignment axes.
type AlignmentDelta struct {
	// LawDelta is the change to law axis (-1.0 to +1.0)
	LawDelta float64

	// GoodDelta is the change to good axis (-1.0 to +1.0)
	GoodDelta float64
}

// ChoiceRewards defines benefits for making a choice.
type ChoiceRewards struct {
	// XP is experience points awarded
	XP int

	// Gold is currency awarded
	Gold int

	// Items are item IDs or types awarded
	Items []string

	// UnlockQuest is a quest ID unlocked by this choice (optional)
	UnlockQuest string
}

// ChoiceConsequences defines negative outcomes for a choice.
type ChoiceConsequences struct {
	// HostileFactions lists factions that become hostile
	HostileFactions []string

	// LoseQuests lists quest IDs that become unavailable
	LoseQuests []string

	// LoseItems lists item IDs that are removed
	LoseItems []string

	// SpawnEnemies is the number of hostile entities to spawn
	SpawnEnemies int
}

// CompletedChoice records a choice that was made and its outcome.
type CompletedChoice struct {
	// ChoiceID references the original MoralChoice ID
	ChoiceID string

	// Description is the situation that was resolved
	Description string

	// SelectedOption is the index of the chosen option
	SelectedOption int

	// OptionLabel is the label of the chosen option
	OptionLabel string

	// Timestamp records when the choice was made
	Timestamp time.Time

	// AlignmentChange records the alignment impact applied
	AlignmentChange AlignmentDelta

	// ReputationChanges records faction reputation changes applied
	ReputationChanges map[string]float64

	// QuestID links this to a quest (optional)
	QuestID string
}

// RedemptionArc represents an ongoing effort to regain lost reputation with a faction.
// Players can work to improve their standing through specific actions.
type RedemptionArc struct {
	// FactionName is the faction being redeemed with
	FactionName string

	// StartingReputation is the reputation value when redemption began
	StartingReputation float64

	// TargetReputation is the goal reputation value
	TargetReputation float64

	// CurrentReputation tracks progress
	CurrentReputation float64

	// RequiredActions describes what must be done
	RequiredActions []RedemptionAction

	// CompletedActions tracks which actions have been done
	CompletedActions int

	// StartTime records when redemption began
	StartTime time.Time

	// TimeLimit is optional deadline for completion (zero value means no limit)
	TimeLimit time.Time
}

// RedemptionAction defines a specific task required for redemption.
type RedemptionAction struct {
	// Type categorizes the action (Kill, Deliver, Donate, etc.)
	Type string

	// Description explains what to do
	Description string

	// Target specifies what/who (enemy type, item type, etc.)
	Target string

	// Quantity is how many times to perform this action
	Quantity int

	// Progress tracks current completion
	Progress int

	// ReputationGain is the reputation awarded upon completion
	ReputationGain float64
}

// NewMoralChoiceComponent creates an empty moral choice component.
func NewMoralChoiceComponent() *MoralChoiceComponent {
	return &MoralChoiceComponent{
		PendingChoices:    make([]MoralChoice, 0),
		ChoiceHistory:     make([]CompletedChoice, 0),
		ActiveRedemptions: make([]RedemptionArc, 0),
	}
}

// AddChoice presents a new moral choice to the entity.
func (m *MoralChoiceComponent) AddChoice(choice MoralChoice) {
	// Set time offered if not already set
	if choice.TimeOffered.IsZero() {
		choice.TimeOffered = time.Now()
	}

	m.PendingChoices = append(m.PendingChoices, choice)
}

// GetPendingChoice retrieves a pending choice by ID.
// Returns nil if not found.
func (m *MoralChoiceComponent) GetPendingChoice(id string) *MoralChoice {
	for i := range m.PendingChoices {
		if m.PendingChoices[i].ID == id {
			return &m.PendingChoices[i]
		}
	}
	return nil
}

// RemovePendingChoice removes a choice from the pending list.
func (m *MoralChoiceComponent) RemovePendingChoice(id string) bool {
	for i, choice := range m.PendingChoices {
		if choice.ID == id {
			// Remove by swapping with last element and truncating
			m.PendingChoices[i] = m.PendingChoices[len(m.PendingChoices)-1]
			m.PendingChoices = m.PendingChoices[:len(m.PendingChoices)-1]
			return true
		}
	}
	return false
}

// RecordChoice saves a completed choice to history.
func (m *MoralChoiceComponent) RecordChoice(completed CompletedChoice) {
	if completed.Timestamp.IsZero() {
		completed.Timestamp = time.Now()
	}
	m.ChoiceHistory = append(m.ChoiceHistory, completed)
}

// GetRecentChoices returns the N most recent completed choices.
func (m *MoralChoiceComponent) GetRecentChoices(count int) []CompletedChoice {
	if count <= 0 {
		return []CompletedChoice{}
	}

	total := len(m.ChoiceHistory)
	if total == 0 {
		return []CompletedChoice{}
	}

	if count > total {
		count = total
	}

	return m.ChoiceHistory[total-count:]
}

// StartRedemption initiates a redemption arc for a faction.
func (m *MoralChoiceComponent) StartRedemption(arc RedemptionArc) {
	if arc.StartTime.IsZero() {
		arc.StartTime = time.Now()
	}
	m.ActiveRedemptions = append(m.ActiveRedemptions, arc)
}

// GetRedemptionArc retrieves an active redemption arc by faction name.
// Returns nil if no redemption is active for that faction.
func (m *MoralChoiceComponent) GetRedemptionArc(factionName string) *RedemptionArc {
	for i := range m.ActiveRedemptions {
		if m.ActiveRedemptions[i].FactionName == factionName {
			return &m.ActiveRedemptions[i]
		}
	}
	return nil
}

// RemoveRedemptionArc removes a completed or failed redemption arc.
func (m *MoralChoiceComponent) RemoveRedemptionArc(factionName string) bool {
	for i, arc := range m.ActiveRedemptions {
		if arc.FactionName == factionName {
			m.ActiveRedemptions[i] = m.ActiveRedemptions[len(m.ActiveRedemptions)-1]
			m.ActiveRedemptions = m.ActiveRedemptions[:len(m.ActiveRedemptions)-1]
			return true
		}
	}
	return false
}

// HasActiveMoralChoices returns true if there are any pending choices.
func (m *MoralChoiceComponent) HasActiveMoralChoices() bool {
	return len(m.PendingChoices) > 0
}

// HasActiveRedemptions returns true if there are any ongoing redemption arcs.
func (m *MoralChoiceComponent) HasActiveRedemptions() bool {
	return len(m.ActiveRedemptions) > 0
}

// IsComplete checks if a redemption action is complete.
func (r *RedemptionAction) IsComplete() bool {
	return r.Progress >= r.Quantity
}

// GetProgress returns the completion percentage (0.0 to 1.0).
func (r *RedemptionAction) GetProgress() float64 {
	if r.Quantity <= 0 {
		return 1.0
	}
	progress := float64(r.Progress) / float64(r.Quantity)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// IsComplete checks if a redemption arc is complete.
func (r *RedemptionArc) IsComplete() bool {
	return r.CompletedActions >= len(r.RequiredActions)
}

// GetProgress returns the overall completion percentage (0.0 to 1.0).
func (r *RedemptionArc) GetProgress() float64 {
	if len(r.RequiredActions) == 0 {
		return 1.0
	}

	totalProgress := 0.0
	for _, action := range r.RequiredActions {
		totalProgress += action.GetProgress()
	}

	progress := totalProgress / float64(len(r.RequiredActions))
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// IsExpired checks if a redemption arc has exceeded its time limit.
func (r *RedemptionArc) IsExpired() bool {
	if r.TimeLimit.IsZero() {
		return false
	}
	return time.Now().After(r.TimeLimit)
}

// IsExpired checks if a moral choice has exceeded its expiration time.
func (m *MoralChoice) IsExpired() bool {
	if m.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(m.ExpiresAt)
}
