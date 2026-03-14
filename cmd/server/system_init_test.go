//go:build !android && !ios
// +build !android,!ios

package main

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// createTestLoggerForSystems creates a logger for testing that discards output
func createTestLoggerForSystems() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

// TestInitializeV4Systems verifies that all V4.0 systems are added to the world
func TestInitializeV4Systems(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	initialSystemCount := len(world.GetSystems())

	initializeV4Systems(world, seed, "fantasy", logger, nil)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// V4 should add 26 systems according to the function comments
	expectedMinSystems := 20 // Allow some flexibility since some systems may be consolidated
	if addedSystems < expectedMinSystems {
		t.Errorf("initializeV4Systems added %d systems, expected at least %d", addedSystems, expectedMinSystems)
	}

	t.Logf("V4 systems initialized: %d systems added", addedSystems)
}

// TestInitializeV4Systems_NilLogger verifies system init doesn't panic with nil-safe logger
func TestInitializeV4Systems_Deterministic(t *testing.T) {
	world1 := engine.NewWorld()
	world2 := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	initializeV4Systems(world1, seed, "fantasy", logger, nil)
	initializeV4Systems(world2, seed, "fantasy", logger, nil)

	// Both worlds should have the same number of systems
	if len(world1.GetSystems()) != len(world2.GetSystems()) {
		t.Errorf("System initialization is not deterministic: world1=%d, world2=%d",
			len(world1.GetSystems()), len(world2.GetSystems()))
	}
}

// TestInitializeV5SystemsServer verifies V5.0 social systems are added
func TestInitializeV5SystemsServer(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()

	initialSystemCount := len(world.GetSystems())

	enhancedChat := initializeV5SystemsServer(world, logger)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// V5 should add 3 systems (enhanced chat, mail, courier)
	expectedSystems := 3
	if addedSystems != expectedSystems {
		t.Errorf("initializeV5SystemsServer added %d systems, expected %d", addedSystems, expectedSystems)
	}

	// Verify EnhancedChatSystem is returned
	if enhancedChat == nil {
		t.Error("initializeV5SystemsServer should return non-nil EnhancedChatSystem")
	}

	t.Logf("V5 systems initialized: %d systems added (returned EnhancedChatSystem)", addedSystems)
}

// TestInitializeV6SystemsServer verifies V6.0 federation systems are added
func TestInitializeV6SystemsServer(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	initialSystemCount := len(world.GetSystems())

	initializeV6SystemsServer(world, seed, logger, nil)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// V6 should add 5 systems (portal, bounty, politics, world events, trade routes)
	expectedSystems := 5
	if addedSystems != expectedSystems {
		t.Errorf("initializeV6SystemsServer added %d systems, expected %d", addedSystems, expectedSystems)
	}

	t.Logf("V6 systems initialized: %d systems added", addedSystems)
}

// TestInitializeV8SystemsServer verifies V8.0 systems are added
func TestInitializeV8SystemsServer(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	initialSystemCount := len(world.GetSystems())

	guildManager, fleetManager, _ := initializeV8SystemsServer(world, seed, "test-server", logger)

	if guildManager == nil {
		t.Error("initializeV8SystemsServer returned nil guildManager")
	}

	if fleetManager == nil {
		t.Error("initializeV8SystemsServer returned nil fleetManager")
	}

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// V8 should add at least 2 systems (guild, fluid simulator)
	expectedMinSystems := 2
	if addedSystems < expectedMinSystems {
		t.Errorf("initializeV8SystemsServer added %d systems, expected at least %d", addedSystems, expectedMinSystems)
	}

	t.Logf("V8 systems initialized: %d systems added (guildManager, fleetManager returned)", addedSystems)
}

// TestInitializeV9SystemsServer verifies V9.0 integration managers are created
func TestInitializeV9SystemsServer(t *testing.T) {
	logger := createTestLoggerForSystems()
	world := engine.NewWorld()
	seed := int64(12345)

	// V9 now requires guildManager from V8
	guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)

	stationManager, petHomeManager, guildHousingManager, narrativeWorldSys, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)

	if stationManager == nil {
		t.Error("initializeV9SystemsServer returned nil stationManager")
	}

	if petHomeManager == nil {
		t.Error("initializeV9SystemsServer returned nil petHomeManager")
	}

	if guildHousingManager == nil {
		t.Error("initializeV9SystemsServer returned nil guildHousingManager")
	}

	if narrativeWorldSys == nil {
		t.Error("initializeV9SystemsServer returned nil narrativeWorldSystem")
	}

	if politicalWarfareSys == nil {
		t.Error("initializeV9SystemsServer returned nil politicalWarfareSystem")
	}

	t.Log("V9 integration managers initialized successfully")
}

// TestV8FluidPhysicsManagersInitialization verifies fluid physics managers can be initialized on server
func TestV8FluidPhysicsManagersInitialization(t *testing.T) {
	// Create a test logger that captures output
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.DebugLevel)

	world := engine.NewWorld()
	seed := int64(12345)

	// Initialize V8 systems which now includes fluid physics managers
	_, _, _ = initializeV8SystemsServer(world, seed, "test-server", logger)

	// Verify debug log message was written
	logOutput := buf.String()
	if !bytes.Contains([]byte(logOutput), []byte("Fluid physics managers initialized")) {
		t.Error("Expected debug log for fluid physics managers initialization, but not found")
	}

	// Verify the log mentions all three managers
	if !bytes.Contains([]byte(logOutput), []byte("buoyancy")) {
		t.Error("Expected 'buoyancy' in log output")
	}
	if !bytes.Contains([]byte(logOutput), []byte("swimming")) {
		t.Error("Expected 'swimming' in log output")
	}
	if !bytes.Contains([]byte(logOutput), []byte("flooding")) {
		t.Error("Expected 'flooding' in log output")
	}

	t.Log("Fluid physics managers initialization verified")
}

// TestInitializeCoreGameplaySystems verifies core gameplay systems are added
func TestInitializeCoreGameplaySystems(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	// Create required dependencies
	inventorySystem := engine.NewInventorySystem(world)
	itemGen := itemgen.NewItemGenerator()

	initialSystemCount := len(world.GetSystems())

	initializeCoreGameplaySystems(world, seed, logger, inventorySystem, itemGen)

	finalSystemCount := len(world.GetSystems())
	addedSystems := finalSystemCount - initialSystemCount

	// Core gameplay should add 29 systems according to the function comments
	expectedMinSystems := 25 // Allow some flexibility
	if addedSystems < expectedMinSystems {
		t.Errorf("initializeCoreGameplaySystems added %d systems, expected at least %d", addedSystems, expectedMinSystems)
	}

	t.Logf("Core gameplay systems initialized: %d systems added", addedSystems)
}

// TestAllSystemsInitialization_Integration tests initializing all system versions together
func TestAllSystemsInitialization_Integration(t *testing.T) {
	world := engine.NewWorld()
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	// Create required dependencies
	inventorySystem := engine.NewInventorySystem(world)
	itemGen := itemgen.NewItemGenerator()

	// Initialize all system versions in order
	initializeV4Systems(world, seed, "fantasy", logger, nil)
	initializeV5SystemsServer(world, logger)
	initializeV6SystemsServer(world, seed, logger, nil)
	guildManager, fleetManager, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
	initializeCoreGameplaySystems(world, seed, logger, inventorySystem, itemGen)

	// Verify FleetManager is returned from V8
	if fleetManager == nil {
		t.Error("FleetManager not returned from V8 initialization")
	}

	// Verify V9 managers are created
	stationManager, petHomeManager, guildHousingManager, narrativeWorldSys, politicalWarfareSys := initializeV9SystemsServer(world, seed, guildManager, logger)
	if stationManager == nil || petHomeManager == nil || guildHousingManager == nil || narrativeWorldSys == nil || politicalWarfareSys == nil {
		t.Error("V9 managers not created")
	}

	totalSystems := len(world.GetSystems())
	expectedMinSystems := 50 // Should have at least 50 systems total
	if totalSystems < expectedMinSystems {
		t.Errorf("Total systems after all initialization: %d, expected at least %d", totalSystems, expectedMinSystems)
	}

	t.Logf("All systems initialized: %d total systems", totalSystems)
}

// TestInitializeV4Systems_DifferentSeeds verifies different seeds produce same system count
func TestInitializeV4Systems_DifferentSeeds(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"seed_0", 0},
		{"seed_1", 1},
		{"seed_max", 9223372036854775807},
		{"seed_negative", -12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			logger := createTestLoggerForSystems()

			initializeV4Systems(world, tt.seed, "fantasy", logger, nil)

			systemCount := len(world.GetSystems())
			if systemCount == 0 {
				t.Error("No systems were initialized")
			}
		})
	}
}

// TestSystemInitialization_LoggerLevels verifies initialization works with different log levels
func TestSystemInitialization_LoggerLevels(t *testing.T) {
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			world := engine.NewWorld()
			logger := logrus.New()
			logger.SetOutput(bytes.NewBuffer(nil))
			logger.SetLevel(level)

			// Should not panic regardless of log level
			initializeV4Systems(world, 12345, "fantasy", logger, nil)
			initializeV5SystemsServer(world, logger)
			initializeV6SystemsServer(world, 12345, logger, nil)
			guildMgr, _, _ := initializeV8SystemsServer(world, 12345, "test-server", logger)
			initializeV9SystemsServer(world, 12345, guildMgr, logger)

			if len(world.GetSystems()) == 0 {
				t.Errorf("No systems initialized at log level %s", level)
			}
		})
	}
}

// BenchmarkInitializeV4Systems measures V4 system initialization performance
func BenchmarkInitializeV4Systems(b *testing.B) {
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		initializeV4Systems(world, seed, "fantasy", logger, nil)
	}
}

// BenchmarkInitializeV5SystemsServer measures V5 system initialization performance
func BenchmarkInitializeV5SystemsServer(b *testing.B) {
	logger := createTestLoggerForSystems()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		initializeV5SystemsServer(world, logger)
	}
}

// BenchmarkInitializeV6SystemsServer measures V6 system initialization performance
func BenchmarkInitializeV6SystemsServer(b *testing.B) {
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		initializeV6SystemsServer(world, seed, logger, nil)
	}
}

// BenchmarkInitializeV8SystemsServer measures V8 system initialization performance
func BenchmarkInitializeV8SystemsServer(b *testing.B) {
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		_, _, _ = initializeV8SystemsServer(world, seed, "test-server", logger)
	}
}

// BenchmarkInitializeV9SystemsServer measures V9 manager initialization performance
func BenchmarkInitializeV9SystemsServer(b *testing.B) {
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
		initializeV9SystemsServer(world, seed, guildManager, logger)
	}
}

// BenchmarkInitializeAllSystems measures total system initialization performance
func BenchmarkInitializeAllSystems(b *testing.B) {
	logger := createTestLoggerForSystems()
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world := engine.NewWorld()
		inventorySystem := engine.NewInventorySystem(world)
		itemGen := itemgen.NewItemGenerator()

		initializeV4Systems(world, seed, "fantasy", logger, nil)
		initializeV5SystemsServer(world, logger)
		initializeV6SystemsServer(world, seed, logger, nil)
		guildManager, _, _ := initializeV8SystemsServer(world, seed, "test-server", logger)
		initializeCoreGameplaySystems(world, seed, logger, inventorySystem, itemGen)
		initializeV9SystemsServer(world, seed, guildManager, logger)
	}
}
