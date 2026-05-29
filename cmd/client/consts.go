//go:build !android && !ios
// +build !android,!ios

package main

// Game configuration constants used throughout the client initialization.
const (
	// Movement and collision
	playerMaxSpeed        = 200.0 // units/second max movement speed
	collisionGridCellSize = 64.0  // spatial partitioning grid cell size in pixels

	// Tile and sprite dimensions
	tileSize             = 32  // standard tile size used throughout the engine
	playerSpriteWidth    = 64  // player sprite width in pixels (Phase 45 enhanced)
	playerSpriteHeight   = 64  // player sprite height in pixels (Phase 45 enhanced)
	playerColliderOffset = -32 // center offset for collider (64/2 = 32)

	// Audio
	audioSampleRate = 44100 // 44.1kHz sample rate for audio synthesis

	// Animation
	animationCacheSize = 300  // max cached animation sequences (WASM optimization)
	playerFrameTime    = 0.15 // seconds per animation frame (~6.7 FPS)
	playerFrameCount   = 8    // frames per animation sequence (V7.0: 8-frame animations)

	// Sprite cache (Phase 1.2)
	spriteCacheMaxSize = 400 * 1024 * 1024 // 400MB limit for sprite cache

	// WASM-specific performance limits.
	// Browser WASM has ~1-2GB total heap; these keep the sprite/animation
	// subsystems well within budget to prevent progressive slowdown.
	wasmSpriteCacheMaxSize = 150 * 1024 * 1024 // 150MB (vs 400MB desktop)
	wasmAnimationCacheSize = 150               // max cached animation sequences (vs 300 desktop)
	wasmMemorySoftLimit    = 120 * 1024 * 1024 // soft eviction threshold
	wasmMemoryHardLimit    = 150 * 1024 * 1024 // hard eviction threshold

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
	defaultDifficulty    = 0.5  // default difficulty level
	defaultDepth         = 1    // starting dungeon depth
	defaultTerrainWidth  = 80   // terrain width in tiles
	defaultTerrainHeight = 50   // terrain height in tiles
	defaultMerchantCount = 2    // merchants per level
	defaultPuzzleCount   = 5    // puzzles per level
	worldPixelsPerTile   = 32.0 // pixels per tile for world coordinate conversion

	// Seed offsets for deterministic generation
	seedOffsetFaction            = 1000  // offset for faction generation
	seedOffsetStation            = 1010  // offset for station generation
	seedOffsetStatusEffect       = 1020  // offset for status effect RNG
	seedOffsetPlayerAnimation    = 1030  // multiplier for player animation seed
	seedOffsetPuzzle             = 2000  // offset for puzzle generation
	seedOffsetLight              = 2010  // offset for environmental light generation
	seedOffsetObject             = 3000  // offset for destructible object generation
	seedOffsetWeather            = 3010  // offset for weather generation
	seedOffsetFirePropagation    = 1090  // offset for fire propagation system
	seedOffsetDestructible       = 1100  // offset for destructible object system
	seedOffsetDestructionPhysics = 1110  // offset for destruction physics (debris generation)
	seedOffsetSpellEffects       = 1200  // offset for spell effect system (V4.0)
	seedOffsetVehicle            = 4000  // offset for vehicle generation (V4.0)
	seedOffsetCompanion          = 5000  // offset for companion generation (V4.0)
	seedOffsetBook               = 6000  // offset for book/bookshelf generation (V4.0)
	seedOffsetStory              = 7000  // offset for story fragment generation (Phase 30)
	seedOffsetArchaeo            = 7010  // offset for archaeological site generation
	seedOffsetTimeline           = 7020  // offset for historical timeline generation
	seedOffsetCrossDungeon       = 7030  // offset for cross-dungeon narrative generation
	seedOffsetReverb             = 8000  // offset for reverb system (Phase 14.4)
	seedOffsetInvestigation      = 9000  // offset for investigation system (Phase 30)
	seedOffsetNPCDialog          = 10000 // offset for NPC dialog system (Phase 31)
	seedOffsetWorldEvents        = 11000 // offset for world events system (Phase 6.3)
	seedOffsetEnvironment        = 12000 // offset for environmental hazard generation (Phase 3.4)
	seedOffsetNarrative          = 13000 // offset for procedural narrative arc generation (Phase 3.6)
	seedOffsetTradeRoutes        = 14000 // offset for trade route system (Phase 4.4)
	seedOffsetFishing            = 15000 // offset for fishing system (Phase 95-96)
	seedOffsetSpecManaBoost      = 16000 // offset for specialization mana boost system (class-mana integration)
	seedOffsetSpecHealthRegen    = 16500 // offset for specialization health regen system (class-health integration)
	seedOffsetSpecSpellDamage    = 17000 // offset for specialization spell damage system (class-spell damage integration)
	seedOffsetSpecAttackSpeed    = 17500 // offset for specialization attack speed system (class-combat integration)
	seedOffsetSpecDefense        = 18000 // offset for specialization defense system (class-defense integration)
	seedOffsetSpecLifesteal      = 18500 // offset for specialization lifesteal system (class-lifesteal integration)
	seedOffsetDualClassSynergy   = 19000 // offset for dual-class synergy system (class combination bonuses)
	seedOffsetSpecCritDamage     = 19500 // offset for specialization crit damage system (class-crit integration)
	seedOffsetSpecEvasion        = 20000 // offset for specialization evasion system (class-evasion integration)

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
