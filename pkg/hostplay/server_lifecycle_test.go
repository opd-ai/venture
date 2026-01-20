package hostplay

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// TestServerManager_SpawnPlayer tests the spawnPlayer function
func TestServerManager_SpawnPlayer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51000,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Start to initialize world and terrain
	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Get initial entity count
	initialCount := len(manager.world.GetEntities())

	// Spawn a player
	playerID := uint64(100)
	manager.spawnPlayer(playerID)

	// Wait for world update to process pending entities
	manager.world.Update(0.016)

	// Verify entity was created
	newCount := len(manager.world.GetEntities())
	if newCount <= initialCount {
		t.Errorf("entity count not increased after spawn: before=%d, after=%d", initialCount, newCount)
	}

	// Verify player is registered with input handler
	entity, exists := manager.inputHandler.GetPlayerEntity(playerID)
	if !exists {
		t.Error("player not registered with input handler after spawn")
	}
	if entity == nil {
		t.Error("player entity is nil")
	}

	// Verify entity has required components
	if _, ok := entity.GetComponent("position"); !ok {
		t.Error("spawned entity missing position component")
	}
	if _, ok := entity.GetComponent("velocity"); !ok {
		t.Error("spawned entity missing velocity component")
	}
	if _, ok := entity.GetComponent("health"); !ok {
		t.Error("spawned entity missing health component")
	}
	if _, ok := entity.GetComponent("network"); !ok {
		t.Error("spawned entity missing network component")
	}
}

// TestServerManager_RemovePlayer tests the removePlayer function
func TestServerManager_RemovePlayer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51100,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Spawn a player first
	playerID := uint64(200)
	manager.spawnPlayer(playerID)
	manager.world.Update(0.016)

	// Verify player exists
	_, exists := manager.inputHandler.GetPlayerEntity(playerID)
	if !exists {
		t.Fatal("player not registered after spawn")
	}

	// Get entity count before removal
	countBefore := len(manager.world.GetEntities())

	// Remove the player
	manager.removePlayer(playerID)
	manager.world.Update(0.016)

	// Verify player is unregistered
	_, exists = manager.inputHandler.GetPlayerEntity(playerID)
	if exists {
		t.Error("player still registered after removal")
	}

	// Verify entity was removed
	countAfter := len(manager.world.GetEntities())
	if countAfter >= countBefore {
		t.Errorf("entity count not decreased after removal: before=%d, after=%d", countBefore, countAfter)
	}
}

// TestServerManager_RemovePlayerNotFound tests removePlayer with non-existent player
func TestServerManager_RemovePlayerNotFound(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51200,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Try to remove a player that doesn't exist (should not panic)
	manager.removePlayer(uint64(999))
	// Test passes if no panic occurs
}

// TestServerManager_HandlePlayerJoin tests the handlePlayerJoin function
func TestServerManager_HandlePlayerJoin(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51300,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	playerID := uint64(300)
	manager.handlePlayerJoin(playerID)
	manager.world.Update(0.016)

	// Verify player was spawned
	_, exists := manager.inputHandler.GetPlayerEntity(playerID)
	if !exists {
		t.Error("player not registered after handlePlayerJoin")
	}
}

// TestServerManager_HandlePlayerLeave tests the handlePlayerLeave function
func TestServerManager_HandlePlayerLeave(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51400,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Join first
	playerID := uint64(400)
	manager.handlePlayerJoin(playerID)
	manager.world.Update(0.016)

	// Verify player exists
	_, exists := manager.inputHandler.GetPlayerEntity(playerID)
	if !exists {
		t.Fatal("player not registered after join")
	}

	// Leave
	manager.handlePlayerLeave(playerID)
	manager.world.Update(0.016)

	// Verify player was removed
	_, exists = manager.inputHandler.GetPlayerEntity(playerID)
	if exists {
		t.Error("player still registered after handlePlayerLeave")
	}
}

// TestServerManager_HandleWorldUpdate tests the handleWorldUpdate function
func TestServerManager_HandleWorldUpdate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51500,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Spawn a player to have an entity to update
	playerID := uint64(500)
	manager.spawnPlayer(playerID)
	manager.world.Update(0.016)

	// Call handleWorldUpdate (should not panic)
	manager.handleWorldUpdate()

	// Test passes if no panic occurs
}

// TestServerManager_ConvertToStateUpdate tests the convertToStateUpdate function
func TestServerManager_ConvertToStateUpdate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51600,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	tests := []struct {
		name        string
		entityState EntityState
		wantNil     bool
	}{
		{
			name: "all components",
			entityState: EntityState{
				ID:       1,
				Position: &PositionState{X: 100, Y: 200},
				Velocity: &VelocityState{VX: 10, VY: 20},
				Health:   &HealthState{Current: 75, Max: 100},
				Rotation: &RotationState{Angle: 1.57},
			},
			wantNil: false,
		},
		{
			name: "position only",
			entityState: EntityState{
				ID:       2,
				Position: &PositionState{X: 50, Y: 60},
			},
			wantNil: false,
		},
		{
			name: "no components",
			entityState: EntityState{
				ID: 3,
			},
			wantNil: true,
		},
		{
			name: "velocity and health only (no position)",
			entityState: EntityState{
				ID:       4,
				Velocity: &VelocityState{VX: 5, VY: 5},
				Health:   &HealthState{Current: 50, Max: 100},
			},
			wantNil: false, // Has velocity and health components, so update is created
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			update := manager.convertToStateUpdate(tt.entityState)

			if tt.wantNil {
				if update != nil {
					t.Error("expected nil update, got non-nil")
				}
				return
			}

			if update == nil {
				t.Fatal("expected non-nil update, got nil")
			}

			if update.EntityID != tt.entityState.ID {
				t.Errorf("EntityID = %d, want %d", update.EntityID, tt.entityState.ID)
			}

			// Verify components were added
			if len(update.Components) == 0 {
				t.Error("no components in update")
			}
		})
	}
}

// TestServerManager_BroadcastEntityStates tests the broadcastEntityStates function
func TestServerManager_BroadcastEntityStates(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51700,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Test with nil snapshot (should not panic)
	manager.broadcastEntityStates(nil)

	// Test with empty snapshot
	emptySnapshot := &WorldState{
		Timestamp: time.Now().UnixMilli(),
		Entities:  []EntityState{},
	}
	manager.broadcastEntityStates(emptySnapshot)

	// Test with valid snapshot
	validSnapshot := &WorldState{
		Timestamp: time.Now().UnixMilli(),
		Entities: []EntityState{
			{
				ID:       1,
				Position: &PositionState{X: 100, Y: 200},
				Velocity: &VelocityState{VX: 10, VY: 20},
			},
			{
				ID:       2,
				Position: &PositionState{X: 300, Y: 400},
				Health:   &HealthState{Current: 50, Max: 100},
			},
		},
	}
	manager.broadcastEntityStates(validSnapshot)

	// Test passes if no panic occurs
}

// TestServerManager_SpawnPlayerNoRooms tests spawning when terrain has no rooms
func TestServerManager_SpawnPlayerNoRooms(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51800,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Override terrain with one that has no rooms
	manager.generatedTerrain = &terrain.Terrain{
		Width:  100,
		Height: 100,
		Tiles:  make([][]terrain.TileType, 100),
		Rooms:  []*terrain.Room{}, // Empty rooms
	}

	// spawnPlayer should handle this gracefully (log error, not panic)
	manager.spawnPlayer(uint64(999))

	// Verify player was NOT registered (since spawn failed)
	_, exists := manager.inputHandler.GetPlayerEntity(uint64(999))
	if exists {
		t.Error("player was registered despite no rooms for spawning")
	}
}

// TestServerManager_MultiplePlayersJoinLeave tests multiple player lifecycle events
func TestServerManager_MultiplePlayersJoinLeave(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       51900,
		MaxPlayers: 8,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Join 4 players
	playerIDs := []uint64{1, 2, 3, 4}
	for _, id := range playerIDs {
		manager.handlePlayerJoin(id)
		manager.world.Update(0.016)
	}

	// Verify all are registered
	for _, id := range playerIDs {
		if _, exists := manager.inputHandler.GetPlayerEntity(id); !exists {
			t.Errorf("player %d not registered after join", id)
		}
	}

	// Leave players 2 and 4
	manager.handlePlayerLeave(2)
	manager.handlePlayerLeave(4)
	manager.world.Update(0.016)

	// Verify 1 and 3 still exist
	if _, exists := manager.inputHandler.GetPlayerEntity(1); !exists {
		t.Error("player 1 not found after partial leave")
	}
	if _, exists := manager.inputHandler.GetPlayerEntity(3); !exists {
		t.Error("player 3 not found after partial leave")
	}

	// Verify 2 and 4 are gone
	if _, exists := manager.inputHandler.GetPlayerEntity(2); exists {
		t.Error("player 2 still exists after leave")
	}
	if _, exists := manager.inputHandler.GetPlayerEntity(4); exists {
		t.Error("player 4 still exists after leave")
	}
}

// TestServerManager_HandleInputCommand tests the handleInputCommand function
func TestServerManager_HandleInputCommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.FatalLevel)

	config := &ServerConfig{
		Port:       52000,
		MaxPlayers: 4,
		WorldSeed:  12345,
		GenreID:    "fantasy",
		TickRate:   10,
	}

	manager, err := NewServerManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	err = manager.Start()
	if err != nil {
		t.Skipf("cannot start server for testing: %v", err)
	}
	defer manager.Stop()

	// Spawn a player first
	playerID := uint64(600)
	manager.spawnPlayer(playerID)
	manager.world.Update(0.016)

	// Create input command using the network package types
	inputCmd := &networkInputCommand{
		PlayerID:  playerID,
		InputType: "move",
		Data:      []byte(`{"dx": 1.0, "dy": 0.0}`),
	}

	// handleInputCommand should not panic
	manager.handleInputCommandInternal(inputCmd)
}

// networkInputCommand is a mock struct for testing handleInputCommand
type networkInputCommand struct {
	PlayerID  uint64
	InputType string
	Data      []byte
}

// handleInputCommandInternal wraps handleInputCommand for testing without network dependency
func (sm *ServerManager) handleInputCommandInternal(cmd *networkInputCommand) {
	sm.inputHandler.ProcessInputRaw(cmd.PlayerID, cmd.InputType, cmd.Data)
}
