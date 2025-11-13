// Package main provides a CLI tool for testing adaptive music composition.package musictest

// This tool demonstrates context-based music generation, layer management,
// and smooth transitions between different gameplay states.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
)

func main() {
	// Command-line flags
	mode := flag.String("mode", "contexts", "Test mode: contexts, layers, transitions, intensity, all")
	genre := flag.String("genre", "fantasy", "Music genre: fantasy, scifi, horror, cyberpunk, post-apocalyptic")
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	duration := flag.Float64("duration", 2.0, "Duration of generated tracks in seconds")
	sampleRate := flag.Int("sample-rate", 44100, "Audio sample rate")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Parse()

	fmt.Printf("Adaptive Music System Test Tool\n")
	fmt.Printf("================================\n\n")
	fmt.Printf("Mode:        %s\n", *mode)
	fmt.Printf("Genre:       %s\n", *genre)
	fmt.Printf("Seed:        %d\n", *seed)
	fmt.Printf("Duration:    %.1fs\n", *duration)
	fmt.Printf("Sample Rate: %d Hz\n", *sampleRate)
	fmt.Printf("\n")

	composer := music.NewAdaptiveComposer(*sampleRate, *seed)
	composer.Initialize(*genre, 60) // Middle C

	switch *mode {
	case "contexts":
		testContexts(composer, *duration, *verbose)
	case "layers":
		testLayers(composer, *duration, *verbose)
	case "transitions":
		testTransitions(composer, *duration, *verbose)
	case "intensity":
		testIntensity(composer, *duration, *verbose)
	case "all":
		testContexts(composer, *duration, *verbose)
		fmt.Println()
		testLayers(composer, *duration, *verbose)
		fmt.Println()
		testTransitions(composer, *duration, *verbose)
		fmt.Println()
		testIntensity(composer, *duration, *verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: contexts, layers, transitions, intensity, all")
		os.Exit(1)
	}
}

// testContexts tests different music contexts
func testContexts(composer *music.AdaptiveComposer, duration float64, verbose bool) {
	fmt.Println("Testing Music Contexts")
	fmt.Println("======================")

	contexts := []struct {
		name        string
		description string
	}{
		{"exploration", "Calm exploration music"},
		{"combat", "Intense combat music with percussion"},
		{"boss", "Epic boss battle music"},
		{"puzzle", "Contemplative puzzle music"},
		{"victory", "Triumphant victory music"},
	}

	for _, ctx := range contexts {
		fmt.Printf("\n%s: %s\n", ctx.name, ctx.description)
		
		composer.SetContext(ctx.name)
		
		if verbose {
			printLayerStates(composer)
		}

		track := composer.GenerateAdaptiveTrack(duration)
		
		stats := analyzeTrack(track)
		fmt.Printf("  Samples:     %d\n", len(track.Data))
		fmt.Printf("  Max Amp:     %.4f\n", stats.maxAmplitude)
		fmt.Printf("  Avg Amp:     %.4f\n", stats.avgAmplitude)
		fmt.Printf("  Has Audio:   %v\n", stats.hasAudio)
		fmt.Printf("  Clipping:    %v\n", stats.hasClipping)
		fmt.Printf("  Active Layers: %d\n", composer.GetActiveLayerCount())
	}
}

// testLayers tests individual layer activation/deactivation
func testLayers(composer *music.AdaptiveComposer, duration float64, verbose bool) {
	fmt.Println("Testing Layer Management")
	fmt.Println("========================")

	layers := []audio.MusicLayer{
		audio.MusicLayerBase,
		audio.MusicLayerMelody,
		audio.MusicLayerHarmony,
		audio.MusicLayerPercussion,
		audio.MusicLayerIntensity,
	}

	for _, layer := range layers {
		fmt.Printf("\nTesting layer: %s\n", layer.String())
		
		// Reset composer
		composer.Initialize("fantasy", 60)
		
		// Add only this layer
		err := composer.AddLayer(layer)
		if err != nil {
			fmt.Printf("  Error adding layer: %v\n", err)
			continue
		}
		
		// Update to activate layer
		composer.Update(1.0)
		
		if verbose {
			printLayerStates(composer)
		}
		
		track := composer.GenerateAdaptiveTrack(duration)
		stats := analyzeTrack(track)
		
		fmt.Printf("  Max Amp:       %.4f\n", stats.maxAmplitude)
		fmt.Printf("  Active Layers: %d\n", composer.GetActiveLayerCount())
		
		// Test removal
		err = composer.RemoveLayer(layer)
		if err != nil {
			fmt.Printf("  Error removing layer: %v\n", err)
		}
		
		// Update to deactivate
		for i := 0; i < 10; i++ {
			composer.Update(0.1)
		}
		
		afterRemoval := composer.GetActiveLayerCount()
		fmt.Printf("  After removal: %d layers\n", afterRemoval)
	}
}

// testTransitions tests smooth transitions between contexts
func testTransitions(composer *music.AdaptiveComposer, duration float64, verbose bool) {
	fmt.Println("Testing Smooth Transitions")
	fmt.Println("==========================")

	transitions := []struct {
		from string
		to   string
	}{
		{"exploration", "combat"},
		{"combat", "boss"},
		{"boss", "victory"},
		{"victory", "exploration"},
	}

	for _, trans := range transitions {
		fmt.Printf("\nTransition: %s → %s\n", trans.from, trans.to)
		
		// Set initial context
		composer.SetContext(trans.from)
		
		// Update to stabilize
		for i := 0; i < 10; i++ {
			composer.Update(0.1)
		}
		
		initialLayers := composer.GetActiveLayerCount()
		fmt.Printf("  Initial layers: %d\n", initialLayers)
		
		if verbose {
			fmt.Println("  Initial state:")
			printLayerStates(composer)
		}
		
		// Change context
		composer.SetContext(trans.to)
		
		// Simulate gradual transition
		start := time.Now()
		steps := 10
		for i := 0; i < steps; i++ {
			composer.Update(0.1)
			time.Sleep(10 * time.Millisecond) // Small delay to simulate real-time
		}
		elapsed := time.Since(start)
		
		finalLayers := composer.GetActiveLayerCount()
		fmt.Printf("  Final layers:   %d\n", finalLayers)
		fmt.Printf("  Transition time: %v\n", elapsed)
		
		if verbose {
			fmt.Println("  Final state:")
			printLayerStates(composer)
		}
	}
}

// testIntensity tests intensity scaling
func testIntensity(composer *music.AdaptiveComposer, duration float64, verbose bool) {
	fmt.Println("Testing Intensity Scaling")
	fmt.Println("=========================")

	composer.SetContext("combat")
	
	intensities := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	
	for _, intensity := range intensities {
		fmt.Printf("\nIntensity: %.2f\n", intensity)
		
		err := composer.UpdateIntensity(intensity)
		if err != nil {
			fmt.Printf("  Error setting intensity: %v\n", err)
			continue
		}
		
		// Update to apply intensity change
		for i := 0; i < 10; i++ {
			composer.Update(0.1)
		}
		
		intensityVolume := composer.GetLayerVolume("intensity")
		fmt.Printf("  Intensity layer volume: %.4f\n", intensityVolume)
		
		track := composer.GenerateAdaptiveTrack(duration)
		stats := analyzeTrack(track)
		fmt.Printf("  Max amplitude: %.4f\n", stats.maxAmplitude)
		
		if verbose {
			printLayerStates(composer)
		}
	}
}

// printLayerStates prints the current state of all layers
func printLayerStates(composer *music.AdaptiveComposer) {
	layers := []string{"ambient", "melody", "harmony", "percussion", "intensity"}
	
	for _, layerName := range layers {
		volume := composer.GetLayerVolume(layerName)
		if volume > 0.01 {
			fmt.Printf("    %-12s: Volume %.3f\n", layerName, volume)
		}
	}
}

// trackStats holds statistics about an audio track
type trackStats struct {
	maxAmplitude float64
	avgAmplitude float64
	hasAudio     bool
	hasClipping  bool
}

// analyzeTrack computes statistics for an audio track
func analyzeTrack(track *audio.AudioSample) trackStats {
	stats := trackStats{}
	
	if len(track.Data) == 0 {
		return stats
	}
	
	sum := 0.0
	for _, sample := range track.Data {
		abs := sample
		if abs < 0 {
			abs = -abs
		}
		
		if abs > stats.maxAmplitude {
			stats.maxAmplitude = abs
		}
		
		if abs > 1.0 {
			stats.hasClipping = true
		}
		
		if sample != 0.0 {
			stats.hasAudio = true
		}
		
		sum += abs
	}
	
	stats.avgAmplitude = sum / float64(len(track.Data))
	
	return stats
}
