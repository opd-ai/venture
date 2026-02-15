package engine

import (
	"image/color"
	"math"
	"testing"
)

func TestNearbyLightTintComponentType(t *testing.T) {
	comp := NewNearbyLightTintComponent()
	if comp.Type() != "nearby_light_tint" {
		t.Errorf("expected type 'nearby_light_tint', got %q", comp.Type())
	}
}

func TestNearbyLightTintComponentDefaults(t *testing.T) {
	comp := NewNearbyLightTintComponent()
	if comp.TintR != 1.0 || comp.TintG != 1.0 || comp.TintB != 1.0 {
		t.Errorf("expected neutral tint (1,1,1), got (%f,%f,%f)", comp.TintR, comp.TintG, comp.TintB)
	}
}

func TestCalculateLightFalloff(t *testing.T) {
	tests := []struct {
		name     string
		dist     float64
		falloff  LightFalloffType
		wantMin  float64
		wantMax  float64
	}{
		{"linear center", 0.0, FalloffLinear, 1.0, 1.0},
		{"linear edge", 1.0, FalloffLinear, 0.0, 0.0},
		{"linear midpoint", 0.5, FalloffLinear, 0.49, 0.51},
		{"quadratic center", 0.0, FalloffQuadratic, 1.0, 1.0},
		{"quadratic edge", 1.0, FalloffQuadratic, 0.0, 0.0},
		{"quadratic midpoint", 0.5, FalloffQuadratic, 0.24, 0.26},
		{"inverse_square center", 0.0, FalloffInverseSquare, 1.0, 1.0},
		{"inverse_square edge", 1.0, FalloffInverseSquare, 0.0, 0.21},
		{"constant center", 0.0, FalloffConstant, 1.0, 1.0},
		{"constant midpoint", 0.5, FalloffConstant, 1.0, 1.0},
		{"constant edge", 1.0, FalloffConstant, 0.0, 0.0},
		{"beyond edge", 1.5, FalloffLinear, 0.0, 0.0},
		{"negative dist", -0.1, FalloffLinear, 1.0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateLightFalloff(tt.dist, tt.falloff)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateLightFalloff(%f, %v) = %f, want [%f, %f]",
					tt.dist, tt.falloff, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestClampLightTint(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.5},
		{0.3, 0.5},
		{0.5, 0.5},
		{0.8, 0.8},
		{1.0, 1.0},
		{1.2, 1.2},
		{1.5, 1.2},
		{2.0, 1.2},
	}
	for _, tt := range tests {
		got := clampLightTint(tt.input)
		if math.Abs(got-tt.expected) > 0.001 {
			t.Errorf("clampLightTint(%f) = %f, want %f", tt.input, got, tt.expected)
		}
	}
}

func TestBuildLightAmbientPresets(t *testing.T) {
	presets := buildLightAmbientPresets()
	expectedGenres := []string{"fantasy", "horror", "scifi", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		preset, ok := presets[genre]
		if !ok {
			t.Errorf("missing ambient preset for genre %q", genre)
			continue
		}
		if preset.R < 0.5 || preset.R > 1.2 || preset.G < 0.5 || preset.G > 1.2 || preset.B < 0.5 || preset.B > 1.2 {
			t.Errorf("genre %q preset out of range: (%f,%f,%f)", genre, preset.R, preset.G, preset.B)
		}
	}
}

func TestNearbyLightEntityTintSystemCreation(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.updateInterval <= 0 {
		t.Error("expected positive update interval")
	}
	if sys.maxLightsPerEntity <= 0 {
		t.Error("expected positive max lights per entity")
	}
}

func TestNearbyLightEntityTintSystemSetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.genreID)
	}
}

func TestNearbyLightEntityTintSystemThrottle(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entities := []*Entity{entity}

	// First call within throttle interval should not apply tints
	sys.Update(entities, 0.01)
	_, hasTint := entity.GetComponent("nearby_light_tint")
	if hasTint {
		t.Error("expected no tint component within throttle interval")
	}

	// After exceeding interval, tint should be applied
	sys.Update(entities, 0.3)
	_, hasTint = entity.GetComponent("nearby_light_tint")
	if !hasTint {
		t.Error("expected tint component after throttle interval")
	}
}

func TestNearbyLightEntityTintSystemAmbientOnly(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("horror")

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entities := []*Entity{entity}

	// Force past throttle
	sys.Update(entities, 1.0)

	comp, ok := entity.GetComponent("nearby_light_tint")
	if !ok {
		t.Fatal("expected nearby_light_tint component")
	}
	tint := comp.(*NearbyLightTintComponent)

	// Horror ambient should darken the entity
	if tint.TintR >= 1.0 || tint.TintG >= 1.0 {
		t.Errorf("expected horror ambient to darken, got (%f,%f,%f)", tint.TintR, tint.TintG, tint.TintB)
	}
}

func TestNearbyLightEntityTintSystemWithLight(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create a warm torch light at (100, 100)
	lightEntity := world.CreateEntity()
	lightEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	lightComp := NewLightComponent(200, color.RGBA{R: 255, G: 180, B: 80, A: 255}, 1.0)
	lightComp.Enabled = true
	lightEntity.AddComponent(lightComp)

	// Create an entity near the light
	nearEntity := world.CreateEntity()
	nearEntity.AddComponent(&EbitenSprite{})
	nearEntity.AddComponent(&PositionComponent{X: 120, Y: 100})

	// Create an entity far from the light
	farEntity := world.CreateEntity()
	farEntity.AddComponent(&EbitenSprite{})
	farEntity.AddComponent(&PositionComponent{X: 500, Y: 500})

	entities := []*Entity{lightEntity, nearEntity, farEntity}

	sys.Update(entities, 1.0)

	// Near entity should have brighter tint than far entity
	nearComp, _ := nearEntity.GetComponent("nearby_light_tint")
	farComp, _ := farEntity.GetComponent("nearby_light_tint")

	if nearComp == nil || farComp == nil {
		t.Fatal("expected tint components on both entities")
	}

	nearTint := nearComp.(*NearbyLightTintComponent)
	farTint := farComp.(*NearbyLightTintComponent)

	// Near entity should be brighter (higher R from warm light)
	if nearTint.TintR <= farTint.TintR {
		t.Errorf("near entity TintR (%f) should be > far entity TintR (%f)",
			nearTint.TintR, farTint.TintR)
	}
}

func TestNearbyLightEntityTintSystemDisabledLight(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create a disabled light
	lightEntity := world.CreateEntity()
	lightEntity.AddComponent(&PositionComponent{X: 100, Y: 100})
	lightComp := NewLightComponent(200, color.RGBA{R: 255, G: 180, B: 80, A: 255}, 1.0)
	lightComp.Enabled = false
	lightEntity.AddComponent(lightComp)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&PositionComponent{X: 120, Y: 100})

	entities := []*Entity{lightEntity, entity}
	sys.Update(entities, 1.0)

	// Should get ambient-only tint (disabled light ignored)
	comp, _ := entity.GetComponent("nearby_light_tint")
	if comp == nil {
		t.Fatal("expected tint component")
	}
	tint := comp.(*NearbyLightTintComponent)
	ambient := sys.getAmbientPreset()
	if math.Abs(tint.TintR-ambient.R) > 0.001 {
		t.Errorf("expected ambient-only TintR (%f), got %f", ambient.R, tint.TintR)
	}
}

func TestNearbyLightEntityTintSystemNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)

	// Entity without sprite should not get tint component
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entities := []*Entity{entity}

	sys.Update(entities, 1.0)

	_, hasTint := entity.GetComponent("nearby_light_tint")
	if hasTint {
		t.Error("entity without sprite should not get tint component")
	}
}

func TestNearbyLightEntityTintSystemMultipleLights(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("fantasy")

	// Create two nearby lights
	light1 := world.CreateEntity()
	light1.AddComponent(&PositionComponent{X: 100, Y: 100})
	lc1 := NewLightComponent(200, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 1.0)
	lc1.Enabled = true
	light1.AddComponent(lc1)

	light2 := world.CreateEntity()
	light2.AddComponent(&PositionComponent{X: 120, Y: 100})
	lc2 := NewLightComponent(200, color.RGBA{R: 0, G: 0, B: 255, A: 255}, 1.0)
	lc2.Enabled = true
	light2.AddComponent(lc2)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&PositionComponent{X: 110, Y: 100})

	entities := []*Entity{light1, light2, entity}
	sys.Update(entities, 1.0)

	comp, _ := entity.GetComponent("nearby_light_tint")
	if comp == nil {
		t.Fatal("expected tint component")
	}
	tint := comp.(*NearbyLightTintComponent)

	// With red + blue lights, both R and B should be boosted above ambient
	ambient := sys.getAmbientPreset()
	if tint.TintR <= ambient.R {
		t.Errorf("expected R boosted above ambient (%f), got %f", ambient.R, tint.TintR)
	}
	if tint.TintB <= ambient.B {
		t.Errorf("expected B boosted above ambient (%f), got %f", ambient.B, tint.TintB)
	}
}

func TestGetAmbientPresetUnknownGenre(t *testing.T) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("unknown_genre")
	preset := sys.getAmbientPreset()
	if preset.R != 1.0 || preset.G != 1.0 || preset.B != 1.0 {
		t.Errorf("expected neutral preset for unknown genre, got (%f,%f,%f)", preset.R, preset.G, preset.B)
	}
}

func BenchmarkNearbyLightEntityTintSystem(b *testing.B) {
	world := NewWorld()
	sys := NewNearbyLightEntityTintSystem(world, 42)
	sys.SetGenre("fantasy")

	// 3 lights + 100 entities
	entities := make([]*Entity, 0, 103)
	for i := 0; i < 3; i++ {
		light := world.CreateEntity()
		light.AddComponent(&PositionComponent{X: float64(i * 200), Y: 100})
		lc := NewLightComponent(200, color.RGBA{R: 255, G: 200, B: 100, A: 255}, 1.0)
		lc.Enabled = true
		light.AddComponent(lc)
		entities = append(entities, light)
	}
	for i := 0; i < 100; i++ {
		e := world.CreateEntity()
		e.AddComponent(&EbitenSprite{})
		e.AddComponent(&PositionComponent{X: float64(i * 8), Y: 100})
		entities = append(entities, e)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = sys.updateInterval // Force update
		sys.Update(entities, 0.016)
	}
}
