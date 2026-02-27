// Package main demonstrates animation timing and frame rate control in the Venture engine.
//
// This example shows:
// - Default animation frame rates (12 FPS for close range)
// - Distance-based LOD (Level of Detail) animation rates
// - Custom frame timing for different animation states
// - How deltaTime affects animation progression
//
// Run with: go run examples/animation_timing_demo.go
//
// Note: This is a pure calculation demo that doesn't require a display or Ebiten runtime.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Venture Animation Timing Demo")
	fmt.Println("=============================")
	fmt.Println()

	// Demonstrate default animation timing
	demonstrateDefaultTiming()

	// Demonstrate distance-based LOD timing
	demonstrateLODTiming()

	// Demonstrate custom timing for different states
	demonstrateCustomTiming()

	// Demonstrate frame progression
	demonstrateFrameProgression()
}

// demonstrateDefaultTiming shows the default 12 FPS animation timing
func demonstrateDefaultTiming() {
	fmt.Println("1. Default Animation Timing")
	fmt.Println("---------------------------")

	// Default values from NewAnimationComponent
	frameTime := 1.0 / 12.0 // 12 FPS for close range (from animation_component.go:219)
	frameCount := 8         // 8 frames per animation state (from animation_component.go:224)

	fmt.Printf("Default FrameTime: %.6f seconds (%.1f FPS)\n", frameTime, 1.0/frameTime)
	fmt.Printf("Frame Count: %d frames per animation state\n", frameCount)
	fmt.Printf("Total Animation Duration: %.2f seconds per cycle\n", frameTime*float64(frameCount))
	fmt.Println()
}

// demonstrateLODTiming shows distance-based frame rate adjustments
func demonstrateLODTiming() {
	fmt.Println("2. Distance-Based LOD Timing")
	fmt.Println("----------------------------")

	distances := []struct {
		name     string
		distance float64
		fps      float64
		usage    string
	}{
		{"Close range", 150.0, 12.0, "Full animation quality for nearby entities"},
		{"Medium range", 300.0, 6.0, "Half rate for medium distance (performance)"},
		{"Far range", 500.0, 3.0, "Minimal animation for distant entities"},
	}

	for _, d := range distances {
		frameTime := 1.0 / d.fps
		duration := frameTime * 8.0 // 8 frames per animation
		fmt.Printf("  %s (≈%.0fpx):\n", d.name, d.distance)
		fmt.Printf("    - Frame Rate: %.1f FPS\n", d.fps)
		fmt.Printf("    - Frame Time: %.4f seconds\n", frameTime)
		fmt.Printf("    - Cycle Duration: %.2f seconds\n", duration)
		fmt.Printf("    - Usage: %s\n", d.usage)
		fmt.Println()
	}
}

// demonstrateCustomTiming shows how to customize animation speed
func demonstrateCustomTiming() {
	fmt.Println("3. Custom Animation Timing Examples")
	fmt.Println("------------------------------------")

	examples := []struct {
		state       string
		frameTime   float64
		description string
	}{
		{"idle", 1.0 / 8.0, "Slow idle breathing (8 FPS)"},
		{"walk", 1.0 / 12.0, "Standard walk animation (12 FPS)"},
		{"run", 1.0 / 16.0, "Fast run animation (16 FPS)"},
		{"attack", 1.0 / 20.0, "Quick attack animation (20 FPS)"},
		{"cast", 1.0 / 10.0, "Deliberate spell cast (10 FPS)"},
	}

	frameCount := 8 // Standard frame count

	for _, ex := range examples {
		fps := 1.0 / ex.frameTime
		duration := ex.frameTime * float64(frameCount)

		fmt.Printf("  %s:\n", ex.state)
		fmt.Printf("    - Frame Time: %.4f seconds (%.1f FPS)\n", ex.frameTime, fps)
		fmt.Printf("    - Total Duration: %.2f seconds\n", duration)
		fmt.Printf("    - Description: %s\n", ex.description)
		fmt.Println()
	}
}

// demonstrateFrameProgression shows how frames advance over time
func demonstrateFrameProgression() {
	fmt.Println("4. Frame Progression Simulation")
	fmt.Println("--------------------------------")

	// Animation state
	frameTime := 1.0 / 12.0 // 12 FPS
	frameCount := 8         // 8 frames per animation
	timeAccumulator := 0.0
	frameIndex := 0
	currentState := "walk"

	// Simulate 60 FPS game loop
	deltaTime := 1.0 / 60.0 // 16.67ms per frame
	fmt.Printf("Simulating 60 FPS game loop (%.4f seconds per frame)\n", deltaTime)
	fmt.Printf("Animation: %s at %.1f FPS\n\n", currentState, 1.0/frameTime)

	fmt.Println("Game Frame | Time (s) | Anim Frame | Accumulator")
	fmt.Println("-----------|----------|------------|------------")

	// Simulate one full animation cycle
	totalTime := 0.0
	gameFrame := 0
	lastAnimFrame := 0

	for totalTime < frameTime*float64(frameCount)*1.5 {
		// Simulate frame update (from animation_system.go:658-694)
		timeAccumulator += deltaTime
		if timeAccumulator >= frameTime {
			timeAccumulator -= frameTime
			frameIndex++
			if frameIndex >= frameCount {
				frameIndex = 0 // Loop back
			}
		}

		// Print frame progression (only when frame changes or every 10 game frames)
		if frameIndex != lastAnimFrame || gameFrame%10 == 0 {
			fmt.Printf("%10d | %8.4f | %10d | %11.4f\n",
				gameFrame, totalTime, frameIndex, timeAccumulator)
			lastAnimFrame = frameIndex
		}

		gameFrame++
		totalTime += deltaTime

		// Safety limit
		if gameFrame > 100 {
			break
		}
	}

	fmt.Println()
	fmt.Printf("Note: At 60 FPS, animation frame changes approximately every %.0f game frames\n",
		frameTime/(1.0/60.0))
	fmt.Println()
}
