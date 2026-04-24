// Package engine provides shared system initialization for all platforms.
// This file contains the centralized system initialization logic used by
// desktop, WASM, and mobile platforms to ensure Version 2.0 feature parity.
package engine

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
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
	InputSystem                                 *InputSystem
	CombatSystem                                *CombatSystem
	CollisionSystem                             *CollisionSystem
	ProjectileSystem                            *ProjectileSystem
	AudioManager                                *AudioManager
	ObjectiveTracker                            *ObjectiveTrackerSystem
	CommerceSystem                              *CommerceSystem
	DialogSystem                                *DialogSystem
	CraftingSystem                              *CraftingSystem
	InteractionSystem                           *InteractionSystem
	MiniGameSystem                              *MiniGameSystem
	AnimationSystem                             *AnimationSystem
	ParticleSystem                              *ParticleSystem
	TutorialSystem                              *EbitenTutorialSystem
	HelpSystem                                  *EbitenHelpSystem
	LevelUpParticleSystem                       *LevelUpParticleSystem
	ItemPickupParticleSystem                    *ItemPickupParticleSystem
	SpellEffectParticleSystem                   *SpellEffectParticleSystem
	DeathParticleSystem                         *DeathParticleSystem
	DamageResistanceParticleSystem              *DamageResistanceParticleSystem
	ShieldAbsorbParticleSystem                  *ShieldAbsorbParticleSystem
	CompanionLevelUpParticleSystem              *CompanionLevelUpParticleSystem
	ItemPickupSystem                            *ItemPickupSystem
	ProgressionSystem                           *ProgressionSystem
	SkillLoadoutSystem                          *SkillLoadoutSystem
	WeaponMasterySystem                         *WeaponMasterySystem
	ClassAffinitySystem                         *ClassAffinitySystem
	AttributeAllocationSystem                   *AttributeAllocationSystem
	TalentSystem                                *TalentSystem
	SkillMutationSystem                         *SkillMutationSystem
	BuildTemplateSystem                         *BuildTemplateSystem
	EquipmentSetBonusSystem                     *EquipmentSetBonusSystem
	ElementalWeaponEffectSystem                 *ElementalWeaponEffectSystem
	CompanionProgressionSystem                  *CompanionProgressionSystem
	WeatherAudioSystem                          *WeatherAudioSystem
	FactionXPBonusSystem                        *FactionXPBonusSystem
	WeatherManaRegenSystem                      *WeatherManaRegenSystem
	WeatherCooldownSystem                       *WeatherCooldownSystem
	TerrainMovementSpeedSystem                  *TerrainMovementSpeedSystem
	TerrainCombatBonusSystem                    *TerrainCombatBonusSystem
	TerrainStealthSystem                        *TerrainStealthSystem
	TerrainStatusEffectSystem                   *TerrainStatusEffectSystem
	TerrainManaRegenSystem                      *TerrainManaRegenSystem
	TerrainSpellDamageSystem                    *TerrainSpellDamageSystem
	LowHealthVFXSystem                          *LowHealthVFXSystem
	ManaRegenParticleSystem                     *ManaRegenParticleSystem
	CompanionAuraParticleSystem                 *CompanionAuraParticleSystem
	SpecializationManaBoostSystem               *SpecializationManaBoostSystem
	SpecializationHealthRegenSystem             *SpecializationHealthRegenSystem
	SpecializationSpellDamageSystem             *SpecializationSpellDamageSystem
	SpecializationAttackSpeedSystem             *SpecializationAttackSpeedSystem
	SpecializationDefenseSystem                 *SpecializationDefenseSystem
	DualClassSynergySystem                      *DualClassSynergySystem
	ElementalComboParticleSystem                *ElementalComboParticleSystem
	WeatherRangedAccuracySystem                 *WeatherRangedAccuracySystem
	StatusEffectEvasionSystem                   *StatusEffectEvasionSystem
	StatusEffectCriticalChanceSystem            *StatusEffectCriticalChanceSystem
	WeatherXPBonusSystem                        *WeatherXPBonusSystem
	ElementalComboDamageSystem                  *ElementalComboDamageSystem
	ElementalCompanionSynergySystem             *ElementalCompanionSynergySystem
	CompanionSpellAmplificationSystem           *CompanionSpellAmplificationSystem
	LifestealSystem                             *LifestealSystem
	StatusEffectManaCostSystem                  *StatusEffectManaCostSystem
	StatusEffectDamageParticleSystem            *StatusEffectDamageParticleSystem
	WeaponSwingParticleSystem                   *WeaponSwingParticleSystem
	FearFleeParticleSystem                      *FearFleeParticleSystem
	StatusEffectDamageBoostSystem               *StatusEffectDamageBoostSystem
	TerrainCompanionBonusSystem                 *TerrainCompanionBonusSystem
	FactionCompanionBehaviorSystem              *FactionCompanionBehaviorSystem
	WeatherCompanionBonusSystem                 *WeatherCompanionBonusSystem
	EvasionParticleSystem                       *EvasionParticleSystem
	FootstepParticleSystem                      *FootstepParticleSystem
	SpecializationStatusResistSystem            *SpecializationStatusResistSystem
	StatusEffectHealthRegenSystem               *StatusEffectHealthRegenSystem
	HealingParticleSystem                       *HealingParticleSystem
	SpecializationCritDamageSystem              *SpecializationCritDamageSystem
	SpecializationEvasionSystem                 *SpecializationEvasionSystem
	ShieldRegenSystem                           *ShieldRegenSystem
	WeatherMeleeDamageSystem                    *WeatherMeleeDamageSystem
	WeatherAttackSpeedSystem                    *WeatherAttackSpeedSystem
	DrowningParticleSystem                      *DrowningParticleSystem
	HazardProximityWarningParticleSystem        *HazardProximityWarningParticleSystem
	DestructionParticleSystem                   *DestructionParticleSystem
	BlockParticleSystem                         *BlockParticleSystem
	TimeOfDayLightingSystem                     *TimeOfDayLightingSystem
	TimeOfDayStealthSystem                      *TimeOfDayStealthSystem
	TimeOfDayXPBonusSystem                      *TimeOfDayXPBonusSystem
	TimeOfDayManaCostSystem                     *TimeOfDayManaCostSystem
	TimeOfDayCriticalChanceSystem               *TimeOfDayCriticalChanceSystem
	TerrainAmbushCritSystem                     *TerrainAmbushCritSystem
	FactionDamageBonusSystem                    *FactionDamageBonusSystem
	GuildCombatBonusSystem                      *GuildCombatBonusSystem
	WeatherFactionResistanceSystem              *WeatherFactionResistanceSystem
	WeatherCritChanceSystem                     *WeatherCritChanceSystem
	WeatherBlockChanceSystem                    *WeatherBlockChanceSystem
	StealthIndicatorParticleSystem              *StealthIndicatorParticleSystem
	FishingWeatherBonusSystem                   *FishingWeatherBonusSystem
	TimeOfDayFishingBonusSystem                 *TimeOfDayFishingBonusSystem
	FishingCatchParticleSystem                  *FishingCatchParticleSystem
	TimeOfDayFishingBonusParticleSystem         *TimeOfDayFishingBonusParticleSystem
	FishingLineTensionParticleSystem            *FishingLineTensionParticleSystem
	TimeOfDayCompanionBonusSystem               *TimeOfDayCompanionBonusSystem
	TimeOfDayHealthRegenSystem                  *TimeOfDayHealthRegenSystem
	TimeOfDayManaRegenSystem                    *TimeOfDayManaRegenSystem
	TimeOfDayBlockChanceSystem                  *TimeOfDayBlockChanceSystem
	TimeOfDayEvasionSystem                      *TimeOfDayEvasionSystem
	TimeOfDaySpellDamageSystem                  *TimeOfDaySpellDamageSystem
	TimeOfDayAttackSpeedSystem                  *TimeOfDayAttackSpeedSystem
	TerrainCombatBonusParticleSystem            *TerrainCombatBonusParticleSystem
	CompanionManaRegenSystem                    *CompanionManaRegenSystem
	WeatherElementalComboBonusSystem            *WeatherElementalComboBonusSystem
	TerrainFishingBonusSystem                   *TerrainFishingBonusSystem
	WeatherSpellDamageSystem                    *WeatherSpellDamageSystem
	SpecializationLifestealSystem               *SpecializationLifestealSystem
	CompanionFishingBonusSystem                 *CompanionFishingBonusSystem
	TerrainEquipmentDurabilitySystem            *TerrainEquipmentDurabilitySystem
	TerrainEquipmentDurabilityParticleSystem    *TerrainEquipmentDurabilityParticleSystem
	TerrainRangedAccuracySystem                 *TerrainRangedAccuracySystem
	WeatherEquipmentDurabilitySystem            *WeatherEquipmentDurabilitySystem
	SpellChannelParticleSystem                  *SpellChannelParticleSystem
	CompanionDamageLifestealSystem              *CompanionDamageLifestealSystem
	TerrainCompanionHealthRegenSystem           *TerrainCompanionHealthRegenSystem
	WeatherMovementSpeedSystem                  *WeatherMovementSpeedSystem
	WeatherMovementSpeedParticleSystem          *WeatherMovementSpeedParticleSystem
	WeatherCompanionBonusParticleSystem         *WeatherCompanionBonusParticleSystem
	CombatEquipmentDurabilityParticleSystem     *CombatEquipmentDurabilityParticleSystem
	ReputationDefenseBonusSystem                *ReputationDefenseBonusSystem
	ReputationDefenseBonusParticleSystem        *ReputationDefenseBonusParticleSystem
	ReputationHealingBonusSystem                *ReputationHealingBonusSystem
	ReputationHealingBonusParticleSystem        *ReputationHealingBonusParticleSystem
	ReputationSpellDamageBonusSystem            *ReputationSpellDamageBonusSystem
	ReputationSpellDamageBonusParticleSystem    *ReputationSpellDamageBonusParticleSystem
	ReputationMovementSpeedSystem               *ReputationMovementSpeedSystem
	ReputationMovementSpeedParticleSystem       *ReputationMovementSpeedParticleSystem
	ReputationCriticalChanceBonusSystem         *ReputationCriticalChanceBonusSystem
	ReputationCriticalChanceParticleSystem      *ReputationCriticalChanceParticleSystem
	ReputationEquipmentDurabilitySystem         *ReputationEquipmentDurabilitySystem
	ReputationEquipmentDurabilityParticleSystem *ReputationEquipmentDurabilityParticleSystem
	ReputationCompanionBonusSystem              *ReputationCompanionBonusSystem
	ReputationCompanionBonusParticleSystem      *ReputationCompanionBonusParticleSystem
	ReputationQuestGatingSystem                 *ReputationQuestGatingSystem
	FactionTerritoryInfluenceSystem             *FactionTerritoryInfluenceSystem
	AmbientEnvironmentParticleSystem            *AmbientEnvironmentParticleSystem
	EquipmentEnchantmentGlowParticleSystem      *EquipmentEnchantmentGlowParticleSystem
	StatusEffectVisualOverlaySystem             *StatusEffectVisualOverlaySystem
	WeatherSpriteTintSystem                     *WeatherSpriteTintSystem
	EntityDropShadowSystem                      *EntityDropShadowSystem
	TimeOfDayShadowDirectionSystem              *TimeOfDayShadowDirectionSystem
	EquipmentMaterialSheenSystem                *EquipmentMaterialSheenSystem
	EquipmentDamageStateTintSystem              *EquipmentDamageStateTintSystem
	CreatureGenreTintSystem                     *CreatureGenreTintSystem
	CreatureSizeProportionSystem                *CreatureSizeProportionSystem
	CreatureAnatomySystem                       *CreatureAnatomySystem
	EquipmentRarityDetailSystem                 *EquipmentRarityDetailSystem
	NpcFacialDetailSystem                       *NpcFacialDetailSystem
	ProjectileTrailParticleSystem               *ProjectileTrailParticleSystem
	EntityIdleAmbientParticleSystem             *EntityIdleAmbientParticleSystem
	WeaponMaterialImpactParticleSystem          *WeaponMaterialImpactParticleSystem
	DamageFlashTintSystem                       *DamageFlashTintSystem
	SprintTrailParticleSystem                   *SprintTrailParticleSystem
	NearbyLightEntityTintSystem                 *NearbyLightEntityTintSystem
	WeatherEquipmentSheenSystem                 *WeatherEquipmentSheenSystem
	CreatureEyeGlowSystem                       *CreatureEyeGlowSystem
	CreatureEyePatternSystem                    *CreatureEyePatternSystem
	CreatureElementalAuraSystem                 *CreatureElementalAuraSystem
	MeleeSwingArcSystem                         *MeleeSwingArcSystem
	CombatReadyAuraSystem                       *CombatReadyAuraSystem
	AIStateBubbleSystem                         *AIStateBubbleSystem
	TerrainReflectionTintSystem                 *TerrainReflectionTintSystem
	MovementBobSystem                           *MovementBobSystem
	MovementLeanSystem                          *MovementLeanSystem
	SpellCastGlowSystem                         *SpellCastGlowSystem
	EquipmentChangeFlashSystem                  *EquipmentChangeFlashSystem
	DodgeAfterimageSystem                       *DodgeAfterimageSystem
	EntityIdleBreathingSystem                   *EntityIdleBreathingSystem
	CombatHitStaggerSystem                      *CombatHitStaggerSystem
	FloatingDamageNumberSystem                  *FloatingDamageNumberSystem
	EntityThreatIndicatorSystem                 *EntityThreatIndicatorSystem
	WeaponMaterialParticleSystem                *WeaponMaterialParticleSystem
	LootRarityBeamSystem                        *LootRarityBeamSystem
	DamageTypeColorFlashSystem                  *DamageTypeColorFlashSystem
	CompanionBondTetherSystem                   *CompanionBondTetherSystem
	CriticalHitScreenShakeSystem                *CriticalHitScreenShakeSystem
	EntitySpawnMaterializeSystem                *EntitySpawnMaterializeSystem
	NPCInteractionProximityGlowSystem           *NPCInteractionProximityGlowSystem
	EntityFactionOutlineSystem                  *EntityFactionOutlineSystem
	ShieldBubbleOverlaySystem                   *ShieldBubbleOverlaySystem
	EnvironmentalBreathVaporSystem              *EnvironmentalBreathVaporSystem
	MovementDustSystem                          *MovementDustSystem
	DynamicExpressionSystem                     *DynamicExpressionSystem
	HealthRegenPulseSystem                      *HealthRegenPulseSystem
	StatusEffectGroundTrailSystem               *StatusEffectGroundTrailSystem
	WaterSurfaceRippleSystem                    *WaterSurfaceRippleSystem
	EntityTargetLockIndicatorSystem             *EntityTargetLockIndicatorSystem
	EntityDeathDissolveSystem                   *EntityDeathDissolveSystem
	ArmorHitSparkSystem                         *ArmorHitSparkSystem
	MeleeEnchantmentArcParticleSystem           *MeleeEnchantmentArcParticleSystem
	BattleWoundOverlaySystem                    *BattleWoundOverlaySystem
	EquipmentDamageCrackOverlaySystem           *EquipmentDamageCrackOverlaySystem
	WeatherEntityWetnessSystem                  *WeatherEntityWetnessSystem
	AttackTelegraphGlowSystem                   *AttackTelegraphGlowSystem
	XPGainBurstSystem                           *XPGainBurstSystem
	EquipmentGleamSweepSystem                   *EquipmentGleamSweepSystem
	SpriteDepthShadingSystem                    *SpriteDepthShadingSystem
	ClothingPatternSystem                       *ClothingPatternSystem
	SurfaceTextureSystem                        *SurfaceTextureSystem
	HumanoidTextureSystem                       *HumanoidTextureSystem
	BodyTypeSystem                              *BodyTypeSystem
	HeadgearAssignmentSystem                    *HeadgearAssignmentSystem
	BackAccessorySystem                         *BackAccessorySystem
	CreatureVisualClassifierSystem              *CreatureVisualClassifierSystem
	SizeSpriteScalingSystem                     *SizeSpriteScalingSystem
	NpcRoleVisualSystem                         *NpcRoleVisualSystem
	DirectionalSpriteSystem                     *DirectionalSpriteSystem
	SpriteDepthEnhanceSystem                    *SpriteDepthEnhanceSystem
	SpriteColorTemperatureSystem                *SpriteColorTemperatureSystem
	SpriteFinalizerSystem                       *SpriteFinalizerSystem
	CompanionQuestSynergySystem                 *CompanionQuestSynergySystem
	AchievementNotificationSystem               *AchievementNotificationSystem
	AvailabilitySystem                          *AvailabilitySystem
	SkillPointGainParticleSystem                *SkillPointGainParticleSystem
	// Seasonal-event cluster (G2/G4 — registered; SeasonalEventComponent seeded at runtime)
	EventCalendarSystem    *EventCalendarSystem
	EventQuestSystem       *EventQuestSystem
	EventDecorationSystem  *EventDecorationSystem
	EventRewardSystem      *EventRewardSystem
	// Mod systems (G2/G3 — basic registration; full repository wiring done in cmd/client)
	ModCompatibilitySystem *ModCompatibilitySystem
	ModBrowserSystem       *ModBrowserSystem
	// G10: ExtendedAchievementSystem is the gameplay tracker (kills/quests/crafting/PvP).
	// Safe to register alongside AchievementSystem — they use different components.
	ExtendedAchievementSystem *ExtendedAchievementSystem

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
// COMPLEXITY JUSTIFICATION: High cyclomatic complexity (19) comes from validating
// multiple required config fields at function start. The function body is actually
// linear—sequentially creating and registering 66 systems with no nested conditionals.
// The 1884-line length is justified by keeping all system initialization in one location
// for discoverability and consistent initialization order.
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
	// Register per-entity cleanup hook so that visitedCells entries are reclaimed
	// when an entity is removed, preventing unbounded per-session memory growth.
	movementSystem.RegisterRemovalHook(game.World)

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

	// 17ai. AchievementNotificationSystem - distributes rewards and fires UI/audio
	// callbacks when achievements are unlocked. Wired here alongside ProgressionSystem
	// because achievement rewards are denominated in XP and gold that ProgressionSystem
	// also manages. Callers set OnRewardGranted/OnPlayUnlockSound callbacks after init.
	achievementNotificationSystem := NewAchievementNotificationSystem(game.World)
	achievementNotificationSystem.SetRewardSeed(config.Seed ^ 0x4143484E) // "ACHN"
	result.AchievementNotificationSystem = achievementNotificationSystem
	game.World.AddSystem(achievementNotificationSystem)

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

	// 17b1. GuildCombatBonusSystem - proximity-based combat bonuses for guild members
	// Connects GuildComponent membership with combat stats when guild members fight together
	result.GuildCombatBonusSystem = regGenre(game.World,
		NewGuildCombatBonusSystem(game.World, config.Seed+5152), config.GenreID)

	// 17b2. WeatherFactionResistanceSystem - weather affects faction combat stats
	// Connects WeatherComponent with FactionComponent for elemental affinity bonuses/penalties
	result.WeatherFactionResistanceSystem = regGenre(game.World,
		NewWeatherFactionResistanceSystem(game.World, config.Seed+5155), config.GenreID)

	// 17c. ReputationDefenseBonusSystem - defense bonus against enemies of allied factions
	// Connects FactionComponent reputation with StatsComponent.Defense for damage reduction
	reputationDefenseBonusSystem := NewReputationDefenseBonusSystem(game.World, config.Seed+5175)
	reputationDefenseBonusSystem.SetFactionSystem(factionSystem)
	reputationDefenseBonusSystem.SetGenre(config.GenreID)
	result.ReputationDefenseBonusSystem = reputationDefenseBonusSystem
	game.World.AddSystem(reputationDefenseBonusSystem)

	// 17d. ReputationDefenseBonusParticleSystem - visual feedback for reputation defense
	// Connects ReputationDefenseBonusSystem damage reduction with particle effects
	reputationDefenseBonusParticleSystem := NewReputationDefenseBonusParticleSystem(game.World, config.Seed+5180)
	reputationDefenseBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationDefenseBonusParticleSystem.SetGenre(config.GenreID)
	result.ReputationDefenseBonusParticleSystem = reputationDefenseBonusParticleSystem
	game.World.AddSystem(reputationDefenseBonusParticleSystem)

	// 17e. ReputationHealingBonusSystem - health regen bonus from allied faction reputation
	reputationHealingBonusSystem := NewReputationHealingBonusSystem(game.World, config.Seed+5185)
	reputationHealingBonusSystem.SetFactionSystem(factionSystem)
	reputationHealingBonusSystem.SetGenre(config.GenreID)
	result.ReputationHealingBonusSystem = reputationHealingBonusSystem
	game.World.AddSystem(reputationHealingBonusSystem)

	// 17f. ReputationHealingBonusParticleSystem - visual feedback for reputation healing
	reputationHealingBonusParticleSystem := NewReputationHealingBonusParticleSystem(game.World, config.Seed+5190)
	reputationHealingBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationHealingBonusParticleSystem.SetHealingSystem(reputationHealingBonusSystem)
	reputationHealingBonusParticleSystem.SetGenre(config.GenreID)
	result.ReputationHealingBonusParticleSystem = reputationHealingBonusParticleSystem
	game.World.AddSystem(reputationHealingBonusParticleSystem)

	// 17g. ReputationSpellDamageBonusSystem - spell damage bonus from allied faction reputation
	reputationSpellDamageBonusSystem := NewReputationSpellDamageBonusSystem(game.World, config.Seed+5195)
	reputationSpellDamageBonusSystem.SetFactionSystem(factionSystem)
	reputationSpellDamageBonusSystem.SetGenre(config.GenreID)
	result.ReputationSpellDamageBonusSystem = reputationSpellDamageBonusSystem
	game.World.AddSystem(reputationSpellDamageBonusSystem)

	// 17h. ReputationSpellDamageBonusParticleSystem - visual feedback for reputation spell damage
	reputationSpellDamageBonusParticleSystem := NewReputationSpellDamageBonusParticleSystem(game.World, config.Seed+5200)
	reputationSpellDamageBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationSpellDamageBonusParticleSystem.SetSpellDamageSystem(reputationSpellDamageBonusSystem)
	reputationSpellDamageBonusParticleSystem.SetGenre(config.GenreID)
	result.ReputationSpellDamageBonusParticleSystem = reputationSpellDamageBonusParticleSystem
	game.World.AddSystem(reputationSpellDamageBonusParticleSystem)

	// 17i. ReputationMovementSpeedSystem - movement speed bonus from allied faction reputation
	reputationMovementSpeedSystem := NewReputationMovementSpeedSystem(game.World, config.Seed+5205)
	reputationMovementSpeedSystem.SetFactionSystem(factionSystem)
	reputationMovementSpeedSystem.SetGenre(config.GenreID)
	result.ReputationMovementSpeedSystem = reputationMovementSpeedSystem
	game.World.AddSystem(reputationMovementSpeedSystem)

	// 17j. ReputationMovementSpeedParticleSystem - visual feedback for reputation speed bonus
	reputationMovementSpeedParticleSystem := NewReputationMovementSpeedParticleSystem(game.World, config.Seed+5210)
	reputationMovementSpeedParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationMovementSpeedParticleSystem.SetSpeedSystem(reputationMovementSpeedSystem)
	reputationMovementSpeedParticleSystem.SetGenre(config.GenreID)
	result.ReputationMovementSpeedParticleSystem = reputationMovementSpeedParticleSystem
	game.World.AddSystem(reputationMovementSpeedParticleSystem)

	// 17k. ReputationCriticalChanceBonusSystem - crit chance bonus from allied faction reputation
	reputationCriticalChanceBonusSystem := NewReputationCriticalChanceBonusSystem(game.World, config.Seed+5215)
	reputationCriticalChanceBonusSystem.SetFactionSystem(factionSystem)
	reputationCriticalChanceBonusSystem.SetGenre(config.GenreID)
	result.ReputationCriticalChanceBonusSystem = reputationCriticalChanceBonusSystem
	game.World.AddSystem(reputationCriticalChanceBonusSystem)

	// 17l. ReputationCriticalChanceParticleSystem - visual feedback for reputation crit bonus
	reputationCriticalChanceParticleSystem := NewReputationCriticalChanceParticleSystem(game.World, config.Seed+5220)
	reputationCriticalChanceParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationCriticalChanceParticleSystem.SetCritSystem(reputationCriticalChanceBonusSystem)
	reputationCriticalChanceParticleSystem.SetGenre(config.GenreID)
	result.ReputationCriticalChanceParticleSystem = reputationCriticalChanceParticleSystem
	game.World.AddSystem(reputationCriticalChanceParticleSystem)

	// 17m. ReputationEquipmentDurabilitySystem - equipment preservation from allied faction reputation
	reputationEquipmentDurabilitySystem := NewReputationEquipmentDurabilitySystem(game.World, config.Seed+5225)
	reputationEquipmentDurabilitySystem.SetFactionSystem(factionSystem)
	reputationEquipmentDurabilitySystem.SetGenre(config.GenreID)
	result.ReputationEquipmentDurabilitySystem = reputationEquipmentDurabilitySystem
	game.World.AddSystem(reputationEquipmentDurabilitySystem)

	// 17n. ReputationEquipmentDurabilityParticleSystem - visual feedback for reputation durability
	reputationEquipmentDurabilityParticleSystem := NewReputationEquipmentDurabilityParticleSystem(game.World, config.Seed+5230)
	reputationEquipmentDurabilityParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationEquipmentDurabilityParticleSystem.SetDurabilitySystem(reputationEquipmentDurabilitySystem)
	reputationEquipmentDurabilityParticleSystem.SetGenre(config.GenreID)
	result.ReputationEquipmentDurabilityParticleSystem = reputationEquipmentDurabilityParticleSystem
	game.World.AddSystem(reputationEquipmentDurabilityParticleSystem)

	// 17o. ReputationCompanionBonusSystem - companion stat boosts from allied faction reputation
	reputationCompanionBonusSystem := NewReputationCompanionBonusSystem(game.World, config.Seed+5235)
	reputationCompanionBonusSystem.SetFactionSystem(factionSystem)
	reputationCompanionBonusSystem.SetGenre(config.GenreID)
	result.ReputationCompanionBonusSystem = reputationCompanionBonusSystem
	game.World.AddSystem(reputationCompanionBonusSystem)

	// 17p. ReputationCompanionBonusParticleSystem - visual feedback for reputation companion bonus
	reputationCompanionBonusParticleSystem := NewReputationCompanionBonusParticleSystem(game.World, config.Seed+5240)
	reputationCompanionBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	reputationCompanionBonusParticleSystem.SetBonusSystem(reputationCompanionBonusSystem)
	reputationCompanionBonusParticleSystem.SetGenre(config.GenreID)
	result.ReputationCompanionBonusParticleSystem = reputationCompanionBonusParticleSystem
	game.World.AddSystem(reputationCompanionBonusParticleSystem)

	// 17q-b. ReputationQuestGatingSystem - gates quests behind faction reputation requirements
	// Integrates reputation standings with quest availability for meaningful progression
	reputationQuestGatingSystem := NewReputationQuestGatingSystem(game.World, logger, config.Seed+5245)
	result.ReputationQuestGatingSystem = reputationQuestGatingSystem
	game.World.AddSystem(reputationQuestGatingSystem)

	// 17q-c. FactionTerritoryInfluenceSystem - faction zones provide combat/progression bonuses
	// Integrates faction reputation with territory mechanics for zone-based buffs/debuffs
	factionTerritoryInfluenceSystem := NewFactionTerritoryInfluenceSystem(game.World, factionSystem)
	result.FactionTerritoryInfluenceSystem = factionTerritoryInfluenceSystem
	game.World.AddSystem(factionTerritoryInfluenceSystem)

	// 17q. AmbientEnvironmentParticleSystem - atmospheric particles based on terrain type
	// Connects terrain data with ParticleSystem for genre-aware ambient effects
	ambientEnvironmentParticleSystem := NewAmbientEnvironmentParticleSystem(game.World, config.Seed+7350)
	ambientEnvironmentParticleSystem.SetParticleSystem(result.ParticleSystem)
	ambientEnvironmentParticleSystem.SetGenre(config.GenreID)
	ambientEnvironmentParticleSystem.SetTileSize(config.TileSize)
	result.AmbientEnvironmentParticleSystem = ambientEnvironmentParticleSystem
	game.World.AddSystem(ambientEnvironmentParticleSystem)

	// 17r. EquipmentEnchantmentGlowParticleSystem - ambient glow for rare+ equipment
	// Connects EquipmentComponent rarity with ParticleSystem for rarity-colored enchantment aura
	equipEnchantGlowSystem := NewEquipmentEnchantmentGlowParticleSystem(game.World, config.Seed+7420)
	equipEnchantGlowSystem.SetParticleSystem(result.ParticleSystem)
	equipEnchantGlowSystem.SetGenre(config.GenreID)
	result.EquipmentEnchantmentGlowParticleSystem = equipEnchantGlowSystem
	game.World.AddSystem(equipEnchantGlowSystem)

	// 17s. StatusEffectVisualOverlaySystem - color tints on sprites from status effects
	// Connects StatusEffectComponent data with VisualFeedbackComponent tint fields
	result.StatusEffectVisualOverlaySystem = regGenre(game.World,
		NewStatusEffectVisualOverlaySystem(game.World), config.GenreID)

	// 17t. WeatherSpriteTintSystem - weather-driven sprite color tints
	// Reads active weather and applies subtle multiplicative tints to entity sprites
	result.WeatherSpriteTintSystem = regGenre(game.World,
		NewWeatherSpriteTintSystem(game.World, config.Seed+8800), config.GenreID)

	// 17u. EntityDropShadowSystem - soft elliptical drop shadows beneath entities
	// Reads entity collider/sprite size and genre to compute genre-aware shadow parameters
	result.EntityDropShadowSystem = regGenre(game.World,
		NewEntityDropShadowSystem(game.World, config.Seed+8900), config.GenreID)

	// 17u2. TimeOfDayShadowDirectionSystem - directional shadow offset from sun position
	// Connects TimeOfDayLightingSystem with DropShadowComponent for sun-arc shadow direction
	result.TimeOfDayShadowDirectionSystem = regGenre(game.World,
		NewTimeOfDayShadowDirectionSystem(game.World, config.Seed+8925), config.GenreID)

	// 17v. EquipmentMaterialSheenSystem - specular highlights from material properties
	// Bridges sprites.GetMaterialVisualProperties (Sheen/Reflectivity/Roughness) with per-entity visual state
	result.EquipmentMaterialSheenSystem = regGenre(game.World,
		NewEquipmentMaterialSheenSystem(game.World, config.Seed+8950), config.GenreID)

	// 17w. EquipmentDamageStateTintSystem - aggregate equipment wear visual tinting
	// Bridges sprites.GetDamageVisualEffects with per-entity render state for darkening/opacity
	result.EquipmentDamageStateTintSystem = regGenre(game.World,
		NewEquipmentDamageStateTintSystem(game.World, config.Seed+8975), config.GenreID)

	// 17x. CreatureGenreTintSystem - genre-aware creature/NPC sprite color tinting
	// Applies genre-specific tint presets to creature sprites for atmospheric cohesion
	result.CreatureGenreTintSystem = regGenre(game.World,
		NewCreatureGenreTintSystem(game.World, config.Seed+9000), config.GenreID)

	// 17y. CreatureSizeProportionSystem - size-based anatomy proportions for creatures
	result.CreatureSizeProportionSystem = regGenre(game.World,
		NewCreatureSizeProportionSystem(game.World, config.Seed+9001), config.GenreID)

	// 17y1. CreatureAnatomySystem - assigns anatomy types (quadruped, arachnid, etc.) to creatures
	result.CreatureAnatomySystem = regGenre(game.World,
		NewCreatureAnatomySystem(game.World, config.Seed+9010), config.GenreID)

	// 17z. EquipmentRarityDetailSystem - rarity-based visual detail scaling
	result.EquipmentRarityDetailSystem = regGenre(game.World,
		NewEquipmentRarityDetailSystem(game.World, config.Seed+9025), config.GenreID)

	// 17aa. NpcFacialDetailSystem - genre-aware NPC facial features
	result.NpcFacialDetailSystem = regGenre(game.World,
		NewNpcFacialDetailSystem(game.World, config.Seed+9050), config.GenreID)

	// 17ab. ProjectileTrailParticleSystem - genre-aware projectile trail particles
	// Connects ProjectileComponent with ParticleSystem for visual trails behind projectiles
	projectileTrailParticleSystem := NewProjectileTrailParticleSystem(game.World, config.Seed+9075)
	projectileTrailParticleSystem.SetParticleSystem(result.ParticleSystem)
	projectileTrailParticleSystem.SetGenre(config.GenreID)
	result.ProjectileTrailParticleSystem = projectileTrailParticleSystem
	game.World.AddSystem(projectileTrailParticleSystem)

	// 17ac. EntityIdleAmbientParticleSystem - genre-aware idle entity ambient particles
	// Connects VelocityComponent with ParticleSystem for subtle particles around stationary entities
	entityIdleAmbientParticleSystem := NewEntityIdleAmbientParticleSystem(game.World, config.Seed+9100)
	entityIdleAmbientParticleSystem.SetParticleSystem(result.ParticleSystem)
	entityIdleAmbientParticleSystem.SetGenre(config.GenreID)
	result.EntityIdleAmbientParticleSystem = entityIdleAmbientParticleSystem
	game.World.AddSystem(entityIdleAmbientParticleSystem)

	// 17ad. WeaponMaterialImpactParticleSystem - material-aware melee impact particles
	// Connects EquipmentComponent weapon material with ParticleSystem for distinct impact visuals
	weaponMaterialImpactParticleSystem := NewWeaponMaterialImpactParticleSystem(game.World, config.Seed+9150)
	weaponMaterialImpactParticleSystem.SetParticleSystem(result.ParticleSystem)
	weaponMaterialImpactParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.AddDamageCallback(weaponMaterialImpactParticleSystem.OnMeleeImpact)
	result.WeaponMaterialImpactParticleSystem = weaponMaterialImpactParticleSystem
	game.World.AddSystem(weaponMaterialImpactParticleSystem)

	// 17ae. DamageFlashTintSystem - genre-aware damage flash tints on health decrease
	// Monitors HealthComponent changes and triggers VisualFeedbackComponent flashes
	result.DamageFlashTintSystem = regGenre(game.World, NewDamageFlashTintSystem(game.World), config.GenreID)

	// 17af. SprintTrailParticleSystem - genre-aware speed trail particles for sprinting entities
	// Monitors VelocityComponent and spawns trail particles behind fast-moving entities
	sprintTrailParticleSystem := NewSprintTrailParticleSystem(game.World, config.Seed+9200)
	sprintTrailParticleSystem.SetParticleSystem(result.ParticleSystem)
	sprintTrailParticleSystem.SetGenre(config.GenreID)
	result.SprintTrailParticleSystem = sprintTrailParticleSystem
	game.World.AddSystem(sprintTrailParticleSystem)

	// 17ag. NearbyLightEntityTintSystem - tints entity sprites based on nearby light sources
	// Uses light color, intensity, and distance-based falloff with genre-aware ambient base
	result.NearbyLightEntityTintSystem = regGenre(game.World,
		NewNearbyLightEntityTintSystem(game.World, config.Seed+9300), config.GenreID)

	// 17ah. WeatherEquipmentSheenSystem - weather-driven equipment sheen modifications
	// Bridges WeatherComponent with MaterialSheenComponent for wet/frost/dust effects
	result.WeatherEquipmentSheenSystem = regGenre(game.World,
		NewWeatherEquipmentSheenSystem(game.World, config.Seed+9400), config.GenreID)

	// 17ai. CreatureEyeGlowSystem - genre-aware glowing eyes for hostile creatures
	// Uses threat level (health, faction, detection range) to scale glow intensity and pulse
	result.CreatureEyeGlowSystem = regGenre(game.World,
		NewCreatureEyeGlowSystem(game.World, config.Seed+9450), config.GenreID)

	// 17ai1. CreatureEyePatternSystem - creature-type-specific eye patterns for nonhumanoids
	// Assigns 8-eye spider patterns, slit-pupil serpent eyes, compound insect eyes, etc.
	result.CreatureEyePatternSystem = regGenre(game.World,
		NewCreatureEyePatternSystem(game.World, config.Seed+9460), config.GenreID)

	// 17ai2. CreatureElementalAuraSystem - persistent elemental aura visuals for creatures
	// Infers elemental affinity from name/tags/attack and assigns colored aura overlays
	result.CreatureElementalAuraSystem = regGenre(game.World,
		NewCreatureElementalAuraSystem(game.World, config.Seed+9475), config.GenreID)

	// 17aj. MeleeSwingArcSystem - genre-aware visual arc overlays during melee attacks
	// Watches AnimationComponent attack transitions and writes MeleeSwingArcComponent
	result.MeleeSwingArcSystem = regGenre(game.World, NewMeleeSwingArcSystem(game.World, config.Seed+9500), config.GenreID)

	// 17ak. CombatReadyAuraSystem - genre-aware aura when AI entities enter hostile states
	// Connects AIComponent state with visual aura overlay for combat readiness feedback
	result.CombatReadyAuraSystem = regGenre(game.World,
		NewCombatReadyAuraSystem(game.World, config.Seed+9550), config.GenreID)

	// 17al. AIStateBubbleSystem - genre-aware floating state indicator bubbles above NPCs
	// Connects AIComponent state with floating symbol overlay for AI behavior feedback
	result.AIStateBubbleSystem = regGenre(game.World, NewAIStateBubbleSystem(game.World, config.Seed+9600), config.GenreID)

	// 17am. TerrainReflectionTintSystem - terrain-driven sprite color tinting
	// Reads entity position, looks up terrain tile type, and writes genre-aware tint
	// values so entities on water appear blue-tinted, lava orange, trees green, etc.
	terrainReflectionTintSystem := NewTerrainReflectionTintSystem(game.World, config.Seed+9650)
	terrainReflectionTintSystem.SetGenre(config.GenreID)
	if config.TileSize > 0 {
		terrainReflectionTintSystem.SetTileSize(config.TileSize)
	}
	result.TerrainReflectionTintSystem = terrainReflectionTintSystem
	game.World.AddSystem(terrainReflectionTintSystem)

	// 17an. MovementBobSystem - genre-aware walk-cycle vertical sprite bobbing
	// Reads VelocityComponent and writes sinusoidal Y offset to MovementBobComponent
	// so the render system displaces moving entities for visual movement weight.
	result.MovementBobSystem = regGenre(game.World, NewMovementBobSystem(game.World, config.Seed+9700), config.GenreID)

	// 17an2. MovementLeanSystem - genre-aware horizontal lean in movement direction
	// Reads VelocityComponent and writes smoothed X offset to MovementLeanComponent
	// so the render system displaces entities horizontally for momentum lean.
	result.MovementLeanSystem = regGenre(game.World, NewMovementLeanSystem(game.World, config.Seed+9710), config.GenreID)

	// 17ao. SpellCastGlowSystem - genre-aware visual glow during spell casting
	// Reads SpellSlotComponent casting state and writes SpellCastGlowComponent
	// with element-derived color, intensity that ramps with CastingBar, and pulse.
	result.SpellCastGlowSystem = regGenre(game.World, NewSpellCastGlowSystem(game.World, config.Seed+9750), config.GenreID)

	// 17ap. EquipmentChangeFlashSystem - genre-aware flash particles on equipment change
	// Reads EquipmentComponent slot state, detects equip/unequip transitions, and
	// spawns sparkle particles via ParticleSystem for immediate visual feedback.
	equipmentChangeFlashSystem := NewEquipmentChangeFlashSystem(game.World, config.Seed+9800)
	equipmentChangeFlashSystem.SetParticleSystem(result.ParticleSystem)
	equipmentChangeFlashSystem.SetGenre(config.GenreID)
	result.EquipmentChangeFlashSystem = equipmentChangeFlashSystem
	game.World.AddSystem(equipmentChangeFlashSystem)

	// 17aq. DodgeAfterimageSystem - genre-aware translucent ghost copies on fast movement
	// Reads VelocityComponent speed, spawns fading ghost positions in AfterimageComponent
	// for a motion trail effect distinct from particle-based sprint trails.
	result.DodgeAfterimageSystem = regGenre(game.World,
		NewDodgeAfterimageSystem(game.World, config.Seed+9850), config.GenreID)

	// 17ar. EntityIdleBreathingSystem - genre-aware subtle idle breathing animation
	// Reads VelocityComponent to detect stationary entities, writes sinusoidal Y
	// offset to IdleBreathingComponent for a lifelike idle visual effect.
	result.EntityIdleBreathingSystem = regGenre(game.World,
		NewEntityIdleBreathingSystem(game.World, config.Seed+9900), config.GenreID)

	// 17as. CombatHitStaggerSystem - genre-aware positional stagger on damage
	// Monitors HealthComponent changes and writes decaying X/Y offsets to HitStaggerComponent
	result.CombatHitStaggerSystem = regGenre(game.World,
		NewCombatHitStaggerSystem(game.World, config.Seed+9950), config.GenreID)

	// 17at. FloatingDamageNumberSystem - genre-aware floating damage/heal numbers
	// Monitors HealthComponent changes and writes rising/fading numbers to FloatingDamageNumberComponent
	result.FloatingDamageNumberSystem = regGenre(game.World,
		NewFloatingDamageNumberSystem(game.World, config.Seed+10000), config.GenreID)

	// 17au. EntityThreatIndicatorSystem - genre-aware threat level rings under AI entities
	// Compares entity level to player level and assigns colored ring indicators
	result.EntityThreatIndicatorSystem = regGenre(game.World,
		NewEntityThreatIndicatorSystem(game.World, config.Seed+10050), config.GenreID)

	// 17av. WeaponMaterialParticleSystem - genre-aware idle particles from weapon material
	// Reads main-hand weapon material type and spawns material-appropriate ambient particles
	weaponMaterialParticleSystem := NewWeaponMaterialParticleSystem(game.World, config.Seed+10100)
	weaponMaterialParticleSystem.SetParticleSystem(result.ParticleSystem)
	weaponMaterialParticleSystem.SetGenre(config.GenreID)
	result.WeaponMaterialParticleSystem = weaponMaterialParticleSystem
	game.World.AddSystem(weaponMaterialParticleSystem)

	// 17aw. LootRarityBeamSystem - genre-aware beam particles on ground items by rarity
	// Items with Uncommon+ rarity emit colored upward-flowing beam particles for loot visibility
	lootRarityBeamSystem := NewLootRarityBeamSystem(game.World, config.Seed+10200)
	lootRarityBeamSystem.SetParticleSystem(result.ParticleSystem)
	lootRarityBeamSystem.SetGenre(config.GenreID)
	result.LootRarityBeamSystem = lootRarityBeamSystem
	game.World.AddSystem(lootRarityBeamSystem)

	// 17ax. DamageTypeColorFlashSystem - genre-aware elemental damage color flashes
	// Hooks into CombatSystem damage callbacks and tints sprites by damage type
	damageTypeColorFlashSystem := NewDamageTypeColorFlashSystem(game.World, config.Seed+10250)
	damageTypeColorFlashSystem.SetGenre(config.GenreID)
	result.CombatSystem.AddDamageCallback(damageTypeColorFlashSystem.OnDamageDealt)
	result.DamageTypeColorFlashSystem = damageTypeColorFlashSystem
	game.World.AddSystem(damageTypeColorFlashSystem)

	// 17ay. CompanionBondTetherSystem - genre-aware visual tether between companion and owner
	// Reads CompanionComponent.OwnerID and PositionComponent to write tether geometry/color
	// into CompanionBondTetherComponent for the renderer.
	result.CompanionBondTetherSystem = regGenre(game.World,
		NewCompanionBondTetherSystem(game.World, config.Seed+10300), config.GenreID)

	// 17az. CriticalHitScreenShakeSystem - genre-aware camera shake on critical hits
	// Connects CombatSystem critical hit callback with CameraSystem.ShakeAdvanced
	// for visceral combat feedback with distance attenuation and genre-tuned intensity.
	criticalHitScreenShakeSystem := NewCriticalHitScreenShakeSystem(game.World, config.Seed+10350)
	criticalHitScreenShakeSystem.SetCameraSystem(game.CameraSystem)
	criticalHitScreenShakeSystem.SetGenre(config.GenreID)
	result.CombatSystem.AddCriticalHitCallback(criticalHitScreenShakeSystem.OnCriticalHit)
	result.CriticalHitScreenShakeSystem = criticalHitScreenShakeSystem
	game.World.AddSystem(criticalHitScreenShakeSystem)

	// 17ba. EntitySpawnMaterializeSystem - genre-aware fade-in and particle burst
	// when entities first appear in the world, providing visual spawn feedback.
	result.EntitySpawnMaterializeSystem = regGenre(game.World,
		NewEntitySpawnMaterializeSystem(game.World, config.Seed+10400), config.GenreID)

	// 17bb. NPCInteractionProximityGlowSystem - genre-aware glow tint on interactable
	// NPCs (dialog/merchant) when a player entity is within proximity range.
	result.NPCInteractionProximityGlowSystem = regGenre(game.World,
		NewNPCInteractionProximityGlowSystem(game.World, config.Seed+10450), config.GenreID)

	// 17bc. EntityFactionOutlineSystem - genre-aware colored outlines around entities
	// based on TeamComponent allegiance and FactionComponent reputation.
	// Allies show green, hostiles pulsing red, neutrals amber.
	result.EntityFactionOutlineSystem = regGenre(game.World,
		NewEntityFactionOutlineSystem(game.World, config.Seed+10500), config.GenreID)

	// 17bd. ShieldBubbleOverlaySystem - genre-aware translucent bubble around
	// entities with active ShieldComponent. Opacity scales with shield amount,
	// flickers at low shield, fades out on expiry.
	result.ShieldBubbleOverlaySystem = regGenre(game.World,
		NewShieldBubbleOverlaySystem(game.World, config.Seed+10550), config.GenreID)

	// 17be. EnvironmentalBreathVaporSystem - cold weather breath vapor puffs
	// Spawns genre-aware vapor particles near entity faces during cold weather (snow, fog, ash)
	result.EnvironmentalBreathVaporSystem = regGenre(game.World,
		NewEnvironmentalBreathVaporSystem(game.World, config.Seed+10600), config.GenreID)

	// 17bf. MovementDustSystem - speed-proportional terrain dust behind fast movers
	// Genre-aware particles: fantasy=dust, horror=ash wisps, scifi=energy sparks
	result.MovementDustSystem = regGenre(game.World, NewMovementDustSystem(game.World, config.Seed+10650), config.GenreID)

	// 17bg. DynamicExpressionSystem - reactive facial expression updates
	// Reads HealthComponent, StatusEffectComponent, and AIComponent to update
	// NpcFacialDetailComponent.ExpressionType, then marks sprites dirty for re-render.
	dynamicExpressionSystem := NewDynamicExpressionSystem(game.World, config.Seed+10700)
	result.DynamicExpressionSystem = dynamicExpressionSystem
	game.World.AddSystem(dynamicExpressionSystem)

	// 18. SkillProgressionSystem - applies skill effects to stats
	skillProgressionSystem := NewSkillProgressionSystem()
	game.World.AddSystem(skillProgressionSystem)

	// 18a. SkillLoadoutSystem - manages saved skill configurations and loadout swapping
	// Connects SkillTreeComponent with SkillLoadoutComponent for build management
	skillLoadoutSystem := NewSkillLoadoutSystem(game.World)
	result.SkillLoadoutSystem = skillLoadoutSystem
	game.World.AddSystem(skillLoadoutSystem)

	// 18b. WeaponMasterySystem - tracks weapon proficiency and awards combat bonuses
	// Players gain XP with each weapon type, unlocking damage/speed/crit bonuses
	weaponMasterySystem := NewWeaponMasterySystem(game.World)
	result.WeaponMasterySystem = weaponMasterySystem
	game.World.AddSystem(weaponMasterySystem)

	// Wire up weapon mastery to combat callbacks
	result.CombatSystem.AddDamageCallback(func(attacker, target *Entity, damage float64) {
		weaponType := weaponMasterySystem.GetEquippedWeaponType(attacker)
		if weaponType != "" {
			weaponMasterySystem.OnHit(attacker, weaponType, damage)
		}
	})
	result.CombatSystem.AddCriticalHitCallback(func(attacker, target *Entity, damage float64) {
		weaponType := weaponMasterySystem.GetEquippedWeaponType(attacker)
		if weaponType != "" {
			weaponMasterySystem.OnCriticalHit(attacker, weaponType)
		}
	})

	// 18b2. ClassAffinitySystem - tracks playstyle progression and combat archetypes
	// Players build affinity XP by using abilities, unlocking passive bonuses for their preferred playstyle
	classAffinitySystem := NewClassAffinitySystem(game.World)
	result.ClassAffinitySystem = classAffinitySystem
	game.World.AddSystem(classAffinitySystem)

	// Wire class affinity to combat callbacks for damage tracking
	result.CombatSystem.AddDamageCallback(func(attacker, target *Entity, damage float64) {
		// Track damage dealt for affinity XP (ability used is determined by last action)
		if classAffinitySystem != nil {
			comp, hasComp := attacker.GetComponent("class_affinity")
			if hasComp {
				affinity := comp.(*ClassAffinityComponent)
				if affinity.TotalAffinityXP > 0 {
					// Award bonus XP for high damage hits
					if damage >= 100 {
						classAffinitySystem.OnAbilityUsed(attacker, "power_strike", damage, 0)
					}
				}
			}
		}
	})
	result.CombatSystem.SetKillCallback(func(attacker, target *Entity) {
		weaponType := weaponMasterySystem.GetEquippedWeaponType(attacker)
		if weaponType != "" {
			weaponMasterySystem.OnKill(attacker, weaponType)
		}
		// Also award affinity XP for kills
		if classAffinitySystem != nil {
			classAffinitySystem.OnKill(attacker, "power_strike")
		}
		// Award base combat XP to the attacker based on target's level/stats
		if progressionSystem != nil {
			xp := progressionSystem.CalculateXPReward(target)
			if xp > 0 {
				if err := progressionSystem.AwardXP(attacker, xp); err != nil {
					logrus.WithFields(logrus.Fields{
						"system": "progression",
						"xp":     xp,
					}).WithError(err).Warn("Failed to award kill XP")
				}
			}
		}
	})

	// 18c. AttributeAllocationSystem - manages core attribute point allocation
	// Players spend attribute points on STR/AGI/INT/VIT/END/LCK to customize builds
	attributeAllocationSystem := NewAttributeAllocationSystem(game.World, config.Seed)
	result.AttributeAllocationSystem = attributeAllocationSystem
	game.World.AddSystem(attributeAllocationSystem)

	// Wire attribute allocation to progression level-ups
	result.ProgressionSystem.AddLevelUpCallback(func(entity *Entity, newLevel int) {
		// Award attribute points per level (default 3)
		comp, ok := entity.GetComponent("attribute_allocation")
		if !ok {
			return
		}
		attrComp, ok := comp.(*AttributeAllocationComponent)
		if !ok {
			return
		}
		attributeAllocationSystem.AwardPoints(entity, attrComp.PointsPerLevel)
	})

	// 18d. TalentSystem - manages talent point allocation and passive bonuses
	// Players spend talent points earned on level-up across 4 categories: Offense, Defense, Utility, Mastery
	talentSystem := NewTalentSystem(game.World)
	result.TalentSystem = talentSystem
	game.World.AddSystem(talentSystem)

	// Wire talent system to progression level-ups (1 talent point per level)
	result.ProgressionSystem.AddLevelUpCallback(func(entity *Entity, newLevel int) {
		comp, ok := entity.GetComponent("talent")
		if !ok {
			// Auto-create talent component on first level up
			talentComp := NewTalentComponent()
			entity.AddComponent(talentComp)
			comp = talentComp
		}
		if talentComp, ok := comp.(*TalentComponent); ok {
			talentComp.AddTalentPoints(1)
		}
	})

	// 18d2. SkillMutationSystem - allows players to customize skills with mutation effects
	// Players collect mutations as loot and apply them to skills for stat/effect modifications
	skillMutationSystem := NewSkillMutationSystemWithLogger(game.World, config.Logger)
	result.SkillMutationSystem = skillMutationSystem
	game.World.AddSystem(skillMutationSystem)

	// 18d3. BuildTemplateSystem - manages complete character build presets
	// Players can save and load full character builds (attributes + talents + skills) as templates
	// Includes predefined archetype presets (Tank, DPS, Support, Hybrid, etc.)
	buildTemplateSystem := NewBuildTemplateSystem(game.World, config.Seed)
	result.BuildTemplateSystem = buildTemplateSystem
	game.World.AddSystem(buildTemplateSystem)

	// 18e. EquipmentSetBonusSystem - tracks equipped set pieces and applies tiered bonuses
	// Players wearing multiple pieces from the same equipment set receive cumulative stat bonuses
	equipmentSetBonusSystem := NewEquipmentSetBonusSystem(config.Seed, config.GenreID, config.Logger)
	result.EquipmentSetBonusSystem = equipmentSetBonusSystem
	game.World.AddSystem(equipmentSetBonusSystem)

	// 18f. ElementalWeaponEffectSystem - animates elemental weapon visual effects
	// Weapons with fire, ice, lightning, poison, holy, or shadow enchantments display
	// animated visual effects (flames, frost, sparks, etc.) that update each frame
	elementalWeaponEffectSystem := NewElementalWeaponEffectSystem(game.World)
	result.ElementalWeaponEffectSystem = elementalWeaponEffectSystem
	game.World.AddSystem(elementalWeaponEffectSystem)

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
	result.SpecializationManaBoostSystem = regGenre(game.World,
		NewSpecializationManaBoostSystem(game.World, config.Seed+6600), config.GenreID)

	// 25b. SpecializationHealthRegenSystem - passive health regen bonuses from class specializations
	// Connects ClassProgressionComponent with HealthComponent for genre-aware tank/healer bonuses
	result.SpecializationHealthRegenSystem = regGenre(game.World,
		NewSpecializationHealthRegenSystem(game.World, config.Seed+6650), config.GenreID)

	// 25c. SpecializationSpellDamageSystem - passive spell damage bonuses from class specializations
	// Connects ClassProgressionComponent with spell damage for genre-aware caster bonuses
	result.SpecializationSpellDamageSystem = regGenre(game.World,
		NewSpecializationSpellDamageSystem(game.World, config.Seed+6660), config.GenreID)

	// 25d. SpecializationAttackSpeedSystem - passive attack speed bonuses from class specializations
	// Connects ClassProgressionComponent with AttackComponent cooldowns for genre-aware melee bonuses
	result.SpecializationAttackSpeedSystem = regGenre(game.World,
		NewSpecializationAttackSpeedSystem(game.World, config.Seed+6670), config.GenreID)

	// 25e. SpecializationDefenseSystem - passive defense bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.Defense for genre-aware tank bonuses
	result.SpecializationDefenseSystem = regGenre(game.World,
		NewSpecializationDefenseSystem(game.World, config.Seed+6680), config.GenreID)

	// 25.2a. DualClassSynergySystem - passive bonuses for dual-classed characters
	// Connects ClassProgressionComponent.SecondaryClass with stat bonuses based on class combination
	result.DualClassSynergySystem = regGenre(game.World,
		NewDualClassSynergySystem(game.World, config.Seed+6685), config.GenreID)

	// 25f. SpecializationStatusResistSystem - status effect duration modifiers from class specializations
	// Connects ClassProgressionComponent with StatusEffectComponent for genre-aware debuff resistance
	result.SpecializationStatusResistSystem = regGenre(game.World,
		NewSpecializationStatusResistSystem(game.World, config.Seed+6690), config.GenreID)

	// 25h. SpecializationCritDamageSystem - critical hit damage bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.CritDamage for rogue/assassin bonuses
	result.SpecializationCritDamageSystem = regGenre(game.World,
		NewSpecializationCritDamageSystem(game.World, config.Seed+6700), config.GenreID)

	// 25i. SpecializationEvasionSystem - evasion bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.Evasion for rogue/monk bonuses
	result.SpecializationEvasionSystem = regGenre(game.World,
		NewSpecializationEvasionSystem(game.World, config.Seed+6710), config.GenreID)

	// 25j. SpecializationLifestealSystem - lifesteal bonuses from class specializations
	// Connects ClassProgressionComponent with StatsComponent.Lifesteal for blood magic/survival bonuses
	result.SpecializationLifestealSystem = regGenre(game.World,
		NewSpecializationLifestealSystem(game.World, config.Seed+6720), config.GenreID)

	// 25g. StatusEffectHealthRegenSystem - health regeneration modifiers from status effects
	// Connects StatusEffectSystem with HealthComponent for genre-aware regen bonuses/penalties
	result.StatusEffectHealthRegenSystem = regGenre(game.World,
		NewStatusEffectHealthRegenSystem(game.World, config.Seed+6695), config.GenreID)

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

	result.CraftingSystem = NewCraftingSystem(game.World, inventorySystem, item.NewItemGenerator())
	game.World.AddSystem(result.CraftingSystem)

	// 30. InteractionSystem - puzzle element interactions (Phase 11.2)
	result.InteractionSystem = NewInteractionSystem(game.World)
	game.World.AddSystem(result.InteractionSystem)

	// 30b. ScriptingSystem - ECS-integrated sandbox for mod script execution
	// Evaluates Script components on entities for the sandboxed modding system
	scriptingSystem := NewScriptingSystem(game.World)
	game.World.AddSystem(scriptingSystem)

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

	// Creature visual classifier — must run before animation system so entity type
	// is available when sprite configs are built.
	result.CreatureVisualClassifierSystem = NewCreatureVisualClassifierSystem(game.World)
	game.World.AddSystem(result.CreatureVisualClassifierSystem)

	// Size sprite scaling — propagates entity size class to sprite pipeline
	// so anatomy proportions adapt to entity size (Tiny→chibi, Huge→massive).
	// Must run after creature visual classifier and before animation system.
	result.SizeSpriteScalingSystem = NewSizeSpriteScalingSystem(game.World)
	game.World.AddSystem(result.SizeSpriteScalingSystem)

	// NPC role visual system — infers humanoid visual roles (mage, warrior, merchant, etc.)
	// Must run before animation/directional systems that build sprite configs.
	result.NpcRoleVisualSystem = NewNpcRoleVisualSystem(game.World, config.Seed+9050)
	game.World.AddSystem(result.NpcRoleVisualSystem)

	result.AnimationSystem = NewAnimationSystem(spriteGenerator)
	result.AnimationSystemWrapper = &animationSystemWrapper{
		system: result.AnimationSystem,
		logger: game.World.GetLogger(),
	}
	game.World.AddSystem(result.AnimationSystemWrapper)

	equipmentVisualSystem := NewEquipmentVisualSystem(spriteGenerator)
	game.World.AddSystem(equipmentVisualSystem)

	// Directional sprites — generates 4-directional sprite variants (up/down/left/right)
	// so the render system displays the correct facing. Runs after animation and equipment
	// visuals but before the sprite finalizer.
	result.DirectionalSpriteSystem = NewDirectionalSpriteSystem(spriteGenerator)
	result.DirectionalSpriteSystem.SetGenre(config.GenreID)
	game.World.AddSystem(result.DirectionalSpriteSystem)

	// Sprite depth enhance — applies form-aware volumetric shading (spherical
	// heads, cylindrical torsos, tubular limbs) for 3D depth perception.
	// Runs after directional sprites, before finalization.
	result.SpriteDepthEnhanceSystem = NewSpriteDepthEnhanceSystem(game.World, config.Seed+9050)
	result.SpriteDepthEnhanceSystem.SetGenre(config.GenreID)
	game.World.AddSystem(result.SpriteDepthEnhanceSystem)

	// Sprite color temperature: warm/cool color grading and specular highlights.
	// Runs after depth enhancement, before finalization, for genre-aware lighting.
	result.SpriteColorTemperatureSystem = NewSpriteColorTemperatureSystem(game.World, config.Seed+9075)
	result.SpriteColorTemperatureSystem.SetGenre(config.GenreID)
	game.World.AddSystem(result.SpriteColorTemperatureSystem)

	// Sprite finalizer — applies adaptive outline, rim lighting, and edge shadow
	// to entity sprites for visual clarity in top-down view. Runs after equipment
	// visuals so finalization covers equipment overlays.
	result.SpriteFinalizerSystem = NewSpriteFinalizerSystem(game.World, config.Seed+9100)
	game.World.AddSystem(result.SpriteFinalizerSystem)

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
	result.StatusEffectEvasionSystem = regGenre(game.World,
		NewStatusEffectEvasionSystem(game.World, config.Seed+2120), config.GenreID)

	// 36b2c. StatusEffectCriticalChanceSystem - applies crit chance modifiers from status effects
	// Connects StatusEffectSystem with CombatSystem for genre-aware critical hit bonuses/penalties
	result.StatusEffectCriticalChanceSystem = regGenre(game.World,
		NewStatusEffectCriticalChanceSystem(game.World, config.Seed+2130), config.GenreID)

	// 36b2d. StatusEffectDamageBoostSystem - applies damage modifiers from status effects
	// Connects StatusEffectSystem with CombatSystem for genre-aware damage bonuses/penalties
	result.StatusEffectDamageBoostSystem = regGenre(game.World,
		NewStatusEffectDamageBoostSystem(game.World, config.Seed+2140), config.GenreID)

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

	// 36b7c. TerrainEquipmentDurabilitySystem - degrades equipment from terrain hazards
	// Connects terrain tiles (lava, water, traps) with EquipmentComponent durability for environmental wear
	terrainEquipmentDurabilitySystem := NewTerrainEquipmentDurabilitySystem(game.World, config.Seed+2194)
	terrainEquipmentDurabilitySystem.SetGenre(config.GenreID)
	terrainEquipmentDurabilitySystem.SetTileSize(config.TileSize)
	result.TerrainEquipmentDurabilitySystem = terrainEquipmentDurabilitySystem
	game.World.AddSystem(terrainEquipmentDurabilitySystem)

	// 36b7c2. TerrainEquipmentDurabilityParticleSystem - visual feedback for terrain equipment damage
	// Connects TerrainEquipmentDurabilitySystem with ParticleSystem for genre-aware damage particles
	terrainEquipmentDurabilityParticleSystem := NewTerrainEquipmentDurabilityParticleSystem(game.World, config.Seed+2195)
	terrainEquipmentDurabilityParticleSystem.SetParticleSystem(result.ParticleSystem)
	terrainEquipmentDurabilityParticleSystem.SetTerrainEquipmentDurabilitySystem(terrainEquipmentDurabilitySystem)
	terrainEquipmentDurabilityParticleSystem.SetGenre(config.GenreID)
	terrainEquipmentDurabilityParticleSystem.SetTileSize(config.TileSize)
	result.TerrainEquipmentDurabilityParticleSystem = terrainEquipmentDurabilityParticleSystem
	game.World.AddSystem(terrainEquipmentDurabilityParticleSystem)

	// 36b7c3. TerrainRangedAccuracySystem - modifies ranged accuracy based on terrain type
	// Connects terrain tiles (corridors, trees, platforms) with ranged combat for tactical positioning
	terrainRangedAccuracySystem := NewTerrainRangedAccuracySystem(game.World, config.Seed+2197)
	terrainRangedAccuracySystem.SetGenre(config.GenreID)
	terrainRangedAccuracySystem.SetTileSize(config.TileSize)
	result.TerrainRangedAccuracySystem = terrainRangedAccuracySystem
	game.World.AddSystem(terrainRangedAccuracySystem)

	// 36b7d. WeatherEquipmentDurabilitySystem - degrades equipment from weather conditions
	// Connects WeatherComponent (rain, snow, sandstorm) with EquipmentComponent durability for environmental wear
	result.WeatherEquipmentDurabilitySystem = regGenre(game.World,
		NewWeatherEquipmentDurabilitySystem(game.World, config.Seed+2196), config.GenreID)

	// 36b8. TerrainCompanionBonusSystem - modifies companion stats based on terrain type
	// Connects terrain tiles with CompanionStatsComponent for tactical companion positioning
	terrainCompanionBonusSystem := NewTerrainCompanionBonusSystem(game.World, config.Seed+2195)
	terrainCompanionBonusSystem.SetGenre(config.GenreID)
	terrainCompanionBonusSystem.SetTileSize(config.TileSize)
	result.TerrainCompanionBonusSystem = terrainCompanionBonusSystem
	game.World.AddSystem(terrainCompanionBonusSystem)

	// 36b8b. TerrainCompanionHealthRegenSystem - heals companions based on terrain type
	// Connects terrain tiles with HealthComponent for type-specific companion health regeneration
	// Companions with PerkExtraHealth bonding perk get boosted terrain regen
	terrainCompanionHealthRegenSystem := NewTerrainCompanionHealthRegenSystem(game.World, config.Seed+2196)
	terrainCompanionHealthRegenSystem.SetGenre(config.GenreID)
	terrainCompanionHealthRegenSystem.SetTileSize(config.TileSize)
	result.TerrainCompanionHealthRegenSystem = terrainCompanionHealthRegenSystem
	game.World.AddSystem(terrainCompanionHealthRegenSystem)

	// 36b9. FactionCompanionBehaviorSystem - modifies companion targeting based on faction reputation
	// Connects FactionSystem reputation with companion AI targeting for faction-aware companions
	result.FactionCompanionBehaviorSystem = regGenre(game.World,
		NewFactionCompanionBehaviorSystem(game.World, config.Seed+2198), config.GenreID)

	// 36b10. WeatherCompanionBonusSystem - modifies companion stats based on weather conditions
	// Connects WeatherSystem with CompanionStatsComponent for tactical companion positioning
	result.WeatherCompanionBonusSystem = regGenre(game.World,
		NewWeatherCompanionBonusSystem(game.World, config.Seed+2200), config.GenreID)

	// 36b10b. WeatherCompanionBonusParticleSystem - visual feedback for weather companion bonuses
	// Connects WeatherCompanionBonusSystem with ParticleSystem for genre-aware companion particles
	weatherCompanionBonusParticleSystem := NewWeatherCompanionBonusParticleSystem(game.World, config.Seed+2205)
	weatherCompanionBonusParticleSystem.SetParticleSystem(result.ParticleSystem)
	weatherCompanionBonusParticleSystem.SetWeatherCompanionBonusSystem(result.WeatherCompanionBonusSystem)
	weatherCompanionBonusParticleSystem.SetGenre(config.GenreID)
	result.WeatherCompanionBonusParticleSystem = weatherCompanionBonusParticleSystem
	game.World.AddSystem(weatherCompanionBonusParticleSystem)

	// 36b11. WeatherMovementSpeedSystem - modifies entity movement speed based on weather
	// Connects WeatherComponent with VelocityComponent for environmental movement effects
	// Rain makes surfaces slick (+5% speed), snow creates resistance (-18%), sandstorm drags (-20%)
	result.WeatherMovementSpeedSystem = regGenre(game.World,
		NewWeatherMovementSpeedSystem(game.World, config.Seed+2210), config.GenreID)

	// 36b11b. WeatherMovementSpeedParticleSystem - visual feedback for weather movement modifiers
	// Connects WeatherMovementSpeedSystem with ParticleSystem for genre-aware speed effect particles
	weatherMovementSpeedParticleSystem := NewWeatherMovementSpeedParticleSystem(game.World, config.Seed+2215)
	weatherMovementSpeedParticleSystem.SetParticleSystem(result.ParticleSystem)
	weatherMovementSpeedParticleSystem.SetWeatherMovementSpeedSystem(result.WeatherMovementSpeedSystem)
	weatherMovementSpeedParticleSystem.SetGenre(config.GenreID)
	result.WeatherMovementSpeedParticleSystem = weatherMovementSpeedParticleSystem
	game.World.AddSystem(weatherMovementSpeedParticleSystem)

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

	// 36d1. SkillPointGainParticleSystem - celebratory particles when a skill point is awarded
	skillPointGainParticleSystem := NewSkillPointGainParticleSystem(game.World, config.Seed+4050)
	skillPointGainParticleSystem.SetParticleSystem(result.ParticleSystem)
	skillPointGainParticleSystem.SetGenre(config.GenreID)
	result.ProgressionSystem.AddSkillPointCallback(skillPointGainParticleSystem.OnSkillPointGain)
	result.SkillPointGainParticleSystem = skillPointGainParticleSystem
	game.World.AddSystem(skillPointGainParticleSystem)

	// 36d2. CompanionProgressionSystem - companion XP and leveling
	companionProgressionSystem := NewCompanionProgressionSystem(game.World)
	result.CompanionProgressionSystem = companionProgressionSystem
	result.CompanionProgressionSystemWrapper = &companionProgressionSystemWrapper{system: companionProgressionSystem}
	game.World.AddSystem(result.CompanionProgressionSystemWrapper)

	// 36d2b. CompanionQuestSynergySystem - integrates companion skills with quest bonuses
	// Connects CompanionComponent skills with QuestTrackerComponent for objective and reward multipliers
	companionQuestSynergySystem := NewCompanionQuestSynergySystem(game.World)
	result.CompanionQuestSynergySystem = companionQuestSynergySystem
	game.World.AddSystem(companionQuestSynergySystem)

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

	// 36f1b. SpellChannelParticleSystem - visual feedback during spell channeling
	// Connects SpellSlotComponent.IsCasting() with ParticleSystem for element-colored particles
	spellChannelParticleSystem := NewSpellChannelParticleSystem(game.World, config.Seed+5550)
	spellChannelParticleSystem.SetParticleSystem(result.ParticleSystem)
	spellChannelParticleSystem.SetGenre(config.GenreID)
	result.SpellChannelParticleSystem = spellChannelParticleSystem
	game.World.AddSystem(spellChannelParticleSystem)

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
	result.WeatherManaRegenSystem = regGenre(game.World,
		NewWeatherManaRegenSystem(game.World, config.Seed+6300), config.GenreID)

	// 36j2. WeatherCooldownSystem - modifies spell cooldown rates based on weather
	// Connects WeatherSystem with SpellCastingSystem for magical gameplay synergy
	result.WeatherCooldownSystem = regGenre(game.World,
		NewWeatherCooldownSystem(game.World, config.Seed+6350), config.GenreID)

	// 36j2b. WeatherSpellDamageSystem - modifies spell damage based on weather-element synergies
	// Connects WeatherSystem with SpellCastingSystem for elemental combat bonuses
	// (e.g., lightning +25% in rain, fire -25% in rain, ice +30% in snow)
	weatherSpellDamageSystem := NewWeatherSpellDamageSystem(game.World, config.Seed+6360)
	weatherSpellDamageSystem.SetGenre(config.GenreID)
	spellCastingSystem.SetWeatherSpellDamageSystem(weatherSpellDamageSystem)
	result.WeatherSpellDamageSystem = weatherSpellDamageSystem
	game.World.AddSystem(weatherSpellDamageSystem)

	// 36j3. FishingWeatherBonusSystem - modifies fishing bonuses based on weather
	// Connects WeatherSystem with FishingSystem for immersive fishing gameplay
	result.FishingWeatherBonusSystem = regGenre(game.World,
		NewFishingWeatherBonusSystem(game.World, config.Seed+6375), config.GenreID)

	// 36j4. TerrainFishingBonusSystem - modifies fishing bonuses based on terrain
	// Connects terrain tiles (deep water, kelp, ruins) with FishingSystem for strategic spot placement
	terrainFishingBonusSystem := NewTerrainFishingBonusSystem(game.World, config.Seed+6380)
	terrainFishingBonusSystem.SetGenre(config.GenreID)
	terrainFishingBonusSystem.SetTileSize(config.TileSize)
	result.TerrainFishingBonusSystem = terrainFishingBonusSystem
	game.World.AddSystem(terrainFishingBonusSystem)

	// 36k. LowHealthVFXSystem - visual feedback for low player health
	// Spawns genre-aware pulsing particles when player health is critically low
	lowHealthVFXSystem := NewLowHealthVFXSystem(game.World, config.Seed+6400)
	lowHealthVFXSystem.SetParticleSystem(result.ParticleSystem)
	lowHealthVFXSystem.SetGenre(config.GenreID)
	result.LowHealthVFXSystem = lowHealthVFXSystem
	game.World.AddSystem(lowHealthVFXSystem)

	// 36k1b. HealthRegenPulseSystem - visual feedback for active health regeneration
	healthRegenPulseSystem := NewHealthRegenPulseSystem(game.World, config.Seed+6425)
	healthRegenPulseSystem.SetParticleSystem(result.ParticleSystem)
	healthRegenPulseSystem.SetGenre(config.GenreID)
	result.HealthRegenPulseSystem = healthRegenPulseSystem
	game.World.AddSystem(healthRegenPulseSystem)

	// 36k1c. StatusEffectGroundTrailSystem - ground-level DoT movement trails
	statusEffectGroundTrailSystem := NewStatusEffectGroundTrailSystem(game.World, config.Seed+6435)
	statusEffectGroundTrailSystem.SetParticleSystem(result.ParticleSystem)
	statusEffectGroundTrailSystem.SetGenre(config.GenreID)
	result.StatusEffectGroundTrailSystem = statusEffectGroundTrailSystem
	game.World.AddSystem(statusEffectGroundTrailSystem)

	// 36k1d. WaterSurfaceRippleSystem - ripple/splash particles in water tiles
	waterSurfaceRippleSystem := NewWaterSurfaceRippleSystem(game.World, config.Seed+6440)
	waterSurfaceRippleSystem.SetParticleSystem(result.ParticleSystem)
	waterSurfaceRippleSystem.SetGenre(config.GenreID)
	result.WaterSurfaceRippleSystem = waterSurfaceRippleSystem
	game.World.AddSystem(waterSurfaceRippleSystem)

	// 36k1e. EntityTargetLockIndicatorSystem - genre-aware targeting reticle on closest hostile
	result.EntityTargetLockIndicatorSystem = regGenre(game.World,
		NewEntityTargetLockIndicatorSystem(game.World, config.Seed+6445), config.GenreID)

	// 36k1f. EntityDeathDissolveSystem - genre-aware visual dissolve on entity death
	result.EntityDeathDissolveSystem = regGenre(game.World,
		NewEntityDeathDissolveSystem(game.World, config.Seed+6448), config.GenreID)

	// 36k1g. ArmorHitSparkSystem - material-aware armor deflection particles on damage
	// Connects HealthComponent damage detection with EquipmentComponent armor material
	armorHitSparkSystem := NewArmorHitSparkSystem(game.World, config.Seed+6452)
	armorHitSparkSystem.SetParticleSystem(result.ParticleSystem)
	armorHitSparkSystem.SetGenre(config.GenreID)
	result.ArmorHitSparkSystem = armorHitSparkSystem
	game.World.AddSystem(armorHitSparkSystem)

	// 36k1h. MeleeEnchantmentArcParticleSystem - rarity-colored particles along swing arcs
	// Bridges MeleeSwingArcComponent active state with equipment enchantment rarity
	meleeEnchantmentArcParticleSystem := NewMeleeEnchantmentArcParticleSystem(game.World, config.Seed+6455)
	meleeEnchantmentArcParticleSystem.SetParticleSystem(result.ParticleSystem)
	meleeEnchantmentArcParticleSystem.SetGenre(config.GenreID)
	result.MeleeEnchantmentArcParticleSystem = meleeEnchantmentArcParticleSystem
	game.World.AddSystem(meleeEnchantmentArcParticleSystem)

	// 36k1i. BattleWoundOverlaySystem - genre-aware wound marks based on health percentage
	// Bridges HealthComponent damage events with per-entity wound visual overlays
	result.BattleWoundOverlaySystem = regGenre(game.World,
		NewBattleWoundOverlaySystem(game.World, config.Seed+6458), config.GenreID)

	// 36k1j. EquipmentDamageCrackOverlaySystem - procedural crack patterns from equipment wear
	// Reads EquipmentWearTintComponent crack density and generates genre-aware crack segments
	result.EquipmentDamageCrackOverlaySystem = regGenre(game.World,
		NewEquipmentDamageCrackOverlaySystem(game.World, config.Seed+6460), config.GenreID)

	// 36k1k. WeatherEntityWetnessSystem - rain-driven entity sprite wetness darkening + sheen
	// Reads WeatherComponent rain state and writes WetnessComponent with darken/tint values
	result.WeatherEntityWetnessSystem = regGenre(game.World,
		NewWeatherEntityWetnessSystem(game.World, config.Seed+6465), config.GenreID)

	// 36k1l. AttackTelegraphGlowSystem - genre-aware attack wind-up warning glow
	// Reads AIComponent state and AttackComponent cooldown to ramp a visual telegraph
	result.AttackTelegraphGlowSystem = regGenre(game.World,
		NewAttackTelegraphGlowSystem(game.World, config.Seed+6470), config.GenreID)

	// 36k1m. XPGainBurstSystem - genre-aware XP gain particle bursts
	// Detects ExperienceComponent.TotalXP changes and spawns upward-rising particles
	xpGainBurstSystem := NewXPGainBurstSystem(game.World, config.Seed+6475)
	xpGainBurstSystem.SetParticleSystem(result.ParticleSystem)
	xpGainBurstSystem.SetGenre(config.GenreID)
	result.XPGainBurstSystem = xpGainBurstSystem
	game.World.AddSystem(xpGainBurstSystem)

	// 36k1n. EquipmentGleamSweepSystem - animated specular gleam sweep across equipment
	// Reads MaterialSheenComponent and animates a moving highlight band per material/rarity
	result.EquipmentGleamSweepSystem = regGenre(game.World,
		NewEquipmentGleamSweepSystem(game.World, config.Seed+6480), config.GenreID)

	// 36k1o. SpriteDepthShadingSystem - per-body-part depth shading for entity sprites
	// Attaches genre-aware shading parameters (highlight, edge darkening, AO, dithering)
	result.SpriteDepthShadingSystem = regGenre(game.World,
		NewSpriteDepthShadingSystem(game.World, config.Seed+6485), config.GenreID)

	// 36k1p. ClothingPatternSystem - seed-based clothing patterns for entity sprites
	// Attaches genre-aware garment patterns (stripes, checks, dots, borders, etc.)
	result.ClothingPatternSystem = regGenre(game.World,
		NewClothingPatternSystem(game.World, config.Seed+6490), config.GenreID)

	// 36k1p2. SurfaceTextureSystem - creature-form-specific procedural surface textures
	// Assigns fur, scales, chitin, metal, bone, ooze, feather, bark textures to
	// nonhumanoid entities based on their CreatureVisualComponent form.
	result.SurfaceTextureSystem = regGenre(game.World,
		NewSurfaceTextureSystem(game.World, config.Seed+6495), config.GenreID)

	// 36k1p3. HumanoidTextureSystem - humanoid-specific surface textures
	// Assigns skin textures (freckled, scarred, weathered, tattooed), clothing
	// fabric textures (linen, leather, silk, wool, chainmail, plate), and hair
	// textures (straight, wavy, curly, braided) to humanoid entities.
	result.HumanoidTextureSystem = regGenre(game.World,
		NewHumanoidTextureSystem(game.World, config.Seed+6497), config.GenreID)

	// 36k1q. BodyTypeSystem - seed-based body type variety for entity sprites
	// Assigns distinct body builds (stocky, lean, muscular, heavy, etc.) per entity
	result.BodyTypeSystem = regGenre(game.World, NewBodyTypeSystem(game.World, config.Seed+6500), config.GenreID)

	// 36k1r. HeadgearAssignmentSystem - seed-based headgear variety for humanoid entities
	// Assigns genre- and role-aware headgear types (crowns, hoods, wizard hats, circlets, etc.)
	result.HeadgearAssignmentSystem = regGenre(game.World,
		NewHeadgearAssignmentSystem(game.World, config.Seed+6510), config.GenreID)

	// 36k1s. BackAccessorySystem - seed-based back accessories for humanoid entities
	// Assigns genre- and role-aware back accessories (capes, cloaks, quivers, backpacks, etc.)
	result.BackAccessorySystem = regGenre(game.World, NewBackAccessorySystem(game.World, config.Seed+6520), config.GenreID)

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
	result.WeatherRangedAccuracySystem = regGenre(game.World,
		NewWeatherRangedAccuracySystem(game.World, config.Seed+6750), config.GenreID)

	// 36n2. WeatherCritChanceSystem - modifies critical hit chance based on weather conditions
	// Connects WeatherSystem with StatsComponent.CritChance for tactical crit bonuses/penalties
	result.WeatherCritChanceSystem = regGenre(game.World,
		NewWeatherCritChanceSystem(game.World, config.Seed+6775), config.GenreID)

	// 36n3. WeatherBlockChanceSystem - modifies block chance based on weather conditions
	// Connects WeatherSystem with StatsComponent.BlockChance for tactical block penalties
	// Rain makes shields slippery, snow slows reflexes, storms hamper defense
	result.WeatherBlockChanceSystem = regGenre(game.World,
		NewWeatherBlockChanceSystem(game.World, config.Seed+6780), config.GenreID)

	// 36o. WeatherXPBonusSystem - modifies XP gains based on weather conditions
	// Connects WeatherSystem with ProgressionSystem for genre-aware XP bonuses
	weatherXPBonusSystem := NewWeatherXPBonusSystem(game.World, config.Seed+6800)
	weatherXPBonusSystem.SetGenre(config.GenreID)
	weatherXPBonusSystem.SetProgressionSystem(progressionSystem)
	result.WeatherXPBonusSystem = weatherXPBonusSystem
	game.World.AddSystem(weatherXPBonusSystem)

	// 36p. ElementalComboDamageSystem - applies bonus damage when elemental status effects combine
	// Connects StatusEffectSystem with CombatSystem for genre-aware elemental combo damage
	result.ElementalComboDamageSystem = regGenre(game.World,
		NewElementalComboDamageSystem(game.World, config.Seed+6850), config.GenreID)

	// 36p2. WeatherElementalComboBonusSystem - modifies elemental combo damage based on weather
	// Connects WeatherSystem with ElementalComboDamageSystem for environmental synergies
	weatherElementalComboBonusSystem := NewWeatherElementalComboBonusSystem(game.World, config.Seed+6855)
	weatherElementalComboBonusSystem.SetElementalComboDamageSystem(result.ElementalComboDamageSystem)
	weatherElementalComboBonusSystem.SetGenre(config.GenreID)
	result.WeatherElementalComboBonusSystem = weatherElementalComboBonusSystem
	game.World.AddSystem(weatherElementalComboBonusSystem)

	// 36q. ElementalCompanionSynergySystem - boosts elemental companions based on owner status effects
	// Connects CompanionComponent (elemental type) with StatusEffectComponent for tactical synergy
	result.ElementalCompanionSynergySystem = regGenre(game.World,
		NewElementalCompanionSynergySystem(game.World, config.Seed+6900), config.GenreID)

	// 36q2. CompanionSpellAmplificationSystem - boosts owner spell damage/healing based on companion bonding
	// Connects CompanionComponent (loyalty, bonding perks) with owner spell effectiveness
	result.CompanionSpellAmplificationSystem = regGenre(game.World,
		NewCompanionSpellAmplificationSystem(game.World, config.Seed+6910), config.GenreID)

	// 36q3. CompanionManaRegenSystem - boosts owner mana regeneration based on companion bonding
	// Connects CompanionComponent (loyalty, type, perks) with ManaComponent for sustained spellcasting
	result.CompanionManaRegenSystem = regGenre(game.World,
		NewCompanionManaRegenSystem(game.World, config.Seed+6920), config.GenreID)

	// 36q4. CompanionFishingBonusSystem - boosts fishing rates based on nearby companions
	// Connects CompanionComponent (type, loyalty) with FishingComponent for companion-assisted fishing
	result.CompanionFishingBonusSystem = regGenre(game.World,
		NewCompanionFishingBonusSystem(game.World, config.Seed+6930), config.GenreID)

	// 36r. LifestealSystem - heals attackers based on damage dealt
	// Connects CombatSystem damage events with HealthComponent healing for sustained combat
	lifestealSystem := NewLifestealSystem(game.World, config.Seed+6950)
	lifestealSystem.SetParticleSystem(result.ParticleSystem)
	lifestealSystem.SetGenre(config.GenreID)
	result.CombatSystem.SetDamageCallback(lifestealSystem.OnDamageDealt)
	result.LifestealSystem = lifestealSystem
	game.World.AddSystem(lifestealSystem)

	// 36r2. CompanionDamageLifestealSystem - heals owners when companions deal damage
	// Connects CompanionComponent with CombatSystem damage events for pet-based sustain builds
	companionDamageLifestealSystem := NewCompanionDamageLifestealSystem(game.World, config.Seed+6960)
	companionDamageLifestealSystem.SetParticleSystem(result.ParticleSystem)
	companionDamageLifestealSystem.SetGenre(config.GenreID)
	result.CombatSystem.AddDamageCallback(companionDamageLifestealSystem.OnCompanionDamageDealt)
	result.CompanionDamageLifestealSystem = companionDamageLifestealSystem
	game.World.AddSystem(companionDamageLifestealSystem)

	// 36s. StatusEffectManaCostSystem - modifies spell mana costs based on status effects
	// Connects StatusEffectSystem with SpellCastingSystem for tactical spell management
	result.StatusEffectManaCostSystem = regGenre(game.World,
		NewStatusEffectManaCostSystem(game.World, config.Seed+7000), config.GenreID)

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

	// 36u2. CombatEquipmentDurabilityParticleSystem - visual feedback for armor damage absorption
	// Connects CombatSystem damage events with ParticleSystem for armor wear particle effects
	combatEquipmentDurabilityParticleSystem := NewCombatEquipmentDurabilityParticleSystem(game.World, config.Seed+7125)
	combatEquipmentDurabilityParticleSystem.SetParticleSystem(result.ParticleSystem)
	combatEquipmentDurabilityParticleSystem.SetGenre(config.GenreID)
	result.CombatSystem.AddDamageCallback(combatEquipmentDurabilityParticleSystem.OnDamageTaken)
	result.CombatEquipmentDurabilityParticleSystem = combatEquipmentDurabilityParticleSystem
	game.World.AddSystem(combatEquipmentDurabilityParticleSystem)

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
	result.WeatherMeleeDamageSystem = regGenre(game.World,
		NewWeatherMeleeDamageSystem(game.World, config.Seed+7400), config.GenreID)

	// 36ab. WeatherAttackSpeedSystem - modifies melee attack cooldowns based on weather
	// Connects WeatherSystem with AttackComponent for genre-aware attack speed modifiers
	// Cold slows attacks, storms enable quick strikes
	result.WeatherAttackSpeedSystem = regGenre(game.World,
		NewWeatherAttackSpeedSystem(game.World, config.Seed+7425), config.GenreID)

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

	// 42b. HazardProximityWarningParticleSystem - visual warning near hazard zones
	// Connects HazardSystem zone tracker with ParticleSystem for genre-aware proximity warnings
	hazardWarningSystem := NewHazardProximityWarningParticleSystem(game.World, config.Seed+7500)
	hazardWarningSystem.SetHazardSystem(hazardSystem)
	hazardWarningSystem.SetParticleSystem(result.ParticleSystem)
	hazardWarningSystem.SetGenre(config.GenreID)
	result.HazardProximityWarningParticleSystem = hazardWarningSystem
	game.World.AddSystem(hazardWarningSystem)

	// 43. NarrativeSystem - story progression (Phase 12.2)
	narrativeSystem := NewNarrativeSystem(game.World)
	game.World.AddSystem(narrativeSystem)

	// 43. ShadowSystem - enhanced lighting (Phase 14)
	shadowSystem := NewShadowSystemWithLogger(game.World, logger)
	game.World.AddSystem(shadowSystem)

	// 43b. TimeOfDayLightingSystem - day/night ambient light modulation (Phase 17.3)
	// Connects palette/timeofday.go with AmbientLightComponent for dynamic lighting
	result.TimeOfDayLightingSystem = regGenre(game.World,
		NewTimeOfDayLightingSystem(game.World, config.Seed+7500), config.GenreID)

	// 43c. TimeOfDayStealthSystem - day/night AI detection modulation
	// Connects TimeOfDayLightingSystem with AIComponent detection ranges for stealth
	timeOfDayStealthSystem := NewTimeOfDayStealthSystem(game.World, config.Seed+7550)
	timeOfDayStealthSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayStealthSystem.SetGenre(config.GenreID)
	result.TimeOfDayStealthSystem = timeOfDayStealthSystem
	game.World.AddSystem(timeOfDayStealthSystem)

	// 43d. TimeOfDayXPBonusSystem - day/night XP bonus modulation
	// Connects TimeOfDayLightingSystem with ProgressionSystem for time-based XP bonuses
	timeOfDayXPBonusSystem := NewTimeOfDayXPBonusSystem(game.World, config.Seed+7575)
	timeOfDayXPBonusSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayXPBonusSystem.SetProgressionSystem(progressionSystem)
	timeOfDayXPBonusSystem.SetGenre(config.GenreID)
	result.TimeOfDayXPBonusSystem = timeOfDayXPBonusSystem
	game.World.AddSystem(timeOfDayXPBonusSystem)

	// 43e. TimeOfDayManaCostSystem - day/night spell mana cost modulation
	// Connects TimeOfDayLightingSystem with SpellCastingSystem for element-based mana costs
	timeOfDayManaCostSystem := NewTimeOfDayManaCostSystem(game.World, config.Seed+7600)
	timeOfDayManaCostSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayManaCostSystem.SetGenre(config.GenreID)
	result.TimeOfDayManaCostSystem = timeOfDayManaCostSystem
	game.World.AddSystem(timeOfDayManaCostSystem)

	// 43f. TimeOfDayCriticalChanceSystem - day/night critical hit chance modulation
	// Connects TimeOfDayLightingSystem with StatsComponent.CritChance for ambush tactics
	timeOfDayCriticalChanceSystem := NewTimeOfDayCriticalChanceSystem(game.World, config.Seed+7625)
	timeOfDayCriticalChanceSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayCriticalChanceSystem.SetGenre(config.GenreID)
	result.TimeOfDayCriticalChanceSystem = timeOfDayCriticalChanceSystem
	game.World.AddSystem(timeOfDayCriticalChanceSystem)

	// 43g. TimeOfDayCompanionBonusSystem - day/night companion stat modulation
	// Connects TimeOfDayLightingSystem with CompanionStatsComponent for companion bonuses
	timeOfDayCompanionBonusSystem := NewTimeOfDayCompanionBonusSystem(game.World, config.Seed+7650)
	timeOfDayCompanionBonusSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayCompanionBonusSystem.SetGenre(config.GenreID)
	result.TimeOfDayCompanionBonusSystem = timeOfDayCompanionBonusSystem
	game.World.AddSystem(timeOfDayCompanionBonusSystem)

	// 43h. TimeOfDayHealthRegenSystem - day/night health regeneration modulation
	// Connects TimeOfDayLightingSystem with HealthComponent for genre-aware health regen
	timeOfDayHealthRegenSystem := NewTimeOfDayHealthRegenSystem(game.World, config.Seed+7675)
	timeOfDayHealthRegenSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayHealthRegenSystem.SetGenre(config.GenreID)
	result.TimeOfDayHealthRegenSystem = timeOfDayHealthRegenSystem
	game.World.AddSystem(timeOfDayHealthRegenSystem)

	// 43i. TimeOfDayManaRegenSystem - day/night mana regeneration modulation
	// Connects TimeOfDayLightingSystem with ManaComponent for genre-aware mana regen
	timeOfDayManaRegenSystem := NewTimeOfDayManaRegenSystem(game.World, config.Seed+7700)
	timeOfDayManaRegenSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayManaRegenSystem.SetGenre(config.GenreID)
	result.TimeOfDayManaRegenSystem = timeOfDayManaRegenSystem
	game.World.AddSystem(timeOfDayManaRegenSystem)

	// 43j. TimeOfDayBlockChanceSystem - day/night block chance modulation
	// Connects TimeOfDayLightingSystem with StatsComponent.BlockChance for defensive timing
	timeOfDayBlockChanceSystem := NewTimeOfDayBlockChanceSystem(game.World, config.Seed+7725)
	timeOfDayBlockChanceSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayBlockChanceSystem.SetGenre(config.GenreID)
	result.TimeOfDayBlockChanceSystem = timeOfDayBlockChanceSystem
	game.World.AddSystem(timeOfDayBlockChanceSystem)

	// 43j2. TimeOfDayEvasionSystem - day/night evasion modulation
	// Connects TimeOfDayLightingSystem with StatsComponent.Evasion for shadow-aided dodging
	// Night enhances evasion (+5% base, up to +8% fantasy), day impairs it (-2% visibility penalty)
	timeOfDayEvasionSystem := NewTimeOfDayEvasionSystem(game.World, config.Seed+7730)
	timeOfDayEvasionSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayEvasionSystem.SetGenre(config.GenreID)
	result.TimeOfDayEvasionSystem = timeOfDayEvasionSystem
	game.World.AddSystem(timeOfDayEvasionSystem)

	// 43k. TimeOfDaySpellDamageSystem - day/night spell damage modulation
	// Connects TimeOfDayLightingSystem with spell damage for element-based bonuses
	// Fire/Light spells stronger during day, Dark/Arcane stronger at night
	timeOfDaySpellDamageSystem := NewTimeOfDaySpellDamageSystem(game.World, config.Seed+7750)
	timeOfDaySpellDamageSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDaySpellDamageSystem.SetGenre(config.GenreID)
	result.TimeOfDaySpellDamageSystem = timeOfDaySpellDamageSystem
	game.World.AddSystem(timeOfDaySpellDamageSystem)

	// 43l. TimeOfDayAttackSpeedSystem - day/night melee attack speed modulation
	// Connects TimeOfDayLightingSystem with AttackComponent.Cooldown for combat pacing
	// Night enables faster attacks (cover), day slightly slows them (visibility aids defense)
	timeOfDayAttackSpeedSystem := NewTimeOfDayAttackSpeedSystem(game.World, config.Seed+7775)
	timeOfDayAttackSpeedSystem.SetLightingSystem(result.TimeOfDayLightingSystem)
	timeOfDayAttackSpeedSystem.SetGenre(config.GenreID)
	result.TimeOfDayAttackSpeedSystem = timeOfDayAttackSpeedSystem
	game.World.AddSystem(timeOfDayAttackSpeedSystem)

	// 43m. Connect TimeOfDayShadowDirectionSystem to TimeOfDayLightingSystem
	result.TimeOfDayShadowDirectionSystem.SetLightingSystem(result.TimeOfDayLightingSystem)

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

	// ========================================================================
	// ADDITIONAL SYSTEMS (previously uninstantiated)
	// ========================================================================

	// WorldPersistenceSystem - manages world state persistence across sessions
	worldPersistenceSystem := NewWorldPersistenceSystem()
	game.World.AddSystem(worldPersistenceSystem)

	// ChallengeSystem - manages daily/weekly challenge lifecycle and reset
	challengeSystem := NewChallengeSystem(game.World)
	game.World.AddSystem(challengeSystem)

	// CollectionSystem - tracks collectible discovery and completion
	collectionSystem := NewCollectionSystem(game.World)
	game.World.AddSystem(collectionSystem)

	// EconomyTerritoryIntegrationSystem - bridges territory control with marketplace pricing
	economyTerritorySystem := NewEconomyTerritoryIntegrationSystem(game.World, EconomyTerritoryConfig{})
	game.World.AddSystem(economyTerritorySystem)

	// VoiceSettingsSystem - applies voice settings (volume, codec quality) at runtime
	voiceSettingsSystem := NewVoiceSettingsSystem(game.World)
	game.World.AddSystem(voiceSettingsSystem)

	// Provide a shared game clock for scheduled systems
	if game.World.Clock == nil {
		game.World.Clock = NewRealTimeClock()
	}

	// CityEvolutionSystem - processes city evolution triggers over time
	cityEvolutionSystem := NewCityEvolutionSystem(game.World, game.World.Clock)
	game.World.AddSystem(cityEvolutionSystem)

	// AvailabilitySystem - enforces operating-hours gating on shops, NPCs, and services.
	// Wired after Clock initialisation so game.World.Clock is guaranteed non-nil.
	availabilitySystem := NewAvailabilitySystem(game.World, game.World.Clock)
	result.AvailabilitySystem = availabilitySystem
	game.World.AddSystem(availabilitySystem)

	// Seasonal-event cluster (G2 / G4): all four systems are registered and the
	// SeasonalEventComponent is seeded on a dedicated world entity so the cluster
	// is fully active from the start.  The seed is derived deterministically so
	// different worlds produce different event calendars.
	eventCalendarSystem := NewEventCalendarSystem(game.World, game.World.Clock)
	result.EventCalendarSystem = eventCalendarSystem
	game.World.AddSystem(eventCalendarSystem)

	eventQuestSystem := NewEventQuestSystem(game.World, game.World.Clock)
	result.EventQuestSystem = eventQuestSystem
	game.World.AddSystem(eventQuestSystem)

	eventDecorationSystem := NewEventDecorationSystem(game.World)
	result.EventDecorationSystem = eventDecorationSystem
	game.World.AddSystem(eventDecorationSystem)

	eventRewardSystem := NewEventRewardSystem(game.World, game.World.Clock)
	result.EventRewardSystem = eventRewardSystem
	game.World.AddSystem(eventRewardSystem)

	// G4: Seed the SeasonalEventComponent on a world-level entity so the event
	// calendar systems have something to act on.  useRealTime=true so events
	// align with wall-clock dates (festivals, holidays) rather than game ticks.
	const seasonalSeedMagic = int64(0x53454153) // "SEAS"
	seasonalWorldEntity := game.World.CreateEntity()
	seasonalWorldEntity.AddComponent(NewSeasonalEventComponent(config.Seed^seasonalSeedMagic, true))
	game.World.AddEntity(seasonalWorldEntity)

	// Mod systems (G2 / G3): ModCompatibilitySystem runs conflict detection;
	// ModBrowserSystem is wired with its repository by cmd/client after init.
	modCompatibilitySystem := NewModCompatibilitySystem(game.World)
	result.ModCompatibilitySystem = modCompatibilitySystem
	game.World.AddSystem(modCompatibilitySystem)

	modBrowserSystem := NewModBrowserSystem(game.World)
	result.ModBrowserSystem = modBrowserSystem
	game.World.AddSystem(modBrowserSystem)

	// G10: ExtendedAchievementSystem tracks gameplay events (kills, quests, crafting,
	// exploration, PvP) with multi-tier progression.  It complements AchievementSystem
	// (expression/social) and is safe to register alongside it.
	extendedAchievementSystem := NewExtendedAchievementSystem(game.World)
	result.ExtendedAchievementSystem = extendedAchievementSystem
	game.World.AddSystem(extendedAchievementSystem)

	if config.EnableVerboseLogging {
		logger.WithFields(logrus.Fields{
			"systemCount": len(game.World.GetSystems()),
			"seed":        config.Seed,
			"genre":       config.GenreID,
		}).Info("game systems initialized successfully (spatial partition system requires terrain)")
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
