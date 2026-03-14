package raids

import (
	"fmt"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.generator == nil {
		t.Error("Manager has nil generator")
	}

	if manager.instanceManager == nil {
		t.Error("Manager has nil instanceManager")
	}

	if manager.lockoutManager == nil {
		t.Error("Manager has nil lockoutManager")
	}
}

func TestManager_GenerateRaid(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	tests := []struct {
		name    string
		tier    RaidTier
		depth   int
		wantErr bool
	}{
		{
			name:    "normal tier",
			tier:    TierNormal,
			depth:   5,
			wantErr: false,
		},
		{
			name:    "heroic tier",
			tier:    TierHeroic,
			depth:   10,
			wantErr: false,
		},
		{
			name:    "mythic tier",
			tier:    TierMythic,
			depth:   15,
			wantErr: false,
		},
		{
			name:    "invalid depth",
			tier:    TierNormal,
			depth:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raid, err := manager.GenerateRaid(tt.tier, tt.depth)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRaid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if raid == nil {
					t.Error("GenerateRaid() returned nil raid")
				}

				if raid.Tier != tt.tier {
					t.Errorf("raid tier = %v, want %v", raid.Tier, tt.tier)
				}

				if len(raid.Bosses) == 0 {
					t.Error("raid has no bosses")
				}
			}
		})
	}
}

func TestManager_CreateInstance(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	tests := []struct {
		name      string
		tier      RaidTier
		depth     int
		groupID   string
		playerIDs []string
		wantErr   bool
	}{
		{
			name:      "valid normal raid",
			tier:      TierNormal,
			depth:     5,
			groupID:   "group1",
			playerIDs: []string{"p1", "p2", "p3", "p4", "p5"},
			wantErr:   false,
		},
		{
			name:      "valid heroic raid",
			tier:      TierHeroic,
			depth:     10,
			groupID:   "group2",
			playerIDs: []string{"p1", "p2", "p3", "p4", "p5", "p6"},
			wantErr:   false,
		},
		{
			name:      "insufficient players",
			tier:      TierNormal,
			depth:     5,
			groupID:   "group3",
			playerIDs: []string{"p1", "p2"},
			wantErr:   true,
		},
		{
			name:      "too many players",
			tier:      TierNormal,
			depth:     5,
			groupID:   "group4",
			playerIDs: []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, err := manager.CreateInstance(tt.tier, tt.depth, tt.groupID, tt.playerIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if instance == nil {
					t.Error("CreateInstance() returned nil instance")
				}

				if instance.GroupID != tt.groupID {
					t.Errorf("instance groupID = %v, want %v", instance.GroupID, tt.groupID)
				}

				if len(instance.PlayerIDs) != len(tt.playerIDs) {
					t.Errorf("instance has %d players, want %d", len(instance.PlayerIDs), len(tt.playerIDs))
				}
			}
		})
	}
}

func TestManager_CompleteRaid(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	// Create instance
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
	instance, err := manager.CreateInstance(TierNormal, 5, "group1", playerIDs)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	// Complete raid
	err = manager.CompleteRaid(instance.InstanceID)
	if err != nil {
		t.Fatalf("CompleteRaid() error = %v", err)
	}

	// Verify instance is marked complete
	completedInstance, exists := manager.GetInstance(instance.InstanceID)
	if !exists {
		t.Fatal("Instance not found after completion")
	}

	if !completedInstance.Completed {
		t.Error("Instance not marked as completed")
	}

	// Verify all players have lockouts
	for _, playerID := range playerIDs {
		lockouts := manager.GetPlayerLockouts(playerID)
		if len(lockouts) == 0 {
			t.Errorf("Player %s has no lockouts after raid completion", playerID)
		}
	}

	// Try to complete again (should fail)
	err = manager.CompleteRaid(instance.InstanceID)
	if err == nil {
		t.Error("CompleteRaid() should fail on already completed instance")
	}
}

func TestManager_CanParticipate(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}

	// Initially all players can participate
	canParticipate, locked := manager.CanParticipate(playerIDs, TierNormal)
	if !canParticipate {
		t.Error("Players should be able to participate initially")
	}

	if len(locked) != 0 {
		t.Errorf("Got %d locked players, want 0", len(locked))
	}

	// Create and complete a raid
	instance, err := manager.CreateInstance(TierNormal, 5, "group1", playerIDs)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	err = manager.CompleteRaid(instance.InstanceID)
	if err != nil {
		t.Fatalf("Failed to complete raid: %v", err)
	}

	// Now players should be locked out
	canParticipate, locked = manager.CanParticipate(playerIDs, TierNormal)
	if canParticipate {
		t.Error("Players should not be able to participate after lockout")
	}

	if len(locked) != len(playerIDs) {
		t.Errorf("Got %d locked players, want %d", len(locked), len(playerIDs))
	}

	// Different tier should still be available
	canParticipate, locked = manager.CanParticipate(playerIDs, TierHeroic)
	if !canParticipate {
		t.Error("Players should be able to participate in different tier")
	}
}

func TestManager_GetPlayerLockouts(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	playerID := "player1"

	// Initially no lockouts
	lockouts := manager.GetPlayerLockouts(playerID)
	if len(lockouts) != 0 {
		t.Errorf("Got %d lockouts, want 0", len(lockouts))
	}

	// Create and complete raids at multiple tiers
	tiers := []RaidTier{TierNormal, TierHeroic, TierMythic}

	for i, tier := range tiers {
		// Adjust player count based on tier requirements
		minPlayers := tier.MinPlayers()
		playerIDs := []string{playerID}
		for j := 1; j < minPlayers; j++ {
			playerIDs = append(playerIDs, fmt.Sprintf("p%d", j+1))
		}

		instance, err := manager.CreateInstance(tier, 5+i, fmt.Sprintf("group%d", i+1), playerIDs)
		if err != nil {
			t.Fatalf("Failed to create instance for tier %v: %v", tier, err)
		}

		err = manager.CompleteRaid(instance.InstanceID)
		if err != nil {
			t.Fatalf("Failed to complete raid for tier %v: %v", tier, err)
		}
	}

	// Should have 3 lockouts
	lockouts = manager.GetPlayerLockouts(playerID)
	if len(lockouts) != 3 {
		t.Errorf("Got %d lockouts, want 3", len(lockouts))
	}

	// Verify each tier is represented
	tierMap := make(map[RaidTier]bool)
	for _, lockout := range lockouts {
		tierMap[lockout.Tier] = true
	}

	for _, tier := range tiers {
		if !tierMap[tier] {
			t.Errorf("Missing lockout for tier %v", tier)
		}
	}
}

func TestManager_CleanupExpired(t *testing.T) {
	// Use short lockout period for testing
	manager := NewManager(12345, "fantasy")
	manager.lockoutManager = NewLockoutManagerWithPeriod(100 * time.Millisecond)
	manager.instanceManager = NewInstanceManagerWithTimeout(100 * time.Millisecond)

	// Create instances and lockouts
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
	instance, err := manager.CreateInstance(TierNormal, 5, "group1", playerIDs)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	err = manager.CompleteRaid(instance.InstanceID)
	if err != nil {
		t.Fatalf("Failed to complete raid: %v", err)
	}

	// Initially should have 0 active instances (completed instances don't count as active)
	if count := manager.GetActiveInstanceCount(); count != 0 {
		t.Errorf("Got %d active instances, want 0 (completed instances don't count)", count)
	}

	if count := manager.GetActiveLockoutCount(); count != 5 {
		t.Errorf("Got %d active lockouts, want 5", count)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Cleanup
	instances, lockouts := manager.CleanupExpired()

	// The instance was already completed, so it might or might not be cleaned up
	// depending on implementation. What matters is lockouts are cleaned.
	t.Logf("Cleaned up %d instances and %d lockouts", instances, lockouts)

	if lockouts != 5 {
		t.Errorf("Cleaned up %d lockouts, want 5", lockouts)
	}

	// Verify counts after cleanup
	if count := manager.GetActiveInstanceCount(); count != 0 {
		t.Errorf("Got %d active instances after cleanup, want 0", count)
	}

	if count := manager.GetActiveLockoutCount(); count != 0 {
		t.Errorf("Got %d active lockouts after cleanup, want 0", count)
	}
}

func TestManager_GetGroupInstance(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	groupID := "group1"
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}

	// Create instance
	instance, err := manager.CreateInstance(TierNormal, 5, groupID, playerIDs)
	if err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	// Retrieve by group
	retrieved, exists := manager.GetGroupInstance(groupID, TierNormal)
	if !exists {
		t.Fatal("GetGroupInstance() returned false for existing instance")
	}

	if retrieved.InstanceID != instance.InstanceID {
		t.Errorf("Got instance ID %v, want %v", retrieved.InstanceID, instance.InstanceID)
	}

	// Different tier should not exist
	_, exists = manager.GetGroupInstance(groupID, TierHeroic)
	if exists {
		t.Error("GetGroupInstance() should return false for non-existent tier")
	}

	// Different group should not exist
	_, exists = manager.GetGroupInstance("group2", TierNormal)
	if exists {
		t.Error("GetGroupInstance() should return false for non-existent group")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager := NewManager(12345, "fantasy")

	done := make(chan bool)

	// Concurrent instance creation
	for i := 0; i < 10; i++ {
		go func(id int) {
			playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}
			groupID := string(rune('a' + id))
			_, err := manager.CreateInstance(TierNormal, 5, groupID, playerIDs)
			if err != nil {
				t.Errorf("Concurrent CreateInstance failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify count
	count := manager.GetActiveInstanceCount()
	if count != 10 {
		t.Errorf("Got %d active instances, want 10", count)
	}
}

// Benchmarks

func BenchmarkManager_GenerateRaid(b *testing.B) {
	manager := NewManager(12345, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.GenerateRaid(TierNormal, 5)
	}
}

func BenchmarkManager_CreateInstance(b *testing.B) {
	manager := NewManager(12345, "fantasy")
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		groupID := string(rune(i % 1000))
		_, _ = manager.CreateInstance(TierNormal, 5, groupID, playerIDs)
	}
}

func BenchmarkManager_CanParticipate(b *testing.B) {
	manager := NewManager(12345, "fantasy")
	playerIDs := []string{"p1", "p2", "p3", "p4", "p5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CanParticipate(playerIDs, TierNormal)
	}
}

func BenchmarkManager_GetPlayerLockouts(b *testing.B) {
	manager := NewManager(12345, "fantasy")

	// Create some lockouts
	playerIDs := []string{"player1", "p2", "p3", "p4", "p5"}
	instance, _ := manager.CreateInstance(TierNormal, 5, "group1", playerIDs)
	manager.CompleteRaid(instance.InstanceID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetPlayerLockouts("player1")
	}
}
