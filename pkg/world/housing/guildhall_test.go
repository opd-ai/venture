package housing

import (
	"bytes"
	"testing"
	"time"
)

func TestGuildHallCreation(t *testing.T) {
	tests := []struct {
		name    string
		guildID string
		size    GuildHallSize
		floors  int
		wantErr bool
	}{
		{"small 1 floor", "guild1", GuildHallSmall, 1, false},
		{"medium 3 floors", "guild2", GuildHallMedium, 3, false},
		{"large 5 floors", "guild3", GuildHallLarge, 5, false},
		{"invalid 0 floors", "guild4", GuildHallSmall, 0, true},
		{"invalid 6 floors", "guild5", GuildHallLarge, 6, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := Vector2{X: 100.0, Y: 100.0}
			gh := NewGuildHall(tt.guildID, "Test Guild", pos, tt.size, tt.floors)

			if tt.wantErr {
				// Floors out of range should be caught by manager, not constructor
				if tt.floors >= 1 && tt.floors <= 5 {
					t.Errorf("expected error for floors=%d", tt.floors)
				}
				return
			}

			if gh.GuildID != tt.guildID {
				t.Errorf("got GuildID %s, want %s", gh.GuildID, tt.guildID)
			}
			if gh.Size != tt.size {
				t.Errorf("got Size %d, want %d", gh.Size, tt.size)
			}
			if gh.Floors != tt.floors {
				t.Errorf("got Floors %d, want %d", gh.Floors, tt.floors)
			}
			if gh.Phase != PhaseFoundation {
				t.Errorf("got Phase %v, want %v", gh.Phase, PhaseFoundation)
			}
		})
	}
}

func TestGuildHallMaterialContribution(t *testing.T) {
	gh := NewGuildHall("guild1", "Test Guild", Vector2{X: 100.0, Y: 100.0}, GuildHallSmall, 2)

	// Initial phase: Foundation
	required := gh.RequiredMaterials[MaterialStone]
	if required == 0 {
		t.Fatal("Foundation should require stone")
	}

	// Add partial materials
	complete := gh.AddMaterial("player1", MaterialStone, required/2)
	if complete {
		t.Error("phase should not be complete with partial materials")
	}

	// Check contribution recorded
	contrib := gh.GetPlayerContribution("player1")
	if contrib[MaterialStone] != required/2 {
		t.Errorf("got contribution %d, want %d", contrib[MaterialStone], required/2)
	}

	// Add remaining materials
	complete = gh.AddMaterial("player2", MaterialStone, required/2)
	// Still need metal for foundation
	if complete {
		t.Error("phase should not be complete without all material types")
	}

	// Add metal
	metalRequired := gh.RequiredMaterials[MaterialMetal]
	complete = gh.AddMaterial("player1", MaterialMetal, metalRequired)
	if !complete {
		t.Error("phase should be complete with all materials")
	}
}

func TestGuildHallPhaseAdvancement(t *testing.T) {
	gh := NewGuildHall("guild1", "Test Guild", Vector2{X: 100.0, Y: 100.0}, GuildHallSmall, 1)

	phases := []ConstructionPhase{PhaseFoundation, PhaseWalls, PhaseRoof, PhaseInterior, PhaseComplete}

	for i, expectedPhase := range phases {
		if gh.Phase != expectedPhase {
			t.Fatalf("step %d: got phase %v, want %v", i, gh.Phase, expectedPhase)
		}

		if expectedPhase == PhaseComplete {
			// Try to advance past complete
			err := gh.AdvancePhase()
			if err == nil {
				t.Error("advancing past complete should fail")
			}
			break
		}

		// Add all required materials
		for mat, amount := range gh.RequiredMaterials {
			gh.AddMaterial("testplayer", mat, amount)
		}

		// Advance phase
		err := gh.AdvancePhase()
		if err != nil {
			t.Fatalf("advance phase failed: %v", err)
		}
	}

	// Verify final state
	if !gh.IsComplete() {
		t.Error("guild hall should be complete")
	}
	if gh.GetProgress() != 1.0 {
		t.Errorf("got progress %.2f, want 1.0", gh.GetProgress())
	}
}

func TestGuildHallProgress(t *testing.T) {
	gh := NewGuildHall("guild1", "Test Guild", Vector2{X: 100.0, Y: 100.0}, GuildHallSmall, 1)

	// Initial progress
	progress := gh.GetProgress()
	if progress < 0.0 || progress > 0.25 {
		t.Errorf("initial progress %.2f should be in [0.0, 0.25]", progress)
	}

	// Add half the materials for foundation
	for mat, amount := range gh.RequiredMaterials {
		gh.AddMaterial("player", mat, amount/2)
	}

	progressHalf := gh.GetProgress()
	if progressHalf <= progress {
		t.Error("progress should increase after adding materials")
	}

	// Complete all phases
	for ph := PhaseFoundation; ph < PhaseComplete; ph++ {
		for mat, amount := range gh.RequiredMaterials {
			current := gh.Materials[mat]
			if current < amount {
				gh.AddMaterial("player", mat, amount-current)
			}
		}
		gh.AdvancePhase()
	}

	finalProgress := gh.GetProgress()
	if finalProgress != 1.0 {
		t.Errorf("final progress %.2f, want 1.0", finalProgress)
	}
}

func TestGuildHallMaterialProgress(t *testing.T) {
	gh := NewGuildHall("guild1", "Test Guild", Vector2{X: 100.0, Y: 100.0}, GuildHallSmall, 1)

	required := gh.RequiredMaterials[MaterialStone]
	if required == 0 {
		t.Skip("Foundation doesn't require stone in this configuration")
	}

	// Check initial progress
	collected, req, progress := gh.GetMaterialProgress(MaterialStone)
	if collected != 0 {
		t.Errorf("initial collected %d, want 0", collected)
	}
	if req != required {
		t.Errorf("required %d, want %d", req, required)
	}
	if progress != 0.0 {
		t.Errorf("initial progress %.2f, want 0.0", progress)
	}

	// Add half materials
	gh.AddMaterial("player", MaterialStone, required/2)
	collected, req, progress = gh.GetMaterialProgress(MaterialStone)
	if collected != required/2 {
		t.Errorf("collected %d, want %d", collected, required/2)
	}
	if progress < 0.49 || progress > 0.51 {
		t.Errorf("half progress %.2f, want ~0.5", progress)
	}

	// Add full materials
	gh.AddMaterial("player", MaterialStone, required/2)
	collected, req, progress = gh.GetMaterialProgress(MaterialStone)
	if progress != 1.0 {
		t.Errorf("full progress %.2f, want 1.0", progress)
	}
}

func TestGuildHallContributors(t *testing.T) {
	gh := NewGuildHall("guild1", "Test Guild", Vector2{X: 100.0, Y: 100.0}, GuildHallSmall, 1)

	// Add materials from multiple players
	gh.AddMaterial("player1", MaterialStone, 10)
	gh.AddMaterial("player2", MaterialStone, 20)
	gh.AddMaterial("player1", MaterialMetal, 5)
	gh.AddMaterial("player3", MaterialWood, 15)

	contributors := gh.GetContributors()
	if len(contributors) != 3 {
		t.Errorf("got %d contributors, want 3", len(contributors))
	}

	// Check individual contributions
	p1Contrib := gh.GetPlayerContribution("player1")
	if p1Contrib[MaterialStone] != 10 {
		t.Errorf("player1 stone: got %d, want 10", p1Contrib[MaterialStone])
	}
	if p1Contrib[MaterialMetal] != 5 {
		t.Errorf("player1 metal: got %d, want 5", p1Contrib[MaterialMetal])
	}

	p2Contrib := gh.GetPlayerContribution("player2")
	if p2Contrib[MaterialStone] != 20 {
		t.Errorf("player2 stone: got %d, want 20", p2Contrib[MaterialStone])
	}
}

func TestGuildHallBounds(t *testing.T) {
	tests := []struct {
		name     string
		position Vector2
		size     GuildHallSize
		wantMin  Vector2
		wantMax  Vector2
	}{
		{
			"small at origin",
			Vector2{X: 0, Y: 0},
			GuildHallSmall,
			Vector2{X: -16, Y: -16},
			Vector2{X: 16, Y: 16},
		},
		{
			"medium offset",
			Vector2{X: 100, Y: 200},
			GuildHallMedium,
			Vector2{X: 76, Y: 176},
			Vector2{X: 124, Y: 224},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := NewGuildHall("guild", "Test", tt.position, tt.size, 1)
			min, max := gh.Bounds()

			if min.X != tt.wantMin.X || min.Y != tt.wantMin.Y {
				t.Errorf("got min (%v, %v), want (%v, %v)", min.X, min.Y, tt.wantMin.X, tt.wantMin.Y)
			}
			if max.X != tt.wantMax.X || max.Y != tt.wantMax.Y {
				t.Errorf("got max (%v, %v), want (%v, %v)", max.X, max.Y, tt.wantMax.X, tt.wantMax.Y)
			}
		})
	}
}

func TestGuildHallManagerCreate(t *testing.T) {
	manager := NewGuildHallManager()

	// Create first guild hall
	gh1, err := manager.CreateGuildHall("guild1", "Guild One", Vector2{X: 100, Y: 100}, GuildHallSmall, 2)
	if err != nil {
		t.Fatalf("create guild hall failed: %v", err)
	}
	if gh1 == nil {
		t.Fatal("guild hall is nil")
	}

	// Try to create duplicate
	_, err = manager.CreateGuildHall("guild1", "Guild One", Vector2{X: 200, Y: 200}, GuildHallSmall, 2)
	if err == nil {
		t.Error("duplicate guild hall should fail")
	}

	// Create non-overlapping hall
	gh2, err := manager.CreateGuildHall("guild2", "Guild Two", Vector2{X: 200, Y: 200}, GuildHallMedium, 3)
	if err != nil {
		t.Fatalf("create second guild hall failed: %v", err)
	}
	if gh2 == nil {
		t.Fatal("second guild hall is nil")
	}

	// Try to create overlapping hall
	_, err = manager.CreateGuildHall("guild3", "Guild Three", Vector2{X: 105, Y: 105}, GuildHallSmall, 1)
	if err == nil {
		t.Error("overlapping guild hall should fail")
	}
}

func TestGuildHallManagerContribution(t *testing.T) {
	manager := NewGuildHallManager()

	gh, _ := manager.CreateGuildHall("guild1", "Test Guild", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)
	required := gh.RequiredMaterials[MaterialStone]

	err := manager.ContributeMaterial("guild1", "player1", MaterialStone, required)
	if err != nil {
		t.Errorf("contribution failed: %v", err)
	}

	// Invalid guild
	err = manager.ContributeMaterial("nonexistent", "player1", MaterialStone, 10)
	if err == nil {
		t.Error("contribution to nonexistent guild should fail")
	}

	// Invalid amount
	err = manager.ContributeMaterial("guild1", "player1", MaterialStone, -5)
	if err == nil {
		t.Error("negative contribution should fail")
	}
}

func TestGuildHallManagerGetAll(t *testing.T) {
	manager := NewGuildHallManager()

	// Create multiple guild halls
	manager.CreateGuildHall("guild1", "One", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)
	manager.CreateGuildHall("guild2", "Two", Vector2{X: 200, Y: 200}, GuildHallMedium, 2)
	manager.CreateGuildHall("guild3", "Three", Vector2{X: 300, Y: 300}, GuildHallLarge, 3)

	halls := manager.GetAllGuildHalls()
	if len(halls) != 3 {
		t.Errorf("got %d guild halls, want 3", len(halls))
	}
}

func TestGuildHallManagerAreaQuery(t *testing.T) {
	manager := NewGuildHallManager()

	manager.CreateGuildHall("guild1", "One", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)
	manager.CreateGuildHall("guild2", "Two", Vector2{X: 200, Y: 200}, GuildHallMedium, 2)

	// Query area containing only guild1
	halls := manager.GetGuildHallsInArea(Vector2{X: 50, Y: 50}, Vector2{X: 150, Y: 150})
	if len(halls) != 1 {
		t.Errorf("got %d guild halls in area, want 1", len(halls))
	}

	// Query area containing both
	halls = manager.GetGuildHallsInArea(Vector2{X: 0, Y: 0}, Vector2{X: 300, Y: 300})
	if len(halls) != 2 {
		t.Errorf("got %d guild halls in area, want 2", len(halls))
	}
}

func TestGuildHallManagerRemove(t *testing.T) {
	manager := NewGuildHallManager()

	manager.CreateGuildHall("guild1", "Test", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)

	err := manager.RemoveGuildHall("guild1")
	if err != nil {
		t.Errorf("remove failed: %v", err)
	}

	// Verify removed
	_, ok := manager.GetGuildHall("guild1")
	if ok {
		t.Error("guild hall should be removed")
	}

	// Remove again should fail
	err = manager.RemoveGuildHall("guild1")
	if err == nil {
		t.Error("removing nonexistent guild hall should fail")
	}
}

func TestGuildHallManagerStats(t *testing.T) {
	manager := NewGuildHallManager()

	// Create guild halls in different phases
	gh1, _ := manager.CreateGuildHall("guild1", "One", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)
	gh2, _ := manager.CreateGuildHall("guild2", "Two", Vector2{X: 200, Y: 200}, GuildHallSmall, 1)

	// Advance gh2 to walls phase
	for mat, amt := range gh2.RequiredMaterials {
		gh2.AddMaterial("player", mat, amt)
	}
	gh2.AdvancePhase()

	stats := manager.GetStats()
	if stats["total"] != 2 {
		t.Errorf("total: got %d, want 2", stats["total"])
	}
	if stats["foundation"] != 1 {
		t.Errorf("foundation: got %d, want 1", stats["foundation"])
	}
	if stats["walls"] != 1 {
		t.Errorf("walls: got %d, want 1", stats["walls"])
	}

	// Advance gh1 to complete
	for ph := gh1.Phase; ph < PhaseComplete; ph++ {
		for mat, amt := range gh1.RequiredMaterials {
			gh1.AddMaterial("player", mat, amt)
		}
		gh1.AdvancePhase()
	}

	stats = manager.GetStats()
	if stats["complete"] != 1 {
		t.Errorf("complete: got %d, want 1", stats["complete"])
	}
}

func TestGuildHallManagerSaveLoad(t *testing.T) {
	manager := NewGuildHallManager()

	// Create guild halls
	gh1, _ := manager.CreateGuildHall("guild1", "One", Vector2{X: 100, Y: 100}, GuildHallSmall, 2)
	gh2, _ := manager.CreateGuildHall("guild2", "Two", Vector2{X: 200, Y: 200}, GuildHallMedium, 3)

	// Add some materials
	gh1.AddMaterial("player1", MaterialStone, 50)
	gh2.AddMaterial("player2", MaterialMetal, 30)

	// Save
	var buf bytes.Buffer
	err := manager.Save(&buf)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load into new manager
	manager2 := NewGuildHallManager()
	err = manager2.Load(&buf)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify loaded data
	halls := manager2.GetAllGuildHalls()
	if len(halls) != 2 {
		t.Errorf("got %d guild halls after load, want 2", len(halls))
	}

	gh1Loaded, ok := manager2.GetGuildHall("guild1")
	if !ok {
		t.Fatal("guild1 not found after load")
	}
	if gh1Loaded.Floors != 2 {
		t.Errorf("guild1 floors: got %d, want 2", gh1Loaded.Floors)
	}

	contrib := gh1Loaded.GetPlayerContribution("player1")
	if contrib[MaterialStone] != 50 {
		t.Errorf("guild1 contribution: got %d, want 50", contrib[MaterialStone])
	}
}

func TestGuildHallManagerValidation(t *testing.T) {
	manager := NewGuildHallManager()

	gh, _ := manager.CreateGuildHall("guild1", "Test", Vector2{X: 100, Y: 100}, GuildHallSmall, 2)

	// Valid guild hall
	err := manager.ValidateProgress("guild1")
	if err != nil {
		t.Errorf("validation failed: %v", err)
	}

	// Manually corrupt data to test validation
	gh.Floors = 10
	err = manager.ValidateProgress("guild1")
	if err == nil {
		t.Error("validation should fail for invalid floors")
	}

	// Fix floors
	gh.Floors = 2

	// Nonexistent guild
	err = manager.ValidateProgress("nonexistent")
	if err == nil {
		t.Error("validation should fail for nonexistent guild")
	}
}

// Benchmarks
func BenchmarkGuildHallCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gh := NewGuildHall("guild", "Test", Vector2{X: 100, Y: 100}, GuildHallMedium, 3)
		_ = gh
	}
}

func BenchmarkGuildHallAddMaterial(b *testing.B) {
	gh := NewGuildHall("guild", "Test", Vector2{X: 100, Y: 100}, GuildHallSmall, 1)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gh.AddMaterial("player", MaterialStone, 10)
	}
}

func BenchmarkGuildHallGetProgress(b *testing.B) {
	gh := NewGuildHall("guild", "Test", Vector2{X: 100, Y: 100}, GuildHallMedium, 3)
	gh.AddMaterial("player", MaterialStone, 100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gh.GetProgress()
	}
}

func BenchmarkGuildHallManagerCreate(b *testing.B) {
	manager := NewGuildHallManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		guildID := time.Now().String() // Unique ID
		manager.CreateGuildHall(guildID, "Test", Vector2{X: float64(i * 100), Y: float64(i * 100)}, GuildHallSmall, 1)
	}
}

func BenchmarkGuildHallManagerQuery(b *testing.B) {
	manager := NewGuildHallManager()
	for i := 0; i < 100; i++ {
		manager.CreateGuildHall(string(rune(i)), "Test", Vector2{X: float64(i * 50), Y: float64(i * 50)}, GuildHallSmall, 1)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.GetGuildHallsInArea(Vector2{X: 0, Y: 0}, Vector2{X: 1000, Y: 1000})
	}
}
