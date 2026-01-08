package modding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMod_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mod     Mod
		wantErr bool
	}{
		{
			name: "valid rule mod",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules:   map[string]interface{}{"difficulty": 2.0},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			mod: Mod{
				Name:    "Test Mod",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			mod: Mod{
				ID:      "test-mod",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "empty version",
			mod: Mod{
				ID:   "test-mod",
				Name: "Test Mod",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "nil rule value",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules:   map[string]interface{}{"difficulty": nil},
			},
			wantErr: true,
		},
		{
			name: "default type",
			mod: Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mod.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Mod.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoader_LoadFromFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test mod file
	testMod := Mod{
		ID:          "test-mod",
		Name:        "Test Mod",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "A test mod",
		Type:        ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
			"permadeath": true,
		},
		Enabled: true,
	}

	modPath := filepath.Join(tmpDir, "test.json")
	data, err := json.Marshal(testMod)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(modPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Create loader
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false // Disable for testing
	loader := NewLoaderWithConfig(config)

	// Test loading
	mod, err := loader.LoadFromFile(modPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if mod.ID != testMod.ID {
		t.Errorf("ID = %v, want %v", mod.ID, testMod.ID)
	}

	if mod.Name != testMod.Name {
		t.Errorf("Name = %v, want %v", mod.Name, testMod.Name)
	}

	if mod.LoadedAt.IsZero() {
		t.Error("LoadedAt should be set")
	}
}

func TestLoader_LoadFromFile_SandboxViolation(t *testing.T) {
	// Create loader with sandbox enabled
	config := DefaultConfig()
	config.ModsDirectory = "/tmp/mods"
	config.EnableSandbox = true
	loader := NewLoaderWithConfig(config)

	// Try to load from outside mods directory
	_, err := loader.LoadFromFile("/etc/passwd")
	if err == nil {
		t.Error("Expected sandbox violation error")
	}
}

func TestLoader_LoadAll(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test mod files
	mods := []Mod{
		{
			ID:      "mod1",
			Name:    "Mod 1",
			Version: "1.0.0",
			Type:    ModTypeRule,
			Enabled: true,
		},
		{
			ID:      "mod2",
			Name:    "Mod 2",
			Version: "1.0.0",
			Type:    ModTypeGenerator,
			Enabled: true,
		},
	}

	for _, mod := range mods {
		data, err := json.Marshal(mod)
		if err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(tmpDir, mod.ID+".json")
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create loader
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	// Test loading all
	loadedMods, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(loadedMods) != len(mods) {
		t.Errorf("LoadAll() loaded %d mods, want %d", len(loadedMods), len(mods))
	}
}

func TestLoader_SaveToFile(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test mod
	mod := &Mod{
		ID:          "test-mod",
		Name:        "Test Mod",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "A test mod",
		Type:        ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
		},
		Enabled: true,
	}

	// Create loader
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	// Save mod
	modPath := filepath.Join(tmpDir, "test.json")
	if err := loader.SaveToFile(mod, modPath); err != nil {
		t.Fatalf("SaveToFile() error = %v", err)
	}

	// Load and verify
	loadedMod, err := loader.LoadFromFile(modPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	if loadedMod.ID != mod.ID {
		t.Errorf("ID = %v, want %v", loadedMod.ID, mod.ID)
	}
}

func TestManager_AddMod(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod() error = %v", err)
	}

	// Verify mod was added
	loadedMod, exists := manager.GetMod("test-mod")
	if !exists {
		t.Fatal("Mod not found after adding")
	}

	if loadedMod.ID != mod.ID {
		t.Errorf("ID = %v, want %v", loadedMod.ID, mod.ID)
	}
}

func TestManager_AddMod_Duplicate(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatalf("AddMod() error = %v", err)
	}

	// Try to add again
	err := manager.AddMod(mod)
	if err == nil {
		t.Error("Expected error for duplicate mod")
	}
}

func TestManager_AddMod_MissingDependency(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:           "test-mod",
		Name:         "Test Mod",
		Version:      "1.0.0",
		Type:         ModTypeRule,
		Dependencies: []string{"missing-mod"},
		Enabled:      true,
	}

	err := manager.AddMod(mod)
	if err == nil {
		t.Error("Expected error for missing dependency")
	}
}

func TestManager_RemoveMod(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.RemoveMod("test-mod"); err != nil {
		t.Fatalf("RemoveMod() error = %v", err)
	}

	// Verify mod was removed
	_, exists := manager.GetMod("test-mod")
	if exists {
		t.Error("Mod still exists after removal")
	}
}

func TestManager_RemoveMod_WithDependents(t *testing.T) {
	manager := NewManager()

	baseMod := &Mod{
		ID:      "base-mod",
		Name:    "Base Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true,
	}

	dependentMod := &Mod{
		ID:           "dependent-mod",
		Name:         "Dependent Mod",
		Version:      "1.0.0",
		Type:         ModTypeRule,
		Dependencies: []string{"base-mod"},
		Enabled:      true,
	}

	if err := manager.AddMod(baseMod); err != nil {
		t.Fatal(err)
	}

	if err := manager.AddMod(dependentMod); err != nil {
		t.Fatal(err)
	}

	// Try to remove base mod
	err := manager.RemoveMod("base-mod")
	if err == nil {
		t.Error("Expected error when removing mod with dependents")
	}
}

func TestManager_ApplyRules(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
			"permadeath": true,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatalf("ApplyRules() error = %v", err)
	}

	// Verify rules were applied
	difficulty, exists := manager.GetRule("difficulty")
	if !exists {
		t.Error("Difficulty rule not found")
	}

	if diffVal, ok := difficulty.(float64); !ok || diffVal != 2.0 {
		t.Errorf("Difficulty = %v, want 2.0", difficulty)
	}
}

func TestManager_GetRuleFloat64(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
			"spawn_rate": 1.5,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	// Test existing rule
	if val := manager.GetRuleFloat64("difficulty", 1.0); val != 2.0 {
		t.Errorf("GetRuleFloat64(difficulty) = %v, want 2.0", val)
	}

	// Test non-existent rule (should return default)
	if val := manager.GetRuleFloat64("nonexistent", 1.0); val != 1.0 {
		t.Errorf("GetRuleFloat64(nonexistent) = %v, want 1.0", val)
	}
}

func TestManager_GetRuleBool(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"permadeath": true,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	// Test existing rule
	if val := manager.GetRuleBool("permadeath", false); val != true {
		t.Errorf("GetRuleBool(permadeath) = %v, want true", val)
	}

	// Test non-existent rule (should return default)
	if val := manager.GetRuleBool("nonexistent", false); val != false {
		t.Errorf("GetRuleBool(nonexistent) = %v, want false", val)
	}
}

func TestManager_RateLimit(t *testing.T) {
	config := DefaultConfig()
	config.RuleChangeRateLimit = 2.0 // Low limit for testing
	manager := NewManagerWithConfig(config)

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"test": 1},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Apply rules twice (within limit)
	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	// Third call should exceed limit
	err := manager.ApplyRules()
	if err == nil {
		t.Error("Expected rate limit error")
	}

	// Wait for rate limit to reset
	time.Sleep(1100 * time.Millisecond)

	// Should work again
	if err := manager.ApplyRules(); err != nil {
		t.Errorf("ApplyRules() after rate limit reset failed: %v", err)
	}
}

func TestManager_TriggerEvent(t *testing.T) {
	manager := NewManager()

	handlerCalled := false
	handler := func(event Event) error {
		handlerCalled = true
		if event.Type != "test_event" {
			t.Errorf("Event type = %v, want test_event", event.Type)
		}
		return nil
	}

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeEvent,
		EventHandlers: map[string]EventHandler{
			"test_event": handler,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Trigger event
	event := Event{
		Type:      "test_event",
		Data:      map[string]interface{}{"key": "value"},
		Timestamp: time.Now(),
	}

	if err := manager.TriggerEvent(event); err != nil {
		t.Fatalf("TriggerEvent() error = %v", err)
	}

	if !handlerCalled {
		t.Error("Event handler was not called")
	}
}

func TestManager_GetStats(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"test": 1},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	stats := manager.GetStats()

	if totalMods, ok := stats["total_mods"].(int); !ok || totalMods != 1 {
		t.Errorf("total_mods = %v, want 1", stats["total_mods"])
	}

	if enabledMods, ok := stats["enabled_mods"].(int); !ok || enabledMods != 1 {
		t.Errorf("enabled_mods = %v, want 1", stats["enabled_mods"])
	}

	if activeRules, ok := stats["active_rules"].(int); !ok || activeRules != 1 {
		t.Errorf("active_rules = %v, want 1", stats["active_rules"])
	}
}

// Benchmarks

func BenchmarkManager_ApplyRules(b *testing.B) {
	// Use higher rate limit for benchmarking
	config := DefaultConfig()
	config.RuleChangeRateLimit = 1000.0 // High enough for benchmarking
	manager := NewManagerWithConfig(config)

	for i := 0; i < 10; i++ {
		mod := &Mod{
			ID:      fmt.Sprintf("mod%d", i),
			Name:    fmt.Sprintf("Mod %d", i),
			Version: "1.0.0",
			Type:    ModTypeRule,
			Rules: map[string]interface{}{
				"rule1": 1.0,
				"rule2": 2.0,
				"rule3": true,
			},
			Enabled: true,
		}

		if err := manager.AddMod(mod); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Sleep briefly to reset rate limit counter
		if i > 0 && i%50 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
		if err := manager.ApplyRules(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManager_GetRuleFloat64(b *testing.B) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		b.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = manager.GetRuleFloat64("difficulty", 1.0)
	}
}

func BenchmarkLoader_LoadFromFile(b *testing.B) {
	tmpDir := b.TempDir()

	mod := Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules:   map[string]interface{}{"test": 1},
		Enabled: true,
	}

	modPath := filepath.Join(tmpDir, "test.json")
	data, err := json.Marshal(mod)
	if err != nil {
		b.Fatal(err)
	}

	if err := os.WriteFile(modPath, data, 0o644); err != nil {
		b.Fatal(err)
	}

	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(modPath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestLoader_LoadFromFile_SandboxErrorFormatting(t *testing.T) {
	// Create temporary directory for test mod
	tmpDir := t.TempDir()

	// Create a mod with multiple entries that should trigger sandbox violations
	mod := Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"system.execute": "rm -rf /",                 // Rule name that violates allowed patterns
			"file.read":      "/etc/passwd",              // Another rule name outside allowed patterns
			"difficulty":     "<script>alert()</script>", // String value with script tag
		},
	}

	// Write mod to file
	modPath := filepath.Join(tmpDir, "bad-mod.json")
	data, err := json.Marshal(mod)
	if err != nil {
		t.Fatalf("Failed to marshal mod: %v", err)
	}
	if err := os.WriteFile(modPath, data, 0o644); err != nil {
		t.Fatalf("Failed to write mod file: %v", err)
	}

	// Create loader with sandbox enabled
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = true
	loader := NewLoaderWithConfig(config)

	// Load the mod - should fail with formatted error message
	_, err = loader.LoadFromFile(modPath)
	if err == nil {
		t.Fatal("Expected error due to sandbox violations")
	}

	// Verify error message contains multiple violations separated by semicolons
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}

	// Verify the error message contains semicolons (multiple errors are joined)
	if !strings.Contains(errMsg, "; ") {
		t.Error("Error message should contain '; ' separator for multiple violations")
	}

	// Count semicolons to verify multiple errors are present
	semicolonCount := strings.Count(errMsg, "; ")
	if semicolonCount < 2 {
		t.Errorf("Expected at least 2 semicolons in error message (for 3+ violations), got %d", semicolonCount)
	}

	// Verify specific violation types are mentioned
	if !strings.Contains(errMsg, "APIRestrictions") && !strings.Contains(errMsg, "rule name") {
		t.Error("Error message should mention rule name violations")
	}

	t.Logf("Error message (validated): %s", errMsg)
}

// TestLoader_LoadAll_ErrorWrapping tests that errors are properly wrapped
// when all mods fail to load, allowing errors.Is() and errors.As() to work.
func TestLoader_LoadAll_ErrorWrapping(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create invalid JSON files that will fail to load
	invalidFiles := []struct {
		name    string
		content string
	}{
		{"invalid1.json", `{invalid json`},
		{"invalid2.json", `{"ID": "test", "Name": "", "Version": "1.0.0"}`}, // Empty name
		{"invalid3.json", `not json at all`},
	}

	for _, file := range invalidFiles {
		path := filepath.Join(tmpDir, file.name)
		if err := os.WriteFile(path, []byte(file.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create loader
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	// Test loading all - should fail with wrapped errors
	_, err := loader.LoadAll()
	if err == nil {
		t.Fatal("Expected error when all mods fail to load")
	}

	// Verify the error message contains our context
	if !strings.Contains(err.Error(), "failed to load any mods") {
		t.Errorf("Error should contain context message, got: %v", err)
	}

	// Verify error wrapping works - the joined errors should be accessible
	// This verifies that we're using %w with errors.Join() instead of %v
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		t.Error("Error should be wrappable with errors.Unwrap()")
	}

	t.Logf("Error properly wrapped: %v", err)
}

// TestNewLoader tests the default loader constructor
func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
	if loader.config.ModsDirectory != "mods" {
		t.Errorf("Default ModsDirectory = %v, want 'mods'", loader.config.ModsDirectory)
	}
	if !loader.config.EnableSandbox {
		t.Error("Default EnableSandbox should be true")
	}
}

// TestLoader_GetModPath tests the mod path generation
func TestLoader_GetModPath(t *testing.T) {
	tests := []struct {
		name      string
		modsDir   string
		modID     string
		wantPath  string
	}{
		{
			name:     "default directory",
			modsDir:  "mods",
			modID:    "test-mod",
			wantPath: "mods/test-mod.json",
		},
		{
			name:     "custom directory",
			modsDir:  "/tmp/custom",
			modID:    "my-mod",
			wantPath: "/tmp/custom/my-mod.json",
		},
		{
			name:     "mod with dashes",
			modsDir:  "mods",
			modID:    "cool-mod-v2",
			wantPath: "mods/cool-mod-v2.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.ModsDirectory = tt.modsDir
			loader := NewLoaderWithConfig(config)
			
			gotPath := loader.GetModPath(tt.modID)
			if gotPath != tt.wantPath {
				t.Errorf("GetModPath(%s) = %v, want %v", tt.modID, gotPath, tt.wantPath)
			}
		})
	}
}

// TestManager_ListMods tests listing all loaded mods
func TestManager_ListMods(t *testing.T) {
	manager := NewManager()

	// Initially empty
	mods := manager.ListMods()
	if len(mods) != 0 {
		t.Errorf("Initial ListMods() returned %d mods, want 0", len(mods))
	}

	// Add some mods
	testMods := []*Mod{
		{
			ID:      "mod1",
			Name:    "Mod 1",
			Version: "1.0.0",
			Type:    ModTypeRule,
			Enabled: true,
		},
		{
			ID:      "mod2",
			Name:    "Mod 2",
			Version: "1.0.0",
			Type:    ModTypeGenerator,
			Enabled: false,
		},
		{
			ID:      "mod3",
			Name:    "Mod 3",
			Version: "1.0.0",
			Type:    ModTypeEvent,
			Enabled: true,
		},
	}

	for _, mod := range testMods {
		if err := manager.AddMod(mod); err != nil {
			t.Fatal(err)
		}
	}

	// List all mods
	mods = manager.ListMods()
	if len(mods) != len(testMods) {
		t.Errorf("ListMods() returned %d mods, want %d", len(mods), len(testMods))
	}

	// Verify all mods are present
	modIDs := make(map[string]bool)
	for _, mod := range mods {
		modIDs[mod.ID] = true
	}

	for _, expected := range testMods {
		if !modIDs[expected.ID] {
			t.Errorf("Mod %s not found in ListMods() result", expected.ID)
		}
	}
}

// TestManager_EnableMod tests enabling a mod
func TestManager_EnableMod(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: false, // Start disabled
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Enable the mod
	if err := manager.EnableMod("test-mod"); err != nil {
		t.Fatalf("EnableMod() error = %v", err)
	}

	// Verify it's enabled
	loadedMod, exists := manager.GetMod("test-mod")
	if !exists {
		t.Fatal("Mod not found after enabling")
	}

	if !loadedMod.Enabled {
		t.Error("Mod should be enabled after EnableMod()")
	}
}

// TestManager_EnableMod_NotFound tests enabling a non-existent mod
func TestManager_EnableMod_NotFound(t *testing.T) {
	manager := NewManager()

	err := manager.EnableMod("nonexistent")
	if err == nil {
		t.Error("Expected error when enabling non-existent mod")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error message should mention 'not found', got: %v", err)
	}
}

// TestManager_DisableMod tests disabling a mod
func TestManager_DisableMod(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true, // Start enabled
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Disable the mod
	if err := manager.DisableMod("test-mod"); err != nil {
		t.Fatalf("DisableMod() error = %v", err)
	}

	// Verify it's disabled
	loadedMod, exists := manager.GetMod("test-mod")
	if !exists {
		t.Fatal("Mod not found after disabling")
	}

	if loadedMod.Enabled {
		t.Error("Mod should be disabled after DisableMod()")
	}
}

// TestManager_DisableMod_NotFound tests disabling a non-existent mod
func TestManager_DisableMod_NotFound(t *testing.T) {
	manager := NewManager()

	err := manager.DisableMod("nonexistent")
	if err == nil {
		t.Error("Expected error when disabling non-existent mod")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error message should mention 'not found', got: %v", err)
	}
}

// TestManager_GetRuleChangeLog tests retrieving rule change history
func TestManager_GetRuleChangeLog(t *testing.T) {
	manager := NewManager()

	// Initially empty
	log := manager.GetRuleChangeLog()
	if len(log) != 0 {
		t.Errorf("Initial GetRuleChangeLog() returned %d entries, want 0", len(log))
	}

	// Add a mod with rules
	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"difficulty": 2.0,
			"permadeath": true,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Apply rules to generate log entries
	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	// Check log
	log = manager.GetRuleChangeLog()
	if len(log) == 0 {
		t.Error("GetRuleChangeLog() should contain entries after ApplyRules()")
	}

	// Verify log entries contain expected data
	foundDifficulty := false
	foundPermadeath := false
	for _, entry := range log {
		if entry.ModID != "test-mod" {
			t.Errorf("Log entry ModID = %v, want 'test-mod'", entry.ModID)
		}
		if entry.RuleName == "difficulty" {
			foundDifficulty = true
			if entry.NewValue != 2.0 {
				t.Errorf("Difficulty NewValue = %v, want 2.0", entry.NewValue)
			}
		}
		if entry.RuleName == "permadeath" {
			foundPermadeath = true
			if entry.NewValue != true {
				t.Errorf("Permadeath NewValue = %v, want true", entry.NewValue)
			}
		}
		if entry.AppliedAt.IsZero() {
			t.Error("Log entry AppliedAt should be set")
		}
	}

	if !foundDifficulty {
		t.Error("Log should contain difficulty rule")
	}
	if !foundPermadeath {
		t.Error("Log should contain permadeath rule")
	}
}

// TestLoadError_Error tests the LoadError error message
func TestLoadError_Error(t *testing.T) {
	tests := []struct {
		name     string
		modID    string
		err      error
		wantText string
	}{
		{
			name:     "basic error",
			modID:    "test-mod",
			err:      fmt.Errorf("file not found"),
			wantText: "failed to load mod test-mod: file not found",
		},
		{
			name:     "validation error",
			modID:    "bad-mod",
			err:      fmt.Errorf("invalid JSON"),
			wantText: "failed to load mod bad-mod: invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadErr := &LoadError{
				ModID: tt.modID,
				Err:   tt.err,
			}

			gotMsg := loadErr.Error()
			if gotMsg != tt.wantText {
				t.Errorf("Error() = %q, want %q", gotMsg, tt.wantText)
			}
		})
	}
}

// TestValidationError_Error tests the ValidationError error message
func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		modID    string
		field    string
		reason   string
		wantText string
	}{
		{
			name:     "empty name field",
			modID:    "test-mod",
			field:    "Name",
			reason:   "cannot be empty",
			wantText: "validation failed for mod test-mod (field Name): cannot be empty",
		},
		{
			name:     "invalid type",
			modID:    "bad-mod",
			field:    "Type",
			reason:   "unsupported type",
			wantText: "validation failed for mod bad-mod (field Type): unsupported type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valErr := &ValidationError{
				ModID:  tt.modID,
				Field:  tt.field,
				Reason: tt.reason,
			}

			gotMsg := valErr.Error()
			if gotMsg != tt.wantText {
				t.Errorf("Error() = %q, want %q", gotMsg, tt.wantText)
			}
		})
	}
}

// TestLoader_SaveToFile_InvalidMod tests saving an invalid mod
func TestLoader_SaveToFile_InvalidMod(t *testing.T) {
	tmpDir := t.TempDir()

	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	// Try to save a mod with empty ID (invalid)
	invalidMod := &Mod{
		ID:      "", // Invalid: empty ID
		Name:    "Test Mod",
		Version: "1.0.0",
	}

	modPath := filepath.Join(tmpDir, "invalid.json")
	err := loader.SaveToFile(invalidMod, modPath)
	if err == nil {
		t.Error("Expected error when saving invalid mod")
	}

	// Verify the file was not created
	if _, statErr := os.Stat(modPath); !os.IsNotExist(statErr) {
		t.Error("Invalid mod file should not be created")
	}
}

// TestLoader_SaveToFile_SandboxViolation tests saving with sandbox enabled
func TestLoader_SaveToFile_SandboxViolation(t *testing.T) {
	config := DefaultConfig()
	config.ModsDirectory = "/tmp/mods"
	config.EnableSandbox = true
	loader := NewLoaderWithConfig(config)

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Enabled: true,
	}

	// Try to save outside mods directory
	err := loader.SaveToFile(mod, "/etc/passwd")
	if err == nil {
		t.Error("Expected sandbox violation error")
	}
}

// TestManager_RemoveMod_NotFound tests removing a non-existent mod
func TestManager_RemoveMod_NotFound(t *testing.T) {
	manager := NewManager()

	err := manager.RemoveMod("nonexistent")
	if err == nil {
		t.Error("Expected error when removing non-existent mod")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error message should mention 'not found', got: %v", err)
	}
}

// TestManager_RemoveMod_EventHandlerCleanup tests that event handlers are cleaned up
func TestManager_RemoveMod_EventHandlerCleanup(t *testing.T) {
	manager := NewManager()

	handlerCalled := false
	handler := func(event Event) error {
		handlerCalled = true
		return nil
	}

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeEvent,
		EventHandlers: map[string]EventHandler{
			"test_event": handler,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	// Trigger event to verify handler works
	event := Event{Type: "test_event", Timestamp: time.Now()}
	if err := manager.TriggerEvent(event); err != nil {
		t.Fatal(err)
	}

	if !handlerCalled {
		t.Error("Handler should be called before removal")
	}

	// Remove mod
	if err := manager.RemoveMod("test-mod"); err != nil {
		t.Fatal(err)
	}

	// Reset flag and try triggering again
	handlerCalled = false
	if err := manager.TriggerEvent(event); err != nil {
		t.Fatal(err)
	}

	if handlerCalled {
		t.Error("Handler should not be called after mod removal")
	}
}

// TestManager_GetRuleFloat64_TypeConversions tests type conversion scenarios
func TestManager_GetRuleFloat64_TypeConversions(t *testing.T) {
	tests := []struct {
		name         string
		ruleValue    interface{}
		defaultValue float64
		want         float64
	}{
		{
			name:         "float64 value",
			ruleValue:    2.5,
			defaultValue: 1.0,
			want:         2.5,
		},
		{
			name:         "float32 value",
			ruleValue:    float32(3.5),
			defaultValue: 1.0,
			want:         3.5,
		},
		{
			name:         "int value",
			ruleValue:    42,
			defaultValue: 1.0,
			want:         42.0,
		},
		{
			name:         "int64 value",
			ruleValue:    int64(100),
			defaultValue: 1.0,
			want:         100.0,
		},
		{
			name:         "string value (unsupported)",
			ruleValue:    "not a number",
			defaultValue: 1.0,
			want:         1.0, // Should return default
		},
		{
			name:         "bool value (unsupported)",
			ruleValue:    true,
			defaultValue: 5.0,
			want:         5.0, // Should return default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()

			mod := &Mod{
				ID:      "test-mod",
				Name:    "Test Mod",
				Version: "1.0.0",
				Type:    ModTypeRule,
				Rules: map[string]interface{}{
					"test_rule": tt.ruleValue,
				},
				Enabled: true,
			}

			if err := manager.AddMod(mod); err != nil {
				t.Fatal(err)
			}

			if err := manager.ApplyRules(); err != nil {
				t.Fatal(err)
			}

			got := manager.GetRuleFloat64("test_rule", tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetRuleFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestManager_GetRuleBool_InvalidType tests bool conversion with invalid types
func TestManager_GetRuleBool_InvalidType(t *testing.T) {
	manager := NewManager()

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeRule,
		Rules: map[string]interface{}{
			"string_rule": "not a bool",
			"number_rule": 42,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	if err := manager.ApplyRules(); err != nil {
		t.Fatal(err)
	}

	// String value should return default
	if val := manager.GetRuleBool("string_rule", false); val != false {
		t.Errorf("GetRuleBool with string should return default, got %v", val)
	}

	// Number value should return default
	if val := manager.GetRuleBool("number_rule", true); val != true {
		t.Errorf("GetRuleBool with number should return default, got %v", val)
	}
}

// TestManager_TriggerEvent_NoHandlers tests triggering an event with no handlers
func TestManager_TriggerEvent_NoHandlers(t *testing.T) {
	manager := NewManager()

	event := Event{
		Type:      "unhandled_event",
		Data:      map[string]interface{}{},
		Timestamp: time.Now(),
	}

	// Should not error when no handlers exist
	if err := manager.TriggerEvent(event); err != nil {
		t.Errorf("TriggerEvent() with no handlers should not error, got: %v", err)
	}
}

// TestManager_TriggerEvent_HandlerError tests triggering an event where handler returns error
func TestManager_TriggerEvent_HandlerError(t *testing.T) {
	manager := NewManager()

	expectedErr := fmt.Errorf("handler failed")
	handler := func(event Event) error {
		return expectedErr
	}

	mod := &Mod{
		ID:      "test-mod",
		Name:    "Test Mod",
		Version: "1.0.0",
		Type:    ModTypeEvent,
		EventHandlers: map[string]EventHandler{
			"error_event": handler,
		},
		Enabled: true,
	}

	if err := manager.AddMod(mod); err != nil {
		t.Fatal(err)
	}

	event := Event{
		Type:      "error_event",
		Timestamp: time.Now(),
	}

	err := manager.TriggerEvent(event)
	if err == nil {
		t.Error("Expected error from handler")
	}

	if !strings.Contains(err.Error(), "event handler failed") {
		t.Errorf("Error should mention handler failure, got: %v", err)
	}
}

// TestLoader_LoadAll_MaxModsLimit tests that LoadAll respects the MaxMods limit
func TestLoader_LoadAll_MaxModsLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 5 mod files
	for i := 1; i <= 5; i++ {
		mod := Mod{
			ID:      fmt.Sprintf("mod%d", i),
			Name:    fmt.Sprintf("Mod %d", i),
			Version: "1.0.0",
			Type:    ModTypeRule,
			Enabled: true,
		}

		data, err := json.Marshal(mod)
		if err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(tmpDir, fmt.Sprintf("mod%d.json", i))
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Set max mods to 3
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	config.MaxMods = 3
	loader := NewLoaderWithConfig(config)

	// Load all - should only load 3
	mods, err := loader.LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(mods) != 3 {
		t.Errorf("LoadAll() with MaxMods=3 loaded %d mods, want 3", len(mods))
	}
}

// TestLoader_LoadAll_NonExistentDirectory tests LoadAll with non-existent directory
func TestLoader_LoadAll_NonExistentDirectory(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "nonexistent")

	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	// Should create directory and return empty list
	mods, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() should create directory, got error: %v", err)
	}

	if len(mods) != 0 {
		t.Errorf("LoadAll() with new directory should return 0 mods, got %d", len(mods))
	}

	// Verify directory was created
	if _, statErr := os.Stat(tmpDir); os.IsNotExist(statErr) {
		t.Error("LoadAll() should create mods directory if it doesn't exist")
	}
}

// TestLoader_LoadFromFile_LargeFile tests loading a file that exceeds size limit
func TestLoader_LoadFromFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a large description string to make file > 1MB
	// Avoid using many rules as they'll fail sandbox validation
	largeDescription := strings.Repeat("This is a very long description. ", 50000)

	mod := Mod{
		ID:          "large-mod",
		Name:        "Large Mod",
		Version:     "1.0.0",
		Description: largeDescription,
		Type:        ModTypeRule,
		Rules:       map[string]interface{}{"difficulty": 1.0},
		Enabled:     true,
	}

	modPath := filepath.Join(tmpDir, "large.json")
	data, err := json.Marshal(mod)
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's actually > 1MB
	if len(data) <= 1024*1024 {
		t.Fatalf("Test data is not large enough: %d bytes", len(data))
	}

	if err := os.WriteFile(modPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	// Try to load with sandbox enabled (1MB limit)
	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = true
	loader := NewLoaderWithConfig(config)

	_, err = loader.LoadFromFile(modPath)
	if err == nil {
		t.Error("Expected error for large file with sandbox enabled")
	}

	if !strings.Contains(err.Error(), "1MB") {
		t.Errorf("Error should mention size limit, got: %v", err)
	}
}

// TestLoader_LoadFromFile_InvalidJSON tests loading a file with invalid JSON
func TestLoader_LoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	modPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(modPath, []byte("{invalid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	config := DefaultConfig()
	config.ModsDirectory = tmpDir
	config.EnableSandbox = false
	loader := NewLoaderWithConfig(config)

	_, err := loader.LoadFromFile(modPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("Error should mention invalid JSON, got: %v", err)
	}
}
