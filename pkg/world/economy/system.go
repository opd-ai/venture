package economy

import (
	"sync"
	"time"
)

// World is the minimal interface for ECS world operations needed by the economy system.
type World interface {
	GetEntities() []Entity
}

// Entity is the minimal interface for ECS entities needed by the economy system.
type Entity interface {
	HasComponent(componentType string) bool
	GetComponent(componentType string) (interface{}, bool)
}

// System is an ECS system that manages the federated marketplace.
// It handles marketplace cleanup, price trend updates, and entity-based transactions.
type System struct {
	world             World
	marketplace       *FederatedMarketplace
	cleanupInterval   time.Duration
	lastCleanup       time.Time
	priceUpdateTicker float64
	mu                sync.RWMutex
}

// NewSystem creates a new economy system with the given world.
func NewSystem(world World) *System {
	// Generate server ID from world pointer to ensure uniqueness
	serverID := "local-server"
	return &System{
		world:           world,
		marketplace:     NewFederatedMarketplace(serverID),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
}

// NewSystemWithServerID creates a new economy system with a specific server ID.
func NewSystemWithServerID(world World, serverID string) *System {
	return &System{
		world:           world,
		marketplace:     NewFederatedMarketplace(serverID),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}
}

// Update is called each frame to update the economy system.
func (s *System) Update(deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Periodic cleanup of expired listings
	s.priceUpdateTicker += deltaTime
	if s.priceUpdateTicker >= 60.0 { // Every 60 seconds
		s.priceUpdateTicker = 0
		if time.Since(s.lastCleanup) >= s.cleanupInterval {
			s.marketplace.CleanupExpiredListings()
			s.lastCleanup = time.Now()
		}
	}
}

// GetMarketplace returns the federated marketplace for direct access.
func (s *System) GetMarketplace() *FederatedMarketplace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.marketplace
}

// CreateListing creates a new marketplace listing.
func (s *System) CreateListing(listing *Listing) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.marketplace.CreateListing(listing)
}

// SearchItems searches for items matching the query.
func (s *System) SearchItems(query ItemQuery) ([]*Listing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.marketplace.SearchItems(query)
}

// PurchaseItem executes a purchase transaction.
func (s *System) PurchaseItem(listingID, buyerID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.marketplace.PurchaseItem(listingID, buyerID, quantity)
}

// GetPriceTrend returns price statistics for an item type.
func (s *System) GetPriceTrend(itemType string) *PriceTrend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.marketplace.GetPriceTrend(itemType)
}

// GetStats returns marketplace statistics.
func (s *System) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.marketplace.GetStats()
}
