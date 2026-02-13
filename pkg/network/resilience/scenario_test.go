package resilience

import (
	"context"
	"testing"
	"time"
)

// TestRunScenario_BasicExecution tests basic scenario execution.
func TestRunScenario_BasicExecution(t *testing.T) {
	// Use deterministic seed for reproducible tests
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()

	// Create a short test scenario
	scenario := &TestScenario{
		Name:                 "Test Scenario",
		Description:          "Short test for basic execution",
		Config:               NetworkConfig{Latency: 50 * time.Millisecond, PacketLossRate: 0.01},
		Duration:             100 * time.Millisecond,
		MaxDesyncRate:        100.0, // Very permissive for testing
		MaxMispredictionRate: 1.0,   // Allow any misprediction rate
		MaxReconnectTime:     30 * time.Second,
	}

	ctx := context.Background()
	result := RunScenario(ctx, scenario, sim, collector)

	if result == nil {
		t.Fatal("RunScenario() returned nil")
	}

	if result.Scenario != scenario {
		t.Error("Result.Scenario does not match input scenario")
	}

	if result.Timestamp.IsZero() {
		t.Error("Result.Timestamp is zero")
	}

	// Should pass with permissive acceptance criteria
	if result.Failed() {
		t.Errorf("Scenario failed unexpectedly: %s", result.FailureReason)
	}
}

// TestRunScenario_NilInputs tests error handling for nil inputs.
func TestRunScenario_NilInputs(t *testing.T) {
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()
	scenario := &LowLatencyScenario

	tests := []struct {
		name        string
		scenario    *TestScenario
		sim         *NetworkSimulator
		collector   *MetricsCollector
		wantFailure string
	}{
		{
			name:        "nil_scenario",
			scenario:    nil,
			sim:         sim,
			collector:   collector,
			wantFailure: "scenario is nil",
		},
		{
			name:        "nil_simulator",
			scenario:    scenario,
			sim:         nil,
			collector:   collector,
			wantFailure: "simulator is nil",
		},
		{
			name:        "nil_collector",
			scenario:    scenario,
			sim:         sim,
			collector:   nil,
			wantFailure: "collector is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result := RunScenario(ctx, tt.scenario, tt.sim, tt.collector)

			if result == nil {
				t.Fatal("RunScenario() returned nil")
			}

			if !result.Failed() {
				t.Error("Expected scenario to fail with nil input")
			}

			if result.FailureReason != tt.wantFailure {
				t.Errorf("FailureReason = %q, want %q", result.FailureReason, tt.wantFailure)
			}
		})
	}
}

// TestRunScenario_ContextCancellation tests that scenarios respect context cancellation.
func TestRunScenario_ContextCancellation(t *testing.T) {
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()

	// Create a long scenario that we'll cancel early
	scenario := &TestScenario{
		Name:                 "Long Scenario",
		Description:          "Should be cancelled early",
		Config:               NetworkConfig{Latency: 10 * time.Millisecond},
		Duration:             10 * time.Second, // Long duration
		MaxDesyncRate:        100.0,
		MaxMispredictionRate: 1.0,
	}

	// Cancel after short delay
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	result := RunScenario(ctx, scenario, sim, collector)
	elapsed := time.Since(start)

	if result == nil {
		t.Fatal("RunScenario() returned nil")
	}

	// Should complete quickly due to cancellation
	if elapsed > 500*time.Millisecond {
		t.Errorf("Scenario took %v, expected <500ms due to cancellation", elapsed)
	}
}

// TestRunScenario_AcceptanceCriteria tests validation of acceptance criteria.
func TestRunScenario_AcceptanceCriteria(t *testing.T) {
	tests := []struct {
		name       string
		scenario   *TestScenario
		wantPassed bool
	}{
		{
			name: "passes_permissive_criteria",
			scenario: &TestScenario{
				Name:                 "Permissive",
				Config:               NetworkConfig{Latency: 10 * time.Millisecond, PacketLossRate: 0.01},
				Duration:             50 * time.Millisecond,
				MaxDesyncRate:        1000.0,
				MaxMispredictionRate: 1.0,
			},
			wantPassed: true,
		},
		{
			name: "zero_duration_uses_default",
			scenario: &TestScenario{
				Name:                 "Zero Duration",
				Config:               NetworkConfig{Latency: 10 * time.Millisecond},
				Duration:             0, // Should use default
				MaxDesyncRate:        1000.0,
				MaxMispredictionRate: 1.0,
			},
			wantPassed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := NewNetworkSimulatorWithSeed(12345)
			collector := NewMetricsCollector()

			ctx := context.Background()
			result := RunScenario(ctx, tt.scenario, sim, collector)

			if result == nil {
				t.Fatal("RunScenario() returned nil")
			}

			if result.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v (reason: %s)", result.Passed, tt.wantPassed, result.FailureReason)
			}
		})
	}
}

// TestRunScenarioWithOptions tests custom options.
func TestRunScenarioWithOptions(t *testing.T) {
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()

	scenario := &TestScenario{
		Name:                 "Options Test",
		Config:               NetworkConfig{Latency: 10 * time.Millisecond},
		Duration:             50 * time.Millisecond,
		MaxDesyncRate:        1000.0,
		MaxMispredictionRate: 1.0,
	}

	opts := ScenarioOptions{
		PacketSize:          512,
		PacketsPerSecond:    60,
		SimulatePredictions: false,
		SimulateDesyncs:     false,
	}

	ctx := context.Background()
	result := RunScenarioWithOptions(ctx, scenario, sim, collector, opts)

	if result == nil {
		t.Fatal("RunScenarioWithOptions() returned nil")
	}

	if result.Failed() {
		t.Errorf("Scenario failed: %s", result.FailureReason)
	}
}

// TestRunAllScenarios tests running all predefined scenarios.
func TestRunAllScenarios(t *testing.T) {
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()

	// Use very short timeout to make test fast
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Use short durations for testing
	opts := ScenarioOptions{
		PacketSize:          64,
		PacketsPerSecond:    10,
		SimulatePredictions: false,
		SimulateDesyncs:     false,
	}

	results := RunAllScenariosWithOptions(ctx, sim, collector, opts)

	// Should have at least one result before timeout
	if len(results) == 0 {
		t.Error("RunAllScenarios() returned no results")
	}

	// Verify each result has valid fields
	for _, result := range results {
		if result.Scenario == nil {
			t.Error("Result has nil scenario")
		}
		if result.Timestamp.IsZero() {
			t.Error("Result has zero timestamp")
		}
	}
}

// TestDefaultScenarioOptions verifies default options.
func TestDefaultScenarioOptions(t *testing.T) {
	opts := DefaultScenarioOptions()

	if opts.PacketSize <= 0 {
		t.Errorf("PacketSize = %d, want > 0", opts.PacketSize)
	}
	if opts.PacketsPerSecond <= 0 {
		t.Errorf("PacketsPerSecond = %d, want > 0", opts.PacketsPerSecond)
	}
	if !opts.SimulatePredictions {
		t.Error("SimulatePredictions should be true by default")
	}
	if !opts.SimulateDesyncs {
		t.Error("SimulateDesyncs should be true by default")
	}
}

// TestFormatFloat tests float formatting helper.
func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.0, "0.00"},
		{0.001, "0.00"},
		{1.0, "1"},
		{1.5, "1.5"},
		{1.23, "1.23"},
		{100.0, "100"},
		{1000.0, "1000"},
		{1234.56, "1234"},
	}

	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestFormatInt tests integer formatting helper.
func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{12, "12"},
		{123, "123"},
		{1234, "1234"},
		{-1, "-1"},
		{-123, "-123"},
	}

	for _, tt := range tests {
		got := formatInt(tt.input)
		if got != tt.want {
			t.Errorf("formatInt(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// BenchmarkRunScenario benchmarks scenario execution.
func BenchmarkRunScenario(b *testing.B) {
	sim := NewNetworkSimulatorWithSeed(12345)
	collector := NewMetricsCollector()

	scenario := &TestScenario{
		Name:                 "Benchmark Scenario",
		Config:               NetworkConfig{Latency: 1 * time.Millisecond},
		Duration:             10 * time.Millisecond,
		MaxDesyncRate:        1000.0,
		MaxMispredictionRate: 1.0,
	}

	opts := ScenarioOptions{
		PacketSize:          64,
		PacketsPerSecond:    100,
		SimulatePredictions: false,
		SimulateDesyncs:     false,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunScenarioWithOptions(ctx, scenario, sim, collector, opts)
		sim.Reset()
		collector.Reset()
	}
}
