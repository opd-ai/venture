# Audit: github.com/opd-ai/venture/pkg/audio/synthesis
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
Audio synthesis package provides deterministic waveform generation with oscillators and ADSR envelopes for procedural audio. Core functionality is solid with excellent test coverage (97.8%), proper seed-based randomness, and structured logging. Interface compliance verified at compile time. Concurrency fixes applied for envelope methods.

## Issues Found
- [x] **high** Interface compliance — Engine doesn't implement audio.Synthesizer interface; has `GenerateTone()` instead of `Generate()` method — **FIXED**: Engine now implements Synthesizer interface via `Generate()` method with compile-time verification (`var _ audio.Synthesizer = (*Engine)(nil)`)
- [x] **high** Concurrency — GenerateChordWithEnvelope calls GenerateChord (releases mutex), then calls env.Apply without mutex protection — **FIXED**: Mutex is now acquired before env.Apply() call
- [x] **high** Concurrency — ApplyEnvelope has no mutex protection while modifying sample data — **FIXED**: ApplyEnvelope now has proper mutex protection with documentation noting caller must ensure sample is not accessed by other goroutines
- [ ] **med** Performance — Excessive debug logging in hot path: 51 log.WithFields() calls in waveform generation (oscillator.go has 27 statements per simple waveform) (`oscillator.go:55-262`)
- [ ] **med** Documentation — Package-level doc comments duplicated across 4 files (doc.go, engine.go, oscillator.go, envelope.go); should have single authoritative package doc (`doc.go:1`, `engine.go:1-3`, `oscillator.go:1-3`, `envelope.go:1-3`)
- [ ] **low** Code organization — Global package-level logger with init() makes testing harder; should inject logger or use context-based logging (`oscillator.go:14-20`)
- [ ] **low** API design — waveformName() helper is unexported; could be useful for debugging/logging in music/sfx packages (`oscillator.go:265`)
- [ ] **low** Documentation — DefaultEnvelope() lacks godoc explaining when/why to use defaults vs custom values (`envelope.go:22`)

## Test Coverage
97.8% (target: 65%) ✅

## Integration Status
**Integration points:**
- `pkg/audio/music/generator.go` — Uses synthesis.Oscillator directly (bypasses Engine API)
- `pkg/audio/music/adaptive.go` — Uses synthesis.Oscillator and Envelope directly  
- `pkg/audio/sfx/effects.go` — Uses synthesis.Envelope for sound effect shaping

**Interface compliance gap:**
- Engine SHOULD implement audio.Synthesizer interface (lines 39-46 in pkg/audio/interfaces.go)
- Method name mismatch: interface expects `Generate()`, Engine provides `GenerateTone()`
- Consequence: Engine cannot be used polymorphically as Synthesizer type

**Design observations:**
- Music/SFX packages bypass Engine and use Oscillator directly → suggests Engine API not ergonomic
- All integrators create their own Envelope instances rather than using Engine methods → possible API gap

**No ECS registration needed:** Pure utility package, not a system or component.

## Recommendations
1. **[HIGH - Interface]** Add `Generate()` method to Engine that delegates to `GenerateTone()`, enabling audio.Synthesizer interface compliance. Alternative: rename GenerateTone→Generate if no breaking change concerns.
2. **[HIGH - Concurrency]** Fix GenerateChordWithEnvelope: acquire mutex before env.Apply() call, or document that returned AudioSample must not be shared across goroutines until envelope application completes.
3. **[HIGH - Concurrency]** Fix ApplyEnvelope: add mutex lock around env.Apply() call to prevent concurrent modification of sample.Data slice.
4. **[MED - Performance]** Reduce logging: move 27+ debug statements in oscillator.go to trace level, or remove from hot paths (waveform generation loops execute per-sample).
5. **[MED - Docs]** Consolidate package documentation: keep doc.go as single source of truth; remove package-level comments from engine.go, oscillator.go, envelope.go.
6. **[LOW - Testing]** Add logger injection to Oscillator constructor: replace global logger with per-instance logger passed via NewOscillator() for better test isolation.
7. **[LOW - API]** Export waveformName() as WaveformName() for reuse in music/sfx packages that log waveform types.
8. **[LOW - Docs]** Document DefaultEnvelope() with usage guidance (e.g., "suitable for short percussive sounds; customize for sustained tones").
