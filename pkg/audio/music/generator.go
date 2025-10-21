package music

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
	"github.com/opd-ai/venture/pkg/audio/synthesis"
)

// Generator creates procedural music tracks.
type Generator struct {
	sampleRate int
	osc        *synthesis.Oscillator
	rng        *rand.Rand
}

// NewGenerator creates a new music generator.
func NewGenerator(sampleRate int, seed int64) *Generator {
	return &Generator{
		sampleRate: sampleRate,
		osc:        synthesis.NewOscillator(sampleRate, seed),
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// GenerateTrack creates a music track for the given context.
func (g *Generator) GenerateTrack(genre string, context string, seed int64, duration float64) *audio.AudioSample {
	localRng := rand.New(rand.NewSource(seed))
	
	// Get musical parameters based on genre and context
	scale := GetScaleForGenre(genre)
	tempo := GetTempoForContext(context)
	rhythm := GetRhythmForContext(context)
	
	// Calculate root note (between C3 and C4)
	rootNote := 48 + localRng.Intn(12)
	
	// Get chord progression
	chords := GetChordProgression(genre, rootNote)
	
	// Generate the track
	numSamples := int(float64(g.sampleRate) * duration)
	track := make([]float64, numSamples)
	
	// Generate melody and harmony
	g.generateMelody(track, scale, rootNote, rhythm, tempo, localRng)
	g.generateHarmony(track, chords, rhythm, tempo, localRng)
	
	// Apply master envelope for fade in/out
	g.applyMasterEnvelope(track, duration)
	
	return &audio.AudioSample{
		SampleRate: g.sampleRate,
		Data:       track,
	}
}

// generateMelody generates the melodic line.
func (g *Generator) generateMelody(track []float64, scale Scale, rootNote int, rhythm Rhythm, tempo float64, rng *rand.Rand) {
	beatDuration := 60.0 / tempo // seconds per beat
	samplePos := 0
	
	// Choose waveform based on preference
	waveforms := []audio.WaveformType{
		audio.WaveformSine,
		audio.WaveformTriangle,
		audio.WaveformSquare,
	}
	waveform := waveforms[rng.Intn(len(waveforms))]
	
	// Generate notes following the rhythm
	for samplePos < len(track) {
		for i := range rhythm.Pattern {
			noteDuration := rhythm.Pattern[i] * beatDuration
			velocity := rhythm.Velocity[i]
			
			// Choose a note from the scale
			octave := 1 + rng.Intn(2) // 1 or 2 octaves above root
			scaleIndex := rng.Intn(len(scale.Intervals))
			noteOffset := scale.Intervals[scaleIndex] + octave*12
			note := rootNote + noteOffset
			
			freq := NoteToFrequency(note)
			
			// Generate the note
			noteSample := g.osc.Generate(waveform, freq, noteDuration)
			
			// Apply envelope
			env := synthesis.Envelope{
				Attack:  0.01,
				Decay:   0.1,
				Sustain: 0.6,
				Release: 0.2,
			}
			env.Apply(noteSample.Data, noteSample.SampleRate)
			
			// Mix into track with velocity
			for j := 0; j < len(noteSample.Data) && samplePos+j < len(track); j++ {
				track[samplePos+j] += noteSample.Data[j] * velocity * 0.3
			}
			
			samplePos += len(noteSample.Data)
			if samplePos >= len(track) {
				return
			}
		}
	}
}

// generateHarmony generates harmonic chords.
func (g *Generator) generateHarmony(track []float64, chords []Chord, rhythm Rhythm, tempo float64, rng *rand.Rand) {
	beatDuration := 60.0 / tempo
	samplePos := 0
	chordIndex := 0
	
	// Use sine wave for harmony (smoother)
	for samplePos < len(track) {
		chord := chords[chordIndex%len(chords)]
		chordIndex++
		
		// Chord duration is sum of rhythm pattern
		var chordDuration float64
		for _, beatLen := range rhythm.Pattern {
			chordDuration += beatLen * beatDuration
		}
		
		// Generate each note in the chord
		numSamples := int(chordDuration * float64(g.sampleRate))
		if samplePos+numSamples > len(track) {
			numSamples = len(track) - samplePos
		}
		
		for _, offset := range chord.Notes {
			note := chord.Root + offset
			freq := NoteToFrequency(note)
			
			noteSample := g.osc.Generate(audio.WaveformSine, freq, chordDuration)
			
			env := synthesis.Envelope{
				Attack:  0.05,
				Decay:   0.1,
				Sustain: 0.7,
				Release: 0.3,
			}
			env.Apply(noteSample.Data, noteSample.SampleRate)
			
			// Mix into track
			for j := 0; j < numSamples && samplePos+j < len(track); j++ {
				if j < len(noteSample.Data) {
					track[samplePos+j] += noteSample.Data[j] * 0.2 / float64(len(chord.Notes))
				}
			}
		}
		
		samplePos += numSamples
	}
}

// applyMasterEnvelope applies fade in/out to the entire track.
func (g *Generator) applyMasterEnvelope(track []float64, duration float64) {
	fadeDuration := 0.5 // seconds
	fadeSamples := int(fadeDuration * float64(g.sampleRate))
	
	if fadeSamples > len(track)/4 {
		fadeSamples = len(track) / 4
	}
	
	// Fade in
	for i := 0; i < fadeSamples && i < len(track); i++ {
		fade := float64(i) / float64(fadeSamples)
		track[i] *= fade
	}
	
	// Fade out
	for i := 0; i < fadeSamples && i < len(track); i++ {
		idx := len(track) - 1 - i
		fade := float64(i) / float64(fadeSamples)
		track[idx] *= fade
	}
}
