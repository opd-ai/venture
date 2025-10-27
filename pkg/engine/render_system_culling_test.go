package engine

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestRenderSystem_ViewportCulling(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 400, Y: 300})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system
	renderSystem := NewRenderSystem(cameraSystem)

	// Create spatial partition system
	spatialPartition := NewSpatialPartitionSystem(3000, 3000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)

	// Create entities
	entities := []*Entity{}

	// Visible entities (near camera at 400,300)
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i + 2))
		entity.AddComponent(&PositionComponent{X: 400 + float64(i*20), Y: 300})
		entity.AddComponent(&EbitenSprite{
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities = append(entities, entity)
	}

	// Off-screen entities (far from camera)
	for i := 0; i < 20; i++ {
		entity := NewEntity(uint64(i + 12))
		entity.AddComponent(&PositionComponent{X: 2500 + float64(i*50), Y: 2500})
		entity.AddComponent(&EbitenSprite{
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities = append(entities, entity)
	}

	// Update spatial partition
	spatialPartition.Update(entities, 0.016)

	// Create a dummy screen
	screen := ebiten.NewImage(800, 600)

	// Draw with culling enabled
	renderSystem.Draw(screen, entities)

	// Check statistics
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 30 {
		t.Errorf("Expected 30 total entities, got %d", stats.TotalEntities)
	}

	// Should have culled the off-screen entities
	if stats.CulledEntities < 15 {
		t.Errorf("Expected at least 15 culled entities, got %d", stats.CulledEntities)
	}

	if stats.RenderedEntities != stats.TotalEntities-stats.CulledEntities {
		t.Error("Rendered count doesn't match total - culled")
	}
}

func TestRenderSystem_CullingDisabled(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 400, Y: 300})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system
	renderSystem := NewRenderSystem(cameraSystem)

	// Disable culling
	renderSystem.EnableCulling(false)

	// Create entities
	entities := []*Entity{}
	sprite := ebiten.NewImage(32, 32) // Create a sprite image
	for i := 0; i < 20; i++ {
		entity := NewEntity(uint64(i + 2))
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: float64(i * 100)})
		entity.AddComponent(&EbitenSprite{
			Image:   sprite,
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities = append(entities, entity)
	}

	screen := ebiten.NewImage(800, 600)

	// Draw with culling disabled
	renderSystem.Draw(screen, entities)

	// Check statistics - all should be rendered when culling is disabled
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 20 {
		t.Errorf("Expected 20 total entities, got %d", stats.TotalEntities)
	}

	// With culling disabled, all entities should be rendered
	if stats.RenderedEntities != 20 {
		t.Errorf("Expected 20 rendered entities (culling disabled), got %d", stats.RenderedEntities)
	}

	if stats.CulledEntities != 0 {
		t.Errorf("Expected 0 culled entities (culling disabled), got %d", stats.CulledEntities)
	}
}

func TestRenderSystem_GetStats(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Initially stats should be zero
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 0 {
		t.Error("Initial total entities should be 0")
	}
	if stats.RenderedEntities != 0 {
		t.Error("Initial rendered entities should be 0")
	}
	if stats.CulledEntities != 0 {
		t.Error("Initial culled entities should be 0")
	}
}

func TestRenderSystem_SetSpatialPartition(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Initially no spatial partition
	if renderSystem.spatialPartition != nil {
		t.Error("Spatial partition should be nil initially")
	}

	// Set spatial partition
	spatialPartition := NewSpatialPartitionSystem(1000, 1000)
	renderSystem.SetSpatialPartition(spatialPartition)

	if renderSystem.spatialPartition == nil {
		t.Error("Spatial partition should be set")
	}
}

func TestRenderSystem_EnableCulling(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Initially culling should be disabled (due to spatial partition issue - see Phase 2.4 bug fix)
	if renderSystem.enableCulling {
		t.Error("Culling should be disabled by default (temporary fix for spatial partition query bug)")
	}

	// Enable culling
	renderSystem.EnableCulling(true)
	if !renderSystem.enableCulling {
		t.Error("Culling should be enabled after EnableCulling(true)")
	}

	// Disable culling
	renderSystem.EnableCulling(false)
	if renderSystem.enableCulling {
		t.Error("Culling should be disabled")
	}

	// Re-enable culling
	renderSystem.EnableCulling(true)
	if !renderSystem.enableCulling {
		t.Error("Culling should be enabled")
	}
}

// Benchmark viewport culling performance
func BenchmarkRenderSystem_ViewportCulling(b *testing.B) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 500, Y: 500})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with spatial partition
	renderSystem := NewRenderSystem(cameraSystem)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)

	// Create 2000 entities scattered across world
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i + 2))
		x := float64((i % 50) * 100)
		y := float64((i / 50) * 100)
		entity.AddComponent(&PositionComponent{X: x, Y: y})
		entity.AddComponent(&EbitenSprite{
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities[i] = entity
	}

	spatialPartition.Update(entities, 0.016)
	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSystem.Draw(screen, entities)
	}
}

// Benchmark without culling for comparison
func BenchmarkRenderSystem_NoCulling(b *testing.B) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 500, Y: 500})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system WITHOUT culling
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableCulling(false)

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i + 2))
		x := float64((i % 50) * 100)
		y := float64((i / 50) * 100)
		entity.AddComponent(&PositionComponent{X: x, Y: y})
		entity.AddComponent(&EbitenSprite{
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities[i] = entity
	}

	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSystem.Draw(screen, entities)
	}
}
