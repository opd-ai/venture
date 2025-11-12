package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// SpellEffectSystem manages spell effects on entities and terrain.
// It processes SpellEffectComponents and executes their effects based on type.
type SpellEffectSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry
}

// NewSpellEffectSystem creates a new spell effect system.
func NewSpellEffectSystem(world *World, rng *rand.Rand) *SpellEffectSystem {
	return NewSpellEffectSystemWithLogger(world, rng, nil)
}

// NewSpellEffectSystemWithLogger creates a new spell effect system with a logger.
func NewSpellEffectSystemWithLogger(world *World, rng *rand.Rand, logger *logrus.Logger) *SpellEffectSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "spell_effect")
	}
	return &SpellEffectSystem{
		world:  world,
		rng:    rng,
		logger: logEntry,
	}
}

// Update processes all active spell effects.
func (s *SpellEffectSystem) Update(entities []*Entity, deltaTime float64) {
	var effectsToRemove []struct {
		entity *Entity
		effect *SpellEffectComponent
	}

	// Process all entities with spell effects
	for _, entity := range entities {
		for _, comp := range entity.Components {
			if effect, ok := comp.(*SpellEffectComponent); ok {
				if !effect.Active {
					continue
				}

				// Update effect timer
				effect.Update(deltaTime)

				// Execute effect based on type
				s.executeEffect(entity, effect, deltaTime)

				// Mark expired effects for removal
				if effect.IsExpired() {
					effectsToRemove = append(effectsToRemove, struct {
						entity *Entity
						effect *SpellEffectComponent
					}{entity, effect})
				}
			}
		}
	}

	// Remove expired effects
	for _, item := range effectsToRemove {
		item.entity.RemoveComponent(item.effect.Type())
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"entity_id":   item.entity.ID,
				"effect_type": item.effect.EffectType.String(),
			}).Debug("Spell effect expired")
		}
	}
}

// executeEffect executes a spell effect based on its type.
func (s *SpellEffectSystem) executeEffect(entity *Entity, effect *SpellEffectComponent, deltaTime float64) {
	switch effect.EffectType {
	case EffectTerrainManipulation:
		s.executeTerrainManipulation(effect)
	case EffectTransmutation:
		s.executeTransmutation(effect)
	case EffectSummoning:
		s.executeSummoning(effect)
	case EffectIllusion:
		s.executeIllusion(entity, effect)
	case EffectTimeManipulation:
		s.executeTimeManipulation(entity, effect, deltaTime)
	case EffectGravityControl:
		s.executeGravityControl(entity, effect, deltaTime)
	case EffectElementalFusion:
		s.executeElementalFusion(effect)
	case EffectLifeDrain:
		s.executeLifeDrain(effect, deltaTime)
	case EffectTeleportation:
		s.executeTeleportation(entity, effect)
	case EffectMetamagic:
		s.executeMetamagic(entity, effect)
	}
}

// executeTerrainManipulation creates, modifies, or destroys terrain.
func (s *SpellEffectSystem) executeTerrainManipulation(effect *SpellEffectComponent) {
	// Terrain manipulation would modify world terrain tiles
	// This is a placeholder for terrain system integration
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":            effect.TargetX,
			"y":            effect.TargetY,
			"radius":       effect.Radius,
			"terrain_type": effect.TerrainModifier,
		}).Debug("Terrain manipulation executed")
	}
}

// executeTransmutation converts materials (stone→gold, water→ice).
func (s *SpellEffectSystem) executeTransmutation(effect *SpellEffectComponent) {
	// Transmutation would convert terrain or entity materials
	// This is a placeholder for material system integration
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"x":           effect.TargetX,
			"y":           effect.TargetY,
			"radius":      effect.Radius,
			"target_type": effect.TerrainModifier,
		}).Debug("Transmutation executed")
	}
}

// executeSummoning spawns temporary allies or objects.
func (s *SpellEffectSystem) executeSummoning(effect *SpellEffectComponent) {
	// Only spawn once at the start
	if effect.ElapsedTime > 0.016 { // Skip after first frame
		return
	}

	// Summoning would create temporary entities
	// This is a placeholder for entity spawning integration
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"template": effect.SummonTemplate,
			"x":        effect.TargetX,
			"y":        effect.TargetY,
			"duration": effect.Duration,
		}).Debug("Summoning executed")
	}
}

// executeIllusion creates decoys, invisibility, or confusion.
func (s *SpellEffectSystem) executeIllusion(entity *Entity, effect *SpellEffectComponent) {
	// Illusion effects modify rendering and AI perception
	// This is a placeholder for rendering/AI system integration

	// For invisibility, we could add/modify components
	if effect.Magnitude >= 0.9 { // Full invisibility threshold
		// Add invisibility marker (would be used by rendering system)
		if !entity.HasComponent("invisible") {
			entity.AddComponent(&GenericComponent{
				componentType: "invisible",
			})
		}
	}

	if s.logger != nil && effect.ElapsedTime < 0.016 {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"magnitude": effect.Magnitude,
			"duration":  effect.Duration,
		}).Debug("Illusion effect applied")
	}
}

// executeTimeManipulation slows, hastes, or rewinds positions.
func (s *SpellEffectSystem) executeTimeManipulation(entity *Entity, effect *SpellEffectComponent, deltaTime float64) {
	// Time manipulation affects movement and action speeds

	// Get velocity component if it exists
	if velComp, hasVel := entity.GetComponent("velocity"); hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			// Apply time scaling to velocity
			// Magnitude < 1.0 = slow, > 1.0 = haste
			if effect.Magnitude < 1.0 {
				// Slow effect - reduce velocity
				vel.VX *= effect.Magnitude
				vel.VY *= effect.Magnitude
			} else if effect.Magnitude > 1.0 {
				// Haste effect - increase velocity
				vel.VX *= effect.Magnitude
				vel.VY *= effect.Magnitude
			}
		}
	}

	if s.logger != nil && effect.ElapsedTime < 0.016 {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"magnitude": effect.Magnitude,
			"type":      "time_manipulation",
		}).Debug("Time manipulation applied")
	}
}

// executeGravityControl levitation, increased weight, orbital effects.
func (s *SpellEffectSystem) executeGravityControl(entity *Entity, effect *SpellEffectComponent, deltaTime float64) {
	// Gravity control affects physics and movement

	// Add or modify gravity component
	if !entity.HasComponent("gravity_modified") {
		entity.AddComponent(&GenericComponent{
			componentType: "gravity_modified",
		})
	}

	// Magnitude determines gravity strength
	// Negative = levitation, Positive = increased weight
	if s.logger != nil && effect.ElapsedTime < 0.016 {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"magnitude": effect.Magnitude,
			"duration":  effect.Duration,
		}).Debug("Gravity control applied")
	}
}

// executeElementalFusion combines elements (fire+ice=steam, earth+lightning=glass).
func (s *SpellEffectSystem) executeElementalFusion(effect *SpellEffectComponent) {
	// Elemental fusion creates new effects based on element combinations
	// This is a placeholder for elemental combo system

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"elements": effect.FusionElements,
			"x":        effect.TargetX,
			"y":        effect.TargetY,
			"radius":   effect.Radius,
			"damage":   effect.Magnitude,
		}).Debug("Elemental fusion executed")
	}
}

// executeLifeDrain transfers HP between entities.
func (s *SpellEffectSystem) executeLifeDrain(effect *SpellEffectComponent, deltaTime float64) {
	// Find caster and target entities
	var caster, target *Entity
	for _, entity := range s.world.GetEntities() {
		if entity.ID == effect.CasterID {
			caster = entity
		}
		if entity.ID == effect.TargetID {
			target = entity
		}
	}

	if caster == nil || target == nil {
		return
	}

	// Get health components
	targetHealthComp, hasTargetHealth := target.GetComponent("health")
	casterHealthComp, hasCasterHealth := caster.GetComponent("health")

	if !hasTargetHealth || !hasCasterHealth {
		return
	}

	targetHealth, okTarget := targetHealthComp.(*HealthComponent)
	casterHealth, okCaster := casterHealthComp.(*HealthComponent)

	if !okTarget || !okCaster {
		return
	}

	// Drain HP from target
	drainAmount := effect.Magnitude * deltaTime
	targetHealth.TakeDamage(drainAmount)

	// Transfer to caster
	casterHealth.Heal(drainAmount * 0.5) // 50% efficiency

	if s.logger != nil && int(effect.ElapsedTime*10)%10 == 0 { // Log every second
		s.logger.WithFields(logrus.Fields{
			"caster_id": effect.CasterID,
			"target_id": effect.TargetID,
			"amount":    drainAmount,
		}).Debug("Life drain tick")
	}
}

// executeTeleportation short-range blink or long-range portal.
func (s *SpellEffectSystem) executeTeleportation(entity *Entity, effect *SpellEffectComponent) {
	// Only execute once
	if effect.ElapsedTime > 0.016 { // Skip after first frame
		return
	}

	// Get position component
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Set new position
	oldX, oldY := pos.X, pos.Y
	pos.X = effect.TargetX
	pos.Y = effect.TargetY

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id": entity.ID,
			"from_x":    oldX,
			"from_y":    oldY,
			"to_x":      effect.TargetX,
			"to_y":      effect.TargetY,
		}).Debug("Teleportation executed")
	}
}

// executeMetamagic enhances other spells (double damage, multi-target).
func (s *SpellEffectSystem) executeMetamagic(entity *Entity, effect *SpellEffectComponent) {
	// Metamagic modifies other active spell effects
	// This is a placeholder for spell modifier system

	// Look for other spell effects on the entity
	for _, comp := range entity.Components {
		if otherEffect, ok := comp.(*SpellEffectComponent); ok {
			if otherEffect == effect {
				continue // Skip self
			}

			// Apply metamagic multiplier to other effects
			otherEffect.Magnitude *= effect.MetamagicMultiplier
		}
	}

	if s.logger != nil && effect.ElapsedTime < 0.016 {
		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"multiplier": effect.MetamagicMultiplier,
		}).Debug("Metamagic applied")
	}
}

// GenericComponent is a simple component for marking entities with flags.
type GenericComponent struct {
	componentType string
}

// Type returns the component type identifier.
func (g *GenericComponent) Type() string {
	return g.componentType
}

// ApplySpellEffect creates and applies a spell effect to an entity.
func (s *SpellEffectSystem) ApplySpellEffect(
	entity *Entity,
	effectType EffectType,
	magnitude float64,
	duration float64,
	targetType TargetType,
	casterID uint64,
	targetX float64,
	targetY float64,
	radius float64,
) {
	effect := &SpellEffectComponent{
		EffectType:  effectType,
		Duration:    duration,
		Magnitude:   magnitude,
		TargetType:  targetType,
		CasterID:    casterID,
		TargetX:     targetX,
		TargetY:     targetY,
		Radius:      radius,
		Active:      true,
		ElapsedTime: 0,
	}

	entity.AddComponent(effect)

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"effect_type": effectType.String(),
			"magnitude":   magnitude,
			"duration":    duration,
		}).Debug("Spell effect applied")
	}
}

// GetEntitiesInRadius returns all entities within a radius of a point.
func (s *SpellEffectSystem) GetEntitiesInRadius(x, y, radius float64) []*Entity {
	var result []*Entity
	radiusSquared := radius * radius

	for _, entity := range s.world.GetEntities() {
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance squared (faster than sqrt)
		dx := pos.X - x
		dy := pos.Y - y
		distSquared := dx*dx + dy*dy

		if distSquared <= radiusSquared {
			result = append(result, entity)
		}
	}

	return result
}

// CalculateDamageWithDistance calculates damage falloff based on distance from center.
func (s *SpellEffectSystem) CalculateDamageWithDistance(baseDamage, distance, maxRadius float64) float64 {
	if distance >= maxRadius {
		return 0
	}

	// Linear falloff: full damage at center, zero at edge
	falloff := 1.0 - (distance / maxRadius)
	return baseDamage * falloff
}

// GetAngleBetweenPoints calculates the angle from point1 to point2 in radians.
func (s *SpellEffectSystem) GetAngleBetweenPoints(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Atan2(dy, dx)
}
