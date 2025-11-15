// Command tradetest is a CLI utility for testing the item trading system.
// It demonstrates trade proposals, acceptance, rejection, trust mechanics,
// proximity validation, and atomic ownership transfer.
//
// Usage:
//
//	tradetest [-verbose]
//
// Examples:
//
//	tradetest          # Run basic trade scenarios
//	tradetest -verbose # Run with detailed logging
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/trade"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

var (
	verbose = flag.Bool("verbose", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	fmt.Println("=== Trade System Test ===")
	fmt.Println()

	// Test 1: Simple trade between two players
	fmt.Println("Test 1: Simple Trade")
	testSimpleTrade()
	fmt.Println()

	// Test 2: Trade rejection
	fmt.Println("Test 2: Trade Rejection")
	testTradeRejection()
	fmt.Println()

	// Test 3: Proximity validation
	fmt.Println("Test 3: Proximity Validation")
	testProximityValidation()
	fmt.Println()

	fmt.Println("=== All Tests Complete ===")
}

func testSimpleTrade() {
	world := engine.NewWorld()
	ts := trade.NewTradeSystem(world)

	// Create two players with items
	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{
		createItem("sword1", "Iron Sword", item.RarityCommon),
		createItem("potion1", "Health Potion", item.RarityCommon),
	})

	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{
		createItem("shield1", "Wooden Shield", item.RarityCommon),
	})

	fmt.Printf("  Alice has: Iron Sword, Health Potion\\n")
	fmt.Printf("  Bob has: Wooden Shield\\n")
	fmt.Printf("  Alice offers: Iron Sword\\n")
	fmt.Printf("  Alice wants: Wooden Shield\\n")

	// Propose trade
	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})
	if err != nil {
		log.Fatalf("  ❌ Failed to propose trade: %v", err)
	}
	fmt.Println("  ✓ Trade proposed successfully")

	// Accept trade
	err = ts.AcceptTrade(player2)
	if err != nil {
		log.Fatalf("  ❌ Failed to accept trade: %v", err)
	}
	fmt.Println("  ✓ Trade accepted and completed")

	// Verify items were exchanged
	p1Entity, _ := world.GetEntity(player1)
	p1Inv, _ := p1Entity.GetComponent("inventory")
	p1Items := p1Inv.(*engine.InventoryComponent).Items

	p2Entity, _ := world.GetEntity(player2)
	p2Inv, _ := p2Entity.GetComponent("inventory")
	p2Items := p2Inv.(*engine.InventoryComponent).Items

	hasShield := false
	for _, itm := range p1Items {
		if itm.ID == "shield1" {
			hasShield = true
		}
	}

	hasSword := false
	for _, itm := range p2Items {
		if itm.ID == "sword1" {
			hasSword = true
		}
	}

	if hasShield && hasSword {
		fmt.Println("  ✓ Items exchanged correctly")
	} else {
		log.Fatal("  ❌ Items not exchanged correctly")
	}
}

func testTradeRejection() {
	world := engine.NewWorld()
	ts := trade.NewTradeSystem(world)

	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{
		createItem("sword1", "Iron Sword", item.RarityCommon),
	})

	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{
		createItem("shield1", "Wooden Shield", item.RarityCommon),
	})

	// Propose trade
	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})
	if err != nil {
		log.Fatalf("  ❌ Failed to propose trade: %v", err)
	}
	fmt.Println("  ✓ Trade proposed")

	// Reject trade
	err = ts.RejectTrade(player2)
	if err != nil {
		log.Fatalf("  ❌ Failed to reject trade: %v", err)
	}
	fmt.Println("  ✓ Trade rejected")

	// Verify trade was cleared
	p2Entity, _ := world.GetEntity(player2)
	p2Trade, _ := p2Entity.GetComponent("trade")
	if p2Trade.(*engine.TradeComponent).ActiveTrade != nil {
		log.Fatal("  ❌ Trade not cleared after rejection")
	}
	fmt.Println("  ✓ Trade cleared correctly")
}

func testProximityValidation() {
	world := engine.NewWorld()
	ts := trade.NewTradeSystem(world)

	// Place players too far apart (>5 tiles)
	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{
		createItem("sword1", "Iron Sword", item.RarityCommon),
	})

	player2 := createPlayer(world, "Bob", 100, 100, []*item.Item{
		createItem("shield1", "Wooden Shield", item.RarityCommon),
	})

	// Attempt trade
	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})
	if err == nil {
		log.Fatal("  ❌ Trade should fail due to distance")
	}
	fmt.Printf("  ✓ Trade rejected due to distance: %v\\n", err)
}

// Helper functions

func createPlayer(world *engine.World, name string, x, y float64, items []*item.Item) uint64 {
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})

	inv := engine.NewInventoryComponent(100, 1000.0)
	for _, itm := range items {
		inv.AddItem(itm)
	}
	entity.AddComponent(inv)

	world.Update(0) // Process entity creation

	return entity.ID
}

func createItem(id, name string, rarity item.Rarity) *item.Item {
	return &item.Item{
		ID:     id,
		Name:   name,
		Type:   item.TypeWeapon,
		Rarity: rarity,
		Stats: item.Stats{
			Weight: 1.0,
			Value:  100,
		},
		Tags: []string{},
	}
}












































































































































































































































































































}	}		Tags: []string{},		},			Value:  100,			Weight: 1.0,		Stats: item.Stats{		Rarity: rarity,		Type:   item.TypeWeapon,		Name:   name,		ID:     id,	return &item.Item{func createItem(id, name string, rarity item.Rarity) *item.Item {}	return entity.ID	world.Update(0) // Process entity creation	entity.AddComponent(inv)	}		inv.AddItem(itm)	for _, itm := range items {	inv := engine.NewInventoryComponent(100, 1000.0)	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})	entity := world.CreateEntity()func createPlayer(world *engine.World, name string, x, y float64, items []*item.Item) uint64 {// Helper functions}	fmt.Printf("  ✓ Trade rejected (item not found): %v\n", err)	}		log.Fatal("  ❌ Trade should fail (item doesn't exist)")	if err == nil {	err := ts.ProposeTrade(player1, player2, []string{"nonexistent"}, []string{"shield1"})	// Attempt trade with non-existent item	})		createItem("shield1", "Wooden Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{	})		createItem("sword1", "Iron Sword", item.RarityCommon),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testOwnershipValidation() {}	fmt.Println("  ✓ Second trade rejected (player has active trade)")	}		log.Fatal("  ❌ Second trade should fail (player has active trade)")	if err == nil {	err = ts.ProposeTrade(player1, player3, []string{"sword1"}, []string{"axe1"})	// Attempt second trade while first is active	fmt.Println("  ✓ First trade proposed")	}		log.Fatalf("  ❌ Failed to propose first trade: %v", err)	if err != nil {	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})	// First trade	})		createItem("axe1", "Battle Axe", item.RarityCommon),	player3 := createPlayer(world, "Charlie", 1.5, 1.5, []*item.Item{	})		createItem("shield1", "Wooden Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{	})		createItem("sword1", "Iron Sword", item.RarityCommon),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testConcurrentTrades() {}	fmt.Printf("  ✓ Trade rejected due to low trust: %v\n", err)	}		log.Fatal("  ❌ Trade should fail due to low trust")	if err == nil {	err := ts.ProposeTrade(player1, player2, []string{"legendary1"}, []string{"common1"})	// Attempt to trade legendary item with low trust	})		TrustScore: 0.2, // Low trust	p1Entity.AddComponent(&engine.TradeComponent{	p1Entity, _ := world.GetEntity(player1)	// Set low trust on player1	})		createItem("common1", "Common Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{	})		createItem("legendary1", "Legendary Sword", item.RarityLegendary),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testTrustMechanics() {}	fmt.Printf("  ✓ Trade rejected due to distance: %v\n", err)	}		log.Fatal("  ❌ Trade should fail due to distance")	if err == nil {	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})	// Attempt trade	})		createItem("shield1", "Wooden Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 100, 100, []*item.Item{	})		createItem("sword1", "Iron Sword", item.RarityCommon),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	// Place players too far apart (>5 tiles)	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testProximityValidation() {}	fmt.Println("  ✓ Trade cleared correctly")	}		log.Fatal("  ❌ Trade not cleared after rejection")	if p2Trade.(*engine.TradeComponent).ActiveTrade != nil {	p2Trade, _ := p2Entity.GetComponent("trade")	p2Entity, _ := world.GetEntity(player2)	// Verify trade was cleared	fmt.Println("  ✓ Trade rejected")	}		log.Fatalf("  ❌ Failed to reject trade: %v", err)	if err != nil {	err = ts.RejectTrade(player2)	// Reject trade	fmt.Println("  ✓ Trade proposed")	}		log.Fatalf("  ❌ Failed to propose trade: %v", err)	if err != nil {	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})	// Propose trade	})		createItem("shield1", "Wooden Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{	})		createItem("sword1", "Iron Sword", item.RarityCommon),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testTradeRejection() {}	}		log.Fatal("  ❌ Items not exchanged correctly")	} else {		fmt.Println("  ✓ Items exchanged correctly")	if hasShield && hasSword {	}		}			hasSword = true		if itm.ID == "sword1" {	for _, itm := range p2Items {	hasSword := false	}		}			hasShield = true		if itm.ID == "shield1" {	for _, itm := range p1Items {	hasShield := false	p2Items := p2Inv.(*engine.InventoryComponent).Items	p2Inv, _ := p2Entity.GetComponent("inventory")	p2Entity, _ := world.GetEntity(player2)	p1Items := p1Inv.(*engine.InventoryComponent).Items	p1Inv, _ := p1Entity.GetComponent("inventory")	p1Entity, _ := world.GetEntity(player1)	// Verify items were exchanged	fmt.Println("  ✓ Trade accepted and completed")	}		log.Fatalf("  ❌ Failed to accept trade: %v", err)	if err != nil {	err = ts.AcceptTrade(player2)	// Accept trade	fmt.Println("  ✓ Trade proposed successfully")	}		log.Fatalf("  ❌ Failed to propose trade: %v", err)	if err != nil {	err := ts.ProposeTrade(player1, player2, []string{"sword1"}, []string{"shield1"})	// Propose trade	fmt.Printf("  Alice wants: Wooden Shield\n")	fmt.Printf("  Alice offers: Iron Sword\n")	fmt.Printf("  Bob has: Wooden Shield\n")	fmt.Printf("  Alice has: Iron Sword, Health Potion\n")	})		createItem("shield1", "Wooden Shield", item.RarityCommon),	player2 := createPlayer(world, "Bob", 1, 1, []*item.Item{	})		createItem("potion1", "Health Potion", item.RarityCommon),		createItem("sword1", "Iron Sword", item.RarityCommon),	player1 := createPlayer(world, "Alice", 0, 0, []*item.Item{	// Create two players with items	ts := trade.NewTradeSystem(world)	world := engine.NewWorld()func testSimpleTrade() {}	fmt.Println("=== All Tests Complete ===")	fmt.Println()	testOwnershipValidation()	fmt.Println("Test 6: Item Ownership Validation")	// Test 6: Item ownership validation	fmt.Println()	testConcurrentTrades()	fmt.Println("Test 5: Concurrent Trade Attempts")	// Test 5: Concurrent trade attempts	fmt.Println()	testTrustMechanics()	fmt.Println("Test 4: Trust Mechanics")	// Test 4: Trust mechanics	fmt.Println()	testProximityValidation()	fmt.Println("Test 3: Proximity Validation")	// Test 3: Proximity validation	fmt.Println()	testTradeRejection()	fmt.Println("Test 2: Trade Rejection")	// Test 2: Trade rejection	fmt.Println()	testSimpleTrade()	fmt.Println("Test 1: Simple Trade")	// Test 1: Simple trade between two players	fmt.Println()	fmt.Println("=== Trade System Test ===")	}		logrus.SetLevel(logrus.InfoLevel)	} else {		logrus.SetLevel(logrus.DebugLevel)	if *verbose {	flag.Parse()func main() {)	verbose = flag.Bool("verbose", false, "enable verbose logging")var ()	"github.com/sirupsen/logrus"	"github.com/opd-ai/venture/pkg/procgen/item"	"github.com/opd-ai/venture/pkg/network/trade"	"github.com/opd-ai/venture/pkg/engine"	"log"	"fmt"	"flag"import (package main//	tradetest -verbose # Run with detailed logging//	tradetest          # Run basic trade scenarios//// Examples:////	tradetest [-verbose]//// Usage://// proximity validation, and atomic ownership transfer.// It demonstrates trade proposals, acceptance, rejection, trust mechanics,