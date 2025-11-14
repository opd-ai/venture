package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/story"
)

func TestNewDiscoverySystem(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	if system == nil {
		t.Fatal("NewDiscoverySystem() returned nil")
	}

	if system.world != world {
		t.Error("DiscoverySystem world not set correctly")
	}

	if system.discoveryRadius <= 0 {
		t.Errorf("discoveryRadius = %f, want > 0", system.discoveryRadius)
	}

	if system.seriesXPBonus < 0 {
		t.Errorf("seriesXPBonus = %f, want >= 0", system.seriesXPBonus)
	}

	if system.seriesFragments == nil {
		t.Error("seriesFragments map not initialized")
	}
}

func TestDiscoverySystem_RegisterSeries(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	seriesID := "test-series"
	fragmentCount := 5

	system.RegisterSeries(seriesID, fragmentCount)

	count := system.getSeriesFragmentCount(seriesID)
	if count != fragmentCount {
		t.Errorf("getSeriesFragmentCount() = %d, want %d", count, fragmentCount)
	}
}

func TestDiscoverySystem_SetDiscoveryRadius(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	// Set valid radius
	system.SetDiscoveryRadius(5.0)
	if system.discoveryRadius != 5.0 {
		t.Errorf("discoveryRadius = %f after SetDiscoveryRadius(5.0), want 5.0", system.discoveryRadius)
	}

	// Zero radius should be ignored
	system.SetDiscoveryRadius(0)
	if system.discoveryRadius != 5.0 {
		t.Errorf("discoveryRadius = %f after SetDiscoveryRadius(0), should not change", system.discoveryRadius)
	}

	// Negative radius should be ignored
	system.SetDiscoveryRadius(-10.0)
	if system.discoveryRadius != 5.0 {
		t.Errorf("discoveryRadius = %f after SetDiscoveryRadius(-10), should not change", system.discoveryRadius)
	}
}

func TestDiscoverySystem_SetSeriesXPBonus(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	// Set valid bonus
	system.SetSeriesXPBonus(200.0)
	if system.seriesXPBonus != 200.0 {
		t.Errorf("seriesXPBonus = %f, want 200.0", system.seriesXPBonus)
	}

	// Zero bonus should be allowed
	system.SetSeriesXPBonus(0)
	if system.seriesXPBonus != 0 {
		t.Errorf("seriesXPBonus = %f after SetSeriesXPBonus(0), want 0", system.seriesXPBonus)
	}

	// Negative bonus should be ignored
	system.seriesXPBonus = 100.0
	system.SetSeriesXPBonus(-50.0)
	if system.seriesXPBonus != 100.0 {
		t.Errorf("seriesXPBonus = %f after SetSeriesXPBonus(-50), should not change", system.seriesXPBonus)
	}
}

func TestDiscoverySystem_Update_NoDiscovery(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	// Create player with journal
	player := world.CreateEntity()
	player.AddComponent(NewStoryJournalComponent())
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&ExperienceComponent{})
	world.Update(0.0) // Commit entity

	// Create fragment far away
	fragment := world.CreateEntity()
	fragment.AddComponent(&PositionComponent{X: 100, Y: 100})
	fragment.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:        story.FragmentNote,
			Content:     "Test content",
			DiscoveryXP: 10.0,
			SeriesID:    "series1",
			SequenceNum: 0,
		},
		Discovered:  false,
		SeriesID:    "series1",
		SequenceNum: 0,
	})
	world.Update(0.0) // Commit entity

	// Update system
	system.Update(0.016)

	// Fragment should not be discovered
	fragCompInterface, ok := fragment.GetComponent("storyfragment")
	if !ok {
		t.Fatal("Fragment missing storyfragment component")
	}
	fragComp := fragCompInterface.(*StoryFragmentComponent)
	if fragComp.Discovered {
		t.Error("Fragment discovered when too far away")
	}
}

func TestDiscoverySystem_Update_DiscoverFragment(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)
	system.SetDiscoveryRadius(2.0)

	// Create player with journal and experience
	player := world.CreateEntity()
	journal := NewStoryJournalComponent()
	player.AddComponent(journal)
	player.AddComponent(&PositionComponent{X: 10, Y: 10})
	expComp := &ExperienceComponent{}
	player.AddComponent(expComp)
	world.Update(0.0) // Commit entity

	// Create fragment nearby
	fragment := world.CreateEntity()
	fragment.AddComponent(&PositionComponent{X: 11, Y: 10}) // 1 unit away
	fragComp := &StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:        story.FragmentNote,
			Content:     "Test content",
			DiscoveryXP: 50.0,
			SeriesID:    "series1",
			SequenceNum: 0,
		},
		Discovered:  false,
		SeriesID:    "series1",
		SequenceNum: 0,
	}
	fragment.AddComponent(fragComp)
	world.Update(0.0) // Commit entity

	// Register series
	system.RegisterSeries("series1", 3)

	// Update system
	system.Update(0.016)

	// Fragment should be discovered
	if !fragComp.Discovered {
		t.Error("Fragment not discovered when within range")
	}

	// Journal should be updated
	if !journal.IsDiscovered("series1", 0) {
		t.Error("Fragment not recorded in journal")
	}

	if journal.TotalDiscoveries != 1 {
		t.Errorf("TotalDiscoveries = %d, want 1", journal.TotalDiscoveries)
	}

	// XP should be awarded
	if expComp.CurrentXP != 50 {
		t.Errorf("CurrentXP = %d, want 50", expComp.CurrentXP)
	}
}

func TestDiscoverySystem_Update_CompleteSeriesBonus(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)
	system.SetDiscoveryRadius(5.0)
	system.SetSeriesXPBonus(100.0)

	// Create player
	player := world.CreateEntity()
	journal := NewStoryJournalComponent()
	player.AddComponent(journal)
	player.AddComponent(&PositionComponent{X: 10, Y: 10})
	expComp := &ExperienceComponent{}
	player.AddComponent(expComp)
	world.Update(0.0)

	// Create 3-fragment series
	seriesID := "complete-series"
	system.RegisterSeries(seriesID, 3)

	// Create fragments all nearby
	for i := 0; i < 3; i++ {
		frag := world.CreateEntity()
		frag.AddComponent(&PositionComponent{X: 10 + float64(i), Y: 10})
		frag.AddComponent(&StoryFragmentComponent{
			Fragment: story.StoryFragment{
				Type:        story.FragmentNote,
				Content:     "Fragment content",
				DiscoveryXP: 10.0,
				SeriesID:    seriesID,
				SequenceNum: i,
			},
			Discovered:  false,
			SeriesID:    seriesID,
			SequenceNum: i,
		})
		world.Update(0.0)
	}

	// Update system - should discover all fragments
	system.Update(0.016)

	// All fragments should be discovered
	if journal.TotalDiscoveries != 3 {
		t.Errorf("TotalDiscoveries = %d, want 3", journal.TotalDiscoveries)
	}

	// Series should be complete
	if !journal.CompletedSeries[seriesID] {
		t.Error("Series not marked as complete")
	}

	if journal.TotalSeriesComplete != 1 {
		t.Errorf("TotalSeriesComplete = %d, want 1", journal.TotalSeriesComplete)
	}

	// XP should include fragments + bonus
	expectedXP := 3*10 + 100 // 30 from fragments + 100 bonus
	if expComp.CurrentXP != expectedXP {
		t.Errorf("CurrentXP = %d, want %d", expComp.CurrentXP, expectedXP)
	}
}

func TestDiscoverySystem_GetDiscoveryStatus(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	// Create player with journal
	player := world.CreateEntity()
	journal := NewStoryJournalComponent()
	player.AddComponent(journal)
	world.Update(0.0)

	// Initial status
	discovered, series, err := system.GetDiscoveryStatus(player)
	if err != nil {
		t.Fatalf("GetDiscoveryStatus() error = %v", err)
	}

	if discovered != 0 || series != 0 {
		t.Errorf("GetDiscoveryStatus() = (%d, %d), want (0, 0)", discovered, series)
	}

	// Add some discoveries
	journal.AddDiscovery("series1", 0)
	journal.AddDiscovery("series1", 1)
	journal.AddDiscovery("series2", 0)
	journal.MarkSeriesComplete("series1")

	// Check updated status
	discovered, series, err = system.GetDiscoveryStatus(player)
	if err != nil {
		t.Fatalf("GetDiscoveryStatus() error = %v", err)
	}

	if discovered != 3 {
		t.Errorf("discovered = %d, want 3", discovered)
	}

	if series != 1 {
		t.Errorf("series = %d, want 1", series)
	}
}

func TestDiscoverySystem_GetDiscoveryStatus_NoJournal(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	// Create player without journal
	player := world.CreateEntity()
	world.Update(0.0)

	// Should return error
	_, _, err := system.GetDiscoveryStatus(player)
	if err == nil {
		t.Error("GetDiscoveryStatus() expected error for player without journal")
	}
}

func TestDiscoverySystem_CountFragmentsInWorld(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)

	seriesID := "world-series"

	// Create multiple fragments with same series
	for i := 0; i < 5; i++ {
		frag := world.CreateEntity()
		frag.AddComponent(&StoryFragmentComponent{
			Fragment: story.StoryFragment{
				SeriesID:    seriesID,
				SequenceNum: i,
			},
			SeriesID:    seriesID,
			SequenceNum: i,
		})
		world.Update(0.0)
	}

	// Create fragment with different series
	otherFrag := world.CreateEntity()
	otherFrag.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{
			SeriesID:    "other-series",
			SequenceNum: 0,
		},
		SeriesID:    "other-series",
		SequenceNum: 0,
	})
	world.Update(0.0)

	// Count should find 5 fragments in series
	count := system.countFragmentsInWorld(seriesID)
	if count != 5 {
		t.Errorf("countFragmentsInWorld() = %d, want 5", count)
	}
}

func TestDiscoverySystem_NoDoubleDiscovery(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)
	system.SetDiscoveryRadius(2.0)

	// Create player
	player := world.CreateEntity()
	journal := NewStoryJournalComponent()
	player.AddComponent(journal)
	player.AddComponent(&PositionComponent{X: 10, Y: 10})
	expComp := NewExperienceComponent()
	player.AddComponent(expComp)
	world.Update(0.0)

	// Register series with 3 fragments (so 1 discovery won't complete it)
	system.RegisterSeries("test", 3)

	// Create fragment
	fragment := world.CreateEntity()
	fragment.AddComponent(&PositionComponent{X: 11, Y: 10})
	fragment.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{
			Type:        story.FragmentNote,
			Content:     "Test",
			DiscoveryXP: 25.0,
			SeriesID:    "test",
			SequenceNum: 0,
		},
		Discovered:  false,
		SeriesID:    "test",
		SequenceNum: 0,
	})
	world.Update(0.0)

	// First update - should discover
	system.Update(0.016)

	if journal.TotalDiscoveries != 1 {
		t.Errorf("First discovery: TotalDiscoveries = %d, want 1", journal.TotalDiscoveries)
	}

	if expComp.CurrentXP != 25 {
		t.Errorf("First discovery: CurrentXP = %d, want 25", expComp.CurrentXP)
	}

	// Second update - should NOT discover again
	system.Update(0.016)

	if journal.TotalDiscoveries != 1 {
		t.Errorf("Second discovery: TotalDiscoveries = %d, should still be 1", journal.TotalDiscoveries)
	}

	if expComp.CurrentXP != 25 {
		t.Errorf("Second discovery: CurrentXP = %d, should still be 25", expComp.CurrentXP)
	}
}

func TestDiscoverySystem_MultiplePlayersIndependent(t *testing.T) {
	world := NewWorld()
	system := NewDiscoverySystem(world)
	system.SetDiscoveryRadius(2.0)

	// Create two players
	player1 := world.CreateEntity()
	journal1 := NewStoryJournalComponent()
	player1.AddComponent(journal1)
	player1.AddComponent(&PositionComponent{X: 10, Y: 10})
	player1.AddComponent(&ExperienceComponent{})
	world.Update(0.0)

	player2 := world.CreateEntity()
	journal2 := NewStoryJournalComponent()
	player2.AddComponent(journal2)
	player2.AddComponent(&PositionComponent{X: 100, Y: 100}) // Far away
	player2.AddComponent(&ExperienceComponent{})
	world.Update(0.0)

	// Create fragment near player1
	fragment := world.CreateEntity()
	fragment.AddComponent(&PositionComponent{X: 11, Y: 10})
	fragment.AddComponent(&StoryFragmentComponent{
		Fragment: story.StoryFragment{
			SeriesID:    "test",
			SequenceNum: 0,
			DiscoveryXP: 10.0,
		},
		SeriesID:    "test",
		SequenceNum: 0,
	})
	world.Update(0.0)

	// Update system
	system.Update(0.016)

	// Player1 should have discovered
	if journal1.TotalDiscoveries != 1 {
		t.Errorf("Player1 TotalDiscoveries = %d, want 1", journal1.TotalDiscoveries)
	}

	// Player2 should not have discovered
	if journal2.TotalDiscoveries != 0 {
		t.Errorf("Player2 TotalDiscoveries = %d, want 0", journal2.TotalDiscoveries)
	}
}
