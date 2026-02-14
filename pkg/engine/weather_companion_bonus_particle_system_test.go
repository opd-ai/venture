package engine

import (
	"testing"
)

func TestNewWeatherCompanionBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherCompanionBonusParticleSystem returned nil")
	}

	if system.world != world {
		t.Error("world not set correctly")
	}

	if system.seed != 12345 {
		t.Errorf("seed = %d, want 12345", system.seed)
	}

	if system.rng == nil {
		t.Error("rng not initialized")
	}

	if system.pulseInterval != 2.5 {
		t.Errorf("pulseInterval = %f, want 2.5", system.pulseInterval)
	}

	if system.lastBonusState == nil {
		t.Error("lastBonusState map not initialized")
	}
}

func TestWeatherCompanionBonusParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particle system not set correctly")
	}
}

func TestWeatherCompanionBonusParticleSystem_SetWeatherCompanionBonusSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	wcbs := NewWeatherCompanionBonusSystem(world, 54321)

	system.SetWeatherCompanionBonusSystem(wcbs)

	if system.weatherCompanionBonusSystem != wcbs {
		t.Error("weather companion bonus system not set correctly")
	}
}

func TestWeatherCompanionBonusParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)

	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			system.SetGenre(tt.genre)
			if system.genreID != tt.genre {
				t.Errorf("genreID = %s, want %s", system.genreID, tt.genre)
			}
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_UpdateWithoutParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)

	// Should not panic with nil particle system
	system.Update([]*Entity{}, 0.016)
}

func TestWeatherCompanionBonusParticleSystem_UpdateWithoutWorld(t *testing.T) {
	system := NewWeatherCompanionBonusParticleSystem(nil, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Should not panic with nil world
	system.Update([]*Entity{}, 0.016)
}

func TestWeatherCompanionBonusParticleSystem_ProcessNonCompanion(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create non-companion entity with position
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should skip non-companion entities without error
	system.Update([]*Entity{entity}, 0.016)
}

func TestWeatherCompanionBonusParticleSystem_ProcessCompanionWithoutBonus(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create companion entity without weather bonus
	companion := NewEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		CompType:    CompanionTypePet,
		OwnerID:     1,
		Name:        "Test Pet",
		Loyalty:     50,
		BondingPerk: PerkNone,
	})

	// Should process without error
	system.Update([]*Entity{companion}, 0.016)
}

func TestWeatherCompanionBonusParticleSystem_ProcessCompanionWithBonus(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create companion entity with weather bonus
	companion := NewEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		CompType:    CompanionTypeElemental,
		OwnerID:     1,
		Name:        "Water Sprite",
		Loyalty:     75,
		BondingPerk: PerkNone,
	})
	companion.AddComponent(&WeatherCompanionBonusComponent{
		AttackBonus:       1.25,
		DefenseBonus:      1.20,
		SpeedBonus:        1.10,
		WeatherType:       "Rain",
		CompanionTypeName: "Elemental",
	})

	// First update should trigger particle spawn
	system.Update([]*Entity{companion}, 0.016)

	// Verify state was cached
	if _, exists := system.lastBonusState[companion.ID]; !exists {
		t.Error("bonus state not cached after first update")
	}
}

func TestWeatherCompanionBonusParticleSystem_DetectsBonusStateChange(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create companion with initial bonus
	companion := NewEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		CompType: CompanionTypeRobot,
		OwnerID:  1,
	})

	bonusComp := &WeatherCompanionBonusComponent{
		AttackBonus:       0.85,
		DefenseBonus:      0.90,
		SpeedBonus:        0.90,
		WeatherType:       "Rain",
		CompanionTypeName: "Robot",
	}
	companion.AddComponent(bonusComp)

	// First update
	system.Update([]*Entity{companion}, 0.016)

	// Change bonus state
	bonusComp.AttackBonus = 1.20
	bonusComp.DefenseBonus = 1.15
	bonusComp.SpeedBonus = 1.10
	bonusComp.WeatherType = "NeonRain"

	// Second update should detect change
	system.Update([]*Entity{companion}, 0.016)

	// Verify updated state
	state, exists := system.lastBonusState[companion.ID]
	if !exists {
		t.Fatal("bonus state not cached")
	}

	if !state.hasAttackBonus {
		t.Error("attack bonus not detected after change")
	}

	if state.weatherType != "NeonRain" {
		t.Errorf("weatherType = %s, want NeonRain", state.weatherType)
	}
}

func TestWeatherCompanionBonusParticleSystem_PulseInterval(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	companion := NewEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		CompType: CompanionTypeSpirit,
		OwnerID:  1,
	})
	companion.AddComponent(&WeatherCompanionBonusComponent{
		AttackBonus:       1.30,
		DefenseBonus:      1.25,
		SpeedBonus:        1.15,
		WeatherType:       "Fog",
		CompanionTypeName: "Spirit",
	})

	// Run multiple updates until pulse should trigger
	for i := 0; i < 200; i++ {
		system.Update([]*Entity{companion}, 0.016) // ~3.2 seconds total
	}

	// timeSinceEmit should have been reset at least once
	if system.timeSinceEmit > system.pulseInterval {
		t.Error("pulse interval not triggering resets")
	}
}

func TestWeatherCompanionBonusParticleSystem_BonusStateChanged(t *testing.T) {
	system := NewWeatherCompanionBonusParticleSystem(nil, 12345)

	tests := []struct {
		name    string
		last    weatherCompanionBonusStateCache
		current weatherCompanionBonusStateCache
		want    bool
	}{
		{
			name: "no change",
			last: weatherCompanionBonusStateCache{
				hasAttackBonus:  true,
				hasDefenseBonus: true,
				weatherType:     "Rain",
			},
			current: weatherCompanionBonusStateCache{
				hasAttackBonus:  true,
				hasDefenseBonus: true,
				weatherType:     "Rain",
			},
			want: false,
		},
		{
			name: "attack bonus changed",
			last: weatherCompanionBonusStateCache{
				hasAttackBonus: false,
			},
			current: weatherCompanionBonusStateCache{
				hasAttackBonus: true,
			},
			want: true,
		},
		{
			name: "defense penalty changed",
			last: weatherCompanionBonusStateCache{
				hasDefensePenalt: false,
			},
			current: weatherCompanionBonusStateCache{
				hasDefensePenalt: true,
			},
			want: true,
		},
		{
			name: "weather type changed",
			last: weatherCompanionBonusStateCache{
				weatherType: "Rain",
			},
			current: weatherCompanionBonusStateCache{
				weatherType: "Snow",
			},
			want: true,
		},
		{
			name: "speed bonus changed",
			last: weatherCompanionBonusStateCache{
				hasSpeedBonus: false,
			},
			current: weatherCompanionBonusStateCache{
				hasSpeedBonus: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.bonusStateChanged(tt.last, tt.current)
			if got != tt.want {
				t.Errorf("bonusStateChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_HasActiveBonus(t *testing.T) {
	system := NewWeatherCompanionBonusParticleSystem(nil, 12345)

	tests := []struct {
		name  string
		state weatherCompanionBonusStateCache
		want  bool
	}{
		{
			name:  "no bonuses",
			state: weatherCompanionBonusStateCache{},
			want:  false,
		},
		{
			name: "attack bonus only",
			state: weatherCompanionBonusStateCache{
				hasAttackBonus: true,
			},
			want: true,
		},
		{
			name: "defense penalty only",
			state: weatherCompanionBonusStateCache{
				hasDefensePenalt: true,
			},
			want: true,
		},
		{
			name: "speed bonus only",
			state: weatherCompanionBonusStateCache{
				hasSpeedBonus: true,
			},
			want: true,
		},
		{
			name: "multiple bonuses",
			state: weatherCompanionBonusStateCache{
				hasAttackBonus:  true,
				hasDefenseBonus: true,
				hasSpeedBonus:   true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.hasActiveBonus(tt.state)
			if got != tt.want {
				t.Errorf("hasActiveBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_GetWeatherParticleType(t *testing.T) {
	system := NewWeatherCompanionBonusParticleSystem(nil, 12345)

	tests := []struct {
		name        string
		weatherType string
		genre       string
		isBonus     bool
	}{
		{"rain bonus", "Rain", "fantasy", true},
		{"rain penalty", "Rain", "fantasy", false},
		{"snow", "Snow", "fantasy", true},
		{"fog", "Fog", "horror", true},
		{"dust", "Dust", "postapoc", true},
		{"radiation horror", "Radiation", "horror", true},
		{"neon rain", "NeonRain", "cyberpunk", true},
		{"blood rain horror", "BloodRain", "horror", false},
		{"unknown weather fantasy", "Unknown", "fantasy", true},
		{"unknown weather scifi", "Unknown", "scifi", true},
		{"unknown weather horror", "Unknown", "horror", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system.SetGenre(tt.genre)
			particleType := system.getWeatherParticleType(tt.weatherType, tt.isBonus)
			// Just verify it returns a valid particle type without panicking
			if particleType < 0 {
				t.Error("invalid particle type returned")
			}
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_CleanupRemovedBonus(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create companion with bonus
	companion := NewEntity()
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		CompType: CompanionTypeUndead,
		OwnerID:  1,
	})
	bonusComp := &WeatherCompanionBonusComponent{
		AttackBonus:  1.25,
		DefenseBonus: 1.20,
		SpeedBonus:   1.05,
		WeatherType:  "Ash",
	}
	companion.AddComponent(bonusComp)

	// First update caches state
	system.Update([]*Entity{companion}, 0.016)

	if _, exists := system.lastBonusState[companion.ID]; !exists {
		t.Fatal("state not cached after update")
	}

	// Remove bonus component
	companion.RemoveComponent("weather_companion_bonus")

	// Update should clean up cached state
	system.Update([]*Entity{companion}, 0.016)

	if _, exists := system.lastBonusState[companion.ID]; exists {
		t.Error("state not cleaned up after bonus removal")
	}
}

func TestWeatherCompanionBonusParticleSystem_ParticleConfigs(t *testing.T) {
	system := NewWeatherCompanionBonusParticleSystem(nil, 12345)
	system.SetGenre("fantasy")

	// Test attack bonus config
	attackConfig := system.getAttackBonusConfig(1000, 1.25, "Rain")
	if attackConfig.Count < 5 {
		t.Error("attack config count too low")
	}
	if attackConfig.Gravity >= 0 {
		t.Error("attack config should have negative gravity (rise)")
	}

	// Test high attack bonus increases count
	highAttackConfig := system.getAttackBonusConfig(1000, 1.30, "Rain")
	if highAttackConfig.Count <= attackConfig.Count {
		t.Error("high attack bonus should increase particle count")
	}

	// Test defense bonus config
	defenseConfig := system.getDefenseBonusConfig(1000, 1.20, "Snow")
	if defenseConfig.Count < 6 {
		t.Error("defense config count too low")
	}
	if defenseConfig.Gravity != 0 {
		t.Error("defense config should hover (zero gravity)")
	}

	// Test defense penalty config
	penaltyConfig := system.getDefensePenaltyConfig(1000, 0.80, "Rain")
	if penaltyConfig.Gravity <= 0 {
		t.Error("penalty config should have positive gravity (fall)")
	}

	// Test speed bonus config
	speedConfig := system.getSpeedBonusConfig(1000, 1.15, "Fog")
	if speedConfig.SpreadX < 20 {
		t.Error("speed config should have wide horizontal spread")
	}

	// Test speed penalty config
	speedPenaltyConfig := system.getSpeedPenaltyConfig(1000, 0.80, "Dust")
	if speedPenaltyConfig.Gravity <= 0 {
		t.Error("speed penalty config should have positive gravity (drag)")
	}

	// Test pulse config
	pulseConfig := system.getPulseConfig(1000, &WeatherCompanionBonusComponent{
		AttackBonus: 1.20,
		WeatherType: "Rain",
	})
	if pulseConfig.Count > 3 {
		t.Error("pulse config should have minimal particles")
	}
}

func TestWeatherCompanionBonusParticleSystem_AllCompanionTypes(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	companionTypes := []CompanionType{
		CompanionTypePet,
		CompanionTypeElemental,
		CompanionTypeSpirit,
		CompanionTypeUndead,
		CompanionTypeRobot,
		CompanionTypeInsect,
	}

	for _, compType := range companionTypes {
		t.Run(compType.String(), func(t *testing.T) {
			companion := NewEntity()
			companion.AddComponent(&PositionComponent{X: 100, Y: 100})
			companion.AddComponent(&CompanionComponent{
				CompType: compType,
				OwnerID:  1,
			})
			companion.AddComponent(&WeatherCompanionBonusComponent{
				AttackBonus:       1.15,
				DefenseBonus:      1.10,
				SpeedBonus:        1.05,
				WeatherType:       "Rain",
				CompanionTypeName: compType.String(),
			})

			// Should process without error
			system.Update([]*Entity{companion}, 0.016)

			if _, exists := system.lastBonusState[companion.ID]; !exists {
				t.Errorf("state not cached for companion type %s", compType.String())
			}
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_AllWeatherTypes(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	weatherTypes := []string{
		"Rain", "Snow", "Fog", "Dust", "Ash",
		"NeonRain", "Smog", "Radiation", "BloodRain",
	}

	for _, weatherType := range weatherTypes {
		t.Run(weatherType, func(t *testing.T) {
			companion := NewEntity()
			companion.AddComponent(&PositionComponent{X: 100, Y: 100})
			companion.AddComponent(&CompanionComponent{
				CompType: CompanionTypeElemental,
				OwnerID:  1,
			})
			companion.AddComponent(&WeatherCompanionBonusComponent{
				AttackBonus:       1.20,
				DefenseBonus:      1.15,
				SpeedBonus:        1.10,
				WeatherType:       weatherType,
				CompanionTypeName: "Elemental",
			})

			// Should process without error for all weather types
			system.Update([]*Entity{companion}, 0.016)

			if _, exists := system.lastBonusState[companion.ID]; !exists {
				t.Errorf("state not cached for weather type %s", weatherType)
			}

			// Clear for next iteration
			delete(system.lastBonusState, companion.ID)
		})
	}
}

func TestWeatherCompanionBonusParticleSystem_AllGenres(t *testing.T) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system.SetGenre(genre)

			companion := NewEntity()
			companion.AddComponent(&PositionComponent{X: 100, Y: 100})
			companion.AddComponent(&CompanionComponent{
				CompType: CompanionTypeSpirit,
				OwnerID:  1,
			})
			companion.AddComponent(&WeatherCompanionBonusComponent{
				AttackBonus:       1.25,
				DefenseBonus:      1.20,
				SpeedBonus:        1.15,
				WeatherType:       "Fog",
				CompanionTypeName: "Spirit",
			})

			// Should process without error for all genres
			system.Update([]*Entity{companion}, 0.016)

			if _, exists := system.lastBonusState[companion.ID]; !exists {
				t.Errorf("state not cached for genre %s", genre)
			}

			// Clear for next iteration
			delete(system.lastBonusState, companion.ID)
		})
	}
}

func BenchmarkWeatherCompanionBonusParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherCompanionBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)

	// Create test companions
	entities := make([]*Entity, 50)
	for i := 0; i < 50; i++ {
		companion := NewEntity()
		companion.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		companion.AddComponent(&CompanionComponent{
			CompType: CompanionType(i % 6),
			OwnerID:  uint64(i),
		})
		companion.AddComponent(&WeatherCompanionBonusComponent{
			AttackBonus:       1.0 + float64(i%30)*0.01,
			DefenseBonus:      1.0 + float64(i%20)*0.01,
			SpeedBonus:        1.0 + float64(i%15)*0.01,
			WeatherType:       "Rain",
			CompanionTypeName: "Test",
		})
		entities[i] = companion
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
