package engine

import "image/color"

// TrailComponent enables entities (especially projectiles) to leave particle trails.
// Trails enhance visual feedback and make projectile movement more visible and appealing.
type TrailComponent struct {
	// Enabled controls whether trail generation is active
	Enabled bool

	// SpawnRate is the number of particles spawned per second
	SpawnRate float64

	// TimeSinceLastSpawn tracks time for particle spawning
	TimeSinceLastSpawn float64

	// ParticleLifetime is how long each trail particle lives (seconds)
	ParticleLifetime float64

	// ParticleSize is the size of trail particles in pixels
	ParticleSize float64

	// Color is the base color of trail particles (nil = use entity sprite color)
	Color *color.RGBA

	// FadeRate controls how quickly particles fade (0.0-1.0, higher = faster fade)
	FadeRate float64

	// SpreadX and SpreadY control the random spread of particles
	SpreadX float64
	SpreadY float64
}

// Type returns the component type identifier.
func (t TrailComponent) Type() string {
	return "trail"
}

// NewTrailComponent creates a trail component with sensible defaults for projectiles.
// ParticleLifetime: 0.5 seconds (particles fade quickly)
// SpawnRate: 30 particles per second (smooth trail)
// ParticleSize: 2.0 pixels (small, subtle particles)
// FadeRate: 0.8 (fade to 80% opacity per second)
func NewTrailComponent() *TrailComponent {
	return &TrailComponent{
		Enabled:            true,
		SpawnRate:          30.0, // 30 particles per second
		TimeSinceLastSpawn: 0.0,
		ParticleLifetime:   0.5, // Half-second lifetime
		ParticleSize:       2.0, // 2 pixel particles
		Color:              nil, // Use entity color by default
		FadeRate:           0.8, // Fast fade
		SpreadX:            2.0, // Small horizontal spread
		SpreadY:            2.0, // Small vertical spread
	}
}

// NewMagicTrailComponent creates a trail component tuned for magical projectiles.
// Brighter, slower-fading particles with more spread for a magical effect.
func NewMagicTrailComponent(color *color.RGBA) *TrailComponent {
	return &TrailComponent{
		Enabled:            true,
		SpawnRate:          40.0, // More particles for magical glow
		TimeSinceLastSpawn: 0.0,
		ParticleLifetime:   0.8,   // Longer lifetime for magical sparkle
		ParticleSize:       3.0,   // Larger particles
		Color:              color, // Custom magical color
		FadeRate:           0.6,   // Slower fade for magical glow
		SpreadX:            4.0,   // More spread for magical effect
		SpreadY:            4.0,
	}
}

// NewPhysicalTrailComponent creates a trail component tuned for physical projectiles.
// Subtle, fast-fading particles for arrows, bullets, etc.
func NewPhysicalTrailComponent() *TrailComponent {
	return &TrailComponent{
		Enabled:            true,
		SpawnRate:          20.0, // Fewer particles for subtle effect
		TimeSinceLastSpawn: 0.0,
		ParticleLifetime:   0.3, // Short lifetime
		ParticleSize:       1.5, // Very small particles
		Color:              nil, // Use projectile color
		FadeRate:           0.9, // Fast fade
		SpreadX:            1.0, // Minimal spread
		SpreadY:            1.0,
	}
}
