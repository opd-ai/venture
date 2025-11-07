// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestNewProcessor(t *testing.T) {
	p := NewProcessor()
	if p == nil {
		t.Fatal("NewProcessor returned nil")
	}

	config := p.GetConfig()
	if config.ColorGrading.Saturation != 1.0 {
		t.Errorf("Default saturation = %v, want 1.0", config.ColorGrading.Saturation)
	}
}

func TestNewProcessorWithConfig(t *testing.T) {
	customConfig := DefaultConfig()
	customConfig.ColorGrading.Saturation = 1.5

	p := NewProcessorWithConfig(customConfig)
	if p == nil {
		t.Fatal("NewProcessorWithConfig returned nil")
	}

	config := p.GetConfig()
	if config.ColorGrading.Saturation != 1.5 {
		t.Errorf("Custom saturation = %v, want 1.5", config.ColorGrading.Saturation)
	}
}

func TestProcessor_SetConfig(t *testing.T) {
	p := NewProcessor()

	newConfig := DefaultConfig()
	newConfig.Vignette.Intensity = 0.8

	p.SetConfig(newConfig)

	config := p.GetConfig()
	if config.Vignette.Intensity != 0.8 {
		t.Errorf("Vignette intensity = %v, want 0.8", config.Vignette.Intensity)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		min   float64
		max   float64
		want  float64
	}{
		{"within range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v",
					tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestClampUint8(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  uint8
	}{
		{"normal", 128.0, 128},
		{"below zero", -10.0, 0},
		{"above 255", 300.0, 255},
		{"zero", 0.0, 0},
		{"max", 255.0, 255},
		{"fractional", 127.5, 127},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampUint8(tt.value)
			if got != tt.want {
				t.Errorf("clampUint8(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		t    float64
		want float64
	}{
		{"start", 0.0, 1.0, 0.0, 0.0},
		{"end", 0.0, 1.0, 1.0, 1.0},
		{"middle", 0.0, 1.0, 0.5, 0.5},
		{"quarter", 0.0, 1.0, 0.25, 0.25},
		{"negative", -1.0, 1.0, 0.5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lerp(tt.a, tt.b, tt.t)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("lerp(%v, %v, %v) = %v, want %v",
					tt.a, tt.b, tt.t, got, tt.want)
			}
		})
	}
}

func TestSmoothstep(t *testing.T) {
	tests := []struct {
		name         string
		edge0, edge1 float64
		x            float64
		wantRange    [2]float64 // min, max
	}{
		{"before edge0", 0.0, 1.0, -0.5, [2]float64{0.0, 0.0}},
		{"at edge0", 0.0, 1.0, 0.0, [2]float64{0.0, 0.0}},
		{"middle", 0.0, 1.0, 0.5, [2]float64{0.4, 0.6}},
		{"at edge1", 0.0, 1.0, 1.0, [2]float64{1.0, 1.0}},
		{"after edge1", 0.0, 1.0, 1.5, [2]float64{1.0, 1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := smoothstep(tt.edge0, tt.edge1, tt.x)
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("smoothstep(%v, %v, %v) = %v, want in range [%v, %v]",
					tt.edge0, tt.edge1, tt.x, got, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestRGBToHSL(t *testing.T) {
	tests := []struct {
		name                string
		r, g, b             float64
		wantH, wantS, wantL float64
		tolerance           float64
	}{
		{"red", 1.0, 0.0, 0.0, 0.0, 1.0, 0.5, 0.01},
		{"green", 0.0, 1.0, 0.0, 0.333, 1.0, 0.5, 0.01},
		{"blue", 0.0, 0.0, 1.0, 0.667, 1.0, 0.5, 0.01},
		{"white", 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.01},
		{"black", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.01},
		{"gray", 0.5, 0.5, 0.5, 0.0, 0.0, 0.5, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)

			if math.Abs(h-tt.wantH) > tt.tolerance {
				t.Errorf("rgbToHSL(%v, %v, %v) hue = %v, want %v",
					tt.r, tt.g, tt.b, h, tt.wantH)
			}
			if math.Abs(s-tt.wantS) > tt.tolerance {
				t.Errorf("rgbToHSL(%v, %v, %v) saturation = %v, want %v",
					tt.r, tt.g, tt.b, s, tt.wantS)
			}
			if math.Abs(l-tt.wantL) > tt.tolerance {
				t.Errorf("rgbToHSL(%v, %v, %v) lightness = %v, want %v",
					tt.r, tt.g, tt.b, l, tt.wantL)
			}
		})
	}
}

func TestHSLToRGB(t *testing.T) {
	tests := []struct {
		name                string
		h, s, l             float64
		wantR, wantG, wantB float64
		tolerance           float64
	}{
		{"red", 0.0, 1.0, 0.5, 1.0, 0.0, 0.0, 0.01},
		{"green", 0.333, 1.0, 0.5, 0.0, 1.0, 0.0, 0.05},
		{"blue", 0.667, 1.0, 0.5, 0.0, 0.0, 1.0, 0.05},
		{"white", 0.0, 0.0, 1.0, 1.0, 1.0, 1.0, 0.01},
		{"black", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.01},
		{"gray", 0.0, 0.0, 0.5, 0.5, 0.5, 0.5, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := hslToRGB(tt.h, tt.s, tt.l)

			if math.Abs(r-tt.wantR) > tt.tolerance {
				t.Errorf("hslToRGB(%v, %v, %v) red = %v, want %v",
					tt.h, tt.s, tt.l, r, tt.wantR)
			}
			if math.Abs(g-tt.wantG) > tt.tolerance {
				t.Errorf("hslToRGB(%v, %v, %v) green = %v, want %v",
					tt.h, tt.s, tt.l, g, tt.wantG)
			}
			if math.Abs(b-tt.wantB) > tt.tolerance {
				t.Errorf("hslToRGB(%v, %v, %v) blue = %v, want %v",
					tt.h, tt.s, tt.l, b, tt.wantB)
			}
		})
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name      string
		r, g, b   float64
		want      float64
		tolerance float64
	}{
		{"white", 1.0, 1.0, 1.0, 1.0, 0.01},
		{"black", 0.0, 0.0, 0.0, 0.0, 0.01},
		{"red", 1.0, 0.0, 0.0, 0.299, 0.01},
		{"green", 0.0, 1.0, 0.0, 0.587, 0.01},
		{"blue", 0.0, 0.0, 1.0, 0.114, 0.01},
		{"gray", 0.5, 0.5, 0.5, 0.5, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := luminance(tt.r, tt.g, tt.b)
			if math.Abs(got-tt.want) > tt.tolerance {
				t.Errorf("luminance(%v, %v, %v) = %v, want %v",
					tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestSampleBilinear(t *testing.T) {
	// Create a simple 2x2 test image
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{0, 0, 0, 255})
	img.Set(1, 0, color.RGBA{255, 0, 0, 255})
	img.Set(0, 1, color.RGBA{0, 255, 0, 255})
	img.Set(1, 1, color.RGBA{255, 255, 255, 255})

	tests := []struct {
		name string
		x, y float64
		want color.RGBA
	}{
		{"top-left", 0.0, 0.0, color.RGBA{0, 0, 0, 255}},
		{"top-right", 1.0, 0.0, color.RGBA{255, 0, 0, 255}},
		{"bottom-left", 0.0, 1.0, color.RGBA{0, 255, 0, 255}},
		{"bottom-right", 1.0, 1.0, color.RGBA{255, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sampleBilinear(img, tt.x, tt.y)
			if got.R != tt.want.R || got.G != tt.want.G || got.B != tt.want.B {
				t.Errorf("sampleBilinear(%v, %v) = %v, want %v",
					tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestBoxBlur(t *testing.T) {
	// Create a simple test image with a single white pixel in the center
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	img.Set(2, 2, color.RGBA{255, 255, 255, 255})

	t.Run("radius 1", func(t *testing.T) {
		result := boxBlur(img, 1)

		// Center pixel should be less bright after blur
		centerColor := result.RGBAAt(2, 2)
		if centerColor.R == 255 {
			t.Error("Center pixel should be blurred (less than 255)")
		}

		// Adjacent pixels should have some brightness
		adjacentColor := result.RGBAAt(2, 1)
		if adjacentColor.R == 0 {
			t.Error("Adjacent pixel should have some brightness after blur")
		}
	})

	t.Run("radius 0", func(t *testing.T) {
		result := boxBlur(img, 0)

		// Should return original image unchanged
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				orig := img.RGBAAt(x, y)
				blurred := result.RGBAAt(x, y)
				if orig != blurred {
					t.Errorf("boxBlur with radius 0 changed pixel at (%d, %d)", x, y)
				}
			}
		}
	})
}

func TestProcessor_ApplyAll(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	t.Run("all effects disabled", func(t *testing.T) {
		p := NewProcessor()
		p.config.ColorGrading.Enabled = false
		p.config.Vignette.Enabled = false

		result := p.ApplyAll(img, nil, nil)

		// Should return original image
		if result == nil {
			t.Fatal("ApplyAll returned nil")
		}
	})

	t.Run("color grading enabled", func(t *testing.T) {
		p := NewProcessor()
		p.config.ColorGrading.Enabled = true
		p.config.ColorGrading.Brightness = 0.1
		p.config.Vignette.Enabled = false

		result := p.ApplyAll(img, nil, nil)

		if result == nil {
			t.Fatal("ApplyAll returned nil")
		}

		// Check that brightness was applied
		centerColor := result.RGBAAt(5, 5)
		if centerColor.R <= 128 {
			t.Error("Brightness adjustment should have increased pixel value")
		}
	})
}
