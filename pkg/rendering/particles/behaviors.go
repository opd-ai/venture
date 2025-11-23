// Package particles provides particle behavior definitions.
// This file implements advanced particle behaviors like physics,
// trails, and attraction for enhanced particle effects.
package particles

import (
	"math"
)

// ParticleBehavior defines behavior flags for particles.
type ParticleBehavior int

const (
	// BehaviorNone represents no special behavior
	BehaviorNone ParticleBehavior = 0
	// BehaviorGravity applies gravity force
	BehaviorGravity ParticleBehavior = 1 << 0
	// BehaviorAirResistance applies velocity damping
	BehaviorAirResistance ParticleBehavior = 1 << 1
	// BehaviorBounce makes particles bounce off surfaces
	BehaviorBounce ParticleBehavior = 1 << 2
	// BehaviorTrail leaves a fading trail behind particle
	BehaviorTrail ParticleBehavior = 1 << 3
	// BehaviorAttract moves particles toward an attractor point
	BehaviorAttract ParticleBehavior = 1 << 4
	// BehaviorOrbit makes particles orbit around a point
	BehaviorOrbit ParticleBehavior = 1 << 5
	// BehaviorRising makes particles float upward
	BehaviorRising ParticleBehavior = 1 << 6
	// BehaviorSink makes particles sink downward
	BehaviorSink ParticleBehavior = 1 << 7
)

// Has checks if a behavior flag is set.
func (b ParticleBehavior) Has(flag ParticleBehavior) bool {
	return b&flag != 0
}

// PhysicsConfig contains parameters for particle physics.
type PhysicsConfig struct {
	// Gravity acceleration (pixels/second^2)
	Gravity float64

	// AirResistance damping coefficient (0.0-1.0)
	// Higher values slow particles faster
	AirResistance float64

	// BounceDamping energy loss on bounce (0.0-1.0)
	// 0.0 = no bounce, 1.0 = perfect elastic bounce
	BounceDamping float64

	// GroundY position where particles bounce
	GroundY float64

	// AttractorX, AttractorY position of attraction point
	AttractorX float64
	AttractorY float64

	// AttractorStrength force of attraction
	AttractorStrength float64

	// OrbitRadius for orbital behavior
	OrbitRadius float64

	// OrbitSpeed angular velocity for orbits (radians/second)
	OrbitSpeed float64
}

// DefaultPhysicsConfig returns sensible default physics parameters.
func DefaultPhysicsConfig() PhysicsConfig {
	return PhysicsConfig{
		Gravity:           200.0, // Moderate gravity
		AirResistance:     0.05,  // Light air resistance
		BounceDamping:     0.5,   // Moderate energy loss
		GroundY:           0,     // Ground at Y=0
		AttractorX:        0,     // Center attraction
		AttractorY:        0,     // Center attraction
		AttractorStrength: 100.0, // Moderate attraction
		OrbitRadius:       50.0,  // Small orbit
		OrbitSpeed:        2.0,   // Slow rotation
	}
}

// ApplyPhysics updates particle velocity and position based on physics behaviors.
// This is called during particle system updates to simulate physical effects.
func ApplyPhysics(p *Particle, behavior ParticleBehavior, config PhysicsConfig, deltaTime float64) {
	applyGravityForces(p, behavior, config, deltaTime)
	applyAttractorPhysics(p, behavior, config, deltaTime)
	applyOrbitalMotion(p, behavior, config, deltaTime)
	applyBouncePhysics(p, behavior, config)
}

// applyGravityForces applies gravity, air resistance, rising, and sinking forces.
func applyGravityForces(p *Particle, behavior ParticleBehavior, config PhysicsConfig, deltaTime float64) {
	if behavior.Has(BehaviorGravity) {
		p.VY += config.Gravity * deltaTime
	}

	if behavior.Has(BehaviorAirResistance) {
		damping := 1.0 - (config.AirResistance * deltaTime)
		if damping < 0 {
			damping = 0
		}
		p.VX *= damping
		p.VY *= damping
	}

	if behavior.Has(BehaviorRising) {
		p.VY -= config.Gravity * 0.5 * deltaTime
	}

	if behavior.Has(BehaviorSink) {
		p.VY += config.Gravity * 0.3 * deltaTime
	}
}

// applyAttractorPhysics applies attraction force toward a point.
func applyAttractorPhysics(p *Particle, behavior ParticleBehavior, config PhysicsConfig, deltaTime float64) {
	if !behavior.Has(BehaviorAttract) {
		return
	}

	dx := config.AttractorX - p.X
	dy := config.AttractorY - p.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 1.0 {
		force := config.AttractorStrength / (dist * dist) * deltaTime
		p.VX += (dx / dist) * force
		p.VY += (dy / dist) * force
	}
}

// applyOrbitalMotion applies centripetal and tangential forces for circular motion.
func applyOrbitalMotion(p *Particle, behavior ParticleBehavior, config PhysicsConfig, deltaTime float64) {
	if !behavior.Has(BehaviorOrbit) {
		return
	}

	dx := p.X - config.AttractorX
	dy := p.Y - config.AttractorY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 0.1 {
		return
	}

	angle := math.Atan2(dy, dx)
	tangentAngle := angle + math.Pi/2

	centripetalForce := config.OrbitSpeed * config.OrbitSpeed * dist
	p.VX += -math.Cos(angle) * centripetalForce * deltaTime
	p.VY += -math.Sin(angle) * centripetalForce * deltaTime

	p.VX += math.Cos(tangentAngle) * config.OrbitSpeed * config.OrbitRadius * deltaTime
	p.VY += math.Sin(tangentAngle) * config.OrbitSpeed * config.OrbitRadius * deltaTime
}

// applyBouncePhysics handles bouncing off ground with damping.
func applyBouncePhysics(p *Particle, behavior ParticleBehavior, config PhysicsConfig) {
	if !behavior.Has(BehaviorBounce) {
		return
	}

	if p.Y > config.GroundY && p.VY > 0 {
		p.Y = config.GroundY
		p.VY = -p.VY * config.BounceDamping
		p.VX *= config.BounceDamping
	}
}

// TrailConfig contains parameters for particle trails.
type TrailConfig struct {
	// MaxTrailLength number of trail segments
	MaxTrailLength int

	// TrailFadeRate how fast trail fades (0.0-1.0 per second)
	TrailFadeRate float64

	// TrailSpacing minimum distance between trail points
	TrailSpacing float64
}

// DefaultTrailConfig returns sensible default trail parameters.
func DefaultTrailConfig() TrailConfig {
	return TrailConfig{
		MaxTrailLength: 10,
		TrailFadeRate:  0.5,
		TrailSpacing:   5.0,
	}
}

// TrailPoint represents a point in a particle's trail.
type TrailPoint struct {
	X, Y  float64
	Alpha float64 // 0.0-1.0, 1.0 is fully visible
}

// ParticleWithTrail wraps a Particle with trail data.
type ParticleWithTrail struct {
	Particle
	Trail         []TrailPoint
	LastTrailX    float64
	LastTrailY    float64
	TrailDistance float64 // Accumulated distance for spacing
}

// UpdateTrail adds new trail points and fades existing ones.
func UpdateTrail(pwt *ParticleWithTrail, config TrailConfig, deltaTime float64) {
	// Calculate distance moved
	dx := pwt.X - pwt.LastTrailX
	dy := pwt.Y - pwt.LastTrailY
	dist := math.Sqrt(dx*dx + dy*dy)
	pwt.TrailDistance += dist

	// Add new trail point if moved enough
	if pwt.TrailDistance >= config.TrailSpacing {
		newPoint := TrailPoint{
			X:     pwt.LastTrailX,
			Y:     pwt.LastTrailY,
			Alpha: 1.0,
		}

		// Add to front of trail
		pwt.Trail = append([]TrailPoint{newPoint}, pwt.Trail...)

		// Limit trail length
		if len(pwt.Trail) > config.MaxTrailLength {
			pwt.Trail = pwt.Trail[:config.MaxTrailLength]
		}

		pwt.LastTrailX = pwt.X
		pwt.LastTrailY = pwt.Y
		pwt.TrailDistance = 0
	}

	// Fade all trail points
	fadeAmount := config.TrailFadeRate * deltaTime
	aliveTrail := make([]TrailPoint, 0, len(pwt.Trail))
	for i := range pwt.Trail {
		pwt.Trail[i].Alpha -= fadeAmount
		if pwt.Trail[i].Alpha > 0 {
			aliveTrail = append(aliveTrail, pwt.Trail[i])
		}
	}
	pwt.Trail = aliveTrail
}
