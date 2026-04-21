//go:build !android && !ios
// +build !android,!ios

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image/color"
	"io"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/config"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/qol"
	"github.com/opd-ai/venture/pkg/hostplay"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/network"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/book"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/environment"
	"github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/minigame"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/skills"
	"github.com/opd-ai/venture/pkg/procgen/story"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/sirupsen/logrus"
)

// Note: All system wrappers have been consolidated in system_wrappers.go for better code organization.

var (
	width            = flag.Int("width", 1920, "Screen width (1280, 1920, 2560, 3840)")
	height           = flag.Int("height", 1080, "Screen height (720, 1080, 1440, 2160)")
	fullscreen       = flag.Bool("fullscreen", false, "Start in fullscreen mode")
	seed             = flag.Int64("seed", seededRandom(), "World generation seed")
	genreID          = flag.String("genre", "random", "Genre ID (fantasy, scifi, horror, cyberpunk, postapoc, random)")
	weatherType      = flag.String("weather", "", "Weather type (rain, snow, fog, dust, ash, neonrain, smog, radiation) - empty for genre-appropriate random")
	weatherIntensity = flag.String("weather-intensity", "heavy", "Weather intensity (light, medium, heavy, extreme)")

	// Post-processing configuration (all enabled by default for maximum visual quality)
	postprocessPreset          = flag.String("postprocess-preset", "cinematic", "Post-processing preset (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic, neutral, cinematic)")
	postprocessColorGrading    = flag.Bool("postprocess-color-grading", true, "Enable color grading effect")
	postprocessVignette        = flag.Bool("postprocess-vignette", true, "Enable vignette effect")
	postprocessChromaticAber   = flag.Bool("postprocess-chromatic", true, "Enable chromatic aberration effect")
	postprocessSaturation      = flag.Float64("postprocess-saturation", 1.1, "Color grading saturation (0.0-2.0)")
	postprocessContrast        = flag.Float64("postprocess-contrast", 1.05, "Color grading contrast (0.0-2.0)")
	postprocessBrightness      = flag.Float64("postprocess-brightness", 0.02, "Color grading brightness (-1.0 to 1.0)")
	postprocessVignetteIntens  = flag.Float64("postprocess-vignette-intensity", 0.6, "Vignette intensity (0.0-1.0)")
	postprocessVignetteSoft    = flag.Float64("postprocess-vignette-softness", 0.4, "Vignette softness (0.0-1.0)")
	postprocessChromaticIntens = flag.Float64("postprocess-chromatic-intensity", 0.3, "Chromatic aberration intensity (0.0-1.0)")

	// Phase 5.4 (PLAN.md): Genre Palette (enhanced defaults for best visual quality)
	paletteHarmony = flag.String("palette-harmony", "triadic", "Color harmony type (complementary, analogous, triadic, tetradic, split-complementary, monochromatic)")
	paletteMood    = flag.String("palette-mood", "vibrant", "Palette mood (normal, bright, dark, saturated, muted, vibrant, pastel, tense, calm, victorious, melancholic, energetic, mystical, ominous, serene, aggressive, playful, somber, ethereal, dangerous, peaceful, chaotic, regal, desolate)")
	paletteRarity  = flag.String("palette-rarity", "epic", "Palette rarity/intensity (common, uncommon, rare, epic, legendary)")

	verbose       = flag.Bool("verbose", true, "Enable verbose logging")
	profile       = flag.Bool("profile", true, "Enable performance profiling with frame time tracking")
	multiplayer   = flag.Bool("multiplayer", false, "Connect to remote multiplayer server (default: starts local server)")
	server        = flag.String("server", "localhost:8080", "Server address (host:port) for multiplayer")
	highLatency   = flag.Bool("high-latency", false, "Use high-latency configuration optimized for Tor/onion services (200-5000ms latency)")
	hostAndPlay   = flag.Bool("host-and-play", false, "Explicitly enable host-and-play mode (default behavior when --multiplayer not specified)")
	hostLAN       = flag.Bool("host-lan", false, "Bind server to 0.0.0.0 for LAN access instead of 127.0.0.1 (requires host-and-play mode)")
	serverPort    = flag.Int("port", 8080, "Server port for --host-and-play mode (will try next 10 ports if occupied)")
	serverPlayers = flag.Int("max-players", 4, "Maximum players for --host-and-play mode")
	serverTick    = flag.Int("tick-rate", 30, "Server tick rate for --host-and-play mode (updates per second)")
	noTutorial    = flag.Bool("no-tutorial", false, "Disable tutorial for experienced players")
	enableVR      = flag.Bool("vr", false, "Enable VR mode (requires VR headset, auto-detects hardware)")
	forceVR       = flag.Bool("force-vr", false, "Force VR mode even without detected hardware (for testing)")
	showVersion   = flag.Bool("version", false, "Print version information and exit")
	spriteCacheMB = flag.Int("sprite-cache-mb", 0, "Sprite cache size in MB (0 = use platform default: 400 desktop, 150 WASM; max 300)")
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

	var clientConfig network.ClientConfig
	if *highLatency {
		clientConfig = network.TorClientConfig()
		clientLogger.Info("using high-latency client configuration (Tor/onion service optimized)")
	} else {
		clientConfig = network.DefaultClientConfig()
	}
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
			// Check if error is due to normal disconnection (EOF, connection closed)
			if errors.Is(err, io.EOF) || strings.Contains(err.Error(), "use of closed") {
				clientLogger.WithError(err).Debug("connection closed")
			} else {
				clientLogger.WithError(err).Error("network error")
			}
		}
	}()

	return networkClient
}

// canonicalGenreID normalizes genre ID aliases to the canonical form used in map keys.
// The CLI flag uses "postapoc" as the canonical ID; some older code used "postapocalyptic".
func canonicalGenreID(genreID string) string {
	if genreID == "postapocalyptic" {
		return "postapoc"
	}
	return genreID
}

// seededRandom returns a random seed based on time for CLI flag default values.
//
// IMPORTANT: This function is called ONLY at program initialization time via
// flag.Int64("seed", seededRandom(), ...) in the var() block. It is NOT called
// during gameplay. The time-based randomness is appropriate here because:
//
//  1. It provides a unique default seed for each program run when users don't
//     specify --seed explicitly
//  2. Once the seed is chosen (either from this default or from --seed flag),
//     ALL procedural generation uses that deterministic seed
//  3. This is a one-time initialization, not ongoing randomness
//
// The time.Now() usage here is explicitly EXEMPT from the "no time.Now() in
// procgen" rule because it's for CLI initialization, not for procedural content
// generation during gameplay.
func seededRandom() int64 {
	nowNano := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(nowNano))
	return rng.Int63()
}

// getGenreTheme returns the genre theme configuration for all generators.
// Uses the world seed for deterministic random genre selection.
func getGenreTheme() *genre.Genre {
	return genre.GetThemeWithSeed(*genreID, *seed)
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

	count += spawnTorchesOnHorizontalWalls(world, room, config, tileSize, spawnChance, rng)
	count += spawnTorchesOnVerticalWalls(world, room, config, tileSize, spawnChance, rng)

	return count
}

// spawnTorchesOnHorizontalWalls spawns torches on top and bottom walls.
func spawnTorchesOnHorizontalWalls(world *engine.World, room *terrain.Room, config lightConfig, tileSize int, spawnChance float64, rng *rand.Rand) int {
	count := 0

	for x := room.X; x < room.X+room.Width; x++ {
		if x%config.torchInterval == 0 {
			// Top wall
			if rng.Float64() < spawnChance {
				worldX := float64(x * tileSize)
				worldY := float64(room.Y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker, rng)
				count++
			}
			// Bottom wall
			if rng.Float64() < spawnChance {
				worldX := float64(x * tileSize)
				worldY := float64((room.Y + room.Height - 1) * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker, rng)
				count++
			}
		}
	}

	return count
}

// spawnTorchesOnVerticalWalls spawns torches on left and right walls.
func spawnTorchesOnVerticalWalls(world *engine.World, room *terrain.Room, config lightConfig, tileSize int, spawnChance float64, rng *rand.Rand) int {
	count := 0

	for y := room.Y; y < room.Y+room.Height; y++ {
		if y%config.torchInterval == 0 {
			// Left wall
			if rng.Float64() < spawnChance {
				worldX := float64(room.X * tileSize)
				worldY := float64(y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker, rng)
				count++
			}
			// Right wall
			if rng.Float64() < spawnChance {
				worldX := float64((room.X + room.Width - 1) * tileSize)
				worldY := float64(y * tileSize)
				spawnTorchLight(world, worldX, worldY, config.torchColor, config.torchRadius, config.torchFlicker, rng)
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
func spawnTorchLight(world *engine.World, x, y float64, color color.RGBA, radius float64, flicker bool, rng *rand.Rand) {
	lightEntity := world.CreateEntity()
	lightEntity.AddComponent(&engine.PositionComponent{X: x, Y: y})

	torchLight := engine.NewTorchLight(radius)
	torchLight.Color = color
	torchLight.Enabled = true
	if flicker {
		torchLight.Flickering = true
		torchLight.FlickerSpeed = 2.0 + (rng.Float64() * 2.0) // Vary flicker speed
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

// calculateHazardPosition determines spawn coordinates within a room with padding.
func calculateHazardPosition(room *terrain.Room, rng *rand.Rand, tileSize, padding int) (worldX, worldY float64, valid bool) {
	if room.Width <= padding*2 || room.Height <= padding*2 {
		return 0, 0, false
	}

	tileX := room.X + padding + rng.Intn(room.Width-padding*2)
	tileY := room.Y + padding + rng.Intn(room.Height-padding*2)
	worldX = float64(tileX * tileSize)
	worldY = float64(tileY * tileSize)
	return worldX, worldY, true
}

// addHazardComponents attaches sprite, collider, and hazard components to entity.
func addHazardComponents(entity *engine.Entity, envObj *environment.EnvironmentalObject, subType environment.SubType) {
	sprite := &engine.EbitenSprite{
		Image:   nil,
		Width:   float64(envObj.Width),
		Height:  float64(envObj.Height),
		Visible: true,
		Layer:   3,
	}
	entity.AddComponent(sprite)

	if envObj.Collidable {
		entity.AddComponent(&engine.ColliderComponent{
			Width:     float64(envObj.Width),
			Height:    float64(envObj.Height),
			Solid:     true,
			IsTrigger: false,
			Layer:     2,
			OffsetX:   -float64(envObj.Width) / 2,
			OffsetY:   -float64(envObj.Height) / 2,
		})
	}

	if envObj.Harmful && envObj.Damage > 0 {
		hazardType := determineHazardType(subType)
		entity.AddComponent(&engine.HazardComponent{
			HazardType:      hazardType,
			Duration:        0,
			DamagePerSecond: float64(envObj.Damage),
		})
	}
}

// spawnEnvironmentalHazards spawns procedural environmental hazards (fire pits, acid pools, spike traps, etc.).
// Phase 3.4: Environment & Legendary Items
func spawnEnvironmentalHazards(world *engine.World, terrain *terrain.Terrain, seed int64, genreID string, logger *logrus.Entry) int {
	rng := rand.New(rand.NewSource(seed))
	hazardCount := 0
	const tileSize = 32
	const padding = 2
	envGen := environment.NewGenerator()

	for i, room := range terrain.Rooms {
		if i == 0 {
			continue
		}

		numHazards := 1 + rng.Intn(3)

		for j := 0; j < numHazards; j++ {
			subType := selectHazardSubType(genreID, rng)

			worldX, worldY, valid := calculateHazardPosition(room, rng, tileSize, padding)
			if !valid {
				continue
			}

			config := environment.Config{
				SubType: subType,
				GenreID: genreID,
				Seed:    seed + int64(i*1000+j),
				Width:   64,
				Height:  64,
			}

			envObj, err := envGen.GenerateFromConfig(config)
			if err != nil {
				if logger != nil && logger.Logger.GetLevel() >= logrus.DebugLevel {
					logger.WithError(err).Debug("failed to generate environmental hazard")
				}
				continue
			}

			hazardEntity := world.CreateEntity()
			hazardEntity.AddComponent(&engine.PositionComponent{X: worldX, Y: worldY})
			addHazardComponents(hazardEntity, envObj, subType)
			hazardCount++
		}
	}

	return hazardCount
}

// determineHazardType maps environment SubType to engine HazardType.
func determineHazardType(subType environment.SubType) engine.HazardType {
	switch subType {
	case environment.SubTypeFirePit, environment.SubTypeLavaPit:
		return engine.HazardPoison // Use poison as generic damage type
	case environment.SubTypeAcidPool, environment.SubTypePoisonGas:
		return engine.HazardPoison
	case environment.SubTypeElectricField:
		return engine.HazardPoison // Use poison as generic damage
	case environment.SubTypeIceField:
		return engine.HazardWater // Ice/water slows movement
	case environment.SubTypeSpikes, environment.SubTypeBearTrap:
		return engine.HazardPoison // Physical damage via poison type
	default:
		return engine.HazardPoison
	}
}

// selectHazardSubType selects an appropriate hazard type for the genre.
func selectHazardSubType(genreID string, rng *rand.Rand) environment.SubType {
	// Define hazard pools per genre
	genreHazards := map[string][]environment.SubType{
		"fantasy": {
			environment.SubTypeFirePit,
			environment.SubTypeSpikes,
			environment.SubTypeBearTrap,
			environment.SubTypePoisonGas,
		},
		"scifi": {
			environment.SubTypeElectricField,
			environment.SubTypeAcidPool,
			environment.SubTypeLavaPit,
			environment.SubTypeIceField,
		},
		"horror": {
			environment.SubTypeBloodstain, // Decorative, not harmful
			environment.SubTypeBearTrap,
			environment.SubTypePoisonGas,
			environment.SubTypeSpikes,
		},
		"cyberpunk": {
			environment.SubTypeElectricField,
			environment.SubTypeAcidPool,
			environment.SubTypePoisonGas,
		},
		"postapoc": {
			environment.SubTypeFirePit,
			environment.SubTypeAcidPool,
			environment.SubTypeSpikes,
			environment.SubTypePoisonGas,
		},
	}

	// Default to fantasy hazards if genre not found
	hazards := genreHazards["fantasy"]
	if genrePool, ok := genreHazards[canonicalGenreID(genreID)]; ok {
		hazards = genrePool
	}

	// Select random hazard from pool
	return hazards[rng.Intn(len(hazards))]
}

// addMinigamesToMerchants attaches procedural minigames to merchant entities.
// Phase 3.5 (PLAN.md): Minigames in shops/taverns for player entertainment
func addMinigamesToMerchants(world *engine.World, minigameGen *minigame.Generator, seed int64, params procgen.GenerationParams, logger *logrus.Entry) int {
	if minigameGen == nil {
		return 0
	}

	minigameCount := 0
	const seedOffsetMinigame = 16000

	// Find all merchant entities
	entities := world.GetEntities()
	for i, entity := range entities {
		// Check if entity is a merchant
		if !entity.HasComponent("merchant") {
			continue
		}

		// Skip if merchant already has a minigame
		if entity.HasComponent("minigame") {
			continue
		}

		// Generate minigame with entity-specific seed
		minigameSeed := seed + seedOffsetMinigame + int64(i)
		minigameResult, err := minigameGen.Generate(minigameSeed, params)
		if err != nil {
			logger.WithError(err).WithField("entityID", entity.ID).Warn("failed to generate minigame")
			continue
		}

		mg, ok := minigameResult.(*minigame.MiniGame)
		if !ok {
			logger.WithField("entityID", entity.ID).Warn("minigame generator returned invalid type")
			continue
		}

		// Use factory function to convert procgen.GameType to engine.MiniGameType
		gameType := minigame.GameTypeToEngineType(mg.Type)

		// Add minigame component to merchant
		entity.AddComponent(&engine.MiniGameComponent{
			GameType:   gameType,
			Active:     false,
			State:      mg.State,
			Difficulty: mg.Difficulty,
			TimeLimit:  mg.TimeLimit,
		})

		logger.WithFields(logrus.Fields{
			"entityID": entity.ID,
			"gameType": gameType.String(),
			"name":     mg.Name,
			"merchant": true,
		}).Debug("added minigame to merchant")

		minigameCount++
	}

	return minigameCount
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
		"postapoc": {
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
	if genreConfig, ok := configs[canonicalGenreID(genreID)]; ok {
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
				pos, ok := posComp.(*engine.PositionComponent)
				if !ok {
					continue
				}
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
			"theme": getGenreTheme(),
		},
	}

	weaponResult, err := itemGen.Generate(seed+1, weaponParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter weapon")
		return
	}

	weapons, ok := weaponResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from weapon generator")
		return
	}
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
			"theme": getGenreTheme(),
		},
	}

	potionResult, err := itemGen.Generate(seed+2, potionParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate healing potions")
		return
	}

	potions, ok := potionResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from potion generator")
		return
	}
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
			"theme": getGenreTheme(),
		},
	}

	armorResult, err := itemGen.Generate(seed+100, armorParams)
	if err != nil {
		logger.WithError(err).Warn("failed to generate starter armor")
		return
	}

	armors, ok := armorResult.([]*item.Item)
	if !ok {
		logger.Warn("unexpected type from armor generator")
		return
	}
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
	inventory, ok := invComp.(*engine.InventoryComponent)
	if !ok {
		return
	}

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
	equipment, ok := equipComp.(*engine.EquipmentComponent)
	if !ok {
		return
	}
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

// generateLegendaryItemDrop generates a legendary item drop for an enemy (1% base, 5% for bosses).
func generateLegendaryItemDrop(world *engine.World, enemy *engine.Entity, x, y float64, seed int64, genreID string) *engine.Entity {
	// Determine drop chance based on enemy stats (bosses have higher stats)
	statsComp, hasStats := enemy.GetComponent("stats")
	baseDropChance := 0.01 // 1% base chance
	if hasStats {
		stats, ok := statsComp.(*engine.StatsComponent)
		if !ok {
			return nil
		}
		// Bosses/elites (high attack/defense) have 5% chance
		if stats.Attack > 20 || stats.Defense > 20 {
			baseDropChance = 0.05
		}
	}

	// Deterministic drop check using enemy ID and seed
	dropRNG := rand.New(rand.NewSource(seed + int64(enemy.ID*7)))
	if dropRNG.Float64() > baseDropChance {
		return nil // No drop
	}

	// Generate legendary item using procgen/item with high quality
	itemGen := item.NewItemGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.9,  // High difficulty for legendary quality
		Depth:      10.0, // High depth for legendary stats
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"quality": "legendary",
		},
	}

	generatedItem, err := itemGen.Generate(seed+int64(enemy.ID*11), params)
	if err != nil {
		return nil
	}

	itm, ok := generatedItem.(*item.Item)
	if !ok {
		return nil
	}

	// Override to ensure legendary rarity
	itm.Rarity = item.RarityLegendary
	itm.Name = "Legendary " + itm.Name // Prefix with "Legendary"

	// Use SpawnItemInWorld to create properly structured item entity
	itemEntity := engine.SpawnItemInWorld(world, itm, x, y)

	// Override sprite color to golden for legendary items
	if spriteComp, ok := itemEntity.GetComponent("sprite"); ok {
		if sprite, ok := spriteComp.(*engine.EbitenSprite); ok {
			sprite.Color = color.RGBA{255, 215, 0, 255} // Golden color
		}
	}

	return itemEntity
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
	magicGen *magic.SpellGenerator,
	skillGen *skills.SkillTreeGenerator,
) {
	// Only for NPCs/enemies, not players
	if enemy.HasComponent("input") {
		return
	}

	// Create deterministic RNG from enemy ID and seed for physics
	physicsRNG := rand.New(rand.NewSource(seed + int64(enemy.ID)))

	lootEntity := engine.GenerateLootDrop(game.World, enemy, pos.X, pos.Y, seed, genreID)
	if lootEntity != nil {
		// Add physics to procedural loot too
		lootEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 30.0, // Random velocity -30 to +30
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 30.0,
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
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 25.0, // Slightly slower velocity for recipes
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 25.0,
		})
		// Add friction for smooth deceleration
		recipeEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(recipeEntity.ID)
	}

	// Phase 3.2: Generate and spawn spell scroll drops (10% base, 25% for bosses)
	spellScrollEntity := engine.GenerateSpellScrollDrop(magicGen, game.World, enemy, pos.X, pos.Y, seed, genreID)
	if spellScrollEntity != nil {
		// Add physics to spell scrolls
		spellScrollEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 28.0, // Medium velocity for scrolls
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 28.0,
		})
		// Add friction for smooth deceleration
		spellScrollEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(spellScrollEntity.ID)
	}

	// Phase 3.2: Generate and spawn skill book drops (5% base, 15% for bosses)
	skillBookEntity := engine.GenerateSkillBookDrop(skillGen, game.World, enemy, pos.X, pos.Y, seed, genreID)
	if skillBookEntity != nil {
		// Add physics to skill books
		skillBookEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 26.0, // Slower velocity for books (heavier)
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 26.0,
		})
		// Add friction for smooth deceleration
		skillBookEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(skillBookEntity.ID)
	}

	// Phase 3.4: Generate and spawn legendary item drops (1% base, 5% for bosses)
	legendaryEntity := generateLegendaryItemDrop(game.World, enemy, pos.X, pos.Y, seed, genreID)
	if legendaryEntity != nil {
		// Add physics to legendary items (dramatic velocity)
		legendaryEntity.AddComponent(&engine.VelocityComponent{
			VX: (physicsRNG.Float64()*2.0 - 1.0) * 35.0, // Higher velocity for legendary (dramatic effect)
			VY: (physicsRNG.Float64()*2.0 - 1.0) * 35.0,
		})
		// Add friction for smooth deceleration
		legendaryEntity.AddComponent(engine.NewFrictionComponent(0.12))

		deadComp.AddDroppedItem(legendaryEntity.ID)
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
// - Spawning death particle effects
// - Tracking quest objectives
// - Awarding XP to the player (AUDIT.md MEDIUM fix)
func createDeathCallback(
	game *engine.EbitenGame,
	playerEntity **engine.Entity,
	objectiveTracker *engine.ObjectiveTrackerSystem,
	audioManager **engine.AudioManager,
	recipeGen *recipe.RecipeGenerator,
	magicGen *magic.SpellGenerator,
	skillGen *skills.SkillTreeGenerator,
	deathParticleSystem *engine.DeathParticleSystem,
	progressionSystem **engine.ProgressionSystem,
	seed int64,
	genreID string,
	logger *logrus.Logger,
	tp TimeProvider,
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
		gameTime := float64(tp.Now().Unix())
		deadComp := engine.NewDeadComponent(gameTime)
		enemy.AddComponent(deadComp)

		// AUDIT.md MEDIUM: Award XP to the player for killing the enemy.
		// XP is derived from enemy max HP (1 XP per max-HP point, minimum 10).
		// Guard: skip if the dead entity is a player-controlled entity (has an
		// input component) or if it is the player entity itself.
		if progressionSystem != nil && *progressionSystem != nil &&
			playerEntity != nil && *playerEntity != nil &&
			!enemy.HasComponent("input") && enemy != *playerEntity {
			xpAmount := calculateEnemyXP(enemy)
			if err := (*progressionSystem).AwardXP(*playerEntity, xpAmount); err != nil {
				if logger.GetLevel() >= logrus.DebugLevel {
					logging.ComponentLogger(logger, "progression").WithError(err).Debug("failed to award kill XP")
				}
			}
		}

		// Spawn death particle effects for visual feedback
		if deathParticleSystem != nil {
			deathParticleSystem.OnDeath(enemy)
		}

		// Priority 1.4: Drop all items from entity's inventory
		dropInventoryItems(game, enemy, pos, deadComp)

		// Priority 1.4: Also drop equipped items
		dropEquippedItems(game, enemy, pos, deadComp)

		// Generate and spawn procedural loot drop (in addition to inventory items)
		spawnProceduralLoot(game, enemy, pos, deadComp, recipeGen, seed, genreID, *playerEntity, objectiveTracker, magicGen, skillGen)

		// GAP-010 REPAIR: Play death sound effect
		if *audioManager != nil {
			if err := (*audioManager).PlaySFX("death", tp.Now().UnixNano()); err != nil {
				if logger.GetLevel() >= logrus.WarnLevel {
					logging.ComponentLogger(logger, "audio").WithError(err).Warn("failed to play death SFX")
				}
			}
		}
	}
}

// calculateEnemyXP returns the XP reward for killing an enemy.
// Derives the amount from the enemy's max HP (1 XP per HP point) with a minimum
// of 10 and a maximum of 10 000 to prevent degenerate values.
func calculateEnemyXP(enemy *engine.Entity) int {
	const minXP = 10
	const maxXP = 10_000
	healthComp, ok := enemy.GetComponent("health")
	if !ok {
		return minXP
	}
	hc, ok := healthComp.(*engine.HealthComponent)
	if !ok {
		return minXP
	}
	xp := int(hc.Max)
	if xp < minXP {
		return minXP
	}
	if xp > maxXP {
		return maxXP
	}
	return xp
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
	serializeOnboardingState(game, playerState)
	serializeContextTutorialState(game, playerState)
	serializeQoLState(player, playerState)

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
		exp, ok := expComp.(*engine.ExperienceComponent)
		if !ok {
			return
		}
		state.Level = exp.Level
		state.Experience = exp.CurrentXP
	}
}

// serializeInventory extracts inventory data to state.
func serializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv, ok := invComp.(*engine.InventoryComponent)
		if !ok {
			return
		}
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
		equipment, ok := equip.(*engine.EquipmentComponent)
		if !ok {
			return
		}
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
		mana, ok := manaComp.(*engine.ManaComponent)
		if !ok {
			return
		}
		state.CurrentMana = mana.Current
		state.MaxMana = mana.Max
	}

	if slotsComp, hasSlots := player.GetComponent("spell_slots"); hasSlots {
		slots, ok := slotsComp.(*engine.SpellSlotComponent)
		if !ok {
			return
		}
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

// serializeOnboardingState extracts onboarding manager state to state.
// Phase 4.5: Enables persistence of onboarding flow progress.
func serializeOnboardingState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.OnboardingManager != nil {
		data := game.OnboardingManager.ExportState()
		state.OnboardingState = &saveload.OnboardingStateData{
			CurrentState: data.CurrentState,
			Enabled:      data.Enabled,
			Skipped:      data.Skipped,
			PlayerClass:  data.PlayerClass,
		}
	}
}

// serializeContextTutorialState extracts context-sensitive tutorial state to state.
// Phase 4.5: Enables persistence of viewed tutorial topics.
func serializeContextTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.ContextualTutorial != nil {
		data := game.ContextualTutorial.ExportState()
		state.ContextTutorialState = &saveload.ContextTutorialStateData{
			Enabled:      data.Enabled,
			ViewedTopics: data.ViewedTopics,
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
	deserializeOnboardingState(game, playerState)
	deserializeContextTutorialState(game, playerState)
	deserializeQoLState(player, playerState)
}

// deserializePosition restores player position from state.
func deserializePosition(player *engine.Entity, state *saveload.PlayerState) {
	if posComp, ok := player.GetComponent("position"); ok {
		pos, ok := posComp.(*engine.PositionComponent)
		if !ok {
			return
		}
		pos.X, pos.Y = state.X, state.Y
	}
}

// deserializeHealth restores player health from state.
func deserializeHealth(player *engine.Entity, state *saveload.PlayerState) {
	if healthComp, ok := player.GetComponent("health"); ok {
		health, ok := healthComp.(*engine.HealthComponent)
		if !ok {
			return
		}
		health.Current, health.Max = state.CurrentHealth, state.MaxHealth
	}
}

// deserializeStats restores player stats from state.
func deserializeStats(player *engine.Entity, state *saveload.PlayerState) {
	if statsComp, ok := player.GetComponent("stats"); ok {
		stats, ok := statsComp.(*engine.StatsComponent)
		if !ok {
			return
		}
		stats.Attack = state.Attack
		stats.Defense = state.Defense
		stats.MagicPower = state.MagicPower
	}
}

// deserializeExperience restores player level and XP from state.
func deserializeExperience(player *engine.Entity, state *saveload.PlayerState) {
	if expComp, ok := player.GetComponent("experience"); ok {
		exp, ok := expComp.(*engine.ExperienceComponent)
		if !ok {
			return
		}
		exp.Level = state.Level
		exp.CurrentXP = state.Experience
	}
}

// deserializeInventory restores inventory from state.
func deserializeInventory(player *engine.Entity, state *saveload.PlayerState) {
	if invComp, ok := player.GetComponent("inventory"); ok {
		inv, ok := invComp.(*engine.InventoryComponent)
		if !ok {
			return
		}
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
		equipment, ok := equipComp.(*engine.EquipmentComponent)
		if !ok {
			return
		}
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
		mana, ok := manaComp.(*engine.ManaComponent)
		if !ok {
			return
		}
		mana.Current, mana.Max = state.CurrentMana, state.MaxMana
	}

	if slotsComp, ok := player.GetComponent("spell_slots"); ok {
		slots, ok := slotsComp.(*engine.SpellSlotComponent)
		if !ok {
			return
		}
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

// deserializeOnboardingState restores onboarding manager state from state.
// Phase 4.5: Restores onboarding flow progress after load.
func deserializeOnboardingState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.OnboardingManager != nil && state.OnboardingState != nil {
		data := engine.OnboardingStateData{
			CurrentState: state.OnboardingState.CurrentState,
			Enabled:      state.OnboardingState.Enabled,
			Skipped:      state.OnboardingState.Skipped,
			PlayerClass:  state.OnboardingState.PlayerClass,
		}
		game.OnboardingManager.ImportState(data)
	}
}

// deserializeContextTutorialState restores context-sensitive tutorial state from state.
// Phase 4.5: Restores viewed tutorial topics after load.
func deserializeContextTutorialState(game *engine.EbitenGame, state *saveload.PlayerState) {
	if game.ContextualTutorial != nil && state.ContextTutorialState != nil {
		data := saveload.ContextTutorialStateData{
			Enabled:      state.ContextTutorialState.Enabled,
			ViewedTopics: state.ContextTutorialState.ViewedTopics,
		}
		game.ContextualTutorial.ImportState(data)
	}
}

// serializeQoLState extracts QoL preferences to state.
// AUDIT.md fix: Allows QoL settings to persist across saves/loads.
func serializeQoLState(player *engine.Entity, state *saveload.PlayerState) {
	qolComp, ok := player.GetComponent("qol")
	if !ok {
		return
	}
	q, ok := qolComp.(*qol.QoLComponent)
	if !ok {
		return
	}

	// Serialize craft queue to JSON
	var craftQueueJSON []byte
	if len(q.CraftQueue) > 0 {
		craftQueueJSON, _ = json.Marshal(q.CraftQueue)
	}

	state.QoLData = &saveload.QoLStateData{
		PlayerID:        q.PlayerID,
		AutoLootEnabled: q.AutoLootEnabled,
		AutoLootRadius:  q.AutoLootRadius,
		CraftQueueJSON:  craftQueueJSON,
		SortPreset:      q.SortPreset,
		MountWhistle:    q.MountWhistle,
		RecipeTracking:  q.RecipeTracking,
	}
}

// deserializeQoLState restores QoL preferences from state.
// AUDIT.md fix: Restores QoL settings after load.
func deserializeQoLState(player *engine.Entity, state *saveload.PlayerState) {
	if state.QoLData == nil {
		return
	}

	// Get or create QoL component
	qolComp, ok := player.GetComponent("qol")
	var q *qol.QoLComponent
	if ok {
		q, ok = qolComp.(*qol.QoLComponent)
		if !ok {
			return
		}
	} else {
		// Create new QoL component if it doesn't exist
		q = &qol.QoLComponent{
			PlayerID: player.ID,
		}
		player.AddComponent(q)
	}

	// Restore QoL settings
	q.PlayerID = state.QoLData.PlayerID
	q.AutoLootEnabled = state.QoLData.AutoLootEnabled
	q.AutoLootRadius = state.QoLData.AutoLootRadius
	q.SortPreset = state.QoLData.SortPreset
	q.MountWhistle = state.QoLData.MountWhistle
	q.RecipeTracking = state.QoLData.RecipeTracking

	// Deserialize craft queue from JSON
	if len(state.QoLData.CraftQueueJSON) > 0 {
		var craftQueue []*qol.CraftQueueEntry
		if err := json.Unmarshal(state.QoLData.CraftQueueJSON, &craftQueue); err == nil {
			q.CraftQueue = craftQueue
		}
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

// convertVehicleType converts vehicle.VehicleType to engine.VehicleType.
func convertVehicleType(vType vehicle.VehicleType) engine.VehicleType {
	switch vType {
	case vehicle.TypeMount:
		return engine.VehicleMount
	case vehicle.TypeCart:
		return engine.VehicleCart
	case vehicle.TypeBoat:
		return engine.VehicleBoat
	case vehicle.TypeGlider:
		return engine.VehicleGlider
	case vehicle.TypeMech:
		return engine.VehicleMech
	default:
		return engine.VehicleMount
	}
}

// convertColorToRGBA converts a uint32 color value to color.RGBA.
func convertColorToRGBA(colorValue uint32) color.RGBA {
	r := uint8((colorValue >> 16) & 0xFF)
	g := uint8((colorValue >> 8) & 0xFF)
	b := uint8(colorValue & 0xFF)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// calculateVehicleSizing determines sprite size and collider dimensions based on vehicle type.
func calculateVehicleSizing(vType vehicle.VehicleType) (size int, colliderSize float64) {
	switch vType {
	case vehicle.TypeMount:
		return 32, 28.0
	case vehicle.TypeCart:
		return 40, 36.0
	case vehicle.TypeBoat:
		return 48, 44.0
	case vehicle.TypeGlider:
		return 36, 32.0
	case vehicle.TypeMech:
		return 44, 40.0
	default:
		return 32, 28.0
	}
}

// createVehicleSpawnData converts a vehicle instance to spawn data.
func createVehicleSpawnData(v *vehicle.Vehicle) engine.VehicleSpawnData {
	engineType := convertVehicleType(v.VehicleType)
	colorRGBA := convertColorToRGBA(v.Color)
	size, colliderSize := calculateVehicleSizing(v.VehicleType)

	return engine.VehicleSpawnData{
		Name:         v.Name,
		VehicleType:  engineType,
		Components:   v.ToComponents(),
		Color:        colorRGBA,
		Size:         size,
		ColliderSize: colliderSize,
	}
}

// spawnVehicles generates and spawns vehicles in terrain rooms (V4.0).
func spawnVehicles(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	vehicleGen := vehicle.NewVehicleGenerator()

	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil
	}

	vehicleCount := 2 + (roomCount-2)/4
	if vehicleCount > 5 {
		vehicleCount = 5
	}

	vehicleResult, err := vehicleGen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("failed to generate vehicles: %w", err)
	}

	vehicles, ok := vehicleResult.([]*vehicle.Vehicle)
	if !ok || len(vehicles) == 0 {
		return 0, fmt.Errorf("vehicle generator returned invalid result")
	}

	if len(vehicles) > vehicleCount {
		vehicles = vehicles[:vehicleCount]
	}

	vehicleSpawnData := make([]engine.VehicleSpawnData, len(vehicles))
	for i, v := range vehicles {
		vehicleSpawnData[i] = createVehicleSpawnData(v)
	}

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

// calculateCompanionSizing determines sprite size and collider dimensions based on companion type.
func calculateCompanionSizing(compType engine.CompanionType) (size int, colliderSize float64) {
	switch compType {
	case engine.CompanionTypePet:
		return 24, 20.0
	case engine.CompanionTypeSummon:
		return 28, 24.0
	case engine.CompanionTypeHireling:
		return 28, 24.0
	case engine.CompanionTypeElemental:
		return 32, 28.0
	case engine.CompanionTypeUndead:
		return 30, 26.0
	case engine.CompanionTypeRobot:
		return 30, 26.0
	case engine.CompanionTypeSpirit:
		return 26, 22.0
	case engine.CompanionTypeInsect:
		return 22, 18.0
	default:
		return 28, 24.0
	}
}

// createCompanionSpawnData converts a companion instance to spawn data.
func createCompanionSpawnData(comp *companion.Companion, companionSeed int64, genreID string) engine.CompanionSpawnData {
	companionColor := generateCompanionColor(comp.Type, genreID, companionSeed)
	size, colliderSize := calculateCompanionSizing(comp.Type)

	return engine.CompanionSpawnData{
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
	}
}

// spawnCompanions generates and spawns companions in terrain rooms (V4.0).
func spawnCompanions(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	companionGen := companion.NewGenerator()

	roomCount := len(terrainMap.Rooms)
	if roomCount < 2 {
		return 0, nil
	}

	companionCount := 1 + (roomCount-2)/5
	if companionCount > 3 {
		companionCount = 3
	}

	companionSpawnData := make([]engine.CompanionSpawnData, 0, companionCount)
	for i := 0; i < companionCount; i++ {
		companionSeed := seed + int64(i*1000)

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

		spawnData := createCompanionSpawnData(comp, companionSeed, params.GenreID)
		companionSpawnData = append(companionSpawnData, spawnData)
	}

	if len(companionSpawnData) == 0 {
		return 0, nil
	}

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

// spawnArchaeologicalSites generates and spawns an archaeological site entity (Phase 30).
// The site is positioned in the terrain using the percentile location from the generator.
func spawnArchaeologicalSites(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	gen := story.NewArchaeologyGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("archaeological site generation failed: %w", err)
	}
	site, ok := result.(*story.ArchaeologicalSite)
	if !ok {
		return 0, fmt.Errorf("archaeological site generator returned invalid type")
	}

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{
		X: site.Location.X * float64(terrainMap.Width) / 100.0,
		Y: site.Location.Y * float64(terrainMap.Height) / 100.0,
	})
	entity.AddComponent(&engine.ArchaeologicalSiteComponent{Site: site})
	entity.AddComponent(&engine.EbitenSprite{Width: 16, Height: 16, Visible: true, Layer: 4})
	entity.AddComponent(&engine.ColliderComponent{Width: 16.0, Height: 16.0})

	logger.WithFields(logrus.Fields{
		"name":  site.Name,
		"genre": site.Genre,
		"depth": site.Depth,
	}).Debug("spawned archaeological site")

	return 1, nil
}

// spawnTimelines generates and spawns a world-history timeline entity (Phase 30).
// One entity is placed at the centre of the map to represent the world's history.
func spawnTimelines(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	gen := story.NewTimelineGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("timeline generation failed: %w", err)
	}
	timeline, ok := result.(*story.Timeline)
	if !ok {
		return 0, fmt.Errorf("timeline generator returned invalid type")
	}

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{
		X: float64(terrainMap.Width) / 2.0,
		Y: float64(terrainMap.Height) / 2.0,
	})
	entity.AddComponent(&engine.TimelineComponent{Timeline: timeline})
	entity.AddComponent(&engine.EbitenSprite{Width: 16, Height: 16, Visible: false, Layer: 4})
	entity.AddComponent(&engine.ColliderComponent{Width: 16.0, Height: 16.0})

	logger.WithFields(logrus.Fields{
		"eraCount": len(timeline.Eras),
	}).Debug("spawned world timeline")

	return 1, nil
}

// spawnCrossDungeonStories generates and spawns a cross-dungeon narrative entity (Phase 30).
// One story entity is placed at the centre of the map; its fragments are distributed
// across dungeon levels through the CrossDungeonStory.Levels slice.
func spawnCrossDungeonStories(world *engine.World, terrainMap *terrain.Terrain, seed int64, params procgen.GenerationParams, logger *logrus.Entry) (int, error) {
	gen := story.NewCrossDungeonGenerator()
	result, err := gen.Generate(seed, params)
	if err != nil {
		return 0, fmt.Errorf("cross-dungeon story generation failed: %w", err)
	}
	cdStory, ok := result.(*story.CrossDungeonStory)
	if !ok {
		return 0, fmt.Errorf("cross-dungeon story generator returned invalid type")
	}

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{
		X: float64(terrainMap.Width) / 2.0,
		Y: float64(terrainMap.Height) / 2.0,
	})
	entity.AddComponent(&engine.CrossDungeonStoryComponent{Story: cdStory})
	entity.AddComponent(&engine.EbitenSprite{Width: 16, Height: 16, Visible: false, Layer: 4})
	entity.AddComponent(&engine.ColliderComponent{Width: 16.0, Height: 16.0})

	logger.WithFields(logrus.Fields{
		"theme":     cdStory.Theme,
		"fragments": len(cdStory.Fragments),
		"levelSpan": cdStory.LevelSpan,
	}).Debug("spawned cross-dungeon story")

	return 1, nil
}

// parsePaletteOptions parses command-line flags into palette.GenerationOptions (Phase 5.4).
//
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (36) is intentional—exhaustively
// validates and maps CLI flag strings to typed enum values for harmony (6), mood (24),
// and rarity (5) options. The switch provides compile-time enum coverage and clear error messages.
func parsePaletteOptions() (*palette.GenerationOptions, error) {
	opts := &palette.GenerationOptions{}

	// Parse harmony type
	switch *paletteHarmony {
	case "complementary":
		opts.Harmony = palette.HarmonyComplementary
	case "analogous":
		opts.Harmony = palette.HarmonyAnalogous
	case "triadic":
		opts.Harmony = palette.HarmonyTriadic
	case "tetradic":
		opts.Harmony = palette.HarmonyTetradic
	case "split-complementary":
		opts.Harmony = palette.HarmonySplitComplementary
	case "monochromatic":
		opts.Harmony = palette.HarmonyMonochromatic
	default:
		return nil, fmt.Errorf("invalid harmony type: %s", *paletteHarmony)
	}

	// Parse mood type
	switch *paletteMood {
	case "normal":
		opts.Mood = palette.MoodNormal
	case "bright":
		opts.Mood = palette.MoodBright
	case "dark":
		opts.Mood = palette.MoodDark
	case "saturated":
		opts.Mood = palette.MoodSaturated
	case "muted":
		opts.Mood = palette.MoodMuted
	case "vibrant":
		opts.Mood = palette.MoodVibrant
	case "pastel":
		opts.Mood = palette.MoodPastel
	case "tense":
		opts.Mood = palette.MoodTense
	case "calm":
		opts.Mood = palette.MoodCalm
	case "victorious":
		opts.Mood = palette.MoodVictorious
	case "melancholic":
		opts.Mood = palette.MoodMelancholic
	case "energetic":
		opts.Mood = palette.MoodEnergetic
	case "mystical":
		opts.Mood = palette.MoodMystical
	case "ominous":
		opts.Mood = palette.MoodOminous
	case "serene":
		opts.Mood = palette.MoodSerene
	case "aggressive":
		opts.Mood = palette.MoodAggressive
	case "playful":
		opts.Mood = palette.MoodPlayful
	case "somber":
		opts.Mood = palette.MoodSomber
	case "ethereal":
		opts.Mood = palette.MoodEthereal
	case "dangerous":
		opts.Mood = palette.MoodDangerous
	case "peaceful":
		opts.Mood = palette.MoodPeaceful
	case "chaotic":
		opts.Mood = palette.MoodChaotic
	case "regal":
		opts.Mood = palette.MoodRegal
	case "desolate":
		opts.Mood = palette.MoodDesolate
	default:
		return nil, fmt.Errorf("invalid mood type: %s", *paletteMood)
	}

	// Parse rarity type
	switch *paletteRarity {
	case "common":
		opts.Rarity = palette.RarityCommon
	case "uncommon":
		opts.Rarity = palette.RarityUncommon
	case "rare":
		opts.Rarity = palette.RarityRare
	case "epic":
		opts.Rarity = palette.RarityEpic
	case "legendary":
		opts.Rarity = palette.RarityLegendary
	default:
		return nil, fmt.Errorf("invalid rarity type: %s", *paletteRarity)
	}

	// Default minimum colors (can be extended later with flag)
	opts.MinColors = 12

	return opts, nil
}

// validateClientConfiguration validates client configuration flags.
// Returns an error if any configuration is invalid, with a helpful message.
func validateClientConfiguration() error {
	validator := config.NewValidator()

	// Build configuration to validate
	cfg := &config.Config{
		Genre: *genreID,
	}

	// Validate host-and-play server configuration if enabled
	if *hostAndPlay {
		cfg.Port = fmt.Sprintf("%d", *serverPort)
		cfg.MaxPlayers = *serverPlayers
		cfg.ValidateMaxPlayers = true
		cfg.TickRate = *serverTick
		cfg.ValidateTickRate = true
	}

	if err := validator.ValidateAll(cfg); err != nil {
		return fmt.Errorf("%w\n\nRun with -help to see all available options", err)
	}

	return nil
}
