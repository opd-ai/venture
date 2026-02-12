package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/engine/qol"
)

// QoLSystemWrapper wraps the qol.Manager to implement the engine.System interface
type QoLSystemWrapper struct {
	manager         *qol.Manager
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// NewQoLSystem creates a new QoL system wrapper
func NewQoLSystem(manager *qol.Manager) *QoLSystemWrapper {
	return &QoLSystemWrapper{
		manager:         manager,
		lastCleanup:     time.Now(),
		cleanupInterval: 5 * time.Minute, // Cleanup expired invitations every 5 minutes
	}
}

// Update implements the engine.System interface.
// It performs periodic cleanup of expired guild invitations.
func (s *QoLSystemWrapper) Update(entities []*Entity, deltaTime float64) {
	// Guard against nil receiver (can happen during lazy initialization)
	if s == nil || s.manager == nil {
		return
	}

	// Periodic cleanup of expired guild invitations
	if time.Since(s.lastCleanup) > s.cleanupInterval {
		s.manager.GuildInvites().CleanupExpired()
		s.lastCleanup = time.Now()
	}

	// Auto-loot is handled by companion AI systems which query the manager
	// Craft queue is handled by crafting system
	// Mount whistle is handled by vehicle systems
	// Storage sorting is player-initiated, not frame-based
	// Recipe tracking is updated by inventory system
}

// Manager returns the underlying qol.Manager for direct access
func (s *QoLSystemWrapper) Manager() *qol.Manager {
	return s.manager
}
