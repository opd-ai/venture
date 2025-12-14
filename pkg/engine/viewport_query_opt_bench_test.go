package engine

import (
	"testing"
)

// BenchmarkViewportQueryInto validates the QueryBoundsInto optimization.
// This method is used by the RenderSystem to avoid per-frame allocations
// in the viewport culling hot path.
func BenchmarkViewportQueryInto(b *testing.B) {
	// Setup spatial partition with entities spread across a large world
	sp := NewSpatialPartitionSystem(2000, 2000)
	entities := make([]*Entity, 200)
	for i := 0; i < 200; i++ {
		e := NewEntity(uint64(i))
		// Spread entities across world
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&EbitenSprite{Width: 32, Height: 32, Visible: true})
		entities[i] = e
	}
	sp.Update(entities, 0.016)

	// Query bounds simulating viewport (returns ~20 entities)
	bounds := Bounds{X: 0, Y: 0, Width: 800, Height: 600}
	buffer := make([]*Entity, 0, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer = buffer[:0]
		buffer = sp.QueryBoundsInto(bounds, buffer)
	}
}
