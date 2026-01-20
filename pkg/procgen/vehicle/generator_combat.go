// Package vehicle provides combat-related generation functions.
// This file contains weapon type selection and special ability generation logic
// for vehicles with combat capabilities (rare+ tier or 20% chance for common/uncommon).
//
// Weapons and abilities are genre-specific, ensuring thematic consistency:
// - Fantasy: Ballista, Dragon Breath, Holy Shield
// - Sci-Fi: Railgun, Cloaking Device, EMP Pulse
// - Horror: Soul Reaper, Terror Aura, Corpse Explosion
// - Cyberpunk: Smartgun, Neural Jack, Time Dilation
// - Post-Apocalyptic: Flamethrower, Nitro Boost, Radiation Pulse
//
// All functions moved from generator.go during Phase 3 reorganization to group
// combat-related functionality separately from core generation logic.
//
// Code relocated from: generator.go
// Phase 21.2: Vehicle Combat & Genre-Specific Features
package vehicle

import (
	"math/rand"
)

// generateWeaponType selects a weapon type based on genre and vehicle type.
// Phase 21.2: Vehicle Combat
// Originally from: generator.go
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
// Originally from: generator.go
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
