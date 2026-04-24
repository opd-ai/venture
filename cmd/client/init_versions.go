//go:build !android && !ios
// +build !android,!ios

// init_versions.go contains version-specific system initialization functions.
// These functions initialize systems for different game versions (V4-V19, VR, Phase 3).
// Extracted from handlers.go to improve maintainability and code organization.

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/physics/vehicle"
	"github.com/opd-ai/venture/pkg/engine/qol"
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"
	"github.com/opd-ai/venture/pkg/integration/narrative_world"
	"github.com/opd-ai/venture/pkg/integration/political_warfare"
	"github.com/opd-ai/venture/pkg/integration/trade_routes"
	"github.com/opd-ai/venture/pkg/integration/world_events"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network/chat"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	mobilefed "github.com/opd-ai/venture/pkg/network/federation/mobile"
	"github.com/opd-ai/venture/pkg/network/trade"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/dialog"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/procgen/legendary"
	"github.com/opd-ai/venture/pkg/procgen/minigame/games"
	"github.com/opd-ai/venture/pkg/rendering/display"
	"github.com/opd-ai/venture/pkg/social/persistence"
	"github.com/opd-ai/venture/pkg/vr"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/opd-ai/venture/pkg/world/economy"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"

	// G3 (AUDIT.md): mod browser install/uninstall callbacks
	"github.com/opd-ai/venture/pkg/modding"
)

// initializeV4Systems initializes Version 4.0 systems (Phase 21-27).
func initializeV4Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 21: Vehicle systems
	sys.vehicleMovementSys = engine.NewVehicleMovementSystem(game.World)
	sys.vehicleDurabilitySys = engine.NewVehicleDurabilitySystem(game.World)
	sys.mountingSystem = engine.NewMountingSystem(game.World)
	sys.vehicleCombatSystem = engine.NewVehicleCombatSystem(game.World)

	// Phase 22: Companion systems
	sys.companionAISystem = engine.NewCompanionAISystem(game.World)
	sys.companionProgressionSys = engine.NewCompanionProgressionSystem(game.World)
	sys.companionQuestSynergySys = engine.NewCompanionQuestSynergySystem(game.World)
	sys.companionLoyaltySys = engine.NewCompanionLoyaltySystem(game.World, clientLogger.Logger)
	sys.companionInventorySys = engine.NewCompanionInventorySystem(game.World)
	sys.companionLearningSys = engine.NewCompanionLearningSystem(game.World) // Phase 4.1: Companion learning
	sys.advancedClassSystem = engine.NewAdvancedClassSystem(game.World)      // Phase 4.2: Advanced classes
	sys.skillInheritanceSys = engine.NewSkillInheritanceSystem(game.World)

	// Phase 23: Book system
	sys.bookReadingSystem = engine.NewBookReadingSystem(game.World)

	// Phase 24: Expanded magic systems (reuse statusEffectRNG for consistency)
	spellRNG := rand.New(rand.NewSource(*seed + seedOffsetSpellEffects))
	sys.spellEffectSystem = engine.NewSpellEffectSystem(game.World, spellRNG)
	sys.spellCombinationSys = engine.NewSpellCombinationSystem(game.World, spellRNG)

	// Phase 25: Class progression system
	sys.classProgressionSys = engine.NewClassProgressionSystem()

	// Phase 25.2a: Dual-class synergy system - connects dual-class combinations with passive bonuses
	sys.dualClassSynergySys = engine.NewDualClassSynergySystem(game.World, *seed+seedOffsetDualClassSynergy)
	sys.dualClassSynergySys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "dual_class_synergy").Debug("Created dual-class synergy system")

	// Phase 25a: Specialization mana boost system - connects class specialization with mana regen
	sys.specializationManaBoostSys = engine.NewSpecializationManaBoostSystem(game.World, *seed+seedOffsetSpecManaBoost)
	sys.specializationManaBoostSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_mana_boost").Debug("Created specialization mana boost system")

	// Phase 25b: Specialization health regen system - connects class specialization with health regen
	sys.specializationHealthRegenSys = engine.NewSpecializationHealthRegenSystem(game.World, *seed+seedOffsetSpecHealthRegen)
	sys.specializationHealthRegenSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_health_regen").Debug("Created specialization health regen system")

	// Phase 25c: Specialization spell damage system - connects class specialization with spell damage
	sys.specializationSpellDamageSys = engine.NewSpecializationSpellDamageSystem(game.World, *seed+seedOffsetSpecSpellDamage)
	sys.specializationSpellDamageSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_spell_damage").Debug("Created specialization spell damage system")

	// Phase 25d: Specialization attack speed system - connects class specialization with attack cooldown
	sys.specializationAttackSpeedSys = engine.NewSpecializationAttackSpeedSystem(game.World, *seed+seedOffsetSpecAttackSpeed)
	sys.specializationAttackSpeedSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_attack_speed").Debug("Created specialization attack speed system")

	// Phase 25e: Specialization defense system - connects class specialization with defense bonuses
	sys.specializationDefenseSys = engine.NewSpecializationDefenseSystem(game.World, *seed+seedOffsetSpecDefense)
	sys.specializationDefenseSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_defense").Debug("Created specialization defense system")

	// Phase 25f: Specialization lifesteal system - connects class specialization with lifesteal bonuses
	sys.specializationLifestealSys = engine.NewSpecializationLifestealSystem(game.World, *seed+seedOffsetSpecLifesteal)
	sys.specializationLifestealSys.SetGenre(*genreID)
	logging.ComponentLogger(clientLogger.Logger, "specialization_lifesteal").Debug("Created specialization lifesteal system")

	// Phase 26: Expression systems (requires audio manager)
	sys.expressionSystem = engine.NewExpressionSystem(game.World, sys.audioManager)
	sys.expressionComboSys = engine.NewExpressionComboSystem(game.World)

	// Phase 27: Mini-game system
	sys.miniGameSystem = engine.NewMiniGameSystem(game.World)

	// Phase 3.4: Minigame implementations
	sys.minigameGamesSystem = games.NewSystem(game.World)
	logging.ComponentLogger(clientLogger.Logger, "minigame_games").Debug("Created minigame games system")

	// Phase 4.1: Choice & consequences system
	sys.choiceConsequencesSystem = engine.NewChoiceConsequencesSystem(game.World)
	logging.ComponentLogger(clientLogger.Logger, "choice_consequences").Debug("Created choice consequences system")

	// Phase 4.2: Guild vehicle integration
	sys.guildVehicleSystem = engine.NewGuildVehicleSystem(game.World)
	logging.ComponentLogger(clientLogger.Logger, "guild_vehicle").Debug("Created guild vehicle system")

	// Phase 4.3: Narrative-world integration
	sys.narrativeWorldSystem = narrative_world.NewSystem(game.World, game.GetWorldSeed())
	logging.ComponentLogger(clientLogger.Logger, "narrative_world").Debug("Created narrative world system")

	// Phase 26.2: Achievement system (social features)
	sys.achievementSystem = engine.NewAchievementSystem(game.World)

	// INTEGRATION FIX [Category A]: Phase 28 - MoralChoiceSystem
	// Gap: MoralChoiceSystem implemented but never initialized or registered
	// Fix: Added system initialization for moral decision tracking and consequences
	// Roadmap: ROADMAP_V4.md Phase 28
	sys.moralChoiceSystem = engine.NewMoralChoiceSystem(game.World, clientLogger.Logger)

	// Phase 30: Environmental Storytelling - Discovery System
	sys.discoverySystem = engine.NewDiscoverySystem(game.World)

	// INTEGRATION FIX [Category A]: Phase 30 - InvestigationSystem
	// Gap: InvestigationSystem implemented but never initialized
	// Fix: Added system initialization for investigating environment and revealing hidden clues
	// Roadmap: ROADMAP_V4.md Phase 30.2
	sys.investigationSystem = engine.NewInvestigationSystem(game.World, *seed+seedOffsetInvestigation)

	// INTEGRATION FIX [Category A]: Phase 31 - NPCDialogSystem
	// Gap: NPCDialogSystem implemented but never initialized
	// Fix: Added system initialization for Markov-chain NPC conversations
	// Roadmap: ROADMAP_V5.md Phase 31
	sys.npcDialogSystem = engine.NewNPCDialogSystem(game.World, *seed+seedOffsetNPCDialog)

	// INTEGRATION FIX [Category A]: Phase 29 - AdaptiveSoundtrackSystem
	// Gap: AdaptiveSoundtrackSystem fully implemented but never initialized
	// Fix: Added system initialization for dynamic music adaptation based on combat, exploration, discovery
	// Roadmap: ROADMAP_V4.md Phase 29
	sys.adaptiveSoundtrackSystem = engine.NewAdaptiveSoundtrackSystem(game.World)

	// INTEGRATION FIX [Category A]: Phase 22.2 - CompanionSystem high-level wrapper
	// Gap: CompanionSystem implemented but never initialized (subsystems CompanionAISystem, etc. initialized above)
	// Fix: Added high-level CompanionSystem for unified companion management
	// Roadmap: ROADMAP_V4.md Phase 22.2
	sys.companionSystem = engine.NewCompanionSystem(game.World)

	// Phase 1.1: Companion Learning System (PLAN.md Integration - Week 1)
	// Unconditional activation: AI companion behavior adaptation and learning baseline feature
	sys.companionLearningSystem = learning.NewCompanionLearningSystem(10 * time.Second)
	clientLogger.WithField("system_name", "companion_learning").Debug("Created companion learning system")

	// INTEGRATION FIX [Category A]: Phase 21.2 - VehicleSystem high-level wrapper
	// Gap: VehicleSystem implemented but never initialized (subsystems VehicleMovementSystem, etc. initialized above)
	// Fix: Added high-level VehicleSystem for unified vehicle management
	// Roadmap: ROADMAP_V4.md Phase 21.2
	sys.vehicleSystem = engine.NewVehicleSystem(game.World)

	// AUDIT FIX: Phase 95-96 - Fishing and Gathering Systems
	// Gap: FishingSystem and GatheringSystem implemented but never initialized on client
	// Fix: Added system initialization for fishing and resource gathering minigames
	// Note: These systems must run on client for deterministic multiplayer sync (see fishing_multiplayer_sync_test.go)
	sys.fishingSystem = engine.NewFishingSystem(game.World, *seed+seedOffsetFishing)
	logging.ComponentLogger(clientLogger.Logger, "fishing").Debug("Created fishing system")
	sys.gatheringSystem = engine.NewGatheringSystem(game.World)
	logging.ComponentLogger(clientLogger.Logger, "gathering").Debug("Created gathering system")

	// FishingWeatherBonusSystem - connects weather with fishing for immersive gameplay
	// Rain increases rare fish bonus, storms attract legendary fish, etc.
	sys.fishingWeatherBonusSystem = engine.NewFishingWeatherBonusSystem(game.World, *seed+seedOffsetFishing+50)
	sys.fishingWeatherBonusSystem.SetGenre(*genreID)
	sys.fishingWeatherBonusSystem.SetFishingSystem(sys.fishingSystem)
	game.World.AddSystem(sys.fishingWeatherBonusSystem)
	logging.ComponentLogger(clientLogger.Logger, "fishing_weather").Debug("Created fishing weather bonus system")

	// TimeOfDayFishingBonusSystem - connects time of day with fishing for immersive gameplay
	// Dawn/dusk boost rare fish catches, night enhances legendary fish chance
	sys.timeOfDayFishingBonusSystem = engine.NewTimeOfDayFishingBonusSystem(game.World, *seed+seedOffsetFishing+100)
	sys.timeOfDayFishingBonusSystem.SetGenre(*genreID)
	sys.timeOfDayFishingBonusSystem.SetLightingSystem(sys.timeOfDayLightingSystem)
	sys.timeOfDayFishingBonusSystem.SetFishingSystem(sys.fishingSystem)
	game.World.AddSystem(sys.timeOfDayFishingBonusSystem)
	logging.ComponentLogger(clientLogger.Logger, "fishing_timeofday").Debug("Created time-of-day fishing bonus system")

	// FishingCatchParticleSystem - visual feedback for fish catches
	// Spawns genre-aware particles when fish are caught, scaled by rarity
	sys.fishingCatchParticleSystem = engine.NewFishingCatchParticleSystem(game.World, *seed+seedOffsetFishing+150)
	sys.fishingCatchParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.fishingCatchParticleSystem.SetFishingSystem(sys.fishingSystem)
	sys.fishingCatchParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.fishingCatchParticleSystem)
	logging.ComponentLogger(clientLogger.Logger, "fishing_particle").Debug("Created fishing catch particle system")

	// FishingLineTensionParticleSystem - visual feedback for line tension during reeling
	// Spawns genre-aware particles based on tension level and fish struggle direction
	sys.fishingLineTensionParticleSystem = engine.NewFishingLineTensionParticleSystem(game.World, *seed+seedOffsetFishing+160)
	sys.fishingLineTensionParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.fishingLineTensionParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.fishingLineTensionParticleSystem)
	logging.ComponentLogger(clientLogger.Logger, "fishing_tension").Debug("Created fishing line tension particle system")

	// TimeOfDayFishingBonusParticleSystem - visual feedback for time-based fishing bonuses
	// Spawns genre-aware particles on fishing spots when dawn/dusk/night bonuses are active
	sys.timeOfDayFishingBonusParticleSystem = engine.NewTimeOfDayFishingBonusParticleSystem(game.World, *seed+seedOffsetFishing+175)
	sys.timeOfDayFishingBonusParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.timeOfDayFishingBonusParticleSystem.SetTimeOfDayFishingBonusSystem(sys.timeOfDayFishingBonusSystem)
	sys.timeOfDayFishingBonusParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.timeOfDayFishingBonusParticleSystem)
	logging.ComponentLogger(clientLogger.Logger, "fishing_timeofday_particle").Debug("Created time-of-day fishing bonus particle system")

	// TerrainFishingBonusSystem - connects terrain tiles to fishing catch rate bonuses
	// Deep water adds +10-40% rare fish, trees (kelp) +15%, structures (ruins) +10%, bridges +20% speed
	sys.terrainFishingBonusSystem = engine.NewTerrainFishingBonusSystem(game.World, *seed+seedOffsetFishing+200)
	sys.terrainFishingBonusSystem.SetGenre(*genreID)
	sys.terrainFishingBonusSystem.SetTileSize(tileSize)
	sys.terrainFishingBonusSystem.SetFishingSystem(sys.fishingSystem)
	game.World.AddSystem(sys.terrainFishingBonusSystem)
	logging.ComponentLogger(clientLogger.Logger, "terrain_fishing").Debug("Created terrain fishing bonus system")

	// CompanionFishingBonusSystem - connects companion types to fishing catch rate bonuses
	// Water elementals +25% rare fish, spirits +20% detection, pets +15% speed, robots +30% legendary
	sys.companionFishingBonusSystem = engine.NewCompanionFishingBonusSystem(game.World, *seed+seedOffsetFishing+250)
	sys.companionFishingBonusSystem.SetGenre(*genreID)
	sys.companionFishingBonusSystem.SetFishingSystem(sys.fishingSystem)
	game.World.AddSystem(sys.companionFishingBonusSystem)
	logging.ComponentLogger(clientLogger.Logger, "companion_fishing").Debug("Created companion fishing bonus system")

	// AUDIT FIX: Phase 112 - CarryOverSystem for New Game Plus
	// Gap: CarryOverSystem implemented but never initialized on client
	// Fix: Added system initialization for NG+ carry-over selection UI and item transfer
	// Note: Client-only system for NG+ progression (works with prestige.System)
	sys.carryoverSystem = engine.NewCarryOverSystem(game.World)
	logging.ComponentLogger(clientLogger.Logger, "carryover").Debug("Created carryover system")

	if *verbose {
		clientLogger.Info("V4.0 systems initialized (vehicles, vehicle-mgmt, mounting, companions, companion-mgmt, skills, books, spells, classes, expressions, minigames, achievements, moral choices, discovery, investigation, NPC dialog, adaptive music, fishing, gathering, carryover)")
	}
}

// initializeV5Systems initializes Version 5.0 social and communication systems.
func initializeV5Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 3.1: Enhanced chat system with E2E encryption and history persistence
	sys.chatSystem = engine.NewEnhancedChatSystem(game.World)
	clientLogger.Info("enhanced chat system initialized with encryption and history persistence")

	// INTEGRATION FIX [Category A]: Phase 33 - TradeSystem
	// Gap: TradeSystem implemented but never initialized
	// Fix: Added system initialization for player-to-player trading with trust-based limits
	// Roadmap: ROADMAP_V5.md Phase 33
	sys.tradeSystem = engine.NewTradeSystem(game.World)

	// INTEGRATION FIX [Category A]: Phase 35 - Terrain Modification Systems
	// Gap: TerrainConstructionSystem and TerrainModificationSystem implemented but never initialized
	// Fix: Added system initialization for buildable walls and destructible terrain
	// Roadmap: ROADMAP_V5.md Phase 35
	sys.terrainConstructionSys = engine.NewTerrainConstructionSystemWithLogger(tileSize, clientLogger.Logger)
	sys.terrainModificationSys = engine.NewTerrainModificationSystemWithLogger(tileSize, clientLogger.Logger)

	// INTEGRATION FIX [Category A]: Phase 36 - MerchantCaravanSystem
	// Gap: MerchantCaravanSystem implemented but never initialized
	// Fix: Added system initialization for traveling merchants between servers
	// Roadmap: ROADMAP_V5.md Phase 36
	sys.merchantCaravanSystem = engine.NewMerchantCaravanSystem(game.World)

	// Phase 40: Mail system for asynchronous messaging
	sys.mailSystem = engine.NewMailSystem(game.World)

	// Phase 40: Courier system for mail delivery simulation (depends on MailSystem)
	sys.courierSystem = engine.NewCourierSystem(game.World, sys.mailSystem)

	// Phase 4.1: Network chat system for multiplayer messaging (PLAN.md)
	sys.networkChatSystem = chat.NewChatSystem(game.World)
	clientLogger.Info("network chat system initialized for multiplayer messaging")

	// Phase 4.2: Network trade system for multiplayer item trading (PLAN.md)
	sys.networkTradeSystem = trade.NewTradeSystem(game.World)
	clientLogger.Info("network trade system initialized for multiplayer trading")

	// Phase 5.1: Quality-of-Life System (PLAN.md)
	sys.qolManager = qol.NewManager(qol.Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	})
	sys.qolSystem = engine.NewQoLSystem(sys.qolManager)
	clientLogger.Info("QoL system initialized (auto-loot, auto-sort, quick-deposit)")

	if *verbose {
		clientLogger.Info("V5.0 systems initialized (chat, trade, terrain construction/modification, merchant caravans, mail, courier, network chat/trade, QoL)")
	}
}

// initializeV6Systems initializes Version 6.0 persistent world and federation systems.
func initializeV6Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 38: Federation protocol for server-to-server communication
	// Use deterministic client ID based on seed to ensure reproducible federation state
	serverID := fmt.Sprintf("client-%d", *seed)
	clientIdentity, err := federation.NewServerIdentity(serverID)
	if err != nil {
		clientLogger.WithError(err).Warn("Failed to create client identity")
		var fallbackErr error
		clientIdentity, fallbackErr = federation.NewServerIdentity("fallback-client")
		if fallbackErr != nil {
			clientLogger.WithError(fallbackErr).Error("Failed to create fallback client identity")
		}
	}
	sys.federationProtocol = federation.NewFederationProtocol(serverID, clientIdentity)

	// Phase 39: Portal system for cross-server travel
	sys.portalSystem = federation.NewPortalSystem(game.World, sys.federationProtocol)
	// AUDIT.md MEDIUM: AuthManager and TransferManager never instantiated.
	// Wire them into the portal system so cross-server player transfers can
	// create and validate session tokens on the client side.
	sys.portalSystem.SetManagers(federation.NewAuthManager(), federation.NewTransferManager())

	// Phase 40: Bounty system for cross-server quests
	sys.bountySystem = engine.NewBountySystem(game.World, game.World.GetLogger().Logger)

	// Phase 41: Politics system for server diplomacy
	sys.politicsSystem = engine.NewPoliticsSystem(game.World)

	// Phase 42: Territory control system
	sys.territoryManager = territory.NewManager()

	// Phase 42: Server ranking system
	sys.rankingManager = world.NewRankingManager()

	// Phase 42: Meta-game event system
	sys.eventManager = world.NewEventManager(*seed)

	if *verbose {
		clientLogger.Info("V6.0 systems initialized (federation, portals, bounties, politics, territories, rankings, events)")
	}
}

// initializeV7Systems initializes V7.0 Display & Viewport System.
// Gap: V7.0 features (display management, viewport optimization) implemented but never initialized
// Fix: Added initialization function for display manager and viewport optimizer
// Roadmap: ROADMAP_V7.md (Phase 43-44)
func initializeV7Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 43: Display Foundation - 1920x1080 default resolution with dynamic scaling
	displayConfig := &display.Config{
		Width:      *width,      // From command-line flag (default 1920)
		Height:     *height,     // From command-line flag (default 1080)
		Fullscreen: *fullscreen, // From command-line flag
		VSync:      true,        // VSync enabled by default
	}
	sys.displayManager = display.NewManager(displayConfig)

	// Apply resolution on initialization
	if switchDuration := sys.displayManager.ApplyResolution(); *verbose {
		clientLogger.WithFields(logrus.Fields{
			"width":          displayConfig.Width,
			"height":         displayConfig.Height,
			"fullscreen":     displayConfig.Fullscreen,
			"switchDuration": switchDuration,
		}).Info("display manager initialized and resolution applied")
	}

	// Phase 44: Viewport Optimization - Enhanced culling for larger resolutions
	sys.viewportOptimizer = engine.NewViewportOptimizer()
	sys.viewportOptimizer.SetTileSize(32.0) // Standard tile size
	sys.viewportOptimizer.SetMarginTiles(1) // 1-tile margin for smooth scrolling

	// Wire up F11 fullscreen toggle to display manager
	if sys.inputSystem != nil && sys.displayManager != nil {
		dm := sys.displayManager
		sys.inputSystem.SetFullscreenToggleCallback(func() error {
			dm.ToggleFullscreen()
			clientLogger.WithFields(logrus.Fields{
				"fullscreen": dm.GetConfig().Fullscreen,
			}).Info("fullscreen toggled (F11)")
			return nil
		})
	}

	if *verbose {
		clientLogger.Info("V7.0 systems initialized (display manager, viewport optimizer)")
	}
}

// initializeV8Systems initializes Version 8.0 housing, physics, and social persistence systems.
// Gap: V8.0 systems fully implemented but never initialized in client
// Fix: Added complete V8.0 system initialization (housing, trust, physics, building/furniture)
// Roadmap: ROADMAP_V8.md (Phase 49-51)
func initializeV8Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 49.1: Housing Core Infrastructure
	sys.housingManager = housing.NewManager()

	// Phase 49.2: Persistent Trust & Reputation System
	sys.trustManager = persistence.NewTrustManager()
	sys.reputationManager = persistence.NewReputationManager()

	// Phase 49.3: Chat History with Delta Compression
	// Note: ChatHistory/ImageGallery are per-player instances. They are initialized
	// in initializeHousingAndGalleryUI after the player entity is created with the
	// actual player ID. We leave them nil here - they will be created later.
	// sys.chatHistory and sys.imageGallery initialized in initializeHousingAndGalleryUI

	// Phase 50.3: Enhanced Vehicle Physics
	sys.enhancedVehicleSys = vehicle.NewEnhancedVehicleSystem()
	game.World.AddSystem(&enhancedVehicleSystemWrapper{system: sys.enhancedVehicleSys}) // AUDIT.md REM-019: Register vehicle physics system

	// Phase 50.4: Fluid Dynamics & Swimming
	fluidConfig := fluids.SimulationConfig{
		GridWidth:       100,
		GridHeight:      100,
		CellSize:        1.0,
		UpdateRate:      30.0,
		Gravity:         9.8,
		PressureFactor:  0.8,
		ViscosityFactor: 0.01,
		MaxIterations:   10,
		Convergence:     0.001,
	}
	sys.fluidSimulator = fluids.NewSimulator(fluidConfig)
	sys.buoyancyCalculator = fluids.NewBuoyancyCalculator(fluidConfig.Gravity)
	sys.swimmingManager = fluids.NewSwimmingManager(fluidConfig.Gravity)
	sys.floodingManager = fluids.NewFloodingManager(sys.fluidSimulator)

	// Phase 51.1: Procedural Building Generation
	sys.buildingGenerator = building.NewGenerator()

	// Phase 51.2: Guild Hall Construction
	sys.guildHallManager = housing.NewGuildHallManager()

	// Phase 51.3: Furniture Generation & Placement
	sys.furnitureGenerator = furniture.NewGenerator()

	if *verbose {
		clientLogger.Info("V8.0 systems initialized (housing, trust, reputation, chat history, images, vehicle physics, fluid dynamics, buildings, guild halls, furniture)")
	}
}

// initializeV9Systems initializes V9.0 Integration Manager.
// Gap: V9.0 integration features (housing crafting, companion housing, guild housing) implemented but never initialized
// Fix: Added initialization function for V9.0 integration managers
// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
func initializeV9Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 55.1: Crafting Stations in Player Housing
	// Enables placing forges, alchemy tables, enchanting stations in player homes
	// Provides skill training facilities and recipe unlocking system
	sys.stationManager = housingcrafting.NewStationManager()

	// Phase 55.2: Companion Housing & Pet Homes
	// Allows companions to live in player housing with bedding quality affecting loyalty
	// Training areas provide XP bonuses, shared storage accessible by companions
	sys.petHomeManager = companionhousing.NewPetHomeManager()

	// Phase 55.3: Guild Housing & Communal Spaces
	// Guild halls with rank-based access permissions
	// Communal crafting stations, guild storage, meeting halls with chat bonuses
	sys.guildHousingManager = guildhousing.NewManager()

	if *verbose {
		clientLogger.Info("V9.0 integration managers initialized (crafting stations, companion housing, guild housing)")
	}
}

// initializeV19Systems initializes V19.0 Priority 1 dormant packages (ROADMAP_V19.md)
func initializeV19Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 99: Entity & Dialog Generation
	// Procedural entity generator for monsters and NPCs
	sys.entityGenerator = entity.NewEntityGeneratorWithLogger(clientLogger.Logger)
	clientLogger.Debug("entity generator initialized")

	// Markov-chain dialog generator for NPC conversations
	// Uses genre-specific corpus for contextual dialogs
	sys.dialogGenerator = dialog.NewMarkovGenerator(*seed, *genreID, dialog.Order2)
	// Train with genre-specific corpus
	if corpus := dialog.GetCorpus(*genreID); corpus != nil {
		sys.dialogGenerator.TrainFromCorpus(corpus.Sentences)
		clientLogger.WithFields(logrus.Fields{
			"genreID":   *genreID,
			"chainSize": sys.dialogGenerator.GetChainSize(),
		}).Debug("dialog generator initialized and trained")
	} else {
		// Fallback to fantasy corpus
		if fantasyCorpus := dialog.GetCorpus("fantasy"); fantasyCorpus != nil {
			sys.dialogGenerator.TrainFromCorpus(fantasyCorpus.Sentences)
		}
		clientLogger.Debug("dialog generator initialized with fallback corpus")
	}

	// Phase 100: Legendary & Economy
	// Legendary quest manager for boss drops and quest rewards
	sys.legendaryManager = legendary.NewQuestManager(nil) // nil raid manager for now
	clientLogger.Debug("legendary quest manager initialized")

	// Federated marketplace and dynamic economy system
	serverID := fmt.Sprintf("client-%d", *seed)
	sys.worldEconomySystem = economy.NewSystemWithServerID(nil, serverID) // nil world for now, entity interface
	clientLogger.Debug("world economy system initialized")

	// Phase 101: Integration Package Activation
	// World event manager for responsive world events
	sys.worldEventManager = world_events.NewEventManager(*seed + seedOffsetWorldEvents)
	clientLogger.Debug("world event manager initialized")

	if *verbose {
		clientLogger.Info("V19.0 systems initialized (entity gen, dialog gen, legendary, economy, choice tracker, fleet manager, world events)")
	}
}

// initializeVRSystems initializes VR systems with hardware detection.
// AUDIT.md Task 7: VR hardware detection for conditional stereoscopic and head tracking initialization.
func initializeVRSystems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Check if VR is enabled via flags
	if !*enableVR && !*forceVR {
		clientLogger.Debug("VR disabled via flags, skipping VR system initialization")
		return
	}

	// Perform hardware detection
	detector := vr.NewDetector()
	if *forceVR {
		detector.SetForceEnable(true)
		clientLogger.Info("VR mode force-enabled via --force-vr flag")
	}

	vrAvailable := detector.DetectHardware()

	if !vrAvailable && !*forceVR {
		clientLogger.Info("VR hardware not detected, skipping VR system initialization")
		clientLogger.Info("Use --force-vr to enable VR mode without hardware (testing)")
		return
	}

	clientLogger.WithFields(logrus.Fields{
		"headset":    detector.IsHeadsetDetected(),
		"controller": detector.IsControllerDetected(),
		"force":      *forceVR,
	}).Info("VR mode enabled, initializing VR systems")

	// Initialize stereoscopic rendering system
	sys.stereoscopicSystem = engine.NewStereoscopicSystem(game.World)
	game.World.AddSystem(sys.stereoscopicSystem)
	clientLogger.Debug("stereoscopic rendering system initialized")

	// Initialize head tracking system
	sys.headTrackingSystem = engine.NewHeadTrackingSystem(game.World)

	// If no headset detected, enable mouse fallback
	if !detector.IsHeadsetDetected() {
		sys.headTrackingSystem.SetUseMouseFallback(true)
		clientLogger.Debug("head tracking system initialized with mouse fallback (no headset)")
	} else {
		headsetAdapter := engine.NewRuntimeHeadsetAdapter()
		sys.headTrackingSystem.SetHeadsetAdapter(headsetAdapter)
		clientLogger.Debug("head tracking system initialized with runtime headset adapter")
	}

	game.World.AddSystem(sys.headTrackingSystem)

	// Initialize VR controller system if controllers detected
	if detector.IsControllerDetected() || *forceVR {
		sys.vrControllerSystem = engine.NewVRControllerSystem(game.World)

		controllerAdapter := engine.NewRuntimeControllerAdapter()
		sys.vrControllerSystem.SetControllerAdapter(controllerAdapter)

		game.World.AddSystem(sys.vrControllerSystem)
		clientLogger.Debug("VR controller system initialized with runtime controller adapter")
	}

	// Initialize VR UI system for VR-optimized rendering
	sys.vrUISystem = engine.NewVRUISystem(game.World)
	game.World.AddSystem(sys.vrUISystem)
	clientLogger.Debug("VR UI system initialized")

	if *verbose {
		clientLogger.Info("VR systems initialized successfully")
	}
}

// initializePhase3Systems initializes Phase 3 systems from PLAN.md (Networking & Social Features)
func initializePhase3Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 3.2: Guild Federation - Cross-server guild management
	guildManager := guild.NewManager()
	sys.guildSystem = engine.NewGuildSystem(game.World, guildManager)
	sys.guildUI = engine.NewGuildUI(game.World, sys.guildSystem, *width, *height)

	// Phase 3.2b: Guild Combat Bonus System - proximity-based combat bonuses for guild members
	sys.guildCombatBonusSystem = engine.NewGuildCombatBonusSystem(game.World, *seed+5152)
	sys.guildCombatBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.guildCombatBonusSystem)
	logging.ComponentLogger(clientLogger.Logger, "guild_combat_bonus").Debug("Created guild combat bonus system")

	// Connect guild system to federation protocol for cross-server sync
	if sys.federationProtocol != nil {
		sys.guildSystem.SetFederation(sys.federationProtocol)
		guildManager.SetTransport(sys.federationProtocol)
		clientLogger.Debug("guild system connected to federation protocol")
	}

	// Phase 4.4: Political Warfare Integration
	sys.politicalWarfareSystem = political_warfare.NewSystem(game.World, guildManager, *seed)
	logging.ComponentLogger(clientLogger.Logger, "political_warfare").Debug("Created political warfare system")

	// Phase 4.5: Territory Siege System
	sys.siegeManager = territory.NewSiegeManager()
	sys.siegeSystem = engine.NewTerritorySiegeSystem(sys.siegeManager)
	logging.ComponentLogger(clientLogger.Logger, "territory_siege").Debug("Created territory siege system")

	// Phase 4.4: Trade Routes System (PLAN.md Integration - Week 4)
	// Generate deterministic server ID from seed for trade route identification
	tradeServerID := fmt.Sprintf("client-%d", *seed)
	sys.tradeRouteManager = trade_routes.NewRouteManager(tradeServerID, *seed+seedOffsetTradeRoutes)
	sys.tradeRouteManager.Start()
	logging.ComponentLogger(clientLogger.Logger, "trade_routes").Debug("Created trade routes system")

	// Phase 5.1: Mobile Federation Support
	mobileConfig := mobilefed.DefaultConfig()
	sys.mobileFederationSystem = engine.NewMobileFederationSystem(mobileConfig)
	logging.ComponentLogger(clientLogger.Logger, "mobile_federation").Debug("Created mobile federation system")

	// Phase 4.3: Territory Control - Guild warfare and territory management
	sys.territorySystem = engine.NewTerritorySystem(sys.territoryManager, clientLogger.WithField("system", "territory"))
	sys.territoryUI = engine.NewTerritoryUI(sys.territorySystem, *width, *height)

	if *verbose {
		clientLogger.Info("Phase 3 systems initialized (guild federation with cross-server sync, territory control)")
	}
}

// initializeModBrowserWiring wires the ModBrowserSystem (already registered in
// the World by system_init.go / scheduleLazyInit) to a modding.Manager so that
// install/uninstall actions bridge into the sandboxed JSON mod loader.
//
// It also implements AUDIT.md G5: loads mods from disk and calls
// world.SetModRules so rule overrides in mods/*.json take effect in
// single-player and host-and-play mode.
//
// Satisfies AUDIT.md G3 and G5.
func initializeModBrowserWiring(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	log := logging.ComponentLogger(clientLogger.Logger, "mod_browser")

	// Start from DefaultConfig so ModsDirectory defaults to "mods" and all
	// other defaults are correct; override only the fields we need to customise.
	modCfg := modding.DefaultConfig()
	modCfg.EnableSandbox = true
	modCfg.MaxMods = 50
	modCfg.RuleChangeRateLimit = 10.0
	loader := modding.NewLoaderWithConfig(modCfg)
	sys.modManager = modding.NewManagerWithConfig(modCfg)

	// G5: load mods from disk and apply rule overrides to the world.
	// Only done when the client owns the world (single-player / host-and-play).
	// In pure multiplayer mode the dedicated server already manages mod rules.
	if !*multiplayer {
		mods, err := loader.LoadAll()
		if err != nil {
			log.WithError(err).Warn("mod loading: some mods failed to load — continuing without them")
		}
		loadedCount := 0
		for _, mod := range mods {
			if err := sys.modManager.AddMod(mod); err != nil {
				log.WithFields(logrus.Fields{
					"mod_id": mod.ID,
					"error":  err.Error(),
				}).Warn("mod_browser: failed to add mod to manager")
				continue
			}
			if err := sys.modManager.EnableMod(mod.ID); err != nil {
				log.WithFields(logrus.Fields{
					"mod_id": mod.ID,
					"error":  err.Error(),
				}).Warn("mod_browser: failed to enable mod")
				continue
			}
			loadedCount++
		}
		log.WithField("count", loadedCount).Info("mods loaded from disk")
		game.World.SetModRules(modding.NewProviderAdapter(sys.modManager))
		log.Debug("mod rules applied to world (single-player / host-and-play)")
	}

	// Reuse the ModBrowserSystem already registered by InitializeGameSystems (if
	// called, e.g. in cmd/mobile). If none is present yet, create and register one
	// now. This prevents a duplicate system being added to the World.
	for _, s := range game.World.GetSystems() {
		if mbs, ok := s.(*engine.ModBrowserSystem); ok {
			sys.modBrowserSys = mbs
			break
		}
	}
	if sys.modBrowserSys == nil {
		sys.modBrowserSys = engine.NewModBrowserSystem(game.World)
		game.World.AddSystem(sys.modBrowserSys)
	}

	// G16 (AUDIT.md): Use the filesystem-backed mod repository so the mod browser
	// shows mods already present in the mods/ directory on disk, rather than an
	// empty in-memory store that was only suitable for testing.
	sys.modBrowserSys.SetRepository(engine.NewFileSystemModRepository(modCfg.ModsDirectory))

	// Install callback: parse raw JSON bytes into a Mod and add it to the manager.
	sys.modBrowserSys.SetInstallCallback(func(modID string, modData []byte) error {
		mod, err := modding.ParseModFromBytes(modData)
		if err != nil {
			return fmt.Errorf("mod_browser: parse %s: %w", modID, err)
		}
		if err := sys.modManager.AddMod(mod); err != nil {
			return fmt.Errorf("mod_browser: install %s: %w", modID, err)
		}
		log.WithField("mod_id", modID).Info("mod installed via ModBrowserSystem")
		return nil
	})

	// Uninstall callback: delegate removal to the manager.
	sys.modBrowserSys.SetUninstallCallback(func(modID string) error {
		if err := sys.modManager.RemoveMod(modID); err != nil {
			return fmt.Errorf("mod_browser: uninstall %s: %w", modID, err)
		}
		log.WithField("mod_id", modID).Info("mod uninstalled via ModBrowserSystem")
		return nil
	})

	log.Debug("ModBrowserSystem wired with modding.Manager callbacks")

	// G15 (AUDIT.md): Register HotReloadSystem for live mod reloading.
	// The system monitors the mods directory for changes and applies updates
	// without requiring a game restart, fulfilling the Modding System goal.
	watcher := engine.NewFileSystemFileWatcher(modCfg.ModsDirectory)
	hotReload := engine.NewHotReloadSystem(game.World)
	hotReload.SetFileWatcher(watcher)

	// Hash callback delegates to the watcher, which auto-detects file changes
	// via modtime comparison — no manual cache invalidation needed.
	hotReload.SetHashCallback(func(modID string) (string, error) {
		return watcher.GetFileHash(modID)
	})

	// Reload callback: parse fresh mod data; use the JSON's own ID as the
	// authoritative key to avoid mismatches between filename and "id" field.
	hotReload.SetReloadCallback(func(modID string, modData []byte) error {
		mod, err := modding.ParseModFromBytes(modData)
		if err != nil {
			return fmt.Errorf("hot_reload: parse %s: %w", modID, err)
		}
		// Normalise: the JSON "id" field is the authoritative mod identifier used
		// by Manager.  The watcher-provided modID is derived from the filename
		// (stem of the .json file).  If they differ we use the JSON id and log a
		// warning so operators can fix the naming inconsistency.  All downstream
		// calls (Remove/Add/Enable) use the same canonicalID for consistency.
		if mod.ID != modID {
			log.WithFields(logrus.Fields{
				"filename_id": modID,
				"json_id":     mod.ID,
			}).Warn("hot_reload: mod JSON id differs from filename — using JSON id")
		}
		canonicalID := mod.ID
		_ = sys.modManager.RemoveMod(canonicalID)
		if err := sys.modManager.AddMod(mod); err != nil {
			return fmt.Errorf("hot_reload: add %s: %w", canonicalID, err)
		}
		if err := sys.modManager.EnableMod(canonicalID); err != nil {
			return fmt.Errorf("hot_reload: enable %s: %w", canonicalID, err)
		}
		log.WithField("mod_id", canonicalID).Info("mod reloaded via HotReloadSystem")
		return nil
	})
	hotReload.SetRollbackCallback(func(modID string, state *engine.ModState) error {
		log.WithFields(logrus.Fields{
			"mod_id":  modID,
			"version": state.Version,
		}).Warn("hot_reload: rollback requested; state migration not configured in this build")
		return fmt.Errorf("hot_reload: rollback not configured for mod %s", modID)
	})
	game.World.AddSystem(hotReload)

	// Seed a world-level entity with HotReloadComponent and start watching all
	// currently enabled mods so the system has something to monitor immediately.
	hotReloadEntity := game.World.CreateEntity()
	hotReloadComp := engine.NewHotReloadComponent()
	hotReloadEntity.AddComponent(hotReloadComp)
	for _, mod := range sys.modManager.ListMods() {
		if !mod.Enabled {
			continue
		}
		if err := hotReload.StartWatchingMod(hotReloadComp, mod.ID); err != nil {
			log.WithFields(logrus.Fields{
				"mod_id": mod.ID,
				"error":  err.Error(),
			}).Warn("hot_reload: failed to start watching mod")
		}
	}
	log.WithField("count", len(hotReloadComp.GetWatchedModIDs())).Debug("HotReloadSystem registered for live mod reloading")
}
