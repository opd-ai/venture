# Audit: pkg/procgen/recipe

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: 90.2%

## Summary

The recipe package generates procedural crafting recipes (potions, enchanting, magic items, cooking, smithing) with genre-specific templates and seed-based determinism.

## Issues Found: 4 (0 high remaining, 2 medium fixed, 2 low fixed)

### MEDIUM-001: Empty Template Fallback Could Panic (FIXED)
- **Severity**: Medium
- **Location**: `generator.go`, `generateRecipe()` method
- **Description**: If no templates existed for a genre AND the fantasy fallback was also empty, `templates[rng.Intn(len(templates))]` would panic with index out of range.
- **Fix**: Added cascading fallback: genre → fantasy → default ("") → any available template.

### MEDIUM-002: Rarity Threshold Inversion at High Depths (FIXED)
- **Severity**: Medium
- **Location**: `generator.go`, `calculateRarity()` method
- **Description**: `rarityBonus` was unbounded, causing Common threshold (`0.50 - rarityBonus`) to go negative at high depth/difficulty values, breaking rarity distribution.
- **Fix**: Clamped `rarityBonus` to max 0.45 using `math.Min()`.

### LOW-001: Success Chance Could Go Negative (FIXED)
- **Severity**: Low
- **Location**: `generator.go`, `generateRecipe()` method
- **Description**: `baseSuccess -= float64(rarity) * 0.05` could produce negative success chances for high-rarity recipes with low base success ranges.
- **Fix**: Clamped success chance to `[0.05, 1.0]` using `math.Max/Min`.

### LOW-002: Zero/Negative Count Not Validated (FIXED)
- **Severity**: Low
- **Location**: `generator.go`, `extractRecipeCount()` method
- **Description**: Passing `count: 0` or `count: -1` in Custom params would generate zero recipes, which then fails Validate(). Now defaults to 5 for invalid counts.
- **Fix**: Added `c > 0` guard to count extraction.

## Code Quality Notes
- Deterministic generation: ✅ Uses `rand.New(rand.NewSource(seed))`
- Error handling: ✅ Validate() covers all edge cases
- Test coverage: 90.2% with table-driven tests
- GoDoc: ✅ Comprehensive package and function documentation
