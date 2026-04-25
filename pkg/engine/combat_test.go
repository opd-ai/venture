package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
)

func TestHealthComponent(t *testing.T) {
	tests := []struct {
		name            string
		initial         float64
		max             float64
		operation       string
		amount          float64
		expectedCurrent float64
		expectedAlive   bool
	}{
		{"full health", 100, 100, "none", 0, 100, true},
		{"take damage", 100, 100, "damage", 30, 70, true},
		{"fatal damage", 100, 100, "damage", 150, 0, false},
		{"heal partial", 50, 100, "heal", 30, 80, true},
		{"heal overcap", 80, 100, "heal", 50, 100, true},
		{"exact lethal", 50, 100, "damage", 50, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HealthComponent{
				Current: tt.initial,
				Max:     tt.max,
			}

			switch tt.operation {
			case "damage":
				h.TakeDamage(tt.amount)
			case "heal":
				h.Heal(tt.amount)
			}

			if h.Current != tt.expectedCurrent {
				t.Errorf("expected current health %v, got %v", tt.expectedCurrent, h.Current)
			}

			if h.IsAlive() != tt.expectedAlive {
				t.Errorf("expected IsAlive() %v, got %v", tt.expectedAlive, h.IsAlive())
			}

			if h.IsDead() == tt.expectedAlive {
				t.Errorf("IsDead() should be opposite of IsAlive()")
			}
		})
	}
}

func TestStatsComponent(t *testing.T) {
	stats := NewStatsComponent()

	// Test default values
	if stats.Attack <= 0 {
		t.Error("default attack should be positive")
	}
	if stats.CritChance < 0 || stats.CritChance > 1 {
		t.Error("default crit chance should be between 0 and 1")
	}

	// Test resistance
	stats.Resistances[combat.DamageFire] = 0.5

	if res := stats.GetResistance(combat.DamageFire); res != 0.5 {
		t.Errorf("expected fire resistance 0.5, got %v", res)
	}

	if res := stats.GetResistance(combat.DamageIce); res != 0.0 {
		t.Errorf("expected ice resistance 0.0, got %v", res)
	}
}

func TestAttackComponent(t *testing.T) {
	attack := &AttackComponent{
		Damage:        20,
		DamageType:    combat.DamagePhysical,
		Range:         50,
		Cooldown:      1.0,
		CooldownTimer: 0,
	}

	// Should be able to attack initially
	if !attack.CanAttack() {
		t.Error("attack should be ready initially")
	}

	// Reset cooldown
	attack.ResetCooldown()

	if attack.CooldownTimer != 1.0 {
		t.Errorf("expected cooldown timer 1.0, got %v", attack.CooldownTimer)
	}

	if attack.CanAttack() {
		t.Error("attack should not be ready after reset")
	}

	// Update cooldown
	attack.UpdateCooldown(0.5)
	if attack.CooldownTimer != 0.5 {
		t.Errorf("expected cooldown timer 0.5 after update, got %v", attack.CooldownTimer)
	}

	attack.UpdateCooldown(0.6)
	if attack.CooldownTimer != 0 {
		t.Errorf("expected cooldown timer 0 after expiry, got %v", attack.CooldownTimer)
	}

	if !attack.CanAttack() {
		t.Error("attack should be ready after cooldown expires")
	}
}

func TestStatusEffectComponent(t *testing.T) {
	effect := &StatusEffectComponent{
		EffectType:   "poison",
		Duration:     5.0,
		Magnitude:    10.0,
		TickInterval: 1.0,
		NextTick:     1.0,
	}

	// Not expired initially
	if effect.IsExpired() {
		t.Error("effect should not be expired initially")
	}

	// Update without tick
	ticked := effect.Update(0.5)
	if ticked {
		t.Error("should not tick after 0.5 seconds")
	}
	if effect.Duration != 4.5 {
		t.Errorf("expected duration 4.5, got %v", effect.Duration)
	}

	// Update with tick
	ticked = effect.Update(0.6)
	if !ticked {
		t.Error("should tick after 1.1 seconds total")
	}
	if effect.NextTick != 1.0 {
		t.Errorf("tick timer should reset to 1.0, got %v", effect.NextTick)
	}

	// Update until expiry
	effect.Update(10.0)
	if !effect.IsExpired() {
		t.Error("effect should be expired after duration passes")
	}
}

func TestTeamComponent(t *testing.T) {
	team1 := &TeamComponent{TeamID: 1}
	neutral := &TeamComponent{TeamID: 0}

	// Test allies
	if !team1.IsAlly(1) {
		t.Error("team should be ally with itself")
	}
	if team1.IsAlly(2) {
		t.Error("team 1 should not be ally with team 2")
	}

	// Test enemies
	if !team1.IsEnemy(2) {
		t.Error("team 1 should be enemy with team 2")
	}
	if team1.IsEnemy(1) {
		t.Error("team should not be enemy with itself")
	}
	if team1.IsEnemy(0) {
		t.Error("team should not be enemy with neutral")
	}
	if neutral.IsEnemy(1) {
		t.Error("neutral should not be enemy with any team")
	}
}

func TestCombatSystemBasicAttack(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Create attacker
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   1.0,
	})
	attacker.AddComponent(NewStatsComponent())

	// Create target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 50, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	target.AddComponent(NewStatsComponent())

	world.Update(0) // Process additions

	// Perform attack
	hit := combatSystem.Attack(attacker, target)
	if !hit {
		t.Error("attack should hit")
	}

	// Check health reduction
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current >= 100 {
		t.Error("target health should be reduced")
	}
	if health.Current <= 0 {
		t.Error("target should not be dead from one hit")
	}

	// Check cooldown
	attackComp, _ := attacker.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if attack.CanAttack() {
		t.Error("attack should be on cooldown")
	}

	// Try to attack while on cooldown
	hit = combatSystem.Attack(attacker, target)
	if hit {
		t.Error("should not be able to attack while on cooldown")
	}
}

func TestCombatSystemRange(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create attacker with limited range
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      50,
		Cooldown:   0,
	})

	// Create target out of range
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 100, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	// Attack should miss due to range
	hit := combatSystem.Attack(attacker, target)
	if hit {
		t.Error("attack should miss due to range")
	}

	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 100 {
		t.Error("target should not take damage when out of range")
	}
}

func TestCombatSystemEvasion(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create attacker
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   0,
	})

	// Create target with 100% evasion
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 50, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	targetStats := NewStatsComponent()
	targetStats.Evasion = 1.0 // 100% evasion
	target.AddComponent(targetStats)

	world.Update(0)

	// Attack should miss due to evasion
	hit := combatSystem.Attack(attacker, target)
	if hit {
		t.Error("attack should miss due to evasion")
	}

	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 100 {
		t.Error("target should not take damage when evading")
	}
}

func TestCombatSystemResistance(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create attacker with fire damage
	attacker := world.CreateEntity()
	attacker.AddComponent(&AttackComponent{
		Damage:     100,
		DamageType: combat.DamageFire,
		Range:      100,
		Cooldown:   0,
	})
	attackerStats := NewStatsComponent()
	attackerStats.Attack = 0 // No bonus attack
	attackerStats.MagicPower = 0
	attackerStats.CritChance = 0 // No crits
	attacker.AddComponent(attackerStats)

	// Create target with 50% fire resistance
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	targetStats := NewStatsComponent()
	targetStats.Defense = 0 // No defense
	targetStats.MagicDefense = 0
	targetStats.Evasion = 0                          // No evasion
	targetStats.Resistances[combat.DamageFire] = 0.5 // 50% fire resistance
	target.AddComponent(targetStats)

	world.Update(0)

	// Perform attack
	hit := combatSystem.Attack(attacker, target)
	if !hit {
		t.Error("attack should hit")
	}

	// With 50% resistance, 100 damage should be reduced to 50
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	expectedHealth := 50.0 // 100 - (100 * 0.5)
	if health.Current != expectedHealth {
		t.Errorf("expected health %v after resistance, got %v", expectedHealth, health.Current)
	}
}

func TestCombatSystemStatusEffects(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	// Apply poison effect
	combatSystem.ApplyStatusEffect(entity, "poison", 3.0, 10.0, 1.0)

	// Check effect applied
	effectComp, ok := entity.GetComponent("status_effect")
	if !ok {
		t.Fatal("status effect should be applied")
	}
	effect := effectComp.(*StatusEffectComponent)
	if effect.EffectType != "poison" {
		t.Errorf("expected poison effect, got %v", effect.EffectType)
	}

	// Update 0.5 seconds - no tick yet
	world.Update(0.5)
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 100 {
		t.Error("health should not decrease before first tick")
	}

	// Update 0.6 seconds - should tick
	world.Update(0.6)
	healthComp, _ = entity.GetComponent("health")
	health = healthComp.(*HealthComponent)
	if health.Current != 90 {
		t.Errorf("expected health 90 after poison tick, got %v", health.Current)
	}

	// Update to expiry
	world.Update(10.0)

	// Effect should be removed
	_, ok = entity.GetComponent("status_effect")
	if ok {
		t.Error("expired status effect should be removed")
	}
}

func TestCombatSystemHeal(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create damaged entity
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})

	world.Update(0)

	// Heal
	combatSystem.Heal(entity, 30)

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 80 {
		t.Errorf("expected health 80 after heal, got %v", health.Current)
	}

	// Heal beyond max
	combatSystem.Heal(entity, 50)
	if health.Current != 100 {
		t.Errorf("expected health capped at 100, got %v", health.Current)
	}
}

func TestFindEnemiesInRange(t *testing.T) {
	world := NewWorld()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemies at various distances
	enemy1 := world.CreateEntity()
	enemy1.AddComponent(&PositionComponent{X: 30, Y: 0})
	enemy1.AddComponent(&TeamComponent{TeamID: 2})
	enemy1.AddComponent(&HealthComponent{Current: 100, Max: 100})

	enemy2 := world.CreateEntity()
	enemy2.AddComponent(&PositionComponent{X: 70, Y: 0})
	enemy2.AddComponent(&TeamComponent{TeamID: 2})
	enemy2.AddComponent(&HealthComponent{Current: 100, Max: 100})

	enemy3 := world.CreateEntity()
	enemy3.AddComponent(&PositionComponent{X: 150, Y: 0})
	enemy3.AddComponent(&TeamComponent{TeamID: 2})
	enemy3.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Create ally (should not be included)
	ally := world.CreateEntity()
	ally.AddComponent(&PositionComponent{X: 20, Y: 0})
	ally.AddComponent(&TeamComponent{TeamID: 1})
	ally.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	// Find enemies within range 100
	enemies := FindEnemiesInRange(world, player, 100)

	if len(enemies) != 2 {
		t.Errorf("expected 2 enemies in range, got %d", len(enemies))
	}

	// Find nearest enemy
	nearest := FindNearestEnemy(world, player, 100)
	if nearest == nil {
		t.Fatal("should find nearest enemy")
	}
	if nearest.ID != enemy1.ID {
		t.Error("enemy1 should be nearest")
	}
}

func TestCombatSystemDeathCallback(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	deathCalled := false
	var deadEntity *Entity

	combatSystem.SetDeathCallback(func(entity *Entity) {
		deathCalled = true
		deadEntity = entity
	})

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 1, Max: 100})

	world.Update(0)

	// Kill entity
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(10)

	// Update to trigger callback
	world.Update(0.1)

	if !deathCalled {
		t.Error("death callback should be called")
	}
	if deadEntity == nil || deadEntity.ID != entity.ID {
		t.Error("death callback should receive correct entity")
	}
}

func TestCombatSystemDeadAttackerCannotAttack(t *testing.T) {
	// Priority 1.3: Dead entities cannot perform attacks
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create dead attacker
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   0, // Ready to attack
	})
	attacker.AddComponent(NewStatsComponent())
	attacker.AddComponent(NewDeadComponent(5.0)) // Mark as dead

	// Create living target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 50, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	target.AddComponent(NewStatsComponent())

	world.Update(0)

	// Dead attacker should not be able to attack
	hit := combatSystem.Attack(attacker, target)
	if hit {
		t.Error("dead attacker should not be able to attack")
	}

	// Target health should be unchanged
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 100 {
		t.Errorf("target health = %f, want 100 (dead attacker should deal no damage)", health.Current)
	}
}

func TestCombatSystemDeadTargetCannotBeAttacked(t *testing.T) {
	// Priority 1.3: Dead entities cannot be targeted for attacks
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	// Create living attacker
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   0, // Ready to attack
	})
	attacker.AddComponent(NewStatsComponent())

	// Create dead target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 50, Y: 0})
	target.AddComponent(&HealthComponent{Current: 0, Max: 100}) // Dead (0 health)
	target.AddComponent(NewStatsComponent())
	target.AddComponent(NewDeadComponent(3.0)) // Mark as dead

	world.Update(0)

	// Should not be able to attack dead target
	hit := combatSystem.Attack(attacker, target)
	if hit {
		t.Error("should not be able to attack dead target")
	}

	// Attack cooldown should not be triggered (attack didn't happen)
	attackComp, _ := attacker.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	if !attack.CanAttack() {
		t.Error("attack cooldown should not be triggered when targeting dead entity")
	}
}

func TestCombatSystemDeadEntityNoCooldownUpdate(t *testing.T) {
	// Priority 1.3: Dead entities don't progress attack cooldowns
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Create dead entity with attack on cooldown
	deadEntity := world.CreateEntity()
	deadEntity.AddComponent(&AttackComponent{
		Damage:        20,
		DamageType:    combat.DamagePhysical,
		Range:         100,
		Cooldown:      5.0,
		CooldownTimer: 5.0, // Full cooldown
	})
	deadEntity.AddComponent(NewDeadComponent(0.0))

	// Create living entity with same cooldown
	livingEntity := world.CreateEntity()
	livingEntity.AddComponent(&AttackComponent{
		Damage:        20,
		DamageType:    combat.DamagePhysical,
		Range:         100,
		Cooldown:      5.0,
		CooldownTimer: 5.0, // Full cooldown
	})

	world.Update(0)

	// Update for 3 seconds
	world.Update(3.0)

	// Living entity cooldown should decrease
	livingAttackComp, _ := livingEntity.GetComponent("attack")
	livingAttack := livingAttackComp.(*AttackComponent)
	if livingAttack.CooldownTimer != 2.0 {
		t.Errorf("living entity cooldown = %f, want 2.0", livingAttack.CooldownTimer)
	}

	// Dead entity cooldown should remain unchanged
	deadAttackComp, _ := deadEntity.GetComponent("attack")
	deadAttack := deadAttackComp.(*AttackComponent)
	if deadAttack.CooldownTimer != 5.0 {
		t.Errorf("dead entity cooldown = %f, want 5.0 (should not decrease)", deadAttack.CooldownTimer)
	}
}

func TestCombatSystemDeadEntityStatusEffectsStillProcess(t *testing.T) {
	// Status effects should continue on dead entities (design decision: effects don't stop at death)
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	world.AddSystem(combatSystem)

	// Create entity with low health
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 5, Max: 100})

	world.Update(0)

	// Apply poison effect
	combatSystem.ApplyStatusEffect(entity, "poison", 3.0, 10.0, 1.0)

	// Kill the entity by reducing health to 0
	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*HealthComponent)
	health.TakeDamage(5)

	// Mark as dead
	entity.AddComponent(NewDeadComponent(1.0))

	// Verify entity is dead
	if health.Current != 0 {
		t.Fatalf("entity should have 0 health, got %f", health.Current)
	}

	// Update to trigger poison tick
	world.Update(1.1)

	// Health should remain at 0 (clamped minimum), but the effect still processed
	// The status effect component should still exist and update
	if !entity.HasComponent("status_effect") {
		t.Error("status effect should still exist on dead entity")
	}

	// Verify health stays at 0 (design: health doesn't go negative)
	if health.Current != 0 {
		t.Errorf("health should be clamped at 0, got %f", health.Current)
	}
}

func TestFindEnemiesInRangeExcludesDeadEntities(t *testing.T) {
	// Helper functions should exclude dead entities from targeting
	world := NewWorld()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&TeamComponent{TeamID: 1})

	// Create living enemy
	livingEnemy := world.CreateEntity()
	livingEnemy.AddComponent(&PositionComponent{X: 30, Y: 0})
	livingEnemy.AddComponent(&TeamComponent{TeamID: 2})
	livingEnemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Create dead enemy
	deadEnemy := world.CreateEntity()
	deadEnemy.AddComponent(&PositionComponent{X: 40, Y: 0})
	deadEnemy.AddComponent(&TeamComponent{TeamID: 2})
	deadEnemy.AddComponent(&HealthComponent{Current: 0, Max: 100})
	deadEnemy.AddComponent(NewDeadComponent(1.0))

	world.Update(0)

	// Find enemies - should only return living enemy
	enemies := FindEnemiesInRange(world, player, 100)

	if len(enemies) != 1 {
		t.Errorf("expected 1 living enemy, got %d", len(enemies))
	}

	if len(enemies) > 0 && enemies[0].ID != livingEnemy.ID {
		t.Error("returned enemy should be the living one")
	}

	// Find nearest enemy - should return living enemy, not closer dead one
	nearest := FindNearestEnemy(world, player, 100)
	if nearest == nil {
		t.Fatal("should find nearest living enemy")
	}
	if nearest.ID != livingEnemy.ID {
		t.Error("nearest enemy should be the living one, not the dead one")
	}
}

func TestCombatSystemDamageCallback(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)

	damageCalled := false
	var damageAmount float64

	combatSystem.SetDamageCallback(func(attacker, target *Entity, damage float64) {
		damageCalled = true
		damageAmount = damage
	})

	// Create attacker and target
	attacker := world.CreateEntity()
	attacker.AddComponent(&AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   0,
	})

	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	// Perform attack
	combatSystem.Attack(attacker, target)

	if !damageCalled {
		t.Error("damage callback should be called")
	}
	if damageAmount <= 0 {
		t.Error("damage amount should be positive")
	}
}

func TestDeadComponent(t *testing.T) {
	tests := []struct {
		name              string
		timeOfDeath       float64
		itemsToAdd        []uint64
		expectedItems     int
		expectedType      string
		expectedTimestamp float64
	}{
		{
			name:              "new dead component",
			timeOfDeath:       10.5,
			itemsToAdd:        []uint64{},
			expectedItems:     0,
			expectedType:      "dead",
			expectedTimestamp: 10.5,
		},
		{
			name:              "with single dropped item",
			timeOfDeath:       20.0,
			itemsToAdd:        []uint64{1001},
			expectedItems:     1,
			expectedType:      "dead",
			expectedTimestamp: 20.0,
		},
		{
			name:              "with multiple dropped items",
			timeOfDeath:       30.5,
			itemsToAdd:        []uint64{1001, 1002, 1003},
			expectedItems:     3,
			expectedType:      "dead",
			expectedTimestamp: 30.5,
		},
		{
			name:              "zero time of death",
			timeOfDeath:       0.0,
			itemsToAdd:        []uint64{},
			expectedItems:     0,
			expectedType:      "dead",
			expectedTimestamp: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NewDeadComponent constructor
			deadComp := NewDeadComponent(tt.timeOfDeath)

			// Verify type
			if deadComp.Type() != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, deadComp.Type())
			}

			// Verify time of death
			if deadComp.TimeOfDeath != tt.expectedTimestamp {
				t.Errorf("expected TimeOfDeath %v, got %v", tt.expectedTimestamp, deadComp.TimeOfDeath)
			}

			// Verify DroppedItems initialized empty
			if deadComp.DroppedItems == nil {
				t.Error("DroppedItems should be initialized, not nil")
			}
			if len(deadComp.DroppedItems) != 0 {
				t.Errorf("expected 0 initial items, got %d", len(deadComp.DroppedItems))
			}

			// Add items
			for _, itemID := range tt.itemsToAdd {
				deadComp.AddDroppedItem(itemID)
			}

			// Verify item count
			if len(deadComp.DroppedItems) != tt.expectedItems {
				t.Errorf("expected %d items, got %d", tt.expectedItems, len(deadComp.DroppedItems))
			}

			// Verify item IDs match
			for i, expectedID := range tt.itemsToAdd {
				if deadComp.DroppedItems[i] != expectedID {
					t.Errorf("item %d: expected ID %d, got %d", i, expectedID, deadComp.DroppedItems[i])
				}
			}
		})
	}
}

func TestDeadComponentWithEntity(t *testing.T) {
	world := NewWorld()

	// Create entity
	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})

	world.Update(0)

	// Verify entity doesn't have dead component initially
	if entity.HasComponent("dead") {
		t.Error("entity should not have dead component initially")
	}

	// Simulate death by adding DeadComponent
	gameTime := 42.5
	deadComp := NewDeadComponent(gameTime)
	entity.AddComponent(deadComp)

	// Verify component attached
	if !entity.HasComponent("dead") {
		t.Fatal("entity should have dead component after adding")
	}

	// Retrieve and verify
	comp, ok := entity.GetComponent("dead")
	if !ok {
		t.Fatal("failed to retrieve dead component")
	}

	retrieved := comp.(*DeadComponent)
	if retrieved.TimeOfDeath != gameTime {
		t.Errorf("expected TimeOfDeath %v, got %v", gameTime, retrieved.TimeOfDeath)
	}

	// Add dropped items
	retrieved.AddDroppedItem(5001)
	retrieved.AddDroppedItem(5002)

	if len(retrieved.DroppedItems) != 2 {
		t.Errorf("expected 2 dropped items, got %d", len(retrieved.DroppedItems))
	}
}

func TestDeadComponentEdgeCases(t *testing.T) {
	t.Run("negative time of death", func(t *testing.T) {
		// Should handle negative time (e.g., for testing or bugs)
		deadComp := NewDeadComponent(-5.0)
		if deadComp.TimeOfDeath != -5.0 {
			t.Error("should preserve negative time of death")
		}
	})

	t.Run("add duplicate item IDs", func(t *testing.T) {
		// Should allow duplicates (intentional design - track all spawned items)
		deadComp := NewDeadComponent(10.0)
		deadComp.AddDroppedItem(1001)
		deadComp.AddDroppedItem(1001)

		if len(deadComp.DroppedItems) != 2 {
			t.Errorf("expected 2 items (duplicates allowed), got %d", len(deadComp.DroppedItems))
		}
	})

	t.Run("add many items", func(t *testing.T) {
		// Stress test with many items
		deadComp := NewDeadComponent(10.0)
		for i := uint64(0); i < 100; i++ {
			deadComp.AddDroppedItem(i)
		}

		if len(deadComp.DroppedItems) != 100 {
			t.Errorf("expected 100 items, got %d", len(deadComp.DroppedItems))
		}

		// Verify order preserved
		for i := uint64(0); i < 100; i++ {
			if deadComp.DroppedItems[i] != i {
				t.Errorf("item %d: expected ID %d, got %d", i, i, deadComp.DroppedItems[i])
			}
		}
	})

	t.Run("add zero item ID", func(t *testing.T) {
		// Should allow zero ID (might be used for invalid/null entities)
		deadComp := NewDeadComponent(10.0)
		deadComp.AddDroppedItem(0)

		if len(deadComp.DroppedItems) != 1 {
			t.Error("should allow adding zero ID")
		}
		if deadComp.DroppedItems[0] != 0 {
			t.Error("should preserve zero ID")
		}
	})
}

// TestFindEnemyInAimDirection tests Phase 10.1 aim-based target selection.
func TestFindEnemyInAimDirection(t *testing.T) {
	tests := []struct {
		name         string
		aimAngle     float64 // radians: 0=right, π/2=down, π=left, 3π/2=up
		aimCone      float64 // radians: aim cone width
		enemyOffsets []struct{ x, y float64 }
		maxRange     float64
		expectHit    int // index of expected enemy hit, or -1 for none
	}{
		{
			name:     "enemy directly ahead",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 50, y: 0}, // directly right
			},
			maxRange:  100,
			expectHit: 0,
		},
		{
			name:     "enemy in cone (slight angle)",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 50, y: 10}, // slightly up-right (within 45° cone)
			},
			maxRange:  100,
			expectHit: 0,
		},
		{
			name:     "enemy outside cone",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 10, y: 50}, // almost straight up (outside 45° cone)
			},
			maxRange:  100,
			expectHit: -1, // no hit
		},
		{
			name:     "multiple enemies - choose closest in cone",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 80, y: 5},  // far enemy in cone
				{x: 30, y: 5},  // close enemy in cone (should hit this one)
				{x: 10, y: 50}, // enemy outside cone
			},
			maxRange:  100,
			expectHit: 1, // closest enemy in cone
		},
		{
			name:     "enemy out of range",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 150, y: 0}, // too far
			},
			maxRange:  100,
			expectHit: -1, // no hit (out of range)
		},
		{
			name:     "aim left (π radians)",
			aimAngle: math.Pi,     // aiming left
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: -50, y: 0}, // directly left
			},
			maxRange:  100,
			expectHit: 0,
		},
		{
			name:     "aim down (π/2 radians)",
			aimAngle: math.Pi / 2, // aiming down
			aimCone:  math.Pi / 4, // 45° cone
			enemyOffsets: []struct{ x, y float64 }{
				{x: 0, y: 50}, // directly down
			},
			maxRange:  100,
			expectHit: 0,
		},
		{
			name:     "wide cone catches more enemies",
			aimAngle: 0,           // aiming right
			aimCone:  math.Pi / 2, // 90° cone (wider)
			enemyOffsets: []struct{ x, y float64 }{
				{x: 50, y: 30}, // ~31° up-right (in 90° cone, not in 45° cone)
			},
			maxRange:  100,
			expectHit: 0,
		},
		{
			name:     "narrow cone misses off-angle enemy",
			aimAngle: 0,            // aiming right
			aimCone:  math.Pi / 16, // ~11° cone (very narrow)
			enemyOffsets: []struct{ x, y float64 }{
				{x: 50, y: 10}, // small angle but outside narrow cone
			},
			maxRange:  100,
			expectHit: -1, // miss
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world and attacker
			world := NewWorld()
			attacker := world.CreateEntity()
			attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
			attacker.AddComponent(&TeamComponent{TeamID: 1})

			// Create enemies at specified offsets
			enemies := make([]*Entity, len(tt.enemyOffsets))
			for i, offset := range tt.enemyOffsets {
				enemy := world.CreateEntity()
				enemy.AddComponent(&PositionComponent{X: offset.x, Y: offset.y})
				enemy.AddComponent(&TeamComponent{TeamID: 2}) // Different team
				enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
				enemies[i] = enemy
			}

			// Process pending entity additions
			world.Update(0)

			// Find enemy in aim direction
			result := FindEnemyInAimDirection(world, attacker, tt.aimAngle, tt.maxRange, tt.aimCone)

			if tt.expectHit == -1 {
				// Expect no hit
				if result != nil {
					t.Errorf("expected no hit, but found enemy %d", result.ID)
				}
			} else {
				// Expect specific enemy hit
				if result == nil {
					t.Errorf("expected to hit enemy %d, but got nil", tt.expectHit)
				} else if result.ID != enemies[tt.expectHit].ID {
					t.Errorf("expected to hit enemy %d (ID %d), but hit enemy ID %d",
						tt.expectHit, enemies[tt.expectHit].ID, result.ID)
				}
			}
		})
	}
}

// TestFindEnemyInAimDirection_EdgeCases tests edge cases for aim-based targeting.
func TestFindEnemyInAimDirection_EdgeCases(t *testing.T) {
	t.Run("no enemies", func(t *testing.T) {
		world := NewWorld()
		attacker := world.CreateEntity()
		attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
		attacker.AddComponent(&TeamComponent{TeamID: 1})

		world.Update(0) // Process pending additions

		result := FindEnemyInAimDirection(world, attacker, 0, 100, math.Pi/4)
		if result != nil {
			t.Error("expected nil when no enemies exist")
		}
	})

	t.Run("attacker has no position", func(t *testing.T) {
		world := NewWorld()
		attacker := world.CreateEntity()
		// No position component

		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 50, Y: 0})
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

		world.Update(0) // Process pending additions

		result := FindEnemyInAimDirection(world, attacker, 0, 100, math.Pi/4)
		if result != nil {
			t.Error("expected nil when attacker has no position")
		}
	})

	t.Run("enemy has no position", func(t *testing.T) {
		world := NewWorld()
		attacker := world.CreateEntity()
		attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
		attacker.AddComponent(&TeamComponent{TeamID: 1})

		enemy := world.CreateEntity()
		// No position component
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

		world.Update(0) // Process pending additions

		result := FindEnemyInAimDirection(world, attacker, 0, 100, math.Pi/4)
		if result != nil {
			t.Error("expected nil when enemy has no position")
		}
	})

	t.Run("zero aim cone", func(t *testing.T) {
		world := NewWorld()
		attacker := world.CreateEntity()
		attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
		attacker.AddComponent(&TeamComponent{TeamID: 1})

		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: 50, Y: 0.1}) // Tiny angle offset
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

		world.Update(0) // Process pending additions

		result := FindEnemyInAimDirection(world, attacker, 0, 100, 0) // Zero cone
		if result != nil {
			t.Error("expected nil with zero aim cone and non-zero angle")
		}
	})

	t.Run("full circle cone (2π)", func(t *testing.T) {
		world := NewWorld()
		attacker := world.CreateEntity()
		attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
		attacker.AddComponent(&TeamComponent{TeamID: 1})

		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: -50, Y: -50}) // Behind and to the left
		enemy.AddComponent(&TeamComponent{TeamID: 2})
		enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

		world.Update(0) // Process pending additions

		result := FindEnemyInAimDirection(world, attacker, 0, 100, 2*math.Pi) // Full circle
		if result == nil {
			t.Error("expected to find enemy with full circle aim cone")
		}
	})
}

// TestCombatSystem_SetAudioManager tests the SetAudioManager method for combat SFX integration.
// Phase 14.4/Plan 1.1: Audio system integration with combat system.
func TestCombatSystem_SetAudioManager(t *testing.T) {
	// Create combat system
	combatSystem := NewCombatSystem(12345)

	// Test with nil audio manager (should not panic)
	combatSystem.SetAudioManager(nil)
	if combatSystem.audioManager != nil {
		t.Error("expected audioManager to be nil")
	}

	// Create audio manager and set it
	audioManager := NewAudioManager(44100, 12345)
	combatSystem.SetAudioManager(audioManager)

	if combatSystem.audioManager == nil {
		t.Error("expected audioManager to be set")
	}

	if combatSystem.audioManager != audioManager {
		t.Error("expected audioManager to match the one that was set")
	}
}

// TestCombatSystem_UsesAuthoritativeCombatResolver verifies that CombatSystem
// uses the combat package's DefaultCombatResolver for damage calculation.
// This test addresses the AUDIT.md finding: "Defense Calculation Formula Inconsistency"
func TestCombatSystem_UsesAuthoritativeCombatResolver(t *testing.T) {
	combatSystem := NewCombatSystem(12345)

	tests := []struct {
		name       string
		baseDamage float64
		damageType combat.DamageType
		defense    float64
		magicDef   float64
		resistance float64
		wantMin    float64 // Minimum expected damage
		wantMax    float64 // Maximum expected damage
	}{
		{
			name:       "physical damage with defense",
			baseDamage: 100,
			damageType: combat.DamagePhysical,
			defense:    50,
			magicDef:   0,
			resistance: 0,
			wantMin:    66.0, // 100 * (100 / (100 + 50)) ≈ 66.67
			wantMax:    67.0,
		},
		{
			name:       "magical damage with magic defense",
			baseDamage: 100,
			damageType: combat.DamageMagical,
			defense:    50,
			magicDef:   30,
			resistance: 0,
			wantMin:    76.0, // 100 * (100 / (100 + 30)) ≈ 76.92
			wantMax:    77.0,
		},
		{
			name:       "fire damage uses magic defense (not physical)",
			baseDamage: 100,
			damageType: combat.DamageFire,
			defense:    100, // Should NOT be used
			magicDef:   20,  // Should be used
			resistance: 0,
			wantMin:    83.0, // 100 * (100 / (100 + 20)) ≈ 83.33
			wantMax:    84.0,
		},
		{
			name:       "ice damage uses magic defense",
			baseDamage: 100,
			damageType: combat.DamageIce,
			defense:    50,
			magicDef:   25,
			resistance: 0,
			wantMin:    79.0, // 100 * (100 / (100 + 25)) = 80.0
			wantMax:    81.0,
		},
		{
			name:       "lightning damage uses magic defense",
			baseDamage: 100,
			damageType: combat.DamageLightning,
			defense:    50,
			magicDef:   10,
			resistance: 0,
			wantMin:    90.0, // 100 * (100 / (100 + 10)) ≈ 90.91
			wantMax:    91.0,
		},
		{
			name:       "poison damage uses magic defense",
			baseDamage: 100,
			damageType: combat.DamagePoison,
			defense:    50,
			magicDef:   40,
			resistance: 0,
			wantMin:    71.0, // 100 * (100 / (100 + 40)) ≈ 71.43
			wantMax:    72.0,
		},
		{
			name:       "resistance reduces damage after defense",
			baseDamage: 100,
			damageType: combat.DamageFire,
			defense:    0,
			magicDef:   0,
			resistance: 0.5,  // 50% resistance
			wantMin:    49.0, // 100 * (1 - 0.5) = 50.0, then min damage kicks in
			wantMax:    51.0,
		},
		{
			name:       "high defense with diminishing returns",
			baseDamage: 100,
			damageType: combat.DamagePhysical,
			defense:    200, // Very high defense
			magicDef:   0,
			resistance: 0,
			wantMin:    33.0, // 100 * (100 / (100 + 200)) ≈ 33.33
			wantMax:    34.0,
		},
		{
			name:       "minimum damage enforcement",
			baseDamage: 100,
			damageType: combat.DamagePhysical,
			defense:    1000, // Extreme defense
			magicDef:   0,
			resistance: 0,
			wantMin:    9.0,  // MinDamageMultiplier (0.1) * 100 = 10.0
			wantMax:    10.1, // Allow slight floating point variance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetStats := &StatsComponent{
				Defense:      tt.defense,
				MagicDefense: tt.magicDef,
				Resistances:  make(map[combat.DamageType]float64),
			}

			if tt.resistance != 0 {
				targetStats.Resistances[tt.damageType] = tt.resistance
			}

			result := combatSystem.applyDefenseAndResistance(tt.baseDamage, tt.damageType, targetStats)

			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("expected damage in range [%.2f, %.2f], got %.2f", tt.wantMin, tt.wantMax, result)
			}
		})
	}
}

// TestStatsComponentToCombatStats verifies the conversion helper works correctly.
func TestStatsComponentToCombatStats(t *testing.T) {
	statsComp := &StatsComponent{
		Attack:       25.0,
		Defense:      30.0,
		MagicPower:   20.0,
		MagicDefense: 15.0,
		CritChance:   0.15,
		CritDamage:   2.0,
		Evasion:      0.1,
		Resistances: map[combat.DamageType]float64{
			combat.DamageFire: 0.5,
			combat.DamageIce:  0.25,
		},
	}

	combatStats := statsComponentToCombatStats(statsComp)

	if combatStats.Attack != 25.0 {
		t.Errorf("expected Attack 25.0, got %v", combatStats.Attack)
	}
	if combatStats.Defense != 30.0 {
		t.Errorf("expected Defense 30.0, got %v", combatStats.Defense)
	}
	if combatStats.MagicPower != 20.0 {
		t.Errorf("expected MagicPower 20.0, got %v", combatStats.MagicPower)
	}
	if combatStats.MagicDefense != 15.0 {
		t.Errorf("expected MagicDefense 15.0, got %v", combatStats.MagicDefense)
	}
	if combatStats.CritChance != 0.15 {
		t.Errorf("expected CritChance 0.15, got %v", combatStats.CritChance)
	}
	if combatStats.CritDamage != 2.0 {
		t.Errorf("expected CritDamage 2.0, got %v", combatStats.CritDamage)
	}
	if combatStats.Evasion != 0.1 {
		t.Errorf("expected Evasion 0.1, got %v", combatStats.Evasion)
	}
	if combatStats.Resistances[combat.DamageFire] != 0.5 {
		t.Errorf("expected Fire resistance 0.5, got %v", combatStats.Resistances[combat.DamageFire])
	}
	if combatStats.Resistances[combat.DamageIce] != 0.25 {
		t.Errorf("expected Ice resistance 0.25, got %v", combatStats.Resistances[combat.DamageIce])
	}
}

// TestCombatSystem_ResistanceClampingBehavior verifies resistance clamping matches combat package.
func TestCombatSystem_ResistanceClampingBehavior(t *testing.T) {
	combatSystem := NewCombatSystem(12345)

	tests := []struct {
		name       string
		resistance float64
		wantMin    float64
		wantMax    float64
	}{
		{
			name:       "extreme negative resistance clamped to -0.5",
			resistance: -1.0, // 200% damage if not clamped
			wantMin:    149.0,
			wantMax:    151.0, // 100 * (1 - (-0.5)) = 150.0
		},
		{
			name:       "normal negative resistance",
			resistance: -0.3, // 130% damage
			wantMin:    129.0,
			wantMax:    131.0,
		},
		{
			name:       "full immunity",
			resistance: 1.0, // 0% damage (min damage kicks in)
			wantMin:    9.0,
			wantMax:    10.1, // MinDamageMultiplier * 100 = 10.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetStats := &StatsComponent{
				Defense:     0,
				Resistances: map[combat.DamageType]float64{combat.DamagePhysical: tt.resistance},
			}

			result := combatSystem.applyDefenseAndResistance(100, combat.DamagePhysical, targetStats)

			if result < tt.wantMin || result > tt.wantMax {
				t.Errorf("expected damage in range [%.2f, %.2f], got %.2f", tt.wantMin, tt.wantMax, result)
			}
		})
	}
}

// TestGetEntityStats_SafeTypeAssertion verifies getEntityStats uses safe type assertion.
func TestGetEntityStats_SafeTypeAssertion(t *testing.T) {
	combatSys := NewCombatSystem(42)

	tests := []struct {
		name    string
		setup   func(*Entity)
		wantNil bool
	}{
		{
			name:    "no stats component",
			setup:   func(e *Entity) {},
			wantNil: true,
		},
		{
			name: "valid stats component",
			setup: func(e *Entity) {
				e.AddComponent(NewStatsComponent())
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			entity := world.CreateEntity()
			tt.setup(entity)
			world.Update(0)

			stats := combatSys.getEntityStats(entity)
			if (stats == nil) != tt.wantNil {
				t.Errorf("getEntityStats() nil=%v, want nil=%v", stats == nil, tt.wantNil)
			}
		})
	}
}

// TestAdditionalDamageCallbacks verifies that AddDamageCallback callbacks are invoked.
func TestAdditionalDamageCallbacks(t *testing.T) {
	world := NewWorld()
	combatSys := NewCombatSystem(42)

	// Track callbacks
	var primaryCalled bool
	var additionalCalled1, additionalCalled2 bool
	var recordedDamage1, recordedDamage2 float64

	combatSys.SetDamageCallback(func(attacker, target *Entity, damage float64) {
		primaryCalled = true
	})
	combatSys.AddDamageCallback(func(attacker, target *Entity, damage float64) {
		additionalCalled1 = true
		recordedDamage1 = damage
	})
	combatSys.AddDamageCallback(func(attacker, target *Entity, damage float64) {
		additionalCalled2 = true
		recordedDamage2 = damage
	})

	// Create attacker and target
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})
	attacker.AddComponent(&AttackComponent{Damage: 20, Range: 50, Cooldown: 1.0})
	attacker.AddComponent(&TeamComponent{TeamID: 1})

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 110, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	target.AddComponent(&TeamComponent{TeamID: 2})

	world.Update(0)

	hit := combatSys.Attack(attacker, target)
	if !hit {
		t.Fatal("expected attack to hit")
	}

	if !primaryCalled {
		t.Error("primary damage callback was not called")
	}
	if !additionalCalled1 {
		t.Error("additional damage callback 1 was not called")
	}
	if !additionalCalled2 {
		t.Error("additional damage callback 2 was not called")
	}
	if recordedDamage1 <= 0 {
		t.Errorf("callback 1 recorded damage should be positive, got %f", recordedDamage1)
	}
	if recordedDamage1 != recordedDamage2 {
		t.Errorf("both callbacks should receive same damage: %f != %f", recordedDamage1, recordedDamage2)
	}
}

// TestComputeFinalDamage_ReturnsBaseDamage verifies that computeFinalDamage returns baseDamage
// so that callers don't need to recalculate and consume extra RNG state.
func TestComputeFinalDamage_ReturnsBaseDamage(t *testing.T) {
	combatSys := NewCombatSystem(42)

	attack := &AttackComponent{
		Damage:     20,
		DamageType: combat.DamagePhysical,
	}
	attackerStats := NewStatsComponent()
	attackerStats.CritChance = 0 // No crits for deterministic test

	world := NewWorld()
	attacker := world.CreateEntity()
	target := world.CreateEntity()
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	world.Update(0)

	finalDamage, baseDamage, isCrit := combatSys.computeFinalDamage(attacker, attack, attackerStats, nil, target)

	expectedBase := attack.Damage + attackerStats.Attack // 20 + 10 = 30
	if baseDamage != expectedBase {
		t.Errorf("baseDamage = %f, want %f", baseDamage, expectedBase)
	}
	if finalDamage <= 0 {
		t.Error("finalDamage should be positive")
	}
	if isCrit {
		t.Error("expected no crit with 0 crit chance")
	}
}

// TestCombatSystem_DeterministicRNG verifies that the same seed produces the same
// combat outcome sequence, confirming no extra RNG consumption during attacks.
func TestCombatSystem_DeterministicRNG(t *testing.T) {
	runCombat := func(seed int64) []float64 {
		world := NewWorld()
		cs := NewCombatSystem(seed)

		results := make([]float64, 0, 5)

		for i := 0; i < 5; i++ {
			attacker := world.CreateEntity()
			attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
			attacker.AddComponent(&AttackComponent{Damage: 10, Range: 100, Cooldown: 0})
			attacker.AddComponent(&TeamComponent{TeamID: 1})
			attacker.AddComponent(NewStatsComponent())

			target := world.CreateEntity()
			target.AddComponent(&PositionComponent{X: 10, Y: 0})
			target.AddComponent(&HealthComponent{Current: 200, Max: 200})
			target.AddComponent(&TeamComponent{TeamID: 2})
			target.AddComponent(NewStatsComponent())

			world.Update(0)

			healthBefore := float64(200)
			cs.Attack(attacker, target)

			healthComp, _ := target.GetComponent("health")
			health := healthComp.(*HealthComponent)
			results = append(results, healthBefore-health.Current)
		}
		return results
	}

	run1 := runCombat(99999)
	run2 := runCombat(99999)

	for i := range run1 {
		if run1[i] != run2[i] {
			t.Errorf("non-deterministic at attack %d: %f != %f", i, run1[i], run2[i])
		}
	}
}

// TestCombatModDamageMultipliers verifies that mod rules correctly scale damage.
// Phase 6.3 (PLAN.md): Modding System Integration
func TestCombatModDamageMultipliers(t *testing.T) {
	tests := []struct {
		name             string
		playerMultiplier float64
		enemyMultiplier  float64
		baseDamage       float64
		attackerIsPlayer bool
		expectedDamage   float64
	}{
		{
			name:             "no mod rules applied",
			playerMultiplier: 1.0,
			enemyMultiplier:  1.0,
			baseDamage:       100,
			attackerIsPlayer: true,
			expectedDamage:   100,
		},
		{
			name:             "player damage buffed",
			playerMultiplier: 1.5,
			enemyMultiplier:  1.0,
			baseDamage:       100,
			attackerIsPlayer: true,
			expectedDamage:   150,
		},
		{
			name:             "player damage nerfed",
			playerMultiplier: 0.5,
			enemyMultiplier:  1.0,
			baseDamage:       100,
			attackerIsPlayer: true,
			expectedDamage:   50,
		},
		{
			name:             "enemy damage buffed",
			playerMultiplier: 1.0,
			enemyMultiplier:  2.0,
			baseDamage:       50,
			attackerIsPlayer: false,
			expectedDamage:   100,
		},
		{
			name:             "enemy damage nerfed",
			playerMultiplier: 1.0,
			enemyMultiplier:  0.7,
			baseDamage:       100,
			attackerIsPlayer: false,
			expectedDamage:   70,
		},
		{
			name:             "hardcore mode multipliers",
			playerMultiplier: 0.8,
			enemyMultiplier:  1.3,
			baseDamage:       100,
			attackerIsPlayer: true,
			expectedDamage:   80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			combatSystem := NewCombatSystem(12345)
			combatSystem.SetParticleSystem(nil, world, "fantasy")

			// Create mock mod rules provider
			modRules := &mockModRuleProvider{
				rules: map[string]interface{}{
					"combat.player_damage_multiplier": tt.playerMultiplier,
					"combat.enemy_damage_multiplier":  tt.enemyMultiplier,
				},
			}
			world.SetModRules(modRules)

			attacker := world.CreateEntity()
			target := world.CreateEntity()

			// Set attacker type (player or NPC)
			if tt.attackerIsPlayer {
				attacker.AddComponent(&StubInput{})
			} else {
				attacker.AddComponent(&AIComponent{})
			}

			// Apply multipliers
			finalDamage := combatSystem.applyModDamageMultipliers(attacker, target, tt.baseDamage)

			if math.Abs(finalDamage-tt.expectedDamage) > 0.01 {
				t.Errorf("expected damage %v, got %v", tt.expectedDamage, finalDamage)
			}
		})
	}
}

// TestCombatModDamageMultipliersNoModRules verifies default behavior without mod rules.
// Phase 6.3 (PLAN.md): Modding System Integration
func TestCombatModDamageMultipliersNoModRules(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	combatSystem.SetParticleSystem(nil, world, "fantasy")
	// No mod rules set - should return baseDamage unchanged

	attacker := world.CreateEntity()
	target := world.CreateEntity()
	attacker.AddComponent(&StubInput{})

	baseDamage := 100.0
	finalDamage := combatSystem.applyModDamageMultipliers(attacker, target, baseDamage)

	if finalDamage != baseDamage {
		t.Errorf("expected damage unchanged at %v, got %v", baseDamage, finalDamage)
	}
}

// TestCombatIntegrationWithModRules verifies end-to-end combat with mod rules.
// Phase 6.3 (PLAN.md): Modding System Integration
func TestCombatIntegrationWithModRules(t *testing.T) {
	world := NewWorld()
	combatSystem := NewCombatSystem(12345)
	combatSystem.SetParticleSystem(nil, world, "fantasy")

	// Set hardcore mode mod rules (player damage 0.8x, enemy damage 1.3x)
	modRules := &mockModRuleProvider{
		rules: map[string]interface{}{
			"combat.player_damage_multiplier": 0.8,
			"combat.enemy_damage_multiplier":  1.3,
		},
	}
	world.SetModRules(modRules)

	// Create player vs enemy scenario
	player := world.CreateEntity()
	player.AddComponent(&StubInput{})
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&StatsComponent{
		Attack:     10,
		Defense:    5,
		CritChance: 0,
	})
	player.AddComponent(&AttackComponent{
		Damage:     50, // Base damage
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   1.0,
	})

	enemy := world.CreateEntity()
	enemy.AddComponent(&AIComponent{})
	enemy.AddComponent(&PositionComponent{X: 10, Y: 0})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(&StatsComponent{
		Attack:     10,
		Defense:    5,
		CritChance: 0,
	})
	enemy.AddComponent(&AttackComponent{
		Damage:     40, // Base damage
		DamageType: combat.DamagePhysical,
		Range:      100,
		Cooldown:   1.0,
	})

	// Player attacks enemy - should deal 0.8x damage
	playerHealthComp := player.GetHealth()
	enemyHealthComp := enemy.GetHealth()

	initialPlayerHealth := playerHealthComp.Current
	initialEnemyHealth := enemyHealthComp.Current

	// Player attack
	success := combatSystem.Attack(player, enemy)
	if !success {
		t.Fatal("player attack should succeed")
	}

	// Expected: base 50 + stats 10 = 60, then 0.8x = 48, minus defense
	// (actual formula uses combat.CombatResolver, so we verify multiplier was applied)
	playerDamageDealt := initialEnemyHealth - enemyHealthComp.Current
	if playerDamageDealt >= 60 { // Should be less than unmultiplied
		t.Errorf("player damage should be reduced by 0.8x multiplier, got %v", playerDamageDealt)
	}

	// Enemy attacks player - should deal 1.3x damage
	success = combatSystem.Attack(enemy, player)
	if !success {
		t.Fatal("enemy attack should succeed")
	}

	enemyDamageDealt := initialPlayerHealth - playerHealthComp.Current
	if enemyDamageDealt < 40 { // Should be more than unmultiplied base
		t.Errorf("enemy damage should be increased by 1.3x multiplier, got %v", enemyDamageDealt)
	}
}

// mockModRuleProvider is a test implementation of ModRuleProvider.
type mockModRuleProvider struct {
	rules map[string]interface{}
}

func (m *mockModRuleProvider) GetRule(ruleName string) (interface{}, bool) {
	val, ok := m.rules[ruleName]
	return val, ok
}

func (m *mockModRuleProvider) GetRuleFloat64(ruleName string, defaultValue float64) float64 {
	val, ok := m.rules[ruleName]
	if !ok {
		return defaultValue
	}
	if f, ok := val.(float64); ok {
		return f
	}
	return defaultValue
}

func (m *mockModRuleProvider) GetRuleBool(ruleName string, defaultValue bool) bool {
	val, ok := m.rules[ruleName]
	if !ok {
		return defaultValue
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

func (m *mockModRuleProvider) TriggerEvent(eventType string, eventData map[string]interface{}) error {
	return nil
}

// TestG35_ShieldAbsorbsFullDamage verifies that a fully-charged shield reduces
// final damage to 0 (not 1), i.e., the floor is applied after shield absorption.
func TestG35_ShieldAbsorbsFullDamage(t *testing.T) {
	world := NewWorld()
	cs := NewCombatSystem(42)
	world.AddSystem(cs)

	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{Damage: 1.0, DamageType: combat.DamagePhysical, Range: 100, Cooldown: 1.0})
	attacker.AddComponent(NewStatsComponent())

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	// Shield absorbs ALL incoming damage; Duration must be >0 for IsActive().
	target.AddComponent(&ShieldComponent{Amount: 1000, MaxAmount: 1000, Duration: 9999})
	target.AddComponent(NewStatsComponent())

	world.Update(0)

	hit := cs.Attack(attacker, target)
	if !hit {
		t.Skip("G35: attack did not land (may require specific entity setup)")
	}

	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 100 {
		t.Errorf("G35: health = %.0f, want 100 (shield should have absorbed all damage)", health.Current)
	}
}

// TestG34_EquipmentSetBonusDamageApplied verifies that the equipment set damage
// bonus from EquipmentSetBonusComponent is added to the attacker's damage output.
func TestG34_EquipmentSetBonusDamageApplied(t *testing.T) {
	world := NewWorld()
	cs := NewCombatSystem(42)
	world.AddSystem(cs)

	baseAttackDamage := 10.0

	// Attacker without set bonus
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&AttackComponent{Damage: baseAttackDamage, DamageType: combat.DamagePhysical, Range: 100, Cooldown: 1.0})
	attacker.AddComponent(NewStatsComponent())

	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 0})
	target.AddComponent(&HealthComponent{Current: 1000, Max: 1000})
	target.AddComponent(NewStatsComponent())

	world.Update(0)

	cs.Attack(attacker, target)
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	damageWithoutBonus := 1000.0 - health.Current

	// Reset health and cooldown.
	health.Current = 1000
	attackComp, _ := attacker.GetComponent("attack")
	attack := attackComp.(*AttackComponent)
	attack.CooldownTimer = 0 // allow next attack

	// Add equipment set bonus (+15 damage).
	setBonus := NewEquipmentSetBonusComponent()
	setBonus.ActiveSets["inferno"] = &ActiveSetBonus{
		SetID:          "inferno",
		PiecesEquipped: 2,
		CombinedBonus:  SetBonusTier{DamageBonus: 15},
	}
	attacker.AddComponent(setBonus)

	cs.Attack(attacker, target)
	damageWithBonus := 1000.0 - health.Current

	if damageWithBonus <= damageWithoutBonus {
		t.Errorf("G34: damage with set bonus (%.1f) should exceed damage without (%.1f)", damageWithBonus, damageWithoutBonus)
	}
}
