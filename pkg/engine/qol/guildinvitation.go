// Package qol - guildinvitation.go
// This file contains the GuildInvitationManager implementation for offline guild invitations.
// Code relocated from: manager.go

package qol

import (
	"fmt"
	"sync"
	"time"
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

// SendInvitation creates a new guild invitation
func (m *GuildInvitationManager) SendInvitation(inv *GuildInvitation) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

// GetPendingInvitations retrieves pending invitations for a player
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

// AcceptInvitation accepts a guild invitation
func (m *GuildInvitationManager) AcceptInvitation(invitationID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv, exists := m.invitations[invitationID]
	if !exists {
		return fmt.Errorf("invitation not found")
	}

	if inv.Accepted {
		return fmt.Errorf("invitation already accepted")
	}

	if inv.IsExpired() {
		return fmt.Errorf("invitation expired")
	}

	inv.Accepted = true
	inv.AcceptedAt = time.Now()
	return nil
}

// CleanupExpired removes expired invitations
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

	return removed
}
