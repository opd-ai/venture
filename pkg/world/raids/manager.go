// Manager provides a unified interface for raid generation, instances, and lockouts.
// This file coordinates between Generator, InstanceManager, and LockoutManager
// to provide a single entry point for all raid operations.
package raids

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// Manager provides a unified interface for raid generation, instance management, and lockout tracking.
type Manager struct {
	generator       *Generator
	instanceManager *InstanceManager
	lockoutManager  *LockoutManager
	genreID         string
	mu              sync.RWMutex
}

// NewManager creates a new raid manager with default settings.
// genreID sets the genre for raid content generation (e.g., "fantasy", "sci-fi").
// Pass an empty string to use the default "fantasy" genre.
func NewManager(seed int64, genreID string) *Manager {
	if genreID == "" {
		genreID = "fantasy"
	}
	return &Manager{
		generator:       NewGenerator(seed),
		instanceManager: NewInstanceManager(),
		lockoutManager:  NewLockoutManager(),
		genreID:         genreID,
	}
}

// GenerateRaid generates a new raid dungeon for the specified tier and depth.
func (m *Manager) GenerateRaid(tier RaidTier, depth int) (*RaidDungeon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	params := procgen.GenerationParams{
		Difficulty: tier.DifficultyMultiplier() / 10.0, // Scale to 0.0-1.0 range
		Depth:      depth,
		GenreID:    m.genreID,
		Custom: map[string]interface{}{
			"tier": tier,
		},
	}

	result, err := m.generator.Generate(m.generator.baseSeed+int64(tier), params)
	if err != nil {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"tier":        tier.String(),
			"depth":       depth,
		}).WithError(err).Error("failed to generate raid dungeon")
		return nil, fmt.Errorf("failed to generate raid: %w", err)
	}

	raid, ok := result.(*RaidDungeon)
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"result_type": fmt.Sprintf("%T", result),
		}).Error("generator returned invalid type")
		return nil, fmt.Errorf("generator returned invalid type")
	}

	return raid, nil
}

// CreateInstance creates a new raid instance for a group if all players can participate.
func (m *Manager) CreateInstance(tier RaidTier, depth int, groupID string, playerIDs []string) (*RaidInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check lockouts for all players
	for _, playerID := range playerIDs {
		if m.lockoutManager.IsLockedOut(playerID, tier) {
			log.WithFields(log.Fields{
				"system_name": "raids",
				"playerID":    playerID,
				"tier":        tier.String(),
			}).Warn("player locked out from raid")
			return nil, fmt.Errorf("player %s is locked out from %s raids", playerID, tier.String())
		}
	}

	// Generate raid dungeon
	params := procgen.GenerationParams{
		Difficulty: tier.DifficultyMultiplier() / 10.0, // Scale to 0.0-1.0 range
		Depth:      depth,
		GenreID:    m.genreID,
		Custom: map[string]interface{}{
			"tier":       tier,
			"group_id":   groupID,
			"group_size": len(playerIDs),
		},
	}

	result, err := m.generator.Generate(m.generator.baseSeed+int64(tier)+int64(depth), params)
	if err != nil {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"tier":        tier.String(),
			"depth":       depth,
			"group_id":    groupID,
		}).WithError(err).Error("failed to generate raid for instance")
		return nil, fmt.Errorf("failed to generate raid: %w", err)
	}

	raid, ok := result.(*RaidDungeon)
	if !ok {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"result_type": fmt.Sprintf("%T", result),
		}).Error("generator returned invalid type for instance")
		return nil, fmt.Errorf("generator returned invalid type")
	}

	// Create instance
	instance, err := m.instanceManager.CreateInstance(raid, groupID, playerIDs)
	if err != nil {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"group_id":    groupID,
			"tier":        tier.String(),
		}).WithError(err).Error("failed to create raid instance")
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	return instance, nil
}

// GetInstance retrieves an active raid instance by ID.
func (m *Manager) GetInstance(instanceID string) (*RaidInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.instanceManager.GetInstance(instanceID)
}

// CompleteRaid marks a raid instance as complete and applies lockouts to all participants.
func (m *Manager) CompleteRaid(instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instanceManager.GetInstance(instanceID)
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"instance_id": instanceID,
		}).Warn("attempt to complete non-existent raid instance")
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	if instance.Completed {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"instance_id": instanceID,
		}).Warn("attempt to complete already completed raid")
		return fmt.Errorf("instance already completed")
	}

	// Mark instance as completed
	if err := m.instanceManager.CompleteInstance(instanceID); err != nil {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"instance_id": instanceID,
		}).WithError(err).Error("failed to mark instance as completed")
		return fmt.Errorf("failed to complete instance: %w", err)
	}

	// Apply lockouts to all players
	for _, playerID := range instance.PlayerIDs {
		m.lockoutManager.RecordClear(playerID, instance.Dungeon.Tier)
	}

	return nil
}

// CanParticipate checks if all players in a group can participate in a raid tier.
func (m *Manager) CanParticipate(playerIDs []string, tier RaidTier) (bool, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lockedPlayers := make([]string, 0)

	for _, playerID := range playerIDs {
		if m.lockoutManager.IsLockedOut(playerID, tier) {
			lockedPlayers = append(lockedPlayers, playerID)
		}
	}

	return len(lockedPlayers) == 0, lockedPlayers
}

// GetPlayerLockouts retrieves all raid lockouts for a player.
func (m *Manager) GetPlayerLockouts(playerID string) []*PlayerLockout {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lockoutManager.GetPlayerLockouts(playerID)
}

// CleanupExpired removes expired instances and lockouts.
func (m *Manager) CleanupExpired() (instances, lockouts int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instances = m.instanceManager.CleanupExpired()
	lockouts = m.lockoutManager.ResetExpiredLockouts()
	return instances, lockouts
}

// GetActiveInstanceCount returns the number of currently active raid instances.
func (m *Manager) GetActiveInstanceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.instanceManager.GetActiveInstanceCount()
}

// GetActiveLockoutCount returns the number of active player lockouts.
func (m *Manager) GetActiveLockoutCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lockoutManager.GetActiveLockoutCount()
}

// GetGroupInstance retrieves the active instance for a group at a specific tier.
func (m *Manager) GetGroupInstance(groupID string, tier RaidTier) (*RaidInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.instanceManager.GetGroupInstance(groupID, tier)
}
