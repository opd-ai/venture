package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestReverbSystem_SetTerrain(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	// Create small room (8x8 = 64 floor tiles)
	width, height := 8, 8
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileFloor
		}
	}

	system.SetTerrain(width, height, tiles)

	if system.terrainWidth != width {
		t.Errorf("terrainWidth = %v, want %v", system.terrainWidth, width)
	}
	if system.terrainHeight != height {
		t.Errorf("terrainHeight = %v, want %v", system.terrainHeight, height)
	}

	reverb := system.GetCurrentReverb()
	if reverb == nil {
		t.Fatal("Expected reverb to be created after SetTerrain")
	}

	// Small room should have RoomSizeSmall
	if reverb.RoomSize != RoomSizeSmall {
		t.Errorf("RoomSize = %v, want RoomSizeSmall", reverb.RoomSize)
	}
}

func TestReverbSystem_CalculateRoomSize(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	tests := []struct {
		name       string
		floorCount int
		wantSize   RoomSize
	}{
		{"tiny room (50 tiles)", 50, RoomSizeSmall},
		{"small room (99 tiles)", 99, RoomSizeSmall},
		{"medium room (100 tiles)", 100, RoomSizeMedium},
		{"medium room (399 tiles)", 399, RoomSizeMedium},
		{"large room (400 tiles)", 400, RoomSizeLarge},
		{"large room (1199 tiles)", 1199, RoomSizeLarge},
		{"huge room (1200 tiles)", 1200, RoomSizeHuge},
		{"huge room (2000 tiles)", 2000, RoomSizeHuge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := system.calculateRoomSize(tt.floorCount)
			if got != tt.wantSize {
				t.Errorf("calculateRoomSize(%d) = %v, want %v", tt.floorCount, got, tt.wantSize)
			}
		})
	}
}

func TestReverbSystem_AnalyzeMaterials(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	tests := []struct {
		name         string
		tileType     terrain.TileType
		wantMaterial RoomMaterial
	}{
		{"stone walls", terrain.TileWall, RoomMaterialStone},
		{"wooden doors", terrain.TileDoor, RoomMaterialWood},
		{"wooden platforms", terrain.TilePlatform, RoomMaterialWood},
		{"water tiles", terrain.TileWaterShallow, RoomMaterialStone},
		{"lava tiles", terrain.TileLavaFlow, RoomMaterialMetal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create terrain filled with specific tile type
			width, height := 10, 10
			tiles := make([][]terrain.TileType, height)
			for i := range tiles {
				tiles[i] = make([]terrain.TileType, width)
				for j := range tiles[i] {
					tiles[i][j] = tt.tileType
				}
			}

			system.SetTerrain(width, height, tiles)

			reverb := system.GetCurrentReverb()
			if reverb == nil {
				t.Fatal("Expected reverb to be created")
			}

			if reverb.Material != tt.wantMaterial {
				t.Errorf("Material = %v, want %v", reverb.Material, tt.wantMaterial)
			}
		})
	}
}

func TestReverbSystem_RoomSizeAffectsDecayTime(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	tests := []struct {
		name         string
		roomSize     int // width=height for square rooms
		wantDecayMin float64
		wantDecayMax float64
	}{
		{"small (5x5)", 5, 0.2, 0.4},
		{"medium (15x15)", 15, 0.6, 1.0},
		{"large (25x25)", 25, 1.2, 1.8},
		{"huge (40x40)", 40, 2.5, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create square room
			tiles := make([][]terrain.TileType, tt.roomSize)
			for i := range tiles {
				tiles[i] = make([]terrain.TileType, tt.roomSize)
				for j := range tiles[i] {
					tiles[i][j] = terrain.TileFloor
				}
			}

			system.SetTerrain(tt.roomSize, tt.roomSize, tiles)

			reverb := system.GetCurrentReverb()
			if reverb == nil {
				t.Fatal("Expected reverb to be created")
			}

			if reverb.DecayTime < tt.wantDecayMin || reverb.DecayTime > tt.wantDecayMax {
				t.Errorf("DecayTime = %v, want range [%v, %v]",
					reverb.DecayTime, tt.wantDecayMin, tt.wantDecayMax)
			}
		})
	}
}

func TestReverbSystem_MaterialAffectsDamping(t *testing.T) {
	_ = NewWorld() // Create world but don't need to use it

	tests := []struct {
		name           string
		material       RoomMaterial
		wantDampingMin float64
		wantDampingMax float64
	}{
		{"stone (reflective)", RoomMaterialStone, 0.0, 0.2},
		{"wood (moderate)", RoomMaterialWood, 0.2, 0.4},
		{"cloth (absorptive)", RoomMaterialCloth, 0.5, 0.7},
		{"metal (very reflective)", RoomMaterialMetal, 0.0, 0.1},
		{"carpet (very absorptive)", RoomMaterialCarpet, 0.7, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reverb := NewReverbComponent(RoomSizeMedium, tt.material)

			if reverb.Damping < tt.wantDampingMin || reverb.Damping > tt.wantDampingMax {
				t.Errorf("Damping = %v, want range [%v, %v]",
					reverb.Damping, tt.wantDampingMin, tt.wantDampingMax)
			}
		})
	}
}

func TestReverbSystem_GetReverbParameters(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	// Initially no reverb
	decay, damp, amount, enabled := system.GetReverbParameters()
	if enabled {
		t.Error("Reverb should be disabled initially (no terrain)")
	}

	// Set terrain to create reverb
	width, height := 10, 10
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileFloor
		}
	}

	system.SetTerrain(width, height, tiles)

	decay, damp, amount, enabled = system.GetReverbParameters()
	if !enabled {
		t.Error("Reverb should be enabled after terrain set")
	}
	if decay <= 0 {
		t.Errorf("DecayTime should be positive, got %v", decay)
	}
	if damp < 0 || damp > 1 {
		t.Errorf("Damping out of range [0, 1]: %v", damp)
	}
	if amount != 0.5 {
		t.Errorf("Amount = %v, want 0.5 (default)", amount)
	}
}

func TestReverbSystem_ApplyReverbToSamples(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	// Create reverb settings
	width, height := 15, 15
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileFloor
		}
	}
	system.SetTerrain(width, height, tiles)

	// Create test samples (simple pulse)
	sampleRate := 44100
	numSamples := 1000
	samples := make([]float64, numSamples)
	samples[0] = 1.0 // Single impulse

	// Apply reverb
	output := system.ApplyReverbToSamples(samples, sampleRate)

	// Check output
	if len(output) != numSamples {
		t.Errorf("Output length = %v, want %v", len(output), numSamples)
	}

	// Reverb should create a tail (non-zero samples after impulse)
	tailFound := false
	for i := 100; i < 500; i++ {
		if output[i] != 0 {
			tailFound = true
			break
		}
	}
	if !tailFound {
		t.Error("Expected reverb tail after impulse, found none")
	}
}

func TestReverbSystem_ApplyReverb_DisabledReverb(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	// No terrain = no reverb
	samples := []float64{1.0, 0.5, 0.0, -0.5, -1.0}
	output := system.ApplyReverbToSamples(samples, 44100)

	// Should return unchanged
	if len(output) != len(samples) {
		t.Errorf("Output length = %v, want %v", len(output), len(samples))
	}

	for i, s := range samples {
		if output[i] != s {
			t.Errorf("Sample[%d] = %v, want %v (no reverb)", i, output[i], s)
		}
	}
}

func TestReverbSystem_SetMaterialMapping(t *testing.T) {
	world := NewWorld()
	system := NewReverbSystem(world, 12345)

	// Set custom material mapping
	system.SetMaterialMapping(terrain.TileFloor, RoomMaterialCarpet)

	// Create terrain with floors
	width, height := 10, 10
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			tiles[i][j] = terrain.TileFloor
		}
	}

	system.SetTerrain(width, height, tiles)

	reverb := system.GetCurrentReverb()
	if reverb == nil {
		t.Fatal("Expected reverb to be created")
	}

	// Should use custom mapping (carpet instead of stone)
	if reverb.Material != RoomMaterialCarpet {
		t.Errorf("Material = %v, want RoomMaterialCarpet", reverb.Material)
	}
}
