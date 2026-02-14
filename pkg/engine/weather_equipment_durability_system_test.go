//go:build ignore

package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/particles"
)

func TestNewWeatherEquipmentDurabilitySystem(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	if sys == nil {
		t.Fatal("NewWeatherEquipmentDurabilitySystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %s, want fantasy", sys.genreID)
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("Update interval = %f, want 1.0", sys.updateInterval)
	}
}

func TestWeatherEquipmentDurabilitySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		genre    string
		wantMult float64
	}{
		{"fantasy", 1.0},
		{"scifi", 0.5},
		{"horror", 1.3},
		{"cyberpunk", 0.7},
		{"postapoc", 1.4},
		{"unknown", 1.0}, // Falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys.SetGenre(tt.genre)
			if sys.genreID != tt.genre {
				t.Errorf("Genre = %s, want %s", sys.genreID, tt.genre)
			}
			mult := sys.getGenreMultiplier()
			if mult != tt.wantMult {
				t.Errorf("Genre multiplier = %f, want %f", mult, tt.wantMult)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_IntensityMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		intensity particles.WeatherIntensity
		wantMult  float64
	}{
		{particles.IntensityLight, 0.25},
		{particles.IntensityMedium, 0.5},
		{particles.IntensityHeavy, 1.0},
		{particles.IntensityExtreme, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.intensity.String(), func(t *testing.T) {
			mult := sys.getIntensityMultiplier(tt.intensity)
			if mult != tt.wantMult {
				t.Errorf("Intensity multiplier = %f, want %f", mult, tt.wantMult)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_WeatherDamageRates(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		weather    particles.WeatherType
		wantDamage float64
		hasDamage  bool
	}{
		{particles.WeatherRain, 0.3, true},
		{particles.WeatherSnow, 0.2, true},
		{particles.WeatherSandstorm, 0.8, true},
		{particles.WeatherRadiation, 1.0, true},
		{particles.WeatherFog, 0.0, true}, // No damage
		{particles.WeatherBloodRain, 0.6, true},
	}

	for _, tt := range tests {
		t.Run(tt.weather.String(), func(t *testing.T) {
			damage, ok := sys.weatherDamageRates[tt.weather]
			if ok != tt.hasDamage {
				t.Errorf("Has damage = %v, want %v", ok, tt.hasDamage)
			}
			if damage != tt.wantDamage {
				t.Errorf("Damage rate = %f, want %f", damage, tt.wantDamage)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_SlotMultipliers(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		weather  particles.WeatherType
		slot     EquipmentSlot
		wantMult float64
	}{
		{particles.WeatherRain, SlotMainHand, 2.0},    // Weapons rust
		{particles.WeatherRain, SlotOffHand, 2.0},     // Weapons rust
		{particles.WeatherRain, SlotChest, 0.8},       // Armor protected
		{particles.WeatherSnow, SlotBoots, 1.5},       // Feet in snow
		{particles.WeatherSandstorm, SlotHead, 1.5},   // Face protection
		{particles.WeatherBloodRain, SlotGloves, 1.5}, // Exposed hands
	}

	for _, tt := range tests {
		t.Run(tt.weather.String()+"_"+tt.slot.String(), func(t *testing.T) {
			mult := sys.getSlotMultiplier(tt.weather, tt.slot)
			if mult != tt.wantMult {
				t.Errorf("Slot multiplier = %f, want %f", mult, tt.wantMult)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_DamageStateChanged(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	tests := []struct {
		name   string
		oldDur int
		newDur int
		maxDur int
		want   bool
	}{
		{"no change 100%", 100, 100, 100, false},
		{"no change 50%", 50, 50, 100, false},
		{"cross 75% threshold", 76, 75, 100, true},
		{"cross 50% threshold", 51, 50, 100, true},
		{"cross 25% threshold", 26, 25, 100, true},
		{"no threshold cross", 80, 77, 100, false},
		{"zero max", 50, 40, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.damageStateChanged(tt.oldDur, tt.newDur, tt.maxDur)
			if got != tt.want {
				t.Errorf("damageStateChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_Update_NoWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	testItem := &item.Item{
		ID:   1,
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	equipComp.Slots[SlotMainHand] = testItem
	entity.AddComponent(equipComp)

	// Update without weather - should not damage equipment
	sys.Update([]*Entity{entity}, 2.0) // 2 seconds to exceed interval

	if testItem.Stats.Durability != 100 {
		t.Errorf("Durability changed without weather: got %d, want 100", testItem.Stats.Durability)
	}
}

func TestWeatherEquipmentDurabilitySystem_Update_WithWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityHeavy,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	testItem := &item.Item{
		ID:   1,
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	equipComp.Slots[SlotMainHand] = testItem
	entity.AddComponent(equipComp)

	// Run multiple updates to accumulate damage
	for i := 0; i < 5; i++ {
		sys.Update([]*Entity{weatherEntity, entity}, 1.1) // Slightly over interval
	}

	// Rain at Heavy intensity should damage weapons
	// Base: 0.3 * genre(1.0) * intensity(1.0) * slot(2.0 for weapon) = 0.6 per update
	// After 5 updates: 100 - 3 = ~97 (int conversion)
	if testItem.Stats.Durability >= 100 {
		t.Errorf("Durability not reduced by weather: got %d", testItem.Stats.Durability)
	}
}

func TestWeatherEquipmentDurabilitySystem_Update_GenreModifiers(t *testing.T) {
	tests := []struct {
		genre          string
		expectedLower  int // Min expected durability after damage
		expectedHigher int // Max expected durability after damage
	}{
		{"scifi", 97, 100},   // Low damage (0.5x)
		{"postapoc", 90, 98}, // High damage (1.4x)
		{"fantasy", 95, 99},  // Baseline (1.0x)
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			world := NewWorld()
			sys := NewWeatherEquipmentDurabilitySystem(world, 12345)
			sys.SetGenre(tt.genre)

			// Create weather entity
			weatherEntity := world.CreateEntity()
			weatherComp := &WeatherComponent{
				Active: true,
				Config: particles.WeatherConfig{
					Type:      particles.WeatherRain,
					Intensity: particles.IntensityHeavy,
				},
			}
			weatherEntity.AddComponent(weatherComp)

			// Create entity with equipment
			entity := world.CreateEntity()
			equipComp := NewEquipmentComponent()
			testItem := &item.Item{
				ID:   1,
				Name: "Test Sword",
				Type: item.TypeWeapon,
				Stats: item.Stats{
					Durability:    100,
					DurabilityMax: 100,
				},
			}
			equipComp.Slots[SlotMainHand] = testItem
			entity.AddComponent(equipComp)

			// Run updates
			for i := 0; i < 3; i++ {
				sys.Update([]*Entity{weatherEntity, entity}, 1.1)
			}

			if testItem.Stats.Durability > tt.expectedHigher || testItem.Stats.Durability < tt.expectedLower {
				t.Errorf("%s durability = %d, want between %d and %d",
					tt.genre, testItem.Stats.Durability, tt.expectedLower, tt.expectedHigher)
			}
		})
	}
}

func TestWeatherEquipmentDurabilitySystem_FindActiveWeather(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	// Test no weather
	_, _, active := sys.findActiveWeather([]*Entity{})
	if active {
		t.Error("Found weather when none exists")
	}

	// Test inactive weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: false,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherSnow,
			Intensity: particles.IntensityMedium,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	_, _, active = sys.findActiveWeather([]*Entity{weatherEntity})
	if active {
		t.Error("Found inactive weather as active")
	}

	// Test active weather
	weatherComp.Active = true
	wType, intensity, active := sys.findActiveWeather([]*Entity{weatherEntity})
	if !active {
		t.Error("Did not find active weather")
	}
	if wType != particles.WeatherSnow {
		t.Errorf("Weather type = %v, want Snow", wType)
	}
	if intensity != particles.IntensityMedium {
		t.Errorf("Intensity = %v, want Medium", intensity)
	}
}

func TestWeatherEquipmentDurabilitySystem_NoEquipment(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	// Create weather entity
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityHeavy,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity without equipment
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not panic
	sys.Update([]*Entity{weatherEntity, entity}, 1.1)
}

func TestWeatherEquipmentDurabilitySystem_FogNoDamage(t *testing.T) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)

	// Create fog weather (should not damage)
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherFog,
			Intensity: particles.IntensityExtreme,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create entity with equipment
	entity := world.CreateEntity()
	equipComp := NewEquipmentComponent()
	testItem := &item.Item{
		ID:   1,
		Name: "Test Sword",
		Type: item.TypeWeapon,
		Stats: item.Stats{
			Durability:    100,
			DurabilityMax: 100,
		},
	}
	equipComp.Slots[SlotMainHand] = testItem
	entity.AddComponent(equipComp)

	// Run updates
	for i := 0; i < 10; i++ {
		sys.Update([]*Entity{weatherEntity, entity}, 1.1)
	}

	// Fog should not damage equipment
	if testItem.Stats.Durability != 100 {
		t.Errorf("Fog damaged equipment: durability = %d, want 100", testItem.Stats.Durability)
	}
}

func TestWeatherEquipmentDurabilitySystem_Determinism(t *testing.T) {
	// Run same scenario twice with same seed
	results := make([]int, 2)

	for run := 0; run < 2; run++ {
		world := NewWorld()
		sys := NewWeatherEquipmentDurabilitySystem(world, 99999)
		sys.SetGenre("fantasy")

		weatherEntity := world.CreateEntity()
		weatherComp := &WeatherComponent{
			Active: true,
			Config: particles.WeatherConfig{
				Type:      particles.WeatherSandstorm,
				Intensity: particles.IntensityHeavy,
			},
		}
		weatherEntity.AddComponent(weatherComp)

		entity := world.CreateEntity()
		equipComp := NewEquipmentComponent()
		testItem := &item.Item{
			ID:   1,
			Name: "Test Helmet",
			Type: item.TypeArmor,
			Stats: item.Stats{
				Durability:    100,
				DurabilityMax: 100,
			},
		}
		equipComp.Slots[SlotHead] = testItem
		entity.AddComponent(equipComp)

		for i := 0; i < 10; i++ {
			sys.Update([]*Entity{weatherEntity, entity}, 1.1)
		}

		results[run] = testItem.Stats.Durability
	}

	if results[0] != results[1] {
		t.Errorf("Non-deterministic results: run1=%d, run2=%d", results[0], results[1])
	}
}

func BenchmarkWeatherEquipmentDurabilitySystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewWeatherEquipmentDurabilitySystem(world, 12345)
	rng := rand.New(rand.NewSource(12345))

	// Create weather
	weatherEntity := world.CreateEntity()
	weatherComp := &WeatherComponent{
		Active: true,
		Config: particles.WeatherConfig{
			Type:      particles.WeatherRain,
			Intensity: particles.IntensityHeavy,
		},
	}
	weatherEntity.AddComponent(weatherComp)

	// Create 100 entities with equipment
	entities := make([]*Entity, 101)
	entities[0] = weatherEntity
	for i := 1; i <= 100; i++ {
		entity := world.CreateEntity()
		equipComp := NewEquipmentComponent()
		equipComp.Slots[SlotMainHand] = &item.Item{
			ID: uint64(i),
			Stats: item.Stats{
				Durability:    100 + rng.Intn(50),
				DurabilityMax: 150,
			},
		}
		equipComp.Slots[SlotChest] = &item.Item{
			ID: uint64(i + 1000),
			Stats: item.Stats{
				Durability:    80 + rng.Intn(40),
				DurabilityMax: 120,
			},
		}
		entity.AddComponent(equipComp)
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 0 // Force update
		sys.Update(entities, 1.1)
	}
}
