package vehicle

import "testing"

// Test helper functions for ECS purity compliance

func TestHelpers_Suspension(t *testing.T) {
	comp := NewSuspensionComponent(4)
	comp.Wheels[0].Load = 100.0
	comp.Wheels[1].Compression = 0.5
	comp.Wheels[2].IsGrounded = false // Explicitly unground other wheels for testing
	comp.Wheels[3].IsGrounded = false

	// Test GetWheelLoad
	if load := GetWheelLoad(comp, 0); load != 100.0 {
		t.Errorf("GetWheelLoad(0) = %f, want 100.0", load)
	}
	if load := GetWheelLoad(comp, -1); load != 0.0 {
		t.Errorf("GetWheelLoad(-1) = %f, want 0.0", load)
	}

	// Test GetWheelCompression
	if comp := GetWheelCompression(comp, 1); comp != 0.5 {
		t.Errorf("GetWheelCompression(1) = %f, want 0.5", comp)
	}

	// Test IsWheelGrounded
	if IsWheelGrounded(comp, 2) {
		t.Error("IsWheelGrounded(2) = true, want false")
	}

	// Test GetGroundedWheelCount (wheels 0 and 1 are grounded)
	if count := GetGroundedWheelCount(comp); count != 2 {
		t.Errorf("GetGroundedWheelCount() = %d, want 2", count)
	}

	// Test SetWheelLoad
	SetWheelLoad(comp, 3, 200.0)
	if comp.Wheels[3].Load != 200.0 {
		t.Errorf("SetWheelLoad(3, 200.0) failed, got %f", comp.Wheels[3].Load)
	}
}

func TestHelpers_WeightTransfer(t *testing.T) {
	comp := &WeightTransferComponent{
		FrontLeftWeight:       0.3,
		FrontRightWeight:      0.2,
		RearLeftWeight:        0.25,
		RearRightWeight:       0.25,
		LastTransferMagnitude: 0.5,
	}

	// Test GetWheelWeights
	weights := GetWheelWeights(comp)
	expected := [4]float64{0.3, 0.2, 0.25, 0.25}
	if weights != expected {
		t.Errorf("GetWheelWeights() = %v, want %v", weights, expected)
	}

	// Test GetFrontAxleWeight
	if front := GetFrontAxleWeight(comp); front != 0.5 {
		t.Errorf("GetFrontAxleWeight() = %f, want 0.5", front)
	}

	// Test GetRearAxleWeight
	if rear := GetRearAxleWeight(comp); rear != 0.5 {
		t.Errorf("GetRearAxleWeight() = %f, want 0.5", rear)
	}

	// Test GetLeftSideWeight
	if left := GetLeftSideWeight(comp); left != 0.55 {
		t.Errorf("GetLeftSideWeight() = %f, want 0.55", left)
	}

	// Test GetRightSideWeight
	if right := GetRightSideWeight(comp); right != 0.45 {
		t.Errorf("GetRightSideWeight() = %f, want 0.45", right)
	}

	// Test GetTransferMagnitude
	if mag := GetTransferMagnitude(comp); mag != 0.5 {
		t.Errorf("GetTransferMagnitude() = %f, want 0.5", mag)
	}

	// Test ResetWeightTransfer
	comp.AccelerationX = 5.0
	ResetWeightTransfer(comp)
	if comp.AccelerationX != 0.0 || comp.FrontLeftWeight != 0.25 {
		t.Error("ResetWeightTransfer() failed")
	}
}

func TestHelpers_CollisionResponse(t *testing.T) {
	comp := &CollisionResponseComponent{
		StructuralIntegrity: 0.5,
		CollisionCount:      3,
		LastImpactForce:     100.0,
		LastImpactVelocity:  50.0,
		DamageThreshold:     30.0,
	}

	// Test GetDamageMultiplier
	mult := GetDamageMultiplier(comp)
	expected := 0.5 + (0.5 * 0.5) // 0.75
	if mult != expected {
		t.Errorf("GetDamageMultiplier() = %f, want %f", mult, expected)
	}

	// Test IsDestroyed
	if IsDestroyed(comp) {
		t.Error("IsDestroyed() = true, want false")
	}
	comp.StructuralIntegrity = 0.0
	if !IsDestroyed(comp) {
		t.Error("IsDestroyed() = false, want true")
	}
	comp.StructuralIntegrity = 0.5 // Reset

	// Test RepairVehicle
	RepairVehicle(comp, 0.3)
	if comp.StructuralIntegrity != 0.8 {
		t.Errorf("RepairVehicle(0.3) resulted in %f, want 0.8", comp.StructuralIntegrity)
	}

	// Test capping at 1.0
	RepairVehicle(comp, 0.5)
	if comp.StructuralIntegrity != 1.0 {
		t.Errorf("RepairVehicle(0.5) should cap at 1.0, got %f", comp.StructuralIntegrity)
	}

	// Test ResetCollisionResponse
	ResetCollisionResponse(comp)
	if comp.CollisionCount != 0 || comp.StructuralIntegrity != 1.0 {
		t.Error("ResetCollisionResponse() failed")
	}

	// Test ShouldCauseDamage
	if !ShouldCauseDamage(comp, 40.0) {
		t.Error("ShouldCauseDamage(40.0) = false, want true")
	}
	if ShouldCauseDamage(comp, 20.0) {
		t.Error("ShouldCauseDamage(20.0) = true, want false")
	}
}

func TestHelpers_TerrainDeformation(t *testing.T) {
	comp := NewTerrainDeformationComponent(12345)
	comp.Tracks = []TrackMark{
		{X: 100, Y: 100, Age: 10, FadeTime: 30},
		{X: 200, Y: 200, Age: 20, FadeTime: 30},
	}

	// Test GetVisibleTracks
	visible := GetVisibleTracks(comp, 50, 50, 150, 150)
	if len(visible) != 1 {
		t.Errorf("GetVisibleTracks() returned %d tracks, want 1", len(visible))
	}

	// Test GetTrackAlpha
	track := comp.Tracks[0]
	alpha := GetTrackAlpha(&track)
	expected := 1.0 - (10.0 / 30.0)
	if alpha < expected-0.01 || alpha > expected+0.01 {
		t.Errorf("GetTrackAlpha() = %f, want ~%f", alpha, expected)
	}

	// Test ClearTracks
	ClearTracks(comp)
	if len(comp.Tracks) != 0 {
		t.Errorf("ClearTracks() left %d tracks, want 0", len(comp.Tracks))
	}
}

func BenchmarkHelpers_GetWheelLoad(b *testing.B) {
	comp := NewSuspensionComponent(4)
	comp.Wheels[0].Load = 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetWheelLoad(comp, i%4)
	}
}

func BenchmarkHelpers_GetDamageMultiplier(b *testing.B) {
	comp := &CollisionResponseComponent{StructuralIntegrity: 0.75}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetDamageMultiplier(comp)
	}
}

func BenchmarkHelpers_GetVisibleTracks(b *testing.B) {
	comp := NewTerrainDeformationComponent(12345)
	sys := NewEnhancedVehicleSystem()
	for i := 0; i < 100; i++ {
		sys.AddTerrainTrack(comp, float64(i*10), float64(i*10), 0, 100, TerrainSoft)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetVisibleTracks(comp, 500.0, 500.0, 700.0, 700.0)
	}
}
