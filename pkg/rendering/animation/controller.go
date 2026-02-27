package animation

import (
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// Controller manages 8-frame animations with articulation and caching.
// This is the main entry point for the Phase 46 animation system.
type Controller struct {
	generator *sprites.Generator
	cache     *AnimationCache
	config    ArticulationConfig

	// Performance tracking
	frameGenerationTime time.Duration
	cacheHitRate        float64
}

// NewController creates a new animation controller.
func NewController(generator *sprites.Generator) *Controller {
	return &Controller{
		generator: generator,
		cache:     NewAnimationCache(50*1024*1024, 1000), // 50MB, 1000 entries
		config:    DefaultArticulationConfig(),
	}
}

// NewControllerWithCache creates a controller with a custom cache.
func NewControllerWithCache(generator *sprites.Generator, cache *AnimationCache) *Controller {
	return &Controller{
		generator: generator,
		cache:     cache,
		config:    DefaultArticulationConfig(),
	}
}

// SetArticulationConfig updates the articulation configuration.
func (c *Controller) SetArticulationConfig(config ArticulationConfig) {
	c.config = config
}

// GenerateFrame generates a single animation frame with articulation.
// This is the core frame generation function that integrates all Phase 46 features.
func (c *Controller) GenerateFrame(seed int64, state string, frameIndex, frameCount int, direction Direction8, spriteConfig sprites.Config) (*ebiten.Image, error) {
	// Check cache first
	cacheKey := CacheKey{
		Seed:       seed,
		State:      state,
		Direction:  direction,
		FrameIndex: frameIndex,
	}

	if cached, found := c.cache.Get(cacheKey); found {
		return cached, nil
	}

	// Generate frame
	// NOTE: time.Now() used here for non-critical performance metrics only.
	// This does not affect animation determinism (all generation is seed-based).
	// Metrics are for monitoring/observability, not gameplay logic.
	startTime := time.Now()
	frame, err := c.generateFrameInternal(seed, state, frameIndex, frameCount, direction, spriteConfig)
	c.frameGenerationTime = time.Since(startTime)

	if err != nil {
		return nil, err
	}

	// Cache for future use
	c.cache.Put(cacheKey, frame)

	return frame, nil
}

// generateFrameInternal performs the actual frame generation with articulation.
func (c *Controller) generateFrameInternal(seed int64, state string, frameIndex, frameCount int, direction Direction8, spriteConfig sprites.Config) (*ebiten.Image, error) {
	// Calculate articulation for this frame
	articulation := CalculateArticulation(state, frameIndex, frameCount, direction, c.config)

	// Update sprite config for direction
	spriteConfig.Custom["facing"] = direction.To4Direction()
	spriteConfig.Custom["direction8"] = direction.String()

	// Generate base sprite
	baseSprite, err := c.generator.Generate(spriteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base sprite: %w", err)
	}

	// Apply articulation transformations
	frame := c.applyArticulation(baseSprite, articulation, spriteConfig)

	return frame, nil
}

// bodyPartRegion defines the proportional bounds of a body part within a sprite.
// Regions are defined as fractions of sprite dimensions (0.0-1.0).
type bodyPartRegion struct {
	yStart float64 // Start Y as fraction of height
	yEnd   float64 // End Y as fraction of height
	xStart float64 // Start X as fraction of width
	xEnd   float64 // End X as fraction of width
}

// humanoidBodyParts returns approximate body part regions for humanoid sprites.
// Proportions based on standard humanoid template: Head 12%, Torso 40%, Legs 48%.
func humanoidBodyParts() map[string]bodyPartRegion {
	return map[string]bodyPartRegion{
		"head": {
			yStart: 0.0,
			yEnd:   0.20, // Top 20% of sprite
			xStart: 0.25,
			xEnd:   0.75,
		},
		"torso": {
			yStart: 0.20,
			yEnd:   0.55, // Middle 35% of sprite
			xStart: 0.20,
			xEnd:   0.80,
		},
		"leftArm": {
			yStart: 0.22,
			yEnd:   0.55,
			xStart: 0.0,
			xEnd:   0.30, // Left side
		},
		"rightArm": {
			yStart: 0.22,
			yEnd:   0.55,
			xStart: 0.70,
			xEnd:   1.0, // Right side
		},
		"leftLeg": {
			yStart: 0.55,
			yEnd:   1.0, // Bottom 45%
			xStart: 0.25,
			xEnd:   0.50, // Left half of legs area
		},
		"rightLeg": {
			yStart: 0.55,
			yEnd:   1.0,
			xStart: 0.50,
			xEnd:   0.75, // Right half of legs area
		},
		"tail": {
			yStart: 0.45,
			yEnd:   0.70,
			xStart: 0.35,
			xEnd:   0.65,
		},
	}
}

// applyArticulation applies body part articulation to a sprite.
// This creates the final animated frame by transforming each body part region
// according to its calculated articulation offsets and rotations.
func (c *Controller) applyArticulation(baseSprite *ebiten.Image, articulation Articulation, config sprites.Config) *ebiten.Image {
	bounds := baseSprite.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create output image with padding for articulation offsets
	// Scaled for 64×64 sprites (Phase 45 migration): 10 → 20 pixels
	padding := 20 // Extra pixels for articulation range (scaled 2× for 64×64)
	outputWidth := width + padding*2
	outputHeight := height + padding*2
	output := ebiten.NewImage(outputWidth, outputHeight)

	// Get body part regions
	regions := humanoidBodyParts()

	// Draw body parts from back to front (Z-order)
	// Order: tail (if present), legs, torso, arms, head
	drawOrder := []struct {
		name string
		art  ArticulationOffset
	}{
		{"tail", articulation.Tail},
		{"leftLeg", articulation.LeftLeg},
		{"rightLeg", articulation.RightLeg},
		{"torso", articulation.Torso},
		{"leftArm", articulation.LeftArm},
		{"rightArm", articulation.RightArm},
		{"head", articulation.Head},
	}

	for _, part := range drawOrder {
		region, exists := regions[part.name]
		if !exists {
			continue
		}

		// Calculate pixel bounds for this region
		x0 := int(region.xStart * float64(width))
		y0 := int(region.yStart * float64(height))
		x1 := int(region.xEnd * float64(width))
		y1 := int(region.yEnd * float64(height))

		// Ensure valid bounds
		if x1 <= x0 || y1 <= y0 {
			continue
		}

		// For sub-image extraction, create a new image and draw the region
		partWidth := x1 - x0
		partHeight := y1 - y0
		partImg := ebiten.NewImage(partWidth, partHeight)

		// Draw the region from baseSprite to partImg
		srcOpts := &ebiten.DrawImageOptions{}
		srcOpts.GeoM.Translate(-float64(x0), -float64(y0))
		partImg.DrawImage(baseSprite, srcOpts)

		// Calculate center point for rotation
		centerX := float64(partWidth) / 2
		centerY := float64(partHeight) / 2

		// Create draw options for this body part
		opts := &ebiten.DrawImageOptions{}

		// Apply rotation around part center
		opts.GeoM.Translate(-centerX, -centerY)
		opts.GeoM.Rotate(part.art.Rotation)
		opts.GeoM.Translate(centerX, centerY)

		// Apply position offset and move to final position
		destX := float64(x0) + float64(padding) + part.art.X
		destY := float64(y0) + float64(padding) + part.art.Y
		opts.GeoM.Translate(destX, destY)

		output.DrawImage(partImg, opts)
	}

	return output
}

// GenerateSequence generates a complete animation sequence (all frames).
// Returns a slice of frames that can be played in order.
func (c *Controller) GenerateSequence(seed int64, state string, direction Direction8, spriteConfig sprites.Config) ([]*ebiten.Image, error) {
	frameCount := GetFrameCount(state)
	frames := make([]*ebiten.Image, frameCount)

	for i := 0; i < frameCount; i++ {
		frame, err := c.GenerateFrame(seed, state, i, frameCount, direction, spriteConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate frame %d: %w", i, err)
		}
		frames[i] = frame
	}

	return frames, nil
}

// PrecomputeCommon pre-generates common animation sequences.
// This improves runtime performance by front-loading generation work.
func (c *Controller) PrecomputeCommon(seeds []int64, spriteConfig sprites.Config) error {
	states := CommonAnimationStates()
	directions := CommonDirections()

	sequences := make([]PrewarmSequence, 0, len(seeds)*len(states)*len(directions))
	for _, seed := range seeds {
		for _, state := range states {
			frameCount := GetFrameCount(state)
			for _, direction := range directions {
				sequences = append(sequences, PrewarmSequence{
					Seed:       seed,
					State:      state,
					Direction:  direction,
					FrameCount: frameCount,
				})
			}
		}
	}

	// Pre-warm cache
	return c.cache.Prewarm(sequences, func(key CacheKey) (*ebiten.Image, error) {
		return c.generateFrameInternal(
			key.Seed,
			key.State,
			key.FrameIndex,
			GetFrameCount(key.State),
			key.Direction,
			spriteConfig,
		)
	})
}

// GetFrameCount returns the number of frames for an animation state.
// Phase 46 uses 8 frames for most animations.
func GetFrameCount(state string) int {
	switch state {
	case "idle":
		return 8 // Smooth breathing cycle
	case "walk":
		return 8 // Full stride cycle (left-right-left-right)
	case "run":
		return 8 // Exaggerated running cycle
	case "attack":
		return 8 // Wind-up, strike, follow-through
	case "cast":
		return 8 // Spell casting gesture
	case "hit":
		return 4 // Quick hit reaction (shorter)
	case "death":
		return 8 // Falling/collapsing
	case "jump":
		return 6 // Crouch, jump, land
	case "crouch":
		return 4 // Crouch down/up
	case "use":
		return 6 // Item usage
	default:
		return 8 // Default to 8 frames
	}
}

// GetFrameTime returns the time per frame in seconds for an animation state.
// Target is 60 FPS with smooth 8-frame cycles.
func GetFrameTime(state string) float64 {
	switch state {
	case "idle":
		return 1.0 / 8.0 // 8 FPS for breathing (0.125s per frame)
	case "walk":
		return 1.0 / 12.0 // 12 FPS for natural walk (0.083s per frame)
	case "run":
		return 1.0 / 16.0 // 16 FPS for fast run (0.0625s per frame)
	case "attack":
		return 1.0 / 16.0 // 16 FPS for quick attack (0.0625s per frame)
	case "cast":
		return 1.0 / 10.0 // 10 FPS for deliberate casting (0.1s per frame)
	case "hit":
		return 1.0 / 20.0 // 20 FPS for quick reaction (0.05s per frame)
	case "death":
		return 1.0 / 8.0 // 8 FPS for dramatic death (0.125s per frame)
	case "jump":
		return 1.0 / 12.0 // 12 FPS for jump (0.083s per frame)
	default:
		return 1.0 / 12.0 // Default 12 FPS
	}
}

// InterpolateFrame creates an interpolated frame between two frames.
// This enables sub-frame smoothness for 60 FPS rendering.
// t is the interpolation factor (0.0 = frameA, 1.0 = frameB)
func (c *Controller) InterpolateFrame(frameA, frameB *ebiten.Image, t float64) *ebiten.Image {
	if t <= 0.0 {
		return frameA
	}
	if t >= 1.0 {
		return frameB
	}

	// Create blended output
	bounds := frameA.Bounds()
	output := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	// Draw frameA with alpha = 1-t
	optsA := &ebiten.DrawImageOptions{}
	optsA.ColorScale.ScaleAlpha(float32(1.0 - t))
	output.DrawImage(frameA, optsA)

	// Draw frameB with alpha = t
	optsB := &ebiten.DrawImageOptions{}
	optsB.ColorScale.ScaleAlpha(float32(t))
	output.DrawImage(frameB, optsB)

	return output
}

// GetCacheStats returns current cache statistics.
func (c *Controller) GetCacheStats() CacheStats {
	stats := c.cache.GetStats()
	c.cacheHitRate = stats.HitRate()
	return stats
}

// GetPerformanceMetrics returns performance metrics for monitoring.
type PerformanceMetrics struct {
	FrameGenerationTime time.Duration
	CacheHitRate        float64
	CacheSize           int64
	CacheCount          int
}

// GetPerformanceMetrics returns current performance metrics.
func (c *Controller) GetPerformanceMetrics() PerformanceMetrics {
	stats := c.cache.GetStats()
	return PerformanceMetrics{
		FrameGenerationTime: c.frameGenerationTime,
		CacheHitRate:        stats.HitRate(),
		CacheSize:           c.cache.Size(),
		CacheCount:          c.cache.Count(),
	}
}

// ClearCache clears the animation cache.
// Useful for memory management or when switching scenes.
func (c *Controller) ClearCache() {
	c.cache.Clear()
}

// CalculateDirection8 is a helper to convert velocity to 8-direction.
// Convenience wrapper for FromVelocity.
func CalculateDirection8(vx, vy float64) Direction8 {
	return FromVelocity(vx, vy)
}

// SmoothDirectionTransition smoothly transitions between directions.
// Returns a blend factor for smooth rotation between old and new directions.
func SmoothDirectionTransition(oldDir, newDir Direction8, deltaTime float64) float64 {
	// Calculate angular difference
	oldAngle := oldDir.Angle()
	newAngle := newDir.Angle()
	diff := newAngle - oldAngle

	// Normalize to [-π, π]
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}

	// Smooth transition over 0.1 seconds
	transitionSpeed := 10.0 // transitions per second
	blend := math.Min(deltaTime*transitionSpeed, 1.0)

	return blend
}
