//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestCompetitiveSystemsInitialization verifies that competitive PvP systems are properly initialized
func TestCompetitiveSystemsInitialization(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(12345)
	logger := createTestLogger()

	initialSystemCount := len(world.GetSystems())

	// Initialize V4 systems which includes competitive systems
	_, _ = initializeV4Systems(world, seed, "fantasy", logger, nil)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// V4 should add 31 systems (including 4 competitive systems)
	expectedMinSystems := 27 // At least 27 systems should be added
	if addedSystems < expectedMinSystems {
		t.Errorf("initializeV4Systems added %d systems, expected at least %d", addedSystems, expectedMinSystems)
	}

	// Verify all 4 competitive systems are added to the world
	systems := world.GetSystems()
	if len(systems) == 0 {
		t.Fatal("No systems were initialized")
	}

	// Count competitive systems by checking their types
	var raidSystemCount, tournamentSystemCount, pvpRatingSystemCount, legendaryQuestSystemCount int

	for _, sys := range systems {
		switch sys.(type) {
		case *engine.RaidSystem:
			raidSystemCount++
		case *engine.TournamentSystem:
			tournamentSystemCount++
		case *engine.PvPRatingSystem:
			pvpRatingSystemCount++
		case *engine.LegendaryQuestSystem:
			legendaryQuestSystemCount++
		}
	}

	// Verify each competitive system is present exactly once
	if raidSystemCount != 1 {
		t.Errorf("Expected 1 RaidSystem, got %d", raidSystemCount)
	}
	if tournamentSystemCount != 1 {
		t.Errorf("Expected 1 TournamentSystem, got %d", tournamentSystemCount)
	}
	if pvpRatingSystemCount != 1 {
		t.Errorf("Expected 1 PvPRatingSystem, got %d", pvpRatingSystemCount)
	}
	if legendaryQuestSystemCount != 1 {
		t.Errorf("Expected 1 LegendaryQuestSystem, got %d", legendaryQuestSystemCount)
	}

	t.Logf("Competitive systems initialized successfully: Raid=%d, Tournament=%d, PvPRating=%d, LegendaryQuest=%d",
		raidSystemCount, tournamentSystemCount, pvpRatingSystemCount, legendaryQuestSystemCount)
}

// TestCompetitiveSystemsDeterminism verifies systems produce deterministic results with same seed
func TestCompetitiveSystemsDeterminism(t *testing.T) {
	seed := int64(44444)
	logger := createTestLogger()

	// Initialize two worlds with same seed
	world1 := engine.NewWorld()
	initializeV4Systems(world1, seed, "fantasy", logger, nil)

	world2 := engine.NewWorld()
	initializeV4Systems(world2, seed, "fantasy", logger, nil)

	// Both worlds should have the same number of systems
	systems1 := world1.GetSystems()
	systems2 := world2.GetSystems()

	if len(systems1) != len(systems2) {
		t.Errorf("Same seed should produce same number of systems: %d vs %d", len(systems1), len(systems2))
	}

	// Count competitive systems in both worlds
	countSystems := func(systems []engine.System) (raid, tournament, pvp, quest int) {
		for _, sys := range systems {
			switch sys.(type) {
			case *engine.RaidSystem:
				raid++
			case *engine.TournamentSystem:
				tournament++
			case *engine.PvPRatingSystem:
				pvp++
			case *engine.LegendaryQuestSystem:
				quest++
			}
		}
		return raid, tournament, pvp, quest
	}

	r1, t1, p1, q1 := countSystems(systems1)
	r2, t2, p2, q2 := countSystems(systems2)

	if r1 != r2 || t1 != t2 || p1 != p2 || q1 != q2 {
		t.Errorf("System counts should match: (%d,%d,%d,%d) vs (%d,%d,%d,%d)",
			r1, t1, p1, q1, r2, t2, p2, q2)
	}

	t.Logf("Determinism verified: both worlds have (%d,%d,%d,%d) competitive systems",
		r1, t1, p1, q1)
}

// TestCompetitiveSystemsWithDifferentSeeds verifies different seeds produce same system count
func TestCompetitiveSystemsWithDifferentSeeds(t *testing.T) {
	logger := createTestLogger()

	seeds := []int64{12345, 54321, 99999, 11111}
	systemCounts := make([]int, len(seeds))

	for i, seed := range seeds {
		world := engine.NewWorld()
		initializeV4Systems(world, seed, "fantasy", logger, nil)
		systemCounts[i] = len(world.GetSystems())
	}

	// All worlds should have the same number of systems regardless of seed
	expectedCount := systemCounts[0]
	for i, count := range systemCounts {
		if count != expectedCount {
			t.Errorf("Seed %d produced %d systems, expected %d", seeds[i], count, expectedCount)
		}
	}

	t.Logf("All seeds produce consistent system count: %d systems", expectedCount)
}

// TestCompetitiveSystemsNoDuplicates verifies no duplicate systems are initialized
func TestCompetitiveSystemsNoDuplicates(t *testing.T) {
	world := engine.NewWorld()
	seed := int64(77777)
	logger := createTestLogger()

	initializeV4Systems(world, seed, "fantasy", logger, nil)

	systems := world.GetSystems()

	// Count each type of competitive system
	counts := make(map[string]int)
	for _, sys := range systems {
		switch sys.(type) {
		case *engine.RaidSystem:
			counts["raid"]++
		case *engine.TournamentSystem:
			counts["tournament"]++
		case *engine.PvPRatingSystem:
			counts["pvp_rating"]++
		case *engine.LegendaryQuestSystem:
			counts["legendary_quest"]++
		}
	}

	// Verify no duplicates
	for systemType, count := range counts {
		if count > 1 {
			t.Errorf("System type %s initialized %d times, expected 1", systemType, count)
		}
	}

	t.Logf("No duplicate competitive systems found: %v", counts)
}
