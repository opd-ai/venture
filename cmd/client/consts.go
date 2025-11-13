//go:build !android && !ios
// +build !android,!ios

package main

// Game configuration constants used throughout the client initialization.
const (
	// Movement and collision
	playerMaxSpeed        = 200.0 // units/second max movement speed
	collisionGridCellSize = 64.0  // spatial partitioning grid cell size in pixels
	
	// Tile and sprite dimensions
	tileSize              = 32    // standard tile size used throughout the engine
	playerSpriteWidth     = 28    // player sprite width in pixels
	playerSpriteHeight    = 28    // player sprite height in pixels
	playerColliderOffset  = -14   // center offset for collider (28/2 = 14)
	
	// Audio
	audioSampleRate = 44100 // 44.1kHz sample rate for audio synthesis
	
	// Animation
	animationCacheSize    = 300  // max cached animation sequences (WASM optimization)
	playerFrameTime       = 0.15 // seconds per animation frame (~6.7 FPS)
	playerFrameCount      = 4    // frames per animation sequence
	
	// Player stats
	playerStartHealth     = 100   // initial and max health
	playerStartMana       = 100   // initial and max mana
	playerManaRegen       = 5.0   // mana regeneration per second
	playerStartGold       = 100   // starting gold amount
	playerInventorySlots  = 20    // maximum inventory slots
	playerInventoryWeight = 100.0 // maximum carry weight
	playerAttackDamage    = 15    // base attack damage
	playerAttackRange     = 50    // attack range in pixels
	playerAttackCooldown  = 0.5   // attack cooldown in seconds
	playerBaseAttack      = 10    // base attack stat
	playerBaseDefense     = 5     // base defense stat
	playerCritChance      = 0.05  // 5% critical hit chance
	playerCritDamage      = 1.5   // 1.5x critical damage multiplier
	playerEvasion         = 0.05  // 5% evasion chance
	
	// Camera and lighting
	cameraSmoothing         = 0.1   // camera follow smoothing factor
	playerRotationSpeed     = 3.0   // radians per second (~172°/s)
	playerTorchRadius       = 200   // torch light radius in pixels
	animationCloseThreshold = 200.0 // close distance threshold for LOD
	animationMidThreshold   = 400.0 // mid distance threshold for LOD
	maxLights               = 16    // maximum dynamic lights
	
	// World generation
	defaultDifficulty    = 0.5 // default difficulty level
	defaultDepth         = 1   // starting dungeon depth
	defaultTerrainWidth  = 80  // terrain width in tiles
	defaultTerrainHeight = 50  // terrain height in tiles
	defaultMerchantCount = 2   // merchants per level
	defaultPuzzleCount   = 5   // puzzles per level
	worldPixelsPerTile   = 32.0 // pixels per tile for world coordinate conversion
	
	// Seed offsets for deterministic generation
	seedOffsetFaction         = 1000 // offset for faction generation
	seedOffsetStation         = 1000 // offset for station generation
	seedOffsetStatusEffect    = 999  // offset for status effect RNG
	seedOffsetPlayerAnimation = 1000 // multiplier for player animation seed
	seedOffsetPuzzle          = 2000 // offset for puzzle generation
	seedOffsetLight           = 2000 // offset for environmental light generation
	seedOffsetObject          = 3000 // offset for destructible object generation
	seedOffsetWeather         = 3000 // offset for weather generation
	seedOffsetFirePropagation = 1090 // offset for fire propagation system
	seedOffsetDestructible    = 1100 // offset for destructible object system
	
	// Fallback positions
	fallbackPlayerX = 400 // fallback X position if no valid spawn
	fallbackPlayerY = 300 // fallback Y position if no valid spawn
	
	// Interaction
	merchantInteractionRange = 64.0 // interaction range for merchants in pixels
	
	// Spatial partitioning
	quadtreeCapacity = 8 // entities per quadtree node before subdivision
	
	// Performance monitoring
	perfMonitorInterval = 10 // seconds between performance metric logs
)
