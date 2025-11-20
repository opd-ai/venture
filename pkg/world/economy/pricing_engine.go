package economy

import (
	"sync"
	"time"
)

// PricingEngine tracks price trends and market dynamics.
type PricingEngine struct {
	trends map[string]*PriceTrend
	mu     sync.RWMutex
}

// NewPricingEngine creates a new pricing engine.
func NewPricingEngine() *PricingEngine {
	return &PricingEngine{
		trends: make(map[string]*PriceTrend),
	}
}

// RecordListing updates price trends when a listing is created.
func (pe *PricingEngine) RecordListing(listing *Listing) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	trend, exists := pe.trends[listing.ItemType]
	if !exists {
		trend = &PriceTrend{
			ItemType:      listing.ItemType,
			AveragePrice:  listing.Price,
			MinPrice:      listing.Price,
			MaxPrice:      listing.Price,
			TotalVolume:   listing.Quantity,
			TotalListings: 1,
			LastUpdated:   time.Now(),
		}
		pe.trends[listing.ItemType] = trend
		return
	}

	// Update trend statistics
	trend.TotalListings++
	trend.TotalVolume += listing.Quantity

	// Update min/max prices
	if listing.Price < trend.MinPrice {
		trend.MinPrice = listing.Price
	}
	if listing.Price > trend.MaxPrice {
		trend.MaxPrice = listing.Price
	}

	// Recalculate average (weighted by quantity)
	totalValue := trend.AveragePrice*trend.TotalVolume + listing.Price*listing.Quantity
	trend.AveragePrice = totalValue / (trend.TotalVolume + listing.Quantity)
	trend.LastUpdated = time.Now()
}

// RecordTransaction updates price trends when a transaction occurs.
func (pe *PricingEngine) RecordTransaction(listing *Listing, quantity int) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	trend, exists := pe.trends[listing.ItemType]
	if !exists {
		// Create trend if it doesn't exist
		trend = &PriceTrend{
			ItemType:      listing.ItemType,
			AveragePrice:  listing.Price,
			MinPrice:      listing.Price,
			MaxPrice:      listing.Price,
			TotalVolume:   quantity,
			TotalListings: 0,
			LastUpdated:   time.Now(),
		}
		pe.trends[listing.ItemType] = trend
		return
	}

	// Update volume (transaction volume is separate from listing volume)
	trend.TotalVolume += quantity
	trend.LastUpdated = time.Now()

	// Transactions influence average price more heavily (2x weight)
	totalValue := trend.AveragePrice*trend.TotalVolume + listing.Price*quantity*2
	trend.AveragePrice = totalValue / (trend.TotalVolume + quantity*2)
}

// GetTrend returns price trend for an item type.
func (pe *PricingEngine) GetTrend(itemType string) *PriceTrend {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	trend, exists := pe.trends[itemType]
	if !exists {
		// Return empty trend
		return &PriceTrend{
			ItemType:      itemType,
			AveragePrice:  0,
			MinPrice:      0,
			MaxPrice:      0,
			TotalVolume:   0,
			TotalListings: 0,
			LastUpdated:   time.Time{},
		}
	}

	// Return copy to avoid external mutation
	trendCopy := *trend
	return &trendCopy
}

// GetAllTrends returns all price trends.
func (pe *PricingEngine) GetAllTrends() []*PriceTrend {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	trends := make([]*PriceTrend, 0, len(pe.trends))
	for _, trend := range pe.trends {
		trendCopy := *trend
		trends = append(trends, &trendCopy)
	}

	return trends
}

// ResetTrends clears all price trend data.
func (pe *PricingEngine) ResetTrends() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.trends = make(map[string]*PriceTrend)
}
