package security

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewAuditor(t *testing.T) {
	auditor := NewAuditor(nil)
	if auditor == nil {
		t.Fatal("NewAuditor returned nil")
	}
	if auditor.checks == nil {
		t.Error("checks slice not initialized")
	}
}

func TestRunFullAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := auditor.RunFullAudit()

	if results == nil {
		t.Fatal("RunFullAudit returned nil")
	}

	// Verify all 30 checks executed (6 domains)
	expectedChecks := 30 // 8 + 6 + 6 + 4 + 3 + 3
	if results.TotalChecks != expectedChecks {
		t.Errorf("Expected %d total checks, got %d", expectedChecks, results.TotalChecks)
	}

	// Verify time tracking
	if results.StartTime.IsZero() {
		t.Error("StartTime not set")
	}
	if results.EndTime.IsZero() {
		t.Error("EndTime not set")
	}
	if results.EndTime.Before(results.StartTime) {
		t.Error("EndTime before StartTime")
	}

	// Verify all domains covered
	domains := make(map[string]bool)
	for _, check := range results.Checks {
		domains[check.Domain] = true
	}

	expectedDomains := []string{
		"Federation Security",
		"Chat & Encryption",
		"Mod Sandbox",
		"Input Validation",
		"Anti-Cheat",
		"Privacy",
	}

	for _, domain := range expectedDomains {
		if !domains[domain] {
			t.Errorf("Missing domain: %s", domain)
		}
	}
}

func TestAuditResults_AllPassed(t *testing.T) {
	tests := []struct {
		name          string
		passed        int
		failed        int
		wantAllPassed bool
	}{
		{"all passed", 30, 0, true},
		{"some failed", 24, 6, false},
		{"all failed", 0, 30, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AuditResults{
				TotalChecks:  tt.passed + tt.failed,
				PassedChecks: tt.passed,
				FailedChecks: tt.failed,
			}
			if got := r.AllPassed(); got != tt.wantAllPassed {
				t.Errorf("AllPassed() = %v, want %v", got, tt.wantAllPassed)
			}
		})
	}
}

func TestAuditResults_HasCritical(t *testing.T) {
	tests := []struct {
		name          string
		criticalCount int
		want          bool
	}{
		{"no critical", 0, false},
		{"one critical", 1, true},
		{"multiple critical", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AuditResults{CriticalCount: tt.criticalCount}
			if got := r.HasCritical(); got != tt.want {
				t.Errorf("HasCritical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuditResults_Summary(t *testing.T) {
	r := &AuditResults{
		TotalChecks:   30,
		PassedChecks:  24,
		FailedChecks:  6,
		CriticalCount: 2,
		HighCount:     2,
		MediumCount:   1,
		LowCount:      1,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(100 * time.Millisecond),
	}

	summary := r.Summary()

	// Verify summary contains key information
	if !strings.Contains(summary, "24/30") {
		t.Error("Summary missing pass/total count")
	}
	if !strings.Contains(summary, "80.0%") {
		t.Error("Summary missing percentage")
	}
	if !strings.Contains(summary, "Critical: 2") {
		t.Error("Summary missing critical count")
	}
	if !strings.Contains(summary, "High: 2") {
		t.Error("Summary missing high count")
	}
}

func TestAuditResults_Summary_ZeroChecks(t *testing.T) {
	r := &AuditResults{
		TotalChecks:  0,
		PassedChecks: 0,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(10 * time.Millisecond),
	}

	summary := r.Summary()

	// Verify no NaN in output when TotalChecks is zero
	if strings.Contains(summary, "NaN") {
		t.Error("Summary should not contain NaN when TotalChecks is 0")
	}
	if !strings.Contains(summary, "0/0") {
		t.Error("Summary should show 0/0 checks")
	}
	if !strings.Contains(summary, "0.0%") {
		t.Error("Summary should show 0.0% when no checks")
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityInfo, "Info"},
		{SeverityLow, "Low"},
		{SeverityMedium, "Medium"},
		{SeverityHigh, "High"},
		{SeverityCritical, "Critical"},
		{Severity(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFederationSecurityAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditFederationSecurity(results)

	// Should have 8 checks
	if len(results.Checks) != 8 {
		t.Errorf("Expected 8 federation security checks, got %d", len(results.Checks))
	}

	// Verify all checks are in Federation Security domain
	for _, check := range results.Checks {
		if check.Domain != "Federation Security" {
			t.Errorf("Check %s has wrong domain: %s", check.Name, check.Domain)
		}

		// All federation checks should pass (implemented features)
		if !check.Passed {
			t.Errorf("Federation check failed: %s - %s", check.Name, check.Message)
		}
	}

	// Verify critical checks exist
	criticalChecks := []string{"Certificate Validation", "MITM Protection"}
	for _, name := range criticalChecks {
		found := false
		for _, check := range results.Checks {
			if check.Name == name && check.Severity == SeverityCritical {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing critical check: %s", name)
		}
	}
}

func TestChatEncryptionAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditChatEncryption(results)

	// Should have 6 checks
	if len(results.Checks) != 6 {
		t.Errorf("Expected 6 chat encryption checks, got %d", len(results.Checks))
	}

	// Verify E2E encryption check passes
	foundE2E := false
	for _, check := range results.Checks {
		if check.Name == "E2E Encryption" {
			foundE2E = true
			if !check.Passed {
				t.Error("E2E encryption check should pass")
			}
			if check.Severity != SeverityCritical {
				t.Error("E2E encryption should be critical severity")
			}
		}
	}
	if !foundE2E {
		t.Error("E2E encryption check not found")
	}
}

func TestModSandboxAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditModSandbox(results)

	// Should have 6 checks
	if len(results.Checks) != 6 {
		t.Errorf("Expected 6 mod sandbox checks, got %d", len(results.Checks))
	}

	// All mod sandbox checks should now pass (sandbox implemented)
	for _, check := range results.Checks {
		if !check.Passed {
			t.Errorf("Mod sandbox check should pass: %s (got: %s)", check.Name, check.Message)
		}
	}

	// Verify no critical failures
	criticalFails := 0
	for _, check := range results.Checks {
		if check.Severity == SeverityCritical && !check.Passed {
			criticalFails++
		}
	}
	if criticalFails != 0 {
		t.Errorf("Expected 0 critical mod sandbox failures, got %d", criticalFails)
	}
}

func TestInputValidationAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditInputValidation(results)

	// Should have 4 checks
	if len(results.Checks) != 4 {
		t.Errorf("Expected 4 input validation checks, got %d", len(results.Checks))
	}

	// Verify all checks passed (validation implemented)
	for _, check := range results.Checks {
		if !check.Passed {
			t.Errorf("Input validation check failed: %s - %s", check.Name, check.Message)
		}
	}
}

func TestAntiCheatAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditAntiCheat(results)

	// Should have 3 checks
	if len(results.Checks) != 3 {
		t.Errorf("Expected 3 anti-cheat checks, got %d", len(results.Checks))
	}

	// All anti-cheat checks should pass (systems operational)
	for _, check := range results.Checks {
		if !check.Passed {
			t.Errorf("Anti-cheat check failed: %s - %s", check.Name, check.Message)
		}
	}

	// Verify inventory integrity is critical
	for _, check := range results.Checks {
		if check.Name == "Inventory Integrity" && check.Severity != SeverityCritical {
			t.Error("Inventory integrity should be critical severity")
		}
	}
}

func TestPrivacyAudit(t *testing.T) {
	auditor := NewAuditor(nil)
	results := &AuditResults{Checks: make([]SecurityCheck, 0)}
	auditor.auditPrivacy(results)

	// Should have 3 checks
	if len(results.Checks) != 3 {
		t.Errorf("Expected 3 privacy checks, got %d", len(results.Checks))
	}

	// All privacy checks should pass
	for _, check := range results.Checks {
		if !check.Passed {
			t.Errorf("Privacy check failed: %s - %s", check.Name, check.Message)
		}
	}
}

func TestValidateIVRandomness(t *testing.T) {
	passed, msg := ValidateIVRandomness()
	if !passed {
		t.Errorf("IV randomness validation failed: %s", msg)
	}
	if !strings.Contains(msg, "100 unique IVs") {
		t.Errorf("Message should confirm 100 unique IVs, got: %s", msg)
	}
}

func TestValidateChatMessageSafety(t *testing.T) {
	passed, msg := ValidateChatMessageSafety()
	if !passed {
		t.Errorf("Chat message safety validation failed: %s", msg)
	}
	if msg == "" {
		t.Error("Validation message should not be empty")
	}
}

func TestValidateCoordinateBounds(t *testing.T) {
	passed, msg := ValidateCoordinateBounds()
	if !passed {
		t.Errorf("Coordinate bounds validation failed: %s", msg)
	}
	if !strings.Contains(msg, "1000000") {
		t.Errorf("Message should mention max coordinate, got: %s", msg)
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name string
		a    []byte
		b    []byte
		want bool
	}{
		{"equal", []byte("test"), []byte("test"), true},
		{"not equal", []byte("test"), []byte("best"), false},
		{"different length", []byte("test"), []byte("testing"), false},
		{"empty equal", []byte{}, []byte{}, true},
		{"one empty", []byte("test"), []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConstantTimeCompare(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ConstantTimeCompare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkRunFullAudit(b *testing.B) {
	auditor := NewAuditor(nil)
	for i := 0; i < b.N; i++ {
		_ = auditor.RunFullAudit()
	}
}

func BenchmarkValidateIVRandomness(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ValidateIVRandomness()
	}
}

func BenchmarkConstantTimeCompare(b *testing.B) {
	a := []byte("this is a test string for comparison")
	b_data := []byte("this is a test string for comparison")
	for i := 0; i < b.N; i++ {
		_ = ConstantTimeCompare(a, b_data)
	}
}

// TestNewAuditor_WithCustomLogger verifies that NewAuditor accepts a custom logger
func TestNewAuditor_WithCustomLogger(t *testing.T) {
	// Create a custom logger with specific settings
	customLogger := logrus.New()
	customLogger.SetLevel(logrus.ErrorLevel)
	customLogger.SetFormatter(&logrus.JSONFormatter{})

	// Create auditor with custom logger
	auditor := NewAuditor(customLogger)

	if auditor == nil {
		t.Fatal("NewAuditor returned nil with custom logger")
	}

	if auditor.logger != customLogger {
		t.Error("Auditor should use the provided custom logger")
	}

	// Run audit to ensure it works with custom logger
	results := auditor.RunFullAudit()

	if results == nil {
		t.Fatal("RunFullAudit returned nil with custom logger")
	}

	if results.TotalChecks != 30 {
		t.Errorf("Expected 30 checks with custom logger, got %d", results.TotalChecks)
	}
}

// TestNewAuditor_WithNilLogger verifies that NewAuditor uses package logger when nil is passed
func TestNewAuditor_WithNilLogger(t *testing.T) {
	// Create auditor with nil logger (should use package-level logger)
	auditor := NewAuditor(nil)

	if auditor == nil {
		t.Fatal("NewAuditor returned nil with nil logger")
	}

	if auditor.logger == nil {
		t.Error("Auditor logger should not be nil (should use package logger)")
	}

	// Run audit to ensure it works with package logger
	results := auditor.RunFullAudit()

	if results == nil {
		t.Fatal("RunFullAudit returned nil with package logger")
	}

	if results.TotalChecks != 30 {
		t.Errorf("Expected 30 checks with package logger, got %d", results.TotalChecks)
	}
}

// TestLoggerConfiguration_EnvironmentVariable verifies LOG_LEVEL env var is respected
func TestLoggerConfiguration_EnvironmentVariable(t *testing.T) {
	// Note: This test verifies the init() function logic but cannot actually
	// test environment variable changes since init() runs once per package.
	// The init() function sets the package logger level based on LOG_LEVEL.

	// Verify that the package logger exists and has a valid level
	if log == nil {
		t.Fatal("Package logger should be initialized")
	}

	// The actual level depends on the LOG_LEVEL environment variable
	// Default should be Info if LOG_LEVEL is not set
	level := log.GetLevel()

	validLevels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}

	found := false
	for _, validLevel := range validLevels {
		if level == validLevel {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Package logger has invalid level: %v", level)
	}
}
