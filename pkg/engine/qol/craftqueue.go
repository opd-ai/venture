// Package qol - craftqueue.go
// This file contains the CraftQueueManager implementation for smart crafting queues.
// Code relocated from: manager.go

package qol

import (
	"fmt"
	"sync"
	"time"
)

// CraftQueueManager manages crafting recipe queue
type CraftQueueManager struct {
	queue map[uint64][]*CraftQueueEntry
	mu    sync.RWMutex
}

// NewCraftQueueManager creates a new craft queue manager
func NewCraftQueueManager() *CraftQueueManager {
	return &CraftQueueManager{
		queue: make(map[uint64][]*CraftQueueEntry),
	}
}

// AddRecipe adds a recipe to the player's craft queue
func (m *CraftQueueManager) AddRecipe(playerID uint64, recipeID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	if _, exists := m.queue[playerID]; !exists {
		m.queue[playerID] = make([]*CraftQueueEntry, 0)
	}

	if len(m.queue[playerID]) >= 50 {
		return fmt.Errorf("craft queue full (max 50 recipes)")
	}

	entry := &CraftQueueEntry{
		RecipeID:       recipeID,
		Quantity:       quantity,
		MaterialsReady: false,
		Position:       len(m.queue[playerID]),
		AddedAt:        time.Now(),
	}

	m.queue[playerID] = append(m.queue[playerID], entry)
	return nil
}

// RemoveRecipe removes a recipe from the queue
func (m *CraftQueueManager) RemoveRecipe(playerID uint64, position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queue[playerID]
	if !exists || position < 0 || position >= len(queue) {
		return fmt.Errorf("invalid queue position")
	}

	m.queue[playerID] = append(queue[:position], queue[position+1:]...)

	for i := range m.queue[playerID] {
		m.queue[playerID][i].Position = i
	}

	return nil
}

// GetQueue retrieves the player's craft queue
func (m *CraftQueueManager) GetQueue(playerID uint64) []*CraftQueueEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queue, exists := m.queue[playerID]
	if !exists {
		return make([]*CraftQueueEntry, 0)
	}

	result := make([]*CraftQueueEntry, len(queue))
	copy(result, queue)
	return result
}

// ClearQueue clears the player's craft queue
func (m *CraftQueueManager) ClearQueue(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.queue, playerID)
}
