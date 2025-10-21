// Package synthesis provides low-level audio waveform generation.
// It implements oscillators for basic waveforms (sine, square, sawtooth, triangle, noise)
// with ADSR envelopes for shaping sound over time.
//
// All waveform generation is deterministic when using seeded random number generators,
// ensuring consistent audio generation across network sessions.
package synthesis
