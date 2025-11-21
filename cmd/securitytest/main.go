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
	// Print summary
	fmt.Println("=== Venture Security Audit - Phase 64.2 ===")
	fmt.Println()
	fmt.Println(results.Summary())
	fmt.Println()

	// Print status indicator
	if results.AllPassed() {
		fmt.Println("✅ PASS - All security checks passed")
	} else if results.HasCritical() {
		fmt.Println("❌ CRITICAL - Critical vulnerabilities detected")
	} else {
		fmt.Println("⚠️  WARNING - Non-critical issues found")
	}
	fmt.Println()

	// Group checks by domain
	domainChecks := make(map[string][]security.SecurityCheck)
	for _, check := range results.Checks {
		domainChecks[check.Domain] = append(domainChecks[check.Domain], check)
	}

	// Print results by domain
	domains := []string{
		"Federation Security",
		"Chat & Encryption",
		"Mod Sandbox",
		"Input Validation",
		"Anti-Cheat",
		"Privacy",
	}

	for _, d := range domains {
		// Skip domain if filter specified and doesn't match
		if *domain != "" && d != *domain {
			continue
		}

		checks := domainChecks[d]
		if len(checks) == 0 {
			continue
		}

		// Count passed/failed
		passed := 0
		failed := 0
		for _, c := range checks {
			if c.Passed {
				passed++
			} else {
				failed++
			}
		}

		// Print domain header
		fmt.Printf("--- %s (%d/%d passed) ---\n", d, passed, len(checks))

		if *verbose {
			// Print all checks
			for _, check := range checks {
				status := "✅"
				if !check.Passed {
					switch check.Severity {
					case security.SeverityCritical:
						status = "❌ CRITICAL"
					case security.SeverityHigh:
						status = "⚠️  HIGH"
					case security.SeverityMedium:
						status = "⚠️  MEDIUM"
					case security.SeverityLow:
						status = "⚠️  LOW"
					}
				}
				fmt.Printf("  %s %s\n", status, check.Name)
				if *verbose {
					fmt.Printf("     %s\n", check.Description)
					fmt.Printf("     %s\n", check.Message)
				}
			}
		} else {
			// Just show failures
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
		fmt.Println()
	}

	// Print recommendations if failures exist
	if !results.AllPassed() {
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

		// Specific recommendations for mod sandbox
		modSandboxFailed := false
		for _, check := range results.Checks {
			if check.Domain == "Mod Sandbox" && !check.Passed {
				modSandboxFailed = true
				break
			}
		}
		if modSandboxFailed {
			fmt.Println("- Mod sandbox security is not implemented (deferred to future release)")
			fmt.Println("  This is acceptable if mods are not enabled in v10.0")
		}
	}
}
