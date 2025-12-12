package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/ux"
)

func main() {
	// Command-line flags
	journeyName := flag.String("journey", "", "Specific journey to test (e.g., 'new_player', 'crafter', 'all')")
	runs := flag.Int("runs", 10, "Number of test runs per journey")
	verbose := flag.Bool("verbose", false, "Show detailed step-by-step results")
	listJourneys := flag.Bool("list", false, "List all available journeys")
	flag.Parse()

	// List journeys mode
	if *listJourneys {
		printJourneyList()
		return
	}

	// Configure validator
	config := ux.DefaultValidationConfig()
	config.Runs = *runs
	validator := ux.NewJourneyValidatorWithConfig(config)

	// Determine which journeys to run
	var results []ux.JourneyResult

	if *journeyName == "" || *journeyName == "all" {
		// Run all journeys
		fmt.Printf("Validating all 20 user journeys (%d runs each)...\n\n", *runs)
		results = validator.ValidateAll()
	} else {
		// Run specific journey
		journeyType := ux.JourneyType(*journeyName)
		fmt.Printf("Validating journey: %s (%d runs)...\n\n", *journeyName, *runs)
		result := validator.ValidateJourney(journeyType)
		results = []ux.JourneyResult{result}
	}

	// Display results
	for i, result := range results {
		if i > 0 {
			fmt.Println()
		}
		printJourneyResult(result, *verbose)
	}

	// Print summary if multiple journeys
	if len(results) > 1 {
		fmt.Println()
		printSummary(results)
	}

	// Exit with error code if any journey failed
	for _, result := range results {
		if !result.Passed {
			os.Exit(1)
		}
	}
}

func printJourneyList() {
	journeys := ux.AllJourneys()
	fmt.Println("Available User Journeys:")
	fmt.Println(strings.Repeat("=", 80))

	for _, journey := range journeys {
		fmt.Printf("\nJourney: %s\n", string(journey.Type))
		fmt.Printf("  Name: %s\n", journey.Name)
		fmt.Printf("  Description: %s\n", journey.Description)
		fmt.Printf("  Expected Duration: %v\n", journey.ExpectedDuration)
		fmt.Printf("  Steps: %d\n", len(journey.Steps))
		fmt.Printf("  Required Features: %s\n", strings.Join(journey.RequiredFeatures, ", "))
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("Total: %d journeys\n", len(journeys))
}

func printJourneyResult(result ux.JourneyResult, verbose bool) {
	// Header
	status := "✅ PASS"
	if !result.Passed {
		status = "❌ FAIL"
	}

	fmt.Printf("%s %s\n", status, result.Name)
	fmt.Println(strings.Repeat("-", 80))

	// Metrics
	fmt.Printf("Completion Rate:  %.1f%% (target: ≥90%%)\n", result.CompletionRate*100)
	fmt.Printf("Satisfaction:     %.1f%% (target: ≥80%%)\n", result.Satisfaction*100)
	fmt.Printf("Error Rate:       %.1f%% (target: ≤5%%)\n", result.ErrorRate*100)
	fmt.Printf("Average Duration: %v\n", formatDuration(result.AverageDuration))

	// Error if failed
	if result.Error != nil {
		fmt.Printf("Error: %v\n", result.Error)
	}

	// Detailed step results if verbose
	if verbose && len(result.Steps) > 0 {
		fmt.Println("\nStep Details:")
		for i, step := range result.Steps {
			stepStatus := "✅"
			if !step.Completed {
				stepStatus = "❌"
			}
			fmt.Printf("  %s Step %d: %s (%.2fs)\n",
				stepStatus, i+1, step.Name, step.Duration.Seconds())
			if step.Error != nil {
				fmt.Printf("    Error: %v\n", step.Error)
			}
		}
	}
}

func printSummary(results []ux.JourneyResult) {
	summary := ux.GetSummary(results)

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Total Journeys:    %d\n", summary.TotalJourneys)
	fmt.Printf("Passed:            %d (%.1f%%)\n",
		summary.PassedJourneys, summary.PassRate*100)
	fmt.Printf("Failed:            %d\n",
		summary.TotalJourneys-summary.PassedJourneys)
	fmt.Println()
	fmt.Printf("Avg Completion:    %.1f%%\n", summary.AverageCompletionRate*100)
	fmt.Printf("Avg Satisfaction:  %.1f%%\n", summary.AverageSatisfaction*100)
	fmt.Printf("Avg Error Rate:    %.1f%%\n", summary.AverageErrorRate*100)

	// Overall assessment
	fmt.Println()
	if summary.PassRate >= 0.90 {
		fmt.Println("✅ Overall Status: EXCELLENT - 90%+ journeys passing")
	} else if summary.PassRate >= 0.80 {
		fmt.Println("⚠️  Overall Status: GOOD - 80-89% journeys passing")
	} else if summary.PassRate >= 0.70 {
		fmt.Println("⚠️  Overall Status: NEEDS IMPROVEMENT - 70-79% journeys passing")
	} else {
		fmt.Println("❌ Overall Status: CRITICAL - <70% journeys passing")
	}
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "N/A"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
