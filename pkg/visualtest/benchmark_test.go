package visualtest

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestRunBenchmark tests the basic benchmark runner.
func TestRunBenchmark(t *testing.T) {
	counter := 0
	result := RunBenchmark("TestBenchmark", "Phase 15.1", 10, func() {
		counter++
	})

	if counter != 10 {
		t.Errorf("Expected 10 iterations, got %d", counter)
	}

	if result.Iterations != 10 {
		t.Errorf("Expected 10 iterations in result, got %d", result.Iterations)
	}

	if result.Name != "TestBenchmark" {
		t.Errorf("Expected name 'TestBenchmark', got %s", result.Name)
	}

	if result.Phase != "Phase 15.1" {
		t.Errorf("Expected phase 'Phase 15.1', got %s", result.Phase)
	}

	if result.NsPerOp <= 0 {
		t.Errorf("Expected positive ns/op, got %d", result.NsPerOp)
	}
}

// TestGetPhaseTargets verifies phase targets are defined.
func TestGetPhaseTargets(t *testing.T) {
	targets := GetPhaseTargets()

	expectedPhases := []string{
		"Phase 15.1", "Phase 15.2", "Phase 15.3",
		"Phase 16.1", "Phase 16.2", "Phase 16.3",
		"Phase 17.1", "Phase 17.2", "Phase 17.3",
		"Phase 18.1", "Phase 18.2", "Phase 18.3",
		"Phase 19.1", "Phase 19.2", "Phase 19.3",
		"Phase 20.1", "Phase 20.2",
	}

	for _, phase := range expectedPhases {
		target, exists := targets[phase]
		if !exists {
			t.Errorf("Missing target for %s", phase)
			continue
		}

		if target.MaxTimeNs <= 0 {
			t.Errorf("Invalid MaxTimeNs for %s: %d", phase, target.MaxTimeNs)
		}

		if target.MaxMemoryBytes < 0 {
			t.Errorf("Invalid MaxMemoryBytes for %s: %d", phase, target.MaxMemoryBytes)
		}
	}

	if len(targets) != len(expectedPhases) {
		t.Errorf("Expected %d phase targets, got %d", len(expectedPhases), len(targets))
	}
}

// TestBenchmarkPhase15Sprites tests Phase 15 sprite benchmarks.
func TestBenchmarkPhase15Sprites(t *testing.T) {
	results := BenchmarkPhase15Sprites(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}

		if result.Phase == "" {
			t.Errorf("Benchmark %s missing phase", result.Name)
		}

		if result.NsPerOp < 0 {
			t.Errorf("Benchmark %s has negative ns/op: %d", result.Name, result.NsPerOp)
		}
	}
}

// TestBenchmarkPhase16Tiles tests Phase 16 tile benchmarks.
func TestBenchmarkPhase16Tiles(t *testing.T) {
	results := BenchmarkPhase16Tiles(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}
	}
}

// TestBenchmarkPhase17Lighting tests Phase 17 lighting benchmarks.
func TestBenchmarkPhase17Lighting(t *testing.T) {
	results := BenchmarkPhase17Lighting(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}
	}
}

// TestBenchmarkPhase18Particles tests Phase 18 particle benchmarks.
func TestBenchmarkPhase18Particles(t *testing.T) {
	results := BenchmarkPhase18Particles(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}
	}
}

// TestBenchmarkPhase19UI tests Phase 19 UI benchmarks.
func TestBenchmarkPhase19UI(t *testing.T) {
	results := BenchmarkPhase19UI(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}
	}
}

// TestBenchmarkPhase20Environment tests Phase 20 environment benchmarks.
func TestBenchmarkPhase20Environment(t *testing.T) {
	results := BenchmarkPhase20Environment(12345)

	if len(results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	for _, result := range results {
		if result.Iterations == 0 {
			t.Errorf("Benchmark %s has 0 iterations", result.Name)
		}
	}
}

// TestRunAllBenchmarks tests the complete benchmark suite.
func TestRunAllBenchmarks(t *testing.T) {
	suite := RunAllBenchmarks(12345)

	if len(suite.Results) == 0 {
		t.Fatal("Expected benchmark results, got none")
	}

	if suite.StartTime.IsZero() {
		t.Error("StartTime not set")
	}

	if suite.EndTime.IsZero() {
		t.Error("EndTime not set")
	}

	if suite.TotalTime <= 0 {
		t.Error("TotalTime not calculated")
	}

	if suite.Version == "" {
		t.Error("Version not set")
	}

	// Check that we have results from all phases
	phaseCount := make(map[string]int)
	for _, result := range suite.Results {
		phaseCount[result.Phase]++
	}

	if len(phaseCount) == 0 {
		t.Error("No phases represented in results")
	}
}

// TestPrintResults verifies output formatting doesn't panic.
func TestPrintResults(t *testing.T) {
	suite := &BenchmarkSuite{
		Results: []BenchmarkResult{
			{
				Name:           "Test",
				Phase:          "Phase 15.1",
				NsPerOp:        1000000,
				BytesPerOp:     5000,
				AllocsPerOp:    10,
				TargetMetNs:    true,
				TargetMetBytes: true,
			},
		},
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintResults panicked: %v", r)
		}
	}()

	suite.PrintResults()
}

// TestFormatDuration tests duration formatting.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ns       int64
		expected string
	}{
		{500, "500ns"},
		{1500, "1.50µs"},
		{1500000, "1.50ms"},
		{1500000000, "1.50s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.ns)
		if result != tt.expected {
			t.Errorf("formatDuration(%d) = %s, want %s", tt.ns, result, tt.expected)
		}
	}
}

// TestTruncate tests string truncation.
func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exact length", 12, "exact length"},
		{"this is a very long string", 10, "this is..."},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

// BenchmarkPhase15SpriteBench is a standard Go benchmark for Phase 15.
func BenchmarkPhase15SpriteBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase15Sprites(12345)
	}
}

// BenchmarkPhase16TilesBench is a standard Go benchmark for Phase 16.
func BenchmarkPhase16TilesBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase16Tiles(12345)
	}
}

// BenchmarkPhase17LightingBench is a standard Go benchmark for Phase 17.
func BenchmarkPhase17LightingBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase17Lighting(12345)
	}
}

// BenchmarkPhase18ParticlesBench is a standard Go benchmark for Phase 18.
func BenchmarkPhase18ParticlesBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase18Particles(12345)
	}
}

// BenchmarkPhase19UIBench is a standard Go benchmark for Phase 19.
func BenchmarkPhase19UIBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase19UI(12345)
	}
}

// BenchmarkPhase20EnvironmentBench is a standard Go benchmark for Phase 20.
func BenchmarkPhase20EnvironmentBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BenchmarkPhase20Environment(12345)
	}
}

// BenchmarkRunAllBench benchmarks the entire suite.
func BenchmarkRunAllBench(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RunAllBenchmarks(12345)
	}
}

// TestPrintResultsTo tests PrintResultsTo writes to the specified writer.
func TestPrintResultsTo(t *testing.T) {
tests := []struct {
name         string
suite        *BenchmarkSuite
wantContains []string
wantNotEmpty bool
}{
{
name: "empty_suite",
suite: &BenchmarkSuite{
Results:   nil,
TotalTime: time.Second,
},
wantContains: []string{
"Performance Benchmark Results",
"Total Benchmarks: 0",
"Passed: 0/0",
},
wantNotEmpty: true,
},
{
name: "suite_with_passing_results",
suite: &BenchmarkSuite{
Results: []BenchmarkResult{
{
Name:           "TestBench1",
Phase:          "Phase 15",
NsPerOp:        1000,
BytesPerOp:     256,
AllocsPerOp:    2,
TargetMetNs:    true,
TargetMetBytes: true,
},
{
Name:           "TestBench2",
Phase:          "Phase 16",
NsPerOp:        2000,
BytesPerOp:     512,
AllocsPerOp:    4,
TargetMetNs:    true,
TargetMetBytes: true,
},
},
TotalTime: 2 * time.Second,
},
wantContains: []string{
"Performance Benchmark Results",
"Total Benchmarks: 2",
"TestBench1",
"TestBench2",
"Phase 15",
"Phase 16",
"Passed: 2/2 (100.0%)",
},
wantNotEmpty: true,
},
{
name: "suite_with_failing_results",
suite: &BenchmarkSuite{
Results: []BenchmarkResult{
{
Name:           "PassingTest",
Phase:          "Phase 17",
NsPerOp:        500,
BytesPerOp:     128,
AllocsPerOp:    1,
TargetMetNs:    true,
TargetMetBytes: true,
},
{
Name:           "FailingTest",
Phase:          "Phase 18",
NsPerOp:        5000000,
BytesPerOp:     10000,
AllocsPerOp:    100,
TargetMetNs:    false,
TargetMetBytes: false,
},
},
TotalTime: time.Second,
},
wantContains: []string{
"Total Benchmarks: 2",
"PassingTest",
"FailingTest",
"Passed: 1/2 (50.0%)",
},
wantNotEmpty: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
var buf bytes.Buffer
tt.suite.PrintResultsTo(&buf)
output := buf.String()

if tt.wantNotEmpty && output == "" {
t.Error("Expected non-empty output")
}

for _, want := range tt.wantContains {
if !strings.Contains(output, want) {
t.Errorf("Output should contain %q, got:\n%s", want, output)
}
}
})
}
}
