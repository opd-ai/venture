package economy

import (
	"testing"
	"time"
)

// mockWorld implements the World interface for testing
type mockWorld struct{}

func (m *mockWorld) GetEntities() []Entity {
	return []Entity{}
}

func TestNewSystem(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.marketplace == nil {
		t.Error("System marketplace is nil")
	}

	if sys.world != world {
		t.Error("System world reference not set correctly")
	}
}

func TestNewSystemWithServerID(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystemWithServerID(world, "test-server")

	if sys == nil {
		t.Fatal("NewSystemWithServerID returned nil")
	}

	if sys.marketplace == nil {
		t.Error("System marketplace is nil")
	}
}

func TestSystem_Update(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	// Should not panic
	sys.Update(0.016) // 60 FPS delta time
	sys.Update(1.0)
	sys.Update(60.0) // Trigger cleanup check
}

func TestSystem_CreateListing(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	listing := &Listing{
		ItemID:   "sword_001",
		ItemName: "Test Sword",
		ItemType: "weapon",
		SellerID: "player_001",
		Price:    1000,
		Quantity: 5,
	}

	err := sys.CreateListing(listing)
	if err != nil {
		t.Fatalf("CreateListing failed: %v", err)
	}

	if listing.ListingID == "" {
		t.Error("Listing ID not generated")
	}
}

func TestSystem_SearchItems(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	// Create listings
	for i := 0; i < 5; i++ {
		sys.CreateListing(&Listing{
			ItemID:   "weapon_" + string(rune('0'+i)),
			ItemName: "Sword",
			ItemType: "weapon",
			SellerID: "player_001",
			Price:    1000 + i*100,
			Quantity: 1,
		})
	}

	// Search for weapons
	results, err := sys.SearchItems(ItemQuery{
		ItemType: "weapon",
	})
	if err != nil {
		t.Fatalf("SearchItems failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

func TestSystem_PurchaseItem(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	listing := &Listing{
		ItemID:   "potion_001",
		ItemName: "Health Potion",
		ItemType: "consumable",
		SellerID: "seller_001",
		Price:    50,
		Quantity: 10,
	}

	sys.CreateListing(listing)

	_, err := sys.PurchaseItem(listing.ListingID, "buyer_001", 3)
	if err != nil {
		t.Fatalf("PurchaseItem failed: %v", err)
	}

	// Verify quantity decreased
	results, _ := sys.SearchItems(ItemQuery{ItemType: "consumable"})
	if len(results) > 0 && results[0].Quantity != 7 {
		t.Errorf("Expected 7 remaining, got %d", results[0].Quantity)
	}
}

func TestSystem_GetPriceTrend(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	// Create multiple listings with different prices
	prices := []int{1000, 1200, 800, 1500, 1100}
	for i, price := range prices {
		sys.CreateListing(&Listing{
			ItemID:   "weapon_" + string(rune('0'+i)),
			ItemType: "weapon",
			SellerID: "seller_001",
			Price:    price,
			Quantity: 1,
		})
	}

	trend := sys.GetPriceTrend("weapon")

	if trend == nil {
		t.Fatal("GetPriceTrend returned nil")
	}

	if trend.MinPrice != 800 {
		t.Errorf("Expected min price 800, got %d", trend.MinPrice)
	}

	if trend.MaxPrice != 1500 {
		t.Errorf("Expected max price 1500, got %d", trend.MaxPrice)
	}

	if trend.TotalListings != 5 {
		t.Errorf("Expected 5 listings, got %d", trend.TotalListings)
	}
}

func TestSystem_GetStats(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	// Create some listings
	for i := 0; i < 10; i++ {
		sys.CreateListing(&Listing{
			ItemID:   "item_" + string(rune('0'+i)),
			ItemType: "weapon",
			SellerID: "seller_001",
			Price:    1000,
			Quantity: 1,
		})
	}

	stats := sys.GetStats()

	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	localListings, ok := stats["local_listings"]
	if !ok {
		t.Error("Stats missing local_listings")
	}

	if localListings.(int) != 10 {
		t.Errorf("Expected 10 local listings, got %v", localListings)
	}
}

func TestSystem_GetMarketplace(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	mp := sys.GetMarketplace()

	if mp == nil {
		t.Fatal("GetMarketplace returned nil")
	}

	if mp != sys.marketplace {
		t.Error("GetMarketplace returned different instance")
	}
}

func TestSystem_ConcurrentAccess(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	done := make(chan bool, 10)

	// Concurrent listing creation
	for i := 0; i < 5; i++ {
		go func(idx int) {
			for j := 0; j < 10; j++ {
				sys.CreateListing(&Listing{
					ItemID:   "item_" + string(rune(idx*10+j)),
					ItemType: "weapon",
					SellerID: "seller_001",
					Price:    1000,
					Quantity: 1,
				})
			}
			done <- true
		}(i)
	}

	// Concurrent searches
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_, _ = sys.SearchItems(ItemQuery{ItemType: "weapon"})
				_ = sys.GetStats()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSystem_CleanupInterval(t *testing.T) {
	world := &mockWorld{}
	sys := NewSystem(world)

	if sys.cleanupInterval != 5*time.Minute {
		t.Errorf("Expected 5 minute cleanup interval, got %v", sys.cleanupInterval)
	}
}
