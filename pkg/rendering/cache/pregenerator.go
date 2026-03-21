package cache

import (
	"context"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	log "github.com/sirupsen/logrus"
)

// GeneratorFunc represents a function that generates a sprite image.
type GeneratorFunc func() (*ebiten.Image, error)

// PreGenerator handles batch pre-generation of sprites to warm the cache.
// Phase 44: Pre-generates common sprites to improve cache hit rate.
type PreGenerator struct {
	mu sync.RWMutex

	cache *SpriteCache

	// Pre-generation queue
	queue []PreGenRequest

	// Statistics
	stats PreGenStats
}

// PreGenRequest represents a sprite pre-generation request.
type PreGenRequest struct {
	Key       CacheKey
	Generator GeneratorFunc
}

// PreGenStats tracks pre-generation metrics.
type PreGenStats struct {
	RequestsQueued   int
	RequestsComplete int
	RequestsFailed   int
	CacheHits        int // Requests that were already cached
}

// NewPreGenerator creates a new pre-generator.
func NewPreGenerator(cache *SpriteCache) *PreGenerator {
	return &PreGenerator{
		cache: cache,
		queue: make([]PreGenRequest, 0, 100),
	}
}

// Queue adds a sprite generation request to the queue.
func (p *PreGenerator) Queue(key CacheKey, generator GeneratorFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already in cache
	if p.cache.Contains(key) {
		p.stats.CacheHits++
		return
	}

	p.queue = append(p.queue, PreGenRequest{
		Key:       key,
		Generator: generator,
	})
	p.stats.RequestsQueued++
}

// QueueBatch adds multiple generation requests.
func (p *PreGenerator) QueueBatch(requests []PreGenRequest) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, req := range requests {
		// Check if already in cache
		if p.cache.Contains(req.Key) {
			p.stats.CacheHits++
			continue
		}

		p.queue = append(p.queue, req)
		p.stats.RequestsQueued++
	}
}

// processRequest generates a single sprite and updates stats.
// Returns true if sprite was generated, false if skipped or failed.
func (p *PreGenerator) processRequest(req PreGenRequest) bool {
	// Double-check cache (another goroutine might have generated it)
	if p.cache.Contains(req.Key) {
		p.mu.Lock()
		p.stats.CacheHits++
		p.mu.Unlock()
		return false
	}

	// Generate sprite
	img, err := req.Generator()
	if err != nil {
		p.mu.Lock()
		p.stats.RequestsFailed++
		failCount := p.stats.RequestsFailed
		p.mu.Unlock()
		log.WithFields(log.Fields{
			"system_name":  "pregenerator",
			"cache_key":    string(req.Key),
			"failed_count": failCount,
			"error":        err.Error(),
		}).Warn("sprite pre-generation failed")
		return false
	}

	// Store in cache
	p.cache.Put(req.Key, img)

	p.mu.Lock()
	p.stats.RequestsComplete++
	p.mu.Unlock()

	return true
}

// Generate processes all queued requests.
// Returns number of sprites generated.
func (p *PreGenerator) Generate() int {
	p.mu.Lock()
	queue := p.queue
	p.queue = make([]PreGenRequest, 0, 100) // Reset queue
	p.mu.Unlock()

	generated := 0
	for _, req := range queue {
		if p.processRequest(req) {
			generated++
		}
	}
	return generated
}

// GenerateAsync processes queued requests in background with cancellation support.
// Returns immediately; generation happens in a goroutine.
// The provided ctx can be used to cancel the operation early.
func (p *PreGenerator) GenerateAsync(ctx context.Context, doneCh chan<- int) {
	go func() {
		count := p.GenerateWithContext(ctx)
		if doneCh != nil {
			doneCh <- count
		}
	}()
}

// GenerateWithContext processes all queued requests, respecting ctx cancellation.
// Returns number of sprites generated before ctx was cancelled or queue was empty.
func (p *PreGenerator) GenerateWithContext(ctx context.Context) int {
	p.mu.Lock()
	queue := p.queue
	p.queue = make([]PreGenRequest, 0, 100)
	p.mu.Unlock()

	generated := 0
	for _, req := range queue {
		select {
		case <-ctx.Done():
			return generated
		default:
		}

		if p.processRequest(req) {
			generated++
		}
	}
	return generated
}

// QueueSize returns number of pending requests.
func (p *PreGenerator) QueueSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.queue)
}

// Stats returns pre-generation statistics.
func (p *PreGenerator) Stats() PreGenStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// Clear removes all queued requests.
func (p *PreGenerator) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.queue = make([]PreGenRequest, 0, 100)
}

// HitRate returns cache hit rate for pre-generation requests.
func (p *PreGenerator) HitRate() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := p.stats.RequestsQueued
	if total == 0 {
		return 0.0
	}

	return float64(p.stats.CacheHits) / float64(total)
}

// SuccessRate returns successful generation rate.
func (p *PreGenerator) SuccessRate() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	attempted := p.stats.RequestsComplete + p.stats.RequestsFailed
	if attempted == 0 {
		return 1.0
	}

	return float64(p.stats.RequestsComplete) / float64(attempted)
}
