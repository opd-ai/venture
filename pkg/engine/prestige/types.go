package prestige

import (
	"encoding/json"
	"time"
)

// ParagonStat represents a stat that can receive paragon points.
type ParagonStat int

const (
	// StatHealth increases maximum health
	StatHealth ParagonStat = iota
	// StatDamage increases damage output
	StatDamage
	// StatDefense increases damage reduction
	StatDefense
	// StatSpeed increases movement and attack speed
	StatSpeed
	// StatCritical increases critical hit chance
	StatCritical
)

// String returns the string representation of a paragon stat.
func (s ParagonStat) String() string {
	switch s {
	case StatHealth:
		return "Health"
	case StatDamage:
		return "Damage"
	case StatDefense:
		return "Defense"
	case StatSpeed:
		return "Speed"
	case StatCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// VisualTier represents the visual effect tier based on prestige level.
type VisualTier int

const (
	// VisualNone represents no visual effect (prestige 1-9)
	VisualNone VisualTier = iota
	// VisualSubtle represents subtle glow (prestige 10-24)
	VisualSubtle
	// VisualModerate represents brighter glow with trail (prestige 25-49)
	VisualModerate
	// VisualIntense represents intense glow with aura (prestige 50-99)
	VisualIntense
	// VisualRadiant represents radiant aura (prestige 100+)
	VisualRadiant
)

// String returns the string representation of a visual tier.
func (v VisualTier) String() string {
	switch v {
	case VisualNone:
		return "None"
	case VisualSubtle:
		return "Subtle"
	case VisualModerate:
		return "Moderate"
	case VisualIntense:
		return "Intense"
	case VisualRadiant:
		return "Radiant"
	default:
		return "Unknown"
	}
}

// PrestigeAbility represents a class-specific prestige ability.
type PrestigeAbility struct {
	// Name is the ability name
	Name string
	// Description explains the ability effect
	Description string
	// UnlockLevel is the prestige level required
	UnlockLevel int
	// ClassName is the class that unlocks this ability
	ClassName string
	// Cooldown in seconds
	Cooldown int
	// ManaCost for using the ability
	ManaCost int
}

// PlayerPrestige tracks a single player's prestige progression.
type PlayerPrestige struct {
	// PlayerID is the unique player identifier
	PlayerID string
	// ClassName is the player's class
	ClassName string
	// PrestigeLevel is the current prestige level
	PrestigeLevel int
	// CurrentXP towards next prestige level
	CurrentXP int
	// TotalXP earned across all prestige levels
	TotalXP int
	// ParagonPoints available to spend
	ParagonPoints int
	// ParagonAllocations maps stat to allocated points
	ParagonAllocations map[ParagonStat]int
	// UnlockedAbilities lists prestige abilities unlocked
	UnlockedAbilities []int
	// LastUpdated timestamp
	LastUpdated time.Time
}

// AccountPrestige tracks account-wide prestige bonuses.
type AccountPrestige struct {
	// AccountID is the unique account identifier
	AccountID string
	// Prestige100Count is number of characters at prestige 100+
	Prestige100Count int
	// XPBonus is the total account-wide XP bonus
	XPBonus float64
	// CharacterIDs lists all characters on account
	CharacterIDs []string
	// LastUpdated timestamp
	LastUpdated time.Time
}

// PrestigeData wraps all prestige data for persistence.
type PrestigeData struct {
	// Players maps playerID to prestige data
	Players map[string]*PlayerPrestige
	// Accounts maps accountID to account data
	Accounts map[string]*AccountPrestige
}

// PrestigeComponent is the ECS component for prestige system.
type PrestigeComponent struct {
	// PlayerID links to PlayerPrestige
	PlayerID string
	// PrestigeLevel for quick access
	PrestigeLevel int
	// VisualTier for rendering
	VisualTier VisualTier
	// ActiveAbilities lists currently active prestige abilities
	ActiveAbilities []string
}

// Type returns the component type identifier.
func (p PrestigeComponent) Type() string {
	return "prestige"
}

// Serialize converts the prestige component to bytes for persistence.
func (p *PrestigeComponent) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize restores the prestige component from bytes.
func (p *PrestigeComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, p)
}

// Constants for prestige system configuration
const (
	// BasePrestigeXP is the XP required for prestige level 1
	BasePrestigeXP = 100000
	// ParagonPointBonus is the stat increase per point (0.1% = 0.001 multiplier)
	ParagonPointBonus = 0.001
	// RespecCostPerPoint is the gold cost to respec one paragon point
	RespecCostPerPoint = 1000
	// AccountXPBonus is the XP bonus per prestige 100 character (5% = 0.05)
	AccountXPBonus = 0.05
)

// MarshalJSON implements custom JSON marshaling.
func (p *PlayerPrestige) MarshalJSON() ([]byte, error) {
	type Alias PlayerPrestige
	return json.Marshal(&struct {
		*Alias
		LastUpdatedUnix int64 `json:"last_updated_unix"`
	}{
		Alias:           (*Alias)(p),
		LastUpdatedUnix: p.LastUpdated.Unix(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling.
func (p *PlayerPrestige) UnmarshalJSON(data []byte) error {
	type Alias PlayerPrestige
	aux := &struct {
		*Alias
		LastUpdatedUnix int64 `json:"last_updated_unix"`
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.LastUpdated = time.Unix(aux.LastUpdatedUnix, 0)
	return nil
}

// MarshalJSON implements custom JSON marshaling for AccountPrestige.
func (a *AccountPrestige) MarshalJSON() ([]byte, error) {
	type Alias AccountPrestige
	return json.Marshal(&struct {
		*Alias
		LastUpdatedUnix int64 `json:"last_updated_unix"`
	}{
		Alias:           (*Alias)(a),
		LastUpdatedUnix: a.LastUpdated.Unix(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for AccountPrestige.
func (a *AccountPrestige) UnmarshalJSON(data []byte) error {
	type Alias AccountPrestige
	aux := &struct {
		*Alias
		LastUpdatedUnix int64 `json:"last_updated_unix"`
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	a.LastUpdated = time.Unix(aux.LastUpdatedUnix, 0)
	return nil
}
