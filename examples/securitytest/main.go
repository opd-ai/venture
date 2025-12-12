// Command securitytest runs the Phase 64.2 security audit framework.
//
// This tool executes all 30 security checks across 6 domains and generates
// a comprehensive report of the system's security posture.
//
// Usage:
//
//	securitytest                  Run full audit with summary
//	securitytest -verbose         Show all check details
//	securitytest -domain=<name>   Run checks for specific domain
//	securitytest -json            Output results as JSON
//
// Exit codes:
//
//	0: All checks passed
//	1: Some checks failed but no critical vulnerabilities
//	2: Critical vulnerabilities detected
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/security"
)

var (
	verbose = flag.Bool("verbose", false, "Show detailed check results")
	domain  = flag.String("domain", "", "Filter by security domain")
	jsonOut = flag.Bool("json", false, "Output results as JSON")
)

func main() {
	flag.Parse()

	// Run security audit
	auditor := security.NewAuditor()
	results := auditor.RunFullAudit()

	// Output results
	if *jsonOut {
		outputJSON(results)
	} else {
		outputText(results)
	}

	// Set exit code based on findings
	if results.HasCritical() {
		os.Exit(2)
	} else if !results.AllPassed() {
		os.Exit(1)
	}
	os.Exit(0)
}

func outputJSON(results *security.AuditResults) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(3)
	}
}

func outputText(results *security.AuditResults) {
	printSecuritySummary(results)
	domainChecks := groupChecksByDomain(results.Checks)
	printDomainResults(domainChecks)
	printRecommendations(results)
}

// printSecuritySummary prints the audit header and overall status.
func printSecuritySummary(results *security.AuditResults) {
	fmt.Println("=== Venture Security Audit - Phase 64.2 ===")
	fmt.Println()
	fmt.Println(results.Summary())
	fmt.Println()

	if results.AllPassed() {
		fmt.Println("✅ PASS - All security checks passed")
	} else if results.HasCritical() {
		fmt.Println("❌ CRITICAL - Critical vulnerabilities detected")
	} else {
		fmt.Println("⚠️  WARNING - Non-critical issues found")
	}
	fmt.Println()
}

// groupChecksByDomain organizes security checks by their domain.
func groupChecksByDomain(checks []security.SecurityCheck) map[string][]security.SecurityCheck {
	domainChecks := make(map[string][]security.SecurityCheck)
	for _, check := range checks {
		domainChecks[check.Domain] = append(domainChecks[check.Domain], check)
	}
	return domainChecks
}

// printDomainResults outputs results for each security domain.
func printDomainResults(domainChecks map[string][]security.SecurityCheck) {
	domains := []string{
		"Federation Security",
		"Chat & Encryption",
		"Mod Sandbox",
		"Input Validation",
		"Anti-Cheat",
		"Privacy",
	}

	for _, d := range domains {
		if *domain != "" && d != *domain {
			continue
		}

		checks := domainChecks[d]
		if len(checks) == 0 {
			continue
		}

		passed, _ := countPassedFailed(checks)
		fmt.Printf("--- %s (%d/%d passed) ---\n", d, passed, len(checks))

		if *verbose {
			printAllChecks(checks)
		} else {
			printFailures(checks)
		}
		fmt.Println()
	}
}

// countPassedFailed tallies passed and failed checks.
func countPassedFailed(checks []security.SecurityCheck) (passed, failed int) {
	for _, c := range checks {
		if c.Passed {
			passed++
		}
	}
	return passed, failed
}

// printAllChecks displays all security checks with details.
func printAllChecks(checks []security.SecurityCheck) {
	for _, check := range checks {
		status := formatCheckStatus(check)
		fmt.Printf("  %s %s\n", status, check.Name)
		if *verbose {
			fmt.Printf("     %s\n", check.Description)
			fmt.Printf("     %s\n", check.Message)
		}
	}
}

// formatCheckStatus returns a formatted status indicator for a check.
func formatCheckStatus(check security.SecurityCheck) string {
	if check.Passed {
		return "✅"
	}
	switch check.Severity {
	case security.SeverityCritical:
		return "❌ CRITICAL"
	case security.SeverityHigh:
		return "⚠️  HIGH"
	case security.SeverityMedium:
		return "⚠️  MEDIUM"
	case security.SeverityLow:
		return "⚠️  LOW"
	default:
		return "⚠️ "
	}
}

// printFailures displays only failed security checks.
func printFailures(checks []security.SecurityCheck) {
	for _, check := range checks {
		if !check.Passed {
			status := "⚠️ "
			if check.Severity == security.SeverityCritical {
				status = "❌"
			}
			fmt.Printf("  %s %s (%s): %s\n", status, check.Name, check.Severity, check.Message)
		}
	}
}

// printRecommendations outputs remediation suggestions for failed checks.
func printRecommendations(results *security.AuditResults) {
	if results.AllPassed() {
		return
	}

	fmt.Println("=== Recommendations ===")
	if results.CriticalCount > 0 {
		fmt.Printf("- Address %d critical vulnerabilities before release\n", results.CriticalCount)
	}
	if results.HighCount > 0 {
		fmt.Printf("- Fix %d high-severity issues\n", results.HighCount)
	}
	if results.MediumCount > 0 {
		fmt.Printf("- Consider addressing %d medium-severity issues\n", results.MediumCount)
	}

	if checkModSandboxFailed(results.Checks) {
		fmt.Println("- Mod sandbox security is not implemented (deferred to future release)")
		fmt.Println("  This is acceptable if mods are not enabled in v10.0")
	}
}

// checkModSandboxFailed determines if any mod sandbox checks failed.
func checkModSandboxFailed(checks []security.SecurityCheck) bool {
	for _, check := range checks {
		if check.Domain == "Mod Sandbox" && !check.Passed {
			return true
		}
	}
	return false
}
