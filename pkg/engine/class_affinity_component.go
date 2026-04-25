// Package engine provides the class affinity component for tracking playstyle preferences.
// This component tracks how players use their class abilities, building affinity
// toward combat archetypes that unlock passive bonuses and ability enhancements.
package engine

// AffinityType represents a combat archetype that players can develop.
type AffinityType int

const (
	// AffinityNone indicates no particular affinity has developed.
	AffinityNone AffinityType = iota

	// Combat Archetypes
	AffinityAggressor   // High damage, offensive focus
	AffinityDefender    // Tanky, damage mitigation focus
	AffinityCaster      // Spell-heavy playstyle
	AffinitySupportive  // Healing/buffing focus
	AffinityStealthy    // Stealth and assassination
	AffinityTactical    // Control and crowd control
	AffinityBurstDamage // High single-target burst
	AffinityAreaDamage  // AOE specialist
	AffinityDrainer     // Life/mana drain specialist
	AffinitySummoner    // Minion-focused playstyle
)

// String returns the display name for an AffinityType.
func (a AffinityType) String() string {
	switch a {
	case AffinityNone:
		return "None"
	case AffinityAggressor:
		return "Aggressor"
	case AffinityDefender:
		return "Defender"
	case AffinityCaster:
		return "Caster"
	case AffinitySupportive:
		return "Supportive"
	case AffinityStealthy:
		return "Stealthy"
	case AffinityTactical:
		return "Tactical"
	case AffinityBurstDamage:
		return "Burst Damage"
	case AffinityAreaDamage:
		return "Area Damage"
	case AffinityDrainer:
		return "Drainer"
	case AffinitySummoner:
		return "Summoner"
	default:
		return "Unknown"
	}
}

// Description returns a brief description of the affinity archetype.
func (a AffinityType) Description() string {
	switch a {
	case AffinityNone:
		return "No dominant playstyle detected yet."
	case AffinityAggressor:
		return "Relentless attacker focused on dealing damage."
	case AffinityDefender:
		return "Stalwart protector who mitigates and absorbs damage."
	case AffinityCaster:
		return "Master of arcane arts who relies on spells."
	case AffinitySupportive:
		return "Ally-focused healer and buffer."
	case AffinityStealthy:
		return "Shadow operative who strikes from concealment."
	case AffinityTactical:
		return "Battlefield controller who manipulates enemy positioning."
	case AffinityBurstDamage:
		return "Assassin who eliminates targets with overwhelming force."
	case AffinityAreaDamage:
		return "Devastator who annihilates groups of enemies."
	case AffinityDrainer:
		return "Life force manipulator who sustains through theft."
	case AffinitySummoner:
		return "Commander who fights through summoned allies."
	default:
		return "Unknown archetype."
	}
}

// AffinityLevel represents the mastery level of an affinity.
type AffinityLevel int

const (
	AffinityLevelNone        AffinityLevel = iota // 0 XP
	AffinityLevelNovice                           // 100 XP
	AffinityLevelApprentice                       // 500 XP
	AffinityLevelJourneyman                       // 1500 XP
	AffinityLevelExpert                           // 4000 XP
	AffinityLevelMaster                           // 10000 XP
	AffinityLevelGrandmaster                      // 25000 XP
)

// String returns the display name for an AffinityLevel.
func (l AffinityLevel) String() string {
	switch l {
	case AffinityLevelNone:
		return "None"
	case AffinityLevelNovice:
		return "Novice"
	case AffinityLevelApprentice:
		return "Apprentice"
	case AffinityLevelJourneyman:
		return "Journeyman"
	case AffinityLevelExpert:
		return "Expert"
	case AffinityLevelMaster:
		return "Master"
	case AffinityLevelGrandmaster:
		return "Grandmaster"
	default:
		return "Unknown"
	}
}

// XPThreshold returns the XP required to reach this level.
func (l AffinityLevel) XPThreshold() int {
	switch l {
	case AffinityLevelNone:
		return 0
	case AffinityLevelNovice:
		return 100
	case AffinityLevelApprentice:
		return 500
	case AffinityLevelJourneyman:
		return 1500
	case AffinityLevelExpert:
		return 4000
	case AffinityLevelMaster:
		return 10000
	case AffinityLevelGrandmaster:
		return 25000
	default:
		return 0
	}
}

// AffinityData tracks progress in a specific affinity.
type AffinityData struct {
	XP               int           // Total XP earned
	Level            AffinityLevel // Current level
	AbilitiesUsed    int           // Total relevant abilities used
	DamageDealt      float64       // Damage dealt with affinity abilities
	TimesTriggered   int           // How many times affinity bonuses triggered
	PeakStreak       int           // Longest streak of consecutive affinity actions
	CurrentStreak    int           // Current streak counter
	LastActivityTime float64       // Game time of last affinity-relevant action
}

// ClassAffinityComponent tracks a player's development of combat archetypes.
// As players use abilities, they build XP in relevant affinities.
type ClassAffinityComponent struct {
	// Affinities maps affinity types to their progress data
	Affinities map[AffinityType]*AffinityData

	// PrimaryAffinity is the highest-level affinity (computed)
	PrimaryAffinity AffinityType

	// SecondaryAffinity is the second-highest affinity (computed)
	SecondaryAffinity AffinityType

	// TotalAffinityXP is the sum of all affinity XP earned
	TotalAffinityXP int

	// BonusesApplied tracks which affinity bonuses are currently applied
	BonusesApplied map[AffinityType]AffinityLevel

	// AppliedManaRegen stores the absolute mana regen value that was added for
	// each affinity at the time it was applied. G37 fix: using the stored value
	// at removal time prevents drift when mana.Max changes between apply and removal.
	AppliedManaRegen map[AffinityType]float64

	// Dirty marks that bonuses need recalculation
	Dirty bool
}

// Type returns the component type identifier.
func (c *ClassAffinityComponent) Type() string {
	return "class_affinity"
}

// NewClassAffinityComponent creates a new class affinity component.
func NewClassAffinityComponent() *ClassAffinityComponent {
	return &ClassAffinityComponent{
		Affinities:        make(map[AffinityType]*AffinityData),
		PrimaryAffinity:   AffinityNone,
		SecondaryAffinity: AffinityNone,
		TotalAffinityXP:   0,
		BonusesApplied:    make(map[AffinityType]AffinityLevel),
		AppliedManaRegen:  make(map[AffinityType]float64),
		Dirty:             false,
	}
}

// GetAffinity returns the affinity data for a type, creating if needed.
func (c *ClassAffinityComponent) GetAffinity(affinity AffinityType) *AffinityData {
	if c.Affinities == nil {
		c.Affinities = make(map[AffinityType]*AffinityData)
	}
	if _, exists := c.Affinities[affinity]; !exists {
		c.Affinities[affinity] = &AffinityData{
			Level: AffinityLevelNone,
		}
	}
	return c.Affinities[affinity]
}

// GetAffinityLevel returns the current level for an affinity type.
func (c *ClassAffinityComponent) GetAffinityLevel(affinity AffinityType) AffinityLevel {
	data := c.GetAffinity(affinity)
	return data.Level
}

// GetXPToNextLevel returns XP needed to reach next level.
func (c *ClassAffinityComponent) GetXPToNextLevel(affinity AffinityType) int {
	data := c.GetAffinity(affinity)
	currentXP := data.XP
	currentLevel := data.Level
	nextLevel := currentLevel + 1

	if nextLevel > AffinityLevelGrandmaster {
		return 0 // Already at max
	}

	return nextLevel.XPThreshold() - currentXP
}

// GetProgressToNextLevel returns progress (0-1) toward next level.
func (c *ClassAffinityComponent) GetProgressToNextLevel(affinity AffinityType) float64 {
	data := c.GetAffinity(affinity)
	currentXP := data.XP
	currentLevel := data.Level
	nextLevel := currentLevel + 1

	if nextLevel > AffinityLevelGrandmaster {
		return 1.0 // Max level
	}

	currentThreshold := currentLevel.XPThreshold()
	nextThreshold := nextLevel.XPThreshold()
	rangeXP := nextThreshold - currentThreshold

	if rangeXP <= 0 {
		return 0
	}

	return float64(currentXP-currentThreshold) / float64(rangeXP)
}

// RecalculatePrimaryAffinities determines primary and secondary affinities.
func (c *ClassAffinityComponent) RecalculatePrimaryAffinities() {
	var first, second AffinityType = AffinityNone, AffinityNone
	var firstXP, secondXP int = 0, 0

	for affinity, data := range c.Affinities {
		if data.XP > firstXP {
			second = first
			secondXP = firstXP
			first = affinity
			firstXP = data.XP
		} else if data.XP > secondXP {
			second = affinity
			secondXP = data.XP
		}
	}

	c.PrimaryAffinity = first
	c.SecondaryAffinity = second
}

// AffinityBonuses defines the stat bonuses granted at each affinity level.
type AffinityBonuses struct {
	// Damage multipliers
	DamageMultiplier  float64
	SpellDamageMult   float64
	HealingMultiplier float64

	// Stat bonuses
	CritChanceBonus float64
	CritDamageBonus float64
	CooldownReduce  float64

	// Defense bonuses
	DamageReduction float64
	EvasionBonus    float64
	ArmorBonus      float64

	// Resource bonuses
	ManaRegenBonus   float64
	HealthRegenBonus float64
}

// GetAffinityBonuses returns the bonuses for an affinity at a given level.
func GetAffinityBonuses(affinity AffinityType, level AffinityLevel) AffinityBonuses {
	// Base bonuses scale with level
	levelMultiplier := float64(level) / float64(AffinityLevelGrandmaster)

	switch affinity {
	case AffinityAggressor:
		return AffinityBonuses{
			DamageMultiplier: 1.0 + (0.25 * levelMultiplier), // Up to +25% damage
			CritChanceBonus:  0.08 * levelMultiplier,         // Up to +8% crit
			CritDamageBonus:  0.20 * levelMultiplier,         // Up to +20% crit damage
		}
	case AffinityDefender:
		return AffinityBonuses{
			DamageReduction:  0.15 * levelMultiplier, // Up to 15% damage reduction
			ArmorBonus:       50 * levelMultiplier,   // Up to +50 armor
			HealthRegenBonus: 0.03 * levelMultiplier, // Up to +3% health regen
		}
	case AffinityCaster:
		return AffinityBonuses{
			SpellDamageMult: 1.0 + (0.30 * levelMultiplier), // Up to +30% spell damage
			ManaRegenBonus:  0.05 * levelMultiplier,         // Up to +5% mana regen
			CooldownReduce:  0.15 * levelMultiplier,         // Up to 15% cooldown reduction
		}
	case AffinitySupportive:
		return AffinityBonuses{
			HealingMultiplier: 1.0 + (0.35 * levelMultiplier), // Up to +35% healing
			ManaRegenBonus:    0.04 * levelMultiplier,         // Up to +4% mana regen
			CooldownReduce:    0.10 * levelMultiplier,         // Up to 10% cooldown reduction
		}
	case AffinityStealthy:
		return AffinityBonuses{
			CritChanceBonus: 0.12 * levelMultiplier, // Up to +12% crit from stealth
			CritDamageBonus: 0.40 * levelMultiplier, // Up to +40% crit damage
			EvasionBonus:    0.10 * levelMultiplier, // Up to +10% evasion
			CooldownReduce:  0.08 * levelMultiplier, // Up to 8% cooldown reduction
		}
	case AffinityTactical:
		return AffinityBonuses{
			CooldownReduce:   0.20 * levelMultiplier,         // Up to 20% cooldown reduction
			ManaRegenBonus:   0.03 * levelMultiplier,         // Up to +3% mana regen
			DamageMultiplier: 1.0 + (0.10 * levelMultiplier), // Up to +10% damage
		}
	case AffinityBurstDamage:
		return AffinityBonuses{
			CritChanceBonus:  0.15 * levelMultiplier,         // Up to +15% crit
			CritDamageBonus:  0.50 * levelMultiplier,         // Up to +50% crit damage
			DamageMultiplier: 1.0 + (0.15 * levelMultiplier), // Up to +15% damage
		}
	case AffinityAreaDamage:
		return AffinityBonuses{
			SpellDamageMult: 1.0 + (0.20 * levelMultiplier), // Up to +20% AOE damage
			ManaRegenBonus:  0.02 * levelMultiplier,         // Up to +2% mana regen
			CooldownReduce:  0.12 * levelMultiplier,         // Up to 12% cooldown reduction
		}
	case AffinityDrainer:
		return AffinityBonuses{
			HealthRegenBonus: 0.08 * levelMultiplier, // Up to +8% health regen
			ManaRegenBonus:   0.06 * levelMultiplier, // Up to +6% mana regen
			SpellDamageMult:  1.0 + (0.15 * levelMultiplier),
		}
	case AffinitySummoner:
		return AffinityBonuses{
			DamageMultiplier: 1.0 + (0.12 * levelMultiplier), // Summon damage
			ManaRegenBonus:   0.04 * levelMultiplier,
			CooldownReduce:   0.15 * levelMultiplier, // Faster summons
		}
	default:
		return AffinityBonuses{DamageMultiplier: 1.0, SpellDamageMult: 1.0, HealingMultiplier: 1.0}
	}
}

// abilityAffinityMap maps ability names to their associated affinities.
var abilityAffinityMap = map[string][]AffinityType{
	// Warrior abilities
	"power_strike":     {AffinityAggressor, AffinityBurstDamage},
	"shield_bash":      {AffinityDefender, AffinityTactical},
	"battle_cry":       {AffinitySupportive, AffinityTactical},
	"cleave":           {AffinityAggressor, AffinityAreaDamage},
	"charge":           {AffinityAggressor, AffinityBurstDamage},
	"defensive_stance": {AffinityDefender},
	"execute":          {AffinityBurstDamage, AffinityAggressor},
	"taunt":            {AffinityDefender, AffinityTactical},

	// Rogue abilities
	"backstab":     {AffinityStealthy, AffinityBurstDamage},
	"dual_wield":   {AffinityAggressor},
	"stealth":      {AffinityStealthy},
	"poison_blade": {AffinityStealthy, AffinityDrainer},
	"evade":        {AffinityStealthy, AffinityDefender},
	"ambush":       {AffinityStealthy, AffinityBurstDamage},
	"shadow_step":  {AffinityStealthy, AffinityTactical},
	"disarm":       {AffinityTactical},

	// Mage abilities
	"fireball":       {AffinityCaster, AffinityAreaDamage},
	"ice_shard":      {AffinityCaster, AffinityTactical},
	"magic_missile":  {AffinityCaster},
	"mana_shield":    {AffinityDefender, AffinityCaster},
	"lightning_bolt": {AffinityCaster, AffinityBurstDamage},
	"frost_nova":     {AffinityCaster, AffinityAreaDamage, AffinityTactical},
	"teleport":       {AffinityStealthy, AffinityTactical},
	"arcane_barrage": {AffinityCaster, AffinityAreaDamage},

	// Ranger abilities
	"aimed_shot":     {AffinityBurstDamage, AffinityAggressor},
	"rapid_fire":     {AffinityAggressor},
	"tame_beast":     {AffinitySummoner},
	"track":          {AffinityTactical},
	"explosive_shot": {AffinityAreaDamage, AffinityBurstDamage},
	"multi_shot":     {AffinityAreaDamage},
	"camouflage":     {AffinityStealthy},
	"hunters_mark":   {AffinityTactical, AffinityBurstDamage},

	// Cleric abilities
	"heal":          {AffinitySupportive},
	"smite":         {AffinityCaster, AffinityAggressor},
	"divine_shield": {AffinityDefender, AffinitySupportive},
	"prayer":        {AffinitySupportive},
	"resurrection":  {AffinitySupportive},
	"holy_light":    {AffinitySupportive, AffinityAreaDamage},
	"purify":        {AffinitySupportive, AffinityTactical},
	"blessing":      {AffinitySupportive},

	// Necromancer abilities
	"raise_dead":       {AffinitySummoner},
	"life_drain":       {AffinityDrainer},
	"curse":            {AffinityTactical, AffinityDrainer},
	"bone_armor":       {AffinityDefender},
	"death_coil":       {AffinityDrainer, AffinityCaster},
	"fear":             {AffinityTactical},
	"corpse_explosion": {AffinityAreaDamage, AffinitySummoner},
	"soul_harvest":     {AffinityDrainer},

	// Specialization abilities - Berserker
	"rage":            {AffinityAggressor, AffinityBurstDamage},
	"whirlwind":       {AffinityAggressor, AffinityAreaDamage},
	"reckless_attack": {AffinityAggressor, AffinityBurstDamage},

	// Defender
	"shield_wall": {AffinityDefender},
	"iron_skin":   {AffinityDefender},

	// Elementalist
	"pyroblast": {AffinityCaster, AffinityBurstDamage, AffinityAreaDamage},

	// Arcanist
	"arcane_blast": {AffinityCaster, AffinityBurstDamage},
	"time_warp":    {AffinityTactical, AffinityCaster},
	"spell_steal":  {AffinityTactical, AffinityCaster},

	// Assassin
	"shadow_strike": {AffinityStealthy, AffinityBurstDamage},
	"deadly_poison": {AffinityStealthy, AffinityDrainer},
	"vanish":        {AffinityStealthy},

	// And more specialization abilities...
	"call_of_wild":   {AffinitySummoner},
	"bestial_wrath":  {AffinitySummoner, AffinityAggressor},
	"mend_pet":       {AffinitySummoner, AffinitySupportive},
	"greater_heal":   {AffinitySupportive},
	"resurrect":      {AffinitySupportive},
	"dispel_magic":   {AffinitySupportive, AffinityTactical},
	"army_of_dead":   {AffinitySummoner},
	"blood_boil":     {AffinityDrainer, AffinityAreaDamage},
	"vampiric_touch": {AffinityDrainer},
	"siphon_life":    {AffinityDrainer},
}

// GetAbilityAffinities returns the affinities associated with an ability.
func GetAbilityAffinities(abilityID string) []AffinityType {
	if affinities, exists := abilityAffinityMap[abilityID]; exists {
		return affinities
	}
	return nil
}

// Serialize converts ClassAffinityComponent to bytes.
func (c *ClassAffinityComponent) Serialize() []byte {
	// Calculate size: 4 bytes for count + (affinity data) per entry
	// Each entry: 1 byte type + 4 bytes XP + 1 byte level + 4 bytes abilities + 8 bytes damage + 4 bytes times
	entrySize := 22
	size := 4 + (len(c.Affinities) * entrySize) + 1 + 1 + 4

	buf := make([]byte, size)
	offset := 0

	// Write number of affinities
	writeInt32(buf[offset:], int32(len(c.Affinities)))
	offset += 4

	// Write each affinity
	for affinityType, data := range c.Affinities {
		buf[offset] = byte(affinityType)
		offset++
		writeInt32(buf[offset:], int32(data.XP))
		offset += 4
		buf[offset] = byte(data.Level)
		offset++
		writeInt32(buf[offset:], int32(data.AbilitiesUsed))
		offset += 4
		writeFloat64(buf[offset:], data.DamageDealt)
		offset += 8
		writeInt32(buf[offset:], int32(data.TimesTriggered))
		offset += 4
	}

	// Write primary and secondary affinities
	buf[offset] = byte(c.PrimaryAffinity)
	offset++
	buf[offset] = byte(c.SecondaryAffinity)
	offset++
	writeInt32(buf[offset:], int32(c.TotalAffinityXP))

	return buf
}

// Deserialize reads ClassAffinityComponent from bytes.
func (c *ClassAffinityComponent) Deserialize(data []byte) error {
	if len(data) < 4 {
		return ErrInvalidComponentData
	}

	offset := 0

	// Read number of affinities
	count := int(readInt32(data[offset:]))
	offset += 4

	c.Affinities = make(map[AffinityType]*AffinityData, count)

	for i := 0; i < count; i++ {
		if offset >= len(data) {
			break
		}

		affinityType := AffinityType(data[offset])
		offset++

		affinityData := &AffinityData{}
		affinityData.XP = int(readInt32(data[offset:]))
		offset += 4
		affinityData.Level = AffinityLevel(data[offset])
		offset++
		affinityData.AbilitiesUsed = int(readInt32(data[offset:]))
		offset += 4
		affinityData.DamageDealt = readFloat64(data[offset:])
		offset += 8
		affinityData.TimesTriggered = int(readInt32(data[offset:]))
		offset += 4

		c.Affinities[affinityType] = affinityData
	}

	// Read primary and secondary affinities
	if offset+6 <= len(data) {
		c.PrimaryAffinity = AffinityType(data[offset])
		offset++
		c.SecondaryAffinity = AffinityType(data[offset])
		offset++
		c.TotalAffinityXP = int(readInt32(data[offset:]))
	}

	c.BonusesApplied = make(map[AffinityType]AffinityLevel)
	c.AppliedManaRegen = make(map[AffinityType]float64)
	c.Dirty = true

	return nil
}
