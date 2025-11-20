package engine

import (
	"testing"
	"time"
)

func TestSimulationClock(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		advances   []float64
		wantOffset time.Duration
	}{
		{
			name:       "new clock starts at epoch",
			seed:       12345,
			advances:   []float64{},
			wantOffset: 0,
		},
		{
			name:       "advance by 1 second",
			seed:       12345,
			advances:   []float64{1.0},
			wantOffset: 1 * time.Second,
		},
		{
			name:       "advance by multiple steps",
			seed:       12345,
			advances:   []float64{1.0, 2.5, 0.5},
			wantOffset: 4 * time.Second,
		},
		{
			name:       "fractional advance",
			seed:       12345,
			advances:   []float64{0.016667}, // ~60 FPS frame time
			wantOffset: 16667 * time.Microsecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := NewSimulationClock(tt.seed)

			// Verify starts at epoch
			start := clock.Now()
			if !start.Equal(time.Unix(0, 0)) {
				t.Errorf("SimulationClock.Now() start = %v, want Unix epoch", start)
			}

			// Apply advances
			for _, delta := range tt.advances {
				clock.Advance(delta)
			}

			// Check final time
			got := clock.Now()
			want := time.Unix(0, 0).Add(tt.wantOffset)
			if !got.Equal(want) {
				t.Errorf("SimulationClock.Now() = %v, want %v", got, want)
			}
		})
	}
}

func TestSimulationClockReset(t *testing.T) {
	clock := NewSimulationClock(12345)

	// Advance time
	clock.Advance(10.0)

	// Reset to a specific time
	resetTime := time.Date(2025, 11, 20, 12, 0, 0, 0, time.UTC)
	clock.Reset(resetTime)

	got := clock.Now()
	if !got.Equal(resetTime) {
		t.Errorf("SimulationClock.Now() after reset = %v, want %v", got, resetTime)
	}

	// Verify advance works after reset
	clock.Advance(5.0)
	want := resetTime.Add(5 * time.Second)
	got = clock.Now()
	if !got.Equal(want) {
		t.Errorf("SimulationClock.Now() after reset+advance = %v, want %v", got, want)
	}
}

func TestSimulationClockConcurrency(t *testing.T) {
	clock := NewSimulationClock(12345)

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = clock.Now()
		}
		done <- true
	}()

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			clock.Advance(0.001)
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	// Verify final state is valid
	finalTime := clock.Now()
	if finalTime.Before(time.Unix(0, 0)) {
		t.Errorf("SimulationClock.Now() after concurrent operations = %v, should not be before epoch", finalTime)
	}
}

func TestRealTimeClock(t *testing.T) {
	clock := NewRealTimeClock()

	// Get time before and after
	before := time.Now()
	got := clock.Now()
	after := time.Now()

	// Verify clock returns actual wall time
	if got.Before(before) || got.After(after) {
		t.Errorf("RealTimeClock.Now() = %v, expected between %v and %v", got, before, after)
	}

	// Verify Advance is a no-op
	timeBefore := clock.Now()
	time.Sleep(1 * time.Millisecond)
	clock.Advance(100.0) // Should have no effect
	timeAfter := clock.Now()

	// Time should have advanced by wall-clock time, not by Advance parameter
	if timeAfter.Sub(timeBefore) > 100*time.Millisecond {
		t.Errorf("RealTimeClock.Advance() affected time: before=%v, after=%v", timeBefore, timeAfter)
	}

	// Verify Reset is a no-op
	resetTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	clock.Reset(resetTime)
	gotAfterReset := clock.Now()

	// Should still be current time, not reset time
	if gotAfterReset.Year() < 2020 {
		t.Errorf("RealTimeClock.Reset() affected time: got year %d, expected current year", gotAfterReset.Year())
	}
}

func TestGameClockInterface(t *testing.T) {
	// Verify both clock types implement GameClock interface
	var _ GameClock = (*SimulationClock)(nil)
	var _ GameClock = (*RealTimeClock)(nil)

	// Test polymorphism
	clocks := []GameClock{
		NewSimulationClock(12345),
		NewRealTimeClock(),
	}

	for i, clock := range clocks {
		// All clocks should support Now()
		now := clock.Now()
		if now.IsZero() {
			t.Errorf("clock[%d].Now() returned zero time", i)
		}

		// All clocks should accept Advance() and Reset() without panicking
		clock.Advance(1.0)
		clock.Reset(time.Now())
	}
}

func BenchmarkSimulationClockNow(b *testing.B) {
	clock := NewSimulationClock(12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = clock.Now()
	}
}

func BenchmarkSimulationClockAdvance(b *testing.B) {
	clock := NewSimulationClock(12345)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clock.Advance(0.016667)
	}
}

func BenchmarkRealTimeClockNow(b *testing.B) {
	clock := NewRealTimeClock()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = clock.Now()
	}
}
