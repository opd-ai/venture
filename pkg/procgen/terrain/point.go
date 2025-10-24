// Package terrain provides point utilities for terrain generation.
// This file defines the Point type and related utility functions.
package terrain

import "math"

// Point represents a 2D coordinate in the terrain grid.
type Point struct {
	X, Y int
}

// NewPoint creates a new Point with the given coordinates.
func NewPoint(x, y int) Point {
	return Point{X: x, Y: y}
}

// Distance calculates the Euclidean distance between two points.
func (p Point) Distance(other Point) float64 {
	dx := float64(p.X - other.X)
	dy := float64(p.Y - other.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// ManhattanDistance calculates the Manhattan (taxicab) distance between two points.
func (p Point) ManhattanDistance(other Point) int {
	dx := p.X - other.X
	if dx < 0 {
		dx = -dx
	}
	dy := p.Y - other.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Equals checks if two points have the same coordinates.
func (p Point) Equals(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

// IsInBounds checks if the point is within the given width and height.
func (p Point) IsInBounds(width, height int) bool {
	return p.X >= 0 && p.X < width && p.Y >= 0 && p.Y < height
}

// Neighbors returns the four orthogonal neighbors of this point.
func (p Point) Neighbors() []Point {
	return []Point{
		{X: p.X, Y: p.Y - 1},     // North
		{X: p.X + 1, Y: p.Y},     // East
		{X: p.X, Y: p.Y + 1},     // South
		{X: p.X - 1, Y: p.Y},     // West
	}
}

// AllNeighbors returns all eight neighbors (orthogonal and diagonal).
func (p Point) AllNeighbors() []Point {
	return []Point{
		{X: p.X, Y: p.Y - 1},     // North
		{X: p.X + 1, Y: p.Y - 1}, // Northeast
		{X: p.X + 1, Y: p.Y},     // East
		{X: p.X + 1, Y: p.Y + 1}, // Southeast
		{X: p.X, Y: p.Y + 1},     // South
		{X: p.X - 1, Y: p.Y + 1}, // Southwest
		{X: p.X - 1, Y: p.Y},     // West
		{X: p.X - 1, Y: p.Y - 1}, // Northwest
	}
}
