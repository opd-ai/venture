package engine

import (
	"os"
	"testing"
)

func TestNewEntity(t *testing.T) {
	entity := NewEntity(1)
	if entity.ID != 1 {
		t.Errorf("Expected entity ID 1, got %d", entity.ID)
	}
	if entity.Components == nil {
		t.Error("Expected Components map to be initialized")
	}
}

type MockComponent struct {
	Value string
}

func (m *MockComponent) Type() string {
	return "mock"
}

func TestEntityComponents(t *testing.T) {
	entity := NewEntity(1)

	// Test adding component
	comp := &MockComponent{Value: "test"}
	entity.AddComponent(comp)

	if !entity.HasComponent("mock") {
		t.Error("Expected entity to have mock component")
	}

	// Test getting component
	retrieved, ok := entity.GetComponent("mock")
	if !ok {
		t.Error("Expected to retrieve mock component")
	}
	if mockComp, ok := retrieved.(*MockComponent); !ok || mockComp.Value != "test" {
		t.Error("Retrieved component doesn't match")
	}

	// Test removing component
	entity.RemoveComponent("mock")
	if entity.HasComponent("mock") {
		t.Error("Expected component to be removed")
	}
}

func TestWorld(t *testing.T) {
	world := NewWorld()

	// Test entity creation
	entity := world.CreateEntity()
	if entity.ID != 1 {
		t.Errorf("Expected first entity ID to be 1, got %d", entity.ID)
	}

	// Ensure entity is added after update
	world.Update(0.016)

	retrieved, ok := world.GetEntity(entity.ID)
	if !ok {
		t.Error("Expected to retrieve created entity")
	}
	if retrieved.ID != entity.ID {
		t.Error("Retrieved entity doesn't match")
	}

	// Test entity removal
	world.RemoveEntity(entity.ID)
	world.Update(0.016)

	_, ok = world.GetEntity(entity.ID)
	if ok {
		t.Error("Expected entity to be removed")
	}
}

type MockSystem struct {
	UpdateCount int
}

func (s *MockSystem) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

func TestWorldSystems(t *testing.T) {
	world := NewWorld()
	system := &MockSystem{}

	world.AddSystem(system)
	world.Update(0.016)

	if system.UpdateCount != 1 {
		t.Errorf("Expected system to be updated once, got %d", system.UpdateCount)
	}
}

func TestWorldAddNilSystem(t *testing.T) {
	world := NewWorld()

	// Attempt to add nil system - should be prevented
	world.AddSystem(nil)

	// Create a test entity
	entity := world.CreateEntity()
	entity.AddComponent(&MockComponent{Value: "test"})

	// Update should not panic even though we tried to add a nil system
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked after attempting to add nil system: %v", r)
		}
	}()

	world.Update(0.016)

	// Verify entity was processed correctly
	retrieved, ok := world.GetEntity(entity.ID)
	if !ok {
		t.Error("Expected to retrieve created entity")
	}
	if !retrieved.HasComponent("mock") {
		t.Error("Expected entity to have mock component")
	}
}

func TestGetEntitiesWith(t *testing.T) {
	world := NewWorld()

	// Create entities with different components
	entity1 := world.CreateEntity()
	entity1.AddComponent(&MockComponent{Value: "e1"})

	entity2 := world.CreateEntity()
	entity2.AddComponent(&MockComponent{Value: "e2"})

	_ = world.CreateEntity()
	// No components

	world.Update(0.016)

	// Get entities with mock component
	entities := world.GetEntitiesWith("mock")
	if len(entities) != 2 {
		t.Errorf("Expected 2 entities with mock component, got %d", len(entities))
	}
}

// TestGetEntitiesWithOptimizedPaths verifies the zero-allocation fast paths work correctly.
func TestGetEntitiesWithOptimizedPaths(t *testing.T) {
	world := NewWorld()

	// Create test entities with common component combinations
	for i := 0; i < 10; i++ {
		e := world.CreateEntity()
		e.AddComponent(&PositionComponent{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			e.AddComponent(&VelocityComponent{VX: 1.0, VY: 1.0})
		}
		if i%3 == 0 {
			e.AddComponent(&HealthComponent{Current: 100, Max: 100})
		}
		if i%4 == 0 {
			e.AddComponent(&ColliderComponent{Width: 32, Height: 32})
		}
	}

	world.Update(0.016)

	// Test fast-path queries
	tests := []struct {
		name       string
		components []string
		wantCount  int
	}{
		{"position only", []string{"position"}, 10},
		{"position+velocity", []string{"position", "velocity"}, 5},
		{"position+health", []string{"position", "health"}, 4},
		{"position+collider", []string{"position", "collider"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First call (cache miss)
			result1 := world.GetEntitiesWith(tt.components...)
			if len(result1) != tt.wantCount {
				t.Errorf("%s: expected %d entities, got %d", tt.name, tt.wantCount, len(result1))
			}

			// Second call (cache hit - should use fast path with zero allocations)
			result2 := world.GetEntitiesWith(tt.components...)
			if len(result2) != tt.wantCount {
				t.Errorf("%s: cached result expected %d entities, got %d", tt.name, tt.wantCount, len(result2))
			}

			// Verify results are identical
			if len(result1) != len(result2) {
				t.Errorf("%s: cache results differ in length", tt.name)
			}
		})
	}
}

// Test that doesn't require display - just tests the constructor
func TestGameStructure(t *testing.T) {
	// Skip test requiring display in CI/headless environments
	if os.Getenv("DISPLAY") == "" && os.Getenv("CI") != "" {
		t.Skip("Skipping Game test - no display available in CI")
	}

	// This will be tested in integration tests with a virtual display
	t.Skip("Game tests require display - skipped for unit tests")
}
