// Package engine provides the skill mutation system for applying and managing mutations.
// This system handles mutation application to spells, effect calculation, and integration
// with the spell casting system.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SkillMutationSystem manages skill mutation effects and application.
// It processes mutations on spells and calculates modified spell stats.
type SkillMutationSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewSkillMutationSystem creates a new skill mutation system.
func NewSkillMutationSystem(world *World) *SkillMutationSystem {
	return NewSkillMutationSystemWithLogger(world, nil)
}

// NewSkillMutationSystemWithLogger creates a new skill mutation system with logging.
func NewSkillMutationSystemWithLogger(world *World, logger *logrus.Logger) *SkillMutationSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system_name", "skill_mutation")
		logEntry.Debug("Creating skill mutation system")
	}
	return &SkillMutationSystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes mutation effects for entities with the mutation component.
// Signature matches System interface: Update(entities []*Entity, deltaTime float64)
func (s *SkillMutationSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("skill_mutation") {
			continue
		}

		comp, ok := entity.GetComponent("skill_mutation")
		if !ok {
			continue
		}
		mutComp, ok := comp.(*SkillMutationComponent)
		if !ok {
			continue
		}

		// Only process if dirty flag is set (mutations changed)
		if !mutComp.Dirty {
			continue
		}

		// Recalculate mutation effects
		s.recalculateMutationEffects(entity, mutComp)
		mutComp.ClearDirtyFlag()

		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":       entity.ID,
				"total_mutations": mutComp.GetTotalMutationCount(),
			}).Debug("Recalculated mutation effects")
		}
	}
}

// recalculateMutationEffects applies mutation modifiers to entity's spell stats.
func (s *SkillMutationSystem) recalculateMutationEffects(entity *Entity, mutComp *SkillMutationComponent) {
	// Apply mutations to spell slots if entity has them
	if entity.HasComponent("spell_slots") {
		s.applySpellSlotMutations(entity, mutComp)
	}
}

// applySpellSlotMutations syncs mutation data with spell slots.
func (s *SkillMutationSystem) applySpellSlotMutations(entity *Entity, mutComp *SkillMutationComponent) {
	comp, ok := entity.GetComponent("spell_slots")
	if !ok {
		return
	}
	slotComp, ok := comp.(*SpellSlotComponent)
	if !ok {
		return
	}

	// Update spell slot assignments in mutated skills
	for slot := 0; slot < 5; slot++ {
		spell := slotComp.GetSlot(slot)
		if spell == nil {
			continue
		}

		skillID := fmt.Sprintf("spell_slot_%d", slot)
		ms := mutComp.GetMutatedSkill(skillID)
		ms.SpellSlot = slot
	}
}

// GenerateMutation creates a random mutation with the given seed and parameters.
func (s *SkillMutationSystem) GenerateMutation(seed int64, rarity MutationRarity, playerLevel int) *SkillMutation {
	rng := rand.New(rand.NewSource(seed))
	mutationType := s.selectMutationType(rng, rarity)
	return s.generateMutationValues(rng, mutationType, rarity, playerLevel, seed)
}

// selectMutationType determines the mutation type based on rarity and random roll.
func (s *SkillMutationSystem) selectMutationType(rng *rand.Rand, rarity MutationRarity) MutationType {
	typeRoll := rng.Intn(100)

	switch rarity {
	case MutationRarityCommon:
		return s.selectCommonMutation(typeRoll)
	case MutationRarityUncommon:
		return s.selectUncommonMutation(typeRoll)
	case MutationRarityRare:
		return s.selectRareMutation(typeRoll)
	case MutationRarityEpic:
		return s.selectEpicMutation(typeRoll)
	case MutationRarityLegendary:
		return s.selectLegendaryMutation(typeRoll)
	default:
		return MutationDamage
	}
}

// selectCommonMutation returns mutation type for common rarity.
func (s *SkillMutationSystem) selectCommonMutation(roll int) MutationType {
	if roll < 40 {
		return MutationDamage
	} else if roll < 70 {
		return MutationCooldown
	}
	return MutationManaCost
}

// selectUncommonMutation returns mutation type for uncommon rarity.
func (s *SkillMutationSystem) selectUncommonMutation(roll int) MutationType {
	switch {
	case roll < 25:
		return MutationDamage
	case roll < 45:
		return MutationCooldown
	case roll < 60:
		return MutationRange
	case roll < 75:
		return MutationArea
	default:
		return MutationDuration
	}
}

// selectRareMutation returns mutation type for rare rarity.
func (s *SkillMutationSystem) selectRareMutation(roll int) MutationType {
	switch {
	case roll < 20:
		return MutationDamage
	case roll < 35:
		return MutationChain
	case roll < 50:
		return MutationLifesteal
	case roll < 65:
		return MutationElemental
	case roll < 80:
		return MutationPierce
	default:
		return MutationRange
	}
}

// selectEpicMutation returns mutation type for epic rarity.
func (s *SkillMutationSystem) selectEpicMutation(roll int) MutationType {
	switch {
	case roll < 25:
		return MutationChain
	case roll < 45:
		return MutationLifesteal
	case roll < 60:
		return MutationSplit
	case roll < 75:
		return MutationEcho
	default:
		return MutationPierce
	}
}

// selectLegendaryMutation returns mutation type for legendary rarity.
func (s *SkillMutationSystem) selectLegendaryMutation(roll int) MutationType {
	switch {
	case roll < 30:
		return MutationEcho
	case roll < 55:
		return MutationSplit
	case roll < 75:
		return MutationChain
	default:
		return MutationLifesteal
	}
}

// generateMutationValues creates the specific values for a mutation.
func (s *SkillMutationSystem) generateMutationValues(rng *rand.Rand, mutationType MutationType, rarity MutationRarity, playerLevel int, seed int64) *SkillMutation {
	// Base values by rarity
	rarityMultiplier := 1.0
	switch rarity {
	case MutationRarityUncommon:
		rarityMultiplier = 1.3
	case MutationRarityRare:
		rarityMultiplier = 1.6
	case MutationRarityEpic:
		rarityMultiplier = 2.0
	case MutationRarityLegendary:
		rarityMultiplier = 2.5
	}

	var primaryValue, secondaryValue, tradeoffValue float64
	var tradeoffType MutationType
	var name, description string
	var incompatible []string

	switch mutationType {
	case MutationDamage:
		// +10% to +30% damage, scaled by rarity
		baseVal := 10.0 + float64(rng.Intn(21))
		primaryValue = baseVal * rarityMultiplier
		// Tradeoff: increased mana cost
		tradeoffType = MutationManaCost
		tradeoffValue = primaryValue * 0.5
		name = s.generateDamageMutationName(rng, rarity)
		description = fmt.Sprintf("+%.0f%% damage, +%.0f%% mana cost", primaryValue, tradeoffValue)

	case MutationCooldown:
		// -10% to -25% cooldown
		baseVal := 10.0 + float64(rng.Intn(16))
		primaryValue = -baseVal * rarityMultiplier
		// Tradeoff: reduced damage
		tradeoffType = MutationDamage
		tradeoffValue = baseVal * 0.4
		name = s.generateCooldownMutationName(rng, rarity)
		description = fmt.Sprintf("%.0f%% cooldown, -%.0f%% damage", primaryValue, tradeoffValue)

	case MutationManaCost:
		// -15% to -35% mana cost
		baseVal := 15.0 + float64(rng.Intn(21))
		primaryValue = -baseVal * rarityMultiplier
		// Tradeoff: reduced range
		tradeoffType = MutationRange
		tradeoffValue = baseVal * 0.3
		name = s.generateManaCostMutationName(rng, rarity)
		description = fmt.Sprintf("%.0f%% mana cost, -%.0f%% range", primaryValue, tradeoffValue)

	case MutationRange:
		// +20% to +50% range
		baseVal := 20.0 + float64(rng.Intn(31))
		primaryValue = baseVal * rarityMultiplier
		// Tradeoff: increased cooldown
		tradeoffType = MutationCooldown
		tradeoffValue = baseVal * 0.3
		name = s.generateRangeMutationName(rng, rarity)
		description = fmt.Sprintf("+%.0f%% range, +%.0f%% cooldown", primaryValue, tradeoffValue)

	case MutationArea:
		// +25% to +60% area
		baseVal := 25.0 + float64(rng.Intn(36))
		primaryValue = baseVal * rarityMultiplier
		// Tradeoff: reduced damage per target
		tradeoffType = MutationDamage
		tradeoffValue = baseVal * 0.4
		name = s.generateAreaMutationName(rng, rarity)
		description = fmt.Sprintf("+%.0f%% area, -%.0f%% damage", primaryValue, tradeoffValue)

	case MutationDuration:
		// +30% to +80% duration
		baseVal := 30.0 + float64(rng.Intn(51))
		primaryValue = baseVal * rarityMultiplier
		// Tradeoff: increased cooldown
		tradeoffType = MutationCooldown
		tradeoffValue = baseVal * 0.25
		name = s.generateDurationMutationName(rng, rarity)
		description = fmt.Sprintf("+%.0f%% duration, +%.0f%% cooldown", primaryValue, tradeoffValue)

	case MutationChain:
		// 1-3 additional targets based on rarity
		chainTargets := 1
		if rarity >= MutationRarityRare {
			chainTargets = 2
		}
		if rarity >= MutationRarityLegendary {
			chainTargets = 3
		}
		primaryValue = float64(chainTargets)
		// Tradeoff: reduced damage per target
		tradeoffType = MutationDamage
		tradeoffValue = float64(chainTargets) * 15
		name = s.generateChainMutationName(rng, rarity)
		description = fmt.Sprintf("Chains to %d additional targets, -%.0f%% damage", chainTargets, tradeoffValue)
		incompatible = []string{"split_"}

	case MutationLifesteal:
		// 5% to 20% lifesteal
		baseVal := 5.0 + float64(rng.Intn(16))
		primaryValue = baseVal * rarityMultiplier
		if primaryValue > 35 {
			primaryValue = 35 // Cap at 35%
		}
		// Tradeoff: increased mana cost
		tradeoffType = MutationManaCost
		tradeoffValue = primaryValue * 0.8
		name = s.generateLifestealMutationName(rng, rarity)
		description = fmt.Sprintf("%.0f%% lifesteal, +%.0f%% mana cost", primaryValue, tradeoffValue)

	case MutationPierce:
		// 10% to 40% resistance ignored
		baseVal := 10.0 + float64(rng.Intn(31))
		primaryValue = baseVal * rarityMultiplier
		if primaryValue > 60 {
			primaryValue = 60 // Cap at 60%
		}
		// Tradeoff: reduced area
		tradeoffType = MutationArea
		tradeoffValue = baseVal * 0.5
		name = s.generatePierceMutationName(rng, rarity)
		description = fmt.Sprintf("Ignores %.0f%% resistance, -%.0f%% area", primaryValue, tradeoffValue)

	case MutationSplit:
		// 2-4 projectiles based on rarity
		splits := 2
		if rarity >= MutationRarityEpic {
			splits = 3
		}
		if rarity >= MutationRarityLegendary {
			splits = 4
		}
		primaryValue = float64(splits)
		// Tradeoff: damage split among projectiles
		tradeoffType = MutationDamage
		tradeoffValue = float64(100 - (100 / splits))
		name = s.generateSplitMutationName(rng, rarity)
		description = fmt.Sprintf("Splits into %d projectiles, each deals %.0f%% damage", splits, 100.0/float64(splits))
		incompatible = []string{"chain_"}

	case MutationEcho:
		// 10% to 30% chance to cast twice
		baseChance := 10.0 + float64(rng.Intn(21))
		secondaryValue = baseChance * rarityMultiplier
		if secondaryValue > 50 {
			secondaryValue = 50 // Cap at 50%
		}
		primaryValue = 0 // Not used for echo
		// Tradeoff: increased cooldown
		tradeoffType = MutationCooldown
		tradeoffValue = secondaryValue * 0.8
		name = s.generateEchoMutationName(rng, rarity)
		description = fmt.Sprintf("%.0f%% chance to cast twice, +%.0f%% cooldown", secondaryValue, tradeoffValue)

	case MutationElemental:
		// Change element to random type
		elements := []int{1, 2, 3, 4, 5, 6, 7, 8} // Fire, Ice, Lightning, Earth, Wind, Light, Dark, Arcane
		secondaryValue = float64(elements[rng.Intn(len(elements))])
		primaryValue = 15 * rarityMultiplier // Bonus damage of new element
		// No tradeoff for element change
		tradeoffType = MutationDamage
		tradeoffValue = 0
		name = s.generateElementalMutationName(rng, rarity, int(secondaryValue))
		description = fmt.Sprintf("Converts to %s element, +%.0f%% elemental damage", s.getElementName(int(secondaryValue)), primaryValue)
	}

	// Generate unique ID
	mutationID := fmt.Sprintf("%s_%d", mutationType.String(), seed)

	return &SkillMutation{
		ID:             mutationID,
		Name:           name,
		Description:    description,
		Type:           mutationType,
		Rarity:         rarity,
		PrimaryValue:   primaryValue,
		SecondaryValue: secondaryValue,
		TradeoffType:   tradeoffType,
		TradeoffValue:  tradeoffValue,
		Seed:           seed,
		RequiredLevel:  s.calculateRequiredLevel(rarity, playerLevel),
		Incompatible:   incompatible,
	}
}

// Name generation helpers
func (s *SkillMutationSystem) generateDamageMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Empowered", "Brutal", "Devastating", "Fierce", "Savage"}
	suffixes := []string{"Force", "Impact", "Strike", "Blow", "Surge"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateCooldownMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Swift", "Rapid", "Quick", "Hastened", "Accelerated"}
	suffixes := []string{"Recovery", "Refresh", "Reset", "Cycle", "Flow"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateManaCostMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Efficient", "Conserved", "Frugal", "Economical", "Balanced"}
	suffixes := []string{"Channeling", "Flow", "Weave", "Cast", "Focus"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateRangeMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Far", "Distant", "Extended", "Reaching", "Longshot"}
	suffixes := []string{"Reach", "Grasp", "Touch", "Strike", "Cast"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateAreaMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Wide", "Expansive", "Broad", "Massive", "Sweeping"}
	suffixes := []string{"Blast", "Wave", "Field", "Zone", "Spread"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateDurationMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Lasting", "Persistent", "Enduring", "Prolonged", "Extended"}
	suffixes := []string{"Effect", "Influence", "Presence", "Aura", "Mark"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateChainMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Chained", "Linked", "Bouncing", "Arcing", "Jumping"}
	suffixes := []string{"Lightning", "Strike", "Cascade", "Surge", "Wave"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateLifestealMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Vampiric", "Draining", "Siphoning", "Leeching", "Absorbing"}
	suffixes := []string{"Touch", "Grasp", "Drain", "Feed", "Hunger"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generatePierceMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Piercing", "Penetrating", "Sundering", "Breaking", "Rending"}
	suffixes := []string{"Strike", "Force", "Edge", "Point", "Tip"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateSplitMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Splitting", "Forking", "Dividing", "Fragmenting", "Scattering"}
	suffixes := []string{"Shot", "Bolt", "Ray", "Beam", "Burst"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateEchoMutationName(rng *rand.Rand, rarity MutationRarity) string {
	prefixes := []string{"Echoing", "Resonant", "Reverberating", "Repeating", "Doubled"}
	suffixes := []string{"Cast", "Strike", "Spell", "Force", "Power"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) generateElementalMutationName(rng *rand.Rand, rarity MutationRarity, element int) string {
	elementPrefixes := map[int][]string{
		1: {"Blazing", "Burning", "Fiery", "Scorching"},      // Fire
		2: {"Frozen", "Glacial", "Icy", "Chilling"},          // Ice
		3: {"Shocking", "Electric", "Voltaic", "Thunderous"}, // Lightning
		4: {"Earthen", "Stone", "Rocky", "Granite"},          // Earth
		5: {"Windswept", "Gale", "Gusty", "Cyclonic"},        // Wind
		6: {"Radiant", "Luminous", "Shining", "Brilliant"},   // Light
		7: {"Shadowy", "Dark", "Void", "Abyssal"},            // Dark
		8: {"Arcane", "Mystic", "Ethereal", "Magical"},       // Arcane
	}
	prefixes := elementPrefixes[element]
	if prefixes == nil {
		prefixes = []string{"Elemental"}
	}
	suffixes := []string{"Infusion", "Conversion", "Transformation", "Shift"}
	return prefixes[rng.Intn(len(prefixes))] + " " + suffixes[rng.Intn(len(suffixes))]
}

func (s *SkillMutationSystem) getElementName(element int) string {
	names := map[int]string{
		1: "Fire", 2: "Ice", 3: "Lightning", 4: "Earth",
		5: "Wind", 6: "Light", 7: "Dark", 8: "Arcane",
	}
	if name, ok := names[element]; ok {
		return name
	}
	return "Unknown"
}

func (s *SkillMutationSystem) calculateRequiredLevel(rarity MutationRarity, playerLevel int) int {
	baseLevel := 1
	switch rarity {
	case MutationRarityUncommon:
		baseLevel = 5
	case MutationRarityRare:
		baseLevel = 15
	case MutationRarityEpic:
		baseLevel = 30
	case MutationRarityLegendary:
		baseLevel = 50
	}
	return baseLevel
}

// ApplyMutationFromInventory applies a mutation from inventory to a skill.
func (s *SkillMutationSystem) ApplyMutationFromInventory(entity *Entity, skillID string, mutationIndex int) error {
	if !entity.HasComponent("skill_mutation") {
		return fmt.Errorf("entity has no skill_mutation component")
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return fmt.Errorf("invalid skill_mutation component")
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return fmt.Errorf("invalid skill_mutation component")
	}

	if mutationIndex < 0 || mutationIndex >= len(mutComp.AvailableMutations) {
		return fmt.Errorf("mutation index out of range")
	}

	mutation := mutComp.AvailableMutations[mutationIndex]
	ms := mutComp.GetMutatedSkill(skillID)

	// Validate
	if !ms.CanAddMutation() {
		return fmt.Errorf("skill cannot accept more mutations (max: %d)", ms.MaxMutations)
	}
	if ms.HasMutation(mutation.ID) {
		return fmt.Errorf("mutation already applied to this skill")
	}
	if !ms.IsMutationCompatible(mutation) {
		return fmt.Errorf("mutation is incompatible with existing mutations")
	}

	// Check player level if we have experience component
	if entity.HasComponent("experience") {
		expComp, hasExp := entity.GetComponent("experience")
		if hasExp {
			if exp, ok := expComp.(*ExperienceComponent); ok {
				if exp.Level < mutation.RequiredLevel {
					return fmt.Errorf("requires level %d (current: %d)", mutation.RequiredLevel, exp.Level)
				}
			}
		}
	}

	// Apply mutation
	if !mutComp.AddMutation(skillID, mutation) {
		return fmt.Errorf("failed to apply mutation")
	}

	// Remove from inventory
	mutComp.RemoveFromInventory(mutationIndex)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"skill_id":    skillID,
			"mutation_id": mutation.ID,
			"rarity":      mutation.Rarity.String(),
		}).Info("Applied mutation to skill")
	}

	return nil
}

// RemoveMutationFromSkill removes a mutation from a skill and returns it to inventory.
func (s *SkillMutationSystem) RemoveMutationFromSkill(entity *Entity, skillID, mutationID string) error {
	if !entity.HasComponent("skill_mutation") {
		return fmt.Errorf("entity has no skill_mutation component")
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return fmt.Errorf("invalid skill_mutation component")
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return fmt.Errorf("invalid skill_mutation component")
	}

	ms, exists := mutComp.MutatedSkills[skillID]
	if !exists {
		return fmt.Errorf("skill not found")
	}

	// Find the mutation
	var removedMutation *SkillMutation
	for _, m := range ms.Mutations {
		if m.ID == mutationID {
			removedMutation = m.Copy()
			break
		}
	}
	if removedMutation == nil {
		return fmt.Errorf("mutation not found on skill")
	}

	// Remove from skill
	if !mutComp.RemoveMutation(skillID, mutationID) {
		return fmt.Errorf("failed to remove mutation")
	}

	// Return to inventory if space available
	if !mutComp.AddToInventory(removedMutation) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   entity.ID,
				"mutation_id": mutationID,
			}).Warn("Inventory full, mutation lost")
		}
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"skill_id":    skillID,
			"mutation_id": mutationID,
		}).Info("Removed mutation from skill")
	}

	return nil
}

// GrantRandomMutation generates and adds a random mutation to entity's inventory.
func (s *SkillMutationSystem) GrantRandomMutation(entity *Entity, seed int64, rarity MutationRarity) (*SkillMutation, error) {
	if !entity.HasComponent("skill_mutation") {
		return nil, fmt.Errorf("entity has no skill_mutation component")
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return nil, fmt.Errorf("invalid skill_mutation component")
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return nil, fmt.Errorf("invalid skill_mutation component")
	}

	// Get player level
	playerLevel := 1
	if entity.HasComponent("experience") {
		expComp, hasExp := entity.GetComponent("experience")
		if hasExp {
			if exp, ok := expComp.(*ExperienceComponent); ok {
				playerLevel = exp.Level
			}
		}
	}

	mutation := s.GenerateMutation(seed, rarity, playerLevel)
	if !mutComp.AddToInventory(mutation) {
		return nil, fmt.Errorf("inventory full")
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"mutation_name": mutation.Name,
			"rarity":        rarity.String(),
		}).Info("Granted mutation to entity")
	}

	return mutation, nil
}

// GetModifiedSpellDamage returns spell damage with mutation modifiers applied.
func (s *SkillMutationSystem) GetModifiedSpellDamage(entity *Entity, spellSlot, baseDamage int) int {
	if !entity.HasComponent("skill_mutation") {
		return baseDamage
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return baseDamage
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return baseDamage
	}

	skillID := fmt.Sprintf("spell_slot_%d", spellSlot)
	multiplier := mutComp.GetSkillDamageMultiplier(skillID)
	return int(float64(baseDamage) * multiplier)
}

// GetModifiedSpellCooldown returns spell cooldown with mutation modifiers applied.
func (s *SkillMutationSystem) GetModifiedSpellCooldown(entity *Entity, spellSlot int, baseCooldown float64) float64 {
	if !entity.HasComponent("skill_mutation") {
		return baseCooldown
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return baseCooldown
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return baseCooldown
	}

	skillID := fmt.Sprintf("spell_slot_%d", spellSlot)
	multiplier := mutComp.GetSkillCooldownMultiplier(skillID)
	return baseCooldown * multiplier
}

// GetModifiedSpellManaCost returns spell mana cost with mutation modifiers applied.
func (s *SkillMutationSystem) GetModifiedSpellManaCost(entity *Entity, spellSlot, baseCost int) int {
	if !entity.HasComponent("skill_mutation") {
		return baseCost
	}

	comp, ok := entity.GetComponent("skill_mutation")
	if !ok {
		return baseCost
	}
	mutComp, ok := comp.(*SkillMutationComponent)
	if !ok {
		return baseCost
	}

	skillID := fmt.Sprintf("spell_slot_%d", spellSlot)
	multiplier := mutComp.GetSkillManaCostMultiplier(skillID)
	return int(float64(baseCost) * multiplier)
}
