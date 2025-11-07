// Package palette provides tests for time-of-day color modulation.
// Phase 17.3: Time-of-Day Color Shifts
package palette

import (
	"image/color"
	"math"
	"testing"
)

func TestTimeOfDay_String(t *testing.T) {
	tests := []struct {
		name      string
		timeOfDay TimeOfDay
		want      string
	}{
		{"dawn", TimeOfDayDawn, "Dawn"},
		{"day", TimeOfDayDay, "Day"},
		{"dusk", TimeOfDayDusk, "Dusk"},
		{"night", TimeOfDayNight, "Night"},
		{"unknown", TimeOfDay(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.timeOfDay.String()
			if got != tt.want {
				t.Errorf("TimeOfDay.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTimeConfig(t *testing.T) {
	config := DefaultTimeConfig()

	if config.CurrentTime != TimeOfDayDay {
		t.Errorf("DefaultTimeConfig().CurrentTime = %v, want %v", config.CurrentTime, TimeOfDayDay)
	}
	if config.TransitionProgress != 0.0 {
		t.Errorf("DefaultTimeConfig().TransitionProgress = %v, want 0.0", config.TransitionProgress)
	}
	if config.IntensityMultiplier != 1.0 {
		t.Errorf("DefaultTimeConfig().IntensityMultiplier = %v, want 1.0", config.IntensityMultiplier)
	}
}

func TestGetModulationForTime(t *testing.T) {
	tests := []struct {
		name      string
		timeOfDay TimeOfDay
		wantWarm  bool // Should have positive temperature shift
		wantDark  bool // Should have negative lightness offset
		wantDesat bool // Should have saturation multiplier < 1.0
	}{
		{"dawn", TimeOfDayDawn, true, false, true},
		{"day", TimeOfDayDay, false, false, false},
		{"dusk", TimeOfDayDusk, true, true, false},
		{"night", TimeOfDayNight, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := GetModulationForTime(tt.timeOfDay)

			if tt.wantWarm && mod.TemperatureShift <= 0.0 {
				t.Errorf("Expected warm temperature shift, got %v", mod.TemperatureShift)
			}
			if !tt.wantWarm && tt.timeOfDay != TimeOfDayDay && mod.TemperatureShift >= 0.0 && tt.name == "night" {
				t.Errorf("Expected cool temperature shift, got %v", mod.TemperatureShift)
			}
			if tt.wantDark && mod.LightnessOffset >= 0.0 {
				t.Errorf("Expected negative lightness offset, got %v", mod.LightnessOffset)
			}
			if tt.wantDesat && mod.SaturationMultiplier >= 1.0 {
				t.Errorf("Expected desaturation (< 1.0), got %v", mod.SaturationMultiplier)
			}
		})
	}
}

func TestGetModulationForTime_DayIsNeutral(t *testing.T) {
	mod := GetModulationForTime(TimeOfDayDay)

	if mod.HueShift != 0.0 {
		t.Errorf("Day HueShift = %v, want 0.0", mod.HueShift)
	}
	if mod.SaturationMultiplier != 1.0 {
		t.Errorf("Day SaturationMultiplier = %v, want 1.0", mod.SaturationMultiplier)
	}
	if mod.LightnessOffset != 0.0 {
		t.Errorf("Day LightnessOffset = %v, want 0.0", mod.LightnessOffset)
	}
	if mod.TemperatureShift != 0.0 {
		t.Errorf("Day TemperatureShift = %v, want 0.0", mod.TemperatureShift)
	}
}

func TestInterpolateModulation(t *testing.T) {
	from := ColorModulation{
		HueShift:             0.0,
		SaturationMultiplier: 1.0,
		LightnessOffset:      0.0,
		TemperatureShift:     0.0,
	}
	to := ColorModulation{
		HueShift:             10.0,
		SaturationMultiplier: 0.5,
		LightnessOffset:      -0.2,
		TemperatureShift:     0.5,
	}

	tests := []struct {
		name     string
		progress float64
		checkFn  func(ColorModulation) bool
	}{
		{
			name:     "progress 0.0 returns from",
			progress: 0.0,
			checkFn: func(m ColorModulation) bool {
				return m.HueShift == from.HueShift &&
					m.SaturationMultiplier == from.SaturationMultiplier
			},
		},
		{
			name:     "progress 1.0 returns to",
			progress: 1.0,
			checkFn: func(m ColorModulation) bool {
				return math.Abs(m.HueShift-to.HueShift) < 0.001 &&
					math.Abs(m.SaturationMultiplier-to.SaturationMultiplier) < 0.001
			},
		},
		{
			name:     "progress 0.5 is between",
			progress: 0.5,
			checkFn: func(m ColorModulation) bool {
				return m.HueShift > from.HueShift && m.HueShift < to.HueShift &&
					m.SaturationMultiplier < from.SaturationMultiplier && m.SaturationMultiplier > to.SaturationMultiplier
			},
		},
		{
			name:     "progress < 0 clamped to 0",
			progress: -0.5,
			checkFn: func(m ColorModulation) bool {
				return m.HueShift == from.HueShift
			},
		},
		{
			name:     "progress > 1 clamped to 1",
			progress: 1.5,
			checkFn: func(m ColorModulation) bool {
				return math.Abs(m.HueShift-to.HueShift) < 0.001
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterpolateModulation(from, to, tt.progress)
			if !tt.checkFn(result) {
				t.Errorf("InterpolateModulation() failed check: %+v", result)
			}
		})
	}
}

func TestInterpolateModulation_SmoothStep(t *testing.T) {
	from := ColorModulation{HueShift: 0.0}
	to := ColorModulation{HueShift: 100.0}

	// Test that smooth step provides easing (not linear)
	linear := InterpolateModulation(from, to, 0.5)

	// At 0.5 progress, smooth step should be exactly 0.5
	// (3*0.5² - 2*0.5³ = 3*0.25 - 2*0.125 = 0.75 - 0.25 = 0.5)
	expectedHue := 50.0 // 0 + (100-0) * 0.5
	if math.Abs(linear.HueShift-expectedHue) > 0.001 {
		t.Errorf("At progress 0.5, expected HueShift ~%v, got %v", expectedHue, linear.HueShift)
	}

	// Test that early progress is slower (easing in)
	early := InterpolateModulation(from, to, 0.25)
	// Smooth step at 0.25: 3*0.25² - 2*0.25³ = 0.1875 - 0.03125 = 0.15625
	expectedEarly := 100.0 * 0.15625
	if math.Abs(early.HueShift-expectedEarly) > 0.5 {
		t.Errorf("At progress 0.25, expected HueShift ~%v, got %v", expectedEarly, early.HueShift)
	}
}

func TestGetNextTimeState(t *testing.T) {
	tests := []struct {
		current TimeOfDay
		want    TimeOfDay
	}{
		{TimeOfDayDawn, TimeOfDayDay},
		{TimeOfDayDay, TimeOfDayDusk},
		{TimeOfDayDusk, TimeOfDayNight},
		{TimeOfDayNight, TimeOfDayDawn},
		{TimeOfDay(999), TimeOfDayDay}, // Unknown fallback
	}

	for _, tt := range tests {
		t.Run(tt.current.String(), func(t *testing.T) {
			got := getNextTimeState(tt.current)
			if got != tt.want {
				t.Errorf("getNextTimeState(%v) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}
}

func TestGetModulationWithTransition_NoTransition(t *testing.T) {
	config := TimeConfig{
		CurrentTime:         TimeOfDayDawn,
		TransitionProgress:  0.0,
		IntensityMultiplier: 1.0,
	}

	result := GetModulationWithTransition(config)
	expected := GetModulationForTime(TimeOfDayDawn)

	if result.HueShift != expected.HueShift {
		t.Errorf("Expected HueShift %v, got %v", expected.HueShift, result.HueShift)
	}
}

func TestGetModulationWithTransition_WithTransition(t *testing.T) {
	config := TimeConfig{
		CurrentTime:         TimeOfDayDawn,
		TransitionProgress:  0.5,
		IntensityMultiplier: 1.0,
	}

	result := GetModulationWithTransition(config)
	dawn := GetModulationForTime(TimeOfDayDawn)
	day := GetModulationForTime(TimeOfDayDay)

	// Result should be between dawn and day
	if result.HueShift < day.HueShift || result.HueShift > dawn.HueShift {
		t.Errorf("Expected HueShift between %v and %v, got %v", day.HueShift, dawn.HueShift, result.HueShift)
	}
}

func TestGetModulationWithTransition_IntensityScaling(t *testing.T) {
	fullIntensity := TimeConfig{
		CurrentTime:         TimeOfDayNight,
		TransitionProgress:  0.0,
		IntensityMultiplier: 1.0,
	}
	halfIntensity := TimeConfig{
		CurrentTime:         TimeOfDayNight,
		TransitionProgress:  0.0,
		IntensityMultiplier: 0.5,
	}

	full := GetModulationWithTransition(fullIntensity)
	half := GetModulationWithTransition(halfIntensity)
	day := GetModulationForTime(TimeOfDayDay) // Neutral

	// Half intensity should be closer to day (neutral) than full intensity
	fullDist := math.Abs(full.LightnessOffset - day.LightnessOffset)
	halfDist := math.Abs(half.LightnessOffset - day.LightnessOffset)

	if halfDist >= fullDist {
		t.Errorf("Half intensity should be closer to neutral than full intensity")
	}
}

func TestApplyTimeModulation_NilPalette(t *testing.T) {
	config := DefaultTimeConfig()
	result := ApplyTimeModulation(nil, config)

	if result != nil {
		t.Errorf("ApplyTimeModulation(nil) should return nil, got %v", result)
	}
}

func TestApplyTimeModulation_PreservesStructure(t *testing.T) {
	// Create a simple palette
	palette := &Palette{
		Primary:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Secondary:  color.RGBA{R: 0, G: 255, B: 0, A: 255},
		Background: color.RGBA{R: 100, G: 100, B: 100, A: 255},
		Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Colors:     []color.Color{color.RGBA{R: 128, G: 128, B: 128, A: 255}},
	}

	config := TimeConfig{
		CurrentTime:         TimeOfDayNight,
		TransitionProgress:  0.0,
		IntensityMultiplier: 1.0,
	}

	result := ApplyTimeModulation(palette, config)

	if result == nil {
		t.Fatal("ApplyTimeModulation returned nil")
	}
	if result.Primary == nil {
		t.Error("Primary color should not be nil")
	}
	if result.Secondary == nil {
		t.Error("Secondary color should not be nil")
	}
	if len(result.Colors) != len(palette.Colors) {
		t.Errorf("Colors length = %v, want %v", len(result.Colors), len(palette.Colors))
	}
}

func TestApplyTimeModulation_NightDarkens(t *testing.T) {
	// Create a bright palette
	bright := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	palette := &Palette{
		Primary: bright,
	}

	config := TimeConfig{
		CurrentTime:         TimeOfDayNight,
		TransitionProgress:  0.0,
		IntensityMultiplier: 1.0,
	}

	result := ApplyTimeModulation(palette, config)

	r1, g1, b1, _ := palette.Primary.RGBA()
	r2, g2, b2, _ := result.Primary.RGBA()

	// Night should darken colors
	if r2 >= r1 || g2 >= g1 || b2 >= b1 {
		t.Errorf("Night should darken colors: original (%v,%v,%v), result (%v,%v,%v)",
			r1>>8, g1>>8, b1>>8, r2>>8, g2>>8, b2>>8)
	}
}

func TestModulateColor_PreservesAlpha(t *testing.T) {
	original := color.RGBA{R: 128, G: 128, B: 128, A: 200}
	mod := GetModulationForTime(TimeOfDayDusk)

	result := modulateColor(original, mod)

	_, _, _, a := result.RGBA()
	expectedAlpha := uint32(200 << 8)

	// Allow small rounding error (within 256, which is 1 in 8-bit)
	diff := int32(a) - int32(expectedAlpha)
	if diff < 0 {
		diff = -diff
	}
	if diff > 256 {
		t.Errorf("Alpha should be preserved (within 1 8-bit unit): got %v (8-bit: %v), want %v (8-bit: %v)",
			a, a>>8, expectedAlpha, expectedAlpha>>8)
	}
}

func TestRGBToHSL_Conversion(t *testing.T) {
	tests := []struct {
		name                string
		r, g, b             uint8
		wantH, wantS, wantL float64
		tolerance           float64
	}{
		{"red", 255, 0, 0, 0.0, 1.0, 0.5, 1.0},
		{"green", 0, 255, 0, 120.0, 1.0, 0.5, 1.0},
		{"blue", 0, 0, 255, 240.0, 1.0, 0.5, 1.0},
		{"white", 255, 255, 255, 0.0, 0.0, 1.0, 1.0},
		{"black", 0, 0, 0, 0.0, 0.0, 0.0, 1.0},
		{"gray", 128, 128, 128, 0.0, 0.0, 0.502, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)

			if math.Abs(h-tt.wantH) > tt.tolerance {
				t.Errorf("Hue = %v, want %v (±%v)", h, tt.wantH, tt.tolerance)
			}
			if math.Abs(s-tt.wantS) > tt.tolerance {
				t.Errorf("Saturation = %v, want %v (±%v)", s, tt.wantS, tt.tolerance)
			}
			if math.Abs(l-tt.wantL) > tt.tolerance {
				t.Errorf("Lightness = %v, want %v (±%v)", l, tt.wantL, tt.tolerance)
			}
		})
	}
}

func TestHSLToRGB_Conversion(t *testing.T) {
	tests := []struct {
		name                string
		h, s, l             float64
		wantR, wantG, wantB uint8
		tolerance           uint8
	}{
		{"red", 0.0, 1.0, 0.5, 255, 0, 0, 2},
		{"green", 120.0, 1.0, 0.5, 0, 255, 0, 2},
		{"blue", 240.0, 1.0, 0.5, 0, 0, 255, 2},
		{"white", 0.0, 0.0, 1.0, 255, 255, 255, 2},
		{"black", 0.0, 0.0, 0.0, 0, 0, 0, 2},
		{"gray", 0.0, 0.0, 0.5, 128, 128, 128, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := hslToRGB(tt.h, tt.s, tt.l)

			if absDiffU8(r, tt.wantR) > tt.tolerance {
				t.Errorf("Red = %v, want %v (±%v)", r, tt.wantR, tt.tolerance)
			}
			if absDiffU8(g, tt.wantG) > tt.tolerance {
				t.Errorf("Green = %v, want %v (±%v)", g, tt.wantG, tt.tolerance)
			}
			if absDiffU8(b, tt.wantB) > tt.tolerance {
				t.Errorf("Blue = %v, want %v (±%v)", b, tt.wantB, tt.tolerance)
			}
		})
	}
}

func TestRGBHSLRoundTrip(t *testing.T) {
	tests := []struct {
		r, g, b uint8
	}{
		{255, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{128, 128, 128},
		{200, 100, 50},
		{50, 200, 100},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			h, s, l := rgbToHSL(tt.r, tt.g, tt.b)
			r2, g2, b2 := hslToRGB(h, s, l)

			// Allow small rounding errors
			if absDiffU8(r2, tt.r) > 2 || absDiffU8(g2, tt.g) > 2 || absDiffU8(b2, tt.b) > 2 {
				t.Errorf("Round trip failed: (%v,%v,%v) -> HSL(%v,%v,%v) -> (%v,%v,%v)",
					tt.r, tt.g, tt.b, h, s, l, r2, g2, b2)
			}
		})
	}
}

// Helper function to calculate absolute difference between two uint8 values
func absDiffU8(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// Benchmark tests for performance validation (<1% frame time overhead target)

func BenchmarkGetModulationForTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetModulationForTime(TimeOfDayNight)
	}
}

func BenchmarkInterpolateModulation(b *testing.B) {
	from := GetModulationForTime(TimeOfDayDawn)
	to := GetModulationForTime(TimeOfDayDay)

	for i := 0; i < b.N; i++ {
		_ = InterpolateModulation(from, to, 0.5)
	}
}

func BenchmarkGetModulationWithTransition(b *testing.B) {
	config := TimeConfig{
		CurrentTime:         TimeOfDayDusk,
		TransitionProgress:  0.3,
		IntensityMultiplier: 1.0,
	}

	for i := 0; i < b.N; i++ {
		_ = GetModulationWithTransition(config)
	}
}

func BenchmarkModulateColor(b *testing.B) {
	c := color.RGBA{R: 128, G: 64, B: 192, A: 255}
	mod := GetModulationForTime(TimeOfDayNight)

	for i := 0; i < b.N; i++ {
		_ = modulateColor(c, mod)
	}
}

func BenchmarkApplyTimeModulation(b *testing.B) {
	palette := &Palette{
		Primary:    color.RGBA{R: 255, G: 100, B: 50, A: 255},
		Secondary:  color.RGBA{R: 50, G: 100, B: 255, A: 255},
		Background: color.RGBA{R: 30, G: 30, B: 30, A: 255},
		Text:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Accent1:    color.RGBA{R: 200, G: 50, B: 50, A: 255},
		Accent2:    color.RGBA{R: 50, G: 200, B: 50, A: 255},
		Accent3:    color.RGBA{R: 50, G: 50, B: 200, A: 255},
		Colors:     make([]color.Color, 12),
	}

	config := TimeConfig{
		CurrentTime:         TimeOfDayNight,
		TransitionProgress:  0.5,
		IntensityMultiplier: 1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyTimeModulation(palette, config)
	}
}

func BenchmarkRGBToHSL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = rgbToHSL(128, 64, 192)
	}
}

func BenchmarkHSLToRGB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = hslToRGB(280.0, 0.5, 0.5)
	}
}
