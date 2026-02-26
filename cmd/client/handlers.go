//go:build !android && !ios
// +build !android,!ios

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/class/advanced"
	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/physics/fluids"
	"github.com/opd-ai/venture/pkg/engine/physics/vehicle"
	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/network/federation"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/building"
	"github.com/opd-ai/venture/pkg/procgen/faction"
	"github.com/opd-ai/venture/pkg/procgen/furniture"
	genrepkg "github.com/opd-ai/venture/pkg/procgen/genre"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/opd-ai/venture/pkg/procgen/recipe"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/quality"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/opd-ai/venture/pkg/social/persistence"
	"github.com/opd-ai/venture/pkg/version"
	"github.com/opd-ai/venture/pkg/world"
	"github.com/opd-ai/venture/pkg/world/housing"
	"github.com/opd-ai/venture/pkg/world/raids"
	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"

	// INTEGRATION FIX [Category A]: V7.0 Display System Import
	// Gap: Display package existed but was never imported for initialization
	// Fix: Added import for display management (1920x1080 default resolution)
	// Roadmap: ROADMAP_V7.md Phase 43
	"github.com/opd-ai/venture/pkg/rendering/display"

	// INTEGRATION FIX [Category A]: V9.0 Integration Manager Imports
	// Gap: Integration packages implemented but never imported for use
	// Fix: Added imports for housing crafting, companion housing, and guild housing managers
	// Roadmap: ROADMAP_V9.md Phase 55.1-55.3
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	guildhousing "github.com/opd-ai/venture/pkg/integration/guild_housing"
	housingcrafting "github.com/opd-ai/venture/pkg/integration/housing_crafting"

	// Phase 4.3: Narrative-World Integration (PLAN.md)
	"github.com/opd-ai/venture/pkg/integration/narrative_world"

	// Phase 4.4: Political Warfare Integration (PLAN.md)
	"github.com/opd-ai/venture/pkg/integration/political_warfare"

	// Phase 6.1: Branching Narratives (PLAN.md)
	"github.com/opd-ai/venture/pkg/narrative/branching"

	// Phase 1.1: Audio Synthesis Engine (PLAN.md)
	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
	"github.com/opd-ai/venture/pkg/audio/sfx"
	"github.com/opd-ai/venture/pkg/audio/synthesis"

	// Phase 1.2: Destruction Physics (PLAN.md)
	"github.com/opd-ai/venture/pkg/engine/physics/destruction"

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	// Note: These packages are wrapped by adapters in pkg/engine/ (LightingAdapter, AnimationAdapter, PostProcessorAdapter)
	// Imports kept for documentation purposes showing integration is complete

	// Ebiten for sprite generation
	"github.com/hajimehoshi/ebiten/v2"
	_ "github.com/opd-ai/venture/pkg/rendering/animation"
	_ "github.com/opd-ai/venture/pkg/rendering/lighting"
	_ "github.com/opd-ai/venture/pkg/rendering/postprocess"

	// Phase 2.3: UI & Shape Rendering (PLAN.md)
	"github.com/opd-ai/venture/pkg/rendering/patterns"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/ui"

	// Phase 3.4: Minigame Implementations (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/minigame/games"

	// Phase 2.4: Rendering Optimization (PLAN.md)
	"github.com/opd-ai/venture/pkg/rendering/parallel"
	"github.com/opd-ai/venture/pkg/rendering/pool"

	// Phase 3.2: Magic & Skills (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/procgen/skills"

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/class"

	// Phase 1.1: Companion Learning System (PLAN.md Integration)
	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/procgen/minigame"
	"github.com/opd-ai/venture/pkg/procgen/puzzle"

	// Phase 3.6: Narrative Integration (PLAN.md)
	"github.com/opd-ai/venture/pkg/procgen/narrative"

	// Phase 4.1: Chat System (PLAN.md)
	"github.com/opd-ai/venture/pkg/network/chat"

	// Phase 4.2: Trade System (PLAN.md)
	"github.com/opd-ai/venture/pkg/network/trade"

	// Phase 5.1: Quality-of-Life System (PLAN.md)
	"github.com/opd-ai/venture/pkg/engine/qol"

	// Phase 1.3: Prestige System (PLAN.md)
	"github.com/opd-ai/venture/pkg/engine/prestige"

	// Phase 4.4: Trade Routes System (PLAN.md Integration - Week 4)
	"github.com/opd-ai/venture/pkg/integration/trade_routes"

	// V19.0: Priority 1 Dormant Package Integration (ROADMAP_V19.md)
	// Phase 99: Entity & Dialog Generation
	"github.com/opd-ai/venture/pkg/procgen/dialog"
	"github.com/opd-ai/venture/pkg/procgen/entity"

	// Phase 100: Legendary & Economy
	"github.com/opd-ai/venture/pkg/procgen/legendary"
	"github.com/opd-ai/venture/pkg/world/economy"

	// Phase 101: Integration Package Activation
	"github.com/opd-ai/venture/pkg/integration/choice_consequences"
	"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
	"github.com/opd-ai/venture/pkg/integration/world_events"
)

// systemsContainer holds all initialized game systems for dependency injection.
type systemsContainer struct {
	inputSystem                                 *engine.InputSystem
	movementSystem                              *engine.MovementSystem
	statisticsSystem                            *engine.StatisticsSystem // Tracks player movement and exploration statistics
	collisionSystem                             *engine.CollisionSystem
	combatSystem                                *engine.CombatSystem
	interactionSystem                           *engine.InteractionSystem
	particleSystem                              *engine.ParticleSystem
	animationSystem                             *engine.AnimationSystem
	equipmentVisualSystem                       *engine.EquipmentVisualSystem
	objectiveTracker                            *engine.ObjectiveTrackerSystem
	aiSystem                                    *engine.AISystem
	progressionSystem                           *engine.ProgressionSystem
	inventorySystem                             *engine.InventorySystem
	commerceSystem                              *engine.CommerceSystem
	reputationPricingSystem                     *engine.ReputationPricingSystem // Connects faction reputation with merchant pricing
	dialogSystem                                *engine.DialogSystem
	craftingSystem                              *engine.CraftingSystem
	audioManager                                *engine.AudioManager
	audioManagerSystem                          *engine.AudioManagerSystem
	itemPickupSystem                            *engine.ItemPickupSystem
	statusEffectSystem                          *engine.StatusEffectSystem
	spellCastingSystem                          *engine.SpellCastingSystem
	playerSpellCasting                          *engine.PlayerSpellCastingSystem
	manaRegenSystem                             *engine.ManaRegenSystem
	playerCombatSystem                          *engine.PlayerCombatSystem
	playerItemUseSystem                         *engine.PlayerItemUseSystem
	rotationSystem                              *engine.RotationSystem
	projectileSystem                            *engine.ProjectileSystem
	revivalSystem                               *engine.RevivalSystem
	behaviorTreeSystem                          *engine.BehaviorTreeSystem
	squadSystem                                 *engine.SquadSystem
	factionSystem                               *engine.FactionSystem
	factionAwareAISystem                        *engine.FactionAwareAISystem                        // Bridges faction reputation with AI hostility
	factionXPBonusSystem                        *engine.FactionXPBonusSystem                        // Bridges faction reputation with XP bonus rewards
	factionDamageBonusSystem                    *engine.FactionDamageBonusSystem                    // Bridges faction reputation with damage bonuses
	weatherFactionResistanceSystem              *engine.WeatherFactionResistanceSystem              // Bridges weather with faction elemental affinities
	reputationDefenseBonusSystem                *engine.ReputationDefenseBonusSystem                // Bridges faction reputation with defense bonuses
	reputationDefenseBonusParticleSystem        *engine.ReputationDefenseBonusParticleSystem        // Visual feedback for reputation defense bonuses
	reputationHealingBonusSystem                *engine.ReputationHealingBonusSystem                // Bridges faction reputation with health regen
	reputationHealingBonusParticleSystem        *engine.ReputationHealingBonusParticleSystem        // Visual feedback for reputation healing
	reputationSpellDamageBonusSystem            *engine.ReputationSpellDamageBonusSystem            // Bridges faction reputation with spell damage
	reputationSpellDamageBonusParticleSystem    *engine.ReputationSpellDamageBonusParticleSystem    // Visual feedback for reputation spell damage
	reputationMovementSpeedSystem               *engine.ReputationMovementSpeedSystem               // Bridges faction reputation with movement speed
	reputationMovementSpeedParticleSystem       *engine.ReputationMovementSpeedParticleSystem       // Visual feedback for reputation speed bonus
	reputationCriticalChanceBonusSystem         *engine.ReputationCriticalChanceBonusSystem         // Bridges faction reputation with critical hit chance
	reputationCriticalChanceParticleSystem      *engine.ReputationCriticalChanceParticleSystem      // Visual feedback for reputation crit bonus
	reputationEquipmentDurabilitySystem         *engine.ReputationEquipmentDurabilitySystem         // Bridges faction reputation with equipment durability
	reputationEquipmentDurabilityParticleSystem *engine.ReputationEquipmentDurabilityParticleSystem // Visual feedback for reputation durability
	reputationCompanionBonusSystem              *engine.ReputationCompanionBonusSystem              // Bridges faction reputation with companion stat bonuses
	reputationCompanionBonusParticleSystem      *engine.ReputationCompanionBonusParticleSystem      // Visual feedback for reputation companion bonus
	reputationQuestGatingSystem                 *engine.ReputationQuestGatingSystem                 // Gates quests behind faction reputation requirements
	factionTerritoryInfluenceSystem             *engine.FactionTerritoryInfluenceSystem             // Faction zones provide combat/progression bonuses based on reputation
	ambientEnvironmentParticleSystem            *engine.AmbientEnvironmentParticleSystem            // Terrain-aware atmospheric ambient particles
	equipmentEnchantmentGlowParticleSystem      *engine.EquipmentEnchantmentGlowParticleSystem      // Rarity-driven enchantment glow particles
	statusEffectVisualOverlaySystem             *engine.StatusEffectVisualOverlaySystem             // Status effect color tints on sprites
	weatherSpriteTintSystem                     *engine.WeatherSpriteTintSystem                     // Weather-driven sprite color tints
	entityDropShadowSystem                      *engine.EntityDropShadowSystem                      // Genre-aware soft drop shadows beneath entities
	timeOfDayShadowDirectionSystem              *engine.TimeOfDayShadowDirectionSystem              // Time-of-day directional shadow offset from sun arc
	equipmentMaterialSheenSystem                *engine.EquipmentMaterialSheenSystem                // Material-based specular highlights on equipment
	equipmentDamageStateTintSystem              *engine.EquipmentDamageStateTintSystem              // Aggregate equipment wear visual tinting
	creatureGenreTintSystem                     *engine.CreatureGenreTintSystem                     // Genre-aware creature/NPC sprite color tinting
	creatureSizeProportionSystem                *engine.CreatureSizeProportionSystem                // Size-based anatomy proportions for creatures
	creatureAnatomySystem                       *engine.CreatureAnatomySystem                       // Assigns anatomy types (quadruped, arachnid, etc.) to creatures
	equipmentRarityDetailSystem                 *engine.EquipmentRarityDetailSystem                 // Rarity-based visual detail scaling for equipment
	npcFacialDetailSystem                       *engine.NpcFacialDetailSystem                       // Genre-aware NPC facial feature parameters
	dynamicExpressionSystem                     *engine.DynamicExpressionSystem                     // Reactive facial expression updates from game state
	projectileTrailParticleSystem               *engine.ProjectileTrailParticleSystem               // Genre-aware projectile trail particles
	entityIdleAmbientParticleSystem             *engine.EntityIdleAmbientParticleSystem             // Genre-aware idle entity ambient particles
	weaponMaterialImpactParticleSystem          *engine.WeaponMaterialImpactParticleSystem          // Material-aware melee impact particles
	damageFlashTintSystem                       *engine.DamageFlashTintSystem                       // Genre-aware damage flash tints
	sprintTrailParticleSystem                   *engine.SprintTrailParticleSystem                   // Genre-aware sprint speed trail particles
	nearbyLightEntityTintSystem                 *engine.NearbyLightEntityTintSystem                 // Light-source-based entity sprite tinting
	weatherEquipmentSheenSystem                 *engine.WeatherEquipmentSheenSystem                 // Weather-driven equipment sheen
	creatureEyeGlowSystem                       *engine.CreatureEyeGlowSystem                       // Genre-aware hostile creature eye glow
	creatureEyePatternSystem                    *engine.CreatureEyePatternSystem                    // Creature-type-specific eye patterns for nonhumanoids
	creatureElementalAuraSystem                 *engine.CreatureElementalAuraSystem                 // Elemental creature aura visuals
	meleeSwingArcSystem                         *engine.MeleeSwingArcSystem                         // Genre-aware melee attack swing arcs
	combatReadyAuraSystem                       *engine.CombatReadyAuraSystem                       // Genre-aware AI combat readiness aura
	aiStateBubbleSystem                         *engine.AIStateBubbleSystem                         // Genre-aware AI state indicator bubbles
	terrainReflectionTintSystem                 *engine.TerrainReflectionTintSystem                 // Terrain-driven sprite color tinting
	movementBobSystem                           *engine.MovementBobSystem                           // Walk-cycle sprite bobbing
	movementLeanSystem                          *engine.MovementLeanSystem                          // Directional movement lean
	spellCastGlowSystem                         *engine.SpellCastGlowSystem                         // Genre-aware spell casting glow
	equipmentChangeFlashSystem                  *engine.EquipmentChangeFlashSystem                  // Equipment change flash particles
	dodgeAfterimageSystem                       *engine.DodgeAfterimageSystem                       // Genre-aware dodge/dash afterimage ghosts
	entityIdleBreathingSystem                   *engine.EntityIdleBreathingSystem                   // Genre-aware idle breathing animation
	combatHitStaggerSystem                      *engine.CombatHitStaggerSystem                      // Genre-aware combat hit stagger offset
	floatingDamageNumberSystem                  *engine.FloatingDamageNumberSystem                  // Genre-aware floating damage numbers
	entityThreatIndicatorSystem                 *engine.EntityThreatIndicatorSystem                 // Genre-aware threat level ring under AI entities
	weaponMaterialParticleSystem                *engine.WeaponMaterialParticleSystem                // Genre-aware idle particles from weapon material type
	lootRarityBeamSystem                        *engine.LootRarityBeamSystem                        // Genre-aware beam particles on ground items by rarity
	damageTypeColorFlashSystem                  *engine.DamageTypeColorFlashSystem                  // Genre-aware elemental damage color flashes
	companionBondTetherSystem                   *engine.CompanionBondTetherSystem                   // Genre-aware visual tether between companion and owner
	criticalHitScreenShakeSystem                *engine.CriticalHitScreenShakeSystem                // Genre-aware camera shake on critical hits
	entitySpawnMaterializeSystem                *engine.EntitySpawnMaterializeSystem                // Genre-aware spawn fade-in visuals
	npcInteractionProximityGlowSystem           *engine.NPCInteractionProximityGlowSystem           // Genre-aware NPC interactability glow
	entityFactionOutlineSystem                  *engine.EntityFactionOutlineSystem                  // Genre-aware faction allegiance outlines
	shieldBubbleOverlaySystem                   *engine.ShieldBubbleOverlaySystem                   // Genre-aware shield bubble visual overlay
	environmentalBreathVaporSystem              *engine.EnvironmentalBreathVaporSystem              // Cold weather breath vapor puffs
	movementDustSystem                          *engine.MovementDustSystem                          // Speed-proportional terrain dust behind fast movers
	healthRegenPulseSystem                      *engine.HealthRegenPulseSystem                      // Genre-aware healing pulse particles
	statusEffectGroundTrailSystem               *engine.StatusEffectGroundTrailSystem               // Ground-level DoT movement trails
	waterSurfaceRippleSystem                    *engine.WaterSurfaceRippleSystem                    // Water tile ripple/splash particles
	entityTargetLockIndicatorSystem             *engine.EntityTargetLockIndicatorSystem             // Genre-aware target lock reticle
	entityDeathDissolveSystem                   *engine.EntityDeathDissolveSystem                   // Genre-aware death dissolve visuals
	armorHitSparkSystem                         *engine.ArmorHitSparkSystem                         // Material-aware armor deflection sparks
	meleeEnchantmentArcParticleSystem           *engine.MeleeEnchantmentArcParticleSystem           // Rarity-colored enchantment arc particles
	battleWoundOverlaySystem                    *engine.BattleWoundOverlaySystem                    // Genre-aware battle wound overlays
	equipmentDamageCrackOverlaySystem           *engine.EquipmentDamageCrackOverlaySystem           // Procedural crack patterns from equipment wear
	weatherEntityWetnessSystem                  *engine.WeatherEntityWetnessSystem                  // Rain-driven entity sprite wetness
	attackTelegraphGlowSystem                   *engine.AttackTelegraphGlowSystem                   // Genre-aware attack wind-up warning glow
	xpGainBurstSystem                           *engine.XPGainBurstSystem                           // Genre-aware XP gain particle bursts
	equipmentGleamSweepSystem                   *engine.EquipmentGleamSweepSystem                   // Animated specular gleam sweep on equipment
	spriteDepthShadingSystem                    *engine.SpriteDepthShadingSystem                    // Per-body-part depth shading for entity sprites
	clothingPatternSystem                       *engine.ClothingPatternSystem                       // Seed-based clothing patterns for entity sprites
	surfaceTextureSystem                        *engine.SurfaceTextureSystem                        // Creature-form surface textures (fur, scales, etc.)
	humanoidTextureSystem                       *engine.HumanoidTextureSystem                       // Humanoid-specific surface textures (skin, fabric, hair)
	bodyTypeSystem                              *engine.BodyTypeSystem                              // Seed-based body type variety for entity sprites
	headgearAssignmentSystem                    *engine.HeadgearAssignmentSystem                    // Seed-based headgear variety for humanoid entities
	backAccessorySystem                         *engine.BackAccessorySystem                         // Seed-based back accessories (capes, cloaks, quivers, etc.)
	creatureVisualClassifierSystem              *engine.CreatureVisualClassifierSystem              // Infers creature visual form from procgen data
	sizeSpriteScalingSystem                     *engine.SizeSpriteScalingSystem                     // Propagates entity size to sprite proportions
	npcRoleVisualSystem                         *engine.NpcRoleVisualSystem                         // Infers humanoid NPC visual role for template selection
	directionalSpriteSystem                     *engine.DirectionalSpriteSystem                     // Generates 4-directional sprite variants for facing
	spriteDepthEnhanceSystem                    *engine.SpriteDepthEnhanceSystem                    // Applies form-aware volumetric depth shading to sprites
	spriteColorTemperatureSystem                *engine.SpriteColorTemperatureSystem                // Applies genre-aware color temperature grading and specular
	spriteFinalizerSystem                       *engine.SpriteFinalizerSystem                       // Applies adaptive outline, rim lighting, edge shadow
	statusEffectAISystem                        *engine.StatusEffectAISystem                        // Bridges status effects with AI (stun/frozen disable AI)
	reputationSystem                            *engine.ReputationSystem
	alignmentSystem                             *engine.AlignmentSystem
	factionReactionSystem                       *engine.FactionReactionSystem
	skillProgressionSystem                      *engine.SkillProgressionSystem
	skillLoadoutSystem                          *engine.SkillLoadoutSystem        // Manages saved skill loadouts for build swapping
	classAffinitySystem                         *engine.ClassAffinitySystem       // Tracks playstyle progression and combat archetypes
	attributeAllocationSystem                   *engine.AttributeAllocationSystem // Manages core attribute point allocation
	talentSystem                                *engine.TalentSystem              // Manages talent point allocation and passive bonuses
	skillMutationSystem                         *engine.SkillMutationSystem       // Manages skill mutations for ability customization
	buildTemplateSystem                         *engine.BuildTemplateSystem       // Manages complete character build presets
	equipmentSetBonusSystem                     *engine.EquipmentSetBonusSystem   // Manages equipment set piece tracking and tiered bonuses
	visualFeedbackSystem                        *engine.VisualFeedbackSystem
	weatherSystem                               *engine.WeatherSystem
	weatherCombatSystem                         *engine.WeatherCombatSystem
	weatherGroundEffectSystem                   *engine.WeatherGroundEffectSystem                // Connects weather to ground impact particle effects
	weatherAudioSystem                          *engine.WeatherAudioSystem                       // Connects weather to ambient audio sounds
	weatherManaRegenSystem                      *engine.WeatherManaRegenSystem                   // Connects weather to mana regeneration rates
	weatherCooldownSystem                       *engine.WeatherCooldownSystem                    // Connects weather to spell cooldown rates
	weatherAttackSpeedSystem                    *engine.WeatherAttackSpeedSystem                 // Connects weather to melee attack speed modifiers
	statusEffectLightingSystem                  *engine.StatusEffectLightingSystem               // Connects status effects to lighting for visual feedback
	statusEffectMovementSystem                  *engine.StatusEffectMovementSystem               // Connects status effects to movement speed modifiers
	statusEffectEvasionSystem                   *engine.StatusEffectEvasionSystem                // Connects status effects to evasion modifiers in combat
	statusEffectCritChanceSystem                *engine.StatusEffectCriticalChanceSystem         // Connects status effects to crit chance modifiers
	terrainMovementSpeedSystem                  *engine.TerrainMovementSpeedSystem               // Connects terrain tiles to movement speed modifiers
	terrainCombatBonusSystem                    *engine.TerrainCombatBonusSystem                 // Connects terrain tiles to combat bonuses (high ground, cover)
	terrainCombatBonusParticleSystem            *engine.TerrainCombatBonusParticleSystem         // Connects terrain combat bonuses to visual particle feedback
	terrainStealthSystem                        *engine.TerrainStealthSystem                     // Connects terrain tiles to AI detection for stealth gameplay
	stealthIndicatorParticleSystem              *engine.StealthIndicatorParticleSystem           // Connects terrain stealth to visual particle feedback
	terrainAmbushCritSystem                     *engine.TerrainAmbushCritSystem                  // Connects terrain stealth to critical hit bonuses for ambush
	terrainStatusEffectSystem                   *engine.TerrainStatusEffectSystem                // Connects terrain tiles (water, lava) to elemental status effects
	terrainManaRegenSystem                      *engine.TerrainManaRegenSystem                   // Connects terrain tiles (water, platforms) to mana regeneration
	terrainSpellDamageSystem                    *engine.TerrainSpellDamageSystem                 // Connects terrain tiles (lava, water) to spell damage modifiers
	terrainEquipmentDurabilitySys               *engine.TerrainEquipmentDurabilitySystem         // Connects terrain hazards (lava, water, traps) to equipment durability
	terrainEquipmentDurabilityParticleSys       *engine.TerrainEquipmentDurabilityParticleSystem // Connects terrain durability to visual particle feedback
	terrainRangedAccuracySys                    *engine.TerrainRangedAccuracySystem              // Connects terrain tiles to ranged combat accuracy
	terrainCompanionBonusSystem                 *engine.TerrainCompanionBonusSystem              // Connects terrain tiles to companion combat stat bonuses
	factionCompanionBehaviorSystem              *engine.FactionCompanionBehaviorSystem           // Connects faction reputation to companion AI targeting
	weatherCompanionBonusSystem                 *engine.WeatherCompanionBonusSystem              // Connects weather conditions to companion combat stat bonuses
	criticalHitParticleSystem                   *engine.CriticalHitParticleSystem                // Connects combat crits to particle effects
	levelUpParticleSystem                       *engine.LevelUpParticleSystem                    // Connects level-ups to particle effects
	itemPickupParticleSystem                    *engine.ItemPickupParticleSystem                 // Connects item pickups to particle effects
	deathParticleSystem                         *engine.DeathParticleSystem                      // Connects entity deaths to particle effects
	spellEffectParticleSystem                   *engine.SpellEffectParticleSystem                // Connects spell effects to particle effects
	damageResistanceParticleSystem              *engine.DamageResistanceParticleSystem           // Connects damage resistance to particle effects
	shieldAbsorbParticleSystem                  *engine.ShieldAbsorbParticleSystem               // Connects shield absorption to particle effects
	shieldRegenSystem                           *engine.ShieldRegenSystem                        // Connects shield regeneration to particle effects
	drowningParticleSystem                      *engine.DrowningParticleSystem                   // Connects drowning state to particle effects
	lowHealthVFXSystem                          *engine.LowHealthVFXSystem                       // Connects low player health to warning particle effects
	manaRegenParticleSystem                     *engine.ManaRegenParticleSystem                  // Connects mana regeneration to visual particle effects
	companionAuraParticleSystem                 *engine.CompanionAuraParticleSystem              // Connects companion bonding perks to aura particles
	elementalComboParticleSystem                *engine.ElementalComboParticleSystem             // Connects elemental status combos to visual effects
	elementalComboDamageSystem                  *engine.ElementalComboDamageSystem               // Connects elemental status combos to bonus damage
	weatherElementalComboBonusSystem            *engine.WeatherElementalComboBonusSystem         // Connects weather to elemental combo damage modifiers
	elementalCompanionSynergySystem             *engine.ElementalCompanionSynergySystem          // Connects elemental companions to owner status effects
	companionSpellAmplificationSystem           *engine.CompanionSpellAmplificationSystem        // Connects companion bonding to owner spell effectiveness
	companionManaRegenSystem                    *engine.CompanionManaRegenSystem                 // Connects companion bonding to owner mana regeneration
	weatherRangedAccuracySystem                 *engine.WeatherRangedAccuracySystem              // Connects weather to ranged attack accuracy modifiers
	weatherCritChanceSystem                     *engine.WeatherCritChanceSystem                  // Connects weather to critical hit chance modifiers
	weatherBlockChanceSystem                    *engine.WeatherBlockChanceSystem                 // Connects weather to block chance modifiers
	weatherXPBonusSystem                        *engine.WeatherXPBonusSystem                     // Connects weather to XP gain bonuses
	lifestealSystem                             *engine.LifestealSystem                          // Connects combat damage to attacker healing
	statusEffectManaCostSystem                  *engine.StatusEffectManaCostSystem               // Connects status effects to spell mana cost modifiers
	statusEffectDamageParticleSystem            *engine.StatusEffectDamageParticleSystem         // Connects status effect ticks (burn, poison, regen) to particle effects
	statusEffectDamageBoostSystem               *engine.StatusEffectDamageBoostSystem            // Connects status effects to damage modifiers for attacks
	fearFleeParticleSystem                      *engine.FearFleeParticleSystem                   // Connects fear status effects with flee particle effects
	combatEquipmentDurabilityParticleSystem     *engine.CombatEquipmentDurabilityParticleSystem  // Connects combat damage to armor durability particle effects
	lifetimeSystem                              *engine.LifetimeSystem
	puzzleSystem                                *engine.PuzzleSystem
	firePropagationSystem                       *engine.FirePropagationSystem
	destructibleSystem                          *engine.DestructibleObjectSystem
	carrySystem                                 *engine.CarrySystem
	hazardSystem                                *engine.HazardSystem
	hazardProximityWarningParticleSystem        *engine.HazardProximityWarningParticleSystem
	narrativeSystem                             *engine.NarrativeSystem
	branchingNarrativeSystem                    *engine.BranchingNarrativeSystem // Phase 6.1: Branching story arc system
	worldEventsSystem                           *engine.WorldEventsSystem        // Phase 6.3: World-responsive events
	shadowSystem                                *engine.ShadowSystem
	timeOfDayLightingSystem                     *engine.TimeOfDayLightingSystem       // Time-of-day ambient lighting modulation
	timeOfDayStealthSystem                      *engine.TimeOfDayStealthSystem        // Connects time-of-day lighting with AI detection for stealth
	timeOfDayXPBonusSystem                      *engine.TimeOfDayXPBonusSystem        // Connects time-of-day lighting with XP bonuses
	timeOfDayManaCostSystem                     *engine.TimeOfDayManaCostSystem       // Connects time-of-day lighting with spell mana costs
	timeOfDayCriticalChanceSystem               *engine.TimeOfDayCriticalChanceSystem // Connects time-of-day lighting with crit chance bonuses
	timeOfDayCompanionBonusSystem               *engine.TimeOfDayCompanionBonusSystem // Connects time-of-day lighting with companion stat bonuses
	timeOfDayManaRegenSystem                    *engine.TimeOfDayManaRegenSystem      // Connects time-of-day lighting with mana regen rates
	timeOfDayBlockChanceSystem                  *engine.TimeOfDayBlockChanceSystem    // Connects time-of-day lighting with block chance bonuses
	timeOfDayEvasionSystem                      *engine.TimeOfDayEvasionSystem        // Connects time-of-day lighting with evasion bonuses
	timeOfDaySpellDamageSystem                  *engine.TimeOfDaySpellDamageSystem    // Connects time-of-day lighting with spell damage modifiers
	timeOfDayAttackSpeedSystem                  *engine.TimeOfDayAttackSpeedSystem    // Connects time-of-day lighting with attack speed modifiers
	spriteGenerator                             *sprites.Generator
	spriteCache                                 *cache.SpriteCache   // Phase 1.2: Sprite caching for animation performance
	memoryMonitor                               *cache.MemoryMonitor // Background cache eviction under memory pressure
	itemGen                                     *item.ItemGenerator
	recipeGen                                   *recipe.RecipeGenerator
	statusEffectRNG                             *rand.Rand
	// V4.0 Systems (Phase 21-27)
	vehicleMovementSys           *engine.VehicleMovementSystem
	vehicleDurabilitySys         *engine.VehicleDurabilitySystem
	mountingSystem               *engine.MountingSystem
	vehicleCombatSystem          *engine.VehicleCombatSystem
	companionAISystem            *engine.CompanionAISystem
	companionProgressionSys      *engine.CompanionProgressionSystem
	companionQuestSynergySys     *engine.CompanionQuestSynergySystem // Companion skill-quest integration bonuses
	companionLoyaltySys          *engine.CompanionLoyaltySystem
	companionInventorySys        *engine.CompanionInventorySystem
	companionLearningSys         *engine.CompanionLearningSystem // Phase 4.1: Companion AI skill progression and personality evolution
	advancedClassSystem          *engine.AdvancedClassSystem     // Phase 4.2: Multi-classing, prestige classes, talent trees
	skillInheritanceSys          *engine.SkillInheritanceSystem
	bookReadingSystem            *engine.BookReadingSystem
	spellEffectSystem            *engine.SpellEffectSystem
	spellCombinationSys          *engine.SpellCombinationSystem
	classProgressionSys          *engine.ClassProgressionSystem
	dualClassSynergySys          *engine.DualClassSynergySystem          // Connects dual-class combinations with passive stat bonuses
	specializationManaBoostSys   *engine.SpecializationManaBoostSystem   // Connects class specialization with mana regen bonuses
	specializationHealthRegenSys *engine.SpecializationHealthRegenSystem // Connects class specialization with health regen bonuses
	specializationSpellDamageSys *engine.SpecializationSpellDamageSystem // Connects class specialization with spell damage bonuses
	specializationAttackSpeedSys *engine.SpecializationAttackSpeedSystem // Connects class specialization with attack speed bonuses
	specializationDefenseSys     *engine.SpecializationDefenseSystem     // Connects class specialization with defense bonuses
	specializationCritDamageSys  *engine.SpecializationCritDamageSystem  // Connects class specialization with crit damage bonuses
	specializationEvasionSys     *engine.SpecializationEvasionSystem     // Connects class specialization with evasion bonuses
	specializationLifestealSys   *engine.SpecializationLifestealSystem   // Connects class specialization with lifesteal bonuses
	expressionSystem             *engine.ExpressionSystem
	expressionComboSys           *engine.ExpressionComboSystem
	miniGameSystem               *engine.MiniGameSystem
	minigameGamesSystem          *games.System // Phase 3.4: Minigame implementations
	achievementSystem            *engine.AchievementSystem
	// Phase 28: Reputation & Moral Choices
	moralChoiceSystem *engine.MoralChoiceSystem
	// Phase 95-96: Resource gathering and fishing minigames
	fishingSystem                       *engine.FishingSystem
	gatheringSystem                     *engine.GatheringSystem
	fishingWeatherBonusSystem           *engine.FishingWeatherBonusSystem
	timeOfDayFishingBonusSystem         *engine.TimeOfDayFishingBonusSystem
	fishingCatchParticleSystem          *engine.FishingCatchParticleSystem
	fishingLineTensionParticleSystem    *engine.FishingLineTensionParticleSystem
	timeOfDayFishingBonusParticleSystem *engine.TimeOfDayFishingBonusParticleSystem
	terrainFishingBonusSystem           *engine.TerrainFishingBonusSystem   // Connects terrain tiles to fishing catch rate bonuses
	companionFishingBonusSystem         *engine.CompanionFishingBonusSystem // Connects companion types to fishing catch rate bonuses
	// Phase 112: New Game Plus carry-over system
	carryoverSystem *engine.CarryOverSystem
	// Phase 30: Environmental Storytelling
	discoverySystem *engine.DiscoverySystem
	// V5.0 Systems (Social & Communication)
	chatSystem    *engine.EnhancedChatSystem // Phase 3.1: Enhanced chat with E2E encryption and history persistence
	mailSystem    *engine.MailSystem
	courierSystem *engine.CourierSystem
	// V6.0 Systems (Persistent Worlds & Federation)
	portalSystem       *federation.PortalSystem
	bountySystem       *engine.BountySystem
	politicsSystem     *engine.PoliticsSystem
	territoryManager   *territory.Manager
	rankingManager     *world.RankingManager
	eventManager       *world.EventManager
	federationProtocol *federation.FederationProtocol

	// INTEGRATION FIX [Category A]: Missing Systems (Phase 14+)
	// Gap: Systems implemented but never instantiated or registered in game loop
	// Fix: Added system fields for investigation, audio enhancements, quality, trade, terrain modification, merchant caravans, NPC dialog
	// Roadmap: ROADMAP_V4.md (Phase 14, 30-31) and ROADMAP_V5.md (Phase 32-36)
	investigationSystem    *engine.InvestigationSystem       // Phase 30: Environmental Storytelling - investigation mechanics
	musicTriggerSystem     *engine.MusicTriggerSystem        // Phase 14.4: Adaptive music context switching
	positionalAudioSystem  *engine.PositionalAudioSystem     // Phase 14.4: 3D positional audio with panning/occlusion
	reverbSystem           *engine.ReverbSystem              // Phase 14.4: Room-based reverb effects
	qualitySystem          *engine.QualitySystem             // Phase 14: Performance-based quality settings
	tradeSystem            *engine.TradeSystem               // Phase 33: Player-to-player trading with trust limits
	terrainConstructionSys *engine.TerrainConstructionSystem // Phase 35: Buildable walls from materials
	terrainModificationSys *engine.TerrainModificationSystem // Phase 35: Destructible terrain
	merchantCaravanSystem  *engine.MerchantCaravanSystem     // Phase 36: Traveling merchants between servers
	npcDialogSystem        *engine.NPCDialogSystem           // Phase 31: Markov-chain NPC conversations

	// INTEGRATION FIX [Category A]: High-Level Management Systems
	// Gap: Wrapper systems CompanionSystem, VehicleSystem, AdaptiveSoundtrackSystem implemented but never initialized
	// Fix: Added system fields for high-level companion/vehicle management and adaptive music
	// Roadmap: ROADMAP_V4.md Phase 21.2 (VehicleSystem), Phase 22.2 (CompanionSystem), Phase 29 (AdaptiveSoundtrackSystem)
	companionSystem          *engine.CompanionSystem          // Phase 22.2: High-level companion management
	vehicleSystem            *engine.VehicleSystem            // Phase 21.2: High-level vehicle management
	adaptiveSoundtrackSystem *engine.AdaptiveSoundtrackSystem // Phase 29: Dynamic music adaptation based on game state

	// INTEGRATION FIX [Category A]: V8.0 Systems (Phase 49-51)
	// Gap: V8.0 systems fully implemented but never instantiated or registered
	// Fix: Added system fields for housing, social persistence, physics, territory, building/furniture generation
	// Roadmap: ROADMAP_V8.md (Phase 49-51)
	housingManager     *housing.Manager               // Phase 49.1: Player housing with plot placement
	trustManager       *persistence.TrustManager      // Phase 49.2: Persistent trust scores cross-server
	reputationManager  *persistence.ReputationManager // Phase 49.2: Reputation tracking (trade, combat, social, quest)
	chatHistory        *persistence.ChatHistory       // Phase 49.3: Chat history with delta compression
	imageGallery       *persistence.ImageGallery      // Phase 49.4: Persistent image storage
	guildHallManager   *housing.GuildHallManager      // Phase 51.2: Guild hall construction system
	enhancedVehicleSys *vehicle.EnhancedVehicleSystem // Phase 50.3: Vehicle physics (suspension, weight, deformation)
	fluidSimulator     *fluids.Simulator              // Phase 50.4: Fluid dynamics simulation
	buoyancyCalculator *fluids.BuoyancyCalculator     // Phase 50.4: Buoyancy and swimming
	swimmingManager    *fluids.SwimmingManager        // Phase 50.4: Swimming mechanics
	floodingManager    *fluids.FloodingManager        // Phase 50.4: Flooding system
	buildingGenerator  *building.Generator            // Phase 51.1: Procedural building generation
	furnitureGenerator *furniture.Generator           // Phase 51.3: Furniture generation and placement

	// INTEGRATION FIX [Category A]: V7.0 Display & Viewport Systems (Phase 43-44)
	// Gap: V7.0 display management and viewport optimization implemented but never initialized
	// Fix: Added system fields for display manager and viewport optimizer
	// Roadmap: ROADMAP_V7.md (Phase 43-44)
	displayManager    *display.Manager          // Phase 43: Display resolution management (1920x1080 default)
	viewportOptimizer *engine.ViewportOptimizer // Phase 44: Enhanced viewport culling for larger resolutions

	// INTEGRATION FIX [Category A]: V9.0 Integration Managers (Phase 55)
	// Gap: V9.0 integration managers implemented but never initialized or connected to systems
	// Fix: Added manager fields for housing crafting, companion housing, and guild housing
	// Roadmap: ROADMAP_V9.md (Phase 55.1-55.3)
	stationManager      *housingcrafting.StationManager  // Phase 55.1: Crafting stations in player housing
	petHomeManager      *companionhousing.PetHomeManager // Phase 55.2: Companion housing and pet homes
	guildHousingManager *guildhousing.Manager            // Phase 55.3: Guild housing and communal spaces

	// Phase 3.2: Guild Federation (PLAN.md)
	guildSystem            *engine.GuildSystem            // Cross-server guild management and sync
	guildUI                *engine.GuildUI                // Guild UI for player interaction
	guildCombatBonusSystem *engine.GuildCombatBonusSystem // Guild proximity combat bonuses

	// Phase 4.3: Territory Control (PLAN.md)
	territorySystem *engine.TerritorySystem // Territory capture and guild warfare mechanics
	territoryUI     *engine.TerritoryUI     // Territory UI for viewing and managing territories

	// Phase 1.2: Destruction Physics (PLAN.md)
	destructionSystem *destruction.System // Building structural integrity and debris physics

	// Phase 1.2: Performance Monitoring System (PLAN.md)
	performanceSystem *engine.PerformanceMonitoringSystem // Real-time performance tracking and metrics

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	lightingAdapter  *engine.LightingAdapter  // Dynamic lighting with multiple light sources
	animationAdapter *engine.AnimationAdapter // Advanced animation with articulation and direction
	// Note: PostProcessorAdapter already exists in game.PostProcessor (see initializeCoreSystems)

	// Phase 2.3: UI & Shape Rendering (PLAN.md)
	uiGenerator      *ui.Generator       // Procedural UI element generation (menus, buttons, panels)
	shapeRenderer    *shapes.Generator   // Geometric shape rendering for sprites and UI
	patternGenerator *patterns.Generator // Texture pattern generation for tiles and materials

	// Phase 2.4: Rendering Optimization (PLAN.md)
	imagePool        *pool.ImagePool      // Image pool for memory efficiency
	parallelRenderer *parallel.WorkerPool // Parallel renderer for performance

	// Phase 3.2: Magic & Skills (PLAN.md)
	magicGenerator *magic.SpellGenerator      // Procedural spell and magic generation for loot and progression
	skillGenerator *skills.SkillTreeGenerator // Skill tree generation for class progression

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	puzzleGenerator   *puzzle.Generator     // Procedural puzzle generation for dungeon rooms
	minigameGenerator *minigame.Generator   // Minigame generation for taverns and social spaces
	classGenerator    *class.ClassGenerator // Class archetype generation for character creation

	// Phase 3.6: Narrative Integration (PLAN.md)
	narrativeGenerator *narrative.StoryArcGenerator // Procedural story arc generation for world narrative

	// Phase 4.1: Chat System (PLAN.md)
	networkChatSystem *chat.ChatSystem // Network-based chat system for multiplayer messaging

	// Phase 4.2: Trade System (PLAN.md)
	networkTradeSystem *trade.TradeSystem // Network-based trade system for multiplayer item trading

	// Phase 5.1: Quality-of-Life System (PLAN.md)
	qolManager *qol.Manager             // Unified QoL manager for auto-loot, crafting queues, guild invitations, etc.
	qolSystem  *engine.QoLSystemWrapper // QoL system wrapper for ECS integration

	// Phase 1.3: Prestige System (PLAN.md)
	prestigeSystem *prestige.System // Post-max-level progression with paragon points and prestige abilities

	// Phase 2.1: Economy System (PLAN.md)
	economySystem *engine.EconomySystem // Cross-server marketplace and guild bank management

	// Phase 2.2: Raid System (PLAN.md)
	raidSystem *engine.RaidSystem // Dynamic raid instance creation and lockout management

	// Phase 3.3: Legendary Quest System (PLAN.md)
	legendaryQuestSystem *engine.LegendaryQuestSystem // Legendary quest management

	// Phase 4.1: Choice & Consequences System (PLAN.md)
	choiceConsequencesSystem *engine.ChoiceConsequencesSystem // Persistent choice tracking and consequences

	// Phase 4.2: Guild Vehicle Integration (PLAN.md)
	guildVehicleSystem *engine.GuildVehicleSystem // Guild vehicle fleet combat with formation bonuses

	// Phase 4.3: Narrative-World Integration (PLAN.md)
	narrativeWorldSystem *narrative_world.System // Companion-driven narrative events and story progression

	// Phase 4.4: Political Warfare Integration (PLAN.md)
	politicalWarfareSystem *political_warfare.System // Guild politics, war declarations, and diplomatic mechanics

	// Phase 4.5: Territory Siege System (PLAN.md)
	siegeManager *territory.SiegeManager      // Territory siege manager (V8.0)
	siegeSystem  *engine.TerritorySiegeSystem // ECS wrapper for siege mechanics

	// Phase 4.4: Trade Routes System (PLAN.md Integration - Week 4)
	tradeRouteManager *trade_routes.RouteManager // AI merchant caravan trading system

	// Phase 5.1: Mobile Federation Support (PLAN.md)
	mobileFederationSystem *engine.MobileFederationSystem // Mobile-optimized federation

	// Phase 1.1: Companion Learning System (PLAN.md Integration - Week 1)
	companionLearningSystem *learning.CompanionLearningSystem // AI companion behavior adaptation and learning

	// V19.0: Priority 1 Dormant Package Integration (ROADMAP_V19.md)
	// Phase 99: Entity & Dialog Generation
	entityGenerator *entity.EntityGenerator // Procedural entity generation for monsters/NPCs
	dialogGenerator *dialog.MarkovGenerator // Markov-chain dialog generation for NPCs

	// Phase 100: Legendary & Economy
	legendaryManager   *legendary.QuestManager // Legendary quest tracking and generation
	worldEconomySystem *economy.System         // Federated marketplace and dynamic economy

	// Phase 101: Integration Package Activation
	choiceTracker     *choice_consequences.ChoiceTracker // Persistent choice tracking and consequences
	guildFleetManager *guild_vehicle.FleetManager        // Guild vehicle fleet management
	worldEventManager *world_events.EventManager         // World-responsive event generation

	// Voice Systems (PLAN.md Phase 1)
	// Gap: Voice systems implemented in pkg/engine but never initialized or registered
	// Fix: Added voice channel and spatial voice systems for multiplayer voice chat
	voiceChannelSystem *engine.VoiceChannelSystem // Voice channel lifecycle and participant synchronization
	spatialVoiceSystem *engine.SpatialVoiceSystem // Distance-based volume and stereo panning for voice

	// VR Systems (AUDIT.md Task 7)
	// Gap: VR systems implemented but never initialized with hardware detection
	// Fix: Added stereoscopic and head tracking systems with conditional initialization
	stereoscopicSystem *engine.StereoscopicSystem // VR stereoscopic rendering (dual-eye cameras)
	headTrackingSystem *engine.HeadTrackingSystem // VR head tracking with mouse fallback
	vrControllerSystem *engine.VRControllerSystem // VR controller input system
	vrUISystem         *engine.VRUISystem         // VR-optimized UI rendering

	// Time abstraction for deterministic testing of gameplay code
	timeProvider TimeProvider

	// Lazy initialization tracking (Performance Audit S4)
	lazyInitStarted   bool // Tracks whether lazy initialization has been triggered
	lazyInitCompleted bool // Tracks whether lazy initialization has finished
	lazyInitMutex     sync.Mutex
}

// initializeCoreSystems creates and initializes all core game systems.
func initializeCoreSystems(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) *systemsContainer {
	startTime := time.Now()
	clientLogger.Info("initializing game systems with parallel optimization")

	sys := &systemsContainer{
		timeProvider: DefaultTimeProvider(),
	}

	// Phase 1: Critical systems (must be sequential)
	sys.performanceSystem = engine.NewPerformanceMonitoringSystem()
	clientLogger.WithField("system_name", "performance").Debug("Created performance monitoring system")

	sys.inputSystem = engine.NewInputSystem()
	sys.movementSystem = engine.NewMovementSystem(playerMaxSpeed)
	sys.statisticsSystem = engine.NewStatisticsSystem(game.World)
	sys.collisionSystem = engine.NewCollisionSystem(collisionGridCellSize)
	sys.movementSystem.SetCollisionSystem(sys.collisionSystem)
	sys.movementSystem.SetStatisticsSystem(sys.statisticsSystem)

	// Phase 2: Parallel initialization of independent systems
	var wg sync.WaitGroup
	wg.Add(3)

	// Group 1: Combat & Interaction systems
	go func() {
		defer wg.Done()
		sys.combatSystem = engine.NewCombatSystemWithLogger(*seed, logger)
		sys.interactionSystem = engine.NewInteractionSystem(game.World)
		sys.particleSystem = engine.NewParticleSystem()
		clientLogger.Debug("combat & interaction systems initialized")
	}()

	// Group 2: Sprite & Animation systems
	// WASM runs in a browser with ~1-2GB total heap. Use reduced cache
	// and animation limits to prevent progressive memory exhaustion.
	effectiveSpriteCacheMax := int64(spriteCacheMaxSize)
	effectiveAnimCacheSize := animationCacheSize
	if mobile.IsWASM() {
		effectiveSpriteCacheMax = wasmSpriteCacheMaxSize
		effectiveAnimCacheSize = wasmAnimationCacheSize
	}
	go func() {
		defer wg.Done()
		sys.spriteCache = cache.NewSpriteCache(effectiveSpriteCacheMax)
		sys.spriteGenerator = sprites.NewGenerator()
		sys.animationSystem = engine.NewAnimationSystem(sys.spriteGenerator)
		sys.animationSystem.SetMaxCacheSize(effectiveAnimCacheSize)
		sys.animationSystem.SetSpriteCache(sys.spriteCache)
		sys.equipmentVisualSystem = engine.NewEquipmentVisualSystem(sys.spriteGenerator)
		clientLogger.WithField("maxSize", effectiveSpriteCacheMax).Debug("sprite & animation systems initialized")
	}()

	// Group 3: Rendering systems
	// WASM uses Low quality to stay within browser resource constraints;
	// desktop uses Medium with selective overrides.
	go func() {
		defer wg.Done()
		var qualityConfig *quality.Config
		if mobile.IsWASM() {
			wasmQuality := quality.LowQualityConfig()
			qualityConfig = &wasmQuality
		} else {
			qualityConfig = &quality.Config{
				Level:                 quality.QualityMedium,
				EnablePostProcessing:  true,
				EnableBloom:           false,
				EnableSoftShadows:     true,
				SpriteDetailLevel:     0.7,
				EnableAntiAliasing:    true,
				AntiAliasingQuality:   1,
				EnableSpriteCache:     true,
				EnableDynamicLighting: true,
				ShadowSampleCount:     2,
			}
		}
		sys.qualitySystem = engine.NewQualitySystem(qualityConfig, 60.0)
		sys.lightingAdapter = engine.NewLightingAdapter(clientLogger.WithField("system", "lighting"))
		sys.animationAdapter = engine.NewAnimationAdapter(sys.spriteGenerator, clientLogger.WithField("system", "animation"))
		sys.uiGenerator = ui.NewGeneratorWithLogger(logger)
		sys.shapeRenderer = shapes.NewGenerator()
		sys.patternGenerator = patterns.NewGeneratorWithLogger(logger)
		sys.imagePool = pool.NewImagePool()
		workerCount := runtime.NumCPU()
		sys.parallelRenderer = parallel.NewWorkerPool(workerCount)
		clientLogger.WithField("workerCount", workerCount).Debug("rendering systems initialized")
	}()

	wg.Wait()

	// Phase 3: Final connections (requires all systems initialized)
	poolAdapter := engine.NewImagePoolAdapter(sys.imagePool)
	parallelAdapter := engine.NewParallelRendererAdapter(sys.parallelRenderer)
	game.RenderSystem.SetPool(poolAdapter)
	game.RenderSystem.SetParallelRenderer(parallelAdapter)

	// Phase 3.1: WASM memory safety — start background memory monitor and
	// tune GC to prevent progressive slowdown in the browser.
	if mobile.IsWASM() {
		sys.memoryMonitor = cache.NewMemoryMonitor(sys.spriteCache)
		sys.memoryMonitor.SetLimits(wasmMemorySoftLimit, wasmMemoryHardLimit)
		sys.memoryMonitor.Start()
		debug.SetGCPercent(50)
		clientLogger.WithFields(logrus.Fields{
			"spriteCacheMB":      wasmSpriteCacheMaxSize / (1024 * 1024),
			"animationCacheSize": wasmAnimationCacheSize,
			"qualityLevel":       "Low",
			"gcPercent":          50,
		}).Info("WASM performance controls active")
	}

	elapsed := time.Since(startTime)
	clientLogger.WithField("duration_ms", elapsed.Milliseconds()).Info("core systems initialized with parallel optimization")

	return sys
}

// scheduleLazyInit triggers background initialization of non-critical systems after first frame.
// This defers audio, environmental, and advanced feature systems to improve startup time.
func (sys *systemsContainer) scheduleLazyInit(game *engine.EbitenGame, logger *logrus.Logger, clientLogger *logrus.Entry) {
	sys.lazyInitMutex.Lock()
	if sys.lazyInitStarted {
		sys.lazyInitMutex.Unlock()
		return
	}
	sys.lazyInitStarted = true
	sys.lazyInitMutex.Unlock()

	// Start background initialization after a small delay to ensure first frame renders
	go func() {
		// Wait for first frame to complete (approximately 16ms for 60fps target)
		time.Sleep(20 * time.Millisecond)

		startTime := time.Now()
		clientLogger.Debug("starting lazy initialization of non-critical systems")

		// Phase 1: Audio system (can be initialized in background)
		initializeAudioSystem(game, sys, clientLogger)

		// Phase 2: Environmental systems (parallel - weather, hazards, etc.)
		var wg sync.WaitGroup
		wg.Add(8)
		go func() {
			defer wg.Done()
			initializeEnvironmentalSystems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV4Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV5Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV6Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV7Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV8Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV9Systems(game, sys, clientLogger)
		}()
		go func() {
			defer wg.Done()
			initializeV19Systems(game, sys, clientLogger)
		}()
		wg.Wait()

		// Phase 3: Guild Federation and advanced systems
		initializePhase3Systems(game, sys, clientLogger)

		// AUDIT.md Task 7: VR hardware detection and conditional system initialization
		initializeVRSystems(game, sys, clientLogger)

		// Register non-critical systems
		registerNonCriticalSystems(game, sys)

		// Phase 4: Sprite cache warming (deferred to idle time)
		// Priority 4 optimization: Warm common sprites to improve cache hit rate
		warmCommonSprites(sys, seed, genreID, clientLogger)

		sys.lazyInitMutex.Lock()
		defer sys.lazyInitMutex.Unlock()
		sys.lazyInitCompleted = true

		elapsed := time.Since(startTime)
		clientLogger.WithField("duration_ms", elapsed.Milliseconds()).Info("lazy initialization completed")
	}()
}

// isLazyInitCompleted returns true if lazy initialization has finished.
func (sys *systemsContainer) isLazyInitCompleted() bool {
	sys.lazyInitMutex.Lock()
	defer sys.lazyInitMutex.Unlock()
	return sys.lazyInitCompleted
}

// warmCommonSprites pre-generates common sprites to improve cache hit rate.
// This function queues sprite generation for:
// - Player idle/walk animations (4 directions)
// - Basic enemy types (idle state)
// Expected impact: -10-20ms startup, improved first-frame responsiveness
func warmCommonSprites(sys *systemsContainer, seed *int64, genre *string, logger *logrus.Entry) {
	if sys.spriteCache == nil || sys.spriteGenerator == nil {
		logger.Warn("sprite cache or generator not available, skipping warming")
		return
	}

	startTime := time.Now()
	logger.Debug("starting sprite cache warming")

	pregen := cache.NewPreGenerator(sys.spriteCache)
	seedVal, genreVal := resolveSeedAndGenre(seed, genre)

	queuePlayerSprites(pregen, sys.spriteGenerator, seedVal, genreVal)
	queueEnemySprites(pregen, sys.spriteGenerator, seedVal, genreVal)

	launchAsyncGeneration(pregen, startTime, logger)
}

// resolveSeedAndGenre resolves the seed and genre values with defaults.
// If genre is "random", it will be resolved to a concrete genre using the seed.
func resolveSeedAndGenre(seed *int64, genre *string) (int64, string) {
	seedVal := int64(12345)
	if seed != nil {
		seedVal = *seed
	}
	genreVal := "fantasy"
	if genre != nil {
		genreVal = *genre
	}
	// Handle "random" genre by selecting a concrete genre using the seed
	if genreVal == "random" {
		selectedGenre := genrepkg.GetRandomTheme(seedVal)
		genreVal = selectedGenre.ID
	}
	return seedVal, genreVal
}

// queuePlayerSprites queues common player sprite animations into the pre-generator.
func queuePlayerSprites(pregen *cache.PreGenerator, gen *sprites.Generator, seedVal int64, genreVal string) {
	playerStates := []string{"idle", "walk"}
	directions := []string{"down", "up", "left", "right"}

	for _, state := range playerStates {
		for frame := 0; frame < 4; frame++ {
			for _, dir := range directions {
				capturedState := state
				capturedFrame := frame
				capturedDir := dir

				key := cache.GenerateKey(seedVal, capturedState+"_"+capturedDir, capturedFrame)
				pregen.Queue(key, func() (*ebiten.Image, error) {
					config := sprites.Config{
						Type:    sprites.SpriteEntity,
						Width:   64,
						Height:  64,
						Seed:    seedVal,
						GenreID: genreVal,
						Custom: map[string]interface{}{
							"entityType": "player",
							"state":      capturedState,
							"frame":      capturedFrame,
							"direction":  capturedDir,
						},
					}
					return gen.Generate(config)
				})
			}
		}
	}
}

// queueEnemySprites queues common enemy idle sprites into the pre-generator.
func queueEnemySprites(pregen *cache.PreGenerator, gen *sprites.Generator, seedVal int64, genreVal string) {
	enemyTypes := []string{"goblin", "skeleton", "orc", "slime"}
	directions := []string{"down", "up", "left", "right"}

	for i, enemyType := range enemyTypes {
		for _, dir := range directions {
			capturedEnemy := enemyType
			capturedDir := dir
			enemySeed := seedVal + int64(i+1)

			key := cache.GenerateKey(enemySeed, "idle_"+capturedDir, 0)
			pregen.Queue(key, func() (*ebiten.Image, error) {
				config := sprites.Config{
					Type:    sprites.SpriteEntity,
					Width:   64,
					Height:  64,
					Seed:    enemySeed,
					GenreID: genreVal,
					Custom: map[string]interface{}{
						"entityType": capturedEnemy,
						"state":      "idle",
						"frame":      0,
						"direction":  capturedDir,
					},
				}
				return gen.Generate(config)
			})
		}
	}
}

// launchAsyncGeneration starts asynchronous sprite generation with logging.
func launchAsyncGeneration(pregen *cache.PreGenerator, startTime time.Time, logger *logrus.Entry) {
	go func() {
		count := pregen.Generate()
		elapsed := time.Since(startTime)
		logger.WithFields(logrus.Fields{
			"sprites_generated": count,
			"duration_ms":       elapsed.Milliseconds(),
		}).Info("sprite cache warming completed")
	}()
}

// initializeGenerators creates item and recipe generators for loot drops.
func initializeGenerators(sys *systemsContainer) {
	sys.itemGen = item.NewItemGenerator()
	sys.recipeGen = recipe.NewRecipeGenerator()

	// Phase 3.2: Magic & Skills (PLAN.md)
	sys.magicGenerator = magic.NewSpellGenerator()
	sys.skillGenerator = skills.NewSkillTreeGenerator()

	// Phase 3.5: Puzzles, Minigames, Class Generation (PLAN.md)
	sys.puzzleGenerator = puzzle.NewGenerator()
	sys.minigameGenerator = minigame.NewGenerator()
	sys.classGenerator = class.NewClassGenerator()

	// Phase 3.6: Narrative Integration (PLAN.md)
	sys.narrativeGenerator = narrative.NewStoryArcGenerator()
}

// initializeObjectiveSystem creates and configures the objective tracker with callbacks.
func initializeObjectiveSystem(sys *systemsContainer, logger *logrus.Logger) {
	sys.objectiveTracker = engine.NewObjectiveTrackerSystem()

	sys.objectiveTracker.SetQuestCompleteCallback(func(entity *engine.Entity, qst *quest.Quest) {
		sys.objectiveTracker.AwardQuestRewards(entity, qst)
		if logger.GetLevel() >= logrus.InfoLevel {
			logging.ComponentLogger(logger, "quest").WithFields(logrus.Fields{
				"questName":   qst.Name,
				"xpReward":    qst.Reward.XP,
				"goldReward":  qst.Reward.Gold,
				"skillPoints": qst.Reward.SkillPoints,
			}).Info("quest completed")
		}
	})
}

// initializeProgressionSystems creates AI, progression, inventory, and commerce systems.
func initializeProgressionSystems(game *engine.EbitenGame, sys *systemsContainer, logger *logrus.Logger) {
	sys.aiSystem = engine.NewAISystem(game.World)
	sys.progressionSystem = engine.NewProgressionSystem(game.World)
	sys.inventorySystem = engine.NewInventorySystem(game.World)
	sys.commerceSystem = engine.NewCommerceSystemWithLogger(game.World, sys.inventorySystem, logger)
	sys.reputationPricingSystem = engine.NewReputationPricingSystem(game.World, *seed+5000)
	sys.dialogSystem = engine.NewDialogSystemWithLogger(game.World, logger)
	sys.craftingSystem = engine.NewCraftingSystem(game.World, sys.inventorySystem, sys.itemGen)

	// Phase 1.3: Prestige system
	sys.prestigeSystem = prestige.NewSystemWithLogger(logger)
	logging.ComponentLogger(logger, "prestige").Debug("Created prestige system")

	// Phase 2.1: Economy system
	// Note: serverID is "local" for client-side economy tracking. In multiplayer mode,
	// the actual economy operations are handled by the authoritative server, and this
	// local instance is used for client-side caching and UI state management only.
	serverID := "local"
	sys.economySystem = engine.NewEconomySystem(game.World, serverID)
	logging.ComponentLogger(logger, "economy").Debug("Created economy system")

	// Phase 2.2: Raid system
	sys.raidSystem = engine.NewRaidSystem(game.World, game.GetWorldSeed())
	logging.ComponentLogger(logger, "raids").Debug("Created raid system")

	// Phase 3.3: Legendary quest system (requires raid manager)
	raidManager := raids.NewManager(game.GetWorldSeed())
	sys.legendaryQuestSystem = engine.NewLegendaryQuestSystem(game.World, game.GetWorldSeed(), raidManager)
	logging.ComponentLogger(logger, "legendary").Debug("Created legendary quest system")

	logging.ComponentLogger(logger, "commerce").Info("commerce system initialized")
	logging.ComponentLogger(logger, "dialog").Info("dialog system initialized")
	logging.ComponentLogger(logger, "crafting").Info("crafting system initialized")
}

// initializeAudioSystem creates and initializes the audio system with full synthesis support.
func initializeAudioSystem(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	// Phase 1.1: Full audio integration - synthesis, music, SFX
	const sampleRate = audioSampleRate
	audioSeed := *seed

	// Create synthesis engine for low-level waveform generation
	synthEngine := synthesis.NewEngineWithSampleRate(sampleRate, audioSeed)

	// Create base audio manager
	audioManager := audio.NewManager(sampleRate, audioSeed)

	// Create adaptive music manager
	musicManager := music.NewAdaptiveMusicManager(sampleRate, audioSeed)
	audioManager.SetMusicManager(musicManager)

	// Create SFX variety manager
	sfxManager := sfx.NewVarietyManager(sampleRate, audioSeed)
	audioManager.SetSFXManager(sfxManager)

	clientLogger.WithFields(logrus.Fields{
		"sampleRate": sampleRate,
		"seed":       audioSeed,
	}).Info("audio system initialized with adaptive music and SFX variety")

	// Store reference (legacy compatibility - keep for existing code)
	sys.audioManager = engine.NewAudioManager(audioSampleRate, *seed)
	sys.audioManagerSystem = engine.NewAudioManagerSystem(sys.audioManager)
	game.SetAudioManager(sys.audioManager)
	sys.audioManager.EnableAdaptiveMusic(true)

	// Phase 1.1 (PLAN.md): Connect synthesis engine to audio manager system
	sys.audioManagerSystem.SetSynthesizer(synthEngine)

	// Initialize enhanced audio systems
	sys.musicTriggerSystem = engine.NewMusicTriggerSystem(game.World, sys.audioManager)
	sys.positionalAudioSystem = engine.NewPositionalAudioSystem(game.World)
	sys.reverbSystem = engine.NewReverbSystemWithLogger(game.World, *seed+seedOffsetReverb, clientLogger.Logger)

	// Voice Systems (PLAN.md Phase 1)
	// Voice channel system manages voice channel lifecycle and participant synchronization
	sys.voiceChannelSystem = engine.NewVoiceChannelSystem(game.World)
	// Spatial voice system calculates distance-based volume and stereo panning
	sys.spatialVoiceSystem = engine.NewSpatialVoiceSystem(game.World)
	clientLogger.Debug("voice systems initialized (voice channels, spatial audio)")

	if *verbose {
		clientLogger.Info("adaptive music composition enabled with motif system, music triggers, positional audio, and reverb")
	}

	if err := sys.audioManager.PlayMusic(*genreID, "exploration"); err != nil {
		logging.ComponentLogger(clientLogger.Logger, "audio").WithError(err).Warn("failed to start background music")
	}

	logging.ComponentLogger(clientLogger.Logger, "audio").Info("audio system initialized (synthesis engine, music and SFX generators, triggers, 3D audio, reverb)")
}

// initializeCombatSystems creates spell casting and player combat systems.
func initializeCombatSystems(game *engine.EbitenGame, sys *systemsContainer) {
	sys.itemPickupSystem = engine.NewItemPickupSystem(game.World)

	sys.statusEffectRNG = rand.New(rand.NewSource(*seed + seedOffsetStatusEffect))
	sys.statusEffectSystem = engine.NewStatusEffectSystem(game.World, sys.statusEffectRNG)
	sys.spellCastingSystem = engine.NewSpellCastingSystem(game.World, sys.statusEffectSystem)
	sys.playerSpellCasting = engine.NewPlayerSpellCastingSystem(sys.spellCastingSystem, game.World)
	sys.manaRegenSystem = &engine.ManaRegenSystem{}
	sys.playerCombatSystem = engine.NewPlayerCombatSystem(sys.combatSystem, game.World)
	sys.playerItemUseSystem = engine.NewPlayerItemUseSystem(sys.inventorySystem, game.World)
}

// initializeEnvironmentalSystems creates weather, lifetime, puzzle, and destruction systems.
func initializeEnvironmentalSystems(game *engine.EbitenGame, sys *systemsContainer, clientLogger *logrus.Entry) {
	sys.weatherSystem = engine.NewWeatherSystem(game.World)
	sys.weatherCombatSystem = engine.NewWeatherCombatSystem(game.World)
	sys.statusEffectLightingSystem = engine.NewStatusEffectLightingSystem(game.World, *seed+2000)
	sys.statusEffectMovementSystem = engine.NewStatusEffectMovementSystem(game.World, *seed+2100)

	// StatusEffectEvasionSystem - applies evasion modifiers from status effects
	sys.statusEffectEvasionSystem = engine.NewStatusEffectEvasionSystem(game.World, *seed+2120)
	sys.statusEffectEvasionSystem.SetGenre(*genreID)

	// StatusEffectCriticalChanceSystem - applies crit chance modifiers from status effects
	sys.statusEffectCritChanceSystem = engine.NewStatusEffectCriticalChanceSystem(game.World, *seed+2130)
	sys.statusEffectCritChanceSystem.SetGenre(*genreID)

	// StatusEffectManaCostSystem - applies mana cost modifiers from status effects
	sys.statusEffectManaCostSystem = engine.NewStatusEffectManaCostSystem(game.World, *seed+7000)
	sys.statusEffectManaCostSystem.SetGenre(*genreID)

	// StatusEffectDamageBoostSystem - applies damage modifiers from status effects
	sys.statusEffectDamageBoostSystem = engine.NewStatusEffectDamageBoostSystem(game.World, *seed+2140)
	sys.statusEffectDamageBoostSystem.SetGenre(*genreID)

	sys.criticalHitParticleSystem = engine.NewCriticalHitParticleSystem(game.World, *seed+3000)
	sys.criticalHitParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.criticalHitParticleSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.SetCriticalHitCallback(sys.criticalHitParticleSystem.OnCriticalHit)
	}

	// LevelUpParticleSystem - visual feedback for level-ups
	sys.levelUpParticleSystem = engine.NewLevelUpParticleSystem(game.World, *seed+4000)
	sys.levelUpParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.levelUpParticleSystem.SetGenre(*genreID)
	if sys.progressionSystem != nil {
		sys.progressionSystem.AddLevelUpCallback(sys.levelUpParticleSystem.OnLevelUp)
	}

	// ItemPickupParticleSystem - visual feedback for item pickups
	sys.itemPickupParticleSystem = engine.NewItemPickupParticleSystem(game.World, *seed+4500)
	sys.itemPickupParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.itemPickupParticleSystem.SetGenre(*genreID)
	if sys.itemPickupSystem != nil {
		sys.itemPickupSystem.SetPickupCallback(sys.itemPickupParticleSystem.OnItemPickup)
	}

	// SpellEffectParticleSystem - visual feedback for spell effects
	sys.spellEffectParticleSystem = engine.NewSpellEffectParticleSystem(game.World, *seed+5500)
	sys.spellEffectParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.spellEffectParticleSystem.SetGenre(*genreID)

	// DeathParticleSystem - visual feedback for entity deaths
	sys.deathParticleSystem = engine.NewDeathParticleSystem(game.World, *seed+5600)
	sys.deathParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.deathParticleSystem.SetGenre(*genreID)

	// DamageResistanceParticleSystem - visual feedback for damage resistance
	sys.damageResistanceParticleSystem = engine.NewDamageResistanceParticleSystem(game.World, *seed+5700)
	sys.damageResistanceParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.damageResistanceParticleSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.SetDamageResistedCallback(sys.damageResistanceParticleSystem.OnDamageResisted)
	}

	// ShieldAbsorbParticleSystem - visual feedback for shield absorption
	sys.shieldAbsorbParticleSystem = engine.NewShieldAbsorbParticleSystem(game.World, *seed+5750)
	sys.shieldAbsorbParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.shieldAbsorbParticleSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.SetShieldAbsorbCallback(sys.shieldAbsorbParticleSystem.OnShieldAbsorb)
	}

	// ShieldRegenSystem - passive shield regeneration with visual feedback
	// Connects ShieldComponent with ParticleSystem for genre-aware shield recovery particles
	sys.shieldRegenSystem = engine.NewShieldRegenSystem(game.World, *seed+5775)
	sys.shieldRegenSystem.SetParticleSystem(sys.particleSystem)
	sys.shieldRegenSystem.SetGenre(*genreID)

	// DrowningParticleSystem - visual feedback for drowning entities
	// Connects SwimmingComponent.Drowning state with ParticleSystem for genre-aware suffocation effects
	sys.drowningParticleSystem = engine.NewDrowningParticleSystem(game.World, *seed+5780)
	sys.drowningParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.drowningParticleSystem.SetGenre(*genreID)

	// WeatherGroundEffectSystem - visual feedback for weather ground impacts
	sys.weatherGroundEffectSystem = engine.NewWeatherGroundEffectSystem(game.World, *seed+6000)
	sys.weatherGroundEffectSystem.SetParticleSystem(sys.particleSystem)
	sys.weatherGroundEffectSystem.SetGenre(*genreID)

	// WeatherAudioSystem - audio feedback for weather conditions
	sys.weatherAudioSystem = engine.NewWeatherAudioSystem(game.World, *seed+6100)
	sys.weatherAudioSystem.SetAudioManager(sys.audioManager)
	sys.weatherAudioSystem.SetGenre(*genreID)

	// WeatherManaRegenSystem - modifies mana regeneration based on weather
	sys.weatherManaRegenSystem = engine.NewWeatherManaRegenSystem(game.World, *seed+6300)
	sys.weatherManaRegenSystem.SetGenre(*genreID)

	// WeatherCooldownSystem - modifies spell cooldown rates based on weather
	sys.weatherCooldownSystem = engine.NewWeatherCooldownSystem(game.World, *seed+6350)
	sys.weatherCooldownSystem.SetGenre(*genreID)

	// WeatherAttackSpeedSystem - modifies melee attack cooldowns based on weather
	// Cold slows attacks, storms enable quick strikes
	sys.weatherAttackSpeedSystem = engine.NewWeatherAttackSpeedSystem(game.World, *seed+7425)
	sys.weatherAttackSpeedSystem.SetGenre(*genreID)

	// WeatherCompanionBonusSystem - modifies companion stats based on weather conditions
	// Connects WeatherSystem with CompanionStatsComponent for tactical companion positioning
	sys.weatherCompanionBonusSystem = engine.NewWeatherCompanionBonusSystem(game.World, *seed+6375)
	sys.weatherCompanionBonusSystem.SetGenre(*genreID)

	// LowHealthVFXSystem - visual feedback for low player health
	sys.lowHealthVFXSystem = engine.NewLowHealthVFXSystem(game.World, *seed+6400)
	sys.lowHealthVFXSystem.SetParticleSystem(sys.particleSystem)
	sys.lowHealthVFXSystem.SetGenre(*genreID)

	// ManaRegenParticleSystem - visual feedback for mana regeneration
	sys.manaRegenParticleSystem = engine.NewManaRegenParticleSystem(game.World, *seed+6450)
	sys.manaRegenParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.manaRegenParticleSystem.SetGenre(*genreID)

	// CompanionAuraParticleSystem - visual feedback for companion bonding perks
	sys.companionAuraParticleSystem = engine.NewCompanionAuraParticleSystem(game.World, *seed+6500)
	sys.companionAuraParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.companionAuraParticleSystem.SetGenre(*genreID)

	// ElementalComboParticleSystem - visual feedback for elemental status combos
	sys.elementalComboParticleSystem = engine.NewElementalComboParticleSystem(game.World, *seed+6700)
	sys.elementalComboParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.elementalComboParticleSystem.SetGenre(*genreID)

	// ElementalComboDamageSystem - applies bonus damage when elemental status effects combine
	sys.elementalComboDamageSystem = engine.NewElementalComboDamageSystem(game.World, *seed+6850)
	sys.elementalComboDamageSystem.SetGenre(*genreID)

	// WeatherElementalComboBonusSystem - modifies elemental combo damage based on weather
	// Rain enhances steam/water combos, storms boost electric combos, snow enhances ice combos
	sys.weatherElementalComboBonusSystem = engine.NewWeatherElementalComboBonusSystem(game.World, *seed+6855)
	sys.weatherElementalComboBonusSystem.SetElementalComboDamageSystem(sys.elementalComboDamageSystem)
	sys.weatherElementalComboBonusSystem.SetGenre(*genreID)

	// ElementalCompanionSynergySystem - boosts elemental companions based on owner status effects
	sys.elementalCompanionSynergySystem = engine.NewElementalCompanionSynergySystem(game.World, *seed+6900)
	sys.elementalCompanionSynergySystem.SetGenre(*genreID)

	// CompanionSpellAmplificationSystem - boosts owner spell damage/healing based on companion bonding
	sys.companionSpellAmplificationSystem = engine.NewCompanionSpellAmplificationSystem(game.World, *seed+6910)
	sys.companionSpellAmplificationSystem.SetGenre(*genreID)

	// CompanionManaRegenSystem - boosts owner mana regeneration based on companion bonding
	// Magical companions (Spirit, Elemental) provide stronger mana bonuses
	sys.companionManaRegenSystem = engine.NewCompanionManaRegenSystem(game.World, *seed+6920)
	sys.companionManaRegenSystem.SetGenre(*genreID)

	// WeatherRangedAccuracySystem - modifies ranged attack accuracy based on weather
	sys.weatherRangedAccuracySystem = engine.NewWeatherRangedAccuracySystem(game.World, *seed+6750)
	sys.weatherRangedAccuracySystem.SetGenre(*genreID)

	// WeatherCritChanceSystem - modifies critical hit chance based on weather conditions
	// Fog and dust increase crit (concealment), rain/snow decrease crit (precision penalty)
	sys.weatherCritChanceSystem = engine.NewWeatherCritChanceSystem(game.World, *seed+6775)
	sys.weatherCritChanceSystem.SetGenre(*genreID)

	// WeatherXPBonusSystem: bridges weather conditions with XP gains
	// Grants bonus XP when players earn experience during specific weather conditions
	sys.weatherXPBonusSystem = engine.NewWeatherXPBonusSystem(game.World, *seed+6800)
	sys.weatherXPBonusSystem.SetGenre(*genreID)
	if sys.progressionSystem != nil {
		sys.weatherXPBonusSystem.SetProgressionSystem(sys.progressionSystem)
	}

	// LifestealSystem - heals attackers based on damage dealt and their Lifesteal stat
	// Connects CombatSystem damage events with HealthComponent healing for sustained combat
	sys.lifestealSystem = engine.NewLifestealSystem(game.World, *seed+6950)
	sys.lifestealSystem.SetParticleSystem(sys.particleSystem)
	sys.lifestealSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.SetDamageCallback(sys.lifestealSystem.OnDamageDealt)
	}

	// StatusEffectDamageParticleSystem - visual feedback for status effect damage ticks
	// Connects StatusEffectSystem tick events with ParticleSystem for burn/poison/regen particles
	sys.statusEffectDamageParticleSystem = engine.NewStatusEffectDamageParticleSystem(game.World, *seed+7050)
	sys.statusEffectDamageParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.statusEffectDamageParticleSystem.SetGenre(*genreID)
	if sys.statusEffectSystem != nil {
		sys.statusEffectSystem.SetTickCallback(sys.statusEffectDamageParticleSystem.OnStatusEffectTick)
	}

	// FearFleeParticleSystem - visual feedback for feared entities fleeing
	// Connects StatusEffectAISystem flee behavior with ParticleSystem for genre-aware fear trails
	sys.fearFleeParticleSystem = engine.NewFearFleeParticleSystem(game.World, *seed+7150)
	sys.fearFleeParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.fearFleeParticleSystem.SetGenre(*genreID)

	sys.lifetimeSystem = engine.NewLifetimeSystemWithLogger(game.World, clientLogger.Logger)
	sys.puzzleSystem = engine.NewPuzzleSystem(game.World)

	sys.firePropagationSystem = engine.NewFirePropagationSystemWithLogger(tileSize, *seed+seedOffsetFirePropagation, clientLogger.Logger)
	sys.firePropagationSystem.SetWorld(game.World)

	sys.destructibleSystem = engine.NewDestructibleObjectSystemWithLogger(tileSize, *seed+seedOffsetDestructible, clientLogger.Logger)
	sys.destructibleSystem.SetWorld(game.World)
	sys.destructibleSystem.SetFireSystem(sys.firePropagationSystem)

	sys.carrySystem = engine.NewCarrySystemWithLogger(clientLogger.Logger)
	sys.carrySystem.SetWorld(game.World)

	sys.hazardSystem = engine.NewHazardSystemWithLogger(clientLogger.Logger)
	sys.hazardSystem.SetWorld(game.World)

	// HazardProximityWarningParticleSystem - visual warnings near hazard zones
	sys.hazardProximityWarningParticleSystem = engine.NewHazardProximityWarningParticleSystem(game.World, *seed+7500)
	sys.hazardProximityWarningParticleSystem.SetHazardSystem(sys.hazardSystem)
	sys.hazardProximityWarningParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.hazardProximityWarningParticleSystem.SetGenre(*genreID)

	sys.narrativeSystem = engine.NewNarrativeSystem(game.World)
	sys.branchingNarrativeSystem = engine.NewBranchingNarrativeSystem(game.World)                                               // Phase 6.1: Branching narratives
	sys.worldEventsSystem = engine.NewWorldEventsSystemWithLogger(game.World, *seed+seedOffsetWorldEvents, clientLogger.Logger) // Phase 6.3: World events
	sys.shadowSystem = engine.NewShadowSystemWithLogger(game.World, clientLogger.Logger)

	// Phase 1.2: Destruction physics
	destructionConfig := destruction.Config{
		EnableIntegrityChecks: true,
		DamagePropagationRate: 0.1,
		CollapseThreshold:     0.3,
		EnableDebris:          true,
		MaxDebrisParticles:    500,
		DebrisLifetime:        10.0,
		Gravity:               980.0,
		AirResistance:         0.05,
		GroundFriction:        0.8,
		MaxFallingObjects:     100,
		UpdateFrequency:       30.0,
		Seed:                  *seed + seedOffsetDestructionPhysics,
	}
	sys.destructionSystem = destruction.NewSystem(&destructionConfig)
	clientLogger.Info("destruction physics system initialized")
}

// registerCriticalSystems registers only systems required for first frame rendering.
// This includes core gameplay, input, movement, collision, combat, and rendering systems.
func registerCriticalSystems(game *engine.EbitenGame, sys *systemsContainer) {
	// Phase 1.2: Performance monitoring - register first to track all systems
	game.World.AddSystem(sys.performanceSystem)

	// Core input and camera systems
	game.World.AddSystem(sys.inputSystem)
	game.World.AddSystem(game.CameraSystem)

	// Essential gameplay systems for first frame
	sys.rotationSystem = engine.NewRotationSystem(game.World)
	game.World.AddSystem(&rotationSystemWrapper{system: sys.rotationSystem})

	game.World.AddSystem(sys.playerCombatSystem)
	game.World.AddSystem(sys.playerItemUseSystem)
	game.World.AddSystem(sys.playerSpellCasting)
	game.World.AddSystem(sys.movementSystem)
	game.World.AddSystem(sys.collisionSystem)
	game.World.AddSystem(sys.statisticsSystem) // Track player movement and exploration statistics

	sys.projectileSystem = engine.NewProjectileSystem(game.World)
	game.World.AddSystem(sys.projectileSystem)

	game.World.AddSystem(sys.combatSystem)
	game.World.AddSystem(sys.statusEffectSystem)

	sys.revivalSystem = engine.NewRevivalSystem(game.World)
	game.World.AddSystem(sys.revivalSystem)

	game.World.AddSystem(sys.aiSystem)

	sys.behaviorTreeSystem = engine.NewBehaviorTreeSystem(game.World)
	game.World.AddSystem(sys.behaviorTreeSystem)

	sys.squadSystem = engine.NewSquadSystem(game.World)
	game.World.AddSystem(&squadSystemWrapper{system: sys.squadSystem})

	game.World.AddSystem(sys.progressionSystem)

	// Phase 1.3: Prestige levels and resets
	game.World.AddSystem(&prestigeSystemWrapper{system: sys.prestigeSystem})

	// Essential game mechanics
	game.World.AddSystem(sys.objectiveTracker)
	game.World.AddSystem(sys.itemPickupSystem)
	game.World.AddSystem(sys.spellCastingSystem)
	game.World.AddSystem(sys.manaRegenSystem)
	game.World.AddSystem(sys.inventorySystem)
	game.World.AddSystem(sys.commerceSystem)
	game.World.AddSystem(sys.reputationPricingSystem) // Adjusts merchant prices based on faction reputation
	game.World.AddSystem(sys.dialogSystem)
	game.World.AddSystem(sys.craftingSystem)
	game.World.AddSystem(sys.interactionSystem)

	// Core rendering systems
	sys.creatureVisualClassifierSystem = engine.NewCreatureVisualClassifierSystem(game.World)
	game.World.AddSystem(sys.creatureVisualClassifierSystem)

	// Size sprite scaling: adapts anatomy proportions based on entity size class
	sys.sizeSpriteScalingSystem = engine.NewSizeSpriteScalingSystem(game.World)
	game.World.AddSystem(sys.sizeSpriteScalingSystem)

	// NPC role visual: infers humanoid visual roles (mage, warrior, merchant, etc.)
	sys.npcRoleVisualSystem = engine.NewNpcRoleVisualSystem(game.World, *seed+9050)
	game.World.AddSystem(sys.npcRoleVisualSystem)

	game.World.AddSystem(&animationSystemWrapper{
		system: sys.animationSystem,
		logger: game.World.GetLogger(),
	})

	game.World.AddSystem(sys.equipmentVisualSystem)

	// Directional sprites: 4-direction variants for correct facing in top-down view
	sys.directionalSpriteSystem = engine.NewDirectionalSpriteSystem(sys.spriteGenerator)
	sys.directionalSpriteSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.directionalSpriteSystem)

	// Sprite depth enhance: form-aware volumetric shading (spherical heads, cylindrical torsos)
	sys.spriteDepthEnhanceSystem = engine.NewSpriteDepthEnhanceSystem(game.World, *seed+9050)
	sys.spriteDepthEnhanceSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.spriteDepthEnhanceSystem)

	// Sprite color temperature: warm/cool grading and specular highlights for genre-aware lighting
	sys.spriteColorTemperatureSystem = engine.NewSpriteColorTemperatureSystem(game.World, *seed+9075)
	sys.spriteColorTemperatureSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.spriteColorTemperatureSystem)

	// Sprite finalizer: adaptive outline, rim lighting, edge shadow for all entity sprites
	sys.spriteFinalizerSystem = engine.NewSpriteFinalizerSystem(game.World, *seed+9100)
	game.World.AddSystem(sys.spriteFinalizerSystem)

	game.World.AddSystem(sys.particleSystem)

	// Phase 2.2: Core Rendering Systems (PLAN.md)
	game.World.AddSystem(sys.lightingAdapter)  // Dynamic lighting with multiple light sources
	game.World.AddSystem(sys.animationAdapter) // Advanced animation features

	// Phase 5.1 (PLAN.md): Quality-of-Life System - Essential for first frame UX
	game.World.AddSystem(sys.qolSystem)
}

// registerNonCriticalSystems registers systems not required for first frame.
// These include environmental effects, audio, advanced features, and federation systems.
func registerNonCriticalSystems(game *engine.EbitenGame, sys *systemsContainer) {
	// Audio system (deferred)
	game.World.AddSystem(sys.audioManagerSystem)

	// Faction and reputation systems
	sys.factionSystem = engine.NewFactionSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(sys.factionSystem)

	// FactionAwareAISystem: bridges faction reputation with AI hostility
	// Makes neutral NPCs become enemies when player reputation drops to hostile (-50 or below)
	sys.factionAwareAISystem = engine.NewFactionAwareAISystem(game.World)
	game.World.AddSystem(sys.factionAwareAISystem)

	// FactionXPBonusSystem: bridges faction reputation with XP bonus rewards
	// Awards bonus XP when player kills enemies of factions they have good standing with
	sys.factionXPBonusSystem = engine.NewFactionXPBonusSystem(game.World, *seed+5100)
	sys.factionXPBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.factionXPBonusSystem.SetProgressionSystem(sys.progressionSystem)
	game.World.AddSystem(sys.factionXPBonusSystem)

	// FactionDamageBonusSystem: bridges faction reputation with damage bonuses
	// Deals bonus damage when player attacks enemies of factions they have good standing with
	sys.factionDamageBonusSystem = engine.NewFactionDamageBonusSystem(game.World, *seed+5150)
	sys.factionDamageBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.factionDamageBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.factionDamageBonusSystem)

	// WeatherFactionResistanceSystem: bridges weather with faction elemental affinities
	// Fire factions weakened by rain, ice factions buffed by snow, etc.
	sys.weatherFactionResistanceSystem = engine.NewWeatherFactionResistanceSystem(game.World, *seed+5155)
	sys.weatherFactionResistanceSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weatherFactionResistanceSystem)

	// ReputationDefenseBonusSystem: bridges faction reputation with defense bonuses
	// Reduces incoming damage when player is attacked by enemies of allied factions
	sys.reputationDefenseBonusSystem = engine.NewReputationDefenseBonusSystem(game.World, *seed+5175)
	sys.reputationDefenseBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationDefenseBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationDefenseBonusSystem)

	// ReputationDefenseBonusParticleSystem: visual feedback for reputation defense bonuses
	// Spawns genre-aware particles when faction reputation reduces incoming damage
	sys.reputationDefenseBonusParticleSystem = engine.NewReputationDefenseBonusParticleSystem(game.World, *seed+5180)
	sys.reputationDefenseBonusParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationDefenseBonusParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationDefenseBonusParticleSystem)

	// ReputationHealingBonusSystem: bridges faction reputation with health regeneration
	// Players with high faction standing passively regenerate health faster
	sys.reputationHealingBonusSystem = engine.NewReputationHealingBonusSystem(game.World, *seed+5185)
	sys.reputationHealingBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationHealingBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationHealingBonusSystem)

	// ReputationHealingBonusParticleSystem: visual feedback for reputation healing
	// Spawns genre-aware upward healing particles when faction reputation heals the player
	sys.reputationHealingBonusParticleSystem = engine.NewReputationHealingBonusParticleSystem(game.World, *seed+5190)
	sys.reputationHealingBonusParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationHealingBonusParticleSystem.SetHealingSystem(sys.reputationHealingBonusSystem)
	sys.reputationHealingBonusParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationHealingBonusParticleSystem)

	// ReputationSpellDamageBonusSystem: bridges faction reputation with spell damage
	// Players with high faction standing gain a passive MagicPower bonus
	sys.reputationSpellDamageBonusSystem = engine.NewReputationSpellDamageBonusSystem(game.World, *seed+5195)
	sys.reputationSpellDamageBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationSpellDamageBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationSpellDamageBonusSystem)

	// ReputationSpellDamageBonusParticleSystem: visual feedback for reputation spell damage
	// Spawns genre-aware arcane particles when faction reputation boosts spell damage
	sys.reputationSpellDamageBonusParticleSystem = engine.NewReputationSpellDamageBonusParticleSystem(game.World, *seed+5200)
	sys.reputationSpellDamageBonusParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationSpellDamageBonusParticleSystem.SetSpellDamageSystem(sys.reputationSpellDamageBonusSystem)
	sys.reputationSpellDamageBonusParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationSpellDamageBonusParticleSystem)

	// ReputationMovementSpeedSystem: bridges faction reputation with movement speed
	// Players with high faction standing gain a passive movement speed bonus
	sys.reputationMovementSpeedSystem = engine.NewReputationMovementSpeedSystem(game.World, *seed+5205)
	sys.reputationMovementSpeedSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationMovementSpeedSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationMovementSpeedSystem)

	// ReputationMovementSpeedParticleSystem: visual feedback for reputation speed bonus
	// Spawns genre-aware wind-trail particles when faction reputation boosts movement speed
	sys.reputationMovementSpeedParticleSystem = engine.NewReputationMovementSpeedParticleSystem(game.World, *seed+5210)
	sys.reputationMovementSpeedParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationMovementSpeedParticleSystem.SetSpeedSystem(sys.reputationMovementSpeedSystem)
	sys.reputationMovementSpeedParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationMovementSpeedParticleSystem)

	// ReputationCriticalChanceBonusSystem: bridges faction reputation with critical hit chance
	// Players with high faction standing gain a passive crit chance bonus
	sys.reputationCriticalChanceBonusSystem = engine.NewReputationCriticalChanceBonusSystem(game.World, *seed+5215)
	sys.reputationCriticalChanceBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationCriticalChanceBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationCriticalChanceBonusSystem)

	// ReputationCriticalChanceParticleSystem: visual feedback for reputation crit bonus
	// Spawns genre-aware glint particles when faction reputation boosts critical chance
	sys.reputationCriticalChanceParticleSystem = engine.NewReputationCriticalChanceParticleSystem(game.World, *seed+5220)
	sys.reputationCriticalChanceParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationCriticalChanceParticleSystem.SetCritSystem(sys.reputationCriticalChanceBonusSystem)
	sys.reputationCriticalChanceParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationCriticalChanceParticleSystem)

	// ReputationEquipmentDurabilitySystem: bridges faction reputation with equipment durability
	// High reputation provides durability protection, hostile reputation accelerates degradation
	sys.reputationEquipmentDurabilitySystem = engine.NewReputationEquipmentDurabilitySystem(game.World, *seed+5225)
	sys.reputationEquipmentDurabilitySystem.SetFactionSystem(sys.factionSystem)
	sys.reputationEquipmentDurabilitySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationEquipmentDurabilitySystem)

	// ReputationEquipmentDurabilityParticleSystem: visual feedback for reputation durability
	// Spawns protective shimmer or corrosion wisps based on reputation modifier
	sys.reputationEquipmentDurabilityParticleSystem = engine.NewReputationEquipmentDurabilityParticleSystem(game.World, *seed+5230)
	sys.reputationEquipmentDurabilityParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationEquipmentDurabilityParticleSystem.SetDurabilitySystem(sys.reputationEquipmentDurabilitySystem)
	sys.reputationEquipmentDurabilityParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationEquipmentDurabilityParticleSystem)

	// ReputationCompanionBonusSystem: bridges faction reputation with companion stat bonuses
	// High reputation passively boosts companion attack, defense, and speed
	sys.reputationCompanionBonusSystem = engine.NewReputationCompanionBonusSystem(game.World, *seed+5235)
	sys.reputationCompanionBonusSystem.SetFactionSystem(sys.factionSystem)
	sys.reputationCompanionBonusSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationCompanionBonusSystem)

	// ReputationCompanionBonusParticleSystem: visual feedback for reputation companion bonus
	// Spawns genre-aware aura particles around companions buffed by owner reputation
	sys.reputationCompanionBonusParticleSystem = engine.NewReputationCompanionBonusParticleSystem(game.World, *seed+5240)
	sys.reputationCompanionBonusParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.reputationCompanionBonusParticleSystem.SetBonusSystem(sys.reputationCompanionBonusSystem)
	sys.reputationCompanionBonusParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.reputationCompanionBonusParticleSystem)

	// ReputationQuestGatingSystem: gates quests behind faction reputation requirements
	// Integrates reputation standings with quest availability for meaningful progression
	sys.reputationQuestGatingSystem = engine.NewReputationQuestGatingSystem(game.World, game.World.GetLogger().Logger, *seed+5245)
	game.World.AddSystem(sys.reputationQuestGatingSystem)

	// FactionTerritoryInfluenceSystem: faction zones provide combat/progression bonuses
	// Applies damage/defense/XP modifiers based on player reputation in faction-controlled territories
	sys.factionTerritoryInfluenceSystem = engine.NewFactionTerritoryInfluenceSystem(game.World, sys.factionSystem)
	game.World.AddSystem(sys.factionTerritoryInfluenceSystem)

	// AmbientEnvironmentParticleSystem: atmospheric particles based on terrain type
	// Spawns genre-aware ambient effects (fireflies, mist, embers, dust motes) near entities
	sys.ambientEnvironmentParticleSystem = engine.NewAmbientEnvironmentParticleSystem(game.World, *seed+7350)
	sys.ambientEnvironmentParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.ambientEnvironmentParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.ambientEnvironmentParticleSystem)

	// EquipmentEnchantmentGlowParticleSystem: ambient glow particles for rare+ equipped items
	// Reads highest equipped rarity and spawns rarity-colored enchantment aura particles
	sys.equipmentEnchantmentGlowParticleSystem = engine.NewEquipmentEnchantmentGlowParticleSystem(game.World, *seed+7420)
	sys.equipmentEnchantmentGlowParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.equipmentEnchantmentGlowParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentEnchantmentGlowParticleSystem)

	// StatusEffectVisualOverlaySystem: color tints on entity sprites from active status effects
	// Reads StatusEffectComponent and writes tint to VisualFeedbackComponent for visual feedback
	sys.statusEffectVisualOverlaySystem = engine.NewStatusEffectVisualOverlaySystem(game.World)
	sys.statusEffectVisualOverlaySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.statusEffectVisualOverlaySystem)

	// WeatherSpriteTintSystem: weather-driven sprite color tints
	// Reads active weather and applies subtle multiplicative color tints to entity sprites
	sys.weatherSpriteTintSystem = engine.NewWeatherSpriteTintSystem(game.World, *seed+8800)
	sys.weatherSpriteTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weatherSpriteTintSystem)

	// EntityDropShadowSystem: genre-aware soft elliptical drop shadows beneath entities
	// Reads entity collider/sprite size and applies genre-tinted shadow parameters
	sys.entityDropShadowSystem = engine.NewEntityDropShadowSystem(game.World, *seed+8900)
	sys.entityDropShadowSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityDropShadowSystem)

	// TimeOfDayShadowDirectionSystem: directional shadow offset from simulated sun position
	// Connects TimeOfDayLightingSystem with DropShadowComponent for sun-arc shadow direction
	sys.timeOfDayShadowDirectionSystem = engine.NewTimeOfDayShadowDirectionSystem(game.World, *seed+8925)
	sys.timeOfDayShadowDirectionSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.timeOfDayShadowDirectionSystem)

	// EquipmentMaterialSheenSystem: material-based specular highlights on equipment
	// Bridges sprites.GetMaterialVisualProperties with per-entity visual state
	sys.equipmentMaterialSheenSystem = engine.NewEquipmentMaterialSheenSystem(game.World, *seed+8950)
	sys.equipmentMaterialSheenSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentMaterialSheenSystem)

	// EquipmentDamageStateTintSystem: aggregate equipment wear visual tinting
	// Bridges sprites.GetDamageVisualEffects with per-entity render state
	sys.equipmentDamageStateTintSystem = engine.NewEquipmentDamageStateTintSystem(game.World, *seed+8975)
	sys.equipmentDamageStateTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentDamageStateTintSystem)

	// CreatureGenreTintSystem: genre-aware creature/NPC sprite color tinting
	// Applies genre-specific tint presets to creature sprites for atmospheric cohesion
	sys.creatureGenreTintSystem = engine.NewCreatureGenreTintSystem(game.World, *seed+9000)
	sys.creatureGenreTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureGenreTintSystem)

	// CreatureSizeProportionSystem: size-based anatomy proportions for creature/NPC sprites
	// Adjusts head/torso/leg ratios and width scale based on creature size with genre modifiers
	sys.creatureSizeProportionSystem = engine.NewCreatureSizeProportionSystem(game.World, *seed+9001)
	sys.creatureSizeProportionSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureSizeProportionSystem)

	// CreatureAnatomySystem: assigns anatomy types (quadruped, arachnid, etc.) to creatures
	// Maps entity names/tags to appropriate nonhumanoid aerial sprite templates
	sys.creatureAnatomySystem = engine.NewCreatureAnatomySystem(game.World, *seed+9010)
	sys.creatureAnatomySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureAnatomySystem)

	// EquipmentRarityDetailSystem: rarity-based visual detail scaling for equipment
	// Scales shape complexity, color vibrancy, border sharpness, and material fidelity by rarity
	sys.equipmentRarityDetailSystem = engine.NewEquipmentRarityDetailSystem(game.World, *seed+9025)
	sys.equipmentRarityDetailSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentRarityDetailSystem)

	// NpcFacialDetailSystem: genre-aware NPC facial features (eyes, mouth, expression, head shape)
	sys.npcFacialDetailSystem = engine.NewNpcFacialDetailSystem(game.World, *seed+9050)
	sys.npcFacialDetailSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.npcFacialDetailSystem)

	// DynamicExpressionSystem: reactive facial expression updates from health, status effects, AI state
	sys.dynamicExpressionSystem = engine.NewDynamicExpressionSystem(game.World, *seed+10700)
	game.World.AddSystem(sys.dynamicExpressionSystem)

	// ProjectileTrailParticleSystem: genre-aware projectile trail particles
	sys.projectileTrailParticleSystem = engine.NewProjectileTrailParticleSystem(game.World, *seed+9075)
	sys.projectileTrailParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.projectileTrailParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.projectileTrailParticleSystem)

	// EntityIdleAmbientParticleSystem: genre-aware idle entity ambient particles
	sys.entityIdleAmbientParticleSystem = engine.NewEntityIdleAmbientParticleSystem(game.World, *seed+9100)
	sys.entityIdleAmbientParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.entityIdleAmbientParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityIdleAmbientParticleSystem)

	// WeaponMaterialImpactParticleSystem: material-aware melee impact particles
	sys.weaponMaterialImpactParticleSystem = engine.NewWeaponMaterialImpactParticleSystem(game.World, *seed+9150)
	sys.weaponMaterialImpactParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.weaponMaterialImpactParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weaponMaterialImpactParticleSystem)

	// DamageFlashTintSystem: triggers genre-aware flash tints when entities take damage
	// Monitors HealthComponent changes and writes to VisualFeedbackComponent
	sys.damageFlashTintSystem = engine.NewDamageFlashTintSystem(game.World)
	sys.damageFlashTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.damageFlashTintSystem)

	// SprintTrailParticleSystem: spawns genre-aware trail particles behind sprinting entities
	sys.sprintTrailParticleSystem = engine.NewSprintTrailParticleSystem(game.World, *seed+9200)
	sys.sprintTrailParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.sprintTrailParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.sprintTrailParticleSystem)

	// NearbyLightEntityTintSystem: tints entity sprites based on nearby light sources
	// Uses light color, intensity, and distance-based falloff with genre-aware ambient base
	sys.nearbyLightEntityTintSystem = engine.NewNearbyLightEntityTintSystem(game.World, *seed+9300)
	sys.nearbyLightEntityTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.nearbyLightEntityTintSystem)

	// WeatherEquipmentSheenSystem: modifies equipment sheen based on weather conditions
	sys.weatherEquipmentSheenSystem = engine.NewWeatherEquipmentSheenSystem(game.World, *seed+9400)
	sys.weatherEquipmentSheenSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weatherEquipmentSheenSystem)

	// CreatureEyeGlowSystem: genre-aware glowing eyes for hostile creatures and bosses
	sys.creatureEyeGlowSystem = engine.NewCreatureEyeGlowSystem(game.World, *seed+9450)
	sys.creatureEyeGlowSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureEyeGlowSystem)

	// CreatureEyePatternSystem: creature-type-specific eye patterns for nonhumanoids
	// Assigns 8-eye spider patterns, slit-pupil serpent eyes, compound insect eyes, etc.
	sys.creatureEyePatternSystem = engine.NewCreatureEyePatternSystem(game.World, *seed+9460)
	sys.creatureEyePatternSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureEyePatternSystem)

	// CreatureElementalAuraSystem: persistent elemental aura visuals for creatures
	sys.creatureElementalAuraSystem = engine.NewCreatureElementalAuraSystem(game.World, *seed+9475)
	sys.creatureElementalAuraSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.creatureElementalAuraSystem)

	// MeleeSwingArcSystem: genre-aware visual arc overlays during melee attacks
	sys.meleeSwingArcSystem = engine.NewMeleeSwingArcSystem(game.World, *seed+9500)
	sys.meleeSwingArcSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.meleeSwingArcSystem)

	// CombatReadyAuraSystem: genre-aware aura when AI entities enter hostile states
	sys.combatReadyAuraSystem = engine.NewCombatReadyAuraSystem(game.World, *seed+9550)
	sys.combatReadyAuraSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.combatReadyAuraSystem)

	// AIStateBubbleSystem: genre-aware floating state indicator bubbles above NPCs
	sys.aiStateBubbleSystem = engine.NewAIStateBubbleSystem(game.World, *seed+9600)
	sys.aiStateBubbleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.aiStateBubbleSystem)

	// TerrainReflectionTintSystem: terrain-driven entity sprite color tinting
	sys.terrainReflectionTintSystem = engine.NewTerrainReflectionTintSystem(game.World, *seed+9650)
	sys.terrainReflectionTintSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.terrainReflectionTintSystem)

	// MovementBobSystem: genre-aware walk-cycle vertical sprite bobbing
	sys.movementBobSystem = engine.NewMovementBobSystem(game.World, *seed+9700)
	sys.movementBobSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.movementBobSystem)

	// MovementLeanSystem: genre-aware horizontal lean in movement direction
	sys.movementLeanSystem = engine.NewMovementLeanSystem(game.World, *seed+9710)
	sys.movementLeanSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.movementLeanSystem)

	// SpellCastGlowSystem: genre-aware visual glow during spell casting
	sys.spellCastGlowSystem = engine.NewSpellCastGlowSystem(game.World, *seed+9750)
	sys.spellCastGlowSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.spellCastGlowSystem)

	// EquipmentChangeFlashSystem: genre-aware flash particles on equipment change
	sys.equipmentChangeFlashSystem = engine.NewEquipmentChangeFlashSystem(game.World, *seed+9800)
	sys.equipmentChangeFlashSystem.SetParticleSystem(sys.particleSystem)
	sys.equipmentChangeFlashSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentChangeFlashSystem)

	// DodgeAfterimageSystem: genre-aware translucent ghost copies on dodge/dash
	sys.dodgeAfterimageSystem = engine.NewDodgeAfterimageSystem(game.World, *seed+9850)
	sys.dodgeAfterimageSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.dodgeAfterimageSystem)

	// EntityIdleBreathingSystem: genre-aware subtle idle breathing animation
	sys.entityIdleBreathingSystem = engine.NewEntityIdleBreathingSystem(game.World, *seed+9900)
	sys.entityIdleBreathingSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityIdleBreathingSystem)

	// CombatHitStaggerSystem: genre-aware positional stagger on damage
	sys.combatHitStaggerSystem = engine.NewCombatHitStaggerSystem(game.World, *seed+9950)
	sys.combatHitStaggerSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.combatHitStaggerSystem)

	// FloatingDamageNumberSystem: genre-aware floating damage/heal numbers on health change
	sys.floatingDamageNumberSystem = engine.NewFloatingDamageNumberSystem(game.World, *seed+10000)
	sys.floatingDamageNumberSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.floatingDamageNumberSystem)

	// EntityThreatIndicatorSystem: genre-aware threat level rings under AI entities
	sys.entityThreatIndicatorSystem = engine.NewEntityThreatIndicatorSystem(game.World, *seed+10050)
	sys.entityThreatIndicatorSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityThreatIndicatorSystem)

	// WeaponMaterialParticleSystem: genre-aware idle particles from equipped weapon material
	// Reads main-hand weapon material type and spawns material-appropriate ambient particles
	sys.weaponMaterialParticleSystem = engine.NewWeaponMaterialParticleSystem(game.World, *seed+10100)
	sys.weaponMaterialParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.weaponMaterialParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weaponMaterialParticleSystem)

	// LootRarityBeamSystem: genre-aware beam particles on ground items by rarity
	// Items with Uncommon+ rarity emit colored upward-flowing beam particles for loot visibility
	sys.lootRarityBeamSystem = engine.NewLootRarityBeamSystem(game.World, *seed+10200)
	sys.lootRarityBeamSystem.SetParticleSystem(sys.particleSystem)
	sys.lootRarityBeamSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.lootRarityBeamSystem)

	// DamageTypeColorFlashSystem: genre-aware elemental damage color flashes on hit
	// Reads attacker DamageType and tints the target sprite with per-element colors
	sys.damageTypeColorFlashSystem = engine.NewDamageTypeColorFlashSystem(game.World, *seed+10250)
	sys.damageTypeColorFlashSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.AddDamageCallback(sys.damageTypeColorFlashSystem.OnDamageDealt)
	}
	game.World.AddSystem(sys.damageTypeColorFlashSystem)

	// CompanionBondTetherSystem: genre-aware visual tether between companion and owner
	// Draws a pulsing line connecting companion to player with loyalty-driven intensity
	sys.companionBondTetherSystem = engine.NewCompanionBondTetherSystem(game.World, *seed+10300)
	sys.companionBondTetherSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.companionBondTetherSystem)

	// CriticalHitScreenShakeSystem: genre-aware camera shake on critical hits
	// Connects CombatSystem critical hit callback with CameraSystem.ShakeAdvanced
	sys.criticalHitScreenShakeSystem = engine.NewCriticalHitScreenShakeSystem(game.World, *seed+10350)
	sys.criticalHitScreenShakeSystem.SetCameraSystem(game.CameraSystem)
	sys.criticalHitScreenShakeSystem.SetGenre(*genreID)
	if sys.combatSystem != nil {
		sys.combatSystem.AddCriticalHitCallback(sys.criticalHitScreenShakeSystem.OnCriticalHit)
	}
	game.World.AddSystem(sys.criticalHitScreenShakeSystem)

	// EntitySpawnMaterializeSystem: genre-aware visual fade-in when entities spawn
	sys.entitySpawnMaterializeSystem = engine.NewEntitySpawnMaterializeSystem(game.World, *seed+10400)
	sys.entitySpawnMaterializeSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entitySpawnMaterializeSystem)

	// NPCInteractionProximityGlowSystem: genre-aware glow tint on interactable NPCs near players
	sys.npcInteractionProximityGlowSystem = engine.NewNPCInteractionProximityGlowSystem(game.World, *seed+10450)
	sys.npcInteractionProximityGlowSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.npcInteractionProximityGlowSystem)

	// EntityFactionOutlineSystem: genre-aware colored outlines based on team/faction allegiance
	sys.entityFactionOutlineSystem = engine.NewEntityFactionOutlineSystem(game.World, *seed+10500)
	sys.entityFactionOutlineSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityFactionOutlineSystem)

	// ShieldBubbleOverlaySystem: genre-aware translucent bubble around shielded entities
	sys.shieldBubbleOverlaySystem = engine.NewShieldBubbleOverlaySystem(game.World, *seed+10550)
	sys.shieldBubbleOverlaySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.shieldBubbleOverlaySystem)

	// EnvironmentalBreathVaporSystem: cold weather breath vapor puffs near entity faces
	sys.environmentalBreathVaporSystem = engine.NewEnvironmentalBreathVaporSystem(game.World, *seed+10600)
	sys.environmentalBreathVaporSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.environmentalBreathVaporSystem)

	// MovementDustSystem: speed-proportional terrain dust behind fast-moving entities
	sys.movementDustSystem = engine.NewMovementDustSystem(game.World, *seed+10650)
	sys.movementDustSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.movementDustSystem)

	// HealthRegenPulseSystem: genre-aware healing pulse particles on health increase
	sys.healthRegenPulseSystem = engine.NewHealthRegenPulseSystem(game.World, *seed+10675)
	sys.healthRegenPulseSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.healthRegenPulseSystem)

	// StatusEffectGroundTrailSystem: ground-level DoT movement trails
	// Drops genre-aware particles behind moving entities afflicted by burning/poison/bleeding/frozen
	sys.statusEffectGroundTrailSystem = engine.NewStatusEffectGroundTrailSystem(game.World, *seed+10690)
	sys.statusEffectGroundTrailSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.statusEffectGroundTrailSystem)

	// WaterSurfaceRippleSystem: ripple/splash particles when entities move through water tiles
	sys.waterSurfaceRippleSystem = engine.NewWaterSurfaceRippleSystem(game.World, *seed+10700)
	sys.waterSurfaceRippleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.waterSurfaceRippleSystem)

	// EntityTargetLockIndicatorSystem: genre-aware targeting reticle on closest hostile
	sys.entityTargetLockIndicatorSystem = engine.NewEntityTargetLockIndicatorSystem(game.World, *seed+10705)
	sys.entityTargetLockIndicatorSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityTargetLockIndicatorSystem)

	// EntityDeathDissolveSystem: genre-aware visual dissolve when entities die
	sys.entityDeathDissolveSystem = engine.NewEntityDeathDissolveSystem(game.World, *seed+10710)
	sys.entityDeathDissolveSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.entityDeathDissolveSystem)

	// ArmorHitSparkSystem: material-aware armor deflection sparks on damage
	sys.armorHitSparkSystem = engine.NewArmorHitSparkSystem(game.World, *seed+10715)
	sys.armorHitSparkSystem.SetParticleSystem(sys.particleSystem)
	sys.armorHitSparkSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.armorHitSparkSystem)

	// MeleeEnchantmentArcParticleSystem: rarity-colored particles along melee swing arcs
	sys.meleeEnchantmentArcParticleSystem = engine.NewMeleeEnchantmentArcParticleSystem(game.World, *seed+10720)
	sys.meleeEnchantmentArcParticleSystem.SetParticleSystem(sys.particleSystem)
	sys.meleeEnchantmentArcParticleSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.meleeEnchantmentArcParticleSystem)

	// BattleWoundOverlaySystem: genre-aware wound marks based on health percentage
	sys.battleWoundOverlaySystem = engine.NewBattleWoundOverlaySystem(game.World, *seed+10725)
	sys.battleWoundOverlaySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.battleWoundOverlaySystem)

	// EquipmentDamageCrackOverlaySystem: procedural crack patterns from equipment wear
	sys.equipmentDamageCrackOverlaySystem = engine.NewEquipmentDamageCrackOverlaySystem(game.World, *seed+10730)
	sys.equipmentDamageCrackOverlaySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentDamageCrackOverlaySystem)

	// WeatherEntityWetnessSystem: rain-driven sprite darkening and genre-aware sheen tint
	sys.weatherEntityWetnessSystem = engine.NewWeatherEntityWetnessSystem(game.World, *seed+10735)
	sys.weatherEntityWetnessSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.weatherEntityWetnessSystem)

	// AttackTelegraphGlowSystem: genre-aware attack wind-up warning glow on hostile AI
	sys.attackTelegraphGlowSystem = engine.NewAttackTelegraphGlowSystem(game.World, *seed+10740)
	sys.attackTelegraphGlowSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.attackTelegraphGlowSystem)

	// XPGainBurstSystem: genre-aware XP gain particle bursts on experience increase
	sys.xpGainBurstSystem = engine.NewXPGainBurstSystem(game.World, *seed+10745)
	sys.xpGainBurstSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.xpGainBurstSystem)

	// EquipmentGleamSweepSystem: animated specular gleam sweep across equipped items
	sys.equipmentGleamSweepSystem = engine.NewEquipmentGleamSweepSystem(game.World, *seed+10750)
	sys.equipmentGleamSweepSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.equipmentGleamSweepSystem)

	// SpriteDepthShadingSystem: per-body-part depth shading for entity sprites
	sys.spriteDepthShadingSystem = engine.NewSpriteDepthShadingSystem(game.World, *seed+10755)
	sys.spriteDepthShadingSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.spriteDepthShadingSystem)

	// ClothingPatternSystem: seed-based clothing patterns for entity sprites
	// Attaches genre-aware garment patterns (stripes, checks, dots, borders) for visual variety
	sys.clothingPatternSystem = engine.NewClothingPatternSystem(game.World, *seed+10760)
	sys.clothingPatternSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.clothingPatternSystem)

	// SurfaceTextureSystem: creature-form surface textures (fur, scales, chitin, metal, etc.)
	// Assigns form-appropriate procedural micro-textures to nonhumanoid entities
	sys.surfaceTextureSystem = engine.NewSurfaceTextureSystem(game.World, *seed+10765)
	sys.surfaceTextureSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.surfaceTextureSystem)

	// HumanoidTextureSystem: humanoid-specific surface textures
	// Assigns skin textures, clothing fabric textures, and hair textures to humanoid entities
	sys.humanoidTextureSystem = engine.NewHumanoidTextureSystem(game.World, *seed+10767)
	sys.humanoidTextureSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.humanoidTextureSystem)

	// BodyTypeSystem: seed-based body type variety for entity sprites
	// Assigns distinct body builds (stocky, lean, muscular, heavy, etc.) per entity
	sys.bodyTypeSystem = engine.NewBodyTypeSystem(game.World, *seed+10770)
	sys.bodyTypeSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.bodyTypeSystem)

	// HeadgearAssignmentSystem: seed-based headgear variety for humanoid entities
	// Assigns genre- and role-aware headgear types (crowns, hoods, wizard hats, etc.)
	sys.headgearAssignmentSystem = engine.NewHeadgearAssignmentSystem(game.World, *seed+10775)
	sys.headgearAssignmentSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.headgearAssignmentSystem)

	// BackAccessorySystem: seed-based back accessories for humanoid entities
	// Assigns genre- and role-aware back accessories (capes, cloaks, quivers, backpacks, etc.)
	sys.backAccessorySystem = engine.NewBackAccessorySystem(game.World, *seed+10780)
	sys.backAccessorySystem.SetGenre(*genreID)
	game.World.AddSystem(sys.backAccessorySystem)

	// FactionCompanionBehaviorSystem: bridges faction reputation with companion AI targeting
	// Companions won't attack allied faction members and prioritize hostile faction enemies
	sys.factionCompanionBehaviorSystem = engine.NewFactionCompanionBehaviorSystem(game.World, *seed+2198)
	sys.factionCompanionBehaviorSystem.SetFactionSystem(sys.factionSystem)
	sys.factionCompanionBehaviorSystem.SetGenre(*genreID)
	game.World.AddSystem(sys.factionCompanionBehaviorSystem)

	// StatusEffectAISystem: bridges status effects with AI behavior
	// Disables AI actions when entities are stunned, frozen, feared, or paralyzed
	sys.statusEffectAISystem = engine.NewStatusEffectAISystem(game.World, *seed+2200)
	game.World.AddSystem(sys.statusEffectAISystem)

	// StatusEffectEvasionSystem: bridges status effects with combat evasion
	// Modifies evasion chance based on status effects (frozen, haste, stunned, etc.)
	game.World.AddSystem(sys.statusEffectEvasionSystem)

	// StatusEffectCriticalChanceSystem: bridges status effects with critical hit chance
	// Modifies crit chance based on status effects (blessed, cursed, focused, etc.)
	game.World.AddSystem(sys.statusEffectCritChanceSystem)

	// StatusEffectManaCostSystem: bridges status effects with spell mana costs
	// Modifies mana costs based on status effects (haste, cursed, focused, etc.)
	game.World.AddSystem(sys.statusEffectManaCostSystem)

	// StatusEffectDamageBoostSystem: bridges status effects with combat damage
	// Modifies Attack/MagicPower based on status effects (enraged, empowered, weakness, etc.)
	game.World.AddSystem(sys.statusEffectDamageBoostSystem)

	sys.reputationSystem = engine.NewReputationSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(&reputationSystemWrapper{system: sys.reputationSystem})

	sys.alignmentSystem = engine.NewAlignmentSystem(game.World)
	game.World.AddSystem(&alignmentSystemWrapper{system: sys.alignmentSystem})

	sys.factionReactionSystem = engine.NewFactionReactionSystem(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(&factionReactionSystemWrapper{system: sys.factionReactionSystem})

	sys.skillProgressionSystem = engine.NewSkillProgressionSystem()
	game.World.AddSystem(sys.skillProgressionSystem)

	sys.skillLoadoutSystem = engine.NewSkillLoadoutSystem(game.World)
	game.World.AddSystem(sys.skillLoadoutSystem)

	// ClassAffinitySystem: tracks playstyle progression and combat archetypes
	// Players build affinity XP by using abilities, unlocking passive bonuses for their preferred playstyle
	sys.classAffinitySystem = engine.NewClassAffinitySystem(game.World)
	game.World.AddSystem(sys.classAffinitySystem)

	sys.attributeAllocationSystem = engine.NewAttributeAllocationSystem(game.World, *seed)
	game.World.AddSystem(sys.attributeAllocationSystem)

	sys.talentSystem = engine.NewTalentSystem(game.World)
	game.World.AddSystem(sys.talentSystem)

	// SkillMutationSystem: allows players to customize skills with mutation effects
	// Players collect mutations as loot and apply them for stat/effect modifications
	sys.skillMutationSystem = engine.NewSkillMutationSystemWithLogger(game.World, game.World.GetLogger().Logger)
	game.World.AddSystem(sys.skillMutationSystem)

	// BuildTemplateSystem: manages complete character build presets
	// Players can save and load full builds (attributes + talents + skills) as templates
	sys.buildTemplateSystem = engine.NewBuildTemplateSystem(game.World, *seed)
	game.World.AddSystem(sys.buildTemplateSystem)

	// EquipmentSetBonusSystem: tracks equipped set pieces and applies tiered bonuses
	// Players wearing multiple pieces from the same equipment set receive cumulative stat bonuses
	sys.equipmentSetBonusSystem = engine.NewEquipmentSetBonusSystem(*seed+seedOffsetDualClassSynergy+100, *genreID, game.World.GetLogger().Logger)
	game.World.AddSystem(sys.equipmentSetBonusSystem)

	sys.visualFeedbackSystem = engine.NewVisualFeedbackSystem()
	game.World.AddSystem(sys.visualFeedbackSystem)

	// Economy and advanced gameplay systems
	game.World.AddSystem(sys.economySystem)        // Phase 2.1: Dynamic economy
	game.World.AddSystem(sys.raidSystem)           // Phase 2.2: Dynamic raids
	game.World.AddSystem(sys.legendaryQuestSystem) // Phase 3.3: Legendary quests

	// Environmental systems (deferred)
	game.World.AddSystem(sys.weatherSystem)
	game.World.AddSystem(sys.weatherCombatSystem)
	game.World.AddSystem(sys.weatherGroundEffectSystem)         // Weather ground impact visual feedback via particles
	game.World.AddSystem(sys.weatherAudioSystem)                // Weather ambient audio feedback
	game.World.AddSystem(sys.weatherManaRegenSystem)            // Weather mana regeneration modifiers
	game.World.AddSystem(sys.weatherCooldownSystem)             // Weather spell cooldown rate modifiers
	game.World.AddSystem(sys.weatherAttackSpeedSystem)          // Weather melee attack speed modifiers
	game.World.AddSystem(sys.weatherCompanionBonusSystem)       // Weather companion combat stat bonuses
	game.World.AddSystem(sys.weatherRangedAccuracySystem)       // Weather ranged attack accuracy modifiers
	game.World.AddSystem(sys.weatherXPBonusSystem)              // Weather XP gain bonuses
	game.World.AddSystem(sys.statusEffectLightingSystem)        // Status effect visual feedback via lighting
	game.World.AddSystem(sys.statusEffectMovementSystem)        // Status effect movement speed modifiers
	game.World.AddSystem(sys.criticalHitParticleSystem)         // Critical hit visual feedback via particles
	game.World.AddSystem(sys.levelUpParticleSystem)             // Level-up visual feedback via particles
	game.World.AddSystem(sys.itemPickupParticleSystem)          // Item pickup visual feedback via particles
	game.World.AddSystem(sys.spellEffectParticleSystem)         // Spell effect visual feedback via particles
	game.World.AddSystem(sys.damageResistanceParticleSystem)    // Damage resistance visual feedback via particles
	game.World.AddSystem(sys.shieldAbsorbParticleSystem)        // Shield absorption visual feedback via particles
	game.World.AddSystem(sys.shieldRegenSystem)                 // Shield regeneration visual feedback via particles
	game.World.AddSystem(sys.lowHealthVFXSystem)                // Low player health warning visual feedback via particles
	game.World.AddSystem(sys.manaRegenParticleSystem)           // Mana regeneration visual feedback via particles
	game.World.AddSystem(sys.companionAuraParticleSystem)       // Companion bonding perk aura visual feedback via particles
	game.World.AddSystem(sys.elementalComboParticleSystem)      // Elemental status combo visual feedback via particles
	game.World.AddSystem(sys.elementalComboDamageSystem)        // Elemental status combo bonus damage
	game.World.AddSystem(sys.weatherElementalComboBonusSystem)  // Weather-based elemental combo damage modifiers
	game.World.AddSystem(sys.elementalCompanionSynergySystem)   // Elemental companion stat bonuses from owner effects
	game.World.AddSystem(sys.companionSpellAmplificationSystem) // Owner spell damage/healing bonuses from companion bonding
	game.World.AddSystem(sys.companionManaRegenSystem)          // Owner mana regeneration bonuses from companion bonding
	game.World.AddSystem(sys.lifestealSystem)                   // Combat lifesteal healing with visual feedback
	game.World.AddSystem(sys.statusEffectDamageParticleSystem)  // Status effect DOT tick visual feedback (burn/poison/regen)
	game.World.AddSystem(sys.fearFleeParticleSystem)            // Fear flee particle trails for genre-aware fear feedback
	game.World.AddSystem(sys.lifetimeSystem)
	game.World.AddSystem(sys.puzzleSystem)
	game.World.AddSystem(sys.firePropagationSystem)
	game.World.AddSystem(sys.destructibleSystem)
	game.World.AddSystem(&destructionSystemWrapper{system: sys.destructionSystem}) // Phase 1.2: Destruction physics
	game.World.AddSystem(sys.carrySystem)
	game.World.AddSystem(sys.hazardSystem)
	game.World.AddSystem(sys.hazardProximityWarningParticleSystem)
	game.World.AddSystem(sys.narrativeSystem)
	game.World.AddSystem(sys.branchingNarrativeSystem) // Phase 6.1: Branching narratives
	game.World.AddSystem(sys.worldEventsSystem)        // Phase 6.3: World events
	game.World.AddSystem(sys.shadowSystem)

	// V4.0 System Registrations (Phase 21-27)
	// Phase 21: Vehicle systems
	game.World.AddSystem(sys.vehicleMovementSys)
	game.World.AddSystem(sys.vehicleDurabilitySys)
	game.World.AddSystem(sys.mountingSystem)
	game.World.AddSystem(sys.vehicleCombatSystem)

	// Phase 22: Companion systems (use wrappers for incompatible signatures)
	game.World.AddSystem(&companionAISystemWrapper{system: sys.companionAISystem})
	game.World.AddSystem(&companionProgressionSystemWrapper{system: sys.companionProgressionSys})
	game.World.AddSystem(sys.companionQuestSynergySys) // Companion-quest synergy bonuses
	game.World.AddSystem(&companionLoyaltySystemWrapper{system: sys.companionLoyaltySys})
	game.World.AddSystem(&companionInventorySystemWrapper{system: sys.companionInventorySys})
	game.World.AddSystem(sys.companionLearningSys) // Phase 4.1: Companion learning (compatible signature)
	game.World.AddSystem(sys.advancedClassSystem)  // Phase 4.2: Advanced classes (compatible signature)
	game.World.AddSystem(&skillInheritanceSystemWrapper{system: sys.skillInheritanceSys})

	// Phase 23: Book system
	game.World.AddSystem(sys.bookReadingSystem)

	// Phase 24: Expanded magic systems
	game.World.AddSystem(sys.spellEffectSystem)
	game.World.AddSystem(sys.spellCombinationSys)

	// Phase 25: Class progression
	game.World.AddSystem(sys.classProgressionSys)

	// Phase 25.2a: Dual-class synergy - connects dual-class combinations with passive stat bonuses
	game.World.AddSystem(sys.dualClassSynergySys)

	// Phase 25a: Specialization mana boost - connects class specialization with mana regen bonuses
	game.World.AddSystem(sys.specializationManaBoostSys)

	// Phase 25b: Specialization health regen - connects class specialization with health regen bonuses
	game.World.AddSystem(sys.specializationHealthRegenSys)

	// Phase 25c: Specialization spell damage - connects class specialization with spell damage bonuses
	game.World.AddSystem(sys.specializationSpellDamageSys)

	// Phase 25d: Specialization attack speed - connects class specialization with attack cooldown bonuses
	game.World.AddSystem(sys.specializationAttackSpeedSys)

	// Phase 25e: Specialization defense - connects class specialization with defense bonuses
	game.World.AddSystem(sys.specializationDefenseSys)

	// Phase 25f: Specialization lifesteal - connects class specialization with lifesteal bonuses
	game.World.AddSystem(sys.specializationLifestealSys)

	// Phase 26: Expression systems (use wrappers)
	game.World.AddSystem(&expressionSystemWrapper{system: sys.expressionSystem})
	game.World.AddSystem(&expressionComboSystemWrapper{system: sys.expressionComboSys})

	// Phase 26.2: Achievement system (social features, use wrapper)
	game.World.AddSystem(&achievementSystemWrapper{system: sys.achievementSystem})

	// INTEGRATION FIX [Category A]: Phase 28 - MoralChoiceSystem registration
	game.World.AddSystem(&moralChoiceSystemWrapper{system: sys.moralChoiceSystem})

	// Phase 27: Mini-game system (use wrapper)
	game.World.AddSystem(&miniGameSystemWrapper{system: sys.miniGameSystem})

	// Phase 3.4: Minigame implementations
	game.World.AddSystem(sys.minigameGamesSystem)

	// AUDIT FIX: Phase 95-96 - Fishing and Gathering Systems
	game.World.AddSystem(sys.fishingSystem)
	game.World.AddSystem(sys.gatheringSystem)

	// AUDIT FIX: Phase 112 - CarryOverSystem for New Game Plus
	game.World.AddSystem(sys.carryoverSystem)

	// Phase 4.1: Choice & consequences
	game.World.AddSystem(sys.choiceConsequencesSystem) // Phase 4.1: Choice tracking and consequences
	game.World.AddSystem(sys.guildVehicleSystem)       // Phase 4.2: Guild vehicle fleets
	game.World.AddSystem(sys.narrativeWorldSystem)     // Phase 4.3: Narrative-world integration
	game.World.AddSystem(sys.politicalWarfareSystem)   // Phase 4.4: Political warfare
	game.World.AddSystem(sys.siegeSystem)              // Phase 4.5: Territory sieges
	game.World.AddSystem(sys.mobileFederationSystem)   // Phase 5.1: Mobile federation

	// Phase 30: Environmental Storytelling - Discovery System (use wrapper)
	game.World.AddSystem(&discoverySystemWrapper{system: sys.discoverySystem})

	// V5.0 System Registrations (Social & Communication)
	// Phase 32: Chat system for player communication
	game.World.AddSystem(&chatSystemWrapper{system: sys.chatSystem})

	// Phase 40: Mail and courier systems for asynchronous messaging
	game.World.AddSystem(&mailSystemWrapper{system: sys.mailSystem})
	game.World.AddSystem(&courierSystemWrapper{system: sys.courierSystem})

	// Phase 4.1-4.2 (PLAN.md): Network Chat and Trade Systems
	game.World.AddSystem(&networkChatSystemWrapper{system: sys.networkChatSystem})
	game.World.AddSystem(&networkTradeSystemWrapper{system: sys.networkTradeSystem})

	// Phase 3.2 (PLAN.md): Guild Federation - Cross-server guild management
	if sys.guildSystem != nil {
		game.World.AddSystem(sys.guildSystem)
	}

	// Phase 4.3 (PLAN.md): Territory Control - Guild warfare and territory capture
	if sys.territorySystem != nil {
		game.World.AddSystem(&territorySystemWrapper{system: sys.territorySystem})
	}

	// V6.0 System Registrations (Persistent Worlds & Federation)
	// Phase 39: Portal system for cross-server travel
	game.World.AddSystem(&portalSystemWrapper{system: sys.portalSystem})

	// Phase 40: Bounty system for cross-server quests
	game.World.AddSystem(&bountySystemWrapper{system: sys.bountySystem})

	// Phase 41: Politics system for server diplomacy
	game.World.AddSystem(&politicsSystemWrapper{system: sys.politicsSystem})

	// INTEGRATION FIX [Category A]: Missing System Registrations
	// Phase 30-31: Environmental Storytelling and NPC Dialog
	game.World.AddSystem(&investigationSystemWrapper{system: sys.investigationSystem})
	game.World.AddSystem(&npcDialogSystemWrapper{system: sys.npcDialogSystem})

	// Phase 14.4: Adaptive Audio Systems
	game.World.AddSystem(&musicTriggerSystemWrapper{system: sys.musicTriggerSystem})
	game.World.AddSystem(&positionalAudioSystemWrapper{system: sys.positionalAudioSystem})
	game.World.AddSystem(&reverbSystemWrapper{system: sys.reverbSystem})

	// Phase 14: Quality System
	game.World.AddSystem(&qualitySystemWrapper{system: sys.qualitySystem})

	// Voice Systems (PLAN.md Phase 1)
	game.World.AddSystem(sys.voiceChannelSystem)
	game.World.AddSystem(sys.spatialVoiceSystem)

	// Phase 33-36: Trade, Terrain, and Merchant Systems
	game.World.AddSystem(&tradeSystemWrapper{system: sys.tradeSystem})
	game.World.AddSystem(&terrainConstructionSystemWrapper{system: sys.terrainConstructionSys})
	game.World.AddSystem(&terrainModificationSystemWrapper{system: sys.terrainModificationSys})
	game.World.AddSystem(&merchantCaravanSystemWrapper{system: sys.merchantCaravanSystem})

	// INTEGRATION FIX [Category A]: High-Level Management System Registrations
	// Phase 21.2: High-level VehicleSystem wrapper
	game.World.AddSystem(&vehicleSystemWrapper{system: sys.vehicleSystem})

	// Phase 22.2: High-level CompanionSystem wrapper
	game.World.AddSystem(&companionSystemWrapper{system: sys.companionSystem})

	// Phase 1.1: Companion Learning System (PLAN.md Integration - Week 1)
	game.World.AddSystem(&companionLearningSystemWrapper{system: sys.companionLearningSystem})

	// Phase 29: AdaptiveSoundtrackSystem for dynamic music
	game.World.AddSystem(&adaptiveSoundtrackSystemWrapper{system: sys.adaptiveSoundtrackSystem})

	// INTEGRATION FIX [Category A]: V8.0 Fluid Simulator System Registration (Phase 50.4)
	if sys.fluidSimulator != nil {
		game.World.AddSystem(&fluidSimulatorWrapper{system: sys.fluidSimulator})
	}

	// V19.0: Priority 1 Dormant Package Integration (ROADMAP_V19.md)
	// Phase 100: World Economy System
	if sys.worldEconomySystem != nil {
		game.World.AddSystem(&worldEconomySystemWrapper{system: sys.worldEconomySystem})
	}

	// Phase 101: World Event Manager
	if sys.worldEventManager != nil {
		game.World.AddSystem(&worldEventManagerWrapper{system: sys.worldEventManager})
	}
}

// registerAllSystems adds all systems to the game world in the correct order.
// DEPRECATED: Use registerCriticalSystems + lazy initialization instead.
// Kept for backwards compatibility and non-lazy initialization paths.
func registerAllSystems(game *engine.EbitenGame, sys *systemsContainer) {
	registerCriticalSystems(game, sys)
	registerNonCriticalSystems(game, sys)
}

// configureSystemConnections wires up interdependent systems.
func configureSystemConnections(game *engine.EbitenGame, sys *systemsContainer) {
	sys.combatSystem.SetCamera(game.CameraSystem)
	sys.combatSystem.SetParticleSystem(sys.particleSystem, game.World, *genreID)
	sys.combatSystem.SetProjectileSystem(sys.projectileSystem)
	// Plan Phase 1.1: Connect combat system to audio manager for combat SFX
	sys.combatSystem.SetAudioManager(sys.audioManager)
	sys.projectileSystem.SetCamera(game.CameraSystem)
	sys.projectileSystem.SetGenre(*genreID)
	sys.projectileSystem.SetSeed(*seed)
	sys.interactionSystem.SetCarrySystem(sys.carrySystem)
}

// generateWorldTerrain creates and validates procedural terrain using async loading.
// Returns the async loader for tracking progress instead of blocking until complete.
func generateWorldTerrain(logger *logrus.Logger, clientLogger *logrus.Entry) *terrain.AsyncLoader {
	clientLogger.Info("starting async terrain generation")

	terrainGen := terrain.NewBSPGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  defaultTerrainWidth,
			"height": defaultTerrainHeight,
			"theme":  getGenreTheme(),
		},
	}

	// Use async loader to generate terrain in background
	loader := terrain.NewAsyncLoader(logger)
	loader.StartGeneration(terrainGen, *seed, params)

	clientLogger.Info("terrain generation started in background")
	return loader
}

// initializeTerrainRendering sets up the terrain rendering system.
func initializeTerrainRendering(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"transitions":    true,
			"parallax":       false,
			"enhanced_walls": true,
		}).Info("initializing terrain rendering system")
	}

	terrainRenderSystem := engine.NewTerrainRenderSystem(tileSize, tileSize, *genreID, *seed)
	terrainRenderSystem.SetTerrain(generatedTerrain)

	// Configure advanced tile rendering features (unconditionally enabled as of Phase 2.1)
	terrainRenderSystem.SetTransitionsEnabled(true)
	terrainRenderSystem.SetParallaxEnabled(false) // Disabled due to performance impact
	terrainRenderSystem.SetEnhancedWallsEnabled(true)

	game.TerrainRenderSystem = terrainRenderSystem

	if *verbose {
		clientLogger.Info("terrain rendering system initialized with advanced features")
	}
}

// configureLightingSystem enables and configures dynamic lighting (unconditionally enabled as of Phase 2.1).
func configureLightingSystem(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	clientLogger.Info("enabling dynamic lighting system")
	game.EnableLighting(true)
	game.SetLightingGenrePreset(*genreID)

	clientLogger.WithFields(logrus.Fields{
		"genre":     *genreID,
		"enabled":   true,
		"maxLights": maxLights,
	}).Info("lighting system configured")
}

// configurePostProcessing initializes and configures the post-processing system (unconditionally enabled as of Phase 2.1).
func configurePostProcessing(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	if game.PostProcessor == nil {
		clientLogger.Warn("PostProcessor not initialized")
		return
	}

	// Enable post-processing unconditionally (Phase 2.1)
	game.PostProcessor.SetEnabled(true)

	// Apply preset if specified
	if *postprocessPreset != "" {
		game.PostProcessor.SetGenrePreset(*postprocessPreset)
		clientLogger.WithFields(logrus.Fields{
			"preset": *postprocessPreset,
		}).Info("post-processing preset applied")
		return
	}

	// Otherwise, configure individual effects
	if *postprocessColorGrading {
		game.PostProcessor.EnableColorGrading(
			*postprocessSaturation,
			*postprocessContrast,
			*postprocessBrightness,
			0.0, // temperature (not exposed via flags yet)
			0.0, // tint (not exposed via flags yet)
		)
		clientLogger.WithFields(logrus.Fields{
			"saturation": *postprocessSaturation,
			"contrast":   *postprocessContrast,
			"brightness": *postprocessBrightness,
		}).Debug("color grading enabled")
	}

	if *postprocessVignette {
		game.PostProcessor.EnableVignette(
			*postprocessVignetteIntens,
			*postprocessVignetteSoft,
		)
		clientLogger.WithFields(logrus.Fields{
			"intensity": *postprocessVignetteIntens,
			"softness":  *postprocessVignetteSoft,
		}).Debug("vignette enabled")
	}

	if *postprocessChromaticAber {
		game.PostProcessor.EnableChromaticAberration(
			*postprocessChromaticIntens,
			1.0, // directionX (outward from center)
			0.0, // directionY
			3,   // samples
		)
		clientLogger.WithFields(logrus.Fields{
			"intensity": *postprocessChromaticIntens,
		}).Debug("chromatic aberration enabled")
	}

	clientLogger.Info("post-processing configured")
}

// configurePaletteOptions sets up genre palette options for sprite generation (Phase 5.4).
func configurePaletteOptions(sys *systemsContainer, clientLogger *logrus.Entry) {
	if sys.animationSystem == nil {
		clientLogger.Warn("AnimationSystem not initialized, skipping palette configuration")
		return
	}

	// Parse palette options from command-line flags
	opts, err := parsePaletteOptions()
	if err != nil {
		clientLogger.WithError(err).Warn("failed to parse palette options, using defaults")
		return
	}

	// Set palette options on animation system (used for sprite generation)
	sys.animationSystem.SetPaletteOptions(opts)

	clientLogger.WithFields(logrus.Fields{
		"harmony": paletteHarmony,
		"mood":    paletteMood,
		"rarity":  paletteRarity,
	}).Info("palette options configured")
}

// initializeTerrainCollision sets up efficient terrain collision checking.
func initializeTerrainCollision(game *engine.EbitenGame, sys *systemsContainer, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing terrain collision system")
	}

	terrainChecker := engine.NewTerrainCollisionChecker(tileSize, tileSize)
	terrainChecker.SetTerrain(generatedTerrain)

	for _, system := range game.World.GetSystems() {
		if collisionSys, ok := system.(*engine.CollisionSystem); ok {
			collisionSys.SetTerrainChecker(terrainChecker)
		}
		if projSys, ok := system.(*engine.ProjectileSystem); ok {
			projSys.SetTerrainChecker(terrainChecker)
		}
		if terrainMoveSys, ok := system.(*engine.TerrainMovementSpeedSystem); ok {
			terrainMoveSys.SetTerrain(generatedTerrain)
			sys.terrainMovementSpeedSystem = terrainMoveSys
		}
		if terrainCombatSys, ok := system.(*engine.TerrainCombatBonusSystem); ok {
			terrainCombatSys.SetTerrain(generatedTerrain)
			sys.terrainCombatBonusSystem = terrainCombatSys
		}
		if terrainCombatParticleSys, ok := system.(*engine.TerrainCombatBonusParticleSystem); ok {
			// Connect to terrain combat bonus system for bonus lookups
			if sys.terrainCombatBonusSystem != nil {
				terrainCombatParticleSys.SetTerrainCombatBonusSystem(sys.terrainCombatBonusSystem)
			}
			sys.terrainCombatBonusParticleSystem = terrainCombatParticleSys
		}
		if terrainStealthSys, ok := system.(*engine.TerrainStealthSystem); ok {
			terrainStealthSys.SetTerrain(generatedTerrain)
			sys.terrainStealthSystem = terrainStealthSys
		}
		if stealthIndicatorParticleSys, ok := system.(*engine.StealthIndicatorParticleSystem); ok {
			// Connect to terrain stealth system for stealth multiplier lookups
			if sys.terrainStealthSystem != nil {
				stealthIndicatorParticleSys.SetTerrainStealthSystem(sys.terrainStealthSystem)
			}
			sys.stealthIndicatorParticleSystem = stealthIndicatorParticleSys
		}
		if terrainAmbushCritSys, ok := system.(*engine.TerrainAmbushCritSystem); ok {
			// Connect to terrain stealth system for stealth multiplier lookups
			if sys.terrainStealthSystem != nil {
				terrainAmbushCritSys.SetTerrainStealthSystem(sys.terrainStealthSystem)
			}
			sys.terrainAmbushCritSystem = terrainAmbushCritSys
		}
		if terrainStatusSys, ok := system.(*engine.TerrainStatusEffectSystem); ok {
			terrainStatusSys.SetTerrain(generatedTerrain)
			sys.terrainStatusEffectSystem = terrainStatusSys
		}
		if terrainManaRegenSys, ok := system.(*engine.TerrainManaRegenSystem); ok {
			terrainManaRegenSys.SetTerrain(generatedTerrain)
			sys.terrainManaRegenSystem = terrainManaRegenSys
		}
		if terrainSpellDamageSys, ok := system.(*engine.TerrainSpellDamageSystem); ok {
			terrainSpellDamageSys.SetTerrain(generatedTerrain)
			sys.terrainSpellDamageSystem = terrainSpellDamageSys
		}
		if terrainEquipDurabilitySys, ok := system.(*engine.TerrainEquipmentDurabilitySystem); ok {
			terrainEquipDurabilitySys.SetTerrain(generatedTerrain)
			sys.terrainEquipmentDurabilitySys = terrainEquipDurabilitySys
		}
		if terrainRangedAccSys, ok := system.(*engine.TerrainRangedAccuracySystem); ok {
			terrainRangedAccSys.SetTerrain(generatedTerrain)
			sys.terrainRangedAccuracySys = terrainRangedAccSys
		}
		if terrainReflectionTintSys, ok := system.(*engine.TerrainReflectionTintSystem); ok {
			terrainReflectionTintSys.SetTerrain(generatedTerrain)
			sys.terrainReflectionTintSystem = terrainReflectionTintSys
		}
		if terrainCompanionBonusSys, ok := system.(*engine.TerrainCompanionBonusSystem); ok {
			terrainCompanionBonusSys.SetTerrain(generatedTerrain)
			sys.terrainCompanionBonusSystem = terrainCompanionBonusSys
		}
		if footstepParticleSys, ok := system.(*engine.FootstepParticleSystem); ok {
			footstepParticleSys.SetTerrain(generatedTerrain)
		}
		if ambientEnvSys, ok := system.(*engine.AmbientEnvironmentParticleSystem); ok {
			ambientEnvSys.SetTerrain(generatedTerrain)
		}
		if timeOfDayLightingSys, ok := system.(*engine.TimeOfDayLightingSystem); ok {
			sys.timeOfDayLightingSystem = timeOfDayLightingSys
		}
		if timeOfDayStealthSys, ok := system.(*engine.TimeOfDayStealthSystem); ok {
			sys.timeOfDayStealthSystem = timeOfDayStealthSys
		}
		if timeOfDayXPBonusSys, ok := system.(*engine.TimeOfDayXPBonusSystem); ok {
			sys.timeOfDayXPBonusSystem = timeOfDayXPBonusSys
		}
		if timeOfDayManaCostSys, ok := system.(*engine.TimeOfDayManaCostSystem); ok {
			sys.timeOfDayManaCostSystem = timeOfDayManaCostSys
		}
		if timeOfDayCritChanceSys, ok := system.(*engine.TimeOfDayCriticalChanceSystem); ok {
			sys.timeOfDayCriticalChanceSystem = timeOfDayCritChanceSys
		}
		if timeOfDayCompanionBonusSys, ok := system.(*engine.TimeOfDayCompanionBonusSystem); ok {
			sys.timeOfDayCompanionBonusSystem = timeOfDayCompanionBonusSys
		}
		if timeOfDayManaRegenSys, ok := system.(*engine.TimeOfDayManaRegenSystem); ok {
			sys.timeOfDayManaRegenSystem = timeOfDayManaRegenSys
		}
		if timeOfDayBlockChanceSys, ok := system.(*engine.TimeOfDayBlockChanceSystem); ok {
			sys.timeOfDayBlockChanceSystem = timeOfDayBlockChanceSys
		}
		if timeOfDayEvasionSys, ok := system.(*engine.TimeOfDayEvasionSystem); ok {
			sys.timeOfDayEvasionSystem = timeOfDayEvasionSys
		}
		if timeOfDayAttackSpeedSys, ok := system.(*engine.TimeOfDayAttackSpeedSystem); ok {
			sys.timeOfDayAttackSpeedSystem = timeOfDayAttackSpeedSys
		}
		if timeOfDayShadowDirSys, ok := system.(*engine.TimeOfDayShadowDirectionSystem); ok {
			sys.timeOfDayShadowDirectionSystem = timeOfDayShadowDirSys
		}
	}

	// Connect TimeOfDayFishingBonusSystem to extracted TimeOfDayLightingSystem
	if sys.timeOfDayFishingBonusSystem != nil && sys.timeOfDayLightingSystem != nil {
		sys.timeOfDayFishingBonusSystem.SetLightingSystem(sys.timeOfDayLightingSystem)
	}

	// Connect TimeOfDayShadowDirectionSystem to extracted TimeOfDayLightingSystem
	if sys.timeOfDayShadowDirectionSystem != nil && sys.timeOfDayLightingSystem != nil {
		sys.timeOfDayShadowDirectionSystem.SetLightingSystem(sys.timeOfDayLightingSystem)
	}

	if *verbose {
		clientLogger.Info("terrain collision system initialized (efficient mode)")
	}
}

// generateWorldFactions creates and registers all world factions.
func generateWorldFactions(game *engine.EbitenGame, params procgen.GenerationParams, clientLogger *logrus.Entry) {
	clientLogger.Info("generating world factions")

	factionGen := faction.NewGenerator()
	factionResult, err := factionGen.Generate(*seed+seedOffsetFaction, params)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to generate factions")
	}

	worldFactions := factionResult.([]*engine.Faction)
	clientLogger.WithFields(logrus.Fields{
		"count": len(worldFactions),
		"genre": *genreID,
	}).Info("factions generated")

	for _, fac := range worldFactions {
		for _, system := range game.World.GetSystems() {
			if facSys, ok := system.(*engine.FactionSystem); ok {
				facSys.AddFaction(fac)
				if *verbose {
					clientLogger.WithFields(logrus.Fields{
						"factionID":   fac.ID,
						"factionName": fac.Name,
						"factionType": fac.Type,
					}).Debug("faction added to world")
				}
			}
		}
	}
}

// initializeSpatialPartitioning creates and configures the spatial partition system.
// Also enables quadtree-based collision optimization for O(n log n) collision detection.
func initializeSpatialPartitioning(game *engine.EbitenGame, sys *systemsContainer, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("initializing spatial partition system for viewport culling and collision optimization")
	}

	worldWidth := float64(generatedTerrain.Width) * worldPixelsPerTile
	worldHeight := float64(generatedTerrain.Height) * worldPixelsPerTile

	spatialSystem := engine.NewSpatialPartitionSystem(worldWidth, worldHeight)
	game.World.AddSystem(spatialSystem)

	spatialSystem.Rebuild(game.World.GetEntities())

	game.RenderSystem.SetSpatialPartition(spatialSystem)
	game.RenderSystem.EnableCulling(true)
	game.RenderSystem.EnableBatching(true)

	// Enable quadtree-based collision optimization (29% faster than grid-based)
	// This reduces collision detection from O(n²) to O(n log n)
	if sys.collisionSystem != nil {
		sys.collisionSystem.SetQuadtree(spatialSystem.GetQuadtree())
		if *verbose {
			clientLogger.Info("quadtree-based collision optimization enabled")
		}
	}

	// Enable quadtree-based AI enemy detection (50-80% faster with 500+ entities)
	// This reduces enemy search from O(n) to O(log n)
	if sys.aiSystem != nil {
		sys.aiSystem.SetQuadtree(spatialSystem.GetQuadtree())
		if *verbose {
			clientLogger.Info("quadtree-based AI optimization enabled")
		}
	}

	clientLogger.WithFields(logrus.Fields{
		"worldWidth":         worldWidth,
		"worldHeight":        worldHeight,
		"cellSize":           quadtreeCapacity,
		"initialCount":       len(game.World.GetEntities()),
		"cullingActive":      true,
		"batchingActive":     true,
		"collisionOptimized": sys.collisionSystem != nil,
		"aiOptimized":        sys.aiSystem != nil,
	}).Info("spatial partition system initialized with initial rebuild, culling, batching, and collision optimization enabled")
}

// connectMapUIToTerrain wires the map UI to terrain data.
func connectMapUIToTerrain(game *engine.EbitenGame, generatedTerrain *terrain.Terrain, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.Info("connecting terrain to Map UI")
	}

	game.MapUI.SetTerrain(generatedTerrain)

	if *verbose {
		clientLogger.Info("Map UI configured with terrain data")
	}
}

// createPlayerEntity creates the player entity with all necessary components.
func createPlayerEntity(game *engine.EbitenGame, playerX, playerY float64, animationSystem *engine.AnimationSystem, clientLogger *logrus.Entry) *engine.Entity {
	if *verbose {
		clientLogger.Info("creating player entity")
	}

	player := game.World.CreateEntity()

	// Add player components
	player.AddComponent(&engine.PositionComponent{X: playerX, Y: playerY})
	player.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&engine.TeamComponent{TeamID: 1}) // Player team

	// Add input component for player control
	player.AddComponent(&engine.EbitenInput{})

	// Phase 10.1: Add rotation and aim components for 360° rotation and mouse aim
	player.AddComponent(engine.NewRotationComponent(0, 3.0)) // Start facing right
	player.AddComponent(engine.NewAimComponent(0))           // Start aiming right

	// Add animated sprite (Phase 45: 64×64 enhanced sprites)
	playerSprite := &engine.EbitenSprite{
		Image:   nil, // Will be created by animation system
		Width:   playerSpriteWidth,
		Height:  playerSpriteHeight,
		Visible: true,
		Layer:   10,
	}
	player.AddComponent(playerSprite)

	// Add animation component
	playerAnim := engine.NewAnimationComponent(*seed + int64(player.ID*1000))
	playerAnim.CurrentState = engine.AnimationStateIdle
	playerAnim.FrameTime = 0.15
	playerAnim.Loop = true
	playerAnim.Playing = true
	playerAnim.FrameCount = 8 // V7.0: 8-frame animations
	playerAnim.Dirty = true
	player.AddComponent(playerAnim)

	// Add equipment visual component
	player.AddComponent(engine.NewEquipmentVisualComponent())

	// Add camera components
	camera := engine.NewCameraComponent()
	camera.Smoothing = 0.1
	player.AddComponent(camera)

	player.AddComponent(engine.NewScreenShakeComponent())
	player.AddComponent(engine.NewHitStopComponent())

	// Add layer component
	layerComp := engine.NewLayerComponent()
	layerComp.CurrentLayer = 0
	player.AddComponent(&layerComp)

	// Add shadow component (Phase 45: shadow matches 64×64 sprite)
	playerShadow := engine.NewShadowComponent(playerSpriteWidth)
	playerShadow.CastsShadow = true
	playerShadow.ShadowType = engine.ShadowTypeSoft
	player.AddComponent(playerShadow)

	// Add player torch (unconditionally enabled as of Phase 2.1)
	playerTorch := engine.NewTorchLight(200)
	playerTorch.Enabled = true
	player.AddComponent(playerTorch)

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"radius":    200,
			"intensity": playerTorch.Intensity,
		}).Info("player torch added")
	}

	// Configure camera and animation systems
	game.CameraSystem.SetActiveCamera(player)
	animationSystem.SetCameraSystem(game.CameraSystem)
	animationSystem.SetPlayerEntity(player)

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"viewportCulling": true,
			"distanceLOD":     true,
			"closeThreshold":  200.0,
			"midThreshold":    400.0,
		}).Info("animation system configured with performance optimizations")
	}

	// Set player for UI systems
	game.HUDSystem.SetPlayerEntity(player)
	game.SetPlayerEntity(player)

	clientLogger.WithField("entityID", player.ID).Info("player entity created")
	return player
}

// addPlayerComponents adds gameplay components to the player entity.
func addPlayerComponents(player *engine.Entity, logger *logrus.Logger, clientLogger *logrus.Entry) (*engine.InventoryComponent, *engine.QuestTrackerComponent) {
	// Add player stats
	playerStats := engine.NewStatsComponent()
	playerStats.Attack = 10
	playerStats.Defense = 5
	playerStats.CritChance = 0.05
	playerStats.CritDamage = 1.5
	playerStats.Evasion = 0.05
	player.AddComponent(playerStats)

	// Add player experience/progression
	player.AddComponent(engine.NewExperienceComponent())

	// Add player inventory
	playerInventory := engine.NewInventoryComponent(20, 100.0)
	playerInventory.Gold = 100
	player.AddComponent(playerInventory)

	// Add player equipment
	player.AddComponent(engine.NewEquipmentComponent())

	// Add mana component
	playerMana := &engine.ManaComponent{
		Current: 100,
		Max:     100,
		Regen:   5.0,
	}
	player.AddComponent(playerMana)

	// Load spells
	if err := engine.LoadPlayerSpells(player, *seed, *genreID, 1); err != nil {
		clientLogger.WithError(err).Fatal("failed to load player spells")
	}
	if *verbose {
		clientLogger.Info("player spells loaded (keys 1-5)")
	}

	// Load skill tree
	if err := engine.LoadPlayerSkillTree(player, *seed, *genreID, 0); err != nil {
		clientLogger.WithError(err).Fatal("failed to load skill tree")
	}
	if *verbose {
		if comp, ok := player.GetComponent("skill_tree"); ok {
			treeComp, ok := comp.(*engine.SkillTreeComponent)
			if ok {
				clientLogger.WithFields(logrus.Fields{
					"treeName":   treeComp.Tree.Name,
					"skillCount": len(treeComp.Tree.Nodes),
				}).Info("skill tree loaded (press K)")
			}
		}
	}

	// Add quest tracker
	questTracker := engine.NewQuestTrackerComponent(5)
	player.AddComponent(questTracker)

	// Add attack capability
	player.AddComponent(&engine.AttackComponent{
		Damage:     15,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0.5,
	})

	// Add collision (Phase 45: collider matches 64×64 sprite)
	player.AddComponent(&engine.ColliderComponent{
		Width:     playerSpriteWidth,
		Height:    playerSpriteHeight,
		Solid:     true,
		IsTrigger: false,
		Layer:     1,
		OffsetX:   playerColliderOffset,
		OffsetY:   playerColliderOffset,
	})

	// Add visual feedback
	player.AddComponent(engine.NewVisualFeedbackComponent())

	// Add achievement tracking (Phase 26.2)
	player.AddComponent(&engine.AchievementComponent{
		Achievements:     []engine.Achievement{},
		ExpressionCount:  0,
		UniqueExpression: make(map[engine.ExpressionType]int),
	})

	// Add story journal for fragment discovery (Phase 30)
	player.AddComponent(engine.NewStoryJournalComponent())

	// Add mail component for mailbox system (Phase 40.3)
	player.AddComponent(&engine.MailComponent{
		Inbox:    []*engine.MailMessage{},
		Outbox:   []*engine.MailMessage{},
		MaxInbox: 50,
	})

	// Add adaptive soundtrack component for dynamic music (Phase 29)
	player.AddComponent(engine.NewAdaptiveSoundtrackComponent(*genreID))

	// Add QoL component for quality of life features persistence (AUDIT.md fix)
	player.AddComponent(&qol.QoLComponent{
		PlayerID:        player.ID,
		AutoLootEnabled: true,
		AutoLootRadius:  7.0,
		SortPreset:      "default",
		MountWhistle:    false,
		RecipeTracking:  false,
	})

	// Add starter items
	clientLogger.Info("adding starter items to inventory")
	addStarterItems(playerInventory, *seed, *genreID, logger)

	// Add tutorial quest
	clientLogger.Info("creating tutorial quest")
	addTutorialQuest(questTracker, *seed, *genreID, logger)

	return playerInventory, questTracker
}

// initializeTutorialAndHelp creates and configures tutorial and help systems.
// The showTutorials parameter comes from GameSettings.ShowTutorials.
// The screenWidth and screenHeight parameters ensure correct button positioning.
// Phase 3.3: Also creates and configures TutorialManager for context-sensitive help.
func initializeTutorialAndHelp(inputSystem *engine.InputSystem, cameraSystem *engine.CameraSystem, showTutorials bool, screenWidth, screenHeight int) (*engine.EbitenTutorialSystem, *engine.EbitenHelpSystem, *ui.TutorialManager) {
	tutorialSystem := engine.NewTutorialSystemWithSize(screenWidth, screenHeight)
	// Start the in-game tutorial disabled; OnboardingManager.TransitionToInGameTutorial()
	// will enable it at the correct time (after character creation completes).
	// For loaded games, deserializeTutorialState() restores the saved enabled/showUI state.
	tutorialSystem.Enabled = false
	tutorialSystem.ShowUI = false

	// If --no-tutorial is set or ShowTutorials is off, tutorials stay permanently disabled
	// (OnboardingManager.SetEnabled(false) handles this separately).
	helpSystem := engine.NewHelpSystem()

	// Phase 3.3: Create context-sensitive tutorial manager
	contextualTutorial := ui.NewTutorialManager()
	if *noTutorial || !showTutorials {
		contextualTutorial.Disable()
	}

	inputSystem.SetHelpSystem(helpSystem)
	inputSystem.SetTutorialSystem(tutorialSystem)
	inputSystem.SetCameraSystem(cameraSystem)

	return tutorialSystem, helpSystem, contextualTutorial
}

// configureSaveLoadSystem initializes the save/load manager and registers callbacks.
func configureSaveLoadSystem(player *engine.Entity, game *engine.EbitenGame, generatedTerrain *terrain.Terrain, inputSystem *engine.InputSystem, clientLogger *logrus.Entry) saveload.Manager {
	clientLogger.Info("initializing save/load system")

	// BUG FIX: Phase 7 - SaveManager auto-selects storage backend based on platform
	// - Desktop/mobile: file-based storage in ./saves directory
	// - WASM: localStorage with 5MB limit, fallback to in-memory
	// Platform: All (conditional compilation selects correct implementation)
	var saveManager saveload.Manager
	fileManager, err := saveload.NewSaveManager("./saves")
	if err != nil {
		// Fall back to in-memory storage to prevent nil pointer dereference
		clientLogger.WithError(err).Warn("failed to initialize file-based save manager, falling back to in-memory storage")
		saveManager = saveload.NewMemorySaveManager()
	} else {
		saveManager = fileManager
	}

	if *verbose {
		clientLogger.Info("save/load system initialized")
	}

	inputSystem.SetQuickSaveCallback(func() error {
		clientLogger.Info("quick save (F5 pressed)")
		gameSave := createGameSave(player, game, generatedTerrain)
		if err := saveManager.SaveGame("quicksave", gameSave); err != nil {
			clientLogger.WithError(err).Error("failed to save game")
			return err
		}
		clientLogger.Info("game saved successfully")
		return nil
	})

	inputSystem.SetQuickLoadCallback(func() error {
		clientLogger.Info("quick load (F9 pressed)")
		gameSave, err := saveManager.LoadGame("quicksave")
		if err != nil {
			clientLogger.WithError(err).Error("failed to load game")
			return err
		}
		loadGameSave(player, gameSave, game)
		clientLogger.Info("game loaded successfully")
		return nil
	})

	if *verbose {
		clientLogger.Info("quick save/load callbacks registered (F5/F9)")
	}

	return saveManager
}

// setupUICallbacks configures all UI input callbacks and menu system integrations.
func setupUICallbacks(game *engine.EbitenGame, player *engine.Entity, generatedTerrain *terrain.Terrain, inputSystem *engine.InputSystem, objectiveTracker *engine.ObjectiveTrackerSystem, dialogSystem *engine.DialogSystem, shopUI *engine.ShopUI, saveManager saveload.Manager, clientLogger *logrus.Entry) error {
	if *verbose {
		clientLogger.Info("setting up UI input callbacks")
	}

	if err := game.SetupInputCallbacks(inputSystem, objectiveTracker); err != nil {
		return fmt.Errorf("failed to setup input callbacks: %w", err)
	}

	connectUIComponentsToInputSystem(game, inputSystem, shopUI, player, clientLogger)

	if *verbose {
		clientLogger.Info("UI callbacks registered (I: Inventory, J: Quests, ESC: Pause Menu)")
		clientLogger.Info("inventory actions: E to equip/use, D to drop")
	}

	if err := setupMerchantInteraction(player, game, dialogSystem, shopUI, inputSystem, clientLogger); err != nil {
		return fmt.Errorf("failed to setup merchant interaction: %w", err)
	}

	if *verbose {
		clientLogger.Info("merchant interaction registered (F key when near merchant)")
	}

	if err := connectMenuSaveLoad(game, player, generatedTerrain, saveManager, clientLogger); err != nil {
		return fmt.Errorf("failed to connect menu save/load: %w", err)
	}

	return nil
}

// connectUIComponentsToInputSystem connects all UI components to input system for ESC key handling.
// BUG FIX: Phase 3 - Menu Trap - Enables dual-exit pattern (toggle key + ESC key close) for ALL UI panels.
func connectUIComponentsToInputSystem(game *engine.EbitenGame, inputSystem *engine.InputSystem, shopUI *engine.ShopUI, player *engine.Entity, clientLogger *logrus.Entry) {
	connectBasicUIComponents(game, inputSystem, shopUI)
	connectAdvancedUIComponents(game, inputSystem)
	connectDialogUI(game, inputSystem, player, clientLogger)
}

func connectBasicUIComponents(game *engine.EbitenGame, inputSystem *engine.InputSystem, shopUI *engine.ShopUI) {
	uiConnections := []struct {
		ui     interface{}
		setter func(interface{})
	}{
		{game.MailboxUI, func(ui interface{}) { inputSystem.SetMailboxUI(ui.(*engine.MailboxUI)) }},
		{game.InventoryUI, func(ui interface{}) { inputSystem.SetInventoryUI(ui.(*engine.EbitenInventoryUI)) }},
		{game.CharacterUI, func(ui interface{}) { inputSystem.SetCharacterUI(ui.(*engine.EbitenCharacterUI)) }},
		{game.SkillsUI, func(ui interface{}) { inputSystem.SetSkillsUI(ui.(*engine.EbitenSkillsUI)) }},
		{game.QuestUI, func(ui interface{}) { inputSystem.SetQuestUI(ui.(*engine.EbitenQuestUI)) }},
		{game.MapUI, func(ui interface{}) { inputSystem.SetMapUI(ui.(*engine.EbitenMapUI)) }},
		{game.CraftingUI, func(ui interface{}) { inputSystem.SetCraftingUI(ui.(*engine.CraftingUI)) }},
		{shopUI, func(ui interface{}) { inputSystem.SetShopUI(ui.(*engine.ShopUI)) }},
		{game.TradeUI, func(ui interface{}) { inputSystem.SetTradeUI(ui.(*engine.TradeUI)) }},
	}

	connectUIComponents(uiConnections)
}

// connectUIComponents connects non-nil UI components using the provided setters.
func connectUIComponents(connections []struct {
	ui     interface{}
	setter func(interface{})
},
) {
	for _, conn := range connections {
		if conn.ui != nil {
			conn.setter(conn.ui)
		}
	}
}

func connectAdvancedUIComponents(game *engine.EbitenGame, inputSystem *engine.InputSystem) {
	if game.AdvancedClassUI != nil {
		inputSystem.SetAdvancedClassUI(game.AdvancedClassUI)
		inputSystem.SetClassesCallback(func() {
			game.AdvancedClassUI.Toggle()
		})
	}
	if game.TerritoryUI != nil {
		inputSystem.SetTerritoryUI(game.TerritoryUI)
		inputSystem.SetTerritoryCallback(func() {
			game.TerritoryUI.Toggle()
		})
	}
	// INTEGRATION FIX [Category B]: V8.0 Housing UI ESC key handling (Phase 49.1)
	// Gap: SetHousingUI was defined but never called, so ESC key couldn't close HousingUI
	// Fix: Wire HousingUI to InputSystem for ESC key dual-exit pattern
	if game.HousingUI != nil {
		inputSystem.SetHousingUI(game.HousingUI)
	}
}

func connectDialogUI(game *engine.EbitenGame, inputSystem *engine.InputSystem, player *engine.Entity, clientLogger *logrus.Entry) {
	if game.DialogUI == nil {
		return
	}

	inputSystem.SetDialogUI(game.DialogUI)
	inputSystem.SetDialogCallback(func() {
		nearestNPC := findNearestNPC(game, player)
		if nearestNPC != nil {
			if err := game.DialogUI.Show(nearestNPC.ID); err != nil {
				clientLogger.WithError(err).Warn("failed to open dialog UI")
			}
		}
	})
}

func findNearestNPC(game *engine.EbitenGame, player *engine.Entity) *engine.Entity {
	if player == nil {
		return nil
	}

	posComp, ok := player.GetComponent("position")
	if !ok {
		return nil
	}
	pos, ok := posComp.(*engine.PositionComponent)
	if !ok {
		return nil
	}

	var nearestNPC *engine.Entity
	minDist := 5.0 * 32.0
	minDistSquared := minDist * minDist

	for _, entity := range game.World.GetEntities() {
		if entity.ID == player.ID {
			continue
		}
		if _, hasDialog := entity.GetComponent("dialog"); !hasDialog {
			continue
		}

		if ePos, ok := entity.GetComponent("position"); ok {
			entityPos, ok := ePos.(*engine.PositionComponent)
			if !ok {
				continue
			}
			dx := entityPos.X - pos.X
			dy := entityPos.Y - pos.Y
			distSquared := dx*dx + dy*dy

			if distSquared < minDistSquared {
				minDistSquared = distSquared
				nearestNPC = entity
			}
		}
	}

	return nearestNPC
}

// setupMerchantInteraction configures the F key interaction callback for merchants.
func setupMerchantInteraction(player *engine.Entity, game *engine.EbitenGame, dialogSystem *engine.DialogSystem, shopUI *engine.ShopUI, inputSystem *engine.InputSystem, clientLogger *logrus.Entry) error {
	return inputSystem.SetInteractCallback(func() {
		if player == nil {
			return
		}
		posComp, ok := player.GetComponent("position")
		if !ok {
			return
		}
		pos, ok := posComp.(*engine.PositionComponent)
		if !ok {
			return
		}

		merchant, dist := engine.FindClosestMerchant(game.World, pos.X, pos.Y, merchantInteractionRange)
		if merchant == nil {
			if *verbose {
				clientLogger.Debug("no merchant nearby to interact with")
			}
			return
		}

		success, err := dialogSystem.StartDialog(player.ID, merchant.ID)
		if err != nil {
			clientLogger.WithError(err).Warn("failed to start dialog")
			return
		}

		if !success {
			if *verbose {
				clientLogger.Debug("dialog could not be started")
			}
			return
		}

		shopUI.Open(merchant)

		if *verbose {
			clientLogger.WithField("distance", dist).Debug("opened shop with merchant")
		}
	})
}

// connectMenuSaveLoad wires save/load callbacks to the menu system.
func connectMenuSaveLoad(game *engine.EbitenGame, player *engine.Entity, generatedTerrain *terrain.Terrain, saveManager saveload.Manager, clientLogger *logrus.Entry) error {
	if game.MenuSystem == nil || saveManager == nil {
		return nil
	}

	if *verbose {
		clientLogger.Info("connecting save/load callbacks to menu system")
	}

	saveCallback := func(saveName string) error {
		if *verbose {
			clientLogger.WithField("saveName", saveName).Info("menu save")
		}
		gameSave := createGameSave(player, game, generatedTerrain)
		if err := saveManager.SaveGame(saveName, gameSave); err != nil {
			clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to save game")
			return err
		}
		clientLogger.WithField("saveName", saveName).Info("game saved successfully")
		return nil
	}

	loadCallback := func(saveName string) error {
		if *verbose {
			clientLogger.WithField("saveName", saveName).Info("menu load")
		}
		gameSave, err := saveManager.LoadGame(saveName)
		if err != nil {
			clientLogger.WithError(err).WithField("saveName", saveName).Error("failed to load game")
			return err
		}
		loadGameSave(player, gameSave, game)
		clientLogger.WithField("saveName", saveName).Info("game loaded successfully")
		return nil
	}

	game.MenuSystem.SetSaveCallback(saveCallback)
	game.MenuSystem.SetLoadCallback(loadCallback)

	if *verbose {
		clientLogger.Info("save/load callbacks connected to menu system")
	}

	return nil
}

// initializeUIIntegration sets up shop UI, crafting UI, mailbox UI, housing UI, gallery UI, and connects them to game systems.
func initializeUIIntegration(game *engine.EbitenGame, player *engine.Entity, commerceSystem *engine.CommerceSystem, dialogSystem *engine.DialogSystem, craftingSystem *engine.CraftingSystem, inventorySystem *engine.InventorySystem, sys *systemsContainer, clientLogger *logrus.Entry) (*engine.ShopUI, *engine.CraftingUI) {
	shopUI := initializeShopAndCommerceUI(game, player, commerceSystem, dialogSystem, clientLogger)
	initializeHousingAndGalleryUI(game, player, sys, clientLogger)
	initializeGuildAndTradeUI(game, player, sys, clientLogger)
	initializeAdvancedAndTerritoryUI(game, player, sys, clientLogger)
	initializeStoryAndDialogUI(game, player, sys, clientLogger)
	craftingUI := initializeCraftingAndMailboxUI(game, player, craftingSystem, inventorySystem, clientLogger)

	return shopUI, craftingUI
}

func initializeShopAndCommerceUI(game *engine.EbitenGame, player *engine.Entity, commerceSystem *engine.CommerceSystem, dialogSystem *engine.DialogSystem, clientLogger *logrus.Entry) *engine.ShopUI {
	shopUI := engine.NewShopUI(*width, *height)
	shopUI.SetPlayerEntity(player)
	shopUI.SetCommerceSystem(commerceSystem)
	shopUI.SetDialogSystem(dialogSystem)
	game.ShopUI = shopUI

	if *verbose {
		clientLogger.Info("shop UI initialized and connected to commerce/dialog systems")
	}

	return shopUI
}

func initializeHousingAndGalleryUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	housingUI := housing.NewHousingUI(*width, *height)
	if sys.housingManager != nil {
		housingUI.SetManagers(sys.housingManager, sys.guildHallManager, sys.buildingGenerator, sys.furnitureGenerator)
		housingUI.SetPlayerID(player.ID)
	}
	game.HousingUI = housingUI

	if *verbose {
		clientLogger.Info("housing UI initialized (H key to open)")
	}

	// Create per-player ChatHistory and ImageGallery with actual player ID
	playerIDStr := fmt.Sprintf("player_%d", player.ID)
	sys.chatHistory = persistence.NewChatHistory(playerIDStr)
	sys.imageGallery = persistence.NewImageGallery(playerIDStr)

	clientLogger.WithField("playerID", playerIDStr).Debug("per-player chat history and image gallery initialized")

	galleryUI := engine.NewGalleryUI(*width, *height)
	if sys.imageGallery != nil {
		galleryUI.SetGallery(sys.imageGallery)
	}
	game.GalleryUI = galleryUI

	if *verbose {
		clientLogger.Info("gallery UI initialized (G key to open)")
	}
}

func initializeGuildAndTradeUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	if sys.guildUI != nil {
		game.GuildUI = sys.guildUI
		if *verbose {
			clientLogger.Info("guild UI initialized (U key to open)")
		}
	}

	tradeUI := engine.NewTradeUI(game.World, sys.tradeSystem, *width, *height)
	tradeUI.SetPlayerEntity(player)
	game.TradeUI = tradeUI

	if *verbose {
		clientLogger.Info("trade UI initialized (T key to open)")
	}
}

func initializeAdvancedAndTerritoryUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	advancedClassUI := engine.NewAdvancedClassUI(game.World, sys.advancedClassSystem, *width, *height)
	advancedClassUI.SetPlayerEntity(player)
	game.AdvancedClassUI = advancedClassUI

	if *verbose {
		clientLogger.Info("advanced class UI initialized (A key to open)")
	}

	sys.territoryUI.SetPlayerEntity(player)
	game.TerritoryUI = sys.territoryUI

	if *verbose {
		clientLogger.Info("territory UI initialized (Y key to open)")
	}
}

func initializeStoryAndDialogUI(game *engine.EbitenGame, player *engine.Entity, sys *systemsContainer, clientLogger *logrus.Entry) {
	storyChoiceUI := engine.NewStoryChoiceUI(game.World, sys.branchingNarrativeSystem, *width, *height)
	storyChoiceUI.SetPlayerEntity(player)
	game.StoryChoiceUI = storyChoiceUI

	if *verbose {
		clientLogger.Info("story choice UI initialized (shows automatically when choices available)")
	}

	initializePlayerNarrative(player, sys.branchingNarrativeSystem, *seed, *genreID, clientLogger)

	dialogUI := engine.NewDialogUI(game.World, sys.dialogSystem, sys.npcDialogSystem, *width, *height)
	dialogUI.SetPlayerEntity(player)
	game.DialogUI = dialogUI

	if *verbose {
		clientLogger.Info("dialog UI initialized (D key to toggle with NPCs)")
	}
}

func initializeCraftingAndMailboxUI(game *engine.EbitenGame, player *engine.Entity, craftingSystem *engine.CraftingSystem, inventorySystem *engine.InventorySystem, clientLogger *logrus.Entry) *engine.CraftingUI {
	craftingUI := engine.NewCraftingUI(*width, *height)
	craftingUI.SetPlayerEntity(player)
	craftingUI.SetCraftingSystem(craftingSystem)
	game.CraftingUI = craftingUI

	if *verbose {
		clientLogger.Info("crafting UI initialized and connected to crafting system")
	}

	mailboxUI := engine.NewMailboxUI(0, 0, *width, *height, *genreID)
	game.MailboxUI = mailboxUI

	if *verbose {
		clientLogger.Info("mailbox UI initialized (Phase 40.3)")
	}

	game.SetInventorySystem(inventorySystem)

	return craftingUI
}

// applyCharacterClass applies character class stats if character data is pending.
func applyCharacterClass(player *engine.Entity, game *engine.EbitenGame, clientLogger *logrus.Entry) {
	charData := game.GetPendingCharacterData()
	if charData == nil {
		return
	}

	clientLogger.WithFields(logrus.Fields{
		"name":  charData.Name,
		"class": charData.Class.String(),
	}).Info("applying character class stats")

	if err := engine.ApplyClassStats(player, charData.Class); err != nil {
		clientLogger.WithError(err).Fatal("failed to apply character class stats")
	}

	// Apply equipment loadout bonuses if a loadout was selected
	applyEquipmentLoadout(player, charData, clientLogger)
}

// applyEquipmentLoadout applies the selected starting equipment loadout bonuses.
func applyEquipmentLoadout(player *engine.Entity, charData *engine.CharacterData, clientLogger *logrus.Entry) {
	if charData.StartingLoadout == nil {
		return
	}

	loadout := charData.StartingLoadout
	clientLogger.WithFields(logrus.Fields{
		"loadout":      loadout.Name,
		"bonusHP":      loadout.BonusHP,
		"bonusAttack":  loadout.BonusAttack,
		"bonusDefense": loadout.BonusDefense,
	}).Info("applying equipment loadout bonuses")

	// Apply HP bonus
	if comp, exists := player.GetComponent("health"); exists {
		if healthComp, ok := comp.(*engine.HealthComponent); ok && healthComp != nil {
			healthComp.Max += float64(loadout.BonusHP)
			healthComp.Current += float64(loadout.BonusHP)
		}
	}

	// Apply Attack and Defense bonuses
	if comp, exists := player.GetComponent("stats"); exists {
		if statsComp, ok := comp.(*engine.StatsComponent); ok && statsComp != nil {
			statsComp.Attack += float64(loadout.BonusAttack)
			statsComp.Defense += float64(loadout.BonusDefense)
		}
	}
}

// initializePlayerAdvancedClass initializes the advanced class system for the player.
// Phase 4.2 (PLAN.md): Multi-classing, prestige classes, and talent trees
func initializePlayerAdvancedClass(player *engine.Entity, game *engine.EbitenGame, clientLogger *logrus.Entry) {
	charData := game.GetPendingCharacterData()
	if charData == nil {
		return
	}

	// Map CharacterClass to advanced.ClassID
	classID := mapCharacterClassToAdvancedClass(charData.Class)
	if classID == "" {
		clientLogger.Warn("character class not mapped to advanced class system")
		return
	}

	// Initialize with level 1 and warrior class (can be customized via character creation)
	// Get the advanced class system from game systems
	// Note: We can't access sys.advancedClassSystem here directly,
	// so we initialize the component manually and the system will pick it up
	player.AddComponent(&advanced.AdvancedClassComponent{
		PrimaryClass: classID,
		Level:        1,
		TalentPoints: advanced.TalentAllocation{
			Talents:     make(map[advanced.TalentID]int),
			PointsTotal: 1,
		},
	})

	clientLogger.WithFields(logrus.Fields{
		"class": classID,
		"level": 1,
	}).Info("initialized advanced class system")
}

// mapCharacterClassToAdvancedClass maps engine.CharacterClass to advanced.ClassID
func mapCharacterClassToAdvancedClass(class engine.CharacterClass) advanced.ClassID {
	switch class {
	case engine.ClassWarrior:
		return advanced.ClassWarrior
	case engine.ClassMage:
		return advanced.ClassMage
	case engine.ClassRogue:
		return advanced.ClassRogue
	case engine.ClassRanger:
		return advanced.ClassRanger
	case engine.ClassCleric:
		return advanced.ClassCleric
	case engine.ClassNecromancer:
		return advanced.ClassNecromancer
	case engine.ClassBattlemage:
		return advanced.ClassWarrior // Default to warrior for hybrids (can be improved)
	default:
		return ""
	}
}

// finalizeGameInitialization processes initial entity updates and logs game start info.
func finalizeGameInitialization(game *engine.EbitenGame, player *engine.Entity, networkClient interface{}, clientLogger *logrus.Entry) {
	game.World.Update(0)

	clientLogger.Info("game initialized successfully")
	clientLogger.Info("controls: WASD to move, Space to attack, E to use item, I: Inventory, J: Quests, L: Mailbox, A: Classes")
	clientLogger.WithFields(logrus.Fields{"genre": *genreID, "seed": *seed, "server": *server}).Info("game settings")
}

// handleHostAndPlay starts embedded server if host-and-play mode is enabled.
// Returns cleanup function that should be deferred by caller.
func handleHostAndPlay(logger *logrus.Logger, clientLogger *logrus.Entry) func() {
	if !*hostAndPlay {
		return func() {} // Return no-op cleanup
	}

	// Log message depends on whether this was explicitly requested or auto-enabled
	// Auto-enabled case already logged in main.go, so we skip redundant logging here
	if !autoEnabledHostAndPlay {
		if *hostLAN {
			clientLogger.Info("host-and-play mode - starting LAN-accessible server on 0.0.0.0")
		} else {
			clientLogger.Info("host-and-play mode - starting localhost server on 127.0.0.1")
		}
	}

	serverAddr, cleanup, err := startEmbeddedServer(logger, *seed, *genreID)
	if err != nil {
		clientLogger.WithError(err).Fatal("failed to start embedded server")
	}

	*server = serverAddr
	*multiplayer = true

	clientLogger.WithField("serverAddr", serverAddr).Info("embedded server started, connecting client")

	return cleanup // Return cleanup to be deferred by caller
}

// createGameInstance initializes the main game instance with logging and profiling.
func createGameInstance(logger *logrus.Logger, clientLogger *logrus.Entry) *engine.EbitenGame {
	game := engine.NewEbitenGameWithLogger(*width, *height, logger)
	game.SetFullscreen(*fullscreen)

	// Set world seed for deterministic character naming
	game.SetWorldSeed(*seed)

	if *profile {
		game.EnableFrameTimeProfiling()
		clientLogger.Info("performance profiling enabled - frame time stats will be logged every 5 seconds")
	}

	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"width":      *width,
			"height":     *height,
			"fullscreen": *fullscreen,
			"seed":       *seed,
		}).Info("display initialized")
	}

	return game
}

// initializeVirtualControls sets up touch controls for mobile/web platforms.
func initializeVirtualControls(inputSystem *engine.InputSystem, clientLogger *logrus.Entry) {
	if !mobile.IsTouchCapable() {
		return
	}

	inputSystem.InitializeVirtualControls(*width, *height)
	clientLogger.WithFields(logrus.Fields{
		"platform": mobile.GetPlatform().String(),
		"width":    *width,
		"height":   *height,
	}).Info("virtual controls initialized for touch-capable platform")
}

// configureDeathCallback sets up the death callback for combat system.
func configureDeathCallback(sys *systemsContainer, game *engine.EbitenGame, logger *logrus.Logger) {
	var playerEntity *engine.Entity
	sys.combatSystem.SetDeathCallback(createDeathCallback(
		game, &playerEntity, sys.objectiveTracker, &sys.audioManager,
		sys.recipeGen, sys.magicGenerator, sys.skillGenerator, sys.deathParticleSystem, *seed, *genreID, logger, sys.timeProvider,
	))
}

// createGenerationParams creates standard generation parameters for world content.
func createGenerationParams() procgen.GenerationParams {
	return procgen.GenerationParams{
		Difficulty: defaultDifficulty,
		Depth:      defaultDepth,
		GenreID:    *genreID,
		Custom: map[string]interface{}{
			"width":  defaultTerrainWidth,
			"height": defaultTerrainHeight,
			"theme":  getGenreTheme(),
		},
	}
}

// cleanupNetworkClient disconnects the network client on shutdown.
func cleanupNetworkClient(networkClient interface{}, clientLogger *logrus.Entry) {
	if networkClient == nil {
		return
	}

	type disconnector interface {
		Disconnect() error
	}

	if dc, ok := networkClient.(disconnector); ok {
		clientLogger.Info("disconnecting from server")
		if err := dc.Disconnect(); err != nil {
			clientLogger.WithError(err).Warn("error disconnecting")
		}
	}
}

// runGameLoop starts the main game loop.
// Note: System wrappers have been moved to system_wrappers.go for better organization.
func runGameLoop(game *engine.EbitenGame, clientLogger *logrus.Entry) {
	windowTitle := fmt.Sprintf("Venture %s - Procedural Action RPG", version.Version)
	if err := game.Run(windowTitle); err != nil {
		clientLogger.WithError(err).Fatal("error running game")
	}
}

// generateNarrativeArcWithLogging generates a procedural story arc for the world narrative with optional verbose logging.
// Phase 3.6 (PLAN.md): Integrates pkg/procgen/narrative to create overarching world story
func generateNarrativeArcWithLogging(w *engine.World, narrativeGen *narrative.StoryArcGenerator, params procgen.GenerationParams, clientLogger *logrus.Entry, tp TimeProvider) {
	if narrativeGen == nil {
		clientLogger.Warn("narrative generator is nil, skipping story arc generation")
		return
	}

	if *verbose {
		clientLogger.Info("generating procedural narrative arc for world")
	}

	arc := generateAndValidateNarrativeArc(narrativeGen, params, clientLogger)
	if arc == nil {
		return
	}

	worldNarrativeEntity := findOrCreateWorldNarrativeEntity(w)
	if worldNarrativeEntity == nil {
		clientLogger.Warn("failed to create world narrative entity")
		return
	}

	narrativeComp := setupNarrativeComponent(worldNarrativeEntity, arc)
	addInitialNarrativeEvent(narrativeComp, arc, tp)
	addPlotPointsAsThreads(narrativeComp, arc)
	logNarrativeArcSuccess(arc, clientLogger)
}

// generateAndValidateNarrativeArc generates and validates a story arc.
func generateAndValidateNarrativeArc(narrativeGen *narrative.StoryArcGenerator, params procgen.GenerationParams, clientLogger *logrus.Entry) *narrative.StoryArc {
	result, err := narrativeGen.Generate(*seed+seedOffsetNarrative, params)
	if err != nil {
		clientLogger.WithError(err).Warn("failed to generate narrative arc")
		return nil
	}

	arc, ok := result.(*narrative.StoryArc)
	if !ok {
		clientLogger.Warn("narrative generator returned invalid type")
		return nil
	}

	if err := narrativeGen.Validate(arc); err != nil {
		clientLogger.WithError(err).Warn("generated narrative arc failed validation")
		return nil
	}

	return arc
}

// setupNarrativeComponent gets or creates narrative component with story arc data.
func setupNarrativeComponent(worldNarrativeEntity *engine.Entity, arc *narrative.StoryArc) *engine.NarrativeComponent {
	var narrativeComp *engine.NarrativeComponent
	if comp, ok := worldNarrativeEntity.GetComponent("narrative"); ok {
		narrativeComp, _ = comp.(*engine.NarrativeComponent)
	}
	if narrativeComp == nil {
		narrativeComp = &engine.NarrativeComponent{
			CurrentAct:      engine.ActSetup,
			EventHistory:    make([]engine.NarrativeEvent, 0),
			MainObjective:   arc.MainConflict,
			StoryProgress:   0.0,
			ActiveThreads:   make([]string, 0),
			ResolvedThreads: make([]string, 0),
			WorldStateFlags: make(map[string]bool),
			Relationships:   make(map[string]float64),
			PlayerDecisions: make([]engine.PlayerDecision, 0),
		}
		worldNarrativeEntity.AddComponent(narrativeComp)
	} else {
		narrativeComp.MainObjective = arc.MainConflict
	}
	return narrativeComp
}

// addInitialNarrativeEvent adds the initial narrative event for the story arc.
func addInitialNarrativeEvent(narrativeComp *engine.NarrativeComponent, arc *narrative.StoryArc, tp TimeProvider) {
	initialEvent := engine.NarrativeEvent{
		Type:        engine.EventDiscovery,
		Timestamp:   tp.Now(),
		Description: fmt.Sprintf("The story begins: %s", arc.Title),
		Importance:  1.0,
		Act:         engine.ActSetup,
	}
	narrativeComp.AddEvent(initialEvent)
}

// addPlotPointsAsThreads adds Act 1 plot points as active narrative threads.
func addPlotPointsAsThreads(narrativeComp *engine.NarrativeComponent, arc *narrative.StoryArc) {
	for _, plotPoint := range arc.PlotPoints {
		if plotPoint.Act == 1 {
			narrativeComp.ActiveThreads = append(narrativeComp.ActiveThreads, plotPoint.Description)
		}
	}
}

// logNarrativeArcSuccess logs successful narrative arc generation with details.
func logNarrativeArcSuccess(arc *narrative.StoryArc, clientLogger *logrus.Entry) {
	if *verbose {
		clientLogger.WithFields(logrus.Fields{
			"arc_title":   arc.Title,
			"antagonist":  arc.Antagonist,
			"ally":        arc.Ally,
			"plot_points": len(arc.PlotPoints),
			"endings":     len(arc.PossibleEndings),
			"difficulty":  arc.Difficulty,
			"seed":        arc.Seed,
		}).Info("procedural narrative arc generated and integrated")
	}
}

// findOrCreateWorldNarrativeEntity finds existing world narrative entity or creates a new one.
func findOrCreateWorldNarrativeEntity(w *engine.World) *engine.Entity {
	// Search for existing world narrative entity
	entities := w.GetEntities()
	for _, entity := range entities {
		if _, ok := entity.GetComponent("narrative"); ok {
			// Check if this is a world narrative entity (not player narrative)
			if _, hasPlayer := entity.GetComponent("player"); !hasPlayer {
				return entity
			}
		}
	}

	// Create new world narrative entity
	worldNarrative := w.CreateEntity()
	return worldNarrative
}

// initializePlayerNarrative initializes branching narrative for the player.
// Phase 6.1 (PLAN.md): Creates a procedural story arc and starts player on narrative journey.
func initializePlayerNarrative(player *engine.Entity, narrativeSystem *engine.BranchingNarrativeSystem, seed int64, genreID string, logger *logrus.Entry) {
	// Add branching narrative component to player
	player.AddComponent(&engine.BranchingNarrativeComponent{
		LastUpdate: 0,
	})

	// Create narrative generator and manager
	generator := branching.NewGenerator()
	manager := branching.NewManager()

	// Generate a story arc for the player
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"theme": getGenreTheme(),
		},
	}

	result, err := generator.Generate(seed, params)
	if err != nil {
		logger.WithError(err).Warn("failed to generate branching narrative arc")
		return
	}

	arc, ok := result.(*branching.StoryArc)
	if !ok {
		logger.Warn("branching narrative generator returned invalid type")
		return
	}

	// Start the story arc
	if err := narrativeSystem.StartStoryArc(player, arc, manager); err != nil {
		logger.WithError(err).Warn("failed to start branching narrative arc")
		return
	}

	logger.WithFields(logrus.Fields{
		"arc_id":     arc.ID,
		"arc_title":  arc.Title,
		"node_count": len(arc.Nodes),
	}).Info("player branching narrative initialized")
}
