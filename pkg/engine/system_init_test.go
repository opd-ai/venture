// Package engine provides tests for shared system initialization.
// This file verifies that InitializeGameSystems correctly registers
// all 44 Version 2.0 systems across all platforms.
package engine

import (
	"reflect"
	"testing"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/sirupsen/logrus"
)

// TestInitializeGameSystems verifies that all 44 systems are registered.
func TestInitializeGameSystems(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	config := DefaultSystemInitConfig(12345, "fantasy", logger)
	config.EnableVerboseLogging = true

	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}

	// Verify result contains expected system references
	if result.InputSystem == nil {
		t.Error("InputSystem not returned in result")
	}
	if result.CombatSystem == nil {
		t.Error("CombatSystem not returned in result")
	}
	if result.CollisionSystem == nil {
		t.Error("CollisionSystem not returned in result")
	}
	if result.ProjectileSystem == nil {
		t.Error("ProjectileSystem not returned in result")
	}
	if result.AudioManager == nil {
		t.Error("AudioManager not returned in result")
	}
	if result.ObjectiveTracker == nil {
		t.Error("ObjectiveTracker not returned in result")
	}
	if result.CommerceSystem == nil {
		t.Error("CommerceSystem not returned in result")
	}
	if result.DialogSystem == nil {
		t.Error("DialogSystem not returned in result")
	}
	if result.CraftingSystem == nil {
		t.Error("CraftingSystem not returned in result")
	}
	if result.InteractionSystem == nil {
		t.Error("InteractionSystem not returned in result")
	}
	if result.AnimationSystem == nil {
		t.Error("AnimationSystem not returned in result")
	}
	if result.ParticleSystem == nil {
		t.Error("ParticleSystem not returned in result")
	}
	if result.TutorialSystem == nil {
		t.Error("TutorialSystem not returned in result")
	}
	if result.HelpSystem == nil {
		t.Error("HelpSystem not returned in result")
	}

	// Verify system wrappers
	if result.AnimationSystemWrapper == nil {
		t.Error("AnimationSystemWrapper not returned in result")
	}
	if result.RotationSystemWrapper == nil {
		t.Error("RotationSystemWrapper not returned in result")
	}
	if result.SquadSystemWrapper == nil {
		t.Error("SquadSystemWrapper not returned in result")
	}

	// Verify systems are registered (not including SpatialPartitionSystem which requires terrain)
	systems := game.World.GetSystems()
	if len(systems) < 200 {
		t.Errorf("Expected at least 200 systems registered, got %d", len(systems))
	}

	// Verify game references are set
	if game.TutorialSystem == nil {
		t.Error("game.TutorialSystem not set")
	}
	if game.HelpSystem == nil {
		t.Error("game.HelpSystem not set")
	}
}

// TestInitializeGameSystems_NilGame verifies error handling for nil game.
func TestInitializeGameSystems_NilGame(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	config := DefaultSystemInitConfig(12345, "fantasy", logger)

	_, err := InitializeGameSystems(nil, config)
	if err == nil {
		t.Error("Expected error for nil game, got nil")
	}
}

// TestInitializeGameSystems_NilConfig verifies error handling for nil config.
func TestInitializeGameSystems_NilConfig(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	_, err := InitializeGameSystems(game, nil)
	if err == nil {
		t.Error("Expected error for nil config, got nil")
	}
}

// TestInitializeGameSystems_NilLogger verifies error handling for nil logger.
func TestInitializeGameSystems_NilLogger(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	config := DefaultSystemInitConfig(12345, "fantasy", nil)

	_, err := InitializeGameSystems(game, config)
	if err == nil {
		t.Error("Expected error for nil logger, got nil")
	}
}

// TestDefaultSystemInitConfig verifies default configuration values.
func TestDefaultSystemInitConfig(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	seed := int64(54321)
	genre := "scifi"

	config := DefaultSystemInitConfig(seed, genre, logger)

	if config.Seed != seed {
		t.Errorf("Seed = %d, want %d", config.Seed, seed)
	}
	if config.GenreID != genre {
		t.Errorf("GenreID = %s, want %s", config.GenreID, genre)
	}
	if config.Logger != logger {
		t.Error("Logger not set correctly")
	}

	// Verify defaults
	if config.MaxSpeed != 200.0 {
		t.Errorf("MaxSpeed = %f, want 200.0", config.MaxSpeed)
	}
	if config.CollisionCellSize != 64.0 {
		t.Errorf("CollisionCellSize = %f, want 64.0", config.CollisionCellSize)
	}
	if config.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", config.SampleRate)
	}
	if config.TileSize != 32 {
		t.Errorf("TileSize = %d, want 32", config.TileSize)
	}
	if config.EnableVerboseLogging {
		t.Error("EnableVerboseLogging should default to false")
	}
}

// TestSystemInitConfig_CustomValues verifies custom configuration values are used.
func TestSystemInitConfig_CustomValues(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	config := &SystemInitConfig{
		Seed:                 99999,
		GenreID:              "horror",
		Logger:               logger,
		MaxSpeed:             150.0,
		CollisionCellSize:    48.0,
		SampleRate:           48000,
		TileSize:             64,
		EnableVerboseLogging: true,
	}

	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}

	// Verify systems use custom values
	// Note: We can't directly access internal system state in many cases,
	// but we verify no errors occurred with custom config

	if result == nil {
		t.Error("Result should not be nil")
	}
}

// TestInitializeGameSystems_MultipleGenres verifies initialization with different genres.
func TestInitializeGameSystems_MultipleGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	logger := logging.TestUtilityLogger("system_init_test")

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			game := NewEbitenGameWithLogger(800, 600, logger)
			config := DefaultSystemInitConfig(12345, genre, logger)

			_, err := InitializeGameSystems(game, config)
			if err != nil {
				t.Errorf("InitializeGameSystems failed for genre %s: %v", genre, err)
			}

			systems := game.World.GetSystems()
			if len(systems) < 200 {
				t.Errorf("Genre %s: expected at least 200 systems, got %d", genre, len(systems))
			}
		})
	}
}

// TestInitializeGameSystems_SystemConnections verifies inter-system connections.
func TestInitializeGameSystems_SystemConnections(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	config := DefaultSystemInitConfig(12345, "fantasy", logger)
	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}

	// Verify InputSystem connections
	if result.InputSystem.cameraSystem == nil {
		t.Error("InputSystem.cameraSystem not connected")
	}
	if result.InputSystem.helpSystem == nil {
		t.Error("InputSystem.helpSystem not connected")
	}
	if result.InputSystem.tutorialSystem == nil {
		t.Error("InputSystem.tutorialSystem not connected")
	}

	// Verify CombatSystem connections
	// Note: These are internal fields that may not be directly accessible
	// The test verifies the initialization completed without errors

	// Verify game references
	if game.TutorialSystem != result.TutorialSystem {
		t.Error("game.TutorialSystem not set to result.TutorialSystem")
	}
	if game.HelpSystem != result.HelpSystem {
		t.Error("game.HelpSystem not set to result.HelpSystem")
	}
}

func TestInitializeGameSystems_WeaponMaterialImpactCallback(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)
	config := DefaultSystemInitConfig(12345, "fantasy", logger)

	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}
	if result.CombatSystem == nil || result.ParticleSystem == nil || result.WeaponMaterialImpactParticleSystem == nil {
		t.Fatal("expected combat, particle, and weapon material impact systems to be initialized")
	}

	expectedPtr := reflect.ValueOf(result.WeaponMaterialImpactParticleSystem.OnMeleeImpact).Pointer()
	found := false
	for _, callback := range result.CombatSystem.additionalDamageCallbacks {
		if reflect.ValueOf(callback).Pointer() == expectedPtr {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected WeaponMaterialImpactParticleSystem.OnMeleeImpact to be registered as a combat damage callback")
	}
}

// TestInitializeGameSystems_DeterministicSeeds verifies deterministic initialization.
func TestInitializeGameSystems_DeterministicSeeds(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	seed := int64(42)

	// Initialize twice with same seed
	game1 := NewEbitenGameWithLogger(800, 600, logger)
	config1 := DefaultSystemInitConfig(seed, "fantasy", logger)
	result1, err1 := InitializeGameSystems(game1, config1)
	if err1 != nil {
		t.Fatalf("First initialization failed: %v", err1)
	}

	game2 := NewEbitenGameWithLogger(800, 600, logger)
	config2 := DefaultSystemInitConfig(seed, "fantasy", logger)
	result2, err2 := InitializeGameSystems(game2, config2)
	if err2 != nil {
		t.Fatalf("Second initialization failed: %v", err2)
	}

	// Both should have same number of systems
	systems1 := game1.World.GetSystems()
	systems2 := game2.World.GetSystems()
	if len(systems1) != len(systems2) {
		t.Errorf("System counts differ: %d vs %d", len(systems1), len(systems2))
	}

	// Both results should have all system references
	if result1 == nil || result2 == nil {
		t.Error("Results should not be nil")
	}
}

// BenchmarkInitializeGameSystems measures initialization performance.
func BenchmarkInitializeGameSystems(b *testing.B) {
	logger := logging.TestUtilityLogger("system_init_benchmark")

	for i := 0; i < b.N; i++ {
		game := NewEbitenGameWithLogger(800, 600, logger)
		config := DefaultSystemInitConfig(12345, "fantasy", logger)
		_, err := InitializeGameSystems(game, config)
		if err != nil {
			b.Fatalf("InitializeGameSystems failed: %v", err)
		}
	}
}

// TestInitializeSpatialPartitionSystem verifies spatial partition system initialization.
func TestInitializeSpatialPartitionSystem(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)

	// First initialize main systems
	config := DefaultSystemInitConfig(12345, "fantasy", logger)
	_, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}

	// Now initialize spatial partition system with mock terrain dimensions
	worldWidth := 2560.0  // 80 tiles * 32 pixels
	worldHeight := 1600.0 // 50 tiles * 32 pixels

	spatialSystem := InitializeSpatialPartitionSystem(game, worldWidth, worldHeight, true, true, logger)

	if spatialSystem == nil {
		t.Error("SpatialPartitionSystem should not be nil")
	}

	// Verify spatial partition system was added to world
	systems := game.World.GetSystems()
	if len(systems) < 200 {
		t.Errorf("Expected at least 200 systems after spatial partition init, got %d", len(systems))
	}
}

// TestSystemInitDebugFlagCaching verifies debug flag caching optimization (S4).
// This test ensures the systemInitDebugEnabled flag is set correctly during initialization.
func TestSystemInitDebugFlagCaching(t *testing.T) {
	tests := []struct {
		name          string
		logLevel      logrus.Level
		expectedDebug bool
	}{
		{"debug level", logrus.DebugLevel, true},
		{"info level", logrus.InfoLevel, false},
		{"warn level", logrus.WarnLevel, false},
		{"error level", logrus.ErrorLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.TestUtilityLogger("system_init_test")
			logger.SetLevel(tt.logLevel)
			game := NewEbitenGameWithLogger(800, 600, logger)

			config := DefaultSystemInitConfig(12345, "fantasy", logger)
			_, err := InitializeGameSystems(game, config)
			if err != nil {
				t.Fatalf("InitializeGameSystems failed: %v", err)
			}

			// Verify the cached flag is set correctly
			if systemInitDebugEnabled != tt.expectedDebug {
				t.Errorf("systemInitDebugEnabled = %v, want %v", systemInitDebugEnabled, tt.expectedDebug)
			}
		})
	}
}

// BenchmarkSystemInitDebugFlagCaching benchmarks initialization with/without debug logging.
// This measures the optimization impact of cached debug flag (S4).
func BenchmarkSystemInitDebugFlagCaching(b *testing.B) {
	benchmarks := []struct {
		name     string
		logLevel logrus.Level
	}{
		{"with_debug_logging", logrus.DebugLevel},
		{"without_debug_logging", logrus.InfoLevel},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			logger := logging.TestUtilityLogger("system_init_benchmark")
			logger.SetLevel(bm.logLevel)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				game := NewEbitenGameWithLogger(800, 600, logger)
				config := DefaultSystemInitConfig(12345, "fantasy", logger)
				_, err := InitializeGameSystems(game, config)
				if err != nil {
					b.Fatalf("InitializeGameSystems failed: %v", err)
				}
			}
		})
	}
}

// TestInitializeGameSystems_PreviouslyDanglingSystemsG2 asserts that the six
// systems identified in G2 of AUDIT.md (event cluster + mod systems) are now
// registered with the World after InitializeGameSystems returns.
// It also verifies the result references are non-nil so downstream callers
// can perform additional wiring (e.g. cmd/client repository injection for G3).
func TestInitializeGameSystems_PreviouslyDanglingSystemsG2(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)
	config := DefaultSystemInitConfig(12345, "fantasy", logger)

	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems failed: %v", err)
	}

	// G2: verify result pointers for the event cluster.
	if result.EventCalendarSystem == nil {
		t.Error("EventCalendarSystem not in result (G2)")
	}
	if result.EventQuestSystem == nil {
		t.Error("EventQuestSystem not in result (G2)")
	}
	if result.EventDecorationSystem == nil {
		t.Error("EventDecorationSystem not in result (G2)")
	}
	if result.EventRewardSystem == nil {
		t.Error("EventRewardSystem not in result (G2)")
	}
	// G2: verify result pointers for mod systems.
	if result.ModCompatibilitySystem == nil {
		t.Error("ModCompatibilitySystem not in result (G2)")
	}
	if result.ModBrowserSystem == nil {
		t.Error("ModBrowserSystem not in result (G2)")
	}

	// Verify the systems appear in the world's registered system list.
	registered := game.World.GetSystems()
	checkType := func(target interface{}, name string) {
		t.Helper()
		targetType := reflect.TypeOf(target)
		for _, sys := range registered {
			if reflect.TypeOf(sys) == targetType {
				return
			}
		}
		t.Errorf("%s not found in world.GetSystems() (G2)", name)
	}
	checkType(result.EventCalendarSystem, "EventCalendarSystem")
	checkType(result.EventQuestSystem, "EventQuestSystem")
	checkType(result.EventDecorationSystem, "EventDecorationSystem")
	checkType(result.EventRewardSystem, "EventRewardSystem")
	checkType(result.ModCompatibilitySystem, "ModCompatibilitySystem")
	checkType(result.ModBrowserSystem, "ModBrowserSystem")
}

// TestInitializeGameSystems_SingleAchievementInstances asserts that exactly one
// ExtendedAchievementSystem is registered in the World (G10).
// Duplicate registrations would cause double-fired achievement rewards.
func TestInitializeGameSystems_SingleAchievementInstances(t *testing.T) {
	logger := logging.TestUtilityLogger("system_init_test")
	game := NewEbitenGameWithLogger(800, 600, logger)
	config := DefaultSystemInitConfig(42, "fantasy", logger)

	result, err := InitializeGameSystems(game, config)
	if err != nil {
		t.Fatalf("InitializeGameSystems: %v", err)
	}

	extendedCount := 0
	for _, sys := range game.World.GetSystems() {
		if _, ok := sys.(*ExtendedAchievementSystem); ok {
			extendedCount++
		}
	}
	if extendedCount != 1 {
		t.Errorf("expected exactly 1 ExtendedAchievementSystem in World, got %d", extendedCount)
	}
	if result.ExtendedAchievementSystem == nil {
		t.Error("GameSystemsResult.ExtendedAchievementSystem is nil")
	}
}
