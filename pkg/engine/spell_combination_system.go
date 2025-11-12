package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/sirupsen/logrus"
)

// SpellCombinationSystem detects and executes spell combinations.
// It processes entities with SpellComboComponent and monitors spell casting
// to trigger combo effects when compatible spells are cast in sequence.
type SpellCombinationSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
}

// NewSpellCombinationSystem creates a new spell combination system.
func NewSpellCombinationSystem(world *World, rng *rand.Rand) *SpellCombinationSystem {
	return NewSpellCombinationSystemWithLogger(world, rng, nil)
}

// NewSpellCombinationSystemWithLogger creates a new spell combination system with a logger.
func NewSpellCombinationSystemWithLogger(world *World, rng *rand.Rand, logger *logrus.Logger) *SpellCombinationSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "spell_combination")
	}
	return &SpellCombinationSystem{
		world:  world,
		rng:    rng,
		logger: logEntry,
	}
}

// Update processes combo detection and active combo effects.
func (s *SpellCombinationSystem) Update(entities []*Entity, deltaTime float64) {
	currentTime := GetCurrentTime()

	for _, entity := range entities {
		// Only process entities with combo tracking
		comboComp, hasCombo := entity.GetComponent("spell_combo")
		if !hasCombo {
			continue
		}
		combo, ok := comboComp.(*SpellComboComponent)
		if !ok {
			continue
		}

		// Clean old casts
		combo.CleanOldCasts(currentTime)

		// Check for combo opportunities
		if combo.GetRecentCastsCount(currentTime) >= 2 {
			s.detectAndTriggerCombo(entity, combo, currentTime)
		}

		// Update active combo duration - clear if expired
		if combo.ActiveCombo != nil {
			if currentTime >= combo.ActiveCombo.StartTime+combo.ActiveCombo.Duration {
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"entity_id": entity.ID,
						"combo":     combo.ActiveCombo.EffectDescription,
					}).Debug("Combo effect expired")
				}
				combo.ActiveCombo = nil
			}
		}
	}
}

// detectAndTriggerCombo checks for valid combos in recent casts.
func (s *SpellCombinationSystem) detectAndTriggerCombo(entity *Entity, combo *SpellComboComponent, currentTime float64) {
	// Need at least 2 casts for a combo
	if len(combo.RecentCasts) < 2 {
		return
	}

	// Check the last two casts
	cast1 := combo.RecentCasts[len(combo.RecentCasts)-2]
	cast2 := combo.RecentCasts[len(combo.RecentCasts)-1]

	// Check if they're within the combo window
	if cast2.CastTime-cast1.CastTime > combo.ComboWindow {
		return
	}

	// Try element-based synergy first (automatic)
	if synergy := s.checkElementalSynergy(cast1.Element, cast2.Element); synergy != nil {
		s.triggerCombo(entity, combo, cast1, cast2, synergy, currentTime)
		return
	}

	// Try known recipes (discovered combinations)
	if recipe := combo.GetRecipe(cast1.SpellName, cast2.SpellName); recipe != nil {
		s.triggerRecipeCombo(entity, combo, cast1, cast2, recipe, currentTime)
		return
	}

	// Check for incompatible combo (backlash)
	if s.checkIncompatibleCombo(cast1.Element, cast2.Element) {
		s.triggerBacklash(entity, combo, cast1, cast2, currentTime)
	}
}

// ElementalSynergy represents automatic element-based combo bonuses.
type ElementalSynergy struct {
	Element1        string
	Element2        string
	PowerMultiplier float64
	EffectName      string
	Description     string
	Duration        float64
}

// checkElementalSynergy returns synergy data if elements are compatible.
func (s *SpellCombinationSystem) checkElementalSynergy(elem1, elem2 string) *ElementalSynergy {
	// Define elemental synergies (complementary elements boost each other)
	synergies := []ElementalSynergy{
		{Element1: "fire", Element2: "wind", PowerMultiplier: 1.5, EffectName: "Inferno", Description: "Fire and wind create a raging inferno!", Duration: 3.0},
		{Element1: "ice", Element2: "fire", PowerMultiplier: 1.3, EffectName: "Steam Burst", Description: "Ice and fire create scalding steam!", Duration: 2.5},
		{Element1: "lightning", Element2: "earth", PowerMultiplier: 1.6, EffectName: "Seismic Shock", Description: "Lightning and earth cause devastating tremors!", Duration: 3.5},
		{Element1: "water", Element2: "lightning", PowerMultiplier: 1.7, EffectName: "Electrified Deluge", Description: "Water conducts lightning with deadly efficiency!", Duration: 3.0},
		{Element1: "earth", Element2: "fire", PowerMultiplier: 1.4, EffectName: "Molten Core", Description: "Earth and fire create molten destruction!", Duration: 4.0},
		{Element1: "ice", Element2: "wind", PowerMultiplier: 1.5, EffectName: "Blizzard", Description: "Ice and wind summon a freezing blizzard!", Duration: 3.5},
		{Element1: "light", Element2: "dark", PowerMultiplier: 2.0, EffectName: "Eclipse", Description: "Light and dark collide in cosmic power!", Duration: 2.0},
		{Element1: "arcane", Element2: "arcane", PowerMultiplier: 1.8, EffectName: "Arcane Resonance", Description: "Pure magic resonates and amplifies!", Duration: 3.0},
	}

	// Check for matching synergy (supports both orders)
	for _, syn := range synergies {
		if (syn.Element1 == elem1 && syn.Element2 == elem2) ||
			(syn.Element1 == elem2 && syn.Element2 == elem1) {
			return &syn
		}
	}

	return nil
}

// checkIncompatibleCombo returns true if elements clash badly.
func (s *SpellCombinationSystem) checkIncompatibleCombo(elem1, elem2 string) bool {
	// Define incompatible pairs that cause backlash
	incompatible := map[string][]string{
		"fire":      {"ice"},
		"ice":       {"fire"},
		"light":     {"light"}, // Too much light is blinding
		"dark":      {"dark"},  // Too much dark is consuming
		"lightning": {"wind"},  // Lightning disperses in wind
	}

	// Check if elem2 is incompatible with elem1
	if incompList, exists := incompatible[elem1]; exists {
		for _, incomp := range incompList {
			if incomp == elem2 {
				// Random chance for backlash (30%)
				return s.rng.Float64() < 0.3
			}
		}
	}

	return false
}

// triggerCombo activates an elemental synergy combo.
func (s *SpellCombinationSystem) triggerCombo(entity *Entity, combo *SpellComboComponent, cast1, cast2 RecentCast, synergy *ElementalSynergy, currentTime float64) {
	combo.ActiveCombo = &ActiveCombo{
		Spell1Name:        cast1.SpellName,
		Spell2Name:        cast2.SpellName,
		PowerMultiplier:   synergy.PowerMultiplier,
		EffectDescription: synergy.Description,
		StartTime:         currentTime,
		Duration:          synergy.Duration,
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"combo_name": synergy.EffectName,
			"multiplier": synergy.PowerMultiplier,
			"spell1":     cast1.SpellName,
			"spell2":     cast2.SpellName,
		}).Info("Spell combo triggered")
	}

	// Apply visual/audio feedback if systems are available
	s.createComboFeedback(entity, synergy.EffectName, synergy.PowerMultiplier)

	// Clear recent casts to prevent re-triggering
	combo.RecentCasts = []RecentCast{}
}

// triggerRecipeCombo activates a known recipe combo.
func (s *SpellCombinationSystem) triggerRecipeCombo(entity *Entity, combo *SpellComboComponent, cast1, cast2 RecentCast, recipe *ComboRecipe, currentTime float64) {
	combo.ActiveCombo = &ActiveCombo{
		Spell1Name:        cast1.SpellName,
		Spell2Name:        cast2.SpellName,
		PowerMultiplier:   recipe.PowerMultiplier,
		EffectDescription: recipe.ResultEffect,
		StartTime:         currentTime,
		Duration:          3.0, // Default recipe duration
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"recipe":     recipe.ResultEffect,
			"multiplier": recipe.PowerMultiplier,
			"spell1":     cast1.SpellName,
			"spell2":     cast2.SpellName,
		}).Info("Recipe combo triggered")
	}

	// Apply visual/audio feedback
	s.createComboFeedback(entity, recipe.ResultEffect, recipe.PowerMultiplier)

	// Clear recent casts
	combo.RecentCasts = []RecentCast{}
}

// triggerBacklash applies negative effects from incompatible combos.
func (s *SpellCombinationSystem) triggerBacklash(entity *Entity, combo *SpellComboComponent, cast1, cast2 RecentCast, currentTime float64) {
	// Backlash reduces effectiveness instead of boosting
	backlashMultiplier := 0.5 // 50% power reduction

	combo.ActiveCombo = &ActiveCombo{
		Spell1Name:        cast1.SpellName,
		Spell2Name:        cast2.SpellName,
		PowerMultiplier:   backlashMultiplier,
		EffectDescription: "Magical backlash! Incompatible energies clash!",
		StartTime:         currentTime,
		Duration:          2.0,
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"spell1":    cast1.SpellName,
			"spell2":    cast2.SpellName,
		}).Warn("Spell backlash triggered")
	}

	// Apply backlash damage to caster (10% of max health)
	if healthComp, hasHealth := entity.GetComponent("health"); hasHealth {
		if health, ok := healthComp.(*HealthComponent); ok {
			backlashDamage := health.Max * 0.1
			health.Current -= backlashDamage
			if health.Current < 0 {
				health.Current = 0
			}

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"entity_id": entity.ID,
					"damage":    backlashDamage,
				}).Debug("Backlash damage applied")
			}
		}
	}

	// Clear recent casts
	combo.RecentCasts = []RecentCast{}
}

// createComboFeedback creates visual/audio effects for successful combos.
func (s *SpellCombinationSystem) createComboFeedback(entity *Entity, effectName string, multiplier float64) {
	// Get entity position for particle effects
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Create particle burst effect at entity position
	// This would integrate with the particle system
	// For now, we just log it
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"effect_name": effectName,
			"multiplier":  multiplier,
			"x":           pos.X,
			"y":           pos.Y,
		}).Debug("Combo feedback created")
	}
}

// OnSpellCast should be called when a spell is cast to track it for combos.
// This is typically called from the SpellCastingSystem.
func (s *SpellCombinationSystem) OnSpellCast(entity *Entity, spell *magic.Spell, slotIndex int) {
	comboComp, hasCombo := entity.GetComponent("spell_combo")
	if !hasCombo {
		// Create combo component if it doesn't exist
		combo := &SpellComboComponent{
			ComboWindow:  1.0, // 1 second window by default
			RecentCasts:  []RecentCast{},
			KnownRecipes: []ComboRecipe{},
		}
		entity.AddComponent(combo)
		comboComp = combo
	}

	combo, ok := comboComp.(*SpellComboComponent)
	if !ok {
		return
	}

	// Add this cast to recent history
	currentTime := GetCurrentTime()
	combo.AddRecentCast(spell.Name, spell.Element.String(), currentTime, slotIndex)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"spell_name": spell.Name,
			"element":    spell.Element.String(),
			"slot":       slotIndex,
		}).Debug("Spell cast tracked for combo")
	}
}

// GetActiveComboMultiplier returns the power multiplier for the active combo.
// Returns 1.0 if no combo is active.
func (s *SpellCombinationSystem) GetActiveComboMultiplier(entity *Entity) float64 {
	comboComp, hasCombo := entity.GetComponent("spell_combo")
	if !hasCombo {
		return 1.0
	}

	combo, ok := comboComp.(*SpellComboComponent)
	if !ok {
		return 1.0
	}

	currentTime := GetCurrentTime()
	if combo.IsComboActive(currentTime) {
		return combo.ActiveCombo.PowerMultiplier
	}

	return 1.0
}

// DiscoverRecipe adds a new recipe to the entity's known recipes.
// This represents the player discovering a new combo through experimentation.
func (s *SpellCombinationSystem) DiscoverRecipe(entity *Entity, spell1Name, spell2Name, element1, element2, resultEffect string, multiplier float64, symmetric bool) {
	comboComp, hasCombo := entity.GetComponent("spell_combo")
	if !hasCombo {
		// Create combo component if it doesn't exist
		combo := &SpellComboComponent{
			ComboWindow:  1.0,
			RecentCasts:  []RecentCast{},
			KnownRecipes: []ComboRecipe{},
		}
		entity.AddComponent(combo)
		comboComp = combo
	}

	combo, ok := comboComp.(*SpellComboComponent)
	if !ok {
		return
	}

	recipe := ComboRecipe{
		Spell1Name:      spell1Name,
		Spell2Name:      spell2Name,
		Element1:        element1,
		Element2:        element2,
		ResultEffect:    resultEffect,
		PowerMultiplier: multiplier,
		IsSymmetric:     symmetric,
	}

	combo.AddRecipe(recipe)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":    entity.ID,
			"spell1":       spell1Name,
			"spell2":       spell2Name,
			"effect":       resultEffect,
			"multiplier":   multiplier,
			"is_symmetric": symmetric,
		}).Info("Combo recipe discovered")
	}
}
