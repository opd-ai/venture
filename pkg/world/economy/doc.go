// Package economy implements cross-server economic systems including federated marketplaces,
// dynamic pricing, and automated trade.
//
// This package integrates V6 Federation, V5 Trading, and V8 Guilds to create a unified
// cross-server economy where players can trade items across the entire federation.
//
// # Federated Marketplace
//
// The marketplace allows players to:
//   - List items for sale on their local server
//   - Search for items across all federated servers
//   - Purchase items from remote servers with automatic delivery
//   - Track price trends and market dynamics
//
// # Dynamic Pricing
//
// The pricing engine automatically adjusts prices based on:
//   - Supply and demand across the federation
//   - Historical transaction data
//   - Server-specific economic conditions
//   - Market manipulation detection
//
// # Transaction Fees
//
// All marketplace transactions incur a 5-15% fee:
//   - 5% base fee for local server transactions
//   - +2% per federated server hop (max 15%)
//   - Fees are split between origin and destination servers
//
// # Delivery System
//
// Items purchased from remote servers are delivered via:
//   - V6 Mail system for small items (instant delivery)
//   - Courier NPCs for large items (10-60 minute delay)
//   - Requires players to be online to receive delivery
//
// # Example Usage
//
//	// Create marketplace
//	marketplace := economy.NewFederatedMarketplace(federation, mailSystem)
//
//	// List an item for sale
//	listing := &economy.Listing{
//	    ItemID:       "sword_123",
//	    SellerID:     "player_456",
//	    Price:        1000,
//	    Quantity:     1,
//	    ServerID:     "server_1",
//	}
//	err := marketplace.CreateListing(listing)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Search for items across federation
//	query := economy.ItemQuery{
//	    ItemType:    "weapon",
//	    MinPrice:    0,
//	    MaxPrice:    2000,
//	    SortBy:      economy.SortByPrice,
//	}
//	results, err := marketplace.SearchItems(query)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Purchase item from remote server
//	err = marketplace.PurchaseItem(results[0].ListingID, "buyer_789", 1)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get price trends
//	trend := marketplace.GetPriceTrend("sword")
//	log.Printf("Average price: %d gold", trend.AveragePrice)
//
// # Performance Targets
//
//   - Listing capacity: 10,000+ active listings per server
//   - Search performance: <100ms across 5+ servers
//   - Price update frequency: every 5 minutes
//   - Transaction volume: 100+ trades/hour peak
//
// # Test Coverage Target
//
// ≥65% coverage for all files in this package.
package economy
