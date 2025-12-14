// Benchmark for sortEntitiesByLayer optimization
package engine

import (
	"image/color"
	"testing"
)

// BenchmarkSortEntitiesByLayer measures sorting performance with different entity counts
func BenchmarkSortEntitiesByLayer(b *testing.B) {
	tests := []struct {
		name  string
		count int
	}{
		{"100 entities", 100},
		{"500 entities", 500},
		{"1000 entities", 1000},
		{"2000 entities", 2000},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create test entities with sprites at different layers
			entities := make([]*Entity, tt.count)
			for i := 0; i < tt.count; i++ {
				entity := NewEntity(uint64(i))
				entity.AddComponent(&PositionComponent{X: float64(i % 100), Y: float64(i / 100)})
				sprite := NewSpriteComponent(16, 16, color.RGBA{R: 255, G: 255, B: 255, A: 255})
				sprite.Layer = i % 10 // Distribute across 10 layers
				entity.AddComponent(sprite)
				entities[i] = entity
			}

			// Create render system
			cameraSystem := NewCameraSystem(800, 600)
			renderSystem := NewRenderSystem(cameraSystem)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = renderSystem.sortEntitiesByLayer(entities)
			}
		})
	}
}
