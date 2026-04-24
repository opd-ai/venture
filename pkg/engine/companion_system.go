package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// CompanionSystem manages companion AI, following, commands, and bonding.
// Phase 22.2: Companion System
type CompanionSystem struct {
	world  *World
	logger *logrus.Entry

	// scoutTimers tracks accumulated time per-companion for direction changes.
	// Used by executeScout to cycle through four cardinal directions.
	scoutTimers     map[uint64]float64
	scoutDirections map[uint64]int
}

// NewCompanionSystem creates a new companion management system.
func NewCompanionSystem(world *World) *CompanionSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system": "companion",
	})
	return &CompanionSystem{
		world:           world,
		logger:          logger,
		scoutTimers:     make(map[uint64]float64),
		scoutDirections: make(map[uint64]int),
	}
}

// Update processes companion AI behavior, following, and bonding.
func (s *CompanionSystem) Update(deltaTime float64) {
	companions := s.world.GetEntitiesWith("companion")

	for _, companion := range companions {
		comp, ok := companion.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp, ok := comp.(*CompanionComponent)
		if !ok {
			continue
		}

		// Get owner entity
		owner, _ := s.world.GetEntity(companionComp.OwnerID)
		if owner == nil {
			// Owner doesn't exist, companion becomes idle
			continue
		}

		// Update bonding time
		s.updateBonding(companion, companionComp, owner, deltaTime)

		// Process current command
		s.processCommand(companion, companionComp, owner, deltaTime)

		// Apply bonding perks
		s.applyBondingPerks(companion, companionComp)
	}
}

// updateBonding increases bonding time and unlocks perks.
func (s *CompanionSystem) updateBonding(companion *Entity, companionComp *CompanionComponent, owner *Entity, deltaTime float64) {
	// Get positions
	compPos, hasCompanionPos := companion.GetComponent("position")
	if !hasCompanionPos {
		return
	}
	companionPos, ok := compPos.(*PositionComponent)
	if !ok {
		return
	}

	ownPos, hasOwnerPos := owner.GetComponent("position")
	if !hasOwnerPos {
		return
	}
	ownerPos, ok := ownPos.(*PositionComponent)
	if !ok {
		return
	}

	// Calculate distance to owner
	dx := companionPos.X - ownerPos.X
	dy := companionPos.Y - ownerPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Only increase bonding when close to owner (within 200 pixels)
	if distance < 200 {
		companionComp.TimeWithOwner += deltaTime

		// Increase loyalty gradually when bonding
		loyaltyIncrease := 0.1 * deltaTime // 0.1 per second
		companionComp.Loyalty += loyaltyIncrease
		if companionComp.Loyalty > 100 {
			companionComp.Loyalty = 100
		}

		// Unlock perks based on time spent with owner
		s.checkPerkUnlocks(companionComp)
	}
}

// checkPerkUnlocks unlocks bonding perks based on time and loyalty.
func (s *CompanionSystem) checkPerkUnlocks(companionComp *CompanionComponent) {
	// Perk thresholds (time in seconds, loyalty level)
	perks := []struct {
		perk         BondingPerk
		timeRequired float64
		loyaltyReq   float64
	}{
		{PerkExtraHealth, 300, 20},       // 5 minutes, 20% loyalty
		{PerkExtraDamage, 600, 40},       // 10 minutes, 40% loyalty
		{PerkFasterLearning, 1200, 60},   // 20 minutes, 60% loyalty
		{PerkLoyalGuard, 1800, 75},       // 30 minutes, 75% loyalty
		{PerkSharedExperience, 2400, 85}, // 40 minutes, 85% loyalty
		{PerkAutoRevive, 3600, 95},       // 60 minutes, 95% loyalty
	}

	for _, perkInfo := range perks {
		// Check if perk already unlocked
		hasPerk := false
		for _, p := range companionComp.BondingPerks {
			if p == perkInfo.perk {
				hasPerk = true
				break
			}
		}

		// Unlock if requirements met and not already unlocked
		if !hasPerk && companionComp.TimeWithOwner >= perkInfo.timeRequired && companionComp.Loyalty >= perkInfo.loyaltyReq {
			companionComp.BondingPerks = append(companionComp.BondingPerks, perkInfo.perk)
			s.logger.WithFields(logrus.Fields{
				"companion_id": companionComp.OwnerID,
				"perk":         perkInfo.perk.String(),
			}).Info("Companion unlocked bonding perk")
		}
	}
}

// processCommand executes the companion's current command.
func (s *CompanionSystem) processCommand(companion *Entity, companionComp *CompanionComponent, owner *Entity, deltaTime float64) {
	// Get the active command (first in list)
	if len(companionComp.Commands) == 0 {
		// Default to Follow if no commands
		companionComp.Commands = []CommandType{CommandFollow}
	}

	currentCommand := companionComp.Commands[0]

	switch currentCommand {
	case CommandFollow:
		s.executeFollow(companion, companionComp, owner, deltaTime)
	case CommandStay:
		s.executeStay(companion, companionComp)
	case CommandAttack:
		s.executeAttack(companion, companionComp, owner)
	case CommandDefend:
		s.executeDefend(companion, companionComp, owner)
	case CommandGather:
		s.executeGather(companion, companionComp)
	case CommandScout:
		s.executeScout(companion, companionComp, owner, deltaTime)
	}
}

// executeFollow makes the companion follow the owner.
func (s *CompanionSystem) executeFollow(companion *Entity, companionComp *CompanionComponent, owner *Entity, deltaTime float64) {
	compPos, hasCompanionPos := companion.GetComponent("position")
	companionPos, _ := compPos.(*PositionComponent)
	ownPos, hasOwnerPos := owner.GetComponent("position")
	ownerPos, _ := ownPos.(*PositionComponent)

	if !hasCompanionPos || !hasOwnerPos {
		return
	}

	// Calculate direction to owner
	dx := ownerPos.X - companionPos.X
	dy := ownerPos.Y - companionPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Only move if beyond follow distance (50 pixels)
	if distance > 50 {
		// Normalize direction
		if distance > 0 {
			dx /= distance
			dy /= distance
		}

		// Set velocity to move toward owner
		moveSpeed := 100.0 // Base follow speed
		velComp, hasVelocity := companion.GetComponent("velocity")
		if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
			velocityComp.VX = dx * moveSpeed
			velocityComp.VY = dy * moveSpeed
		}
	} else {
		// Close enough, stop moving
		velComp, hasVelocity := companion.GetComponent("velocity")
		if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
			velocityComp.VX = 0
			velocityComp.VY = 0
		}
	}
}

// executeStay makes the companion stay in place.
func (s *CompanionSystem) executeStay(companion *Entity, companionComp *CompanionComponent) {
	velComp, hasVelocity := companion.GetComponent("velocity")
	if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
		velocityComp.VX = 0
		velocityComp.VY = 0
	}
}

// executeAttack makes the companion attack nearby enemies.
func (s *CompanionSystem) executeAttack(companion *Entity, companionComp *CompanionComponent, owner *Entity) {
	companionPos := s.getCompanionPosition(companion)
	if companionPos == nil {
		return
	}

	nearestEnemy := s.findNearestEnemy(companion, owner, companionPos)
	if nearestEnemy != nil {
		s.engageEnemy(companion, nearestEnemy, companionPos)
	}
}

// getCompanionPosition retrieves the position component of a companion.
func (s *CompanionSystem) getCompanionPosition(companion *Entity) *PositionComponent {
	compPos, hasPos := companion.GetComponent("position")
	if !hasPos {
		return nil
	}
	companionPos, ok := compPos.(*PositionComponent)
	if !ok {
		return nil
	}
	return companionPos
}

// findNearestEnemy finds the nearest hostile entity within attack range.
func (s *CompanionSystem) findNearestEnemy(companion, owner *Entity, companionPos *PositionComponent) *Entity {
	entities := s.world.GetEntitiesWith("position", "health")
	var nearestEnemy *Entity
	minDistance := 300.0

	for _, entity := range entities {
		if entity.ID == companion.ID || entity.ID == owner.ID {
			continue
		}

		distance := s.calculateDistanceToEntity(entity, companionPos)
		if distance >= 0 && distance < minDistance {
			minDistance = distance
			nearestEnemy = entity
		}
	}

	return nearestEnemy
}

// calculateDistanceToEntity calculates distance from companion to entity.
func (s *CompanionSystem) calculateDistanceToEntity(entity *Entity, companionPos *PositionComponent) float64 {
	entPos, _ := entity.GetComponent("position")
	entityPos, ok := entPos.(*PositionComponent)
	if !ok {
		return -1
	}
	dx := entityPos.X - companionPos.X
	dy := entityPos.Y - companionPos.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// engageEnemy moves toward or attacks an enemy based on distance.
func (s *CompanionSystem) engageEnemy(companion, enemy *Entity, companionPos *PositionComponent) {
	enPos, _ := enemy.GetComponent("position")
	enemyPos, ok := enPos.(*PositionComponent)
	if !ok {
		return
	}

	dx := enemyPos.X - companionPos.X
	dy := enemyPos.Y - companionPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance > 32 {
		s.moveTowardsTarget(companion, dx, dy, distance)
	}
}

// moveTowardsTarget updates companion velocity to move toward target.
func (s *CompanionSystem) moveTowardsTarget(companion *Entity, dx, dy, distance float64) {
	if distance > 0 {
		dx /= distance
		dy /= distance
	}

	velComp, hasVelocity := companion.GetComponent("velocity")
	if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
		velocityComp.VX = dx * 150.0
		velocityComp.VY = dy * 150.0
	}
}

// executeDefend makes the companion defend the owner.
func (s *CompanionSystem) executeDefend(companion *Entity, companionComp *CompanionComponent, owner *Entity) {
	// Stay close to owner and intercept threats
	// Similar to Follow but with tighter radius
	compPos, hasCompanionPos := companion.GetComponent("position")
	companionPos, _ := compPos.(*PositionComponent)
	ownPos, hasOwnerPos := owner.GetComponent("position")
	ownerPos, _ := ownPos.(*PositionComponent)

	if !hasCompanionPos || !hasOwnerPos {
		return
	}

	dx := ownerPos.X - companionPos.X
	dy := ownerPos.Y - companionPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Maintain defensive position (within 64 pixels)
	if distance > 64 || distance < 32 {
		if distance > 0 {
			dx /= distance
			dy /= distance
		}
		targetDistance := 48.0 // Optimal defensive distance
		velComp, hasVelocity := companion.GetComponent("velocity")
		if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
			if distance > targetDistance {
				velocityComp.VX = dx * 100.0
				velocityComp.VY = dy * 100.0
			} else {
				// Too close, back away slightly
				velocityComp.VX = -dx * 50.0
				velocityComp.VY = -dy * 50.0
			}
		}
	}
}

// executeGather makes the companion collect nearby items.
// Companions search for item_entity components within gather range, move toward
// the nearest item, and pick it up when close enough. Items are added to the
// companion's inventory if it has a CompanionInventoryComponent.
func (s *CompanionSystem) executeGather(companion *Entity, companionComp *CompanionComponent) {
	compPos, hasPos := companion.GetComponent("position")
	companionPos, ok := compPos.(*PositionComponent)
	if !hasPos || !ok {
		return
	}

	// Find nearest item entity within gather range (200 pixels)
	const gatherRange = 200.0
	const pickupRange = 32.0
	nearestItem := s.findNearestItem(companionPos, gatherRange)

	if nearestItem == nil {
		// No items nearby, stop moving
		s.stopCompanionMovement(companion)
		return
	}

	// Get item position
	itemPosComp, hasItemPos := nearestItem.GetComponent("position")
	if !hasItemPos {
		s.stopCompanionMovement(companion)
		return
	}
	itemPos := itemPosComp.(*PositionComponent)

	// Calculate distance to item
	dx := itemPos.X - companionPos.X
	dy := itemPos.Y - companionPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// If close enough, pick up the item
	if distance <= pickupRange {
		s.pickupItem(companion, nearestItem)
		s.stopCompanionMovement(companion)
		return
	}

	// Move toward the item
	s.moveCompanionToward(companion, dx, dy, distance, 100.0)
}

// findNearestItem finds the closest item_entity within the specified range.
func (s *CompanionSystem) findNearestItem(companionPos *PositionComponent, maxRange float64) *Entity {
	items := s.world.GetEntitiesWith("item_entity", "position")
	var nearestItem *Entity
	minDistance := maxRange

	for _, itemEntity := range items {
		itemPosComp, ok := itemEntity.GetComponent("position")
		if !ok {
			continue
		}
		itemPos := itemPosComp.(*PositionComponent)

		dx := itemPos.X - companionPos.X
		dy := itemPos.Y - companionPos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance < minDistance {
			minDistance = distance
			nearestItem = itemEntity
		}
	}

	return nearestItem
}

// pickupItem adds the item to companion inventory and removes it from world.
func (s *CompanionSystem) pickupItem(companion, itemEntity *Entity) {
	// Get item data
	itemCompRaw, hasItem := itemEntity.GetComponent("item_entity")
	if !hasItem {
		return
	}
	itemEntityComp := itemCompRaw.(*ItemEntityComponent)
	if itemEntityComp.Item == nil {
		return
	}

	// Try to add to companion inventory
	invCompRaw, hasInv := companion.GetComponent("companioninventory")
	if hasInv {
		invComp := invCompRaw.(*CompanionInventoryComponent)
		if invComp.AddItem(itemEntityComp.Item) {
			s.world.RemoveEntity(itemEntity.ID)
			s.logger.WithFields(logrus.Fields{
				"companion": companion.ID,
				"item":      itemEntityComp.Item.Name,
			}).Debug("companion gathered item")
			return
		}
	}

	// If no inventory or full, try to transfer directly to owner
	companionCompRaw, hasCompanion := companion.GetComponent("companion")
	if !hasCompanion {
		return
	}
	companionComp := companionCompRaw.(*CompanionComponent)

	owner, ownerExists := s.world.GetEntity(companionComp.OwnerID)
	if !ownerExists || owner == nil {
		return
	}

	ownerInvRaw, hasOwnerInv := owner.GetComponent("inventory")
	if !hasOwnerInv {
		return
	}
	ownerInv := ownerInvRaw.(*InventoryComponent)

	if ownerInv.AddItem(itemEntityComp.Item) {
		s.world.RemoveEntity(itemEntity.ID)
		s.logger.WithFields(logrus.Fields{
			"companion": companion.ID,
			"owner":     owner.ID,
			"item":      itemEntityComp.Item.Name,
		}).Debug("companion gathered item for owner")
	}
}

// stopCompanionMovement sets companion velocity to zero.
func (s *CompanionSystem) stopCompanionMovement(companion *Entity) {
	velComp, hasVelocity := companion.GetComponent("velocity")
	if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
		velocityComp.VX = 0
		velocityComp.VY = 0
	}
}

// moveCompanionToward sets companion velocity to move toward a target.
func (s *CompanionSystem) moveCompanionToward(companion *Entity, dx, dy, distance, speed float64) {
	if distance <= 0 {
		return
	}
	// Normalize direction
	dx /= distance
	dy /= distance

	velComp, hasVelocity := companion.GetComponent("velocity")
	if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
		velocityComp.VX = dx * speed
		velocityComp.VY = dy * speed
	}
}

// executeScout makes the companion explore nearby areas.
// Cycles through four cardinal directions every scoutIntervalSecs seconds to avoid
// moving diagonally north-east indefinitely (G19 fix).
func (s *CompanionSystem) executeScout(companion *Entity, _ *CompanionComponent, _ *Entity, deltaTime float64) {
	const scoutSpeed = 80.0
	const scoutIntervalSecs = 3.0 // seconds per direction segment

	// Cardinal direction vectors: E, S, W, N
	dirs := [4][2]float64{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}

	id := companion.ID
	s.scoutTimers[id] += deltaTime
	if s.scoutTimers[id] >= scoutIntervalSecs {
		s.scoutTimers[id] = 0
		s.scoutDirections[id] = (s.scoutDirections[id] + 1) % 4
	}

	dir := dirs[s.scoutDirections[id]]
	velComp, hasVelocity := companion.GetComponent("velocity")
	if velocityComp, ok := velComp.(*VelocityComponent); hasVelocity && ok {
		velocityComp.VX = dir[0] * scoutSpeed
		velocityComp.VY = dir[1] * scoutSpeed
	}
}

// applyBondingPerks applies stat bonuses from unlocked perks.
func (s *CompanionSystem) applyBondingPerks(companion *Entity, companionComp *CompanionComponent) {
	comphasHealth, hasHealth := companion.GetComponent("health")
	if !hasHealth {
		return
	}
	healthComp := comphasHealth.(*HealthComponent)

	for _, perk := range companionComp.BondingPerks {
		switch perk {
		case PerkExtraHealth:
			// Apply 20% max health bonus
			if hasHealth {
				baseMaxHealth := 100.0 // Would come from companion stats
				healthComp.Max = baseMaxHealth * 1.2
			}

		case PerkExtraDamage:
			// Apply 15% damage bonus
			// Would modify combat stats in full implementation

		case PerkFasterLearning:
			// Double XP gain
			// Would be applied during XP award in full implementation

		case PerkLoyalGuard:
			// 30% damage reduction
			// Would modify defense stats in full implementation

		case PerkSharedExperience:
			// 10% XP share to owner
			// Would be applied during XP award in full implementation

		case PerkAutoRevive:
			// One-time revival per day
			// Would be checked in death system in full implementation
		}
	}
}

// IssueCommand adds a command to the companion's command queue.
func (s *CompanionSystem) IssueCommand(companion *Entity, command CommandType) error {
	compok, ok := companion.GetComponent("companion")
	if !ok {
		return ErrComponentNotFound
	}
	companionComp, ok := compok.(*CompanionComponent)
	if !ok {
		return ErrComponentNotFound
	}

	// Replace current command with new one
	companionComp.Commands = []CommandType{command}

	s.logger.WithFields(logrus.Fields{
		"companion": companion.ID,
		"command":   command,
	}).Info("Companion received command")

	return nil
}

// GetLoyalty returns the companion's current loyalty level.
func (s *CompanionSystem) GetLoyalty(companion *Entity) float64 {
	compok, ok := companion.GetComponent("companion")
	if !ok {
		return 0
	}
	companionComp, ok := compok.(*CompanionComponent)
	if !ok {
		return 0
	}
	return companionComp.Loyalty
}

// HasPerk checks if a companion has unlocked a specific bonding perk.
func (s *CompanionSystem) HasPerk(companion *Entity, perk BondingPerk) bool {
	compok, ok := companion.GetComponent("companion")
	if !ok {
		return false
	}
	companionComp, ok := compok.(*CompanionComponent)
	if !ok {
		return false
	}

	for _, p := range companionComp.BondingPerks {
		if p == perk {
			return true
		}
	}
	return false
}
