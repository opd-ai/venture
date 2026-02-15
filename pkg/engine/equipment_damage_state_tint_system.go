// Package engine provides the EquipmentDamageStateTintSystem which aggregates
// the damage states of all equipped items per entity and writes an
// EquipmentWearTintComponent with cumulative visual degradation parameters.
// This bridges sprites.GetDamageVisualEffects with per-entity render state,
// allowing the render pipeline to darken and reduce opacity of worn entities.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// genreWearPreset holds genre-specific wear tint colors.
type genreWearPreset struct {
	R, G, B        float64
	DarkenScale    float64 // Multiplier on aggregate darkening
	DirtinessScale float64 // Multiplier on grime overlay
}

// EquipmentDamageStateTintSystem computes aggregate equipment wear parameters
// from the damage states of all equipped items and writes them to an
// EquipmentWearTintComponent for the render pipeline.
type EquipmentDamageStateTintSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  genreWearPreset

	// Throttle full entity scans for lazy component attachment
	updateInterval float64
	timeSinceCheck float64
}

// NewEquipmentDamageStateTintSystem creates a new equipment damage state tint system.
func NewEquipmentDamageStateTintSystem(world *World, seed int64) *EquipmentDamageStateTintSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_damage_state_tint")
		logEntry.Debug("equipment damage state tint system created")
	}

	sys := &EquipmentDamageStateTintSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.preset = sys.getPreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware wear tint colors and scaling.
func (s *EquipmentDamageStateTintSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for damage state tint")
	}
}

// getPreset returns genre-specific wear tint colors and intensity scaling.
func (s *EquipmentDamageStateTintSystem) getPreset(genreID string) genreWearPreset {
	switch genreID {
	case "horror":
		// Reddish-brown grime, extra darkening
		return genreWearPreset{R: 0.7, G: 0.3, B: 0.2, DarkenScale: 1.3, DirtinessScale: 1.4}
	case "cyberpunk":
		// Yellowish industrial grime, moderate darkening
		return genreWearPreset{R: 0.8, G: 0.7, B: 0.3, DarkenScale: 0.9, DirtinessScale: 1.1}
	case "sci-fi", "scifi":
		// Cool blue-grey wear, minimal dirtiness
		return genreWearPreset{R: 0.5, G: 0.5, B: 0.6, DarkenScale: 0.8, DirtinessScale: 0.6}
	case "post-apocalyptic", "postapoc":
		// Heavy dust/rust, maximum dirtiness
		return genreWearPreset{R: 0.7, G: 0.5, B: 0.3, DarkenScale: 1.2, DirtinessScale: 1.5}
	default: // fantasy
		// Warm brown wear stains
		return genreWearPreset{R: 0.6, G: 0.5, B: 0.4, DarkenScale: 1.0, DirtinessScale: 1.0}
	}
}

// Update iterates entities with equipment and computes aggregate wear tint.
func (s *EquipmentDamageStateTintSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	fullScan := s.timeSinceCheck >= s.updateInterval
	if fullScan {
		s.timeSinceCheck = 0.0
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("equipment_wear_tint")
		tint, hasTint := comp.(*EquipmentWearTintComponent)

		if !hasTint {
			if !fullScan {
				continue
			}
			if !entity.HasComponent("equipment") {
				continue
			}
			tint = NewEquipmentWearTintComponent()
			entity.AddComponent(tint)
		}

		// Recompute aggregate only on full scans (throttled to 1 Hz)
		if fullScan {
			s.computeWearTint(entity, tint)
		}
	}
}

// computeWearTint aggregates damage visual effects from all equipped items.
func (s *EquipmentDamageStateTintSystem) computeWearTint(entity *Entity, tint *EquipmentWearTintComponent) {
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		tint.Enabled = false
		return
	}

	slots := []EquipmentSlot{
		SlotMainHand, SlotOffHand, SlotHead, SlotChest,
		SlotLegs, SlotBoots, SlotGloves,
	}

	var totalOpacity, totalDarken, totalCrack, totalEdge, totalDirt float64
	var count int
	var worstState sprites.DamageState

	for _, slot := range slots {
		itm := equipComp.GetEquipped(slot)
		if itm == nil {
			continue
		}

		state := sprites.GetDamageStateFromDurability(itm.Stats.Durability, itm.Stats.DurabilityMax)
		effects := sprites.GetDamageVisualEffects(state)

		totalOpacity += effects.OpacityMultiplier
		totalDarken += effects.ColorDarken
		totalCrack += effects.CrackDensity
		totalEdge += effects.EdgeRoughness
		totalDirt += effects.Dirtiness
		count++

		if state > worstState {
			worstState = state
		}
	}

	if count == 0 {
		tint.Enabled = false
		tint.OpacityMultiplier = 1.0
		tint.ColorDarken = 0.0
		tint.CrackDensity = 0.0
		tint.EdgeRoughness = 0.0
		tint.Dirtiness = 0.0
		tint.EquippedCount = 0
		tint.WorstState = "none"
		return
	}

	n := float64(count)
	tint.OpacityMultiplier = totalOpacity / n
	tint.ColorDarken = clampFloat(totalDarken/n*s.preset.DarkenScale, 0.0, 1.0)
	tint.CrackDensity = clampFloat(totalCrack/n, 0.0, 1.0)
	tint.EdgeRoughness = clampFloat(totalEdge/n, 0.0, 1.0)
	tint.Dirtiness = clampFloat(totalDirt/n*s.preset.DirtinessScale, 0.0, 1.0)

	tint.TintR = s.preset.R
	tint.TintG = s.preset.G
	tint.TintB = s.preset.B

	tint.EquippedCount = count
	tint.WorstState = worstState.String()

	// Enable only if any item is non-pristine
	tint.Enabled = worstState > sprites.DamageStatePristine
}

// getEquipmentComponent retrieves the typed equipment component from an entity.
func (s *EquipmentDamageStateTintSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}
