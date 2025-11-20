package territory_siege

import (
	"testing"
	"time"
)

// TestSiegePhase_String tests the String() method for SiegePhase.
func TestSiegePhase_String(t *testing.T) {
	tests := []struct {
		phase    SiegePhase
		expected string
	}{
		{PhasePreparation, "Preparation"},
		{PhaseAssault, "Assault"},
		{PhaseResolution, "Resolution"},
		{SiegePhase(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.phase.String(); got != tt.expected {
			t.Errorf("SiegePhase(%d).String() = %s, want %s", tt.phase, got, tt.expected)
		}
	}
}

// TestStructureType_String tests the String() method for StructureType.
func TestStructureType_String(t *testing.T) {
	tests := []struct {
		structType StructureType
		expected   string
	}{
		{StructureWall, "Wall"},
		{StructureTower, "Tower"},
		{StructureGate, "Gate"},
		{StructureBarracks, "Barracks"},
		{StructureKeep, "Keep"},
		{StructureType(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.structType.String(); got != tt.expected {
			t.Errorf("StructureType(%d).String() = %s, want %s", tt.structType, got, tt.expected)
		}
	}
}

// TestVictoryCondition_String tests the String() method for VictoryCondition.
func TestVictoryCondition_String(t *testing.T) {
	tests := []struct {
		condition VictoryCondition
		expected  string
	}{
		{VictoryConditionNone, "Ongoing"},
		{VictoryConditionAllPointsCaptured, "All Control Points Captured"},
		{VictoryConditionGuildHallDestroyed, "Guild Hall Destroyed"},
		{VictoryConditionTimeExpired, "Time Expired"},
		{VictoryConditionAttackersEliminated, "Attackers Eliminated"},
		{VictoryCondition(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.condition.String(); got != tt.expected {
			t.Errorf("VictoryCondition(%d).String() = %s, want %s", tt.condition, got, tt.expected)
		}
	}
}

// TestDefensiveStructure_IsStructureDestroyed tests structure destruction checking.
func TestDefensiveStructure_IsStructureDestroyed(t *testing.T) {
	tests := []struct {
		name        string
		currentHP   int
		isDestroyed bool
		expected    bool
	}{
		{"Full HP", 100, false, false},
		{"Partial damage", 50, false, false},
		{"Zero HP", 0, false, true},
		{"Negative HP", -10, false, true},
		{"Flagged destroyed", 100, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DefensiveStructure{
				CurrentHP:   tt.currentHP,
				IsDestroyed: tt.isDestroyed,
			}
			if got := ds.IsStructureDestroyed(); got != tt.expected {
				t.Errorf("IsStructureDestroyed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDefensiveStructure_TakeDamage tests damage application to structures.
func TestDefensiveStructure_TakeDamage(t *testing.T) {
	tests := []struct {
		name          string
		initialHP     int
		damage        int
		expectedHP    int
		shouldDestroy bool
	}{
		{"Partial damage", 100, 30, 70, false},
		{"Exact destruction", 100, 100, 0, true},
		{"Overkill damage", 100, 150, 0, true},
		{"Already destroyed", 0, 50, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DefensiveStructure{
				CurrentHP:   tt.initialHP,
				MaxHP:       100,
				IsDestroyed: tt.initialHP <= 0,
			}

			before := time.Now().Unix()
			ds.TakeDamage(tt.damage)
			after := time.Now().Unix()

			if ds.CurrentHP != tt.expectedHP {
				t.Errorf("CurrentHP = %d, want %d", ds.CurrentHP, tt.expectedHP)
			}

			if ds.IsDestroyed != tt.shouldDestroy {
				t.Errorf("IsDestroyed = %v, want %v", ds.IsDestroyed, tt.shouldDestroy)
			}

			// Check timestamp updated (unless already destroyed at start)
			if !tt.shouldDestroy || tt.initialHP > 0 {
				if ds.LastDamageAt < before || ds.LastDamageAt > after {
					t.Errorf("LastDamageAt not updated correctly: %d not in [%d, %d]", ds.LastDamageAt, before, after)
				}
			}
		})
	}
}

// TestSiege_GetElapsedTime tests elapsed time calculation.
func TestSiege_GetElapsedTime(t *testing.T) {
	siege := &Siege{
		PhaseStartTime: time.Now().Unix() - 60, // 60 seconds ago
	}

	elapsed := siege.GetElapsedTime()
	if elapsed < 59 || elapsed > 61 {
		t.Errorf("GetElapsedTime() = %d, want ~60", elapsed)
	}
}

// TestSiege_GetRemainingTime tests remaining time calculation for each phase.
func TestSiege_GetRemainingTime(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name         string
		phase        SiegePhase
		phaseStart   int64
		prepDuration int64
		assaultDur   int64
		wantMin      int64
		wantMax      int64
	}{
		{"Preparation just started", PhasePreparation, now, 3600, 7200, 3598, 3600},
		{"Preparation halfway", PhasePreparation, now - 1800, 3600, 7200, 1798, 1800},
		{"Assault just started", PhaseAssault, now, 3600, 7200, 7198, 7200},
		{"Assault halfway", PhaseAssault, now - 3600, 3600, 7200, 3598, 3600},
		{"Resolution phase", PhaseResolution, now, 3600, 7200, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			siege := &Siege{
				CurrentPhase:        tt.phase,
				PhaseStartTime:      tt.phaseStart,
				PreparationDuration: tt.prepDuration,
				AssaultDuration:     tt.assaultDur,
			}

			got := siege.GetRemainingTime()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetRemainingTime() = %d, want in range [%d, %d]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestSiege_ShouldAdvancePhase tests phase advancement logic.
func TestSiege_ShouldAdvancePhase(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name       string
		phase      SiegePhase
		phaseStart int64
		expected   bool
	}{
		{"Preparation not expired", PhasePreparation, now, false},
		{"Preparation expired", PhasePreparation, now - 3601, true},
		{"Assault not expired", PhaseAssault, now, false},
		{"Assault expired", PhaseAssault, now - 7201, true},
		{"Resolution phase", PhaseResolution, now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			siege := &Siege{
				CurrentPhase:        tt.phase,
				PhaseStartTime:      tt.phaseStart,
				PreparationDuration: 3600,
				AssaultDuration:     7200,
			}

			if got := siege.ShouldAdvancePhase(); got != tt.expected {
				t.Errorf("ShouldAdvancePhase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSiege_CountDestroyedStructures tests counting destroyed structures.
func TestSiege_CountDestroyedStructures(t *testing.T) {
	siege := &Siege{
		DefensiveStructures: []*DefensiveStructure{
			{CurrentHP: 100, IsDestroyed: false},
			{CurrentHP: 0, IsDestroyed: true},
			{CurrentHP: 50, IsDestroyed: false},
			{CurrentHP: 0, IsDestroyed: false}, // HP zero but not flagged (still counts)
			{CurrentHP: 100, IsDestroyed: false},
		},
	}

	got := siege.CountDestroyedStructures()
	want := 2 // Two structures with CurrentHP <= 0 or IsDestroyed=true

	if got != want {
		t.Errorf("CountDestroyedStructures() = %d, want %d", got, want)
	}
}

// TestSiege_GetDestructionPercentage tests destruction percentage calculation.
func TestSiege_GetDestructionPercentage(t *testing.T) {
	tests := []struct {
		name        string
		destroyed   int
		total       int
		expectedPct float64
	}{
		{"No structures", 0, 0, 0.0},
		{"None destroyed", 0, 10, 0.0},
		{"Half destroyed", 5, 10, 0.5},
		{"All destroyed", 10, 10, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structures := make([]*DefensiveStructure, tt.total)
			for i := 0; i < tt.total; i++ {
				if i < tt.destroyed {
					structures[i] = &DefensiveStructure{CurrentHP: 0, IsDestroyed: true}
				} else {
					structures[i] = &DefensiveStructure{CurrentHP: 100, IsDestroyed: false}
				}
			}

			siege := &Siege{DefensiveStructures: structures}
			got := siege.GetDestructionPercentage()

			if got != tt.expectedPct {
				t.Errorf("GetDestructionPercentage() = %f, want %f", got, tt.expectedPct)
			}
		})
	}
}

// BenchmarkDefensiveStructure_TakeDamage benchmarks damage application.
func BenchmarkDefensiveStructure_TakeDamage(b *testing.B) {
	ds := &DefensiveStructure{
		CurrentHP:   10000,
		MaxHP:       10000,
		IsDestroyed: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ds.TakeDamage(10)
	}
}

// BenchmarkSiege_GetElapsedTime benchmarks elapsed time calculation.
func BenchmarkSiege_GetElapsedTime(b *testing.B) {
	siege := &Siege{
		PhaseStartTime: time.Now().Unix() - 60,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = siege.GetElapsedTime()
	}
}

// BenchmarkSiege_CountDestroyedStructures benchmarks structure counting.
func BenchmarkSiege_CountDestroyedStructures(b *testing.B) {
	structures := make([]*DefensiveStructure, 15)
	for i := range structures {
		structures[i] = &DefensiveStructure{
			CurrentHP:   100 - i*10,
			IsDestroyed: i%2 == 0,
		}
	}

	siege := &Siege{DefensiveStructures: structures}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = siege.CountDestroyedStructures()
	}
}
