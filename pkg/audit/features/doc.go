// Package features provides feature completeness validation for Phase 65.1 of ROADMAP_V10.md.
//
// # Overview
//
// The feature completeness validation ensures all documented game features meet three criteria:
//
// 1. **Accessibility**: Reachable within 30 minutes of gameplay
// 2. **Tutorial**: Explained in in-game help or tutorial (70%+ completeness)
// 3. **Integration**: Works with at least 2 other systems
//
// This package implements the acceptance criteria for Phase 65.1:
// - All 100+ features functional and accessible
// - No dead-end features (inaccessible due to bugs/design)
// - User testing: 90%+ feature discovery rate
//
// # Usage
//
// Create a registry and validate all features:
//
//	registry := features.GetDefaultRegistry()
//	report := registry.ValidateAll()
//
//	if report.IsAcceptable() {
//	    fmt.Printf("PASS: %.1f%% features valid (need 90%%)\n", report.PassRate*100)
//	} else {
//	    fmt.Printf("FAIL: %.1f%% features valid (need 90%%)\n", report.PassRate*100)
//	    for _, issue := range report.Issues {
//	        fmt.Printf("  %s: %v\n", issue.FeatureName, issue.Issues)
//	    }
//	}
//
// # Feature Categories
//
// Features are organized into 10 categories:
//
// - Core Gameplay: Movement, combat, inventory, progression
// - Advanced Systems: Vehicles, companions, classes, mini-games
// - Vehicles: Mount, combat, upgrades
// - Social: Chat, expressions, reputation
// - Housing: Claim, build, furniture, storage
// - Guilds: Create, join, resources, territory, warfare
// - Combat: Melee, ranged, magic, status effects
// - Economy: Trading, crafting, marketplace
// - Content: Quests, storytelling, lore
// - Meta-Game: Tutorial, settings, save/load, HUD, map
//
// # Feature Registration
//
// Features are registered in separate files by category:
//
// - core_features.go: Core gameplay (movement, combat, inventory, progression, skills)
// - advanced_features.go: Advanced systems (quests, crafting, vehicles, companions, classes)
// - social_housing_guilds.go: Social, housing, and guild features
// - meta_features.go: Meta-game and UI features
//
// Each feature specifies:
//
//	&Feature{
//	    ID:                   "movement.8dir",
//	    Name:                 "8-Direction Movement",
//	    Category:             CategoryCore,
//	    Accessible:           true,
//	    AccessibilityTime:    1 * time.Minute,
//	    AccessibilityPath:    "Tutorial: press WASD keys",
//	    HasTutorial:          true,
//	    TutorialLocation:     "First-time tutorial popup",
//	    TutorialCompleteness: 1.0,
//	    IntegratedSystems:    []string{"collision", "viewport", "animation"},
//	    IntegrationCount:     3,
//	    Implemented:          true,
//	    Functional:           true,
//	}
//
// # Validation Report
//
// The validation report provides detailed results:
//
//	type ValidationReport struct {
//	    TotalFeatures   int
//	    PassedFeatures  int
//	    FailedFeatures  int
//	    PassRate        float64
//	    ByCategory      map[FeatureCategory]*CategoryReport
//	    Issues          []FeatureIssue
//	}
//
// Category reports show per-category pass rates:
//
//	type CategoryReport struct {
//	    Category FeatureCategory
//	    Total    int
//	    Passed   int
//	    Failed   int
//	}
//
// # Acceptance Criteria
//
// For Phase 65.1 completion:
//
// - 90%+ features must pass validation (IsAcceptable() returns true)
// - All features must be implemented (Implemented=true)
// - All features must be functional (Functional=true)
// - All features must be accessible within 30 minutes (Accessible=true)
// - All features must have tutorial coverage ≥70% (TutorialCompleteness≥0.7)
// - All features must integrate with ≥2 systems (IntegrationCount≥2)
//
// # Testing
//
// Run tests to verify feature registry:
//
//	go test ./pkg/audit/features/
//
// Run CLI tool for interactive validation:
//
//	go run ./cmd/featureaudit/
package features
