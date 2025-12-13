package territory_siege

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
	"github.com/opd-ai/venture/pkg/world"
)

// TestNewSystem tests system creation.
func TestNewSystem(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	if system == nil {
		t.Fatal("NewSystem() returned nil")
	}

	if system.world != w {
		t.Error("World not set correctly")
	}

	if system.manager == nil {
		t.Error("Manager not initialized")
	}

	if system.logger == nil {
		t.Error("Logger not initialized")
	}
}

// TestSystem_GetManager tests manager accessor.
func TestSystem_GetManager(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)
	manager := system.GetManager()

	if manager == nil {
		t.Fatal("GetManager() returned nil")
	}

	if manager != system.manager {
		t.Error("GetManager() returned different manager")
	}
}

// TestSystem_Update_NoEntities tests update with no entities.
func TestSystem_Update_NoEntities(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	// Should not panic with empty entity list
	entities := make([]*engine.Entity, 0)
	system.Update(entities, 0.016)
}

// TestSystem_Update_NonSiegeEntities tests update with entities without siege components.
func TestSystem_Update_NonSiegeEntities(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	// Create entities without siege_participant components
	entities := []*engine.Entity{
		engine.NewEntity(uint64(time.Now().UnixNano())),
		engine.NewEntity(uint64(time.Now().UnixNano())),
	}

	// Should process without errors
	system.Update(entities, 0.016)
}

// TestSystem_Update_WithSiegeParticipants tests update with siege participants.
func TestSystem_Update_WithSiegeParticipants(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	// Create test guilds
	attackerID, _ := gm.CreateGuild("fantasy", "player1")
	defenderID, _ := gm.CreateGuild("fantasy", "player2")

	attackerGuild, _ := gm.GetGuild(attackerID)
	defenderGuild, _ := gm.GetGuild(defenderID)

	// Create test zone
	zone := tm.CreateBorderZone("zone1", "server1", "server2", 3)
	zone.OwnerFaction = defenderGuild.ID

	system := NewSystem(w, tm, ps, gm)

	// Declare a siege
	siege, err := system.manager.DeclareSiege(attackerGuild.ID, defenderGuild.ID, "zone1")
	if err != nil {
		t.Fatalf("Failed to declare siege: %v", err)
	}

	// Create entity with siege participant component
	entity := engine.NewEntity(uint64(time.Now().UnixNano()))
	comp := &SiegeParticipantComponent{
		SiegeID:      siege.SiegeID,
		IsAttacker:   true,
		IsActive:     true,
		LastSeenTime: time.Now().Unix(),
	}
	entity.AddComponent(comp)

	entities := []*engine.Entity{entity}

	// Update should process participant
	system.Update(entities, 0.016)

	// Component should still be active (siege just started)
	updatedComp, ok := entity.GetComponent("siege_participant")
	if !ok {
		t.Fatal("Component not found after update")
	}

	siegeComp, ok := updatedComp.(*SiegeParticipantComponent)
	if !ok {
		t.Fatal("Component type assertion failed")
	}

	if !siegeComp.IsActive {
		t.Error("Participant should still be active in preparation phase")
	}
}

// TestSystem_Update_InvalidComponent tests handling of invalid component type.
func TestSystem_Update_InvalidComponent(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	// Create entity with wrong component type
	entity := engine.NewEntity(uint64(time.Now().UnixNano()))
	entity.AddComponent(&mockComponent{componentType: "siege_participant"})

	entities := []*engine.Entity{entity}

	// Should handle gracefully without panic
	system.Update(entities, 0.016)
}

// TestSystem_Update_InactiveSiege tests participant deactivation when siege ends.
func TestSystem_Update_InactiveSiege(t *testing.T) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	// Create entity with participant for non-existent siege
	entity := engine.NewEntity(uint64(time.Now().UnixNano()))
	comp := &SiegeParticipantComponent{
		SiegeID:      "nonexistent_siege",
		IsAttacker:   true,
		IsActive:     true,
		LastSeenTime: time.Now().Unix(),
	}
	entity.AddComponent(comp)

	entities := []*engine.Entity{entity}

	// Update should deactivate participant
	system.Update(entities, 0.016)

	updatedComp, ok := entity.GetComponent("siege_participant")
	if !ok {
		t.Fatal("Component not found after update")
	}

	siegeComp, ok := updatedComp.(*SiegeParticipantComponent)
	if !ok {
		t.Fatal("Component type assertion failed")
	}

	if siegeComp.IsActive {
		t.Error("Participant should be inactive for non-existent siege")
	}
}

// TestSiegeParticipantComponent_Type tests component type identification.
func TestSiegeParticipantComponent_Type(t *testing.T) {
	comp := &SiegeParticipantComponent{
		SiegeID:      "siege_test_123",
		IsAttacker:   true,
		IsActive:     true,
		LastSeenTime: time.Now().Unix(),
	}

	if got := comp.Type(); got != "siege_participant" {
		t.Errorf("Type() = %s, want siege_participant", got)
	}
}

// mockComponent is a test component with wrong type for testing error handling.
type mockComponent struct {
	componentType string
}

func (c *mockComponent) Type() string {
	return c.componentType
}

// BenchmarkSystem_Update benchmarks system update performance.
func BenchmarkSystem_Update(b *testing.B) {
	w := engine.NewWorld()
	tm := world.NewTerritoryManager()
	ps := engine.NewPoliticsSystem(w)
	gm := guild.NewManager()

	system := NewSystem(w, tm, ps, gm)

	// Create test entities
	entities := make([]*engine.Entity, 100)
	for i := range entities {
		entities[i] = engine.NewEntity(uint64(time.Now().UnixNano()))
		if i%10 == 0 {
			// Every 10th entity is a siege participant
			comp := &SiegeParticipantComponent{
				SiegeID:      "siege_test",
				IsAttacker:   i%2 == 0,
				IsActive:     true,
				LastSeenTime: time.Now().Unix(),
			}
			entities[i].AddComponent(comp)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
