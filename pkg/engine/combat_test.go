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
			aimAngle: 0,                  // aiming right
			aimCone:  math.Pi / 4,        // 45° cone
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
				{x: 30, y: 40}, // 45° up-right (in 90° cone, not in 45° cone)
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

		result := FindEnemyInAimDirection(world, attacker, 0, 100, 2*math.Pi) // Full circle
		if result == nil {
			t.Error("expected to find enemy with full circle aim cone")
		}
	})
}
