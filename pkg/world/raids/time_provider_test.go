package raids

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestRealTimeProvider(t *testing.T) {
	provider := RealTimeProvider{}
	before := time.Now()
	result := provider.Now()
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Errorf("RealTimeProvider.Now() = %v, want time between %v and %v", result, before, after)
	}
}

func TestFixedTimeProvider(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	provider := FixedTimeProvider{FixedTime: fixedTime}

	result := provider.Now()
	if !result.Equal(fixedTime) {
		t.Errorf("FixedTimeProvider.Now() = %v, want %v", result, fixedTime)
	}

	// Should always return the same time
	result2 := provider.Now()
	if !result2.Equal(fixedTime) {
		t.Errorf("FixedTimeProvider.Now() = %v, want %v", result2, fixedTime)
	}
}

func TestDefaultTimeProvider(t *testing.T) {
	provider := DefaultTimeProvider()
	if provider == nil {
		t.Fatal("DefaultTimeProvider() returned nil")
	}

	// Should return RealTimeProvider
	_, ok := provider.(RealTimeProvider)
	if !ok {
		t.Errorf("DefaultTimeProvider() type = %T, want RealTimeProvider", provider)
	}
}

func TestLockoutManager_WithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	provider := FixedTimeProvider{FixedTime: fixedTime}
	lm := NewLockoutManagerWithProvider(24*time.Hour, provider)

	// Record a clear
	lm.RecordClear("player1", TierNormal)

	// Check lockout status at fixed time
	if !lm.IsLockedOut("player1", TierNormal) {
		t.Error("Player should be locked out immediately after clear")
	}

	// Advance time past lockout period
	provider.FixedTime = fixedTime.Add(25 * time.Hour)
	lm.timeProvider = provider

	if lm.IsLockedOut("player1", TierNormal) {
		t.Error("Player should not be locked out after period expires")
	}
}

func TestInstanceManager_WithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
	provider := FixedTimeProvider{FixedTime: fixedTime}
	im := NewInstanceManagerWithProvider(1*time.Hour, provider)

	gen := NewGenerator(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":     TierNormal,
			"group_id": "test-group",
		},
	}
	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}
	raid := result.(*RaidDungeon)

	// Create instance at fixed time
	instance, err := im.CreateInstance(raid, "group1", []string{"p1", "p2", "p3", "p4", "p5"})
	if err != nil {
		t.Fatalf("CreateInstance() failed: %v", err)
	}

	// Check instance exists at fixed time
	retrieved, exists := im.GetInstance(instance.InstanceID)
	if !exists {
		t.Fatal("Instance should exist at creation time")
	}
	if retrieved.InstanceID != instance.InstanceID {
		t.Errorf("Retrieved instance ID = %s, want %s", retrieved.InstanceID, instance.InstanceID)
	}

	// Advance time past expiration
	provider.FixedTime = fixedTime.Add(2 * time.Hour)
	im.timeProvider = provider

	// Instance should now be expired
	_, exists = im.GetInstance(instance.InstanceID)
	if exists {
		t.Error("Instance should be expired after timeout period")
	}
}

func TestLockoutManager_Determinism(t *testing.T) {
	// Test that same time produces same lockout expiration
	fixedTime := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	provider1 := FixedTimeProvider{FixedTime: fixedTime}
	provider2 := FixedTimeProvider{FixedTime: fixedTime}

	lm1 := NewLockoutManagerWithProvider(7*24*time.Hour, provider1)
	lm2 := NewLockoutManagerWithProvider(7*24*time.Hour, provider2)

	lm1.RecordClear("player1", TierHeroic)
	lm2.RecordClear("player1", TierHeroic)

	lockout1, _ := lm1.GetLockout("player1", TierHeroic)
	lockout2, _ := lm2.GetLockout("player1", TierHeroic)

	if !lockout1.LastRun.Equal(lockout2.LastRun) {
		t.Errorf("LastRun times differ: %v vs %v", lockout1.LastRun, lockout2.LastRun)
	}

	if !lockout1.NextReset.Equal(lockout2.NextReset) {
		t.Errorf("NextReset times differ: %v vs %v", lockout1.NextReset, lockout2.NextReset)
	}
}

func TestInstanceManager_Determinism(t *testing.T) {
	// Test that same time produces same instance creation time
	fixedTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
	provider1 := FixedTimeProvider{FixedTime: fixedTime}
	provider2 := FixedTimeProvider{FixedTime: fixedTime}

	im1 := NewInstanceManagerWithProvider(4*time.Hour, provider1)
	im2 := NewInstanceManagerWithProvider(4*time.Hour, provider2)

	gen := NewGenerator(54321)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"tier":     TierMythic,
			"group_id": "test-group",
		},
	}
	result, _ := gen.Generate(54321, params)
	raid := result.(*RaidDungeon)

	instance1, _ := im1.CreateInstance(raid, "groupA", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"})
	instance2, _ := im2.CreateInstance(raid, "groupA", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"})

	if !instance1.CreatedAt.Equal(instance2.CreatedAt) {
		t.Errorf("CreatedAt times differ: %v vs %v", instance1.CreatedAt, instance2.CreatedAt)
	}

	if !instance1.ExpiresAt.Equal(instance2.ExpiresAt) {
		t.Errorf("ExpiresAt times differ: %v vs %v", instance1.ExpiresAt, instance2.ExpiresAt)
	}
}

func BenchmarkRealTimeProvider(b *testing.B) {
	provider := RealTimeProvider{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Now()
	}
}

func BenchmarkFixedTimeProvider(b *testing.B) {
	provider := FixedTimeProvider{FixedTime: time.Now()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Now()
	}
}
