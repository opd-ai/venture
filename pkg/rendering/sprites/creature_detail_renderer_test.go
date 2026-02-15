package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestRenderCreatureDetails_AllForms(t *testing.T) {
	forms := []struct {
		name string
		form string
	}{
		{"quadruped", "quadruped"},
		{"arachnid", "arachnid"},
		{"serpentine", "serpentine"},
		{"flying", "flying"},
		{"blob", "blob"},
		{"mechanical", "mechanical"},
		{"undead", "undead"},
	}

	for _, tt := range forms {
		t.Run(tt.name, func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := CreatureDetailParams{
				Width:     32,
				Height:    32,
				Form:      tt.form,
				Direction: "down",
				Seed:      42,
				SizeClass: "medium",
				Genre:     "fantasy",
			}
			RenderCreatureDetails(buf, params)

			// Verify at least some non-transparent pixels were drawn
			nonZero := countNonZeroPixels(buf)
			if nonZero == 0 {
				t.Errorf("form %q produced no visible pixels", tt.form)
			}
		})
	}
}

func TestRenderCreatureDetails_Directions(t *testing.T) {
	directions := []string{"up", "down", "left", "right"}
	for _, dir := range directions {
		t.Run(dir, func(t *testing.T) {
			buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := CreatureDetailParams{
				Width: 32, Height: 32,
				Form:      "quadruped",
				Direction: dir,
				Seed:      123,
				SizeClass: "medium",
			}
			RenderCreatureDetails(buf, params)
			if countNonZeroPixels(buf) == 0 {
				t.Errorf("direction %q produced no pixels", dir)
			}
		})
	}
}

func TestRenderCreatureDetails_DeterministicOutput(t *testing.T) {
	forms := []string{"quadruped", "arachnid", "serpentine", "flying", "blob", "mechanical", "undead"}
	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			params := CreatureDetailParams{
				Width: 32, Height: 32,
				Form:      form,
				Direction: "down",
				Seed:      999,
				SizeClass: "medium",
				Genre:     "fantasy",
			}

			buf1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			buf2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			RenderCreatureDetails(buf1, params)
			RenderCreatureDetails(buf2, params)

			if !rgbaEqual(buf1, buf2) {
				t.Errorf("form %q is not deterministic with same seed", form)
			}
		})
	}
}

func TestRenderCreatureDetails_DifferentSeedsDiffer(t *testing.T) {
	forms := []string{"arachnid", "blob", "undead"}
	for _, form := range forms {
		t.Run(form, func(t *testing.T) {
			params1 := CreatureDetailParams{
				Width: 32, Height: 32,
				Form: form, Direction: "down",
				Seed: 1, SizeClass: "medium",
			}
			params2 := params1
			params2.Seed = 9999

			buf1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			buf2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
			RenderCreatureDetails(buf1, params1)
			RenderCreatureDetails(buf2, params2)

			if rgbaEqual(buf1, buf2) {
				t.Errorf("form %q produced identical output for different seeds", form)
			}
		})
	}
}

func TestRenderCreatureDetails_NilBuffer(t *testing.T) {
	// Should not panic
	RenderCreatureDetails(nil, CreatureDetailParams{
		Width: 32, Height: 32, Form: "blob",
		Direction: "down", Seed: 1,
	})
}

func TestRenderCreatureDetails_ZeroDimensions(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderCreatureDetails(buf, CreatureDetailParams{
		Width: 0, Height: 0, Form: "blob",
		Direction: "down", Seed: 1,
	})
	// Should not panic, should not draw anything
	if countNonZeroPixels(buf) != 0 {
		t.Error("zero dimensions should produce no pixels")
	}
}

func TestRenderCreatureDetails_HumanoidSkipped(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderCreatureDetails(buf, CreatureDetailParams{
		Width: 32, Height: 32, Form: "humanoid",
		Direction: "down", Seed: 1,
	})
	if countNonZeroPixels(buf) != 0 {
		t.Error("humanoid form should produce no creature details")
	}
}

func TestRenderCreatureDetails_UnknownFormSkipped(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderCreatureDetails(buf, CreatureDetailParams{
		Width: 32, Height: 32, Form: "unknown_creature",
		Direction: "down", Seed: 1,
	})
	if countNonZeroPixels(buf) != 0 {
		t.Error("unknown form should produce no creature details")
	}
}

func TestRenderCreatureDetails_GenreVariation(t *testing.T) {
	// Undead has genre-specific eye colors (horror vs default)
	params := CreatureDetailParams{
		Width: 32, Height: 32, Form: "undead",
		Direction: "down", Seed: 77, SizeClass: "medium",
	}

	buf1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	params.Genre = "fantasy"
	RenderCreatureDetails(buf1, params)

	buf2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	params.Genre = "horror"
	RenderCreatureDetails(buf2, params)

	if rgbaEqual(buf1, buf2) {
		t.Error("undead should look different in horror vs fantasy genre")
	}
}

func TestRenderCreatureDetails_LargerSprite(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 64, 64))
	params := CreatureDetailParams{
		Width: 64, Height: 64, Form: "flying",
		Direction: "down", Seed: 42, SizeClass: "large",
	}
	RenderCreatureDetails(buf, params)
	if countNonZeroPixels(buf) == 0 {
		t.Error("64x64 flying creature should produce visible pixels")
	}
}

func TestIsOpaqueAt(t *testing.T) {
	buf := image.NewRGBA(image.Rect(0, 0, 4, 4))
	buf.SetRGBA(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 128})

	if isOpaqueAt(buf, 0, 0) {
		t.Error("transparent pixel should not be opaque")
	}
	if !isOpaqueAt(buf, 1, 1) {
		t.Error("pixel with alpha > 0 should be opaque")
	}
	if isOpaqueAt(buf, -1, -1) {
		t.Error("out-of-bounds should not be opaque")
	}
}

func TestClampU8(t *testing.T) {
	tests := []struct {
		in  float64
		out uint8
	}{
		{-10, 0},
		{0, 0},
		{128.5, 128},
		{255, 255},
		{300, 255},
	}
	for _, tt := range tests {
		got := clampU8(tt.in)
		if got != tt.out {
			t.Errorf("clampU8(%v) = %d, want %d", tt.in, got, tt.out)
		}
	}
}

func TestHeadOffset(t *testing.T) {
	tests := []struct {
		dir    string
		wantDx int // relative to cx
		wantDy int // relative to cy
	}{
		{"up", 0, -1},
		{"down", 0, 1},
		{"left", -1, 0},
		{"right", 1, 0},
	}
	cx, cy := 16, 16
	for _, tt := range tests {
		hx, hy := headOffset(cx, cy, tt.dir, 32, 32)
		dx := sign(hx - cx)
		dy := sign(hy - cy)
		if dx != tt.wantDx || dy != tt.wantDy {
			t.Errorf("headOffset dir=%s: got offset (%d,%d), want sign (%d,%d)", tt.dir, dx, dy, tt.wantDx, tt.wantDy)
		}
	}
}

func TestTailOffset(t *testing.T) {
	// Tail should be opposite to head
	cx, cy := 16, 16
	hx, _ := headOffset(cx, cy, "up", 32, 32)
	tx, _ := tailOffset(cx, cy, "up", 32, 32)
	if hx != cx || tx != cx {
		// x should stay centered for up/down
	}
	_, hy := headOffset(cx, cy, "up", 32, 32)
	_, ty := tailOffset(cx, cy, "up", 32, 32)
	if hy >= ty {
		t.Error("head should be above tail when facing up")
	}
}

// --- helpers ---

func countNonZeroPixels(buf *image.RGBA) int {
	count := 0
	for y := buf.Rect.Min.Y; y < buf.Rect.Max.Y; y++ {
		for x := buf.Rect.Min.X; x < buf.Rect.Max.X; x++ {
			c := buf.RGBAAt(x, y)
			if c.A > 0 {
				count++
			}
		}
	}
	return count
}

func rgbaEqual(a, b *image.RGBA) bool {
	if a.Rect != b.Rect {
		return false
	}
	for i := range a.Pix {
		if a.Pix[i] != b.Pix[i] {
			return false
		}
	}
	return true
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
