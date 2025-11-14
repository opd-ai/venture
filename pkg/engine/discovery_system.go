package engine

import (
	"fmt"
	"math"

	"github.com/opd-ai/venture/pkg/procgen/quest"
	log "github.com/sirupsen/logrus"
)

// DiscoverySystem handles player interaction with story fragments.
// It detects when players are near fragments, handles discovery events,
// awards XP, and tracks series completion.
type DiscoverySystem struct {
	world           *World
	discoveryRadius float64        // Distance within which fragments can be discovered
	seriesXPBonus   float64        // Extra XP awarded for completing a series
	seriesFragments map[string]int // SeriesID → total fragment count
}

// NewDiscoverySystem creates a new discovery system.
func NewDiscoverySystem(world *World) *DiscoverySystem {
	return &DiscoverySystem{
		world:           world,
		discoveryRadius: 2.0,   // 2 tile radius for discovery
		seriesXPBonus:   100.0, // 100 XP bonus for completing a series
		seriesFragments: make(map[string]int),
	}
}

// Update checks for nearby fragments and processes discoveries.
// This should be called every frame with the time delta.
func (s *DiscoverySystem) Update(deltaTime float64) {
	// Get all fragment entities
	fragmentEntities := s.world.GetEntitiesWith("storyfragment")

	// Get all player entities (entities with storyjournal component)
	playerEntities := s.world.GetEntitiesWith("storyjournal", "position")

	for _, player := range playerEntities {
		playerPos, ok := player.GetComponent("position")
		if !ok {
			continue
		}

		posComp, ok := playerPos.(*PositionComponent)
		if !ok {
			continue
		}

		journal, ok := player.GetComponent("storyjournal")
		if !ok {
			continue
		}

		journalComp, ok := journal.(*StoryJournalComponent)
		if !ok {
			continue
		}

		// Check all fragments for discovery
		for _, fragment := range fragmentEntities {
			s.checkFragmentDiscovery(player, fragment, posComp, journalComp)
		}
	}
}

// checkFragmentDiscovery checks if a player should discover a fragment.
func (s *DiscoverySystem) checkFragmentDiscovery(player, fragment *Entity, playerPos *PositionComponent, journal *StoryJournalComponent) {
	// Get fragment position and component
	fragPosComp, ok := fragment.GetComponent("position")
	if !ok {
		return
	}

	fragPos, ok := fragPosComp.(*PositionComponent)
	if !ok {
		return
	}

	// Get fragment component
	fragComp, ok := fragment.GetComponent("storyfragment")
	if !ok {
		return
	}

	storyFragComp, ok := fragComp.(*StoryFragmentComponent)
	if !ok {
		return
	}

	// Skip if already discovered
	if storyFragComp.Discovered {
		return
	}

	// Check distance
	dx := playerPos.X - fragPos.X
	dy := playerPos.Y - fragPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance <= s.discoveryRadius {
		// Discover the fragment
		s.discoverFragment(player, fragment, storyFragComp, journal)
	}
}

// discoverFragment processes a fragment discovery event.
func (s *DiscoverySystem) discoverFragment(player, fragment *Entity, fragComp *StoryFragmentComponent, journal *StoryJournalComponent) {
	// Mark fragment as discovered
	fragComp.Discovered = true

	// Add to journal
	isNew := journal.AddDiscovery(fragComp.SeriesID, fragComp.SequenceNum)
	if !isNew {
		return // Already in journal (shouldn't happen but defensive)
	}

	// Award discovery XP
	s.awardDiscoveryXP(player, fragComp.Fragment.DiscoveryXP)

	// Check if series is complete
	totalFragments := s.getSeriesFragmentCount(fragComp.SeriesID)
	if journal.IsSeriesComplete(fragComp.SeriesID, totalFragments) {
		// Mark series complete
		journal.MarkSeriesComplete(fragComp.SeriesID)

		// Award series completion bonus
		s.awardDiscoveryXP(player, s.seriesXPBonus)

		// Unlock any story-linked quests (Phase 30.2)
		s.unlockStoryQuests(player, fragComp.SeriesID)
	}
}

// awardDiscoveryXP awards experience points to a player entity.
func (s *DiscoverySystem) awardDiscoveryXP(player *Entity, xpAmount float64) {
	expComp, ok := player.GetComponent("experience")
	if !ok {
		return
	}

	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		return
	}

	// Convert float64 XP to int
	exp.AddXP(int(xpAmount))
}

// RegisterSeries registers a story series with its fragment count.
// This should be called when generating story content for a dungeon.
func (s *DiscoverySystem) RegisterSeries(seriesID string, fragmentCount int) {
	s.seriesFragments[seriesID] = fragmentCount
}

// getSeriesFragmentCount returns the total number of fragments in a series.
func (s *DiscoverySystem) getSeriesFragmentCount(seriesID string) int {
	count, exists := s.seriesFragments[seriesID]
	if !exists {
		// Default to counting fragments in world if not registered
		return s.countFragmentsInWorld(seriesID)
	}
	return count
}

// countFragmentsInWorld counts fragments with a specific seriesID in the world.
func (s *DiscoverySystem) countFragmentsInWorld(seriesID string) int {
	count := 0
	fragments := s.world.GetEntitiesWith("storyfragment")

	for _, frag := range fragments {
		fragComp, ok := frag.GetComponent("storyfragment")
		if !ok {
			continue
		}

		storyFrag, ok := fragComp.(*StoryFragmentComponent)
		if !ok {
			continue
		}

		if storyFrag.SeriesID == seriesID {
			count++
		}
	}

	return count
}

// GetDiscoveryStatus returns discovery information for a player.
func (s *DiscoverySystem) GetDiscoveryStatus(player *Entity) (totalDiscovered, totalSeries int, err error) {
	journal, ok := player.GetComponent("storyjournal")
	if !ok {
		return 0, 0, fmt.Errorf("player has no story journal component")
	}

	journalComp, ok := journal.(*StoryJournalComponent)
	if !ok {
		return 0, 0, fmt.Errorf("invalid journal component type")
	}

	return journalComp.TotalDiscoveries, journalComp.TotalSeriesComplete, nil
}

// SetDiscoveryRadius sets the radius within which fragments can be discovered.
func (s *DiscoverySystem) SetDiscoveryRadius(radius float64) {
	if radius > 0 {
		s.discoveryRadius = radius
	}
}

// SetSeriesXPBonus sets the XP bonus awarded for completing a series.
func (s *DiscoverySystem) SetSeriesXPBonus(bonus float64) {
	if bonus >= 0 {
		s.seriesXPBonus = bonus
	}
}

// unlockStoryQuests checks if the completed series unlocks any quests.
// This is called when a story series is completed (Phase 30.2).
func (s *DiscoverySystem) unlockStoryQuests(player *Entity, seriesID string) {
	questTrackerComp, ok := player.GetComponent("questtracker")
	if !ok {
		return // Player doesn't have quest tracker
	}

	questTracker, ok := questTrackerComp.(*QuestTrackerComponent)
	if !ok {
		return
	}

	// Unlock quests for this series
	// The questGenerator would need to be provided by a quest generation system
	// For now, we'll just use a nil generator which means no quests are actually generated
	// In a full implementation, this would call into a quest generation system
	unlockedCount := questTracker.UnlockStoryQuests(seriesID, func(questID string) *quest.Quest {
		// TODO: Integrate with quest generation system
		// For now, return nil (no quests unlocked)
		return nil
	})

	if unlockedCount > 0 {
		log.WithFields(log.Fields{
			"player":   player.ID,
			"seriesID": seriesID,
			"unlocked": unlockedCount,
		}).Info("Story completion unlocked new quests")
	}
}
