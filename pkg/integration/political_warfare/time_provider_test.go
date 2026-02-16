package political_warfare

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

// TestTimeProviderInterface verifies the TimeProvider implementations.
func TestTimeProviderInterface(t *testing.T) {
	t.Run("RealTimeProvider returns current time", func(t *testing.T) {
		provider := RealTimeProvider{}
		before := time.Now()
		result := provider.Now()
		after := time.Now()

		if result.Before(before) || result.After(after) {
			t.Errorf("RealTimeProvider.Now() returned time outside expected range")
		}
	})

	t.Run("FixedTimeProvider returns fixed time", func(t *testing.T) {
		fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
		provider := FixedTimeProvider{FixedTime: fixedTime}

		result1 := provider.Now()
		result2 := provider.Now()

		if !result1.Equal(fixedTime) {
			t.Errorf("Expected fixed time %v, got %v", fixedTime, result1)
		}
		if !result1.Equal(result2) {
			t.Errorf("FixedTimeProvider should return same time on every call")
		}
	})

	t.Run("SetTimeProvider and ResetTimeProvider", func(t *testing.T) {
		fixedTime := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
		defer ResetTimeProvider()

		result := now()
		if !result.Equal(fixedTime) {
			t.Errorf("Expected fixed time %v from now(), got %v", fixedTime, result)
		}
	})

	t.Run("ResetTimeProvider restores real time", func(t *testing.T) {
		fixedTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
		ResetTimeProvider()

		result := now()
		if result.Year() < 2024 {
			t.Errorf("After reset, now() should return real time, got %v", result)
		}
	})
}

// TestDeclareWarDeterministicTimestamps verifies war declaration uses TimeProvider.
func TestDeclareWarDeterministicTimestamps(t *testing.T) {
	fixedTime := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)
	prepPeriod := 24 * time.Hour

	war, err := manager.DeclareWar(guildID1, guildID2, prepPeriod)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	if !war.DeclaredAt.Equal(fixedTime) {
		t.Errorf("Expected DeclaredAt %v, got %v", fixedTime, war.DeclaredAt)
	}
	expectedPrep := fixedTime.Add(prepPeriod)
	if !war.PreparationEnds.Equal(expectedPrep) {
		t.Errorf("Expected PreparationEnds %v, got %v", expectedPrep, war.PreparationEnds)
	}
}

// TestSignPeaceTreatyDeterministicTimestamps verifies treaty uses TimeProvider.
func TestSignPeaceTreatyDeterministicTimestamps(t *testing.T) {
	fixedTime := time.Date(2026, 4, 1, 8, 30, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)
	duration := 7 * 24 * time.Hour

	treaty, err := manager.SignPeaceTreaty(guildID1, guildID2, duration)
	if err != nil {
		t.Fatalf("SignPeaceTreaty failed: %v", err)
	}

	if !treaty.SignedAt.Equal(fixedTime) {
		t.Errorf("Expected SignedAt %v, got %v", fixedTime, treaty.SignedAt)
	}
	if !treaty.ExpiresAt.Equal(fixedTime.Add(duration)) {
		t.Errorf("Expected ExpiresAt %v, got %v", fixedTime.Add(duration), treaty.ExpiresAt)
	}
	if !treaty.CooldownEnds.Equal(fixedTime.Add(duration)) {
		t.Errorf("Expected CooldownEnds %v, got %v", fixedTime.Add(duration), treaty.CooldownEnds)
	}
}

// TestImposeEmbargoDeterministicTimestamps verifies embargo uses TimeProvider.
func TestImposeEmbargoDeterministicTimestamps(t *testing.T) {
	fixedTime := time.Date(2026, 5, 10, 14, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)

	embargo, err := manager.ImposeEmbargo(guildID1, guildID2, 0.75)
	if err != nil {
		t.Fatalf("ImposeEmbargo failed: %v", err)
	}

	if !embargo.ImposedAt.Equal(fixedTime) {
		t.Errorf("Expected ImposedAt %v, got %v", fixedTime, embargo.ImposedAt)
	}
}

// TestReputationPenaltyDeterministicTimestamps verifies penalties use TimeProvider.
func TestReputationPenaltyDeterministicTimestamps(t *testing.T) {
	fixedTime := time.Date(2026, 6, 20, 16, 45, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)

	// DeclareWar triggers an internal reputation penalty
	_, err := manager.DeclareWar(guildID1, guildID2, time.Hour)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	penalties := manager.GetReputationPenalties()
	if len(penalties) == 0 {
		t.Fatal("Expected at least one penalty")
	}
	if !penalties[0].AppliedAt.Equal(fixedTime) {
		t.Errorf("Expected penalty AppliedAt %v, got %v", fixedTime, penalties[0].AppliedAt)
	}
}

// TestUpdateDeterministicTimeComparison verifies Update uses TimeProvider for comparisons.
func TestUpdateDeterministicTimeComparison(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)

	// Set time to T0 for war declaration
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: t0})
	defer ResetTimeProvider()

	prepPeriod := time.Hour
	war, err := manager.DeclareWar(guildID1, guildID2, prepPeriod)
	if err != nil {
		t.Fatalf("DeclareWar failed: %v", err)
	}

	// Update at T0 + 30min: war should NOT be active
	SetTimeProvider(FixedTimeProvider{FixedTime: t0.Add(30 * time.Minute)})
	manager.Update(0)
	if war.Active {
		t.Error("War should not be active before preparation ends")
	}

	// Update at T0 + 2h: war should be active
	SetTimeProvider(FixedTimeProvider{FixedTime: t0.Add(2 * time.Hour)})
	manager.Update(0)
	if !war.Active {
		t.Error("War should be active after preparation period")
	}
}

// TestGetTradeDiscountDeterministicTime verifies discount expiration uses TimeProvider.
func TestGetTradeDiscountDeterministicTime(t *testing.T) {
	world := engine.NewWorld()
	guildManager := guild.NewManager()
	guildID1, _ := guildManager.CreateGuild("fantasy", "Player1", 12345)
	guildID2, _ := guildManager.CreateGuild("fantasy", "Player2", 23456)

	manager := NewManager(world, guildManager)

	// Manually insert a trade concession with known expiration
	applyTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	manager.appliedConcessions = append(manager.appliedConcessions, AppliedConcession{
		Type:              ConcessionTrade,
		AttackerGuildID:   guildID1,
		DefenderGuildID:   guildID2,
		AppliedAt:         applyTime,
		TradeDiscountPct:  0.15,
		TradeDiscountEnds: applyTime.Add(TradeDiscountDuration),
	})

	// Before expiration: discount should be active
	SetTimeProvider(FixedTimeProvider{FixedTime: applyTime.Add(10 * 24 * time.Hour)})
	defer ResetTimeProvider()

	discount := manager.GetTradeDiscount(guildID1, guildID2)
	if discount != 0.15 {
		t.Errorf("Expected discount 0.15, got %f", discount)
	}

	// After expiration: discount should be zero
	SetTimeProvider(FixedTimeProvider{FixedTime: applyTime.Add(31 * 24 * time.Hour)})
	discount = manager.GetTradeDiscount(guildID1, guildID2)
	if discount != 0 {
		t.Errorf("Expected expired discount 0, got %f", discount)
	}
}
