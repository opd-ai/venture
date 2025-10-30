package engine

import (
	"image/color"
	"testing"
)

func TestNewLightingSystem(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()

	system := NewLightingSystem(world, config)

	if system.world != world {
		t.Error("World not set correctly")
	}
	if system.config != config {
		t.Error("Config not set correctly")
	}
	if system.visibleLights == nil {
		t.Error("Visible lights slice not initialized")
	}
}

func TestNewLightingSystemWithLogger(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	logger := createTestLogger()

	system := NewLightingSystemWithLogger(world, config, logger)

	if system.logger == nil {
		t.Error("Logger not set")
	}
}

func TestLightingSystem_SetViewport(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	system.SetViewport(100, 200, 800, 600)

	if system.cameraX != 100 {
		t.Errorf("cameraX = %v, want 100", system.cameraX)
	}
	if system.cameraY != 200 {
		t.Errorf("cameraY = %v, want 200", system.cameraY)
	}
	if system.viewportW != 800 {
		t.Errorf("viewportW = %v, want 800", system.viewportW)
	}
	if system.viewportH != 600 {
		t.Errorf("viewportH = %v, want 600", system.viewportH)
	}
	if !system.viewportSet {
		t.Error("viewportSet should be true")
	}
}

func TestLightingSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Create entity with light
	entity := world.CreateEntity()
	light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	light.Flickering = true
	entity.AddComponent(light)

	initialTime := light.internalTime

	// Update
	entities := []*Entity{entity}
	system.Update(entities, 0.016) // ~60 FPS

	if light.internalTime <= initialTime {
		t.Error("Internal time should have increased")
	}
}

func TestLightingSystem_UpdateDisabled(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = false
	system := NewLightingSystem(world, config)

	entity := world.CreateEntity()
	light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	entity.AddComponent(light)

	initialTime := light.internalTime

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Time should not update when disabled
	if light.internalTime != initialTime {
		t.Error("Internal time should not change when system is disabled")
	}
}

func TestLightingSystem_CollectVisibleLights(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Create entities with lights
	entity1 := world.CreateEntity()
	light1 := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	entity1.AddComponent(light1)
	entity1.AddComponent(&PositionComponent{X: 100, Y: 100})

	entity2 := world.CreateEntity()
	light2 := NewLightComponent(100, color.RGBA{255, 0, 0, 255}, 1.0)
	entity2.AddComponent(light2)
	entity2.AddComponent(&PositionComponent{X: 200, Y: 200})

	// Entity without position (should be skipped)
	entity3 := world.CreateEntity()
	light3 := NewLightComponent(100, color.RGBA{0, 255, 0, 255}, 1.0)
	entity3.AddComponent(light3)

	// Disabled light (should be skipped)
	entity4 := world.CreateEntity()
	light4 := NewLightComponent(100, color.RGBA{0, 0, 255, 255}, 1.0)
	light4.Enabled = false
	entity4.AddComponent(light4)
	entity4.AddComponent(&PositionComponent{X: 300, Y: 300})

	entities := []*Entity{entity1, entity2, entity3, entity4}
	lights := system.CollectVisibleLights(entities)

	if len(lights) != 2 {
		t.Errorf("CollectVisibleLights() returned %d lights, want 2", len(lights))
	}

	// Verify correct lights were collected
	found1, found2 := false, false
	for _, lwp := range lights {
		if lwp.x == 100 && lwp.y == 100 {
			found1 = true
		}
		if lwp.x == 200 && lwp.y == 200 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Expected lights not found in collection")
	}
}

func TestLightingSystem_CollectVisibleLightsWithCulling(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)
	system.SetViewport(0, 0, 800, 600)

	// Light in viewport
	entity1 := world.CreateEntity()
	light1 := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	entity1.AddComponent(light1)
	entity1.AddComponent(&PositionComponent{X: 400, Y: 300})

	// Light outside viewport (should be culled)
	entity2 := world.CreateEntity()
	light2 := NewLightComponent(100, color.RGBA{255, 0, 0, 255}, 1.0)
	entity2.AddComponent(light2)
	entity2.AddComponent(&PositionComponent{X: 2000, Y: 2000})

	entities := []*Entity{entity1, entity2}
	lights := system.CollectVisibleLights(entities)

	if len(lights) != 1 {
		t.Errorf("CollectVisibleLights() with culling returned %d lights, want 1", len(lights))
	}

	if lights[0].x != 400 || lights[0].y != 300 {
		t.Error("Wrong light was collected")
	}
}

func TestLightingSystem_CollectVisibleLightsMaxLimit(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.MaxLights = 3
	system := NewLightingSystem(world, config)

	// Create 5 lights (should only collect 3)
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
		entity.AddComponent(light)
		entity.AddComponent(&PositionComponent{X: float64(i * 100), Y: 100})
	}

	entities := world.GetAllEntities()
	lights := system.CollectVisibleLights(entities)

	if len(lights) != 3 {
		t.Errorf("CollectVisibleLights() returned %d lights, want 3 (max limit)", len(lights))
	}
}

func TestLightingSystem_isLightInViewport(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)
	system.SetViewport(0, 0, 800, 600)

	tests := []struct {
		name   string
		x      float64
		y      float64
		radius float64
		want   bool
	}{
		{"center of viewport", 400, 300, 100, true},
		{"left edge", 0, 300, 100, true},
		{"right edge", 800, 300, 100, true},
		{"top edge", 400, 0, 100, true},
		{"bottom edge", 400, 600, 100, true},
		{"far left (out of range)", -500, 300, 100, false},
		{"far right (out of range)", 1500, 300, 100, false},
		{"far top (out of range)", 400, -500, 100, false},
		{"far bottom (out of range)", 400, 1500, 100, false},
		{"barely in range (radius)", -50, 300, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.isLightInViewport(tt.x, tt.y, tt.radius)
			if got != tt.want {
				t.Errorf("isLightInViewport(%v, %v, %v) = %v, want %v", tt.x, tt.y, tt.radius, got, tt.want)
			}
		})
	}
}

func TestLightingSystem_calculateFalloff(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	tests := []struct {
		name         string
		dist         float64
		radius       float64
		falloffType  LightFalloffType
		wantMin      float64
		wantMax      float64
	}{
		{"linear at center", 0, 100, FalloffLinear, 1.0, 1.0},
		{"linear at half", 50, 100, FalloffLinear, 0.5, 0.5},
		{"linear at edge", 100, 100, FalloffLinear, 0.0, 0.0},
		{"quadratic at center", 0, 100, FalloffQuadratic, 1.0, 1.0},
		{"quadratic at half", 50, 100, FalloffQuadratic, 0.7, 0.8},
		{"constant within radius", 50, 100, FalloffConstant, 1.0, 1.0},
		{"constant at edge", 99, 100, FalloffConstant, 1.0, 1.0},
		{"beyond radius", 150, 100, FalloffLinear, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.calculateFalloff(tt.dist, tt.radius, tt.falloffType)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateFalloff(%v, %v, %v) = %v, want range [%v, %v]",
					tt.dist, tt.radius, tt.falloffType, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestLightingSystem_CalculateLightIntensityAt(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Create a light at (100, 100) with radius 200
	entity := world.CreateEntity()
	light := NewLightComponent(200, color.RGBA{255, 255, 255, 255}, 1.0)
	entity.AddComponent(light)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}

	tests := []struct {
		name    string
		x       float64
		y       float64
		wantMin float64
	}{
		{"at light center", 100, 100, 0.8}, // ambient + full light
		{"near light", 150, 150, 0.5},      // ambient + some light
		{"far from light", 500, 500, 0.3},  // only ambient
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intensity := system.CalculateLightIntensityAt(tt.x, tt.y, entities)
			if intensity < tt.wantMin {
				t.Errorf("CalculateLightIntensityAt(%v, %v) = %v, want >= %v", tt.x, tt.y, intensity, tt.wantMin)
			}
			if intensity > 1.0 {
				t.Errorf("CalculateLightIntensityAt(%v, %v) = %v, should be clamped to 1.0", tt.x, tt.y, intensity)
			}
		})
	}
}

func TestLightingSystem_CalculateLightIntensityAtDisabled(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = false
	system := NewLightingSystem(world, config)

	entities := []*Entity{}
	intensity := system.CalculateLightIntensityAt(0, 0, entities)

	if intensity != 1.0 {
		t.Errorf("CalculateLightIntensityAt() with disabled lighting = %v, want 1.0", intensity)
	}
}

func TestLightingSystem_SetEnabled(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	system.SetEnabled(false)
	if system.IsEnabled() {
		t.Error("System should be disabled")
	}

	system.SetEnabled(true)
	if !system.IsEnabled() {
		t.Error("System should be enabled")
	}
}

func TestLightingSystem_GetSetConfig(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	newConfig := NewLightingConfig()
	newConfig.MaxLights = 32
	newConfig.AmbientIntensity = 0.5

	system.SetConfig(newConfig)

	got := system.GetConfig()
	if got.MaxLights != 32 {
		t.Errorf("MaxLights = %v, want 32", got.MaxLights)
	}
	if got.AmbientIntensity != 0.5 {
		t.Errorf("AmbientIntensity = %v, want 0.5", got.AmbientIntensity)
	}
}

func TestLightingSystem_CollectVisibleLightsDisabled(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.Enabled = false
	system := NewLightingSystem(world, config)

	entity := world.CreateEntity()
	light := NewLightComponent(100, color.RGBA{255, 255, 255, 255}, 1.0)
	entity.AddComponent(light)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{entity}
	lights := system.CollectVisibleLights(entities)

	if len(lights) != 0 {
		t.Errorf("CollectVisibleLights() with disabled system returned %d lights, want 0", len(lights))
	}
}

func TestLightingSystem_WithAmbientLightComponent(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Create ambient light entity
	ambientEntity := world.CreateEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 255, 255}, 0.6)
	ambientEntity.AddComponent(ambient)

	// Calculate intensity (should use ambient component instead of config)
	entities := []*Entity{ambientEntity}
	intensity := system.CalculateLightIntensityAt(0, 0, entities)

	// Should be close to the ambient component's intensity
	if intensity < 0.5 || intensity > 0.7 {
		t.Errorf("CalculateLightIntensityAt() = %v, want ~0.6 (from ambient component)", intensity)
	}
}
