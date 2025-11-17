// Package engine provides pixel-perfect collision detection with sub-pixel precision.
// This file implements Phase 48: Pixel-Perfect Collision features including
// 0.1-pixel precision, smooth wall sliding, and collision shapes for rounded corners.
package engine

import (
	"math"
)

// CollisionPrecision defines the sub-pixel precision for collision detection.
// Set to 0.1 pixels for Phase 48 requirements.
const CollisionPrecision = 0.1

// EdgeNormal represents a surface normal vector for smooth wall sliding.
type EdgeNormal struct {
	NX, NY float64 // Normalized direction perpendicular to surface
}

// CollisionShape defines geometric collision primitives beyond simple AABB.
type CollisionShape int

const (
	// ShapeAABB is axis-aligned bounding box (default, fastest)
	ShapeAABB CollisionShape = iota
	// ShapeCircle for rounded collision detection
	ShapeCircle
	// ShapeRoundedRect for rounded corners (hybrid AABB + circles at corners)
	ShapeRoundedRect
)

// PreciseColliderComponent extends ColliderComponent with pixel-perfect features.
type PreciseColliderComponent struct {
	// Base collider properties
	Width, Height    float64
	OffsetX, OffsetY float64
	Solid            bool
	IsTrigger        bool
	Layer            int

	// Precise collision features
	Shape        CollisionShape
	CornerRadius float64 // For ShapeRoundedRect

	// Edge normals for wall sliding (computed dynamically for terrain)
	EdgeNormals []EdgeNormal
}

// Type returns the component type identifier.
func (p *PreciseColliderComponent) Type() string {
	return "precise_collider"
}

// QuantizePosition rounds a position to the nearest collision precision unit.
// This ensures consistent collision detection at 0.1-pixel precision.
func QuantizePosition(x, y float64) (float64, float64) {
	qx := math.Round(x/CollisionPrecision) * CollisionPrecision
	qy := math.Round(y/CollisionPrecision) * CollisionPrecision
	return qx, qy
}

// GetBounds returns the axis-aligned bounding box for this collider.
// Returns min and max coordinates with sub-pixel precision.
func (p *PreciseColliderComponent) GetBounds(x, y float64) (minX, minY, maxX, maxY float64) {
	minX = x + p.OffsetX
	minY = y + p.OffsetY
	maxX = minX + p.Width
	maxY = minY + p.Height

	// Quantize to collision precision
	minX, minY = QuantizePosition(minX, minY)
	maxX, maxY = QuantizePosition(maxX, maxY)

	return minX, minY, maxX, maxY
}

// IntersectsAABB checks AABB intersection with sub-pixel precision.
func (p *PreciseColliderComponent) IntersectsAABB(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	minX1, minY1, maxX1, maxY1 := p.GetBounds(x1, y1)
	minX2, minY2, maxX2, maxY2 := other.GetBounds(x2, y2)

	// Use epsilon for floating-point comparison at 0.1px precision
	epsilon := CollisionPrecision / 2.0

	return !(maxX1 <= minX2+epsilon ||
		maxX2 <= minX1+epsilon ||
		maxY1 <= minY2+epsilon ||
		maxY2 <= minY1+epsilon)
}

// IntersectsCircle checks circle-circle intersection.
func (p *PreciseColliderComponent) IntersectsCircle(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// Calculate centers
	centerX1 := x1 + p.OffsetX + p.Width/2.0
	centerY1 := y1 + p.OffsetY + p.Height/2.0
	centerX2 := x2 + other.OffsetX + other.Width/2.0
	centerY2 := y2 + other.OffsetY + other.Height/2.0

	// Use average of width/height as radius
	radius1 := (p.Width + p.Height) / 4.0
	radius2 := (other.Width + other.Height) / 4.0

	// Distance between centers
	dx := centerX2 - centerX1
	dy := centerY2 - centerY1
	distSq := dx*dx + dy*dy
	radiusSum := radius1 + radius2

	return distSq <= radiusSum*radiusSum+CollisionPrecision
}

// IntersectsRoundedRect checks intersection with rounded rectangle.
// Uses hybrid approach: AABB for core, circles for corners.
func (p *PreciseColliderComponent) IntersectsRoundedRect(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// First check core AABB (excluding corner radius)
	coreMinX1 := x1 + p.OffsetX + p.CornerRadius
	coreMinY1 := y1 + p.OffsetY + p.CornerRadius
	coreMaxX1 := x1 + p.OffsetX + p.Width - p.CornerRadius
	coreMaxY1 := y1 + p.OffsetY + p.Height - p.CornerRadius

	coreMinX2 := x2 + other.OffsetX + other.CornerRadius
	coreMinY2 := y2 + other.OffsetY + other.CornerRadius
	coreMaxX2 := x2 + other.OffsetX + other.Width - other.CornerRadius
	coreMaxY2 := y2 + other.OffsetY + other.Height - other.CornerRadius

	epsilon := CollisionPrecision / 2.0

	// Check if AABBs overlap
	if !(coreMaxX1 <= coreMinX2+epsilon ||
		coreMaxX2 <= coreMinX1+epsilon ||
		coreMaxY1 <= coreMinY2+epsilon ||
		coreMaxY2 <= coreMinY1+epsilon) {
		return true
	}

	// Check corner circles if AABBs don't overlap
	corners1 := p.getCornerPositions(x1, y1)
	corners2 := other.getCornerPositions(x2, y2)

	for _, c1 := range corners1 {
		for _, c2 := range corners2 {
			dx := c2[0] - c1[0]
			dy := c2[1] - c1[1]
			distSq := dx*dx + dy*dy
			radiusSum := p.CornerRadius + other.CornerRadius
			if distSq <= radiusSum*radiusSum+CollisionPrecision {
				return true
			}
		}
	}

	return false
}

// getCornerPositions returns the center positions of the four corner circles.
func (p *PreciseColliderComponent) getCornerPositions(x, y float64) [4][2]float64 {
	baseX := x + p.OffsetX
	baseY := y + p.OffsetY

	return [4][2]float64{
		{baseX + p.CornerRadius, baseY + p.CornerRadius},                      // Top-left
		{baseX + p.Width - p.CornerRadius, baseY + p.CornerRadius},            // Top-right
		{baseX + p.CornerRadius, baseY + p.Height - p.CornerRadius},           // Bottom-left
		{baseX + p.Width - p.CornerRadius, baseY + p.Height - p.CornerRadius}, // Bottom-right
	}
}

// Intersects checks intersection based on shape type.
func (p *PreciseColliderComponent) Intersects(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// Use most specific shape check
	if p.Shape == ShapeCircle && other.Shape == ShapeCircle {
		return p.IntersectsCircle(x1, y1, other, x2, y2)
	}

	if (p.Shape == ShapeRoundedRect || other.Shape == ShapeRoundedRect) &&
		(p.CornerRadius > 0 || other.CornerRadius > 0) {
		return p.IntersectsRoundedRect(x1, y1, other, x2, y2)
	}

	// Default to AABB
	return p.IntersectsAABB(x1, y1, other, x2, y2)
}

// ComputeWallNormal calculates the surface normal for a wall collision.
// This enables smooth wall sliding instead of getting stuck on walls.
// Returns a normalized vector perpendicular to the wall surface.
func ComputeWallNormal(entityX, entityY, wallX, wallY float64) EdgeNormal {
	// Vector from wall to entity
	dx := entityX - wallX
	dy := entityY - wallY

	// Normalize
	length := math.Sqrt(dx*dx + dy*dy)
	if length < CollisionPrecision {
		// Default to upward normal if too close
		return EdgeNormal{NX: 0, NY: -1}
	}

	return EdgeNormal{
		NX: dx / length,
		NY: dy / length,
	}
}

// ApplyWallSlide projects velocity along a wall surface for smooth sliding.
// Takes current velocity and wall normal, returns adjusted velocity that slides along the wall.
func ApplyWallSlide(vx, vy float64, normal EdgeNormal) (newVX, newVY float64) {
	// Project velocity onto wall tangent (perpendicular to normal)
	// Tangent is (-normal.NY, normal.NX) for 2D
	tangentX := -normal.NY
	tangentY := normal.NX

	// Dot product with tangent
	dotProduct := vx*tangentX + vy*tangentY

	// Project velocity onto tangent
	newVX = dotProduct * tangentX
	newVY = dotProduct * tangentY

	return newVX, newVY
}

// ResolveWallCollision resolves collision with terrain wall using smooth sliding.
// Updates entity position to slide along wall instead of stopping.
// Returns the adjusted position.
func ResolveWallCollision(entity *Entity, normal EdgeNormal) (adjustedX, adjustedY float64) {
	posComp, hasPoscomp := entity.GetComponent("position")
	if !hasPoscomp {
		return 0, 0
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return 0, 0
	}

	// Get velocity if present
	velComp, hasVel := entity.GetComponent("velocity")
	if hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			// Apply wall slide to velocity
			vel.VX, vel.VY = ApplyWallSlide(vel.VX, vel.VY, normal)
		}
	}

	// Push entity slightly away from wall along normal
	pushDistance := CollisionPrecision * 2.0 // Push 0.2 pixels away
	adjustedX = pos.X + normal.NX*pushDistance
	adjustedY = pos.Y + normal.NY*pushDistance

	// Quantize to maintain precision
	adjustedX, adjustedY = QuantizePosition(adjustedX, adjustedY)

	return adjustedX, adjustedY
}

// CheckPreciseCollision performs pixel-perfect collision check between two entities.
// Returns true if entities collide at sub-pixel precision (0.1px).
func CheckPreciseCollision(e1, e2 *Entity) bool {
	// Try precise collider first
	pc1Comp, has1 := e1.GetComponent("precise_collider")
	pc2Comp, has2 := e2.GetComponent("precise_collider")

	if has1 && has2 {
		pc1, ok1 := pc1Comp.(*PreciseColliderComponent)
		pc2, ok2 := pc2Comp.(*PreciseColliderComponent)

		if ok1 && ok2 {
			pos1Comp, hasPos1 := e1.GetComponent("position")
			pos2Comp, hasPos2 := e2.GetComponent("position")

			if hasPos1 && hasPos2 {
				if p1, ok := pos1Comp.(*PositionComponent); ok {
					if p2, ok := pos2Comp.(*PositionComponent); ok {
						return pc1.Intersects(p1.X, p1.Y, pc2, p2.X, p2.Y)
					}
				}
			}
		}
	}

	// Fallback to standard collider
	c1Comp, has1 := e1.GetComponent("collider")
	c2Comp, has2 := e2.GetComponent("collider")

	if has1 && has2 {
		c1, ok1 := c1Comp.(*ColliderComponent)
		c2, ok2 := c2Comp.(*ColliderComponent)

		if ok1 && ok2 {
			pos1Comp, hasPos1 := e1.GetComponent("position")
			pos2Comp, hasPos2 := e2.GetComponent("position")

			if hasPos1 && hasPos2 {
				if p1, ok := pos1Comp.(*PositionComponent); ok {
					if p2, ok := pos2Comp.(*PositionComponent); ok {
						return c1.Intersects(p1.X, p1.Y, c2, p2.X, p2.Y)
					}
				}
			}
		}
	}

	return false
}

// GetCollisionAlignment calculates the visual/collision alignment error.
// Returns the distance between visual center and collision center.
// Used for debug visualization to ensure <0.5px alignment.
func GetCollisionAlignment(entity *Entity) (alignmentError float64) {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return 0
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return 0
	}

	// Get collision bounds
	var collisionCenterX, collisionCenterY float64

	if pcComp, hasPrecise := entity.GetComponent("precise_collider"); hasPrecise {
		if pc, ok := pcComp.(*PreciseColliderComponent); ok {
			minX, minY, maxX, maxY := pc.GetBounds(pos.X, pos.Y)
			collisionCenterX = (minX + maxX) / 2.0
			collisionCenterY = (minY + maxY) / 2.0
		}
	} else if cComp, hasCollider := entity.GetComponent("collider"); hasCollider {
		if c, ok := cComp.(*ColliderComponent); ok {
			minX, minY, maxX, maxY := c.GetBounds(pos.X, pos.Y)
			collisionCenterX = (minX + maxX) / 2.0
			collisionCenterY = (minY + maxY) / 2.0
		}
	} else {
		return 0
	}

	// Visual center is entity position (sprite is centered)
	visualCenterX := pos.X
	visualCenterY := pos.Y

	// Calculate distance
	dx := collisionCenterX - visualCenterX
	dy := collisionCenterY - visualCenterY
	alignmentError = math.Sqrt(dx*dx + dy*dy)

	return alignmentError
}
