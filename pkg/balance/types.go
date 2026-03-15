package balance

import "context"

// BalanceValidator defines the interface for domain-specific balance validators.
// Each validator simulates gameplay scenarios and measures balance metrics.
type BalanceValidator interface {
	// Validate runs balance tests and returns results.
	// Context allows cancellation of long-running validations.
	Validate(ctx context.Context) (*ValidationResult, error)

	// GetDomain returns the balance domain name (e.g., "Combat", "Economic").
	GetDomain() string
}

// ValidationResult contains the outcome of a balance validation run.
type ValidationResult struct {
	// Domain is the balance domain tested (e.g., "Combat").
	Domain string

	// Passed indicates if all acceptance criteria were met.
	Passed bool

	// Metrics maps metric names to their measured values.
	// Examples: "class_win_rate_variance", "weapon_usage_max", "xp_curve_r_squared"
	Metrics map[string]float64

	// Issues lists detected balance problems (empty if Passed is true).
	Issues []string

	// Recommendations suggests fixes for detected issues.
	Recommendations []string

	// SimulationCount is the number of scenarios tested.
	SimulationCount int

	// Duration is the validation runtime in seconds.
	Duration float64
}

// BalanceConfig holds global configuration for balance validation.
type BalanceConfig struct {
	// Seed for deterministic simulation.
	Seed int64

	// SimulationCounts per domain (default: 1000 per domain).
	SimulationCounts map[string]int

	// AcceptanceThresholds override default thresholds (optional).
	// Example: {"class_win_rate_min": 0.40, "class_win_rate_max": 0.60}
	AcceptanceThresholds map[string]float64

	// ProgressLogInterval controls how often progress is logged as a percentage
	// of total iterations (e.g., 10 = log every 10%). Zero or negative values
	// use the default of 10.
	ProgressLogInterval int
}

// NewDefaultConfig returns a BalanceConfig with standard settings.
func NewDefaultConfig() *BalanceConfig {
	return &BalanceConfig{
		Seed:                12345,
		ProgressLogInterval: 10, // log every 10% by default
		SimulationCounts: map[string]int{
			"Combat":      10000, // 10k simulated battles
			"Economic":    5000,  // 5k transactions
			"Progression": 1000,  // 1k player journeys
			"Social":      2000,  // 2k interactions
			"Housing":     500,   // 500 builds
			"Vehicle":     2000,  // 2k trips
			"Companion":   1000,  // 1k companion lives
			"Quest":       2000,  // 2k quest runs
		},
		AcceptanceThresholds: map[string]float64{
			// Combat thresholds
			"class_win_rate_min":      0.45,
			"class_win_rate_max":      0.55,
			"weapon_usage_min":        0.05,
			"weapon_usage_max":        0.40,
			"enemy_scaling_r_squared": 0.95,
			"boss_failure_rate_min":   0.60,
			"boss_failure_rate_max":   0.80,

			// Economic thresholds
			"loot_value_correlation": 0.80,
			"crafting_profit_min":    0.10,
			"crafting_profit_max":    0.30,
			"gold_sink_ratio_min":    0.80,
			"gold_sink_ratio_max":    1.20,

			// Progression thresholds
			"xp_curve_r_squared":       0.98,
			"skill_power_increase_min": 0.10,
			"skill_power_increase_max": 0.20,
			"stat_scaling_r_squared":   0.98,

			// Social thresholds
			"defender_advantage_min": 0.55,
			"defender_advantage_max": 0.60,
			"fraud_rate_max":         0.01,
			"chat_reject_rate_max":   0.05,
		},
	}
}

// GetThreshold retrieves an acceptance threshold, falling back to default if not set.
func (c *BalanceConfig) GetThreshold(name string) float64 {
	if val, ok := c.AcceptanceThresholds[name]; ok {
		return val
	}
	// Fallback defaults (shouldn't happen with NewDefaultConfig)
	defaults := map[string]float64{
		"class_win_rate_min": 0.45,
		"class_win_rate_max": 0.55,
	}
	if val, ok := defaults[name]; ok {
		return val
	}
	return 0.0 // Unknown threshold
}

// GetSimulationCount retrieves the simulation count for a domain, defaulting to 1000.
func (c *BalanceConfig) GetSimulationCount(domain string) int {
	if count, ok := c.SimulationCounts[domain]; ok {
		return count
	}
	return 1000 // Default
}

// GetProgressInterval returns the number of iterations between progress log entries
// for a loop of totalIterations. Uses ProgressLogInterval as a percentage.
func (c *BalanceConfig) GetProgressInterval(totalIterations int) int {
	pct := c.ProgressLogInterval
	if pct <= 0 {
		pct = 10
	}
	interval := totalIterations * pct / 100
	if interval <= 0 {
		interval = 1
	}
	return interval
}
