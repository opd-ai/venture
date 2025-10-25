package engine

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestRenderSystem_Batching(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 400, Y: 300})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with batching enabled
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(true)

	// Create shared sprite images for batching
	sprite1 := ebiten.NewImage(32, 32)
	sprite2 := ebiten.NewImage(32, 32)

	// Create entities with shared sprites
	entities := []*Entity{}

	// 10 entities with sprite1 (should batch together)
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i + 2))
		entity.AddComponent(&PositionComponent{X: 400 + float64(i*20), Y: 300})
		entity.AddComponent(&EbitenSprite{
			Image:   sprite1,
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities = append(entities, entity)
	}

	// 5 entities with sprite2 (should batch together)
	for i := 0; i < 5; i++ {
		entity := NewEntity(uint64(i + 12))
		entity.AddComponent(&PositionComponent{X: 500 + float64(i*20), Y: 350})
		entity.AddComponent(&EbitenSprite{
			Image:   sprite2,
			Width:   32,
			Height:  32,
			Color:   color.White,
			Visible: true,
		})
		entities = append(entities, entity)
	}

	screen := ebiten.NewImage(800, 600)

	// Draw with batching
	renderSystem.Draw(screen, entities)

	// Check statistics
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 15 {
		t.Errorf("Expected 15 total entities, got %d", stats.TotalEntities)
	}

	if stats.RenderedEntities != 15 {
		t.Errorf("Expected 15 rendered entities, got %d", stats.RenderedEntities)
	}

	// Should create 2 batches (one for each sprite image)
	if stats.BatchCount != 2 {
		t.Errorf("Expected 2 batches, got %d", stats.BatchCount)
	}
}

func TestRenderSystem_BatchingDisabled(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 400, Y: 300})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with batching DISABLED
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(false)

	// Create shared sprite
	sprite := ebiten.NewImage(32, 32)

	// Create entities with same sprite
	entities := []*Entity{}
	for i := 0; i < 10; i++ {
		entity := NewEntity(uint64(i + 2))
		entity.AddComponent(&PositionComponent{X: 400 + float64(i*20), Y: 300})
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

	// Draw without batching
	renderSystem.Draw(screen, entities)

	// Check statistics
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 10 {
		t.Errorf("Expected 10 total entities, got %d", stats.TotalEntities)
	}

	if stats.RenderedEntities != 10 {
		t.Errorf("Expected 10 rendered entities, got %d", stats.RenderedEntities)
	}

	// No batches when batching is disabled
	if stats.BatchCount != 0 {
		t.Errorf("Expected 0 batches (batching disabled), got %d", stats.BatchCount)
	}
}

func TestRenderSystem_BatchingWithMultipleSprites(t *testing.T) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 400, Y: 300})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(true)

	// Create 5 unique sprites
	sprites := make([]*ebiten.Image, 5)
	for i := range sprites {
		sprites[i] = ebiten.NewImage(32, 32)
	}

	// Create entities - 4 entities per sprite
	entities := []*Entity{}
	entityID := uint64(2)
	for _, sprite := range sprites {
		for j := 0; j < 4; j++ {
			entity := NewEntity(entityID)
			entityID++
			entity.AddComponent(&PositionComponent{X: float64(100 + j*50), Y: float64(100 + j*50)})
			entity.AddComponent(&EbitenSprite{
				Image:   sprite,
				Width:   32,
				Height:  32,
				Color:   color.White,
				Visible: true,
			})
			entities = append(entities, entity)
		}
	}

	screen := ebiten.NewImage(800, 600)

	// Draw with batching
	renderSystem.Draw(screen, entities)

	// Check statistics
	stats := renderSystem.GetStats()

	if stats.TotalEntities != 20 {
		t.Errorf("Expected 20 total entities, got %d", stats.TotalEntities)
	}

	if stats.RenderedEntities != 20 {
		t.Errorf("Expected 20 rendered entities, got %d", stats.RenderedEntities)
	}

	// Should create 5 batches (one for each unique sprite)
	if stats.BatchCount != 5 {
		t.Errorf("Expected 5 batches, got %d", stats.BatchCount)
	}
}

func TestRenderSystem_EnableBatching(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Initially batching should be enabled
	if !renderSystem.enableBatching {
		t.Error("Batching should be enabled by default")
	}

	// Disable batching
	renderSystem.EnableBatching(false)
	if renderSystem.enableBatching {
		t.Error("Batching should be disabled")
	}

	// Re-enable batching
	renderSystem.EnableBatching(true)
	if !renderSystem.enableBatching {
		t.Error("Batching should be enabled")
	}
}

func TestRenderSystem_BatchMapPooling(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)

	// Pool should be empty initially
	if len(renderSystem.batchPool) != 0 {
		t.Error("Batch pool should be empty initially")
	}

	// Get a batch map
	batches1 := renderSystem.getBatchMap()
	if batches1 == nil {
		t.Error("getBatchMap should return a valid map")
	}

	// Return it to pool
	renderSystem.returnBatchMap(batches1)
	if len(renderSystem.batchPool) != 1 {
		t.Error("Batch pool should have 1 entry after return")
	}

	// Get it again - should reuse from pool
	batches2 := renderSystem.getBatchMap()
	if len(renderSystem.batchPool) != 0 {
		t.Error("Batch pool should be empty after reuse")
	}

	// Should be the same map
	if &batches1 != &batches2 {
		// Note: Pointer comparison may not work, check pool size instead
		t.Log("Maps are different (acceptable)")
	}
}

// Benchmark batching vs no batching
func BenchmarkRenderSystem_Batching(b *testing.B) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 500, Y: 500})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with batching enabled
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(true)

	// Create 10 shared sprites
	sprites := make([]*ebiten.Image, 10)
	for i := range sprites {
		sprites[i] = ebiten.NewImage(32, 32)
	}

	// Create 2000 entities - 200 per sprite (high reuse)
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i + 2))
		x := float64((i % 50) * 100)
		y := float64((i / 50) * 100)
		entity.AddComponent(&PositionComponent{X: x, Y: y})

		// Assign sprite based on index (creates 10 batches)
		spriteIdx := i % 10
		entity.AddComponent(&EbitenSprite{
			Image:   sprites[spriteIdx],
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

// Benchmark without batching for comparison
func BenchmarkRenderSystem_NoBatching(b *testing.B) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 500, Y: 500})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with batching DISABLED
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(false)

	// Create 10 shared sprites
	sprites := make([]*ebiten.Image, 10)
	for i := range sprites {
		sprites[i] = ebiten.NewImage(32, 32)
	}

	// Create 2000 entities
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i + 2))
		x := float64((i % 50) * 100)
		y := float64((i / 50) * 100)
		entity.AddComponent(&PositionComponent{X: x, Y: y})

		spriteIdx := i % 10
		entity.AddComponent(&EbitenSprite{
			Image:   sprites[spriteIdx],
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

// Benchmark with culling AND batching
func BenchmarkRenderSystem_CullingAndBatching(b *testing.B) {
	// Create camera system
	cameraSystem := NewCameraSystem(800, 600)
	camera := NewEntity(1)
	camera.AddComponent(&PositionComponent{X: 500, Y: 500})
	camera.AddComponent(NewCameraComponent())
	cameraSystem.SetActiveCamera(camera)

	// Create render system with BOTH optimizations
	renderSystem := NewRenderSystem(cameraSystem)
	spatialPartition := NewSpatialPartitionSystem(5000, 5000)
	renderSystem.SetSpatialPartition(spatialPartition)
	renderSystem.EnableCulling(true)
	renderSystem.EnableBatching(true)

	// Create 10 shared sprites
	sprites := make([]*ebiten.Image, 10)
	for i := range sprites {
		sprites[i] = ebiten.NewImage(32, 32)
	}

	// Create 2000 entities scattered across world
	entities := make([]*Entity, 2000)
	for i := 0; i < 2000; i++ {
		entity := NewEntity(uint64(i + 2))
		x := float64((i % 50) * 100)
		y := float64((i / 50) * 100)
		entity.AddComponent(&PositionComponent{X: x, Y: y})

		spriteIdx := i % 10
		entity.AddComponent(&EbitenSprite{
			Image:   sprites[spriteIdx],
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
