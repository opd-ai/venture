package engine

import (
	"image/color"
	"math"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/magic"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

// ManaComponent tracks entity's magical energy.
type ManaComponent struct {
	Current int
	Max     int
	Regen   float64 // Mana regenerated per second
}

// Type implements Component interface.
func (m *ManaComponent) Type() string {
	return "mana"
}

// SpellSlotComponent stores equipped spells in slots 1-5.
type SpellSlotComponent struct {
	Slots      [5]*magic.Spell
	Cooldowns  [5]float64 // Remaining cooldown time for each slot
	CastingBar float64    // Progress of current cast (0.0 to 1.0)
	Casting    int        // Which slot is being cast (-1 = none)
}

// Type implements Component interface.
func (s *SpellSlotComponent) Type() string {
	return "spell_slots"
}

// GetSlot returns the spell in the given slot (0-4), or nil if empty.
func (s *SpellSlotComponent) GetSlot(slot int) *magic.Spell {
	if slot < 0 || slot >= 5 {
		return nil
	}
	return s.Slots[slot]
}

// SetSlot assigns a spell to a slot.
func (s *SpellSlotComponent) SetSlot(slot int, spell *magic.Spell) {
	if slot >= 0 && slot < 5 {
		s.Slots[slot] = spell
	}
}

// IsOnCooldown returns true if the slot is on cooldown.
func (s *SpellSlotComponent) IsOnCooldown(slot int) bool {
	if slot < 0 || slot >= 5 {
		return true
	}
	return s.Cooldowns[slot] > 0
}

// IsCasting returns true if currently casting a spell.
func (s *SpellSlotComponent) IsCasting() bool {
	return s.Casting >= 0
}

// SpellCastingSystem handles spell execution and cooldowns.
type SpellCastingSystem struct {
	world           *World
	statusEffectSys *StatusEffectSystem
	particleSys     *ParticleSystem       // For visual effects
	audioMgr        *AudioManager         // For sound effects
	tutorialSys     *EbitenTutorialSystem // For notifications
}

// NewSpellCastingSystem creates a new spell casting system.
func NewSpellCastingSystem(world *World, statusEffectSys *StatusEffectSystem) *SpellCastingSystem {
	return &SpellCastingSystem{
		world:           world,
		statusEffectSys: statusEffectSys,
		particleSys:     NewParticleSystem(),
		audioMgr:        nil, // Will be set via SetAudioManager()
	}
}

// SetAudioManager sets the audio manager for sound effects.
// This allows deferred initialization when audio system is ready.
func (s *SpellCastingSystem) SetAudioManager(audioMgr *AudioManager) {
	s.audioMgr = audioMgr
}

// SetParticleSystem sets the particle system for visual effects.
// This allows using a shared particle system if desired.
func (s *SpellCastingSystem) SetParticleSystem(particleSys *ParticleSystem) {
	s.particleSys = particleSys
}

// SetTutorialSystem sets the tutorial system for notifications.
// This allows displaying feedback messages to the player.
func (s *SpellCastingSystem) SetTutorialSystem(tutorialSys *EbitenTutorialSystem) {
	s.tutorialSys = tutorialSys
}

// Update processes spell casting and cooldowns.
func (s *SpellCastingSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Only process entities with spell slots
		spellComp, hasSpells := entity.GetComponent("spell_slots")
		if !hasSpells {
			continue
		}
		slots := spellComp.(*SpellSlotComponent)

		// Update cooldowns
		for i := range slots.Cooldowns {
			if slots.Cooldowns[i] > 0 {
				slots.Cooldowns[i] -= deltaTime
				if slots.Cooldowns[i] < 0 {
					slots.Cooldowns[i] = 0
				}
			}
		}

		// Update casting progress
		if slots.IsCasting() {
			spell := slots.GetSlot(slots.Casting)
			if spell == nil {
				slots.Casting = -1
				slots.CastingBar = 0
				continue
			}

			// Advance casting progress
			if spell.Stats.CastTime > 0 {
				slots.CastingBar += deltaTime / spell.Stats.CastTime
			} else {
				// Instant cast
				slots.CastingBar = 1.0
			}

			// Complete cast when bar reaches 1.0
			if slots.CastingBar >= 1.0 {
				s.executeCast(entity, spell, slots.Casting)

				// Start cooldown
				slots.Cooldowns[slots.Casting] = spell.Stats.Cooldown

				// Reset casting state
				slots.Casting = -1
				slots.CastingBar = 0
			}
		}
	}
}

// executeCast performs the spell effect.
func (s *SpellCastingSystem) executeCast(caster *Entity, spell *magic.Spell, slotIndex int) {
	// Check mana cost
	manaComp, hasMana := caster.GetComponent("mana")
	if !hasMana {
		return
	}
	mana := manaComp.(*ManaComponent)

	if mana.Current < spell.Stats.ManaCost {
		// Not enough mana - show notification to player
		if s.tutorialSys != nil {
			s.tutorialSys.ShowNotification("Not enough mana!", 1.5)
		}
		return
	}

	// Deduct mana cost
	mana.Current -= spell.Stats.ManaCost
	if mana.Current < 0 {
		mana.Current = 0
	}

	// Get caster position for targeting
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		return
	}
	pos := posComp.(*PositionComponent)

	// Apply spell effects based on type
	switch spell.Type {
	case magic.TypeOffensive:
		s.castOffensiveSpell(caster, spell, pos.X, pos.Y)
	case magic.TypeHealing:
		s.castHealingSpell(caster, spell)
	case magic.TypeDefensive:
		s.castDefensiveSpell(caster, spell)
	case magic.TypeBuff:
		s.castBuffSpell(caster, spell)
	case magic.TypeDebuff:
		s.castDebuffSpell(caster, spell, pos.X, pos.Y)
	case magic.TypeUtility:
		s.castUtilitySpell(caster, spell)
	}

	// Play cast sound effect (genre-aware)
	if s.audioMgr != nil {
		effectType := "magic" // Generic magic sound
		if err := s.audioMgr.PlaySFX(effectType, int64(caster.ID)); err != nil {
			// Audio failure is non-critical, continue
			_ = err
		}
	}

	// Spawn cast visual effect (magic particles at caster position)
	if s.particleSys != nil {
		s.particleSys.SpawnMagicParticles(s.world, pos.X, pos.Y, int64(caster.ID), "fantasy")
	}

	// Phase 5.3: Spawn spell light for dynamic lighting
	// Light duration matches typical spell effect duration (2-3 seconds)
	s.spawnSpellLight(pos.X, pos.Y, spell, 2.5)
}

// castOffensiveSpell deals damage to enemies in range.
func (s *SpellCastingSystem) castOffensiveSpell(caster *Entity, spell *magic.Spell, x, y float64) {
	// Find targets based on spell target type
	targets := s.findTargets(caster, spell, x, y)

	for _, target := range targets {
		// Apply damage
		healthComp, hasHealth := target.GetComponent("health")
		if !hasHealth {
			continue
		}
		health := healthComp.(*HealthComponent)

		health.Current -= float64(spell.Stats.Damage)
		if health.Current < 0 {
			health.Current = 0
		}

		// Apply elemental effects based on spell element
		if s.statusEffectSys != nil {
			s.applyElementalEffect(target, spell)
		}

		// Spawn damage visual effect based on element
		if s.particleSys != nil {
			targetPos, hasPos := target.GetComponent("position")
			if hasPos {
				pos := targetPos.(*PositionComponent)
				// Spawn element-specific particles
				s.spawnElementalHitEffect(pos.X, pos.Y, spell.Element, target.ID)
			}
		}

		// Play impact sound effect
		if s.audioMgr != nil {
			_ = s.audioMgr.PlaySFX("impact", int64(target.ID))
		}
	}
}

// castHealingSpell restores health to caster or allies.
func (s *SpellCastingSystem) castHealingSpell(caster *Entity, spell *magic.Spell) {
	target := caster
	if spell.Target == magic.TargetSingle {
		// Find nearest injured ally in range
		ally := s.findNearestInjuredAlly(caster, spell.Stats.Range)
		if ally != nil {
			target = ally
		}
	} else if spell.Target == magic.TargetArea || spell.Target == magic.TargetAllAllies {
		// Heal multiple allies
		allies := s.findAlliesInRange(caster, spell.Stats.AreaSize)
		for _, ally := range allies {
			s.healTarget(ally, spell)
		}
		return
	}

	s.healTarget(target, spell)
}

// healTarget applies healing to a single target.
func (s *SpellCastingSystem) healTarget(target *Entity, spell *magic.Spell) {
	healthComp, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return
	}
	health := healthComp.(*HealthComponent)

	health.Current += float64(spell.Stats.Healing)
	if health.Current > health.Max {
		health.Current = health.Max
	}

	// Spawn healing visual effect (green/gold particles rising upward)
	if s.particleSys != nil {
		targetPos, hasPos := target.GetComponent("position")
		if hasPos {
			pos := targetPos.(*PositionComponent)
			config := particles.Config{
				Type:     particles.ParticleMagic,
				Count:    20,
				GenreID:  "fantasy",
				Seed:     int64(target.ID),
				Duration: 1.0,
				SpreadX:  60.0,
				SpreadY:  60.0,
				Gravity:  -80.0, // Rise upward for healing
				MinSize:  4.0,
				MaxSize:  8.0,
				Custom:   map[string]interface{}{"color": "healing"},
			}
			s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
		}
	}

	// Play healing sound effect
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(target.ID))
	}
}

// findNearestInjuredAlly finds the nearest ally that needs healing.
func (s *SpellCastingSystem) findNearestInjuredAlly(caster *Entity, maxRange float64) *Entity {
	entities := s.world.GetEntities()
	var nearestAlly *Entity
	minDist := maxRange

	// Get caster's team
	var casterTeamID int
	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
		casterTeamID = teamComp.(*TeamComponent).TeamID
	}

	for _, entity := range entities {
		if entity == caster {
			continue
		}

		// Check if ally
		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
			team := teamComp.(*TeamComponent)
			if !team.IsAlly(casterTeamID) {
				continue
			}
		} else {
			// No team component - skip
			continue
		}

		// Check if injured
		healthComp, hasHealth := entity.GetComponent("health")
		if !hasHealth {
			continue
		}
		health := healthComp.(*HealthComponent)
		if health.Current >= health.Max {
			continue // At full health
		}

		// Check distance
		dist := GetDistance(caster, entity)
		if dist <= minDist {
			nearestAlly = entity
			minDist = dist
		}
	}

	return nearestAlly
}

// findAlliesInRange finds all allies within range.
func (s *SpellCastingSystem) findAlliesInRange(caster *Entity, maxRange float64) []*Entity {
	entities := s.world.GetEntities()
	var allies []*Entity

	// Get caster's team
	var casterTeamID int
	if teamComp, hasTeam := caster.GetComponent("team"); hasTeam {
		casterTeamID = teamComp.(*TeamComponent).TeamID
	}

	for _, entity := range entities {
		// Check if ally (including self)
		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
			team := teamComp.(*TeamComponent)
			if !team.IsAlly(casterTeamID) {
				continue
			}
		} else if entity != caster {
			continue
		}

		// Check if has health
		if !entity.HasComponent("health") {
			continue
		}

		// Check distance
		dist := GetDistance(caster, entity)
		if dist <= maxRange {
			allies = append(allies, entity)
		}
	}

	return allies
}

// castDefensiveSpell applies shields or defensive buffs.
func (s *SpellCastingSystem) castDefensiveSpell(caster *Entity, spell *magic.Spell) {
	// Apply shield using the damage stat as shield strength
	if s.statusEffectSys != nil {
		shieldAmount := float64(spell.Stats.Damage)
		if shieldAmount <= 0 {
			shieldAmount = 50.0 // Default shield if no damage stat
		}

		duration := spell.Stats.Duration
		if duration <= 0 {
			duration = 30.0 // Default duration
		}

		s.statusEffectSys.ApplyShield(caster, shieldAmount, duration)
	}
}

// castBuffSpell applies stat boosts.
func (s *SpellCastingSystem) castBuffSpell(caster *Entity, spell *magic.Spell) {
	if s.statusEffectSys == nil {
		return
	}

	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 30.0 // Default duration
	}

	// Determine buff type based on spell element
	switch spell.Element {
	case magic.ElementWind:
		// Haste - increased attack speed (represented as attack boost)
		s.statusEffectSys.ApplyStatusEffect(caster, "haste", 0.5, duration, 0)
	case magic.ElementLight:
		// Strength - increased attack
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.3, duration, 0)
	case magic.ElementEarth:
		// Fortify - increased defense
		s.statusEffectSys.ApplyStatusEffect(caster, "fortify", 0.3, duration, 0)
	default:
		// Generic buff - small attack and defense boost
		s.statusEffectSys.ApplyStatusEffect(caster, "strength", 0.2, duration, 0)
	}
}

// castDebuffSpell applies stat reductions to enemies.
func (s *SpellCastingSystem) castDebuffSpell(caster *Entity, spell *magic.Spell, x, y float64) {
	targets := s.findTargets(caster, spell, x, y)

	for _, target := range targets {
		// Apply minor damage if any
		if spell.Stats.Damage > 0 {
			healthComp, hasHealth := target.GetComponent("health")
			if hasHealth {
				health := healthComp.(*HealthComponent)
				health.Current -= float64(spell.Stats.Damage)
				if health.Current < 0 {
					health.Current = 0
				}
			}
		}

		// Apply debuff effects
		if s.statusEffectSys != nil {
			duration := spell.Stats.Duration
			if duration <= 0 {
				duration = 10.0 // Default duration
			}

			// Determine debuff type based on spell element
			switch spell.Element {
			case magic.ElementDark:
				// Weakness - reduced attack
				s.statusEffectSys.ApplyStatusEffect(target, "weakness", 0.7, duration, 0)
			case magic.ElementEarth:
				// Vulnerability - reduced defense
				s.statusEffectSys.ApplyStatusEffect(target, "vulnerability", 0.7, duration, 0)
			default:
				// Generic debuff - small attack reduction
				s.statusEffectSys.ApplyStatusEffect(target, "weakness", 0.8, duration, 0)
			}
		}
	}
}

// applyElementalEffect applies status effects based on spell element.
func (s *SpellCastingSystem) applyElementalEffect(target *Entity, spell *magic.Spell) {
	switch spell.Element {
	case magic.ElementFire:
		// Burning: 10 damage per second for 3 seconds
		s.statusEffectSys.ApplyStatusEffect(target, "burning", 10.0, 3.0, 1.0)

	case magic.ElementIce:
		// Frozen: 50% movement slow for 2 seconds (visual indicator only, actual movement handled by AI)
		s.statusEffectSys.ApplyStatusEffect(target, "frozen", 0.5, 2.0, 0)

	case magic.ElementLightning:
		// Shocked: chain to nearby enemies
		if spell.Target == magic.TargetSingle || spell.Target == magic.TargetArea {
			s.statusEffectSys.ChainLightning(nil, target, float64(spell.Stats.Damage)*0.5, 2, 15.0)
		}
		// Apply shocked marker for visual effects
		s.statusEffectSys.ApplyStatusEffect(target, "shocked", 0, 2.0, 0)

	case magic.ElementEarth:
		// Earth spells can apply poison effect
		// Poison: 5 damage per second ignoring armor for 5 seconds
		if s.shouldApplyPoison() {
			s.statusEffectSys.ApplyStatusEffect(target, "poisoned", 5.0, 5.0, 1.0)
		}
	}
}

// shouldApplyPoison returns true 30% of the time for Earth spells.
func (s *SpellCastingSystem) shouldApplyPoison() bool {
	// Use status effect system's RNG if available
	if s.statusEffectSys != nil && s.statusEffectSys.rng != nil {
		return s.statusEffectSys.rng.Float64() < 0.3
	}
	return true
}

// castUtilitySpell handles non-combat spells.
func (s *SpellCastingSystem) castUtilitySpell(caster *Entity, spell *magic.Spell) {
	// Determine utility spell type based on tags and element
	switch {
	case containsTag(spell.Tags, "teleport"):
		s.castTeleportSpell(caster, spell)
	case containsTag(spell.Tags, "light"), containsTag(spell.Tags, "reveal"):
		s.castRevealSpell(caster, spell)
	case containsTag(spell.Tags, "speed"), containsTag(spell.Tags, "haste"):
		s.castSpeedBoostSpell(caster, spell)
	default:
		// Generic utility effect based on element
		switch spell.Element {
		case magic.ElementLight:
			s.castRevealSpell(caster, spell)
		case magic.ElementWind:
			s.castSpeedBoostSpell(caster, spell)
		case magic.ElementArcane:
			s.castTeleportSpell(caster, spell)
		}
	}
}

// castTeleportSpell teleports the caster to a nearby safe location.
// Teleport distance is based on spell range, and landing spot must be walkable.
func (s *SpellCastingSystem) castTeleportSpell(caster *Entity, spell *magic.Spell) {
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		return
	}
	pos := posComp.(*PositionComponent)

	// Calculate teleport direction and distance
	// Use spell range as max teleport distance
	maxDist := spell.Stats.Range
	if maxDist <= 0 {
		maxDist = 100.0 // Default teleport range
	}

	// For simplicity, teleport forward (could be enhanced with mouse targeting later)
	// Use velocity direction if available, otherwise default direction
	dirX, dirY := 0.0, 1.0 // Default: down
	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
		vel := velComp.(*VelocityComponent)
		if vel.VX != 0 || vel.VY != 0 {
			// Normalize velocity to get direction
			mag := math.Sqrt(vel.VX*vel.VX + vel.VY*vel.VY)
			if mag > 0 {
				dirX = vel.VX / mag
				dirY = vel.VY / mag
			}
		}
	}

	// Calculate target position
	targetX := pos.X + dirX*maxDist
	targetY := pos.Y + dirY*maxDist

	// Validate landing position (check collision)
	if s.isPositionWalkable(targetX, targetY, caster) {
		// Teleport successful
		pos.X = targetX
		pos.Y = targetY

		// Spawn teleport visual effect at departure
		if s.particleSys != nil {
			config := particles.Config{
				Type:     particles.ParticleMagic,
				Count:    30,
				GenreID:  "fantasy",
				Seed:     int64(caster.ID),
				Duration: 0.5,
				SpreadX:  80.0,
				SpreadY:  80.0,
				Gravity:  0.0,
				MinSize:  4.0,
				MaxSize:  8.0,
				Custom:   map[string]interface{}{"color": "teleport"},
			}
			s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
		}

		// Play teleport sound
		if s.audioMgr != nil {
			_ = s.audioMgr.PlaySFX("magic", int64(caster.ID))
		}
	}
	// If position not walkable, teleport fails (mana still consumed as per executeCast)
}

// castRevealSpell reveals fog of war in an area around the caster.
// Useful for exploration and finding hidden enemies/items.
func (s *SpellCastingSystem) castRevealSpell(caster *Entity, spell *magic.Spell) {
	posComp, hasPos := caster.GetComponent("position")
	if !hasPos {
		return
	}
	pos := posComp.(*PositionComponent)

	// Determine reveal radius from spell stats
	revealRadius := spell.Stats.AreaSize
	if revealRadius <= 0 {
		revealRadius = 200.0 // Default reveal radius
	}

	// Reveal fog of war (requires access to map UI system)
	// This is handled at a higher level in the game loop
	// For now, we mark this with a temporary status effect that the map system can detect
	if s.statusEffectSys != nil {
		duration := 0.1 // Brief marker effect
		s.statusEffectSys.ApplyStatusEffect(caster, "revealing", revealRadius, duration, 0)
	}

	// Spawn light particles to indicate reveal
	if s.particleSys != nil {
		config := particles.Config{
			Type:     particles.ParticleSpark,
			Count:    40,
			GenreID:  "fantasy",
			Seed:     int64(caster.ID),
			Duration: 1.5,
			SpreadX:  revealRadius,
			SpreadY:  revealRadius,
			Gravity:  -20.0, // Slow rise
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   map[string]interface{}{"color": "light"},
		}
		s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
	}

	// Play light sound
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(caster.ID))
	}
}

// castSpeedBoostSpell applies a temporary speed boost to the caster.
// Increases movement speed for exploration or combat mobility.
func (s *SpellCastingSystem) castSpeedBoostSpell(caster *Entity, spell *magic.Spell) {
	if s.statusEffectSys == nil {
		return
	}

	// Determine duration from spell stats
	duration := spell.Stats.Duration
	if duration <= 0 {
		duration = 10.0 // Default speed boost duration
	}

	// Speed multiplier based on spell power
	speedMultiplier := 1.5 // 50% speed increase
	if spell.Rarity >= magic.RarityRare {
		speedMultiplier = 2.0 // 100% speed increase for rare+ spells
	}

	// Apply speed boost as a status effect
	// The movement system will need to check for this effect
	s.statusEffectSys.ApplyStatusEffect(caster, "speed_boost", speedMultiplier, duration, 0)

	// Spawn speed particles (fast-moving wind particles)
	if s.particleSys != nil {
		posComp, hasPos := caster.GetComponent("position")
		if hasPos {
			pos := posComp.(*PositionComponent)
			config := particles.Config{
				Type:     particles.ParticleDust,
				Count:    25,
				GenreID:  "fantasy",
				Seed:     int64(caster.ID),
				Duration: duration,
				SpreadX:  100.0,
				SpreadY:  60.0,
				Gravity:  10.0,
				MinSize:  2.0,
				MaxSize:  4.0,
				Custom:   map[string]interface{}{"color": "wind"},
			}
			s.particleSys.SpawnParticles(s.world, config, pos.X, pos.Y)
		}
	}

	// Play speed boost sound
	if s.audioMgr != nil {
		_ = s.audioMgr.PlaySFX("powerup", int64(caster.ID))
	}
}

// isPositionWalkable checks if a position is valid for teleportation.
// Returns true if the position has no solid colliders.
func (s *SpellCastingSystem) isPositionWalkable(x, y float64, caster *Entity) bool {
	entities := s.world.GetEntities()

	// Check for collisions with solid entities
	for _, entity := range entities {
		if entity == caster {
			continue
		}

		// Check if entity has collider
		colliderComp, hasCollider := entity.GetComponent("collider")
		if !hasCollider {
			continue
		}
		collider := colliderComp.(*ColliderComponent)

		// Skip non-solid colliders
		if !collider.Solid {
			continue
		}

		// Check if position intersects with this collider
		entityPos, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		pos := entityPos.(*PositionComponent)

		// Get caster collider for size checking
		casterCollider, hasCasterCollider := caster.GetComponent("collider")
		if hasCasterCollider {
			cc := casterCollider.(*ColliderComponent)
			// Create temporary collider at target position
			tempCollider := &ColliderComponent{
				Width:   cc.Width,
				Height:  cc.Height,
				OffsetX: cc.OffsetX,
				OffsetY: cc.OffsetY,
				Solid:   true,
				Layer:   cc.Layer,
			}
			if tempCollider.Intersects(x, y, collider, pos.X, pos.Y) {
				return false // Collision detected
			}
		} else {
			// No caster collider, check point collision
			minX, minY, maxX, maxY := collider.GetBounds(pos.X, pos.Y)
			if x >= minX && x <= maxX && y >= minY && y <= maxY {
				return false
			}
		}
	}

	return true // Position is walkable
}

// containsTag checks if a spell has a specific tag.
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// findTargets returns entities affected by the spell.
func (s *SpellCastingSystem) findTargets(caster *Entity, spell *magic.Spell, x, y float64) []*Entity {
	var targets []*Entity

	entities := s.world.GetEntities()

	// Helper to check if entity is valid enemy target
	isEnemyTarget := func(entity *Entity) bool {
		if entity == caster {
			return false
		}
		// Player has input component
		if entity.HasComponent("input") {
			return false
		}
		// Must have health to be a valid target
		if !entity.HasComponent("health") {
			return false
		}
		return true
	}

	switch spell.Target {
	case magic.TargetSelf:
		targets = append(targets, caster)

	case magic.TargetSingle:
		// Find nearest enemy in range
		var nearest *Entity
		nearestDist := spell.Stats.Range

		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			dist := GetDistance(caster, entity)
			if dist <= nearestDist {
				nearest = entity
				nearestDist = dist
			}
		}

		if nearest != nil {
			targets = append(targets, nearest)
		}

	case magic.TargetArea:
		// Find all enemies in area
		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			dist := GetDistance(caster, entity)
			if dist <= spell.Stats.AreaSize {
				targets = append(targets, entity)
			}
		}

	case magic.TargetAllEnemies:
		// All enemies regardless of distance
		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			targets = append(targets, entity)
		}

	case magic.TargetCone:
		// Cone targeting: entities within angle from caster's facing direction
		casterPos, hasCasterPos := caster.GetComponent("position")
		if !hasCasterPos {
			break
		}
		casterPosComp := casterPos.(*PositionComponent)

		// Get caster's facing direction (use velocity or mouse aim)
		dirX, dirY := s.getCasterDirection(caster, x, y)
		if dirX == 0 && dirY == 0 {
			dirX = 1.0 // Default to facing right
		}

		// Normalize direction
		dirLength := math.Sqrt(dirX*dirX + dirY*dirY)
		dirX /= dirLength
		dirY /= dirLength

		// Cone parameters
		coneAngle := 45.0 * math.Pi / 180.0 // 45-degree cone (adjustable)

		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			entityPos, hasPos := entity.GetComponent("position")
			if !hasPos {
				continue
			}
			entityPosComp := entityPos.(*PositionComponent)

			// Vector from caster to entity
			toEntityX := entityPosComp.X - casterPosComp.X
			toEntityY := entityPosComp.Y - casterPosComp.Y
			dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

			// Check if within range
			if dist > spell.Stats.Range || dist < 0.1 {
				continue
			}

			// Normalize to entity vector
			toEntityX /= dist
			toEntityY /= dist

			// Calculate angle between direction and to-entity vector
			dotProduct := dirX*toEntityX + dirY*toEntityY
			angle := math.Acos(math.Max(-1.0, math.Min(1.0, dotProduct)))

			// Check if within cone angle
			if angle <= coneAngle {
				targets = append(targets, entity)
			}
		}

	case magic.TargetLine:
		// Line targeting: entities along a line from caster in facing direction
		casterPos, hasCasterPos := caster.GetComponent("position")
		if !hasCasterPos {
			break
		}
		casterPosComp := casterPos.(*PositionComponent)

		// Get caster's facing direction (use velocity or mouse aim)
		dirX, dirY := s.getCasterDirection(caster, x, y)
		if dirX == 0 && dirY == 0 {
			dirX = 1.0 // Default to facing right
		}

		// Normalize direction
		dirLength := math.Sqrt(dirX*dirX + dirY*dirY)
		dirX /= dirLength
		dirY /= dirLength

		// Line width tolerance (in pixels)
		lineWidth := 32.0

		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			entityPos, hasPos := entity.GetComponent("position")
			if !hasPos {
				continue
			}
			entityPosComp := entityPos.(*PositionComponent)

			// Vector from caster to entity
			toEntityX := entityPosComp.X - casterPosComp.X
			toEntityY := entityPosComp.Y - casterPosComp.Y
			dist := math.Sqrt(toEntityX*toEntityX + toEntityY*toEntityY)

			// Check if within range
			if dist > spell.Stats.Range || dist < 0.1 {
				continue
			}

			// Calculate perpendicular distance from line
			// Project entity position onto line direction
			projection := toEntityX*dirX + toEntityY*dirY
			if projection < 0 {
				continue // Behind caster
			}

			// Calculate perpendicular distance
			perpX := toEntityX - projection*dirX
			perpY := toEntityY - projection*dirY
			perpDist := math.Sqrt(perpX*perpX + perpY*perpY)

			// Check if within line width
			if perpDist <= lineWidth {
				targets = append(targets, entity)
			}
		}
	}

	return targets
}

// getCasterDirection determines the caster's facing direction for directional spells.
// Uses velocity if moving, otherwise uses direction towards target point (x, y).
func (s *SpellCastingSystem) getCasterDirection(caster *Entity, targetX, targetY float64) (dirX, dirY float64) {
	// Try to use velocity for moving entities
	if velComp, hasVel := caster.GetComponent("velocity"); hasVel {
		vel := velComp.(*VelocityComponent)
		if vel.VX != 0 || vel.VY != 0 {
			return vel.VX, vel.VY
		}
	}

	// Fall back to direction towards target point
	if posComp, hasPos := caster.GetComponent("position"); hasPos {
		pos := posComp.(*PositionComponent)
		dirX = targetX - pos.X
		dirY = targetY - pos.Y
		return dirX, dirY
	}

	// Default to facing right if no position
	return 1.0, 0.0
}

// StartCast initiates casting a spell from a slot.
func (s *SpellCastingSystem) StartCast(entity *Entity, slotIndex int) bool {
	spellComp, hasSpells := entity.GetComponent("spell_slots")
	if !hasSpells {
		return false
	}
	slots := spellComp.(*SpellSlotComponent)

	// Check if already casting
	if slots.IsCasting() {
		return false
	}

	// Check slot validity
	spell := slots.GetSlot(slotIndex)
	if spell == nil {
		return false
	}

	// Check cooldown
	if slots.IsOnCooldown(slotIndex) {
		return false
	}

	// Check mana
	manaComp, hasMana := entity.GetComponent("mana")
	if !hasMana {
		return false
	}
	mana := manaComp.(*ManaComponent)
	if mana.Current < spell.Stats.ManaCost {
		return false
	}

	// Start casting
	slots.Casting = slotIndex
	slots.CastingBar = 0

	return true
}

// CancelCast interrupts current spell cast.
func (s *SpellCastingSystem) CancelCast(entity *Entity) {
	spellComp, hasSpells := entity.GetComponent("spell_slots")
	if !hasSpells {
		return
	}
	slots := spellComp.(*SpellSlotComponent)

	slots.Casting = -1
	slots.CastingBar = 0
}

// spawnElementalHitEffect creates element-specific particle effects for spell hits.
func (s *SpellCastingSystem) spawnElementalHitEffect(x, y float64, element magic.ElementType, seed uint64) {
	if s.particleSys == nil {
		return
	}

	var config particles.Config
	switch element {
	case magic.ElementFire:
		// Fire: Orange/red flames rising upward
		config = particles.Config{
			Type:     particles.ParticleFlame,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.8,
			SpreadX:  80.0,
			SpreadY:  80.0,
			Gravity:  -100.0, // Rise upward
			MinSize:  3.0,
			MaxSize:  7.0,
			Custom:   make(map[string]interface{}),
		}

	case magic.ElementIce:
		// Ice: Blue/white crystals with slower movement
		config = particles.Config{
			Type:     particles.ParticleMagic, // Use magic particles with blue tint
			Count:    15,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.2,
			SpreadX:  60.0,
			SpreadY:  60.0,
			Gravity:  50.0, // Slow fall
			MinSize:  4.0,
			MaxSize:  8.0,
			Custom:   map[string]interface{}{"color": "ice"},
		}

	case magic.ElementLightning:
		// Lightning: Fast yellow/white sparks
		config = particles.Config{
			Type:     particles.ParticleSpark,
			Count:    25,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.4,
			SpreadX:  120.0,
			SpreadY:  120.0,
			Gravity:  0.0, // No gravity, pure energy
			MinSize:  2.0,
			MaxSize:  5.0,
			Custom:   map[string]interface{}{"color": "lightning"},
		}

	case magic.ElementEarth:
		// Earth: Brown/green dust/rock particles (can apply poison)
		config = particles.Config{
			Type:     particles.ParticleDust,
			Count:    18,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.2,
			SpreadX:  70.0,
			SpreadY:  70.0,
			Gravity:  100.0, // Fall to ground
			MinSize:  3.0,
			MaxSize:  7.0,
			Custom:   map[string]interface{}{"color": "earth"},
		}

	case magic.ElementWind:
		// Wind: Fast-moving dust particles
		config = particles.Config{
			Type:     particles.ParticleDust,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.6,
			SpreadX:  150.0,
			SpreadY:  80.0,
			Gravity:  20.0,
			MinSize:  2.0,
			MaxSize:  4.0,
			Custom:   map[string]interface{}{"color": "wind"},
		}

	case magic.ElementLight:
		// Light: Bright white/yellow particles
		config = particles.Config{
			Type:     particles.ParticleSpark,
			Count:    22,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.0,
			SpreadX:  100.0,
			SpreadY:  100.0,
			Gravity:  -30.0, // Slow rise
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   map[string]interface{}{"color": "light"},
		}

	case magic.ElementDark:
		// Dark: Purple/black smoke particles
		config = particles.Config{
			Type:     particles.ParticleSmoke,
			Count:    20,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 1.5,
			SpreadX:  90.0,
			SpreadY:  90.0,
			Gravity:  -10.0, // Slow rise
			MinSize:  4.0,
			MaxSize:  8.0,
			Custom:   map[string]interface{}{"color": "dark"},
		}

	default:
		// Generic magic effect for other elements (none, arcane)
		config = particles.Config{
			Type:     particles.ParticleMagic,
			Count:    15,
			GenreID:  "fantasy",
			Seed:     int64(seed),
			Duration: 0.8,
			SpreadX:  90.0,
			SpreadY:  90.0,
			Gravity:  -50.0,
			MinSize:  3.0,
			MaxSize:  6.0,
			Custom:   make(map[string]interface{}),
		}
	}

	s.particleSys.SpawnParticles(s.world, config, x, y)
}

// ManaRegenSystem regenerates mana over time.
type ManaRegenSystem struct{}

// Update regenerates mana for all entities.
func (s *ManaRegenSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		manaComp, hasMana := entity.GetComponent("mana")
		if !hasMana {
			continue
		}
		mana := manaComp.(*ManaComponent)

		// Regenerate mana
		mana.Current += int(mana.Regen * deltaTime)
		if mana.Current > mana.Max {
			mana.Current = mana.Max
		}
	}
}

// LoadPlayerSpells generates and equips spells for the player.
func LoadPlayerSpells(player *Entity, seed int64, genreID string, depth int) error {
	// Generate spells using procgen system
	generator := magic.NewSpellGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      depth,
		GenreID:    genreID,
		Custom: map[string]interface{}{
			"count": 5, // Generate 5 spells for the 5 slots
		},
	}

	result, err := generator.Generate(seed, params)
	if err != nil {
		return err
	}

	spells := result.([]*magic.Spell)

	// Create spell slots component if doesn't exist
	var slots *SpellSlotComponent
	if !player.HasComponent("spell_slots") {
		slots = &SpellSlotComponent{
			Casting: -1,
		}
		player.AddComponent(slots)
	} else {
		slotsComp, _ := player.GetComponent("spell_slots")
		slots = slotsComp.(*SpellSlotComponent)
	}

	// Equip spells to slots
	for i := 0; i < 5 && i < len(spells); i++ {
		slots.SetSlot(i, spells[i])
	}

	return nil
}

// spawnSpellLight creates a temporary light entity at the spell cast position.
// The light color and intensity are based on the spell's elemental type.
// This function is part of Phase 5.3: Dynamic Lighting System Integration.
func (s *SpellCastingSystem) spawnSpellLight(x, y float64, spell *magic.Spell, duration float64) {
	// Get light color based on spell element
	lightColor := getElementLightColor(spell.Element)
	
	// Create light entity
	lightEntity := s.world.CreateEntity()
	
	// Add position component
	lightEntity.AddComponent(&PositionComponent{X: x, Y: y})
	
	// Create spell light with appropriate radius and color
	// Radius scaled by spell power (damage/healing amount)
	baseRadius := 100.0
	powerScale := math.Min(float64(spell.Stats.Damage+spell.Stats.Healing)/50.0, 2.0)
	radius := baseRadius * powerScale
	
	spellLight := NewSpellLight(radius, lightColor)
	spellLight.Pulsing = true      // Spells have pulsing lights
	spellLight.PulseSpeed = 4.0    // Fast pulse for dramatic effect
	spellLight.PulseAmount = 0.3   // Moderate pulse intensity
	lightEntity.AddComponent(spellLight)
	
	// Add lifetime component so light despawns automatically
	lightEntity.AddComponent(&LifetimeComponent{
		Duration: duration,
		Elapsed:  0,
	})
}

// getElementLightColor returns the appropriate light color for a spell element.
// Colors are chosen to match the visual theme of each element while providing
// good visibility and atmosphere.
func getElementLightColor(element magic.ElementType) color.RGBA {
	switch element {
	case magic.ElementFire:
		return color.RGBA{255, 100, 0, 255} // Orange-red (warm fire)
	case magic.ElementIce:
		return color.RGBA{100, 200, 255, 255} // Cyan (cold ice)
	case magic.ElementLightning:
		return color.RGBA{255, 255, 150, 255} // Bright yellow (electric)
	case magic.ElementEarth:
		return color.RGBA{139, 90, 43, 255} // Brown (earthy)
	case magic.ElementWind:
		return color.RGBA{200, 230, 255, 255} // Light cyan (airy)
	case magic.ElementLight:
		return color.RGBA{255, 255, 220, 255} // Bright white-yellow (holy)
	case magic.ElementDark:
		return color.RGBA{100, 50, 150, 255} // Purple (shadowy)
	case magic.ElementArcane:
		return color.RGBA{200, 100, 255, 255} // Magenta (pure magic)
	case magic.ElementNone:
		return color.RGBA{180, 180, 200, 255} // Neutral grey-blue
	default:
		return color.RGBA{180, 180, 200, 255} // Default grey-blue
	}
}

// LifetimeComponent marks an entity for automatic despawn after a duration.
// Used for temporary entities like spell lights and particle effects.
type LifetimeComponent struct {
	Duration float64 // Total lifetime in seconds
	Elapsed  float64 // Time elapsed since creation
}

// Type implements Component interface.
func (l *LifetimeComponent) Type() string {
	return "lifetime"
}
