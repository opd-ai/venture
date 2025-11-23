package migration

import (
	"testing"
)

func TestValidator_NewValidator(t *testing.T) {
	config := Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	}

	validator := NewValidator(config)
	if validator == nil {
		t.Fatal("expected validator instance, got nil")
	}
	if validator.config.TargetVersion != "v10.0" {
		t.Errorf("expected target version v10.0, got %s", validator.config.TargetVersion)
	}
}

func TestValidator_Defaults(t *testing.T) {
	validator := NewValidator(Config{})

	if validator.config.TargetVersion != "v10.0" {
		t.Errorf("expected default target version v10.0, got %s", validator.config.TargetVersion)
	}
	if validator.config.TestDataPath != "testdata/saves/" {
		t.Errorf("expected default test data path, got %s", validator.config.TestDataPath)
	}
}

func TestValidator_ValidateMigration_V1ToV10(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("v1.0", "v10.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	if result.SourceVersion != "v1.0" {
		t.Errorf("expected source version v1.0, got %s", result.SourceVersion)
	}
	if result.TargetVersion != "v10.0" {
		t.Errorf("expected target version v10.0, got %s", result.TargetVersion)
	}
	if result.MigrationTime == 0 {
		t.Error("expected non-zero migration time")
	}
	if len(result.ComponentsPreserved) == 0 {
		t.Error("expected components to be preserved")
	}
}

func TestValidator_ValidateMigration_V5ToV10(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("v5.0", "v10.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	if len(result.ComponentsPreserved) < 3 {
		t.Errorf("expected at least 3 preserved components, got %d", len(result.ComponentsPreserved))
	}
}

func TestValidator_ValidateMigration_V9ToV10(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("v9.0", "v10.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	// v9.0 to v10.0 should preserve all components
	if len(result.ComponentsPreserved) < 3 {
		t.Errorf("expected at least 3 components, got %d", len(result.ComponentsPreserved))
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	results, err := validator.ValidateAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.TotalCount != 9 { // v1.0 through v9.0
		t.Errorf("expected 9 migrations, got %d", results.TotalCount)
	}

	// All migrations should pass with synthetic data
	if results.FailedCount > 0 {
		t.Errorf("expected 0 failures, got %d", results.FailedCount)
		for _, migration := range results.Migrations {
			if !migration.Passed {
				t.Logf("Failed migration %s→%s: %s", migration.SourceVersion, migration.TargetVersion, migration.Error)
			}
		}
	}

	if results.PassedCount != results.TotalCount {
		t.Errorf("expected all %d migrations to pass, got %d passed", results.TotalCount, results.PassedCount)
	}
}

func TestValidator_GenerateSyntheticSave(t *testing.T) {
	validator := NewValidator(Config{})

	save := validator.generateSyntheticSave("v1.0")

	// Verify required fields
	if version, ok := save["version"].(string); !ok || version != "v1.0" {
		t.Error("expected version field v1.0")
	}
	if _, ok := save["player"]; !ok {
		t.Error("expected player field")
	}
	if _, ok := save["world"]; !ok {
		t.Error("expected world field")
	}
	if _, ok := save["inventory"]; !ok {
		t.Error("expected inventory field")
	}
}

func TestValidator_ApplyMigrationRules_V10(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "v1.0",
		"player":  map[string]interface{}{"level": 10},
	}

	err := validator.applyMigrationRules(data, "v1.0", "v10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify v10.0 fields were added
	if _, exists := data["audit_flags"]; !exists {
		t.Error("expected audit_flags to be added for v10.0")
	}
}

func TestValidator_ApplyMigrationRules_V9(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "v1.0",
		"player":  map[string]interface{}{"level": 10},
	}

	err := validator.applyMigrationRules(data, "v1.0", "v9.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify v9.0 fields were added
	if _, exists := data["performance"]; !exists {
		t.Error("expected performance to be added for v9.0")
	}
}

func TestValidator_ValidateData(t *testing.T) {
	validator := NewValidator(Config{})

	// Valid data
	validData := map[string]interface{}{
		"version": "v10.0",
		"player":  map[string]interface{}{"level": 10},
		"world":   map[string]interface{}{"seed": 12345},
	}

	components, err := validator.validateData(validData, "v10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(components) < 3 {
		t.Errorf("expected at least 3 components, got %d", len(components))
	}

	// Missing required field
	invalidData := map[string]interface{}{
		"version": "v10.0",
		"player":  map[string]interface{}{"level": 10},
		// Missing "world"
	}

	_, err = validator.validateData(invalidData, "v10.0")
	if err == nil {
		t.Error("expected error for missing required field")
	}

	// Version mismatch
	mismatchData := map[string]interface{}{
		"version": "v9.0",
		"player":  map[string]interface{}{"level": 10},
		"world":   map[string]interface{}{"seed": 12345},
	}

	_, err = validator.validateData(mismatchData, "v10.0")
	if err == nil {
		t.Error("expected error for version mismatch")
	}
}

func TestValidator_ComponentPreservation(t *testing.T) {
	validator := NewValidator(Config{
		ValidateData: true,
	})

	// Create save with optional components
	saveData := map[string]interface{}{
		"version":    "v5.0",
		"player":     map[string]interface{}{"level": 20},
		"world":      map[string]interface{}{"seed": 12345},
		"inventory":  []interface{}{},
		"quests":     []interface{}{},
		"companions": []interface{}{},
		"vehicles":   []interface{}{},
	}

	migratedData, _, err := validator.performMigration(saveData, "v5.0", "v10.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	components, err := validator.validateData(migratedData, "v10.0")
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}

	// Verify all components preserved
	expectedComponents := []string{"version", "player", "world", "inventory", "quests", "companions", "vehicles"}
	for _, expected := range expectedComponents {
		found := false
		for _, component := range components {
			if component == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected component %s to be preserved", expected)
		}
	}
}

func BenchmarkValidator_ValidateMigration(b *testing.B) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  false, // Skip validation for speed
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateMigration("v1.0", "v10.0")
	}
}

func BenchmarkValidator_ValidateAll(b *testing.B) {
	validator := NewValidator(Config{
		TargetVersion: "v10.0",
		TestDataPath:  "testdata/",
		ValidateData:  false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateAll()
	}
}
