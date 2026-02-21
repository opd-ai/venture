// Package engine provides the FactionTerritoryInfluenceComponent for tracking
// faction control over territory zones and their gameplay effects.
package engine

// FactionTerritoryInfluenceComponent tracks faction influence zones in the world.
// Territories controlled by factions provide buffs/debuffs based on player reputation.
// This is a pure data component - all logic is in FactionTerritoryInfluenceSystem.
type FactionTerritoryInfluenceComponent struct {
	// FactionID identifies which faction controls this territory zone
	FactionID string

	// InfluenceStrength represents how strong faction control is (0.0 to 1.0)
	// Higher strength = stronger bonus/penalty effects
	InfluenceStrength float64

	// ZoneX and ZoneZ identify the territory grid coordinates
	ZoneX int
	ZoneZ int

	// ZoneRadius is the radius of influence in world units
	ZoneRadius float64

	// EffectMultiplier scales all influence effects (default 1.0)
	EffectMultiplier float64

	// Bonuses applied to friendly entities in this zone
	FriendlyDamageBonus  float64 // Percentage damage increase (0.0 to 0.5)
	FriendlyDefenseBonus float64 // Percentage damage reduction (0.0 to 0.3)
	FriendlyXPBonus      float64 // Percentage XP increase (0.0 to 0.25)
	FriendlyHealingBonus float64 // Percentage healing increase (0.0 to 0.2)

	// Penalties applied to hostile entities in this zone
	HostileDamagePenalty  float64 // Percentage damage reduction (0.0 to 0.3)
	HostileDefensePenalty float64 // Percentage damage taken increase (0.0 to 0.25)
	HostileDetectionBonus float64 // Enemy detection range increase (0.0 to 0.5)

	// IsContested indicates if another faction is challenging control
	IsContested bool

	// ContestingFactionID is the faction attempting to take over (if contested)
	ContestingFactionID string

	// ContestProgress tracks capture progress (0.0 to 1.0)
	ContestProgress float64
}

// Type returns the component type identifier for ECS registration.
func (c *FactionTerritoryInfluenceComponent) Type() string {
	return "faction_territory_influence"
}

// NewFactionTerritoryInfluenceComponent creates an influence zone with default values.
func NewFactionTerritoryInfluenceComponent(factionID string, zoneX, zoneZ int) *FactionTerritoryInfluenceComponent {
	return &FactionTerritoryInfluenceComponent{
		FactionID:             factionID,
		InfluenceStrength:     1.0,
		ZoneX:                 zoneX,
		ZoneZ:                 zoneZ,
		ZoneRadius:            256.0, // Default 256 units radius
		EffectMultiplier:      1.0,
		FriendlyDamageBonus:   0.15, // 15% damage boost
		FriendlyDefenseBonus:  0.10, // 10% damage reduction
		FriendlyXPBonus:       0.10, // 10% XP bonus
		FriendlyHealingBonus:  0.10, // 10% healing bonus
		HostileDamagePenalty:  0.10, // 10% damage reduction
		HostileDefensePenalty: 0.15, // 15% more damage taken
		HostileDetectionBonus: 0.25, // 25% detection range increase
		IsContested:           false,
	}
}

// FactionTerritoryModifierComponent is applied to entities in faction zones.
// This tracks active territory-based modifiers on the entity.
type FactionTerritoryModifierComponent struct {
	// ActiveFactionID is the faction zone the entity is currently in
	ActiveFactionID string

	// ReputationLevel with the controlling faction (-100 to 100)
	ReputationLevel int

	// EffectiveDamageModifier is the current damage multiplier
	EffectiveDamageModifier float64

	// EffectiveDefenseModifier is the current defense multiplier
	EffectiveDefenseModifier float64

	// EffectiveXPModifier is the current XP multiplier
	EffectiveXPModifier float64

	// EffectiveHealingModifier is the current healing multiplier
	EffectiveHealingModifier float64

	// EffectiveDetectionModifier is the current detection range multiplier
	EffectiveDetectionModifier float64

	// InFactionZone indicates if entity is currently in any faction zone
	InFactionZone bool

	// ZoneX, ZoneZ store current zone coordinates for change detection
	ZoneX int
	ZoneZ int

	// Dirty indicates modifiers need recalculation
	Dirty bool
}

// Type returns the component type identifier for ECS registration.
func (c *FactionTerritoryModifierComponent) Type() string {
	return "faction_territory_modifier"
}

// NewFactionTerritoryModifierComponent creates a modifier component with neutral defaults.
func NewFactionTerritoryModifierComponent() *FactionTerritoryModifierComponent {
	return &FactionTerritoryModifierComponent{
		EffectiveDamageModifier:    1.0,
		EffectiveDefenseModifier:   1.0,
		EffectiveXPModifier:        1.0,
		EffectiveHealingModifier:   1.0,
		EffectiveDetectionModifier: 1.0,
		InFactionZone:              false,
		Dirty:                      true,
	}
}

// Reset clears all modifiers to neutral values.
func (c *FactionTerritoryModifierComponent) Reset() {
	c.ActiveFactionID = ""
	c.ReputationLevel = 0
	c.EffectiveDamageModifier = 1.0
	c.EffectiveDefenseModifier = 1.0
	c.EffectiveXPModifier = 1.0
	c.EffectiveHealingModifier = 1.0
	c.EffectiveDetectionModifier = 1.0
	c.InFactionZone = false
	c.ZoneX = 0
	c.ZoneZ = 0
	c.Dirty = false
}

// IsFriendly returns true if reputation indicates friendly standing.
func (c *FactionTerritoryModifierComponent) IsFriendly() bool {
	return c.ReputationLevel >= 51
}

// IsHostile returns true if reputation indicates hostile standing.
func (c *FactionTerritoryModifierComponent) IsHostile() bool {
	return c.ReputationLevel <= -50
}

// IsNeutral returns true if reputation is neither friendly nor hostile.
func (c *FactionTerritoryModifierComponent) IsNeutral() bool {
	return c.ReputationLevel > -50 && c.ReputationLevel < 51
}
