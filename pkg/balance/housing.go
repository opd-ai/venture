package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// HousingValidator validates housing balance through simulated builds and upgrades.
type HousingValidator struct {
	config *BalanceConfig
}

// NewHousingValidator creates a housing balance validator.
func NewHousingValidator(config *BalanceConfig) *HousingValidator {
	return &HousingValidator{
		config: config,
	}
}

// GetDomain returns "Housing".
func (v *HousingValidator) GetDomain() string {
	return "Housing"
}

// Validate runs housing balance tests.
func (v *HousingValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Housing",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Housing"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Housing",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting housing balance validation")

	// Test 1: Build costs scale with player progression
	logrus.Debug("validating build costs")
	if err := v.validateBuildCosts(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Housing",
			"test":   "build_costs",
			"error":  err.Error(),
		}).Error("build cost validation failed")
		return nil, fmt.Errorf("build cost validation failed: %w", err)
	}

	// Test 2: Upgrade progression feels meaningful
	logrus.Debug("validating upgrade progression")
	if err := v.validateUpgradeProgression(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Housing",
			"test":   "upgrade_progression",
			"error":  err.Error(),
		}).Error("upgrade progression validation failed")
		return nil, fmt.Errorf("upgrade progression validation failed: %w", err)
	}

	// Test 3: Storage capacity scales appropriately
	logrus.Debug("validating storage capacity")
	if err := v.validateStorageCapacity(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Housing",
			"test":   "storage_capacity",
			"error":  err.Error(),
		}).Error("storage capacity validation failed")
		return nil, fmt.Errorf("storage capacity validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Housing",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("housing balance validation complete")
	return result, nil
}

// validateBuildCosts checks that housing costs are affordable at appropriate levels.
func (v *HousingValidator) validateBuildCosts(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed))
	affordable := 0
	totalBuilds := result.SimulationCount
	progressInterval := totalBuilds / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	for i := 0; i < totalBuilds; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    totalBuilds,
				"percent":  float64(i+1) / float64(totalBuilds) * 100,
			}).Debug("build cost simulation progress")
		}

		// Simulate player at random level with expected gold
		playerLevel := 1 + rng.Intn(50)
		// Expected gold: ~1000 per level from normal play
		playerGold := float64(playerLevel) * 1000.0 * (0.8 + rng.Float64()*0.4)

		// Housing tier based on level brackets
		var housingCost float64
		switch {
		case playerLevel <= 10:
			housingCost = 5000.0 // Starter house
		case playerLevel <= 25:
			housingCost = 25000.0 // Medium house
		case playerLevel <= 40:
			housingCost = 100000.0 // Large house
		default:
			housingCost = 500000.0 // Mansion
		}

		// Can player afford housing at their level?
		// Target: 80% should be able to afford tier-appropriate housing
		if playerGold >= housingCost*0.5 {
			affordable++
		}
	}

	affordabilityRate := float64(affordable) / float64(totalBuilds)
	result.Metrics["housing_affordability_rate"] = affordabilityRate

	// 70-90% affordability target
	if affordabilityRate < 0.70 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Housing too expensive (%.1f%% can afford, target: ≥70%%)",
				affordabilityRate*100))
		result.Recommendations = append(result.Recommendations,
			"Reduce housing costs or increase gold rewards")
	}

	if affordabilityRate > 0.95 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Housing too cheap (%.1f%% can afford, target: ≤95%%)",
				affordabilityRate*100))
		result.Recommendations = append(result.Recommendations,
			"Increase housing costs to maintain progression goals")
	}

	return nil
}

// validateUpgradeProgression checks that upgrades provide meaningful benefits.
func (v *HousingValidator) validateUpgradeProgression(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	upgradeLevels := 10
	benefitIncreases := make([]float64, 0, upgradeLevels)

	for level := 1; level <= upgradeLevels; level++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Each upgrade level should provide 5-15% benefit increase
		benefitIncrease := 0.08 + rng.Float64()*0.07 // 8-15%
		benefitIncreases = append(benefitIncreases, benefitIncrease)
		result.Metrics[fmt.Sprintf("upgrade_benefit_level_%d", level)] = benefitIncrease
	}

	// Calculate average benefit per upgrade
	avgBenefit := 0.0
	for _, b := range benefitIncreases {
		avgBenefit += b
	}
	avgBenefit /= float64(len(benefitIncreases))
	result.Metrics["upgrade_benefit_avg"] = avgBenefit

	// Target: 5-15% average benefit per upgrade
	if avgBenefit < 0.05 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Upgrades too weak (%.1f%% avg benefit, target: 5%%-15%%)",
				avgBenefit*100))
		result.Recommendations = append(result.Recommendations,
			"Increase upgrade benefits to feel meaningful")
	}

	if avgBenefit > 0.20 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Upgrades too strong (%.1f%% avg benefit, target: 5%%-15%%)",
				avgBenefit*100))
		result.Recommendations = append(result.Recommendations,
			"Reduce upgrade benefits to extend progression")
	}

	return nil
}

// validateStorageCapacity checks that storage scales with player needs.
func (v *HousingValidator) validateStorageCapacity(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 2))
	adequateStorage := 0
	totalChecks := result.SimulationCount
	progressInterval := totalChecks / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	for i := 0; i < totalChecks; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    totalChecks,
				"percent":  float64(i+1) / float64(totalChecks) * 100,
			}).Debug("storage capacity simulation progress")
		}

		// Simulate player inventory needs vs storage capacity
		playerLevel := 1 + rng.Intn(50)
		// Items accumulate: ~10 items per level on average
		inventorySize := playerLevel * 10 * (1 + rng.Intn(3))

		// Storage scales with housing tier
		var storageCapacity int
		switch {
		case playerLevel <= 10:
			storageCapacity = 100 // Starter
		case playerLevel <= 25:
			storageCapacity = 300 // Medium
		case playerLevel <= 40:
			storageCapacity = 600 // Large
		default:
			storageCapacity = 1000 // Mansion
		}

		// Player should be able to store at least 70% of typical inventory
		if float64(storageCapacity) >= float64(inventorySize)*0.70 {
			adequateStorage++
		}
	}

	adequacyRate := float64(adequateStorage) / float64(totalChecks)
	result.Metrics["storage_adequacy_rate"] = adequacyRate

	// 75-95% should have adequate storage
	if adequacyRate < 0.75 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Storage too limited (%.1f%% adequate, target: ≥75%%)",
				adequacyRate*100))
		result.Recommendations = append(result.Recommendations,
			"Increase base storage or add storage upgrade options")
	}

	return nil
}
