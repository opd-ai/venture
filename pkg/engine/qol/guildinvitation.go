// Package qol - guildinvitation.go
// This file contains the GuildInvitationManager implementation for offline guild invitations.
// Code relocated from: manager.go

package qol

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// GuildInvitationManager manages offline guild invitations
type GuildInvitationManager struct {
	invitations map[string]*GuildInvitation // invitationID -> invitation
	byInvitee   map[string][]string         // inviteeID -> invitationIDs
	mu          sync.RWMutex
}

// NewGuildInvitationManager creates a new guild invitation manager
func NewGuildInvitationManager() *GuildInvitationManager {
	return &GuildInvitationManager{
		invitations: make(map[string]*GuildInvitation),
		byInvitee:   make(map[string][]string),
	}
}

// SendInvitation creates and stores a new guild invitation.
// If ExpiresAt is not set, defaults to 7 days from now (real-time expiry for gameplay).
// If SentAt is not set, defaults to time.Now() (real-time timestamp for UI display).
func (m *GuildInvitationManager) SendInvitation(inv *GuildInvitation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Note: time.Now() is intentional here as guild invitations use real-time expiry
	// for gameplay mechanics (7-day offline invitation window)
	if inv.ExpiresAt.IsZero() {
		inv.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	if inv.SentAt.IsZero() {
		inv.SentAt = time.Now()
	}

	m.invitations[inv.InvitationID] = inv

	if _, exists := m.byInvitee[inv.InviteeID]; !exists {
		m.byInvitee[inv.InviteeID] = make([]string, 0)
	}
	m.byInvitee[inv.InviteeID] = append(m.byInvitee[inv.InviteeID], inv.InvitationID)
}

// GetPendingInvitations retrieves all non-expired, non-accepted invitations for a player.
// Returns an empty slice if the player has no pending invitations.
func (m *GuildInvitationManager) GetPendingInvitations(playerID string) []*GuildInvitation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	invIDs, exists := m.byInvitee[playerID]
	if !exists {
		return make([]*GuildInvitation, 0)
	}

	result := make([]*GuildInvitation, 0)
	for _, invID := range invIDs {
		inv, exists := m.invitations[invID]
		if exists && !inv.Accepted && !inv.IsExpired() {
			result = append(result, inv)
		}
	}

	return result
}

// AcceptInvitation marks a guild invitation as accepted.
// Returns an error if the invitation is not found, already accepted, or expired.
// The AcceptedAt timestamp uses time.Now() for accurate acceptance tracking.
func (m *GuildInvitationManager) AcceptInvitation(invitationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv, exists := m.invitations[invitationID]
	if !exists {
		log.WithFields(log.Fields{
			"invitation_id": invitationID,
		}).Warn("guild invitation: not found")
		return fmt.Errorf("invitation not found")
	}

	if inv.Accepted {
		log.WithFields(log.Fields{
			"invitation_id": invitationID,
			"guild_id":      inv.GuildID,
		}).Warn("guild invitation: already accepted")
		return fmt.Errorf("invitation already accepted")
	}

	if inv.IsExpired() {
		log.WithFields(log.Fields{
			"invitation_id": invitationID,
			"guild_id":      inv.GuildID,
			"expired_at":    inv.ExpiresAt,
		}).Warn("guild invitation: expired")
		return fmt.Errorf("invitation expired")
	}

	inv.Accepted = true
	inv.AcceptedAt = time.Now() // Real-time timestamp for acceptance tracking
	return nil
}

// CleanupExpired removes all expired invitations from the manager.
// Returns the number of invitations removed.
func (m *GuildInvitationManager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for invID, inv := range m.invitations {
		if inv.IsExpired() {
			delete(m.invitations, invID)
			removed++
		}
	}

	for inviteeID, invIDs := range m.byInvitee {
		newIDs := make([]string, 0)
		for _, invID := range invIDs {
			if _, exists := m.invitations[invID]; exists {
				newIDs = append(newIDs, invID)
			}
		}
		if len(newIDs) == 0 {
			delete(m.byInvitee, inviteeID)
		} else {
			m.byInvitee[inviteeID] = newIDs
		}
	}

	if removed > 0 {
		log.WithFields(log.Fields{
			"removed_count": removed,
		}).Debug("guild invitations: cleaned up expired")
	}

	return removed
}
