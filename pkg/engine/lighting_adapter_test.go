package engine

import (
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/lighting"
)

func TestNewLightingAdapter(t *testing.T) {
	adapter := NewLightingAdapter(nil)
	if adapter == nil {
		t.Fatal("NewLightingAdapter returned nil")
	}
	if !adapter.IsEnabled() {
		t.Error("expected lighting adapter to be enabled by default")
	}
	if adapter.LightCount() != 0 {
		t.Errorf("expected 0 lights, got %d", adapter.LightCount())
	}
}

func TestLightingAdapter_SetEnabled(t *testing.T) {
	adapter := NewLightingAdapter(nil)

	adapter.SetEnabled(false)
	if adapter.IsEnabled() {
		t.Error("expected lighting adapter to be disabled")
	}

	adapter.SetEnabled(true)
	if !adapter.IsEnabled() {
		t.Error("expected lighting adapter to be enabled")
	}
}

func TestLightingAdapter_AddLight(t *testing.T) {
	adapter := NewLightingAdapter(nil)

	light := lighting.Light{
		Type:      lighting.TypePoint,
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    100.0,
		Enabled:   true,
	}

	err := adapter.AddLight(light)
	if err != nil {
		t.Fatalf("failed to add light: %v", err)
	}

	if adapter.LightCount() != 1 {
		t.Errorf("expected 1 light, got %d", adapter.LightCount())
	}
}

func TestLightingAdapter_Update(t *testing.T) {
	adapter := NewLightingAdapter(nil)

	// Create entity with light and position components
	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}

	// Add position component
	posComp := &PositionComponent{X: 100, Y: 200}
	entity.AddComponent(posComp)

	// Add light component
	lightComp := &LightComponent{
		Color:     color.RGBA{255, 255, 0, 255},
		Radius:    150.0,
		Intensity: 0.8,
		Falloff:   FalloffLinear,
		Enabled:   true,
	}
	entity.AddComponent(lightComp)

	// Update should extract lights from entities
	adapter.Update([]*Entity{entity}, 0.016)

	if adapter.LightCount() != 1 {
		t.Errorf("expected 1 light after update, got %d", adapter.LightCount())
	}
}

func TestLightingAdapter_UpdateDisabled(t *testing.T) {
	adapter := NewLightingAdapter(nil)
	adapter.SetEnabled(false)

	entity := &Entity{
		ID:         1,
		Components: make(map[string]Component),
	}
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&LightComponent{
		Color:     color.RGBA{255, 255, 255, 255},
		Radius:    100,
		Intensity: 1.0,
		Enabled:   true,
	})

	adapter.Update([]*Entity{entity}, 0.016)

	// Should not add lights when disabled
	if adapter.LightCount() != 0 {
		t.Errorf("expected 0 lights when disabled, got %d", adapter.LightCount())
	}
}

func TestLightingAdapter_ClearLights(t *testing.T) {
	adapter := NewLightingAdapter(nil)

	light := lighting.Light{
		Type:      lighting.TypePoint,
		Color:     color.RGBA{255, 255, 255, 255},
		Intensity: 1.0,
		Radius:    100.0,
		Enabled:   true,
	}

	adapter.AddLight(light)
	adapter.AddLight(light)

	if adapter.LightCount() != 2 {
		t.Fatalf("expected 2 lights, got %d", adapter.LightCount())
	}

	adapter.ClearLights()

	if adapter.LightCount() != 0 {
		t.Errorf("expected 0 lights after clear, got %d", adapter.LightCount())
	}
}
