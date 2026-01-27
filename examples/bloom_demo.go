// Package main demonstrates bloom effect integration in the lighting system.
// This example shows how to enable and configure bloom for magical lights.
package main

import (
	"fmt"
	"image/color"

	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	fmt.Println("Bloom Effect Integration Demo")
	fmt.Println("==============================")

	// Create world
	world := engine.NewWorld()

	// Configure lighting with bloom enabled
	config := engine.NewLightingConfig()
	config.EnableBloom = true
	config.BloomThreshold = 0.7  // Only bright lights bloom
	config.BloomIntensity = 1.5  // Strong glow effect
	config.BloomRadius = 12      // Medium spread distance

	fmt.Printf("\nLighting Configuration:\n")
	fmt.Printf("  Bloom Enabled: %v\n", config.EnableBloom)
	fmt.Printf("  Bloom Threshold: %.2f (0.0-1.0)\n", config.BloomThreshold)
	fmt.Printf("  Bloom Intensity: %.2f\n", config.BloomIntensity)
	fmt.Printf("  Bloom Radius: %d pixels\n", config.BloomRadius)

	// Create lighting system
	system := engine.NewLightingSystem(world, config)

	fmt.Printf("\nLighting system created successfully\n")

	// Example 1: Magical spell light (will bloom)
	magicalEntity := world.CreateEntity()
	magicalLight := engine.NewSpellLight(100, color.RGBA{R: 255, G: 100, B: 255, A: 255})
	magicalLight.Intensity = 2.0 // Very bright = will bloom
	magicalPos := &engine.PositionComponent{X: 400, Y: 300}
	magicalEntity.AddComponent(magicalLight)
	magicalEntity.AddComponent(magicalPos)

	fmt.Printf("\nCreated magical spell light:\n")
	fmt.Printf("  Position: (%.0f, %.0f)\n", magicalPos.X, magicalPos.Y)
	fmt.Printf("  Color: RGB(%d, %d, %d)\n", magicalLight.Color.R, magicalLight.Color.G, magicalLight.Color.B)
	fmt.Printf("  Intensity: %.1f (bright enough to bloom)\n", magicalLight.Intensity)
	fmt.Printf("  Radius: %.0f pixels\n", magicalLight.Radius)

	// Example 2: Regular torch light (may not bloom)
	torchEntity := world.CreateEntity()
	torchLight := engine.NewTorchLight(80)
	torchLight.Intensity = 1.0 // Normal brightness = may not bloom
	torchPos := &engine.PositionComponent{X: 200, Y: 200}
	torchEntity.AddComponent(torchLight)
	torchEntity.AddComponent(torchPos)

	fmt.Printf("\nCreated torch light:\n")
	fmt.Printf("  Position: (%.0f, %.0f)\n", torchPos.X, torchPos.Y)
	fmt.Printf("  Color: RGB(%d, %d, %d)\n", torchLight.Color.R, torchLight.Color.G, torchLight.Color.B)
	fmt.Printf("  Intensity: %.1f (normal brightness)\n", torchLight.Intensity)
	fmt.Printf("  Radius: %.0f pixels\n", torchLight.Radius)

	// Collect lights
	system.SetViewport(0, 0, 800, 600)
	entities := []*engine.Entity{magicalEntity, torchEntity}
	lights := system.CollectVisibleLights(entities)

	fmt.Printf("\nCollected %d visible lights\n", len(lights))

	fmt.Println("\n✅ Bloom Effect Integration Successful!")
	fmt.Println("\nExpected Visual Result:")
	fmt.Println("  - Magical spell: Bright purple glow spreading beyond light radius")
	fmt.Println("  - Torch: Warm orange light with minimal bloom")
	fmt.Println("  - Bloom adds soft halo around bright lights for enhanced atmosphere")

	fmt.Println("\nUsage in your game:")
	fmt.Println("  1. Create LightingConfig with EnableBloom = true")
	fmt.Println("  2. Adjust BloomThreshold (0.0-1.0) for brightness cutoff")
	fmt.Println("  3. Adjust BloomIntensity (0.5-2.0) for glow strength")
	fmt.Println("  4. Adjust BloomRadius (8-20) for spread distance")
	fmt.Println("  5. Bright lights (intensity > threshold) automatically bloom")
}
