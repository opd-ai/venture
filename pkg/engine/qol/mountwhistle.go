// Package qol - mountwhistle.go
// This file contains the MountWhistleManager implementation for vehicle summoning.
// Code relocated from: manager.go

package qol

import (
	"math"
	"sync"
	"time"
)

// MountWhistleManager manages vehicle summoning via whistle mechanic.
// Tracks active summon requests and their estimated arrival times.
type MountWhistleManager struct {
	summons map[uint64]*MountSummon
	mu      sync.RWMutex
}

// NewMountWhistleManager creates a new mount whistle manager.
func NewMountWhistleManager() *MountWhistleManager {
	return &MountWhistleManager{
		summons: make(map[uint64]*MountSummon),
	}
}

// SummonMount initiates a vehicle summon to the player's location.
// Calculates distance and estimated arrival time based on current and target positions.
// The RequestTime uses time.Now() for accurate summon timing (real-time gameplay feature).
func (m *MountWhistleManager) SummonMount(summon *MountSummon) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summon.Distance = math.Sqrt(
		math.Pow(summon.TargetPos[0]-summon.CurrentPos[0], 2) +
			math.Pow(summon.TargetPos[1]-summon.CurrentPos[1], 2),
	)
	summon.EstimatedTime = EstimateArrivalTime(summon.Distance)
	summon.RequestTime = time.Now() // Real-time timestamp for summon timing
	summon.Completed = false

	m.summons[summon.PlayerID] = summon
}

// GetActiveSummon retrieves the current active summon for a player.
// Returns nil if the player has no active summon request.
func (m *MountWhistleManager) GetActiveSummon(playerID uint64) *MountSummon {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summons[playerID]
}

// CompleteSummon marks a summon as completed when the vehicle arrives.
func (m *MountWhistleManager) CompleteSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if summon, exists := m.summons[playerID]; exists {
		summon.Completed = true
	}
}

// CancelSummon cancels an active summon request for a player.
func (m *MountWhistleManager) CancelSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.summons, playerID)
}
