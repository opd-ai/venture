// Package engine provides aiming functionality for entities.
// This file implements AimComponent which manages independent aim direction
// separate from movement direction, enabling dual-stick shooter mechanics.
package engine

import "math"

// Vector2D represents a 2D position or direction
type Vector2D struct {
	X float64
	Y float64
}

// AimComponent stores the aim direction for an entity, separate from movement.
// This enables dual-stick shooter mechanics where movement (WASD) and aim
// (mouse/right joystick) are independent. The component supports both direct
// angle specification and target-based aiming (e.g., mouse cursor position).
type AimComponent struct {
	// AimAngle is the current aim direction in radians [0, 2π)
	// Uses same convention as RotationComponent (0=right, π/2=down, π=left, 3π/2=up)
	AimAngle float64

	// AimTarget is the world-space position being aimed at
	// Used for mouse/touch aiming where target position is known
	AimTarget Vector2D

	// HasTarget indicates if AimTarget is valid
	// When true, AimAngle is calculated from entity position to AimTarget
	// When false, AimAngle is used directly
	HasTarget bool

	// AutoAim enables aim assist for mobile/controller input
	// When enabled, aim automatically snaps to nearby enemies within SnapRadius
	AutoAim bool

	// SnapRadius is the maximum distance for auto-aim targeting (pixels)
	// Enemies within this radius are candidates for auto-aim
	// Default: 100 pixels
	SnapRadius float64

	// AutoAimStrength controls how much auto-aim affects the aim direction
	// 0.0 = no auto-aim, 1.0 = full snap to target
	// Values between allow partial aim correction
	AutoAimStrength float64
}

// Type returns the component type identifier.
func (a *AimComponent) Type() string {
	return "aim"
}

// NewAimComponent creates an aim component with default values.
// initialAngle: starting aim direction in radians
func NewAimComponent(initialAngle float64) *AimComponent {
	return &AimComponent{
		AimAngle:        normalizeAngle(initialAngle),
		AimTarget:       Vector2D{0, 0},
		HasTarget:       false,
		AutoAim:         false,
		SnapRadius:      100.0,
		AutoAimStrength: 0.3, // Subtle aim assist
	}
}

// SetAimAngle directly sets the aim direction.
// Use this for gamepad right-stick input or when aim direction is known.
func (a *AimComponent) SetAimAngle(angle float64) {
	a.AimAngle = normalizeAngle(angle)
	a.HasTarget = false
}

// SetAimTarget sets the world-space position to aim at.
// Use this for mouse/touch input where target position is known.
// The aim angle will be calculated when UpdateAimAngle is called.
func (a *AimComponent) SetAimTarget(targetX, targetY float64) {
	a.AimTarget.X = targetX
	a.AimTarget.Y = targetY
	a.HasTarget = true
}

// ClearAimTarget disables target-based aiming.
// AimAngle will be used directly until a new target is set.
func (a *AimComponent) ClearAimTarget() {
	a.HasTarget = false
}

// UpdateAimAngle calculates aim angle from entity position to target.
// entityX, entityY: current entity world position
// Returns the calculated aim angle in radians.
// If HasTarget is false, returns current AimAngle unchanged.
func (a *AimComponent) UpdateAimAngle(entityX, entityY float64) float64 {
	if !a.HasTarget {
		return a.AimAngle
	}

	dx := a.AimTarget.X - entityX
	dy := a.AimTarget.Y - entityY

	// Calculate angle from entity to target
	a.AimAngle = normalizeAngle(math.Atan2(dy, dx))
	return a.AimAngle
}

// GetAimDirection returns the unit vector in the aim direction.
// Returns (x, y) where x = cos(AimAngle), y = sin(AimAngle)
func (a *AimComponent) GetAimDirection() (float64, float64) {
	return math.Cos(a.AimAngle), math.Sin(a.AimAngle)
}

// GetAttackOrigin calculates the position where a projectile should spawn.
// entityX, entityY: entity center position
// weaponOffset: distance from entity center to weapon (pixels)
// Returns (x, y) position offset in aim direction
func (a *AimComponent) GetAttackOrigin(entityX, entityY, weaponOffset float64) (float64, float64) {
	dx, dy := a.GetAimDirection()
	return entityX + dx*weaponOffset, entityY + dy*weaponOffset
}

// ApplyAutoAim adjusts aim angle towards the nearest enemy within snap radius.
// nearestEnemyX, nearestEnemyY: position of closest enemy
// entityX, entityY: current entity position
// Returns true if auto-aim was applied
func (a *AimComponent) ApplyAutoAim(entityX, entityY, nearestEnemyX, nearestEnemyY float64) bool {
	if !a.AutoAim || a.AutoAimStrength <= 0 {
		return false
	}

	// Calculate distance to enemy
	dx := nearestEnemyX - entityX
	dy := nearestEnemyY - entityY
	distSq := dx*dx + dy*dy

	// Check if within snap radius
	if distSq > a.SnapRadius*a.SnapRadius {
		return false
	}

	// Calculate angle to enemy
	targetAngle := normalizeAngle(math.Atan2(dy, dx))

	// Interpolate between current aim and target based on strength
	// Use shortest angular distance for smooth interpolation
	angularDiff := shortestAngularDistance(a.AimAngle, targetAngle)
	a.AimAngle = normalizeAngle(a.AimAngle + angularDiff*a.AutoAimStrength)

	return true
}

// IsAimingAt checks if the aim direction points towards a target position.
// targetX, targetY: position to check
// entityX, entityY: current entity position
// tolerance: angle tolerance in radians (e.g., π/16 = ~11 degrees)
// Returns true if aim is within tolerance of target direction
func (a *AimComponent) IsAimingAt(entityX, entityY, targetX, targetY, tolerance float64) bool {
	// Calculate angle to target
	dx := targetX - entityX
	dy := targetY - entityY
	angleToTarget := normalizeAngle(math.Atan2(dy, dx))

	// Check if current aim is within tolerance
	angularDiff := math.Abs(shortestAngularDistance(a.AimAngle, angleToTarget))
	return angularDiff <= tolerance
}
