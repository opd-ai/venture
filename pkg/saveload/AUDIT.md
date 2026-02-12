# Audit: pkg/saveload
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The saveload package provides persistent game state management with file-based serialization (desktop) and localStorage (WASM). Overall health is good with 83.8% test coverage, comprehensive error handling, and robust recovery mechanisms. Critical risk: API inconsistency between desktop and WASM implementations could cause compilation errors when migrator functionality is needed on WASM.

## Issues Found
- [x] **high** API Inconsistency — WASM `SaveManager` missing `SetMigrator`, `NewSaveManagerWithLogger`, `NewSaveManagerWithMigrator` methods that exist in desktop version (`storage_wasm.go:30-70`, `manager.go:39-79`)
- [x] **med** Documentation Gap — WASM-specific limitations not documented in package doc.go (migration not supported on WASM) (`storage_wasm.go:384-399`, `doc.go:1-52`)
- [x] **low** Code Duplication — `validateSaveName` implementation duplicated between manager.go and storage_wasm.go instead of shared helper (`manager.go:281-299`, `storage_wasm.go:362-380`)

## Test Coverage
83.8% (target: 65%) ✅

Coverage by file:
- manager.go: Well covered with table-driven tests
- migrator.go: Comprehensive migration path testing
- recovery.go: Backup/checksum scenarios covered
- serialization.go: Item/spell conversion tested
- storage_wasm.go: Not directly testable (requires js runtime)

## Integration Status
Successfully integrated with:
- **Engine**: `pkg/engine/menu_system.go` uses SaveManager for save/load UI (lines 14, 60, 85, 762, 785)
- **Migration**: `pkg/migration/validator.go` imports for save validation
- **No registration needed**: Pure data package, no system/component registration

Missing integrations:
- None identified - package is complete for its scope

## Recommendations
1. **HIGH PRIORITY**: Add missing methods to WASM SaveManager to match desktop API:
   ```go
   // storage_wasm.go after line 70
   func NewSaveManagerWithLogger(saveDir string, logger *logrus.Logger) (*SaveManager, error)
   func NewSaveManagerWithMigrator(saveDir string, logger *logrus.Logger, migrator Migrator) (*SaveManager, error)
   func (m *SaveManager) SetMigrator(migrator Migrator)
   ```
   Implementation should log warning that migration is not supported on WASM and set migrator to nil.

2. **MEDIUM PRIORITY**: Document WASM limitations in doc.go:
   ```go
   // ## Platform-Specific Behavior
   //
   // Desktop (Linux/macOS/Windows):
   //   - Uses file-based persistence with .sav extension
   //   - Supports save migration via Migrator interface
   //   - SHA256 checksums for integrity validation
   //
   // WASM (Browser):
   //   - Uses localStorage API (5MB limit per origin)
   //   - No migration support (incompatible versions rejected)
   //   - FNV-1a checksums for integrity validation
   //   - Falls back to in-memory storage if localStorage unavailable
   ```

3. **LOW PRIORITY**: Extract `validateSaveName` to shared helper function to eliminate duplication:
   ```go
   // types.go or new validation.go
   func ValidateSaveName(name string) error { ... }
   ```
   Update both manager.go and storage_wasm.go to use shared implementation.

## Compliance Checklist

### Stub/Incomplete Code ✅
- No functions returning only nil/zero values
- No TODO/FIXME/placeholder comments found
- All methods have complete implementations

### ECS Compliance ✅
- Not applicable - saveload is pure data/persistence package
- No components or systems defined
- Correctly avoids circular dependencies with engine

### Deterministic Procgen ✅
- No random number generation used
- `time.Now()` only used for save timestamps (appropriate)
- All data serialization is deterministic (JSON marshaling)

### Network Interfaces ✅
- Not applicable - no network code in this package

### Error Handling ✅
- All errors checked and propagated with context
- Structured logging with logrus.WithFields on error paths
- Comprehensive validation before operations
- Examples:
  - `manager.go:61` - Error wrapping with context
  - `manager.go:137` - Error propagation with validation
  - `storage_wasm.go:166` - Validation errors logged to console

### Test Coverage ✅
- 83.8% overall coverage exceeds 65% target
- Table-driven tests for manager, migrator, recovery, serialization
- Benchmarks present in saveload_bench_test.go
- Edge cases tested (nil values, invalid names, corruption scenarios)

### Doc Coverage ✅
- Package doc.go with comprehensive overview
- All exported types have godoc comments (36 types documented)
- All exported functions have godoc comments
- README.md with usage examples and feature list

### Integration Points ✅
- Used by engine/menu_system.go for save/load UI
- Used by migration/validator.go for data validation
- No system registration needed (not an ECS system)
- No component serialization needed (handles data types only)

## Additional Notes

**Strengths**:
- Excellent separation of desktop vs WASM implementations with build tags
- Robust error recovery with backup/checksum validation
- Clean API design with sensible defaults
- Comprehensive test suite with good edge case coverage
- Security conscious (save name validation prevents path traversal)

**Architecture Decisions**:
- JSON format chosen for human readability (good for modding/debugging)
- WASM uses FNV-1a instead of SHA256 for WASM compatibility (acceptable trade-off)
- In-memory fallback on WASM when localStorage unavailable (good UX)
- Migrator optional (nil = reject incompatible versions) - allows strict version control

**Performance Notes**:
- JSON marshaling may be slow for large save files (acceptable for save/load operations)
- localStorage 5MB limit documented and enforced on WASM
- No performance-critical code paths identified

**Security Considerations**:
- Save name validation prevents directory traversal (lines manager.go:281-299)
- No executable code in save files (pure JSON data)
- Checksum validation detects tampering/corruption

## Files Audited
- doc.go (52 lines)
- types.go (639 lines)
- manager.go (435 lines)
- migrator.go (194 lines)
- recovery.go (459 lines)
- serialization.go (323 lines)
- storage_wasm.go (767 lines)
- manager_test.go (1274 lines)
- migrator_test.go (348 lines)
- recovery_test.go (520 lines)
- serialization_test.go (749 lines)
- saveload_bench_test.go (617 lines)
- animation_test.go (375 lines)

**Total**: 6,752 lines (3,869 implementation + 2,883 tests)
