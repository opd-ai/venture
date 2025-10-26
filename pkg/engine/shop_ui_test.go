package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestShopMode_String tests the string representation of shop modes.
func TestShopMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode ShopMode
		want string
	}{
		{"buy mode", ShopModeBuy, "Buy"},
		{"sell mode", ShopModeSell, "Sell"},
		{"unknown mode", ShopMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("ShopMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewShopUI tests shop UI creation.
func TestNewShopUI(t *testing.T) {
	ui := NewShopUI(800, 600)

	if ui == nil {
		t.Fatal("NewShopUI() returned nil")
	}

	// Verify initial state
	if ui.visible {
		t.Error("New shop should not be visible")
	}

	if ui.mode != ShopModeBuy {
		t.Errorf("Initial mode = %v, want %v", ui.mode, ShopModeBuy)
	}

	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}

	if ui.selectedSlot != -1 {
		t.Errorf("Initial selectedSlot = %d, want -1", ui.selectedSlot)
	}

	if ui.hoveredSlot != -1 {
		t.Errorf("Initial hoveredSlot = %d, want -1", ui.hoveredSlot)
	}
}

// TestShopUI_SettersAndGetters tests setter and getter methods.
func TestShopUI_SettersAndGetters(t *testing.T) {
	world := NewWorld()
	ui := NewShopUI(800, 600)

	// Create test entities
	player := world.CreateEntity()
	merchant := world.CreateEntity()

	// Test SetPlayerEntity
	ui.SetPlayerEntity(player)
	if ui.playerEntity != player {
		t.Error("SetPlayerEntity() did not set player entity")
	}

	// Test SetMerchantEntity
	ui.SetMerchantEntity(merchant)
	if ui.merchantEntity != merchant {
		t.Error("SetMerchantEntity() did not set merchant entity")
	}

	// Create systems
	commerceSystem := NewCommerceSystem(world, NewInventorySystem(world))
	dialogSystem := NewDialogSystem(world)

	// Test SetCommerceSystem
	ui.SetCommerceSystem(commerceSystem)
	if ui.commerceSystem != commerceSystem {
		t.Error("SetCommerceSystem() did not set commerce system")
	}

	// Test SetDialogSystem
	ui.SetDialogSystem(dialogSystem)
	if ui.dialogSystem != dialogSystem {
		t.Error("SetDialogSystem() did not set dialog system")
	}

	// Test GetMode
	if ui.GetMode() != ShopModeBuy {
		t.Errorf("GetMode() = %v, want %v", ui.GetMode(), ShopModeBuy)
	}

	// Test SetMode
	ui.SetMode(ShopModeSell)
	if ui.GetMode() != ShopModeSell {
		t.Errorf("After SetMode(Sell), GetMode() = %v, want %v", ui.GetMode(), ShopModeSell)
	}

	// Verify selection cleared on mode change
	ui.selectedSlot = 5
	ui.SetMode(ShopModeBuy)
	if ui.selectedSlot != -1 {
		t.Errorf("SetMode() did not clear selection: selectedSlot = %d, want -1", ui.selectedSlot)
	}
}

// TestShopUI_OpenClose tests opening and closing the shop.
func TestShopUI_OpenClose(t *testing.T) {
	world := NewWorld()
	ui := NewShopUI(800, 600)
	merchant := world.CreateEntity()

	// Test initial state
	if ui.IsVisible() {
		t.Error("Shop should not be visible initially")
	}

	// Test Open
	ui.Open(merchant)
	if !ui.IsVisible() {
		t.Error("Shop should be visible after Open()")
	}

	if ui.merchantEntity != merchant {
		t.Error("Open() did not set merchant entity")
	}

	if ui.mode != ShopModeBuy {
		t.Errorf("Open() set mode to %v, want %v", ui.mode, ShopModeBuy)
	}

	if ui.selectedSlot != -1 {
		t.Error("Open() did not reset selectedSlot")
	}

	// Test Close
	ui.selectedSlot = 3
	ui.Close()
	if ui.IsVisible() {
		t.Error("Shop should not be visible after Close()")
	}

	if ui.merchantEntity != nil {
		t.Error("Close() did not clear merchant entity")
	}

	if ui.selectedSlot != -1 {
		t.Error("Close() did not reset selectedSlot")
	}
}

// TestShopUI_Toggle tests toggling visibility.
func TestShopUI_Toggle(t *testing.T) {
	world := NewWorld()
	ui := NewShopUI(800, 600)
	merchant := world.CreateEntity()

	// Open shop first
	ui.Open(merchant)
	if !ui.IsVisible() {
		t.Fatal("Shop should be visible after Open()")
	}

	// Toggle off
	ui.Toggle()
	if ui.IsVisible() {
		t.Error("Shop should not be visible after first Toggle()")
	}

	// Verify state was cleaned up
	if ui.merchantEntity != nil {
		t.Error("Toggle(off) did not clean up merchant entity")
	}

	// Toggle on
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("Shop should be visible after second Toggle()")
	}

	// Note: merchantEntity will still be nil since Toggle doesn't set it,
	// only Open() does. This is expected behavior.
}

// TestShopUI_ModeSwitching tests switching between buy and sell modes.
func TestShopUI_ModeSwitching(t *testing.T) {
	ui := NewShopUI(800, 600)

	// Test initial mode
	if ui.mode != ShopModeBuy {
		t.Errorf("Initial mode = %v, want %v", ui.mode, ShopModeBuy)
	}

	// Set to sell mode
	ui.selectedSlot = 5 // Set selection to verify it gets cleared
	ui.SetMode(ShopModeSell)

	if ui.mode != ShopModeSell {
		t.Errorf("After SetMode(Sell), mode = %v, want %v", ui.mode, ShopModeSell)
	}

	if ui.selectedSlot != -1 {
		t.Error("SetMode() should clear selection")
	}

	// Set back to buy mode
	ui.selectedSlot = 3
	ui.SetMode(ShopModeBuy)

	if ui.mode != ShopModeBuy {
		t.Errorf("After SetMode(Buy), mode = %v, want %v", ui.mode, ShopModeBuy)
	}

	if ui.selectedSlot != -1 {
		t.Error("SetMode() should clear selection")
	}
}

// TestShopUI_ShowMessage tests transaction message display.
func TestShopUI_ShowMessage(t *testing.T) {
	ui := NewShopUI(800, 600)

	// Show a message
	ui.showMessage("Test message")

	if ui.lastTransactionMessage != "Test message" {
		t.Errorf("lastTransactionMessage = %q, want %q", ui.lastTransactionMessage, "Test message")
	}

	if ui.transactionMessageTime != 3.0 {
		t.Errorf("transactionMessageTime = %f, want 3.0", ui.transactionMessageTime)
	}

	// Simulate time passing
	ui.Update(nil, 1.5)
	if ui.transactionMessageTime != 1.5 {
		t.Errorf("After 1.5s, transactionMessageTime = %f, want 1.5", ui.transactionMessageTime)
	}

	// Message should still be visible
	if ui.lastTransactionMessage == "" {
		t.Error("Message cleared too early")
	}

	// Simulate more time passing
	ui.Update(nil, 2.0)
	if ui.transactionMessageTime > 0 {
		t.Error("transactionMessageTime should be 0 after 3+ seconds")
	}

	if ui.lastTransactionMessage != "" {
		t.Error("Message should be cleared after timer expires")
	}
}

// TestShopUI_ExecuteTransaction_Buy tests buying items from merchant.
func TestShopUI_ExecuteTransaction_Buy(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	commerceSystem := NewCommerceSystem(world, invSystem)

	// Create player with gold
	player := world.CreateEntity()
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 200
	player.AddComponent(playerInv)

	// Create merchant with items
	merchant := world.CreateEntity()
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	merchantComp.AddItem(sword)
	merchant.AddComponent(merchantComp)

	// Process entities
	world.Update(0)

	// Create and configure shop UI
	ui := NewShopUI(800, 600)
	ui.SetPlayerEntity(player)
	ui.SetMerchantEntity(merchant)
	ui.SetCommerceSystem(commerceSystem)
	ui.Open(merchant)
	ui.SetMode(ShopModeBuy)
	ui.selectedSlot = 0 // Select first item

	// Execute purchase
	ui.executeTransaction()

	// Verify transaction occurred
	if playerInv.Gold != 50 { // 200 - 150 (100 * 1.5)
		t.Errorf("Player gold = %d, want 50", playerInv.Gold)
	}

	if len(playerInv.Items) != 1 {
		t.Errorf("Player has %d items, want 1", len(playerInv.Items))
	}

	if len(merchantComp.Inventory) != 0 {
		t.Errorf("Merchant has %d items, want 0", len(merchantComp.Inventory))
	}

	// Verify message was shown
	if ui.lastTransactionMessage == "" {
		t.Error("No transaction message shown")
	}
}

// TestShopUI_ExecuteTransaction_Sell tests selling items to merchant.
func TestShopUI_ExecuteTransaction_Sell(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	commerceSystem := NewCommerceSystem(world, invSystem)

	// Create player with item
	player := world.CreateEntity()
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 50
	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	playerInv.AddItem(sword)
	player.AddComponent(playerInv)

	// Create merchant
	merchant := world.CreateEntity()
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)

	// Process entities
	world.Update(0)

	// Create and configure shop UI
	ui := NewShopUI(800, 600)
	ui.SetPlayerEntity(player)
	ui.SetMerchantEntity(merchant)
	ui.SetCommerceSystem(commerceSystem)
	ui.Open(merchant)
	ui.SetMode(ShopModeSell)
	ui.selectedSlot = 0 // Select first item

	// Execute sale
	ui.executeTransaction()

	// Verify transaction occurred
	if playerInv.Gold != 100 { // 50 + 50 (100 * 0.5)
		t.Errorf("Player gold = %d, want 100", playerInv.Gold)
	}

	if len(playerInv.Items) != 0 {
		t.Errorf("Player has %d items, want 0", len(playerInv.Items))
	}

	if len(merchantComp.Inventory) != 1 {
		t.Errorf("Merchant has %d items, want 1", len(merchantComp.Inventory))
	}

	// Verify message was shown
	if ui.lastTransactionMessage == "" {
		t.Error("No transaction message shown")
	}
}

// TestShopUI_ExecuteTransaction_InsufficientGold tests buying with insufficient gold.
func TestShopUI_ExecuteTransaction_InsufficientGold(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	commerceSystem := NewCommerceSystem(world, invSystem)

	// Create player with insufficient gold
	player := world.CreateEntity()
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 50 // Not enough for 150 gold sword
	player.AddComponent(playerInv)

	// Create merchant with expensive item
	merchant := world.CreateEntity()
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	merchantComp.AddItem(sword)
	merchant.AddComponent(merchantComp)

	// Process entities
	world.Update(0)

	// Create and configure shop UI
	ui := NewShopUI(800, 600)
	ui.SetPlayerEntity(player)
	ui.SetMerchantEntity(merchant)
	ui.SetCommerceSystem(commerceSystem)
	ui.Open(merchant)
	ui.SetMode(ShopModeBuy)
	ui.selectedSlot = 0

	// Execute purchase (should fail)
	ui.executeTransaction()

	// Verify transaction failed (nothing changed)
	if playerInv.Gold != 50 {
		t.Errorf("Player gold changed to %d, should remain 50", playerInv.Gold)
	}

	if len(playerInv.Items) != 0 {
		t.Errorf("Player has %d items, want 0", len(playerInv.Items))
	}

	if len(merchantComp.Inventory) != 1 {
		t.Errorf("Merchant has %d items, want 1", len(merchantComp.Inventory))
	}

	// Verify error message was shown
	if ui.lastTransactionMessage == "" {
		t.Error("No error message shown for insufficient gold")
	}
}

// TestShopUI_ExecuteTransaction_NoCommerceSystem tests transaction without commerce system.
func TestShopUI_ExecuteTransaction_NoCommerceSystem(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()
	merchant := world.CreateEntity()

	ui := NewShopUI(800, 600)
	ui.SetPlayerEntity(player)
	ui.SetMerchantEntity(merchant)
	// Deliberately not setting commerce system
	ui.selectedSlot = 0

	// Execute transaction (should fail gracefully)
	ui.executeTransaction()

	// Verify error message was shown
	if ui.lastTransactionMessage == "" {
		t.Error("No error message shown when commerce system missing")
	}
}

// TestShopUI_Draw_NotVisible tests that drawing is skipped when not visible.
func TestShopUI_Draw_NotVisible(t *testing.T) {
	ui := NewShopUI(800, 600)

	// Create screen
	screen := ebiten.NewImage(800, 600)

	// Draw should not panic when not visible
	ui.Draw(screen)

	// Test passes if no panic occurs
}

// TestShopUI_Draw_NoEntities tests drawing with missing entities.
func TestShopUI_Draw_NoEntities(t *testing.T) {
	ui := NewShopUI(800, 600)
	ui.visible = true // Force visible

	// Create screen
	screen := ebiten.NewImage(800, 600)

	// Draw should not panic with missing entities
	ui.Draw(screen)

	// Test passes if no panic occurs
}

// TestShopUI_Draw_WithValidSetup tests drawing with complete setup.
func TestShopUI_Draw_WithValidSetup(t *testing.T) {
	world := NewWorld()

	// Create player with inventory
	player := world.CreateEntity()
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 100
	player.AddComponent(playerInv)

	// Create merchant with items
	merchant := world.CreateEntity()
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchantComp.MerchantName = "Test Merchant"
	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	merchantComp.AddItem(sword)
	merchant.AddComponent(merchantComp)

	// Create UI
	ui := NewShopUI(800, 600)
	ui.SetPlayerEntity(player)
	ui.SetMerchantEntity(merchant)
	ui.Open(merchant)

	// Create screen
	screen := ebiten.NewImage(800, 600)

	// Draw should not panic with valid setup
	ui.Draw(screen)

	// Test passes if no panic occurs
}

// TestShopUI_Update_NotVisible tests that update is skipped when not visible.
func TestShopUI_Update_NotVisible(t *testing.T) {
	ui := NewShopUI(800, 600)

	// Update should not panic when not visible
	ui.Update(nil, 0.016)

	// Test passes if no panic occurs
}

// TestShopUI_Update_NoEntities tests update with missing entities.
func TestShopUI_Update_NoEntities(t *testing.T) {
	ui := NewShopUI(800, 600)
	ui.visible = true // Force visible

	// Update should not panic with missing entities
	ui.Update(nil, 0.016)

	// Test passes if no panic occurs
}
