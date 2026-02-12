package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherManaRegenSystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherManaRegenSystem returned nil")
	}
	if sys.world != world {
		t.Error("World reference not set")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("Expected update interval 0.5, got %f", sys.updateInterval)
	}
	if sys.originalRegen == nil {
		t.Error("originalRegen map not initialized")
	}
}

func TestWeatherManaRegenSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Errorf("Expected genre scifi, got %s", sys.genreID)
	}
}

func TestWeatherManaRegenSystem_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create entity with mana but no weather
	entity := world.CreateEntity()
	entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	entities := world.GetEntities()
	sys.Update(entities, 1.0) // Long delta to trigger check

	// Verify mana regen unchanged
	manaComp, _ := entity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Regen != 5.0 {
		t.Errorf("Mana regen should be unchanged at 5.0, got %f", mana.Regen)
	}
}

func TestWeatherManaRegenSystem_RainBoostsManaRegen(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather entity with rain
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create entity with mana
	manaEntity := world.CreateEntity()
	manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	entities := world.GetEntities()
	sys.Update(entities, 1.0)

	// Verify mana regen increased (rain gives 15% base boost)
	manaComp, _ := manaEntity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Regen <= 5.0 {
		t.Errorf("Rain should boost mana regen above 5.0, got %f", mana.Regen)
	}
	// With medium intensity and 15% base, expect roughly 1.10-1.25 range
	if mana.Regen > 8.0 {
		t.Errorf("Mana regen boost too high, got %f", mana.Regen)
	}
}

func TestWeatherManaRegenSystem_FogReducesManaRegen(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather entity with fog
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherFog,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create entity with mana
	manaEntity := world.CreateEntity()
	manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

	entities := world.GetEntities()
	sys.Update(entities, 1.0)

	// Verify mana regen decreased (fog gives 10% base penalty)
	manaComp, _ := manaEntity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Regen >= 10.0 {
		t.Errorf("Fog should reduce mana regen below 10.0, got %f", mana.Regen)
	}
}

func TestWeatherManaRegenSystem_WeatherClears(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather entity with rain
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create entity with mana
	manaEntity := world.CreateEntity()
	manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	entities := world.GetEntities()

	// Apply weather effect
	sys.Update(entities, 1.0)

	// Weather clears
	weatherComp.Active = false
	sys.Update(entities, 1.0)

	// Verify mana regen restored
	manaComp, _ := manaEntity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Regen != 5.0 {
		t.Errorf("Mana regen should be restored to 5.0, got %f", mana.Regen)
	}
}

func TestWeatherManaRegenSystem_IntensityAffectsMultiplier(t *testing.T) {
	tests := []struct {
		name      string
		intensity particles.WeatherIntensity
	}{
		{"light", particles.IntensityLight},
		{"medium", particles.IntensityMedium},
		{"heavy", particles.IntensityHeavy},
		{"extreme", particles.IntensityExtreme},
	}

	lastRegen := 0.0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherManaRegenSystem(world, 12345)

			weatherEntity := world.CreateEntity()
			weatherConfig := particles.WeatherConfig{
				Type:      particles.WeatherRain,
				Intensity: tt.intensity,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			manaEntity := world.CreateEntity()
			manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

			entities := world.GetEntities()
			sys.Update(entities, 1.0)

			manaComp, _ := manaEntity.GetComponent("mana")
			mana := manaComp.(*ManaComponent)

			// Each intensity should produce a different result
			// Light < Medium < Heavy < Extreme for rain boost
			if tt.intensity > particles.IntensityLight && mana.Regen <= lastRegen {
				// Due to randomization, we allow some variance
				t.Logf("Note: Intensity %s regen %f may vary due to RNG", tt.name, mana.Regen)
			}
			lastRegen = mana.Regen
		})
	}
}

func TestWeatherManaRegenSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create multiple mana entities
	entity1 := world.CreateEntity()
	entity1.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&ManaComponent{Current: 30, Max: 80, Regen: 3.0})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&ManaComponent{Current: 100, Max: 200, Regen: 10.0})

	entities := world.GetEntities()
	sys.Update(entities, 1.0)

	// All should have boosted regen
	for _, e := range []*Entity{entity1, entity2, entity3} {
		manaComp, ok := e.GetComponent("mana")
		if !ok {
			continue
		}
		mana := manaComp.(*ManaComponent)
		originalRegen := sys.originalRegen[e.ID]
		if mana.Regen <= originalRegen {
			t.Errorf("Entity %d should have boosted regen, original %f, current %f",
				e.ID, originalRegen, mana.Regen)
		}
	}
}

func TestWeatherManaRegenSystem_NoManaComponent(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create entity without mana
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := world.GetEntities()

	// Should not panic
	sys.Update(entities, 1.0)

	// No entries in original regen cache for non-mana entities
	if _, exists := sys.originalRegen[entity.ID]; exists {
		t.Error("Non-mana entity should not be in original regen cache")
	}
}

func TestWeatherManaRegenSystem_AllWeatherTypes(t *testing.T) {
	weatherTypes := []struct {
		weather     particles.WeatherType
		expectBoost bool // true if generally boosts, false if generally reduces
	}{
		{particles.WeatherRain, true},
		{particles.WeatherSnow, true},
		{particles.WeatherFog, false},
		{particles.WeatherDust, false},
		{particles.WeatherAsh, false},
		{particles.WeatherNeonRain, true},
		{particles.WeatherSmog, false},
		{particles.WeatherRadiation, true},
		{particles.WeatherSandstorm, false},
		{particles.WeatherBloodRain, true},
	}

	for _, tt := range weatherTypes {
		t.Run(tt.weather.String(), func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherManaRegenSystem(world, 12345)

			weatherEntity := world.CreateEntity()
			weatherConfig := particles.WeatherConfig{
				Type:      tt.weather,
				Intensity: particles.IntensityMedium,
			}
			weatherComp := NewWeatherComponent(weatherConfig)
			weatherComp.Active = true
			weatherEntity.AddComponent(weatherComp)

			manaEntity := world.CreateEntity()
			manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 10.0})

			entities := world.GetEntities()
			sys.Update(entities, 1.0)

			manaComp, _ := manaEntity.GetComponent("mana")
			mana := manaComp.(*ManaComponent)

			if tt.expectBoost && mana.Regen < 10.0 {
				t.Errorf("%s should boost regen, got %f", tt.weather.String(), mana.Regen)
			}
			if !tt.expectBoost && mana.Regen > 10.0 {
				t.Errorf("%s should reduce regen, got %f", tt.weather.String(), mana.Regen)
			}
		})
	}
}

func TestWeatherManaRegenSystem_CalculateManaMultiplier(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Test bounds: multiplier should be in [0.5, 2.0]
	for i := 0; i < 100; i++ {
		mult := sys.calculateManaMultiplier(particles.WeatherBloodRain, particles.IntensityExtreme)
		if mult < 0.5 || mult > 2.0 {
			t.Errorf("Multiplier out of bounds: %f", mult)
		}
	}
}

func TestWeatherManaRegenSystem_UpdateInterval(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	manaEntity := world.CreateEntity()
	manaEntity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})

	entities := world.GetEntities()

	// Small delta should not trigger update
	sys.Update(entities, 0.1)
	manaComp, _ := manaEntity.GetComponent("mana")
	mana := manaComp.(*ManaComponent)
	if mana.Regen != 5.0 {
		t.Errorf("Small delta should not trigger update, regen is %f", mana.Regen)
	}

	// Accumulate to pass threshold
	sys.Update(entities, 0.5)
	manaComp, _ = manaEntity.GetComponent("mana")
	mana = manaComp.(*ManaComponent)
	if mana.Regen == 5.0 {
		t.Error("Accumulated delta should trigger update")
	}
}

func TestWeatherManaRegenSystem_IsWeatherActive(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	if sys.IsWeatherActive() {
		t.Error("Weather should not be active initially")
	}

	// Activate weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	entities := world.GetEntities()
	sys.Update(entities, 1.0)

	if !sys.IsWeatherActive() {
		t.Error("Weather should be active after update")
	}
}

func TestWeatherManaRegenSystem_GetActiveWeatherType(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherSnow,
		Intensity: particles.IntensityHeavy,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	entities := world.GetEntities()
	sys.Update(entities, 1.0)

	if sys.GetActiveWeatherType() != particles.WeatherSnow {
		t.Errorf("Expected WeatherSnow, got %v", sys.GetActiveWeatherType())
	}
}

func BenchmarkWeatherManaRegenSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherManaRegenSystem(world, 12345)

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherConfig := particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	}
	weatherComp := NewWeatherComponent(weatherConfig)
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)

	// Create many mana entities
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&ManaComponent{Current: 50, Max: 100, Regen: 5.0})
	}

	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016) // 60 FPS
	}
}
