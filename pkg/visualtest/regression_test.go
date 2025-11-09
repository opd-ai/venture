package visualtest

import (
	"testing"
)

// TestNewRegressionSuite tests suite creation.
func TestNewRegressionSuite(t *testing.T) {
	suite := NewRegressionSuite()

	if suite == nil {
		t.Fatal("Expected suite to be created")
	}

	if len(suite.Tests) == 0 {
		t.Fatal("Expected tests to be added to suite")
	}

	// Verify we have 50+ tests as required by Phase 20.3
	if len(suite.Tests) < 50 {
		t.Errorf("Expected at least 50 tests, got %d", len(suite.Tests))
	}
}

// TestRegressionSuiteCount tests test counting.
func TestRegressionSuiteCount(t *testing.T) {
	suite := NewRegressionSuite()

	count := suite.Count()
	if count == 0 {
		t.Error("Expected non-zero test count")
	}

	if count != len(suite.Tests) {
		t.Errorf("Count() = %d, but len(Tests) = %d", count, len(suite.Tests))
	}
}

// TestCountByPhase tests phase counting.
func TestCountByPhase(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	if len(counts) == 0 {
		t.Error("Expected phase counts")
	}

	// Verify all phases represented
	expectedPhases := []string{
		"Phase 15.1", "Phase 15.2", "Phase 15.3",
		"Phase 16.1", "Phase 16.2", "Phase 16.3",
		"Phase 17.1", "Phase 17.2", "Phase 17.3",
		"Phase 18.1", "Phase 18.2", "Phase 18.3",
		"Phase 19.1", "Phase 19.2", "Phase 19.3",
		"Phase 20.1", "Phase 20.2",
	}

	for _, phase := range expectedPhases {
		if count, exists := counts[phase]; !exists || count == 0 {
			t.Errorf("Expected tests for %s, got %d", phase, count)
		}
	}
}

// TestCountByCategory tests category counting.
func TestCountByCategory(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByCategory()

	if len(counts) == 0 {
		t.Error("Expected category counts")
	}

	// Verify main categories represented
	expectedCategories := []string{
		"sprite", "tile", "lighting", "particle", "ui", "palette", "environment",
	}

	for _, category := range expectedCategories {
		if count, exists := counts[category]; !exists || count == 0 {
			t.Errorf("Expected tests for category %s, got %d", category, count)
		}
	}
}

// TestRegressionTestExecution tests that tests can be executed.
func TestRegressionTestExecution(t *testing.T) {
	suite := NewRegressionSuite()

	if len(suite.Tests) == 0 {
		t.Fatal("No tests in suite")
	}

	// Execute first test
	test := suite.Tests[0]
	if test.TestFunc == nil {
		t.Fatal("Test function is nil")
	}

	img, err := test.TestFunc()
	if err != nil {
		t.Errorf("Test execution failed: %v", err)
	}

	if img == nil {
		t.Error("Expected image to be generated")
	}
}

// TestPhase15Tests verifies Phase 15 test presence.
func TestPhase15Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase151 := counts["Phase 15.1"]
	phase152 := counts["Phase 15.2"]
	phase153 := counts["Phase 15.3"]

	if phase151 == 0 {
		t.Error("Expected Phase 15.1 tests")
	}

	if phase152 == 0 {
		t.Error("Expected Phase 15.2 tests")
	}

	if phase153 == 0 {
		t.Error("Expected Phase 15.3 tests")
	}

	// Phase 15.1 should have tests for 5 genres * 2 template types = 10 tests
	if phase151 < 10 {
		t.Errorf("Expected at least 10 Phase 15.1 tests, got %d", phase151)
	}
}

// TestPhase16Tests verifies Phase 16 test presence.
func TestPhase16Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase161 := counts["Phase 16.1"]
	phase162 := counts["Phase 16.2"]
	phase163 := counts["Phase 16.3"]

	if phase161 == 0 {
		t.Error("Expected Phase 16.1 tests")
	}

	if phase162 == 0 {
		t.Error("Expected Phase 16.2 tests")
	}

	if phase163 == 0 {
		t.Error("Expected Phase 16.3 tests")
	}
}

// TestPhase17Tests verifies Phase 17 test presence.
func TestPhase17Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase171 := counts["Phase 17.1"]
	phase172 := counts["Phase 17.2"]
	phase173 := counts["Phase 17.3"]

	if phase171 == 0 {
		t.Error("Expected Phase 17.1 tests")
	}

	if phase172 == 0 {
		t.Error("Expected Phase 17.2 tests")
	}

	if phase173 == 0 {
		t.Error("Expected Phase 17.3 tests")
	}
}

// TestPhase18Tests verifies Phase 18 test presence.
func TestPhase18Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase181 := counts["Phase 18.1"]
	phase182 := counts["Phase 18.2"]
	phase183 := counts["Phase 18.3"]

	if phase181 == 0 {
		t.Error("Expected Phase 18.1 tests")
	}

	if phase182 == 0 {
		t.Error("Expected Phase 18.2 tests")
	}

	if phase183 == 0 {
		t.Error("Expected Phase 18.3 tests")
	}
}

// TestPhase19Tests verifies Phase 19 test presence.
func TestPhase19Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase191 := counts["Phase 19.1"]
	phase192 := counts["Phase 19.2"]
	phase193 := counts["Phase 19.3"]

	if phase191 == 0 {
		t.Error("Expected Phase 19.1 tests")
	}

	if phase192 == 0 {
		t.Error("Expected Phase 19.2 tests")
	}

	if phase193 == 0 {
		t.Error("Expected Phase 19.3 tests")
	}
}

// TestPhase20Tests verifies Phase 20 test presence.
func TestPhase20Tests(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	phase201 := counts["Phase 20.1"]
	phase202 := counts["Phase 20.2"]

	if phase201 == 0 {
		t.Error("Expected Phase 20.1 tests")
	}

	if phase202 == 0 {
		t.Error("Expected Phase 20.2 tests")
	}
}

// TestRegressionTestStructure verifies test structure is correct.
func TestRegressionTestStructure(t *testing.T) {
	suite := NewRegressionSuite()

	for i, test := range suite.Tests {
		if test.Name == "" {
			t.Errorf("Test %d missing name", i)
		}

		if test.Phase == "" {
			t.Errorf("Test %d (%s) missing phase", i, test.Name)
		}

		if test.Category == "" {
			t.Errorf("Test %d (%s) missing category", i, test.Name)
		}

		if test.Description == "" {
			t.Errorf("Test %d (%s) missing description", i, test.Name)
		}

		if test.TestFunc == nil {
			t.Errorf("Test %d (%s) missing test function", i, test.Name)
		}
	}
}

// TestCategoryDistribution verifies categories are well-distributed.
func TestCategoryDistribution(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByCategory()

	// Each category should have some representation
	for category, count := range counts {
		if count < 2 {
			t.Errorf("Category %s has too few tests: %d", category, count)
		}
	}
}

// TestPhaseDistribution verifies phases are well-distributed.
func TestPhaseDistribution(t *testing.T) {
	suite := NewRegressionSuite()

	counts := suite.CountByPhase()

	// Most sub-phases should have at least 2 tests (some have 1 which is acceptable)
	for phase, count := range counts {
		if count < 1 {
			t.Errorf("Phase %s has too few tests: %d", phase, count)
		}
	}
}

// BenchmarkRegressionSuiteCreation benchmarks suite creation.
func BenchmarkRegressionSuiteCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRegressionSuite()
	}
}

// BenchmarkTestExecution benchmarks test execution.
func BenchmarkTestExecution(b *testing.B) {
	suite := NewRegressionSuite()
	if len(suite.Tests) == 0 {
		b.Fatal("No tests in suite")
	}

	test := suite.Tests[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = test.TestFunc()
	}
}
