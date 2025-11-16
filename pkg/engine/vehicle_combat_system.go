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
	// Get vehicle speed and position
	vehicleComp, hasVehicle := vehicle.GetComponent("vehicle")
	if !hasVehicle {
		return
	}
	v, ok := vehicleComp.(*VehicleComponent)
	if !ok {
		return
	}

	// Check if can ram (cooldown ready and sufficient speed)
	if !combat.CanRam(v.Speed) {
		return
	}

	posComp, hasPos := vehicle.GetComponent("position")
	if !hasPos {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Check for entities in ramming range (small radius around vehicle)
	rammingRadius := 20.0 // pixels
	for _, target := range entities {
		if target.ID == vehicle.ID {
			continue // Don't ram self
		}

		// Check if target is a valid combat target
		targetHealth, hasHealth := target.GetComponent("health")
		if !hasHealth {
			continue // Can't damage entities without health
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

		// Check if within ramming range
		if distance <= rammingRadius {
			// Calculate ramming damage based on speed
			damage := combat.CalculateRammingDamage(v.Speed)

			// Apply damage to target
			if health, ok := targetHealth.(*HealthComponent); ok {
				health.TakeDamage(damage)
			} else {
				continue
			}

			// Execute ramming attack (set cooldown)
			combat.ExecuteRam()

			// Log ramming attack
			if vcs.logger != nil {
				vcs.logger.WithFields(logrus.Fields{
					"vehicle_id": vehicle.ID,
					"target_id":  target.ID,
					"damage":     damage,
					"speed":      v.Speed,
				}).Debug("Vehicle ramming attack")
			}

			// Vehicle also takes some damage from impact (10% of dealt damage)
			v.TakeDamage(damage * 0.1)

			// Only ram one target per update
			break
		}
	}
}

// processMountedWeapons handles mounted weapon attacks.
func (vcs *VehicleCombatSystem) processMountedWeapons(vehicle *Entity, combat *VehicleCombatComponent, entities []*Entity) {
	// Check if weapon can shoot
	if !combat.CanShoot() {
		return
	}

	// Get vehicle position
	posComp, hasPos := vehicle.GetComponent("position")
	if !hasPos {
		return
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Check if vehicle is player-controlled or has AI targeting
	hasControl := vcs.hasTargetInput(vehicle)
	if !hasControl {
		return // Only shoot when controlled
	}

	// Find nearest target in weapon range
	target := vcs.findNearestTarget(vehicle, pos, combat.WeaponRange, entities)
	if target == nil {
		return // No target in range
	}

	// Fire weapon (create projectile if projectile system available)
	if vcs.projectileSystem != nil {
		// Calculate direction to target
		targetPos, _ := target.GetComponent("position")
		if targetPos == nil {
			return
		}
		tPos, ok := targetPos.(*PositionComponent)
		if !ok {
			return
		}
		dx := tPos.X - pos.X
		dy := tPos.Y - pos.Y
		targetAngle := math.Atan2(dy, dx)

		// Create projectile entity with weapon stats
		// Note: This is a simplified version - full implementation would use
		// projectile system's spawn methods with proper configuration
		vcs.spawnWeaponProjectile(vehicle, pos, targetAngle, combat)
	} else {
		// Direct damage application (fallback if no projectile system)
		targetHealth, hasHealth := target.GetComponent("health")
		if hasHealth {
			if health, ok := targetHealth.(*HealthComponent); ok {
				health.TakeDamage(combat.WeaponDamage)
			}
		}
	}

	// Execute shot (set cooldown)
	combat.ExecuteShot()

	// Log weapon attack
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
// This is a simplified version - full implementation would integrate with
// projectile system properly.
func (vcs *VehicleCombatSystem) spawnWeaponProjectile(vehicle *Entity, pos *PositionComponent, angle float64, combat *VehicleCombatComponent) {
	// INTEGRATION FIX [Category A]: Vehicle Weapon Projectile Spawning
	// Gap: Vehicle mounted weapons should spawn projectile entities, not placeholder
	// Fix: Call world.CreateEntity(), add ProjectileComponent with weapon stats, Position at vehicle
	// Roadmap: ROADMAP_V4.md Phase 21.2 - Vehicle Combat (mounted weapon systems)
	// Integration: ProjectileSystem exists, supports vehicle-spawned projectiles via OwnerID field
	// with appropriate speed, damage, and visual based on weapon type
	if vcs.logger != nil {
		vcs.logger.WithFields(logrus.Fields{
			"vehicle_id":  vehicle.ID,
			"weapon_type": combat.WeaponType,
			"angle":       angle,
		}).Debug("Spawning weapon projectile (placeholder)")
	}
}

// ApplyUpgrades applies all installed upgrades to vehicle stats.
// This should be called when upgrades are added/removed.
func ApplyUpgrades(vehicle *VehicleComponent, combat *VehicleCombatComponent, upgrades *UpgradeSlotComponent) {
	// Track base stats (would need to be stored separately for proper upgrade removal)
	// For now, we'll apply upgrades additively/multiplicatively

	for _, upgrade := range upgrades.Slots {
		switch upgrade.Type {
		case UpgradeSpeed:
			if upgrade.IsMultiplicative {
				vehicle.MaxSpeed *= upgrade.Value
			} else {
				vehicle.MaxSpeed += upgrade.Value
			}

		case UpgradeAcceleration:
			if upgrade.IsMultiplicative {
				vehicle.Acceleration *= upgrade.Value
			} else {
				vehicle.Acceleration += upgrade.Value
			}

		case UpgradeHandling:
			if upgrade.IsMultiplicative {
				vehicle.Handling *= upgrade.Value
			} else {
				vehicle.Handling += upgrade.Value
			}

		case UpgradeDurability:
			if upgrade.IsMultiplicative {
				vehicle.MaxDurability *= upgrade.Value
			} else {
				vehicle.MaxDurability += upgrade.Value
			}
			// Scale current durability proportionally
			ratio := vehicle.Durability / vehicle.MaxDurability
			if upgrade.IsMultiplicative {
				vehicle.Durability *= upgrade.Value
			} else {
				vehicle.Durability = (vehicle.MaxDurability + upgrade.Value) * ratio
			}

		case UpgradeArmor:
			if combat != nil {
				if upgrade.IsMultiplicative {
					combat.ArmorRating *= upgrade.Value
				} else {
					combat.ArmorRating += upgrade.Value
				}
			}

		case UpgradeCapacity:
			if upgrade.IsMultiplicative {
				vehicle.Capacity = int(float64(vehicle.Capacity) * upgrade.Value)
			} else {
				vehicle.Capacity += int(upgrade.Value)
			}

		case UpgradeFuelCapacity:
			if upgrade.IsMultiplicative {
				vehicle.FuelCapacity *= upgrade.Value
			} else {
				vehicle.FuelCapacity += upgrade.Value
			}
			// Scale current fuel proportionally
			ratio := vehicle.FuelAmount / vehicle.FuelCapacity
			if upgrade.IsMultiplicative {
				vehicle.FuelAmount *= upgrade.Value
			} else {
				vehicle.FuelAmount = (vehicle.FuelCapacity + upgrade.Value) * ratio
			}

		case UpgradeWeaponDamage:
			if combat != nil {
				if upgrade.IsMultiplicative {
					combat.WeaponDamage *= upgrade.Value
				} else {
					combat.WeaponDamage += upgrade.Value
				}
			}
		}
	}
}
