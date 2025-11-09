// Command benchmarkphases runs performance benchmarks for Phase 15-20 features.
//
// This tool executes comprehensive performance benchmarks for all visual enhancement
// phases and reports timing, memory usage, and target compliance.
//
// Usage:
//
//	benchmarkphases [options]
//
// Options:
//   -seed int
//       Random seed for generation (default: 12345)
//   -phase string
//       Run benchmarks for specific phase only (e.g., "Phase 15", "Phase 16.1")
//   -verbose
//       Enable verbose output
//   -profile-memory
//       Enable memory profiling
//
// Examples:
//
//	# Run all phase benchmarks
//	./benchmarkphases
//
//	# Run only Phase 15 benchmarks
//	./benchmarkphases -phase "Phase 15"
//
//	# Run with memory profiling
//	./benchmarkphases -profile-memory -verbose
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/visualtest"
)

func main() {
	// Parse flags
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	phaseFilter := flag.String("phase", "", "Run benchmarks for specific phase only")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	profileMemory := flag.Bool("profile-memory", false, "Enable memory profiling")

	flag.Parse()

	fmt.Println("=== Phase 15-20 Performance Benchmark Suite ===")
	fmt.Printf("Seed: %d\n", *seed)
	if *phaseFilter != "" {
		fmt.Printf("Phase Filter: %s\n", *phaseFilter)
	}
	fmt.Println()

	// Run benchmarks
	var suite *visualtest.BenchmarkSuite

	if *profileMemory {
		// Profile memory while running benchmarks
		profile := visualtest.ProfileFunction("AllBenchmarks", 1, func() {
			suite = visualtest.RunAllBenchmarks(*seed)
		})

		fmt.Println("\n=== Memory Profile ===")
		profile.PrintProfile()
		fmt.Println()
	} else {
		suite = visualtest.RunAllBenchmarks(*seed)
	}

	// Filter results if requested
	filteredResults := suite.Results
	if *phaseFilter != "" {
		filtered := []visualtest.BenchmarkResult{}
		for _, result := range suite.Results {
			if strings.HasPrefix(result.Phase, *phaseFilter) {
				filtered = append(filtered, result)
			}
		}
		filteredResults = filtered
	}

	// Print results
	if len(filteredResults) == 0 {
		fmt.Println("No benchmarks match the filter criteria")
		os.Exit(1)
	}

	// Create filtered suite for printing
	filteredSuite := &visualtest.BenchmarkSuite{
		Results:   filteredResults,
		StartTime: suite.StartTime,
		EndTime:   suite.EndTime,
		TotalTime: suite.TotalTime,
		Version:   suite.Version,
	}

	filteredSuite.PrintResults()

	// Print detailed analysis if verbose
	if *verbose {
		printDetailedAnalysis(filteredResults)
	}

	// Check if all targets met
	allPassed := true
	for _, result := range filteredResults {
		if !result.TargetMetNs || !result.TargetMetBytes {
			allPassed = false
			break
		}
	}

	// Exit with appropriate code
	if !allPassed {
		fmt.Println("\n⚠ Some benchmarks did not meet performance targets")
		os.Exit(1)
	}

	fmt.Println("\n✓ All benchmarks met performance targets")
	os.Exit(0)
}

// printDetailedAnalysis prints detailed benchmark analysis.
func printDetailedAnalysis(results []visualtest.BenchmarkResult) {
	fmt.Println("\n=== Detailed Analysis ===")

	// Group by phase
	byPhase := make(map[string][]visualtest.BenchmarkResult)
	for _, result := range results {
		byPhase[result.Phase] = append(byPhase[result.Phase], result)
	}

	for phase, phaseResults := range byPhase {
		fmt.Printf("\n%s:\n", phase)

		var totalNs int64
		var totalBytes int64
		passed := 0

		for _, result := range phaseResults {
			totalNs += result.NsPerOp
			totalBytes += result.BytesPerOp

			if result.TargetMetNs && result.TargetMetBytes {
				passed++
			}

			// Print individual result details
			status := "✓"
			if !result.TargetMetNs || !result.TargetMetBytes {
				status = "✗"
			}

			fmt.Printf("  %s %-30s: %10s/op, %8d bytes/op\n",
				status,
				truncate(result.Name, 30),
				formatDuration(result.NsPerOp),
				result.BytesPerOp,
			)
		}

		avgNs := totalNs / int64(len(phaseResults))
		avgBytes := totalBytes / int64(len(phaseResults))

		fmt.Printf("  Average: %s/op, %d bytes/op\n", formatDuration(avgNs), avgBytes)
		fmt.Printf("  Pass Rate: %d/%d (%.1f%%)\n",
			passed, len(phaseResults),
			float64(passed)*100.0/float64(len(phaseResults)))
	}
}

// formatDuration formats nanoseconds into a human-readable string.
func formatDuration(ns int64) string {
	if ns < 1000 {
		return fmt.Sprintf("%dns", ns)
	} else if ns < 1_000_000 {
		return fmt.Sprintf("%.2fµs", float64(ns)/1000.0)
	} else if ns < 1_000_000_000 {
		return fmt.Sprintf("%.2fms", float64(ns)/1_000_000.0)
	}
	return fmt.Sprintf("%.2fs", float64(ns)/1_000_000_000.0)
}

// truncate truncates a string to a maximum length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
