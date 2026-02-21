// Package sprites provides tests for creature eye pattern rendering.
package sprites

import (
	"image"
	"image/color"
	"testing"
)

func TestRenderCreatureEyes(t *testing.T) {
	// Create a test image
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	// Create test params for a spider (8 eyes)
	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.3,
			RelativeWidth:  0.35,
			RelativeHeight: 0.3,
		},
		EyePattern: "arachnid_8",
		EyeCount:   8,
		EyePositions: []float64{
			0.35, 0.25, // Eye 1
			0.65, 0.25, // Eye 2
			0.25, 0.30, // Eye 3
			0.75, 0.30, // Eye 4
			0.30, 0.40, // Eye 5
			0.70, 0.40, // Eye 6
			0.20, 0.45, // Eye 7
			0.80, 0.45, // Eye 8
		},
		EyeSizes:      []float64{1.2, 1.2, 1.0, 1.0, 0.7, 0.7, 0.5, 0.5},
		EyeR:          0.3,
		EyeG:          0.2,
		EyeB:          0.1,
		PupilStyle:    "none",
		GlowIntensity: 0.0,
		Direction:     DirDown,
		Seed:          12345,
	}

	// Should not panic
	RenderCreatureEyes(img, params)

	// Check that some pixels were drawn (non-black)
	hasContent := false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a > 0 && (r > 0 || g > 0 || b > 0) {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Expected some pixels to be drawn")
	}
}

func TestRenderCreatureEyesSlitPupil(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.3,
			RelativeWidth:  0.4,
			RelativeHeight: 0.35,
		},
		EyePattern: "serpent_slit",
		EyeCount:   2,
		EyePositions: []float64{
			0.30, 0.40,
			0.70, 0.40,
		},
		EyeSizes:      []float64{1.3, 1.3},
		EyeR:          0.8,
		EyeG:          0.6,
		EyeB:          0.1,
		PupilStyle:    "slit_vertical",
		GlowIntensity: 0.0,
		Direction:     DirDown,
	}

	RenderCreatureEyes(img, params)

	// Should have drawn something
	hasContent := false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Expected slit pupil eyes to be drawn")
	}
}

func TestRenderCreatureEyesFaceted(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.25,
			RelativeWidth:  0.45,
			RelativeHeight: 0.35,
		},
		EyePattern: "insect_compound",
		EyeCount:   2,
		EyePositions: []float64{
			0.22, 0.40,
			0.78, 0.40,
		},
		EyeSizes:      []float64{1.8, 1.8},
		EyeR:          0.4,
		EyeG:          0.5,
		EyeB:          0.3,
		PupilStyle:    "faceted",
		GlowIntensity: 0.0,
		Direction:     DirDown,
	}

	RenderCreatureEyes(img, params)

	// Verify content was drawn
	hasContent := false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Expected compound faceted eyes to be drawn")
	}
}

func TestRenderCreatureEyesMechanicalGlow(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.3,
			RelativeWidth:  0.35,
			RelativeHeight: 0.3,
		},
		EyePattern: "mechanical_sensors",
		EyeCount:   3,
		EyePositions: []float64{
			0.50, 0.30,
			0.30, 0.50,
			0.70, 0.50,
		},
		EyeSizes:      []float64{1.5, 0.8, 0.8},
		EyeR:          0.2,
		EyeG:          0.8,
		EyeB:          0.9,
		PupilStyle:    "none",
		GlowIntensity: 0.6, // Mechanical glow
		Direction:     DirDown,
	}

	RenderCreatureEyes(img, params)

	// Verify content was drawn
	hasContent := false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}

	if !hasContent {
		t.Error("Expected mechanical sensor eyes with glow to be drawn")
	}
}

func TestRenderCreatureEyesDirectionAdjustment(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))

			params := CreatureEyeRenderParams{
				SpriteWidth:  32,
				SpriteHeight: 32,
				HeadSpec: PartSpec{
					RelativeX:      0.5,
					RelativeY:      0.3,
					RelativeWidth:  0.4,
					RelativeHeight: 0.35,
				},
				EyePattern: "quadruped_2",
				EyeCount:   2,
				EyePositions: []float64{
					0.30, 0.35,
					0.70, 0.35,
				},
				EyeSizes:      []float64{1.0, 1.0},
				EyeR:          0.4,
				EyeG:          0.3,
				EyeB:          0.2,
				PupilStyle:    "round",
				GlowIntensity: 0.0,
				Direction:     dir,
			}

			// Should not panic for any direction
			RenderCreatureEyes(img, params)
		})
	}
}

func TestRenderCreatureEyesEmptyParams(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	// Empty params should not panic
	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		EyeCount:     0,
	}

	RenderCreatureEyes(img, params) // Should not panic
}

func TestRenderCreatureEyesTooSmallHead(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	// Very small head should not panic
	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.5,
			RelativeWidth:  0.05, // Very small
			RelativeHeight: 0.05,
		},
		EyeCount: 2,
		EyePositions: []float64{
			0.3, 0.5,
			0.7, 0.5,
		},
		EyeSizes:   []float64{1.0, 1.0},
		PupilStyle: "round",
	}

	RenderCreatureEyes(img, params) // Should not panic
}

func TestAdjustEyePositionForDirection(t *testing.T) {
	tests := []struct {
		dir  Direction
		relX float64
		relY float64
	}{
		{DirUp, 0.5, 0.5},
		{DirDown, 0.5, 0.5},
		{DirLeft, 0.5, 0.5},
		{DirRight, 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(string(tt.dir), func(t *testing.T) {
			adjX, adjY := adjustEyePositionForDirection(tt.relX, tt.relY, tt.dir)

			// Verify adjusted positions are within bounds
			if adjX < 0.0 || adjX > 1.0 {
				t.Errorf("Adjusted X out of bounds: %f", adjX)
			}
			if adjY < 0.0 || adjY > 1.0 {
				t.Errorf("Adjusted Y out of bounds: %f", adjY)
			}
		})
	}
}

func TestClampCEP(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{128, 0, 255, 128},
		{300, 0, 255, 255},
		{-10, 0, 255, 0},
	}

	for _, tt := range tests {
		got := clampCEP(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clampCEP(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestDarkencolorCEP(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	darkened := darkenColorCEP(c, 0.5)

	if darkened.R != 100 {
		t.Errorf("Expected R=100, got %d", darkened.R)
	}
	if darkened.G != 50 {
		t.Errorf("Expected G=50, got %d", darkened.G)
	}
	if darkened.A != 255 {
		t.Errorf("Alpha should be preserved")
	}
}

func TestBrightencolorCEP(t *testing.T) {
	c := color.RGBA{R: 100, G: 50, B: 25, A: 255}
	brightened := brightenColorCEP(c, 2.0)

	if brightened.R != 200 {
		t.Errorf("Expected R=200, got %d", brightened.R)
	}
	if brightened.G != 100 {
		t.Errorf("Expected G=100, got %d", brightened.G)
	}
	// B should clamp to 50 (25*2)
	if brightened.A != 255 {
		t.Errorf("Alpha should be preserved")
	}
}

func BenchmarkRenderCreatureEyes(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	params := CreatureEyeRenderParams{
		SpriteWidth:  32,
		SpriteHeight: 32,
		HeadSpec: PartSpec{
			RelativeX:      0.5,
			RelativeY:      0.3,
			RelativeWidth:  0.35,
			RelativeHeight: 0.3,
		},
		EyePattern: "arachnid_8",
		EyeCount:   8,
		EyePositions: []float64{
			0.35, 0.25, 0.65, 0.25, 0.25, 0.30, 0.75, 0.30,
			0.30, 0.40, 0.70, 0.40, 0.20, 0.45, 0.80, 0.45,
		},
		EyeSizes:      []float64{1.2, 1.2, 1.0, 1.0, 0.7, 0.7, 0.5, 0.5},
		EyeR:          0.3,
		EyeG:          0.2,
		EyeB:          0.1,
		PupilStyle:    "none",
		GlowIntensity: 0.0,
		Direction:     DirDown,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderCreatureEyes(img, params)
	}
}
