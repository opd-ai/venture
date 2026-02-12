# Audit: github.com/opd-ai/venture/pkg/audio/synthesis
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The audio synthesis package provides deterministic waveform generation with ADSR envelopes and is well-tested (97.8% coverage). Core functionality is solid with proper seed-based randomness and structured logging. However, it has interface compliance issues (doesn't implement audio.Synthesizer), concurrency bugs in envelope application methods, excessive debug logging, and documentation inconsistencies across multiple files.

## Issues Found
- [ ] **high** Interface compliance — Engine doesn't implement audio.Synthesizer interface; method naming mismatch: `GenerateTone()` vs `Generate()` (`engine.go:59`)
- [ ] **high** Concurrency — GenerateChordWithEnvelope missing mutex lock when applying envelope to returned sample data (`engine.go:133-136`)
- [ ] **high** Concurrency — ApplyEnvelope has no mutex protection while modifying sample data that could be accessed concurrently (`engine.go:176-178`)
- [ ] **med** Documentation — Multiple package-level doc comments across files (doc.go, engine.go, oscillator.go, envelope.go) create inconsistent/redundant documentation (`doc.go:1`, `engine.go:1`, `oscillator.go:1`, `envelope.go:1`)
- [ ] **low** Performance — Excessive debug logging in hot path waveform generation (27 log statements for simple operations) (`oscillator.go:55-262`)
- [ ] **low** Documentation — waveformName() is unexported but could be useful for debugging/logging in other packages (`oscillator.go:265`)

## Test Coverage
97.8% (target: 65%) ✅

## Integration Status
**Direct integrations:**
- `pkg/audio/music/generator.go` — Uses synthesis.Oscillator directly (bypasses Engine)
- `pkg/audio/music/adaptive.go` — Uses synthesis.Oscillator directly (bypasses Engine)  
- `pkg/audio/sfx/effects.go` — Uses synthesis.Envelope for shaping sound effects

**Missing integration:**
- Engine should implement audio.Synthesizer interface but doesn't due to method name mismatch
- Music/SFX packages use Oscillator directly rather than Engine (suggests Engine API not ergonomic)

**No ECS registration needed:** This is a pure utility package, not a system/component.

## Recommendations
1. **[HIGH] Fix interface compliance**: Rename `GenerateTone()` to `Generate()` to implement audio.Synthesizer interface, or add wrapper method
2. **[HIGH] Fix concurrency bugs**: Add mutex lock to GenerateChordWithEnvelope and ApplyEnvelope, or document that callers must not share AudioSample between goroutines
3. **[MED] Consolidate documentation**: Keep only doc.go package comment; remove duplicate package docs from engine.go, oscillator.go, envelope.go
4. **[LOW] Reduce logging verbosity**: Move debug logging to trace level or remove from hot paths (waveform generation loops)
5. **[LOW] Export waveformName()**: Make it WaveformName() for reuse in music/sfx packages
