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
	"github.com/opd-ai/venture/pkg/procgen/book"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/story"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/saveload"
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

// V4.0 System Wrappers (Phase 21-27)

// companionAISystemWrapper adapts CompanionAISystem to System interface
type companionAISystemWrapper struct {
	system *engine.CompanionAISystem
}

func (w *companionAISystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionProgressionSystemWrapper adapts CompanionProgressionSystem to System interface
type companionProgressionSystemWrapper struct {
	system *engine.CompanionProgressionSystem
}

func (w *companionProgressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionLoyaltySystemWrapper adapts CompanionLoyaltySystem to System interface
type companionLoyaltySystemWrapper struct {
	system *engine.CompanionLoyaltySystem
}

func (w *companionLoyaltySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// companionInventorySystemWrapper adapts CompanionInventorySystem to System interface
type companionInventorySystemWrapper struct {
	system *engine.CompanionInventorySystem
}

func (w *companionInventorySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// skillInheritanceSystemWrapper adapts SkillInheritanceSystem to System interface
type skillInheritanceSystemWrapper struct {
	system *engine.SkillInheritanceSystem
}

func (w *skillInheritanceSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// expressionSystemWrapper adapts ExpressionSystem to System interface
type expressionSystemWrapper struct {
	system *engine.ExpressionSystem
}

func (w *expressionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// expressionComboSystemWrapper adapts ExpressionComboSystem to System interface
type expressionComboSystemWrapper struct {
	system *engine.ExpressionComboSystem
}

func (w *expressionComboSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// miniGameSystemWrapper adapts MiniGameSystem to System interface
type miniGameSystemWrapper struct {
	system *engine.MiniGameSystem
}

func (w *miniGameSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// achievementSystemWrapper adapts AchievementSystem to System interface
type achievementSystemWrapper struct {
	system *engine.AchievementSystem
}

func (w *achievementSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// V5.0 System Wrappers (Social & Communication)

type enhancedChatSystemWrapper struct {
	system *engine.EnhancedChatSystem
}

func (w *enhancedChatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type chatSystemWrapper struct {
	system *engine.EnhancedChatSystem
}

func (w *chatSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type mailSystemWrapper struct {
	system *engine.MailSystem
}

func (w *mailSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type courierSystemWrapper struct {
	system *engine.CourierSystem
}

func (w *courierSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// INTEGRATION FIX [Category A]: System Wrapper Types
// Gap: New systems need adapter wrappers to match World.System interface
// Fix: Added wrapper types for all newly integrated systems
// Roadmap: ROADMAP_V4.md (Phase 14, 30-31) and ROADMAP_V5.md (Phase 32-36)

type investigationSystemWrapper struct {
	system *engine.InvestigationSystem
}

func (w *investigationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type npcDialogSystemWrapper struct {
	system *engine.NPCDialogSystem
}

func (w *npcDialogSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type musicTriggerSystemWrapper struct {
	system *engine.MusicTriggerSystem
}

func (w *musicTriggerSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type positionalAudioSystemWrapper struct {
	system *engine.PositionalAudioSystem
}

func (w *positionalAudioSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type reverbSystemWrapper struct {
	system *engine.ReverbSystem
}

func (w *reverbSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type qualitySystemWrapper struct {
	system *engine.QualitySystem
}

func (w *qualitySystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type tradeSystemWrapper struct {
	system *engine.TradeSystem
}

func (w *tradeSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type terrainConstructionSystemWrapper struct {
	system *engine.TerrainConstructionSystem
}

func (w *terrainConstructionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type terrainModificationSystemWrapper struct {
	system *engine.TerrainModificationSystem
}

func (w *terrainModificationSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(entities, deltaTime)
}

type merchantCaravanSystemWrapper struct {
	system *engine.MerchantCaravanSystem
}

func (w *merchantCaravanSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// INTEGRATION FIX [Category A]: High-Level Management System Wrappers
// Gap: CompanionSystem, VehicleSystem, AdaptiveSoundtrackSystem wrappers missing
// Fix: Added wrapper types for high-level management systems
// Roadmap: ROADMAP_V4.md Phase 21.2, 22.2, 29

type companionSystemWrapper struct {
	system *engine.CompanionSystem
}

func (w *companionSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type vehicleSystemWrapper struct {
	system *engine.VehicleSystem
}

func (w *vehicleSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

type adaptiveSoundtrackSystemWrapper struct {
	system *engine.AdaptiveSoundtrackSystem
}

func (w *adaptiveSoundtrackSystemWrapper) Update(entities []*engine.Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}

// Note: fluidSimulatorWrapper is defined in handlers.go along with other V8.0 system wrappers

var (
	width            = flag.Int("width", 1920, "Screen width (1280, 1920, 2560, 3840)")
	height           = flag.Int("height", 1080, "Screen height (720, 1080, 1440, 2160)")
	fullscreen       = flag.Bool("fullscreen", false, "Start in fullscreen mode")
	seed             = flag.Int64("seed", seededRandom(), "World generation seed")
	genreID          = flag.String("genre", randomGenre(), "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)")
	enableLighting   = flag.Bool("enable-lighting", true, "Enable dynamic lighting system")
	enableWeather    = flag.Bool("enable-weather", true, "Enable procedural weather effects")
	weatherType      = flag.String("weather", "", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation) - empty for genre-appropriate random")
	weatherIntensity = flag.String("weather-intensity", "medium", "Weather intensity (light, medium, heavy, extreme)")
	verbose          = flag.Bool("verbose", false, "Enable verbose logging")
	profile          = flag.Bool("profile", false, "Enable performance profiling with frame time tracking")
	multiplayer      = flag.Bool("multiplayer", false, "Connect to remote multiplayer server (default: starts local server)")
	server           = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
	hostAndPlay      = flag.Bool("host-and-play", false, "Explicitly enable host-and-play mode (default behavior when --multiplayer not specified)")
	hostLAN          = flag.Bool("host-lan", false, "Bind server to 0.0.0.0 for LAN access instead of 127.0.0.1 (requires host-and-play mode)")
	serverPort       = flag.Int("port", 8080, "Server port for --host-and-play mode (will try next 10 ports if occupied)")
	serverPlayers    = flag.Int("max-players", 4, "Maximum players for --host-and-play mode")
	serverTick       = flag.Int("tick-rate", 20, "Server tick rate for --host-and-play mode (updates per second)")
	noTutorial       = flag.Bool("no-tutorial", false, "Disable tutorial for experienced players")
	enableHousing    = flag.Bool("enable-housing", true, "Enable player housing system (V8.0)")
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

// initializeNetworkClient sets up network connection for multiplayer.
// With auto-enable logic, host-and-play always sets *multiplayer=true before this is called,
// so this function always creates a network client (either to embedded or remote server).
func initializeNetworkClient(logger *logrus.Logger, clientLogger *logrus.Entry) network.ClientConnection {
	clientLogger.WithField("server", *server).Info("connecting to server")

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

// serializePlayerState extracts all player state for saving.
func serializePlayerState(player *engine.Entity, game *engine.EbitenGame) *saveload.PlayerState {
	playerState := &saveload.PlayerState{EntityID: player.ID}

	serializePosition(player, playerState)
	serializeHealth(player, playerState)
	serializeStats(player, playerState)
	serializeExperience(player, playerState)
	serializeInventory(player, playerState)
	serializeEquipment(player, playerState)
	serializeManaAndSpells(player, playerState)
	serializeTutorialState(game, playerState)

	playerState.Speed = 1.0
	return playerState
}

// serializePosition extracts player position to state.
func serializePosition(player *engine.Entity, state *saveload.PlayerState) {
	if posComp, ok := player.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		state.X, state.Y = pos.X, pos.Y
	}
}

// serializeHealth extracts player health to state.
func serializeHealth(player *engine.Entity, state *saveload.PlayerState) {
	if healthComp, ok := player.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		state.CurrentHealth, state.MaxHealth = health.Current, health.Max
	}
}

// serializeStats extracts player stats to state.
func serializeStats(player *engine.Entity, state *saveload.PlayerState) {
	if statsComp, ok := player.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		state.Attack = stats.Attack
		state.Defense = stats.Defense
		state.MagicPower = stats.MagicPower
	}
}

// serializeExperience extracts player level and XP to state.
func serializeExperience(player *engine.Entity, state *saveload.PlayerState) {
	if expComp, ok := player.GetComponent("experience"); ok {
		exp := expComp.(*engine.ExperienceComponent)
		state.Level = exp.Level
		state.Experience = exp.CurrentXP
	}
}

// serializeInventory extracts inventory data to state.
func serializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv := invComp.(*engine.InventoryComponent)
		state.Gold = inv.Gold
		state.Items = make([]saveload.ItemData, 0, len(inv.Items))
		for _, itm := range inv.Items {
			state.Items = append(state.Items, saveload.ItemToData(itm))
		}
	}
}

// serializeEquipment extracts equipped items to state.
func serializeEquipment(player *engine.Entity, state *saveload.PlayerState) {
	if equip, hasEquip := player.GetComponent("equipment"); hasEquip {
		equipment := equip.(*engine.EquipmentComponent)
		if weapon := equipment.Slots[engine.SlotMainHand]; weapon != nil {
			weaponData := saveload.ItemToData(weapon)
			state.EquippedItems.Weapon = &weaponData
		}
		if armor := equipment.Slots[engine.SlotChest]; armor != nil {
			armorData := saveload.ItemToData(armor)
			state.EquippedItems.Armor = &armorData
		}
		if accessory := equipment.Slots[engine.SlotAccessory1]; accessory != nil {
			accessoryData := saveload.ItemToData(accessory)
			state.EquippedItems.Accessory = &accessoryData
		}
	}
}

// serializeManaAndSpells extracts mana and spell data to state.
func serializeManaAndSpells(player *engine.Entity, state *saveload.PlayerState) {
	if manaComp, hasMana := player.GetComponent("mana"); hasMana {
		mana := manaComp.(*engine.ManaComponent)
		state.CurrentMana = mana.Current
		state.MaxMana = mana.Max
	}

	if slotsComp, hasSlots := player.GetComponent("spell_slots"); hasSlots {
		slots := slotsComp.(*engine.SpellSlotComponent)
		state.Spells = make([]saveload.SpellData, 0, 5)
		for i := 0; i < 5; i++ {
			if spell := slots.GetSlot(i); spell != nil {
				state.Spells = append(state.Spells, saveload.SpellToData(spell))
			}
		}
	}
}

// serializeTutorialState extracts tutorial state to state.
func serializeTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.TutorialSystem != nil {
		enabled, showUI, currentStep, completed := game.TutorialSystem.ExportState()
		state.TutorialState = &saveload.TutorialStateData{
			Enabled:        enabled,
			ShowUI:         showUI,
			CurrentStepIdx: currentStep,
			CompletedSteps: completed,
		}
	}
}

// deserializePlayerState restores all player state from a save.
func deserializePlayerState(player *engine.Entity, playerState *saveload.PlayerState, game *engine.EbitenGame) {
	deserializePosition(player, playerState)
	deserializeHealth(player, playerState)
	deserializeStats(player, playerState)
	deserializeExperience(player, playerState)
	deserializeInventory(player, playerState)
	deserializeEquipment(player, playerState)
	deserializeManaAndSpells(player, playerState)
	deserializeTutorialState(game, playerState)
}

// deserializePosition restores player position from state.
func deserializePosition(player *engine.Entity, state *saveload.PlayerState) {
	if posComp, ok := player.GetComponent("position"); ok {
		pos := posComp.(*engine.PositionComponent)
		pos.X, pos.Y = state.X, state.Y
	}
}

// deserializeHealth restores player health from state.
func deserializeHealth(player *engine.Entity, state *saveload.PlayerState) {
	if healthComp, ok := player.GetComponent("health"); ok {
		health := healthComp.(*engine.HealthComponent)
		health.Current, health.Max = state.CurrentHealth, state.MaxHealth
	}
}

// deserializeStats restores player stats from state.
func deserializeStats(player *engine.Entity, state *saveload.PlayerState) {
	if statsComp, ok := player.GetComponent("stats"); ok {
		stats := statsComp.(*engine.StatsComponent)
		stats.Attack = state.Attack
		stats.Defense = state.Defense
		stats.MagicPower = state.MagicPower
	}
}

// deserializeExperience restores player level and XP from state.
func deserializeExperience(player *engine.Entity, state *saveload.PlayerState) {
	if expComp, ok := player.GetComponent("experience"); ok {
		exp := expComp.(*engine.ExperienceComponent)
		exp.Level = state.Level
		exp.CurrentXP = state.Experience
	}
}

// deserializeInventory restores inventory from state.
func deserializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv := invComp.(*engine.InventoryComponent)
		inv.Items = make([]*item.Item, 0, len(state.Items))
		for _, itemData := range state.Items {
			restoredItem := saveload.DataToItem(itemData)
			inv.Items = append(inv.Items, restoredItem)
		}
		inv.Gold = state.Gold
	}
}

// deserializeEquipment restores equipped items from state.
func deserializeEquipment(player *engine.Entity, state *saveload.PlayerState) {
	if equipComp, ok := player.GetComponent("equipment"); ok {
		equipment := equipComp.(*engine.EquipmentComponent)
		equipment.Slots = make(map[engine.EquipmentSlot]*item.Item)

		if state.EquippedItems.Weapon != nil {
			weapon := saveload.DataToItem(*state.EquippedItems.Weapon)
			equipment.Slots[engine.SlotMainHand] = weapon
		}
		if state.EquippedItems.Armor != nil {
			armor := saveload.DataToItem(*state.EquippedItems.Armor)
			equipment.Slots[engine.SlotChest] = armor
		}
		if state.EquippedItems.Accessory != nil {
			accessory := saveload.DataToItem(*state.EquippedItems.Accessory)
			equipment.Slots[engine.SlotAccessory1] = accessory
		}
		equipment.StatsDirty = true
	}
}

// deserializeManaAndSpells restores mana and spells from state.
func deserializeManaAndSpells(player *engine.Entity, state *saveload.PlayerState) {
	if manaComp, ok := player.GetComponent("mana"); ok {
		mana := manaComp.(*engine.ManaComponent)
		mana.Current, mana.Max = state.CurrentMana, state.MaxMana
	}

	if slotsComp, ok := player.GetComponent("spell_slots"); ok {
		slots := slotsComp.(*engine.SpellSlotComponent)
		for i := 0; i < 5; i++ {
			slots.Slots[i] = nil
		}
		for i, spellData := range state.Spells {
			if i < 5 {
				restoredSpell := saveload.DataToSpell(spellData)
				slots.SetSlot(i, restoredSpell)
			}
		}
	}
}

// deserializeTutorialState restores tutorial state from state.
func deserializeTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.TutorialSystem != nil && state.TutorialState != nil {
		tutState := state.TutorialState
		game.TutorialSystem.ImportState(
			tutState.Enabled,
			tutState.ShowUI,
			tutState.CurrentStepIdx,
			tutState.CompletedSteps,
		)
	}
}

// createGameSave creates a complete game save from current game state.
func createGameSave(player *engine.Entity, game *engine.EbitenGame, generatedTerrain *terrain.Terrain) *saveload.GameSave {
	playerState := serializePlayerState(player, game)

	// Get fog of war data
	var fogOfWar [][]bool
	if game.MapUI != nil {
		fogOfWar = game.MapUI.GetFogOfWar()
	}

	return &saveload.GameSave{
		Version:     saveload.SaveVersion,
		PlayerState: playerState,
		WorldState: &saveload.WorldState{
			Seed:       *seed,
			GenreID:    *genreID,
			Width:      generatedTerrain.Width,
			Height:     generatedTerrain.Height,
			Difficulty: 0.5,
			Depth:      1,
			FogOfWar:   fogOfWar,
		},
		Settings: &saveload.GameSettings{
			ScreenWidth:  *width,
			ScreenHeight: *height,
			Fullscreen:   false,
			VSync:        true,
			MasterVolume: 1.0,
			MusicVolume:  0.7,
			SFXVolume:    0.8,
			KeyBindings:  make(map[string]string),
		},
	}
}

// loadGameSave restores game state from a saved game.
func loadGameSave(player *engine.Entity, gameSave *saveload.GameSave, game *engine.EbitenGame) {
	deserializePlayerState(player, gameSave.PlayerState, game)

	// Restore fog of war
	if game.MapUI != nil && gameSave.WorldState != nil && gameSave.WorldState.FogOfWar != nil {
		game.MapUI.SetFogOfWar(gameSave.WorldState.FogOfWar)
	}
}

// spawnVehicles generates and spawns vehicles in terrain rooms (V4.0).
func spawnVehicles(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	// Import vehicle generator
	vehicleGen := vehicle.NewVehicleGenerator()

	// Determine number of vehicles based on room count (2-5 vehicles per level)
	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil // Need at least 2 rooms (entrance + 1 other)
	}

	vehicleCount := 2 + (roomCount-2)/4 // Start with 2, +1 per 4 additional rooms
	if vehicleCount > 5 {
		vehicleCount = 5 // Cap at 5 vehicles
	}

	// Generate vehicles
	vehicleResult, err := vehicleGen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate vehicles: %w", err)
	}

	vehicles, ok := vehicleResult.([]*vehicle.Vehicle)
	if !ok || len(vehicles) == 0 {
		return 0, fmt.Errorf("vehicle generator returned invalid result")
	}

	// Limit to vehicleCount
	if len(vehicles) > vehicleCount {
		vehicles = vehicles[:vehicleCount]
	}

	// Convert vehicles to VehicleSpawnData
	vehicleSpawnData := make([]engine.VehicleSpawnData, len(vehicles))
	for i, v := range vehicles {
		// Convert vehicle.VehicleType to engine.VehicleType
		var engineType engine.VehicleType
		switch v.VehicleType {
		case vehicle.TypeMount:
			engineType = engine.VehicleMount
		case vehicle.TypeCart:
			engineType = engine.VehicleCart
		case vehicle.TypeBoat:
			engineType = engine.VehicleBoat
		case vehicle.TypeGlider:
			engineType = engine.VehicleGlider
		case vehicle.TypeMech:
			engineType = engine.VehicleMech
		}

		// Convert uint32 color to color.RGBA
		r := uint8((v.Color >> 16) & 0xFF)
		g := uint8((v.Color >> 8) & 0xFF)
		b := uint8(v.Color & 0xFF)
		colorRGBA := color.RGBA{R: r, G: g, B: b, A: 255}

		// Determine sprite size based on vehicle type
		var size int
		var colliderSize float64
		switch v.VehicleType {
		case vehicle.TypeMount:
			size = 32
			colliderSize = 28.0
		case vehicle.TypeCart:
			size = 40
			colliderSize = 36.0
		case vehicle.TypeBoat:
			size = 48
			colliderSize = 44.0
		case vehicle.TypeGlider:
			size = 36
			colliderSize = 32.0
		case vehicle.TypeMech:
			size = 44
			colliderSize = 40.0
		default:
			size = 32
			colliderSize = 28.0
		}

		vehicleSpawnData[i] = engine.VehicleSpawnData{
			Name:         v.Name,
			VehicleType:  engineType,
			Components:   v.ToComponents(), // Generate components from vehicle
			Color:        colorRGBA,
			Size:         size,
			ColliderSize: colliderSize,
		}
	}

	// Spawn vehicles in terrain
	spawned, err := engine.SpawnVehiclesInTerrain(world, terrainMap, vehicleSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn vehicles in terrain: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"requested": vehicleCount,
		"generated": len(vehicles),
		"spawned":   spawned,
	}).Debug("vehicle spawning complete")

	return spawned, nil
}

// spawnCompanions generates and spawns companions in terrain rooms (V4.0).
func spawnCompanions(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	// Import companion generator
	companionGen := companion.NewGenerator()

	// Determine number of companions based on room count (1-3 companions per level)
	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil // Need at least 2 rooms (entrance + 1 other)
	}

	companionCount := 1 + (roomCount-2)/5 // Start with 1, +1 per 5 additional rooms
	if companionCount > 3 {
		companionCount = 3 // Cap at 3 companions
	}

	// Generate companions
	companionSpawnData := make([]engine.CompanionSpawnData, 0, companionCount)
	for i := 0; i < companionCount; i++ {
		companionSeed := seed + int64(i*1000) // Unique seed per companion

		companionResult, err := companionGen.Generate(companionSeed, params)
		if err != nil {
			logger.WithError(err).Warn("failed to generate companion")
			continue
		}

		comp, ok := companionResult.(*companion.Companion)
		if !ok {
			logger.Warn("companion generator returned invalid result")
			continue
		}

		// Convert companion.Type (engine.CompanionType) to spawn data
		// Generate color based on companion type and genre
		companionColor := generateCompanionColor(comp.Type, params.GenreID, companionSeed)

		// Determine sprite size based on companion type
		var size int
		var colliderSize float64
		switch comp.Type {
		case engine.CompanionTypePet:
			size = 24
			colliderSize = 20.0
		case engine.CompanionTypeSummon:
			size = 28
			colliderSize = 24.0
		case engine.CompanionTypeHireling:
			size = 28
			colliderSize = 24.0
		case engine.CompanionTypeElemental:
			size = 32
			colliderSize = 28.0
		case engine.CompanionTypeUndead:
			size = 30
			colliderSize = 26.0
		case engine.CompanionTypeRobot:
			size = 30
			colliderSize = 26.0
		case engine.CompanionTypeSpirit:
			size = 26
			colliderSize = 22.0
		case engine.CompanionTypeInsect:
			size = 22
			colliderSize = 18.0
		default:
			size = 28
			colliderSize = 24.0
		}

		companionSpawnData = append(companionSpawnData, engine.CompanionSpawnData{
			Name:          comp.Name,
			CompanionType: comp.Type,
			Level:         comp.Level,
			Attack:        comp.Attack,
			Defense:       comp.Defense,
			Speed:         comp.Speed,
			HP:            comp.HP,
			MaxHP:         comp.MaxHP,
			Loyalty:       comp.Loyalty,
			Commands:      comp.Commands,
			Color:         companionColor,
			Size:          size,
			ColliderSize:  colliderSize,
		})
	}

	if len(companionSpawnData) == 0 {
		return 0, nil // No companions generated
	}

	// Spawn companions in terrain
	spawned, err := engine.SpawnCompanionsInTerrain(world, terrainMap, companionSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn companions in terrain: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"requested": companionCount,
		"generated": len(companionSpawnData),
		"spawned":   spawned,
	}).Debug("companion spawning complete")

	return spawned, nil
}

// generateCompanionColor generates a color for a companion based on type and genre.
func generateCompanionColor(companionType engine.CompanionType, genreID string, seed int64) color.RGBA {
	rng := rand.New(rand.NewSource(seed))

	// Base colors per companion type
	var baseR, baseG, baseB uint8
	switch companionType {
	case engine.CompanionTypePet:
		baseR, baseG, baseB = 180, 140, 100 // Brown tones
	case engine.CompanionTypeSummon:
		baseR, baseG, baseB = 120, 100, 200 // Purple/magical
	case engine.CompanionTypeHireling:
		baseR, baseG, baseB = 160, 160, 140 // Earthy/armor tones
	case engine.CompanionTypeElemental:
		baseR, baseG, baseB = 100, 180, 220 // Elemental blue
	case engine.CompanionTypeUndead:
		baseR, baseG, baseB = 140, 160, 140 // Pale/decayed
	case engine.CompanionTypeRobot:
		baseR, baseG, baseB = 180, 180, 200 // Metallic
	case engine.CompanionTypeSpirit:
		baseR, baseG, baseB = 200, 200, 240 // Ethereal white/blue
	case engine.CompanionTypeInsect:
		baseR, baseG, baseB = 100, 160, 80 // Green/chitinous
	default:
		baseR, baseG, baseB = 150, 150, 150 // Gray default
	}

	// Add genre-based color variation
	genreVariation := func(base uint8, variance int) uint8 {
		offset := rng.Intn(variance*2+1) - variance
		result := int(base) + offset
		if result < 0 {
			result = 0
		}
		if result > 255 {
			result = 255
		}
		return uint8(result)
	}

	return color.RGBA{
		R: genreVariation(baseR, 30),
		G: genreVariation(baseG, 30),
		B: genreVariation(baseB, 30),
		A: 255,
	}
}

// spawnBookshelves generates and spawns bookshelves with books in terrain rooms (V4.0).
func spawnBookshelves(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	bookGen := book.NewGenerator()

	bookshelfCount := calculateBookshelfCount(terrainMap)
	if bookshelfCount == 0 {
		return 0, nil
	}

	bookshelfSpawnData := generateBookshelves(bookGen, bookshelfCount, seed, params, logger)
	if len(bookshelfSpawnData) == 0 {
		return 0, nil
	}

	spawned, err := engine.SpawnBookshelvesInTerrain(world, terrainMap, bookshelfSpawnData, seed)
	if err != nil {
		return 0, fmt.Errorf("failed to spawn bookshelves in terrain: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"requested": bookshelfCount,
		"generated": len(bookshelfSpawnData),
		"spawned":   spawned,
	}).Debug("bookshelf spawning complete")

	return spawned, nil
}

// calculateBookshelfCount determines number of bookshelves based on room count.
func calculateBookshelfCount(terrainMap *terrain.Terrain) int {
	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0
	}

	bookshelfCount := 1
	if roomCount >= 8 {
		bookshelfCount = 2
	}
	return bookshelfCount
}

// generateBookshelves creates bookshelf spawn data with generated books.
func generateBookshelves(bookGen *book.Generator, count int, seed int64, params procgen.GenerationParams, logger *logrus.Entry) []engine.BookshelfSpawnData {
	bookshelfSpawnData := make([]engine.BookshelfSpawnData, 0, count)

	for i := 0; i < count; i++ {
		bookshelfSeed := seed + int64(i*10000)
		books := generateBooksForShelf(bookGen, i, bookshelfSeed, params, logger)

		if len(books) == 0 {
			continue
		}

		shelfColor := generateBookshelfColor(params.GenreID, bookshelfSeed)
		bookshelfSpawnData = append(bookshelfSpawnData, engine.BookshelfSpawnData{
			Books:        books,
			ShelfColor:   shelfColor,
			ShelfSize:    32,
			ColliderSize: 28.0,
		})
	}

	return bookshelfSpawnData
}

// generateBooksForShelf creates books for a single bookshelf.
func generateBooksForShelf(bookGen *book.Generator, shelfIndex int, bookshelfSeed int64, params procgen.GenerationParams, logger *logrus.Entry) []*engine.BookComponent {
	booksPerShelf := 3 + (shelfIndex % 6)
	books := make([]*engine.BookComponent, 0, booksPerShelf)

	for j := 0; j < booksPerShelf; j++ {
		bookComp := generateSingleBook(bookGen, bookshelfSeed+int64(j*100), params, logger)
		if bookComp != nil {
			books = append(books, bookComp)
		}
	}

	return books
}

// generateSingleBook creates a single book with appropriate parameters.
func generateSingleBook(bookGen *book.Generator, bookSeed int64, params procgen.GenerationParams, logger *logrus.Entry) *engine.BookComponent {
	bookType := selectBookType(bookSeed)
	bookParams := createBookParams(params, bookType, bookSeed)

	bookResult, err := bookGen.Generate(bookSeed, bookParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate book")
		return nil
	}

	bookComp, ok := bookResult.(*engine.BookComponent)
	if !ok {
		logger.Warn("book generator returned invalid result")
		return nil
	}

	return bookComp
}

// selectBookType randomly selects a book type.
func selectBookType(bookSeed int64) engine.BookType {
	bookTypes := []engine.BookType{
		engine.BookTypeSkill,
		engine.BookTypeLore,
		engine.BookTypeRecipe,
		engine.BookTypeHistory,
	}
	rng := rand.New(rand.NewSource(bookSeed))
	return bookTypes[rng.Intn(len(bookTypes))]
}

// createBookParams creates generation parameters for a book.
func createBookParams(params procgen.GenerationParams, bookType engine.BookType, bookSeed int64) procgen.GenerationParams {
	bookParams := params
	bookParams.Custom = map[string]interface{}{
		"book_type": bookType,
	}

	if bookType == engine.BookTypeSkill {
		skills := []string{"Combat", "Magic", "Crafting", "Stealth", "Survival"}
		rng := rand.New(rand.NewSource(bookSeed))
		bookParams.Custom["skill_name"] = skills[rng.Intn(len(skills))]
		bookParams.Custom["skill_bonus"] = float64(params.Depth) * 0.1
	}

	return bookParams
}

// generateBookshelfColor generates a color for a bookshelf based on genre.
func generateBookshelfColor(genreID string, seed int64) color.RGBA {
	rng := rand.New(rand.NewSource(seed))

	// Base colors per genre
	var baseR, baseG, baseB uint8
	switch genreID {
	case "fantasy":
		baseR, baseG, baseB = 101, 67, 33 // Dark wood brown
	case "sci-fi":
		baseR, baseG, baseB = 120, 130, 140 // Metallic gray
	case "horror":
		baseR, baseG, baseB = 60, 50, 45 // Very dark brown
	case "cyberpunk":
		baseR, baseG, baseB = 80, 80, 100 // Dark bluish
	case "post-apocalyptic":
		baseR, baseG, baseB = 90, 85, 70 // Weathered brown
	default:
		baseR, baseG, baseB = 100, 80, 60 // Default wood
	}

	// Add random variation
	genreVariation := func(base uint8, variance int) uint8 {
		offset := rng.Intn(variance*2+1) - variance
		result := int(base) + offset
		if result < 0 {
			result = 0
		}
		if result > 255 {
			result = 255
		}
		return uint8(result)
	}

	return color.RGBA{
		R: genreVariation(baseR, 20),
		G: genreVariation(baseG, 20),
		B: genreVariation(baseB, 20),
		A: 255,
	}
}

// spawnStoryFragments generates and spawns story fragments in the terrain (Phase 30).
func spawnStoryFragments(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	storyGen := story.NewFragmentGenerator()

	// Generate story sequence
	storyResult, err := storyGen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate story: %w", err)
	}

	sequence, ok := storyResult.(*story.StorySequence)
	if !ok {
		return 0, fmt.Errorf("story generator returned invalid result type")
	}

	// Validate story quality
	if err := storyGen.Validate(sequence); err != nil {
		logger.WithError(err).Warn("generated story failed validation, using anyway")
	}

	// Register story series with discovery system if available
	// We'll set this after system initialization when it's available
	// For now, the discovery system will auto-detect series by counting fragments

	// Spawn fragment entities in terrain
	spawned := 0
	for _, fragment := range sequence.Fragments {
		// Create entity for fragment
		entity := world.CreateEntity()

		// Add position component (fragment.Location is already set)
		posComp := &engine.PositionComponent{
			X: fragment.Location.X,
			Y: fragment.Location.Y,
		}
		entity.AddComponent(posComp)

		// Add story fragment component
		fragComp := &engine.StoryFragmentComponent{
			Fragment:    fragment,
			Discovered:  false,
			SeriesID:    sequence.SeriesID,
			SequenceNum: fragment.SequenceNum,
		}
		entity.AddComponent(fragComp)

		// Add sprite component for visual representation
		spriteComp := &engine.EbitenSprite{
			Image:   nil, // Will be created by rendering system
			Width:   16,  // Small sprite for fragments
			Height:  16,
			Visible: true,
			Layer:   5, // Lower layer than player
		}
		entity.AddComponent(spriteComp)

		// Add collider for interaction detection
		colliderComp := &engine.ColliderComponent{
			Width:  16.0,
			Height: 16.0,
		}
		entity.AddComponent(colliderComp)

		spawned++
	}

	logger.WithFields(logrus.Fields{
		"seriesID":      sequence.SeriesID,
		"theme":         sequence.Theme,
		"fragmentCount": len(sequence.Fragments),
		"spawned":       spawned,
	}).Debug("spawned story fragments")

	return spawned, nil
}
