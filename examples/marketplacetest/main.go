package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/world/economy"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, create, search, purchase, trends, stats, all")
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo()
	case "create":
		testCreateListings()
	case "search":
		testSearchItems()
	case "purchase":
		testPurchaseItems()
	case "trends":
		testPriceTrends()
	case "stats":
		testMarketplaceStats()
	case "all":
		runDemo()
		testCreateListings()
		testSearchItems()
		testPurchaseItems()
		testPriceTrends()
		testMarketplaceStats()
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

func runDemo() {
	fmt.Println("=== Federated Marketplace Demo ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	// Create some listings
	listings := []*economy.Listing{
		{
			ItemID:       "sword_001",
			ItemName:     "Iron Sword",
			ItemType:     "weapon",
			SellerID:     "player_123",
			SellerName:   "Alice",
			Price:        1000,
			Quantity:     3,
			MaxStackSize: 1,
		},
		{
			ItemID:       "sword_002",
			ItemName:     "Steel Sword",
			ItemType:     "weapon",
			SellerID:     "player_456",
			SellerName:   "Bob",
			Price:        2500,
			Quantity:     1,
			MaxStackSize: 1,
		},
		{
			ItemID:       "armor_001",
			ItemName:     "Iron Armor",
			ItemType:     "armor",
			SellerID:     "player_789",
			SellerName:   "Charlie",
			Price:        1500,
			Quantity:     2,
			MaxStackSize: 1,
		},
		{
			ItemID:       "potion_001",
			ItemName:     "Health Potion",
			ItemType:     "consumable",
			SellerID:     "player_101",
			SellerName:   "Dana",
			Price:        50,
			Quantity:     20,
			MaxStackSize: 10,
		},
	}

	for _, listing := range listings {
		err := marketplace.CreateListing(listing)
		if err != nil {
			log.Printf("Failed to create listing: %v", err)
		} else {
			fmt.Printf("Created listing: %s - %s (%d gold, qty: %d)\n",
				listing.ListingID, listing.ItemName, listing.Price, listing.Quantity)
		}
	}

	fmt.Println()
	fmt.Println("=== Marketplace Statistics ===")
	stats := marketplace.GetStats()
	for key, value := range stats {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Println()
}

func testCreateListings() {
	fmt.Println("=== Testing Listing Creation ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	// Test valid listing
	listing := &economy.Listing{
		ItemID:   "test_item_001",
		ItemName: "Test Sword",
		ItemType: "weapon",
		SellerID: "player_001",
		Price:    1000,
		Quantity: 5,
	}

	err := marketplace.CreateListing(listing)
	if err != nil {
		fmt.Printf("❌ Failed to create listing: %v\n", err)
	} else {
		fmt.Printf("✅ Created listing: %s\n", listing.ListingID)
		fmt.Printf("   Server ID: %s\n", listing.ServerID)
		fmt.Printf("   Created At: %s\n", listing.CreatedAt.Format(time.RFC3339))
		fmt.Printf("   Expires At: %s\n", listing.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("   Delivery: %s\n", listing.DeliveryMethod.String())
	}

	// Test invalid listing (missing seller)
	invalidListing := &economy.Listing{
		ItemID:   "test_item_002",
		Price:    1000,
		Quantity: 1,
	}

	err = marketplace.CreateListing(invalidListing)
	if err != nil {
		fmt.Printf("✅ Correctly rejected invalid listing: %v\n", err)
	} else {
		fmt.Println("❌ Should have rejected invalid listing")
	}

	fmt.Println()
}

func testSearchItems() {
	fmt.Println("=== Testing Item Search ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	// Create test data
	for i := 0; i < 10; i++ {
		marketplace.CreateListing(&economy.Listing{
			ItemID:   fmt.Sprintf("weapon_%d", i),
			ItemName: fmt.Sprintf("Sword Level %d", i),
			ItemType: "weapon",
			SellerID: "player_001",
			Price:    1000 + i*100,
			Quantity: 1,
		})
	}

	for i := 0; i < 5; i++ {
		marketplace.CreateListing(&economy.Listing{
			ItemID:   fmt.Sprintf("armor_%d", i),
			ItemName: fmt.Sprintf("Armor Level %d", i),
			ItemType: "armor",
			SellerID: "player_002",
			Price:    1500 + i*150,
			Quantity: 1,
		})
	}

	// Search all weapons
	fmt.Println("Search: All weapons")
	results, _ := marketplace.SearchItems(economy.ItemQuery{
		ItemType: "weapon",
	})
	fmt.Printf("Found %d weapons\n", len(results))

	// Search with price range
	fmt.Println("\nSearch: Weapons priced 1000-1500 gold")
	results, _ = marketplace.SearchItems(economy.ItemQuery{
		ItemType: "weapon",
		MinPrice: 1000,
		MaxPrice: 1500,
		SortBy:   economy.SortByPrice,
	})
	fmt.Printf("Found %d matching weapons:\n", len(results))
	for _, listing := range results {
		fmt.Printf("  - %s: %d gold\n", listing.ItemName, listing.Price)
	}

	// Search with limit
	fmt.Println("\nSearch: Top 3 items by price")
	results, _ = marketplace.SearchItems(economy.ItemQuery{
		SortBy: economy.SortByPrice,
		Limit:  3,
	})
	fmt.Printf("Top 3 cheapest items:\n")
	for i, listing := range results {
		fmt.Printf("  %d. %s: %d gold\n", i+1, listing.ItemName, listing.Price)
	}

	fmt.Println()
}

func testPurchaseItems() {
	fmt.Println("=== Testing Item Purchase ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	listing := &economy.Listing{
		ItemID:   "potion_001",
		ItemName: "Health Potion",
		ItemType: "consumable",
		SellerID: "seller_001",
		Price:    50,
		Quantity: 10,
	}

	marketplace.CreateListing(listing)

	fmt.Printf("Listing created: %s (qty: %d, price: %d each)\n",
		listing.ItemName, listing.Quantity, listing.Price)

	// Purchase 3 potions
	fmt.Println("\nBuyer purchases 3 potions...")
	totalCost := listing.GetTotalCost(3)
	fmt.Printf("Total cost: %d gold (includes transaction fee)\n", totalCost)

	err := marketplace.PurchaseItem(listing.ListingID, "buyer_001", 3)
	if err != nil {
		fmt.Printf("❌ Purchase failed: %v\n", err)
	} else {
		fmt.Println("✅ Purchase successful")

		// Check remaining quantity
		updated, _ := marketplace.GetListing(listing.ListingID)
		fmt.Printf("Remaining quantity: %d\n", updated.Quantity)
	}

	// Try to purchase more than available
	fmt.Println("\nAttempting to purchase 20 potions (more than available)...")
	err = marketplace.PurchaseItem(listing.ListingID, "buyer_002", 20)
	if err != nil {
		fmt.Printf("✅ Correctly rejected: %v\n", err)
	} else {
		fmt.Println("❌ Should have rejected insufficient quantity")
	}

	// Purchase all remaining
	fmt.Println("\nBuyer purchases all remaining potions...")
	updated, _ := marketplace.GetListing(listing.ListingID)
	err = marketplace.PurchaseItem(listing.ListingID, "buyer_003", updated.Quantity)
	if err != nil {
		fmt.Printf("❌ Purchase failed: %v\n", err)
	} else {
		fmt.Println("✅ Purchase successful")
		fmt.Println("Listing should be removed now")

		_, err = marketplace.GetListing(listing.ListingID)
		if err != nil {
			fmt.Println("✅ Listing removed as expected")
		} else {
			fmt.Println("❌ Listing should have been removed")
		}
	}

	fmt.Println()
}

func testPriceTrends() {
	fmt.Println("=== Testing Price Trends ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	// Create multiple weapon listings
	weaponPrices := []int{1000, 1200, 1500, 1100, 1300}
	for i, price := range weaponPrices {
		marketplace.CreateListing(&economy.Listing{
			ItemID:   fmt.Sprintf("weapon_%d", i),
			ItemName: "Iron Sword",
			ItemType: "weapon",
			SellerID: "seller_001",
			Price:    price,
			Quantity: 1,
		})
	}

	trend := marketplace.GetPriceTrend("weapon")
	fmt.Printf("Weapon Price Trend:\n")
	fmt.Printf("  Average: %d gold\n", trend.AveragePrice)
	fmt.Printf("  Min: %d gold\n", trend.MinPrice)
	fmt.Printf("  Max: %d gold\n", trend.MaxPrice)
	fmt.Printf("  Total Listings: %d\n", trend.TotalListings)
	fmt.Printf("  Total Volume: %d items\n", trend.TotalVolume)

	// Simulate purchase to see impact on trends
	fmt.Println("\nSimulating purchase to update trends...")
	listing, _ := marketplace.SearchItems(economy.ItemQuery{ItemType: "weapon", Limit: 1})
	if len(listing) > 0 {
		marketplace.PurchaseItem(listing[0].ListingID, "buyer_001", 1)

		trend = marketplace.GetPriceTrend("weapon")
		fmt.Printf("Updated Average Price: %d gold\n", trend.AveragePrice)
	}

	fmt.Println()
}

func testMarketplaceStats() {
	fmt.Println("=== Testing Marketplace Statistics ===")
	fmt.Println()

	marketplace := economy.NewFederatedMarketplace("server1")

	// Create local listings
	for i := 0; i < 50; i++ {
		marketplace.CreateListing(&economy.Listing{
			ItemID:   fmt.Sprintf("item_%d", i),
			ItemType: "weapon",
			SellerID: "seller_001",
			Price:    1000 + i*10,
			Quantity: 1,
		})
	}

	// Add remote cache
	remoteListings := make([]*economy.Listing, 0)
	for i := 0; i < 30; i++ {
		remoteListings = append(remoteListings, &economy.Listing{
			ItemID:    fmt.Sprintf("remote_%d", i),
			ServerID:  "server2",
			ItemType:  "armor",
			Price:     2000,
			Quantity:  1,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
	}
	marketplace.UpdateRemoteCache("server2", remoteListings)

	// Get stats
	stats := marketplace.GetStats()
	fmt.Println("Marketplace Statistics:")
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Test cleanup
	fmt.Println("\nRunning expired listing cleanup...")
	removed := marketplace.CleanupExpiredListings()
	fmt.Printf("Removed %d expired listings\n", removed)

	fmt.Println()
}
