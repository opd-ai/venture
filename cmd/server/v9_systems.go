//go:build !android && !ios
// +build !android,!ios

// INTEGRATION FIX [Category A]: V9.0 Server System Initialization
// Gap: V9.0 integration managers (housing crafting, companion housing, guild housing) were client-only
// Fix: Created v9_systems.go for server-side V9.0 integration manager initialization
// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3, Phase 56.3, Phase 58.1)
// Impact: Enables server-authoritative validation of crafting bonuses, companion loyalty, guild permissions, and political warfare

package main

import (
	"github.com/opd-ai/venture/pkg/engine"
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"
	narrativeworld "github.com/opd-ai/venture/pkg/integration/narrative_world"
	politicalwarfare "github.com/opd-ai/venture/pkg/integration/political_warfare"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"
)

// initializeV9SystemsServer initializes Version 9.0 integration managers on the server.
// Server-side managers provide authoritative validation for:
//   - Crafting station bonuses (prevent XP/speed exploits)
//   - Companion housing loyalty calculations (server-validated loyalty gains)
//   - Guild housing permissions (access control enforcement)
//   - Companion narrative events (quest generation, memory tracking, conflicts)
//   - Political warfare (guild wars, treaties, embargoes, diplomatic victories)
//
// These managers do NOT run as ECS systems, but are used by existing systems
// (CraftingSystem, CompanionLoyaltySystem, NarrativeSystem) to validate client claims.
// The PoliticalWarfareSystem runs as an ECS system for time-based updates.
func initializeV9SystemsServer(world *engine.World, seed int64, guildManager *guild.Manager, logger *logrus.Logger) (
	*housingcrafting.StationManager,
	*companionhousing.PetHomeManager,
	*guildhousing.Manager,
	*narrativeworld.System,
	*politicalwarfare.System,
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

	// Phase 56.3: Political Warfare Integration (guild wars, treaties, embargoes)
	// Server manages guild-level political warfare with preparation periods and cooldowns
	// Integrates with V6 Politics, V8 Guilds, and V6 Federation Market
	politicalWarfareSystem := politicalwarfare.NewSystem(world, guildManager, seed)
	world.AddSystem(politicalWarfareSystem) // Runs as ECS system for time-based war/treaty updates

	if logger.GetLevel() >= logrus.DebugLevel {
		serverLogger.WithFields(logrus.Fields{
			"stationManager":         "initialized",
			"petHomeManager":         "initialized",
			"guildHousingManager":    "initialized",
			"narrativeWorldSystem":   "initialized",
			"politicalWarfareSystem": "initialized",
			"integrationSystems":     5,
			"note":                   "V9 managers + narrative/political systems for server-authoritative validation",
		}).Info("V9.0 integration managers initialized on server (housing crafting, companion housing, guild housing, companion narrative, political warfare)")
	}

	return stationManager, petHomeManager, guildHousingManager, narrativeWorldSystem, politicalWarfareSystem
}

// initializeTerritorySystemsServer initializes territory control and siege systems on the server.
// AUDIT.md REM-018: Territory systems must be server-authoritative to prevent exploits:
//   - Territory capture progress validation
//   - Guild war declaration cost enforcement (1000g)
//   - Siege phase progression (Preparation 1h → Assault 2h → Resolution → Ended)
//   - Prevents clients from sending fake capture progress packets
//   - Prevents bypassing war declaration costs via client manipulation
//   - Ensures multiplayer sync of territory state
func initializeTerritorySystemsServer(world *engine.World, logger *logrus.Logger) (*territory.Manager, *engine.TerritorySystem, *territory.SiegeManager, *engine.TerritorySiegeSystem) {
	serverLogger := logger.WithField("component", "territory_systems")

	// Phase 4.3: Territory Control
	// Server manages authoritative territory state
	territoryManager := territory.NewManager()
	territorySystem := engine.NewTerritorySystem(territoryManager, serverLogger.WithField("system", "territory"))
	world.AddSystem(territorySystem)

	// Phase 4.5: Territory Siege System
	// Server validates siege phase transitions and participant eligibility
	siegeManager := territory.NewSiegeManager()
	siegeSystem := engine.NewTerritorySiegeSystem(siegeManager)
	world.AddSystem(siegeSystem)

	if logger.GetLevel() >= logrus.DebugLevel {
		serverLogger.WithFields(logrus.Fields{
			"territorySystem": "initialized",
			"siegeSystem":     "initialized",
			"note":            "Server-authoritative territory control prevents client exploits",
		}).Info("Territory systems initialized on server")
	}

	return territoryManager, territorySystem, siegeManager, siegeSystem
}
