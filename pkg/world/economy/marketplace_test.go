package economy

import (
	"testing"
	"time"
)

func TestNewFederatedMarketplace(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	if marketplace.localServerID != "server1" {
		t.Errorf("Expected server ID 'server1', got '%s'", marketplace.localServerID)
	}
	if marketplace.maxListingsLocal != 10000 {
		t.Errorf("Expected max listings 10000, got %d", marketplace.maxListingsLocal)
	}
	if len(marketplace.localListings) != 0 {
		t.Errorf("Expected empty listings, got %d", len(marketplace.localListings))
	}
}

func TestFederatedMarketplace_CreateListing(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	tests := []struct {
		name    string
		listing *Listing
		wantErr bool
	}{
		{
			name: "Valid listing",
			listing: &Listing{
				ItemID:     "sword123",
				ItemName:   "Iron Sword",
				ItemType:   "weapon",
				SellerID:   "player1",
				SellerName: "TestPlayer",
				Price:      1000,
				Quantity:   1,
			},
			wantErr: false,
		},
		{
			name: "Missing item ID",
			listing: &Listing{
				SellerID: "player1",
				Price:    1000,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "Missing seller ID",
			listing: &Listing{
				ItemID:   "sword123",
				Price:    1000,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "Invalid price (zero)",
			listing: &Listing{
				ItemID:   "sword123",
				SellerID: "player1",
				Price:    0,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "Invalid price (negative)",
			listing: &Listing{
				ItemID:   "sword123",
				SellerID: "player1",
				Price:    -100,
				Quantity: 1,
			},
			wantErr: true,
		},
		{
			name: "Invalid quantity (zero)",
			listing: &Listing{
				ItemID:   "sword123",
				SellerID: "player1",
				Price:    1000,
				Quantity: 0,
			},
			wantErr: true,
		},
		{
			name: "Invalid quantity (negative)",
			listing: &Listing{
				ItemID:   "sword123",
				SellerID: "player1",
				Price:    1000,
				Quantity: -5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := marketplace.CreateListing(tt.listing)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateListing() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify listing was created
				if tt.listing.ListingID == "" {
					t.Error("Expected listing ID to be generated")
				}
				if tt.listing.ServerID != "server1" {
					t.Errorf("Expected server ID 'server1', got '%s'", tt.listing.ServerID)
				}
				if tt.listing.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.listing.ExpiresAt.IsZero() {
					t.Error("Expected ExpiresAt to be set")
				}
			}
		})
	}
}

func TestFederatedMarketplace_GetListing(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	listing := &Listing{
		ItemID:   "sword123",
		SellerID: "player1",
		Price:    1000,
		Quantity: 1,
	}

	err := marketplace.CreateListing(listing)
	if err != nil {
		t.Fatalf("CreateListing() failed: %v", err)
	}

	// Get existing listing
	retrieved, err := marketplace.GetListing(listing.ListingID)
	if err != nil {
		t.Errorf("GetListing() error = %v", err)
	}
	if retrieved.ItemID != "sword123" {
		t.Errorf("Expected ItemID 'sword123', got '%s'", retrieved.ItemID)
	}

	// Get non-existent listing
	_, err = marketplace.GetListing("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent listing")
	}
}

func TestFederatedMarketplace_RemoveListing(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	listing := &Listing{
		ItemID:   "sword123",
		SellerID: "player1",
		Price:    1000,
		Quantity: 1,
	}

	err := marketplace.CreateListing(listing)
	if err != nil {
		t.Fatalf("CreateListing() failed: %v", err)
	}

	// Remove existing listing
	err = marketplace.RemoveListing(listing.ListingID)
	if err != nil {
		t.Errorf("RemoveListing() error = %v", err)
	}

	// Verify removal
	_, err = marketplace.GetListing(listing.ListingID)
	if err == nil {
		t.Error("Expected error for removed listing")
	}

	// Remove non-existent listing
	err = marketplace.RemoveListing("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent listing")
	}
}

func TestFederatedMarketplace_SearchItems(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	// Create test listings
	listings := []*Listing{
		{ItemID: "sword1", ItemName: "Iron Sword", ItemType: "weapon", SellerID: "player1", Price: 1000, Quantity: 1},
		{ItemID: "sword2", ItemName: "Steel Sword", ItemType: "weapon", SellerID: "player2", Price: 2000, Quantity: 1},
		{ItemID: "armor1", ItemName: "Iron Armor", ItemType: "armor", SellerID: "player1", Price: 1500, Quantity: 1},
		{ItemID: "potion1", ItemName: "Health Potion", ItemType: "consumable", SellerID: "player3", Price: 50, Quantity: 10},
	}

	for _, listing := range listings {
		err := marketplace.CreateListing(listing)
		if err != nil {
			t.Fatalf("CreateListing() failed: %v", err)
		}
	}

	tests := []struct {
		name      string
		query     ItemQuery
		wantCount int
	}{
		{
			name:      "Search all items",
			query:     ItemQuery{},
			wantCount: 4,
		},
		{
			name: "Search by type (weapon)",
			query: ItemQuery{
				ItemType: "weapon",
			},
			wantCount: 2,
		},
		{
			name: "Search by type (armor)",
			query: ItemQuery{
				ItemType: "armor",
			},
			wantCount: 1,
		},
		{
			name: "Search by price range",
			query: ItemQuery{
				MinPrice: 1000,
				MaxPrice: 2000,
			},
			wantCount: 3,
		},
		{
			name: "Search by seller",
			query: ItemQuery{
				SellerID: "player1",
			},
			wantCount: 2,
		},
		{
			name: "Search with limit",
			query: ItemQuery{
				Limit: 2,
			},
			wantCount: 2,
		},
		{
			name: "No results",
			query: ItemQuery{
				ItemType: "mount",
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := marketplace.SearchItems(tt.query)
			if err != nil {
				t.Errorf("SearchItems() error = %v", err)
			}
			if len(results) != tt.wantCount {
				t.Errorf("SearchItems() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestFederatedMarketplace_PurchaseItem(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	listing := &Listing{
		ItemID:   "sword123",
		SellerID: "player1",
		Price:    1000,
		Quantity: 5,
	}

	err := marketplace.CreateListing(listing)
	if err != nil {
		t.Fatalf("CreateListing() failed: %v", err)
	}

	// Purchase partial quantity
	_, err = marketplace.PurchaseItem(listing.ListingID, "buyer1", 2)
	if err != nil {
		t.Errorf("PurchaseItem() error = %v", err)
	}

	// Verify quantity updated
	updated, _ := marketplace.GetListing(listing.ListingID)
	if updated.Quantity != 3 {
		t.Errorf("Expected quantity 3, got %d", updated.Quantity)
	}

	// Purchase remaining quantity
	_, err = marketplace.PurchaseItem(listing.ListingID, "buyer2", 3)
	if err != nil {
		t.Errorf("PurchaseItem() error = %v", err)
	}

	// Verify listing removed
	_, err = marketplace.GetListing(listing.ListingID)
	if err == nil {
		t.Error("Expected listing to be removed after purchasing all quantity")
	}

	// Purchase from non-existent listing
	_, err = marketplace.PurchaseItem("nonexistent", "buyer3", 1)
	if err == nil {
		t.Error("Expected error for non-existent listing")
	}

	// Create new listing to test insufficient quantity
	listing2 := &Listing{
		ItemID:   "potion1",
		SellerID: "player2",
		Price:    50,
		Quantity: 2,
	}
	marketplace.CreateListing(listing2)

	_, err = marketplace.PurchaseItem(listing2.ListingID, "buyer4", 5)
	if err == nil {
		t.Error("Expected error for insufficient quantity")
	}
}

func TestFederatedMarketplace_CleanupExpiredListings(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	// Create active listing
	activeListing := &Listing{
		ItemID:   "sword1",
		SellerID: "player1",
		Price:    1000,
		Quantity: 1,
	}
	marketplace.CreateListing(activeListing)

	// Create expired listing
	expiredListing := &Listing{
		ItemID:    "sword2",
		SellerID:  "player2",
		Price:     2000,
		Quantity:  1,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	marketplace.CreateListing(expiredListing)

	// Cleanup
	removed := marketplace.CleanupExpiredListings()

	if removed != 1 {
		t.Errorf("Expected 1 expired listing removed, got %d", removed)
	}

	// Verify active listing still exists
	_, err := marketplace.GetListing(activeListing.ListingID)
	if err != nil {
		t.Error("Active listing should not be removed")
	}

	// Verify expired listing removed
	_, err = marketplace.GetListing(expiredListing.ListingID)
	if err == nil {
		t.Error("Expired listing should be removed")
	}
}

func TestFederatedMarketplace_UpdateRemoteCache(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	remoteListings := []*Listing{
		{
			ItemID:    "remote1",
			ServerID:  "server2",
			Price:     500,
			Quantity:  1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Set expiration
		},
		{
			ItemID:    "remote2",
			ServerID:  "server2",
			Price:     600,
			Quantity:  1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Set expiration
		},
	}

	marketplace.UpdateRemoteCache("server2", remoteListings)

	// Verify remote cache
	if len(marketplace.remoteCache) != 1 {
		t.Errorf("Expected 1 remote server cached, got %d", len(marketplace.remoteCache))
	}

	// Verify hops are set
	for _, listing := range remoteListings {
		if listing.EstimatedHops == 0 {
			t.Error("Expected EstimatedHops to be set for remote listings")
		}
	}

	// Search should include remote listings
	query := ItemQuery{}
	results, _ := marketplace.SearchItems(query)
	if len(results) != 2 {
		t.Errorf("Expected 2 results (remote listings), got %d", len(results))
	}
}

func TestFederatedMarketplace_GetStats(t *testing.T) {
	marketplace := NewFederatedMarketplace("server1")

	// Create local listing
	marketplace.CreateListing(&Listing{
		ItemID:   "sword1",
		SellerID: "player1",
		Price:    1000,
		Quantity: 1,
	})

	// Add remote cache
	marketplace.UpdateRemoteCache("server2", []*Listing{
		{ItemID: "remote1", Price: 500, Quantity: 1},
	})

	stats := marketplace.GetStats()

	if stats["local_listings"].(int) != 1 {
		t.Errorf("Expected 1 local listing, got %v", stats["local_listings"])
	}
	if stats["cached_servers"].(int) != 1 {
		t.Errorf("Expected 1 cached server, got %v", stats["cached_servers"])
	}
	if stats["total_cached"].(int) != 1 {
		t.Errorf("Expected 1 total cached listing, got %v", stats["total_cached"])
	}
}

func BenchmarkFederatedMarketplace_CreateListing(b *testing.B) {
	marketplace := NewFederatedMarketplace("server1")
	listing := &Listing{
		ItemID:   "bench_item",
		SellerID: "player1",
		Price:    1000,
		Quantity: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listing.ListingID = "" // Reset to test ID generation
		marketplace.CreateListing(listing)
	}
}

func BenchmarkFederatedMarketplace_SearchItems(b *testing.B) {
	marketplace := NewFederatedMarketplace("server1")

	// Create 100 listings
	for i := 0; i < 100; i++ {
		marketplace.CreateListing(&Listing{
			ItemID:   "item_" + string(rune(i)),
			SellerID: "player1",
			Price:    1000 + i*10,
			Quantity: 1,
			ItemType: "weapon",
		})
	}

	query := ItemQuery{
		ItemType: "weapon",
		MinPrice: 1000,
		MaxPrice: 2000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		marketplace.SearchItems(query)
	}
}

func BenchmarkFederatedMarketplace_PurchaseItem(b *testing.B) {
	marketplace := NewFederatedMarketplace("server1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		listing := &Listing{
			ItemID:   "bench_item",
			SellerID: "player1",
			Price:    1000,
			Quantity: 1,
		}
		marketplace.CreateListing(listing)
		b.StartTimer()

		marketplace.PurchaseItem(listing.ListingID, "buyer1", 1)
	}
}
