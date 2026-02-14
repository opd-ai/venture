//go:build ignore

package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewFishingWeatherBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFishingWeatherBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
	if sys.updateInterval != 0.5 {
		t.Errorf("Update interval = %f, want 0.5", sys.updateInterval)
	}
}

func TestFishingWeatherBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	tests := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range tests {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%s): genre = %s", genre, sys.genreID)
		}
	}
}

func TestFishingWeatherBonusSystem_SetFishingSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	fishingSys := NewFishingSystem(world, 12345)
	sys.SetFishingSystem(fishingSys)

	if sys.fishingSystem != fishingSys {
		t.Error("FishingSystem not set correctly")
	}

	// Verify callback was set
	weather := fishingSys.CurrentWeather()
	if weather != "clear" {
		t.Errorf("Initial weather = %s, want clear", weather)
	}
}

func TestFishingWeatherBonusSystem_Update_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	// Create fishing spot
	spot := NewEntity(1)
	spot.AddComponent(NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "forest"))
	spot.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(spot)

	entities := world.GetAllEntities()

	// Update without weather - should not change bonuses
	sys.Update(entities, 1.0)

	spotComp := spot.GetTypedComponent("fishing_spot").(*FishingSpotComponent)
	if spotComp.RareFishBonus != 1.0 {
		t.Errorf("RareFishBonus = %f, want 1.0 (no weather)", spotComp.RareFishBonus)
	}
}

func TestFishingWeatherBonusSystem_Update_WithRain(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	// Create fishing spot (freshwater for max bonus)
	spot := NewEntity(1)
	spotComp := NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "forest")
	spot.AddComponent(spotComp)
	spot.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(spot)

	// Create weather entity with rain
	weatherEntity := NewEntity(2)
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
		Width:     800,
		Height:    600,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)
	world.AddEntity(weatherEntity)

	entities := world.GetAllEntities()

	// Update with rain - should increase bonus
	sys.Update(entities, 1.0)

	// Rain + Freshwater = 1.4, IntensityMedium = 1.0, Fantasy bonus = 1.15
	expectedBonus := 1.4 * 1.0 * 1.15 // 1.61
	if spotComp.RareFishBonus < 1.5 || spotComp.RareFishBonus > 1.7 {
		t.Errorf("RareFishBonus = %f, want ~%f (rain)", spotComp.RareFishBonus, expectedBonus)
	}
}

func TestFishingWeatherBonusSystem_Update_WeatherClears(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	// Create fishing spot with custom bonus
	spot := NewEntity(1)
	spotComp := NewFishingSpotComponent(WaterTypeSaltwater, DepthMedium, "ocean")
	spotComp.RareFishBonus = 1.5 // Custom bonus
	spot.AddComponent(spotComp)
	spot.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(spot)

	// Create weather entity
	weatherEntity := NewEntity(2)
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)
	world.AddEntity(weatherEntity)

	entities := world.GetAllEntities()

	// Activate weather
	sys.Update(entities, 1.0)
	bonusWithWeather := spotComp.RareFishBonus
	if bonusWithWeather <= 1.5 {
		t.Errorf("Bonus should increase with rain, got %f", bonusWithWeather)
	}

	// Deactivate weather
	weatherComp.Active = false
	sys.Update(entities, 1.0)

	// Should restore original bonus
	if spotComp.RareFishBonus != 1.5 {
		t.Errorf("RareFishBonus = %f, want 1.5 (restored)", spotComp.RareFishBonus)
	}
}

func TestFishingWeatherBonusSystem_getBaseModifier(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	tests := []struct {
		name      string
		weather   particles.WeatherType
		waterType WaterType
		wantMin   float64
		wantMax   float64
	}{
		{"rain freshwater", particles.WeatherRain, WaterTypeFreshwater, 1.35, 1.45},
		{"rain saltwater", particles.WeatherRain, WaterTypeSaltwater, 1.2, 1.3},
		{"snow any", particles.WeatherSnow, WaterTypeFreshwater, 0.8, 0.9},
		{"fog any", particles.WeatherFog, WaterTypeSaltwater, 0.9, 1.0},
		{"smog any", particles.WeatherSmog, WaterTypeFreshwater, 0.5, 0.7},
		{"neon magical", particles.WeatherNeonRain, WaterTypeMagical, 1.5, 1.7},
		{"blood rain", particles.WeatherBloodRain, WaterTypeSaltwater, 1.7, 1.9},
		{"radiation", particles.WeatherRadiation, WaterTypeMagical, 1.4, 1.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := sys.getBaseModifier(tt.weather, tt.waterType)
			if mod < tt.wantMin || mod > tt.wantMax {
				t.Errorf("getBaseModifier(%v, %v) = %f, want [%f, %f]",
					tt.weather, tt.waterType, mod, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFishingWeatherBonusSystem_getIntensityMultiplier(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	tests := []struct {
		intensity particles.WeatherIntensity
		want      float64
	}{
		{particles.IntensityLight, 0.8},
		{particles.IntensityMedium, 1.0},
		{particles.IntensityHeavy, 1.3},
		{particles.IntensityExtreme, 1.6},
	}

	for _, tt := range tests {
		mult := sys.getIntensityMultiplier(tt.intensity)
		if mult != tt.want {
			t.Errorf("getIntensityMultiplier(%v) = %f, want %f", tt.intensity, mult, tt.want)
		}
	}
}

func TestFishingWeatherBonusSystem_getGenreBonus(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	tests := []struct {
		genre   string
		weather particles.WeatherType
		wantMin float64
	}{
		{"fantasy", particles.WeatherRain, 1.1},
		{"scifi", particles.WeatherRadiation, 1.15},
		{"horror", particles.WeatherBloodRain, 1.2},
		{"cyberpunk", particles.WeatherNeonRain, 1.15},
		{"postapoc", particles.WeatherSandstorm, 1.1},
		{"fantasy", particles.WeatherSmog, 1.0}, // No bonus
	}

	for _, tt := range tests {
		t.Run(tt.genre+"_"+tt.weather.String(), func(t *testing.T) {
			sys.SetGenre(tt.genre)
			bonus := sys.getGenreBonus(tt.weather)
			if bonus < tt.wantMin {
				t.Errorf("getGenreBonus(%v) = %f, want >= %f", tt.weather, bonus, tt.wantMin)
			}
		})
	}
}

func TestFishingWeatherBonusSystem_getCurrentWeatherString(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	// Initially no weather
	weather := sys.getCurrentWeatherString()
	if weather != "clear" {
		t.Errorf("Initial weather = %s, want clear", weather)
	}

	// Set active weather
	sys.lastWeatherActive = true
	sys.lastWeatherType = particles.WeatherRain

	weather = sys.getCurrentWeatherString()
	if weather != "Rain" {
		t.Errorf("Weather with rain = %s, want Rain", weather)
	}

	// Try blood rain
	sys.lastWeatherType = particles.WeatherBloodRain
	weather = sys.getCurrentWeatherString()
	if weather != "BloodRain" {
		t.Errorf("Weather with blood rain = %s, want BloodRain", weather)
	}
}

func TestFishingWeatherBonusSystem_GetCurrentWeatherType(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	// Initially inactive
	_, active := sys.GetCurrentWeatherType()
	if active {
		t.Error("Initial weather should be inactive")
	}

	// Activate
	sys.lastWeatherActive = true
	sys.lastWeatherType = particles.WeatherSnow

	wtype, active := sys.GetCurrentWeatherType()
	if !active {
		t.Error("Weather should be active")
	}
	if wtype != particles.WeatherSnow {
		t.Errorf("Weather type = %v, want Snow", wtype)
	}
}

func TestFishingWeatherBonusSystem_GetBonusModifier(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)
	sys.SetGenre("fantasy")

	// No weather = 1.0
	mod := sys.GetBonusModifier(WaterTypeFreshwater)
	if mod != 1.0 {
		t.Errorf("Modifier without weather = %f, want 1.0", mod)
	}

	// With rain
	sys.lastWeatherActive = true
	sys.lastWeatherType = particles.WeatherRain
	sys.lastWeatherIntensity = particles.IntensityHeavy

	mod = sys.GetBonusModifier(WaterTypeFreshwater)
	// 1.4 (rain+freshwater) * 1.3 (heavy) * 1.15 (fantasy) = 2.093
	if mod < 2.0 || mod > 2.2 {
		t.Errorf("Modifier with heavy rain = %f, want ~2.1", mod)
	}
}

func TestFishingWeatherBonusSystem_Update_NilWorld(t *testing.T) {
	sys := NewFishingWeatherBonusSystem(nil, 12345)

	// Should not panic
	sys.Update(nil, 1.0)
	sys.Update([]*Entity{}, 1.0)
}

func TestFishingWeatherBonusSystem_Update_MultipleSpots(t *testing.T) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	// Create multiple fishing spots with different water types
	spots := []*Entity{
		NewEntity(1),
		NewEntity(2),
		NewEntity(3),
	}

	waterTypes := []WaterType{WaterTypeFreshwater, WaterTypeSaltwater, WaterTypeMagical}

	for i, spot := range spots {
		spotComp := NewFishingSpotComponent(waterTypes[i], DepthMedium, "test")
		spot.AddComponent(spotComp)
		spot.AddComponent(&PositionComponent{X: float64(i * 100), Y: 100})
		world.AddEntity(spot)
	}

	// Add weather
	weatherEntity := NewEntity(100)
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityMedium,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)
	world.AddEntity(weatherEntity)

	entities := world.GetAllEntities()
	sys.Update(entities, 1.0)

	// Verify all spots were affected
	for i, spot := range spots {
		spotComp := spot.GetTypedComponent("fishing_spot").(*FishingSpotComponent)
		if spotComp.RareFishBonus <= 1.0 {
			t.Errorf("Spot %d (water=%v): bonus = %f, want > 1.0",
				i, waterTypes[i], spotComp.RareFishBonus)
		}
	}

	// Freshwater should have highest bonus
	fresh := spots[0].GetTypedComponent("fishing_spot").(*FishingSpotComponent)
	salt := spots[1].GetTypedComponent("fishing_spot").(*FishingSpotComponent)
	if fresh.RareFishBonus <= salt.RareFishBonus {
		t.Errorf("Freshwater bonus (%f) should be > Saltwater (%f)",
			fresh.RareFishBonus, salt.RareFishBonus)
	}
}

func BenchmarkFishingWeatherBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFishingWeatherBonusSystem(world, 12345)

	// Create 100 fishing spots
	for i := 0; i < 100; i++ {
		spot := NewEntity(uint64(i))
		spot.AddComponent(NewFishingSpotComponent(WaterTypeFreshwater, DepthMedium, "test"))
		spot.AddComponent(&PositionComponent{X: float64(i * 10), Y: 100})
		world.AddEntity(spot)
	}

	// Add weather
	weatherEntity := NewEntity(1000)
	weatherComp := NewWeatherComponent(particles.WeatherConfig{
		Type:      particles.WeatherRain,
		Intensity: particles.IntensityHeavy,
	})
	weatherComp.Active = true
	weatherEntity.AddComponent(weatherComp)
	world.AddEntity(weatherEntity)

	entities := world.GetAllEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0.5 // Force update
		sys.Update(entities, 0.016)
	}
}
