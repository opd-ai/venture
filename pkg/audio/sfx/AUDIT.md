# Audit: pkg/audio/sfx
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The `pkg/audio/sfx` package provides procedural sound effect generation with genre-specific variations, pitch/volume modulation, and caching via VarietyManager. The implementation is largely complete with 87.5% test coverage, excellent deterministic design, and proper integration with the engine's AudioManager. However, critical gaps include missing tests for GenerateWithGenre (the main public API), no test coverage for genre-specific modifications (scifi, horror, cyberpunk, postapoc), and helper functions (pitchRatioFromSemitones, pow2) that lack exported documentation despite being important for audio processing accuracy.

## Issues Found
- [ ] high test-coverage — GenerateWithGenre method has no tests despite being the primary public API used by AudioManager (`generator.go:54`)
- [ ] high test-coverage — Genre-specific modifications (applyScifiModifications, applyHorrorModifications, applyCyberpunkModifications, applyPostApocalypticModifications) have zero test coverage (`generator.go:122-148`)
- [ ] med doc-coverage — pitchRatioFromSemitones lacks godoc comment despite being a critical audio calculation utility (`helpers.go:8`)
- [ ] med doc-coverage — pow2 lacks godoc comment despite being a critical mathematical approximation function with Taylor series expansion (`helpers.go:23`)
- [ ] med test-coverage — applyGenreModifications has no direct tests; only tested indirectly through untested GenerateWithGenre (`generator.go:108`)
- [ ] low doc-coverage — EffectType constants lack individual godoc comments explaining use cases for each effect (EffectImpact through EffectPowerup) (`types.go:10-20`)

## Test Coverage
87.5% (target: 65%) ✅

**Note**: While overall coverage exceeds the target, the missing tests for GenerateWithGenre represent a significant quality gap since this is the primary API method used by the engine (see `pkg/engine/audio_manager.go:149`).

## Integration Status
**Fully Integrated** ✅

The package is properly integrated into the game engine:
- Imported and used by `pkg/engine/audio_manager.go` (line 8)
- AudioManager creates sfxGen instance with deterministic seed (line 42)
- AudioManager calls `sfxGen.GenerateWithGenre(effectType, effectSeed, genre)` at line 149
- Imported by `cmd/client/handlers.go` (line 74) for client-side audio handling
- No system registration needed (SFX is consumed as a library, not an ECS system)

The package correctly follows the deterministic generation pattern:
- Uses `rand.New(rand.NewSource(seed))` throughout
- No global `rand`, `time.Now()`, or OS entropy usage ✅
- Each generation method takes a seed parameter for reproducibility

## Recommendations
1. **HIGH PRIORITY**: Add table-driven tests for `GenerateWithGenre` covering all genres (fantasy, scifi, horror, cyberpunk, postapoc) and all effect types. This is the primary API method and must have comprehensive test coverage.
2. **HIGH PRIORITY**: Add tests for genre-specific modification functions verifying that:
   - scifi increases pitch (~1.3x) and reduces amplitude
   - horror decreases pitch (~0.7x) and applies vibrato
   - cyberpunk increases pitch (~1.4x) and applies hard clipping
   - postapoc applies soft clipping and moderate pitch reduction
3. **MEDIUM PRIORITY**: Add godoc comments to `pitchRatioFromSemitones` and `pow2` explaining the audio mathematics (semitones to frequency ratio conversion, Taylor series approximation for 2^x).
4. **MEDIUM PRIORITY**: Add godoc comments to each EffectType constant explaining intended use case and typical characteristics (e.g., "EffectImpact - Short, punchy sound for collisions and melee hits").
5. **LOW PRIORITY**: Consider adding benchmark tests for `GenerateWithGenre` to ensure genre modifications don't significantly impact performance compared to baseline `Generate`.
