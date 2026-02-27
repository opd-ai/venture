// InstanceManager manages active raid instances with timeout and cleanup.
// This file provides instance lifecycle management, ensuring groups have
// isolated dungeon instances that expire after 4 hours.
package raids

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// InstanceManager manages active raid instances.
type InstanceManager struct {
	instances       map[string]*RaidInstance
	instanceTimeout time.Duration
	timeProvider    TimeProvider
	mu              sync.RWMutex
}

// NewInstanceManager creates a new instance manager with 4-hour timeout.
func NewInstanceManager() *InstanceManager {
	return &InstanceManager{
		instances:       make(map[string]*RaidInstance),
		instanceTimeout: 4 * time.Hour,
		timeProvider:    DefaultTimeProvider(),
	}
}

// NewInstanceManagerWithTimeout creates an instance manager with custom timeout.
func NewInstanceManagerWithTimeout(timeout time.Duration) *InstanceManager {
	return &InstanceManager{
		instances:       make(map[string]*RaidInstance),
		instanceTimeout: timeout,
		timeProvider:    DefaultTimeProvider(),
	}
}

// NewInstanceManagerWithProvider creates an instance manager with custom timeout and time provider.
// This constructor is primarily for testing with deterministic time.
func NewInstanceManagerWithProvider(timeout time.Duration, provider TimeProvider) *InstanceManager {
	return &InstanceManager{
		instances:       make(map[string]*RaidInstance),
		instanceTimeout: timeout,
		timeProvider:    provider,
	}
}

// CreateInstance creates a new raid instance for a group.
func (im *InstanceManager) CreateInstance(raid *RaidDungeon, groupID string, playerIDs []string) (*RaidInstance, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Validate group size
	if len(playerIDs) < raid.Tier.MinPlayers() {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"group_id":    groupID,
			"players":     len(playerIDs),
			"min_players": raid.Tier.MinPlayers(),
		}).Warn("insufficient players for raid")
		return nil, fmt.Errorf("insufficient players: need %d, got %d", raid.Tier.MinPlayers(), len(playerIDs))
	}
	if len(playerIDs) > raid.Tier.MaxPlayers() {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"group_id":    groupID,
			"players":     len(playerIDs),
			"max_players": raid.Tier.MaxPlayers(),
		}).Warn("too many players for raid")
		return nil, fmt.Errorf("too many players: max %d, got %d", raid.Tier.MaxPlayers(), len(playerIDs))
	}

	// Check if group already has an active instance
	existingKey := instanceKeyByGroup(groupID, raid.Tier)
	if existing, exists := im.instances[existingKey]; exists {
		if !existing.Completed && im.timeProvider.Now().Before(existing.ExpiresAt) {
			return nil, fmt.Errorf("group already has active instance")
		}
	}

	now := im.timeProvider.Now()
	instanceID := fmt.Sprintf("instance-%s-%d", groupID, now.Unix())

	instance := &RaidInstance{
		InstanceID: instanceID,
		RaidID:     raid.ID,
		Dungeon:    raid,
		GroupID:    groupID,
		PlayerIDs:  playerIDs,
		CreatedAt:  now,
		ExpiresAt:  now.Add(im.instanceTimeout),
		Completed:  false,
	}

	im.instances[instanceID] = instance
	return instance, nil
}

// GetInstance retrieves a raid instance by ID.
func (im *InstanceManager) GetInstance(instanceID string) (*RaidInstance, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instance, exists := im.instances[instanceID]
	if !exists {
		return nil, false
	}

	// Check if instance expired
	if im.timeProvider.Now().After(instance.ExpiresAt) {
		return nil, false
	}

	return instance, true
}

// GetGroupInstance retrieves the active instance for a group.
func (im *InstanceManager) GetGroupInstance(groupID string, tier RaidTier) (*RaidInstance, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	for _, instance := range im.instances {
		if instance.GroupID == groupID && instance.Dungeon.Tier == tier {
			if !instance.Completed && im.timeProvider.Now().Before(instance.ExpiresAt) {
				return instance, true
			}
		}
	}

	return nil, false
}

// CompleteInstance marks an instance as completed.
func (im *InstanceManager) CompleteInstance(instanceID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	instance, exists := im.instances[instanceID]
	if !exists {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"instance_id": instanceID,
		}).Warn("attempt to complete non-existent instance")
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	if instance.Completed {
		log.WithFields(log.Fields{
			"system_name": "raids",
			"instance_id": instanceID,
		}).Warn("attempt to complete already completed instance")
		return fmt.Errorf("instance already completed")
	}

	instance.Completed = true
	return nil
}

// RemoveInstance removes an instance from the manager.
func (im *InstanceManager) RemoveInstance(instanceID string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.instances, instanceID)
}

// CleanupExpired removes all expired instances.
func (im *InstanceManager) CleanupExpired() int {
	im.mu.Lock()
	defer im.mu.Unlock()

	now := im.timeProvider.Now()
	removed := 0

	for id, instance := range im.instances {
		if now.After(instance.ExpiresAt) {
			delete(im.instances, id)
			removed++
		}
	}

	return removed
}

// GetActiveInstanceCount returns the number of active instances.
func (im *InstanceManager) GetActiveInstanceCount() int {
	im.mu.RLock()
	defer im.mu.RUnlock()

	count := 0
	now := im.timeProvider.Now()

	for _, instance := range im.instances {
		if !instance.Completed && now.Before(instance.ExpiresAt) {
			count++
		}
	}

	return count
}

// GetGroupInstances returns all instances for a group.
func (im *InstanceManager) GetGroupInstances(groupID string) []*RaidInstance {
	im.mu.RLock()
	defer im.mu.RUnlock()

	instances := make([]*RaidInstance, 0)

	for _, instance := range im.instances {
		if instance.GroupID == groupID {
			instances = append(instances, instance)
		}
	}

	return instances
}

// instanceKeyByGroup creates a lookup key for group instances.
func instanceKeyByGroup(groupID string, tier RaidTier) string {
	return fmt.Sprintf("%s-%d", groupID, tier)
}
