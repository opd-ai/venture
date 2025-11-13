//go:build !android && !ios
// +build !android,!ios

// Package main provides the desktop client application.
// For mobile platforms (Android/iOS), use cmd/mobile with ebitenmobile build tool.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/hostplay"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

// animationSystemWrapper adapts AnimationSystem (returns error) to System interface (no return)
type animationSystemWrapper struct {
	system *engine.AnimationSystem
	logger *logrus.Entry
}

func (w *animationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	if err := w.system.Update(entities, deltaTime); err != nil {
		if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
			w.logger.WithError(err).Debug("animation system error")
		}
	}
}

// rotationSystemWrapper adapts RotationSystem to System interface
type rotationSystemWrapper struct {
	system *engine.RotationSystem
}

func (w *rotationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// squadSystemWrapper adapts SquadSystem to System interface
type squadSystemWrapper struct {
	system *engine.SquadSystem
}

func (w *squadSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// reputationSystemWrapper adapts ReputationSystem to System interface
type reputationSystemWrapper struct {
	system *engine.ReputationSystem
}

func (w *reputationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// alignmentSystemWrapper adapts AlignmentSystem to System interface
type alignmentSystemWrapper struct {
	system *engine.AlignmentSystem
}

func (w *alignmentSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// factionReactionSystemWrapper adapts FactionReactionSystem to System interface
type factionReactionSystemWrapper struct {
	system *engine.FactionReactionSystem
}

func (w *factionReactionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

var (
	width            = flag.Int("width", 800, "Screen width")
	height           = flag.Int("height", 600, "Screen height")
	seed             = flag.Int64("seed", seededRandom(), "World generation seed")
	genreID          = flag.String("genre", randomGenre(), "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	enableLighting   = flag.Bool("enable-lighting", true, "Enable dynamic lighting system")
	enableWeather    = flag.Bool("enable-weather", true, "Enable procedural weather effects")
	weatherType      = flag.String("weather", "", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation) - empty for genre-appropriate random")
	weatherIntensity = flag.String("weather-intensity", "medium", "Weather intensity (light, medium, heavy, extreme)")
	verbose          = flag.Bool("verbose", false, "Enable verbose logging")
	profile          = flag.Bool("profile", false, "Enable performance profiling with frame time tracking")
	multiplayer      = flag.Bool("multiplayer", false, "Enable multiplayer mode (connect to server)")
	server           = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
	hostAndPlay      = flag.Bool("host-and-play", false, "Host server and auto-connect (single command LAN party mode)")
	hostLAN          = flag.Bool("host-lan", false, "Bind server to 0.0.0.0 for LAN access (use with --host-and-play, default is localhost only)")
	serverPort       = flag.Int("port", 8080, "Server port for --host-and-play mode (will try next 10 ports if occupied)")
	serverPlayers    = flag.Int("max-players", 4, "Maximum players for --host-and-play mode")
	serverTick       = flag.Int("tick-rate", 20, "Server tick rate for --host-and-play mode (updates per second)")
	noTutorial       = flag.Bool("no-tutorial", false, "Disable tutorial for experienced players")
)

// initializeLogger creates and configures the logger based on environment variables and flags.
func initializeLogger() (*logrus.Logger, *logrus.Entry) {
	logConfig := logging.DefaultConfig()

	// Check for JSON format from environment (default to text for client)
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat == "json" {
		logConfig.Format = logging.JSONFormat
	} else {
		logConfig.Format = logging.TextFormat
		logConfig.EnableColor = true
	}

	// Set log level from environment or use Info as default
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		logConfig.Level = logging.LogLevel(logLevel)
	} else if *verbose {
		logConfig.Level = logging.DebugLevel
	} else {
		logConfig.Level = logging.InfoLevel
	}

	logger := logging.NewLogger(logConfig)
	clientLogger := logger.WithFields(logrus.Fields{
		"component": "client",
		"genre":     *genreID,
		"seed":      *seed,
	})

	clientLogger.Infof("Starting Venture %s", version.FullVersion)
	clientLogger.WithFields(logrus.Fields{
		"width":  *width,
		"height": *height,
		"seed":   *seed,
		"genre":  *genreID,
	}).Info("client configuration")

	return logger, clientLogger
}

// initializeNetworkClient sets up network connection if multiplayer mode is enabled.
// Returns the network client or nil for single-player mode.
func initializeNetworkClient(logger *logrus.Logger, clientLogger *logrus.Entry) network.ClientConnection {
	if !*multiplayer {
		clientLogger.Info("single-player mode (use -multiplayer flag to connect to server)")
		return nil
	}

	clientLogger.WithField("server", *server).Info("multiplayer mode enabled - connecting to server")

	clientConfig := network.DefaultClientConfig()
	clientConfig.ServerAddress = *server
	networkClient := network.NewClientWithLogger(clientConfig, logger)

	// Connect to server
	if err := networkClient.Connect(); err != nil {
		clientLogger.WithError(err).Fatal("failed to connect to server")
	}

	clientLogger.Info("connected to server successfully")

	// Handle network errors in background
	go func() {
		for err := range networkClient.ReceiveError() {
			clientLogger.WithError(err).Error("network error")
		}
	}()

	return networkClient
}

// return a random seed
func seededRandom() int64 {
	time := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(time))
	return rand.Int63()
}

// return a random genre
func randomGenre() string {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	time := time.Now().UnixNano()
	rand := rand.New(rand.NewSource(time))
	return genres[rand.Intn(len(genres))]
}

// spawnEnvironmentalLights creates atmospheric lighting throughout the dungeon.
// Spawns wall torches, magical crystals, and genre-specific lights based on the world seed.
// This function is part of Phase 5.3: Dynamic Lighting System Integration.
// lightConfig defines lighting configuration per genre.
type lightConfig struct {
	torchInterval int // Every N tiles along walls/corridors
	crystalChance float64
	torchColor    color.RGBA
	crystalColor  color.RGBA
	torchRadius   float64
	crystalRadius float64
	torchFlicker  bool
	crystalPulse  bool
}

// getLightConfig returns the lighting configuration for the given genre.
func getLightConfig(genreID string) lightConfig {
	configs := map[string]lightConfig{
		"fantasy": {
			torchInterval: 5,
			crystalChance: 0.15,
			torchColor:    color.RGBA{255, 150, 80, 255},  // Warm torch light
			crystalColor:  color.RGBA{150, 200, 255, 255}, // Blue magical crystal
			torchRadius:   150,
			crystalRadius: 120,
			torchFlicker:  true,
			crystalPulse:  true,
		},
		"scifi": {
			torchInterval: 4,
			crystalChance: 0.20,
			torchColor:    color.RGBA{150, 200, 255, 255}, // Cool neon blue
			crystalColor:  color.RGBA{0, 255, 200, 255},   // Cyan tech light
			torchRadius:   180,
			crystalRadius: 140,
			torchFlicker:  false,
			crystalPulse:  true,
		},
		"horror": {
			torchInterval: 7,
			crystalChance: 0.08,
			torchColor:    color.RGBA{180, 140, 100, 255}, // Dim yellowish
			crystalColor:  color.RGBA{120, 80, 80, 255},   // Faint reddish
			torchRadius:   100,
			crystalRadius: 80,
			torchFlicker:  true,
			crystalPulse:  false,
		},
		"cyberpunk": {
			torchInterval: 3,
			crystalChance: 0.25,
			torchColor:    color.RGBA{255, 0, 150, 255}, // Neon pink
			crystalColor:  color.RGBA{0, 255, 255, 255}, // Cyan hologram
			torchRadius:   160,
			crystalRadius: 130,
			torchFlicker:  false,
			crystalPulse:  true,
		},
		"postapoc": {
			torchInterval: 6,
			crystalChance: 0.10,
			torchColor:    color.RGBA{200, 180, 140, 255}, // Dusty yellow
			crystalColor:  color.RGBA{100, 255, 100, 255}, // Radioactive green
			torchRadius:   120,
			crystalRadius: 100,
			torchFlicker:  true,
			crystalPulse:  true,
		},
	}

	// Get configuration for this genre (default to fantasy if unknown)
	config, ok := configs[genreID]
	if !ok {
		config = configs["fantasy"]
	}
	return config
}

// spawnWallTorches spawns torch lights along the perimeter walls of a room.
func spawnWallTorches(world *engine.World, room *terrain.Room, config lightConfig, rng *rand.Rand) int {
	count := 0
	const tileSize = 32
	const spawnChance = 0.6 // 60% chance per position

	// Top and bottom walls
	for x := room.X; x < room.X+room.Width; x++ {
		if x%config.torchInterval == 0 {
			// Top wall
			if rng.Float64() < spawnChance {
				worldX := float64(x * tileSize)
				worldY := float64(room.Y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker)
				count++
			}
			// Bottom wall
			if rng.Float64() < spawnChance {
				worldX := float64(x * tileSize)
				worldY := float64((room.Y + room.Height - 1) * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker)
				count++
			}
		}
	}

	// Left and right walls
	for y := room.Y; y < room.Y+room.Height; y++ {
		if y%config.torchInterval == 0 {
			// Left wall
			if rng.Float64() < spawnChance {
				worldX := float64(room.X * tileSize)
				worldY := float64(y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker)
				count++
			}
			// Right wall
			if rng.Float64() < spawnChance {
				worldX := float64((room.X + room.Width - 1) * tileSize)
				worldY := float64(y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker)
				count++
			}
		}
	}

	return count
}

func spawnEnvironmentalLights(world *engine.World, terrain *terrain.Terrain, seed int64, genreID string) int {
	rng := rand.New(rand.NewSource(seed))
	lightCount := 0
	config := getLightConfig(genreID)

	// Spawn lights in each room
	for _, room := range terrain.Rooms {
		// Skip entrance room (index 0) - keep it dark for dramatic effect
		if room == terrain.Rooms[0] {
			continue
		}

		// Spawn wall torches around room perimeter
		lightCount += spawnWallTorches(world, room, config, rng)

		// Spawn magical crystals in room centers (boss rooms, treasure rooms)
		if rng.Float64() < config.crystalChance {
			cx, cy := room.Center()
			worldX := float64(cx * 32)
			worldY := float64(cy * 32)
			spawnCrystalLight(world, worldX, worldY, config.crystalColor, config.crystalRadius, config.crystalPulse)
			lightCount++
		}
	}

	return lightCount
}

// spawnTorchLight creates a wall torch light entity.
func spawnTorchLight(world *engine.World, x, y float64, color color.RGBA, radius float64, flicker bool) {
	lightEntity := world.CreateEntity()
	lightEntity.AddComponent(&engine.PositionComponent{X: x, Y: y})

	torchLight := engine.NewTorchLight(radius)
	torchLight.Color = color
	torchLight.Enabled = true
	if flicker {
		torchLight.Flickering = true
		torchLight.FlickerSpeed = 2.0 + (rand.Float64() * 2.0) // Vary flicker speed
		torchLight.FlickerAmount = 0.15
	}
	lightEntity.AddComponent(torchLight)
}

// spawnCrystalLight creates a magical crystal light entity.
func spawnCrystalLight(world *engine.World, x, y float64, color color.RGBA, radius float64, pulse bool) {
	lightEntity := world.CreateEntity()
	lightEntity.AddComponent(&engine.PositionComponent{X: x, Y: y})

	crystalLight := engine.NewCrystalLight(radius, color)
	crystalLight.Enabled = true
	if pulse {
		crystalLight.Pulsing = true
		crystalLight.PulseSpeed = 1.5
		crystalLight.PulseAmount = 0.25
	}
	lightEntity.AddComponent(crystalLight)
}

// spawnWeather creates a weather effect entity.
// Phase 5.4: Weather Particle System Integration
func spawnWeather(world *engine.World, screenWidth, screenHeight int, seed int64, genreID, weatherTypeStr, intensityStr string) *engine.Entity {
	rng := rand.New(rand.NewSource(seed))

	// Parse weather intensity
	var intensity particles.WeatherIntensity
	switch strings.ToLower(intensityStr) {
	case "light":
		intensity = particles.IntensityLight
	case "medium":
		intensity = particles.IntensityMedium
	case "heavy":
		intensity = particles.IntensityHeavy
	case "extreme":
		intensity = particles.IntensityExtreme
	default:
		intensity = particles.IntensityMedium
	}

	// Determine weather type
	var weatherType particles.WeatherType
	if weatherTypeStr == "" {
		// Select genre-appropriate random weather
		genreWeathers := particles.GetGenreWeather(genreID)
		if len(genreWeathers) > 0 {
			weatherType = genreWeathers[rng.Intn(len(genreWeathers))]
		} else {
			weatherType = particles.WeatherRain
		}
	} else {
		// Parse explicit weather type
		switch strings.ToLower(weatherTypeStr) {
		case "rain":
			weatherType = particles.WeatherRain
		case "snow":
			weatherType = particles.WeatherSnow
		case "fog":
			weatherType = particles.WeatherFog
		case "dust":
			weatherType = particles.WeatherDust
		case "ash":
			weatherType = particles.WeatherAsh
		case "neonrain":
			weatherType = particles.WeatherNeonRain
		case "smog":
			weatherType = particles.WeatherSmog
		case "radiation":
			weatherType = particles.WeatherRadiation
		default:
			weatherType = particles.WeatherRain
		}
	}

	// Create weather configuration
	config := particles.WeatherConfig{
		Type:      weatherType,
		Intensity: intensity,
		Width:     screenWidth * 2,  // Cover larger area than screen for smooth edges
		Height:    screenHeight * 2, // Cover larger area than screen
		GenreID:   genreID,
		Seed:      seed,
		WindX:     (rng.Float64() - 0.5) * 20.0, // Random wind: -10 to +10 px/s
		WindY:     0.0,                          // No vertical wind
		Custom:    make(map[string]interface{}),
	}

	// Create weather entity
	weatherEntity := world.CreateEntity()

	// Add weather component
	weatherComp := engine.NewWeatherComponent(config)

	// Start weather immediately with fade-in
	if err := weatherComp.StartWeather(); err != nil {
		// Log error but don't fail - weather is optional
		return nil
	}

	weatherEntity.AddComponent(weatherComp)

	return weatherEntity
}

// spawnDestructibleObjects spawns destructible objects (crates, barrels, furniture) in dungeon rooms.
// Phase 11.3: Environmental Destruction & Manipulation
// objectConfig defines spawn probabilities and counts for destructible objects per genre.
type objectConfig struct {
	crateChance           float64 // Probability of crate spawn per suitable room
	barrelChance          float64 // Probability of barrel spawn per suitable room
	furnitureChance       float64 // Probability of furniture spawn per suitable room
	explosiveBarrelChance float64 // Probability of explosive barrel (vs regular barrel)
	poisonContainerChance float64 // Probability of poison container spawn
	objectsPerRoom        int     // Number of objects to spawn in each eligible room
}

// getObjectConfig returns the spawning configuration for the given genre.
func getObjectConfig(genreID string) objectConfig {
	configs := map[string]objectConfig{
		"fantasy": {
			crateChance:           0.6,
			barrelChance:          0.5,
			furnitureChance:       0.4,
			explosiveBarrelChance: 0.1,
			poisonContainerChance: 0.05,
			objectsPerRoom:        3,
		},
		"scifi": {
			crateChance:           0.7,
			barrelChance:          0.6,
			furnitureChance:       0.3,
			explosiveBarrelChance: 0.15,
			poisonContainerChance: 0.1,
			objectsPerRoom:        4,
		},
		"horror": {
			crateChance:           0.4,
			barrelChance:          0.3,
			furnitureChance:       0.7,
			explosiveBarrelChance: 0.05,
			poisonContainerChance: 0.15,
			objectsPerRoom:        2,
		},
		"cyberpunk": {
			crateChance:           0.65,
			barrelChance:          0.55,
			furnitureChance:       0.5,
			explosiveBarrelChance: 0.2,
			poisonContainerChance: 0.08,
			objectsPerRoom:        3,
		},
		"postapocalyptic": {
			crateChance:           0.5,
			barrelChance:          0.7,
			furnitureChance:       0.6,
			explosiveBarrelChance: 0.12,
			poisonContainerChance: 0.1,
			objectsPerRoom:        4,
		},
	}

	// Default configuration if genre not found
	config := configs["fantasy"]
	if genreConfig, ok := configs[genreID]; ok {
		config = genreConfig
	}
	return config
}

// selectObjectType randomly selects an object type based on configuration probabilities.
func selectObjectType(rng *rand.Rand, config objectConfig) (engine.ObjectType, bool) {
	roll := rng.Float64()

	if roll < config.crateChance {
		return engine.ObjectCrate, true
	} else if roll < config.crateChance+config.barrelChance {
		// Decide between regular and explosive barrel
		if rng.Float64() < config.explosiveBarrelChance {
			return engine.ObjectExplosiveBarrel, true
		}
		return engine.ObjectBarrel, true
	} else if roll < config.crateChance+config.barrelChance+config.furnitureChance {
		return engine.ObjectFurniture, true
	} else if roll < config.crateChance+config.barrelChance+config.furnitureChance+config.poisonContainerChance {
		return engine.ObjectPoisonContainer, true
	}

	return engine.ObjectCrate, false // No spawn
}

// findValidSpawnLocation attempts to find a valid spawn location within the given room.
func findValidSpawnLocation(room *terrain.Room, terrainMap *terrain.Terrain, world *engine.World, rng *rand.Rand, tileSize int) (float64, float64, bool) {
	maxAttempts := 10
	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Random position within room (leave 1-tile border)
		tx := room.X + 1 + rng.Intn(room.Width-2)
		ty := room.Y + 1 + rng.Intn(room.Height-2)

		// Check if tile is walkable
		tile := terrainMap.GetTile(tx, ty)
		if tile == terrain.TileWall || tile == terrain.TileWallNE ||
			tile == terrain.TileWallNW || tile == terrain.TileWallSE ||
			tile == terrain.TileWallSW {
			continue
		}

		// Convert tile coordinates to world coordinates (center of tile)
		worldX := float64(tx*tileSize + tileSize/2)
		worldY := float64(ty*tileSize + tileSize/2)

		// Check if position is clear of other entities
		entities := world.GetEntities()
		tooClose := false
		for _, entity := range entities {
			if posComp, ok := entity.GetComponent("position"); ok {
				pos := posComp.(*engine.PositionComponent)
				dx := pos.X - worldX
				dy := pos.Y - worldY
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < float64(tileSize) { // Objects must be at least 1 tile apart
					tooClose = true
					break
				}
			}
		}

		if !tooClose {
			return worldX, worldY, true
		}
	}
	return 0, 0, false
}

// createDestructibleObject creates and configures a destructible object entity.
func createDestructibleObject(world *engine.World, objectType engine.ObjectType, worldX, worldY float64, logger *logrus.Logger, roomIndex int) {
	objectEntity := world.CreateEntity()

	// Add position component
	posComp := &engine.PositionComponent{
		X: worldX,
		Y: worldY,
	}
	objectEntity.AddComponent(posComp)

	// Add destructible object component
	destructibleComp := engine.NewDestructibleObjectComponent(objectType)
	objectEntity.AddComponent(destructibleComp)

	// Add carriable component (lighter objects can be picked up)
	weight := 0.5 // Default medium weight
	if objectType == engine.ObjectCrate {
		weight = 0.3 // Light crate
	} else if objectType == engine.ObjectBarrel || objectType == engine.ObjectExplosiveBarrel {
		weight = 0.7 // Heavy barrel
	} else if objectType == engine.ObjectFurniture {
		weight = 0.6 // Medium furniture
	} else if objectType == engine.ObjectPoisonContainer {
		weight = 0.4 // Light container
	}
	carriableComp := engine.NewCarriableComponent(weight)
	objectEntity.AddComponent(carriableComp)

	// Add context action component for interaction prompts
	actionType := engine.ActionPickup
	actionText := "Pickup"
	if objectType == engine.ObjectCrate {
		actionText = "Open Crate"
		actionType = engine.ActionOpen
	}
	contextComp := engine.NewContextActionComponent(actionType, actionText)
	objectEntity.AddComponent(contextComp)

	if logger != nil && logger.GetLevel() >= logrus.DebugLevel {
		logger.WithFields(logrus.Fields{
			"objectType": objectType.String(),
			"x":          worldX,
			"y":          worldY,
			"roomIndex":  roomIndex,
		}).Debug("spawned destructible object")
	}
}

func spawnDestructibleObjects(world *engine.World, terrainMap *terrain.Terrain, seed int64, genreID string, logger *logrus.Logger) int {
	rng := rand.New(rand.NewSource(seed))
	objectCount := 0
	tileSize := 32

	config := getObjectConfig(genreID)

	// Iterate through rooms (skip entrance room at index 0)
	for i := 1; i < len(terrainMap.Rooms); i++ {
		room := terrainMap.Rooms[i]

		// Skip very small rooms (< 4x4 tiles)
		if room.Width < 4 || room.Height < 4 {
			continue
		}

		// Spawn objects in this room based on configuration
		for objectsSpawned := 0; objectsSpawned < config.objectsPerRoom; objectsSpawned++ {
			// Randomly select object type
			objectType, shouldSpawn := selectObjectType(rng, config)
			if !shouldSpawn {
				break // Move to next room
			}

			// Find a valid spawn location within room
			worldX, worldY, found := findValidSpawnLocation(room, terrainMap, world, rng, tileSize)
			if !found {
				break // Couldn't find valid location, stop trying for this room
			}

			// Create and configure the object entity
			createDestructibleObject(world, objectType, worldX, worldY, logger, i)
			objectCount++
		}
	}

	return objectCount
}

// addStarterItems generates and adds starting items to the player's inventory.
// generateStarterWeapon creates a rusty starter weapon for new players.
func generateStarterWeapon(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
	weaponParams := procgen.GenerationParams{
		Difficulty: 0.0, // Easy starter weapon
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "weapon",
		},
	}

	weaponResult, err := itemGen.Generate(seed+1, weaponParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter weapon")
		return
	}

	weapons := weaponResult.([]*item.Item)
	if len(weapons) > 0 {
		weapon := weapons[0]
		weapon.Name = "Rusty " + weapon.Name // Make it clearly a starter item
		weapon.Stats.Value = 5               // Low value
		inventory.Items = append(inventory.Items, weapon)
		if logger.Logger.GetLevel() >= logrus.InfoLevel {
			logger.WithFields(logrus.Fields{
				"weaponName": weapon.Name,
				"damage":     weapon.Stats.Damage,
			}).Info("added starter weapon")
		}
	}
}

// generateStarterPotions creates minor healing potions for new players.
func generateStarterPotions(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
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
		logger.WithError(err).Warn("failed to generate healing potions")
		return
	}

	potions := potionResult.([]*item.Item)
	for _, potion := range potions {
		potion.Name = "Minor Health Potion"
		potion.Stats.Value = 10
		potion.Stats.Weight = 0.2
		inventory.Items = append(inventory.Items, potion)
	}
	if logger.Logger.GetLevel() >= logrus.InfoLevel && len(potions) > 0 {
		logger.WithField("count", len(potions)).Info("added healing potions")
	}
}

// generateStarterArmor creates worn starter armor for new players.
func generateStarterArmor(inventory *engine.InventoryComponent, itemGen *item.ItemGenerator, seed int64, genreID string, logger *logrus.Entry) {
	armorParams := procgen.GenerationParams{
		Difficulty: 0.0,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 1,
			"type":  "armor",
		},
	}

	armorResult, err := itemGen.Generate(seed+100, armorParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter armor")
		return
	}

	armors := armorResult.([]*item.Item)
	if len(armors) > 0 {
		armor := armors[0]
		armor.Name = "Worn " + armor.Name
		armor.Stats.Value = 8
		inventory.Items = append(inventory.Items, armor)
		if logger.Logger.GetLevel() >= logrus.InfoLevel {
			logger.WithFields(logrus.Fields{
				"armorName": armor.Name,
				"defense":   armor.Stats.Defense,
			}).Info("added starter armor")
		}
	}
}

func addStarterItems(inventory *engine.InventoryComponent, seed int64, genreID string, logger *logrus.Logger) {
	itemGen := item.NewItemGenerator()
	itemLogger := logging.GeneratorLogger(logger, "item", seed, genreID)

	// Generate starting equipment
	generateStarterWeapon(inventory, itemGen, seed, genreID, itemLogger)
	generateStarterPotions(inventory, itemGen, seed, genreID, itemLogger)
	generateStarterArmor(inventory, itemGen, seed, genreID, itemLogger)

	if logger.GetLevel() >= logrus.InfoLevel {
		itemLogger.WithField("itemCount", len(inventory.Items)).Info("starter items added")
	}
}

// addTutorialQuest creates and adds a tutorial quest to the player's quest tracker.
func addTutorialQuest(tracker *engine.QuestTrackerComponent, seed int64, genreID string, logger *logrus.Logger) {
	// Create a simple tutorial quest manually (more reliable than generation)
	tutorialQuest := &quest.Quest{
		ID:            fmt.Sprintf("tutorial_%d", seed),
		Name:          "Welcome to Venture",
		Type:          quest.TypeExplore,
		Difficulty:    quest.DifficultyTrivial,
		Description:   "Learn the basics of survival in this procedurally generated world. Explore your surroundings, manage your inventory, and prepare for adventure!",
		RequiredLevel: 1,
		Status:        quest.StatusActive,
		Seed:          seed,
		Tags:          []string{"tutorial", "starter"},
		GiverNPC:      "System",
		Objectives: []quest.Objective{
			{
				Description: "Open your inventory (press I)",
				Target:      "inventory",
				Required:    1,
				Current:     0,
			},
			{
				Description: "Check your quest log (press J)",
				Target:      "questlog",
				Required:    1,
				Current:     1, // Auto-complete since they're viewing it now!
			},
			{
				Description: "Explore the dungeon (move with WASD)",
				Target:      "explore",
				Required:    10, // Move 10 tiles
				Current:     0,
			},
		},
		Reward: quest.Reward{
			XP:          50,
			Gold:        25,
			Items:       []string{},
			SkillPoints: 0,
		},
	}

	// Accept the quest
	tracker.AcceptQuest(tutorialQuest, 0)

	if logger.GetLevel() >= logrus.InfoLevel {
		logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
			"questName":      tutorialQuest.Name,
			"objectiveCount": len(tutorialQuest.Objectives),
		}).Info("tutorial quest added")
	}
}

// startEmbeddedServer starts a server in a background goroutine for host-and-play mode
// Design: Uses ServerManager from pkg/hostplay for lifecycle management
// Why: Reuses server implementation with proper resource management
//
// Returns: (serverAddress, cleanupFunction, error)
func startEmbeddedServer(logger *logrus.Logger, seed int64, genreID string) (string, func(), error) {
	serverLogger := logger.WithFields(logrus.Fields{
		"component": "embedded-server",
		"seed":      seed,
		"genre":     genreID,
	})

	serverLogger.Info("starting server in background")

	// Create server configuration
	serverConfig := &hostplay.ServerConfig{
		Port:       *serverPort,
		MaxPlayers: *serverPlayers,
		BindLAN:    *hostLAN,
		WorldSeed:  seed,
		GenreID:    genreID,
		Difficulty: 0.5,
		TickRate:   *serverTick,
	}

	// Create server manager
	manager, err := hostplay.NewServerManager(serverConfig, logger)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create server manager: %w", err)
	}

	// Start the server (blocks until listening or error)
	if err := manager.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start server: %w", err)
	}

	serverAddr := manager.Address()
	port := manager.Port()

	if *hostLAN {
		serverLogger.WithField("bindAddr", "0.0.0.0").Warn("server accessible on LAN - firewall may block connections")

		// Try to get LAN IP for display
		if lanAddr := manager.GetLANAddress(); lanAddr != "" {
			serverLogger.WithField("lanAddress", lanAddr).Info("LAN players can connect to this address")
		}
	} else {
		serverLogger.WithField("bindAddr", "127.0.0.1").Info("server bound to localhost only")
	}

	serverLogger.WithFields(logrus.Fields{
		"address":    serverAddr,
		"port":       port,
		"maxPlayers": *serverPlayers,
		"tickRate":   *serverTick,
	}).Info("server ready for connections")

	// Return cleanup function
	cleanup := func() {
		serverLogger.Info("initiating graceful shutdown")
		if err := manager.Stop(); err != nil {
			serverLogger.WithError(err).Error("shutdown error")
		}
	}

	return serverAddr, cleanup, nil
}

// dropInventoryItems drops all items from an entity's inventory with scatter physics.
// Returns true if items were dropped.
func dropInventoryItems(
	game *engine.EbitenGame,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
) {
	invComp, hasInv := enemy.GetComponent("inventory")
	if !hasInv {
		return
	}
	inventory := invComp.(*engine.InventoryComponent)

	// Spawn each item in the inventory with scatter physics
	for i, itm := range inventory.Items {
		if itm == nil {
			continue
		}

		// Calculate scatter offset using circular distribution
		angle := float64(i) * 6.28318 / float64(len(inventory.Items)) // 2*PI radians
		scatterDist := 20.0 + float64(i)*5.0                          // Items spread 20-50 pixels out
		offsetX := scatterDist * math.Cos(angle)
		offsetY := scatterDist * math.Sin(angle)

		// Spawn item entity at scattered position
		itemEntity := engine.SpawnItemInWorld(game.World, itm, pos.X+offsetX, pos.Y+offsetY)
		if itemEntity != nil {
			// Add physics velocity for scatter effect (items fly outward then slow down)
			velocityX := offsetX * 3.0 // Initial velocity proportional to offset
			velocityY := offsetY * 3.0
			itemEntity.AddComponent(&engine.VelocityComponent{
				VX: velocityX,
				VY: velocityY,
			})

			// Add friction to slow down items over time
			itemEntity.AddComponent(engine.NewFrictionComponent(0.12)) // 12% friction per frame (at 60 FPS)

			// Track dropped item in DeadComponent
			deadComp.AddDroppedItem(itemEntity.ID)
		}
	}

	// Clear inventory after dropping all items
	inventory.Clear()
}

// dropEquippedItems drops all equipped items from an entity with scatter physics.
func dropEquippedItems(
	game *engine.EbitenGame,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
) {
	equipComp, hasEquip := enemy.GetComponent("equipment")
	if !hasEquip {
		return
	}
	equipment := equipComp.(*engine.EquipmentComponent)
	equippedItems := equipment.UnequipAll()

	// Spawn equipped items with additional scatter
	for i, itm := range equippedItems {
		if itm == nil {
			continue
		}

		// Use different angle range for equipped items (opposite side)
		angle := (float64(i) * 6.28318 / float64(len(equippedItems))) + 3.14159 // Offset by PI
		scatterDist := 30.0 + float64(i)*5.0
		offsetX := scatterDist * math.Cos(angle)
		offsetY := scatterDist * math.Sin(angle)

		itemEntity := engine.SpawnItemInWorld(game.World, itm, pos.X+offsetX, pos.Y+offsetY)
		if itemEntity != nil {
			velocityX := offsetX * 3.0
			velocityY := offsetY * 3.0
			itemEntity.AddComponent(&engine.VelocityComponent{
				VX: velocityX,
				VY: velocityY,
			})

			// Add friction for smooth deceleration
			itemEntity.AddComponent(engine.NewFrictionComponent(0.12))

			deadComp.AddDroppedItem(itemEntity.ID)
		}
	}
}

// spawnProceduralLoot generates and spawns procedural loot drops for non-player entities.
func spawnProceduralLoot(
	game *engine.EbitenGame,
	enemy *engine.Entity,
	pos *engine.PositionComponent,
	deadComp *engine.DeadComponent,
	recipeGen *recipe.RecipeGenerator,
	seed int64,
	genreID string,
	playerEntity *engine.Entity,
	objectiveTracker *engine.ObjectiveTrackerSystem,
) {
	// Only for NPCs/enemies, not players
	if enemy.HasComponent("input") {
		return
	}

	lootEntity := engine.GenerateLootDrop(game.World, enemy, pos.X, pos.Y, seed, genreID)
	if lootEntity != nil {
		// Add physics to procedural loot too
		lootEntity.AddComponent(&engine.VelocityComponent{
			VX: (rand.Float64()*2.0 - 1.0) * 30.0, // Random velocity -30 to +30
			VY: (rand.Float64()*2.0 - 1.0) * 30.0,
		})
		// Add friction for smooth deceleration
		lootEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(lootEntity.ID)
	}

	// Generate and spawn recipe drops (rarer than item drops)
	recipeEntity := engine.GenerateRecipeDrop(recipeGen, game.World, enemy, pos.X, pos.Y, seed, genreID)
	if recipeEntity != nil {
		// Add physics to recipe drops
		recipeEntity.AddComponent(&engine.VelocityComponent{
			VX: (rand.Float64()*2.0 - 1.0) * 25.0, // Slightly slower velocity for recipes
			VY: (rand.Float64()*2.0 - 1.0) * 25.0,
		})
		// Add friction for smooth deceleration
		recipeEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(recipeEntity.ID)
	}

	// Track enemy kill for quest objectives
	if playerEntity != nil {
		objectiveTracker.OnEnemyKilled(playerEntity, enemy)
	}
}

// createDeathCallback creates the death callback function for the combat system.
// This callback handles entity death by:
// - Dropping inventory and equipped items with scatter physics
// - Generating procedural loot drops for non-player entities
// - Playing death sound effects
// - Tracking quest objectives
func createDeathCallback(
	game *engine.EbitenGame,
	playerEntity **engine.Entity,
	objectiveTracker *engine.ObjectiveTrackerSystem,
	audioManager **engine.AudioManager,
	recipeGen *recipe.RecipeGenerator,
	seed int64,
	genreID string,
	logger *logrus.Logger,
) func(*engine.Entity) {
	return func(enemy *engine.Entity) {
		// Priority 1.4: Only process death once (callback called every frame while entity is dead)
		if enemy.HasComponent("dead") {
			return
		}

		// Get enemy position
		posComp, hasPos := enemy.GetComponent("position")
		if !hasPos {
			return
		}
		pos := posComp.(*engine.PositionComponent)

		// Priority 1.4: Add DeadComponent to mark entity as dead
		gameTime := float64(time.Now().Unix()) // Use game time if available
		deadComp := engine.NewDeadComponent(gameTime)
		enemy.AddComponent(deadComp)

		// Priority 1.4: Drop all items from entity's inventory
		dropInventoryItems(game, enemy, pos, deadComp)

		// Priority 1.4: Also drop equipped items
		dropEquippedItems(game, enemy, pos, deadComp)

		// Generate and spawn procedural loot drop (in addition to inventory items)
		spawnProceduralLoot(game, enemy, pos, deadComp, recipeGen, seed, genreID, *playerEntity, objectiveTracker)

		// GAP-010 REPAIR: Play death sound effect
		if *audioManager != nil {
			if err := (*audioManager).PlaySFX("death", time.Now().UnixNano()); err != nil {
				if logger.GetLevel() >= logrus.WarnLevel {
					logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to play death SFX")
				}
			}
		}
	}
}
