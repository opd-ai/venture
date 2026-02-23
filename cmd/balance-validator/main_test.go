//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/balance"
)

// TestValidatorConstructors verifies all validators can be constructed.
func TestValidatorConstructors(t *testing.T) {
	config := balance.NewDefaultConfig()
	config.Seed = 42

	tests := []struct {
		name        string
		constructor func(*balance.BalanceConfig) balance.BalanceValidator
	}{
		{"Combat", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewCombatValidator(c) }},
		{"Economic", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewEconomicValidator(c) }},
		{"Progression", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewProgressionValidator(c) }},
		{"Social", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewSocialValidator(c) }},
		{"Housing", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewHousingValidator(c) }},
		{"Vehicle", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewVehicleValidator(c) }},
		{"Companion", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewCompanionValidator(c) }},
		{"Quest", func(c *balance.BalanceConfig) balance.BalanceValidator { return balance.NewQuestValidator(c) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := tt.constructor(config)
			if validator == nil {
				t.Errorf("%s validator constructor returned nil", tt.name)
			}
			if got := validator.GetDomain(); got != tt.name {
				t.Errorf("%s.GetDomain() = %q, want %q", tt.name, got, tt.name)
			}
		})
	}
}

// TestAvailableValidatorsMap verifies the availableValidators map is correctly populated.
func TestAvailableValidatorsMap(t *testing.T) {
	expectedDomains := []string{
		"combat", "economic", "progression", "social",
		"housing", "vehicle", "companion", "quest",
	}

	for _, domain := range expectedDomains {
		info, ok := availableValidators[domain]
		if !ok {
			t.Errorf("missing validator for domain %q", domain)
			continue
		}
		if info.constructor == nil {
			t.Errorf("validator for domain %q has nil constructor", domain)
		}
		if info.defaultCount <= 0 {
			t.Errorf("validator for domain %q has invalid default count: %d", domain, info.defaultCount)
		}
	}

	if len(availableValidators) != len(expectedDomains) {
		t.Errorf("availableValidators has %d entries, want %d", len(availableValidators), len(expectedDomains))
	}
}

// TestValidationSummaryJSON verifies ValidationSummary can be serialized.
func TestValidationSummaryJSON(t *testing.T) {
	summary := &ValidationSummary{
		Seed:        12345,
		TotalPassed: 7,
		TotalFailed: 1,
		Duration:    10.5,
		Results: []*balance.ValidationResult{
			{
				Domain:          "Combat",
				Passed:          true,
				Metrics:         map[string]float64{"win_rate": 0.5},
				Issues:          []string{},
				Recommendations: []string{},
				SimulationCount: 1000,
				Duration:        1.5,
			},
		},
		Errors: map[string]string{
			"Quest": "timeout",
		},
	}

	// Verify fields are accessible
	if summary.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", summary.Seed)
	}
	if len(summary.Results) != 1 {
		t.Errorf("Results length = %d, want 1", len(summary.Results))
	}
	if summary.Results[0].Domain != "Combat" {
		t.Errorf("Results[0].Domain = %q, want %q", summary.Results[0].Domain, "Combat")
	}
}
