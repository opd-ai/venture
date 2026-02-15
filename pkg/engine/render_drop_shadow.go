// Package engine provides soft elliptical drop shadow rendering for entities.
// This file generates and caches radial-gradient shadow images keyed by
// bucketed dimensions, then draws them beneath entity sprites during the
// render pass. Drop shadows ground top-down sprites visually and provide
// depth cues that make entities look like 3D objects seen from above.
package engine

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// shadowCacheEntry holds a pre-rendered shadow image for a given size bucket.
type shadowCacheEntry struct {
	img    *ebiten.Image
	width  int
	height int
}

// dropShadowCache is a size-bucketed LRU cache for shadow images.
// Shadows are bucketed to 4-pixel increments so nearby sizes share images.
type dropShadowCache struct {
	mu      sync.Mutex
	entries map[uint64]*shadowCacheEntry
	maxSize int
}

// newDropShadowCache creates a cache with the given capacity.
func newDropShadowCache(maxSize int) *dropShadowCache {
	return &dropShadowCache{
		entries: make(map[uint64]*shadowCacheEntry, maxSize),
		maxSize: maxSize,
	}
}

// bucketKey produces a cache key from bucketed width and height.
func bucketKey(w, h int) uint64 {
	return uint64(w)<<32 | uint64(h)
}

// bucketSize rounds a dimension up to the nearest 4-pixel boundary.
func bucketSize(v int) int {
	if v <= 0 {
		return 4
	}
	return ((v + 3) / 4) * 4
}

// get retrieves a cached shadow image or nil if not present.
func (c *dropShadowCache) get(bw, bh int) *ebiten.Image {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := bucketKey(bw, bh)
	if entry, ok := c.entries[key]; ok {
		return entry.img
	}
	return nil
}

// put stores a shadow image in the cache, evicting the oldest entry if full.
func (c *dropShadowCache) put(bw, bh int, img *ebiten.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := bucketKey(bw, bh)
	if len(c.entries) >= c.maxSize {
		// Simple eviction: remove one arbitrary entry
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
	c.entries[key] = &shadowCacheEntry{img: img, width: bw, height: bh}
}

// generateShadowImage creates a soft elliptical shadow with radial gradient falloff.
// The shadow is an axis-aligned ellipse filling the given dimensions, with
// quadratic alpha falloff from center (fully opaque) to edge (transparent).
func generateShadowImage(w, h int, r, g, b float64, baseOpacity float64) *ebiten.Image {
	buf := image.NewRGBA(image.Rect(0, 0, w, h))

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0
	rx := cx // semi-axis X
	ry := cy // semi-axis Y

	if rx < 1 {
		rx = 1
	}
	if ry < 1 {
		ry = 1
	}

	cr := uint8(clampFShadow(r*255, 0, 255))
	cg := uint8(clampFShadow(g*255, 0, 255))
	cb := uint8(clampFShadow(b*255, 0, 255))

	for py := 0; py < h; py++ {
		dy := (float64(py) + 0.5 - cy) / ry
		dy2 := dy * dy
		for px := 0; px < w; px++ {
			dx := (float64(px) + 0.5 - cx) / rx
			dist2 := dx*dx + dy2
			if dist2 >= 1.0 {
				continue
			}
			// Quadratic falloff: strongest at center, zero at edge
			alpha := (1.0 - dist2) * baseOpacity
			a := uint8(clampFShadow(alpha*255, 0, 255))
			buf.SetRGBA(px, py, color.RGBA{R: cr, G: cg, B: cb, A: a})
		}
	}

	return ebiten.NewImageFromImage(buf)
}

// clampFShadow clamps a float64 to [lo, hi].
func clampFShadow(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// getOrCreateShadow returns a cached shadow image or generates and caches one.
func (c *dropShadowCache) getOrCreate(bw, bh int, r, g, b, opacity float64) *ebiten.Image {
	if img := c.get(bw, bh); img != nil {
		return img
	}
	img := generateShadowImage(bw, bh, r, g, b, opacity)
	c.put(bw, bh, img)
	return img
}

// drawDropShadow renders a soft elliptical ground shadow beneath an entity.
// It reads the entity's DropShadowComponent for size, color, opacity, and offset,
// then draws a radial-gradient ellipse at the entity's screen position.
func (r *EbitenRenderSystem) drawDropShadow(entity *Entity, screenX, screenY float64) {
	ds := entity.GetDropShadow()
	if ds == nil || !ds.Enabled || ds.Opacity <= 0 {
		return
	}

	w := int(math.Round(ds.ShadowWidth))
	h := int(math.Round(ds.ShadowHeight))
	if w < 2 || h < 2 {
		return
	}

	bw := bucketSize(w)
	bh := bucketSize(h)

	shadowImg := r.shadowCache.getOrCreate(bw, bh, ds.ColorR, ds.ColorG, ds.ColorB, ds.Opacity)

	r.shadowDrawOpts.GeoM.Reset()
	r.shadowDrawOpts.ColorScale.Reset()

	// Center the shadow image at the entity position + offset
	r.shadowDrawOpts.GeoM.Translate(
		-float64(bw)/2+ds.OffsetX,
		-float64(bh)/2+ds.OffsetY,
	)
	r.shadowDrawOpts.GeoM.Translate(screenX, screenY)
	r.screen.DrawImage(shadowImg, &r.shadowDrawOpts)
}
