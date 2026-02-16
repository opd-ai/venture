# Audit: pkg/procgen/class

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: 93.0% (up from 86.4%)

## Summary

The class package provides procedural character class generation with 21 class presets (6 base + 15 hybrid). Code is clean, deterministic, and well-structured. Four issues were found and all resolved.

## Issues Found: 4 (0 high, 1 medium, 3 low) — All Fixed

### MEDIUM

1. **Fragile random class selection using hardcoded magic number** — `Generate()` used `rng.Intn(21)` to select a random class, assuming enum values 0–20 are contiguous and exactly match the preset count. If classes are added/removed, this would silently produce errors for some seeds.
   - **Fix**: Replaced with `randomClassType()` helper that selects from actual preset map keys.

### LOW

2. **`GetAllPresets` used hardcoded loop bound** — Iterated `0..20` instead of using `len(g.presets)`.
   - **Fix**: Changed to `len(g.presets)` for the loop bound.

3. **`Validate` missing defense/speed checks** — `StartingDefense` and `StartingSpeed` were not validated, allowing zero or negative values.
   - **Fix**: Added `StartingDefense > 0` and `StartingSpeed > 0` validation.

4. **`Validate` missing name/description checks** — Empty `Name` or `Description` strings would pass validation.
   - **Fix**: Added non-empty checks for both fields.

## Test Improvements

- Added test cases: empty name, empty description, invalid defense, invalid speed, wrong type
- Added `TestGetPreset` for the `GetPreset()` method
- Added `TestGenerateAndValidateAllClasses` to verify all 21 classes generate and validate correctly
- Coverage: 86.4% → 93.0%

## No Remaining Issues
