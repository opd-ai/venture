package guild_vehicle

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

// MembershipValidator validates guild membership for fleet access control.
// Implement this interface with pkg/network/federation/guild.Manager to enforce
// that only actual guild members can be granted access to fleet vehicles.
type MembershipValidator interface {
	// IsMember returns true if playerID is an active member of guildID.
	IsMember(guildID, playerID string) bool
}

// VehicleSyncer synchronizes GuildVehicleFleetComponent onto vehicle entities.
// Implement this interface in the ECS engine layer (pkg/engine) to propagate
// fleet membership data onto vehicle entities when they join or leave a fleet.
// The engine already imports guild_vehicle, so it can register itself as the syncer.
type VehicleSyncer interface {
	// SyncVehicleFleetComponent attaches or updates the GuildVehicleFleetComponent
	// on the entity identified by vehicleID.
	SyncVehicleFleetComponent(vehicleID uint64, comp *GuildVehicleFleetComponent)
	// ClearVehicleFleetComponent removes the GuildVehicleFleetComponent from the
	// entity identified by vehicleID.
	ClearVehicleFleetComponent(vehicleID uint64)
}

// StructureDamager applies siege engine damage to territory defensive structures.
// Implement this interface with pkg/world/territory.Manager to bridge siege
// vehicle payloads to defensive structure hit-points.
type StructureDamager interface {
	// DamageStructure applies damage to structureID in the given territory.
	DamageStructure(territoryID, structureID string, damage float64) error
}

// FormationOffset describes the relative position and heading adjustment for a
// single vehicle slot in a fleet formation. Offsets are expressed in the fleet
// commander's local coordinate frame (Y = forward, X = right, angle in radians).
type FormationOffset struct {
	// SlotIndex is the 0-based position within the formation (0 = leader).
	SlotIndex int
	// OffsetX is the right/left displacement from the leader position (world units).
	OffsetX float64
	// OffsetY is the forward/backward displacement from the leader position (world units).
	// Negative values place the vehicle behind the leader.
	OffsetY float64
	// AngleAdjust is the rotation delta in radians relative to the leader's heading.
	AngleAdjust float64
}

// formationSpacing is the default distance between vehicles in a formation (world units).
const formationSpacing = 50.0

// SetMembershipValidator injects a MembershipValidator into the fleet manager.
// When set, GrantAccess will reject any playerID that is not a member of the guild.
// Pass nil to disable membership validation (default; useful for testing).
func (m *FleetManager) SetMembershipValidator(v MembershipValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.membershipValidator = v
	log.WithFields(log.Fields{
		"system_name": "fleet_manager",
		"validator":   v != nil,
	}).Debug("fleet manager membership validator updated")
}

// SetVehicleSyncer injects a VehicleSyncer into the fleet manager.
// When set, AddVehicle/AddVehicleWithType and RemoveVehicle will notify the syncer
// so that ECS vehicle entities receive an up-to-date GuildVehicleFleetComponent.
// Pass nil to disable ECS synchronization (default).
func (m *FleetManager) SetVehicleSyncer(s VehicleSyncer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.vehicleSyncer = s
	log.WithFields(log.Fields{
		"system_name": "fleet_manager",
		"syncer":      s != nil,
	}).Debug("fleet manager vehicle syncer updated")
}

// SetStructureDamager injects a StructureDamager into the fleet manager.
// When set, ApplySiegeDamage will delegate structure damage to the territory system.
// Pass nil to disable territory integration (default).
func (m *FleetManager) SetStructureDamager(d StructureDamager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.structureDamager = d
	log.WithFields(log.Fields{
		"system_name": "fleet_manager",
		"damager":     d != nil,
	}).Debug("fleet manager structure damager updated")
}

// GetFormationOffsets returns the ideal per-slot position offsets for the named fleet
// based on its current formation type. Offsets are in the fleet commander's local
// coordinate frame and can be fed into pkg/engine/physics/vehicle per-frame target
// positions. Slot 0 (the leader) always has zero offset.
// Returns nil if the fleet does not exist.
func (m *FleetManager) GetFormationOffsets(guildID, fleetID string) []FormationOffset {
	m.mu.RLock()
	key := m.getFleetKey(guildID, fleetID)
	fleet, exists := m.fleets[key]
	if !exists {
		m.mu.RUnlock()
		return nil
	}
	formation := fleet.Formation
	count := len(fleet.Vehicles)
	m.mu.RUnlock()

	return calculateFormationOffsets(formation, count)
}

// calculateFormationOffsets computes position offsets for each vehicle slot.
func calculateFormationOffsets(f FormationType, count int) []FormationOffset {
	offsets := make([]FormationOffset, count)
	for i := range offsets {
		offsets[i].SlotIndex = i
	}
	if count <= 1 {
		return offsets
	}

	for i := 1; i < count; i++ {
		offsets[i] = slotOffset(f, i, count)
	}
	return offsets
}

// slotOffset returns the formation offset for a single non-leader slot.
func slotOffset(f FormationType, slot, total int) FormationOffset {
	switch f {
	case FormationLine:
		return lineOffset(slot)
	case FormationWedge:
		return wedgeOffset(slot)
	case FormationColumn:
		return FormationOffset{SlotIndex: slot, OffsetY: -float64(slot) * formationSpacing}
	case FormationCircle:
		return circleOffset(slot, total)
	default:
		return FormationOffset{SlotIndex: slot}
	}
}

func lineOffset(slot int) FormationOffset {
	side := float64((slot+1)/2) * formationSpacing
	if slot%2 == 1 {
		return FormationOffset{SlotIndex: slot, OffsetX: side}
	}
	return FormationOffset{SlotIndex: slot, OffsetX: -side}
}

func wedgeOffset(slot int) FormationOffset {
	row := float64((slot + 1) / 2)
	side := row * formationSpacing
	if slot%2 == 1 {
		return FormationOffset{SlotIndex: slot, OffsetX: side, OffsetY: -row * formationSpacing}
	}
	return FormationOffset{SlotIndex: slot, OffsetX: -side, OffsetY: -row * formationSpacing}
}

func circleOffset(slot, total int) FormationOffset {
	angle := (2 * math.Pi * float64(slot)) / float64(total)
	return FormationOffset{
		SlotIndex:   slot,
		OffsetX:     math.Cos(angle) * formationSpacing,
		OffsetY:     math.Sin(angle) * formationSpacing,
		AngleAdjust: angle,
	}
}

// ApplySiegeDamage looks up the vehicle's siege engine type, multiplies weaponDamage
// by the siege multiplier, and delegates the result to the StructureDamager.
// This bridges fleet vehicle payloads to territory structure hit-points
// (pkg/world/territory.Manager satisfies StructureDamager).
// Returns an error if the vehicle is not found, no StructureDamager is configured,
// or the underlying damage call fails.
func (m *FleetManager) ApplySiegeDamage(vehicleID uint64, territoryID, structureID string, weaponDamage float64) error {
	m.mu.RLock()
	siegeType, found := m.findVehicleSiegeType(vehicleID)
	damager := m.structureDamager
	m.mu.RUnlock()

	if !found {
		return fmt.Errorf("vehicle %d not found in any fleet", vehicleID)
	}
	if damager == nil {
		return fmt.Errorf("no StructureDamager configured; call SetStructureDamager first")
	}

	totalDamage := weaponDamage * siegeType.GetSiegeDamageMultiplier()
	log.WithFields(log.Fields{
		"system_name":  "fleet_manager",
		"vehicle_id":   vehicleID,
		"siege_type":   siegeType.String(),
		"weapon_dmg":   weaponDamage,
		"total_dmg":    totalDamage,
		"territory_id": territoryID,
		"structure_id": structureID,
	}).Debug("applying siege vehicle damage to structure")
	return damager.DamageStructure(territoryID, structureID, totalDamage)
}

// findVehicleSiegeType scans all fleets for a vehicle with the given ID.
// Must be called with m.mu.RLock held.
func (m *FleetManager) findVehicleSiegeType(vehicleID uint64) (SiegeEngineType, bool) {
	for _, fleet := range m.fleets {
		if v, ok := fleet.Vehicles[vehicleID]; ok {
			return v.SiegeType, true
		}
	}
	return SiegeNone, false
}
