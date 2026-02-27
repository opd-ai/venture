// Package performance provides performance optimization tools for Venture Phase 60.2.
//
// This package implements:
//   - Memory profiling and monitoring
//   - Network message batching
//   - Cache management and tuning
//   - Background resource loading
//   - Level-of-detail (LOD) system
//   - Database indexing and query optimization
//
// # Memory Profiling
//
// The MemoryProfiler tracks memory usage across all V9 systems:
//
//	profiler := performance.NewMemoryProfiler()
//	profiler.TrackAllocation("raid_dungeons", 1024*1024*50) // 50MB
//	stats := profiler.GetStats()
//	logrus.WithFields(logrus.Fields{
//	    "total_mb": stats.TotalMB,
//	    "peak_mb": stats.PeakMB,
//	}).Info("memory profiler stats")
//
// # Network Batching
//
// The NetworkBatcher combines multiple small messages into batched packets:
//
//	sendFunc := func(batch *BatchedMessage) {
//	    // Send the batched messages
//	}
//	batcher := performance.NewNetworkBatcher(100, sendFunc) // 100ms window
//	batcher.Start()
//	batcher.QueueMessage("guild_update", data, playerID)
//	batcher.QueueMessage("player_position", posData, playerID)
//	// Messages automatically batched and sent every 100ms
//	batcher.Stop()
//
// # Cache Management
//
// The CacheManager handles sprite and resource caching with V9 content:
//
//	cache := performance.NewCacheManager(400*1024*1024) // 400MB limit
//	cache.SetSprites(sprites)
//	cache.SetRaidData(raidID, data)
//	cache.Cleanup() // Evict LRU items if over limit
//
// # Background Loading
//
// The BackgroundLoader preloads resources during travel:
//
//	loader := performance.NewBackgroundLoader()
//	loader.PreloadRaid(raidID, func(data interface{}) {
//	    // Raid loaded and ready
//	})
//
// # Level of Detail
//
// The LODManager reduces detail for distant objects:
//
//	lodMgr := performance.NewLODManager()
//	lod := lodMgr.GetLODLevel(distanceFromCamera)
//	// Returns: LODHigh, LODMedium, LODLow, or LODVeryLow
//
// # Performance Targets
//
// Phase 60.2 targets:
//   - Memory: <550MB total (+50MB from V8)
//   - Network: <85KB/s (+10KB/s from V8)
//   - Load time: <3s for raid instances
//   - Frame time: <16.67ms (60 FPS maintained)
//
// All systems are thread-safe and designed for concurrent use.
package performance
