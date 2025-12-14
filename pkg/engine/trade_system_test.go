package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/social"
)

// TestProposeTrade_Success verifies successful trade proposal.
func TestProposeTrade_Success(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	// Create two players with inventories
	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0}) // 3 tiles away
	recipientInv := NewInventoryComponent(10, 100.0)
	recipient.AddComponent(recipientInv)

	world.Update(0.0) // Commit entities

	// Add items to proposer
	item1 := &item.Item{ID: "sword1", Name: "Iron Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(item1)

	// Add items to recipient
	item2 := &item.Item{ID: "potion1", Name: "Health Potion", Rarity: item.RarityCommon}
	recipientInv.AddItem(item2)

	// Propose trade
	err := sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{"potion1"})
	if err != nil {
		t.Fatalf("ProposeTrade failed: %v", err)
	}

	// Verify trade component created
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	if proposerTrade.ActiveTrade == nil {
		t.Fatal("Proposer should have active trade")
	}

	if proposerTrade.ActiveTrade.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", proposerTrade.ActiveTrade.Status)
	}

	if len(proposerTrade.ActiveTrade.OfferedItems) != 1 || proposerTrade.ActiveTrade.OfferedItems[0] != "sword1" {
		t.Error("Offered items mismatch")
	}
}

// TestProposeTrade_ProximityFailure verifies proximity check.
func TestProposeTrade_ProximityFailure(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposer.AddComponent(NewInventoryComponent(10, 100.0))

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 10, Y: 0}) // 10 tiles away, >5 limit
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	err := sys.ProposeTrade(proposer.ID, recipient.ID, []string{}, []string{})
	if err == nil {
		t.Fatal("Expected proximity error, got nil")
	}

	// Verify it's a social proximity error
	socialErr, ok := social.IsSocialError(err)
	if !ok {
		t.Fatalf("Expected SocialError, got: %v", err)
	}
	if socialErr.Type != social.ErrorTypeProximity {
		t.Errorf("Expected proximity error type, got: %v", socialErr.Type)
	}
}

// TestProposeTrade_OwnershipValidation verifies item ownership check.
func TestProposeTrade_OwnershipValidation(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	// Try to offer item not owned
	err := sys.ProposeTrade(proposer.ID, recipient.ID, []string{"nonexistent"}, []string{})
	if err == nil {
		t.Fatal("Expected ownership error, got nil")
	}
}

// TestProposeTrade_TrustLimits verifies trust-based restrictions.
func TestProposeTrade_TrustLimits(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	proposerTrade.TrustScore = 0.2 // Low trust

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	// Add legendary item
	legendary := &item.Item{
		ID:     "legendary1",
		Name:   "Excalibur",
		Rarity: item.RarityLegendary,
	}
	proposerInv.AddItem(legendary)

	// Try to trade legendary with low trust
	err := sys.ProposeTrade(proposer.ID, recipient.ID, []string{"legendary1"}, []string{})
	if err == nil {
		t.Fatal("Expected trust limit error, got nil")
	}
}

// TestProposeTrade_LowTrustItemLimit verifies 5-item limit for low trust.
func TestProposeTrade_LowTrustItemLimit(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(20, 200.0)
	proposer.AddComponent(proposerInv)
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	proposerTrade.TrustScore = 0.2 // Low trust

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	// Add 6 common items
	itemIDs := []string{}
	for i := 0; i < 6; i++ {
		itm := &item.Item{
			ID:     fmt.Sprintf("item%d", i),
			Name:   fmt.Sprintf("Item %d", i),
			Rarity: item.RarityCommon,
		}
		proposerInv.AddItem(itm)
		itemIDs = append(itemIDs, itm.ID)
	}

	// Try to trade 6 items with low trust (limit is 5)
	err := sys.ProposeTrade(proposer.ID, recipient.ID, itemIDs, []string{})
	if err == nil {
		t.Fatal("Expected item count error for low trust, got nil")
	}
}

// TestAcceptTrade verifies trade acceptance flow.
func TestAcceptTrade(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)
	item1 := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(item1)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	// Propose trade
	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})

	// Accept trade
	err := sys.AcceptTrade(recipient.ID)
	if err != nil {
		t.Fatalf("AcceptTrade failed: %v", err)
	}

	recipientTrade := sys.getOrCreateTradeComponent(recipient)
	if recipientTrade.ActiveTrade.Status != "accepted" {
		t.Errorf("Expected status 'accepted', got '%s'", recipientTrade.ActiveTrade.Status)
	}
}

// TestRejectTrade verifies trade rejection.
func TestRejectTrade(t *testing.T) {
	world := NewWorld()
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)
	item1 := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(item1)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0) // Commit entities

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})

	// Reject trade
	err := sys.RejectTrade(recipient.ID)
	if err != nil {
		t.Fatalf("RejectTrade failed: %v", err)
	}

	// Verify both parties have no active trade
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	recipientTrade := sys.getOrCreateTradeComponent(recipient)

	if proposerTrade.ActiveTrade != nil {
		t.Error("Proposer should have no active trade after rejection")
	}
	if recipientTrade.ActiveTrade != nil {
		t.Error("Recipient should have no active trade after rejection")
	}
}

// TestCommitTrade_Success verifies successful trade commit.
func TestCommitTrade_Success(t *testing.T) {
	world := NewWorld()
	world.Update(0.0) // Commit pending entities
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipientInv := NewInventoryComponent(10, 100.0)
	recipient.AddComponent(recipientInv)

	world.Update(0.0) // Commit entities

	// Add items
	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	potion := &item.Item{ID: "potion1", Name: "Potion", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)
	recipientInv.AddItem(potion)

	// Propose and accept
	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{"potion1"})
	sys.AcceptTrade(recipient.ID)

	// Commit trade
	err := sys.CommitTrade(proposer.ID)
	if err != nil {
		t.Fatalf("CommitTrade failed: %v", err)
	}

	// Verify item transfer
	if len(proposerInv.Items) != 1 || proposerInv.Items[0].ID != "potion1" {
		t.Error("Proposer should have potion after trade")
	}
	if len(recipientInv.Items) != 1 || recipientInv.Items[0].ID != "sword1" {
		t.Error("Recipient should have sword after trade")
	}

	// Verify trust increased
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	if proposerTrade.TrustScore != TrustDefault+TrustIncrement {
		t.Errorf("Expected trust %.2f, got %.2f", TrustDefault+TrustIncrement, proposerTrade.TrustScore)
	}

	// Verify trade history
	if len(proposerTrade.TradeHistory) != 1 || !proposerTrade.TradeHistory[0].Success {
		t.Error("Trade history should record successful trade")
	}
}

// TestCommitTrade_ProximityViolation verifies rollback on proximity violation.
func TestCommitTrade_ProximityViolation(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposerPos := &PositionComponent{X: 0, Y: 0}
	proposer.AddComponent(proposerPos)
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipientPos := &PositionComponent{X: 3, Y: 0}
	recipient.AddComponent(recipientPos)
	recipientInv := NewInventoryComponent(10, 100.0)
	recipient.AddComponent(recipientInv)

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})
	sys.AcceptTrade(recipient.ID)

	// Move recipient too far
	recipientPos.X = 15

	err := sys.CommitTrade(proposer.ID)
	if err == nil {
		t.Fatal("Expected proximity violation error, got nil")
	}

	// Verify trust decreased
	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	if proposerTrade.TrustScore >= TrustDefault {
		t.Error("Trust should decrease after failed trade")
	}
}

// TestCommitTrade_ItemNoLongerOwned verifies rollback when item is moved.
func TestCommitTrade_ItemNoLongerOwned(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})
	sys.AcceptTrade(recipient.ID)

	// Remove the sword (dropped/sold)
	proposerInv.RemoveItem(0)

	err := sys.CommitTrade(proposer.ID)
	if err == nil {
		t.Fatal("Expected ownership error, got nil")
	}
}

// TestConcurrentTrades verifies that concurrent trades for same item fail.
func TestConcurrentTrades(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	player1 := world.CreateEntity()
	player1.AddComponent(&PositionComponent{X: 0, Y: 0})
	inv1 := NewInventoryComponent(10, 100.0)
	player1.AddComponent(inv1)

	player2 := world.CreateEntity()
	player2.AddComponent(&PositionComponent{X: 3, Y: 0})
	player2.AddComponent(NewInventoryComponent(10, 100.0))

	player3 := world.CreateEntity()
	player3.AddComponent(&PositionComponent{X: 3, Y: 3})
	player3.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	inv1.AddItem(sword)

	// Player1 proposes trade with Player2
	err1 := sys.ProposeTrade(player1.ID, player2.ID, []string{"sword1"}, []string{})
	if err1 != nil {
		t.Fatalf("First proposal failed: %v", err1)
	}

	// Player1 tries to propose concurrent trade with Player3 (should fail)
	err2 := sys.ProposeTrade(player1.ID, player3.ID, []string{"sword1"}, []string{})
	if err2 == nil {
		t.Fatal("Concurrent trade should fail (player already has active trade)")
	}
}

// TestCancelTrade verifies trade cancellation.
func TestCancelTrade(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})

	err := sys.CancelTrade(proposer.ID)
	if err != nil {
		t.Fatalf("CancelTrade failed: %v", err)
	}

	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	if proposerTrade.ActiveTrade != nil {
		t.Error("Active trade should be cleared after cancel")
	}
}

// TestUpdate_Timeout verifies trade timeout handling.
func TestUpdate_Timeout(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)
	proposerTrade := sys.getOrCreateTradeComponent(proposer)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})

	// Manually set proposal time to past
	proposerTrade.ActiveTrade.ProposalTime = time.Now().Unix() - TradeTimeout - 1

	// Update should detect timeout
	sys.Update(1.0)

	if proposerTrade.ActiveTrade != nil {
		t.Error("Timeout should clear active trade")
	}

	// Trust should decrease
	if proposerTrade.TrustScore >= TrustDefault {
		t.Error("Trust should decrease after timeout")
	}
}

// TestUpdate_ProximityDuringNegotiation verifies proximity monitoring.
func TestUpdate_ProximityDuringNegotiation(t *testing.T) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposerPos := &PositionComponent{X: 0, Y: 0}
	proposer.AddComponent(proposerPos)
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipientPos := &PositionComponent{X: 3, Y: 0}
	recipient.AddComponent(recipientPos)
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})

	// Move recipient too far during negotiation
	recipientPos.X = 15

	sys.Update(1.0)

	proposerTrade := sys.getOrCreateTradeComponent(proposer)
	if proposerTrade.ActiveTrade != nil {
		t.Error("Proximity violation during negotiation should cancel trade")
	}
}

// BenchmarkProposeTrade benchmarks trade proposal.
func BenchmarkProposeTrade(b *testing.B) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	proposer := world.CreateEntity()
	proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
	proposerInv := NewInventoryComponent(10, 100.0)
	proposer.AddComponent(proposerInv)

	recipient := world.CreateEntity()
	recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
	recipient.AddComponent(NewInventoryComponent(10, 100.0))

	world.Update(0.0)

	sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
	proposerInv.AddItem(sword)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})
		sys.CancelTrade(proposer.ID)
	}
}

// BenchmarkCommitTrade benchmarks trade commit.
func BenchmarkCommitTrade(b *testing.B) {
	world := NewWorld()
	world.Update(0.0)
	sys := NewTradeSystem(world)

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		proposer := world.CreateEntity()
		proposer.AddComponent(&PositionComponent{X: 0, Y: 0})
		proposerInv := NewInventoryComponent(10, 100.0)
		proposer.AddComponent(proposerInv)

		recipient := world.CreateEntity()
		recipient.AddComponent(&PositionComponent{X: 3, Y: 0})
		recipientInv := NewInventoryComponent(10, 100.0)
		recipient.AddComponent(recipientInv)

		world.Update(0.0)

		sword := &item.Item{ID: "sword1", Name: "Sword", Rarity: item.RarityCommon}
		proposerInv.AddItem(sword)

		sys.ProposeTrade(proposer.ID, recipient.ID, []string{"sword1"}, []string{})
		sys.AcceptTrade(recipient.ID)

		b.StartTimer()
		sys.CommitTrade(proposer.ID)
	}
}
