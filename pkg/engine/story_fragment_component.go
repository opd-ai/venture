package engine

import (
	"fmt"
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
//
// Deprecated: Use JournalAddDiscovery helper function instead for ECS compliance.
// This method exists for backward compatibility but will be removed in future versions.
func (j *StoryJournalComponent) AddDiscovery(seriesID string, sequenceNum int) bool {
	return JournalAddDiscovery(j, seriesID, sequenceNum, time.Now())
}

// IsDiscovered checks if a fragment has been discovered.
//
// Deprecated: Use JournalIsDiscovered helper function instead for ECS compliance.
// This method exists for backward compatibility but will be removed in future versions.
func (j *StoryJournalComponent) IsDiscovered(seriesID string, sequenceNum int) bool {
	return JournalIsDiscovered(j, seriesID, sequenceNum)
}

// IsSeriesComplete checks if all fragments in a series have been discovered.
//
// Deprecated: Use JournalIsSeriesComplete helper function instead for ECS compliance.
// This method exists for backward compatibility but will be removed in future versions.
func (j *StoryJournalComponent) IsSeriesComplete(seriesID string, totalFragments int) bool {
	return JournalIsSeriesComplete(j, seriesID, totalFragments)
}

// MarkSeriesComplete marks a story series as complete.
// Should be called after verifying IsSeriesComplete returns true.
//
// Deprecated: Use JournalMarkSeriesComplete helper function instead for ECS compliance.
// This method exists for backward compatibility but will be removed in future versions.
func (j *StoryJournalComponent) MarkSeriesComplete(seriesID string) {
	JournalMarkSeriesComplete(j, seriesID)
}

// GetDiscoveryCount returns the number of discovered fragments in a series.
//
// Deprecated: Use JournalGetDiscoveryCount helper function instead for ECS compliance.
// This method exists for backward compatibility but will be removed in future versions.
func (j *StoryJournalComponent) GetDiscoveryCount(seriesID string, totalFragments int) int {
	return JournalGetDiscoveryCount(j, seriesID, totalFragments)
}

// --- ECS-Compliant Helper Functions ---
// These functions operate on StoryJournalComponent data without embedding behavior
// in the component itself, following ECS pure-data component principles.

// JournalAddDiscovery records a fragment discovery in the journal.
// Returns true if this is a new discovery, false if already discovered.
// The discoveryTime parameter enables deterministic testing and replay.
func JournalAddDiscovery(j *StoryJournalComponent, seriesID string, sequenceNum int, discoveryTime time.Time) bool {
	if j == nil {
		return false
	}
	key := fragmentKey(seriesID, sequenceNum)
	if j.DiscoveredFragments[key] {
		return false // Already discovered
	}

	j.DiscoveredFragments[key] = true
	j.TotalDiscoveries++
	j.LastDiscoveryTime = discoveryTime
	return true
}

// JournalIsDiscovered checks if a fragment has been discovered.
func JournalIsDiscovered(j *StoryJournalComponent, seriesID string, sequenceNum int) bool {
	if j == nil {
		return false
	}
	key := fragmentKey(seriesID, sequenceNum)
	return j.DiscoveredFragments[key]
}

// JournalIsSeriesComplete checks if all fragments in a series have been discovered.
func JournalIsSeriesComplete(j *StoryJournalComponent, seriesID string, totalFragments int) bool {
	if j == nil {
		return false
	}
	for i := 0; i < totalFragments; i++ {
		if !JournalIsDiscovered(j, seriesID, i) {
			return false
		}
	}
	return true
}

// JournalMarkSeriesComplete marks a story series as complete.
// Should be called after verifying JournalIsSeriesComplete returns true.
func JournalMarkSeriesComplete(j *StoryJournalComponent, seriesID string) {
	if j == nil {
		return
	}
	if !j.CompletedSeries[seriesID] {
		j.CompletedSeries[seriesID] = true
		j.TotalSeriesComplete++
	}
}

// JournalGetDiscoveryCount returns the number of discovered fragments in a series.
// totalFragments must match the total number of fragments in the series; only
// fragments with sequence numbers in [0, totalFragments) are counted.
func JournalGetDiscoveryCount(j *StoryJournalComponent, seriesID string, totalFragments int) int {
	if j == nil {
		return 0
	}
	count := 0
	for i := 0; i < totalFragments; i++ {
		if JournalIsDiscovered(j, seriesID, i) {
			count++
		}
	}
	return count
}

// fragmentKey generates a unique key for fragment lookup.
// Format: "seriesID:sequenceNum" where sequenceNum is formatted as a decimal integer.
// This supports sequence numbers of any size (not limited to single digits 0-9).
func fragmentKey(seriesID string, sequenceNum int) string {
	return fmt.Sprintf("%s:%d", seriesID, sequenceNum)
}
