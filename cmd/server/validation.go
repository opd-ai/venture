//go:build !android && !ios
// +build !android,!ios

// validation.go contains startup validation functions for security, balance, migration, and UX.
// Code relocated from: main.go
package main

import (
	"context"
	"time"

	"github.com/opd-ai/venture/pkg/balance"
	"github.com/opd-ai/venture/pkg/migration"
	"github.com/opd-ai/venture/pkg/security"
	"github.com/opd-ai/venture/pkg/ux"
	"github.com/sirupsen/logrus"
)

// runSecurityAudit performs security validation at server startup.
// Phase 2.4 (PLAN.md): Unconditional security package integration
func runSecurityAudit(serverLogger *logrus.Entry) {
	serverLogger.Info("running security audit at startup")
	auditor := security.NewAuditor(nil)
	results := auditor.RunFullAudit()

	serverLogger.WithFields(logrus.Fields{
		"total_checks": results.TotalChecks,
		"passed":       results.PassedChecks,
		"failed":       results.FailedChecks,
		"critical":     results.CriticalCount,
		"high":         results.HighCount,
		"pass_rate":    float64(results.PassedChecks) / float64(results.TotalChecks) * 100.0,
		"duration_ms":  results.EndTime.Sub(results.StartTime).Milliseconds(),
	}).Info("security audit completed")

	if results.HasCritical() {
		serverLogger.Warn("critical security vulnerabilities detected - review security audit results")
	}
}

// runBalanceValidation performs combat and economic balance validation at startup.
// Phase 6.1 (PLAN.md): Balance package integration for gameplay validation
func runBalanceValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running balance validation at startup")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	config := balance.NewDefaultConfig()
	config.Seed = *seed

	// Use reduced simulation counts for faster startup validation
	config.SimulationCounts["Combat"] = 1000
	config.SimulationCounts["Economic"] = 500

	// Run combat validation
	combatValidator := balance.NewCombatValidator(config)
	combatResult, err := combatValidator.Validate(ctx)
	if err != nil {
		serverLogger.WithError(err).Error("combat balance validation failed")
	} else {
		logBalanceResult(serverLogger, combatResult)
	}

	// Run economic validation
	economicValidator := balance.NewEconomicValidator(config)
	economicResult, err := economicValidator.Validate(ctx)
	if err != nil {
		serverLogger.WithError(err).Error("economic balance validation failed")
	} else {
		logBalanceResult(serverLogger, economicResult)
	}

	serverLogger.Info("balance validation completed")
}

// logBalanceResult logs the results of a balance validation run.
func logBalanceResult(serverLogger *logrus.Entry, result *balance.ValidationResult) {
	fields := logrus.Fields{
		"domain":           result.Domain,
		"passed":           result.Passed,
		"simulation_count": result.SimulationCount,
		"duration_sec":     result.Duration,
		"issue_count":      len(result.Issues),
	}

	// Add key metrics to log output
	for key, value := range result.Metrics {
		fields[key] = value
	}

	if result.Passed {
		serverLogger.WithFields(fields).Info("balance validation passed")
	} else {
		serverLogger.WithFields(fields).Warn("balance validation detected issues")
		for _, issue := range result.Issues {
			serverLogger.WithField("domain", result.Domain).Warn(issue)
		}
		for _, rec := range result.Recommendations {
			serverLogger.WithField("domain", result.Domain).Info("recommendation: " + rec)
		}
	}
}

// runMigrationValidation validates that all supported save file versions can be migrated.
// Phase 6.2 (PLAN.md): Migration package integration for save file compatibility
func runMigrationValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running migration validation at startup")

	config := migration.Config{
		TargetVersion: "v10.0",
		ValidateData:  true,
	}

	validator := migration.NewValidator(config)
	results, err := validator.ValidateAll()
	if err != nil {
		serverLogger.WithError(err).Error("migration validation failed")
		return
	}

	serverLogger.WithFields(logrus.Fields{
		"total_migrations": results.TotalCount,
		"passed":           results.PassedCount,
		"failed":           results.FailedCount,
		"pass_rate":        float64(results.PassedCount) / float64(results.TotalCount) * 100.0,
	}).Info("migration validation completed")

	if results.FailedCount > 0 {
		serverLogger.Warn("some migrations failed - save file compatibility may be affected")
		for _, m := range results.Migrations {
			if !m.Passed {
				serverLogger.WithFields(logrus.Fields{
					"source_version": m.SourceVersion,
					"target_version": m.TargetVersion,
					"error":          m.Error,
				}).Warn("migration failed")
			}
		}
	}
}

// runUXValidation performs user experience journey validation at startup.
// Phase 6.4 (PLAN.md): UX package integration for user journey validation
func runUXValidation(serverLogger *logrus.Entry) {
	serverLogger.Info("running UX journey validation at startup")

	config := ux.ValidationConfig{
		Runs:                 5,
		TimeTolerancePercent: 20.0,
		MinCompletionRate:    0.90,
		MinSatisfaction:      0.80,
		MaxErrorRate:         0.05,
	}

	validator := ux.NewJourneyValidatorWithConfig(config)
	results := validator.ValidateAll()
	summary := validator.GetSummary(results)

	serverLogger.WithFields(logrus.Fields{
		"total_journeys":   summary.TotalJourneys,
		"passed":           summary.PassedJourneys,
		"pass_rate":        summary.PassRate * 100.0,
		"avg_completion":   summary.AverageCompletionRate * 100.0,
		"avg_satisfaction": summary.AverageSatisfaction * 100.0,
		"avg_error_rate":   summary.AverageErrorRate * 100.0,
	}).Info("UX journey validation completed")

	if summary.PassedJourneys < summary.TotalJourneys {
		serverLogger.Warn("some user journeys failed validation - UX issues may exist")
		for _, result := range results {
			if !result.Passed {
				serverLogger.WithFields(logrus.Fields{
					"journey":         result.Name,
					"completion_rate": result.CompletionRate * 100.0,
					"satisfaction":    result.Satisfaction * 100.0,
					"error_rate":      result.ErrorRate * 100.0,
				}).Warn("journey failed")
			}
		}
	}
}
