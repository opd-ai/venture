package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/world/territory"
)

// TestNewTerritorySiegeSystem tests system creation.
func TestNewTerritorySiegeSystem(t *testing.T) {
	manager := territory.NewSiegeManager()
	system := NewTerritorySiegeSystem(manager)

	if system == nil {
		t.Fatal("NewTerritorySiegeSystem() returned nil")
	}

	if system.manager != manager {
		t.Error("Manager not set correctly")
	}
}

// TestTerritorySiegeSystem_GetSiegeManager tests manager accessor.
func TestTerritorySiegeSystem_GetSiegeManager(t *testing.T) {
	manager := territory.NewSiegeManager()
	system := NewTerritorySiegeSystem(manager)

	retrieved := system.GetSiegeManager()
	if retrieved != manager {
		t.Error("GetSiegeManager() returned different manager")
	}
}

// TestTerritorySiegeSystem_Update tests update with no entities.
func TestTerritorySiegeSystem_Update(t *testing.T) {
	manager := territory.NewSiegeManager()
	system := NewTerritorySiegeSystem(manager)

	// Should not panic with empty entity list
	entities := make([]*Entity, 0)
	system.Update(entities, 0.016)

	// Should not panic with nil entities
	system.Update(nil, 0.016)
}

// BenchmarkTerritorySiegeSystem_Update benchmarks system update performance.
func BenchmarkTerritorySiegeSystem_Update(b *testing.B) {
	manager := territory.NewSiegeManager()
	system := NewTerritorySiegeSystem(manager)
	entities := make([]*Entity, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
