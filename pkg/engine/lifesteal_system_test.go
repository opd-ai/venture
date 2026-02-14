package engine

import (
	"testing"
)

func getHealthComp(entity *Entity) *HealthComponent {
	comp, ok := entity.GetComponent("health")
	if !ok || comp == nil {
		return nil
	}
	return comp.(*HealthComponent)
}

func TestNewLifestealSystem(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	if system == nil {
		t.Fatal("NewLifestealSystem returned nil")
	}
	if system.world != world {
		t.Error("world reference not set correctly")
	}
	if system.seed != 12345 {
		t.Errorf("seed = %d, want 12345", system.seed)
	}
	if system.rng == nil {
		t.Error("rng not initialized")
	}
	if system.particleCount != 8 {
		t.Errorf("particleCount = %d, want 8", system.particleCount)
	}
}

func TestLifestealSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)
	ps := NewParticleSystem()
	system.SetParticleSystem(ps)
	if system.particleSystem != ps {
		t.Error("particle system not set")
	}
}

func TestLifestealSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)
	system.SetGenre("fantasy")
	if system.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", system.genreID, "fantasy")
	}
}

func TestLifestealSystem_OnDamageDealt_NoLifesteal(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 0
	attacker.AddComponent(stats)
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 100)

	health := getHealthComp(attacker)
	if health.Current != 50 {
		t.Errorf("health changed without lifesteal: got %f, want 50", health.Current)
	}
}

func TestLifestealSystem_OnDamageDealt_WithLifesteal(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 0.2
	attacker.AddComponent(stats)
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 100)

	health := getHealthComp(attacker)
	if health.Current != 70 {
		t.Errorf("health after lifesteal = %f, want 70", health.Current)
	}
}

func TestLifestealSystem_OnDamageDealt_CappedHealing(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 1.0
	attacker.AddComponent(stats)
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 200)

	health := getHealthComp(attacker)
	if health.Current != 75 {
		t.Errorf("health after capped lifesteal = %f, want 75", health.Current)
	}
}

func TestLifestealSystem_OnDamageDealt_NilAttacker(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)
	target := world.CreateEntity()
	system.OnDamageDealt(nil, target, 100)
}

func TestLifestealSystem_OnDamageDealt_ZeroDamage(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 0.5
	attacker.AddComponent(stats)
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 0)

	health := getHealthComp(attacker)
	if health.Current != 50 {
		t.Errorf("health changed with zero damage: got %f, want 50", health.Current)
	}
}

func TestLifestealSystem_OnDamageDealt_MissingHealthComponent(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 0.5
	attacker.AddComponent(stats)

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 100)
}

func TestLifestealSystem_OnDamageDealt_MissingStatsComponent(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})

	target := world.CreateEntity()
	system.OnDamageDealt(attacker, target, 100)

	health := getHealthComp(attacker)
	if health.Current != 50 {
		t.Errorf("health changed without stats: got %f, want 50", health.Current)
	}
}

func TestLifestealSystem_GetLifestealAmount(t *testing.T) {
	tests := []struct {
		name      string
		lifesteal float64
		damage    float64
		expected  float64
	}{
		{"no lifesteal", 0, 100, 0},
		{"20% lifesteal", 0.2, 100, 20},
		{"50% lifesteal", 0.5, 50, 25},
		{"zero damage", 0.5, 0, 0},
		{"negative damage", 0.5, -10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewLifestealSystem(world, 12345)
			attacker := world.CreateEntity()
			stats := NewStatsComponent()
			stats.Lifesteal = tt.lifesteal
			attacker.AddComponent(stats)

			result := system.GetLifestealAmount(attacker, tt.damage)
			if result != tt.expected {
				t.Errorf("GetLifestealAmount() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestLifestealSystem_GetLifestealAmount_NilAttacker(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)
	result := system.GetLifestealAmount(nil, 100)
	if result != 0 {
		t.Errorf("GetLifestealAmount(nil) = %f, want 0", result)
	}
}

func TestLifestealSystem_Update_NoOp(t *testing.T) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)
	entities := []*Entity{world.CreateEntity(), world.CreateEntity()}
	system.Update(entities, 0.016)
}

func BenchmarkLifestealSystem_OnDamageDealt(b *testing.B) {
	world := NewWorld()
	system := NewLifestealSystem(world, 12345)

	attacker := world.CreateEntity()
	stats := NewStatsComponent()
	stats.Lifesteal = 0.2
	attacker.AddComponent(stats)
	attacker.AddComponent(&HealthComponent{Current: 50, Max: 100})
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := world.CreateEntity()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		health := getHealthComp(attacker)
		health.Current = 50
		system.OnDamageDealt(attacker, target, 100)
	}
}
