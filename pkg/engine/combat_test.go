package engine

import (
	"testing"
	
	"github.com/opd-ai/venture/pkg/combat"
)

func TestHealthComponent(t *testing.T) {
	tests := []struct {
		name           string
		initial        float64
		max            float64
		operation      string
		amount         float64
		expectedCurrent float64
		expectedAlive  bool
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
	targetStats.Evasion = 0 // No evasion
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
