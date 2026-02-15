package sprites

import (
	"image"
	"image/color"
	"testing"
)

// makeTestSprite creates a small test sprite with a filled circle.
func makeTestSprite(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	cx, cy := w/2, h/2
	r := w / 3
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				img.SetRGBA(x, y, color.RGBA{R: 180, G: 120, B: 80, A: 255})
			}
		}
	}
	return img
}

func TestFinalizeEntitySprite_NilInput(t *testing.T) {
	result := FinalizeEntitySprite(nil, DefaultFinalizerConfig(42))
	if result == nil {
		t.Fatal("expected non-nil result for nil input")
	}
}

func TestFinalizeEntitySprite_EmptyImage(t *testing.T) {
	empty := image.NewRGBA(image.Rect(0, 0, 0, 0))
	result := FinalizeEntitySprite(empty, DefaultFinalizerConfig(42))
	if result == nil {
		t.Fatal("expected non-nil result for empty image")
	}
}

func TestFinalizeEntitySprite_ProducesOutline(t *testing.T) {
	src := makeTestSprite(16, 16)
	cfg := DefaultFinalizerConfig(12345)
	result := FinalizeEntitySprite(src, cfg)

	// Count opaque pixels in source vs result — result should have more due to outline
	srcOpaque := countOpaquePixels(src)
	resOpaque := countOpaquePixels(result)

	if resOpaque <= srcOpaque {
		t.Errorf("expected more opaque pixels after outline; src=%d result=%d", srcOpaque, resOpaque)
	}
}

func TestFinalizeEntitySprite_Deterministic(t *testing.T) {
	src := makeTestSprite(16, 16)
	cfg := DefaultFinalizerConfig(9999)

	r1 := FinalizeEntitySprite(src, cfg)
	r2 := FinalizeEntitySprite(src, cfg)

	if !imagesEqual(r1, r2) {
		t.Error("expected deterministic output for same seed")
	}
}

func TestFinalizeEntitySprite_DifferentSeeds(t *testing.T) {
	src := makeTestSprite(16, 16)
	r1 := FinalizeEntitySprite(src, DefaultFinalizerConfig(1))
	r2 := FinalizeEntitySprite(src, DefaultFinalizerConfig(2))

	// Different seeds should produce different per-pixel variation
	if imagesEqual(r1, r2) {
		t.Error("expected different output for different seeds")
	}
}

func TestFinalizeEntitySprite_OutlineOnly(t *testing.T) {
	src := makeTestSprite(16, 16)
	cfg := FinalizerConfig{
		EnableOutline:       true,
		OutlineDarkenFactor: 0.4,
		Seed:                42,
	}
	result := FinalizeEntitySprite(src, cfg)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Source pixels should be preserved
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			sc := src.RGBAAt(x, y)
			if sc.A > 0 {
				rc := result.RGBAAt(x, y)
				if rc.A == 0 {
					t.Fatalf("opaque source pixel at (%d,%d) was erased", x, y)
				}
			}
		}
	}
}

func TestFinalizeEntitySprite_RimLightBrightensTopEdge(t *testing.T) {
	src := makeTestSprite(32, 32)
	cfgNoRim := FinalizerConfig{
		EnableOutline: false,
		EnableRimLight: false, Seed: 42,
	}
	cfgWithRim := FinalizerConfig{
		EnableOutline: false,
		EnableRimLight: true, RimLightIntensity: 0.5, RimLightAngle: 315,
		Seed: 42,
	}

	noRim := FinalizeEntitySprite(src, cfgNoRim)
	withRim := FinalizeEntitySprite(src, cfgWithRim)

	// Sum brightness of entire image — rim version should be at least slightly brighter
	noRimBright := regionBrightness(noRim, 0, 0, 32, 32)
	withRimBright := regionBrightness(withRim, 0, 0, 32, 32)

	if withRimBright <= noRimBright {
		t.Errorf("rim lighting should increase total brightness; without=%d with=%d", noRimBright, withRimBright)
	}
}

func TestFinalizeEntitySprite_EdgeShadowDarkensBottom(t *testing.T) {
	src := makeTestSprite(32, 32)
	cfgNoShadow := FinalizerConfig{
		EnableOutline:    false,
		EnableEdgeShadow: false, Seed: 42,
	}
	cfgWithShadow := FinalizerConfig{
		EnableOutline:       false,
		EnableEdgeShadow:    true,
		EdgeShadowIntensity: 0.5,
		RimLightAngle:       315,
		Seed:                42,
	}

	noShadow := FinalizeEntitySprite(src, cfgNoShadow)
	withShadow := FinalizeEntitySprite(src, cfgWithShadow)

	// Sum brightness of entire image — shadow version should be darker overall
	noShadowBright := regionBrightness(noShadow, 0, 0, 32, 32)
	withShadowBright := regionBrightness(withShadow, 0, 0, 32, 32)

	if withShadowBright >= noShadowBright {
		t.Errorf("edge shadow should reduce total brightness; without=%d with=%d", noShadowBright, withShadowBright)
	}
}

func TestDefaultFinalizerConfig(t *testing.T) {
	cfg := DefaultFinalizerConfig(42)
	if !cfg.EnableOutline {
		t.Error("expected outline enabled by default")
	}
	if !cfg.EnableRimLight {
		t.Error("expected rim light enabled by default")
	}
	if !cfg.EnableEdgeShadow {
		t.Error("expected edge shadow enabled by default")
	}
	if cfg.OutlineDarkenFactor <= 0 || cfg.OutlineDarkenFactor >= 1 {
		t.Errorf("outline darken factor out of range: %f", cfg.OutlineDarkenFactor)
	}
}

func TestComputeDominantEdgeColor_AllTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	opaque := make([]bool, 64)
	c := computeDominantEdgeColor(img, opaque, 8, 8)
	// Should return fallback dark color
	if c.R > 40 || c.G > 40 || c.B > 40 {
		t.Errorf("expected dark fallback color, got %v", c)
	}
}

func TestIsEdgePixel(t *testing.T) {
	opaque := []bool{
		false, false, false,
		false, true, false,
		false, false, false,
	}
	// Single opaque pixel at center (1,1) is always an edge pixel
	if !isEdgePixel(opaque, 1, 1, 3, 3) {
		t.Error("center pixel should be an edge pixel")
	}
}

func TestHasOpaqueNeighbor(t *testing.T) {
	opaque := []bool{
		false, true, false,
		false, false, false,
		false, false, false,
	}
	// (0,0) has opaque neighbor at (1,0)
	if !hasOpaqueNeighbor(opaque, 0, 0, 3, 3) {
		t.Error("(0,0) should have opaque neighbor")
	}
	// (2,2) has no opaque neighbor
	if hasOpaqueNeighbor(opaque, 2, 2, 3, 3) {
		t.Error("(2,2) should not have opaque neighbor")
	}
}

func TestEdgeNormal(t *testing.T) {
	// 3x3 grid with center opaque, all others transparent
	opaque := []bool{
		false, false, false,
		false, true, false,
		false, false, false,
	}
	nx, ny := edgeNormal(opaque, 1, 1, 3, 3)
	// Normal magnitude should be ~1
	length := nx*nx + ny*ny
	if length < 0.5 || length > 1.5 {
		t.Errorf("expected unit-ish normal, got (%f,%f) length^2=%f", nx, ny, length)
	}
}

func TestDarkenColor(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	d := darkenColor(c, 0.5)
	if d.R != 100 || d.G != 50 || d.B != 25 {
		t.Errorf("expected (100,50,25), got (%d,%d,%d)", d.R, d.G, d.B)
	}
}

func TestClampU8_Float64(t *testing.T) {
	tests := []struct {
		in   float64
		want uint8
	}{
		{-10, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}
	for _, tt := range tests {
		got := clampU8(tt.in)
		if got != tt.want {
			t.Errorf("clampU8(%f) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFinalizeEntitySprite_32x32(b *testing.B) {
	src := makeTestSprite(32, 32)
	cfg := DefaultFinalizerConfig(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalizeEntitySprite(src, cfg)
	}
}

func BenchmarkFinalizeEntitySprite_64x64(b *testing.B) {
	src := makeTestSprite(64, 64)
	cfg := DefaultFinalizerConfig(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalizeEntitySprite(src, cfg)
	}
}

// --- Helpers ---

func countOpaquePixels(img *image.RGBA) int {
	bounds := img.Bounds()
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.RGBAAt(x, y).A > 0 {
				count++
			}
		}
	}
	return count
}

func imagesEqual(a, b *image.RGBA) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.RGBAAt(x, y) != b.RGBAAt(x, y) {
				return false
			}
		}
	}
	return true
}

func regionBrightness(img *image.RGBA, x0, y0, x1, y1 int) int {
	total := 0
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := img.RGBAAt(x, y)
			total += int(c.R) + int(c.G) + int(c.B)
		}
	}
	return total
}
