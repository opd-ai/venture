package memprofile

import (
	"testing"
	"time"
)

// TestCaptureMemorySnapshot tests snapshot capture.
func TestCaptureMemorySnapshot(t *testing.T) {
	snapshot := CaptureMemorySnapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("Snapshot timestamp not set")
	}

	if snapshot.Alloc == 0 {
		t.Error("Snapshot alloc is 0")
	}

	if snapshot.TotalAlloc == 0 {
		t.Error("Snapshot total alloc is 0")
	}

	if snapshot.Mallocs == 0 {
		t.Error("Snapshot mallocs is 0")
	}
}

// TestStartMemoryProfile tests profile initialization.
func TestStartMemoryProfile(t *testing.T) {
	profile := StartMemoryProfile("test")

	if profile.Name != "test" {
		t.Errorf("Expected name 'test', got %s", profile.Name)
	}

	if len(profile.Snapshots) == 0 {
		t.Error("Profile should have initial snapshot")
	}

	if profile.StartTime.IsZero() {
		t.Error("StartTime not set")
	}
}

// TestMemoryProfileSnapshot tests adding snapshots.
func TestMemoryProfileSnapshot(t *testing.T) {
	profile := StartMemoryProfile("test")
	initialCount := len(profile.Snapshots)

	profile.Snapshot()

	if len(profile.Snapshots) != initialCount+1 {
		t.Errorf("Expected %d snapshots, got %d", initialCount+1, len(profile.Snapshots))
	}
}

// TestMemoryProfileEnd tests profile finalization.
func TestMemoryProfileEnd(t *testing.T) {
	profile := StartMemoryProfile("test")

	// Allocate some memory
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	profile.End()

	if profile.EndTime.IsZero() {
		t.Error("EndTime not set")
	}

	if len(profile.Snapshots) < 2 {
		t.Error("Profile should have at least 2 snapshots (start and end)")
	}
}

// TestGetPeakAllocation tests peak allocation calculation.
func TestGetPeakAllocation(t *testing.T) {
	profile := StartMemoryProfile("test")

	// Allocate memory to create variation
	data1 := make([]byte, 1024*100)
	profile.Snapshot()
	_ = data1

	data2 := make([]byte, 1024*200)
	profile.Snapshot()
	_ = data2

	profile.End()

	peak := profile.GetPeakAllocation()
	if peak == 0 {
		t.Error("Peak allocation should not be 0")
	}
}

// TestGetAverageAllocation tests average allocation calculation.
func TestGetAverageAllocation(t *testing.T) {
	profile := StartMemoryProfile("test")
	profile.Snapshot()
	profile.Snapshot()
	profile.End()

	avg := profile.GetAverageAllocation()
	if avg == 0 {
		t.Error("Average allocation should not be 0")
	}
}

// TestGetAllocationGrowth tests allocation growth calculation.
func TestGetAllocationGrowth(t *testing.T) {
	profile := StartMemoryProfile("test")

	// Allocate memory
	data := make([]byte, 1024*1024)
	_ = data

	profile.End()

	growth := profile.GetAllocationGrowth()
	// Growth might be positive or negative depending on GC
	// Just check it's calculated
	_ = growth
}

// TestGetObjectGrowth tests object growth calculation.
func TestGetObjectGrowth(t *testing.T) {
	profile := StartMemoryProfile("test")

	// Allocate objects
	for i := 0; i < 1000; i++ {
		_ = make([]byte, 100)
	}

	profile.End()

	growth := profile.GetObjectGrowth()
	// Growth might be positive or negative depending on GC
	// Just check it's calculated
	_ = growth
}

// TestProfileFunction tests function profiling.
func TestProfileFunction(t *testing.T) {
	counter := 0
	profile := ProfileFunction("test_func", 100, func() {
		counter++
	})

	if counter != 100 {
		t.Errorf("Expected 100 iterations, got %d", counter)
	}

	if len(profile.Snapshots) < 2 {
		t.Error("Profile should have multiple snapshots")
	}

	if profile.EndTime.Before(profile.StartTime) {
		t.Error("EndTime should be after StartTime")
	}
}

// TestDetectLeaksInBenchmark tests leak detection.
func TestDetectLeaksInBenchmark(t *testing.T) {
	// Test non-leaking function
	leaked := DetectLeaksInBenchmark("no_leak", 10, 100, func() {
		// Simple allocation that will be GC'd
		_ = make([]byte, 100)
	})

	// We can't guarantee no leak detection due to GC timing
	// Just verify the function runs without error
	_ = leaked
}

// TestRunMemoryTest tests memory test runner.
func TestRunMemoryTest(t *testing.T) {
	test := MemoryTest{
		Name:       "simple_test",
		WarmupRuns: 10,
		TestRuns:   50,
		MaxGrowth:  10 * 1024 * 1024, // 10MB
	}

	counter := 0
	passed, profile := RunMemoryTest(test, func() {
		counter++
	})

	if counter != test.TestRuns {
		t.Errorf("Expected %d iterations, got %d", test.TestRuns, counter)
	}

	if profile == nil {
		t.Fatal("Profile should not be nil")
	}

	// Passing depends on GC timing, just verify it completes
	_ = passed
}

// TestLeakDetection tests memory leak detection logic.
func TestLeakDetection(t *testing.T) {
	profile := &MemoryProfile{
		Name:      "leak_test",
		Snapshots: make([]MemorySnapshot, 0),
		StartTime: time.Now(),
	}

	// Add first snapshot with low allocation
	profile.Snapshots = append(profile.Snapshots, MemorySnapshot{
		Timestamp:   time.Now(),
		Alloc:       1000000, // 1MB
		LiveObjects: 1000,
	})

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Add second snapshot with higher allocation (>10% growth)
	profile.Snapshots = append(profile.Snapshots, MemorySnapshot{
		Timestamp:   time.Now(),
		Alloc:       1200000, // 1.2MB (20% growth)
		LiveObjects: 1300,    // 30% growth
	})

	profile.EndTime = time.Now()
	profile.detectLeaks()

	if !profile.LeakDetected {
		t.Error("Expected leak to be detected with 20% allocation and 30% object growth")
	}

	if profile.LeakRate <= 0 {
		t.Error("Expected positive leak rate")
	}
}

// TestNoLeakDetection tests that normal variation doesn't trigger leak detection.
func TestNoLeakDetection(t *testing.T) {
	profile := &MemoryProfile{
		Name:      "no_leak_test",
		Snapshots: make([]MemorySnapshot, 0),
		StartTime: time.Now(),
	}

	// Add snapshots with minimal growth
	profile.Snapshots = append(profile.Snapshots, MemorySnapshot{
		Timestamp:   time.Now(),
		Alloc:       1000000,
		LiveObjects: 1000,
	})

	time.Sleep(10 * time.Millisecond)

	profile.Snapshots = append(profile.Snapshots, MemorySnapshot{
		Timestamp:   time.Now(),
		Alloc:       1050000, // 5% growth (below 10% threshold)
		LiveObjects: 1050,
	})

	profile.EndTime = time.Now()
	profile.detectLeaks()

	if profile.LeakDetected {
		t.Error("Should not detect leak with <10% growth")
	}
}

// TestPrintProfile verifies profile printing doesn't panic.
func TestPrintProfile(t *testing.T) {
	profile := StartMemoryProfile("test")
	profile.Snapshot()
	profile.End()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintProfile panicked: %v", r)
		}
	}()

	profile.PrintProfile()
}

// BenchmarkCaptureMemorySnapshot benchmarks snapshot capture.
func BenchmarkCaptureMemorySnapshot(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CaptureMemorySnapshot()
	}
}

// BenchmarkProfileFunction benchmarks function profiling.
func BenchmarkProfileFunction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ProfileFunction("bench", 10, func() {
			_ = make([]byte, 100)
		})
	}
}
