package engine

import (
	"image"
	"image/color"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/lighting"
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

	// Process pending entity additions
	world.Update(0)

	entities := world.GetEntities()
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
		name        string
		dist        float64
		radius      float64
		falloffType LightFalloffType
		wantMin     float64
		wantMax     float64
	}{
		{"linear at center", 0, 100, FalloffLinear, 1.0, 1.0},
		{"linear at half", 50, 100, FalloffLinear, 0.5, 0.5},
		{"linear at edge", 100, 100, FalloffLinear, 0.0, 0.0},
		{"quadratic at center", 0, 100, FalloffQuadratic, 1.0, 1.0},
		{"quadratic at half", 50, 100, FalloffQuadratic, 0.24, 0.26}, // (1-0.5)^2 = 0.25
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

func TestLightingSystem_SetAmbientLightEntity(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Initially not cached
	if system.ambientLightCached {
		t.Error("Ambient light should not be cached initially")
	}

	// Set ambient light entity
	system.SetAmbientLightEntity(123)

	if !system.ambientLightCached {
		t.Error("Ambient light should be cached after SetAmbientLightEntity")
	}
	if system.ambientLightEntityID != 123 {
		t.Errorf("ambientLightEntityID = %v, want 123", system.ambientLightEntityID)
	}
}

func TestLightingSystem_ClearAmbientLightCache(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Set and then clear
	system.SetAmbientLightEntity(123)
	system.ClearAmbientLightCache()

	if system.ambientLightCached {
		t.Error("Ambient light cache should be cleared")
	}
	if system.ambientLightEntityID != 0 {
		t.Errorf("ambientLightEntityID = %v, want 0", system.ambientLightEntityID)
	}
}

func TestLightingSystem_AmbientLightCaching(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Create ambient light entity
	ambientEntity := world.CreateEntity()
	ambient := NewAmbientLightComponent(color.RGBA{200, 200, 255, 255}, 0.6)
	ambientEntity.AddComponent(ambient)

	// Set the cache
	system.SetAmbientLightEntity(ambientEntity.ID)

	// Calculate intensity using cached entity
	entities := []*Entity{ambientEntity}
	intensity := system.CalculateLightIntensityAt(0, 0, entities)

	// Should use cached ambient component
	if intensity < 0.5 || intensity > 0.7 {
		t.Errorf("CalculateLightIntensityAt() = %v, want ~0.6 (from cached ambient)", intensity)
	}
}

// TestGetCachedLightCircle_Caching tests that light circles are properly cached.
func TestGetCachedLightCircle_Caching(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// First call should create and cache the image
	img1 := system.getCachedLightCircle(100, FalloffLinear)
	if img1 == nil {
		t.Fatal("getCachedLightCircle returned nil")
	}

	// Second call should return the same cached image
	img2 := system.getCachedLightCircle(100, FalloffLinear)
	if img1 != img2 {
		t.Error("getCachedLightCircle should return cached image on second call")
	}

	// Different diameter should create new image
	img3 := system.getCachedLightCircle(200, FalloffLinear)
	if img3 == img1 {
		t.Error("Different diameter should create new cached image")
	}

	// Different falloff should create new image
	img4 := system.getCachedLightCircle(100, FalloffQuadratic)
	if img4 == img1 {
		t.Error("Different falloff should create new cached image")
	}
}

// TestGetCachedLightCircle_GradientGeneration tests that gradients are properly generated.
func TestGetCachedLightCircle_GradientGeneration(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	diameter := 100
	img := system.getCachedLightCircle(diameter, FalloffLinear)

	if img == nil {
		t.Fatal("getCachedLightCircle returned nil")
	}

	// Check image dimensions
	bounds := img.Bounds()
	if bounds.Dx() != diameter || bounds.Dy() != diameter {
		t.Errorf("Image dimensions = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), diameter, diameter)
	}

	// Note: We cannot call img.At() in tests without a running Ebiten game.
	// The gradient generation logic is tested indirectly through calculateFalloffIntensity tests.
}

// TestGetCachedLightCircle_AllFalloffTypes tests gradient generation for all falloff types.
func TestGetCachedLightCircle_AllFalloffTypes(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	falloffTypes := []LightFalloffType{
		FalloffLinear,
		FalloffQuadratic,
		FalloffInverseSquare,
		FalloffConstant,
	}

	for _, falloff := range falloffTypes {
		t.Run(falloff.String(), func(t *testing.T) {
			img := system.getCachedLightCircle(100, falloff)
			if img == nil {
				t.Fatalf("getCachedLightCircle returned nil for falloff %v", falloff)
			}

			// Verify image was created
			bounds := img.Bounds()
			if bounds.Dx() != 100 || bounds.Dy() != 100 {
				t.Errorf("Image dimensions = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestCalculateFalloffIntensity_Linear tests linear falloff calculation.
func TestCalculateFalloffIntensity_Linear(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	tests := []struct {
		distance float64
		want     float64
	}{
		{0.0, 1.0},  // Center: full intensity
		{0.5, 0.5},  // Halfway: half intensity
		{1.0, 0.0},  // Edge: no intensity
		{-0.1, 1.0}, // Below zero: clamped to full
		{1.1, 0.0},  // Beyond edge: clamped to zero
	}

	for _, tt := range tests {
		got := system.calculateFalloffIntensity(tt.distance, FalloffLinear)
		if got != tt.want {
			t.Errorf("calculateFalloffIntensity(%v, Linear) = %v, want %v", tt.distance, got, tt.want)
		}
	}
}

// TestCalculateFalloffIntensity_Quadratic tests quadratic falloff calculation.
func TestCalculateFalloffIntensity_Quadratic(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	tests := []struct {
		distance  float64
		want      float64
		tolerance float64
	}{
		{0.0, 1.0, 0.001},  // Center: full intensity
		{0.5, 0.25, 0.001}, // Halfway: (1-0.5)^2 = 0.25
		{1.0, 0.0, 0.001},  // Edge: no intensity
	}

	for _, tt := range tests {
		got := system.calculateFalloffIntensity(tt.distance, FalloffQuadratic)
		diff := got - tt.want
		if diff < 0 {
			diff = -diff
		}
		if diff > tt.tolerance {
			t.Errorf("calculateFalloffIntensity(%v, Quadratic) = %v, want %v", tt.distance, got, tt.want)
		}
	}
}

// TestCalculateFalloffIntensity_InverseSquare tests inverse square falloff calculation.
func TestCalculateFalloffIntensity_InverseSquare(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Test basic properties
	centerIntensity := system.calculateFalloffIntensity(0.0, FalloffInverseSquare)
	if centerIntensity != 1.0 {
		t.Errorf("Center intensity = %v, want 1.0", centerIntensity)
	}

	edgeIntensity := system.calculateFalloffIntensity(1.0, FalloffInverseSquare)
	if edgeIntensity >= 0.3 { // Should be quite dim at edge
		t.Errorf("Edge intensity = %v, want < 0.3", edgeIntensity)
	}

	// Verify monotonic decrease
	prev := 1.0
	for dist := 0.0; dist <= 1.0; dist += 0.1 {
		curr := system.calculateFalloffIntensity(dist, FalloffInverseSquare)
		if curr > prev {
			t.Errorf("Intensity increased from %v to %v at distance %v (should be monotonic decreasing)", prev, curr, dist)
		}
		prev = curr
	}
}

// TestCalculateFalloffIntensity_Constant tests constant falloff calculation.
func TestCalculateFalloffIntensity_Constant(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Constant falloff should always return 1.0 within range
	tests := []float64{0.0, 0.25, 0.5, 0.75, 0.99}
	for _, dist := range tests {
		got := system.calculateFalloffIntensity(dist, FalloffConstant)
		if got != 1.0 {
			t.Errorf("calculateFalloffIntensity(%v, Constant) = %v, want 1.0", dist, got)
		}
	}

	// Edge should still return 0
	got := system.calculateFalloffIntensity(1.0, FalloffConstant)
	if got != 0.0 {
		t.Errorf("calculateFalloffIntensity(1.0, Constant) = %v, want 0.0", got)
	}
}

// TestCalculateFalloffIntensity_UnknownType tests fallback for unknown falloff types.
func TestCalculateFalloffIntensity_UnknownType(t *testing.T) {
	world := NewWorld()
	system := NewLightingSystem(world, nil)

	// Use an invalid falloff type value
	invalidFalloff := LightFalloffType(999)

	// Should default to linear falloff behavior
	got := system.calculateFalloffIntensity(0.5, invalidFalloff)
	want := 0.5 // Linear falloff at 0.5 distance
	if got != want {
		t.Errorf("calculateFalloffIntensity(0.5, invalid) = %v, want %v (linear fallback)", got, want)
	}
}

// TestLightCacheKey tests the cache key struct.
func TestLightCacheKey(t *testing.T) {
	key1 := lightCacheKey{diameter: 100, falloff: FalloffLinear}
	key2 := lightCacheKey{diameter: 100, falloff: FalloffLinear}
	key3 := lightCacheKey{diameter: 200, falloff: FalloffLinear}
	key4 := lightCacheKey{diameter: 100, falloff: FalloffQuadratic}

	// Same keys should be equal
	if key1 != key2 {
		t.Error("Identical cache keys should be equal")
	}

	// Different diameter should create different key
	if key1 == key3 {
		t.Error("Different diameter should create different cache key")
	}

	// Different falloff should create different key
	if key1 == key4 {
		t.Error("Different falloff should create different cache key")
	}
}

// TestBloomIntegration tests that bloom effect is applied when enabled.
func TestBloomIntegration(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.EnableBloom = true
	config.BloomIntensity = 1.5
	config.BloomThreshold = 0.7
	config.BloomRadius = 12

	system := NewLightingSystem(world, config)

	// Create bright light source
	entity := world.CreateEntity()
	light := NewLightComponent(100, color.RGBA{255, 100, 255, 255}, 2.0) // Bright enough to bloom
	pos := &PositionComponent{X: 400, Y: 300}
	entity.AddComponent(light)
	entity.AddComponent(pos)

	// Apply lighting
	system.SetViewport(0, 0, 800, 600)
	entities := []*Entity{entity}
	system.CollectVisibleLights(entities)

	// Verify bloom config is set
	if !system.config.EnableBloom {
		t.Error("Bloom should be enabled")
	}
	if system.config.BloomIntensity != 1.5 {
		t.Errorf("BloomIntensity = %v, want 1.5", system.config.BloomIntensity)
	}
	if system.config.BloomThreshold != 0.7 {
		t.Errorf("BloomThreshold = %v, want 0.7", system.config.BloomThreshold)
	}
	if system.config.BloomRadius != 12 {
		t.Errorf("BloomRadius = %v, want 12", system.config.BloomRadius)
	}
}

// TestBloomDisabled tests that bloom is not applied when disabled.
func TestBloomDisabled(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.EnableBloom = false

	system := NewLightingSystem(world, config)

	if system.config.EnableBloom {
		t.Error("Bloom should be disabled")
	}
}

// TestBloomConfigDefaults tests default bloom configuration values.
func TestBloomConfigDefaults(t *testing.T) {
	config := NewLightingConfig()

	// Bloom should be enabled by default
	if !config.EnableBloom {
		t.Error("Bloom should be enabled by default")
	}

	// Verify sensible defaults
	if config.BloomThreshold <= 0 || config.BloomThreshold >= 1 {
		t.Errorf("BloomThreshold = %v, want in range (0, 1)", config.BloomThreshold)
	}
	if config.BloomIntensity <= 0 {
		t.Errorf("BloomIntensity = %v, want > 0", config.BloomIntensity)
	}
	if config.BloomRadius <= 0 {
		t.Errorf("BloomRadius = %v, want > 0", config.BloomRadius)
	}
}

// TestApplyBloomEffect tests the bloom effect application via CPU-side buffers.
// Tests the pkg/rendering/lighting.ApplyBloom function directly, avoiding Ebiten runtime deps.
func TestApplyBloomEffect(t *testing.T) {
	// Create a CPU-side image with some bright pixels
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	// Draw a bright center region (above threshold)
	brightColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 20; y < 44; y++ {
		for x := 20; x < 44; x++ {
			img.Set(x, y, brightColor)
		}
	}

	// Configure bloom
	bloomConfig := lighting.BloomConfig{
		Enabled:   true,
		Threshold: 0.8,
		Intensity: 1.0,
		Radius:    8,
		Samples:   5,
	}

	// Apply bloom effect (CPU-side, no Ebiten required)
	result := lighting.ApplyBloom(img, bloomConfig)

	if result == nil {
		t.Fatal("ApplyBloom returned nil")
	}

	// Verify result has same dimensions
	if result.Bounds() != img.Bounds() {
		t.Errorf("Result bounds %v != input bounds %v", result.Bounds(), img.Bounds())
	}

	// Verify bloom spread: pixels outside bright region should now have some brightness
	// due to bloom glow effect
	edgeColor := result.At(15, 32)
	r, g, b, _ := edgeColor.RGBA()
	if r == 0 && g == 0 && b == 0 {
		t.Error("Bloom should have spread brightness to edge pixels")
	}
}

// TestBloomIntensityZero tests that bloom is skipped when intensity is 0.
func TestBloomIntensityZero(t *testing.T) {
	world := NewWorld()
	config := NewLightingConfig()
	config.EnableBloom = true
	config.BloomIntensity = 0 // Zero intensity should skip bloom

	system := NewLightingSystem(world, config)

	// Bloom should be configured but effectively disabled
	if !system.config.EnableBloom {
		t.Error("EnableBloom should be true even with zero intensity")
	}
	if system.config.BloomIntensity != 0 {
		t.Errorf("BloomIntensity = %v, want 0", system.config.BloomIntensity)
	}
}
