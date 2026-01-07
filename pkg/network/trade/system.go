// Package trade provides item trading between players with two-phase commit,
// proximity validation, trust mechanics, and atomic ownership transfer.
package trade

import (
	"fmt"
	"math"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/validation"
)

const (
	// Trade proximity and distance constraints
	maxProposalDistance = 5.0  // Maximum distance to initiate trade (tiles)
	maxTradeDistance    = 10.0 // Maximum distance during trade (auto-cancel if exceeded)

	// Trust score thresholds
	highTrustThreshold = 0.8 // Can trade rare/legendary items
	lowTrustThreshold  = 0.3 // Limited to common/uncommon, max 5 items
	maxLowTrustItems   = 5   // Maximum items for low-trust trades

	// Trust score adjustments
	trustSuccessBonus   = 0.05 // Trust increase on successful trade
	trustFailurePenalty = 0.1  // Trust decrease on failed trade
	defaultTrustScore   = 0.5  // Default trust score for new players

	// Trade timeouts
	proposalTimeout = 30 * time.Second // Auto-cancel if no response
)

// TradeStatus represents the state of a trade
type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"   // Waiting for recipient response
	TradeStatusAccepted  TradeStatus = "accepted"  // Recipient accepted, validating
	TradeStatusRejected  TradeStatus = "rejected"  // Recipient rejected
	TradeStatusCommitted TradeStatus = "committed" // Items transferred successfully
	TradeStatusCancelled TradeStatus = "cancelled" // Cancelled (timeout, disconnect, distance)
	TradeStatusFailed    TradeStatus = "failed"    // Validation failed
)

// TradeFailureReason describes why a trade failed
type TradeFailureReason string

const (
	ReasonProximity  TradeFailureReason = "players too far apart"
	ReasonTrust      TradeFailureReason = "insufficient trust score"
	ReasonOwnership  TradeFailureReason = "item no longer owned"
	ReasonInventory  TradeFailureReason = "inventory space insufficient"
	ReasonTradable   TradeFailureReason = "item not tradable"
	ReasonTimeout    TradeFailureReason = "trade proposal expired"
	ReasonConcurrent TradeFailureReason = "item in another trade"
	ReasonDisconnect TradeFailureReason = "player disconnected"
	ReasonRarity     TradeFailureReason = "rarity exceeds trust level"

	// TradeRateLimit is the maximum number of trade requests per player per second
	// This prevents spam and DoS attacks while allowing normal trading
	TradeRateLimit = 10

	// TradeRateLimitWindow is the time window for trade rate limiting
	TradeRateLimitWindow = time.Second
)

// TradeSystem manages item trading with two-phase commit protocol and input validation
type TradeSystem struct {
	world     *engine.World
	validator *validation.TradeValidator
	limiter   *validation.RateLimiter
}

// NewTradeSystem creates a new trade system with validation and rate limiting
func NewTradeSystem(world *engine.World) *TradeSystem {
	return &TradeSystem{
		world:     world,
		validator: validation.NewTradeValidator(),
		limiter:   validation.NewRateLimiter(TradeRateLimit, TradeRateLimitWindow),
	}
}

// Update processes active trades (check timeouts, proximity)
func (s *TradeSystem) Update(deltaTime float64) {
	if s.world == nil {
		return
	}

	// Get all entities with trade components
	entities := s.world.GetEntitiesWith("trade")
	now := time.Now()

	for _, entity := range entities {
		tradeComp := s.getTradeComponent(entity)
		if tradeComp == nil || tradeComp.ActiveTrade == nil {
			continue
		}

		proposal := tradeComp.ActiveTrade

		// Check for timeout
		if proposal.Status == string(TradeStatusPending) {
			proposalTime := time.Unix(proposal.ProposalTime, 0)
			if now.Sub(proposalTime) > proposalTimeout {
				s.cancelTrade(entity.ID, ReasonTimeout)
				continue
			}
		}

		// Check proximity for active trades
		if proposal.Status == string(TradeStatusPending) || proposal.Status == string(TradeStatusAccepted) {
			if !s.validateProximity(proposal.ProposerID, proposal.RecipientID, maxTradeDistance) {
				s.cancelTrade(entity.ID, ReasonProximity)
			}
		}
	}
}

// ProposeTrade proposes a trade between two players with validation and rate limiting
func (s *TradeSystem) ProposeTrade(proposerID, recipientID uint64, offeredItemIDs, requestedItemIDs []string) error {
	if s.world == nil {
		return fmt.Errorf("world is nil")
	}

	// Check rate limit
	if !s.limiter.Allow(proposerID) {
		return fmt.Errorf("rate limit exceeded (maximum 10 trade requests per second)")
	}

	// Validate item IDs format before processing
	if err := s.validator.ValidateTradeRequest(offeredItemIDs, requestedItemIDs); err != nil {
		return fmt.Errorf("trade validation failed: %w", err)
	}

	proposer, recipient, err := s.getTradeEntities(proposerID, recipientID)
	if err != nil {
		return err
	}

	if !s.validateProximity(proposerID, recipientID, maxProposalDistance) {
		return fmt.Errorf("players too far apart (max %v tiles)", maxProposalDistance)
	}

	proposerTradeComp, recipientTradeComp, err := s.getOrCreateTradeComponents(proposer, recipient)
	if err != nil {
		return err
	}

	proposerInv, recipientInv, err := s.getInventories(proposer, recipient)
	if err != nil {
		return err
	}

	_, _, err = s.resolveAndValidateTradeItems(
		proposerInv, recipientInv, offeredItemIDs, requestedItemIDs,
		proposerTradeComp.TrustScore, recipientTradeComp.TrustScore)
	if err != nil {
		return err
	}

	proposal := s.createTradeProposal(proposerID, recipientID, offeredItemIDs, requestedItemIDs)
	proposerTradeComp.ActiveTrade = proposal
	recipientTradeComp.ActiveTrade = proposal

	return nil
}

// getTradeEntities retrieves and validates proposer and recipient entities.
func (s *TradeSystem) getTradeEntities(proposerID, recipientID uint64) (*engine.Entity, *engine.Entity, error) {
	proposer, ok := s.world.GetEntity(proposerID)
	if !ok || proposer == nil {
		return nil, nil, fmt.Errorf("proposer not found")
	}

	recipient, ok := s.world.GetEntity(recipientID)
	if !ok || recipient == nil {
		return nil, nil, fmt.Errorf("recipient not found")
	}

	return proposer, recipient, nil
}

// getOrCreateTradeComponents gets or creates trade components for both participants.
func (s *TradeSystem) getOrCreateTradeComponents(proposer, recipient *engine.Entity) (*engine.TradeComponent, *engine.TradeComponent, error) {
	proposerTradeComp := s.getOrCreateTradeComponent(proposer)
	if proposerTradeComp.ActiveTrade != nil {
		return nil, nil, fmt.Errorf("proposer already has an active trade")
	}

	recipientTradeComp := s.getOrCreateTradeComponent(recipient)
	if recipientTradeComp.ActiveTrade != nil {
		return nil, nil, fmt.Errorf("recipient already has an active trade")
	}

	return proposerTradeComp, recipientTradeComp, nil
}

// resolveAndValidateTradeItems resolves item IDs and validates trust/tradability.
func (s *TradeSystem) resolveAndValidateTradeItems(proposerInv, recipientInv *engine.InventoryComponent,
	offeredItemIDs, requestedItemIDs []string, proposerTrust, recipientTrust float64,
) ([]*item.Item, []*item.Item, error) {
	offeredItemRefs, err := s.resolveItems(proposerInv, offeredItemIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("offered items invalid: %w", err)
	}

	requestedItemRefs, err := s.resolveItems(recipientInv, requestedItemIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("requested items invalid: %w", err)
	}

	if err := s.validateTrust(proposerTrust, offeredItemRefs); err != nil {
		return nil, nil, fmt.Errorf("trust validation failed for offered items: %w", err)
	}

	if err := s.validateTrust(recipientTrust, requestedItemRefs); err != nil {
		return nil, nil, fmt.Errorf("trust validation failed for requested items: %w", err)
	}

	if err := s.validateTradability(offeredItemRefs); err != nil {
		return nil, nil, fmt.Errorf("offered items validation failed: %w", err)
	}

	if err := s.validateTradability(requestedItemRefs); err != nil {
		return nil, nil, fmt.Errorf("requested items validation failed: %w", err)
	}

	return offeredItemRefs, requestedItemRefs, nil
}

// createTradeProposal creates a new trade proposal with current timestamp.
func (s *TradeSystem) createTradeProposal(proposerID, recipientID uint64, offeredItems, requestedItems []string) *engine.TradeProposal {
	now := time.Now()
	return &engine.TradeProposal{
		ProposerID:     proposerID,
		RecipientID:    recipientID,
		OfferedItems:   offeredItems,
		RequestedItems: requestedItems,
		Status:         string(TradeStatusPending),
		ProposalTime:   now.Unix(),
		FailureReason:  "",
	}
}

// AcceptTrade accepts a trade proposal and commits the transfer
func (s *TradeSystem) AcceptTrade(recipientID uint64) error {
	if s.world == nil {
		return fmt.Errorf("world is nil")
	}

	recipient, ok := s.world.GetEntity(recipientID)
	if !ok || recipient == nil {
		return fmt.Errorf("recipient not found")
	}

	tradeComp := s.getTradeComponent(recipient)
	if tradeComp == nil || tradeComp.ActiveTrade == nil {
		return fmt.Errorf("no active trade proposal")
	}

	proposal := tradeComp.ActiveTrade
	if proposal.Status != string(TradeStatusPending) {
		return fmt.Errorf("trade not in pending state: %s", proposal.Status)
	}

	// Change status to accepted
	proposal.Status = string(TradeStatusAccepted)

	// Validate and commit the trade
	if err := s.commitTrade(proposal.ProposerID, recipientID); err != nil {
		proposal.Status = string(TradeStatusFailed)
		proposal.FailureReason = err.Error()
		// Clear trade from both participants to prevent stuck state
		s.clearTrade(proposal.ProposerID)
		s.clearTrade(recipientID)
		return fmt.Errorf("trade commit failed: %w", err)
	}

	proposal.Status = string(TradeStatusCommitted)
	return nil
}

// RejectTrade rejects a trade proposal
func (s *TradeSystem) RejectTrade(recipientID uint64) error {
	if s.world == nil {
		return fmt.Errorf("world is nil")
	}

	recipient, ok := s.world.GetEntity(recipientID)
	if !ok || recipient == nil {
		return fmt.Errorf("recipient not found")
	}

	tradeComp := s.getTradeComponent(recipient)
	if tradeComp == nil || tradeComp.ActiveTrade == nil {
		return fmt.Errorf("no active trade proposal")
	}

	proposal := tradeComp.ActiveTrade
	proposal.Status = string(TradeStatusRejected)

	// Clear trade from both participants
	s.clearTrade(proposal.ProposerID)
	s.clearTrade(recipientID)

	return nil
}

// commitTrade performs two-phase commit for atomic item transfer
func (s *TradeSystem) commitTrade(proposerID, recipientID uint64) error {
	proposer, recipient, proposal, err := s.validateTradeParticipants(proposerID, recipientID)
	if err != nil {
		return err
	}

	proposerInv, recipientInv, err := s.getInventories(proposer, recipient)
	if err != nil {
		return err
	}

	offeredItems, requestedItems, err := s.resolveAndValidateItems(proposerInv, recipientInv, proposal)
	if err != nil {
		return err
	}

	if err := s.executeItemTransfer(proposerInv, recipientInv, offeredItems, requestedItems, proposal); err != nil {
		return err
	}

	s.finalizeTradeSuccess(proposerID, recipientID, proposer, recipient)
	return nil
}

// validateTradeParticipants validates trade participants and proximity.
func (s *TradeSystem) validateTradeParticipants(proposerID, recipientID uint64) (*engine.Entity, *engine.Entity, *engine.TradeProposal, error) {
	proposer, ok := s.world.GetEntity(proposerID)
	if !ok || proposer == nil {
		return nil, nil, nil, fmt.Errorf("%s: proposer not found", ReasonDisconnect)
	}

	recipient, ok := s.world.GetEntity(recipientID)
	if !ok || recipient == nil {
		return nil, nil, nil, fmt.Errorf("%s: recipient not found", ReasonDisconnect)
	}

	proposerTradeComp := s.getTradeComponent(proposer)
	if proposerTradeComp == nil || proposerTradeComp.ActiveTrade == nil {
		return nil, nil, nil, fmt.Errorf("proposer has no active trade")
	}

	if !s.validateProximity(proposerID, recipientID, maxTradeDistance) {
		return nil, nil, nil, fmt.Errorf("%s", ReasonProximity)
	}

	return proposer, recipient, proposerTradeComp.ActiveTrade, nil
}

// getInventories retrieves inventory components for both participants.
func (s *TradeSystem) getInventories(proposer, recipient *engine.Entity) (*engine.InventoryComponent, *engine.InventoryComponent, error) {
	proposerInv := s.getInventoryComponent(proposer)
	recipientInv := s.getInventoryComponent(recipient)

	if proposerInv == nil || recipientInv == nil {
		return nil, nil, fmt.Errorf("missing inventory component")
	}

	return proposerInv, recipientInv, nil
}

// resolveAndValidateItems resolves items and validates inventory space.
func (s *TradeSystem) resolveAndValidateItems(proposerInv, recipientInv *engine.InventoryComponent, proposal *engine.TradeProposal) ([]*item.Item, []*item.Item, error) {
	offeredItems, err := s.resolveItems(proposerInv, proposal.OfferedItems)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", ReasonOwnership, err)
	}

	requestedItems, err := s.resolveItems(recipientInv, proposal.RequestedItems)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", ReasonOwnership, err)
	}

	if err := s.validateInventorySpace(proposerInv, requestedItems, offeredItems); err != nil {
		return nil, nil, fmt.Errorf("%s (proposer): %w", ReasonInventory, err)
	}

	if err := s.validateInventorySpace(recipientInv, offeredItems, requestedItems); err != nil {
		return nil, nil, fmt.Errorf("%s (recipient): %w", ReasonInventory, err)
	}

	return offeredItems, requestedItems, nil
}

// executeItemTransfer performs the atomic item transfer between participants.
// Tracks removed items so they can be properly restored on rollback.
func (s *TradeSystem) executeItemTransfer(proposerInv, recipientInv *engine.InventoryComponent, offeredItems, requestedItems []*item.Item, proposal *engine.TradeProposal) error {
	// Track items removed during transfer for proper rollback
	var removedFromProposer []*item.Item
	var removedFromRecipient []*item.Item
	var addedToRecipient []*item.Item
	var addedToProposer []*item.Item

	// Helper to restore items on failure
	rollback := func() {
		// Restore items removed from proposer
		for _, itm := range removedFromProposer {
			proposerInv.AddItem(itm)
		}
		// Restore items removed from recipient
		for _, itm := range removedFromRecipient {
			recipientInv.AddItem(itm)
		}
		// Remove items added to recipient (reverse of add)
		for _, itm := range addedToRecipient {
			recipientInv.RemoveItemByReference(itm)
		}
		// Remove items added to proposer (reverse of add)
		for _, itm := range addedToProposer {
			proposerInv.RemoveItemByReference(itm)
		}
	}

	// Phase 1: Remove offered items from proposer
	for _, itm := range offeredItems {
		if !proposerInv.RemoveItemByReference(itm) {
			rollback()
			return fmt.Errorf("%s: failed to remove item from proposer", ReasonOwnership)
		}
		removedFromProposer = append(removedFromProposer, itm)
	}

	// Phase 2: Remove requested items from recipient
	for _, itm := range requestedItems {
		if !recipientInv.RemoveItemByReference(itm) {
			rollback()
			return fmt.Errorf("%s: failed to remove item from recipient", ReasonOwnership)
		}
		removedFromRecipient = append(removedFromRecipient, itm)
	}

	// Phase 3: Add offered items to recipient
	for _, itm := range offeredItems {
		if !recipientInv.AddItem(itm) {
			rollback()
			return fmt.Errorf("%s: failed to add item to recipient", ReasonInventory)
		}
		addedToRecipient = append(addedToRecipient, itm)
	}

	// Phase 4: Add requested items to proposer
	for _, itm := range requestedItems {
		if !proposerInv.AddItem(itm) {
			rollback()
			return fmt.Errorf("%s: failed to add item to proposer", ReasonInventory)
		}
		addedToProposer = append(addedToProposer, itm)
	}

	return nil
}

// finalizeTradeSuccess updates trust scores and clears trade state.
func (s *TradeSystem) finalizeTradeSuccess(proposerID, recipientID uint64, proposer, recipient *engine.Entity) {
	proposerTradeComp := s.getTradeComponent(proposer)
	recipientTradeComp := s.getTradeComponent(recipient)

	s.updateTrustScore(proposerTradeComp, true)
	s.updateTrustScore(recipientTradeComp, true)

	s.recordTrade(proposerTradeComp, recipientID, true)
	s.recordTrade(recipientTradeComp, proposerID, true)

	s.clearTrade(proposerID)
	s.clearTrade(recipientID)
}

// cancelTrade cancels an active trade with a reason
func (s *TradeSystem) cancelTrade(entityID uint64, reason TradeFailureReason) error {
	if s.world == nil {
		return fmt.Errorf("world is nil")
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return fmt.Errorf("entity not found")
	}

	tradeComp := s.getTradeComponent(entity)
	if tradeComp == nil || tradeComp.ActiveTrade == nil {
		return fmt.Errorf("no active trade")
	}

	proposal := tradeComp.ActiveTrade
	proposal.Status = string(TradeStatusCancelled)
	proposal.FailureReason = string(reason)

	// Update trust scores (failed trade)
	proposerEntity, ok := s.world.GetEntity(proposal.ProposerID)
	if ok && proposerEntity != nil {
		if pc := s.getTradeComponent(proposerEntity); pc != nil {
			s.updateTrustScore(pc, false)
			s.recordTrade(pc, proposal.RecipientID, false)
		}
	}

	recipientEntity, ok := s.world.GetEntity(proposal.RecipientID)
	if ok && recipientEntity != nil {
		if rc := s.getTradeComponent(recipientEntity); rc != nil {
			s.updateTrustScore(rc, false)
			s.recordTrade(rc, proposal.ProposerID, false)
		}
	}

	// Clear trades
	s.clearTrade(proposal.ProposerID)
	s.clearTrade(proposal.RecipientID)

	return nil
}

// Helper functions

func (s *TradeSystem) validateProximity(id1, id2 uint64, maxDist float64) bool {
	if s.world == nil {
		return false
	}

	e1, ok1 := s.world.GetEntity(id1)
	e2, ok2 := s.world.GetEntity(id2)

	if !ok1 || !ok2 || e1 == nil || e2 == nil {
		return false
	}

	pos1 := s.getPosition(e1)
	pos2 := s.getPosition(e2)

	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= maxDist
}

func (s *TradeSystem) getPosition(entity *engine.Entity) engine.PositionComponent {
	if comp, ok := entity.GetComponent("position"); ok {
		if pos, ok := comp.(*engine.PositionComponent); ok {
			return *pos
		}
	}
	return engine.PositionComponent{X: 0, Y: 0}
}

func (s *TradeSystem) getTradeComponent(entity *engine.Entity) *engine.TradeComponent {
	if comp, ok := entity.GetComponent("trade"); ok {
		if trade, ok := comp.(*engine.TradeComponent); ok {
			return trade
		}
	}
	return nil
}

func (s *TradeSystem) getOrCreateTradeComponent(entity *engine.Entity) *engine.TradeComponent {
	tradeComp := s.getTradeComponent(entity)
	if tradeComp == nil {
		tradeComp = &engine.TradeComponent{
			ActiveTrade:  nil,
			TradeHistory: []engine.TradeRecord{},
			TrustScore:   defaultTrustScore,
		}
		entity.AddComponent(tradeComp)
	}
	return tradeComp
}

func (s *TradeSystem) getInventoryComponent(entity *engine.Entity) *engine.InventoryComponent {
	if comp, ok := entity.GetComponent("inventory"); ok {
		if inv, ok := comp.(*engine.InventoryComponent); ok {
			return inv
		}
	}
	return nil
}

func (s *TradeSystem) resolveItems(inventory *engine.InventoryComponent, itemIDs []string) ([]*item.Item, error) {
	// Check for duplicate item IDs to prevent double-trading the same item
	seenIDs := make(map[string]bool)
	for _, id := range itemIDs {
		if seenIDs[id] {
			return nil, fmt.Errorf("duplicate item ID in trade request: %s", id)
		}
		seenIDs[id] = true
	}

	var items []*item.Item
	for _, id := range itemIDs {
		found := false
		for _, itm := range inventory.Items {
			if itm.ID == id {
				items = append(items, itm)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("item %s not found in inventory", id)
		}
	}
	return items, nil
}

func (s *TradeSystem) validateTrust(trustScore float64, items []*item.Item) error {
	if trustScore < lowTrustThreshold {
		// Low trust: max 5 items, common/uncommon only
		if len(items) > maxLowTrustItems {
			return fmt.Errorf("low trust players can trade max %d items (trust: %.2f)", maxLowTrustItems, trustScore)
		}
		for _, itm := range items {
			if itm.Rarity == item.RarityRare || itm.Rarity == item.RarityEpic || itm.Rarity == item.RarityLegendary {
				return fmt.Errorf("low trust players cannot trade rare/epic/legendary items (trust: %.2f)", trustScore)
			}
		}
	} else if trustScore < highTrustThreshold {
		// Medium trust: no legendary items
		for _, itm := range items {
			if itm.Rarity == item.RarityLegendary {
				return fmt.Errorf("trust score %.2f too low for legendary items (need %.2f)", trustScore, highTrustThreshold)
			}
		}
	}
	// High trust: no restrictions
	return nil
}

func (s *TradeSystem) validateTradability(items []*item.Item) error {
	for _, itm := range items {
		// Check for untradable item types
		// Currently all items are tradable except quest items (which we'll check via tags)
		for _, tag := range itm.Tags {
			if tag == "quest" || tag == "bound" || tag == "unique" {
				return fmt.Errorf("item %s cannot be traded (tag: %s)", itm.Name, tag)
			}
		}
	}
	return nil
}

func (s *TradeSystem) validateInventorySpace(inventory *engine.InventoryComponent, incomingItems, outgoingItems []*item.Item) error {
	// Calculate net change in item count and weight
	netItemChange := len(incomingItems) - len(outgoingItems)

	// Calculate total weight change
	incomingWeight := 0.0
	for _, itm := range incomingItems {
		incomingWeight += itm.Stats.Weight
	}
	outgoingWeight := 0.0
	for _, itm := range outgoingItems {
		outgoingWeight += itm.Stats.Weight
	}
	netWeightChange := incomingWeight - outgoingWeight

	// Check if resulting inventory would exceed slot limits
	resultingItems := len(inventory.Items) + netItemChange
	if resultingItems > inventory.MaxItems {
		return fmt.Errorf("insufficient inventory slots (%d/%d used, net change +%d)",
			len(inventory.Items), inventory.MaxItems, netItemChange)
	}

	// Check if resulting inventory would exceed weight limits
	resultingWeight := inventory.GetCurrentWeight() + netWeightChange
	if resultingWeight > inventory.MaxWeight {
		return fmt.Errorf("insufficient inventory weight capacity (%.2f/%.2f used, net change +%.2f)",
			inventory.GetCurrentWeight(), inventory.MaxWeight, netWeightChange)
	}

	return nil
}

func (s *TradeSystem) updateTrustScore(tradeComp *engine.TradeComponent, success bool) {
	if success {
		tradeComp.TrustScore += trustSuccessBonus
		if tradeComp.TrustScore > 1.0 {
			tradeComp.TrustScore = 1.0
		}
	} else {
		tradeComp.TrustScore -= trustFailurePenalty
		if tradeComp.TrustScore < 0.0 {
			tradeComp.TrustScore = 0.0
		}
	}
}

func (s *TradeSystem) recordTrade(tradeComp *engine.TradeComponent, partnerID uint64, success bool) {
	record := engine.TradeRecord{
		Timestamp: time.Now().Unix(),
		PartnerID: partnerID,
		Success:   success,
	}
	tradeComp.TradeHistory = append(tradeComp.TradeHistory, record)
}

func (s *TradeSystem) clearTrade(entityID uint64) {
	if s.world == nil {
		return
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return
	}

	if tradeComp := s.getTradeComponent(entity); tradeComp != nil {
		tradeComp.ActiveTrade = nil
	}
}

// GetActiveTrade returns the active trade for an entity, if any
func (s *TradeSystem) GetActiveTrade(entityID uint64) *engine.TradeProposal {
	if s.world == nil {
		return nil
	}

	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return nil
	}

	if tradeComp := s.getTradeComponent(entity); tradeComp != nil {
		return tradeComp.ActiveTrade
	}

	return nil
}

// CancelTrade cancels an active trade (public wrapper)
func (s *TradeSystem) CancelTrade(entityID uint64) error {
	return s.cancelTrade(entityID, TradeFailureReason("user cancelled"))
}
