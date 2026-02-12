package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/venture/pkg/modding"
)

func TestNewFileSystemModRepository(t *testing.T) {
	repo := NewFileSystemModRepository("")
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	if repo.modsDir != "mods" {
		t.Errorf("expected modsDir='mods', got %s", repo.modsDir)
	}

	customRepo := NewFileSystemModRepository("/custom/path")
	if customRepo.modsDir != "/custom/path" {
		t.Errorf("expected modsDir='/custom/path', got %s", customRepo.modsDir)
	}
}

func TestFileSystemModRepository_FetchMods(t *testing.T) {
	tempDir := t.TempDir()

	mod1 := modding.Mod{
		ID:          "test-mod-1",
		Name:        "Test Mod 1",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test description 1",
		Type:        modding.ModTypeRule,
		Rules:       map[string]interface{}{"difficulty": 2.0},
		Enabled:     true,
	}

	mod2 := modding.Mod{
		ID:          "test-mod-2",
		Name:        "Test Mod 2",
		Version:     "2.0.0",
		Author:      "Test Author 2",
		Description: "Test description 2",
		Type:        modding.ModTypeGenerator,
		Enabled:     false,
	}

	createModFile(t, tempDir, "test-mod-1.json", mod1)
	createModFile(t, tempDir, "test-mod-2.json", mod2)

	repo := NewFileSystemModRepository(tempDir)
	mods, err := repo.FetchMods()
	if err != nil {
		t.Fatalf("FetchMods failed: %v", err)
	}

	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}

	for _, mod := range mods {
		if mod.ID != "test-mod-1" && mod.ID != "test-mod-2" {
			t.Errorf("unexpected mod ID: %s", mod.ID)
		}
		if mod.Size == 0 {
			t.Errorf("expected non-zero size for mod %s", mod.ID)
		}
	}
}

func TestFileSystemModRepository_FetchMods_NonExistentDirectory(t *testing.T) {
	repo := NewFileSystemModRepository("/nonexistent/directory")
	mods, err := repo.FetchMods()
	if err != nil {
		t.Errorf("expected no error for nonexistent directory, got: %v", err)
	}

	if len(mods) != 0 {
		t.Errorf("expected 0 mods from nonexistent directory, got %d", len(mods))
	}
}

func TestFileSystemModRepository_FetchMods_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	invalidJSON := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte("{invalid json}"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	validMod := modding.Mod{
		ID:          "valid-mod",
		Name:        "Valid Mod",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Valid",
		Type:        modding.ModTypeRule,
	}
	createModFile(t, tempDir, "valid-mod.json", validMod)

	repo := NewFileSystemModRepository(tempDir)
	mods, err := repo.FetchMods()
	if err != nil {
		t.Fatalf("FetchMods failed: %v", err)
	}

	if len(mods) != 1 {
		t.Errorf("expected 1 valid mod (invalid ignored), got %d", len(mods))
	}

	if mods[0].ID != "valid-mod" {
		t.Errorf("expected valid mod, got %s", mods[0].ID)
	}
}

func TestFileSystemModRepository_FetchMods_SkipsNonJSON(t *testing.T) {
	tempDir := t.TempDir()

	// Create non-JSON files
	os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("readme"), 0o644)
	os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte("config"), 0o644)

	// Create valid mod
	mod := modding.Mod{
		ID:          "valid-mod",
		Name:        "Valid Mod",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Valid",
		Type:        modding.ModTypeRule,
	}
	createModFile(t, tempDir, "valid-mod.json", mod)

	repo := NewFileSystemModRepository(tempDir)
	mods, err := repo.FetchMods()
	if err != nil {
		t.Fatalf("FetchMods failed: %v", err)
	}

	if len(mods) != 1 {
		t.Errorf("expected 1 mod (non-JSON skipped), got %d", len(mods))
	}
}

func TestFileSystemModRepository_DownloadMod(t *testing.T) {
	tempDir := t.TempDir()

	mod := modding.Mod{
		ID:          "download-test",
		Name:        "Download Test Mod",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test",
		Type:        modding.ModTypeRule,
	}
	createModFile(t, tempDir, "download-test.json", mod)

	repo := NewFileSystemModRepository(tempDir)
	repo.FetchMods()

	data, err := repo.DownloadMod("download-test", nil)
	if err != nil {
		t.Fatalf("DownloadMod failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty mod data")
	}

	var loadedMod modding.Mod
	if err := json.Unmarshal(data, &loadedMod); err != nil {
		t.Fatalf("failed to parse downloaded mod: %v", err)
	}

	if loadedMod.ID != "download-test" {
		t.Errorf("expected mod ID 'download-test', got %s", loadedMod.ID)
	}
}

func TestFileSystemModRepository_DownloadMod_WithProgress(t *testing.T) {
	tempDir := t.TempDir()

	mod := modding.Mod{
		ID:          "progress-test",
		Name:        "Progress Test",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Test",
		Type:        modding.ModTypeRule,
	}
	createModFile(t, tempDir, "progress-test.json", mod)

	repo := NewFileSystemModRepository(tempDir)

	progressCalls := 0
	var lastDownloaded, lastTotal int64

	data, err := repo.DownloadMod("progress-test", func(downloaded, total int64) {
		progressCalls++
		lastDownloaded = downloaded
		lastTotal = total
	})
	if err != nil {
		t.Fatalf("DownloadMod failed: %v", err)
	}

	if progressCalls == 0 {
		t.Error("expected progress callback to be called")
	}

	if lastDownloaded != lastTotal {
		t.Errorf("expected final download to equal total, got %d/%d", lastDownloaded, lastTotal)
	}

	if int64(len(data)) != lastTotal {
		t.Errorf("expected data length to match total, got %d vs %d", len(data), lastTotal)
	}
}

func TestFileSystemModRepository_DownloadMod_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewFileSystemModRepository(tempDir)

	_, err := repo.DownloadMod("nonexistent-mod", nil)
	if err == nil {
		t.Error("expected error for nonexistent mod")
	}
}

func TestFileSystemModRepository_GetModDetails(t *testing.T) {
	tempDir := t.TempDir()

	mod := modding.Mod{
		ID:           "details-test",
		Name:         "Details Test Mod",
		Version:      "3.0.0",
		Author:       "Test Author",
		Description:  "Detailed description",
		Type:         modding.ModTypeEvent,
		Dependencies: []string{"dep1", "dep2"},
	}
	createModFile(t, tempDir, "details-test.json", mod)

	repo := NewFileSystemModRepository(tempDir)
	repo.FetchMods()

	details, err := repo.GetModDetails("details-test")
	if err != nil {
		t.Fatalf("GetModDetails failed: %v", err)
	}

	if details.ID != "details-test" {
		t.Errorf("expected ID 'details-test', got %s", details.ID)
	}

	if details.Name != "Details Test Mod" {
		t.Errorf("expected name 'Details Test Mod', got %s", details.Name)
	}

	if details.Version != "3.0.0" {
		t.Errorf("expected version '3.0.0', got %s", details.Version)
	}

	if len(details.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(details.Dependencies))
	}
}

func TestFileSystemModRepository_GetModDetails_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	repo := NewFileSystemModRepository(tempDir)

	_, err := repo.GetModDetails("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent mod")
	}
}

func TestFileSystemModRepository_CategoriesFromModType(t *testing.T) {
	tests := []struct {
		modType  modding.ModType
		expected []string
	}{
		{modding.ModTypeRule, []string{"gameplay", "balance"}},
		{modding.ModTypeGenerator, []string{"content", "generator"}},
		{modding.ModTypeEvent, []string{"events", "gameplay"}},
		{modding.ModType("unknown"), []string{"gameplay"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.modType), func(t *testing.T) {
			categories := categoriesFromModType(tt.modType)
			if len(categories) != len(tt.expected) {
				t.Errorf("expected %d categories, got %d", len(tt.expected), len(categories))
				return
			}
			for i, cat := range categories {
				if cat != tt.expected[i] {
					t.Errorf("expected category %s, got %s", tt.expected[i], cat)
				}
			}
		})
	}
}

func TestFileSystemModRepository_Caching(t *testing.T) {
	tempDir := t.TempDir()

	mod := modding.Mod{
		ID:          "cache-test",
		Name:        "Cache Test",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Test",
		Type:        modding.ModTypeRule,
	}
	createModFile(t, tempDir, "cache-test.json", mod)

	repo := NewFileSystemModRepository(tempDir)

	mods1, err := repo.FetchMods()
	if err != nil {
		t.Fatalf("first FetchMods failed: %v", err)
	}

	if len(mods1) != 1 {
		t.Fatalf("expected 1 mod, got %d", len(mods1))
	}

	details, err := repo.GetModDetails("cache-test")
	if err != nil {
		t.Errorf("GetModDetails should work after FetchMods: %v", err)
	}

	if details.ID != "cache-test" {
		t.Errorf("expected cached mod, got %s", details.ID)
	}
}

func TestFileSystemModRepository_ThreadSafety(t *testing.T) {
	tempDir := t.TempDir()

	for i := 0; i < 10; i++ {
		mod := modding.Mod{
			ID:          "thread-test",
			Name:        "Thread Test",
			Version:     "1.0.0",
			Author:      "Test",
			Description: "Test",
			Type:        modding.ModTypeRule,
		}
		createModFile(t, tempDir, "thread-test.json", mod)
	}

	repo := NewFileSystemModRepository(tempDir)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			repo.FetchMods()
			repo.GetModDetails("thread-test")
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper function to create a mod JSON file
func createModFile(t *testing.T, dir, filename string, mod modding.Mod) {
	data, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal mod: %v", err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("failed to write mod file: %v", err)
	}
}

func BenchmarkFileSystemModRepository_FetchMods(b *testing.B) {
	tempDir := b.TempDir()

	for i := 0; i < 50; i++ {
		mod := modding.Mod{
			ID:          "bench-mod",
			Name:        "Benchmark Mod",
			Version:     "1.0.0",
			Author:      "Test",
			Description: "Test",
			Type:        modding.ModTypeRule,
		}
		data, _ := json.Marshal(mod)
		os.WriteFile(filepath.Join(tempDir, "bench-mod.json"), data, 0o644)
	}

	repo := NewFileSystemModRepository(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.FetchMods()
	}
}

func BenchmarkFileSystemModRepository_DownloadMod(b *testing.B) {
	tempDir := b.TempDir()

	mod := modding.Mod{
		ID:          "bench-download",
		Name:        "Benchmark Download",
		Version:     "1.0.0",
		Author:      "Test",
		Description: "Test",
		Type:        modding.ModTypeRule,
	}
	data, _ := json.Marshal(mod)
	os.WriteFile(filepath.Join(tempDir, "bench-download.json"), data, 0o644)

	repo := NewFileSystemModRepository(tempDir)
	repo.FetchMods()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.DownloadMod("bench-download", nil)
	}
}
