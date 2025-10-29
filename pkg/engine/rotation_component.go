// Package engine provides rotation functionality for entities.
// This file implements RotationComponent which stores entity facing direction
// in 2D space, enabling 360° rotation for enhanced combat and movement.
package engine

import "math"

// RotationComponent stores the facing direction of an entity in 2D space.
// Angles are measured in radians with the following convention:
//
//	0 radians = facing right (positive X axis)
//	π/2 radians = facing down (positive Y axis)
//	π radians = facing left (negative X axis)
//	3π/2 radians = facing up (negative Y axis)
//
// The component supports smooth rotation via angular velocity and configurable
// rotation speed limits. This enables dual-stick shooter mechanics where
// movement direction is independent from facing direction.
type RotationComponent struct {
	// Angle is the current facing direction in radians [0, 2π)
	Angle float64

	// TargetAngle is the desired facing direction for smooth rotation
	TargetAngle float64

	// AngularVelocity is the current rotation speed in radians per second
	// Positive values rotate clockwise, negative counter-clockwise
	AngularVelocity float64

	// RotationSpeed is the maximum rotation rate in radians per second
	// Default: 3.0 rad/s provides responsive but smooth rotation
	RotationSpeed float64

	// SmoothRotation enables interpolation between current and target angles
	// When true, rotation uses AngularVelocity for smooth transitions
	// When false, Angle snaps instantly to TargetAngle
	SmoothRotation bool
}

// Type returns the component type identifier.
func (r *RotationComponent) Type() string {
	return "rotation"
}

// NewRotationComponent creates a rotation component with default values.
// initialAngle: starting facing direction in radians
// rotationSpeed: maximum rotation rate in radians per second (0 = use default 3.0)
func NewRotationComponent(initialAngle, rotationSpeed float64) *RotationComponent {
	if rotationSpeed <= 0 {
		rotationSpeed = 3.0 // Default: ~172 degrees per second
	}

	return &RotationComponent{
		Angle:           normalizeAngle(initialAngle),
		TargetAngle:     normalizeAngle(initialAngle),
		AngularVelocity: 0,
		RotationSpeed:   rotationSpeed,
		SmoothRotation:  true,
	}
}

// SetTargetAngle sets the desired facing direction.
// The entity will rotate towards this angle at RotationSpeed rate.
func (r *RotationComponent) SetTargetAngle(angle float64) {
	r.TargetAngle = normalizeAngle(angle)
}

// SetAngleImmediate instantly sets the facing direction without interpolation.
// Use this for teleports, respawns, or when instant rotation is desired.
func (r *RotationComponent) SetAngleImmediate(angle float64) {
	angle = normalizeAngle(angle)
	r.Angle = angle
	r.TargetAngle = angle
	r.AngularVelocity = 0
}

// Update performs smooth rotation interpolation.
// deltaTime: elapsed time in seconds since last update
// Returns true if rotation is complete (Angle == TargetAngle)
func (r *RotationComponent) Update(deltaTime float64) bool {
	if !r.SmoothRotation {
		r.Angle = r.TargetAngle
		r.AngularVelocity = 0
		return true
	}

	// Check if already at target (with small epsilon for floating point)
	angleDiff := shortestAngularDistance(r.Angle, r.TargetAngle)
	if math.Abs(angleDiff) < 0.01 { // ~0.57 degrees tolerance
		r.Angle = r.TargetAngle
		r.AngularVelocity = 0
		return true
	}

	// Calculate required rotation direction and speed
	rotationDirection := 1.0
	if angleDiff < 0 {
		rotationDirection = -1.0
	}

	// Apply rotation with speed limit
	maxRotation := r.RotationSpeed * deltaTime
	actualRotation := math.Min(math.Abs(angleDiff), maxRotation)
	r.AngularVelocity = rotationDirection * actualRotation / deltaTime

	r.Angle = normalizeAngle(r.Angle + rotationDirection*actualRotation)

	// Check if rotation is now complete after the update
	if actualRotation >= math.Abs(angleDiff) {
		r.Angle = r.TargetAngle
		r.AngularVelocity = 0
		return true
	}

	return false
}

// GetDirectionVector returns the unit vector in the facing direction.
// Useful for calculating attack origins, forward movement, etc.
// Returns (x, y) where x = cos(Angle), y = sin(Angle)
func (r *RotationComponent) GetDirectionVector() (float64, float64) {
	return math.Cos(r.Angle), math.Sin(r.Angle)
}

// GetCardinalDirection returns the nearest cardinal direction (0-7).
// 0=right, 1=down-right, 2=down, 3=down-left, 4=left, 5=up-left, 6=up, 7=up-right
// Useful for sprite caching systems that store rotated sprites at 8 directions.
func (r *RotationComponent) GetCardinalDirection() int {
	// Divide circle into 8 equal sections (π/4 radians each)
	sector := int((r.Angle + math.Pi/8) / (math.Pi / 4))
	return sector % 8
}

// normalizeAngle constrains an angle to the range [0, 2π)
func normalizeAngle(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}

// shortestAngularDistance calculates the shortest rotation from angle1 to angle2.
// Returns positive for clockwise rotation, negative for counter-clockwise.
// Result is in the range (-π, π].
func shortestAngularDistance(angle1, angle2 float64) float64 {
	diff := normalizeAngle(angle2 - angle1)
	if diff > math.Pi {
		diff -= 2 * math.Pi
	}
	return diff
}
