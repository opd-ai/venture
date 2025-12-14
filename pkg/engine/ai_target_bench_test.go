package engine

import (
	"math/rand"
	"testing"
)

// BenchmarkAISystem_FindNearestEnemy benchmarks the target finding hot path
func BenchmarkAISystem_FindNearestEnemy(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create 1000 entities with position and team components
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 50),
			Y: float64(i / 50),
		})
		entity.AddComponent(&TeamComponent{
			TeamID: i % 2, // Two teams
		})
	}

	// Create AI-controlled entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 25, Y: 10})
	aiEntity.AddComponent(&TeamComponent{TeamID: 0})
	aiEntity.AddComponent(&AIComponent{
		State:          AIStateIdle,
		DetectionRange: 50.0,
	})

	posComp, _ := aiEntity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	var result *Entity
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = aiSystem.findNearestEnemy(aiEntity, pos, 50.0)
	}
	// Prevent optimization
	if result != nil {
		b.Log("Found target")
	}
}

// BenchmarkAISystem_Update benchmarks the full AI system update
func BenchmarkAISystem_Update(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create 100 AI entities
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 10),
			Y: float64(i / 10),
		})
		entity.AddComponent(&TeamComponent{TeamID: 0})
		entity.AddComponent(&AIComponent{
			State:            AIStateIdle,
			DetectionRange:   20.0,
			DecisionInterval: 1.0,
			StateTimer:       1.0, // Trigger decision immediately
		})
		entities[i] = entity
	}

	// Create 500 potential targets
	for i := 0; i < 500; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{
			X: float64(i % 25),
			Y: float64(i / 25),
		})
		entity.AddComponent(&TeamComponent{TeamID: 1}) // Enemy team
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aiSystem.Update(entities, 0.016) // 60 FPS
	}
}

// BenchmarkAISystem_ExecuteCombatAttack benchmarks the combat attack execution path
// This specifically tests the optimization of reusing CombatSystem instead of creating new instances
func BenchmarkAISystem_ExecuteCombatAttack(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create attacker with required components
	attacker := world.CreateEntity()
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&VelocityComponent{})
	attacker.AddComponent(&TeamComponent{TeamID: 0})
	attacker.AddComponent(&AttackComponent{
		Damage:        10,
		Range:         50,
		Cooldown:      1.0,
		CooldownTimer: 0,
	})
	attacker.AddComponent(&HealthComponent{Current: 100, Max: 100})
	attacker.AddComponent(&StatsComponent{Attack: 10, Defense: 5})

	// Create AI component with target
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 10, Y: 10})
	target.AddComponent(&TeamComponent{TeamID: 1})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})
	target.AddComponent(&StatsComponent{Attack: 5, Defense: 5})

	aiComp := &AIComponent{
		State:  AIStateAttack,
		Target: target,
	}
	attacker.AddComponent(aiComp)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aiSystem.executeCombatAttack(attacker, aiComp)
	}
}

// BenchmarkNewCombatSystem_Allocation measures the allocation cost of creating new CombatSystem
// This demonstrates the cost that was eliminated by the optimization
func BenchmarkNewCombatSystem_Allocation(b *testing.B) {
	var cs *CombatSystem
	rng := rand.New(rand.NewSource(12345))
	_ = rng

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cs = NewCombatSystem(12345)
	}
	// Prevent optimization
	if cs == nil {
		b.Log("nil")
	}
}
