package engine

import (
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
	particleSys     *ParticleSystem // For visual effects
	audioMgr        *AudioManager   // For sound effects
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
		// Not enough mana
		// TODO: Show "Not enough mana" message
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
	// TODO: Implement utility spells (teleport, light, reveal map, etc.)
	// For now, just consume mana
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

	case magic.TargetCone, magic.TargetLine:
		// TODO: Implement directional targeting
		// For now, treat like single target
		for _, entity := range entities {
			if !isEnemyTarget(entity) {
				continue
			}

			dist := GetDistance(caster, entity)
			if dist <= spell.Stats.Range {
				targets = append(targets, entity)
				break // Just one for now
			}
		}
	}

	return targets
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
