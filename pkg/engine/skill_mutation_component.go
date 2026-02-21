// Package engine provides the skill mutation component for character customization.
// This component tracks skill mutations that modify spell/skill behavior.
// Mutations allow players to customize skills by trading off stats, changing
// elements, adding effects, or altering targeting behavior.
package engine

import (
	"encoding/json"
)

// MutationType represents the category of a skill mutation.
type MutationType int

const (
	// MutationDamage modifies damage output (positive or negative).
	MutationDamage MutationType = iota
	// MutationCooldown modifies cooldown duration (positive = longer, negative = shorter).
	MutationCooldown
	// MutationManaCost modifies mana consumption.
	MutationManaCost
	// MutationRange modifies spell range.
	MutationRange
	// MutationArea modifies area of effect size.
	MutationArea
	// MutationDuration modifies buff/debuff duration.
	MutationDuration
	// MutationElemental changes the spell's element.
	MutationElemental
	// MutationChain adds chain effect (hits multiple targets).
	MutationChain
	// MutationLifesteal adds life drain to damage.
	MutationLifesteal
	// MutationPierce makes spell ignore some resistance.
	MutationPierce
	// MutationSplit divides spell into multiple smaller projectiles.
	MutationSplit
	// MutationEcho has a chance to cast spell twice.
	MutationEcho
)

// String returns the display name for a mutation type.
func (m MutationType) String() string {
	switch m {
	case MutationDamage:
		return "Damage"
	case MutationCooldown:
		return "Cooldown"
	case MutationManaCost:
		return "Mana Cost"
	case MutationRange:
		return "Range"
	case MutationArea:
		return "Area"
	case MutationDuration:
		return "Duration"
	case MutationElemental:
		return "Elemental"
	case MutationChain:
		return "Chain"
	case MutationLifesteal:
		return "Lifesteal"
	case MutationPierce:
		return "Pierce"
	case MutationSplit:
		return "Split"
	case MutationEcho:
		return "Echo"
	default:
		return "Unknown"
	}
}

// MutationRarity indicates how rare/powerful a mutation is.
type MutationRarity int

const (
	// MutationRarityCommon is basic mutations with modest effects.
	MutationRarityCommon MutationRarity = iota
	// MutationRarityUncommon has better effects.
	MutationRarityUncommon
	// MutationRarityRare has strong effects.
	MutationRarityRare
	// MutationRarityEpic has powerful effects.
	MutationRarityEpic
	// MutationRarityLegendary has game-changing effects.
	MutationRarityLegendary
)

// String returns the display name for a mutation rarity.
func (r MutationRarity) String() string {
	switch r {
	case MutationRarityCommon:
		return "Common"
	case MutationRarityUncommon:
		return "Uncommon"
	case MutationRarityRare:
		return "Rare"
	case MutationRarityEpic:
		return "Epic"
	case MutationRarityLegendary:
		return "Legendary"
	default:
		return "Unknown"
	}
}

// SkillMutation represents a single mutation that can be applied to a skill.
// Mutations modify skill behavior by adjusting stats or adding new effects.
type SkillMutation struct {
	// ID is the unique identifier for this mutation.
	ID string `json:"id"`
	// Name is the display name of the mutation.
	Name string `json:"name"`
	// Description explains what the mutation does.
	Description string `json:"description"`
	// Type categorizes the mutation effect.
	Type MutationType `json:"type"`
	// Rarity indicates how rare/powerful this mutation is.
	Rarity MutationRarity `json:"rarity"`
	// PrimaryValue is the main effect magnitude (interpretation depends on Type).
	// For damage: percentage modifier (+20 = +20% damage).
	// For cooldown: percentage modifier (-15 = -15% cooldown).
	// For chain: number of additional targets.
	// For lifesteal: percentage of damage healed.
	PrimaryValue float64 `json:"primary_value"`
	// SecondaryValue is an optional secondary effect.
	// For elemental: new element type (cast to ElementType).
	// For echo: chance percentage (0-100).
	// For pierce: resistance ignored percentage.
	SecondaryValue float64 `json:"secondary_value"`
	// TradeoffType is what stat is reduced to balance the mutation.
	TradeoffType MutationType `json:"tradeoff_type"`
	// TradeoffValue is the magnitude of the tradeoff (always positive, treated as penalty).
	TradeoffValue float64 `json:"tradeoff_value"`
	// Seed is the generation seed for this mutation.
	Seed int64 `json:"seed"`
	// RequiredLevel is the minimum character level to use this mutation.
	RequiredLevel int `json:"required_level"`
	// Incompatible lists mutation IDs that cannot be combined with this one.
	Incompatible []string `json:"incompatible"`
}

// Copy creates a deep copy of the mutation.
func (m *SkillMutation) Copy() *SkillMutation {
	incompatible := make([]string, len(m.Incompatible))
	copy(incompatible, m.Incompatible)
	return &SkillMutation{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		Type:           m.Type,
		Rarity:         m.Rarity,
		PrimaryValue:   m.PrimaryValue,
		SecondaryValue: m.SecondaryValue,
		TradeoffType:   m.TradeoffType,
		TradeoffValue:  m.TradeoffValue,
		Seed:           m.Seed,
		RequiredLevel:  m.RequiredLevel,
		Incompatible:   incompatible,
	}
}

// GetPowerScore returns a numerical assessment of mutation power (0-100).
func (m *SkillMutation) GetPowerScore() int {
	// Base power from primary value
	basePower := m.PrimaryValue
	if m.Type == MutationCooldown || m.Type == MutationManaCost {
		// For cost reductions, negative values are beneficial
		basePower = -basePower
	}

	// Adjust for tradeoff
	tradeoffPenalty := m.TradeoffValue * 0.5

	// Rarity multiplier
	rarityMult := 1.0
	switch m.Rarity {
	case MutationRarityUncommon:
		rarityMult = 1.2
	case MutationRarityRare:
		rarityMult = 1.5
	case MutationRarityEpic:
		rarityMult = 2.0
	case MutationRarityLegendary:
		rarityMult = 3.0
	}

	power := int((basePower - tradeoffPenalty) * rarityMult)
	if power < 0 {
		power = 0
	}
	if power > 100 {
		power = 100
	}
	return power
}

// MutatedSkill tracks mutations applied to a specific skill.
type MutatedSkill struct {
	// SkillID is the identifier of the skill being mutated.
	SkillID string `json:"skill_id"`
	// SpellSlot is which spell slot this applies to (0-4, -1 if not a spell slot).
	SpellSlot int `json:"spell_slot"`
	// Mutations is the list of applied mutations.
	Mutations []*SkillMutation `json:"mutations"`
	// MaxMutations is the maximum mutations allowed on this skill.
	MaxMutations int `json:"max_mutations"`
	// Locked prevents further mutation changes if true.
	Locked bool `json:"locked"`
}

// Copy creates a deep copy of the mutated skill.
func (ms *MutatedSkill) Copy() *MutatedSkill {
	mutations := make([]*SkillMutation, len(ms.Mutations))
	for i, m := range ms.Mutations {
		mutations[i] = m.Copy()
	}
	return &MutatedSkill{
		SkillID:      ms.SkillID,
		SpellSlot:    ms.SpellSlot,
		Mutations:    mutations,
		MaxMutations: ms.MaxMutations,
		Locked:       ms.Locked,
	}
}

// GetMutationCount returns the number of mutations applied.
func (ms *MutatedSkill) GetMutationCount() int {
	return len(ms.Mutations)
}

// CanAddMutation checks if another mutation can be added.
func (ms *MutatedSkill) CanAddMutation() bool {
	return !ms.Locked && len(ms.Mutations) < ms.MaxMutations
}

// HasMutation checks if a mutation with the given ID is already applied.
func (ms *MutatedSkill) HasMutation(mutationID string) bool {
	for _, m := range ms.Mutations {
		if m.ID == mutationID {
			return true
		}
	}
	return false
}

// IsMutationCompatible checks if a mutation can be combined with existing mutations.
func (ms *MutatedSkill) IsMutationCompatible(mutation *SkillMutation) bool {
	for _, existing := range ms.Mutations {
		// Check if new mutation is incompatible with existing
		for _, incomp := range mutation.Incompatible {
			if existing.ID == incomp {
				return false
			}
		}
		// Check if existing mutation is incompatible with new
		for _, incomp := range existing.Incompatible {
			if mutation.ID == incomp {
				return false
			}
		}
	}
	return true
}

// GetTotalModifier returns the combined modifier for a given mutation type.
// Returns the sum of all PrimaryValues for mutations of that type.
func (ms *MutatedSkill) GetTotalModifier(mutationType MutationType) float64 {
	total := 0.0
	for _, m := range ms.Mutations {
		if m.Type == mutationType {
			total += m.PrimaryValue
		}
		// Also apply tradeoff penalties
		if m.TradeoffType == mutationType {
			total -= m.TradeoffValue
		}
	}
	return total
}

// GetEffectiveDamageMultiplier returns the damage multiplier (1.0 = no change).
func (ms *MutatedSkill) GetEffectiveDamageMultiplier() float64 {
	modifier := ms.GetTotalModifier(MutationDamage)
	return 1.0 + (modifier / 100.0)
}

// GetEffectiveCooldownMultiplier returns the cooldown multiplier (1.0 = no change).
func (ms *MutatedSkill) GetEffectiveCooldownMultiplier() float64 {
	modifier := ms.GetTotalModifier(MutationCooldown)
	return 1.0 + (modifier / 100.0)
}

// GetEffectiveManaCostMultiplier returns the mana cost multiplier (1.0 = no change).
func (ms *MutatedSkill) GetEffectiveManaCostMultiplier() float64 {
	modifier := ms.GetTotalModifier(MutationManaCost)
	return 1.0 + (modifier / 100.0)
}

// GetChainTargets returns the number of additional chain targets (0 = no chain).
func (ms *MutatedSkill) GetChainTargets() int {
	targets := 0
	for _, m := range ms.Mutations {
		if m.Type == MutationChain {
			targets += int(m.PrimaryValue)
		}
	}
	return targets
}

// GetLifestealPercent returns the lifesteal percentage (0-100).
func (ms *MutatedSkill) GetLifestealPercent() float64 {
	percent := 0.0
	for _, m := range ms.Mutations {
		if m.Type == MutationLifesteal {
			percent += m.PrimaryValue
		}
	}
	if percent > 100 {
		percent = 100
	}
	return percent
}

// GetEchoChance returns the echo (double cast) chance percentage (0-100).
func (ms *MutatedSkill) GetEchoChance() float64 {
	chance := 0.0
	for _, m := range ms.Mutations {
		if m.Type == MutationEcho {
			chance += m.SecondaryValue
		}
	}
	if chance > 100 {
		chance = 100
	}
	return chance
}

// SkillMutationComponent tracks all skill mutations for an entity.
// This is the main component attached to entities for the mutation system.
type SkillMutationComponent struct {
	// MutatedSkills maps skill ID to its mutation data.
	MutatedSkills map[string]*MutatedSkill `json:"mutated_skills"`
	// AvailableMutations stores mutations the player can apply (inventory).
	AvailableMutations []*SkillMutation `json:"available_mutations"`
	// MaxMutationsPerSkill is the default max mutations per skill.
	MaxMutationsPerSkill int `json:"max_mutations_per_skill"`
	// TotalMutationsApplied tracks total mutations ever applied (for achievements).
	TotalMutationsApplied int `json:"total_mutations_applied"`
	// MutationSlots is how many mutations can be stored in inventory.
	MutationSlots int `json:"mutation_slots"`
	// Dirty flag indicates if mutation effects need recalculation.
	Dirty bool `json:"-"`
}

// Type returns the component type identifier.
func (c *SkillMutationComponent) Type() string {
	return "skill_mutation"
}

// NewSkillMutationComponent creates a new skill mutation component with defaults.
func NewSkillMutationComponent() *SkillMutationComponent {
	return &SkillMutationComponent{
		MutatedSkills:         make(map[string]*MutatedSkill),
		AvailableMutations:    make([]*SkillMutation, 0),
		MaxMutationsPerSkill:  3,
		TotalMutationsApplied: 0,
		MutationSlots:         20,
		Dirty:                 false,
	}
}

// GetMutatedSkill returns the mutation data for a skill, creating if needed.
func (c *SkillMutationComponent) GetMutatedSkill(skillID string) *MutatedSkill {
	if c.MutatedSkills == nil {
		c.MutatedSkills = make(map[string]*MutatedSkill)
	}
	if ms, exists := c.MutatedSkills[skillID]; exists {
		return ms
	}
	ms := &MutatedSkill{
		SkillID:      skillID,
		SpellSlot:    -1,
		Mutations:    make([]*SkillMutation, 0),
		MaxMutations: c.MaxMutationsPerSkill,
		Locked:       false,
	}
	c.MutatedSkills[skillID] = ms
	return ms
}

// GetMutatedSkillBySlot returns mutation data for a spell slot (0-4).
func (c *SkillMutationComponent) GetMutatedSkillBySlot(slot int) *MutatedSkill {
	for _, ms := range c.MutatedSkills {
		if ms.SpellSlot == slot {
			return ms
		}
	}
	return nil
}

// AddMutation applies a mutation to a skill. Returns true on success.
func (c *SkillMutationComponent) AddMutation(skillID string, mutation *SkillMutation) bool {
	ms := c.GetMutatedSkill(skillID)
	if !ms.CanAddMutation() {
		return false
	}
	if ms.HasMutation(mutation.ID) {
		return false
	}
	if !ms.IsMutationCompatible(mutation) {
		return false
	}

	ms.Mutations = append(ms.Mutations, mutation.Copy())
	c.TotalMutationsApplied++
	c.Dirty = true
	return true
}

// RemoveMutation removes a mutation from a skill by mutation ID. Returns true on success.
func (c *SkillMutationComponent) RemoveMutation(skillID, mutationID string) bool {
	ms, exists := c.MutatedSkills[skillID]
	if !exists {
		return false
	}
	if ms.Locked {
		return false
	}

	for i, m := range ms.Mutations {
		if m.ID == mutationID {
			ms.Mutations = append(ms.Mutations[:i], ms.Mutations[i+1:]...)
			c.Dirty = true
			return true
		}
	}
	return false
}

// AddToInventory adds a mutation to the available mutations inventory.
func (c *SkillMutationComponent) AddToInventory(mutation *SkillMutation) bool {
	if len(c.AvailableMutations) >= c.MutationSlots {
		return false
	}
	c.AvailableMutations = append(c.AvailableMutations, mutation.Copy())
	return true
}

// RemoveFromInventory removes a mutation from inventory by index.
func (c *SkillMutationComponent) RemoveFromInventory(index int) *SkillMutation {
	if index < 0 || index >= len(c.AvailableMutations) {
		return nil
	}
	removed := c.AvailableMutations[index]
	c.AvailableMutations = append(c.AvailableMutations[:index], c.AvailableMutations[index+1:]...)
	return removed
}

// GetInventoryCount returns the number of mutations in inventory.
func (c *SkillMutationComponent) GetInventoryCount() int {
	return len(c.AvailableMutations)
}

// GetMutationByID finds a mutation in inventory by ID. Returns index and mutation.
func (c *SkillMutationComponent) GetMutationByID(mutationID string) (int, *SkillMutation) {
	for i, m := range c.AvailableMutations {
		if m.ID == mutationID {
			return i, m
		}
	}
	return -1, nil
}

// GetSkillDamageMultiplier returns damage multiplier for a skill (1.0 = no change).
func (c *SkillMutationComponent) GetSkillDamageMultiplier(skillID string) float64 {
	ms, exists := c.MutatedSkills[skillID]
	if !exists {
		return 1.0
	}
	return ms.GetEffectiveDamageMultiplier()
}

// GetSkillCooldownMultiplier returns cooldown multiplier for a skill.
func (c *SkillMutationComponent) GetSkillCooldownMultiplier(skillID string) float64 {
	ms, exists := c.MutatedSkills[skillID]
	if !exists {
		return 1.0
	}
	return ms.GetEffectiveCooldownMultiplier()
}

// GetSkillManaCostMultiplier returns mana cost multiplier for a skill.
func (c *SkillMutationComponent) GetSkillManaCostMultiplier(skillID string) float64 {
	ms, exists := c.MutatedSkills[skillID]
	if !exists {
		return 1.0
	}
	return ms.GetEffectiveManaCostMultiplier()
}

// GetTotalMutationCount returns the total number of applied mutations across all skills.
func (c *SkillMutationComponent) GetTotalMutationCount() int {
	count := 0
	for _, ms := range c.MutatedSkills {
		count += len(ms.Mutations)
	}
	return count
}

// ClearDirtyFlag resets the dirty flag after recalculation.
func (c *SkillMutationComponent) ClearDirtyFlag() {
	c.Dirty = false
}

// Serialize encodes the component for persistence.
func (c *SkillMutationComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes the component from persistence data.
func (c *SkillMutationComponent) Deserialize(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	// Ensure maps are initialized
	if c.MutatedSkills == nil {
		c.MutatedSkills = make(map[string]*MutatedSkill)
	}
	if c.AvailableMutations == nil {
		c.AvailableMutations = make([]*SkillMutation, 0)
	}
	c.Dirty = true
	return nil
}
