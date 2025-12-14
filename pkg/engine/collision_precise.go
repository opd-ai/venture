// Package engine provides pixel-perfect collision detection with sub-pixel precision.
// This file implements Phase 48: Pixel-Perfect Collision features including
// 0.1-pixel precision, smooth wall sliding, and collision shapes for rounded corners.
package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// Package-level logger for collision operations
var collisionLog *logrus.Logger

func init() {
	collisionLog = logrus.New()
	collisionLog.SetReportCaller(true)
	collisionLog.SetLevel(logrus.InfoLevel)
}

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
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"input_x":   x,
			"input_y":   y,
			"precision": CollisionPrecision,
		}).Debug("Quantizing position")
	}

	qx := math.Round(x/CollisionPrecision) * CollisionPrecision
	qy := math.Round(y/CollisionPrecision) * CollisionPrecision

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"quantized_x": qx,
			"quantized_y": qy,
		}).Debug("Position quantized")
	}

	return qx, qy
}

// GetBounds returns the axis-aligned bounding box for this collider.
// Returns min and max coordinates with sub-pixel precision.
func (p *PreciseColliderComponent) GetBounds(x, y float64) (minX, minY, maxX, maxY float64) {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"component_type": "precise_collider",
			"position_x":     x,
			"position_y":     y,
			"width":          p.Width,
			"height":         p.Height,
			"offset_x":       p.OffsetX,
			"offset_y":       p.OffsetY,
		}).Debug("Getting collider bounds")
	}

	minX = x + p.OffsetX
	minY = y + p.OffsetY
	maxX = minX + p.Width
	maxY = minY + p.Height

	// Quantize to collision precision
	minX, minY = QuantizePosition(minX, minY)
	maxX, maxY = QuantizePosition(maxX, maxY)

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"min_x": minX,
			"min_y": minY,
			"max_x": maxX,
			"max_y": maxY,
		}).Debug("Bounds calculated and quantized")
	}

	return minX, minY, maxX, maxY
}

// IntersectsAABB checks AABB intersection with sub-pixel precision.
func (p *PreciseColliderComponent) IntersectsAABB(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	if collisionLog.GetLevel() >= logrus.DebugLevel {
		collisionLog.WithFields(logrus.Fields{
			"operation":      "intersects_aabb",
			"component_type": "precise_collider",
			"pos1_x":         x1,
			"pos1_y":         y1,
			"pos2_x":         x2,
			"pos2_y":         y2,
		}).Debug("Checking AABB intersection")
	}

	minX1, minY1, maxX1, maxY1 := p.GetBounds(x1, y1)
	minX2, minY2, maxX2, maxY2 := other.GetBounds(x2, y2)

	// Use epsilon for floating-point comparison at 0.1px precision
	epsilon := CollisionPrecision / 2.0

	intersects := !(maxX1 <= minX2+epsilon ||
		maxX2 <= minX1+epsilon ||
		maxY1 <= minY2+epsilon ||
		maxY2 <= minY1+epsilon)

	// OPTIMIZATION: Check log level before allocating Fields map
	if collisionLog.GetLevel() >= logrus.DebugLevel {
		collisionLog.WithFields(logrus.Fields{
			"intersects": intersects,
			"epsilon":    epsilon,
		}).Debug("AABB intersection check completed")
	}

	return intersects
}

// IntersectsCircle checks circle-circle intersection.
func (p *PreciseColliderComponent) IntersectsCircle(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"operation":      "intersects_circle",
			"component_type": "precise_collider",
			"pos1_x":         x1,
			"pos1_y":         y1,
			"pos2_x":         x2,
			"pos2_y":         y2,
		}).Debug("Checking circle intersection")
	}

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

	intersects := distSq <= radiusSum*radiusSum+CollisionPrecision

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"intersects":  intersects,
			"distance_sq": distSq,
			"radius_sum":  radiusSum,
			"radius1":     radius1,
			"radius2":     radius2,
		}).Debug("Circle intersection check completed")
	}

	return intersects
}

// IntersectsRoundedRect checks intersection with rounded rectangle.
// Uses hybrid approach: AABB for core, circles for corners.
func (p *PreciseColliderComponent) IntersectsRoundedRect(x1, y1 float64, other *PreciseColliderComponent, x2, y2 float64) bool {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"operation":      "intersects_rounded_rect",
			"component_type": "precise_collider",
			"pos1_x":         x1,
			"pos1_y":         y1,
			"pos2_x":         x2,
			"pos2_y":         y2,
			"corner_radius1": p.CornerRadius,
			"corner_radius2": other.CornerRadius,
		}).Debug("Checking rounded rectangle intersection")
	}

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
		if debugEnabled {
			collisionLog.WithFields(logrus.Fields{
				"result": true,
				"reason": "core_aabb_overlap",
			}).Debug("Rounded rectangle intersection: AABB cores overlap")
		}
		return true
	}

	// Check corner circles if AABBs don't overlap
	if debugEnabled {
		collisionLog.Debug("Checking corner circles for rounded rectangle intersection")
	}
	corners1 := p.getCornerPositions(x1, y1)
	corners2 := other.getCornerPositions(x2, y2)

	for i, c1 := range corners1 {
		for j, c2 := range corners2 {
			dx := c2[0] - c1[0]
			dy := c2[1] - c1[1]
			distSq := dx*dx + dy*dy
			radiusSum := p.CornerRadius + other.CornerRadius
			if distSq <= radiusSum*radiusSum+CollisionPrecision {
				if debugEnabled {
					collisionLog.WithFields(logrus.Fields{
						"result":      true,
						"reason":      "corner_collision",
						"corner1_idx": i,
						"corner2_idx": j,
						"distance_sq": distSq,
					}).Debug("Rounded rectangle intersection: corners collide")
				}
				return true
			}
		}
	}

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"result": false,
		}).Debug("Rounded rectangle intersection check completed: no collision")
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
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"operation":      "intersects",
			"component_type": "precise_collider",
			"shape1":         p.Shape,
			"shape2":         other.Shape,
			"pos1_x":         x1,
			"pos1_y":         y1,
			"pos2_x":         x2,
			"pos2_y":         y2,
		}).Debug("Determining collision check method based on shapes")
	}

	// Use most specific shape check
	if p.Shape == ShapeCircle && other.Shape == ShapeCircle {
		if debugEnabled {
			collisionLog.Debug("Using circle-circle intersection")
		}
		return p.IntersectsCircle(x1, y1, other, x2, y2)
	}

	if (p.Shape == ShapeRoundedRect || other.Shape == ShapeRoundedRect) &&
		(p.CornerRadius > 0 || other.CornerRadius > 0) {
		if debugEnabled {
			collisionLog.Debug("Using rounded rectangle intersection")
		}
		return p.IntersectsRoundedRect(x1, y1, other, x2, y2)
	}

	// Default to AABB
	if debugEnabled {
		collisionLog.Debug("Using AABB intersection (default)")
	}
	return p.IntersectsAABB(x1, y1, other, x2, y2)
}

// ComputeWallNormal calculates the surface normal for a wall collision.
// This enables smooth wall sliding instead of getting stuck on walls.
// Returns a normalized vector perpendicular to the wall surface.
func ComputeWallNormal(entityX, entityY, wallX, wallY float64) EdgeNormal {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"operation": "compute_wall_normal",
			"entity_x":  entityX,
			"entity_y":  entityY,
			"wall_x":    wallX,
			"wall_y":    wallY,
			"precision": CollisionPrecision,
		}).Debug("Computing wall normal for collision")
	}

	// Vector from wall to entity
	dx := entityX - wallX
	dy := entityY - wallY

	// Normalize
	length := math.Sqrt(dx*dx + dy*dy)
	if length < CollisionPrecision {
		// Default to upward normal if too close
		if debugEnabled {
			collisionLog.WithFields(logrus.Fields{
				"result":   "default_normal",
				"length":   length,
				"normal_x": 0.0,
				"normal_y": -1.0,
			}).Debug("Entities too close, using default upward normal")
		}
		return EdgeNormal{NX: 0, NY: -1}
	}

	normal := EdgeNormal{
		NX: dx / length,
		NY: dy / length,
	}

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"normal_x": normal.NX,
			"normal_y": normal.NY,
			"length":   length,
		}).Debug("Wall normal computed and normalized")
	}

	return normal
}

// ApplyWallSlide projects velocity along a wall surface for smooth sliding.
// Takes current velocity and wall normal, returns adjusted velocity that slides along the wall.
func ApplyWallSlide(vx, vy float64, normal EdgeNormal) (newVX, newVY float64) {
	// OPTIMIZATION: Check log level before allocating Fields map to avoid per-call allocations
	debugEnabled := collisionLog.GetLevel() >= logrus.DebugLevel
	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"operation":  "apply_wall_slide",
			"velocity_x": vx,
			"velocity_y": vy,
			"normal_x":   normal.NX,
			"normal_y":   normal.NY,
		}).Debug("Applying wall slide to velocity")
	}

	// Project velocity onto wall tangent (perpendicular to normal)
	// Tangent is (-normal.NY, normal.NX) for 2D
	tangentX := -normal.NY
	tangentY := normal.NX

	// Dot product with tangent
	dotProduct := vx*tangentX + vy*tangentY

	// Project velocity onto tangent
	newVX = dotProduct * tangentX
	newVY = dotProduct * tangentY

	if debugEnabled {
		collisionLog.WithFields(logrus.Fields{
			"new_velocity_x": newVX,
			"new_velocity_y": newVY,
			"tangent_x":      tangentX,
			"tangent_y":      tangentY,
			"dot_product":    dotProduct,
		}).Debug("Wall slide velocity calculated")
	}

	return newVX, newVY
}

// ResolveWallCollision resolves collision with terrain wall using smooth sliding.
// Updates entity position to slide along wall instead of stopping.
// Returns the adjusted position.
func ResolveWallCollision(entity *Entity, normal EdgeNormal) (adjustedX, adjustedY float64) {
	collisionLog.WithFields(logrus.Fields{
		"operation": "resolve_wall_collision",
		"entity_id": entity.ID,
		"normal_x":  normal.NX,
		"normal_y":  normal.NY,
	}).Debug("Resolving wall collision for entity")

	posComp, hasPoscomp := entity.GetComponent("position")
	if !hasPoscomp {
		collisionLog.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Warn("Entity missing position component, cannot resolve wall collision")
		return 0, 0
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		collisionLog.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "position",
		}).Error("Failed to type assert position component")
		return 0, 0
	}

	originalX := pos.X
	originalY := pos.Y

	// Get velocity if present
	velComp, hasVel := entity.GetComponent("velocity")
	if hasVel {
		if vel, ok := velComp.(*VelocityComponent); ok {
			oldVX := vel.VX
			oldVY := vel.VY
			// Apply wall slide to velocity
			vel.VX, vel.VY = ApplyWallSlide(vel.VX, vel.VY, normal)
			collisionLog.WithFields(logrus.Fields{
				"entity_id":      entity.ID,
				"old_velocity_x": oldVX,
				"old_velocity_y": oldVY,
				"new_velocity_x": vel.VX,
				"new_velocity_y": vel.VY,
			}).Debug("Applied wall slide to entity velocity")
		}
	}

	// Push entity slightly away from wall along normal
	pushDistance := CollisionPrecision * 2.0 // Push 0.2 pixels away
	adjustedX = pos.X + normal.NX*pushDistance
	adjustedY = pos.Y + normal.NY*pushDistance

	// Quantize to maintain precision
	adjustedX, adjustedY = QuantizePosition(adjustedX, adjustedY)

	collisionLog.WithFields(logrus.Fields{
		"entity_id":     entity.ID,
		"original_x":    originalX,
		"original_y":    originalY,
		"adjusted_x":    adjustedX,
		"adjusted_y":    adjustedY,
		"push_distance": pushDistance,
	}).Debug("Wall collision resolved with position adjustment")

	return adjustedX, adjustedY
}

// CheckPreciseCollision performs pixel-perfect collision check between two entities.
// Returns true if entities collide at sub-pixel precision (0.1px).
func CheckPreciseCollision(e1, e2 *Entity) bool {
	collisionLog.WithFields(logrus.Fields{
		"operation":  "check_precise_collision",
		"entity1_id": e1.ID,
		"entity2_id": e2.ID,
	}).Debug("Checking precise collision between entities")

	if result, ok := checkPreciseColliders(e1, e2); ok {
		return result
	}

	return checkStandardColliders(e1, e2)
}

// checkPreciseColliders tests collision using precise collider components.
func checkPreciseColliders(e1, e2 *Entity) (result, ok bool) {
	pc1Comp, has1 := e1.GetComponent("precise_collider")
	pc2Comp, has2 := e2.GetComponent("precise_collider")

	if !has1 || !has2 {
		return false, false
	}

	collisionLog.WithFields(logrus.Fields{
		"entity1_id": e1.ID,
		"entity2_id": e2.ID,
	}).Debug("Both entities have precise colliders, using precise collision")

	pc1, ok1 := pc1Comp.(*PreciseColliderComponent)
	pc2, ok2 := pc2Comp.(*PreciseColliderComponent)

	if !ok1 || !ok2 {
		collisionLog.WithFields(logrus.Fields{
			"entity1_id":   e1.ID,
			"entity2_id":   e2.ID,
			"type_assert1": ok1,
			"type_assert2": ok2,
		}).Warn("Failed to type assert precise collider components")
		return false, false
	}

	pos1Comp, hasPos1 := e1.GetComponent("position")
	pos2Comp, hasPos2 := e2.GetComponent("position")

	if !hasPos1 || !hasPos2 {
		collisionLog.WithFields(logrus.Fields{
			"entity1_id": e1.ID,
			"entity2_id": e2.ID,
			"has_pos1":   hasPos1,
			"has_pos2":   hasPos2,
		}).Warn("Entities missing position components for precise collision")
		return false, false
	}

	p1, ok1 := pos1Comp.(*PositionComponent)
	p2, ok2 := pos2Comp.(*PositionComponent)

	if !ok1 || !ok2 {
		return false, false
	}

	result = pc1.Intersects(p1.X, p1.Y, pc2, p2.X, p2.Y)
	collisionLog.WithFields(logrus.Fields{
		"entity1_id": e1.ID,
		"entity2_id": e2.ID,
		"collides":   result,
	}).Debug("Precise collision check completed")
	return result, true
}

// checkStandardColliders tests collision using standard collider components.
func checkStandardColliders(e1, e2 *Entity) bool {
	collisionLog.WithFields(logrus.Fields{
		"entity1_id": e1.ID,
		"entity2_id": e2.ID,
	}).Debug("Falling back to standard collider")

	c1Comp, has1 := e1.GetComponent("collider")
	c2Comp, has2 := e2.GetComponent("collider")

	if !has1 || !has2 {
		collisionLog.WithFields(logrus.Fields{
			"entity1_id": e1.ID,
			"entity2_id": e2.ID,
			"result":     false,
		}).Debug("No valid colliders found, returning false")
		return false
	}

	c1, ok1 := c1Comp.(*ColliderComponent)
	c2, ok2 := c2Comp.(*ColliderComponent)

	if !ok1 || !ok2 {
		collisionLog.WithFields(logrus.Fields{
			"entity1_id":   e1.ID,
			"entity2_id":   e2.ID,
			"type_assert1": ok1,
			"type_assert2": ok2,
		}).Warn("Failed to type assert standard collider components")
		return false
	}

	pos1Comp, hasPos1 := e1.GetComponent("position")
	pos2Comp, hasPos2 := e2.GetComponent("position")

	if !hasPos1 || !hasPos2 {
		collisionLog.WithFields(logrus.Fields{
			"entity1_id": e1.ID,
			"entity2_id": e2.ID,
			"has_pos1":   hasPos1,
			"has_pos2":   hasPos2,
		}).Warn("Entities missing position components for standard collision")
		return false
	}

	p1, ok1 := pos1Comp.(*PositionComponent)
	p2, ok2 := pos2Comp.(*PositionComponent)

	if !ok1 || !ok2 {
		return false
	}

	result := c1.Intersects(p1.X, p1.Y, c2, p2.X, p2.Y)
	collisionLog.WithFields(logrus.Fields{
		"entity1_id": e1.ID,
		"entity2_id": e2.ID,
		"collides":   result,
	}).Debug("Standard collision check completed")
	return result
}

// GetCollisionAlignment calculates the visual/collision alignment error.
// Returns the distance between visual center and collision center.
// Used for debug visualization to ensure <0.5px alignment.
func GetCollisionAlignment(entity *Entity) (alignmentError float64) {
	collisionLog.WithFields(logrus.Fields{
		"operation": "get_collision_alignment",
		"entity_id": entity.ID,
	}).Debug("Calculating collision alignment error")

	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		collisionLog.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Warn("Entity missing position component, cannot calculate alignment")
		return 0
	}
	pos, ok := posComp.(*PositionComponent)
	if !ok {
		collisionLog.WithFields(logrus.Fields{
			"entity_id":      entity.ID,
			"component_type": "position",
		}).Error("Failed to type assert position component")
		return 0
	}

	// Get collision bounds
	var collisionCenterX, collisionCenterY float64
	var hasCollider bool

	if pcComp, hasPrecise := entity.GetComponent("precise_collider"); hasPrecise {
		if pc, ok := pcComp.(*PreciseColliderComponent); ok {
			minX, minY, maxX, maxY := pc.GetBounds(pos.X, pos.Y)
			collisionCenterX = (minX + maxX) / 2.0
			collisionCenterY = (minY + maxY) / 2.0
			hasCollider = true
			collisionLog.WithFields(logrus.Fields{
				"entity_id":          entity.ID,
				"collider_type":      "precise",
				"collision_center_x": collisionCenterX,
				"collision_center_y": collisionCenterY,
			}).Debug("Using precise collider bounds")
		}
	} else if cComp, hasCollider := entity.GetComponent("collider"); hasCollider {
		if c, ok := cComp.(*ColliderComponent); ok {
			minX, minY, maxX, maxY := c.GetBounds(pos.X, pos.Y)
			collisionCenterX = (minX + maxX) / 2.0
			collisionCenterY = (minY + maxY) / 2.0
			hasCollider = true
			collisionLog.WithFields(logrus.Fields{
				"entity_id":          entity.ID,
				"collider_type":      "standard",
				"collision_center_x": collisionCenterX,
				"collision_center_y": collisionCenterY,
			}).Debug("Using standard collider bounds")
		}
	}

	if !hasCollider {
		collisionLog.WithFields(logrus.Fields{
			"entity_id": entity.ID,
		}).Debug("Entity has no collider, alignment error is 0")
		return 0
	}

	// Visual center is entity position (sprite is centered)
	visualCenterX := pos.X
	visualCenterY := pos.Y

	// Calculate distance
	dx := collisionCenterX - visualCenterX
	dy := collisionCenterY - visualCenterY
	alignmentError = math.Sqrt(dx*dx + dy*dy)

	collisionLog.WithFields(logrus.Fields{
		"entity_id":          entity.ID,
		"visual_center_x":    visualCenterX,
		"visual_center_y":    visualCenterY,
		"collision_center_x": collisionCenterX,
		"collision_center_y": collisionCenterY,
		"alignment_error":    alignmentError,
		"within_threshold":   alignmentError < 0.5,
	}).Debug("Collision alignment calculated")

	return alignmentError
}
