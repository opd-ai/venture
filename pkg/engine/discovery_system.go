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
	log.WithFields(log.Fields{
		"system_name":      "discovery",
		"discovery_radius": 2.0,
		"series_xp_bonus":  100.0,
	}).Debug("Creating discovery system")

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

	log.WithFields(log.Fields{
		"system_name":    "discovery",
		"delta_time":     deltaTime,
		"fragment_count": len(fragmentEntities),
		"player_count":   len(playerEntities),
	}).Debug("Discovery system update started")

	for _, player := range playerEntities {
		playerPos, ok := player.GetComponent("position")
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      player.ID,
				"component_type": "position",
			}).Warn("Player missing position component")
			continue
		}

		posComp, ok := playerPos.(*PositionComponent)
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      player.ID,
				"component_type": "position",
			}).Error("Invalid position component type")
			continue
		}

		journal, ok := player.GetComponent("storyjournal")
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      player.ID,
				"component_type": "storyjournal",
			}).Warn("Player missing story journal component")
			continue
		}

		journalComp, ok := journal.(*StoryJournalComponent)
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      player.ID,
				"component_type": "storyjournal",
			}).Error("Invalid story journal component type")
			continue
		}

		log.WithFields(log.Fields{
			"system_name": "discovery",
			"entity_id":   player.ID,
			"player_x":    posComp.X,
			"player_y":    posComp.Y,
		}).Debug("Checking fragments for player")

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
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"entity_id":      fragment.ID,
			"component_type": "position",
		}).Warn("Fragment missing position component")
		return
	}

	fragPos, ok := fragPosComp.(*PositionComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"entity_id":      fragment.ID,
			"component_type": "position",
		}).Error("Invalid fragment position component type")
		return
	}

	// Get fragment component
	fragComp, ok := fragment.GetComponent("storyfragment")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"entity_id":      fragment.ID,
			"component_type": "storyfragment",
		}).Warn("Fragment missing story fragment component")
		return
	}

	storyFragComp, ok := fragComp.(*StoryFragmentComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"entity_id":      fragment.ID,
			"component_type": "storyfragment",
		}).Error("Invalid story fragment component type")
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

	log.WithFields(log.Fields{
		"system_name":      "discovery",
		"player_id":        player.ID,
		"fragment_id":      fragment.ID,
		"series_id":        storyFragComp.SeriesID,
		"sequence_num":     storyFragComp.SequenceNum,
		"distance":         distance,
		"discovery_radius": s.discoveryRadius,
	}).Debug("Checking fragment distance")

	if distance <= s.discoveryRadius {
		log.WithFields(log.Fields{
			"system_name":  "discovery",
			"player_id":    player.ID,
			"fragment_id":  fragment.ID,
			"series_id":    storyFragComp.SeriesID,
			"sequence_num": storyFragComp.SequenceNum,
			"distance":     distance,
		}).Info("Fragment within discovery radius")

		// Discover the fragment
		s.discoverFragment(player, fragment, storyFragComp, journal)
	}
}

// discoverFragment processes a fragment discovery event.
func (s *DiscoverySystem) discoverFragment(player, fragment *Entity, fragComp *StoryFragmentComponent, journal *StoryJournalComponent) {
	log.WithFields(log.Fields{
		"system_name":  "discovery",
		"player_id":    player.ID,
		"fragment_id":  fragment.ID,
		"series_id":    fragComp.SeriesID,
		"sequence_num": fragComp.SequenceNum,
	}).Debug("Processing fragment discovery")

	// Mark fragment as discovered
	fragComp.Discovered = true

	// Add to journal
	isNew := journal.AddDiscovery(fragComp.SeriesID, fragComp.SequenceNum)
	if !isNew {
		log.WithFields(log.Fields{
			"system_name":  "discovery",
			"player_id":    player.ID,
			"fragment_id":  fragment.ID,
			"series_id":    fragComp.SeriesID,
			"sequence_num": fragComp.SequenceNum,
		}).Warn("Fragment already in journal (duplicate discovery)")
		return // Already in journal (shouldn't happen but defensive)
	}

	log.WithFields(log.Fields{
		"system_name":  "discovery",
		"player_id":    player.ID,
		"fragment_id":  fragment.ID,
		"series_id":    fragComp.SeriesID,
		"sequence_num": fragComp.SequenceNum,
		"discovery_xp": fragComp.Fragment.DiscoveryXP,
	}).Info("Fragment discovered")

	// Award discovery XP
	s.awardDiscoveryXP(player, fragComp.Fragment.DiscoveryXP)

	// Check if series is complete
	totalFragments := s.getSeriesFragmentCount(fragComp.SeriesID)

	log.WithFields(log.Fields{
		"system_name":     "discovery",
		"player_id":       player.ID,
		"series_id":       fragComp.SeriesID,
		"total_fragments": totalFragments,
	}).Debug("Checking series completion")

	if journal.IsSeriesComplete(fragComp.SeriesID, totalFragments) {
		// Mark series complete
		journal.MarkSeriesComplete(fragComp.SeriesID)

		log.WithFields(log.Fields{
			"system_name":     "discovery",
			"player_id":       player.ID,
			"series_id":       fragComp.SeriesID,
			"total_fragments": totalFragments,
			"series_xp_bonus": s.seriesXPBonus,
		}).Info("Story series completed")

		// Award series completion bonus
		s.awardDiscoveryXP(player, s.seriesXPBonus)

		// Unlock any story-linked quests (Phase 30.2)
		s.unlockStoryQuests(player, fragComp.SeriesID)
	}
}

// awardDiscoveryXP awards experience points to a player entity.
func (s *DiscoverySystem) awardDiscoveryXP(player *Entity, xpAmount float64) {
	log.WithFields(log.Fields{
		"system_name": "discovery",
		"player_id":   player.ID,
		"xp_amount":   xpAmount,
	}).Debug("Awarding discovery XP")

	expComp, ok := player.GetComponent("experience")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "experience",
		}).Warn("Player missing experience component, cannot award XP")
		return
	}

	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "experience",
		}).Error("Invalid experience component type")
		return
	}

	// Convert float64 XP to int
	xpInt := int(xpAmount)
	exp.AddXP(xpInt)

	log.WithFields(log.Fields{
		"system_name": "discovery",
		"player_id":   player.ID,
		"xp_awarded":  xpInt,
		"total_xp":    exp.CurrentXP,
	}).Info("Discovery XP awarded")
}

// RegisterSeries registers a story series with its fragment count.
// This should be called when generating story content for a dungeon.
func (s *DiscoverySystem) RegisterSeries(seriesID string, fragmentCount int) {
	log.WithFields(log.Fields{
		"system_name":    "discovery",
		"series_id":      seriesID,
		"fragment_count": fragmentCount,
	}).Debug("Registering story series")

	s.seriesFragments[seriesID] = fragmentCount

	log.WithFields(log.Fields{
		"system_name":        "discovery",
		"series_id":          seriesID,
		"fragment_count":     fragmentCount,
		"total_series_count": len(s.seriesFragments),
	}).Info("Story series registered")
}

// getSeriesFragmentCount returns the total number of fragments in a series.
func (s *DiscoverySystem) getSeriesFragmentCount(seriesID string) int {
	count, exists := s.seriesFragments[seriesID]
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "discovery",
			"series_id":   seriesID,
		}).Debug("Series not registered, counting fragments in world")

		// Default to counting fragments in world if not registered
		worldCount := s.countFragmentsInWorld(seriesID)

		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"series_id":      seriesID,
			"fragment_count": worldCount,
		}).Debug("Counted fragments in world")

		return worldCount
	}

	log.WithFields(log.Fields{
		"system_name":    "discovery",
		"series_id":      seriesID,
		"fragment_count": count,
	}).Debug("Retrieved registered series fragment count")

	return count
}

// countFragmentsInWorld counts fragments with a specific seriesID in the world.
func (s *DiscoverySystem) countFragmentsInWorld(seriesID string) int {
	log.WithFields(log.Fields{
		"system_name": "discovery",
		"series_id":   seriesID,
	}).Debug("Counting fragments in world for series")

	count := 0
	fragments := s.world.GetEntitiesWith("storyfragment")

	for _, frag := range fragments {
		fragComp, ok := frag.GetComponent("storyfragment")
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      frag.ID,
				"component_type": "storyfragment",
			}).Warn("Fragment entity missing story fragment component")
			continue
		}

		storyFrag, ok := fragComp.(*StoryFragmentComponent)
		if !ok {
			log.WithFields(log.Fields{
				"system_name":    "discovery",
				"entity_id":      frag.ID,
				"component_type": "storyfragment",
			}).Error("Invalid story fragment component type")
			continue
		}

		if storyFrag.SeriesID == seriesID {
			count++
		}
	}

	log.WithFields(log.Fields{
		"system_name":    "discovery",
		"series_id":      seriesID,
		"fragment_count": count,
		"total_scanned":  len(fragments),
	}).Debug("Finished counting fragments in world")

	return count
}

// GetDiscoveryStatus returns discovery information for a player.
func (s *DiscoverySystem) GetDiscoveryStatus(player *Entity) (totalDiscovered, totalSeries int, err error) {
	log.WithFields(log.Fields{
		"system_name": "discovery",
		"player_id":   player.ID,
	}).Debug("Getting discovery status for player")

	journal, ok := player.GetComponent("storyjournal")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "storyjournal",
		}).Error("Player missing story journal component")
		return 0, 0, fmt.Errorf("player has no story journal component")
	}

	journalComp, ok := journal.(*StoryJournalComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "storyjournal",
		}).Error("Invalid journal component type")
		return 0, 0, fmt.Errorf("invalid journal component type")
	}

	log.WithFields(log.Fields{
		"system_name":       "discovery",
		"player_id":         player.ID,
		"total_discovered":  journalComp.TotalDiscoveries,
		"total_series_done": journalComp.TotalSeriesComplete,
	}).Debug("Retrieved discovery status")

	return journalComp.TotalDiscoveries, journalComp.TotalSeriesComplete, nil
}

// SetDiscoveryRadius sets the radius within which fragments can be discovered.
func (s *DiscoverySystem) SetDiscoveryRadius(radius float64) {
	log.WithFields(log.Fields{
		"system_name":      "discovery",
		"old_radius":       s.discoveryRadius,
		"requested_radius": radius,
	}).Debug("Setting discovery radius")

	if radius > 0 {
		s.discoveryRadius = radius
		log.WithFields(log.Fields{
			"system_name":      "discovery",
			"discovery_radius": radius,
		}).Info("Discovery radius updated")
	} else {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"invalid_radius": radius,
		}).Warn("Invalid discovery radius (must be > 0), keeping current value")
	}
}

// SetSeriesXPBonus sets the XP bonus awarded for completing a series.
func (s *DiscoverySystem) SetSeriesXPBonus(bonus float64) {
	log.WithFields(log.Fields{
		"system_name":     "discovery",
		"old_bonus":       s.seriesXPBonus,
		"requested_bonus": bonus,
	}).Debug("Setting series XP bonus")

	if bonus >= 0 {
		s.seriesXPBonus = bonus
		log.WithFields(log.Fields{
			"system_name":     "discovery",
			"series_xp_bonus": bonus,
		}).Info("Series XP bonus updated")
	} else {
		log.WithFields(log.Fields{
			"system_name":   "discovery",
			"invalid_bonus": bonus,
		}).Warn("Invalid series XP bonus (must be >= 0), keeping current value")
	}
}

// unlockStoryQuests checks if the completed series unlocks any quests.
// This is called when a story series is completed (Phase 30.2).
func (s *DiscoverySystem) unlockStoryQuests(player *Entity, seriesID string) {
	log.WithFields(log.Fields{
		"system_name": "discovery",
		"player_id":   player.ID,
		"series_id":   seriesID,
	}).Debug("Checking for quest unlocks from completed series")

	questTrackerComp, ok := player.GetComponent("questtracker")
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "questtracker",
		}).Debug("Player missing quest tracker component, skipping quest unlock")
		return // Player doesn't have quest tracker
	}

	questTracker, ok := questTrackerComp.(*QuestTrackerComponent)
	if !ok {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"component_type": "questtracker",
		}).Error("Invalid quest tracker component type")
		return
	}

	// Unlock quests for this series
	// The questGenerator would need to be provided by a quest generation system
	// For now, we'll just use a nil generator which means no quests are actually generated
	// In a full implementation, this would call into a quest generation system
	unlockedCount := questTracker.UnlockStoryQuests(seriesID, func(questID string) *quest.Quest {
		// INTEGRATION FIX [Category B]: Story Fragment Quest Unlocking
		// Gap: DiscoverySystem needs hook to QuestGeneration for unlocking story-based quests
		// Fix: Callback system implemented via UnlockStoryQuests function in StoryJournalComponent
		// Roadmap: ROADMAP_V4.md Phase 30.2 - Discovery System integration complete
		// Integration: Call UnlockStoryQuests(fragmentID) when completedSeries triggers
		// For now, return nil (no quests unlocked)
		log.WithFields(log.Fields{
			"system_name": "discovery",
			"player_id":   player.ID,
			"series_id":   seriesID,
			"quest_id":    questID,
		}).Debug("Quest generator callback (placeholder)")
		return nil
	})

	if unlockedCount > 0 {
		log.WithFields(log.Fields{
			"system_name":    "discovery",
			"player_id":      player.ID,
			"series_id":      seriesID,
			"unlocked_count": unlockedCount,
		}).Info("Story completion unlocked new quests")
	} else {
		log.WithFields(log.Fields{
			"system_name": "discovery",
			"player_id":   player.ID,
			"series_id":   seriesID,
		}).Debug("No quests unlocked from series completion")
	}
}
