package modding

import (
	"path/filepath"
	"testing"
)

func TestSandbox_ValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		modsDir   string
		path      string
		wantError bool
	}{
		{
			name:      "valid path within mods directory",
			modsDir:   "testmods",
			path:      "testmods/mymod.json",
			wantError: false,
		},
		{
			name:      "path outside mods directory",
			modsDir:   "testmods",
			path:      "/etc/passwd",
			wantError: true,
		},
		{
			name:      "directory traversal attempt",
			modsDir:   "testmods",
			path:      "testmods/../etc/passwd",
			wantError: true,
		},
		{
			name:      "nested valid path",
			modsDir:   "testmods",
			path:      "testmods/category/mymod.json",
			wantError: false,
		},
		{
			name:      "absolute path outside",
			modsDir:   "testmods",
			path:      "/tmp/malicious.json",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSandboxConfig()
			config.ModsDirectory = tt.modsDir
			sandbox := NewSandboxWithConfig(config)

			err := sandbox.ValidatePath(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePath() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestSandbox_ValidateMod(t *testing.T) {
	sandbox := NewSandbox()

	tests := []struct {
		name       string
		mod        Mod
		wantValid  bool
		wantErrors int
	}{
		{
			name: "valid mod with allowed rules",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"difficulty":      2.0,
					"loot.drop_rate":  1.5,
					"spawn.frequency": 0.8,
				},
			},
			wantValid:  true,
			wantErrors: 0,
		},
		{
			name: "mod with disallowed rule name",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"system.execute": "command",
				},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "mod with script injection attempt",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"difficulty": "<script>alert('xss')</script>",
				},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "mod with eval injection attempt",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"combat.modifier": "eval(dangerous)",
				},
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "mod with nested valid rules",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"difficulty": map[string]interface{}{
						"base":       1.0,
						"multiplier": 1.5,
					},
				},
			},
			wantValid:  true,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sandbox.ValidateMod(&tt.mod)
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateMod() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("ValidateMod() errors = %d, want %d: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}
		})
	}
}

func TestSandbox_ValidateMod_ExcessiveRules(t *testing.T) {
	config := DefaultSandboxConfig()
	config.MaxRules = 5
	sandbox := NewSandboxWithConfig(config)

	mod := Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   make(map[string]interface{}),
	}

	// Add more rules than allowed
	for i := 0; i < 10; i++ {
		mod.Rules["difficulty"] = float64(i) // Using same allowed key
	}

	// Actually create unique keys for the test
	mod.Rules = map[string]interface{}{
		"difficulty":       1.0,
		"loot.drop_rate":   1.0,
		"spawn.rate":       1.0,
		"combat.damage":    1.0,
		"economy.gold":     1.0,
		"quest.difficulty": 1.0, // This exceeds the limit of 5
	}

	result := sandbox.ValidateMod(&mod)
	if result.Valid {
		t.Error("Expected validation to fail due to excessive rules")
	}

	foundResourceError := false
	for _, err := range result.Errors {
		if err.Check == "ResourceLimits" {
			foundResourceError = true
		}
	}
	if !foundResourceError {
		t.Error("Expected ResourceLimits error")
	}
}

func TestSandbox_ValidateMod_DeepNesting(t *testing.T) {
	config := DefaultSandboxConfig()
	config.MaxNestingDepth = 3
	sandbox := NewSandboxWithConfig(config)

	// Create deeply nested structure
	deepNested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"value": 1.0,
					},
				},
			},
		},
	}

	mod := Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   deepNested,
	}

	result := sandbox.ValidateMod(&mod)
	if result.Valid {
		t.Error("Expected validation to fail due to deep nesting")
	}
}

func TestSandbox_GenerateSecurityReport(t *testing.T) {
	sandbox := NewSandbox()
	report := sandbox.GenerateSecurityReport()

	// All checks should pass for the data-driven mod system
	if !report.AllChecksPassed() {
		t.Errorf("Expected all checks to pass, got: FileSystem=%v, Network=%v, Memory=%v, CPU=%v, API=%v, Code=%v",
			report.FileSystemIsolation, report.NetworkIsolation, report.MemoryLimits,
			report.CPULimits, report.APIRestrictions, report.CodeExecution)
	}

	if report.PassedCount() != 6 {
		t.Errorf("Expected 6 passing checks, got %d", report.PassedCount())
	}

	// Verify details are populated
	expectedDetails := []string{
		"FileSystemIsolation",
		"NetworkIsolation",
		"MemoryLimits",
		"CPULimits",
		"APIRestrictions",
		"CodeExecution",
	}

	for _, key := range expectedDetails {
		if report.Details[key] == "" {
			t.Errorf("Expected detail for %s to be populated", key)
		}
	}
}

func TestSandbox_validateStringValue(t *testing.T) {
	sandbox := NewSandbox()

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"safe string", "hello world", false},
		{"numeric string", "123.456", false},
		{"script tag", "<script>alert('xss')</script>", true},
		{"javascript url", "javascript:void(0)", true},
		{"eval call", "eval(code)", true},
		{"exec call", "exec(command)", true},
		{"python import", "__import__('os')", true},
		{"require call", "require('fs')", true},
		{"template literal", "${variable}", true},
		{"bash expansion", "$((1+1))", true},
		{"Function constructor", "new Function('return 1')", true},
		{"import call", "import('module')", true},
		{"normal text with eval word", "evaluate the results", false},
		{"safe json key", "combat_modifier", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sandbox.validateStringValue("test", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("validateStringValue(%q) error = %v, wantError %v", tt.value, err, tt.wantError)
			}
		})
	}
}

func TestSandbox_isAllowedRuleName(t *testing.T) {
	sandbox := NewSandbox()

	tests := []struct {
		name    string
		rule    string
		allowed bool
	}{
		{"difficulty base", "difficulty", true},
		{"difficulty modifier", "difficulty.modifier", true},
		{"loot drop rate", "loot.drop_rate", true},
		{"spawn rate", "spawn.rate", true},
		{"combat damage", "combat.damage", true},
		{"economy gold", "economy.gold", true},
		{"quest reward", "quest.reward", true},
		{"world size", "world.size", true},
		{"player health", "player.health", true},
		{"system execute", "system.execute", false},
		{"file read", "file.read", false},
		{"network request", "network.request", false},
		{"arbitrary key", "arbitrary_key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sandbox.isAllowedRuleName(tt.rule)
			if result != tt.allowed {
				t.Errorf("isAllowedRuleName(%q) = %v, want %v", tt.rule, result, tt.allowed)
			}
		})
	}
}

func TestSandboxError(t *testing.T) {
	err := SandboxError{
		Check:   "FileSystemIsolation",
		Message: "path outside mods directory",
	}

	expected := "[FileSystemIsolation] path outside mods directory"
	if err.Error() != expected {
		t.Errorf("SandboxError.Error() = %q, want %q", err.Error(), expected)
	}
}

func TestSecurityReport_PassedCount(t *testing.T) {
	tests := []struct {
		name     string
		report   SecurityReport
		expected int
	}{
		{
			name:     "all pass",
			report:   SecurityReport{true, true, true, true, true, true, nil},
			expected: 6,
		},
		{
			name:     "none pass",
			report:   SecurityReport{false, false, false, false, false, false, nil},
			expected: 0,
		},
		{
			name:     "some pass",
			report:   SecurityReport{true, false, true, false, true, false, nil},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.report.PassedCount(); got != tt.expected {
				t.Errorf("PassedCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDefaultSandboxConfig(t *testing.T) {
	config := DefaultSandboxConfig()

	if config.ModsDirectory != "mods" {
		t.Errorf("Expected ModsDirectory = 'mods', got %q", config.ModsDirectory)
	}

	if config.MaxModSizeBytes != 1024*1024 {
		t.Errorf("Expected MaxModSizeBytes = 1MB, got %d", config.MaxModSizeBytes)
	}

	if config.MaxRules != 100 {
		t.Errorf("Expected MaxRules = 100, got %d", config.MaxRules)
	}

	if config.MaxNestingDepth != 5 {
		t.Errorf("Expected MaxNestingDepth = 5, got %d", config.MaxNestingDepth)
	}

	if len(config.AllowedRulePatterns) == 0 {
		t.Error("Expected AllowedRulePatterns to be populated")
	}
}

func TestNewSandbox(t *testing.T) {
	sandbox := NewSandbox()
	if sandbox == nil {
		t.Error("NewSandbox() returned nil")
	}

	// Verify default config is applied
	if sandbox.config.ModsDirectory != "mods" {
		t.Errorf("Expected default ModsDirectory = 'mods', got %q", sandbox.config.ModsDirectory)
	}
}

func TestNewSandboxWithConfig(t *testing.T) {
	config := SandboxConfig{
		ModsDirectory:   "custom_mods",
		MaxModSizeBytes: 500000,
		MaxRules:        50,
		MaxNestingDepth: 3,
	}

	sandbox := NewSandboxWithConfig(config)

	if sandbox.config.ModsDirectory != "custom_mods" {
		t.Errorf("Expected ModsDirectory = 'custom_mods', got %q", sandbox.config.ModsDirectory)
	}
	if sandbox.config.MaxModSizeBytes != 500000 {
		t.Errorf("Expected MaxModSizeBytes = 500000, got %d", sandbox.config.MaxModSizeBytes)
	}
}

func BenchmarkSandbox_ValidateMod(b *testing.B) {
	sandbox := NewSandbox()
	mod := &Mod{
		ID:      "benchmark-mod",
		Name:    "Benchmark Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty":       2.0,
			"loot.drop_rate":   1.5,
			"spawn.frequency":  0.8,
			"combat.damage":    1.2,
			"economy.gold":     2.0,
			"quest.difficulty": 1.5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sandbox.ValidateMod(mod)
	}
}

func BenchmarkSandbox_ValidatePath(b *testing.B) {
	sandbox := NewSandbox()
	path := filepath.Join("mods", "testmod.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sandbox.ValidatePath(path)
	}
}

func BenchmarkSandbox_GenerateSecurityReport(b *testing.B) {
	sandbox := NewSandbox()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sandbox.GenerateSecurityReport()
	}
}
