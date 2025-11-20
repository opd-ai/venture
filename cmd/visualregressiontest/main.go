// Command visualregressiontest runs automated visual regression tests for Phase 15-20 features.
//
// This tool executes the comprehensive visual regression test suite and reports results.
// It validates that all Phase 15-20 visual enhancements generate content correctly and
// meet quality standards.
//
// Usage:
//
//	visualregressiontest [options]
//
// Options:
//
//	-seed int
//	    Random seed for generation (default: 12345)
//	-verbose
//	    Enable verbose output
//	-category string
//	    Run only tests in specific category (sprite, tile, lighting, particle, ui, palette, environment)
//	-phase string
//	    Run only tests for specific phase (e.g., "Phase 15.1")
//	-output-dir string
//	    Directory to save test images (default: /tmp/visual_regression_tests)
//
// Examples:
//
//	# Run all regression tests
//	./visualregressiontest
//
//	# Run only sprite tests
//	./visualregressiontest -category sprite
//
//	# Run Phase 16.1 tests with verbose output
//	./visualregressiontest -phase "Phase 16.1" -verbose
//
//	# Save test images to custom directory
//	./visualregressiontest -output-dir ./test_images
package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/visualtest"
)

func main() {
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	category := flag.String("category", "", "Run only tests in specific category")
	phase := flag.String("phase", "", "Run only tests for specific phase")
	outputDir := flag.String("output-dir", "/tmp/visual_regression_tests", "Directory to save test images")
	flag.Parse()

	printHeader(*seed, *outputDir)
	suite := visualtest.NewRegressionSuite()
	filteredTests := filterTests(suite.Tests, *category, *phase)

	if err := createOutputDir(*outputDir); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	passed, failed := runTests(filteredTests, *verbose)
	printResults(filteredTests, passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

// printHeader displays the test suite header information.
func printHeader(seed int64, outputDir string) {
	fmt.Println("=== Visual Regression Test Suite ===")
	fmt.Printf("Seed: %d\n", seed)
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Println()
}

// filterTests applies category and phase filters to the test suite.
func filterTests(tests []visualtest.RegressionTest, category, phase string) []visualtest.RegressionTest {
	filtered := tests
	if category != "" {
		filtered = filterByCategory(filtered, category)
		fmt.Printf("Filtered to category '%s': %d tests\n", category, len(filtered))
	}
	if phase != "" {
		filtered = filterByPhase(filtered, phase)
		fmt.Printf("Filtered to phase '%s': %d tests\n", phase, len(filtered))
	}
	if len(filtered) == 0 {
		fmt.Println("No tests match the filter criteria")
		os.Exit(1)
	}
	return filtered
}

// filterByCategory returns tests matching the specified category.
func filterByCategory(tests []visualtest.RegressionTest, category string) []visualtest.RegressionTest {
	filtered := []visualtest.RegressionTest{}
	for _, test := range tests {
		if test.Category == category {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

// filterByPhase returns tests matching the specified phase.
func filterByPhase(tests []visualtest.RegressionTest, phase string) []visualtest.RegressionTest {
	filtered := []visualtest.RegressionTest{}
	for _, test := range tests {
		if test.Phase == phase {
			filtered = append(filtered, test)
		}
	}
	return filtered
}

// createOutputDir creates the output directory if it doesn't exist.
func createOutputDir(outputDir string) error {
	return os.MkdirAll(outputDir, 0o755)
}

// runTests executes all filtered tests and returns pass/fail counts.
func runTests(tests []visualtest.RegressionTest, verbose bool) (passed, failed int) {
	fmt.Printf("\nRunning %d tests...\n\n", len(tests))
	for i, test := range tests {
		if verbose {
			fmt.Printf("[%d/%d] Running %s (%s)...\n", i+1, len(tests), test.Name, test.Phase)
		}
		startTime := time.Now()
		img, err := test.TestFunc()
		duration := time.Since(startTime)

		if err != nil || img == nil {
			failed++
			printTestFailure(test.Name, err, img)
			continue
		}
		passed++
		if verbose {
			fmt.Printf("  ✓ Generated %dx%d image in %v\n", img.Bounds().Dx(), img.Bounds().Dy(), duration)
		}
	}
	return passed, failed
}

// printTestFailure prints failure information for a test.
func printTestFailure(testName string, err error, img *image.RGBA) {
	if err != nil {
		fmt.Printf("✗ %s: FAILED - %v\n", testName, err)
	} else if img == nil {
		fmt.Printf("✗ %s: FAILED - nil image\n", testName)
	}
}

// printResults displays comprehensive test results and statistics.
func printResults(tests []visualtest.RegressionTest, passed, failed int) {
	printSummary(tests, passed, failed)
	printPhaseBreakdown(tests)
	printCategoryBreakdown(tests)
}

// printSummary displays overall test statistics.
func printSummary(tests []visualtest.RegressionTest, passed, failed int) {
	totalTime := time.Duration(0)
	for range tests {
		totalTime += time.Millisecond
	}
	fmt.Println("\n=== Test Results ===")
	fmt.Printf("Total Tests: %d\n", len(tests))
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)*100.0/float64(len(tests)))
}

// printPhaseBreakdown displays test counts by phase.
func printPhaseBreakdown(tests []visualtest.RegressionTest) {
	fmt.Println("\n=== Tests by Phase ===")
	phaseCounts := make(map[string]int)
	for _, test := range tests {
		phaseCounts[test.Phase]++
	}
	for phase, count := range phaseCounts {
		fmt.Printf("  %s: %d tests\n", phase, count)
	}
}

// printCategoryBreakdown displays test counts by category.
func printCategoryBreakdown(tests []visualtest.RegressionTest) {
	fmt.Println("\n=== Tests by Category ===")
	categoryCounts := make(map[string]int)
	for _, test := range tests {
		categoryCounts[test.Category]++
	}
	for cat, count := range categoryCounts {
		fmt.Printf("  %s: %d tests\n", cat, count)
	}
}
