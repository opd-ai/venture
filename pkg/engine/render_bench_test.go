package engine

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkRenderSystemNoBatching measures rendering performance without batching
func BenchmarkRenderSystemNoBatching(b *testing.B) {
	// Create test world with entities
	world := NewWorld()

	// Create camera system first
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 0, Y: 0})
	camera.AddComponent(&CameraComponent{
		Zoom: 1.0,
	})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)
	renderSys.enableBatching = false // Disable batching

	// Create a dummy sprite image (1x1 white pixel)
	spriteImg := ebiten.NewImage(16, 16)
	spriteImg.Fill(color.White)

	// Create 100 entities with sprites at different positions
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i * 10),
			Y: float64(i * 10),
		})
		entity.AddComponent(&EbitenSprite{
			Image:   spriteImg,
			Width:   16,
			Height:  16,
			Visible: true,
			Color:   color.White,
		})
		entities[i] = entity
	}

	// Create screen image
	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSys.Draw(screen, entities)
	}
}

// BenchmarkRenderSystemWithBatching measures rendering performance with batching enabled
func BenchmarkRenderSystemWithBatching(b *testing.B) {
	// Create test world with entities
	world := NewWorld()

	// Create camera system first
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 0, Y: 0})
	camera.AddComponent(&CameraComponent{
		Zoom: 1.0,
	})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)
	renderSys.enableBatching = true // Enable batching

	// Create a dummy sprite image (1x1 white pixel)
	spriteImg := ebiten.NewImage(16, 16)
	spriteImg.Fill(color.White)

	// Create 100 entities with sprites at different positions
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i * 10),
			Y: float64(i * 10),
		})
		entity.AddComponent(&EbitenSprite{
			Image:   spriteImg,
			Width:   16,
			Height:  16,
			Visible: true,
			Color:   color.White,
		})
		entities[i] = entity
	}

	// Create screen image
	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSys.Draw(screen, entities)
	}
}

// BenchmarkRenderSystemManySprites measures rendering with 500 sprites (realistic game load)
func BenchmarkRenderSystemManySprites(b *testing.B) {
	// Create test world with entities
	world := NewWorld()

	// Create camera system first
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 250, Y: 250})
	camera.AddComponent(&CameraComponent{
		Zoom: 1.0,
	})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)
	renderSys.enableBatching = true // Enable batching

	// Create 5 different sprite images to test multi-batch scenario
	spriteImages := make([]*ebiten.Image, 5)
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Red
		color.RGBA{0, 255, 0, 255},   // Green
		color.RGBA{0, 0, 255, 255},   // Blue
		color.RGBA{255, 255, 0, 255}, // Yellow
		color.RGBA{255, 0, 255, 255}, // Magenta
	}
	for i := 0; i < 5; i++ {
		spriteImages[i] = ebiten.NewImage(16, 16)
		spriteImages[i].Fill(colors[i])
	}

	// Create 500 entities with sprites (100 per sprite image)
	entities := make([]*Entity, 500)
	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64((i % 50) * 10),
			Y: float64((i / 50) * 10),
		})
		entity.AddComponent(&EbitenSprite{
			Image:   spriteImages[i%5], // Distribute across 5 sprite images
			Width:   16,
			Height:  16,
			Visible: true,
			Color:   colors[i%5],
		})
		entities[i] = entity
	}

	// Create screen image
	screen := ebiten.NewImage(800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSys.Draw(screen, entities)
	}
}

// BenchmarkDrawBatchSmall measures batching performance with small batch (10 sprites)
func BenchmarkDrawBatchSmall(b *testing.B) {
	world := NewWorld()

	// Create camera system first
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 0, Y: 0})
	camera.AddComponent(&CameraComponent{
		Zoom: 1.0,
	})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)

	spriteImg := ebiten.NewImage(16, 16)
	spriteImg.Fill(color.White)

	// Create 10 entities
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i * 10),
			Y: float64(i * 10),
		})
		entity.AddComponent(&EbitenSprite{
			Image:   spriteImg,
			Width:   16,
			Height:  16,
			Visible: true,
			Color:   color.White,
		})
		entities[i] = entity
	}

	screen := ebiten.NewImage(800, 600)
	renderSys.screen = screen

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSys.drawBatch(entities)
	}
}

// BenchmarkDrawBatchLarge measures batching performance with large batch (100 sprites)
func BenchmarkDrawBatchLarge(b *testing.B) {
	world := NewWorld()

	// Create camera system first
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 0, Y: 0})
	camera.AddComponent(&CameraComponent{
		Zoom: 1.0,
	})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)

	spriteImg := ebiten.NewImage(16, 16)
	spriteImg.Fill(color.White)

	// Create 100 entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64((i % 10) * 10),
			Y: float64((i / 10) * 10),
		})
		entity.AddComponent(&EbitenSprite{
			Image:   spriteImg,
			Width:   16,
			Height:  16,
			Visible: true,
			Color:   color.White,
		})
		entities[i] = entity
	}

	screen := ebiten.NewImage(800, 600)
	renderSys.screen = screen

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderSys.drawBatch(entities)
	}
}

// BenchmarkDrawSpriteImage measures individual sprite drawing performance
// This tests the hot path for entities that don't go through batch rendering
// (e.g., entities with directional sprites or visual effects)
func BenchmarkDrawSpriteImage(b *testing.B) {
	world := NewWorld()

	// Create camera system
	cameraSys := NewCameraSystem(800, 600)
	camera := world.CreateEntity()
	camera.AddComponent(&PositionComponent{X: 0, Y: 0})
	camera.AddComponent(&CameraComponent{Zoom: 1.0})
	cameraSys.SetActiveCamera(camera)

	// Create render system with camera
	renderSys := NewRenderSystem(cameraSys)

	// Create sprite image
	spriteImg := ebiten.NewImage(32, 32)
	spriteImg.Fill(color.White)

	// Create sprite component
	sprite := &EbitenSprite{
		Image:    spriteImg,
		Width:    32,
		Height:   32,
		Visible:  true,
		Color:    color.White,
		Rotation: 0.5,
	}

	// Create screen
	screen := ebiten.NewImage(800, 600)
	renderSys.screen = screen

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Simulate rendering 100 sprites individually (like directional sprites)
		for j := 0; j < 100; j++ {
			renderSys.drawSpriteImage(spriteImg, sprite, float64(j*8), float64(j*8), 0, 0, 1.0, 1.0, 1.0, 1.0)
		}
	}
}
