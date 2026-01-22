package mobile

import (
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// Game is the mobile game instance
var (
	gameInstance      *engine.EbitenGame
	logger            *logrus.Logger
	systemsInitResult *engine.SystemInitResult
	playerEntity      *engine.Entity
	worldSeed         int64
	genreID           string
)

func init() {
	// Initialize logger for mobile
	logConfig := logging.DefaultConfig()
	logConfig.Format = logging.TextFormat
	logConfig.Level = logging.InfoLevel
	logger = logging.NewLogger(logConfig)

	// Generate default seed and genre
	worldSeed = time.Now().UnixNano()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	rng := rand.New(rand.NewSource(worldSeed))
	genreID = genres[rng.Intn(len(genres))]

	// Initialize the game immediately for ebitenmobile
	initializeGame()
}

// initializeGame initializes the game for mobile platforms.
func initializeGame() {
	if gameInstance != nil {
		return // Already initialized
	}

	logger.WithFields(logrus.Fields{
		"seed":  worldSeed,
		"genre": genreID,
	}).Info("initializing mobile game with Version 2.0 systems")

	// Create the game instance with mobile-friendly dimensions
	// Portrait mode: 720x1280 (9:16 aspect ratio)
	gameInstance = engine.NewEbitenGameWithLogger(720, 1280, logger)

	// Initialize all game systems using shared initialization
	config := engine.DefaultSystemInitConfig(worldSeed, genreID, logger)
	config.EnableVerboseLogging = true

	var err error
	systemsInitResult, err = engine.InitializeGameSystems(gameInstance, config)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize game systems")
		return
	}

	logger.WithField("systemCount", 43).Info("game systems initialized successfully")

	// Generate initial world terrain
	logger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"width":  60, // Smaller for mobile
			"height": 40,
		},
	}

	terrainResult, err := terrainGen.Generate(worldSeed, params)
	if err != nil {
		logger.WithError(err).Fatal("failed to generate terrain")
		return
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	logger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("terrain generated")

	// Initialize terrain rendering system
	terrainRenderSystem := engine.NewTerrainRenderSystem(32, 32, genreID, worldSeed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	gameInstance.TerrainRenderSystem = terrainRenderSystem

	// Initialize terrain collision checking
	terrainChecker := engine.NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(generatedTerrain)

	// Connect terrain checker to collision and projectile systems
	systemsInitResult.CollisionSystem.SetTerrainChecker(terrainChecker)
	systemsInitResult.ProjectileSystem.SetTerrainChecker(terrainChecker)

	// Initialize spatial partition system (system #44)
	worldWidth := float64(generatedTerrain.Width) * 32.0
	worldHeight := float64(generatedTerrain.Height) * 32.0
	engine.InitializeSpatialPartitionSystem(gameInstance, worldWidth, worldHeight, true, true, logger)

	logger.Info("all 44 systems initialized")

	// Spawn enemies in terrain
	enemyParams := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
	}

	enemyCount, err := engine.SpawnEnemiesInTerrain(gameInstance.World, generatedTerrain, worldSeed, enemyParams)
	if err != nil {
		logger.WithError(err).Warn("failed to spawn enemies")
	} else {
		logger.WithField("enemyCount", enemyCount).Info("spawned enemies")
	}

	// Create player entity
	playerEntity = gameInstance.World.CreateEntity()

	// Calculate player spawn position from first room
	var playerX, playerY float64
	if len(generatedTerrain.Rooms) > 0 {
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		playerX = float64(cx * 32)
		playerY = float64(cy * 32)
	} else {
		playerX, playerY = 400, 300
	}

	// Add player components
	playerEntity.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	playerEntity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	playerEntity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	playerEntity.AddComponent(&engine.TeamComponent{TeamID: 1})
	playerEntity.AddComponent(&engine.EbitenInput{})

	// Phase 10.1: Add rotation and aim components
	playerEntity.AddComponent(engine.NewRotationComponent(0, 3.0))
	playerEntity.AddComponent(engine.NewAimComponent(0))

	// Add sprite and animation (Phase 45: 64×64 enhanced sprites)
	const playerSpriteSize = 64  // Phase 45 standard sprite size
	const playerColliderOff = -32 // Center offset (64/2 = 32)
	playerSprite := &engine.EbitenSprite{
		Image:   nil, // Will be generated by animation system
		Width:   playerSpriteSize,
		Height:  playerSpriteSize,
		Visible: true,
		Layer:   10,
	}
	playerEntity.AddComponent(playerSprite)

	playerAnim := engine.NewAnimationComponent(worldSeed + int64(playerEntity.ID*1000))
	playerAnim.CurrentState = engine.AnimationStateIdle
	playerAnim.FrameTime = 0.15
	playerAnim.Loop = true
	playerAnim.Playing = true
	playerAnim.FrameCount = 4
	playerEntity.AddComponent(playerAnim)

	// Add equipment visual component
	equipmentVisualComp := engine.NewEquipmentVisualComponent()
	playerEntity.AddComponent(equipmentVisualComp)

	// Add camera
	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	playerEntity.AddComponent(camera)

	// Phase 10.3: Add screen shake and hit-stop
	playerEntity.AddComponent(engine.NewScreenShakeComponent())
	playerEntity.AddComponent(engine.NewHitStopComponent())

	// Phase 11.1: Add layer component
	layerComp := engine.NewLayerComponent()
	layerComp.CurrentLayer = 0
	playerEntity.AddComponent(&layerComp)

	// Phase 14: Add shadow (Phase 45: shadow matches 64×64 sprite)
	playerShadow := engine.NewShadowComponent(playerSpriteSize)
	playerShadow.CastsShadow = true
	playerShadow.ShadowType = engine.ShadowTypeSoft
	playerEntity.AddComponent(playerShadow)

	// Set player as active camera
	gameInstance.CameraSystem.SetActiveCamera(playerEntity)

	// Configure animation system
	systemsInitResult.AnimationSystem.SetCameraSystem(gameInstance.CameraSystem)
	systemsInitResult.AnimationSystem.SetPlayerEntity(playerEntity)

	// Set player for HUD and UI
	gameInstance.HUDSystem.SetPlayerEntity(playerEntity)
	gameInstance.SetPlayerEntity(playerEntity)

	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	playerStats.CritChance = 0.05
	playerStats.CritDamage = 1.5
	playerStats.Evasion = 0.05
	playerEntity.AddComponent(playerStats)

	// Add player experience
	playerExp := engine.NewExperienceComponent()
	playerEntity.AddComponent(playerExp)

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0)
	playerInventory.Gold = 100
	playerEntity.AddComponent(playerInventory)

	// Add player equipment
	playerEquipment := engine.NewEquipmentComponent()
	playerEntity.AddComponent(playerEquipment)

	// Add mana
	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0,
	}
	playerEntity.AddComponent(playerMana)

	// Load spells
	err = engine.LoadPlayerSpells(playerEntity, worldSeed, genreID, 1)
	if err != nil {
		logger.WithError(err).Warn("failed to load player spells")
	}

	// Load skill tree
	err = engine.LoadPlayerSkillTree(playerEntity, worldSeed, genreID, 0)
	if err != nil {
		logger.WithError(err).Warn("failed to load skill tree")
	}

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5)
	playerEntity.AddComponent(questTracker)

	// Add attack capability
	playerEntity.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: 1, // Physical
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision (Phase 45: collider matches 64×64 sprite)
	playerEntity.AddComponent(&engine.ColliderComponent{
		Width:     playerSpriteSize,
		Height:    playerSpriteSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   playerColliderOff,
		OffsetY:   playerColliderOff,
	})

	// Add visual feedback
	playerEntity.AddComponent(engine.NewVisualFeedbackComponent())

	// Add adaptive soundtrack component for dynamic music (Phase 29)
	playerEntity.AddComponent(engine.NewAdaptiveSoundtrackComponent(genreID))

	// Add starter items
	addStarterItems(playerInventory, worldSeed, genreID)

	// Setup audio
	if err := systemsInitResult.AudioManager.PlayMusic(genreID, "exploration"); err != nil {
		logger.WithError(err).Warn("failed to start background music")
	}

	// Process initial entity additions
	gameInstance.World.Update(0)

	logger.Info("mobile game fully initialized with all Version 2.0 systems")

	// Register the game with ebitenmobile
	mobile.SetGame(gameInstance)
}

// addStarterItems generates and adds starting items to the player's inventory.
func addStarterItems(inventory *engine.InventoryComponent, seed int64, genreID string) {
	itemGen := item.NewItemGenerator()

	// Generate a starting weapon
	weaponParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "weapon",
		},
	}

	weaponResult, err := itemGen.Generate(seed+1, weaponParams)
	if err == nil {
		weapons := weaponResult.([]*item.Item)
		if len(weapons) > 0 {
			weapon := weapons[0]
			weapon.Name = "Rusty " + weapon.Name
			weapon.Stats.Value = 5
			inventory.Items = append(inventory.Items, weapon)
		}
	}

	// Generate 2 healing potions
	potionParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 2,
			"type":  "consumable",
		},
	}

	potionResult, err := itemGen.Generate(seed+2, potionParams)
	if err == nil {
		potions := potionResult.([]*item.Item)
		for _, potion := range potions {
			potion.Name = "Minor Health Potion"
			potion.Stats.Value = 10
			potion.Stats.Weight = 0.2
			inventory.Items = append(inventory.Items, potion)
		}
	}
}

// Start starts the game loop.
// This is called automatically by the mobile platform.
func Start() {
	if gameInstance == nil {
		initializeGame()
	}
}

// Update updates the game state.
// Returns true to continue running, false to quit.
func Update() bool {
	return gameInstance != nil
}

// GetScreenWidth returns the screen width.
func GetScreenWidth() int {
	if gameInstance == nil {
		return 0
	}
	return 720
}

// GetScreenHeight returns the screen height.
func GetScreenHeight() int {
	if gameInstance == nil {
		return 0
	}
	return 1280
}
