// Package balance provides automated balance validation and testing for Venture's gameplay systems.
//
// This package implements Phase 65.2 of V10.0 (Production Readiness Audit), providing statistical
// analysis and validation of game balance across 8 domains:
//  1. Combat Balance: Class win rates, weapon viability, enemy scaling, boss difficulty
//  2. Economic Balance: Loot value, crafting profitability, market pricing, gold sinks
//  3. Progression Balance: XP curves, skill unlocks, stat scaling, prestige content
//  4. Social Balance: Guild size optimization, territory control, trade limits, chat rates
//  5. Housing Balance: Build costs, upgrade progression, decoration limits, storage capacity
//  6. Vehicle Balance: Speed/durability trade-offs, fuel consumption, terrain compatibility
//  7. Companion Balance: Loyalty progression, skill learning rates, combat effectiveness
//  8. Quest Balance: Reward scaling, difficulty rating accuracy, completion times
//
// # Design Philosophy
//
// The balance framework uses statistical analysis and simulation rather than hardcoded thresholds.
// Each validator runs thousands of simulated scenarios and reports deviations from target distributions.
//
// # Core Interfaces
//
// BalanceValidator defines the contract for domain-specific validators:
//
//	type BalanceValidator interface {
//	    Validate(ctx context.Context) (*ValidationResult, error)
//	    GetDomain() string
//	}
//
// ValidationResult contains pass/fail status, metrics, and recommendations:
//
//	type ValidationResult struct {
//	    Domain       string
//	    Passed       bool
//	    Metrics      map[string]float64
//	    Issues       []string
//	    Recommendations []string
//	}
//
// # Usage Example
//
//	import "github.com/opd-ai/venture/pkg/balance"
//
//	// Create combat balance validator
//	validator := balance.NewCombatValidator(world, config)
//
//	// Run validation (simulates 1000+ combat scenarios)
//	result, err := validator.Validate(context.Background())
//	if err != nil {
//	    log.Fatalf("Validation failed: %v", err)
//	}
//
//	// Check results
//	if !result.Passed {
//	    fmt.Printf("Balance issues in %s:\n", result.Domain)
//	    for _, issue := range result.Issues {
//	        fmt.Printf("  - %s\n", issue)
//	    }
//	}
//
//	// Review metrics
//	fmt.Printf("Win rate variance: %.2f%%\n", result.Metrics["win_rate_variance"])
//
// # Statistical Methods
//
// Validators use statistical analysis to detect balance issues:
//   - Chi-square tests for distribution uniformity (e.g., weapon usage)
//   - T-tests for mean comparisons (e.g., average loot value vs. enemy difficulty)
//   - Regression analysis for scaling curves (e.g., XP requirements per level)
//   - Gini coefficients for inequality measurement (e.g., wealth distribution)
//
// # Acceptance Criteria (from ROADMAP_V10.md Phase 65.2)
//
// Combat Balance:
//   - Class balance: win rates within 45-55% in simulated PvP
//   - Weapon balance: all weapon types viable (usage >5%, <40%)
//   - Enemy scaling: difficulty curve smooth (R² >0.95)
//   - Boss difficulty: 60-80% simulated first-attempt failure rate
//
// Economic Balance:
//   - Loot value: matches enemy difficulty (correlation >0.8)
//   - Crafting profitability: 10-30% profit margin
//   - Market pricing: supply/demand curves realistic
//   - Gold sinks: prevent inflation (sink/source ratio 0.8-1.2)
//
// Progression Balance:
//   - XP curve: 1-2 hours per level (levels 1-20), 3-5 hours (21-50)
//   - Skill unlocks: meaningful every 5 levels (power increase 10-20%)
//   - Stat scaling: linear growth (R² >0.98), no power spikes
//   - Prestige balance: post-50 content requires 100+ hours
//
// Social Balance:
//   - Guild size: 10-100 members optimal (diminishing returns curve)
//   - Territory control: defender advantage 55-60%
//   - Trade limits: trust-based system prevents scams (<1% fraud rate)
//   - Chat rates: spam prevention without limiting normal use (reject rate <5%)
//
// # Performance
//
// Validation is computationally intensive but runs offline:
//   - Combat validation: ~30 seconds for 10,000 simulated battles
//   - Economic validation: ~20 seconds for 5,000 simulated transactions
//   - Progression validation: ~10 seconds for 1,000 simulated player journeys
//   - Total suite: ~5 minutes for all 8 domains
//
// # Determinism
//
// Validation uses deterministic simulation with fixed seeds for reproducibility.
// Test results are identical across platforms and runs for the same game version.
package balance
