package engine

import (
	"testing"
)

func TestNewFactionAwareAISystem(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	if system == nil {
		t.Fatal("NewFactionAwareAISystem returned nil")
	}
	if system.world != world {
		t.Error("system.world not set correctly")
	}
	if system.updateInterval != 1.0 {
		t.Errorf("expected updateInterval 1.0, got %f", system.updateInterval)
	}
}

func TestFactionAwareAISystem_Update_NoPlayer(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	npc := NewEntity(1)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{npc}, 0.1) // Should not panic
}

func TestFactionAwareAISystem_Update_HostileFaction(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "bandits", Reputation: -60, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
	npc.AddComponent(&TeamComponent{TeamID: 0})
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	team := npc.GetTeam()
	if team == nil || team.TeamID != 2 {
		t.Errorf("expected NPC team ID 2 (enemy), got %v", team)
	}
}

func TestFactionAwareAISystem_Update_NeutralFaction(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "merchants", Reputation: 30, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "merchants", IsPlayerFaction: false})
	npc.AddComponent(&TeamComponent{TeamID: 0})
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	if npc.GetTeam().TeamID != 0 {
		t.Error("expected NPC team ID 0 (neutral)")
	}
}

func TestFactionAwareAISystem_Update_ReputationImproved(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "guards", Reputation: -40, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "guards", IsPlayerFaction: false})
	npc.AddComponent(&TeamComponent{TeamID: 2}) // Was enemy
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	if npc.GetTeam().TeamID != 0 {
		t.Error("expected NPC to become neutral after rep improved")
	}
}

func TestFactionAwareAISystem_NPCWithoutAI(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "bandits", Reputation: -60, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2) // No AI component
	npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
	npc.AddComponent(&TeamComponent{TeamID: 0})
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	if npc.GetTeam().TeamID != 0 {
		t.Error("NPC without AI should be unchanged")
	}
}

func TestFactionAwareAISystem_AddTeamToNPCWithoutOne(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "bandits", Reputation: -60, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
	// No TeamComponent
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	team := npc.GetTeam()
	if team == nil || team.TeamID != 2 {
		t.Error("expected system to add TeamComponent with ID 2")
	}
}

func TestFactionAwareAISystem_ValueTypeFactionComponent(t *testing.T) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "bandits", Reputation: -60, IsPlayerFaction: true})
	world.AddEntity(player)

	npc := NewEntity(2)
	npc.AddComponent(&AIComponent{})
	npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
	npc.AddComponent(&TeamComponent{TeamID: 0})
	world.AddEntity(npc)

	system.timeSinceCheck = 1.0
	system.Update([]*Entity{player, npc}, 0.1)

	if npc.GetTeam().TeamID != 2 {
		t.Error("value-type faction should work")
	}
}

func TestCalculateTeamFromReputation(t *testing.T) {
	system := NewFactionAwareAISystem(NewWorld())
	tests := []struct {
		name        string
		reputation  int
		currentTeam int
		wantTeam    int
	}{
		{"hostile makes enemy", -60, 0, 2},
		{"exactly hostile", -50, 0, 2},
		{"suspicious stays", -40, 0, 0},
		{"neutral stays", 30, 0, 0},
		{"improved from enemy", -40, 2, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.calculateTeamFromReputation(tt.reputation, tt.currentTeam)
			if got != tt.wantTeam {
				t.Errorf("got %d, want %d", got, tt.wantTeam)
			}
		})
	}
}

func TestFactionAwareAISystem_SetUpdateInterval(t *testing.T) {
	system := NewFactionAwareAISystem(NewWorld())
	system.SetUpdateInterval(2.0)
	if system.updateInterval != 2.0 {
		t.Error("expected 2.0")
	}
	system.SetUpdateInterval(-1.0) // Invalid - should be ignored
	if system.updateInterval != 2.0 {
		t.Error("negative should be ignored")
	}
}

func BenchmarkFactionAwareAISystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewFactionAwareAISystem(world)

	player := NewEntity(1)
	player.AddComponent(NewStubInput())
	player.AddComponent(&FactionComponent{FactionID: "bandits", Reputation: -60, IsPlayerFaction: true})
	world.AddEntity(player)

	entities := []*Entity{player}
	for i := uint64(2); i < 102; i++ {
		npc := NewEntity(i)
		npc.AddComponent(&AIComponent{})
		npc.AddComponent(&FactionComponent{FactionID: "bandits", IsPlayerFaction: false})
		npc.AddComponent(&TeamComponent{TeamID: 0})
		world.AddEntity(npc)
		entities = append(entities, npc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.timeSinceCheck = 1.0
		system.Update(entities, 0.016)
	}
}
