package housing

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewBlueprint(t *testing.T) {
	buildingDef := &BuildingDefinition{
		Type:   0, // TypeHouse
		Style:  0, // StyleMedieval
		Width:  24,
		Height: 24,
		Floors: 2,
		Seed:   12345,
	}

	bp := NewBlueprintWithTime("Test House", "player1", "fantasy", buildingDef, defaultTimeProvider)

	if bp.ID == "" {
		t.Error("Blueprint ID should not be empty")
	}
	if bp.Name != "Test House" {
		t.Errorf("Name = %v, want Test House", bp.Name)
	}
	if bp.Author != "player1" {
		t.Errorf("Author = %v, want player1", bp.Author)
	}
	if bp.GenreID != "fantasy" {
		t.Errorf("GenreID = %v, want fantasy", bp.GenreID)
	}
	if bp.BuildingDef != buildingDef {
		t.Error("BuildingDef should match provided definition")
	}
	if bp.GetRating() != 0.0 {
		t.Errorf("Initial Rating = %v, want 0.0", bp.GetRating())
	}
	if bp.GetRatingCount() != 0 {
		t.Errorf("Initial RatingCount = %v, want 0", bp.GetRatingCount())
	}
	if bp.GetDownloads() != 0 {
		t.Errorf("Initial Downloads = %v, want 0", bp.GetDownloads())
	}
	if bp.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if bp.ModifiedAt.IsZero() {
		t.Error("ModifiedAt should be set")
	}
}

func TestBlueprintAddRating(t *testing.T) {
	tests := []struct {
		name        string
		ratings     []float64
		wantAvg     float64
		wantCount   int
		expectError bool
	}{
		{
			name:        "single rating",
			ratings:     []float64{4.5},
			wantAvg:     4.5,
			wantCount:   1,
			expectError: false,
		},
		{
			name:        "multiple ratings",
			ratings:     []float64{5.0, 4.0, 3.0},
			wantAvg:     4.0,
			wantCount:   3,
			expectError: false,
		},
		{
			name:        "rating too high",
			ratings:     []float64{5.5},
			expectError: true,
		},
		{
			name:        "rating too low",
			ratings:     []float64{-1.0},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBlueprintWithTime("Test", "author", "fantasy", nil, defaultTimeProvider)
			var err error
			for _, rating := range tt.ratings {
				err = bp.AddRating(rating)
				if err != nil && !tt.expectError {
					t.Fatalf("AddRating(%v) unexpected error: %v", rating, err)
				}
				if tt.expectError && err != nil {
					return // Expected error, test passed
				}
			}

			if tt.expectError && err == nil {
				t.Fatal("Expected error but got none")
			}

			if !tt.expectError {
				if bp.GetRatingCount() != tt.wantCount {
					t.Errorf("RatingCount = %v, want %v", bp.GetRatingCount(), tt.wantCount)
				}
				// Allow small floating point difference
				if diff := bp.GetRating() - tt.wantAvg; diff > 0.001 || diff < -0.001 {
					t.Errorf("Rating = %v, want %v", bp.GetRating(), tt.wantAvg)
				}
			}
		})
	}
}

func TestBlueprintIncrementDownloads(t *testing.T) {
	bp := NewBlueprintWithTime("Test", "author", "fantasy", nil, defaultTimeProvider)

	if bp.GetDownloads() != 0 {
		t.Errorf("Initial downloads = %v, want 0", bp.GetDownloads())
	}

	bp.IncrementDownloads()
	if bp.GetDownloads() != 1 {
		t.Errorf("Downloads after increment = %v, want 1", bp.GetDownloads())
	}

	bp.IncrementDownloads()
	bp.IncrementDownloads()
	if bp.GetDownloads() != 3 {
		t.Errorf("Downloads after 3 increments = %v, want 3", bp.GetDownloads())
	}
}

func TestBlueprintExportImport(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "test_blueprint.json.gz")

	buildingDef := &BuildingDefinition{
		Type:   1,
		Style:  2,
		Width:  24,
		Height: 24,
		Floors: 3,
		Seed:   67890,
	}

	original := NewBlueprintWithTime("Test Manor", "player123", "fantasy", buildingDef, defaultTimeProvider)
	original.Description = "A beautiful test manor"
	original.Tags = []string{"medieval", "manor", "large"}
	original.AddRating(4.5)
	original.AddRating(4.5) // Average will be 4.5
	for i := 0; i < 10; i++ {
		original.IncrementDownloads()
	}
	for i := 0; i < 15; i++ {
		original.IncrementDownloads()
	}

	// Export
	err := original.Export(filepath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Import
	imported, err := ImportBlueprint(filepath)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify all fields match
	if imported.ID != original.ID {
		t.Errorf("ID = %v, want %v", imported.ID, original.ID)
	}
	if imported.Name != original.Name {
		t.Errorf("Name = %v, want %v", imported.Name, original.Name)
	}
	if imported.Description != original.Description {
		t.Errorf("Description = %v, want %v", imported.Description, original.Description)
	}
	if imported.Author != original.Author {
		t.Errorf("Author = %v, want %v", imported.Author, original.Author)
	}
	if imported.GenreID != original.GenreID {
		t.Errorf("GenreID = %v, want %v", imported.GenreID, original.GenreID)
	}
	if imported.GetRating() != original.GetRating() {
		t.Errorf("Rating = %v, want %v", imported.GetRating(), original.GetRating())
	}
	if imported.GetRatingCount() != original.GetRatingCount() {
		t.Errorf("RatingCount = %v, want %v", imported.GetRatingCount(), original.GetRatingCount())
	}
	if imported.GetDownloads() != original.GetDownloads() {
		t.Errorf("Downloads = %v, want %v", imported.GetDownloads(), original.GetDownloads())
	}
	if len(imported.Tags) != len(original.Tags) {
		t.Fatalf("Tags length = %v, want %v", len(imported.Tags), len(original.Tags))
	}
	for i, tag := range imported.Tags {
		if tag != original.Tags[i] {
			t.Errorf("Tag[%d] = %v, want %v", i, tag, original.Tags[i])
		}
	}
	if imported.BuildingDef.Type != original.BuildingDef.Type {
		t.Errorf("BuildingDef.Type = %v, want %v", imported.BuildingDef.Type, original.BuildingDef.Type)
	}
	if imported.BuildingDef.Width != original.BuildingDef.Width {
		t.Errorf("BuildingDef.Width = %v, want %v", imported.BuildingDef.Width, original.BuildingDef.Width)
	}
	if imported.BuildingDef.Seed != original.BuildingDef.Seed {
		t.Errorf("BuildingDef.Seed = %v, want %v", imported.BuildingDef.Seed, original.BuildingDef.Seed)
	}
}

func TestImportBlueprintInvalidFile(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{"nonexistent file", "/nonexistent/path/blueprint.json.gz"},
		{"invalid gzip", "/tmp/invalid.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ImportBlueprint(tt.filepath)
			if err == nil {
				t.Error("Expected error for invalid file, got nil")
			}
		})
	}
}

func TestNewBlueprintLibrary(t *testing.T) {
	library := NewBlueprintLibrary()
	if library == nil {
		t.Fatal("NewBlueprintLibrary returned nil")
	}
	if library.Count() != 0 {
		t.Errorf("New library count = %v, want 0", library.Count())
	}
}

func TestBlueprintLibraryAdd(t *testing.T) {
	library := NewBlueprintLibrary()
	bp1 := NewBlueprintWithTime("House 1", "player1", "fantasy", nil, defaultTimeProvider)
	bp2 := NewBlueprintWithTime("House 2", "player2", "scifi", nil, defaultTimeProvider)

	library.Add(bp1)
	if library.Count() != 1 {
		t.Errorf("Count after 1 add = %v, want 1", library.Count())
	}

	library.Add(bp2)
	if library.Count() != 2 {
		t.Errorf("Count after 2 adds = %v, want 2", library.Count())
	}

	// Replace existing blueprint without copying the struct (avoids copying embedded mutex)
	bp1.Name = "Updated House"
	library.Add(bp1)
	if library.Count() != 2 {
		t.Errorf("Count after replacement = %v, want 2", library.Count())
	}

	retrieved := library.Get(bp1.ID)
	if retrieved.Name != "Updated House" {
		t.Errorf("Retrieved name = %v, want Updated House", retrieved.Name)
	}
}

func TestBlueprintLibraryGet(t *testing.T) {
	library := NewBlueprintLibrary()
	bp := NewBlueprintWithTime("Test", "player1", "fantasy", nil, defaultTimeProvider)
	library.Add(bp)

	retrieved := library.Get(bp.ID)
	if retrieved == nil {
		t.Fatal("Get returned nil for existing blueprint")
	}
	if retrieved.ID != bp.ID {
		t.Errorf("Retrieved ID = %v, want %v", retrieved.ID, bp.ID)
	}

	notFound := library.Get("nonexistent_id")
	if notFound != nil {
		t.Error("Get should return nil for nonexistent ID")
	}
}

func TestBlueprintLibraryRemove(t *testing.T) {
	library := NewBlueprintLibrary()
	bp := NewBlueprintWithTime("Test", "player1", "fantasy", nil, defaultTimeProvider)
	library.Add(bp)

	// Remove existing
	removed := library.Remove(bp.ID)
	if !removed {
		t.Error("Remove should return true for existing blueprint")
	}
	if library.Count() != 0 {
		t.Errorf("Count after remove = %v, want 0", library.Count())
	}

	// Remove nonexistent
	removed = library.Remove("nonexistent_id")
	if removed {
		t.Error("Remove should return false for nonexistent blueprint")
	}
}

func TestBlueprintLibraryList(t *testing.T) {
	library := NewBlueprintLibrary()

	// Empty library
	list := library.List()
	if len(list) != 0 {
		t.Errorf("Empty library list length = %v, want 0", len(list))
	}

	// Add blueprints
	bp1 := NewBlueprintWithTime("House 1", "player1", "fantasy", nil, defaultTimeProvider)
	bp2 := NewBlueprintWithTime("House 2", "player2", "scifi", nil, defaultTimeProvider)
	bp3 := NewBlueprintWithTime("House 3", "player3", "horror", nil, defaultTimeProvider)

	library.Add(bp1)
	library.Add(bp2)
	library.Add(bp3)

	list = library.List()
	if len(list) != 3 {
		t.Errorf("List length = %v, want 3", len(list))
	}

	// Verify all blueprints are in the list
	ids := make(map[string]bool)
	for _, bp := range list {
		ids[bp.ID] = true
	}
	if !ids[bp1.ID] || !ids[bp2.ID] || !ids[bp3.ID] {
		t.Error("List should contain all added blueprints")
	}
}

func TestBlueprintLibraryFilter(t *testing.T) {
	library := NewBlueprintLibrary()

	// Create test blueprints
	bp1 := NewBlueprintWithTime("Medieval Manor", "player1", "fantasy", &BuildingDefinition{Width: 24, Height: 24}, defaultTimeProvider)
	bp1.Tags = []string{"medieval", "manor"}
	bp1.AddRating(4.5)
	library.Add(bp1)

	bp2 := NewBlueprintWithTime("Sci-Fi Station", "player2", "scifi", &BuildingDefinition{Width: 32, Height: 32}, defaultTimeProvider)
	bp2.Tags = []string{"scifi", "station"}
	bp2.AddRating(3.5)
	library.Add(bp2)

	bp3 := NewBlueprintWithTime("Medieval Castle", "player1", "fantasy", &BuildingDefinition{Width: 48, Height: 48}, defaultTimeProvider)
	bp3.Tags = []string{"medieval", "castle"}
	bp3.AddRating(4.8)
	library.Add(bp3)

	tests := []struct {
		name      string
		opts      FilterOptions
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "no filters",
			opts:      FilterOptions{},
			wantCount: 3,
		},
		{
			name:      "filter by author",
			opts:      FilterOptions{Author: "player1"},
			wantCount: 2,
			wantIDs:   []string{bp1.ID, bp3.ID},
		},
		{
			name:      "filter by genre",
			opts:      FilterOptions{GenreID: "fantasy"},
			wantCount: 2,
		},
		{
			name:      "filter by tag",
			opts:      FilterOptions{Tags: []string{"medieval"}},
			wantCount: 2,
		},
		{
			name:      "filter by multiple tags",
			opts:      FilterOptions{Tags: []string{"medieval", "manor"}},
			wantCount: 1,
			wantIDs:   []string{bp1.ID},
		},
		{
			name:      "filter by min rating",
			opts:      FilterOptions{MinRating: 4.0},
			wantCount: 2,
		},
		{
			name:      "filter by max rating",
			opts:      FilterOptions{MaxRating: 4.0},
			wantCount: 1,
		},
		{
			name:      "filter by size range",
			opts:      FilterOptions{MinSize: 30, MaxSize: 50},
			wantCount: 2,
		},
		{
			name:      "combined filters",
			opts:      FilterOptions{Author: "player1", MinRating: 4.6},
			wantCount: 1,
			wantIDs:   []string{bp3.ID},
		},
		{
			name:      "no matches",
			opts:      FilterOptions{Author: "nonexistent"},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := library.Filter(tt.opts)
			if len(results) != tt.wantCount {
				t.Errorf("Filter result count = %v, want %v", len(results), tt.wantCount)
			}

			if tt.wantIDs != nil {
				resultIDs := make(map[string]bool)
				for _, bp := range results {
					resultIDs[bp.ID] = true
				}
				for _, wantID := range tt.wantIDs {
					if !resultIDs[wantID] {
						t.Errorf("Expected ID %v in results", wantID)
					}
				}
			}
		})
	}
}

func TestBlueprintLibrarySort(t *testing.T) {
	library := NewBlueprintLibrary()

	// Create blueprints with different values
	now := time.Now()
	bp1 := NewBlueprintWithTime("Alpha", "player1", "fantasy", nil, defaultTimeProvider)
	bp1.AddRating(3.0)
	for i := 0; i < 100; i++ {
		bp1.IncrementDownloads()
	}
	bp1.CreatedAt = now.Add(-3 * time.Hour)
	bp1.ModifiedAt = now.Add(-1 * time.Hour)

	bp2 := NewBlueprintWithTime("Beta", "player2", "scifi", nil, defaultTimeProvider)
	bp2.AddRating(5.0)
	for i := 0; i < 50; i++ {
		bp2.IncrementDownloads()
	}
	bp2.CreatedAt = now.Add(-2 * time.Hour)
	bp2.ModifiedAt = now.Add(-2 * time.Hour)

	bp3 := NewBlueprintWithTime("Gamma", "player3", "horror", nil, defaultTimeProvider)
	bp3.AddRating(4.0)
	for i := 0; i < 150; i++ {
		bp3.IncrementDownloads()
	}
	bp3.CreatedAt = now.Add(-1 * time.Hour)
	bp3.ModifiedAt = now.Add(-3 * time.Hour)

	blueprints := []*Blueprint{bp1, bp2, bp3}

	tests := []struct {
		name       string
		field      SortField
		descending bool
		wantOrder  []string
	}{
		{
			name:       "by rating ascending",
			field:      SortByRating,
			descending: false,
			wantOrder:  []string{"Alpha", "Gamma", "Beta"},
		},
		{
			name:       "by rating descending",
			field:      SortByRating,
			descending: true,
			wantOrder:  []string{"Beta", "Gamma", "Alpha"},
		},
		{
			name:       "by downloads ascending",
			field:      SortByDownloads,
			descending: false,
			wantOrder:  []string{"Beta", "Alpha", "Gamma"},
		},
		{
			name:       "by downloads descending",
			field:      SortByDownloads,
			descending: true,
			wantOrder:  []string{"Gamma", "Alpha", "Beta"},
		},
		{
			name:       "by name ascending",
			field:      SortByName,
			descending: false,
			wantOrder:  []string{"Alpha", "Beta", "Gamma"},
		},
		{
			name:       "by name descending",
			field:      SortByName,
			descending: true,
			wantOrder:  []string{"Gamma", "Beta", "Alpha"},
		},
		{
			name:       "by created ascending",
			field:      SortByCreated,
			descending: false,
			wantOrder:  []string{"Alpha", "Beta", "Gamma"},
		},
		{
			name:       "by modified descending",
			field:      SortByModified,
			descending: true,
			wantOrder:  []string{"Alpha", "Beta", "Gamma"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying original
			testBlueprints := make([]*Blueprint, len(blueprints))
			copy(testBlueprints, blueprints)

			library.Sort(testBlueprints, tt.field, tt.descending)

			for i, expectedName := range tt.wantOrder {
				if testBlueprints[i].Name != expectedName {
					t.Errorf("Position %d: got %v, want %v", i, testBlueprints[i].Name, expectedName)
				}
			}
		})
	}
}

func TestBlueprintLibraryConcurrency(t *testing.T) {
	library := NewBlueprintLibrary()

	// Test concurrent adds
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			bp := NewBlueprintWithTime("Concurrent", "player1", "fantasy", nil, defaultTimeProvider)
			library.Add(bp)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if library.Count() != 10 {
		t.Errorf("Concurrent adds: count = %v, want 10", library.Count())
	}

	// Test concurrent reads
	bp := NewBlueprintWithTime("Read Test", "player1", "fantasy", nil, defaultTimeProvider)
	library.Add(bp)

	for i := 0; i < 10; i++ {
		go func() {
			retrieved := library.Get(bp.ID)
			if retrieved == nil {
				t.Error("Concurrent read failed")
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestContainsAllTags(t *testing.T) {
	tests := []struct {
		name          string
		blueprintTags []string
		requiredTags  []string
		want          bool
	}{
		{
			name:          "all tags present",
			blueprintTags: []string{"medieval", "manor", "large"},
			requiredTags:  []string{"medieval", "manor"},
			want:          true,
		},
		{
			name:          "missing tag",
			blueprintTags: []string{"medieval", "manor"},
			requiredTags:  []string{"medieval", "castle"},
			want:          false,
		},
		{
			name:          "case insensitive",
			blueprintTags: []string{"Medieval", "Manor"},
			requiredTags:  []string{"medieval", "manor"},
			want:          true,
		},
		{
			name:          "empty required",
			blueprintTags: []string{"medieval"},
			requiredTags:  []string{},
			want:          true,
		},
		{
			name:          "empty blueprint tags",
			blueprintTags: []string{},
			requiredTags:  []string{"medieval"},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAllTags(tt.blueprintTags, tt.requiredTags)
			if got != tt.want {
				t.Errorf("containsAllTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
