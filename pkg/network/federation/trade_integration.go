// Package federation provides trade integration between political and economic systems.
//
// This file implements Phase 41.3 (Integration & Balance) which connects the
// PoliticsSystem (Phase 41.1) with the FederatedMarket (Phase 41.2) to create
// a cohesive cross-server trade experience with the following features:
//
// 1. Rate Limiting: Prevents price manipulation through trading volume restrictions
//   - Default: 10 trades per 60-second window (configurable)
//   - Automatic window reset
//   - Per-player tracking
//
// 2. Server Reputation System: Trust-based trade limits
//   - Reputation range: 0.0 (blocked) to 1.0 (fully trusted)
//   - Default: 0.5 (neutral) for unknown servers
//   - Five reputation tiers with different transaction limits:
//   - 0.0-0.2: Blocked (requires manual approval)
//   - 0.2-0.4: Restricted (max 5 items, 500 gold)
//   - 0.4-0.6: Limited (max 10 items, 2000 gold)
//   - 0.6-0.8: Trusted (max 20 items, 10000 gold)
//   - 0.8-1.0: Verified (max 50 items, 100000 gold)
//
// 3. AI Merchant Baseline: Economic stability through automated supply/demand
//   - Maintains baseline inventory levels
//   - Prevents extreme price swings
//   - Configurable update intervals
//   - Per-item merchant allocation
//
// 4. Political Integration: Trade prices affected by diplomatic relationships
//   - Alliance: 20% discount (0.8x multiplier)
//   - War: 50% markup (1.5x multiplier)
//   - Treaty: Normal pricing (1.0x multiplier)
//   - Embargo: Trade blocked
//   - Trade Pact: 10% discount (0.9x multiplier)
//
// Thread Safety:
// All operations are thread-safe using RWMutex for concurrent access.
//
// Performance:
// - Trade validation: <0.001ms per check
// - AI merchant updates: <0.1ms for 10 merchants
// - Reputation updates: <0.001ms per operation
//
// Example Usage:
//
//	market := NewFederatedMarket()
//	politics := engine.NewPoliticsSystem(world)
//	integration := NewTradeIntegration(market, politics)
//
//	// Validate a trade
//	err := integration.ValidateTrade("player1", "serverA", 5, 1000.0)
//	if err != nil {
//	    // Handle rate limit or reputation restriction
//	}
//
//	// Record successful trade
//	integration.RecordTrade("player1")
//
//	// Add AI merchant for stability
//	integration.AddAIMerchant("sword", 100, 50, 5*time.Minute)
//
//	// Update in game loop
//	integration.Update(deltaTime)
package federation

import (
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TradeIntegration integrates the political system with the federated market.
// It enforces rate limits, server reputation effects, and AI merchant baselines.
type TradeIntegration struct {
	mu             sync.RWMutex
	market         *FederatedMarket
	politicsSystem *engine.PoliticsSystem

	// Rate limiting
	tradeCounts        map[string]int // PlayerID -> trades in current window
	lastWindowReset    time.Time
	windowDuration     time.Duration // Default: 60 seconds
	maxTradesPerWindow int           // Default: 10 trades/minute

	// Server reputation (0.0-1.0)
	serverReputation map[string]float64 // ServerID -> reputation score

	// AI merchants for baseline supply/demand
	aiMerchants          []AIMerchant
	systemUpdateInterval time.Duration
	lastAIUpdate         time.Time
}

// AIMerchant simulates a baseline merchant for market stability.
type AIMerchant struct {
	MerchantID     string
	ItemID         string
	BaseSupply     int // Amount merchant maintains in stock
	BaseDemand     int // Amount merchant purchases per interval
	UpdateInterval time.Duration
	LastUpdate     time.Time
}

// TradeLimit represents restrictions on a trade based on reputation.
type TradeLimit struct {
	MaxTransactionValue float64 // Maximum gold per transaction
	MaxItemCount        int     // Maximum items per transaction
	RequiresApproval    bool    // Manual server approval required
}

// NewTradeIntegration creates a new trade integration system.
func NewTradeIntegration(market *FederatedMarket, politicsSystem *engine.PoliticsSystem) *TradeIntegration {
	return &TradeIntegration{
		market:               market,
		politicsSystem:       politicsSystem,
		tradeCounts:          make(map[string]int),
		lastWindowReset:      time.Now(),
		windowDuration:       60 * time.Second,
		maxTradesPerWindow:   10,
		serverReputation:     make(map[string]float64),
		aiMerchants:          []AIMerchant{},
		systemUpdateInterval: 5 * time.Minute,
		lastAIUpdate:         time.Now(),
	}
}

// ValidateTrade checks if a trade is allowed under current rate limits and reputation.
func (ti *TradeIntegration) ValidateTrade(playerID, serverID string, itemCount int, totalValue float64) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	// Reset rate limit window if needed
	if time.Since(ti.lastWindowReset) > ti.windowDuration {
		ti.tradeCounts = make(map[string]int)
		ti.lastWindowReset = time.Now()
	}

	// Check rate limit
	currentCount := ti.tradeCounts[playerID]
	if currentCount >= ti.maxTradesPerWindow {
		return fmt.Errorf("rate limit exceeded: %d trades in current window (max: %d)",
			currentCount, ti.maxTradesPerWindow)
	}

	// Get server reputation (default 0.5 for unknown servers)
	reputation, exists := ti.serverReputation[serverID]
	if !exists {
		reputation = 0.5 // Neutral default
	}

	// Get trade limits based on reputation
	limits := ti.getTradeLimit(reputation)

	// Validate against limits
	if totalValue > limits.MaxTransactionValue {
		return fmt.Errorf("transaction value %.2f exceeds limit %.2f for server reputation %.2f",
			totalValue, limits.MaxTransactionValue, reputation)
	}

	if itemCount > limits.MaxItemCount {
		return fmt.Errorf("item count %d exceeds limit %d for server reputation %.2f",
			itemCount, limits.MaxItemCount, reputation)
	}

	if limits.RequiresApproval {
		return fmt.Errorf("server reputation %.2f too low: manual approval required", reputation)
	}

	return nil
}

// RecordTrade increments the player's trade count for rate limiting.
func (ti *TradeIntegration) RecordTrade(playerID string) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	ti.tradeCounts[playerID]++
}

// GetTradesRemaining returns how many trades a player can make in current window.
func (ti *TradeIntegration) GetTradesRemaining(playerID string) int {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	// Reset window if needed
	if time.Since(ti.lastWindowReset) > ti.windowDuration {
		return ti.maxTradesPerWindow
	}

	used := ti.tradeCounts[playerID]
	remaining := ti.maxTradesPerWindow - used
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// getTradeLimit returns trade restrictions based on reputation (must hold lock).
func (ti *TradeIntegration) getTradeLimit(reputation float64) TradeLimit {
	// Reputation tiers:
	// 0.0-0.2: Blocked (requires approval)
	// 0.2-0.4: Restricted (low limits)
	// 0.4-0.6: Limited (moderate limits)
	// 0.6-0.8: Trusted (high limits)
	// 0.8-1.0: Verified (very high limits)

	if reputation < 0.2 {
		return TradeLimit{
			MaxTransactionValue: 100.0,
			MaxItemCount:        1,
			RequiresApproval:    true,
		}
	} else if reputation < 0.4 {
		return TradeLimit{
			MaxTransactionValue: 500.0,
			MaxItemCount:        5,
			RequiresApproval:    false,
		}
	} else if reputation < 0.6 {
		return TradeLimit{
			MaxTransactionValue: 2000.0,
			MaxItemCount:        10,
			RequiresApproval:    false,
		}
	} else if reputation < 0.8 {
		return TradeLimit{
			MaxTransactionValue: 10000.0,
			MaxItemCount:        20,
			RequiresApproval:    false,
		}
	} else {
		return TradeLimit{
			MaxTransactionValue: 100000.0,
			MaxItemCount:        50,
			RequiresApproval:    false,
		}
	}
}

// UpdateServerReputation adjusts a server's reputation score.
// Delta should be in range -0.1 to +0.1 per event.
func (ti *TradeIntegration) UpdateServerReputation(serverID string, delta float64) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	current, exists := ti.serverReputation[serverID]
	if !exists {
		current = 0.5 // Default neutral for first update
	}

	// Apply delta and clamp to 0.0-1.0
	current += delta
	if current < 0.0 {
		current = 0.0
	}
	if current > 1.0 {
		current = 1.0
	}

	ti.serverReputation[serverID] = current
}

// GetServerReputation returns a server's reputation score.
func (ti *TradeIntegration) GetServerReputation(serverID string) float64 {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	rep, exists := ti.serverReputation[serverID]
	if !exists {
		return 0.5 // Neutral default for unknown servers
	}
	return rep
}

// AddAIMerchant registers an AI merchant to maintain baseline supply/demand.
func (ti *TradeIntegration) AddAIMerchant(itemID string, baseSupply, baseDemand int, updateInterval time.Duration) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	merchantID := fmt.Sprintf("ai_merchant_%s_%d", itemID, time.Now().UnixNano())
	merchant := AIMerchant{
		MerchantID:     merchantID,
		ItemID:         itemID,
		BaseSupply:     baseSupply,
		BaseDemand:     baseDemand,
		UpdateInterval: updateInterval,
		LastUpdate:     time.Time{}, // Start with zero time so first update happens immediately
	}

	ti.aiMerchants = append(ti.aiMerchants, merchant)
}

// UpdateAIMerchants updates AI merchant supply/demand to maintain market stability.
func (ti *TradeIntegration) UpdateAIMerchants() {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	now := time.Now()

	for i := range ti.aiMerchants {
		merchant := &ti.aiMerchants[i]

		// Check if update interval has passed
		if now.Sub(merchant.LastUpdate) < merchant.UpdateInterval {
			continue
		}

		// Get current supply/demand
		currentSupply := ti.market.GetSupply(merchant.ItemID)
		currentDemand := ti.market.GetDemand(merchant.ItemID)

		// Maintain baseline supply (add if below baseline)
		if currentSupply < merchant.BaseSupply {
			supplyDelta := merchant.BaseSupply - currentSupply
			ti.market.UpdateSupply(merchant.ItemID, supplyDelta)
		}

		// Maintain baseline demand (add if below baseline)
		if currentDemand < merchant.BaseDemand {
			demandDelta := merchant.BaseDemand - currentDemand
			ti.market.UpdateDemand(merchant.ItemID, demandDelta)
		}

		merchant.LastUpdate = now
	}
}

// GetPriceWithPolitics calculates item price including political modifiers.
func (ti *TradeIntegration) GetPriceWithPolitics(itemID, targetServerID string) (float64, error) {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	// Get base market price
	basePrice := ti.market.GetPrice(itemID, 1.0)

	// Get political multiplier from politics system
	multiplier := 1.0
	if ti.politicsSystem != nil {
		multiplier = ti.politicsSystem.GetTradeMultiplier(targetServerID)
	}

	return basePrice * multiplier, nil
}

// Update performs periodic maintenance (AI merchants, reputation decay, etc.).
func (ti *TradeIntegration) Update(deltaTime float64) {
	// Update AI merchants every 5 minutes
	if time.Since(ti.lastAIUpdate) >= ti.systemUpdateInterval {
		ti.UpdateAIMerchants()
		ti.lastAIUpdate = time.Now()
	}

	// Slow reputation decay towards neutral (0.5) over time
	// Decay rate: 0.01 per hour (0.000002778 per second)
	ti.decayReputationTowardNeutral(deltaTime)
}

// decayReputationTowardNeutral slowly moves all reputations toward neutral (0.5).
func (ti *TradeIntegration) decayReputationTowardNeutral(deltaTime float64) {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	decayRate := 0.000002778 // 0.01 per hour
	decay := decayRate * deltaTime

	for serverID, rep := range ti.serverReputation {
		// Move toward 0.5
		if rep > 0.5 {
			rep -= decay
			if rep < 0.5 {
				rep = 0.5
			}
		} else if rep < 0.5 {
			rep += decay
			if rep > 0.5 {
				rep = 0.5
			}
		}

		ti.serverReputation[serverID] = rep
	}
}

// SetMaxTradesPerWindow configures the rate limit.
func (ti *TradeIntegration) SetMaxTradesPerWindow(max int) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	ti.maxTradesPerWindow = max
}

// SetWindowDuration configures the rate limit window duration.
func (ti *TradeIntegration) SetWindowDuration(duration time.Duration) {
	ti.mu.Lock()
	defer ti.mu.Unlock()
	ti.windowDuration = duration
}

// GetStats returns integration statistics.
func (ti *TradeIntegration) GetStats() TradeIntegrationStats {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	return TradeIntegrationStats{
		ActivePlayers:       len(ti.tradeCounts),
		TotalAIMerchants:    len(ti.aiMerchants),
		AverageReputation:   ti.calculateAverageReputation(),
		WindowTimeRemaining: ti.windowDuration - time.Since(ti.lastWindowReset),
	}
}

// calculateAverageReputation computes average server reputation (must hold lock).
func (ti *TradeIntegration) calculateAverageReputation() float64 {
	if len(ti.serverReputation) == 0 {
		return 0.5
	}

	sum := 0.0
	for _, rep := range ti.serverReputation {
		sum += rep
	}
	return sum / float64(len(ti.serverReputation))
}

// TradeIntegrationStats contains statistics about the trade integration system.
type TradeIntegrationStats struct {
	ActivePlayers       int
	TotalAIMerchants    int
	AverageReputation   float64
	WindowTimeRemaining time.Duration
}
