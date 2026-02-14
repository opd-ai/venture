package engine

import (
	"testing"
)

func TestNewTerrainCombatBonusParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewTerrainCombatBonusParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.seed != 12345 {
		t.Errorf("Seed = %d, want 12345", sys.seed)
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Default genre = %q, want fantasy", sys.genreID)
	}
	if sys.pulseInterval != 2.0 {
		t.Errorf("Pulse interval = %f, want 2.0", sys.pulseInterval)
	}
}

func TestTerrainCombatBonusParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()

	sys.SetParticleSystem(ps)

	if sys.particleSystem != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestTerrainCombatBonusParticleSystem_SetTerrainCombatBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	tcbs := NewTerrainCombatBonusSystem(world, 12345)

	sys.SetTerrainCombatBonusSystem(tcbs)

	if sys.terrainCombatBonusSystem != tcbs {
		t.Error("TerrainCombatBonusSystem not set correctly")
	}
}

func TestTerrainCombatBonusParticleSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewTerrainCombatBonusParticleSystem(world, 12345)

			sys.SetGenre(tt.genreID)

			if sys.genreID != tt.genreID {
				t.Errorf("Genre = %q, want %q", sys.genreID, tt.genreID)
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_Update_NilParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	// Don't set particle system

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.10,
		DefenseBonus: 1.0,
		EvasionBonus: 0.0,
		TerrainType:  "platform",
	})

	// Should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestTerrainCombatBonusParticleSystem_Update_NilWorld(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Should not panic
	sys.Update([]*Entity{}, 0.016)
}

func TestTerrainCombatBonusParticleSystem_Update_EntityWithoutBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	// No TerrainCombatBonusComponent

	sys.Update([]*Entity{entity}, 0.016)

	// Should process without error
	if _, exists := sys.lastBonusState[entity.ID]; exists {
		t.Error("Should not cache state for entity without bonus")
	}
}

func TestTerrainCombatBonusParticleSystem_Update_EntityWithDamageBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.10,
		DefenseBonus: 1.0,
		EvasionBonus: 0.0,
		TerrainType:  "platform",
	})

	sys.Update([]*Entity{entity}, 0.016)

	// Should cache state
	state, exists := sys.lastBonusState[entity.ID]
	if !exists {
		t.Fatal("State not cached for entity with bonus")
	}
	if !state.hasDamageBonus {
		t.Error("hasDamageBonus should be true")
	}
	if state.terrainType != "platform" {
		t.Errorf("TerrainType = %q, want platform", state.terrainType)
	}
}

func TestTerrainCombatBonusParticleSystem_Update_EntityWithEvasionBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.0,
		DefenseBonus: 1.0,
		EvasionBonus: 0.10, // Cover bonus
		TerrainType:  "wall_cover",
	})

	sys.Update([]*Entity{entity}, 0.016)

	state, exists := sys.lastBonusState[entity.ID]
	if !exists {
		t.Fatal("State not cached")
	}
	if !state.hasEvasionBonus {
		t.Error("hasEvasionBonus should be true")
	}
}

func TestTerrainCombatBonusParticleSystem_Update_EntityWithDefenseDebuff(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.0,
		DefenseBonus: 0.85, // Water penalty
		EvasionBonus: 0.0,
		TerrainType:  "shallow_water",
	})

	sys.Update([]*Entity{entity}, 0.016)

	state, exists := sys.lastBonusState[entity.ID]
	if !exists {
		t.Fatal("State not cached")
	}
	if !state.hasDefenseDebuf {
		t.Error("hasDefenseDebuf should be true")
	}
}

func TestTerrainCombatBonusParticleSystem_Update_StateChange(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	bonus := &TerrainCombatBonusComponent{
		DamageBonus:  1.0,
		DefenseBonus: 1.0,
		EvasionBonus: 0.0,
		TerrainType:  "floor",
	}
	entity.AddComponent(bonus)

	// First update - no bonus active
	sys.Update([]*Entity{entity}, 0.016)

	// Change to high ground
	bonus.DamageBonus = 1.10
	bonus.TerrainType = "platform"

	// Second update - should detect state change
	sys.Update([]*Entity{entity}, 0.016)

	state := sys.lastBonusState[entity.ID]
	if !state.hasDamageBonus {
		t.Error("Should detect damage bonus change")
	}
	if state.terrainType != "platform" {
		t.Error("Should update terrain type")
	}
}

func TestTerrainCombatBonusParticleSystem_Update_PulseInterval(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.10,
		DefenseBonus: 1.0,
		EvasionBonus: 0.0,
		TerrainType:  "platform",
	})

	// First update caches state
	sys.Update([]*Entity{entity}, 0.016)

	// Update without reaching pulse interval
	sys.Update([]*Entity{entity}, 1.0)
	if sys.timeSinceEmit >= sys.pulseInterval {
		t.Errorf("Should not reach pulse yet: %f", sys.timeSinceEmit)
	}

	// Update past pulse interval
	sys.Update([]*Entity{entity}, 2.0)
	if sys.timeSinceEmit >= sys.pulseInterval {
		t.Error("Timer should reset after pulse")
	}
}

func TestTerrainCombatBonusParticleSystem_Update_EntityLosesBonus(t *testing.T) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	bonus := &TerrainCombatBonusComponent{
		DamageBonus:  1.10,
		DefenseBonus: 1.0,
		EvasionBonus: 0.0,
		TerrainType:  "platform",
	}
	entity.AddComponent(bonus)

	// Update with bonus
	sys.Update([]*Entity{entity}, 0.016)
	if _, exists := sys.lastBonusState[entity.ID]; !exists {
		t.Fatal("State should be cached")
	}

	// Remove bonus component
	entity.RemoveComponent("terrain_combat_bonus")

	// Update without bonus
	sys.Update([]*Entity{entity}, 0.016)

	// State should be cleaned up
	if _, exists := sys.lastBonusState[entity.ID]; exists {
		t.Error("State should be cleaned up when bonus removed")
	}
}

func TestTerrainCombatBonusParticleSystem_bonusStateChanged(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)

	tests := []struct {
		name    string
		last    bonusStateCache
		current bonusStateCache
		want    bool
	}{
		{
			name:    "no change",
			last:    bonusStateCache{hasDamageBonus: true, terrainType: "platform"},
			current: bonusStateCache{hasDamageBonus: true, terrainType: "platform"},
			want:    false,
		},
		{
			name:    "damage bonus changed",
			last:    bonusStateCache{hasDamageBonus: false},
			current: bonusStateCache{hasDamageBonus: true},
			want:    true,
		},
		{
			name:    "defense debuff changed",
			last:    bonusStateCache{hasDefenseDebuf: false},
			current: bonusStateCache{hasDefenseDebuf: true},
			want:    true,
		},
		{
			name:    "evasion changed",
			last:    bonusStateCache{hasEvasionBonus: false},
			current: bonusStateCache{hasEvasionBonus: true},
			want:    true,
		},
		{
			name:    "terrain type changed",
			last:    bonusStateCache{terrainType: "floor"},
			current: bonusStateCache{terrainType: "platform"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.bonusStateChanged(tt.last, tt.current)
			if got != tt.want {
				t.Errorf("bonusStateChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_hasActiveBonus(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)

	tests := []struct {
		name  string
		state bonusStateCache
		want  bool
	}{
		{"no bonus", bonusStateCache{}, false},
		{"damage bonus", bonusStateCache{hasDamageBonus: true}, true},
		{"defense buff", bonusStateCache{hasDefenseBuff: true}, true},
		{"defense debuff", bonusStateCache{hasDefenseDebuf: true}, true},
		{"evasion bonus", bonusStateCache{hasEvasionBonus: true}, true},
		{"multiple bonuses", bonusStateCache{hasDamageBonus: true, hasEvasionBonus: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.hasActiveBonus(tt.state)
			if got != tt.want {
				t.Errorf("hasActiveBonus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getHighGroundParticleType(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewTerrainCombatBonusParticleSystem(nil, 12345)
			sys.SetGenre(tt.genre)

			pType := sys.getHighGroundParticleType()
			// Just verify it returns a valid type without panicking
			if pType < 0 {
				t.Error("Invalid particle type")
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getCoverParticleType(t *testing.T) {
	tests := []struct {
		genre string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewTerrainCombatBonusParticleSystem(nil, 12345)
			sys.SetGenre(tt.genre)

			pType := sys.getCoverParticleType()
			if pType < 0 {
				t.Error("Invalid particle type")
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getHighGroundConfig(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)

	tests := []struct {
		name        string
		damageBonus float64
		wantCount   int
	}{
		{"low bonus", 1.05, 6},
		{"high bonus", 1.10, 8},
		{"very high bonus", 1.15, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := sys.getHighGroundConfig(12345, tt.damageBonus)
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.Gravity >= 0 {
				t.Error("High ground particles should rise (negative gravity)")
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getCoverBonusConfig(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)

	tests := []struct {
		name         string
		evasionBonus float64
		wantCount    int
	}{
		{"low evasion", 0.05, 5},
		{"high evasion", 0.10, 7},
		{"very high evasion", 0.15, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := sys.getCoverBonusConfig(12345, tt.evasionBonus)
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.Gravity != 0 {
				t.Error("Cover particles should hover (zero gravity)")
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getWaterPenaltyConfig(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)

	tests := []struct {
		name         string
		defenseBonus float64
		wantCount    int
	}{
		{"mild penalty", 0.85, 4},
		{"severe penalty", 0.80, 6},
		{"very severe", 0.75, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := sys.getWaterPenaltyConfig(12345, tt.defenseBonus)
			if config.Count != tt.wantCount {
				t.Errorf("Count = %d, want %d", config.Count, tt.wantCount)
			}
			if config.Gravity <= 0 {
				t.Error("Water penalty particles should fall (positive gravity)")
			}
		})
	}
}

func TestTerrainCombatBonusParticleSystem_getPulseConfig(t *testing.T) {
	sys := NewTerrainCombatBonusParticleSystem(nil, 12345)
	bonus := &TerrainCombatBonusComponent{
		DamageBonus: 1.10,
	}

	config := sys.getPulseConfig(12345, bonus)

	if config.Count != 2 {
		t.Errorf("Pulse count = %d, want 2", config.Count)
	}
	if config.Duration != 0.4 {
		t.Errorf("Pulse duration = %f, want 0.4", config.Duration)
	}
}

func BenchmarkTerrainCombatBonusParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	// Create entities with terrain bonuses
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 32), Y: float64(i * 32)})
		entity.AddComponent(&TerrainCombatBonusComponent{
			DamageBonus:  1.10,
			DefenseBonus: 1.0,
			EvasionBonus: 0.10,
			TerrainType:  "platform",
		})
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkTerrainCombatBonusParticleSystem_ProcessEntity(b *testing.B) {
	world := NewWorld()
	sys := NewTerrainCombatBonusParticleSystem(world, 12345)
	ps := NewParticleSystem()
	sys.SetParticleSystem(ps)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&TerrainCombatBonusComponent{
		DamageBonus:  1.10,
		DefenseBonus: 0.85,
		EvasionBonus: 0.10,
		TerrainType:  "platform",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.processEntity(entity, false)
	}
}
