package engine

import (
	"testing"
	"time"
)

func TestFrameTimeTracker_New(t *testing.T) {
	tracker := NewFrameTimeTracker(1000)
	if tracker == nil {
		t.Fatal("NewFrameTimeTracker returned nil")
	}
	if tracker.maxSamples != 1000 {
		t.Errorf("maxSamples = %d, want 1000", tracker.maxSamples)
	}
}

func TestFrameTimeTracker_RecordAndGet(t *testing.T) {
	tracker := NewFrameTimeTracker(100)

	// Record 100 consistent frames at 16ms (60 FPS)
	for i := 0; i < 100; i++ {
		tracker.RecordFrame(16 * time.Millisecond)
	}

	stats := tracker.GetStats()
	if stats.Average != 16*time.Millisecond {
		t.Errorf("Average = %v, want 16ms", stats.Average)
	}
	if stats.IsStuttering() {
		t.Error("Should not detect stuttering for consistent 16ms frames")
	}
}

func TestFrameTimeTracker_StutteringDetection(t *testing.T) {
	tracker := NewFrameTimeTracker(1000)

	// 95% good frames, 5% bad frames
	for i := 0; i < 950; i++ {
		tracker.RecordFrame(16 * time.Millisecond)
	}
	for i := 0; i < 50; i++ {
		tracker.RecordFrame(30 * time.Millisecond) // Spikes
	}

	stats := tracker.GetStats()
	if !stats.IsStuttering() {
		t.Error("Should detect stuttering with 30ms spikes")
	}
}

func TestFrameTimeStats_GetFPS(t *testing.T) {
	stats := FrameTimeStats{Average: 16 * time.Millisecond}
	fps := stats.GetFPS()
	if fps < 60 || fps > 65 {
		t.Errorf("GetFPS() = %.2f, want ~62.5", fps)
	}
}

func BenchmarkFrameTimeTracker_Record(b *testing.B) {
	tracker := NewFrameTimeTracker(1000)
	frameTime := 16 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.RecordFrame(frameTime)
	}
}

func BenchmarkFrameTimeTracker_GetStats(b *testing.B) {
	tracker := NewFrameTimeTracker(1000)
	for i := 0; i < 1000; i++ {
		tracker.RecordFrame(16 * time.Millisecond)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tracker.GetStats()
	}
}
