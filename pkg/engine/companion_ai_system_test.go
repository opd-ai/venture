// Package engine provides tests for companion AI behavior.
package engine

import (
	"testing"
)

// TestNewCompanionAISystem tests the companion AI system constructor.
func TestNewCompanionAISystem(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	if system == nil {
		t.Fatal("NewCompanionAISystem returned nil")
	}

	if system.world != world {
		t.Error("Expected system to have correct world reference")
	}
}

// TestCompanionAISystemUpdate tests the Update function with different behaviors.
func TestCompanionAISystemUpdate(t *testing.T) {
	tests := []struct {
		name     string
		behavior BehaviorMode
	}{
		{"aggressive", BehaviorAggressive},
		{"defensive", BehaviorDefensive},
		{"passive", BehaviorPassive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewCompanionAISystem(world)

			// Create owner
			owner := NewEntity(1)
			owner.AddComponent(&PositionComponent{X: 100, Y: 100})
			world.AddEntity(owner)

			// Create companion
			companion := NewEntity(2)
			companion.AddComponent(&PositionComponent{X: 200, Y: 200})
			companion.AddComponent(&CompanionComponent{
				OwnerID:  owner.ID,
				Behavior: tt.behavior,
			})
			world.AddEntity(companion)

			// Update the system - should not crash
			system.Update(0.016)
		})
	}
}

// TestCompanionAIFollowOwner tests that companions follow their owner.
func TestCompanionAIFollowOwner(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create owner at position (100, 100)
	owner := NewEntity(1)
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(owner)

	// Create companion far from owner at (300, 300)
	companion := NewEntity(2)
	companionPos := &PositionComponent{X: 300, Y: 300}
	companion.AddComponent(companionPos)
	companion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorPassive, // Passive only follows
	})
	world.AddEntity(companion)

	// Record initial position
	initialX := companionPos.X
	initialY := companionPos.Y

	// Update multiple times
	for i := 0; i < 10; i++ {
		system.Update(0.1)
	}

	// Companion should have moved toward owner
	if companionPos.X >= initialX && companionPos.Y >= initialY {
		t.Errorf("Companion should have moved toward owner: initial=(%.1f,%.1f), current=(%.1f,%.1f)",
			initialX, initialY, companionPos.X, companionPos.Y)
	}
}

// TestCompanionAIStaysCloseToOwner tests that companions don't move when already close.
func TestCompanionAIStaysCloseToOwner(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create owner
	owner := NewEntity(1)
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(owner)

	// Create companion very close to owner (within 30 units)
	companion := NewEntity(2)
	companionPos := &PositionComponent{X: 110, Y: 110}
	companion.AddComponent(companionPos)
	companion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorPassive,
	})
	world.AddEntity(companion)

	// Record initial position
	initialX := companionPos.X
	initialY := companionPos.Y

	// Update
	system.Update(0.016)

	// Position should be unchanged (distance < 30)
	if companionPos.X != initialX || companionPos.Y != initialY {
		t.Errorf("Close companion should not move: initial=(%.1f,%.1f), current=(%.1f,%.1f)",
			initialX, initialY, companionPos.X, companionPos.Y)
	}
}

// TestCompanionAIMissingOwner tests behavior when owner doesn't exist.
func TestCompanionAIMissingOwner(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create companion with non-existent owner
	companion := NewEntity(2)
	companionPos := &PositionComponent{X: 200, Y: 200}
	companion.AddComponent(companionPos)
	companion.AddComponent(&CompanionComponent{
		OwnerID:  999, // Non-existent owner
		Behavior: BehaviorPassive,
	})
	world.AddEntity(companion)

	initialX := companionPos.X
	initialY := companionPos.Y

	// Update should not crash and position should be unchanged
	system.Update(0.016)

	if companionPos.X != initialX || companionPos.Y != initialY {
		t.Error("Companion with missing owner should not move")
	}
}

// TestCompanionAIAggressiveFindEnemies tests aggressive companions finding enemies.
func TestCompanionAIAggressiveFindEnemies(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create owner
	owner := NewEntity(1)
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(owner)

	// Create aggressive companion
	companion := NewEntity(2)
	companionPos := &PositionComponent{X: 120, Y: 120}
	companion.AddComponent(companionPos)
	companion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorAggressive,
	})
	companion.AddComponent(&AIComponent{}) // For target assignment
	world.AddEntity(companion)

	// Create enemy within range (200 units)
	enemy := NewEntity(3)
	enemy.AddComponent(&PositionComponent{X: 150, Y: 150})
	enemy.AddComponent(&AIComponent{}) // Has AI component to be detected as enemy
	world.AddEntity(enemy)

	// Update
	system.Update(0.016)

	// Check if target was set
	aiComp, ok := companion.GetComponent("ai")
	if !ok {
		t.Fatal("Companion should have AI component")
	}
	ai := aiComp.(*AIComponent)

	if ai.Target == nil {
		t.Error("Aggressive companion should have found a target")
	}
}

// TestCompanionAIDefensiveStaysNearOwner tests defensive behavior.
func TestCompanionAIDefensiveStaysNearOwner(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create owner
	owner := NewEntity(1)
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(owner)

	// Create defensive companion at moderate distance
	companion := NewEntity(2)
	companionPos := &PositionComponent{X: 180, Y: 180} // About 113 units away
	companion.AddComponent(companionPos)
	companion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorDefensive,
	})
	world.AddEntity(companion)

	// Record initial position
	initialX := companionPos.X
	initialY := companionPos.Y

	// Update
	system.Update(0.1)

	// Defensive companion should move closer (distance > 50)
	if companionPos.X >= initialX || companionPos.Y >= initialY {
		t.Error("Defensive companion should move toward owner when too far")
	}
}

// TestCompanionAIWithStats tests that companion uses its speed stat.
func TestCompanionAIWithStats(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create owner
	owner := NewEntity(1)
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	world.AddEntity(owner)

	// Create fast companion
	fastCompanion := NewEntity(2)
	fastPos := &PositionComponent{X: 300, Y: 300}
	fastCompanion.AddComponent(fastPos)
	fastCompanion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorPassive,
	})
	fastCompanion.AddComponent(&CompanionStatsComponent{Speed: 200})
	world.AddEntity(fastCompanion)

	// Create slow companion
	slowCompanion := NewEntity(3)
	slowPos := &PositionComponent{X: 300, Y: 300}
	slowCompanion.AddComponent(slowPos)
	slowCompanion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorPassive,
	})
	slowCompanion.AddComponent(&CompanionStatsComponent{Speed: 50})
	world.AddEntity(slowCompanion)

	// Update once
	system.Update(0.1)

	// Fast companion should have moved more
	fastDistX := 300 - fastPos.X
	slowDistX := 300 - slowPos.X

	if fastDistX <= slowDistX {
		t.Errorf("Fast companion should move farther: fast=%.1f, slow=%.1f", fastDistX, slowDistX)
	}
}

// TestCompanionAIDistance tests the distance calculation helper.
func TestCompanionAIDistance(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	tests := []struct {
		name     string
		pos1     *PositionComponent
		pos2     *PositionComponent
		expected float64
	}{
		{
			name:     "same position",
			pos1:     &PositionComponent{X: 100, Y: 100},
			pos2:     &PositionComponent{X: 100, Y: 100},
			expected: 0,
		},
		{
			name:     "horizontal distance",
			pos1:     &PositionComponent{X: 0, Y: 0},
			pos2:     &PositionComponent{X: 100, Y: 0},
			expected: 100,
		},
		{
			name:     "vertical distance",
			pos1:     &PositionComponent{X: 0, Y: 0},
			pos2:     &PositionComponent{X: 0, Y: 100},
			expected: 100,
		},
		{
			name:     "diagonal 3-4-5 triangle",
			pos1:     &PositionComponent{X: 0, Y: 0},
			pos2:     &PositionComponent{X: 30, Y: 40},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := system.distance(tt.pos1, tt.pos2)
			if dist != tt.expected {
				t.Errorf("distance() = %v, want %v", dist, tt.expected)
			}
		})
	}
}

// TestCompanionAIFindNearbyEnemies tests enemy detection.
func TestCompanionAIFindNearbyEnemies(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create companion position
	companionPos := &PositionComponent{X: 100, Y: 100}

	// Create enemies at various distances
	// Close enemy (within 200 radius)
	closeEnemy := NewEntity(1)
	closeEnemy.AddComponent(&PositionComponent{X: 150, Y: 150})
	closeEnemy.AddComponent(&AIComponent{})
	world.AddEntity(closeEnemy)

	// Far enemy (outside 200 radius)
	farEnemy := NewEntity(2)
	farEnemy.AddComponent(&PositionComponent{X: 500, Y: 500})
	farEnemy.AddComponent(&AIComponent{})
	world.AddEntity(farEnemy)

	// Non-enemy (no AI component)
	nonEnemy := NewEntity(3)
	nonEnemy.AddComponent(&PositionComponent{X: 110, Y: 110})
	world.AddEntity(nonEnemy)

	enemies := system.findNearbyEnemies(companionPos, 200.0)

	// Should only find the close enemy
	if len(enemies) != 1 {
		t.Errorf("Expected 1 nearby enemy, got %d", len(enemies))
	}

	if len(enemies) > 0 && enemies[0].ID != closeEnemy.ID {
		t.Error("Wrong enemy detected")
	}
}

// TestCompanionAIMissingComponents tests behavior with missing components.
func TestCompanionAIMissingComponents(t *testing.T) {
	world := NewWorld()
	system := NewCompanionAISystem(world)

	// Create companion without position
	companionNoPos := NewEntity(1)
	companionNoPos.AddComponent(&CompanionComponent{
		OwnerID:  999,
		Behavior: BehaviorPassive,
	})
	world.AddEntity(companionNoPos)

	// Create owner without position
	owner := NewEntity(2)
	world.AddEntity(owner)

	companion := NewEntity(3)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&CompanionComponent{
		OwnerID:  owner.ID,
		Behavior: BehaviorPassive,
	})
	world.AddEntity(companion)

	// Update should not panic
	system.Update(0.016)
}
