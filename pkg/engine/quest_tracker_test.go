package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// TestQuestTrackerComponent_RegisterStoryQuest tests registering story-unlocked quests.
func TestQuestTrackerComponent_RegisterStoryQuest(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	tracker.RegisterStoryQuest("ancient_curse", "quest_001")
	tracker.RegisterStoryQuest("ancient_curse", "quest_002")
	tracker.RegisterStoryQuest("fallen_kingdom", "quest_003")

	if len(tracker.StoryUnlockedQuests) != 2 {
		t.Errorf("Expected 2 series with registered quests, got %d", len(tracker.StoryUnlockedQuests))
	}

	ancientCurseQuests := tracker.StoryUnlockedQuests["ancient_curse"]
	if len(ancientCurseQuests) != 2 {
		t.Errorf("Expected 2 quests for ancient_curse, got %d", len(ancientCurseQuests))
	}

	if ancientCurseQuests[0] != "quest_001" {
		t.Errorf("Expected quest_001, got %s", ancientCurseQuests[0])
	}

	if ancientCurseQuests[1] != "quest_002" {
		t.Errorf("Expected quest_002, got %s", ancientCurseQuests[1])
	}
}

// TestQuestTrackerComponent_UnlockStoryQuests tests unlocking quests.
func TestQuestTrackerComponent_UnlockStoryQuests(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	// Register quests
	tracker.RegisterStoryQuest("ancient_curse", "quest_001")
	tracker.RegisterStoryQuest("ancient_curse", "quest_002")

	// Create quest generator
	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{
			ID:          questID,
			Name:        "Test Quest " + questID,
			Description: "Description for " + questID,
			Type:        quest.TypeKill,
		}
	}

	// Unlock quests for series
	count := tracker.UnlockStoryQuests("ancient_curse", questGen)

	if count != 2 {
		t.Errorf("Expected 2 quests unlocked, got %d", count)
	}

	if len(tracker.PendingStoryQuests) != 2 {
		t.Errorf("Expected 2 pending quests, got %d", len(tracker.PendingStoryQuests))
	}

	// Verify quest data
	if tracker.PendingStoryQuests[0].ID != "quest_001" {
		t.Errorf("Expected quest_001, got %s", tracker.PendingStoryQuests[0].ID)
	}

	if tracker.PendingStoryQuests[1].ID != "quest_002" {
		t.Errorf("Expected quest_002, got %s", tracker.PendingStoryQuests[1].ID)
	}
}

// TestQuestTrackerComponent_UnlockStoryQuests_NoQuestsForSeries tests series with no quests.
func TestQuestTrackerComponent_UnlockStoryQuests_NoQuestsForSeries(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{ID: questID}
	}

	// Unlock for non-existent series
	count := tracker.UnlockStoryQuests("nonexistent", questGen)

	if count != 0 {
		t.Errorf("Expected 0 quests unlocked, got %d", count)
	}

	if len(tracker.PendingStoryQuests) != 0 {
		t.Errorf("Expected 0 pending quests, got %d", len(tracker.PendingStoryQuests))
	}
}

// TestQuestTrackerComponent_UnlockStoryQuests_DuplicatePrevention tests duplicate prevention.
func TestQuestTrackerComponent_UnlockStoryQuests_DuplicatePrevention(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	tracker.RegisterStoryQuest("test_series", "quest_001")

	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{ID: questID}
	}

	// Unlock once
	count1 := tracker.UnlockStoryQuests("test_series", questGen)

	if count1 != 1 {
		t.Errorf("First unlock: expected 1, got %d", count1)
	}

	// Try unlocking again (should not add duplicates)
	count2 := tracker.UnlockStoryQuests("test_series", questGen)

	if count2 != 0 {
		t.Errorf("Second unlock: expected 0, got %d", count2)
	}

	if len(tracker.PendingStoryQuests) != 1 {
		t.Errorf("Expected 1 pending quest, got %d", len(tracker.PendingStoryQuests))
	}
}

// TestQuestTrackerComponent_UnlockStoryQuests_NilGenerator tests nil quest handling.
func TestQuestTrackerComponent_UnlockStoryQuests_NilGenerator(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	tracker.RegisterStoryQuest("test_series", "quest_001")

	// Generator returns nil
	questGen := func(questID string) *quest.Quest {
		return nil
	}

	count := tracker.UnlockStoryQuests("test_series", questGen)

	if count != 0 {
		t.Errorf("Expected 0 quests unlocked (nil quest), got %d", count)
	}

	if len(tracker.PendingStoryQuests) != 0 {
		t.Errorf("Expected 0 pending quests, got %d", len(tracker.PendingStoryQuests))
	}
}

// TestQuestTrackerComponent_GetPendingStoryQuests tests retrieving pending quests.
func TestQuestTrackerComponent_GetPendingStoryQuests(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	// Initially empty
	pending := tracker.GetPendingStoryQuests()
	if len(pending) != 0 {
		t.Errorf("Expected 0 pending quests initially, got %d", len(pending))
	}

	// Add some quests
	tracker.PendingStoryQuests = []*quest.Quest{
		{ID: "quest_001"},
		{ID: "quest_002"},
	}

	pending = tracker.GetPendingStoryQuests()
	if len(pending) != 2 {
		t.Errorf("Expected 2 pending quests, got %d", len(pending))
	}
}

// TestQuestTrackerComponent_ClearPendingStoryQuest tests clearing individual quests.
func TestQuestTrackerComponent_ClearPendingStoryQuest(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	tracker.PendingStoryQuests = []*quest.Quest{
		{ID: "quest_001"},
		{ID: "quest_002"},
		{ID: "quest_003"},
	}

	// Clear middle quest
	tracker.ClearPendingStoryQuest("quest_002")

	if len(tracker.PendingStoryQuests) != 2 {
		t.Errorf("Expected 2 quests after clearing, got %d", len(tracker.PendingStoryQuests))
	}

	// Verify correct quest was removed
	for _, q := range tracker.PendingStoryQuests {
		if q.ID == "quest_002" {
			t.Error("quest_002 should have been removed")
		}
	}

	// Clear non-existent quest (should not crash)
	tracker.ClearPendingStoryQuest("nonexistent")

	if len(tracker.PendingStoryQuests) != 2 {
		t.Errorf("Expected 2 quests after clearing nonexistent, got %d", len(tracker.PendingStoryQuests))
	}
}

// TestQuestTrackerComponent_HasPendingStoryQuests tests pending quest check.
func TestQuestTrackerComponent_HasPendingStoryQuests(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	// Initially false
	if tracker.HasPendingStoryQuests() {
		t.Error("Expected false initially")
	}

	// Add quest
	tracker.PendingStoryQuests = []*quest.Quest{
		{ID: "quest_001"},
	}

	if !tracker.HasPendingStoryQuests() {
		t.Error("Expected true after adding quest")
	}

	// Clear all
	tracker.PendingStoryQuests = []*quest.Quest{}

	if tracker.HasPendingStoryQuests() {
		t.Error("Expected false after clearing all")
	}
}

// TestQuestTrackerComponent_Integration tests full workflow.
func TestQuestTrackerComponent_Integration(t *testing.T) {
	tracker := NewQuestTrackerComponent(5)

	// Register quests for multiple series
	tracker.RegisterStoryQuest("series1", "quest_A")
	tracker.RegisterStoryQuest("series1", "quest_B")
	tracker.RegisterStoryQuest("series2", "quest_C")

	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{
			ID:          questID,
			Name:        "Quest " + questID,
			Description: "Description",
			Type:        quest.TypeCollect,
		}
	}

	// Unlock series1 quests
	count1 := tracker.UnlockStoryQuests("series1", questGen)
	if count1 != 2 {
		t.Errorf("Expected 2 quests from series1, got %d", count1)
	}

	// Unlock series2 quests
	count2 := tracker.UnlockStoryQuests("series2", questGen)
	if count2 != 1 {
		t.Errorf("Expected 1 quest from series2, got %d", count2)
	}

	// Check total pending
	if len(tracker.PendingStoryQuests) != 3 {
		t.Errorf("Expected 3 total pending quests, got %d", len(tracker.PendingStoryQuests))
	}

	// Accept one quest
	tracker.AcceptQuest(tracker.PendingStoryQuests[0], 1234567890)
	tracker.ClearPendingStoryQuest(tracker.PendingStoryQuests[0].ID)

	// Check active and pending counts
	if len(tracker.ActiveQuests) != 1 {
		t.Errorf("Expected 1 active quest, got %d", len(tracker.ActiveQuests))
	}

	if len(tracker.PendingStoryQuests) != 2 {
		t.Errorf("Expected 2 pending quests after accepting one, got %d", len(tracker.PendingStoryQuests))
	}
}

// TestQuestTrackerComponent_MultipleSeriesUnlock tests unlocking multiple series.
func TestQuestTrackerComponent_MultipleSeriesUnlock(t *testing.T) {
	tracker := NewQuestTrackerComponent(10)

	// Register quests for 5 different series
	for i := 1; i <= 5; i++ {
		seriesID := string(rune('A' + i - 1))
		tracker.RegisterStoryQuest(seriesID, "quest_"+seriesID+"1")
		tracker.RegisterStoryQuest(seriesID, "quest_"+seriesID+"2")
	}

	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{ID: questID}
	}

	// Unlock all series
	totalUnlocked := 0
	for i := 1; i <= 5; i++ {
		seriesID := string(rune('A' + i - 1))
		count := tracker.UnlockStoryQuests(seriesID, questGen)
		totalUnlocked += count
	}

	if totalUnlocked != 10 {
		t.Errorf("Expected 10 total quests unlocked, got %d", totalUnlocked)
	}

	if len(tracker.PendingStoryQuests) != 10 {
		t.Errorf("Expected 10 pending quests, got %d", len(tracker.PendingStoryQuests))
	}
}

// Benchmark quest tracker story operations
func BenchmarkQuestTrackerComponent_RegisterStoryQuest(b *testing.B) {
	tracker := NewQuestTrackerComponent(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.RegisterStoryQuest("series", "quest")
	}
}

func BenchmarkQuestTrackerComponent_UnlockStoryQuests(b *testing.B) {
	tracker := NewQuestTrackerComponent(100)

	for i := 0; i < 10; i++ {
		tracker.RegisterStoryQuest("series", "quest_"+string(rune('0'+i)))
	}

	questGen := func(questID string) *quest.Quest {
		return &quest.Quest{ID: questID}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset pending quests each iteration
		tracker.PendingStoryQuests = []*quest.Quest{}
		tracker.UnlockStoryQuests("series", questGen)
	}
}
