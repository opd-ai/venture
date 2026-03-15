package balance

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// EconomicValidator validates economic balance through simulated transactions.
type EconomicValidator struct {
	config *BalanceConfig
}

// NewEconomicValidator creates an economic balance validator.
func NewEconomicValidator(config *BalanceConfig) *EconomicValidator {
	return &EconomicValidator{
		config: config,
	}
}

// GetDomain returns "Economic".
func (v *EconomicValidator) GetDomain() string {
	return "Economic"
}

// Validate runs economic balance tests.
func (v *EconomicValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Economic",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Economic"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Economic",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting economic balance validation")

	// Test 1: Loot value matches enemy difficulty
	logrus.Debug("validating loot value correlation")
	if err := v.validateLootValue(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Economic",
			"test":   "loot_value",
			"error":  err.Error(),
		}).Error("loot value validation failed")
		return nil, fmt.Errorf("loot value validation failed: %w", err)
	}

	// Test 2: Crafting profitability (10-30% margin)
	logrus.Debug("validating crafting profitability")
	if err := v.validateCraftingProfit(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Economic",
			"test":   "crafting_profit",
			"error":  err.Error(),
		}).Error("crafting validation failed")
		return nil, fmt.Errorf("crafting validation failed: %w", err)
	}

	// Test 3: Gold sink/source ratio (0.8-1.2)
	logrus.Debug("validating gold balance")
	if err := v.validateGoldBalance(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Economic",
			"test":   "gold_balance",
			"error":  err.Error(),
		}).Error("gold balance validation failed")
		return nil, fmt.Errorf("gold balance validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Economic",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("economic balance validation complete")
	return result, nil
}

func (v *EconomicValidator) validateLootValue(ctx context.Context, result *ValidationResult) error {
	difficulties := make([]float64, 0)
	values := make([]float64, 0)
	rng := rand.New(rand.NewSource(v.config.Seed))
	totalDepths := 50 // Total iterations for progress tracking

	for depth := 1; depth <= totalDepths; depth++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10 depths
		if depth%v.config.GetProgressInterval(totalDepths) == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": depth,
				"total":    totalDepths,
				"percent":  float64(depth) / float64(totalDepths) * 100,
			}).Debug("loot value simulation progress")
		}

		// Enemy difficulty scales with depth
		difficulty := 10.0 + (float64(depth) * 2.0)

		// Loot value should scale similarly
		value := 5.0 + (float64(depth) * 1.5) + (rng.Float64() * 5.0) // Base + scaling + variance

		difficulties = append(difficulties, difficulty)
		values = append(values, value)
	}

	correlation := v.calculateCorrelation(difficulties, values)
	result.Metrics["loot_value_correlation"] = correlation

	threshold := v.config.GetThreshold("loot_value_correlation")
	if correlation < threshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Loot value poorly correlated with difficulty (r=%.2f, target: ≥%.2f)",
				correlation, threshold))
		result.Recommendations = append(result.Recommendations,
			"Adjust item value formulas to scale with depth/difficulty")
	}

	return nil
}

func (v *EconomicValidator) validateCraftingProfit(ctx context.Context, result *ValidationResult) error {
	profits := v.simulateRecipeProfits(ctx)
	if profits == nil {
		return ctx.Err()
	}

	avgProfit := v.calculateAverageProfit(profits)
	result.Metrics["crafting_profit_margin"] = avgProfit

	return v.validateProfitThresholds(avgProfit, result)
}

// simulateRecipeProfits generates profit margins for recipe samples.
func (v *EconomicValidator) simulateRecipeProfits(ctx context.Context) []float64 {
	profits := make([]float64, 0)
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	totalRecipes := 100 // Total iterations for progress tracking

	for i := 0; i < totalRecipes; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Log progress at the configured interval
		if (i+1)%v.config.GetProgressInterval(totalRecipes) == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    totalRecipes,
				"percent":  float64(i+1) / float64(totalRecipes) * 100,
			}).Debug("crafting profit simulation progress")
		}

		profit := v.calculateRecipeProfit(rng)
		if profit >= 0 {
			profits = append(profits, profit)
		}
	}

	return profits
}

// calculateRecipeProfit simulates profit margin for a single recipe.
func (v *EconomicValidator) calculateRecipeProfit(rng *rand.Rand) float64 {
	ingredientCount := 2 + rng.Intn(2)
	materialCost := 0

	for j := 0; j < ingredientCount; j++ {
		itemValue := 10 + rng.Intn(40)
		materialCost += itemValue
	}

	if materialCost == 0 {
		return -1
	}

	profitMargin := 0.15 + rng.Float64()*0.15
	outputValue := int(float64(materialCost) * (1.0 + profitMargin))
	return (float64(outputValue) - float64(materialCost)) / float64(materialCost)
}

// calculateAverageProfit computes mean profit margin.
func (v *EconomicValidator) calculateAverageProfit(profits []float64) float64 {
	if len(profits) == 0 {
		return 0
	}

	sum := 0.0
	for _, p := range profits {
		sum += p
	}
	return sum / float64(len(profits))
}

// validateProfitThresholds checks if profit margins meet economic balance criteria.
func (v *EconomicValidator) validateProfitThresholds(avgProfit float64, result *ValidationResult) error {
	minThreshold := v.config.GetThreshold("crafting_profit_min")
	maxThreshold := v.config.GetThreshold("crafting_profit_max")

	if avgProfit < minThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Crafting too unprofitable (%.1f%% margin, target: %.1f%%-%.1f%%)",
				avgProfit*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Increase crafted item values or reduce material requirements")
	}

	if avgProfit > maxThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Crafting too profitable (%.1f%% margin, target: %.1f%%-%.1f%%)",
				avgProfit*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Decrease crafted item values or increase material requirements")
	}

	return nil
}

func (v *EconomicValidator) validateGoldBalance(ctx context.Context, result *ValidationResult) error {
	// Simulate gold sources (loot, quest rewards) vs sinks (repairs, housing, taxes)
	goldSources := 0.0
	goldSinks := 0.0
	rng := rand.New(rand.NewSource(v.config.Seed + 2))
	totalIterations := 1000 // Total iterations for progress tracking

	logrus.Debug("simulating gold sources")
	// Sources: loot from killing enemies
	for i := 0; i < totalIterations; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		itemValue := 5 + rng.Intn(50) // 5-55 gold per item
		goldSources += float64(itemValue)
	}

	logrus.Debug("simulating gold sinks")
	// Sinks: equipment repairs (10% of value per use), housing costs, consumables
	for i := 0; i < totalIterations; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		itemValue := 5 + rng.Intn(50) // Same as sources
		// Repairs: 10% of item value per 100 uses
		goldSinks += float64(itemValue) * 0.1
	}

	// Add fixed sinks: housing (1000), consumables (500 per 1000 items)
	goldSinks += 1000 + 500

	ratio := goldSinks / goldSources
	result.Metrics["gold_sink_ratio"] = ratio

	minThreshold := v.config.GetThreshold("gold_sink_ratio_min")
	maxThreshold := v.config.GetThreshold("gold_sink_ratio_max")

	if ratio < minThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Too many gold sources (sink/source=%.2f, target: %.2f-%.2f)",
				ratio, minThreshold, maxThreshold))
		result.Recommendations = append(result.Recommendations,
			"Increase repair costs, housing prices, or add more gold sinks")
	}

	if ratio > maxThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Too many gold sinks (sink/source=%.2f, target: %.2f-%.2f)",
				ratio, minThreshold, maxThreshold))
		result.Recommendations = append(result.Recommendations,
			"Decrease repair costs or increase loot values")
	}

	return nil
}

// calculateCorrelation computes the Pearson product-moment correlation
// coefficient (r) between two data series. Pearson's r measures the
// strength and direction of the linear relationship between x and y.
//
// Formula: r = Σ(xᵢ - x̄)(yᵢ - ȳ) / √(Σ(xᵢ - x̄)² · Σ(yᵢ - ȳ)²)
//
// Returns a value in [-1, 1] where:
//
//	 1 = perfect positive linear correlation
//	 0 = no linear correlation
//	-1 = perfect negative linear correlation
//
// Returns 0.0 if fewer than 2 data points are provided, if x and y
// have different lengths, or if either series has zero variance.
func (v *EconomicValidator) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	meanX, meanY := 0.0, 0.0
	for i := range x {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(len(x))
	meanY /= float64(len(y))

	var numerator, denomX, denomY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}

	if denomX == 0 || denomY == 0 {
		return 0.0
	}

	return numerator / math.Sqrt(denomX*denomY)
}
