package engine

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkRenderSystem_BatchGeometry benchmarks batch geometry building
// to measure allocation reduction from reusing vertex/index buffers.
func BenchmarkRenderSystem_BatchGeometry(b *testing.B) {
	// Create test entities with sprites (all share same sprite image for batching)
	entities := make([]*Entity, 200)
	spriteImage := ebiten.NewImage(16, 16)

	for i := 0; i < 200; i++ {
		entity := NewEntity(uint64(i))
		entity.AddComponent(&PositionComponent{X: float64(i * 20), Y: float64(i * 20)})
		sprite := NewSpriteComponent(16, 16, color.White)
		sprite.Image = spriteImage
		entity.AddComponent(sprite)
		entities[i] = entity
	}

	cameraSystem := NewCameraSystem(800, 600)
	renderSystem := NewRenderSystem(cameraSystem)
	renderSystem.EnableBatching(true)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = renderSystem.buildBatchGeometry(entities, spriteImage)
	}
}
