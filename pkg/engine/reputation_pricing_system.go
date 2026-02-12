// Package engine provides the ReputationPricingSystem which connects faction reputation
// with merchant pricing. When a player has high reputation with a merchant's faction,
// they receive discounts; when reputation is low or hostile, prices are increased or
// trading is refused entirely.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// ReputationPricingSystem adjusts merchant prices based on player faction reputation.
// It scans merchants and updates their PriceMultiplier based on the player's standing
// with the merchant's faction. This connects FactionComponent reputation data with
// MerchantComponent pricing logic.
type ReputationPricingSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	// updateInterval controls how often prices are recalculated (in seconds)
	updateInterval  float64
	timeSinceUpdate float64

	// Cache player entity to avoid repeated lookups
	cachedPlayerID uint64
}

// NewReputationPricingSystem creates a new reputation pricing system.
func NewReputationPricingSystem(world *World, seed int64) *ReputationPricingSystem {
	var logger *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		logger = world.GetLogger().WithField("system_name", "reputation_pricing")
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"seed": seed,
		}).Debug("Creating reputation pricing system")
	}

	return &ReputationPricingSystem{
		world:           world,
		logger:          logger,
		rng:             rand.New(rand.NewSource(seed)),
		updateInterval:  1.0, // Update every second
		timeSinceUpdate: 0,
		cachedPlayerID:  0,
	}
}

// Update processes merchant pricing adjustments based on player reputation.
func (s *ReputationPricingSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	// Throttle updates for performance
	s.timeSinceUpdate += deltaTime
	if s.timeSinceUpdate < s.updateInterval {
		return
	}
	s.timeSinceUpdate = 0

	// Find player entity
	player := s.findPlayerEntity(entities)
	if player == nil {
		return
	}

	// Get player's faction components (tracks reputation with different factions)
	playerFactions := s.getPlayerFactionReputations(player)
	if len(playerFactions) == 0 {
		return
	}

	// Process merchants and adjust their prices
	merchantsUpdated := 0
	for _, entity := range entities {
		if !entity.HasComponent("merchant") {
			continue
		}

		merchantComp, ok := entity.GetComponent("merchant")
		if !ok {
			continue
		}
		merchant := merchantComp.(*MerchantComponent)

		// Get merchant's faction (if any)
		merchantFactionID := s.getMerchantFaction(entity)
		if merchantFactionID == "" {
			continue // Merchant has no faction, prices unaffected
		}

		// Look up player's reputation with this merchant's faction
		priceMultiplier := s.calculatePriceMultiplier(playerFactions, merchantFactionID)
		if priceMultiplier == 0 {
			continue // Reputation not tracked
		}

		// Apply reputation-based price adjustment
		s.applyPriceModifier(merchant, priceMultiplier, merchantFactionID)
		merchantsUpdated++
	}

	if s.logger != nil && merchantsUpdated > 0 {
		s.logger.WithFields(logrus.Fields{
			"merchants_updated": merchantsUpdated,
			"factions_tracked":  len(playerFactions),
		}).Debug("Updated merchant prices based on reputation")
	}
}

// findPlayerEntity locates the player entity from the entity list.
func (s *ReputationPricingSystem) findPlayerEntity(entities []*Entity) *Entity {
	// Use cached player ID if available
	if s.cachedPlayerID != 0 {
		for _, e := range entities {
			if e.ID == s.cachedPlayerID && e.HasComponent("input") {
				return e
			}
		}
		// Cache invalidated
		s.cachedPlayerID = 0
	}

	// Search for player
	for _, e := range entities {
		if e.HasComponent("input") {
			s.cachedPlayerID = e.ID
			return e
		}
	}
	return nil
}

// getPlayerFactionReputations extracts all faction reputation data from player.
// Returns a map of factionID -> reputation value.
func (s *ReputationPricingSystem) getPlayerFactionReputations(player *Entity) map[string]int {
	reputations := make(map[string]int)

	// Check for faction component
	if factionComp, ok := player.GetComponent("faction"); ok {
		fc, ok := factionComp.(*FactionComponent)
		if ok && fc.IsPlayerFaction {
			reputations[fc.FactionID] = fc.Reputation
		}
	}

	// Also check for faction_reputation component if present (for multiple factions)
	if repComp, ok := player.GetComponent("faction_reputation"); ok {
		if frComp, ok := repComp.(*FactionReputationComponent); ok {
			for factionID, rep := range frComp.Reputations {
				reputations[factionID] = rep
			}
		}
	}

	return reputations
}

// getMerchantFaction returns the faction ID a merchant belongs to.
func (s *ReputationPricingSystem) getMerchantFaction(merchant *Entity) string {
	factionComp, ok := merchant.GetComponent("faction")
	if !ok {
		return ""
	}

	fc, ok := factionComp.(*FactionComponent)
	if !ok || fc.IsPlayerFaction {
		return "" // This is a player faction tracker, not NPC faction membership
	}

	return fc.FactionID
}

// calculatePriceMultiplier computes the price modifier based on reputation.
// Uses the FactionComponent.GetPriceMultiplier() logic for consistency.
func (s *ReputationPricingSystem) calculatePriceMultiplier(playerReps map[string]int, merchantFactionID string) float64 {
	rep, exists := playerReps[merchantFactionID]
	if !exists {
		// No reputation tracked, use neutral pricing
		return 1.0
	}

	// Create a temporary FactionComponent to use its GetPriceMultiplier logic
	fc := &FactionComponent{
		FactionID:       merchantFactionID,
		Reputation:      rep,
		IsPlayerFaction: true,
	}

	return fc.GetPriceMultiplier()
}

// applyPriceModifier updates the merchant's price multiplier based on reputation.
func (s *ReputationPricingSystem) applyPriceModifier(merchant *MerchantComponent, multiplier float64, factionID string) {
	// Base price multiplier is typically 1.5 for merchants
	// We scale by the reputation multiplier
	basePriceMultiplier := 1.5

	// Apply reputation adjustment
	// hostile (0) = refuse to trade (handled at shop UI level)
	// suspicious (1.5) = 50% more expensive * base = 2.25x
	// neutral (1.0) = normal base price = 1.5x
	// friendly (0.75-1.0) = discounted base price = 1.125-1.5x
	newMultiplier := basePriceMultiplier * multiplier

	// Only update if changed significantly (avoid float comparison issues)
	if repPricingAbsFloat(merchant.PriceMultiplier-newMultiplier) > 0.01 {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"faction_id":           factionID,
				"reputation_modifier":  multiplier,
				"old_price_multiplier": merchant.PriceMultiplier,
				"new_price_multiplier": newMultiplier,
			}).Debug("Adjusted merchant price based on reputation")
		}
		merchant.PriceMultiplier = newMultiplier
	}
}

// repPricingAbsFloat returns the absolute value of a float64.
func repPricingAbsFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// FactionReputationComponent tracks a player's reputation with multiple factions.
// This is an alternative to having multiple FactionComponent instances.
type FactionReputationComponent struct {
	// Reputations maps faction IDs to reputation values (-100 to +100)
	Reputations map[string]int
}

// Type returns the component type identifier.
func (f *FactionReputationComponent) Type() string {
	return "faction_reputation"
}

// NewFactionReputationComponent creates a new faction reputation tracker.
func NewFactionReputationComponent() *FactionReputationComponent {
	return &FactionReputationComponent{
		Reputations: make(map[string]int),
	}
}

// GetReputation returns the player's reputation with a faction.
func (f *FactionReputationComponent) GetReputation(factionID string) int {
	if rep, ok := f.Reputations[factionID]; ok {
		return rep
	}
	return 0 // Neutral by default
}

// SetReputation sets the player's reputation with a faction.
func (f *FactionReputationComponent) SetReputation(factionID string, reputation int) {
	// Clamp to valid range
	if reputation < -100 {
		reputation = -100
	}
	if reputation > 100 {
		reputation = 100
	}
	f.Reputations[factionID] = reputation
}

// ModifyReputation adjusts reputation by a delta value.
func (f *FactionReputationComponent) ModifyReputation(factionID string, delta int) {
	current := f.GetReputation(factionID)
	f.SetReputation(factionID, current+delta)
}
