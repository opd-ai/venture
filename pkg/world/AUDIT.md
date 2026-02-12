# Audit: github.com/opd-ai/venture/pkg/world
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The `pkg/world` package provides comprehensive world state management including persistence, chunk streaming, meta-game events, server rankings, and territory control. Overall implementation is solid with 88.4% test coverage exceeding the 65% target. The primary issues are non-deterministic timestamp usage in meta-game events and territory systems (acceptable for real-time multiplayer), and a housing test failure due to Ebiten/GLFW initialization requirements.

## Issues Found
- [ ] <severity:low> **Non-deterministic timestamps** — Extensive use of `time.Now()` for event timing in meta-game, territory, and economy systems (`metagame.go:85-185`, `territory.go:61-190`, `chunk_modification.go:102`, `economy/pricing_engine.go:35-181`, `economy/guild_bank.go:117-502`, `economy/marketplace.go:66-307`, `raids/generator.go:236`, `raids/lockout.go:47-164`, `raids/instance.go:51-158`, `housing/types.go:123-316`, `housing/guildhall.go:123-193`, `territory/siege.go:102-419`, `territory/manager.go:46-298`). This is acceptable for multiplayer game events requiring wall-clock synchronization, but should be documented.
- [ ] <severity:low> **Test isolation** — Housing package tests fail due to Ebiten GLFW initialization requirement (`pkg/world/housing` test failure: "The DISPLAY environment variable is missing"). Tests should use stub implementations or skip graphics tests in headless environments.
- [ ] <severity:low> **No structured logging** — Package does not use `logrus.WithFields` for logging. All logging is via comments in doc.go examples only. Consider adding structured logging for error paths, especially in persistence operations.
- [ ] <severity:low> **Missing integration documentation** — While persistence and chunk systems are well-documented, integration points with engine ECS systems (how to register ChunkLoaderSystem, ChunkModificationSystem, ChunkCompressionSystem) are not documented. Add integration examples to doc.go.

## Test Coverage
88.4% (main package), 88.5% (economy), 91.8% (raids), 93.9% (territory) — **Exceeds 65% target ✓**

Housing sub-package test failure is environmental (GLFW/display), not code quality issue.

## Integration Status
**Core Integration**: The package provides foundational world state types (PersistentWorldState, Chunk, EntityState, WorldEvent) and managers (WorldPersistence, EventManager, RankingManager, TerritoryManager) used throughout the codebase.

**Persistence**: WorldPersistence integrates with saveload system. Uses JSON+gzip serialization with backup rotation. Entity components serialize via generic `map[string]interface{}` - serialization responsibility left to component owners.

**ECS Compliance**: Housing package provides `HousingComponent` (pure data + `Type()` method only) - **compliant ✓**

**Registration**: Chunk systems (ChunkLoaderSystem, ChunkModificationSystem, ChunkCompressionSystem) are NOT registered in engine system_init.go. These appear to be standalone utilities rather than ECS systems - **acceptable design choice**.

**Deterministic Generation**: EventManager uses seed-based RNG (`rand.New(rand.NewSource(seed))` at `metagame.go:70`) for procedural content (threat levels at `metagame.go:128`). Other randomness is time-based for multiplayer events - **compliant with multiplayer exemption**.

## Recommendations
1. **Add environment detection to housing tests** — Skip Ebiten-dependent tests when `DISPLAY` is unavailable or add headless test mode. See `pkg/visualtest` for examples of headless testing patterns.
2. **Document time.Now() usage policy** — Add comment to metagame.go, territory.go explaining that wall-clock time is intentional for multiplayer event synchronization (not a determinism violation).
3. **Add structured logging to persistence** — Use `logrus.WithFields` in SaveWorld, LoadWorld, and backup rotation operations for better observability. Example: `log.WithFields(log.Fields{"save_path": w.SavePath, "chunks": len(state.ChunkData)}).Info("saving world")`.
4. **Document integration patterns** — Add section to doc.go showing how to integrate chunk systems with engine ECS, or clarify they are standalone utilities not requiring ECS registration.
5. **Consider deterministic event IDs** — EventManager uses sequential IDs (`em.nextID++`). This could desync in distributed scenarios. Consider UUID or hash-based IDs for federated events.
