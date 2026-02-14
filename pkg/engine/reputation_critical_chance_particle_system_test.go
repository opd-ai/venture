//go:build ignore

package engine

import (
	"testing"
)

func TestNewReputationCriticalChanceParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationCriticalChanceParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.particleCount != 4 {
		t.Errorf("particleCount = %d, want 4", sys.particleCount)
	}
	if sys.emitInterval != 1.2 {
		t.Errorf("emitInterval = %f, want 1.2", sys.emitInterval)
	}
}

func TestReputationCriticalChanceParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	ps := &ParticleSystem{}

	sys.SetParticleSystem(ps)

	if sys.partSys != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestReputationCriticalChanceParticleSystem_SetCritSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	critSys := NewReputationCriticalChanceBonusSystem(world, 12345)

	sys.SetCritSystem(critSys)

	if sys.critSys != critSys {
		t.Error("CritSystem not set correctly")
	}
}

func TestReputationCriticalChanceParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("GenreID = %s, want horror", sys.genreID)
	}
}

func TestReputationCriticalChanceParticleSystem_Update_NilDeps(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)

	// Should not panic with nil dependencies
	sys.Update([]*Entity{}, 0.016)
}

func TestReputationCriticalChanceParticleSystem_Update_RespectInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	sys.SetParticleSystem(&ParticleSystem{})
	critSys := NewReputationCriticalChanceBonusSystem(world, 12345)
	sys.SetCritSystem(critSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Should not emit - interval not reached
	sys.Update([]*Entity{player}, 0.016)

	if sys.timeSinceEmit < 0.015 {
		t.Error("timeSinceEmit should accumulate")
	}
}

func TestReputationCriticalChanceParticleSystem_GenreParticleTypes(t *testing.T) {
	world := NewWorld()

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
			sys := NewReputationCriticalChanceParticleSystem(world, 12345)
			sys.SetGenre(tt.genre)
			// Should not panic
			pType := sys.getParticleTypeForGenre()
			if pType < 0 {
				t.Error("invalid particle type")
			}
		})
	}
}

func TestReputationCriticalChanceParticleSystem_Update_SkipsNonPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	sys.SetParticleSystem(&ParticleSystem{})
	critSys := NewReputationCriticalChanceBonusSystem(world, 12345)
	sys.SetCritSystem(critSys)

	npc := NewEntity(0)
	npc.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.timeSinceEmit = 2.0
	// Should not panic - skips entities without input
	sys.Update([]*Entity{npc}, 0.016)
}

func TestReputationCriticalChanceParticleSystem_Update_SkipsLowBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	sys.SetParticleSystem(&ParticleSystem{})
	critSys := NewReputationCriticalChanceBonusSystem(world, 12345)
	sys.SetCritSystem(critSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	sys.timeSinceEmit = 2.0
	// critSys has no bonuses applied, so bonus is 0
	sys.Update([]*Entity{player}, 0.016)
	// Should complete without spawning (no error)
}

func BenchmarkReputationCriticalChanceParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationCriticalChanceParticleSystem(world, 12345)
	sys.SetParticleSystem(&ParticleSystem{})
	critSys := NewReputationCriticalChanceBonusSystem(world, 12345)
	sys.SetCritSystem(critSys)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceEmit = 2.0
		sys.Update(entities, 0.016)
	}
}
