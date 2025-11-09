// Package vehicle provides procedural vehicle generation.
// This file implements the vehicle generator with stat scaling and validation.
//
// Phase 21.1: Vehicle Foundation
package vehicle

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// VehicleGenerator generates procedural vehicles.
type VehicleGenerator struct {
	templates map[string][]VehicleTemplate
	logger    *logrus.Entry
}

// NewVehicleGenerator creates a new vehicle generator.
func NewVehicleGenerator() *VehicleGenerator {
	return NewVehicleGeneratorWithLogger(nil)
}

// NewVehicleGeneratorWithLogger creates a new vehicle generator with a logger.
func NewVehicleGeneratorWithLogger(logger *logrus.Logger) *VehicleGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "vehicle")
	}

	gen := &VehicleGenerator{
		templates: make(map[string][]VehicleTemplate),
		logger:    logEntry,
	}

	// Register genre templates
	gen.templates["fantasy"] = GetFantasyTemplates()
	gen.templates["scifi"] = GetSciFiTemplates()
	gen.templates["horror"] = GetHorrorTemplates()
	gen.templates["cyberpunk"] = GetCyberpunkTemplates()
	gen.templates["postapoc"] = GetPostApocTemplates()
	gen.templates[""] = GetFantasyTemplates() // default

	if logEntry != nil {
		logEntry.Debug("vehicle generator initialized")
	}

	return gen
}

// Generate creates vehicles based on the seed and parameters.
// Returns []*Vehicle or error.
func (g *VehicleGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	count := 5
	if params.Custom != nil {
		if c, ok := params.Custom["count"].(int); ok {
			count = c
		}
	}

	templates := g.templates[params.GenreID]
	if templates == nil {
		templates = g.templates[""]
	}

	rng := rand.New(rand.NewSource(seed))

	vehicles := make([]*Vehicle, count)
	for i := 0; i < count; i++ {
		vehicleSeed := seed + int64(i)*1000
		vehicles[i] = g.generateSingleVehicle(vehicleSeed, params, templates, rng)
	}

	return vehicles, nil
}

// generateSingleVehicle creates one vehicle.
func (g *VehicleGenerator) generateSingleVehicle(seed int64, params procgen.GenerationParams, templates []VehicleTemplate, rng *rand.Rand) *Vehicle {
	// Select template
	template := templates[rng.Intn(len(templates))]

	// Determine rarity based on depth
	rarity := g.determineRarity(params.Depth, rng)

	// Scale stats based on depth, difficulty, and rarity
	depthScale := 1.0 + (float64(params.Depth) * 0.1)
	difficultyScale := 0.5 + (params.Difficulty * 0.5) // 0.5x to 1.0x
	rarityScale := rarity.GetMultiplier()
	totalScale := depthScale * difficultyScale * rarityScale

	// Generate vehicle
	vehicle := &Vehicle{
		Name:          g.generateName(template, rarity, rng),
		Description:   template.DescriptionTemplate,
		VehicleType:   template.VehicleType,
		Rarity:        rarity,
		MaxSpeed:      template.BaseMaxSpeed * totalScale,
		Acceleration:  template.BaseAcceleration * totalScale,
		Handling:      template.BaseHandling * (1.0 + (rarity.GetMultiplier()-1.0)*0.2), // Handling scales less with rarity
		MaxDurability: template.BaseDurability * totalScale,
		FuelCapacity:  template.BaseFuelCapacity * totalScale,
		Capacity:      template.BaseCapacity,
		FuelType:      template.FuelType,
		GenreID:       params.GenreID,
		Depth:         params.Depth,
		Color:         g.generateColor(template.BaseColor, rarity, rng),
	}

	return vehicle
}

// determineRarity calculates rarity based on depth and randomness.
func (g *VehicleGenerator) determineRarity(depth int, rng *rand.Rand) Rarity {
	// Base probabilities
	roll := rng.Float64()

	// Depth increases rare chances
	depthBonus := float64(depth) * 0.02 // +2% per depth level

	// Common: 60% - depth*2%
	// Uncommon: 25%
	// Rare: 10% + depth*1%
	// Epic: 4% + depth*0.8%
	// Legendary: 1% + depth*0.2%

	commonChance := 0.60 - depthBonus
	uncommonChance := commonChance + 0.25
	rareChance := uncommonChance + 0.10 + (depthBonus * 0.5)
	epicChance := rareChance + 0.04 + (depthBonus * 0.4)
	// Legendary is everything above epicChance

	if roll < commonChance {
		return RarityCommon
	} else if roll < uncommonChance {
		return RarityUncommon
	} else if roll < rareChance {
		return RarityRare
	} else if roll < epicChance {
		return RarityEpic
	}
	return RarityLegendary
}

// generateName creates a name from template and rarity.
func (g *VehicleGenerator) generateName(template VehicleTemplate, rarity Rarity, rng *rand.Rand) string {
	prefix := template.NamePrefix
	suffix := template.NameSuffix

	// Add rarity modifier to uncommon and above
	if rarity >= RarityUncommon {
		modifiers := []string{"Enhanced", "Superior", "Reinforced", "Advanced", "Masterwork"}
		prefix = modifiers[rng.Intn(len(modifiers))] + " " + prefix
	}

	return prefix + " " + suffix
}

// generateColor varies the base color based on rarity.
func (g *VehicleGenerator) generateColor(baseColor uint32, rarity Rarity, rng *rand.Rand) uint32 {
	// Extract RGB components
	r := float64((baseColor >> 16) & 0xFF)
	green := float64((baseColor >> 8) & 0xFF)
	b := float64(baseColor & 0xFF)

	// Apply rarity-based variation
	variation := 30.0
	if rarity >= RarityRare {
		variation = 50.0 // More variation for rare+
	}

	r += (rng.Float64()*2.0 - 1.0) * variation
	green += (rng.Float64()*2.0 - 1.0) * variation
	b += (rng.Float64()*2.0 - 1.0) * variation

	// Clamp to valid range
	r = clamp(r, 0, 255)
	green = clamp(green, 0, 255)
	b = clamp(b, 0, 255)

	return (uint32(r) << 16) | (uint32(green) << 8) | uint32(b)
}

// Validate checks if generated vehicles meet quality standards.
func (g *VehicleGenerator) Validate(result interface{}) error {
	vehicles, ok := result.([]*Vehicle)
	if !ok {
		return fmt.Errorf("result is not []*Vehicle")
	}

	if len(vehicles) == 0 {
		return fmt.Errorf("no vehicles generated")
	}

	for i, vehicle := range vehicles {
		if vehicle == nil {
			return fmt.Errorf("vehicle %d is nil", i)
		}

		// Validate stats are positive
		if vehicle.MaxSpeed <= 0 {
			return fmt.Errorf("vehicle %d has invalid MaxSpeed: %f", i, vehicle.MaxSpeed)
		}
		if vehicle.Acceleration <= 0 {
			return fmt.Errorf("vehicle %d has invalid Acceleration: %f", i, vehicle.Acceleration)
		}
		if vehicle.Handling <= 0 {
			return fmt.Errorf("vehicle %d has invalid Handling: %f", i, vehicle.Handling)
		}
		if vehicle.MaxDurability <= 0 {
			return fmt.Errorf("vehicle %d has invalid MaxDurability: %f", i, vehicle.MaxDurability)
		}
		if vehicle.FuelCapacity <= 0 {
			return fmt.Errorf("vehicle %d has invalid FuelCapacity: %f", i, vehicle.FuelCapacity)
		}
		if vehicle.Capacity <= 0 {
			return fmt.Errorf("vehicle %d has invalid Capacity: %d", i, vehicle.Capacity)
		}

		// Validate name is not empty
		if vehicle.Name == "" {
			return fmt.Errorf("vehicle %d has empty name", i)
		}
	}

	return nil
}

// clamp constrains a value to a range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
