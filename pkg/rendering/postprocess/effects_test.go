// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"testing"
)

// Helper function to create a solid color image
func createTestImage(width, height int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// Helper function to create a gradient image (horizontal)
func createGradientImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := uint8((float64(x) / float64(width)) * 255.0)
			img.Set(x, y, color.RGBA{value, value, value, 255})
		}
	}
	return img
}

func TestApplyMotionBlur_Disabled(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})
	p := NewProcessor()
	p.config.MotionBlur.Enabled = false

	result := p.ApplyMotionBlur(img, nil)

	if result != img {
		t.Error("ApplyMotionBlur with disabled effect should return original image")
	}
}

func TestApplyMotionBlur_NoVelocity(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})
	p := NewProcessor()
	p.config.MotionBlur.Enabled = true
	p.config.MotionBlur.VelocityX = 0.0
	p.config.MotionBlur.VelocityY = 0.0

	result := p.ApplyMotionBlur(img, nil)

	// Should return nearly identical image (no velocity)
	centerOrig := img.RGBAAt(5, 5)
	centerResult := result.RGBAAt(5, 5)

	if centerOrig != centerResult {
		t.Error("ApplyMotionBlur with zero velocity should not change image")
	}
}

func TestApplyMotionBlur_WithVelocity(t *testing.T) {
	img := createGradientImage(20, 20)
	p := NewProcessor()
	p.config.MotionBlur.Enabled = true
	p.config.MotionBlur.Intensity = 0.5
	p.config.MotionBlur.Samples = 5
	p.config.MotionBlur.VelocityX = 3.0

	result := p.ApplyMotionBlur(img, nil)

	if result == nil {
		t.Fatal("ApplyMotionBlur returned nil")
	}

	// Motion blur should affect the image
	bounds := result.Bounds()
	if bounds != img.Bounds() {
		t.Error("Result bounds should match input bounds")
	}
}

func TestCreateUniformVelocityMap(t *testing.T) {
	bounds := image.Rect(0, 0, 10, 10)
	vx, vy := 5.0, 3.0

	velMap := CreateUniformVelocityMap(bounds, vx, vy)

	// Check a few pixels
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			gotVx, gotVy := velMap.GetVelocity(x, y)
			if gotVx != vx || gotVy != vy {
				t.Errorf("Velocity at (%d, %d) = (%v, %v), want (%v, %v)",
					x, y, gotVx, gotVy, vx, vy)
			}
		}
	}
}

func TestCreateRadialVelocityMap(t *testing.T) {
	bounds := image.Rect(0, 0, 10, 10)
	centerX, centerY := 5, 5
	strength := 2.0

	velMap := CreateRadialVelocityMap(bounds, centerX, centerY, strength)

	// Check that center has minimal velocity
	vx, vy := velMap.GetVelocity(centerX, centerY)
	if vx != 0 || vy != 0 {
		t.Errorf("Center velocity should be (0, 0), got (%v, %v)", vx, vy)
	}

	// Check that edges have non-zero velocity
	vx, vy = velMap.GetVelocity(0, 0)
	if vx == 0 && vy == 0 {
		t.Error("Edge velocity should be non-zero")
	}
}

func TestApplyDepthBlur_Disabled(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})
	depthMap := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.DepthBlur.Enabled = false

	result := p.ApplyDepthBlur(img, depthMap)

	if result != img {
		t.Error("ApplyDepthBlur with disabled effect should return original image")
	}
}

func TestApplyDepthBlur_NilDepthMap(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.DepthBlur.Enabled = true

	result := p.ApplyDepthBlur(img, nil)

	if result != img {
		t.Error("ApplyDepthBlur with nil depth map should return original image")
	}
}

func TestGenerateDepthMapFromLuminance(t *testing.T) {
	// Create gradient image (bright to dark)
	img := createGradientImage(10, 10)

	depthMap := GenerateDepthMapFromLuminance(img)

	if depthMap == nil {
		t.Fatal("GenerateDepthMapFromLuminance returned nil")
	}

	// Check bounds
	if depthMap.Bounds() != img.Bounds() {
		t.Error("Depth map bounds should match image bounds")
	}

	// Check that bright areas have lower depth values (inverted)
	brightPixel := img.RGBAAt(9, 0) // Right side (bright)
	brightDepth := depthMap.RGBAAt(9, 0)

	darkPixel := img.RGBAAt(0, 0) // Left side (dark)
	darkDepth := depthMap.RGBAAt(0, 0)

	if brightPixel.R > darkPixel.R && brightDepth.R >= darkDepth.R {
		t.Error("Brighter pixels should have lower depth values")
	}
}

func TestGenerateDepthMapFromY(t *testing.T) {
	bounds := image.Rect(0, 0, 10, 10)
	depthMap := GenerateDepthMapFromY(bounds)

	if depthMap == nil {
		t.Fatal("GenerateDepthMapFromY returned nil")
	}

	// Top should be near (low depth value)
	topColor := depthMap.RGBAAt(5, 0)

	// Bottom should be far (high depth value)
	bottomColor := depthMap.RGBAAt(5, 9)

	if topColor.R >= bottomColor.R {
		t.Error("Top pixels should have lower depth than bottom pixels")
	}
}

func TestApplyColorGrading_Disabled(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.ColorGrading.Enabled = false

	result := p.ApplyColorGrading(img)

	if result != img {
		t.Error("ApplyColorGrading with disabled effect should return original image")
	}
}

func TestApplyColorGrading_Brightness(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.ColorGrading.Enabled = true
	p.config.ColorGrading.Brightness = 0.2
	p.config.ColorGrading.Saturation = 1.0
	p.config.ColorGrading.Contrast = 1.0

	result := p.ApplyColorGrading(img)

	centerColor := result.RGBAAt(5, 5)

	// Should be brighter than original
	if centerColor.R <= 128 {
		t.Error("Brightness adjustment should increase pixel value")
	}
}

func TestApplyColorGrading_Contrast(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.ColorGrading.Enabled = true
	p.config.ColorGrading.Brightness = 0.0
	p.config.ColorGrading.Saturation = 1.0
	p.config.ColorGrading.Contrast = 1.5

	result := p.ApplyColorGrading(img)

	// Contrast adjustment around midpoint (128) should not change much
	// because it pivots around 0.5
	if result == nil {
		t.Fatal("ApplyColorGrading returned nil")
	}
}

func TestApplyGrayscale(t *testing.T) {
	// Create colorful image
	img := createTestImage(10, 10, color.RGBA{255, 100, 50, 255})

	result := ApplyGrayscale(img)

	if result == nil {
		t.Fatal("ApplyGrayscale returned nil")
	}

	// Check that colors are equal (grayscale)
	centerColor := result.RGBAAt(5, 5)
	if centerColor.R != centerColor.G || centerColor.G != centerColor.B {
		t.Error("Grayscale should have equal RGB values")
	}
}

func TestApplySepia(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	result := ApplySepia(img)

	if result == nil {
		t.Fatal("ApplySepia returned nil")
	}

	// Sepia should have warmer tones (more red/yellow)
	centerColor := result.RGBAAt(5, 5)
	if centerColor.R < centerColor.B {
		t.Error("Sepia should have more red than blue")
	}
}

func TestApplyHueShift(t *testing.T) {
	// Create red image
	img := createTestImage(10, 10, color.RGBA{255, 0, 0, 255})

	// Shift hue by 1/3 (should become greenish)
	result := ApplyHueShift(img, 0.333)

	if result == nil {
		t.Fatal("ApplyHueShift returned nil")
	}

	centerColor := result.RGBAAt(5, 5)

	// After shifting red by 1/3, green should be dominant
	if centerColor.G < centerColor.R {
		t.Error("Hue shift of red by 1/3 should produce more green")
	}
}

func TestApplyVignette_Disabled(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.Vignette.Enabled = false

	result := p.ApplyVignette(img)

	if result != img {
		t.Error("ApplyVignette with disabled effect should return original image")
	}
}

func TestApplyVignette_Effect(t *testing.T) {
	img := createTestImage(100, 100, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.Vignette.Enabled = true
	p.config.Vignette.Intensity = 0.8
	p.config.Vignette.InnerRadius = 0.3
	p.config.Vignette.OuterRadius = 1.0

	result := p.ApplyVignette(img)

	if result == nil {
		t.Fatal("ApplyVignette returned nil")
	}

	// Center should be brighter than edges
	centerColor := result.RGBAAt(50, 50)
	edgeColor := result.RGBAAt(0, 0)

	if centerColor.R <= edgeColor.R {
		t.Error("Vignette should darken edges more than center")
	}
}

func TestApplyRadialGradient(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{128, 128, 128, 255})
	centerColor := color.RGBA{255, 255, 255, 255}
	edgeColor := color.RGBA{0, 0, 0, 255}

	result := ApplyRadialGradient(img, centerColor, edgeColor, 0.8)

	if result == nil {
		t.Fatal("ApplyRadialGradient returned nil")
	}

	// Center should be lighter than edges
	centerPixel := result.RGBAAt(25, 25)
	edgePixel := result.RGBAAt(0, 0)

	if centerPixel.R <= edgePixel.R {
		t.Error("Radial gradient should have lighter center")
	}
}

func TestApplyChromaticAberration_Disabled(t *testing.T) {
	img := createTestImage(10, 10, color.RGBA{128, 128, 128, 255})

	p := NewProcessor()
	p.config.ChromaticAberration.Enabled = false

	result := p.ApplyChromaticAberration(img)

	if result != img {
		t.Error("ApplyChromaticAberration with disabled effect should return original image")
	}
}

func TestApplyChromaticAberration_Effect(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{255, 255, 255, 255})

	p := NewProcessor()
	p.config.ChromaticAberration.Enabled = true
	p.config.ChromaticAberration.Intensity = 0.5
	p.config.ChromaticAberration.Samples = 3

	result := p.ApplyChromaticAberration(img)

	if result == nil {
		t.Fatal("ApplyChromaticAberration returned nil")
	}

	// Effect should be visible at edges (color separation)
	// Center should remain relatively unchanged
	bounds := result.Bounds()
	if bounds != img.Bounds() {
		t.Error("Result bounds should match input bounds")
	}
}

func TestApplyPrismaticAberration(t *testing.T) {
	img := createTestImage(50, 50, color.RGBA{255, 255, 255, 255})

	result := ApplyPrismaticAberration(img, 0.3, 0.0)

	if result == nil {
		t.Fatal("ApplyPrismaticAberration returned nil")
	}

	bounds := result.Bounds()
	if bounds != img.Bounds() {
		t.Error("Result bounds should match input bounds")
	}
}

// Benchmark tests
func BenchmarkApplyColorGrading(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})
	p := NewProcessor()
	p.config.ColorGrading.Enabled = true
	p.config.ColorGrading.Saturation = 1.2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ApplyColorGrading(img)
	}
}

func BenchmarkApplyVignette(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})
	p := NewProcessor()
	p.config.Vignette.Enabled = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ApplyVignette(img)
	}
}

func BenchmarkApplyMotionBlur(b *testing.B) {
	img := createTestImage(200, 200, color.RGBA{128, 128, 128, 255})
	velMap := CreateUniformVelocityMap(img.Bounds(), 3.0, 0.0)
	p := NewProcessor()
	p.config.MotionBlur.Enabled = true
	p.config.MotionBlur.Intensity = 0.5
	p.config.MotionBlur.Samples = 7

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ApplyMotionBlur(img, velMap)
	}
}
