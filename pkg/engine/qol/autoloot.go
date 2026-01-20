// Package qol - autoloot.go
// This file contains the AutoLootManager implementation for companion auto-loot behavior.
// Code relocated from: manager.go

package qol

import (
	"sync"
)

// AutoLootManager manages companion auto-loot behavior
type AutoLootManager struct {
	configs map[uint64]*AutoLootConfig
	mu      sync.RWMutex
}

// NewAutoLootManager creates a new auto-loot manager
func NewAutoLootManager() *AutoLootManager {
	return &AutoLootManager{
		configs: make(map[uint64]*AutoLootConfig),
	}
}

// SetConfig sets auto-loot configuration for a companion
func (m *AutoLootManager) SetConfig(config *AutoLootConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[config.CompanionID] = config
}

// GetConfig retrieves auto-loot configuration for a companion
func (m *AutoLootManager) GetConfig(companionID uint64) *AutoLootConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if config, exists := m.configs[companionID]; exists {
		return config
	}
	return DefaultAutoLootConfig(companionID)
}

// SetRadius sets the loot collection radius for a companion
func (m *AutoLootManager) SetRadius(companionID uint64, radius float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config, exists := m.configs[companionID]; exists {
		config.Radius = clampRadius(radius)
	} else {
		config := DefaultAutoLootConfig(companionID)
		config.Radius = clampRadius(radius)
		m.configs[companionID] = config
	}
}

// ShouldCollect checks if an item should be auto-collected
func (m *AutoLootManager) ShouldCollect(companionID uint64, itemRarity int, itemType string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, exists := m.configs[companionID]
	if !exists || !config.Enabled {
		return false
	}

	if itemRarity < config.MinRarity {
		return false
	}

	for _, ignore := range config.IgnoreTypes {
		if itemType == ignore {
			return false
		}
	}

	if len(config.FilterTypes) > 0 {
		found := false
		for _, filter := range config.FilterTypes {
			if itemType == filter {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// clampRadius clamps the auto-loot radius to valid range (5-10 tiles)
func clampRadius(radius float64) float64 {
	if radius < 5.0 {
		return 5.0
	}
	if radius > 10.0 {
		return 10.0
	}
	return radius
}
