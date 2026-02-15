package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestHairStyleString(t *testing.T) {
	tests := []struct {
		style HairStyle
		want  string
	}{
		{HairBald, "bald"},
		{HairBuzz, "buzz"},
		{HairShort, "short"},
		{HairMedium, "medium"},
		{HairLong, "long"},
		{HairPonytail, "ponytail"},
		{HairMohawk, "mohawk"},
		{HairSpiky, "spiky"},
		{HairTopKnot, "topknot"},
		{HairHooded, "hooded"},
		{HairBraided, "braided"},
		{HairStyle(-1), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.style.String(); got != tt.want {
				t.Errorf("HairStyle(%d).String() = %q, want %q", tt.style, got, tt.want)
			}
		})
	}
}

func TestRenderHairOverlay_AllStyles(t *testing.T) {
	baseColor := color.RGBA{R: 140, G: 100, B: 60, A: 255}
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for style := HairBald; style < HairStyleCount; style++ {
		for _, dir := range directions {
			t.Run(style.String()+"_"+string(dir), func(t *testing.T) {
				img := image.NewRGBA(image.Rect(0, 0, 32, 32))
				params := HairRenderParams{
					HeadCenterX: 16,
					HeadCenterY: 8,
					HeadRadius:  5,
					Style:       style,
					Color:       baseColor,
					Direction:   dir,
					Seed:        42,
				}
				RenderHairOverlay(img, params)

				// Verify at least some pixels were drawn
				nonTransparent := countNonTransparent(img)
				if nonTransparent == 0 {
					t.Errorf("style %s dir %s produced zero hair pixels", style, dir)
				}
			})
		}
	}
}

func TestRenderHairOverlay_Deterministic(t *testing.T) {
	params := HairRenderParams{
		HeadCenterX: 16,
		HeadCenterY: 8,
		HeadRadius:  5,
		Style:       HairShort,
		Color:       color.RGBA{R: 100, G: 70, B: 40, A: 255},
		Direction:   DirDown,
		Seed:        12345,
	}

	img1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderHairOverlay(img1, params)

	img2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	RenderHairOverlay(img2, params)

	// Pixel-perfect match
	for i := range img1.Pix {
		if img1.Pix[i] != img2.Pix[i] {
			t.Fatal("same params produced different output — not deterministic")
		}
	}
}

func TestRenderHairOverlay_DifferentSeeds(t *testing.T) {
	base := HairRenderParams{
		HeadCenterX: 16,
		HeadCenterY: 8,
		HeadRadius:  5,
		Style:       HairMedium,
		Color:       color.RGBA{R: 100, G: 70, B: 40, A: 255},
		Direction:   DirDown,
	}

	img1 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	p1 := base
	p1.Seed = 111
	RenderHairOverlay(img1, p1)

	img2 := image.NewRGBA(image.Rect(0, 0, 32, 32))
	p2 := base
	p2.Seed = 222
	RenderHairOverlay(img2, p2)

	// Should differ (different seeds)
	differ := false
	for i := range img1.Pix {
		if img1.Pix[i] != img2.Pix[i] {
			differ = true
			break
		}
	}
	if !differ {
		t.Error("different seeds produced identical output")
	}
}

func TestRenderHairOverlay_ZeroRadius(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	params := HairRenderParams{
		HeadCenterX: 16,
		HeadCenterY: 8,
		HeadRadius:  0,
		Style:       HairShort,
		Color:       color.RGBA{R: 100, G: 70, B: 40, A: 255},
		Direction:   DirDown,
		Seed:        42,
	}
	RenderHairOverlay(img, params)
	// Should not panic and produce no pixels
	if countNonTransparent(img) != 0 {
		t.Error("zero radius should produce no pixels")
	}
}

func TestRenderHairOverlay_BoundsRespected(t *testing.T) {
	// Head near edge of small image — must not panic
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for style := HairBald; style < HairStyleCount; style++ {
		params := HairRenderParams{
			HeadCenterX: 2,
			HeadCenterY: 2,
			HeadRadius:  4,
			Style:       style,
			Color:       color.RGBA{R: 200, G: 150, B: 100, A: 255},
			Direction:   DirDown,
			Seed:        99,
		}
		RenderHairOverlay(img, params) // Must not panic
	}
}

func TestComputeHairRenderParams(t *testing.T) {
	traits := &AvatarTraits{
		HairColor: color.RGBA{R: 100, G: 70, B: 40, A: 255},
		HairStyle: HairPonytail,
	}
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.2,
		RelativeWidth:  0.45,
		RelativeHeight: 0.35,
	}
	p := ComputeHairRenderParams(32, 32, headSpec, traits, DirDown, 42)

	if p.HeadCenterX != 16 {
		t.Errorf("HeadCenterX = %d, want 16", p.HeadCenterX)
	}
	if p.HeadCenterY != 6 {
		t.Errorf("HeadCenterY = %d, want 6", p.HeadCenterY)
	}
	if p.HeadRadius < 1 {
		t.Errorf("HeadRadius = %d, want >= 1", p.HeadRadius)
	}
	if p.Style != HairPonytail {
		t.Errorf("Style = %v, want HairPonytail", p.Style)
	}
	if p.Color != traits.HairColor {
		t.Error("Color mismatch")
	}
}

func TestComputeHairRenderParams_PixelDimensions(t *testing.T) {
	traits := &AvatarTraits{
		HairColor: color.RGBA{R: 100, G: 70, B: 40, A: 255},
		HairStyle: HairShort,
	}
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.2,
		RelativeWidth:  0.179,
		RelativeHeight: 0.179,
		PreferredPixelSize: &PixelDimensions{
			Width:  5,
			Height: 5,
		},
	}
	p := ComputeHairRenderParams(28, 28, headSpec, traits, DirUp, 100)
	if p.HeadRadius != 2 {
		t.Errorf("HeadRadius = %d, want 2 (from 5/2)", p.HeadRadius)
	}
}

func TestDirectionOffsets(t *testing.T) {
	tests := []struct {
		dir    Direction
		backX  float64
		backY  float64
		frontX float64
		frontY float64
	}{
		{DirUp, 0, 1, 0, -1},
		{DirDown, 0, -1, 0, 1},
		{DirLeft, 1, 0, -1, 0},
		{DirRight, -1, 0, 1, 0},
	}
	for _, tt := range tests {
		t.Run(string(tt.dir), func(t *testing.T) {
			bx, by := directionBackOffset(tt.dir)
			if bx != tt.backX || by != tt.backY {
				t.Errorf("back(%s) = (%v,%v), want (%v,%v)", tt.dir, bx, by, tt.backX, tt.backY)
			}
			fx, fy := directionFrontOffset(tt.dir)
			if fx != tt.frontX || fy != tt.frontY {
				t.Errorf("front(%s) = (%v,%v), want (%v,%v)", tt.dir, fx, fy, tt.frontX, tt.frontY)
			}
		})
	}
}

func TestSetPixelSafe_OutOfBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Should not panic
	setPixelSafe(img, -1, -1, color.RGBA{R: 255, A: 255})
	setPixelSafe(img, 4, 4, color.RGBA{R: 255, A: 255})
	setPixelSafe(img, 100, 100, color.RGBA{R: 255, A: 255})

	// Valid pixel
	setPixelSafe(img, 2, 2, color.RGBA{R: 255, G: 128, B: 64, A: 255})
	r, g, b, a := img.At(2, 2).RGBA()
	if r>>8 != 255 || g>>8 != 128 || b>>8 != 64 || a>>8 != 255 {
		t.Errorf("pixel at (2,2) wrong: r=%d g=%d b=%d a=%d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestBlendPixelSafe(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// Set base pixel
	setPixelSafe(img, 1, 1, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	// Blend semi-transparent overlay
	blendPixelSafe(img, 1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 128})

	r, g, _, _ := img.At(1, 1).RGBA()
	// R should be reduced (darkened by ~50% blend of black)
	if r>>8 > 150 {
		t.Errorf("blended pixel R = %d, expected darkened", r>>8)
	}
	if g>>8 > 80 {
		t.Errorf("blended pixel G = %d, expected darkened", g>>8)
	}
}

func TestShadeColor(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	darker := shadeColor(c, 0.5)
	if darker.R != 50 || darker.G != 50 || darker.B != 50 {
		t.Errorf("shade(0.5) = (%d,%d,%d), want (50,50,50)", darker.R, darker.G, darker.B)
	}
	lighter := shadeColor(c, 1.5)
	if lighter.R != 150 || lighter.G != 150 || lighter.B != 150 {
		t.Errorf("shade(1.5) = (%d,%d,%d), want (150,150,150)", lighter.R, lighter.G, lighter.B)
	}
	// Clamp at 255
	maxed := shadeColor(color.RGBA{R: 200, G: 200, B: 200, A: 255}, 2.0)
	if maxed.R != 255 {
		t.Errorf("shade(2.0) R = %d, want clamped 255", maxed.R)
	}
}

func TestGenerateAvatarTraits_HasHairStyle(t *testing.T) {
	seeds := []int64{0, 1, 42, 12345, 999999}
	for _, seed := range seeds {
		traits := GenerateAvatarTraits(seed)
		if traits.HairStyle < 0 || traits.HairStyle >= HairStyleCount {
			t.Errorf("seed %d: HairStyle = %d, want [0, %d)", seed, traits.HairStyle, HairStyleCount)
		}
	}
}

func TestGenerateAvatarTraits_HairStyleVariety(t *testing.T) {
	seen := make(map[HairStyle]bool)
	for seed := int64(0); seed < 200; seed++ {
		traits := GenerateAvatarTraits(seed)
		seen[traits.HairStyle] = true
	}
	// With 11 styles and 200 seeds, we should see at least 6 distinct styles
	if len(seen) < 6 {
		t.Errorf("only %d distinct hair styles in 200 seeds, want >= 6", len(seen))
	}
}

func TestLightenRGBA(t *testing.T) {
	c := color.RGBA{R: 100, G: 80, B: 60, A: 255}
	l := lightenRGBA(c, 1.5)
	if l.R != 150 || l.G != 120 || l.B != 90 {
		t.Errorf("lighten(1.5) = (%d,%d,%d), want (150,120,90)", l.R, l.G, l.B)
	}
}

func TestMaxInt(t *testing.T) {
	if maxInt(3, 5) != 5 {
		t.Error("maxInt(3,5) != 5")
	}
	if maxInt(5, 3) != 5 {
		t.Error("maxInt(5,3) != 5")
	}
	if maxInt(3, 3) != 3 {
		t.Error("maxInt(3,3) != 3")
	}
}

// countNonTransparent counts pixels with alpha > 0.
func countNonTransparent(img *image.RGBA) int {
	count := 0
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			if img.Pix[idx+3] > 0 {
				count++
			}
		}
	}
	return count
}
