package engine

import (
	"testing"
)

// TestPlayerCombatSystem_AttackInRange tests combat when enemy is in range.
func TestPlayerCombatSystem_AttackInRange(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{
		Damage:   15,
		Range:    50,
		Cooldown: 1.0,
	})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemy in range
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 120, Y: 100}) // 20 pixels away
	enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&StatsComponent{})

	world.Update(0) // Process additions

	// Run player combat system
	playerCombatSys.Update(world.GetEntities(), 0.016)

	// Verify enemy took damage
	healthComp, _ := enemy.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current >= 50 {
		t.Errorf("Enemy should have taken damage, health still at %f", health.Current)
	}

	// Verify input was consumed
	inputComp, _ := player.GetComponent("input")
	input := inputComp.(*StubInput)
	if input.ActionPressed {
		t.Error("ActionPressed should be false after use")
	}
}

// TestPlayerCombatSystem_AttackOutOfRange tests combat when no enemy in range.
func TestPlayerCombatSystem_AttackOutOfRange(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{
		Damage:   15,
		Range:    50,
		Cooldown: 1.0,
	})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemy out of range
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 200, Y: 100}) // 100 pixels away
	enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
	enemy.AddComponent(&TeamComponent{TeamID: 2})

	world.Update(0)

	// Run player combat system
	playerCombatSys.Update(world.GetEntities(), 0.016)

	// Verify enemy took NO damage
	healthComp, _ := enemy.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 50 {
		t.Errorf("Enemy should not have taken damage (out of range), got %f", health.Current)
	}
}

// TestPlayerCombatSystem_AttackOnCooldown tests combat when attack is on cooldown.
func TestPlayerCombatSystem_AttackOnCooldown(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	attackComp := &AttackComponent{
		Damage:        15,
		Range:         50,
		Cooldown:      1.0,
		CooldownTimer: 0.5, // Still on cooldown
	}
	player.AddComponent(attackComp)
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemy in range
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 120, Y: 100})
	enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
	enemy.AddComponent(&TeamComponent{TeamID: 2})

	world.Update(0)

	// Run player combat system
	playerCombatSys.Update(world.GetEntities(), 0.016)

	// Verify enemy took NO damage (attack on cooldown)
	healthComp, _ := enemy.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 50 {
		t.Errorf("Enemy should not have taken damage (cooldown), got %f", health.Current)
	}
}

// TestPlayerCombatSystem_NoInputComponent tests system with non-player entity.
func TestPlayerCombatSystem_NoInputComponent(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create entity without input (not player-controlled)
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 1.0})

	world.Update(0)

	// Should not panic
	playerCombatSys.Update(world.GetEntities(), 0.016)
}

// TestPlayerCombatSystem_NoAttackComponent tests player without attack ability.
func TestPlayerCombatSystem_NoAttackComponent(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player without attack component
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})

	world.Update(0)

	// Should not panic
	playerCombatSys.Update(world.GetEntities(), 0.016)
}

// TestPlayerCombatSystem_MultipleEnemies tests targeting nearest enemy.
func TestPlayerCombatSystem_MultipleEnemies(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{Damage: 15, Range: 100, Cooldown: 1.0})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create multiple enemies at different distances
	far := world.CreateEntity()
	far.AddComponent(&PositionComponent{X: 180, Y: 100}) // 80 pixels
	far.AddComponent(&HealthComponent{Current: 50, Max: 50})
	far.AddComponent(&TeamComponent{TeamID: 2})
	far.AddComponent(&StatsComponent{})

	near := world.CreateEntity()
	near.AddComponent(&PositionComponent{X: 130, Y: 100}) // 30 pixels
	near.AddComponent(&HealthComponent{Current: 50, Max: 50})
	near.AddComponent(&TeamComponent{TeamID: 2})
	near.AddComponent(&StatsComponent{})

	world.Update(0)

	// Run player combat system
	playerCombatSys.Update(world.GetEntities(), 0.016)

	// Nearest enemy should be damaged
	nearHealthComp, _ := near.GetComponent("health")
	nearHealth := nearHealthComp.(*HealthComponent)
	if nearHealth.Current >= 50 {
		t.Error("Nearest enemy should have taken damage")
	}

	// Farther enemy should be untouched (only one target per attack)
	farHealthComp, _ := far.GetComponent("health")
	farHealth := farHealthComp.(*HealthComponent)
	if farHealth.Current != 50 {
		t.Error("Farther enemy should not have taken damage")
	}
}

// TestPlayerCombatSystem_NoEnemies tests combat with no valid targets.
func TestPlayerCombatSystem_NoEnemies(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player only
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 1.0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	world.Update(0)

	// Should not panic
	playerCombatSys.Update(world.GetEntities(), 0.016)

	// Input should be consumed anyway
	inputComp, _ := player.GetComponent("input")
	input := inputComp.(*StubInput)
	if input.ActionPressed {
		t.Error("ActionPressed should be consumed even if no target")
	}
}

// TestPlayerCombatSystem_DeadEnemy tests that dead enemies are not targeted.
func TestPlayerCombatSystem_DeadEnemy(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 1.0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create dead enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 120, Y: 100})
	enemy.AddComponent(&HealthComponent{Current: 0, Max: 50}) // Dead
	enemy.AddComponent(&TeamComponent{TeamID: 2})

	world.Update(0)

	// Should not panic, dead enemy should not be targeted
	playerCombatSys.Update(world.GetEntities(), 0.016)
}

// Benchmark for player combat system performance
func BenchmarkPlayerCombatSystem(b *testing.B) {
	world := NewWorld()
	combatSys := NewCombatSystem(12345)
	playerCombatSys := NewPlayerCombatSystem(combatSys, world)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&StubInput{ActionPressed: true})
	player.AddComponent(&AttackComponent{Damage: 15, Range: 50, Cooldown: 0.0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create 10 enemies
	for i := 0; i < 10; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: float64(110 + i*5), Y: 100})
		enemy.AddComponent(&HealthComponent{Current: 50, Max: 50})
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&StatsComponent{})
	}

	world.Update(0)
	entities := world.GetEntities()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerCombatSys.Update(entities, 0.016)
	}
}
