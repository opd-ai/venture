package balance

import (
	"context"
	"testing"
	"time"
)

// TestAllValidators uses table-driven pattern to test all 8 validators
func TestAllValidators(t *testing.T) {
	tests := []struct {
		name            string
		domain          string
		simulationCount int
		createValidator func(*BalanceConfig) BalanceValidator
		expectedMetrics []string
	}{
		{
			name:            "Combat",
			domain:          "Combat",
			simulationCount: 100,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewCombatValidator(c) },
			expectedMetrics: []string{"class_win_rate_variance", "enemy_scaling_r_squared", "boss_failure_rate"},
		},
		{
			name:            "Economic",
			domain:          "Economic",
			simulationCount: 100,
			createValidator: func(c *BalanceConfig) BalanceValidator { return NewEconomicValidator(c) },
			expectedMetrics: []string{"loot_value_correlation", "crafting_profit_margin", "gold_sink_ratio"},
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

			if result.SimulationCount != tt.simulationCount {
				t.Errorf("Expected %d simulations, got %d", tt.simulationCount, result.SimulationCount)
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
			for _, issue := range result.Issues {
				t.Logf("  ISSUE: %s", issue)
			}
		})
	}
}

func TestBalanceConfig(t *testing.T) {
	config := NewDefaultConfig()

	// Test threshold retrieval
	threshold := config.GetThreshold("class_win_rate_min")
	if threshold != 0.45 {
		t.Errorf("Expected threshold 0.45, got %f", threshold)
	}

	// Test simulation count retrieval
	count := config.GetSimulationCount("Combat")
	if count != 10000 {
		t.Errorf("Expected 10000 simulations, got %d", count)
	}

	// Test unknown domain defaults to 1000
	count = config.GetSimulationCount("Unknown")
	if count != 1000 {
		t.Errorf("Expected 1000 default simulations, got %d", count)
	}
}

func TestCombatClassBalance(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Combat"] = 600 // 100 per class pair
	validator := NewCombatValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Combat",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: 600,
	}

	err := validator.validateClassBalance(ctx, result)
	if err != nil {
		t.Fatalf("Class balance validation failed: %v", err)
	}

	// Verify class win rates exist
	classes := []string{"Warrior", "Rogue", "Mage", "Ranger", "Cleric", "Necromancer"}
	for _, class := range classes {
		metricKey := "class_win_rate_" + class
		if _, ok := result.Metrics[metricKey]; !ok {
			t.Errorf("Missing win rate for %s", class)
		}
	}

	// Log variance
	variance := result.Metrics["class_win_rate_variance"]
	t.Logf("Class win rate variance: %.4f", variance)

	// Variance should be reasonable (not everyone at 50%, but not wildly different)
	if variance > 0.20 {
		t.Errorf("Class variance too high: %.4f (suggests severe imbalance)", variance)
	}
}

func TestCombatWeaponBalance(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Combat"] = 600
	validator := NewCombatValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Combat",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: 600,
	}

	err := validator.validateWeaponBalance(ctx, result)
	if err != nil {
		t.Fatalf("Weapon balance validation failed: %v", err)
	}

	weapons := []string{"Sword", "Axe", "Bow", "Staff", "Dagger", "Spear"}
	for _, weapon := range weapons {
		metricKey := "weapon_usage_" + weapon
		if _, ok := result.Metrics[metricKey]; !ok {
			t.Errorf("Missing usage rate for %s", weapon)
		}
	}
}

func TestCombatEnemyScaling(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewCombatValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Combat",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateEnemyScaling(ctx, result)
	if err != nil {
		t.Fatalf("Enemy scaling validation failed: %v", err)
	}

	rSquared := result.Metrics["enemy_scaling_r_squared"]
	t.Logf("Enemy scaling R²: %.4f", rSquared)

	// R² should be high (enemy difficulty scales linearly with depth)
	if rSquared < 0.90 {
		t.Errorf("Enemy scaling R² too low: %.4f (target: ≥0.95)", rSquared)
	}
}

func TestCombatBossDifficulty(t *testing.T) {
	config := NewDefaultConfig()
	config.SimulationCounts["Combat"] = 500 // 100 boss battles
	validator := NewCombatValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:          "Combat",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		SimulationCount: 500,
	}

	err := validator.validateBossDifficulty(ctx, result)
	if err != nil {
		t.Fatalf("Boss difficulty validation failed: %v", err)
	}

	failureRate := result.Metrics["boss_failure_rate"]
	t.Logf("Boss failure rate: %.1f%%", failureRate*100)

	// Boss difficulty should be challenging but not impossible
	if failureRate < 0.40 || failureRate > 0.90 {
		t.Logf("WARNING: Boss failure rate outside typical range (%.1f%%, expected 60-80%%)", failureRate*100)
	}
}

func TestEconomicLootValue(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewEconomicValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Economic",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateLootValue(ctx, result)
	if err != nil {
		t.Fatalf("Loot value validation failed: %v", err)
	}

	correlation := result.Metrics["loot_value_correlation"]
	t.Logf("Loot-difficulty correlation: %.4f", correlation)

	// Correlation should be positive and strong
	if correlation < 0.50 {
		t.Logf("WARNING: Weak loot-difficulty correlation: %.4f (target: ≥0.80)", correlation)
	}
}

func TestEconomicCraftingProfit(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewEconomicValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Economic",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateCraftingProfit(ctx, result)
	if err != nil {
		t.Fatalf("Crafting profit validation failed: %v", err)
	}

	profit := result.Metrics["crafting_profit_margin"]
	t.Logf("Crafting profit margin: %.1f%%", profit*100)

	// Crafting should be profitable but not exploitable
	if profit < -0.10 || profit > 0.50 {
		t.Logf("WARNING: Crafting profit outside typical range (%.1f%%, expected 10-30%%)", profit*100)
	}
}

func TestEconomicGoldBalance(t *testing.T) {
	config := NewDefaultConfig()
	validator := NewEconomicValidator(config)

	ctx := context.Background()
	result := &ValidationResult{
		Domain:  "Economic",
		Passed:  true,
		Metrics: make(map[string]float64),
		Issues:  make([]string, 0),
	}

	err := validator.validateGoldBalance(ctx, result)
	if err != nil {
		t.Fatalf("Gold balance validation failed: %v", err)
	}

	ratio := result.Metrics["gold_sink_ratio"]
	t.Logf("Gold sink/source ratio: %.4f", ratio)

	// Ratio should be close to 1.0 (balanced economy)
	if ratio < 0.50 || ratio > 1.50 {
		t.Logf("WARNING: Gold sink/source ratio outside typical range (%.4f, expected 0.8-1.2)", ratio)
	}
}

// Benchmarks
func BenchmarkCombatValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Combat"] = 1000
	validator := NewCombatValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(ctx)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

func BenchmarkEconomicValidation(b *testing.B) {
	config := NewDefaultConfig()
	config.SimulationCounts["Economic"] = 500
	validator := NewEconomicValidator(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validator.Validate(ctx)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}
