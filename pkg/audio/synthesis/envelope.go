package synthesis

// Envelope represents an ADSR (Attack, Decay, Sustain, Release) envelope.
type Envelope struct {
	// Attack time in seconds
	Attack float64

	// Decay time in seconds
	Decay float64

	// Sustain level (0.0 to 1.0)
	Sustain float64

	// Release time in seconds
	Release float64
}

// DefaultEnvelope returns a standard ADSR envelope suitable for short percussive
// sounds. For sustained tones or pads, customize the values: increase Attack for
// softer onset, increase Sustain (closer to 1.0) for longer held notes, and
// increase Release for gradual fadeout.
func DefaultEnvelope() Envelope {
	return Envelope{
		Attack:  0.01,
		Decay:   0.1,
		Sustain: 0.7,
		Release: 0.2,
	}
}

// Apply applies the ADSR envelope to an audio sample.
// The data slice is modified in-place. Callers should pass a copy if the original is needed.
// The envelope is applied over the entire sample duration, divided into ADSR phases.
func (e *Envelope) Apply(data []float64, sampleRate int) {
	numSamples := len(data)
	if numSamples == 0 {
		return
	}

	phases := e.calculatePhaseLengths(numSamples, sampleRate)
	e.applyAllPhases(data, phases)
}

// envelopePhases holds the calculated lengths of each ADSR phase.
type envelopePhases struct {
	attack  int
	decay   int
	sustain int
	release int
}

// calculatePhaseLengths computes the sample counts for each envelope phase.
func (e *Envelope) calculatePhaseLengths(numSamples, sampleRate int) envelopePhases {
	attack := e.clampToSampleLength(int(e.Attack*float64(sampleRate)), numSamples)
	decay := e.clampToSampleLength(int(e.Decay*float64(sampleRate)), numSamples-attack)
	release := e.clampToSampleLength(int(e.Release*float64(sampleRate)), numSamples)

	sustain := numSamples - attack - decay - release
	if sustain < 0 {
		sustain = 0
	}

	return envelopePhases{
		attack:  attack,
		decay:   decay,
		sustain: sustain,
		release: release,
	}
}

// clampToSampleLength ensures a phase length does not exceed the maximum allowed.
func (e *Envelope) clampToSampleLength(phase, max int) int {
	if phase > max {
		return max
	}
	return phase
}

// applyAllPhases applies attack, decay, sustain, and release phases to the audio data.
func (e *Envelope) applyAllPhases(data []float64, phases envelopePhases) {
	idx := 0
	idx = e.applyAttack(data, idx, phases.attack)
	idx = e.applyDecay(data, idx, phases.decay)
	idx = e.applySustain(data, idx, phases.sustain)
	e.applyRelease(data, idx, phases.release)
}

// applyAttack applies the attack phase (ramp from 0 to 1).
func (e *Envelope) applyAttack(data []float64, start, length int) int {
	idx := start
	for i := 0; i < length && idx < len(data); i++ {
		envelope := float64(i) / float64(length)
		data[idx] *= envelope
		idx++
	}
	return idx
}

// applyDecay applies the decay phase (ramp from 1 to sustain level).
func (e *Envelope) applyDecay(data []float64, start, length int) int {
	idx := start
	for i := 0; i < length && idx < len(data); i++ {
		envelope := 1.0 - (1.0-e.Sustain)*(float64(i)/float64(length))
		data[idx] *= envelope
		idx++
	}
	return idx
}

// applySustain applies the sustain phase (constant at sustain level).
func (e *Envelope) applySustain(data []float64, start, length int) int {
	idx := start
	for i := 0; i < length && idx < len(data); i++ {
		data[idx] *= e.Sustain
		idx++
	}
	return idx
}

// applyRelease applies the release phase (ramp from sustain to 0).
func (e *Envelope) applyRelease(data []float64, start, length int) int {
	idx := start
	for i := 0; i < length && idx < len(data); i++ {
		envelope := e.Sustain * (1.0 - float64(i)/float64(length))
		data[idx] *= envelope
		idx++
	}
	return idx
}
