package raids

import (
	"testing"
	"time"
)

func TestLockoutManager_IsLockedOut(t *testing.T) {
	lm := NewLockoutManagerWithPeriod(1 * time.Hour)

	playerID := "player-1"
	tier := TierNormal

	// Not locked out initially
	if lm.IsLockedOut(playerID, tier) {
		t.Error("Player should not be locked out initially")
	}

	// Record a clear
	lm.RecordClear(playerID, tier)

	// Now locked out
	if !lm.IsLockedOut(playerID, tier) {
		t.Error("Player should be locked out after clear")
	}

	// Different tier not locked out
	if lm.IsLockedOut(playerID, TierHeroic) {
		t.Error("Player should not be locked out for different tier")
	}
}

func TestLockoutManager_RecordClear(t *testing.T) {
	lm := NewLockoutManager()

	playerID := "player-test"
	tier := TierMythic

	lm.RecordClear(playerID, tier)

	lockout, exists := lm.GetLockout(playerID, tier)
	if !exists {
		t.Fatal("Lockout should exist after RecordClear")
	}

	if lockout.PlayerID != playerID {
		t.Errorf("PlayerID = %q, want %q", lockout.PlayerID, playerID)
	}

	if lockout.Tier != tier {
		t.Errorf("Tier = %v, want %v", lockout.Tier, tier)
	}

	if time.Since(lockout.LastRun) > time.Second {
		t.Error("LastRun should be recent")
	}
}

func TestLockoutManager_GetLockout(t *testing.T) {
	lm := NewLockoutManager()

	// Non-existent lockout
	_, exists := lm.GetLockout("nonexistent", TierNormal)
	if exists {
		t.Error("GetLockout should return false for non-existent lockout")
	}

	// Create lockout
	playerID := "player-get"
	tier := TierLegendary
	lm.RecordClear(playerID, tier)

	// Retrieve lockout
	lockout, exists := lm.GetLockout(playerID, tier)
	if !exists {
		t.Fatal("GetLockout should return true for existing lockout")
	}

	if lockout.PlayerID != playerID {
		t.Errorf("PlayerID = %q, want %q", lockout.PlayerID, playerID)
	}
}

func TestLockoutManager_TimeUntilReset(t *testing.T) {
	lm := NewLockoutManagerWithPeriod(2 * time.Hour)

	playerID := "player-time"
	tier := TierNormal

	// No lockout
	if duration := lm.TimeUntilReset(playerID, tier); duration != 0 {
		t.Errorf("TimeUntilReset = %v, want 0 for non-existent lockout", duration)
	}

	// Record clear
	lm.RecordClear(playerID, tier)

	// Check time remaining
	duration := lm.TimeUntilReset(playerID, tier)
	if duration < time.Hour || duration > 2*time.Hour {
		t.Errorf("TimeUntilReset = %v, want between 1h and 2h", duration)
	}
}

func TestLockoutManager_ClearLockout(t *testing.T) {
	lm := NewLockoutManager()

	playerID := "player-clear"
	tier := TierHeroic

	// Create lockout
	lm.RecordClear(playerID, tier)

	if !lm.IsLockedOut(playerID, tier) {
		t.Error("Player should be locked out")
	}

	// Clear lockout
	lm.ClearLockout(playerID, tier)

	if lm.IsLockedOut(playerID, tier) {
		t.Error("Player should not be locked out after clearing")
	}
}

func TestLockoutManager_ClearAllLockouts(t *testing.T) {
	lm := NewLockoutManager()

	playerID := "player-clearall"

	// Create multiple lockouts
	lm.RecordClear(playerID, TierNormal)
	lm.RecordClear(playerID, TierHeroic)
	lm.RecordClear(playerID, TierMythic)

	// Clear all
	lm.ClearAllLockouts(playerID)

	// Verify all cleared
	if lm.IsLockedOut(playerID, TierNormal) {
		t.Error("TierNormal should not be locked out")
	}
	if lm.IsLockedOut(playerID, TierHeroic) {
		t.Error("TierHeroic should not be locked out")
	}
	if lm.IsLockedOut(playerID, TierMythic) {
		t.Error("TierMythic should not be locked out")
	}
}

func TestLockoutManager_ResetExpiredLockouts(t *testing.T) {
	lm := NewLockoutManagerWithPeriod(10 * time.Millisecond)

	playerID := "player-expire"

	// Create lockout
	lm.RecordClear(playerID, TierNormal)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Reset expired
	removed := lm.ResetExpiredLockouts()
	if removed != 1 {
		t.Errorf("ResetExpiredLockouts removed %d, want 1", removed)
	}

	// Verify lockout no longer active
	if lm.IsLockedOut(playerID, TierNormal) {
		t.Error("Expired lockout should be cleaned up")
	}
}

func TestLockoutManager_GetPlayerLockouts(t *testing.T) {
	lm := NewLockoutManager()

	playerID := "player-multi"

	// Create multiple lockouts
	lm.RecordClear(playerID, TierNormal)
	lm.RecordClear(playerID, TierHeroic)
	lm.RecordClear(playerID, TierMythic)

	lockouts := lm.GetPlayerLockouts(playerID)
	if len(lockouts) != 3 {
		t.Errorf("GetPlayerLockouts returned %d lockouts, want 3", len(lockouts))
	}

	// Verify tiers
	tiers := make(map[RaidTier]bool)
	for _, lockout := range lockouts {
		tiers[lockout.Tier] = true
	}

	if !tiers[TierNormal] || !tiers[TierHeroic] || !tiers[TierMythic] {
		t.Error("Not all tiers present in lockouts")
	}
}

func TestLockoutManager_GetActiveLockoutCount(t *testing.T) {
	lm := NewLockoutManagerWithPeriod(1 * time.Hour)

	if count := lm.GetActiveLockoutCount(); count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	// Add lockouts
	lm.RecordClear("player-1", TierNormal)
	lm.RecordClear("player-2", TierHeroic)
	lm.RecordClear("player-3", TierMythic)

	if count := lm.GetActiveLockoutCount(); count != 3 {
		t.Errorf("Count after adds = %d, want 3", count)
	}
}

func TestLockoutManager_Concurrency(t *testing.T) {
	lm := NewLockoutManager()

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			playerID := "concurrent-player"
			tier := RaidTier(id % 5)

			lm.RecordClear(playerID, tier)
			lm.IsLockedOut(playerID, tier)
			lm.GetLockout(playerID, tier)
			lm.TimeUntilReset(playerID, tier)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// No deadlock = success
}

func BenchmarkLockoutManager_IsLockedOut(b *testing.B) {
	lm := NewLockoutManager()
	lm.RecordClear("bench-player", TierNormal)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.IsLockedOut("bench-player", TierNormal)
	}
}

func BenchmarkLockoutManager_RecordClear(b *testing.B) {
	lm := NewLockoutManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.RecordClear("bench-player", TierNormal)
	}
}

func BenchmarkLockoutManager_GetPlayerLockouts(b *testing.B) {
	lm := NewLockoutManager()

	// Setup lockouts
	for i := 0; i < 5; i++ {
		lm.RecordClear("bench-player", RaidTier(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.GetPlayerLockouts("bench-player")
	}
}
