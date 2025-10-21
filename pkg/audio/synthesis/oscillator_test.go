package synthesis

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/audio"
)

func TestOscillator_Generate(t *testing.T) {
	tests := []struct {
		name      string
		waveform  audio.WaveformType
		frequency float64
		duration  float64
		wantLen   int
	}{
		{
			name:      "sine wave 440Hz 1 second",
			waveform:  audio.WaveformSine,
			frequency: 440.0,
			duration:  1.0,
			wantLen:   44100,
		},
		{
			name:      "square wave 220Hz 0.5 seconds",
			waveform:  audio.WaveformSquare,
			frequency: 220.0,
			duration:  0.5,
			wantLen:   22050,
		},
		{
			name:      "sawtooth wave 880Hz 0.25 seconds",
			waveform:  audio.WaveformSawtooth,
			frequency: 880.0,
			duration:  0.25,
			wantLen:   11025,
		},
		{
			name:      "triangle wave 110Hz 2 seconds",
			waveform:  audio.WaveformTriangle,
			frequency: 110.0,
			duration:  2.0,
			wantLen:   88200,
		},
		{
			name:      "noise 0.1 seconds",
			waveform:  audio.WaveformNoise,
			frequency: 0,
			duration:  0.1,
			wantLen:   4410,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osc := NewOscillator(44100, 12345)
			sample := osc.Generate(tt.waveform, tt.frequency, tt.duration)

			if sample == nil {
				t.Fatal("Generate returned nil")
			}

			if sample.SampleRate != 44100 {
				t.Errorf("SampleRate = %d, want 44100", sample.SampleRate)
			}

			if len(sample.Data) != tt.wantLen {
				t.Errorf("len(Data) = %d, want %d", len(sample.Data), tt.wantLen)
			}

			// Check that samples are in valid range [-1, 1]
			for i, v := range sample.Data {
				if v < -1.0 || v > 1.0 {
					t.Errorf("sample[%d] = %f, out of range [-1, 1]", i, v)
					break
				}
			}
		})
	}
}

func TestOscillator_GenerateNote(t *testing.T) {
	tests := []struct {
		name     string
		note     audio.Note
		waveform audio.WaveformType
	}{
		{
			name: "A4 quarter note full velocity",
			note: audio.Note{
				Frequency: 440.0,
				Duration:  0.5,
				Velocity:  1.0,
			},
			waveform: audio.WaveformSine,
		},
		{
			name: "C4 half note soft velocity",
			note: audio.Note{
				Frequency: 261.63,
				Duration:  1.0,
				Velocity:  0.5,
			},
			waveform: audio.WaveformTriangle,
		},
		{
			name: "E5 eighth note medium velocity",
			note: audio.Note{
				Frequency: 659.25,
				Duration:  0.25,
				Velocity:  0.75,
			},
			waveform: audio.WaveformSquare,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osc := NewOscillator(44100, 12345)
			sample := osc.GenerateNote(tt.note, tt.waveform)

			if sample == nil {
				t.Fatal("GenerateNote returned nil")
			}

			expectedLen := int(44100 * tt.note.Duration)
			if len(sample.Data) != expectedLen {
				t.Errorf("len(Data) = %d, want %d", len(sample.Data), expectedLen)
			}

			// Check that velocity was applied (samples should be scaled)
			maxAmplitude := 0.0
			for _, v := range sample.Data {
				if math.Abs(v) > maxAmplitude {
					maxAmplitude = math.Abs(v)
				}
			}

			// Max amplitude should be approximately equal to velocity
			if math.Abs(maxAmplitude-tt.note.Velocity) > 0.1 {
				t.Errorf("maxAmplitude = %f, want ~%f", maxAmplitude, tt.note.Velocity)
			}
		})
	}
}

func TestOscillator_Determinism(t *testing.T) {
	seed := int64(54321)
	
	osc1 := NewOscillator(44100, seed)
	sample1 := osc1.Generate(audio.WaveformNoise, 0, 0.1)
	
	osc2 := NewOscillator(44100, seed)
	sample2 := osc2.Generate(audio.WaveformNoise, 0, 0.1)
	
	if len(sample1.Data) != len(sample2.Data) {
		t.Fatal("samples have different lengths")
	}
	
	for i := range sample1.Data {
		if sample1.Data[i] != sample2.Data[i] {
			t.Errorf("sample mismatch at index %d: %f != %f", i, sample1.Data[i], sample2.Data[i])
			break
		}
	}
}

func TestOscillator_WaveformCharacteristics(t *testing.T) {
	osc := NewOscillator(44100, 12345)

	t.Run("sine wave is smooth", func(t *testing.T) {
		sample := osc.Generate(audio.WaveformSine, 440.0, 0.1)
		
		// Sine wave should have no abrupt changes
		for i := 1; i < len(sample.Data); i++ {
			diff := math.Abs(sample.Data[i] - sample.Data[i-1])
			if diff > 0.1 {
				t.Errorf("abrupt change at index %d: %f", i, diff)
				break
			}
		}
	})

	t.Run("square wave has sharp transitions", func(t *testing.T) {
		sample := osc.Generate(audio.WaveformSquare, 440.0, 0.01)
		
		// Square wave values should be close to -1 or 1
		nearExtremes := 0
		for _, v := range sample.Data {
			if math.Abs(v) > 0.9 {
				nearExtremes++
			}
		}
		
		// Most samples should be near extremes
		percentage := float64(nearExtremes) / float64(len(sample.Data))
		if percentage < 0.8 {
			t.Errorf("only %f%% of samples near extremes, want >80%%", percentage*100)
		}
	})

	t.Run("triangle wave is linear", func(t *testing.T) {
		sample := osc.Generate(audio.WaveformTriangle, 440.0, 0.01)
		
		// Triangle wave should have constant slope within each ramp
		// Just verify it oscillates
		hasPositive := false
		hasNegative := false
		for _, v := range sample.Data {
			if v > 0.5 {
				hasPositive = true
			}
			if v < -0.5 {
				hasNegative = true
			}
		}
		
		if !hasPositive || !hasNegative {
			t.Error("triangle wave doesn't oscillate properly")
		}
	})
}

func TestEnvelope_Apply(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope
		duration float64
	}{
		{
			name:     "default envelope",
			envelope: DefaultEnvelope(),
			duration: 1.0,
		},
		{
			name: "fast attack slow release",
			envelope: Envelope{
				Attack:  0.001,
				Decay:   0.01,
				Sustain: 0.8,
				Release: 0.5,
			},
			duration: 1.0,
		},
		{
			name: "no sustain",
			envelope: Envelope{
				Attack:  0.1,
				Decay:   0.1,
				Sustain: 0.0,
				Release: 0.1,
			},
			duration: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			osc := NewOscillator(44100, 12345)
			sample := osc.Generate(audio.WaveformSine, 440.0, tt.duration)
			
			originalData := make([]float64, len(sample.Data))
			copy(originalData, sample.Data)
			
			tt.envelope.Apply(sample.Data, sample.SampleRate)
			
			// Check that envelope was applied (data changed)
			changed := false
			for i := range sample.Data {
				if sample.Data[i] != originalData[i] {
					changed = true
					break
				}
			}
			
			if !changed {
				t.Error("envelope did not modify the sample")
			}
			
			// Check that first sample is close to 0 (attack starts at 0)
			if math.Abs(sample.Data[0]) > 0.1 {
				t.Errorf("first sample = %f, want ~0", sample.Data[0])
			}
			
			// Check that last sample is close to 0 (release ends at 0)
			lastIdx := len(sample.Data) - 1
			if math.Abs(sample.Data[lastIdx]) > 0.1 {
				t.Errorf("last sample = %f, want ~0", sample.Data[lastIdx])
			}
		})
	}
}

func TestEnvelope_EmptyData(t *testing.T) {
	env := DefaultEnvelope()
	data := []float64{}
	
	// Should not panic
	env.Apply(data, 44100)
}

func BenchmarkOscillator_GenerateSine(b *testing.B) {
	osc := NewOscillator(44100, 12345)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		osc.Generate(audio.WaveformSine, 440.0, 1.0)
	}
}

func BenchmarkOscillator_GenerateNoise(b *testing.B) {
	osc := NewOscillator(44100, 12345)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		osc.Generate(audio.WaveformNoise, 0, 1.0)
	}
}

func BenchmarkEnvelope_Apply(b *testing.B) {
	osc := NewOscillator(44100, 12345)
	sample := osc.Generate(audio.WaveformSine, 440.0, 1.0)
	env := DefaultEnvelope()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Apply(sample.Data, sample.SampleRate)
	}
}
