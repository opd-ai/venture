package engine

import (
	"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
	"github.com/sirupsen/logrus"
)

// GuildVehicleSystem manages guild vehicle fleets with formation bonuses and shared access.
// Integrates V8 Guilds, V4 Vehicles, and V8 Vehicle Physics for coordinated fleet warfare.
type GuildVehicleSystem struct {
	world   *World
	manager *guild_vehicle.FleetManager
	logger  *logrus.Entry
}

// NewGuildVehicleSystem creates a new guild vehicle system.
func NewGuildVehicleSystem(world *World) *GuildVehicleSystem {
	logger := logrus.WithField("system", "guild_vehicle")
	sys := &GuildVehicleSystem{
		world:   world,
		manager: guild_vehicle.NewFleetManager(),
		logger:  logger,
	}
	// Wire self as VehicleSyncer so ECS fleet components stay in sync
	// whenever FleetManager.AddVehicle or RemoveVehicle is called.
	sys.manager.SetVehicleSyncer(sys)
	return sys
}

// Update processes entities with guild vehicle fleet components.
// Two-pass: first collects leader positions, then applies formation bonuses
// and proportional steering toward formation target positions.
func (s *GuildVehicleSystem) Update(entities []*Entity, deltaTime float64) {
	type member struct {
		entity    *Entity
		fleetKey  string
		fleetComp *guild_vehicle.GuildVehicleFleetComponent
	}

	// Pass 1: collect leader (slot 0) positions keyed by "guildID/fleetID".
	leaderPos := make(map[string]*PositionComponent)
	members := make([]member, 0, len(entities))

	for _, entity := range entities {
		if !entity.HasComponent("guild_vehicle_fleet") {
			continue
		}
		comp, ok := entity.GetComponent("guild_vehicle_fleet")
		if !ok {
			continue
		}
		fleetComp, ok := comp.(*guild_vehicle.GuildVehicleFleetComponent)
		if !ok {
			s.logger.WithFields(logrus.Fields{
				"entityID":       entity.ID,
				"component_type": "guild_vehicle_fleet",
			}).Warn("Component type assertion failed")
			continue
		}

		key := fleetComp.GuildID + "/" + fleetComp.FleetID
		if fleetComp.FormationPosition == 0 {
			if pos := entity.GetPosition(); pos != nil {
				leaderPos[key] = pos
			}
		}
		members = append(members, member{entity, key, fleetComp})
	}

	// Pass 2: apply formation bonuses and steer followers toward their slots.
	for _, m := range members {
		s.applyFormationBonuses(m.entity, m.fleetComp)

		if m.fleetComp.FormationPosition <= 0 {
			continue // leader needs no steering
		}
		lp, hasLeader := leaderPos[m.fleetKey]
		if !hasLeader {
			continue
		}
		offsets := s.manager.GetFormationOffsets(m.fleetComp.GuildID, m.fleetComp.FleetID)
		slot := m.fleetComp.FormationPosition
		if slot >= len(offsets) {
			continue
		}
		off := offsets[slot]
		targetX := lp.X + off.OffsetX
		targetY := lp.Y + off.OffsetY

		vel := m.entity.GetVelocity()
		currPos := m.entity.GetPosition()
		if vel == nil || currPos == nil {
			continue
		}
		// Use additive, deltaTime-scaled steering so formation forces blend with
		// existing momentum instead of overriding it abruptly.
		const steerStrength = 4.0
		vel.VX += (targetX - currPos.X) * steerStrength * deltaTime
		vel.VY += (targetY - currPos.Y) * steerStrength * deltaTime
	}
}

// SyncVehicleFleetComponent attaches or updates the GuildVehicleFleetComponent on
// the entity identified by vehicleID and mirrors FleetID onto VehicleComponent.
// Implements guild_vehicle.VehicleSyncer.
func (s *GuildVehicleSystem) SyncVehicleFleetComponent(vehicleID uint64, comp *guild_vehicle.GuildVehicleFleetComponent) {
	entity, ok := s.world.GetEntity(vehicleID)
	if !ok {
		s.logger.WithField("vehicleID", vehicleID).Warn("SyncVehicleFleetComponent: entity not found")
		return
	}
	entity.AddComponent(comp)
	if raw, exists := entity.GetComponent("vehicle"); exists {
		if v, ok := raw.(*VehicleComponent); ok {
			v.FleetID = comp.FleetID
		}
	}
	s.logger.WithFields(logrus.Fields{
		"vehicleID": vehicleID,
		"guildID":   comp.GuildID,
		"fleetID":   comp.FleetID,
	}).Debug("fleet component synced to vehicle entity")
}

// ClearVehicleFleetComponent removes the GuildVehicleFleetComponent from the entity
// and clears VehicleComponent.FleetID. Implements guild_vehicle.VehicleSyncer.
func (s *GuildVehicleSystem) ClearVehicleFleetComponent(vehicleID uint64) {
	entity, ok := s.world.GetEntity(vehicleID)
	if !ok {
		return
	}
	entity.RemoveComponent("guild_vehicle_fleet")
	if raw, exists := entity.GetComponent("vehicle"); exists {
		if v, ok := raw.(*VehicleComponent); ok {
			v.FleetID = ""
		}
	}
	s.logger.WithField("vehicleID", vehicleID).Debug("fleet component cleared from vehicle entity")
}

// applyFormationBonuses applies fleet formation bonuses to vehicle combat stats.
func (s *GuildVehicleSystem) applyFormationBonuses(entity *Entity, fleetComp *guild_vehicle.GuildVehicleFleetComponent) {
	bonuses := s.manager.GetFleetBonuses(fleetComp.GuildID, fleetComp.FleetID)

	// Apply damage multiplier to vehicle combat component
	if entity.HasComponent("vehicle_combat") {
		comp, ok := entity.GetComponent("vehicle_combat")
		if !ok {
			return
		}
		combatComp, ok := comp.(*VehicleCombatComponent)
		if !ok {
			return
		}

		// Apply formation bonus and siege engine multiplier to ramming damage
		baseRammingDamage := combatComp.RammingDamage
		siegeMultiplier := fleetComp.SiegeType.GetSiegeDamageMultiplier()
		combatComp.RammingDamage = baseRammingDamage * bonuses.DamageMultiplier * siegeMultiplier

		// Apply bonuses to weapon damage if mounted
		if combatComp.WeaponMounted {
			baseWeaponDamage := combatComp.WeaponDamage
			combatComp.WeaponDamage = baseWeaponDamage * bonuses.DamageMultiplier * siegeMultiplier
		}

		// Apply defense bonus to armor rating
		baseArmor := combatComp.ArmorRating
		combatComp.ArmorRating = baseArmor * bonuses.DefenseMultiplier
	}

	s.logger.WithFields(logrus.Fields{
		"entityID":   entity.ID,
		"guild_id":   fleetComp.GuildID,
		"fleet_id":   fleetComp.FleetID,
		"formation":  bonuses.Formation.String(),
		"dmg_mult":   bonuses.DamageMultiplier,
		"def_mult":   bonuses.DefenseMultiplier,
		"siege_type": fleetComp.SiegeType.String(),
	}).Debug("Applied formation bonuses")
}

// AddVehicleToFleet adds a vehicle entity to a guild fleet.
func (s *GuildVehicleSystem) AddVehicleToFleet(guildID string, vehicleID uint64, fleetID string) error {
	if err := s.manager.AddVehicle(guildID, vehicleID, fleetID); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"guild_id":   guildID,
			"vehicle_id": vehicleID,
			"fleet_id":   fleetID,
		}).Error("Failed to add vehicle to fleet")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"vehicle_id": vehicleID,
		"fleet_id":   fleetID,
	}).Info("Added vehicle to fleet")

	return nil
}

// SetFormation sets the formation for a guild fleet.
func (s *GuildVehicleSystem) SetFormation(guildID, fleetID string, formation guild_vehicle.FormationType) error {
	if err := s.manager.SetFormation(guildID, fleetID, formation); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"guild_id":  guildID,
			"fleet_id":  fleetID,
			"formation": formation.String(),
		}).Error("Failed to set formation")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"guild_id":  guildID,
		"fleet_id":  fleetID,
		"formation": formation.String(),
	}).Info("Set fleet formation")

	return nil
}

// GrantAccess grants a player access to a guild vehicle.
func (s *GuildVehicleSystem) GrantAccess(guildID string, vehicleID uint64, playerID string) error {
	if err := s.manager.GrantAccess(guildID, vehicleID, playerID); err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"guild_id":   guildID,
			"vehicle_id": vehicleID,
			"player_id":  playerID,
		}).Error("Failed to grant access")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"guild_id":   guildID,
		"vehicle_id": vehicleID,
		"player_id":  playerID,
	}).Info("Granted vehicle access")

	return nil
}

// CheckAccess checks if a player has access to a guild vehicle.
func (s *GuildVehicleSystem) CheckAccess(guildID string, vehicleID uint64, playerID string) bool {
	return s.manager.CheckAccess(guildID, vehicleID, playerID)
}

// GetFleetManager returns the underlying fleet manager for advanced operations.
func (s *GuildVehicleSystem) GetFleetManager() *guild_vehicle.FleetManager {
	return s.manager
}
