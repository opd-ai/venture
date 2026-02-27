// Package economy implements cross-server economic systems including federated marketplaces,
// guild banks, dynamic pricing, and automated trade.
//
// This package integrates V6 Federation, V5 Trading, and V8 Guilds to create a unified
// cross-server economy where players can trade items across the entire federation and
// manage shared guild resources with persistent storage.
//
// # Federated Marketplace
//
// The marketplace allows players to:
//   - List items for sale on their local server
//   - Search for items across all federated servers
//   - Purchase items from remote servers with automatic delivery
//   - Track price trends and market dynamics
//
// # Guild Banks
//
// Guild banks provide shared storage and treasury management:
//   - Vault capacity: 5,000+ unique items per guild
//   - Cross-server treasury synchronization
//   - Rank-based withdrawal limits (0-10k gold/day)
//   - Resource pooling for guild projects
//   - Bank interest: 0.1-1.0% daily on deposited gold
//   - Comprehensive audit logs (30 day retention)
//   - Save/load persistence with gzip compression
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
// Note: Examples use log.Fatal and log.Printf for simplicity.
// Production code should use logrus.WithError() and logrus.WithFields()
// for structured logging per project guidelines.
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
//	// Create guild bank
//	bankManager := economy.NewGuildBankManager()
//	err = bankManager.CreateVault("guild_123", 0.005) // 0.5% daily interest
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Set withdrawal limits by rank
//	bankManager.SetWithdrawalLimit("guild_123", "recruit", 500)
//	bankManager.SetWithdrawalLimit("guild_123", "member", 2000)
//	bankManager.SetWithdrawalLimit("guild_123", "officer", 5000)
//
//	// Deposit gold into guild vault
//	err = bankManager.DepositGold("guild_123", "player_456", "Alice", 10000)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Withdraw gold with rank-based limits
//	err = bankManager.WithdrawGold("guild_123", "player_789", "Bob", "recruit", 300)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Deposit items into vault
//	err = bankManager.DepositItem("guild_123", "player_456", "Alice", "sword_123", "Iron Sword", "weapon", 5, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Calculate daily interest
//	err = bankManager.CalculateInterest("guild_123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Retrieve audit log
//	entries, err := bankManager.GetAuditLog("guild_123", 50)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, entry := range entries {
//	    log.Printf("%s: %s by %s", entry.Timestamp, entry.ActionType, entry.MemberName)
//	}
//
// # Performance Targets
//
//   - Listing capacity: 10,000+ active listings per server
//   - Search performance: <100ms across 5+ servers
//   - Price update frequency: every 5 minutes
//   - Transaction volume: 100+ trades/hour peak
//   - Vault capacity: 5,000+ unique items per guild
//   - Sync latency: <5 seconds cross-server
//   - Transaction log retention: 30 days
//   - Interest calculation: daily at server midnight
//
// # Test Coverage Target
//
// ≥40% coverage for all files in this package.
package economy
