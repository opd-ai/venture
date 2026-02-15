package engine

import (
	"testing"
)

// stubInputTag is a test stub for marking entities as players (type "input").
type stubInputTag struct{}

func (s *stubInputTag) Type() string { return "input" }

func TestNPCProximityGlowComponent_Type(t *testing.T) {
	c := &NPCProximityGlowComponent{}
	if got := c.Type(); got != "npc_proximity_glow" {
		t.Errorf("Type() = %q, want %q", got, "npc_proximity_glow")
	}
}

func TestNewNPCInteractionProximityGlowSystem(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)
	if sys == nil {
		t.Fatal("NewNPCInteractionProximityGlowSystem returned nil")
	}
	if sys.GlowRadius != 80.0 {
		t.Errorf("GlowRadius = %v, want 80.0", sys.GlowRadius)
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "fantasy")
	}
}

func TestNPCInteractionProximityGlowSystem_SetGenre(t *testing.T) {
	genres := []struct {
		name  string
		wantR uint8
	}{
		{"fantasy", 255},
		{"horror", 80},
		{"scifi", 60},
		{"cyberpunk", 220},
		{"postapoc", 220},
	}
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)
	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetGenre(tt.name)
			if sys.genreID != tt.name {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.name)
			}
			if sys.preset.Color.R != tt.wantR {
				t.Errorf("preset.Color.R = %d, want %d", sys.preset.Color.R, tt.wantR)
			}
		})
	}
}

func TestNPCInteractionProximityGlowSystem_Update_NoEntities(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)
	// Should not panic with empty entity list
	sys.Update(nil, 0.016)
	sys.Update([]*Entity{}, 0.016)
}

func TestNPCInteractionProximityGlowSystem_Update_GlowActivates(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	// Create player entity at (50,50)
	player := NewEntity(1000)
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&stubInputTag{})

	// Create NPC with dialog at (70,50) — 20px away, within inner radius
	npc := NewEntity(1001)
	npc.AddComponent(&PositionComponent{X: 70, Y: 50})
	npc.AddComponent(NewDialogComponent(nil))
	npc.AddComponent(NewVisualFeedbackComponent())

	entities := []*Entity{player, npc}

	// Run several frames to let glow ramp up
	for i := 0; i < 10; i++ {
		sys.Update(entities, 0.016)
	}

	// NPC should now have glow component
	comp, ok := npc.GetComponent("npc_proximity_glow")
	if !ok {
		t.Fatal("expected npc_proximity_glow component on NPC")
	}
	glow := comp.(*NPCProximityGlowComponent)
	if !glow.Active {
		t.Error("expected glow to be active when player is in range")
	}
	if glow.GlowIntensity <= 0 {
		t.Errorf("expected positive GlowIntensity, got %f", glow.GlowIntensity)
	}

	// Check visual feedback tint was modified
	feedback := npc.GetVisualFeedback()
	if feedback == nil {
		t.Fatal("expected VisualFeedbackComponent on NPC")
	}
	// Fantasy glow is warm gold (R=255, G=215, B=80) → TintR should be ~1.0, TintB < 1.0
	if feedback.TintB >= 1.0 {
		t.Errorf("expected TintB < 1.0 for gold glow, got %f", feedback.TintB)
	}
}

func TestNPCInteractionProximityGlowSystem_Update_GlowDecays(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	// Create player far away
	player := NewEntity(1002)
	player.AddComponent(&PositionComponent{X: 500, Y: 500})
	player.AddComponent(&stubInputTag{})

	// Create NPC with merchant component
	npc := NewEntity(1003)
	npc.AddComponent(&PositionComponent{X: 50, Y: 50})
	npc.AddComponent(&MerchantComponent{})
	npc.AddComponent(NewVisualFeedbackComponent())

	// Pre-seed glow on the NPC
	glow := &NPCProximityGlowComponent{GlowIntensity: 1.0, Active: true}
	npc.AddComponent(glow)

	entities := []*Entity{player, npc}

	// Run several frames — glow should decay since player is far away
	for i := 0; i < 20; i++ {
		sys.Update(entities, 0.016)
	}

	if glow.Active {
		t.Error("expected glow to deactivate after player leaves range")
	}
	if glow.GlowIntensity > 0.01 {
		t.Errorf("expected near-zero GlowIntensity, got %f", glow.GlowIntensity)
	}
}

func TestNPCInteractionProximityGlowSystem_Update_DistanceFalloff(t *testing.T) {
	tests := []struct {
		name         string
		npcX         float64 // Player always at (0,0)
		wantHighGlow bool
	}{
		{"at_inner_radius", 30, true},
		{"at_mid_range", 60, true},
		{"beyond_glow_radius", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNPCInteractionProximityGlowSystem(world, 42)

			player := NewEntity(1004)
			player.AddComponent(&PositionComponent{X: 0, Y: 0})
			player.AddComponent(&stubInputTag{})

			npc := NewEntity(1005)
			npc.AddComponent(&PositionComponent{X: tt.npcX, Y: 0})
			npc.AddComponent(NewDialogComponent(nil))

			entities := []*Entity{player, npc}
			for i := 0; i < 15; i++ {
				sys.Update(entities, 0.016)
			}

			comp, ok := npc.GetComponent("npc_proximity_glow")
			if !ok && tt.wantHighGlow {
				t.Fatal("expected npc_proximity_glow component")
			}
			if ok {
				glow := comp.(*NPCProximityGlowComponent)
				if tt.wantHighGlow && glow.GlowIntensity <= 0 {
					t.Errorf("expected positive glow at distance %f", tt.npcX)
				}
				if !tt.wantHighGlow && glow.Active {
					t.Errorf("expected inactive glow at distance %f", tt.npcX)
				}
			}
		})
	}
}

func TestNPCInteractionProximityGlowSystem_Update_NoGlowOnNonNPC(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	player := NewEntity(1006)
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&stubInputTag{})

	// Entity without dialog or merchant — should not get glow
	plain := NewEntity(1007)
	plain.AddComponent(&PositionComponent{X: 55, Y: 50})

	entities := []*Entity{player, plain}
	for i := 0; i < 10; i++ {
		sys.Update(entities, 0.016)
	}

	_, ok := plain.GetComponent("npc_proximity_glow")
	if ok {
		t.Error("non-NPC entity should not have npc_proximity_glow component")
	}
}

func TestNPCInteractionProximityGlowSystem_Update_MerchantGlows(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	player := NewEntity(1008)
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&stubInputTag{})

	merchant := NewEntity(1009)
	merchant.AddComponent(&PositionComponent{X: 55, Y: 50})
	merchant.AddComponent(&MerchantComponent{})

	entities := []*Entity{player, merchant}
	for i := 0; i < 10; i++ {
		sys.Update(entities, 0.016)
	}

	comp, ok := merchant.GetComponent("npc_proximity_glow")
	if !ok {
		t.Fatal("expected npc_proximity_glow on merchant")
	}
	glow := comp.(*NPCProximityGlowComponent)
	if !glow.Active {
		t.Error("expected merchant glow to be active near player")
	}
}

func TestNPCInteractionProximityGlowSystem_Update_NoPlayers(t *testing.T) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	npc := NewEntity(1010)
	npc.AddComponent(&PositionComponent{X: 50, Y: 50})
	npc.AddComponent(NewDialogComponent(nil))
	glow := &NPCProximityGlowComponent{GlowIntensity: 0.5, Active: true}
	npc.AddComponent(glow)

	entities := []*Entity{npc}
	// No players — glow should decay
	for i := 0; i < 20; i++ {
		sys.Update(entities, 0.016)
	}

	if glow.Active {
		t.Error("expected glow to deactivate with no players present")
	}
}

func BenchmarkNPCInteractionProximityGlowSystem(b *testing.B) {
	world := NewWorld()
	sys := NewNPCInteractionProximityGlowSystem(world, 42)

	entities := make([]*Entity, 0, 102)
	// 2 players
	for i := 0; i < 2; i++ {
		p := NewEntity(uint64(2000 + i))
		p.AddComponent(&PositionComponent{X: float64(i * 100), Y: 50})
		p.AddComponent(&stubInputTag{})
		entities = append(entities, p)
	}
	// 100 NPCs
	for i := 0; i < 100; i++ {
		n := NewEntity(uint64(3000 + i))
		n.AddComponent(&PositionComponent{X: float64(i * 10), Y: 50})
		n.AddComponent(NewDialogComponent(nil))
		n.AddComponent(NewVisualFeedbackComponent())
		entities = append(entities, n)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
