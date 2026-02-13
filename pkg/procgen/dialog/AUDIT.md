# Audit: github.com/opd-ai/venture/pkg/procgen/dialog
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The dialog package provides runtime NPC dialog generation using Markov chains with genre-specific corpora and personality traits. Overall package health is strong with 88.7% test coverage and solid deterministic procgen patterns. Three medium-severity issues identified: unchecked error in hash seeding, genre ID mismatch between getter and corpus retrieval, and missing structured logging for debugging. No blocking issues found.

## Issues Found
- [x] **severity:med** Error handling — `binary.Write()` error ignored in `deriveRuntimeSeed()` (`markov.go:289`) — **FIXED**: Added error check with fallback seed derivation using hash64()
- [x] **severity:med** API consistency — `GetAvailableGenres()` returns `"postapoc"` but `GetCorpus()` expects `"postapocalyptic"`, causing nil corpus retrieval (`corpus.go:694` vs `corpus.go:30`) — **FIXED**: Updated GetAvailableGenres() to return "postapocalyptic" matching GetCorpus() switch cases
- [ ] **severity:low** Logging — No structured logging with `logrus.WithFields` for generation failures or fallbacks (all files)
- [ ] **severity:low** Documentation — `GenerateParams.Temperature` field not documented in procgen.Generator integration example in `doc.go`
- [ ] **severity:low** Test coverage — Missing benchmark for `deriveRuntimeSeed()` hash performance (should verify <1ms target)

## Test Coverage
88.7% (target: 65%) ✅

**Coverage breakdown:**
- `corpus.go`: 100% (simple data structures)
- `markov.go`: ~85% (core generation logic well-tested)
- `personality.go`: ~90% (personality traits and greetings)
- `utils.go`: ~85% (utility functions)

## Integration Status
**Integration points:**
1. ✅ Implements `procgen.Generator` interface (`markov.go:344-435`)
   - `Generate(seed, params)` method converts procgen params to dialog params
   - `Validate(result)` checks dialog text validity
2. ✅ Used by `pkg/engine/markov_dialog_provider.go` for NPC dialog
3. ✅ Used by `pkg/engine/dialog_ui.go` for UI rendering
4. ✅ Used by `pkg/engine/companion_spawning.go` for companion dialog
5. ✅ Used by `pkg/engine/book_spawning.go` for book content generation

**No registration required** — Package is pure procgen library, not an ECS system.

**Deterministic generation:** ✅ All randomness uses seeded `*rand.Rand` instances, never global `rand` functions.

**No serialization needed** — Dialog is generated on-demand, not persisted as component state.

## Recommendations
1. **Fix error handling** — Check `binary.Write()` error in `markov.go:289`:
   ```go
   if err := binary.Write(h, binary.LittleEndian, m.seed); err != nil {
       // Fallback to simpler seed derivation or log error
       return m.seed ^ hash64(playerInput) ^ hash64(conversationID)
   }
   ```
2. **Fix genre ID mismatch** — Update `GetAvailableGenres()` to return `"postapocalyptic"` instead of `"postapoc"` for consistency with `GetCorpus()` switch cases (`corpus.go:694`).
3. **Add structured logging** — Use `logrus.WithFields` for generation failures:
   ```go
   if result == "" {
       logger.WithFields(logrus.Fields{
           "genreID": m.genreID,
           "chainSize": len(m.chain),
           "params": params,
       }).Warn("dialog generation returned empty result")
   }
   ```
4. **Add benchmark** — Verify `deriveRuntimeSeed()` performance meets <1ms target for real-time conversation.
5. **Document temperature** — Add `Temperature` field documentation to `doc.go` procgen.Generator example.
