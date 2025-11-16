package animation

import "math"

// Direction8 represents 8-directional movement for animation.
// Extends the 4-direction system with diagonal directions.
type Direction8 int

const (
	// Primary directions
	Dir8North Direction8 = iota
	Dir8NorthEast
	Dir8East
	Dir8SouthEast
	Dir8South
	Dir8SouthWest
	Dir8West
	Dir8NorthWest
)

// String returns the string representation of the direction.
func (d Direction8) String() string {
	switch d {
	case Dir8North:
		return "north"
	case Dir8NorthEast:
		return "northeast"
	case Dir8East:
		return "east"
	case Dir8SouthEast:
		return "southeast"
	case Dir8South:
		return "south"
	case Dir8SouthWest:
		return "southwest"
	case Dir8West:
		return "west"
	case Dir8NorthWest:
		return "northwest"
	default:
		return "unknown"
	}
}

// Angle returns the angle in radians for the direction (0 = East, PI/2 = North).
func (d Direction8) Angle() float64 {
	switch d {
	case Dir8East:
		return 0
	case Dir8NorthEast:
		return math.Pi / 4
	case Dir8North:
		return math.Pi / 2
	case Dir8NorthWest:
		return 3 * math.Pi / 4
	case Dir8West:
		return math.Pi
	case Dir8SouthWest:
		return 5 * math.Pi / 4
	case Dir8South:
		return 3 * math.Pi / 2
	case Dir8SouthEast:
		return 7 * math.Pi / 4
	default:
		return 0
	}
}

// IsDiagonal returns true if the direction is diagonal (NE, SE, SW, NW).
func (d Direction8) IsDiagonal() bool {
	return d == Dir8NorthEast || d == Dir8SouthEast ||
		d == Dir8SouthWest || d == Dir8NorthWest
}

// FromVelocity determines the 8-direction from velocity components.
// Uses 8-way directional segmentation for accurate diagonal detection.
func FromVelocity(vx, vy float64) Direction8 {
	// Threshold for considering movement significant
	const threshold = 0.01
	if math.Abs(vx) < threshold && math.Abs(vy) < threshold {
		return Dir8South // Default to south (facing camera)
	}

	// Calculate angle from velocity vector
	angle := math.Atan2(-vy, vx) // Negative Y because screen coordinates
	if angle < 0 {
		angle += 2 * math.Pi
	}

	// Determine direction from angle (8 segments of 45 degrees each)
	// East = 0, rotating counter-clockwise
	segment := int(math.Round(angle / (math.Pi / 4)))
	segment = segment % 8

	switch segment {
	case 0:
		return Dir8East
	case 1:
		return Dir8NorthEast
	case 2:
		return Dir8North
	case 3:
		return Dir8NorthWest
	case 4:
		return Dir8West
	case 5:
		return Dir8SouthWest
	case 6:
		return Dir8South
	case 7:
		return Dir8SouthEast
	default:
		return Dir8South
	}
}

// To4Direction converts 8-direction to legacy 4-direction for compatibility.
// Diagonal directions are mapped to their primary components.
func (d Direction8) To4Direction() string {
	switch d {
	case Dir8North, Dir8NorthEast, Dir8NorthWest:
		return "up"
	case Dir8South, Dir8SouthEast, Dir8SouthWest:
		return "down"
	case Dir8East:
		return "right"
	case Dir8West:
		return "left"
	default:
		return "down"
	}
}

// Opposite returns the opposite direction.
func (d Direction8) Opposite() Direction8 {
	return (d + 4) % 8
}

// RotateClockwise rotates the direction 45 degrees clockwise.
func (d Direction8) RotateClockwise() Direction8 {
	return (d + 1) % 8
}

// RotateCounterClockwise rotates the direction 45 degrees counter-clockwise.
func (d Direction8) RotateCounterClockwise() Direction8 {
	return (d + 7) % 8
}
