package synthesis

import (
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
)

// TestEngine_ImplementsSynthesizer verifies that Engine implements audio.Synthesizer.
func TestEngine_ImplementsSynthesizer(t *testing.T) {
	engine := NewEngine(12345)

	// This will fail to compile if Engine doesn't implement audio.Synthesizer
	var synth audio.Synthesizer = engine
	if synth == nil {
		t.Fatal("Engine should implement audio.Synthesizer")
	}

	// Verify interface methods work correctly
	sample := synth.Generate(audio.WaveformSine, 440.0, 0.5)
	if sample == nil {
		t.Fatal("Synthesizer.Generate returned nil")
	}

	note := audio.Note{Frequency: 440.0, Duration: 0.5, Velocity: 0.8}
	noteSample := synth.GenerateNote(note, audio.WaveformSine)
	if noteSample == nil {
		t.Fatal("Synthesizer.GenerateNote returned nil")
	}
}

func TestNewEngine(t *testing.T) {
	engine := NewEngine(12345)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.GetSampleRate() != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", engine.GetSampleRate())
	}

	if engine.GetSeed() != 12345 {
		t.Errorf("Expected seed 12345, got %d", engine.GetSeed())
	}
}

func TestNewEngineWithSampleRate(t *testing.T) {
	engine := NewEngineWithSampleRate(48000, 54321)

	if engine.GetSampleRate() != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", engine.GetSampleRate())
	}

	if engine.GetSeed() != 54321 {
		t.Errorf("Expected seed 54321, got %d", engine.GetSeed())
	}
}

func TestNewEngineWithSampleRate_InvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		sampleRate   int
		expectedRate int
	}{
		{
			name:         "zero sample rate defaults to 44100",
			sampleRate:   0,
			expectedRate: 44100,
		},
		{
			name:         "negative sample rate defaults to 44100",
			sampleRate:   -1000,
			expectedRate: 44100,
		},
		{
			name:         "very negative sample rate defaults to 44100",
			sampleRate:   -999999,
			expectedRate: 44100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngineWithSampleRate(tt.sampleRate, 12345)

			if engine == nil {
				t.Fatal("NewEngineWithSampleRate returned nil")
			}

			if engine.GetSampleRate() != tt.expectedRate {
				t.Errorf("Expected sample rate %d, got %d", tt.expectedRate, engine.GetSampleRate())
			}

			// Verify oscillator was created with corrected sample rate
			sample := engine.Generate(audio.WaveformSine, 440.0, 0.1)
			if sample == nil {
				t.Fatal("Engine with corrected sample rate failed to generate tone")
			}
			if sample.SampleRate != tt.expectedRate {
				t.Errorf("Generated sample has rate %d, expected %d", sample.SampleRate, tt.expectedRate)
			}
		})
	}
}

func TestEngine_Generate(t *testing.T) {
	engine := NewEngine(12345)

	tests := []struct {
		name      string
		waveform  audio.WaveformType
		frequency float64
		duration  float64
	}{
		{"sine_440hz", audio.WaveformSine, 440.0, 0.5},
		{"square_880hz", audio.WaveformSquare, 880.0, 0.25},
		{"sawtooth_220hz", audio.WaveformSawtooth, 220.0, 1.0},
		{"triangle_330hz", audio.WaveformTriangle, 330.0, 0.1},
		{"noise_short", audio.WaveformNoise, 0.0, 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := engine.Generate(tt.waveform, tt.frequency, tt.duration)

			if sample == nil {
				t.Fatal("Generate returned nil")
			}

			expectedSamples := int(float64(engine.GetSampleRate()) * tt.duration)
			if len(sample.Data) != expectedSamples {
				t.Errorf("Expected %d samples, got %d", expectedSamples, len(sample.Data))
			}

			if sample.SampleRate != engine.GetSampleRate() {
				t.Errorf("Expected sample rate %d, got %d", engine.GetSampleRate(), sample.SampleRate)
			}
		})
	}
}

func TestEngine_GenerateToneWithEnvelope(t *testing.T) {
	engine := NewEngine(12345)
	env := DefaultEnvelope()

	sample := engine.GenerateToneWithEnvelope(audio.WaveformSine, 440.0, 1.0, env)

	if sample == nil {
		t.Fatal("GenerateToneWithEnvelope returned nil")
	}

	// Check that envelope was applied (attack phase starts at 0)
	if sample.Data[0] != 0.0 {
		t.Errorf("Expected first sample to be 0 after envelope attack, got %f", sample.Data[0])
	}
}

func TestEngine_GenerateNote(t *testing.T) {
	engine := NewEngine(12345)

	note := audio.Note{
		Frequency: 440.0,
		Duration:  0.5,
		Velocity:  0.8,
	}

	sample := engine.GenerateNote(note, audio.WaveformSine)

	if sample == nil {
		t.Fatal("GenerateNote returned nil")
	}

	expectedSamples := int(float64(engine.GetSampleRate()) * note.Duration)
	if len(sample.Data) != expectedSamples {
		t.Errorf("Expected %d samples, got %d", expectedSamples, len(sample.Data))
	}
}

func TestEngine_GenerateNoteWithEnvelope(t *testing.T) {
	engine := NewEngine(12345)
	env := DefaultEnvelope()

	note := audio.Note{
		Frequency: 440.0,
		Duration:  1.0,
		Velocity:  1.0,
	}

	sample := engine.GenerateNoteWithEnvelope(note, audio.WaveformSine, env)

	if sample == nil {
		t.Fatal("GenerateNoteWithEnvelope returned nil")
	}

	// Envelope should start at 0
	if sample.Data[0] != 0.0 {
		t.Errorf("Expected first sample to be 0, got %f", sample.Data[0])
	}
}

func TestEngine_GenerateChord(t *testing.T) {
	engine := NewEngine(12345)

	// C major chord (C4, E4, G4)
	notes := []audio.Note{
		{Frequency: 261.63, Duration: 0.5, Velocity: 0.8},
		{Frequency: 329.63, Duration: 0.5, Velocity: 0.8},
		{Frequency: 392.00, Duration: 0.5, Velocity: 0.8},
	}

	sample := engine.GenerateChord(notes, audio.WaveformSine)

	if sample == nil {
		t.Fatal("GenerateChord returned nil")
	}

	expectedSamples := int(float64(engine.GetSampleRate()) * 0.5)
	if len(sample.Data) != expectedSamples {
		t.Errorf("Expected %d samples, got %d", expectedSamples, len(sample.Data))
	}

	// Check normalization: all values should be in [-1, 1] range
	for i, v := range sample.Data {
		if v < -1.0 || v > 1.0 {
			t.Errorf("Sample %d out of range: %f", i, v)
			break
		}
	}
}

func TestEngine_GenerateChord_Empty(t *testing.T) {
	engine := NewEngine(12345)

	sample := engine.GenerateChord([]audio.Note{}, audio.WaveformSine)

	if sample == nil {
		t.Fatal("GenerateChord returned nil for empty notes")
	}

	if len(sample.Data) != 0 {
		t.Errorf("Expected 0 samples for empty chord, got %d", len(sample.Data))
	}
}

func TestEngine_GenerateChordWithEnvelope(t *testing.T) {
	engine := NewEngine(12345)
	env := DefaultEnvelope()

	notes := []audio.Note{
		{Frequency: 261.63, Duration: 1.0, Velocity: 1.0},
		{Frequency: 329.63, Duration: 1.0, Velocity: 1.0},
	}

	sample := engine.GenerateChordWithEnvelope(notes, audio.WaveformSine, env)

	if sample == nil {
		t.Fatal("GenerateChordWithEnvelope returned nil")
	}

	// Attack phase should start at 0
	if sample.Data[0] != 0.0 {
		t.Errorf("Expected first sample to be 0, got %f", sample.Data[0])
	}
}

func TestEngine_MixSamples(t *testing.T) {
	engine := NewEngine(12345)

	sample1 := engine.Generate(audio.WaveformSine, 440.0, 0.5)
	sample2 := engine.Generate(audio.WaveformSquare, 880.0, 0.3)

	mixed := engine.MixSamples([]*audio.AudioSample{sample1, sample2})

	if mixed == nil {
		t.Fatal("MixSamples returned nil")
	}

	// Mixed sample should be length of longest input
	if len(mixed.Data) != len(sample1.Data) {
		t.Errorf("Expected %d samples, got %d", len(sample1.Data), len(mixed.Data))
	}
}

func TestEngine_MixSamples_Empty(t *testing.T) {
	engine := NewEngine(12345)

	mixed := engine.MixSamples([]*audio.AudioSample{})

	if mixed == nil {
		t.Fatal("MixSamples returned nil for empty input")
	}

	if len(mixed.Data) != 0 {
		t.Errorf("Expected 0 samples for empty mix, got %d", len(mixed.Data))
	}
}

func TestEngine_ApplyEnvelope(t *testing.T) {
	engine := NewEngine(12345)
	env := DefaultEnvelope()

	sample := engine.Generate(audio.WaveformSine, 440.0, 1.0)
	originalFirst := sample.Data[0]

	engine.ApplyEnvelope(sample, env)

	// After envelope, first sample should be 0 (attack starts at 0)
	if sample.Data[0] != 0.0 {
		t.Errorf("Expected 0 after envelope, got %f (was %f)", sample.Data[0], originalFirst)
	}
}

func TestEngine_Determinism(t *testing.T) {
	engine1 := NewEngine(12345)
	engine2 := NewEngine(12345)

	sample1 := engine1.Generate(audio.WaveformSine, 440.0, 0.5)
	sample2 := engine2.Generate(audio.WaveformSine, 440.0, 0.5)

	if len(sample1.Data) != len(sample2.Data) {
		t.Fatalf("Samples have different lengths: %d vs %d", len(sample1.Data), len(sample2.Data))
	}

	for i := range sample1.Data {
		if sample1.Data[i] != sample2.Data[i] {
			t.Errorf("Sample %d differs: %f vs %f", i, sample1.Data[i], sample2.Data[i])
			break
		}
	}
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	engine := NewEngine(12345)

	done := make(chan bool, 10)

	// Run multiple goroutines accessing the engine
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = engine.Generate(audio.WaveformSine, 440.0, 0.01)
				_ = engine.GetSampleRate()
				_ = engine.GetSeed()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestEngine_ConcurrentEnvelopeMethods(t *testing.T) {
	engine := NewEngine(12345)
	env := DefaultEnvelope()

	done := make(chan bool, 10)

	// Run multiple goroutines calling envelope methods concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				// Test GenerateChordWithEnvelope concurrent access
				notes := []audio.Note{
					{Frequency: 261.63, Duration: 0.05, Velocity: 0.8},
					{Frequency: 329.63, Duration: 0.05, Velocity: 0.8},
				}
				_ = engine.GenerateChordWithEnvelope(notes, audio.WaveformSine, env)

				// Test ApplyEnvelope concurrent access (each goroutine uses its own sample)
				sample := engine.Generate(audio.WaveformSine, 440.0, 0.05)
				engine.ApplyEnvelope(sample, env)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkEngine_Generate(b *testing.B) {
	engine := NewEngine(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Generate(audio.WaveformSine, 440.0, 0.1)
	}
}

func BenchmarkEngine_GenerateChord(b *testing.B) {
	engine := NewEngine(12345)
	notes := []audio.Note{
		{Frequency: 261.63, Duration: 0.1, Velocity: 0.8},
		{Frequency: 329.63, Duration: 0.1, Velocity: 0.8},
		{Frequency: 392.00, Duration: 0.1, Velocity: 0.8},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.GenerateChord(notes, audio.WaveformSine)
	}
}
