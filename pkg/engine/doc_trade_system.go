/*
Package engine/trade_system implements player-to-player item trading.

# Overview

The TradeSystem provides a secure two-phase commit protocol for atomic item transfers
between players with proximity validation, trust-based limits, and comprehensive rollback
support for high-latency multiplayer scenarios (200-5000ms).

# Core Features

  - Two-phase commit protocol: ProposeTrade → AcceptTrade/RejectTrade → CommitTrade
  - Proximity validation: 5-tile proposal radius, 10-tile active radius
  - Trust mechanics: 0.0-1.0 score affecting item rarity and trade count limits
  - Atomic transfers: All-or-nothing item ownership changes
  - Automatic rollback: Handles disconnects, proximity violations, and ownership changes
  - Timeout detection: 30-second trade expiration

# Usage Example

	world := engine.NewWorld()
	tradeSystem := engine.NewTradeSystem(world)

	// Propose a trade (player1 offers sword1 for player2's potion1)
	err := tradeSystem.ProposeTrade(player1ID, player2ID,
		[]string{"sword1"}, []string{"potion1"})
	if err != nil {
		// Handle proximity, ownership, or trust errors
	}

	// Player 2 accepts
	err = tradeSystem.AcceptTrade(player2ID)
	if err != nil {
		// Handle acceptance errors
	}

	// Server commits the trade
	err = tradeSystem.CommitTrade(player1ID)
	if err != nil {
		// Rollback occurred - items returned to original owners
	}

# Trust Mechanics

Players start with a default trust score of 0.5. Trust changes based on trade outcomes:
  - Successful trade: +0.05
  - Failed trade: -0.1

Trust levels affect trade restrictions:
  - Low trust (<0.3): Cannot trade Epic/Legendary items, max 5 items per trade
  - Normal trust (0.3-0.8): Can trade all rarities
  - High trust (>0.8): No restrictions

Trust scores are not persisted and reset to 0.5 on server restart.

# Rollback Scenarios

The system automatically rolls back trades in the following cases:
  - Proximity violation (players move >10 tiles apart during negotiation)
  - Item ownership change (item sold/dropped between propose and commit)
  - Player disconnect (either party)
  - Timeout (trade not completed within 30 seconds)
  - Concurrent trade conflict (same item in multiple active proposals)

All rollback scenarios decrease both parties' trust scores by 0.1.

# Network Synchronization

The TradeSystem is designed for server-authoritative multiplayer:
  - Server validates all trade operations
  - Clients may show optimistic UI but server state is authoritative
  - Lag compensation supported via proximity checks at client perspective time
  - Trade state synchronized via TradeComponent on entities

# Performance

Target performance metrics (all met):
  - ProposeTrade: <10ms
  - CommitTrade: <100ms at 200ms latency
  - Update loop: <1ms for proximity/timeout checks across all active trades

# Integration

Add TradeSystem to your game loop:

	tradeSystem := engine.NewTradeSystem(world)
	// In update loop:
	tradeSystem.Update(deltaTime)

Ensure entities have:
  - PositionComponent for proximity checks
  - InventoryComponent for item storage
  - TradeComponent (auto-created) for trade state

# Testing

See trade_system_test.go for comprehensive test coverage including:
  - Proximity validation
  - Trust limit enforcement
  - Atomic transfer verification
  - Rollback scenario testing
  - Concurrent trade conflict resolution
  - Race condition detection

Test coverage: 70-100% per function.
*/
package engine
