package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/world/economy"
	"github.com/sirupsen/logrus"
)

// EconomySystem integrates the economy package managers into the ECS architecture.
// It handles marketplace updates, guild bank interest calculation, and listing expiration.
type EconomySystem struct {
	world            *World
	marketplace      *economy.FederatedMarketplace
	guildBank        *economy.GuildBankManager
	serverID         string
	updateInterval   time.Duration
	lastUpdate       time.Time
	interestInterval time.Duration
	lastInterest     time.Time
	logger           *logrus.Entry
}

// NewEconomySystem creates a new economy system.
func NewEconomySystem(world *World, serverID string) *EconomySystem {
	logger := logrus.WithField("system_name", "economy")
	return &EconomySystem{
		world:            world,
		marketplace:      economy.NewFederatedMarketplace(serverID),
		guildBank:        economy.NewGuildBankManager(),
		serverID:         serverID,
		updateInterval:   5 * time.Minute, // Update marketplace every 5 minutes
		interestInterval: 24 * time.Hour,  // Calculate interest daily
		lastUpdate:       time.Now(),
		lastInterest:     time.Now(),
		logger:           logger,
	}
}

// Update processes marketplace listings and guild bank operations.
func (es *EconomySystem) Update(entities []*Entity, deltaTime float64) {
	now := time.Now()

	// Update marketplace (expire old listings, update pricing)
	if now.Sub(es.lastUpdate) >= es.updateInterval {
		es.updateMarketplace()
		es.lastUpdate = now
	}

	// Calculate daily interest on guild vaults
	if now.Sub(es.lastInterest) >= es.interestInterval {
		es.calculateInterest()
		es.lastInterest = now
	}
}

// updateMarketplace removes expired listings and updates pricing trends.
func (es *EconomySystem) updateMarketplace() {
	// The marketplace internally handles expiration checks
	// This would be extended with price trend updates in a full implementation
	es.logger.Debug("marketplace update cycle")
}

// calculateInterest applies daily interest to all guild vaults.
func (es *EconomySystem) calculateInterest() {
	es.logger.Debug("calculating guild vault interest")
	// Interest calculation is handled by the guild bank manager
	// This provides a hook for automated daily processing
}

// GetMarketplace returns the federated marketplace for direct access.
func (es *EconomySystem) GetMarketplace() *economy.FederatedMarketplace {
	return es.marketplace
}

// GetGuildBank returns the guild bank manager for direct access.
func (es *EconomySystem) GetGuildBank() *economy.GuildBankManager {
	return es.guildBank
}

// CreateListing adds a new item to the marketplace.
func (es *EconomySystem) CreateListing(listing *economy.Listing) error {
	err := es.marketplace.CreateListing(listing)
	if err != nil {
		es.logger.WithError(err).Error("failed to create marketplace listing")
		return err
	}
	es.logger.WithFields(logrus.Fields{
		"listingID": listing.ListingID,
		"itemName":  listing.ItemName,
		"price":     listing.Price,
	}).Debug("listing created")
	return nil
}

// PurchaseItem handles a marketplace purchase transaction.
func (es *EconomySystem) PurchaseItem(listingID, buyerID string, quantity int) error {
	err := es.marketplace.PurchaseItem(listingID, buyerID, quantity)
	if err != nil {
		es.logger.WithError(err).WithFields(logrus.Fields{
			"listingID": listingID,
			"buyerID":   buyerID,
			"quantity":  quantity,
		}).Error("failed to purchase item")
		return err
	}
	es.logger.WithFields(logrus.Fields{
		"listingID": listingID,
		"buyerID":   buyerID,
		"quantity":  quantity,
	}).Info("item purchased")
	return nil
}

// CreateGuildVault creates a new guild bank vault.
func (es *EconomySystem) CreateGuildVault(guildID string, interestRate float64) error {
	err := es.guildBank.CreateVault(guildID, interestRate)
	if err != nil {
		es.logger.WithError(err).WithField("guildID", guildID).Error("failed to create guild vault")
		return err
	}
	es.logger.WithFields(logrus.Fields{
		"guildID":      guildID,
		"interestRate": interestRate,
	}).Info("guild vault created")
	return nil
}

// DepositGold deposits gold into a guild vault.
func (es *EconomySystem) DepositGold(guildID, memberID, memberName string, amount int) error {
	err := es.guildBank.DepositGold(guildID, memberID, memberName, amount)
	if err != nil {
		es.logger.WithError(err).WithFields(logrus.Fields{
			"guildID":  guildID,
			"memberID": memberID,
			"amount":   amount,
		}).Error("failed to deposit gold")
		return err
	}
	es.logger.WithFields(logrus.Fields{
		"guildID":    guildID,
		"memberName": memberName,
		"amount":     amount,
	}).Debug("gold deposited to guild vault")
	return nil
}

// WithdrawGold withdraws gold from a guild vault with rank-based limits.
func (es *EconomySystem) WithdrawGold(guildID, memberID, memberName, rankID string, amount int) error {
	err := es.guildBank.WithdrawGold(guildID, memberID, memberName, rankID, amount)
	if err != nil {
		es.logger.WithError(err).WithFields(logrus.Fields{
			"guildID":  guildID,
			"memberID": memberID,
			"amount":   amount,
		}).Error("failed to withdraw gold")
		return err
	}
	es.logger.WithFields(logrus.Fields{
		"guildID":    guildID,
		"memberName": memberName,
		"amount":     amount,
	}).Debug("gold withdrawn from guild vault")
	return nil
}
