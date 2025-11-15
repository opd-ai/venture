package ui

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/story"
)

// TestNewStoryJournalUI tests UI creation.
func TestNewStoryJournalUI(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")

	if ui.X != 0 || ui.Y != 0 {
		t.Errorf("Expected position (0,0), got (%d,%d)", ui.X, ui.Y)
	}

	if ui.Width != 800 || ui.Height != 600 {
		t.Errorf("Expected size (800,600), got (%d,%d)", ui.Width, ui.Height)
	}

	if ui.GenreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", ui.GenreID)
	}

	if ui.ViewMode != ViewSeriesList {
		t.Errorf("Expected view mode ViewSeriesList, got %v", ui.ViewMode)
	}

	if ui.BackgroundColor == nil {
		t.Error("BackgroundColor should be initialized")
	}
}

// TestNewStoryJournalUI_DifferentGenres tests genre-specific styling.
func TestNewStoryJournalUI_DifferentGenres(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"Fantasy", "fantasy"},
		{"Sci-Fi", "scifi"},
		{"Horror", "horror"},
		{"Cyberpunk", "cyberpunk"},
		{"Post-Apocalyptic", "postapocalyptic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := NewStoryJournalUI(0, 0, 800, 600, tt.genreID)

			if ui.GenreID != tt.genreID {
				t.Errorf("Expected genre '%s', got '%s'", tt.genreID, ui.GenreID)
			}

			// Verify colors are set (not nil)
			if ui.BackgroundColor == nil {
				t.Error("BackgroundColor not set")
			}
			if ui.TextColor == nil {
				t.Error("TextColor not set")
			}
			if ui.HighlightColor == nil {
				t.Error("HighlightColor not set")
			}
		})
	}
}

// TestStoryJournalUI_LoadFromJournal tests loading journal data.
func TestStoryJournalUI_LoadFromJournal(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	world := engine.NewWorld()
	journal := engine.NewStoryJournalComponent()

	// Create some fragments
	frag1 := world.CreateEntity()
	frag1.AddComponent(&engine.StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:    story.FragmentNote,
			Content: "Fragment 1",
		},
		SeriesID:    "ancient_curse",
		SequenceNum: 1,
	})

	frag2 := world.CreateEntity()
	frag2.AddComponent(&engine.StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:    story.FragmentCarving,
			Content: "Fragment 2",
		},
		SeriesID:    "ancient_curse",
		SequenceNum: 2,
	})

	frag3 := world.CreateEntity()
	frag3.AddComponent(&engine.StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:    story.FragmentRelic,
			Content: "Fragment 3",
		},
		SeriesID:    "fallen_kingdom",
		SequenceNum: 1,
	})

	world.Update(0.0) // Commit entities

	// Mark some fragments as discovered
	journal.AddDiscovery("ancient_curse", 1)

	// Load into UI
	ui.LoadFromJournal(journal, world)

	if len(ui.SeriesList) != 2 {
		t.Errorf("Expected 2 series, got %d", len(ui.SeriesList))
	}

	// Check series data
	var ancientCurse *SeriesEntry
	for i := range ui.SeriesList {
		if ui.SeriesList[i].SeriesID == "ancient_curse" {
			ancientCurse = &ui.SeriesList[i]
			break
		}
	}

	if ancientCurse == nil {
		t.Fatal("ancient_curse series not found")
	}

	if ancientCurse.FragmentCount != 2 {
		t.Errorf("Expected 2 fragments in ancient_curse, got %d", ancientCurse.FragmentCount)
	}

	if ancientCurse.DiscoveredCount != 1 {
		t.Errorf("Expected 1 discovered in ancient_curse, got %d", ancientCurse.DiscoveredCount)
	}
}

// TestStoryJournalUI_LoadFromJournal_EmptyJournal tests empty state.
func TestStoryJournalUI_LoadFromJournal_EmptyJournal(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	world := engine.NewWorld()
	journal := engine.NewStoryJournalComponent()

	ui.LoadFromJournal(journal, world)

	if len(ui.SeriesList) != 0 {
		t.Errorf("Expected 0 series with empty journal, got %d", len(ui.SeriesList))
	}
}

// TestStoryJournalUI_LoadFragmentsForSeries tests fragment loading.
func TestStoryJournalUI_LoadFragmentsForSeries(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	world := engine.NewWorld()
	journal := engine.NewStoryJournalComponent()

	// Create fragments for a series
	for i := 1; i <= 3; i++ {
		frag := world.CreateEntity()
		frag.AddComponent(&engine.StoryFragmentComponent{
			Fragment: story.StoryFragment{
				Type:    story.FragmentNote,
				Content: "Content",
			},
			SeriesID:    "test_series",
			SequenceNum: i,
		})
	}

	world.Update(0.0)

	// Mark one as discovered
	journal.AddDiscovery("test_series", 2)

	// Load series list
	ui.LoadFromJournal(journal, world)

	// Load fragments for first series
	ui.SelectedSeriesIndex = 0
	ui.LoadFragmentsForSeries(journal, world)

	if len(ui.VisibleFragments) != 3 {
		t.Errorf("Expected 3 fragments, got %d", len(ui.VisibleFragments))
	}

	// Check discovered status
	discoveredCount := 0
	for _, frag := range ui.VisibleFragments {
		if frag.IsDiscovered {
			discoveredCount++
		}
	}

	if discoveredCount != 1 {
		t.Errorf("Expected 1 discovered fragment, got %d", discoveredCount)
	}

	// Check sorting by sequence
	for i, frag := range ui.VisibleFragments {
		if frag.SequenceNum != i+1 {
			t.Errorf("Fragment %d has sequence %d, expected %d", i, frag.SequenceNum, i+1)
		}
	}
}

// TestStoryJournalUI_NavigateUp tests upward navigation.
func TestStoryJournalUI_NavigateUp(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	ui.SeriesList = []SeriesEntry{
		{SeriesID: "series1"},
		{SeriesID: "series2"},
		{SeriesID: "series3"},
	}
	ui.SelectedSeriesIndex = 2

	ui.NavigateUp()

	if ui.SelectedSeriesIndex != 1 {
		t.Errorf("Expected index 1, got %d", ui.SelectedSeriesIndex)
	}

	ui.NavigateUp()

	if ui.SelectedSeriesIndex != 0 {
		t.Errorf("Expected index 0, got %d", ui.SelectedSeriesIndex)
	}

	// Test boundary (should stay at 0)
	ui.NavigateUp()

	if ui.SelectedSeriesIndex != 0 {
		t.Errorf("Expected index to stay at 0, got %d", ui.SelectedSeriesIndex)
	}
}

// TestStoryJournalUI_NavigateDown tests downward navigation.
func TestStoryJournalUI_NavigateDown(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	ui.SeriesList = []SeriesEntry{
		{SeriesID: "series1"},
		{SeriesID: "series2"},
		{SeriesID: "series3"},
	}
	ui.SelectedSeriesIndex = 0

	ui.NavigateDown()

	if ui.SelectedSeriesIndex != 1 {
		t.Errorf("Expected index 1, got %d", ui.SelectedSeriesIndex)
	}

	ui.NavigateDown()

	if ui.SelectedSeriesIndex != 2 {
		t.Errorf("Expected index 2, got %d", ui.SelectedSeriesIndex)
	}

	// Test boundary (should stay at 2)
	ui.NavigateDown()

	if ui.SelectedSeriesIndex != 2 {
		t.Errorf("Expected index to stay at 2, got %d", ui.SelectedSeriesIndex)
	}
}

// TestStoryJournalUI_NavigateSelect tests selection navigation.
func TestStoryJournalUI_NavigateSelect(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	world := engine.NewWorld()
	journal := engine.NewStoryJournalComponent()

	// Create a series with fragments
	frag := world.CreateEntity()
	frag.AddComponent(&engine.StoryFragmentComponent{
		Fragment:    story.StoryFragment{Type: story.FragmentNote, Content: "Test"},
		SeriesID:    "test",
		SequenceNum: 1,
	})
	world.Update(0.0)

	journal.AddDiscovery("test", 1)

	ui.LoadFromJournal(journal, world)

	// Initially in series list view
	if ui.ViewMode != ViewSeriesList {
		t.Fatalf("Expected ViewSeriesList, got %v", ui.ViewMode)
	}

	// Select series (should move to fragment list)
	ui.NavigateSelect(journal, world)

	if ui.ViewMode != ViewFragmentList {
		t.Errorf("Expected ViewFragmentList, got %v", ui.ViewMode)
	}

	// Select fragment (should move to detail view)
	ui.NavigateSelect(journal, world)

	if ui.ViewMode != ViewFragmentDetail {
		t.Errorf("Expected ViewFragmentDetail, got %v", ui.ViewMode)
	}
}

// TestStoryJournalUI_NavigateBack tests back navigation.
func TestStoryJournalUI_NavigateBack(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")

	// Start at fragment detail
	ui.ViewMode = ViewFragmentDetail

	ui.NavigateBack()

	if ui.ViewMode != ViewFragmentList {
		t.Errorf("Expected ViewFragmentList, got %v", ui.ViewMode)
	}

	ui.NavigateBack()

	if ui.ViewMode != ViewSeriesList {
		t.Errorf("Expected ViewSeriesList, got %v", ui.ViewMode)
	}

	// Back from series list should do nothing
	ui.NavigateBack()

	if ui.ViewMode != ViewSeriesList {
		t.Errorf("Expected to stay at ViewSeriesList, got %v", ui.ViewMode)
	}
}

// TestStoryJournalUI_Render tests rendering.
func TestStoryJournalUI_Render(t *testing.T) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")

	img := ui.Render()

	if img == nil {
		t.Fatal("Render() returned nil")
	}

	if img.Bounds().Dx() != 800 || img.Bounds().Dy() != 600 {
		t.Errorf("Expected image size (800,600), got (%d,%d)",
			img.Bounds().Dx(), img.Bounds().Dy())
	}
}

// TestFormatSeriesName tests series name formatting.
func TestFormatSeriesName(t *testing.T) {
	tests := []struct {
		name     string
		seriesID string
		want     string
	}{
		{"Single word", "ancient", "Ancient"},
		{"Two words", "ancient_curse", "Ancient Curse"},
		{"Three words", "fallen_dark_kingdom", "Fallen Dark Kingdom"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSeriesName(tt.seriesID)

			if result != tt.want {
				t.Errorf("formatSeriesName(%s) = %s, want %s", tt.seriesID, result, tt.want)
			}
		})
	}
}

// TestTruncateText tests text truncation.
func TestTruncateText(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		maxLen int
		want   string
	}{
		{"Short text", "Hello", 10, "Hello"},
		{"Exact length", "Hello", 5, "Hello"},
		{"Truncate", "Hello World", 8, "Hello..."},
		{"Long truncate", "This is a very long text", 15, "This is a ve..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)

			if result != tt.want {
				t.Errorf("truncateText(%s, %d) = %s, want %s", tt.text, tt.maxLen, result, tt.want)
			}
		})
	}
}

// TestWrapText tests text wrapping.
func TestWrapText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  int
		wantLines int
	}{
		{"Short text", "Hello", 20, 1},
		{"Exact width", "Hello World", 11, 1},
		{"Two lines", "Hello World Test", 11, 2},
		{"Multiple lines", "The quick brown fox jumps over the lazy dog", 15, 3},
		{"Empty", "", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapText(tt.text, tt.maxWidth)

			if len(lines) != tt.wantLines {
				t.Errorf("wrapText() returned %d lines, want %d", len(lines), tt.wantLines)
			}

			// Verify no line exceeds max width
			for i, line := range lines {
				if len(line) > tt.maxWidth {
					t.Errorf("Line %d exceeds max width: %d > %d", i, len(line), tt.maxWidth)
				}
			}
		})
	}
}

// Benchmark UI operations
func BenchmarkStoryJournalUI_LoadFromJournal(b *testing.B) {
	world := engine.NewWorld()
	journal := engine.NewStoryJournalComponent()

	// Create test data
	for i := 0; i < 20; i++ {
		frag := world.CreateEntity()
		frag.AddComponent(&engine.StoryFragmentComponent{
			Fragment:    story.StoryFragment{Type: story.FragmentNote, Content: "Test"},
			SeriesID:    "series",
			SequenceNum: i,
		})
	}
	world.Update(0.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
		ui.LoadFromJournal(journal, world)
	}
}

func BenchmarkStoryJournalUI_Render(b *testing.B) {
	ui := NewStoryJournalUI(0, 0, 800, 600, "fantasy")
	ui.SeriesList = []SeriesEntry{
		{SeriesID: "s1", SeriesName: "Series 1", FragmentCount: 5},
		{SeriesID: "s2", SeriesName: "Series 2", FragmentCount: 3},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.Render()
	}
}
