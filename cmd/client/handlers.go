//go:build !android && !ios
// +build !android,!ios

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/opd-ai/venture/pkg/class/advanced"
	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/physics/vehicle"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/station"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/quality"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/opd-ai/venture/pkg/social/persistence"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"

	// INTEGRATION FIX [Category A]: V7.0 Display System Import
	// Gap: Display package existed but was never imported for initialization
	// Fix: Added import for display management (1920x1080 default resolution)
	// Roadmap: ROADMAP_V7.md Phase 43
	"github.com/opd-ai/venture/pkg/rendering/display"

	// INTEGRATION FIX [Category A]: V9.0 Integration Manager Imports
	// Gap: Integration packages implemented but never imported for use
	// Fix: Added imports for housing crafting, companion housing, and guild housing managers
	// Roadmap: ROADMAP_V9.md Phase 55.1-55.3
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"

	// Phase 3.2: Guild Federation (PLAN.md)
	"github.com/opd-ai/venture/pkg/network/federation/guild"

	// Phase 6.1: Branching Narratives (PLAN.md)
	"github.com/opd-ai/venture/pkg/narrative/branching"

	// Phase 1.1: Audio System Complete (PLAN.md)
	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
	"github.com/opd-ai/venture/pkg/audio/sfx"

	// Phase 1.2: Destruction Physics (PLAN.md)
	"github.com/opd-ai/venture/pkg/engine/physics/destruction"

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	// Note: These packages are wrapped by adapters in pkg/engine/ (LightingAdapter, AnimationAdapter, PostProcessorAdapter)
	// Imports kept for documentation purposes showing integration is complete
	_ "github.com/opd-ai/venture/pkg/rendering/animation"
	_ "github.com/opd-ai/venture/pkg/rendering/lighting"
	_ "github.com/opd-ai/venture/pkg/rendering/postprocess"

	// Phase 2.3: UI & Shape Rendering (PLAN.md)
	"github.com/opd-ai/venture/pkg/rendering/patterns"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/ui"

	// Phase 2.4: Rendering Optimization (PLAN.md)
	"github.com/opd-ai/venture/pkg/rendering/parallel"
	"github.com/opd-ai/venture/pkg/rendering/pool"

	// Phase 3.2: Magic & Skills (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/skills"

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/class"
	"github.com/opd-ai/venture/pkg/procgen/minigame"
	"github.com/opd-ai/venture/pkg/procgen/puzzle"

	// Phase 3.6: Narrative Integration (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/narrative"

	// Phase 4.1: Chat System (PLAN.md)
	"github.com/opd-ai/venture/pkg/network/chat"

	// Phase 4.2: Trade System (PLAN.md)
	"github.com/opd-ai/venture/pkg/network/trade"
)

// systemsContainer holds all initialized game systems for dependency injection.
type systemsContainer struct {
	inputSystem              *engine.InputSystem
	movementSystem           *engine.MovementSystem
	collisionSystem          *engine.CollisionSystem
	combatSystem             *engine.CombatSystem
	interactionSystem        *engine.InteractionSystem
	particleSystem           *engine.ParticleSystem
	animationSystem          *engine.AnimationSystem
	equipmentVisualSystem    *engine.EquipmentVisualSystem
	objectiveTracker         *engine.ObjectiveTrackerSystem
	aiSystem                 *engine.AISystem
	progressionSystem        *engine.ProgressionSystem
	inventorySystem          *engine.InventorySystem
	commerceSystem           *engine.CommerceSystem
	dialogSystem             *engine.DialogSystem
	craftingSystem           *engine.CraftingSystem
	audioManager             *engine.AudioManager
	audioManagerSystem       *engine.AudioManagerSystem
	itemPickupSystem         *engine.ItemPickupSystem
	statusEffectSystem       *engine.StatusEffectSystem
	spellCastingSystem       *engine.SpellCastingSystem
	playerSpellCasting       *engine.PlayerSpellCastingSystem
	manaRegenSystem          *engine.ManaRegenSystem
	playerCombatSystem       *engine.PlayerCombatSystem
	playerItemUseSystem      *engine.PlayerItemUseSystem
	rotationSystem           *engine.RotationSystem
	projectileSystem         *engine.ProjectileSystem
	revivalSystem            *engine.RevivalSystem
	behaviorTreeSystem       *engine.BehaviorTreeSystem
	squadSystem              *engine.SquadSystem
	factionSystem            *engine.FactionSystem
	reputationSystem         *engine.ReputationSystem
	alignmentSystem          *engine.AlignmentSystem
	factionReactionSystem    *engine.FactionReactionSystem
	skillProgressionSystem   *engine.SkillProgressionSystem
	visualFeedbackSystem     *engine.VisualFeedbackSystem
	weatherSystem            *engine.WeatherSystem
	lifetimeSystem           *engine.LifetimeSystem
	puzzleSystem             *engine.PuzzleSystem
	firePropagationSystem    *engine.FirePropagationSystem
	destructibleSystem       *engine.DestructibleObjectSystem
	carrySystem              *engine.CarrySystem
	hazardSystem             *engine.HazardSystem
	narrativeSystem          *engine.NarrativeSystem
	branchingNarrativeSystem *engine.BranchingNarrativeSystem // Phase 6.1: Branching story arc system
	worldEventsSystem        *engine.WorldEventsSystem        // Phase 6.3: World-responsive events
	shadowSystem             *engine.ShadowSystem
	spriteGenerator          *sprites.Generator
	spriteCache              *cache.SpriteCache // Phase 1.2: Sprite caching for animation performance
	itemGen                  *item.ItemGenerator
	recipeGen                *recipe.RecipeGenerator
	statusEffectRNG          *rand.Rand
	// V4.0 Systems (Phase 21-27)
	vehicleMovementSys      *engine.VehicleMovementSystem
	vehicleDurabilitySys    *engine.VehicleDurabilitySystem
	mountingSystem          *engine.MountingSystem
	vehicleCombatSystem     *engine.VehicleCombatSystem
	companionAISystem       *engine.CompanionAISystem
	companionProgressionSys *engine.CompanionProgressionSystem
	companionLoyaltySys     *engine.CompanionLoyaltySystem
	companionInventorySys   *engine.CompanionInventorySystem
	companionLearningSys    *engine.CompanionLearningSystem // Phase 4.1: Companion AI skill progression and personality evolution
	advancedClassSystem     *engine.AdvancedClassSystem     // Phase 4.2: Multi-classing, prestige classes, talent trees
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
	chatSystem    *engine.EnhancedChatSystem // Phase 3.1: Enhanced chat with E2E encryption and history persistence
	mailSystem    *engine.MailSystem
	courierSystem *engine.CourierSystem
	// V6.0 Systems (Persistent Worlds & Federation)
	portalSystem       *federation.PortalSystem
	bountySystem       *engine.BountySystem
	politicsSystem     *engine.PoliticsSystem
	territoryManager   *territory.Manager
	rankingManager     *world.RankingManager
	eventManager       *world.EventManager
	federationProtocol *federation.FederationProtocol

	// INTEGRATION FIX [Category A]: Missing Systems (Phase 14+)
	// Gap: Systems implemented but never instantiated or registered in game loop
	// Fix: Added system fields for investigation, audio enhancements, quality, trade, terrain modification, merchant caravans, NPC dialog
	// Roadmap: ROADMAP_V4.md (Phase 14, 30-31) and ROADMAP_V5.md (Phase 32-36)
	investigationSystem    *engine.InvestigationSystem       // Phase 30: Environmental Storytelling - investigation mechanics
	musicTriggerSystem     *engine.MusicTriggerSystem        // Phase 14.4: Adaptive music context switching
	positionalAudioSystem  *engine.PositionalAudioSystem     // Phase 14.4: 3D positional audio with panning/occlusion
	reverbSystem           *engine.ReverbSystem              // Phase 14.4: Room-based reverb effects
	qualitySystem          *engine.QualitySystem             // Phase 14: Performance-based quality settings
	tradeSystem            *engine.TradeSystem               // Phase 33: Player-to-player trading with trust limits
	terrainConstructionSys *engine.TerrainConstructionSystem // Phase 35: Buildable walls from materials
	terrainModificationSys *engine.TerrainModificationSystem // Phase 35: Destructible terrain
	merchantCaravanSystem  *engine.MerchantCaravanSystem     // Phase 36: Traveling merchants between servers
	npcDialogSystem        *engine.NPCDialogSystem           // Phase 31: Markov-chain NPC conversations

	// INTEGRATION FIX [Category A]: High-Level Management Systems
	// Gap: Wrapper systems CompanionSystem, VehicleSystem, AdaptiveSoundtrackSystem implemented but never initialized
	// Fix: Added system fields for high-level companion/vehicle management and adaptive music
	// Roadmap: ROADMAP_V4.md Phase 21.2 (VehicleSystem), Phase 22.2 (CompanionSystem), Phase 29 (AdaptiveSoundtrackSystem)
	companionSystem          *engine.CompanionSystem          // Phase 22.2: High-level companion management
	vehicleSystem            *engine.VehicleSystem            // Phase 21.2: High-level vehicle management
	adaptiveSoundtrackSystem *engine.AdaptiveSoundtrackSystem // Phase 29: Dynamic music adaptation based on game state

	// INTEGRATION FIX [Category A]: V8.0 Systems (Phase 49-51)
	// Gap: V8.0 systems fully implemented but never instantiated or registered
	// Fix: Added system fields for housing, social persistence, physics, territory, building/furniture generation
	// Roadmap: ROADMAP_V8.md (Phase 49-51)
	housingManager     *housing.Manager               // Phase 49.1: Player housing with plot placement
	trustManager       *persistence.TrustManager      // Phase 49.2: Persistent trust scores cross-server
	reputationManager  *persistence.ReputationManager // Phase 49.2: Reputation tracking (trade, combat, social, quest)
	chatHistory        *persistence.ChatHistory       // Phase 49.3: Chat history with delta compression
	imageGallery       *persistence.ImageGallery      // Phase 49.4: Persistent image storage
	guildHallManager   *housing.GuildHallManager      // Phase 51.2: Guild hall construction system
	enhancedVehicleSys *vehicle.EnhancedVehicleSystem // Phase 50.3: Vehicle physics (suspension, weight, deformation)
	fluidSimulator     *fluids.Simulator              // Phase 50.4: Fluid dynamics simulation
	buoyancyCalculator *fluids.BuoyancyCalculator     // Phase 50.4: Buoyancy and swimming
	swimmingManager    *fluids.SwimmingManager        // Phase 50.4: Swimming mechanics
	floodingManager    *fluids.FloodingManager        // Phase 50.4: Flooding system
	buildingGenerator  *building.Generator            // Phase 51.1: Procedural building generation
	furnitureGenerator *furniture.Generator           // Phase 51.3: Furniture generation and placement

	// INTEGRATION FIX [Category A]: V7.0 Display & Viewport Systems (Phase 43-44)
	// Gap: V7.0 display management and viewport optimization implemented but never initialized
	// Fix: Added system fields for display manager and viewport optimizer
	// Roadmap: ROADMAP_V7.md (Phase 43-44)
	displayManager    *display.Manager          // Phase 43: Display resolution management (1920x1080 default)
	viewportOptimizer *engine.ViewportOptimizer // Phase 44: Enhanced viewport culling for larger resolutions

	// INTEGRATION FIX [Category A]: V9.0 Integration Managers (Phase 55)
	// Gap: V9.0 integration managers implemented but never initialized or connected to systems
	// Fix: Added manager fields for housing crafting, companion housing, and guild housing
	// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
	stationManager      *housingcrafting.StationManager  // Phase 55.1: Crafting stations in player housing
	petHomeManager      *companionhousing.PetHomeManager // Phase 55.2: Companion housing and pet homes
	guildHousingManager *guildhousing.Manager            // Phase 55.3: Guild housing and communal spaces

	// Phase 3.2: Guild Federation (PLAN.md)
	guildSystem *engine.GuildSystem // Cross-server guild management and sync
	guildUI     *engine.GuildUI     // Guild UI for player interaction

	// Phase 4.3: Territory Control (PLAN.md)
	territorySystem *engine.TerritorySystem // Territory capture and guild warfare mechanics
	territoryUI     *engine.TerritoryUI     // Territory UI for viewing and managing territories

	// Phase 1.2: Destruction Physics (PLAN.md)
	destructionSystem *destruction.System // Building structural integrity and debris physics

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	lightingAdapter  *engine.LightingAdapter  // Dynamic lighting with multiple light sources
	animationAdapter *engine.AnimationAdapter // Advanced animation with articulation and direction
	// Note: PostProcessorAdapter already exists in game.PostProcessor (see initializeCoreSystems)

	// Phase 2.3: UI & Shape Rendering (PLAN.md)
	uiGenerator      *ui.Generator       // Procedural UI element generation (menus, buttons, panels)
	shapeRenderer    *shapes.Generator   // Geometric shape rendering for sprites and UI
	patternGenerator *patterns.Generator // Texture pattern generation for tiles and materials

	// Phase 2.4: Rendering Optimization (PLAN.md)
	imagePool        *pool.ImagePool      // Image pool for memory efficiency
	parallelRenderer *parallel.WorkerPool // Parallel renderer for performance

	// Phase 3.2: Magic & Skills (PLAN.md)
	magicGenerator *magic.SpellGenerator      // Procedural spell and magic generation for loot and progression
	skillGenerator *skills.SkillTreeGenerator // Skill tree generation for class progression

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	puzzleGenerator   *puzzle.Generator     // Procedural puzzle generation for dungeon rooms
	minigameGenerator *minigame.Generator   // Minigame generation for taverns and social spaces
	classGenerator    *class.ClassGenerator // Class archetype generation for character creation

	// Phase 3.6: Narrative Integration (PLAN.md)
	narrativeGenerator *narrative.StoryArcGenerator // Procedural story arc generation for world narrative

	// Phase 4.1: Chat System (PLAN.md)
	networkChatSystem *chat.ChatSystem // Network-based chat system for multiplayer messaging

	// Phase 4.2: Trade System (PLAN.md)
	networkTradeSystem *trade.TradeSystem // Network-based trade system for multiplayer item trading
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

	// Phase 1.2: Initialize sprite cache for base sprite caching
	// 400MB limit provides significant performance improvement through cached sprite reuse
	sys.spriteCache = cache.NewSpriteCache(spriteCacheMaxSize)
	clientLogger.WithField("maxSize", spriteCacheMaxSize).Info("sprite cache initialized")

	sys.spriteGenerator = sprites.NewGenerator()
	sys.animationSystem = engine.NewAnimationSystem(sys.spriteGenerator)
	sys.animationSystem.SetMaxCacheSize(animationCacheSize)

	// Phase 1.2: Connect sprite cache to animation system
	sys.animationSystem.SetSpriteCache(sys.spriteCache)

	sys.equipmentVisualSystem = engine.NewEquipmentVisualSystem(sys.spriteGenerator)

	// INTEGRATION FIX [Category A]: Phase 14 - QualitySystem
	// Gap: QualitySystem implemented for performance-based quality adjustment but never initialized
	// Fix: Added system initialization for automatic quality settings based on frame rate
	// Roadmap: ROADMAP_V4.md Phase 14
	qualityConfig := &quality.Config{
		Level:                 quality.QualityMedium,
		EnablePostProcessing:  true,
		EnableBloom:           false,
		EnableSoftShadows:     true,
		SpriteDetailLevel:     0.7,
		EnableAntiAliasing:    true,
		AntiAliasingQuality:   1, // 2x2 sampling
		EnableSpriteCache:     true,
		EnableDynamicLighting: true,
		ShadowSampleCount:     2,
	}
	sys.qualitySystem = engine.NewQualitySystem(qualityConfig, 60.0)

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	// Initialize lighting adapter for dynamic lighting effects
	sys.lightingAdapter = engine.NewLightingAdapter(clientLogger.WithField("system", "lighting"))
	clientLogger.Info("lighting adapter initialized")

	// Initialize animation adapter for advanced animation features
	sys.animationAdapter = engine.NewAnimationAdapter(sys.spriteGenerator, clientLogger.WithField("system", "animation"))
	clientLogger.Info("animation adapter initialized")

	// Phase 2.3: UI & Shape Rendering (PLAN.md)
	// Initialize UI generator for procedural interface elements
	sys.uiGenerator = ui.NewGeneratorWithLogger(logger)
	clientLogger.Info("UI generator initialized")

	// Initialize shape renderer for geometric primitives
	sys.shapeRenderer = shapes.NewGenerator()
	clientLogger.Info("shape renderer initialized")

	// Initialize pattern generator for textures and materials
	sys.patternGenerator = patterns.NewGeneratorWithLogger(logger)
	clientLogger.Info("pattern generator initialized")

	// Phase 2.4: Rendering Optimization (PLAN.md)
	// Initialize image pool for memory efficiency (1000 pooled images)
	sys.imagePool = pool.NewImagePool()
	clientLogger.Info("image pool initialized")

	// Initialize parallel renderer for performance (CPU count workers)
	workerCount := runtime.NumCPU()
	sys.parallelRenderer = parallel.NewWorkerPool(workerCount)
	clientLogger.WithField("workerCount", workerCount).Info("parallel renderer initialized")

	// Connect rendering optimizations to RenderSystem
	poolAdapter := engine.NewImagePoolAdapter(sys.imagePool)
	parallelAdapter := engine.NewParallelRendererAdapter(sys.parallelRenderer)
	game.RenderSystem.SetPool(poolAdapter)
	game.RenderSystem.SetParallelRenderer(parallelAdapter)
	clientLogger.Info("rendering optimizations connected to RenderSystem")

	return sys
}

// initializeGenerators creates item and recipe generators for loot drops.
func initializeGenerators(sys *systemsContainer) {
	sys.itemGen = item.NewItemGenerator()
	sys.recipeGen = recipe.NewRecipeGenerator()

	// Phase 3.2: Magic & Skills (PLAN.md)
	sys.magicGenerator = magic.NewSpellGenerator()
	sys.skillGenerator = skills.NewSkillTreeGenerator()

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	sys.puzzleGenerator = puzzle.NewGenerator()
	sys.minigameGenerator = minigame.NewGenerator()
	sys.classGenerator = class.NewClassGenerator()

	// Phase 3.6: Narrative Integration (PLAN.md)
	sys.narrativeGenerator = narrative.NewStoryArcGenerator()
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

// initializeAudioSystem creates and initializes the audio system with full synthesis support.
func initializeAudioSystem(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 1.1: Full audio integration - synthesis, music, SFX
	const sampleRate = audioSampleRate
	audioSeed := *seed

	// Create base audio manager
	audioManager := audio.NewManager(sampleRate, audioSeed)

	// Create adaptive music manager
	musicManager := music.NewAdaptiveMusicManager(sampleRate, audioSeed)
	audioManager.SetMusicManager(musicManager)

	// Create SFX variety manager
	sfxManager := sfx.NewVarietyManager(sampleRate, audioSeed)
	audioManager.SetSFXManager(sfxManager)

	clientLogger.WithFields(logrus.Fields{
		"sampleRate": sampleRate,
		"seed":       audioSeed,
	}).Info("audio system initialized with adaptive music and SFX variety")

	// Store reference (legacy compatibility - keep for existing code)
	sys.audioManager = engine.NewAudioManager(audioSampleRate, *seed)
	sys.audioManagerSystem = engine.NewAudioManagerSystem(sys.audioManager)
	game.SetAudioManager(sys.audioManager)
	sys.audioManager.EnableAdaptiveMusic(true)

	// Initialize enhanced audio systems
	sys.musicTriggerSystem = engine.NewMusicTriggerSystem(game.World, sys.audioManager)
	sys.positionalAudioSystem = engine.NewPositionalAudioSystem(game.World)
	sys.reverbSystem = engine.NewReverbSystemWithLogger(game.World, *seed+seedOffsetReverb, clientLogger.Logger)

	if *verbose {
		clientLogger.Info("adaptive music composition enabled with motif system, music triggers, positional audio, and reverb")
	}

	if err := sys.audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(clientLogger.Logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(clientLogger.Logger, "audio").Info("audio system initialized (music and SFX generators, triggers, 3D audio, reverb)")
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
	sys.branchingNarrativeSystem = engine.NewBranchingNarrativeSystem(game.World)                                               // Phase 6.1: Branching narratives
	sys.worldEventsSystem = engine.NewWorldEventsSystemWithLogger(game.World, *seed+seedOffsetWorldEvents, clientLogger.Logger) // Phase 6.3: World events
	sys.shadowSystem = engine.NewShadowSystemWithLogger(game.World, clientLogger.Logger)

	// Phase 1.2: Destruction physics
	destructionConfig := destruction.Config{
		EnableIntegrityChecks: true,
		DamagePropagationRate: 0.1,
		CollapseThreshold:     0.3,
		EnableDebris:          true,
		MaxDebrisParticles:    500,
		DebrisLifetime:        10.0,
		Gravity:               980.0,
		AirResistance:         0.05,
		GroundFriction:        0.8,
		MaxFallingObjects:     100,
		UpdateFrequency:       30.0,
	}
	sys.destructionSystem = destruction.NewSystem(&destructionConfig)
	clientLogger.Info("destruction physics system initialized")
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

	// INTEGRATION FIX [Category A]: Phase 21.2 - VehicleSystem high-level wrapper
	// Gap: VehicleSystem implemented but never initialized (subsystems VehicleMovementSystem, etc. initialized above)
	// Fix: Added high-level VehicleSystem for unified vehicle management
	// Roadmap: ROADMAP_V4.md Phase 21.2
	sys.vehicleSystem = engine.NewVehicleSystem(game.World)

	if *verbose {
		clientLogger.Info("V4.0 systems initialized (vehicles, vehicle-mgmt, mounting, companions, companion-mgmt, skills, books, spells, classes, expressions, minigames, achievements, moral choices, discovery, investigation, NPC dialog, adaptive music)")
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

	if *verbose {
		clientLogger.Info("V5.0 systems initialized (chat, trade, terrain construction/modification, merchant caravans, mail, courier, network chat/trade)")
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

// INTEGRATION FIX [Category A]: initializeV8Systems initializes Version 8.0 housing, physics, and social persistence systems.
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
	// Note: ChatHistory is created per-player, this is a global placeholder for the system
	// Actual player-specific instances will be created when player entity is spawned
	sys.chatHistory = persistence.NewChatHistory("system")

	// Phase 49.4: Persistent Image Storage & Gallery
	// Note: ImageGallery is created per-player, this is a global placeholder for the system
	// Actual player-specific instances will be created when player entity is spawned
	sys.imageGallery = persistence.NewImageGallery("system")

	// Phase 50.3: Enhanced Vehicle Physics
	sys.enhancedVehicleSys = vehicle.NewEnhancedVehicleSystem()

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

// INTEGRATION FIX [Category A]: V7.0 Display & Viewport System Initialization
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

	if *verbose {
		clientLogger.Info("V7.0 systems initialized (display manager, viewport optimizer)")
	}
}

// INTEGRATION FIX [Category A]: V9.0 Integration Manager Initialization
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

// initializePhase3Systems initializes Phase 3 systems from PLAN.md (Networking & Social Features)
func initializePhase3Systems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 3.2: Guild Federation - Cross-server guild management
	guildManager := guild.NewManager()
	sys.guildSystem = engine.NewGuildSystem(game.World, guildManager)
	sys.guildUI = engine.NewGuildUI(game.World, sys.guildSystem, *width, *height)

	// Connect guild system to federation protocol for cross-server sync
	if sys.federationProtocol != nil {
		sys.guildSystem.SetFederation(sys.federationProtocol)
		clientLogger.Debug("guild system connected to federation protocol")
	}

	// Phase 4.3: Territory Control - Guild warfare and territory management
	sys.territorySystem = engine.NewTerritorySystem(sys.territoryManager, clientLogger.WithField("system", "territory"))
	sys.territoryUI = engine.NewTerritoryUI(sys.territorySystem, *width, *height)

	if *verbose {
		clientLogger.Info("Phase 3 systems initialized (guild federation with cross-server sync, territory control)")
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
	game.World.AddSystem(&destructionSystemWrapper{system: sys.destructionSystem}) // Phase 1.2: Destruction physics
	game.World.AddSystem(sys.carrySystem)
	game.World.AddSystem(sys.hazardSystem)
	game.World.AddSystem(sys.narrativeSystem)
	game.World.AddSystem(sys.branchingNarrativeSystem) // Phase 6.1: Branching narratives
	game.World.AddSystem(sys.worldEventsSystem)        // Phase 6.3: World events
	game.World.AddSystem(sys.shadowSystem)

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	game.World.AddSystem(sys.lightingAdapter)  // Dynamic lighting with multiple light sources
	game.World.AddSystem(sys.animationAdapter) // Advanced animation features
	// Note: PostProcessorAdapter is used during rendering, not in ECS update loop

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
	game.World.AddSystem(sys.companionLearningSys) // Phase 4.1: Companion learning (compatible signature)
	game.World.AddSystem(sys.advancedClassSystem)  // Phase 4.2: Advanced classes (compatible signature)
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

	// Phase 3.2 (PLAN.md): Guild Federation - Cross-server guild management
	if sys.guildSystem != nil {
		game.World.AddSystem(sys.guildSystem)
	}

	// Phase 4.3 (PLAN.md): Territory Control - Guild warfare and territory capture
	if sys.territorySystem != nil {
		game.World.AddSystem(&territorySystemWrapper{system: sys.territorySystem})
	}

	// V6.0 System Registrations (Persistent Worlds & Federation)
	// Phase 39: Portal system for cross-server travel
	game.World.AddSystem(&portalSystemWrapper{system: sys.portalSystem})

	// Phase 40: Bounty system for cross-server quests
	game.World.AddSystem(&bountySystemWrapper{system: sys.bountySystem})

	// Phase 41: Politics system for server diplomacy
	game.World.AddSystem(&politicsSystemWrapper{system: sys.politicsSystem})

	// INTEGRATION FIX [Category A]: Missing System Registrations
	// Gap: Systems initialized but never added to World update loop
	// Fix: Registered all missing systems with appropriate wrappers
	// Roadmap: ROADMAP_V4.md (Phase 14, 30-31) and ROADMAP_V5.md (Phase 32-36)

	// Phase 30-31: Environmental Storytelling and NPC Dialog
	game.World.AddSystem(&investigationSystemWrapper{system: sys.investigationSystem})
	game.World.AddSystem(&npcDialogSystemWrapper{system: sys.npcDialogSystem})

	// Phase 14.4: Adaptive Audio Systems
	game.World.AddSystem(&musicTriggerSystemWrapper{system: sys.musicTriggerSystem})
	game.World.AddSystem(&positionalAudioSystemWrapper{system: sys.positionalAudioSystem})
	game.World.AddSystem(&reverbSystemWrapper{system: sys.reverbSystem})

	// Phase 14: Quality System
	game.World.AddSystem(&qualitySystemWrapper{system: sys.qualitySystem})

	// Phase 33-36: Trade, Terrain, and Merchant Systems
	game.World.AddSystem(&tradeSystemWrapper{system: sys.tradeSystem})
	game.World.AddSystem(&terrainConstructionSystemWrapper{system: sys.terrainConstructionSys})
	game.World.AddSystem(&terrainModificationSystemWrapper{system: sys.terrainModificationSys})
	game.World.AddSystem(&merchantCaravanSystemWrapper{system: sys.merchantCaravanSystem})

	// INTEGRATION FIX [Category A]: High-Level Management System Registrations
	// Gap: CompanionSystem, VehicleSystem, AdaptiveSoundtrackSystem initialized but never registered
	// Fix: Registered high-level wrapper systems for unified management
	// Roadmap: ROADMAP_V4.md Phase 21.2, 22.2, 29

	// Phase 21.2: High-level VehicleSystem wrapper
	game.World.AddSystem(&vehicleSystemWrapper{system: sys.vehicleSystem})

	// Phase 22.2: High-level CompanionSystem wrapper
	game.World.AddSystem(&companionSystemWrapper{system: sys.companionSystem})

	// Phase 29: AdaptiveSoundtrackSystem for dynamic music
	game.World.AddSystem(&adaptiveSoundtrackSystemWrapper{system: sys.adaptiveSoundtrackSystem})

	// INTEGRATION FIX [Category A]: V8.0 Fluid Simulator System Registration (Phase 50.4)
	// Gap: FluidSimulator has Update() method and should be registered for fluid dynamics
	// Fix: Added fluid simulator to World update loop for water flow simulation
	// Roadmap: ROADMAP_V8.md Phase 50.4
	// Note: Other V8 managers (EnhancedVehicleSystem, SwimmingManager, FloodingManager) are
	// helper utilities used by other systems, not standalone systems requiring registration
	if sys.fluidSimulator != nil {
		game.World.AddSystem(&fluidSimulatorWrapper{system: sys.fluidSimulator})
	}
}

// configureSystemConnections wires up interdependent systems.
func configureSystemConnections(game *engine.EbitenGame, sys *systemsContainer) {
	sys.combatSystem.SetCamera(game.CameraSystem)
	sys.combatSystem.SetParticleSystem(sys.particleSystem, game.World, *genreID)
	sys.combatSystem.SetProjectileSystem(sys.projectileSystem)
	// Plan Phase 1.1: Connect combat system to audio manager for combat SFX
	sys.combatSystem.SetAudioManager(sys.audioManager)
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
			"theme":  getGenreTheme(),
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
		clientLogger.WithFields(logrus.Fields{
			"transitions":    true,
			"parallax":       false,
			"enhanced_walls": true,
		}).Info("initializing terrain rendering system")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(tileSize, tileSize, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)

	// Configure advanced tile rendering features (unconditionally enabled as of Phase 2.1)
	terrainRenderSystem.SetTransitionsEnabled(true)
	terrainRenderSystem.SetParallaxEnabled(false) // Disabled due to performance impact
	terrainRenderSystem.SetEnhancedWallsEnabled(true)

	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		clientLogger.Info("terrain rendering system initialized with advanced features")
	}
}

// configureLightingSystem enables and configures dynamic lighting (unconditionally enabled as of Phase 2.1).
func configureLightingSystem(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	clientLogger.Info("enabling dynamic lighting system")
	game.EnableLighting(true)
	game.SetLightingGenrePreset(*genreID)

	clientLogger.WithFields(logrus.Fields{
		"genre":     *genreID,
		"enabled":   true,
		"maxLights": maxLights,
	}).Info("lighting system configured")
}

// configurePostProcessing initializes and configures the post-processing system (unconditionally enabled as of Phase 2.1).
func configurePostProcessing(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if game.PostProcessor == nil {
		clientLogger.Warn("PostProcessor not initialized")
		return
	}

	// Enable post-processing unconditionally (Phase 2.1)
	game.PostProcessor.SetEnabled(true)

	// Apply preset if specified
	if *postprocessPreset != "" {
		game.PostProcessor.SetGenrePreset(*postprocessPreset)
		clientLogger.WithFields(logrus.Fields{
			"preset": *postprocessPreset,
		}).Info("post-processing preset applied")
		return
	}

	// Otherwise, configure individual effects
	if *postprocessColorGrading {
		game.PostProcessor.EnableColorGrading(
			*postprocessSaturation,
			*postprocessContrast,
			*postprocessBrightness,
			0.0, // temperature (not exposed via flags yet)
			0.0, // tint (not exposed via flags yet)
		)
		clientLogger.WithFields(logrus.Fields{
			"saturation": *postprocessSaturation,
			"contrast":   *postprocessContrast,
			"brightness": *postprocessBrightness,
		}).Debug("color grading enabled")
	}

	if *postprocessVignette {
		game.PostProcessor.EnableVignette(
			*postprocessVignetteIntens,
			*postprocessVignetteSoft,
		)
		clientLogger.WithFields(logrus.Fields{
			"intensity": *postprocessVignetteIntens,
			"softness":  *postprocessVignetteSoft,
		}).Debug("vignette enabled")
	}

	if *postprocessChromaticAber {
		game.PostProcessor.EnableChromaticAberration(
			*postprocessChromaticIntens,
			1.0, // directionX (outward from center)
			0.0, // directionY
			3,   // samples
		)
		clientLogger.WithFields(logrus.Fields{
			"intensity": *postprocessChromaticIntens,
		}).Debug("chromatic aberration enabled")
	}

	clientLogger.Info("post-processing configured")
}

// configurePaletteOptions sets up genre palette options for sprite generation (Phase 5.4).
func configurePaletteOptions(sys *systemsContainer, clientLogger *logrus.Entry) {
	if sys.animationSystem == nil {
		clientLogger.Warn("AnimationSystem not initialized, skipping palette configuration")
		return
	}

	// Parse palette options from command-line flags
	opts, err := parsePaletteOptions()
	if err != nil {
		clientLogger.WithError(err).Warn("failed to parse palette options, using defaults")
		return
	}

	// Set palette options on animation system (used for sprite generation)
	sys.animationSystem.SetPaletteOptions(opts)

	clientLogger.WithFields(logrus.Fields{
		"harmony": paletteHarmony,
		"mood":    paletteMood,
		"rarity":  paletteRarity,
	}).Info("palette options configured")
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
func spawnWorldEntities(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, sys *systemsContainer, clientLogger *logrus.Entry) {
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"theme": getGenreTheme(),
		},
	}

	spawnEnemiesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnMerchantsWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnMinigamesWithLogging(game.World, sys.minigameGenerator, params, clientLogger)
	spawnStationsWithLogging(game.World, generatedTerrain, clientLogger)
	spawnPuzzlesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnObjectsWithLogging(game.World, generatedTerrain, clientLogger)
	spawnVehiclesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnCompanionsWithLogging(game.World, generatedTerrain, params, clientLogger, sys.companionLearningSys)
	spawnBookshelvesWithLogging(game.World, generatedTerrain, params, clientLogger)
	spawnStoryFragmentsWithLogging(game.World, generatedTerrain, params, clientLogger)
	generateNarrativeArcWithLogging(game.World, sys.narrativeGenerator, params, clientLogger)
}

// spawnEnemiesWithLogging spawns enemies in terrain with optional verbose logging.
func spawnEnemiesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning enemies in dungeon rooms")
	}
	enemyCount, err := engine.SpawnEnemiesInTerrain(w, generatedTerrain, *seed, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn enemies")
	} else if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"enemyCount": enemyCount,
			"roomCount":  len(generatedTerrain.Rooms) - 1,
		}).Info("spawned enemies")
	}
}

// spawnMerchantsWithLogging spawns merchants in terrain with optional verbose logging.
func spawnMerchantsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning merchants in dungeon")
	}
	merchantCount, err := engine.SpawnMerchantsInTerrain(w, generatedTerrain, *seed, params, defaultMerchantCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn merchants")
	} else if *verbose {
		clientLogger.WithField("merchantCount", merchantCount).Info("spawned merchants")
	}
}

// spawnMinigamesWithLogging adds procedural minigames to merchants with optional verbose logging.
// Phase 3.5 (PLAN.md): Minigames in shops/taverns for player entertainment
func spawnMinigamesWithLogging(w *engine.World, minigameGen *minigame.Generator, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if minigameGen == nil {
		return
	}
	if *verbose {
		clientLogger.Info("adding minigames to merchants")
	}
	minigameCount := addMinigamesToMerchants(w, minigameGen, *seed, params, clientLogger)
	if *verbose {
		clientLogger.WithField("minigameCount", minigameCount).Info("added minigames to merchants")
	}
}

// spawnStationsWithLogging spawns crafting stations in terrain with optional verbose logging.
func spawnStationsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning crafting stations in dungeon")
	}
	stationGen := station.NewStationGenerator()
	stationCount := engine.SpawnStationsInTerrain(w, stationGen, generatedTerrain, tileSize, *seed+seedOffsetStation, *genreID)
	if *verbose {
		clientLogger.WithField("stationCount", stationCount).Info("spawned crafting stations")
	}
}

// spawnPuzzlesWithLogging spawns procedural puzzles in terrain with optional verbose logging.
func spawnPuzzlesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning procedural puzzles in dungeon")
	}
	puzzleCount, err := engine.SpawnPuzzlesInTerrain(w, generatedTerrain, *seed+seedOffsetPuzzle, params, defaultPuzzleCount)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn puzzles")
	} else {
		clientLogger.WithFields(logrus.Fields{
			"puzzleCount": puzzleCount,
			"targetCount": defaultPuzzleCount,
			"roomCount":   len(generatedTerrain.Rooms) - 1,
		}).Info("spawned procedural puzzles")
	}
}

// spawnObjectsWithLogging spawns destructible objects in terrain with logging.
func spawnObjectsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning destructible objects in dungeon")
	}
	objectCount := spawnDestructibleObjects(w, generatedTerrain, *seed+seedOffsetObject, *genreID, clientLogger.Logger)
	clientLogger.WithFields(logrus.Fields{
		"objectCount": objectCount,
		"roomCount":   len(generatedTerrain.Rooms) - 1,
		"genre":       *genreID,
	}).Info("spawned destructible objects")
}

// spawnVehiclesWithLogging spawns vehicles in terrain with optional verbose logging.
func spawnVehiclesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning vehicles in dungeon")
	}
	vehicleCount, err := spawnVehicles(w, generatedTerrain, *seed+seedOffsetVehicle, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn vehicles")
	} else if *verbose {
		clientLogger.WithField("vehicleCount", vehicleCount).Info("spawned vehicles")
	}
}

// spawnCompanionsWithLogging spawns companions in terrain with optional verbose logging.
func spawnCompanionsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry, companionLearningSys *engine.CompanionLearningSystem) {
	if *verbose {
		clientLogger.Info("spawning companions in dungeon")
	}
	companionCount, err := spawnCompanions(w, generatedTerrain, *seed+seedOffsetCompanion, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn companions")
	} else if *verbose {
		clientLogger.WithField("companionCount", companionCount).Info("spawned companions")
	}

	// Initialize learning for spawned companions (Phase 4.1)
	if companionLearningSys != nil {
		companionEntities := w.GetEntitiesWith("companion")
		for _, entity := range companionEntities {
			_, hasLearning := entity.GetComponent("companion_learning")
			if !hasLearning {
				learningRate := 1.0 + (float64(companionCount) * 0.1)
				if learningRate > 2.0 {
					learningRate = 2.0
				}
				err := companionLearningSys.AddCompanionLearning(entity.ID, learningRate)
				if err != nil {
					clientLogger.WithError(err).WithField("companionID", entity.ID).Warn("failed to initialize companion learning")
				} else if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"companionID":  entity.ID,
						"learningRate": learningRate,
					}).Debug("initialized companion learning")
				}
			}
		}
	}
}

// spawnBookshelvesWithLogging spawns bookshelves with books in terrain with optional verbose logging.
func spawnBookshelvesWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning bookshelves in dungeon")
	}
	bookshelfCount, err := spawnBookshelves(w, generatedTerrain, *seed+seedOffsetBook, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn bookshelves")
	} else if *verbose {
		clientLogger.WithField("bookshelfCount", bookshelfCount).Info("spawned bookshelves")
	}
}

// spawnStoryFragmentsWithLogging spawns story fragments in terrain with optional verbose logging.
func spawnStoryFragmentsWithLogging(w *engine.World, generatedTerrain *terrain.Terrain, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning story fragments in dungeon")
	}
	fragmentCount, err := spawnStoryFragments(w, generatedTerrain, *seed+seedOffsetStory, params, clientLogger)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to spawn story fragments")
	} else if *verbose {
		clientLogger.WithField("fragmentCount", fragmentCount).Info("spawned story fragments")
	}
}

// spawnEnvironmentalEffects spawns lights, weather effects, and environmental hazards (unconditionally enabled).
func spawnEnvironmentalEffects(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("spawning environmental lights in dungeon")
	}
	lightCount := spawnEnvironmentalLights(game.World, generatedTerrain, *seed+seedOffsetLight, *genreID)
	clientLogger.WithFields(logrus.Fields{
		"lightCount": lightCount,
		"genre":      *genreID,
	}).Info("spawned environmental lights")

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

	// Phase 3.4: Spawn environmental hazards (fire pits, acid pools, spike traps, etc.)
	if *verbose {
		clientLogger.Info("spawning environmental hazards")
	}
	hazardCount := spawnEnvironmentalHazards(game.World, generatedTerrain, *seed+seedOffsetEnvironment, *genreID, clientLogger)
	clientLogger.WithFields(logrus.Fields{
		"hazardCount": hazardCount,
		"genre":       *genreID,
	}).Info("spawned environmental hazards")
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

	// Add player torch (unconditionally enabled as of Phase 2.1)
	playerTorch := engine.NewTorchLight(200)
	playerTorch.Enabled = true
	player.AddComponent(playerTorch)

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"radius":    200,
			"intensity": playerTorch.Intensity,
		}).Info("player torch added")
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

	// BUG FIX: Phase 7 - SaveManager auto-selects storage backend based on platform
	// - Desktop/mobile: file-based storage in ./saves directory
	// - WASM: localStorage with 5MB limit, fallback to in-memory
	// Platform: All (conditional compilation selects correct implementation)
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

	connectUIComponentsToInputSystem(game, inputSystem, shopUI, player, clientLogger)

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

// connectUIComponentsToInputSystem connects all UI components to input system for ESC key handling.
// BUG FIX: Phase 3 - Menu Trap - Enables dual-exit pattern (toggle key + ESC key close) for ALL UI panels.
func connectUIComponentsToInputSystem(game *engine.EbitenGame, inputSystem *engine.InputSystem, shopUI *engine.ShopUI, player *engine.Entity, clientLogger *logrus.Entry) {
	connectBasicUIComponents(game, inputSystem, shopUI)
	connectAdvancedUIComponents(game, inputSystem)
	connectDialogUI(game, inputSystem, player, clientLogger)
}

func connectBasicUIComponents(game *engine.EbitenGame, inputSystem *engine.InputSystem, shopUI *engine.ShopUI) {
	if game.MailboxUI != nil {
		inputSystem.SetMailboxUI(game.MailboxUI)
	}
	if game.InventoryUI != nil {
		inputSystem.SetInventoryUI(game.InventoryUI)
	}
	if game.CharacterUI != nil {
		inputSystem.SetCharacterUI(game.CharacterUI)
	}
	if game.SkillsUI != nil {
		inputSystem.SetSkillsUI(game.SkillsUI)
	}
	if game.QuestUI != nil {
		inputSystem.SetQuestUI(game.QuestUI)
	}
	if game.MapUI != nil {
		inputSystem.SetMapUI(game.MapUI)
	}
	if game.CraftingUI != nil {
		inputSystem.SetCraftingUI(game.CraftingUI)
	}
	if shopUI != nil {
		inputSystem.SetShopUI(shopUI)
	}
	if game.TradeUI != nil {
		inputSystem.SetTradeUI(game.TradeUI)
	}
}

func connectAdvancedUIComponents(game *engine.EbitenGame, inputSystem *engine.InputSystem) {
	if game.AdvancedClassUI != nil {
		inputSystem.SetAdvancedClassUI(game.AdvancedClassUI)
		inputSystem.SetClassesCallback(func() {
			game.AdvancedClassUI.Toggle()
		})
	}
	if game.TerritoryUI != nil {
		inputSystem.SetTerritoryUI(game.TerritoryUI)
		inputSystem.SetTerritoryCallback(func() {
			game.TerritoryUI.Toggle()
		})
	}
}

func connectDialogUI(game *engine.EbitenGame, inputSystem *engine.InputSystem, player *engine.Entity, clientLogger *logrus.Entry) {
	if game.DialogUI == nil {
		return
	}

	inputSystem.SetDialogUI(game.DialogUI)
	inputSystem.SetDialogCallback(func() {
		nearestNPC := findNearestNPC(game, player)
		if nearestNPC != nil {
			if err := game.DialogUI.Show(nearestNPC.ID); err != nil {
				clientLogger.WithError(err).Warn("failed to open dialog UI")
			}
		}
	})
}

func findNearestNPC(game *engine.EbitenGame, player *engine.Entity) *engine.Entity {
	if player == nil {
		return nil
	}

	posComp, ok := player.GetComponent("position")
	if !ok {
		return nil
	}
	pos := posComp.(*engine.PositionComponent)

	var nearestNPC *engine.Entity
	minDist := 5.0 * 32.0
	minDistSquared := minDist * minDist

	for _, entity := range game.World.GetEntities() {
		if entity.ID == player.ID {
			continue
		}
		if _, hasDialog := entity.GetComponent("dialog"); !hasDialog {
			continue
		}

		if ePos, ok := entity.GetComponent("position"); ok {
			entityPos := ePos.(*engine.PositionComponent)
			dx := entityPos.X - pos.X
			dy := entityPos.Y - pos.Y
			distSquared := dx*dx + dy*dy

			if distSquared < minDistSquared {
				minDistSquared = distSquared
				nearestNPC = entity
			}
		}
	}

	return nearestNPC
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

// initializeUIIntegration sets up shop UI, crafting UI, mailbox UI, housing UI, gallery UI, and connects them to game systems.
func initializeUIIntegration(game *engine.EbitenGame, player *engine.Entity, commerceSystem *engine.CommerceSystem, dialogSystem *engine.DialogSystem, craftingSystem *engine.CraftingSystem, inventorySystem *engine.InventorySystem, sys *systemsContainer, clientLogger *logrus.Entry) (*engine.ShopUI, *engine.CraftingUI) {
	shopUI := initializeShopAndCommerceUI(game, player, commerceSystem, dialogSystem, clientLogger)
	initializeHousingAndGalleryUI(game, player, sys, clientLogger)
	initializeGuildAndTradeUI(game, player, sys, clientLogger)
	initializeAdvancedAndTerritoryUI(game, player, sys, clientLogger)
	initializeStoryAndDialogUI(game, player, sys, clientLogger)
	craftingUI := initializeCraftingAndMailboxUI(game, player, craftingSystem, inventorySystem, clientLogger)

	return shopUI, craftingUI
}

func initializeShopAndCommerceUI(game *engine.EbitenGame, player *engine.Entity, commerceSystem *engine.CommerceSystem, dialogSystem *engine.DialogSystem, clientLogger *logrus.Entry) *engine.ShopUI {
	shopUI := engine.NewShopUI(*width, *height)
	shopUI.SetPlayerEntity(player)
	shopUI.SetCommerceSystem(commerceSystem)
	shopUI.SetDialogSystem(dialogSystem)
	game.ShopUI = shopUI

	if *verbose {
		clientLogger.Info("shop UI initialized and connected to commerce/dialog systems")
	}

	return shopUI
}

func initializeHousingAndGalleryUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	housingUI := housing.NewHousingUI(*width, *height)
	if sys.housingManager != nil {
		housingUI.SetManagers(sys.housingManager, sys.guildHallManager, sys.buildingGenerator, sys.furnitureGenerator)
		housingUI.SetPlayerID(player.ID)
	}
	game.HousingUI = housingUI

	if *verbose {
		clientLogger.Info("housing UI initialized (H key to open)")
	}

	galleryUI := engine.NewGalleryUI(*width, *height)
	if sys.imageGallery != nil {
		galleryUI.SetGallery(sys.imageGallery)
	}
	game.GalleryUI = galleryUI

	if *verbose {
		clientLogger.Info("gallery UI initialized (G key to open)")
	}
}

func initializeGuildAndTradeUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	if sys.guildUI != nil {
		game.GuildUI = sys.guildUI
		if *verbose {
			clientLogger.Info("guild UI initialized (U key to open)")
		}
	}

	tradeUI := engine.NewTradeUI(game.World, sys.tradeSystem, *width, *height)
	tradeUI.SetPlayerEntity(player)
	game.TradeUI = tradeUI

	if *verbose {
		clientLogger.Info("trade UI initialized (T key to open)")
	}
}

func initializeAdvancedAndTerritoryUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	advancedClassUI := engine.NewAdvancedClassUI(game.World, sys.advancedClassSystem, *width, *height)
	advancedClassUI.SetPlayerEntity(player)
	game.AdvancedClassUI = advancedClassUI

	if *verbose {
		clientLogger.Info("advanced class UI initialized (A key to open)")
	}

	sys.territoryUI.SetPlayerEntity(player)
	game.TerritoryUI = sys.territoryUI

	if *verbose {
		clientLogger.Info("territory UI initialized (Y key to open)")
	}
}

func initializeStoryAndDialogUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	storyChoiceUI := engine.NewStoryChoiceUI(game.World, sys.branchingNarrativeSystem, *width, *height)
	storyChoiceUI.SetPlayerEntity(player)
	game.StoryChoiceUI = storyChoiceUI

	if *verbose {
		clientLogger.Info("story choice UI initialized (shows automatically when choices available)")
	}

	initializePlayerNarrative(player, sys.branchingNarrativeSystem, *seed, *genreID, clientLogger)

	dialogUI := engine.NewDialogUI(game.World, sys.dialogSystem, sys.npcDialogSystem, *width, *height)
	dialogUI.SetPlayerEntity(player)
	game.DialogUI = dialogUI

	if *verbose {
		clientLogger.Info("dialog UI initialized (D key to toggle with NPCs)")
	}
}

func initializeCraftingAndMailboxUI(game *engine.EbitenGame, player *engine.Entity, craftingSystem *engine.CraftingSystem, inventorySystem *engine.InventorySystem, clientLogger *logrus.Entry) *engine.CraftingUI {
	craftingUI := engine.NewCraftingUI(*width, *height)
	craftingUI.SetPlayerEntity(player)
	craftingUI.SetCraftingSystem(craftingSystem)
	game.CraftingUI = craftingUI

	if *verbose {
		clientLogger.Info("crafting UI initialized and connected to crafting system")
	}

	mailboxUI := engine.NewMailboxUI(0, 0, *width, *height, *genreID)
	game.MailboxUI = mailboxUI

	if *verbose {
		clientLogger.Info("mailbox UI initialized (Phase 40.3)")
	}

	game.SetInventorySystem(inventorySystem)

	return craftingUI
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

// initializePlayerAdvancedClass initializes the advanced class system for the player.
// Phase 4.2 (PLAN.md): Multi-classing, prestige classes, and talent trees
func initializePlayerAdvancedClass(player *engine.Entity, game *engine.EbitenGame, clientLogger *logrus.Entry) {
	charData := game.GetPendingCharacterData()
	if charData == nil {
		return
	}

	// Map CharacterClass to advanced.ClassID
	classID := mapCharacterClassToAdvancedClass(charData.Class)
	if classID == "" {
		clientLogger.Warn("character class not mapped to advanced class system")
		return
	}

	// Initialize with level 1 and warrior class (can be customized via character creation)
	// Get the advanced class system from game systems
	// Note: We can't access sys.advancedClassSystem here directly,
	// so we initialize the component manually and the system will pick it up
	player.AddComponent(&advanced.AdvancedClassComponent{
		PrimaryClass: classID,
		Level:        1,
		TalentPoints: advanced.TalentAllocation{
			Talents:     make(map[advanced.TalentID]int),
			PointsTotal: 1,
		},
	})

	clientLogger.WithFields(logrus.Fields{
		"class": classID,
		"level": 1,
	}).Info("initialized advanced class system")
}

// mapCharacterClassToAdvancedClass maps engine.CharacterClass to advanced.ClassID
func mapCharacterClassToAdvancedClass(class engine.CharacterClass) advanced.ClassID {
	switch class {
	case engine.ClassWarrior:
		return advanced.ClassWarrior
	case engine.ClassMage:
		return advanced.ClassMage
	case engine.ClassRogue:
		return advanced.ClassRogue
	case engine.ClassRanger:
		return advanced.ClassRanger
	case engine.ClassCleric:
		return advanced.ClassCleric
	case engine.ClassNecromancer:
		return advanced.ClassNecromancer
	case engine.ClassBattlemage:
		return advanced.ClassWarrior // Default to warrior for hybrids (can be improved)
	default:
		return ""
	}
}

// finalizeGameInitialization processes initial entity updates and logs game start info.
func finalizeGameInitialization(game *engine.EbitenGame, player *engine.Entity, networkClient interface{}, clientLogger *logrus.Entry) {
	game.World.Update(0)

	clientLogger.Info("game initialized successfully")
	clientLogger.Info("controls: WASD to move, Space to attack, E to use item, I: Inventory, J: Quests, L: Mailbox, A: Classes")
	clientLogger.WithFields(logrus.Fields{"genre": *genreID, "seed": *seed, "server": *server}).Info("game settings")
}

// handleHostAndPlay starts embedded server if host-and-play mode is enabled.
// Returns cleanup function that should be deferred by caller.
func handleHostAndPlay(logger *logrus.Logger, clientLogger *logrus.Entry) func() {
	if !*hostAndPlay {
		return func() {} // Return no-op cleanup
	}

	// Log message depends on whether this was explicitly requested or auto-enabled
	// Auto-enabled case already logged in main.go, so we skip redundant logging here
	if !autoEnabledHostAndPlay {
		if *hostLAN {
			clientLogger.Info("host-and-play mode - starting LAN-accessible server on 0.0.0.0")
		} else {
			clientLogger.Info("host-and-play mode - starting localhost server on 127.0.0.1")
		}
	}

	serverAddr, cleanup, err := startEmbeddedServer(logger, *seed, *genreID)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to start embedded server")
	}

	*server = serverAddr
	*multiplayer = true

	clientLogger.WithField("serverAddr", serverAddr).Info("embedded server started, connecting client")

	return cleanup // Return cleanup to be deferred by caller
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
		sys.recipeGen, sys.magicGenerator, sys.skillGenerator, *seed, *genreID, logger,
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
			"theme":  getGenreTheme(),
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

// INTEGRATION FIX [Category A]: V8.0 Fluid Simulator System Wrapper (Phase 50.4)
// Gap: FluidSimulator needed wrapper to adapt to World.System interface
// Fix: Added wrapper for fluid dynamics simulation system
// Roadmap: ROADMAP_V8.md Phase 50.4
// Note: EnhancedVehicleSystem, SwimmingManager, and FloodingManager are helper managers
// used by vehicle and entity systems directly, not standalone systems requiring wrappers

// fluidSimulatorWrapper adapts Simulator (fluid dynamics) to the System interface.
type fluidSimulatorWrapper struct {
	system *fluids.Simulator
}

func (w *fluidSimulatorWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// Phase 1.2: Destruction Physics System Wrapper (PLAN.md)
// destructionSystemWrapper adapts destruction.System to the World.System interface.
type destructionSystemWrapper struct {
	system *destruction.System
}

func (w *destructionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// territorySystemWrapper adapts TerritorySystem to the System interface.
type territorySystemWrapper struct {
	system *engine.TerritorySystem
}

func (w *territorySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

// runGameLoop starts the main game loop.
func runGameLoop(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	windowTitle := fmt.Sprintf("Venture %s - Procedural Action RPG", version.Version)
	if err := game.Run(windowTitle); err != nil {
		clientLogger.WithError(err).Fatal("error running game")
	}
}

// generateNarrativeArcWithLogging generates a procedural story arc for the world narrative with optional verbose logging.
// Phase 3.6 (PLAN.md): Integrates pkg/procgen/narrative to create overarching world story
func generateNarrativeArcWithLogging(w *engine.World, narrativeGen *narrative.StoryArcGenerator, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	if narrativeGen == nil {
		clientLogger.Warn("narrative generator is nil, skipping story arc generation")
		return
	}

	if *verbose {
		clientLogger.Info("generating procedural narrative arc for world")
	}

	// Generate story arc using world seed
	result, err := narrativeGen.Generate(*seed+seedOffsetNarrative, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to generate narrative arc")
		return
	}

	// Validate the generated arc
	arc, ok := result.(*narrative.StoryArc)
	if !ok {
		clientLogger.Warn("narrative generator returned invalid type")
		return
	}

	if err := narrativeGen.Validate(arc); err != nil {
		clientLogger.WithError(err).Warn("generated narrative arc failed validation")
		return
	}

	// Find or create world narrative entity
	worldNarrativeEntity := findOrCreateWorldNarrativeEntity(w)
	if worldNarrativeEntity == nil {
		clientLogger.Warn("failed to create world narrative entity")
		return
	}

	// Get or create narrative component
	var narrativeComp *engine.NarrativeComponent
	if comp, ok := worldNarrativeEntity.GetComponent("narrative"); ok {
		narrativeComp, _ = comp.(*engine.NarrativeComponent)
	}
	if narrativeComp == nil {
		narrativeComp = &engine.NarrativeComponent{
			CurrentAct:      engine.ActSetup,
			EventHistory:    make([]engine.NarrativeEvent, 0),
			MainObjective:   arc.MainConflict,
			StoryProgress:   0.0,
			ActiveThreads:   make([]string, 0),
			ResolvedThreads: make([]string, 0),
			WorldStateFlags: make(map[string]bool),
			Relationships:   make(map[string]float64),
			PlayerDecisions: make([]engine.PlayerDecision, 0),
		}
		worldNarrativeEntity.AddComponent(narrativeComp)
	}

	// Update narrative component with story arc data
	narrativeComp.MainObjective = arc.MainConflict

	// Add initial narrative event for the story arc
	initialEvent := engine.NarrativeEvent{
		Type:        engine.EventDiscovery,
		Timestamp:   time.Now(),
		Description: fmt.Sprintf("The story begins: %s", arc.Title),
		Importance:  1.0,
		Act:         engine.ActSetup,
	}
	narrativeComp.AddEvent(initialEvent)

	// Add plot points as active threads
	for _, plotPoint := range arc.PlotPoints {
		if plotPoint.Act == 1 {
			narrativeComp.ActiveThreads = append(narrativeComp.ActiveThreads, plotPoint.Description)
		}
	}

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"arc_title":   arc.Title,
			"antagonist":  arc.Antagonist,
			"ally":        arc.Ally,
			"plot_points": len(arc.PlotPoints),
			"endings":     len(arc.PossibleEndings),
			"difficulty":  arc.Difficulty,
			"seed":        arc.Seed,
		}).Info("procedural narrative arc generated and integrated")
	}
}

// findOrCreateWorldNarrativeEntity finds existing world narrative entity or creates a new one.
func findOrCreateWorldNarrativeEntity(w *engine.World) *engine.Entity {
	// Search for existing world narrative entity
	entities := w.GetEntities()
	for _, entity := range entities {
		if _, ok := entity.GetComponent("narrative"); ok {
			// Check if this is a world narrative entity (not player narrative)
			if _, hasPlayer := entity.GetComponent("player"); !hasPlayer {
				return entity
			}
		}
	}

	// Create new world narrative entity
	worldNarrative := w.CreateEntity()
	return worldNarrative
}

// initializePlayerNarrative initializes branching narrative for the player.
// Phase 6.1 (PLAN.md): Creates a procedural story arc and starts player on narrative journey.
func initializePlayerNarrative(player *engine.Entity, narrativeSystem *engine.BranchingNarrativeSystem, seed int64, genreID string, logger *logrus.Entry) {
	// Add branching narrative component to player
	player.AddComponent(&engine.BranchingNarrativeComponent{
		LastUpdate: 0,
	})

	// Create narrative generator and manager
	generator := branching.NewGenerator()
	manager := branching.NewManager()

	// Generate a story arc for the player
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      0,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"theme": getGenreTheme(),
		},
	}

	result, err := generator.Generate(seed, params)
	if err != nil {
		logger.WithError(err).Warn("failed to generate branching narrative arc")
		return
	}

	arc, ok := result.(*branching.StoryArc)
	if !ok {
		logger.Warn("branching narrative generator returned invalid type")
		return
	}

	// Start the story arc
	if err := narrativeSystem.StartStoryArc(player, arc, manager); err != nil {
		logger.WithError(err).Warn("failed to start branching narrative arc")
		return
	}

	logger.WithFields(logrus.Fields{
		"arc_id":     arc.ID,
		"arc_title":  arc.Title,
		"node_count": len(arc.Nodes),
	}).Info("player branching narrative initialized")
}
