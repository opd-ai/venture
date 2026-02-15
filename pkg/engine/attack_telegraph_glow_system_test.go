package engine

import (
	"testing"
)

// stubTelegraphInputComponent is a tag component identifying the player entity.
type stubTelegraphInputComponent struct{}

func (s *stubTelegraphInputComponent) Type() string { return "input" }

func TestAttackTelegraphComponent_Type(t *testing.T) {
	c := NewAttackTelegraphComponent()
	if c.Type() != "attack_telegraph" {
		t.Errorf("expected type 'attack_telegraph', got %q", c.Type())
	}
}

func TestAttackTelegraphComponent_Defaults(t *testing.T) {
	c := NewAttackTelegraphComponent()
	if c.Active {
		t.Error("expected Active=false by default")
	}
	if c.Intensity != 0 {
		t.Errorf("expected Intensity=0, got %f", c.Intensity)
	}
	if c.Radius != 16.0 {
		t.Errorf("expected Radius=16.0, got %f", c.Radius)
	}
}

func TestAttackTelegraphGlowSystem_SetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		wantR   float64
	}{
		{"fantasy", "fantasy", 0.9},
		{"horror", "horror", 0.85},
		{"cyberpunk", "cyberpunk", 0.9},
		{"scifi", "sci-fi", 0.1},
		{"postapoc", "postapoc", 0.9},
		{"default", "unknown", 0.9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewAttackTelegraphGlowSystem(nil, 42)
			sys.SetGenre(tt.genreID)
			if sys.preset.R != tt.wantR {
				t.Errorf("genre %q: expected R=%f, got %f", tt.genreID, tt.wantR, sys.preset.R)
			}
		})
	}
}

func TestAttackTelegraphGlowSystem_Update_NonAIIgnored(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	// Player entity (has input component) should be ignored
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&stubTelegraphInputComponent{})
	player.AddComponent(&AIComponent{State: AIStateAttack})
	player.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.2})

	sys.Update([]*Entity{player}, 0.6) // trigger fullScan

	comp, found := player.GetComponent("attack_telegraph")
	if !found {
		// No telegraph should be attached to player entities — correct
	} else {
		telegraph, ok := comp.(*AttackTelegraphComponent)
		if ok && telegraph.Active {
			t.Error("player entity should not get active attack telegraph")
		}
	}
}

func TestAttackTelegraphGlowSystem_Update_IdleAI(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateIdle})
	entity.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.2})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})

	sys.Update([]*Entity{entity}, 0.6) // fullScan

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph, ok := comp.(*AttackTelegraphComponent)
	if !ok {
		t.Fatal("expected AttackTelegraphComponent to be attached")
	}
	if telegraph.Active {
		t.Error("idle AI should not have active telegraph")
	}
}

func TestAttackTelegraphGlowSystem_Update_AttackingAI(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)
	sys.SetGenre("fantasy")

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(&AttackComponent{Cooldown: 2.0, CooldownTimer: 0.4}) // 20% remaining < 40% threshold
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})

	sys.Update([]*Entity{entity}, 0.6) // fullScan

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph, ok := comp.(*AttackTelegraphComponent)
	if !ok {
		t.Fatal("expected AttackTelegraphComponent to be attached")
	}
	if !telegraph.Active {
		t.Error("attacking AI near cooldown ready should have active telegraph")
	}
	if telegraph.Intensity <= 0 {
		t.Errorf("expected positive intensity, got %f", telegraph.Intensity)
	}
	if telegraph.Intensity > 1.0 {
		t.Errorf("expected intensity <= 1.0, got %f", telegraph.Intensity)
	}
	if telegraph.ColorR != 0.9 {
		t.Errorf("expected fantasy ColorR=0.9, got %f", telegraph.ColorR)
	}
}

func TestAttackTelegraphGlowSystem_Update_ChasingAI(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateChase})
	entity.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.1}) // 10% remaining
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})

	sys.Update([]*Entity{entity}, 0.6) // fullScan

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph, ok := comp.(*AttackTelegraphComponent)
	if !ok {
		t.Fatal("expected AttackTelegraphComponent to be attached")
	}
	if !telegraph.Active {
		t.Error("chasing AI near cooldown ready should have active telegraph")
	}
}

func TestAttackTelegraphGlowSystem_Update_HighCooldownNoTelegraph(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	// 80% cooldown remaining — above 40% threshold, no telegraph
	entity.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.8})
	entity.AddComponent(&ColliderComponent{Width: 32, Height: 32})

	sys.Update([]*Entity{entity}, 0.6) // fullScan

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph, ok := comp.(*AttackTelegraphComponent)
	if !ok {
		t.Fatal("expected AttackTelegraphComponent to be attached")
	}
	if telegraph.Active {
		t.Error("high cooldown remaining should not trigger telegraph")
	}
}

func TestAttackTelegraphGlowSystem_IntensityRamping(t *testing.T) {
	tests := []struct {
		name          string
		cooldownTimer float64
		cooldown      float64
		wantActive    bool
		wantHigher    bool // true if intensity should be > 0.5
	}{
		{"just_attacked", 1.0, 1.0, false, false},     // 100% remaining
		{"mid_cooldown", 0.6, 1.0, false, false},       // 60% remaining
		{"threshold_edge", 0.4, 1.0, true, false},        // exactly at threshold → just enters telegraph
		{"below_threshold", 0.3, 1.0, true, false},      // 30% remaining → 25% through telegraph
		{"near_ready", 0.1, 1.0, true, false},            // 10% remaining → 75% through telegraph
		{"ready", 0.0, 1.0, true, true},                  // fully ready → max intensity
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewAttackTelegraphGlowSystem(world, 42)

			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 10, Y: 10})
			entity.AddComponent(&AIComponent{State: AIStateAttack})
			entity.AddComponent(&AttackComponent{Cooldown: tt.cooldown, CooldownTimer: tt.cooldownTimer})

			sys.Update([]*Entity{entity}, 0.6) // fullScan

			comp, _ := entity.GetComponent("attack_telegraph")
			telegraph, ok := comp.(*AttackTelegraphComponent)
			if !ok {
				t.Fatal("expected AttackTelegraphComponent")
			}
			if telegraph.Active != tt.wantActive {
				t.Errorf("Active: got %v, want %v", telegraph.Active, tt.wantActive)
			}
			if tt.wantHigher && telegraph.Intensity < 0.5 {
				t.Errorf("expected high intensity (>0.5), got %f", telegraph.Intensity)
			}
		})
	}
}

func TestAttackTelegraphGlowSystem_RadiusFromCollider(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.1})
	entity.AddComponent(&ColliderComponent{Width: 64, Height: 48})

	sys.Update([]*Entity{entity}, 0.6)

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph := comp.(*AttackTelegraphComponent)
	expectedRadius := 64.0 * 0.6 // max(64, 48) * 0.6
	if telegraph.Radius != expectedRadius {
		t.Errorf("expected radius %f, got %f", expectedRadius, telegraph.Radius)
	}
}

func TestAttackTelegraphGlowSystem_ZeroCooldownNoTelegraph(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(&AttackComponent{Cooldown: 0, CooldownTimer: 0}) // No cooldown configured

	sys.Update([]*Entity{entity}, 0.6)

	comp, _ := entity.GetComponent("attack_telegraph")
	telegraph, ok := comp.(*AttackTelegraphComponent)
	if !ok {
		t.Fatal("expected AttackTelegraphComponent")
	}
	if telegraph.Active {
		t.Error("zero cooldown attack should not telegraph")
	}
}

func TestAttackTelegraphGlowSystem_Throttling(t *testing.T) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 10})
	entity.AddComponent(&AIComponent{State: AIStateAttack})
	entity.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.2})

	// First update with small deltaTime — no fullScan, entity won't get component
	sys.Update([]*Entity{entity}, 0.1)
	comp, _ := entity.GetComponent("attack_telegraph")
	if comp != nil {
		t.Error("expected no telegraph on non-fullScan frame")
	}

	// Second update pushes past updateInterval threshold
	sys.Update([]*Entity{entity}, 0.5)
	comp, _ = entity.GetComponent("attack_telegraph")
	if comp == nil {
		t.Error("expected telegraph after fullScan")
	}
}

func BenchmarkAttackTelegraphGlowSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewAttackTelegraphGlowSystem(world, 42)

	entities := make([]*Entity, 200)
	for i := range entities {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		e.AddComponent(&AIComponent{State: AIStateAttack})
		e.AddComponent(&AttackComponent{Cooldown: 1.0, CooldownTimer: 0.3})
		e.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		e.AddComponent(NewAttackTelegraphComponent())
		entities[i] = e
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
