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

// systemInitDebugEnabled caches whether debug-level system initialization logging is enabled.
// This variable is set once per InitializeGameSystems call to avoid 44 GetLevel() checks.
// Eliminates 5-15ms startup overhead in debug builds.
var systemInitDebugEnabled bool

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
	InputSystem                         *InputSystem
	CombatSystem                        *CombatSystem
	CollisionSystem                     *CollisionSystem
	ProjectileSystem                    *ProjectileSystem
	AudioManager                        *AudioManager
	ObjectiveTracker                    *ObjectiveTrackerSystem
	CommerceSystem                      *CommerceSystem
	DialogSystem                        *DialogSystem
	CraftingSystem                      *CraftingSystem
	InteractionSystem                   *InteractionSystem
	MiniGameSystem                      *MiniGameSystem
	AnimationSystem                     *AnimationSystem
	ParticleSystem                      *ParticleSystem
	TutorialSystem                      *EbitenTutorialSystem
	HelpSystem                          *EbitenHelpSystem
	LevelUpParticleSystem               *LevelUpParticleSystem
	ItemPickupParticleSystem            *ItemPickupParticleSystem
	SpellEffectParticleSystem           *SpellEffectParticleSystem
	DeathParticleSystem                 *DeathParticleSystem
	DamageResistanceParticleSystem      *DamageResistanceParticleSystem
	ShieldAbsorbParticleSystem          *ShieldAbsorbParticleSystem
	CompanionLevelUpParticleSystem      *CompanionLevelUpParticleSystem
	ItemPickupSystem                    *ItemPickupSystem
	ProgressionSystem                   *ProgressionSystem
	CompanionProgressionSystem          *CompanionProgressionSystem
	WeatherAudioSystem                  *WeatherAudioSystem
	FactionXPBonusSystem                *FactionXPBonusSystem
	WeatherManaRegenSystem              *WeatherManaRegenSystem
	WeatherCooldownSystem               *WeatherCooldownSystem
	TerrainMovementSpeedSystem          *TerrainMovementSpeedSystem
	TerrainCombatBonusSystem            *TerrainCombatBonusSystem
	TerrainStealthSystem                *TerrainStealthSystem
	TerrainStatusEffectSystem           *TerrainStatusEffectSystem
	TerrainManaRegenSystem              *TerrainManaRegenSystem
	TerrainSpellDamageSystem            *TerrainSpellDamageSystem
	LowHealthVFXSystem                  *LowHealthVFXSystem
	ManaRegenParticleSystem             *ManaRegenParticleSystem
	CompanionAuraParticleSystem         *CompanionAuraParticleSystem
	SpecializationManaBoostSystem       *SpecializationManaBoostSystem
	SpecializationHealthRegenSystem     *SpecializationHealthRegenSystem
	SpecializationSpellDamageSystem     *SpecializationSpellDamageSystem
	SpecializationAttackSpeedSystem     *SpecializationAttackSpeedSystem
	SpecializationDefenseSystem         *SpecializationDefenseSystem
	ElementalComboParticleSystem        *ElementalComboParticleSystem
	WeatherRangedAccuracySystem         *WeatherRangedAccuracySystem
	StatusEffectEvasionSystem           *StatusEffectEvasionSystem
	StatusEffectCriticalChanceSystem    *StatusEffectCriticalChanceSystem
	WeatherXPBonusSystem                *WeatherXPBonusSystem
	ElementalComboDamageSystem          *ElementalComboDamageSystem
	ElementalCompanionSynergySystem     *ElementalCompanionSynergySystem
	CompanionSpellAmplificationSystem   *CompanionSpellAmplificationSystem
	LifestealSystem                     *LifestealSystem
	StatusEffectManaCostSystem          *StatusEffectManaCostSystem
	StatusEffectDamageParticleSystem    *StatusEffectDamageParticleSystem
	WeaponSwingParticleSystem           *WeaponSwingParticleSystem
	FearFleeParticleSystem              *FearFleeParticleSystem
	StatusEffectDamageBoostSystem       *StatusEffectDamageBoostSystem
	TerrainCompanionBonusSystem         *TerrainCompanionBonusSystem
	FactionCompanionBehaviorSystem      *FactionCompanionBehaviorSystem
	WeatherCompanionBonusSystem         *WeatherCompanionBonusSystem
	EvasionParticleSystem               *EvasionParticleSystem
	FootstepParticleSystem              *FootstepParticleSystem
	SpecializationStatusResistSystem    *SpecializationStatusResistSystem
	StatusEffectHealthRegenSystem       *StatusEffectHealthRegenSystem
	HealingParticleSystem               *HealingParticleSystem
	SpecializationCritDamageSystem      *SpecializationCritDamageSystem
	SpecializationEvasionSystem         *SpecializationEvasionSystem
	ShieldRegenSystem                   *ShieldRegenSystem
	WeatherMeleeDamageSystem            *WeatherMeleeDamageSystem
	DrowningParticleSystem              *DrowningParticleSystem
	DestructionParticleSystem           *DestructionParticleSystem
	BlockParticleSystem                 *BlockParticleSystem
	TimeOfDayLightingSystem             *TimeOfDayLightingSystem
	TimeOfDayStealthSystem              *TimeOfDayStealthSystem
	TimeOfDayXPBonusSystem              *TimeOfDayXPBonusSystem
	TimeOfDayManaCostSystem             *TimeOfDayManaCostSystem
	TimeOfDayCriticalChanceSystem       *TimeOfDayCriticalChanceSystem
	TerrainAmbushCritSystem             *TerrainAmbushCritSystem
	FactionDamageBonusSystem            *FactionDamageBonusSystem
	WeatherCritChanceSystem             *WeatherCritChanceSystem
	StealthIndicatorParticleSystem      *StealthIndicatorParticleSystem
	FishingWeatherBonusSystem           *FishingWeatherBonusSystem
	TimeOfDayFishingBonusSystem         *TimeOfDayFishingBonusSystem
	FishingCatchParticleSystem          *FishingCatchParticleSystem
	TimeOfDayFishingBonusParticleSystem *TimeOfDayFishingBonusParticleSystem
	TimeOfDayCompanionBonusSystem       *TimeOfDayCompanionBonusSystem
	TimeOfDayHealthRegenSystem          *TimeOfDayHealthRegenSystem
	TimeOfDayManaRegenSystem            *TimeOfDayManaRegenSystem
	TimeOfDayBlockChanceSystem          *TimeOfDayBlockChanceSystem
	TerrainCombatBonusParticleSystem    *TerrainCombatBonusParticleSystem
	CompanionManaRegenSystem            *CompanionManaRegenSystem

	// System wrappers
	AnimationSystemWrapper            System
	RotationSystemWrapper             System
	SquadSystemWrapper                System
	CompanionProgressionSystemWrapper System
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

	// Cache debug flag once to avoid 44 GetLevel() checks during system creation
	systemInitDebugEnabled = logger.GetLevel() >= logrus.DebugLevel

	result := &SystemInitResult{}

	if config.EnableVerboseLogging {
		logger.Info("initializing game systems (66 total)")
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

	// 16a. FactionAwareAISystem - bridges faction reputation with AI behavior
	// Makes neutral NPCs hostile when player reputation drops below -50
	factionAwareAISystem := NewFactionAwareAISystem(game.World)
	game.World.AddSystem(factionAwareAISystem)

	// 16b. StatusEffectAISystem - bridges status effects with AI behavior
	// Disables AI actions when entities are stunned, frozen, feared, or paralyzed
	statusEffectAISystem := NewStatusEffectAISystem(game.World, config.Seed+2200)
	game.World.AddSystem(statusEffectAISystem)

	// ========================================================================
	// PROGRESSION SYSTEMS (5 systems)
	// ========================================================================

	// 17. ProgressionSystem - XP and leveling
	progressionSystem := NewProgressionSystem(game.World)
	result.ProgressionSystem = progressionSystem
	game.World.AddSystem(progressionSystem)

	// 17a. FactionXPBonusSystem - awards bonus XP for killing enemies of allied factions
	// Connects FactionComponent reputation with ProgressionSystem XP rewards
	factionXPBonusSystem := NewFactionXPBonusSystem(game.World, config.Seed+5100)
	factionXPBonusSystem.SetFactionSystem(factionSystem)
	factionXPBonusSystem.SetProgressionSystem(progressionSystem)
	result.FactionXPBonusSystem = factionXPBonusSystem
	game.World.AddSystem(factionXPBonusSystem)

	// 17b. FactionDamageBonusSystem - bonus damage against enemies of allied factions
	// Connects FactionComponent reputation with CombatSystem damage modifiers
	factionDamageBonusSystem := NewFactionDamageBonusSystem(game.World, config.Seed+5150)
	factionDamageBonusSystem.SetFactionSystem(factionSystem)
	factionDamageBonusSystem.SetGenre(config.GenreID)
	result.FactionDamageBonusSystem = factionDamageBonusSystem
	game.World.AddSystem(factionDamageBonusSystem)

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
	result.ItemPickupSystem = itemPickupSystem
	game.World.AddSystem(itemPickupSystem)

	// 23-25. Magic systems
	game.World.AddSystem(spellCastingSystem)
	manaRegenSystem := &ManaRegenSystem{}
	game.World.AddSystem(manaRegenSystem)

	// 25a. SpecializationManaBoostSystem - passive mana regen bonuses from class specializations
	// Connects ClassProgressionComponent with ManaComponent for genre-aware caster bonuses
	specializationManaBoostSystem := NewSpecializationManaBoostSystem(game.World, config.Seed+6600)
	specializationManaBoostSystem.SetGenre(config.GenreID)
	result.SpecializationManaBoostSystem = specializationManaBoostSystem
	game.World.AddSystem(specializationManaBoostSystem)

	// 25b. SpecializationHealthRegenSystem - passive health regen bonuses from class specializations
	// Connects ClassProgressionComponent with HealthComponent for genre-aware tank/healer bonuses
	specializationHealthRegenSystem := NewSpecializationHealthRegenSystem(game.World, config.Seed+6650)
	specializationHealthRegenSystem.SetGenre(config.GenreID)
	result.SpecializationHealthRegenSystem = specializationHealthRegenSystem
	game.World.AddSystem(specializationHealthRegenSystem)

	// 25c. SpecializationSpellDamageSystem - passive spell damage bonuses from class specializations
	// Connects ClassProgressionComponent with spell damage for genre-aware caster bonuses
	specializationSpellDamageSystem := NewSpecializationSpellDamageSystem(game.World, config.Seed+6660)
	specializationSpellDamageSystem.SetGenre(config.GenreID)
	result.SpecializationSpellDamageSystem = specializationSpellDamageSystem
	game.World.AddSystem(specializationSpellDamageSystem)

	// 25d. SpecializationAttackSpeedSystem - passive attack speed bonuses from class specializations
	// Connects ClassProgressionComponent with AttackComponent cooldowns for genre-aware melee bonuses
	specializationAttackSpeedSystem := NewSpecializationAttackSpeedSystem(game.World, config.Seed+6670)
	specializationAttackSpeedSystem.SetGenre(config.GenreID)
	result.SpecializationAttackSpeedSystem = specializationAttackSpeedSystem
	game.World.AddSystem(specializationAttackSpeedSystem)

	// 25e. SpecializationDefenseSystem - passive defense bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.Defense for genre-aware tank bonuses
	specializationDefenseSystem := NewSpecializationDefenseSystem(game.World, config.Seed+6680)
	specializationDefenseSystem.SetGenre(config.GenreID)
	result.SpecializationDefenseSystem = specializationDefenseSystem
	game.World.AddSystem(specializationDefenseSystem)

	// 25f. SpecializationStatusResistSystem - status effect duration modifiers from class specializations
	// Connects ClassProgressionComponent with StatusEffectComponent for genre-aware debuff resistance
	specializationStatusResistSystem := NewSpecializationStatusResistSystem(game.World, config.Seed+6690)
	specializationStatusResistSystem.SetGenre(config.GenreID)
	result.SpecializationStatusResistSystem = specializationStatusResistSystem
	game.World.AddSystem(specializationStatusResistSystem)

	// 25h. SpecializationCritDamageSystem - critical hit damage bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.CritDamage for rogue/assassin bonuses
	specializationCritDamageSystem := NewSpecializationCritDamageSystem(game.World, config.Seed+6700)
	specializationCritDamageSystem.SetGenre(config.GenreID)
	result.SpecializationCritDamageSystem = specializationCritDamageSystem
	game.World.AddSystem(specializationCritDamageSystem)

	// 25i. SpecializationEvasionSystem - evasion bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.Evasion for rogue/monk bonuses
	specializationEvasionSystem := NewSpecializationEvasionSystem(game.World, config.Seed+6710)
	specializationEvasionSystem.SetGenre(config.GenreID)
	result.SpecializationEvasionSystem = specializationEvasionSystem
	game.World.AddSystem(specializationEvasionSystem)

	// 25g. StatusEffectHealthRegenSystem - health regeneration modifiers from status effects
	// Connects StatusEffectSystem with HealthComponent for genre-aware regen bonuses/penalties
	statusEffectHealthRegenSystem := NewStatusEffectHealthRegenSystem(game.World, config.Seed+6695)
	statusEffectHealthRegenSystem.SetGenre(config.GenreID)
	result.StatusEffectHealthRegenSystem = statusEffectHealthRegenSystem
	game.World.AddSystem(statusEffectHealthRegenSystem)

	// 26. InventorySystem - item management
	game.World.AddSystem(inventorySystem)

	// 27-29. Commerce and crafting
	result.CommerceSystem = NewCommerceSystemWithLogger(game.World, inventorySystem, logger)
	game.World.AddSystem(result.CommerceSystem)

	// 29a. ReputationPricingSystem - adjusts merchant prices based on player faction reputation
	// Connects FactionComponent reputation with MerchantComponent pricing
	reputationPricingSystem := NewReputationPricingSystem(game.World, config.Seed+5000)
	game.World.AddSystem(reputationPricingSystem)

	result.DialogSystem = NewDialogSystemWithLogger(game.World, logger)
	game.World.AddSystem(result.DialogSystem)

	result.CraftingSystem = NewCraftingSystem(game.World, inventorySystem, nil) // itemGen set later
	game.World.AddSystem(result.CraftingSystem)

	// 30. InteractionSystem - puzzle element interactions (Phase 11.2)
	result.InteractionSystem = NewInteractionSystem(game.World)
	game.World.AddSystem(result.InteractionSystem)

	// 30a. MiniGameSystem - manages mini-game lifecycle (Phase 27.1)
	// Note: This is NOT added as a System because it has a custom Update signature.
	// It will be called manually from the game loop.
	miniGameSystem := NewMiniGameSystem(game.World)
	miniGameSystem.SetSeed(config.Seed)

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

	// 36a. WeatherCombatSystem - applies weather-based status effects
	weatherCombatSystem := NewWeatherCombatSystem(game.World)
	game.World.AddSystem(weatherCombatSystem)

	// 36b. StatusEffectLightingSystem - visual feedback for status effects via lighting
	statusEffectLightingSystem := NewStatusEffectLightingSystem(game.World, config.Seed+2000)
	game.World.AddSystem(statusEffectLightingSystem)

	// 36b2. StatusEffectMovementSystem - applies movement speed modifiers from status effects
	// Connects WeatherCombatSystem (chilled, wet) with MovementSystem velocity scaling
	statusEffectMovementSystem := NewStatusEffectMovementSystem(game.World, config.Seed+2100)
	game.World.AddSystem(statusEffectMovementSystem)

	// 36b2b. StatusEffectEvasionSystem - applies evasion modifiers from status effects
	// Connects StatusEffectSystem with CombatSystem for genre-aware evasion bonuses/penalties
	statusEffectEvasionSystem := NewStatusEffectEvasionSystem(game.World, config.Seed+2120)
	statusEffectEvasionSystem.SetGenre(config.GenreID)
	result.StatusEffectEvasionSystem = statusEffectEvasionSystem
	game.World.AddSystem(statusEffectEvasionSystem)

	// 36b2c. StatusEffectCriticalChanceSystem - applies crit chance modifiers from status effects
	// Connects StatusEffectSystem with CombatSystem for genre-aware critical hit bonuses/penalties
	statusEffectCriticalChanceSystem := NewStatusEffectCriticalChanceSystem(game.World, config.Seed+2130)
	statusEffectCriticalChanceSystem.SetGenre(config.GenreID)
	result.StatusEffectCriticalChanceSystem = statusEffectCriticalChanceSystem
	game.World.AddSystem(statusEffectCriticalChanceSystem)

	// 36b2d. StatusEffectDamageBoostSystem - applies damage modifiers from status effects
	// Connects StatusEffectSystem with CombatSystem for genre-aware damage bonuses/penalties
	statusEffectDamageBoostSystem := NewStatusEffectDamageBoostSystem(game.World, config.Seed+2140)
	statusEffectDamageBoostSystem.SetGenre(config.GenreID)
	result.StatusEffectDamageBoostSystem = statusEffectDamageBoostSystem
	game.World.AddSystem(statusEffectDamageBoostSystem)

	// 36b3. TerrainMovementSpeedSystem - applies movement speed modifiers from terrain type
	// Connects terrain generation with MovementSystem for genre-aware terrain effects
	terrainMovementSpeedSystem := NewTerrainMovementSpeedSystem(game.World, config.Seed+2150)
	terrainMovementSpeedSystem.SetGenre(config.GenreID)
	terrainMovementSpeedSystem.SetTileSize(config.TileSize)
	result.TerrainMovementSpeedSystem = terrainMovementSpeedSystem
	game.World.AddSystem(terrainMovementSpeedSystem)

	// 36b4. TerrainCombatBonusSystem - applies combat stat modifiers from terrain type
	// Connects terrain generation with CombatSystem for tactical bonuses (high ground, cover, water)
	terrainCombatBonusSystem := NewTerrainCombatBonusSystem(game.World, config.Seed+2160)
	terrainCombatBonusSystem.SetGenre(config.GenreID)
	terrainCombatBonusSystem.SetTileSize(config.TileSize)
	result.TerrainCombatBonusSystem = terrainCombatBonusSystem
	game.World.AddSystem(terrainCombatBonusSystem)

	// 36b4b. TerrainCombatBonusParticleSystem - visual feedback for terrain combat bonuses
	// Connects TerrainCombatBonusSystem with ParticleSystem for genre-aware high ground/cover particles
	terrainCombatBonusParticleSystem := NewTerrainCombatBonusParticleSystem(game.World, config.Seed+2165)
	terrainCombatBonusParticleSystem.SetTerrainCombatBonusSystem(terrainCombatBonusSystem)
	terrainCombatBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	terrainCombatBonusParticleSystem.SetGenre(config.GenreID)
	result.TerrainCombatBonusParticleSystem = terrainCombatBonusParticleSystem
	game.World.AddSystem(terrainCombatBonusParticleSystem)

	// 36b5. TerrainStealthSystem - modifies AI detection based on target terrain
	// Connects terrain generation with AISystem for tactical stealth gameplay
	terrainStealthSystem := NewTerrainStealthSystem(game.World, config.Seed+2170)
	terrainStealthSystem.SetGenre(config.GenreID)
	terrainStealthSystem.SetTileSize(config.TileSize)
	result.TerrainStealthSystem = terrainStealthSystem
	game.World.AddSystem(terrainStealthSystem)

	// 36b5b. TerrainAmbushCritSystem - grants critical hit bonuses for terrain concealment
	// Connects TerrainStealthSystem with StatsComponent for tactical ambush attacks
	terrainAmbushCritSystem := NewTerrainAmbushCritSystem(game.World, config.Seed+2175)
	terrainAmbushCritSystem.SetTerrainStealthSystem(terrainStealthSystem)
	terrainAmbushCritSystem.SetGenre(config.GenreID)
	result.TerrainAmbushCritSystem = terrainAmbushCritSystem
	game.World.AddSystem(terrainAmbushCritSystem)

	// 36b5c. StealthIndicatorParticleSystem - visual feedback for terrain cover state changes
	// Connects TerrainStealthSystem with ParticleSystem for genre-aware stealth particles
	stealthIndicatorParticleSystem := NewStealthIndicatorParticleSystem(game.World, config.Seed+2177)
	stealthIndicatorParticleSystem.SetTerrainStealthSystem(terrainStealthSystem)
	stealthIndicatorParticleSystem.SetParticleSystem(result.ParticleSystem)
	stealthIndicatorParticleSystem.SetGenre(config.GenreID)
	result.StealthIndicatorParticleSystem = stealthIndicatorParticleSystem
	game.World.AddSystem(stealthIndicatorParticleSystem)

	// 36b6. TerrainStatusEffectSystem - applies elemental status effects from terrain type
	// Connects terrain tiles (water, lava) with StatusEffectSystem for elemental combos
	terrainStatusEffectSystem := NewTerrainStatusEffectSystem(game.World, config.Seed+2180)
	terrainStatusEffectSystem.SetGenre(config.GenreID)
	terrainStatusEffectSystem.SetTileSize(config.TileSize)
	result.TerrainStatusEffectSystem = terrainStatusEffectSystem
	game.World.AddSystem(terrainStatusEffectSystem)

	// 36b7. TerrainManaRegenSystem - modifies mana regeneration based on terrain type
	// Connects terrain tiles (water, platforms) with ManaComponent for tactical spellcasting
	terrainManaRegenSystem := NewTerrainManaRegenSystem(game.World, config.Seed+2190)
	terrainManaRegenSystem.SetGenre(config.GenreID)
	terrainManaRegenSystem.SetTileSize(config.TileSize)
	result.TerrainManaRegenSystem = terrainManaRegenSystem
	game.World.AddSystem(terrainManaRegenSystem)

	// 36b7b. TerrainSpellDamageSystem - modifies spell damage based on terrain type
	// Connects terrain tiles (lava, water, platforms) with spell damage for tactical spellcasting
	terrainSpellDamageSystem := NewTerrainSpellDamageSystem(game.World, config.Seed+2192)
	terrainSpellDamageSystem.SetGenre(config.GenreID)
	terrainSpellDamageSystem.SetTileSize(config.TileSize)
	result.TerrainSpellDamageSystem = terrainSpellDamageSystem
	game.World.AddSystem(terrainSpellDamageSystem)

	// 36b8. TerrainCompanionBonusSystem - modifies companion stats based on terrain type
	// Connects terrain tiles with CompanionStatsComponent for tactical companion positioning
	terrainCompanionBonusSystem := NewTerrainCompanionBonusSystem(game.World, config.Seed+2195)
	terrainCompanionBonusSystem.SetGenre(config.GenreID)
	terrainCompanionBonusSystem.SetTileSize(config.TileSize)
	result.TerrainCompanionBonusSystem = terrainCompanionBonusSystem
	game.World.AddSystem(terrainCompanionBonusSystem)

	// 36b9. FactionCompanionBehaviorSystem - modifies companion targeting based on faction reputation
	// Connects FactionSystem reputation with companion AI targeting for faction-aware companions
	factionCompanionBehaviorSystem := NewFactionCompanionBehaviorSystem(game.World, config.Seed+2198)
	factionCompanionBehaviorSystem.SetGenre(config.GenreID)
	result.FactionCompanionBehaviorSystem = factionCompanionBehaviorSystem
	game.World.AddSystem(factionCompanionBehaviorSystem)

	// 36b10. WeatherCompanionBonusSystem - modifies companion stats based on weather conditions
	// Connects WeatherSystem with CompanionStatsComponent for tactical companion positioning
	weatherCompanionBonusSystem := NewWeatherCompanionBonusSystem(game.World, config.Seed+2200)
	weatherCompanionBonusSystem.SetGenre(config.GenreID)
	result.WeatherCompanionBonusSystem = weatherCompanionBonusSystem
	game.World.AddSystem(weatherCompanionBonusSystem)

	// 36c. CriticalHitParticleSystem - visual feedback for critical hits
	criticalHitParticleSystem := NewCriticalHitParticleSystem(game.World, config.Seed+3000)
	criticalHitParticleSystem.SetParticleSystem(result.ParticleSystem)
	criticalHitParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetCriticalHitCallback(criticalHitParticleSystem.OnCriticalHit)
	game.World.AddSystem(criticalHitParticleSystem)

	// 36d. LevelUpParticleSystem - visual feedback for level-ups
	levelUpParticleSystem := NewLevelUpParticleSystem(game.World, config.Seed+4000)
	levelUpParticleSystem.SetParticleSystem(result.ParticleSystem)
	levelUpParticleSystem.SetGenre(config.GenreID)
	result.ProgressionSystem.AddLevelUpCallback(levelUpParticleSystem.OnLevelUp)
	result.LevelUpParticleSystem = levelUpParticleSystem
	game.World.AddSystem(levelUpParticleSystem)

	// 36d2. CompanionProgressionSystem - companion XP and leveling
	companionProgressionSystem := NewCompanionProgressionSystem(game.World)
	result.CompanionProgressionSystem = companionProgressionSystem
	result.CompanionProgressionSystemWrapper = &companionProgressionSystemWrapper{system: companionProgressionSystem}
	game.World.AddSystem(result.CompanionProgressionSystemWrapper)

	// 36d3. CompanionLevelUpParticleSystem - visual feedback for companion level-ups
	companionLevelUpParticleSystem := NewCompanionLevelUpParticleSystem(game.World, config.Seed+4100)
	companionLevelUpParticleSystem.SetParticleSystem(result.ParticleSystem)
	companionLevelUpParticleSystem.SetGenre(config.GenreID)
	companionProgressionSystem.AddLevelUpCallback(companionLevelUpParticleSystem.OnCompanionLevelUp)
	result.CompanionLevelUpParticleSystem = companionLevelUpParticleSystem
	game.World.AddSystem(companionLevelUpParticleSystem)

	// 36e. ItemPickupParticleSystem - visual feedback for item pickups
	itemPickupParticleSystem := NewItemPickupParticleSystem(game.World, config.Seed+4500)
	itemPickupParticleSystem.SetParticleSystem(result.ParticleSystem)
	itemPickupParticleSystem.SetGenre(config.GenreID)
	result.ItemPickupSystem.SetPickupCallback(itemPickupParticleSystem.OnItemPickup)
	result.ItemPickupParticleSystem = itemPickupParticleSystem
	game.World.AddSystem(itemPickupParticleSystem)

	// 36f. SpellEffectParticleSystem - visual feedback for spell effects
	spellEffectParticleSystem := NewSpellEffectParticleSystem(game.World, config.Seed+5500)
	spellEffectParticleSystem.SetParticleSystem(result.ParticleSystem)
	spellEffectParticleSystem.SetGenre(config.GenreID)
	result.SpellEffectParticleSystem = spellEffectParticleSystem
	game.World.AddSystem(spellEffectParticleSystem)

	// 36f2. DeathParticleSystem - visual feedback for entity deaths
	// Spawns genre-aware particles (smoke, debris, blood) when entities die
	deathParticleSystem := NewDeathParticleSystem(game.World, config.Seed+5600)
	deathParticleSystem.SetParticleSystem(result.ParticleSystem)
	deathParticleSystem.SetGenre(config.GenreID)
	result.DeathParticleSystem = deathParticleSystem
	game.World.AddSystem(deathParticleSystem)

	// 36f3. DamageResistanceParticleSystem - visual feedback for damage resistance
	// Spawns genre-aware particles when damage is significantly reduced by resistances
	damageResistanceParticleSystem := NewDamageResistanceParticleSystem(game.World, config.Seed+5700)
	damageResistanceParticleSystem.SetParticleSystem(result.ParticleSystem)
	damageResistanceParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetDamageResistedCallback(damageResistanceParticleSystem.OnDamageResisted)
	result.DamageResistanceParticleSystem = damageResistanceParticleSystem
	game.World.AddSystem(damageResistanceParticleSystem)

	// 36f4. ShieldAbsorbParticleSystem - visual feedback for shield damage absorption
	// Spawns genre-aware particles when shields absorb incoming damage
	shieldAbsorbParticleSystem := NewShieldAbsorbParticleSystem(game.World, config.Seed+5750)
	shieldAbsorbParticleSystem.SetParticleSystem(result.ParticleSystem)
	shieldAbsorbParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetShieldAbsorbCallback(shieldAbsorbParticleSystem.OnShieldAbsorb)
	result.ShieldAbsorbParticleSystem = shieldAbsorbParticleSystem
	game.World.AddSystem(shieldAbsorbParticleSystem)

	// 36g. WeatherGroundEffectSystem - visual feedback for weather ground impacts
	// Spawns rain splashes, snow puffs, dust clouds when weather particles hit ground
	weatherGroundEffectSystem := NewWeatherGroundEffectSystem(game.World, config.Seed+6000)
	weatherGroundEffectSystem.SetParticleSystem(result.ParticleSystem)
	weatherGroundEffectSystem.SetGenre(config.GenreID)
	game.World.AddSystem(weatherGroundEffectSystem)

	// 36h. WeatherAudioSystem - audio feedback for weather conditions
	// Plays ambient sounds (rain, wind, etc.) when weather is active
	weatherAudioSystem := NewWeatherAudioSystem(game.World, config.Seed+6100)
	weatherAudioSystem.SetAudioManager(result.AudioManager)
	weatherAudioSystem.SetGenre(config.GenreID)
	result.WeatherAudioSystem = weatherAudioSystem
	game.World.AddSystem(weatherAudioSystem)

	// 36i. WeatherAwareAISystem - modifies AI detection ranges based on weather
	// Connects WeatherSystem with AISystem for tactical stealth gameplay
	weatherAwareAISystem := NewWeatherAwareAISystem(game.World, config.Seed+6200)
	game.World.AddSystem(weatherAwareAISystem)

	// 36j. WeatherManaRegenSystem - modifies mana regeneration based on weather
	// Connects WeatherSystem with ManaComponent for magical gameplay synergy
	weatherManaRegenSystem := NewWeatherManaRegenSystem(game.World, config.Seed+6300)
	weatherManaRegenSystem.SetGenre(config.GenreID)
	result.WeatherManaRegenSystem = weatherManaRegenSystem
	game.World.AddSystem(weatherManaRegenSystem)

	// 36j2. WeatherCooldownSystem - modifies spell cooldown rates based on weather
	// Connects WeatherSystem with SpellCastingSystem for magical gameplay synergy
	weatherCooldownSystem := NewWeatherCooldownSystem(game.World, config.Seed+6350)
	weatherCooldownSystem.SetGenre(config.GenreID)
	result.WeatherCooldownSystem = weatherCooldownSystem
	game.World.AddSystem(weatherCooldownSystem)

	// 36j3. FishingWeatherBonusSystem - modifies fishing bonuses based on weather
	// Connects WeatherSystem with FishingSystem for immersive fishing gameplay
	fishingWeatherBonusSystem := NewFishingWeatherBonusSystem(game.World, config.Seed+6375)
	fishingWeatherBonusSystem.SetGenre(config.GenreID)
	result.FishingWeatherBonusSystem = fishingWeatherBonusSystem
	game.World.AddSystem(fishingWeatherBonusSystem)

	// 36k. LowHealthVFXSystem - visual feedback for low player health
	// Spawns genre-aware pulsing particles when player health is critically low
	lowHealthVFXSystem := NewLowHealthVFXSystem(game.World, config.Seed+6400)
	lowHealthVFXSystem.SetParticleSystem(result.ParticleSystem)
	lowHealthVFXSystem.SetGenre(config.GenreID)
	result.LowHealthVFXSystem = lowHealthVFXSystem
	game.World.AddSystem(lowHealthVFXSystem)

	// 36k2. ManaRegenParticleSystem - visual feedback for mana regeneration
	// Spawns genre-aware particles when entities actively regenerate mana
	manaRegenParticleSystem := NewManaRegenParticleSystem(game.World, config.Seed+6450)
	manaRegenParticleSystem.SetParticleSystem(result.ParticleSystem)
	manaRegenParticleSystem.SetGenre(config.GenreID)
	result.ManaRegenParticleSystem = manaRegenParticleSystem
	game.World.AddSystem(manaRegenParticleSystem)

	// 36l. CompanionAuraParticleSystem - visual feedback for companion bonding perks
	// Spawns genre-aware aura particles around companions with active bonding perks
	companionAuraParticleSystem := NewCompanionAuraParticleSystem(game.World, config.Seed+6500)
	companionAuraParticleSystem.SetParticleSystem(result.ParticleSystem)
	companionAuraParticleSystem.SetGenre(config.GenreID)
	result.CompanionAuraParticleSystem = companionAuraParticleSystem
	game.World.AddSystem(companionAuraParticleSystem)

	// 36m. ElementalComboParticleSystem - visual feedback for elemental status combos
	// Spawns genre-aware particles when elemental effects combine (fire+ice=steam, etc.)
	elementalComboParticleSystem := NewElementalComboParticleSystem(game.World, config.Seed+6700)
	elementalComboParticleSystem.SetParticleSystem(result.ParticleSystem)
	elementalComboParticleSystem.SetGenre(config.GenreID)
	result.ElementalComboParticleSystem = elementalComboParticleSystem
	game.World.AddSystem(elementalComboParticleSystem)

	// 36n. WeatherRangedAccuracySystem - modifies ranged attack accuracy based on weather
	// Connects WeatherSystem with ranged combat for tactical depth in poor visibility
	weatherRangedAccuracySystem := NewWeatherRangedAccuracySystem(game.World, config.Seed+6750)
	weatherRangedAccuracySystem.SetGenre(config.GenreID)
	result.WeatherRangedAccuracySystem = weatherRangedAccuracySystem
	game.World.AddSystem(weatherRangedAccuracySystem)

	// 36n2. WeatherCritChanceSystem - modifies critical hit chance based on weather conditions
	// Connects WeatherSystem with StatsComponent.CritChance for tactical crit bonuses/penalties
	weatherCritChanceSystem := NewWeatherCritChanceSystem(game.World, config.Seed+6775)
	weatherCritChanceSystem.SetGenre(config.GenreID)
	result.WeatherCritChanceSystem = weatherCritChanceSystem
	game.World.AddSystem(weatherCritChanceSystem)

	// 36o. WeatherXPBonusSystem - modifies XP gains based on weather conditions
	// Connects WeatherSystem with ProgressionSystem for genre-aware XP bonuses
	weatherXPBonusSystem := NewWeatherXPBonusSystem(game.World, config.Seed+6800)
	weatherXPBonusSystem.SetGenre(config.GenreID)
	weatherXPBonusSystem.SetProgressionSystem(progressionSystem)
	result.WeatherXPBonusSystem = weatherXPBonusSystem
	game.World.AddSystem(weatherXPBonusSystem)

	// 36p. ElementalComboDamageSystem - applies bonus damage when elemental status effects combine
	// Connects StatusEffectSystem with CombatSystem for genre-aware elemental combo damage
	elementalComboDamageSystem := NewElementalComboDamageSystem(game.World, config.Seed+6850)
	elementalComboDamageSystem.SetGenre(config.GenreID)
	result.ElementalComboDamageSystem = elementalComboDamageSystem
	game.World.AddSystem(elementalComboDamageSystem)

	// 36q. ElementalCompanionSynergySystem - boosts elemental companions based on owner status effects
	// Connects CompanionComponent (elemental type) with StatusEffectComponent for tactical synergy
	elementalCompanionSynergySystem := NewElementalCompanionSynergySystem(game.World, config.Seed+6900)
	elementalCompanionSynergySystem.SetGenre(config.GenreID)
	result.ElementalCompanionSynergySystem = elementalCompanionSynergySystem
	game.World.AddSystem(elementalCompanionSynergySystem)

	// 36q2. CompanionSpellAmplificationSystem - boosts owner spell damage/healing based on companion bonding
	// Connects CompanionComponent (loyalty, bonding perks) with owner spell effectiveness
	companionSpellAmplificationSystem := NewCompanionSpellAmplificationSystem(game.World, config.Seed+6910)
	companionSpellAmplificationSystem.SetGenre(config.GenreID)
	result.CompanionSpellAmplificationSystem = companionSpellAmplificationSystem
	game.World.AddSystem(companionSpellAmplificationSystem)

	// 36q3. CompanionManaRegenSystem - boosts owner mana regeneration based on companion bonding
	// Connects CompanionComponent (loyalty, type, perks) with ManaComponent for sustained spellcasting
	companionManaRegenSystem := NewCompanionManaRegenSystem(game.World, config.Seed+6920)
	companionManaRegenSystem.SetGenre(config.GenreID)
	result.CompanionManaRegenSystem = companionManaRegenSystem
	game.World.AddSystem(companionManaRegenSystem)

	// 36r. LifestealSystem - heals attackers based on damage dealt
	// Connects CombatSystem damage events with HealthComponent healing for sustained combat
	lifestealSystem := NewLifestealSystem(game.World, config.Seed+6950)
	lifestealSystem.SetParticleSystem(result.ParticleSystem)
	lifestealSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetDamageCallback(lifestealSystem.OnDamageDealt)
	result.LifestealSystem = lifestealSystem
	game.World.AddSystem(lifestealSystem)

	// 36s. StatusEffectManaCostSystem - modifies spell mana costs based on status effects
	// Connects StatusEffectSystem with SpellCastingSystem for tactical spell management
	statusEffectManaCostSystem := NewStatusEffectManaCostSystem(game.World, config.Seed+7000)
	statusEffectManaCostSystem.SetGenre(config.GenreID)
	result.StatusEffectManaCostSystem = statusEffectManaCostSystem
	game.World.AddSystem(statusEffectManaCostSystem)

	// 36t. StatusEffectDamageParticleSystem - visual feedback for status effect damage ticks
	// Connects StatusEffectSystem tick events (burning, poison, regen) with ParticleSystem
	statusEffectDamageParticleSystem := NewStatusEffectDamageParticleSystem(game.World, config.Seed+7050)
	statusEffectDamageParticleSystem.SetParticleSystem(result.ParticleSystem)
	statusEffectDamageParticleSystem.SetGenre(config.GenreID)
	statusEffectSystem.SetTickCallback(statusEffectDamageParticleSystem.OnStatusEffectTick)
	result.StatusEffectDamageParticleSystem = statusEffectDamageParticleSystem
	game.World.AddSystem(statusEffectDamageParticleSystem)

	// 36u. WeaponSwingParticleSystem - visual feedback for melee weapon attacks
	// Connects CombatSystem damage events with ParticleSystem for rarity-aware weapon trails
	weaponSwingParticleSystem := NewWeaponSwingParticleSystem(game.World, config.Seed+7100)
	weaponSwingParticleSystem.SetParticleSystem(result.ParticleSystem)
	weaponSwingParticleSystem.SetGenre(config.GenreID)
	result.WeaponSwingParticleSystem = weaponSwingParticleSystem
	game.World.AddSystem(weaponSwingParticleSystem)

	// 36v. FearFleeParticleSystem - visual feedback for feared fleeing entities
	// Connects StatusEffectAISystem flee behavior with ParticleSystem for genre-aware fear effects
	fearFleeParticleSystem := NewFearFleeParticleSystem(game.World, config.Seed+7150)
	fearFleeParticleSystem.SetParticleSystem(result.ParticleSystem)
	fearFleeParticleSystem.SetGenre(config.GenreID)
	result.FearFleeParticleSystem = fearFleeParticleSystem
	game.World.AddSystem(fearFleeParticleSystem)

	// 36w. EvasionParticleSystem - visual feedback for dodged attacks
	// Connects CombatSystem evasion events with ParticleSystem for genre-aware dodge effects
	evasionParticleSystem := NewEvasionParticleSystem(game.World, config.Seed+7200)
	evasionParticleSystem.SetParticleSystem(result.ParticleSystem)
	evasionParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetEvasionCallback(evasionParticleSystem.OnEvasion)
	result.EvasionParticleSystem = evasionParticleSystem
	game.World.AddSystem(evasionParticleSystem)

	// 36w2. BlockParticleSystem - visual feedback for blocked attacks
	// Connects CombatSystem block events with ParticleSystem for genre-aware shield particles
	blockParticleSystem := NewBlockParticleSystem(game.World, config.Seed+7225)
	blockParticleSystem.SetParticleSystem(result.ParticleSystem)
	blockParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetBlockCallback(blockParticleSystem.OnBlock)
	result.BlockParticleSystem = blockParticleSystem
	game.World.AddSystem(blockParticleSystem)

	// 36x. FootstepParticleSystem - visual feedback for terrain-aware movement
	// Connects MovementSystem with ParticleSystem for genre-aware footstep effects
	footstepParticleSystem := NewFootstepParticleSystem(game.World, config.Seed+7250)
	footstepParticleSystem.SetParticleSystem(result.ParticleSystem)
	footstepParticleSystem.SetGenre(config.GenreID)
	footstepParticleSystem.SetTileSize(config.TileSize)
	result.FootstepParticleSystem = footstepParticleSystem
	game.World.AddSystem(footstepParticleSystem)

	// 36y. HealingParticleSystem - visual feedback for healing effects
	// Connects CombatSystem heal events with ParticleSystem for genre-aware heal particles
	healingParticleSystem := NewHealingParticleSystem(game.World, config.Seed+7300)
	healingParticleSystem.SetParticleSystem(result.ParticleSystem)
	healingParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetHealCallback(healingParticleSystem.OnHeal)
	result.HealingParticleSystem = healingParticleSystem
	game.World.AddSystem(healingParticleSystem)

	// 36z. ShieldRegenSystem - passive shield regeneration with visual feedback
	// Connects ShieldComponent with ParticleSystem for genre-aware shield recovery particles
	shieldRegenSystem := NewShieldRegenSystem(game.World, config.Seed+7350)
	shieldRegenSystem.SetParticleSystem(result.ParticleSystem)
	shieldRegenSystem.SetGenre(config.GenreID)
	result.ShieldRegenSystem = shieldRegenSystem
	game.World.AddSystem(shieldRegenSystem)

	// 36ab. DrowningParticleSystem - visual feedback for drowning entities
	// Connects SwimmingComponent.Drowning state with ParticleSystem for genre-aware suffocation effects
	drowningParticleSystem := NewDrowningParticleSystem(game.World, config.Seed+7450)
	drowningParticleSystem.SetParticleSystem(result.ParticleSystem)
	drowningParticleSystem.SetGenre(config.GenreID)
	result.DrowningParticleSystem = drowningParticleSystem
	game.World.AddSystem(drowningParticleSystem)

	// 36aa. WeatherMeleeDamageSystem - modifies physical/melee damage based on weather
	// Connects WeatherSystem with CombatSystem for genre-aware melee damage modifiers
	weatherMeleeDamageSystem := NewWeatherMeleeDamageSystem(game.World, config.Seed+7400)
	weatherMeleeDamageSystem.SetGenre(config.GenreID)
	result.WeatherMeleeDamageSystem = weatherMeleeDamageSystem
	game.World.AddSystem(weatherMeleeDamageSystem)

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

	// 39b. DestructionParticleSystem - visual feedback for destructible objects
	// Connects DestructibleObjectSystem with ParticleSystem for genre-aware destruction effects
	destructionParticleSystem := NewDestructionParticleSystem(game.World, config.Seed+1110)
	destructionParticleSystem.SetParticleSystem(result.ParticleSystem)
	destructionParticleSystem.SetGenre(config.GenreID)
	result.DestructionParticleSystem = destructionParticleSystem
	game.World.AddSystem(destructionParticleSystem)

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

	// 43b. TimeOfDayLightingSystem - day/night ambient light modulation (Phase 17.3)
	// Connects palette/timeofday.go with AmbientLightComponent for dynamic lighting
	timeOfDayLightingSystem := NewTimeOfDayLightingSystem(game.World, config.Seed+7500)
	timeOfDayLightingSystem.SetGenre(config.GenreID)
	result.TimeOfDayLightingSystem = timeOfDayLightingSystem
	game.World.AddSystem(timeOfDayLightingSystem)

	// 43c. TimeOfDayStealthSystem - day/night AI detection modulation
	// Connects TimeOfDayLightingSystem with AIComponent detection ranges for stealth
	timeOfDayStealthSystem := NewTimeOfDayStealthSystem(game.World, config.Seed+7550)
	timeOfDayStealthSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayStealthSystem.SetGenre(config.GenreID)
	result.TimeOfDayStealthSystem = timeOfDayStealthSystem
	game.World.AddSystem(timeOfDayStealthSystem)

	// 43d. TimeOfDayXPBonusSystem - day/night XP bonus modulation
	// Connects TimeOfDayLightingSystem with ProgressionSystem for time-based XP bonuses
	timeOfDayXPBonusSystem := NewTimeOfDayXPBonusSystem(game.World, config.Seed+7575)
	timeOfDayXPBonusSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayXPBonusSystem.SetProgressionSystem(progressionSystem)
	timeOfDayXPBonusSystem.SetGenre(config.GenreID)
	result.TimeOfDayXPBonusSystem = timeOfDayXPBonusSystem
	game.World.AddSystem(timeOfDayXPBonusSystem)

	// 43e. TimeOfDayManaCostSystem - day/night spell mana cost modulation
	// Connects TimeOfDayLightingSystem with SpellCastingSystem for element-based mana costs
	timeOfDayManaCostSystem := NewTimeOfDayManaCostSystem(game.World, config.Seed+7600)
	timeOfDayManaCostSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayManaCostSystem.SetGenre(config.GenreID)
	result.TimeOfDayManaCostSystem = timeOfDayManaCostSystem
	game.World.AddSystem(timeOfDayManaCostSystem)

	// 43f. TimeOfDayCriticalChanceSystem - day/night critical hit chance modulation
	// Connects TimeOfDayLightingSystem with StatsComponent.CritChance for ambush tactics
	timeOfDayCriticalChanceSystem := NewTimeOfDayCriticalChanceSystem(game.World, config.Seed+7625)
	timeOfDayCriticalChanceSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayCriticalChanceSystem.SetGenre(config.GenreID)
	result.TimeOfDayCriticalChanceSystem = timeOfDayCriticalChanceSystem
	game.World.AddSystem(timeOfDayCriticalChanceSystem)

	// 43g. TimeOfDayCompanionBonusSystem - day/night companion stat modulation
	// Connects TimeOfDayLightingSystem with CompanionStatsComponent for companion bonuses
	timeOfDayCompanionBonusSystem := NewTimeOfDayCompanionBonusSystem(game.World, config.Seed+7650)
	timeOfDayCompanionBonusSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayCompanionBonusSystem.SetGenre(config.GenreID)
	result.TimeOfDayCompanionBonusSystem = timeOfDayCompanionBonusSystem
	game.World.AddSystem(timeOfDayCompanionBonusSystem)

	// 43h. TimeOfDayHealthRegenSystem - day/night health regeneration modulation
	// Connects TimeOfDayLightingSystem with HealthComponent for genre-aware health regen
	timeOfDayHealthRegenSystem := NewTimeOfDayHealthRegenSystem(game.World, config.Seed+7675)
	timeOfDayHealthRegenSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayHealthRegenSystem.SetGenre(config.GenreID)
	result.TimeOfDayHealthRegenSystem = timeOfDayHealthRegenSystem
	game.World.AddSystem(timeOfDayHealthRegenSystem)

	// 43i. TimeOfDayManaRegenSystem - day/night mana regeneration modulation
	// Connects TimeOfDayLightingSystem with ManaComponent for genre-aware mana regen
	timeOfDayManaRegenSystem := NewTimeOfDayManaRegenSystem(game.World, config.Seed+7700)
	timeOfDayManaRegenSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayManaRegenSystem.SetGenre(config.GenreID)
	result.TimeOfDayManaRegenSystem = timeOfDayManaRegenSystem
	game.World.AddSystem(timeOfDayManaRegenSystem)

	// 43j. TimeOfDayBlockChanceSystem - day/night block chance modulation
	// Connects TimeOfDayLightingSystem with StatsComponent.BlockChance for defensive timing
	timeOfDayBlockChanceSystem := NewTimeOfDayBlockChanceSystem(game.World, config.Seed+7725)
	timeOfDayBlockChanceSystem.SetLightingSystem(timeOfDayLightingSystem)
	timeOfDayBlockChanceSystem.SetGenre(config.GenreID)
	result.TimeOfDayBlockChanceSystem = timeOfDayBlockChanceSystem
	game.World.AddSystem(timeOfDayBlockChanceSystem)

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
	// LOCK-PICKING INTEGRATION: Connect InteractionSystem to MiniGameSystem for lock-picking
	result.InteractionSystem.SetMiniGameSystem(miniGameSystem)

	// Store MiniGameSystem reference in result for external access
	result.MiniGameSystem = miniGameSystem

	// Store UI system references in game
	game.TutorialSystem = result.TutorialSystem
	game.HelpSystem = result.HelpSystem

	// Wire audio manager to game
	game.SetAudioManager(result.AudioManager)

	if config.EnableVerboseLogging {
		logger.WithFields(logrus.Fields{
			"systemCount": 65,
			"seed":        config.Seed,
			"genre":       config.GenreID,
		}).Info("game systems initialized successfully (65th system requires terrain)")
	}

	return result, nil
}

// InitializeSpatialPartitionSystem initializes the SpatialPartitionSystem (system #65)
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

// companionProgressionSystemWrapper adapts CompanionProgressionSystem to System interface
type companionProgressionSystemWrapper struct {
	system *CompanionProgressionSystem
}

func (w *companionProgressionSystemWrapper) Update(entities []*Entity, deltaTime float64) {
	w.system.Update(deltaTime)
}
