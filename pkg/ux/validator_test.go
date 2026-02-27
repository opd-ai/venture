package ux

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewJourneyValidator(t *testing.T) {
	v := NewJourneyValidator()
	if v == nil {
		t.Fatal("NewJourneyValidator returned nil")
	}
	if v.config.MinCompletionRate != 0.90 {
		t.Errorf("Expected MinCompletionRate 0.90, got %f", v.config.MinCompletionRate)
	}
}

func TestValidateJourney_NewPlayer(t *testing.T) {
	config := ValidationConfig{
		Runs:                 5,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}
	v := NewJourneyValidatorWithConfig(config)

	result := v.ValidateJourney(JourneyNewPlayer)

	if result.Type != JourneyNewPlayer {
		t.Errorf("Expected journey type %s, got %s", JourneyNewPlayer, result.Type)
	}

	if result.CompletionRate < 0.0 || result.CompletionRate > 1.0 {
		t.Errorf("Invalid completion rate: %f", result.CompletionRate)
	}

	if result.ErrorRate < 0.0 || result.ErrorRate > 1.0 {
		t.Errorf("Invalid error rate: %f", result.ErrorRate)
	}

	if result.Satisfaction < 0.0 || result.Satisfaction > 1.0 {
		t.Errorf("Invalid satisfaction: %f", result.Satisfaction)
	}

	t.Logf("New Player Journey: Completion=%.1f%%, Satisfaction=%.1f%%, Errors=%.1f%%, Passed=%v",
		result.CompletionRate*100, result.Satisfaction*100, result.ErrorRate*100, result.Passed)
}

func TestValidateJourney_Crafter(t *testing.T) {
	config := ValidationConfig{
		Runs:                 5,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}
	v := NewJourneyValidatorWithConfig(config)

	result := v.ValidateJourney(JourneyCrafter)

	if !result.Passed {
		t.Errorf("Crafter journey should pass, got Passed=%v, Error=%v", result.Passed, result.Error)
	}

	if len(result.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(result.Steps))
	}

	t.Logf("Crafter Journey: Completion=%.1f%%, Duration=%v",
		result.CompletionRate*100, result.AverageDuration)
}

func TestValidateAll(t *testing.T) {
	config := ValidationConfig{
		Runs:                 3,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}
	v := NewJourneyValidatorWithConfig(config)

	results := v.ValidateAll()

	if len(results) != 20 {
		t.Errorf("Expected 20 journey results, got %d", len(results))
	}

	passCount := 0
	for _, result := range results {
		if result.Passed {
			passCount++
		}
		t.Logf("Journey %s: Passed=%v, Completion=%.1f%%, Satisfaction=%.1f%%",
			result.Name, result.Passed, result.CompletionRate*100, result.Satisfaction*100)
	}

	if passCount < 15 {
		t.Errorf("Expected at least 15/20 journeys to pass, got %d", passCount)
	}
}

func TestGetSummary(t *testing.T) {
	validator := NewJourneyValidator()

	tests := []struct {
		name                string
		results             []JourneyResult
		wantTotal           int
		wantPassed          int
		wantPassRate        float64
		wantAvgCompletion   float64
		wantAvgSatisfaction float64
		wantAvgErrorRate    float64
	}{
		{
			name: "mixed results",
			results: []JourneyResult{
				{Passed: true, CompletionRate: 1.0, Satisfaction: 0.95, ErrorRate: 0.0},
				{Passed: true, CompletionRate: 0.95, Satisfaction: 0.90, ErrorRate: 0.02},
				{Passed: false, CompletionRate: 0.80, Satisfaction: 0.75, ErrorRate: 0.10},
			},
			wantTotal:           3,
			wantPassed:          2,
			wantPassRate:        2.0 / 3.0,
			wantAvgCompletion:   (1.0 + 0.95 + 0.80) / 3.0,
			wantAvgSatisfaction: (0.95 + 0.90 + 0.75) / 3.0,
			wantAvgErrorRate:    (0.0 + 0.02 + 0.10) / 3.0,
		},
		{
			name:                "empty results",
			results:             []JourneyResult{},
			wantTotal:           0,
			wantPassed:          0,
			wantPassRate:        0.0,
			wantAvgCompletion:   0.0,
			wantAvgSatisfaction: 0.0,
			wantAvgErrorRate:    0.0,
		},
		{
			name: "all passed",
			results: []JourneyResult{
				{Passed: true, CompletionRate: 1.0, Satisfaction: 0.95, ErrorRate: 0.0},
				{Passed: true, CompletionRate: 1.0, Satisfaction: 0.98, ErrorRate: 0.0},
			},
			wantTotal:           2,
			wantPassed:          2,
			wantPassRate:        1.0,
			wantAvgCompletion:   1.0,
			wantAvgSatisfaction: (0.95 + 0.98) / 2.0,
			wantAvgErrorRate:    0.0,
		},
		{
			name: "all failed",
			results: []JourneyResult{
				{Passed: false, CompletionRate: 0.5, Satisfaction: 0.4, ErrorRate: 0.3},
				{Passed: false, CompletionRate: 0.6, Satisfaction: 0.5, ErrorRate: 0.2},
			},
			wantTotal:           2,
			wantPassed:          0,
			wantPassRate:        0.0,
			wantAvgCompletion:   0.55,
			wantAvgSatisfaction: 0.45,
			wantAvgErrorRate:    0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := validator.GetSummary(tt.results)

			if summary.TotalJourneys != tt.wantTotal {
				t.Errorf("TotalJourneys = %d, want %d", summary.TotalJourneys, tt.wantTotal)
			}
			if summary.PassedJourneys != tt.wantPassed {
				t.Errorf("PassedJourneys = %d, want %d", summary.PassedJourneys, tt.wantPassed)
			}
			if summary.PassRate != tt.wantPassRate {
				t.Errorf("PassRate = %.4f, want %.4f", summary.PassRate, tt.wantPassRate)
			}
			if summary.AverageCompletionRate != tt.wantAvgCompletion {
				t.Errorf("AverageCompletionRate = %.4f, want %.4f", summary.AverageCompletionRate, tt.wantAvgCompletion)
			}
			if summary.AverageSatisfaction != tt.wantAvgSatisfaction {
				t.Errorf("AverageSatisfaction = %.4f, want %.4f", summary.AverageSatisfaction, tt.wantAvgSatisfaction)
			}
			if summary.AverageErrorRate != tt.wantAvgErrorRate {
				t.Errorf("AverageErrorRate = %.4f, want %.4f", summary.AverageErrorRate, tt.wantAvgErrorRate)
			}
		})
	}

	// Legacy compatibility test - ensure basic behavior still works
	results := []JourneyResult{
		{Passed: true, CompletionRate: 1.0, Satisfaction: 0.95, ErrorRate: 0.0},
		{Passed: true, CompletionRate: 0.95, Satisfaction: 0.90, ErrorRate: 0.02},
		{Passed: false, CompletionRate: 0.80, Satisfaction: 0.75, ErrorRate: 0.10},
	}
	summary := validator.GetSummary(results)
	t.Logf("Summary: Pass rate=%.1f%%, Avg completion=%.1f%%, Avg satisfaction=%.1f%%",
		summary.PassRate*100, summary.AverageCompletionRate*100, summary.AverageSatisfaction*100)
}

func TestJourneyContext(t *testing.T) {
	ctx := &JourneyContext{
		PlayerID:  12345,
		WorldSeed: 67890,
		Data:      make(map[string]interface{}),
	}

	err := createCharacter(ctx)
	if err != nil {
		t.Fatalf("createCharacter failed: %v", err)
	}

	if !ctx.Data["character_created"].(bool) {
		t.Error("Expected character_created to be true")
	}

	err = completeTutorial(ctx)
	if err != nil {
		t.Fatalf("completeTutorial failed: %v", err)
	}

	if !ctx.Data["tutorial_complete"].(bool) {
		t.Error("Expected tutorial_complete to be true")
	}
}

func TestStepDependencies(t *testing.T) {
	ctx := &JourneyContext{
		PlayerID:  1,
		WorldSeed: 1,
		Data:      make(map[string]interface{}),
	}

	// Try to complete tutorial without creating character
	err := completeTutorial(ctx)
	if err == nil {
		t.Error("Expected error when completing tutorial without character creation")
	}

	// Create character first
	createCharacter(ctx)
	err = completeTutorial(ctx)
	if err != nil {
		t.Errorf("Expected success after character creation, got: %v", err)
	}
}

func TestCraftingWorkflow(t *testing.T) {
	ctx := &JourneyContext{
		PlayerID:  1,
		WorldSeed: 1,
		Data:      make(map[string]interface{}),
	}

	// Full crafting workflow
	if err := gatherMaterials(ctx); err != nil {
		t.Fatalf("gatherMaterials failed: %v", err)
	}

	if err := findRecipe(ctx); err != nil {
		t.Fatalf("findRecipe failed: %v", err)
	}

	if err := accessCraftingStation(ctx); err != nil {
		t.Fatalf("accessCraftingStation failed: %v", err)
	}

	if err := craftItem(ctx); err != nil {
		t.Fatalf("craftItem failed: %v", err)
	}

	if err := equipItem(ctx); err != nil {
		t.Fatalf("equipItem failed: %v", err)
	}

	if !ctx.Data["item_equipped"].(bool) {
		t.Error("Expected item to be equipped")
	}
}

func TestInsufficientMaterials(t *testing.T) {
	ctx := &JourneyContext{
		PlayerID:  1,
		WorldSeed: 1,
		Data:      make(map[string]interface{}),
	}

	ctx.Data["materials"] = 3 // Less than required 5

	err := craftItem(ctx)
	if err == nil {
		t.Error("Expected error when crafting with insufficient materials")
	}
}

func TestAllJourneyDefinitions(t *testing.T) {
	journeys := AllJourneys()

	if len(journeys) != 20 {
		t.Errorf("Expected 20 journey definitions, got %d", len(journeys))
	}

	uniqueTypes := make(map[JourneyType]bool)
	for _, journey := range journeys {
		if journey.Type == "" {
			t.Errorf("Journey %s has empty type", journey.Name)
		}
		if journey.Name == "" {
			t.Errorf("Journey type %s has empty name", journey.Type)
		}
		if len(journey.Steps) == 0 {
			t.Errorf("Journey %s has no steps", journey.Name)
		}
		if journey.ExpectedDuration == 0 {
			t.Errorf("Journey %s has zero expected duration", journey.Name)
		}

		uniqueTypes[journey.Type] = true
	}

	if len(uniqueTypes) != 20 {
		t.Errorf("Expected 20 unique journey types, got %d", len(uniqueTypes))
	}
}

func TestDurationTolerance(t *testing.T) {
	v := NewJourneyValidator()
	v.config.TimeTolerancePercent = 20.0

	tests := []struct {
		name     string
		actual   time.Duration
		expected time.Duration
		want     bool
	}{
		{"exact match", 30 * time.Minute, 30 * time.Minute, true},
		{"within tolerance low", 25 * time.Minute, 30 * time.Minute, true},
		{"within tolerance high", 35 * time.Minute, 30 * time.Minute, true},
		{"outside tolerance low", 20 * time.Minute, 30 * time.Minute, false},
		{"outside tolerance high", 40 * time.Minute, 30 * time.Minute, false},
		{"zero expected", 10 * time.Minute, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.checkDurationWithinTolerance(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("checkDurationWithinTolerance(%v, %v) = %v, want %v",
					tt.actual, tt.expected, got, tt.want)
			}
		})
	}
}

func TestSatisfactionCalculation(t *testing.T) {
	v := NewJourneyValidator()
	v.config.TimeTolerancePercent = 20.0

	tests := []struct {
		name             string
		completionRate   float64
		actualDuration   time.Duration
		expectedDuration time.Duration
		minSatisfaction  float64
	}{
		{"perfect completion on time", 1.0, 30 * time.Minute, 30 * time.Minute, 1.0},
		{"perfect completion fast", 1.0, 20 * time.Minute, 30 * time.Minute, 1.0},
		{"perfect completion slow", 1.0, 40 * time.Minute, 30 * time.Minute, 0.70},
		{"partial completion on time", 0.85, 30 * time.Minute, 30 * time.Minute, 0.80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			satisfaction := v.calculateSatisfaction(tt.completionRate, tt.actualDuration, tt.expectedDuration)
			if satisfaction < tt.minSatisfaction {
				t.Errorf("Satisfaction %.2f below minimum %.2f", satisfaction, tt.minSatisfaction)
			}
			if satisfaction < 0.0 || satisfaction > 1.0 {
				t.Errorf("Satisfaction %.2f out of valid range [0.0, 1.0]", satisfaction)
			}
		})
	}
}

func BenchmarkValidateJourney(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateJourney(JourneyNewPlayer)
	}
}

func BenchmarkValidateAll(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateAll()
	}
}

func BenchmarkStepExecution(b *testing.B) {
	ctx := &JourneyContext{
		PlayerID:  1,
		WorldSeed: 1,
		Data:      make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createCharacter(ctx)
	}
}

func TestDeterministicRandom(t *testing.T) {
	// Validate that same seed produces same results
	seed := int64(12345)

	v1 := &JourneyValidator{
		config: DefaultValidationConfig(),
		rng:    rand.New(rand.NewSource(seed)),
	}

	v2 := &JourneyValidator{
		config: DefaultValidationConfig(),
		rng:    rand.New(rand.NewSource(seed)),
	}

	// Both validators should generate same player IDs
	id1 := v1.rng.Intn(10000)
	id2 := v2.rng.Intn(10000)

	if id1 != id2 {
		t.Errorf("Expected deterministic RNG: got %d and %d", id1, id2)
	}
}

func TestValidationConfig(t *testing.T) {
	config := DefaultValidationConfig()

	if config.Runs != 10 {
		t.Errorf("Expected 10 runs, got %d", config.Runs)
	}

	if config.MinCompletionRate != 0.90 {
		t.Errorf("Expected 0.90 min completion rate, got %f", config.MinCompletionRate)
	}

	if config.MinSatisfaction != 0.80 {
		t.Errorf("Expected 0.80 min satisfaction, got %f", config.MinSatisfaction)
	}

	if config.MaxErrorRate != 0.05 {
		t.Errorf("Expected 0.05 max error rate, got %f", config.MaxErrorRate)
	}

	if config.TimeTolerancePercent != 20.0 {
		t.Errorf("Expected 20.0%% time tolerance, got %f", config.TimeTolerancePercent)
	}
}

// BenchmarkNewPlayerJourney benchmarks the complete new player journey workflow.
func BenchmarkNewPlayerJourney(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateJourney(JourneyNewPlayer)
	}
}

// BenchmarkCrafterJourney benchmarks the crafter journey workflow.
func BenchmarkCrafterJourney(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateJourney(JourneyCrafter)
	}
}

// BenchmarkPvPJourney benchmarks the PvP journey workflow.
func BenchmarkPvPJourney(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidateJourney(JourneyPvPer)
	}
}

// BenchmarkFullValidation benchmarks validation of all 20 journeys.
func BenchmarkFullValidation(b *testing.B) {
	v := NewJourneyValidator()
	v.config.Runs = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := v.ValidateAll()
		_ = v.GetSummary(results)
	}
}

// BenchmarkJourneyStep benchmarks individual journey step execution.
func BenchmarkJourneyStep(b *testing.B) {
	steps := map[string]func(*JourneyContext) error{
		"createCharacter":       createCharacter,
		"completeTutorial":      completeTutorial,
		"gatherMaterials":       gatherMaterials,
		"findRecipe":            findRecipe,
		"accessCraftingStation": accessCraftingStation,
		"craftItem":             craftItem,
		"equipItem":             equipItem,
		"acceptQuest":           acceptQuest,
		"completeObjectives":    completeObjectives,
		"returnToNPC":           returnToNPC,
	}

	ctx := &JourneyContext{
		PlayerID:  1,
		WorldSeed: 1,
		Data:      make(map[string]interface{}),
	}

	for name, step := range steps {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Reset context for clean execution
				ctx.Data = make(map[string]interface{})
				// Set common dependencies
				ctx.Data["character_created"] = true
				ctx.Data["tutorial_complete"] = true
				ctx.Data["materials"] = 10
				ctx.Data["recipe_known"] = true
				ctx.Data["crafting_station_found"] = true
				ctx.Data["item_crafted"] = true

				step(ctx)
			}
		})
	}
}
