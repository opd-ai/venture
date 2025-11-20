package territory_siege

import (
	"math/rand"
	"testing"
)

// TestStructureGenerator_GenerateStructures tests structure generation.
func TestStructureGenerator_GenerateStructures(t *testing.T) {
	source := rand.NewSource(12345)
	sg := NewStructureGenerator(source)

	tests := []struct {
		name  string
		count int
	}{
		{"Minimum structures", 5},
		{"Medium structures", 10},
		{"Maximum structures", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			structures := sg.GenerateStructures("zone1", tt.count)

			if len(structures) != tt.count {
				t.Errorf("GenerateStructures() generated %d structures, want %d", len(structures), tt.count)
			}

			// Verify first structure is always a keep
			if structures[0].Type != StructureKeep {
				t.Errorf("First structure is %s, want Keep", structures[0].Type.String())
			}

			// Verify all structures have valid HP and positions
			for i, s := range structures {
				if s.MaxHP <= 0 {
					t.Errorf("Structure %d has invalid MaxHP: %d", i, s.MaxHP)
				}

				if s.CurrentHP != s.MaxHP {
					t.Errorf("Structure %d CurrentHP (%d) != MaxHP (%d)", i, s.CurrentHP, s.MaxHP)
				}

				if s.X < 0 || s.X > 1000 || s.Y < 0 || s.Y > 1000 {
					t.Errorf("Structure %d position out of bounds: (%f, %f)", i, s.X, s.Y)
				}

				if s.IsDestroyed {
					t.Errorf("Structure %d is destroyed on generation", i)
				}

				// Verify HP is within type range
				min, max := GetHPRange(s.Type)
				if s.MaxHP < min || s.MaxHP > max {
					t.Errorf("Structure %d (%s) HP %d not in range [%d, %d]", i, s.Type.String(), s.MaxHP, min, max)
				}
			}
		})
	}
}

// TestStructureGenerator_Determinism tests deterministic generation.
func TestStructureGenerator_Determinism(t *testing.T) {
	seed := int64(12345)

	// Generate structures twice with same seed
	sg1 := NewStructureGenerator(rand.NewSource(seed))
	structures1 := sg1.GenerateStructures("zone1", 10)

	sg2 := NewStructureGenerator(rand.NewSource(seed))
	structures2 := sg2.GenerateStructures("zone1", 10)

	if len(structures1) != len(structures2) {
		t.Fatalf("Determinism broken: got %d and %d structures", len(structures1), len(structures2))
	}

	for i := range structures1 {
		s1 := structures1[i]
		s2 := structures2[i]

		if s1.Type != s2.Type {
			t.Errorf("Structure %d type mismatch: %s vs %s", i, s1.Type.String(), s2.Type.String())
		}

		if s1.MaxHP != s2.MaxHP {
			t.Errorf("Structure %d HP mismatch: %d vs %d", i, s1.MaxHP, s2.MaxHP)
		}

		if s1.X != s2.X || s1.Y != s2.Y {
			t.Errorf("Structure %d position mismatch: (%f, %f) vs (%f, %f)", i, s1.X, s1.Y, s2.X, s2.Y)
		}
	}
}

// TestStructureGenerator_TypeDistribution tests structure type distribution.
func TestStructureGenerator_TypeDistribution(t *testing.T) {
	source := rand.NewSource(12345)
	sg := NewStructureGenerator(source)

	structures := sg.GenerateStructures("zone1", 15)

	typeCounts := make(map[StructureType]int)
	for _, s := range structures {
		typeCounts[s.Type]++
	}

	// Verify we have exactly one keep
	if typeCounts[StructureKeep] != 1 {
		t.Errorf("Got %d keeps, want exactly 1", typeCounts[StructureKeep])
	}

	// Verify we have multiple types (diversity check)
	if len(typeCounts) < 3 {
		t.Errorf("Only %d structure types generated, want at least 3 for variety", len(typeCounts))
	}

	// Walls should be most common (~33%)
	if typeCounts[StructureWall] < 3 {
		t.Errorf("Only %d walls generated, expected more for 15-structure set", typeCounts[StructureWall])
	}
}

// TestGetHPRange tests HP range retrieval.
func TestGetHPRange(t *testing.T) {
	tests := []struct {
		structType StructureType
		wantMin    int
		wantMax    int
	}{
		{StructureWall, 1000, 5000},
		{StructureTower, 500, 1500},
		{StructureGate, 800, 2000},
		{StructureBarracks, 300, 800},
		{StructureKeep, 10000, 20000},
		{StructureType(99), 0, 0}, // Invalid type
	}

	for _, tt := range tests {
		t.Run(tt.structType.String(), func(t *testing.T) {
			gotMin, gotMax := GetHPRange(tt.structType)

			if gotMin != tt.wantMin {
				t.Errorf("GetHPRange(%s) min = %d, want %d", tt.structType.String(), gotMin, tt.wantMin)
			}

			if gotMax != tt.wantMax {
				t.Errorf("GetHPRange(%s) max = %d, want %d", tt.structType.String(), gotMax, tt.wantMax)
			}
		})
	}
}

// BenchmarkStructureGenerator_GenerateStructures benchmarks structure generation.
func BenchmarkStructureGenerator_GenerateStructures(b *testing.B) {
	source := rand.NewSource(12345)
	sg := NewStructureGenerator(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sg.GenerateStructures("zone1", 10)
	}
}

// BenchmarkStructureGenerator_generateStructure benchmarks single structure generation.
func BenchmarkStructureGenerator_generateStructure(b *testing.B) {
	source := rand.NewSource(12345)
	sg := NewStructureGenerator(source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sg.generateStructure("zone1", StructureWall, i)
	}
}
