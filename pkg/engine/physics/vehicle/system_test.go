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
	frontWeight := GetFrontAxleWeight(weightTransfer)
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

	trackCount := len(deformation.Tracks)
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
	expectedVelX := state.VelocityX * GetDamageMultiplier(collision)
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
	if len(deformation.Tracks) == 0 {
		t.Error("accelerating vehicle should leave tracks")
	}

	rearWeight := weightTransfer.GetRearAxleWeight()
	frontWeight := GetFrontAxleWeight(weightTransfer)
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

// --- ECS System Interface Tests ---

// MockEntity implements the Entity interface for testing.
type MockEntity struct {
	components map[string]interface{}
}

func NewMockEntity() *MockEntity {
	return &MockEntity{
		components: make(map[string]interface{}),
	}
}

func (e *MockEntity) HasComponent(componentType string) bool {
	_, exists := e.components[componentType]
	return exists
}

func (e *MockEntity) GetComponent(componentType string) interface{} {
	return e.components[componentType]
}

func (e *MockEntity) AddComponent(componentType string, component interface{}) {
	e.components[componentType] = component
}

// MockPositionComponent for testing.
type MockPositionComponent struct {
	X, Y float64
}

func (p *MockPositionComponent) GetX() float64 { return p.X }
func (p *MockPositionComponent) GetY() float64 { return p.Y }

// MockVelocityComponent for testing.
type MockVelocityComponent struct {
	VelX, VelY float64
}

func (v *MockVelocityComponent) GetVelX() float64 { return v.VelX }
func (v *MockVelocityComponent) GetVelY() float64 { return v.VelY }

func (v *MockVelocityComponent) SetVelocity(vx, vy float64) {
	v.VelX = vx
	v.VelY = vy
}

// MockRotationComponent for testing.
type MockRotationComponent struct {
	Angle float64
}

func (r *MockRotationComponent) GetAngle() float64 { return r.Angle }

func TestEnhancedVehicleSystem_Update_NoVehicles(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	// Empty entity list
	entities := []Entity{}
	sys.Update(entities, 0.016)

	// No crash, no errors - test passes
}

func TestEnhancedVehicleSystem_Update_NonVehicleEntity(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	// Entity without vehicle physics components
	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})

	entities := []Entity{entity}
	sys.Update(entities, 0.016)

	// Should skip entity without vehicle components
}

func TestEnhancedVehicleSystem_Update_WithSuspension(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})
	entity.AddComponent("rotation", &MockRotationComponent{Angle: 0})

	suspension := NewSuspensionComponent(4) // 4 wheels
	entity.AddComponent("suspension", suspension)

	entities := []Entity{entity}
	sys.Update(entities, 0.016)

	// Verify suspension was updated
	if suspension.LastUpdateTime == 0 {
		t.Error("suspension LastUpdateTime should be updated")
	}
}

func TestEnhancedVehicleSystem_Update_WithWeightTransfer(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})

	weightTransfer := NewWeightTransferComponent()
	entity.AddComponent("weight_transfer", weightTransfer)

	entities := []Entity{entity}

	// First update to establish initial velocity
	sys.Update(entities, 0.016)

	// Change velocity to trigger weight transfer
	vel := entity.GetComponent("velocity").(*MockVelocityComponent)
	vel.VelX = 100 // Acceleration

	// Second update should calculate weight transfer
	sys.Update(entities, 0.016)

	if weightTransfer.AccelerationX == 0 {
		t.Error("weight transfer should calculate acceleration")
	}
}

func TestEnhancedVehicleSystem_Update_WithDeformation(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})
	entity.AddComponent("rotation", &MockRotationComponent{Angle: 0})

	suspension := NewSuspensionComponent(4) // 4 wheels
	entity.AddComponent("suspension", suspension)

	deformation := NewTerrainDeformationComponent(12345)
	entity.AddComponent("terrain_deformation", deformation)

	entities := []Entity{entity}
	sys.Update(entities, 0.016)

	// Deformation system should process (tracks added if moving fast enough)
	// At 50 pixels/s, may not create tracks (threshold is 10.0 in updateTireTracks)
}

func TestEnhancedVehicleSystem_Update_WithCollision(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})

	collision := NewCollisionResponseComponent(1000.0)
	collision.StructuralIntegrity = 0.5 // Damaged vehicle
	entity.AddComponent("collision_response", collision)

	entities := []Entity{entity}

	originalVelX := entity.GetComponent("velocity").(*MockVelocityComponent).VelX

	sys.Update(entities, 0.016)

	// Velocity should be reduced by damage multiplier
	newVelX := entity.GetComponent("velocity").(*MockVelocityComponent).VelX
	if newVelX >= originalVelX {
		t.Error("damaged vehicle should have reduced velocity")
	}
}

func TestEnhancedVehicleSystem_Update_Disabled(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	sys.Disable()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})

	suspension := NewSuspensionComponent(4) // 4 wheels
	entity.AddComponent("suspension", suspension)

	entities := []Entity{entity}

	initialTime := suspension.LastUpdateTime
	sys.Update(entities, 0.016)

	// When disabled, no updates should occur
	if suspension.LastUpdateTime != initialTime {
		t.Error("disabled system should not update components")
	}
}

func TestEnhancedVehicleSystem_Update_MultipleEntities(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	// Create 3 vehicle entities
	entities := make([]Entity, 3)
	for i := 0; i < 3; i++ {
		entity := NewMockEntity()
		entity.AddComponent("position", &MockPositionComponent{X: float64(i * 100), Y: 100})
		entity.AddComponent("velocity", &MockVelocityComponent{VelX: float64(i * 10), VelY: 0})
		suspension := NewSuspensionComponent(4) // 4 wheels
		entity.AddComponent("suspension", suspension)
		entities[i] = entity
	}

	sys.Update(entities, 0.016)

	// Verify all entities were updated
	for i, e := range entities {
		entity := e.(*MockEntity)
		suspension := entity.GetComponent("suspension").(*SuspensionComponent)
		if suspension.LastUpdateTime == 0 {
			t.Errorf("entity %d suspension was not updated", i)
		}
	}
}

func TestEnhancedVehicleSystem_Update_AllComponents(t *testing.T) {
	sys := NewEnhancedVehicleSystem()

	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})
	entity.AddComponent("rotation", &MockRotationComponent{Angle: 0})

	suspension := NewSuspensionComponent(4) // 4 wheels
	weightTransfer := NewWeightTransferComponent()
	deformation := NewTerrainDeformationComponent(12345)
	collision := NewCollisionResponseComponent(1000.0)

	entity.AddComponent("suspension", suspension)
	entity.AddComponent("weight_transfer", weightTransfer)
	entity.AddComponent("terrain_deformation", deformation)
	entity.AddComponent("collision_response", collision)

	entities := []Entity{entity}

	sys.Update(entities, 0.016)

	// All components should be processed
	if suspension.LastUpdateTime == 0 {
		t.Error("suspension was not updated")
	}
}

func BenchmarkEnhancedVehicleSystem_Update(b *testing.B) {
	sys := NewEnhancedVehicleSystem()

	// Create entity with all physics components
	entity := NewMockEntity()
	entity.AddComponent("position", &MockPositionComponent{X: 100, Y: 100})
	entity.AddComponent("velocity", &MockVelocityComponent{VelX: 50, VelY: 0})
	entity.AddComponent("rotation", &MockRotationComponent{Angle: 0})

	suspension := NewSuspensionComponent(4) // 4 wheels
	weightTransfer := NewWeightTransferComponent()
	deformation := NewTerrainDeformationComponent(12345)
	collision := NewCollisionResponseComponent(1000.0)

	entity.AddComponent("suspension", suspension)
	entity.AddComponent("weight_transfer", weightTransfer)
	entity.AddComponent("terrain_deformation", deformation)
	entity.AddComponent("collision_response", collision)

	entities := []Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
