package economy

import (
	"time"
)

// SortCriteria defines how search results are sorted.
type SortCriteria int

const (
	// SortByPrice sorts items by price (ascending).
	SortByPrice SortCriteria = iota

	// SortByPriceDesc sorts items by price (descending).
	SortByPriceDesc

	// SortByQuantity sorts items by quantity available.
	SortByQuantity

	// SortByDeliveryTime sorts items by estimated delivery time.
	SortByDeliveryTime

	// SortByRelevance sorts items by search relevance.
	SortByRelevance
)

// String returns the name of the sort criteria.
func (sc SortCriteria) String() string {
	switch sc {
	case SortByPrice:
		return "Price (Ascending)"
	case SortByPriceDesc:
		return "Price (Descending)"
	case SortByQuantity:
		return "Quantity"
	case SortByDeliveryTime:
		return "Delivery Time"
	case SortByRelevance:
		return "Relevance"
	default:
		return "Unknown"
	}
}

// DeliveryMethod defines how items are delivered.
type DeliveryMethod int

const (
	// DeliveryMail uses V6 mail system (instant).
	DeliveryMail DeliveryMethod = iota

	// DeliveryCourier uses NPC courier (10-60 minute delay).
	DeliveryCourier
)

// String returns the name of the delivery method.
func (dm DeliveryMethod) String() string {
	switch dm {
	case DeliveryMail:
		return "Mail"
	case DeliveryCourier:
		return "Courier"
	default:
		return "Unknown"
	}
}

// Listing represents an item for sale on the marketplace.
type Listing struct {
	ListingID      string
	ItemID         string
	ItemName       string
	ItemType       string
	SellerID       string
	SellerName     string
	ServerID       string
	Price          int
	Quantity       int
	MaxStackSize   int
	CreatedAt      time.Time
	ExpiresAt      time.Time
	DeliveryMethod DeliveryMethod
	EstimatedHops  int // Number of server hops for delivery
}

// ItemQuery defines search parameters for marketplace items.
type ItemQuery struct {
	ItemType string
	ItemName string
	MinPrice int
	MaxPrice int
	ServerID string // Empty = search all servers
	SellerID string // Empty = all sellers
	SortBy   SortCriteria
	Limit    int // Max results to return
}

// PriceTrend contains pricing statistics for an item type.
type PriceTrend struct {
	ItemType      string
	AveragePrice  int
	MinPrice      int
	MaxPrice      int
	TotalVolume   int
	TotalListings int
	LastUpdated   time.Time
}

// Transaction represents a completed marketplace transaction.
type Transaction struct {
	TransactionID  string
	ListingID      string
	BuyerID        string
	SellerID       string
	ItemID         string
	Quantity       int
	Price          int
	TransactionFee int
	Timestamp      time.Time
	OriginServer   string
	DestServer     string
}

// CalculateTransactionFee computes the fee for a transaction.
// Base fee is 5%, with +2% per server hop (max 15%).
func CalculateTransactionFee(price, hops int) int {
	baseFee := 0.05
	hopFee := float64(hops) * 0.02
	totalFee := baseFee + hopFee
	if totalFee > 0.15 {
		totalFee = 0.15
	}
	return int(float64(price) * totalFee)
}

// IsExpired returns true if the listing has expired.
func (l *Listing) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// GetDeliveryTime estimates delivery time in seconds.
func (l *Listing) GetDeliveryTime() int {
	switch l.DeliveryMethod {
	case DeliveryMail:
		return 0 // Instant
	case DeliveryCourier:
		// 10-60 minutes based on server hops
		baseTime := 600                  // 10 minutes
		hopTime := l.EstimatedHops * 300 // 5 minutes per hop
		totalTime := baseTime + hopTime
		if totalTime > 3600 {
			totalTime = 3600 // Max 60 minutes
		}
		return totalTime
	default:
		return 0
	}
}

// GetTotalCost returns the total cost including transaction fee.
func (l *Listing) GetTotalCost(quantity int) int {
	basePrice := l.Price * quantity
	fee := CalculateTransactionFee(basePrice, l.EstimatedHops)
	return basePrice + fee
}
