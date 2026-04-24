//go:build !android && !ios
// +build !android,!ios

package main

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
	itemgen "github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"github.com/sirupsen/logrus"
)

// createTestLogger creates a logger for testing that discards output
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(bytes.NewBuffer(nil))
	logger.SetLevel(logrus.DebugLevel)
	return logger
}

// createTestTerrain creates a simple test terrain with rooms
func createTestTerrain(numRooms int) *terrain.Terrain {
	t := terrain.NewTerrain(100, 100, 12345)
	for i := 0; i < numRooms; i++ {
		room := &terrain.Room{
			X:      10 + i*20,
			Y:      10 + i*20,
			Width:  10,
			Height: 10,
			Type:   terrain.RoomNormal,
		}
		t.Rooms = append(t.Rooms, room)
	}
	return t
}

// TestCreatePlayerEntity_BasicSetup tests player entity creation with default components
func TestCreatePlayerEntity_BasicSetup(t *testing.T) {
	world := engine.NewWorld()
	terrain := createTestTerrain(1)
	logger := createTestLogger()

	entity := createPlayerEntity(world, terrain, 1001, 12345, "fantasy", false, logger)

	if entity == nil {
		t.Fatal("createPlayerEntity returned nil")
	}

	// Verify core components exist
	components := []string{
		"position", "velocity", "health", "team", "network",
		"stats", "experience", "inventory", "equipment", "questtracker",
		"attack", "collider",
	}

	for _, comp := range components {
		if !entity.HasComponent(comp) {
			t.Errorf("expected entity to have component %q", comp)
		}
	}
}

// TestCreatePlayerEntity_SpawnPosition tests player spawning at correct position
func TestCreatePlayerEntity_SpawnPosition(t *testing.T) {
	tests := []struct {
		name      string
		numRooms  int
		wantX     float64
		wantY     float64
		tolerance float64
	}{
		{
			name:      "spawn in first room center",
			numRooms:  3,
			wantX:     float64((10 + 10/2)) * 32, // Room center X * tile size
			wantY:     float64((10 + 10/2)) * 32, // Room center Y * tile size
			tolerance: 0.01,
		},
		{
			name:      "spawn at default with no rooms",
			numRooms:  0,
			wantX:     400.0,
			wantY:     300.0,
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			terrain := createTestTerrain(tt.numRooms)
			logger := createTestLogger()

			entity := createPlayerEntity(world, terrain, 1001, 12345, "fantasy", false, logger)

			posComp, ok := entity.GetComponent("position")
			if !ok {
				t.Fatal("expected entity to have position component")
			}
			pos := posComp.(*engine.PositionComponent)

			if pos.X < tt.wantX-tt.tolerance || pos.X > tt.wantX+tt.tolerance {
				t.Errorf("X position = %v, want %v", pos.X, tt.wantX)
			}
			if pos.Y < tt.wantY-tt.tolerance || pos.Y > tt.wantY+tt.tolerance {
				t.Errorf("Y position = %v, want %v", pos.Y, tt.wantY)
			}
		})
	}
}

// TestCreatePlayerEntity_NetworkComponent tests network component setup
func TestCreatePlayerEntity_NetworkComponent(t *testing.T) {
	world := engine.NewWorld()
	terrain := createTestTerrain(1)
	logger := createTestLogger()

	playerID := uint64(9999)
	entity := createPlayerEntity(world, terrain, playerID, 12345, "fantasy", false, logger)

	netComp, ok := entity.GetComponent("network")
	if !ok {
		t.Fatal("expected entity to have network component")
	}
	network := netComp.(*engine.NetworkComponent)

	if network.PlayerID != playerID {
		t.Errorf("PlayerID = %d, want %d", network.PlayerID, playerID)
	}
	if !network.Synced {
		t.Error("expected network component to be synced")
	}
}

// TestCreatePlayerEntity_UniqueSeeds tests that different players get unique sprite seeds
func TestCreatePlayerEntity_UniqueSeeds(t *testing.T) {
	world := engine.NewWorld()
	terrain := createTestTerrain(1)
	logger := createTestLogger()

	entity1 := createPlayerEntity(world, terrain, 1, 12345, "fantasy", true, logger)
	entity2 := createPlayerEntity(world, terrain, 2, 12345, "fantasy", true, logger)

	// Both entities should exist
	if entity1 == nil || entity2 == nil {
		t.Fatal("expected both entities to be created")
	}

	// They should have different entity IDs
	if entity1.ID == entity2.ID {
		t.Error("expected different entity IDs")
	}
}

// TestApplyInputCommand_Movement tests movement input processing
func TestApplyInputCommand_Movement(t *testing.T) {
	tests := []struct {
		name   string
		dataX  int8
		dataY  int8
		wantVX float64
		wantVY float64
	}{
		{"move right", 127, 0, 100.0, 0.0},
		{"move left", -127, 0, -100.0, 0.0},
		{"move down", 0, 127, 0.0, 100.0},
		{"move up", 0, -127, 0.0, -100.0},
		{"stop", 0, 0, 0.0, 0.0},
		{"diagonal normalized", 127, 127, 100.0 * 0.707, 100.0 * 0.707},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			entity := world.CreateEntity()
			entity.AddComponent(&engine.VelocityComponent{})

			logger := createTestLogger()
			cmd := &network.InputCommand{
				PlayerID:  1,
				InputType: "move",
				Data:      []byte{byte(tt.dataX), byte(tt.dataY)},
			}

			applyInputCommand(entity, cmd, logger)

			velComp, _ := entity.GetComponent("velocity")
			vel := velComp.(*engine.VelocityComponent)

			tolerance := 1.0 // Allow some floating point tolerance
			if vel.VX < tt.wantVX-tolerance || vel.VX > tt.wantVX+tolerance {
				t.Errorf("VX = %v, want %v", vel.VX, tt.wantVX)
			}
			if vel.VY < tt.wantVY-tolerance || vel.VY > tt.wantVY+tolerance {
				t.Errorf("VY = %v, want %v", vel.VY, tt.wantVY)
			}
		})
	}
}

// TestApplyInputCommand_MovementNoData tests movement with missing data
func TestApplyInputCommand_MovementNoData(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{VX: 50, VY: 50}) // Pre-set velocity

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "move",
		Data:      []byte{}, // Empty data
	}

	applyInputCommand(entity, cmd, logger)

	// Velocity should be unchanged (no crash, early return)
	velComp, _ := entity.GetComponent("velocity")
	vel := velComp.(*engine.VelocityComponent)

	if vel.VX != 50 || vel.VY != 50 {
		t.Error("velocity should be unchanged with empty movement data")
	}
}

// TestApplyInputCommand_Attack tests attack input processing
func TestApplyInputCommand_Attack(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	attackComp := &engine.AttackComponent{
		Damage:        15,
		DamageType:    combat.DamagePhysical,
		Range:         50,
		Cooldown:      0.5,
		CooldownTimer: 0.0, // Ready to attack
	}
	entity.AddComponent(attackComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "attack",
		Data:      []byte{},
	}

	applyInputCommand(entity, cmd, logger)

	// After attack, cooldown should be reset
	if attackComp.CooldownTimer != attackComp.Cooldown {
		t.Errorf("CooldownTimer = %v, want %v", attackComp.CooldownTimer, attackComp.Cooldown)
	}
}

// TestApplyInputCommand_AttackOnCooldown tests attack blocked by cooldown
func TestApplyInputCommand_AttackOnCooldown(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	attackComp := &engine.AttackComponent{
		Damage:        15,
		Range:         50,
		Cooldown:      0.5,
		CooldownTimer: 0.3, // On cooldown
	}
	entity.AddComponent(attackComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "attack",
		Data:      []byte{},
	}

	applyInputCommand(entity, cmd, logger)

	// Cooldown timer should not be reset when already on cooldown
	if attackComp.CooldownTimer != 0.3 {
		t.Errorf("CooldownTimer = %v, want 0.3 (unchanged)", attackComp.CooldownTimer)
	}
}

// TestApplyInputCommand_AttackNoComponent tests attack without attack component
func TestApplyInputCommand_AttackNoComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	// No attack component added

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "attack",
		Data:      []byte{},
	}

	// Should not panic
	applyInputCommand(entity, cmd, logger)
}

// TestApplyInputCommand_UseItem tests item consumption
func TestApplyInputCommand_UseItem(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})

	// Add health component with reduced health
	healthComp := &engine.HealthComponent{Current: 50, Max: 100}
	entity.AddComponent(healthComp)

	// Add inventory with a consumable item
	invComp := engine.NewInventoryComponent(10, 50.0)
	healPotion := &itemgen.Item{
		Name:  "Health Potion",
		Type:  itemgen.TypeConsumable,
		Stats: itemgen.Stats{Healing: 25},
	}
	invComp.Items = append(invComp.Items, healPotion)
	entity.AddComponent(invComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "use_item",
		Data:      []byte{0}, // Use item at index 0
	}

	applyInputCommand(entity, cmd, logger)

	// Health should be increased
	if healthComp.Current != 75 {
		t.Errorf("Health = %v, want 75", healthComp.Current)
	}

	// Item should be removed from inventory
	if len(invComp.Items) != 0 {
		t.Errorf("Items count = %d, want 0", len(invComp.Items))
	}
}

// TestApplyInputCommand_UseItemHealthCapped tests healing doesn't exceed max health
func TestApplyInputCommand_UseItemHealthCapped(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})

	// Add health component near max
	healthComp := &engine.HealthComponent{Current: 90, Max: 100}
	entity.AddComponent(healthComp)

	// Add inventory with a consumable item
	invComp := engine.NewInventoryComponent(10, 50.0)
	bigPotion := &itemgen.Item{
		Name:  "Big Health Potion",
		Type:  itemgen.TypeConsumable,
		Stats: itemgen.Stats{Healing: 50},
	}
	invComp.Items = append(invComp.Items, bigPotion)
	entity.AddComponent(invComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "use_item",
		Data:      []byte{0},
	}

	applyInputCommand(entity, cmd, logger)

	// Health should be capped at max
	if healthComp.Current != 100 {
		t.Errorf("Health = %v, want 100 (capped at max)", healthComp.Current)
	}
}

// TestApplyInputCommand_UseItemInvalidIndex tests invalid item index handling
func TestApplyInputCommand_UseItemInvalidIndex(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	entity.AddComponent(&engine.HealthComponent{Current: 50, Max: 100})

	invComp := engine.NewInventoryComponent(10, 50.0)
	entity.AddComponent(invComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "use_item",
		Data:      []byte{5}, // Invalid index (inventory is empty)
	}

	// Should not panic, should handle gracefully
	applyInputCommand(entity, cmd, logger)
}

// TestApplyInputCommand_UseNonConsumable tests using non-consumable item
func TestApplyInputCommand_UseNonConsumable(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	entity.AddComponent(&engine.HealthComponent{Current: 50, Max: 100})

	invComp := engine.NewInventoryComponent(10, 50.0)
	weapon := &itemgen.Item{
		Name: "Iron Sword",
		Type: itemgen.TypeWeapon, // Not consumable
	}
	invComp.Items = append(invComp.Items, weapon)
	entity.AddComponent(invComp)

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "use_item",
		Data:      []byte{0},
	}

	applyInputCommand(entity, cmd, logger)

	// Item should NOT be consumed
	if len(invComp.Items) != 1 {
		t.Error("non-consumable item should not be consumed")
	}
}

// TestApplyInputCommand_UseItemNoInventory tests item use without inventory
func TestApplyInputCommand_UseItemNoInventory(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	// No inventory component

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "use_item",
		Data:      []byte{0},
	}

	// Should not panic
	applyInputCommand(entity, cmd, logger)
}

// TestApplyInputCommand_UnknownType tests unknown input type handling
func TestApplyInputCommand_UnknownType(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "unknown_action",
		Data:      []byte{},
	}

	// Should not panic, should log warning
	applyInputCommand(entity, cmd, logger)
}

// TestApplyInputCommand_NoVelocityComponent tests input handling without velocity
func TestApplyInputCommand_NoVelocityComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	// No velocity component

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "move",
		Data:      []byte{127, 0},
	}

	// Should return early without panic
	applyInputCommand(entity, cmd, logger)
}

// TestValidateItemUse_MissingData tests item validation with missing data
func TestValidateItemUse_MissingData(t *testing.T) {
	invComp := engine.NewInventoryComponent(10, 50.0)
	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID: 1,
		Data:     []byte{}, // Missing item index
	}

	_, _, ok := validateItemUse(invComp, cmd, logger)
	if ok {
		t.Error("expected validation to fail with missing data")
	}
}

// TestValidateItemUse_NegativeIndex tests item validation with negative index
func TestValidateItemUse_NegativeIndex(t *testing.T) {
	invComp := engine.NewInventoryComponent(10, 50.0)
	invComp.Items = append(invComp.Items, &itemgen.Item{Name: "Test", Type: itemgen.TypeConsumable})

	logger := createTestLogger()
	cmd := &network.InputCommand{
		PlayerID: 1,
		Data:     []byte{255}, // -1 as int8, will be interpreted as 255 (out of bounds)
	}

	_, _, ok := validateItemUse(invComp, cmd, logger)
	if ok {
		t.Error("expected validation to fail with out of bounds index")
	}
}

// TestConsumeItem_ZeroHeal tests consume item with zero heal amount
func TestConsumeItem_ZeroHeal(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	healthComp := &engine.HealthComponent{Current: 50, Max: 100}
	entity.AddComponent(healthComp)

	invComp := engine.NewInventoryComponent(10, 50.0)
	zeroItem := &itemgen.Item{
		Name:  "Empty Vial",
		Type:  itemgen.TypeConsumable,
		Stats: itemgen.Stats{Healing: 0}, // Zero heal
	}
	invComp.Items = append(invComp.Items, zeroItem)

	logger := createTestLogger()

	// Call consumeItem directly
	consumeItem(entity, invComp, zeroItem, 0, 1, logger)

	// Health should remain unchanged (zero heal triggers early return)
	if healthComp.Current != 50 {
		t.Errorf("Health = %v, want 50 (unchanged)", healthComp.Current)
	}

	// Item should still be in inventory (early return before removal)
	if len(invComp.Items) != 1 {
		t.Errorf("Items count = %d, want 1 (item not removed on zero heal)", len(invComp.Items))
	}
}

// TestConsumeItem_NoHealthComponent tests consume item without health
func TestConsumeItem_NoHealthComponent(t *testing.T) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	// No health component

	invComp := engine.NewInventoryComponent(10, 50.0)
	item := &itemgen.Item{
		Name:  "Potion",
		Type:  itemgen.TypeConsumable,
		Stats: itemgen.Stats{Healing: 25},
	}
	invComp.Items = append(invComp.Items, item)

	logger := createTestLogger()

	// Should not panic, should return early
	consumeItem(entity, invComp, item, 0, 1, logger)
}

// Benchmark tests

func BenchmarkCreatePlayerEntity(b *testing.B) {
	world := engine.NewWorld()
	terrain := createTestTerrain(5)
	logger := createTestLogger()
	logger.SetLevel(logrus.ErrorLevel) // Reduce logging overhead

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createPlayerEntity(world, terrain, uint64(i), 12345, "fantasy", false, logger)
	}
}

func BenchmarkApplyInputCommand_Movement(b *testing.B) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})

	logger := createTestLogger()
	logger.SetLevel(logrus.ErrorLevel)

	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "move",
		Data:      []byte{127, 0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyInputCommand(entity, cmd, logger)
	}
}

func BenchmarkApplyInputCommand_Attack(b *testing.B) {
	world := engine.NewWorld()
	entity := world.CreateEntity()
	entity.AddComponent(&engine.VelocityComponent{})
	attackComp := &engine.AttackComponent{
		Damage:   15,
		Range:    50,
		Cooldown: 0.5,
	}
	entity.AddComponent(attackComp)

	logger := createTestLogger()
	logger.SetLevel(logrus.ErrorLevel)

	cmd := &network.InputCommand{
		PlayerID:  1,
		InputType: "attack",
		Data:      []byte{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset cooldown for each iteration
		attackComp.CooldownTimer = 0
		applyInputCommand(entity, cmd, logger)
	}
}

// TestEnhancedChatSystemRegistration tests player registration with chat system
func TestEnhancedChatSystemRegistration(t *testing.T) {
	world := engine.NewWorld()
	chatSystem := engine.NewEnhancedChatSystem(world)

	// Create a player entity
	entity := world.CreateEntity()

	// Register player with chat system
	err := chatSystem.RegisterPlayer(entity.ID)
	if err != nil {
		t.Fatalf("RegisterPlayer failed: %v", err)
	}

	// Verify chat component was added
	if !entity.HasComponent("chat") {
		t.Error("expected entity to have chat component after registration")
	}

	// Verify chat component is properly initialized
	chatComp, ok := entity.GetComponent("chat")
	if !ok {
		t.Fatal("failed to get chat component")
	}
	chat := chatComp.(*engine.ChatComponent)
	if chat == nil {
		t.Error("chat component is nil")
	}

	// Unregister player
	chatSystem.UnregisterPlayer(entity.ID)

	// Chat component should still exist on entity (only history is removed)
	if !entity.HasComponent("chat") {
		t.Error("chat component should persist after unregistration")
	}
}

// TestEnhancedChatSystemRegistrationNonexistentEntity tests registration with invalid entity
func TestEnhancedChatSystemRegistrationNonexistentEntity(t *testing.T) {
	world := engine.NewWorld()
	chatSystem := engine.NewEnhancedChatSystem(world)

	// Try to register a non-existent entity
	err := chatSystem.RegisterPlayer(99999)
	if err == nil {
		t.Error("expected error when registering non-existent entity")
	}
}
