// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

// TerrainType identifies different terrain surfaces for deformation.
// Originally from: terrain_deformation.go
type TerrainType int

const (
	TerrainHard  TerrainType = iota // Stone, concrete - no deformation
	TerrainFirm                     // Packed dirt - minimal deformation
	TerrainSoft                     // Mud, sand - significant deformation
	TerrainSnow                     // Snow - deep tracks
	TerrainWater                    // Water - no permanent tracks
)

// String returns the string representation of terrain type.
func (t TerrainType) String() string {
	switch t {
	case TerrainHard:
		return "Hard"
	case TerrainFirm:
		return "Firm"
	case TerrainSoft:
		return "Soft"
	case TerrainSnow:
		return "Snow"
	case TerrainWater:
		return "Water"
	default:
		return "Unknown"
	}
}

// ImpactResult contains the results of processing a collision.
// Originally from: collision_response.go
type ImpactResult struct {
	DamageDealt       float64 // Damage to vehicle durability
	VelocityReduction float64 // How much velocity was lost
	BounceVelocityX   float64 // Velocity after bounce
	BounceVelocityY   float64
	IntegrityLoss     float64 // Reduction in structural integrity
}

// Wheel represents a single wheel with suspension physics.
// Originally from: suspension.go
type Wheel struct {
	// Position relative to vehicle center (local coordinates)
	LocalX float64
	LocalY float64

	// Suspension parameters
	RestLength      float64 // Natural length of suspension spring (pixels)
	Compression     float64 // Current compression amount (0.0 = fully extended, 1.0 = fully compressed)
	CompressionRate float64 // Rate of compression change

	// Spring-damper model parameters
	SpringStiffness float64 // Spring constant (N/m equivalent)
	DamperStrength  float64 // Damping coefficient

	// Load on this wheel (affected by weight transfer)
	Load float64 // Force in newtons (approximated)

	// Contact with terrain
	IsGrounded bool
}

// VehicleState contains current vehicle dynamics for physics calculations.
// Originally from: system.go
type VehicleState struct {
	PositionX     float64
	PositionY     float64
	VelocityX     float64
	VelocityY     float64
	Rotation      float64 // Radians
	AngularVel    float64 // Radians/s
	Speed         float64 // Magnitude of velocity
	Acceleration  float64
	IsGrounded    bool
	TerrainHeight []float64 // Height at each wheel position
	TerrainTypes  []TerrainType
}

// TrackMark represents a single tire track deformation in terrain.
// Originally from: terrain_deformation.go
type TrackMark struct {
	X           float64 // World position
	Y           float64
	Angle       float64     // Track orientation (radians)
	Depth       float64     // Deformation depth (0.0 = none, 1.0 = max)
	Width       float64     // Track width (pixels)
	Age         float64     // Time since creation (seconds)
	TerrainType TerrainType // Terrain surface type
	FadeTime    float64     // Time to fade completely (seconds)
}
