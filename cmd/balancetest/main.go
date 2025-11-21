// Command balancetest runs Phase 65.2 balance validation tests.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/balance"
)

func main() {
	// Command-line flags
	domain := flag.String("domain", "all", "Balance domain to test (Combat, Economic, all)")
	simCount := flag.Int("simulations", 0, "Override simulation count (0 = use defaults)")
	seed := flag.Int64("seed", 12345, "Random seed for deterministic testing")
	verbose := flag.Bool("verbose", false, "Show detailed metrics")
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

	validators := make(map[string]balance.BalanceValidator)

	switch *domain {
	case "Combat":
		validators["Combat"] = balance.NewCombatValidator(config)
	case "Economic":
		validators["Economic"] = balance.NewEconomicValidator(config)
	case "all":
		validators["Combat"] = balance.NewCombatValidator(config)
		validators["Economic"] = balance.NewEconomicValidator(config)
	default:
		fmt.Fprintf(os.Stderr, "Unknown domain: %s\n", *domain)
		fmt.Fprintf(os.Stderr, "Valid domains: Combat, Economic, all\n")
		os.Exit(1)
	}

	ctx := context.Background()
	passed := 0
	failed := 0
	totalIssues := 0

	for name, validator := range validators {
		fmt.Printf("--- %s Balance ---\n", name)
		result, err := validator.Validate(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			failed++
			continue
		}

		// Print summary
		status := "✅ PASS"
		if !result.Passed {
			status = "❌ FAIL"
			failed++
		} else {
			passed++
		}

		fmt.Printf("Status: %s\n", status)
		fmt.Printf("Simulations: %d\n", result.SimulationCount)
		fmt.Printf("Duration: %.2fs\n", result.Duration)
		fmt.Printf("Issues: %d\n", len(result.Issues))
		fmt.Println()

		// Print metrics
		if *verbose {
			fmt.Println("Metrics:")
			for key, value := range result.Metrics {
				fmt.Printf("  %-35s %.4f\n", key+":", value)
			}
			fmt.Println()
		}

		// Print issues
		if len(result.Issues) > 0 {
			totalIssues += len(result.Issues)
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
	}

	// Final summary
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
