package vehicle

import (
	"math"
	"testing"
)

func TestNewEnhancedVehicleSystem(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	if sys == nil {
		t.Fatal("NewEnhancedVehicleSystem returned nil")
	}
	if !sys.IsEnabled() {
		t.Error("system should be enabled by default")
	}
}

func TestEnhancedVehicleSystem_EnableDisable(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	if !sys.IsEnabled() {
		t.Error("system should start enabled")
	}

	sys.Disable()
	if sys.IsEnabled() {
		t.Error("system should be disabled after Disable()")
	}

	sys.Enable()
	if !sys.IsEnabled() {
		t.Error("system should be enabled after Enable()")
	}
}

func TestEnhancedVehicleSystem_UpdateVehiclePhysics_Disabled(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	sys.Disable()

	state := VehicleState{
		PositionX: 100.0,
		PositionY: 100.0,
		VelocityX: 50.0,
		VelocityY: 0.0,
		Speed:     50.0,
	}

	newState := sys.UpdateVehiclePhysics(nil, nil, nil, nil, state, 0.016)

	// State should be unchanged when system disabled
	if newState.PositionX != state.PositionX || newState.VelocityX != state.VelocityX {
		t.Error("state should be unchanged when system is disabled")
	}
}

func TestEnhancedVehicleSystem_UpdateVehiclePhysics_WithSuspension(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	suspension := NewSuspensionComponent(4)

	state := VehicleState{
		PositionX:     100.0,
		PositionY:     100.0,
		VelocityX:     50.0,
		VelocityY:     0.0,
		Speed:         50.0,
		IsGrounded:    true,
		TerrainHeight: []float64{0, 0, 0, 0},
	}

	newState := sys.UpdateVehiclePhysics(suspension, nil, nil, nil, state, 0.016)

	// State should be processed
	if newState.PositionX != state.PositionX {
		t.Error("position should not change from physics update alone")
	}

	// Check grounded status updated
	if !newState.IsGrounded {
		t.Error("vehicle should be grounded with all wheels touching")
	}
}

func TestEnhancedVehicleSystem_UpdateVehiclePhysics_WithWeightTransfer(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	suspension := NewSuspensionComponent(4)
	weightTransfer := NewWeightTransferComponent()

	state := VehicleState{
		PositionX:     100.0,
		PositionY:     100.0,
		VelocityX:     50.0,
		VelocityY:     0.0,
		AngularVel:    0.0,
		Speed:         50.0,
		TerrainHeight: []float64{0, 0, 0, 0},
	}

	// Run update twice to calculate acceleration
	sys.UpdateVehiclePhysics(suspension, weightTransfer, nil, nil, state, 0.016)

	state.VelocityX = 100.0 // Accelerate
	state.Speed = 100.0
	newState := sys.UpdateVehiclePhysics(suspension, weightTransfer, nil, nil, state, 0.016)

	// Weight transfer should have been calculated
	weights := weightTransfer.GetWheelWeights()
	totalWeight := weights[0] + weights[1] + weights[2] + weights[3]
	if math.Abs(totalWeight-1.0) > 0.01 {
		t.Errorf("total weight should be 1.0, got %f", totalWeight)
	}

	// During acceleration, rear wheels should have more weight
	rearWeight := weightTransfer.GetRearAxleWeight()
	frontWeight := weightTransfer.GetFrontAxleWeight()
	if rearWeight <= frontWeight {
		t.Errorf("during acceleration, rear weight (%f) should exceed front weight (%f)", rearWeight, frontWeight)
	}

	_ = newState // Use newState
}

func TestEnhancedVehicleSystem_UpdateVehiclePhysics_WithDeformation(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	suspension := NewSuspensionComponent(4)
	deformation := NewTerrainDeformationComponent(12345)

	state := VehicleState{
		PositionX:     100.0,
		PositionY:     100.0,
		VelocityX:     50.0,
		VelocityY:     0.0,
		Rotation:      0.0,
		Speed:         50.0, // Above 10 px/s threshold
		IsGrounded:    true,
		TerrainHeight: []float64{0, 0, 0, 0},
		TerrainTypes:  []TerrainType{TerrainSoft, TerrainSoft, TerrainSoft, TerrainSoft},
	}

	// Update should create tracks
	sys.UpdateVehiclePhysics(suspension, nil, deformation, nil, state, 0.016)

	trackCount := deformation.GetTrackCount()
	if trackCount == 0 {
		t.Error("moving vehicle on soft terrain should create tracks")
	}
}

func TestEnhancedVehicleSystem_UpdateVehiclePhysics_WithDamage(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	collision := NewCollisionResponseComponent(1000.0)

	// Damage vehicle to 50% integrity
	collision.StructuralIntegrity = 0.5

	state := VehicleState{
		PositionX: 100.0,
		PositionY: 100.0,
		VelocityX: 100.0,
		VelocityY: 0.0,
		Speed:     100.0,
	}

	newState := sys.UpdateVehiclePhysics(nil, nil, nil, collision, state, 0.016)

	// Speed should be reduced due to damage
	if newState.Speed >= state.Speed {
		t.Errorf("damaged vehicle speed (%f) should be less than input speed (%f)", newState.Speed, state.Speed)
	}

	// Velocity should also be scaled
	expectedVelX := state.VelocityX * collision.GetDamageMultiplier()
	if math.Abs(newState.VelocityX-expectedVelX) > 0.1 {
		t.Errorf("velocity X: got %f, want ~%f", newState.VelocityX, expectedVelX)
	}
}

func TestEnhancedVehicleSystem_ProcessVehicleCollision(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	collision := NewCollisionResponseComponent(1000.0)

	state := VehicleState{
		VelocityX: 100.0,
		VelocityY: 0.0,
		Speed:     100.0,
	}

	// Process collision with wall (normal pointing left)
	normalX, normalY := -1.0, 0.0
	newVelX, newVelY, damage := sys.ProcessVehicleCollision(collision, state, normalX, normalY)

	// Velocity should be reversed (bounced)
	if newVelX >= 0.0 {
		t.Errorf("collision should reverse velocity, got vx=%f", newVelX)
	}

	// Should deal damage above threshold
	if damage <= 0.0 {
		t.Error("high speed collision should deal damage")
	}

	// Structural integrity should be reduced
	if collision.GetIntegrity() >= 1.0 {
		t.Error("collision should reduce structural integrity")
	}

	_ = newVelY // Use newVelY
}

func TestEnhancedVehicleSystem_ProcessVehicleCollision_NilComponent(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	state := VehicleState{
		VelocityX: 100.0,
		VelocityY: 0.0,
	}

	// Should not crash with nil collision component
	newVelX, newVelY, damage := sys.ProcessVehicleCollision(nil, state, -1.0, 0.0)

	// Should return original velocity and zero damage
	if newVelX != state.VelocityX || newVelY != state.VelocityY {
		t.Error("with nil collision component, velocity should be unchanged")
	}
	if damage != 0.0 {
		t.Error("with nil collision component, damage should be zero")
	}
}

func TestEnhancedVehicleSystem_IntegratedPhysics(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	suspension := NewSuspensionComponent(4)
	weightTransfer := NewWeightTransferComponent()
	deformation := NewTerrainDeformationComponent(12345)
	collision := NewCollisionResponseComponent(1000.0)

	state := VehicleState{
		PositionX:     100.0,
		PositionY:     100.0,
		VelocityX:     0.0,
		VelocityY:     0.0,
		Rotation:      0.0,
		AngularVel:    0.0,
		Speed:         0.0,
		IsGrounded:    true,
		TerrainHeight: []float64{0, 0, 0, 0},
		TerrainTypes:  []TerrainType{TerrainSoft, TerrainSoft, TerrainSoft, TerrainSoft},
	}

	// Simulate vehicle accelerating
	for i := 0; i < 10; i++ {
		// Increase speed
		state.VelocityX += 10.0
		state.Speed = state.VelocityX
		state.PositionX += state.VelocityX * 0.016

		// Update physics
		state = sys.UpdateVehiclePhysics(suspension, weightTransfer, deformation, collision, state, 0.016)
	}

	// Check results
	if deformation.GetTrackCount() == 0 {
		t.Error("accelerating vehicle should leave tracks")
	}

	rearWeight := weightTransfer.GetRearAxleWeight()
	frontWeight := weightTransfer.GetFrontAxleWeight()
	if rearWeight <= frontWeight {
		t.Error("accelerating vehicle should have more weight on rear")
	}

	if !state.IsGrounded {
		t.Error("vehicle should remain grounded")
	}
}

// Benchmark tests
func BenchmarkEnhancedVehicleSystem_UpdateVehiclePhysics(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	suspension := NewSuspensionComponent(4)
	weightTransfer := NewWeightTransferComponent()
	deformation := NewTerrainDeformationComponent(12345)
	collision := NewCollisionResponseComponent(1000.0)

	state := VehicleState{
		PositionX:     100.0,
		PositionY:     100.0,
		VelocityX:     50.0,
		VelocityY:     0.0,
		Rotation:      0.0,
		AngularVel:    0.0,
		Speed:         50.0,
		TerrainHeight: []float64{0, 0, 0, 0},
		TerrainTypes:  []TerrainType{TerrainSoft, TerrainSoft, TerrainSoft, TerrainSoft},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state = sys.UpdateVehiclePhysics(suspension, weightTransfer, deformation, collision, state, 0.016)
	}
}

func BenchmarkEnhancedVehicleSystem_ProcessVehicleCollision(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	collision := NewCollisionResponseComponent(1000.0)

	state := VehicleState{
		VelocityX: 100.0,
		VelocityY: 0.0,
		Speed:     100.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collision.Reset() // Reset to avoid accumulating damage
		_, _, _ = sys.ProcessVehicleCollision(collision, state, -1.0, 0.0)
	}
}
