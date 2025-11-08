// Package main provides a CLI tool for testing advanced particle physics.
// This tool visualizes fluid (SPH), fire, smoke, and debris simulations
// to verify correctness and performance of Phase 18.2 implementations.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

var (
	physicsType = flag.String("type", "fluid", "Physics type: fluid, fire, smoke, debris")
	count       = flag.Int("count", 100, "Number of particles")
	seed        = flag.Int64("seed", 12345, "Random seed for deterministic generation")
	frames      = flag.Int("frames", 60, "Number of simulation frames")
	output      = flag.String("output", "physics_test.png", "Output PNG file")
	width       = flag.Int("width", 800, "Output image width")
	height      = flag.Int("height", 600, "Output image height")
	verbose     = flag.Bool("v", false, "Verbose logging")
)

func main() {
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.WithFields(logrus.Fields{
		"type":   *physicsType,
		"count":  *count,
		"seed":   *seed,
		"frames": *frames,
	}).Info("Starting particle physics test")

	// Create particles based on type
	physParticles := make([]particles.PhysicsParticle, *count)
	rng := rand.New(rand.NewSource(*seed))

	switch *physicsType {
	case "fluid":
		initFluidParticles(physParticles, rng)
	case "fire":
		initFireParticles(physParticles, rng)
	case "smoke":
		initSmokeParticles(physParticles, rng)
	case "debris":
		initDebrisParticles(physParticles, rng)
	default:
		logrus.Fatalf("Unknown physics type: %s", *physicsType)
	}

	// Simulate
	deltaTime := 1.0 / 60.0 // 60 FPS
	startTime := time.Now()

	for i := 0; i < *frames; i++ {
		switch *physicsType {
		case "fluid":
			config := particles.DefaultSPHConfig()
			particles.UpdateSPH(physParticles, config, deltaTime)
		case "fire":
			config := particles.DefaultFireConfig()
			particles.UpdateFire(physParticles, config, deltaTime, rng)
		case "smoke":
			config := particles.DefaultSmokeConfig()
			particles.UpdateSmoke(physParticles, config, deltaTime, float64(i)*deltaTime)
		case "debris":
			config := particles.DefaultDebrisConfig()
			particles.UpdateDebris(physParticles, config, deltaTime, float64(*height-50))
		}

		// Update positions for all types
		for j := range physParticles {
			p := &physParticles[j]
			p.X += p.VX * deltaTime
			p.Y += p.VY * deltaTime
			p.Life -= deltaTime / p.InitialLife
		}
	}

	elapsed := time.Since(startTime)
	logrus.WithFields(logrus.Fields{
		"duration":  elapsed,
		"fps":       float64(*frames) / elapsed.Seconds(),
		"per_frame": elapsed / time.Duration(*frames),
	}).Info("Simulation complete")

	// Render to image
	img := renderParticles(physParticles, *width, *height)

	// Save PNG
	f, err := os.Create(*output)
	if err != nil {
		logrus.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		logrus.Fatalf("Failed to encode PNG: %v", err)
	}

	logrus.WithField("file", *output).Info("Output saved")

	// Print statistics
	printStatistics(physParticles)
}

func initFluidParticles(ps []particles.PhysicsParticle, rng *rand.Rand) {
	// Create fluid blob in center
	centerX := float64(*width) / 2
	centerY := float64(*height) / 4

	for i := range ps {
		radius := rng.Float64() * 50

		ps[i] = particles.PhysicsParticle{
			Particle: particles.Particle{
				X:           centerX + radius*rng.Float64()*10,
				Y:           centerY + radius*rng.Float64()*5,
				VX:          (rng.Float64() - 0.5) * 20,
				VY:          rng.Float64() * 50,
				Size:        2.0 + rng.Float64()*2,
				Life:        1.0,
				InitialLife: 10.0,
				Color:       color.RGBA{R: 100, G: 150, B: 255, A: 200},
			},
			PhysicsType: particles.PhysicsFluid,
		}
	}
}

func initFireParticles(ps []particles.PhysicsParticle, rng *rand.Rand) {
	// Create fire at bottom
	baseY := float64(*height) - 50

	for i := range ps {
		ps[i] = particles.PhysicsParticle{
			Particle: particles.Particle{
				X:           float64(*width)/2 + (rng.Float64()-0.5)*100,
				Y:           baseY + rng.Float64()*50,
				VX:          (rng.Float64() - 0.5) * 10,
				VY:          -rng.Float64() * 20,
				Size:        3.0 + rng.Float64()*3,
				Life:        1.0,
				InitialLife: 2.0,
				Color:       color.RGBA{R: 255, G: uint8(100 + rng.Intn(100)), B: 0, A: 255},
			},
			PhysicsType: particles.PhysicsFire,
			Heat:        0.3 + rng.Float64()*0.7,
			Ignited:     rng.Float64() < 0.3,
			FuelRemain:  1.0,
		}
	}
}

func initSmokeParticles(ps []particles.PhysicsParticle, rng *rand.Rand) {
	// Create smoke rising from bottom
	baseY := float64(*height) - 50

	for i := range ps {
		ps[i] = particles.PhysicsParticle{
			Particle: particles.Particle{
				X:           float64(*width)/2 + (rng.Float64()-0.5)*80,
				Y:           baseY + rng.Float64()*100,
				VX:          (rng.Float64() - 0.5) * 5,
				VY:          -10 - rng.Float64()*10,
				Size:        2.0 + rng.Float64()*2,
				Life:        1.0,
				InitialLife: 3.0,
				Color:       color.RGBA{R: 100, G: 100, B: 100, A: 150},
			},
			PhysicsType:     particles.PhysicsSmoke,
			TurbulencePhase: rng.Float64() * 6.28,
		}
	}
}

func initDebrisParticles(ps []particles.PhysicsParticle, rng *rand.Rand) {
	// Create debris explosion from center
	centerX := float64(*width) / 2
	centerY := float64(*height) / 2

	for i := range ps {
		speed := 50 + rng.Float64()*100

		ps[i] = particles.PhysicsParticle{
			Particle: particles.Particle{
				X:           centerX,
				Y:           centerY,
				VX:          speed * rng.Float64() * 2,
				VY:          speed * rng.Float64() * 2,
				Size:        2.0 + rng.Float64()*4,
				Life:        1.0,
				InitialLife: 5.0,
				Color:       color.RGBA{R: uint8(100 + rng.Intn(155)), G: uint8(50 + rng.Intn(100)), B: uint8(50 + rng.Intn(100)), A: 255},
			},
			PhysicsType:     particles.PhysicsDebris,
			AngularVelocity: (rng.Float64() - 0.5) * 10,
		}
	}
}

func renderParticles(ps []particles.PhysicsParticle, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Clear to black
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	// Draw particles
	for i := range ps {
		p := &ps[i]
		if p.Life <= 0 {
			continue
		}

		// Convert particle position to image coordinates
		px := int(p.X)
		py := int(p.Y)

		if px < 0 || px >= width || py < 0 || py >= height {
			continue
		}

		// Draw particle as a circle
		size := int(p.Size)
		for dy := -size; dy <= size; dy++ {
			for dx := -size; dx <= size; dx++ {
				if dx*dx+dy*dy <= size*size {
					x := px + dx
					y := py + dy
					if x >= 0 && x < width && y >= 0 && y < height {
						// Blend with existing color
						existing := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
						newCol := color.RGBAModel.Convert(p.Color).(color.RGBA)

						// Simple alpha blending
						alpha := float64(newCol.A) / 255.0 * p.Life
						blended := color.RGBA{
							R: uint8(float64(existing.R)*(1-alpha) + float64(newCol.R)*alpha),
							G: uint8(float64(existing.G)*(1-alpha) + float64(newCol.G)*alpha),
							B: uint8(float64(existing.B)*(1-alpha) + float64(newCol.B)*alpha),
							A: 255,
						}
						img.Set(x, y, blended)
					}
				}
			}
		}
	}

	return img
}

func printStatistics(ps []particles.PhysicsParticle) {
	if len(ps) == 0 {
		return
	}

	alive := 0
	var avgVel, maxVel float64
	var minX, maxX, minY, maxY float64 = 1e9, -1e9, 1e9, -1e9

	for i := range ps {
		p := &ps[i]
		if p.Life > 0 {
			alive++
			vel := p.VX*p.VX + p.VY*p.VY
			avgVel += vel
			if vel > maxVel {
				maxVel = vel
			}
		}

		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	if alive > 0 {
		avgVel /= float64(alive)
	}

	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Alive particles: %d / %d (%.1f%%)\n", alive, len(ps), 100.0*float64(alive)/float64(len(ps)))
	fmt.Printf("  Average velocity²: %.2f\n", avgVel)
	fmt.Printf("  Max velocity²: %.2f\n", maxVel)
	fmt.Printf("  Bounds: X[%.1f, %.1f], Y[%.1f, %.1f]\n", minX, maxX, minY, maxY)

	// Type-specific stats
	switch *physicsType {
	case "fluid":
		var avgDensity, avgPressure float64
		for i := range ps {
			avgDensity += ps[i].Density
			avgPressure += ps[i].Pressure
		}
		fmt.Printf("  Avg density: %.2f\n", avgDensity/float64(len(ps)))
		fmt.Printf("  Avg pressure: %.2f\n", avgPressure/float64(len(ps)))

	case "fire":
		var avgHeat float64
		ignited := 0
		for i := range ps {
			avgHeat += ps[i].Heat
			if ps[i].Ignited {
				ignited++
			}
		}
		fmt.Printf("  Avg heat: %.2f\n", avgHeat/float64(len(ps)))
		fmt.Printf("  Ignited: %d (%.1f%%)\n", ignited, 100.0*float64(ignited)/float64(len(ps)))

	case "debris":
		var avgAngVel float64
		for i := range ps {
			avgAngVel += ps[i].AngularVelocity
		}
		fmt.Printf("  Avg angular velocity: %.2f rad/s\n", avgAngVel/float64(len(ps)))
	}
}
