package trade

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
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

// TestProposeTrade tests trade proposal
func TestProposeTrade(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*engine.World) (proposerID, recipientID uint64)
		offeredItems   []uint64
		requestedItems []uint64
		wantErr        bool
		errMsg         string
	}{
		{
			name: "simple trade proposal",
			setup: func(w *engine.World) (uint64, uint64) {
				proposer := w.CreateEntity()
				recipient := w.CreateEntity()
				return proposer.ID, recipient.ID
			},
			offeredItems:   []uint64{1, 2, 3},
			requestedItems: []uint64{4, 5},
			wantErr:        false,
		},
		{
			name: "empty offered items",
			setup: func(w *engine.World) (uint64, uint64) {
				proposer := w.CreateEntity()
				recipient := w.CreateEntity()
				return proposer.ID, recipient.ID
			},
			offeredItems:   []uint64{},
			requestedItems: []uint64{1, 2},
			wantErr:        false,
		},
		{
			name: "empty requested items",
			setup: func(w *engine.World) (uint64, uint64) {
				proposer := w.CreateEntity()
				recipient := w.CreateEntity()
				return proposer.ID, recipient.ID
			},
			offeredItems:   []uint64{1, 2},
			requestedItems: []uint64{},
			wantErr:        false,
		},
		{
			name: "both empty",
			setup: func(w *engine.World) (uint64, uint64) {
				proposer := w.CreateEntity()
				recipient := w.CreateEntity()
				return proposer.ID, recipient.ID
			},
			offeredItems:   []uint64{},
			requestedItems: []uint64{},
			wantErr:        false,
		},
		{
			name: "proposer not found",
			setup: func(w *engine.World) (uint64, uint64) {
				recipient := w.CreateEntity()
				return 9999, recipient.ID // non-existent proposer
			},
			offeredItems:   []uint64{1},
			requestedItems: []uint64{2},
			wantErr:        true,
			errMsg:         "proposer not found",
		},
		{
			name: "proposer has existing trade",
			setup: func(w *engine.World) (uint64, uint64) {
				proposer := w.CreateEntity()
				recipient := w.CreateEntity()
				return proposer.ID, recipient.ID
			},
			offeredItems:   []uint64{1},
			requestedItems: []uint64{2},
			wantErr:        false, // First proposal succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := engine.NewWorld()
			ts := NewTradeSystem(world)

			proposerID, recipientID := tt.setup(world)
			// Process pending entity additions
			world.Update(0)

			err := ts.ProposeTrade(proposerID, recipientID, tt.offeredItems, tt.requestedItems)

			if (err != nil) != tt.wantErr {
				t.Errorf("ProposeTrade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("error message = %q, want %q", err.Error(), tt.errMsg)
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
					if len(proposal.OfferedItems) != len(tt.offeredItems) {
						t.Errorf("offered items count = %d, want %d", len(proposal.OfferedItems), len(tt.offeredItems))
					}
					if len(proposal.RequestedItems) != len(tt.requestedItems) {
						t.Errorf("requested items count = %d, want %d", len(proposal.RequestedItems), len(tt.requestedItems))
					}
					if proposal.Status != "pending" {
						t.Errorf("status = %q, want pending", proposal.Status)
					}
				}
			}
		})
	}
}

// TestProposeTradeWithExistingTrade tests proposing when trade already exists
func TestProposeTradeWithExistingTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := world.CreateEntity()
	recipient1 := world.CreateEntity()
	recipient2 := world.CreateEntity()
	world.Update(0)

	// First trade
	err := ts.ProposeTrade(proposer.ID, recipient1.ID, []uint64{1}, []uint64{2})
	if err != nil {
		t.Fatalf("first ProposeTrade() failed: %v", err)
	}

	// Second trade should fail
	err = ts.ProposeTrade(proposer.ID, recipient2.ID, []uint64{3}, []uint64{4})
	if err == nil {
		t.Error("expected error for existing trade, got nil")
	} else if err.Error() != "proposer already has an active trade" {
		t.Errorf("error = %q, want proposer already has an active trade", err.Error())
	}
}

// TestTradeComponentCreation tests automatic trade component creation
func TestTradeComponentCreation(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := world.CreateEntity()
	recipient := world.CreateEntity()
	world.Update(0)

	// Initially, no trade component
	_, ok := proposer.GetComponent("trade")
	if ok {
		t.Error("trade component should not exist initially")
	}

	// Propose trade should create component
	err := ts.ProposeTrade(proposer.ID, recipient.ID, []uint64{1}, []uint64{2})
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	// Now should have trade component
	tradeCompRaw, ok := proposer.GetComponent("trade")
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

	recipient := world.CreateEntity()
	world.Update(0)

	// AcceptTrade currently not implemented (returns nil)
	err := ts.AcceptTrade(recipient.ID)
	if err != nil {
		t.Errorf("AcceptTrade() error = %v, want nil", err)
	}
}

// TestRejectTrade tests rejecting a trade
func TestRejectTrade(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	recipient := world.CreateEntity()
	world.Update(0)

	// RejectTrade currently not implemented (returns nil)
	err := ts.RejectTrade(recipient.ID)
	if err != nil {
		t.Errorf("RejectTrade() error = %v, want nil", err)
	}
}

// TestTradeProposalFields tests trade proposal structure
func TestTradeProposalFields(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := world.CreateEntity()
	recipient := world.CreateEntity()
	world.Update(0)

	offeredItems := []uint64{10, 20, 30}
	requestedItems := []uint64{40, 50}

	err := ts.ProposeTrade(proposer.ID, recipient.ID, offeredItems, requestedItems)
	if err != nil {
		t.Fatalf("ProposeTrade() failed: %v", err)
	}

	tradeCompRaw, _ := proposer.GetComponent("trade")
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	proposal := tradeComp.ActiveTrade

	// Verify all fields
	if proposal.ProposerID != proposer.ID {
		t.Errorf("ProposerID = %v, want %v", proposal.ProposerID, proposer.ID)
	}
	if proposal.RecipientID != recipient.ID {
		t.Errorf("RecipientID = %v, want %v", proposal.RecipientID, recipient.ID)
	}
	for i, item := range proposal.OfferedItems {
		if item != offeredItems[i] {
			t.Errorf("OfferedItems[%d] = %v, want %v", i, item, offeredItems[i])
		}
	}
	for i, item := range proposal.RequestedItems {
		if item != requestedItems[i] {
			t.Errorf("RequestedItems[%d] = %v, want %v", i, item, requestedItems[i])
		}
	}
	if proposal.Status != "pending" {
		t.Errorf("Status = %q, want pending", proposal.Status)
	}
}

// TestTradeWithLargeItemLists tests trading with many items
func TestTradeWithLargeItemLists(t *testing.T) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	proposer := world.CreateEntity()
	recipient := world.CreateEntity()
	world.Update(0)

	// Create large item lists
	offeredItems := make([]uint64, 100)
	requestedItems := make([]uint64, 100)
	for i := 0; i < 100; i++ {
		offeredItems[i] = uint64(i)
		requestedItems[i] = uint64(i + 1000)
	}

	err := ts.ProposeTrade(proposer.ID, recipient.ID, offeredItems, requestedItems)
	if err != nil {
		t.Fatalf("ProposeTrade() with large lists failed: %v", err)
	}

	tradeCompRaw, _ := proposer.GetComponent("trade")
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	proposal := tradeComp.ActiveTrade

	if len(proposal.OfferedItems) != 100 {
		t.Errorf("offered items count = %d, want 100", len(proposal.OfferedItems))
	}
	if len(proposal.RequestedItems) != 100 {
		t.Errorf("requested items count = %d, want 100", len(proposal.RequestedItems))
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

	proposers := make([]*engine.Entity, b.N)
	recipients := make([]*engine.Entity, b.N)
	for i := 0; i < b.N; i++ {
		proposers[i] = world.CreateEntity()
		recipients[i] = world.CreateEntity()
	}
	world.Update(0)

	offeredItems := []uint64{1, 2, 3}
	requestedItems := []uint64{4, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.ProposeTrade(proposers[i].ID, recipients[i].ID, offeredItems, requestedItems)
	}
}

// BenchmarkProposeTradeWithExistingComponent benchmarks when component exists
func BenchmarkProposeTradeWithExistingComponent(b *testing.B) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)

	// Create entities with trade components
	proposers := make([]*engine.Entity, b.N)
	recipients := make([]*engine.Entity, b.N)
	for i := 0; i < b.N; i++ {
		proposers[i] = world.CreateEntity()
		recipients[i] = world.CreateEntity()
		proposers[i].AddComponent(&engine.TradeComponent{
			ActiveTrade:  nil,
			TradeHistory: []engine.TradeRecord{},
			TrustScore:   0.5,
		})
	}
	world.Update(0)

	offeredItems := []uint64{1, 2, 3}
	requestedItems := []uint64{4, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.ProposeTrade(proposers[i].ID, recipients[i].ID, offeredItems, requestedItems)
	}
}

// BenchmarkAcceptTrade benchmarks accepting a trade
func BenchmarkAcceptTrade(b *testing.B) {
	world := engine.NewWorld()
	ts := NewTradeSystem(world)
	recipient := world.CreateEntity()
	world.Update(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts.AcceptTrade(recipient.ID)
	}
}
