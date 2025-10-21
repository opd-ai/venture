package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
	"github.com/opd-ai/venture/pkg/audio/sfx"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
)

func main() {
	// Parse command line flags
	testType := flag.String("type", "sfx", "Type of audio to test: oscillator, sfx, music")
	genre := flag.String("genre", "fantasy", "Genre for music generation")
	context := flag.String("context", "combat", "Context for music generation")
	effectType := flag.String("effect", "impact", "Effect type for SFX")
	waveform := flag.String("waveform", "sine", "Waveform type for oscillator")
	frequency := flag.Float64("frequency", 440.0, "Frequency in Hz for oscillator")
	duration := flag.Float64("duration", 1.0, "Duration in seconds")
	seed := flag.Int64("seed", 12345, "Random seed for generation")
	verbose := flag.Bool("verbose", false, "Show verbose output")
	
	flag.Parse()
	
	sampleRate := 44100
	
	switch *testType {
	case "oscillator":
		testOscillator(sampleRate, *seed, *waveform, *frequency, *duration, *verbose)
	case "sfx":
		testSFX(sampleRate, *seed, *effectType, *verbose)
	case "music":
		testMusic(sampleRate, *seed, *genre, *context, *duration, *verbose)
	default:
		fmt.Fprintf(os.Stderr, "Unknown test type: %s\n", *testType)
		flag.Usage()
		os.Exit(1)
	}
}

func testOscillator(sampleRate int, seed int64, waveformStr string, frequency, duration float64, verbose bool) {
	fmt.Printf("=== Testing Oscillator ===\n")
	fmt.Printf("Waveform: %s\n", waveformStr)
	fmt.Printf("Frequency: %.2f Hz\n", frequency)
	fmt.Printf("Duration: %.2f seconds\n", duration)
	fmt.Printf("Seed: %d\n\n", seed)
	
	osc := synthesis.NewOscillator(sampleRate, seed)
	
	var waveform audio.WaveformType
	switch waveformStr {
	case "sine":
		waveform = audio.WaveformSine
	case "square":
		waveform = audio.WaveformSquare
	case "sawtooth":
		waveform = audio.WaveformSawtooth
	case "triangle":
		waveform = audio.WaveformTriangle
	case "noise":
		waveform = audio.WaveformNoise
	default:
		fmt.Fprintf(os.Stderr, "Unknown waveform: %s\n", waveformStr)
		os.Exit(1)
	}
	
	sample := osc.Generate(waveform, frequency, duration)
	
	fmt.Printf("Generated:\n")
	fmt.Printf("  Sample Rate: %d Hz\n", sample.SampleRate)
	fmt.Printf("  Samples: %d\n", len(sample.Data))
	fmt.Printf("  Duration: %.3f seconds\n", float64(len(sample.Data))/float64(sample.SampleRate))
	
	if verbose {
		printSampleStats(sample)
	}
}

func testSFX(sampleRate int, seed int64, effectType string, verbose bool) {
	fmt.Printf("=== Testing Sound Effects ===\n")
	fmt.Printf("Effect Type: %s\n", effectType)
	fmt.Printf("Seed: %d\n\n", seed)
	
	gen := sfx.NewGenerator(sampleRate, seed)
	sample := gen.Generate(effectType, seed)
	
	fmt.Printf("Generated:\n")
	fmt.Printf("  Sample Rate: %d Hz\n", sample.SampleRate)
	fmt.Printf("  Samples: %d\n", len(sample.Data))
	fmt.Printf("  Duration: %.3f seconds\n", float64(len(sample.Data))/float64(sample.SampleRate))
	
	if verbose {
		printSampleStats(sample)
		fmt.Println("\nAvailable effects:")
		effects := []string{"impact", "explosion", "magic", "laser", "pickup", "hit", "jump", "death", "powerup"}
		for _, e := range effects {
			fmt.Printf("  - %s\n", e)
		}
	}
}

func testMusic(sampleRate int, seed int64, genre, context string, duration float64, verbose bool) {
	fmt.Printf("=== Testing Music Generation ===\n")
	fmt.Printf("Genre: %s\n", genre)
	fmt.Printf("Context: %s\n", context)
	fmt.Printf("Duration: %.2f seconds\n", duration)
	fmt.Printf("Seed: %d\n\n", seed)
	
	gen := music.NewGenerator(sampleRate, seed)
	sample := gen.GenerateTrack(genre, context, seed, duration)
	
	fmt.Printf("Generated:\n")
	fmt.Printf("  Sample Rate: %d Hz\n", sample.SampleRate)
	fmt.Printf("  Samples: %d\n", len(sample.Data))
	fmt.Printf("  Duration: %.3f seconds\n", float64(len(sample.Data))/float64(sample.SampleRate))
	
	if verbose {
		printSampleStats(sample)
		
		scale := music.GetScaleForGenre(genre)
		tempo := music.GetTempoForContext(context)
		
		fmt.Printf("\nMusical Properties:\n")
		fmt.Printf("  Scale: %s (%d notes)\n", scale.Name, len(scale.Intervals))
		fmt.Printf("  Tempo: %.0f BPM\n", tempo)
		
		fmt.Println("\nAvailable genres:")
		genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}
		for _, g := range genres {
			fmt.Printf("  - %s\n", g)
		}
		
		fmt.Println("\nAvailable contexts:")
		contexts := []string{"combat", "exploration", "ambient", "victory"}
		for _, c := range contexts {
			fmt.Printf("  - %s\n", c)
		}
	}
}

func printSampleStats(sample *audio.AudioSample) {
	var min, max, sum, sumSquares float64
	min = sample.Data[0]
	max = sample.Data[0]
	
	for _, v := range sample.Data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
		sumSquares += v * v
	}
	
	mean := sum / float64(len(sample.Data))
	rms := sumSquares / float64(len(sample.Data))
	
	fmt.Printf("\nSample Statistics:\n")
	fmt.Printf("  Min: %.6f\n", min)
	fmt.Printf("  Max: %.6f\n", max)
	fmt.Printf("  Mean: %.6f\n", mean)
	fmt.Printf("  RMS: %.6f\n", rms)
}
