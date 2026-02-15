// Package sprites provides color temperature and specular highlight rendering
// for top-down entity sprites. ApplyColorTemperature shifts the warm/cool
// balance across each body part: pixels on the light-facing side receive a warm
// (golden/amber) shift while shadow-side pixels shift cool (blue/violet). Sharp
// specular highlight spots are placed on the brightest light-facing surface to
// simulate reflective sheen. Both effects are genre-aware and seed-deterministic.
//
// Call this AFTER depth enhancement and BEFORE sprite finalization in the
// rendering pipeline so outlines and rim lighting wrap the color-graded result.
package sprites

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// ColorTemperatureConfig controls the warm/cool color shift and specular highlights.
type ColorTemperatureConfig struct {
	// WarmShift is the maximum RGB warm shift for highlight pixels (0.0-1.0).
	WarmShift float64
	// CoolShift is the maximum RGB cool shift for shadow pixels (0.0-1.0).
	CoolShift float64
	// WarmColor is the color added on the light side.
	WarmColor color.RGBA
	// CoolColor is the color added on the shadow side.
	CoolColor color.RGBA
	// LightAngle is the light source direction in radians (0 = top).
	LightAngle float64
	// SpecularIntensity controls how bright specular highlights are (0.0-1.0).
	SpecularIntensity float64
	// SpecularTightness controls the falloff exponent for specular highlights.
	// Higher values produce smaller, sharper highlights (8-64).
	SpecularTightness float64
	// SpecularColor is the color of specular highlight spots.
	SpecularColor color.RGBA
	// Seed for deterministic variation.
	Seed int64
}

// GenreColorTemperatureConfig returns a genre-tuned color temperature config.
func GenreColorTemperatureConfig(genre string, seed int64) ColorTemperatureConfig {
	cfg := ColorTemperatureConfig{
		WarmShift:         0.12,
		CoolShift:         0.10,
		WarmColor:         color.RGBA{R: 255, G: 220, B: 170, A: 255},
		CoolColor:         color.RGBA{R: 140, G: 160, B: 220, A: 255},
		LightAngle:        5.5, // ~315 degrees = top-left light
		SpecularIntensity: 0.35,
		SpecularTightness: 24.0,
		SpecularColor:     color.RGBA{R: 255, G: 252, B: 245, A: 255},
		Seed:              seed,
	}

	switch genre {
	case "fantasy":
		cfg.WarmShift = 0.15
		cfg.CoolShift = 0.08
		cfg.WarmColor = color.RGBA{R: 255, G: 215, B: 150, A: 255} // golden warmth
		cfg.CoolColor = color.RGBA{R: 130, G: 150, B: 200, A: 255}
		cfg.SpecularIntensity = 0.40
		cfg.SpecularColor = color.RGBA{R: 255, G: 248, B: 220, A: 255}
	case "horror":
		cfg.WarmShift = 0.06
		cfg.CoolShift = 0.16
		cfg.WarmColor = color.RGBA{R: 200, G: 190, B: 170, A: 255} // desaturated warm
		cfg.CoolColor = color.RGBA{R: 100, G: 140, B: 180, A: 255} // cold blue
		cfg.SpecularIntensity = 0.20
		cfg.SpecularTightness = 32.0
		cfg.SpecularColor = color.RGBA{R: 220, G: 230, B: 240, A: 255} // cold highlight
	case "scifi":
		cfg.WarmShift = 0.08
		cfg.CoolShift = 0.12
		cfg.WarmColor = color.RGBA{R: 240, G: 240, B: 235, A: 255} // neutral white
		cfg.CoolColor = color.RGBA{R: 150, G: 170, B: 220, A: 255}
		cfg.SpecularIntensity = 0.45
		cfg.SpecularTightness = 16.0
		cfg.SpecularColor = color.RGBA{R: 240, G: 245, B: 255, A: 255} // cool white
	case "cyberpunk":
		cfg.WarmShift = 0.10
		cfg.CoolShift = 0.14
		cfg.WarmColor = color.RGBA{R: 255, G: 200, B: 180, A: 255}
		cfg.CoolColor = color.RGBA{R: 120, G: 100, B: 200, A: 255} // purple-blue
		cfg.SpecularIntensity = 0.50
		cfg.SpecularTightness = 12.0
		cfg.SpecularColor = color.RGBA{R: 230, G: 220, B: 255, A: 255} // neon tinge
	case "postapoc":
		cfg.WarmShift = 0.10
		cfg.CoolShift = 0.08
		cfg.WarmColor = color.RGBA{R: 230, G: 200, B: 150, A: 255} // dusty warm
		cfg.CoolColor = color.RGBA{R: 140, G: 150, B: 160, A: 255} // muted cool
		cfg.SpecularIntensity = 0.25
		cfg.SpecularTightness = 20.0
		cfg.SpecularColor = color.RGBA{R: 240, G: 235, B: 220, A: 255}
	}

	return cfg
}

// ApplyColorTemperature applies warm/cool color temperature shift and specular
// highlights to a sprite image. Only opaque pixels are modified. The light
// direction determines which side receives warm vs cool tinting, and a sharp
// specular highlight is placed at the point of maximum light reflection.
// Returns the number of pixels modified.
func ApplyColorTemperature(img *image.RGBA, cfg ColorTemperatureConfig) int {
	if img == nil {
		return 0
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return 0
	}

	// Precompute light direction vector
	lightDx := math.Sin(cfg.LightAngle)
	lightDy := -math.Cos(cfg.LightAngle)

	cx := float64(w) / 2.0
	cy := float64(h) / 2.0

	rng := rand.New(rand.NewSource(cfg.Seed))

	// Precompute warm/cool color components as floats
	warmR := float64(cfg.WarmColor.R) / 255.0
	warmG := float64(cfg.WarmColor.G) / 255.0
	warmB := float64(cfg.WarmColor.B) / 255.0
	coolR := float64(cfg.CoolColor.R) / 255.0
	coolG := float64(cfg.CoolColor.G) / 255.0
	coolB := float64(cfg.CoolColor.B) / 255.0
	specR := float64(cfg.SpecularColor.R) / 255.0
	specG := float64(cfg.SpecularColor.G) / 255.0
	specB := float64(cfg.SpecularColor.B) / 255.0

	// Build opacity mask for finding the shape center-of-mass
	var totalWeight float64
	var comX, comY float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			a := img.Pix[idx+3]
			if a > 0 {
				fa := float64(a) / 255.0
				comX += float64(x) * fa
				comY += float64(y) * fa
				totalWeight += fa
			}
		}
	}
	if totalWeight > 0 {
		cx = comX / totalWeight
		cy = comY / totalWeight
	}

	// Find the specular highlight anchor: the opaque pixel closest to pure
	// light-direction offset from center-of-mass.
	specAnchorX := cx + lightDx*float64(w)*0.25
	specAnchorY := cy + lightDy*float64(h)*0.25

	// Add deterministic jitter so specular spots vary per entity
	specAnchorX += (rng.Float64() - 0.5) * 2.0
	specAnchorY += (rng.Float64() - 0.5) * 2.0

	modified := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			a := img.Pix[idx+3]
			if a == 0 {
				continue
			}

			r := float64(img.Pix[idx])
			g := float64(img.Pix[idx+1])
			b := float64(img.Pix[idx+2])

			// Compute the pixel's position relative to center-of-mass
			nx := (float64(x) - cx)
			ny := (float64(y) - cy)
			// Normalize
			dist := math.Sqrt(nx*nx + ny*ny)
			if dist < 0.001 {
				dist = 0.001
			}
			dirX := nx / dist
			dirY := ny / dist

			// Dot product: positive = light-facing (warm), negative = shadow (cool)
			dot := dirX*lightDx + dirY*lightDy

			// Apply color temperature shift
			if dot > 0 {
				// Warm shift — proportional to alignment with light direction
				strength := dot * cfg.WarmShift
				r += strength * warmR * 60.0
				g += strength * warmG * 40.0
				b += strength * warmB * 20.0
			} else {
				// Cool shift — proportional to alignment away from light
				strength := -dot * cfg.CoolShift
				r += strength * coolR * 15.0
				g += strength * coolG * 25.0
				b += strength * coolB * 45.0
			}

			// Specular highlight: sharp bright spot near light anchor
			specDx := float64(x) - specAnchorX
			specDy := float64(y) - specAnchorY
			specDist := math.Sqrt(specDx*specDx+specDy*specDy) / math.Max(float64(w), float64(h))
			// Phong-like specular: (1 - dist)^tightness
			specFactor := math.Pow(math.Max(0, 1.0-specDist*3.0), cfg.SpecularTightness)
			if specFactor > 0.01 {
				specStrength := specFactor * cfg.SpecularIntensity * 255.0
				r += specStrength * specR
				g += specStrength * specG
				b += specStrength * specB
			}

			img.Pix[idx] = clampByte(r)
			img.Pix[idx+1] = clampByte(g)
			img.Pix[idx+2] = clampByte(b)
			modified++
		}
	}

	return modified
}
