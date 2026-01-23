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
			comp.AddTrack(100.0, 100.0, 0.0, tt.wheelLoad, tt.terrainType)

			trackCount := comp.GetTrackCount()
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
	comp := NewTerrainDeformationComponent(12345)

	// Add first track
	comp.AddTrack(100.0, 100.0, 0.0, 1000.0, TerrainSoft)
	if comp.GetTrackCount() != 1 {
		t.Fatalf("first track not created")
	}

	// Try to add track too close (should be rejected)
	comp.AddTrack(101.0, 101.0, 0.0, 1000.0, TerrainSoft) // Distance ~1.4 pixels
	if comp.GetTrackCount() != 1 {
		t.Error("track should not be created when too close to previous track")
	}

	// Add track far enough away (should succeed)
	comp.AddTrack(110.0, 110.0, 0.0, 1000.0, TerrainSoft) // Distance ~14.1 pixels
	if comp.GetTrackCount() != 2 {
		t.Error("track should be created when far enough from previous track")
	}
}

func TestTerrainDeformationComponent_Update(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)

	// Add tracks with different fade times
	comp.AddTrack(100.0, 100.0, 0.0, 1000.0, TerrainFirm) // 30s fade
	comp.AddTrack(200.0, 200.0, 0.0, 1000.0, TerrainSoft) // 120s fade

	initialCount := comp.GetTrackCount()
	if initialCount != 2 {
		t.Fatalf("expected 2 tracks, got %d", initialCount)
	}

	// Age tracks
	comp.Update(1.0) // 1 second

	// Check ages
	for i := range comp.Tracks {
		if comp.Tracks[i].Age < 0.9 || comp.Tracks[i].Age > 1.1 {
			t.Errorf("track %d age=%f, expected ~1.0", i, comp.Tracks[i].Age)
		}
	}

	// Age past fade time for firm terrain (30s)
	comp.Update(30.0)

	// Firm track should be removed, soft track should remain
	if comp.GetTrackCount() != 1 {
		t.Errorf("expected 1 track after fading, got %d", comp.GetTrackCount())
	}
}

func TestTerrainDeformationComponent_GetVisibleTracks(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)

	// Add tracks at different positions
	comp.AddTrack(50.0, 50.0, 0.0, 1000.0, TerrainSoft)
	comp.AddTrack(150.0, 150.0, 0.0, 1000.0, TerrainSoft)
	comp.AddTrack(250.0, 250.0, 0.0, 1000.0, TerrainSoft)

	// Query viewport (100, 100) to (200, 200)
	visible := comp.GetVisibleTracks(100.0, 100.0, 200.0, 200.0)

	// Only middle track should be visible
	if len(visible) != 1 {
		t.Errorf("expected 1 visible track, got %d", len(visible))
	}
	if len(visible) > 0 && (visible[0].X != 150.0 || visible[0].Y != 150.0) {
		t.Errorf("wrong track visible: pos=(%f,%f), want (150,150)", visible[0].X, visible[0].Y)
	}
}

func TestTerrainDeformationComponent_GetTrackAlpha(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)

	track := TrackMark{
		Age:      10.0,
		FadeTime: 30.0,
	}

	alpha := comp.GetTrackAlpha(&track)
	expected := 1.0 - (10.0 / 30.0) // ~0.667

	if alpha < expected-0.01 || alpha > expected+0.01 {
		t.Errorf("got alpha=%f, want ~%f", alpha, expected)
	}

	// Test fully faded track
	track.Age = 30.0
	alpha = comp.GetTrackAlpha(&track)
	if alpha != 0.0 {
		t.Errorf("fully faded track should have alpha=0.0, got %f", alpha)
	}
}

func TestTerrainDeformationComponent_Clear(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)

	// Add some tracks
	comp.AddTrack(100.0, 100.0, 0.0, 1000.0, TerrainSoft)
	comp.AddTrack(200.0, 200.0, 0.0, 1000.0, TerrainSoft)

	if comp.GetTrackCount() != 2 {
		t.Fatalf("expected 2 tracks, got %d", comp.GetTrackCount())
	}

	// Clear
	comp.Clear()

	if comp.GetTrackCount() != 0 {
		t.Errorf("after clear, expected 0 tracks, got %d", comp.GetTrackCount())
	}
}

func TestTerrainDeformationComponent_MaxTracksLimit(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)
	comp.MaxTracks = 10 // Set low limit for testing

	// Add more tracks than the limit
	for i := 0; i < 15; i++ {
		x := float64(i * 10)
		comp.AddTrack(x, 100.0, 0.0, 1000.0, TerrainSoft)
	}

	// Should not exceed max
	if comp.GetTrackCount() > comp.MaxTracks {
		t.Errorf("track count %d exceeds MaxTracks %d", comp.GetTrackCount(), comp.MaxTracks)
	}
}

func TestTerrainDeformationComponent_Determinism(t *testing.T) {
	seed := int64(42)

	// Create two components with same seed
	comp1 := NewTerrainDeformationComponent(seed)
	comp2 := NewTerrainDeformationComponent(seed)

	// Add same tracks
	for i := 0; i < 5; i++ {
		x := float64(i * 20)
		comp1.AddTrack(x, 100.0, 0.0, 1000.0, TerrainSoft)
		comp2.AddTrack(x, 100.0, 0.0, 1000.0, TerrainSoft)
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
	comp := NewTerrainDeformationComponent(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64(i % 100 * 10)
		y := float64(i / 100 * 10)
		comp.AddTrack(x, y, 0.0, 1000.0, TerrainSoft)
	}
}

func BenchmarkTerrainDeformationComponent_Update(b *testing.B) {
	comp := NewTerrainDeformationComponent(12345)

	// Add some tracks
	for i := 0; i < 100; i++ {
		x := float64(i * 10)
		comp.AddTrack(x, 100.0, 0.0, 1000.0, TerrainSoft)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Update(0.016)
	}
}

func BenchmarkTerrainDeformationComponent_GetVisibleTracks(b *testing.B) {
	comp := NewTerrainDeformationComponent(12345)

	// Add many tracks
	for i := 0; i < 200; i++ {
		x := float64(i * 10)
		y := float64(i * 10)
		comp.AddTrack(x, y, 0.0, 1000.0, TerrainSoft)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetVisibleTracks(500.0, 500.0, 700.0, 700.0)
	}
}

// TestTerrainDeformationComponent_GetVisibleTracks_BufferReuse verifies that
// the buffer reuse optimization works correctly across multiple calls.
// Note: Callers must use the returned slice immediately and not store it,
// as it will be overwritten by the next call to GetVisibleTracks.
func TestTerrainDeformationComponent_GetVisibleTracks_BufferReuse(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)

	// Add tracks at different positions
	comp.AddTrack(50.0, 50.0, 0.0, 1000.0, TerrainSoft)
	comp.AddTrack(150.0, 150.0, 0.0, 1000.0, TerrainSoft)
	comp.AddTrack(250.0, 250.0, 0.0, 1000.0, TerrainSoft)

	// First call - should return middle track
	visible1 := comp.GetVisibleTracks(100.0, 100.0, 200.0, 200.0)
	if len(visible1) != 1 {
		t.Errorf("first call: expected 1 visible track, got %d", len(visible1))
	}
	if visible1[0].X != 150.0 || visible1[0].Y != 150.0 {
		t.Errorf("first call: wrong track position (%f,%f), want (150,150)", visible1[0].X, visible1[0].Y)
	}

	// Second call with different bounds - should return last track
	visible2 := comp.GetVisibleTracks(200.0, 200.0, 300.0, 300.0)
	if len(visible2) != 1 {
		t.Errorf("second call: expected 1 visible track, got %d", len(visible2))
	}
	if visible2[0].X != 250.0 || visible2[0].Y != 250.0 {
		t.Errorf("second call: wrong track position (%f,%f), want (250,250)", visible2[0].X, visible2[0].Y)
	}

	// Third call with bounds that includes all tracks
	visible3 := comp.GetVisibleTracks(0.0, 0.0, 300.0, 300.0)
	if len(visible3) != 3 {
		t.Errorf("third call: expected 3 visible tracks, got %d", len(visible3))
	}

	// Verify positions are correct
	expectedPositions := []struct{ x, y float64 }{
		{50.0, 50.0},
		{150.0, 150.0},
		{250.0, 250.0},
	}
	for i, expected := range expectedPositions {
		if i >= len(visible3) {
			break
		}
		if visible3[i].X != expected.x || visible3[i].Y != expected.y {
			t.Errorf("track %d: got position (%f,%f), want (%f,%f)",
				i, visible3[i].X, visible3[i].Y, expected.x, expected.y)
		}
	}

	// Fourth call with empty bounds - should return no tracks
	visible4 := comp.GetVisibleTracks(500.0, 500.0, 600.0, 600.0)
	if len(visible4) != 0 {
		t.Errorf("fourth call: expected 0 visible tracks, got %d", len(visible4))
	}
}
