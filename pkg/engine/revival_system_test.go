package engine

import (
	"testing"
)

// TestRevivalSystemBasic tests basic revival functionality
// Priority 1.5: Multiplayer Revival System
func TestRevivalSystemBasic(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: true} // E key pressed
	livingPlayer.AddComponent(livingInput)

	// Create dead player nearby
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 110, Y: 100}) // 10 pixels away
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{}) // Player input component
	deadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Trigger revival
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify dead player is revived
	if deadPlayer.HasComponent("dead") {
		t.Error("dead player should no longer have dead component after revival")
	}

	// Verify health restored
	healthComp, _ := deadPlayer.GetComponent("health")
	health := healthComp.(*HealthComponent)
	expectedHealth := 100.0 * 0.2 // 20% of max
	if health.Current != expectedHealth {
		t.Errorf("revived player health = %f, want %f", health.Current, expectedHealth)
	}
}

// TestRevivalSystemOutOfRange tests that revival doesn't work beyond range
func TestRevivalSystemOutOfRange(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: true}
	livingPlayer.AddComponent(livingInput)

	// Create dead player far away (out of range)
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 200, Y: 200}) // ~141 pixels away (sqrt(10000+10000))
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{})
	deadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Try to revive (should fail due to range)
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify dead player is still dead
	if !deadPlayer.HasComponent("dead") {
		t.Error("dead player out of range should not be revived")
	}

	// Health should still be 0
	healthComp, _ := deadPlayer.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 0 {
		t.Errorf("dead player out of range should have 0 health, got %f", health.Current)
	}
}

// TestRevivalSystemNoInput tests that revival doesn't happen without input
func TestRevivalSystemNoInput(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create living player WITHOUT pressing E key
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: false} // E key NOT pressed
	livingPlayer.AddComponent(livingInput)

	// Create dead player nearby
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 110, Y: 100})
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{})
	deadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Try to revive (should fail - no input)
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify dead player is still dead
	if !deadPlayer.HasComponent("dead") {
		t.Error("dead player should not be revived without input")
	}
}

// TestRevivalSystemMultipleDeadPlayers tests reviving closest player
func TestRevivalSystemMultipleDeadPlayers(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: true}
	livingPlayer.AddComponent(livingInput)

	// Create two dead players at different distances
	closerDeadPlayer := world.CreateEntity()
	closerDeadPlayer.AddComponent(&PositionComponent{X: 110, Y: 100}) // 10 pixels away
	closerDeadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	closerDeadPlayer.AddComponent(&EbitenInput{})
	closerDeadPlayer.AddComponent(NewDeadComponent(0.0))

	fartherDeadPlayer := world.CreateEntity()
	fartherDeadPlayer.AddComponent(&PositionComponent{X: 120, Y: 100}) // 20 pixels away
	fartherDeadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	fartherDeadPlayer.AddComponent(&EbitenInput{})
	fartherDeadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Trigger revival - should revive closer player only
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify closer player is revived
	if closerDeadPlayer.HasComponent("dead") {
		t.Error("closer dead player should be revived")
	}

	// Verify farther player is still dead
	if !fartherDeadPlayer.HasComponent("dead") {
		t.Error("farther dead player should still be dead")
	}
}

// TestRevivalSystemCustomParameters tests custom revival range and amount
func TestRevivalSystemCustomParameters(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Set custom parameters
	revivalSystem.RevivalRange = 64.0 // Double range
	revivalSystem.RevivalAmount = 0.5 // 50% health instead of 20%

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: true}
	livingPlayer.AddComponent(livingInput)

	// Create dead player at 50 pixels away (within custom range)
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 150, Y: 100}) // 50 pixels
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{})
	deadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Trigger revival
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify revival worked with custom range
	if deadPlayer.HasComponent("dead") {
		t.Error("dead player within custom range should be revived")
	}

	// Verify custom health amount
	healthComp, _ := deadPlayer.GetComponent("health")
	health := healthComp.(*HealthComponent)
	expectedHealth := 100.0 * 0.5 // 50%
	if health.Current != expectedHealth {
		t.Errorf("revived player health = %f, want %f (50%%)", health.Current, expectedHealth)
	}
}

// TestRevivalSystemDeadLivingPlayerCannotRevive tests dead players can't revive others
func TestRevivalSystemDeadLivingPlayerCannotRevive(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create "living" player who is actually dead (has dead component)
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100}) // Dead
	livingInput := &EbitenInput{UseItemPressed: true}
	livingPlayer.AddComponent(livingInput)
	livingPlayer.AddComponent(NewDeadComponent(0.0)) // Dead!

	// Create dead player nearby
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 110, Y: 100})
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{})
	deadPlayer.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Try to revive (should fail - reviver is dead)
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify dead player is still dead (not revived by another dead player)
	if !deadPlayer.HasComponent("dead") {
		t.Error("dead player should not be revived by another dead player")
	}
}

// TestIsPlayerRevivable tests the revivability checker function
func TestIsPlayerRevivable(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Entity
		revivable bool
	}{
		{
			name: "revivable dead player",
			setup: func() *Entity {
				e := &Entity{ID: 1, Components: make(map[string]Component)}
				e.AddComponent(&EbitenInput{})
				e.AddComponent(NewDeadComponent(0.0))
				e.AddComponent(&HealthComponent{Current: 0, Max: 100})
				return e
			},
			revivable: true,
		},
		{
			name: "living player not revivable",
			setup: func() *Entity {
				e := &Entity{ID: 2, Components: make(map[string]Component)}
				e.AddComponent(&EbitenInput{})
				e.AddComponent(&HealthComponent{Current: 50, Max: 100})
				// No dead component
				return e
			},
			revivable: false,
		},
		{
			name: "dead NPC not revivable",
			setup: func() *Entity {
				e := &Entity{ID: 3, Components: make(map[string]Component)}
				// No input component (not a player)
				e.AddComponent(NewDeadComponent(0.0))
				e.AddComponent(&HealthComponent{Current: 0, Max: 100})
				return e
			},
			revivable: false,
		},
		{
			name: "dead player without health not revivable",
			setup: func() *Entity {
				e := &Entity{ID: 4, Components: make(map[string]Component)}
				e.AddComponent(&EbitenInput{})
				e.AddComponent(NewDeadComponent(0.0))
				// No health component
				return e
			},
			revivable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := tt.setup()
			result := IsPlayerRevivable(entity)
			if result != tt.revivable {
				t.Errorf("IsPlayerRevivable() = %v, want %v", result, tt.revivable)
			}
		})
	}
}

// TestFindRevivablePlayersInRange tests finding revivable players
func TestFindRevivablePlayersInRange(t *testing.T) {
	world := NewWorld()

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingPlayer.AddComponent(&EbitenInput{})

	// Create dead player in range
	deadPlayerInRange := world.CreateEntity()
	deadPlayerInRange.AddComponent(&PositionComponent{X: 120, Y: 100}) // 20 pixels
	deadPlayerInRange.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayerInRange.AddComponent(&EbitenInput{})
	deadPlayerInRange.AddComponent(NewDeadComponent(0.0))

	// Create dead player out of range
	deadPlayerOutOfRange := world.CreateEntity()
	deadPlayerOutOfRange.AddComponent(&PositionComponent{X: 200, Y: 100}) // 100 pixels
	deadPlayerOutOfRange.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayerInRange.AddComponent(&EbitenInput{})
	deadPlayerOutOfRange.AddComponent(NewDeadComponent(0.0))

	// Create dead NPC (not revivable - no input component)
	deadNPC := world.CreateEntity()
	deadNPC.AddComponent(&PositionComponent{X: 110, Y: 100})
	deadNPC.AddComponent(&HealthComponent{Current: 0, Max: 100})
	// No input component
	deadNPC.AddComponent(NewDeadComponent(0.0))

	world.Update(0)

	// Find revivable players within 50 pixel range
	revivablePlayers := FindRevivablePlayersInRange(world, livingPlayer, 50.0)

	// Should find only the one dead player in range
	if len(revivablePlayers) != 1 {
		t.Errorf("expected 1 revivable player in range, got %d", len(revivablePlayers))
	}

	if len(revivablePlayers) > 0 && revivablePlayers[0].ID != deadPlayerInRange.ID {
		t.Error("found player should be the one in range")
	}
}

// TestRevivalSystemNoPlayers tests system handles empty world gracefully
func TestRevivalSystemNoPlayers(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Empty world - should not crash
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Test passed if no panic
}

// TestRevivalSystemNonPlayerEntities tests system ignores non-player entities
func TestRevivalSystemNonPlayerEntities(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create NPC enemies (no input component)
	enemy1 := world.CreateEntity()
	enemy1.AddComponent(&PositionComponent{X: 100, Y: 100})
	enemy1.AddComponent(&HealthComponent{Current: 100, Max: 100})
	// No input component

	enemy2 := world.CreateEntity()
	enemy2.AddComponent(&PositionComponent{X: 110, Y: 100})
	enemy2.AddComponent(&HealthComponent{Current: 0, Max: 100})
	enemy2.AddComponent(NewDeadComponent(0.0))
	// No input component

	world.Update(0)

	// Should not crash and not revive NPCs
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Enemy should still be dead
	if !enemy2.HasComponent("dead") {
		t.Error("dead NPC should not be revived by revival system")
	}
}

// TestRevivalWithDroppedItems tests that revived players don't lose dropped items
func TestRevivalWithDroppedItems(t *testing.T) {
	world := NewWorld()
	revivalSystem := NewRevivalSystem(world)

	// Create living player
	livingPlayer := world.CreateEntity()
	livingPlayer.AddComponent(&PositionComponent{X: 100, Y: 100})
	livingPlayer.AddComponent(&HealthComponent{Current: 100, Max: 100})
	livingInput := &EbitenInput{UseItemPressed: true}
	livingPlayer.AddComponent(livingInput)

	// Create dead player with dropped items tracked
	deadPlayer := world.CreateEntity()
	deadPlayer.AddComponent(&PositionComponent{X: 110, Y: 100})
	deadPlayer.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadPlayer.AddComponent(&EbitenInput{})

	deadComp := NewDeadComponent(0.0)
	deadComp.AddDroppedItem(1001) // Simulate dropped item entities
	deadComp.AddDroppedItem(1002)
	deadPlayer.AddComponent(deadComp)

	world.Update(0)

	// Trigger revival
	revivalSystem.Update(world.GetEntities(), 0.016)

	// Verify player is revived
	if deadPlayer.HasComponent("dead") {
		t.Fatal("player should be revived")
	}

	// Note: Dropped items remain in world - players need to pick them back up
	// This is intentional game design - revival doesn't auto-restore inventory
}
