package engine

import (
	"testing"
)

func TestNewViewportOptimizer(t *testing.T) {
	vo := NewViewportOptimizer()
	if vo == nil {
		t.Fatal("NewViewportOptimizer returned nil")
	}

	if vo.tileSize != 32.0 {
		t.Errorf("Expected tileSize 32.0, got %f", vo.tileSize)
	}

	if vo.marginTiles != 1 {
		t.Errorf("Expected marginTiles 1, got %d", vo.marginTiles)
	}

	// Verify visibleBuffer is initialized (AUDIT.md Priority 1.2)
	if vo.visibleBuffer == nil {
		t.Error("Expected visibleBuffer to be initialized")
	}
	if cap(vo.visibleBuffer) != 256 {
		t.Errorf("Expected visibleBuffer capacity 256, got %d", cap(vo.visibleBuffer))
	}
}

func TestSetTileSize(t *testing.T) {
	vo := NewViewportOptimizer()
	vo.SetTileSize(64.0)

	if vo.tileSize != 64.0 {
		t.Errorf("Expected tileSize 64.0, got %f", vo.tileSize)
	}
}

func TestSetMarginTiles(t *testing.T) {
	vo := NewViewportOptimizer()
	vo.SetMarginTiles(2)

	if vo.marginTiles != 2 {
		t.Errorf("Expected marginTiles 2, got %d", vo.marginTiles)
	}
}

func TestCalculateViewportBounds(t *testing.T) {
	tests := []struct {
		name           string
		cameraX        float64
		cameraY        float64
		viewportWidth  float64
		viewportHeight float64
		zoom           float64
		tileSize       float64
		marginTiles    int
		wantX          float64
		wantY          float64
		wantWidth      float64
		wantHeight     float64
	}{
		{
			name:           "1920x1080 at zoom 1.0",
			cameraX:        500,
			cameraY:        500,
			viewportWidth:  1920,
			viewportHeight: 1080,
			zoom:           1.0,
			tileSize:       32,
			marginTiles:    1,
			wantX:          500 - 1920/2 - 32,
			wantY:          500 - 1080/2 - 32,
			wantWidth:      1920 + 64,
			wantHeight:     1080 + 64,
		},
		{
			name:           "1920x1080 at zoom 2.0",
			cameraX:        500,
			cameraY:        500,
			viewportWidth:  1920,
			viewportHeight: 1080,
			zoom:           2.0,
			tileSize:       32,
			marginTiles:    1,
			wantX:          500 - 1920/4 - 32,
			wantY:          500 - 1080/4 - 32,
			wantWidth:      1920/2 + 64,
			wantHeight:     1080/2 + 64,
		},
		{
			name:           "800x600 at zoom 1.0",
			cameraX:        0,
			cameraY:        0,
			viewportWidth:  800,
			viewportHeight: 600,
			zoom:           1.0,
			tileSize:       32,
			marginTiles:    1,
			wantX:          -400 - 32,
			wantY:          -300 - 32,
			wantWidth:      800 + 64,
			wantHeight:     600 + 64,
		},
		{
			name:           "2-tile margin",
			cameraX:        100,
			cameraY:        100,
			viewportWidth:  1920,
			viewportHeight: 1080,
			zoom:           1.0,
			tileSize:       32,
			marginTiles:    2,
			wantX:          100 - 960 - 64,
			wantY:          100 - 540 - 64,
			wantWidth:      1920 + 128,
			wantHeight:     1080 + 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo := NewViewportOptimizer()
			vo.SetTileSize(tt.tileSize)
			vo.SetMarginTiles(tt.marginTiles)

			bounds := vo.CalculateViewportBounds(
				tt.cameraX, tt.cameraY,
				tt.viewportWidth, tt.viewportHeight,
				tt.zoom,
			)

			if bounds.X != tt.wantX {
				t.Errorf("X: got %f, want %f", bounds.X, tt.wantX)
			}
			if bounds.Y != tt.wantY {
				t.Errorf("Y: got %f, want %f", bounds.Y, tt.wantY)
			}
			if bounds.Width != tt.wantWidth {
				t.Errorf("Width: got %f, want %f", bounds.Width, tt.wantWidth)
			}
			if bounds.Height != tt.wantHeight {
				t.Errorf("Height: got %f, want %f", bounds.Height, tt.wantHeight)
			}
		})
	}
}

func TestFrustumCull(t *testing.T) {
	tests := []struct {
		name           string
		entityX        float64
		entityY        float64
		entityWidth    float64
		entityHeight   float64
		viewportBounds Bounds
		wantVisible    bool
	}{
		{
			name:         "entity fully inside viewport",
			entityX:      500,
			entityY:      500,
			entityWidth:  32,
			entityHeight: 32,
			viewportBounds: Bounds{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
			wantVisible: true,
		},
		{
			name:         "entity fully outside viewport",
			entityX:      2500,
			entityY:      2500,
			entityWidth:  32,
			entityHeight: 32,
			viewportBounds: Bounds{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
			wantVisible: false,
		},
		{
			name:         "entity partially inside viewport",
			entityX:      1900,
			entityY:      500,
			entityWidth:  64,
			entityHeight: 64,
			viewportBounds: Bounds{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
			wantVisible: true,
		},
		{
			name:         "entity at edge (touching)",
			entityX:      1920,
			entityY:      500,
			entityWidth:  32,
			entityHeight: 32,
			viewportBounds: Bounds{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
			wantVisible: true,
		},
		{
			name:         "entity just beyond edge",
			entityX:      1952,
			entityY:      500,
			entityWidth:  32,
			entityHeight: 32,
			viewportBounds: Bounds{
				X:      0,
				Y:      0,
				Width:  1920,
				Height: 1080,
			},
			wantVisible: false,
		},
	}

	vo := NewViewportOptimizer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visible := vo.FrustumCull(
				tt.entityX, tt.entityY,
				tt.entityWidth, tt.entityHeight,
				tt.viewportBounds,
			)

			if visible != tt.wantVisible {
				t.Errorf("FrustumCull() = %v, want %v", visible, tt.wantVisible)
			}
		})
	}
}

func TestOptimizeVisibleSet(t *testing.T) {
	// Create world and spatial partition
	world := NewWorld()
	partition := NewSpatialPartitionSystem(1000, 1000)

	// Create camera
	camera := NewCameraComponent()
	camera.X = 500
	camera.Y = 500
	camera.Zoom = 1.0

	// Create entities at various positions
	entities := []*Entity{
		// Visible entities (near camera)
		createTestEntityAt(world, 500, 500, true),  // center
		createTestEntityAt(world, 600, 600, false), // visible
		createTestEntityAt(world, 400, 400, false), // visible

		// Off-screen entities (far from camera)
		createTestEntityAt(world, 2000, 2000, false), // far off-screen
		createTestEntityAt(world, 50, 50, false),     // far off-screen
	}

	// Build spatial partition
	partition.Rebuild(entities)

	// Optimize visible set
	vo := NewViewportOptimizer()
	visible := vo.OptimizeVisibleSet(camera, 1920, 1080, partition, entities)

	// Check stats
	stats := vo.Stats()
	if stats.TotalEntities != 5 {
		t.Errorf("TotalEntities: got %d, want 5", stats.TotalEntities)
	}

	// Should include player entity always
	hasPlayer := false
	for _, e := range visible {
		if e.HasComponent("input") {
			hasPlayer = true
			break
		}
	}
	if !hasPlayer {
		t.Error("Player entity not included in visible set")
	}

	// Culled entities should be > 0 (far entities culled)
	if stats.CulledEntities == 0 {
		t.Error("Expected some entities to be culled")
	}
}

func TestOptimizeVisibleSet_NilCamera(t *testing.T) {
	world := NewWorld()
	partition := NewSpatialPartitionSystem(1000, 1000)

	entities := []*Entity{
		createTestEntityAt(world, 500, 500, false),
	}

	vo := NewViewportOptimizer()
	visible := vo.OptimizeVisibleSet(nil, 1920, 1080, partition, entities)

	// Without camera, should return all entities
	if len(visible) != len(entities) {
		t.Errorf("Expected all entities without camera, got %d, want %d", len(visible), len(entities))
	}
}

func TestOptimizeVisibleSet_NilPartition(t *testing.T) {
	world := NewWorld()
	camera := NewCameraComponent()

	entities := []*Entity{
		createTestEntityAt(world, 500, 500, false),
	}

	vo := NewViewportOptimizer()
	visible := vo.OptimizeVisibleSet(camera, 1920, 1080, nil, entities)

	// Without partition, should return all entities
	if len(visible) != len(entities) {
		t.Errorf("Expected all entities without partition, got %d, want %d", len(visible), len(entities))
	}
}

func TestOffScreenPercentage(t *testing.T) {
	vo := NewViewportOptimizer()

	// Manually set stats
	vo.mu.Lock()
	vo.stats.VisibleEntities = 100
	vo.stats.OffScreenRendered = 3
	vo.mu.Unlock()

	pct := vo.OffScreenPercentage()
	expected := 3.0
	if pct != expected {
		t.Errorf("OffScreenPercentage: got %f, want %f", pct, expected)
	}
}

func TestOffScreenPercentage_ZeroVisible(t *testing.T) {
	vo := NewViewportOptimizer()

	pct := vo.OffScreenPercentage()
	if pct != 0.0 {
		t.Errorf("Expected 0.0 with zero visible entities, got %f", pct)
	}
}

func TestCullingEfficiency(t *testing.T) {
	vo := NewViewportOptimizer()

	// Manually set stats
	vo.mu.Lock()
	vo.stats.TotalEntities = 1000
	vo.stats.CulledEntities = 950
	vo.mu.Unlock()

	eff := vo.CullingEfficiency()
	expected := 0.95
	if eff != expected {
		t.Errorf("CullingEfficiency: got %f, want %f", eff, expected)
	}
}

func TestCullingEfficiency_ZeroEntities(t *testing.T) {
	vo := NewViewportOptimizer()

	eff := vo.CullingEfficiency()
	if eff != 1.0 {
		t.Errorf("Expected 1.0 with zero entities, got %f", eff)
	}
}

func TestValidateMetrics(t *testing.T) {
	tests := []struct {
		name              string
		visibleEntities   int
		offScreenRendered int
		wantValid         bool
	}{
		{
			name:              "below 5% threshold",
			visibleEntities:   100,
			offScreenRendered: 4,
			wantValid:         true,
		},
		{
			name:              "at 5% threshold",
			visibleEntities:   100,
			offScreenRendered: 5,
			wantValid:         false,
		},
		{
			name:              "above 5% threshold",
			visibleEntities:   100,
			offScreenRendered: 10,
			wantValid:         false,
		},
		{
			name:              "0% off-screen",
			visibleEntities:   100,
			offScreenRendered: 0,
			wantValid:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo := NewViewportOptimizer()
			vo.mu.Lock()
			vo.stats.VisibleEntities = tt.visibleEntities
			vo.stats.OffScreenRendered = tt.offScreenRendered
			vo.mu.Unlock()

			valid := vo.ValidateMetrics()
			if valid != tt.wantValid {
				t.Errorf("ValidateMetrics() = %v, want %v (off-screen: %.1f%%)",
					valid, tt.wantValid, vo.OffScreenPercentage())
			}
		})
	}
}

// Helper to create test entity at position
func createTestEntityAt(world *World, x, y float64, isPlayer bool) *Entity {
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: x, Y: y})

	sprite := &EbitenSprite{
		Width:   32,
		Height:  32,
		Visible: true,
	}
	entity.AddComponent(sprite)

	if isPlayer {
		// Use a stub input provider for testing
		entity.AddComponent(&StubInput{})
	}

	return entity
}

func BenchmarkCalculateViewportBounds(b *testing.B) {
	vo := NewViewportOptimizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vo.CalculateViewportBounds(500, 500, 1920, 1080, 1.0)
	}
}

func BenchmarkFrustumCull(b *testing.B) {
	vo := NewViewportOptimizer()
	viewportBounds := Bounds{X: 0, Y: 0, Width: 1920, Height: 1080}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vo.FrustumCull(500, 500, 32, 32, viewportBounds)
	}
}

func BenchmarkOptimizeVisibleSet(b *testing.B) {
	// Setup
	world := NewWorld()
	partition := NewSpatialPartitionSystem(5000, 5000)

	camera := NewCameraComponent()
	camera.X = 2500
	camera.Y = 2500
	camera.Zoom = 1.0

	// Create 1000 entities
	entities := make([]*Entity, 1000)
	for i := 0; i < 1000; i++ {
		x := float64(i%100) * 50
		y := float64(i/100) * 50
		entities[i] = createTestEntityAt(world, x, y, i == 0)
	}

	partition.Rebuild(entities)
	vo := NewViewportOptimizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vo.OptimizeVisibleSet(camera, 1920, 1080, partition, entities)
	}
}

// BenchmarkOptimizeVisibleSet_Allocations verifies zero-allocation queries (AUDIT.md Priority 1.2)
func BenchmarkOptimizeVisibleSet_Allocations(b *testing.B) {
	world := NewWorld()
	partition := NewSpatialPartitionSystem(5000, 5000)

	camera := NewCameraComponent()
	camera.X = 2500
	camera.Y = 2500
	camera.Zoom = 1.0

	// Create 500 entities in viewport
	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		x := 2500 + float64(i%20)*40
		y := 2500 + float64(i/20)*40
		entities[i] = createTestEntityAt(world, x, y, i == 0)
	}

	partition.Rebuild(entities)
	vo := NewViewportOptimizer()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vo.OptimizeVisibleSet(camera, 1920, 1080, partition, entities)
	}
}
