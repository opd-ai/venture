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
