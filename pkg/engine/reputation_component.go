package engine

import "time"

// ReputationComponent tracks a player's standing with various factions and their moral alignment.
// It stores faction reputation values, alignment on law/good axes, and a history of significant
// deeds that influenced the character's reputation.
type ReputationComponent struct {
	// Factions maps faction names to reputation values (-100 to +100).
	// Negative values indicate hostility, positive values indicate favor.
	Factions map[string]float64

	// Alignment represents the character's moral and ethical position
	Alignment Alignment

	// KarmaDeeds tracks significant actions that affected reputation or alignment
	KarmaDeeds []Deed
}

// Type returns the component type identifier.
func (r ReputationComponent) Type() string {
	return "reputation"
}

// Alignment represents a character's ethical and moral position on two axes.
// The law axis ranges from Chaotic (-1) to Lawful (+1).
// The good axis ranges from Evil (-1) to Good (+1).
type Alignment struct {
	// LawAxis represents lawful vs chaotic (-1.0 = chaotic, 0 = neutral, +1.0 = lawful)
	LawAxis float64

	// GoodAxis represents good vs evil (-1.0 = evil, 0 = neutral, +1.0 = good)
	GoodAxis float64
}

// String returns a human-readable alignment description (e.g., "Lawful Good", "Neutral Evil").
func (a Alignment) String() string {
	law := "Neutral"
	if a.LawAxis > 0.33 {
		law = "Lawful"
	} else if a.LawAxis < -0.33 {
		law = "Chaotic"
	}

	good := "Neutral"
	if a.GoodAxis > 0.33 {
		good = "Good"
	} else if a.GoodAxis < -0.33 {
		good = "Evil"
	}

	// Special case for true neutral
	if law == "Neutral" && good == "Neutral" {
		return "True Neutral"
	}

	return law + " " + good
}

// Deed represents a significant action that affected reputation or alignment.
// Each deed records what was done, its impact on factions/alignment, and when it occurred.
type Deed struct {
	// Description is a human-readable description of the action
	Description string

	// Timestamp records when the deed occurred
	Timestamp time.Time

	// FactionImpact maps faction names to reputation changes caused by this deed
	FactionImpact map[string]float64

	// LawImpact is the change to the law axis (-1.0 to +1.0)
	LawImpact float64

	// GoodImpact is the change to the good axis (-1.0 to +1.0)
	GoodImpact float64

	// Location is an optional field indicating where the deed occurred
	Location string
}

// NewReputationComponent creates a new ReputationComponent with neutral alignment
// and no faction standings.
func NewReputationComponent() *ReputationComponent {
	return &ReputationComponent{
		Factions: make(map[string]float64),
		Alignment: Alignment{
			LawAxis:  0.0,
			GoodAxis: 0.0,
		},
		KarmaDeeds: make([]Deed, 0),
	}
}

// GetReputation returns the current reputation with a faction.
// If the faction is unknown, returns 0 (neutral).
func (r *ReputationComponent) GetReputation(faction string) float64 {
	if rep, ok := r.Factions[faction]; ok {
		return rep
	}
	return 0.0
}

// SetReputation sets the reputation with a faction, clamping to [-100, +100].
func (r *ReputationComponent) SetReputation(faction string, value float64) {
	// Clamp to valid range
	if value > 100.0 {
		value = 100.0
	} else if value < -100.0 {
		value = -100.0
	}
	r.Factions[faction] = value
}

// AdjustReputation modifies reputation with a faction by the given delta.
// The final value is clamped to [-100, +100].
func (r *ReputationComponent) AdjustReputation(faction string, delta float64) {
	current := r.GetReputation(faction)
	r.SetReputation(faction, current+delta)
}

// GetReputationTier returns a string describing the reputation level with a faction.
// Tiers: Revered (75+), Honored (50+), Friendly (25+), Neutral (-24.99 to +24.99),
// Unfriendly (-49.99 to -25), Hostile (-74.99 to -50), Hated (<-75).
func (r *ReputationComponent) GetReputationTier(faction string) string {
	rep := r.GetReputation(faction)

	if rep >= 75.0 {
		return "Revered"
	} else if rep >= 50.0 {
		return "Honored"
	} else if rep >= 25.0 {
		return "Friendly"
	} else if rep > -25.0 {
		return "Neutral"
	} else if rep > -50.0 {
		return "Unfriendly"
	} else if rep > -75.0 {
		return "Hostile"
	}
	return "Hated"
}

// IsHostile returns true if the reputation with a faction is at or below -50 (Hostile or Hated).
func (r *ReputationComponent) IsHostile(faction string) bool {
	return r.GetReputation(faction) <= -50.0
}

// IsFriendly returns true if the reputation with a faction is above 25 (Friendly or better).
func (r *ReputationComponent) IsFriendly(faction string) bool {
	return r.GetReputation(faction) >= 25.0
}

// AdjustAlignment modifies the alignment axes by the given deltas, clamping to [-1, +1].
func (r *ReputationComponent) AdjustAlignment(lawDelta, goodDelta float64) {
	r.Alignment.LawAxis += lawDelta
	r.Alignment.GoodAxis += goodDelta

	// Clamp to valid range
	if r.Alignment.LawAxis > 1.0 {
		r.Alignment.LawAxis = 1.0
	} else if r.Alignment.LawAxis < -1.0 {
		r.Alignment.LawAxis = -1.0
	}

	if r.Alignment.GoodAxis > 1.0 {
		r.Alignment.GoodAxis = 1.0
	} else if r.Alignment.GoodAxis < -1.0 {
		r.Alignment.GoodAxis = -1.0
	}
}

// RecordDeed adds a deed to the karma history and applies its impacts to reputation and alignment.
func (r *ReputationComponent) RecordDeed(deed Deed) {
	// Set timestamp if not already set
	if deed.Timestamp.IsZero() {
		deed.Timestamp = time.Now()
	}

	// Apply faction impacts
	for faction, impact := range deed.FactionImpact {
		r.AdjustReputation(faction, impact)
	}

	// Apply alignment impacts
	r.AdjustAlignment(deed.LawImpact, deed.GoodImpact)

	// Add to history
	r.KarmaDeeds = append(r.KarmaDeeds, deed)
}

// GetRecentDeeds returns the N most recent deeds.
func (r *ReputationComponent) GetRecentDeeds(count int) []Deed {
	if count <= 0 {
		return []Deed{}
	}

	totalDeeds := len(r.KarmaDeeds)
	if totalDeeds == 0 {
		return []Deed{}
	}

	if count > totalDeeds {
		count = totalDeeds
	}

	// Return last N deeds
	return r.KarmaDeeds[totalDeeds-count:]
}
