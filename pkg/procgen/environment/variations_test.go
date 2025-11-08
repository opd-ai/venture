// Package environment provides procedural generation of environmental objects.
package environment

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestVisualVariation_Default(t *testing.T) {
	v := DefaultVariation()

	if v.Rotation != 0 {
		t.Errorf("Expected rotation 0, got %f", v.Rotation)
	}
	if v.Scale != 1.0 {
		t.Errorf("Expected scale 1.0, got %f", v.Scale)
	}
	if v.ColorShift != 0 {
		t.Errorf("Expected colorShift 0, got %f", v.ColorShift)
	}
	if v.Brightness != 0 {
		t.Errorf("Expected brightness 0, got %f", v.Brightness)
	}
	if v.FlipHorizontal {
		t.Error("Expected flipHorizontal false")
	}
	if v.FlipVertical {
		t.Error("Expected flipVertical false")
	}
}

func TestGenerateVariation_Deterministic(t *testing.T) {
	seed := int64(12345)
	rng1 := rand.New(rand.NewSource(seed))
	rng2 := rand.New(rand.NewSource(seed))

	v1 := GenerateVariation(SubTypePlant, rng1)
	v2 := GenerateVariation(SubTypePlant, rng2)

	if v1.Rotation != v2.Rotation {
		t.Errorf("Rotation mismatch: %f != %f", v1.Rotation, v2.Rotation)
	}
	if v1.Scale != v2.Scale {
		t.Errorf("Scale mismatch: %f != %f", v1.Scale, v2.Scale)
	}
	if v1.ColorShift != v2.ColorShift {
		t.Errorf("ColorShift mismatch: %f != %f", v1.ColorShift, v2.ColorShift)
	}
	if v1.Brightness != v2.Brightness {
		t.Errorf("Brightness mismatch: %f != %f", v1.Brightness, v2.Brightness)
	}
}

func TestGenerateVariation_Decorations(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	decorations := []SubType{
		SubTypePlant, SubTypeStatue, SubTypePainting, SubTypeBanner,
		SubTypeSconce, SubTypeBloodstain, SubTypeGrass, SubTypeMushroom,
	}

	for _, subType := range decorations {
		t.Run(subType.String(), func(t *testing.T) {
			v := GenerateVariation(subType, rng)

			// Decorations should get full rotation range
			if v.Rotation < 0 || v.Rotation >= 360 {
				t.Errorf("Rotation out of range: %f", v.Rotation)
			}

			// Scale should be reasonable
			if v.Scale < 0.5 || v.Scale > 2.0 {
				t.Errorf("Scale out of range: %f", v.Scale)
			}
		})
	}
}

func TestGenerateVariation_Furniture(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	furniture := []SubType{SubTypeTable, SubTypeChair, SubTypeChest}

	for _, subType := range furniture {
		t.Run(subType.String(), func(t *testing.T) {
			v := GenerateVariation(subType, rng)

			// Furniture should get 90-degree increments
			if v.Rotation != 0 && v.Rotation != 90 && v.Rotation != 180 && v.Rotation != 270 {
				t.Errorf("Furniture rotation should be 90-degree increment, got %f", v.Rotation)
			}
		})
	}
}

func TestRGBToHSL_Black(t *testing.T) {
	h, s, l := rgbToHSL(0, 0, 0)

	if h != 0 {
		t.Errorf("Black hue should be 0, got %f", h)
	}
	if s != 0 {
		t.Errorf("Black saturation should be 0, got %f", s)
	}
	if l != 0 {
		t.Errorf("Black lightness should be 0, got %f", l)
	}
}

func TestRGBToHSL_White(t *testing.T) {
	h, s, l := rgbToHSL(255, 255, 255)

	if h != 0 {
		t.Errorf("White hue should be 0, got %f", h)
	}
	if s != 0 {
		t.Errorf("White saturation should be 0, got %f", s)
	}
	if l != 1.0 {
		t.Errorf("White lightness should be 1.0, got %f", l)
	}
}

func TestRGBToHSL_Red(t *testing.T) {
	h, s, l := rgbToHSL(255, 0, 0)

	if h < -1 || h > 361 {
		t.Errorf("Red hue out of range: %f", h)
	}
	if s < 0.9 {
		t.Errorf("Red should have high saturation, got %f", s)
	}
	if l < 0.4 || l > 0.6 {
		t.Errorf("Red lightness should be ~0.5, got %f", l)
	}
}

func TestHSLToRGB_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
	}{
		{"black", 0, 0, 0},
		{"white", 255, 255, 255},
		{"red", 255, 0, 0},
		{"green", 0, 255, 0},
		{"blue", 0, 0, 255},
		{"gray", 128, 128, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
			r2, g2, b2 := hslToRGB(h, s, l)

			// Allow small rounding error
			if absInt(int(tt.r)-int(r2)) > 2 || absInt(int(tt.g)-int(g2)) > 2 || absInt(int(tt.b)-int(b2)) > 2 {
				t.Errorf("RGB->HSL->RGB roundtrip failed: (%d,%d,%d) -> (%f,%f,%f) -> (%d,%d,%d)",
					tt.r, tt.g, tt.b, h, s, l, r2, g2, b2)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min, max float64
		expected float64
	}{
		{"below min", -5.0, 0.0, 10.0, 0.0},
		{"above max", 15.0, 0.0, 10.0, 10.0},
		{"in range", 5.0, 0.0, 10.0, 5.0},
		{"at min", 0.0, 0.0, 10.0, 0.0},
		{"at max", 10.0, 0.0, 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("clamp(%f, %f, %f) = %f, expected %f",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestBilinearSample(t *testing.T) {
	// Create a 4x4 test image with a gradient
	img := createTestImage(4, 4)

	// Test sampling at exact pixel center
	c := bilinearSample(img, 1.5, 1.5)
	if _, ok := c.(color.RGBA); !ok {
		t.Error("bilinearSample should return color.RGBA")
	}

	// Test sampling out of bounds
	c = bilinearSample(img, -1, -1)
	rgba := c.(color.RGBA)
	if rgba.A != 0 {
		t.Error("Out-of-bounds sample should be transparent")
	}
}

func TestApplyFlips_Horizontal(t *testing.T) {
	img := createTestImage(4, 4)
	original := img.At(0, 0)

	flipped := applyFlips(img, true, false)

	// Check that the image was flipped
	if flipped.At(3, 0) != original {
		t.Error("Horizontal flip failed")
	}
}

func TestApplyFlips_Vertical(t *testing.T) {
	img := createTestImage(4, 4)
	original := img.At(0, 0)

	flipped := applyFlips(img, false, true)

	// Check that the image was flipped
	if flipped.At(0, 3) != original {
		t.Error("Vertical flip failed")
	}
}

func TestApplyFlips_Both(t *testing.T) {
	img := createTestImage(4, 4)
	original := img.At(0, 0)

	flipped := applyFlips(img, true, true)

	// Check that the image was flipped both ways
	if flipped.At(3, 3) != original {
		t.Error("Both flips failed")
	}
}

func TestApplyColorAdjustments_Brightness(t *testing.T) {
	img := createTestImage(32, 32)

	// Apply brightness increase
	adjusted := applyColorAdjustments(img, 0, 0.2)

	// Verify image was modified
	if adjusted.At(16, 16) == img.At(16, 16) {
		t.Error("Brightness adjustment should modify colors")
	}
}

func TestApplyColorAdjustments_HueShift(t *testing.T) {
	img := createTestImage(32, 32)

	// Apply hue shift
	adjusted := applyColorAdjustments(img, 60, 0)

	// Verify image was modified
	if adjusted.At(16, 16) == img.At(16, 16) {
		t.Error("Hue shift should modify colors")
	}
}

func TestApplyTransform_NoChange(t *testing.T) {
	img := createTestImage(32, 32)

	// No rotation or scale change
	transformed := applyTransform(img, 0, 1.0)

	// Dimensions should be the same
	if transformed.Bounds().Dx() != 32 || transformed.Bounds().Dy() != 32 {
		t.Error("No transform should maintain dimensions")
	}
}

func TestApplyTransform_Scale(t *testing.T) {
	img := createTestImage(32, 32)

	// Scale up by 2x
	transformed := applyTransform(img, 0, 2.0)

	// Dimensions should double
	if transformed.Bounds().Dx() != 64 || transformed.Bounds().Dy() != 64 {
		t.Errorf("2x scale should double dimensions, got %dx%d",
			transformed.Bounds().Dx(), transformed.Bounds().Dy())
	}
}

func TestApplyVariation_NoChange(t *testing.T) {
	img := createTestImage(32, 32)
	v := DefaultVariation()

	result := ApplyVariation(img, v)

	// Should return same image reference
	if result != img {
		t.Error("Default variation should not modify image")
	}
}

func TestApplyVariation_WithChanges(t *testing.T) {
	img := createTestImage(32, 32)
	v := VisualVariation{
		Rotation:   45,
		Scale:      1.2,
		ColorShift: 30,
		Brightness: 0.1,
	}

	result := ApplyVariation(img, v)

	// Should return different image
	if result == img {
		t.Error("Variation should create new image")
	}
}

// Helper function to create a test image
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8(x * 255 / width)
			g := uint8(y * 255 / height)
			b := uint8(128)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

// absInt returns the absolute value of an integer.
// This is a test helper function to avoid dependency on unexported functions.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests

func BenchmarkGenerateVariation(b *testing.B) {
	rng := rand.New(rand.NewSource(12345))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GenerateVariation(SubTypePlant, rng)
	}
}

func BenchmarkApplyColorAdjustments(b *testing.B) {
	img := createTestImage(32, 32)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		applyColorAdjustments(img, 30, 0.1)
	}
}

func BenchmarkApplyFlips(b *testing.B) {
	img := createTestImage(32, 32)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		applyFlips(img, true, false)
	}
}

func BenchmarkApplyTransform(b *testing.B) {
	img := createTestImage(32, 32)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		applyTransform(img, 45, 1.0)
	}
}

func BenchmarkApplyVariation(b *testing.B) {
	img := createTestImage(32, 32)
	v := VisualVariation{
		Rotation:   45,
		Scale:      1.2,
		ColorShift: 30,
		Brightness: 0.1,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ApplyVariation(img, v)
	}
}
