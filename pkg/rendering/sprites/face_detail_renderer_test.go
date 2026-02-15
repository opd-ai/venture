package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestComputeFaceParams(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		direction Direction
	}{
		{"seed_1_down", 1, DirDown},
		{"seed_2_left", 2, DirLeft},
		{"seed_42_right", 42, DirRight},
		{"seed_99_up", 99, DirUp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headSpec := PartSpec{
				RelativeX:      0.5,
				RelativeY:      0.175,
				RelativeWidth:  0.45,
				RelativeHeight: 0.35,
			}
			params := ComputeFaceParams(32, 32, headSpec, tt.direction, tt.seed)
			if params.EyeSize < 1 || params.EyeSize > 3 {
				t.Errorf("EyeSize %d out of range [1,3]", params.EyeSize)
			}
			if params.EyeColor.A != 255 {
				t.Errorf("EyeColor alpha should be 255, got %d", params.EyeColor.A)
			}
			if params.Direction != tt.direction {
				t.Errorf("Direction mismatch: got %v want %v", params.Direction, tt.direction)
			}
		})
	}
}

func TestComputeFaceParamsDeterministic(t *testing.T) {
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	p1 := ComputeFaceParams(32, 32, headSpec, DirDown, 12345)
	p2 := ComputeFaceParams(32, 32, headSpec, DirDown, 12345)
	if p1.EyeColor != p2.EyeColor {
		t.Error("Same seed should produce same EyeColor")
	}
	if p1.MouthColor != p2.MouthColor {
		t.Error("Same seed should produce same MouthColor")
	}
}

func TestComputeFaceParamsVariety(t *testing.T) {
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	seen := make(map[color.RGBA]bool)
	for seed := int64(0); seed < 50; seed++ {
		p := ComputeFaceParams(32, 32, headSpec, DirDown, seed)
		seen[p.EyeColor] = true
	}
	if len(seen) < 5 {
		t.Errorf("Expected at least 5 distinct eye colors from 50 seeds, got %d", len(seen))
	}
}

func TestComputeFaceParamsFromComponent(t *testing.T) {
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	params := ComputeFaceParamsFromComponent(32, 32, headSpec, DirDown, 1,
		0.85, 0.65, 0.20, // eye RGB
		0.75, 0.45, 0.40, // mouth RGB
		2.0, 1.0, "hostile")
	if params.Expression != "hostile" {
		t.Errorf("Expected hostile expression, got %s", params.Expression)
	}
	if params.EyeSize != 2 {
		t.Errorf("Expected eye size 2, got %d", params.EyeSize)
	}
	if params.EyeColor.R != 216 { // 0.85*255 ≈ 216
		t.Errorf("Expected eye R ~216, got %d", params.EyeColor.R)
	}
}

func TestComputeFaceParamsFromComponentClamping(t *testing.T) {
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	params := ComputeFaceParamsFromComponent(32, 32, headSpec, DirDown, 1,
		1.5, -0.5, 0.5, 0.5, 0.5, 0.5,
		5.0, 0.0, "neutral")
	if params.EyeColor.R != 255 {
		t.Errorf("Expected clamped R=255, got %d", params.EyeColor.R)
	}
	if params.EyeColor.G != 0 {
		t.Errorf("Expected clamped G=0, got %d", params.EyeColor.G)
	}
	if params.EyeSize != 3 {
		t.Errorf("Expected clamped eye size 3, got %d", params.EyeSize)
	}
	if params.MouthSize != 1 {
		t.Errorf("Expected clamped mouth size 1, got %d", params.MouthSize)
	}
}

func TestRenderFaceDetailDown(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	headSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.175,
		RelativeWidth:  0.45,
		RelativeHeight: 0.35,
	}
	params := FaceRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec:     headSpec,
		EyeColor:     color.RGBA{R: 80, G: 50, B: 25, A: 255},
		EyeSize:      2,
		MouthColor:   color.RGBA{R: 160, G: 90, B: 90, A: 255},
		MouthSize:    1,
		Expression:   "neutral",
		Direction:    DirDown,
		Seed:         42,
	}
	RenderFaceDetail(img, params)

	// Verify at least some non-zero pixels were drawn in the head area
	nonZero := 0
	sprH := float64(32)
	headTop := int(sprH*0.175) - int(sprH*0.35)/2
	headBot := headTop + int(sprH*0.35)
	for y := headTop; y < headBot; y++ {
		for x := 0; x < 32; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 || a > 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Error("Expected non-zero pixels in head area for DirDown face")
	}
}

func TestRenderFaceDetailUpNoFeatures(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	params := FaceRenderParams{
		SpriteWidth: 32, SpriteHeight: 32, HeadSpec: headSpec,
		EyeColor: color.RGBA{R: 80, G: 50, B: 25, A: 255}, EyeSize: 2,
		Direction: DirUp, Seed: 42,
	}
	RenderFaceDetail(img, params)

	// Should draw nothing for DirUp
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r > 0 || g > 0 || b > 0 || a > 0 {
				t.Fatal("Expected no pixels drawn for DirUp face (back of head)")
			}
		}
	}
}

func TestRenderFaceDetailLeftRight(t *testing.T) {
	for _, dir := range []Direction{DirLeft, DirRight} {
		t.Run(string(dir), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
			params := FaceRenderParams{
				SpriteWidth: 32, SpriteHeight: 32, HeadSpec: headSpec,
				EyeColor: color.RGBA{R: 80, G: 50, B: 25, A: 255}, EyeSize: 2,
				MouthColor: color.RGBA{R: 160, G: 90, B: 90, A: 255}, MouthSize: 1,
				Direction: dir, Seed: 42, Expression: "neutral",
			}
			RenderFaceDetail(img, params)

			nonZero := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("Expected non-zero pixels for %s face", dir)
			}
		})
	}
}

func TestRenderFaceDetailExpressions(t *testing.T) {
	expressions := []string{"neutral", "hostile", "friendly", "scared"}
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.175, RelativeWidth: 0.45, RelativeHeight: 0.35}
	for _, expr := range expressions {
		t.Run(expr, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			params := FaceRenderParams{
				SpriteWidth: 32, SpriteHeight: 32, HeadSpec: headSpec,
				EyeColor: color.RGBA{R: 80, G: 50, B: 25, A: 255}, EyeSize: 2,
				MouthColor: color.RGBA{R: 160, G: 90, B: 90, A: 255}, MouthSize: 1,
				Expression: expr, Direction: DirDown, Seed: 42,
			}
			RenderFaceDetail(img, params)
			// Just verify no panic and something was drawn
			nonZero := 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("Expected pixels drawn for expression %s", expr)
			}
		})
	}
}

func TestRenderFaceDetailTinyHead(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	headSpec := PartSpec{RelativeX: 0.5, RelativeY: 0.5, RelativeWidth: 0.3, RelativeHeight: 0.3}
	params := FaceRenderParams{
		SpriteWidth: 8, SpriteHeight: 8, HeadSpec: headSpec,
		EyeColor: color.RGBA{R: 80, G: 50, B: 25, A: 255}, EyeSize: 2,
		Direction: DirDown, Seed: 1,
	}
	// Should not panic even with very small head
	RenderFaceDetail(img, params)
}

func TestBrightenRGBA(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	bright := brightenRGBA(c, 1.5)
	if bright.R != 150 {
		t.Errorf("Expected R=150, got %d", bright.R)
	}
	// Test clamping
	bright2 := brightenRGBA(c, 3.0)
	if bright2.R != 255 {
		t.Errorf("Expected clamped R=255, got %d", bright2.R)
	}
}
