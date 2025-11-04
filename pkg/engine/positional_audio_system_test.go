package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

func TestPositionalAudioSystem_SetListener(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)

	system.SetListener(100, 200)

	if system.listenerX != 100 {
		t.Errorf("listenerX = %v, want 100", system.listenerX)
	}
	if system.listenerY != 200 {
		t.Errorf("listenerY = %v, want 200", system.listenerY)
	}
}

func TestPositionalAudioSystem_SetPerformanceMode(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)

	if system.performanceMode {
		t.Error("performanceMode should be false by default")
	}

	system.SetPerformanceMode(true)
	if !system.performanceMode {
		t.Error("performanceMode should be true after SetPerformanceMode(true)")
	}

	system.SetPerformanceMode(false)
	if system.performanceMode {
		t.Error("performanceMode should be false after SetPerformanceMode(false)")
	}
}

func TestPositionalAudioSystem_Update(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(500, 500)

	// Create entity with positional audio at (600, 500) - 100 pixels to the right
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 600, Y: 500})
	entity.AddComponent(NewPositionalAudioComponent(1000))

	// Update should calculate volume and pan
	system.Update(0.016)

	comp, ok := entity.GetComponent("positional_audio")
	if !ok {
		t.Fatal("Component positional_audio not found")
	}
	audioComp, ok := comp.(*PositionalAudioComponent)
	if !ok {
		t.Fatal("Failed to convert to PositionalAudioComponent")
	}

	// Check position updated
	if audioComp.X != 600 || audioComp.Y != 500 {
		t.Errorf("Position not updated: got (%v, %v), want (600, 500)", audioComp.X, audioComp.Y)
	}

	// Check volume (should be high, close to source)
	if audioComp.VolumeMultiplier < 0.9 {
		t.Errorf("Volume too low: %v, expected >0.9", audioComp.VolumeMultiplier)
	}

	// Check pan (should be positive, source is to the right)
	if audioComp.StereoPan <= 0 {
		t.Errorf("Pan should be positive (right), got %v", audioComp.StereoPan)
	}
}

func TestPositionalAudioSystem_DistanceFalloff(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(0, 0)

	tests := []struct {
		name        string
		distance    float64
		maxDistance float64
		wantLow     float64
		wantHigh    float64
	}{
		{
			name:        "very close (50px)",
			distance:    50,
			maxDistance: 1000,
			wantLow:     0.95,
			wantHigh:    1.0,
		},
		{
			name:        "medium distance (500px)",
			distance:    500,
			maxDistance: 1000,
			wantLow:     0.6,
			wantHigh:    0.85,
		},
		{
			name:        "far distance (900px)",
			distance:    900,
			maxDistance: 1000,
			wantLow:     0.0,
			wantHigh:    0.2,
		},
		{
			name:        "at max distance (1000px)",
			distance:    1000,
			maxDistance: 1000,
			wantLow:     0.0,
			wantHigh:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: tt.distance, Y: 0})
			entity.AddComponent(NewPositionalAudioComponent(tt.maxDistance))

			system.Update(0.016)

			comp, ok := entity.GetComponent("positional_audio")
			if !ok {
				t.Fatal("Component positional_audio not found")
			}
			audioComp, ok := comp.(*PositionalAudioComponent)
			if !ok {
				t.Fatal("Failed to convert to PositionalAudioComponent")
			}
			if audioComp.VolumeMultiplier < tt.wantLow || audioComp.VolumeMultiplier > tt.wantHigh {
				t.Errorf("Volume = %v, want range [%v, %v]", audioComp.VolumeMultiplier, tt.wantLow, tt.wantHigh)
			}
		})
	}
}

func TestPositionalAudioSystem_StereoPanning(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(500, 500)

	tests := []struct {
		name    string
		x       float64
		y       float64
		wantPan string // "left", "center", "right"
	}{
		{
			name:    "far left",
			x:       0,
			y:       500,
			wantPan: "left",
		},
		{
			name:    "center",
			x:       500,
			y:       500,
			wantPan: "center",
		},
		{
			name:    "far right",
			x:       1000,
			y:       500,
			wantPan: "right",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: tt.x, Y: tt.y})
			entity.AddComponent(NewPositionalAudioComponent(1000))

			system.Update(0.016)

			comp, ok := entity.GetComponent("positional_audio")
			if !ok {
				t.Fatal("Component positional_audio not found")
			}
			audioComp, ok := comp.(*PositionalAudioComponent)
			if !ok {
				t.Fatal("Failed to convert to PositionalAudioComponent")
			}
			switch tt.wantPan {
			case "left":
				if audioComp.StereoPan >= -0.1 {
					t.Errorf("Expected left pan (<-0.1), got %v", audioComp.StereoPan)
				}
			case "center":
				if audioComp.StereoPan < -0.1 || audioComp.StereoPan > 0.1 {
					t.Errorf("Expected center pan ([-0.1, 0.1]), got %v", audioComp.StereoPan)
				}
			case "right":
				if audioComp.StereoPan <= 0.1 {
					t.Errorf("Expected right pan (>0.1), got %v", audioComp.StereoPan)
				}
			}
		})
	}
}

func TestPositionalAudioSystem_Occlusion(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(100, 100)

	// Create simple terrain with walls
	width, height := 10, 10
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			if i == 5 { // Wall at y=5
				tiles[i][j] = terrain.TileWall
			} else {
				tiles[i][j] = terrain.TileFloor
			}
		}
	}
	system.SetTerrain(width, height, tiles)

	tests := []struct {
		name         string
		entityX      float64
		entityY      float64
		wantOccluded bool
	}{
		{
			name:         "same side as listener (no occlusion)",
			entityX:      150,
			entityY:      100,
			wantOccluded: false,
		},
		{
			name:         "across wall (occluded)",
			entityX:      100,
			entityY:      250, // Across the wall at y=5*32=160
			wantOccluded: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&PositionComponent{X: tt.entityX, Y: tt.entityY})
			audioComp := NewPositionalAudioComponent(1000)
			audioComp.OcclusionEnabled = true
			entity.AddComponent(audioComp)

			system.Update(0.016)

			comp, ok := entity.GetComponent("positional_audio")
			if !ok {
				t.Fatal("Component positional_audio not found")
			}
			audioComp, ok = comp.(*PositionalAudioComponent)
			if !ok {
				t.Fatal("Failed to convert to PositionalAudioComponent")
			}
			if audioComp.IsOccluded != tt.wantOccluded {
				t.Errorf("IsOccluded = %v, want %v", audioComp.IsOccluded, tt.wantOccluded)
			}

			// If occluded, volume should be reduced
			if tt.wantOccluded && audioComp.VolumeMultiplier > audioComp.OcclusionFactor {
				t.Errorf("Occluded volume should be reduced, got %v", audioComp.VolumeMultiplier)
			}
		})
	}
}

func TestPositionalAudioSystem_GetAudioParameters(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(100, 100)

	// Create entity with positional audio
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 200, Y: 100})
	entity.AddComponent(NewPositionalAudioComponent(1000))

	system.Update(0.016)

	volume, pan, ok := system.GetAudioParameters(entity)
	if !ok {
		t.Fatal("GetAudioParameters returned ok=false for entity with component")
	}

	if volume <= 0 || volume > 1.0 {
		t.Errorf("Volume out of range: %v", volume)
	}

	if pan < -1.0 || pan > 1.0 {
		t.Errorf("Pan out of range: %v", pan)
	}

	// Test entity without component
	entity2 := world.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 300, Y: 100})

	_, _, ok = system.GetAudioParameters(entity2)
	if ok {
		t.Error("GetAudioParameters returned ok=true for entity without component")
	}
}

func TestPositionalAudioSystem_PerformanceMode(t *testing.T) {
	world := NewWorld()
	system := NewPositionalAudioSystem(world)
	system.SetListener(100, 100)

	// Create terrain with wall
	width, height := 10, 10
	tiles := make([][]terrain.TileType, height)
	for i := range tiles {
		tiles[i] = make([]terrain.TileType, width)
		for j := range tiles[i] {
			if i == 5 {
				tiles[i][j] = terrain.TileWall
			} else {
				tiles[i][j] = terrain.TileFloor
			}
		}
	}
	system.SetTerrain(width, height, tiles)

	// Create entity across wall
	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 250})
	audioComp := NewPositionalAudioComponent(1000)
	audioComp.OcclusionEnabled = true
	entity.AddComponent(audioComp)

	// With performance mode off, should detect occlusion
	system.SetPerformanceMode(false)
	system.Update(0.016)
	comp, ok := entity.GetComponent("positional_audio")
	if !ok {
		t.Fatal("Component positional_audio not found")
	}
	audioComp, ok = comp.(*PositionalAudioComponent)
	if !ok {
		t.Fatal("Failed to convert to PositionalAudioComponent")
	}
	if !audioComp.IsOccluded {
		t.Error("Should detect occlusion with performance mode off")
	}

	// With performance mode on, should skip occlusion
	system.SetPerformanceMode(true)
	system.Update(0.016)
	comp, ok = entity.GetComponent("positional_audio")
	if !ok {
		t.Fatal("Component positional_audio not found")
	}
	audioComp, ok = comp.(*PositionalAudioComponent)
	if !ok {
		t.Fatal("Failed to convert to PositionalAudioComponent")
	}
	if audioComp.IsOccluded {
		t.Error("Should not detect occlusion with performance mode on")
	}
}
