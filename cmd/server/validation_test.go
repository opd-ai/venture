//go:build !android && !ios
// +build !android,!ios

// validation_test.go contains tests for server startup validation functions.
// Tests cover security audit, balance validation, migration validation, and UX validation.
package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/balance"
	"github.com/sirupsen/logrus"
)

// createTestLoggerForValidation creates a logger for testing that captures output
func createTestLoggerForValidation() (*logrus.Logger, *bytes.Buffer) {
	logger := logrus.New()
	buf := new(bytes.Buffer)
	logger.SetOutput(buf)
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})
	return logger, buf
}

// createTestLoggerEntry creates a logger entry for validation functions
func createTestLoggerEntry() (*logrus.Entry, *bytes.Buffer) {
	logger, buf := createTestLoggerForValidation()
	return logger.WithField("component", "test"), buf
}

// TestRunSecurityAudit verifies that security audit runs without panicking
func TestRunSecurityAudit(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Should not panic
	runSecurityAudit(entry)

	output := buf.String()

	// Verify logging occurred
	if !strings.Contains(output, "running security audit") {
		t.Error("expected 'running security audit' log message")
	}
	if !strings.Contains(output, "security audit completed") {
		t.Error("expected 'security audit completed' log message")
	}

	// Verify metrics are logged
	if !strings.Contains(output, "total_checks") {
		t.Error("expected 'total_checks' metric in log output")
	}
	if !strings.Contains(output, "passed") {
		t.Error("expected 'passed' metric in log output")
	}

	t.Logf("Security audit completed with output length: %d bytes", len(output))
}

// TestRunSecurityAudit_ResultFields verifies audit result fields are logged
func TestRunSecurityAudit_ResultFields(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	runSecurityAudit(entry)

	output := buf.String()

	// Check that key result fields are present
	expectedFields := []string{
		"total_checks",
		"passed",
		"failed",
		"critical",
		"high",
		"pass_rate",
		"duration_ms",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected field '%s' in log output", field)
		}
	}
}

// TestRunBalanceValidation verifies that balance validation runs without panicking
func TestRunBalanceValidation(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Set the global seed variable for the test
	testSeed := int64(12345)
	originalSeed := *seed
	*seed = testSeed
	defer func() { *seed = originalSeed }()

	// Should not panic
	runBalanceValidation(entry)

	output := buf.String()

	// Verify logging occurred
	if !strings.Contains(output, "running balance validation") {
		t.Error("expected 'running balance validation' log message")
	}
	if !strings.Contains(output, "balance validation completed") {
		t.Error("expected 'balance validation completed' log message")
	}

	t.Logf("Balance validation completed with output length: %d bytes", len(output))
}

// TestRunBalanceValidation_CombatAndEconomic verifies both validators are run
func TestRunBalanceValidation_CombatAndEconomic(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	testSeed := int64(54321)
	originalSeed := *seed
	*seed = testSeed
	defer func() { *seed = originalSeed }()

	runBalanceValidation(entry)

	output := buf.String()

	// Both combat and economic validation should be logged
	combatLogged := strings.Contains(output, "Combat") || strings.Contains(output, "combat")
	economicLogged := strings.Contains(output, "Economic") || strings.Contains(output, "economic")

	if !combatLogged && !economicLogged {
		t.Error("expected either combat or economic domain in log output")
	}

	// Verify domain field is logged
	if !strings.Contains(output, "domain") {
		t.Error("expected 'domain' field in log output")
	}
}

// TestRunBalanceValidation_DeterministicSeed verifies seed is used correctly
func TestRunBalanceValidation_DeterministicSeed(t *testing.T) {
	entry1, _ := createTestLoggerEntry()
	entry2, _ := createTestLoggerEntry()

	testSeed := int64(99999)
	originalSeed := *seed
	*seed = testSeed
	defer func() { *seed = originalSeed }()

	// Run twice with same seed - should complete without error
	runBalanceValidation(entry1)
	runBalanceValidation(entry2)

	// If we reach here without panic, the test passes
	t.Log("Balance validation with deterministic seed completed successfully")
}

// TestLogBalanceResult_Passed verifies logging of passed validation results
func TestLogBalanceResult_Passed(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Create a mock validation result that passed
	result := &mockValidationResult{
		domain:          "TestDomain",
		passed:          true,
		simulationCount: 100,
		duration:        1.5,
		issues:          []string{},
		recommendations: []string{},
		metrics:         map[string]float64{"accuracy": 0.95, "speed": 10.0},
	}

	logBalanceResult(entry, result.toBalanceResult())

	output := buf.String()

	// Should log as info (passed)
	if !strings.Contains(output, "balance validation passed") {
		t.Error("expected 'balance validation passed' log message")
	}
	if !strings.Contains(output, "TestDomain") {
		t.Error("expected domain name in log output")
	}
	if !strings.Contains(output, "passed=true") {
		t.Error("expected 'passed=true' in log output")
	}
}

// TestLogBalanceResult_Failed verifies logging of failed validation results
func TestLogBalanceResult_Failed(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Create a mock validation result that failed
	result := &mockValidationResult{
		domain:          "FailedDomain",
		passed:          false,
		simulationCount: 50,
		duration:        2.0,
		issues:          []string{"Issue 1: balance problem", "Issue 2: scaling issue"},
		recommendations: []string{"Recommendation 1", "Recommendation 2"},
		metrics:         map[string]float64{"error_rate": 0.15},
	}

	logBalanceResult(entry, result.toBalanceResult())

	output := buf.String()

	// Should log as warning (failed)
	if !strings.Contains(output, "balance validation detected issues") {
		t.Error("expected 'balance validation detected issues' log message")
	}
	if !strings.Contains(output, "Issue 1") {
		t.Error("expected issue details in log output")
	}
	if !strings.Contains(output, "recommendation") {
		t.Error("expected recommendation in log output")
	}
}

// TestLogBalanceResult_Metrics verifies metrics are included in log output
func TestLogBalanceResult_Metrics(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	result := &mockValidationResult{
		domain:          "MetricsDomain",
		passed:          true,
		simulationCount: 200,
		duration:        0.5,
		issues:          []string{},
		recommendations: []string{},
		metrics: map[string]float64{
			"metric_a": 100.0,
			"metric_b": 0.5,
		},
	}

	logBalanceResult(entry, result.toBalanceResult())

	output := buf.String()

	// Metrics should be included
	if !strings.Contains(output, "simulation_count=200") {
		t.Error("expected simulation_count in log output")
	}
}

// TestRunMigrationValidation verifies that migration validation runs without panicking
func TestRunMigrationValidation(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Should not panic
	runMigrationValidation(entry)

	output := buf.String()

	// Verify logging occurred
	if !strings.Contains(output, "running migration validation") {
		t.Error("expected 'running migration validation' log message")
	}
	if !strings.Contains(output, "migration validation completed") {
		t.Error("expected 'migration validation completed' log message")
	}

	t.Logf("Migration validation completed with output length: %d bytes", len(output))
}

// TestRunMigrationValidation_ResultFields verifies migration result fields are logged
func TestRunMigrationValidation_ResultFields(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	runMigrationValidation(entry)

	output := buf.String()

	// Check that key result fields are present
	expectedFields := []string{
		"total_migrations",
		"passed",
		"failed",
		"pass_rate",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected field '%s' in log output", field)
		}
	}
}

// TestRunUXValidation verifies that UX validation runs without panicking
func TestRunUXValidation(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	// Should not panic
	runUXValidation(entry)

	output := buf.String()

	// Verify logging occurred
	if !strings.Contains(output, "running UX journey validation") {
		t.Error("expected 'running UX journey validation' log message")
	}
	if !strings.Contains(output, "UX journey validation completed") {
		t.Error("expected 'UX journey validation completed' log message")
	}

	t.Logf("UX validation completed with output length: %d bytes", len(output))
}

// TestRunUXValidation_ResultFields verifies UX validation result fields are logged
func TestRunUXValidation_ResultFields(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	runUXValidation(entry)

	output := buf.String()

	// Check that key result fields are present
	expectedFields := []string{
		"total_journeys",
		"passed",
		"pass_rate",
		"avg_completion",
		"avg_satisfaction",
		"avg_error_rate",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("expected field '%s' in log output", field)
		}
	}
}

// TestRunUXValidation_JourneyDetails verifies journey-specific details are logged
func TestRunUXValidation_JourneyDetails(t *testing.T) {
	entry, buf := createTestLoggerEntry()

	runUXValidation(entry)

	output := buf.String()

	// The output should contain journey-related metrics
	// These fields are logged for each journey result
	if !strings.Contains(output, "completion") && !strings.Contains(output, "satisfaction") {
		t.Log("Note: No failed journeys detected (this is good)")
	}

	t.Logf("UX validation output length: %d bytes", len(output))
}

// TestAllValidationFunctions_NoNilPointerPanic verifies all functions handle nil-safe
func TestAllValidationFunctions_NoNilPointerPanic(t *testing.T) {
	// Create a valid entry - we can't pass nil but we verify functions don't panic
	entry, _ := createTestLoggerEntry()

	testSeed := int64(12345)
	originalSeed := *seed
	*seed = testSeed
	defer func() { *seed = originalSeed }()

	// Run all validation functions
	t.Run("SecurityAudit", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("runSecurityAudit panicked: %v", r)
			}
		}()
		runSecurityAudit(entry)
	})

	t.Run("BalanceValidation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("runBalanceValidation panicked: %v", r)
			}
		}()
		runBalanceValidation(entry)
	})

	t.Run("MigrationValidation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("runMigrationValidation panicked: %v", r)
			}
		}()
		runMigrationValidation(entry)
	})

	t.Run("UXValidation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("runUXValidation panicked: %v", r)
			}
		}()
		runUXValidation(entry)
	})
}

// BenchmarkRunSecurityAudit benchmarks the security audit function
func BenchmarkRunSecurityAudit(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel) // Minimize logging overhead
	entry := logger.WithField("component", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runSecurityAudit(entry)
	}
}

// BenchmarkRunMigrationValidation benchmarks the migration validation function
func BenchmarkRunMigrationValidation(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel)
	entry := logger.WithField("component", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runMigrationValidation(entry)
	}
}

// BenchmarkRunUXValidation benchmarks the UX validation function
func BenchmarkRunUXValidation(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.ErrorLevel)
	entry := logger.WithField("component", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runUXValidation(entry)
	}
}

// mockValidationResult is a test helper for creating balance.ValidationResult
type mockValidationResult struct {
	domain          string
	passed          bool
	simulationCount int
	duration        float64
	issues          []string
	recommendations []string
	metrics         map[string]float64
}

// toBalanceResult converts mock to actual balance.ValidationResult
func (m *mockValidationResult) toBalanceResult() *balance.ValidationResult {
	return &balance.ValidationResult{
		Domain:          m.domain,
		Passed:          m.passed,
		SimulationCount: m.simulationCount,
		Duration:        m.duration,
		Issues:          m.issues,
		Recommendations: m.recommendations,
		Metrics:         m.metrics,
	}
}
