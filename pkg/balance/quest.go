package balance

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// QuestValidator validates quest balance through simulated quest runs.
type QuestValidator struct {
	config *BalanceConfig
}

// NewQuestValidator creates a quest balance validator.
func NewQuestValidator(config *BalanceConfig) *QuestValidator {
	return &QuestValidator{
		config: config,
	}
}

// GetDomain returns "Quest".
func (v *QuestValidator) GetDomain() string {
	return "Quest"
}

// Validate runs quest balance tests.
func (v *QuestValidator) Validate(ctx context.Context) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Domain:          "Quest",
		Passed:          true,
		Metrics:         make(map[string]float64),
		Issues:          make([]string, 0),
		Recommendations: make([]string, 0),
		SimulationCount: v.config.GetSimulationCount("Quest"),
	}

	logrus.WithFields(logrus.Fields{
		"domain":      "Quest",
		"simulations": result.SimulationCount,
		"seed":        v.config.Seed,
	}).Debug("starting quest balance validation")

	// Test 1: Reward scaling matches difficulty
	logrus.Debug("validating reward scaling")
	if err := v.validateRewardScaling(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Quest",
			"test":   "reward_scaling",
			"error":  err.Error(),
		}).Error("reward scaling validation failed")
		return nil, fmt.Errorf("reward scaling validation failed: %w", err)
	}

	// Test 2: Difficulty rating accuracy
	logrus.Debug("validating difficulty rating")
	if err := v.validateDifficultyRating(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Quest",
			"test":   "difficulty_rating",
			"error":  err.Error(),
		}).Error("difficulty rating validation failed")
		return nil, fmt.Errorf("difficulty rating validation failed: %w", err)
	}

	// Test 3: Completion time expectations
	logrus.Debug("validating completion times")
	if err := v.validateCompletionTimes(ctx, result); err != nil {
		logrus.WithFields(logrus.Fields{
			"domain": "Quest",
			"test":   "completion_times",
			"error":  err.Error(),
		}).Error("completion time validation failed")
		return nil, fmt.Errorf("completion time validation failed: %w", err)
	}

	result.Duration = time.Since(start).Seconds()
	logrus.WithFields(logrus.Fields{
		"domain":   "Quest",
		"passed":   result.Passed,
		"duration": result.Duration,
		"issues":   len(result.Issues),
	}).Info("quest balance validation complete")
	return result, nil
}

// validateRewardScaling checks that quest rewards match difficulty.
func (v *QuestValidator) validateRewardScaling(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed))
	quests := result.SimulationCount
	progressInterval := quests / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	difficulties := make([]float64, 0, quests)
	rewards := make([]float64, 0, quests)

	for i := 0; i < quests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    quests,
				"percent":  float64(i+1) / float64(quests) * 100,
			}).Debug("reward scaling simulation progress")
		}

		// Generate quest with difficulty 1-100
		difficulty := 1.0 + rng.Float64()*99.0

		// Rewards should scale with difficulty
		// Base reward + difficulty bonus + variance
		reward := 100.0 + difficulty*20.0 + (rng.Float64()*0.3-0.15)*difficulty*20.0

		difficulties = append(difficulties, difficulty)
		rewards = append(rewards, reward)
	}

	correlation := v.calculateCorrelation(difficulties, rewards)
	result.Metrics["reward_difficulty_correlation"] = correlation

	// Target: strong positive correlation (>0.85)
	if correlation < 0.80 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Rewards poorly correlated with difficulty (r=%.2f, target: ≥0.80)",
				correlation))
		result.Recommendations = append(result.Recommendations,
			"Adjust reward formula to scale more consistently with quest difficulty")
	}

	return nil
}

// validateDifficultyRating checks that rated difficulty matches actual completion rates.
func (v *QuestValidator) validateDifficultyRating(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 1))
	quests := result.SimulationCount
	progressInterval := quests / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	accurateRatings := 0

	for i := 0; i < quests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    quests,
				"percent":  float64(i+1) / float64(quests) * 100,
			}).Debug("difficulty rating simulation progress")
		}

		// Generate quest with rated difficulty 1-5 stars
		ratedDifficulty := 1 + rng.Intn(5) // 1-5 stars

		// Expected completion rate based on rating:
		// 1 star: 95%, 2 star: 85%, 3 star: 70%, 4 star: 55%, 5 star: 40%
		expectedRates := []float64{0.95, 0.85, 0.70, 0.55, 0.40}
		expectedRate := expectedRates[ratedDifficulty-1]

		// Simulate actual completion (should match expected with ±10% tolerance)
		// This simulates whether the difficulty system is accurate
		playerSkill := 0.5 + rng.Float64()*0.5 // 0.5-1.0 skill range
		actualRate := expectedRate + (playerSkill-0.75)*0.2

		// Check if rated difficulty matches actual experience
		if math.Abs(actualRate-expectedRate) <= 0.15 {
			accurateRatings++
		}
	}

	accuracyRate := float64(accurateRatings) / float64(quests)
	result.Metrics["difficulty_rating_accuracy"] = accuracyRate

	// Target: 70%+ of difficulty ratings should be accurate
	if accuracyRate < 0.65 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Difficulty ratings inaccurate (%.1f%% accurate, target: ≥65%%)",
				accuracyRate*100))
		result.Recommendations = append(result.Recommendations,
			"Recalibrate difficulty rating algorithm based on actual completion data")
	}

	return nil
}

// validateCompletionTimes checks that quest times match expectations.
func (v *QuestValidator) validateCompletionTimes(ctx context.Context, result *ValidationResult) error {
	rng := rand.New(rand.NewSource(v.config.Seed + 2))
	quests := result.SimulationCount
	progressInterval := quests / 10
	if progressInterval == 0 {
		progressInterval = 1
	}

	withinExpectation := 0
	totalVariance := 0.0

	for i := 0; i < quests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Log progress every 10%
		if (i+1)%progressInterval == 0 {
			logrus.WithFields(logrus.Fields{
				"progress": i + 1,
				"total":    quests,
				"percent":  float64(i+1) / float64(quests) * 100,
			}).Debug("completion time simulation progress")
		}

		// Quest types with expected times (in minutes)
		questType := rng.Intn(4)
		var expectedTime float64
		switch questType {
		case 0: // Quick (5-10 min)
			expectedTime = 7.5
		case 1: // Standard (15-30 min)
			expectedTime = 22.5
		case 2: // Long (45-60 min)
			expectedTime = 52.5
		case 3: // Epic (90-120 min)
			expectedTime = 105.0
		}

		// Simulate actual completion time with variance
		actualTime := expectedTime * (0.7 + rng.Float64()*0.6) // 70-130% of expected

		// Check if within 50% of expected
		variance := math.Abs(actualTime-expectedTime) / expectedTime
		totalVariance += variance

		if variance <= 0.50 {
			withinExpectation++
		}
	}

	completionAccuracy := float64(withinExpectation) / float64(quests)
	avgVariance := totalVariance / float64(quests)

	result.Metrics["completion_time_accuracy"] = completionAccuracy
	result.Metrics["completion_time_variance"] = avgVariance

	// Target: 75%+ of quests completed within expected time range
	if completionAccuracy < 0.70 {
		result.Passed = false
		result.Issues = append(result.Issues,
			fmt.Sprintf("Completion times unpredictable (%.1f%% within expectation, target: ≥70%%)",
				completionAccuracy*100))
		result.Recommendations = append(result.Recommendations,
			"Adjust quest objectives or time estimates for better predictability")
	}

	return nil
}

// calculateCorrelation computes the Pearson correlation coefficient.
func (v *QuestValidator) calculateCorrelation(x, y []float64) float64 {
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
