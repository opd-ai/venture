// Package main provides a CLI tool for testing enhanced music composition.package compositiontest

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
)

func main() {
	genreFlag := flag.String("genre", "fantasy", "Music genre (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)")
	contextFlag := flag.String("context", "exploration", "Music context (exploration, combat, boss, puzzle, victory)")
	durationFlag := flag.Float64("duration", 4.0, "Track duration in seconds")
	
	flag.Parse()

	fmt.Printf("Enhanced Composition Test\n")
	fmt.Printf("=========================\n\n")
	fmt.Printf("Genre: %s\n", *genreFlag)
	fmt.Printf("Context: %s\n", *contextFlag)
	fmt.Printf("Duration: %.1fs\n\n", *durationFlag)

	// Create composer
	composer := music.NewAdaptiveComposer(44100, 12345)
	composer.Initialize(*genreFlag, 120)

	// Create manager for interface
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize(*genreFlag, 120)

	// Set context
	ctx := audio.MusicContext{
		Location:   *contextFlag,
		Combat:     *contextFlag == "combat" || *contextFlag == "boss",
		BossNearby: *contextFlag == "boss",
		TimeOfDay:  "day",
		Danger:     0.0,
	}
	
	if *contextFlag == "combat" {
		ctx.Danger = 0.6
	} else if *contextFlag == "boss" {
		ctx.Danger = 1.0
	}

	err := manager.SetContext(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting context: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating track with enhanced composition...")
	fmt.Println("\nEnhancements:")
	fmt.Println("  • Melody: Patterns (ascending, descending, arpeggio, wave, repeat)")
	fmt.Println("  • Harmony: Chord progressions (5 types based on context)")
	fmt.Println("  • Percussion: Genre-specific drum patterns")
	fmt.Println("")

	// Generate track
	track := manager.GenerateTrack(*durationFlag)

	// Analyze track
	if track != nil && len(track.Data) > 0 {
		fmt.Printf("Track Generated Successfully!\n\n")
		fmt.Printf("Sample Rate: %d Hz\n", track.SampleRate)
		fmt.Printf("Samples: %d\n", len(track.Data))
		fmt.Printf("Duration: %.2f seconds\n", float64(len(track.Data))/float64(track.SampleRate))
		
		// Calculate RMS amplitude
		var sumSquares float64
		for _, sample := range track.Data {
			sumSquares += sample * sample
		}
		rms := 0.0
		if len(track.Data) > 0 {
			rms = sumSquares / float64(len(track.Data))
		}
		fmt.Printf("RMS Amplitude: %.4f\n", rms)
		
		// Find peak amplitude
		peak := 0.0
		for _, sample := range track.Data {
			if abs := sample; abs < 0 {
				abs = -abs
			} else if abs > peak {
				peak = abs
			}
		}
		fmt.Printf("Peak Amplitude: %.4f\n", peak)
		
		// Check for clipping
		if peak > 0.99 {
			fmt.Println("\n⚠️  Warning: Track may be clipping!")
		} else {
			fmt.Println("\n✓ Track amplitude within safe range")
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate track\n")
		os.Exit(1)
	}

	fmt.Println("\n" + "Test complete!")
}
