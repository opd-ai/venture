package mobile

import (
	"math/rand"
	"os"
	"runtime/debug"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/opd-ai/venture/cmd/mobile/config"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/logging"
	mobilepkg "github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/quality"
	"github.com/sirupsen/logrus"
)

// Default mobile screen dimensions (landscape orientation)
const (
	// DefaultScreenWidth is the default logical width for mobile devices (landscape)
	DefaultScreenWidth = 1280
	// DefaultScreenHeight is the default logical height for mobile devices (landscape)
	DefaultScreenHeight = 720
)

// Mobile-specific performance limits.
// These are significantly lower than the desktop defaults to prevent
// memory exhaustion on RAM-constrained mobile devices.
const (
	// mobileSpriteCacheMaxSize is the hard cap for the sprite cache (100MB).
	// Desktop uses 400MB; mobile devices typically have 2-4GB shared RAM.
	mobileSpriteCacheMaxSize = 100 * 1024 * 1024

	// mobileAnimationCacheSize is the max number of cached animation sequences.
	// Desktop uses 300; reduced to prevent memory growth from diverse animations.
	mobileAnimationCacheSize = 100

	// mobileMemorySoftLimit is the sprite-cache soft limit that triggers LRU cleanup.
	mobileMemorySoftLimit = 80 * 1024 * 1024 // 80MB

	// mobileMemoryHardLimit is the sprite-cache hard limit that forces eviction.
	mobileMemoryHardLimit = 100 * 1024 * 1024 // 100MB
)

// Game is the mobile game instance
var (
	gameInstance      *engine.EbitenGame
	logger            *logrus.Logger
	systemsInitResult *engine.SystemInitResult
	playerEntity      *engine.Entity
	worldSeed         int64
	genreID           string

	// Mobile performance management
	spriteCache   *cache.SpriteCache
	memoryMonitor *cache.MemoryMonitor
	qualitySystem *engine.QualitySystem

	// Mobile input controls
	mobileInput *mobilepkg.MobileInputAdapter
)

func init() {
	// Initialize logger for mobile
	logConfig := logging.DefaultConfig()
	logConfig.Format = logging.TextFormat
	logConfig.Level = logging.InfoLevel
	logger = logging.NewLogger(logConfig)

	// Generate seed (check environment variable first for testing/debugging)
	worldSeed = config.GetSeedFromEnv(logger)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	rng := rand.New(rand.NewSource(worldSeed))
	genreID = config.GetGenreFromEnv(genres, rng, logger)

	// Skip game initialization during testing (mobile.SetGame panics in non-mobile env)
	// Tests set VENTURE_SKIP_MOBILE_INIT=1 to prevent this
	// Also auto-detect Go test binaries to avoid panics
	if os.Getenv("VENTURE_SKIP_MOBILE_INIT") == "" && !isTestBinary() {
		// Initialize the game immediately for ebitenmobile
		initializeGame()
	}
}

// isTestBinary detects if the current binary is a Go test binary
// by checking for -test.* flags that go test always passes.
func isTestBinary() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}

// initializeGame initializes the game for mobile platforms.
func initializeGame() {
	if gameInstance != nil {
		return
	}

	logger.WithFields(logrus.Fields{
		"seed":  worldSeed,
		"genre": genreID,
	}).Info("initializing mobile game with Version 2.0 systems")

	initializeGameInstance()
	generatedTerrain := initializeTerrainAndSystems()
	spawnEnemiesInWorld(generatedTerrain)
	createAndConfigurePlayer(generatedTerrain)
	finalizeGameSetup()

	logger.Info("mobile game fully initialized with all Version 2.0 systems")

	mobile.SetGame(gameInstance)
}

// initializeGameInstance creates the game instance and initializes all game systems.
func initializeGameInstance() {
	gameInstance = engine.NewEbitenGameWithLogger(DefaultScreenWidth, DefaultScreenHeight, logger)

	config := engine.DefaultSystemInitConfig(worldSeed, genreID, logger)
	config.EnableVerboseLogging = true

	var err error
	systemsInitResult, err = engine.InitializeGameSystems(gameInstance, config)
	if err != nil {
		logger.WithError(err).Fatal("failed to initialize game systems")
		// Fatal exits the program, so this is unreachable
	}

	// --- Mobile performance controls (not present in InitializeGameSystems) ---

	// 1. Sprite cache with memory-safe limits
	spriteCache = cache.NewSpriteCache(mobileSpriteCacheMaxSize)
	systemsInitResult.AnimationSystem.SetMaxCacheSize(mobileAnimationCacheSize)
	systemsInitResult.AnimationSystem.SetSpriteCache(spriteCache)

	// 2. Memory monitor — runs a background goroutine that evicts sprites
	//    when the cache exceeds soft/hard limits, preventing OOM crashes.
	memoryMonitor = cache.NewMemoryMonitor(spriteCache)
	memoryMonitor.SetLimits(mobileMemorySoftLimit, mobileMemoryHardLimit)
	memoryMonitor.Start()

	// 3. Quality system — uses Low preset for mobile with auto-adjustment.
	//    This caps particle counts, disables expensive post-processing,
	//    and scales back visual effects to keep frame rate stable.
	lowQuality := quality.LowQualityConfig()
	qualitySystem = engine.NewQualitySystem(&lowQuality, 60.0)
	gameInstance.World.AddSystem(&mobileQualitySystemWrapper{system: qualitySystem})

	// 4. Performance monitoring system — tracks frame times so the quality
	//    auto-adjuster has data to react to.
	perfSystem := engine.NewPerformanceMonitoringSystem()
	gameInstance.World.AddSystem(perfSystem)

	// 5. Aggressive GC tuning for mobile — lower the target percentage so the
	//    garbage collector runs more often, keeping peak RSS in check.
	debug.SetGCPercent(50)

	logger.WithFields(logrus.Fields{
		"systemCount":        43,
		"spriteCacheMB":      mobileSpriteCacheMaxSize / (1024 * 1024),
		"animationCacheSize": mobileAnimationCacheSize,
		"qualityLevel":       "Low",
	}).Info("game systems initialized with mobile performance controls")
}

// initializeTerrainAndSystems generates terrain and configures terrain-dependent systems.
func initializeTerrainAndSystems() *terrain.Terrain {
	logger.Info("generating procedural terrain")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"width":  60,
			"height": 40,
		},
	}

	terrainResult, err := terrainGen.Generate(worldSeed, params)
	if err != nil {
		logger.WithError(err).Fatal("failed to generate terrain")
		// Fatal exits the program, so this is unreachable
	}

	generatedTerrain := terrainResult.(*terrain.Terrain)
	logger.WithFields(logrus.Fields{
		"width":     generatedTerrain.Width,
		"height":    generatedTerrain.Height,
		"roomCount": len(generatedTerrain.Rooms),
	}).Info("terrain generated")

	setupTerrainSystems(generatedTerrain)
	return generatedTerrain
}

// setupTerrainSystems initializes terrain rendering, collision, and spatial partition systems.
func setupTerrainSystems(generatedTerrain *terrain.Terrain) {
	terrainRenderSystem := engine.NewTerrainRenderSystem(32, 32, genreID, worldSeed)
	terrainRenderSystem.SetTerrain(generatedTerrain)
	gameInstance.TerrainRenderSystem = terrainRenderSystem

	terrainChecker := engine.NewTerrainCollisionChecker(32, 32)
	terrainChecker.SetTerrain(generatedTerrain)

	systemsInitResult.CollisionSystem.SetTerrainChecker(terrainChecker)
	systemsInitResult.ProjectileSystem.SetTerrainChecker(terrainChecker)

	worldWidth := float64(generatedTerrain.Width) * 32.0
	worldHeight := float64(generatedTerrain.Height) * 32.0
	engine.InitializeSpatialPartitionSystem(gameInstance, worldWidth, worldHeight, true, true, logger)

	logger.Info("all 44 systems initialized")
}

// spawnEnemiesInWorld spawns enemies in the generated terrain.
func spawnEnemiesInWorld(generatedTerrain *terrain.Terrain) {
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
}

// createAndConfigurePlayer creates the player entity and adds all necessary components.
func createAndConfigurePlayer(generatedTerrain *terrain.Terrain) {
	playerEntity = gameInstance.World.CreateEntity()

	playerX, playerY := calculatePlayerSpawnPosition(generatedTerrain)

	addBasicPlayerComponents(playerX, playerY)
	addVisualPlayerComponents()
	configurePlayerSystems()
	addPlayerStatsAndInventory()
	loadPlayerAbilities()
	addPlayerCombatComponents()
}

// calculatePlayerSpawnPosition determines the player's starting position.
// If terrain has rooms, spawns player in the center of the first room.
// Otherwise, uses a safe default position in the middle of the screen.
func calculatePlayerSpawnPosition(generatedTerrain *terrain.Terrain) (float64, float64) {
	if generatedTerrain != nil && len(generatedTerrain.Rooms) > 0 {
		firstRoom := generatedTerrain.Rooms[0]
		cx, cy := firstRoom.Center()
		return float64(cx * 32), float64(cy * 32)
	}
	return 400, 300
}

// addBasicPlayerComponents adds core components to the player entity.
func addBasicPlayerComponents(playerX, playerY float64) {
	playerEntity.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	playerEntity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	playerEntity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	playerEntity.AddComponent(&engine.TeamComponent{TeamID: 1})

	// Create mobile touch input adapter instead of desktop keyboard/mouse input
	mobileInput = mobilepkg.NewMobileInputAdapter(DefaultScreenWidth, DefaultScreenHeight)
	playerEntity.AddComponent(mobileInput)

	playerEntity.AddComponent(engine.NewRotationComponent(0, 3.0))
	playerEntity.AddComponent(engine.NewAimComponent(0))

	logger.WithFields(logrus.Fields{
		"input_type":    "mobile_touch",
		"screen_width":  DefaultScreenWidth,
		"screen_height": DefaultScreenHeight,
	}).Info("configured mobile touch input")
}

// addVisualPlayerComponents adds sprite, animation, camera, and visual effect components.
func addVisualPlayerComponents() {
	const playerSpriteSize = 64
	const playerColliderOff = -32

	playerSprite := &engine.EbitenSprite{
		Image:   nil,
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
	playerAnim.FrameCount = 8
	playerEntity.AddComponent(playerAnim)

	playerEntity.AddComponent(engine.NewEquipmentVisualComponent())

	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	playerEntity.AddComponent(camera)

	playerEntity.AddComponent(engine.NewScreenShakeComponent())
	playerEntity.AddComponent(engine.NewHitStopComponent())

	layerComp := engine.NewLayerComponent()
	layerComp.CurrentLayer = 0
	playerEntity.AddComponent(&layerComp)

	playerShadow := engine.NewShadowComponent(playerSpriteSize)
	playerShadow.CastsShadow = true
	playerShadow.ShadowType = engine.ShadowTypeSoft
	playerEntity.AddComponent(playerShadow)

	playerEntity.AddComponent(&engine.ColliderComponent{
		Width:     playerSpriteSize,
		Height:    playerSpriteSize,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   playerColliderOff,
		OffsetY:   playerColliderOff,
	})

	playerEntity.AddComponent(engine.NewVisualFeedbackComponent())
}

// configurePlayerSystems configures camera, animation, and HUD systems for the player.
func configurePlayerSystems() {
	gameInstance.CameraSystem.SetActiveCamera(playerEntity)
	systemsInitResult.AnimationSystem.SetCameraSystem(gameInstance.CameraSystem)
	systemsInitResult.AnimationSystem.SetPlayerEntity(playerEntity)
	gameInstance.HUDSystem.SetPlayerEntity(playerEntity)
	gameInstance.SetPlayerEntity(playerEntity)
}

// addPlayerStatsAndInventory adds stats, experience, inventory, equipment, and mana.
func addPlayerStatsAndInventory() {
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	playerStats.CritChance = 0.05
	playerStats.CritDamage = 1.5
	playerStats.Evasion = 0.05
	playerEntity.AddComponent(playerStats)

	playerEntity.AddComponent(engine.NewExperienceComponent())

	playerInventory := engine.NewInventoryComponent(20, 100.0)
	playerInventory.Gold = 100
	playerEntity.AddComponent(playerInventory)

	playerEntity.AddComponent(engine.NewEquipmentComponent())

	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0,
	}
	playerEntity.AddComponent(playerMana)

	addStarterItems(playerInventory, worldSeed, genreID)
}

// loadPlayerAbilities loads spells and skill trees for the player.
func loadPlayerAbilities() {
	err := engine.LoadPlayerSpells(playerEntity, worldSeed, genreID, 1)
	if err != nil {
		logger.WithError(err).Warn("failed to load player spells")
	}

	err = engine.LoadPlayerSkillTree(playerEntity, worldSeed, genreID, 0)
	if err != nil {
		logger.WithError(err).Warn("failed to load skill tree")
	}
}

// addPlayerCombatComponents adds quest tracking, attack, and soundtrack components.
func addPlayerCombatComponents() {
	playerEntity.AddComponent(engine.NewQuestTrackerComponent(5))

	playerEntity.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: 1,
		Range:      50,
		Cooldown:   0.5,
	})

	playerEntity.AddComponent(engine.NewAdaptiveSoundtrackComponent(genreID))
}

// finalizeGameSetup starts background music and processes initial entity additions.
func finalizeGameSetup() {
	if err := systemsInitResult.AudioManager.PlayMusic(genreID, "exploration"); err != nil {
		logger.WithError(err).Warn("failed to start background music")
	}

	gameInstance.World.Update(0)
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
	if err != nil {
		logger.WithError(err).Debug("starter weapon generation failed, skipping")
	} else {
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
	if err != nil {
		logger.WithError(err).Debug("starter potion generation failed, skipping")
	} else {
		potions := potionResult.([]*item.Item)
		for _, potion := range potions {
			potion.Name = "Minor Health Potion"
			potion.Stats.Value = 10
			potion.Stats.Weight = 0.2
			inventory.Items = append(inventory.Items, potion)
		}
	}
}

// Start initializes and starts the game loop for mobile platforms.
// This function is called automatically by the iOS/Android platform binding
// when the application launches. It ensures the game is initialized before
// the main loop begins. If called multiple times, subsequent calls are no-ops.
func Start() {
	if gameInstance == nil {
		initializeGame()
	}
}

// Update is called each frame by the mobile platform to update game state.
// It returns true to continue running the game loop, or false to signal
// that the application should quit. Currently returns true if the game
// instance is initialized, false otherwise.
func Update() bool {
	return gameInstance != nil
}

// GetScreenWidth returns the game's logical screen width in pixels.
// For mobile, this defaults to DefaultScreenWidth (landscapee orientation).
// Returns 0 if the game instance has not been initialized.
func GetScreenWidth() int {
	if gameInstance == nil {
		return 0
	}
	return DefaultScreenWidth
}

// GetScreenHeight returns the game's logical screen height in pixels.
// For mobile, this defaults to DefaultScreenHeight (landscapee orientation).
// Returns 0 if the game instance has not been initialized.
func GetScreenHeight() int {
	if gameInstance == nil {
		return 0
	}
	return DefaultScreenHeight
}

// mobileQualitySystemWrapper adapts QualitySystem to the ECS System interface
// so it can be registered with World.AddSystem and updated each frame.
type mobileQualitySystemWrapper struct {
	system *engine.QualitySystem
}

func (w *mobileQualitySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}
