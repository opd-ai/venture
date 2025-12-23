// Package engine provides vehicle combat system functionality.
// This file implements VehicleCombatSystem which handles vehicle-based
// combat including ramming damage calculations and mounted weapon attacks.
//
// Phase 21.2: Advanced Vehicle Features
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// VehicleCombatSystem manages vehicle combat interactions.
// It processes vehicles with VehicleCombatComponent to handle ramming
// and mounted weapon attacks.
type VehicleCombatSystem struct {
	world  *World
	logger *logrus.Entry

	// Combat system reference for damage dealing
	combatSystem *CombatSystem

	// Projectile system for mounted weapon projectiles
	projectileSystem *ProjectileSystem
}

// NewVehicleCombatSystem creates a new vehicle combat system.
func NewVehicleCombatSystem(world *World) *VehicleCombatSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system", "vehicle_combat")
		if logEntry.Logger.GetLevel() >= logrus.DebugLevel {
			logEntry.Debug("Vehicle combat system created")
		}
	}
	return &VehicleCombatSystem{
		world:  world,
		logger: logEntry,
	}
}

// SetCombatSystem sets the combat system reference for damage dealing.
func (vcs *VehicleCombatSystem) SetCombatSystem(cs *CombatSystem) {
	vcs.combatSystem = cs
}

// SetProjectileSystem sets the projectile system reference for weapon attacks.
func (vcs *VehicleCombatSystem) SetProjectileSystem(ps *ProjectileSystem) {
	vcs.projectileSystem = ps
}

// Update processes vehicle combat for all entities with vehicle combat components.
func (vcs *VehicleCombatSystem) Update(entities []*Entity, deltaTime float64) {
	for _, vehicle := range entities {
		// Check if entity has vehicle combat component
		combatComp, hasCombat := vehicle.GetComponent("vehicle_combat")
		if !hasCombat {
			continue
		}
		combat, ok := combatComp.(*VehicleCombatComponent)
		if !ok {
			continue
		}

		// Update cooldowns
		combat.UpdateCooldowns(deltaTime)

		// Process ramming damage
		vcs.processRamming(vehicle, combat, entities)

		// Process mounted weapon attacks
		vcs.processMountedWeapons(vehicle, combat, entities)
	}
}

// processRamming handles ramming damage when vehicle collides with entities.
func (vcs *VehicleCombatSystem) processRamming(vehicle *Entity, combat *VehicleCombatComponent, entities []*Entity) {
	vehicleComp, pos := vcs.getVehicleComponents(vehicle, combat)
	if vehicleComp == nil || pos == nil {
		return
	}

	target := vcs.findRammingTarget(vehicle, pos, entities)
	if target == nil {
		return
	}

	vcs.applyRammingDamage(vehicle, vehicleComp, combat, target)
}

// getVehicleComponents retrieves and validates vehicle and position components.
func (vcs *VehicleCombatSystem) getVehicleComponents(vehicle *Entity, combat *VehicleCombatComponent) (*VehicleComponent, *PositionComponent) {
	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if !hasVehicle {
		return nil, nil
	}
	v, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return nil, nil
	}

	if !combat.CanRam(v.Speed) {
		return nil, nil
	}

	posComp, hasPos := vehicle.GetComponent("position")
	if !hasPos {
		return nil, nil
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil, nil
	}

	return v, pos
}

// findRammingTarget finds a valid target within ramming range.
func (vcs *VehicleCombatSystem) findRammingTarget(vehicle *Entity, pos *PositionComponent, entities []*Entity) *Entity {
	const rammingRadius = 20.0 // pixels

	for _, target := range entities {
		if target.ID == vehicle.ID {
			continue
		}

		if !vcs.isValidRammingTarget(target, pos, rammingRadius) {
			continue
		}

		return target
	}

	return nil
}

// isValidRammingTarget checks if target is valid and within ramming range.
func (vcs *VehicleCombatSystem) isValidRammingTarget(target *Entity, vehiclePos *PositionComponent, rammingRadius float64) bool {
	targetHealth, hasHealth := target.GetComponent("health")
	if !hasHealth {
		return false
	}

	targetPos, hasTargetPos := target.GetComponent("position")
	if !hasTargetPos {
		return false
	}
	tPos, ok := targetPos.(*PositionComponent)
	if !ok {
		return false
	}

	dx := tPos.X - vehiclePos.X
	dy := tPos.Y - vehiclePos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= rammingRadius && targetHealth != nil
}

// applyRammingDamage applies damage to target and vehicle from ramming.
func (vcs *VehicleCombatSystem) applyRammingDamage(vehicle *Entity, vehicleComp *VehicleComponent, combat *VehicleCombatComponent, target *Entity) {
	damage := combat.CalculateRammingDamage(vehicleComp.Speed)

	targetHealth, _ := target.GetComponent("health")
	if health, ok := targetHealth.(*HealthComponent); ok {
		health.TakeDamage(damage)
	}

	combat.ExecuteRam()

	if vcs.logger != nil {
		vcs.logger.WithFields(logrus.Fields{
			"vehicle_id": vehicle.ID,
			"target_id":  target.ID,
			"damage":     damage,
			"speed":      vehicleComp.Speed,
		}).Debug("Vehicle ramming attack")
	}

	vehicleComp.TakeDamage(damage * 0.1)
}

// processMountedWeapons handles mounted weapon attacks.
func (vcs *VehicleCombatSystem) processMountedWeapons(vehicle *Entity, combat *VehicleCombatComponent, entities []*Entity) {
	if !combat.CanShoot() {
		return
	}

	pos := vcs.getVehiclePosition(vehicle)
	if pos == nil {
		return
	}

	if !vcs.hasTargetInput(vehicle) {
		return
	}

	target := vcs.findNearestTarget(vehicle, pos, combat.WeaponRange, entities)
	if target == nil {
		return
	}

	vcs.fireWeapon(vehicle, pos, target, combat)
	combat.ExecuteShot()
	vcs.logWeaponAttack(vehicle, target, combat)
}

// getVehiclePosition retrieves and validates vehicle position component
func (vcs *VehicleCombatSystem) getVehiclePosition(vehicle *Entity) *PositionComponent {
	posComp, hasPos := vehicle.GetComponent("position")
	if !hasPos {
		return nil
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return nil
	}
	return pos
}

// fireWeapon fires weapon at target using projectile system or direct damage
func (vcs *VehicleCombatSystem) fireWeapon(vehicle *Entity, pos *PositionComponent, target *Entity, combat *VehicleCombatComponent) {
	if vcs.projectileSystem != nil {
		vcs.fireProjectileWeapon(vehicle, pos, target, combat)
	} else {
		vcs.applyDirectDamage(target, combat.WeaponDamage)
	}
}

// fireProjectileWeapon spawns projectile towards target
func (vcs *VehicleCombatSystem) fireProjectileWeapon(vehicle *Entity, pos *PositionComponent, target *Entity, combat *VehicleCombatComponent) {
	targetPos := vcs.getTargetPosition(target)
	if targetPos == nil {
		return
	}

	dx := targetPos.X - pos.X
	dy := targetPos.Y - pos.Y
	targetAngle := math.Atan2(dy, dx)
	vcs.spawnWeaponProjectile(vehicle, pos, targetAngle, combat)
}

// getTargetPosition retrieves target position component
func (vcs *VehicleCombatSystem) getTargetPosition(target *Entity) *PositionComponent {
	targetPos, _ := target.GetComponent("position")
	if targetPos == nil {
		return nil
	}
	tPos, ok := targetPos.(*PositionComponent)
	if !ok {
		return nil
	}
	return tPos
}

// applyDirectDamage applies damage directly to target health
func (vcs *VehicleCombatSystem) applyDirectDamage(target *Entity, damage float64) {
	targetHealth, hasHealth := target.GetComponent("health")
	if hasHealth {
		if health, ok := targetHealth.(*HealthComponent); ok {
			health.TakeDamage(damage)
		}
	}
}

// logWeaponAttack logs vehicle weapon attack event
func (vcs *VehicleCombatSystem) logWeaponAttack(vehicle, target *Entity, combat *VehicleCombatComponent) {
	if vcs.logger != nil && target != nil {
		vcs.logger.WithFields(logrus.Fields{
			"vehicle_id":  vehicle.ID,
			"target_id":   target.ID,
			"damage":      combat.WeaponDamage,
			"weapon_type": combat.WeaponType,
		}).Debug("Vehicle weapon attack")
	}
}

// hasTargetInput checks if vehicle has active targeting (player or AI).
func (vcs *VehicleCombatSystem) hasTargetInput(vehicle *Entity) bool {
	// Check for player input (player-controlled vehicle)
	if _, hasInput := vehicle.GetComponent("input"); hasInput {
		return true
	}

	// Check if mounted by player
	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if hasVehicle {
		if v, ok := vehicleComp.(*VehicleComponent); ok {
			if v.CurrentPassengers > 0 {
				return true
			}
		}
	}

	// Check for AI component (AI-controlled vehicle)
	if _, hasAI := vehicle.GetComponent("ai"); hasAI {
		return true
	}

	return false
}

// findNearestTarget finds the closest enemy entity within weapon range.
func (vcs *VehicleCombatSystem) findNearestTarget(vehicle *Entity, pos *PositionComponent, weaponRange float64, entities []*Entity) *Entity {
	var nearestTarget *Entity
	nearestDistance := weaponRange + 1.0 // Start beyond range

	for _, target := range entities {
		if target.ID == vehicle.ID {
			continue // Don't target self
		}

		// Check if target has health (is attackable)
		if _, hasHealth := target.GetComponent("health"); !hasHealth {
			continue
		}

		// Get target position
		targetPos, hasTargetPos := target.GetComponent("position")
		if !hasTargetPos {
			continue
		}
		tPos, ok := targetPos.(*PositionComponent)
		if !ok {
			continue
		}

		// Calculate distance
		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		// Check if within range and closer than current nearest
		if distance <= weaponRange && distance < nearestDistance {
			nearestTarget = target
			nearestDistance = distance
		}
	}

	return nearestTarget
}

// spawnWeaponProjectile creates a projectile for mounted weapon attack.
// This method spawns actual projectile entities using the projectile system
// when available, with appropriate speed, damage, and visual based on weapon type.
func (vcs *VehicleCombatSystem) spawnWeaponProjectile(vehicle *Entity, pos *PositionComponent, angle float64, combat *VehicleCombatComponent) {
	// Validate projectile system and world are available
	if vcs.projectileSystem == nil || vcs.world == nil {
		if vcs.logger != nil {
			vcs.logger.WithFields(logrus.Fields{
				"vehicle_id":  vehicle.ID,
				"weapon_type": combat.WeaponType,
			}).Debug("Cannot spawn projectile: projectile system or world unavailable")
		}
		return
	}

	// Calculate velocity from angle and projectile speed
	projectileSpeed := combat.WeaponProjectileSpeed
	if projectileSpeed <= 0 {
		projectileSpeed = 300.0 // Default speed if not configured
	}
	vx := math.Cos(angle) * projectileSpeed
	vy := math.Sin(angle) * projectileSpeed

	// Map weapon type to projectile type for visual representation
	projectileType := vcs.weaponTypeToProjectileType(combat.WeaponType)

	// Calculate projectile lifetime based on weapon range and speed
	lifetime := combat.WeaponRange / projectileSpeed
	if lifetime <= 0 {
		lifetime = 2.0 // Default 2 seconds if range not configured
	}

	// Create projectile component with weapon stats
	projComp := NewProjectileComponent(
		combat.WeaponDamage, // damage
		projectileSpeed,    // speed (stored for reference)
		lifetime,           // lifetime
		projectileType,     // projectile type for visual
		vehicle.ID,         // owner ID
	)

	// Spawn projectile using the projectile system
	projectile := vcs.projectileSystem.SpawnProjectile(pos.X, pos.Y, vx, vy, projComp)

	if vcs.logger != nil {
		projectileID := uint64(0)
		if projectile != nil {
			projectileID = projectile.ID
		}
		vcs.logger.WithFields(logrus.Fields{
			"vehicle_id":      vehicle.ID,
			"projectile_id":   projectileID,
			"weapon_type":     combat.WeaponType,
			"projectile_type": projectileType,
			"angle":           angle,
			"damage":          combat.WeaponDamage,
			"speed":           projectileSpeed,
		}).Debug("Vehicle spawned weapon projectile")
	}
}

// weaponTypeToProjectileType maps vehicle weapon types to projectile visual types.
func (vcs *VehicleCombatSystem) weaponTypeToProjectileType(weaponType string) string {
	switch weaponType {
	case "Cannon":
		return "bullet"
	case "MachineGun":
		return "bullet"
	case "Laser":
		return "lightning_bolt"
	case "Magic":
		return "magic_missile"
	case "Ballista":
		return "arrow"
	default:
		return "bullet"
	}
}

// ApplyUpgrades applies all installed upgrades to vehicle stats.
// This should be called when upgrades are added/removed.
func ApplyUpgrades(vehicle *VehicleComponent, combat *VehicleCombatComponent, upgrades *UpgradeSlotComponent) {
	for _, upgrade := range upgrades.Slots {
		applyUpgradeByType(vehicle, combat, upgrade)
	}
}

// applyUpgradeByType routes upgrade application to the appropriate handler.
func applyUpgradeByType(vehicle *VehicleComponent, combat *VehicleCombatComponent, upgrade *VehicleUpgrade) {
	switch upgrade.Type {
	case UpgradeSpeed:
		applySimpleUpgrade(&vehicle.MaxSpeed, upgrade)
	case UpgradeAcceleration:
		applySimpleUpgrade(&vehicle.Acceleration, upgrade)
	case UpgradeHandling:
		applySimpleUpgrade(&vehicle.Handling, upgrade)
	case UpgradeDurability:
		applyDurabilityUpgrade(vehicle, upgrade)
	case UpgradeArmor:
		applyArmorUpgrade(combat, upgrade)
	case UpgradeCapacity:
		applyCapacityUpgrade(vehicle, upgrade)
	case UpgradeFuelCapacity:
		applyFuelCapacityUpgrade(vehicle, upgrade)
	case UpgradeWeaponDamage:
		applyWeaponDamageUpgrade(combat, upgrade)
	}
}

// applySimpleUpgrade applies an upgrade value to a float64 stat.
func applySimpleUpgrade(stat *float64, upgrade *VehicleUpgrade) {
	if upgrade.IsMultiplicative {
		*stat *= upgrade.Value
	} else {
		*stat += upgrade.Value
	}
}

// applyDurabilityUpgrade applies durability upgrade with proportional scaling.
func applyDurabilityUpgrade(vehicle *VehicleComponent, upgrade *VehicleUpgrade) {
	ratio := vehicle.Durability / vehicle.MaxDurability
	applySimpleUpgrade(&vehicle.MaxDurability, upgrade)
	if upgrade.IsMultiplicative {
		vehicle.Durability *= upgrade.Value
	} else {
		vehicle.Durability = (vehicle.MaxDurability + upgrade.Value) * ratio
	}
}

// applyArmorUpgrade applies armor upgrade to combat component.
func applyArmorUpgrade(combat *VehicleCombatComponent, upgrade *VehicleUpgrade) {
	if combat != nil {
		applySimpleUpgrade(&combat.ArmorRating, upgrade)
	}
}

// applyCapacityUpgrade applies capacity upgrade to vehicle.
func applyCapacityUpgrade(vehicle *VehicleComponent, upgrade *VehicleUpgrade) {
	if upgrade.IsMultiplicative {
		vehicle.Capacity = int(float64(vehicle.Capacity) * upgrade.Value)
	} else {
		vehicle.Capacity += int(upgrade.Value)
	}
}

// applyFuelCapacityUpgrade applies fuel capacity upgrade with proportional scaling.
func applyFuelCapacityUpgrade(vehicle *VehicleComponent, upgrade *VehicleUpgrade) {
	ratio := vehicle.FuelAmount / vehicle.FuelCapacity
	applySimpleUpgrade(&vehicle.FuelCapacity, upgrade)
	if upgrade.IsMultiplicative {
		vehicle.FuelAmount *= upgrade.Value
	} else {
		vehicle.FuelAmount = (vehicle.FuelCapacity + upgrade.Value) * ratio
	}
}

// applyWeaponDamageUpgrade applies weapon damage upgrade to combat component.
func applyWeaponDamageUpgrade(combat *VehicleCombatComponent, upgrade *VehicleUpgrade) {
	if combat != nil {
		applySimpleUpgrade(&combat.WeaponDamage, upgrade)
	}
}
