// Package trade provides item trading between players with two-phase commit protocol,
// proximity validation, trust mechanics, and atomic ownership transfer.
//
// # Trade System Overview
//
// The trade system implements a comprehensive item trading mechanism for multiplayer
// gameplay with the following key features:
//
// - Two-phase commit protocol for atomic item transfers
// - Proximity-based validation (players must be within 5 tiles to propose, 10 tiles during trade)
// - Trust score mechanics (0.0-1.0) affecting tradable item rarity and quantity
// - Automatic timeout and cancellation (30-second proposal timeout)
// - Rollback on disconnect or validation failure
//
// # Trade Workflow
//
// 1. **Propose**: Player A proposes a trade to Player B with offered and requested items
// 2. **Review**: Player B reviews the proposal (can accept, reject, or counter-propose)
// 3. **Validate**: Server validates proximity, trust, ownership, and inventory space
// 4. **Commit**: Server atomically transfers items between players
// 5. **Complete**: Both players' trust scores and trade history are updated
//
// # Trust Mechanics
//
// Trust scores range from 0.0 to 1.0 (default: 0.5) and affect trading capabilities:
//
// - Low trust (<0.3): Max 5 items per trade, common/uncommon only
// - Medium trust (0.3-0.8): No restrictions except legendary items
// - High trust (>0.8): Can trade all rarities including legendary
//
// Trust scores are updated based on trade outcomes:
//
// - Successful trade: +0.05
// - Failed trade: -0.10
//
// # Proximity Rules
//
// - Maximum proposal distance: 5 tiles (Euclidean distance)
// - Maximum active trade distance: 10 tiles (auto-cancel if exceeded)
// - Distance validation uses lag compensation for multiplayer fairness
//
// # Example Usage
//
//	// Create trade system
//	world := engine.NewWorld()
//	ts := trade.NewTradeSystem(world)
//
//	// Setup players with inventories
//	proposer := world.CreateEntity()
//	proposer.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
//	proposer.AddComponent(engine.NewInventoryComponent(100, 1000.0))
//
//	recipient := world.CreateEntity()
//	recipient.AddComponent(&engine.PositionComponent{X: 2, Y: 2})
//	recipient.AddComponent(engine.NewInventoryComponent(100, 1000.0))
//
//	// Propose trade
//	offeredItems := []string{"item_id_1", "item_id_2"}
//	requestedItems := []string{"item_id_3"}
//	err := ts.ProposeTrade(proposer.ID, recipient.ID, offeredItems, requestedItems)
//	if err != nil {
//	    // Handle error (too far apart, already trading, etc.)
//	}
//
//	// Quantity-bearing proposals are also supported
//	err = ts.ProposeTradeWithQuantities(
//	    proposer.ID,
//	    recipient.ID,
//	    []engine.TradeLineItem{{ItemID: "potion", Quantity: 2}},
//	    []engine.TradeLineItem{{ItemID: "herb", Quantity: 3}},
//	)
//
//	// Recipient accepts trade
//	err = ts.AcceptTrade(recipient.ID)
//	if err != nil {
//	    // Handle error (items moved, inventory full, etc.)
//	}
//
//	// Or recipient rejects trade
//	err = ts.RejectTrade(recipient.ID)
//
// # Integration with Network Layer
//
// The trade system is designed to work with the network protocol for multiplayer:
//
// - Trade proposals are validated on the server (authoritative)
// - Client can display optimistic UI but server determines success/failure
// - All validation (proximity, trust, ownership) happens server-side
// - Rollback mechanism handles disconnect and concurrent modification
//
// ## Validation Order in ProposeTrade
//
// Trade proposals follow a strict validation sequence for performance and security:
//
// 1. **Rate Limiting** (system.go:137): Check if proposer exceeds 10 requests/second - fail fast
// 2. **Format Validation** (system.go): Validate item ID and quantity formats via pkg/validation
// 3. **Entity Validation**: Verify proposer and recipient exist with required components
// 4. **Proximity Validation**: Ensure players are within 5 tiles
// 5. **Trust Validation**: Check trust score allows proposed items
// 6. **Inventory Validation**: Verify ownership and space availability
//
// This order minimizes wasted computation by rejecting invalid requests early.
//
// # Performance Considerations
//
// - Trade proposals: <10ms validation time
// - Trade commits: <100ms (includes item transfer and inventory updates)
// - Memory overhead: ~1KB per active trade
// - Update loop processes timeouts and proximity checks for all active trades
//
// # Thread Safety
//
// The trade system is NOT thread-safe. It is designed to be used from a single
// game loop thread. For multiplayer servers, ensure all trade operations are
// executed on the main server thread.
//
// # Known Limitations
//
// ## Dual Trade System Architecture
//
// The Venture codebase contains TWO separate trade system implementations:
//
// 1. **pkg/network/trade** (this package): Network layer implementation
//   - Provides validation, rate limiting, and time abstraction (TimeProvider)
//   - Focuses on multiplayer synchronization and security
//   - Registered in cmd/client as networkTradeSystemWrapper
//
// 2. **pkg/engine/trade_system.go**: Engine layer implementation
//   - Provides social integration and single-player functionality
//   - Handles trade UI interactions (pkg/engine/trade_ui.go)
//   - Registered in cmd/client as tradeSystemWrapper
//
// Both systems are registered as separate ECS systems in the client (cmd/client/handlers.go:2132, :2172).
// The engine layer system delegates validation and rate limiting to this network layer in multiplayer
// scenarios. This separation enables:
//
// - Testable network validation logic independent of game engine
// - Reusable validation rules across single-player and multiplayer
// - Clear separation of concerns (network vs gameplay)
//
// For server integration, the authoritative server uses pkg/engine/trade_system.go directly rather than
// this network layer package.
package trade
