//go:build !android && !ios
// +build !android,!ios

// Package main implements a standalone CLI tool for offline balance validation.
// This enables balance validation in CI/CD pipelines and during development
// without requiring a full server startup.
//
// Usage:
//
//	venture-balance-validator --seed 12345 --domain combat --simulations 10000
//	venture-balance-validator --domain all --json
//	venture-balance-validator --help
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/venture/pkg/balance"
	"github.com/sirupsen/logrus"
)

var (
	seed        = flag.Int64("seed", 12345, "Random seed for deterministic simulation")
	domain      = flag.String("domain", "all", "Balance domain to validate: combat, economic, progression, social, housing, vehicle, companion, quest, or all")
	simulations = flag.Int("simulations", 0, "Override default simulation count (0 = use domain defaults)")
	timeout     = flag.Duration("timeout", 5*time.Minute, "Maximum validation duration")
	jsonOutput  = flag.Bool("json", false, "Output results as JSON")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	quiet       = flag.Bool("quiet", false, "Suppress non-error output")
)

// validatorInfo holds validator constructor and default simulation count.
type validatorInfo struct {
	name         string
	constructor  func(*balance.BalanceConfig) balance.BalanceValidator
	defaultCount int
}

// availableValidators maps domain names to their validator constructors.
var availableValidators = map[string]validatorInfo{
	"combat":      {"Combat", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewCombatValidator(c) }, 10000},
	"economic":    {"Economic", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewEconomicValidator(c) }, 5000},
	"progression": {"Progression", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewProgressionValidator(c) }, 1000},
	"social":      {"Social", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewSocialValidator(c) }, 2000},
	"housing":     {"Housing", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewHousingValidator(c) }, 500},
	"vehicle":     {"Vehicle", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewVehicleValidator(c) }, 2000},
	"companion":   {"Companion", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewCompanionValidator(c) }, 1000},
	"quest":       {"Quest", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewQuestValidator(c) }, 2000},
}

// ValidationSummary holds aggregated results for JSON output.
type ValidationSummary struct {
	Seed        int64                       `json:"seed"`
	TotalPassed int                         `json:"total_passed"`
	TotalFailed int                         `json:"total_failed"`
	Duration    float64                     `json:"duration_seconds"`
	Results     []*balance.ValidationResult `json:"results"`
	Errors      map[string]string           `json:"errors,omitempty"`
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Balance validation tool for Venture game server.\n")
		fmt.Fprintf(os.Stderr, "Validates game balance across 8 domains using deterministic simulation.\n\n")
		fmt.Fprintf(os.Stderr, "Available domains: combat, economic, progression, social, housing, vehicle, companion, quest, all\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --domain combat --seed 42\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --domain all --json > balance-report.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --domain economic --simulations 1000 --verbose\n", os.Args[0])
	}
	flag.Parse()

	// Configure logging
	logger := logrus.New()
	if *quiet {
		logger.SetLevel(logrus.ErrorLevel)
	} else if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	if *jsonOutput {
		logger.SetLevel(logrus.ErrorLevel) // Suppress logging for clean JSON
	}
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Validate domain argument
	domainLower := strings.ToLower(*domain)
	if domainLower != "all" {
		if _, ok := availableValidators[domainLower]; !ok {
			fmt.Fprintf(os.Stderr, "Error: unknown domain %q\n", *domain)
			fmt.Fprintf(os.Stderr, "Available domains: combat, economic, progression, social, housing, vehicle, companion, quest, all\n")
			os.Exit(1)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Configure balance validation
	config := balance.NewDefaultConfig()
	config.Seed = *seed

	// Override simulation count if specified
	if *simulations > 0 {
		for domain := range config.SimulationCounts {
			config.SimulationCounts[domain] = *simulations
		}
	}

	// Run validation
	start := time.Now()
	summary := &ValidationSummary{
		Seed:    *seed,
		Results: make([]*balance.ValidationResult, 0),
		Errors:  make(map[string]string),
	}

	// Determine which validators to run
	var domainsToRun []string
	if domainLower == "all" {
		domainsToRun = []string{"combat", "economic", "progression", "social", "housing", "vehicle", "companion", "quest"}
	} else {
		domainsToRun = []string{domainLower}
	}

	// Run each validator
	for _, d := range domainsToRun {
		info := availableValidators[d]

		if !*quiet && !*jsonOutput {
			logger.WithFields(logrus.Fields{
				"domain": info.name,
				"seed":   *seed,
			}).Info("starting validation")
		}

		validator := info.constructor(config)
		result, err := validator.Validate(ctx)
		if err != nil {
			summary.Errors[info.name] = err.Error()
			summary.TotalFailed++
			if !*jsonOutput {
				logger.WithFields(logrus.Fields{
					"domain": info.name,
					"error":  err.Error(),
				}).Error("validation failed")
			}
			continue
		}

		summary.Results = append(summary.Results, result)
		if result.Passed {
			summary.TotalPassed++
		} else {
			summary.TotalFailed++
		}

		if !*quiet && !*jsonOutput {
			logResult(logger, result)
		}
	}

	summary.Duration = time.Since(start).Seconds()

	// Output results
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(summary); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
	} else if !*quiet {
		// NOTE: fmt.Printf usage below is intentional for CLI tool output.
		// CLI tools (cmd/*) are exempt from structured logging requirements.
		fmt.Println()
		fmt.Printf("=== Balance Validation Summary ===\n")
		fmt.Printf("Seed:      %d\n", summary.Seed)
		fmt.Printf("Passed:    %d/%d\n", summary.TotalPassed, summary.TotalPassed+summary.TotalFailed)
		fmt.Printf("Duration:  %.2fs\n", summary.Duration)
		fmt.Println()
	}

	// Exit with appropriate code
	if summary.TotalFailed > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

// logResult logs a single validation result.
func logResult(logger *logrus.Logger, result *balance.ValidationResult) {
	fields := logrus.Fields{
		"domain":      result.Domain,
		"passed":      result.Passed,
		"simulations": result.SimulationCount,
		"duration":    fmt.Sprintf("%.2fs", result.Duration),
	}

	if result.Passed {
		logger.WithFields(fields).Info("validation passed")
	} else {
		logger.WithFields(fields).Warn("validation failed")
		for _, issue := range result.Issues {
			logger.WithField("domain", result.Domain).Warn("  - " + issue)
		}
		for _, rec := range result.Recommendations {
			logger.WithField("domain", result.Domain).Info("  recommendation: " + rec)
		}
	}
}
