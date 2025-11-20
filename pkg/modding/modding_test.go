package modding

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
