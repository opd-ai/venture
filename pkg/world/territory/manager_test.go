package territory

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil Manager")
	}
	if m.captureRadius != 50.0 {
		t.Errorf("expected captureRadius 50.0, got %f", m.captureRadius)
	}
	if len(m.territories) != 0 {
		t.Errorf("expected 0 territories, got %d", len(m.territories))
	}
}

func TestCreateTerritory(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}

	territory, err := m.CreateTerritory("terr-1", coords)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if territory == nil {
		t.Fatal("expected non-nil territory")
	}
	if territory.ID != "terr-1" {
		t.Errorf("expected ID 'terr-1', got '%s'", territory.ID)
	}
	if territory.Coords.ChunkX != 10 || territory.Coords.ChunkZ != 20 {
		t.Errorf("expected coords (10, 20), got (%d, %d)", territory.Coords.ChunkX, territory.Coords.ChunkZ)
	}
	if territory.Status != StatusNeutral {
		t.Errorf("expected status Neutral, got %s", territory.Status)
	}
	if territory.ResourceBonus != BaseResourceBonus {
		t.Errorf("expected ResourceBonus %f, got %f", BaseResourceBonus, territory.ResourceBonus)
	}
	if territory.XPBonus != BaseXPBonus {
		t.Errorf("expected XPBonus %f, got %f", BaseXPBonus, territory.XPBonus)
	}
}

func TestCreateTerritory_Duplicate(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}

	_, err := m.CreateTerritory("terr-1", coords)
	if err != nil {
		t.Fatalf("unexpected error on first create: %v", err)
	}

	_, err = m.CreateTerritory("terr-1", coords)
	if err == nil {
		t.Error("expected error for duplicate territory ID")
	}
}

func TestGetTerritory(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing territory", "terr-1", false},
		{"non-existent territory", "terr-2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			territory, err := m.GetTerritory(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if territory == nil {
					t.Error("expected non-nil territory")
				}
			}
		})
	}
}

func TestAssignOwner(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)

	err := m.AssignOwner("terr-1", "guild-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	territory, _ := m.GetTerritory("terr-1")
	if territory.OwnerGuildID != "guild-123" {
		t.Errorf("expected OwnerGuildID 'guild-123', got '%s'", territory.OwnerGuildID)
	}
	if territory.Status != StatusOwned {
		t.Errorf("expected status Owned, got %s", territory.Status)
	}
	if territory.CaptureProgress != 0.0 {
		t.Errorf("expected CaptureProgress 0.0, got %f", territory.CaptureProgress)
	}
}

func TestAssignOwner_NonExistent(t *testing.T) {
	m := NewManager()

	err := m.AssignOwner("terr-1", "guild-123")
	if err == nil {
		t.Error("expected error for non-existent territory")
	}
}

func TestUpdateCaptureProgress_Attackers(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	territory, _ := m.CreateTerritory("terr-1", coords)

	territory.LastUpdate = time.Now().Add(-10 * time.Second)

	err := m.UpdateCaptureProgress("terr-1", 5, 0, "guild-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	territory, _ = m.GetTerritory("terr-1")
	if territory.CaptureProgress <= 0.0 {
		t.Error("expected capture progress > 0")
	}
	if territory.CapturingGuild != "guild-456" {
		t.Errorf("expected CapturingGuild 'guild-456', got '%s'", territory.CapturingGuild)
	}
}

func TestUpdateCaptureProgress_Defenders(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	territory, _ := m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	// Set up internal state via CreateTerritory pointer
	territory.CaptureProgress = 0.5
	territory.CapturingGuild = "guild-456"
	territory.Status = StatusContested
	territory.LastUpdate = time.Now().Add(-10 * time.Second)

	err := m.UpdateCaptureProgress("terr-1", 2, 5, "guild-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, _ := m.GetTerritory("terr-1")
	if result.CaptureProgress > 0.5 {
		t.Error("expected capture progress to decay with more defenders")
	}
}

func TestUpdateCaptureProgress_Completion(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	territory, _ := m.CreateTerritory("terr-1", coords)

	territory.CaptureProgress = 0.9
	territory.CapturingGuild = "guild-456"
	territory.LastUpdate = time.Now().Add(-60 * time.Second)

	err := m.UpdateCaptureProgress("terr-1", 5, 0, "guild-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	territory, _ = m.GetTerritory("terr-1")
	if territory.CaptureProgress != 1.0 {
		t.Errorf("expected full capture (1.0), got %f", territory.CaptureProgress)
	}
	if territory.OwnerGuildID != "guild-456" {
		t.Errorf("expected OwnerGuildID 'guild-456', got '%s'", territory.OwnerGuildID)
	}
	if territory.Status != StatusOwned {
		t.Errorf("expected status Owned, got %s", territory.Status)
	}
	if territory.CapturingGuild != "" {
		t.Errorf("expected empty CapturingGuild after completion, got '%s'", territory.CapturingGuild)
	}
}

func TestUpdateCaptureProgress_NoAttackingGuild(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)

	err := m.UpdateCaptureProgress("terr-1", 5, 0, "")
	if err == nil {
		t.Error("expected error when attacking guild not specified")
	}
}

func TestBuildDefensiveStructure_Wall(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	structure, err := m.BuildDefensiveStructure("terr-1", StructureTypeWall, 100.0, 100.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if structure == nil {
		t.Fatal("expected non-nil structure")
	}
	if structure.Type != StructureTypeWall {
		t.Errorf("expected type Wall, got %s", structure.Type)
	}
	if structure.MaxHP != WallBaseHP {
		t.Errorf("expected MaxHP %f, got %f", WallBaseHP, structure.MaxHP)
	}
	if structure.HP != structure.MaxHP {
		t.Error("expected HP to equal MaxHP for new structure")
	}
	if structure.X != 100.0 || structure.Y != 100.0 {
		t.Errorf("expected position (100.0, 100.0), got (%f, %f)", structure.X, structure.Y)
	}
}

func TestBuildDefensiveStructure_Tower(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	structure, err := m.BuildDefensiveStructure("terr-1", StructureTypeTower, 200.0, 200.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if structure.Type != StructureTypeTower {
		t.Errorf("expected type Tower, got %s", structure.Type)
	}
	if structure.Damage != TowerDamage {
		t.Errorf("expected Damage %f, got %f", TowerDamage, structure.Damage)
	}
	if structure.MaxHP != TowerBaseHP {
		t.Errorf("expected MaxHP %f, got %f", TowerBaseHP, structure.MaxHP)
	}
}

func TestBuildDefensiveStructure_Guard(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	structure, err := m.BuildDefensiveStructure("terr-1", StructureTypeGuard, 300.0, 300.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if structure.Type != StructureTypeGuard {
		t.Errorf("expected type Guard, got %s", structure.Type)
	}
	if structure.Level != GuardLevel {
		t.Errorf("expected Level %d, got %d", GuardLevel, structure.Level)
	}
	if structure.MaxHP != GuardBaseHP {
		t.Errorf("expected MaxHP %f, got %f", GuardBaseHP, structure.MaxHP)
	}
}

func TestBuildDefensiveStructure_Unowned(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)

	_, err := m.BuildDefensiveStructure("terr-1", StructureTypeWall, 100.0, 100.0)
	if err == nil {
		t.Error("expected error for building in unowned territory")
	}
}

func TestDamageStructure(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	structure, _ := m.BuildDefensiveStructure("terr-1", StructureTypeWall, 100.0, 100.0)
	initialHP := structure.HP

	err := m.DamageStructure("terr-1", structure.ID, 200.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	territory, _ := m.GetTerritory("terr-1")
	found := false
	for _, s := range territory.Structures {
		if s.ID == structure.ID {
			found = true
			if s.HP >= initialHP {
				t.Error("expected HP to decrease after damage")
			}
		}
	}
	if !found {
		t.Error("structure still exists after taking non-lethal damage")
	}
}

func TestDamageStructure_Destruction(t *testing.T) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	m.CreateTerritory("terr-1", coords)
	m.AssignOwner("terr-1", "guild-123")

	structure, _ := m.BuildDefensiveStructure("terr-1", StructureTypeWall, 100.0, 100.0)

	err := m.DamageStructure("terr-1", structure.ID, 10000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	territory, _ := m.GetTerritory("terr-1")
	for _, s := range territory.Structures {
		if s.ID == structure.ID {
			t.Error("structure should be removed after destruction")
		}
	}
}

func TestDeclareWar(t *testing.T) {
	m := NewManager()

	war, err := m.DeclareWar("guild-123", "guild-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if war == nil {
		t.Fatal("expected non-nil war declaration")
	}
	if war.AttackerGuild != "guild-123" {
		t.Errorf("expected AttackerGuild 'guild-123', got '%s'", war.AttackerGuild)
	}
	if war.DefenderGuild != "guild-456" {
		t.Errorf("expected DefenderGuild 'guild-456', got '%s'", war.DefenderGuild)
	}
	if !war.Active {
		t.Error("expected war to be active")
	}
	if war.Cost != WarDeclarationCost {
		t.Errorf("expected Cost %d, got %d", WarDeclarationCost, war.Cost)
	}
}

func TestDeclareWar_SelfWar(t *testing.T) {
	m := NewManager()

	_, err := m.DeclareWar("guild-123", "guild-123")
	if err == nil {
		t.Error("expected error for guild declaring war on itself")
	}
}

func TestDeclareWar_Duplicate(t *testing.T) {
	m := NewManager()

	_, err := m.DeclareWar("guild-123", "guild-456")
	if err != nil {
		t.Fatalf("unexpected error on first war: %v", err)
	}

	_, err = m.DeclareWar("guild-123", "guild-456")
	if err == nil {
		t.Error("expected error for duplicate war declaration")
	}
}

func TestEndWar(t *testing.T) {
	m := NewManager()

	war, _ := m.DeclareWar("guild-123", "guild-456")

	err := m.EndWar(war.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if war.Active {
		t.Error("expected war to be inactive after ending")
	}
}

func TestIsAtWar(t *testing.T) {
	m := NewManager()

	if m.IsAtWar("guild-123", "guild-456") {
		t.Error("expected no war before declaration")
	}

	m.DeclareWar("guild-123", "guild-456")

	if !m.IsAtWar("guild-123", "guild-456") {
		t.Error("expected war after declaration")
	}
	if !m.IsAtWar("guild-456", "guild-123") {
		t.Error("expected war to be symmetric")
	}
}

func TestGetGuildTerritories(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})
	m.CreateTerritory("terr-3", TerritoryCoords{ChunkX: 30, ChunkZ: 30})

	m.AssignOwner("terr-1", "guild-123")
	m.AssignOwner("terr-2", "guild-123")
	m.AssignOwner("terr-3", "guild-456")

	territories := m.GetGuildTerritories("guild-123")
	if len(territories) != 2 {
		t.Errorf("expected 2 territories for guild-123, got %d", len(territories))
	}

	territories = m.GetGuildTerritories("guild-456")
	if len(territories) != 1 {
		t.Errorf("expected 1 territory for guild-456, got %d", len(territories))
	}

	territories = m.GetGuildTerritories("guild-789")
	if len(territories) != 0 {
		t.Errorf("expected 0 territories for guild-789, got %d", len(territories))
	}
}

func TestGetResourceBonus(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})

	m.AssignOwner("terr-1", "guild-123")
	m.AssignOwner("terr-2", "guild-123")

	bonus := m.GetResourceBonus("guild-123")
	expected := 2 * BaseResourceBonus
	if bonus != expected {
		t.Errorf("expected resource bonus %f, got %f", expected, bonus)
	}

	bonus = m.GetResourceBonus("guild-456")
	if bonus != 0.0 {
		t.Errorf("expected resource bonus 0.0 for guild-456, got %f", bonus)
	}
}

func TestGetXPBonus(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})
	m.CreateTerritory("terr-3", TerritoryCoords{ChunkX: 30, ChunkZ: 30})

	m.AssignOwner("terr-1", "guild-123")
	m.AssignOwner("terr-2", "guild-123")
	m.AssignOwner("terr-3", "guild-123")

	bonus := m.GetXPBonus("guild-123")
	expected := 3 * BaseXPBonus
	tolerance := 0.0001
	if bonus < expected-tolerance || bonus > expected+tolerance {
		t.Errorf("expected XP bonus %f, got %f", expected, bonus)
	}
}

// TestGetBonusesForGuild verifies that GetBonusesForGuild returns both resource and XP bonuses.
// This method implements the engine.TerritoryBonusProvider interface for HUD display.
func TestGetBonusesForGuild(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})

	m.AssignOwner("terr-1", "guild-hud-test")
	m.AssignOwner("terr-2", "guild-hud-test")

	resourceBonus, xpBonus := m.GetBonusesForGuild("guild-hud-test")
	expectedResource := 2 * BaseResourceBonus
	expectedXP := 2 * BaseXPBonus
	tolerance := 0.0001

	if resourceBonus < expectedResource-tolerance || resourceBonus > expectedResource+tolerance {
		t.Errorf("expected resource bonus %f, got %f", expectedResource, resourceBonus)
	}
	if xpBonus < expectedXP-tolerance || xpBonus > expectedXP+tolerance {
		t.Errorf("expected XP bonus %f, got %f", expectedXP, xpBonus)
	}

	// Test guild with no territories
	resourceBonus, xpBonus = m.GetBonusesForGuild("unknown-guild")
	if resourceBonus != 0.0 || xpBonus != 0.0 {
		t.Errorf("expected 0 bonuses for unknown guild, got resource=%f xp=%f", resourceBonus, xpBonus)
	}
}

func TestGetContestedTerritories(t *testing.T) {
	m := NewManager()
	territory1, _ := m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})

	territory1.Status = StatusContested

	contested := m.GetContestedTerritories()
	if len(contested) != 1 {
		t.Errorf("expected 1 contested territory, got %d", len(contested))
	}
}

func TestGetAllTerritories(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 10})
	m.CreateTerritory("terr-2", TerritoryCoords{ChunkX: 20, ChunkZ: 20})
	m.CreateTerritory("terr-3", TerritoryCoords{ChunkX: 30, ChunkZ: 30})

	territories := m.GetAllTerritories()
	if len(territories) != 3 {
		t.Errorf("expected 3 territories, got %d", len(territories))
	}
}

func TestGetActiveWars(t *testing.T) {
	m := NewManager()
	war1, _ := m.DeclareWar("guild-123", "guild-456")
	m.DeclareWar("guild-789", "guild-012")

	m.EndWar(war1.ID)

	active := m.GetActiveWars()
	if len(active) != 1 {
		t.Errorf("expected 1 active war, got %d", len(active))
	}
}

func TestGetGuildWars(t *testing.T) {
	m := NewManager()
	m.DeclareWar("guild-123", "guild-456")
	m.DeclareWar("guild-123", "guild-789")
	m.DeclareWar("guild-456", "guild-012")

	wars := m.GetGuildWars("guild-123")
	if len(wars) != 2 {
		t.Errorf("expected 2 wars for guild-123, got %d", len(wars))
	}

	wars = m.GetGuildWars("guild-456")
	if len(wars) != 2 {
		t.Errorf("expected 2 wars for guild-456, got %d", len(wars))
	}

	wars = m.GetGuildWars("guild-999")
	if len(wars) != 0 {
		t.Errorf("expected 0 wars for guild-999, got %d", len(wars))
	}
}

func TestDeclareWar_Duration(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	m := NewManagerWithTimeProvider(tp)

	war, err := m.DeclareWar("guild-123", "guild-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEnd := fixedTime.Add(time.Duration(WarDurationDays) * 24 * time.Hour)
	if !war.EndsAt.Equal(expectedEnd) {
		t.Errorf("expected EndsAt %v, got %v", expectedEnd, war.EndsAt)
	}

	// Verify duration is exactly 7 days
	duration := war.EndsAt.Sub(war.DeclaredAt)
	if duration != 7*24*time.Hour {
		t.Errorf("expected war duration 7 days, got %v", duration)
	}
}

func TestGetTerritory_DefensiveCopy(t *testing.T) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 20})
	m.AssignOwner("terr-1", "guild-123")
	m.BuildDefensiveStructure("terr-1", StructureTypeWall, 100.0, 100.0)

	// Get a copy and mutate it
	copy1, _ := m.GetTerritory("terr-1")
	copy1.OwnerGuildID = "mutated"
	copy1.Status = StatusContested
	copy1.Structures[0].HP = 0

	// Get another copy and verify internal state was not affected
	copy2, _ := m.GetTerritory("terr-1")
	if copy2.OwnerGuildID != "guild-123" {
		t.Errorf("internal state mutated: OwnerGuildID = %s, want guild-123", copy2.OwnerGuildID)
	}
	if copy2.Status != StatusOwned {
		t.Errorf("internal state mutated: Status = %s, want Owned", copy2.Status)
	}
	if copy2.Structures[0].HP == 0 {
		t.Error("internal state mutated: structure HP should not be 0")
	}
}

func TestGetActiveWars_DefensiveCopy(t *testing.T) {
	m := NewManager()
	m.DeclareWar("guild-123", "guild-456")

	// Get a copy and mutate it
	wars := m.GetActiveWars()
	wars[0].Active = false

	// Verify internal state was not affected
	wars2 := m.GetActiveWars()
	if len(wars2) != 1 {
		t.Errorf("expected 1 active war after mutating copy, got %d", len(wars2))
	}
}

// Benchmarks

func BenchmarkCreateTerritory(b *testing.B) {
	m := NewManager()
	coords := TerritoryCoords{ChunkX: 10, ChunkZ: 20}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.CreateTerritory("terr-"+string(rune(i)), coords)
	}
}

func BenchmarkUpdateCaptureProgress(b *testing.B) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 20})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.UpdateCaptureProgress("terr-1", 5, 2, "guild-456")
	}
}

func BenchmarkBuildDefensiveStructure(b *testing.B) {
	m := NewManager()
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 20})
	m.AssignOwner("terr-1", "guild-123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.BuildDefensiveStructure("terr-1", StructureTypeWall, float64(i), float64(i))
	}
}

func BenchmarkGetResourceBonus(b *testing.B) {
	m := NewManager()
	for i := 0; i < 100; i++ {
		m.CreateTerritory("terr-"+string(rune(i)), TerritoryCoords{ChunkX: i, ChunkZ: i})
		m.AssignOwner("terr-"+string(rune(i)), "guild-123")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetResourceBonus("guild-123")
	}
}

func TestManager_SetConfig(t *testing.T) {
	m := NewManager()

	// Verify default config is set
	if m.GetConfig() == nil {
		t.Fatal("expected default config")
	}
	if m.GetConfig().BaseCaptureTime != BaseCaptureTime {
		t.Errorf("expected default BaseCaptureTime %d, got %d", BaseCaptureTime, m.GetConfig().BaseCaptureTime)
	}

	// Set custom config
	customConfig := &TerritoryConfig{
		BaseCaptureTime:   120,
		DefenderTimeBonus: 60,
		BaseResourceBonus: 0.20,
		BaseXPBonus:       0.10,
		WallBaseHP:        2000.0,
		TowerBaseHP:       1000.0,
		GuardBaseHP:       1000.0,
		TowerDamage:       200.0,
		GuardLevel:        50,
		GuildHallMaxHP:    20000.0,
	}
	m.SetConfig(customConfig)

	if m.GetConfig().BaseCaptureTime != 120 {
		t.Errorf("expected custom BaseCaptureTime 120, got %d", m.GetConfig().BaseCaptureTime)
	}
	if m.GetConfig().BaseResourceBonus != 0.20 {
		t.Errorf("expected custom BaseResourceBonus 0.20, got %f", m.GetConfig().BaseResourceBonus)
	}

	// Reset to nil restores defaults
	m.SetConfig(nil)
	if m.GetConfig().BaseCaptureTime != BaseCaptureTime {
		t.Errorf("expected default BaseCaptureTime after nil set, got %d", m.GetConfig().BaseCaptureTime)
	}
}

func TestManager_ConfigAffectsBonuses(t *testing.T) {
	m := NewManager()

	// Create territory and assign to guild
	m.CreateTerritory("terr-1", TerritoryCoords{ChunkX: 10, ChunkZ: 20})
	m.AssignOwner("terr-1", "guild-123")

	// Check default bonuses
	defaultResourceBonus := m.GetResourceBonus("guild-123")
	if defaultResourceBonus != BaseResourceBonus {
		t.Errorf("expected default resource bonus %f, got %f", BaseResourceBonus, defaultResourceBonus)
	}

	// Set custom config with higher bonuses
	customConfig := DefaultTerritoryConfig()
	customConfig.BaseResourceBonus = 0.25
	customConfig.BaseXPBonus = 0.15
	m.SetConfig(customConfig)

	// Check custom bonuses
	customResourceBonus := m.GetResourceBonus("guild-123")
	if customResourceBonus != 0.25 {
		t.Errorf("expected custom resource bonus 0.25, got %f", customResourceBonus)
	}
	customXPBonus := m.GetXPBonus("guild-123")
	if customXPBonus != 0.15 {
		t.Errorf("expected custom XP bonus 0.15, got %f", customXPBonus)
	}
}
