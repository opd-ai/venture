// Package trade provides item trading between players.
package trade

import (
	"fmt"
	
	"github.com/opd-ai/venture/pkg/engine"
)

// TradeSystem manages item trading
type TradeSystem struct {
	world *engine.World
}

// NewTradeSystem creates a new trade system
func NewTradeSystem(world *engine.World) *TradeSystem {
	return &TradeSystem{world: world}
}

// Update processes active trades
func (s *TradeSystem) Update(deltaTime float64) {
	// Trade is event-driven
}

// ProposeTrade proposes a trade between two players
func (s *TradeSystem) ProposeTrade(proposerID, recipientID uint64, offeredItems, requestedItems []uint64) error {
	// Validate proximity (TODO: Add distance check)
	
	// Get proposer's trade component
	proposer, ok := s.world.GetEntity(proposerID)
	if !ok || proposer == nil {
		return fmt.Errorf("proposer not found")
	}
	
	tradeCompRaw, ok := proposer.GetComponent("trade")
	if !ok {
		// Create trade component
		tradeComp := &engine.TradeComponent{
			ActiveTrade:  nil,
			TradeHistory: []engine.TradeRecord{},
			TrustScore:   0.5,
		}
		proposer.AddComponent(tradeComp)
		tradeCompRaw, _ = proposer.GetComponent("trade")
	}
	
	tradeComp := tradeCompRaw.(*engine.TradeComponent)
	
	// Check for existing trade
	if tradeComp.ActiveTrade != nil {
		return fmt.Errorf("proposer already has an active trade")
	}
	
	// Create proposal
	proposal := &engine.TradeProposal{
		ProposerID:     proposerID,
		RecipientID:    recipientID,
		OfferedItems:   offeredItems,
		RequestedItems: requestedItems,
		Status:         "pending",
	}
	
	tradeComp.ActiveTrade = proposal
	
	// TODO: Notify recipient
	// TODO: Implement two-phase commit
	// TODO: Validate ownership
	// TODO: Check trust score
	
	return nil
}

// AcceptTrade accepts a trade proposal
func (s *TradeSystem) AcceptTrade(recipientID uint64) error {
	// TODO: Implement two-phase commit
	// TODO: Atomic ownership transfer
	// TODO: Update trust scores
	return nil
}

// RejectTrade rejects a trade proposal
func (s *TradeSystem) RejectTrade(recipientID uint64) error {
	// TODO: Cancel trade
	return nil
}
