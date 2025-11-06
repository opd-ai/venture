// Package music provides procedural music composition with motif system.
// This file implements leitmotif generation for characters, factions, and locations.
package music

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/audio"
)

// Motif represents a musical leitmotif (recurring theme).
// Motifs are short melodic patterns that represent characters, factions, or locations.
type Motif struct {
	// ID uniquely identifies this motif
	ID string
	// Name is the human-readable motif name
	Name string
	// Type categorizes the motif (character, faction, location)
	Type MotifType
	// Scale is the musical scale for this motif
	Scale Scale
	// Notes are the melodic pattern (MIDI note offsets from scale root)
	Notes []int
	// Rhythm defines the note durations and velocities
	Rhythm Rhythm
	// Tempo is the base tempo in BPM
	Tempo float64
	// Waveform is the primary sound wave type
	Waveform audio.WaveformType
}

// MotifType categorizes musical motifs
type MotifType int

const (
	// MotifTypeCharacter represents character themes
	MotifTypeCharacter MotifType = iota
	// MotifTypeFaction represents faction themes
	MotifTypeFaction
	// MotifTypeLocation represents location themes
	MotifTypeLocation
)

// String returns the string representation of the motif type
func (mt MotifType) String() string {
	switch mt {
	case MotifTypeCharacter:
		return "character"
	case MotifTypeFaction:
		return "faction"
	case MotifTypeLocation:
		return "location"
	default:
		return "unknown"
	}
}

// MotifGenerator creates procedural musical motifs.
type MotifGenerator struct {
	sampleRate int
	seed       int64
	rng        *rand.Rand
}

// NewMotifGenerator creates a new motif generator.
func NewMotifGenerator(sampleRate int, seed int64) *MotifGenerator {
	return &MotifGenerator{
		sampleRate: sampleRate,
		seed:       seed,
		rng:        rand.New(rand.NewSource(seed)),
	}
}

// GenerateMotif creates a musical motif for the given entity and genre.
// The motif is deterministic based on the entity ID and genre.
func (mg *MotifGenerator) GenerateMotif(entityID, genre string, motifType MotifType) *Motif {
	// Create local RNG for deterministic generation
	localSeed := hashString(entityID) + mg.seed + int64(motifType)*1000
	localRng := rand.New(rand.NewSource(localSeed))

	// Select scale based on genre
	scale := GetScaleForGenre(genre)

	// Generate melodic pattern (4-8 notes)
	noteCount := 4 + localRng.Intn(5) // 4-8 notes
	notes := make([]int, noteCount)

	// Generate interval pattern using scale degrees
	for i := 0; i < noteCount; i++ {
		// Use scale degrees (0-6) mapped to scale intervals
		scaleDegree := localRng.Intn(len(scale.Intervals))
		octave := localRng.Intn(2) // 0-1 octave above root
		notes[i] = scale.Intervals[scaleDegree] + octave*12
	}

	// Ensure motif has a strong identity - repeat first note at end
	if noteCount > 2 {
		notes[noteCount-1] = notes[0]
	}

	// Generate rhythm pattern
	rhythm := mg.generateMotifRhythm(noteCount, motifType, localRng)

	// Select tempo based on motif type
	tempo := mg.getTempoForMotifType(motifType, localRng)

	// Select waveform based on genre and type
	waveform := mg.getWaveformForMotif(genre, motifType, localRng)

	return &Motif{
		ID:       entityID,
		Name:     entityID, // Can be overridden
		Type:     motifType,
		Scale:    scale,
		Notes:    notes,
		Rhythm:   rhythm,
		Tempo:    tempo,
		Waveform: waveform,
	}
}

// generateMotifRhythm creates a rhythm pattern for the motif.
func (mg *MotifGenerator) generateMotifRhythm(noteCount int, motifType MotifType, rng *rand.Rand) Rhythm {
	pattern := make([]float64, noteCount)
	velocity := make([]float64, noteCount)

	// Base duration (quarter notes = 1.0, eighth notes = 0.5)
	baseDuration := 1.0
	if motifType == MotifTypeCharacter {
		// Characters get more varied rhythms
		baseDuration = 0.5 + rng.Float64()*0.5 // 0.5-1.0
	}

	for i := 0; i < noteCount; i++ {
		// Add slight variation to duration
		pattern[i] = baseDuration * (0.8 + rng.Float64()*0.4) // ±20% variation

		// Emphasize first and last notes
		if i == 0 || i == noteCount-1 {
			velocity[i] = 0.8 + rng.Float64()*0.2 // 0.8-1.0
		} else {
			velocity[i] = 0.6 + rng.Float64()*0.3 // 0.6-0.9
		}
	}

	return Rhythm{
		Pattern:  pattern,
		Velocity: velocity,
	}
}

// getTempoForMotifType returns appropriate tempo for the motif type.
func (mg *MotifGenerator) getTempoForMotifType(motifType MotifType, rng *rand.Rand) float64 {
	baseTempo := 120.0

	switch motifType {
	case MotifTypeCharacter:
		// Characters: moderate tempo with variation
		return baseTempo + (rng.Float64()-0.5)*40.0 // 100-140 BPM
	case MotifTypeFaction:
		// Factions: slightly slower, more authoritative
		return 100.0 + rng.Float64()*20.0 // 100-120 BPM
	case MotifTypeLocation:
		// Locations: slower, more atmospheric
		return 80.0 + rng.Float64()*30.0 // 80-110 BPM
	default:
		return baseTempo
	}
}

// getWaveformForMotif selects an appropriate waveform for the motif.
func (mg *MotifGenerator) getWaveformForMotif(genre string, motifType MotifType, rng *rand.Rand) audio.WaveformType {
	switch genre {
	case "fantasy":
		// Fantasy uses organic waveforms
		waveforms := []audio.WaveformType{audio.WaveformSine, audio.WaveformTriangle}
		return waveforms[rng.Intn(len(waveforms))]
	case "scifi":
		// Sci-fi uses synthetic waveforms
		waveforms := []audio.WaveformType{audio.WaveformSquare, audio.WaveformSawtooth}
		return waveforms[rng.Intn(len(waveforms))]
	case "horror":
		// Horror uses darker, more dissonant waveforms
		waveforms := []audio.WaveformType{audio.WaveformSawtooth, audio.WaveformSquare}
		return waveforms[rng.Intn(len(waveforms))]
	case "cyberpunk":
		// Cyberpunk uses square waves
		return audio.WaveformSquare
	case "post-apocalyptic":
		// Post-apocalyptic uses sawtooth for harshness
		return audio.WaveformSawtooth
	default:
		return audio.WaveformSine
	}
}

// hashString creates a deterministic hash from a string.
func hashString(s string) int64 {
	var hash int64
	for i := 0; i < len(s); i++ {
		hash = hash*31 + int64(s[i])
	}
	return hash
}
