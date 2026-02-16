# Audit: github.com/opd-ai/venture/pkg/engine/saves
**Date**: 2026-02-16
**Status**: Incomplete

## Summary
The `pkg/engine/saves/` directory exists but is **completely empty** with no Go source files. Save/load functionality is implemented in `pkg/saveload/` instead. This directory appears to be unused scaffolding that should either be populated with engine-specific save state management or removed to avoid confusion.

## Issues Found
- [ ] high stub/incomplete — Directory exists but contains zero Go files; purpose unclear (`pkg/engine/saves/`)
- [ ] high integration — Save/load functionality exists in `pkg/saveload/` but `pkg/engine/saves/` is empty, suggesting incomplete migration or abandoned refactoring (`pkg/engine/saves/` vs `pkg/saveload/`)
- [ ] med documentation — Directory lacks README.md or doc.go explaining its intended purpose vs `pkg/saveload/` (`pkg/engine/saves/`)
- [ ] low cleanup — Empty directory tracked in git with no clear purpose; consider removal if truly unused (git history shows creation in commit 6fc68261)

## Test Coverage
0% (no Go files)

## Integration Status
**Not integrated.** The directory is empty. All save/load functionality resides in `pkg/saveload/` which includes:
- `manager.go` - Save/load manager
- `serialization.go` - Entity/world serialization
- `recovery.go` - Save corruption recovery
- `migrator.go` - Save file migration between versions
- `validation.go` - Save data validation
- `storage_wasm.go` - WASM-specific storage

The `pkg/engine/` package does not reference `pkg/engine/saves/` anywhere. The client (`cmd/client/`) imports `pkg/saveload` directly.

**Potential Intent**: Based on naming convention, `pkg/engine/saves/` might have been intended for:
1. Engine-specific save state (component snapshots, entity serialization helpers)
2. Integration layer between ECS engine and `pkg/saveload`
3. Abandoned refactoring to move save logic closer to engine

## Recommendations
1. **CRITICAL**: Clarify architectural intent — decide whether to populate this package or remove it
2. **HIGH**: If keeping, create doc.go explaining separation of concerns between `pkg/engine/saves/` and `pkg/saveload/`
3. **HIGH**: If removing, update AUDIT.md to reflect removal and ensure no future references
4. **MED**: If populating, implement engine-specific save helpers:
   - Component batch serialization (using cached hot-path components)
   - Entity snapshot utilities
   - World state diff generation for incremental saves
   - Integration with `pkg/saveload.Manager`
5. **LOW**: Review git history to determine original architectural plan
