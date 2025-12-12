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

	printHeader(*genreFlag, *contextFlag, *durationFlag)

	manager := initializeMusicManager(*genreFlag)
	ctx := createMusicContext(*contextFlag)

	if err := manager.SetContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting context: %v\n", err)
		os.Exit(1)
	}

	printEnhancements()
	track := manager.GenerateTrack(*durationFlag)

	if track == nil || len(track.Data) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Failed to generate track\n")
		os.Exit(1)
	}

	analyzeAndPrintTrack(track)
	fmt.Println("\n" + "Test complete!")
}

// printHeader prints the test header with configuration.
func printHeader(genre, context string, duration float64) {
	fmt.Printf("Enhanced Composition Test\n")
	fmt.Printf("=========================\n\n")
	fmt.Printf("Genre: %s\n", genre)
	fmt.Printf("Context: %s\n", context)
	fmt.Printf("Duration: %.1fs\n\n", duration)
}

// initializeMusicManager creates and initializes the music manager.
func initializeMusicManager(genre string) *music.AdaptiveMusicManager {
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize(genre, 120)
	return manager
}

// createMusicContext creates a music context based on the context flag.
func createMusicContext(contextFlag string) audio.MusicContext {
	ctx := audio.MusicContext{
		Location:   contextFlag,
		Combat:     contextFlag == "combat" || contextFlag == "boss",
		BossNearby: contextFlag == "boss",
		TimeOfDay:  "day",
		Danger:     0.0,
	}

	if contextFlag == "combat" {
		ctx.Danger = 0.6
	} else if contextFlag == "boss" {
		ctx.Danger = 1.0
	}

	return ctx
}

// printEnhancements prints the composition enhancements description.
func printEnhancements() {
	fmt.Println("Generating track with enhanced composition...")
	fmt.Println("\nEnhancements:")
	fmt.Println("  • Melody: Patterns (ascending, descending, arpeggio, wave, repeat)")
	fmt.Println("  • Harmony: Chord progressions (5 types based on context)")
	fmt.Println("  • Percussion: Genre-specific drum patterns")
	fmt.Println("")
}

// analyzeAndPrintTrack analyzes and prints track metrics.
func analyzeAndPrintTrack(track *audio.AudioSample) {
	fmt.Printf("Track Generated Successfully!\n\n")
	fmt.Printf("Sample Rate: %d Hz\n", track.SampleRate)
	fmt.Printf("Samples: %d\n", len(track.Data))
	fmt.Printf("Duration: %.2f seconds\n", float64(len(track.Data))/float64(track.SampleRate))

	rms, peak := calculateAmplitudeMetrics(track.Data)
	fmt.Printf("RMS Amplitude: %.4f\n", rms)
	fmt.Printf("Peak Amplitude: %.4f\n", peak)

	if peak > 0.99 {
		fmt.Println("\n⚠️  Warning: Track may be clipping!")
	} else {
		fmt.Println("\n✓ Track amplitude within safe range")
	}
}

// calculateAmplitudeMetrics computes RMS and peak amplitude from sample data.
func calculateAmplitudeMetrics(data []float64) (rms, peak float64) {
	var sumSquares float64
	for _, sample := range data {
		sumSquares += sample * sample
		abs := sample
		if abs < 0 {
			abs = -abs
		}
		if abs > peak {
			peak = abs
		}
	}

	if len(data) > 0 {
		rms = sumSquares / float64(len(data))
	}

	return rms, peak
}
