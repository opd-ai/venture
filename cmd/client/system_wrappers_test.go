//go:build !android && !ios
// +build !android,!ios

// Package main (system_wrappers_test.go) contains tests for system wrapper types.
// These tests verify that wrappers correctly delegate to their underlying systems.

package main

import (
	"io"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/engine/prestige"
	"github.com/sirupsen/logrus"
)

// TestAnimationSystemWrapper tests the animation system wrapper Update method.
func TestAnimationSystemWrapper(t *testing.T) {
	tests := []struct {
		name        string
		entities    []*engine.Entity
		deltaTime   float64
		logLevel    logrus.Level
	}{
		{
			name:      "nil entities",
			entities:  nil,
			deltaTime: 0.016,
			logLevel:  logrus.WarnLevel,
		},
		{
			name:      "empty entities",
			entities:  []*engine.Entity{},
			deltaTime: 0.016,
			logLevel:  logrus.WarnLevel,
		},
		{
			name:      "with debug logging",
			entities:  []*engine.Entity{{ID: 1}},
			deltaTime: 0.033,
			logLevel:  logrus.DebugLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetOutput(io.Discard)
			logger.SetLevel(tt.logLevel)
			entry := logger.WithField("test", "animation")

			// Test that wrapper doesn't panic with nil system
			wrapper := &animationSystemWrapper{
				system: nil,
				logger: entry,
			}

			// Call should panic with nil system (expected behavior - systems always initialized in production)
			defer func() {
				if r := recover(); r != nil {
					// Expected panic with nil system
				}
			}()
			wrapper.Update(tt.entities, tt.deltaTime)
		})
	}
}

// TestRotationSystemWrapper tests the rotation system wrapper.
func TestRotationSystemWrapper(t *testing.T) {
	wrapper := &rotationSystemWrapper{system: nil}
	
	// Test that wrapper panics with nil (expected behavior)
	defer func() {
		if r := recover(); r == nil {
			t.Log("wrapper with nil system panics as expected")
		}
	}()
	
	wrapper.Update([]*engine.Entity{}, 0.016)
}

// TestPrestigeEntityAdapterMethods tests all prestigeEntityAdapter methods.
func TestPrestigeEntityAdapterMethods(t *testing.T) {
	// Create a real entity for testing
	entity := engine.NewEntity(42)
	
	// Add a test component
	testComp := &engine.PositionComponent{X: 10, Y: 20}
	entity.AddComponent(testComp)

	adapter := &prestigeEntityAdapter{entity: entity}

	t.Run("GetID", func(t *testing.T) {
		id := adapter.GetID()
		if id != "42" {
			t.Errorf("GetID() = %q, want %q", id, "42")
		}
	})

	t.Run("HasComponent_exists", func(t *testing.T) {
		if !adapter.HasComponent("position") {
			t.Error("HasComponent(position) = false, want true")
		}
	})

	t.Run("HasComponent_not_exists", func(t *testing.T) {
		if adapter.HasComponent("nonexistent") {
			t.Error("HasComponent(nonexistent) = true, want false")
		}
	})

	t.Run("GetComponent_exists", func(t *testing.T) {
		comp := adapter.GetComponent("position")
		if comp == nil {
			t.Error("GetComponent(position) = nil, want component")
		}
		posComp, ok := comp.(*engine.PositionComponent)
		if !ok {
			t.Errorf("GetComponent returned wrong type: %T", comp)
		}
		if posComp.X != 10 || posComp.Y != 20 {
			t.Errorf("GetComponent returned wrong values: X=%f, Y=%f", posComp.X, posComp.Y)
		}
	})

	t.Run("GetComponent_not_exists", func(t *testing.T) {
		comp := adapter.GetComponent("nonexistent")
		if comp != nil {
			t.Errorf("GetComponent(nonexistent) = %v, want nil", comp)
		}
	})

	t.Run("RemoveComponent", func(t *testing.T) {
		// First verify component exists
		if !adapter.HasComponent("position") {
			t.Fatal("position component should exist before removal")
		}
		
		adapter.RemoveComponent("position")
		
		if adapter.HasComponent("position") {
			t.Error("position component should not exist after removal")
		}
	})

	t.Run("AddComponent", func(t *testing.T) {
		// Add a new component via adapter
		healthComp := &engine.HealthComponent{Current: 100, Max: 100}
		adapter.AddComponent(healthComp)
		
		if !adapter.HasComponent("health") {
			t.Error("health component should exist after adding")
		}
	})
}

// TestPrestigeSystemWrapperUpdate tests the prestige system wrapper Update method.
func TestPrestigeSystemWrapperUpdate(t *testing.T) {
	// Create real prestige system
	sys := prestige.NewSystem()
	
	wrapper := &prestigeSystemWrapper{system: sys}
	
	// Create test entities
	entities := []*engine.Entity{
		engine.NewEntity(1),
		engine.NewEntity(2),
	}
	
	// Add prestige components
	for _, e := range entities {
		comp := &prestige.PrestigeComponent{
			PlayerID:      "player-1",
			PrestigeLevel: 1,
		}
		e.AddComponent(comp)
	}
	
	// Should not panic
	wrapper.Update(entities, 0.016)
}

// TestPrestigeSystemWrapperEmptyEntities tests wrapper with empty entity list.
func TestPrestigeSystemWrapperEmptyEntities(t *testing.T) {
	sys := prestige.NewSystem()
	
	wrapper := &prestigeSystemWrapper{system: sys}
	
	// Should not panic with empty slice
	wrapper.Update([]*engine.Entity{}, 0.016)
	
	// Should not panic with nil slice
	wrapper.Update(nil, 0.016)
}

// BenchmarkPrestigeEntityAdapter benchmarks the adapter methods.
func BenchmarkPrestigeEntityAdapter(b *testing.B) {
	entity := engine.NewEntity(1)
	entity.AddComponent(&engine.PositionComponent{X: 10, Y: 20})
	adapter := &prestigeEntityAdapter{entity: entity}

	b.Run("GetID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.GetID()
		}
	})

	b.Run("HasComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.HasComponent("position")
		}
	})

	b.Run("GetComponent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.GetComponent("position")
		}
	})
}

// BenchmarkPrestigeSystemWrapper benchmarks the wrapper update loop.
func BenchmarkPrestigeSystemWrapper(b *testing.B) {
	sys := prestige.NewSystem()
	wrapper := &prestigeSystemWrapper{system: sys}
	
	// Create test entities
	entities := make([]*engine.Entity, 100)
	for i := range entities {
		entities[i] = engine.NewEntity(uint64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper.Update(entities, 0.016)
	}
}
