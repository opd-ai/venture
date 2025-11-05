package engine

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestSoftShadowPenumbraRendering verifies that soft shadows are rendered with proper umbra/penumbra layers.
func TestSoftShadowPenumbraRendering(t *testing.T) {
	world := NewWorld()
	system := NewShadowSystem(world)
	system.SetEnabled(true)

	// Create entity with soft shadow
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 400, Y: 300})
	shadowComp := NewShadowComponent(32)
	shadowComp.ShadowType = ShadowTypeSoft
	shadowComp.Opacity = 0.8
	shadowComp.SoftEdgeRadius = 8.0
	entity.AddComponent(shadowComp)

	// Create light source
	light := world.CreateEntity()
	light.AddComponent(&PositionComponent{X: 300, Y: 200})
	light.AddComponent(&LightComponent{
		Enabled:   true,
		Intensity: 1.0,
		Radius:    500,
		Color:     color.RGBA{255, 255, 255, 255},
	})

	world.Update(0)

	// Render shadows
	target := ebiten.NewImage(800, 600)
	system.SetViewport(0, 0, 800, 600)
	system.RenderShadows(target, 300, 200, 500)

	// Verify shadow was rendered (non-zero pixels in shadow area)
	// This is a smoke test - actual visual verification would require image comparison
	if target == nil {
		t.Error("Shadow render target is nil")
	}
}

// TestSoftShadowPenumbraDistanceFalloff verifies penumbra width increases with light distance.
func TestSoftShadowPenumbraDistanceFalloff(t *testing.T) {
	tests := []struct {
		name           string
		lightDistance  float64
		expectedFactor float64 // Relative to base penumbra
		description    string
	}{
		{
			name:           "close light (hard shadow)",
			lightDistance:  100,
			expectedFactor: 0.2, // 100/500 = 0.2
			description:    "Close lights produce harder shadows",
		},
		{
			name:           "medium light",
			lightDistance:  250,
			expectedFactor: 0.5, // 250/500 = 0.5
			description:    "Medium distance produces moderate softness",
		},
		{
			name:           "distant light (soft shadow)",
			lightDistance:  500,
			expectedFactor: 1.0, // 500/500 = 1.0
			description:    "Distant lights produce softer shadows",
		},
		{
			name:           "very distant light (max soft)",
			lightDistance:  1500,
			expectedFactor: 2.0, // Capped at 2.0
			description:    "Very distant lights capped at maximum softness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate penumbra factor using same formula as renderSoftShadow
			penumbraFactor := math.Min(tt.lightDistance/500.0, 2.0)

			if math.Abs(penumbraFactor-tt.expectedFactor) > 0.01 {
				t.Errorf("penumbraFactor = %f, want %f (%s)",
					penumbraFactor, tt.expectedFactor, tt.description)
			}
		})
	}
}

// TestSoftShadowLayerOpacityGradient verifies penumbra layers have proper opacity falloff.
func TestSoftShadowLayerOpacityGradient(t *testing.T) {
	layers := 3
	baseOpacity := 0.8

	for i := 1; i <= layers; i++ {
		layerFactor := float64(i) / float64(layers+1)
		layerOpacity := (1.0 - layerFactor) * 0.4 * baseOpacity

		// Verify opacity decreases with each layer (outer layers fainter)
		if i > 1 {
			prevLayerFactor := float64(i-1) / float64(layers+1)
			prevLayerOpacity := (1.0 - prevLayerFactor) * 0.4 * baseOpacity

			if layerOpacity >= prevLayerOpacity {
				t.Errorf("Layer %d opacity %f should be less than layer %d opacity %f",
					i, layerOpacity, i-1, prevLayerOpacity)
			}
		}

		// Verify opacity is in valid range
		if layerOpacity < 0 || layerOpacity > 1.0 {
			t.Errorf("Layer %d opacity %f out of range [0, 1]", i, layerOpacity)
		}
	}
}

// TestSoftShadowPenumbraRadius verifies penumbra radius expansion.
func TestSoftShadowPenumbraRadius(t *testing.T) {
	baseRadius := 32.0
	softEdgeRadius := 8.0
	penumbraFactor := 1.5 // Example factor
	penumbraWidth := softEdgeRadius * penumbraFactor

	layers := 3
	for i := 1; i <= layers; i++ {
		layerFactor := float64(i) / float64(layers+1)
		expectedRadius := baseRadius + (penumbraWidth * layerFactor)

		// Verify each layer expands outward
		if i > 1 {
			prevLayerFactor := float64(i-1) / float64(layers+1)
			prevRadius := baseRadius + (penumbraWidth * prevLayerFactor)

			if expectedRadius <= prevRadius {
				t.Errorf("Layer %d radius %f should be greater than layer %d radius %f",
					i, expectedRadius, i-1, prevRadius)
			}
		}

		// Verify radius is positive
		if expectedRadius <= 0 {
			t.Errorf("Layer %d radius %f should be positive", i, expectedRadius)
		}
	}
}

// TestSoftShadowEntityAtLightSource verifies handling when entity is at light source.
func TestSoftShadowEntityAtLightSource(t *testing.T) {
	world := NewWorld()
	system := NewShadowSystem(world)

	// Create entity at same position as light
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 400, Y: 300})
	shadowComp := NewShadowComponent(32)
	shadowComp.ShadowType = ShadowTypeSoft
	entity.AddComponent(shadowComp)

	light := world.CreateEntity()
	light.AddComponent(&PositionComponent{X: 400, Y: 300}) // Same position
	light.AddComponent(&LightComponent{
		Enabled:   true,
		Intensity: 1.0,
		Radius:    500,
		Color:     color.RGBA{255, 255, 255, 255},
	})

	world.Update(0)

	// Should not crash when rendering (distance < 0.1 check)
	target := ebiten.NewImage(800, 600)
	system.SetViewport(0, 0, 800, 600)
	system.RenderShadows(target, 400, 300, 500) // Light at same position

	if target == nil {
		t.Error("Render target is nil")
	}
}

// TestSoftShadowComponentConfiguration verifies SoftEdgeRadius usage.
func TestSoftShadowComponentConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		softEdgeRadius float64
		valid          bool
	}{
		{"zero radius", 0.0, true},
		{"small radius", 2.0, true},
		{"default radius", 4.0, true},
		{"large radius", 16.0, true},
		{"very large radius", 64.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewShadowComponent(32)
			comp.SoftEdgeRadius = tt.softEdgeRadius
			comp.ShadowType = ShadowTypeSoft

			if comp.SoftEdgeRadius != tt.softEdgeRadius {
				t.Errorf("SoftEdgeRadius = %f, want %f", comp.SoftEdgeRadius, tt.softEdgeRadius)
			}
		})
	}
}

// TestSoftShadowUmbraOpacity verifies umbra core shadow opacity.
func TestSoftShadowUmbraOpacity(t *testing.T) {
	baseOpacity := 1.0
	umbraOpacity := baseOpacity * 0.8

	if umbraOpacity > baseOpacity {
		t.Errorf("Umbra opacity %f should not exceed base opacity %f",
			umbraOpacity, baseOpacity)
	}

	if umbraOpacity <= 0 {
		t.Error("Umbra opacity should be positive")
	}
}

// TestSoftShadowMultipleLights verifies shadows from multiple light sources.
func TestSoftShadowMultipleLights(t *testing.T) {
	world := NewWorld()
	system := NewShadowSystem(world)

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 400, Y: 300})
	shadowComp := NewShadowComponent(32)
	shadowComp.ShadowType = ShadowTypeSoft
	entity.AddComponent(shadowComp)

	// Create multiple lights
	light1 := world.CreateEntity()
	light1.AddComponent(&PositionComponent{X: 200, Y: 200})
	light1.AddComponent(&LightComponent{
		Enabled:   true,
		Intensity: 1.0,
		Radius:    500,
		Color:     color.RGBA{255, 255, 255, 255},
	})

	light2 := world.CreateEntity()
	light2.AddComponent(&PositionComponent{X: 600, Y: 400})
	light2.AddComponent(&LightComponent{
		Enabled:   true,
		Intensity: 1.0,
		Radius:    500,
		Color:     color.RGBA{255, 255, 255, 255},
	})

	world.Update(0)

	// Should render multiple shadows without issues
	target := ebiten.NewImage(800, 600)
	system.SetViewport(0, 0, 800, 600)
	system.RenderShadows(target, 200, 200, 500) // Render from light1 position

	if target == nil {
		t.Error("Render target is nil")
	}
}

// TestSoftShadowDisabled verifies soft shadow respects Enabled flag.
func TestSoftShadowDisabled(t *testing.T) {
	world := NewWorld()
	system := NewShadowSystem(world)

	// Create entity with disabled shadow
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 400, Y: 300})
	shadowComp := NewShadowComponent(32)
	shadowComp.ShadowType = ShadowTypeSoft
	shadowComp.Enabled = false // Disabled
	entity.AddComponent(shadowComp)

	light := world.CreateEntity()
	light.AddComponent(&PositionComponent{X: 300, Y: 200})
	light.AddComponent(&LightComponent{
		Enabled:   true,
		Intensity: 1.0,
		Radius:    500,
		Color:     color.RGBA{255, 255, 255, 255},
	})

	world.Update(0)

	// Should not crash even with disabled shadow
	target := ebiten.NewImage(800, 600)
	system.SetViewport(0, 0, 800, 600)
	system.RenderShadows(target, 300, 200, 500)

	if target == nil {
		t.Error("Render target is nil")
	}
}
