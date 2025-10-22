// Package magic provides magic type definitions.
// This file defines spell types, elements, targeting, and effect types
// used by the magic generator.
package magic

// SpellType represents the classification of a spell.
type SpellType int

const (
	// TypeOffensive represents damage-dealing spells
	TypeOffensive SpellType = iota
	// TypeDefensive represents protective and shielding spells
	TypeDefensive
	// TypeUtility represents non-combat spells (teleport, light, etc.)
	TypeUtility
	// TypeHealing represents health restoration spells
	TypeHealing
	// TypeBuff represents stat-boosting spells
	TypeBuff
	// TypeDebuff represents stat-reducing spells on enemies
	TypeDebuff
	// TypeSummon represents spells that summon entities
	TypeSummon
)

// String returns the string representation of a spell type.
func (t SpellType) String() string {
	switch t {
	case TypeOffensive:
		return "offensive"
	case TypeDefensive:
		return "defensive"
	case TypeUtility:
		return "utility"
	case TypeHealing:
		return "healing"
	case TypeBuff:
		return "buff"
	case TypeDebuff:
		return "debuff"
	case TypeSummon:
		return "summon"
	default:
		return "unknown"
	}
}

// ElementType represents the elemental affinity of a spell.
type ElementType int

const (
	// ElementNone represents non-elemental magic
	ElementNone ElementType = iota
	// ElementFire represents fire-based spells
	ElementFire
	// ElementIce represents ice and cold spells
	ElementIce
	// ElementLightning represents electric spells
	ElementLightning
	// ElementEarth represents earth and stone spells
	ElementEarth
	// ElementWind represents air and wind spells
	ElementWind
	// ElementLight represents holy/light spells
	ElementLight
	// ElementDark represents shadow/dark spells
	ElementDark
	// ElementArcane represents pure magical energy
	ElementArcane
)

// String returns the string representation of an element type.
func (e ElementType) String() string {
	switch e {
	case ElementNone:
		return "none"
	case ElementFire:
		return "fire"
	case ElementIce:
		return "ice"
	case ElementLightning:
		return "lightning"
	case ElementEarth:
		return "earth"
	case ElementWind:
		return "wind"
	case ElementLight:
		return "light"
	case ElementDark:
		return "dark"
	case ElementArcane:
		return "arcane"
	default:
		return "unknown"
	}
}

// Rarity represents how rare/special a spell is.
type Rarity int

const (
	// RarityCommon represents frequently available spells
	RarityCommon Rarity = iota
	// RarityUncommon represents moderately rare spells
	RarityUncommon
	// RarityRare represents rare spells with better effects
	RarityRare
	// RarityEpic represents very rare, powerful spells
	RarityEpic
	// RarityLegendary represents extremely rare, unique spells
	RarityLegendary
)

// String returns the string representation of a rarity level.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "common"
	case RarityUncommon:
		return "uncommon"
	case RarityRare:
		return "rare"
	case RarityEpic:
		return "epic"
	case RarityLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// TargetType represents what a spell can target.
type TargetType int

const (
	// TargetSelf affects only the caster
	TargetSelf TargetType = iota
	// TargetSingle affects one target
	TargetSingle
	// TargetArea affects all targets in an area
	TargetArea
	// TargetCone affects targets in a cone
	TargetCone
	// TargetLine affects targets in a line
	TargetLine
	// TargetAllAllies affects all allies
	TargetAllAllies
	// TargetAllEnemies affects all enemies
	TargetAllEnemies
)

// String returns the string representation of a target type.
func (t TargetType) String() string {
	switch t {
	case TargetSelf:
		return "self"
	case TargetSingle:
		return "single"
	case TargetArea:
		return "area"
	case TargetCone:
		return "cone"
	case TargetLine:
		return "line"
	case TargetAllAllies:
		return "all_allies"
	case TargetAllEnemies:
		return "all_enemies"
	default:
		return "unknown"
	}
}

// Stats represents the core statistics of a spell.
type Stats struct {
	// Damage dealt by offensive spells
	Damage int
	// Healing provided by healing spells
	Healing int
	// ManaCost is the mana required to cast
	ManaCost int
	// Cooldown in seconds before spell can be cast again
	Cooldown float64
	// CastTime in seconds required to cast the spell
	CastTime float64
	// Range in game units the spell can reach
	Range float64
	// AreaSize for area-effect spells
	AreaSize float64
	// Duration in seconds for buffs/debuffs
	Duration float64
	// RequiredLevel to learn the spell
	RequiredLevel int
}

// Spell represents a generated magic spell.
type Spell struct {
	// Name is the procedurally generated name
	Name string
	// Type categorizes the spell
	Type SpellType
	// Element indicates the elemental affinity
	Element ElementType
	// Rarity indicates how special/rare the spell is
	Rarity Rarity
	// Target indicates what the spell can target
	Target TargetType
	// Stats contains all numerical attributes
	Stats Stats
	// Seed is the generation seed for this spell
	Seed int64
	// Tags are additional descriptive labels
	Tags []string
	// Description is generated flavor text
	Description string
}

// IsOffensive returns true if the spell deals damage.
func (s *Spell) IsOffensive() bool {
	return s.Type == TypeOffensive || s.Type == TypeDebuff
}

// IsSupport returns true if the spell supports allies.
func (s *Spell) IsSupport() bool {
	return s.Type == TypeHealing || s.Type == TypeBuff || s.Type == TypeDefensive
}

// GetPowerLevel returns a numerical power assessment (0-100).
func (s *Spell) GetPowerLevel() int {
	// Calculate power based on stats
	basePower := 0

	if s.Stats.Damage > 0 {
		basePower += s.Stats.Damage * 2
	}
	if s.Stats.Healing > 0 {
		basePower += s.Stats.Healing * 2
	}
	if s.Stats.Duration > 0 {
		basePower += int(s.Stats.Duration * 3)
	}
	if s.Stats.AreaSize > 0 {
		basePower += int(s.Stats.AreaSize * 5)
	}

	// Adjust for cost
	if s.Stats.ManaCost > 0 {
		basePower = basePower * 100 / s.Stats.ManaCost
	}

	// Multiply by rarity
	rarityMultiplier := 1.0
	switch s.Rarity {
	case RarityUncommon:
		rarityMultiplier = 1.2
	case RarityRare:
		rarityMultiplier = 1.5
	case RarityEpic:
		rarityMultiplier = 2.0
	case RarityLegendary:
		rarityMultiplier = 3.0
	}

	power := int(float64(basePower) * rarityMultiplier)

	// Cap at 100
	if power > 100 {
		power = 100
	}

	return power
}

// SpellTemplate defines a template for generating spells.
type SpellTemplate struct {
	BaseType      SpellType
	BaseElement   ElementType
	BaseTarget    TargetType
	NamePrefixes  []string
	NameSuffixes  []string
	Tags          []string
	DamageRange   [2]int
	HealingRange  [2]int
	ManaCostRange [2]int
	CooldownRange [2]float64
	CastTimeRange [2]float64
	RangeRange    [2]float64
	AreaSizeRange [2]float64
	DurationRange [2]float64
}

// GetFantasyOffensiveTemplates returns offensive spell templates for fantasy genre.
func GetFantasyOffensiveTemplates() []SpellTemplate {
	return []SpellTemplate{
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementFire,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Fire", "Flame", "Inferno", "Blaze", "Burning"},
			NameSuffixes:  []string{"Bolt", "Strike", "Blast", "Arrow", "Ray"},
			Tags:          []string{"fire", "damage", "burn"},
			DamageRange:   [2]int{20, 50},
			ManaCostRange: [2]int{15, 30},
			CooldownRange: [2]float64{2.0, 5.0},
			CastTimeRange: [2]float64{0.5, 1.5},
			RangeRange:    [2]float64{10.0, 25.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementIce,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Ice", "Frost", "Frozen", "Glacial", "Arctic"},
			NameSuffixes:  []string{"Storm", "Nova", "Explosion", "Wave", "Blast"},
			Tags:          []string{"ice", "area", "slow"},
			DamageRange:   [2]int{30, 80},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{5.0, 10.0},
			CastTimeRange: [2]float64{1.0, 2.0},
			RangeRange:    [2]float64{5.0, 15.0},
			AreaSizeRange: [2]float64{5.0, 10.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementLightning,
			BaseTarget:    TargetLine,
			NamePrefixes:  []string{"Lightning", "Thunder", "Shock", "Electric", "Volt"},
			NameSuffixes:  []string{"Bolt", "Strike", "Chain", "Beam", "Arc"},
			Tags:          []string{"lightning", "chain", "fast"},
			DamageRange:   [2]int{25, 60},
			ManaCostRange: [2]int{20, 40},
			CooldownRange: [2]float64{3.0, 6.0},
			CastTimeRange: [2]float64{0.3, 1.0},
			RangeRange:    [2]float64{15.0, 30.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementEarth,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Stone", "Rock", "Boulder", "Earth", "Granite"},
			NameSuffixes:  []string{"Throw", "Spike", "Fist", "Lance", "Barrage"},
			Tags:          []string{"earth", "physical", "stun"},
			DamageRange:   [2]int{35, 70},
			ManaCostRange: [2]int{20, 35},
			CooldownRange: [2]float64{4.0, 7.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			RangeRange:    [2]float64{8.0, 20.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementDark,
			BaseTarget:    TargetCone,
			NamePrefixes:  []string{"Shadow", "Dark", "Void", "Curse", "Doom"},
			NameSuffixes:  []string{"Bolt", "Wave", "Beam", "Touch", "Blast"},
			Tags:          []string{"dark", "curse", "fear"},
			DamageRange:   [2]int{28, 65},
			ManaCostRange: [2]int{25, 45},
			CooldownRange: [2]float64{4.0, 8.0},
			CastTimeRange: [2]float64{0.7, 1.8},
			RangeRange:    [2]float64{8.0, 15.0},
			AreaSizeRange: [2]float64{3.0, 8.0},
		},
	}
}

// GetFantasySupportTemplates returns support spell templates for fantasy genre.
func GetFantasySupportTemplates() []SpellTemplate {
	return []SpellTemplate{
		{
			BaseType:      TypeHealing,
			BaseElement:   ElementLight,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Heal", "Cure", "Mend", "Restore", "Divine"},
			NameSuffixes:  []string{"Touch", "Light", "Grace", "Blessing", "Aid"},
			Tags:          []string{"healing", "light", "holy"},
			HealingRange:  [2]int{30, 80},
			ManaCostRange: [2]int{20, 40},
			CooldownRange: [2]float64{3.0, 8.0},
			CastTimeRange: [2]float64{0.5, 1.5},
			RangeRange:    [2]float64{5.0, 15.0},
		},
		{
			BaseType:      TypeDefensive,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Mana", "Magic", "Arcane", "Mystic", "Energy"},
			NameSuffixes:  []string{"Shield", "Barrier", "Ward", "Protection", "Armor"},
			Tags:          []string{"defense", "shield", "protection"},
			ManaCostRange: [2]int{15, 35},
			CooldownRange: [2]float64{10.0, 20.0},
			CastTimeRange: [2]float64{0.5, 1.0},
			DurationRange: [2]float64{15.0, 45.0},
		},
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementWind,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Haste", "Swift", "Quick", "Speed", "Rush"},
			NameSuffixes:  []string{"Blessing", "Enchantment", "Boost", "Enhancement"},
			Tags:          []string{"buff", "speed", "haste"},
			ManaCostRange: [2]int{10, 25},
			CooldownRange: [2]float64{15.0, 30.0},
			CastTimeRange: [2]float64{0.3, 0.8},
			RangeRange:    [2]float64{5.0, 10.0},
			DurationRange: [2]float64{20.0, 60.0},
		},
		{
			BaseType:      TypeDebuff,
			BaseElement:   ElementDark,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Weakness", "Slow", "Curse", "Hex", "Bane"},
			NameSuffixes:  []string{"Touch", "Affliction", "Plague", "Spell"},
			Tags:          []string{"debuff", "curse", "weaken"},
			DamageRange:   [2]int{5, 15},
			ManaCostRange: [2]int{12, 28},
			CooldownRange: [2]float64{8.0, 15.0},
			CastTimeRange: [2]float64{0.5, 1.2},
			RangeRange:    [2]float64{8.0, 20.0},
			DurationRange: [2]float64{10.0, 30.0},
		},
	}
}

// GetSciFiOffensiveTemplates returns offensive spell templates for sci-fi genre.
func GetSciFiOffensiveTemplates() []SpellTemplate {
	return []SpellTemplate{
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementLightning, // Tech as lightning
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Plasma", "Ion", "Photon", "Laser", "Particle"},
			NameSuffixes:  []string{"Beam", "Burst", "Lance", "Cannon", "Pulse"},
			Tags:          []string{"energy", "tech", "precision"},
			DamageRange:   [2]int{25, 60},
			ManaCostRange: [2]int{18, 35},
			CooldownRange: [2]float64{2.5, 6.0},
			CastTimeRange: [2]float64{0.3, 1.2},
			RangeRange:    [2]float64{15.0, 40.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementFire, // Explosive as fire
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Fusion", "Quantum", "Nuclear", "Thermal", "Explosive"},
			NameSuffixes:  []string{"Blast", "Detonation", "Missile", "Grenade", "Bomb"},
			Tags:          []string{"explosive", "area", "tech"},
			DamageRange:   [2]int{35, 90},
			ManaCostRange: [2]int{35, 65},
			CooldownRange: [2]float64{6.0, 12.0},
			CastTimeRange: [2]float64{1.2, 2.5},
			RangeRange:    [2]float64{10.0, 25.0},
			AreaSizeRange: [2]float64{6.0, 12.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementIce, // Cryo as ice
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Cryo", "Freeze", "Stasis", "Zero", "Cold"},
			NameSuffixes:  []string{"Beam", "Ray", "Field", "Shot", "Blast"},
			Tags:          []string{"cryo", "freeze", "slow"},
			DamageRange:   [2]int{22, 55},
			ManaCostRange: [2]int{20, 38},
			CooldownRange: [2]float64{3.5, 7.0},
			CastTimeRange: [2]float64{0.5, 1.3},
			RangeRange:    [2]float64{12.0, 28.0},
		},
	}
}

// GetSciFiSupportTemplates returns support spell templates for sci-fi genre.
func GetSciFiSupportTemplates() []SpellTemplate {
	return []SpellTemplate{
		{
			BaseType:      TypeHealing,
			BaseElement:   ElementLight, // Medical as light
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Nano", "Medical", "Bio", "Regen", "Heal"},
			NameSuffixes:  []string{"Injection", "Field", "Spray", "Boost", "Pack"},
			Tags:          []string{"medical", "healing", "tech"},
			HealingRange:  [2]int{35, 90},
			ManaCostRange: [2]int{22, 42},
			CooldownRange: [2]float64{4.0, 10.0},
			CastTimeRange: [2]float64{0.4, 1.2},
			RangeRange:    [2]float64{5.0, 18.0},
		},
		{
			BaseType:      TypeDefensive,
			BaseElement:   ElementArcane, // Tech as arcane
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Energy", "Quantum", "Force", "Kinetic", "Shield"},
			NameSuffixes:  []string{"Barrier", "Field", "Shield", "Matrix", "Wall"},
			Tags:          []string{"defense", "shield", "tech"},
			ManaCostRange: [2]int{18, 40},
			CooldownRange: [2]float64{12.0, 25.0},
			CastTimeRange: [2]float64{0.4, 1.0},
			DurationRange: [2]float64{18.0, 50.0},
		},
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementLightning, // Tech boost
			BaseTarget:    TargetAllAllies,
			NamePrefixes:  []string{"Combat", "Tactical", "Battle", "War", "System"},
			NameSuffixes:  []string{"Stimulant", "Boost", "Enhancement", "Override", "Protocol"},
			Tags:          []string{"buff", "combat", "tech"},
			ManaCostRange: [2]int{25, 50},
			CooldownRange: [2]float64{20.0, 40.0},
			CastTimeRange: [2]float64{0.5, 1.5},
			DurationRange: [2]float64{25.0, 70.0},
		},
	}
}
