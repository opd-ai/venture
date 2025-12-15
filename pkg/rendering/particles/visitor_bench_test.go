package particles

import (
	"image/color"
	"testing"
)

// BenchmarkGetAliveParticles benchmarks the slice-returning method (allocates).
func BenchmarkGetAliveParticles(b *testing.B) {
	ps := createBenchmarkParticleSystem(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alive := ps.GetAliveParticles()
		_ = len(alive) // Prevent dead code elimination
	}
}

// BenchmarkVisitAliveParticles benchmarks the zero-allocation visitor pattern.
func BenchmarkVisitAliveParticles(b *testing.B) {
	ps := createBenchmarkParticleSystem(100)
	count := 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.VisitAliveParticles(func(p *Particle) {
			count++
		})
	}
	_ = count // Prevent dead code elimination
}

// createBenchmarkParticleSystem creates a particle system for benchmarking.
func createBenchmarkParticleSystem(count int) *ParticleSystem {
	ps := &ParticleSystem{
		Particles: make([]Particle, count),
		Type:      ParticleSpark,
		Config:    DefaultConfig(),
	}

	// Half alive, half dead
	for i := 0; i < count; i++ {
		ps.Particles[i] = Particle{
			X:           float64(i * 10),
			Y:           float64(i * 5),
			VX:          100.0,
			VY:          -50.0,
			Color:       color.RGBA{255, 100, 50, 255},
			Size:        3.0,
			Life:        0.0, // Will be set below
			InitialLife: 1.0,
		}
		// Make half alive, half dead
		if i%2 == 0 {
			ps.Particles[i].Life = 0.5
		}
	}

	return ps
}
