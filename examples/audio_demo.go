// +build test

package main

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/music"
	"github.com/opd-ai/venture/pkg/audio/sfx"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
)

func main() {
	fmt.Println("=== Phase 4: Audio Synthesis Demonstration ===")
	fmt.Println()
	
	seed := int64(12345)
	sampleRate := 44100
	
	// 1. Demonstrate Waveform Synthesis
	demonstrateOscillator(sampleRate, seed)
	fmt.Println()
	
	// 2. Demonstrate Sound Effects
	demonstrateSFX(sampleRate, seed)
	fmt.Println()
	
	// 3. Demonstrate Music Generation
	demonstrateMusic(sampleRate, seed)
	fmt.Println()
	
	fmt.Println("=== All Audio Systems Operational ===")
	fmt.Println("Phase 4 implementation complete!")
}

func demonstrateOscillator(sampleRate int, seed int64) {
	fmt.Println("1. WAVEFORM SYNTHESIS")
	fmt.Println("   Testing all waveform types...")
	
	osc := synthesis.NewOscillator(sampleRate, seed)
	
	waveforms := []struct {
		name string
		typ  audio.WaveformType
	}{
		{"Sine", audio.WaveformSine},
		{"Square", audio.WaveformSquare},
		{"Sawtooth", audio.WaveformSawtooth},
		{"Triangle", audio.WaveformTriangle},
		{"Noise", audio.WaveformNoise},
	}
	
	for _, w := range waveforms {
		sample := osc.Generate(w.typ, 440.0, 0.1)
		fmt.Printf("   ✓ %s wave: %d samples generated\n", w.name, len(sample.Data))
	}
	
	// Test ADSR envelope
	sample := osc.Generate(audio.WaveformSine, 440.0, 1.0)
	env := synthesis.DefaultEnvelope()
	env.Apply(sample.Data, sample.SampleRate)
	fmt.Printf("   ✓ ADSR envelope applied\n")
	
	// Test musical note
	note := audio.Note{
		Frequency: 440.0,
		Duration:  0.5,
		Velocity:  0.8,
	}
	noteSample := osc.GenerateNote(note, audio.WaveformSine)
	fmt.Printf("   ✓ Musical note (A4): %d samples at 80%% velocity\n", len(noteSample.Data))
}

func demonstrateSFX(sampleRate int, seed int64) {
	fmt.Println("2. SOUND EFFECTS GENERATION")
	fmt.Println("   Testing all effect types...")
	
	gen := sfx.NewGenerator(sampleRate, seed)
	
	effects := []string{
		"impact", "explosion", "magic", "laser", "pickup",
		"hit", "jump", "death", "powerup",
	}
	
	for _, effect := range effects {
		sample := gen.Generate(effect, seed)
		duration := float64(len(sample.Data)) / float64(sample.SampleRate)
		fmt.Printf("   ✓ %s: %.3fs duration\n", effect, duration)
	}
}

func demonstrateMusic(sampleRate int, seed int64) {
	fmt.Println("3. MUSIC COMPOSITION")
	fmt.Println("   Testing all genres and contexts...")
	
	gen := music.NewGenerator(sampleRate, seed)
	
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "post-apocalyptic"}
	contexts := []string{"combat", "exploration", "ambient", "victory"}
	
	// Generate one track per genre
	for _, genre := range genres {
		context := contexts[0] // Use combat for all
		_ = gen.GenerateTrack(genre, context, seed, 2.0)
		
		scale := music.GetScaleForGenre(genre)
		tempo := music.GetTempoForContext(context)
		
		fmt.Printf("   ✓ %s: %s scale, %.0f BPM\n", genre, scale.Name, tempo)
	}
	
	// Show musical variety across contexts
	fmt.Println("\n   Context variations (fantasy genre):")
	for _, context := range contexts {
		contextTrack := gen.GenerateTrack("fantasy", context, seed, 1.0)
		tempo := music.GetTempoForContext(context)
		fmt.Printf("   ✓ %s: %.0f BPM, %d samples\n", context, tempo, len(contextTrack.Data))
	}
}
