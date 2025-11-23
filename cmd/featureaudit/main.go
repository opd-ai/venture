// Feature Completeness Audit Tool
//
// This CLI tool validates all game features for Phase 65.1 of ROADMAP_V10.md.
//
// Usage:
//
//	go run ./cmd/featureaudit/                    # Full validation report
//	go run ./cmd/featureaudit/ -category core     # Validate specific category
//	go run ./cmd/featureaudit/ -feature movement.8dir  # Validate specific feature
//	go run ./cmd/featureaudit/ -summary           # Summary only
//	go run ./cmd/featureaudit/ -verbose           # Detailed output
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/opd-ai/venture/pkg/audit/features"
)

var (
	categoryFlag = flag.String("category", "", "Validate specific category (core, advanced, vehicles, social, housing, guilds, combat, economy, content, meta)")
	featureFlag  = flag.String("feature", "", "Validate specific feature by ID")
	summaryFlag  = flag.Bool("summary", false, "Show summary only")
	verboseFlag  = flag.Bool("verbose", false, "Show verbose output")
	failedFlag   = flag.Bool("failed", false, "Show only failed features")
)

func main() {
	flag.Parse()

	registry := features.GetDefaultRegistry()

	// Handle specific feature validation
	if *featureFlag != "" {
		validateFeature(registry, *featureFlag)
		return
	}

	// Handle specific category validation
	if *categoryFlag != "" {
		validateCategory(registry, *categoryFlag)
		return
	}

	// Full validation
	validateAll(registry)
}

func validateFeature(registry *features.FeatureRegistry, id string) {
	feature := registry.GetFeature(id)
	if feature == nil {
		fmt.Printf("ERROR: Feature '%s' not found\n", id)
		os.Exit(1)
	}

	fmt.Printf("=== Feature: %s ===\n", feature.Name)
	fmt.Printf("ID: %s\n", feature.ID)
	fmt.Printf("Category: %s\n", feature.Category)
	fmt.Printf("Description: %s\n", feature.Description)
	fmt.Println()

	fmt.Printf("Status:\n")
	fmt.Printf("  Implemented: %v\n", feature.Implemented)
	fmt.Printf("  Functional: %v\n", feature.Functional)
	fmt.Printf("  Tested: %v\n", feature.Tested)
	fmt.Println()

	fmt.Printf("Accessibility:\n")
	fmt.Printf("  Accessible: %v\n", feature.Accessible)
	fmt.Printf("  Time: %v\n", feature.AccessibilityTime)
	fmt.Printf("  Path: %s\n", feature.AccessibilityPath)
	fmt.Println()

	fmt.Printf("Tutorial:\n")
	fmt.Printf("  Has Tutorial: %v\n", feature.HasTutorial)
	fmt.Printf("  Completeness: %.1f%%\n", feature.TutorialCompleteness*100)
	fmt.Printf("  Location: %s\n", feature.TutorialLocation)
	fmt.Println()

	fmt.Printf("Integration:\n")
	fmt.Printf("  Count: %d\n", feature.IntegrationCount)
	fmt.Printf("  Systems: %v\n", feature.IntegratedSystems)
	fmt.Println()

	valid, issues := feature.Validate()
	if valid {
		fmt.Printf("✅ PASS: Feature meets all validation criteria\n")
	} else {
		fmt.Printf("❌ FAIL: Feature has issues:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}
}

func validateCategory(registry *features.FeatureRegistry, catName string) {
	cat := mapCategoryName(catName)
	categoryFeatures := registry.GetCategory(cat)
	if len(categoryFeatures) == 0 {
		fmt.Printf("No features in category '%s'\n", cat)
		return
	}

	fmt.Printf("=== Category: %s (%d features) ===\n\n", cat, len(categoryFeatures))
	passed, failed := validateFeatures(categoryFeatures)
	printCategorySummary(passed, failed, len(categoryFeatures))
}

// mapCategoryName maps a category name string to its corresponding FeatureCategory.
func mapCategoryName(catName string) features.FeatureCategory {
	catMap := map[string]features.FeatureCategory{
		"core":     features.CategoryCore,
		"advanced": features.CategoryAdvanced,
		"vehicles": features.CategoryVehicles,
		"social":   features.CategorySocial,
		"housing":  features.CategoryHousing,
		"guilds":   features.CategoryGuilds,
		"combat":   features.CategoryCombat,
		"economy":  features.CategoryEconomy,
		"content":  features.CategoryContent,
		"meta":     features.CategoryMeta,
	}

	cat, ok := catMap[strings.ToLower(catName)]
	if !ok {
		fmt.Printf("ERROR: Unknown category '%s'\n", catName)
		fmt.Printf("Valid categories: core, advanced, vehicles, social, housing, guilds, combat, economy, content, meta\n")
		os.Exit(1)
	}
	return cat
}

// validateFeatures validates a list of features and returns pass/fail counts.
func validateFeatures(categoryFeatures []*features.Feature) (passed, failed int) {
	for _, feature := range categoryFeatures {
		valid, issues := feature.Validate()

		if *failedFlag && valid {
			continue
		}

		if valid {
			passed++
			printPassedFeature(feature)
		} else {
			failed++
			printFailedFeature(feature, issues)
		}
	}
	return passed, failed
}

// printPassedFeature prints a passed feature with optional verbose details.
func printPassedFeature(feature *features.Feature) {
	if !*summaryFlag {
		fmt.Printf("✅ %s (%s)\n", feature.Name, feature.ID)
		if *verboseFlag {
			fmt.Printf("   Accessibility: %v, Tutorial: %.0f%%, Integration: %d systems\n",
				feature.AccessibilityTime, feature.TutorialCompleteness*100, feature.IntegrationCount)
		}
	}
}

// printFailedFeature prints a failed feature with its validation issues.
func printFailedFeature(feature *features.Feature, issues []string) {
	if !*summaryFlag {
		fmt.Printf("❌ %s (%s)\n", feature.Name, feature.ID)
		for _, issue := range issues {
			fmt.Printf("   - %s\n", issue)
		}
	}
}

// printCategorySummary prints the category validation summary with pass rate.
func printCategorySummary(passed, failed, total int) {
	fmt.Println()
	passRate := float64(passed) / float64(total) * 100
	fmt.Printf("Summary: %d/%d passed (%.1f%%)\n", passed, total, passRate)

	if passRate >= 90.0 {
		fmt.Printf("✅ Category meets 90%% acceptance criteria\n")
	} else {
		fmt.Printf("❌ Category does not meet 90%% acceptance criteria\n")
	}
}

func validateAll(registry *features.FeatureRegistry) {
	report := registry.ValidateAll()

	fmt.Printf("=== Feature Completeness Audit (Phase 65.1) ===\n\n")

	if !*summaryFlag && !*failedFlag {
		fmt.Printf("Total Features: %d\n", report.TotalFeatures)
		fmt.Printf("Passed: %d\n", report.PassedFeatures)
		fmt.Printf("Failed: %d\n", report.FailedFeatures)
		fmt.Printf("Pass Rate: %.2f%%\n\n", report.PassRate*100)
	}

	// Category breakdown
	if !*summaryFlag {
		fmt.Printf("--- Category Breakdown ---\n")

		// Sort categories for consistent output
		categories := make([]features.FeatureCategory, 0, len(report.ByCategory))
		for cat := range report.ByCategory {
			categories = append(categories, cat)
		}
		sort.Slice(categories, func(i, j int) bool {
			return string(categories[i]) < string(categories[j])
		})

		for _, cat := range categories {
			catReport := report.ByCategory[cat]
			passRate := catReport.PassRate() * 100
			status := "✅"
			if passRate < 90.0 {
				status = "❌"
			}
			fmt.Printf("%s %-20s: %2d/%2d (%.1f%%)\n",
				status, cat, catReport.Passed, catReport.Total, passRate)
		}
		fmt.Println()
	}

	// Failed features
	if len(report.Issues) > 0 && !*summaryFlag {
		fmt.Printf("--- Failed Features (%d) ---\n", len(report.Issues))
		for i, issue := range report.Issues {
			fmt.Printf("%d. %s (%s)\n", i+1, issue.FeatureName, issue.FeatureID)
			for _, issueDesc := range issue.Issues {
				fmt.Printf("   - %s\n", issueDesc)
			}
		}
		fmt.Println()
	}

	// Final verdict
	fmt.Printf("--- Acceptance Criteria ---\n")
	fmt.Printf("Target: 90%% features valid\n")
	fmt.Printf("Actual: %.2f%% features valid\n", report.PassRate*100)
	fmt.Println()

	if report.IsAcceptable() {
		fmt.Printf("✅ PASS: Feature completeness meets Phase 65.1 acceptance criteria\n")
		os.Exit(0)
	} else {
		fmt.Printf("❌ FAIL: Feature completeness does not meet Phase 65.1 acceptance criteria\n")
		fmt.Printf("\nTo pass, fix the %d failed features listed above.\n", report.FailedFeatures)
		os.Exit(1)
	}
}
