package engine

import (
	"math/rand"
	"testing"
)

func TestNewWeatherMovementSpeedParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)

	if system == nil {
		t.Fatal("NewWeatherMovementSpeedParticleSystem returned nil")
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
	if system.genreID != "fantasy" {
		t.Errorf("genreID = %s, want fantasy", system.genreID)
	}
	if system.pulseInterval <= 0 {
		t.Error("pulseInterval should be positive")
	}
	if system.lastModifierState == nil {
		t.Error("lastModifierState map not initialized")
	}
}

func TestWeatherMovementSpeedParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem()

	system.SetParticleSystem(ps)

	if system.particleSystem != ps {
		t.Error("particleSystem not set correctly")
	}
}

func TestWeatherMovementSpeedParticleSystem_SetWeatherMovementSpeedSystem(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	wmss := NewWeatherMovementSpeedSystem(world, 12345)

	system.SetWeatherMovementSpeedSystem(wmss)

	if system.weatherMovementSpeedSystem != wmss {
		t.Error("weatherMovementSpeedSystem not set correctly")
	}
}

func TestWeatherMovementSpeedParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)

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

func TestWeatherMovementSpeedParticleSystem_Update_NilDependencies(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)

	// Should not panic with nil dependencies
	entity := NewEntity(1)
	entities := []*Entity{entity}

	// No particle system
	system.Update(entities, 0.016)

	// Add particle system but no weather movement speed system
	system.SetParticleSystem(NewParticleSystem())
	system.Update(entities, 0.016)

	// Should handle gracefully
}

func TestWeatherMovementSpeedParticleSystem_Update_NoVelocity(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetWeatherMovementSpeedSystem(NewWeatherMovementSpeedSystem(world, 12345))

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No velocity component

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should skip entity without velocity
	if _, exists := system.lastModifierState[entity.ID]; exists {
		t.Error("should not track entity without velocity")
	}
}

func TestWeatherMovementSpeedParticleSystem_Update_SlowEntity(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	wmss := NewWeatherMovementSpeedSystem(world, 12345)
	system.SetWeatherMovementSpeedSystem(wmss)

	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{VX: 1, VY: 1}) // Very slow
	world.AddEntity(entity)

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should skip slow entities
	if _, exists := system.lastModifierState[entity.ID]; exists {
		t.Error("should not track slow entity")
	}
}

func TestWeatherMovementSpeedParticleSystem_Update_NoPosition(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	system.SetWeatherMovementSpeedSystem(NewWeatherMovementSpeedSystem(world, 12345))

	entity := NewEntity(1)
	entity.AddComponent(&VelocityComponent{VX: 50, VY: 50}) // Fast enough
	// No position component

	entities := []*Entity{entity}
	system.Update(entities, 0.016)

	// Should skip entity without position
	if _, exists := system.lastModifierState[entity.ID]; exists {
		t.Error("should not track entity without position")
	}
}

func TestWeatherMovementSpeedParticleSystem_ModifiersEqual(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want bool
	}{
		{"equal", 1.0, 1.0, true},
		{"close", 1.0, 1.005, true},
		{"different", 1.0, 1.1, false},
		{"negative difference", 1.1, 1.0, false},
		{"penalty equal", 0.85, 0.85, true},
		{"penalty close", 0.85, 0.855, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := modifiersEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("modifiersEqual(%f, %f) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestWeatherMovementSpeedParticleSystem_GetBoostParticleType(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "cyberpunk", "horror", "postapoc", "unknown"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system.SetGenre(genre)
			ptype := system.getBoostParticleType()
			// Just verify it returns a valid type (not zero value)
			if ptype == 0 {
				t.Errorf("getBoostParticleType() returned zero for genre %s", genre)
			}
		})
	}
}

func TestWeatherMovementSpeedParticleSystem_GetPenaltyParticleType(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "cyberpunk", "horror", "postapoc", "unknown"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			system.SetGenre(genre)
			ptype := system.getPenaltyParticleType()
			// Just verify it returns a valid type (not zero value)
			if ptype == 0 {
				t.Errorf("getPenaltyParticleType() returned zero for genre %s", genre)
			}
		})
	}
}

func TestWeatherMovementSpeedParticleSystem_GetSpeedBoostConfig(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	vel := &VelocityComponent{VX: 50, VY: 0}

	tests := []struct {
		name      string
		modifier  float64
		wantCount int
	}{
		{"small bonus", 1.05, 4},
		{"large bonus", 1.10, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := system.getSpeedBoostConfig(12345, tt.modifier, vel)
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.Gravity >= 0 {
				t.Error("boost gravity should be negative (upward)")
			}
			customVal, ok := config.Custom["weather_speed_boost"]
			if !ok || customVal != true {
				t.Error("missing weather_speed_boost custom flag")
			}
		})
	}
}

func TestWeatherMovementSpeedParticleSystem_GetSpeedPenaltyConfig(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetGenre("horror")

	vel := &VelocityComponent{VX: 50, VY: 0}

	tests := []struct {
		name      string
		modifier  float64
		wantCount int
	}{
		{"small penalty", 0.90, 3},
		{"large penalty", 0.80, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := system.getSpeedPenaltyConfig(12345, tt.modifier, vel)
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.Gravity <= 0 {
				t.Error("penalty gravity should be positive (downward)")
			}
			customVal, ok := config.Custom["weather_speed_penalty"]
			if !ok || customVal != true {
				t.Error("missing weather_speed_penalty custom flag")
			}
		})
	}
}

func TestWeatherMovementSpeedParticleSystem_GetPulseConfig(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetGenre("scifi")

	vel := &VelocityComponent{VX: 50, VY: 0}

	// Test bonus pulse
	bonusConfig := system.getPulseConfig(12345, 1.05, true, vel)
	if bonusConfig.Count != 2 {
		t.Errorf("bonus pulse count = %d, want 2", bonusConfig.Count)
	}
	if bonusConfig.Gravity >= 0 {
		t.Error("bonus pulse gravity should be negative")
	}

	// Test penalty pulse
	penaltyConfig := system.getPulseConfig(12345, 0.85, false, vel)
	if penaltyConfig.Count != 2 {
		t.Errorf("penalty pulse count = %d, want 2", penaltyConfig.Count)
	}
	if penaltyConfig.Gravity <= 0 {
		t.Error("penalty pulse gravity should be positive")
	}
}

func TestWeatherMovementSpeedParticleSystem_GetModifierConfig(t *testing.T) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetGenre("fantasy")

	vel := &VelocityComponent{VX: 50, VY: 0}

	// Test bonus config
	bonusConfig := system.getModifierConfig(12345, 1.05, true, vel)
	if _, ok := bonusConfig.Custom["weather_speed_boost"]; !ok {
		t.Error("bonus config should have weather_speed_boost flag")
	}

	// Test penalty config
	penaltyConfig := system.getModifierConfig(12345, 0.85, false, vel)
	if _, ok := penaltyConfig.Custom["weather_speed_penalty"]; !ok {
		t.Error("penalty config should have weather_speed_penalty flag")
	}
}

func TestWeatherMovementSpeedParticleSystem_Determinism(t *testing.T) {
	// Test that same seed produces same particle configs
	world1 := NewWorld()
	system1 := NewWeatherMovementSpeedParticleSystem(world1, 42)
	system1.SetGenre("fantasy")

	world2 := NewWorld()
	system2 := NewWeatherMovementSpeedParticleSystem(world2, 42)
	system2.SetGenre("fantasy")

	vel := &VelocityComponent{VX: 50, VY: 0}

	config1 := system1.getSpeedBoostConfig(100, 1.05, vel)
	config2 := system2.getSpeedBoostConfig(100, 1.05, vel)

	if config1.Duration != config2.Duration {
		t.Error("determinism failure: duration differs")
	}
	if config1.Count != config2.Count {
		t.Error("determinism failure: count differs")
	}
	if config1.Gravity != config2.Gravity {
		t.Error("determinism failure: gravity differs")
	}
}

func BenchmarkWeatherMovementSpeedParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetParticleSystem(NewParticleSystem())
	wmss := NewWeatherMovementSpeedSystem(world, 12345)
	system.SetWeatherMovementSpeedSystem(wmss)
	system.SetGenre("fantasy")

	// Create 100 moving entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		e := NewEntity(uint64(i + 1))
		e.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		e.AddComponent(&VelocityComponent{VX: 50 + float64(i), VY: 30 + float64(i)})
		entities[i] = e
		world.AddEntity(e)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}

func BenchmarkWeatherMovementSpeedParticleSystem_GetConfig(b *testing.B) {
	world := NewWorld()
	system := NewWeatherMovementSpeedParticleSystem(world, 12345)
	system.SetGenre("cyberpunk")

	rng := rand.New(rand.NewSource(12345))
	vel := &VelocityComponent{VX: 50, VY: 30}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed := int64(rng.Intn(10000))
		modifier := 0.8 + rng.Float64()*0.4 // 0.8 to 1.2
		isBonus := modifier > 1.0
		_ = system.getModifierConfig(seed, modifier, isBonus, vel)
	}
}
