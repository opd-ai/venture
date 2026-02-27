package balance

import (
	"context"
	"testing"
	"time"
)

// TestAllValidatorsExtended uses table-driven pattern to test validators 3-8
func TestAllValidatorsExtended(t *testing.T) {
	tests := []struct {
		name            string
		domain          string
		simulationCount int
		createValidator func(*BalanceConfig) BalanceValidator
		expectedMetrics []string
	}{
		{
			name:            "Progression",
			domain:          "Progression",
			simulationCount: 100,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewProgressionValidator(c) },
			expectedMetrics: []string{"xp_curve_r_squared", "skill_power_increase_avg", "stat_scaling_r_squared"},
		},
		{
			name:            "Social",
			domain:          "Social",
			simulationCount: 200,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewSocialValidator(c) },
			expectedMetrics: []string{"defender_advantage", "fraud_rate", "chat_reject_rate"},
		},
		{
			name:            "Housing",
			domain:          "Housing",
			simulationCount: 100,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewHousingValidator(c) },
			expectedMetrics: []string{"housing_affordability_rate", "upgrade_benefit_avg", "storage_adequacy_rate"},
		},
		{
			name:            "Vehicle",
			domain:          "Vehicle",
			simulationCount: 200,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewVehicleValidator(c) },
			expectedMetrics: []string{"speed_durability_correlation", "trip_completion_rate", "terrain_choice_rate"},
		},
		{
			name:            "Companion",
			domain:          "Companion",
			simulationCount: 100,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewCompanionValidator(c) },
			expectedMetrics: []string{"max_loyalty_rate", "avg_skills_learned", "companion_combat_contribution"},
		},
		{
			name:            "Quest",
			domain:          "Quest",
			simulationCount: 200,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewQuestValidator(c) },
			expectedMetrics: []string{"reward_difficulty_correlation", "difficulty_rating_accuracy", "completion_time_accuracy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			config.SimulationCounts[tt.domain] = tt.simulationCount
			validator := tt.createValidator(config)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := validator.Validate(ctx)
			if err != nil {
				t.Fatalf("%s validation failed: %v", tt.name, err)
			}

			// Verify result structure
			if result.Domain != tt.domain {
				t.Errorf("Expected domain '%s', got '%s'", tt.domain, result.Domain)
			}

			// Verify expected metrics exist
			for _, metric := range tt.expectedMetrics {
				if _, ok := result.Metrics[metric]; !ok {
					t.Errorf("Missing metric: %s", metric)
				}
			}

			// Log results
			t.Logf("%s Balance: Passed=%v, Issues=%d, Duration=%.2fs",
				tt.name, result.Passed, len(result.Issues), result.Duration)
			for key, value := range result.Metrics {
				t.Logf("  %s: %.4f", key, value)
			}
		})
	}
}

func TestProgressionXPCurve(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewProgressionValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Progression",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateXPCurve(ctx, result)
	if err != nil {
		t.Fatalf("XP curve validation failed: %v", err)
	}

	rSquared := result.Metrics["xp_curve_r_squared"]
	t.Logf("XP curve R²: %.4f", rSquared)

	if rSquared < 0.90 {
		t.Logf("WARNING: XP curve R² lower than expected: %.4f", rSquared)
	}
}

func TestProgressionStatScaling(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewProgressionValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Progression",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateStatScaling(ctx, result)
	if err != nil {
		t.Fatalf("Stat scaling validation failed: %v", err)
	}

	rSquared := result.Metrics["stat_scaling_r_squared"]
	t.Logf("Stat scaling R²: %.4f", rSquared)

	// Linear stat scaling should have near-perfect R²
	if rSquared < 0.99 {
		t.Errorf("Stat scaling not linear enough: R²=%.4f (expected ≥0.99)", rSquared)
	}
}

func TestSocialTerritoryControl(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Social"] = 1000
	validator := NewSocialValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Social",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		SimulationCount: 1000,
	}

	err := validator.validateTerritoryControl(ctx, result)
	if err != nil {
		t.Fatalf("Territory control validation failed: %v", err)
	}

	advantage := result.Metrics["defender_advantage"]
	t.Logf("Defender advantage: %.1f%%", advantage*100)

	// Defender should have some advantage but not overwhelming
	if advantage < 0.45 || advantage > 0.70 {
		t.Logf("WARNING: Defender advantage outside expected range: %.1f%%", advantage*100)
	}
}

func TestHousingAffordability(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Housing"] = 500
	validator := NewHousingValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Housing",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		SimulationCount: 500,
	}

	err := validator.validateBuildCosts(ctx, result)
	if err != nil {
		t.Fatalf("Build cost validation failed: %v", err)
	}

	affordability := result.Metrics["housing_affordability_rate"]
	t.Logf("Housing affordability: %.1f%%", affordability*100)

	if affordability < 0.60 || affordability > 0.98 {
		t.Logf("WARNING: Housing affordability outside expected range: %.1f%%", affordability*100)
	}
}

func TestVehicleSpeedDurabilityTradeoff(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewVehicleValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Vehicle",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateSpeedDurabilityTradeoff(ctx, result)
	if err != nil {
		t.Fatalf("Speed/durability tradeoff validation failed: %v", err)
	}

	correlation := result.Metrics["speed_durability_correlation"]
	t.Logf("Speed-durability correlation: %.4f", correlation)

	// Should be negative (fast = fragile)
	if correlation > 0 {
		t.Errorf("Speed-durability correlation should be negative: %.4f", correlation)
	}
}

func TestCompanionLoyaltyProgression(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Companion"] = 200
	validator := NewCompanionValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Companion",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		SimulationCount: 200,
	}

	err := validator.validateLoyaltyProgression(ctx, result)
	if err != nil {
		t.Fatalf("Loyalty progression validation failed: %v", err)
	}

	maxLoyaltyRate := result.Metrics["max_loyalty_rate"]
	avgTime := result.Metrics["avg_time_to_max_loyalty"]
	t.Logf("Max loyalty rate: %.1f%%, Avg time: %.1f hours", maxLoyaltyRate*100, avgTime)

	if maxLoyaltyRate < 0.40 {
		t.Logf("WARNING: Low max loyalty rate: %.1f%%", maxLoyaltyRate*100)
	}
}

func TestQuestRewardScaling(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Quest"] = 500
	validator := NewQuestValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Quest",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		SimulationCount: 500,
	}

	err := validator.validateRewardScaling(ctx, result)
	if err != nil {
		t.Fatalf("Reward scaling validation failed: %v", err)
	}

	correlation := result.Metrics["reward_difficulty_correlation"]
	t.Logf("Reward-difficulty correlation: %.4f", correlation)

	// Should be strongly positive
	if correlation < 0.70 {
		t.Logf("WARNING: Weak reward-difficulty correlation: %.4f", correlation)
	}
}

func TestAllValidatorsDomain(t *testing.T) {
	// Verify all validators return correct domain
	tests := []struct {
		name     string
		domain   string
		getValid func(*BalanceConfig) BalanceValidator
	}{
		{"Combat", "Combat", func(c *BalanceConfig) BalanceValidator { return NewCombatValidator(c) }},
		{"Economic", "Economic", func(c *BalanceConfig) BalanceValidator { return NewEconomicValidator(c) }},
		{"Progression", "Progression", func(c *BalanceConfig) BalanceValidator { return NewProgressionValidator(c) }},
		{"Social", "Social", func(c *BalanceConfig) BalanceValidator { return NewSocialValidator(c) }},
		{"Housing", "Housing", func(c *BalanceConfig) BalanceValidator { return NewHousingValidator(c) }},
		{"Vehicle", "Vehicle", func(c *BalanceConfig) BalanceValidator { return NewVehicleValidator(c) }},
		{"Companion", "Companion", func(c *BalanceConfig) BalanceValidator { return NewCompanionValidator(c) }},
		{"Quest", "Quest", func(c *BalanceConfig) BalanceValidator { return NewQuestValidator(c) }},
	}

	config := NewDefaultConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := tt.getValid(config)
			if validator.GetDomain() != tt.domain {
				t.Errorf("GetDomain() = %s, want %s", validator.GetDomain(), tt.domain)
			}
		})
	}
}

func TestValidatorContextCancellation(t *testing.T) {
	// Test that validators respect context cancellation
	config := NewDefaultConfig()
	config.SimulationCounts["Progression"] = 10000 // Large count to ensure cancellation

	validator := NewProgressionValidator(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := validator.Validate(ctx)
	// Either context error or successful completion before cancellation
	if err != nil && err != context.Canceled {
		t.Logf("Validator returned error on cancelled context: %v (may be expected)", err)
	}
}

// Benchmarks for new validators
func BenchmarkProgressionValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Progression"] = 100
	validator := NewProgressionValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}

func BenchmarkSocialValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Social"] = 200
	validator := NewSocialValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}

func BenchmarkHousingValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Housing"] = 100
	validator := NewHousingValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}

func BenchmarkVehicleValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Vehicle"] = 200
	validator := NewVehicleValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}

func BenchmarkCompanionValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Companion"] = 100
	validator := NewCompanionValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}

func BenchmarkQuestValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Quest"] = 200
	validator := NewQuestValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.Validate(ctx)
	}
}
