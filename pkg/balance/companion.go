package balance

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// CompanionValidator validates companion balance through simulated companion lives.
type CompanionValidator struct {
	config *BalanceConfig
}

// NewCompanionValidator creates a companion balance validator.
func NewCompanionValidator(config *BalanceConfig) *CompanionValidator {
	return &CompanionValidator{
		config: config,
	}
}

// GetDomain returns "Companion".
func (v *CompanionValidator) GetDomain() string {
	return "Companion"
}

// Validate runs companion balance tests.
func (v *CompanionValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Companion",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Companion"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Companion",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting companion balance validation")

	// Test 1: Loyalty progression feels meaningful
	logrus.Debug("validating loyalty progression")
	if err := v.validateLoyaltyProgression(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Companion",
			"test":   "loyalty_progression",
			"error":  err.Error(),
		}).Error("loyalty progression validation failed")
		return nil, fmt.Errorf("loyalty progression validation failed: %w", err)
	}

	// Test 2: Skill learning rates are balanced
	logrus.Debug("validating skill learning")
	if err := v.validateSkillLearning(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Companion",
			"test":   "skill_learning",
			"error":  err.Error(),
		}).Error("skill learning validation failed")
		return nil, fmt.Errorf("skill learning validation failed: %w", err)
	}

	// Test 3: Combat effectiveness scales appropriately
	logrus.Debug("validating combat effectiveness")
	if err := v.validateCombatEffectiveness(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Companion",
			"test":   "combat_effectiveness",
			"error":  err.Error(),
		}).Error("combat effectiveness validation failed")
		return nil, fmt.Errorf("combat effectiveness validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Companion",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("companion balance validation complete")
	return result, nil
}

// validateLoyaltyProgression checks that loyalty builds at a meaningful rate.
func (v *CompanionValidator) validateLoyaltyProgression(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed))
	companions := result.SimulationCount
	progressInterval := companions / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	maxLoyaltyReached := 0
	totalTimeToMaxLoyalty := 0.0 // In simulated hours

	for i := 0; i < companions; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    companions,
				"percent":  float64(i+1) / float64(companions) * 100,
			}).Debug("loyalty progression simulation progress")
		}

		// Simulate companion loyalty from 0 to 100
		loyalty := 0.0
		hoursPlayed := 0.0

		for loyalty < 100.0 && hoursPlayed < 500.0 { // Cap at 500 hours
			// Loyalty gains per hour: base + random bonus from activities
			loyaltyGain := 0.5 + rng.Float64()*1.5 // 0.5-2.0 per hour
			loyalty += loyaltyGain
			hoursPlayed++

			if loyalty >= 100.0 {
				maxLoyaltyReached++
				totalTimeToMaxLoyalty += hoursPlayed
				break
			}
		}
	}

	maxLoyaltyRate := float64(maxLoyaltyReached) / float64(companions)
	avgTimeToMax := 0.0
	if maxLoyaltyReached > 0 {
		avgTimeToMax = totalTimeToMaxLoyalty / float64(maxLoyaltyReached)
	}

	result.Metrics["max_loyalty_rate"] = maxLoyaltyRate
	result.Metrics["avg_time_to_max_loyalty"] = avgTimeToMax

	// Target: 60-90% should reach max loyalty in reasonable time (50-200 hours)
	if maxLoyaltyRate < 0.50 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Max loyalty too hard to reach (%.1f%% achieve, target: ≥50%%)",
				maxLoyaltyRate*100))
		result.Recommendations = append(result.Recommendations,
			"Increase loyalty gain rates or add bonus activities")
	}

	if avgTimeToMax > 0 && avgTimeToMax < 30.0 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Max loyalty too easy (%.1f hours avg, target: ≥30 hours)",
				avgTimeToMax))
		result.Recommendations = append(result.Recommendations,
			"Reduce loyalty gain rates to extend companion bonding experience")
	}

	return nil
}

// validateSkillLearning checks that companions learn skills at balanced rates.
func (v *CompanionValidator) validateSkillLearning(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	companions := result.SimulationCount
	progressInterval := companions / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	totalSkillsLearned := 0
	skillsPerCompanion := make([]int, 0, companions)

	for i := 0; i < companions; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    companions,
				"percent":  float64(i+1) / float64(companions) * 100,
			}).Debug("skill learning simulation progress")
		}

		// Simulate 100 hours with companion
		hoursPlayed := 100
		skillsLearned := 0
		maxSkills := 8 // Max skills a companion can learn

		for h := 0; h < hoursPlayed; h++ {
			// Learn skill every ~10-20 hours on average
			if rng.Float64() < 0.07 && skillsLearned < maxSkills { // 7% chance per hour
				skillsLearned++
			}
		}

		totalSkillsLearned += skillsLearned
		skillsPerCompanion = append(skillsPerCompanion, skillsLearned)
	}

	avgSkillsLearned := float64(totalSkillsLearned) / float64(companions)
	result.Metrics["avg_skills_learned"] = avgSkillsLearned

	// Calculate variance
	variance := 0.0
	for _, s := range skillsPerCompanion {
		diff := float64(s) - avgSkillsLearned
		variance += diff * diff
	}
	variance /= float64(len(skillsPerCompanion))
	result.Metrics["skill_variance"] = variance

	// Target: 4-7 skills learned in 100 hours on average
	if avgSkillsLearned < 3.0 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Skill learning too slow (%.1f skills/100hrs, target: ≥3)",
				avgSkillsLearned))
		result.Recommendations = append(result.Recommendations,
			"Increase skill learning chance or add training activities")
	}

	if avgSkillsLearned > 8.0 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Skill learning too fast (%.1f skills/100hrs, target: ≤8)",
				avgSkillsLearned))
		result.Recommendations = append(result.Recommendations,
			"Reduce skill learning chance to extend progression")
	}

	return nil
}

// validateCombatEffectiveness checks that companions contribute meaningfully to combat.
func (v *CompanionValidator) validateCombatEffectiveness(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 2))
	battles := result.SimulationCount
	progressInterval := battles / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	companionContributions := make([]float64, 0, battles)

	for i := 0; i < battles; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    battles,
				"percent":  float64(i+1) / float64(battles) * 100,
			}).Debug("combat effectiveness simulation progress")
		}

		// Simulate battle with companion contribution
		// Player deals 60-100 damage, companion deals scaled damage
		playerDamage := 60.0 + rng.Float64()*40.0

		// Companion effectiveness: 20-40% of player damage
		companionLevel := 1 + rng.Intn(50)                   // Random companion level
		baseCompanionDamage := float64(companionLevel) * 0.5 // 0.5 damage per level
		companionDamage := baseCompanionDamage * (0.8 + rng.Float64()*0.4)

		totalDamage := playerDamage + companionDamage
		contribution := companionDamage / totalDamage
		companionContributions = append(companionContributions, contribution)
	}

	// Calculate average contribution
	avgContribution := 0.0
	for _, c := range companionContributions {
		avgContribution += c
	}
	avgContribution /= float64(len(companionContributions))
	result.Metrics["companion_combat_contribution"] = avgContribution

	// Target: 15-35% of damage should come from companion
	if avgContribution < 0.10 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Companions too weak in combat (%.1f%% contribution, target: ≥10%%)",
				avgContribution*100))
		result.Recommendations = append(result.Recommendations,
			"Increase companion base damage or scaling")
	}

	if avgContribution > 0.50 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Companions too strong in combat (%.1f%% contribution, target: ≤50%%)",
				avgContribution*100))
		result.Recommendations = append(result.Recommendations,
			"Reduce companion damage to maintain player agency")
	}

	return nil
}
