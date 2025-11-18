//go:build !android && !ios
// +build !android,!ios

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/sirupsen/logrus"
)

// systemsContainer holds all initialized game systems for dependency injection.
type systemsContainer struct {
	inputSystem            *engine.InputSystem
	movementSystem         *engine.MovementSystem
	collisionSystem        *engine.CollisionSystem
	combatSystem           *engine.CombatSystem
	interactionSystem      *engine.InteractionSystem
	particleSystem         *engine.ParticleSystem
	animationSystem        *engine.AnimationSystem
	equipmentVisualSystem  *engine.EquipmentVisualSystem
	objectiveTracker       *engine.ObjectiveTrackerSystem
	aiSystem               *engine.AISystem
	progressionSystem      *engine.ProgressionSystem
	inventorySystem        *engine.InventorySystem
	commerceSystem         *engine.CommerceSystem
	dialogSystem           *engine.DialogSystem
	craftingSystem         *engine.CraftingSystem
	audioManager           *engine.AudioManager
	audioManagerSystem     *engine.AudioManagerSystem
	itemPickupSystem       *engine.ItemPickupSystem
	statusEffectSystem     *engine.StatusEffectSystem
	spellCastingSystem     *engine.SpellCastingSystem
	playerSpellCasting     *engine.PlayerSpellCastingSystem
	manaRegenSystem        *engine.ManaRegenSystem
	playerCombatSystem     *engine.PlayerCombatSystem
	playerItemUseSystem    *engine.PlayerItemUseSystem
	rotationSystem         *engine.RotationSystem
	projectileSystem       *engine.ProjectileSystem
	revivalSystem          *engine.RevivalSystem
	behaviorTreeSystem     *engine.BehaviorTreeSystem
	squadSystem            *engine.SquadSystem
	factionSystem          *engine.FactionSystem
	reputationSystem       *engine.ReputationSystem
	alignmentSystem        *engine.AlignmentSystem
	factionReactionSystem  *engine.FactionReactionSystem
	skillProgressionSystem *engine.SkillProgressionSystem
	visualFeedbackSystem   *engine.VisualFeedbackSystem
	weatherSystem          *engine.WeatherSystem
	lifetimeSystem         *engine.LifetimeSystem
	puzzleSystem           *engine.PuzzleSystem
	firePropagationSystem  *engine.FirePropagationSystem
	destructibleSystem     *engine.DestructibleObjectSystem
	carrySystem            *engine.CarrySystem
	hazardSystem           *engine.HazardSystem
	narrativeSystem        *engine.NarrativeSystem
	shadowSystem           *engine.ShadowSystem
	spriteGenerator        *sprites.Generator
	itemGen                *item.ItemGenerator
	recipeGen              *recipe.RecipeGenerator
	statusEffectRNG        *rand.Rand
	// V4.0 Systems (Phase 21-27)
	vehicleMovementSys      *engine.VehicleMovementSystem
	vehicleDurabilitySys    *engine.VehicleDurabilitySystem
	mountingSystem          *engine.MountingSystem
	vehicleCombatSystem     *engine.VehicleCombatSystem
	companionAISystem       *engine.CompanionAISystem
	companionProgressionSys *engine.CompanionProgressionSystem
	companionLoyaltySys     *engine.CompanionLoyaltySystem
	companionInventorySys   *engine.CompanionInventorySystem
	skillInheritanceSys     *engine.SkillInheritanceSystem
	bookReadingSystem       *engine.BookReadingSystem
	spellEffectSystem       *engine.SpellEffectSystem
	spellCombinationSys     *engine.SpellCombinationSystem
	classProgressionSys     *engine.ClassProgressionSystem
	expressionSystem        *engine.ExpressionSystem
	expressionComboSys      *engine.ExpressionComboSystem
	miniGameSystem          *engine.MiniGameSystem
	achievementSystem       *engine.AchievementSystem
	// Phase 28: Reputation & Moral Choices
	moralChoiceSystem *engine.MoralChoiceSystem
	// Phase 30: Environmental Storytelling
	discoverySystem *engine.DiscoverySystem
	// V5.0 Systems (Social & Communication)
	chatSystem    *engine.ChatSystem
	mailSystem    *engine.MailSystem
	courierSystem *engine.CourierSystem
	// V6.0 Systems (Persistent Worlds & Federation)
	portalSystem       *federation.PortalSystem
	bountySystem       *engine.BountySystem
	politicsSystem     *engine.PoliticsSystem
	territoryManager   *world.TerritoryManager
	rankingManager     *world.RankingManager
	eventManager       *world.EventManager
	federationProtocol *federation.FederationProtocol
}

// initializeCoreSystems creates and initializes all core game systems.
func initializeCoreSystems(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *systemsContainer {
	clientLogger.Info("initializing game systems")

	sys := &systemsContainer{}

	// Core gameplay systems
	sys.inputSystem = engine.NewInputSystem()

	sys.movementSystem = engine.NewMovementSystem(playerMaxSpeed)
	sys.collisionSystem = engine.NewCollisionSystem(collisionGridCellSize)
	sys.movementSystem.SetCollisionSystem(sys.collisionSystem)

	sys.combatSystem = engine.NewCombatSystemWithLogger(*seed, logger)
	sys.interactionSystem = engine.NewInteractionSystem(game.World)
	sys.particleSystem = engine.NewParticleSystem()

	sys.spriteGenerator = sprites.NewGenerator()
	sys.animationSystem = engine.NewAnimationSystem(sys.spriteGenerator)
	sys.animationSystem.SetMaxCacheSize(animationCacheSize)

	sys.equipmentVisualSystem = engine.NewEquipmentVisualSystem(sys.spriteGenerator)

	return sys
}

// initializeGenerators creates item and recipe generators for loot drops.
func initializeGenerators(sys *systemsContainer) {
	sys.itemGen = item.NewItemGenerator()
	sys.recipeGen = recipe.NewRecipeGenerator()
}

// initializeObjectiveSystem creates and configures the objective tracker with callbacks.
func initializeObjectiveSystem(sys *systemsContainer, logger *logrus.Logger) {
	sys.objectiveTracker = engine.NewObjectiveTrackerSystem()

	sys.objectiveTracker.SetQuestCompleteCallback(func(entity *engine.Entity, qst *quest.Quest) {
		sys.objectiveTracker.AwardQuestRewards(entity, qst)
		if logger.GetLevel() >= logrus.InfoLevel {
			logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
				"questName":   qst.Name,
				"xpReward":    qst.Reward.XP,
				"goldReward":  qst.Reward.Gold,
				"skillPoints": qst.Reward.SkillPoints,
			}).Info("quest completed")
		}
	})
}

// initializeProgressionSystems creates AI, progression, inventory, and commerce systems.
func initializeProgressionSystems(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Logger) {
	sys.aiSystem = engine.NewAISystem(game.World)
	sys.progressionSystem = engine.NewProgressionSystem(game.World)
	sys.inventorySystem = engine.NewInventorySystem(game.World)
	sys.commerceSystem = engine.NewCommerceSystemWithLogger(game.World, sys.inventorySystem, logger)
	sys.dialogSystem = engine.NewDialogSystemWithLogger(game.World, logger)
	sys.craftingSystem = engine.NewCraftingSystem(game.World, sys.inventorySystem, sys.itemGen)

	logging.ComponentLogger(logger, "commerce").Info("commerce system initialized")
	logging.ComponentLogger(logger, "dialog").Info("dialog system initialized")
	logging.ComponentLogger(logger, "crafting").Info("crafting system initialized")
}

// initializeAudioSystem creates and configures the audio manager with music playback.
func initializeAudioSystem(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	sys.audioManager = engine.NewAudioManager(audioSampleRate, *seed)
	sys.audioManagerSystem = engine.NewAudioManagerSystem(sys.audioManager)

	game.SetAudioManager(sys.audioManager)
	sys.audioManager.EnableAdaptiveMusic(true)

	if *verbose {
		clientLogger.Info("adaptive music composition enabled with motif system")
	}

	if err := sys.audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(clientLogger.Logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(clientLogger.Logger, "audio").Info("audio system initialized (music and SFX generators)")
}

// initializeCombatSystems creates spell casting and player combat systems.
func initializeCombatSystems(game *engine.EbitenGame, sys *systemsContainer) {
	sys.itemPickupSystem = engine.NewItemPickupSystem(game.World)

	sys.statusEffectRNG = rand.New(rand.NewSource(*seed + seedOffsetStatusEffect))
	sys.statusEffectSystem = engine.NewStatusEffectSystem(game.World, sys.statusEffectRNG)
	sys.spellCastingSystem = engine.NewSpellCastingSystem(game.World, sys.statusEffectSystem)
	sys.playerSpellCasting = engine.NewPlayerSpellCastingSystem(sys.spellCastingSystem, game.World)
	sys.manaRegenSystem = &engine.ManaRegenSystem{}
	sys.playerCombatSystem = engine.NewPlayerCombatSystem(sys.combatSystem, game.World)
	sys.playerItemUseSystem = engine.NewPlayerItemUseSystem(sys.inventorySystem, game.World)
}

// initializeEnvironmentalSystems creates weather, lifetime, puzzle, and destruction systems.
func initializeEnvironmentalSystems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	sys.weatherSystem = engine.NewWeatherSystem(game.World)
	sys.lifetimeSystem = engine.NewLifetimeSystemWithLogger(game.World, clientLogger.Logger)
	sys.puzzleSystem = engine.NewPuzzleSystem(game.World)

	sys.firePropagationSystem = engine.NewFirePropagationSystemWithLogger(tileSize, *seed+seedOffsetFirePropagation, clientLogger.Logger)
	sys.firePropagationSystem.SetWorld(game.World)

	sys.destructibleSystem = engine.NewDestructibleObjectSystemWithLogger(tileSize, *seed+seedOffsetDestructible, clientLogger.Logger)
	sys.destructibleSystem.SetWorld(game.World)
	sys.destructibleSystem.SetFireSystem(sys.firePropagationSystem)

	sys.carrySystem = engine.NewCarrySystemWithLogger(clientLogger.Logger)
	sys.carrySystem.SetWorld(game.World)

	sys.hazardSystem = engine.NewHazardSystemWithLogger(clientLogger.Logger)
	sys.hazardSystem.SetWorld(game.World)

	sys.narrativeSystem = engine.NewNarrativeSystem(game.World)
	sys.shadowSystem = engine.NewShadowSystemWithLogger(game.World, clientLogger.Logger)
}

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
	sys.companionLoyaltySys = engine.NewCompanionLoyaltySystem(game.World, clientLogger.Logger)
	sys.companionInventorySys = engine.NewCompanionInventorySystem(game.World)
	sys.skillInheritanceSys = engine.NewSkillInheritanceSystem(game.World)

	// Phase 23: Book system
	sys.bookReadingSystem = engine.NewBookReadingSystem(game.World)

	// Phase 24: Expanded magic systems (reuse statusEffectRNG for consistency)
	spellRNG := rand.New(rand.NewSource(*seed + seedOffsetSpellEffects))
	sys.spellEffectSystem = engine.NewSpellEffectSystem(game.World, spellRNG)
	sys.spellCombinationSys = engine.NewSpellCombinationSystem(game.World, spellRNG)

	// Phase 25: Class progression system
	sys.classProgressionSys = engine.NewClassProgressionSystem()

	// Phase 26: Expression systems (requires audio manager)
	sys.expressionSystem = engine.NewExpressionSystem(game.World, sys.audioManager)
	sys.expressionComboSys = engine.NewExpressionComboSystem(game.World)

	// Phase 27: Mini-game system
	sys.miniGameSystem = engine.NewMiniGameSystem(game.World)

	// Phase 26.2: Achievement system (social features)
	sys.achievementSystem = engine.NewAchievementSystem(game.World)

	// INTEGRATION FIX [Category A]: Phase 28 - MoralChoiceSystem
	// Gap: MoralChoiceSystem implemented but never initialized or registered
	// Fix: Added system initialization for moral decision tracking and consequences
	// Roadmap: ROADMAP_V4.md Phase 28
	sys.moralChoiceSystem = engine.NewMoralChoiceSystem(game.World, clientLogger.Logger)

	// Phase 30: Environmental Storytelling - Discovery System
	sys.discoverySystem = engine.NewDiscoverySystem(game.World)

	if *verbose {
		clientLogger.Info("V4.0 systems initialized (vehicles, companions, books, magic, classes, expressions, mini-games, achievements, moral choices, story discovery)")
	}
}

// initializeV5Systems initializes Version 5.0 social and communication systems.
func initializeV5Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 32: Chat system for player-to-player communication
	sys.chatSystem = engine.NewChatSystem(game.World)

	// Phase 40: Mail system for asynchronous messaging
	sys.mailSystem = engine.NewMailSystem(game.World)

	// Phase 40: Courier system for mail delivery simulation (depends on MailSystem)
	sys.courierSystem = engine.NewCourierSystem(game.World, sys.mailSystem)

	if *verbose {
		clientLogger.Info("V5.0 systems initialized (chat, mail, courier)")
	}
}

// initializeV6Systems initializes Version 6.0 persistent world and federation systems.
func initializeV6Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 38: Federation protocol for server-to-server communication
	serverID := fmt.Sprintf("client-%d", time.Now().Unix())
	clientIdentity, err := federation.NewServerIdentity(serverID)
	if err != nil {
		clientLogger.WithError(err).Warn("Failed to create client identity")
		clientIdentity, _ = federation.NewServerIdentity("fallback-client")
	}
	sys.federationProtocol = federation.NewFederationProtocol(serverID, clientIdentity)

	// Phase 39: Portal system for cross-server travel
	sys.portalSystem = federation.NewPortalSystem(game.World, sys.federationProtocol)

	// Phase 40: Bounty system for cross-server quests
	sys.bountySystem = engine.NewBountySystem(game.World, game.World.GetLogger().Logger)

	// Phase 41: Politics system for server diplomacy
	sys.politicsSystem = engine.NewPoliticsSystem(game.World)

	// Phase 42: Territory control system
	sys.territoryManager = world.NewTerritoryManager()

	// Phase 42: Server ranking system
	sys.rankingManager = world.NewRankingManager()

	// Phase 42: Meta-game event system
	sys.eventManager = world.NewEventManager(*seed)

	if *verbose {
		clientLogger.Info("V6.0 systems initialized (federation, portals, bounties, politics, territories, rankings, events)")
	}
}

// registerAllSystems adds all systems to the game world in the correct order.
func registerAllSystems(game *engine.EbitenGame, sys *systemsContainer) {
	game.World.AddSystem(sys.inputSystem)
	game.World.AddSystem(game.CameraSystem)

	sys.rotationSystem = engine.NewRotationSystem(game.World)
	game.World.AddSystem(&rotationSystemWrapper{system: sys.rotationSystem})

	game.World.AddSystem(sys.playerCombatSystem)
	game.World.AddSystem(sys.playerItemUseSystem)
	game.World.AddSystem(sys.playerSpellCasting)
	game.World.AddSystem(sys.movementSystem)
	game.World.AddSystem(sys.collisionSystem)

	sys.projectileSystem = engine.NewProjectileSystem(game.World)
	game.World.AddSystem(sys.projectileSystem)

	game.World.AddSystem(sys.combatSystem)
	game.World.AddSystem(sys.statusEffectSystem)

	sys.revivalSystem = engine.NewRevivalSystem(game.World)
	game.World.AddSystem(sys.revivalSystem)

	game.World.AddSystem(sys.aiSystem)

	sys.behaviorTreeSystem = engine.NewBehaviorTreeSystem(game.World)
	game.World.AddSystem(sys.behaviorTreeSystem)

	sys.squadSystem = engine.NewSquadSystem(game.World)
	game.World.AddSystem(&squadSystemWrapper{system: sys.squadSystem})

	game.World.AddSystem(sys.progressionSystem)

	sys.factionSystem = engine.NewFactionSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(sys.factionSystem)

	sys.reputationSystem = engine.NewReputationSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(&reputationSystemWrapper{system: sys.reputationSystem})

	sys.alignmentSystem = engine.NewAlignmentSystem(game.World)
	game.World.AddSystem(&alignmentSystemWrapper{system: sys.alignmentSystem})

	sys.factionReactionSystem = engine.NewFactionReactionSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(&factionReactionSystemWrapper{system: sys.factionReactionSystem})

	sys.skillProgressionSystem = engine.NewSkillProgressionSystem()
	game.World.AddSystem(sys.skillProgressionSystem)

	sys.visualFeedbackSystem = engine.NewVisualFeedbackSystem()
	game.World.AddSystem(sys.visualFeedbackSystem)

	game.World.AddSystem(sys.audioManagerSystem)
	game.World.AddSystem(sys.objectiveTracker)
	game.World.AddSystem(sys.itemPickupSystem)
	game.World.AddSystem(sys.spellCastingSystem)
	game.World.AddSystem(sys.manaRegenSystem)
	game.World.AddSystem(sys.inventorySystem)
	game.World.AddSystem(sys.commerceSystem)
	game.World.AddSystem(sys.dialogSystem)
	game.World.AddSystem(sys.craftingSystem)
	game.World.AddSystem(sys.interactionSystem)

	game.World.AddSystem(&animationSystemWrapper{
		system: sys.animationSystem,
		logger: game.World.GetLogger(),
	})

	game.World.AddSystem(sys.equipmentVisualSystem)
	game.World.AddSystem(sys.particleSystem)
	game.World.AddSystem(sys.weatherSystem)
	game.World.AddSystem(sys.lifetimeSystem)
	game.World.AddSystem(sys.puzzleSystem)
	game.World.AddSystem(sys.firePropagationSystem)
	game.World.AddSystem(sys.destructibleSystem)
	game.World.AddSystem(sys.carrySystem)
	game.World.AddSystem(sys.hazardSystem)
	game.World.AddSystem(sys.narrativeSystem)
	game.World.AddSystem(sys.shadowSystem)

	// V4.0 System Registrations (Phase 21-27)
	// Phase 21: Vehicle systems
	game.World.AddSystem(sys.vehicleMovementSys)
	game.World.AddSystem(sys.vehicleDurabilitySys)
	game.World.AddSystem(sys.mountingSystem)
	game.World.AddSystem(sys.vehicleCombatSystem)

	// Phase 22: Companion systems (use wrappers for incompatible signatures)
	game.World.AddSystem(&companionAISystemWrapper{system: sys.companionAISystem})
	game.World.AddSystem(&companionProgressionSystemWrapper{system: sys.companionProgressionSys})
	game.World.AddSystem(&companionLoyaltySystemWrapper{system: sys.companionLoyaltySys})
	game.World.AddSystem(&companionInventorySystemWrapper{system: sys.companionInventorySys})
	game.World.AddSystem(&skillInheritanceSystemWrapper{system: sys.skillInheritanceSys})

	// Phase 23: Book system
	game.World.AddSystem(sys.bookReadingSystem)

	// Phase 24: Expanded magic systems
	game.World.AddSystem(sys.spellEffectSystem)
	game.World.AddSystem(sys.spellCombinationSys)

	// Phase 25: Class progression
	game.World.AddSystem(sys.classProgressionSys)

	// Phase 26: Expression systems (use wrappers)
	game.World.AddSystem(&expressionSystemWrapper{system: sys.expressionSystem})
	game.World.AddSystem(&expressionComboSystemWrapper{system: sys.expressionComboSys})

	// Phase 26.2: Achievement system (social features, use wrapper)
	game.World.AddSystem(&achievementSystemWrapper{system: sys.achievementSystem})

	// INTEGRATION FIX [Category A]: Phase 28 - MoralChoiceSystem registration
	// Gap: MoralChoiceSystem created but never added to world update loop
	// Fix: Registered system with wrapper for quest-driven moral decisions
	// Roadmap: ROADMAP_V4.md Phase 28.2
	game.World.AddSystem(&moralChoiceSystemWrapper{system: sys.moralChoiceSystem})

	// Phase 27: Mini-game system (use wrapper)
	game.World.AddSystem(&miniGameSystemWrapper{system: sys.miniGameSystem})

	// Phase 30: Environmental Storytelling - Discovery System (use wrapper)
	game.World.AddSystem(&discoverySystemWrapper{system: sys.discoverySystem})

	// V5.0 System Registrations (Social & Communication)
	// Phase 32: Chat system for player communication
	game.World.AddSystem(&chatSystemWrapper{system: sys.chatSystem})

	// Phase 40: Mail and courier systems for asynchronous messaging
	game.World.AddSystem(&mailSystemWrapper{system: sys.mailSystem})
	game.World.AddSystem(&courierSystemWrapper{system: sys.courierSystem})

	// V6.0 System Registrations (Persistent Worlds & Federation)
	// Phase 39: Portal system for cross-server travel
	game.World.AddSystem(&portalSystemWrapper{system: sys.portalSystem})

	// Phase 40: Bounty system for cross-server quests
	game.World.AddSystem(&bountySystemWrapper{system: sys.bountySystem})

	// Phase 41: Politics system for server diplomacy
	game.World.AddSystem(&politicsSystemWrapper{system: sys.politicsSystem})
}

// configureSystemConnections wires up interdependent systems.
func configureSystemConnections(game *engine.EbitenGame, sys *systemsContainer) {
	sys.combatSystem.SetCamera(game.CameraSystem)
	sys.combatSystem.SetParticleSystem(sys.particleSystem, game.World, *genreID)
	sys.combatSystem.SetProjectileSystem(sys.projectileSystem)
	sys.projectileSystem.SetCamera(game.CameraSystem)
	sys.projectileSystem.SetGenre(*genreID)
	sys.projectileSystem.SetSeed(*seed)
	sys.interactionSystem.SetCarrySystem(sys.carrySystem)
}

// startPerformanceMonitoring begins background performance metric logging.
func startPerformanceMonitoring(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if !*verbose {
		return
	}

	perfMonitor := engine.NewPerformanceMonitor(game.World)
	clientLogger.Info("performance monitoring initialized")

	go func() {
		ticker := time.NewTicker(perfMonitorInterval * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := perfMonitor.GetMetrics()
			clientLogger.WithField("metrics", metrics.String()).Info("performance metrics")
		}
	}()
}

// generateWorldTerrain creates and validates procedural terrain.
func generateWorldTerrain(logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.Terrain {
	clientLogger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  defaultTerrainWidth,
			"height": defaultTerrainHeight,
		},
	}

	terrainResult, err := terrainGen.Generate(*seed, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate terrain")
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	clientLogger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("terrain generated")

	return generatedTerrain
}

// initializeTerrainRendering sets up the terrain rendering system.
func initializeTerrainRendering(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing terrain rendering system")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(tileSize, tileSize, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		clientLogger.Info("terrain rendering system initialized")
	}
}

// configureLightingSystem enables and configures dynamic lighting.
func configureLightingSystem(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if !*enableLighting {
		return
	}

	clientLogger.Info("enabling dynamic lighting system")
	game.EnableLighting(true)
	game.SetLightingGenrePreset(*genreID)

	clientLogger.WithFields(logrus.Fields{
		"genre":     *genreID,
		"enabled":   true,
		"maxLights": maxLights,
	}).Info("lighting system configured")
}

// initializeTerrainCollision sets up efficient terrain collision checking.
func initializeTerrainCollision(game *engine.EbitenGame, sys *systemsContainer, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing terrain collision system")
	}

	terrainChecker := engine.NewTerrainCollisionChecker(tileSize, tileSize)
	terrainChecker.SetTerrain(generatedTerrain)

	for _, system := range game.World.GetSystems() {
		if collisionSys, ok := system.(*engine.CollisionSystem); ok {
			collisionSys.SetTerrainChecker(terrainChecker)
		}
		if projSys, ok := system.(*engine.ProjectileSystem); ok {
			projSys.SetTerrainChecker(terrainChecker)
		}
	}

	if *verbose {
		clientLogger.Info("terrain collision system initialized (efficient mode)")
	}
}

// generateWorldFactions creates and registers all world factions.
func generateWorldFactions(game *engine.EbitenGame, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	clientLogger.Info("generating world factions")

	factionGen := faction.NewGenerator()
	factionResult, err := factionGen.Generate(*seed+seedOffsetFaction, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate factions")
	}

	worldFactions := factionResult.([]*engine.Faction)
	clientLogger.WithFields(logrus.Fields{
		"count": len(worldFactions),
		"genre": *genreID,
	}).Info("factions generated")

	for _, fac := range worldFactions {
		for _, system := range game.World.GetSystems() {
			if facSys, ok := system.(*engine.FactionSystem); ok {
				facSys.AddFaction(fac)
				if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"factionID":   fac.ID,
						"factionName": fac.Name,
						"factionType": fac.Type,
					}).Debug("faction added to world")
				}
			}
		}
	}
}

// initializeSpatialPartitioning creates and configures the spatial partition system.
func initializeSpatialPartitioning(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing spatial partition system for viewport culling")
	}

	worldWidth := float64(generatedTerrain.Width) * worldPixelsPerTile
	worldHeight := float64(generatedTerrain.Height) * worldPixelsPerTile

	spatialSystem := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)
	game.World.AddSystem(spatialSystem)

	spatialSystem.Rebuild(game.World.GetEntities())

	game.RenderSystem.SetSpatialPartition(spatialSystem)
	game.RenderSystem.EnableCulling(true)
	game.RenderSystem.EnableBatching(true)

	clientLogger.WithFields(logrus.Fields{
		"worldWidth":     worldWidth,
		"worldHeight":    worldHeight,
		"cellSize":       quadtreeCapacity,
		"initialCount":   len(game.World.GetEntities()),
		"cullingActive":  true,
		"batchingActive": true,
	}).Info("spatial partition system initialized with initial rebuild, culling and batching enabled")
}

// connectMapUIToTerrain wires the map UI to terrain data.
func connectMapUIToTerrain(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("connecting terrain to Map UI")
	}

	game.MapUI.SetTerrain(generatedTerrain)

	if *verbose {
		clientLogger.Info("Map UI configured with terrain data")
	}
}

// spawnWorldEntities spawns enemies, merchants, stations, puzzles, and objects.
func spawnWorldEntities(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
	}

	// Spawn enemies
	if *verbose {
		clientLogger.Info("spawning enemies in dungeon rooms")
	}
	enemyCount, err := engine.SpawnEnemiesInTerrain(game.World, generatedTerrain, *seed, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn enemies")
	} else if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"enemyCount": enemyCount,
			"roomCount":  len(generatedTerrain.Rooms) - 1,
		}).Info("spawned enemies")
	}

	// Spawn merchants
	if *verbose {
		clientLogger.Info("spawning merchants in dungeon")
	}
	merchantCount, err := engine.SpawnMerchantsInTerrain(game.World, generatedTerrain, *seed, params, defaultMerchantCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn merchants")
	} else if *verbose {
		clientLogger.WithField("merchantCount", merchantCount).Info("spawned merchants")
	}

	// Spawn crafting stations
	if *verbose {
		clientLogger.Info("spawning crafting stations in dungeon")
	}
	stationGen := station.NewStationGenerator()
	stationCount := engine.SpawnStationsInTerrain(game.World, stationGen, generatedTerrain, tileSize, *seed+seedOffsetStation, *genreID)
	if *verbose {
		clientLogger.WithField("stationCount", stationCount).Info("spawned crafting stations")
	}

	// Spawn puzzles
	if *verbose {
		clientLogger.Info("spawning procedural puzzles in dungeon")
	}
	puzzleCount, err := engine.SpawnPuzzlesInTerrain(game.World, generatedTerrain, *seed+seedOffsetPuzzle, params, defaultPuzzleCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn puzzles")
	} else {
		clientLogger.WithFields(logrus.Fields{
			"puzzleCount": puzzleCount,
			"targetCount": defaultPuzzleCount,
			"roomCount":   len(generatedTerrain.Rooms) - 1,
		}).Info("spawned procedural puzzles")
	}

	// Spawn destructible objects
	if *verbose {
		clientLogger.Info("spawning destructible objects in dungeon")
	}
	objectCount := spawnDestructibleObjects(game.World, generatedTerrain, *seed+seedOffsetObject, *genreID, clientLogger.Logger)
	clientLogger.WithFields(logrus.Fields{
		"objectCount": objectCount,
		"roomCount":   len(generatedTerrain.Rooms) - 1,
		"genre":       *genreID,
	}).Info("spawned destructible objects")

	// Spawn vehicles (V4.0)
	if *verbose {
		clientLogger.Info("spawning vehicles in dungeon")
	}
	vehicleCount, err := spawnVehicles(game.World, generatedTerrain, *seed+seedOffsetVehicle, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn vehicles")
	} else if *verbose {
		clientLogger.WithField("vehicleCount", vehicleCount).Info("spawned vehicles")
	}

	// Spawn companions (V4.0)
	if *verbose {
		clientLogger.Info("spawning companions in dungeon")
	}
	companionCount, err := spawnCompanions(game.World, generatedTerrain, *seed+seedOffsetCompanion, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn companions")
	} else if *verbose {
		clientLogger.WithField("companionCount", companionCount).Info("spawned companions")
	}

	// Spawn bookshelves with books (V4.0)
	if *verbose {
		clientLogger.Info("spawning bookshelves in dungeon")
	}
	bookshelfCount, err := spawnBookshelves(game.World, generatedTerrain, *seed+seedOffsetBook, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn bookshelves")
	} else if *verbose {
		clientLogger.WithField("bookshelfCount", bookshelfCount).Info("spawned bookshelves")
	}

	// Spawn story fragments (Phase 30: Environmental Storytelling)
	if *verbose {
		clientLogger.Info("spawning story fragments in dungeon")
	}
	fragmentCount, err := spawnStoryFragments(game.World, generatedTerrain, *seed+seedOffsetStory, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn story fragments")
	} else if *verbose {
		clientLogger.WithField("fragmentCount", fragmentCount).Info("spawned story fragments")
	}
}

// spawnEnvironmentalEffects spawns lights and weather effects.
func spawnEnvironmentalEffects(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *enableLighting {
		if *verbose {
			clientLogger.Info("spawning environmental lights in dungeon")
		}
		lightCount := spawnEnvironmentalLights(game.World, generatedTerrain, *seed+seedOffsetLight, *genreID)
		clientLogger.WithFields(logrus.Fields{
			"lightCount": lightCount,
			"genre":      *genreID,
		}).Info("spawned environmental lights")
	}

	if *enableWeather {
		if *verbose {
			clientLogger.Info("spawning weather effects")
		}
		weatherEntity := spawnWeather(game.World, *width, *height, *seed+seedOffsetWeather, *genreID, *weatherType, *weatherIntensity)
		if weatherEntity != nil {
			clientLogger.WithFields(logrus.Fields{
				"type":      *weatherType,
				"intensity": *weatherIntensity,
				"genre":     *genreID,
			}).Info("weather effects spawned")
		}
	}
}

// calculatePlayerSpawnPosition determines where the player should spawn.
func calculatePlayerSpawnPosition(generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) (float64, float64) {
	if len(generatedTerrain.Rooms) > 0 {
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		playerX := float64(cx * tileSize)
		playerY := float64(cy * tileSize)

		if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"tileX":  cx,
				"tileY":  cy,
				"worldX": playerX,
				"worldY": playerY,
			}).Info("player spawning in first room")
		}

		return playerX, playerY
	}

	clientLogger.Warn("no rooms in terrain, using default spawn position")
	return fallbackPlayerX, fallbackPlayerY
}

// createPlayerEntity creates the player entity with all necessary components.
func createPlayerEntity(game *engine.EbitenGame, playerX, playerY float64, animationSystem *engine.AnimationSystem, clientLogger *logrus.Entry) *engine.Entity {
	if *verbose {
		clientLogger.Info("creating player entity")
	}

	player := game.World.CreateEntity()

	// Add player components
	player.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add input component for player control
	player.AddComponent(&engine.EbitenInput{})

	// Phase 10.1: Add rotation and aim components for 360° rotation and mouse aim
	player.AddComponent(engine.NewRotationComponent(0, 3.0)) // Start facing right
	player.AddComponent(engine.NewAimComponent(0))           // Start aiming right

	// Add animated sprite
	playerSprite := &engine.EbitenSprite{
		Image:   nil, // Will be created by animation system
		Width:   28,
		Height:  28,
		Visible: true,
		Layer:   10,
	}
	player.AddComponent(playerSprite)

	// Add animation component
	playerAnim := engine.NewAnimationComponent(*seed + int64(player.ID*1000))
	playerAnim.CurrentState = engine.AnimationStateIdle
	playerAnim.FrameTime = 0.15
	playerAnim.Loop = true
	playerAnim.Playing = true
	playerAnim.FrameCount = 4
	playerAnim.Dirty = true
	player.AddComponent(playerAnim)

	// Add equipment visual component
	player.AddComponent(engine.NewEquipmentVisualComponent())

	// Add camera components
	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	player.AddComponent(camera)

	player.AddComponent(engine.NewScreenShakeComponent())
	player.AddComponent(engine.NewHitStopComponent())

	// Add layer component
	layerComp := engine.NewLayerComponent()
	layerComp.CurrentLayer = 0
	player.AddComponent(&layerComp)

	// Add shadow component
	playerShadow := engine.NewShadowComponent(28)
	playerShadow.CastsShadow = true
	playerShadow.ShadowType = engine.ShadowTypeSoft
	player.AddComponent(playerShadow)

	// Add player torch if lighting enabled
	if *enableLighting {
		playerTorch := engine.NewTorchLight(200)
		playerTorch.Enabled = true
		player.AddComponent(playerTorch)

		if *verbose {
			clientLogger.WithFields(logrus.Fields{
				"radius":    200,
				"intensity": playerTorch.Intensity,
			}).Info("player torch added")
		}
	}

	// Configure camera and animation systems
	game.CameraSystem.SetActiveCamera(player)
	animationSystem.SetCameraSystem(game.CameraSystem)
	animationSystem.SetPlayerEntity(player)

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"viewportCulling": true,
			"distanceLOD":     true,
			"closeThreshold":  200.0,
			"midThreshold":    400.0,
		}).Info("animation system configured with performance optimizations")
	}

	// Set player for UI systems
	game.HUDSystem.SetPlayerEntity(player)
	game.SetPlayerEntity(player)

	clientLogger.WithField("entityID", player.ID).Info("player entity created")
	return player
}

// addPlayerComponents adds gameplay components to the player entity.
func addPlayerComponents(player *engine.Entity, logger *logrus.Logger, clientLogger *logrus.Entry) (*engine.InventoryComponent, *engine.QuestTrackerComponent) {
	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	playerStats.CritChance = 0.05
	playerStats.CritDamage = 1.5
	playerStats.Evasion = 0.05
	player.AddComponent(playerStats)

	// Add player experience/progression
	player.AddComponent(engine.NewExperienceComponent())

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0)
	playerInventory.Gold = 100
	player.AddComponent(playerInventory)

	// Add player equipment
	player.AddComponent(engine.NewEquipmentComponent())

	// Add mana component
	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0,
	}
	player.AddComponent(playerMana)

	// Load spells
	if err := engine.LoadPlayerSpells(player, *seed, *genreID, 1); err != nil {
		clientLogger.WithError(err).Fatal("failed to load player spells")
	}
	if *verbose {
		clientLogger.Info("player spells loaded (keys 1-5)")
	}

	// Load skill tree
	if err := engine.LoadPlayerSkillTree(player, *seed, *genreID, 0); err != nil {
		clientLogger.WithError(err).Fatal("failed to load skill tree")
	}
	if *verbose {
		if comp, ok := player.GetComponent("skill_tree"); ok {
			treeComp := comp.(*engine.SkillTreeComponent)
			clientLogger.WithFields(logrus.Fields{
				"treeName":   treeComp.Tree.Name,
				"skillCount": len(treeComp.Tree.Nodes),
			}).Info("skill tree loaded (press K)")
		}
	}

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5)
	player.AddComponent(questTracker)

	// Add attack capability
	player.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision
	player.AddComponent(&engine.ColliderComponent{
		Width:     28,
		Height:    28,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   -14,
		OffsetY:   -14,
	})

	// Add visual feedback
	player.AddComponent(engine.NewVisualFeedbackComponent())

	// Add achievement tracking (Phase 26.2)
	player.AddComponent(&engine.AchievementComponent{
		Achievements:     []engine.Achievement{},
		ExpressionCount:  0,
		UniqueExpression: make(map[engine.ExpressionType]int),
	})

	// Add story journal for fragment discovery (Phase 30)
	player.AddComponent(engine.NewStoryJournalComponent())

	// Add mail component for mailbox system (Phase 40.3)
	player.AddComponent(&engine.MailComponent{
		Inbox:    []*engine.MailMessage{},
		Outbox:   []*engine.MailMessage{},
		MaxInbox: 50,
	})

	// Add starter items
	clientLogger.Info("adding starter items to inventory")
	addStarterItems(playerInventory, *seed, *genreID, logger)

	// Add tutorial quest
	clientLogger.Info("creating tutorial quest")
	addTutorialQuest(questTracker, *seed, *genreID, logger)

	return playerInventory, questTracker
}

// initializeTutorialAndHelp creates and configures tutorial and help systems.
func initializeTutorialAndHelp(inputSystem *engine.InputSystem, cameraSystem *engine.CameraSystem) (*engine.EbitenTutorialSystem, *engine.EbitenHelpSystem) {
	tutorialSystem := engine.NewTutorialSystem()
	if *noTutorial {
		tutorialSystem.Enabled = false
		tutorialSystem.ShowUI = false
	}
	helpSystem := engine.NewHelpSystem()

	inputSystem.SetHelpSystem(helpSystem)
	inputSystem.SetTutorialSystem(tutorialSystem)
	inputSystem.SetCameraSystem(cameraSystem)

	return tutorialSystem, helpSystem
}

// configureSaveLoadSystem initializes the save/load manager and registers callbacks.
func configureSaveLoadSystem(player *engine.Entity, game *engine.EbitenGame, generatedTerrain *terrain.Terrain, inputSystem *engine.InputSystem, clientLogger *logrus.Entry) *saveload.SaveManager {
	clientLogger.Info("initializing save/load system")

	saveManager, err := saveload.NewSaveManager("./saves")
	if err != nil {
		clientLogger.WithError(err).Warn("failed to initialize save manager, save/load functionality will be unavailable")
		return nil
	}

	if *verbose {
		clientLogger.Info("save/load system initialized")
	}

	inputSystem.SetQuickSaveCallback(func() error {
		clientLogger.Info("quick save (F5 pressed)")
		gameSave := createGameSave(player, game, generatedTerrain)
		if err := saveManager.SaveGame("quicksave", gameSave); err != nil {
			clientLogger.WithError(err).Error("failed to save game")
			return err
		}
		clientLogger.Info("game saved successfully")
		return nil
	})

	inputSystem.SetQuickLoadCallback(func() error {
		clientLogger.Info("quick load (F9 pressed)")
		gameSave, err := saveManager.LoadGame("quicksave")
		if err != nil {
			clientLogger.WithError(err).Error("failed to load game")
			return err
		}
		loadGameSave(player, gameSave, game)
		clientLogger.Info("game loaded successfully")
		return nil
	})

	if *verbose {
		clientLogger.Info("quick save/load callbacks registered (F5/F9)")
	}

	return saveManager
}

// setupUICallbacks configures all UI input callbacks and menu system integrations.
func setupUICallbacks(game *engine.EbitenGame, player *engine.Entity, generatedTerrain *terrain.Terrain, inputSystem *engine.InputSystem, objectiveTracker *engine.ObjectiveTrackerSystem, dialogSystem *engine.DialogSystem, shopUI *engine.ShopUI, saveManager *saveload.SaveManager, clientLogger *logrus.Entry) error {
	if *verbose {
		clientLogger.Info("setting up UI input callbacks")
	}

	if err := game.SetupInputCallbacks(inputSystem, objectiveTracker); err != nil {
		return err
	}

	// BUG FIX: Menu Trap - Connect mailbox UI to input system for ESC key handling
	// Resolution: Enables dual-exit pattern (L key toggle + ESC key close) for mailbox UI
	if game.MailboxUI != nil {
		inputSystem.SetMailboxUI(game.MailboxUI)
	}

	if *verbose {
		clientLogger.Info("UI callbacks registered (I: Inventory, J: Quests, ESC: Pause Menu)")
		clientLogger.Info("inventory actions: E to equip/use, D to drop")
	}

	if err := setupMerchantInteraction(player, game, dialogSystem, shopUI, inputSystem, clientLogger); err != nil {
		return err
	}

	if *verbose {
		clientLogger.Info("merchant interaction registered (F key when near merchant)")
	}

	if err := connectMenuSaveLoad(game, player, generatedTerrain, saveManager, clientLogger); err != nil {
		return err
	}

	return nil
}

// setupMerchantInteraction configures the F key interaction callback for merchants.
func setupMerchantInteraction(player *engine.Entity, game *engine.EbitenGame, dialogSystem *engine.DialogSystem, shopUI *engine.ShopUI, inputSystem *engine.InputSystem, clientLogger *logrus.Entry) error {
	return inputSystem.SetInteractCallback(func() {
		if player == nil {
			return
		}
		posComp, ok := player.GetComponent("position")
		if !ok {
			return
		}
		pos := posComp.(*engine.PositionComponent)

		merchant, dist := engine.FindClosestMerchant(game.World, pos.X, pos.Y, merchantInteractionRange)
		if merchant == nil {
			if *verbose {
				clientLogger.Debug("no merchant nearby to interact with")
			}
			return
		}

		success, err := dialogSystem.StartDialog(player.ID, merchant.ID)
		if err != nil {
			clientLogger.WithError(err).Warn("failed to start dialog")
			return
		}

		if !success {
			if *verbose {
				clientLogger.Debug("dialog could not be started")
			}
			return
		}

		shopUI.Open(merchant)

		if *verbose {
			clientLogger.WithField("distance", dist).Debug("opened shop with merchant")
		}
	})
}

// connectMenuSaveLoad wires save/load callbacks to the menu system.
func connectMenuSaveLoad(game *engine.EbitenGame, player *engine.Entity, generatedTerrain *terrain.Terrain, saveManager *saveload.SaveManager, clientLogger *logrus.Entry) error {
	if game.MenuSystem == nil || saveManager == nil {
		return nil
	}

	if *verbose {
		clientLogger.Info("connecting save/load callbacks to menu system")
	}

	saveCallback := func(saveName string) error {
		if *verbose {
			clientLogger.WithField("saveName", saveName).Info("menu save")
		}
		gameSave := createGameSave(player, game, generatedTerrain)
		if err := saveManager.SaveGame(saveName, gameSave); err != nil {
			clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to save game")
			return err
		}
		clientLogger.WithField("saveName", saveName).Info("game saved successfully")
		return nil
	}

	loadCallback := func(saveName string) error {
		if *verbose {
			clientLogger.WithField("saveName", saveName).Info("menu load")
		}
		gameSave, err := saveManager.LoadGame(saveName)
		if err != nil {
			clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to load game")
			return err
		}
		loadGameSave(player, gameSave, game)
		clientLogger.WithField("saveName", saveName).Info("game loaded successfully")
		return nil
	}

	game.MenuSystem.SetSaveCallback(saveCallback)
	game.MenuSystem.SetLoadCallback(loadCallback)

	if *verbose {
		clientLogger.Info("save/load callbacks connected to menu system")
	}

	return nil
}

// initializeUIIntegration sets up shop UI, crafting UI, mailbox UI, and connects them to game systems.
func initializeUIIntegration(game *engine.EbitenGame, player *engine.Entity, commerceSystem *engine.CommerceSystem, dialogSystem *engine.DialogSystem, craftingSystem *engine.CraftingSystem, inventorySystem *engine.InventorySystem, clientLogger *logrus.Entry) (*engine.ShopUI, *engine.CraftingUI) {
	shopUI := engine.NewShopUI(*width, *height)
	shopUI.SetPlayerEntity(player)
	shopUI.SetCommerceSystem(commerceSystem)
	shopUI.SetDialogSystem(dialogSystem)
	game.ShopUI = shopUI

	if *verbose {
		clientLogger.Info("shop UI initialized and connected to commerce/dialog systems")
	}

	craftingUI := engine.NewCraftingUI(*width, *height)
	craftingUI.SetPlayerEntity(player)
	craftingUI.SetCraftingSystem(craftingSystem)
	game.CraftingUI = craftingUI

	if *verbose {
		clientLogger.Info("crafting UI initialized and connected to crafting system")
	}

	// Phase 40.3: Initialize mailbox UI
	mailboxUI := engine.NewMailboxUI(0, 0, *width, *height, *genreID)
	game.MailboxUI = mailboxUI

	if *verbose {
		clientLogger.Info("mailbox UI initialized (Phase 40.3)")
	}

	game.SetInventorySystem(inventorySystem)

	return shopUI, craftingUI
}

// applyCharacterClass applies character class stats if character data is pending.
func applyCharacterClass(player *engine.Entity, game *engine.EbitenGame, clientLogger *logrus.Entry) {
	charData := game.GetPendingCharacterData()
	if charData == nil {
		return
	}

	clientLogger.WithFields(logrus.Fields{
		"name":  charData.Name,
		"class": charData.Class.String(),
	}).Info("applying character class stats")

	if err := engine.ApplyClassStats(player, charData.Class); err != nil {
		clientLogger.WithError(err).Fatal("failed to apply character class stats")
	}
}

// finalizeGameInitialization processes initial entity updates and logs game start info.
func finalizeGameInitialization(game *engine.EbitenGame, player *engine.Entity, networkClient interface{}, clientLogger *logrus.Entry) {
	game.World.Update(0)

	clientLogger.Info("game initialized successfully")
	clientLogger.Info("controls: WASD to move, Space to attack, E to use item, I: Inventory, J: Quests, L: Mailbox")
	clientLogger.WithFields(logrus.Fields{"genre": *genreID, "seed": *seed}).Info("game settings")

	if *multiplayer {
		clientLogger.WithField("server", *server).Info("multiplayer connected")
	}
}

// handleHostAndPlay starts embedded server if host-and-play mode is enabled.
func handleHostAndPlay(logger *logrus.Logger, clientLogger *logrus.Entry) {
	if !*hostAndPlay {
		return
	}

	// Log message depends on whether this was explicitly requested or auto-enabled
	if *hostLAN {
		clientLogger.Info("host-and-play mode enabled (explicit) - starting LAN-accessible server")
	} else {
		clientLogger.Info("host-and-play mode enabled - starting localhost server")
	}

	serverAddr, cleanup, err := startEmbeddedServer(logger, *seed, *genreID)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to start embedded server")
	}
	defer cleanup()

	*server = serverAddr
	*multiplayer = true

	clientLogger.WithField("serverAddr", serverAddr).Info("embedded server started, connecting client")
}

// createGameInstance initializes the main game instance with logging and profiling.
func createGameInstance(logger *logrus.Logger, clientLogger *logrus.Entry) *engine.EbitenGame {
	game := engine.NewEbitenGameWithLogger(*width, *height, logger)
	game.SetFullscreen(*fullscreen)

	// Set world seed for deterministic character naming
	game.SetWorldSeed(*seed)

	if *profile {
		game.EnableFrameTimeProfiling()
		clientLogger.Info("performance profiling enabled - frame time stats will be logged every 5 seconds")
	}

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"width":      *width,
			"height":     *height,
			"fullscreen": *fullscreen,
			"seed":       *seed,
		}).Info("display initialized")
	}

	return game
}

// initializeVirtualControls sets up touch controls for mobile/web platforms.
func initializeVirtualControls(inputSystem *engine.InputSystem, clientLogger *logrus.Entry) {
	if !mobile.IsTouchCapable() {
		return
	}

	inputSystem.InitializeVirtualControls(*width, *height)
	clientLogger.WithFields(logrus.Fields{
		"platform": mobile.GetPlatform().String(),
		"width":    *width,
		"height":   *height,
	}).Info("virtual controls initialized for touch-capable platform")
}

// configureDeathCallback sets up the death callback for combat system.
func configureDeathCallback(sys *systemsContainer, game *engine.EbitenGame, logger *logrus.Logger) {
	var playerEntity *engine.Entity
	sys.combatSystem.SetDeathCallback(createDeathCallback(
		game, &playerEntity, sys.objectiveTracker, &sys.audioManager,
		sys.recipeGen, *seed, *genreID, logger,
	))
}

// createGenerationParams creates standard generation parameters for world content.
func createGenerationParams() procgen.GenerationParams {
	return procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  defaultTerrainWidth,
			"height": defaultTerrainHeight,
		},
	}
}

// cleanupNetworkClient disconnects the network client on shutdown.
func cleanupNetworkClient(networkClient interface{}, clientLogger *logrus.Entry) {
	if networkClient == nil {
		return
	}

	type disconnector interface {
		Disconnect() error
	}

	if dc, ok := networkClient.(disconnector); ok {
		clientLogger.Info("disconnecting from server")
		if err := dc.Disconnect(); err != nil {
			clientLogger.WithError(err).Warn("error disconnecting")
		}
	}
}

// INTEGRATION FIX [Category A]: System Wrappers for Client V4.0+
// Gap: Missing wrappers prevented proper system registration
// Fix: Added wrappers to adapt Update() signatures to ECS interface
// Roadmap: ROADMAP_V4.md Phases 28, 30

// discoverySystemWrapper adapts DiscoverySystem to the System interface.
type discoverySystemWrapper struct {
	system *engine.DiscoverySystem
}

func (w *discoverySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// moralChoiceSystemWrapper adapts MoralChoiceSystem to the System interface.
type moralChoiceSystemWrapper struct {
	system *engine.MoralChoiceSystem
}

func (w *moralChoiceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// V6.0 System Wrappers (Persistent Worlds & Federation)
// Roadmap: ROADMAP_V6.md Phases 39-42

// portalSystemWrapper adapts PortalSystem to the System interface.
type portalSystemWrapper struct {
	system *federation.PortalSystem
}

func (w *portalSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// bountySystemWrapper adapts BountySystem to the System interface.
type bountySystemWrapper struct {
	system *engine.BountySystem
}

func (w *bountySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// politicsSystemWrapper adapts PoliticsSystem to the System interface.
type politicsSystemWrapper struct {
	system *engine.PoliticsSystem
}

func (w *politicsSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// runGameLoop starts the main game loop.
func runGameLoop(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	windowTitle := fmt.Sprintf("Venture %s - Procedural Action RPG", version.Version)
	if err := game.Run(windowTitle); err != nil {
		clientLogger.WithError(err).Fatal("error running game")
	}
}
