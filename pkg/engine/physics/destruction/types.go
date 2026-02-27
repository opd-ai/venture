package destruction

import (
	"image/color"
	"math"

	log "github.com/sirupsen/logrus"
)

// IntegrityState represents the structural integrity state of a building
type IntegrityState int

const (
	IntegrityPristine  IntegrityState = iota // 100% structural integrity
	IntegrityDamaged                         // 50-99% structural integrity
	IntegrityCritical                        // 1-49% structural integrity
	IntegrityCollapsed                       // 0% structural integrity (destroyed)
)

// String returns the string representation of IntegrityState.
// Returns "Pristine", "Damaged", "Critical", "Collapsed", or "Unknown" for invalid states.
func (s IntegrityState) String() string {
	switch s {
	case IntegrityPristine:
		return "Pristine"
	case IntegrityDamaged:
		return "Damaged"
	case IntegrityCritical:
		return "Critical"
	case IntegrityCollapsed:
		return "Collapsed"
	default:
		return "Unknown"
	}
}

// StructuralIntegrity represents a building's structural health
type StructuralIntegrity struct {
	BuildingID     string
	TotalHealth    float64        // Total structural health points
	CurrentHealth  float64        // Current structural health (0.0-1.0 of total)
	Supports       []SupportPoint // Load-bearing structures
	DamagedAreas   []DamageArea   // Areas with structural damage
	State          IntegrityState // Current integrity state
	CollapseRisk   float64        // Risk of collapse (0.0-1.0)
	LastDamageTime float64        // Time of last damage (for propagation)
}

// SupportPoint represents a load-bearing structure (wall, column, beam)
type SupportPoint struct {
	X            int // Tile X position
	Y            int // Tile Y position
	Floor        int // Floor level (0 = ground floor)
	Type         SupportType
	Health       float64 // 0.0-1.0
	LoadBearing  bool    // Is this critical for structural stability?
	LoadCapacity float64 // Weight capacity (normalized 0.0-1.0)
	CurrentLoad  float64 // Current load (0.0-1.0)
}

// SupportType represents different types of structural supports
type SupportType int

const (
	SupportWall SupportType = iota
	SupportColumn
	SupportBeam
	SupportFoundation
)

// String returns the string representation of SupportType.
// Returns "Wall", "Column", "Beam", "Foundation", or "Unknown" for invalid types.
func (s SupportType) String() string {
	switch s {
	case SupportWall:
		return "Wall"
	case SupportColumn:
		return "Column"
	case SupportBeam:
		return "Beam"
	case SupportFoundation:
		return "Foundation"
	default:
		return "Unknown"
	}
}

// DamageArea represents a damaged region in a building
type DamageArea struct {
	X            int     // Center X position
	Y            int     // Center Y position
	Floor        int     // Floor level
	Radius       float64 // Damage radius in tiles
	Severity     float64 // Damage severity (0.0-1.0)
	PropagateAge float64 // Age of damage for propagation (seconds)
}

// DebrisParticle represents a piece of falling debris
type DebrisParticle struct {
	X        float64 // World X position
	Y        float64 // World Y position
	VelX     float64 // X velocity
	VelY     float64 // Y velocity
	RotVel   float64 // Rotational velocity (radians/sec)
	Angle    float64 // Current rotation angle
	Mass     float64 // Mass (affects gravity and bouncing)
	Size     float64 // Particle size (pixels)
	Material MaterialType
	Color    color.RGBA
	Life     float64 // Remaining lifetime (seconds)
}

// MaterialType represents different building materials
type MaterialType int

const (
	MaterialWood MaterialType = iota
	MaterialStone
	MaterialMetal
	MaterialGlass
	MaterialConcrete
	MaterialBrick
)

// String returns the string representation of MaterialType.
// Returns "Wood", "Stone", "Metal", "Glass", "Concrete", "Brick", or "Unknown" for invalid types.
func (m MaterialType) String() string {
	switch m {
	case MaterialWood:
		return "Wood"
	case MaterialStone:
		return "Stone"
	case MaterialMetal:
		return "Metal"
	case MaterialGlass:
		return "Glass"
	case MaterialConcrete:
		return "Concrete"
	case MaterialBrick:
		return "Brick"
	default:
		return "Unknown"
	}
}

// GetMaterialProperties returns physical properties for a material type
func GetMaterialProperties(material MaterialType) MaterialProperties {
	switch material {
	case MaterialWood:
		return MaterialProperties{
			Density:    0.6,
			Bounciness: 0.3,
			Friction:   0.4,
			Durability: 0.5,
			Flammable:  true,
		}
	case MaterialStone:
		return MaterialProperties{
			Density:    0.9,
			Bounciness: 0.1,
			Friction:   0.6,
			Durability: 0.9,
			Flammable:  false,
		}
	case MaterialMetal:
		return MaterialProperties{
			Density:    0.8,
			Bounciness: 0.4,
			Friction:   0.3,
			Durability: 0.95,
			Flammable:  false,
		}
	case MaterialGlass:
		return MaterialProperties{
			Density:    0.5,
			Bounciness: 0.2,
			Friction:   0.1,
			Durability: 0.1,
			Flammable:  false,
		}
	case MaterialConcrete:
		return MaterialProperties{
			Density:    1.0,
			Bounciness: 0.05,
			Friction:   0.7,
			Durability: 0.85,
			Flammable:  false,
		}
	case MaterialBrick:
		return MaterialProperties{
			Density:    0.7,
			Bounciness: 0.15,
			Friction:   0.5,
			Durability: 0.75,
			Flammable:  false,
		}
	default:
		// Log warning for unknown material type to help catch bugs
		log.WithFields(log.Fields{
			"material":  int(material),
			"component": "destruction",
			"function":  "GetMaterialProperties",
		}).Warn("Unknown material type, using default properties")
		return MaterialProperties{
			Density:    0.5,
			Bounciness: 0.3,
			Friction:   0.4,
			Durability: 0.5,
			Flammable:  false,
		}
	}
}

// MaterialProperties defines physical properties of materials
type MaterialProperties struct {
	Density    float64 // Mass per unit volume (0.0-1.0)
	Bounciness float64 // Restitution coefficient (0.0-1.0)
	Friction   float64 // Friction coefficient (0.0-1.0)
	Durability float64 // Resistance to damage (0.0-1.0)
	Flammable  bool    // Can catch fire
}

// FallingObject represents a physics-simulated falling object
type FallingObject struct {
	X          float64
	Y          float64
	Z          float64
	VelX       float64
	VelY       float64
	VelZ       float64
	RotVel     float64
	Angle      float64
	Mass       float64
	Width      float64
	Height     float64
	Material   MaterialType
	OnGround   bool
	Bounces    int
	MaxBounces int
}

// IsGrounded checks if the falling object has settled on the ground
func (f *FallingObject) IsGrounded() bool {
	return f.OnGround || (f.Z <= 0.1 && math.Abs(f.VelZ) < 0.1)
}

// Config holds configuration for the destruction system
type Config struct {
	EnableIntegrityChecks bool
	DamagePropagationRate float64
	CollapseThreshold     float64
	EnableDebris          bool
	MaxDebrisParticles    int
	DebrisLifetime        float64
	Gravity               float64
	AirResistance         float64
	GroundFriction        float64
	MaxFallingObjects     int
	UpdateFrequency       float64
	// Seed is used for deterministic debris generation. If zero, a default
	// seed (12345) is used. Required for multiplayer sync and test reproducibility.
	Seed int64
}

// DefaultConfig returns default configuration for the destruction system
func DefaultConfig() *Config {
	return &Config{
		EnableIntegrityChecks: true,
		DamagePropagationRate: 0.1,
		CollapseThreshold:     0.15,
		EnableDebris:          true,
		MaxDebrisParticles:    500,
		DebrisLifetime:        10.0,
		Gravity:               980.0,
		AirResistance:         0.05,
		GroundFriction:        0.8,
		MaxFallingObjects:     100,
		UpdateFrequency:       30.0,
		Seed:                  12345,
	}
}
