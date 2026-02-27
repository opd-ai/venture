package performance

import (
	"container/list"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/recovery"
)

// CacheManager handles resource caching with LRU eviction
type CacheManager struct {
	mu        sync.RWMutex
	maxSizeMB uint64
	entries   map[string]*CacheEntry
	lruList   *list.List
	lruMap    map[string]*list.Element
	stats     *CacheStats
	hitCount  uint64 // Total cache hits
	missCount uint64 // Total cache misses
}

// NewCacheManager creates a new cache manager with the specified maximum size.
// maxSizeBytes must be at least 1MB (1048576 bytes); values below this will be
// rounded up to 1MB to ensure the cache can store at least one entry.
//
// Zero or very small maxSizeBytes values would cause immediate eviction of all
// entries, making the cache non-functional. This minimum prevents such behavior.
func NewCacheManager(maxSizeBytes uint64) *CacheManager {
	maxSizeMB := maxSizeBytes / (1024 * 1024)
	if maxSizeMB == 0 {
		maxSizeMB = 1
		log.WithFields(log.Fields{
			"requested_bytes": maxSizeBytes,
			"enforced_mb":     maxSizeMB,
		}).Warn("cache size too small, enforcing 1MB minimum")
	}

	return &CacheManager{
		maxSizeMB: maxSizeMB,
		entries:   make(map[string]*CacheEntry),
		lruList:   list.New(),
		lruMap:    make(map[string]*list.Element),
		stats: &CacheStats{
			MaxSizeMB: maxSizeMB,
		},
	}
}

// Set adds or updates a cache entry
func (cm *CacheManager) Set(key string, data interface{}, sizeBytes uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove existing entry if present
	if entry, exists := cm.entries[key]; exists {
		cm.removeEntry(entry)
	}

	// Create new entry
	entry := &CacheEntry{
		Key:          key,
		Data:         data,
		SizeBytes:    sizeBytes,
		LastAccessed: time.Now(),
		AccessCount:  1,
	}

	cm.entries[key] = entry

	// Add to LRU
	element := cm.lruList.PushFront(key)
	cm.lruMap[key] = element

	// Update stats
	cm.stats.ItemCount = len(cm.entries)
	cm.updateCurrentSize()

	// Evict if necessary
	cm.evictIfNeeded()
}

// Get retrieves a cache entry
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry, exists := cm.entries[key]
	if !exists {
		cm.missCount++
		return nil, false
	}

	// Track cache hit
	cm.hitCount++

	// Update access stats
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	// Move to front of LRU
	if element, found := cm.lruMap[key]; found {
		cm.lruList.MoveToFront(element)
	}

	return entry.Data, true
}

// Remove deletes a cache entry
func (cm *CacheManager) Remove(key string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if entry, exists := cm.entries[key]; exists {
		cm.removeEntry(entry)
		cm.updateCurrentSize()
	}
}

// removeEntry removes an entry (must be called with lock held)
func (cm *CacheManager) removeEntry(entry *CacheEntry) {
	delete(cm.entries, entry.Key)

	if element, found := cm.lruMap[entry.Key]; found {
		cm.lruList.Remove(element)
		delete(cm.lruMap, entry.Key)
	}

	cm.stats.ItemCount = len(cm.entries)
}

// evictIfNeeded evicts LRU entries if over size limit
func (cm *CacheManager) evictIfNeeded() {
	for cm.stats.CurrentSizeMB > cm.stats.MaxSizeMB && cm.lruList.Len() > 0 {
		// Evict least recently used
		element := cm.lruList.Back()
		if element == nil {
			break
		}

		key := element.Value.(string)
		if entry, exists := cm.entries[key]; exists {
			log.WithFields(log.Fields{
				"key":             key,
				"size_bytes":      entry.SizeBytes,
				"current_size_mb": cm.stats.CurrentSizeMB,
				"max_size_mb":     cm.stats.MaxSizeMB,
			}).Debug("cache evicting LRU entry")
			cm.removeEntry(entry)
			cm.stats.EvictionCount++
		}

		cm.updateCurrentSize()
	}
}

// updateCurrentSize recalculates current cache size
func (cm *CacheManager) updateCurrentSize() {
	var totalBytes uint64
	for _, entry := range cm.entries {
		totalBytes += entry.SizeBytes
	}
	cm.stats.CurrentSizeMB = totalBytes / (1024 * 1024)
}

// Cleanup forces eviction of old entries
func (cm *CacheManager) Cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.evictIfNeeded()
	cm.stats.LastCleanup = time.Now()
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := *cm.stats

	// Calculate actual hit rate from tracked hits/misses
	totalRequests := cm.hitCount + cm.missCount
	if totalRequests > 0 {
		stats.HitRate = float64(cm.hitCount) / float64(totalRequests)
	} else {
		stats.HitRate = 0.0 // No requests yet
	}

	return &stats
}

// Clear removes all entries
func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.entries = make(map[string]*CacheEntry)
	cm.lruList = list.New()
	cm.lruMap = make(map[string]*list.Element)
	cm.stats.ItemCount = 0
	cm.stats.CurrentSizeMB = 0
	cm.hitCount = 0
	cm.missCount = 0
}

// ResourceLoader defines the interface for loading resources
type ResourceLoader interface {
	// Load loads the resource identified by the request and returns the loaded data
	Load(request *LoadRequest) (interface{}, error)
}

// DefaultResourceLoader provides a no-op implementation for testing
type DefaultResourceLoader struct{}

// Load returns nil data (no-op implementation)
func (d *DefaultResourceLoader) Load(request *LoadRequest) (interface{}, error) {
	// Default implementation returns nil - actual resource loading
	// is provided by injecting a real ResourceLoader implementation
	return nil, nil
}

// BackgroundLoader preloads resources asynchronously
type BackgroundLoader struct {
	mu       sync.RWMutex
	queue    []*LoadRequest
	workers  int
	workChan chan *LoadRequest
	stopChan chan struct{}
	running  bool
	loader   ResourceLoader // Injected resource loader
}

// NewBackgroundLoader creates a new background loader
func NewBackgroundLoader(workers int) *BackgroundLoader {
	return &BackgroundLoader{
		queue:    make([]*LoadRequest, 0),
		workers:  workers,
		workChan: make(chan *LoadRequest, 100),
		stopChan: make(chan struct{}),
		loader:   &DefaultResourceLoader{},
	}
}

// NewBackgroundLoaderWithLoader creates a new background loader with a custom resource loader
func NewBackgroundLoaderWithLoader(workers int, loader ResourceLoader) *BackgroundLoader {
	if loader == nil {
		loader = &DefaultResourceLoader{}
	}
	return &BackgroundLoader{
		queue:    make([]*LoadRequest, 0),
		workers:  workers,
		workChan: make(chan *LoadRequest, 100),
		stopChan: make(chan struct{}),
		loader:   loader,
	}
}

// Start begins background loading
func (bl *BackgroundLoader) Start() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if bl.running {
		return
	}

	bl.running = true

	for i := 0; i < bl.workers; i++ {
		go bl.worker()
	}
}

// Stop halts background loading
func (bl *BackgroundLoader) Stop() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if !bl.running {
		return
	}

	bl.running = false
	close(bl.stopChan)
}

// PreloadRaid queues a raid for background loading
func (bl *BackgroundLoader) PreloadRaid(raidID string, callback func(interface{})) {
	request := &LoadRequest{
		ID:       raidID,
		Type:     "raid",
		Priority: 5,
		Callback: callback,
	}
	bl.Queue(request)
}

// PreloadGuildHall queues a guild hall for background loading
func (bl *BackgroundLoader) PreloadGuildHall(hallID string, callback func(interface{})) {
	request := &LoadRequest{
		ID:       hallID,
		Type:     "guild_hall",
		Priority: 3,
		Callback: callback,
	}
	bl.Queue(request)
}

// Queue adds a load request
func (bl *BackgroundLoader) Queue(request *LoadRequest) {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if !bl.running {
		return
	}

	bl.queue = append(bl.queue, request)

	// Send to workers (non-blocking)
	select {
	case bl.workChan <- request:
	default:
		// Channel full, will process from queue later
	}
}

// worker processes load requests
func (bl *BackgroundLoader) worker() {
	defer recovery.RecoverPanicWithLogger("background_loader", "worker goroutine", nil)()

	for {
		select {
		case request := <-bl.workChan:
			// Load resource using the injected loader
			data, err := bl.loader.Load(request)
			if err != nil {
				log.WithFields(log.Fields{
					"request_id":   request.ID,
					"request_type": request.Type,
					"error":        err.Error(),
				}).Warn("background loader failed to load resource")
			} else if data != nil {
				request.Data = data
			}

			if request.Callback != nil {
				request.Callback(request.Data)
			}

		case <-bl.stopChan:
			return
		}
	}
}

// GetQueueSize returns current queue length
func (bl *BackgroundLoader) GetQueueSize() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return len(bl.queue)
}

// LODManager manages level-of-detail for rendering
type LODManager struct {
	mu             sync.RWMutex
	distanceHigh   float64
	distanceMedium float64
	distanceLow    float64
	enabled        bool
}

// NewLODManager creates a new LOD manager
func NewLODManager() *LODManager {
	config := DefaultPerformanceConfig()
	return &LODManager{
		distanceHigh:   config.LODDistanceHigh,
		distanceMedium: config.LODDistanceMedium,
		distanceLow:    config.LODDistanceLow,
		enabled:        true,
	}
}

// GetLODLevel returns the appropriate LOD level for a distance
func (lm *LODManager) GetLODLevel(distance float64) LODLevel {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if !lm.enabled {
		return LODVeryHigh
	}

	if distance < lm.distanceHigh {
		return LODHigh
	} else if distance < lm.distanceMedium {
		return LODMedium
	} else if distance < lm.distanceLow {
		return LODLow
	}
	return LODVeryLow
}

// SetDistances updates LOD distance thresholds
func (lm *LODManager) SetDistances(high, medium, low float64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.distanceHigh = high
	lm.distanceMedium = medium
	lm.distanceLow = low
}

// Enable enables LOD system
func (lm *LODManager) Enable() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.enabled = true
}

// Disable disables LOD system (everything uses highest quality)
func (lm *LODManager) Disable() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.enabled = false
}

// IsEnabled returns LOD enabled state
func (lm *LODManager) IsEnabled() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.enabled
}
