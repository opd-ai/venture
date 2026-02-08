//go:build !android && !ios
// +build !android,!ios

// INTEGRATION FIX [Category A]: V9.0 Server System Initialization
// Gap: V9.0 integration managers (housing crafting, companion housing, guild housing) were client-only
// Fix: Created v9_systems.go for server-side V9.0 integration manager initialization
// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
// Impact: Enables server-authoritative validation of crafting bonuses, companion loyalty, and guild permissions

package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"
	narrativeworld "github.com/opd-ai/venture/pkg/integration/narrative_world"
	"github.com/sirupsen/logrus"
)

// initializeV9SystemsServer initializes Version 9.0 integration managers on the server.
// Server-side managers provide authoritative validation for:
//   - Crafting station bonuses (prevent XP/speed exploits)
//   - Companion housing loyalty calculations (server-validated loyalty gains)
//   - Guild housing permissions (access control enforcement)
//   - Companion narrative events (quest generation, memory tracking, conflicts)
//
// These managers do NOT run as ECS systems, but are used by existing systems
// (CraftingSystem, CompanionLoyaltySystem, NarrativeSystem) to validate client claims.
func initializeV9SystemsServer(world *engine.World, seed int64, logger *logrus.Logger) (
	*housingcrafting.StationManager,
	*companionhousing.PetHomeManager,
	*guildhousing.Manager,
	*narrativeworld.System,
) {
	serverLogger := logger.WithField("component", "v9_systems")

	// Phase 55.1: Crafting Stations & Skill Training
	// Server validates crafting bonus claims from clients
	// Prevents exploits where clients claim higher quality stations than they own
	stationManager := housingcrafting.NewStationManager()

	// Phase 55.2: Companion Housing Interactions
	// Server validates companion loyalty bonus calculations
	// Ensures companions can't gain loyalty from non-existent housing
	petHomeManager := companionhousing.NewPetHomeManager()

	// Phase 55.3: Guild Housing & Communal Spaces
	// Server enforces rank-based access permissions
	// Validates guild resource access (storage, crafting stations, upgrades)
	guildHousingManager := guildhousing.NewManager()

	// Phase 58.1: Companion-Driven Narrative (companion quests, memory-based dialogue, conflicts)
	// Server tracks companion personal quests, memory events, and personality conflicts
	// Enables narrative progression based on companion loyalty and interactions
	narrativeWorldSystem := narrativeworld.NewSystem(world, seed)
	world.AddSystem(narrativeWorldSystem) // Runs as ECS system to track companion interactions

	if logger.GetLevel() >= logrus.DebugLevel {
		serverLogger.WithFields(logrus.Fields{
			"stationManager":       "initialized",
			"petHomeManager":       "initialized",
			"guildHousingManager":  "initialized",
			"narrativeWorldSystem": "initialized",
			"integrationSystems":   4,
			"note":                 "V9 managers + narrative system for server-authoritative validation",
		}).Info("V9.0 integration managers initialized on server (housing crafting, companion housing, guild housing, companion narrative)")
	}

	return stationManager, petHomeManager, guildHousingManager, narrativeWorldSystem
}
