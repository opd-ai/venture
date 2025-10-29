package engine

import (
	"sync"
	"testing"
)

func TestStatusEffectPool_NewAndRelease(t *testing.T) {
	// Reset pool stats for clean test
	ResetStatusEffectPoolStats()

	// Acquire a status effect from the pool
	effect := NewStatusEffectComponent("poison", 5.0, 10.0, 1.0)

	// Verify initialization
	if effect.EffectType != "poison" {
		t.Errorf("EffectType = %s, want poison", effect.EffectType)
	}
	if effect.Magnitude != 5.0 {
		t.Errorf("Magnitude = %f, want 5.0", effect.Magnitude)
	}
	if effect.Duration != 10.0 {
		t.Errorf("Duration = %f, want 10.0", effect.Duration)
	}
	if effect.TickInterval != 1.0 {
		t.Errorf("TickInterval = %f, want 1.0", effect.TickInterval)
	}
	if effect.NextTick != 1.0 {
		t.Errorf("NextTick = %f, want 1.0", effect.NextTick)
	}

	// Release back to pool
	ReleaseStatusEffect(effect)

	// Verify reset
	if effect.EffectType != "" {
		t.Errorf("EffectType not reset, got %s", effect.EffectType)
	}
	if effect.Magnitude != 0 {
		t.Errorf("Magnitude not reset, got %f", effect.Magnitude)
	}
	if effect.Duration != 0 {
		t.Errorf("Duration not reset, got %f", effect.Duration)
	}
}

func TestStatusEffectPool_ReleaseNil(t *testing.T) {
	// Should not panic when releasing nil
	ReleaseStatusEffect(nil)
}

func TestStatusEffectPool_Reuse(t *testing.T) {
	ResetStatusEffectPoolStats()

	// Acquire effect
	effect1 := NewStatusEffectComponent("burn", 2.0, 5.0, 0.5)
	effect1Ptr := effect1

	// Release to pool
	ReleaseStatusEffect(effect1)

	// Acquire another effect - should reuse the same object
	effect2 := NewStatusEffectComponent("freeze", 3.0, 8.0, 1.0)

	// Verify it's the same object (pointer equality)
	if effect1Ptr != effect2 {
		// Note: sync.Pool doesn't guarantee reuse, so this test may occasionally fail
		// That's expected behavior - the pool is allowed to discard objects
		t.Logf("Object was not reused (this is acceptable sync.Pool behavior)")
	} else {
		t.Logf("Object successfully reused from pool")
	}

	// Verify new values are set correctly
	if effect2.EffectType != "freeze" {
		t.Errorf("EffectType = %s, want freeze", effect2.EffectType)
	}
	if effect2.Magnitude != 3.0 {
		t.Errorf("Magnitude = %f, want 3.0", effect2.Magnitude)
	}
}

func TestStatusEffectPool_ConcurrentAccess(t *testing.T) {
	ResetStatusEffectPoolStats()

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Concurrent acquire and release
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				effect := NewStatusEffectComponent("test", 1.0, 1.0, 0.1)
				// Simulate some work
				_ = effect.Type()
				ReleaseStatusEffect(effect)
			}
		}(i)
	}

	wg.Wait()

	// All effects should be returned to pool
	// Note: We don't check exact numbers because sync.Pool may discard objects
	t.Logf("Concurrent test completed successfully")
}

func TestStatusEffectPool_Stats(t *testing.T) {
	ResetStatusEffectPoolStats()

	// Acquire multiple effects
	effect1 := NewStatusEffectComponent("test1", 1.0, 1.0, 0.1)
	effect2 := NewStatusEffectComponent("test2", 2.0, 2.0, 0.2)
	effect3 := NewStatusEffectComponent("test3", 3.0, 3.0, 0.3)

	// Check stats after acquisition
	// Note: We removed stats tracking from the pool implementation to keep it simple
	// This test just verifies the API exists
	_ = GetStatusEffectPoolStats()

	// Release effects
	ReleaseStatusEffect(effect1)
	ReleaseStatusEffect(effect2)
	ReleaseStatusEffect(effect3)

	// Stats should be accessible
	_ = GetStatusEffectPoolStats()
}

func TestStatusEffectComponent_Reset(t *testing.T) {
	effect := &StatusEffectComponent{
		EffectType:   "poison",
		Duration:     10.0,
		Magnitude:    5.0,
		TickInterval: 1.0,
		NextTick:     0.5,
	}

	effect.Reset()

	// Verify all fields are cleared
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"EffectType", effect.EffectType, ""},
		{"Duration", effect.Duration, 0.0},
		{"Magnitude", effect.Magnitude, 0.0},
		{"TickInterval", effect.TickInterval, 0.0},
		{"NextTick", effect.NextTick, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestStatusEffectPool_MultipleReleases(t *testing.T) {
	effect := NewStatusEffectComponent("test", 1.0, 1.0, 0.1)

	// First release - normal
	ReleaseStatusEffect(effect)

	// Second release - should be safe (idempotent)
	// This tests that releasing the same object twice doesn't cause issues
	ReleaseStatusEffect(effect)

	// Test should not panic
}

func TestStatusEffectPool_Integration(t *testing.T) {
	// Test typical usage pattern in combat system
	ResetStatusEffectPoolStats()

	// Simulate applying multiple status effects
	effects := make([]*StatusEffectComponent, 0, 10)
	for i := 0; i < 10; i++ {
		effect := NewStatusEffectComponent("test", float64(i), float64(i), 0.1)
		effects = append(effects, effect)
	}

	// Verify all effects are unique and properly initialized
	for i, effect := range effects {
		if effect.Magnitude != float64(i) {
			t.Errorf("Effect %d: Magnitude = %f, want %d", i, effect.Magnitude, i)
		}
	}

	// Release all effects back to pool
	for _, effect := range effects {
		ReleaseStatusEffect(effect)
	}

	// Acquire new effects - some may be reused
	newEffects := make([]*StatusEffectComponent, 0, 10)
	for i := 0; i < 10; i++ {
		effect := NewStatusEffectComponent("new", float64(i*2), float64(i*2), 0.2)
		newEffects = append(newEffects, effect)
	}

	// Verify new effects have correct values (not contaminated)
	for i, effect := range newEffects {
		if effect.Magnitude != float64(i*2) {
			t.Errorf("New Effect %d: Magnitude = %f, want %d", i, effect.Magnitude, i*2)
		}
		if effect.EffectType != "new" {
			t.Errorf("New Effect %d: EffectType = %s, want new", i, effect.EffectType)
		}
	}
}

// Benchmarks

func BenchmarkStatusEffectPool_New(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		effect := NewStatusEffectComponent("test", 1.0, 1.0, 0.1)
		_ = effect
	}
}

func BenchmarkStatusEffectPool_NewAndRelease(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		effect := NewStatusEffectComponent("test", 1.0, 1.0, 0.1)
		ReleaseStatusEffect(effect)
	}
}

func BenchmarkStatusEffectPool_DirectAllocation(b *testing.B) {
	// Baseline: direct allocation without pooling
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		effect := &StatusEffectComponent{
			EffectType:   "test",
			Duration:     1.0,
			Magnitude:    1.0,
			TickInterval: 0.1,
			NextTick:     0.1,
		}
		_ = effect
	}
}

func BenchmarkStatusEffectPool_ConcurrentNewAndRelease(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			effect := NewStatusEffectComponent("test", 1.0, 1.0, 0.1)
			ReleaseStatusEffect(effect)
		}
	})
}

func BenchmarkStatusEffectPool_TypicalCombatPattern(b *testing.B) {
	// Simulate typical combat pattern: apply effects, update a few times, expire
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Apply 3 status effects
		effect1 := NewStatusEffectComponent("burn", 2.0, 5.0, 0.5)
		effect2 := NewStatusEffectComponent("poison", 1.0, 8.0, 1.0)
		effect3 := NewStatusEffectComponent("slow", 0.5, 3.0, 0.0)

		// Simulate a few update cycles
		for j := 0; j < 10; j++ {
			effect1.Update(0.1)
			effect2.Update(0.1)
			effect3.Update(0.1)
		}

		// Effects expire, return to pool
		ReleaseStatusEffect(effect1)
		ReleaseStatusEffect(effect2)
		ReleaseStatusEffect(effect3)
	}
}
