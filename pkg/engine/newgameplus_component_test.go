// Package engine provides tests for the New Game Plus component.
package engine

import (
	"testing"
)

func TestNewGamePlusComponent_Type(t *testing.T) {
	ngp := NewNewGamePlusComponent()
	if ngp.Type() != "newgameplus" {
		t.Errorf("Type() = %q, want %q", ngp.Type(), "newgameplus")
	}
}

func TestNewNewGamePlusComponent(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	if ngp.Cycle != 0 {
		t.Errorf("Cycle = %d, want 0", ngp.Cycle)
	}
	if ngp.MaxCycleReached != 0 {
		t.Errorf("MaxCycleReached = %d, want 0", ngp.MaxCycleReached)
	}
	if ngp.CarryOverSlots != 3 {
		t.Errorf("CarryOverSlots = %d, want 3", ngp.CarryOverSlots)
	}
	if ngp.CurrencyCarryOverPercent != 50.0 {
		t.Errorf("CurrencyCarryOverPercent = %f, want 50.0", ngp.CurrencyCarryOverPercent)
	}
	if ngp.LegacyStats == nil {
		t.Error("LegacyStats should not be nil")
	}
	if ngp.IsNewGamePlus() {
		t.Error("IsNewGamePlus() should be false for first playthrough")
	}
}

func TestNewGamePlusComponent_GetCycle(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	if ngp.GetCycle() != 0 {
		t.Errorf("GetCycle() = %d, want 0", ngp.GetCycle())
	}

	ngp.Cycle = 5
	if ngp.GetCycle() != 5 {
		t.Errorf("GetCycle() = %d, want 5", ngp.GetCycle())
	}
}

func TestNewGamePlusComponent_IsNewGamePlus(t *testing.T) {
	tests := []struct {
		name  string
		cycle int
		want  bool
	}{
		{"first playthrough", 0, false},
		{"NG+1", 1, true},
		{"NG+5", 5, true},
		{"NG+99", 99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.cycle
			if got := ngp.IsNewGamePlus(); got != tt.want {
				t.Errorf("IsNewGamePlus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGamePlusComponent_LegacyStats(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Initially zero
	if ngp.GetLegacyStat("enemies_killed") != 0 {
		t.Error("Initial legacy stat should be 0")
	}

	// Add to stat
	ngp.AddToLegacyStat("enemies_killed", 100)
	if ngp.GetLegacyStat("enemies_killed") != 100 {
		t.Errorf("GetLegacyStat() = %d, want 100", ngp.GetLegacyStat("enemies_killed"))
	}

	// Add more
	ngp.AddToLegacyStat("enemies_killed", 50)
	if ngp.GetLegacyStat("enemies_killed") != 150 {
		t.Errorf("GetLegacyStat() = %d, want 150", ngp.GetLegacyStat("enemies_killed"))
	}

	// Different stat
	ngp.AddToLegacyStat("quests_completed", 10)
	if ngp.GetLegacyStat("quests_completed") != 10 {
		t.Errorf("GetLegacyStat() = %d, want 10", ngp.GetLegacyStat("quests_completed"))
	}
}

func TestNewGamePlusComponent_Playtime(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Initial state
	if ngp.GetTotalPlaytime() != 0 {
		t.Error("Initial total playtime should be 0")
	}
	if ngp.GetCurrentCyclePlaytime() != 0 {
		t.Error("Initial cycle playtime should be 0")
	}

	// Update playtime
	ngp.UpdatePlaytime(3600) // 1 hour
	if ngp.GetTotalPlaytime() != 3600 {
		t.Errorf("GetTotalPlaytime() = %d, want 3600", ngp.GetTotalPlaytime())
	}
	if ngp.GetCurrentCyclePlaytime() != 3600 {
		t.Errorf("GetCurrentCyclePlaytime() = %d, want 3600", ngp.GetCurrentCyclePlaytime())
	}

	// Add more
	ngp.UpdatePlaytime(1800) // 30 minutes
	if ngp.GetTotalPlaytime() != 5400 {
		t.Errorf("GetTotalPlaytime() = %d, want 5400", ngp.GetTotalPlaytime())
	}
}

func TestNewGamePlusComponent_Bonuses(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Initially no bonuses
	if ngp.HasBonus("speed_boost") {
		t.Error("Should not have bonus initially")
	}

	// Unlock bonus
	if !ngp.UnlockBonus("speed_boost") {
		t.Error("UnlockBonus should return true for new bonus")
	}
	if !ngp.HasBonus("speed_boost") {
		t.Error("Should have bonus after unlock")
	}

	// Try to unlock again (should fail)
	if ngp.UnlockBonus("speed_boost") {
		t.Error("UnlockBonus should return false for duplicate")
	}

	// Unlock different bonus
	ngp.UnlockBonus("damage_boost")
	if !ngp.HasBonus("damage_boost") {
		t.Error("Should have second bonus")
	}
}

func TestNewGamePlusComponent_StartNewCycle(t *testing.T) {
	ngp := NewNewGamePlusComponent()
	ngp.CurrentCyclePlaytime = 7200 // 2 hours

	stats := map[string]int64{
		"enemies_killed":   500,
		"quests_completed": 20,
		"deaths":           5,
	}

	// Start NG+1
	ngp.StartNewCycle(stats)

	if ngp.Cycle != 1 {
		t.Errorf("Cycle = %d, want 1", ngp.Cycle)
	}
	if ngp.MaxCycleReached != 1 {
		t.Errorf("MaxCycleReached = %d, want 1", ngp.MaxCycleReached)
	}
	if ngp.CurrentCyclePlaytime != 0 {
		t.Error("CurrentCyclePlaytime should reset to 0")
	}
	if ngp.CarryOverSlots != 4 {
		t.Errorf("CarryOverSlots = %d, want 4", ngp.CarryOverSlots)
	}
	if ngp.CurrencyCarryOverPercent != 55.0 {
		t.Errorf("CurrencyCarryOverPercent = %f, want 55.0", ngp.CurrencyCarryOverPercent)
	}

	// Check legacy stats accumulated
	if ngp.GetLegacyStat("enemies_killed") != 500 {
		t.Errorf("Legacy enemies_killed = %d, want 500", ngp.GetLegacyStat("enemies_killed"))
	}

	// Check cycle record
	cycles := ngp.GetCompletedCycles()
	if len(cycles) != 1 {
		t.Errorf("CompletedCycles length = %d, want 1", len(cycles))
	}
	if cycles[0].CycleNumber != 0 {
		t.Errorf("Recorded cycle number = %d, want 0", cycles[0].CycleNumber)
	}
	if cycles[0].EnemiesKilled != 500 {
		t.Errorf("Recorded enemies killed = %d, want 500", cycles[0].EnemiesKilled)
	}
}

func TestNewGamePlusComponent_MultipleCycles(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Complete multiple cycles
	for i := 0; i < 10; i++ {
		stats := map[string]int64{
			"enemies_killed":   100,
			"quests_completed": 10,
			"deaths":           1,
		}
		ngp.StartNewCycle(stats)
	}

	if ngp.Cycle != 10 {
		t.Errorf("Cycle = %d, want 10", ngp.Cycle)
	}
	if ngp.MaxCycleReached != 10 {
		t.Errorf("MaxCycleReached = %d, want 10", ngp.MaxCycleReached)
	}

	// Check carry-over caps
	if ngp.CarryOverSlots != 10 {
		t.Errorf("CarryOverSlots = %d, want 10 (capped)", ngp.CarryOverSlots)
	}
	if ngp.CurrencyCarryOverPercent != 100.0 {
		t.Errorf("CurrencyCarryOverPercent = %f, want 100.0 (capped)", ngp.CurrencyCarryOverPercent)
	}

	// Check legacy stats accumulated
	if ngp.GetLegacyStat("enemies_killed") != 1000 {
		t.Errorf("Legacy enemies_killed = %d, want 1000", ngp.GetLegacyStat("enemies_killed"))
	}

	// Check all cycles recorded
	cycles := ngp.GetCompletedCycles()
	if len(cycles) != 10 {
		t.Errorf("CompletedCycles length = %d, want 10", len(cycles))
	}
}

func TestNewGamePlusComponent_Serialize(t *testing.T) {
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 3
	ngp.MaxCycleReached = 5
	ngp.AddToLegacyStat("enemies_killed", 1000)
	ngp.TotalPlaytime = 36000
	ngp.UnlockBonus("speed_boost")

	data, err := ngp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("Serialized data should not be empty")
	}

	// Deserialize into new component
	ngp2 := NewNewGamePlusComponent()
	if err := ngp2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if ngp2.Cycle != 3 {
		t.Errorf("Deserialized Cycle = %d, want 3", ngp2.Cycle)
	}
	if ngp2.MaxCycleReached != 5 {
		t.Errorf("Deserialized MaxCycleReached = %d, want 5", ngp2.MaxCycleReached)
	}
	if ngp2.GetLegacyStat("enemies_killed") != 1000 {
		t.Errorf("Deserialized enemies_killed = %d, want 1000", ngp2.GetLegacyStat("enemies_killed"))
	}
	if ngp2.TotalPlaytime != 36000 {
		t.Errorf("Deserialized TotalPlaytime = %d, want 36000", ngp2.TotalPlaytime)
	}
	if !ngp2.HasBonus("speed_boost") {
		t.Error("Deserialized component should have speed_boost bonus")
	}
}

func TestNewGamePlusComponent_DeserializeInvalid(t *testing.T) {
	ngp := NewNewGamePlusComponent()
	err := ngp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize should return error for invalid JSON")
	}
}

func TestNewGamePlusComponent_GetNGPlusLabel(t *testing.T) {
	tests := []struct {
		cycle int
		want  string
	}{
		{0, ""},
		{1, "NG+"},
		{2, "NG+2"},
		{5, "NG+5"},
		{9, "NG+9"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.cycle
			if got := ngp.GetNGPlusLabel(); got != tt.want {
				t.Errorf("GetNGPlusLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewGamePlusComponent_GetNGPlusLabelFull(t *testing.T) {
	tests := []struct {
		cycle int
		want  string
	}{
		{0, "First Playthrough"},
		{1, "New Game Plus"},
		{2, "New Game Plus 2"},
		{5, "New Game Plus 5"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			ngp := NewNewGamePlusComponent()
			ngp.Cycle = tt.cycle
			if got := ngp.GetNGPlusLabelFull(); got != tt.want {
				t.Errorf("GetNGPlusLabelFull() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewGamePlusComponent_ThreadSafety(t *testing.T) {
	ngp := NewNewGamePlusComponent()
	done := make(chan bool)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				ngp.GetCycle()
				ngp.GetLegacyStat("test")
				ngp.AddToLegacyStat("test", 1)
				ngp.UpdatePlaytime(1)
				ngp.IsNewGamePlus()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state is consistent
	if ngp.GetLegacyStat("test") != 1000 {
		t.Errorf("Expected 1000 from concurrent adds, got %d", ngp.GetLegacyStat("test"))
	}
}

func TestNewGamePlusComponent_CarryOverSlotsCap(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Force to max and beyond
	for i := 0; i < 20; i++ {
		ngp.StartNewCycle(map[string]int64{})
	}

	if ngp.CarryOverSlots > 10 {
		t.Errorf("CarryOverSlots = %d, should be capped at 10", ngp.CarryOverSlots)
	}
}

func TestNewGamePlusComponent_CurrencyCarryOverCap(t *testing.T) {
	ngp := NewNewGamePlusComponent()

	// Force to max and beyond
	for i := 0; i < 20; i++ {
		ngp.StartNewCycle(map[string]int64{})
	}

	if ngp.CurrencyCarryOverPercent > 100.0 {
		t.Errorf("CurrencyCarryOverPercent = %f, should be capped at 100.0", ngp.CurrencyCarryOverPercent)
	}
}

func TestFormatCycleNumber(t *testing.T) {
	tests := []struct {
		cycle int
		want  string
	}{
		{1, "1"},
		{5, "5"},
		{9, "9"},
		{10, "10"},
		{25, "25"},
		{99, "99"},
	}

	for _, tt := range tests {
		got := formatCycleNumber(tt.cycle)
		if got != tt.want {
			t.Errorf("formatCycleNumber(%d) = %q, want %q", tt.cycle, got, tt.want)
		}
	}
}

func BenchmarkNewGamePlusComponent_GetCycle(b *testing.B) {
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ngp.GetCycle()
	}
}

func BenchmarkNewGamePlusComponent_AddToLegacyStat(b *testing.B) {
	ngp := NewNewGamePlusComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ngp.AddToLegacyStat("enemies_killed", 1)
	}
}

func BenchmarkNewGamePlusComponent_Serialize(b *testing.B) {
	ngp := NewNewGamePlusComponent()
	ngp.Cycle = 5
	ngp.AddToLegacyStat("enemies_killed", 10000)
	ngp.UnlockBonus("test_bonus")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ngp.Serialize()
	}
}
