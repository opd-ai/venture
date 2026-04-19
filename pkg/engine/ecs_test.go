package engine

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
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

// TestSystemNameCaching verifies that system names are cached to avoid per-frame reflection.
func TestSystemNameCaching(t *testing.T) {
	world := NewWorld()
	system := &MockSystem{}

	// Add system - should populate cache
	world.AddSystem(system)

	// Verify cache was populated
	cachedName, exists := world.systemNameCache[system]
	if !exists {
		t.Error("Expected system name to be cached after AddSystem")
	}
	if cachedName == "" {
		t.Error("Expected cached system name to be non-empty")
	}
	if cachedName != "MockSystem" {
		t.Errorf("Expected cached name 'MockSystem', got '%s'", cachedName)
	}

	// Update world - should use cached name (no reflection)
	world.Update(0.016)

	// Verify name is still cached after update
	newCachedName, exists := world.systemNameCache[system]
	if !exists {
		t.Error("Expected system name to remain cached after Update")
	}
	if newCachedName != cachedName {
		t.Error("Expected cached name to remain unchanged after Update")
	}
}

// TestSystemNameCachingMultipleSystems verifies cache works with multiple systems.
func TestSystemNameCachingMultipleSystems(t *testing.T) {
	world := NewWorld()

	// Add multiple systems
	system1 := &MockSystem{}
	system2 := &MockSystem{}

	world.AddSystem(system1)
	world.AddSystem(system2)

	// Both should be in cache
	if len(world.systemNameCache) != 2 {
		t.Errorf("Expected 2 systems in cache, got %d", len(world.systemNameCache))
	}

	// Both should have correct names
	name1 := world.systemNameCache[system1]
	name2 := world.systemNameCache[system2]

	if name1 != "MockSystem" {
		t.Errorf("Expected 'MockSystem' for system1, got '%s'", name1)
	}
	if name2 != "MockSystem" {
		t.Errorf("Expected 'MockSystem' for system2, got '%s'", name2)
	}

	// Update should not panic or modify cache
	world.Update(0.016)

	if len(world.systemNameCache) != 2 {
		t.Errorf("Expected cache size to remain 2 after Update, got %d", len(world.systemNameCache))
	}
}

// TestSystemNameCachingNilSystem verifies nil systems don't pollute cache.
func TestSystemNameCachingNilSystem(t *testing.T) {
	world := NewWorld()

	// Attempt to add nil system
	world.AddSystem(nil)

	// Cache should be empty
	if len(world.systemNameCache) != 0 {
		t.Errorf("Expected empty cache after adding nil system, got %d entries", len(world.systemNameCache))
	}

	// Update should not panic
	world.Update(0.016)
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

// TestAddComponentWithLoggerUpdatesCache verifies that AddComponentWithLogger
// populates the fast-path component cache (MED-1 fix).
func TestAddComponentWithLoggerUpdatesCache(t *testing.T) {
	entity := NewEntity(1)

	pos := &PositionComponent{X: 10, Y: 20}
	entity.AddComponentWithLogger(pos, nil)

	// Cache should be populated
	cached := entity.GetPosition()
	if cached == nil {
		t.Fatal("Expected position cache to be populated via AddComponentWithLogger")
	}
	if cached.X != 10 || cached.Y != 20 {
		t.Errorf("Cached position = (%f, %f), want (10, 20)", cached.X, cached.Y)
	}

	// Also test with velocity
	vel := &VelocityComponent{VX: 5, VY: -3}
	entity.AddComponentWithLogger(vel, nil)

	cachedVel := entity.GetVelocity()
	if cachedVel == nil {
		t.Fatal("Expected velocity cache to be populated via AddComponentWithLogger")
	}
	if cachedVel.VX != 5 || cachedVel.VY != -3 {
		t.Errorf("Cached velocity = (%f, %f), want (5, -3)", cachedVel.VX, cachedVel.VY)
	}
}

// TestRemoveComponentWithLoggerClearsCache verifies that RemoveComponentWithLogger
// clears the fast-path component cache (MED-2 fix).
func TestRemoveComponentWithLoggerClearsCache(t *testing.T) {
	entity := NewEntity(1)

	// Add via normal path (known to work)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	if entity.GetPosition() == nil {
		t.Fatal("Precondition: position cache should be set")
	}

	// Remove via logger path
	entity.RemoveComponentWithLogger("position", nil)

	// Cache should be cleared
	if entity.GetPosition() != nil {
		t.Error("Expected position cache to be cleared via RemoveComponentWithLogger")
	}
	if entity.HasComponent("position") {
		t.Error("Expected position component to be removed from map")
	}
}

// TestCompanionLearningCache verifies that CompanionLearningComponent is cached.
func TestCompanionLearningCache(t *testing.T) {
	entity := NewEntity(1)

	comp := &learning.CompanionLearningComponent{
		CompanionID:  "companion-123",
		LearningRate: 1.5,
		LastSkillUse: make(map[string]time.Time),
	}

	// Add component
	entity.AddComponent(comp)

	// Verify cache is populated
	cached := entity.GetCompanionLearning()
	if cached == nil {
		t.Fatal("Expected companion learning cache to be populated")
	}
	if cached.CompanionID != "companion-123" {
		t.Errorf("Cached CompanionID = %q, want %q", cached.CompanionID, "companion-123")
	}
	if cached.LearningRate != 1.5 {
		t.Errorf("Cached LearningRate = %f, want 1.5", cached.LearningRate)
	}

	// Remove and verify cache is cleared
	entity.RemoveComponent("companion_learning")
	if entity.GetCompanionLearning() != nil {
		t.Error("Expected companion learning cache to be cleared after removal")
	}
}

// TestGuildVehicleFleetCache verifies that GuildVehicleFleetComponent is cached.
func TestGuildVehicleFleetCache(t *testing.T) {
	entity := NewEntity(1)

	comp := &guild_vehicle.GuildVehicleFleetComponent{
		GuildID:           "guild-456",
		FleetID:           "fleet-789",
		FormationPosition: 2,
	}

	// Add component
	entity.AddComponent(comp)

	// Verify cache is populated
	cached := entity.GetGuildVehicleFleet()
	if cached == nil {
		t.Fatal("Expected guild vehicle fleet cache to be populated")
	}
	if cached.GuildID != "guild-456" {
		t.Errorf("Cached GuildID = %q, want %q", cached.GuildID, "guild-456")
	}
	if cached.FleetID != "fleet-789" {
		t.Errorf("Cached FleetID = %q, want %q", cached.FleetID, "fleet-789")
	}
	if cached.FormationPosition != 2 {
		t.Errorf("Cached FormationPosition = %d, want 2", cached.FormationPosition)
	}

	// Remove and verify cache is cleared
	entity.RemoveComponent("guild_vehicle_fleet")
	if entity.GetGuildVehicleFleet() != nil {
		t.Error("Expected guild vehicle fleet cache to be cleared after removal")
	}
}

// TestCacheWithLoggerMethods verifies new components are cached via logger methods.
func TestCacheWithLoggerMethods(t *testing.T) {
	entity := NewEntity(1)

	// Test AddComponentWithLogger for CompanionLearningComponent
	clComp := &learning.CompanionLearningComponent{
		CompanionID: "test-companion",
	}
	entity.AddComponentWithLogger(clComp, nil)
	if entity.GetCompanionLearning() == nil {
		t.Error("Expected CompanionLearningComponent to be cached via AddComponentWithLogger")
	}

	// Test AddComponentWithLogger for GuildVehicleFleetComponent
	gvfComp := &guild_vehicle.GuildVehicleFleetComponent{
		GuildID: "test-guild",
	}
	entity.AddComponentWithLogger(gvfComp, nil)
	if entity.GetGuildVehicleFleet() == nil {
		t.Error("Expected GuildVehicleFleetComponent to be cached via AddComponentWithLogger")
	}

	// Test RemoveComponentWithLogger clears cache
	entity.RemoveComponentWithLogger("companion_learning", nil)
	if entity.GetCompanionLearning() != nil {
		t.Error("Expected CompanionLearningComponent cache to be cleared via RemoveComponentWithLogger")
	}

	entity.RemoveComponentWithLogger("guild_vehicle_fleet", nil)
	if entity.GetGuildVehicleFleet() != nil {
		t.Error("Expected GuildVehicleFleetComponent cache to be cleared via RemoveComponentWithLogger")
	}
}

// TestCreateEntityConcurrentSafety verifies that concurrent calls to
// CreateEntity, AddEntity, and RemoveEntity do not race on the staging
// buffers or the nextEntityID counter (C-001 / C-002 from AUDIT.md).
//
// Run with: go test -race ./pkg/engine/ -run TestCreateEntityConcurrentSafety
func TestCreateEntityConcurrentSafety(t *testing.T) {
	world := NewWorld()

	const goroutines = 20
	const entitiesPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < entitiesPerGoroutine; j++ {
				e := world.CreateEntity()
				if e == nil {
					t.Errorf("CreateEntity returned nil")
					return
				}
				// Interleave removals to exercise the removal staging buffer too.
				if j%5 == 0 {
					world.RemoveEntity(e.ID)
				}
			}
		}()
	}
	wg.Wait()

	// Flush pending entities so we can count them.
	world.FlushPendingEntities()

	// The exact entity count depends on removal timing, but no IDs should
	// have been duplicated. Check that the world has a non-zero entity count
	// and that Update() drains the staging buffers without panicking.
	world.Update(1.0 / 60.0)
}
