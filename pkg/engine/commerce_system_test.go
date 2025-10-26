package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestTransactionType_String(t *testing.T) {
	tests := []struct {
		name string
		tt   TransactionType
		want string
	}{
		{"buy", TransactionBuy, "buy"},
		{"sell", TransactionSell, "sell"},
		{"unknown", TransactionType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tt.String(); got != tt.want {
				t.Errorf("TransactionType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCommerceSystem(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	if system == nil {
		t.Fatal("NewCommerceSystem returned nil")
	}
	if system.world != world {
		t.Error("CommerceSystem world not set correctly")
	}
	if system.inventory != invSystem {
		t.Error("CommerceSystem inventory system not set correctly")
	}
	if system.validator == nil {
		t.Error("CommerceSystem validator is nil, should have default")
	}
}

func TestCommerceSystem_SetValidator(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	customValidator := NewDefaultTransactionValidator()
	system.SetValidator(customValidator)

	if system.validator != customValidator {
		t.Error("SetValidator did not update validator")
	}
}

func TestCommerceSystem_BuyItem(t *testing.T) {
	tests := []struct {
		name          string
		playerGold    int
		itemIndex     int
		wantSuccess   bool
		wantGoldDelta int
		wantErrMsg    string
	}{
		{"buy sword success", 200, 0, true, -150, ""},
		{"not enough gold", 100, 0, false, 0, "Not enough gold"},
		{"buy potion success", 100, 1, true, -75, ""},
		{"invalid index negative", 200, -1, false, 0, "Invalid item index"},
		{"invalid index too high", 200, 99, false, 0, "Invalid item index"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh world and system for each test
			world := NewWorld()
			invSystem := NewInventorySystem(world)
			system := NewCommerceSystem(world, invSystem)

			// Create player
			player := world.CreateEntity()
			playerID := player.ID
			playerInv := NewInventoryComponent(10, 100.0)
			playerInv.Gold = tt.playerGold
			player.AddComponent(playerInv)

			// Create merchant
			merchant := world.CreateEntity()
			merchantID := merchant.ID
			merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
			merchant.AddComponent(merchantComp)

			// Add items to merchant
			sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
			potion := &item.Item{Name: "Potion", Stats: item.Stats{Value: 50}}
			merchantComp.AddItem(sword)
			merchantComp.AddItem(potion)

			// Process entities
			world.Update(0)

			initialGold := playerInv.Gold
			initialPlayerItems := len(playerInv.Items)
			initialMerchantItems := len(merchantComp.Inventory)

			result, err := system.BuyItem(playerID, merchantID, tt.itemIndex)
			if err != nil {
				t.Fatalf("BuyItem() error = %v, want nil", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("BuyItem() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if tt.wantSuccess {
				// Verify gold changed
				if playerInv.Gold != initialGold+tt.wantGoldDelta {
					t.Errorf("Player gold = %d, want %d", playerInv.Gold, initialGold+tt.wantGoldDelta)
				}
				// Verify item added to player
				if len(playerInv.Items) != initialPlayerItems+1 {
					t.Errorf("Player items = %d, want %d", len(playerInv.Items), initialPlayerItems+1)
				}
				// Verify item removed from merchant
				if len(merchantComp.Inventory) != initialMerchantItems-1 {
					t.Errorf("Merchant items = %d, want %d", len(merchantComp.Inventory), initialMerchantItems-1)
				}
			} else {
				// Verify nothing changed on failure
				if playerInv.Gold != initialGold {
					t.Errorf("Player gold changed on failure: %d, want %d", playerInv.Gold, initialGold)
				}
				if len(playerInv.Items) != initialPlayerItems {
					t.Errorf("Player items changed on failure: %d, want %d", len(playerInv.Items), initialPlayerItems)
				}
				if tt.wantErrMsg != "" && result.ErrorMessage == "" {
					t.Error("Expected error message but got none")
				}
			}
		})
	}
}

func TestCommerceSystem_BuyItem_InventoryFull(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	// Create player with full inventory
	player := world.CreateEntity()
	playerID := player.ID
	playerInv := NewInventoryComponent(2, 100.0)
	playerInv.Gold = 500
	player.AddComponent(playerInv)

	// Fill player inventory
	playerInv.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Value: 10}})
	playerInv.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Value: 10}})

	// Create merchant
	merchant := world.CreateEntity()
	merchantID := merchant.ID
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)
	merchantComp.AddItem(&item.Item{Name: "Sword", Stats: item.Stats{Value: 100}})

	// Process entities
	world.Update(0)

	result, err := system.BuyItem(playerID, merchantID, 0)
	if err != nil {
		t.Fatalf("BuyItem() error = %v, want nil", err)
	}

	if result.Success {
		t.Error("BuyItem() succeeded with full inventory, want failure")
	}
	if result.ErrorMessage == "" {
		t.Error("Expected error message for full inventory")
	}
}

func TestCommerceSystem_SellItem(t *testing.T) {
	tests := []struct {
		name          string
		itemIndex     int
		wantSuccess   bool
		wantGoldDelta int
		wantErrMsg    string
	}{
		{"sell sword success", 0, true, 50, ""},  // 100 * 0.5 = 50
		{"sell potion success", 1, true, 25, ""}, // 50 * 0.5 = 25
		{"invalid index negative", -1, false, 0, "Invalid item index"},
		{"invalid index too high", 99, false, 0, "Invalid item index"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh world and system for each test
			world := NewWorld()
			invSystem := NewInventorySystem(world)
			system := NewCommerceSystem(world, invSystem)

			// Create player
			player := world.CreateEntity()
			playerID := player.ID
			playerInv := NewInventoryComponent(10, 100.0)
			playerInv.Gold = 50
			player.AddComponent(playerInv)

			// Add items to player
			sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
			potion := &item.Item{Name: "Potion", Stats: item.Stats{Value: 50}}
			playerInv.AddItem(sword)
			playerInv.AddItem(potion)

			// Create merchant
			merchant := world.CreateEntity()
			merchantID := merchant.ID
			merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
			merchant.AddComponent(merchantComp)

			// Process entities
			world.Update(0)

			initialGold := playerInv.Gold
			initialPlayerItems := len(playerInv.Items)
			initialMerchantItems := len(merchantComp.Inventory)

			result, err := system.SellItem(playerID, merchantID, tt.itemIndex)
			if err != nil {
				t.Fatalf("SellItem() error = %v, want nil", err)
			}

			if result.Success != tt.wantSuccess {
				t.Errorf("SellItem() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if tt.wantSuccess {
				// Verify gold changed
				if playerInv.Gold != initialGold+tt.wantGoldDelta {
					t.Errorf("Player gold = %d, want %d", playerInv.Gold, initialGold+tt.wantGoldDelta)
				}
				// Verify item removed from player
				if len(playerInv.Items) != initialPlayerItems-1 {
					t.Errorf("Player items = %d, want %d", len(playerInv.Items), initialPlayerItems-1)
				}
				// Verify item added to merchant
				if len(merchantComp.Inventory) != initialMerchantItems+1 {
					t.Errorf("Merchant items = %d, want %d", len(merchantComp.Inventory), initialMerchantItems+1)
				}
			} else {
				// Verify nothing changed on failure
				if playerInv.Gold != initialGold {
					t.Errorf("Player gold changed on failure: %d, want %d", playerInv.Gold, initialGold)
				}
				if len(playerInv.Items) != initialPlayerItems {
					t.Errorf("Player items changed on failure: %d, want %d", len(playerInv.Items), initialPlayerItems)
				}
			}
		})
	}
}

func TestCommerceSystem_SellItem_MerchantInventoryFull(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	// Create player
	player := world.CreateEntity()
	playerID := player.ID
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 50
	player.AddComponent(playerInv)
	playerInv.AddItem(&item.Item{Name: "Sword", Stats: item.Stats{Value: 100}})

	// Create merchant with full inventory
	merchant := world.CreateEntity()
	merchantID := merchant.ID
	merchantComp := NewMerchantComponent(2, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)
	merchantComp.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Value: 10}})
	merchantComp.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Value: 10}})

	// Process entities
	world.Update(0)

	result, err := system.SellItem(playerID, merchantID, 0)
	if err != nil {
		t.Fatalf("SellItem() error = %v, want nil", err)
	}

	if result.Success {
		t.Error("SellItem() succeeded with full merchant inventory, want failure")
	}
	if result.ErrorMessage == "" {
		t.Error("Expected error message for full merchant inventory")
	}
}

func TestCommerceSystem_BuyItem_InvalidEntities(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	player := world.CreateEntity()
	playerID := player.ID
	playerInv := NewInventoryComponent(10, 100.0)
	playerInv.Gold = 200
	player.AddComponent(playerInv)

	world.Update(0)

	tests := []struct {
		name       string
		playerID   uint64
		merchantID uint64
		wantErr    bool
	}{
		{"invalid player", 9999, 1, true},
		{"invalid merchant", playerID, 9999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := system.BuyItem(tt.playerID, tt.merchantID, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuyItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommerceSystem_SellItem_InvalidEntities(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	player := world.CreateEntity()
	playerID := player.ID
	playerInv := NewInventoryComponent(10, 100.0)
	player.AddComponent(playerInv)
	playerInv.AddItem(&item.Item{Name: "Sword", Stats: item.Stats{Value: 100}})

	world.Update(0)

	tests := []struct {
		name       string
		playerID   uint64
		merchantID uint64
		wantErr    bool
	}{
		{"invalid player", 9999, 1, true},
		{"invalid merchant", playerID, 9999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := system.SellItem(tt.playerID, tt.merchantID, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("SellItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommerceSystem_GetMerchantInventory(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	merchant := world.CreateEntity()
	merchantID := merchant.ID
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)

	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	potion := &item.Item{Name: "Potion", Stats: item.Stats{Value: 50}}
	merchantComp.AddItem(sword)
	merchantComp.AddItem(potion)

	world.Update(0)

	inventory, err := system.GetMerchantInventory(merchantID)
	if err != nil {
		t.Fatalf("GetMerchantInventory() error = %v, want nil", err)
	}

	if len(inventory) != 2 {
		t.Errorf("GetMerchantInventory() length = %d, want 2", len(inventory))
	}

	// Verify it's a copy (modifying shouldn't affect merchant)
	inventory[0] = nil
	if merchantComp.Inventory[0] == nil {
		t.Error("GetMerchantInventory() did not return a copy")
	}
}

func TestCommerceSystem_GetMerchantInventory_InvalidMerchant(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	_, err := system.GetMerchantInventory(9999)
	if err == nil {
		t.Error("GetMerchantInventory() with invalid merchant should return error")
	}
}

func TestCommerceSystem_GetMerchantPrices(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	merchant := world.CreateEntity()
	merchantID := merchant.ID
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)

	world.Update(0)

	itm := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}

	sellPrice, buyPrice, err := system.GetMerchantPrices(merchantID, itm)
	if err != nil {
		t.Fatalf("GetMerchantPrices() error = %v, want nil", err)
	}

	wantSellPrice := 150 // 100 * 1.5
	wantBuyPrice := 50   // 100 * 0.5

	if sellPrice != wantSellPrice {
		t.Errorf("sellPrice = %d, want %d", sellPrice, wantSellPrice)
	}
	if buyPrice != wantBuyPrice {
		t.Errorf("buyPrice = %d, want %d", buyPrice, wantBuyPrice)
	}
}

func TestCommerceSystem_GetMerchantPrices_InvalidMerchant(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	itm := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	_, _, err := system.GetMerchantPrices(9999, itm)
	if err == nil {
		t.Error("GetMerchantPrices() with invalid merchant should return error")
	}
}

func TestCommerceSystem_BuyItem_Rollback(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	system := NewCommerceSystem(world, invSystem)

	// Create player with limited inventory space
	player := world.CreateEntity()
	playerID := player.ID
	playerInv := NewInventoryComponent(1, 100.0)
	playerInv.Gold = 500
	player.AddComponent(playerInv)

	// Create merchant
	merchant := world.CreateEntity()
	merchantID := merchant.ID
	merchantComp := NewMerchantComponent(10, MerchantFixed, 1.5)
	merchant.AddComponent(merchantComp)
	sword := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100, Weight: 200.0}}
	merchantComp.AddItem(sword)

	world.Update(0)

	// This should fail weight check and rollback
	result, err := system.BuyItem(playerID, merchantID, 0)
	if err != nil {
		t.Fatalf("BuyItem() error = %v, want nil", err)
	}

	if result.Success {
		t.Error("BuyItem() should fail due to weight limit")
	}

	// Verify rollback: item should still be in merchant inventory
	if len(merchantComp.Inventory) != 1 {
		t.Errorf("Merchant inventory = %d items, want 1 (rollback should have restored item)", len(merchantComp.Inventory))
	}

	// Verify rollback: player gold should be unchanged
	if playerInv.Gold != 500 {
		t.Errorf("Player gold = %d, want 500 (rollback should have refunded)", playerInv.Gold)
	}
}
