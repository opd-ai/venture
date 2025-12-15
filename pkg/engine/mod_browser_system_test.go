package engine

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewModBrowserSystem(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	if system == nil {
		t.Fatal("expected non-nil system")
	}
	if system.world != world {
		t.Error("expected world to be set")
	}
}

func TestModBrowserSystem_SetRepository(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)
	repo := NewInMemoryModRepository()

	system.SetRepository(repo)

	// Verify repository is set by fetching mods
	if system.repository == nil {
		t.Error("expected repository to be set")
	}
}

func TestModBrowserSystem_SetCallbacks(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	installCalled := false
	uninstallCalled := false

	system.SetInstallCallback(func(modID string, modData []byte) error {
		installCalled = true
		return nil
	})

	system.SetUninstallCallback(func(modID string) error {
		uninstallCalled = true
		return nil
	})

	if system.installCallback == nil {
		t.Error("expected install callback to be set")
	}
	if system.uninstallCallback == nil {
		t.Error("expected uninstall callback to be set")
	}

	// Callbacks will be tested through InstallMod/UninstallMod
	_ = installCalled
	_ = uninstallCalled
}

func TestModBrowserSystem_RefreshRepository(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)
	repo := NewInMemoryModRepository()

	// Add test mods to repository
	repo.AddMod(ModListing{ID: "mod1", Name: "Test Mod 1"}, []byte("data1"))
	repo.AddMod(ModListing{ID: "mod2", Name: "Test Mod 2"}, []byte("data2"))

	system.SetRepository(repo)

	// Create entity with mod browser component
	entity := world.CreateEntity()
	comp := NewModBrowserComponent()
	entity.AddComponent(comp)

	// Trigger refresh
	system.RefreshRepository(comp)

	if !comp.RefreshPending {
		t.Error("expected RefreshPending to be true")
	}

	// Process refresh via Update
	system.Update([]*Entity{entity}, 0.016)

	// Give time for async processing
	time.Sleep(50 * time.Millisecond)

	if comp.RefreshPending {
		t.Error("expected RefreshPending to be false after update")
	}
	if comp.GetAvailableModCount() != 2 {
		t.Errorf("expected 2 available mods, got %d", comp.GetAvailableModCount())
	}
}

func TestModBrowserSystem_RefreshWithoutRepository(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	entity := world.CreateEntity()
	comp := NewModBrowserComponent()
	entity.AddComponent(comp)

	// Trigger refresh without repository
	system.RefreshRepository(comp)
	system.processRefresh(comp)

	if comp.RefreshPending {
		t.Error("expected RefreshPending to be false when no repository")
	}
}

func TestModBrowserSystem_InstallMod(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)
	repo := NewInMemoryModRepository()

	repo.AddMod(ModListing{ID: "mod1", Name: "Test Mod", Size: 1024}, []byte("moddata"))
	system.SetRepository(repo)

	installCalled := false
	var mu sync.Mutex
	system.SetInstallCallback(func(modID string, modData []byte) error {
		mu.Lock()
		installCalled = true
		mu.Unlock()
		return nil
	})

	entity := world.CreateEntity()
	comp := NewModBrowserComponent()
	entity.AddComponent(comp)

	// Refresh to populate available mods
	comp.SetAvailableMods([]ModListing{{ID: "mod1", Name: "Test Mod", Size: 1024}})

	// Install mod
	err := system.InstallMod(comp, "mod1")
	if err != nil {
		t.Errorf("unexpected error installing mod: %v", err)
	}

	// Check download was started
	download, exists := comp.GetDownload("mod1")
	if !exists {
		t.Error("expected download to be started")
	}
	if download.Status != "pending" {
		t.Errorf("expected status 'pending', got %s", download.Status)
	}

	// Process downloads - manually call downloadMod to avoid async timing issues
	system.downloadMod(comp, "mod1")

	// Wait briefly for completion
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	called := installCalled
	mu.Unlock()
	if !called {
		t.Error("expected install callback to be called")
	}

	// Verify mod is installed
	if !comp.IsInstalled("mod1") {
		t.Error("expected mod to be installed after download")
	}
}

func TestModBrowserSystem_InstallAlreadyInstalled(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{{ID: "mod1", Name: "Test Mod", Size: 1024}})
	comp.SetInstalled("mod1", true)

	err := system.InstallMod(comp, "mod1")
	if err == nil {
		t.Error("expected error when installing already installed mod")
	}
}

func TestModBrowserSystem_InstallMissingDependencies(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{
		{ID: "mod1", Name: "Test Mod", Size: 1024, Dependencies: []string{"dep1", "dep2"}},
	})

	err := system.InstallMod(comp, "mod1")
	if err == nil {
		t.Error("expected error when dependencies are missing")
	}
}

func TestModBrowserSystem_InstallModNotFound(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()

	err := system.InstallMod(comp, "nonexistent")
	if err == nil {
		t.Error("expected error when mod not found")
	}
}

func TestModBrowserSystem_UninstallMod(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	uninstallCalled := false
	system.SetUninstallCallback(func(modID string) error {
		uninstallCalled = true
		return nil
	})

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{{ID: "mod1", Name: "Test Mod"}})
	comp.SetInstalled("mod1", true)

	err := system.UninstallMod(comp, "mod1")
	if err != nil {
		t.Errorf("unexpected error uninstalling mod: %v", err)
	}

	if comp.IsInstalled("mod1") {
		t.Error("expected mod to be uninstalled")
	}

	if !uninstallCalled {
		t.Error("expected uninstall callback to be called")
	}
}

func TestModBrowserSystem_UninstallNotInstalled(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()

	err := system.UninstallMod(comp, "mod1")
	if err == nil {
		t.Error("expected error when uninstalling non-installed mod")
	}
}

func TestModBrowserSystem_UninstallWithDependents(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{
		{ID: "base-mod", Name: "Base Mod"},
		{ID: "dependent-mod", Name: "Dependent Mod", Dependencies: []string{"base-mod"}},
	})
	comp.SetInstalled("base-mod", true)
	comp.SetInstalled("dependent-mod", true)

	err := system.UninstallMod(comp, "base-mod")
	if err == nil {
		t.Error("expected error when uninstalling mod with dependents")
	}
}

func TestModBrowserSystem_GetModsByCategory(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{
		{ID: "mod1", Name: "Mod 1", Categories: []string{"gameplay"}},
		{ID: "mod2", Name: "Mod 2", Categories: []string{"graphics"}},
		{ID: "mod3", Name: "Mod 3", Categories: []string{"gameplay"}},
	})

	mods := system.GetModsByCategory(comp, "gameplay")
	if len(mods) != 2 {
		t.Errorf("expected 2 gameplay mods, got %d", len(mods))
	}

	// Test nil component
	mods = system.GetModsByCategory(nil, "gameplay")
	if mods != nil {
		t.Error("expected nil for nil component")
	}
}

func TestModBrowserSystem_SearchMods(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{
		{ID: "mod1", Name: "Combat Overhaul"},
		{ID: "mod2", Name: "Better Graphics"},
		{ID: "mod3", Name: "Combat AI"},
	})

	mods := system.SearchMods(comp, "combat")
	if len(mods) != 2 {
		t.Errorf("expected 2 combat mods, got %d", len(mods))
	}

	// Test nil component
	mods = system.SearchMods(nil, "combat")
	if mods != nil {
		t.Error("expected nil for nil component")
	}
}

func TestModBrowserSystem_GetRecommendedMods(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{
		{ID: "mod1", Name: "Mod 1", Rating: 4.5, Categories: []string{"gameplay"}},
		{ID: "mod2", Name: "Mod 2", Rating: 3.0, Categories: []string{"graphics"}},
		{ID: "mod3", Name: "Mod 3", Rating: 4.0, Categories: []string{"gameplay"}, Featured: true},
		{ID: "mod4", Name: "Mod 4", Rating: 5.0, Categories: []string{"gameplay"}},
	})

	// Install a gameplay mod
	comp.SetInstalled("mod1", true)

	// Get recommendations
	recommended := system.GetRecommendedMods(comp, 2)

	if len(recommended) != 2 {
		t.Errorf("expected 2 recommended mods, got %d", len(recommended))
	}

	// Installed mod should not be in recommendations
	for _, mod := range recommended {
		if mod.ID == "mod1" {
			t.Error("installed mod should not be in recommendations")
		}
	}

	// Test nil component
	recommended = system.GetRecommendedMods(nil, 5)
	if recommended != nil {
		t.Error("expected nil for nil component")
	}
}

func TestModBrowserSystem_CheckGameVersionCompatibility(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	tests := []struct {
		name        string
		mod         *ModListing
		gameVersion string
		compatible  bool
	}{
		{
			name:        "compatible version",
			mod:         &ModListing{GameVersion: "10.0.0"},
			gameVersion: "10.0.0",
			compatible:  true,
		},
		{
			name:        "newer game version",
			mod:         &ModListing{GameVersion: "10.0.0"},
			gameVersion: "11.0.0",
			compatible:  true,
		},
		{
			name:        "older game version",
			mod:         &ModListing{GameVersion: "11.0.0"},
			gameVersion: "10.0.0",
			compatible:  false,
		},
		{
			name:        "no version requirement",
			mod:         &ModListing{GameVersion: ""},
			gameVersion: "10.0.0",
			compatible:  true,
		},
		{
			name:        "nil mod",
			mod:         nil,
			gameVersion: "10.0.0",
			compatible:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.CheckGameVersionCompatibility(tt.mod, tt.gameVersion)
			if result != tt.compatible {
				t.Errorf("expected %v, got %v", tt.compatible, result)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"10.0.0", "9.0.0", 1},
		{"1.0", "1.0.0", 0},
		{"1", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.v1, tt.v2), func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestInMemoryModRepository(t *testing.T) {
	repo := NewInMemoryModRepository()

	if repo == nil {
		t.Fatal("expected non-nil repository")
	}

	// Test AddMod and FetchMods
	repo.AddMod(ModListing{ID: "mod1", Name: "Test Mod 1"}, []byte("data1"))
	repo.AddMod(ModListing{ID: "mod2", Name: "Test Mod 2"}, []byte("data2"))

	mods, err := repo.FetchMods()
	if err != nil {
		t.Errorf("unexpected error fetching mods: %v", err)
	}
	if len(mods) != 2 {
		t.Errorf("expected 2 mods, got %d", len(mods))
	}

	// Test GetModDetails
	mod, err := repo.GetModDetails("mod1")
	if err != nil {
		t.Errorf("unexpected error getting mod details: %v", err)
	}
	if mod.Name != "Test Mod 1" {
		t.Errorf("expected 'Test Mod 1', got %s", mod.Name)
	}

	// Test GetModDetails not found
	_, err = repo.GetModDetails("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent mod")
	}

	// Test DownloadMod
	var progressCalled bool
	data, err := repo.DownloadMod("mod1", func(downloaded, total int64) {
		progressCalled = true
	})
	if err != nil {
		t.Errorf("unexpected error downloading mod: %v", err)
	}
	if string(data) != "data1" {
		t.Errorf("expected 'data1', got %s", string(data))
	}
	if !progressCalled {
		t.Error("expected progress callback to be called")
	}

	// Test DownloadMod not found
	_, err = repo.DownloadMod("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent mod download")
	}
}

func TestGenerateModListing(t *testing.T) {
	// Test determinism
	mod1 := GenerateModListing(12345)
	mod2 := GenerateModListing(12345)

	if mod1.ID != mod2.ID {
		t.Error("expected same ID for same seed")
	}
	if mod1.Name != mod2.Name {
		t.Error("expected same Name for same seed")
	}
	if mod1.Rating != mod2.Rating {
		t.Error("expected same Rating for same seed")
	}

	// Test different seeds produce different results
	mod3 := GenerateModListing(99999)
	if mod1.ID == mod3.ID {
		t.Error("expected different ID for different seed")
	}
}

func TestSerializeModBrowserState(t *testing.T) {
	comp := NewModBrowserComponent()
	comp.SetInstalled("mod1", true)
	comp.SetInstalled("mod2", true)
	comp.LastRefresh = 12345

	data, err := SerializeModBrowserState(comp)
	if err != nil {
		t.Errorf("unexpected error serializing: %v", err)
	}

	// Test nil component
	_, err = SerializeModBrowserState(nil)
	if err == nil {
		t.Error("expected error for nil component")
	}

	// Deserialize
	comp2 := NewModBrowserComponent()
	err = DeserializeModBrowserState(comp2, data)
	if err != nil {
		t.Errorf("unexpected error deserializing: %v", err)
	}

	if !comp2.IsInstalled("mod1") || !comp2.IsInstalled("mod2") {
		t.Error("expected installed mods to be restored")
	}
	if comp2.LastRefresh != 12345 {
		t.Errorf("expected LastRefresh 12345, got %d", comp2.LastRefresh)
	}

	// Test nil component for deserialize
	err = DeserializeModBrowserState(nil, data)
	if err == nil {
		t.Error("expected error for nil component")
	}
}

func TestModBrowserSystem_InstallWithNilComponent(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	err := system.InstallMod(nil, "mod1")
	if err == nil {
		t.Error("expected error for nil component")
	}
}

func TestModBrowserSystem_UninstallWithNilComponent(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	err := system.UninstallMod(nil, "mod1")
	if err == nil {
		t.Error("expected error for nil component")
	}
}

func TestModBrowserSystem_RefreshWithNilComponent(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	// Should not panic
	system.RefreshRepository(nil)
}

func TestModBrowserSystem_UpdateWithoutBrowserComponent(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	// Create entity without mod browser component
	entity := world.CreateEntity()

	// Should not panic
	system.Update([]*Entity{entity}, 0.016)
}

func TestModBrowserSystem_DownloadFailure(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	// Repository that always fails downloads
	failingRepo := &failingModRepository{}
	system.SetRepository(failingRepo)

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{{ID: "mod1", Name: "Test Mod", Size: 1024}})

	// Start install
	err := system.InstallMod(comp, "mod1")
	if err != nil {
		t.Errorf("unexpected error starting install: %v", err)
	}

	// Manually process download
	system.downloadMod(comp, "mod1")

	// Check download failed
	download, _ := comp.GetDownload("mod1")
	if download.Status != "failed" {
		t.Errorf("expected status 'failed', got %s", download.Status)
	}
}

// failingModRepository is a mock repository that fails all operations.
type failingModRepository struct{}

func (r *failingModRepository) FetchMods() ([]ModListing, error) {
	return nil, fmt.Errorf("fetch failed")
}

func (r *failingModRepository) DownloadMod(modID string, progressCallback func(downloaded, total int64)) ([]byte, error) {
	return nil, fmt.Errorf("download failed")
}

func (r *failingModRepository) GetModDetails(modID string) (*ModListing, error) {
	return nil, fmt.Errorf("get details failed")
}

func TestModBrowserSystem_InstallCallbackFailure(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)
	repo := NewInMemoryModRepository()

	repo.AddMod(ModListing{ID: "mod1", Name: "Test Mod", Size: 1024}, []byte("moddata"))
	system.SetRepository(repo)

	// Callback that fails
	system.SetInstallCallback(func(modID string, modData []byte) error {
		return fmt.Errorf("install callback failed")
	})

	comp := NewModBrowserComponent()
	comp.SetAvailableMods([]ModListing{{ID: "mod1", Name: "Test Mod", Size: 1024}})

	// Start install
	system.InstallMod(comp, "mod1")

	// Manually process download
	system.downloadMod(comp, "mod1")

	// Check download failed
	download, _ := comp.GetDownload("mod1")
	if download.Status != "failed" {
		t.Errorf("expected status 'failed', got %s", download.Status)
	}
}

func TestModBrowserSystem_UninstallCallbackFailure(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	// Callback that fails
	system.SetUninstallCallback(func(modID string) error {
		return fmt.Errorf("uninstall callback failed")
	})

	comp := NewModBrowserComponent()
	comp.SetInstalled("mod1", true)

	err := system.UninstallMod(comp, "mod1")
	if err == nil {
		t.Error("expected error when uninstall callback fails")
	}

	// Mod should still be installed since uninstall failed
	if !comp.IsInstalled("mod1") {
		t.Error("expected mod to still be installed after callback failure")
	}
}
