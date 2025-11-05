// Package engine provides basic physics components for the ECS.
// This file defines fundamental components: PositionComponent, VelocityComponent,
// ColliderComponent, and BoundsComponent used across all game systems.
package engine

import "math"

// PositionComponent represents an entity's position in 2D space.
type PositionComponent struct {
	X, Y float64
}

// Type returns the component type identifier.
func (p *PositionComponent) Type() string {
	return "position"
}

// VelocityComponent represents an entity's velocity in 2D space.
type VelocityComponent struct {
	VX, VY float64
}

// Type returns the component type identifier.
func (v *VelocityComponent) Type() string {
	return "velocity"
}

// ColliderComponent represents an entity's collision bounds.
// Uses axis-aligned bounding box (AABB) for efficient collision detection.
type ColliderComponent struct {
	// Width and height of the collision box
	Width, Height float64

	// Whether this collider is solid (blocks movement)
	Solid bool

	// Whether this collider is a trigger (detects collision but doesn't block)
	IsTrigger bool

	// Layer for collision filtering (0 = all layers)
	Layer int

	// Offset from position (for centered colliders)
	OffsetX, OffsetY float64
}

// Type returns the component type identifier.
func (c *ColliderComponent) Type() string {
	return "collider"
}

// GetBounds returns the axis-aligned bounding box for this collider.
// Returns min and max coordinates.
// Note: Does not account for rotation. Use GetRotatedBounds() for rotated entities.
func (c *ColliderComponent) GetBounds(x, y float64) (minX, minY, maxX, maxY float64) {
	minX = x + c.OffsetX
	minY = y + c.OffsetY
	maxX = minX + c.Width
	maxY = minY + c.Height
	return minX, minY, maxX, maxY
}

// Issue #20 FIX: GetRotatedBounds returns the axis-aligned bounding box that encompasses
// a rotated collider. This provides conservative collision detection for rotated entities.
// The returned AABB is guaranteed to fully contain the rotated collision box.
//
// Parameters:
//   - x, y: Entity position
//   - angle: Rotation angle in radians (0 = facing right, positive = clockwise)
//
// Returns the min/max coordinates of the AABB that contains the rotated collider.
func (c *ColliderComponent) GetRotatedBounds(x, y, angle float64) (minX, minY, maxX, maxY float64) {
	// If no rotation, use regular bounds for efficiency
	if angle == 0 {
		return c.GetBounds(x, y)
	}

	// Calculate the four corners of the unrotated box
	// relative to the entity position (accounting for offset)
	centerX := x + c.OffsetX + c.Width/2
	centerY := y + c.OffsetY + c.Height/2

	halfWidth := c.Width / 2
	halfHeight := c.Height / 2

	// Four corners relative to center
	corners := [4][2]float64{
		{-halfWidth, -halfHeight}, // Top-left
		{halfWidth, -halfHeight},  // Top-right
		{-halfWidth, halfHeight},  // Bottom-left
		{halfWidth, halfHeight},   // Bottom-right
	}

	// Rotate each corner and track min/max to find bounding box
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)

	minX = math.MaxFloat64
	minY = math.MaxFloat64
	maxX = -math.MaxFloat64
	maxY = -math.MaxFloat64

	for _, corner := range corners {
		// Apply 2D rotation matrix
		rotatedX := corner[0]*cosAngle - corner[1]*sinAngle + centerX
		rotatedY := corner[0]*sinAngle + corner[1]*cosAngle + centerY

		// Update bounds
		if rotatedX < minX {
			minX = rotatedX
		}
		if rotatedX > maxX {
			maxX = rotatedX
		}
		if rotatedY < minY {
			minY = rotatedY
		}
		if rotatedY > maxY {
			maxY = rotatedY
		}
	}

	return minX, minY, maxX, maxY
}

// Intersects checks if this collider intersects with another collider.
// Uses axis-aligned bounding boxes without rotation.
func (c *ColliderComponent) Intersects(x1, y1 float64, other *ColliderComponent, x2, y2 float64) bool {
	minX1, minY1, maxX1, maxY1 := c.GetBounds(x1, y1)
	minX2, minY2, maxX2, maxY2 := other.GetBounds(x2, y2)

	return !(maxX1 <= minX2 || maxX2 <= minX1 || maxY1 <= minY2 || maxY2 <= minY1)
}

// Issue #20 FIX: IntersectsRotated checks intersection accounting for rotation angles.
// Uses conservative AABB approach - computes rotated bounding boxes and checks intersection.
// This is more expensive than Intersects() but handles rotated entities correctly.
//
// Parameters:
//   - x1, y1: First entity position
//   - angle1: First entity rotation in radians
//   - other: Second collider component
//   - x2, y2: Second entity position
//   - angle2: Second entity rotation in radians
//
// Returns true if the rotated bounding boxes intersect.
func (c *ColliderComponent) IntersectsRotated(x1, y1, angle1 float64, other *ColliderComponent, x2, y2, angle2 float64) bool {
	minX1, minY1, maxX1, maxY1 := c.GetRotatedBounds(x1, y1, angle1)
	minX2, minY2, maxX2, maxY2 := other.GetRotatedBounds(x2, y2, angle2)

	return !(maxX1 <= minX2 || maxX2 <= minX1 || maxY1 <= minY2 || maxY2 <= minY1)
}

// BoundsComponent represents world boundaries for an entity.
type BoundsComponent struct {
	// Minimum and maximum coordinates
	MinX, MinY float64
	MaxX, MaxY float64

	// Whether to wrap around boundaries (for infinite worlds)
	Wrap bool
}

// Type returns the component type identifier.
func (b *BoundsComponent) Type() string {
	return "bounds"
}

// Clamp restricts a position to within the bounds.
func (b *BoundsComponent) Clamp(x, y float64) (float64, float64) {
	if b.Wrap {
		// Wrap around
		if x < b.MinX {
			x = b.MaxX - (b.MinX - x)
		} else if x > b.MaxX {
			x = b.MinX + (x - b.MaxX)
		}
		if y < b.MinY {
			y = b.MaxY - (b.MinY - y)
		} else if y > b.MaxY {
			y = b.MinY + (y - b.MaxY)
		}
	} else {
		// Clamp to bounds
		x = math.Max(b.MinX, math.Min(b.MaxX, x))
		y = math.Max(b.MinY, math.Min(b.MaxY, y))
	}
	return x, y
}

// FrictionComponent applies drag/friction to slow down moving entities.
// Used for items dropped on death to create a realistic scatter effect.
// Priority 1.4: Loot Drop System
type FrictionComponent struct {
	// Coefficient is the friction multiplier (0.0 = no friction, 1.0 = stops instantly)
	// Typical values: 0.05-0.15 for smooth deceleration
	Coefficient float64
}

// Type returns the component type identifier.
func (f *FrictionComponent) Type() string {
	return "friction"
}

// NewFrictionComponent creates a friction component with the specified coefficient.
func NewFrictionComponent(coefficient float64) *FrictionComponent {
	return &FrictionComponent{
		Coefficient: coefficient,
	}
}
