// Package engine provides the combat system for damage and status effects.
// This file implements CombatSystem which handles damage calculation, combat
// interactions, and status effect management using the combat package.
package engine

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// CombatSystem handles combat interactions, damage calculation, and status effects.
type CombatSystem struct {
	rng *rand.Rand

	// Camera reference for screen shake feedback (GAP-012)
	camera *CameraSystem

	// GAP-016 REPAIR: Particle system for hit effects
	particleSystem *ParticleSystem
	world          *World
	seed           int64
	genreID        string

	// Phase 10.2: Projectile system for ranged weapon physics
	projectileSystem *ProjectileSystem

	// Plan Phase 1.1: Audio manager for combat SFX
	audioManager *AudioManager

	// Combat resolver for damage calculation (uses authoritative combat package)
	combatResolver combat.CombatResolver

	// Callback for when an entity dies
	onDeathCallback func(entity *Entity)

	// Callback for when damage is dealt
	onDamageCallback func(attacker, target *Entity, damage float64)

	// Additional damage callbacks for multiple systems
	additionalDamageCallbacks []func(attacker, target *Entity, damage float64)

	// Callback for when a critical hit occurs
	onCriticalHitCallback func(attacker, target *Entity, damage float64)

	// Additional critical hit callbacks for multiple systems
	additionalCriticalHitCallbacks []func(attacker, target *Entity, damage float64)

	// Callback for when damage is significantly reduced by resistance
	onDamageResistedCallback func(target *Entity, damageType combat.DamageType, originalDamage, finalDamage, resistance float64)

	// Callback for when a shield absorbs damage
	onShieldAbsorbCallback func(target *Entity, absorbed, remaining float64)

	// Callback for when an attack is evaded
	onEvasionCallback func(attacker, target *Entity, evasionChance float64)

	// Callback for when an entity is healed
	onHealCallback func(healer, target *Entity, amount float64)

	// Callback for when an attack is blocked (damage reduced)
	onBlockCallback func(attacker, target *Entity, blockChance, originalDamage, reducedDamage float64)

	// Callback for when an entity is killed by another entity (with attacker info)
	onKillCallback func(attacker, target *Entity)

	// Logger for combat events
	logger *logrus.Entry

	// processedDeaths tracks entities whose death callback was already invoked
	// this session.  It prevents re-invocation on frames where the entity is
	// still in the entity list but its health is still ≤0.  The callback is
	// invoked BEFORE DeadComponent is attached, preserving the original contract
	// where the callback is responsible for adding DeadComponent itself.
	//
	// Lifetime: the map is never explicitly cleared.  A new CombatSystem is
	// created for each World/level, so entity IDs from a prior World will not
	// collide.  Call ClearProcessedDeaths() if entity IDs are recycled within a
	// single long-lived World instance.
	processedDeaths map[uint64]struct{}
}

// ClearProcessedDeaths removes all entries from the internal death-tracking map.
// Call this when entity IDs may be recycled (e.g. object-pool entity reuse) to
// ensure the death callback fires correctly for reused IDs.
func (s *CombatSystem) ClearProcessedDeaths() {
	s.processedDeaths = make(map[uint64]struct{})
}

// NewCombatSystem creates a new combat system with a given random seed.
func NewCombatSystem(seed int64) *CombatSystem {
	return NewCombatSystemWithLogger(seed, nil)
}

// NewCombatSystemWithLogger creates a new combat system with a logger.
func NewCombatSystemWithLogger(seed int64, logger *logrus.Logger) *CombatSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "combat",
			"seed":   seed,
		})
		logEntry.Debug("combat system created")
	}

	// Use the authoritative combat package's damage calculation formulas
	combatResolver := combat.NewDefaultCombatResolver(nil)

	return &CombatSystem{
		rng:             rand.New(rand.NewSource(seed)),
		seed:            seed,
		logger:          logEntry,
		combatResolver:  combatResolver,
		processedDeaths: make(map[uint64]struct{}),
	}
}

// SetCamera sets the camera reference for screen shake feedback (GAP-012).
func (s *CombatSystem) SetCamera(camera *CameraSystem) {
	if s.logger != nil {
		s.logger.Debug("camera system linked to combat system")
	}
	s.camera = camera
}

// GAP-016 REPAIR: SetParticleSystem sets the particle system reference for hit effects.
func (s *CombatSystem) SetParticleSystem(ps *ParticleSystem, world *World, genreID string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"genre_id":            genreID,
			"has_particle_system": ps != nil,
			"has_world":           world != nil,
		}).Debug("particle system linked to combat system")
	}
	s.particleSystem = ps
	s.world = world
	s.genreID = genreID
}

// SetProjectileSystem sets the projectile system reference for ranged weapon spawning (Phase 10.2).
func (s *CombatSystem) SetProjectileSystem(ps *ProjectileSystem) {
	if s.logger != nil {
		s.logger.WithField("has_projectile_system", ps != nil).Debug("projectile system linked to combat system")
	}
	s.projectileSystem = ps
}

// SetAudioManager sets the audio manager reference for combat SFX (Plan Phase 1.1).
func (s *CombatSystem) SetAudioManager(am *AudioManager) {
	if s.logger != nil {
		s.logger.WithField("has_audio_manager", am != nil).Debug("audio manager linked to combat system")
	}
	s.audioManager = am
}

// Update implements the System interface.
// Updates attack cooldowns and processes status effects.
func (s *CombatSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		s.updateEntityCombat(entity, deltaTime)
	}
	s.processDeadEntities(entities)
}

// updateEntityCombat updates attack cooldowns and status effects for an entity.
func (s *CombatSystem) updateEntityCombat(entity *Entity, deltaTime float64) {
	isDead := entity.HasComponent("dead")

	if entity.HasComponent("input") && isDead && s.logger != nil {
		s.logger.WithField("entityID", entity.ID).Warn("player entity has dead component")
	}

	if !isDead {
		s.updateAttackCooldown(entity, deltaTime)
		s.applyBaseHealthRegen(entity, deltaTime)
	}

	s.processStatusEffects(entity, deltaTime)
}

// applyBaseHealthRegen ticks HealthComponent.RegenRate into health.Current for
// living entities.  Written by AttributeAllocationSystem (G26 VIT→regen fix).
func (s *CombatSystem) applyBaseHealthRegen(entity *Entity, deltaTime float64) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok || health.RegenRate <= 0 || health.Current >= health.Max {
		return
	}
	health.Current += health.RegenRate * deltaTime
	if health.Current > health.Max {
		health.Current = health.Max
	}
}

// updateAttackCooldown updates attack cooldown for living entities.
func (s *CombatSystem) updateAttackCooldown(entity *Entity, deltaTime float64) {
	attackComp, ok := entity.GetComponent("attack")
	if !ok {
		return
	}

	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return
	}

	beforeCooldown := attack.CooldownTimer
	attack.UpdateCooldown(deltaTime)

	if entity.HasComponent("input") && beforeCooldown > 0 && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID":       entity.ID,
			"cooldownBefore": beforeCooldown,
			"cooldownAfter":  attack.CooldownTimer,
			"deltaTime":      deltaTime,
		}).Debug("player attack cooldown updated")
	}
}

// processStatusEffects processes status effects for an entity.
func (s *CombatSystem) processStatusEffects(entity *Entity, deltaTime float64) {
	statusComp, ok := entity.GetComponent("status_effect")
	if !ok {
		return
	}

	status, ok := statusComp.(*StatusEffectComponent)
	if !ok {
		return
	}

	if ticked := status.Update(deltaTime); ticked {
		s.applyStatusEffectTick(entity, status)
	}

	if status.IsExpired() {
		entity.RemoveComponent("status_effect")
	}
}

// processDeadEntities handles death callbacks for dead entities.
func (s *CombatSystem) processDeadEntities(entities []*Entity) {
	for _, entity := range entities {
		s.handleEntityDeath(entity)
	}
}

// handleEntityDeath checks and handles entity death.
// Death callback contract (G28 / review fix):
//   - The callback is invoked BEFORE DeadComponent is attached, preserving the
//     original contract: the callback is the single transaction responsible for
//     adding DeadComponent, awarding XP, dropping loot, etc.
//   - An internal processedDeaths map prevents re-invocation on subsequent frames
//     where the entity is still in the entity list but health is still ≤ 0.
//   - If no callback is registered (or the callback does not add DeadComponent),
//     handleEntityDeath adds DeadComponent itself as a fallback.
func (s *CombatSystem) handleEntityDeath(entity *Entity) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok || !health.IsDead() {
		return
	}

	// Entity already has a DeadComponent from a previous invocation or an
	// external source (e.g. save-file loading).  Skip entirely.
	if entity.HasComponent("dead") {
		return
	}

	// Internal guard: prevent re-invocation on subsequent frames before the
	// entity is removed from the world.
	if _, alreadyFired := s.processedDeaths[entity.ID]; alreadyFired {
		return
	}
	s.processedDeaths[entity.ID] = struct{}{}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"entityID":      entity.ID,
			"currentHealth": health.Current,
		}).Info("entity death")
	}

	// Invoke callback first.  The callback (cmd/client/client_loot.go:
	// createDeathCallback) is the authoritative source for loot, XP, and
	// adding DeadComponent in one atomic transaction.
	if s.onDeathCallback != nil {
		s.onDeathCallback(entity)
	}

	// Fallback: if there is no callback, or the callback did not add
	// DeadComponent, add it here to ensure the entity is marked dead and the
	// guard above fires on subsequent frames.
	if !entity.HasComponent("dead") {
		entity.AddComponent(NewDeadComponent(0.0))
	}
}

// applyStatusEffectTick applies periodic status effect damage/healing.
func (s *CombatSystem) applyStatusEffectTick(entity *Entity, effect *StatusEffectComponent) {
	healthComp, ok := entity.GetComponent("health")
	if !ok {
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	switch effect.EffectType {
	case "poison", "burn":
		health.TakeDamage(effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":    entity.ID,
				"effect_type":  effect.EffectType,
				"damage":       effect.Magnitude,
				"health_after": health.Current,
			}).Debug("status effect damage applied")
		}
	case "regeneration":
		health.Heal(effect.Magnitude)
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":    entity.ID,
				"effect_type":  effect.EffectType,
				"healing":      effect.Magnitude,
				"health_after": health.Current,
			}).Debug("status effect healing applied")
		}
	}
}

// validateAttackEntities checks if attacker and target entities are in a valid state for combat.
func (s *CombatSystem) validateAttackEntities(attacker, target *Entity) bool {
	// G30 fix: an entity cannot damage itself.
	if attacker.ID == target.ID {
		if s.logger != nil {
			s.logger.WithField("entity_id", attacker.ID).Debug("attack blocked - self-attack")
		}
		return false
	}

	// Priority 1.3: Dead entities cannot attack
	if attacker.HasComponent("dead") {
		if s.logger != nil {
			s.logger.WithField("entity_id", attacker.ID).Debug("attack blocked - attacker is dead")
		}
		return false
	}

	// Priority 1.3: Dead entities cannot be targeted for attacks
	if target.HasComponent("dead") {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"attacker_id": attacker.ID,
				"target_id":   target.ID,
			}).Debug("attack blocked - target is dead")
		}
		return false
	}
	return true
}

// getAttackComponent retrieves and validates the attack component from an entity.
func (s *CombatSystem) getAttackComponent(attacker *Entity) (*AttackComponent, bool) {
	attackComp, ok := attacker.GetComponent("attack")
	if !ok {
		if s.logger != nil {
			s.logger.WithField("entity_id", attacker.ID).Debug("attack blocked - no attack component")
		}
		return nil, false
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("entity_id", attacker.ID).Error("attack component type assertion failed")
		}
		return nil, false
	}
	return attack, true
}

// validateAttackRange checks if the target is within attack range.
func (s *CombatSystem) validateAttackRange(attacker, target *Entity, attackRange float64) bool {
	_, attackerHasPos := attacker.GetComponent("position")
	_, targetHasPos := target.GetComponent("position")
	if attackerHasPos && targetHasPos {
		distance := GetDistance(attacker, target)
		if distance > attackRange {
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"attacker_id": attacker.ID,
					"target_id":   target.ID,
					"distance":    distance,
					"range":       attackRange,
				}).Debug("attack blocked - target out of range")
			}
			return false
		}
	}
	return true
}

// tryProjectileAttack attempts to perform a projectile-based attack if the attacker has a ranged weapon.
// Returns (success, isProjectile) where isProjectile indicates if this was a projectile attack attempt.
func (s *CombatSystem) tryProjectileAttack(attacker, target *Entity, attack *AttackComponent) (bool, bool) {
	// Phase 10.2: Check if attacker has a projectile weapon equipped
	// If so, spawn a projectile instead of doing instant damage
	if equipComp, hasEquip := attacker.GetComponent("equipment"); hasEquip {
		if equipment, ok := equipComp.(*EquipmentComponent); ok {
			if weapon, hasWeapon := equipment.Slots[SlotMainHand]; hasWeapon && weapon != nil {
				if weapon.Stats.IsProjectile {
					if s.logger != nil {
						s.logger.WithFields(logrus.Fields{
							"attacker_id":     attacker.ID,
							"target_id":       target.ID,
							"weapon_name":     weapon.Name,
							"projectile_type": weapon.Stats.ProjectileType,
						}).Debug("attempting projectile attack")
					}
					// Spawn projectile for ranged weapon
					success := s.spawnProjectile(attacker, target, weapon, attack)
					if success {
						attack.ResetCooldown()
					}
					return success, true
				}
			}
		}
	}
	return false, false
}

// getTargetHealth retrieves and validates the health component from the target entity.
func (s *CombatSystem) getTargetHealth(target *Entity) (*HealthComponent, bool) {
	targetHealth, ok := target.GetComponent("health")
	if !ok {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Debug("target has no health component")
		}
		return nil, false
	}
	health, ok := targetHealth.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Error("health component type assertion failed")
		}
		return nil, false
	}

	// Check if target is already dead
	if health.IsDead() {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Debug("target is already dead")
		}
		return nil, false
	}
	return health, true
}

// Attack performs an attack from attacker to target.
// Returns true if the attack hit, false if it missed or was invalid.
func (s *CombatSystem) Attack(attacker, target *Entity) bool {
	s.logAttackInitiated(attacker, target)

	if !s.validateAttackEntities(attacker, target) {
		return false
	}

	attack, ok := s.getAttackComponent(attacker)
	if !ok {
		return false
	}

	if !s.checkAttackCooldown(attacker, attack) {
		return false
	}

	if success, isProjectile := s.tryProjectileAttack(attacker, target, attack); isProjectile {
		return success
	}

	if !s.validateAttackRange(attacker, target, attack.Range) {
		return false
	}

	return s.executeMeleeAttack(attacker, target, attack)
}

// checkAttackCooldown validates if the attack is off cooldown.
func (s *CombatSystem) checkAttackCooldown(attacker *Entity, attack *AttackComponent) bool {
	if !attack.CanAttack() {
		s.logCooldownBlocked(attacker, attack)
		return false
	}
	return true
}

// executeMeleeAttack executes a melee attack after validation.
func (s *CombatSystem) executeMeleeAttack(attacker, target *Entity, attack *AttackComponent) bool {
	health, ok := s.getTargetHealth(target)
	if !ok {
		return false
	}

	attackerStats := s.getEntityStats(attacker)
	targetStats := s.getEntityStats(target)

	if s.checkEvasion(attacker, target, targetStats) {
		attack.ResetCooldown()
		return false
	}

	// Check for block - reduces damage by 50% if successful
	blocked := s.checkBlock(attacker, target, targetStats)

	finalDamage, baseDamage, isCrit := s.computeFinalDamage(attacker, attack, attackerStats, targetStats, target)
	if finalDamage <= 0 {
		attack.ResetCooldown()
		s.logDamageAbsorbed(attacker, target)
		return true
	}

	// Apply block damage reduction
	originalDamage := finalDamage
	if blocked {
		finalDamage *= 0.5
		if s.onBlockCallback != nil && targetStats != nil {
			s.onBlockCallback(attacker, target, targetStats.BlockChance, originalDamage, finalDamage)
		}
	}

	s.applyDamageAndFeedback(attacker, target, health, attack, finalDamage, baseDamage, isCrit)
	return true
}

// computeFinalDamage calculates the final damage after all modifiers.
// Returns (finalDamage, baseDamage, isCrit) to avoid redundant RNG-consuming recalculation.
func (s *CombatSystem) computeFinalDamage(attacker *Entity, attack *AttackComponent, attackerStats, targetStats *StatsComponent, target *Entity) (float64, float64, bool) {
	baseDamage, isCrit := s.calculateDamage(attack, attackerStats)

	// G34: apply equipment set bonus damage from attacker.
	baseDamage += s.getEquipmentSetDamageBonus(attacker)

	// Apply mod rule damage multipliers (Phase 6.3: Modding System Integration)
	baseDamage = s.applyModDamageMultipliers(attacker, target, baseDamage)

	// G34: boost effective defense with target's equipment set bonus before resist.
	effectiveTargetStats := s.applyEquipmentSetDefenseBonus(target, targetStats)

	damageAfterResist := s.applyDefenseAndResistance(baseDamage, attack.DamageType, effectiveTargetStats)

	// Check if significant damage was resisted and trigger callback
	if s.onDamageResistedCallback != nil && targetStats != nil {
		resistance := targetStats.GetResistance(attack.DamageType)
		if resistance > 0 && baseDamage > damageAfterResist {
			s.onDamageResistedCallback(target, attack.DamageType, baseDamage, damageAfterResist, resistance)
		}
	}

	finalDamage := damageAfterResist

	// G35 fix: apply shield absorption before the minimum-damage floor so that a
	// fully-charged shield can reduce damage to 0. The floor only guards against
	// near-zero resist math; it must not bypass intentional full-block shields.
	finalDamage = s.applyShieldAbsorption(target, finalDamage)
	if finalDamage < 1.0 && finalDamage > 0 {
		finalDamage = 1.0
	}

	return finalDamage, baseDamage, isCrit
}

// getEquipmentSetDamageBonus returns the additional flat damage from the
// attacker's active equipment set bonuses (G34).
func (s *CombatSystem) getEquipmentSetDamageBonus(attacker *Entity) float64 {
	comp, ok := attacker.GetComponent("equipment_set_bonus")
	if !ok {
		return 0
	}
	setBonus, ok := comp.(*EquipmentSetBonusComponent)
	if !ok {
		return 0
	}
	return float64(setBonus.GetTotalDamageBonus())
}

// applyEquipmentSetDefenseBonus returns a copy of the targetStats with the
// target's equipment set defense bonus folded in. If the target has no set
// bonus component or targetStats is nil, the original pointer is returned
// unchanged (G34).
func (s *CombatSystem) applyEquipmentSetDefenseBonus(target *Entity, targetStats *StatsComponent) *StatsComponent {
	if targetStats == nil {
		return targetStats
	}
	comp, ok := target.GetComponent("equipment_set_bonus")
	if !ok {
		return targetStats
	}
	setBonus, ok := comp.(*EquipmentSetBonusComponent)
	if !ok {
		return targetStats
	}
	defenseBonus := float64(setBonus.GetTotalDefenseBonus())
	if defenseBonus == 0 {
		return targetStats
	}
	// Return a shallow copy so that we do not mutate the live component.
	adjusted := *targetStats
	adjusted.Defense += defenseBonus
	return &adjusted
}

// applyDamageAndFeedback applies damage to target and triggers all feedback.
func (s *CombatSystem) applyDamageAndFeedback(attacker, target *Entity, health *HealthComponent, attack *AttackComponent, finalDamage, baseDamage float64, isCrit bool) {
	health.TakeDamage(finalDamage)
	s.applyAttackFeedback(attacker, target, finalDamage, baseDamage, attack.DamageType, isCrit, health.Current)
	attack.ResetCooldown()

	if s.onDamageCallback != nil {
		s.onDamageCallback(attacker, target, finalDamage)
	}

	// Trigger additional damage callbacks
	for _, cb := range s.additionalDamageCallbacks {
		cb(attacker, target, finalDamage)
	}

	// Trigger critical hit callback for visual effects
	if isCrit && s.onCriticalHitCallback != nil {
		s.onCriticalHitCallback(attacker, target, finalDamage)
	}

	// Trigger additional critical hit callbacks
	if isCrit {
		for _, cb := range s.additionalCriticalHitCallbacks {
			cb(attacker, target, finalDamage)
		}
	}

	// Check if target died from this attack and trigger kill callback
	if health.IsDead() && s.onKillCallback != nil {
		s.onKillCallback(attacker, target)
	}
}

// logAttackInitiated logs the initiation of an attack.
func (s *CombatSystem) logAttackInitiated(attacker, target *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id": attacker.ID,
			"target_id":   target.ID,
		}).Debug("attack initiated")
	}
}

// logCooldownBlocked logs when an attack is blocked by cooldown.
func (s *CombatSystem) logCooldownBlocked(attacker *Entity, attack *AttackComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id":    attacker.ID,
			"cooldown_timer": attack.CooldownTimer,
		}).Debug("attack blocked - cooldown active")
	}
}

// logDamageAbsorbed logs when damage is fully absorbed by shield.
func (s *CombatSystem) logDamageAbsorbed(attacker, target *Entity) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"attacker_id": attacker.ID,
			"target_id":   target.ID,
		}).Debug("attack damage fully absorbed by shield")
	}
}

// applyAttackFeedback triggers all visual and audio feedback for an attack.
func (s *CombatSystem) applyAttackFeedback(attacker, target *Entity, finalDamage, baseDamage float64, damageType combat.DamageType, isCrit bool, targetHealth float64) {
	s.triggerAttackAnimation(attacker)
	s.triggerHurtAnimation(target)
	s.logDamageEvent(attacker, target, finalDamage, baseDamage, damageType, isCrit, targetHealth)
	s.spawnHitParticles(target)
	s.applyVisualFeedback(target, finalDamage)
	s.triggerScreenShake(target, finalDamage, isCrit)
	s.playCombatSFX(target, damageType, isCrit)
}

// playCombatSFX plays combat sound effects for damage events (Plan Phase 1.1).
func (s *CombatSystem) playCombatSFX(target *Entity, damageType combat.DamageType, isCrit bool) {
	if s.audioManager == nil {
		return
	}

	// Select effect type based on damage type
	var effectType string
	switch damageType {
	case combat.DamageMagical:
		effectType = "magic"
	case combat.DamageFire, combat.DamageIce, combat.DamageLightning, combat.DamagePoison:
		effectType = "magic"
	default:
		effectType = "hit"
	}

	// Use stronger impact sound for critical hits
	if isCrit {
		effectType = "impact"
	}

	// Use target entity ID as seed for deterministic but varied sounds
	if err := s.audioManager.PlaySFX(effectType, int64(target.ID)); err != nil {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithError(err).WithField("effectType", effectType).Debug("failed to play combat SFX")
		}
	}
}

// getEntityStats retrieves the stats component from an entity, returns nil if not present.
func (s *CombatSystem) getEntityStats(entity *Entity) *StatsComponent {
	statsComp, ok := entity.GetComponent("stats")
	if !ok || statsComp == nil {
		return nil
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return nil
	}
	return stats
}

// checkEvasion determines if an attack is evaded based on target evasion stats.
func (s *CombatSystem) checkEvasion(attacker, target *Entity, targetStats *StatsComponent) bool {
	if targetStats != nil && s.rollChance(targetStats.Evasion) {
		// Attack missed
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"attackerID": attacker.ID,
				"targetID":   target.ID,
				"evasion":    targetStats.Evasion,
			}).Debug("attack evaded")
		}
		// Trigger evasion callback for visual effects
		if s.onEvasionCallback != nil {
			s.onEvasionCallback(attacker, target, targetStats.Evasion)
		}
		return true
	}
	return false
}

// checkBlock determines if an attack is blocked (damage reduced) based on target block stats.
func (s *CombatSystem) checkBlock(attacker, target *Entity, targetStats *StatsComponent) bool {
	if targetStats != nil && targetStats.BlockChance > 0 && s.rollChance(targetStats.BlockChance) {
		if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
			s.logger.WithFields(logrus.Fields{
				"attackerID":  attacker.ID,
				"targetID":    target.ID,
				"blockChance": targetStats.BlockChance,
			}).Debug("attack blocked")
		}
		return true
	}
	return false
}

// calculateDamage computes the base damage and determines if the attack is a critical hit.
func (s *CombatSystem) calculateDamage(attack *AttackComponent, attackerStats *StatsComponent) (damage float64, isCrit bool) {
	baseDamage := attack.Damage
	isCrit = false

	// Apply attacker stats
	if attackerStats != nil {
		if attack.DamageType == combat.DamageMagical {
			baseDamage += attackerStats.MagicPower
		} else {
			baseDamage += attackerStats.Attack
		}

		// Check for critical hit
		if s.rollChance(attackerStats.CritChance) {
			baseDamage *= attackerStats.CritDamage
			isCrit = true
			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"base_damage":     attack.Damage,
					"crit_damage":     baseDamage,
					"crit_multiplier": attackerStats.CritDamage,
				}).Debug("critical hit")
			}
		}
	}

	return baseDamage, isCrit
}

// applyModDamageMultipliers applies mod-defined damage multipliers to base damage.
// Queries mod rules based on entity type (player vs NPC).
// Phase 6.3 (PLAN.md): Modding System Integration
func (s *CombatSystem) applyModDamageMultipliers(attacker, target *Entity, baseDamage float64) float64 {
	if s.world == nil {
		return baseDamage
	}

	modRules := s.world.GetModRules()
	if modRules == nil {
		return baseDamage
	}

	// Apply attacker damage multiplier
	attackerMultiplier := 1.0
	if attacker.HasComponent("input") {
		// Player attacker: query player damage multiplier
		attackerMultiplier = modRules.GetRuleFloat64("combat.player_damage_multiplier", 1.0)
	} else if attacker.HasComponent("ai") || attacker.HasComponent("npc") {
		// Enemy attacker: query enemy damage multiplier
		attackerMultiplier = modRules.GetRuleFloat64("combat.enemy_damage_multiplier", 1.0)
	}

	finalDamage := baseDamage * attackerMultiplier

	if s.logger != nil && attackerMultiplier != 1.0 {
		s.logger.WithFields(logrus.Fields{
			"attacker_id":        attacker.ID,
			"base_damage":        baseDamage,
			"final_damage":       finalDamage,
			"damage_multiplier":  attackerMultiplier,
			"is_player_attacker": attacker.HasComponent("input"),
		}).Debug("mod damage multiplier applied")
	}

	return finalDamage
}

// applyDefenseAndResistance reduces damage based on target's defense and resistances.
// Uses the authoritative combat.CombatResolver for consistent damage formulas across the codebase.
func (s *CombatSystem) applyDefenseAndResistance(baseDamage float64, damageType combat.DamageType, targetStats *StatsComponent) float64 {
	if targetStats == nil {
		return baseDamage
	}

	// Convert StatsComponent to combat.Stats for resolver
	stats := statsComponentToCombatStats(targetStats)

	// Use combat package's authoritative damage calculation
	damage := combat.Damage{
		Amount: baseDamage,
		Type:   damageType,
	}

	return s.combatResolver.CalculateDamage(damage, &stats)
}

// statsComponentToCombatStats converts engine StatsComponent to combat.Stats.
// This allows us to use the combat package's authoritative damage formulas.
func statsComponentToCombatStats(stats *StatsComponent) combat.Stats {
	return combat.Stats{
		Attack:       stats.Attack,
		MagicPower:   stats.MagicPower,
		Defense:      stats.Defense,
		MagicDefense: stats.MagicDefense,
		CritChance:   stats.CritChance,
		CritDamage:   stats.CritDamage,
		Evasion:      stats.Evasion,
		Resistances:  stats.Resistances,
	}
}

// applyShieldAbsorption reduces damage by shield absorption, returns remaining damage.
func (s *CombatSystem) applyShieldAbsorption(target *Entity, damage float64) float64 {
	finalDamage := damage
	if shieldComp, hasShield := target.GetComponent("shield"); hasShield {
		if shield, ok := shieldComp.(*ShieldComponent); ok {
			if shield.IsActive() {
				// Shield absorbs damage
				absorbed := shield.AbsorbDamage(finalDamage)
				finalDamage -= absorbed
				if s.logger != nil && absorbed > 0 {
					s.logger.WithFields(logrus.Fields{
						"target_id":        target.ID,
						"damage_absorbed":  absorbed,
						"damage_remaining": finalDamage,
						"shield_amount":    shield.Amount,
					}).Debug("shield absorbed damage")
				}
				// Trigger shield absorb callback for visual effects
				if absorbed > 0 && s.onShieldAbsorbCallback != nil {
					s.onShieldAbsorbCallback(target, absorbed, shield.Amount)
				}
			}
		}
	}
	return finalDamage
}

// triggerAttackAnimation triggers the attack animation for the attacker entity.
func (s *CombatSystem) triggerAttackAnimation(attacker *Entity) {
	if animComp, hasAnim := attacker.GetComponent("animation"); hasAnim {
		if anim, ok := animComp.(*AnimationComponent); ok {
			// Log animation trigger for player when debugging
			if attacker.HasComponent("input") && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
				s.logger.WithFields(logrus.Fields{
					"attackerID":    attacker.ID,
					"previousState": anim.CurrentState,
					"newState":      "ATTACK",
				}).Debug("player attack animation triggered")
			}

			anim.SetState(AnimationStateAttack)
			// Set callback to return to idle after attack animation completes
			anim.OnComplete = func() {
				// Check if entity is moving to set appropriate idle/walk state
				if velComp, hasVel := attacker.GetComponent("velocity"); hasVel {
					if vel, ok := velComp.(*VelocityComponent); ok {
						speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
						if speed > 0.1 {
							anim.SetState(AnimationStateWalk)
						} else {
							anim.SetState(AnimationStateIdle)
						}
					}
				} else {
					anim.SetState(AnimationStateIdle)
				}

				if attacker.HasComponent("input") && s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
					s.logger.WithField("attackerID", attacker.ID).Debug("player attack animation complete")
				}
			}
		}
	}
}

// triggerHurtAnimation triggers the hurt animation for the target entity.
func (s *CombatSystem) triggerHurtAnimation(target *Entity) {
	if animComp, hasAnim := target.GetComponent("animation"); hasAnim {
		if anim, ok := animComp.(*AnimationComponent); ok {
			anim.SetState(AnimationStateHit)
			// Set a callback to return to idle after hurt animation
			anim.OnComplete = func() {
				// Check if entity is moving to set appropriate idle/walk state
				if velComp, hasVel := target.GetComponent("velocity"); hasVel {
					if vel, ok := velComp.(*VelocityComponent); ok {
						speed := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
						if speed > 0.1 {
							anim.SetState(AnimationStateWalk)
						} else {
							anim.SetState(AnimationStateIdle)
						}
					}
				} else {
					anim.SetState(AnimationStateIdle)
				}
			}
		}
	}
}

// logDamageEvent logs information about the damage dealt.
func (s *CombatSystem) logDamageEvent(attacker, target *Entity, finalDamage, baseDamage float64, damageType combat.DamageType, isCrit bool, targetHealth float64) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.InfoLevel {
		s.logger.WithFields(logrus.Fields{
			"attackerID":   attacker.ID,
			"targetID":     target.ID,
			"damage":       finalDamage,
			"baseDamage":   baseDamage,
			"damageType":   damageType,
			"critical":     isCrit,
			"targetHealth": targetHealth,
		}).Info("damage dealt")
	}
}

// spawnHitParticles creates particle effects at the target's position.
func (s *CombatSystem) spawnHitParticles(target *Entity) {
	// GAP-016 REPAIR: Spawn hit particles at target position
	if s.particleSystem != nil && s.world != nil {
		if posComp, ok := target.GetComponent("position"); ok {
			if pos, ok := posComp.(*PositionComponent); ok {
				// Use timestamp for particle seed variation
				particleSeed := s.seed + int64(pos.X*1000) + int64(pos.Y*1000)
				s.particleSystem.SpawnHitSparks(s.world, pos.X, pos.Y, particleSeed, s.genreID)
				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"target_id": target.ID,
						"x":         pos.X,
						"y":         pos.Y,
						"genre_id":  s.genreID,
					}).Debug("hit particles spawned")
				}
			}
		}
	}
}

// applyVisualFeedback triggers visual hit flash on the target entity.
// Uses cached GetVisualFeedback() getter for ~93x faster access.
func (s *CombatSystem) applyVisualFeedback(target *Entity, finalDamage float64) {
	// GAP-012 REPAIR: Trigger hit flash on damage
	// Phase 10.3: Respects accessibility settings via camera
	feedback := target.GetVisualFeedback()
	if feedback == nil {
		return
	}

	// Check accessibility settings for visual flash
	if s.camera != nil && s.camera.Accessibility.ShouldApplyVisualFlash() {
		// Flash intensity scales with damage (0.3-1.0 range)
		flashIntensity := 0.3 + (finalDamage / 100.0)
		if flashIntensity > 1.0 {
			flashIntensity = 1.0
		}
		feedback.TriggerFlash(flashIntensity)
	}
}

// triggerScreenShake applies camera shake based on damage dealt.
func (s *CombatSystem) triggerScreenShake(target *Entity, finalDamage float64, isCrit bool) {
	// GAP-012 REPAIR: Trigger screen shake on damage
	// Phase 10.3: Enhanced shake with procedural intensity and duration
	if s.camera != nil {
		// Use advanced shake if available (with duration control)
		targetHealthComp, _ := target.GetComponent("health")
		var maxHP float64 = 100 // Default
		if targetHealthComp != nil {
			if health, ok := targetHealthComp.(*HealthComponent); ok {
				maxHP = health.Max
			}
		}

		// Calculate shake intensity based on damage relative to max HP
		shakeIntensity := CalculateShakeIntensity(finalDamage, maxHP,
			CombatShakeScaleFactor, CombatShakeMinIntensity, CombatShakeMaxIntensity)
		shakeDuration := CalculateShakeDuration(shakeIntensity,
			CombatShakeBaseDuration, CombatShakeAdditionalDuration, CombatShakeMaxIntensity)

		// Critical hits get extra shake and hit-stop
		if isCrit {
			shakeIntensity *= CriticalHitShakeMultiplier
			shakeDuration *= CriticalHitDurationMultiplier

			// Trigger hit-stop on critical hits
			s.camera.TriggerHitStop(CriticalHitStopDuration, 0.0)
		}

		// Use advanced shake if available, fallback to basic
		s.camera.ShakeAdvanced(shakeIntensity, shakeDuration)
	}
}

// rollChance returns true if a random roll succeeds based on the given chance (0.0 to 1.0).
func (s *CombatSystem) rollChance(chance float64) bool {
	if chance <= 0 {
		return false
	}
	if chance >= 1.0 {
		return true
	}
	return s.rng.Float64() < chance
}

// CanAttackTarget checks if an attacker can attack a target (range and cooldown check).
func (s *CombatSystem) CanAttackTarget(attacker, target *Entity) bool {
	attackComp, ok := attacker.GetComponent("attack")
	if !ok {
		return false
	}
	attack, ok := attackComp.(*AttackComponent)
	if !ok {
		return false
	}

	if !attack.CanAttack() {
		return false
	}

	targetHealth, ok := target.GetComponent("health")
	if !ok {
		return false
	}
	if health, ok := targetHealth.(*HealthComponent); ok {
		if health.IsDead() {
			return false
		}
	} else {
		return false
	}

	// Check range if both have positions
	_, attackerHasPos := attacker.GetComponent("position")
	_, targetHasPos := target.GetComponent("position")
	if attackerHasPos && targetHasPos {
		distance := GetDistance(attacker, target)
		if distance > attack.Range {
			return false
		}
	}

	return true
}

// ApplyStatusEffect applies a status effect to an entity.
func (s *CombatSystem) ApplyStatusEffect(target *Entity, effectType string, duration, magnitude, tickInterval float64) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":     target.ID,
			"effect_type":   effectType,
			"duration":      duration,
			"magnitude":     magnitude,
			"tick_interval": tickInterval,
		}).Debug("applying status effect")
	}

	// Use pooled status effect to reduce GC pressure
	effect := NewStatusEffectComponent(effectType, magnitude, duration, tickInterval)

	// Replace any existing status effect (simplification)
	target.AddComponent(effect)
}

// Heal heals a target entity by the given amount.
func (s *CombatSystem) Heal(target *Entity, amount float64) {
	s.HealWithHealer(nil, target, amount)
}

// HealWithHealer heals a target entity by the given amount, tracking the healer.
// If healer is nil, it is treated as environmental healing or self-heal.
func (s *CombatSystem) HealWithHealer(healer, target *Entity, amount float64) {
	if target == nil || amount <= 0 {
		return
	}

	healthComp, ok := target.GetComponent("health")
	if !ok {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Debug("heal failed - no health component")
		}
		return
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Error("health component type assertion failed in heal")
		}
		return
	}

	beforeHealth := health.Current
	health.Heal(amount)
	actualHeal := health.Current - beforeHealth

	// Invoke heal callback if meaningful healing occurred
	if actualHeal > 0 && s.onHealCallback != nil {
		s.onHealCallback(healer, target, actualHeal)
	}

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"target_id":     target.ID,
			"heal_amount":   amount,
			"actual_healed": actualHeal,
			"health_before": beforeHealth,
			"health_after":  health.Current,
		}).Debug("entity healed")
	}
}

// SetDeathCallback sets the callback function for entity deaths.
func (s *CombatSystem) SetDeathCallback(callback func(entity *Entity)) {
	if s.logger != nil {
		s.logger.Debug("death callback registered")
	}
	s.onDeathCallback = callback
}

// SetKillCallback sets the callback function for kills (includes attacker).
// Called when an entity is killed by an attack, with both attacker and target info.
func (s *CombatSystem) SetKillCallback(callback func(attacker, target *Entity)) {
	if s.logger != nil {
		s.logger.Debug("kill callback registered")
	}
	s.onKillCallback = callback
}

// SetDamageCallback sets the callback function for damage dealt.
func (s *CombatSystem) SetDamageCallback(callback func(attacker, target *Entity, damage float64)) {
	if s.logger != nil {
		s.logger.Debug("damage callback registered")
	}
	s.onDamageCallback = callback
}

// AddDamageCallback adds an additional callback for damage events without replacing existing ones.
// Use this for systems that need to react to damage but shouldn't override the primary callback.
func (s *CombatSystem) AddDamageCallback(callback func(attacker, target *Entity, damage float64)) {
	if s.logger != nil {
		s.logger.Debug("additional damage callback registered")
	}
	s.additionalDamageCallbacks = append(s.additionalDamageCallbacks, callback)
}

// SetCriticalHitCallback sets the callback function for critical hits.
func (s *CombatSystem) SetCriticalHitCallback(callback func(attacker, target *Entity, damage float64)) {
	if s.logger != nil {
		s.logger.Debug("critical hit callback registered")
	}
	s.onCriticalHitCallback = callback
}

// AddCriticalHitCallback adds an additional callback for critical hit events without replacing existing ones.
// Use this for systems that need to react to critical hits but shouldn't override the primary callback.
func (s *CombatSystem) AddCriticalHitCallback(callback func(attacker, target *Entity, damage float64)) {
	if s.logger != nil {
		s.logger.Debug("additional critical hit callback registered")
	}
	s.additionalCriticalHitCallbacks = append(s.additionalCriticalHitCallbacks, callback)
}

// SetDamageResistedCallback sets the callback function for when damage is resisted.
// The callback receives the target entity, damage type, original damage before resistance,
// final damage after resistance, and the resistance value (0.0-1.0).
func (s *CombatSystem) SetDamageResistedCallback(callback func(target *Entity, damageType combat.DamageType, originalDamage, finalDamage, resistance float64)) {
	if s.logger != nil {
		s.logger.Debug("damage resisted callback registered")
	}
	s.onDamageResistedCallback = callback
}

// SetShieldAbsorbCallback sets the callback function for when a shield absorbs damage.
// The callback receives the target entity, amount absorbed, and remaining shield amount.
func (s *CombatSystem) SetShieldAbsorbCallback(callback func(target *Entity, absorbed, remaining float64)) {
	if s.logger != nil {
		s.logger.Debug("shield absorb callback registered")
	}
	s.onShieldAbsorbCallback = callback
}

// SetEvasionCallback sets the callback function for when an attack is evaded.
// The callback receives the attacker, target, and the evasion chance that succeeded.
func (s *CombatSystem) SetEvasionCallback(callback func(attacker, target *Entity, evasionChance float64)) {
	if s.logger != nil {
		s.logger.Debug("evasion callback registered")
	}
	s.onEvasionCallback = callback
}

// SetHealCallback sets the callback function for when an entity is healed.
// The callback receives the healer (may be nil), target, and actual amount healed.
func (s *CombatSystem) SetHealCallback(callback func(healer, target *Entity, amount float64)) {
	if s.logger != nil {
		s.logger.Debug("heal callback registered")
	}
	s.onHealCallback = callback
}

// SetBlockCallback sets the callback function for when an attack is blocked.
// The callback receives the attacker, target, block chance, original damage, and reduced damage.
func (s *CombatSystem) SetBlockCallback(callback func(attacker, target *Entity, blockChance, originalDamage, reducedDamage float64)) {
	if s.logger != nil {
		s.logger.Debug("block callback registered")
	}
	s.onBlockCallback = callback
}

// isValidEnemyTarget checks if entity is a valid enemy target.
func isValidEnemyTarget(entity *Entity, attackerTeamID int) bool {
	if entity.HasComponent("dead") {
		return false
	}

	targetTeam, hasTeam := entity.GetComponent("team")
	if hasTeam {
		if team, ok := targetTeam.(*TeamComponent); ok {
			if !team.IsEnemy(attackerTeamID) {
				return false
			}
		}
	}

	healthComp, hasHealth := entity.GetComponent("health")
	if !hasHealth {
		return false
	}
	if health, ok := healthComp.(*HealthComponent); ok {
		if health.IsDead() {
			return false
		}
	} else {
		return false
	}

	return true
}

// FindEnemiesInRange finds all enemy entities within the given range of the attacker.
func FindEnemiesInRange(world *World, attacker *Entity, maxRange float64) []*Entity {
	_, ok := attacker.GetComponent("position")
	if !ok {
		return nil
	}

	attackerTeam, _ := attacker.GetComponent("team")
	var attackerTeamID int
	if attackerTeam != nil {
		if team, ok := attackerTeam.(*TeamComponent); ok {
			attackerTeamID = team.TeamID
		}
	}

	enemies := make([]*Entity, 0, 16)

	for _, entity := range world.GetEntities() {
		if entity.ID == attacker.ID {
			continue
		}

		if !isValidEnemyTarget(entity, attackerTeamID) {
			continue
		}

		_, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		distance := GetDistance(attacker, entity)
		if distance <= maxRange {
			enemies = append(enemies, entity)
		}
	}

	return enemies
}

// FindNearestEnemy finds the closest enemy to the attacker within the given range.
func FindNearestEnemy(world *World, attacker *Entity, maxRange float64) *Entity {
	enemies := FindEnemiesInRange(world, attacker, maxRange)
	if len(enemies) == 0 {
		return nil
	}

	var nearest *Entity
	nearestDistance := math.MaxFloat64

	for _, enemy := range enemies {
		distance := GetDistance(attacker, enemy)
		if distance < nearestDistance {
			nearestDistance = distance
			nearest = enemy
		}
	}

	return nearest
}

// FindEnemyInAimDirection finds an enemy in the aim direction within attack range.
// Phase 10.1: Uses AimComponent to determine attack direction for dual-stick shooter mechanics.
// aimAngle: aim direction in radians (0 = right, π/2 = down, π = left, 3π/2 = up)
// maxRange: maximum attack range
// aimCone: angle cone in radians (e.g., π/4 = 45° cone for forgiving aim)
// Returns the closest enemy within the aim cone, or nil if none found.
func FindEnemyInAimDirection(world *World, attacker *Entity, aimAngle, maxRange, aimCone float64) *Entity {
	enemies := FindEnemiesInRange(world, attacker, maxRange)
	if len(enemies) == 0 {
		return nil
	}

	attackerPos := getAttackerPosition(attacker)
	if attackerPos == nil {
		return nil
	}

	return findBestEnemyInAimCone(enemies, attackerPos, aimAngle, aimCone)
}

// getAttackerPosition retrieves the position component from an entity.
func getAttackerPosition(attacker *Entity) *PositionComponent {
	attackerPosComp, hasPos := attacker.GetComponent("position")
	if !hasPos {
		return nil
	}

	pos, ok := attackerPosComp.(*PositionComponent)
	if !ok {
		return nil
	}

	return pos
}

// findBestEnemyInAimCone finds the closest enemy within the aim cone.
func findBestEnemyInAimCone(enemies []*Entity, attackerPos *PositionComponent, aimAngle, aimCone float64) *Entity {
	var bestEnemy *Entity
	bestDistanceSquared := math.MaxFloat64

	for _, enemy := range enemies {
		enemyPos := getAttackerPosition(enemy)
		if enemyPos == nil {
			continue
		}

		if isEnemyInAimCone(attackerPos, enemyPos, aimAngle, aimCone, &bestDistanceSquared, &bestEnemy, enemy) {
			// Enemy was updated in isEnemyInAimCone if it's closer
		}
	}

	return bestEnemy
}

// isEnemyInAimCone checks if enemy is within aim cone and updates best enemy if closer.
func isEnemyInAimCone(attackerPos, enemyPos *PositionComponent, aimAngle, aimCone float64, bestDistSquared *float64, bestEnemy **Entity, enemy *Entity) bool {
	dx := enemyPos.X - attackerPos.X
	dy := enemyPos.Y - attackerPos.Y
	angleToEnemy := math.Atan2(dy, dx)

	angleDiff := normalizeAngleDifference(angleToEnemy - aimAngle)

	if math.Abs(angleDiff) <= aimCone/2 {
		distanceSquared := dx*dx + dy*dy
		if distanceSquared < *bestDistSquared {
			*bestDistSquared = distanceSquared
			*bestEnemy = enemy
			return true
		}
	}
	return false
}

// normalizeAngleDifference normalizes an angle difference to the range [-π, π].
func normalizeAngleDifference(angleDiff float64) float64 {
	for angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	}
	for angleDiff < -math.Pi {
		angleDiff += 2 * math.Pi
	}
	return angleDiff
}

// spawnProjectile creates a projectile entity for ranged weapon attacks (Phase 10.2).
// Returns true if projectile was successfully spawned, false otherwise.
func (s *CombatSystem) spawnProjectile(attacker, target *Entity, weapon *item.Item, attack *AttackComponent) bool {
	if s.projectileSystem == nil || s.world == nil {
		if s.logger != nil {
			s.logger.WithField("attacker_id", attacker.ID).Warn("projectile spawn failed - system not initialized")
		}
		return false
	}

	attackerPos, ok := s.getAttackerPosition(attacker)
	if !ok {
		return false
	}

	aimAngle, ok := s.calculateAimAngle(attacker, target, attackerPos)
	if !ok {
		return false
	}

	spawnX, spawnY := s.calculateSpawnPosition(attackerPos, aimAngle)
	velocityX, velocityY := s.calculateProjectileVelocity(weapon, aimAngle)
	baseDamage := s.calculateProjectileDamage(attacker, attack)
	projComp := s.createProjectileComponent(weapon, baseDamage, attacker.ID)

	projectile := s.spawnProjectileEntity(spawnX, spawnY, velocityX, velocityY, aimAngle, projComp)
	s.addProjectileSprite(projectile, projComp, aimAngle)
	s.logProjectileSpawn(attacker, projectile, baseDamage, projComp)

	return true
}

// getAttackerPosition retrieves the position component from the attacker entity.
func (s *CombatSystem) getAttackerPosition(attacker *Entity) (*PositionComponent, bool) {
	attackerPosComp, hasPos := attacker.GetComponent("position")
	if !hasPos {
		if s.logger != nil {
			s.logger.WithField("attacker_id", attacker.ID).Debug("projectile spawn failed - attacker has no position")
		}
		return nil, false
	}
	attackerPos, ok := attackerPosComp.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("attacker_id", attacker.ID).Error("position component type assertion failed")
		}
		return nil, false
	}
	return attackerPos, true
}

// calculateAimAngle determines the aim angle from aim component, rotation component, or target position.
func (s *CombatSystem) calculateAimAngle(attacker, target *Entity, attackerPos *PositionComponent) (float64, bool) {
	if aimComp, hasAim := attacker.GetComponent("aim"); hasAim {
		if aim, ok := aimComp.(*AimComponent); ok {
			return aim.AimAngle, true
		}
	}

	if rotComp, hasRot := attacker.GetComponent("rotation"); hasRot {
		if rot, ok := rotComp.(*RotationComponent); ok {
			return rot.Angle, true
		}
	}

	return s.calculateAngleToTarget(target, attackerPos)
}

// calculateAngleToTarget calculates the angle from attacker to target position.
func (s *CombatSystem) calculateAngleToTarget(target *Entity, attackerPos *PositionComponent) (float64, bool) {
	targetPosComp, hasTargetPos := target.GetComponent("position")
	if !hasTargetPos {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Debug("cannot calculate aim angle - target has no position")
		}
		return 0, false
	}
	targetPos, ok := targetPosComp.(*PositionComponent)
	if !ok {
		if s.logger != nil {
			s.logger.WithField("target_id", target.ID).Error("target position component type assertion failed")
		}
		return 0, false
	}
	dx := targetPos.X - attackerPos.X
	dy := targetPos.Y - attackerPos.Y
	return math.Atan2(dy, dx), true
}

// calculateSpawnPosition computes the projectile spawn position offset from the attacker.
func (s *CombatSystem) calculateSpawnPosition(attackerPos *PositionComponent, aimAngle float64) (float64, float64) {
	spawnOffset := 20.0
	spawnX := attackerPos.X + math.Cos(aimAngle)*spawnOffset
	spawnY := attackerPos.Y + math.Sin(aimAngle)*spawnOffset
	return spawnX, spawnY
}

// calculateProjectileVelocity computes the projectile velocity components from weapon speed and aim angle.
func (s *CombatSystem) calculateProjectileVelocity(weapon *item.Item, aimAngle float64) (float64, float64) {
	speed := weapon.Stats.ProjectileSpeed
	if speed <= 0 {
		speed = 400.0
	}
	velocityX := math.Cos(aimAngle) * speed
	velocityY := math.Sin(aimAngle) * speed
	return velocityX, velocityY
}

// calculateProjectileDamage computes the projectile damage including attacker stats and critical hits.
func (s *CombatSystem) calculateProjectileDamage(attacker *Entity, attack *AttackComponent) float64 {
	baseDamage := attack.Damage

	attackerStatsComp, hasStats := attacker.GetComponent("stats")
	if !hasStats {
		return baseDamage
	}
	attackerStats, ok := attackerStatsComp.(*StatsComponent)
	if !ok {
		return baseDamage
	}

	if attack.DamageType == combat.DamageMagical {
		baseDamage += attackerStats.MagicPower
	} else {
		baseDamage += attackerStats.Attack
	}

	if s.rollChance(attackerStats.CritChance) {
		baseDamage *= attackerStats.CritDamage
	}

	return baseDamage
}

// createProjectileComponent creates and configures a projectile component with weapon properties.
func (s *CombatSystem) createProjectileComponent(weapon *item.Item, baseDamage float64, attackerID uint64) *ProjectileComponent {
	lifetime := weapon.Stats.ProjectileLifetime
	if lifetime <= 0 {
		lifetime = 3.0
	}

	projectileType := weapon.Stats.ProjectileType
	if projectileType == "" {
		projectileType = "arrow"
	}

	speed := weapon.Stats.ProjectileSpeed
	if speed <= 0 {
		speed = 400.0
	}

	projComp := NewProjectileComponent(baseDamage, speed, lifetime, projectileType, attackerID)
	projComp.Pierce = weapon.Stats.Pierce
	projComp.Bounce = weapon.Stats.Bounce
	projComp.Explosive = weapon.Stats.Explosive
	projComp.ExplosionRadius = weapon.Stats.ExplosionRadius

	return projComp
}

// spawnProjectileEntity creates the projectile entity with position, velocity, and projectile components.
func (s *CombatSystem) spawnProjectileEntity(spawnX, spawnY, velocityX, velocityY, aimAngle float64, projComp *ProjectileComponent) *Entity {
	projectile := s.world.CreateEntity()
	projectile.AddComponent(&PositionComponent{X: spawnX, Y: spawnY})
	projectile.AddComponent(&VelocityComponent{VX: velocityX, VY: velocityY})
	projectile.AddComponent(projComp)
	projectile.AddComponent(&RotationComponent{Angle: aimAngle})
	return projectile
}

// addProjectileSprite generates and adds the sprite component to the projectile entity.
func (s *CombatSystem) addProjectileSprite(projectile *Entity, projComp *ProjectileComponent, aimAngle float64) {
	spriteSize := 12
	if projComp.Explosive {
		spriteSize = 16
	}

	spriteSeed := s.seed + int64(projectile.ID)
	spriteImage := sprites.GenerateProjectileSprite(spriteSeed, projComp.ProjectileType, s.genreID, spriteSize)

	spriteComp := NewSpriteComponent(float64(spriteSize), float64(spriteSize), color.RGBA{255, 255, 255, 255})
	spriteComp.Image = spriteImage
	spriteComp.Rotation = aimAngle
	projectile.AddComponent(spriteComp)
}

// logProjectileSpawn logs debug information about the spawned projectile.
func (s *CombatSystem) logProjectileSpawn(attacker, projectile *Entity, baseDamage float64, projComp *ProjectileComponent) {
	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"attackerID":     attacker.ID,
			"projectileID":   projectile.ID,
			"damage":         baseDamage,
			"speed":          projComp.Speed,
			"projectileType": projComp.ProjectileType,
			"pierce":         projComp.Pierce,
			"bounce":         projComp.Bounce,
			"explosive":      projComp.Explosive,
		}).Debug("projectile spawned")
	}
}
