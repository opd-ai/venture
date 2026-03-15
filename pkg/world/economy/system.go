package economy

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// System is an ECS system that manages the federated marketplace.
// It handles marketplace cleanup, price trend updates, and entity-based transactions.
type System struct {
	world             World
	marketplace       *FederatedMarketplace
	cleanupInterval   time.Duration
	lastCleanup       time.Time
	priceUpdateTicker float64
	timeProvider      TimeProvider
	mu                sync.RWMutex
}

// NewSystem creates a new economy system with the given world.
func NewSystem(world World) *System {
	return NewSystemWithServerID(world, "local-server")
}

// NewSystemWithServerID creates a new economy system with a specific server ID.
func NewSystemWithServerID(world World, serverID string) *System {
	return NewSystemWithTimeProvider(world, serverID, DefaultTimeProvider())
}

// NewSystemWithTimeProvider creates a new economy system with a custom time provider.
func NewSystemWithTimeProvider(world World, serverID string, tp TimeProvider) *System {
	return &System{
		world:           world,
		marketplace:     NewFederatedMarketplaceWithTime(serverID, tp),
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     tp.Now(),
		timeProvider:    tp,
	}
}

// Update is called each frame to update the economy system.
// NOTE: pkg/world/economy cannot import pkg/engine directly (circular dependency via
// pkg/engine/economy_system.go). Integration with the ECS world uses the adapter
// pattern in cmd/server/system_wrappers.go instead of direct engine.System registration.
func (s *System) Update(deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Periodic cleanup of expired listings
	s.priceUpdateTicker += deltaTime
	if s.priceUpdateTicker >= 60.0 { // Every 60 seconds
		s.priceUpdateTicker = 0
		if s.timeProvider.Now().Sub(s.lastCleanup) >= s.cleanupInterval {
			removed := s.marketplace.CleanupExpiredListings()
			s.lastCleanup = s.timeProvider.Now()
			if removed > 0 {
				log.WithFields(log.Fields{
					"system_name": "economy",
					"removed":     removed,
				}).Debug("economy system cleanup cycle")
			}
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

// PurchaseItem executes a purchase transaction and returns the transaction record.
func (s *System) PurchaseItem(listingID, buyerID string, quantity int) (*Transaction, error) {
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

// ApplyTradeImpact applies price changes from external trade activity.
// This method enables trade routes and other systems to influence market prices.
// itemType: The type of item affected (e.g., "Timber", "Ore")
// priceChange: Multiplier for price impact (1.1 = +10%, 0.9 = -10%)
// volume: Quantity traded, used to weight the impact
func (s *System) ApplyTradeImpact(itemType string, priceChange float64, volume int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pricingEngine := s.marketplace.GetPricingEngine()
	pricingEngine.ApplyTradeImpact(itemType, priceChange, volume)
}
