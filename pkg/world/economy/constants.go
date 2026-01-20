// Package economy constants consolidation.
// This file contains all enum constant definitions for marketplace sorting
// and delivery methods, consolidated from types.go for centralized management.
package economy

// SortCriteria constants
// Originally from: types.go
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

// DeliveryMethod constants
// Originally from: types.go
const (
	// DeliveryMail uses V6 mail system (instant).
	DeliveryMail DeliveryMethod = iota

	// DeliveryCourier uses NPC courier (10-60 minute delay).
	DeliveryCourier
)
