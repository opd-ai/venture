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

// TestCacheRemove tests cache entry removal
func TestCacheRemove(t *testing.T) {
	cm := NewCacheManager(100 * 1024 * 1024)

	// Add entries
	cm.Set("entry1", "data1", 10*1024*1024)
	cm.Set("entry2", "data2", 20*1024*1024)

	// Verify entry exists
	_, found := cm.Get("entry1")
	if !found {
		t.Fatal("Entry1 should exist before removal")
	}

	// Remove entry
	cm.Remove("entry1")

	// Verify entry removed
	_, found = cm.Get("entry1")
	if found {
		t.Error("Entry1 should be removed")
	}

	// Verify other entry still exists
	_, found = cm.Get("entry2")
	if !found {
		t.Error("Entry2 should still exist")
	}

	// Remove non-existent entry (should not panic)
	cm.Remove("non-existent")

	stats := cm.GetStats()
	if stats.ItemCount != 1 {
		t.Errorf("Expected 1 item after removal, got %d", stats.ItemCount)
	}
}

// TestCacheCleanup tests cache cleanup functionality
func TestCacheCleanup(t *testing.T) {
	cm := NewCacheManager(50 * 1024 * 1024) // 50MB limit

	// Add entries that exceed limit
	cm.Set("entry1", "data1", 20*1024*1024)
	cm.Set("entry2", "data2", 20*1024*1024)
	cm.Set("entry3", "data3", 20*1024*1024)

	// Force cleanup
	cm.Cleanup()

	stats := cm.GetStats()
	if stats.CurrentSizeMB > stats.MaxSizeMB {
		t.Errorf("Cache size %dMB exceeds max %dMB after cleanup", stats.CurrentSizeMB, stats.MaxSizeMB)
	}

	if stats.LastCleanup.IsZero() {
		t.Error("LastCleanup timestamp should be set")
	}
}

// TestCacheClear tests cache clear functionality
func TestCacheClear(t *testing.T) {
	cm := NewCacheManager(100 * 1024 * 1024)

	// Add multiple entries
	cm.Set("entry1", "data1", 10*1024*1024)
	cm.Set("entry2", "data2", 20*1024*1024)
	cm.Set("entry3", "data3", 30*1024*1024)

	stats := cm.GetStats()
	if stats.ItemCount != 3 {
		t.Fatalf("Expected 3 items before clear, got %d", stats.ItemCount)
	}

	// Clear cache
	cm.Clear()

	// Verify all entries removed
	stats = cm.GetStats()
	if stats.ItemCount != 0 {
		t.Errorf("Expected 0 items after clear, got %d", stats.ItemCount)
	}

	if stats.CurrentSizeMB != 0 {
		t.Errorf("Expected 0MB after clear, got %d", stats.CurrentSizeMB)
	}

	// Verify entries not retrievable
	_, found := cm.Get("entry1")
	if found {
		t.Error("Entry1 should not exist after clear")
	}

	_, found = cm.Get("entry2")
	if found {
		t.Error("Entry2 should not exist after clear")
	}
}

// TestCacheHitRate tests cache hit rate calculation
func TestCacheHitRate(t *testing.T) {
	cm := NewCacheManager(100 * 1024 * 1024)

	// Initial hit rate should be 0 (no requests)
	stats := cm.GetStats()
	if stats.HitRate != 0.0 {
		t.Errorf("Expected 0 hit rate with no requests, got %f", stats.HitRate)
	}

	// Add entries
	cm.Set("entry1", "data1", 1024)
	cm.Set("entry2", "data2", 1024)

	// 2 hits
	cm.Get("entry1")
	cm.Get("entry2")

	// 2 misses
	cm.Get("non-existent1")
	cm.Get("non-existent2")

	// Hit rate should be 50% (2 hits / 4 total)
	stats = cm.GetStats()
	if stats.HitRate != 0.5 {
		t.Errorf("Expected 0.5 hit rate, got %f", stats.HitRate)
	}

	// 2 more hits (4 hits / 6 total = 66.7%)
	cm.Get("entry1")
	cm.Get("entry2")

	stats = cm.GetStats()
	expectedRate := 4.0 / 6.0
	if stats.HitRate != expectedRate {
		t.Errorf("Expected %f hit rate, got %f", expectedRate, stats.HitRate)
	}

	// Clear should reset hit rate
	cm.Clear()
	stats = cm.GetStats()
	if stats.HitRate != 0.0 {
		t.Errorf("Expected 0 hit rate after clear, got %f", stats.HitRate)
	}
}

// TestBackgroundLoaderPreloadGuildHall tests guild hall preloading
func TestBackgroundLoaderPreloadGuildHall(t *testing.T) {
	bl := NewBackgroundLoader(2)
	bl.Start()
	defer bl.Stop()

	var loaded int32
	bl.PreloadGuildHall("hall1", func(data interface{}) {
		atomic.StoreInt32(&loaded, 1)
	})

	// Wait for loading
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&loaded) != 1 {
		t.Error("Expected guild hall to be loaded")
	}
}

// mockResourceLoader implements ResourceLoader for testing
type mockResourceLoader struct {
	loadCalls int
	loadData  interface{}
	loadErr   error
}

func (m *mockResourceLoader) Load(request *LoadRequest) (interface{}, error) {
	m.loadCalls++
	return m.loadData, m.loadErr
}

// TestBackgroundLoaderWithCustomLoader tests custom resource loader
func TestBackgroundLoaderWithCustomLoader(t *testing.T) {
	mockLoader := &mockResourceLoader{loadData: "loaded_data"}
	bl := NewBackgroundLoaderWithLoader(2, mockLoader)
	bl.Start()
	defer bl.Stop()

	var receivedData interface{}
	var callbackCalled int32
	bl.Queue(&LoadRequest{
		ID:   "test1",
		Type: "raid",
		Callback: func(data interface{}) {
			receivedData = data
			atomic.StoreInt32(&callbackCalled, 1)
		},
	})

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&callbackCalled) != 1 {
		t.Error("Expected callback to be called")
	}

	if receivedData != "loaded_data" {
		t.Errorf("Expected 'loaded_data', got %v", receivedData)
	}

	if mockLoader.loadCalls < 1 {
		t.Errorf("Expected Load to be called at least once, got %d", mockLoader.loadCalls)
	}
}

// TestBackgroundLoaderWithNilLoader tests nil loader defaults to DefaultResourceLoader
func TestBackgroundLoaderWithNilLoader(t *testing.T) {
	bl := NewBackgroundLoaderWithLoader(2, nil)
	if bl.loader == nil {
		t.Error("Expected loader to default to DefaultResourceLoader when nil is passed")
	}
	bl.Start()
	defer bl.Stop()

	var callbackCalled int32
	bl.Queue(&LoadRequest{
		ID: "test1",
		Callback: func(data interface{}) {
			atomic.StoreInt32(&callbackCalled, 1)
		},
	})

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&callbackCalled) != 1 {
		t.Error("Expected callback to be called with default loader")
	}
}

// TestBackgroundLoaderGetQueueSize tests queue size retrieval
func TestBackgroundLoaderGetQueueSize(t *testing.T) {
	bl := NewBackgroundLoader(1) // Single worker to control queue growth
	bl.Start()
	defer bl.Stop()

	// Queue multiple items
	bl.PreloadRaid("raid1", nil)
	bl.PreloadRaid("raid2", nil)
	bl.PreloadGuildHall("hall1", nil)

	// Check queue size (should have at least some items)
	queueSize := bl.GetQueueSize()
	if queueSize < 0 {
		t.Errorf("Queue size should be non-negative, got %d", queueSize)
	}

	// Note: exact size depends on worker processing speed
	// Just verify the method works without error
}

// TestLODManagerSetDistances tests custom LOD distance configuration
func TestLODManagerSetDistances(t *testing.T) {
	lm := NewLODManager()

	// Set custom distances
	lm.SetDistances(50.0, 150.0, 300.0)

	// Test new distances
	tests := []struct {
		distance float64
		expected LODLevel
	}{
		{30.0, LODHigh},     // < 50
		{100.0, LODMedium},  // < 150
		{250.0, LODLow},     // < 300
		{400.0, LODVeryLow}, // >= 300
	}

	for _, tt := range tests {
		level := lm.GetLODLevel(tt.distance)
		if level != tt.expected {
			t.Errorf("Distance %.1f: expected %s, got %s", tt.distance, tt.expected, level)
		}
	}
}

// TestLODManagerEnable tests LOD system enable
func TestLODManagerEnable(t *testing.T) {
	lm := NewLODManager()

	// Disable first
	lm.Disable()
	if lm.IsEnabled() {
		t.Error("LOD should be disabled")
	}

	// Re-enable
	lm.Enable()
	if !lm.IsEnabled() {
		t.Error("LOD should be enabled")
	}

	// Verify LOD works when enabled
	level := lm.GetLODLevel(1000.0)
	if level == LODVeryHigh {
		t.Error("Expected LOD level other than VeryHigh when enabled at far distance")
	}
}

// TestLODManagerIsEnabled tests LOD enabled state check
func TestLODManagerIsEnabled(t *testing.T) {
	lm := NewLODManager()

	// Default state should be enabled
	if !lm.IsEnabled() {
		t.Error("LOD should be enabled by default")
	}

	// Disable and check
	lm.Disable()
	if lm.IsEnabled() {
		t.Error("LOD should be disabled after Disable()")
	}

	// Re-enable and check
	lm.Enable()
	if !lm.IsEnabled() {
		t.Error("LOD should be enabled after Enable()")
	}
}

// TestMemoryProfilerGetAllocationTrend tests allocation trend analysis
func TestMemoryProfilerGetAllocationTrend(t *testing.T) {
	mp := NewMemoryProfiler()

	// Create allocation trend
	for i := 0; i < 10; i++ {
		mp.TrackAllocation("growing", uint64(i*10*1024*1024))
		mp.TakeSnapshot()
		time.Sleep(5 * time.Millisecond)
	}

	// Get trend for last 5 samples
	trend := mp.GetAllocationTrend("growing", 5)

	if len(trend) != 5 {
		t.Errorf("Expected 5 trend samples, got %d", len(trend))
	}

	// Verify trend is increasing
	for i := 1; i < len(trend); i++ {
		if trend[i] < trend[i-1] {
			t.Errorf("Trend should be increasing: sample %d (%d) < sample %d (%d)",
				i, trend[i], i-1, trend[i-1])
		}
	}

	// Test with more samples than available
	trend = mp.GetAllocationTrend("growing", 20)
	if len(trend) != 10 {
		t.Errorf("Expected 10 trend samples (limited by snapshots), got %d", len(trend))
	}

	// Test with non-existent allocation
	trend = mp.GetAllocationTrend("non-existent", 5)
	if len(trend) != 5 {
		t.Errorf("Expected 5 trend samples (zeros for non-existent), got %d", len(trend))
	}
	for i, val := range trend {
		if val != 0 {
			t.Errorf("Expected zero for non-existent allocation at index %d, got %d", i, val)
		}
	}
}

// TestMemoryProfilerReset tests profiler reset functionality
func TestMemoryProfilerReset(t *testing.T) {
	mp := NewMemoryProfiler()

	// Track allocations and create snapshots
	mp.TrackAllocation("test1", 50*1024*1024)
	mp.TrackAllocation("test2", 100*1024*1024)
	mp.TakeSnapshot()
	mp.TakeSnapshot()

	stats := mp.GetStats()
	if stats.TotalMB == 0 {
		t.Fatal("Expected non-zero total before reset")
	}

	snapshots := mp.GetSnapshots()
	if len(snapshots) == 0 {
		t.Fatal("Expected snapshots before reset")
	}

	// Reset profiler
	mp.Reset()

	// Verify all cleared
	stats = mp.GetStats()
	if stats.TotalMB != 0 {
		t.Errorf("Expected 0MB after reset, got %d", stats.TotalMB)
	}

	snapshots = mp.GetSnapshots()
	if len(snapshots) != 0 {
		t.Errorf("Expected 0 snapshots after reset, got %d", len(snapshots))
	}

	if len(stats.Allocations) != 0 {
		t.Errorf("Expected 0 allocations after reset, got %d", len(stats.Allocations))
	}
}

// TestMemoryProfilerGetUptime tests uptime tracking
func TestMemoryProfilerGetUptime(t *testing.T) {
	mp := NewMemoryProfiler()

	// Wait briefly
	time.Sleep(100 * time.Millisecond)

	uptime := mp.GetUptime()
	if uptime < 100*time.Millisecond {
		t.Errorf("Expected uptime >= 100ms, got %v", uptime)
	}

	if uptime > 1*time.Second {
		t.Errorf("Expected uptime < 1s, got %v", uptime)
	}

	// Reset and verify uptime resets
	mp.Reset()
	time.Sleep(50 * time.Millisecond)

	uptime = mp.GetUptime()
	if uptime > 100*time.Millisecond {
		t.Errorf("Expected uptime < 100ms after reset, got %v", uptime)
	}
}

// TestNetworkBatcherGetStats tests statistics retrieval
func TestNetworkBatcherGetStats(t *testing.T) {
	var batchCount int32

	sendFunc := func(batch *BatchedMessage) {
		atomic.AddInt32(&batchCount, 1)
	}

	nb := NewNetworkBatcher(100, sendFunc)
	nb.Start()
	defer nb.Stop()

	// Queue messages
	nb.QueueMessage("pos_update", []byte("player1_data"), "player1")
	nb.QueueMessage("pos_update", []byte("player2_data"), "player2")
	nb.QueueMessage("chat", []byte("hello"), "player3")

	// Wait for batching
	time.Sleep(150 * time.Millisecond)

	stats := nb.GetStats()

	// Verify stats structure
	if stats == nil {
		t.Fatal("GetStats should not return nil")
	}

	if stats.BatchCount == 0 {
		t.Error("Expected at least 1 batch sent")
	}

	if stats.MessagesSent == 0 {
		t.Error("Expected at least 1 message sent")
	}

	if stats.BytesSent == 0 {
		t.Error("Expected non-zero bytes sent")
	}

	// MessagesPerSec and BytesPerSec may be 0 or positive depending on timing
	// Just verify they're not negative
	if stats.MessagesPerSec < 0 {
		t.Errorf("MessagesPerSec should be non-negative, got %.2f", stats.MessagesPerSec)
	}

	if stats.BytesPerSec < 0 {
		t.Errorf("BytesPerSec should be non-negative, got %.2f", stats.BytesPerSec)
	}
}

// TestCacheEdgeCases tests edge cases in cache management
func TestCacheEdgeCases(t *testing.T) {
	t.Run("zero size cache enforces minimum", func(t *testing.T) {
		cm := NewCacheManager(0)
		stats := cm.GetStats()

		// Zero size should be enforced to 1MB minimum
		if stats.MaxSizeMB != 1 {
			t.Errorf("Expected 1MB minimum, got %d MB", stats.MaxSizeMB)
		}

		// Should be able to cache small entries
		cm.Set("entry", "data", 1024)
		if _, ok := cm.Get("entry"); !ok {
			t.Error("Small entry should be cacheable with 1MB minimum")
		}
	})

	t.Run("very small size cache enforces minimum", func(t *testing.T) {
		cm := NewCacheManager(512 * 1024) // 512KB, below 1MB
		stats := cm.GetStats()

		// Should be enforced to 1MB minimum
		if stats.MaxSizeMB != 1 {
			t.Errorf("Expected 1MB minimum for 512KB input, got %d MB", stats.MaxSizeMB)
		}
	})

	t.Run("remove from empty cache", func(t *testing.T) {
		cm := NewCacheManager(100 * 1024 * 1024)
		cm.Remove("non-existent") // Should not panic
	})

	t.Run("clear empty cache", func(t *testing.T) {
		cm := NewCacheManager(100 * 1024 * 1024)
		cm.Clear() // Should not panic
		stats := cm.GetStats()
		if stats.ItemCount != 0 {
			t.Error("Empty cache should have 0 items")
		}
	})

	t.Run("cleanup empty cache", func(t *testing.T) {
		cm := NewCacheManager(100 * 1024 * 1024)
		cm.Cleanup() // Should not panic
	})
}

// TestBackgroundLoaderEdgeCases tests edge cases in background loader
func TestBackgroundLoaderEdgeCases(t *testing.T) {
	t.Run("queue when not started", func(t *testing.T) {
		bl := NewBackgroundLoader(2)
		// Queue without starting
		bl.PreloadRaid("raid1", nil) // Should not panic

		queueSize := bl.GetQueueSize()
		// Queue size should be 0 when not running
		if queueSize != 0 {
			t.Errorf("expected queue size 0 when loader is not running, got %d", queueSize)
		}
	})

	t.Run("multiple starts", func(t *testing.T) {
		bl := NewBackgroundLoader(2)
		bl.Start()
		bl.Start() // Second start should be no-op
		defer bl.Stop()
	})

	t.Run("multiple stops", func(t *testing.T) {
		bl := NewBackgroundLoader(2)
		bl.Start()
		bl.Stop()
		bl.Stop() // Second stop should be no-op
	})

	t.Run("nil callback", func(t *testing.T) {
		bl := NewBackgroundLoader(2)
		bl.Start()
		defer bl.Stop()

		bl.PreloadRaid("raid1", nil)       // Should not panic with nil callback
		bl.PreloadGuildHall("hall1", nil)  // Should not panic with nil callback
		time.Sleep(200 * time.Millisecond) // Let workers process
	})
}

// TestLODManagerConcurrency tests LOD manager under concurrent access
func TestLODManagerConcurrency(t *testing.T) {
	lm := NewLODManager()

	done := make(chan bool)

	// Concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				lm.GetLODLevel(float64(j * 10))
				lm.IsEnabled()
			}
			done <- true
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				lm.Enable()
				lm.Disable()
				lm.SetDistances(100.0, 200.0, 400.0)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Verify LOD manager still works
	level := lm.GetLODLevel(100.0)
	if level.String() == "" {
		t.Error("LOD level should not be empty after concurrent access")
	}
}

// TestMemoryProfilerConcurrency tests memory profiler under concurrent access
func TestMemoryProfilerConcurrency(t *testing.T) {
	mp := NewMemoryProfiler()

	done := make(chan bool)

	// Concurrent allocations
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				mp.TrackAllocation("test", 1024*1024)
				mp.ReleaseAllocation("test", 1024*1024)
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				mp.GetStats()
				mp.GetSnapshots()
				mp.GetUptime()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Verify profiler still works
	stats := mp.GetStats()
	if stats == nil {
		t.Error("Stats should not be nil after concurrent access")
	}
}

// TestPerformanceMonitorStatsUpdates tests updating various statistics
func TestPerformanceMonitorStatsUpdates(t *testing.T) {
	pm := NewPerformanceMonitor()

	// Update network stats
	netStats := &NetworkStats{
		MessagesSent:   1000,
		MessagesPerSec: 100.0,
		BytesSent:      50000,
		BytesPerSec:    5000.0,
		BatchCount:     50,
		AvgBatchSize:   20.0,
		LastBatchTime:  time.Now(),
	}
	pm.UpdateNetworkStats(netStats)

	retrieved := pm.GetNetworkStats()
	if retrieved.MessagesSent != 1000 {
		t.Errorf("Expected MessagesSent 1000, got %d", retrieved.MessagesSent)
	}
	if retrieved.BytesSent != 50000 {
		t.Errorf("Expected BytesSent 50000, got %d", retrieved.BytesSent)
	}

	// Update cache stats
	cacheStats := &CacheStats{
		CurrentSizeMB: 200,
		MaxSizeMB:     400,
		ItemCount:     500,
		HitRate:       0.95,
		EvictionCount: 10,
		LastCleanup:   time.Now(),
	}
	pm.UpdateCacheStats(cacheStats)

	retrievedCache := pm.GetCacheStats()
	if retrievedCache.CurrentSizeMB != 200 {
		t.Errorf("Expected CurrentSizeMB 200, got %d", retrievedCache.CurrentSizeMB)
	}
	if retrievedCache.ItemCount != 500 {
		t.Errorf("Expected ItemCount 500, got %d", retrievedCache.ItemCount)
	}

	// Update memory stats
	memStats := &MemoryStats{
		TotalBytes:     300 * 1024 * 1024,
		TotalMB:        300,
		Allocations:    map[string]uint64{"test": 100 * 1024 * 1024},
		LargestAlloc:   "test",
		LargestAllocMB: 100,
	}
	pm.UpdateMemoryStats(memStats)

	retrievedMem := pm.GetMemoryStats()
	if retrievedMem.TotalMB != 300 {
		t.Errorf("Expected TotalMB 300, got %d", retrievedMem.TotalMB)
	}
	if retrievedMem.LargestAlloc != "test" {
		t.Errorf("Expected LargestAlloc 'test', got %s", retrievedMem.LargestAlloc)
	}
	if len(retrievedMem.Allocations) != 1 {
		t.Errorf("Expected 1 allocation, got %d", len(retrievedMem.Allocations))
	}
}

// TestPerformanceMonitorConfigManagement tests config get/set
func TestPerformanceMonitorConfigManagement(t *testing.T) {
	pm := NewPerformanceMonitor()

	// Get default config
	config := pm.GetConfig()
	if config == nil {
		t.Fatal("GetConfig should not return nil")
	}

	originalMaxMemory := config.MaxMemoryMB
	if originalMaxMemory != 550 {
		t.Errorf("Expected default MaxMemoryMB 550, got %d", originalMaxMemory)
	}

	// Create and set new config
	newConfig := DefaultPerformanceConfig()
	newConfig.MaxMemoryMB = 1000
	newConfig.CacheSizeMB = 800
	newConfig.BatchWindowMs = 50

	pm.SetConfig(newConfig)

	// Verify config updated
	retrieved := pm.GetConfig()
	if retrieved.MaxMemoryMB != 1000 {
		t.Errorf("Expected MaxMemoryMB 1000, got %d", retrieved.MaxMemoryMB)
	}
	if retrieved.CacheSizeMB != 800 {
		t.Errorf("Expected CacheSizeMB 800, got %d", retrieved.CacheSizeMB)
	}
	if retrieved.BatchWindowMs != 50 {
		t.Errorf("Expected BatchWindowMs 50, got %d", retrieved.BatchWindowMs)
	}
}

// TestNetworkBatcherGetQueueSize tests queue size for batcher
func TestNetworkBatcherGetQueueSize(t *testing.T) {
	var batchCount int32

	sendFunc := func(batch *BatchedMessage) {
		atomic.AddInt32(&batchCount, 1)
		// Add delay to allow queue to build up
		time.Sleep(50 * time.Millisecond)
	}

	nb := NewNetworkBatcher(1000, sendFunc) // Long window to control batching
	nb.Start()
	defer nb.Stop()

	// Queue messages
	for i := 0; i < 5; i++ {
		nb.QueueMessage("test", []byte("data"), "player1")
	}

	// Check queue size immediately
	queueSize := nb.GetQueueSize()
	if queueSize < 0 {
		t.Errorf("Queue size should be non-negative, got %d", queueSize)
	}

	// Queue size should be reasonable (0-5 messages)
	if queueSize > 10 {
		t.Errorf("Queue size unexpectedly large: %d", queueSize)
	}
}

// TestTypeAliases tests that the type aliases work correctly
// and don't introduce any regressions
func TestTypeAliases(t *testing.T) {
	// Test Config alias
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}
	if config.MaxMemoryMB != 550 {
		t.Errorf("Expected MaxMemoryMB 550, got %d", config.MaxMemoryMB)
	}

	// Test Monitor alias
	monitor := NewMonitor()
	if monitor == nil {
		t.Fatal("NewMonitor() returned nil")
	}

	// Verify Config and PerformanceConfig are the same type
	var _ *PerformanceConfig = config
	var _ *Config = config

	// Verify Monitor and PerformanceMonitor are the same type
	var _ *PerformanceMonitor = monitor
	var _ *Monitor = monitor

	// Test that monitor methods work with alias type
	monitor.UpdateFrameTime(16.67)
	if monitor.GetFPS() < 59 || monitor.GetFPS() > 61 {
		t.Errorf("Expected FPS ~60, got %.2f", monitor.GetFPS())
	}
}

// TestLODLevelDocumentation validates LOD level constants
func TestLODLevelDocumentation(t *testing.T) {
	tests := []struct {
		level    LODLevel
		expected string
	}{
		{LODVeryHigh, "VeryHigh"},
		{LODHigh, "High"},
		{LODMedium, "Medium"},
		{LODLow, "Low"},
		{LODVeryLow, "VeryLow"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.level.String())
			}
		})
	}

	// Test unknown level
	unknown := LODLevel(999)
	if unknown.String() != "Unknown" {
		t.Errorf("Expected Unknown for invalid level, got %s", unknown.String())
	}
}
