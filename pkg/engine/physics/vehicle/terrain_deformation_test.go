package vehicle

import (
	"testing"
)

func TestNewTerrainDeformationComponent(t *testing.T) {
	seed := int64(12345)
	comp := NewTerrainDeformationComponent(seed)

	if comp == nil {
		t.Fatal("NewTerrainDeformationComponent returned nil")
	}
	if comp.Type() != "terrain_deformation" {
		t.Errorf("got type %q, want %q", comp.Type(), "terrain_deformation")
	}
	if comp.Seed != seed {
		t.Errorf("got seed %d, want %d", comp.Seed, seed)
	}
	if comp.MaxTracks != 200 {
		t.Errorf("got MaxTracks=%d, want 200", comp.MaxTracks)
	}
}

func TestTerrainDeformationComponent_AddTrack(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	tests := []struct {
		name         string
		terrainType  TerrainType
		wheelLoad    float64
		shouldCreate bool
	}{
		{"hard terrain no track", TerrainHard, 1000.0, false},
		{"firm terrain creates track", TerrainFirm, 1000.0, true},
		{"soft terrain creates track", TerrainSoft, 1000.0, true},
		{"snow creates track", TerrainSnow, 1000.0, true},
		{"water no permanent track", TerrainWater, 1000.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewTerrainDeformationComponent(12345)

			// Add track
			sys.AddTerrainTrack(comp, 100.0, 100.0, 0.0, tt.wheelLoad, tt.terrainType)

			trackCount := len(comp.Tracks)
			if tt.shouldCreate && trackCount == 0 {
				t.Error("expected track to be created, but count is 0")
			}
			if !tt.shouldCreate && trackCount > 0 {
				t.Error("expected no track, but track was created")
			}
		})
	}
}

func TestTerrainDeformationComponent_TrackSpacing(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add first track
	sys.AddTerrainTrack(comp, 100.0, 100.0, 0.0, 1000.0, TerrainSoft)
	if len(comp.Tracks) != 1 {
		t.Fatalf("first track not created")
	}

	// Try to add track too close (should be rejected)
	sys.AddTerrainTrack(comp, 101.0, 101.0, 0.0, 1000.0, TerrainSoft) // Distance ~1.4 pixels
	if len(comp.Tracks) != 1 {
		t.Error("track should not be created when too close to previous track")
	}

	// Add track far enough away (should succeed)
	sys.AddTerrainTrack(comp, 110.0, 110.0, 0.0, 1000.0, TerrainSoft) // Distance ~14.1 pixels
	if len(comp.Tracks) != 2 {
		t.Error("track should be created when far enough from previous track")
	}
}

func TestTerrainDeformationComponent_Update(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add tracks with different fade times
	sys.AddTerrainTrack(comp, 100.0, 100.0, 0.0, 1000.0, TerrainFirm) // 30s fade
	sys.AddTerrainTrack(comp, 200.0, 200.0, 0.0, 1000.0, TerrainSoft) // 120s fade

	initialCount := len(comp.Tracks)
	if initialCount != 2 {
		t.Fatalf("expected 2 tracks, got %d", initialCount)
	}

	// Age tracks
	sys.UpdateTerrainTracks(comp, 1.0) // 1 second

	// Check ages
	for i := range comp.Tracks {
		if comp.Tracks[i].Age < 0.9 || comp.Tracks[i].Age > 1.1 {
			t.Errorf("track %d age=%f, expected ~1.0", i, comp.Tracks[i].Age)
		}
	}

	// Age past fade time for firm terrain (30s)
	sys.UpdateTerrainTracks(comp, 30.0)

	// Firm track should be removed, soft track should remain
	if len(comp.Tracks) != 1 {
		t.Errorf("expected 1 track after fading, got %d", len(comp.Tracks))
	}
}

func TestTerrainDeformationComponent_GetVisibleTracks(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add tracks at different positions
	sys.AddTerrainTrack(comp, 50.0, 50.0, 0.0, 1000.0, TerrainSoft)
	sys.AddTerrainTrack(comp, 150.0, 150.0, 0.0, 1000.0, TerrainSoft)
	sys.AddTerrainTrack(comp, 250.0, 250.0, 0.0, 1000.0, TerrainSoft)

	// Query viewport (100, 100) to (200, 200)
	visible := GetVisibleTracks(comp, 100.0, 100.0, 200.0, 200.0)

	// Only middle track should be visible
	if len(visible) != 1 {
		t.Errorf("expected 1 visible track, got %d", len(visible))
	}
	if len(visible) > 0 && (visible[0].X != 150.0 || visible[0].Y != 150.0) {
		t.Errorf("wrong track visible: pos=(%f,%f), want (150,150)", visible[0].X, visible[0].Y)
	}
}

func TestTerrainDeformationComponent_GetTrackAlpha(t *testing.T) {
	track := TrackMark{
		Age:      10.0,
		FadeTime: 30.0,
	}

	alpha := GetTrackAlpha(&track)
	expected := 1.0 - (10.0 / 30.0) // ~0.667

	if alpha < expected-0.01 || alpha > expected+0.01 {
		t.Errorf("got alpha=%f, want ~%f", alpha, expected)
	}

	// Test fully faded track
	track.Age = 30.0
	alpha = GetTrackAlpha(&track)
	if alpha != 0.0 {
		t.Errorf("fully faded track should have alpha=0.0, got %f", alpha)
	}
}

func TestTerrainDeformationComponent_Clear(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add some tracks
	sys.AddTerrainTrack(comp, 100.0, 100.0, 0.0, 1000.0, TerrainSoft)
	sys.AddTerrainTrack(comp, 200.0, 200.0, 0.0, 1000.0, TerrainSoft)

	if len(comp.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(comp.Tracks))
	}

	// Clear
	comp.Clear()

	if len(comp.Tracks) != 0 {
		t.Errorf("after clear, expected 0 tracks, got %d", len(comp.Tracks))
	}
}

func TestTerrainDeformationComponent_MaxTracksLimit(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)
	comp.MaxTracks = 10 // Set low limit for testing

	// Add more tracks than the limit
	for i := 0; i < 15; i++ {
		x := float64(i * 10)
		sys.AddTerrainTrack(comp, x, 100.0, 0.0, 1000.0, TerrainSoft)
	}

	// Should not exceed max
	if len(comp.Tracks) > comp.MaxTracks {
		t.Errorf("track count %d exceeds MaxTracks %d", len(comp.Tracks), comp.MaxTracks)
	}
}

func TestTerrainDeformationComponent_Determinism(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	seed := int64(42)

	// Create two components with same seed
	comp1 := NewTerrainDeformationComponent(seed)
	comp2 := NewTerrainDeformationComponent(seed)

	// Add same tracks
	for i := 0; i < 5; i++ {
		x := float64(i * 20)
		sys.AddTerrainTrack(comp1, x, 100.0, 0.0, 1000.0, TerrainSoft)
		sys.AddTerrainTrack(comp2, x, 100.0, 0.0, 1000.0, TerrainSoft)
	}

	// Tracks should be identical
	if comp1.GetTrackCount() != comp2.GetTrackCount() {
		t.Fatalf("track counts differ: %d vs %d", comp1.GetTrackCount(), comp2.GetTrackCount())
	}

	for i := range comp1.Tracks {
		t1 := comp1.Tracks[i]
		t2 := comp2.Tracks[i]

		if t1.X != t2.X || t1.Y != t2.Y {
			t.Errorf("track %d position differs", i)
		}
		if t1.Depth != t2.Depth {
			t.Errorf("track %d depth differs: %f vs %f", i, t1.Depth, t2.Depth)
		}
		if t1.Width != t2.Width {
			t.Errorf("track %d width differs: %f vs %f", i, t1.Width, t2.Width)
		}
	}
}

func TestGetTerrainTypeFromTile(t *testing.T) {
	tests := []struct {
		tileType int
		want     TerrainType
	}{
		{0, TerrainHard},  // Wall
		{1, TerrainHard},  // Floor (hard surface)
		{2, TerrainFirm},  // Corridor
		{3, TerrainFirm},  // Door
		{4, TerrainWater}, // Water shallow
		{5, TerrainWater}, // Water deep
		{11, TerrainSoft}, // Tree (soft ground)
		{99, TerrainFirm}, // Unknown defaults to firm
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			got := GetTerrainTypeFromTile(tt.tileType)
			if got != tt.want {
				t.Errorf("tile %d: got %s, want %s", tt.tileType, got.String(), tt.want.String())
			}
		})
	}
}

// Benchmark tests
func BenchmarkTerrainDeformationComponent_AddTrack(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i % 100 * 10)
		y := float64(i / 100 * 10)
		sys.AddTerrainTrack(comp, x, y, 0.0, 1000.0, TerrainSoft)
	}
}

func BenchmarkTerrainDeformationComponent_Update(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add some tracks
	for i := 0; i < 100; i++ {
		x := float64(i * 10)
		sys.AddTerrainTrack(comp, x, 100.0, 0.0, 1000.0, TerrainSoft)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.UpdateTerrainTracks(comp, 0.016)
	}
}

func BenchmarkTerrainDeformationComponent_GetVisibleTracks(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewTerrainDeformationComponent(12345)

	// Add many tracks
	for i := 0; i < 200; i++ {
		x := float64(i * 10)
		y := float64(i * 10)
		sys.AddTerrainTrack(comp, x, y, 0.0, 1000.0, TerrainSoft)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVisibleTracks(comp, 500.0, 500.0, 700.0, 700.0)
	}
}
