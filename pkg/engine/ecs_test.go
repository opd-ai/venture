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
	if entity.ID != 0 {
		t.Errorf("Expected first entity ID to be 0, got %d", entity.ID)
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

// Test that doesn't require display - just tests the constructor
func TestGameStructure(t *testing.T) {
	// Skip test requiring display in CI/headless environments
	if os.Getenv("DISPLAY") == "" && os.Getenv("CI") != "" {
		t.Skip("Skipping Game test - no display available in CI")
	}
	
	// This will be tested in integration tests with a virtual display
	t.Skip("Game tests require display - skipped for unit tests")
}
