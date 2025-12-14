package engine

import (
	"image/color"
	"testing"
)

// BenchmarkShadowSystem_CollectShadowCasters benchmarks the shadow caster collection.
// This tests the hot path of finding all shadow-casting entities.
func BenchmarkShadowSystem_CollectShadowCasters(b *testing.B) {
	world := NewWorld()
	system := NewShadowSystem(world)

	// Create 200 entities with shadow components (simulating a scene with shadows)
	for i := 0; i < 200; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i%20) * 50,
			Y: float64(i/20) * 50,
		})
		entity.AddComponent(&ShadowComponent{
			Enabled:     true,
			CastsShadow: true,
			Radius:      16.0,
			Opacity:     0.5,
			Color:       color.RGBA{0, 0, 0, 128},
			ShadowType:  ShadowTypeHard,
		})
	}

	// Process pending entities
	world.Update(0.016)

	// Set viewport to include all entities
	system.SetViewport(0, 0, 1000, 500)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Collect shadow casters from a light at center
		casters := system.collectShadowCasters(500, 250, 1000)
		_ = casters
	}
}

// BenchmarkShadowSystem_CollectShadowCasters_LargeScene benchmarks with more entities.
func BenchmarkShadowSystem_CollectShadowCasters_LargeScene(b *testing.B) {
	world := NewWorld()
	system := NewShadowSystem(world)
	system.SetMaxShadows(100) // Limit to 100 shadows

	// Create 500 entities (only 100 will be collected due to limit)
	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i%25) * 40,
			Y: float64(i/25) * 40,
		})
		entity.AddComponent(&ShadowComponent{
			Enabled:     true,
			CastsShadow: true,
			Radius:      16.0,
			Opacity:     0.5,
			Color:       color.RGBA{0, 0, 0, 128},
			ShadowType:  ShadowTypeHard,
		})
	}

	world.Update(0.016)
	system.SetViewport(0, 0, 1000, 800)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		casters := system.collectShadowCasters(500, 400, 1000)
		_ = casters
	}
}
