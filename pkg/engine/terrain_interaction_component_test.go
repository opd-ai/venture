// Package engine provides tests for terrain interaction components.
// Phase 21.2: Advanced Vehicle Features
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestTerrainInteractionComponent_Type(t *testing.T) {
	comp := NewTerrainInteractionComponent()
	if comp.Type() != "terrain_interaction" {
		t.Errorf("Expected type 'terrain_interaction', got '%s'", comp.Type())
	}
}

func TestTerrainInteractionComponent_UpdateTerrainType(t *testing.T) {
	comp := NewTerrainInteractionComponent()
	comp.ActiveEffects = []uint64{1, 2, 3}

	// Update to same terrain - effects should remain
	comp.UpdateTerrainType(terrain.TileFloor)
	if len(comp.ActiveEffects) != 3 {
		t.Errorf("Expected 3 effects after same terrain update, got %d", len(comp.ActiveEffects))
	}

	// Update to different terrain - effects should clear
	comp.UpdateTerrainType(terrain.TileWaterShallow)
	if len(comp.ActiveEffects) != 0 {
		t.Errorf("Expected 0 effects after terrain change, got %d", len(comp.ActiveEffects))
	}
	if comp.CurrentTerrainType != terrain.TileWaterShallow {
		t.Errorf("Expected terrain type TileWaterShallow, got %v", comp.CurrentTerrainType)
	}
}

func TestTerrainInteractionComponent_CanSpawnEffect(t *testing.T) {
	comp := NewTerrainInteractionComponent()
	comp.EffectSpawnInterval = 0.1
	comp.LastEffectTime = 1.0

	tests := []struct {
		name        string
		currentTime float64
		expected    bool
	}{
		{"too soon", 1.05, false},
		{"just enough time", 1.1, true},
		{"plenty of time", 2.0, true},
		{"before last effect", 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := comp.CanSpawnEffect(tt.currentTime); got != tt.expected {
				t.Errorf("CanSpawnEffect(%f) = %v, want %v", tt.currentTime, got, tt.expected)
			}
		})
	}
}

func TestTerrainInteractionComponent_RecordEffect(t *testing.T) {
	comp := NewTerrainInteractionComponent()
	comp.LastEffectTime = 0.0

	comp.RecordEffect(123, 1.5)

	if len(comp.ActiveEffects) != 1 {
		t.Errorf("Expected 1 effect recorded, got %d", len(comp.ActiveEffects))
	}
	if comp.ActiveEffects[0] != 123 {
		t.Errorf("Expected effect ID 123, got %d", comp.ActiveEffects[0])
	}
	if comp.LastEffectTime != 1.5 {
		t.Errorf("Expected last effect time 1.5, got %f", comp.LastEffectTime)
	}

	// Record another effect
	comp.RecordEffect(456, 2.0)
	if len(comp.ActiveEffects) != 2 {
		t.Errorf("Expected 2 effects recorded, got %d", len(comp.ActiveEffects))
	}
	if comp.LastEffectTime != 2.0 {
		t.Errorf("Expected last effect time 2.0, got %f", comp.LastEffectTime)
	}
}

func TestTerrainInteractionComponent_ClearOldEffects(t *testing.T) {
	comp := NewTerrainInteractionComponent()

	// Add many effects
	for i := uint64(0); i < 150; i++ {
		comp.ActiveEffects = append(comp.ActiveEffects, i)
	}

	if len(comp.ActiveEffects) != 150 {
		t.Fatalf("Setup failed: expected 150 effects, got %d", len(comp.ActiveEffects))
	}

	comp.ClearOldEffects()

	if len(comp.ActiveEffects) != 50 {
		t.Errorf("Expected 50 effects after clearing, got %d", len(comp.ActiveEffects))
	}

	// Verify we kept the most recent effects (100-149)
	expectedFirst := uint64(100)
	if comp.ActiveEffects[0] != expectedFirst {
		t.Errorf("Expected first effect to be %d, got %d", expectedFirst, comp.ActiveEffects[0])
	}

	// Test with fewer than threshold
	comp2 := NewTerrainInteractionComponent()
	for i := uint64(0); i < 50; i++ {
		comp2.ActiveEffects = append(comp2.ActiveEffects, i)
	}

	comp2.ClearOldEffects()
	if len(comp2.ActiveEffects) != 50 {
		t.Errorf("Expected 50 effects (unchanged), got %d", len(comp2.ActiveEffects))
	}
}

func TestGetEffectColorForTerrain(t *testing.T) {
	tests := []struct {
		name    string
		terrain terrain.TileType
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantA   uint8
	}{
		{"shallow water", terrain.TileWaterShallow, 100, 150, 255, 200},
		{"deep water", terrain.TileWaterDeep, 100, 150, 255, 200},
		{"tree", terrain.TileTree, 139, 90, 43, 180},
		{"floor", terrain.TileFloor, 180, 180, 160, 150},
		{"wall", terrain.TileWall, 180, 180, 160, 150},
		{"corridor", terrain.TileCorridor, 180, 180, 160, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, a := GetEffectColorForTerrain(tt.terrain)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB || a != tt.wantA {
				t.Errorf("GetEffectColorForTerrain(%v) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					tt.terrain, r, g, b, a, tt.wantR, tt.wantG, tt.wantB, tt.wantA)
			}
		})
	}
}

func TestGetEffectTypeForTerrain(t *testing.T) {
	tests := []struct {
		terrain  terrain.TileType
		expected string
	}{
		{terrain.TileWaterShallow, "splash_small"},
		{terrain.TileWaterDeep, "splash_large"},
		{terrain.TileTree, "leaves"},
		{terrain.TileFloor, "dust"},
		{terrain.TileWall, "dust"},
		{terrain.TileCorridor, "dust"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := GetEffectTypeForTerrain(tt.terrain)
			if got != tt.expected {
				t.Errorf("GetEffectTypeForTerrain(%v) = %v, want %v",
					tt.terrain, got, tt.expected)
			}
		})
	}
}
