package engine

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// BenchmarkLightingSystem_ApplyPointLight benchmarks point light application
// which was identified as a hot path with per-frame allocations.
func BenchmarkLightingSystem_ApplyPointLight(b *testing.B) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = true
	config.MaxLights = 20
	system := NewLightingSystem(world, config)
	system.SetViewport(0, 0, 800, 600)

	// Create scene and buffer images
	scene := ebiten.NewImage(800, 600)
	buffer := ebiten.NewImage(800, 600)

	// Create a test light
	lwp := &lightWithPosition{
		light: &LightComponent{
			Enabled:   true,
			Intensity: 1.0,
			Radius:    100,
			Color:     color.RGBA{255, 255, 200, 255},
			Falloff:   FalloffLinear,
		},
		x: 400,
		y: 300,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.applyPointLight(buffer, scene, lwp)
	}
}

// BenchmarkLightingSystem_ApplyMultipleLights benchmarks multiple light application
// to measure the cumulative impact of the per-light allocation.
func BenchmarkLightingSystem_ApplyMultipleLights(b *testing.B) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = true
	config.MaxLights = 20
	system := NewLightingSystem(world, config)
	system.SetViewport(0, 0, 800, 600)

	// Create scene and buffer images
	scene := ebiten.NewImage(800, 600)
	buffer := ebiten.NewImage(800, 600)

	// Create 10 test lights (typical game scenario)
	lights := make([]lightWithPosition, 10)
	for i := 0; i < 10; i++ {
		lights[i] = lightWithPosition{
			light: &LightComponent{
				Enabled:   true,
				Intensity: 1.0,
				Radius:    50 + float64(i*10), // Varying radii
				Color:     color.RGBA{255, 255, 200, 255},
				Falloff:   FalloffLinear,
			},
			x: float64(100 + i*60),
			y: float64(100 + i*40),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buffer.Clear()
		for j := range lights {
			system.applyPointLight(buffer, scene, &lights[j])
		}
	}
}

// BenchmarkLightingSystem_CollectVisibleLights benchmarks the light collection hot path
// which allocates lightWithPosition structs for each visible light.
func BenchmarkLightingSystem_CollectVisibleLights(b *testing.B) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = true
	config.MaxLights = 20
	system := NewLightingSystem(world, config)
	system.SetViewport(0, 0, 800, 600)

	// Create entities with light and position components
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(100 + i*60),
			Y: float64(100 + i*40),
		})
		entity.AddComponent(&LightComponent{
			Enabled:   true,
			Intensity: 1.0,
			Radius:    50 + float64(i*10),
			Color:     color.RGBA{255, 255, 200, 255},
			Falloff:   FalloffLinear,
		})
		entities[i] = entity
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		system.CollectVisibleLights(entities)
	}
}
