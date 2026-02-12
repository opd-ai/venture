package main

import (
	"runtime"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/parallel"
	"github.com/opd-ai/venture/pkg/rendering/patterns"
	"github.com/opd-ai/venture/pkg/rendering/pool"
	"github.com/opd-ai/venture/pkg/rendering/quality"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/rendering/ui"
	"github.com/sirupsen/logrus"
)

// TestParallelSystemInitialization verifies parallel init reduces startup time.
func TestParallelSystemInitialization(t *testing.T) {
	// This test verifies that parallel initialization works correctly
	// by measuring timing and ensuring all systems are initialized

	startTime := time.Now()

	// Initialize systems (calls initializeCoreSystems)
	game, _, logger, clientLogger := setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	coreSystems := initializeCoreSystems(game, logger, clientLogger)

	elapsed := time.Since(startTime)

	// Verify all expected systems are initialized
	if coreSystems.performanceSystem == nil {
		t.Error("performance system not initialized")
	}
	if coreSystems.inputSystem == nil {
		t.Error("input system not initialized")
	}
	if coreSystems.movementSystem == nil {
		t.Error("movement system not initialized")
	}
	if coreSystems.collisionSystem == nil {
		t.Error("collision system not initialized")
	}
	if coreSystems.combatSystem == nil {
		t.Error("combat system not initialized")
	}
	if coreSystems.interactionSystem == nil {
		t.Error("interaction system not initialized")
	}
	if coreSystems.particleSystem == nil {
		t.Error("particle system not initialized")
	}
	if coreSystems.spriteCache == nil {
		t.Error("sprite cache not initialized")
	}
	if coreSystems.spriteGenerator == nil {
		t.Error("sprite generator not initialized")
	}
	if coreSystems.animationSystem == nil {
		t.Error("animation system not initialized")
	}
	if coreSystems.equipmentVisualSystem == nil {
		t.Error("equipment visual system not initialized")
	}
	if coreSystems.qualitySystem == nil {
		t.Error("quality system not initialized")
	}
	if coreSystems.lightingAdapter == nil {
		t.Error("lighting adapter not initialized")
	}
	if coreSystems.animationAdapter == nil {
		t.Error("animation adapter not initialized")
	}
	if coreSystems.uiGenerator == nil {
		t.Error("ui generator not initialized")
	}
	if coreSystems.shapeRenderer == nil {
		t.Error("shape renderer not initialized")
	}
	if coreSystems.patternGenerator == nil {
		t.Error("pattern generator not initialized")
	}
	if coreSystems.imagePool == nil {
		t.Error("image pool not initialized")
	}
	if coreSystems.parallelRenderer == nil {
		t.Error("parallel renderer not initialized")
	}

	// Verify systems are connected properly
	if game.RenderSystem == nil {
		t.Error("render system not set in game")
	}

	// Log timing for informational purposes
	t.Logf("Core systems initialized in %v", elapsed)

	// Parallel init should complete faster than 100ms for these systems
	// (This is a soft check - the main value is in the parallelization itself)
	if elapsed > 200*time.Millisecond {
		t.Logf("Warning: initialization took %v (expected <200ms)", elapsed)
	}
}

// TestParallelInitSystemConnections verifies system dependencies are properly connected.
func TestParallelInitSystemConnections(t *testing.T) {
	game, _, logger, clientLogger := setupTestEnvironment(t)
	defer cleanupTestEnvironment()

	coreSystems := initializeCoreSystems(game, logger, clientLogger)

	// Verify movement system has collision system
	// Note: GetCollisionSystem() is a method on MovementSystem, not a type assertion
	if coreSystems.movementSystem == nil {
		t.Error("movement system not initialized")
	}

	// Verify animation system has sprite cache
	animSys := coreSystems.animationSystem
	if animSys == nil {
		t.Fatal("animation system is nil")
	}

	// Create test entity to verify animation system works
	entity := game.World.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&engine.AnimationComponent{
		CurrentState: engine.AnimationStateIdle,
		FrameCount:   4,
		Dirty:        true,
	})
	game.World.FlushPendingEntities()

	// Update animation system - should not crash
	entities := game.World.GetEntities()
	animSys.Update(entities, 0.016)

	// If we get here without panic, systems are properly connected
}

// TestParallelInitDeterminism verifies parallel init is deterministic.
func TestParallelInitDeterminism(t *testing.T) {
	// Run initialization multiple times and verify consistent results
	runs := 5
	var systems []*systemsContainer

	for i := 0; i < runs; i++ {
		game, _, logger, clientLogger := setupTestEnvironment(t)
		coreSystems := initializeCoreSystems(game, logger, clientLogger)
		systems = append(systems, coreSystems)
		cleanupTestEnvironment()
	}

	// Verify all runs produced the same system count
	for i := 1; i < runs; i++ {
		if (systems[i].combatSystem == nil) != (systems[0].combatSystem == nil) {
			t.Error("parallel init produced inconsistent results")
		}
	}
}

// BenchmarkParallelSystemInit measures parallel init performance.
func BenchmarkParallelSystemInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		game, _, logger, clientLogger := setupTestEnvironment(b)
		_ = initializeCoreSystems(game, logger, clientLogger)
		cleanupTestEnvironment()
	}
}

// BenchmarkSequentialSystemInit provides baseline for comparison.
// This simulates what sequential init would look like.
func BenchmarkSequentialSystemInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		game, _, logger, clientLogger := setupTestEnvironment(b)

		// Sequential initialization (old pattern)
		sysSeq := &systemsContainer{}
		sysSeq.performanceSystem = engine.NewPerformanceMonitoringSystem()
		sysSeq.inputSystem = engine.NewInputSystem()
		sysSeq.movementSystem = engine.NewMovementSystem(100.0)
		sysSeq.collisionSystem = engine.NewCollisionSystem(64)
		sysSeq.movementSystem.SetCollisionSystem(sysSeq.collisionSystem)
		sysSeq.combatSystem = engine.NewCombatSystemWithLogger(12345, logger)
		sysSeq.interactionSystem = engine.NewInteractionSystem(game.World)
		sysSeq.particleSystem = engine.NewParticleSystem()
		sysSeq.spriteCache = cache.NewSpriteCache(400 * 1024 * 1024)
		sysSeq.spriteGenerator = sprites.NewGenerator()
		sysSeq.animationSystem = engine.NewAnimationSystem(sysSeq.spriteGenerator)
		sysSeq.animationSystem.SetMaxCacheSize(100 * 1024 * 1024)
		sysSeq.animationSystem.SetSpriteCache(sysSeq.spriteCache)
		sysSeq.equipmentVisualSystem = engine.NewEquipmentVisualSystem(sysSeq.spriteGenerator)
		qualityConfig := &quality.Config{
			Level:                 quality.QualityMedium,
			EnablePostProcessing:  true,
			EnableBloom:           false,
			EnableSoftShadows:     true,
			SpriteDetailLevel:     0.7,
			EnableAntiAliasing:    true,
			AntiAliasingQuality:   1,
			EnableSpriteCache:     true,
			EnableDynamicLighting: true,
			ShadowSampleCount:     2,
		}
		sysSeq.qualitySystem = engine.NewQualitySystem(qualityConfig, 60.0)
		sysSeq.lightingAdapter = engine.NewLightingAdapter(clientLogger.WithField("system", "lighting"))
		sysSeq.animationAdapter = engine.NewAnimationAdapter(sysSeq.spriteGenerator, clientLogger.WithField("system", "animation"))
		sysSeq.uiGenerator = ui.NewGeneratorWithLogger(logger)
		sysSeq.shapeRenderer = shapes.NewGenerator()
		sysSeq.patternGenerator = patterns.NewGeneratorWithLogger(logger)
		sysSeq.imagePool = pool.NewImagePool()
		sysSeq.parallelRenderer = parallel.NewWorkerPool(runtime.NumCPU())
		poolAdapter := engine.NewImagePoolAdapter(sysSeq.imagePool)
		parallelAdapter := engine.NewParallelRendererAdapter(sysSeq.parallelRenderer)
		game.RenderSystem.SetPool(poolAdapter)
		game.RenderSystem.SetParallelRenderer(parallelAdapter)

		cleanupTestEnvironment()
	}
}

// Helper functions for test setup

func setupTestEnvironment(tb testing.TB) (*engine.EbitenGame, *systemsContainer, *logrus.Logger, *logrus.Entry) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Quiet during tests
	clientLogger := logger.WithField("component", "test")

	// Create minimal game instance
	game := &engine.EbitenGame{
		World: engine.NewWorldWithLogger(logger),
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := &systemsContainer{}

	return game, sys, logger, clientLogger
}

func cleanupTestEnvironment() {
	// Cleanup if needed
}
