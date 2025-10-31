// Package engine provides diagonal wall collision detection tests.
// Phase 11.1 Week 3: Diagonal Wall Collision System Tests
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestDiagonalWallCollision tests collision detection with all four diagonal wall orientations.
func TestDiagonalWallCollision(t *testing.T) {
	tests := []struct {
		name         string
		wallType     terrain.TileType
		entityBounds [4]float64 // minX, minY, maxX, maxY
		expectHit    bool
		description  string
	}{
		// TileWallNE: / diagonal (bottom-left to top-right)
		{
			name:         "NE diagonal - center hit",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{12, 12, 20, 20}, // Center of tile
			expectHit:    true,
			description:  "Entity in center of NE diagonal should collide",
		},
		{
			name:         "NE diagonal - top-left miss",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{2, 2, 8, 8}, // Top-left corner (open space)
			expectHit:    false,
			description:  "Entity in top-left of NE diagonal should NOT collide",
		},
		{
			name:         "NE diagonal - bottom-right corner hit",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{24, 24, 28, 28}, // Bottom-right corner (solid)
			expectHit:    true,
			description:  "Entity in bottom-right of NE diagonal should collide",
		},

		// TileWallNW: \ diagonal (bottom-right to top-left)
		{
			name:         "NW diagonal - center hit",
			wallType:     terrain.TileWallNW,
			entityBounds: [4]float64{12, 12, 20, 20},
			expectHit:    true,
			description:  "Entity in center of NW diagonal should collide",
		},
		{
			name:         "NW diagonal - top-right miss",
			wallType:     terrain.TileWallNW,
			entityBounds: [4]float64{24, 2, 28, 8}, // Top-right corner (open space)
			expectHit:    false,
			description:  "Entity in top-right of NW diagonal should NOT collide",
		},
		{
			name:         "NW diagonal - bottom-left corner hit",
			wallType:     terrain.TileWallNW,
			entityBounds: [4]float64{2, 24, 8, 28}, // Bottom-left corner (solid)
			expectHit:    true,
			description:  "Entity in bottom-left of NW diagonal should collide",
		},

		// TileWallSE: \ diagonal (top-left to bottom-right)
		{
			name:         "SE diagonal - center hit",
			wallType:     terrain.TileWallSE,
			entityBounds: [4]float64{12, 12, 20, 20},
			expectHit:    true,
			description:  "Entity in center of SE diagonal should collide",
		},
		{
			name:         "SE diagonal - bottom-left miss",
			wallType:     terrain.TileWallSE,
			entityBounds: [4]float64{2, 24, 8, 28}, // Bottom-left corner (open space)
			expectHit:    false,
			description:  "Entity in bottom-left of SE diagonal should NOT collide",
		},
		{
			name:         "SE diagonal - top-right corner hit",
			wallType:     terrain.TileWallSE,
			entityBounds: [4]float64{24, 2, 28, 8}, // Top-right corner (solid)
			expectHit:    true,
			description:  "Entity in top-right of SE diagonal should collide",
		},

		// TileWallSW: / diagonal (top-right to bottom-left)
		{
			name:         "SW diagonal - center hit",
			wallType:     terrain.TileWallSW,
			entityBounds: [4]float64{12, 12, 20, 20},
			expectHit:    true,
			description:  "Entity in center of SW diagonal should collide",
		},
		{
			name:         "SW diagonal - bottom-right miss",
			wallType:     terrain.TileWallSW,
			entityBounds: [4]float64{24, 24, 28, 28}, // Bottom-right corner (open space)
			expectHit:    false,
			description:  "Entity in bottom-right of SW diagonal should NOT collide",
		},
		{
			name:         "SW diagonal - top-left corner hit",
			wallType:     terrain.TileWallSW,
			entityBounds: [4]float64{2, 2, 8, 8}, // Top-left corner (solid)
			expectHit:    true,
			description:  "Entity in top-left of SW diagonal should collide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create terrain with single diagonal wall tile at (0, 0)
			terr := &terrain.Terrain{
				Width:  1,
				Height: 1,
				Tiles:  [][]terrain.TileType{{tt.wallType}},
			}

			// Create terrain collision checker
			checker := NewTerrainCollisionChecker(32, 32) // 32x32 pixel tiles
			checker.SetTerrain(terr)

			// Test collision
			hit := checker.CheckCollisionBounds(
				tt.entityBounds[0],
				tt.entityBounds[1],
				tt.entityBounds[2],
				tt.entityBounds[3],
			)

			if hit != tt.expectHit {
				t.Errorf("%s: expected hit=%v, got hit=%v\nDescription: %s",
					tt.name, tt.expectHit, hit, tt.description)
			}
		})
	}
}

// TestDiagonalWallEdgeCases tests edge cases and boundary conditions.
func TestDiagonalWallEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		wallType     terrain.TileType
		entityBounds [4]float64
		expectHit    bool
	}{
		{
			name:         "Zero-size entity",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{16, 16, 16, 16},
			expectHit:    true, // Point in triangle
		},
		{
			name:         "Entity exactly on diagonal edge - NE",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{15, 15, 17, 17}, // Straddles diagonal
			expectHit:    true,
		},
		{
			name:         "Entity exactly on tile boundary - top edge",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{0, 0, 4, 4},
			expectHit:    false, // Top-left corner is open for NE
		},
		{
			name:         "Entity larger than tile",
			wallType:     terrain.TileWallNE,
			entityBounds: [4]float64{-10, -10, 42, 42},
			expectHit:    true, // Definitely overlaps
		},
		{
			name:         "Entity partially overlapping - NW diagonal",
			wallType:     terrain.TileWallNW,
			entityBounds: [4]float64{14, 14, 18, 18}, // Small entity near center
			expectHit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			terr := &terrain.Terrain{
				Width:  1,
				Height: 1,
				Tiles:  [][]terrain.TileType{{tt.wallType}},
			}

			checker := NewTerrainCollisionChecker(32, 32)
			checker.SetTerrain(terr)

			hit := checker.CheckCollisionBounds(
				tt.entityBounds[0],
				tt.entityBounds[1],
				tt.entityBounds[2],
				tt.entityBounds[3],
			)

			if hit != tt.expectHit {
				t.Errorf("%s: expected hit=%v, got hit=%v", tt.name, tt.expectHit, hit)
			}
		})
	}
}

// TestTriangleAABBIntersection tests the low-level triangle-AABB intersection algorithm.
func TestTriangleAABBIntersection(t *testing.T) {
	tests := []struct {
		name     string
		triangle [6]float64 // v1X, v1Y, v2X, v2Y, v3X, v3Y
		aabb     [4]float64 // minX, minY, maxX, maxY
		expected bool
	}{
		{
			name:     "Triangle contains AABB",
			triangle: [6]float64{0, 0, 100, 0, 50, 100},
			aabb:     [4]float64{40, 10, 60, 30},
			expected: true,
		},
		{
			name:     "AABB contains triangle",
			triangle: [6]float64{10, 10, 20, 10, 15, 20},
			aabb:     [4]float64{0, 0, 100, 100},
			expected: true,
		},
		{
			name:     "Partial overlap - vertex inside",
			triangle: [6]float64{0, 0, 30, 0, 15, 30},
			aabb:     [4]float64{10, 10, 50, 50},
			expected: true,
		},
		{
			name:     "Partial overlap - edge intersection",
			triangle: [6]float64{0, 10, 30, 10, 15, 40},
			aabb:     [4]float64{10, 0, 20, 50},
			expected: true,
		},
		{
			name:     "No overlap - separated",
			triangle: [6]float64{0, 0, 10, 0, 5, 10},
			aabb:     [4]float64{20, 20, 30, 30},
			expected: false,
		},
		{
			name:     "No overlap - adjacent",
			triangle: [6]float64{0, 0, 10, 0, 5, 10},
			aabb:     [4]float64{10, 0, 20, 10},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := triangleAABBIntersection(
				tt.triangle[0], tt.triangle[1],
				tt.triangle[2], tt.triangle[3],
				tt.triangle[4], tt.triangle[5],
				tt.aabb[0], tt.aabb[1], tt.aabb[2], tt.aabb[3],
			)

			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

// TestPointInTriangle tests the point-in-triangle algorithm.
func TestPointInTriangle(t *testing.T) {
	// Equilateral triangle
	triangle := [6]float64{0, 0, 20, 0, 10, 17.32}

	tests := []struct {
		name     string
		point    [2]float64
		expected bool
	}{
		{"Center point", [2]float64{10, 8}, true},
		{"Vertex 1", [2]float64{0, 0}, true},
		{"Vertex 2", [2]float64{20, 0}, true},
		{"Vertex 3", [2]float64{10, 17.32}, true},
		{"Outside top", [2]float64{10, 20}, false},
		{"Outside bottom", [2]float64{10, -5}, false},
		{"Outside left", [2]float64{-5, 8}, false},
		{"Outside right", [2]float64{25, 8}, false},
		{"On edge", [2]float64{10, 0}, true}, // Edge is considered inside
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pointInTriangle(
				tt.point[0], tt.point[1],
				triangle[0], triangle[1],
				triangle[2], triangle[3],
				triangle[4], triangle[5],
			)

			if result != tt.expected {
				t.Errorf("%s: point (%.1f, %.1f) expected %v, got %v",
					tt.name, tt.point[0], tt.point[1], tt.expected, result)
			}
		})
	}
}

// TestLineSegmentsIntersect tests the line segment intersection algorithm.
func TestLineSegmentsIntersect(t *testing.T) {
	tests := []struct {
		name     string
		seg1     [4]float64 // p1X, p1Y, p2X, p2Y
		seg2     [4]float64 // q1X, q1Y, q2X, q2Y
		expected bool
	}{
		{
			name:     "Crossing at center",
			seg1:     [4]float64{0, 10, 20, 10},
			seg2:     [4]float64{10, 0, 10, 20},
			expected: true,
		},
		{
			name:     "Parallel - no intersection",
			seg1:     [4]float64{0, 0, 10, 0},
			seg2:     [4]float64{0, 5, 10, 5},
			expected: false,
		},
		{
			name:     "Touching at endpoint",
			seg1:     [4]float64{0, 0, 10, 0},
			seg2:     [4]float64{10, 0, 20, 0},
			expected: true, // Endpoint touch counts as intersection
		},
		{
			name:     "Non-intersecting - separated",
			seg1:     [4]float64{0, 0, 10, 0},
			seg2:     [4]float64{20, 20, 30, 20},
			expected: false,
		},
		{
			name:     "Perpendicular intersection",
			seg1:     [4]float64{0, 0, 10, 10},
			seg2:     [4]float64{0, 10, 10, 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lineSegmentsIntersect(
				tt.seg1[0], tt.seg1[1], tt.seg1[2], tt.seg1[3],
				tt.seg2[0], tt.seg2[1], tt.seg2[2], tt.seg2[3],
			)

			if result != tt.expected {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

// TestMixedWallTypes tests terrain with both regular and diagonal walls.
func TestMixedWallTypes(t *testing.T) {
	// Create 3x3 terrain with mixed wall types
	// Layout:
	// [ / ][ W ][ \ ]
	// [ W ][   ][ W ]
	// [ \ ][ W ][ / ]
	terr := &terrain.Terrain{
		Width:  3,
		Height: 3,
		Tiles: [][]terrain.TileType{
			{terrain.TileWallNE, terrain.TileWall, terrain.TileWallNW}, // Row 0
			{terrain.TileWall, terrain.TileFloor, terrain.TileWall},    // Row 1
			{terrain.TileWallSE, terrain.TileWall, terrain.TileWallSW}, // Row 2
		},
	}

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(terr)

	tests := []struct {
		name         string
		entityBounds [4]float64
		expectHit    bool
	}{
		{
			name:         "Center floor - no collision",
			entityBounds: [4]float64{40, 40, 56, 56}, // Center tile (1,1)
			expectHit:    false,
		},
		{
			name:         "Regular wall - collision",
			entityBounds: [4]float64{40, 8, 56, 24}, // Top center (1,0)
			expectHit:    true,
		},
		{
			name:         "NE diagonal solid area",
			entityBounds: [4]float64{24, 24, 28, 28}, // Bottom-right of (0,0)
			expectHit:    true,
		},
		{
			name:         "NE diagonal open area",
			entityBounds: [4]float64{4, 4, 12, 12}, // Top-left of (0,0)
			expectHit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hit := checker.CheckCollisionBounds(
				tt.entityBounds[0],
				tt.entityBounds[1],
				tt.entityBounds[2],
				tt.entityBounds[3],
			)

			if hit != tt.expectHit {
				t.Errorf("%s: expected hit=%v, got hit=%v", tt.name, tt.expectHit, hit)
			}
		})
	}
}

// BenchmarkDiagonalWallCollision benchmarks diagonal wall collision detection.
func BenchmarkDiagonalWallCollision(b *testing.B) {
	// Create terrain with diagonal wall
	terr := &terrain.Terrain{
		Width:  1,
		Height: 1,
		Tiles:  [][]terrain.TileType{{terrain.TileWallNE}},
	}

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(terr)

	// Entity bounds in center
	minX, minY := 12.0, 12.0
	maxX, maxY := 20.0, 20.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckCollisionBounds(minX, minY, maxX, maxY)
	}
}

// BenchmarkRegularWallCollision benchmarks regular wall collision for comparison.
func BenchmarkRegularWallCollision(b *testing.B) {
	// Create terrain with regular wall
	terr := &terrain.Terrain{
		Width:  1,
		Height: 1,
		Tiles:  [][]terrain.TileType{{terrain.TileWall}},
	}

	checker := NewTerrainCollisionChecker(32, 32)
	checker.SetTerrain(terr)

	minX, minY := 12.0, 12.0
	maxX, maxY := 20.0, 20.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.CheckCollisionBounds(minX, minY, maxX, maxY)
	}
}
