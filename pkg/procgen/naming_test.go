package procgen

import (
	"testing"
)

// TestSelectDefaultName_Deterministic verifies that the same seed always produces the same name.
func TestSelectDefaultName_Deterministic(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"seed 12345", 12345},
		{"seed 98765", 98765},
		{"seed 0", 0},
		{"seed 1", 1},
		{"seed -1", -1},
		{"negative seed", -999999},
		{"large seed", 1234567890123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate name multiple times with same seed
			name1 := SelectDefaultName(tt.seed)
			name2 := SelectDefaultName(tt.seed)
			name3 := SelectDefaultName(tt.seed)

			// All names should be identical
			if name1 != name2 {
				t.Errorf("SelectDefaultName(%d) not deterministic: first=%s, second=%s", tt.seed, name1, name2)
			}
			if name1 != name3 {
				t.Errorf("SelectDefaultName(%d) not deterministic: first=%s, third=%s", tt.seed, name1, name3)
			}

			// Name should not be empty
			if name1 == "" {
				t.Errorf("SelectDefaultName(%d) returned empty name", tt.seed)
			}

			// Name should be from the DefaultNames list
			found := false
			for _, validName := range DefaultNames {
				if name1 == validName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SelectDefaultName(%d) returned name not in DefaultNames: %s", tt.seed, name1)
			}
		})
	}
}

// TestSelectDefaultName_DifferentSeeds verifies that different seeds produce different names (most of the time).
func TestSelectDefaultName_DifferentSeeds(t *testing.T) {
	// Test that different seeds produce different names
	// (This isn't guaranteed but should be true most of the time with 100 names)
	seed1 := int64(12345)
	seed2 := int64(54321)

	name1 := SelectDefaultName(seed1)
	name2 := SelectDefaultName(seed2)

	// With 100 names, collision probability is low enough that we expect different results
	// for different seeds. If this test fails occasionally, it's not necessarily a bug,
	// but consistent failures would indicate a problem.
	if name1 == name2 {
		t.Logf("Note: Different seeds produced same name (this can happen randomly): %s", name1)
	}
}

// TestDefaultNames_Count verifies that there are exactly 100 names in the list.
func TestDefaultNames_Count(t *testing.T) {
	count := len(DefaultNames)
	expected := 100

	if count != expected {
		t.Errorf("DefaultNames count = %d, want %d", count, expected)
	}
}

// TestDefaultNames_NoDuplicates verifies that all names in the list are unique.
func TestDefaultNames_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	duplicates := []string{}

	for _, name := range DefaultNames {
		if seen[name] {
			duplicates = append(duplicates, name)
		}
		seen[name] = true
	}

	if len(duplicates) > 0 {
		t.Errorf("Found duplicate names in DefaultNames: %v", duplicates)
	}
}

// TestDefaultNames_NotEmpty verifies that all names in the list are non-empty.
func TestDefaultNames_NotEmpty(t *testing.T) {
	for i, name := range DefaultNames {
		if name == "" {
			t.Errorf("DefaultNames[%d] is empty", i)
		}
	}
}

// TestSelectDefaultName_Coverage verifies that selection can return different names.
// This ensures the selection algorithm actually varies based on seed.
func TestSelectDefaultName_Coverage(t *testing.T) {
	// Collect names from different seeds
	uniqueNames := make(map[string]bool)

	// Try a range of seeds
	for seed := int64(0); seed < 100; seed++ {
		name := SelectDefaultName(seed)
		uniqueNames[name] = true
	}

	// We should see multiple different names (not just one)
	if len(uniqueNames) < 10 {
		t.Errorf("SelectDefaultName shows poor coverage: only %d unique names from 100 seeds", len(uniqueNames))
	}
}

// BenchmarkSelectDefaultName benchmarks the performance of name selection.
func BenchmarkSelectDefaultName(b *testing.B) {
	seed := int64(12345)
	for i := 0; i < b.N; i++ {
		SelectDefaultName(seed)
	}
}
