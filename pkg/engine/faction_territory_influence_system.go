// Package engine provides the FactionTerritoryInfluenceSystem that applies
// faction-based combat and progression modifiers when entities enter faction-controlled zones.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// FactionTerritoryInfluenceSystem manages faction zone effects on entities.
// It tracks which faction zones entities are in and applies appropriate
// buffs or debuffs based on reputation with the controlling faction.
type FactionTerritoryInfluenceSystem struct {
	world           *World
	logger          *logrus.Entry
	factionSystem   *FactionSystem
	updateInterval  float64
	timeAccumulator float64
}

// NewFactionTerritoryInfluenceSystem creates a new faction territory influence system.
func NewFactionTerritoryInfluenceSystem(world *World, factionSystem *FactionSystem) *FactionTerritoryInfluenceSystem {
	logger := logrus.WithField("system_name", "faction_territory_influence")
	logger.Debug("Creating faction territory influence system")

	return &FactionTerritoryInfluenceSystem{
		world:           world,
		logger:          logger,
		factionSystem:   factionSystem,
		updateInterval:  0.25, // Update 4 times per second for responsiveness
		timeAccumulator: 0.0,
	}
}

// Update processes all entities and updates their faction territory modifiers.
func (s *FactionTerritoryInfluenceSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeAccumulator += deltaTime
	if s.timeAccumulator < s.updateInterval {
		return
	}
	s.timeAccumulator -= s.updateInterval

	// Collect all faction influence zones
	zones := s.collectInfluenceZones(entities)
	if len(zones) == 0 {
		return
	}

	// Process entities that can be affected by zones
	for _, entity := range entities {
		s.processEntity(entity, zones)
	}
}

// collectInfluenceZones gathers all active faction influence zone entities.
func (s *FactionTerritoryInfluenceSystem) collectInfluenceZones(entities []*Entity) []*Entity {
	zones := make([]*Entity, 0, 32)
	for _, entity := range entities {
		if entity.HasComponent("faction_territory_influence") {
			zones = append(zones, entity)
		}
	}
	return zones
}

// processEntity updates an entity's faction territory modifiers based on position.
func (s *FactionTerritoryInfluenceSystem) processEntity(entity *Entity, zones []*Entity) {
	// Entity needs position to check zone membership
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos || posComp == nil {
		return
	}
	pos := posComp.(*PositionComponent)

	// Get or create modifier component
	modComp, hasModifier := entity.GetComponent("faction_territory_modifier")
	var modifier *FactionTerritoryModifierComponent
	if hasModifier && modComp != nil {
		modifier = modComp.(*FactionTerritoryModifierComponent)
	} else {
		// Only add modifier to entities with faction component (players, NPCs)
		if !entity.HasComponent("faction") {
			return
		}
		modifier = NewFactionTerritoryModifierComponent()
		entity.AddComponent(modifier)
	}

	// Find the strongest overlapping zone
	strongestZone := s.findStrongestZone(pos.X, pos.Y, zones)

	if strongestZone == nil {
		// Entity left all faction zones
		if modifier.InFactionZone {
			s.clearModifiers(entity, modifier)
		}
		return
	}

	// Get zone component
	zoneComp, _ := strongestZone.GetComponent("faction_territory_influence")
	zone := zoneComp.(*FactionTerritoryInfluenceComponent)

	// Check if zone changed
	if modifier.ActiveFactionID != zone.FactionID ||
		modifier.ZoneX != zone.ZoneX ||
		modifier.ZoneZ != zone.ZoneZ {
		modifier.Dirty = true
	}

	if modifier.Dirty {
		s.applyZoneModifiers(entity, modifier, zone)
	}
}

// findStrongestZone returns the faction zone with highest influence at given position.
func (s *FactionTerritoryInfluenceSystem) findStrongestZone(x, y float64, zones []*Entity) *Entity {
	var strongest *Entity
	var strongestInfluence float64

	for _, zoneEntity := range zones {
		zoneComp, ok := zoneEntity.GetComponent("faction_territory_influence")
		if !ok || zoneComp == nil {
			continue
		}
		zone := zoneComp.(*FactionTerritoryInfluenceComponent)

		// Calculate zone center position (convert grid coords to world coords)
		centerX := float64(zone.ZoneX) * zone.ZoneRadius * 2
		centerY := float64(zone.ZoneZ) * zone.ZoneRadius * 2

		// Check if position is within zone
		dist := math.Sqrt((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY))
		if dist > zone.ZoneRadius {
			continue
		}

		// Calculate effective influence (stronger near center)
		distanceFactor := 1.0 - (dist / zone.ZoneRadius)
		effectiveInfluence := zone.InfluenceStrength * distanceFactor * zone.EffectMultiplier

		// Reduce influence if contested
		if zone.IsContested {
			effectiveInfluence *= (1.0 - zone.ContestProgress*0.5)
		}

		if effectiveInfluence > strongestInfluence {
			strongestInfluence = effectiveInfluence
			strongest = zoneEntity
		}
	}

	return strongest
}

// applyZoneModifiers calculates and applies faction zone effects to an entity.
func (s *FactionTerritoryInfluenceSystem) applyZoneModifiers(entity *Entity, modifier *FactionTerritoryModifierComponent, zone *FactionTerritoryInfluenceComponent) {
	// Get entity's reputation with the controlling faction
	reputation := s.getEntityReputation(entity, zone.FactionID)

	// Update modifier state
	modifier.ActiveFactionID = zone.FactionID
	modifier.ReputationLevel = reputation
	modifier.ZoneX = zone.ZoneX
	modifier.ZoneZ = zone.ZoneZ
	modifier.InFactionZone = true
	modifier.Dirty = false

	// Calculate modifiers based on reputation
	if modifier.IsFriendly() {
		// Friendly: apply bonuses
		friendlyFactor := float64(reputation-50) / 50.0 // 0.0 to 1.0
		modifier.EffectiveDamageModifier = 1.0 + zone.FriendlyDamageBonus*friendlyFactor*zone.InfluenceStrength
		modifier.EffectiveDefenseModifier = 1.0 + zone.FriendlyDefenseBonus*friendlyFactor*zone.InfluenceStrength
		modifier.EffectiveXPModifier = 1.0 + zone.FriendlyXPBonus*friendlyFactor*zone.InfluenceStrength
		modifier.EffectiveHealingModifier = 1.0 + zone.FriendlyHealingBonus*friendlyFactor*zone.InfluenceStrength
		modifier.EffectiveDetectionModifier = 1.0 // No detection penalty for friendlies

		s.logger.WithFields(logrus.Fields{
			"entity_id":   entity.ID,
			"faction":     zone.FactionID,
			"reputation":  reputation,
			"damage_mod":  modifier.EffectiveDamageModifier,
			"defense_mod": modifier.EffectiveDefenseModifier,
		}).Debug("Applied friendly faction zone modifiers")

	} else if modifier.IsHostile() {
		// Hostile: apply penalties
		hostileFactor := float64(-50-reputation) / 50.0 // 0.0 to 1.0
		modifier.EffectiveDamageModifier = 1.0 - zone.HostileDamagePenalty*hostileFactor*zone.InfluenceStrength
		modifier.EffectiveDefenseModifier = 1.0 - zone.HostileDefensePenalty*hostileFactor*zone.InfluenceStrength
		modifier.EffectiveXPModifier = 1.0      // No XP penalty
		modifier.EffectiveHealingModifier = 1.0 // No healing penalty
		modifier.EffectiveDetectionModifier = 1.0 + zone.HostileDetectionBonus*hostileFactor*zone.InfluenceStrength

		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"faction":       zone.FactionID,
			"reputation":    reputation,
			"damage_mod":    modifier.EffectiveDamageModifier,
			"detection_mod": modifier.EffectiveDetectionModifier,
		}).Debug("Applied hostile faction zone modifiers")

	} else {
		// Neutral: minimal effects
		modifier.EffectiveDamageModifier = 1.0
		modifier.EffectiveDefenseModifier = 1.0
		modifier.EffectiveXPModifier = 1.0
		modifier.EffectiveHealingModifier = 1.0
		modifier.EffectiveDetectionModifier = 1.0

		s.logger.WithFields(logrus.Fields{
			"entity_id":  entity.ID,
			"faction":    zone.FactionID,
			"reputation": reputation,
		}).Debug("Applied neutral faction zone modifiers")
	}

	// Apply contested zone reduction
	if zone.IsContested {
		contestFactor := 1.0 - zone.ContestProgress*0.3 // Up to 30% reduction
		modifier.EffectiveDamageModifier *= contestFactor
		modifier.EffectiveDefenseModifier *= contestFactor
		modifier.EffectiveXPModifier *= contestFactor
	}
}

// getEntityReputation returns the entity's reputation with a specific faction.
func (s *FactionTerritoryInfluenceSystem) getEntityReputation(entity *Entity, factionID string) int {
	factionComp, ok := entity.GetComponent("faction")
	if !ok || factionComp == nil {
		return 0 // Default neutral
	}
	fc := factionComp.(*FactionComponent)

	// If entity tracks player faction reputation
	if fc.IsPlayerFaction && fc.FactionID == factionID {
		return fc.Reputation
	}

	// If entity is a member of the same faction
	if fc.FactionID == factionID {
		return 100 // Maximum friendly
	}

	// Check faction system for inter-faction relationships
	if s.factionSystem != nil {
		entityFaction := s.factionSystem.GetFaction(fc.FactionID)
		if entityFaction != nil {
			return entityFaction.GetRelationship(factionID)
		}
	}

	return 0 // Default neutral
}

// clearModifiers resets an entity's faction territory modifiers.
func (s *FactionTerritoryInfluenceSystem) clearModifiers(entity *Entity, modifier *FactionTerritoryModifierComponent) {
	prevFaction := modifier.ActiveFactionID
	modifier.Reset()

	s.logger.WithFields(logrus.Fields{
		"entity_id":    entity.ID,
		"prev_faction": prevFaction,
	}).Debug("Entity left faction influence zone")
}

// CreateFactionZone is a helper to create a faction influence zone entity.
func (s *FactionTerritoryInfluenceSystem) CreateFactionZone(factionID string, zoneX, zoneZ int, radius float64) *Entity {
	if s.world == nil {
		return nil
	}

	zone := s.world.CreateEntity()
	influenceComp := NewFactionTerritoryInfluenceComponent(factionID, zoneX, zoneZ)
	influenceComp.ZoneRadius = radius

	// Set zone position at grid center
	posComp := &PositionComponent{
		X: float64(zoneX) * radius * 2,
		Y: float64(zoneZ) * radius * 2,
	}

	zone.AddComponent(influenceComp)
	zone.AddComponent(posComp)

	s.logger.WithFields(logrus.Fields{
		"zone_id":    zone.ID,
		"faction_id": factionID,
		"zone_x":     zoneX,
		"zone_z":     zoneZ,
		"radius":     radius,
	}).Info("Created faction influence zone")

	return zone
}

// GetEntityZoneModifier returns the active faction zone modifier for an entity, if any.
func (s *FactionTerritoryInfluenceSystem) GetEntityZoneModifier(entity *Entity) *FactionTerritoryModifierComponent {
	modComp, ok := entity.GetComponent("faction_territory_modifier")
	if !ok || modComp == nil {
		return nil
	}
	return modComp.(*FactionTerritoryModifierComponent)
}
