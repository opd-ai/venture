package stability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/errors"
)

// Config defines stability test parameters.
type Config struct {
	// Duration is the total test runtime (e.g., 72 hours for production validation)
	Duration time.Duration
	// CheckInterval is the time between health checks (default: 30 seconds)
	CheckInterval time.Duration
	// MemoryLimit is the maximum allowed memory usage in bytes (default: 500MB)
	MemoryLimit uint64
	// MinFPS is the minimum acceptable frames per second (default: 60)
	MinFPS float64
	// ReportPath is the output path for the stability report (default: stdout)
	ReportPath string
	// CrashThreshold is the maximum allowed crashes before test fails (default: 0)
	CrashThreshold int
	// MemoryLeakThreshold is the maximum allowed memory growth rate (bytes/second, default: 1KB/s)
	MemoryLeakThreshold float64
}

// DefaultConfig returns recommended stability test configuration for 72-hour validation.
func DefaultConfig() Config {
	return Config{
		Duration:            72 * time.Hour,
		CheckInterval:       30 * time.Second,
		MemoryLimit:         500 * 1024 * 1024, // 500MB
		MinFPS:              60.0,
		ReportPath:          "stability_report.json",
		CrashThreshold:      0,
		MemoryLeakThreshold: 1024.0, // 1KB/s
	}
}

// Report contains results from a stability test run.
type Report struct {
	// StartTime is when the test began
	StartTime time.Time
	// EndTime is when the test completed
	EndTime time.Time
	// TotalUptime is the duration the server was operational
	TotalUptime time.Duration
	// CrashCount is the number of detected crashes
	CrashCount int
	// MemoryLeakCount is the number of detected memory leaks
	MemoryLeakCount int
	// PerformanceDegradations is the number of FPS drops below threshold
	PerformanceDegradations int
	// AvgFPS is the average frames per second during test
	AvgFPS float64
	// PeakMemory is the maximum memory usage observed (bytes)
	PeakMemory uint64
	// AvgMemory is the average memory usage (bytes)
	AvgMemory uint64
	// Checks is the total number of health checks performed
	Checks int
	// Passed indicates whether all acceptance criteria were met
	Passed bool
	// FailureReason explains why the test failed (if applicable)
	FailureReason string
}

// FPSProvider is an interface for providing FPS metrics from the renderer.
type FPSProvider interface {
	// CurrentFPS returns the current frames per second
	CurrentFPS() float64
}

// defaultFPSProvider returns a constant 60 FPS for testing when no provider is set.
type defaultFPSProvider struct{}

func (d *defaultFPSProvider) CurrentFPS() float64 { return 60.0 }

// Monitor performs continuous stability monitoring.
type Monitor struct {
	config      Config
	mu          sync.RWMutex
	checks      []HealthCheck
	peakMem     uint64
	sumMem      uint64
	sumFPS      float64
	fpsCount    int
	running     bool
	cancel      context.CancelFunc
	fpsProvider FPSProvider
}

// HealthCheck represents a single health check snapshot.
type HealthCheck struct {
	Timestamp  time.Time
	Memory     uint64
	FPS        float64
	Goroutines int
}

// NewMonitor creates a stability monitor with the given configuration.
func NewMonitor(config Config) *Monitor {
	// Apply defaults for zero values
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.MemoryLimit == 0 {
		config.MemoryLimit = 500 * 1024 * 1024
	}
	if config.MinFPS == 0 {
		config.MinFPS = 60.0
	}
	if config.MemoryLeakThreshold == 0 {
		config.MemoryLeakThreshold = 1024.0
	}

	return &Monitor{
		config:      config,
		checks:      make([]HealthCheck, 0, 8640), // 72 hours at 30s intervals
		fpsProvider: &defaultFPSProvider{},
	}
}

// SetFPSProvider sets the FPS provider for obtaining actual FPS metrics.
// If not set, the monitor defaults to 60 FPS for testing purposes.
func (m *Monitor) SetFPSProvider(provider FPSProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fpsProvider = provider
}

// Run executes the stability test for the configured duration.
// The test runs independently of any specific server implementation.
func (m *Monitor) Run(ctx context.Context) (*Report, error) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil, fmt.Errorf("monitor already running")
	}

	m.running = true
	testCtx, cancel := context.WithTimeout(ctx, m.config.Duration)
	m.cancel = cancel
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.running = false
		m.cancel = nil
		m.mu.Unlock()
	}()

	// NOTE: time.Now() is intentionally used throughout this monitoring package
	// for wall-clock measurements. Stability monitoring requires real-time metrics
	// (memory, FPS, goroutines) at actual intervals, not deterministic simulation.
	// This is exempt from the deterministic procgen rules per audit guidelines.
	startTime := time.Now()
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-testCtx.Done():
			// Test duration completed
			return m.generateReport(startTime, time.Now()), nil
		case <-ticker.C:
			if err := m.performHealthCheck(); err != nil {
				return m.generateReport(startTime, time.Now()), fmt.Errorf("health check failed: %w", err)
			}
		}
	}
}

// Stop gracefully stops the stability monitor.
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
	}
	m.running = false
}

// performHealthCheck collects current system metrics.
func (m *Monitor) performHealthCheck() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	check := HealthCheck{
		Timestamp:  time.Now(),
		Memory:     getCurrentMemory(),
		FPS:        m.fpsProvider.CurrentFPS(),
		Goroutines: runtime.NumGoroutine(),
	}

	m.checks = append(m.checks, check)

	// Track peak memory
	if check.Memory > m.peakMem {
		m.peakMem = check.Memory
	}

	// Accumulate for averages
	m.sumMem += check.Memory
	m.sumFPS += check.FPS
	m.fpsCount++

	// Check for memory limit violation
	if check.Memory > m.config.MemoryLimit {
		return fmt.Errorf("memory limit exceeded: %d bytes > %d bytes", check.Memory, m.config.MemoryLimit)
	}

	return nil
}

// generateReport creates the final stability report.
func (m *Monitor) generateReport(start, end time.Time) *Report {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := &Report{
		StartTime:   start,
		EndTime:     end,
		TotalUptime: end.Sub(start),
		Checks:      len(m.checks),
		PeakMemory:  m.peakMem,
	}

	if m.fpsCount > 0 {
		report.AvgFPS = m.sumFPS / float64(m.fpsCount)
		report.AvgMemory = m.sumMem / uint64(m.fpsCount)
	}

	// Count performance degradations
	for _, check := range m.checks {
		if check.FPS < m.config.MinFPS {
			report.PerformanceDegradations++
		}
	}

	// Detect memory leaks (linear regression would be better, but simple check for now)
	// Only consider positive growth rates - negative rates indicate GC reclamation, not leaks
	if len(m.checks) >= 2 {
		timeDiff := m.checks[len(m.checks)-1].Timestamp.Sub(m.checks[0].Timestamp).Seconds()
		if timeDiff > 0 {
			memDiff := float64(m.checks[len(m.checks)-1].Memory) - float64(m.checks[0].Memory)
			growthRate := memDiff / timeDiff

			// Only report as leak if growth rate is positive and exceeds threshold
			if growthRate > 0 && growthRate > m.config.MemoryLeakThreshold {
				report.MemoryLeakCount++
				report.FailureReason = fmt.Sprintf("memory leak detected: %.2f bytes/sec growth rate", growthRate)
			}
		}
	}

	// Determine pass/fail
	report.Passed = report.CrashCount <= m.config.CrashThreshold &&
		report.MemoryLeakCount == 0 &&
		report.PeakMemory <= m.config.MemoryLimit &&
		report.AvgFPS >= m.config.MinFPS

	if !report.Passed && report.FailureReason == "" {
		report.FailureReason = "acceptance criteria not met"
	}

	return report
}

// getCurrentMemory returns current heap memory usage in bytes.
func getCurrentMemory() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem.HeapAlloc
}

// WriteReport writes the report to the configured ReportPath.
// If ReportPath is empty or "-", it writes to stdout.
func (m *Monitor) WriteReport(report *Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return errors.SerializationWrap(err, "failed to marshal stability report")
	}

	m.mu.RLock()
	reportPath := m.config.ReportPath
	m.mu.RUnlock()

	// Write to stdout if no path or "-" specified
	if reportPath == "" || reportPath == "-" {
		_, err = os.Stdout.Write(append(data, '\n'))
		if err != nil {
			return errors.FileSystemWrap(err, "failed to write stability report to stdout")
		}
		return nil
	}

	// Write to file
	if err := os.WriteFile(reportPath, data, 0o644); err != nil {
		return errors.FileSystemWrap(err, fmt.Sprintf("failed to write stability report to %s", reportPath))
	}

	return nil
}
