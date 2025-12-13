package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine/qol"
)

func TestNewQoLSystem(t *testing.T) {
	config := qol.Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	}
	manager := qol.NewManager(config)
	system := NewQoLSystem(manager)

	if system == nil {
		t.Fatal("NewQoLSystem returned nil")
	}

	if system.manager == nil {
		t.Error("system.manager is nil")
	}

	if system.cleanupInterval != 5*time.Minute {
		t.Errorf("cleanupInterval = %v, want %v", system.cleanupInterval, 5*time.Minute)
	}

	// Verify manager accessor
	gotManager := system.Manager()
	if gotManager != manager {
		t.Error("Manager() returned different manager instance")
	}
}

func TestQoLSystemUpdate(t *testing.T) {
	manager := qol.NewManager(qol.Config{})
	system := NewQoLSystem(manager)

	// Create test entities
	entities := []*Entity{
		NewEntity(1),
		NewEntity(2),
		NewEntity(3),
	}

	// Test update with fresh system (should not cleanup yet)
	system.Update(entities, 0.016)

	// Verify lastCleanup was set
	if system.lastCleanup.IsZero() {
		t.Error("lastCleanup is zero after initialization")
	}

	// Test update with old lastCleanup (should trigger cleanup)
	system.lastCleanup = time.Now().Add(-10 * time.Minute)
	oldCleanup := system.lastCleanup

	system.Update(entities, 0.016)

	// Verify cleanup was triggered
	if !system.lastCleanup.After(oldCleanup) {
		t.Error("lastCleanup was not updated after cleanup interval")
	}
}

func TestQoLSystemCleanupExpired(t *testing.T) {
	manager := qol.NewManager(qol.Config{})
	system := NewQoLSystem(manager)

	// Add expired invitation
	expiredInv := &qol.GuildInvitation{
		InvitationID: "expired-1",
		GuildID:      "guild-1",
		InviteeID:    "player-1",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
	}
	manager.GuildInvites().SendInvitation(expiredInv)

	// Add valid invitation
	validInv := &qol.GuildInvitation{
		InvitationID: "valid-1",
		GuildID:      "guild-2",
		InviteeID:    "player-2",
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Expires tomorrow
	}
	manager.GuildInvites().SendInvitation(validInv)

	// Verify pending invitations (expired should not show as pending)
	pending1 := manager.GuildInvites().GetPendingInvitations("player-1")
	pending2 := manager.GuildInvites().GetPendingInvitations("player-2")
	if len(pending1) != 0 { // Expired should not show as pending
		t.Errorf("expired invitation showing as pending")
	}
	if len(pending2) != 1 {
		t.Errorf("valid invitation count = %d, want 1", len(pending2))
	}

	// Force cleanup by setting old lastCleanup and calling Update
	system.lastCleanup = time.Now().Add(-10 * time.Minute)
	system.Update([]*Entity{}, 0.016)

	// Verify cleanup was triggered (lastCleanup should be recent)
	if time.Since(system.lastCleanup) > 1*time.Second {
		t.Error("cleanup was not triggered after interval")
	}

	// After cleanup, expired invitations should be removed from internal storage
	// (calling CleanupExpired again should find 0 expired invitations)
	removed := manager.GuildInvites().CleanupExpired()
	if removed != 0 {
		t.Errorf("found %d expired invitations after cleanup, expected 0", removed)
	}
}

func TestQoLSystemMultipleUpdates(t *testing.T) {
	manager := qol.NewManager(qol.Config{})
	system := NewQoLSystem(manager)

	entities := []*Entity{NewEntity(1), NewEntity(2)}

	// Simulate 60 frames at 60 FPS (1 second of game time)
	for i := 0; i < 60; i++ {
		system.Update(entities, 0.016)
	}

	// Verify system is still functional
	if system.Manager() == nil {
		t.Error("Manager() returned nil after multiple updates")
	}

	// Verify lastCleanup is recent (within last second)
	if time.Since(system.lastCleanup) > 2*time.Second {
		t.Error("lastCleanup is too old after recent initialization")
	}
}

func TestQoLSystemManagerAccessors(t *testing.T) {
	config := qol.Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: false,
	}
	manager := qol.NewManager(config)
	system := NewQoLSystem(manager)

	// Test manager accessor
	gotManager := system.Manager()
	if gotManager == nil {
		t.Error("Manager() returned nil")
	}

	// Test sub-managers through system
	if gotManager.AutoLoot() == nil {
		t.Error("AutoLoot() returned nil")
	}
	if gotManager.CraftQueue() == nil {
		t.Error("CraftQueue() returned nil")
	}
	if gotManager.GuildInvites() == nil {
		t.Error("GuildInvites() returned nil")
	}
	if gotManager.MountWhistle() == nil {
		t.Error("MountWhistle() returned nil")
	}
	if gotManager.StorageSorter() == nil {
		t.Error("StorageSorter() returned nil")
	}
	if gotManager.RecipeTracker() == nil {
		t.Error("RecipeTracker() returned nil")
	}

	// Verify config is preserved
	gotConfig := gotManager.GetConfig()
	if gotConfig.AutoLoot != config.AutoLoot {
		t.Errorf("AutoLoot = %v, want %v", gotConfig.AutoLoot, config.AutoLoot)
	}
	if gotConfig.AutoSort != config.AutoSort {
		t.Errorf("AutoSort = %v, want %v", gotConfig.AutoSort, config.AutoSort)
	}
	if gotConfig.QuickDeposit != config.QuickDeposit {
		t.Errorf("QuickDeposit = %v, want %v", gotConfig.QuickDeposit, config.QuickDeposit)
	}
}

func TestQoLSystemIntegration(t *testing.T) {
	// Create system with full configuration
	manager := qol.NewManager(qol.Config{
		AutoLoot:     true,
		AutoSort:     true,
		QuickDeposit: true,
	})
	system := NewQoLSystem(manager)

	// Setup auto-loot for companion
	system.Manager().AutoLoot().SetRadius(100, 7.0)

	// Add craft queue entries
	err := system.Manager().CraftQueue().AddRecipe(1, "iron_sword", 3)
	if err != nil {
		t.Errorf("AddRecipe failed: %v", err)
	}

	// Add guild invitation
	inv := &qol.GuildInvitation{
		InvitationID: "inv-1",
		GuildID:      "guild-1",
		InviteeID:    "player-1",
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
	system.Manager().GuildInvites().SendInvitation(inv)

	// Summon mount
	summon := &qol.MountSummon{
		PlayerID:    1,
		VehicleID:   200,
		VehicleType: "horse",
		CurrentPos:  [2]float64{0, 0},
		TargetPos:   [2]float64{5, 5},
	}
	system.Manager().MountWhistle().SummonMount(summon)

	// Run system update
	entities := []*Entity{NewEntity(1)}
	system.Update(entities, 0.016)

	// Verify all features still work
	autoLootConfig := system.Manager().AutoLoot().GetConfig(100)
	if autoLootConfig.Radius != 7.0 {
		t.Errorf("AutoLoot radius = %v, want 7.0", autoLootConfig.Radius)
	}

	queue := system.Manager().CraftQueue().GetQueue(1)
	if len(queue) != 1 {
		t.Errorf("craft queue length = %d, want 1", len(queue))
	}

	pending := system.Manager().GuildInvites().GetPendingInvitations("player-1")
	if len(pending) != 1 {
		t.Errorf("pending invitations = %d, want 1", len(pending))
	}

	activeSummon := system.Manager().MountWhistle().GetActiveSummon(1)
	if activeSummon == nil {
		t.Error("active summon is nil")
	}
}
