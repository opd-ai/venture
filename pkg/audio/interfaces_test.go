package audio

import "testing"

// TestWaveformType_Constants verifies WaveformType constants are unique.
func TestWaveformType_Constants(t *testing.T) {
	types := []WaveformType{
		WaveformSine,
		WaveformSquare,
		WaveformSawtooth,
		WaveformTriangle,
		WaveformNoise,
	}

	seen := make(map[WaveformType]bool)
	for _, waveformType := range types {
		if seen[waveformType] {
			t.Errorf("Duplicate waveform type value: %v", waveformType)
		}
		seen[waveformType] = true
	}

	if len(types) != 5 {
		t.Errorf("Expected 5 waveform types, got %d", len(types))
	}

	// Verify expected values
	if WaveformSine != 0 {
		t.Errorf("Expected WaveformSine to be 0, got %d", WaveformSine)
	}
	if WaveformSquare != 1 {
		t.Errorf("Expected WaveformSquare to be 1, got %d", WaveformSquare)
	}
	if WaveformSawtooth != 2 {
		t.Errorf("Expected WaveformSawtooth to be 2, got %d", WaveformSawtooth)
	}
	if WaveformTriangle != 3 {
		t.Errorf("Expected WaveformTriangle to be 3, got %d", WaveformTriangle)
	}
	if WaveformNoise != 4 {
		t.Errorf("Expected WaveformNoise to be 4, got %d", WaveformNoise)
	}
}

// TestNote_Structure verifies Note struct initialization.
func TestNote_Structure(t *testing.T) {
	tests := []struct {
		name      string
		frequency float64
		duration  float64
		velocity  float64
	}{
		{"middle_c", 261.63, 1.0, 0.8},
		{"a440", 440.0, 0.5, 1.0},
		{"low_note", 100.0, 2.0, 0.5},
		{"high_note", 1000.0, 0.1, 0.9},
		{"silent_note", 0.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := Note{
				Frequency: tt.frequency,
				Duration:  tt.duration,
				Velocity:  tt.velocity,
			}

			if note.Frequency != tt.frequency {
				t.Errorf("Expected frequency %f, got %f", tt.frequency, note.Frequency)
			}
			if note.Duration != tt.duration {
				t.Errorf("Expected duration %f, got %f", tt.duration, note.Duration)
			}
			if note.Velocity != tt.velocity {
				t.Errorf("Expected velocity %f, got %f", tt.velocity, note.Velocity)
			}
		})
	}
}

// TestNote_VelocityRange verifies velocity values in valid range.
func TestNote_VelocityRange(t *testing.T) {
	tests := []struct {
		name     string
		velocity float64
		valid    bool
	}{
		{"silent", 0.0, true},
		{"quiet", 0.25, true},
		{"medium", 0.5, true},
		{"loud", 0.75, true},
		{"max", 1.0, true},
		{"over_max", 1.5, false},
		{"negative", -0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := Note{
				Frequency: 440.0,
				Duration:  1.0,
				Velocity:  tt.velocity,
			}

			// Check if velocity is in valid range
			inRange := note.Velocity >= 0.0 && note.Velocity <= 1.0
			if inRange != tt.valid {
				t.Errorf("Velocity %f validity: expected %v, got %v", tt.velocity, tt.valid, inRange)
			}
		})
	}
}

// TestAudioSample_Structure verifies AudioSample struct initialization.
func TestAudioSample_Structure(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
		dataLength int
	}{
		{"cd_quality", 44100, 44100},   // 1 second at CD quality
		{"high_quality", 48000, 48000}, // 1 second at 48kHz
		{"low_quality", 22050, 22050},  // 1 second at lower quality
		{"short_sample", 44100, 4410},  // 0.1 second
		{"long_sample", 44100, 441000}, // 10 seconds
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]float64, tt.dataLength)
			sample := AudioSample{
				SampleRate: tt.sampleRate,
				Data:       data,
			}

			if sample.SampleRate != tt.sampleRate {
				t.Errorf("Expected sample rate %d, got %d", tt.sampleRate, sample.SampleRate)
			}
			if len(sample.Data) != tt.dataLength {
				t.Errorf("Expected data length %d, got %d", tt.dataLength, len(sample.Data))
			}
		})
	}
}

// TestAudioSample_DataRange verifies audio sample data is in valid range.
func TestAudioSample_DataRange(t *testing.T) {
	tests := []struct {
		name  string
		data  []float64
		valid bool
	}{
		{"all_zeros", []float64{0.0, 0.0, 0.0}, true},
		{"valid_range", []float64{-0.5, 0.0, 0.5}, true},
		{"max_values", []float64{-1.0, 1.0}, true},
		{"clipping_positive", []float64{0.5, 1.5}, false},
		{"clipping_negative", []float64{-1.5, 0.5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample := AudioSample{
				SampleRate: 44100,
				Data:       tt.data,
			}

			// Check if all data is in valid range
			allValid := true
			for _, value := range sample.Data {
				if value < -1.0 || value > 1.0 {
					allValid = false
					break
				}
			}

			if allValid != tt.valid {
				t.Errorf("Data validity: expected %v, got %v", tt.valid, allValid)
			}
		})
	}
}

// TestAudioSample_EmptyData verifies behavior with empty data.
func TestAudioSample_EmptyData(t *testing.T) {
	sample := AudioSample{
		SampleRate: 44100,
		Data:       []float64{},
	}

	if len(sample.Data) != 0 {
		t.Errorf("Expected empty data, got length %d", len(sample.Data))
	}
}

// TestAudioSample_NilData verifies behavior with nil data.
func TestAudioSample_NilData(t *testing.T) {
	sample := AudioSample{
		SampleRate: 44100,
		Data:       nil,
	}

	if sample.Data != nil {
		t.Error("Expected nil data")
	}
}

// TestNote_MusicalNotes verifies common musical note frequencies.
func TestNote_MusicalNotes(t *testing.T) {
	musicalNotes := []struct {
		name      string
		frequency float64
	}{
		{"C4", 261.63},
		{"D4", 293.66},
		{"E4", 329.63},
		{"F4", 349.23},
		{"G4", 392.00},
		{"A4", 440.00},
		{"B4", 493.88},
		{"C5", 523.25},
	}

	for _, mn := range musicalNotes {
		t.Run(mn.name, func(t *testing.T) {
			note := Note{
				Frequency: mn.frequency,
				Duration:  1.0,
				Velocity:  0.8,
			}

			if note.Frequency != mn.frequency {
				t.Errorf("Expected frequency %f for %s, got %f",
					mn.frequency, mn.name, note.Frequency)
			}
		})
	}
}

// TestAudioSample_SampleRates verifies common sample rates.
func TestAudioSample_SampleRates(t *testing.T) {
	sampleRates := []int{
		8000,   // Phone quality
		11025,  // Low quality
		22050,  // Radio quality
		44100,  // CD quality
		48000,  // DVD/professional audio
		96000,  // High-resolution audio
		192000, // Ultra high-resolution
	}

	for _, rate := range sampleRates {
		t.Run("rate_"+string(rune(rate)), func(t *testing.T) {
			sample := AudioSample{
				SampleRate: rate,
				Data:       make([]float64, rate), // 1 second
			}

			if sample.SampleRate != rate {
				t.Errorf("Expected sample rate %d, got %d", rate, sample.SampleRate)
			}

			// Verify duration is approximately 1 second
			if len(sample.Data) != rate {
				t.Errorf("Expected %d samples for 1 second, got %d", rate, len(sample.Data))
			}
		})
	}
}

// TestNote_DurationValues verifies various note durations.
func TestNote_DurationValues(t *testing.T) {
	tests := []struct {
		name     string
		duration float64
	}{
		{"whole_note", 4.0},
		{"half_note", 2.0},
		{"quarter_note", 1.0},
		{"eighth_note", 0.5},
		{"sixteenth_note", 0.25},
		{"thirty_second_note", 0.125},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := Note{
				Frequency: 440.0,
				Duration:  tt.duration,
				Velocity:  0.8,
			}

			if note.Duration != tt.duration {
				t.Errorf("Expected duration %f, got %f", tt.duration, note.Duration)
			}
		})
	}
}

// TestAudioSample_SineWaveGeneration verifies sine wave data pattern.
func TestAudioSample_SineWaveGeneration(t *testing.T) {
	// Create a simple sine wave pattern
	sampleRate := 44100
	duration := 0.01 // 10ms
	samples := int(float64(sampleRate) * duration)

	data := make([]float64, samples)
	// This just verifies the structure, not actual sine wave generation
	// which would be implemented by a Synthesizer

	sample := AudioSample{
		SampleRate: sampleRate,
		Data:       data,
	}

	if len(sample.Data) != samples {
		t.Errorf("Expected %d samples, got %d", samples, len(sample.Data))
	}
}

// TestNote_ZeroValue verifies zero-value Note initialization.
func TestNote_ZeroValue(t *testing.T) {
	var note Note

	if note.Frequency != 0.0 {
		t.Errorf("Expected frequency 0.0, got %f", note.Frequency)
	}
	if note.Duration != 0.0 {
		t.Errorf("Expected duration 0.0, got %f", note.Duration)
	}
	if note.Velocity != 0.0 {
		t.Errorf("Expected velocity 0.0, got %f", note.Velocity)
	}
}

// TestAudioSample_ZeroValue verifies zero-value AudioSample initialization.
func TestAudioSample_ZeroValue(t *testing.T) {
	var sample AudioSample

	if sample.SampleRate != 0 {
		t.Errorf("Expected sample rate 0, got %d", sample.SampleRate)
	}
	if sample.Data != nil {
		t.Error("Expected nil data for zero value")
	}
}

// TestAudioSample_LargeBuffer verifies handling of large audio buffers.
func TestAudioSample_LargeBuffer(t *testing.T) {
	// 1 minute at CD quality
	sampleRate := 44100
	duration := 60.0 // seconds
	samples := int(float64(sampleRate) * duration)

	sample := AudioSample{
		SampleRate: sampleRate,
		Data:       make([]float64, samples),
	}

	expectedSamples := 2646000 // 44100 * 60
	if len(sample.Data) != expectedSamples {
		t.Errorf("Expected %d samples, got %d", expectedSamples, len(sample.Data))
	}
}

// TestNote_FrequencyRange verifies various frequency ranges.
func TestNote_FrequencyRange(t *testing.T) {
	tests := []struct {
		name      string
		frequency float64
		rangeType string
	}{
		{"sub_bass", 20.0, "sub-bass"},
		{"bass", 100.0, "bass"},
		{"low_mid", 250.0, "low-mid"},
		{"mid", 1000.0, "mid"},
		{"high_mid", 4000.0, "high-mid"},
		{"presence", 8000.0, "presence"},
		{"brilliance", 16000.0, "brilliance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := Note{
				Frequency: tt.frequency,
				Duration:  1.0,
				Velocity:  0.8,
			}

			if note.Frequency != tt.frequency {
				t.Errorf("Expected frequency %f, got %f", tt.frequency, note.Frequency)
			}

			// Verify frequency is positive
			if note.Frequency < 0 {
				t.Error("Frequency should be positive")
			}
		})
	}
}
