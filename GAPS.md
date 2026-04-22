# Technical Debt Gaps — 2026-04-22

> **Scope.** This document tracks **maintainability** debt — complexity, duplication, coupling, documentation freshness, and quality-infrastructure gaps. Implementation-completeness gaps from earlier revisions (trade-quantity contract, material-impact particle wiring, VR OpenXR adapter, unused sentinel errors, REM-### TODOs) are referenced as `LEGACY-IGAP-1..5` in `AUDIT.md` and remain tracked in `.github/copilot-instructions.md:336–339` and `ROADMAP.md`.
>
> Cross-references in source code (`cmd/client/handlers.go:115,594,766,853,1136`, `cmd/server/main.go:130-521`) that mention "AUDIT.md Gap N" point at those legacy implementation gaps and are intentionally preserved.

---

## Gap 1: Behavior-tree `Tick` methods are 9-way duplicated and over-complex
- **Current State**: `pkg/engine/behavior_tree_advanced_nodes.go` contains nine `Tick` methods at cyclomatic 15–23 (lines 41, 175, 312, 421, 556, 787, 1034, 1167) plus two more in `pkg/engine/behavior_tree_tactical_nodes.go:198,419`. go-stats-generator detects a 40-line renamed clone group at `behavior_tree_advanced_nodes.go:922 ↔ :1067` and many smaller (28–34 line) clones across the family. Average pattern: faction lookup → target resolution → distance/timeout bookkeeping → per-node decision.
- **Impact**: Every fix to common preflight logic must be repeated 9–11 times; bug-fix latency is high in the AI subsystem and behaviour drift between nodes is invisible. AI is on the player-facing critical path.
- **Remediation**: Extract a shared scaffold in the same package, e.g. `runMovementTick(entity *Entity, bb *Blackboard, dt float64, p tickParams, decide func(ctx tickCtx) NodeStatus) NodeStatus`, that owns faction lookup, target acquisition, distance test, and timeout/`timeMoving` updates. Migrate one `Tick` per PR. Do **not** try to unify behaviourally-different nodes into a generic engine — extract the *common* prefix only.
- **Effort**: medium (≈4–6 hours total; one node per PR keeps blast radius small)
- **Dependencies**: none
- **Quick Win**: no — multi-PR sequence
- **Validation**: `xvfb-run -a go test ./pkg/engine/... -run BehaviorTree -race` after each migration; `go-stats-generator analyze . --skip-tests | grep behavior_tree` should show all `Tick` cyclomatic dropping below 15.

## Gap 2: `pkg/engine` is a god package (583 files, 9,534 funcs, 64 deps)
- **Current State**: `pkg/engine` accounts for ~94 % of repo functions and 64 of the project's package dependencies. Cohesion analysis classifies it as healthy internally, but the surface area makes targeted reasoning nearly impossible. Highest-affinity natural sub-packages already exist as filename prefixes: `behavior_tree_*`, `specialization_*`, `weather_*`, `physics/*` (already extracted).
- **Impact**: Onboarding hostile (a new contributor has to read >200 KLOC); IDE and `gopls` slow; every PR touches a "shared" namespace, increasing merge conflicts. Bus factor risk.
- **Remediation**: Incremental extraction by responsibility. Recommended first three sub-packages, in order: (a) `pkg/engine/ai/behavior` ← all `behavior_tree_*.go`; (b) `pkg/engine/specialization` ← all `specialization_*_system.go`; (c) `pkg/engine/weather` ← all `weather_*_system.go`. Each move = `git mv` + package-clause edit + `goimports -w` + adjust `pkg/engine/system_init.go` import + verify tests. Avoid touching `ecs.go`, `interfaces.go`, `system_init.go` themselves until at least three sub-packages are out.
- **Effort**: large (multi-week, one sub-package per PR; first PR ≈ 1 day including review)
- **Dependencies**: Gap 1 should land first so the behaviour-tree extraction is the "cleanest" move and demonstrates the pattern.
- **Quick Win**: no
- **Validation**: `go build ./... && xvfb-run -a go test ./pkg/engine/... ./pkg/integration/... ./cmd/...`. Re-run `go-stats-generator` and confirm `engine` package file/function counts decrease.

## Gap 3: `cmd/client/handlers.go` is highest-churn-and-duplicated file
- **Current State**: 3,880 lines, 25 modifications in the last 200 commits. go-stats-generator flags clone groups at `handlers.go:1730 ↔ :1823 ↔ :1731 ↔ :1824` (35–36 lines, 7+ instances) and a 28-line clone at `handlers.go:1864`. Mixes voice-transport wiring (lines 115, 594, 766, 853, 1136 — referenced as `AUDIT.md Gap 1` legacy markers), UI callback registration (line 2989), and system initialization (line 3271). Function `setupUICallbacks` takes 9 params; `initializeUIIntegration` takes 8.
- **Impact**: Highest single-file maintainability risk in the codebase: large + churned + duplicated + cross-cutting. Any refactor in voice/UI/system wiring carries broad regression risk.
- **Remediation**: Two independent steps. (1) Extract the duplicated 28–36-line callback-registration blocks at `handlers.go:1730–1864` into a single `registerSystemCallbacks(g *EbitenGame, deps callbackDeps) error` helper in a new `cmd/client/handlers_callbacks.go`. (2) Split `handlers.go` by responsibility into `handlers_voice.go` (voice transport wiring), `handlers_ui_callbacks.go` (UI/menu/HUD callbacks), `handlers_system_init.go` (post-`InitializeGameSystems` wiring). Pure file moves with package-private symbols — no API change.
- **Effort**: medium (Step 1 ≈ 2–3 hours; Step 2 ≈ 2 hours, mostly mechanical)
- **Dependencies**: none — this file is package `main`, so no external importers are affected.
- **Quick Win**: yes (Step 1 only — one PR, immediate clone-count reduction).
- **Validation**: `go build ./cmd/client/... && go vet ./cmd/client/...`; `xvfb-run -a go test ./cmd/client/...`.

## Gap 4: Stat-modifier dispatch duplicated across specialization / weather / dual-class systems
- **Current State**: Renamed clones of 34–39 lines: `pkg/engine/specialization_defense_system.go:80 ↔ specialization_evasion_system.go:97`; `specialization_crit_damage_system.go:76 ↔ specialization_lifesteal_system.go:76`; `weather_block_chance_system.go:111 ↔ weather_crit_chance_system.go:97`; plus `pkg/engine/dual_class_synergy_system.go:349` (`applySynergyBonus`, cyc 16, 60 lines) and `:413` (`removeSynergyBonus`, cyc 14). All implement the same "apply or remove a stat-bonus block on an entity" pattern with minor variation.
- **Impact**: Adding a new specialization / weather modifier requires copy-paste from the nearest sibling. Bonuses are silently inconsistent if one is updated and others are not (a common kind of game-balance bug).
- **Remediation**: Introduce `pkg/engine/statmod` (or a sub-file in engine until Gap 2 lands) exporting `Apply(entity *Entity, mod StatModSpec) AppliedHandle` and `Remove(entity *Entity, h AppliedHandle)`. Migrate one consumer per PR; the dual-class pair is the easiest first because its tests are already isolated.
- **Effort**: medium (≈1 hour per migration × ~8 systems)
- **Dependencies**: ideally after Gap 2 step (a) so the new helper has a natural home, but can be done in `pkg/engine` directly first.
- **Quick Win**: no — incremental
- **Validation**: `xvfb-run -a go test ./pkg/engine/... -run 'Specialization|Weather|DualClass'` after each migration; `go-stats-generator` clone count drops.

## Gap 5: `InitializeGameSystems` has 12 internal duplication groups
- **Current State**: `pkg/engine/system_init.go:301` is intentionally a 1944-line linear initializer (justified inline at `:289–297`), but go-stats-generator detects 12 internal clone groups (28–34 lines) at lines 433, 596, 650, 652, 656, 657, 658, 778–787, 1425, 1432, 1528, 1705–1713 — these are repeated `if config.X { sys := NewX(...); world.AddSystem(sys); result.X = sys }` patterns.
- **Impact**: Adding a new system means yet another copy-paste; 11 recent commits to this file show the churn already happens.
- **Remediation**: Add `registerSystemBatch(world *World, regs []systemRegistration)` helper where `systemRegistration` is `{name string, enabled bool, factory func() System, postInit func(System)}`. Migrate one logical batch (e.g. all `weather_*` registrations, or all `specialization_*` registrations) per PR. Keep the function's single-location-of-truth property — do **not** scatter registrations across multiple files.
- **Effort**: small per batch (~30 min); cumulatively medium (~6 batches)
- **Dependencies**: none
- **Quick Win**: yes (first batch only)
- **Validation**: Startup logs unchanged; `xvfb-run -a go test ./pkg/engine/... ./cmd/server/...`.

## Gap 6: 13- and 11-parameter helpers in `cmd/client/util.go`
- **Current State**: `createDeathCallback` at `cmd/client/util.go:1440` takes **13** parameters; `spawnProceduralLoot` at `:1336` takes 11. Both are used from `handlers.go` and `init_versions.go`. They are by far the worst signatures in the codebase (next is 10-param `buildFurniture` in `pkg/procgen/furniture/generator.go:83`).
- **Impact**: Call sites are unreadable and order-sensitive; adding any new dependency to these helpers cascades to every caller.
- **Remediation**: Introduce parameter structs in the same file: `type DeathCallbackDeps struct { … }` and `type LootSpawnDeps struct { … }`. Change the function signatures to accept the struct (and a few hot params if needed). Migrate call sites mechanically. Keep the same field names as the old parameter names to minimize diff noise.
- **Effort**: small (~2 hours total)
- **Dependencies**: none
- **Quick Win**: yes
- **Validation**: `go vet ./cmd/client/... && go build ./cmd/client/...`; manual smoke test with `xvfb-run -a ./venture-client` (existing CI builds the client).

## Gap 7: `cmd/client/util.go` is a 2698-line generic dumping ground
- **Current State**: 85 functions / 2 types in one file; flagged by go-stats-generator as `generic_name` violation. Per repo memory it contains `canonicalGenreID` (genre normalization), procedural loot spawners, death callbacks, and miscellaneous client helpers.
- **Impact**: New contributors cannot guess where a helper lives; `gopls` operations are slower; `git blame` is noisy because every client change touches it.
- **Remediation**: Split into responsibility-named files: `client_loot.go`, `client_callbacks.go`, `client_genre.go`, `client_input.go`, `client_misc.go`. Pure `git mv` of function bodies — no signature changes. Best done **after** Gap 6 so the new parameter structs land cleanly.
- **Effort**: small (~1 hour)
- **Dependencies**: Gap 6 (preferred ordering)
- **Quick Win**: yes
- **Validation**: `go build ./cmd/client/...`.

## Gap 8: Engine package is not race-tested in CI
- **Current State**: `.github/workflows/test.yml` runs `-race` only on `pkg/world/housing/...` and `pkg/integration/...`. Per repo memory, `pkg/engine/ecs_test.go:468/477`, `pkg/engine/expression_combo_test.go:84`, `pkg/engine/guild_vehicle_system_test.go:100`, and `pkg/procgen/audit/determinism_test.go:306` are known-flaky/racy.
- **Impact**: New concurrency bugs in the largest package (where systems share `World` state across goroutines for networking and audio) cannot be caught at PR time.
- **Remediation**: One PR per flaky test to stabilize: deterministic ordering, sync.Mutex on shared state, or remove `t.Parallel()` for tests that mutate package-level fixtures. After all four are green, append `xvfb-run -s "-screen 0 1920x1080x24" go test -race ./pkg/engine/...` to `.github/workflows/test.yml` (after the existing `Run integration tests with race detector` step).
- **Effort**: medium (≈1–2 hours per test + CI step)
- **Dependencies**: none
- **Quick Win**: no — sequential
- **Validation**: each PR must show `xvfb-run -a go test -race -run <TestName> -count=20 ./pkg/engine/` passing; final PR adds the CI step.

## Gap 9: `make lint` runs no static analyzer beyond `go vet`
- **Current State**: `Makefile` `lint:` target runs `go vet ./...` and `scripts/validate-network-types.sh`. CI mirrors this. No `staticcheck`, `golangci-lint`, `errcheck`, or `ineffassign`.
- **Impact**: Common maintainability issues (unused parameters, error-shadowing, ineffectual assignments, deprecated APIs) accrue silently and only get caught by reviewer attention.
- **Remediation**: Add `staticcheck` (single-binary, fast, low false-positive rate) to `make lint`: `go install honnef.co/go/tools/cmd/staticcheck@latest && staticcheck ./...`. Add a corresponding step to `.github/workflows/test.yml`. **Do not** require zero warnings on day one — first run will be noisy; merge with `staticcheck ./... || true` initially and tighten in a follow-up PR after triaging findings.
- **Effort**: small (~1 hour)
- **Dependencies**: none
- **Quick Win**: yes (the *adoption* PR; subsequent fix-up PRs vary)
- **Validation**: `make lint` exits cleanly locally; CI step visible on next PR.

## Gap 10: `pkg/procgen/audit` has 16 deps in only 3 files / 5 funcs
- **Current State**: `pkg/procgen/audit` is the highest deps-per-file outlier in the repo (per go-stats-generator package coupling analysis). It imports many `pkg/procgen/*` subsystems to drive cross-cutting determinism / quality audits.
- **Impact**: Any rename in an audited subsystem ripples here; conversely, this package becomes a bottleneck for adding new audits.
- **Remediation**: Move each subsystem-specific audit helper into the package it audits (e.g. `pkg/procgen/quest/audit.go`, `pkg/procgen/story/audit.go`) exposing a small `Audit(t *testing.T)` or `Audit(ctx) Result` function. The top-level `pkg/procgen/audit` package becomes a thin driver that imports a registry slice from each subsystem. Net effect: dependency count drops to ~1 (just the registry types).
- **Effort**: small per subsystem (~30 min × ~5 subsystems)
- **Dependencies**: none
- **Quick Win**: no — needs coordination across subsystems
- **Validation**: `xvfb-run -a go test ./pkg/procgen/... ./pkg/procgen/audit/...`; verify new dep count via `go-stats-generator`.

## Gap 11: `pkg/errors` package collides with stdlib name
- **Current State**: go-stats-generator naming analysis reports `pkg/errors` as `stdlib_collision`.
- **Impact**: Confuses readers and tooling that auto-import `errors`. Minor but real cognitive cost.
- **Remediation**: Rename `pkg/errors` → `pkg/venterr` (or `pkg/errs`) using `gopls rename` from the package directory; commit as a single PR. All callers update mechanically.
- **Effort**: small (~30 min)
- **Dependencies**: none
- **Quick Win**: yes
- **Validation**: `go build ./... && xvfb-run -a go test ./...`.

## Gap 12: Resolved REM-### TODO comments left in source
- **Current State**: 7 TODO lines reference work that the same comment marks as resolved:
  - `pkg/engine/vr_openxr_adapters.go:56` — `TODO: uncomment when OpenXR SDK is available` (legitimate; **keep**)
  - `pkg/procgen/quest/generator.go:89`, `quest/templates.go:6`, `quest/types.go:286` — `TODO (REM-144 resolved)` — **delete**
  - `pkg/procgen/story/generator.go:14`, `:24`, `:328` — `TODO (REM-096 resolved)` — **delete**
- **Impact**: Noise; readers waste time investigating "tracked" items that are already done.
- **Remediation**: Single PR deleting the six `REM-09*`/`REM-14*` TODO lines. Verify with `grep -nR 'REM-09\|REM-14' pkg/procgen/`.
- **Effort**: small (~10 min)
- **Dependencies**: none
- **Quick Win**: yes
- **Validation**: `go build ./...`; the grep above returns no hits.

## Gap 13: `make lint` does not catch package-name underscore convention violations
- **Current State**: 10 packages use `snake_case` names (`choice_consequences`, `companion_housing`, `guild_housing`, `guild_vehicle`, `housing_crafting`, `narrative_world`, `political_warfare`, `trade_routes`, `world_events`, plus underscored test helpers). Established convention in this repo, but Go style discourages it.
- **Impact**: Cosmetic; tooling like `revive` complains; cross-repo extraction (per the Procedural Game Suite roadmap) will need to address it eventually.
- **Remediation**: **Defer.** Renaming touches every importer for no behavior gain on a single-maintainer project. Track for the next major version (`v8.x`) where API breakage is acceptable; rename in one PR per package via `gopls rename`.
- **Effort**: medium (when done — ~1 hour total)
- **Dependencies**: major version boundary
- **Quick Win**: no
- **Validation**: when undertaken — `go build ./... && xvfb-run -a go test ./...`.

## Gap 14: 63 `Deprecated:` annotations awaiting next major-version cleanup
- **Current State**: 63 methods carry `// Deprecated:` annotations with replacement guidance (e.g. `pkg/engine/collision.go:233`, `pkg/engine/expression_system.go:184`, `pkg/engine/faction_component.go:137,144`, `pkg/engine/physics/vehicle/collision_response.go:41,48,54,60`, `cmd/client/handlers.go:2393`).
- **Impact**: None active — these are correct API stewardship. Becomes debt only if the deprecation backlog grows unbounded.
- **Remediation**: At the next major version (`v8.x`), bulk-delete deprecated symbols and run `go vet ./... && xvfb-run -a go test ./...` to find any remaining callers. Track the count per release; if it grows past ~100 between majors, tighten the policy.
- **Effort**: medium (when undertaken)
- **Dependencies**: major version boundary
- **Quick Win**: no
- **Validation**: full build + test suite passes after deletions.

---

## Priority Summary

| Priority | Gaps |
|----------|------|
| **P1 (next sprint)** | Gap 1 (behavior-tree dup), Gap 2 (engine god-package; first sub-package), Gap 3 (handlers.go split — Step 1 is a quick win) |
| **P2 (this quarter)** | Gap 4 (stat-mod helper), Gap 5 (`registerSystemBatch`), Gap 6 (parameter structs), Gap 7 (util.go split), Gap 8 (engine race tests in CI), Gap 9 (`staticcheck`) |
| **P3 (opportunistic / next major)** | Gap 10 (audit deps), Gap 11 (`pkg/errors` rename), Gap 12 (REM-### TODO sweep — quick win), Gap 13 (snake_case packages — defer), Gap 14 (`Deprecated:` cleanup at v8) |

Quick wins shippable in a single PR each: **Gap 3 Step 1**, **Gap 5 first batch**, **Gap 6**, **Gap 7**, **Gap 9** (adoption), **Gap 11**, **Gap 12**.

— end of gaps —
