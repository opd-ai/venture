// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

// SuspensionComponent manages vehicle suspension system with spring-damper model.
type SuspensionComponent struct {
	// Wheels attached to vehicle
	Wheels []Wheel

	// Global suspension parameters
	TotalMass        float64 // Vehicle + cargo mass (kg)
	SuspensionTravel float64 // Maximum suspension travel distance (pixels)

	// Performance metrics
	LastUpdateTime float64
}

// Type returns the component type identifier.
func (s *SuspensionComponent) Type() string {
	return "suspension"
}

// NewSuspensionComponent creates a suspension component with the specified number of wheels.
func NewSuspensionComponent(wheelCount int) *SuspensionComponent {
	if wheelCount < 1 {
		wheelCount = 4 // Default to 4-wheel suspension
	}

	wheels := make([]Wheel, wheelCount)

	// Configure wheel positions based on count
	switch wheelCount {
	case 1:
		// Single wheel (unicycle/motorcycle)
		wheels[0] = Wheel{LocalX: 0, LocalY: 0}
	case 2:
		// Two wheels (motorcycle/bike)
		wheels[0] = Wheel{LocalX: -16, LocalY: 0} // Front
		wheels[1] = Wheel{LocalX: 16, LocalY: 0}  // Rear
	case 3:
		// Three wheels (trike)
		wheels[0] = Wheel{LocalX: -16, LocalY: 0} // Front
		wheels[1] = Wheel{LocalX: 12, LocalY: -8} // Rear left
		wheels[2] = Wheel{LocalX: 12, LocalY: 8}  // Rear right
	case 4:
		// Four wheels (car/cart) - most common
		wheels[0] = Wheel{LocalX: -16, LocalY: -8} // Front left
		wheels[1] = Wheel{LocalX: -16, LocalY: 8}  // Front right
		wheels[2] = Wheel{LocalX: 16, LocalY: -8}  // Rear left
		wheels[3] = Wheel{LocalX: 16, LocalY: 8}   // Rear right
	default:
		// More than 4 wheels - arrange in rows
		// For trucks, tanks, etc.
		rowCount := (wheelCount + 1) / 2
		for i := 0; i < wheelCount; i++ {
			row := i / 2
			side := i % 2
			xPos := -20.0 + float64(row)*40.0/float64(rowCount-1)
			yPos := -8.0 + float64(side)*16.0
			wheels[i] = Wheel{LocalX: xPos, LocalY: yPos}
		}
	}

	// Initialize wheel parameters
	restLength := 10.0        // 10 pixels natural spring length
	springStiffness := 5000.0 // Stiff spring
	damperStrength := 500.0   // Moderate damping

	for i := range wheels {
		wheels[i].RestLength = restLength
		wheels[i].SpringStiffness = springStiffness
		wheels[i].DamperStrength = damperStrength
		wheels[i].Compression = 0.5 // Start at 50% compression
		wheels[i].IsGrounded = true
	}

	return &SuspensionComponent{
		Wheels:           wheels,
		TotalMass:        1000.0, // Default 1000kg
		SuspensionTravel: 20.0,   // 20 pixels max travel
		LastUpdateTime:   0.0,
	}
}

// GetWheelLoad returns the load (force) on a specific wheel.
// Deprecated: Use vehicle.GetWheelLoad(suspension, wheelIndex) instead to maintain ECS purity.
func (s *SuspensionComponent) GetWheelLoad(wheelIndex int) float64 {
	return GetWheelLoad(s, wheelIndex)
}

// GetWheelCompression returns the compression ratio [0.0, 1.0] for a specific wheel.
// Deprecated: Use vehicle.GetWheelCompression(suspension, wheelIndex) instead to maintain ECS purity.
func (s *SuspensionComponent) GetWheelCompression(wheelIndex int) float64 {
	return GetWheelCompression(s, wheelIndex)
}

// IsWheelGrounded checks if a specific wheel is in contact with terrain.
// Deprecated: Use vehicle.IsWheelGrounded(suspension, wheelIndex) instead to maintain ECS purity.
func (s *SuspensionComponent) IsWheelGrounded(wheelIndex int) bool {
	return IsWheelGrounded(s, wheelIndex)
}

// GetGroundedWheelCount returns the number of wheels currently touching terrain.
// Deprecated: Use vehicle.GetGroundedWheelCount(suspension) instead to maintain ECS purity.
func (s *SuspensionComponent) GetGroundedWheelCount() int {
	return GetGroundedWheelCount(s)
}

// SetWheelLoad sets the load force on a specific wheel.
// Used by the weight transfer system to modify individual wheel loading.
// Deprecated: Use vehicle.SetWheelLoad(suspension, wheelIndex, load) instead to maintain ECS purity.
func (s *SuspensionComponent) SetWheelLoad(wheelIndex int, load float64) {
	SetWheelLoad(s, wheelIndex, load)
}
