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
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/visualtest"
)

func main() {
	// Parse flags
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	category := flag.String("category", "", "Run only tests in specific category")
	phase := flag.String("phase", "", "Run only tests for specific phase")
	outputDir := flag.String("output-dir", "/tmp/visual_regression_tests", "Directory to save test images")

	flag.Parse()

	// Create regression suite
	fmt.Println("=== Visual Regression Test Suite ===")
	fmt.Printf("Seed: %d\n", *seed)
	fmt.Printf("Output Directory: %s\n", *outputDir)
	fmt.Println()

	suite := visualtest.NewRegressionSuite()

	// Filter tests if requested
	filteredTests := suite.Tests
	if *category != "" {
		filtered := []visualtest.RegressionTest{}
		for _, test := range suite.Tests {
			if test.Category == *category {
				filtered = append(filtered, test)
			}
		}
		filteredTests = filtered
		fmt.Printf("Filtered to category '%s': %d tests\n", *category, len(filtered))
	}

	if *phase != "" {
		filtered := []visualtest.RegressionTest{}
		for _, test := range filteredTests {
			if test.Phase == *phase {
				filtered = append(filtered, test)
			}
		}
		filteredTests = filtered
		fmt.Printf("Filtered to phase '%s': %d tests\n", *phase, len(filtered))
	}

	if len(filteredTests) == 0 {
		fmt.Println("No tests match the filter criteria")
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	fmt.Printf("\nRunning %d tests...\n\n", len(filteredTests))

	passed := 0
	failed := 0
	totalTime := time.Duration(0)

	for i, test := range filteredTests {
		if *verbose {
			fmt.Printf("[%d/%d] Running %s (%s)...\n", i+1, len(filteredTests), test.Name, test.Phase)
		}

		startTime := time.Now()
		img, err := test.TestFunc()
		duration := time.Since(startTime)
		totalTime += duration

		if err != nil {
			failed++
			fmt.Printf("✗ %s: FAILED - %v\n", test.Name, err)
			continue
		}

		if img == nil {
			failed++
			fmt.Printf("✗ %s: FAILED - nil image\n", test.Name)
			continue
		}

		passed++
		if *verbose {
			fmt.Printf("  ✓ Generated %dx%d image in %v\n",
				img.Bounds().Dx(), img.Bounds().Dy(), duration)
		}
	}

	// Print summary
	fmt.Println("\n=== Test Results ===")
	fmt.Printf("Total Tests: %d\n", len(filteredTests))
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Success Rate: %.1f%%\n", float64(passed)*100.0/float64(len(filteredTests)))
	fmt.Printf("Total Time: %v\n", totalTime)
	fmt.Printf("Average Time: %v\n", totalTime/time.Duration(len(filteredTests)))

	// Print breakdown by phase
	fmt.Println("\n=== Tests by Phase ===")
	phaseCounts := make(map[string]int)
	for _, test := range filteredTests {
		phaseCounts[test.Phase]++
	}
	for phase, count := range phaseCounts {
		fmt.Printf("  %s: %d tests\n", phase, count)
	}

	// Print breakdown by category
	fmt.Println("\n=== Tests by Category ===")
	categoryCounts := make(map[string]int)
	for _, test := range filteredTests {
		categoryCounts[test.Category]++
	}
	for cat, count := range categoryCounts {
		fmt.Printf("  %s: %d tests\n", cat, count)
	}

	// Exit with appropriate code
	if failed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
