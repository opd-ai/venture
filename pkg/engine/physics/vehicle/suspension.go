// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"
)

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

// Update simulates suspension physics for one frame.
// Returns the average vertical offset to apply to vehicle rendering.
func (s *SuspensionComponent) Update(deltaTime float64, terrainHeights []float64) float64 {
	if len(terrainHeights) != len(s.Wheels) {
		// Safety check - need terrain height for each wheel
		return 0.0
	}

	s.LastUpdateTime += deltaTime

	totalVerticalOffset := 0.0
	groundedWheels := 0

	// Update each wheel's suspension
	for i := range s.Wheels {
		wheel := &s.Wheels[i]
		_ = terrainHeights[i] // Future use: for terrain height queries

		// Calculate spring force based on compression
		// F = -k * x (Hooke's law)
		compressionDistance := wheel.Compression * s.SuspensionTravel
		springForce := -wheel.SpringStiffness * compressionDistance

		// Calculate damper force based on compression rate
		// F = -c * v (damping)
		damperForce := -wheel.DamperStrength * wheel.CompressionRate

		// Total suspension force
		totalForce := springForce + damperForce

		// Calculate weight on this wheel (from weight transfer)
		// This will be modified by WeightTransferComponent
		wheelWeight := (s.TotalMass * 9.81) / float64(len(s.Wheels)) // Distribute mass evenly by default

		// Net force determines compression change
		netForce := wheelWeight + totalForce

		// Update compression rate (acceleration = F/m)
		// Approximate wheel mass as 1/10 of total mass per wheel
		wheelMass := s.TotalMass / (10.0 * float64(len(s.Wheels)))
		compressionAccel := netForce / wheelMass

		// Update compression with clamping
		wheel.CompressionRate += compressionAccel * deltaTime
		wheel.Compression += wheel.CompressionRate * deltaTime

		// Clamp compression to valid range [0, 1]
		if wheel.Compression < 0.0 {
			wheel.Compression = 0.0
			wheel.CompressionRate = 0.0
			wheel.IsGrounded = false
		} else if wheel.Compression > 1.0 {
			wheel.Compression = 1.0
			wheel.CompressionRate = 0.0 // Bottomed out
			wheel.IsGrounded = true
		} else {
			wheel.IsGrounded = true
		}

		// Store load on wheel for physics calculations
		wheel.Load = math.Abs(springForce + damperForce)

		if wheel.IsGrounded {
			// Calculate vertical offset from compression
			verticalOffset := (1.0 - wheel.Compression) * s.SuspensionTravel
			totalVerticalOffset += verticalOffset
			groundedWheels++
		}
	}

	// Return average vertical offset for grounded wheels
	if groundedWheels > 0 {
		return totalVerticalOffset / float64(groundedWheels)
	}
	return 0.0
}

// GetWheelLoad returns the load (force) on a specific wheel.
func (s *SuspensionComponent) GetWheelLoad(wheelIndex int) float64 {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return 0.0
	}
	return s.Wheels[wheelIndex].Load
}

// GetWheelCompression returns the compression ratio [0.0, 1.0] for a specific wheel.
func (s *SuspensionComponent) GetWheelCompression(wheelIndex int) float64 {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return 0.0
	}
	return s.Wheels[wheelIndex].Compression
}

// IsWheelGrounded checks if a specific wheel is in contact with terrain.
func (s *SuspensionComponent) IsWheelGrounded(wheelIndex int) bool {
	if wheelIndex < 0 || wheelIndex >= len(s.Wheels) {
		return false
	}
	return s.Wheels[wheelIndex].IsGrounded
}

// GetGroundedWheelCount returns the number of wheels currently touching terrain.
func (s *SuspensionComponent) GetGroundedWheelCount() int {
	count := 0
	for i := range s.Wheels {
		if s.Wheels[i].IsGrounded {
			count++
		}
	}
	return count
}

// SetWheelLoad allows external systems (like WeightTransferComponent) to modify wheel loading.
func (s *SuspensionComponent) SetWheelLoad(wheelIndex int, load float64) {
	if wheelIndex >= 0 && wheelIndex < len(s.Wheels) {
		s.Wheels[wheelIndex].Load = load
	}
}
