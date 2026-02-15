// Package engine provides the EquipmentRarityDetailSystem which reads the
// rarity tiers of all equipped items per entity, computes aggregate visual
// detail parameters, and writes a RarityDetailComponent. The render pipeline
// uses this to scale shape complexity, color vibrancy, border quality, and
// material fidelity proportional to equipment rarity.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// rarityDetailLevel maps item.Rarity to the detail level float (0.0–1.0).
var rarityDetailLevel = map[item.Rarity]float64{
	item.RarityCommon:    0.3,
	item.RarityUncommon:  0.4,
	item.RarityRare:      0.6,
	item.RarityEpic:      0.8,
	item.RarityLegendary: 1.0,
}

// genreRarityPreset holds genre-specific multipliers for rarity detail scaling.
type genreRarityPreset struct {
	ComplexityScale float64 // Multiplier on shape complexity
	VibrancyScale   float64 // Multiplier on color vibrancy
	SharpnessScale  float64 // Multiplier on border sharpness
	FidelityScale   float64 // Multiplier on material fidelity
}

// EquipmentRarityDetailSystem computes aggregate visual detail level from the
// rarity tiers of equipped items and lazily attaches a RarityDetailComponent.
type EquipmentRarityDetailSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  genreRarityPreset

	updateInterval float64
	timeSinceCheck float64
}

// NewEquipmentRarityDetailSystem creates a new equipment rarity detail system.
func NewEquipmentRarityDetailSystem(world *World, seed int64) *EquipmentRarityDetailSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_rarity_detail")
		logEntry.Debug("equipment rarity detail system created")
	}

	sys := &EquipmentRarityDetailSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.preset = sys.getPreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware detail scaling multipliers.
func (s *EquipmentRarityDetailSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for rarity detail")
	}
}

// getPreset returns genre-specific multipliers for rarity visual detail scaling.
func (s *EquipmentRarityDetailSystem) getPreset(genreID string) genreRarityPreset {
	switch genreID {
	case "horror":
		// Muted vibrancy, rough borders, high material fidelity for grime
		return genreRarityPreset{ComplexityScale: 0.9, VibrancyScale: 0.6, SharpnessScale: 0.7, FidelityScale: 1.2}
	case "cyberpunk":
		// High vibrancy and sharpness for neon aesthetic
		return genreRarityPreset{ComplexityScale: 1.1, VibrancyScale: 1.3, SharpnessScale: 1.2, FidelityScale: 1.0}
	case "sci-fi", "scifi":
		// Clean lines, moderate vibrancy
		return genreRarityPreset{ComplexityScale: 1.0, VibrancyScale: 1.0, SharpnessScale: 1.3, FidelityScale: 1.1}
	case "post-apocalyptic", "postapoc":
		// Low vibrancy, low sharpness for worn aesthetic
		return genreRarityPreset{ComplexityScale: 0.8, VibrancyScale: 0.5, SharpnessScale: 0.6, FidelityScale: 1.3}
	case "fantasy":
		return genreRarityPreset{ComplexityScale: 1.0, VibrancyScale: 1.0, SharpnessScale: 1.0, FidelityScale: 1.0}
	default:
		return genreRarityPreset{ComplexityScale: 1.0, VibrancyScale: 1.0, SharpnessScale: 1.0, FidelityScale: 1.0}
	}
}

// Update iterates entities with equipment and computes rarity-based detail levels.
func (s *EquipmentRarityDetailSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	fullScan := s.timeSinceCheck >= s.updateInterval
	if fullScan {
		s.timeSinceCheck = 0.0
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("rarity_detail")
		detail, hasDetail := comp.(*RarityDetailComponent)

		if !hasDetail {
			if !fullScan {
				continue
			}
			if !entity.HasComponent("equipment") {
				continue
			}
			detail = NewRarityDetailComponent()
			entity.AddComponent(detail)
		}

		if fullScan {
			s.computeRarityDetail(entity, detail)
		}
	}
}

// computeRarityDetail aggregates rarity levels from all equipped items.
func (s *EquipmentRarityDetailSystem) computeRarityDetail(entity *Entity, detail *RarityDetailComponent) {
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		detail.Enabled = false
		return
	}

	slots := []EquipmentSlot{
		SlotMainHand, SlotOffHand, SlotHead, SlotChest,
		SlotLegs, SlotBoots, SlotGloves,
	}

	var totalDetail float64
	var count int
	var highestRarity item.Rarity

	for _, slot := range slots {
		itm := equipComp.GetEquipped(slot)
		if itm == nil {
			continue
		}

		dl, ok := rarityDetailLevel[itm.Rarity]
		if !ok {
			dl = 0.3 // Default to common
		}
		totalDetail += dl
		count++

		if itm.Rarity > highestRarity {
			highestRarity = itm.Rarity
		}
	}

	if count == 0 {
		detail.Enabled = false
		detail.DetailLevel = 0.3
		detail.ShapeComplexity = 0.3
		detail.ColorVibrancy = 0.3
		detail.BorderSharpness = 0.3
		detail.MaterialFidelity = 0.3
		detail.HighestRarity = "none"
		detail.EquippedCount = 0
		return
	}

	avgDetail := totalDetail / float64(count)

	// The highest rarity item pulls the overall detail upward (weighted blend)
	peakDetail := rarityDetailLevel[highestRarity]
	blendedDetail := avgDetail*0.6 + peakDetail*0.4

	detail.DetailLevel = clampFloat(blendedDetail, 0.0, 1.0)
	detail.ShapeComplexity = clampFloat(blendedDetail*s.preset.ComplexityScale, 0.0, 1.0)
	detail.ColorVibrancy = clampFloat(blendedDetail*s.preset.VibrancyScale, 0.0, 1.0)
	detail.BorderSharpness = clampFloat(blendedDetail*s.preset.SharpnessScale, 0.0, 1.0)
	detail.MaterialFidelity = clampFloat(blendedDetail*s.preset.FidelityScale, 0.0, 1.0)
	detail.HighestRarity = highestRarity.String()
	detail.EquippedCount = count
	detail.Enabled = true
}

// getEquipmentComponent retrieves the typed equipment component from an entity.
func (s *EquipmentRarityDetailSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
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
