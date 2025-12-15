package engine

import (
	"testing"
)

func TestNewModBrowserComponent(t *testing.T) {
	comp := NewModBrowserComponent()

	if comp == nil {
		t.Fatal("expected non-nil component")
	}
	if comp.AvailableMods == nil {
		t.Error("expected non-nil AvailableMods")
	}
	if comp.InstalledMods == nil {
		t.Error("expected non-nil InstalledMods")
	}
	if comp.Categories == nil || len(comp.Categories) == 0 {
		t.Error("expected default categories")
	}
	if comp.SortBy != ModSortByRating {
		t.Errorf("expected default SortBy to be rating, got %s", comp.SortBy)
	}
	if !comp.SortDescending {
		t.Error("expected default SortDescending to be true")
	}
	if comp.Downloads == nil {
		t.Error("expected non-nil Downloads map")
	}
}

func TestModBrowserComponent_Type(t *testing.T) {
	comp := NewModBrowserComponent()
	if comp.Type() != "mod_browser" {
		t.Errorf("expected type 'mod_browser', got %s", comp.Type())
	}
}

func TestModBrowserComponent_SetAvailableMods(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Test Mod 1", Categories: []string{"gameplay", "new-category"}},
		{ID: "mod2", Name: "Test Mod 2", Categories: []string{"graphics"}},
	}

	comp.SetAvailableMods(mods)

	if len(comp.AvailableMods) != 2 {
		t.Errorf("expected 2 available mods, got %d", len(comp.AvailableMods))
	}

	if comp.LastRefresh == 0 {
		t.Error("expected LastRefresh to be set")
	}

	// Check that new category was added
	found := false
	for _, cat := range comp.Categories {
		if cat == "new-category" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected new-category to be added to categories")
	}
}

func TestModBrowserComponent_GetFilteredMods(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Combat Overhaul", Rating: 4.5, Categories: []string{"gameplay"}},
		{ID: "mod2", Name: "Better Graphics", Rating: 3.5, Categories: []string{"graphics"}},
		{ID: "mod3", Name: "Combat AI", Rating: 4.0, Categories: []string{"gameplay"}},
	}
	comp.SetAvailableMods(mods)

	// Test no filter
	filtered := comp.GetFilteredMods()
	if len(filtered) != 3 {
		t.Errorf("expected 3 mods without filter, got %d", len(filtered))
	}

	// Test search filter
	comp.SetSearchQuery("combat")
	filtered = comp.GetFilteredMods()
	if len(filtered) != 2 {
		t.Errorf("expected 2 mods matching 'combat', got %d", len(filtered))
	}

	// Test category filter
	comp.SetSearchQuery("")
	comp.SetActiveCategory("gameplay")
	filtered = comp.GetFilteredMods()
	if len(filtered) != 2 {
		t.Errorf("expected 2 gameplay mods, got %d", len(filtered))
	}

	// Test combined filters
	comp.SetSearchQuery("combat")
	comp.SetActiveCategory("gameplay")
	filtered = comp.GetFilteredMods()
	if len(filtered) != 2 {
		t.Errorf("expected 2 combat gameplay mods, got %d", len(filtered))
	}
}

func TestModBrowserComponent_Sorting(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Zeta Mod", Rating: 3.0, Downloads: 100, UploadedAt: 1000},
		{ID: "mod2", Name: "Alpha Mod", Rating: 4.5, Downloads: 500, UploadedAt: 2000},
		{ID: "mod3", Name: "Beta Mod", Rating: 4.0, Downloads: 200, UploadedAt: 3000},
	}
	comp.SetAvailableMods(mods)

	// Test sort by rating (descending by default)
	comp.SetSortBy(ModSortByRating, true)
	filtered := comp.GetFilteredMods()
	if filtered[0].ID != "mod2" {
		t.Errorf("expected highest rated mod first, got %s", filtered[0].ID)
	}

	// Test sort by name ascending
	comp.SetSortBy(ModSortByName, false)
	filtered = comp.GetFilteredMods()
	if filtered[0].ID != "mod2" { // Alpha Mod
		t.Errorf("expected Alpha Mod first when sorting by name ascending, got %s", filtered[0].Name)
	}

	// Test sort by downloads descending
	comp.SetSortBy(ModSortByDownloads, true)
	filtered = comp.GetFilteredMods()
	if filtered[0].ID != "mod2" { // 500 downloads
		t.Errorf("expected most downloaded mod first, got %s", filtered[0].ID)
	}

	// Test sort by date descending
	comp.SetSortBy(ModSortByDate, true)
	filtered = comp.GetFilteredMods()
	if filtered[0].ID != "mod3" { // newest
		t.Errorf("expected newest mod first, got %s", filtered[0].ID)
	}
}

func TestModBrowserComponent_GetMod(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Test Mod"},
	}
	comp.SetAvailableMods(mods)

	mod, found := comp.GetMod("mod1")
	if !found {
		t.Error("expected to find mod1")
	}
	if mod.Name != "Test Mod" {
		t.Errorf("expected name 'Test Mod', got %s", mod.Name)
	}

	_, found = comp.GetMod("nonexistent")
	if found {
		t.Error("expected not to find nonexistent mod")
	}
}

func TestModBrowserComponent_Installation(t *testing.T) {
	comp := NewModBrowserComponent()

	// Test IsInstalled on empty
	if comp.IsInstalled("mod1") {
		t.Error("expected mod1 not to be installed initially")
	}

	// Test SetInstalled
	comp.SetInstalled("mod1", true)
	if !comp.IsInstalled("mod1") {
		t.Error("expected mod1 to be installed after SetInstalled(true)")
	}

	// Test uninstall
	comp.SetInstalled("mod1", false)
	if comp.IsInstalled("mod1") {
		t.Error("expected mod1 not to be installed after SetInstalled(false)")
	}

	// Test GetInstalledModCount
	comp.SetInstalled("mod1", true)
	comp.SetInstalled("mod2", true)
	if comp.GetInstalledModCount() != 2 {
		t.Errorf("expected 2 installed mods, got %d", comp.GetInstalledModCount())
	}

	// Test GetInstalledMods
	installed := comp.GetInstalledMods()
	if len(installed) != 2 {
		t.Errorf("expected 2 installed mod IDs, got %d", len(installed))
	}
}

func TestModBrowserComponent_Downloads(t *testing.T) {
	comp := NewModBrowserComponent()

	// Test StartDownload
	err := comp.StartDownload("mod1", 1000)
	if err != nil {
		t.Errorf("unexpected error starting download: %v", err)
	}

	// Test duplicate download
	err = comp.StartDownload("mod1", 1000)
	if err == nil {
		t.Error("expected error starting duplicate download")
	}

	// Test GetDownload
	download, found := comp.GetDownload("mod1")
	if !found {
		t.Error("expected to find download")
	}
	if download.Status != "pending" {
		t.Errorf("expected status 'pending', got %s", download.Status)
	}
	if download.TotalBytes != 1000 {
		t.Errorf("expected total bytes 1000, got %d", download.TotalBytes)
	}

	// Test UpdateDownloadProgress
	err = comp.UpdateDownloadProgress("mod1", 500)
	if err != nil {
		t.Errorf("unexpected error updating progress: %v", err)
	}
	download, _ = comp.GetDownload("mod1")
	if download.Progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", download.Progress)
	}
	if download.Status != "downloading" {
		t.Errorf("expected status 'downloading', got %s", download.Status)
	}

	// Test SetDownloadStatus
	err = comp.SetDownloadStatus("mod1", "installing", "")
	if err != nil {
		t.Errorf("unexpected error setting status: %v", err)
	}
	download, _ = comp.GetDownload("mod1")
	if download.Status != "installing" {
		t.Errorf("expected status 'installing', got %s", download.Status)
	}

	// Test CompleteDownload
	comp.CompleteDownload("mod1")
	download, _ = comp.GetDownload("mod1")
	if download.Status != "complete" {
		t.Errorf("expected status 'complete', got %s", download.Status)
	}
	if !comp.IsInstalled("mod1") {
		t.Error("expected mod to be marked as installed after complete")
	}

	// Test CancelDownload
	comp.StartDownload("mod2", 500)
	comp.CancelDownload("mod2")
	_, found = comp.GetDownload("mod2")
	if found {
		t.Error("expected download to be cancelled")
	}

	// Test GetActiveDownloads
	comp.StartDownload("mod3", 200)
	downloads := comp.GetActiveDownloads()
	if len(downloads) != 2 { // mod1 (complete) and mod3 (pending)
		t.Errorf("expected 2 active downloads, got %d", len(downloads))
	}
}

func TestModBrowserComponent_DownloadErrors(t *testing.T) {
	comp := NewModBrowserComponent()

	// Test UpdateDownloadProgress on nonexistent
	err := comp.UpdateDownloadProgress("nonexistent", 100)
	if err == nil {
		t.Error("expected error updating progress for nonexistent download")
	}

	// Test SetDownloadStatus on nonexistent
	err = comp.SetDownloadStatus("nonexistent", "complete", "")
	if err == nil {
		t.Error("expected error setting status for nonexistent download")
	}
}

func TestModBrowserComponent_GetFeaturedMods(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Regular Mod", Featured: false},
		{ID: "mod2", Name: "Featured Mod", Featured: true},
		{ID: "mod3", Name: "Another Featured", Featured: true},
	}
	comp.SetAvailableMods(mods)

	featured := comp.GetFeaturedMods()
	if len(featured) != 2 {
		t.Errorf("expected 2 featured mods, got %d", len(featured))
	}
}

func TestModBrowserComponent_CheckDependencies(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "base-mod", Name: "Base Mod", Dependencies: []string{}},
		{ID: "dependent-mod", Name: "Dependent Mod", Dependencies: []string{"base-mod", "other-mod"}},
	}
	comp.SetAvailableMods(mods)

	// Check with no mods installed
	missing, ok := comp.CheckDependencies("dependent-mod")
	if ok {
		t.Error("expected dependencies check to fail with no mods installed")
	}
	if len(missing) != 2 {
		t.Errorf("expected 2 missing dependencies, got %d", len(missing))
	}

	// Install one dependency
	comp.SetInstalled("base-mod", true)
	missing, ok = comp.CheckDependencies("dependent-mod")
	if ok {
		t.Error("expected dependencies check to fail with one missing")
	}
	if len(missing) != 1 {
		t.Errorf("expected 1 missing dependency, got %d", len(missing))
	}

	// Install all dependencies
	comp.SetInstalled("other-mod", true)
	missing, ok = comp.CheckDependencies("dependent-mod")
	if !ok {
		t.Error("expected dependencies check to pass with all installed")
	}
	if len(missing) != 0 {
		t.Errorf("expected 0 missing dependencies, got %d", len(missing))
	}

	// Check nonexistent mod
	_, ok = comp.CheckDependencies("nonexistent")
	if ok {
		t.Error("expected dependencies check to fail for nonexistent mod")
	}
}

func TestModBrowserComponent_Serialization(t *testing.T) {
	comp := NewModBrowserComponent()

	// Set up state
	comp.SetInstalled("mod1", true)
	comp.SetInstalled("mod2", true)
	comp.SetSearchQuery("test")
	comp.SetSortBy(ModSortByDownloads, false)
	comp.SetActiveCategory("gameplay")
	comp.LastRefresh = 12345

	// Serialize
	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	// Create new component and deserialize
	comp2 := NewModBrowserComponent()
	err = comp2.Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error deserializing: %v", err)
	}

	// Verify state
	if !comp2.IsInstalled("mod1") || !comp2.IsInstalled("mod2") {
		t.Error("expected installed mods to be restored")
	}
	if comp2.SearchQuery != "test" {
		t.Errorf("expected search query 'test', got %s", comp2.SearchQuery)
	}
	if comp2.SortBy != ModSortByDownloads {
		t.Errorf("expected sort by downloads, got %s", comp2.SortBy)
	}
	if comp2.SortDescending {
		t.Error("expected sort descending to be false")
	}
	if comp2.ActiveCategory != "gameplay" {
		t.Errorf("expected active category 'gameplay', got %s", comp2.ActiveCategory)
	}
	if comp2.LastRefresh != 12345 {
		t.Errorf("expected last refresh 12345, got %d", comp2.LastRefresh)
	}
}

func TestModBrowserComponent_GetAvailableModCount(t *testing.T) {
	comp := NewModBrowserComponent()

	if comp.GetAvailableModCount() != 0 {
		t.Errorf("expected 0 available mods initially, got %d", comp.GetAvailableModCount())
	}

	mods := []ModListing{
		{ID: "mod1"},
		{ID: "mod2"},
		{ID: "mod3"},
	}
	comp.SetAvailableMods(mods)

	if comp.GetAvailableModCount() != 3 {
		t.Errorf("expected 3 available mods, got %d", comp.GetAvailableModCount())
	}
}

func TestModBrowserComponent_SearchByAuthor(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Mod One", Author: "John"},
		{ID: "mod2", Name: "Mod Two", Author: "Jane"},
	}
	comp.SetAvailableMods(mods)

	comp.SetSearchQuery("john")
	filtered := comp.GetFilteredMods()
	if len(filtered) != 1 {
		t.Errorf("expected 1 mod by author search, got %d", len(filtered))
	}
	if filtered[0].ID != "mod1" {
		t.Errorf("expected mod1 in author search results, got %s", filtered[0].ID)
	}
}

func TestModBrowserComponent_SortByUpdated(t *testing.T) {
	comp := NewModBrowserComponent()

	mods := []ModListing{
		{ID: "mod1", Name: "Mod One", UpdatedAt: 1000},
		{ID: "mod2", Name: "Mod Two", UpdatedAt: 3000},
		{ID: "mod3", Name: "Mod Three", UpdatedAt: 2000},
	}
	comp.SetAvailableMods(mods)

	comp.SetSortBy(ModSortByUpdated, true)
	filtered := comp.GetFilteredMods()
	if filtered[0].ID != "mod2" {
		t.Errorf("expected most recently updated mod first, got %s", filtered[0].ID)
	}
}

func TestDefaultModCategories(t *testing.T) {
	cats := defaultModCategories()
	if len(cats) == 0 {
		t.Error("expected non-empty default categories")
	}

	// Check some expected categories
	expectedCats := []string{"gameplay", "graphics", "audio", "ui"}
	for _, expected := range expectedCats {
		found := false
		for _, cat := range cats {
			if cat == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected category %s in default categories", expected)
		}
	}
}
