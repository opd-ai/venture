//go:build test
// +build test

// Package engine provides tests for network components.
package engine

import (
	"testing"
)

func TestNetworkComponent_Creation(t *testing.T) {
	netComp := &NetworkComponent{
		PlayerID:      42,
		Synced:        true,
		LastUpdateSeq: 100,
	}

	if netComp.Type() != "network" {
		t.Errorf("Expected type 'network', got '%s'", netComp.Type())
	}

	if netComp.PlayerID != 42 {
		t.Errorf("Expected PlayerID 42, got %d", netComp.PlayerID)
	}

	if !netComp.Synced {
		t.Error("Expected Synced to be true")
	}

	if netComp.LastUpdateSeq != 100 {
		t.Errorf("Expected LastUpdateSeq 100, got %d", netComp.LastUpdateSeq)
	}
}

func TestNetworkComponent_Type(t *testing.T) {
	netComp := &NetworkComponent{}
	if netComp.Type() != "network" {
		t.Errorf("Expected type 'network', got '%s'", netComp.Type())
	}
}

func TestNetworkComponent_EntityIntegration(t *testing.T) {
	world := NewWorld()
	entity := world.CreateEntity()

	// Add network component
	netComp := &NetworkComponent{
		PlayerID: 1,
		Synced:   true,
	}
	entity.AddComponent(netComp)

	// Verify component retrieval
	retrieved, ok := entity.GetComponent("network")
	if !ok {
		t.Fatal("Network component not found on entity")
	}

	nc := retrieved.(*NetworkComponent)
	if nc.PlayerID != 1 {
		t.Errorf("Expected PlayerID 1, got %d", nc.PlayerID)
	}

	if !nc.Synced {
		t.Error("Expected Synced to be true")
	}
}

func TestNetworkComponent_DefaultValues(t *testing.T) {
	netComp := &NetworkComponent{}

	// Verify zero values
	if netComp.PlayerID != 0 {
		t.Errorf("Expected default PlayerID 0, got %d", netComp.PlayerID)
	}

	if netComp.Synced {
		t.Error("Expected default Synced to be false")
	}

	if netComp.LastUpdateSeq != 0 {
		t.Errorf("Expected default LastUpdateSeq 0, got %d", netComp.LastUpdateSeq)
	}
}

func TestNetworkComponent_MultipleEntities(t *testing.T) {
	world := NewWorld()

	// Create multiple player entities
	players := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		players[i] = world.CreateEntity()
		players[i].AddComponent(&NetworkComponent{
			PlayerID: uint64(i + 1),
			Synced:   true,
		})
	}

	// Verify each entity has correct PlayerID
	for i, player := range players {
		comp, ok := player.GetComponent("network")
		if !ok {
			t.Fatalf("Player %d missing network component", i)
		}

		nc := comp.(*NetworkComponent)
		expectedID := uint64(i + 1)
		if nc.PlayerID != expectedID {
			t.Errorf("Player %d has PlayerID %d, expected %d", i, nc.PlayerID, expectedID)
		}
	}
}

func TestNetworkComponent_SyncedFlag(t *testing.T) {
	tests := []struct {
		name   string
		synced bool
	}{
		{"Player entity synced", true},
		{"NPC entity not synced", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netComp := &NetworkComponent{
				Synced: tt.synced,
			}

			if netComp.Synced != tt.synced {
				t.Errorf("Expected Synced=%v, got %v", tt.synced, netComp.Synced)
			}
		})
	}
}

func TestNetworkComponent_SequenceTracking(t *testing.T) {
	netComp := &NetworkComponent{
		LastUpdateSeq: 0,
	}

	// Simulate sequence updates
	sequences := []uint32{1, 5, 10, 100, 1000}
	for _, seq := range sequences {
		netComp.LastUpdateSeq = seq
		if netComp.LastUpdateSeq != seq {
			t.Errorf("Expected LastUpdateSeq %d, got %d", seq, netComp.LastUpdateSeq)
		}
	}
}
