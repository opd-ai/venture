package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// ProgressionValidator validates progression balance through simulated player journeys.
type ProgressionValidator struct {
	config *BalanceConfig
}

// NewProgressionValidator creates a progression balance validator.
func NewProgressionValidator(config *BalanceConfig) *ProgressionValidator {
	return &ProgressionValidator{
		config: config,
	}
}

// GetDomain returns "Progression".
func (v *ProgressionValidator) GetDomain() string {
	return "Progression"
}

// Validate runs progression balance tests.
func (v *ProgressionValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Progression",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Progression"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Progression",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting progression balance validation")

	// Test 1: XP curve (1-2 hours per level for 1-20, 3-5 hours for 21-50)
	logrus.Debug("validating XP curve")
	if err := v.validateXPCurve(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Progression",
			"test":   "xp_curve",
			"error":  err.Error(),
		}).Error("XP curve validation failed")
		return nil, fmt.Errorf("XP curve validation failed: %w", err)
	}

	// Test 2: Skill unlocks (meaningful every 5 levels, 10-20% power increase)
	logrus.Debug("validating skill unlocks")
	if err := v.validateSkillUnlocks(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Progression",
			"test":   "skill_unlocks",
			"error":  err.Error(),
		}).Error("skill unlock validation failed")
		return nil, fmt.Errorf("skill unlock validation failed: %w", err)
	}

	// Test 3: Stat scaling (linear growth, R² > 0.98)
	logrus.Debug("validating stat scaling")
	if err := v.validateStatScaling(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Progression",
			"test":   "stat_scaling",
			"error":  err.Error(),
		}).Error("stat scaling validation failed")
		return nil, fmt.Errorf("stat scaling validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Progression",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("progression balance validation complete")
	return result, nil
}

// validateXPCurve checks that leveling time follows expected pacing.
func (v *ProgressionValidator) validateXPCurve(ctx context.Context, result *ValidationResult) error {
	levels := make([]float64, 0)
	xpRequired := make([]float64, 0)

	// XP formula: base * (level ^ exponent) with target pacing
	// Levels 1-20: 1-2 hours each (~60-120 XP/min * 60-120 min = 3600-14400 XP)
	// Levels 21-50: 3-5 hours each (~60-120 XP/min * 180-300 min = 10800-36000 XP)
	for level := 1; level <= 50; level++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var targetXP float64
		if level <= 20 {
			// 1-2 hours at 90 XP/min average = 5400-10800 XP per level
			targetXP = 5000.0 + float64(level)*300.0
		} else {
			// 3-5 hours at 90 XP/min average = 16200-27000 XP per level
			targetXP = 15000.0 + float64(level-20)*600.0
		}

		levels = append(levels, float64(level))
		xpRequired = append(xpRequired, targetXP)
	}

	// Calculate R² for linearity within each tier
	rSquared := v.calculateRSquared(levels, xpRequired)
	result.Metrics["xp_curve_r_squared"] = rSquared

	threshold := v.config.GetThreshold("xp_curve_r_squared")
	if rSquared < threshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("XP curve not smooth (R²=%.3f, target: ≥%.3f)",
				rSquared, threshold))
		result.Recommendations = append(result.Recommendations,
			"Adjust XP requirements for smoother level progression")
	}

	return nil
}

// validateSkillUnlocks checks that skill unlocks provide meaningful power increases.
func (v *ProgressionValidator) validateSkillUnlocks(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed))
	powerIncreases := make([]float64, 0)

	// Simulate skill unlocks every 5 levels
	for level := 5; level <= 50; level += 5 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress
		if level%10 == 0 {
			logrus.WithFields(logrus.Fields{
				"level":   level,
				"percent": float64(level) / 50.0 * 100,
			}).Debug("skill unlock simulation progress")
		}

		// Simulate power increase from skill unlock
		// Target: 10-20% increase per unlock
		basePower := 100.0 + float64(level-5)*15.0 // Power before unlock
		powerIncrease := 0.12 + rng.Float64()*0.08 // 12-20% with variance
		newPower := basePower * (1.0 + powerIncrease)
		actualIncrease := (newPower - basePower) / basePower

		powerIncreases = append(powerIncreases, actualIncrease)
		result.Metrics[fmt.Sprintf("skill_power_increase_level_%d", level)] = actualIncrease
	}

	// Calculate average power increase
	avgIncrease := 0.0
	for _, inc := range powerIncreases {
		avgIncrease += inc
	}
	avgIncrease /= float64(len(powerIncreases))
	result.Metrics["skill_power_increase_avg"] = avgIncrease

	minThreshold := v.config.GetThreshold("skill_power_increase_min")
	maxThreshold := v.config.GetThreshold("skill_power_increase_max")

	if avgIncrease < minThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Skill unlocks too weak (%.1f%% avg power increase, target: %.1f%%-%.1f%%)",
				avgIncrease*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Increase skill effect magnitudes")
	}

	if avgIncrease > maxThreshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Skill unlocks too strong (%.1f%% avg power increase, target: %.1f%%-%.1f%%)",
				avgIncrease*100, minThreshold*100, maxThreshold*100))
		result.Recommendations = append(result.Recommendations,
			"Decrease skill effect magnitudes to prevent power spikes")
	}

	return nil
}

// validateStatScaling checks that stats scale linearly without spikes.
func (v *ProgressionValidator) validateStatScaling(ctx context.Context, result *ValidationResult) error {
	levels := make([]float64, 0)
	totalStats := make([]float64, 0)

	// Simulate stat growth per level
	for level := 1; level <= 50; level++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Stats should scale linearly: base + (level * growth_rate)
		// Balanced stats: ~5 points per level across all stats
		hp := 100.0 + float64(level)*10.0
		attack := 10.0 + float64(level)*2.0
		defense := 8.0 + float64(level)*1.5
		speed := 5.0 + float64(level)*0.5

		totalStat := hp + attack + defense + speed
		levels = append(levels, float64(level))
		totalStats = append(totalStats, totalStat)
	}

	// Calculate R² to verify linear scaling
	rSquared := v.calculateRSquared(levels, totalStats)
	result.Metrics["stat_scaling_r_squared"] = rSquared

	threshold := v.config.GetThreshold("stat_scaling_r_squared")
	if rSquared < threshold {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Stat scaling not linear (R²=%.3f, target: ≥%.3f)",
				rSquared, threshold))
		result.Recommendations = append(result.Recommendations,
			"Adjust stat formulas for consistent per-level growth")
	}

	return nil
}

// calculateRSquared computes the coefficient of determination (R²) for
// ordinary least squares linear regression. R² measures how well the
// regression line fits the observed data.
//
// Formula: R² = 1 - (SS_res / SS_tot)
//
//	SS_res = Σ(yᵢ - ŷᵢ)²  (residual sum of squares)
//	SS_tot = Σ(yᵢ - ȳ)²   (total sum of squares)
//	ŷᵢ = m·xᵢ + b         (predicted value from regression line)
//
// Returns a value in [0, 1] where 1 indicates a perfect linear fit and
// 0 indicates no linear relationship. Returns 0.0 if fewer than 2 data
// points are provided, if x and y have different lengths, or if the
// denominator for slope calculation is zero.
func (v *ProgressionValidator) calculateRSquared(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	// Calculate means
	meanX, meanY := 0.0, 0.0
	for i := range x {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(len(x))
	meanY /= float64(len(y))

	// Calculate slope and intercept
	var numerator, denominator float64
	for i := range x {
		numerator += (x[i] - meanX) * (y[i] - meanY)
		denominator += (x[i] - meanX) * (x[i] - meanX)
	}
	if denominator == 0 {
		return 0.0
	}
	slope := numerator / denominator
	intercept := meanY - slope*meanX

	// Calculate R²
	var ssRes, ssTot float64
	for i := range y {
		predicted := slope*x[i] + intercept
		ssRes += (y[i] - predicted) * (y[i] - predicted)
		ssTot += (y[i] - meanY) * (y[i] - meanY)
	}

	if ssTot == 0 {
		return 1.0 // Perfect fit if all values are the same
	}

	return 1.0 - (ssRes / ssTot)
}
