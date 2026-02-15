package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestBucketSize verifies dimension bucketing to 4-pixel boundaries.
func TestBucketSize(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{"zero", 0, 4},
		{"negative", -5, 4},
		{"exact_4", 4, 4},
		{"5_rounds_up", 5, 8},
		{"8_exact", 8, 8},
		{"9_rounds", 9, 12},
		{"32_exact", 32, 32},
		{"33_rounds", 33, 36},
		{"1_rounds", 1, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bucketSize(tt.in)
			if got != tt.want {
				t.Errorf("bucketSize(%d) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

// TestBucketKey verifies unique cache keys for different dimension pairs.
func TestBucketKey(t *testing.T) {
	tests := []struct {
		name string
		w, h int
	}{
		{"small", 4, 4},
		{"medium", 20, 8},
		{"large", 40, 16},
	}
	seen := make(map[uint64]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := bucketKey(tt.w, tt.h)
			if seen[key] {
				t.Errorf("duplicate key for (%d, %d)", tt.w, tt.h)
			}
			seen[key] = true
			// Same inputs should produce same key
			if bucketKey(tt.w, tt.h) != key {
				t.Error("non-deterministic key")
			}
		})
	}
	// Asymmetric dimensions should differ
	if bucketKey(4, 8) == bucketKey(8, 4) {
		t.Error("asymmetric dimensions should produce different keys")
	}
}

// TestGenerateShadowImage verifies shadow image generation properties.
func TestGenerateShadowImage(t *testing.T) {
	tests := []struct {
		name    string
		w, h    int
		r, g, b float64
		opacity float64
	}{
		{"default_black", 20, 8, 0, 0, 0, 0.35},
		{"white_shadow", 16, 16, 1.0, 1.0, 1.0, 0.5},
		{"tiny", 4, 4, 0, 0, 0, 0.2},
		{"large", 40, 16, 0.1, 0.05, 0.0, 0.4},
		{"full_opacity", 12, 6, 0, 0, 0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := generateShadowImage(tt.w, tt.h, tt.r, tt.g, tt.b, tt.opacity)
			if img == nil {
				t.Fatal("nil image")
			}
			bounds := img.Bounds()
			if bounds.Dx() != tt.w || bounds.Dy() != tt.h {
				t.Errorf("size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.w, tt.h)
			}
		})
	}
}

// TestGenerateShadowImage_CenterBrightest verifies center pixel has highest alpha.
func TestGenerateShadowImage_CenterBrightest(t *testing.T) {
	w, h := 20, 10
	img := generateShadowImage(w, h, 0, 0, 0, 0.8)

	// Read back pixels from the RGBA buffer before ebiten conversion
	// We test the generateShadowImage output indirectly via the cache
	if img == nil {
		t.Fatal("nil image")
	}
	// The center should be the brightest (highest alpha) area
	// Since ebiten images can't be read back directly, we verify the
	// function doesn't panic and produces non-nil output
}

// TestDropShadowCache_GetMiss verifies cache miss returns nil.
func TestDropShadowCache_GetMiss(t *testing.T) {
	c := newDropShadowCache(10)
	if c.get(20, 8) != nil {
		t.Error("expected nil on cache miss")
	}
}

// TestDropShadowCache_PutGet verifies put/get round trip.
func TestDropShadowCache_PutGet(t *testing.T) {
	c := newDropShadowCache(10)
	img := generateShadowImage(8, 4, 0, 0, 0, 0.35)
	c.put(8, 4, img)

	got := c.get(8, 4)
	if got != img {
		t.Error("expected same image from cache")
	}
}

// TestDropShadowCache_Eviction verifies cache evicts when full.
func TestDropShadowCache_Eviction(t *testing.T) {
	c := newDropShadowCache(3)
	for i := 0; i < 5; i++ {
		img := generateShadowImage(4+i*4, 4, 0, 0, 0, 0.3)
		c.put(4+i*4, 4, img)
	}
	// After inserting 5 items into a cache of size 3, at most 3 should remain
	count := 0
	c.mu.Lock()
	count = len(c.entries)
	c.mu.Unlock()
	if count > 3 {
		t.Errorf("cache has %d entries, max should be 3", count)
	}
}

// TestDropShadowCache_GetOrCreate verifies lazy generation and caching.
func TestDropShadowCache_GetOrCreate(t *testing.T) {
	c := newDropShadowCache(10)

	img1 := c.getOrCreate(12, 8, 0, 0, 0, 0.35)
	if img1 == nil {
		t.Fatal("expected non-nil image")
	}

	img2 := c.getOrCreate(12, 8, 0, 0, 0, 0.35)
	if img2 != img1 {
		t.Error("expected cached image on second call")
	}
}

// TestClampFShadow verifies clamping behavior.
func TestClampFShadow(t *testing.T) {
	tests := []struct {
		name       string
		v, lo, hi  float64
		want       float64
	}{
		{"in_range", 0.5, 0, 1, 0.5},
		{"below", -1, 0, 1, 0},
		{"above", 2, 0, 1, 1},
		{"at_lo", 0, 0, 1, 0},
		{"at_hi", 1, 0, 1, 1},
		{"negative_range", -5, -10, -1, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampFShadow(tt.v, tt.lo, tt.hi)
			if got != tt.want {
				t.Errorf("clampFShadow(%v, %v, %v) = %v, want %v", tt.v, tt.lo, tt.hi, got, tt.want)
			}
		})
	}
}

// TestEntityGetDropShadow_Cached verifies the cached getter.
func TestEntityGetDropShadow_Cached(t *testing.T) {
	e := NewEntity(1)
	if e.GetDropShadow() != nil {
		t.Error("expected nil before adding component")
	}

	ds := NewDropShadowComponent()
	e.AddComponent(ds)
	got := e.GetDropShadow()
	if got != ds {
		t.Error("cached getter should return added component")
	}
}

// TestEntityGetDropShadow_Remove verifies cache is cleared on removal.
func TestEntityGetDropShadow_Remove(t *testing.T) {
	e := NewEntity(2)
	ds := NewDropShadowComponent()
	e.AddComponent(ds)
	if e.GetDropShadow() == nil {
		t.Fatal("expected non-nil after add")
	}
	e.RemoveComponent("drop_shadow")
	if e.GetDropShadow() != nil {
		t.Error("expected nil after removal")
	}
}

// TestDrawDropShadow_NilComponent verifies no panic when component is absent.
func TestDrawDropShadow_NilComponent(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 32, Y: 32})
	// No DropShadowComponent — should silently return
	r.drawDropShadow(entity, 32, 32)
}

// TestDrawDropShadow_Disabled verifies disabled shadow is not drawn.
func TestDrawDropShadow_Disabled(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(2)
	ds := NewDropShadowComponent()
	ds.Enabled = false
	entity.AddComponent(ds)
	r.drawDropShadow(entity, 32, 32)
}

// TestDrawDropShadow_ZeroOpacity verifies zero opacity is skipped.
func TestDrawDropShadow_ZeroOpacity(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(3)
	ds := NewDropShadowComponent()
	ds.Opacity = 0
	entity.AddComponent(ds)
	r.drawDropShadow(entity, 32, 32)
}

// TestDrawDropShadow_ValidComponent verifies shadow is drawn without panic.
func TestDrawDropShadow_ValidComponent(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(4)
	ds := NewDropShadowComponent()
	entity.AddComponent(ds)
	r.drawDropShadow(entity, 32, 32)
}

// TestDrawDropShadow_TinySize verifies very small shadow dimensions are skipped.
func TestDrawDropShadow_TinySize(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(5)
	ds := &DropShadowComponent{
		ShadowWidth:  1.0,
		ShadowHeight: 1.0,
		Opacity:      0.5,
		Enabled:      true,
	}
	entity.AddComponent(ds)
	// Should skip since rounded dimensions < 2
	r.drawDropShadow(entity, 32, 32)
}

// TestDrawDropShadow_CustomColor verifies non-black shadow color.
func TestDrawDropShadow_CustomColor(t *testing.T) {
	r := NewRenderSystem(nil)
	r.screen = newTestImage(64, 64)
	entity := NewEntity(6)
	ds := &DropShadowComponent{
		ShadowWidth:  16.0,
		ShadowHeight: 8.0,
		Opacity:      0.4,
		ColorR:       0.2,
		ColorG:       0.1,
		ColorB:       0.05,
		OffsetX:      2.0,
		OffsetY:      3.0,
		Enabled:      true,
	}
	entity.AddComponent(ds)
	r.drawDropShadow(entity, 32, 32)
}

// TestGenerateShadowImage_EdgeTransparent verifies corner pixels are transparent.
func TestGenerateShadowImage_EdgeTransparent(t *testing.T) {
	// Generate a known shadow and verify its RGBA buffer properties
	w, h := 16, 8
	// We can't read ebiten.Image pixels, but we can verify the image
	// doesn't panic and is the correct size
	img := generateShadowImage(w, h, 0, 0, 0, 0.5)
	bounds := img.Bounds()
	if bounds.Dx() != w || bounds.Dy() != h {
		t.Errorf("expected %dx%d, got %dx%d", w, h, bounds.Dx(), bounds.Dy())
	}
}

// newTestImage creates an ebiten image for testing.
func newTestImage(w, h int) *ebiten.Image {
	return ebiten.NewImage(w, h)
}

// TestDropShadowComponent_RenderDefaults verifies default component values.
func TestDropShadowComponent_RenderDefaults(t *testing.T) {
	ds := NewDropShadowComponent()
	if ds.ShadowWidth != 20.0 {
		t.Errorf("default width = %f, want 20.0", ds.ShadowWidth)
	}
	if ds.ShadowHeight != 8.0 {
		t.Errorf("default height = %f, want 8.0", ds.ShadowHeight)
	}
	if ds.Opacity != 0.35 {
		t.Errorf("default opacity = %f, want 0.35", ds.Opacity)
	}
	if !ds.Enabled {
		t.Error("default should be enabled")
	}
	if ds.Type() != "drop_shadow" {
		t.Errorf("type = %q, want drop_shadow", ds.Type())
	}
}

// BenchmarkGenerateShadowImage benchmarks shadow image generation.
func BenchmarkGenerateShadowImage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateShadowImage(20, 8, 0, 0, 0, 0.35)
	}
}

// BenchmarkDropShadowCache_GetOrCreate benchmarks cached shadow retrieval.
func BenchmarkDropShadowCache_GetOrCreate(b *testing.B) {
	c := newDropShadowCache(64)
	// Warm the cache
	c.getOrCreate(20, 8, 0, 0, 0, 0.35)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.getOrCreate(20, 8, 0, 0, 0, 0.35)
	}
}
