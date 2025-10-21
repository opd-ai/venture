package music

import (
	"math"
)

// Scale represents a musical scale.
type Scale struct {
	Name      string
	Intervals []int // semitones from root
}

// Common musical scales
var (
	ScaleMajor = Scale{
		Name:      "Major",
		Intervals: []int{0, 2, 4, 5, 7, 9, 11},
	}
	ScaleMinor = Scale{
		Name:      "Minor",
		Intervals: []int{0, 2, 3, 5, 7, 8, 10},
	}
	ScalePentatonic = Scale{
		Name:      "Pentatonic",
		Intervals: []int{0, 2, 4, 7, 9},
	}
	ScaleBlues = Scale{
		Name:      "Blues",
		Intervals: []int{0, 3, 5, 6, 7, 10},
	}
	ScaleChromatic = Scale{
		Name:      "Chromatic",
		Intervals: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
	}
)

// NoteToFrequency converts a MIDI note number to frequency in Hz.
// A4 (note 69) = 440 Hz
func NoteToFrequency(note int) float64 {
	return 440.0 * math.Pow(2.0, float64(note-69)/12.0)
}

// GetScaleForGenre returns an appropriate scale for the given genre.
func GetScaleForGenre(genre string) Scale {
	switch genre {
	case "fantasy":
		return ScaleMajor
	case "scifi":
		return ScaleChromatic
	case "horror":
		return ScaleMinor
	case "cyberpunk":
		return ScaleBlues
	case "post-apocalyptic":
		return ScalePentatonic
	default:
		return ScaleMajor
	}
}

// Chord represents a musical chord.
type Chord struct {
	Root  int   // MIDI note number
	Notes []int // semitone offsets from root
}

// Common chord types
var (
	ChordMajor      = []int{0, 4, 7}
	ChordMinor      = []int{0, 3, 7}
	ChordDiminished = []int{0, 3, 6}
	ChordAugmented  = []int{0, 4, 8}
	ChordSeventh    = []int{0, 4, 7, 10}
)

// GetChordProgression returns a chord progression for a genre.
func GetChordProgression(genre string, rootNote int) []Chord {
	switch genre {
	case "fantasy":
		return []Chord{
			{Root: rootNote, Notes: ChordMajor},
			{Root: rootNote + 5, Notes: ChordMajor},
			{Root: rootNote + 7, Notes: ChordMinor},
			{Root: rootNote, Notes: ChordMajor},
		}
	case "horror":
		return []Chord{
			{Root: rootNote, Notes: ChordMinor},
			{Root: rootNote - 2, Notes: ChordDiminished},
			{Root: rootNote + 3, Notes: ChordMinor},
			{Root: rootNote - 5, Notes: ChordMinor},
		}
	case "scifi":
		return []Chord{
			{Root: rootNote, Notes: ChordMajor},
			{Root: rootNote + 7, Notes: ChordSeventh},
			{Root: rootNote + 3, Notes: ChordMinor},
			{Root: rootNote + 10, Notes: ChordMajor},
		}
	default:
		return []Chord{
			{Root: rootNote, Notes: ChordMajor},
			{Root: rootNote + 7, Notes: ChordMajor},
			{Root: rootNote, Notes: ChordMajor},
			{Root: rootNote + 5, Notes: ChordMajor},
		}
	}
}

// Rhythm represents a rhythmic pattern.
type Rhythm struct {
	Pattern  []float64 // note durations in beats
	Velocity []float64 // velocity for each note
}

// GetRhythmForContext returns a rhythm pattern for the given context.
func GetRhythmForContext(context string) Rhythm {
	switch context {
	case "combat":
		return Rhythm{
			Pattern:  []float64{0.25, 0.25, 0.25, 0.25},
			Velocity: []float64{0.8, 0.6, 0.7, 0.6},
		}
	case "exploration":
		return Rhythm{
			Pattern:  []float64{0.5, 0.5, 1.0},
			Velocity: []float64{0.5, 0.5, 0.6},
		}
	case "ambient":
		return Rhythm{
			Pattern:  []float64{2.0, 2.0},
			Velocity: []float64{0.3, 0.3},
		}
	case "victory":
		return Rhythm{
			Pattern:  []float64{0.25, 0.25, 0.5, 1.0},
			Velocity: []float64{0.7, 0.8, 0.9, 1.0},
		}
	default:
		return Rhythm{
			Pattern:  []float64{1.0, 1.0},
			Velocity: []float64{0.5, 0.5},
		}
	}
}

// GetTempoForContext returns BPM for the given context.
func GetTempoForContext(context string) float64 {
	switch context {
	case "combat":
		return 140.0
	case "exploration":
		return 90.0
	case "ambient":
		return 60.0
	case "victory":
		return 120.0
	default:
		return 100.0
	}
}
