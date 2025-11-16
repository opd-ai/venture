package network

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestSnapshotBuilder_V4Components tests serialization of V4 components
func TestSnapshotBuilder_V4Components(t *testing.T) {
	builder := NewSnapshotBuilder()
	entity := engine.NewEntity(123)

	// Add V4 components
	expressionComp := &engine.ExpressionComponent{
		ActiveExpression: engine.ExpressionWave,
		ExpressionTime:   1.5,
		Cooldown:         3.0,
	}
	entity.AddComponent(expressionComp)

	classComp := &engine.ClassProgressionComponent{
		Class:          engine.ClassWarrior,
		Level:          10,
		Experience:     5000,
		Specialization: engine.SpecializationBerserker,
	}
	entity.AddComponent(classComp)

	// Build snapshot
	timestamp := time.Now()
	snapshot := builder.BuildEntitySnapshot(entity, timestamp, 1)

	// Verify components are in snapshot
	if _, ok := snapshot.Components["expression"]; !ok {
		t.Error("Expected expression component in snapshot")
	}
	if _, ok := snapshot.Components["class_progression"]; !ok {
		t.Error("Expected class_progression component in snapshot")
	}

	// Apply snapshot to new entity
	newEntity := engine.NewEntity(456)
	newEntity.AddComponent(&engine.ExpressionComponent{})
	newEntity.AddComponent(&engine.ClassProgressionComponent{})

	if err := builder.ApplySnapshotToEntity(newEntity, snapshot); err != nil {
		t.Fatalf("Failed to apply snapshot: %v", err)
	}

	// Verify deserialization
	if exprComp, ok := newEntity.GetComponent("expression"); ok {
		expr := exprComp.(*engine.ExpressionComponent)
		if expr.ActiveExpression != engine.ExpressionWave {
			t.Errorf("Expected ActiveExpression=%v, got %v", engine.ExpressionWave, expr.ActiveExpression)
		}
		if expr.ExpressionTime != 1.5 {
			t.Errorf("Expected ExpressionTime=1.5, got %v", expr.ExpressionTime)
		}
		if expr.Cooldown != 3.0 {
			t.Errorf("Expected Cooldown=3.0, got %v", expr.Cooldown)
		}
	} else {
		t.Error("Expression component not found after deserialization")
	}

	if classCompApplied, ok := newEntity.GetComponent("class_progression"); ok {
		class := classCompApplied.(*engine.ClassProgressionComponent)
		if class.Class != engine.ClassWarrior {
			t.Errorf("Expected Class=%v, got %v", engine.ClassWarrior, class.Class)
		}
		if class.Level != 10 {
			t.Errorf("Expected Level=10, got %v", class.Level)
		}
		if class.Experience != 5000 {
			t.Errorf("Expected Experience=5000, got %v", class.Experience)
		}
		if class.Specialization != engine.SpecializationBerserker {
			t.Errorf("Expected Specialization=%v, got %v", engine.SpecializationBerserker, class.Specialization)
		}
	} else {
		t.Error("Class progression component not found after deserialization")
	}
}

// TestSnapshotBuilder_WorldSnapshotWithV4 tests world snapshot with V4 components
func TestSnapshotBuilder_WorldSnapshotWithV4(t *testing.T) {
	builder := NewSnapshotBuilder()

	// Create multiple entities with V4 components
	entities := make([]*engine.Entity, 3)
	for i := 0; i < 3; i++ {
		entity := engine.NewEntity(uint64(i + 1))
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 10), Y: float64(i * 20)})
		entity.AddComponent(&engine.ExpressionComponent{
			ActiveExpression: engine.ExpressionType(i),
			ExpressionTime:   float64(i),
			Cooldown:         float64(i * 2),
		})
		entities[i] = entity
	}

	// Build world snapshot
	timestamp := time.Now()
	worldSnapshot := builder.BuildWorldSnapshot(entities, timestamp, 1)

	// Verify all entities are in snapshot
	if len(worldSnapshot.Entities) != 3 {
		t.Errorf("Expected 3 entities in snapshot, got %d", len(worldSnapshot.Entities))
	}

	for i := uint64(1); i <= 3; i++ {
		if _, ok := worldSnapshot.Entities[i]; !ok {
			t.Errorf("Entity %d not found in world snapshot", i)
		}
	}
}

// BenchmarkSnapshotBuilder_V4Components benchmarks V4 component serialization
func BenchmarkSnapshotBuilder_V4Components(b *testing.B) {
	builder := NewSnapshotBuilder()
	entity := engine.NewEntity(123)

	entity.AddComponent(&engine.ExpressionComponent{
		ActiveExpression: engine.ExpressionWave,
		ExpressionTime:   1.5,
		Cooldown:         3.0,
	})
	entity.AddComponent(&engine.ClassProgressionComponent{
		Class:          engine.ClassWarrior,
		Level:          10,
		Experience:     5000,
		Specialization: engine.SpecializationBerserker,
	})

	timestamp := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildEntitySnapshot(entity, timestamp, uint32(i))
	}
}
