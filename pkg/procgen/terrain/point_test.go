//go:build test
// +build test

package terrain

import (
	"math"
	"testing"
)

// TestNewPoint tests point creation.
func TestNewPoint(t *testing.T) {
	p := NewPoint(5, 10)
	if p.X != 5 || p.Y != 10 {
		t.Errorf("NewPoint(5, 10) = (%d, %d), want (5, 10)", p.X, p.Y)
	}
}

// TestPoint_Distance tests Euclidean distance calculation.
func TestPoint_Distance(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Point
		expected float64
	}{
		{"Same point", Point{0, 0}, Point{0, 0}, 0.0},
		{"Horizontal distance", Point{0, 0}, Point{3, 0}, 3.0},
		{"Vertical distance", Point{0, 0}, Point{0, 4}, 4.0},
		{"Diagonal distance", Point{0, 0}, Point{3, 4}, 5.0},
		{"Negative coordinates", Point{-1, -1}, Point{2, 3}, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.Distance(tt.p2)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("Distance(%v, %v) = %v, want %v", tt.p1, tt.p2, result, tt.expected)
			}
		})
	}
}

// TestPoint_ManhattanDistance tests Manhattan distance calculation.
func TestPoint_ManhattanDistance(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Point
		expected int
	}{
		{"Same point", Point{0, 0}, Point{0, 0}, 0},
		{"Horizontal distance", Point{0, 0}, Point{3, 0}, 3},
		{"Vertical distance", Point{0, 0}, Point{0, 4}, 4},
		{"Diagonal distance", Point{0, 0}, Point{3, 4}, 7},
		{"Negative coordinates", Point{-1, -1}, Point{2, 3}, 7},
		{"Both negative", Point{-5, -5}, Point{-2, -3}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.ManhattanDistance(tt.p2)
			if result != tt.expected {
				t.Errorf("ManhattanDistance(%v, %v) = %v, want %v", tt.p1, tt.p2, result, tt.expected)
			}
		})
	}
}

// TestPoint_Equals tests point equality.
func TestPoint_Equals(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   Point
		expected bool
	}{
		{"Same coordinates", Point{5, 10}, Point{5, 10}, true},
		{"Different X", Point{5, 10}, Point{6, 10}, false},
		{"Different Y", Point{5, 10}, Point{5, 11}, false},
		{"Both different", Point{5, 10}, Point{6, 11}, false},
		{"Zero points", Point{0, 0}, Point{0, 0}, true},
		{"Negative points", Point{-1, -2}, Point{-1, -2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.Equals(tt.p2)
			if result != tt.expected {
				t.Errorf("Equals(%v, %v) = %v, want %v", tt.p1, tt.p2, result, tt.expected)
			}
		})
	}
}

// TestPoint_IsInBounds tests bounds checking.
func TestPoint_IsInBounds(t *testing.T) {
	tests := []struct {
		name          string
		point         Point
		width, height int
		expected      bool
	}{
		{"Origin in 10x10", Point{0, 0}, 10, 10, true},
		{"Center in 10x10", Point{5, 5}, 10, 10, true},
		{"Max corner in 10x10", Point{9, 9}, 10, 10, true},
		{"X too large", Point{10, 5}, 10, 10, false},
		{"Y too large", Point{5, 10}, 10, 10, false},
		{"Negative X", Point{-1, 5}, 10, 10, false},
		{"Negative Y", Point{5, -1}, 10, 10, false},
		{"Both negative", Point{-1, -1}, 10, 10, false},
		{"Point in 1x1", Point{0, 0}, 1, 1, true},
		{"Point outside 1x1", Point{1, 0}, 1, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.point.IsInBounds(tt.width, tt.height)
			if result != tt.expected {
				t.Errorf("IsInBounds(%v, %d, %d) = %v, want %v",
					tt.point, tt.width, tt.height, result, tt.expected)
			}
		})
	}
}

// TestPoint_Neighbors tests orthogonal neighbor generation.
func TestPoint_Neighbors(t *testing.T) {
	p := Point{5, 5}
	neighbors := p.Neighbors()

	// Should have exactly 4 neighbors
	if len(neighbors) != 4 {
		t.Errorf("Neighbors() returned %d neighbors, want 4", len(neighbors))
	}

	// Check expected neighbors (order matters for this implementation)
	expected := []Point{
		{5, 4}, // North
		{6, 5}, // East
		{5, 6}, // South
		{4, 5}, // West
	}

	for i, exp := range expected {
		if !neighbors[i].Equals(exp) {
			t.Errorf("Neighbor %d = %v, want %v", i, neighbors[i], exp)
		}
	}
}

// TestPoint_AllNeighbors tests all 8 neighbor generation.
func TestPoint_AllNeighbors(t *testing.T) {
	p := Point{5, 5}
	neighbors := p.AllNeighbors()

	// Should have exactly 8 neighbors
	if len(neighbors) != 8 {
		t.Errorf("AllNeighbors() returned %d neighbors, want 8", len(neighbors))
	}

	// Check that all expected neighbors are present
	expected := []Point{
		{5, 4}, // North
		{6, 4}, // Northeast
		{6, 5}, // East
		{6, 6}, // Southeast
		{5, 6}, // South
		{4, 6}, // Southwest
		{4, 5}, // West
		{4, 4}, // Northwest
	}

	for i, exp := range expected {
		if !neighbors[i].Equals(exp) {
			t.Errorf("Neighbor %d = %v, want %v", i, neighbors[i], exp)
		}
	}
}

// TestPoint_NeighborsAtBounds tests neighbors at map boundaries.
func TestPoint_NeighborsAtBounds(t *testing.T) {
	// Test corner point
	p := Point{0, 0}
	neighbors := p.Neighbors()

	// Even at corner, should return 4 neighbors (some out of bounds)
	if len(neighbors) != 4 {
		t.Errorf("Neighbors at corner returned %d neighbors, want 4", len(neighbors))
	}

	// Check that some neighbors are out of bounds for a 10x10 map
	width, height := 10, 10
	inBoundsCount := 0
	for _, n := range neighbors {
		if n.IsInBounds(width, height) {
			inBoundsCount++
		}
	}

	// At corner (0,0), only 2 neighbors should be in bounds: East and South
	if inBoundsCount != 2 {
		t.Errorf("Corner point has %d in-bounds neighbors, want 2", inBoundsCount)
	}
}
