// Package federation provides server-to-server federation protocol for Venture.
//
// The federation system enables multiple Venture servers to connect and form a network,
// allowing players to travel between servers, trade items, send mail, and participate
// in cross-server political and territorial systems.
//
// # Architecture
//
// Federation uses a peer-to-peer model where each server maintains direct connections
// to other servers. There is no central authority - servers discover each other via:
//   - LAN broadcast (UDP discovery on port 8090)
//   - Manual server addition (IP:port + fingerprint verification)
//   - Optional relay servers for NAT traversal
//   - Gossip protocol for multi-hop server discovery
//
// # Security Model
//
// Federation uses ed25519 public-key cryptography for server identity and authentication:
//   - Each server generates a unique keypair on first launch
//   - Server ID is the SHA-256 fingerprint of the public key (64 hex characters)
//   - All handshakes are signed with the server's private key
//   - Replay attacks are prevented via nonce tracking (16-byte random values)
//   - Trust-On-First-Use (TOFU) model: players manually verify fingerprints
//
// # Trust Levels
//
// Servers are classified into three trust levels:
//   - Unknown: No prior interaction, read-only federation, player travel disabled
//   - Verified: Certificate exchange complete, limited features (travel + trade only)
//   - Trusted: Known server, full feature access (all game mechanics enabled)
//
// # Handshake Protocol
//
// Server connection establishment follows this flow:
//  1. Server A creates ServerIdentity (ed25519 keypair)
//  2. Server A sends FederationHandshake to Server B
//  3. Server B verifies signature and timestamp (60-second window)
//  4. Server B checks nonce for replay attacks (5-minute expiry)
//  5. Server B creates response handshake
//  6. Both servers negotiate common features
//  7. Connection established with agreed feature set
//
// # Example Usage
//
//	// Create server identity
//	identity, err := federation.NewServerIdentity("MyServer")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Display fingerprint for manual verification
//	fmt.Println("Server fingerprint:", identity.GetFingerprint())
//
//	// Create handshake
//	handshake, err := identity.CreateHandshake(
//	    "6.0.0",
//	    []string{"travel", "trade", "post"},
//	    federation.TrustVerified,
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Verify received handshake
//	err = federation.VerifyHandshake(handshake)
//	if err != nil {
//	    log.Fatal("Invalid handshake:", err)
//	}
//
//	// Process with HandshakeManager for replay prevention
//	manager := federation.NewHandshakeManager(identity)
//	err = manager.ProcessHandshake(handshake)
//	if err != nil {
//	    log.Fatal("Replay attack detected:", err)
//	}
//
//	// Negotiate features
//	common := federation.NegotiateFeatures(
//	    []string{"travel", "trade", "post"},
//	    []string{"travel", "post", "bounty"},
//	)
//	fmt.Println("Common features:", common) // ["travel", "post"]
//
// # State Synchronization
//
// Federation state management uses FederationState and SyncManager:
//
//	// Create federation state
//	state := federation.NewFederationState()
//
//	// Add connected server
//	state.AddServer(&federation.ServerInfo{
//	    ServerID:    "server1",
//	    ServerName:  "Fantasy Server",
//	    Address:     "192.168.1.100:8080",
//	    Version:     "6.0.0",
//	    Features:    []string{"travel", "trade"},
//	    PlayerCount: 25,
//	    Reputation:  0.9,
//	})
//
//	// Create sync manager
//	syncMgr := federation.NewSyncManager(state)
//	syncMgr.Start()
//	defer syncMgr.Stop()
//
//	// Process heartbeat from peer
//	heartbeat := syncMgr.CreateHeartbeat("local-server", 30)
//	syncMgr.ProcessHeartbeat(heartbeat)
//
//	// Sync market prices
//	state.UpdateMarketPrice("sword", 100.0)
//	marketSync := syncMgr.CreateMarketSync("local-server")
//	syncMgr.ProcessMarketSync(marketSync)
//
//	// Check for stale servers
//	staleServers := state.CheckStaleServers(30 * time.Second)
//	for _, serverID := range staleServers {
//	    fmt.Printf("Server %s is offline\n", serverID)
//	}
//
// # Trade Network
//
// The trade network provides dynamic cross-server item trading with market-driven pricing:
//
//	// Create federated market
//	market := federation.NewFederatedMarket()
//	market.Start() // Begin 60-second price updates
//	defer market.Stop()
//
//	// Register items
//	market.RegisterItem("sword", "server1", 100.0)
//	market.RegisterItem("potion", "server1", 25.0)
//
//	// Update supply and demand
//	market.UpdateSupply("sword", 50)  // 50 swords available
//	market.UpdateDemand("sword", 75)  // 75 buy orders
//
//	// Calculate price (dynamic based on supply/demand)
//	price := market.CalculatePrice("sword", 1.0) // serverMultiplier = 1.0
//	fmt.Printf("Sword price: %.2f gold (%.1fx base)\n", price, price/100.0)
//
//	// Calculate shipping cost
//	shipping := federation.CalculateShippingCost(price, 3) // 3 server hops
//	totalCost := price + shipping
//	fmt.Printf("Total cost with shipping: %.2f gold\n", totalCost)
//
//	// Get price history
//	history, _ := market.GetPriceHistory("sword")
//	fmt.Printf("Current: %.2f, Base: %.2f, History: %d points\n",
//	    history.CurrentPrice, history.BasePrice, len(history.History))
//
//	// Market statistics
//	stats := market.GetStats()
//	fmt.Printf("Market: %d items, %d total supply, %d total demand\n",
//	    stats.TotalItems, stats.TotalSupply, stats.TotalDemand)
//
// Pricing formula: Price = BasePrice × (Demand / Supply) × ServerMultiplier
//   - Ratio clamped to 0.2x-5.0x (prevents extreme price swings)
//   - Zero supply triggers 3x base price (scarcity premium)
//   - ServerMultiplier from political system (0.8x ally, 1.0x neutral, 1.5x enemy)
//   - Shipping adds 10% per server hop
//
// Price history tracks 288 data points (24 hours at 5-minute intervals).
// Market updates run every 60 seconds to recalculate prices based on current supply/demand.
//
// # Merchant Caravans
//
// Merchant caravans are NPC traders that travel between servers carrying goods:
//
//	// Create caravan system (in engine package)
//	import "github.com/opd-ai/venture/pkg/engine"
//
//	world := engine.NewWorld()
//	caravanSys := engine.NewMerchantCaravanSystem(world)
//	caravanSys.SetHopDuration(300.0) // 5 minutes per server hop
//
//	// Create caravan inventory
//	inventory := []engine.CaravanItem{
//	    {ItemID: "sword", Quantity: 10, PurchasePrice: 100, SalePrice: 120},
//	    {ItemID: "potion", Quantity: 50, PurchasePrice: 25, SalePrice: 30},
//	}
//
//	// Spawn caravan
//	caravan := caravanSys.CreateCaravan("server1", "server3", inventory)
//
//	// Update travel (called each frame)
//	caravanSys.Update(deltaTime)
//
//	// Check arrival time
//	eta := caravanSys.EstimateArrivalTime(caravan)
//	fmt.Printf("ETA: %s\n", time.Unix(eta, 0))
//
//	// Get caravans at a server
//	caravansAtServer := caravanSys.GetCaravansAtServer("server2")
//	fmt.Printf("Caravans at server2: %d\n", len(caravansAtServer))
//
//	// Calculate sale price with markup
//	salePrice := caravanSys.CalculateSalePrice(100.0, 3) // 3 hops
//	fmt.Printf("Sale price: %.2f (markup from distance)\n", salePrice)
//
// Merchant markup formula: 10% minimum + distance-based scaling up to 50% maximum
//   - 0 hops: 10% markup (local trade)
//   - 5 hops: ~30% markup
//   - 10+ hops: 50% markup (capped)
//
// Caravans rest at destinations for 10 minutes before returning. Travel time is
// 5 minutes per server hop by default (configurable).
//
// # Protocol Version Compatibility
//
// Version compatibility uses semantic versioning (major.minor.patch):
//   - Compatible if major versions match (6.0.0 and 6.1.0 are compatible)
//   - Incompatible if major versions differ (6.0.0 and 7.0.0 are incompatible)
//   - Feature negotiation handles minor version differences
//
// # Performance Characteristics
//
// Operations are optimized for minimal latency:
//
// Handshake operations:
//   - ServerIdentity creation: ~16µs per keypair
//   - Handshake creation: ~21µs per handshake
//   - Handshake verification: ~48µs per verification
//   - Replay check: ~104µs including nonce tracking
//   - Feature negotiation: ~183ns
//
// State synchronization:
//   - AddServer: ~66ns (0 allocations)
//   - UpdateServer: ~60ns (0 allocations)
//   - UpdateMarketPrice: ~54ns (0 allocations)
//   - ProcessHeartbeat: ~59ns (0 allocations)
//   - ProcessMarketSync: ~204ns (0 allocations)
//
// Memory usage is minimal:
//   - ServerIdentity: 384 bytes
//   - FederationHandshake: 680 bytes
//   - Nonce cache: 16 bytes per tracked nonce (auto-cleanup after 5 minutes)
//   - ServerInfo: ~200 bytes per connected server
//   - FederationState: 48 bytes base + (ServerInfo × server count)
//
// # Network Budget
//
// Federation protocol targets <5KB/s per server connection:
//   - Handshake: ~1KB one-time
//   - Heartbeat: 10-second intervals (minimal overhead)
//   - State sync: 60-second intervals for market prices
//   - Political events: Immediate (event-driven)
//
// # Thread Safety
//
// All federation types are safe for concurrent use:
//   - ServerIdentity uses RWMutex for keypair access
//   - HandshakeManager uses RWMutex for nonce tracking
//   - FederationState uses RWMutex for all state operations
//   - SyncManager uses goroutines with proper synchronization
//   - Nonce cleanup runs asynchronously without blocking
//
// # Testing
//
// Use the federationtest CLI tool to test handshake functionality:
//
//	go run cmd/federationtest/main.go -mode list
//	go run cmd/federationtest/main.go -mode identity -name "MyServer"
//	go run cmd/federationtest/main.go -mode handshake -verbose
//	go run cmd/federationtest/main.go -mode verify
//	go run cmd/federationtest/main.go -mode negotiate -features "travel,trade"
package federation
