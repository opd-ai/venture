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

// DefaultEnvelope returns a standard ADSR envelope.
func DefaultEnvelope() Envelope {
	return Envelope{
		Attack:  0.01,
		Decay:   0.1,
		Sustain: 0.7,
		Release: 0.2,
	}
}

// Apply applies the ADSR envelope to an audio sample.
func (e *Envelope) Apply(data []float64, sampleRate int) {
	numSamples := len(data)
	if numSamples == 0 {
		return
	}
	
	attackSamples := int(e.Attack * float64(sampleRate))
	decaySamples := int(e.Decay * float64(sampleRate))
	releaseSamples := int(e.Release * float64(sampleRate))
	
	// Ensure we don't exceed the sample length
	if attackSamples > numSamples {
		attackSamples = numSamples
	}
	if attackSamples+decaySamples > numSamples {
		decaySamples = numSamples - attackSamples
	}
	if releaseSamples > numSamples {
		releaseSamples = numSamples
	}
	
	sustainSamples := numSamples - attackSamples - decaySamples - releaseSamples
	if sustainSamples < 0 {
		sustainSamples = 0
	}
	
	idx := 0
	
	// Attack phase: ramp from 0 to 1
	for i := 0; i < attackSamples && idx < len(data); i++ {
		envelope := float64(i) / float64(attackSamples)
		data[idx] *= envelope
		idx++
	}
	
	// Decay phase: ramp from 1 to sustain level
	for i := 0; i < decaySamples && idx < len(data); i++ {
		envelope := 1.0 - (1.0-e.Sustain)*(float64(i)/float64(decaySamples))
		data[idx] *= envelope
		idx++
	}
	
	// Sustain phase: constant at sustain level
	for i := 0; i < sustainSamples && idx < len(data); i++ {
		data[idx] *= e.Sustain
		idx++
	}
	
	// Release phase: ramp from sustain to 0
	for i := 0; i < releaseSamples && idx < len(data); i++ {
		envelope := e.Sustain * (1.0 - float64(i)/float64(releaseSamples))
		data[idx] *= envelope
		idx++
	}
}
