package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/world/economy"
)

func TestNewEconomySystem(t *testing.T) {
	world := NewWorld()
	serverID := "test-server-1"

	sys := NewEconomySystem(world, serverID)

	if sys == nil {
		t.Fatal("expected non-nil economy system")
	}
	if sys.world != world {
		t.Error("world not set correctly")
	}
	if sys.serverID != serverID {
		t.Error("serverID not set correctly")
	}
	if sys.marketplace == nil {
		t.Error("marketplace not initialized")
	}
	if sys.guildBank == nil {
		t.Error("guild bank not initialized")
	}
}

func TestEconomySystemUpdate(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	// Create test entity (not used by system, but required by interface)
	entity := world.CreateEntity()
	entities := []*Entity{entity}

	// Should not panic
	sys.Update(entities, 0.016)

	// Verify accumulators track deltaTime
	if sys.updateAccumulator < 0.016 {
		t.Error("update accumulator should have increased")
	}

	// Force marketplace update by accumulating enough time
	sys.updateAccumulator = sys.updateInterval + 1.0
	sys.interestAccumulator = sys.interestInterval + 1.0

	// Update should trigger marketplace update and interest calculation
	sys.Update(entities, 0.016)

	// Accumulators should have been reduced by subtracting interval
	if sys.updateAccumulator >= sys.updateInterval {
		t.Error("update accumulator should have been reduced after triggering")
	}
	if sys.interestAccumulator >= sys.interestInterval {
		t.Error("interest accumulator should have been reduced after triggering")
	}
}

func TestEconomySystemCreateListing(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	tests := []struct {
		name    string
		listing *economy.Listing
		wantErr bool
	}{
		{
			name: "valid listing",
			listing: &economy.Listing{
				ItemID:     "sword-123",
				ItemName:   "Iron Sword",
				ItemType:   "weapon",
				SellerID:   "player-1",
				SellerName: "TestPlayer",
				Price:      100,
				Quantity:   1,
			},
			wantErr: false,
		},
		{
			name: "missing item ID",
			listing: &economy.Listing{
				SellerID: "player-1",
				Price:    100,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "missing seller ID",
			listing: &economy.Listing{
				ItemID:   "sword-123",
				Price:    100,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			listing: &economy.Listing{
				ItemID:   "sword-123",
				SellerID: "player-1",
				Price:    0,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid quantity",
			listing: &economy.Listing{
				ItemID:   "sword-123",
				SellerID: "player-1",
				Price:    100,
				Quantity: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.CreateListing(tt.listing)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateListing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEconomySystemPurchaseItem(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	// Create a listing first
	listing := &economy.Listing{
		ItemID:     "potion-1",
		ItemName:   "Health Potion",
		ItemType:   "consumable",
		SellerID:   "player-1",
		SellerName: "Seller",
		Price:      50,
		Quantity:   5,
	}

	err := sys.CreateListing(listing)
	if err != nil {
		t.Fatalf("failed to create listing: %v", err)
	}

	// Purchase item
	err = sys.PurchaseItem(listing.ListingID, "player-2", 2)
	if err != nil {
		t.Errorf("PurchaseItem() failed: %v", err)
	}

	// Try to purchase with invalid listing ID
	err = sys.PurchaseItem("invalid-id", "player-2", 1)
	if err == nil {
		t.Error("expected error for invalid listing ID")
	}
}

func TestEconomySystemCreateGuildVault(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	tests := []struct {
		name         string
		guildID      string
		interestRate float64
		wantErr      bool
	}{
		{
			name:         "valid vault",
			guildID:      "guild-1",
			interestRate: 0.005,
			wantErr:      false,
		},
		{
			name:         "duplicate vault",
			guildID:      "guild-1",
			interestRate: 0.005,
			wantErr:      true,
		},
		{
			name:         "invalid interest rate - too low",
			guildID:      "guild-2",
			interestRate: 0.0005,
			wantErr:      true,
		},
		{
			name:         "invalid interest rate - too high",
			guildID:      "guild-3",
			interestRate: 0.015,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.CreateGuildVault(tt.guildID, tt.interestRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGuildVault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEconomySystemDepositGold(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	// Create vault first
	guildID := "test-guild"
	err := sys.CreateGuildVault(guildID, 0.005)
	if err != nil {
		t.Fatalf("failed to create guild vault: %v", err)
	}

	// Deposit gold
	err = sys.DepositGold(guildID, "player-1", "TestPlayer", 1000)
	if err != nil {
		t.Errorf("DepositGold() failed: %v", err)
	}

	// Try to deposit to non-existent vault
	err = sys.DepositGold("invalid-guild", "player-1", "TestPlayer", 1000)
	if err == nil {
		t.Error("expected error for invalid guild ID")
	}
}

func TestEconomySystemWithdrawGold(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	// Create vault and deposit gold
	guildID := "test-guild"
	err := sys.CreateGuildVault(guildID, 0.005)
	if err != nil {
		t.Fatalf("failed to create guild vault: %v", err)
	}

	// Set withdrawal limits via direct access (system doesn't expose this)
	sys.GetGuildBank().SetWithdrawalLimit(guildID, "member", 500)

	err = sys.DepositGold(guildID, "player-1", "Player1", 1000)
	if err != nil {
		t.Fatalf("failed to deposit gold: %v", err)
	}

	// Withdraw within limit
	err = sys.WithdrawGold(guildID, "player-2", "Player2", "member", 300)
	if err != nil {
		t.Errorf("WithdrawGold() failed: %v", err)
	}

	// Try to withdraw from non-existent vault
	err = sys.WithdrawGold("invalid-guild", "player-2", "Player2", "member", 100)
	if err == nil {
		t.Error("expected error for invalid guild ID")
	}
}

func TestEconomySystemGetters(t *testing.T) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	marketplace := sys.GetMarketplace()
	if marketplace == nil {
		t.Error("GetMarketplace() returned nil")
	}

	guildBank := sys.GetGuildBank()
	if guildBank == nil {
		t.Error("GetGuildBank() returned nil")
	}
}

func BenchmarkEconomySystemUpdate(b *testing.B) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")
	entity := world.CreateEntity()
	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkCreateListing(b *testing.B) {
	world := NewWorld()
	sys := NewEconomySystem(world, "test-server")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listing := &economy.Listing{
			ItemID:     "item-" + string(rune(i)),
			ItemName:   "Test Item",
			ItemType:   "weapon",
			SellerID:   "seller-1",
			SellerName: "Seller",
			Price:      100,
			Quantity:   1,
		}
		sys.CreateListing(listing)
	}
}
