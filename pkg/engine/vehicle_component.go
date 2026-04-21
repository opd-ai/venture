// Package engine provides vehicle component functionality.
// This file implements VehicleComponent and MountComponent which manage
// vehicle physics, stats, and rider-vehicle relationships.
//
// Phase 21.1: Vehicle Foundation
package engine

import "github.com/opd-ai/venture/pkg/procgen/terrain"

// VehicleType identifies different categories of vehicles.
type VehicleType int

const (
	// VehicleMount represents a living creature mount (horse, dragon, etc.)
	VehicleMount VehicleType = iota
	// VehicleCart represents a wheeled cart or wagon
	VehicleCart
	// VehicleBoat represents a water-based vehicle
	VehicleBoat
	// VehicleGlider represents an air vehicle (limited duration)
	VehicleGlider
	// VehicleMech represents a mechanical vehicle or robot suit
	VehicleMech
)

// String returns the string representation of a vehicle type.
func (v VehicleType) String() string {
	switch v {
	case VehicleMount:
		return "Mount"
	case VehicleCart:
		return "Cart"
	case VehicleBoat:
		return "Boat"
	case VehicleGlider:
		return "Glider"
	case VehicleMech:
		return "Mech"
	default:
		return "Unknown"
	}
}

// VehicleComponent stores vehicle stats and state.
// This component enables entity movement via mounted vehicles with
// physics (momentum, turning), durability, and fuel consumption.
type VehicleComponent struct {
	// VehicleType identifies the vehicle category (mount, cart, boat, glider, mech)
	VehicleType VehicleType

	// Speed is the current movement speed (pixels per second)
	Speed float64

	// MaxSpeed is the maximum movement speed achievable
	MaxSpeed float64

	// Acceleration is the rate of speed increase (pixels per second squared)
	Acceleration float64

	// Handling is the turn rate in radians per second
	// Higher values = tighter turning radius
	Handling float64

	// Durability is the current health/integrity of the vehicle
	Durability float64

	// MaxDurability is the maximum durability when undamaged
	MaxDurability float64

	// FuelType identifies the resource consumed (Stamina, Oil, Magic, Energy)
	FuelType string

	// FuelAmount is the current fuel available
	FuelAmount float64

	// FuelCapacity is the maximum fuel storage
	FuelCapacity float64

	// TerrainTypes lists terrain types this vehicle can traverse
	// Uses terrain.TileType constants from pkg/procgen/terrain
	TerrainTypes []terrain.TileType

	// Capacity is the number of passenger/cargo slots
	Capacity int

	// CurrentPassengers tracks the number of mounted riders
	CurrentPassengers int

	// FleetID is the guild fleet this vehicle belongs to. Empty when not in a fleet.
	// Set by GuildVehicleSystem.SyncVehicleFleetComponent when joining a fleet.
	FleetID string
}

// Type returns the component type identifier.
func (v *VehicleComponent) Type() string {
	return "vehicle"
}

// NewVehicleComponent creates a vehicle component with default values.
func NewVehicleComponent(vehicleType VehicleType) *VehicleComponent {
	// Default stats based on vehicle type
	var maxSpeed, acceleration, handling, maxDurability, fuelCapacity float64
	var capacity int
	var fuelType string
	var terrainTypes []terrain.TileType

	switch vehicleType {
	case VehicleMount:
		// Fast, agile, moderate durability
		maxSpeed = 200.0
		acceleration = 100.0
		handling = 3.0
		maxDurability = 100.0
		fuelType = "Stamina"
		fuelCapacity = 100.0
		capacity = 1
		// Can traverse most ground terrain
		terrainTypes = []terrain.TileType{
			terrain.TileFloor, terrain.TileCorridor, terrain.TileBridge,
			terrain.TileWaterShallow, terrain.TileRamp,
		}

	case VehicleCart:
		// Slow, poor handling, high capacity
		maxSpeed = 120.0
		acceleration = 50.0
		handling = 1.5
		maxDurability = 150.0
		fuelType = "Stamina"
		fuelCapacity = 200.0
		capacity = 4
		// Limited terrain - roads and flat ground
		terrainTypes = []terrain.TileType{
			terrain.TileFloor, terrain.TileCorridor, terrain.TileBridge,
		}

	case VehicleBoat:
		// Water-only, moderate speed
		maxSpeed = 150.0
		acceleration = 60.0
		handling = 2.0
		maxDurability = 120.0
		fuelType = "Energy"
		fuelCapacity = 150.0
		capacity = 2
		// Water terrain only
		terrainTypes = []terrain.TileType{
			terrain.TileWaterShallow, terrain.TileWaterDeep,
		}

	case VehicleGlider:
		// Fast, limited fuel (altitude), fragile
		maxSpeed = 250.0
		acceleration = 80.0
		handling = 2.5
		maxDurability = 60.0
		fuelType = "Energy"
		fuelCapacity = 60.0 // Limited flight time
		capacity = 1
		// Can fly over most terrain
		terrainTypes = []terrain.TileType{
			terrain.TileFloor, terrain.TileCorridor, terrain.TileDoor,
			terrain.TileWaterShallow, terrain.TileWaterDeep, terrain.TileTree,
			terrain.TileBridge, terrain.TilePlatform, terrain.TileRamp,
		}

	case VehicleMech:
		// Balanced stats, can traverse difficult terrain
		maxSpeed = 180.0
		acceleration = 70.0
		handling = 2.2
		maxDurability = 200.0
		fuelType = "Energy"
		fuelCapacity = 120.0
		capacity = 1
		// Can traverse all terrain including difficult areas
		terrainTypes = []terrain.TileType{
			terrain.TileFloor, terrain.TileCorridor, terrain.TileDoor,
			terrain.TileWaterShallow, terrain.TileWaterDeep, terrain.TileTree,
			terrain.TileBridge, terrain.TileStructure, terrain.TilePlatform,
			terrain.TileRamp, terrain.TileRampUp, terrain.TileRampDown,
		}
	}

	return &VehicleComponent{
		VehicleType:       vehicleType,
		Speed:             0.0, // Start stationary
		MaxSpeed:          maxSpeed,
		Acceleration:      acceleration,
		Handling:          handling,
		Durability:        maxDurability,
		MaxDurability:     maxDurability,
		FuelType:          fuelType,
		FuelAmount:        fuelCapacity,
		FuelCapacity:      fuelCapacity,
		TerrainTypes:      terrainTypes,
		Capacity:          capacity,
		CurrentPassengers: 0,
	}
}

// GetSpeed returns the current speed.
func (v *VehicleComponent) GetSpeed() float64 {
	return v.Speed
}

// GetMaxSpeed returns the maximum speed.
func (v *VehicleComponent) GetMaxSpeed() float64 {
	return v.MaxSpeed
}

// GetAcceleration returns the acceleration rate.
func (v *VehicleComponent) GetAcceleration() float64 {
	return v.Acceleration
}

// GetHandling returns the handling/turn rate.
func (v *VehicleComponent) GetHandling() float64 {
	return v.Handling
}

// CanTraverse checks if the vehicle can move on a specific terrain type.
func (v *VehicleComponent) CanTraverse(terrainType int) bool {
	tileType := terrain.TileType(terrainType)
	for _, allowed := range v.TerrainTypes {
		if allowed == tileType {
			return true
		}
	}
	return false
}

// GetFuelCost returns the fuel consumption per tile.
func (v *VehicleComponent) GetFuelCost() float64 {
	// Base fuel cost scales with speed
	// Faster vehicles consume more fuel
	return v.Speed / 100.0
}

// ConsumeFuel reduces fuel by the specified amount.
// Returns false if insufficient fuel available.
// Consumes remaining fuel even if less than requested amount.
func (v *VehicleComponent) ConsumeFuel(amount float64) bool {
	if v.FuelAmount <= 0 {
		return false
	}

	hadEnough := v.FuelAmount >= amount
	v.FuelAmount -= amount
	if v.FuelAmount < 0 {
		v.FuelAmount = 0
	}
	return hadEnough
}

// RefillFuel adds fuel up to capacity.
func (v *VehicleComponent) RefillFuel(amount float64) {
	v.FuelAmount += amount
	if v.FuelAmount > v.FuelCapacity {
		v.FuelAmount = v.FuelCapacity
	}
}

// TakeDamage reduces durability by the specified amount.
// Returns true if vehicle is destroyed (durability <= 0).
func (v *VehicleComponent) TakeDamage(damage float64) bool {
	v.Durability -= damage
	if v.Durability < 0 {
		v.Durability = 0
	}
	return v.Durability <= 0
}

// Repair increases durability up to maximum.
func (v *VehicleComponent) Repair(amount float64) {
	v.Durability += amount
	if v.Durability > v.MaxDurability {
		v.Durability = v.MaxDurability
	}
}

// CanAddPassenger checks if the vehicle has capacity for another rider.
func (v *VehicleComponent) CanAddPassenger() bool {
	return v.CurrentPassengers < v.Capacity
}

// AddPassenger increments the passenger count.
// Returns false if vehicle is at capacity.
func (v *VehicleComponent) AddPassenger() bool {
	if !v.CanAddPassenger() {
		return false
	}
	v.CurrentPassengers++
	return true
}

// RemovePassenger decrements the passenger count.
func (v *VehicleComponent) RemovePassenger() {
	if v.CurrentPassengers > 0 {
		v.CurrentPassengers--
	}
}

// IsFuelDepleted checks if fuel is exhausted.
func (v *VehicleComponent) IsFuelDepleted() bool {
	return v.FuelAmount <= 0
}

// IsDestroyed checks if durability is depleted.
func (v *VehicleComponent) IsDestroyed() bool {
	return v.Durability <= 0
}

// MountComponent stores the relationship between a rider and vehicle.
// This component can be attached to either the rider or vehicle entity.
type MountComponent struct {
	// MountedEntityID is the ID of the vehicle entity being ridden (when on rider)
	MountedEntityID uint64

	// RiderID is the ID of the rider entity (when on vehicle)
	RiderID uint64

	// IsMounted indicates if mounting is currently active
	IsMounted bool

	// MountOffset is the visual offset from vehicle center
	// Used for rendering rider at correct position relative to vehicle
	MountOffset Vector2D
}

// Type returns the component type identifier.
func (m *MountComponent) Type() string {
	return "mount"
}

// NewMountComponent creates a mount component linking rider to vehicle.
func NewMountComponent(vehicleID uint64, offsetX, offsetY float64) *MountComponent {
	return &MountComponent{
		MountedEntityID: vehicleID,
		MountOffset: Vector2D{
			X: offsetX,
			Y: offsetY,
		},
	}
}

// Serialize converts MountComponent to bytes for network transmission.
// Format: vehicleID(8) + offsetX(8) + offsetY(8) = 24 bytes
func (m *MountComponent) Serialize() []byte {
	buf := make([]byte, 24)

	// MountedEntityID (8 bytes)
	writeUint64(buf[0:], m.MountedEntityID)

	// MountOffset.X (8 bytes)
	writeFloat64(buf[8:], m.MountOffset.X)

	// MountOffset.Y (8 bytes)
	writeFloat64(buf[16:], m.MountOffset.Y)

	return buf
}

// Deserialize restores MountComponent from bytes.
func (m *MountComponent) Deserialize(data []byte) error {
	if len(data) < 24 {
		return ErrInvalidComponentData
	}

	// MountedEntityID
	m.MountedEntityID = readUint64(data[0:])

	// MountOffset.X
	m.MountOffset.X = readFloat64(data[8:])

	// MountOffset.Y
	m.MountOffset.Y = readFloat64(data[16:])

	return nil
}

// Serialize converts VehicleComponent to bytes for network transmission.
// Binary format: type(1) + speed(8) + maxSpeed(8) + accel(8) + handling(8) +
// durability(8) + maxDur(8) + fuelAmt(8) + fuelCap(8) + capacity(4) + passengers(4) + terrainCount(4) = 77 bytes
// (excluding variable-length terrainTypes)
func (v *VehicleComponent) Serialize() []byte {
	// Allocate buffer: 1 + 8*8 + 4*3 = 77 bytes base + terrain types (4 bytes each)
	buf := make([]byte, 77+len(v.TerrainTypes)*4)

	// Vehicle type (1 byte)
	buf[0] = byte(v.VehicleType)

	// Float64 values (8 bytes each)
	offset := 1
	writeFloat64(buf[offset:], v.Speed)
	offset += 8
	writeFloat64(buf[offset:], v.MaxSpeed)
	offset += 8
	writeFloat64(buf[offset:], v.Acceleration)
	offset += 8
	writeFloat64(buf[offset:], v.Handling)
	offset += 8
	writeFloat64(buf[offset:], v.Durability)
	offset += 8
	writeFloat64(buf[offset:], v.MaxDurability)
	offset += 8
	writeFloat64(buf[offset:], v.FuelAmount)
	offset += 8
	writeFloat64(buf[offset:], v.FuelCapacity)
	offset += 8

	// Int32 values (4 bytes each)
	writeInt32(buf[offset:], int32(v.Capacity))
	offset += 4
	writeInt32(buf[offset:], int32(v.CurrentPassengers))
	offset += 4

	// Terrain types count (4 bytes) + each type (4 bytes)
	writeInt32(buf[offset:], int32(len(v.TerrainTypes)))
	offset += 4
	for _, tt := range v.TerrainTypes {
		writeInt32(buf[offset:], int32(tt))
		offset += 4
	}

	return buf
}

// Deserialize restores VehicleComponent from bytes.
func (v *VehicleComponent) Deserialize(data []byte) error {
	if len(data) < 73 {
		return ErrInvalidComponentData
	}

	// Vehicle type
	v.VehicleType = VehicleType(data[0])

	// Float64 values
	offset := 1
	v.Speed = readFloat64(data[offset:])
	offset += 8
	v.MaxSpeed = readFloat64(data[offset:])
	offset += 8
	v.Acceleration = readFloat64(data[offset:])
	offset += 8
	v.Handling = readFloat64(data[offset:])
	offset += 8
	v.Durability = readFloat64(data[offset:])
	offset += 8
	v.MaxDurability = readFloat64(data[offset:])
	offset += 8
	v.FuelAmount = readFloat64(data[offset:])
	offset += 8
	v.FuelCapacity = readFloat64(data[offset:])
	offset += 8

	// Int32 values
	v.Capacity = int(readInt32(data[offset:]))
	offset += 4
	v.CurrentPassengers = int(readInt32(data[offset:]))
	offset += 4

	// Terrain types
	terrainCount := int(readInt32(data[offset:]))
	offset += 4
	v.TerrainTypes = make([]terrain.TileType, terrainCount)
	for i := 0; i < terrainCount; i++ {
		v.TerrainTypes[i] = terrain.TileType(readInt32(data[offset:]))
		offset += 4
	}

	return nil
}
