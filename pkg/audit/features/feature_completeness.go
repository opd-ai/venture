// Package features provides feature completeness validation for Phase 65.1.
//
// The feature completeness validation ensures all documented features are:
// 1. Accessible within 30 minutes of gameplay
// 2. Explained in in-game help or tutorial
// 3. Integrated with at least 2 other systems
//
// This package implements the acceptance criteria for Phase 65.1:
// - All 100+ features functional and accessible
// - No dead-end features (inaccessible due to bugs/design)
// - User testing: 90%+ feature discovery rate
package features

import (
	"fmt"
	"time"

	logrus "github.com/sirupsen/logrus"
)

// Validation thresholds for feature completeness criteria.
const (
	// TutorialCompletenessThreshold is the minimum tutorial coverage required
	// for a feature to pass validation (70%).
	TutorialCompletenessThreshold = 0.7

	// AcceptancePassRateThreshold is the minimum pass rate required for
	// the overall feature set to be acceptable per Phase 65.1 criteria (90%).
	AcceptancePassRateThreshold = 0.90
)

// Feature represents a single game feature
type Feature struct {
	ID          string
	Name        string
	Category    FeatureCategory
	Description string

	// Accessibility: reachable within 30 minutes of gameplay
	Accessible        bool
	AccessibilityTime time.Duration
	AccessibilityPath string // How to access

	// Tutorial: explained in in-game help or tutorial
	HasTutorial          bool
	TutorialLocation     string  // Where explained
	TutorialCompleteness float64 // 0.0-1.0

	// Integration: works with at least 2 other systems
	IntegratedSystems []string
	IntegrationCount  int

	// Status
	Implemented bool
	Tested      bool
	Functional  bool
}

// Validate checks if feature meets all three criteria
func (f *Feature) Validate() (bool, []string) {
	var issues []string

	if !f.Accessible {
		issues = append(issues, "not accessible within 30 minutes")
	}

	if !f.HasTutorial || f.TutorialCompleteness < TutorialCompletenessThreshold {
		issues = append(issues, fmt.Sprintf("tutorial incomplete (%.1f%%)", f.TutorialCompleteness*100))
	}

	if f.IntegrationCount < 2 {
		issues = append(issues, fmt.Sprintf("only %d integrations (need 2+)", f.IntegrationCount))
	}

	if !f.Implemented {
		issues = append(issues, "not implemented")
	}

	if !f.Functional {
		issues = append(issues, "not functional")
	}

	return len(issues) == 0, issues
}

// IsComplete returns true if feature passes all validation criteria
func (f *Feature) IsComplete() bool {
	valid, _ := f.Validate()
	return valid
}

// FeatureRegistry manages all game features
type FeatureRegistry struct {
	features   map[string]*Feature
	categories map[FeatureCategory][]*Feature
}

// NewFeatureRegistry creates a new feature registry
func NewFeatureRegistry() *FeatureRegistry {
	return &FeatureRegistry{
		features:   make(map[string]*Feature),
		categories: make(map[FeatureCategory][]*Feature),
	}
}

// Register adds a feature to the registry.
// Nil features are silently ignored to prevent panics.
func (r *FeatureRegistry) Register(f *Feature) {
	if f == nil {
		logrus.WithFields(logrus.Fields{
			"operation": "register_feature",
		}).Warn("attempted to register nil feature")
		return
	}
	r.features[f.ID] = f
	r.categories[f.Category] = append(r.categories[f.Category], f)
}

// GetFeature retrieves a feature by ID
func (r *FeatureRegistry) GetFeature(id string) *Feature {
	return r.features[id]
}

// GetCategory retrieves all features in a category
func (r *FeatureRegistry) GetCategory(cat FeatureCategory) []*Feature {
	return r.categories[cat]
}

// GetAll returns all features
func (r *FeatureRegistry) GetAll() []*Feature {
	features := make([]*Feature, 0, len(r.features))
	for _, f := range r.features {
		features = append(features, f)
	}
	return features
}

// ValidateAll checks all features and returns results
func (r *FeatureRegistry) ValidateAll() *ValidationReport {
	report := &ValidationReport{
		TotalFeatures: len(r.features),
		ByCategory:    make(map[FeatureCategory]*CategoryReport),
	}

	for _, feature := range r.features {
		valid, issues := feature.Validate()

		if valid {
			report.PassedFeatures++
		} else {
			report.FailedFeatures++
			report.Issues = append(report.Issues, FeatureIssue{
				FeatureID:   feature.ID,
				FeatureName: feature.Name,
				Issues:      issues,
			})
		}

		// Update category report
		catReport, ok := report.ByCategory[feature.Category]
		if !ok {
			catReport = &CategoryReport{Category: feature.Category}
			report.ByCategory[feature.Category] = catReport
		}
		catReport.Total++
		if valid {
			catReport.Passed++
		} else {
			catReport.Failed++
		}
	}

	// Calculate pass rate
	if report.TotalFeatures > 0 {
		report.PassRate = float64(report.PassedFeatures) / float64(report.TotalFeatures)
	}

	return report
}

// ValidationReport contains results of feature validation
type ValidationReport struct {
	TotalFeatures  int
	PassedFeatures int
	FailedFeatures int
	PassRate       float64
	ByCategory     map[FeatureCategory]*CategoryReport
	Issues         []FeatureIssue
}

// CategoryReport contains validation results for a category
type CategoryReport struct {
	Category FeatureCategory
	Total    int
	Passed   int
	Failed   int
}

// PassRate returns the pass rate for a category
func (cr *CategoryReport) PassRate() float64 {
	if cr.Total == 0 {
		return 0.0
	}
	return float64(cr.Passed) / float64(cr.Total)
}

// FeatureIssue represents validation issues for a feature
type FeatureIssue struct {
	FeatureID   string
	FeatureName string
	Issues      []string
}

// IsAcceptable returns true if pass rate meets acceptance criteria (90%)
func (r *ValidationReport) IsAcceptable() bool {
	return r.PassRate >= AcceptancePassRateThreshold
}
