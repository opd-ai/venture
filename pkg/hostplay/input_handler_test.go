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
