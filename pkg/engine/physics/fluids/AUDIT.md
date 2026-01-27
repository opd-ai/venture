# Package Audit: pkg/engine/physics/fluids
Generated: 2026-01-27

## AUDIT SUMMARY

| Category | Count |
|----------|-------|
| CRITICAL BUG | 0 |
| FUNCTIONAL MISMATCH | 3 |
| MISSING FEATURE | 0 |
| EDGE CASE BUG | 2 |
| PERFORMANCE ISSUE | 0 |
| DOCUMENTATION DISCREPANCY | 3 |

**Total Issues: 8**

## Code Quality Metrics
- **Test Coverage**: 95.3% of statements
- **Total Tests**: 57 tests (all passing)
- **Benchmarks**: 11 benchmark functions
- **Race Conditions**: None detected
- **Go Vet Issues**: None

---

## DETAILED FINDINGS

---

### FUNCTIONAL MISMATCH: Lava Density Inconsistent with Documentation

**File:** types.go:163-170
**Severity:** Low
**Description:** The README.md documents lava density as 3100 kg/m³, but the implementation uses 3000 kg/m³. While this affects buoyancy calculations, the difference is minor.
**Expected Behavior:** Lava density should be 3100 kg/m³ as documented in README.md (line 105)
**Actual Behavior:** Lava density is 3000 kg/m³ in GetFluidProperties()
**Impact:** Buoyancy calculations for objects in lava will be slightly less accurate than documented. Objects may float slightly less than expected.
**Reproduction:** Call `GetFluidProperties(FluidLava)` and check `.Density`
**Code Reference:**
```go
// types.go:163-170
case FluidLava:
    return FluidProperties{
        Viscosity:    0.6,
        Density:      3000.0,  // README says 3100
        FlowRate:     0.3,
        Damage:       50.0,
        Color:        color.RGBA{R: 255, G: 69, B: 0, A: 220},
        Transparency: 0.2,
    }
```

---

### FUNCTIONAL MISMATCH: Oil Density Inconsistent with Documentation

**File:** types.go:171-180
**Severity:** Low
**Description:** The README.md documents oil density as 800 kg/m³, but the implementation uses 900 kg/m³.
**Expected Behavior:** Oil density should be 800 kg/m³ as documented in README.md (line 105)
**Actual Behavior:** Oil density is 900 kg/m³ in GetFluidProperties()
**Impact:** Oil will be denser than documented, which affects buoyancy behavior (objects float more in oil than documented).
**Reproduction:** Call `GetFluidProperties(FluidOil)` and check `.Density`
**Code Reference:**
```go
// types.go:171-180
case FluidOil:
    return FluidProperties{
        Viscosity:    0.3,
        Density:      900.0,  // README says 800
        FlowRate:     0.6,
        Damage:       0.0,
        Color:        color.RGBA{R: 50, G: 50, B: 50, A: 150},
        Transparency: 0.5,
    }
```

---

### FUNCTIONAL MISMATCH: Lava Viscosity Documented as "High" but Uses 0.6

**File:** types.go:164, README.md:105
**Severity:** Low
**Description:** README describes lava as having "High viscosity" (0.8 per AUDIT line 86), but implementation uses 0.6 which is closer to "medium" viscosity (oil uses 0.3, water uses 0.1).
**Expected Behavior:** Lava should have viscosity of ~0.8 to match "high viscosity" documentation
**Actual Behavior:** Lava viscosity is 0.6
**Impact:** Lava flows faster than documented "high viscosity" would suggest
**Reproduction:** Call `GetFluidProperties(FluidLava)` and check `.Viscosity`
**Code Reference:**
```go
// types.go:163-170
case FluidLava:
    return FluidProperties{
        Viscosity:    0.6,  // "High" should be ~0.8
        ...
    }
```

---

### EDGE CASE BUG: GetSwimSpeedMultiplier Division by Zero

**File:** swimming.go:70
**Severity:** Medium
**Description:** The `GetSwimSpeedMultiplier` function divides by `swimming.MaxStamina` without checking for zero, which could cause a division by zero panic if MaxStamina is zero.
**Expected Behavior:** Function should safely handle MaxStamina=0 case
**Actual Behavior:** Division by zero produces NaN or Inf, which propagates through physics calculations
**Impact:** Could cause erratic behavior or NaN propagation in physics if a SwimmingComponent is created with MaxStamina=0
**Reproduction:** Create SwimmingComponent with MaxStamina=0 and IsSwimming=true, then call GetSwimSpeedMultiplier
**Code Reference:**
```go
// swimming.go:64-72
func (s *SwimmingManager) GetSwimSpeedMultiplier(swimming *SwimmingComponent) float64 {
    if !swimming.IsSwimming {
        return 1.0
    }

    // Speed scales with stamina
    staminaRatio := swimming.Stamina / swimming.MaxStamina  // Division by zero possible
    return swimming.SwimSpeed * (0.5 + 0.5*staminaRatio)
}
```

---

### EDGE CASE BUG: UpdateSwimming Division by Zero with MaxStamina=0

**File:** swimming.go:39
**Severity:** Medium
**Description:** Similar to GetSwimSpeedMultiplier, the StaminaRegen calculation `math.Min(swimming.MaxStamina, swimming.Stamina+...)` works correctly, but stamina could exceed MaxStamina if MaxStamina is 0 and StaminaRegen > 0, leading to staminaRatio > 1 in speed calculations.
**Expected Behavior:** Stamina should never exceed MaxStamina; MaxStamina=0 should be handled gracefully
**Actual Behavior:** With MaxStamina=0, stamina can become positive through regen, breaking ratio assumptions
**Impact:** Speed multiplier calculations produce unexpected results when MaxStamina=0
**Reproduction:** Set MaxStamina=0, Stamina=0, StaminaRegen=20, call UpdateSwimming with fluidAmount=0 (on land)
**Code Reference:**
```go
// swimming.go:33-40
if !inWater {
    swimming.IsSwimming = false
    swimming.TreadingWater = false
    swimming.Drowning = false
    // If MaxStamina=0 and StaminaRegen>0, Stamina becomes positive
    swimming.Stamina = math.Min(swimming.MaxStamina, swimming.Stamina+swimming.StaminaRegen*deltaTime)
    return
}
```

---

### DOCUMENTATION DISCREPANCY: DefaultSimulationConfig MaxFluidPerCell Not Implemented

**File:** README.md:179, types.go:134-147
**Severity:** Low
**Description:** README.md line 179 references `config.MaxFluidPerCell = 1.0` in DefaultSimulationConfig, but this field does not exist in SimulationConfig struct.
**Expected Behavior:** SimulationConfig should have MaxFluidPerCell field as documented
**Actual Behavior:** SimulationConfig has no MaxFluidPerCell field; max is hardcoded to 1.0 in AddFluid
**Impact:** Users cannot configure max fluid per cell as documented example suggests
**Reproduction:** Check SimulationConfig struct - no MaxFluidPerCell field exists
**Code Reference:**
```go
// README.md:179 (documentation)
// config.MaxFluidPerCell = 1.0

// types.go:122-132 - actual struct
type SimulationConfig struct {
    GridWidth       int
    GridHeight      int
    CellSize        float64
    UpdateRate      float64
    Gravity         float64
    PressureFactor  float64
    ViscosityFactor float64
    MaxIterations   int
    Convergence     float64
    // MaxFluidPerCell is MISSING
}
```

---

### DOCUMENTATION DISCREPANCY: FloodingComponent FloodLevel Documentation Incorrect

**File:** README.md:86-87, types.go:100-107
**Severity:** Low
**Description:** README.md (lines 86-87) shows FloodLevel as units (0.0 to 100.0) in the example, but the FloodingComponent comment documents it as 0.0-1.0 normalized range.
**Expected Behavior:** Consistent documentation between README and code comments
**Actual Behavior:** README shows 100.0 max, code comment says 0.0-1.0
**Impact:** Confusion about expected FloodLevel ranges
**Reproduction:** Compare README.md line 87 (`MaxFloodLevel: 100.0`) with types.go line 103 comment (`FloodLevel float64 // Current flood level (0.0-1.0)`)
**Code Reference:**
```go
// README.md:86-87
flooding := &fluids.FloodingComponent{
    FloodLevel:     0.0,
    MaxFloodLevel:  100.0,  // Uses 100.0
    ...
}

// types.go:103
FloodLevel    float64 // Current flood level (0.0-1.0)  // Says 0.0-1.0
```

---

### DOCUMENTATION DISCREPANCY: SwimmingComponent Default DrowningDamage Mismatch

**File:** doc.go:49, README.md:73
**Severity:** Low
**Description:** doc.go line 49 states default drowning damage is 10 damage/sec, but README.md line 73 shows `DrowningDamage: 5.0`. These should be consistent.
**Expected Behavior:** Consistent default values in documentation
**Actual Behavior:** doc.go says 10, README says 5.0
**Impact:** User confusion about expected default values
**Reproduction:** Compare doc.go line 49 with README.md line 73
**Code Reference:**
```go
// doc.go:49
//   - Drowning damage: configurable damage/sec (default: 10)

// README.md:73
//     DrowningDamage: 5.0,  // per second
```

---

## Previous Audit Status (2026-01-20)

The previous audit reported 0 issues. This updated audit identified 8 issues:
- 3 functional mismatches between documented and implemented fluid properties
- 2 edge case bugs with zero MaxStamina handling
- 3 documentation discrepancies

All tests continue to pass (57/57) with 95.3% coverage. No race conditions detected. The issues identified are primarily documentation/implementation alignment and edge case handling.

## Recommendations

1. **High Priority**: Add MaxStamina=0 guard in GetSwimSpeedMultiplier to prevent NaN
2. **Medium Priority**: Align fluid property values (density, viscosity) with README documentation
3. **Low Priority**: Add MaxFluidPerCell to SimulationConfig or remove from README
4. **Low Priority**: Standardize documentation for FloodLevel range (0.0-1.0 or units)

## Conclusion

The pkg/engine/physics/fluids package is well-implemented with high test coverage (95.3%). The identified issues are primarily documentation-code alignment discrepancies and edge case handling for zero MaxStamina. No critical bugs or race conditions were found. Core functionality operates correctly for standard use cases.
