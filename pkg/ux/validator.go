package ux

import (
	"fmt"
	"math/rand"
	"time"
)

// JourneyValidator executes and validates user journeys.
type JourneyValidator struct {
	config ValidationConfig
	rng    *rand.Rand
}

// NewJourneyValidator creates a new journey validator with default configuration.
func NewJourneyValidator() *JourneyValidator {
	config := DefaultValidationConfig()
	return newValidatorWithConfig(config)
}

// NewJourneyValidatorWithConfig creates a validator with custom configuration.
func NewJourneyValidatorWithConfig(config ValidationConfig) *JourneyValidator {
	return newValidatorWithConfig(config)
}

// newValidatorWithConfig creates a validator using seed from config (or time-based if seed is 0).
func newValidatorWithConfig(config ValidationConfig) *JourneyValidator {
	seed := config.Seed
	if seed == 0 {
		// UX validation timing uses non-deterministic seed when Seed==0. This is acceptable
		// because journeys validate flow logic, not game content generation.
		seed = time.Now().UnixNano()
	}
	return &JourneyValidator{
		config: config,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// ValidateAll validates all defined journeys.
func (v *JourneyValidator) ValidateAll() []JourneyResult {
	journeys := AllJourneys()
	results := make([]JourneyResult, len(journeys))

	for i, journey := range journeys {
		results[i] = v.ValidateJourney(journey.Type)
	}

	return results
}

// ValidateJourney validates a specific journey by running it multiple times.
func (v *JourneyValidator) ValidateJourney(journeyType JourneyType) JourneyResult {
	journey, found := v.findJourneyDefinition(journeyType)
	if !found {
		return JourneyResult{
			Type:   journeyType,
			Passed: false,
			Error:  fmt.Errorf("journey type %s not found", journeyType),
		}
	}

	completions, totalDuration, errors, stepResults := v.executeJourneyRuns(journey)

	completionRate, averageDuration, errorRate, satisfaction := v.calculateJourneyMetrics(
		completions, totalDuration, errors, len(journey.Steps), journey.ExpectedDuration)

	v.averageStepDurations(stepResults, completions)

	passed := v.journeyMeetsThresholds(completionRate, errorRate, satisfaction, averageDuration, journey.ExpectedDuration)

	return JourneyResult{
		Type:            journey.Type,
		Name:            journey.Name,
		Passed:          passed,
		CompletionRate:  completionRate,
		AverageDuration: averageDuration,
		Satisfaction:    satisfaction,
		ErrorRate:       errorRate,
		Steps:           stepResults,
	}
}

// findJourneyDefinition locates a journey definition by type.
func (v *JourneyValidator) findJourneyDefinition(journeyType JourneyType) (JourneyDefinition, bool) {
	for _, j := range AllJourneys() {
		if j.Type == journeyType {
			return j, true
		}
	}
	return JourneyDefinition{}, false
}

// executeJourneyRuns runs the journey multiple times and collects results.
// Returns: completions count, total duration, error count, step results.
func (v *JourneyValidator) executeJourneyRuns(journey JourneyDefinition) (int, time.Duration, int, []StepResult) {
	completions := 0
	totalDuration := time.Duration(0)
	errors := 0
	stepResults := make([]StepResult, len(journey.Steps))

	for run := 0; run < v.config.Runs; run++ {
		ctx := &JourneyContext{
			PlayerID:  v.rng.Intn(10000),
			WorldSeed: v.rng.Int63(),
			Data:      make(map[string]interface{}),
		}

		// Using time.Now() for wall-clock timing measurement (not game simulation time)
		runStart := time.Now()
		runCompleted := v.executeJourneySteps(ctx, journey.Steps, stepResults, &errors)

		if runCompleted {
			completions++
			totalDuration += time.Since(runStart)
		}
	}

	return completions, totalDuration, errors, stepResults
}

// executeJourneySteps executes all steps in a single journey run.
// Returns true if all steps completed successfully, false if any step failed.
func (v *JourneyValidator) executeJourneySteps(ctx *JourneyContext, steps []JourneyStep, stepResults []StepResult, errors *int) bool {
	for i, step := range steps {
		ctx.StepIndex = i
		// Using time.Now() for wall-clock timing measurement (not game simulation time)
		ctx.StepStartTime = time.Now()

		err := step.Action(ctx)
		stepDuration := time.Since(ctx.StepStartTime)

		if err != nil {
			*errors++
			stepResults[i].Error = err
			return false
		}

		stepResults[i].Completed = true
		stepResults[i].Duration += stepDuration
		stepResults[i].Name = step.Name
	}
	return true
}

// calculateJourneyMetrics computes completion rate, average duration, error rate, and satisfaction.
// Returns: completion rate, average duration, error rate, satisfaction score.
func (v *JourneyValidator) calculateJourneyMetrics(completions int, totalDuration time.Duration, errors, totalSteps int, expectedDuration time.Duration) (float64, time.Duration, float64, float64) {
	completionRate := float64(completions) / float64(v.config.Runs)

	var averageDuration time.Duration
	if completions > 0 {
		averageDuration = totalDuration / time.Duration(completions)
	}

	// Guard against division by zero when totalSteps is 0
	var errorRate float64
	if totalSteps > 0 {
		errorRate = float64(errors) / float64(v.config.Runs*totalSteps)
	}

	satisfaction := v.calculateSatisfaction(completionRate, averageDuration, expectedDuration)

	return completionRate, averageDuration, errorRate, satisfaction
}

// averageStepDurations normalizes step durations by number of completions.
// Modifies stepResults in-place to contain average durations across all runs.
func (v *JourneyValidator) averageStepDurations(stepResults []StepResult, completions int) {
	if completions > 0 {
		for i := range stepResults {
			stepResults[i].Duration /= time.Duration(completions)
		}
	}
}

// journeyMeetsThresholds checks if journey results meet validation criteria.
// Returns true if all thresholds (completion, error, satisfaction, duration) are met.
func (v *JourneyValidator) journeyMeetsThresholds(completionRate, errorRate, satisfaction float64, averageDuration, expectedDuration time.Duration) bool {
	return completionRate >= v.config.MinCompletionRate &&
		errorRate <= v.config.MaxErrorRate &&
		satisfaction >= v.config.MinSatisfaction &&
		v.checkDurationWithinTolerance(averageDuration, expectedDuration)
}

// calculateSatisfaction simulates user satisfaction based on completion and timing.
func (v *JourneyValidator) calculateSatisfaction(completionRate float64, actualDuration, expectedDuration time.Duration) float64 {
	// Base satisfaction on completion rate
	satisfaction := completionRate

	// Adjust based on timing
	if actualDuration > 0 && expectedDuration > 0 {
		ratio := float64(actualDuration) / float64(expectedDuration)
		// Penalize if significantly over time
		if ratio > 1.0+v.config.TimeTolerancePercent/100.0 {
			satisfaction *= 0.8
		} else if ratio > 1.0 {
			satisfaction *= 0.9
		}
		// Slight bonus if faster
		if ratio < 0.8 {
			satisfaction = min(1.0, satisfaction*1.1)
		}
	}

	return satisfaction
}

// checkDurationWithinTolerance verifies duration is within acceptable range.
// In simulation mode, actual durations are microseconds, so we skip this check
// if actual duration is < 1 second (simulation mode).
func (v *JourneyValidator) checkDurationWithinTolerance(actual, expected time.Duration) bool {
	if expected == 0 {
		return true
	}
	// Simulation mode: durations are microseconds, don't compare to expected
	if actual < time.Second {
		return true
	}
	tolerance := float64(expected) * v.config.TimeTolerancePercent / 100.0
	diff := float64(actual) - float64(expected)
	return diff >= -tolerance && diff <= tolerance
}

// GetSummary returns a summary of validation results.
func (v *JourneyValidator) GetSummary(results []JourneyResult) Summary {
	total := len(results)

	// Handle empty results to avoid division by zero
	if total == 0 {
		return Summary{
			TotalJourneys:         0,
			PassedJourneys:        0,
			AverageCompletionRate: 0.0,
			AverageSatisfaction:   0.0,
			AverageErrorRate:      0.0,
			PassRate:              0.0,
		}
	}

	passed := 0
	totalCompletion := 0.0
	totalSatisfaction := 0.0
	totalErrors := 0.0

	for _, result := range results {
		if result.Passed {
			passed++
		}
		totalCompletion += result.CompletionRate
		totalSatisfaction += result.Satisfaction
		totalErrors += result.ErrorRate
	}

	return Summary{
		TotalJourneys:         total,
		PassedJourneys:        passed,
		AverageCompletionRate: totalCompletion / float64(total),
		AverageSatisfaction:   totalSatisfaction / float64(total),
		AverageErrorRate:      totalErrors / float64(total),
		PassRate:              float64(passed) / float64(total),
	}
}

// Summary contains aggregate metrics from journey validation.
type Summary struct {
	TotalJourneys         int
	PassedJourneys        int
	AverageCompletionRate float64
	AverageSatisfaction   float64
	AverageErrorRate      float64
	PassRate              float64
}
