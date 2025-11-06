package hostplay

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

func TestNewStateBroadcaster(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	rate := 20

	broadcaster := NewStateBroadcaster(world, rate, logger)

	if broadcaster == nil {
		t.Fatal("NewStateBroadcaster returned nil")
	}
	if broadcaster.world != world {
		t.Error("world not set correctly")
	}
	if broadcaster.broadcastRate != rate {
		t.Errorf("broadcastRate = %d, want %d", broadcaster.broadcastRate, rate)
	}
}

func TestStateBroadcaster_ShouldBroadcast(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger) // 20 Hz = 50ms interval

	// Should broadcast immediately (first time)
	if !broadcaster.ShouldBroadcast() {
		t.Error("ShouldBroadcast() = false, want true on first call")
	}

	// Create a snapshot to update lastBroadcast
	broadcaster.CreateSnapshot()

	// Should not broadcast immediately after
	if broadcaster.ShouldBroadcast() {
		t.Error("ShouldBroadcast() = true, want false immediately after broadcast")
	}

	// Wait for interval
	time.Sleep(60 * time.Millisecond)

	// Should broadcast after interval
	if !broadcaster.ShouldBroadcast() {
		t.Error("ShouldBroadcast() = false, want true after interval")
	}
}

func TestStateBroadcaster_CreateSnapshot(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	// Create test entities
	entity1 := world.CreateEntity()
	entity1.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	entity1.AddComponent(&engine.VelocityComponent{VX: 10, VY: 20})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&engine.PositionComponent{X: 300, Y: 400})
	entity2.AddComponent(&engine.HealthComponent{Current: 75, Max: 100})

	// Force entity addition
	world.Update(0)

	snapshot, err := broadcaster.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	if snapshot == nil {
		t.Fatal("snapshot is nil")
	}

	if snapshot.Timestamp == 0 {
		t.Error("timestamp not set")
	}

	if len(snapshot.Entities) != 2 {
		t.Errorf("len(Entities) = %d, want 2", len(snapshot.Entities))
	}
}

func TestStateBroadcaster_SerializeSnapshot(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	world.Update(0)

	snapshot, err := broadcaster.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	data, err := broadcaster.SerializeSnapshot(snapshot)
	if err != nil {
		t.Fatalf("SerializeSnapshot() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("serialized data is empty")
	}

	// Verify JSON is valid
	var decoded WorldState
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("failed to unmarshal: %v", err)
	}

	if len(decoded.Entities) != 1 {
		t.Errorf("decoded len(Entities) = %d, want 1", len(decoded.Entities))
	}
}

func TestStateBroadcaster_Broadcast(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	world.Update(0)

	// First broadcast should succeed
	data, shouldBroadcast, err := broadcaster.Broadcast()
	if err != nil {
		t.Fatalf("Broadcast() error = %v", err)
	}
	if !shouldBroadcast {
		t.Error("shouldBroadcast = false, want true on first call")
	}
	if len(data) == 0 {
		t.Error("broadcast data is empty")
	}

	// Immediate second broadcast should be skipped
	_, shouldBroadcast, err = broadcaster.Broadcast()
	if err != nil {
		t.Fatalf("Broadcast() error = %v", err)
	}
	if shouldBroadcast {
		t.Error("shouldBroadcast = true, want false immediately after")
	}
}

func TestStateBroadcaster_SerializeEntityComponents(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 100, Y: 200})
	entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 20})
	entity.AddComponent(&engine.HealthComponent{Current: 75, Max: 100})
	entity.AddComponent(&engine.RotationComponent{Angle: 1.57})
	world.Update(0)

	snapshot, err := broadcaster.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	if len(snapshot.Entities) != 1 {
		t.Fatalf("len(Entities) = %d, want 1", len(snapshot.Entities))
	}

	entityState := snapshot.Entities[0]

	// Check all components serialized
	if entityState.Position == nil {
		t.Error("Position not serialized")
	} else {
		if entityState.Position.X != 100 || entityState.Position.Y != 200 {
			t.Errorf("Position = (%f, %f), want (100, 200)", entityState.Position.X, entityState.Position.Y)
		}
	}

	if entityState.Velocity == nil {
		t.Error("Velocity not serialized")
	} else {
		if entityState.Velocity.VX != 10 || entityState.Velocity.VY != 20 {
			t.Errorf("Velocity = (%f, %f), want (10, 20)", entityState.Velocity.VX, entityState.Velocity.VY)
		}
	}

	if entityState.Health == nil {
		t.Error("Health not serialized")
	} else {
		if entityState.Health.Current != 75 || entityState.Health.Max != 100 {
			t.Errorf("Health = (%f, %f), want (75, 100)", entityState.Health.Current, entityState.Health.Max)
		}
	}

	if entityState.Rotation == nil {
		t.Error("Rotation not serialized")
	} else {
		if entityState.Rotation.Angle != 1.57 {
			t.Errorf("Rotation.Angle = %f, want 1.57", entityState.Rotation.Angle)
		}
	}
}

func TestStateBroadcaster_SetBroadcastRate(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	// Test valid rate
	broadcaster.SetBroadcastRate(30)
	if broadcaster.GetBroadcastRate() != 30 {
		t.Errorf("GetBroadcastRate() = %d, want 30", broadcaster.GetBroadcastRate())
	}

	// Test invalid rates (should be ignored)
	broadcaster.SetBroadcastRate(0)
	if broadcaster.GetBroadcastRate() != 30 {
		t.Error("rate changed to invalid value 0")
	}

	broadcaster.SetBroadcastRate(100)
	if broadcaster.GetBroadcastRate() != 30 {
		t.Error("rate changed to invalid value 100")
	}
}

func TestStateBroadcaster_EmptyWorld(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	snapshot, err := broadcaster.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	if len(snapshot.Entities) != 0 {
		t.Errorf("len(Entities) = %d, want 0 for empty world", len(snapshot.Entities))
	}
}

func TestStateBroadcaster_EntityWithoutPosition(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	broadcaster := NewStateBroadcaster(world, 20, logger)

	// Create entity with only velocity (no position)
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{VX: 10, VY: 20})
	world.Update(0)

	snapshot, err := broadcaster.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	// Entity without position should not be included
	if len(snapshot.Entities) != 0 {
		t.Errorf("len(Entities) = %d, want 0 (entity without position)", len(snapshot.Entities))
	}
}
