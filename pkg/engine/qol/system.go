package qol

import (
	"sync"
)

// Config holds configuration for all QoL features
type Config struct {
	AutoLoot     bool
	AutoSort     bool
	QuickDeposit bool
}

// Manager is the unified QoL manager for all quality-of-life features.
// It combines auto-loot, crafting queues, guild invitations, mount whistles,
// storage sorting, and recipe tracking into a single system.
type Manager struct {
	config Config

	autoLoot      *AutoLootManager
	craftQueue    *CraftQueueManager
	guildInvites  *GuildInvitationManager
	mountWhistle  *MountWhistleManager
	storageSorter *StorageSorter
	recipeTracker *RecipeTracker

	mu sync.RWMutex
}

// NewManager creates a new QoL manager with the specified configuration
func NewManager(config Config) *Manager {
	return &Manager{
		config:        config,
		autoLoot:      NewAutoLootManager(),
		craftQueue:    NewCraftQueueManager(),
		guildInvites:  NewGuildInvitationManager(),
		mountWhistle:  NewMountWhistleManager(),
		storageSorter: NewStorageSorter(),
		recipeTracker: NewRecipeTracker(),
	}
}

// AutoLoot returns the auto-loot manager
func (m *Manager) AutoLoot() *AutoLootManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.autoLoot
}

// CraftQueue returns the crafting queue manager
func (m *Manager) CraftQueue() *CraftQueueManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.craftQueue
}

// GuildInvites returns the guild invitation manager
func (m *Manager) GuildInvites() *GuildInvitationManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.guildInvites
}

// MountWhistle returns the mount whistle manager
func (m *Manager) MountWhistle() *MountWhistleManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mountWhistle
}

// StorageSorter returns the storage sorter
func (m *Manager) StorageSorter() *StorageSorter {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.storageSorter
}

// RecipeTracker returns the recipe tracker
func (m *Manager) RecipeTracker() *RecipeTracker {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.recipeTracker
}

// GetConfig returns the current QoL configuration
func (m *Manager) GetConfig() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// SetConfig updates the QoL configuration
func (m *Manager) SetConfig(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}
