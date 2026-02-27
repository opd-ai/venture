package vehicle

import (
	"testing"

	procvehicle "github.com/opd-ai/venture/pkg/procgen/vehicle"
)

func TestGetVisibleTracks_SmallCount_AABB(t *testing.T) {
	// Test AABB culling path (<= 200 tracks)
	comp := &TerrainDeformationComponent{
		MaxTracks: 200,
		Tracks:    make([]TrackMark, 0, 200),
	}

	// Add tracks in a grid pattern
	for x := 0.0; x < 10.0; x += 1.0 {
		for y := 0.0; y < 10.0; y += 1.0 {
			comp.Tracks = append(comp.Tracks, TrackMark{
				X:        x * 10.0,
				Y:        y * 10.0,
				Depth:    0.1,
				Width:    0.2,
				Age:      0.0,
				FadeTime: 60.0,
			})
		}
	}
	// Total: 100 tracks (< 200 threshold)

	// Query viewport [20, 20] to [50, 50]
	visible := GetVisibleTracks(comp, 20.0, 20.0, 50.0, 50.0)

	// Expected: tracks at (20,20), (20,30), (20,40), (20,50),
	//                     (30,20), (30,30), (30,40), (30,50),
	//                     (40,20), (40,30), (40,40), (40,50),
	//                     (50,20), (50,30), (50,40), (50,50)
	// = 4x4 = 16 tracks
	if len(visible) != 16 {
		t.Errorf("expected 16 visible tracks, got %d", len(visible))
	}

	// Verify all visible tracks are within bounds
	for _, track := range visible {
		if track.X < 20.0 || track.X > 50.0 || track.Y < 20.0 || track.Y > 50.0 {
			t.Errorf("track (%f, %f) outside bounds [20, 50]", track.X, track.Y)
		}
	}
}

func TestGetVisibleTracks_LargeCount_SpatialHash(t *testing.T) {
	// Test spatial hash grid path (> 200 tracks)
	comp := &TerrainDeformationComponent{
		MaxTracks: 500,
		Tracks:    make([]TrackMark, 0, 500),
	}

	// Add 250 tracks in a grid pattern (> 200 threshold)
	for x := 0.0; x < 25.0; x += 1.0 {
		for y := 0.0; y < 10.0; y += 1.0 {
			comp.Tracks = append(comp.Tracks, TrackMark{
				X:        x * 10.0,
				Y:        y * 10.0,
				Depth:    0.1,
				Width:    0.2,
				Age:      0.0,
				FadeTime: 60.0,
			})
		}
	}
	// Total: 250 tracks (> 200 threshold, triggers spatial hash)

	// Query viewport [100, 50] to [150, 80]
	visible := GetVisibleTracks(comp, 100.0, 50.0, 150.0, 80.0)

	// Expected: tracks at x=[100,110,120,130,140,150] (6 cols) and y=[50,60,70,80] (4 rows)
	// = 6x4 = 24 tracks
	if len(visible) != 24 {
		t.Errorf("expected 24 visible tracks with spatial hash, got %d", len(visible))
	}

	// Verify all visible tracks are within bounds
	for _, track := range visible {
		if track.X < 100.0 || track.X > 150.0 || track.Y < 50.0 || track.Y > 80.0 {
			t.Errorf("track (%f, %f) outside bounds [100,150]x[50,80]", track.X, track.Y)
		}
	}
}

func TestGetVisibleTracks_BufferReuse(t *testing.T) {
	// Verify buffer reuse across multiple queries
	comp := &TerrainDeformationComponent{
		MaxTracks: 100,
		Tracks:    make([]TrackMark, 0, 100),
	}

	for i := 0; i < 50; i++ {
		comp.Tracks = append(comp.Tracks, TrackMark{
			X:        float64(i),
			Y:        float64(i),
			Depth:    0.1,
			Width:    0.2,
			Age:      0.0,
			FadeTime: 60.0,
		})
	}

	// First query
	visible1 := GetVisibleTracks(comp, 10.0, 10.0, 20.0, 20.0)
	if comp.visibleBuffer == nil {
		t.Error("visibleBuffer should be allocated after first query")
	}
	cap1 := cap(comp.visibleBuffer)

	// Second query (buffer should be reused, not reallocated)
	visible2 := GetVisibleTracks(comp, 30.0, 30.0, 40.0, 40.0)
	cap2 := cap(comp.visibleBuffer)

	if cap2 != cap1 {
		t.Errorf("buffer capacity changed (reallocation): %d -> %d", cap1, cap2)
	}

	if len(visible1) == 0 || len(visible2) == 0 {
		t.Error("queries should return non-empty results")
	}
}

func TestCreatePhysicsComponents_Mount(t *testing.T) {
	vehicle := &procvehicle.Vehicle{
		Name:          "Test Horse",
		VehicleType:   procvehicle.TypeMount,
		Handling:      100.0,
		MaxDurability: 100.0,
		CargoWeight:   50.0,
	}

	susp, wt, coll, terrain := CreatePhysicsComponents(vehicle)

	// Mounts should have 4 wheels (legs)
	if susp == nil {
		t.Fatal("suspension component should not be nil for mount")
	}
	if len(susp.Wheels) != 4 {
		t.Errorf("mount should have 4 wheels, got %d", len(susp.Wheels))
	}

	// Weight transfer always present
	if wt == nil {
		t.Fatal("weight transfer component should not be nil")
	}

	// Collision response always present
	if coll == nil {
		t.Fatal("collision response component should not be nil")
	}
	if coll.StructuralIntegrity != 1.0 {
		t.Errorf("new vehicle should have full integrity, got %f", coll.StructuralIntegrity)
	}

	// Mounts should have terrain deformation (hoofprints)
	if terrain == nil {
		t.Error("mount should have terrain deformation component")
	}
}

func TestCreatePhysicsComponents_Cart(t *testing.T) {
	vehicle := &procvehicle.Vehicle{
		Name:          "Test Cart",
		VehicleType:   procvehicle.TypeCart,
		Handling:      80.0,
		MaxDurability: 150.0,
		CargoWeight:   200.0,
	}

	susp, _, _, terrain := CreatePhysicsComponents(vehicle)

	if susp == nil {
		t.Fatal("suspension component should not be nil for cart")
	}
	if len(susp.Wheels) != 4 {
		t.Errorf("cart should have 4 wheels, got %d", len(susp.Wheels))
	}

	// Cart should have terrain deformation (wheel tracks)
	if terrain == nil {
		t.Error("cart should have terrain deformation component")
	}
}

func TestCreatePhysicsComponents_Boat(t *testing.T) {
	vehicle := &procvehicle.Vehicle{
		Name:          "Test Boat",
		VehicleType:   procvehicle.TypeBoat,
		Handling:      60.0,
		MaxDurability: 200.0,
		CargoWeight:   500.0,
	}

	susp, wt, coll, terrain := CreatePhysicsComponents(vehicle)

	// Boats have no wheels
	if susp != nil {
		t.Error("boat should not have suspension component (no wheels)")
	}

	// Weight transfer still needed for motion dynamics
	if wt == nil {
		t.Fatal("boat should have weight transfer component")
	}

	// Collision response needed
	if coll == nil {
		t.Fatal("boat should have collision response component")
	}

	// Boats have no terrain deformation (water instead)
	if terrain != nil {
		t.Error("boat should not have terrain deformation component")
	}
}

func TestCreatePhysicsComponents_Glider(t *testing.T) {
	vehicle := &procvehicle.Vehicle{
		Name:          "Test Glider",
		VehicleType:   procvehicle.TypeGlider,
		Handling:      120.0,
		MaxDurability: 50.0,
		CargoWeight:   10.0,
	}

	susp, wt, coll, terrain := CreatePhysicsComponents(vehicle)

	// Gliders have no wheels
	if susp != nil {
		t.Error("glider should not have suspension component")
	}

	// Weight transfer needed for aerodynamic physics
	if wt == nil {
		t.Fatal("glider should have weight transfer component")
	}

	// Collision response needed
	if coll == nil {
		t.Fatal("glider should have collision response component")
	}

	// Gliders have no terrain deformation (airborne)
	if terrain != nil {
		t.Error("glider should not have terrain deformation component")
	}

	// Gliders should have relatively low damage threshold (fragile)
	// Formula: 5.0 + (durability/100.0 * 10.0) = 5.0 + (50/100 * 10) = 5.0 + 5.0 = 10.0
	// This is acceptable for a fragile glider
	if coll.DamageThreshold < 5.0 {
		t.Errorf("glider damage threshold too low: %f", coll.DamageThreshold)
	}
}

func TestCreatePhysicsComponents_Mech(t *testing.T) {
	vehicle := &procvehicle.Vehicle{
		Name:          "Test Mech",
		VehicleType:   procvehicle.TypeMech,
		Handling:      90.0,
		MaxDurability: 300.0,
		CargoWeight:   100.0,
	}

	susp, wt, coll, terrain := CreatePhysicsComponents(vehicle)

	// Mechs have 2 wheels (feet for bipedal)
	if susp == nil {
		t.Fatal("mech should have suspension component")
	}
	if len(susp.Wheels) != 2 {
		t.Errorf("mech should have 2 wheels (bipedal), got %d", len(susp.Wheels))
	}

	// Weight transfer needed
	if wt == nil {
		t.Fatal("mech should have weight transfer component")
	}

	// Collision response needed
	if coll == nil {
		t.Fatal("mech should have collision response component")
	}

	// Mechs should have high damage threshold (armored)
	if coll.DamageThreshold <= 5.0 {
		t.Errorf("mech should have high damage threshold, got %f", coll.DamageThreshold)
	}

	// Mechs have terrain deformation (footprints)
	if terrain == nil {
		t.Error("mech should have terrain deformation component")
	}

	// Heavy mechs should have deeper tracks (check soft terrain depth)
	if terrain != nil && terrain.DeformationDepth[TerrainSoft] < 0.8 {
		t.Errorf("mech should have deep tracks in soft terrain, got %f", terrain.DeformationDepth[TerrainSoft])
	}
}

func TestCreatePhysicsComponents_HandlingScaling(t *testing.T) {
	// Test that handling stat affects spring stiffness
	tests := []struct {
		name      string
		handling  float64
		expectMin float64
		expectMax float64
	}{
		{"low handling", 50.0, 5000.0, 10000.0},
		{"medium handling", 100.0, 10000.0, 20000.0},
		{"high handling", 200.0, 20000.0, 40000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle := &procvehicle.Vehicle{
				VehicleType:   procvehicle.TypeCart,
				Handling:      tt.handling,
				MaxDurability: 100.0,
				CargoWeight:   0.0,
			}

			susp, _, _, _ := CreatePhysicsComponents(vehicle)
			if susp == nil || len(susp.Wheels) == 0 {
				t.Fatal("suspension should be created")
			}

			stiffness := susp.Wheels[0].SpringStiffness
			if stiffness < tt.expectMin || stiffness > tt.expectMax {
				t.Errorf("spring stiffness %f outside expected range [%f, %f]",
					stiffness, tt.expectMin, tt.expectMax)
			}
		})
	}
}

func TestCreatePhysicsComponents_DurabilityScaling(t *testing.T) {
	// Test that durability stat affects damage threshold
	// Formula: 5.0 + (durability / 100.0 * 10.0)
	tests := []struct {
		name       string
		durability float64
		expectMin  float64
		expectMax  float64
	}{
		{"low durability", 50.0, 9.0, 11.0},      // 5.0 + 5.0 = 10.0
		{"medium durability", 100.0, 14.0, 16.0}, // 5.0 + 10.0 = 15.0
		{"high durability", 200.0, 24.0, 26.0},   // 5.0 + 20.0 = 25.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vehicle := &procvehicle.Vehicle{
				VehicleType:   procvehicle.TypeCart,
				Handling:      100.0,
				MaxDurability: tt.durability,
				CargoWeight:   0.0,
			}

			_, _, coll, _ := CreatePhysicsComponents(vehicle)
			if coll == nil {
				t.Fatal("collision component should be created")
			}

			if coll.DamageThreshold < tt.expectMin || coll.DamageThreshold > tt.expectMax {
				t.Errorf("damage threshold %f outside expected range [%f, %f]",
					coll.DamageThreshold, tt.expectMin, tt.expectMax)
			}
		})
	}
}

// Benchmark spatial hash vs AABB for large track counts
func BenchmarkGetVisibleTracks_AABB_100(b *testing.B) {
	comp := &TerrainDeformationComponent{
		MaxTracks: 200,
		Tracks:    make([]TrackMark, 100),
	}
	for i := range comp.Tracks {
		comp.Tracks[i] = TrackMark{X: float64(i), Y: float64(i), Depth: 0.1, Width: 0.2, Age: 0.0, FadeTime: 60.0}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetVisibleTracks(comp, 25.0, 25.0, 75.0, 75.0)
	}
}

func BenchmarkGetVisibleTracks_SpatialHash_300(b *testing.B) {
	comp := &TerrainDeformationComponent{
		MaxTracks: 500,
		Tracks:    make([]TrackMark, 300),
	}
	for i := range comp.Tracks {
		comp.Tracks[i] = TrackMark{X: float64(i % 100), Y: float64(i / 100), Depth: 0.1, Width: 0.2, Age: 0.0, FadeTime: 60.0}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetVisibleTracks(comp, 25.0, 0.0, 75.0, 2.0)
	}
}

func BenchmarkCreatePhysicsComponents_Cart(b *testing.B) {
	vehicle := &procvehicle.Vehicle{
		VehicleType:   procvehicle.TypeCart,
		Handling:      100.0,
		MaxDurability: 100.0,
		CargoWeight:   50.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreatePhysicsComponents(vehicle)
	}
}
