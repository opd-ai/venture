// Package qol - storagesorter.go
// This file contains the StorageSorter implementation for inventory and storage sorting.
// Code relocated from: manager.go

package qol

import (
	"sort"
	"sync"
)

// StorageSorter handles inventory and storage sorting with customizable presets.
// Provides default presets for common sorting patterns (by type, rarity, value).
type StorageSorter struct {
	presets map[string]*StorageSortPreset
	mu      sync.RWMutex
}

// NewStorageSorter creates a new storage sorter with default presets.
func NewStorageSorter() *StorageSorter {
	s := &StorageSorter{
		presets: make(map[string]*StorageSortPreset),
	}

	s.presets["default"] = &StorageSortPreset{
		Name:              "Default",
		PrimaryCriteria:   SortByType,
		SecondaryCriteria: SortByRarity,
		Descending:        false,
		GroupByType:       true,
	}

	s.presets["rarity"] = &StorageSortPreset{
		Name:              "Rarity",
		PrimaryCriteria:   SortByRarity,
		SecondaryCriteria: SortByName,
		Descending:        true,
		GroupByType:       false,
	}

	s.presets["value"] = &StorageSortPreset{
		Name:              "Value",
		PrimaryCriteria:   SortByValue,
		SecondaryCriteria: SortByQuantity,
		Descending:        true,
		GroupByType:       false,
	}

	return s
}

// AddPreset adds or updates a custom sort preset by name.
func (s *StorageSorter) AddPreset(preset *StorageSortPreset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.presets[preset.Name] = preset
}

// GetPreset retrieves a sort preset by name.
// Returns nil if the preset does not exist.
func (s *StorageSorter) GetPreset(name string) *StorageSortPreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.presets[name]
}

// SortItems sorts a slice of items in-place using the specified criteria.
// Uses stable sort to preserve relative order of equal elements.
func (s *StorageSorter) SortItems(items []*Item, criteria SortCriteria) {
	sort.SliceStable(items, func(i, j int) bool {
		switch criteria {
		case SortByType:
			return items[i].Type < items[j].Type
		case SortByRarity:
			return items[i].Rarity > items[j].Rarity
		case SortByName:
			return items[i].Name < items[j].Name
		case SortByValue:
			return items[i].Value > items[j].Value
		case SortByQuantity:
			return items[i].Quantity > items[j].Quantity
		default:
			return false
		}
	})
}
