package guild_vehicle

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// FormationType represents fleet formation patterns
type FormationType int

const (
	// FormationNone indicates no formation (scattered vehicles)
	FormationNone FormationType = iota
	// FormationLine represents lateral line formation (5% damage bonus)
	FormationLine
	// FormationWedge represents frontal assault wedge (7% damage bonus)
	FormationWedge
	// FormationColumn represents single-file column (10% defense bonus)
	FormationColumn
	// FormationCircle represents defensive perimeter (8% defense bonus)
	FormationCircle
)

// String returns the string representation of formation type
func (f FormationType) String() string {
	switch f {
	case FormationNone:
		return "None"
	case FormationLine:
		return "Line"
	case FormationWedge:
		return "Wedge"
	case FormationColumn:
		return "Column"
	case FormationCircle:
		return "Circle"
	default:
		return "Unknown"
	}
}

// SiegeEngineType represents specialized siege vehicle categories
type SiegeEngineType int

const (
	// SiegeNone indicates regular vehicle (not a siege engine)
	SiegeNone SiegeEngineType = iota
	// SiegeBatteringRam for breaking through walls (3x damage vs walls)
	SiegeBatteringRam
	// SiegeCatapult for long-range structure bombardment (5x damage)
	SiegeCatapult
	// SiegeTower for scaling walls (2x damage, wall climbing)
	SiegeTower
	// SiegeBallistaBattery for destroying defensive towers (4x damage)
	SiegeBallistaBattery
)

// String returns the string representation of siege engine type
func (s SiegeEngineType) String() string {
	switch s {
	case SiegeNone:
		return "None"
	case SiegeBatteringRam:
		return "BatteringRam"
	case SiegeCatapult:
		return "Catapult"
	case SiegeTower:
		return "SiegeTower"
	case SiegeBallistaBattery:
		return "BallistaBattery"
	default:
		return "Unknown"
	}
}

// GetSiegeDamageMultiplier returns damage multiplier for siege engines
func (s SiegeEngineType) GetSiegeDamageMultiplier() float64 {
	switch s {
	case SiegeBatteringRam:
		return 3.0
	case SiegeCatapult:
		return 5.0
	case SiegeTower:
		return 2.0
	case SiegeBallistaBattery:
		return 4.0
	default:
		return 1.0
	}
}

// FleetBonus represents combat bonuses from formation
type FleetBonus struct {
	// DamageMultiplier affects outgoing damage (1.0 = no bonus)
	DamageMultiplier float64
	// DefenseMultiplier affects incoming damage reduction (1.0 = no bonus)
	DefenseMultiplier float64
	// Formation type providing the bonus
	Formation FormationType
}

// GetFormationBonus returns bonuses for a specific formation type
func GetFormationBonus(formation FormationType) FleetBonus {
	switch formation {
	case FormationLine:
		return FleetBonus{
			DamageMultiplier:  1.05, // 5% damage bonus
			DefenseMultiplier: 1.0,
			Formation:         FormationLine,
		}
	case FormationWedge:
		return FleetBonus{
			DamageMultiplier:  1.07, // 7% damage bonus
			DefenseMultiplier: 1.0,
			Formation:         FormationWedge,
		}
	case FormationColumn:
		return FleetBonus{
			DamageMultiplier:  1.0,
			DefenseMultiplier: 1.10, // 10% defense bonus
			Formation:         FormationColumn,
		}
	case FormationCircle:
		return FleetBonus{
			DamageMultiplier:  1.0,
			DefenseMultiplier: 1.08, // 8% defense bonus
			Formation:         FormationCircle,
		}
	default:
		return FleetBonus{
			DamageMultiplier:  1.0,
			DefenseMultiplier: 1.0,
			Formation:         FormationNone,
		}
	}
}

// GuildVehicle represents a vehicle owned by a guild
type GuildVehicle struct {
	// VehicleID is the entity ID of the vehicle
	VehicleID uint64
	// GuildID is the owning guild identifier
	GuildID string
	// FleetID identifies which fleet this vehicle belongs to
	FleetID string
	// SiegeType indicates if this is a siege engine
	SiegeType SiegeEngineType
	// SharedAccess maps player IDs to access permissions
	SharedAccess map[string]bool
	// MaintenanceCost is daily gold cost from guild treasury
	MaintenanceCost int
	// FormationSlot is the stable slot index assigned at add time; never changes after assignment.
	// Slot 0 is the fleet commander (leader); higher slots are followers in formation order.
	FormationSlot int
	// AddedAt tracks when vehicle was added to fleet
	AddedAt time.Time
	// LastMaintenance tracks last maintenance payment
	LastMaintenance time.Time
}

// Fleet represents a coordinated group of guild vehicles
type Fleet struct {
	// FleetID uniquely identifies this fleet within the guild
	FleetID string
	// GuildID is the owning guild identifier
	GuildID string
	// Vehicles maps vehicle entity IDs to GuildVehicle structs
	Vehicles map[uint64]*GuildVehicle
	// Formation is the current fleet formation
	Formation FormationType
	// CommanderID is the player ID of fleet commander (can issue formation commands)
	CommanderID string
	// nextSlot is the monotonically increasing slot counter for stable FormationSlot assignment.
	// It is never decremented on removal, so slots remain unique and stable across the vehicle lifecycle.
	nextSlot int
	// CreatedAt tracks fleet creation time
	CreatedAt time.Time
	// UpdatedAt tracks last fleet modification
	UpdatedAt time.Time
}

// GetVehicleCount returns number of vehicles in fleet
func (f *Fleet) GetVehicleCount() int {
	return len(f.Vehicles)
}

// GetTotalMaintenanceCost calculates sum of all vehicle maintenance costs
func (f *Fleet) GetTotalMaintenanceCost() int {
	total := 0
	for _, vehicle := range f.Vehicles {
		total += vehicle.MaintenanceCost
	}
	return total
}

// GetSiegeEngineCount returns number of siege engines in fleet
func (f *Fleet) GetSiegeEngineCount() int {
	count := 0
	for _, vehicle := range f.Vehicles {
		if vehicle.SiegeType != SiegeNone {
			count++
		}
	}
	return count
}

// HasAccess checks if a player has access to a specific vehicle
func (v *GuildVehicle) HasAccess(playerID string) bool {
	if v.SharedAccess == nil {
		return false
	}
	return v.SharedAccess[playerID]
}

// GuildVehicleFleetComponent is an ECS component for guild fleet integration
type GuildVehicleFleetComponent struct {
	// GuildID identifies the owning guild
	GuildID string
	// FleetID identifies which fleet this vehicle belongs to
	FleetID string
	// SiegeType indicates siege engine category
	SiegeType SiegeEngineType
	// FormationPosition is index in formation (0 = leader, 1+ = followers)
	FormationPosition int
	// LastFormationUpdate tracks when formation was last changed
	LastFormationUpdate time.Time
}

// Type returns the component type identifier
func (g *GuildVehicleFleetComponent) Type() string {
	return "guild_vehicle_fleet"
}

// Serialize encodes the component to JSON bytes for persistence.
// Returns the encoded bytes and any error encountered during marshaling.
func (g *GuildVehicleFleetComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type":     "guild_vehicle_fleet",
		"guild_id":           g.GuildID,
		"fleet_id":           g.FleetID,
		"formation_position": g.FormationPosition,
	}).Debug("Serializing guild vehicle fleet component")

	data, err := json.Marshal(g)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "guild_vehicle_fleet",
			"error":          err.Error(),
		}).Error("Failed to serialize guild vehicle fleet component")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "guild_vehicle_fleet",
		"bytes":          len(data),
	}).Debug("Guild vehicle fleet component serialized successfully")

	return data, nil
}

// Deserialize decodes the component from JSON bytes.
// Returns any error encountered during unmarshaling.
func (g *GuildVehicleFleetComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "guild_vehicle_fleet",
		"bytes":          len(data),
	}).Debug("Deserializing guild vehicle fleet component")

	if err := json.Unmarshal(data, g); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "guild_vehicle_fleet",
			"error":          err.Error(),
		}).Error("Failed to deserialize guild vehicle fleet component")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"component_type":     "guild_vehicle_fleet",
		"guild_id":           g.GuildID,
		"fleet_id":           g.FleetID,
		"formation_position": g.FormationPosition,
	}).Debug("Guild vehicle fleet component deserialized successfully")

	return nil
}
