//go:build !android && !ios
// +build !android,!ios

package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/book"
	"github.com/opd-ai/venture/pkg/procgen/companion"
	"github.com/opd-ai/venture/pkg/procgen/environment"
	"github.com/opd-ai/venture/pkg/procgen/minigame"
	"github.com/opd-ai/venture/pkg/procgen/story"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

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
	switch canonicalGenreID(genreID) {
	case "fantasy":
		baseR, baseG, baseB = 101, 67, 33 // Dark wood brown
	case "scifi":
		baseR, baseG, baseB = 120, 130, 140 // Metallic gray
	case "horror":
		baseR, baseG, baseB = 60, 50, 45 // Very dark brown
	case "cyberpunk":
		baseR, baseG, baseB = 80, 80, 100 // Dark bluish
	case "postapoc":
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
