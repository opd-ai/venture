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
func (c *ColliderComponent) GetBounds(x, y float64) (minX, minY, maxX, maxY float64) {
	minX = x + c.OffsetX
	minY = y + c.OffsetY
	maxX = minX + c.Width
	maxY = minY + c.Height
	return
}

// Intersects checks if this collider intersects with another collider.
func (c *ColliderComponent) Intersects(x1, y1 float64, other *ColliderComponent, x2, y2 float64) bool {
	minX1, minY1, maxX1, maxY1 := c.GetBounds(x1, y1)
	minX2, minY2, maxX2, maxY2 := other.GetBounds(x2, y2)
	
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
