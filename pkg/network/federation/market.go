package federation

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
)

// FederatedMarket manages cross-server item trading and pricing.
type FederatedMarket struct {
	mu           sync.RWMutex
	itemPrices   map[string]*PriceHistory // ItemID -> price history
	supply       map[string]int           // ItemID -> available quantity
	demand       map[string]int           // ItemID -> buy order quantity
	lastUpdate   time.Time
	updateTicker *time.Ticker
	stopChan     chan struct{}

	// G14 (AUDIT.md): startOnce/stopOnce prevent double-start and double-stop panics.
	startOnce sync.Once
	stopOnce  sync.Once
}

// PriceHistory tracks price changes for a specific item.
type PriceHistory struct {
	ItemID       string
	ServerID     string
	CurrentPrice float64
	History      []PricePoint // Last 24 hours of price points
	BasePrice    float64      // Original item price without market forces
}

// PricePoint represents a single price observation.
type PricePoint struct {
	Timestamp time.Time
	Price     float64
	Supply    int
	Demand    int
}

// MarketListing represents an item available for sale.
type MarketListing struct {
	ID           string
	SellerID     string
	ServerID     string
	ItemID       string
	ItemName     string
	Quantity     int
	PricePerUnit float64
	TotalPrice   float64
	ListedAt     time.Time
	ExpiresAt    time.Time
}

// BuyOrder represents a buy request for an item.
type BuyOrder struct {
	ID        string
	BuyerID   string
	ServerID  string
	ItemID    string
	Quantity  int
	MaxPrice  float64
	PlacedAt  time.Time
	ExpiresAt time.Time
}

// Transaction represents a completed trade.
type Transaction struct {
	ID           string
	BuyerID      string
	SellerID     string
	ItemID       string
	Quantity     int
	PricePerUnit float64
	TotalPrice   float64
	ShippingCost float64
	ServerHops   int
	CompletedAt  time.Time
}

// NewFederatedMarket creates a new federated market instance.
func NewFederatedMarket() *FederatedMarket {
	return &FederatedMarket{
		itemPrices: make(map[string]*PriceHistory),
		supply:     make(map[string]int),
		demand:     make(map[string]int),
		lastUpdate: time.Now(),
		stopChan:   make(chan struct{}),
	}
}

// Start begins the market update loop (price updates every 60 seconds).
// Idempotent: subsequent calls after the first are no-ops.
func (m *FederatedMarket) Start() {
	m.startOnce.Do(func() {
		m.updateTicker = time.NewTicker(60 * time.Second)
		go func() {
			defer recovery.RecoverPanicWithLogger("federation_market", "update loop", nil)()
			for {
				select {
				case <-m.updateTicker.C:
					m.UpdatePrices()
				case <-m.stopChan:
					return
				}
			}
		}()
	})
}

// Stop halts the market update loop.
// Idempotent: subsequent calls after the first are no-ops.
func (m *FederatedMarket) Stop() {
	m.stopOnce.Do(func() {
		if m.updateTicker != nil {
			m.updateTicker.Stop()
		}
		close(m.stopChan)
	})
}

// RegisterItem initializes price tracking for an item.
func (m *FederatedMarket) RegisterItem(itemID, serverID string, basePrice float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.itemPrices[itemID]; !exists {
		m.itemPrices[itemID] = &PriceHistory{
			ItemID:       itemID,
			ServerID:     serverID,
			CurrentPrice: basePrice,
			BasePrice:    basePrice,
			History:      make([]PricePoint, 0, 288), // 24h at 5min intervals = 288 points
		}
		m.supply[itemID] = 0
		m.demand[itemID] = 0
	}
}

// UpdateSupply modifies the supply count for an item.
func (m *FederatedMarket) UpdateSupply(itemID string, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.supply[itemID] += delta
	if m.supply[itemID] < 0 {
		m.supply[itemID] = 0
	}
}

// UpdateDemand modifies the demand count for an item.
func (m *FederatedMarket) UpdateDemand(itemID string, delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.demand[itemID] += delta
	if m.demand[itemID] < 0 {
		m.demand[itemID] = 0
	}
}

// GetPrice returns the current market price for an item with server multiplier.
func (m *FederatedMarket) GetPrice(itemID string, serverMultiplier float64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.itemPrices[itemID]
	if !exists {
		return 0.0
	}

	return history.CurrentPrice * serverMultiplier
}

// CalculatePrice computes price based on supply/demand dynamics.
// Formula: Price = BasePrice × (Demand / Supply) × ServerMultiplier
func (m *FederatedMarket) CalculatePrice(itemID string, serverMultiplier float64) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.itemPrices[itemID]
	if !exists {
		return 0.0
	}

	supply := m.supply[itemID]
	demand := m.demand[itemID]

	// Handle edge cases
	if supply == 0 && demand == 0 {
		return history.BasePrice * serverMultiplier
	}
	if supply == 0 {
		// High demand, no supply: 3x base price
		return history.BasePrice * 3.0 * serverMultiplier
	}

	// Calculate supply/demand ratio (clamped to 0.2x - 5.0x)
	ratio := float64(demand) / float64(supply)
	ratio = math.Max(0.2, math.Min(5.0, ratio))

	return history.BasePrice * ratio * serverMultiplier
}

// UpdatePrices recalculates all item prices based on current supply/demand.
func (m *FederatedMarket) UpdatePrices() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	for itemID, history := range m.itemPrices {
		supply := m.supply[itemID]
		demand := m.demand[itemID]

		// Calculate new price with neutral multiplier (1.0)
		newPrice := m.calculatePriceInternal(history, supply, demand, 1.0)
		history.CurrentPrice = newPrice

		// Record price point
		point := PricePoint{
			Timestamp: now,
			Price:     newPrice,
			Supply:    supply,
			Demand:    demand,
		}
		history.History = append(history.History, point)

		// Trim history to last 24 hours (288 points at 5min intervals)
		if len(history.History) > 288 {
			history.History = history.History[len(history.History)-288:]
		}
	}

	m.lastUpdate = now
}

// calculatePriceInternal is the internal price calculation (must hold lock).
func (m *FederatedMarket) calculatePriceInternal(history *PriceHistory, supply, demand int, serverMultiplier float64) float64 {
	if supply == 0 && demand == 0 {
		return history.BasePrice * serverMultiplier
	}
	if supply == 0 {
		return history.BasePrice * 3.0 * serverMultiplier
	}

	ratio := float64(demand) / float64(supply)
	ratio = math.Max(0.2, math.Min(5.0, ratio))

	return history.BasePrice * ratio * serverMultiplier
}

// GetPriceHistory returns the price history for an item.
func (m *FederatedMarket) GetPriceHistory(itemID string) (*PriceHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	history, exists := m.itemPrices[itemID]
	if !exists {
		return nil, fmt.Errorf("item %s not registered in market", itemID)
	}

	// Return a copy to avoid concurrent modification
	historyCopy := &PriceHistory{
		ItemID:       history.ItemID,
		ServerID:     history.ServerID,
		CurrentPrice: history.CurrentPrice,
		BasePrice:    history.BasePrice,
		History:      make([]PricePoint, len(history.History)),
	}
	copy(historyCopy.History, history.History)

	return historyCopy, nil
}

// CalculateShippingCost computes shipping cost based on server hops.
// Formula: +10% per server hop
func CalculateShippingCost(basePrice float64, serverHops int) float64 {
	if serverHops <= 0 {
		return 0.0
	}
	return basePrice * float64(serverHops) * 0.10
}

// GetSupply returns the current supply for an item.
func (m *FederatedMarket) GetSupply(itemID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.supply[itemID]
}

// GetDemand returns the current demand for an item.
func (m *FederatedMarket) GetDemand(itemID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.demand[itemID]
}

// GetStats returns market statistics for monitoring.
func (m *FederatedMarket) GetStats() MarketStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MarketStats{
		TotalItems:  len(m.itemPrices),
		TotalSupply: sumMap(m.supply),
		TotalDemand: sumMap(m.demand),
		LastUpdate:  m.lastUpdate,
	}
}

// MarketStats provides market overview statistics.
type MarketStats struct {
	TotalItems  int
	TotalSupply int
	TotalDemand int
	LastUpdate  time.Time
}

// sumMap sums all values in an integer map.
func sumMap(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
