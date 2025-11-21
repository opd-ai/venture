package features

import (
	"testing"
	"time"
)

func TestFeatureValidation(t *testing.T) {
	tests := []struct {
		name       string
		feature    *Feature
		wantValid  bool
		wantIssues int
	}{
		{
			name: "fully valid feature",
			feature: &Feature{
				ID:                   "test.valid",
				Name:                 "Valid Feature",
				Category:             CategoryCore,
				Accessible:           true,
				AccessibilityTime:    5 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 1.0,
				IntegratedSystems:    []string{"system1", "system2", "system3"},
				IntegrationCount:     3,
				Implemented:          true,
				Functional:           true,
			},
			wantValid:  true,
			wantIssues: 0,
		},
		{
			name: "not accessible",
			feature: &Feature{
				ID:                   "test.inaccessible",
				Name:                 "Inaccessible Feature",
				Category:             CategoryCore,
				Accessible:           false, // Problem
				AccessibilityTime:    40 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 1.0,
				IntegratedSystems:    []string{"system1", "system2"},
				IntegrationCount:     2,
				Implemented:          true,
				Functional:           true,
			},
			wantValid:  false,
			wantIssues: 1,
		},
		{
			name: "incomplete tutorial",
			feature: &Feature{
				ID:                   "test.badtutorial",
				Name:                 "Bad Tutorial",
				Category:             CategoryCore,
				Accessible:           true,
				AccessibilityTime:    5 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 0.5, // Too low
				IntegratedSystems:    []string{"system1", "system2"},
				IntegrationCount:     2,
				Implemented:          true,
				Functional:           true,
			},
			wantValid:  false,
			wantIssues: 1,
		},
		{
			name: "insufficient integration",
			feature: &Feature{
				ID:                   "test.nointeg",
				Name:                 "No Integration",
				Category:             CategoryCore,
				Accessible:           true,
				AccessibilityTime:    5 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 1.0,
				IntegratedSystems:    []string{"system1"}, // Only 1
				IntegrationCount:     1,
				Implemented:          true,
				Functional:           true,
			},
			wantValid:  false,
			wantIssues: 1,
		},
		{
			name: "not implemented",
			feature: &Feature{
				ID:                   "test.notimpl",
				Name:                 "Not Implemented",
				Category:             CategoryCore,
				Accessible:           true,
				AccessibilityTime:    5 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 1.0,
				IntegratedSystems:    []string{"system1", "system2"},
				IntegrationCount:     2,
				Implemented:          false, // Problem
				Functional:           true,
			},
			wantValid:  false,
			wantIssues: 1,
		},
		{
			name: "not functional",
			feature: &Feature{
				ID:                   "test.broken",
				Name:                 "Broken",
				Category:             CategoryCore,
				Accessible:           true,
				AccessibilityTime:    5 * time.Minute,
				HasTutorial:          true,
				TutorialCompleteness: 1.0,
				IntegratedSystems:    []string{"system1", "system2"},
				IntegrationCount:     2,
				Implemented:          true,
				Functional:           false, // Problem
			},
			wantValid:  false,
			wantIssues: 1,
		},
		{
			name: "multiple issues",
			feature: &Feature{
				ID:                   "test.multiple",
				Name:                 "Multiple Issues",
				Category:             CategoryCore,
				Accessible:           false, // Issue 1
				AccessibilityTime:    40 * time.Minute,
				HasTutorial:          false, // Issue 2
				TutorialCompleteness: 0.0,
				IntegratedSystems:    []string{}, // Issue 3
				IntegrationCount:     0,
				Implemented:          false, // Issue 4
				Functional:           false, // Issue 5
			},
			wantValid:  false,
			wantIssues: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, issues := tt.feature.Validate()

			if valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", valid, tt.wantValid)
			}

			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() issues count = %d, want %d. Issues: %v", len(issues), tt.wantIssues, issues)
			}

			if tt.feature.IsComplete() != tt.wantValid {
				t.Errorf("IsComplete() = %v, want %v", tt.feature.IsComplete(), tt.wantValid)
			}
		})
	}
}

func TestFeatureRegistry(t *testing.T) {
	r := NewFeatureRegistry()

	// Test empty registry
	if len(r.GetAll()) != 0 {
		t.Errorf("Empty registry should have 0 features, got %d", len(r.GetAll()))
	}

	// Register some features
	f1 := &Feature{
		ID:       "test.f1",
		Name:     "Feature 1",
		Category: CategoryCore,
	}
	f2 := &Feature{
		ID:       "test.f2",
		Name:     "Feature 2",
		Category: CategoryCore,
	}
	f3 := &Feature{
		ID:       "test.f3",
		Name:     "Feature 3",
		Category: CategoryAdvanced,
	}

	r.Register(f1)
	r.Register(f2)
	r.Register(f3)

	// Test GetAll
	all := r.GetAll()
	if len(all) != 3 {
		t.Errorf("Registry should have 3 features, got %d", len(all))
	}

	// Test GetFeature
	retrieved := r.GetFeature("test.f1")
	if retrieved == nil {
		t.Error("GetFeature should return feature, got nil")
	} else if retrieved.ID != "test.f1" {
		t.Errorf("GetFeature returned wrong feature: %s", retrieved.ID)
	}

	// Test GetCategory
	coreFeatures := r.GetCategory(CategoryCore)
	if len(coreFeatures) != 2 {
		t.Errorf("Core category should have 2 features, got %d", len(coreFeatures))
	}

	advancedFeatures := r.GetCategory(CategoryAdvanced)
	if len(advancedFeatures) != 1 {
		t.Errorf("Advanced category should have 1 feature, got %d", len(advancedFeatures))
	}
}

func TestValidationReport(t *testing.T) {
	r := NewFeatureRegistry()

	// Add mix of valid and invalid features
	r.Register(&Feature{
		ID:                   "test.valid1",
		Name:                 "Valid 1",
		Category:             CategoryCore,
		Accessible:           true,
		HasTutorial:          true,
		TutorialCompleteness: 1.0,
		IntegrationCount:     3,
		Implemented:          true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "test.valid2",
		Name:                 "Valid 2",
		Category:             CategoryCore,
		Accessible:           true,
		HasTutorial:          true,
		TutorialCompleteness: 0.9,
		IntegrationCount:     2,
		Implemented:          true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "test.invalid1",
		Name:                 "Invalid 1",
		Category:             CategoryAdvanced,
		Accessible:           false, // Problem
		HasTutorial:          true,
		TutorialCompleteness: 1.0,
		IntegrationCount:     2,
		Implemented:          true,
		Functional:           true,
	})

	report := r.ValidateAll()

	// Test counts
	if report.TotalFeatures != 3 {
		t.Errorf("Report should show 3 total features, got %d", report.TotalFeatures)
	}

	if report.PassedFeatures != 2 {
		t.Errorf("Report should show 2 passed features, got %d", report.PassedFeatures)
	}

	if report.FailedFeatures != 1 {
		t.Errorf("Report should show 1 failed feature, got %d", report.FailedFeatures)
	}

	// Test pass rate
	expectedPassRate := 2.0 / 3.0
	if report.PassRate != expectedPassRate {
		t.Errorf("Pass rate = %.2f, want %.2f", report.PassRate, expectedPassRate)
	}

	// Test IsAcceptable (should fail, need 90%)
	if report.IsAcceptable() {
		t.Error("Report with 66% pass rate should not be acceptable (need 90%)")
	}

	// Test category reports
	if len(report.ByCategory) != 2 {
		t.Errorf("Report should have 2 category reports, got %d", len(report.ByCategory))
	}

	coreReport := report.ByCategory[CategoryCore]
	if coreReport == nil {
		t.Fatal("Core category report should exist")
	}
	if coreReport.Total != 2 || coreReport.Passed != 2 || coreReport.Failed != 0 {
		t.Errorf("Core report: total=%d passed=%d failed=%d, want total=2 passed=2 failed=0",
			coreReport.Total, coreReport.Passed, coreReport.Failed)
	}

	advancedReport := report.ByCategory[CategoryAdvanced]
	if advancedReport == nil {
		t.Fatal("Advanced category report should exist")
	}
	if advancedReport.Total != 1 || advancedReport.Passed != 0 || advancedReport.Failed != 1 {
		t.Errorf("Advanced report: total=%d passed=%d failed=%d, want total=1 passed=0 failed=1",
			advancedReport.Total, advancedReport.Passed, advancedReport.Failed)
	}

	// Test issues
	if len(report.Issues) != 1 {
		t.Errorf("Report should have 1 issue, got %d", len(report.Issues))
	} else {
		issue := report.Issues[0]
		if issue.FeatureID != "test.invalid1" {
			t.Errorf("Issue feature ID = %s, want test.invalid1", issue.FeatureID)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	r := GetDefaultRegistry()

	all := r.GetAll()
	if len(all) < 50 {
		t.Errorf("Default registry should have at least 50 features, got %d", len(all))
	}

	// Validate all features
	report := r.ValidateAll()

	t.Logf("Total features: %d", report.TotalFeatures)
	t.Logf("Passed: %d", report.PassedFeatures)
	t.Logf("Failed: %d", report.FailedFeatures)
	t.Logf("Pass rate: %.2f%%", report.PassRate*100)

	// All default features should pass
	if report.FailedFeatures > 0 {
		t.Errorf("Default registry has %d failed features:", report.FailedFeatures)
		for _, issue := range report.Issues {
			t.Errorf("  %s (%s): %v", issue.FeatureID, issue.FeatureName, issue.Issues)
		}
	}

	// Should meet acceptance criteria
	if !report.IsAcceptable() {
		t.Errorf("Default registry pass rate %.2f%% does not meet 90%% acceptance criteria", report.PassRate*100)
	}

	// Check category distribution
	for cat, catReport := range report.ByCategory {
		t.Logf("Category %s: %d features (%.1f%% pass rate)",
			cat, catReport.Total, catReport.PassRate()*100)
	}
}

func TestCategoryPassRate(t *testing.T) {
	tests := []struct {
		name     string
		report   *CategoryReport
		wantRate float64
	}{
		{
			name:     "all passed",
			report:   &CategoryReport{Total: 10, Passed: 10, Failed: 0},
			wantRate: 1.0,
		},
		{
			name:     "half passed",
			report:   &CategoryReport{Total: 10, Passed: 5, Failed: 5},
			wantRate: 0.5,
		},
		{
			name:     "none passed",
			report:   &CategoryReport{Total: 10, Passed: 0, Failed: 10},
			wantRate: 0.0,
		},
		{
			name:     "empty category",
			report:   &CategoryReport{Total: 0, Passed: 0, Failed: 0},
			wantRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.report.PassRate()
			if got != tt.wantRate {
				t.Errorf("PassRate() = %.2f, want %.2f", got, tt.wantRate)
			}
		})
	}
}

func BenchmarkFeatureValidation(b *testing.B) {
	f := &Feature{
		ID:                   "bench.feature",
		Name:                 "Benchmark Feature",
		Category:             CategoryCore,
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		HasTutorial:          true,
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"s1", "s2", "s3"},
		IntegrationCount:     3,
		Implemented:          true,
		Functional:           true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Validate()
	}
}

func BenchmarkRegistryValidateAll(b *testing.B) {
	r := GetDefaultRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ValidateAll()
	}
}

func BenchmarkRegistryGetFeature(b *testing.B) {
	r := GetDefaultRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.GetFeature("movement.8dir")
	}
}
