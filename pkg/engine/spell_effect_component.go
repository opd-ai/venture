package engine

// EffectType represents the type of spell effect.
type EffectType int

const (
	// EffectTerrainManipulation creates, modifies, or destroys terrain (walls, bridges, pits)
	EffectTerrainManipulation EffectType = iota
	// EffectTransmutation converts materials (stone→gold, water→ice)
	EffectTransmutation
	// EffectSummoning spawns temporary allies or objects
	EffectSummoning
	// EffectIllusion creates decoys, invisibility, or confusion
	EffectIllusion
	// EffectTimeManipulation slows, hastes, or rewinds positions
	EffectTimeManipulation
	// EffectGravityControl levitation, increased weight, orbital effects
	EffectGravityControl
	// EffectElementalFusion combines elements (fire+ice=steam, earth+lightning=glass)
	EffectElementalFusion
	// EffectLifeDrain transfers HP between entities
	EffectLifeDrain
	// EffectTeleportation short-range blink or long-range portal
	EffectTeleportation
	// EffectMetamagic enhances other spells (double damage, multi-target)
	EffectMetamagic
)

// String returns the string representation of an effect type.
func (e EffectType) String() string {
	switch e {
	case EffectTerrainManipulation:
		return "terrain_manipulation"
	case EffectTransmutation:
		return "transmutation"
	case EffectSummoning:
		return "summoning"
	case EffectIllusion:
		return "illusion"
	case EffectTimeManipulation:
		return "time_manipulation"
	case EffectGravityControl:
		return "gravity_control"
	case EffectElementalFusion:
		return "elemental_fusion"
	case EffectLifeDrain:
		return "life_drain"
	case EffectTeleportation:
		return "teleportation"
	case EffectMetamagic:
		return "metamagic"
	default:
		return "unknown"
	}
}

// TargetType represents what a spell effect can target.
type TargetType int

const (
	// TargetSelf affects only the caster
	TargetSelf TargetType = iota
	// TargetEntity affects a single entity
	TargetEntity
	// TargetArea affects all targets in an area
	TargetArea
	// TargetTerrain affects terrain tiles
	TargetTerrain
)

// String returns the string representation of a target type.
func (t TargetType) String() string {
	switch t {
	case TargetSelf:
		return "self"
	case TargetEntity:
		return "entity"
	case TargetArea:
		return "area"
	case TargetTerrain:
		return "terrain"
	default:
		return "unknown"
	}
}

// SpellEffectComponent represents an active spell effect in the world.
// This component stores the configuration and state of a spell effect,
// while SpellEffectSystem handles the actual execution logic.
type SpellEffectComponent struct {
	// EffectType identifies the kind of effect
	EffectType EffectType

	// Duration is how long the effect lasts in seconds (0 = instant)
	Duration float64

	// Magnitude is the effect strength (meaning depends on effect type)
	Magnitude float64

	// TargetType indicates what the effect can target
	TargetType TargetType

	// TerrainModifier specifies terrain changes (for terrain manipulation)
	TerrainModifier int

	// SummonTemplate specifies what to summon (for summoning effects)
	SummonTemplate string

	// FusionElements stores element IDs for fusion effects (comma-separated)
	FusionElements string

	// MetamagicMultiplier is the damage/effect multiplier (for metamagic)
	MetamagicMultiplier float64

	// CasterID is the entity that cast this effect
	CasterID uint64

	// TargetID is the entity being affected (0 = area effect)
	TargetID uint64

	// TargetX and TargetY are the world coordinates of the effect center
	TargetX float64
	TargetY float64

	// Radius is the effect area size (for area effects)
	Radius float64

	// Active indicates if the effect is currently executing
	Active bool

	// ElapsedTime tracks how long the effect has been active
	ElapsedTime float64
}

// Type returns the component type identifier.
func (s *SpellEffectComponent) Type() string {
	return "spell_effect"
}

// IsExpired returns true if the effect duration has expired.
func (s *SpellEffectComponent) IsExpired() bool {
	if s.Duration <= 0 {
		return true // Instant effect
	}
	return s.ElapsedTime >= s.Duration
}

// Update updates the effect elapsed time.
func (s *SpellEffectComponent) Update(deltaTime float64) {
	s.ElapsedTime += deltaTime
}

// GetProgress returns the effect progress from 0.0 to 1.0.
func (s *SpellEffectComponent) GetProgress() float64 {
	if s.Duration <= 0 {
		return 1.0
	}
	progress := s.ElapsedTime / s.Duration
	if progress > 1.0 {
		progress = 1.0
	}
	return progress
}
