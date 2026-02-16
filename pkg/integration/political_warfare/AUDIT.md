# Audit: pkg/integration/political_warfare

**Date**: 2026-02-16
**Coverage**: 94.7%
**Status**: Complete

## Issues Found and Fixed

### Medium

1. **Missing self-guild validation** — `DeclareWar`, `SignPeaceTreaty`, `ImposeEmbargo`, and `CallReinforcementAllies` did not check if both guild IDs were the same, allowing a guild to declare war on itself, embargo itself, etc. Added early validation to all four methods.

2. **Missing reverse war check in DeclareWar** — Only checked for existing war key `attacker_defender`; did not check `defender_attacker`. Two guilds could have simultaneous mutual wars. Added reverse direction check.

### Low (Documented, Not Fixed)

1. **Typo in `ResponingAllies` field** (types.go:55) — Should be `RespondingAllies`. Not fixed because it would be a breaking change for any serialized data using JSON field `responding_allies`. The JSON tag is correct (`"responding_allies"`), so only Go code referencing the field is affected.

2. **`calculateConcessionValue` gold type assertion** — Uses `concession.Value.(int)` but JSON deserialization produces `float64`. Could cause silent failures for concession values loaded from JSON. Low impact since concessions are currently only created in-memory.

### Low (Fixed)

3. **Doc.go example showed incorrect embargo value** — Example used `90` as embargo price increase but `ImposeEmbargo` expects a fraction (0.5–0.9). Fixed to show `0.9` with clarifying comment.

## Test Coverage

- 40 tests, all passing
- 94.7% statement coverage
- Includes unit tests, edge cases, benchmarks, save/load round-trip, and integration tests
