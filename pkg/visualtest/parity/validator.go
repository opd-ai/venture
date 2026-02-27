package parity

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// ParityResult represents the result of a parity test
type ParityResult struct {
	TestName      string
	Platform      Platform
	Passed        bool
	Errors        []string
	Warnings      []string
	Metrics       map[string]float64
	VisualDiffPct float64 // Percentage of visual difference (0-100%)
}

// ParityValidator validates visual consistency across platforms
type Validator struct {
	platform  Platform
	baselines map[string]*Baseline
	tolerance ColorTolerance
}

// Baseline stores reference data for comparison
type Baseline struct {
	Platform     Platform
	ImageData    *image.RGBA
	ColorProfile []color.RGBA
	FrameRate    float64
	Resolution   image.Point
}

// ColorTolerance defines acceptable color difference thresholds
type ColorTolerance struct {
	MaxRGBDelta    uint8   // Max allowed difference per RGB channel (default: 1)
	MaxPercentDiff float64 // Max allowed percentage of differing pixels (default: 5.0%)
}

// DefaultColorTolerance returns the standard tolerance for parity testing
func DefaultColorTolerance() ColorTolerance {
	return ColorTolerance{
		MaxRGBDelta:    1,
		MaxPercentDiff: 5.0,
	}
}

// NewValidator creates a new parity validator for the current platform
func NewValidator() *Validator {
	return &Validator{
		platform:  DetectPlatform(),
		baselines: make(map[string]*Baseline),
		tolerance: DefaultColorTolerance(),
	}
}

// SetTolerance updates the color tolerance settings
func (v *Validator) SetTolerance(tolerance ColorTolerance) {
	v.tolerance = tolerance
}

// SetBaseline stores a baseline for comparison
func (v *Validator) SetBaseline(testName string, baseline *Baseline) {
	v.baselines[testName] = baseline
}

// GetBaseline retrieves a stored baseline
func (v *Validator) GetBaseline(testName string) (*Baseline, bool) {
	baseline, exists := v.baselines[testName]
	return baseline, exists
}

// CompareImages compares two images and returns the percentage difference
func (v *Validator) CompareImages(img1, img2 *image.RGBA) (float64, []string) {
	if img1 == nil || img2 == nil {
		return 100.0, []string{"one or both images are nil"}
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1 != bounds2 {
		return 100.0, []string{
			fmt.Sprintf("image dimensions differ: %v vs %v", bounds1, bounds2),
		}
	}

	totalPixels := bounds1.Dx() * bounds1.Dy()
	differentPixels := 0
	errors := []string{}

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		for x := bounds1.Min.X; x < bounds1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			// Convert from 16-bit to 8-bit
			r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
			r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

			deltaR := absDiff(uint8(r1), uint8(r2))
			deltaG := absDiff(uint8(g1), uint8(g2))
			deltaB := absDiff(uint8(b1), uint8(b2))
			deltaA := absDiff(uint8(a1), uint8(a2))

			// Count any pixel that differs (even by 1)
			if deltaR > 0 || deltaG > 0 || deltaB > 0 || deltaA > 0 {
				differentPixels++
			}
		}
	}

	var diffPct float64
	if totalPixels > 0 {
		diffPct = float64(differentPixels) / float64(totalPixels) * 100.0
	}

	// Report error if too many pixels differ, regardless of how much they differ
	if diffPct > v.tolerance.MaxPercentDiff {
		errors = append(errors, fmt.Sprintf(
			"%.2f%% of pixels differ (threshold: %.2f%%)",
			diffPct, v.tolerance.MaxPercentDiff,
		))
	}

	return diffPct, errors
}

// CompareColors compares two color profiles and returns differences
func (v *Validator) CompareColors(colors1, colors2 []color.RGBA) (float64, []string) {
	if len(colors1) != len(colors2) {
		return 100.0, []string{
			fmt.Sprintf("color profile lengths differ: %d vs %d", len(colors1), len(colors2)),
		}
	}

	if len(colors1) == 0 {
		// Empty color profiles are considered identical (0.0% difference).
		// This is expected behavior: no colors to compare means no difference.
		return 0.0, nil
	}

	totalDiff := 0.0
	errors := []string{}

	for i := 0; i < len(colors1); i++ {
		c1 := colors1[i]
		c2 := colors2[i]

		deltaR := absDiff(c1.R, c2.R)
		deltaG := absDiff(c1.G, c2.G)
		deltaB := absDiff(c1.B, c2.B)
		deltaA := absDiff(c1.A, c2.A)

		maxDelta := maxUint8(deltaR, deltaG, deltaB, deltaA)
		if maxDelta > v.tolerance.MaxRGBDelta {
			totalDiff += float64(maxDelta)
			errors = append(errors, fmt.Sprintf(
				"color[%d]: RGBA(%d,%d,%d,%d) vs RGBA(%d,%d,%d,%d), delta=%d",
				i, c1.R, c1.G, c1.B, c1.A, c2.R, c2.G, c2.B, c2.A, maxDelta,
			))
		}
	}

	avgDiff := totalDiff / float64(len(colors1))
	diffPct := (avgDiff / 255.0) * 100.0

	return diffPct, errors
}

// CompareFrameRate compares frame rates and returns the percentage difference.
// The first return value is the difference percentage, the second contains error
// messages if the difference meets or exceeds the 20% threshold.
func (v *Validator) CompareFrameRate(fps1, fps2 float64) (float64, []string) {
	if fps1 <= 0 || fps2 <= 0 {
		return 100.0, []string{"invalid frame rate values"}
	}

	diff := math.Abs(fps1-fps2) / math.Max(fps1, fps2) * 100.0
	errors := []string{}

	// Allow up to 20% variance per Phase 63.3 requirements (strictly less than 20%)
	if diff >= 20.0 {
		errors = append(errors, fmt.Sprintf(
			"frame rate difference %.2f%% meets or exceeds 20%% threshold (%.2f vs %.2f FPS)",
			diff, fps1, fps2,
		))
	}

	return diff, errors
}

// ValidateSprites validates that sprite rendering is identical across platforms
func (v *Validator) ValidateSprites(testImage *image.RGBA) ParityResult {
	result := ParityResult{
		TestName: "sprite_rendering",
		Platform: v.platform,
		Passed:   true,
		Metrics:  make(map[string]float64),
	}

	baseline, exists := v.baselines["sprite_rendering"]
	if !exists {
		result.Warnings = append(result.Warnings, "no baseline available for comparison")
		return result
	}

	diffPct, errors := v.CompareImages(baseline.ImageData, testImage)
	result.VisualDiffPct = diffPct
	result.Metrics["visual_diff_pct"] = diffPct

	if len(errors) > 0 {
		result.Passed = false
		result.Errors = errors
	}

	return result
}

// ValidateColors validates color accuracy across platforms
func (v *Validator) ValidateColors(colorProfile []color.RGBA) ParityResult {
	result := ParityResult{
		TestName: "color_accuracy",
		Platform: v.platform,
		Passed:   true,
		Metrics:  make(map[string]float64),
	}

	baseline, exists := v.baselines["color_accuracy"]
	if !exists {
		result.Warnings = append(result.Warnings, "no baseline available for comparison")
		return result
	}

	diffPct, errors := v.CompareColors(baseline.ColorProfile, colorProfile)
	result.VisualDiffPct = diffPct
	result.Metrics["color_diff_pct"] = diffPct

	if len(errors) > 0 {
		result.Passed = false
		result.Errors = errors
	}

	return result
}

// ValidateFrameRate validates that frame rate targets are met
func (v *Validator) ValidateFrameRate(currentFPS float64) ParityResult {
	result := ParityResult{
		TestName: "frame_rate",
		Platform: v.platform,
		Passed:   true,
		Metrics:  make(map[string]float64),
	}

	result.Metrics["current_fps"] = currentFPS

	// Target: 60 FPS minimum per Phase 63.3
	if currentFPS < 60.0 {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"frame rate %.2f FPS below 60 FPS target",
			currentFPS,
		))
	}

	baseline, exists := v.baselines["frame_rate"]
	if exists {
		diffPct, errors := v.CompareFrameRate(baseline.FrameRate, currentFPS)
		result.VisualDiffPct = diffPct
		result.Metrics["fps_diff_pct"] = diffPct

		if len(errors) > 0 {
			result.Warnings = append(result.Warnings, errors...)
		}
	}

	return result
}

// ValidateResolution validates that resolution scaling works correctly
func (v *Validator) ValidateResolution(currentRes image.Point) ParityResult {
	result := ParityResult{
		TestName: "resolution_scaling",
		Platform: v.platform,
		Passed:   true,
		Metrics:  make(map[string]float64),
	}

	result.Metrics["width"] = float64(currentRes.X)
	result.Metrics["height"] = float64(currentRes.Y)

	// Validate resolution is supported
	supportedResolutions := []image.Point{
		{1280, 720},
		{1920, 1080},
		{2560, 1440},
		{3840, 2160},
	}

	supported := false
	for _, res := range supportedResolutions {
		if currentRes == res {
			supported = true
			break
		}
	}

	if !supported {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"resolution %dx%d not in standard list",
			currentRes.X, currentRes.Y,
		))
	}

	return result
}

// Helper functions

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxUint8(values ...uint8) uint8 {
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
