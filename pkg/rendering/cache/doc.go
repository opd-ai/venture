// Package cache provides caching mechanisms for rendered sprites and images.
//
// The cache package implements an LRU (Least Recently Used) cache for Ebiten images,
// specifically designed for caching procedurally generated sprites, animation frames,
// and composite images. This reduces redundant generation and improves performance.
//
// Phase 44 enhancements:
//   - Support for 64x64 sprites (16KB each)
//   - MemoryMonitor for automatic cleanup (<300MB limit)
//   - PreGenerator for batch sprite pre-generation
//   - Enhanced statistics and monitoring
//
// Key features:
//   - LRU eviction policy to manage memory usage
//   - Size-based limits (configurable max cache size in bytes)
//   - Thread-safe operations with fine-grained locking
//   - Cache hit/miss statistics for monitoring
//   - Memory monitoring with soft/hard limits
//   - Batch pre-generation for cache warming
//
// Basic Usage:
//
//	cache := cache.NewSpriteCache(300 * 1024 * 1024) // 300MB limit
//	key := cache.GenerateKey(seed, "idle", 0)
//
//	// Try to get from cache
//	if img, ok := cache.Get(key); ok {
//	    return img // Cache hit
//	}
//
//	// Generate and store in cache
//	img := generateSprite(seed, "idle", 0)
//	cache.Put(key, img)
//
//	// Check statistics
//	stats := cache.Stats()
//	logrus.WithField("hit_rate", stats.HitRate()*100).Info("cache statistics")
//
// Memory Monitoring (Phase 44):
//
//	monitor := cache.NewMemoryMonitor(cache)
//	monitor.SetLimits(250*1024*1024, 300*1024*1024) // 250MB soft, 300MB hard
//	monitor.Start()
//	defer monitor.Stop()
//
//	// Check health
//	if monitor.IsHealthy() {
//	    logrus.WithField("usage_percent", monitor.UsagePercentage()).Info("cache health status")
//	}
//
// Pre-Generation (Phase 44):
//
//	pregen := cache.NewPreGenerator(cache)
//
//	// Queue sprites for pre-generation
//	for i := 0; i < 100; i++ {
//	    key := cache.GenerateKey(int64(i), "idle", 0)
//	    pregen.Queue(key, func() (*ebiten.Image, error) {
//	        return generateSprite(i), nil
//	    })
//	}
//
//	// Generate in batch
//	count := pregen.Generate()
//	logrus.WithField("sprite_count", count).Info("pre-generated sprites")
package cache
