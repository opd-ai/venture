package sprites

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// SilhouetteAnalysis contains metrics about a sprite's readability and visual clarity.
type SilhouetteAnalysis struct {
	// Compactness measures how compact the shape is (0.0-1.0, higher is better)
	// Calculated as 4π * area / perimeter^2
	Compactness float64

	// Coverage measures the percentage of sprite canvas used (0.0-1.0)
	Coverage float64

	// EdgeClarity measures the distinctness of the sprite's edge (0.0-1.0)
	// Higher values indicate clearer boundaries
	EdgeClarity float64

	// OverallScore is a weighted combination of all metrics (0.0-1.0)
	OverallScore float64

	// OpaquePixels is the count of non-transparent pixels
	OpaquePixels int

	// PerimeterPixels is the count of edge pixels
	PerimeterPixels int

	// TotalPixels is the total sprite canvas size
	TotalPixels int
}

// AnalyzeSilhouette analyzes a sprite's visual readability and returns metrics.
// This helps determine if a sprite has a clear, recognizable silhouette.
func AnalyzeSilhouette(sprite *ebiten.Image) SilhouetteAnalysis {
	if sprite == nil {
		return SilhouetteAnalysis{}
	}

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	totalPixels := width * height

	if totalPixels == 0 {
		return SilhouetteAnalysis{TotalPixels: totalPixels}
	}

	// Count opaque pixels and perimeter
	opaquePixels := 0
	perimeterPixels := 0
	edgeContrast := 0.0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			_, _, _, a := sprite.At(x, y).RGBA()
			alpha := float64(a) / 65535.0

			if alpha > 0.5 { // Semi-transparent threshold
				opaquePixels++

				// Check if on perimeter (has transparent neighbor)
				if isOnPerimeter(sprite, x, y) {
					perimeterPixels++
					// Measure edge contrast
					contrast := measureEdgeContrast(sprite, x, y)
					edgeContrast += contrast
				}
			}
		}
	}

	// Calculate metrics
	coverage := float64(opaquePixels) / float64(totalPixels)

	var compactness float64
	if perimeterPixels > 0 && opaquePixels > 0 {
		// Compactness: 4π * area / perimeter^2
		// Circle = 1.0, irregular shapes < 1.0
		area := float64(opaquePixels)
		perimeter := float64(perimeterPixels)
		compactness = (4.0 * math.Pi * area) / (perimeter * perimeter)
		if compactness > 1.0 {
			compactness = 1.0
		}
	}

	var edgeClarity float64
	if perimeterPixels > 0 {
		edgeClarity = edgeContrast / float64(perimeterPixels)
	}

	// Overall score: weighted combination
	// 40% coverage, 30% compactness, 30% edge clarity
	overallScore := (0.4 * coverage) + (0.3 * compactness) + (0.3 * edgeClarity)

	return SilhouetteAnalysis{
		Compactness:     compactness,
		Coverage:        coverage,
		EdgeClarity:     edgeClarity,
		OverallScore:    overallScore,
		OpaquePixels:    opaquePixels,
		PerimeterPixels: perimeterPixels,
		TotalPixels:     totalPixels,
	}
}

// isOnPerimeter checks if a pixel is on the sprite's edge (has transparent neighbor).
func isOnPerimeter(img *ebiten.Image, x, y int) bool {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Check 4-connected neighbors
	neighbors := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, n := range neighbors {
		nx, ny := x+n[0], y+n[1]

		// Edge of image
		if nx < 0 || ny < 0 || nx >= width || ny >= height {
			return true
		}

		// Has transparent neighbor
		_, _, _, a := img.At(nx, ny).RGBA()
		if a < 128*257 { // Semi-transparent threshold
			return true
		}
	}
	return false
}

// measureEdgeContrast measures the contrast between a pixel and its neighbors.
// Returns a value 0.0-1.0 indicating how distinct the edge is.
func measureEdgeContrast(img *ebiten.Image, x, y int) float64 {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	r1, g1, b1, a1 := img.At(x, y).RGBA()
	if a1 < 128*257 {
		return 0.0
	}

	// Calculate luminance of current pixel
	lum1 := (0.299*float64(r1) + 0.587*float64(g1) + 0.114*float64(b1)) / 65535.0

	maxContrast := 0.0

	// Check 8-connected neighbors for maximum contrast
	neighbors := [][2]int{
		{-1, -1},
		{0, -1},
		{1, -1},
		{-1, 0},
		{1, 0},
		{-1, 1},
		{0, 1},
		{1, 1},
	}

	for _, n := range neighbors {
		nx, ny := x+n[0], y+n[1]
		if nx < 0 || ny < 0 || nx >= width || ny >= height {
			continue
		}

		r2, g2, b2, a2 := img.At(nx, ny).RGBA()
		if a2 < 128*257 {
			// Transparent neighbor = maximum contrast
			maxContrast = 1.0
			continue
		}

		// Calculate luminance difference
		lum2 := (0.299*float64(r2) + 0.587*float64(g2) + 0.114*float64(b2)) / 65535.0
		contrast := math.Abs(lum1 - lum2)

		if contrast > maxContrast {
			maxContrast = contrast
		}
	}

	return maxContrast
}

// GenerateSilhouette creates a black silhouette version of a sprite.
// Useful for testing shape recognition without color.
func GenerateSilhouette(sprite *ebiten.Image) *ebiten.Image {
	if sprite == nil {
		return nil
	}

	bounds := sprite.Bounds()
	silhouette := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	// Create black silhouette
	blackPixels := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	blackPixels.Fill(color.RGBA{0, 0, 0, 255})

	// Copy alpha channel from original sprite
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a := sprite.At(x, y).RGBA()
			if a > 128*257 {
				silhouette.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	return silhouette
}

// AddOutline adds a dark outline around a sprite to improve visibility.
// outlineColor specifies the outline color (typically dark gray or black).
// thickness specifies the outline width in pixels (1-2 recommended).
func AddOutline(sprite *ebiten.Image, outlineColor color.Color, thickness int) *ebiten.Image {
	if sprite == nil || thickness <= 0 {
		return sprite
	}

	bounds := sprite.Bounds()
	outlined := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	// Draw outline first (behind sprite)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			_, _, _, a := sprite.At(x, y).RGBA()
			if a > 128*257 {
				// Draw outline around this pixel
				for dy := -thickness; dy <= thickness; dy++ {
					for dx := -thickness; dx <= thickness; dx++ {
						if dx == 0 && dy == 0 {
							continue
						}
						ox, oy := x+dx, y+dy
						if ox >= 0 && oy >= 0 && ox < bounds.Dx() && oy < bounds.Dy() {
							// Only draw outline where sprite is transparent
							_, _, _, oa := sprite.At(ox, oy).RGBA()
							if oa < 128*257 {
								outlined.Set(ox, oy, outlineColor)
							}
						}
					}
				}
			}
		}
	}

	// Draw original sprite on top of outline
	op := &ebiten.DrawImageOptions{}
	outlined.DrawImage(sprite, op)

	return outlined
}

// ValidateContrast checks if a sprite has sufficient contrast between body parts.
// Returns true if the sprite meets minimum contrast requirements.
func ValidateContrast(sprite *ebiten.Image, minLuminanceDiff float64) bool {
	if sprite == nil || minLuminanceDiff <= 0 {
		return true
	}

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Sample luminance values across the sprite
	var luminances []float64

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := sprite.At(x, y).RGBA()
			if a < 128*257 {
				continue
			}

			// Calculate luminance (perceived brightness)
			lum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			luminances = append(luminances, lum)
		}
	}

	if len(luminances) < 2 {
		return true
	}

	// Find min and max luminance
	minLum, maxLum := luminances[0], luminances[0]
	for _, lum := range luminances {
		if lum < minLum {
			minLum = lum
		}
		if lum > maxLum {
			maxLum = lum
		}
	}

	// Check if difference meets threshold
	diff := maxLum - minLum
	return diff >= minLuminanceDiff
}

// TestOnBackground composites a sprite onto a background color and returns the result.
// Useful for testing sprite visibility on different terrain types.
func TestOnBackground(sprite *ebiten.Image, bgColor color.Color) *ebiten.Image {
	if sprite == nil {
		return nil
	}

	bounds := sprite.Bounds()
	composite := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	// Fill with background
	composite.Fill(bgColor)

	// Draw sprite on top
	op := &ebiten.DrawImageOptions{}
	composite.DrawImage(sprite, op)

	return composite
}

// ContrastScore calculates how well a sprite stands out against a background.
// Returns a score 0.0-1.0, higher = better visibility.
func ContrastScore(sprite *ebiten.Image, bgColor color.Color) float64 {
	if sprite == nil {
		return 0.0
	}

	bounds := sprite.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate background luminance
	bgR, bgG, bgB, _ := bgColor.RGBA()
	bgLum := (0.299*float64(bgR) + 0.587*float64(bgG) + 0.114*float64(bgB)) / 65535.0

	// Calculate average contrast of sprite against background
	totalContrast := 0.0
	opaquePixels := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := sprite.At(x, y).RGBA()
			if a < 128*257 {
				continue
			}

			spriteLum := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			contrast := math.Abs(spriteLum - bgLum)
			totalContrast += contrast
			opaquePixels++
		}
	}

	if opaquePixels == 0 {
		return 0.0
	}

	return totalContrast / float64(opaquePixels)
}

// OutlineConfig contains parameters for outline generation.
type OutlineConfig struct {
	Color     color.Color
	Thickness int
	Enabled   bool
}

// DefaultOutlineConfig returns sensible defaults for sprite outlines.
func DefaultOutlineConfig() OutlineConfig {
	return OutlineConfig{
		Color:     color.RGBA{20, 20, 20, 255}, // Dark gray
		Thickness: 1,
		Enabled:   true,
	}
}

// SilhouetteQuality categorizes silhouette score into quality levels.
type SilhouetteQuality int

const (
	// QualityPoor indicates poor silhouette (score < 0.4)
	QualityPoor SilhouetteQuality = iota
	// QualityFair indicates fair silhouette (score 0.4-0.6)
	QualityFair
	// QualityGood indicates good silhouette (score 0.6-0.8)
	QualityGood
	// QualityExcellent indicates excellent silhouette (score > 0.8)
	QualityExcellent
)

// String returns the string representation of silhouette quality.
func (q SilhouetteQuality) String() string {
	switch q {
	case QualityPoor:
		return "poor"
	case QualityFair:
		return "fair"
	case QualityGood:
		return "good"
	case QualityExcellent:
		return "excellent"
	default:
		return "unknown"
	}
}

// GetQuality categorizes the analysis into a quality level.
func (a SilhouetteAnalysis) GetQuality() SilhouetteQuality {
	score := a.OverallScore
	if score < 0.4 {
		return QualityPoor
	} else if score < 0.6 {
		return QualityFair
	} else if score < 0.8 {
		return QualityGood
	}
	return QualityExcellent
}

// NeedsImprovement returns true if the silhouette quality is below acceptable standards.
func (a SilhouetteAnalysis) NeedsImprovement() bool {
	return a.GetQuality() < QualityGood
}
