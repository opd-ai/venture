package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestMerchantType_String(t *testing.T) {
	tests := []struct {
		name string
		mt   MerchantType
		want string
	}{
		{"fixed merchant", MerchantFixed, "fixed"},
		{"nomadic merchant", MerchantNomadic, "nomadic"},
		{"unknown merchant", MerchantType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mt.String(); got != tt.want {
				t.Errorf("MerchantType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDialogAction_String(t *testing.T) {
	tests := []struct {
		name string
		da   DialogAction
		want string
	}{
		{"none", ActionNone, "none"},
		{"open shop", ActionOpenShop, "open_shop"},
		{"close dialog", ActionCloseDialog, "close_dialog"},
		{"start quest", ActionStartQuest, "start_quest"},
		{"give item", ActionGiveItem, "give_item"},
		{"unknown", DialogAction(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.da.String(); got != tt.want {
				t.Errorf("DialogAction.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMerchantComponent(t *testing.T) {
	tests := []struct {
		name            string
		maxInventory    int
		merchantType    MerchantType
		priceMultiplier float64
		wantMaxInv      int
		wantMultiplier  float64
	}{
		{
			name:            "valid defaults",
			maxInventory:    20,
			merchantType:    MerchantFixed,
			priceMultiplier: 1.5,
			wantMaxInv:      20,
			wantMultiplier:  1.5,
		},
		{
			name:            "zero inventory uses default",
			maxInventory:    0,
			merchantType:    MerchantNomadic,
			priceMultiplier: 2.0,
			wantMaxInv:      20,
			wantMultiplier:  2.0,
		},
		{
			name:            "negative inventory uses default",
			maxInventory:    -5,
			merchantType:    MerchantFixed,
			priceMultiplier: 1.2,
			wantMaxInv:      20,
			wantMultiplier:  1.2,
		},
		{
			name:            "zero multiplier uses default",
			maxInventory:    30,
			merchantType:    MerchantFixed,
			priceMultiplier: 0,
			wantMaxInv:      30,
			wantMultiplier:  1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant := NewMerchantComponent(tt.maxInventory, tt.merchantType, tt.priceMultiplier)

			if merchant.MaxInventory != tt.wantMaxInv {
				t.Errorf("MaxInventory = %v, want %v", merchant.MaxInventory, tt.wantMaxInv)
			}
			if merchant.PriceMultiplier != tt.wantMultiplier {
				t.Errorf("PriceMultiplier = %v, want %v", merchant.PriceMultiplier, tt.wantMultiplier)
			}
			if merchant.MerchantType != tt.merchantType {
				t.Errorf("MerchantType = %v, want %v", merchant.MerchantType, tt.merchantType)
			}
			if merchant.BuyBackPercentage != 0.5 {
				t.Errorf("BuyBackPercentage = %v, want 0.5", merchant.BuyBackPercentage)
			}
			if merchant.RestockTimeSec != 300 {
				t.Errorf("RestockTimeSec = %v, want 300", merchant.RestockTimeSec)
			}
		})
	}
}

func TestMerchantComponent_Type(t *testing.T) {
	merchant := NewMerchantComponent(20, MerchantFixed, 1.5)
	if got := merchant.Type(); got != "merchant" {
		t.Errorf("Type() = %v, want merchant", got)
	}
}

func TestMerchantComponent_GetSellPrice(t *testing.T) {
	merchant := NewMerchantComponent(20, MerchantFixed, 1.5)

	tests := []struct {
		name      string
		itemValue int
		want      int
	}{
		{"low value item", 10, 15},
		{"medium value item", 100, 150},
		{"high value item", 1000, 1500},
		{"zero value item", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itm := &item.Item{
				Stats: item.Stats{Value: tt.itemValue},
			}
			if got := merchant.GetSellPrice(itm); got != tt.want {
				t.Errorf("GetSellPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerchantComponent_GetBuyPrice(t *testing.T) {
	merchant := NewMerchantComponent(20, MerchantFixed, 1.5)

	tests := []struct {
		name      string
		itemValue int
		want      int
	}{
		{"low value item", 10, 5},
		{"medium value item", 100, 50},
		{"high value item", 1000, 500},
		{"zero value item", 0, 0},
		{"odd value item", 11, 5}, // 11 * 0.5 = 5.5, truncated to 5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itm := &item.Item{
				Stats: item.Stats{Value: tt.itemValue},
			}
			if got := merchant.GetBuyPrice(itm); got != tt.want {
				t.Errorf("GetBuyPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerchantComponent_InventoryManagement(t *testing.T) {
	merchant := NewMerchantComponent(3, MerchantFixed, 1.5)

	// Create test items
	item1 := &item.Item{Name: "Sword", Stats: item.Stats{Value: 100}}
	item2 := &item.Item{Name: "Potion", Stats: item.Stats{Value: 50}}
	item3 := &item.Item{Name: "Armor", Stats: item.Stats{Value: 200}}
	item4 := &item.Item{Name: "Shield", Stats: item.Stats{Value: 150}}

	t.Run("add items", func(t *testing.T) {
		if !merchant.CanAddItem() {
			t.Error("CanAddItem() = false, want true for empty inventory")
		}

		if !merchant.AddItem(item1) {
			t.Error("AddItem(item1) = false, want true")
		}
		if len(merchant.Inventory) != 1 {
			t.Errorf("Inventory length = %d, want 1", len(merchant.Inventory))
		}

		merchant.AddItem(item2)
		merchant.AddItem(item3)

		if len(merchant.Inventory) != 3 {
			t.Errorf("Inventory length = %d, want 3", len(merchant.Inventory))
		}
	})

	t.Run("inventory full", func(t *testing.T) {
		if merchant.CanAddItem() {
			t.Error("CanAddItem() = true, want false for full inventory")
		}

		if merchant.AddItem(item4) {
			t.Error("AddItem() = true, want false for full inventory")
		}
	})

	t.Run("remove items", func(t *testing.T) {
		removed := merchant.RemoveItem(1) // Remove item2
		if removed != item2 {
			t.Errorf("RemoveItem(1) = %v, want %v", removed, item2)
		}
		if len(merchant.Inventory) != 2 {
			t.Errorf("Inventory length = %d, want 2", len(merchant.Inventory))
		}

		// Can add now that there's space
		if !merchant.CanAddItem() {
			t.Error("CanAddItem() = false, want true after removal")
		}
	})

	t.Run("remove invalid index", func(t *testing.T) {
		if got := merchant.RemoveItem(-1); got != nil {
			t.Errorf("RemoveItem(-1) = %v, want nil", got)
		}
		if got := merchant.RemoveItem(10); got != nil {
			t.Errorf("RemoveItem(10) = %v, want nil", got)
		}
	})
}

func TestMerchantComponent_NeedsRestock(t *testing.T) {
	tests := []struct {
		name        string
		restockTime float64
		lastRestock float64
		currentTime float64
		want        bool
	}{
		{"needs restock", 300, 0, 301, true},
		{"does not need restock", 300, 0, 299, false},
		{"exactly at restock time", 300, 0, 300, true},
		{"no restocking enabled", 0, 0, 1000, false},
		{"negative restock time", -100, 0, 500, false},
		{"multiple restock periods", 60, 100, 250, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merchant := NewMerchantComponent(20, MerchantFixed, 1.5)
			merchant.RestockTimeSec = tt.restockTime
			merchant.LastRestockTime = tt.lastRestock

			if got := merchant.NeedsRestock(tt.currentTime); got != tt.want {
				t.Errorf("NeedsRestock(%v) = %v, want %v", tt.currentTime, got, tt.want)
			}
		})
	}
}

func TestDialogComponent_Type(t *testing.T) {
	dialog := NewDialogComponent(nil)
	if got := dialog.Type(); got != "dialog" {
		t.Errorf("Type() = %v, want dialog", got)
	}
}

func TestDialogComponent_ActivateDeactivate(t *testing.T) {
	provider := NewMerchantDialogProvider("Test Merchant")
	dialog := NewDialogComponent(provider)

	t.Run("initial state", func(t *testing.T) {
		if dialog.IsActive {
			t.Error("IsActive = true, want false initially")
		}
	})

	t.Run("activate", func(t *testing.T) {
		dialog.Activate()

		if !dialog.IsActive {
			t.Error("IsActive = false, want true after activate")
		}
		if dialog.CurrentDialog == "" {
			t.Error("CurrentDialog is empty after activate")
		}
		if len(dialog.Options) == 0 {
			t.Error("Options is empty after activate")
		}
	})

	t.Run("deactivate", func(t *testing.T) {
		dialog.Deactivate()

		if dialog.IsActive {
			t.Error("IsActive = true, want false after deactivate")
		}
		if dialog.CurrentDialog != "" {
			t.Errorf("CurrentDialog = %q, want empty after deactivate", dialog.CurrentDialog)
		}
		if len(dialog.Options) != 0 {
			t.Errorf("Options length = %d, want 0 after deactivate", len(dialog.Options))
		}
	})

	t.Run("activate without provider", func(t *testing.T) {
		dialogNoProvider := NewDialogComponent(nil)
		dialogNoProvider.Activate()

		if !dialogNoProvider.IsActive {
			t.Error("IsActive = false, want true even without provider")
		}
	})
}

func TestMerchantDialogProvider_GetDialog(t *testing.T) {
	provider := NewMerchantDialogProvider("Elara the Merchant")

	text, options := provider.GetDialog()

	if text == "" {
		t.Error("GetDialog() text is empty")
	}

	if len(options) != 2 {
		t.Errorf("GetDialog() returned %d options, want 2", len(options))
	}

	// Verify shop option exists
	foundShop := false
	foundLeave := false
	for _, opt := range options {
		if opt.Action == ActionOpenShop {
			foundShop = true
			if !opt.Enabled {
				t.Error("Shop option is disabled, want enabled")
			}
		}
		if opt.Action == ActionCloseDialog {
			foundLeave = true
			if !opt.Enabled {
				t.Error("Leave option is disabled, want enabled")
			}
		}
	}

	if !foundShop {
		t.Error("Shop option not found in dialog options")
	}
	if !foundLeave {
		t.Error("Leave option not found in dialog options")
	}
}

func TestDefaultTransactionValidator_CanBuyItem(t *testing.T) {
	validator := NewDefaultTransactionValidator()

	tests := []struct {
		name            string
		playerGold      int
		itemPrice       int
		inventoryFull   bool
		wantValid       bool
		wantErrContains string
	}{
		{"valid purchase", 100, 50, false, true, ""},
		{"exact gold", 50, 50, false, true, ""},
		{"not enough gold", 25, 50, false, false, "Not enough gold"},
		{"inventory full", 100, 50, true, false, "Inventory full"},
		{"not enough gold and inventory full", 25, 50, true, false, "Not enough gold"},
		{"zero gold but zero price", 0, 0, false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := validator.CanBuyItem(tt.playerGold, tt.itemPrice, tt.inventoryFull)

			if valid != tt.wantValid {
				t.Errorf("CanBuyItem() valid = %v, want %v", valid, tt.wantValid)
			}

			if tt.wantErrContains != "" {
				if errMsg == "" {
					t.Errorf("CanBuyItem() errMsg is empty, want containing %q", tt.wantErrContains)
				}
			} else {
				if errMsg != "" {
					t.Errorf("CanBuyItem() errMsg = %q, want empty", errMsg)
				}
			}
		})
	}
}

func TestDefaultTransactionValidator_CanSellItem(t *testing.T) {
	validator := NewDefaultTransactionValidator()

	tests := []struct {
		name            string
		merchantGold    int
		itemPrice       int
		inventoryFull   bool
		wantValid       bool
		wantErrContains string
	}{
		{"valid sale", 1000, 50, false, true, ""},
		{"merchant inventory full", 1000, 50, true, false, "Merchant inventory full"},
		{"merchant has no gold but can still buy", 0, 50, false, true, ""},
		{"merchant has insufficient gold but can still buy", 25, 50, false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := validator.CanSellItem(tt.merchantGold, tt.itemPrice, tt.inventoryFull)

			if valid != tt.wantValid {
				t.Errorf("CanSellItem() valid = %v, want %v", valid, tt.wantValid)
			}

			if tt.wantErrContains != "" {
				if errMsg == "" {
					t.Errorf("CanSellItem() errMsg is empty, want containing %q", tt.wantErrContains)
				}
			} else {
				if errMsg != "" {
					t.Errorf("CanSellItem() errMsg = %q, want empty", errMsg)
				}
			}
		})
	}
}
