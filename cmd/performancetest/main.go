// Command performancetest demonstrates Phase 60.2 Performance Optimization.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/engine/performance"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, memory, network, cache, loader, lod, benchmark, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "memory":
		testMemoryProfiler(*verbose)
	case "network":
		testNetworkBatcher(*verbose)
	case "cache":
		testCacheManager(*verbose)
	case "loader":
		testBackgroundLoader(*verbose)
	case "lod":
		testLODManager(*verbose)
	case "benchmark":
		runBenchmark(*verbose)
	case "all":
		runDemo(*verbose)
		testMemoryProfiler(*verbose)
		testNetworkBatcher(*verbose)
		testCacheManager(*verbose)
		testBackgroundLoader(*verbose)
		testLODManager(*verbose)
		runBenchmark(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func runDemo(verbose bool) {
	fmt.Println("=== Phase 60.2: Performance Optimization Demo ===")
	fmt.Println()
	fmt.Println("Features implemented:")
	fmt.Println("✅ Memory Profiling (tracks allocations, detects leaks)")
	fmt.Println("✅ Network Batching (combines messages, reduces bandwidth)")
	fmt.Println("✅ Cache Management (400MB limit, LRU eviction)")
	fmt.Println("✅ Background Loading (preload raids during travel)")
	fmt.Println("✅ LOD System (4 quality levels based on distance)")
	fmt.Println("✅ Performance Monitoring (FPS, memory, network stats)")
	fmt.Println()

	// Show performance targets
	config := performance.DefaultPerformanceConfig()
	fmt.Println("V9 Performance Targets:")
	fmt.Printf("  Memory: %d MB (V8: 500MB)\n", config.MaxMemoryMB)
	fmt.Printf("  Cache: %d MB (V8: 300MB)\n", config.CacheSizeMB)
	fmt.Printf("  Bandwidth: %d KB/s (V8: 75KB/s)\n", config.TargetBandwidthKBs)
	fmt.Printf("  Frame Time: <16.67ms (60 FPS)\n")
	fmt.Printf("  Load Time: <%ds for raid instances\n", config.LoadTimeoutSec)
	fmt.Println()
}

func testMemoryProfiler(verbose bool) {
	fmt.Println("=== Testing Memory Profiler ===")

	profiler := performance.NewMemoryProfiler()

	// Simulate V9 allocations
	fmt.Println("\nTracking V9 allocations:")
	allocations := map[string]uint64{
		"sprites":       300 * 1024 * 1024,
		"raid_dungeons": 50 * 1024 * 1024,
		"guild_halls":   40 * 1024 * 1024,
		"terrain":       60 * 1024 * 1024,
		"companions":    20 * 1024 * 1024,
		"vehicles":      15 * 1024 * 1024,
		"narratives":    10 * 1024 * 1024,
	}

	for name, bytes := range allocations {
		profiler.TrackAllocation(name, bytes)
		if verbose {
			fmt.Printf("  %s: %d MB\n", name, bytes/(1024*1024))
		}
	}

	stats := profiler.GetStats()
	fmt.Printf("\nTotal Memory: %d MB\n", stats.TotalMB)
	fmt.Printf("Largest Allocation: %s (%d MB)\n", stats.LargestAlloc, stats.LargestAllocMB)

	// Test snapshots
	fmt.Println("\nTaking snapshots...")
	for i := 0; i < 5; i++ {
		profiler.TakeSnapshot()
		time.Sleep(100 * time.Millisecond)
	}

	snapshots := profiler.GetSnapshots()
	fmt.Printf("  Captured %d snapshots over %.1fs\n", len(snapshots), profiler.GetUptime().Seconds())

	// Test leak detection
	fmt.Println("\nChecking for memory leaks...")
	leaks := profiler.IdentifyLeaks(10)
	if len(leaks) > 0 {
		fmt.Printf("  ⚠ Found %d potential leaks: %v\n", len(leaks), leaks)
	} else {
		fmt.Println("  ✓ No memory leaks detected")
	}

	fmt.Println()
}

func testNetworkBatcher(verbose bool) {
	fmt.Println("=== Testing Network Batcher ===")

	batchesSent := 0
	sendFunc := func(batch *performance.BatchedMessage) {
		batchesSent++
		if verbose {
			fmt.Printf("  Batch sent: %d messages, %d bytes\n", len(batch.Messages), batch.Size)
		}
	}

	batcher := performance.NewNetworkBatcher(100, sendFunc)
	batcher.Start()
	defer batcher.Stop()

	// Queue messages
	fmt.Println("\nQueuing 10 messages...")
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("message_%d_data", i))
		batcher.QueueMessage("test", data, fmt.Sprintf("player%d", i%3))
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for batching
	time.Sleep(150 * time.Millisecond)

	stats := batcher.GetStats()
	fmt.Printf("\nNetwork Stats:\n")
	fmt.Printf("  Batches Sent: %d\n", stats.BatchCount)
	fmt.Printf("  Messages Sent: %d\n", stats.MessagesSent)
	fmt.Printf("  Bytes Sent: %d\n", stats.BytesSent)
	fmt.Printf("  Avg Batch Size: %.1f messages\n", stats.AvgBatchSize)
	fmt.Println()
}

func testCacheManager(verbose bool) {
	fmt.Println("=== Testing Cache Manager ===")

	cache := performance.NewCacheManager(400 * 1024 * 1024) // 400MB

	// Add cache entries
	fmt.Println("\nAdding cache entries...")
	entries := []struct {
		key  string
		size uint64
	}{
		{"raid_1_sprites", 50 * 1024 * 1024},
		{"raid_2_sprites", 50 * 1024 * 1024},
		{"guild_hall_1", 30 * 1024 * 1024},
		{"guild_hall_2", 30 * 1024 * 1024},
		{"terrain_chunk_1", 20 * 1024 * 1024},
		{"terrain_chunk_2", 20 * 1024 * 1024},
	}

	for _, entry := range entries {
		cache.Set(entry.key, fmt.Sprintf("data_%s", entry.key), entry.size)
		if verbose {
			fmt.Printf("  Set %s (%d MB)\n", entry.key, entry.size/(1024*1024))
		}
	}

	stats := cache.GetStats()
	fmt.Printf("\nCache Stats:\n")
	fmt.Printf("  Current Size: %d MB / %d MB\n", stats.CurrentSizeMB, stats.MaxSizeMB)
	fmt.Printf("  Item Count: %d\n", stats.ItemCount)
	fmt.Printf("  Hit Rate: %.1f%%\n", stats.HitRate*100)
	fmt.Printf("  Evictions: %d\n", stats.EvictionCount)

	// Test retrieval
	fmt.Println("\nRetrieving entries...")
	data, found := cache.Get("raid_1_sprites")
	if found {
		fmt.Printf("  ✓ Retrieved: %v\n", data)
	} else {
		fmt.Println("  ✗ Entry not found")
	}

	fmt.Println()
}

func testBackgroundLoader(verbose bool) {
	fmt.Println("=== Testing Background Loader ===")

	loader := performance.NewBackgroundLoader(4)
	loader.Start()
	defer loader.Stop()

	// Preload raids
	fmt.Println("\nPreloading raid dungeons...")
	loadedCount := 0
	for i := 1; i <= 5; i++ {
		raidID := fmt.Sprintf("raid_%d", i)
		loader.PreloadRaid(raidID, func(data interface{}) {
			loadedCount++
			if verbose {
				fmt.Printf("  ✓ Loaded: %s\n", raidID)
			}
		})
	}

	// Wait for loading
	time.Sleep(600 * time.Millisecond)

	fmt.Printf("\nLoaded %d/5 raids in background\n", loadedCount)
	fmt.Printf("Queue size: %d\n", loader.GetQueueSize())
	fmt.Println()
}

func testLODManager(verbose bool) {
	fmt.Println("=== Testing LOD Manager ===")

	lodMgr := performance.NewLODManager()

	distances := []float64{50, 150, 400, 700}

	fmt.Println("\nLOD levels by distance:")
	for _, distance := range distances {
		level := lodMgr.GetLODLevel(distance)
		fmt.Printf("  %6.1f units: %s\n", distance, level)
	}

	// Test disabled mode
	fmt.Println("\nWith LOD disabled:")
	lodMgr.Disable()
	for _, distance := range distances {
		level := lodMgr.GetLODLevel(distance)
		fmt.Printf("  %6.1f units: %s\n", distance, level)
	}

	fmt.Println()
}

func runBenchmark(verbose bool) {
	fmt.Println("=== Running Performance Benchmarks ===")

	// Simulate game loop performance
	monitor := performance.NewPerformanceMonitor()

	fmt.Println("\nSimulating 60 FPS game loop...")
	for i := 0; i < 10; i++ {
		start := time.Now()

		// Simulate frame work
		time.Sleep(10 * time.Millisecond)

		elapsed := time.Since(start)
		monitor.UpdateFrameTime(float64(elapsed.Milliseconds()))
	}

	fps := monitor.GetFPS()
	frameTime := monitor.GetFrameTime()

	fmt.Printf("\nPerformance Results:\n")
	fmt.Printf("  Frame Time: %.2f ms\n", frameTime)
	fmt.Printf("  FPS: %.1f\n", fps)
	fmt.Printf("  Target Met: %v\n", monitor.CheckPerformanceTarget())

	// Set memory stats
	memStats := &performance.MemoryStats{
		TotalBytes: 495 * 1024 * 1024,
		TotalMB:    495,
	}
	monitor.UpdateMemoryStats(memStats)

	fmt.Printf("\nMemory Status:\n")
	fmt.Printf("  Usage: %d MB / 550 MB\n", memStats.TotalMB)
	fmt.Printf("  Warning: %v\n", monitor.CheckMemoryWarning())

	fmt.Println()
}
