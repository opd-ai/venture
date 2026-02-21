package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/story"
)

func TestStoryFragmentComponent_Type(t *testing.T) {
	comp := StoryFragmentComponent{}
	if got := comp.Type(); got != "storyfragment" {
		t.Errorf("Type() = %v, want %v", got, "storyfragment")
	}
}

func TestStoryJournalComponent_Type(t *testing.T) {
	comp := StoryJournalComponent{}
	if got := comp.Type(); got != "storyjournal" {
		t.Errorf("Type() = %v, want %v", got, "storyjournal")
	}
}

func TestNewStoryJournalComponent(t *testing.T) {
	journal := NewStoryJournalComponent()

	if journal == nil {
		t.Fatal("NewStoryJournalComponent() returned nil")
	}

	if journal.DiscoveredFragments == nil {
		t.Error("DiscoveredFragments map not initialized")
	}

	if journal.CompletedSeries == nil {
		t.Error("CompletedSeries map not initialized")
	}

	if journal.TotalDiscoveries != 0 {
		t.Errorf("TotalDiscoveries = %d, want 0", journal.TotalDiscoveries)
	}

	if journal.TotalSeriesComplete != 0 {
		t.Errorf("TotalSeriesComplete = %d, want 0", journal.TotalSeriesComplete)
	}
}

func TestStoryJournalComponent_AddDiscovery(t *testing.T) {
	journal := NewStoryJournalComponent()

	// First discovery should return true
	isNew := journal.AddDiscovery("series1", 0)
	if !isNew {
		t.Error("First AddDiscovery() returned false, want true")
	}

	if journal.TotalDiscoveries != 1 {
		t.Errorf("TotalDiscoveries = %d, want 1", journal.TotalDiscoveries)
	}

	// Duplicate discovery should return false
	isNew = journal.AddDiscovery("series1", 0)
	if isNew {
		t.Error("Duplicate AddDiscovery() returned true, want false")
	}

	if journal.TotalDiscoveries != 1 {
		t.Errorf("TotalDiscoveries = %d after duplicate, want 1", journal.TotalDiscoveries)
	}

	// Different fragment should return true
	isNew = journal.AddDiscovery("series1", 1)
	if !isNew {
		t.Error("Different fragment AddDiscovery() returned false, want true")
	}

	if journal.TotalDiscoveries != 2 {
		t.Errorf("TotalDiscoveries = %d, want 2", journal.TotalDiscoveries)
	}
}

func TestStoryJournalComponent_IsDiscovered(t *testing.T) {
	journal := NewStoryJournalComponent()

	// Should not be discovered initially
	if journal.IsDiscovered("series1", 0) {
		t.Error("IsDiscovered() returned true before discovery")
	}

	// Add discovery
	journal.AddDiscovery("series1", 0)

	// Should be discovered now
	if !journal.IsDiscovered("series1", 0) {
		t.Error("IsDiscovered() returned false after discovery")
	}

	// Different fragment should not be discovered
	if journal.IsDiscovered("series1", 1) {
		t.Error("IsDiscovered() returned true for undiscovered fragment")
	}
}

func TestStoryJournalComponent_IsSeriesComplete(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"
	totalFragments := 3

	// Empty series should not be complete
	if journal.IsSeriesComplete(seriesID, totalFragments) {
		t.Error("IsSeriesComplete() returned true for empty series")
	}

	// Add one fragment
	journal.AddDiscovery(seriesID, 0)
	if journal.IsSeriesComplete(seriesID, totalFragments) {
		t.Error("IsSeriesComplete() returned true with 1/3 fragments")
	}

	// Add second fragment
	journal.AddDiscovery(seriesID, 1)
	if journal.IsSeriesComplete(seriesID, totalFragments) {
		t.Error("IsSeriesComplete() returned true with 2/3 fragments")
	}

	// Add third fragment
	journal.AddDiscovery(seriesID, 2)
	if !journal.IsSeriesComplete(seriesID, totalFragments) {
		t.Error("IsSeriesComplete() returned false with 3/3 fragments")
	}
}

func TestStoryJournalComponent_MarkSeriesComplete(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"

	// Mark series complete
	journal.MarkSeriesComplete(seriesID)

	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d, want 1", journal.TotalSeriesComplete)
	}

	if !journal.CompletedSeries[seriesID] {
		t.Error("CompletedSeries does not contain marked series")
	}

	// Mark same series again (should not increment)
	journal.MarkSeriesComplete(seriesID)
	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d after duplicate mark, want 1", journal.TotalSeriesComplete)
	}

	// Mark different series
	journal.MarkSeriesComplete("series2")
	if journal.TotalSeriesComplete != 2 {
		t.Errorf("TotalSeriesComplete = %d, want 2", journal.TotalSeriesComplete)
	}
}

func TestStoryJournalComponent_GetDiscoveryCount(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"
	totalFragments := 5

	// Empty series
	count := journal.GetDiscoveryCount(seriesID, totalFragments)
	if count != 0 {
		t.Errorf("GetDiscoveryCount() = %d for empty series, want 0", count)
	}

	// Add 3 fragments
	journal.AddDiscovery(seriesID, 0)
	journal.AddDiscovery(seriesID, 2)
	journal.AddDiscovery(seriesID, 4)

	count = journal.GetDiscoveryCount(seriesID, totalFragments)
	if count != 3 {
		t.Errorf("GetDiscoveryCount() = %d, want 3", count)
	}
}

func TestStoryJournalComponent_LastDiscoveryTime(t *testing.T) {
	journal := NewStoryJournalComponent()

	// Zero time initially
	if !journal.LastDiscoveryTime.IsZero() {
		t.Error("LastDiscoveryTime should be zero initially")
	}

	// Record before adding discovery
	beforeTime := time.Now()
	time.Sleep(10 * time.Millisecond) // Ensure time passes

	journal.AddDiscovery("series1", 0)

	// Check time was updated
	if journal.LastDiscoveryTime.IsZero() {
		t.Error("LastDiscoveryTime still zero after discovery")
	}

	if journal.LastDiscoveryTime.Before(beforeTime) {
		t.Error("LastDiscoveryTime is before discovery was added")
	}
}

func TestStoryFragmentComponent_WithFragmentData(t *testing.T) {
	fragment := story.StoryFragment{
		Type:        story.FragmentNote,
		Content:     "A mysterious note was found here.",
		Location:    story.Vector2{X: 10.0, Y: 20.0},
		DiscoveryXP: 50.0,
		SeriesID:    "ancient-curse",
		SequenceNum: 0,
	}

	comp := StoryFragmentComponent{
		Fragment:    fragment,
		Discovered:  false,
		SeriesID:    "ancient-curse",
		SequenceNum: 0,
	}

	if comp.Fragment.Type != story.FragmentNote {
		t.Errorf("Fragment.Type = %v, want %v", comp.Fragment.Type, story.FragmentNote)
	}

	if comp.Fragment.Content != "A mysterious note was found here." {
		t.Errorf("Fragment.Content = %v, want specific message", comp.Fragment.Content)
	}

	if comp.Fragment.DiscoveryXP != 50.0 {
		t.Errorf("Fragment.DiscoveryXP = %v, want 50.0", comp.Fragment.DiscoveryXP)
	}

	if comp.Discovered {
		t.Error("Discovered should be false initially")
	}
}

func TestStoryJournalComponent_MultipleSeries(t *testing.T) {
	journal := NewStoryJournalComponent()

	// Add fragments from multiple series
	journal.AddDiscovery("series1", 0)
	journal.AddDiscovery("series1", 1)
	journal.AddDiscovery("series2", 0)
	journal.AddDiscovery("series3", 0)
	journal.AddDiscovery("series3", 1)
	journal.AddDiscovery("series3", 2)

	if journal.TotalDiscoveries != 6 {
		t.Errorf("TotalDiscoveries = %d, want 6", journal.TotalDiscoveries)
	}

	// Check individual series counts
	count1 := journal.GetDiscoveryCount("series1", 3)
	if count1 != 2 {
		t.Errorf("Series1 discovery count = %d, want 2", count1)
	}

	count2 := journal.GetDiscoveryCount("series2", 3)
	if count2 != 1 {
		t.Errorf("Series2 discovery count = %d, want 1", count2)
	}

	count3 := journal.GetDiscoveryCount("series3", 3)
	if count3 != 3 {
		t.Errorf("Series3 discovery count = %d, want 3", count3)
	}

	// Mark series3 as complete
	journal.MarkSeriesComplete("series3")

	if !journal.CompletedSeries["series3"] {
		t.Error("Series3 not marked as complete")
	}

	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d, want 1", journal.TotalSeriesComplete)
	}
}

// --- Tests for ECS-Compliant Helper Functions ---

func TestJournalAddDiscovery(t *testing.T) {
	journal := NewStoryJournalComponent()
	fixedTime := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	// First discovery should return true
	isNew := JournalAddDiscovery(journal, "series1", 0, fixedTime)
	if !isNew {
		t.Error("JournalAddDiscovery() returned false for new discovery, want true")
	}

	if journal.TotalDiscoveries != 1 {
		t.Errorf("TotalDiscoveries = %d, want 1", journal.TotalDiscoveries)
	}

	if !journal.LastDiscoveryTime.Equal(fixedTime) {
		t.Errorf("LastDiscoveryTime = %v, want %v", journal.LastDiscoveryTime, fixedTime)
	}

	// Duplicate discovery should return false
	isNew = JournalAddDiscovery(journal, "series1", 0, fixedTime.Add(time.Hour))
	if isNew {
		t.Error("JournalAddDiscovery() returned true for duplicate, want false")
	}

	// Time should not be updated for duplicate
	if !journal.LastDiscoveryTime.Equal(fixedTime) {
		t.Errorf("LastDiscoveryTime changed for duplicate discovery")
	}

	// Nil journal should return false
	isNew = JournalAddDiscovery(nil, "series1", 0, fixedTime)
	if isNew {
		t.Error("JournalAddDiscovery(nil) returned true, want false")
	}
}

func TestJournalIsDiscovered(t *testing.T) {
	journal := NewStoryJournalComponent()
	fixedTime := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	// Should not be discovered initially
	if JournalIsDiscovered(journal, "series1", 0) {
		t.Error("JournalIsDiscovered() returned true before discovery")
	}

	// Add discovery
	JournalAddDiscovery(journal, "series1", 0, fixedTime)

	// Should be discovered now
	if !JournalIsDiscovered(journal, "series1", 0) {
		t.Error("JournalIsDiscovered() returned false after discovery")
	}

	// Nil journal should return false
	if JournalIsDiscovered(nil, "series1", 0) {
		t.Error("JournalIsDiscovered(nil) returned true, want false")
	}
}

func TestJournalIsSeriesComplete(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"
	totalFragments := 3
	fixedTime := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	// Empty series should not be complete
	if JournalIsSeriesComplete(journal, seriesID, totalFragments) {
		t.Error("JournalIsSeriesComplete() returned true for empty series")
	}

	// Add all fragments
	JournalAddDiscovery(journal, seriesID, 0, fixedTime)
	JournalAddDiscovery(journal, seriesID, 1, fixedTime)
	JournalAddDiscovery(journal, seriesID, 2, fixedTime)

	if !JournalIsSeriesComplete(journal, seriesID, totalFragments) {
		t.Error("JournalIsSeriesComplete() returned false with all fragments")
	}

	// Nil journal should return false
	if JournalIsSeriesComplete(nil, seriesID, totalFragments) {
		t.Error("JournalIsSeriesComplete(nil) returned true, want false")
	}
}

func TestJournalMarkSeriesComplete(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"

	// Mark series complete
	JournalMarkSeriesComplete(journal, seriesID)

	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d, want 1", journal.TotalSeriesComplete)
	}

	if !journal.CompletedSeries[seriesID] {
		t.Error("CompletedSeries does not contain marked series")
	}

	// Mark same series again (should not increment)
	JournalMarkSeriesComplete(journal, seriesID)
	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d after duplicate mark, want 1", journal.TotalSeriesComplete)
	}

	// Nil journal should not panic
	JournalMarkSeriesComplete(nil, seriesID) // Just verify no panic
}

func TestJournalGetDiscoveryCount(t *testing.T) {
	journal := NewStoryJournalComponent()
	seriesID := "test-series"
	totalFragments := 5
	fixedTime := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	// Empty series
	count := JournalGetDiscoveryCount(journal, seriesID, totalFragments)
	if count != 0 {
		t.Errorf("JournalGetDiscoveryCount() = %d for empty series, want 0", count)
	}

	// Add 3 fragments
	JournalAddDiscovery(journal, seriesID, 0, fixedTime)
	JournalAddDiscovery(journal, seriesID, 2, fixedTime)
	JournalAddDiscovery(journal, seriesID, 4, fixedTime)

	count = JournalGetDiscoveryCount(journal, seriesID, totalFragments)
	if count != 3 {
		t.Errorf("JournalGetDiscoveryCount() = %d, want 3", count)
	}

	// Nil journal should return 0
	count = JournalGetDiscoveryCount(nil, seriesID, totalFragments)
	if count != 0 {
		t.Errorf("JournalGetDiscoveryCount(nil) = %d, want 0", count)
	}
}

func TestJournalHelpersDeterministic(t *testing.T) {
	// Verify that helper functions produce deterministic results with fixed time
	fixedTime := time.Date(2026, 2, 21, 10, 0, 0, 0, time.UTC)

	journal1 := NewStoryJournalComponent()
	journal2 := NewStoryJournalComponent()

	// Apply same operations
	JournalAddDiscovery(journal1, "series1", 0, fixedTime)
	JournalAddDiscovery(journal1, "series1", 1, fixedTime)
	JournalMarkSeriesComplete(journal1, "series1")

	JournalAddDiscovery(journal2, "series1", 0, fixedTime)
	JournalAddDiscovery(journal2, "series1", 1, fixedTime)
	JournalMarkSeriesComplete(journal2, "series1")

	// Both journals should be identical
	if journal1.TotalDiscoveries != journal2.TotalDiscoveries {
		t.Errorf("TotalDiscoveries mismatch: %d vs %d", journal1.TotalDiscoveries, journal2.TotalDiscoveries)
	}

	if journal1.TotalSeriesComplete != journal2.TotalSeriesComplete {
		t.Errorf("TotalSeriesComplete mismatch: %d vs %d", journal1.TotalSeriesComplete, journal2.TotalSeriesComplete)
	}

	if !journal1.LastDiscoveryTime.Equal(journal2.LastDiscoveryTime) {
		t.Errorf("LastDiscoveryTime mismatch: %v vs %v", journal1.LastDiscoveryTime, journal2.LastDiscoveryTime)
	}
}
