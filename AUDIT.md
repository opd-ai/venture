# IMPLEMENTATION GAP AUDIT — 2026-04-22

## Project Architecture Overview
Venture is a seed-deterministic, zero-asset multiplayer action-RPG using ECS (`pkg/engine`), procedural generation (`pkg/procgen`), runtime rendering/audio (`pkg/rendering`, `pkg/audio`), authoritative networking (`pkg/network`), and persistence/world systems (`pkg/world`, `pkg/saveload`).

Phase-0 evidence reviewed: `README.md`, `ROADMAP.md`, `docs/ARCHITECTURE.md`, `docs/TECHNICAL_SPEC.md`, `go.mod`, package inventory (`go list ./...` = 124 packages), dependency graph (`/tmp/gap-go-deps.txt`), and go-stats metrics (`/tmp/gap-audit-metrics.json`).

Online research (brief): repository currently has no open issues (`list_issues`/`search_issues` returned 0); recent PR stream is audit/backlog-heavy (#472–#483), indicating ongoing gap-closure cycles rather than milestone-tracked open gaps.

Baseline execution:
- `go-stats-generator analyze . --skip-tests` completed.
- `go build ./...` and `go vet ./...` both failed in this environment due missing X11 headers (`X11/Xlib.h`), not a Go code compile error in project sources (`/tmp/gap-build-results.txt`, `/tmp/gap-vet-results.txt`).

## Gap Summary
| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 2 | 0 | 0 | 1 | 1 |
| Dead Code | 2 | 0 | 0 | 1 | 1 |
| Partially Wired | 2 | 0 | 1 | 1 | 0 |
| Interface Gaps | 2 | 0 | 1 | 0 | 1 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |

## Implementation Completeness by Package
| Package | Exported Functions | Implemented | Stubs | Dead | Coverage |
|---------|-------------------:|------------:|------:|-----:|----------|
| engine | 5542 | 5540 | 1 | 1 | N/A (skip-tests baseline) |
| trade | 14 | 13 | 0 | 0 | N/A (skip-tests baseline) |
| validation | 20 | 19 | 0 | 1 | N/A (skip-tests baseline) |
| story | 42 | 40 | 0 | 2 | N/A (skip-tests baseline) |
| prestige | 51 | 49 | 0 | 2 | N/A (skip-tests baseline) |
| quest | 44 | 43 | 1 | 0 | N/A (skip-tests baseline) |

## Findings
### HIGH
- [ ] **Trade quantity contract exists but is non-functional in runtime model** — `pkg/validation/trade.go:95-102`, `pkg/network/trade/doc.go:11`, `pkg/engine/chat_trade_components.go:269-277`, `pkg/network/trade/system.go:131-167` — validator/API/docs describe quantity validation, but trade proposals only carry item IDs and no quantity field, so quantity rules cannot execute in production — **Remediation:** introduce quantity-bearing trade line items in `TradeProposal`, update `ProposeTrade`/validation flow to call `ValidateTradeQuantity`, and add end-to-end trade quantity tests; validate with `go test ./pkg/network/trade ./pkg/validation`.
- [ ] **Weapon material impact particles are instantiated but not connected to combat damage callbacks** — `pkg/engine/weapon_material_impact_particle_system.go:125-130`, `pkg/engine/system_init.go:682-688`, `pkg/engine/system_init.go:1899,1932`, `cmd/client/handlers.go:1808-1812` — system states callback-driven behavior (`OnMeleeImpact`) and is registered in world update loop, but no `CombatSystem.AddDamageCallback` registration exists for it, so melee-impact path is unreachable in live combat — **Remediation:** wire `OnMeleeImpact` via combat callback registration in shared init path(s) and add integration test asserting particle emission on melee hit; validate with `go test ./pkg/engine`.

### MEDIUM
- [ ] **VR hardware adapter path is structurally present but functionally stubbed and unreachable from runtime initialization** — `pkg/engine/vr_openxr_adapters.go:94-229`, `cmd/client/init_versions.go:572-586` — OpenXR adapter methods return placeholder zero/no-op values with TODOs, while VR initialization always installs stub adapters; no runtime path selects OpenXR adapters — **Remediation:** add a `-tags vr` initialization branch that instantiates OpenXR adapters when available and falls back to stubs only on detection/init failure; add vr-tagged integration tests for adapter selection and non-zero pose/controller state; validate with `go test -tags vr ./pkg/engine/...`.
- [ ] **OpenXR adapter constructors have no in-repo callers (dead feature path)** — `pkg/engine/vr_openxr_adapters.go:91,157` and caller search results show no usage outside definitions/docs — feature compiles only under `vr` tag but remains disconnected from command initialization — **Remediation:** consume constructors in vr-tagged client init code, or explicitly deprecate/remove until activation is planned; validate by searching for constructor call sites and running vr-tagged build.

### LOW
- [ ] **Exported sentinel errors defined but never returned** — `pkg/procgen/story/generator.go:62-63`, `pkg/engine/prestige/manager.go:22-23` — `ErrInvalidChoiceIndex`, `ErrInvalidOptionIndex`, `ErrAccountNotFound`, and `ErrUnknownParagonCategory` are declared but have no non-definition references, creating contract noise — **Remediation:** either return these errors from concrete validation paths or remove them; validate by grep/reference scan and package tests.
- [ ] **Stale TODO marker remains after REM-144 migration** — `pkg/procgen/quest/templates.go:6` vs resolved markers in `pkg/procgen/quest/generator.go:89` and `pkg/procgen/quest/types.go:286` — implementation appears complete but TODO text remains partially unresolved, causing audit ambiguity — **Remediation:** update comment to resolved state and link canonical migration note; validate by TODO scan.

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `WeaponMaterialImpactParticleSystem.Update` is empty (`pkg/engine/weapon_material_impact_particle_system.go:126`) | Empty `Update` is intentional for callback-driven architecture; real gap is missing callback wiring, not method body shape. |
| Headless GPU no-op methods (e.g., `pkg/rendering/lighting/gpu_bloom_headless.go:35,38`) | Build-target fallback behavior is intentional for headless paths and does not violate documented behavior. |
| TODO markers with `resolved` suffix in story/quest files | These are historical references to completed removals and not active implementation gaps. |
| `go build`/`go vet` failures during baseline | Failures were environment/toolchain header issues (`X11/Xlib.h` missing), not code-level implementation defects in repo logic. |
