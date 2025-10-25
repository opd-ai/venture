//go:build !test
// +build !test

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
func (g *Generator) GenerateAnimationFrame(config Config, state string, frameIndex, frameCount int) (*ebiten.Image, error) {
	// Generate palette if not provided
	if config.Palette == nil {
		pal, err := g.paletteGen.Generate(config.GenreID, config.Seed)
		if err != nil {
			return nil, fmt.Errorf("failed to generate palette: %w", err)
		}
		config.Palette = pal
	}

	// Create frame-specific seed
	frameSeed := config.Seed + int64(frameIndex) + hashString(state)
	rng := rand.New(rand.NewSource(frameSeed))

	// Apply state-specific transformations
	offset := calculateAnimationOffset(state, frameIndex, frameCount)
	rotation := calculateAnimationRotation(state, frameIndex, frameCount)
	scale := calculateAnimationScale(state, frameIndex, frameCount)

	// Generate base sprite
	baseConfig := config
	baseConfig.Seed = frameSeed
	img := ebiten.NewImage(config.Width, config.Height)

	// Generate body with transformations
	bodyConfig := shapes.Config{
		Type:      shapes.ShapeType(rng.Intn(3)), // Circle, Rectangle, Triangle
		Width:     int(float64(config.Width) * 0.7 * scale),
		Height:    int(float64(config.Height) * 0.7 * scale),
		Color:     config.Palette.Primary,
		Seed:      frameSeed,
		Smoothing: 0.2,
	}

	bodyShape, err := g.shapeGen.Generate(bodyConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate body shape: %w", err)
	}

	// Draw body with offset and rotation
	opts := &ebiten.DrawImageOptions{}

	// Apply rotation if needed
	if rotation != 0 {
		opts.GeoM.Translate(-float64(config.Width)/2, -float64(config.Height)/2)
		opts.GeoM.Rotate(rotation)
		opts.GeoM.Translate(float64(config.Width)/2, float64(config.Height)/2)
	}

	// Apply position offset
	opts.GeoM.Translate(offset.X, offset.Y)

	img.DrawImage(bodyShape, opts)

	// Add details based on complexity
	if config.Complexity > 0.3 {
		g.addAnimationDetails(img, config, rng, frameIndex, frameCount)
	}

	return img, nil
}

// calculateAnimationOffset computes position offset for animation frame.
func calculateAnimationOffset(state string, frameIndex, frameCount int) struct{ X, Y float64 } {
	t := float64(frameIndex) / float64(frameCount)
	offset := struct{ X, Y float64 }{X: 0, Y: 0}

	switch state {
	case "walk", "run":
		// Bobbing motion
		cycle := math.Sin(t * 2 * math.Pi)
		offset.Y = cycle * 2.0 // 2 pixel vertical bob

	case "jump":
		// Parabolic arc
		offset.Y = -4.0 * (t - t*t) * 10.0 // Jump up and down

	case "attack":
		// Forward lunge
		if t < 0.5 {
			offset.X = t * 4.0 // Move forward
		} else {
			offset.X = (1.0 - t) * 4.0 // Return
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
	case "attack":
		// Swing arc
		if t < 0.3 {
			return -t * 0.5 // Wind up
		} else if t < 0.6 {
			return (t - 0.3) * 1.5 // Swing through
		} else {
			return (1.0 - t) * 0.3 // Follow through
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
		// Slight scale up during strike
		if t > 0.3 && t < 0.6 {
			return 1.0 + (t-0.3)*0.3
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
