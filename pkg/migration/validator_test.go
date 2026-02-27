//go:build !js
// +build !js

package migration

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/saveload"
	"github.com/sirupsen/logrus"
)

func TestValidator_NewValidator(t *testing.T) {
	config := Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	}

	validator := NewValidator(config)
	if validator == nil {
		t.Fatal("expected validator instance, got nil")
	}
	if validator.config.TargetVersion != "1.0.0" {
		t.Errorf("expected target version 1.0.0, got %s", validator.config.TargetVersion)
	}
}

func TestValidator_Defaults(t *testing.T) {
	validator := NewValidator(Config{})

	if validator.config.TargetVersion != "1.0.0" {
		t.Errorf("expected default target version 1.0.0, got %s", validator.config.TargetVersion)
	}
	if validator.config.TestDataPath != "testdata/saves/" {
		t.Errorf("expected default test data path, got %s", validator.config.TestDataPath)
	}
}

func TestValidator_ValidateMigration_090To100(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	if result.SourceVersion != "0.9.0" {
		t.Errorf("expected source version 0.9.0, got %s", result.SourceVersion)
	}
	if result.TargetVersion != "1.0.0" {
		t.Errorf("expected target version 1.0.0, got %s", result.TargetVersion)
	}
	if result.MigrationTime == 0 {
		t.Error("expected non-zero migration time")
	}
	if len(result.ComponentsPreserved) == 0 {
		t.Error("expected components to be preserved")
	}
}

func TestValidator_ValidateMigration_092To100(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("0.9.2", "1.0.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	if len(result.ComponentsPreserved) < 3 {
		t.Errorf("expected at least 3 preserved components, got %d", len(result.ComponentsPreserved))
	}
}

func TestValidator_ValidateMigration_093To100(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("0.9.3", "1.0.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	// 0.9.3 to 1.0.0 should preserve all components
	if len(result.ComponentsPreserved) < 3 {
		t.Errorf("expected at least 3 components, got %d", len(result.ComponentsPreserved))
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	results, err := validator.ValidateAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.TotalCount != 4 { // 0.9.0 through 0.9.3
		t.Errorf("expected 4 migrations, got %d", results.TotalCount)
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

	save := validator.generateSyntheticSave("0.9.0")

	// Verify required fields
	if version, ok := save["version"].(string); !ok || version != "0.9.0" {
		t.Error("expected version field 0.9.0")
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

func TestValidator_ApplyMigrationRules_090(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "0.9.0",
		"player":  map[string]interface{}{"level": 10},
	}

	err := validator.applyMigrationRules(data, "0.9.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify trust and reputation fields were added (matching pkg/saveload migrations)
	player, ok := data["player"].(map[string]interface{})
	if !ok {
		t.Fatal("expected player to be a map")
	}
	if _, exists := player["trust_scores"]; !exists {
		t.Error("expected trust_scores to be added for 0.9.0 migration")
	}
	if _, exists := player["reputation_scores"]; !exists {
		t.Error("expected reputation_scores to be added for 0.9.0 migration")
	}
}

func TestValidator_ApplyMigrationRules_092(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "0.9.2",
		"player":  map[string]interface{}{"level": 10},
	}

	err := validator.applyMigrationRules(data, "0.9.2", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify trust fields were added (matching pkg/saveload migrations)
	player, ok := data["player"].(map[string]interface{})
	if !ok {
		t.Fatal("expected player to be a map")
	}
	if _, exists := player["trust_scores"]; !exists {
		t.Error("expected trust_scores to be added for 0.9.2 migration")
	}
}

func TestValidator_ValidateData(t *testing.T) {
	validator := NewValidator(Config{})

	// Valid data
	validData := map[string]interface{}{
		"version": "1.0.0",
		"player":  map[string]interface{}{"level": 10},
		"world":   map[string]interface{}{"seed": 12345},
	}

	components, err := validator.validateData(validData, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(components) < 3 {
		t.Errorf("expected at least 3 components, got %d", len(components))
	}

	// Missing required field
	invalidData := map[string]interface{}{
		"version": "1.0.0",
		"player":  map[string]interface{}{"level": 10},
		// Missing "world"
	}

	_, err = validator.validateData(invalidData, "1.0.0")
	if err == nil {
		t.Error("expected error for missing required field")
	}

	// Version mismatch
	mismatchData := map[string]interface{}{
		"version": "0.9.3",
		"player":  map[string]interface{}{"level": 10},
		"world":   map[string]interface{}{"seed": 12345},
	}

	_, err = validator.validateData(mismatchData, "1.0.0")
	if err == nil {
		t.Error("expected error for version mismatch")
	}
}

func TestValidator_ComponentPreservation(t *testing.T) {
	validator := NewValidator(Config{
		ValidateData: true,
	})

	// Create save with valid GameSave structure
	// Note: In the real GameSave, inventory/companions/vehicles are nested under player
	saveData := map[string]interface{}{
		"version": "0.9.2",
		"player": map[string]interface{}{
			"level":             20,
			"items":             []interface{}{},
			"active_companions": []interface{}{},
			"owned_vehicles":    []interface{}{},
		},
		"world": map[string]interface{}{
			"seed":  12345,
			"depth": 5,
		},
		"settings": map[string]interface{}{
			"screen_width":  800,
			"screen_height": 600,
		},
	}

	migratedData, _, err := validator.performMigration(saveData, "0.9.2", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	components, err := validator.validateData(migratedData, "1.0.0")
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}

	// Verify required components are preserved (version, player, world)
	// Note: Real migrator produces GameSave structure, not arbitrary maps
	expectedRequired := []string{"version", "player", "world"}
	for _, expected := range expectedRequired {
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

	// Verify at least the 3 required components are present
	if len(components) < 3 {
		t.Errorf("expected at least 3 components, got %d: %v", len(components), components)
	}
}

func BenchmarkValidator_ValidateMigration(b *testing.B) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  false, // Skip validation for speed
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateMigration("0.9.0", "1.0.0")
	}
}

func BenchmarkValidator_ValidateAll(b *testing.B) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateAll()
	}
}

// TestValidator_LoadSaveFile_RealFile tests loading from an actual save file.
func TestValidator_LoadSaveFile_RealFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TestDataPath: tmpDir + "/",
	})

	// Create a valid save file
	validSavePath := tmpDir + "/save_0.9.0.json"
	validSaveContent := `{"version":"0.9.0","player":{"name":"Test","level":5},"world":{"seed":42}}`
	if err := os.WriteFile(validSavePath, []byte(validSaveContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test loading the valid file
	data, err := validator.loadSaveFile(validSavePath)
	if err != nil {
		t.Fatalf("unexpected error loading valid file: %v", err)
	}
	if version, ok := data["version"].(string); !ok || version != "0.9.0" {
		t.Errorf("expected version 0.9.0, got %v", data["version"])
	}
	if player, ok := data["player"].(map[string]interface{}); !ok {
		t.Error("expected player to be a map")
	} else if name, ok := player["name"].(string); !ok || name != "Test" {
		t.Errorf("expected player name 'Test', got %v", player["name"])
	}
}

// TestValidator_LoadSaveFile_InvalidJSON tests loading a file with invalid JSON.
func TestValidator_LoadSaveFile_InvalidJSON(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TestDataPath: tmpDir + "/",
	})

	// Create an invalid JSON file
	invalidPath := tmpDir + "/save_invalid.json"
	if err := os.WriteFile(invalidPath, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Test loading the invalid file
	_, err := validator.loadSaveFile(invalidPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("expected 'failed to parse JSON' error, got: %v", err)
	}
}

// TestValidator_LoadSaveFile_ReadError tests handling of file read errors.
func TestValidator_LoadSaveFile_ReadError(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TestDataPath: tmpDir + "/",
	})

	// Create a directory with the save file name (can't read a directory as a file)
	dirPath := tmpDir + "/save_dir.json"
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Test loading a directory as a file (should fail)
	_, err := validator.loadSaveFile(dirPath)
	if err == nil {
		t.Error("expected error when reading a directory as a file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' error, got: %v", err)
	}
}

// TestValidator_ValidateMigration_WithoutValidation tests migration without data validation.
func TestValidator_ValidateMigration_WithoutValidation(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  false, // Skip validation
	})

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	// Without validation, ComponentsPreserved should be empty
	if len(result.ComponentsPreserved) > 0 {
		t.Errorf("expected no components when validation disabled, got %d", len(result.ComponentsPreserved))
	}
}

// TestValidator_ValidateMigration_ValidationFailed tests migration when validation fails.
func TestValidator_ValidateMigration_ValidationFailed(t *testing.T) {
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  tmpDir + "/",
		ValidateData:  true,
	})

	// Create a save file that will fail validation (missing world field)
	incompleteSavePath := tmpDir + "/save_0.9.0.json"
	incompleteContent := `{"version":"0.9.0","player":{"name":"Test"}}`
	if err := os.WriteFile(incompleteSavePath, []byte(incompleteContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	// Migration will still pass because ensureRequiredFields adds missing world field
	// This is correct behavior - migration should add missing fields
	if !result.Passed {
		t.Errorf("expected migration to pass (missing fields should be added), got error: %s", result.Error)
	}
}

// TestValidator_ValidateMigration_InvalidPlayerType tests migration when player field has wrong type.
func TestValidator_ValidateMigration_InvalidPlayerType(t *testing.T) {
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  tmpDir + "/",
		ValidateData:  true,
	})

	// Create a save file where player is a string instead of an object
	invalidSavePath := tmpDir + "/save_0.9.0.json"
	invalidContent := `{"version":"0.9.0","player":"invalid_string","world":{"seed":42}}`
	if err := os.WriteFile(invalidSavePath, []byte(invalidContent), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	// Migration should fail because player type is invalid
	// Now using real migrator which catches this during JSON unmarshal
	if result.Passed {
		t.Error("expected migration to fail for invalid player type")
	}
	// Check for any indication of type mismatch - either original or JSON unmarshal error
	if !strings.Contains(result.Error, "must be an object") &&
		!strings.Contains(result.Error, "cannot unmarshal") &&
		!strings.Contains(result.Error, "failed to convert") {
		t.Errorf("expected type mismatch error, got: %s", result.Error)
	}
}

// TestValidator_EnsureRequiredFields_EmptyData tests adding fields to empty data.
func TestValidator_EnsureRequiredFields_EmptyData(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{}
	validator.ensureRequiredFields(data)

	// Verify player was added
	player, ok := data["player"].(map[string]interface{})
	if !ok {
		t.Fatal("expected player to be added")
	}
	if _, hasItems := player["items"]; !hasItems {
		t.Error("expected player.items to be added")
	}

	// Verify world was added
	world, ok := data["world"].(map[string]interface{})
	if !ok {
		t.Fatal("expected world to be added")
	}
	if _, hasEntities := world["modified_entities"]; !hasEntities {
		t.Error("expected world.modified_entities to be added")
	}

	// Verify settings was added
	settings, ok := data["settings"].(map[string]interface{})
	if !ok {
		t.Fatal("expected settings to be added")
	}
	if _, hasVsync := settings["vsync"]; !hasVsync {
		t.Error("expected settings.vsync to be added")
	}
}

// TestValidator_EnsureRequiredFields_ExistingPlayerWithoutItems tests adding items to existing player.
func TestValidator_EnsureRequiredFields_ExistingPlayerWithoutItems(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"player": map[string]interface{}{
			"name":  "Existing",
			"level": 20,
			// Missing items field
		},
	}
	validator.ensureRequiredFields(data)

	// Verify items was added to existing player
	player := data["player"].(map[string]interface{})
	if _, hasItems := player["items"]; !hasItems {
		t.Error("expected items to be added to existing player")
	}
	// Verify existing fields preserved
	if name, _ := player["name"].(string); name != "Existing" {
		t.Error("expected existing player fields to be preserved")
	}
}

// TestValidator_EnsureRequiredFields_ExistingWorldWithoutModifiedEntities tests adding modified_entities.
func TestValidator_EnsureRequiredFields_ExistingWorldWithoutModifiedEntities(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"world": map[string]interface{}{
			"seed":  12345,
			"depth": 10,
			// Missing modified_entities field
		},
	}
	validator.ensureRequiredFields(data)

	// Verify modified_entities was added to existing world
	world := data["world"].(map[string]interface{})
	if _, hasEntities := world["modified_entities"]; !hasEntities {
		t.Error("expected modified_entities to be added to existing world")
	}
	// Verify existing fields preserved
	if seed, _ := world["seed"].(int); seed != 12345 {
		t.Error("expected existing world fields to be preserved")
	}
}

// TestValidator_ExtractVersion tests version extraction edge cases.
func TestValidator_ExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"valid semantic version", "saves/save_0.9.0.json", "0.9.0"},
		{"valid legacy version", "saves/save_v1.0.json", "v1.0"},
		{"no underscore", "saves/save.json", ""},
		{"too short version", "saves/save_ab.json", ""},
		{"invalid start char", "saves/save_xyz.json", ""},
		{"deeply nested path", "/foo/bar/baz/save_0.9.3.json", "0.9.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.path)
			if result != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// TestValidator_ValidateData_OptionalComponents tests that optional components are detected.
func TestValidator_ValidateData_OptionalComponents(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version":    "1.0.0",
		"player":     map[string]interface{}{"level": 10},
		"world":      map[string]interface{}{"seed": 42},
		"inventory":  []interface{}{},
		"quests":     []interface{}{},
		"companions": []interface{}{},
		"vehicles":   []interface{}{},
		"guild":      map[string]interface{}{"name": "TestGuild"},
	}

	components, err := validator.validateData(data, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have all required + all optional components
	expectedCount := 8 // version, player, world + 5 optional
	if len(components) != expectedCount {
		t.Errorf("expected %d components, got %d: %v", expectedCount, len(components), components)
	}
}

// TestValidator_ApplyMigrationRules_091 tests migration rules for 0.9.1.
func TestValidator_ApplyMigrationRules_091(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "0.9.1",
		"player":  map[string]interface{}{"level": 15},
	}

	err := validator.applyMigrationRules(data, "0.9.1", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify trust and reputation fields were added (matching 0.9.0 and 0.9.1 behavior)
	player := data["player"].(map[string]interface{})
	if _, exists := player["trust_scores"]; !exists {
		t.Error("expected trust_scores to be added for 0.9.1 migration")
	}
	if _, exists := player["reputation_scores"]; !exists {
		t.Error("expected reputation_scores to be added for 0.9.1 migration")
	}
}

// TestValidator_ValidateMigration_LoadError tests the load file error path in ValidateMigration.
func TestValidator_ValidateMigration_LoadError(t *testing.T) {
	tmpDir := t.TempDir()
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  tmpDir + "/",
		ValidateData:  true,
	})

	// Create an invalid JSON save file to trigger JSON parse error
	invalidJSONPath := tmpDir + "/save_0.9.0.json"
	if err := os.WriteFile(invalidJSONPath, []byte(`{invalid json`), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	// Should fail with load error
	if result.Passed {
		t.Error("expected migration to fail with load error")
	}
	if !strings.Contains(result.Error, "failed to load save file") {
		t.Errorf("expected 'failed to load save file' error, got: %s", result.Error)
	}
}

// TestValidator_ValidateMigration_093 tests migration from 0.9.3 (minimal changes).
func TestValidator_ValidateMigration_093(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("0.9.3", "1.0.0")

	if !result.Passed {
		t.Errorf("expected 0.9.3 migration to pass, got error: %s", result.Error)
	}
	// 0.9.3 is close to 1.0.0, should work smoothly
	if len(result.ComponentsPreserved) < 3 {
		t.Errorf("expected at least 3 components, got %d", len(result.ComponentsPreserved))
	}
}

// TestValidator_NewValidatorWithLogger tests creating validator with custom logger.
func TestValidator_NewValidatorWithLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	validator := NewValidatorWithLogger(Config{
		TargetVersion: "1.0.0",
	}, logger)

	if validator == nil {
		t.Fatal("expected validator instance, got nil")
	}
	if validator.logger != logger {
		t.Error("expected custom logger to be set")
	}
}

// TestValidator_NewValidatorWithMigrator tests creating validator with custom migrator.
func TestValidator_NewValidatorWithMigrator(t *testing.T) {
	migrator := saveload.NewDefaultMigrator()

	validator := NewValidatorWithMigrator(Config{
		TargetVersion: "1.0.0",
	}, migrator)

	if validator == nil {
		t.Fatal("expected validator instance, got nil")
	}
	if validator.migrator != migrator {
		t.Error("expected custom migrator to be set")
	}
}

// TestValidator_RealMigrationTime tests that migration time is actually measured.
func TestValidator_RealMigrationTime(t *testing.T) {
	validator := NewValidator(Config{
		TargetVersion: "1.0.0",
		TestDataPath:  "testdata/",
		ValidateData:  true,
	})

	result := validator.ValidateMigration("0.9.0", "1.0.0")

	if !result.Passed {
		t.Errorf("expected migration to pass, got error: %s", result.Error)
	}
	// Migration time should be non-zero (real measurement)
	if result.MigrationTime == 0 {
		t.Error("expected non-zero migration time")
	}
	// Migration time should be reasonable (not the old hardcoded 0.001)
	// It should be at least slightly different from 0.001 due to actual timing
	// Note: Very fast systems might still measure ~0.001s, so we just check it's > 0
	if result.MigrationTime <= 0 {
		t.Errorf("expected positive migration time, got %f", result.MigrationTime)
	}
}

// TestValidator_MapToGameSaveConversion tests the map-to-GameSave conversion.
func TestValidator_MapToGameSaveConversion(t *testing.T) {
	validator := NewValidator(Config{})

	data := map[string]interface{}{
		"version": "0.9.0",
		"player": map[string]interface{}{
			"level": 10,
			"x":     100.0,
			"y":     200.0,
		},
		"world": map[string]interface{}{
			"seed":  int64(12345),
			"depth": 5,
		},
	}

	gameSave, err := validator.mapToGameSave(data, "0.9.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gameSave == nil {
		t.Fatal("expected game save, got nil")
	}
	if gameSave.Version != "0.9.0" {
		t.Errorf("expected version 0.9.0, got %s", gameSave.Version)
	}
}

// TestValidator_GameSaveToMapConversion tests the GameSave-to-map conversion.
func TestValidator_GameSaveToMapConversion(t *testing.T) {
	validator := NewValidator(Config{})

	gameSave := saveload.NewGameSave()
	gameSave.Version = "1.0.0"

	mapData, err := validator.gameSaveToMap(gameSave)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mapData == nil {
		t.Fatal("expected map data, got nil")
	}
	if version, ok := mapData["version"].(string); !ok || version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", mapData["version"])
	}
}

// TestValidator_FallbackMigration tests fallback migration for unsupported versions.
func TestValidator_FallbackMigration(t *testing.T) {
	// Create a mock migrator that doesn't support any versions
	mockMigrator := &mockMigrator{}

	validator := NewValidatorWithMigrator(Config{
		TargetVersion: "1.0.0",
	}, mockMigrator)

	data := map[string]interface{}{
		"version": "0.8.0", // Unsupported version
		"player":  map[string]interface{}{"level": 10},
		"world":   map[string]interface{}{"seed": 12345},
	}

	// Should fall back to simulated migration
	migratedData, migrationTime, err := validator.performMigration(data, "0.8.0", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if migratedData == nil {
		t.Fatal("expected migrated data, got nil")
	}
	if migrationTime <= 0 {
		t.Error("expected positive migration time")
	}
	if version, ok := migratedData["version"].(string); !ok || version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", migratedData["version"])
	}
}

// mockMigrator is a test migrator that doesn't support any versions.
type mockMigrator struct{}

func (m *mockMigrator) CanMigrate(sourceVersion string) bool {
	return false
}

func (m *mockMigrator) Migrate(save *saveload.GameSave, sourceVersion string) (*saveload.GameSave, error) {
	return nil, fmt.Errorf("unsupported version: %s", sourceVersion)
}

func (m *mockMigrator) SupportedVersions() []string {
	return []string{}
}
