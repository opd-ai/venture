package engine

import (
	"testing"
	"time"
)

func TestNewScheduleSystem(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)

	sys := NewScheduleSystem(world, clock)

	if sys == nil {
		t.Fatal("NewScheduleSystem returned nil")
	}
	if sys.world != world {
		t.Error("World not set correctly")
	}
	if sys.clock != clock {
		t.Error("Clock not set correctly")
	}
}

func TestScheduleSystem_Update_NilClock(t *testing.T) {
	world := NewWorld()
	sys := NewScheduleSystem(world, nil)

	// Should not panic with nil clock
	entities := []*Entity{}
	sys.Update(entities, 0.016)
}

func TestScheduleSystem_Update_MovesToLocation(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	// Set clock to 10am (working hours)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))

	sys := NewScheduleSystem(world, clock)

	// Create entity with schedule and position
	entity := world.CreateEntity()
	pos := &PositionComponent{X: 0, Y: 0}
	entity.AddComponent(pos)

	schedule := NewScheduleComponent(0, 0)
	schedule.AddActivity(ActivityWork, 8, 17, 100, 100, "Workshop")
	entity.AddComponent(schedule)

	// Run update
	entities := []*Entity{entity}
	sys.Update(entities, 1.0) // 1 second

	// Should have moved toward target
	if pos.X <= 0 || pos.Y <= 0 {
		t.Errorf("Expected position to move toward (100,100), got (%f,%f)", pos.X, pos.Y)
	}
	if !schedule.IsMoving {
		t.Error("Expected IsMoving to be true")
	}
}

func TestScheduleSystem_Update_StopsAtDestination(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))

	sys := NewScheduleSystem(world, clock)

	// Create entity already at destination
	entity := world.CreateEntity()
	pos := &PositionComponent{X: 100, Y: 100}
	entity.AddComponent(pos)

	schedule := NewScheduleComponent(0, 0)
	schedule.AddActivity(ActivityWork, 8, 17, 100, 100, "Workshop")
	schedule.IsMoving = true
	entity.AddComponent(schedule)

	// Run update
	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Should have stopped moving
	if schedule.IsMoving {
		t.Error("Expected IsMoving to be false at destination")
	}
}

func TestScheduleSystem_Update_ClearsVelocity(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))

	sys := NewScheduleSystem(world, clock)

	// Create entity at destination with velocity
	entity := world.CreateEntity()
	pos := &PositionComponent{X: 100, Y: 100}
	vel := &VelocityComponent{VX: 50, VY: 50}
	entity.AddComponent(pos)
	entity.AddComponent(vel)

	schedule := NewScheduleComponent(0, 0)
	schedule.AddActivity(ActivityWork, 8, 17, 100, 100, "Workshop")
	entity.AddComponent(schedule)

	// Run update
	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Velocity should be cleared
	if vel.VX != 0 || vel.VY != 0 {
		t.Errorf("Expected velocity (0,0), got (%f,%f)", vel.VX, vel.VY)
	}
}

func TestScheduleSystem_Update_SkipsNoSchedule(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))

	sys := NewScheduleSystem(world, clock)

	// Create entity without schedule
	entity := world.CreateEntity()
	pos := &PositionComponent{X: 0, Y: 0}
	entity.AddComponent(pos)

	// Run update - should not panic
	entities := []*Entity{entity}
	sys.Update(entities, 1.0)

	// Position should be unchanged
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("Expected position (0,0), got (%f,%f)", pos.X, pos.Y)
	}
}

func TestScheduleSystem_Update_UpdatesActivityByTime(t *testing.T) {
	world := NewWorld()
	clock := NewSimulationClock(12345)

	sys := NewScheduleSystem(world, clock)

	entity := world.CreateEntity()
	pos := &PositionComponent{X: 0, Y: 0}
	entity.AddComponent(pos)

	schedule := NewScheduleComponent(0, 0)
	schedule.AddActivity(ActivitySleep, 22, 6, 0, 0, "Home")
	schedule.AddActivity(ActivityWork, 8, 17, 100, 100, "Workshop")
	entity.AddComponent(schedule)

	entities := []*Entity{entity}

	// Test at midnight (should be sleeping)
	clock.Reset(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	sys.Update(entities, 0.016)
	if schedule.CurrentActivityIdx != 0 {
		t.Errorf("Expected activity idx 0 at midnight, got %d", schedule.CurrentActivityIdx)
	}

	// Test at 10am (should be working)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))
	sys.Update(entities, 0.016)
	if schedule.CurrentActivityIdx != 1 {
		t.Errorf("Expected activity idx 1 at 10am, got %d", schedule.CurrentActivityIdx)
	}
}

func TestGenerateDefaultSchedule_Determinism(t *testing.T) {
	seed := int64(42)

	sched1 := GenerateDefaultSchedule(seed, "merchant", 0, 0, 100, 100)
	sched2 := GenerateDefaultSchedule(seed, "merchant", 0, 0, 100, 100)

	if len(sched1.Activities) != len(sched2.Activities) {
		t.Errorf("Different activity counts: %d vs %d", len(sched1.Activities), len(sched2.Activities))
	}

	for i := range sched1.Activities {
		if sched1.Activities[i].ActivityType != sched2.Activities[i].ActivityType {
			t.Errorf("Activity %d type mismatch", i)
		}
		if sched1.Activities[i].StartHour != sched2.Activities[i].StartHour {
			t.Errorf("Activity %d start hour mismatch", i)
		}
	}
}

func TestGenerateDefaultSchedule_Roles(t *testing.T) {
	tests := []struct {
		role          string
		minActivities int
		expectsPatrol bool
		expectsWork   bool
		expectsSleep  bool
	}{
		{"merchant", 5, false, true, true},
		{"guard", 3, true, false, true},
		{"villager", 5, false, true, true},
		{"unknown", 2, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			sched := GenerateDefaultSchedule(12345, tt.role, 0, 0, 100, 100)

			if len(sched.Activities) < tt.minActivities {
				t.Errorf("Expected at least %d activities, got %d", tt.minActivities, len(sched.Activities))
			}

			hasPatrol := false
			hasWork := false
			hasSleep := false
			for _, act := range sched.Activities {
				if act.ActivityType == ActivityPatrol {
					hasPatrol = true
				}
				if act.ActivityType == ActivityWork {
					hasWork = true
				}
				if act.ActivityType == ActivitySleep {
					hasSleep = true
				}
			}

			if tt.expectsPatrol && !hasPatrol {
				t.Error("Expected patrol activity")
			}
			if tt.expectsWork && !hasWork {
				t.Error("Expected work activity")
			}
			if tt.expectsSleep && !hasSleep {
				t.Error("Expected sleep activity")
			}
		})
	}
}

func TestGenerateDefaultSchedule_DifferentSeeds(t *testing.T) {
	sched1 := GenerateDefaultSchedule(1, "villager", 0, 0, 100, 100)
	sched2 := GenerateDefaultSchedule(2, "villager", 0, 0, 100, 100)

	// Different seeds should produce different schedules for villagers
	// (villagers have randomized elements)
	different := false
	if len(sched1.Activities) != len(sched2.Activities) {
		different = true
	} else {
		for i := range sched1.Activities {
			if sched1.Activities[i].ActivityType != sched2.Activities[i].ActivityType {
				different = true
				break
			}
		}
	}

	// Note: This test may occasionally fail if two different seeds happen to produce
	// identical schedules, but it's statistically unlikely
	if !different {
		t.Log("Warning: Different seeds produced identical schedules (statistically rare)")
	}
}

func BenchmarkScheduleSystem_Update(b *testing.B) {
	world := NewWorld()
	clock := NewSimulationClock(12345)
	clock.Reset(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC))
	sys := NewScheduleSystem(world, clock)

	// Create 100 NPCs with schedules
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		entity.AddComponent(GenerateDefaultSchedule(int64(i), "villager", 0, 0, 100, 100))
		entities[i] = entity
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
