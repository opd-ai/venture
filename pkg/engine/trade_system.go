// Package engine provides the trade system for player-to-player item transfers.
// This system implements two-phase commit protocol with proximity validation,
// trust-based limits, and atomic ownership transfer.
package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/social"
)

const (
	// Trade proximity constants
	ProposalProximity = 5.0  // Max distance to initiate trade
	ActiveProximity   = 10.0 // Max distance during negotiation

	// Trust thresholds
	TrustHigh      = 0.8
	TrustLow       = 0.3
	TrustDefault   = 0.5
	TrustIncrement = 0.05 // Successful trade
	TrustDecrement = 0.1  // Failed trade

	// Trade limits
	LowTrustMaxItems = 5
	TradeTimeout     = 30 // seconds
)

// TradeSystem handles item trading between players with proximity validation,
// trust mechanics, and two-phase commit protocol.
type TradeSystem struct {
	world *World
}

// NewTradeSystem creates a new trade system.
func NewTradeSystem(world *World) *TradeSystem {
	return &TradeSystem{world: world}
}

// ProposeTrade initiates a trade between two players.
// Returns error if proximity, trust, or ownership validation fails.
func (ts *TradeSystem) ProposeTrade(proposerID, recipientID uint64, offeredItemIDs, requestedItemIDs []string) error {
	proposer, recipient, err := ts.validateProposalParticipants(proposerID, recipientID)
	if err != nil {
		return err
	}

	proposerTrade := ts.getOrCreateTradeComponent(proposer)
	recipientTrade := ts.getOrCreateTradeComponent(recipient)

	if err := ts.validateOfferedItems(proposer, proposerTrade, offeredItemIDs); err != nil {
		return err
	}

	if err := ts.validateRequestedItems(recipient, recipientTrade, requestedItemIDs); err != nil {
		return err
	}

	ts.createTradeProposal(proposerID, recipientID, proposerTrade, recipientTrade, offeredItemIDs, requestedItemIDs)
	return nil
}

// validateProposalParticipants validates that both entities exist and are in proximity.
func (ts *TradeSystem) validateProposalParticipants(proposerID, recipientID uint64) (*Entity, *Entity, error) {
	proposer, _ := ts.world.GetEntity(proposerID)
	recipient, _ := ts.world.GetEntity(recipientID)

	if proposer == nil || recipient == nil {
		return nil, nil, fmt.Errorf("invalid entity IDs")
	}

	if !ts.checkProximity(proposer, recipient, ProposalProximity) {
		distance := ts.calculateDistance(proposer, recipient)
		return nil, nil, social.ErrProximity(ProposalProximity, distance)
	}

	return proposer, recipient, nil
}

// validateOfferedItems validates that proposer owns and can trade offered items.
func (ts *TradeSystem) validateOfferedItems(proposer *Entity, proposerTrade *TradeComponent, offeredItemIDs []string) error {
	if proposerTrade.ActiveTrade != nil {
		return fmt.Errorf("proposer already has an active trade")
	}

	proposerInv := ts.getInventoryComponent(proposer)
	if proposerInv == nil {
		return fmt.Errorf("proposer has no inventory")
	}

	for _, itemID := range offeredItemIDs {
		if !ts.ownsItem(proposerInv, itemID) {
			return social.ErrOwnership(itemID)
		}
	}

	if err := ts.checkTrustLimits(proposerTrade.TrustScore, proposerInv, offeredItemIDs); err != nil {
		return fmt.Errorf("proposer trust check failed: %w", err)
	}

	return nil
}

// validateRequestedItems validates that recipient owns and can trade requested items.
func (ts *TradeSystem) validateRequestedItems(recipient *Entity, recipientTrade *TradeComponent, requestedItemIDs []string) error {
	if recipientTrade.ActiveTrade != nil {
		return fmt.Errorf("recipient already has an active trade")
	}

	if len(requestedItemIDs) == 0 {
		return nil
	}

	recipientInv := ts.getInventoryComponent(recipient)
	if recipientInv == nil {
		return fmt.Errorf("recipient has no inventory")
	}

	for _, itemID := range requestedItemIDs {
		if !ts.ownsItem(recipientInv, itemID) {
			return social.ErrOwnership(itemID)
		}
	}

	if err := ts.checkTrustLimits(recipientTrade.TrustScore, recipientInv, requestedItemIDs); err != nil {
		return fmt.Errorf("recipient trust check failed: %w", err)
	}

	return nil
}

// createTradeProposal creates and assigns a trade proposal to both participants.
func (ts *TradeSystem) createTradeProposal(proposerID, recipientID uint64, proposerTrade, recipientTrade *TradeComponent, offeredItemIDs, requestedItemIDs []string) {
	offeredLineItems := make([]TradeLineItem, 0, len(offeredItemIDs))
	for _, itemID := range offeredItemIDs {
		offeredLineItems = append(offeredLineItems, TradeLineItem{ItemID: itemID, Quantity: 1})
	}
	requestedLineItems := make([]TradeLineItem, 0, len(requestedItemIDs))
	for _, itemID := range requestedItemIDs {
		requestedLineItems = append(requestedLineItems, TradeLineItem{ItemID: itemID, Quantity: 1})
	}

	proposal := &TradeProposal{
		ProposerID:         proposerID,
		RecipientID:        recipientID,
		OfferedItems:       offeredItemIDs,
		RequestedItems:     requestedItemIDs,
		OfferedLineItems:   offeredLineItems,
		RequestedLineItems: requestedLineItems,
		Status:             "pending",
		ProposalTime:       time.Now().Unix(),
	}

	proposerTrade.ActiveTrade = proposal
	recipientTrade.ActiveTrade = proposal
}

// AcceptTrade accepts a pending trade proposal.
// Returns error if validation fails or proximity violated.
func (ts *TradeSystem) AcceptTrade(recipientID uint64) error {
	recipient, _ := ts.world.GetEntity(recipientID)
	if recipient == nil {
		return fmt.Errorf("invalid recipient ID")
	}

	recipientTrade := ts.getOrCreateTradeComponent(recipient)
	if recipientTrade.ActiveTrade == nil {
		return fmt.Errorf("no active trade")
	}

	if recipientTrade.ActiveTrade.Status != "pending" {
		return fmt.Errorf("trade not in pending status: %s", recipientTrade.ActiveTrade.Status)
	}

	recipientTrade.ActiveTrade.Status = "accepted"
	return nil
}

// RejectTrade rejects a pending trade proposal.
func (ts *TradeSystem) RejectTrade(recipientID uint64) error {
	recipient, _ := ts.world.GetEntity(recipientID)
	if recipient == nil {
		return fmt.Errorf("invalid recipient ID")
	}

	recipientTrade := ts.getOrCreateTradeComponent(recipient)
	if recipientTrade.ActiveTrade == nil {
		return fmt.Errorf("no active trade")
	}

	proposal := recipientTrade.ActiveTrade
	proposal.Status = "rejected"

	// Clear trade from both parties
	ts.clearTrade(proposal.ProposerID, proposal.RecipientID)

	return nil
}

// CommitTrade executes an accepted trade with full validation and atomic transfer.
// This is the server-authoritative commit phase.
func (ts *TradeSystem) CommitTrade(proposalEntityID uint64) error {
	proposal, err := ts.validateTradeCommit(proposalEntityID)
	if err != nil {
		return err
	}

	proposer, recipient, proposerInv, recipientInv, err := ts.validateTradeParticipants(proposal)
	if err != nil {
		return err
	}

	offeredItems, requestedItems, err := ts.executeAtomicTransfer(proposal, proposerInv, recipientInv)
	if err != nil {
		return err
	}

	ts.transferItemsToNewOwners(offeredItems, requestedItems, proposerInv, recipientInv)
	ts.updateTradeRecords(proposal, proposer, recipient)

	proposal.Status = "committed"
	ts.clearTrade(proposal.ProposerID, proposal.RecipientID)

	return nil
}

// validateTradeCommit validates the initial trade proposal and returns it if valid.
func (ts *TradeSystem) validateTradeCommit(proposalEntityID uint64) (*TradeProposal, error) {
	entity, _ := ts.world.GetEntity(proposalEntityID)
	if entity == nil {
		return nil, fmt.Errorf("invalid entity ID")
	}

	tradeComp := ts.getOrCreateTradeComponent(entity)
	if tradeComp.ActiveTrade == nil {
		return nil, fmt.Errorf("no active trade")
	}

	proposal := tradeComp.ActiveTrade
	if proposal.Status != "accepted" {
		return nil, fmt.Errorf("trade not accepted: %s", proposal.Status)
	}

	return proposal, nil
}

// validateTradeParticipants validates entities, proximity, inventories, and item ownership.
func (ts *TradeSystem) validateTradeParticipants(proposal *TradeProposal) (*Entity, *Entity, *InventoryComponent, *InventoryComponent, error) {
	proposer, _ := ts.world.GetEntity(proposal.ProposerID)
	recipient, _ := ts.world.GetEntity(proposal.RecipientID)

	if proposer == nil || recipient == nil {
		return nil, nil, nil, nil, ts.rollbackTrade(proposal, "entity no longer exists")
	}

	if !ts.checkProximity(proposer, recipient, ActiveProximity) {
		return nil, nil, nil, nil, ts.rollbackTrade(proposal, fmt.Sprintf("proximity violated (>%.1f tiles)", ActiveProximity))
	}

	proposerInv := ts.getInventoryComponent(proposer)
	recipientInv := ts.getInventoryComponent(recipient)

	if proposerInv == nil || recipientInv == nil {
		return nil, nil, nil, nil, ts.rollbackTrade(proposal, "inventory missing")
	}

	if err := ts.validateItemOwnership(proposal, proposerInv, recipientInv); err != nil {
		return nil, nil, nil, nil, err
	}

	return proposer, recipient, proposerInv, recipientInv, nil
}

// validateItemOwnership verifies that both parties still own their items.
func (ts *TradeSystem) validateItemOwnership(proposal *TradeProposal, proposerInv, recipientInv *InventoryComponent) error {
	for _, itemID := range proposal.OfferedItems {
		if !ts.ownsItem(proposerInv, itemID) {
			return ts.rollbackTrade(proposal, fmt.Sprintf("proposer no longer owns: %s", itemID))
		}
	}

	for _, itemID := range proposal.RequestedItems {
		if !ts.ownsItem(recipientInv, itemID) {
			return ts.rollbackTrade(proposal, fmt.Sprintf("recipient no longer owns: %s", itemID))
		}
	}

	return nil
}

// executeAtomicTransfer removes items from both inventories atomically with rollback on failure.
func (ts *TradeSystem) executeAtomicTransfer(proposal *TradeProposal, proposerInv, recipientInv *InventoryComponent) ([]*item.Item, []*item.Item, error) {
	var offeredItems []*item.Item
	var requestedItems []*item.Item

	for _, itemID := range proposal.OfferedItems {
		itm := ts.removeItemByID(proposerInv, itemID)
		if itm == nil {
			ts.rollbackItemRemoval(offeredItems, proposerInv)
			return nil, nil, ts.rollbackTrade(proposal, fmt.Sprintf("failed to remove offered item: %s", itemID))
		}
		offeredItems = append(offeredItems, itm)
	}

	for _, itemID := range proposal.RequestedItems {
		itm := ts.removeItemByID(recipientInv, itemID)
		if itm == nil {
			ts.rollbackItemRemoval(offeredItems, proposerInv)
			ts.rollbackItemRemoval(requestedItems, recipientInv)
			return nil, nil, ts.rollbackTrade(proposal, fmt.Sprintf("failed to remove requested item: %s", itemID))
		}
		requestedItems = append(requestedItems, itm)
	}

	return offeredItems, requestedItems, nil
}

// rollbackItemRemoval restores items to an inventory during rollback.
func (ts *TradeSystem) rollbackItemRemoval(items []*item.Item, inv *InventoryComponent) {
	for _, prevItem := range items {
		inv.AddItem(prevItem)
	}
}

// transferItemsToNewOwners adds items to their new owners' inventories.
func (ts *TradeSystem) transferItemsToNewOwners(offeredItems, requestedItems []*item.Item, proposerInv, recipientInv *InventoryComponent) {
	for _, itm := range offeredItems {
		if !recipientInv.AddItem(itm) {
			proposerInv.AddItem(itm)
		}
	}

	for _, itm := range requestedItems {
		if !proposerInv.AddItem(itm) {
			recipientInv.AddItem(itm)
		}
	}
}

// updateTradeRecords updates trust scores and trade history for both participants.
func (ts *TradeSystem) updateTradeRecords(proposal *TradeProposal, proposer, recipient *Entity) {
	proposerTrade := ts.getOrCreateTradeComponent(proposer)
	recipientTrade := ts.getOrCreateTradeComponent(recipient)

	proposerTrade.TrustScore = clamp(proposerTrade.TrustScore+TrustIncrement, 0.0, 1.0)
	recipientTrade.TrustScore = clamp(recipientTrade.TrustScore+TrustIncrement, 0.0, 1.0)

	record := TradeRecord{
		Timestamp: time.Now().Unix(),
		PartnerID: proposal.RecipientID,
		Success:   true,
	}
	proposerTrade.TradeHistory = append(proposerTrade.TradeHistory, record)

	recipientRecord := TradeRecord{
		Timestamp: time.Now().Unix(),
		PartnerID: proposal.ProposerID,
		Success:   true,
	}
	recipientTrade.TradeHistory = append(recipientTrade.TradeHistory, recipientRecord)
}

// CancelTrade cancels an active trade (can be called by either party).
func (ts *TradeSystem) CancelTrade(entityID uint64) error {
	entity, _ := ts.world.GetEntity(entityID)
	if entity == nil {
		return fmt.Errorf("invalid entity ID")
	}

	tradeComp := ts.getOrCreateTradeComponent(entity)
	if tradeComp.ActiveTrade == nil {
		return fmt.Errorf("no active trade")
	}

	proposal := tradeComp.ActiveTrade
	proposal.Status = "cancelled"

	ts.clearTrade(proposal.ProposerID, proposal.RecipientID)
	return nil
}

// Update checks for trade timeouts and proximity violations.
func (ts *TradeSystem) Update(deltaTime float64) {
	// Iterate through all entities with trade components
	entities := ts.world.GetEntitiesWith("trade")

	for _, entity := range entities {
		comp, ok := entity.GetComponent("trade")
		if !ok {
			continue
		}
		tradeComp, ok := comp.(*TradeComponent)
		if !ok || tradeComp.ActiveTrade == nil {
			continue
		}

		proposal := tradeComp.ActiveTrade

		// Check timeout
		elapsed := time.Now().Unix() - proposal.ProposalTime
		if elapsed > TradeTimeout {
			ts.rollbackTrade(proposal, "timeout exceeded")
			continue
		}

		// Check proximity for active trades
		proposer, _ := ts.world.GetEntity(proposal.ProposerID)
		recipient, _ := ts.world.GetEntity(proposal.RecipientID)

		if proposer != nil && recipient != nil {
			if !ts.checkProximity(proposer, recipient, ActiveProximity) {
				ts.rollbackTrade(proposal, "proximity violated during negotiation")
			}
		}
	}
}

// Helper functions

func (ts *TradeSystem) checkProximity(e1, e2 *Entity, maxDistance float64) bool {
	pos1Comp, ok1 := e1.GetComponent("position")
	pos2Comp, ok2 := e2.GetComponent("position")

	if !ok1 || !ok2 {
		return false
	}

	pos1, ok1 := pos1Comp.(*PositionComponent)
	pos2, ok2 := pos2Comp.(*PositionComponent)

	if !ok1 || !ok2 {
		return false
	}

	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= maxDistance
}

// calculateDistance returns the distance between two entities.
func (ts *TradeSystem) calculateDistance(e1, e2 *Entity) float64 {
	pos1Comp, ok1 := e1.GetComponent("position")
	pos2Comp, ok2 := e2.GetComponent("position")

	if !ok1 || !ok2 {
		return math.MaxFloat64
	}

	pos1, ok1 := pos1Comp.(*PositionComponent)
	pos2, ok2 := pos2Comp.(*PositionComponent)

	if !ok1 || !ok2 {
		return math.MaxFloat64
	}

	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (ts *TradeSystem) getOrCreateTradeComponent(entity *Entity) *TradeComponent {
	comp, ok := entity.GetComponent("trade")
	if ok {
		if tc, ok := comp.(*TradeComponent); ok {
			return tc
		}
	}

	// Create new trade component with default trust
	tc := &TradeComponent{
		TrustScore:   TrustDefault,
		TradeHistory: make([]TradeRecord, 0),
	}
	entity.AddComponent(tc)
	return tc
}

func (ts *TradeSystem) getInventoryComponent(entity *Entity) *InventoryComponent {
	comp, ok := entity.GetComponent("inventory")
	if !ok {
		return nil
	}
	inv, ok := comp.(*InventoryComponent)
	if !ok {
		return nil
	}
	return inv
}

func (ts *TradeSystem) ownsItem(inv *InventoryComponent, itemID string) bool {
	for _, itm := range inv.Items {
		if itm.ID == itemID {
			return true
		}
	}
	return false
}

func (ts *TradeSystem) removeItemByID(inv *InventoryComponent, itemID string) *item.Item {
	for i, itm := range inv.Items {
		if itm.ID == itemID {
			return inv.RemoveItem(i)
		}
	}
	return nil
}

func (ts *TradeSystem) checkTrustLimits(trustScore float64, inv *InventoryComponent, itemIDs []string) error {
	// Low trust players can only trade common/uncommon items
	if trustScore < TrustLow {
		// Check item count limit
		if len(itemIDs) > LowTrustMaxItems {
			return social.ErrTrust(TrustLow, trustScore).
				WithContext("reason", "item_count_exceeded").
				WithContext("max_items", LowTrustMaxItems)
		}

		// Check rarity limits
		for _, itemID := range itemIDs {
			itm := ts.getItemByID(inv, itemID)
			if itm == nil {
				continue
			}

			// Legendary and Epic forbidden for low trust
			if itm.Rarity == item.RarityLegendary || itm.Rarity == item.RarityEpic {
				return social.ErrTrust(TrustLow, trustScore).
					WithContext("reason", "rarity_restricted").
					WithContext("item_rarity", itm.Rarity.String())
			}
		}
	}

	// High trust can trade anything (no restrictions)

	return nil
}

func (ts *TradeSystem) getItemByID(inv *InventoryComponent, itemID string) *item.Item {
	for _, itm := range inv.Items {
		if itm.ID == itemID {
			return itm
		}
	}
	return nil
}

func (ts *TradeSystem) rollbackTrade(proposal *TradeProposal, reason string) error {
	proposal.Status = "failed"
	proposal.FailureReason = reason

	// Decrease trust scores
	proposer, _ := ts.world.GetEntity(proposal.ProposerID)
	recipient, _ := ts.world.GetEntity(proposal.RecipientID)

	if proposer != nil {
		proposerTrade := ts.getOrCreateTradeComponent(proposer)
		proposerTrade.TrustScore = clamp(proposerTrade.TrustScore-TrustDecrement, 0.0, 1.0)

		// Record failed trade
		proposerTrade.TradeHistory = append(proposerTrade.TradeHistory, TradeRecord{
			Timestamp: time.Now().Unix(),
			PartnerID: proposal.RecipientID,
			Success:   false,
		})
	}

	if recipient != nil {
		recipientTrade := ts.getOrCreateTradeComponent(recipient)
		recipientTrade.TrustScore = clamp(recipientTrade.TrustScore-TrustDecrement, 0.0, 1.0)

		// Record failed trade
		recipientTrade.TradeHistory = append(recipientTrade.TradeHistory, TradeRecord{
			Timestamp: time.Now().Unix(),
			PartnerID: proposal.ProposerID,
			Success:   false,
		})
	}

	ts.clearTrade(proposal.ProposerID, proposal.RecipientID)

	return fmt.Errorf("trade rolled back: %s", reason)
}

func (ts *TradeSystem) clearTrade(proposerID, recipientID uint64) {
	proposer, _ := ts.world.GetEntity(proposerID)
	recipient, _ := ts.world.GetEntity(recipientID)

	if proposer != nil {
		if comp, ok := proposer.GetComponent("trade"); ok {
			if tc, ok := comp.(*TradeComponent); ok {
				tc.ActiveTrade = nil
			}
		}
	}

	if recipient != nil {
		if comp, ok := recipient.GetComponent("trade"); ok {
			if tc, ok := comp.(*TradeComponent); ok {
				tc.ActiveTrade = nil
			}
		}
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
