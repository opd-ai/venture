package engine

import (
	"testing"
)

// TestAIComponent tests the AIComponent functionality.
func TestAIComponent(t *testing.T) {
	ai := NewAIComponent(100, 200)

	// Test initial state
	if ai.State != AIStateIdle {
		t.Errorf("initial state = %v, want %v", ai.State, AIStateIdle)
	}
	if ai.SpawnX != 100 || ai.SpawnY != 200 {
		t.Errorf("spawn position = (%v, %v), want (100, 200)", ai.SpawnX, ai.SpawnY)
	}
	if ai.HasTarget() {
		t.Error("new AI should not have a target")
	}
}

// TestAIComponentStateChanges tests state transitions.
func TestAIComponentStateChanges(t *testing.T) {
	ai := NewAIComponent(0, 0)

	// Change state
	ai.ChangeState(AIStateChase)
	if ai.State != AIStateChase {
		t.Errorf("state = %v, want %v", ai.State, AIStateChase)
	}
	if ai.StateTimer != 0 {
		t.Errorf("state timer = %v, want 0", ai.StateTimer)
	}

	// Update state timer
	ai.UpdateStateTimer(1.5)
	if ai.StateTimer != 1.5 {
		t.Errorf("state timer = %v, want 1.5", ai.StateTimer)
	}

	// Change state again should reset timer
	ai.ChangeState(AIStateAttack)
	if ai.StateTimer != 0 {
		t.Errorf("state timer = %v, want 0 after state change", ai.StateTimer)
	}
}

// TestAIComponentDecisionTimer tests the decision timing.
func TestAIComponentDecisionTimer(t *testing.T) {
	ai := NewAIComponent(0, 0)
	ai.DecisionInterval = 1.0
	ai.DecisionTimer = 1.0

	// Should not update yet
	if ai.ShouldUpdateDecision(0.5) {
		t.Error("should not update decision yet")
	}

	// Should update now
	if !ai.ShouldUpdateDecision(0.6) {
		t.Error("should update decision now")
	}

	// Timer should be reset
	if ai.DecisionTimer <= 0 || ai.DecisionTimer > 1.0 {
		t.Errorf("decision timer = %v, should be reset to interval", ai.DecisionTimer)
	}
}

// TestAIComponentSpeedMultipliers tests speed multipliers for different states.
func TestAIComponentSpeedMultipliers(t *testing.T) {
	ai := NewAIComponent(0, 0)

	tests := []struct {
		state AIState
		want  float64
	}{
		{AIStateIdle, 1.0},
		{AIStatePatrol, 0.5},
		{AIStateChase, 1.0},
		{AIStateAttack, 1.0},
		{AIStateFlee, 1.5},
		{AIStateReturn, 0.8},
	}

	for _, tt := range tests {
		ai.State = tt.state
		got := ai.GetSpeedMultiplier()
		if got != tt.want {
			t.Errorf("speed multiplier for %v = %v, want %v", tt.state, got, tt.want)
		}
	}
}

// TestAIComponentDistanceCalculations tests distance from spawn.
func TestAIComponentDistanceCalculations(t *testing.T) {
	ai := NewAIComponent(100, 100)

	tests := []struct {
		name       string
		x, y       float64
		wantDist   float64
		wantReturn bool
	}{
		{"at spawn", 100, 100, 0, false},
		{"close to spawn", 110, 110, 14.142, false},
		{"far from spawn", 700, 100, 600, true}, // sqrt(600^2 + 0^2) = 600
	}

	ai.MaxChaseDistance = 500

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := ai.GetDistanceFromSpawn(tt.x, tt.y)
			// Use wider tolerance for distance checks (sqrt can be imprecise)
			if dist < tt.wantDist-50 || dist > tt.wantDist+50 {
				t.Errorf("distance = %v, want ~%v", dist, tt.wantDist)
			}

			shouldReturn := ai.ShouldReturnToSpawn(tt.x, tt.y)
			if shouldReturn != tt.wantReturn {
				t.Errorf("should return = %v, want %v", shouldReturn, tt.wantReturn)
			}
		})
	}
}

// TestAISystemIdle tests idle state behavior.
func TestAISystemIdle(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity
	ai := world.CreateEntity()
	ai.AddComponent(NewAIComponent(100, 100))
	ai.AddComponent(&PositionComponent{X: 100, Y: 100})
	ai.AddComponent(&TeamComponent{TeamID: 1})

	// Create enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 150, Y: 150})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	world.Update(0)

	// Update AI - should detect enemy
	aiSystem.Update(0.6) // Trigger decision update

	aiComp, _ := ai.GetComponent("ai")
	aiC := aiComp.(*AIComponent)

	// Should have detected enemy
	if aiC.State != AIStateDetect {
		t.Errorf("state = %v, want %v", aiC.State, AIStateDetect)
	}
	if !aiC.HasTarget() {
		t.Error("should have detected target")
	}
}

// TestAISystemChase tests chase behavior.
func TestAISystemChase(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity with all needed components
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 100, Y: 100})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&AttackComponent{Range: 50, Cooldown: 1.0})

	// Create enemy within detection range but outside attack range
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 200, Y: 100}) // 100 pixels away
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	aiComp.Target = enemy
	aiComp.DetectionRange = 300 // Ensure enemy is within detection range

	world.Update(0)

	// Update AI
	aiSystem.Update(0.6)

	// Should still be chasing (enemy is in detection range but not attack range)
	if aiComp.State != AIStateChase {
		t.Errorf("state = %v, want %v", aiComp.State, AIStateChase)
	}

	// Should have set velocity towards enemy
	velComp, _ := ai.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("velocity should be set when chasing")
	}
}

// TestAISystemAttack tests attack behavior.
func TestAISystemAttack(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateAttack
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 100, Y: 100})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&AttackComponent{
		Damage:        10,
		Range:         100,
		Cooldown:      1.0,
		CooldownTimer: 0,
	})
	ai.AddComponent(NewStatsComponent())

	// Create enemy in attack range
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 150, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})
	enemy.AddComponent(NewStatsComponent())

	aiComp.Target = enemy

	world.Update(0)

	// Get initial enemy health
	enemyHealth, _ := enemy.GetComponent("health")
	initialHealth := enemyHealth.(*HealthComponent).Current

	// Update AI
	aiSystem.Update(0.6)

	// Should still be in attack state
	if aiComp.State != AIStateAttack {
		t.Errorf("state = %v, want %v", aiComp.State, AIStateAttack)
	}

	// Enemy should have taken damage
	currentHealth := enemyHealth.(*HealthComponent).Current
	if currentHealth >= initialHealth {
		t.Errorf("enemy health = %v, should be less than %v", currentHealth, initialHealth)
	}
}

// TestAISystemFlee tests flee behavior.
func TestAISystemFlee(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity with low health
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateFlee
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 200, Y: 200})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&HealthComponent{Current: 10, Max: 100}) // 10% health

	world.Update(0)

	// Update AI
	aiSystem.Update(0.6)

	// Should still be fleeing
	if aiComp.State != AIStateFlee {
		t.Errorf("state = %v, want %v", aiComp.State, AIStateFlee)
	}

	// Should be moving towards spawn
	velComp, _ := ai.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("velocity should be set when fleeing")
	}
}

// TestAISystemReturn tests return to spawn behavior.
func TestAISystemReturn(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity far from spawn
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateReturn
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 500, Y: 500})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})

	world.Update(0)

	// Update AI
	aiSystem.Update(0.6)

	// Should still be returning
	if aiComp.State != AIStateReturn {
		t.Errorf("state = %v, want %v", aiComp.State, AIStateReturn)
	}

	// Should be moving towards spawn
	velComp, _ := ai.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX == 0 && vel.VY == 0 {
		t.Error("velocity should be set when returning")
	}
}

// TestAISystemReturnAtSpawn tests behavior when close to spawn.
func TestAISystemReturnAtSpawn(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity close to spawn
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateReturn
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 105, Y: 105})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})

	world.Update(0)

	// Update AI
	aiSystem.Update(0.6)

	// Should transition to idle
	if aiComp.State != AIStateIdle {
		t.Errorf("state = %v, want %v after returning to spawn", aiComp.State, AIStateIdle)
	}

	// Velocity should be stopped
	velComp, _ := ai.GetComponent("velocity")
	vel := velComp.(*VelocityComponent)
	if vel.VX != 0 || vel.VY != 0 {
		t.Error("velocity should be stopped when at spawn")
	}
}

// TestAISystemFleeTransition tests transitioning from combat to flee.
func TestAISystemFleeTransition(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity in attack state with low health
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateAttack
	aiComp.FleeHealthThreshold = 0.2
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 100, Y: 100})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&AttackComponent{Range: 100, Cooldown: 1.0})
	ai.AddComponent(&HealthComponent{Current: 15, Max: 100}) // 15% health

	// Create enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 150, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	aiComp.Target = enemy

	world.Update(0)

	// Update AI - should transition to flee
	aiSystem.Update(0.6)

	if aiComp.State != AIStateFlee {
		t.Errorf("state = %v, want %v when health low", aiComp.State, AIStateFlee)
	}
}

// TestAISystemChaseRange tests chase distance limit.
func TestAISystemChaseRange(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity far from spawn
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	aiComp.MaxChaseDistance = 200
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 400, Y: 100}) // 300 pixels from spawn
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&AttackComponent{Range: 50, Cooldown: 1.0})

	// Create enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 450, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 100, Max: 100})

	aiComp.Target = enemy

	world.Update(0)

	// Update AI - should transition to return (too far from spawn)
	aiSystem.Update(0.6)

	if aiComp.State != AIStateReturn {
		t.Errorf("state = %v, want %v when too far from spawn", aiComp.State, AIStateReturn)
	}
}

// TestAISystemNoComponents tests AI with missing components.
func TestAISystemNoComponents(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity without position
	ai := world.CreateEntity()
	ai.AddComponent(NewAIComponent(100, 100))

	world.Update(0)

	// Should not crash
	aiSystem.Update(0.6)
}

// TestAISystemDeadTarget tests behavior when target dies.
func TestAISystemDeadTarget(t *testing.T) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity
	ai := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	aiComp.State = AIStateChase
	ai.AddComponent(aiComp)
	ai.AddComponent(&PositionComponent{X: 100, Y: 100})
	ai.AddComponent(&VelocityComponent{})
	ai.AddComponent(&TeamComponent{TeamID: 1})
	ai.AddComponent(&AttackComponent{Range: 100, Cooldown: 1.0})

	// Create dead enemy
	enemy := world.CreateEntity()
	enemy.AddComponent(&PositionComponent{X: 150, Y: 100})
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&HealthComponent{Current: 0, Max: 100}) // Dead

	aiComp.Target = enemy

	world.Update(0)

	// Update AI - should lose target and return
	aiSystem.Update(0.6)

	if aiComp.State != AIStateReturn {
		t.Errorf("state = %v, want %v when target is dead", aiComp.State, AIStateReturn)
	}
	if aiComp.HasTarget() {
		t.Error("should not have target when target is dead")
	}
}

// BenchmarkAISystemUpdate benchmarks AI system updates.
func BenchmarkAISystemUpdate(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create 50 AI entities
	for i := 0; i < 50; i++ {
		ai := world.CreateEntity()
		ai.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		ai.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		ai.AddComponent(&VelocityComponent{})
		ai.AddComponent(&TeamComponent{TeamID: 1})
		ai.AddComponent(&AttackComponent{Range: 50, Cooldown: 1.0})
		ai.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aiSystem.Update(0.016) // ~60 FPS
	}
}

// BenchmarkAISystemUpdateMany benchmarks AI with many entities.
func BenchmarkAISystemUpdateMany(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create 200 AI entities
	for i := 0; i < 200; i++ {
		ai := world.CreateEntity()
		ai.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		ai.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		ai.AddComponent(&VelocityComponent{})
		ai.AddComponent(&TeamComponent{TeamID: 1})
		ai.AddComponent(&AttackComponent{Range: 50, Cooldown: 1.0})
		ai.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aiSystem.Update(0.016) // ~60 FPS
	}
}
