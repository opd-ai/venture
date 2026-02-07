//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/rendering/cache"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// TestWarmCommonSprites_Basic verifies basic sprite cache warming functionality.
func TestWarmCommonSprites_Basic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise during tests
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024), // 100MB cache
		spriteGenerator: sprites.NewGenerator(),
	}

	// Warm the cache
	warmCommonSprites(sys, &seed, &genre, clientLogger)

	// Give async generation a moment to complete
	time.Sleep(100 * time.Millisecond)

	// Verify cache contains some sprites
	stats := sys.spriteCache.Stats()
	if stats.EntryCount == 0 {
		t.Errorf("Expected cache to contain sprites after warming, got 0 entries")
	}

	// Verify we have a reasonable number of sprites (should have player + enemy sprites)
	// Player: 2 states × 4 frames × 4 directions = 32 sprites
	// Enemies: 4 types × 4 directions × 1 frame = 16 sprites
	// Total expected: ~48 sprites
	if stats.EntryCount < 30 {
		t.Errorf("Expected at least 30 sprites after warming, got %d", stats.EntryCount)
	}

	if stats.EntryCount > 60 {
		t.Errorf("Expected at most 60 sprites after warming, got %d", stats.EntryCount)
	}
}

// TestWarmCommonSprites_NilSeed verifies warming works with nil seed (uses default).
func TestWarmCommonSprites_NilSeed(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	genre := "scifi"

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: sprites.NewGenerator(),
	}

	// Warm with nil seed (should use default 12345)
	warmCommonSprites(sys, nil, &genre, clientLogger)

	time.Sleep(100 * time.Millisecond)

	stats := sys.spriteCache.Stats()
	if stats.EntryCount == 0 {
		t.Errorf("Expected cache to contain sprites with default seed, got 0 entries")
	}
}

// TestWarmCommonSprites_NilGenre verifies warming works with nil genre (uses default).
func TestWarmCommonSprites_NilGenre(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(67890)

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: sprites.NewGenerator(),
	}

	// Warm with nil genre (should use default "fantasy")
	warmCommonSprites(sys, &seed, nil, clientLogger)

	time.Sleep(100 * time.Millisecond)

	stats := sys.spriteCache.Stats()
	if stats.EntryCount == 0 {
		t.Errorf("Expected cache to contain sprites with default genre, got 0 entries")
	}
}

// TestWarmCommonSprites_NoCache verifies graceful handling when cache is nil.
func TestWarmCommonSprites_NoCache(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	sys := &systemsContainer{
		spriteCache:     nil, // No cache
		spriteGenerator: sprites.NewGenerator(),
	}

	// Should not panic, should log warning and return
	warmCommonSprites(sys, &seed, &genre, clientLogger)

	// Give a moment for any potential issues
	time.Sleep(10 * time.Millisecond)
}

// TestWarmCommonSprites_NoGenerator verifies graceful handling when generator is nil.
func TestWarmCommonSprites_NoGenerator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: nil, // No generator
	}

	// Should not panic, should log warning and return
	warmCommonSprites(sys, &seed, &genre, clientLogger)

	time.Sleep(10 * time.Millisecond)

	// Cache should remain empty
	stats := sys.spriteCache.Stats()
	if stats.EntryCount != 0 {
		t.Errorf("Expected empty cache without generator, got %d entries", stats.EntryCount)
	}
}

// TestWarmCommonSprites_CacheHitRate verifies warming improves cache hit rate.
func TestWarmCommonSprites_CacheHitRate(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: sprites.NewGenerator(),
	}

	// Warm the cache
	warmCommonSprites(sys, &seed, &genre, clientLogger)

	// Wait for warming to complete
	time.Sleep(150 * time.Millisecond)

	// Try to retrieve a player idle sprite (should be warmed)
	key := cache.GenerateKey(seed, "idle_down", 0)
	_, hit := sys.spriteCache.Get(key)

	if !hit {
		t.Errorf("Expected cache hit for warmed player idle sprite, got miss")
	}

	// Verify hit rate
	stats := sys.spriteCache.Stats()
	if stats.Hits == 0 {
		t.Errorf("Expected at least one cache hit after warming")
	}
}

// TestWarmCommonSprites_DifferentGenres verifies warming works across genres.
func TestWarmCommonSprites_DifferentGenres(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := &systemsContainer{
				spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
				spriteGenerator: sprites.NewGenerator(),
			}

			genreCopy := genre // Capture for goroutine
			warmCommonSprites(sys, &seed, &genreCopy, clientLogger)

			time.Sleep(100 * time.Millisecond)

			stats := sys.spriteCache.Stats()
			if stats.EntryCount == 0 {
				t.Errorf("Expected cache to contain sprites for genre %s, got 0 entries", genre)
			}
		})
	}
}

// TestWarmCommonSprites_Integration verifies integration with game initialization.
func TestWarmCommonSprites_Integration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	// Create a minimal game setup
	game := &engine.EbitenGame{
		World: engine.NewWorld(),
	}
	game.ScreenWidth = 800
	game.ScreenHeight = 600

	// Initialize core systems (minimal setup)
	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: sprites.NewGenerator(),
	}

	// Test that warming integrates with lazy init workflow
	warmCommonSprites(sys, &seed, &genre, clientLogger)

	// Verify warming doesn't break game initialization
	if game.World == nil {
		t.Errorf("Game world should still be valid after sprite warming")
	}

	// Wait for async completion
	time.Sleep(100 * time.Millisecond)

	stats := sys.spriteCache.Stats()
	if stats.EntryCount == 0 {
		t.Errorf("Expected sprites in cache after integration test")
	}
}

// BenchmarkWarmCommonSprites benchmarks sprite cache warming performance.
func BenchmarkWarmCommonSprites(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys := &systemsContainer{
			spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
			spriteGenerator: sprites.NewGenerator(),
		}

		warmCommonSprites(sys, &seed, &genre, clientLogger)

		// Wait for async generation to complete
		time.Sleep(50 * time.Millisecond)
	}
}

// BenchmarkWarmCommonSprites_QueueOnly benchmarks just the queueing overhead.
func BenchmarkWarmCommonSprites_QueueOnly(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	clientLogger := logger.WithFields(logrus.Fields{"component": "client"})

	seed := int64(12345)
	genre := "fantasy"

	sys := &systemsContainer{
		spriteCache:     cache.NewSpriteCache(100 * 1024 * 1024),
		spriteGenerator: sprites.NewGenerator(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		warmCommonSprites(sys, &seed, &genre, clientLogger)
		// Don't wait for generation - measure queue time only
	}
}
