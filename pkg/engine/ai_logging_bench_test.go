package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// BenchmarkAISystem_Update_DebugEnabled benchmarks AI system update with debug logging enabled.
// This tests the hot path with debug-level logging to measure the impact of aiDebugEnabled caching.
func BenchmarkAISystem_Update_DebugEnabled(b *testing.B) {
	world := NewWorld()
	world.logger.Logger.SetLevel(logrus.DebugLevel)
	aiSystem := NewAISystem(world)

	// Create AI entities with all required components
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&TeamComponent{TeamID: 1})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	entities := world.GetEntities()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aiSystem.Update(entities, 0.016) // 60 FPS delta
	}
}

// BenchmarkAISystem_Update_DebugDisabled benchmarks AI system update with debug logging disabled.
// This provides a baseline comparison showing debug logging is effectively skipped.
func BenchmarkAISystem_Update_DebugDisabled(b *testing.B) {
	world := NewWorld()
	world.logger.Logger.SetLevel(logrus.InfoLevel)
	aiSystem := NewAISystem(world)

	// Create AI entities with all required components
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(NewAIComponent(float64(i*10), float64(i*10)))
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&TeamComponent{TeamID: 1})
		entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	entities := world.GetEntities()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aiSystem.Update(entities, 0.016) // 60 FPS delta
	}
}

// BenchmarkAIDebugFlagCheck_Cached benchmarks the cached debug flag check.
// This demonstrates the performance of using the cached aiDebugEnabled variable.
func BenchmarkAIDebugFlagCheck_Cached(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logger.WithField("system", "ai")
	SetAIDebugEnabled(entry)

	var result bool
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = aiDebugEnabled
	}
	// Prevent optimization
	if result {
		b.Log("debug enabled")
	}
}

// BenchmarkAIDebugFlagCheck_Reflection benchmarks the reflection-based log level check.
// This demonstrates the cost of calling GetLevel() on every check.
func BenchmarkAIDebugFlagCheck_Reflection(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logger.WithField("system", "ai")

	var result bool
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result = entry.Logger.GetLevel() >= logrus.DebugLevel
	}
	// Prevent optimization
	if result {
		b.Log("debug enabled")
	}
}

// BenchmarkAISystem_ProcessIdle_DebugEnabled benchmarks processIdle with debug logging.
func BenchmarkAISystem_ProcessIdle_DebugEnabled(b *testing.B) {
	world := NewWorld()
	world.logger.Logger.SetLevel(logrus.DebugLevel)
	aiSystem := NewAISystem(world)

	// Create AI entity
	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	entity.AddComponent(aiComp)
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)
	entity.AddComponent(&TeamComponent{TeamID: 1})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aiSystem.processIdle(entity, aiComp, pos)
	}
}

// BenchmarkAISystem_ProcessIdle_DebugDisabled benchmarks processIdle without debug logging.
func BenchmarkAISystem_ProcessIdle_DebugDisabled(b *testing.B) {
	world := NewWorld()
	world.logger.Logger.SetLevel(logrus.InfoLevel)
	aiSystem := NewAISystem(world)

	// Create AI entity
	entity := world.CreateEntity()
	aiComp := NewAIComponent(100, 100)
	entity.AddComponent(aiComp)
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)
	entity.AddComponent(&TeamComponent{TeamID: 1})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aiSystem.processIdle(entity, aiComp, pos)
	}
}
