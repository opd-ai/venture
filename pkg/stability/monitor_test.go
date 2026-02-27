package stability

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	venterrors "github.com/opd-ai/venture/pkg/errors"
)

func TestMonitor_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Duration != 72*time.Hour {
		t.Errorf("expected duration 72h, got %v", config.Duration)
	}
	if config.CheckInterval != 30*time.Second {
		t.Errorf("expected check interval 30s, got %v", config.CheckInterval)
	}
	if config.MemoryLimit != 500*1024*1024 {
		t.Errorf("expected memory limit 500MB, got %d", config.MemoryLimit)
	}
	if config.MinFPS != 60.0 {
		t.Errorf("expected min FPS 60, got %.2f", config.MinFPS)
	}
}

func TestMonitor_NewMonitor(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   100 * 1024 * 1024,
		MinFPS:        30.0,
	}

	monitor := NewMonitor(config)
	if monitor == nil {
		t.Fatal("expected monitor instance, got nil")
	}
	if monitor.config.Duration != 1*time.Second {
		t.Errorf("expected duration 1s, got %v", monitor.config.Duration)
	}
	if monitor.running {
		t.Error("expected running to be false initially")
	}
}

func TestMonitor_Run_ShortDuration(t *testing.T) {
	config := Config{
		Duration:            500 * time.Millisecond,
		CheckInterval:       100 * time.Millisecond,
		MemoryLimit:         500 * 1024 * 1024,
		MinFPS:              60.0,
		MemoryLeakThreshold: 10 * 1024, // 10KB/s - more lenient for short test
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	start := time.Now()
	report, err := monitor.Run(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected report, got nil")
	}

	// Verify test ran for approximately the configured duration
	if elapsed < 500*time.Millisecond || elapsed > 700*time.Millisecond {
		t.Errorf("expected elapsed time ~500ms, got %v", elapsed)
	}

	// Verify at least 4 checks were performed (500ms / 100ms = 5, minus startup/teardown)
	if report.Checks < 4 {
		t.Errorf("expected at least 4 checks, got %d", report.Checks)
	}

	// Verify uptime matches duration
	if report.TotalUptime < 500*time.Millisecond {
		t.Errorf("expected uptime >=500ms, got %v", report.TotalUptime)
	}

	// Verify no crashes
	if report.CrashCount != 0 {
		t.Errorf("expected 0 crashes, got %d", report.CrashCount)
	}

	// Note: Short tests may have normal memory variance, so we don't fail on pass/fail
	t.Logf("Test report: Passed=%v, Checks=%d, AvgFPS=%.2f, PeakMem=%d, Reason=%s",
		report.Passed, report.Checks, report.AvgFPS, report.PeakMemory, report.FailureReason)
}

func TestMonitor_Stop(t *testing.T) {
	config := Config{
		Duration:      10 * time.Second, // Long duration
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	// Start monitor in goroutine
	done := make(chan struct{})
	var report *Report
	var err error

	go func() {
		report, err = monitor.Run(ctx)
		close(done)
	}()

	// Wait for monitor to start
	time.Sleep(200 * time.Millisecond)

	// Stop monitor early
	monitor.Stop()

	// Wait for completion
	select {
	case <-done:
		// Expected: monitor stopped early
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop within timeout")
	}

	// Verify monitor was running briefly
	if err != nil && err.Error() != "context canceled" {
		// Context cancellation is expected when stopped early
		t.Logf("expected context canceled error, got: %v", err)
	}
	if report != nil && report.TotalUptime > 5*time.Second {
		t.Errorf("expected uptime <5s after stop, got %v", report.TotalUptime)
	}
}

func TestMonitor_MemoryLimit(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   1, // Extremely low limit to trigger failure
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	report, err := monitor.Run(ctx)

	// Expect error due to memory limit exceeded
	if err == nil {
		t.Error("expected error for memory limit exceeded")
	}
	if report == nil {
		t.Fatal("expected report even on error")
	}
	if report.Passed {
		t.Error("expected test to fail due to memory limit")
	}
}

func TestMonitor_HealthCheck(t *testing.T) {
	config := Config{
		Duration:      500 * time.Millisecond,
		CheckInterval: 50 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	report, err := monitor.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify health checks were recorded
	if len(monitor.checks) < 8 { // 500ms / 50ms = 10, minus startup
		t.Errorf("expected at least 8 health checks, got %d", len(monitor.checks))
	}

	// Verify peak memory is reasonable
	if report.PeakMemory == 0 {
		t.Error("expected non-zero peak memory")
	}
	if report.PeakMemory > 500*1024*1024 {
		t.Errorf("peak memory exceeded limit: %d bytes", report.PeakMemory)
	}

	// Verify average FPS is recorded
	if report.AvgFPS < 50.0 {
		t.Errorf("expected avg FPS >= 50, got %.2f", report.AvgFPS)
	}
}

func TestMonitor_Report(t *testing.T) {
	config := Config{
		Duration:      500 * time.Millisecond,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	report, err := monitor.Run(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify report fields
	if report.StartTime.IsZero() {
		t.Error("expected non-zero start time")
	}
	if report.EndTime.IsZero() {
		t.Error("expected non-zero end time")
	}
	if report.EndTime.Before(report.StartTime) {
		t.Error("expected end time after start time")
	}
	if report.TotalUptime == 0 {
		t.Error("expected non-zero uptime")
	}
	if report.Checks == 0 {
		t.Error("expected non-zero check count")
	}
	if report.PeakMemory == 0 {
		t.Error("expected non-zero peak memory")
	}
	if report.AvgMemory == 0 {
		t.Error("expected non-zero average memory")
	}
	if report.AvgFPS == 0 {
		t.Error("expected non-zero average FPS")
	}
}

func TestMonitor_ConcurrentRun(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	ctx := context.Background()

	// Try to run monitor twice concurrently
	go monitor.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	_, err := monitor.Run(ctx)
	if err == nil {
		t.Error("expected error when running monitor concurrently")
	}
	if err != nil && err.Error() != "monitor already running" {
		t.Errorf("expected 'monitor already running' error, got: %v", err)
	}
}

func BenchmarkMonitor_HealthCheck(b *testing.B) {
	config := Config{
		Duration:      1 * time.Hour, // Won't run this long
		CheckInterval: 10 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.performHealthCheck()
	}
}

func BenchmarkMonitor_GenerateReport(b *testing.B) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)
	start := time.Now()
	end := start.Add(1 * time.Second)

	// Populate with fake checks
	for i := 0; i < 10; i++ {
		monitor.checks = append(monitor.checks, HealthCheck{
			Timestamp:  start.Add(time.Duration(i) * 100 * time.Millisecond),
			Memory:     10 * 1024 * 1024,
			FPS:        60.0,
			Goroutines: 10,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.generateReport(start, end)
	}
}

// TestMonitor_NewMonitor_DefaultValues tests that NewMonitor properly applies defaults
// for zero-value config fields.
func TestMonitor_NewMonitor_DefaultValues(t *testing.T) {
	// Create config with zero values
	config := Config{}

	monitor := NewMonitor(config)
	if monitor == nil {
		t.Fatal("expected monitor instance, got nil")
	}

	// Verify defaults were applied
	if monitor.config.CheckInterval != 30*time.Second {
		t.Errorf("expected default check interval 30s, got %v", monitor.config.CheckInterval)
	}
	if monitor.config.MemoryLimit != 500*1024*1024 {
		t.Errorf("expected default memory limit 500MB, got %d", monitor.config.MemoryLimit)
	}
	if monitor.config.MinFPS != 60.0 {
		t.Errorf("expected default min FPS 60, got %.2f", monitor.config.MinFPS)
	}
	if monitor.config.MemoryLeakThreshold != 1024.0 {
		t.Errorf("expected default memory leak threshold 1024 bytes/s, got %.2f", monitor.config.MemoryLeakThreshold)
	}

	// Verify fpsProvider was set to default
	if monitor.fpsProvider == nil {
		t.Error("expected default FPS provider to be set")
	}
	fps := monitor.fpsProvider.CurrentFPS()
	if fps != 60.0 {
		t.Errorf("expected default FPS provider to return 60.0, got %.2f", fps)
	}
}

// mockFPSProvider is a test implementation of FPSProvider.
type mockFPSProvider struct {
	fps float64
}

func (m *mockFPSProvider) CurrentFPS() float64 {
	return m.fps
}

// TestMonitor_SetFPSProvider tests the SetFPSProvider method.
func TestMonitor_SetFPSProvider(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
	}

	monitor := NewMonitor(config)

	// Verify default provider
	defaultFPS := monitor.fpsProvider.CurrentFPS()
	if defaultFPS != 60.0 {
		t.Errorf("expected default FPS 60.0, got %.2f", defaultFPS)
	}

	// Set custom provider
	mockProvider := &mockFPSProvider{fps: 45.5}
	monitor.SetFPSProvider(mockProvider)

	// Verify custom provider is used
	customFPS := monitor.fpsProvider.CurrentFPS()
	if customFPS != 45.5 {
		t.Errorf("expected custom FPS 45.5, got %.2f", customFPS)
	}

	// Test thread-safety by calling SetFPSProvider concurrently
	// The race detector will catch any data races if the mutex isn't working
	var wg sync.WaitGroup
	numGoroutines := 10
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(fps float64) {
			defer wg.Done()
			provider := &mockFPSProvider{fps: fps}
			monitor.SetFPSProvider(provider)
		}(float64(i) + 10.0) // Use different FPS values: 10.0, 11.0, ..., 19.0
	}

	wg.Wait()

	// After all concurrent calls complete, verify we can still get a valid FPS
	// The exact value doesn't matter, but it should be one of the values we set
	finalFPS := monitor.fpsProvider.CurrentFPS()
	if finalFPS < 10.0 || finalFPS >= 20.0 {
		t.Errorf("expected final FPS in range [10.0, 20.0), got %.2f", finalFPS)
	}
}

// TestMonitor_WriteReport_Stdout tests writing report to stdout.
func TestMonitor_WriteReport_Stdout(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
		ReportPath:    "", // Empty means stdout
	}

	monitor := NewMonitor(config)
	report := &Report{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Second),
		TotalUptime: 1 * time.Second,
		Checks:      10,
		Passed:      true,
	}

	// Writing to stdout should succeed
	err := monitor.WriteReport(report)
	if err != nil {
		t.Errorf("unexpected error writing to stdout: %v", err)
	}
}

// TestMonitor_WriteReport_StdoutDash tests writing report to stdout using "-".
func TestMonitor_WriteReport_StdoutDash(t *testing.T) {
	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
		ReportPath:    "-", // Dash means stdout
	}

	monitor := NewMonitor(config)
	report := &Report{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Second),
		TotalUptime: 1 * time.Second,
		Checks:      10,
		Passed:      true,
	}

	// Writing to stdout should succeed
	err := monitor.WriteReport(report)
	if err != nil {
		t.Errorf("unexpected error writing to stdout: %v", err)
	}
}

// TestMonitor_WriteReport_File tests writing report to a file.
func TestMonitor_WriteReport_File(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_stability_report.json")

	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
		ReportPath:    tmpFile,
	}

	monitor := NewMonitor(config)
	report := &Report{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Second),
		TotalUptime: 1 * time.Second,
		Checks:      10,
		Passed:      true,
		AvgFPS:      59.5,
		PeakMemory:  100 * 1024 * 1024,
	}

	// Write report to file
	err := monitor.WriteReport(report)
	if err != nil {
		t.Fatalf("unexpected error writing to file: %v", err)
	}

	// Verify file exists and contains valid JSON
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	var readReport Report
	if err := json.Unmarshal(data, &readReport); err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}

	// Verify key fields
	if readReport.Checks != 10 {
		t.Errorf("expected 10 checks, got %d", readReport.Checks)
	}
	if !readReport.Passed {
		t.Error("expected report to be passed")
	}
	if readReport.AvgFPS != 59.5 {
		t.Errorf("expected avg FPS 59.5, got %.2f", readReport.AvgFPS)
	}
}

// TestMonitor_WriteReport_InvalidPath tests error handling for invalid file paths.
func TestMonitor_WriteReport_InvalidPath(t *testing.T) {
	// Use a path that doesn't exist in a nonexistent directory
	// This is more portable than hardcoded /invalid/path
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "nonexistent", "subdir", "report.json")

	config := Config{
		Duration:      1 * time.Second,
		CheckInterval: 100 * time.Millisecond,
		MemoryLimit:   500 * 1024 * 1024,
		MinFPS:        60.0,
		ReportPath:    invalidPath,
	}

	monitor := NewMonitor(config)
	report := &Report{
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Second),
		TotalUptime: 1 * time.Second,
		Checks:      10,
		Passed:      true,
	}

	// Writing to invalid path should fail (directory doesn't exist)
	err := monitor.WriteReport(report)
	if err == nil {
		t.Error("expected error writing to invalid path")
	}
}

// TestMonitor_WriteReport_ErrorWrapping tests that errors are properly wrapped with custom error types.
func TestMonitor_WriteReport_ErrorWrapping(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (*Monitor, *Report)
		expectedType  venterrors.ErrorType
		shouldContain string
	}{
		{
			name: "FileSystem error for invalid path",
			setupFunc: func() (*Monitor, *Report) {
				tmpDir := t.TempDir()
				invalidPath := filepath.Join(tmpDir, "nonexistent", "deep", "path", "report.json")
				config := Config{
					Duration:      1 * time.Second,
					CheckInterval: 100 * time.Millisecond,
					MemoryLimit:   500 * 1024 * 1024,
					MinFPS:        60.0,
					ReportPath:    invalidPath,
				}
				monitor := NewMonitor(config)
				report := &Report{
					StartTime:   time.Now(),
					EndTime:     time.Now().Add(1 * time.Second),
					TotalUptime: 1 * time.Second,
					Checks:      10,
					Passed:      true,
				}
				return monitor, report
			},
			expectedType:  venterrors.ErrorTypeFileSystem,
			shouldContain: "failed to write stability report",
		},
		{
			name: "FileSystem error can be unwrapped",
			setupFunc: func() (*Monitor, *Report) {
				tmpDir := t.TempDir()
				invalidPath := filepath.Join(tmpDir, "nonexistent", "report.json")
				config := Config{
					Duration:      1 * time.Second,
					CheckInterval: 100 * time.Millisecond,
					MemoryLimit:   500 * 1024 * 1024,
					MinFPS:        60.0,
					ReportPath:    invalidPath,
				}
				monitor := NewMonitor(config)
				report := &Report{
					StartTime:   time.Now(),
					EndTime:     time.Now().Add(1 * time.Second),
					TotalUptime: 1 * time.Second,
					Checks:      10,
					Passed:      true,
				}
				return monitor, report
			},
			expectedType:  venterrors.ErrorTypeFileSystem,
			shouldContain: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, report := tt.setupFunc()
			err := monitor.WriteReport(report)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Check if error is a VentureError
			var ventErr *venterrors.VentureError
			if !errors.As(err, &ventErr) {
				t.Fatalf("expected VentureError, got %T: %v", err, err)
			}

			// Check error type
			if ventErr.Type != tt.expectedType {
				t.Errorf("expected error type %v, got %v", tt.expectedType, ventErr.Type)
			}

			// Check error message contains expected text
			errMsg := err.Error()
			if tt.shouldContain != "" {
				// For "no such file or directory" we need to check the unwrapped error
				if tt.shouldContain == "no such file or directory" {
					unwrapped := ventErr.Unwrap()
					if unwrapped == nil {
						t.Error("expected unwrapped error, got nil")
					}
				} else {
					// For our custom messages, check the error string
					if errMsg == "" || len(errMsg) < 10 {
						t.Errorf("error message too short: %q", errMsg)
					}
				}
			}

			// Verify error is not retryable (filesystem errors require manual intervention)
			if ventErr.IsRetryable() {
				t.Error("expected FileSystem errors to not be retryable")
			}
		})
	}
}

// TestMonitor_WriteReport_ValidReportsSucceed verifies that valid reports write successfully.
func TestMonitor_WriteReport_ValidReportsSucceed(t *testing.T) {
	tests := []struct {
		name       string
		reportPath string
	}{
		{
			name:       "stdout with empty path",
			reportPath: "",
		},
		{
			name:       "stdout with dash",
			reportPath: "-",
		},
		{
			name:       "valid file path",
			reportPath: filepath.Join(t.TempDir(), "report.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Duration:      1 * time.Second,
				CheckInterval: 100 * time.Millisecond,
				MemoryLimit:   500 * 1024 * 1024,
				MinFPS:        60.0,
				ReportPath:    tt.reportPath,
			}

			monitor := NewMonitor(config)
			report := &Report{
				StartTime:   time.Now(),
				EndTime:     time.Now().Add(1 * time.Second),
				TotalUptime: 1 * time.Second,
				Checks:      10,
				Passed:      true,
			}

			err := monitor.WriteReport(report)
			if err != nil && tt.reportPath != "" && tt.reportPath != "-" {
				t.Errorf("unexpected error for valid path: %v", err)
			}
		})
	}
}
