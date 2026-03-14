// Package cache provides sprite caching with predictive warming.
// This file implements predictive cache warming based on access pattern analysis.
package cache

import (
	"sync"
)

// PredictiveCacheWarmer tracks sprite access patterns and predicts which sprites
// will be needed soon, enabling proactive cache warming for higher hit rates.
//
// The warmer uses a sliding window of recent accesses to identify:
// - Frequently accessed sprites (hot sprites)
// - Sequential access patterns (animation frames)
// - Nearby entity sprites (spatial locality)
//
// Target: >98% cache hit rate by preloading predicted sprites.
type PredictiveCacheWarmer struct {
	mu sync.RWMutex

	cache     *SpriteCache
	pregen    *PreGenerator
	accessLog []AccessRecord
	patterns  map[CacheKey]*AccessPattern

	// Configuration
	windowSize     int // Number of accesses to track
	hotThreshold   int // Accesses in window to be considered "hot"
	maxPredictions int // Maximum predictions per warmup cycle
}

// AccessRecord tracks a single cache access.
type AccessRecord struct {
	Key       CacheKey
	Hit       bool
	Timestamp int64 // Game tick or frame number
}

// AccessPattern tracks access statistics for a single sprite.
type AccessPattern struct {
	Key         CacheKey
	AccessCount int
	LastAccess  int64
	NextKeys    []CacheKey // Keys typically accessed after this one
}

// PredictiveWarmerConfig configures the predictive warmer.
type PredictiveWarmerConfig struct {
	WindowSize     int // Access history size (default: 1000)
	HotThreshold   int // Accesses to be "hot" (default: 5)
	MaxPredictions int // Max predictions per cycle (default: 50)
}

// DefaultWarmerConfig returns sensible defaults for predictive warming.
func DefaultWarmerConfig() PredictiveWarmerConfig {
	return PredictiveWarmerConfig{
		WindowSize:     1000,
		HotThreshold:   5,
		MaxPredictions: 50,
	}
}

// NewPredictiveCacheWarmer creates a predictive cache warmer.
func NewPredictiveCacheWarmer(cache *SpriteCache, pregen *PreGenerator, config PredictiveWarmerConfig) *PredictiveCacheWarmer {
	if config.WindowSize <= 0 {
		config.WindowSize = 1000 // Default: track last 1000 accesses.
	}
	if config.HotThreshold <= 0 {
		config.HotThreshold = 5 // Default: 5 accesses makes a key "hot".
	}
	if config.MaxPredictions <= 0 {
		config.MaxPredictions = 50 // Default: pre-warm up to 50 sprites per cycle.
	}

	return &PredictiveCacheWarmer{
		cache:          cache,
		pregen:         pregen,
		accessLog:      make([]AccessRecord, 0, config.WindowSize),
		patterns:       make(map[CacheKey]*AccessPattern),
		windowSize:     config.WindowSize,
		hotThreshold:   config.HotThreshold,
		maxPredictions: config.MaxPredictions,
	}
}

// RecordAccess logs a cache access for pattern analysis.
// Call this after every cache Get() operation.
func (w *PredictiveCacheWarmer) RecordAccess(key CacheKey, hit bool, tick int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Add record to sliding window
	record := AccessRecord{
		Key:       key,
		Hit:       hit,
		Timestamp: tick,
	}

	w.accessLog = append(w.accessLog, record)

	// Trim window if too large
	if len(w.accessLog) > w.windowSize {
		w.accessLog = w.accessLog[1:]
	}

	// Update pattern tracking
	pattern, exists := w.patterns[key]
	if !exists {
		pattern = &AccessPattern{
			Key:      key,
			NextKeys: make([]CacheKey, 0, 4),
		}
		w.patterns[key] = pattern
	}

	pattern.AccessCount++
	pattern.LastAccess = tick

	// Track sequential access patterns
	if len(w.accessLog) >= 2 {
		prevKey := w.accessLog[len(w.accessLog)-2].Key
		if prevPattern, ok := w.patterns[prevKey]; ok {
			// Add this key as a "next" key if not already tracked
			found := false
			for _, k := range prevPattern.NextKeys {
				if k == key {
					found = true
					break
				}
			}
			if !found && len(prevPattern.NextKeys) < 8 {
				prevPattern.NextKeys = append(prevPattern.NextKeys, key)
			}
		}
	}
}

// GetHotSprites returns the most frequently accessed sprites in the window.
func (w *PredictiveCacheWarmer) GetHotSprites() []CacheKey {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var hot []CacheKey
	for key, pattern := range w.patterns {
		if pattern.AccessCount >= w.hotThreshold {
			hot = append(hot, key)
		}
	}
	return hot
}

// PredictNext returns sprites likely to be accessed soon based on patterns.
// Analyzes the most recent access and returns sprites that typically follow it.
func (w *PredictiveCacheWarmer) PredictNext() []CacheKey {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.predictNextLocked()
}

// predictNextLocked is the lock-free implementation of PredictNext.
// Caller must hold at least a read lock on w.mu.
func (w *PredictiveCacheWarmer) predictNextLocked() []CacheKey {
	if len(w.accessLog) == 0 {
		return nil
	}

	// Get the most recent access
	lastKey := w.accessLog[len(w.accessLog)-1].Key

	pattern, exists := w.patterns[lastKey]
	if !exists || len(pattern.NextKeys) == 0 {
		return nil
	}

	// Return predicted next sprites that aren't already cached
	var predictions []CacheKey
	for _, nextKey := range pattern.NextKeys {
		if !w.cache.Contains(nextKey) {
			predictions = append(predictions, nextKey)
		}
		if len(predictions) >= w.maxPredictions {
			break
		}
	}

	return predictions
}

// WarmerStats contains predictive warmer statistics.
type WarmerStats struct {
	AccessLogSize   int
	PatternCount    int
	HotSpriteCount  int
	PredictionCount int
	WindowHitRate   float64 // Hit rate within sliding window
	WindowMissRate  float64 // Miss rate within sliding window
}

// Stats returns current warmer statistics.
func (w *PredictiveCacheWarmer) Stats() WarmerStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	hits := 0
	misses := 0
	for _, record := range w.accessLog {
		if record.Hit {
			hits++
		} else {
			misses++
		}
	}

	total := hits + misses
	hitRate := 0.0
	missRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total)
		missRate = float64(misses) / float64(total)
	}

	hotCount := 0
	for _, pattern := range w.patterns {
		if pattern.AccessCount >= w.hotThreshold {
			hotCount++
		}
	}

	return WarmerStats{
		AccessLogSize:   len(w.accessLog),
		PatternCount:    len(w.patterns),
		HotSpriteCount:  hotCount,
		PredictionCount: len(w.predictNextLocked()),
		WindowHitRate:   hitRate,
		WindowMissRate:  missRate,
	}
}

// Reset clears all access history and patterns.
func (w *PredictiveCacheWarmer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.accessLog = make([]AccessRecord, 0, w.windowSize)
	w.patterns = make(map[CacheKey]*AccessPattern)
}

// QueuePredictedSprites queues predicted sprites for pre-generation.
// Call this periodically (e.g., every 60 frames) to maintain high hit rates.
// The generator function is used to create sprites for predicted keys.
func (w *PredictiveCacheWarmer) QueuePredictedSprites(generator func(key CacheKey) GeneratorFunc) int {
	predictions := w.PredictNext()

	for _, key := range predictions {
		if genFunc := generator(key); genFunc != nil {
			w.pregen.Queue(key, genFunc)
		}
	}

	return len(predictions)
}

// AnalyzeAnimationSequence analyzes a sequence of animation frame keys
// and registers them as a sequential access pattern.
// This enables preloading entire animation sequences when the first frame is accessed.
func (w *PredictiveCacheWarmer) AnalyzeAnimationSequence(frameKeys []CacheKey) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := 0; i < len(frameKeys)-1; i++ {
		key := frameKeys[i]
		nextKey := frameKeys[i+1]

		pattern, exists := w.patterns[key]
		if !exists {
			pattern = &AccessPattern{
				Key:      key,
				NextKeys: make([]CacheKey, 0, 8),
			}
			w.patterns[key] = pattern
		}

		// Add next frame as predicted access
		found := false
		for _, k := range pattern.NextKeys {
			if k == nextKey {
				found = true
				break
			}
		}
		if !found && len(pattern.NextKeys) < 8 {
			pattern.NextKeys = append(pattern.NextKeys, nextKey)
		}
	}
}

// PredictAnimationFrames returns all predicted frames for a given sprite.
// Useful for preloading entire animation sequences.
func (w *PredictiveCacheWarmer) PredictAnimationFrames(startKey CacheKey) []CacheKey {
	w.mu.RLock()
	defer w.mu.RUnlock()

	visited := make(map[CacheKey]bool)
	var frames []CacheKey
	currentKey := startKey

	// Follow the chain of next keys
	for i := 0; i < 16; i++ { // Max 16 frames to prevent infinite loops
		pattern, exists := w.patterns[currentKey]
		if !exists || len(pattern.NextKeys) == 0 {
			break
		}

		nextKey := pattern.NextKeys[0] // Take first predicted next
		if visited[nextKey] {
			break // Avoid cycles
		}

		frames = append(frames, nextKey)
		visited[nextKey] = true
		currentKey = nextKey
	}

	return frames
}
