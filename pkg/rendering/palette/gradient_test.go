// Package palette provides gradient generation tests.
// Phase 19.2: Dynamic Palette System - Gradient Generation Tests
package palette

import (
	"image/color"
	"math"
	"testing"
)

func TestGradientType_String(t *testing.T) {
	tests := []struct {
		name string
		g    GradientType
		want string
	}{
		{"Linear", GradientLinear, "Linear"},
		{"Radial", GradientRadial, "Radial"},
		{"Angular", GradientAngular, "Angular"},
		{"Diamond", GradientDiamond, "Diamond"},
		{"Spiral", GradientSpiral, "Spiral"},
		{"Conic", GradientConic, "Conic"},
		{"Unknown", GradientType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.String(); got != tt.want {
				t.Errorf("GradientType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultGradientConfig(t *testing.T) {
	config := DefaultGradientConfig()

	if config.Type != GradientLinear {
		t.Errorf("DefaultGradientConfig().Type = %v, want %v", config.Type, GradientLinear)
	}
	if len(config.Colors) != 2 {
		t.Errorf("DefaultGradientConfig() colors length = %d, want 2", len(config.Colors))
	}
	if config.CenterX != 0.5 || config.CenterY != 0.5 {
		t.Errorf("DefaultGradientConfig() center = (%f, %f), want (0.5, 0.5)", config.CenterX, config.CenterY)
	}
}

func TestGenerateGradient_Linear(t *testing.T) {
	config := GradientConfig{
		Type:   GradientLinear,
		Colors: []color.Color{color.Black, color.White},
		Angle:  0.0,
	}

	img := GenerateGradient(100, 100, config)

	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 100 {
		t.Errorf("GenerateGradient() size = %dx%d, want 100x100", img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Check corners - left should be darker than right for angle 0
	left := img.At(0, 50)
	right := img.At(99, 50)

	lr, lg, lb, _ := left.RGBA()
	rr, rg, rb, _ := right.RGBA()

	leftBrightness := float64(lr+lg+lb) / 3.0
	rightBrightness := float64(rr+rg+rb) / 3.0

	if leftBrightness >= rightBrightness {
		t.Errorf("Linear gradient left brightness %f >= right brightness %f", leftBrightness, rightBrightness)
	}
}

func TestGenerateGradient_Radial(t *testing.T) {
	config := GradientConfig{
		Type:    GradientRadial,
		Colors:  []color.Color{color.White, color.Black},
		CenterX: 0.5,
		CenterY: 0.5,
		Radius:  0.5,
	}

	img := GenerateGradient(100, 100, config)

	// Check center is brighter than edges
	center := img.At(50, 50)
	edge := img.At(0, 0)

	cr, cg, cb, _ := center.RGBA()
	er, eg, eb, _ := edge.RGBA()

	centerBrightness := float64(cr+cg+cb) / 3.0
	edgeBrightness := float64(er+eg+eb) / 3.0

	if centerBrightness <= edgeBrightness {
		t.Errorf("Radial gradient center brightness %f <= edge brightness %f", centerBrightness, edgeBrightness)
	}
}

func TestGenerateGradient_Angular(t *testing.T) {
	config := GradientConfig{
		Type:    GradientAngular,
		Colors:  []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}},
		CenterX: 0.5,
		CenterY: 0.5,
	}

	img := GenerateGradient(100, 100, config)

	// Just verify it generates without panic
	if img == nil {
		t.Error("GenerateGradient() returned nil image")
	}
}

func TestGenerateGradient_Diamond(t *testing.T) {
	config := GradientConfig{
		Type:    GradientDiamond,
		Colors:  []color.Color{color.White, color.Black},
		CenterX: 0.5,
		CenterY: 0.5,
	}

	img := GenerateGradient(100, 100, config)

	if img == nil {
		t.Error("GenerateGradient() returned nil image")
	}
}

func TestGenerateGradient_Spiral(t *testing.T) {
	config := GradientConfig{
		Type:            GradientSpiral,
		Colors:          []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}},
		CenterX:         0.5,
		CenterY:         0.5,
		SpiralRotations: 3.0,
	}

	img := GenerateGradient(100, 100, config)

	if img == nil {
		t.Error("GenerateGradient() returned nil image")
	}
}

func TestGenerateGradient_Conic(t *testing.T) {
	config := GradientConfig{
		Type:    GradientConic,
		Colors:  []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}},
		CenterX: 0.5,
		CenterY: 0.5,
		Angle:   45.0,
	}

	img := GenerateGradient(100, 100, config)

	if img == nil {
		t.Error("GenerateGradient() returned nil image")
	}
}

func TestGenerateGradient_Reverse(t *testing.T) {
	config := GradientConfig{
		Type:    GradientLinear,
		Colors:  []color.Color{color.Black, color.White},
		Angle:   0.0,
		Reverse: true,
	}

	img := GenerateGradient(100, 100, config)

	// Check corners - right should be darker than left for reversed
	left := img.At(0, 50)
	right := img.At(99, 50)

	lr, lg, lb, _ := left.RGBA()
	rr, rg, rb, _ := right.RGBA()

	leftBrightness := float64(lr+lg+lb) / 3.0
	rightBrightness := float64(rr+rg+rb) / 3.0

	if leftBrightness <= rightBrightness {
		t.Errorf("Reversed gradient left brightness %f <= right brightness %f", leftBrightness, rightBrightness)
	}
}

func TestGenerateGradient_EmptyColors(t *testing.T) {
	config := GradientConfig{
		Type:   GradientLinear,
		Colors: []color.Color{}, // Empty - should default to black-white
		Angle:  0.0,
	}

	img := GenerateGradient(100, 100, config)

	if img == nil {
		t.Error("GenerateGradient() with empty colors returned nil")
	}
}

func TestGenerateGradient_SingleColor(t *testing.T) {
	config := GradientConfig{
		Type:   GradientLinear,
		Colors: []color.Color{color.RGBA{128, 128, 128, 255}},
		Angle:  0.0,
	}

	img := GenerateGradient(100, 100, config)

	if img == nil {
		t.Error("GenerateGradient() with single color returned nil")
	}
}

func TestGenerateGradient_ZeroDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"zero width", 0, 100},
		{"zero height", 100, 0},
		{"both zero", 0, 0},
		{"negative width", -5, 100},
		{"negative height", 100, -5},
		{"both negative", -5, -5},
	}

	config := GradientConfig{
		Type:   GradientLinear,
		Colors: []color.Color{color.Black, color.White},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic, returns 1x1 minimum image
			img := GenerateGradient(tt.width, tt.height, config)
			if img == nil {
				t.Error("GenerateGradient() returned nil for edge case dimensions")
			}
			// Verify dimensions are at least 1x1
			if img.Bounds().Dx() < 1 || img.Bounds().Dy() < 1 {
				t.Errorf("GenerateGradient() returned invalid dimensions: %dx%d",
					img.Bounds().Dx(), img.Bounds().Dy())
			}
		})
	}
}

func TestCalculateRadialGradient_ZeroRadius(t *testing.T) {
	tests := []struct {
		name   string
		radius float64
	}{
		{"zero radius", 0.0},
		{"negative radius", -0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic, treats as radius=1.0
			got := calculateRadialGradient(0.5, 0.5, 0.0, 0.0, tt.radius)
			if got < 0 || got > 1 {
				t.Errorf("calculateRadialGradient with radius=%f returned %f, want [0, 1]", tt.radius, got)
			}
		})
	}
}

func TestCalculateLinearGradient(t *testing.T) {
	tests := []struct {
		name  string
		x, y  float64
		angle float64
		want  float64
	}{
		{"horizontal_start", 0.0, 0.5, 0.0, 0.0},
		{"horizontal_end", 1.0, 0.5, 0.0, 1.0},
		{"vertical_start", 0.5, 0.0, 90.0, 0.0},
		{"vertical_end", 0.5, 1.0, 90.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateLinearGradient(tt.x, tt.y, tt.angle)
			if math.Abs(got-tt.want) > 0.1 {
				t.Errorf("calculateLinearGradient(%f, %f, %f) = %f, want %f", tt.x, tt.y, tt.angle, got, tt.want)
			}
		})
	}
}

func TestCalculateRadialGradient(t *testing.T) {
	tests := []struct {
		name       string
		x, y       float64
		cx, cy     float64
		radius     float64
		wantApprox float64
		tolerance  float64
	}{
		{"center", 0.5, 0.5, 0.5, 0.5, 0.5, 0.0, 0.01},
		{"edge", 1.0, 0.5, 0.5, 0.5, 0.5, 1.0, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRadialGradient(tt.x, tt.y, tt.cx, tt.cy, tt.radius)
			if math.Abs(got-tt.wantApprox) > tt.tolerance {
				t.Errorf("calculateRadialGradient() = %f, want ~%f", got, tt.wantApprox)
			}
		})
	}
}

func TestCalculateAngularGradient(t *testing.T) {
	// Test that angular gradient returns values in [0, 1]
	tests := []struct {
		name   string
		x, y   float64
		cx, cy float64
	}{
		{"right", 1.0, 0.5, 0.5, 0.5},
		{"top", 0.5, 0.0, 0.5, 0.5},
		{"left", 0.0, 0.5, 0.5, 0.5},
		{"bottom", 0.5, 1.0, 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateAngularGradient(tt.x, tt.y, tt.cx, tt.cy)
			if got < 0 || got > 1 {
				t.Errorf("calculateAngularGradient() = %f, want [0, 1]", got)
			}
		})
	}
}

func TestCalculateDiamondGradient(t *testing.T) {
	// Test center and corner
	center := calculateDiamondGradient(0.5, 0.5, 0.5, 0.5)
	corner := calculateDiamondGradient(1.0, 1.0, 0.5, 0.5)

	if center >= corner {
		t.Errorf("Diamond gradient center %f >= corner %f", center, corner)
	}
}

func TestCalculateSpiralGradient(t *testing.T) {
	// Test that spiral gradient returns values in [0, 1]
	got := calculateSpiralGradient(0.7, 0.3, 0.5, 0.5, 2.0)
	if got < 0 || got > 1 {
		t.Errorf("calculateSpiralGradient() = %f, want [0, 1]", got)
	}
}

func TestCalculateConicGradient(t *testing.T) {
	// Test that conic gradient returns values in [0, 1]
	tests := []struct {
		name       string
		x, y       float64
		cx, cy     float64
		startAngle float64
	}{
		{"right", 1.0, 0.5, 0.5, 0.5, 0.0},
		{"with_offset", 0.7, 0.3, 0.5, 0.5, 45.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateConicGradient(tt.x, tt.y, tt.cx, tt.cy, tt.startAngle)
			if got < 0 || got > 1 {
				t.Errorf("calculateConicGradient() = %f, want [0, 1]", got)
			}
		})
	}
}

func TestApplySmoothness(t *testing.T) {
	tests := []struct {
		name       string
		t          float64
		smoothness float64
		wantMin    float64
		wantMax    float64
	}{
		{"no_smoothing", 0.5, 0.0, 0.5, 0.5},
		{"full_smoothing", 0.5, 1.0, 0.0, 1.0},
		{"half_smoothing", 0.5, 0.5, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applySmoothness(tt.t, tt.smoothness)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("applySmoothness(%f, %f) = %f, want [%f, %f]", tt.t, tt.smoothness, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestInterpolateColors_TwoColors(t *testing.T) {
	colors := []color.Color{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}

	// Test start
	c := interpolateColors(colors, 0.0)
	r, g, b, _ := c.RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("interpolateColors at 0.0 = (%d, %d, %d), want (0, 0, 0)", r>>8, g>>8, b>>8)
	}

	// Test end
	c = interpolateColors(colors, 1.0)
	r, g, b, _ = c.RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("interpolateColors at 1.0 = (%d, %d, %d), want (255, 255, 255)", r>>8, g>>8, b>>8)
	}

	// Test middle
	c = interpolateColors(colors, 0.5)
	r, g, b, _ = c.RGBA()
	// Should be approximately gray
	if r>>8 < 100 || r>>8 > 155 {
		t.Errorf("interpolateColors at 0.5 red = %d, want ~127", r>>8)
	}
}

func TestInterpolateColors_ThreeColors(t *testing.T) {
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255}, // Red
		color.RGBA{0, 255, 0, 255}, // Green
		color.RGBA{0, 0, 255, 255}, // Blue
	}

	// Test at position 0.0 (should be red)
	c := interpolateColors(colors, 0.0)
	r, g, b, _ := c.RGBA()
	if r>>8 < 200 || g>>8 > 55 || b>>8 > 55 {
		t.Errorf("interpolateColors at 0.0 = (%d, %d, %d), want red", r>>8, g>>8, b>>8)
	}

	// Test at position 1.0 (should be blue)
	c = interpolateColors(colors, 1.0)
	r, g, b, _ = c.RGBA()
	if r>>8 > 55 || g>>8 > 55 || b>>8 < 200 {
		t.Errorf("interpolateColors at 1.0 = (%d, %d, %d), want blue", r>>8, g>>8, b>>8)
	}
}

func TestInterpolateColors_EmptyColors(t *testing.T) {
	c := interpolateColors([]color.Color{}, 0.5)
	// interpolateColors with empty slice returns color.Black (0,0,0,255 in alpha-premultiplied form).
	// color.Black.RGBA() returns (0, 0, 0, 0xffff); >> 8 gives (0, 0, 0, 255).
	r, g, b, _ := c.RGBA()
	if r>>8 != 0 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("interpolateColors with empty array = (%d,%d,%d), want (0,0,0)", r>>8, g>>8, b>>8)
	}
}

func TestInterpolateColors_SingleColor(t *testing.T) {
	colors := []color.Color{color.RGBA{128, 128, 128, 255}}
	c := interpolateColors(colors, 0.5)
	r, g, b, _ := c.RGBA()
	if r>>8 != 128 || g>>8 != 128 || b>>8 != 128 {
		t.Errorf("interpolateColors single color = (%d, %d, %d), want (128, 128, 128)", r>>8, g>>8, b>>8)
	}
}

func TestInterpolateColors_Clamping(t *testing.T) {
	colors := []color.Color{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}

	// Test below 0
	c := interpolateColors(colors, -0.5)
	r, _, _, _ := c.RGBA()
	if r>>8 != 0 {
		t.Errorf("interpolateColors at -0.5 = %d, want 0 (clamped)", r>>8)
	}

	// Test above 1
	c = interpolateColors(colors, 1.5)
	r, _, _, _ = c.RGBA()
	if r>>8 != 255 {
		t.Errorf("interpolateColors at 1.5 = %d, want 255 (clamped)", r>>8)
	}
}

func TestCreateGradientPalette(t *testing.T) {
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	palette := CreateGradientPalette(colors, 20)

	if palette == nil {
		t.Fatal("CreateGradientPalette returned nil")
	}

	if len(palette.Colors) != 20 {
		t.Errorf("CreateGradientPalette colors length = %d, want 20", len(palette.Colors))
	}

	// Check that primary is set
	if palette.Primary == nil {
		t.Error("CreateGradientPalette Primary is nil")
	}

	// Check that text color is set based on background
	if palette.Text == nil {
		t.Error("CreateGradientPalette Text is nil")
	}

	// Check UI colors are set
	if palette.Danger == nil {
		t.Error("CreateGradientPalette Danger is nil")
	}
}

func TestCreateGradientPalette_EmptyColors(t *testing.T) {
	palette := CreateGradientPalette([]color.Color{}, 12)

	if palette == nil {
		t.Fatal("CreateGradientPalette with empty colors returned nil")
	}

	if len(palette.Colors) != 12 {
		t.Errorf("CreateGradientPalette colors length = %d, want 12", len(palette.Colors))
	}
}

func TestCreateGradientPalette_MinSteps(t *testing.T) {
	colors := []color.Color{color.RGBA{255, 0, 0, 255}}
	palette := CreateGradientPalette(colors, 5) // Below minimum

	if len(palette.Colors) != 12 {
		t.Errorf("CreateGradientPalette with steps=5 got %d colors, want 12 (minimum)", len(palette.Colors))
	}
}

func BenchmarkGenerateGradient_Linear(b *testing.B) {
	config := GradientConfig{
		Type:   GradientLinear,
		Colors: []color.Color{color.Black, color.White},
		Angle:  45.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateGradient(256, 256, config)
	}
}

func BenchmarkGenerateGradient_Radial(b *testing.B) {
	config := GradientConfig{
		Type:    GradientRadial,
		Colors:  []color.Color{color.White, color.Black},
		CenterX: 0.5,
		CenterY: 0.5,
		Radius:  0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateGradient(256, 256, config)
	}
}

func BenchmarkGenerateGradient_Angular(b *testing.B) {
	config := GradientConfig{
		Type:    GradientAngular,
		Colors:  []color.Color{color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}},
		CenterX: 0.5,
		CenterY: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateGradient(256, 256, config)
	}
}

func BenchmarkInterpolateColors(b *testing.B) {
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		interpolateColors(colors, 0.5)
	}
}

func BenchmarkCreateGradientPalette(b *testing.B) {
	colors := []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateGradientPalette(colors, 20)
	}
}
