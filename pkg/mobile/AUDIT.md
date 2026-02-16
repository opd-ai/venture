# Audit: pkg/mobile/

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: 65.4%

## Summary

Audited the mobile platform package covering touch input, virtual controls, gesture detection, accessibility, and mobile UI widgets.

## Issues Found and Fixed

### Medium Severity

1. **Fixed: Joystick direction calculation no-op** (`dual_joystick.go:303-304`)
   - `math.Max(1.0, j.Magnitude)` always returned 1.0 since Magnitude is 0-1
   - Simplified to `dx / j.Radius` to correctly represent normalized stick position
   - DirectionX/Y now clearly represent raw position; Magnitude gives dead-zone-adjusted intensity

2. **Fixed: Swipe vertical detection misclassified left swipes** (`ui.go:206`)
   - Old check `direction > 1.0 || direction < -1.0` incorrectly treated left swipes (±π radians) as vertical
   - Replaced with proper angle range check: `absDir > π/4 && absDir < 3π/4`

3. **Fixed: Division by zero with empty menu items** (`ui.go:133,157`)
   - `handleTapInput()` and `handleLongPressInput()` divided by `len(m.Items)` without guard
   - Added early return when `len(m.Items) == 0`

4. **Fixed: Inconsistent item height across layout methods** (`ui.go`)
   - `calculateMaxScroll()` used hardcoded 50px, `handleTapInput()` used dynamic height
   - Introduced `getItemHeight()` helper with 48px minimum (touch target safe) used consistently across tap detection, scroll calculation, long-press detection, and rendering

### Low Severity (Not Fixed)

5. **Resize() doesn't validate screenWidth/screenHeight > 0** (`controls.go:740`)
   - Could produce negative/zero sizes; unlikely in practice since Ebiten validates screen dimensions

6. **ApplyInputResponseCurve doesn't validate curvePower > 0** (`dual_joystick.go:363`)
   - Negative curvePower with `math.Pow()` could produce unexpected results
   - Low risk: callers are expected to pass valid values

## Tests Added

- `TestMobileMenu_GetItemHeight` — Table-driven test for consistent item height calculation
- `TestMobileMenu_HandleTapEmptyMenu` — Verifies no panic on empty menu tap/long-press
- `TestMobileMenu_SwipeVerticalDetection` — Verifies swipe angle classification correctness
- Updated `TestMobileMenu_CalculateMaxScroll` to match new dynamic height behavior
