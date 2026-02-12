package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewFileSystemFileWatcher(t *testing.T) {
	watcher := NewFileSystemFileWatcher("test-mods")
	if watcher.modsDir != "test-mods" {
		t.Errorf("Expected modsDir 'test-mods', got '%s'", watcher.modsDir)
	}
	if watcher.cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

func TestNewFileSystemFileWatcher_DefaultDirectory(t *testing.T) {
	watcher := NewFileSystemFileWatcher("")
	if watcher.modsDir != "mods" {
		t.Errorf("Expected default modsDir 'mods', got '%s'", watcher.modsDir)
	}
}

func TestFileSystemFileWatcher_GetFileHash(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	// Create test mod file
	modID := "test-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"name":    "Test Mod",
		"version": "1.0.0",
		"type":    "rule",
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Get file hash
	hash1, err := watcher.GetFileHash(modID)
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash1 == "" {
		t.Error("Expected non-empty hash")
	}

	// Second call should return same hash (cached)
	hash2, err := watcher.GetFileHash(modID)
	if err != nil {
		t.Fatalf("GetFileHash (cached) failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("Expected cached hash to match: %s != %s", hash1, hash2)
	}
}

func TestFileSystemFileWatcher_GetFileHash_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	_, err := watcher.GetFileHash("nonexistent-mod")
	if err == nil {
		t.Error("Expected error for nonexistent mod")
	}
}

func TestFileSystemFileWatcher_GetModData(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "data-test-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"name":    "Data Test Mod",
		"version": "2.0.0",
		"type":    "rule",
		"rules": map[string]interface{}{
			"difficulty_multiplier": 1.5,
		},
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Get mod data
	data, err := watcher.GetModData(modID)
	if err != nil {
		t.Fatalf("GetModData failed: %v", err)
	}

	// Parse and verify data
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse returned data: %v", err)
	}
	if parsed["id"] != modID {
		t.Errorf("Expected id '%s', got '%v'", modID, parsed["id"])
	}
	if parsed["version"] != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%v'", parsed["version"])
	}
}

func TestFileSystemFileWatcher_GetModData_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	_, err := watcher.GetModData("missing-mod")
	if err == nil {
		t.Error("Expected error for missing mod")
	}
}

func TestFileSystemFileWatcher_GetModVersion(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "version-test-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"name":    "Version Test Mod",
		"version": "3.2.1",
		"type":    "rule",
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Get mod version
	version, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("GetModVersion failed: %v", err)
	}
	if version != "3.2.1" {
		t.Errorf("Expected version '3.2.1', got '%s'", version)
	}

	// Second call should return cached version
	version2, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("GetModVersion (cached) failed: %v", err)
	}
	if version != version2 {
		t.Errorf("Expected cached version to match: %s != %s", version, version2)
	}
}

func TestFileSystemFileWatcher_GetModVersion_DefaultVersion(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "no-version-mod"
	modData := map[string]interface{}{
		"id":   modID,
		"name": "No Version Mod",
		"type": "rule",
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Should return default version when not specified
	version, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("GetModVersion failed: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", version)
	}
}

func TestFileSystemFileWatcher_InvalidateCache(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "invalidate-test-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"version": "1.0.0",
		"type":    "rule",
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Get version to populate cache
	_, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("Initial GetModVersion failed: %v", err)
	}

	// Verify cache populated
	watcher.mu.RLock()
	_, inCache := watcher.cache[modID]
	watcher.mu.RUnlock()
	if !inCache {
		t.Error("Expected mod to be in cache")
	}

	// Invalidate cache
	watcher.InvalidateCache(modID)

	// Verify cache cleared
	watcher.mu.RLock()
	_, stillInCache := watcher.cache[modID]
	watcher.mu.RUnlock()
	if stillInCache {
		t.Error("Expected mod to be removed from cache")
	}

	// Verify we can still read from filesystem
	version, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("GetModVersion after invalidate failed: %v", err)
	}
	if version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", version)
	}
}

func TestFileSystemFileWatcher_InvalidateAllCache(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	// Create multiple mods
	modIDs := []string{"mod1", "mod2", "mod3"}
	for _, modID := range modIDs {
		modData := map[string]interface{}{
			"id":      modID,
			"version": "1.0.0",
			"type":    "rule",
		}
		createTestModFile(t, tmpDir, modID, modData)
		// Populate cache
		_, _ = watcher.GetModVersion(modID)
	}

	// Verify all in cache
	watcher.mu.RLock()
	cacheSize := len(watcher.cache)
	watcher.mu.RUnlock()
	if cacheSize != 3 {
		t.Errorf("Expected 3 items in cache, got %d", cacheSize)
	}

	// Invalidate all
	watcher.InvalidateAllCache()

	// Verify cache empty
	watcher.mu.RLock()
	cacheSize = len(watcher.cache)
	watcher.mu.RUnlock()
	if cacheSize != 0 {
		t.Errorf("Expected empty cache, got %d items", cacheSize)
	}
}

func TestFileSystemFileWatcher_HashChangesWithContent(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "hash-change-mod"
	modData1 := map[string]interface{}{
		"id":      modID,
		"version": "1.0.0",
		"type":    "rule",
	}
	createTestModFile(t, tmpDir, modID, modData1)

	// Get initial hash
	hash1, err := watcher.GetFileHash(modID)
	if err != nil {
		t.Fatalf("Initial GetFileHash failed: %v", err)
	}

	// Update mod file
	modData2 := map[string]interface{}{
		"id":      modID,
		"version": "2.0.0",
		"type":    "rule",
	}
	createTestModFile(t, tmpDir, modID, modData2)

	// Invalidate cache to force re-read
	watcher.InvalidateCache(modID)

	// Get new hash
	hash2, err := watcher.GetFileHash(modID)
	if err != nil {
		t.Fatalf("Updated GetFileHash failed: %v", err)
	}

	// Hashes should differ
	if hash1 == hash2 {
		t.Error("Expected hash to change after file update")
	}
}

func TestFileSystemFileWatcher_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	// Create test mods
	for i := 0; i < 10; i++ {
		modID := filepath.Base(tmpDir) + "-concurrent-" + string(rune('a'+i))
		modData := map[string]interface{}{
			"id":      modID,
			"version": "1.0.0",
			"type":    "rule",
		}
		createTestModFile(t, tmpDir, modID, modData)
	}

	// Concurrent access
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)
		modID := filepath.Base(tmpDir) + "-concurrent-" + string(rune('a'+i))

		// GetFileHash
		go func(id string) {
			defer wg.Done()
			_, _ = watcher.GetFileHash(id)
		}(modID)

		// GetModData
		go func(id string) {
			defer wg.Done()
			_, _ = watcher.GetModData(id)
		}(modID)

		// GetModVersion
		go func(id string) {
			defer wg.Done()
			_, _ = watcher.GetModVersion(id)
		}(modID)
	}

	wg.Wait()
	// Test passes if no race conditions detected
}

func TestFileSystemFileWatcher_SetLogger(t *testing.T) {
	watcher := NewFileSystemFileWatcher("test-mods")
	customLogger := logrus.New()
	customLogger.SetLevel(logrus.DebugLevel)

	watcher.SetLogger(customLogger)

	watcher.mu.RLock()
	if watcher.logger != customLogger {
		t.Error("Expected custom logger to be set")
	}
	watcher.mu.RUnlock()
}

func TestFileSystemFileWatcher_IntegrationWithHotReloadSystem(t *testing.T) {
	tmpDir := t.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	// Create test mod
	modID := "hot-reload-integration-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"name":    "Hot Reload Integration Test",
		"version": "1.5.0",
		"type":    "rule",
		"rules": map[string]interface{}{
			"test_value": 42,
		},
	}
	createTestModFile(t, tmpDir, modID, modData)

	// Test FileWatcher interface compliance
	var _ FileWatcher = watcher

	// Verify all interface methods work
	hash, err := watcher.GetFileHash(modID)
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	data, err := watcher.GetModData(modID)
	if err != nil {
		t.Fatalf("GetModData failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	version, err := watcher.GetModVersion(modID)
	if err != nil {
		t.Fatalf("GetModVersion failed: %v", err)
	}
	if version != "1.5.0" {
		t.Errorf("Expected version '1.5.0', got '%s'", version)
	}
}

// Helper function to create test mod JSON files
func createTestModFile(t *testing.T, dir, modID string, data map[string]interface{}) {
	t.Helper()
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test mod data: %v", err)
	}
	filename := filepath.Join(dir, modID+".json")
	if err := os.WriteFile(filename, jsonData, 0o644); err != nil {
		t.Fatalf("Failed to write test mod file: %v", err)
	}
}

func BenchmarkFileSystemFileWatcher_GetFileHash(b *testing.B) {
	tmpDir := b.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "benchmark-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"version": "1.0.0",
		"type":    "rule",
	}
	jsonData, _ := json.Marshal(modData)
	filename := filepath.Join(tmpDir, modID+".json")
	_ = os.WriteFile(filename, jsonData, 0o644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = watcher.GetFileHash(modID)
	}
}

func BenchmarkFileSystemFileWatcher_GetModData(b *testing.B) {
	tmpDir := b.TempDir()
	watcher := NewFileSystemFileWatcher(tmpDir)

	modID := "benchmark-data-mod"
	modData := map[string]interface{}{
		"id":      modID,
		"version": "1.0.0",
		"type":    "rule",
		"rules": map[string]interface{}{
			"value1": 100,
			"value2": 200,
		},
	}
	jsonData, _ := json.Marshal(modData)
	filename := filepath.Join(tmpDir, modID+".json")
	_ = os.WriteFile(filename, jsonData, 0o644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = watcher.GetModData(modID)
	}
}
