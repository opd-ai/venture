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
	// templates maps genre IDs to their available vehicle templates.
	templates map[string][]VehicleTemplate
	// logger is an optional logrus entry for structured logging.
	logger *logrus.Entry
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

	// Phase 21.2: Generate combat capabilities for rare+ vehicles
	hasCombat := rarity >= RarityRare || rng.Float64() < 0.2 // 20% for common/uncommon
	hasWeapon := hasCombat && (rarity >= RarityEpic || rng.Float64() < 0.3)

	// Phase 21.2: Generate cargo capacity based on vehicle type
	cargoSlots := g.generateCargoSlots(template.VehicleType, rarity, rng)
	cargoWeight := g.generateCargoWeight(template.VehicleType, rarity)

	// Phase 21.2: Generate upgrade slots based on rarity
	upgradeSlots := g.generateUpgradeSlots(rarity)

	// Phase 21.3: Generate visual variation
	decorations := g.generateDecorations(params.GenreID, template.VehicleType, rarity, rng)
	damageState := g.generateDamageState(rng)
	secondaryColor := g.generateSecondaryColor(template.BaseColor, rarity, rng)
	decalPattern := g.generateDecalPattern(params.GenreID, rarity, rng)

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
		// Phase 21.2: Advanced features
		HasCombat:      hasCombat,
		HasWeapon:      hasWeapon,
		WeaponType:     g.generateWeaponType(params.GenreID, template.VehicleType, rng),
		CargoSlots:     cargoSlots,
		CargoWeight:    cargoWeight,
		UpgradeSlots:   upgradeSlots,
		SpecialAbility: g.generateSpecialAbility(params.GenreID, template.VehicleType, rarity, rng),
		// Phase 21.3: Visual variation
		Decorations:    decorations,
		DamageState:    damageState,
		SecondaryColor: secondaryColor,
		DecalPattern:   decalPattern,
	}

	return vehicle
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
		if err := validateVehicleBasics(vehicle, i); err != nil {
			return err
		}

		if err := validateVehicleStats(vehicle, i); err != nil {
			return err
		}
	}

	return nil
}
