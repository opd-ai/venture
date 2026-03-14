// Package resilience scenario implements test scenario execution
// for validating network behavior under specific conditions.
package resilience

import (
	"context"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// RunScenario executes a test scenario using the provided simulator and collector,
// running simulated network traffic for the scenario's duration and validating
// the results against acceptance criteria.
//
// The function configures the simulator with the scenario's network settings,
// generates synthetic traffic, and collects metrics during execution.
// After the duration expires, it validates the collected stats against
// the scenario's acceptance criteria (max desync rate, max misprediction rate,
// max reconnect time).
//
// Example usage:
//
//	sim := resilience.NewNetworkSimulatorWithSeed(12345)
//	collector := resilience.NewMetricsCollector()
//	result := resilience.RunScenario(context.Background(), &resilience.HighLatencyScenario, sim, collector)
//	if result.Failed() {
//	    logrus.WithFields(logrus.Fields{
//	        "scenario": "HighLatencyScenario",
//	        "failure_reason": result.FailureReason,
//	    }).Error("Scenario failed")
//	}
func RunScenario(ctx context.Context, scenario *TestScenario, sim *NetworkSimulator, collector *MetricsCollector) *ScenarioResult {
	return RunScenarioWithOptions(ctx, scenario, sim, collector, DefaultScenarioOptions())
}

// ScenarioOptions configures scenario execution behavior.
type ScenarioOptions struct {
	// PacketSize is the size of synthetic test packets in bytes.
	PacketSize int

	// PacketsPerSecond is the rate of synthetic traffic generation.
	PacketsPerSecond int

	// SimulatePredictions enables misprediction simulation based on latency.
	SimulatePredictions bool

	// SimulateDesyncs enables desync simulation based on packet loss.
	SimulateDesyncs bool

	// Logger is the logger for scenario execution. If nil, logging is disabled.
	Logger *logrus.Logger
}

// DefaultScenarioOptions returns default scenario execution options.
func DefaultScenarioOptions() ScenarioOptions {
	return ScenarioOptions{
		PacketSize:          256,
		PacketsPerSecond:    30,
		SimulatePredictions: true,
		SimulateDesyncs:     true,
		Logger:              nil,
	}
}

// RunScenarioWithOptions executes a test scenario with custom options.
func RunScenarioWithOptions(ctx context.Context, scenario *TestScenario, sim *NetworkSimulator, collector *MetricsCollector, opts ScenarioOptions) *ScenarioResult {
	result := &ScenarioResult{
		Scenario:  scenario,
		Timestamp: time.Now(),
		Passed:    true,
	}

	// Validate inputs
	if scenario == nil {
		result.Passed = false
		result.FailureReason = "scenario is nil"
		return result
	}
	if sim == nil {
		result.Passed = false
		result.FailureReason = "simulator is nil"
		return result
	}
	if collector == nil {
		result.Passed = false
		result.FailureReason = "collector is nil"
		return result
	}

	// Configure simulator with scenario settings
	if err := sim.SetConfig(scenario.Config); err != nil {
		result.Passed = false
		result.FailureReason = "failed to configure simulator: " + err.Error()
		return result
	}

	// Reset collector for fresh metrics
	collector.Reset()

	// Log scenario start
	if opts.Logger != nil {
		opts.Logger.WithFields(logrus.Fields{
			"scenario":    scenario.Name,
			"latency":     scenario.Config.Latency,
			"packet_loss": scenario.Config.PacketLossRate,
			"duration":    scenario.Duration,
		}).Info("Starting scenario execution")
	}

	// Calculate timing
	duration := scenario.Duration
	if duration == 0 {
		duration = 5 * time.Second // Default duration for testing
	}
	tickInterval := time.Second / time.Duration(opts.PacketsPerSecond)
	if tickInterval < time.Millisecond {
		tickInterval = time.Millisecond
	}

	// Create cancellable context with timeout
	execCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	// Run scenario execution loop
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	packetData := make([]byte, opts.PacketSize)

	for {
		select {
		case <-execCtx.Done():
			// Duration complete or context cancelled
			goto evaluate
		case <-ticker.C:
			// Send synthetic packet
			err := sim.Send(packetData)
			if err == ErrPacketDropped {
				collector.RecordPacketLoss()
				if opts.SimulateDesyncs {
					// Simulate desync on high packet loss
					if sim.GetConfig().PacketLossRate > 0.15 {
						collector.RecordDesync()
					}
				}
			} else if err == ErrBandwidthExceeded {
				// Bandwidth exceeded - record but don't count as loss
				collector.RecordPacketSent(opts.PacketSize)
			} else if err == nil {
				collector.RecordPacketSent(opts.PacketSize)
				collector.RecordLatency(scenario.Config.Latency)

				// Simulate predictions
				if opts.SimulatePredictions {
					// Higher latency = higher misprediction rate
					mispredicted := scenario.Config.Latency > 500*time.Millisecond
					collector.RecordPrediction(mispredicted)
				}
			}
		}
	}

evaluate:
	// Collect final stats
	result.Stats = collector.GetStats()

	// Evaluate acceptance criteria
	evaluateResult(result, scenario, opts.Logger)

	return result
}

// evaluateResult checks the collected stats against scenario acceptance criteria.
func evaluateResult(result *ScenarioResult, scenario *TestScenario, logger *logrus.Logger) {
	stats := result.Stats

	// Calculate desync rate (per hour)
	durationHours := stats.Duration.Hours()
	if durationHours < 0.001 {
		durationHours = 0.001 // Prevent division by zero
	}
	desyncRate := float64(stats.DesyncCount) / durationHours

	// Check desync rate
	if scenario.MaxDesyncRate > 0 && desyncRate > scenario.MaxDesyncRate {
		result.Passed = false
		result.FailureReason = formatFailure("desync rate exceeded",
			desyncRate, scenario.MaxDesyncRate, "desyncs/hour")
		logFailure(logger, result.FailureReason, scenario.Name)
		return
	}

	// Check misprediction rate
	if scenario.MaxMispredictionRate > 0 && stats.MispredictionRate > scenario.MaxMispredictionRate {
		result.Passed = false
		result.FailureReason = formatFailure("misprediction rate exceeded",
			stats.MispredictionRate, scenario.MaxMispredictionRate, "rate")
		logFailure(logger, result.FailureReason, scenario.Name)
		return
	}

	// Check reconnect time (only if reconnects occurred)
	if stats.ReconnectCount > 0 && scenario.MaxReconnectTime > 0 {
		if stats.AvgReconnectTime > scenario.MaxReconnectTime {
			result.Passed = false
			result.FailureReason = formatFailure("reconnect time exceeded",
				stats.AvgReconnectTime.Seconds(), scenario.MaxReconnectTime.Seconds(), "seconds")
			logFailure(logger, result.FailureReason, scenario.Name)
			return
		}
	}

	// Log success
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"scenario":           scenario.Name,
			"desync_rate":        desyncRate,
			"misprediction_rate": stats.MispredictionRate,
			"packets_sent":       stats.PacketsSent,
			"packets_dropped":    stats.PacketsDropped,
		}).Info("Scenario passed acceptance criteria")
	}
}

// formatFailure creates a formatted failure message.
func formatFailure(msg string, actual, max float64, unit string) string {
	return msg + ": " + formatFloat(actual) + " > " + formatFloat(max) + " " + unit
}

// formatFloat formats a float64 to a reasonable precision.
func formatFloat(v float64) string {
	if v < 0.01 {
		return "0.00"
	}
	if v >= 1000 {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// formatInt formats an int64 as a decimal string.
func formatInt(v int64) string {
	return strconv.FormatInt(v, 10)
}

// logFailure logs a scenario failure with structured fields.
func logFailure(logger *logrus.Logger, reason, scenarioName string) {
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"scenario": scenarioName,
			"reason":   reason,
		}).Warn("Scenario failed acceptance criteria")
	}
}

// RunAllScenarios executes all pre-defined scenarios and returns results.
// This is useful for comprehensive resilience testing.
func RunAllScenarios(ctx context.Context, sim *NetworkSimulator, collector *MetricsCollector) []*ScenarioResult {
	return RunAllScenariosWithOptions(ctx, sim, collector, DefaultScenarioOptions())
}

// RunAllScenariosWithOptions executes all scenarios with custom options.
func RunAllScenariosWithOptions(ctx context.Context, sim *NetworkSimulator, collector *MetricsCollector, opts ScenarioOptions) []*ScenarioResult {
	results := make([]*ScenarioResult, 0, len(AllScenarios))

	for _, scenario := range AllScenarios {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results
		default:
		}

		result := RunScenarioWithOptions(ctx, scenario, sim, collector, opts)
		results = append(results, result)

		// Reset between scenarios
		sim.Reset()
		collector.Reset()
	}

	return results
}
