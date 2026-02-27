// Package fluids - types.go
// This file contains all type definitions, constants, and helper functions for fluid simulation.

package fluids

import (
	"encoding/binary"
	"errors"
	"image/color"
	"math"
	"sync"
)

// ErrInvalidData is returned when deserialization encounters invalid data
var ErrInvalidData = errors.New("invalid component data")

// FluidType represents the type of fluid
type FluidType int

const (
	FluidWater FluidType = iota
	FluidLava
	FluidOil
	FluidAcid
	FluidPoison
)

// String returns the human-readable name of the fluid type
func (f FluidType) String() string {
	switch f {
	case FluidWater:
		return "Water"
	case FluidLava:
		return "Lava"
	case FluidOil:
		return "Oil"
	case FluidAcid:
		return "Acid"
	case FluidPoison:
		return "Poison"
	default:
		return "Unknown"
	}
}

// FluidProperties defines the physical properties of a fluid type
type FluidProperties struct {
	Viscosity    float64    // Resistance to flow (0.0-1.0)
	Density      float64    // Mass per unit volume (kg/m³)
	FlowRate     float64    // Speed of flow propagation (0.0-1.0)
	Damage       float64    // Damage per second when submerged
	Color        color.RGBA // Visual color of the fluid
	Transparency float64    // Visual transparency (0.0=opaque, 1.0=transparent)
}

// Cell represents a single grid cell in the fluid simulation
type Cell struct {
	Pressure  float64   // Pressure value (0.0-1.0)
	VelocityX float64   // X-axis velocity
	VelocityY float64   // Y-axis velocity
	Type      FluidType // Type of fluid in this cell
	Amount    float64   // Amount of fluid (0.0=empty, 1.0=full)
}

// Grid represents the fluid simulation grid
type Grid struct {
	Width  int      // Grid width in cells
	Height int      // Grid height in cells
	Cells  [][]Cell // 2D array of cells
	mu     sync.RWMutex
}

// BuoyancyComponent defines buoyancy properties for an entity
type BuoyancyComponent struct {
	Mass         float64 // Entity mass (kg)
	Volume       float64 // Entity volume (m³)
	Density      float64 // Entity density (kg/m³) = Mass/Volume
	Buoyant      bool    // Whether entity floats
	Submerged    float64 // Percentage submerged (0.0-1.0)
	BuoyantForce float64 // Upward force from buoyancy (N)
}

// Type returns the component type identifier
func (b BuoyancyComponent) Type() string {
	return "buoyancy"
}

// Serialize encodes the component to bytes for persistence
func (b *BuoyancyComponent) Serialize() ([]byte, error) {
	// 6 float64s (48 bytes) + 1 bool (1 byte) = 49 bytes
	buf := make([]byte, 49)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(b.Mass))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(b.Volume))
	binary.LittleEndian.PutUint64(buf[16:24], math.Float64bits(b.Density))
	binary.LittleEndian.PutUint64(buf[24:32], math.Float64bits(b.Submerged))
	binary.LittleEndian.PutUint64(buf[32:40], math.Float64bits(b.BuoyantForce))
	if b.Buoyant {
		buf[48] = 1
	} else {
		buf[48] = 0
	}
	return buf, nil
}

// Deserialize decodes the component from bytes
func (b *BuoyancyComponent) Deserialize(data []byte) error {
	if len(data) < 49 {
		return ErrInvalidData
	}
	b.Mass = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	b.Volume = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	b.Density = math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	b.Submerged = math.Float64frombits(binary.LittleEndian.Uint64(data[24:32]))
	b.BuoyantForce = math.Float64frombits(binary.LittleEndian.Uint64(data[32:40]))
	b.Buoyant = data[48] != 0
	return nil
}

// SwimmingComponent defines swimming mechanics for an entity
type SwimmingComponent struct {
	IsSwimming     bool    // Currently in water and swimming
	Stamina        float64 // Swimming stamina (0.0-100.0)
	MaxStamina     float64 // Maximum stamina
	StaminaDrain   float64 // Stamina drain per second
	StaminaRegen   float64 // Stamina regen per second (on land)
	SwimSpeed      float64 // Swimming speed multiplier (0.0-1.0)
	TreadingWater  bool    // Staying afloat without moving
	Drowning       bool    // Out of stamina, taking damage
	DrowningDamage float64 // Damage per second when drowning
}

// Type returns the component type identifier
func (s SwimmingComponent) Type() string {
	return "swimming"
}

// Serialize encodes the component to bytes for persistence
func (s *SwimmingComponent) Serialize() ([]byte, error) {
	// 6 float64s (48 bytes) + 3 bools (3 bytes) = 51 bytes
	buf := make([]byte, 51)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(s.Stamina))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(s.MaxStamina))
	binary.LittleEndian.PutUint64(buf[16:24], math.Float64bits(s.StaminaDrain))
	binary.LittleEndian.PutUint64(buf[24:32], math.Float64bits(s.StaminaRegen))
	binary.LittleEndian.PutUint64(buf[32:40], math.Float64bits(s.SwimSpeed))
	binary.LittleEndian.PutUint64(buf[40:48], math.Float64bits(s.DrowningDamage))
	if s.IsSwimming {
		buf[48] = 1
	}
	if s.TreadingWater {
		buf[49] = 1
	}
	if s.Drowning {
		buf[50] = 1
	}
	return buf, nil
}

// Deserialize decodes the component from bytes
func (s *SwimmingComponent) Deserialize(data []byte) error {
	if len(data) < 51 {
		return ErrInvalidData
	}
	s.Stamina = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	s.MaxStamina = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	s.StaminaDrain = math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	s.StaminaRegen = math.Float64frombits(binary.LittleEndian.Uint64(data[24:32]))
	s.SwimSpeed = math.Float64frombits(binary.LittleEndian.Uint64(data[32:40]))
	s.DrowningDamage = math.Float64frombits(binary.LittleEndian.Uint64(data[40:48]))
	s.IsSwimming = data[48] != 0
	s.TreadingWater = data[49] != 0
	s.Drowning = data[50] != 0
	return nil
}

// FloodingComponent tracks flooding state for an enclosed area
type FloodingComponent struct {
	AreaID        string        // Identifier for the area being flooded
	FloodLevel    float64       // Current flood level (0.0-1.0)
	FloodRate     float64       // Rate of flooding (units per second)
	MaxFloodLevel float64       // Maximum flood level before area is full
	Sources       []FloodSource // Water entry points
}

// Type returns the component type identifier
func (f FloodingComponent) Type() string {
	return "flooding"
}

// Serialize encodes the component to bytes for persistence
func (f *FloodingComponent) Serialize() ([]byte, error) {
	// Calculate size: AreaID length prefix (4) + AreaID + 3 float64s (24) + source count (4) + sources
	sourceSize := len(f.Sources) * (4 + 4 + 8) // x(4) + y(4) + flowRate(8) per source
	areaIDLen := len(f.AreaID)
	totalSize := 4 + areaIDLen + 24 + 4 + sourceSize
	buf := make([]byte, totalSize)

	offset := 0
	// Write AreaID length and content
	binary.LittleEndian.PutUint32(buf[offset:], uint32(areaIDLen))
	offset += 4
	copy(buf[offset:], f.AreaID)
	offset += areaIDLen

	// Write float64 fields
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(f.FloodLevel))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(f.FloodRate))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(f.MaxFloodLevel))
	offset += 8

	// Write sources
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(f.Sources)))
	offset += 4
	for _, source := range f.Sources {
		binary.LittleEndian.PutUint32(buf[offset:], uint32(source.X))
		offset += 4
		binary.LittleEndian.PutUint32(buf[offset:], uint32(source.Y))
		offset += 4
		binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(source.FlowRate))
		offset += 8
	}
	return buf, nil
}

// Deserialize decodes the component from bytes
func (f *FloodingComponent) Deserialize(data []byte) error {
	if len(data) < 32 { // minimum: 4 (areaID len) + 0 (areaID) + 24 (floats) + 4 (source count)
		return ErrInvalidData
	}
	offset := 0

	// Read AreaID
	areaIDLen := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+areaIDLen+28 {
		return ErrInvalidData
	}
	f.AreaID = string(data[offset : offset+areaIDLen])
	offset += areaIDLen

	// Read float64 fields
	f.FloodLevel = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	f.FloodRate = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8
	f.MaxFloodLevel = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8

	// Read sources
	sourceCount := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+sourceCount*16 {
		return ErrInvalidData
	}
	f.Sources = make([]FloodSource, sourceCount)
	for i := 0; i < sourceCount; i++ {
		f.Sources[i].X = int(binary.LittleEndian.Uint32(data[offset:]))
		offset += 4
		f.Sources[i].Y = int(binary.LittleEndian.Uint32(data[offset:]))
		offset += 4
		f.Sources[i].FlowRate = math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
		offset += 8
	}
	return nil
}

// FloodSource represents a water entry point
type FloodSource struct {
	X        int     // Grid X position
	Y        int     // Grid Y position
	FlowRate float64 // Amount of water per second
}

// SimulationConfig defines configuration for fluid simulation
type SimulationConfig struct {
	GridWidth       int     // Width of simulation grid
	GridHeight      int     // Height of simulation grid
	CellSize        float64 // Size of each cell in world units
	UpdateRate      float64 // Updates per second (default: 30 FPS)
	Gravity         float64 // Gravity constant (m/s²)
	PressureFactor  float64 // Pressure calculation factor
	ViscosityFactor float64 // Viscosity calculation factor
	MaxIterations   int     // Max iterations per update
	Convergence     float64 // Convergence threshold
}

// DefaultSimulationConfig returns default configuration values
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		GridWidth:       100,
		GridHeight:      100,
		CellSize:        1.0,
		UpdateRate:      30.0,
		Gravity:         9.81,
		PressureFactor:  1.0,
		ViscosityFactor: 0.95,
		MaxIterations:   10,
		Convergence:     0.001,
	}
}

// GetFluidProperties returns the properties for a given fluid type
func GetFluidProperties(fluidType FluidType) FluidProperties {
	switch fluidType {
	case FluidWater:
		return FluidProperties{
			Viscosity:    0.1,
			Density:      1000.0,
			FlowRate:     0.8,
			Damage:       0.0,
			Color:        color.RGBA{R: 30, G: 144, B: 255, A: 180},
			Transparency: 0.7,
		}
	case FluidLava:
		return FluidProperties{
			Viscosity:    0.6,
			Density:      3000.0,
			FlowRate:     0.3,
			Damage:       50.0,
			Color:        color.RGBA{R: 255, G: 69, B: 0, A: 220},
			Transparency: 0.2,
		}
	case FluidOil:
		return FluidProperties{
			Viscosity:    0.3,
			Density:      900.0,
			FlowRate:     0.6,
			Damage:       0.0,
			Color:        color.RGBA{R: 50, G: 50, B: 50, A: 150},
			Transparency: 0.5,
		}
	case FluidAcid:
		return FluidProperties{
			Viscosity:    0.15,
			Density:      1200.0,
			FlowRate:     0.7,
			Damage:       25.0,
			Color:        color.RGBA{R: 0, G: 255, B: 0, A: 200},
			Transparency: 0.6,
		}
	case FluidPoison:
		return FluidProperties{
			Viscosity:    0.12,
			Density:      1050.0,
			FlowRate:     0.75,
			Damage:       15.0,
			Color:        color.RGBA{R: 128, G: 0, B: 128, A: 190},
			Transparency: 0.65,
		}
	default:
		return FluidProperties{
			Viscosity:    0.1,
			Density:      1000.0,
			FlowRate:     0.5,
			Damage:       0.0,
			Color:        color.RGBA{R: 255, G: 255, B: 255, A: 255},
			Transparency: 1.0,
		}
	}
}

// RegisterComponentFactories registers factory functions for all fluid components with the ECS.
// This is a convenience function to simplify component registration for consuming code.
//
// Usage:
//
//	import "github.com/opd-ai/venture/pkg/engine/physics/fluids"
//	fluids.RegisterComponentFactories()
//
// After calling this function, components can be created via:
//   - entity.AddComponent(&fluids.BuoyancyComponent{...})
//   - entity.AddComponent(&fluids.SwimmingComponent{...})
//   - entity.AddComponent(&fluids.FloodingComponent{...})
//
// This function is idempotent and safe to call multiple times.
// It does not require a World instance - component registration is global.
func RegisterComponentFactories() {
	// Note: In the current ECS architecture, components are registered
	// simply by being added to entities via AddComponent().
	// This function is provided as a documentation aid and future-proofing
	// in case explicit component type registration becomes necessary.
	//
	// The three fluid components that should be registered are:
	//   - BuoyancyComponent (type: "buoyancy")
	//   - SwimmingComponent (type: "swimming")
	//   - FloodingComponent (type: "flooding")
}
