package performance

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestPerformanceMonitor tests the performance monitor
func TestPerformanceMonitor(t *testing.T) {
	pm := NewPerformanceMonitor()

	// Test frame time updates
	pm.UpdateFrameTime(10.0)
	if pm.GetFrameTime() != 10.0 {
		t.Errorf("Expected frame time 10.0, got %.2f", pm.GetFrameTime())
	}

	// Test FPS calculation
	fps := pm.GetFPS()
	expected := 100.0 // 1000ms / 10ms = 100 FPS
	if fps != expected {
		t.Errorf("Expected FPS %.2f, got %.2f", expected, fps)
	}

	// Test performance target (60 FPS = 16.67ms)
	pm.UpdateFrameTime(16.67)
	if !pm.CheckPerformanceTarget() {
		t.Error("Expected to meet performance target at 16.67ms")
	}

	pm.UpdateFrameTime(20.0)
	if pm.CheckPerformanceTarget() {
		t.Error("Expected to fail performance target at 20ms")
	}
}

// TestMemoryProfiler tests memory profiling
func TestMemoryProfiler(t *testing.T) {
	mp := NewMemoryProfiler()

	// Track allocations
	mp.TrackAllocation("raid_dungeons", 50*1024*1024) // 50MB
	mp.TrackAllocation("sprites", 300*1024*1024)      // 300MB
	mp.TrackAllocation("terrain", 100*1024*1024)      // 100MB

	stats := mp.GetStats()

	if stats.TotalMB != 450 {
		t.Errorf("Expected 450MB total, got %d", stats.TotalMB)
	}

	if stats.LargestAlloc != "sprites" {
		t.Errorf("Expected largest alloc 'sprites', got %s", stats.LargestAlloc)
	}

	// Test release
	mp.ReleaseAllocation("terrain", 50*1024*1024)
	stats = mp.GetStats()
	if stats.TotalMB != 400 {
		t.Errorf("Expected 400MB after release, got %d", stats.TotalMB)
	}
}

// TestMemorySnapshot tests snapshot functionality
func TestMemorySnapshot(t *testing.T) {
	mp := NewMemoryProfiler()

	// Create snapshots over time
	for i := 0; i < 5; i++ {
		mp.TrackAllocation("test", uint64(i*10*1024*1024))
		mp.TakeSnapshot()
		time.Sleep(10 * time.Millisecond)
	}

	snapshots := mp.GetSnapshots()
	if len(snapshots) != 5 {
		t.Errorf("Expected 5 snapshots, got %d", len(snapshots))
	}

	// Verify timestamps are increasing
	for i := 1; i < len(snapshots); i++ {
		if !snapshots[i].Timestamp.After(snapshots[i-1].Timestamp) {
			t.Error("Snapshot timestamps not increasing")
		}
	}
}

// TestMemoryLeakDetection tests leak identification
func TestMemoryLeakDetection(t *testing.T) {
	mp := NewMemoryProfiler()

	// Simulate a leak: allocation that only grows
	for i := 0; i < 12; i++ {
		mp.TrackAllocation("leak", uint64(i*10*1024*1024))
		mp.TakeSnapshot()
	}

	// Simulate normal allocation: grows and shrinks
	for i := 0; i < 12; i++ {
		if i%2 == 0 {
			mp.TrackAllocation("normal", 10*1024*1024)
		} else {
			mp.ReleaseAllocation("normal", 10*1024*1024)
		}
		mp.TakeSnapshot()
	}

	leaks := mp.IdentifyLeaks(10) // 10MB minimum growth

	if len(leaks) == 0 {
		t.Error("Expected to identify 'leak' allocation")
	}

	found := false
	for _, leak := range leaks {
		if leak == "leak" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Leak detection did not identify 'leak' allocation")
	}
}

// TestNetworkBatcher tests network batching
func TestNetworkBatcher(t *testing.T) {
	var batchReceived int32
	var receivedBatch atomic.Value

	sendFunc := func(batch *BatchedMessage) {
		atomic.StoreInt32(&batchReceived, 1)
		receivedBatch.Store(batch)
	}

	nb := NewNetworkBatcher(100, sendFunc)
	nb.Start()
	defer nb.Stop()

	// Queue messages
	nb.QueueMessage("pos_update", []byte("player1_data"), "player1")
	nb.QueueMessage("pos_update", []byte("player2_data"), "player2")

	// Wait for batch window
	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&batchReceived) != 1 {
		t.Error("Expected batch to be sent")
	}

	batch := receivedBatch.Load()
	if batch == nil {
		t.Fatal("Received batch is nil")
	}

	batchMsg := batch.(*BatchedMessage)
	if len(batchMsg.Messages) != 2 {
		t.Errorf("Expected 2 messages in batch, got %d", len(batchMsg.Messages))
	}
}

// TestNetworkBatcherMaxSize tests max batch size enforcement
func TestNetworkBatcherMaxSize(t *testing.T) {
	var batchCount int32

	sendFunc := func(batch *BatchedMessage) {
		atomic.AddInt32(&batchCount, 1)
	}

	nb := NewNetworkBatcher(1000, sendFunc) // Long window
	nb.SetMaxBatchSize(3)
	nb.Start()
	defer nb.Stop()

	// Queue 5 messages (should trigger 2 batches: 3 + 2)
	for i := 0; i < 5; i++ {
		nb.QueueMessage("test", []byte("data"), "player1")
	}

	// Wait briefly
	time.Sleep(50 * time.Millisecond)

	// Should have 1 batch from max size trigger, 1 still queued
	if atomic.LoadInt32(&batchCount) != 1 {
		t.Errorf("Expected 1 batch from max size trigger, got %d", atomic.LoadInt32(&batchCount))
	}
}

// TestCacheManager tests cache management
func TestCacheManager(t *testing.T) {
	cm := NewCacheManager(100 * 1024 * 1024) // 100MB limit

	// Add entries
	cm.Set("sprite1", "data1", 10*1024*1024)
	cm.Set("sprite2", "data2", 20*1024*1024)
	cm.Set("sprite3", "data3", 30*1024*1024)

	// Retrieve entry
	data, found := cm.Get("sprite2")
	if !found {
		t.Error("Expected to find sprite2")
	}
	if data != "data2" {
		t.Errorf("Expected data2, got %v", data)
	}

	// Check stats
	stats := cm.GetStats()
	if stats.ItemCount != 3 {
		t.Errorf("Expected 3 items, got %d", stats.ItemCount)
	}
}

// TestCacheLRUEviction tests LRU eviction
func TestCacheLRUEviction(t *testing.T) {
	cm := NewCacheManager(50 * 1024 * 1024) // 50MB limit

	// Add entries that exceed limit
	cm.Set("entry1", "data1", 20*1024*1024)
	cm.Set("entry2", "data2", 20*1024*1024)
	cm.Set("entry3", "data3", 20*1024*1024) // Should trigger eviction

	// entry1 should be evicted (LRU)
	_, found := cm.Get("entry1")
	if found {
		t.Error("Expected entry1 to be evicted")
	}

	// entry2 and entry3 should remain
	_, found = cm.Get("entry2")
	if !found {
		t.Error("Expected entry2 to remain")
	}

	_, found = cm.Get("entry3")
	if !found {
		t.Error("Expected entry3 to remain")
	}
}

// TestBackgroundLoader tests background loading
func TestBackgroundLoader(t *testing.T) {
	bl := NewBackgroundLoader(2)
	bl.Start()
	defer bl.Stop()

	var loaded int32
	bl.PreloadRaid("raid1", func(data interface{}) {
		atomic.StoreInt32(&loaded, 1)
	})

	// Wait for loading
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&loaded) != 1 {
		t.Error("Expected raid to be loaded")
	}
}

// TestLODManager tests LOD level selection
func TestLODManager(t *testing.T) {
	lm := NewLODManager()

	// Test LOD levels at different distances
	tests := []struct {
		distance float64
		expected LODLevel
	}{
		{50.0, LODHigh},
		{150.0, LODMedium},
		{400.0, LODLow},
		{700.0, LODVeryLow},
	}

	for _, tt := range tests {
		level := lm.GetLODLevel(tt.distance)
		if level != tt.expected {
			t.Errorf("Distance %.1f: expected %s, got %s", tt.distance, tt.expected, level)
		}
	}
}

// TestLODManagerDisabled tests LOD when disabled
func TestLODManagerDisabled(t *testing.T) {
	lm := NewLODManager()
	lm.Disable()

	// All distances should return VeryHigh when disabled
	level := lm.GetLODLevel(1000.0)
	if level != LODVeryHigh {
		t.Errorf("Expected VeryHigh when disabled, got %s", level)
	}
}

// TestPerformanceConfig tests configuration
func TestPerformanceConfig(t *testing.T) {
	config := DefaultPerformanceConfig()

	if config.MaxMemoryMB != 550 {
		t.Errorf("Expected MaxMemoryMB 550, got %d", config.MaxMemoryMB)
	}

	if config.CacheSizeMB != 400 {
		t.Errorf("Expected CacheSizeMB 400, got %d", config.CacheSizeMB)
	}

	if config.TargetBandwidthKBs != 85 {
		t.Errorf("Expected TargetBandwidthKBs 85, got %d", config.TargetBandwidthKBs)
	}
}

// TestMemoryWarning tests memory warning threshold
func TestMemoryWarning(t *testing.T) {
	pm := NewPerformanceMonitor()

	// Set memory near limit
	stats := &MemoryStats{
		TotalBytes: 500 * 1024 * 1024,
		TotalMB:    500,
	}
	pm.UpdateMemoryStats(stats)

	// Should trigger warning at 90% (495MB of 550MB)
	if !pm.CheckMemoryWarning() {
		t.Error("Expected memory warning at 500MB (90% of 550MB)")
	}

	// Lower memory
	stats.TotalMB = 400
	pm.UpdateMemoryStats(stats)

	if pm.CheckMemoryWarning() {
		t.Error("Expected no warning at 400MB (72% of 550MB)")
	}
}

// BenchmarkMemoryProfilerTrack benchmarks memory tracking
func BenchmarkMemoryProfilerTrack(b *testing.B) {
	mp := NewMemoryProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.TrackAllocation("test", 1024)
	}
}

// BenchmarkCacheSet benchmarks cache set operations
func BenchmarkCacheSet(b *testing.B) {
	cm := NewCacheManager(1024 * 1024 * 1024) // 1GB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Set("key", "data", 1024)
	}
}

// BenchmarkCacheGet benchmarks cache get operations
func BenchmarkCacheGet(b *testing.B) {
	cm := NewCacheManager(1024 * 1024 * 1024)
	cm.Set("key", "data", 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Get("key")
	}
}

// BenchmarkLODGetLevel benchmarks LOD level calculation
func BenchmarkLODGetLevel(b *testing.B) {
	lm := NewLODManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.GetLODLevel(250.0)
	}
}
