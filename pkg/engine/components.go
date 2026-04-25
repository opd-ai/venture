// Package engine provides basic physics components for the ECS.
// This file defines fundamental components: PositionComponent, VelocityComponent,
// ColliderComponent, and BoundsComponent used across all game systems.
package engine

import (
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
)

var componentsLog *logrus.Logger

func init() {
	componentsLog = logrus.New()
	componentsLog.SetReportCaller(true)
	componentsLog.SetLevel(logrus.InfoLevel)
}

// PositionComponent represents an entity's position in 2D space.
// PrevX/PrevY store the position from the previous simulation tick,
// enabling smooth interpolation in the render path between Update() and Draw().
// Initialized is set to true by MovementSystem after the first physics tick,
// distinguishing a genuine (0,0) spawn from an uninitialized component.
type PositionComponent struct {
	X, Y         float64
	PrevX, PrevY float64 // Previous tick position for render interpolation
	Initialized  bool    // G38: true once MovementSystem has run at least one tick
}

// Type returns the component type identifier.
func (p *PositionComponent) Type() string {
	return "position"
}

// Serialize encodes the component to bytes for persistence
func (p *PositionComponent) Serialize() ([]byte, error) {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "position",
		"x":              p.X,
		"y":              p.Y,
	}).Debug("Serializing position component")

	buf := make([]byte, 16) // 2 float64s = 16 bytes
	writeFloat64(buf[0:8], p.X)
	writeFloat64(buf[8:16], p.Y)

	componentsLog.WithFields(logrus.Fields{
		"component_type": "position",
		"bytes":          len(buf),
	}).Debug("Position component serialized successfully")

	return buf, nil
}

// Deserialize decodes the component from bytes
func (p *PositionComponent) Deserialize(data []byte) error {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "position",
		"bytes":          len(data),
	}).Debug("Deserializing position component")

	if len(data) < 16 {
		componentsLog.WithFields(logrus.Fields{
			"component_type": "position",
			"expected_bytes": 16,
			"received_bytes": len(data),
		}).Error("Insufficient data for position component deserialization")
		return ErrInvalidComponentData
	}
	p.X = readFloat64(data[0:8])
	p.Y = readFloat64(data[8:16])

	componentsLog.WithFields(logrus.Fields{
		"component_type": "position",
		"x":              p.X,
		"y":              p.Y,
	}).Debug("Position component deserialized successfully")

	return nil
}

// String returns a human-readable representation of the position.
func (p *PositionComponent) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", p.X, p.Y)
}

// NameComponent stores an entity's display name.
// Used for UI display, logging, and elemental affinity inference from name keywords.
type NameComponent struct {
	Name string
}

// Type returns the component type identifier.
func (n *NameComponent) Type() string {
	return "name"
}

// VelocityComponent represents an entity's velocity in 2D space.
type VelocityComponent struct {
	VX, VY float64
}

// Type returns the component type identifier.
func (v *VelocityComponent) Type() string {
	return "velocity"
}

// Serialize encodes the component to bytes for persistence
func (v *VelocityComponent) Serialize() ([]byte, error) {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "velocity",
		"vx":             v.VX,
		"vy":             v.VY,
	}).Debug("Serializing velocity component")

	buf := make([]byte, 16) // 2 float64s = 16 bytes
	writeFloat64(buf[0:8], v.VX)
	writeFloat64(buf[8:16], v.VY)

	componentsLog.WithFields(logrus.Fields{
		"component_type": "velocity",
		"bytes":          len(buf),
	}).Debug("Velocity component serialized successfully")

	return buf, nil
}

// Deserialize decodes the component from bytes
func (v *VelocityComponent) Deserialize(data []byte) error {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "velocity",
		"bytes":          len(data),
	}).Debug("Deserializing velocity component")

	if len(data) < 16 {
		componentsLog.WithFields(logrus.Fields{
			"component_type": "velocity",
			"expected_bytes": 16,
			"received_bytes": len(data),
		}).Error("Insufficient data for velocity component deserialization")
		return ErrInvalidComponentData
	}
	v.VX = readFloat64(data[0:8])
	v.VY = readFloat64(data[8:16])

	componentsLog.WithFields(logrus.Fields{
		"component_type": "velocity",
		"vx":             v.VX,
		"vy":             v.VY,
	}).Debug("Velocity component deserialized successfully")

	return nil
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

// Serialize encodes the component to bytes for persistence
func (c *ColliderComponent) Serialize() ([]byte, error) {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"width":          c.Width,
		"height":         c.Height,
		"solid":          c.Solid,
		"is_trigger":     c.IsTrigger,
		"layer":          c.Layer,
		"offset_x":       c.OffsetX,
		"offset_y":       c.OffsetY,
	}).Debug("Serializing collider component")

	buf := make([]byte, 38) // 4 float64s (32 bytes) + 2 bools (2 bytes) + 1 int32 (4 bytes) = 38 bytes
	writeFloat64(buf[0:8], c.Width)
	writeFloat64(buf[8:16], c.Height)
	writeBool(buf[16:17], c.Solid)
	writeBool(buf[17:18], c.IsTrigger)
	writeInt32(buf[18:22], int32(c.Layer))
	writeFloat64(buf[22:30], c.OffsetX)
	writeFloat64(buf[30:38], c.OffsetY)

	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"bytes":          len(buf),
	}).Debug("Collider component serialized successfully")

	return buf, nil
}

// Deserialize decodes the component from bytes
func (c *ColliderComponent) Deserialize(data []byte) error {
	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"bytes":          len(data),
	}).Debug("Deserializing collider component")

	if len(data) < 38 {
		componentsLog.WithFields(logrus.Fields{
			"component_type": "collider",
			"expected_bytes": 38,
			"received_bytes": len(data),
		}).Error("Insufficient data for collider component deserialization")
		return ErrInvalidComponentData
	}
	c.Width = readFloat64(data[0:8])
	c.Height = readFloat64(data[8:16])
	c.Solid = readBool(data[16:17])
	c.IsTrigger = readBool(data[17:18])
	c.Layer = int(readInt32(data[18:22]))
	c.OffsetX = readFloat64(data[22:30])
	c.OffsetY = readFloat64(data[30:38])

	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"width":          c.Width,
		"height":         c.Height,
		"solid":          c.Solid,
		"is_trigger":     c.IsTrigger,
		"layer":          c.Layer,
		"offset_x":       c.OffsetX,
		"offset_y":       c.OffsetY,
	}).Debug("Collider component deserialized successfully")

	return nil
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
	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"x":              x,
		"y":              y,
		"angle":          angle,
		"width":          c.Width,
		"height":         c.Height,
	}).Debug("Calculating rotated collider bounds")

	// If no rotation, use regular bounds for efficiency
	if angle == 0 {
		componentsLog.WithFields(logrus.Fields{
			"component_type": "collider",
			"optimization":   "no_rotation",
		}).Debug("Using non-rotated bounds optimization")
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

	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"min_x":          minX,
		"min_y":          minY,
		"max_x":          maxX,
		"max_y":          maxY,
		"rotation":       angle,
	}).Debug("Rotated collider bounds calculated")

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
	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"operation":      "intersects_rotated",
		"x1":             x1,
		"y1":             y1,
		"angle1":         angle1,
		"x2":             x2,
		"y2":             y2,
		"angle2":         angle2,
	}).Debug("Checking rotated collider intersection")

	minX1, minY1, maxX1, maxY1 := c.GetRotatedBounds(x1, y1, angle1)
	minX2, minY2, maxX2, maxY2 := other.GetRotatedBounds(x2, y2, angle2)

	intersects := !(maxX1 <= minX2 || maxX2 <= minX1 || maxY1 <= minY2 || maxY2 <= minY1)

	componentsLog.WithFields(logrus.Fields{
		"component_type": "collider",
		"operation":      "intersects_rotated",
		"result":         intersects,
	}).Debug("Rotated collider intersection check completed")

	return intersects
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
	componentsLog.WithFields(logrus.Fields{
		"component_type": "bounds",
		"operation":      "clamp",
		"input_x":        x,
		"input_y":        y,
		"wrap":           b.Wrap,
		"min_x":          b.MinX,
		"min_y":          b.MinY,
		"max_x":          b.MaxX,
		"max_y":          b.MaxY,
	}).Debug("Clamping position to bounds")

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

		componentsLog.WithFields(logrus.Fields{
			"component_type": "bounds",
			"operation":      "clamp",
			"mode":           "wrap",
			"output_x":       x,
			"output_y":       y,
		}).Debug("Position wrapped to bounds")
	} else {
		// Clamp to bounds
		x = math.Max(b.MinX, math.Min(b.MaxX, x))
		y = math.Max(b.MinY, math.Min(b.MaxY, y))

		componentsLog.WithFields(logrus.Fields{
			"component_type": "bounds",
			"operation":      "clamp",
			"mode":           "clamp",
			"output_x":       x,
			"output_y":       y,
		}).Debug("Position clamped to bounds")
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
	componentsLog.WithFields(logrus.Fields{
		"component_type": "friction",
		"coefficient":    coefficient,
	}).Debug("Creating friction component")

	return &FrictionComponent{
		Coefficient: coefficient,
	}
}
