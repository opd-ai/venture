package engine

import (
	"testing"
)

func TestLifetimeSystem_EntityDespawn(t *testing.T) {
	tests := []struct {
		name          string
		duration      float64
		updateTime    float64
		shouldDespawn bool
	}{
		{
			name:          "entity despawns after duration",
			duration:      2.0,
			updateTime:    2.1,
			shouldDespawn: true,
		},
		{
			name:          "entity remains before duration",
			duration:      2.0,
			updateTime:    1.0,
			shouldDespawn: false,
		},
		{
			name:          "entity despawns at exact duration",
			duration:      1.5,
			updateTime:    1.5,
			shouldDespawn: true,
		},
		{
			name:          "short lifetime entity",
			duration:      0.5,
			updateTime:    0.6,
			shouldDespawn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewLifetimeSystem(world)

			// Create entity with lifetime
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			entity.AddComponent(&LifetimeComponent{
				Duration: tt.duration,
				Elapsed:  0,
			})

			// Process pending entities
			world.Update(0.0)

			// Get all entities
			entities := world.GetEntities()

			// Verify entity exists
			if len(entities) != 1 {
				t.Fatalf("Expected 1 entity, got %d", len(entities))
			}

			// Update system
			system.Update(entities, tt.updateTime)

			// Process entity removals
			world.Update(0.0)

			// Check if entity still exists
			entitiesAfter := world.GetEntities()

			if tt.shouldDespawn {
				if len(entitiesAfter) != 0 {
					t.Errorf("Expected entity to be despawned, but %d entities remain", len(entitiesAfter))
				}
			} else {
				if len(entitiesAfter) != 1 {
					t.Errorf("Expected entity to remain, but got %d entities", len(entitiesAfter))
				}
			}
		})
	}
}

func TestLifetimeSystem_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewLifetimeSystem(world)

	// Create multiple entities with different lifetimes
	entity1 := world.CreateEntity()
	entity1.AddComponent(&LifetimeComponent{Duration: 1.0, Elapsed: 0})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&LifetimeComponent{Duration: 2.0, Elapsed: 0})

	entity3 := world.CreateEntity()
	entity3.AddComponent(&LifetimeComponent{Duration: 0.5, Elapsed: 0})

	// Create entity without lifetime (should not be affected)
	entity4 := world.CreateEntity()
	entity4.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Process pending entities
	world.Update(0.0)

	entities := world.GetEntities()
	if len(entities) != 4 {
		t.Fatalf("Expected 4 entities, got %d", len(entities))
	}

	// Update for 0.6 seconds (should despawn entity3)
	system.Update(entities, 0.6)
	world.Update(0.0) // Process removals

	entitiesAfter := world.GetEntities()
	if len(entitiesAfter) != 3 {
		t.Errorf("Expected 3 entities after 0.6s, got %d", len(entitiesAfter))
	}

	// Update for another 0.6 seconds (total 1.2s, should despawn entity1)
	entitiesAfter = world.GetEntities()
	system.Update(entitiesAfter, 0.6)
	world.Update(0.0) // Process removals

	entitiesAfter = world.GetEntities()
	if len(entitiesAfter) != 2 {
		t.Errorf("Expected 2 entities after 1.2s, got %d", len(entitiesAfter))
	}

	// Update for another 1.0 seconds (total 2.2s, should despawn entity2)
	entitiesAfter = world.GetEntities()
	system.Update(entitiesAfter, 1.0)
	world.Update(0.0) // Process removals

	entitiesAfter = world.GetEntities()
	if len(entitiesAfter) != 1 {
		t.Errorf("Expected 1 entity after 2.2s (entity4 without lifetime), got %d", len(entitiesAfter))
	}

	// Verify remaining entity is entity4 (no lifetime)
	_, hasLifetime := entitiesAfter[0].GetComponent("lifetime")
	if hasLifetime {
		t.Error("Expected remaining entity to have no lifetime component")
	}
}

func TestLifetimeSystem_IncrementalUpdates(t *testing.T) {
	world := NewWorld()
	system := NewLifetimeSystem(world)

	entity := world.CreateEntity()
	entity.AddComponent(&LifetimeComponent{Duration: 2.0, Elapsed: 0})

	// Process pending entities
	world.Update(0.0)

	// Multiple small updates
	entities := world.GetEntities()
	for i := 0; i < 5; i++ {
		entities = world.GetEntities()
		system.Update(entities, 0.3)
		world.Update(0.0) // Process any removals
	}

	// Total update time: 1.5 seconds (entity should still exist)
	entitiesAfter := world.GetEntities()
	if len(entitiesAfter) != 1 {
		t.Errorf("Expected entity to remain after 1.5s, got %d entities", len(entitiesAfter))
	}

	// One more update to exceed duration
	entities = world.GetEntities()
	system.Update(entities, 0.6)
	world.Update(0.0) // Process removal

	// Total: 2.1 seconds (entity should be despawned)
	entitiesAfter = world.GetEntities()
	if len(entitiesAfter) != 0 {
		t.Errorf("Expected entity to be despawned after 2.1s, got %d entities", len(entitiesAfter))
	}
}

func TestLifetimeComponent_Type(t *testing.T) {
	comp := &LifetimeComponent{}
	if comp.Type() != "lifetime" {
		t.Errorf("Expected component type 'lifetime', got '%s'", comp.Type())
	}
}
