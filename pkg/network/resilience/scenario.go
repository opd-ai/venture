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

	if err := validateScenarioInputs(scenario, sim, collector); err != nil {
		result.Passed = false
		result.FailureReason = err.Error()
		return result
	}

	if err := initializeScenario(scenario, sim, collector, opts); err != nil {
		result.Passed = false
		result.FailureReason = err.Error()
		return result
	}

	duration, tickInterval := calculateScenarioTiming(scenario, opts)
	execCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	runScenarioLoop(execCtx, sim, collector, opts, scenario, tickInterval)

	result.Stats = collector.GetStats()
	evaluateResult(result, scenario, opts.Logger)

	return result
}

// validateScenarioInputs checks that all required inputs are non-nil.
func validateScenarioInputs(scenario *TestScenario, sim *NetworkSimulator, collector *MetricsCollector) error {
	if scenario == nil {
		return errScenarioNil
	}
	if sim == nil {
		return errSimulatorNil
	}
	if collector == nil {
		return errCollectorNil
	}
	return nil
}

// Sentinel errors for input validation.
var (
	errScenarioNil  = scenarioError("scenario is nil")
	errSimulatorNil = scenarioError("simulator is nil")
	errCollectorNil = scenarioError("collector is nil")
)

type scenarioError string

func (e scenarioError) Error() string { return string(e) }

// initializeScenario configures the simulator and resets the collector.
func initializeScenario(scenario *TestScenario, sim *NetworkSimulator, collector *MetricsCollector, opts ScenarioOptions) error {
	if err := sim.SetConfig(scenario.Config); err != nil {
		return scenarioError("failed to configure simulator: " + err.Error())
	}
	collector.Reset()
	logScenarioStart(scenario, opts.Logger)
	return nil
}

// logScenarioStart logs the start of scenario execution.
func logScenarioStart(scenario *TestScenario, logger *logrus.Logger) {
	if logger == nil {
		return
	}
	logger.WithFields(logrus.Fields{
		"scenario":    scenario.Name,
		"latency":     scenario.Config.Latency,
		"packet_loss": scenario.Config.PacketLossRate,
		"duration":    scenario.Duration,
	}).Info("Starting scenario execution")
}

// calculateScenarioTiming returns the duration and tick interval for the scenario.
func calculateScenarioTiming(scenario *TestScenario, opts ScenarioOptions) (time.Duration, time.Duration) {
	duration := scenario.Duration
	if duration == 0 {
		duration = 5 * time.Second
	}
	tickInterval := time.Second / time.Duration(opts.PacketsPerSecond)
	if tickInterval < time.Millisecond {
		tickInterval = time.Millisecond
	}
	return duration, tickInterval
}

// runScenarioLoop sends packets and collects metrics until context expires.
func runScenarioLoop(ctx context.Context, sim *NetworkSimulator, collector *MetricsCollector, opts ScenarioOptions, scenario *TestScenario, tickInterval time.Duration) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	packetData := make([]byte, opts.PacketSize)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processPacketSend(sim, collector, opts, scenario, packetData)
		}
	}
}

// processPacketSend handles sending a packet and recording the appropriate metric.
func processPacketSend(sim *NetworkSimulator, collector *MetricsCollector, opts ScenarioOptions, scenario *TestScenario, packetData []byte) {
	err := sim.Send(packetData)
	switch err {
	case ErrPacketDropped:
		collector.RecordPacketLoss()
		if opts.SimulateDesyncs && sim.GetConfig().PacketLossRate > 0.15 {
			collector.RecordDesync()
		}
	case ErrBandwidthExceeded:
		collector.RecordPacketSent(opts.PacketSize)
	case nil:
		collector.RecordPacketSent(opts.PacketSize)
		collector.RecordLatency(scenario.Config.Latency)
		if opts.SimulatePredictions {
			mispredicted := scenario.Config.Latency > 500*time.Millisecond
			collector.RecordPrediction(mispredicted)
		}
	}
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

// formatFloat formats a float64 to a reasonable precision, trimming trailing zeros.
func formatFloat(v float64) string {
	if v < 0.01 {
		return "0.00"
	}
	if v >= 1000 {
		return strconv.FormatInt(int64(v), 10)
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	// Trim trailing zeros after decimal point.
	if len(s) > 2 && s[len(s)-2] == '.' {
		return s[:len(s)-1]
	}
	if len(s) > 1 && s[len(s)-1] == '0' && s[len(s)-2] != '.' {
		s = s[:len(s)-1]
	}
	if len(s) > 1 && s[len(s)-1] == '0' && s[len(s)-2] == '.' {
		s = s[:len(s)-2]
	}
	return s
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
