package companion_housing

// StorageChest represents shared inventory accessible by companions.
// This file contains the StorageChest type and its methods for managing
// items in shared storage between players and companions.
//
// Code relocated from: types.go

// StorageChest represents shared inventory accessible by companions.
type StorageChest struct {
	FurnitureID    string   // Unique furniture identifier
	HouseID        string   // Owner house identifier
	Capacity       int      // Total slot count (50-100 typical)
	SharedWithPets bool     // If true, companions can deposit/withdraw
	Items          []string // Item IDs stored in chest
}

// AddItem adds an item to the chest if space available.
// Returns false if chest is full.
func (s *StorageChest) AddItem(itemID string) bool {
	if len(s.Items) >= s.Capacity {
		return false
	}
	s.Items = append(s.Items, itemID)
	return true
}

// RemoveItem removes an item from the chest if present.
// Returns false if item not found.
func (s *StorageChest) RemoveItem(itemID string) bool {
	for i, id := range s.Items {
		if id == itemID {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

// AvailableSlots returns number of empty slots in chest.
func (s *StorageChest) AvailableSlots() int {
	return s.Capacity - len(s.Items)
}
