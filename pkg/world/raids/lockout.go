// LockoutManager manages raid lockouts for players with 7-day reset periods.
// This file tracks which players have completed which raid tiers and when
// they can participate again.
package raids

import (
	"fmt"
	"sync"
	"time"
)

// LockoutManager manages raid lockouts for players.
type LockoutManager struct {
	lockouts      map[string]*PlayerLockout // key: playerID-tier
	lockoutPeriod time.Duration
	timeProvider  TimeProvider
	mu            sync.RWMutex
}

// NewLockoutManager creates a new lockout manager with the default 7-day lockout period.
func NewLockoutManager() *LockoutManager {
	return &LockoutManager{
		lockouts:      make(map[string]*PlayerLockout),
		lockoutPeriod: 7 * 24 * time.Hour, // 1 week
		timeProvider:  DefaultTimeProvider(),
	}
}

// NewLockoutManagerWithPeriod creates a lockout manager with a custom period.
func NewLockoutManagerWithPeriod(period time.Duration) *LockoutManager {
	return &LockoutManager{
		lockouts:      make(map[string]*PlayerLockout),
		lockoutPeriod: period,
		timeProvider:  DefaultTimeProvider(),
	}
}

// NewLockoutManagerWithProvider creates a lockout manager with custom period and time provider.
// This constructor is primarily for testing with deterministic time.
func NewLockoutManagerWithProvider(period time.Duration, provider TimeProvider) *LockoutManager {
	return &LockoutManager{
		lockouts:      make(map[string]*PlayerLockout),
		lockoutPeriod: period,
		timeProvider:  provider,
	}
}

// IsLockedOut checks if a player is locked out from a raid tier.
func (lm *LockoutManager) IsLockedOut(playerID string, tier RaidTier) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	key := lockoutKey(playerID, tier)
	lockout, exists := lm.lockouts[key]
	if !exists {
		return false
	}

	// Check if lockout has expired
	return lm.timeProvider.Now().Before(lockout.NextReset)
}

// RecordClear records that a player cleared a raid tier.
func (lm *LockoutManager) RecordClear(playerID string, tier RaidTier) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := lm.timeProvider.Now()
	key := lockoutKey(playerID, tier)

	lockout := &PlayerLockout{
		PlayerID:  playerID,
		Tier:      tier,
		LastRun:   now,
		NextReset: now.Add(lm.lockoutPeriod),
	}

	lm.lockouts[key] = lockout
}

// GetLockout retrieves a player's lockout information for a tier.
func (lm *LockoutManager) GetLockout(playerID string, tier RaidTier) (*PlayerLockout, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	key := lockoutKey(playerID, tier)
	lockout, exists := lm.lockouts[key]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	lockoutCopy := *lockout
	return &lockoutCopy, true
}

// TimeUntilReset returns the time until a player's lockout resets for a tier.
func (lm *LockoutManager) TimeUntilReset(playerID string, tier RaidTier) time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	key := lockoutKey(playerID, tier)
	lockout, exists := lm.lockouts[key]
	if !exists {
		return 0
	}

	remaining := time.Until(lockout.NextReset)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ClearLockout removes a player's lockout for a tier (admin function).
func (lm *LockoutManager) ClearLockout(playerID string, tier RaidTier) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	key := lockoutKey(playerID, tier)
	delete(lm.lockouts, key)
}

// ClearAllLockouts removes all lockouts for a player (admin function).
func (lm *LockoutManager) ClearAllLockouts(playerID string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	for key := range lm.lockouts {
		if lockout, exists := lm.lockouts[key]; exists && lockout.PlayerID == playerID {
			delete(lm.lockouts, key)
		}
	}
}

// ResetExpiredLockouts removes all expired lockouts from memory.
func (lm *LockoutManager) ResetExpiredLockouts() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := lm.timeProvider.Now()
	removed := 0

	for key, lockout := range lm.lockouts {
		if now.After(lockout.NextReset) {
			delete(lm.lockouts, key)
			removed++
		}
	}

	return removed
}

// GetPlayerLockouts returns all lockouts for a player.
func (lm *LockoutManager) GetPlayerLockouts(playerID string) []*PlayerLockout {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lockouts := make([]*PlayerLockout, 0)

	for _, lockout := range lm.lockouts {
		if lockout.PlayerID == playerID {
			lockoutCopy := *lockout
			lockouts = append(lockouts, &lockoutCopy)
		}
	}

	return lockouts
}

// GetActiveLockoutCount returns the total number of active lockouts.
func (lm *LockoutManager) GetActiveLockoutCount() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	count := 0
	now := lm.timeProvider.Now()

	for _, lockout := range lm.lockouts {
		if now.Before(lockout.NextReset) {
			count++
		}
	}

	return count
}

// lockoutKey creates a unique key for a player-tier combination.
func lockoutKey(playerID string, tier RaidTier) string {
	return fmt.Sprintf("%s-%d", playerID, tier)
}
