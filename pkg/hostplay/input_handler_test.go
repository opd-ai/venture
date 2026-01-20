package hostplay

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

func TestNewInputHandler(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())

	handler := NewInputHandler(world, logger)

	if handler == nil {
		t.Fatal("NewInputHandler returned nil")
	}
	if handler.world != world {
		t.Error("world not set correctly")
	}
	if handler.playerMap == nil {
		t.Error("playerMap not initialized")
	}
	if handler.moveSpeed <= 0 {
		t.Error("moveSpeed not set")
	}
}

func TestInputHandler_RegisterUnregisterPlayer(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	playerID := uint64(123)

	// Test registration
	handler.RegisterPlayer(playerID, entity)

	if handler.PlayerCount() != 1 {
		t.Errorf("PlayerCount = %d, want 1", handler.PlayerCount())
	}

	retrievedEntity, exists := handler.GetPlayerEntity(playerID)
	if !exists {
		t.Error("player not found after registration")
	}
	if retrievedEntity != entity {
		t.Error("retrieved entity doesn't match registered entity")
	}

	// Test unregistration
	handler.UnregisterPlayer(playerID)

	if handler.PlayerCount() != 0 {
		t.Errorf("PlayerCount = %d, want 0 after unregister", handler.PlayerCount())
	}

	_, exists = handler.GetPlayerEntity(playerID)
	if exists {
		t.Error("player still exists after unregistration")
	}
}

func TestInputHandler_ProcessMovement(t *testing.T) {
	tests := []struct {
		name     string
		dx       float64
		dy       float64
		wantZero bool
	}{
		{"right", 1.0, 0.0, false},
		{"left", -1.0, 0.0, false},
		{"up", 0.0, -1.0, false},
		{"down", 0.0, 1.0, false},
		{"diagonal", 1.0, 1.0, false},
		{"stop", 0.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			logger := logrus.NewEntry(logrus.New())
			handler := NewInputHandler(world, logger)

			entity := world.CreateEntity()
			velocity := &engine.VelocityComponent{VX: 0, VY: 0}
			entity.AddComponent(velocity)

			playerID := uint64(1)
			handler.RegisterPlayer(playerID, entity)

			data := map[string]interface{}{
				"dx": tt.dx,
				"dy": tt.dy,
			}

			handler.ProcessInput(playerID, "move", data)

			if tt.wantZero {
				if velocity.VX != 0 || velocity.VY != 0 {
					t.Errorf("velocity = (%f, %f), want (0, 0)", velocity.VX, velocity.VY)
				}
			} else {
				if velocity.VX == 0 && velocity.VY == 0 {
					t.Error("velocity is zero when it should be non-zero")
				}
			}
		})
	}
}

func TestInputHandler_ProcessAttack(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	aim := &engine.AimComponent{AimAngle: 0}
	entity.AddComponent(aim)

	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	targetAngle := 1.5708 // 90 degrees in radians
	data := map[string]interface{}{
		"angle": targetAngle,
	}

	handler.ProcessInput(playerID, "attack", data)

	if aim.AimAngle != targetAngle {
		t.Errorf("AimAngle = %f, want %f", aim.AimAngle, targetAngle)
	}
}

func TestInputHandler_ProcessUnknownPlayer(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	// Process input for non-existent player (should not panic)
	data := map[string]interface{}{"dx": 1.0, "dy": 0.0}
	handler.ProcessInput(999, "move", data)

	// Test passes if no panic occurs
}

func TestInputHandler_ProcessInvalidData(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	velocity := &engine.VelocityComponent{VX: 100, VY: 100}
	entity.AddComponent(velocity)

	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Invalid data (missing dx/dy)
	data := map[string]interface{}{}
	handler.ProcessInput(playerID, "move", data)

	// Velocity should remain unchanged
	if velocity.VX != 100 || velocity.VY != 100 {
		t.Error("velocity changed with invalid data")
	}
}

func TestInputHandler_MultipleUsers(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	// Register 3 players
	for i := uint64(1); i <= 3; i++ {
		entity := world.CreateEntity()
		entity.AddComponent(&engine.VelocityComponent{})
		handler.RegisterPlayer(i, entity)
	}

	if handler.PlayerCount() != 3 {
		t.Errorf("PlayerCount = %d, want 3", handler.PlayerCount())
	}

	// Unregister middle player
	handler.UnregisterPlayer(2)

	if handler.PlayerCount() != 2 {
		t.Errorf("PlayerCount = %d, want 2 after unregister", handler.PlayerCount())
	}

	// Verify player 1 and 3 still exist
	if _, exists := handler.GetPlayerEntity(1); !exists {
		t.Error("player 1 not found")
	}
	if _, exists := handler.GetPlayerEntity(3); !exists {
		t.Error("player 3 not found")
	}
	if _, exists := handler.GetPlayerEntity(2); exists {
		t.Error("player 2 still exists")
	}
}

func TestInputHandler_ProcessInputRaw(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{VX: 0, VY: 0})
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Test valid JSON input
	rawData := []byte(`{"dx": 1.0, "dy": 0.0}`)
	handler.ProcessInputRaw(playerID, "move", rawData)

	// Verify velocity was updated
	comp, ok := entity.GetComponent("velocity")
	if !ok {
		t.Fatal("velocity component not found")
	}
	velocity := comp.(*engine.VelocityComponent)
	if velocity.VX == 0 && velocity.VY == 0 {
		t.Error("velocity not updated by ProcessInputRaw")
	}

	// Test invalid JSON
	invalidData := []byte(`{invalid json}`)
	handler.ProcessInputRaw(playerID, "move", invalidData)
	// Should not crash, just log warning

	// Test unknown player
	handler.ProcessInputRaw(uint64(999), "move", rawData)
	// Should not crash, just log warning
}

// TestInputHandler_ProcessItemUse tests the processItemUse function
func TestInputHandler_ProcessItemUse(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Test with valid slot
	data := map[string]interface{}{
		"slot": float64(0),
	}
	handler.ProcessInput(playerID, "use_item", data)
	// Test passes if no panic occurs

	// Test with different slot
	data["slot"] = float64(5)
	handler.ProcessInput(playerID, "use_item", data)
	// Test passes if no panic occurs

	// Test with missing slot (should be skipped gracefully)
	emptyData := map[string]interface{}{}
	handler.ProcessInput(playerID, "use_item", emptyData)
	// Test passes if no panic occurs

	// Test with invalid slot type (should be skipped gracefully)
	invalidData := map[string]interface{}{
		"slot": "not a number",
	}
	handler.ProcessInput(playerID, "use_item", invalidData)
	// Test passes if no panic occurs
}

// TestInputHandler_ProcessInteraction tests the processInteraction function
func TestInputHandler_ProcessInteraction(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Test with empty data
	data := map[string]interface{}{}
	handler.ProcessInput(playerID, "interact", data)
	// Test passes if no panic occurs

	// Test with some data
	data["target_id"] = float64(123)
	handler.ProcessInput(playerID, "interact", data)
	// Test passes if no panic occurs
}

// TestInputHandler_UnknownInputType tests handling of unknown input types
func TestInputHandler_UnknownInputType(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Test with various unknown types
	unknownTypes := []string{"jump", "crouch", "roll", "special", "", "invalid_action"}
	for _, inputType := range unknownTypes {
		t.Run(inputType, func(t *testing.T) {
			data := map[string]interface{}{}
			handler.ProcessInput(playerID, inputType, data)
			// Test passes if no panic occurs
		})
	}
}

// TestInputHandler_MovementEntityWithoutVelocity tests movement on entity without velocity
func TestInputHandler_MovementEntityWithoutVelocity(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	// Intentionally not adding velocity component
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	data := map[string]interface{}{
		"dx": 1.0,
		"dy": 0.0,
	}
	handler.ProcessInput(playerID, "move", data)
	// Test passes if no panic occurs
}

// TestInputHandler_AttackEntityWithoutAim tests attack on entity without aim component
func TestInputHandler_AttackEntityWithoutAim(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	// Intentionally not adding aim component
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	data := map[string]interface{}{
		"angle": 1.57,
	}
	handler.ProcessInput(playerID, "attack", data)
	// Test passes if no panic occurs
}

// TestInputHandler_AttackWithMissingAngle tests attack without angle in data
func TestInputHandler_AttackWithMissingAngle(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	aim := &engine.AimComponent{AimAngle: 0.5}
	entity.AddComponent(aim)
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Attack without angle should not change aim
	data := map[string]interface{}{}
	handler.ProcessInput(playerID, "attack", data)

	if aim.AimAngle != 0.5 {
		t.Errorf("AimAngle changed from %f when no angle provided", 0.5)
	}
}

// TestInputHandler_MovementNormalization tests that movement is properly normalized
func TestInputHandler_MovementNormalization(t *testing.T) {
	world := engine.NewWorld()
	logger := logrus.NewEntry(logrus.New())
	handler := NewInputHandler(world, logger)

	entity := world.CreateEntity()
	velocity := &engine.VelocityComponent{VX: 0, VY: 0}
	entity.AddComponent(velocity)
	playerID := uint64(1)
	handler.RegisterPlayer(playerID, entity)

	// Test diagonal movement (should be normalized)
	data := map[string]interface{}{
		"dx": 1.0,
		"dy": 1.0,
	}
	handler.ProcessInput(playerID, "move", data)

	// Velocity magnitude should equal moveSpeed (200.0)
	magnitude := velocity.VX*velocity.VX + velocity.VY*velocity.VY
	expectedMagnitudeSq := handler.moveSpeed * handler.moveSpeed
	tolerance := 0.01

	if magnitude < expectedMagnitudeSq-tolerance || magnitude > expectedMagnitudeSq+tolerance {
		t.Errorf("magnitude squared = %f, want %f (tolerance %f)",
			magnitude, expectedMagnitudeSq, tolerance)
	}
}

// TestInputHandler_MovementDirections tests all cardinal and diagonal directions
func TestInputHandler_MovementDirections(t *testing.T) {
	directions := []struct {
		name string
		dx   float64
		dy   float64
	}{
		{"right", 1.0, 0.0},
		{"left", -1.0, 0.0},
		{"up", 0.0, -1.0},
		{"down", 0.0, 1.0},
		{"up-right", 1.0, -1.0},
		{"up-left", -1.0, -1.0},
		{"down-right", 1.0, 1.0},
		{"down-left", -1.0, 1.0},
	}

	for _, dir := range directions {
		t.Run(dir.name, func(t *testing.T) {
			world := engine.NewWorld()
			logger := logrus.NewEntry(logrus.New())
			handler := NewInputHandler(world, logger)

			entity := world.CreateEntity()
			velocity := &engine.VelocityComponent{VX: 0, VY: 0}
			entity.AddComponent(velocity)
			handler.RegisterPlayer(1, entity)

			data := map[string]interface{}{"dx": dir.dx, "dy": dir.dy}
			handler.ProcessInput(1, "move", data)

			// Verify direction is correct (sign matches input)
			if dir.dx > 0 && velocity.VX <= 0 {
				t.Error("expected positive VX")
			}
			if dir.dx < 0 && velocity.VX >= 0 {
				t.Error("expected negative VX")
			}
			if dir.dy > 0 && velocity.VY <= 0 {
				t.Error("expected positive VY")
			}
			if dir.dy < 0 && velocity.VY >= 0 {
				t.Error("expected negative VY")
			}
		})
	}
}
