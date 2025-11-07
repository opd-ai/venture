// Package lighting provides dynamic lighting effects for rendered scenes.
package lighting

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// AOConfig contains configuration for ambient occlusion.
type AOConfig struct {
	// Enabled toggles ambient occlusion
	Enabled bool

	// Intensity controls darkening strength (0.0-1.0)
	Intensity float64

	// Radius is the occlusion sampling radius in pixels (4-32 typical)
	Radius int

	// Samples is the number of directional samples (4-32 typical)
	// Higher values provide better quality but cost more performance
	Samples int

	// Bias prevents self-occlusion artifacts (0.0-0.1 typical)
	Bias float64

	// Power controls occlusion falloff curve (1.0-4.0 typical)
	// Higher values create sharper, more pronounced occlusion
	Power float64

	// Seed for deterministic random sampling
	Seed int64
}

// DefaultAOConfig returns a sensible default AO configuration.
func DefaultAOConfig() AOConfig {
	return AOConfig{
		Enabled:   true,
		Intensity: 0.5,
		Radius:    16,
		Samples:   16,
		Bias:      0.02,
		Power:     2.0,
		Seed:      12345,
	}
}

// ApplyAmbientOcclusion applies screen-space ambient occlusion to an image.
// Uses a simplified SSAO algorithm with random sampling for efficiency.
// The depthMap should be a grayscale image where brightness indicates depth
// (brighter = further from camera). If depthMap is nil, uses luminance as depth.
func ApplyAmbientOcclusion(img *image.RGBA, depthMap *image.RGBA, config AOConfig) *image.RGBA {
	if !config.Enabled || config.Intensity <= 0 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Generate depth map if not provided
	if depthMap == nil {
		depthMap = generateDepthFromLuminance(img)
	}

	// Pre-generate sampling directions for deterministic results
	rng := rand.New(rand.NewSource(config.Seed))
	sampleDirs := generateSampleDirections(config.Samples, rng)

	// Process each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get pixel depth
			dr, _, _, _ := depthMap.At(x, y).RGBA()
			depth := float64(dr) / 65535.0

			// Calculate occlusion
			occlusion := calculateOcclusion(x, y, depth, depthMap, config, sampleDirs)

			// Apply occlusion to pixel
			r, g, b, a := img.At(x, y).RGBA()
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0

			// Darken based on occlusion
			factor := 1.0 - (occlusion * config.Intensity)
			result.Set(x, y, color.RGBA{
				R: uint8(rf * factor * 255),
				G: uint8(gf * factor * 255),
				B: uint8(bf * factor * 255),
				A: uint8(float64(a) / 257),
			})
		}
	}

	return result
}

// calculateOcclusion samples surrounding pixels to determine occlusion factor.
func calculateOcclusion(x, y int, depth float64, depthMap *image.RGBA, config AOConfig, sampleDirs [][2]float64) float64 {
	bounds := depthMap.Bounds()
	var occlusionSum float64
	validSamples := 0

	for _, dir := range sampleDirs {
		// Calculate sample position
		sampleX := x + int(dir[0]*float64(config.Radius))
		sampleY := y + int(dir[1]*float64(config.Radius))

		// Check bounds
		if sampleX < bounds.Min.X || sampleX >= bounds.Max.X ||
			sampleY < bounds.Min.Y || sampleY >= bounds.Max.Y {
			continue
		}

		// Get sample depth
		sr, _, _, _ := depthMap.At(sampleX, sampleY).RGBA()
		sampleDepth := float64(sr) / 65535.0

		// Calculate depth difference
		depthDiff := depth - sampleDepth

		// Apply bias to prevent self-occlusion
		if depthDiff > config.Bias {
			// Calculate occlusion contribution with falloff
			distance := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1])
			falloff := 1.0 - clamp(distance/float64(config.Radius), 0, 1)

			// Apply power curve for sharper occlusion
			occlusion := math.Pow(clamp(depthDiff, 0, 1)*falloff, config.Power)
			occlusionSum += occlusion
			validSamples++
		}
	}

	// Average occlusion
	if validSamples > 0 {
		return occlusionSum / float64(validSamples)
	}
	return 0
}

// generateSampleDirections creates a set of evenly distributed sample directions.
// Uses stratified random sampling for better coverage than pure random.
func generateSampleDirections(count int, rng *rand.Rand) [][2]float64 {
	directions := make([][2]float64, count)

	// Generate samples in a stratified grid for better distribution
	sqrtCount := int(math.Sqrt(float64(count)))
	for i := 0; i < count; i++ {
		// Stratified cell
		cellX := i % sqrtCount
		cellY := i / sqrtCount

		// Random jitter within cell
		jitterX := (float64(cellX) + rng.Float64()) / float64(sqrtCount)
		jitterY := (float64(cellY) + rng.Float64()) / float64(sqrtCount)

		// Convert to angle
		angle := 2 * math.Pi * (float64(i) + rng.Float64()) / float64(count)

		// Generate direction with stratified radius
		radius := math.Sqrt(jitterX*jitterX + jitterY*jitterY)
		directions[i] = [2]float64{
			math.Cos(angle) * radius,
			math.Sin(angle) * radius,
		}
	}

	return directions
}

// generateDepthFromLuminance creates a depth map from image luminance.
// Darker areas are considered closer (more depth), lighter areas further.
func generateDepthFromLuminance(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	depthMap := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// Calculate luminance
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0
			lum := 0.2126*rf + 0.7152*gf + 0.0722*bf

			// Invert so dark = near, bright = far
			depth := 1.0 - lum

			d := uint8(depth * 255)
			depthMap.Set(x, y, color.RGBA{R: d, G: d, B: d, A: uint8(float64(a) / 257)})
		}
	}

	return depthMap
}

// EnhancedAOConfig provides quality presets for ambient occlusion.
type EnhancedAOConfig struct {
	AOConfig

	// CornerIntensity provides extra darkening at corners (0.0-1.0)
	CornerIntensity float64

	// EdgeIntensity provides extra darkening at edges (0.0-1.0)
	EdgeIntensity float64
}

// DefaultEnhancedAOConfig returns enhanced AO with corner/edge detection.
func DefaultEnhancedAOConfig() EnhancedAOConfig {
	return EnhancedAOConfig{
		AOConfig:        DefaultAOConfig(),
		CornerIntensity: 0.3,
		EdgeIntensity:   0.2,
	}
}

// ApplyEnhancedAO applies ambient occlusion with corner and edge enhancement.
func ApplyEnhancedAO(img *image.RGBA, depthMap *image.RGBA, config EnhancedAOConfig) *image.RGBA {
	// Apply base AO
	result := ApplyAmbientOcclusion(img, depthMap, config.AOConfig)

	// Enhance corners if configured
	if config.CornerIntensity > 0 {
		result = enhanceCorners(result, depthMap, config.CornerIntensity)
	}

	// Enhance edges if configured
	if config.EdgeIntensity > 0 {
		result = enhanceEdges(result, depthMap, config.EdgeIntensity)
	}

	return result
}

// enhanceCorners adds extra darkening at corners using convexity detection.
func enhanceCorners(img *image.RGBA, depthMap *image.RGBA, intensity float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Use depth map if available, otherwise use luminance
	if depthMap == nil {
		depthMap = generateDepthFromLuminance(img)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Detect corners using 3x3 depth gradient
			cornerFactor := detectCorner(x, y, depthMap)

			// Apply corner darkening
			r, g, b, a := img.At(x, y).RGBA()
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0

			darken := 1.0 - (cornerFactor * intensity)
			result.Set(x, y, color.RGBA{
				R: uint8(rf * darken * 255),
				G: uint8(gf * darken * 255),
				B: uint8(bf * darken * 255),
				A: uint8(float64(a) / 257),
			})
		}
	}

	return result
}

// enhanceEdges adds extra darkening at edges using edge detection.
func enhanceEdges(img *image.RGBA, depthMap *image.RGBA, intensity float64) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	if depthMap == nil {
		depthMap = generateDepthFromLuminance(img)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Detect edges using Sobel operator
			edgeFactor := detectEdge(x, y, depthMap)

			// Apply edge darkening
			r, g, b, a := img.At(x, y).RGBA()
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0

			darken := 1.0 - (edgeFactor * intensity)
			result.Set(x, y, color.RGBA{
				R: uint8(rf * darken * 255),
				G: uint8(gf * darken * 255),
				B: uint8(bf * darken * 255),
				A: uint8(float64(a) / 257),
			})
		}
	}

	return result
}

// detectCorner uses 3x3 convexity detection to find corners.
func detectCorner(x, y int, depthMap *image.RGBA) float64 {
	bounds := depthMap.Bounds()

	// Get center depth
	cr, _, _, _ := depthMap.At(x, y).RGBA()
	center := float64(cr) / 65535.0

	var higherCount int
	var totalDiff float64

	// Check 8 neighbors
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := x+dx, y+dy
			if nx < bounds.Min.X || nx >= bounds.Max.X || ny < bounds.Min.Y || ny >= bounds.Max.Y {
				continue
			}

			nr, _, _, _ := depthMap.At(nx, ny).RGBA()
			neighbor := float64(nr) / 65535.0

			diff := neighbor - center
			if diff > 0.1 { // Threshold for significant depth difference
				higherCount++
				totalDiff += diff
			}
		}
	}

	// Convex corners have many neighbors at higher depth
	if higherCount >= 5 { // 5 or more neighbors higher
		return clamp(totalDiff/float64(higherCount), 0, 1)
	}

	return 0
}

// detectEdge uses Sobel edge detection on depth map.
func detectEdge(x, y int, depthMap *image.RGBA) float64 {
	bounds := depthMap.Bounds()

	// Sobel kernels
	sobelX := [3][3]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	sobelY := [3][3]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	var gx, gy float64

	// Apply Sobel operator
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx < bounds.Min.X || nx >= bounds.Max.X || ny < bounds.Min.Y || ny >= bounds.Max.Y {
				continue
			}

			r, _, _, _ := depthMap.At(nx, ny).RGBA()
			depth := float64(r) / 65535.0

			gx += depth * sobelX[dy+1][dx+1]
			gy += depth * sobelY[dy+1][dx+1]
		}
	}

	// Calculate gradient magnitude
	magnitude := math.Sqrt(gx*gx + gy*gy)
	return clamp(magnitude, 0, 1)
}
