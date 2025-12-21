package pool

import (
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

// Common sprite sizes used in the game.
// Phase 45: Default sprite/tile size is now 64×64 (previously 32×32).
// Memory per sprite: 32×32=4KB, 64×64=16KB, 128×128=64KB (RGBA).
const (
	SizePlayer  = 28          // Player sprite size (fixed, legacy)
	SizeSmall   = 32          // Small sprites (particles, icons)
	SizeDefault = 64          // Default sprites (entities, tiles, objects) - Phase 45 standard
	SizeMedium  = SizeDefault // Medium sprites (alias for SizeDefault for compatibility)
	SizeLarge   = 128         // Large sprites (bosses, effects)
)

// ImagePool manages pools of Ebiten images by size.
type ImagePool struct {
	// Pools for common sizes
	pool28  sync.Pool
	pool32  sync.Pool
	pool64  sync.Pool
	pool128 sync.Pool

	// Statistics
	gets    uint64
	puts    uint64
	creates uint64
}

// globalPool is the default image pool instance.
var globalPool = NewImagePool()

// NewImagePool creates a new image pool with pre-configured size pools.
func NewImagePool() *ImagePool {
	p := &ImagePool{}

	// Initialize pools with constructors
	p.pool28.New = func() interface{} {
		atomic.AddUint64(&p.creates, 1)
		return ebiten.NewImage(SizePlayer, SizePlayer)
	}
	p.pool32.New = func() interface{} {
		atomic.AddUint64(&p.creates, 1)
		return ebiten.NewImage(SizeSmall, SizeSmall)
	}
	p.pool64.New = func() interface{} {
		atomic.AddUint64(&p.creates, 1)
		return ebiten.NewImage(SizeMedium, SizeMedium)
	}
	p.pool128.New = func() interface{} {
		atomic.AddUint64(&p.creates, 1)
		return ebiten.NewImage(SizeLarge, SizeLarge)
	}

	return p
}

// GetImage retrieves an image from the appropriate pool.
// Returns a pooled image for standard sizes (28, 32, 64, 128),
// or creates a new image for non-standard sizes.
func (p *ImagePool) GetImage(width, height int) *ebiten.Image {
	atomic.AddUint64(&p.gets, 1)

	// Use pooled images for square sprites of common sizes
	if width == height {
		switch width {
		case SizePlayer:
			return p.pool28.Get().(*ebiten.Image)
		case SizeSmall:
			return p.pool32.Get().(*ebiten.Image)
		case SizeMedium:
			return p.pool64.Get().(*ebiten.Image)
		case SizeLarge:
			return p.pool128.Get().(*ebiten.Image)
		}
	}

	// Non-standard size: create new image (not pooled)
	atomic.AddUint64(&p.creates, 1)
	return ebiten.NewImage(width, height)
}

// PutImage returns an image to the appropriate pool.
// The image is cleared before being returned to the pool.
// Only standard-sized square images are actually pooled.
func (p *ImagePool) PutImage(img *ebiten.Image) {
	if img == nil {
		return
	}

	atomic.AddUint64(&p.puts, 1)

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Clear the image before returning to pool
	img.Clear()

	// Only pool square images of standard sizes
	if width == height {
		switch width {
		case SizePlayer:
			p.pool28.Put(img)
			return
		case SizeSmall:
			p.pool32.Put(img)
			return
		case SizeMedium:
			p.pool64.Put(img)
			return
		case SizeLarge:
			p.pool128.Put(img)
			return
		}
	}

	// Non-standard size: let it be garbage collected
}

// Statistics holds pool usage statistics.
type Statistics struct {
	Gets    uint64 // Number of Get calls
	Puts    uint64 // Number of Put calls
	Creates uint64 // Number of new allocations
}

// Stats returns a copy of the current pool statistics.
func (p *ImagePool) Stats() Statistics {
	return Statistics{
		Gets:    atomic.LoadUint64(&p.gets),
		Puts:    atomic.LoadUint64(&p.puts),
		Creates: atomic.LoadUint64(&p.creates),
	}
}

// ReuseRate returns the percentage of Get calls that were served from pool (0.0 to 1.0).
func (s Statistics) ReuseRate() float64 {
	if s.Gets == 0 {
		return 0.0
	}
	reused := s.Gets - s.Creates
	return float64(reused) / float64(s.Gets)
}

// Global pool functions for convenience

// GetImage retrieves an image from the global pool.
func GetImage(width, height int) *ebiten.Image {
	return globalPool.GetImage(width, height)
}

// PutImage returns an image to the global pool.
func PutImage(img *ebiten.Image) {
	globalPool.PutImage(img)
}

// Stats returns statistics from the global pool.
func Stats() Statistics {
	return globalPool.Stats()
}

// ResetStats resets the global pool statistics.
func ResetStats() {
	atomic.StoreUint64(&globalPool.gets, 0)
	atomic.StoreUint64(&globalPool.puts, 0)
	atomic.StoreUint64(&globalPool.creates, 0)
}
