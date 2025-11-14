package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/procgen/story"
)

// StoryFragmentComponent marks an entity as a discoverable story fragment.
// Story fragments are environmental storytelling elements that players can find
// and examine to piece together narratives about the dungeon's history.
type StoryFragmentComponent struct {
	Fragment      story.StoryFragment // The fragment data
	Discovered    bool                // Whether the player has found this fragment
	DiscoveryTime time.Time           // When the fragment was discovered
	SeriesID      string              // Links fragments into complete stories
	SequenceNum   int                 // Order within the story series
}

// Type returns the component type identifier.
func (s StoryFragmentComponent) Type() string {
	return "storyfragment"
}

// StoryJournalComponent tracks discovered fragments for a player.
// This component should be attached to the player entity.
type StoryJournalComponent struct {
	DiscoveredFragments map[string]bool // Fragment SeriesID + SequenceNum → discovered
	CompletedSeries     map[string]bool // SeriesID → completed
	TotalDiscoveries    int             // Count of discovered fragments
	TotalSeriesComplete int             // Count of completed story series
	LastDiscoveryTime   time.Time       // When last fragment was discovered
}

// Type returns the component type identifier.
func (j StoryJournalComponent) Type() string {
	return "storyjournal"
}

// NewStoryJournalComponent creates a new journal component for tracking discoveries.
func NewStoryJournalComponent() *StoryJournalComponent {
	return &StoryJournalComponent{
		DiscoveredFragments: make(map[string]bool),
		CompletedSeries:     make(map[string]bool),
		TotalDiscoveries:    0,
		TotalSeriesComplete: 0,
	}
}

// AddDiscovery records a fragment discovery in the journal.
// Returns true if this is a new discovery, false if already discovered.
func (j *StoryJournalComponent) AddDiscovery(seriesID string, sequenceNum int) bool {
	key := fragmentKey(seriesID, sequenceNum)
	if j.DiscoveredFragments[key] {
		return false // Already discovered
	}

	j.DiscoveredFragments[key] = true
	j.TotalDiscoveries++
	j.LastDiscoveryTime = time.Now()
	return true
}

// IsDiscovered checks if a fragment has been discovered.
func (j *StoryJournalComponent) IsDiscovered(seriesID string, sequenceNum int) bool {
	key := fragmentKey(seriesID, sequenceNum)
	return j.DiscoveredFragments[key]
}

// IsSeriesComplete checks if all fragments in a series have been discovered.
func (j *StoryJournalComponent) IsSeriesComplete(seriesID string, totalFragments int) bool {
	for i := 0; i < totalFragments; i++ {
		if !j.IsDiscovered(seriesID, i) {
			return false
		}
	}
	return true
}

// MarkSeriesComplete marks a story series as complete.
// Should be called after verifying IsSeriesComplete returns true.
func (j *StoryJournalComponent) MarkSeriesComplete(seriesID string) {
	if !j.CompletedSeries[seriesID] {
		j.CompletedSeries[seriesID] = true
		j.TotalSeriesComplete++
	}
}

// GetDiscoveryCount returns the number of discovered fragments in a series.
func (j *StoryJournalComponent) GetDiscoveryCount(seriesID string, totalFragments int) int {
	count := 0
	for i := 0; i < totalFragments; i++ {
		if j.IsDiscovered(seriesID, i) {
			count++
		}
	}
	return count
}

// fragmentKey generates a unique key for fragment lookup.
func fragmentKey(seriesID string, sequenceNum int) string {
	// Use simple string concatenation with separator
	// Format: "seriesID:sequenceNum"
	return seriesID + ":" + string(rune('0'+sequenceNum))
}
