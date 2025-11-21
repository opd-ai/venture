package parity

import (
	"image"
	"image/color"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()

	if v == nil {
		t.Fatal("NewValidator returned nil")
	}

	if v.platform == PlatformUnknown {
		t.Error("Validator has unknown platform")
	}

	if v.baselines == nil {
		t.Error("Validator baselines map is nil")
	}

	if v.tolerance.MaxRGBDelta != 1 {
		t.Errorf("expected MaxRGBDelta=1, got %d", v.tolerance.MaxRGBDelta)
	}

	if v.tolerance.MaxPercentDiff != 5.0 {
		t.Errorf("expected MaxPercentDiff=5.0, got %.2f", v.tolerance.MaxPercentDiff)
	}
}

func TestSetTolerance(t *testing.T) {
	v := NewValidator()

	customTolerance := ColorTolerance{
		MaxRGBDelta:    2,
		MaxPercentDiff: 10.0,
	}

	v.SetTolerance(customTolerance)

	if v.tolerance.MaxRGBDelta != 2 {
		t.Errorf("expected MaxRGBDelta=2, got %d", v.tolerance.MaxRGBDelta)
	}

	if v.tolerance.MaxPercentDiff != 10.0 {
		t.Errorf("expected MaxPercentDiff=10.0, got %.2f", v.tolerance.MaxPercentDiff)
	}
}

func TestBaselineStorage(t *testing.T) {
	v := NewValidator()

	baseline := &Baseline{
		Platform:  PlatformLinux,
		FrameRate: 60.0,
	}

	// Test setting baseline
	v.SetBaseline("test", baseline)

	// Test retrieving baseline
	retrieved, exists := v.GetBaseline("test")
	if !exists {
		t.Fatal("baseline not found after setting")
	}

	if retrieved.Platform != PlatformLinux {
		t.Errorf("expected Platform=PlatformLinux, got %v", retrieved.Platform)
	}

	if retrieved.FrameRate != 60.0 {
		t.Errorf("expected FrameRate=60.0, got %.2f", retrieved.FrameRate)
	}

	// Test non-existent baseline
	_, exists = v.GetBaseline("nonexistent")
	if exists {
		t.Error("non-existent baseline reported as existing")
	}
}

func TestCompareImages_Identical(t *testing.T) {
	v := NewValidator()

	// Create identical images
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with same color
	c := color.RGBA{100, 150, 200, 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img1.Set(x, y, c)
			img2.Set(x, y, c)
		}
	}

	diffPct, errors := v.CompareImages(img1, img2)

	if diffPct != 0.0 {
		t.Errorf("expected 0%% difference for identical images, got %.2f%%", diffPct)
	}

	if len(errors) > 0 {
		t.Errorf("expected no errors for identical images, got %v", errors)
	}
}

func TestCompareImages_Different(t *testing.T) {
	v := NewValidator()

	// Create different images
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Fill with different colors
	c1 := color.RGBA{100, 150, 200, 255}
	c2 := color.RGBA{110, 160, 210, 255} // Different by 10

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img1.Set(x, y, c1)
			img2.Set(x, y, c2)
		}
	}

	diffPct, errors := v.CompareImages(img1, img2)

	if diffPct == 0.0 {
		t.Error("expected non-zero difference for different images")
	}

	// With default tolerance (MaxRGBDelta=1), should fail
	if len(errors) == 0 {
		t.Error("expected errors for images exceeding tolerance")
	}
}

func TestCompareImages_WithinTolerance(t *testing.T) {
	v := NewValidator()

	// Create slightly different images within tolerance
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))

	c1 := color.RGBA{100, 150, 200, 255}
	c2 := color.RGBA{101, 151, 200, 255} // Different by 1

	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img1.Set(x, y, c1)
			img2.Set(x, y, c2)
		}
	}

	diffPct, errors := v.CompareImages(img1, img2)

	// Should have 100% different pixels, but within tolerance
	if diffPct != 100.0 {
		t.Errorf("expected 100%% pixels to differ by <=1, got %.2f%%", diffPct)
	}

	// Should fail because >5% threshold
	if len(errors) == 0 {
		t.Error("expected errors for 100% difference exceeding 5% threshold")
	}
}

func TestCompareImages_DifferentDimensions(t *testing.T) {
	v := NewValidator()

	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 20, 20))

	diffPct, errors := v.CompareImages(img1, img2)

	if diffPct != 100.0 {
		t.Errorf("expected 100%% difference for different dimensions, got %.2f%%", diffPct)
	}

	if len(errors) == 0 {
		t.Error("expected errors for different dimensions")
	}
}

func TestCompareImages_Nil(t *testing.T) {
	v := NewValidator()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	diffPct, errors := v.CompareImages(nil, img)
	if diffPct != 100.0 || len(errors) == 0 {
		t.Error("expected 100% difference and errors for nil image")
	}

	diffPct, errors = v.CompareImages(img, nil)
	if diffPct != 100.0 || len(errors) == 0 {
		t.Error("expected 100% difference and errors for nil image")
	}
}

func TestCompareColors_Identical(t *testing.T) {
	v := NewValidator()

	colors1 := []color.RGBA{
		{100, 150, 200, 255},
		{50, 100, 150, 255},
	}
	colors2 := []color.RGBA{
		{100, 150, 200, 255},
		{50, 100, 150, 255},
	}

	diffPct, errors := v.CompareColors(colors1, colors2)

	if diffPct != 0.0 {
		t.Errorf("expected 0%% difference for identical colors, got %.2f%%", diffPct)
	}

	if len(errors) > 0 {
		t.Errorf("expected no errors for identical colors, got %v", errors)
	}
}

func TestCompareColors_Different(t *testing.T) {
	v := NewValidator()

	colors1 := []color.RGBA{{100, 150, 200, 255}}
	colors2 := []color.RGBA{{110, 160, 210, 255}}

	diffPct, errors := v.CompareColors(colors1, colors2)

	if diffPct == 0.0 {
		t.Error("expected non-zero difference for different colors")
	}

	if len(errors) == 0 {
		t.Error("expected errors for colors exceeding tolerance")
	}
}

func TestCompareFrameRate(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		fps1       float64
		fps2       float64
		wantErrors bool
	}{
		{"identical", 60.0, 60.0, false},
		{"within 20%", 60.0, 65.0, false},
		{"exceeds 20%", 60.0, 75.0, true},
		{"invalid fps1", 0.0, 60.0, true},
		{"invalid fps2", 60.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := v.CompareFrameRate(tt.fps1, tt.fps2)

			hasErrors := len(errors) > 0
			if hasErrors != tt.wantErrors {
				t.Errorf("wantErrors=%v, got hasErrors=%v", tt.wantErrors, hasErrors)
			}
		})
	}
}

func TestValidateSprites(t *testing.T) {
	v := NewValidator()

	testImage := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			testImage.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}

	// No baseline - should warn but pass
	result := v.ValidateSprites(testImage)
	if !result.Passed {
		t.Error("ValidateSprites should pass without baseline")
	}
	if len(result.Warnings) == 0 {
		t.Error("ValidateSprites should warn about missing baseline")
	}

	// Set baseline
	v.SetBaseline("sprite_rendering", &Baseline{
		Platform:  v.platform,
		ImageData: testImage,
	})

	// With identical baseline - should pass
	result = v.ValidateSprites(testImage)
	if !result.Passed {
		t.Errorf("ValidateSprites failed: %v", result.Errors)
	}
	if result.VisualDiffPct != 0.0 {
		t.Errorf("expected 0%% diff, got %.2f%%", result.VisualDiffPct)
	}
}

func TestValidateFrameRate(t *testing.T) {
	v := NewValidator()

	// Below target
	result := v.ValidateFrameRate(30.0)
	if result.Passed {
		t.Error("ValidateFrameRate should fail for 30 FPS")
	}

	// Above target
	result = v.ValidateFrameRate(60.0)
	if !result.Passed {
		t.Errorf("ValidateFrameRate failed for 60 FPS: %v", result.Errors)
	}

	// Well above target
	result = v.ValidateFrameRate(120.0)
	if !result.Passed {
		t.Errorf("ValidateFrameRate failed for 120 FPS: %v", result.Errors)
	}
}

func TestValidateResolution(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		resolution image.Point
		wantWarn   bool
	}{
		{"1280x720", image.Pt(1280, 720), false},
		{"1920x1080", image.Pt(1920, 1080), false},
		{"2560x1440", image.Pt(2560, 1440), false},
		{"3840x2160", image.Pt(3840, 2160), false},
		{"unsupported", image.Pt(640, 480), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ValidateResolution(tt.resolution)

			if !result.Passed {
				t.Error("ValidateResolution should always pass")
			}

			hasWarnings := len(result.Warnings) > 0
			if hasWarnings != tt.wantWarn {
				t.Errorf("wantWarn=%v, got hasWarnings=%v", tt.wantWarn, hasWarnings)
			}
		})
	}
}

func BenchmarkCompareImages(b *testing.B) {
	v := NewValidator()

	img1 := image.NewRGBA(image.Rect(0, 0, 64, 64))
	img2 := image.NewRGBA(image.Rect(0, 0, 64, 64))

	c := color.RGBA{100, 150, 200, 255}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img1.Set(x, y, c)
			img2.Set(x, y, c)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.CompareImages(img1, img2)
	}
}

func BenchmarkValidateSprites(b *testing.B) {
	v := NewValidator()

	testImage := image.NewRGBA(image.Rect(0, 0, 64, 64))
	v.SetBaseline("sprite_rendering", &Baseline{
		Platform:  v.platform,
		ImageData: testImage,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateSprites(testImage)
	}
}
