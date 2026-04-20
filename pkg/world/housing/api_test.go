package housing

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/building"
)

// TestAPIMatchesDocumentation verifies that the implementation matches the documented API in doc.go
func TestAPIMatchesDocumentation(t *testing.T) {
	// This test follows the exact example from doc.go to ensure the API matches

	// Create a blueprint library
	library := NewBlueprintLibrary()
	if library == nil {
		t.Fatal("NewBlueprintLibrary() returned nil")
	}

	// Create a blueprint matching doc.go example using the constructor
	bp := NewBlueprintWithTime("Fantasy Manor", "Player1", "fantasy", &BuildingDefinition{
		Type:   int(building.TypeManor),
		Style:  int(building.StyleMedieval),
		Width:  24,
		Height: 24,
		Floors: 3,
		Seed:   12345,
	}, defaultTimeProvider)
	bp.ID = "bp001" // Set ID to match doc example
	bp.Tags = []string{"medieval", "manor"}

	// Add to library
	library.Add(bp)
	if library.Count() != 1 {
		t.Errorf("Library count = %v, want 1", library.Count())
	}

	// Export blueprint (doc.go shows: bp.Export("blueprints/fantasy_manor.json.gz"))
	tmpFile := t.TempDir() + "/fantasy_manor.json.gz"
	err := bp.Export(tmpFile)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import blueprint (doc.go shows: housing.ImportBlueprint(...))
	imported, err := ImportBlueprint(tmpFile)
	if err != nil {
		t.Fatalf("ImportBlueprint failed: %v", err)
	}
	if imported.ID != bp.ID {
		t.Errorf("Imported ID = %v, want %v", imported.ID, bp.ID)
	}

	// Search blueprints (doc.go shows: library.Filter(housing.FilterOptions{...}))
	results := library.Filter(FilterOptions{
		Author:    "Player1",
		MinRating: 4.0,
		Tags:      []string{"medieval"},
	})
	// Note: The blueprint has 0.0 rating initially, so filter with MinRating: 4.0 returns empty
	// This is correct behavior - the test in doc.go assumes a rated blueprint

	// Test with rating added
	bp.AddRating(4.5)
	library.Add(bp) // Re-add with updated rating
	results = library.Filter(FilterOptions{
		Author:    "Player1",
		MinRating: 4.0,
		Tags:      []string{"medieval"},
	})
	if len(results) != 1 {
		t.Errorf("Filter with valid criteria returned %v results, want 1", len(results))
	}

	// Sort by rating (doc.go shows: library.Sort(results, housing.SortByRating, true))
	library.Sort(results, SortByRating, true)
	// After sort, first item should be the highest rated
	if len(results) > 0 && results[0].GetRating() < 4.0 {
		t.Errorf("Sorted result rating = %v, want >= 4.0", results[0].GetRating())
	}

	t.Log("API matches documentation successfully")
}

// TestDocumentedFilterOptions verifies that all filter options from doc.go work correctly
func TestDocumentedFilterOptions(t *testing.T) {
	library := NewBlueprintLibrary()

	// Add test blueprints
	bp1 := NewBlueprintWithTime("Manor 1", "Player1", "fantasy", &BuildingDefinition{Width: 24, Height: 24}, defaultTimeProvider)
	bp1.Tags = []string{"medieval", "manor"}
	bp1.AddRating(4.5)
	library.Add(bp1)

	bp2 := NewBlueprintWithTime("Manor 2", "Player2", "fantasy", &BuildingDefinition{Width: 32, Height: 32}, defaultTimeProvider)
	bp2.Tags = []string{"medieval", "manor", "large"}
	bp2.AddRating(3.5)
	library.Add(bp2)

	bp3 := NewBlueprintWithTime("Tower 1", "Player1", "fantasy", &BuildingDefinition{Width: 16, Height: 16}, defaultTimeProvider)
	bp3.Tags = []string{"medieval", "tower"}
	bp3.AddRating(4.8)
	library.Add(bp3)

	tests := []struct {
		name      string
		filter    FilterOptions
		wantCount int
	}{
		{
			name:      "filter by author",
			filter:    FilterOptions{Author: "Player1"},
			wantCount: 2,
		},
		{
			name:      "filter by min rating",
			filter:    FilterOptions{MinRating: 4.0},
			wantCount: 2,
		},
		{
			name:      "filter by tags",
			filter:    FilterOptions{Tags: []string{"medieval"}},
			wantCount: 3,
		},
		{
			name: "combined filters",
			filter: FilterOptions{
				Author:    "Player1",
				MinRating: 4.0,
				Tags:      []string{"medieval"},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := library.Filter(tt.filter)
			if len(results) != tt.wantCount {
				t.Errorf("Filter returned %v results, want %v", len(results), tt.wantCount)
			}
		})
	}
}

// TestDocumentedSortFields verifies that all sort fields from doc.go work correctly
func TestDocumentedSortFields(t *testing.T) {
	library := NewBlueprintLibrary()

	// Create blueprints with different values for each sort field
	bp1 := NewBlueprintWithTime("Alpha", "Player1", "fantasy", nil, defaultTimeProvider)
	bp1.AddRating(3.0)
	for i := 0; i < 100; i++ {
		bp1.IncrementDownloads()
	}
	library.Add(bp1)

	bp2 := NewBlueprintWithTime("Beta", "Player2", "scifi", nil, defaultTimeProvider)
	bp2.AddRating(5.0)
	for i := 0; i < 50; i++ {
		bp2.IncrementDownloads()
	}
	library.Add(bp2)

	bp3 := NewBlueprintWithTime("Gamma", "Player3", "horror", nil, defaultTimeProvider)
	bp3.AddRating(4.0)
	for i := 0; i < 150; i++ {
		bp3.IncrementDownloads()
	}
	library.Add(bp3)

	allBlueprints := library.List()

	tests := []struct {
		name       string
		field      SortField
		descending bool
		wantFirst  string
	}{
		{
			name:       "sort by rating descending",
			field:      SortByRating,
			descending: true,
			wantFirst:  "Beta", // Highest rating (5.0)
		},
		{
			name:       "sort by rating ascending",
			field:      SortByRating,
			descending: false,
			wantFirst:  "Alpha", // Lowest rating (3.0)
		},
		{
			name:       "sort by downloads descending",
			field:      SortByDownloads,
			descending: true,
			wantFirst:  "Gamma", // Most downloads (150)
		},
		{
			name:       "sort by name ascending",
			field:      SortByName,
			descending: false,
			wantFirst:  "Alpha", // Alphabetically first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the original slice
			testBlueprints := make([]*Blueprint, len(allBlueprints))
			copy(testBlueprints, allBlueprints)

			library.Sort(testBlueprints, tt.field, tt.descending)

			if testBlueprints[0].Name != tt.wantFirst {
				t.Errorf("After sort, first blueprint = %v, want %v", testBlueprints[0].Name, tt.wantFirst)
			}
		})
	}
}
