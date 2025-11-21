package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/balance"
)

var (
	verbose      *bool
	domainFilter *string
)

func main() {
	config := parseFlags()
	validators := setupValidators(config)
	passed, failed, totalIssues := runValidations(validators)
	printFinalSummary(passed, failed, totalIssues)
}

// parseFlags processes command-line arguments and returns configured BalanceConfig.
func parseFlags() *balance.BalanceConfig {
	domain := flag.String("domain", "all", "Balance domain to test (Combat, Economic, all)")
	simCount := flag.Int("simulations", 0, "Override simulation count (0 = use defaults)")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic testing")
	verbose = flag.Bool("verbose", false, "Show detailed metrics")
	domainFilter = domain
	flag.Parse()

	config := balance.NewDefaultConfig()
	config.Seed = *seed

	if *simCount > 0 {
		for key := range config.SimulationCounts {
			config.SimulationCounts[key] = *simCount
		}
	}

	fmt.Println("=== Venture Phase 65.2: Balance & Progression Testing ===")
	fmt.Printf("Seed: %d\n", config.Seed)
	fmt.Println()

	return config
}

// setupValidators creates validators based on domain selection.
func setupValidators(config *balance.BalanceConfig) map[string]balance.BalanceValidator {
	validators := make(map[string]balance.BalanceValidator)
	domain := *domainFilter

	switch domain {
	case "Combat":
		validators["Combat"] = balance.NewCombatValidator(config)
	case "Economic":
		validators["Economic"] = balance.NewEconomicValidator(config)
	case "all":
		validators["Combat"] = balance.NewCombatValidator(config)
		validators["Economic"] = balance.NewEconomicValidator(config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown domain: %s\n", domain)
		fmt.Fprintf(os.Stderr, "Valid domains: Combat, Economic, all\n")
		os.Exit(1)
	}

	return validators
}

// runValidations executes all validators and prints results.
func runValidations(validators map[string]balance.BalanceValidator) (passed, failed, totalIssues int) {
	ctx := context.Background()

	for name, validator := range validators {
		fmt.Printf("--- %s Balance ---\n", name)
		result, err := validator.Validate(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			failed++
			continue
		}

		printValidationResult(result, &passed, &failed, &totalIssues)
	}
	return passed, failed, totalIssues
}

// printValidationResult displays validation result details.
func printValidationResult(result *balance.ValidationResult, passed, failed, totalIssues *int) {
	status := "✅ PASS"
	if !result.Passed {
		status = "❌ FAIL"
		*failed++
	} else {
		*passed++
	}

	fmt.Printf("Status: %s\n", status)
	fmt.Printf("Simulations: %d\n", result.SimulationCount)
	fmt.Printf("Duration: %.2fs\n", result.Duration)
	fmt.Printf("Issues: %d\n", len(result.Issues))
	fmt.Println()

	if *verbose {
		printMetrics(result.Metrics)
	}

	if len(result.Issues) > 0 {
		*totalIssues += len(result.Issues)
		printIssuesAndRecommendations(result)
	}
}

// printMetrics displays detailed validation metrics.
func printMetrics(metrics map[string]float64) {
	fmt.Println("Metrics:")
	for key, value := range metrics {
		fmt.Printf("  %-35s %.4f\n", key+":", value)
	}
	fmt.Println()
}

// printIssuesAndRecommendations displays detected issues and remediation suggestions.
func printIssuesAndRecommendations(result *balance.ValidationResult) {
	fmt.Println("Issues Detected:")
	for i, issue := range result.Issues {
		fmt.Printf("  %d. %s\n", i+1, issue)
	}
	fmt.Println()

	if len(result.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for i, rec := range result.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
		fmt.Println()
	}
}

// printFinalSummary outputs overall test results and exits.
func printFinalSummary(passed, failed, totalIssues int) {
	fmt.Println("=== Summary ===")
	fmt.Printf("Domains Tested: %d\n", passed+failed)
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total Issues: %d\n", totalIssues)

	if failed > 0 {
		fmt.Println("\n❌ Balance validation FAILED")
		fmt.Println("Review issues above and apply recommendations.")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ Balance validation PASSED")
		fmt.Println("All acceptance criteria met!")
	}
}
