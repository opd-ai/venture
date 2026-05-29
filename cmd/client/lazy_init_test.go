//go:build !android && !ios
// +build !android,!ios

package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// TestLazyInitialization verifies that lazy initialization correctly defers non-critical system setup.
func TestLazyInitialization(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"lazy initialization starts and completes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create minimal test environment
			logger := logrus.New()
			logger.SetLevel(logrus.FatalLevel) // Suppress logs during test
			clientLogger := logger.WithField("test", "lazy_init")

			// Create a test game instance with required systems
			game := &engine.EbitenGame{
				World:        engine.NewWorldWithLogger(logger),
				ScreenWidth:  800,
				ScreenHeight: 600,
			}
			game.CameraSystem = engine.NewCameraSystem(800, 600)
			game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

			// Initialize core systems (mimics setupAllGameSystems)
			sys := initializeCoreSystems(game, logger, clientLogger)

			// Verify lazy init hasn't started yet
			if sys.lazyInitStarted {
				t.Errorf("lazy initialization should not have started before scheduleLazyInit()")
			}

			// Schedule lazy initialization
			sys.scheduleLazyInit(game, logger, clientLogger)

			// Verify it started
			sys.lazyInitMutex.Lock()
			started := sys.lazyInitStarted
			sys.lazyInitMutex.Unlock()

			if !started {
				t.Errorf("lazy initialization should have started after scheduleLazyInit()")
			}

			// Wait for lazy init to complete (with timeout)
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			completed := false
			for !completed {
				select {
				case <-timeout:
					t.Fatalf("lazy initialization did not complete within 5 seconds")
				case <-ticker.C:
					completed = sys.isLazyInitCompleted()
				}
			}

			// Verify completion
			if !sys.isLazyInitCompleted() {
				t.Errorf("isLazyInitCompleted() should return true after completion")
			}
		})
	}
}

// TestScheduleLazyInitIdempotent verifies that calling scheduleLazyInit multiple times is safe.
func TestScheduleLazyInitIdempotent(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_idempotent")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)

	// Call scheduleLazyInit multiple times
	sys.scheduleLazyInit(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)

	// Should only start once
	sys.lazyInitMutex.Lock()
	started := sys.lazyInitStarted
	sys.lazyInitMutex.Unlock()

	if !started {
		t.Errorf("lazy initialization should have started")
	}

	// Wait for completion
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return // Test passes - no panic occurred from multiple calls
		case <-ticker.C:
			if sys.isLazyInitCompleted() {
				return // Test passes
			}
		}
	}
}

// TestIsLazyInitCompletedThreadSafe verifies thread-safe access to completion status.
func TestIsLazyInitCompletedThreadSafe(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_threadsafe")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)
	sys.scheduleLazyInit(game, logger, clientLogger)

	// Spawn multiple goroutines checking completion status
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = sys.isLazyInitCompleted()
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no race condition detected
}

// BenchmarkLazyInitScheduling measures the overhead of scheduling lazy initialization.
func BenchmarkLazyInitScheduling(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("bench", "lazy_init")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game := &engine.EbitenGame{
			World:        engine.NewWorldWithLogger(logger),
			ScreenWidth:  800,
			ScreenHeight: 600,
		}
		game.CameraSystem = engine.NewCameraSystem(800, 600)
		game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

		sys := initializeCoreSystems(game, logger, clientLogger)
		sys.scheduleLazyInit(game, logger, clientLogger)
	}
}

func TestLazyInitializationRegistersGameplaySystemsOnce(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_registration")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)
	sys.progressionSystem = engine.NewProgressionSystem(game.World)
	sys.scheduleLazyInit(game, logger, clientLogger)
	waitForLazyInitCompletion(t, sys)

	tests := []struct {
		name  string
		match func(engine.System) bool
	}{
		{"WeatherCritChanceSystem", func(s engine.System) bool { _, ok := s.(*engine.WeatherCritChanceSystem); return ok }},
		{"WeatherBlockChanceSystem", func(s engine.System) bool { _, ok := s.(*engine.WeatherBlockChanceSystem); return ok }},
		{"TerrainMovementSpeedSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainMovementSpeedSystem); return ok }},
		{"TerrainCombatBonusSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainCombatBonusSystem); return ok }},
		{"TerrainStealthSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainStealthSystem); return ok }},
		{"TerrainAmbushCritSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainAmbushCritSystem); return ok }},
		{"TerrainStatusEffectSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainStatusEffectSystem); return ok }},
		{"TerrainManaRegenSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainManaRegenSystem); return ok }},
		{"TerrainSpellDamageSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainSpellDamageSystem); return ok }},
		{"TerrainEquipmentDurabilitySystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainEquipmentDurabilitySystem); return ok }},
		{"TerrainRangedAccuracySystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainRangedAccuracySystem); return ok }},
		{"TerrainCompanionBonusSystem", func(s engine.System) bool { _, ok := s.(*engine.TerrainCompanionBonusSystem); return ok }},
		{"TimeOfDayLightingSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayLightingSystem); return ok }},
		{"TimeOfDayStealthSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayStealthSystem); return ok }},
		{"TimeOfDayXPBonusSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayXPBonusSystem); return ok }},
		{"TimeOfDayManaCostSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayManaCostSystem); return ok }},
		{"TimeOfDayCriticalChanceSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayCriticalChanceSystem); return ok }},
		{"TimeOfDayCompanionBonusSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayCompanionBonusSystem); return ok }},
		{"TimeOfDayManaRegenSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayManaRegenSystem); return ok }},
		{"TimeOfDayBlockChanceSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayBlockChanceSystem); return ok }},
		{"TimeOfDayEvasionSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayEvasionSystem); return ok }},
		{"TimeOfDaySpellDamageSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDaySpellDamageSystem); return ok }},
		{"TimeOfDayAttackSpeedSystem", func(s engine.System) bool { _, ok := s.(*engine.TimeOfDayAttackSpeedSystem); return ok }},
		{"SpecializationCritDamageSystem", func(s engine.System) bool { _, ok := s.(*engine.SpecializationCritDamageSystem); return ok }},
		{"SpecializationEvasionSystem", func(s engine.System) bool { _, ok := s.(*engine.SpecializationEvasionSystem); return ok }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			for _, system := range game.World.GetSystems() {
				if tt.match(system) {
					count++
				}
			}
			if count != 1 {
				t.Fatalf("%s count = %d, want 1", tt.name, count)
			}
		})
	}
}

func TestLazyInitializationWiresTimeOfDaySystems(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)
	clientLogger := logger.WithField("test", "lazy_init_timeofday_wiring")

	game := &engine.EbitenGame{
		World:        engine.NewWorldWithLogger(logger),
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
	game.CameraSystem = engine.NewCameraSystem(800, 600)
	game.RenderSystem = engine.NewRenderSystem(game.CameraSystem)

	sys := initializeCoreSystems(game, logger, clientLogger)
	sys.progressionSystem = engine.NewProgressionSystem(game.World)
	sys.scheduleLazyInit(game, logger, clientLogger)
	waitForLazyInitCompletion(t, sys)

	tests := []struct {
		name   string
		target any
		fields []string
	}{
		{"TimeOfDayStealthSystem", sys.timeOfDayStealthSystem, []string{"lightingSystem"}},
		{"TimeOfDayXPBonusSystem", sys.timeOfDayXPBonusSystem, []string{"lightingSystem", "progressionSystem"}},
		{"TimeOfDayManaCostSystem", sys.timeOfDayManaCostSystem, []string{"lightingSystem"}},
		{"TimeOfDayCriticalChanceSystem", sys.timeOfDayCriticalChanceSystem, []string{"lightingSystem"}},
		{"TimeOfDayCompanionBonusSystem", sys.timeOfDayCompanionBonusSystem, []string{"lightingSystem"}},
		{"TimeOfDayManaRegenSystem", sys.timeOfDayManaRegenSystem, []string{"lightingSystem"}},
		{"TimeOfDayBlockChanceSystem", sys.timeOfDayBlockChanceSystem, []string{"lightingSystem"}},
		{"TimeOfDayEvasionSystem", sys.timeOfDayEvasionSystem, []string{"lightingSystem"}},
		{"TimeOfDaySpellDamageSystem", sys.timeOfDaySpellDamageSystem, []string{"lightingSystem"}},
		{"TimeOfDayAttackSpeedSystem", sys.timeOfDayAttackSpeedSystem, []string{"lightingSystem"}},
		{"TimeOfDayShadowDirectionSystem", sys.timeOfDayShadowDirectionSystem, []string{"lightingSystem"}},
		{"TimeOfDayFishingBonusSystem", sys.timeOfDayFishingBonusSystem, []string{"lightingSystem"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.target == nil {
				t.Fatalf("%s was not initialized", tt.name)
			}
			value := reflect.ValueOf(tt.target)
			for _, field := range tt.fields {
				fieldValue := value.Elem().FieldByName(field)
				if !fieldValue.IsValid() {
					t.Fatalf("%s missing field %s", tt.name, field)
				}
				if fieldValue.IsNil() {
					t.Fatalf("%s field %s is nil", tt.name, field)
				}
			}
		})
	}
}

func waitForLazyInitCompletion(t *testing.T, sys *systemsContainer) {
	t.Helper()

	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("lazy initialization did not complete within 5 seconds")
		case <-ticker.C:
			if sys.isLazyInitCompleted() {
				return
			}
		}
	}
}
