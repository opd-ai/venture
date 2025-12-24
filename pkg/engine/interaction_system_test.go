package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestPlayerHasItemByID_WithMatchingItem tests that playerHasItemByID returns true when the player has the item.
func TestPlayerHasItemByID_WithMatchingItem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player entity with inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)

	// Add test item with specific ID
	testItem := &item.Item{
		ID:   "key-001",
		Name: "Ancient Key",
		Type: item.TypeConsumable,
	}
	inventory.AddItem(testItem)
	player.AddComponent(inventory)

	// Test: Player has the item
	if !system.playerHasItemByID(player, "key-001") {
		t.Error("expected playerHasItemByID to return true when player has the item")
	}
}

// TestPlayerHasItemByID_WithoutMatchingItem tests that playerHasItemByID returns false when the player doesn't have the item.
func TestPlayerHasItemByID_WithoutMatchingItem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player entity with inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)

	// Add different item
	testItem := &item.Item{
		ID:   "key-002",
		Name: "Bronze Key",
		Type: item.TypeConsumable,
	}
	inventory.AddItem(testItem)
	player.AddComponent(inventory)

	// Test: Player doesn't have the requested item
	if system.playerHasItemByID(player, "key-001") {
		t.Error("expected playerHasItemByID to return false when player doesn't have the item")
	}
}

// TestPlayerHasItemByID_EmptyInventory tests that playerHasItemByID returns false when inventory is empty.
func TestPlayerHasItemByID_EmptyInventory(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player entity with empty inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	player.AddComponent(inventory)

	// Test: Empty inventory
	if system.playerHasItemByID(player, "key-001") {
		t.Error("expected playerHasItemByID to return false when inventory is empty")
	}
}

// TestPlayerHasItemByID_NoInventoryComponent tests that playerHasItemByID returns false when player has no inventory.
func TestPlayerHasItemByID_NoInventoryComponent(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player entity without inventory
	player := world.CreateEntity()

	// Test: No inventory component
	if system.playerHasItemByID(player, "key-001") {
		t.Error("expected playerHasItemByID to return false when player has no inventory component")
	}
}

// TestPlayerHasItemByID_NilItem tests that playerHasItemByID handles nil items gracefully.
func TestPlayerHasItemByID_NilItem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player entity with inventory containing nil
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	// Directly append nil to test edge case - AddItem would reject nil
	inventory.Items = append(inventory.Items, nil)
	player.AddComponent(inventory)

	// Test: Inventory contains nil items (should not crash)
	if system.playerHasItemByID(player, "key-001") {
		t.Error("expected playerHasItemByID to return false when inventory contains only nil items")
	}
}

// TestHandleBookshelfRead_UnlockedBookshelf tests normal bookshelf access.
func TestHandleBookshelfRead_UnlockedBookshelf(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player (no PlayerComponent needed - handleBookshelfRead doesn't check for it)
	player := world.CreateEntity()

	// Create unlocked bookshelf with books
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1) // Add a book
	bookshelf.IsLocked = false
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should allow access to unlocked bookshelf
	// This should not panic and should log successful browsing
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf remains unlocked
	if bookshelf.IsLocked {
		t.Error("bookshelf should remain unlocked")
	}
}

// TestHandleBookshelfRead_LockedWithoutKey tests that locked bookshelf blocks access without key.
func TestHandleBookshelfRead_LockedWithoutKey(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player with empty inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	player.AddComponent(inventory)

	// Create locked bookshelf requiring a key
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1) // Add a book
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "key-ancient-library"
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should block access when player doesn't have key
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf remains locked
	if !bookshelf.IsLocked {
		t.Error("bookshelf should remain locked when player doesn't have key")
	}
}

// TestHandleBookshelfRead_LockedWithKey tests that locked bookshelf unlocks with correct key.
func TestHandleBookshelfRead_LockedWithKey(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player with the required key in inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	keyItem := &item.Item{
		ID:   "key-ancient-library",
		Name: "Ancient Library Key",
		Type: item.TypeConsumable,
	}
	inventory.AddItem(keyItem)
	player.AddComponent(inventory)

	// Create locked bookshelf requiring the key
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1) // Add a book
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "key-ancient-library"
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should unlock bookshelf when player has key
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf is now unlocked
	if bookshelf.IsLocked {
		t.Error("bookshelf should be unlocked after player uses correct key")
	}
}

// TestHandleBookshelfRead_LockedWithWrongKey tests that locked bookshelf doesn't unlock with wrong key.
func TestHandleBookshelfRead_LockedWithWrongKey(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player with a different key in inventory
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	wrongKey := &item.Item{
		ID:   "key-bronze-chest",
		Name: "Bronze Chest Key",
		Type: item.TypeConsumable,
	}
	inventory.AddItem(wrongKey)
	player.AddComponent(inventory)

	// Create locked bookshelf requiring different key
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1) // Add a book
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "key-ancient-library"
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should remain locked when player has wrong key
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf remains locked
	if !bookshelf.IsLocked {
		t.Error("bookshelf should remain locked when player has wrong key")
	}
}

// TestHandleBookshelfRead_LockedNoKeyRequired tests bookshelf locked without key requirement.
func TestHandleBookshelfRead_LockedNoKeyRequired(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	player.AddComponent(inventory)

	// Create locked bookshelf that doesn't require a key (edge case)
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1) // Add a book
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = false // Locked but no key required
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should block access (locked without unlock mechanism)
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf remains locked
	if !bookshelf.IsLocked {
		t.Error("bookshelf should remain locked when no key mechanism configured")
	}
}

// TestHandleBookshelfRead_EmptyBookshelf tests that empty bookshelves are handled properly.
func TestHandleBookshelfRead_EmptyBookshelf(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player
	player := world.CreateEntity()

	// Create empty unlocked bookshelf
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.IsLocked = false
	// Don't add any books
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should handle empty bookshelf gracefully
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// No assertions needed - just verify no panic
}

// TestHandleBookshelfRead_UnlockThenBrowse tests full flow of unlocking and browsing.
func TestHandleBookshelfRead_UnlockThenBrowse(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player with key
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)
	keyItem := &item.Item{
		ID:   "master-key",
		Name: "Master Key",
		Type: item.TypeConsumable,
	}
	inventory.AddItem(keyItem)
	player.AddComponent(inventory)

	// Create locked bookshelf with books
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1)
	bookshelf.AddBook(2)
	bookshelf.AddBook(3)
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "master-key"
	bookshelfEntity.AddComponent(bookshelf)

	// First interaction: Unlock
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)
	if bookshelf.IsLocked {
		t.Error("bookshelf should be unlocked after first interaction with key")
	}

	// Second interaction: Browse (already unlocked)
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)
	if bookshelf.IsLocked {
		t.Error("bookshelf should remain unlocked after second interaction")
	}
}

// TestHandleBookshelfRead_MultipleKeys tests player with multiple keys including correct one.
func TestHandleBookshelfRead_MultipleKeys(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)

	// Create player with multiple keys
	player := world.CreateEntity()
	inventory := NewInventoryComponent(10, 100.0)

	// Add several keys
	inventory.AddItem(&item.Item{
		ID:   "key-001",
		Name: "Bronze Key",
		Type: item.TypeConsumable,
	})
	inventory.AddItem(&item.Item{
		ID:   "key-002",
		Name: "Silver Key",
		Type: item.TypeConsumable,
	})
	inventory.AddItem(&item.Item{
		ID:   "key-ancient-tome",
		Name: "Ancient Tome Key",
		Type: item.TypeConsumable,
	})
	player.AddComponent(inventory)

	// Create locked bookshelf requiring specific key
	bookshelfEntity := world.CreateEntity()
	bookshelf := NewBookshelfComponent(10, "fantasy")
	bookshelf.AddBook(1)
	bookshelf.IsLocked = true
	bookshelf.RequiresKey = true
	bookshelf.KeyItemID = "key-ancient-tome"
	bookshelfEntity.AddComponent(bookshelf)

	// Test: Should unlock with correct key from inventory
	system.handleBookshelfRead(player, bookshelfEntity, bookshelf)

	// Verify bookshelf is unlocked
	if bookshelf.IsLocked {
		t.Error("bookshelf should be unlocked when player has correct key among multiple keys")
	}
}

// TestInteractionSystem_SetMiniGameSystem tests the SetMiniGameSystem setter.
func TestInteractionSystem_SetMiniGameSystem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)

	system.SetMiniGameSystem(miniGameSystem)

	if system.miniGameSystem != miniGameSystem {
		t.Error("expected miniGameSystem to be set")
	}
}

// TestInteractionSystem_SetMiniGameSystem_Nil tests setting nil mini-game system.
func TestInteractionSystem_SetMiniGameSystem_Nil(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)

	// Set system first
	system.SetMiniGameSystem(miniGameSystem)

	// Then set to nil (graceful handling)
	system.SetMiniGameSystem(nil)

	if system.miniGameSystem != nil {
		t.Error("expected miniGameSystem to be nil")
	}
}

// TestInteractionSystem_HandleOpenAction_WithLockPicking tests lock-picking mini-game start.
func TestInteractionSystem_HandleOpenAction_WithLockPicking(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create locked door entity
	door := world.CreateEntity()
	door.AddComponent(&PositionComponent{X: 150, Y: 100})
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		ActionText:          "Open",
		InteractionRange:    50.0,
		IsAvailable:         true,
		RequiresLockPicking: true,
		LockDifficulty:      0.5,
	}
	door.AddComponent(ctx)

	// Flush entities so they are available in world.entities map
	world.FlushPendingEntities()

	// Test: handleOpenAction should start lock-picking mini-game
	system.handleOpenAction(player, door)

	// Verify mini-game was started by checking if component exists
	compRaw, hasComp := player.GetComponent("minigame")
	if !hasComp {
		t.Fatal("expected player to have minigame component")
	}

	gameComp, ok := compRaw.(*MiniGameComponent)
	if !ok || gameComp == nil {
		t.Fatal("expected mini-game component to be created")
	}

	if gameComp.GameType != MiniGameLockPicking {
		t.Errorf("expected game type MiniGameLockPicking, got %v", gameComp.GameType)
	}

	if gameComp.Difficulty != 0.5 {
		t.Errorf("expected difficulty 0.5, got %v", gameComp.Difficulty)
	}

	if !gameComp.Active {
		t.Error("expected mini-game to be active")
	}

	// Verify locked entity ID is stored in state
	stateMap, ok := gameComp.State.(map[string]interface{})
	if !ok {
		t.Fatal("expected state to be a map")
	}

	lockedEntityID, ok := stateMap["lockedEntityID"].(uint64)
	if !ok || lockedEntityID != door.ID {
		t.Errorf("expected lockedEntityID to be %d, got %v", door.ID, lockedEntityID)
	}

	// Verify door remains locked (not opened yet)
	if ctx.ActionType != ActionOpen {
		t.Errorf("expected door to remain locked with ActionOpen, got %v", ctx.ActionType)
	}
}

// TestInteractionSystem_HandleOpenAction_NoMiniGameSystem tests graceful degradation without mini-game system.
func TestInteractionSystem_HandleOpenAction_NoMiniGameSystem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	// Don't set miniGameSystem - test graceful degradation

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	// Create locked door entity
	door := world.CreateEntity()
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		ActionText:          "Open",
		RequiresLockPicking: true,
		LockDifficulty:      0.5,
	}
	door.AddComponent(ctx)

	// Test: handleOpenAction should not crash without mini-game system
	system.handleOpenAction(player, door)

	// Verify door remains locked
	if ctx.ActionType != ActionOpen {
		t.Errorf("expected door to remain locked with ActionOpen, got %v", ctx.ActionType)
	}
}

// TestInteractionSystem_HandleOpenAction_NormalDoor tests normal door opening without lock-picking.
func TestInteractionSystem_HandleOpenAction_NormalDoor(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	// Create unlocked door entity
	door := world.CreateEntity()
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		ActionText:          "Open",
		RequiresLockPicking: false, // No lock-picking required
	}
	door.AddComponent(ctx)

	// Test: handleOpenAction should open door normally
	system.handleOpenAction(player, door)

	// Verify door is opened (ActionType changed to Close)
	if ctx.ActionType != ActionClose {
		t.Errorf("expected door to be opened with ActionClose, got %v", ctx.ActionType)
	}

	if ctx.ActionText != "Close" {
		t.Errorf("expected action text to be 'Close', got %q", ctx.ActionText)
	}

	// Verify no mini-game was started
	gameComp := miniGameSystem.GetGameComponent(player.ID)
	if gameComp != nil {
		t.Error("expected no mini-game to be started for normal door")
	}
}

// TestInteractionSystem_HandleOpenAction_DifficultyNormalization tests difficulty clamping.
func TestInteractionSystem_HandleOpenAction_DifficultyNormalization(t *testing.T) {
	tests := []struct {
		name               string
		inputDifficulty    float64
		expectedDifficulty float64
	}{
		{"negative difficulty", -0.5, 0.0},
		{"above max difficulty", 1.5, 1.0},
		{"normal difficulty", 0.7, 0.7},
		{"zero difficulty", 0.0, 0.0},
		{"max difficulty", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			system := NewInteractionSystem(world)
			miniGameSystem := NewMiniGameSystem(world)
			system.SetMiniGameSystem(miniGameSystem)

			player := world.CreateEntity()
			player.AddComponent(&PlayerComponent{})

			door := world.CreateEntity()
			ctx := &ContextActionComponent{
				ActionType:          ActionOpen,
				RequiresLockPicking: true,
				LockDifficulty:      tt.inputDifficulty,
			}
			door.AddComponent(ctx)

			// Flush entities so they are available in world.entities map
			world.FlushPendingEntities()

			system.handleOpenAction(player, door)

			gameComp := miniGameSystem.GetGameComponent(player.ID)
			if gameComp == nil {
				t.Fatal("expected mini-game component to be created")
			}

			if gameComp.Difficulty != tt.expectedDifficulty {
				t.Errorf("expected difficulty %v, got %v", tt.expectedDifficulty, gameComp.Difficulty)
			}
		})
	}
}

// TestInteractionSystem_ProcessLockPickingCompletion_Success tests successful lock-picking.
func TestInteractionSystem_ProcessLockPickingCompletion_Success(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	// Create player and locked door
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	door := world.CreateEntity()
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		ActionText:          "Open",
		RequiresLockPicking: true,
		LockDifficulty:      0.5,
	}
	door.AddComponent(ctx)

	// Flush entities so they are available in world.entities map
	world.FlushPendingEntities()

	// Start lock-picking mini-game
	system.handleOpenAction(player, door)

	// Simulate successful mini-game completion
	system.ProcessLockPickingCompletion(player.ID, true)

	// Verify door is now opened
	if ctx.ActionType != ActionClose {
		t.Errorf("expected door to be opened with ActionClose, got %v", ctx.ActionType)
	}

	if ctx.ActionText != "Close" {
		t.Errorf("expected action text to be 'Close', got %q", ctx.ActionText)
	}

	if ctx.RequiresLockPicking {
		t.Error("expected RequiresLockPicking to be false after successful pick")
	}
}

// TestInteractionSystem_ProcessLockPickingCompletion_Failure tests failed lock-picking.
func TestInteractionSystem_ProcessLockPickingCompletion_Failure(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	// Create player and locked door
	player := world.CreateEntity()
	player.AddComponent(&PlayerComponent{})

	door := world.CreateEntity()
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		ActionText:          "Open",
		RequiresLockPicking: true,
		LockDifficulty:      0.5,
	}
	door.AddComponent(ctx)

	// Flush entities so they are available in world.entities map
	world.FlushPendingEntities()

	// Start lock-picking mini-game
	system.handleOpenAction(player, door)

	// Simulate failed mini-game completion
	system.ProcessLockPickingCompletion(player.ID, false)

	// Verify door remains locked
	if ctx.ActionType != ActionOpen {
		t.Errorf("expected door to remain locked with ActionOpen, got %v", ctx.ActionType)
	}

	if ctx.ActionText != "Open" {
		t.Errorf("expected action text to remain 'Open', got %q", ctx.ActionText)
	}

	if !ctx.RequiresLockPicking {
		t.Error("expected RequiresLockPicking to remain true after failed pick")
	}
}

// TestInteractionSystem_ProcessLockPickingCompletion_NoMiniGameSystem tests graceful handling without system.
func TestInteractionSystem_ProcessLockPickingCompletion_NoMiniGameSystem(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	// Don't set miniGameSystem

	player := world.CreateEntity()

	// Should not crash
	system.ProcessLockPickingCompletion(player.ID, true)
}

// TestInteractionSystem_ProcessLockPickingCompletion_NoGameComponent tests handling when no game is active.
func TestInteractionSystem_ProcessLockPickingCompletion_NoGameComponent(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	player := world.CreateEntity()

	// No mini-game was started, so this should be a no-op
	system.ProcessLockPickingCompletion(player.ID, true)

	// Should not crash and should not affect any entities
}

// TestInteractionSystem_ProcessLockPickingCompletion_WrongGameType tests handling of non-lock-picking games.
func TestInteractionSystem_ProcessLockPickingCompletion_WrongGameType(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	player := world.CreateEntity()
	door := world.CreateEntity()
	ctx := &ContextActionComponent{
		ActionType:          ActionOpen,
		RequiresLockPicking: true,
	}
	door.AddComponent(ctx)

	// Flush entities so they are available in world.entities map
	world.FlushPendingEntities()

	// Start a different game type (not lock-picking)
	miniGameSystem.StartGame(player.ID, MiniGameCard, 0.5)

	// Try to process as lock-picking completion
	system.ProcessLockPickingCompletion(player.ID, true)

	// Door should remain locked since it wasn't a lock-picking game
	if ctx.ActionType != ActionOpen {
		t.Errorf("expected door to remain locked, got %v", ctx.ActionType)
	}
}

// TestInteractionSystem_ProcessLockPickingCompletion_InvalidEntityID tests handling of invalid locked entity.
func TestInteractionSystem_ProcessLockPickingCompletion_InvalidEntityID(t *testing.T) {
	world := NewWorld()
	system := NewInteractionSystem(world)
	miniGameSystem := NewMiniGameSystem(world)
	system.SetMiniGameSystem(miniGameSystem)

	player := world.CreateEntity()

	// Flush entities so they are available in world.entities map
	world.FlushPendingEntities()

	// Start lock-picking with manually created state
	miniGameSystem.StartGame(player.ID, MiniGameLockPicking, 0.5)
	gameComp := miniGameSystem.GetGameComponent(player.ID)
	if gameComp != nil {
		gameComp.State = map[string]interface{}{
			"lockedEntityID": uint64(99999), // Non-existent entity
		}
	}

	// Should not crash
	system.ProcessLockPickingCompletion(player.ID, true)
}
