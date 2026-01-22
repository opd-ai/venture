package economy

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FederatedMarketplace manages cross-server item listings and transactions.
type FederatedMarketplace struct {
	localServerID    string
	localListings    map[string]*Listing
	remoteCache      map[string][]*Listing // serverID -> listings
	pricingEngine    *PricingEngine
	transactions     []*Transaction
	maxListingsLocal int
	mu               sync.RWMutex
}

// NewFederatedMarketplace creates a new federated marketplace.
func NewFederatedMarketplace(serverID string) *FederatedMarketplace {
	return &FederatedMarketplace{
		localServerID:    serverID,
		localListings:    make(map[string]*Listing),
		remoteCache:      make(map[string][]*Listing),
		pricingEngine:    NewPricingEngine(),
		transactions:     make([]*Transaction, 0),
		maxListingsLocal: 10000,
	}
}

// CreateListing adds a new item listing to the marketplace.
func (fm *FederatedMarketplace) CreateListing(listing *Listing) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Validate listing
	if listing.ItemID == "" {
		return fmt.Errorf("item ID is required")
	}
	if listing.SellerID == "" {
		return fmt.Errorf("seller ID is required")
	}
	if listing.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if listing.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Check capacity
	if len(fm.localListings) >= fm.maxListingsLocal {
		return fmt.Errorf("marketplace at capacity (%d listings)", fm.maxListingsLocal)
	}

	// Generate listing ID if not provided
	if listing.ListingID == "" {
		listing.ListingID = uuid.New().String()
	}

	// Set server ID and timestamps
	listing.ServerID = fm.localServerID
	listing.CreatedAt = time.Now()
	if listing.ExpiresAt.IsZero() {
		// Default 7-day expiration
		listing.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	// Set delivery method based on item properties
	if listing.MaxStackSize > 1 {
		listing.DeliveryMethod = DeliveryMail
	} else {
		listing.DeliveryMethod = DeliveryCourier
	}
	listing.EstimatedHops = 0 // Local server, no hops

	fm.localListings[listing.ListingID] = listing

	// Update pricing engine
	fm.pricingEngine.RecordListing(listing)

	return nil
}

// GetListing retrieves a listing by ID.
func (fm *FederatedMarketplace) GetListing(listingID string) (*Listing, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	listing, exists := fm.localListings[listingID]
	if !exists {
		return nil, fmt.Errorf("listing not found: %s", listingID)
	}

	return listing, nil
}

// RemoveListing removes a listing from the marketplace.
func (fm *FederatedMarketplace) RemoveListing(listingID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.localListings[listingID]; !exists {
		return fmt.Errorf("listing not found: %s", listingID)
	}

	delete(fm.localListings, listingID)
	return nil
}

// SearchItems searches for items across local and remote servers.
func (fm *FederatedMarketplace) SearchItems(query ItemQuery) ([]*Listing, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	results := fm.collectMatchingListings(query)
	validResults := filterExpiredListings(results)
	fm.sortListings(validResults, query.SortBy)
	return applyResultLimit(validResults, query.Limit), nil
}

// collectMatchingListings gathers all listings matching the search query.
func (fm *FederatedMarketplace) collectMatchingListings(query ItemQuery) []*Listing {
	results := fm.searchLocalListings(query)
	if fm.shouldSearchRemote(query) {
		results = append(results, fm.searchRemoteListings(query)...)
	}
	return results
}

// searchLocalListings searches listings on the local server.
func (fm *FederatedMarketplace) searchLocalListings(query ItemQuery) []*Listing {
	var results []*Listing
	for _, listing := range fm.localListings {
		if fm.matchesQuery(listing, query) {
			results = append(results, listing)
		}
	}
	return results
}

// shouldSearchRemote determines if remote cache should be searched.
func (fm *FederatedMarketplace) shouldSearchRemote(query ItemQuery) bool {
	return query.ServerID == "" || query.ServerID != fm.localServerID
}

// searchRemoteListings searches cached remote listings.
func (fm *FederatedMarketplace) searchRemoteListings(query ItemQuery) []*Listing {
	var results []*Listing
	for _, remoteListings := range fm.remoteCache {
		for _, listing := range remoteListings {
			if fm.matchesQuery(listing, query) {
				results = append(results, listing)
			}
		}
	}
	return results
}

// filterExpiredListings removes expired listings from results.
func filterExpiredListings(listings []*Listing) []*Listing {
	validResults := make([]*Listing, 0)
	for _, listing := range listings {
		if !listing.IsExpired() {
			validResults = append(validResults, listing)
		}
	}
	return validResults
}

// applyResultLimit limits the number of results returned.
func applyResultLimit(listings []*Listing, limit int) []*Listing {
	if limit > 0 && len(listings) > limit {
		return listings[:limit]
	}
	return listings
}

// matchesQuery checks if a listing matches the search query.
func (fm *FederatedMarketplace) matchesQuery(listing *Listing, query ItemQuery) bool {
	// Type filter
	if query.ItemType != "" && listing.ItemType != query.ItemType {
		return false
	}

	// Name filter (case-insensitive substring match)
	if query.ItemName != "" {
		// Simple contains check (could be enhanced with fuzzy matching)
		nameMatch := false
		for i := 0; i < len(listing.ItemName)-len(query.ItemName)+1; i++ {
			if listing.ItemName[i:i+len(query.ItemName)] == query.ItemName {
				nameMatch = true
				break
			}
		}
		if !nameMatch {
			return false
		}
	}

	// Price range filter
	if query.MinPrice > 0 && listing.Price < query.MinPrice {
		return false
	}
	if query.MaxPrice > 0 && listing.Price > query.MaxPrice {
		return false
	}

	// Server filter
	if query.ServerID != "" && listing.ServerID != query.ServerID {
		return false
	}

	// Seller filter
	if query.SellerID != "" && listing.SellerID != query.SellerID {
		return false
	}

	return true
}

// sortListings sorts listings by the specified criteria.
func (fm *FederatedMarketplace) sortListings(listings []*Listing, sortBy SortCriteria) {
	switch sortBy {
	case SortByPrice:
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].Price < listings[j].Price
		})
	case SortByPriceDesc:
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].Price > listings[j].Price
		})
	case SortByQuantity:
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].Quantity > listings[j].Quantity
		})
	case SortByDeliveryTime:
		sort.Slice(listings, func(i, j int) bool {
			return listings[i].GetDeliveryTime() < listings[j].GetDeliveryTime()
		})
	case SortByRelevance:
		// Relevance = combination of price and delivery time
		sort.Slice(listings, func(i, j int) bool {
			scoreI := float64(listings[i].Price) + float64(listings[i].GetDeliveryTime())*0.1
			scoreJ := float64(listings[j].Price) + float64(listings[j].GetDeliveryTime())*0.1
			return scoreI < scoreJ
		})
	}
}

// PurchaseItem executes a purchase transaction.
func (fm *FederatedMarketplace) PurchaseItem(listingID, buyerID string, quantity int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Find listing
	listing, exists := fm.localListings[listingID]
	if !exists {
		return fmt.Errorf("listing not found: %s", listingID)
	}

	// Check expiration
	if listing.IsExpired() {
		return fmt.Errorf("listing has expired")
	}

	// Check quantity
	if quantity > listing.Quantity {
		return fmt.Errorf("insufficient quantity (requested: %d, available: %d)", quantity, listing.Quantity)
	}

	// Calculate cost and fee
	totalCost := listing.GetTotalCost(quantity)

	// Create transaction record
	transaction := &Transaction{
		TransactionID:  uuid.New().String(),
		ListingID:      listingID,
		BuyerID:        buyerID,
		SellerID:       listing.SellerID,
		ItemID:         listing.ItemID,
		Quantity:       quantity,
		Price:          listing.Price * quantity,
		TransactionFee: CalculateTransactionFee(listing.Price*quantity, listing.EstimatedHops),
		Timestamp:      time.Now(),
		OriginServer:   listing.ServerID,
		DestServer:     fm.localServerID,
	}

	fm.transactions = append(fm.transactions, transaction)

	// Update pricing engine
	fm.pricingEngine.RecordTransaction(listing, quantity)

	// Update or remove listing
	listing.Quantity -= quantity
	if listing.Quantity == 0 {
		delete(fm.localListings, listingID)
	}

	// In real implementation:
	// - Deduct gold from buyer
	// - Transfer gold to seller
	// - Deliver item via mail/courier
	// For now, just log the total cost
	_ = totalCost

	return nil
}

// GetPriceTrend returns price statistics for an item type.
func (fm *FederatedMarketplace) GetPriceTrend(itemType string) *PriceTrend {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.pricingEngine.GetTrend(itemType)
}

// UpdateRemoteCache updates the cache of remote server listings.
func (fm *FederatedMarketplace) UpdateRemoteCache(serverID string, listings []*Listing) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Set estimated hops based on server distance
	// For now, assume 1 hop per remote server
	for _, listing := range listings {
		listing.EstimatedHops = 1
		if listing.ServerID != serverID {
			listing.EstimatedHops = 2 // Multi-hop
		}
	}

	fm.remoteCache[serverID] = listings
}

// GetTransactionHistory returns recent transactions.
func (fm *FederatedMarketplace) GetTransactionHistory(limit int) []*Transaction {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if limit <= 0 || limit > len(fm.transactions) {
		limit = len(fm.transactions)
	}

	// Return most recent transactions
	start := len(fm.transactions) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*Transaction, limit)
	copy(history, fm.transactions[start:])
	return history
}

// CleanupExpiredListings removes expired listings from the marketplace.
func (fm *FederatedMarketplace) CleanupExpiredListings() int {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	removed := 0
	for listingID, listing := range fm.localListings {
		if listing.IsExpired() {
			delete(fm.localListings, listingID)
			removed++
		}
	}

	return removed
}

// GetStats returns marketplace statistics.
func (fm *FederatedMarketplace) GetStats() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return map[string]interface{}{
		"local_listings":   len(fm.localListings),
		"cached_servers":   len(fm.remoteCache),
		"total_cached":     fm.countRemoteListings(),
		"transactions":     len(fm.transactions),
		"capacity":         fm.maxListingsLocal,
		"capacity_percent": float64(len(fm.localListings)) / float64(fm.maxListingsLocal) * 100.0,
	}
}

// countRemoteListings counts total remote listings in cache.
func (fm *FederatedMarketplace) countRemoteListings() int {
	count := 0
	for _, listings := range fm.remoteCache {
		count += len(listings)
	}
	return count
}
