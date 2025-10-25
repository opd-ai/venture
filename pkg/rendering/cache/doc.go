// Package cache provides caching mechanisms for rendered sprites and images.
//
// The cache package implements an LRU (Least Recently Used) cache for Ebiten images,
// specifically designed for caching procedurally generated sprites, animation frames,
// and composite images. This reduces redundant generation and improves performance.
//
// Key features:
//   - LRU eviction policy to manage memory usage
//   - Size-based limits (configurable max cache size in bytes)
//   - Thread-safe operations with fine-grained locking
//   - Cache hit/miss statistics for monitoring
//   - Configurable eviction callbacks for cleanup
//
// Usage:
//
//	cache := cache.NewSpriteCache(200 * 1024 * 1024) // 200MB limit
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
//	fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate()*100)
package cache
