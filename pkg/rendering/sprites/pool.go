package sprites

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// ImagePool provides object pooling for Ebiten images to reduce GC pressure.
// Images are expensive to allocate due to GPU texture creation overhead.
// Pooling allows reuse of images instead of creating new ones each frame.
type ImagePool struct {
	pool   sync.Pool
	width  int
	height int
}

// NewImagePool creates a new image pool for images of a specific size.
func NewImagePool(width, height int) *ImagePool {
	return &ImagePool{
		pool: sync.Pool{
			New: func() interface{} {
				return ebiten.NewImage(width, height)
			},
		},
		width:  width,
		height: height,
	}
}

// Get retrieves an image from the pool or creates a new one.
// The returned image will be cleared (transparent).
func (p *ImagePool) Get() *ebiten.Image {
	img := p.pool.Get().(*ebiten.Image)
	// Clear the image for reuse
	img.Clear()
	return img
}

// Put returns an image to the pool for future reuse.
// The image should not be used after being returned to the pool.
func (p *ImagePool) Put(img *ebiten.Image) {
	if img == nil {
		return
	}

	// Verify image size matches pool
	bounds := img.Bounds()
	if bounds.Dx() != p.width || bounds.Dy() != p.height {
		// Don't pool wrong-sized images
		return
	}

	p.pool.Put(img)
}

// ShapePool provides pooling for temporary shape images during generation.
// Sprite generation creates many intermediate images for layers/parts.
// Pooling these reduces allocation overhead significantly.
type ShapePool struct {
	pools map[int]*ImagePool
	mutex sync.RWMutex
}

// NewShapePool creates a new shape pool supporting multiple image sizes.
func NewShapePool() *ShapePool {
	return &ShapePool{
		pools: make(map[int]*ImagePool),
	}
}

// Get retrieves an image of the specified size from the appropriate pool.
func (sp *ShapePool) Get(width, height int) *ebiten.Image {
	key := width*10000 + height // Simple key encoding

	sp.mutex.RLock()
	pool, exists := sp.pools[key]
	sp.mutex.RUnlock()

	if !exists {
		sp.mutex.Lock()
		// Double-check after acquiring write lock
		pool, exists = sp.pools[key]
		if !exists {
			pool = NewImagePool(width, height)
			sp.pools[key] = pool
		}
		sp.mutex.Unlock()
	}

	return pool.Get()
}

// Put returns an image to the appropriate pool.
func (sp *ShapePool) Put(img *ebiten.Image) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	key := width*10000 + height

	sp.mutex.RLock()
	pool, exists := sp.pools[key]
	sp.mutex.RUnlock()

	if exists {
		pool.Put(img)
	}
	// If pool doesn't exist, just discard the image
}

// Clear removes all pooled images and resets the pool.
// Useful for memory cleanup during loading screens or scene transitions.
func (sp *ShapePool) Clear() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.pools = make(map[int]*ImagePool)
}

// PooledGenerator wraps a Generator with image pooling for performance.
type PooledGenerator struct {
	generator *Generator
	shapePool *ShapePool
	enabled   bool
}

// NewPooledGenerator creates a generator with image pooling enabled.
func NewPooledGenerator() *PooledGenerator {
	return &PooledGenerator{
		generator: NewGenerator(),
		shapePool: NewShapePool(),
		enabled:   true,
	}
}

// Generate generates a sprite using pooled images for intermediate steps.
func (pg *PooledGenerator) Generate(config Config) (*ebiten.Image, error) {
	// For now, delegate to standard generator
	// Pooling will be used internally when generator is refactored
	// to use the pool for temporary images
	return pg.generator.Generate(config)
}

// SetPoolingEnabled enables or disables image pooling.
func (pg *PooledGenerator) SetPoolingEnabled(enabled bool) {
	pg.enabled = enabled
}

// IsPoolingEnabled returns whether pooling is currently enabled.
func (pg *PooledGenerator) IsPoolingEnabled() bool {
	return pg.enabled
}

// ClearPool clears all pooled images.
func (pg *PooledGenerator) ClearPool() {
	pg.shapePool.Clear()
}

// GetPool returns the underlying shape pool for direct access.
func (pg *PooledGenerator) GetPool() *ShapePool {
	return pg.shapePool
}

// CombinedGenerator combines caching and pooling for maximum performance.
type CombinedGenerator struct {
	generator    *Generator
	cache        *Cache
	shapePool    *ShapePool
	cacheEnabled bool
	poolEnabled  bool
}

// NewCombinedGenerator creates a generator with both caching and pooling.
func NewCombinedGenerator(cacheCapacity int) *CombinedGenerator {
	return &CombinedGenerator{
		generator:    NewGenerator(),
		cache:        NewCache(cacheCapacity),
		shapePool:    NewShapePool(),
		cacheEnabled: true,
		poolEnabled:  true,
	}
}

// Generate generates a sprite using both caching and pooling.
func (cg *CombinedGenerator) Generate(config Config) (*ebiten.Image, error) {
	// Try cache first if enabled
	if cg.cacheEnabled {
		if sprite := cg.cache.Get(config); sprite != nil {
			return sprite, nil
		}
	}

	// Cache miss - generate new sprite
	// (pooling will be used internally when generator is refactored)
	sprite, err := cg.generator.Generate(config)
	if err != nil {
		return nil, err
	}

	// Store in cache if enabled
	if cg.cacheEnabled {
		cg.cache.Put(config, sprite)
	}

	return sprite, nil
}

// SetCacheEnabled enables or disables caching.
func (cg *CombinedGenerator) SetCacheEnabled(enabled bool) {
	cg.cacheEnabled = enabled
}

// SetPoolingEnabled enables or disables pooling.
func (cg *CombinedGenerator) SetPoolingEnabled(enabled bool) {
	cg.poolEnabled = enabled
}

// IsCacheEnabled returns whether caching is currently enabled.
func (cg *CombinedGenerator) IsCacheEnabled() bool {
	return cg.cacheEnabled
}

// IsPoolingEnabled returns whether pooling is currently enabled.
func (cg *CombinedGenerator) IsPoolingEnabled() bool {
	return cg.poolEnabled
}

// ClearCache clears the sprite cache.
func (cg *CombinedGenerator) ClearCache() {
	cg.cache.Clear()
}

// ClearPool clears all pooled images.
func (cg *CombinedGenerator) ClearPool() {
	cg.shapePool.Clear()
}

// Stats returns cache performance statistics.
func (cg *CombinedGenerator) Stats() CacheStats {
	return cg.cache.Stats()
}

// GetPool returns the underlying shape pool.
func (cg *CombinedGenerator) GetPool() *ShapePool {
	return cg.shapePool
}

// GetCache returns the underlying cache.
func (cg *CombinedGenerator) GetCache() *Cache {
	return cg.cache
}

// BatchGenerate generates multiple sprites using caching and pooling.
func (cg *CombinedGenerator) BatchGenerate(batchConfig BatchConfig) ([]*ebiten.Image, error) {
	if len(batchConfig.Configs) == 0 {
		return nil, nil
	}

	results := make([]*ebiten.Image, len(batchConfig.Configs))

	if !batchConfig.Concurrent || batchConfig.MaxWorkers <= 1 {
		// Sequential generation
		for i, config := range batchConfig.Configs {
			sprite, err := cg.Generate(config)
			if err != nil {
				if batchConfig.OnError != nil {
					batchConfig.OnError(i, err)
				}
				continue
			}
			results[i] = sprite

			if batchConfig.OnProgress != nil {
				batchConfig.OnProgress(i+1, len(batchConfig.Configs))
			}
		}
		return results, nil
	}

	// Parallel generation
	workers := batchConfig.MaxWorkers
	if workers <= 0 {
		workers = 4
	}

	type job struct {
		index  int
		config Config
	}

	type result struct {
		index  int
		sprite *ebiten.Image
		err    error
	}

	jobs := make(chan job, len(batchConfig.Configs))
	resultsChan := make(chan result, len(batchConfig.Configs))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				sprite, err := cg.Generate(job.config)
				resultsChan <- result{
					index:  job.index,
					sprite: sprite,
					err:    err,
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for i, config := range batchConfig.Configs {
			jobs <- job{index: i, config: config}
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	completed := 0
	for res := range resultsChan {
		if res.err != nil {
			if batchConfig.OnError != nil {
				batchConfig.OnError(res.index, res.err)
			}
			continue
		}
		results[res.index] = res.sprite
		completed++

		if batchConfig.OnProgress != nil {
			batchConfig.OnProgress(completed, len(batchConfig.Configs))
		}
	}

	return results, nil
}
