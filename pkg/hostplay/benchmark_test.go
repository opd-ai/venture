package hostplay

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// =============================================================================
// Snapshot Serialization Benchmarks
// =============================================================================

// BenchmarkCreateSnapshot measures the performance of creating world state snapshots.
// This is a hot-path operation called every tick to capture entity states.
func BenchmarkCreateSnapshot(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel) // Disable logging overhead
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	// Create test entities with typical component distribution
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
		if i%2 == 0 {
			entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
		}
		if i%3 == 0 {
			entity.AddComponent(&engine.RotationComponent{Angle: float64(i) * 0.1})
		}
	}
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tp.Advance(100 * time.Millisecond)
		_, err := broadcaster.CreateSnapshot()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateSnapshot_EntityCount benchmarks CreateSnapshot with varying entity counts.
func BenchmarkCreateSnapshot_EntityCount(b *testing.B) {
	entityCounts := []int{10, 50, 100, 200, 500}

	for _, count := range entityCounts {
		name := fmt.Sprintf("%d_entities", count)
		b.Run(name, func(b *testing.B) {
			world := engine.NewWorld()
			logger := logrus.NewEntry(logrus.New())
			logger.Logger.SetLevel(logrus.ErrorLevel)
			fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
			tp := &MockTimeProvider{fixedTime: fixedTime}
			broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

			for i := 0; i < count; i++ {
				entity := world.CreateEntity()
				entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
				entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
				entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
			}
			world.Update(0)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				tp.Advance(100 * time.Millisecond)
				_, err := broadcaster.CreateSnapshot()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkSerializeSnapshot measures JSON serialization performance of WorldState.
// This is the second hot-path step after CreateSnapshot.
func BenchmarkSerializeSnapshot(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel)
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
		entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
		entity.AddComponent(&engine.RotationComponent{Angle: float64(i) * 0.1})
	}
	world.Update(0)

	snapshot, _ := broadcaster.CreateSnapshot()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := broadcaster.SerializeSnapshot(snapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSerializeSnapshot_Size benchmarks serialization with different snapshot sizes.
func BenchmarkSerializeSnapshot_Size(b *testing.B) {
	sizes := []int{10, 50, 100, 200}

	for _, size := range sizes {
		name := fmt.Sprintf("%d_entities", size)
		b.Run(name, func(b *testing.B) {
			world := engine.NewWorld()
			logger := logrus.NewEntry(logrus.New())
			logger.Logger.SetLevel(logrus.ErrorLevel)
			fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
			tp := &MockTimeProvider{fixedTime: fixedTime}
			broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

			for i := 0; i < size; i++ {
				entity := world.CreateEntity()
				entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
				entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
				entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
			}
			world.Update(0)

			snapshot, _ := broadcaster.CreateSnapshot()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := broadcaster.SerializeSnapshot(snapshot)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// =============================================================================
// Delta Snapshot Benchmarks
// =============================================================================

// BenchmarkCreateDeltaSnapshot measures delta snapshot creation performance.
// Delta snapshots only include changed entities, reducing network bandwidth.
func BenchmarkCreateDeltaSnapshot(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel)
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	// Create 100 entities
	entities := make([]*engine.Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
		entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
		entities[i] = entity
	}
	world.Update(0)

	previousSnapshot, _ := broadcaster.CreateSnapshot()
	tp.Advance(100 * time.Millisecond)

	// Modify 10% of entities each frame (typical scenario)
	for i := 0; i < 10; i++ {
		if pos, ok := entities[i].GetComponent("position"); ok {
			if p, ok := pos.(*engine.PositionComponent); ok {
				p.X += 1.0
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tp.Advance(100 * time.Millisecond)
		_, err := broadcaster.CreateDeltaSnapshot(previousSnapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCreateDeltaSnapshot_ChangeRate benchmarks delta snapshots with varying change rates.
func BenchmarkCreateDeltaSnapshot_ChangeRate(b *testing.B) {
	changeRates := []int{0, 10, 25, 50, 100} // Percentage of entities changed

	for _, rate := range changeRates {
		name := fmt.Sprintf("%dpct_changed", rate)
		b.Run(name, func(b *testing.B) {
			world := engine.NewWorld()
			logger := logrus.NewEntry(logrus.New())
			logger.Logger.SetLevel(logrus.ErrorLevel)
			fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
			tp := &MockTimeProvider{fixedTime: fixedTime}
			broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

			entities := make([]*engine.Entity, 100)
			for i := 0; i < 100; i++ {
				entity := world.CreateEntity()
				entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
				entities[i] = entity
			}
			world.Update(0)

			previousSnapshot, _ := broadcaster.CreateSnapshot()
			tp.Advance(100 * time.Millisecond)

			// Modify 'rate' percent of entities
			numChanges := rate
			for i := 0; i < numChanges; i++ {
				if pos, ok := entities[i].GetComponent("position"); ok {
					if p, ok := pos.(*engine.PositionComponent); ok {
						p.X += 1.0
					}
				}
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				tp.Advance(100 * time.Millisecond)
				_, err := broadcaster.CreateDeltaSnapshot(previousSnapshot)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// =============================================================================
// State Broadcast Benchmarks
// =============================================================================

// BenchmarkBroadcast measures the full broadcast pipeline (snapshot + serialization).
func BenchmarkBroadcast(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel)
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	// Create 50 entities (typical game scenario)
	for i := 0; i < 50; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(&engine.VelocityComponent{VX: float64(i % 10), VY: float64(i % 5)})
		entity.AddComponent(&engine.HealthComponent{Current: 100, Max: 100})
	}
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tp.Advance(100 * time.Millisecond)
		data, shouldBroadcast, err := broadcaster.Broadcast()
		if err != nil {
			b.Fatal(err)
		}
		if !shouldBroadcast {
			b.Fatal("expected broadcast")
		}
		if len(data) == 0 {
			b.Fatal("empty broadcast data")
		}
	}
}

// =============================================================================
// Component Serialization Benchmarks
// =============================================================================

// BenchmarkSerializeEntity measures single entity serialization performance.
func BenchmarkSerializeEntity(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel)
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 20})
	entity.AddComponent(&engine.HealthComponent{Current: 75, Max: 100})
	entity.AddComponent(&engine.RotationComponent{Angle: 1.57})
	world.Update(0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = broadcaster.serializeEntity(entity)
	}
}

// BenchmarkEntityStateComparison measures entity state comparison performance
// used in delta snapshot generation.
func BenchmarkEntityStateComparison(b *testing.B) {
	prev := &EntityState{
		ID: 1,
		Position: &PositionState{X: 100, Y: 200},
		Velocity: &VelocityState{VX: 10, VY: 20},
		Health:   &HealthState{Current: 75, Max: 100},
		Rotation: &RotationState{Angle: 1.57},
	}

	current := &EntityState{
		ID: 1,
		Position: &PositionState{X: 101, Y: 200}, // Position changed
		Velocity: &VelocityState{VX: 10, VY: 20},
		Health:   &HealthState{Current: 75, Max: 100},
		Rotation: &RotationState{Angle: 1.57},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = entityStateChanged(prev, current)
	}
}

// BenchmarkEntityStateComparison_NoChange measures comparison when nothing changed.
func BenchmarkEntityStateComparison_NoChange(b *testing.B) {
	prev := &EntityState{
		ID: 1,
		Position: &PositionState{X: 100, Y: 200},
		Velocity: &VelocityState{VX: 10, VY: 20},
		Health:   &HealthState{Current: 75, Max: 100},
		Rotation: &RotationState{Angle: 1.57},
	}

	current := &EntityState{
		ID: 1,
		Position: &PositionState{X: 100, Y: 200}, // Same
		Velocity: &VelocityState{VX: 10, VY: 20},
		Health:   &HealthState{Current: 75, Max: 100},
		Rotation: &RotationState{Angle: 1.57},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = entityStateChanged(prev, current)
	}
}

// =============================================================================
// JSON Marshal/Unmarshal Benchmarks
// =============================================================================

// BenchmarkWorldStateJSONMarshal measures raw JSON marshaling performance.
func BenchmarkWorldStateJSONMarshal(b *testing.B) {
	snapshot := &WorldState{
		Timestamp: 1708689600000,
		Entities:  make([]EntityState, 100),
	}

	for i := 0; i < 100; i++ {
		snapshot.Entities[i] = EntityState{
			ID: uint64(i),
			Position: &PositionState{X: float64(i * 10), Y: float64(i * 10)},
			Velocity: &VelocityState{VX: float64(i % 10), VY: float64(i % 5)},
			Health:   &HealthState{Current: 100, Max: 100},
			Rotation: &RotationState{Angle: float64(i) * 0.1},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(snapshot)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWorldStateJSONUnmarshal measures JSON unmarshaling performance.
func BenchmarkWorldStateJSONUnmarshal(b *testing.B) {
	snapshot := &WorldState{
		Timestamp: 1708689600000,
		Entities:  make([]EntityState, 100),
	}

	for i := 0; i < 100; i++ {
		snapshot.Entities[i] = EntityState{
			ID: uint64(i),
			Position: &PositionState{X: float64(i * 10), Y: float64(i * 10)},
			Velocity: &VelocityState{VX: float64(i % 10), VY: float64(i % 5)},
			Health:   &HealthState{Current: 100, Max: 100},
			Rotation: &RotationState{Angle: float64(i) * 0.1},
		}
	}

	data, _ := json.Marshal(snapshot)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var decoded WorldState
		err := json.Unmarshal(data, &decoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// ShouldBroadcast Benchmarks
// =============================================================================

// BenchmarkShouldBroadcast measures the rate-limiting check performance.
func BenchmarkShouldBroadcast(b *testing.B) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	logger.Logger.SetLevel(logrus.ErrorLevel)
	fixedTime := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	tp := &MockTimeProvider{fixedTime: fixedTime}
	broadcaster := NewStateBroadcasterWithTimeProvider(world, 20, logger, tp)

	// Pre-populate lastBroadcast
	broadcaster.CreateSnapshot()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = broadcaster.ShouldBroadcast()
	}
}
