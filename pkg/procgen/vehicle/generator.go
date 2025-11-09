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

	// Phase 21.2: Generate combat capabilities for rare+ vehicles
	hasCombat := rarity >= RarityRare || rng.Float64() < 0.2 // 20% for common/uncommon
	hasWeapon := hasCombat && (rarity >= RarityEpic || rng.Float64() < 0.3)

	// Phase 21.2: Generate cargo capacity based on vehicle type
	cargoSlots := g.generateCargoSlots(template.VehicleType, rarity, rng)
	cargoWeight := g.generateCargoWeight(template.VehicleType, rarity)

	// Phase 21.2: Generate upgrade slots based on rarity
	upgradeSlots := g.generateUpgradeSlots(rarity)

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
	}

	return vehicle
}

// generateCargoSlots determines cargo capacity based on vehicle type and rarity.
// Phase 21.2: Cargo/Passenger System
func (g *VehicleGenerator) generateCargoSlots(vehicleType VehicleType, rarity Rarity, rng *rand.Rand) int {
	baseSlots := 0
	switch vehicleType {
	case TypeMount:
		baseSlots = 2 // Saddlebags
	case TypeCart:
		baseSlots = 20 // Large cargo area
	case TypeBoat:
		baseSlots = 10 // Medium storage
	case TypeGlider:
		baseSlots = 1 // Minimal storage
	case TypeMech:
		baseSlots = 5 // Internal compartment
	}

	// Scale with rarity (10-50 range)
	rarityBonus := int(float64(baseSlots) * (rarity.GetMultiplier() - 1.0) * 0.5)
	totalSlots := baseSlots + rarityBonus + rng.Intn(5) // +0 to 4 random

	// Clamp to valid range
	if totalSlots < 1 {
		totalSlots = 1
	}
	if totalSlots > 50 {
		totalSlots = 50
	}

	return totalSlots
}

// generateCargoWeight determines weight capacity based on vehicle type and rarity.
// Phase 21.2: Cargo/Passenger System
func (g *VehicleGenerator) generateCargoWeight(vehicleType VehicleType, rarity Rarity) float64 {
	baseWeight := 0.0
	switch vehicleType {
	case TypeMount:
		baseWeight = 50.0 // kg
	case TypeCart:
		baseWeight = 500.0 // kg
	case TypeBoat:
		baseWeight = 200.0 // kg
	case TypeGlider:
		baseWeight = 10.0 // kg
	case TypeMech:
		baseWeight = 100.0 // kg
	}

	return baseWeight * rarity.GetMultiplier()
}

// generateUpgradeSlots determines upgrade slot count based on rarity.
// Phase 21.2: Upgrade System
func (g *VehicleGenerator) generateUpgradeSlots(rarity Rarity) int {
	switch rarity {
	case RarityCommon:
		return 2
	case RarityUncommon:
		return 3
	case RarityRare:
		return 4
	case RarityEpic:
		return 5
	case RarityLegendary:
		return 6
	default:
		return 2
	}
}

// generateWeaponType selects a weapon type based on genre and vehicle type.
// Phase 21.2: Vehicle Combat
func (g *VehicleGenerator) generateWeaponType(genreID string, vehicleType VehicleType, rng *rand.Rand) string {
	genreWeapons := map[string][]string{
		"fantasy":   {"Ballista", "Catapult", "Magic Crystal", "Flame Thrower"},
		"scifi":     {"Laser", "Plasma Cannon", "Railgun", "Missile Launcher"},
		"horror":    {"Soul Reaper", "Bone Spikes", "Curse Projector", "Blood Cannon"},
		"cyberpunk": {"Smartgun", "EMP Launcher", "Plasma Rifle", "Hacking Beam"},
		"postapoc":  {"Machine Gun", "Flamethrower", "Scrap Cannon", "Spike Launcher"},
	}

	weapons := genreWeapons[genreID]
	if weapons == nil {
		weapons = genreWeapons["fantasy"] // Default
	}

	return weapons[rng.Intn(len(weapons))]
}

// generateSpecialAbility creates a genre-specific special ability.
// Phase 21.2: Genre-Specific Features
func (g *VehicleGenerator) generateSpecialAbility(genreID string, vehicleType VehicleType, rarity Rarity, rng *rand.Rand) string {
	// Only legendary and epic vehicles get special abilities
	if rarity < RarityEpic {
		return ""
	}

	genreAbilities := map[string][]string{
		"fantasy": {
			"Teleport Dash: Instantly blink 20 meters forward",
			"Holy Shield: Temporary invulnerability for 3 seconds",
			"Dragon Breath: Exhale flames in a cone",
			"Wind Burst: Create gust that knocks back enemies",
		},
		"scifi": {
			"Cloaking Device: Become invisible for 5 seconds",
			"Shield Generator: Deploy energy shield",
			"Overdrive: Double speed for 10 seconds",
			"EMP Pulse: Disable nearby electronics",
		},
		"horror": {
			"Soul Harvest: Drain life from nearby enemies",
			"Shadow Meld: Pass through walls for 2 seconds",
			"Terror Aura: Fear nearby enemies",
			"Corpse Explosion: Detonate nearby corpses",
		},
		"cyberpunk": {
			"Neural Jack: Hack nearby vehicles",
			"Time Dilation: Slow time for 3 seconds",
			"Neon Afterburner: Leave damaging trail",
			"Hologram Decoy: Create fake duplicate",
		},
		"postapoc": {
			"Nitro Boost: Extreme speed burst",
			"Scrap Armor: Temporary damage reduction",
			"Radiation Pulse: Area damage over time",
			"Salvage Drone: Auto-collect nearby items",
		},
	}

	abilities := genreAbilities[genreID]
	if abilities == nil {
		abilities = genreAbilities["fantasy"] // Default
	}

	return abilities[rng.Intn(len(abilities))]
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
