//go:build ignore

package engine

import (
	"testing"
)

func TestNewReputationMovementSpeedParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewReputationMovementSpeedParticleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
	if sys.particleCount != 5 {
		t.Errorf("particleCount = %d, want 5", sys.particleCount)
	}
}

func TestReputationMovementSpeedParticleSystem_SetParticleSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem(world, 12345)

	sys.SetParticleSystem(ps)

	if sys.particleSys != ps {
		t.Error("ParticleSystem not set correctly")
	}
}

func TestReputationMovementSpeedParticleSystem_SetSpeedSystem(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ss := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetSpeedSystem(ss)

	if sys.speedSystem != ss {
		t.Error("SpeedSystem not set correctly")
	}
}

func TestReputationMovementSpeedParticleSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)

	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("genreID = %s, want scifi", sys.genreID)
	}
}

func TestReputationMovementSpeedParticleSystem_Update_NilSystems(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Should not panic with nil particle/speed systems
	sys.Update([]*Entity{player}, 0.016)
}

func TestReputationMovementSpeedParticleSystem_Update_RespectInterval(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem(world, 12345)
	ss := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetParticleSystem(ps)
	sys.SetSpeedSystem(ss)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Small delta should not trigger emit
	sys.Update([]*Entity{player}, 0.016)

	// timeSinceEmit should accumulate but not reset
	if sys.timeSinceEmit < 0.015 {
		t.Errorf("timeSinceEmit = %f, should have accumulated", sys.timeSinceEmit)
	}
}

func TestReputationMovementSpeedParticleSystem_Update_SkipsNonPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem(world, 12345)
	ss := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetParticleSystem(ps)
	sys.SetSpeedSystem(ss)

	npc := NewEntity(0)
	npc.AddComponent(&VelocityComponent{VX: 50.0, VY: 25.0})
	world.AddEntity(npc)

	sys.timeSinceEmit = 1.0
	// Should not panic on entity without input
	sys.Update([]*Entity{npc}, 0.016)
}

func TestReputationMovementSpeedParticleSystem_Update_SkipsLowBonus(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem(world, 12345)
	ss := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetParticleSystem(ps)
	sys.SetSpeedSystem(ss)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 200})
	world.AddEntity(player)

	// Speed bonus is 0 (no faction system set)
	sys.timeSinceEmit = 1.0
	sys.Update([]*Entity{player}, 0.016)
	// Should not spawn particles for 0 bonus
}

func TestReputationMovementSpeedParticleSystem_GetParticleType(t *testing.T) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)

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
			sys.SetGenre(tt.genre)
			pt := sys.getParticleTypeForGenre()
			// Just verify it returns a valid type without panicking
			if pt < 0 {
				t.Errorf("Invalid particle type for genre %s", tt.genre)
			}
		})
	}
}

func BenchmarkReputationMovementSpeedParticleSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewReputationMovementSpeedParticleSystem(world, 12345)
	ps := NewParticleSystem(world, 12345)
	ss := NewReputationMovementSpeedSystem(world, 12345)

	sys.SetParticleSystem(ps)
	sys.SetSpeedSystem(ss)

	player := NewEntity(0)
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 200})
	world.AddEntity(player)

	entities := []*Entity{player}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceEmit = 1.0
		sys.Update(entities, 0.016)
	}
}
