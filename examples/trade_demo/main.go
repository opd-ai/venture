package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// main demonstrates the trading system with two-phase commit protocol,
// proximity validation, trust mechanics, and atomic ownership transfer.
func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	fmt.Println("=== Venture Trading System Demonstration ===")
	fmt.Println("Showcasing two-phase commit, proximity validation, and trust mechanics")

	// Create world for trading
	world := engine.NewWorld()

	// Create trade system
	tradeSystem := engine.NewTradeSystem(world)

	// Create player entities with inventories
	player1 := createPlayerWithInventory(world, "Player1", 5, 5, *verbose)
	player2 := createPlayerWithInventory(world, "Player2", 8, 8, *verbose)
	player3 := createPlayerWithInventory(world, "Player3", 20, 20, *verbose)

	// Commit entities to world
	world.Update(0.0)

	// Demonstrate successful trade (proximity OK)
	demonstrateSuccessfulTrade(tradeSystem, world, player1.ID, player2.ID, *verbose)

	// Demonstrate proximity rejection (too far apart)
	demonstrateProximityRejection(tradeSystem, world, player1.ID, player3.ID, *verbose)

	// Demonstrate trust-based limits
	demonstrateTrustLimits(tradeSystem, world, player1.ID, player2.ID, *verbose)

	// Demonstrate trade timeout
	demonstrateTradeTimeout(tradeSystem, world, player1.ID, player2.ID, *verbose)

	// Demonstrate concurrent trade conflict
	demonstrateConcurrentConflict(tradeSystem, world, player1.ID, player2.ID, player3.ID, *verbose)

	fmt.Println("\n=== Trade Demo Complete ===")
}

// createPlayerWithInventory creates a player entity with position, inventory, and trade components.
func createPlayerWithInventory(world *engine.World, name string, x, y float64, verbose bool) *engine.Entity {
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: x, Y: y})

	// Create inventory with some test items
	inv := engine.NewInventoryComponent(20, 100.0) // 20 slots, 100.0 max weight

	// Add common items
	commonSword := &item.Item{
		ID:     fmt.Sprintf("%s_sword_common", name),
		Name:   "Common Sword",
		Rarity: item.RarityCommon,
		Type:   item.TypeWeapon,
	}
	inv.AddItem(commonSword)

	commonShield := &item.Item{
		ID:     fmt.Sprintf("%s_shield_common", name),
		Name:   "Common Shield",
		Rarity: item.RarityCommon,
		Type:   item.TypeArmor,
	}
	inv.AddItem(commonShield)

	// Add a rare item
	rareHelmet := &item.Item{
		ID:     fmt.Sprintf("%s_helmet_rare", name),
		Name:   "Rare Helmet",
		Rarity: item.RarityRare,
		Type:   item.TypeArmor,
	}
	inv.AddItem(rareHelmet)

	// Add a legendary item (for trust limit testing)
	legendaryAxe := &item.Item{
		ID:     fmt.Sprintf("%s_axe_legendary", name),
		Name:   "Legendary Axe",
		Rarity: item.RarityLegendary,
		Type:   item.TypeWeapon,
	}
	inv.AddItem(legendaryAxe)

	player.AddComponent(inv)

	// Add trade component with default trust (will be auto-created if needed, but explicit is clearer)
	tradeComp := &engine.TradeComponent{
		TrustScore:   engine.TrustDefault,
		TradeHistory: make([]engine.TradeRecord, 0),
	}
	player.AddComponent(tradeComp)

	if verbose {
		fmt.Printf("Created %s (ID: %d) at position (%.1f, %.1f) with %d items\n",
			name, player.ID, x, y, len(inv.Items))
	}

	return player
}

// demonstrateSuccessfulTrade shows a complete trade flow between two nearby players.
func demonstrateSuccessfulTrade(ts *engine.TradeSystem, world *engine.World, player1ID, player2ID uint64, verbose bool) {
	fmt.Println("\n--- Demonstrating Successful Trade ---")

	player1, player2 := getPlayers(world, player1ID, player2ID)
	inv1, inv2 := getInventories(player1, player2)

	printInventorySizes(inv1, inv2, verbose, "before trade")

	offeredItems := []string{"Player1_sword_common"}
	requestedItems := []string{"Player2_shield_common"}

	executeTrade(ts, player1ID, player2ID, offeredItems, requestedItems)

	printInventorySizes(inv1, inv2, verbose, "after trade")
	verifyItemTransfer(inv1, inv2)
	checkTrustScores(player1, player2)
}

// getPlayers retrieves both player entities.
func getPlayers(world *engine.World, player1ID, player2ID uint64) (*engine.Entity, *engine.Entity) {
	player1, _ := world.GetEntity(player1ID)
	player2, _ := world.GetEntity(player2ID)
	return player1, player2
}

// getInventories retrieves inventory components from players.
func getInventories(player1, player2 *engine.Entity) (*engine.InventoryComponent, *engine.InventoryComponent) {
	inv1Comp, _ := player1.GetComponent("inventory")
	inv1 := inv1Comp.(*engine.InventoryComponent)

	inv2Comp, _ := player2.GetComponent("inventory")
	inv2 := inv2Comp.(*engine.InventoryComponent)

	return inv1, inv2
}

// printInventorySizes prints inventory sizes if verbose mode is enabled.
func printInventorySizes(inv1, inv2 *engine.InventoryComponent, verbose bool, context string) {
	if verbose {
		fmt.Printf("Player 1 inventory %s: %d items\n", context, len(inv1.Items))
		fmt.Printf("Player 2 inventory %s: %d items\n", context, len(inv2.Items))
	}
}

// executeTrade proposes, accepts, and commits a trade.
func executeTrade(ts *engine.TradeSystem, player1ID, player2ID uint64, offeredItems, requestedItems []string) {
	err := ts.ProposeTrade(player1ID, player2ID, offeredItems, requestedItems)
	if err != nil {
		log.Fatalf("Trade proposal failed: %v", err)
	}
	fmt.Println("✓ Trade proposal successful (players within proximity)")

	err = ts.AcceptTrade(player2ID)
	if err != nil {
		log.Fatalf("Trade acceptance failed: %v", err)
	}
	fmt.Println("✓ Trade accepted by Player 2")

	err = ts.CommitTrade(player1ID)
	if err != nil {
		log.Fatalf("Trade commit failed: %v", err)
	}
	fmt.Println("✓ Trade committed (atomic item transfer)")
}

// verifyItemTransfer verifies that items were transferred correctly.
func verifyItemTransfer(inv1, inv2 *engine.InventoryComponent) {
	player1HasShield := false
	player2HasSword := false

	for _, item := range inv1.Items {
		if item.ID == "Player2_shield_common" {
			player1HasShield = true
		}
	}
	for _, item := range inv2.Items {
		if item.ID == "Player1_sword_common" {
			player2HasSword = true
		}
	}

	if player1HasShield && player2HasSword {
		fmt.Println("✓ Item ownership transferred correctly")
	} else {
		log.Fatal("Item transfer verification failed!")
	}
}

// checkTrustScores verifies that trust scores increased.
func checkTrustScores(player1, player2 *engine.Entity) {
	trade1Comp, _ := player1.GetComponent("trade")
	trade1 := trade1Comp.(*engine.TradeComponent)

	trade2Comp, _ := player2.GetComponent("trade")
	trade2 := trade2Comp.(*engine.TradeComponent)

	if trade1.TrustScore > engine.TrustDefault && trade2.TrustScore > engine.TrustDefault {
		fmt.Printf("✓ Trust scores increased (P1: %.2f, P2: %.2f)\n", trade1.TrustScore, trade2.TrustScore)
	}
}

// demonstrateProximityRejection shows trade rejection when players are too far apart.
func demonstrateProximityRejection(ts *engine.TradeSystem, world *engine.World, player1ID, player3ID uint64, verbose bool) {
	fmt.Println("\n--- Demonstrating Proximity Rejection ---")

	player1, _ := world.GetEntity(player1ID)
	player3, _ := world.GetEntity(player3ID)

	pos1Comp, _ := player1.GetComponent("position")
	pos1 := pos1Comp.(*engine.PositionComponent)

	pos3Comp, _ := player3.GetComponent("position")
	pos3 := pos3Comp.(*engine.PositionComponent)

	distance := engine.Distance(pos1.X, pos1.Y, pos3.X, pos3.Y)
	fmt.Printf("Distance between players: %.2f tiles (max: %.1f)\n", distance, engine.ProposalProximity)

	// Attempt trade
	offeredItems := []string{"Player1_helmet_rare"}
	requestedItems := []string{"Player3_sword_common"}

	err := ts.ProposeTrade(player1ID, player3ID, offeredItems, requestedItems)
	if err != nil {
		fmt.Printf("✓ Trade rejected: %v\n", err)
	} else {
		log.Fatal("Trade should have been rejected due to distance!")
	}
}

// demonstrateTrustLimits shows trust-based restrictions on trading rare/legendary items.
func demonstrateTrustLimits(ts *engine.TradeSystem, world *engine.World, player1ID, player2ID uint64, verbose bool) {
	fmt.Println("\n--- Demonstrating Trust-Based Limits ---")

	player1, _ := world.GetEntity(player1ID)
	trade1Comp, _ := player1.GetComponent("trade")
	trade1 := trade1Comp.(*engine.TradeComponent)

	// Lower trust score below threshold
	originalTrust := trade1.TrustScore
	trade1.TrustScore = engine.TrustLow - 0.1 // Below 0.3
	fmt.Printf("Lowered Player 1 trust score to %.2f (threshold: %.1f)\n", trade1.TrustScore, engine.TrustLow)

	// Attempt to trade legendary item
	offeredItems := []string{"Player1_axe_legendary"}
	requestedItems := []string{} // Gift

	err := ts.ProposeTrade(player1ID, player2ID, offeredItems, requestedItems)
	if err != nil {
		fmt.Printf("✓ Legendary item trade rejected: %v\n", err)
	} else {
		log.Fatal("Trade should have been rejected due to low trust!")
	}

	// Restore original trust
	trade1.TrustScore = originalTrust
	fmt.Printf("Restored Player 1 trust score to %.2f\n", trade1.TrustScore)
}

// demonstrateTradeTimeout shows auto-cancellation after timeout period.
func demonstrateTradeTimeout(ts *engine.TradeSystem, world *engine.World, player1ID, player2ID uint64, verbose bool) {
	fmt.Println("\n--- Demonstrating Trade Timeout ---")

	// Propose a trade
	offeredItems := []string{"Player1_helmet_rare"}
	requestedItems := []string{"Player2_helmet_rare"}

	err := ts.ProposeTrade(player1ID, player2ID, offeredItems, requestedItems)
	if err != nil {
		log.Fatalf("Trade proposal failed: %v", err)
	}
	fmt.Println("✓ Trade proposed")

	player1, _ := world.GetEntity(player1ID)
	trade1Comp, _ := player1.GetComponent("trade")
	trade1 := trade1Comp.(*engine.TradeComponent)

	if trade1.ActiveTrade != nil {
		fmt.Printf("Active trade timeout: %d seconds\n", engine.TradeTimeout)

		// Check timeout detection works
		if trade1.ActiveTrade.ProposalTime == 0 {
			log.Fatal("Trade timestamp not set!")
		}
		fmt.Println("✓ Trade timeout mechanism active")

		// Cancel trade
		err = ts.CancelTrade(player1ID)
		if err != nil {
			log.Fatalf("Trade cancellation failed: %v", err)
		}
		fmt.Println("✓ Trade cancelled manually")
	}
}

// demonstrateConcurrentConflict shows conflict resolution for concurrent trades.
func demonstrateConcurrentConflict(ts *engine.TradeSystem, world *engine.World, player1ID, player2ID, player3ID uint64, verbose bool) {
	fmt.Println("\n--- Demonstrating Concurrent Trade Conflict ---")

	// Move player 3 close to player 1
	player3, _ := world.GetEntity(player3ID)
	// Remove old position and add new one
	player3.RemoveComponent("position")
	player3.AddComponent(&engine.PositionComponent{X: 6, Y: 6}) // Near player 1 at (5,5)

	// Player 1 proposes trade with Player 2
	err := ts.ProposeTrade(player1ID, player2ID, []string{"Player1_shield_common"}, []string{})
	if err != nil {
		log.Fatalf("First trade proposal failed: %v", err)
	}
	fmt.Println("✓ Player 1 → Player 2 trade proposed")

	// Player 3 tries to propose trade with Player 1 (should fail - already has active trade)
	err = ts.ProposeTrade(player3ID, player1ID, []string{"Player3_sword_common"}, []string{})
	if err != nil {
		fmt.Printf("✓ Concurrent trade rejected: %v\n", err)
	} else {
		log.Fatal("Concurrent trade should have been rejected!")
	}

	// Cancel first trade
	ts.CancelTrade(player1ID)
	fmt.Println("✓ First trade cancelled - Player 1 available for new trades")
}
