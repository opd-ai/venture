// Package engine provides shared system initialization for all platforms.
// This file contains the centralized system initialization logic used by
// desktop, WASM, and mobile platforms to ensure Version 2.0 feature parity.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// SystemInitConfig contains configuration parameters for system initialization.
// This struct is passed to InitializeGameSystems to configure system behavior.
type SystemInitConfig struct {
	// Core configuration
	Seed    int64  // World seed for deterministic generation
	GenreID string // Genre identifier (fantasy, scifi, horror, cyberpunk, postapoc)
	Logger  *logrus.Logger

	// System-specific settings
	MaxSpeed          float64 // MovementSystem max speed (default: 200.0)
	CollisionCellSize float64 // CollisionSystem spatial grid cell size (default: 64.0)
	SampleRate        int     // AudioManager sample rate (default: 44100)
	TileSize          int     // Tile size for terrain systems (default: 32)

	// Feature flags
	EnableVerboseLogging bool // Enable detailed system initialization logging
}

// DefaultSystemInitConfig returns a configuration with sensible defaults.
func DefaultSystemInitConfig(seed int64, genreID string, logger *logrus.Logger) *SystemInitConfig {
	return &SystemInitConfig{
		Seed:                 seed,
		GenreID:              genreID,
		Logger:               logger,
		MaxSpeed:             200.0,
		CollisionCellSize:    64.0,
		SampleRate:           44100,
		TileSize:             32,
		EnableVerboseLogging: false,
	}
}

// SystemInitResult contains references to initialized systems that may need
// further configuration after initialization (e.g., setting callbacks).
type SystemInitResult struct {
	// Systems that often need post-initialization configuration
	InputSystem       *InputSystem
	CombatSystem      *CombatSystem
	CollisionSystem   *CollisionSystem
	ProjectileSystem  *ProjectileSystem
	AudioManager      *AudioManager
	ObjectiveTracker  *ObjectiveTrackerSystem
	CommerceSystem    *CommerceSystem
	DialogSystem      *DialogSystem
	CraftingSystem    *CraftingSystem
	InteractionSystem *InteractionSystem
	AnimationSystem   *AnimationSystem
	ParticleSystem    *ParticleSystem
	TutorialSystem    *EbitenTutorialSystem
	HelpSystem        *EbitenHelpSystem

	// System wrappers
	AnimationSystemWrapper System
	RotationSystemWrapper  System
	SquadSystemWrapper     System
}

// InitializeGameSystems initializes all game systems for Version 2.0 feature parity.
// This function registers 43 systems in the correct dependency order.
// The 44th system (SpatialPartitionSystem) must be initialized separately after
// terrain generation using InitializeSpatialPartitionSystem().
//
// Returns: SystemInitResult containing references to systems that may need
// further configuration (callbacks, connections, etc.).
//
// Note: This function only registers systems. Additional setup like terrain
// generation, entity spawning, and callback wiring must be done separately.
func InitializeGameSystems(game *EbitenGame, config *SystemInitConfig) (*SystemInitResult, error) {
	if game == nil {
		return nil, fmt.Errorf("game instance is nil")
	}
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	logger := config.Logger
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	result := &SystemInitResult{}

	if config.EnableVerboseLogging {
		logger.Info("initializing game systems (44 total)")
	}

	// ========================================================================
	// CORE GAMEPLAY SYSTEMS (13 systems)
	// ========================================================================

	// 1. InputSystem - captures player actions
	result.InputSystem = NewInputSystem()
	game.World.AddSystem(result.InputSystem)

	// 2. CameraSystem - processes screen shake, hit-stop, accessibility settings
	game.World.AddSystem(game.CameraSystem)

	// 3. RotationSystem - updates entity facing direction based on aim (Phase 10.1)
	rotationSystem := NewRotationSystem(game.World)
	result.RotationSystemWrapper = &rotationSystemWrapper{system: rotationSystem}
	game.World.AddSystem(result.RotationSystemWrapper)

	// 4-6. Player input processing systems
	result.CombatSystem = NewCombatSystemWithLogger(config.Seed, logger)
	playerCombatSystem := NewPlayerCombatSystem(result.CombatSystem, game.World)
	game.World.AddSystem(playerCombatSystem)

	inventorySystem := NewInventorySystem(game.World)
	playerItemUseSystem := NewPlayerItemUseSystem(inventorySystem, game.World)
	game.World.AddSystem(playerItemUseSystem)

	statusEffectRNG := rand.New(rand.NewSource(config.Seed + 999))
	statusEffectSystem := NewStatusEffectSystem(game.World, statusEffectRNG)
	spellCastingSystem := NewSpellCastingSystem(game.World, statusEffectSystem)
	playerSpellCastingSystem := NewPlayerSpellCastingSystem(spellCastingSystem, game.World)
	game.World.AddSystem(playerSpellCastingSystem)

	// 7-8. Physics systems
	movementSystem := NewMovementSystem(config.MaxSpeed)
	result.CollisionSystem = NewCollisionSystem(config.CollisionCellSize)
	movementSystem.SetCollisionSystem(result.CollisionSystem)
	game.World.AddSystem(movementSystem)
	game.World.AddSystem(result.CollisionSystem)

	// 9. ProjectileSystem - ranged weapon physics (Phase 10.2)
	result.ProjectileSystem = NewProjectileSystem(game.World)
	result.ProjectileSystem.SetCamera(game.CameraSystem)
	result.ProjectileSystem.SetGenre(config.GenreID)
	result.ProjectileSystem.SetSeed(config.Seed)
	game.World.AddSystem(result.ProjectileSystem)

	// 10-11. Combat and status effects
	game.World.AddSystem(result.CombatSystem)
	game.World.AddSystem(statusEffectSystem)

	// 12. RevivalSystem - multiplayer death mechanics
	revivalSystem := NewRevivalSystem(game.World)
	game.World.AddSystem(revivalSystem)

	// 13. AISystem - enemy decision-making
	aiSystem := NewAISystem(game.World)
	game.World.AddSystem(aiSystem)

	// ========================================================================
	// PHASE 13: ADVANCED AI (3 systems)
	// ========================================================================

	// 14. BehaviorTreeSystem - executes behavior trees (Phase 13.1)
	behaviorTreeSystem := NewBehaviorTreeSystem(game.World)
	game.World.AddSystem(behaviorTreeSystem)

	// 15. SquadSystem - coordinated enemy tactics (Phase 13.2)
	squadSystem := NewSquadSystem(game.World)
	result.SquadSystemWrapper = &squadSystemWrapper{system: squadSystem}
	game.World.AddSystem(result.SquadSystemWrapper)

	// 16. FactionSystem - reputation and relationships (Phase 13.3)
	factionSystem := NewFactionSystem(game.World, logger)
	game.World.AddSystem(factionSystem)

	// ========================================================================
	// PROGRESSION SYSTEMS (5 systems)
	// ========================================================================

	// 17. ProgressionSystem - XP and leveling
	progressionSystem := NewProgressionSystem(game.World)
	game.World.AddSystem(progressionSystem)

	// 18. SkillProgressionSystem - applies skill effects to stats
	skillProgressionSystem := NewSkillProgressionSystem()
	game.World.AddSystem(skillProgressionSystem)

	// 19. VisualFeedbackSystem - hit flashes and tints
	visualFeedbackSystem := NewVisualFeedbackSystem()
	game.World.AddSystem(visualFeedbackSystem)

	// 20. AudioManagerSystem - updates music based on game context
	result.AudioManager = NewAudioManager(config.SampleRate, config.Seed)
	result.AudioManager.EnableAdaptiveMusic(true)
	audioManagerSystem := NewAudioManagerSystem(result.AudioManager)
	game.World.AddSystem(audioManagerSystem)

	// 21. ObjectiveTrackerSystem - updates quest progress
	result.ObjectiveTracker = NewObjectiveTrackerSystem()
	game.World.AddSystem(result.ObjectiveTracker)

	// ========================================================================
	// INVENTORY & ECONOMY SYSTEMS (8 systems)
	// ========================================================================

	// 22. ItemPickupSystem - collects nearby items
	itemPickupSystem := NewItemPickupSystem(game.World)
	game.World.AddSystem(itemPickupSystem)

	// 23-25. Magic systems
	game.World.AddSystem(spellCastingSystem)
	manaRegenSystem := &ManaRegenSystem{}
	game.World.AddSystem(manaRegenSystem)

	// 26. InventorySystem - item management
	game.World.AddSystem(inventorySystem)

	// 27-29. Commerce and crafting
	result.CommerceSystem = NewCommerceSystemWithLogger(game.World, inventorySystem, logger)
	game.World.AddSystem(result.CommerceSystem)

	result.DialogSystem = NewDialogSystemWithLogger(game.World, logger)
	game.World.AddSystem(result.DialogSystem)

	result.CraftingSystem = NewCraftingSystem(game.World, inventorySystem, nil) // itemGen set later
	game.World.AddSystem(result.CraftingSystem)

	// 30. InteractionSystem - puzzle element interactions (Phase 11.2)
	result.InteractionSystem = NewInteractionSystem(game.World)
	game.World.AddSystem(result.InteractionSystem)

	// ========================================================================
	// VISUAL & ENVIRONMENTAL SYSTEMS (14 systems)
	// ========================================================================

	// 31-32. Animation and equipment visuals
	spriteGenerator := sprites.NewGenerator()
	result.AnimationSystem = NewAnimationSystem(spriteGenerator)
	result.AnimationSystemWrapper = &animationSystemWrapper{
		system: result.AnimationSystem,
		logger: game.World.GetLogger(),
	}
	game.World.AddSystem(result.AnimationSystemWrapper)

	equipmentVisualSystem := NewEquipmentVisualSystem(spriteGenerator)
	game.World.AddSystem(equipmentVisualSystem)

	// 33-34. UI systems
	result.TutorialSystem = NewTutorialSystem()
	game.World.AddSystem(result.TutorialSystem)

	result.HelpSystem = NewHelpSystem()
	game.World.AddSystem(result.HelpSystem)

	// 35. ParticleSystem - visual effects
	result.ParticleSystem = NewParticleSystem()
	game.World.AddSystem(result.ParticleSystem)

	// 36. WeatherSystem - atmospheric effects (Phase 5.4)
	weatherSystem := NewWeatherSystem(game.World)
	game.World.AddSystem(weatherSystem)

	// 37. LifetimeSystem - temporary entities (Phase 5.3)
	lifetimeSystem := NewLifetimeSystemWithLogger(game.World, logger)
	game.World.AddSystem(lifetimeSystem)

	// 38. PuzzleSystem - procedural puzzle management (Phase 11.2)
	puzzleSystem := NewPuzzleSystem(game.World)
	game.World.AddSystem(puzzleSystem)

	// 39-42. Environmental destruction systems (Phase 11.3)
	firePropagationSystem := NewFirePropagationSystemWithLogger(config.TileSize, config.Seed+1090, logger)
	firePropagationSystem.SetWorld(game.World)
	game.World.AddSystem(firePropagationSystem)

	destructibleObjectSystem := NewDestructibleObjectSystemWithLogger(config.TileSize, config.Seed+1100, logger)
	destructibleObjectSystem.SetWorld(game.World)
	destructibleObjectSystem.SetFireSystem(firePropagationSystem)
	game.World.AddSystem(destructibleObjectSystem)

	carrySystem := NewCarrySystemWithLogger(logger)
	carrySystem.SetWorld(game.World)
	game.World.AddSystem(carrySystem)

	hazardSystem := NewHazardSystemWithLogger(logger)
	hazardSystem.SetWorld(game.World)
	game.World.AddSystem(hazardSystem)

	// 43. NarrativeSystem - story progression (Phase 12.2)
	narrativeSystem := NewNarrativeSystem(game.World)
	game.World.AddSystem(narrativeSystem)

	// 43. ShadowSystem - enhanced lighting (Phase 14)
	shadowSystem := NewShadowSystemWithLogger(game.World, logger)
	game.World.AddSystem(shadowSystem)

	// Note: SpatialPartitionSystem (system #44) is initialized separately
	// after terrain generation via InitializeSpatialPartitionSystem()
	// because it requires world dimensions from terrain.

	// ========================================================================
	// POST-INITIALIZATION CONNECTIONS
	// ========================================================================

	// Connect systems that have interdependencies
	result.InputSystem.SetCameraSystem(game.CameraSystem)
	result.InputSystem.SetHelpSystem(result.HelpSystem)
	result.InputSystem.SetTutorialSystem(result.TutorialSystem)

	result.CombatSystem.SetCamera(game.CameraSystem)
	result.CombatSystem.SetParticleSystem(result.ParticleSystem, game.World, config.GenreID)
	result.CombatSystem.SetProjectileSystem(result.ProjectileSystem)

	result.InteractionSystem.SetCarrySystem(carrySystem)
	// INPUT CONFLICT FIX: Connect InteractionSystem to InputSystem for state checking
	result.InteractionSystem.SetInputSystem(result.InputSystem)

	// Store UI system references in game
	game.TutorialSystem = result.TutorialSystem
	game.HelpSystem = result.HelpSystem

	// Wire audio manager to game
	game.SetAudioManager(result.AudioManager)

	if config.EnableVerboseLogging {
		logger.WithFields(logrus.Fields{
			"systemCount": 43,
			"seed":        config.Seed,
			"genre":       config.GenreID,
		}).Info("game systems initialized successfully (44th system requires terrain)")
	}

	return result, nil
}

// InitializeSpatialPartitionSystem initializes the SpatialPartitionSystem (system #44)
// after terrain generation. This must be called separately from InitializeGameSystems()
// because it requires world dimensions from generated terrain.
//
// Parameters:
//   - game: The game instance
//   - worldWidth: Width of the game world in pixels (from terrain)
//   - worldHeight: Height of the game world in pixels (from terrain)
//   - enableCulling: Whether to enable viewport culling optimization
//   - verbose: Whether to log detailed information
//   - logger: Logger for messages
//
// Returns: The initialized SpatialPartitionSystem
func InitializeSpatialPartitionSystem(
	game *EbitenGame,
	worldWidth, worldHeight float64,
	enableCulling bool,
	verbose bool,
	logger *logrus.Logger,
) *SpatialPartitionSystem {
	if verbose && logger != nil {
		logger.Info("initializing spatial partition system for viewport culling")
	}

	// Create spatial partition system with quadtree-based structure
	spatialSystem := NewSpatialPartitionSystem(worldWidth, worldHeight)

	// Register with ECS World for automatic updates every 60 frames
	game.World.AddSystem(spatialSystem)

	// Connect to render system for viewport culling
	game.RenderSystem.SetSpatialPartition(spatialSystem)

	// Enable or disable culling based on parameter
	game.RenderSystem.EnableCulling(enableCulling)

	if verbose && logger != nil {
		cullingStatus := "enabled"
		if !enableCulling {
			cullingStatus = "disabled"
		}
		logger.WithFields(logrus.Fields{
			"worldWidth":  worldWidth,
			"worldHeight": worldHeight,
			"cellSize":    8, // Quadtree capacity per node
			"culling":     cullingStatus,
		}).Info("spatial partition system initialized")
	}

	return spatialSystem
}

// animationSystemWrapper adapts AnimationSystem (returns error) to System interface (no return)
type animationSystemWrapper struct {
	system *AnimationSystem
	logger *logrus.Entry
}

func (w *animationSystemWrapper) Update(entities []*Entity, deltaTime float64) {
	if err := w.system.Update(entities, deltaTime); err != nil {
		if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
			w.logger.WithError(err).Debug("animation system error")
		}
	}
}

// rotationSystemWrapper adapts RotationSystem to System interface
type rotationSystemWrapper struct {
	system *RotationSystem
}

func (w *rotationSystemWrapper) Update(entities []*Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// squadSystemWrapper adapts SquadSystem to System interface
type squadSystemWrapper struct {
	system *SquadSystem
}

func (w *squadSystemWrapper) Update(entities []*Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}
