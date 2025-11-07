// Package sprites provides animation frame generation for procedural sprites.
package sprites

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// GenerateAnimationFrame creates a single frame of an animation sequence.
// Uses deterministic generation based on seed, state, and frame index.
// CRITICAL: Uses the full sprite generation pipeline (with anatomical templates)
// and applies transformations, rather than generating simple shapes.
func (g *Generator) GenerateAnimationFrame(config Config, state string, frameIndex, frameCount int) (*ebiten.Image, error) {
	// Generate palette if not provided
	if config.Palette == nil {
		pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
		if err != nil {
			return nil, fmt.Errorf("failed to generate palette: %w", err)
		}
		config.Palette = pal
	}

	// CRITICAL FIX: Use the SAME seed for all frames in an animation!
	// Only the animation state affects the seed, NOT the frame index
	// This ensures the sprite looks consistent across all frames
	baseConfig := config
	baseConfig.Seed = config.Seed // Keep seed consistent across frames

	// Apply state-specific transformations (this is what changes between frames)
	offset := calculateAnimationOffset(state, frameIndex, frameCount)
	rotation := calculateAnimationRotation(state, frameIndex, frameCount)
	scale := calculateAnimationScale(state, frameIndex, frameCount)

	// Generate the FULL sprite using the proper generation pipeline
	// This ensures we get anatomical templates, layering, and all visual details
	baseSprite, err := g.Generate(baseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base sprite: %w", err)
	}

	// Create output image with exact config dimensions
	// The image must match the requested width and height exactly
	img := ebiten.NewImage(config.Width, config.Height)

	// Apply transformations to the generated sprite
	opts := &ebiten.DrawImageOptions{}

	// Center sprite in output image
	centerX := float64(config.Width) / 2
	centerY := float64(config.Height) / 2

	// Apply scale
	if scale != 1.0 {
		opts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply rotation around center
	if rotation != 0 {
		opts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		opts.GeoM.Rotate(rotation)
		opts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply position offset and center in output
	opts.GeoM.Translate(centerX-float64(config.Width)/2+offset.X, centerY-float64(config.Height)/2+offset.Y)

	img.DrawImage(baseSprite, opts)

	return img, nil
}

// calculateAnimationOffset computes position offset for animation frame.
func calculateAnimationOffset(state string, frameIndex, frameCount int) struct{ X, Y float64 } {
	t := float64(frameIndex) / float64(frameCount)
	offset := struct{ X, Y float64 }{X: 0, Y: 0}

	switch state {
	case "idle":
		// Phase 15.2: Subtle breathing animation
		// Gentle vertical oscillation with slight horizontal sway
		breathCycle := math.Sin(t * 2 * math.Pi)
		offset.Y = breathCycle * 0.8           // Very subtle 0.8px vertical breathing
		offset.X = math.Sin(t*4*math.Pi) * 0.3 // Even more subtle horizontal sway

	case "walk", "run":
		// Bobbing motion
		cycle := math.Sin(t * 2 * math.Pi)
		offset.Y = cycle * 2.0 // 2 pixel vertical bob

	case "jump":
		// Parabolic arc
		offset.Y = -4.0 * (t - t*t) * 10.0 // Jump up and down

	case "attack":
		// Phase 15.2: Enhanced forward lunge with better follow-through
		// Wind-up (0-0.2), strike (0.2-0.5), follow-through (0.5-1.0)
		if t < 0.2 {
			// Wind-up: slight backward movement
			offset.X = -(t / 0.2) * 2.0
		} else if t < 0.5 {
			// Strike: rapid forward lunge
			strikeT := (t - 0.2) / 0.3
			offset.X = -2.0 + strikeT*18.0 // From -2 to +16 pixels
		} else {
			// Follow-through: gradual return with slight overextension
			followT := (t - 0.5) / 0.5
			offset.X = 16.0 - followT*followT*16.0 // Quadratic easing for smooth return
		}

	case "hit":
		// Knockback
		offset.X = -(1.0 - t) * 3.0 // Move backward and recover

	case "death":
		// Fall down
		offset.Y = t * 8.0 // Move down
	}

	return offset
}

// calculateAnimationRotation computes rotation for animation frame.
func calculateAnimationRotation(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "idle":
		// Phase 15.2: Very subtle head tilt for breathing animation
		return math.Sin(t*2*math.Pi) * 0.03 // Tiny oscillation (0.03 radians ≈ 1.7 degrees)

	case "attack":
		// Phase 15.2: Enhanced swing arc with better follow-through
		// Wind-up (0-0.2), strike (0.2-0.5), follow-through (0.5-1.0)
		if t < 0.2 {
			// Wind up: slight backward rotation
			windupT := t / 0.2
			return -windupT * 0.4 // -0.4 radians (~23 degrees)
		} else if t < 0.5 {
			// Swing through: rapid forward rotation
			strikeT := (t - 0.2) / 0.3
			return -0.4 + strikeT*1.8 // From -0.4 to +1.4 radians
		} else {
			// Follow through: continued rotation with deceleration
			followT := (t - 0.5) / 0.5
			// Use sine easing for smooth deceleration
			easedT := math.Sin(followT * math.Pi / 2)
			return 1.4 - easedT*0.8 // From +1.4 to +0.6 radians, smooth follow-through
		}

	case "death":
		// Rotate while falling
		return t * math.Pi / 2 // 90 degree rotation

	case "cast":
		// Gentle sway
		return math.Sin(t*2*math.Pi) * 0.1
	}

	return 0
}

// calculateAnimationScale computes scale factor for animation frame.
func calculateAnimationScale(state string, frameIndex, frameCount int) float64 {
	t := float64(frameIndex) / float64(frameCount)

	switch state {
	case "idle":
		// Phase 15.2: Subtle breathing scale (chest expansion/contraction)
		breathCycle := math.Sin(t * 2 * math.Pi)
		return 1.0 + breathCycle*0.015 // Very subtle 1.5% scale change

	case "jump":
		// Squash and stretch
		if t < 0.2 {
			return 1.0 - t*0.5 // Squash before jump
		} else if t < 0.8 {
			return 0.9 + (t-0.2)*0.3 // Stretch during jump
		} else {
			return 1.0 - (t-0.8)*0.5 // Squash on landing
		}

	case "hit":
		// Squash on impact
		return 1.0 - t*0.2

	case "attack":
		// Phase 15.2: Enhanced anticipation and follow-through scale
		// Slight anticipation squat, then expansion during strike
		if t < 0.2 {
			// Anticipation: slight compression
			return 1.0 - (t/0.2)*0.05
		} else if t < 0.5 {
			// Strike: expand for power
			strikeT := (t - 0.2) / 0.3
			return 0.95 + strikeT*0.15 // From 0.95 to 1.10
		} else {
			// Follow-through: gradual return to normal
			followT := (t - 0.5) / 0.5
			return 1.10 - followT*0.10 // From 1.10 to 1.00
		}
	}

	return 1.0
}

// addAnimationDetails adds additional visual details to animation frames.
func (g *Generator) addAnimationDetails(img *ebiten.Image, config Config, rng *rand.Rand, frameIndex, frameCount int) {
	// Add particle effects for certain states
	t := float64(frameIndex) / float64(frameCount)

	// Example: Add motion blur effect for fast movements
	if rng.Float64() < config.Complexity {
		detailConfig := shapes.Config{
			Type:      shapes.ShapeCircle,
			Width:     2 + rng.Intn(3),
			Height:    2 + rng.Intn(3),
			Color:     config.Palette.Accent1,
			Seed:      config.Seed + int64(frameIndex),
			Smoothing: 0.5,
		}

		detail, err := g.shapeGen.Generate(detailConfig)
		if err == nil {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(
				float64(rng.Intn(config.Width)),
				float64(rng.Intn(config.Height)),
			)
			opts.ColorScale.ScaleAlpha(float32(0.3 + t*0.3))
			img.DrawImage(detail, opts)
		}
	}
}

// hashString computes a simple hash of a string for seed derivation.
func hashString(s string) int64 {
	var hash int64
	for i, c := range s {
		hash += int64(c) * int64(i+1)
	}
	return hash
}
