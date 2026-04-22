package trade

import (
	"fmt"
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestNewTradeSystem tests creation of trade system
func TestNewTradeSystem(t *testing.T) {
	tests := []struct {
		name  string
		world *engine.World
	}{
		{"valid world", engine.NewWorld()},
		{"nil world", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTradeSystem(tt.world)
			if ts == nil {
				t.Fatal("NewTradeSystem returned nil")
			}
			if ts.world != tt.world {
				t.Error("world not set correctly")
			}
		})
	}
}

// TestTradeSystemUpdate tests trade system update
func TestTradeSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	tests := []struct {
		name      string
		deltaTime float64
	}{
		{"zero delta", 0.0},
		{"normal delta", 0.016},
		{"large delta", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			ts.Update(tt.deltaTime)
		})
	}
}

// Helper function to create a test entity with position and inventory
func createTestPlayer(world *engine.World, x, y float64, items []*item.Item) uint64 {
	entity := world.CreateEntity()
	entity.AddComponent(&engine.PositionComponent{X: x, Y: y})

	inv := engine.NewInventoryComponent(100, 1000.0)
	for _, itm := range items {
		inv.AddItem(itm)
	}
	entity.AddComponent(inv)

	return entity.ID
}

// Helper function to create a test item
func createTestItem(id, name string, rarity item.Rarity) *item.Item {
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

func countItemsByID(items []*item.Item, itemID string) int {
	count := 0
	for _, itm := range items {
		if itm.ID == itemID {
			count++
		}
	}
	return count
}

// TestProposeTrade tests trade proposal
func TestProposeTrade(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*engine.World) (proposerID, recipientID uint64, offeredItems, requestedItems []string)
		wantErr bool
		errMsg  string
	}{
		{
			name: "simple trade proposal",
			setup: func(w *engine.World) (uint64, uint64, []string, []string) {
				item1 := createTestItem("item1", "Sword", item.RarityCommon)
				item2 := createTestItem("item2", "Shield", item.RarityCommon)
				item3 := createTestItem("item3", "Potion", item.RarityCommon)

				proposer := createTestPlayer(w, 0, 0, []*item.Item{item1, item2})
				recipient := createTestPlayer(w, 1, 1, []*item.Item{item3})

				w.Update(0) // Process entity additions

				return proposer, recipient, []string{"item1"}, []string{"item3"}
			},
			wantErr: false,
		},
		{
			name: "empty offered items",
			setup: func(w *engine.World) (uint64, uint64, []string, []string) {
				item1 := createTestItem("item1", "Potion", item.RarityCommon)

				proposer := createTestPlayer(w, 0, 0, []*item.Item{})
				recipient := createTestPlayer(w, 1, 1, []*item.Item{item1})

				w.Update(0)

				return proposer, recipient, []string{}, []string{"item1"}
			},
			wantErr: false,
		},
		{
			name: "proposer not found",
			setup: func(w *engine.World) (uint64, uint64, []string, []string) {
				item1 := createTestItem("item1", "Item", item.RarityCommon)
				recipient := createTestPlayer(w, 0, 0, []*item.Item{item1})
				w.Update(0)

				return 9999, recipient, []string{}, []string{"item1"}
			},
			wantErr: true,
			errMsg:  "proposer not found",
		},
		{
			name: "players too far apart",
			setup: func(w *engine.World) (uint64, uint64, []string, []string) {
				item1 := createTestItem("item1", "Sword", item.RarityCommon)
				item2 := createTestItem("item2", "Shield", item.RarityCommon)

				// Place players >5 tiles apart (max proposal distance)
				proposer := createTestPlayer(w, 0, 0, []*item.Item{item1})
				recipient := createTestPlayer(w, 100, 100, []*item.Item{item2})

				w.Update(0)

				return proposer, recipient, []string{"item1"}, []string{"item2"}
			},
			wantErr: true,
			errMsg:  "too far apart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			ts := NewTradeSystem(world)

			proposerID, recipientID, offeredItems, requestedItems := tt.setup(world)

			err := ts.ProposeTrade(proposerID, recipientID, offeredItems, requestedItems)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProposeTrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}

			// If successful, verify trade proposal was created
			if !tt.wantErr {
				proposer, ok := world.GetEntity(proposerID)
				if !ok {
					t.Fatal("proposer not found after trade proposal")
				}

				tradeCompRaw, ok := proposer.GetComponent("trade")
				if !ok {
					t.Fatal("trade component not found")
				}

				tradeComp := tradeCompRaw.(*engine.TradeComponent)
				if tradeComp.ActiveTrade == nil {
					t.Error("no active trade")
				} else {
					proposal := tradeComp.ActiveTrade
					if proposal.ProposerID != proposerID {
						t.Errorf("proposer ID = %v, want %v", proposal.ProposerID, proposerID)
					}
					if proposal.RecipientID != recipientID {
						t.Errorf("recipient ID = %v, want %v", proposal.RecipientID, recipientID)
					}
					if len(proposal.OfferedItems) != len(offeredItems) {
						t.Errorf("offered items count = %d, want %d", len(proposal.OfferedItems), len(offeredItems))
					}
					if len(proposal.RequestedItems) != len(requestedItems) {
						t.Errorf("requested items count = %d, want %d", len(proposal.RequestedItems), len(requestedItems))
					}
					if proposal.Status != "pending" {
						t.Errorf("status = %q, want pending", proposal.Status)
					}
				}
			}
		})
	}
}

func TestProposeTradeWithQuantities(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{
		createTestItem("stack_potion", "Potion A", item.RarityCommon),
		createTestItem("stack_potion", "Potion B", item.RarityCommon),
	})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{
		createTestItem("stack_herb", "Herb A", item.RarityCommon),
		createTestItem("stack_herb", "Herb B", item.RarityCommon),
	})
	world.Update(0)

	err := ts.ProposeTradeWithQuantities(
		proposer,
		recipient,
		[]engine.TradeLineItem{{ItemID: "stack_potion", Quantity: 2}},
		[]engine.TradeLineItem{{ItemID: "stack_herb", Quantity: 2}},
	)
	if err != nil {
		t.Fatalf("ProposeTradeWithQuantities() failed: %v", err)
	}

	proposal := ts.GetActiveTrade(proposer)
	if proposal == nil {
		t.Fatal("expected active trade proposal")
	}
	if len(proposal.OfferedLineItems) != 1 || proposal.OfferedLineItems[0].Quantity != 2 {
		t.Fatalf("unexpected offered line items: %+v", proposal.OfferedLineItems)
	}
	if len(proposal.RequestedLineItems) != 1 || proposal.RequestedLineItems[0].Quantity != 2 {
		t.Fatalf("unexpected requested line items: %+v", proposal.RequestedLineItems)
	}

	if err := ts.AcceptTrade(recipient); err != nil {
		t.Fatalf("AcceptTrade() failed: %v", err)
	}

	proposerEntity, _ := world.GetEntity(proposer)
	proposerInvRaw, _ := proposerEntity.GetComponent("inventory")
	proposerInv := proposerInvRaw.(*engine.InventoryComponent)
	if got := countItemsByID(proposerInv.Items, "stack_herb"); got != 2 {
		t.Fatalf("proposer received %d stack_herb items, want 2", got)
	}
}

func TestProposeTradeWithQuantities_RejectsInvalidQuantity(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{
		createTestItem("item1", "Sword", item.RarityCommon),
	})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{
		createTestItem("item2", "Shield", item.RarityCommon),
	})
	world.Update(0)

	err := ts.ProposeTradeWithQuantities(
		proposer,
		recipient,
		[]engine.TradeLineItem{{ItemID: "item1", Quantity: 0}},
		[]engine.TradeLineItem{{ItemID: "item2", Quantity: 1}},
	)
	if err == nil {
		t.Fatal("expected quantity validation error")
	}
	if !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("expected quantity error, got: %v", err)
	}
}

// TestProposeTradeWithExistingTrade tests proposing when trade already exists
func TestProposeTradeWithExistingTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)
	item3 := createTestItem("item3", "Potion", item.RarityCommon)
	item4 := createTestItem("item4", "Axe", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1, item3})
	recipient1 := createTestPlayer(world, 1, 1, []*item.Item{item2})
	recipient2 := createTestPlayer(world, 1.5, 1.5, []*item.Item{item4})
	world.Update(0)

	// First trade
	err := ts.ProposeTrade(proposer, recipient1, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("first ProposeTrade() failed: %v", err)
	}

	// Second trade should fail
	err = ts.ProposeTrade(proposer, recipient2, []string{"item3"}, []string{"item4"})
	if err == nil {
		t.Error("expected error for existing trade, got nil")
	}
}

// TestTradeComponentCreation tests automatic trade component creation
func TestTradeComponentCreation(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item2})
	world.Update(0)

	proposerEntity, _ := world.GetEntity(proposer)

	// Initially, no trade component
	_, ok := proposerEntity.GetComponent("trade")
	if ok {
		t.Error("trade component should not exist initially")
	}

	// Propose trade should create component
	err := ts.ProposeTrade(proposer, recipient, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Now should have trade component
	tradeCompRaw, ok := proposerEntity.GetComponent("trade")
	if !ok {
		t.Fatal("trade component not created")
	}

	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	if tradeComp.ActiveTrade == nil {
		t.Error("no active trade in component")
	}
	if tradeComp.TrustScore != 0.5 {
		t.Errorf("trust score = %v, want 0.5", tradeComp.TrustScore)
	}
	if tradeComp.TradeHistory == nil {
		t.Error("trade history is nil")
	}
}

// TestAcceptTrade tests accepting a trade
func TestAcceptTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item2})
	world.Update(0)

	// Create a trade first
	err := ts.ProposeTrade(proposer, recipient, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Accept the trade
	err = ts.AcceptTrade(recipient)
	if err != nil {
		t.Errorf("AcceptTrade() error = %v", err)
	}

	// Verify trade was completed
	proposerEntity, _ := world.GetEntity(proposer)
	proposerInvRaw, _ := proposerEntity.GetComponent("inventory")
	proposerInv := proposerInvRaw.(*engine.InventoryComponent)

	// Proposer should now have item2
	found := false
	for _, itm := range proposerInv.Items {
		if itm.ID == "item2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("proposer should have received item2")
	}
}

// TestRejectTrade tests rejecting a trade
func TestRejectTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item2})
	world.Update(0)

	// Create a trade first
	err := ts.ProposeTrade(proposer, recipient, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Reject the trade
	err = ts.RejectTrade(recipient)
	if err != nil {
		t.Errorf("RejectTrade() error = %v", err)
	}

	// Verify trade was cancelled
	recipientEntity, _ := world.GetEntity(recipient)
	tradeCompRaw, _ := recipientEntity.GetComponent("trade")
	tradeComp := tradeCompRaw.(*engine.TradeComponent)

	if tradeComp.ActiveTrade != nil {
		t.Error("active trade should be cleared after rejection")
	}
}

// TestTradeProposalFields tests trade proposal structure
func TestTradeProposalFields(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Axe", item.RarityCommon)
	item3 := createTestItem("item3", "Potion", item.RarityCommon)
	item4 := createTestItem("item4", "Shield", item.RarityCommon)
	item5 := createTestItem("item5", "Helmet", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1, item2, item3})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item4, item5})
	world.Update(0)

	offeredItems := []string{"item1", "item2", "item3"}
	requestedItems := []string{"item4", "item5"}

	err := ts.ProposeTrade(proposer, recipient, offeredItems, requestedItems)
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	proposerEntity, _ := world.GetEntity(proposer)
	tradeCompRaw, _ := proposerEntity.GetComponent("trade")
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	proposal := tradeComp.ActiveTrade

	// Verify all fields
	if proposal.ProposerID != proposer {
		t.Errorf("ProposerID = %v, want %v", proposal.ProposerID, proposer)
	}
	if proposal.RecipientID != recipient {
		t.Errorf("RecipientID = %v, want %v", proposal.RecipientID, recipient)
	}
	if len(proposal.OfferedItems) != len(offeredItems) {
		t.Errorf("offered items count = %d, want %d", len(proposal.OfferedItems), len(offeredItems))
	}
	if len(proposal.RequestedItems) != len(requestedItems) {
		t.Errorf("requested items count = %d, want %d", len(proposal.RequestedItems), len(requestedItems))
	}
	if proposal.Status != "pending" {
		t.Errorf("Status = %q, want pending", proposal.Status)
	}
}

// TestTradeWithLargeItemLists tests trading with many items
func TestTradeWithLargeItemLists(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	// Create large item lists
	var offeredItemsData []*item.Item
	var requestedItemsData []*item.Item
	offeredItems := make([]string, 20)
	requestedItems := make([]string, 20)

	for i := 0; i < 20; i++ {
		offeredItems[i] = fmt.Sprintf("offered_%d", i)
		requestedItems[i] = fmt.Sprintf("requested_%d", i)
		offeredItemsData = append(offeredItemsData, createTestItem(offeredItems[i], fmt.Sprintf("Weapon%d", i), item.RarityCommon))
		requestedItemsData = append(requestedItemsData, createTestItem(requestedItems[i], fmt.Sprintf("Armor%d", i), item.RarityCommon))
	}

	proposer := createTestPlayer(world, 0, 0, offeredItemsData)
	recipient := createTestPlayer(world, 1, 1, requestedItemsData)
	world.Update(0)

	err := ts.ProposeTrade(proposer, recipient, offeredItems, requestedItems)
	if err != nil {
		t.Fatalf("ProposeTrade() with large lists failed: %v", err)
	}

	proposerEntity, _ := world.GetEntity(proposer)
	tradeCompRaw, _ := proposerEntity.GetComponent("trade")
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	proposal := tradeComp.ActiveTrade

	if len(proposal.OfferedItems) != 20 {
		t.Errorf("offered items count = %d, want 20", len(proposal.OfferedItems))
	}
	if len(proposal.RequestedItems) != 20 {
		t.Errorf("requested items count = %d, want 20", len(proposal.RequestedItems))
	}
}

// BenchmarkNewTradeSystem benchmarks trade system creation
func BenchmarkNewTradeSystem(b *testing.B) {
	world := engine.NewWorld()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewTradeSystem(world)
	}
}

// BenchmarkProposeTrade benchmarks proposing a trade
func BenchmarkProposeTrade(b *testing.B) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)
	item3 := createTestItem("item3", "Potion", item.RarityCommon)

	proposers := make([]uint64, b.N)
	recipients := make([]uint64, b.N)
	for i := 0; i < b.N; i++ {
		proposers[i] = createTestPlayer(world, float64(i*10), 0, []*item.Item{item1, item2})
		recipients[i] = createTestPlayer(world, float64(i*10)+1, 0, []*item.Item{item3})
	}
	world.Update(0)

	offeredItems := []string{"item1", "item2"}
	requestedItems := []string{"item3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.ProposeTrade(proposers[i], recipients[i], offeredItems, requestedItems)
	}
}

// BenchmarkProposeTradeWithExistingComponent benchmarks when component exists
func BenchmarkProposeTradeWithExistingComponent(b *testing.B) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)
	item3 := createTestItem("item3", "Potion", item.RarityCommon)

	// Create entities with trade components
	proposers := make([]uint64, b.N)
	recipients := make([]uint64, b.N)
	for i := 0; i < b.N; i++ {
		proposers[i] = createTestPlayer(world, float64(i*10), 0, []*item.Item{item1, item2})
		recipients[i] = createTestPlayer(world, float64(i*10)+1, 0, []*item.Item{item3})

		entity, _ := world.GetEntity(proposers[i])
		entity.AddComponent(&engine.TradeComponent{
			ActiveTrade:  nil,
			TradeHistory: []engine.TradeRecord{},
			TrustScore:   0.5,
		})
	}
	world.Update(0)

	offeredItems := []string{"item1", "item2"}
	requestedItems := []string{"item3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.ProposeTrade(proposers[i], recipients[i], offeredItems, requestedItems)
	}
}

// BenchmarkAcceptTrade benchmarks accepting a trade
func BenchmarkAcceptTrade(b *testing.B) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	recipients := make([]uint64, b.N)
	proposers := make([]uint64, b.N)

	for i := 0; i < b.N; i++ {
		proposers[i] = createTestPlayer(world, float64(i*10), 0, []*item.Item{item1})
		recipients[i] = createTestPlayer(world, float64(i*10)+1, 0, []*item.Item{item2})
		ts.ProposeTrade(proposers[i], recipients[i], []string{"item1"}, []string{"item2"})
	}
	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.AcceptTrade(recipients[i])
	}
}

// TestCancelTrade tests trade cancellation
func TestCancelTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item2})
	world.Update(0)

	// Propose trade
	err := ts.ProposeTrade(proposer, recipient, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Cancel trade
	err = ts.CancelTrade(proposer)
	if err != nil {
		t.Errorf("CancelTrade() failed: %v", err)
	}

	// Verify trade is cleared
	proposerEntity, _ := world.GetEntity(proposer)
	tradeCompRaw, ok := proposerEntity.GetComponent("trade")
	if ok {
		tradeComp := tradeCompRaw.(*engine.TradeComponent)
		if tradeComp.ActiveTrade != nil {
			t.Error("active trade should be cleared after cancellation")
		}
	}

	recipientEntity, _ := world.GetEntity(recipient)
	tradeCompRaw, ok = recipientEntity.GetComponent("trade")
	if ok {
		tradeComp := tradeCompRaw.(*engine.TradeComponent)
		if tradeComp.ActiveTrade != nil {
			t.Error("recipient's active trade should be cleared after cancellation")
		}
	}
}

// TestCancelTradeNoActiveTrade tests cancelling when no trade exists
func TestCancelTradeNoActiveTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	player := createTestPlayer(world, 0, 0, []*item.Item{item1})
	world.Update(0)

	// Try to cancel when no trade exists
	err := ts.CancelTrade(player)
	if err == nil {
		t.Error("expected error when cancelling non-existent trade")
	}
}

// TestGetActiveTrade tests getting active trade
func TestGetActiveTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	item1 := createTestItem("item1", "Sword", item.RarityCommon)
	item2 := createTestItem("item2", "Shield", item.RarityCommon)

	proposer := createTestPlayer(world, 0, 0, []*item.Item{item1})
	recipient := createTestPlayer(world, 1, 1, []*item.Item{item2})
	world.Update(0)

	// No trade initially
	proposal := ts.GetActiveTrade(proposer)
	if proposal != nil {
		t.Error("expected nil for no active trade")
	}

	// Propose trade
	err := ts.ProposeTrade(proposer, recipient, []string{"item1"}, []string{"item2"})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Now should have active trade
	proposal = ts.GetActiveTrade(proposer)
	if proposal == nil {
		t.Error("expected active trade proposal")
	} else {
		if proposal.ProposerID != proposer {
			t.Errorf("proposer ID = %d, want %d", proposal.ProposerID, proposer)
		}
		if proposal.RecipientID != recipient {
			t.Errorf("recipient ID = %d, want %d", proposal.RecipientID, recipient)
		}
	}

	// Recipient should also see the trade
	proposal = ts.GetActiveTrade(recipient)
	if proposal == nil {
		t.Error("recipient should see active trade proposal")
	}
}
