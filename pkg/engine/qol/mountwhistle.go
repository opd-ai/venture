// Package qol - mountwhistle.go
// This file contains the MountWhistleManager implementation for vehicle summoning.
// Code relocated from: manager.go

package qol

import (
	"math"
	"sync"
	"time"
)

// MountWhistleManager manages vehicle summoning
type MountWhistleManager struct {
	summons map[uint64]*MountSummon
	mu      sync.RWMutex
}

// NewMountWhistleManager creates a new mount whistle manager
func NewMountWhistleManager() *MountWhistleManager {
	return &MountWhistleManager{
		summons: make(map[uint64]*MountSummon),
	}
}

// SummonMount summons a vehicle to the player
func (m *MountWhistleManager) SummonMount(summon *MountSummon) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summon.Distance = math.Sqrt(
		math.Pow(summon.TargetPos[0]-summon.CurrentPos[0], 2) +
			math.Pow(summon.TargetPos[1]-summon.CurrentPos[1], 2),
	)
	summon.EstimatedTime = EstimateArrivalTime(summon.Distance)
	summon.RequestTime = time.Now()
	summon.Completed = false

	m.summons[summon.PlayerID] = summon
}

// GetActiveSummon retrieves the active summon for a player
func (m *MountWhistleManager) GetActiveSummon(playerID uint64) *MountSummon {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summons[playerID]
}

// CompleteSummon marks a summon as completed
func (m *MountWhistleManager) CompleteSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if summon, exists := m.summons[playerID]; exists {
		summon.Completed = true
	}
}

// CancelSummon cancels an active summon
func (m *MountWhistleManager) CancelSummon(playerID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.summons, playerID)
}
