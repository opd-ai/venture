package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/world/territory"
	"github.com/sirupsen/logrus"
)

func TestNewTerritorySystem(t *testing.T) {
	manager := territory.NewManager()
	logger := logrus.WithField("test", "territory")

	sys := NewTerritorySystem(manager, logger)

	if sys == nil {
		t.Fatal("expected system, got nil")
	}
	if sys.manager != manager {
		t.Error("manager not set correctly")
	}
	if sys.updateInterval != 1.0 {
		t.Errorf("expected update interval 1.0, got %f", sys.updateInterval)
	}
}

func TestNewTerritorySystem_NilLogger(t *testing.T) {
	manager := territory.NewManager()

	sys := NewTerritorySystem(manager, nil)

	if sys == nil {
		t.Fatal("expected system, got nil")
	}
	if sys.logger == nil {
		t.Error("logger should be initialized when nil passed")
	}
}

func TestTerritorySystem_GetTerritoryIDFromPosition(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	tests := []struct {
		name     string
		x, y     float64
		expected string
	}{
		{"origin", 0, 0, "territory_0_0"},
		{"first territory", 250, 250, "territory_0_0"},
		{"second territory X", 500, 0, "territory_1_0"},
		{"second territory Y", 0, 500, "territory_0_1"},
		{"diagonal territory", 750, 1000, "territory_1_2"},
		{"negative coordinates", -100, -100, "territory_-1_-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.getTerritoryIDFromPosition(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTerritorySystem_EnsureTerritoryExists(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	terr, err := sys.EnsureTerritoryExists(250, 250)
	if err != nil {
		t.Fatalf("failed to create territory: %v", err)
	}
	if terr == nil {
		t.Fatal("expected territory, got nil")
	}
	if terr.ID != "territory_0_0" {
		t.Errorf("expected territory_0_0, got %s", terr.ID)
	}
	if terr.Status != territory.StatusNeutral {
		t.Errorf("expected neutral status, got %v", terr.Status)
	}

	// Ensure calling again returns same territory
	terr2, err := sys.EnsureTerritoryExists(250, 250)
	if err != nil {
		t.Fatalf("failed to get existing territory: %v", err)
	}
	if terr2.ID != terr.ID {
		t.Error("expected same territory ID")
	}
}

func TestTerritorySystem_GetTerritoryAtPosition(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// Non-existent territory
	_, err := sys.GetTerritoryAtPosition(250, 250)
	if err == nil {
		t.Error("expected error for non-existent territory")
	}

	// Create territory
	sys.EnsureTerritoryExists(250, 250)

	// Should now exist
	terr, err := sys.GetTerritoryAtPosition(250, 250)
	if err != nil {
		t.Fatalf("failed to get territory: %v", err)
	}
	if terr == nil {
		t.Fatal("expected territory, got nil")
	}
}

func TestTerritorySystem_Update(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// Create a test territory
	terr, _ := sys.EnsureTerritoryExists(250, 250)

	// Create test entities with guild membership in territory
	guildA := "guild_a"
	guildB := "guild_b"

	// Assign territory to guild A
	manager.AssignOwner(terr.ID, guildA)

	entities := []*Entity{
		createEntityAtPosition(250, 250, guildA), // defender
		createEntityAtPosition(260, 260, guildA), // defender
		createEntityAtPosition(270, 270, guildB), // attacker
		createEntityAtPosition(280, 280, guildB), // attacker
		createEntityAtPosition(290, 290, guildB), // attacker
	}

	// Update should process after 1 second
	sys.Update(entities, 0.5)

	// No update yet (< 1.0s)
	updatedTerr, _ := manager.GetTerritory(terr.ID)
	if updatedTerr.Status == territory.StatusContested {
		t.Error("territory should not be contested yet (< 1s)")
	}

	// Trigger update
	sys.Update(entities, 0.6)

	// Should now be contested (3 attackers > 2 defenders)
	updatedTerr, _ = manager.GetTerritory(terr.ID)
	if updatedTerr.Status != territory.StatusContested {
		t.Errorf("expected contested status, got %v", updatedTerr.Status)
	}
	if updatedTerr.CapturingGuild != guildB {
		t.Errorf("expected guild B capturing, got %s", updatedTerr.CapturingGuild)
	}
}

func TestTerritorySystem_Update_NoGuildEntities(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// Create entities without guild component
	entities := []*Entity{
		{
			ID:         1,
			Components: make(map[string]Component),
		},
	}
	entities[0].AddComponent(&PositionComponent{X: 250, Y: 250})

	// Should not panic
	sys.Update(entities, 1.0)
}

func TestTerritorySystem_ProcessCombatDamage(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// Create territory and assign owner
	terr, _ := sys.EnsureTerritoryExists(250, 250)
	manager.AssignOwner(terr.ID, "guild_a")

	// Build a defensive structure
	structure, err := manager.BuildDefensiveStructure(terr.ID, territory.StructureTypeWall, 250, 250)
	if err != nil {
		t.Fatalf("failed to build structure: %v", err)
	}

	// Deal damage
	err = sys.ProcessCombatDamage(terr.ID, structure.ID, 100.0)
	if err != nil {
		t.Fatalf("failed to damage structure: %v", err)
	}

	// Verify damage applied
	updatedTerr, _ := manager.GetTerritory(terr.ID)
	if len(updatedTerr.Structures) == 0 {
		t.Fatal("structure disappeared")
	}
	if updatedTerr.Structures[0].HP != structure.MaxHP-100.0 {
		t.Errorf("expected HP %f, got %f", structure.MaxHP-100.0, updatedTerr.Structures[0].HP)
	}
}

func TestTerritorySystem_GetManager(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	if sys.GetManager() != manager {
		t.Error("GetManager returned wrong manager")
	}
}

func TestTerritorySystem_GetBonusesForGuild(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// Test with empty guild ID
	resourceBonus, xpBonus := sys.GetBonusesForGuild("")
	if resourceBonus != 0 || xpBonus != 0 {
		t.Errorf("expected no bonuses for empty guild, got resource=%f xp=%f", resourceBonus, xpBonus)
	}

	// Test with guild that has no territories
	resourceBonus, xpBonus = sys.GetBonusesForGuild("guild_a")
	if resourceBonus != 0 || xpBonus != 0 {
		t.Errorf("expected no bonuses for guild with no territories, got resource=%f xp=%f", resourceBonus, xpBonus)
	}

	// Create territory and assign to guild
	terr, err := sys.EnsureTerritoryExists(250, 250)
	if err != nil {
		t.Fatalf("failed to create territory: %v", err)
	}
	err = manager.AssignOwner(terr.ID, "guild_a")
	if err != nil {
		t.Fatalf("failed to assign owner: %v", err)
	}

	// Now guild_a should have bonuses
	resourceBonus, xpBonus = sys.GetBonusesForGuild("guild_a")
	if resourceBonus == 0 || xpBonus == 0 {
		t.Errorf("expected bonuses for guild with territory, got resource=%f xp=%f", resourceBonus, xpBonus)
	}

	// Guild B still has no territories
	resourceBonus, xpBonus = sys.GetBonusesForGuild("guild_b")
	if resourceBonus != 0 || xpBonus != 0 {
		t.Errorf("expected no bonuses for guild_b, got resource=%f xp=%f", resourceBonus, xpBonus)
	}
}

// TestTerritorySystem_ImplementsBonusProvider verifies TerritorySystem implements TerritoryBonusProvider.
func TestTerritorySystem_ImplementsBonusProvider(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	// This should compile if TerritorySystem implements TerritoryBonusProvider
	var provider TerritoryBonusProvider = sys
	if provider == nil {
		t.Error("TerritorySystem should implement TerritoryBonusProvider")
	}
}

// Helper function to create entity at position with guild
func createEntityAtPosition(x, y float64, guildID string) *Entity {
	entity := &Entity{
		ID:         nextEntityID(),
		Components: make(map[string]Component),
	}
	entity.AddComponent(&PositionComponent{X: x, Y: y})
	entity.AddComponent(&GuildComponent{GuildID: guildID})
	return entity
}

var testEntityIDCounter uint64 = 1000

func nextEntityID() uint64 {
	testEntityIDCounter++
	return testEntityIDCounter
}
