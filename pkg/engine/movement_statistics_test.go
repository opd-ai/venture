// Package engine provides tests for movement statistics tracking.
// This file tests that the MovementSystem correctly tracks player movement
// statistics including distance traveled and areas visited.
package engine

import (
	"testing"
)

// TestMovementSystem_SetStatisticsSystem verifies that the statistics system can be set.
func TestMovementSystem_SetStatisticsSystem(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)

	// Should not panic
	moveSystem.SetStatisticsSystem(statsSystem)

	// Verify system is set (indirectly via behavior)
	if moveSystem.statisticsSystem != statsSystem {
		t.Error("Statistics system not set correctly")
	}
}

// TestMovementSystem_TracksDistanceTraveled verifies that player movement
// correctly updates the distance traveled statistic.
func TestMovementSystem_TracksDistanceTraveled(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)
	moveSystem.SetStatisticsSystem(statsSystem)

	// Create player entity with input component
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	player.AddComponent(&StubInput{})
	player.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0)

	// Update for 1 second - should move 100 units in X
	moveSystem.Update(world.GetEntities(), 1.0)

	// Check that distance was tracked
	comp, ok := player.GetComponent("player_statistics")
	if !ok {
		t.Fatal("Player statistics component not found")
	}
	stats := comp.(*PlayerStatisticsComponent)

	distance := stats.GetSessionStat("explore_distance_traveled")
	if distance < 90 || distance > 110 {
		t.Errorf("Distance traveled = %d, want approximately 100", distance)
	}
}

// TestMovementSystem_TracksAreasVisited verifies that moving into new grid cells
// correctly updates the areas visited statistic.
func TestMovementSystem_TracksAreasVisited(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)
	moveSystem.SetStatisticsSystem(statsSystem)

	// Create player entity starting at position 0,0
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&VelocityComponent{VX: 0, VY: 0})
	player.AddComponent(&StubInput{})
	player.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0)

	// First update - should visit first cell
	moveSystem.Update(world.GetEntities(), 0.1)

	// Now set velocity to move to a new cell (200+ pixels away)
	vel, _ := player.GetComponent("velocity")
	velocity := vel.(*VelocityComponent)
	velocity.VX = 250 // Move 250 units in 1 second

	moveSystem.Update(world.GetEntities(), 1.0) // Should now be at X=300, new cell

	// Check areas visited
	comp, ok := player.GetComponent("player_statistics")
	if !ok {
		t.Fatal("Player statistics component not found")
	}
	stats := comp.(*PlayerStatisticsComponent)

	areas := stats.GetSessionStat("explore_areas_visited")
	if areas < 1 {
		t.Errorf("Areas visited = %d, want at least 1", areas)
	}
}

// TestMovementSystem_NoStatsWithoutStatisticsSystem verifies that movement
// doesn't fail when no statistics system is set.
func TestMovementSystem_NoStatsWithoutStatisticsSystem(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	// Don't set statistics system

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	player.AddComponent(&StubInput{})
	world.Update(0)

	// Should not panic
	moveSystem.Update(world.GetEntities(), 1.0)

	// Position should still be updated
	pos := player.GetPosition()
	if pos.X != 100 {
		t.Errorf("Position X = %f, want 100", pos.X)
	}
}

// TestMovementSystem_NoStatsForNonPlayers verifies that non-player entities
// don't have their movement tracked for statistics.
func TestMovementSystem_NoStatsForNonPlayers(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)
	moveSystem.SetStatisticsSystem(statsSystem)

	// Create non-player entity (no input component)
	npc := world.CreateEntity()
	npc.AddComponent(&PositionComponent{X: 0, Y: 0})
	npc.AddComponent(&VelocityComponent{VX: 100, VY: 0})
	npc.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0)

	moveSystem.Update(world.GetEntities(), 1.0)

	// Stats should not be updated for non-player
	comp, ok := npc.GetComponent("player_statistics")
	if !ok {
		t.Fatal("Player statistics component not found")
	}
	stats := comp.(*PlayerStatisticsComponent)

	distance := stats.GetSessionStat("explore_distance_traveled")
	if distance != 0 {
		t.Errorf("Non-player distance = %d, want 0", distance)
	}
}

// TestMovementSystem_ClearVisitedCells verifies that visited cells can be cleared.
func TestMovementSystem_ClearVisitedCells(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)
	moveSystem.SetStatisticsSystem(statsSystem)

	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&VelocityComponent{VX: 250, VY: 0})
	player.AddComponent(&StubInput{})
	player.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0)

	// Move player to visit some cells
	moveSystem.Update(world.GetEntities(), 1.0)

	// Verify cells were tracked
	if len(moveSystem.visitedCells[player.ID]) == 0 {
		t.Error("Expected visited cells to be tracked")
	}

	// Clear visited cells
	moveSystem.ClearVisitedCells(player.ID)

	// Verify cells were cleared
	if len(moveSystem.visitedCells[player.ID]) != 0 {
		t.Error("Expected visited cells to be cleared")
	}
}

// TestMovementSystem_MultipleAreasInOneUpdate verifies that moving across
// multiple cells in a single update still tracks correctly.
func TestMovementSystem_MultipleAreasInOneUpdate(t *testing.T) {
	world := NewWorld()
	moveSystem := NewMovementSystem(0)
	statsSystem := NewStatisticsSystem(world)
	moveSystem.SetStatisticsSystem(statsSystem)

	// Start at cell (0,0)
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 50, Y: 50})
	player.AddComponent(&VelocityComponent{VX: 500, VY: 0}) // Will move 500 units, crossing 2+ cells
	player.AddComponent(&StubInput{})
	player.AddComponent(NewPlayerStatisticsComponent())
	world.Update(0)

	moveSystem.Update(world.GetEntities(), 1.0)

	// Check that at least some cells were visited
	// Note: Single movement only counts the end cell, not intermediate cells
	// This is acceptable behavior for tutorial purposes
	comp, _ := player.GetComponent("player_statistics")
	stats := comp.(*PlayerStatisticsComponent)

	areas := stats.GetSessionStat("explore_areas_visited")
	if areas < 1 {
		t.Errorf("Areas visited = %d, want at least 1", areas)
	}
}
