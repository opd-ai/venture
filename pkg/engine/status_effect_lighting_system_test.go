package engine

import (
	"image/color"
	"testing"
)

func TestNewStatusEffectLightingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewStatusEffectLightingSystem returned nil")
	}

	if sys.world != world {
		t.Error("World reference not set")
	}

	if sys.rng == nil {
		t.Error("RNG not initialized")
	}

	if len(sys.effectConfigs) == 0 {
		t.Error("Default effect configs not initialized")
	}

	// Verify default configs exist for common effects
	expectedEffects := []string{"burn", "poison", "frost", "regeneration", "shock", "stun"}
	for _, effect := range expectedEffects {
		if _, exists := sys.effectConfigs[effect]; !exists {
			t.Errorf("Missing default config for effect: %s", effect)
		}
	}
}

func TestStatusEffectLightingSystem_SetEffectConfig(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	customConfig := StatusEffectLightConfig{
		Color:      color.RGBA{R: 255, G: 0, B: 255, A: 255},
		Radius:     100.0,
		Intensity:  0.9,
		Flickering: true,
		FlickerAmt: 0.4,
		FlickerSpd: 10.0,
	}

	sys.SetEffectConfig("custom_effect", customConfig)

	config, exists := sys.effectConfigs["custom_effect"]
	if !exists {
		t.Fatal("Custom config not added")
	}

	if config.Radius != 100.0 {
		t.Errorf("Radius mismatch: got %f, want 100.0", config.Radius)
	}

	if config.Intensity != 0.9 {
		t.Errorf("Intensity mismatch: got %f, want 0.9", config.Intensity)
	}
}

func TestStatusEffectLightingSystem_Update_NoStatusEffects(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	// Create entity with position but no status effects
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should not have any lights added
	if sys.GetActiveEffectLightCount() != 0 {
		t.Errorf("Expected 0 active lights, got %d", sys.GetActiveEffectLightCount())
	}

	// Entity should not have light component
	if _, hasLight := entity.GetComponent("light"); hasLight {
		t.Error("Entity without status effects should not have light")
	}
}

func TestStatusEffectLightingSystem_Update_WithBurnEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "burn",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should have one active light
	if sys.GetActiveEffectLightCount() != 1 {
		t.Errorf("Expected 1 active light, got %d", sys.GetActiveEffectLightCount())
	}

	// Entity should have light component
	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		t.Fatal("Entity with burn effect should have light")
	}

	light := lightComp.(*LightComponent)
	burnConfig := sys.effectConfigs["burn"]

	if light.Radius != burnConfig.Radius {
		t.Errorf("Light radius mismatch: got %f, want %f", light.Radius, burnConfig.Radius)
	}

	if !light.Flickering {
		t.Error("Burn light should be flickering")
	}

	// Verify orange-ish color (burn effect)
	if light.Color.R < 200 || light.Color.G > 150 {
		t.Errorf("Burn light color doesn't match expected orange: %v", light.Color)
	}
}

func TestStatusEffectLightingSystem_Update_WithPoisonEffect(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poison",
		Duration:   5.0,
		Magnitude:  5.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	lightComp, hasLight := entity.GetComponent("light")
	if !hasLight {
		t.Fatal("Entity with poison effect should have light")
	}

	light := lightComp.(*LightComponent)

	// Poison should pulse, not flicker
	if light.Flickering {
		t.Error("Poison light should not flicker")
	}

	if !light.Pulsing {
		t.Error("Poison light should pulse")
	}

	// Verify greenish color
	if light.Color.G < 150 {
		t.Errorf("Poison light should be green-ish: %v", light.Color)
	}
}

func TestStatusEffectLightingSystem_Update_EffectExpires(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	effect := &StatusEffectComponent{
		EffectType: "burn",
		Duration:   5.0,
		Magnitude:  10.0,
	}
	entity.AddComponent(effect)

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Verify light was added
	if sys.GetActiveEffectLightCount() != 1 {
		t.Errorf("Expected 1 active light after adding effect, got %d", sys.GetActiveEffectLightCount())
	}

	// Remove the status effect (simulating expiration)
	entity.RemoveComponent("status_effect")

	sys.Update(entities, 0.016)

	// Light should be removed
	if sys.GetActiveEffectLightCount() != 0 {
		t.Errorf("Expected 0 active lights after effect expired, got %d", sys.GetActiveEffectLightCount())
	}

	if _, hasLight := entity.GetComponent("light"); hasLight {
		t.Error("Light should be removed when status effect expires")
	}
}

func TestStatusEffectLightingSystem_Update_EffectTypeChange(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "burn",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Get initial light color
	lightComp, _ := entity.GetComponent("light")
	initialColor := lightComp.(*LightComponent).Color

	// Change effect type to frost
	entity.RemoveComponent("status_effect")
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "frost",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	sys.Update(entities, 0.016)

	// Light should now be blue-ish
	lightComp, _ = entity.GetComponent("light")
	newColor := lightComp.(*LightComponent).Color

	// Color should have changed (frost is blue, burn is orange)
	if newColor.B <= initialColor.B {
		t.Error("Light color should change when effect type changes (burn->frost)")
	}
}

func TestStatusEffectLightingSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	// Create entity with status effect but no position
	entity := world.CreateEntity()
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "burn",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should not add light to entity without position
	if sys.GetActiveEffectLightCount() != 0 {
		t.Errorf("Should not create light for entity without position")
	}
}

func TestStatusEffectLightingSystem_Update_PriorityOrder(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Add multiple effects - burn should take priority over poison
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "poison",
		Duration:   5.0,
		Magnitude:  10.0,
	})
	entity.AddComponent(&StatusEffectComponent{
		EffectType: "burn",
		Duration:   5.0,
		Magnitude:  10.0,
	})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Only one light should be added (for the highest priority effect)
	if sys.GetActiveEffectLightCount() != 1 {
		t.Errorf("Expected 1 active light for multiple effects, got %d", sys.GetActiveEffectLightCount())
	}

	// Light should match burn config (higher priority than poison)
	lightComp, _ := entity.GetComponent("light")
	light := lightComp.(*LightComponent)
	burnConfig := sys.effectConfigs["burn"]

	if light.Flickering != burnConfig.Flickering {
		t.Error("Light should match burn effect (flickering) since burn has higher priority")
	}
}

func TestStatusEffectLightingSystem_ScaleIntensityByMagnitude(t *testing.T) {
	tests := []struct {
		name          string
		baseIntensity float64
		magnitude     float64
		wantMin       float64
		wantMax       float64
	}{
		{"low magnitude", 0.5, 5.0, 0.25, 0.5},          // Below base, clamped
		{"base magnitude", 0.5, 10.0, 0.49, 0.51},       // 1.0x multiplier
		{"medium magnitude", 0.5, 50.0, 0.7, 0.8},       // ~1.5x multiplier
		{"high magnitude", 0.5, 100.0, 0.95, 1.05},      // 2.0x multiplier
		{"very high magnitude", 0.5, 200.0, 0.95, 1.05}, // Clamped to 2.0x
	}

	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.ScaleIntensityByMagnitude(tt.baseIntensity, tt.magnitude)
			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("ScaleIntensityByMagnitude(%f, %f) = %f, want between %f and %f",
					tt.baseIntensity, tt.magnitude, result, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestStatusEffectLightingSystem_DefaultConfigs(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	tests := []struct {
		effectType    string
		expectFlicker bool
		expectPulse   bool
	}{
		{"burn", true, false},
		{"shock", true, false},
		{"poison", false, true},
		{"frost", false, true},
		{"regeneration", false, true},
		{"stun", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.effectType, func(t *testing.T) {
			config, exists := sys.effectConfigs[tt.effectType]
			if !exists {
				t.Fatalf("Config for %s not found", tt.effectType)
			}

			if config.Flickering != tt.expectFlicker {
				t.Errorf("%s: Flickering = %v, want %v", tt.effectType, config.Flickering, tt.expectFlicker)
			}

			if config.Pulsing != tt.expectPulse {
				t.Errorf("%s: Pulsing = %v, want %v", tt.effectType, config.Pulsing, tt.expectPulse)
			}

			if config.Radius <= 0 {
				t.Errorf("%s: Radius should be positive, got %f", tt.effectType, config.Radius)
			}

			if config.Intensity <= 0 || config.Intensity > 1.0 {
				t.Errorf("%s: Intensity should be in (0, 1], got %f", tt.effectType, config.Intensity)
			}
		})
	}
}

func TestStatusEffectLightingSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	// Create multiple entities with different effects
	entity1 := world.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity1.AddComponent(&StatusEffectComponent{EffectType: "burn", Duration: 5.0})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 200, Y: 200})
	entity2.AddComponent(&StatusEffectComponent{EffectType: "frost", Duration: 5.0})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 300, Y: 300})
	// No status effect

	entities := []*Entity{entity1, entity2, entity3}
	sys.Update(entities, 0.016)

	// Should have exactly 2 active lights
	if sys.GetActiveEffectLightCount() != 2 {
		t.Errorf("Expected 2 active lights, got %d", sys.GetActiveEffectLightCount())
	}

	// Verify each affected entity has a light
	if _, hasLight := entity1.GetComponent("light"); !hasLight {
		t.Error("Entity1 (burn) should have light")
	}

	if _, hasLight := entity2.GetComponent("light"); !hasLight {
		t.Error("Entity2 (frost) should have light")
	}

	if _, hasLight := entity3.GetComponent("light"); hasLight {
		t.Error("Entity3 (no effect) should not have light")
	}
}

func BenchmarkStatusEffectLightingSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewStatusEffectLightingSystem(world, 12345)

	// Create 100 entities, half with status effects
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		if i%2 == 0 {
			entity.AddComponent(&StatusEffectComponent{
				EffectType: "burn",
				Duration:   5.0,
				Magnitude:  10.0,
			})
		}
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
