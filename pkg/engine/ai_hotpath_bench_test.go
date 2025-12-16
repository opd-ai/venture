package engine

import (
	"testing"
)

// BenchmarkAISystem_IsValidTarget benchmarks the isValidTarget hot path
// which was optimized to use typed getters instead of map lookup + type assertion.
func BenchmarkAISystem_IsValidTarget(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 25, Y: 10})
	aiEntity.AddComponent(&TeamComponent{TeamID: 0})
	aiEntity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	pos := aiEntity.GetPosition()

	// Create target entity
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 30, Y: 15})
	target.AddComponent(&TeamComponent{TeamID: 1})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	var result bool
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = aiSystem.isValidTarget(target, aiEntity, pos, 50.0)
	}
	// Prevent optimization
	if !result {
		b.Log("Invalid target")
	}
}

// BenchmarkAISystem_ShouldFlee benchmarks the shouldFlee hot path
// which was optimized to use typed getter instead of map lookup + type assertion.
func BenchmarkAISystem_ShouldFlee(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity with health
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 25, Y: 10})
	aiEntity.AddComponent(&HealthComponent{Current: 30, Max: 100}) // Low health for flee check

	aiComp := &AIComponent{
		State:               AIStateChase,
		FleeHealthThreshold: 0.25,
	}
	aiEntity.AddComponent(aiComp)

	var result bool
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = aiSystem.shouldFlee(aiEntity, aiComp)
	}
	// Prevent optimization
	if result {
		b.Log("Should flee")
	}
}

// BenchmarkAISystem_GetTargetPosition benchmarks the getTargetPosition hot path
// which was optimized to use typed getter instead of map lookup + type assertion.
func BenchmarkAISystem_GetTargetPosition(b *testing.B) {
	world := NewWorld()
	aiSystem := NewAISystem(world)

	// Create AI entity
	aiEntity := world.CreateEntity()
	aiEntity.AddComponent(&PositionComponent{X: 25, Y: 10})

	// Create target entity
	target := world.CreateEntity()
	target.AddComponent(&PositionComponent{X: 30, Y: 15})

	aiComp := &AIComponent{
		State:  AIStateAttack,
		Target: target,
	}
	aiEntity.AddComponent(aiComp)

	var result *PositionComponent
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = aiSystem.getTargetPosition(aiEntity, aiComp)
	}
	// Prevent optimization
	if result == nil {
		b.Log("nil")
	}
}
