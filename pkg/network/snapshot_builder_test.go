// Package network provides snapshot building tests.
package network

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TestBuildEntitySnapshot_CoreComponents tests snapshot building with core components.
func TestBuildEntitySnapshot_CoreComponents(t *testing.T) {
	builder := NewSnapshotBuilder()
	entity := &engine.Entity{ID: 42}

	// Add core components
	entity.AddComponent(&engine.PositionComponent{X: 100.5, Y: 200.5})
	entity.AddComponent(&engine.VelocityComponent{VX: 10.0, VY: -5.0})
	entity.AddComponent(&engine.HealthComponent{Current: 80.0, Max: 100.0})
	entity.AddComponent(&engine.StatsComponent{Attack: 50.0, Defense: 30.0, MagicPower: 40.0})
	entity.AddComponent(&engine.TeamComponent{TeamID: 1})
	entity.AddComponent(&engine.ExperienceComponent{Level: 5, CurrentXP: 250})

	now := time.Now()
	snapshot := builder.BuildEntitySnapshot(entity, now, 123)

	if snapshot.EntityID != 42 {
		t.Errorf("EntityID = %d, want 42", snapshot.EntityID)
	}

	if snapshot.Sequence != 123 {
		t.Errorf("Sequence = %d, want 123", snapshot.Sequence)
	}

	if snapshot.Position.X != 100.5 || snapshot.Position.Y != 200.5 {
		t.Errorf("Position = (%f, %f), want (100.5, 200.5)", snapshot.Position.X, snapshot.Position.Y)
	}

	if snapshot.Velocity.VX != 10.0 || snapshot.Velocity.VY != -5.0 {
		t.Errorf("Velocity = (%f, %f), want (10.0, -5.0)", snapshot.Velocity.VX, snapshot.Velocity.VY)
	}

	// Verify all component data is present
	requiredComponents := []string{"position", "velocity", "health", "stats", "team", "level"}
	for _, compType := range requiredComponents {
		if _, ok := snapshot.Components[compType]; !ok {
			t.Errorf("Component %s missing from snapshot", compType)
		}
	}
}

// TestBuildEntitySnapshot_V4Components tests snapshot building with V4.0 components.
func TestBuildEntitySnapshot_V4Components(t *testing.T) {
	builder := NewSnapshotBuilder()
	entity := &engine.Entity{ID: 99}

	// Add position (required for snapshot)
	entity.AddComponent(&engine.PositionComponent{X: 50.0, Y: 75.0})

	// Add V4.0 components
	vehicleComp := engine.NewVehicleComponent(engine.VehicleMount)
	vehicleComp.Speed = 120.0
	vehicleComp.Durability = 80.0
	entity.AddComponent(vehicleComp)

	companionComp := &engine.CompanionComponent{
		OwnerID:       100,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       75.0,
		Level:         3,
		Behavior:      engine.BehaviorAggressive,
	}
	entity.AddComponent(companionComp)

	mountComp := engine.NewMountComponent(50, 0.0, -16.0)
	entity.AddComponent(mountComp)

	achievementComp := &engine.AchievementComponent{
		Achievements:    []engine.Achievement{{Type: engine.AchievementFirstExpression}},
		ExpressionCount: 10,
	}
	entity.AddComponent(achievementComp)

	expressionComp := &engine.ExpressionComponent{
		ActiveExpression: engine.ExpressionWave,
		ExpressionTime:   0.5,
		Cooldown:         2.5,
	}
	entity.AddComponent(expressionComp)

	snapshot := builder.BuildEntitySnapshot(entity, time.Now(), 456)

	// Verify V4 components are serialized
	v4Components := []string{"vehicle", "companion", "mount", "achievement", "expression"}
	for _, compType := range v4Components {
		if _, ok := snapshot.Components[compType]; !ok {
			t.Errorf("V4 component %s missing from snapshot", compType)
		}
	}

	// Verify component data is non-empty
	if len(snapshot.Components["vehicle"]) == 0 {
		t.Error("Vehicle component data is empty")
	}
	if len(snapshot.Components["companion"]) == 0 {
		t.Error("Companion component data is empty")
	}
}

// TestApplySnapshotToEntity_CoreComponents tests applying snapshot to entity.
func TestApplySnapshotToEntity_CoreComponents(t *testing.T) {
	builder := NewSnapshotBuilder()

	// Create entity with components
	entity := &engine.Entity{ID: 42}
	entity.AddComponent(&engine.PositionComponent{X: 0.0, Y: 0.0})
	entity.AddComponent(&engine.VelocityComponent{VX: 0.0, VY: 0.0})
	entity.AddComponent(&engine.HealthComponent{Current: 50.0, Max: 100.0})
	entity.AddComponent(&engine.StatsComponent{Attack: 20.0, Defense: 10.0, MagicPower: 15.0})
	entity.AddComponent(&engine.TeamComponent{TeamID: 0})
	entity.AddComponent(&engine.ExperienceComponent{Level: 1, CurrentXP: 0})

	// Build snapshot with different values
	snapshot := EntitySnapshot{
		EntityID:  42,
		Timestamp: time.Now(),
		Sequence:  789,
		Position:  Position{X: 150.5, Y: 250.5},
		Velocity:  Velocity{VX: 20.0, VY: -10.0},
		Components: map[string][]byte{
			"position": builder.serializer.SerializePosition(150.5, 250.5),
			"velocity": builder.serializer.SerializeVelocity(20.0, -10.0),
			"health":   builder.serializer.SerializeHealth(90.0, 120.0),
			"stats":    builder.serializer.SerializeStats(60.0, 40.0, 50.0),
			"team":     builder.serializer.SerializeTeam(2),
			"level":    builder.serializer.SerializeLevel(10, 500),
		},
	}

	// Apply snapshot
	err := builder.ApplySnapshotToEntity(entity, snapshot)
	if err != nil {
		t.Fatalf("ApplySnapshotToEntity failed: %v", err)
	}

	// Verify updated values
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*engine.PositionComponent)
	if pos.X != 150.5 || pos.Y != 250.5 {
		t.Errorf("Position = (%f, %f), want (150.5, 250.5)", pos.X, pos.Y)
	}

	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*engine.VelocityComponent)
	if vel.VX != 20.0 || vel.VY != -10.0 {
		t.Errorf("Velocity = (%f, %f), want (20.0, -10.0)", vel.VX, vel.VY)
	}

	healthComp, _ := entity.GetComponent("health")
	health := healthComp.(*engine.HealthComponent)
	if health.Current != 90.0 || health.Max != 120.0 {
		t.Errorf("Health = (%f, %f), want (90.0, 120.0)", health.Current, health.Max)
	}

	statsComp, _ := entity.GetComponent("stats")
	stats := statsComp.(*engine.StatsComponent)
	if stats.Attack != 60.0 || stats.Defense != 40.0 || stats.MagicPower != 50.0 {
		t.Errorf("Stats = (%f, %f, %f), want (60.0, 40.0, 50.0)", stats.Attack, stats.Defense, stats.MagicPower)
	}

	teamComp, _ := entity.GetComponent("team")
	team := teamComp.(*engine.TeamComponent)
	if team.TeamID != 2 {
		t.Errorf("TeamID = %d, want 2", team.TeamID)
	}

	expComp, _ := entity.GetComponent("experience")
	exp := expComp.(*engine.ExperienceComponent)
	if exp.Level != 10 || exp.CurrentXP != 500 {
		t.Errorf("Level/XP = (%d, %d), want (10, 500)", exp.Level, exp.CurrentXP)
	}
}

// TestApplySnapshotToEntity_V4Components tests applying V4 component data.
func TestApplySnapshotToEntity_V4Components(t *testing.T) {
	builder := NewSnapshotBuilder()

	// Create source entity with V4 components
	sourceEntity := &engine.Entity{ID: 100}
	sourceEntity.AddComponent(&engine.PositionComponent{X: 50.0, Y: 75.0})

	vehicleComp := engine.NewVehicleComponent(engine.VehicleCart)
	vehicleComp.Speed = 150.0
	vehicleComp.Durability = 95.0
	sourceEntity.AddComponent(vehicleComp)

	companionComp := &engine.CompanionComponent{
		OwnerID:       200,
		CompanionType: engine.CompanionTypeRobot,
		Loyalty:       85.0,
		Level:         5,
		Behavior:      engine.BehaviorDefensive,
	}
	sourceEntity.AddComponent(companionComp)

	// Build snapshot from source
	snapshot := builder.BuildEntitySnapshot(sourceEntity, time.Now(), 999)

	// Create target entity with default V4 components
	targetEntity := &engine.Entity{ID: 100}
	targetEntity.AddComponent(&engine.PositionComponent{X: 0.0, Y: 0.0})
	targetEntity.AddComponent(engine.NewVehicleComponent(engine.VehicleMount))
	targetEntity.AddComponent(&engine.CompanionComponent{})

	// Apply snapshot
	err := builder.ApplySnapshotToEntity(targetEntity, snapshot)
	if err != nil {
		t.Fatalf("ApplySnapshotToEntity failed: %v", err)
	}

	// Verify vehicle component was updated
	targetVehicleComp, _ := targetEntity.GetComponent("vehicle")
	targetVehicle := targetVehicleComp.(*engine.VehicleComponent)
	if targetVehicle.VehicleType != engine.VehicleCart {
		t.Errorf("VehicleType = %d, want %d", targetVehicle.VehicleType, engine.VehicleCart)
	}
	if targetVehicle.Speed != 150.0 {
		t.Errorf("Vehicle Speed = %f, want 150.0", targetVehicle.Speed)
	}
	if targetVehicle.Durability != 95.0 {
		t.Errorf("Vehicle Durability = %f, want 95.0", targetVehicle.Durability)
	}

	// Verify companion component was updated
	targetCompanionComp, _ := targetEntity.GetComponent("companion")
	targetCompanion := targetCompanionComp.(*engine.CompanionComponent)
	if targetCompanion.OwnerID != 200 {
		t.Errorf("Companion OwnerID = %d, want 200", targetCompanion.OwnerID)
	}
	if targetCompanion.CompanionType != engine.CompanionTypeRobot {
		t.Errorf("CompanionType = %d, want %d", targetCompanion.CompanionType, engine.CompanionTypeRobot)
	}
	if targetCompanion.Loyalty != 85.0 {
		t.Errorf("Companion Loyalty = %f, want 85.0", targetCompanion.Loyalty)
	}
	if targetCompanion.Level != 5 {
		t.Errorf("Companion Level = %d, want 5", targetCompanion.Level)
	}
}

// TestBuildWorldSnapshot tests building a complete world snapshot.
func TestBuildWorldSnapshot(t *testing.T) {
	builder := NewSnapshotBuilder()

	// Create multiple entities
	entities := []*engine.Entity{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	for _, entity := range entities {
		entity.AddComponent(&engine.PositionComponent{X: float64(entity.ID * 10), Y: float64(entity.ID * 20)})
		entity.AddComponent(&engine.HealthComponent{Current: 100.0, Max: 100.0})
	}

	now := time.Now()
	worldSnapshot := builder.BuildWorldSnapshot(entities, now, 555)

	if worldSnapshot.Sequence != 555 {
		t.Errorf("WorldSnapshot Sequence = %d, want 555", worldSnapshot.Sequence)
	}

	if len(worldSnapshot.Entities) != 3 {
		t.Errorf("WorldSnapshot has %d entities, want 3", len(worldSnapshot.Entities))
	}

	// Verify each entity snapshot
	for _, entity := range entities {
		entitySnap, exists := worldSnapshot.Entities[entity.ID]
		if !exists {
			t.Errorf("Entity %d missing from world snapshot", entity.ID)
			continue
		}

		expectedX := float64(entity.ID * 10)
		expectedY := float64(entity.ID * 20)
		if entitySnap.Position.X != expectedX || entitySnap.Position.Y != expectedY {
			t.Errorf("Entity %d position = (%f, %f), want (%f, %f)",
				entity.ID, entitySnap.Position.X, entitySnap.Position.Y, expectedX, expectedY)
		}
	}
}

// BenchmarkBuildEntitySnapshot benchmarks snapshot creation.
func BenchmarkBuildEntitySnapshot(b *testing.B) {
	builder := NewSnapshotBuilder()
	entity := &engine.Entity{ID: 42}
	entity.AddComponent(&engine.PositionComponent{X: 100.5, Y: 200.5})
	entity.AddComponent(&engine.VelocityComponent{VX: 10.0, VY: -5.0})
	entity.AddComponent(&engine.HealthComponent{Current: 80.0, Max: 100.0})
	entity.AddComponent(&engine.StatsComponent{Attack: 50.0, Defense: 30.0, MagicPower: 40.0})
	entity.AddComponent(engine.NewVehicleComponent(engine.VehicleMount))
	entity.AddComponent(&engine.CompanionComponent{OwnerID: 100, Loyalty: 75.0})

	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.BuildEntitySnapshot(entity, now, uint32(i))
	}
}

// BenchmarkApplySnapshotToEntity benchmarks snapshot application.
func BenchmarkApplySnapshotToEntity(b *testing.B) {
	builder := NewSnapshotBuilder()

	entity := &engine.Entity{ID: 42}
	entity.AddComponent(&engine.PositionComponent{X: 0.0, Y: 0.0})
	entity.AddComponent(&engine.VelocityComponent{VX: 0.0, VY: 0.0})
	entity.AddComponent(&engine.HealthComponent{Current: 50.0, Max: 100.0})
	entity.AddComponent(&engine.StatsComponent{Attack: 20.0, Defense: 10.0, MagicPower: 15.0})

	snapshot := EntitySnapshot{
		EntityID: 42,
		Position: Position{X: 150.5, Y: 250.5},
		Velocity: Velocity{VX: 20.0, VY: -10.0},
		Components: map[string][]byte{
			"position": builder.serializer.SerializePosition(150.5, 250.5),
			"velocity": builder.serializer.SerializeVelocity(20.0, -10.0),
			"health":   builder.serializer.SerializeHealth(90.0, 120.0),
			"stats":    builder.serializer.SerializeStats(60.0, 40.0, 50.0),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.ApplySnapshotToEntity(entity, snapshot)
	}
}
