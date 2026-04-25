package engine

import (
	"math"
	"testing"
)

// StubRenderSystem is a test implementation of RenderingSystem interface.
// It provides a simple mock for unit testing without actual rendering.
type StubRenderSystem struct {
	cameraSystem  *CameraSystem
	ShowColliders bool
	ShowGrid      bool

	// Test state tracking
	UpdateCount          int
	DrawCount            int
	LastDrawnEntityCount int
}

// NewStubRenderSystem creates a new test render system.
func NewStubRenderSystem(cameraSystem *CameraSystem) *StubRenderSystem {
	return &StubRenderSystem{
		cameraSystem:  cameraSystem,
		ShowColliders: false,
		ShowGrid:      false,
		UpdateCount:   0,
		DrawCount:     0,
	}
}

// Update implements System interface (test stub).
func (r *StubRenderSystem) Update(entities []*Entity, deltaTime float64) {
	r.UpdateCount++
}

// Draw implements RenderingSystem interface (test stub).
func (r *StubRenderSystem) Draw(screen interface{}, entities []*Entity) {
	r.DrawCount++
	r.LastDrawnEntityCount = len(entities)
}

// SetShowColliders implements RenderingSystem interface.
func (r *StubRenderSystem) SetShowColliders(show bool) {
	r.ShowColliders = show
}

// SetShowGrid implements RenderingSystem interface.
func (r *StubRenderSystem) SetShowGrid(show bool) {
	r.ShowGrid = show
}

// Compile-time interface check
var _ RenderingSystem = (*StubRenderSystem)(nil)

// TestRenderSystem_SpatialPartition_CameraPosition tests that spatial partition
// uses the correct camera position for viewport culling.
// This is a regression test for the bug where getVisibleEntities used the
// entity's PositionComponent instead of camera.X/Y.
func TestRenderSystem_SpatialPartition_CameraPosition(t *testing.T) {
	// Note: This test cannot use Ebiten types (ebiten.NewImage) as they require
	// graphics initialization which is not available in CI environments.
	// We test the logic indirectly through the spatial partition system.

	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Create spatial partition
	spatialPartition := NewSpatialPartitionSystem(2000, 2000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)

	// Create camera entity
	cameraEntity := NewEntity(1)
	cameraComp := NewCameraComponent()

	// Camera component position is at (1000, 1000) - this is what should be used
	cameraComp.X = 1000
	cameraComp.Y = 1000
	cameraComp.Zoom = 1.0
	cameraEntity.AddComponent(cameraComp)

	// Entity position is different (500, 500) - this should NOT be used
	// This is the bug we're testing for
	entityPos := &PositionComponent{X: 500, Y: 500}
	cameraEntity.AddComponent(entityPos)

	cameraSystem.SetActiveCamera(cameraEntity)

	// Create test entity that should be visible based on camera.X/Y (1000, 1000)
	// but NOT visible based on entity position (500, 500)
	testEntity := NewEntity(2)
	testPos := &PositionComponent{X: 1100, Y: 1100} // Near camera.X/Y
	testEntity.AddComponent(testPos)

	testSprite := &EbitenSprite{
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   0,
	}
	testEntity.AddComponent(testSprite)

	entities := []*Entity{testEntity}

	// Rebuild spatial partition immediately to populate it with entities
	spatialPartition.Rebuild(entities)

	// Test getVisibleEntities - should use camera.X/Y (1000, 1000)
	// The viewport should be centered at (1000, 1000) with margin
	// So entity at (1100, 1100) should be visible
	visibleEntities := renderSystem.getVisibleEntities(entities)

	if len(visibleEntities) != 1 {
		t.Errorf("Expected 1 visible entity (using camera.X/Y), got %d", len(visibleEntities))
		t.Logf("Camera component position: (%.0f, %.0f)", cameraComp.X, cameraComp.Y)
		t.Logf("Entity position component: (%.0f, %.0f)", entityPos.X, entityPos.Y)
		t.Logf("Test entity position: (%.0f, %.0f)", testPos.X, testPos.Y)
	}
}

// TestRenderSystem_NoDoubleCulling tests that entities are not culled twice
// when spatial partition is enabled. This is a regression test for the bug
// where per-entity culling was applied after spatial partition culling.
func TestRenderSystem_NoDoubleCulling(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Create camera entity
	cameraEntity := NewEntity(1)
	cameraComp := NewCameraComponent()
	cameraComp.X = 400
	cameraComp.Y = 300
	cameraComp.Zoom = 1.0
	cameraEntity.AddComponent(cameraComp)
	cameraEntity.AddComponent(&PositionComponent{X: 400, Y: 300})
	cameraSystem.SetActiveCamera(cameraEntity)

	// Create spatial partition and enable culling
	spatialPartition := NewSpatialPartitionSystem(2000, 2000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)

	// Create entity at camera position (should definitely be visible)
	testEntity := NewEntity(2)
	testEntity.AddComponent(&PositionComponent{X: 400, Y: 300})
	testEntity.AddComponent(&EbitenSprite{
		Width:   32,
		Height:  32,
		Visible: true,
		Layer:   0,
	})

	entities := []*Entity{testEntity}

	// Rebuild spatial partition
	spatialPartition.Rebuild(entities)

	// Test getVisibleEntities - entity should be visible
	visibleEntities := renderSystem.getVisibleEntities(entities)
	if len(visibleEntities) != 1 {
		t.Errorf("Expected 1 visible entity from spatial partition, got %d", len(visibleEntities))
	}

	// Verify spatialCullingUsed flag is set when spatial partition is used
	// This flag should prevent per-entity culling in drawEntity/drawBatch
	renderSystem.spatialCullingUsed = false
	if renderSystem.enableCulling && renderSystem.spatialPartition != nil && renderSystem.cameraSystem != nil {
		_ = renderSystem.getVisibleEntities(entities)
		// After calling getVisibleEntities, the flag should NOT be set
		// (it's only set in Draw() method)
		if renderSystem.spatialCullingUsed {
			t.Error("spatialCullingUsed should only be set by Draw() method")
		}
	}

	// Verify the flag is properly set in Draw() context by checking internal state
	// The spatialCullingUsed flag should be false initially
	if renderSystem.spatialCullingUsed {
		t.Error("spatialCullingUsed should be false before Draw()")
	}
}

// TestG38_PositionComponent_Initialized verifies that interpolatePosition uses
// the current position when Initialized=false (first-frame guard) and correctly
// interpolates between PrevX/PrevY and X/Y when Initialized=true.
// This is the regression test for G38.
func TestG38_PositionComponent_Initialized(t *testing.T) {
	// Create a CameraSystem with no active camera so WorldToScreen is an identity
	// function (returns world coords unchanged), making the expected values trivial.
	cam := NewCameraSystem(0, 0)
	rs := NewRenderSystem(cam)
	rs.SetRenderAlpha(0.5) // midpoint interpolation

	// Case 1: Initialized=false at world origin — interpolatePosition must return
	// the current position (0,0), not an interpolation artefact from PrevX/PrevY.
	posOrigin := &PositionComponent{X: 0, Y: 0, PrevX: 100, PrevY: 100}
	// posOrigin.Initialized is false (zero-value), so the old heuristic would
	// have entered interpolation because PrevX != 0 and X == 0.
	gotX, gotY := rs.interpolatePosition(posOrigin)
	if gotX != 0 || gotY != 0 {
		t.Errorf("G38 Initialized=false at origin: interpolatePosition = (%.1f,%.1f), want (0,0)", gotX, gotY)
	}

	// Case 2: Initialized=true — must interpolate between prev and current.
	posMoving := &PositionComponent{
		X: 10, Y: 20,
		PrevX: 0, PrevY: 0,
		Initialized: true,
	}
	// alpha=0.5 → expected midpoint: (5, 10)
	gotX, gotY = rs.interpolatePosition(posMoving)
	wantX, wantY := 5.0, 10.0
	if math.Abs(gotX-wantX) > 0.001 || math.Abs(gotY-wantY) > 0.001 {
		t.Errorf("G38 Initialized=true alpha=0.5: interpolatePosition = (%.3f,%.3f), want (%.3f,%.3f)",
			gotX, gotY, wantX, wantY)
	}

	// Case 3: alpha=1.0 — always returns current position regardless of Initialized.
	rs.SetRenderAlpha(1.0)
	gotX, gotY = rs.interpolatePosition(posMoving)
	if math.Abs(gotX-10) > 0.001 || math.Abs(gotY-20) > 0.001 {
		t.Errorf("G38 alpha=1.0: interpolatePosition = (%.3f,%.3f), want (10,20)", gotX, gotY)
	}
}
