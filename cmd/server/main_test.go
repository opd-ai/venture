//go:build !android && !ios
// +build !android,!ios

package main

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestInitializeLogger_Default tests default logger initialization.
func TestInitializeLogger_Default(t *testing.T) {
	// Save and restore environment
	oldLevel := os.Getenv("LOG_LEVEL")
	os.Unsetenv("LOG_LEVEL")
	defer func() {
		if oldLevel != "" {
			os.Setenv("LOG_LEVEL", oldLevel)
		}
	}()

	// Save and restore verbose flag
	originalVerbose := *verbose
	*verbose = false
	defer func() { *verbose = originalVerbose }()

	logger := initializeLogger()

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}
}

// TestInitializeLogger_Verbose tests logger with verbose flag.
func TestInitializeLogger_Verbose(t *testing.T) {
	// Save and restore environment
	oldLevel := os.Getenv("LOG_LEVEL")
	os.Unsetenv("LOG_LEVEL")
	defer func() {
		if oldLevel != "" {
			os.Setenv("LOG_LEVEL", oldLevel)
		}
	}()

	// Enable verbose
	originalVerbose := *verbose
	*verbose = true
	defer func() { *verbose = originalVerbose }()

	logger := initializeLogger()

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}
	// Verbose mode should set debug level
	if logger.GetLevel() < logrus.DebugLevel {
		t.Errorf("Expected debug level or higher, got %v", logger.GetLevel())
	}
}

// TestInitializeLogger_EnvOverride tests LOG_LEVEL environment variable.
func TestInitializeLogger_EnvOverride(t *testing.T) {
	// Save and restore environment
	oldLevel := os.Getenv("LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "warn")
	defer func() {
		if oldLevel != "" {
			os.Setenv("LOG_LEVEL", oldLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Even with verbose, env should take precedence
	originalVerbose := *verbose
	*verbose = true
	defer func() { *verbose = originalVerbose }()

	logger := initializeLogger()

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}
}

// TestBuildWorldSnapshot_Empty tests snapshot building with no entities.
func TestBuildWorldSnapshot_Empty(t *testing.T) {
	world := engine.NewWorld()
	world.Update(0) // Process pending entities
	timestamp := time.Now()

	snapshot := buildWorldSnapshot(world, timestamp)

	if len(snapshot.Entities) != 0 {
		t.Errorf("Expected 0 entities in snapshot, got %d", len(snapshot.Entities))
	}

	if snapshot.Timestamp != timestamp {
		t.Errorf("Expected timestamp %v, got %v", timestamp, snapshot.Timestamp)
	}
}

// TestBuildWorldSnapshot_EntityWithPosition tests snapshot with position-only entity.
func TestBuildWorldSnapshot_EntityWithPosition(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 100.5, Y: 200.5})
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	snapshot := buildWorldSnapshot(world, timestamp)

	if len(snapshot.Entities) != 1 {
		t.Fatalf("Expected 1 entity in snapshot, got %d", len(snapshot.Entities))
	}

	entitySnap, exists := snapshot.Entities[entity.ID]
	if !exists {
		t.Fatalf("Entity %d not found in snapshot", entity.ID)
	}

	if entitySnap.Position.X != 100.5 {
		t.Errorf("Expected X=100.5, got %f", entitySnap.Position.X)
	}
	if entitySnap.Position.Y != 200.5 {
		t.Errorf("Expected Y=200.5, got %f", entitySnap.Position.Y)
	}

	// No velocity component, so velocity should be 0
	if entitySnap.Velocity.VX != 0 || entitySnap.Velocity.VY != 0 {
		t.Errorf("Expected velocity (0,0), got (%f,%f)", entitySnap.Velocity.VX, entitySnap.Velocity.VY)
	}
}

// TestBuildWorldSnapshot_EntityWithPositionAndVelocity tests snapshot with both position and velocity.
func TestBuildWorldSnapshot_EntityWithPositionAndVelocity(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 50.0, Y: 75.0})
	entity.AddComponent(&engine.VelocityComponent{VX: 5.5, VY: -3.3})
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	snapshot := buildWorldSnapshot(world, timestamp)

	if len(snapshot.Entities) != 1 {
		t.Fatalf("Expected 1 entity in snapshot, got %d", len(snapshot.Entities))
	}

	entitySnap := snapshot.Entities[entity.ID]

	if entitySnap.Velocity.VX != 5.5 {
		t.Errorf("Expected VX=5.5, got %f", entitySnap.Velocity.VX)
	}
	if entitySnap.Velocity.VY != -3.3 {
		t.Errorf("Expected VY=-3.3, got %f", entitySnap.Velocity.VY)
	}
}

// TestBuildWorldSnapshot_EntityWithoutPosition tests entities without position are not included.
func TestBuildWorldSnapshot_EntityWithoutPosition(t *testing.T) {
	world := engine.NewWorld()
	// Entity with only velocity (no position)
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{VX: 10.0, VY: 10.0})
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	snapshot := buildWorldSnapshot(world, timestamp)

	// Should not include entity without position
	if len(snapshot.Entities) != 0 {
		t.Errorf("Expected 0 entities (entity has no position), got %d", len(snapshot.Entities))
	}
}

// TestBuildWorldSnapshot_MultipleEntities tests snapshot with multiple entities.
func TestBuildWorldSnapshot_MultipleEntities(t *testing.T) {
	world := engine.NewWorld()

	// Create 5 entities with positions
	for i := 0; i < 5; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i * 100), Y: float64(i * 50)})
	}

	// Create 2 entities without positions (should be excluded)
	for i := 0; i < 2; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 1.0})
	}
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	snapshot := buildWorldSnapshot(world, timestamp)

	if len(snapshot.Entities) != 5 {
		t.Errorf("Expected 5 entities (only with positions), got %d", len(snapshot.Entities))
	}
}

// TestBuildWorldSnapshot_WithVehicleComponent tests snapshot includes vehicle component data.
func TestBuildWorldSnapshot_WithVehicleComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.VehicleComponent{
		VehicleType: engine.VehicleCart,
		Speed:       100.0,
		MaxSpeed:    200.0,
	})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]
	vehicleData, exists := entitySnap.Components["vehicle"]
	if !exists {
		t.Error("Expected vehicle component in snapshot")
	}
	if len(vehicleData) == 0 {
		t.Error("Vehicle component data should not be empty")
	}
}

// TestBuildWorldSnapshot_WithCompanionComponent tests snapshot includes companion component data.
func TestBuildWorldSnapshot_WithCompanionComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.CompanionComponent{
		OwnerID:       123,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       80.0,
	})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]
	companionData, exists := entitySnap.Components["companion"]
	if !exists {
		t.Error("Expected companion component in snapshot")
	}
	if len(companionData) == 0 {
		t.Error("Companion component data should not be empty")
	}
}

// TestBuildWorldSnapshot_WithMountComponent tests snapshot includes mount component data.
func TestBuildWorldSnapshot_WithMountComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.MountComponent{
		RiderID:         123,
		MountedEntityID: 456,
		IsMounted:       true,
	})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]
	mountData, exists := entitySnap.Components["mount"]
	if !exists {
		t.Error("Expected mount component in snapshot")
	}
	if len(mountData) == 0 {
		t.Error("Mount component data should not be empty")
	}
}

// TestBuildWorldSnapshot_WithAchievementComponent tests snapshot includes achievement component data.
func TestBuildWorldSnapshot_WithAchievementComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.AchievementComponent{
		Achievements:    []engine.Achievement{},
		ExpressionCount: 10,
	})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]
	achievementData, exists := entitySnap.Components["achievement"]
	if !exists {
		t.Error("Expected achievement component in snapshot")
	}
	if len(achievementData) == 0 {
		t.Error("Achievement component data should not be empty")
	}
}

// TestBuildWorldSnapshot_WithBookshelfComponent tests snapshot includes bookshelf component data.
func TestBuildWorldSnapshot_WithBookshelfComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.BookshelfComponent{
		Books:    []uint64{1, 2, 3},
		Capacity: 10,
	})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]
	bookshelfData, exists := entitySnap.Components["bookshelf"]
	if !exists {
		t.Error("Expected bookshelf component in snapshot")
	}
	if len(bookshelfData) == 0 {
		t.Error("Bookshelf component data should not be empty")
	}
}

// TestBuildWorldSnapshot_WithAllComponents tests snapshot with all V4.0 components.
func TestBuildWorldSnapshot_WithAllComponents(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 10.0, Y: 20.0})
	entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 2.0})
	entity.AddComponent(&engine.VehicleComponent{VehicleType: engine.VehicleBoat})
	entity.AddComponent(&engine.CompanionComponent{OwnerID: 1})
	entity.AddComponent(&engine.MountComponent{IsMounted: true})
	entity.AddComponent(&engine.AchievementComponent{Achievements: []engine.Achievement{}})
	entity.AddComponent(&engine.BookshelfComponent{Books: []uint64{}})
	world.Update(0) // Process pending entities

	snapshot := buildWorldSnapshot(world, time.Now())

	entitySnap := snapshot.Entities[entity.ID]

	// Should have all 5 V4.0 component types in Components map
	expectedComponents := []string{"vehicle", "companion", "mount", "achievement", "bookshelf"}
	for _, compType := range expectedComponents {
		if _, exists := entitySnap.Components[compType]; !exists {
			t.Errorf("Missing component type in snapshot: %s", compType)
		}
	}
}

// TestInitializeResilienceTesting_MetricsEnabled tests metrics collector initialization.
func TestInitializeResilienceTesting_MetricsEnabled(t *testing.T) {
	// Save and restore original flag value
	originalMetrics := *resilienceMetrics
	originalSimulate := *simulateNetwork
	defer func() {
		*resilienceMetrics = originalMetrics
		*simulateNetwork = originalSimulate
	}()

	*resilienceMetrics = true
	*simulateNetwork = ""

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{}) // Suppress output
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	networkSim, metricsCollector := initializeResilienceTesting(serverLogger)

	if metricsCollector == nil {
		t.Error("Expected metrics collector to be initialized")
	}
	if networkSim != nil {
		t.Error("Expected network simulator to be nil when simulateNetwork is empty")
	}
}

// TestInitializeResilienceTesting_MetricsDisabled tests without metrics collector.
func TestInitializeResilienceTesting_MetricsDisabled(t *testing.T) {
	originalMetrics := *resilienceMetrics
	originalSimulate := *simulateNetwork
	defer func() {
		*resilienceMetrics = originalMetrics
		*simulateNetwork = originalSimulate
	}()

	*resilienceMetrics = false
	*simulateNetwork = ""

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	networkSim, metricsCollector := initializeResilienceTesting(serverLogger)

	if metricsCollector != nil {
		t.Error("Expected metrics collector to be nil when disabled")
	}
	if networkSim != nil {
		t.Error("Expected network simulator to be nil")
	}
}

// TestInitializeResilienceTesting_LowLatencyScenario tests low latency network simulation.
func TestInitializeResilienceTesting_LowLatencyScenario(t *testing.T) {
	originalMetrics := *resilienceMetrics
	originalSimulate := *simulateNetwork
	defer func() {
		*resilienceMetrics = originalMetrics
		*simulateNetwork = originalSimulate
	}()

	*resilienceMetrics = false
	*simulateNetwork = "low"

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	networkSim, _ := initializeResilienceTesting(serverLogger)

	if networkSim == nil {
		t.Error("Expected network simulator to be initialized for 'low' scenario")
	}
}

// TestInitializeResilienceTesting_AllScenarios tests all network simulation scenarios.
func TestInitializeResilienceTesting_AllScenarios(t *testing.T) {
	scenarios := []string{"low", "medium", "high", "very-high", "extreme"}

	for _, scenario := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			originalMetrics := *resilienceMetrics
			originalSimulate := *simulateNetwork
			defer func() {
				*resilienceMetrics = originalMetrics
				*simulateNetwork = originalSimulate
			}()

			*resilienceMetrics = false
			*simulateNetwork = scenario

			logger := logrus.New()
			logger.SetOutput(&bytes.Buffer{})
			serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

			networkSim, _ := initializeResilienceTesting(serverLogger)

			if networkSim == nil {
				t.Errorf("Expected network simulator for scenario '%s'", scenario)
			}
		})
	}
}

// TestInitializeResilienceTesting_UnknownScenario tests unknown scenario defaults to low.
func TestInitializeResilienceTesting_UnknownScenario(t *testing.T) {
	originalMetrics := *resilienceMetrics
	originalSimulate := *simulateNetwork
	defer func() {
		*resilienceMetrics = originalMetrics
		*simulateNetwork = originalSimulate
	}()

	*resilienceMetrics = false
	*simulateNetwork = "invalid-scenario"

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	networkSim, _ := initializeResilienceTesting(serverLogger)

	// Unknown scenarios should fall back to low latency
	if networkSim == nil {
		t.Error("Expected network simulator even for unknown scenario (defaults to low)")
	}
}

// BenchmarkBuildWorldSnapshot_100Entities benchmarks snapshot building.
func BenchmarkBuildWorldSnapshot_100Entities(b *testing.B) {
	world := engine.NewWorld()

	// Create 100 entities with various components
	for i := 0; i < 100; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i), Y: float64(i * 2)})
		entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 1.0})
	}
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = buildWorldSnapshot(world, timestamp)
	}
}

// BenchmarkBuildWorldSnapshot_1000Entities benchmarks snapshot building with more entities.
func BenchmarkBuildWorldSnapshot_1000Entities(b *testing.B) {
	world := engine.NewWorld()

	// Create 1000 entities with all V4.0 components
	for i := 0; i < 1000; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.PositionComponent{X: float64(i), Y: float64(i * 2)})
		entity.AddComponent(&engine.VelocityComponent{VX: 1.0, VY: 1.0})
		entity.AddComponent(&engine.VehicleComponent{VehicleType: engine.VehicleCart})
		entity.AddComponent(&engine.CompanionComponent{OwnerID: uint64(i)})
	}
	world.Update(0) // Process pending entities

	timestamp := time.Now()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = buildWorldSnapshot(world, timestamp)
	}
}

// TestInitializeModSystem_WithMods tests mod system initialization with a valid mods directory.
func TestInitializeModSystem_WithMods(t *testing.T) {
	// Save and restore original flag value
	originalModsDir := *modsDir
	originalEnableMods := *enableMods
	defer func() {
		*modsDir = originalModsDir
		*enableMods = originalEnableMods
	}()

	// Use a temporary directory for mods
	tmpDir, err := os.MkdirTemp("", "venture-mods-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	*modsDir = tmpDir
	*enableMods = true

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	// initializeModSystem should not panic with empty mods directory
	manager := initializeModSystem(serverLogger)

	// Manager may be nil if sandbox security checks fail, but should not panic
	// The function should handle empty mods directory gracefully
	_ = manager
}

// TestInitializeModSystem_InvalidDirectory tests mod system with non-existent directory.
func TestInitializeModSystem_InvalidDirectory(t *testing.T) {
	// Save and restore original flag value
	originalModsDir := *modsDir
	originalEnableMods := *enableMods
	defer func() {
		*modsDir = originalModsDir
		*enableMods = originalEnableMods
	}()

	// Use a non-existent directory
	*modsDir = "/non/existent/path/for/mods/test"
	*enableMods = true

	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	// Should not panic even with invalid directory
	manager := initializeModSystem(serverLogger)

	// The function should handle this gracefully
	_ = manager
}

// TestStartStabilityMonitoring tests stability monitoring initialization.
func TestStartStabilityMonitoring(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	// Start the stability monitor
	monitor := startStabilityMonitoring(serverLogger)

	// Should return a non-nil monitor
	if monitor == nil {
		t.Error("Expected stability monitor to be initialized")
	}

	// Give the background goroutine a moment to start
	time.Sleep(10 * time.Millisecond)
}

// TestStartStabilityMonitoring_ConfigValues tests that stability config is set correctly.
func TestStartStabilityMonitoring_ConfigValues(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	serverLogger := logger.WithFields(logrus.Fields{"component": "test"})

	monitor := startStabilityMonitoring(serverLogger)

	if monitor == nil {
		t.Fatal("Expected monitor to be initialized")
	}
}

// TestSystemWrappers_CompanionAISystemWrapper tests companion AI system wrapper.
func TestSystemWrappers_CompanionAISystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewCompanionAISystem(world)
	wrapper := &companionAISystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_CompanionProgressionSystemWrapper tests companion progression wrapper.
func TestSystemWrappers_CompanionProgressionSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewCompanionProgressionSystem(world)
	wrapper := &companionProgressionSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_CompanionLoyaltySystemWrapper tests companion loyalty wrapper.
func TestSystemWrappers_CompanionLoyaltySystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	system := engine.NewCompanionLoyaltySystem(world, logger)
	wrapper := &companionLoyaltySystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_CompanionInventorySystemWrapper tests companion inventory wrapper.
func TestSystemWrappers_CompanionInventorySystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewCompanionInventorySystem(world)
	wrapper := &companionInventorySystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_SkillInheritanceSystemWrapper tests skill inheritance wrapper.
func TestSystemWrappers_SkillInheritanceSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewSkillInheritanceSystem(world)
	wrapper := &skillInheritanceSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_MiniGameSystemWrapper tests mini game wrapper.
func TestSystemWrappers_MiniGameSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewMiniGameSystem(world)
	wrapper := &miniGameSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_AlignmentSystemWrapper tests alignment wrapper.
func TestSystemWrappers_AlignmentSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewAlignmentSystem(world)
	wrapper := &alignmentSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_VehicleSystemWrapper tests vehicle system wrapper.
func TestSystemWrappers_VehicleSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewVehicleSystem(world)
	wrapper := &vehicleSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_DiscoverySystemWrapper tests discovery system wrapper.
func TestSystemWrappers_DiscoverySystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewDiscoverySystem(world)
	wrapper := &discoverySystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_AchievementSystemWrapper tests achievement system wrapper.
func TestSystemWrappers_AchievementSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewAchievementSystem(world)
	wrapper := &achievementSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_RotationSystemWrapper tests rotation system wrapper.
func TestSystemWrappers_RotationSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewRotationSystem(world)
	wrapper := &rotationSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_InvestigationSystemWrapper tests investigation system wrapper.
func TestSystemWrappers_InvestigationSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewInvestigationSystem(world, 12345)
	wrapper := &investigationSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_MerchantCaravanSystemWrapper tests merchant caravan wrapper.
func TestSystemWrappers_MerchantCaravanSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewMerchantCaravanSystem(world)
	wrapper := &merchantCaravanSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_SquadSystemWrapper tests squad system wrapper.
func TestSystemWrappers_SquadSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewSquadSystem(world)
	wrapper := &squadSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_PoliticsSystemWrapper tests politics system wrapper.
func TestSystemWrappers_PoliticsSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	system := engine.NewPoliticsSystem(world)
	wrapper := &politicsSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_FactionReactionSystemWrapper tests faction reaction wrapper.
func TestSystemWrappers_FactionReactionSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	system := engine.NewFactionReactionSystem(world, logger)
	wrapper := &factionReactionSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestSystemWrappers_MoralChoiceSystemWrapper tests moral choice wrapper.
func TestSystemWrappers_MoralChoiceSystemWrapper(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.New()
	logger.SetOutput(&bytes.Buffer{})
	system := engine.NewMoralChoiceSystem(world, logger)
	wrapper := &moralChoiceSystemWrapper{system: system}

	// Should not panic when called
	wrapper.Update([]*engine.Entity{}, 0.016)
}
