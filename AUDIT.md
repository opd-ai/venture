# MAINTAINABILITY AUDIT — 2026-04-22

> **Note on legacy IDs.** Previous `AUDIT.md` / `GAPS.md` revisions were "implementation gap" audits (Gap 1 = Trade quantity, Gap 2 = Material impact particle wiring, Gap 3 = VR OpenXR adapter, etc.) and are referenced from `cmd/client/handlers.go:115,766,1136`, `.github/copilot-instructions.md:336–339`, and `ROADMAP.md`. Those implementation-completeness items are **out of scope** for this maintainability audit. Their IDs are preserved as `LEGACY-IGAP-1..5` in the False Positives table so cross-references remain traceable.

## Project Maintenance Context

| Dimension | Observation | Source |
|-----------|-------------|--------|
| Purpose | Zero-asset, fully-procedural multiplayer action-RPG (Go + Ebiten) | `README.md`, `.github/copilot-instructions.md` |
| Maturity stage | **Active mature** — 5,364 total commits, tag `v7.0.0` released, 120 commits in last 30 days | `git log` |
| Team size / bus factor | Small core (`user`, `opdai`) + heavy Copilot-agent contributions (`copilot-swe-agent[bot]`, `Copilot`); functionally **single-maintainer-with-agents** | `git shortlog -sn` |
| Change velocity | High; large engine package (`pkg/engine`) and entry points churn constantly | `git log --name-only -n200` |
| Quality infrastructure | CI (`.github/workflows/test.yml`) runs: `go test` (xvfb), `go test -race` on integration, `go vet`, `gofmt -s -l`, `govulncheck`, builds for client/server. `make lint` = `go vet` + custom `scripts/validate-network-types.sh`. **No** `golangci-lint`/`staticcheck`/coverage gate in CI. | `Makefile`, `.github/workflows/test.yml` |
| High-churn files (top 5) | `cmd/client/handlers.go` (25), `cmd/server/main.go` (19), `pkg/engine/system_init.go` (11), `pkg/procgen/story/generator.go` (7), `pkg/network/server.go` (7) | `git log --name-only` |
| Codebase size | 215,987 LOC · 1,266 non-test Go files · 107 packages · 4,083 functions + 12,881 methods · 2,260 structs · 107 interfaces | `go-stats-generator` |

The repository follows the venture-style layout (`cmd/{client,server,mobile}`, `pkg/...`) shared across the 8-repo "Procedural Game Suite". Maintainability expectations are **production / mature**, with a single human reviewer, so refactoring recommendations are weighted toward **high-impact, low-effort, mechanically-verifiable** changes.

## Maintainability Scorecard

| Category | Status | Key Metric |
|----------|--------|------------|
| Complexity | ⚠️ | 89 functions cyclomatic > 10; only **2** > 20 (`InitializeGameSystems`=23, `Tick@1167`=23). Avg complexity 4.0 ✅. Long-function tail: **675** funcs > 50 lines, **104** > 100 lines. |
| Duplication | ⚠️ | 690 clone groups; 30,249 duplicated lines; **duplication ratio 7.05 %**; 56 exact + 634 renamed; largest clone 40 lines. Concentrated in `pkg/engine/behavior_tree_*`, `pkg/engine/system_init.go`, `cmd/client/handlers.go`, and the `*_system.go` "specialization_*"/"weather_*" pattern families. |
| Coupling | ⚠️ | **0 circular dependencies** ✅. Fan-out outliers: `main` (93 deps — expected for entry point), `engine` (64 deps, **god package**: 583 files / 9,534 funcs / 203,240 LOC), `mobile` (19), `audit` (16 deps in only 3 files / 5 funcs — coupling without substance). |
| Documentation | ✅ | Overall doc coverage **90 %** (packages 99 %, functions 97 %, types 89 %, methods 88 %). Quality score 100. |
| Code Freshness | ✅ | `go.mod` declares Go 1.24.5 (current minor is 1.24.x, fully current). No `ioutil` usage in production code (one comment-only mention in `pkg/social/persistence/doc.go`). 7 active `TODO`s — all dated/tracked (REM-### or VR build-tag); 0 `FIXME`; 2 `HACK` annotations (one is a misdetected string literal). 63 `Deprecated:` annotations are **intentional API deprecation notices** with replacement guidance, not rot. |
| Testability | ⚠️ | ECS architecture is interface-first (`pkg/engine/interfaces.go`) — generally testable. Stub providers (`StubInput`, `StubSprite`, `StubImage`) exist. Known race in `TestCreateEntityConcurrentSafety` (`pkg/engine/ecs_test.go:477`) per repo memory. CI runs `-race` only on `pkg/world/housing/...` and `pkg/integration/...`, **not engine-wide** — race coverage gap. |

Legend: ✅ healthy · ⚠️ improvable but not blocking · ❌ urgent debt.

## Complexity Hotspots

Top offenders (overall = weighted score combining cyclomatic, cognitive, nesting, length):

| Function | File:Line | Cyclomatic | Lines | Params | Nest | Risk | Notes |
|----------|-----------|-----------:|------:|-------:|-----:|------|-------|
| `InitializeGameSystems` | `pkg/engine/system_init.go:301` | 23 | 1944 | 2 | 4 | MEDIUM | Linear init of 66 systems; **complexity is justified inline** (`system_init.go:289–297` comment). High-churn (11 modifications). |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:1167` | 23 | 134 | 3 | 3 | HIGH | Behavior-tree node logic; multiple sibling `Tick` methods at cyclomatic 19–20 in same file → repeated pattern. |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:556` | 20 | 98 | 3 | 3 | HIGH | Same family. |
| `main` | `cmd/balance-validator/main.go:67` | 19 | 134 | 0 | 3 | MEDIUM | CLI tool entry-point; intentional flag/dispatch logic. |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:175` | 19 | 104 | 3 | 2 | HIGH | Same family. |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:421` | 16 | 99 | 3 | **5** | HIGH | Deepest nesting in file. |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:41` | 16 | 100 | 3 | 3 | MEDIUM | |
| `applySynergyBonus` | `pkg/engine/dual_class_synergy_system.go:349` | 16 | 60 | 2 | 3 | MEDIUM | Stat-bonus dispatch; consider table-driven. |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:787` | 16 | 78 | 3 | 2 | MEDIUM | |
| `Tick` | `pkg/engine/behavior_tree_advanced_nodes.go:312` | 15 | 77 | 3 | 4 | MEDIUM | |
| `OnDamageTaken` | `pkg/engine/combat_equipment_durability_particle_system.go:93` | 15 | 84 | 2 | 3 | MEDIUM | |
| `renderAccessory` | `pkg/rendering/sprites/equipment_renderer.go:506` | 14 | 69 | **5** | **5** | MEDIUM | High-param + deep-nest combination. |
| `GetCollisionAlignment` | `pkg/engine/collision_precise.go:684` | 14 | 92 | 1 | 4 | LOW | Geometric algorithm; intrinsic. |

### High-parameter functions (signature bloat)

| Function | File:Line | Params |
|----------|-----------|-------:|
| `createDeathCallback` | `cmd/client/util.go:1440` | **13** |
| `spawnProceduralLoot` | `cmd/client/util.go:1336` | 11 |
| `buildFurniture` | `pkg/procgen/furniture/generator.go:83` | 10 |
| `setupUICallbacks` | `cmd/client/handlers.go:2989` | 9 |
| `runGameLoop` | `cmd/server/main.go:894` | 9 |
| `ApplySpellEffect` | `pkg/engine/spell_effect_system.go:498` | 9 |
| `initializeUIIntegration` | `cmd/client/handlers.go:3271` | 8 |
| `executeGameLoop` | `cmd/server/main.go:326` | 8 |
| `runBattlesForClass` | `pkg/balance/combat.go:142` | 8 |
| `addAdvancedComponents` | `pkg/engine/entity_spawning.go:238` | 8 |

> 80 functions exceed the > 5 parameter threshold; the 13- and 11-parameter outliers in `cmd/client/util.go` are the most actionable (introduce a parameter struct).

### Oversized files (>1000 lines, top 12)

| File | Lines | Notes |
|------|------:|-------|
| `cmd/client/handlers.go` | 3880 | High-churn (25 commits). Mixes UI wiring, voice transport, system callbacks. |
| `pkg/engine/spell_casting.go` | 3191 | 158 funcs / 5 types — split candidate. |
| `pkg/rendering/sprites/anatomy_template.go` | 2844 | Mostly data tables (low cyclomatic). |
| `cmd/client/util.go` | 2698 | Generic file name + many unrelated helpers. |
| `pkg/engine/character_creation.go` | 2573 | 108 funcs / 7 types. |
| `pkg/engine/system_init.go` | 2339 | 1944-line `InitializeGameSystems`. |
| `pkg/engine/input_system.go` | 2225 | 161 funcs. |
| `pkg/engine/animation_system.go` | 2216 | |
| `pkg/engine/game.go` | 2133 | |
| `pkg/procgen/book/grammar.go` | 1814 | Grammar table — likely intentional. |
| `pkg/engine/combat_system.go` | 1527 | |
| `pkg/procgen/skills/templates.go` | 1421 | Templates table. |

344 files report as "oversized" by go-stats-generator's burden metric, but most are template/data-heavy rather than complex.

## Findings

### CRITICAL
*(none — there are no circular dependencies, no >30-cyclomatic untested critical-path functions, and no coupling that prevents independent change. The bar set in the task spec is not met.)*

### HIGH
- [x] **Duplicated behavior-tree `Tick` implementations** — `pkg/engine/behavior_tree_advanced_nodes.go` (lines 41/175/312/421/556/787/1034/1167) and `pkg/engine/behavior_tree_tactical_nodes.go` (lines 198/419) — clone group `40L renamed @ behavior_tree_advanced_nodes.go:922 ↔ :1067`, plus 9 sibling `Tick` methods all at cyclomatic 15–23. **Impact:** every behavior-node fix has to be copy-edited across 9+ near-identical methods; high regression risk in AI subsystem. **Remediation:** extract a shared scaffold helper `runMovementTick(entity, blackboard, dt, params, step func(...) NodeStatus) NodeStatus` that owns the common preflight (faction lookup, target resolution, distance/timeout bookkeeping) and delegates only the per-node movement decision. Validate with `go test ./pkg/engine/... -run BehaviorTree -race` and rerun `go-stats-generator analyze . --skip-tests | grep behavior_tree_` to confirm the cyclomatic count drops below 15.
- [x] **`pkg/engine` is a god package** — 583 files / 9,534 functions / 203,240 LOC / 64 dependencies (`go-stats-generator` package analysis). **Impact:** every change forces re-reading huge surface area; bus-factor and onboarding hostile. **Remediation:** carve out cohesive sub-packages incrementally without breaking imports — start with the cleanest seams (`pkg/engine/behavior_tree_*` → `pkg/engine/ai/behavior`, `pkg/engine/specialization_*_system.go` → `pkg/engine/specialization`, `pkg/engine/weather_*_system.go` → `pkg/engine/weather`). Each move is a single PR (move file + update imports + `goimports -w`). Verify with `go build ./... && go test ./pkg/engine/... ./cmd/...`. **✅ First extraction complete** — `pkg/engine/aitypes` (NodeStatus, Blackboard, EntityContext) and `pkg/engine/ai/behavior` (BehaviorNode + 5 composite nodes: Sequence, Selector, Parallel, Inverter, Repeat) extracted; all engine aliases preserved; full build + tests pass.
- [x] **`cmd/client/handlers.go` (3880 lines, 25 recent commits) has self-clones** — clone group `36L renamed @ handlers.go:1730 ↔ :1823 ↔ :1731 ↔ :1824` (7+ instances). High-churn + duplication = highest-risk maintainability hotspot in the repo. **Impact:** every UI/voice/system-wiring change must be repeated; recent `AUDIT.md Gap 1` voice-transport wiring (lines 115, 766, 1136) shows this file already accumulates cross-cutting wiring. **Remediation:** (1) extract the duplicated 28–36-line callback-registration blocks at lines 1730–1864 into a single `registerSystemCallbacks(g, deps)` helper; (2) split file by responsibility — proposed: `handlers_voice.go`, `handlers_ui_callbacks.go`, `handlers_system_init.go` (no behavior change, only file moves). Validate with `go build ./cmd/client/...`, `go vet ./cmd/client/...`, and existing client tests under xvfb.

### MEDIUM
- [x] **Repeated stat-bonus dispatch logic across `*_system.go` "specialization_*" / "weather_*" pattern families** — clone groups: `34L @ specialization_defense_system.go:80 ↔ specialization_evasion_system.go:97`; `34L @ weather_block_chance_system.go:111 ↔ weather_crit_chance_system.go:97`; `39L @ specialization_crit_damage_system.go:76 ↔ specialization_lifesteal_system.go:76`; plus the `applySynergyBonus` (cyc 16) and `removeSynergyBonus` (cyc 14) pair in `pkg/engine/dual_class_synergy_system.go:349,413`. **Impact:** every new specialization or weather modifier requires copy-paste from a sibling file. **Remediation:** introduce a shared `pkg/engine/statmod` helper exporting `Apply(entity *Entity, mod StatModSpec)` / `Remove(...)` and migrate 1 file per PR. Each migration is verifiable: `go test ./pkg/engine/... -run Specialization|Weather|DualClass`.
- [x] **`InitializeGameSystems` is 1944 lines / cyclomatic 23** — `pkg/engine/system_init.go:301`. The file also contains the densest internal duplication in the codebase (clone groups at lines 596, 650, 652, 656, 657, 658, 778–787, 1425, 1432, 1528, 1705–1713 — most are 28–34 lines of repeated init/check patterns). **Impact:** every new system bloats this function; high-churn (11 commits). The function is intentionally linear (justified by inline comment at `:289–297`), so the *length* is acceptable, but the **internal duplication is not**. **Remediation:** extract a `registerSystemBatch(world, []systemRegistration{...})` helper that consumes a slice of `{name string, factory func() System, postInit func(System)}`. Migrate one batch at a time. Validate by ensuring `go test ./pkg/engine/... -run InitializeGameSystems` and runtime startup logs are unchanged.
- [x] **High-parameter outliers in `cmd/client/util.go`** — `createDeathCallback` (13 params, line 1440) and `spawnProceduralLoot` (11 params, line 1336). **Impact:** call-sites are unreadable and order-sensitive; refactors of either require touching every caller. **Remediation:** introduce a `DeathCallbackDeps` / `LootSpawnDeps` parameter struct in the same file, default-initialize, and migrate call sites. Pure mechanical refactor, ~100-line PR each. Validate with `go vet ./cmd/client/... && go build ./cmd/client/...`.
- [x] **`cmd/client/util.go` is a generic dumping ground** — 2698 lines, 85 funcs / 2 types, file-name reported as `generic_name` violation by go-stats-generator. **Impact:** discoverability suffers; new contributors cannot guess where helpers live. **Remediation:** split into `client_loot.go`, `client_callbacks.go`, `client_genre.go` (uses existing `canonicalGenreID` helper noted in repo memory), `client_input.go` based on existing function clusters. No behavior change. Validate with `go build ./cmd/client/...`.
- [x] **Engine package not race-tested in CI** — `.github/workflows/test.yml` runs `-race` only on `pkg/world/housing/...` and `pkg/integration/...`. Per repo memory, `pkg/engine/ecs_test.go:468/477`, `expression_combo_test.go:84`, `guild_vehicle_system_test.go:100`, and `pkg/procgen/audit/determinism_test.go:306` have known races/flakes. **Impact:** new concurrency bugs in the engine (where most of the code lives) cannot be caught at PR time. **Remediation:** stabilize the four named tests in separate PRs (deterministic ordering, sync.Mutex on shared state, or t.Parallel removal), then add `xvfb-run go test -race ./pkg/engine/...` to `test.yml` once green. Track stabilization PRs against this finding.
- [ ] **Audit package coupling-without-substance** — `pkg/procgen/audit` is reported with **16 dependencies, only 3 files, 5 functions** (highest deps-per-file in the repo). **Impact:** the package imports many subsystems for tiny scope, creating an implicit central registry that ripples on any subsystem rename. **Remediation:** move audit-specific helpers into the package they audit (one helper per subsystem package, e.g. `pkg/procgen/quest/audit.go`) and have the top-level audit driver iterate over a registered list. Verify with `go test ./pkg/procgen/audit/...` and dependency count drop via `go-stats-generator`.
- [ ] **`make lint` only runs `go vet` + a custom shell script** — no `staticcheck`, no `golangci-lint`, no `errcheck`. **Impact:** common maintainability issues (unused parameters, error-shadowing, ineffectual assignments) are not caught automatically; new debt accrues silently. **Remediation:** add `staticcheck` to the `lint` target (`go install honnef.co/go/tools/cmd/staticcheck@latest && staticcheck ./...`) and to `.github/workflows/test.yml`. First run will be noisy — file initial findings as separate issues; do **not** block the PR adding the linter on fixing every existing warning. Verify by running locally before merge.

### LOW
- [ ] **10 packages use snake_case names** (Go convention violation reported by go-stats-generator naming analysis): `choice_consequences`, `companion_housing`, `guild_housing`, `guild_vehicle`, `housing_crafting`, `narrative_world`, `political_warfare`, `trade_routes`, `world_events`, plus `errors` (collides with stdlib). **Impact:** mostly cosmetic, but the `errors` collision can confuse readers and tooling (`golang.org/x/tools/go/analysis` import graph). **Remediation:** rename `pkg/errors` → `pkg/venterr` (or similar) in a single PR with `gopls rename`. Defer the underscore renames — they touch every importer for no behavior gain on a single-maintainer project.
- [x] **Stale documentation freshness in REM-### TODOs** — 7 `TODO` comments in `pkg/procgen/quest/{generator,templates,types}.go` and `pkg/procgen/story/generator.go` reference completed `REM-096` / `REM-144` work but were left in place. **Impact:** noise; readers waste time investigating "resolved" markers. **Remediation:** delete the seven TODO lines (one PR). Verify with `grep -n 'REM-09\|REM-14' pkg/procgen/...` returning empty.
- [x] **VR OpenXR adapter `TODO: uncomment when OpenXR SDK is available`** — `pkg/engine/vr_openxr_adapters.go:56`. This is **tracked legacy implementation gap** (was `LEGACY-IGAP-3` in prior `GAPS.md`); per repo memory `pkg/vr/doc.go:66-73` documents the intended build-tagged adapter pattern. **Impact:** none active — stub is wired by default. **Remediation:** keep TODO; classify as tracked/non-urgent. Re-evaluate when OpenXR SDK becomes a project priority.
- [x] **`HACK` annotation in `pkg/security/audit.go:831` (`prevention`)** is the only real hack annotation; the second match (`pkg/procgen/skills/templates.go:21` `%s systems remotely`) is a misdetected string literal in a template. **Impact:** trivial. **Remediation:** convert `audit.go:831` `// HACK:` to `// NOTE:` if the workaround is intentional and permanent, or file an issue if it is not.
- [x] **63 `Deprecated:` annotations** on legacy methods (e.g. `pkg/engine/collision.go:233`, `pkg/engine/expression_system.go:184`, `pkg/engine/faction_component.go:137,144`, `pkg/engine/physics/vehicle/collision_response.go:41,48,54,60`). **All carry replacement guidance.** **Impact:** intentional API stewardship, not debt. **Remediation:** keep until next major version (`v8.x`); at that boundary, delete in one PR and run `go vet ./...` + full test suite to confirm no remaining callers. (Filed for completeness — not urgent.)
- [ ] **Misplaced functions/methods (heuristic)** — go-stats-generator reports 2,396 misplaced funcs / 1,314 misplaced methods (e.g. `BalanceConfig` methods split between `pkg/balance/types.go` and `pkg/procgen/magic/balance/...`). **Impact:** marginal — most are intentional layering. **Remediation:** review only the top 10 affinity-gain candidates (output above) opportunistically when touching adjacent code. Do not run a sweeping move.

## Technical Debt Inventory

| # | Debt Item | Category | Effort | Impact | Priority |
|---|-----------|----------|--------|--------|----------|
| 1 | Behavior-tree `Tick` duplication (9+ near-clones, all cyc 15–23) | duplication + complexity | M | Cuts AI-subsystem regression risk; lowers all listed `Tick` to <15 | **P1** |
| 2 | `pkg/engine` god-package decomposition (incremental sub-package extraction) | coupling | L | Long-term; halves the largest package and unlocks parallel work | **P1** |
| 3 | `cmd/client/handlers.go` extract callback-registration helper + file split | duplication + complexity | M | Reduces highest-churn-file risk | **P1** |
| 4 | Specialization/weather/dual-class stat-mod helper | duplication | M | One-time helper amortizes across ~10 systems | P2 |
| 5 | `InitializeGameSystems` `registerSystemBatch` helper | duplication | S–M | Removes ~12 internal clone groups | P2 |
| 6 | `cmd/client/util.go` parameter-struct refactor (13- and 11-param funcs) | complexity (signature) | S | Improves call-site readability; mechanical | P2 |
| 7 | `cmd/client/util.go` file split | coupling/discoverability | S | New-contributor friction | P2 |
| 8 | Add `staticcheck` to `make lint` and CI | quality infra | S | Catches future debt automatically | P2 |
| 9 | Race-test the engine in CI (after stabilizing 4 known flaky tests) | quality infra | M | Prevents regressions in the largest package | P2 |
| 10 | Collapse `pkg/procgen/audit` deps by relocating helpers | coupling | S–M | Removes top deps-per-file outlier | P3 |
| 11 | Rename `pkg/errors` to avoid stdlib collision | naming/freshness | S | Clarity | P3 |
| 12 | Delete resolved `REM-###` TODO comments | docs freshness | S (1 PR) | Quick win | P3 |
| 13 | Sweep `Deprecated:` removals at next major version | freshness | M | Aligned with version cadence | P3 |
| 14 | `vr_openxr_adapters.go` TODO | tracked gap | L (depends on SDK) | None until VR is prioritized | P3 |

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|-----------------|
| `InitializeGameSystems` 1944 lines is critical debt | The function is intentionally linear (66 sequential `world.AddSystem` calls) and the design rationale is documented inline at `pkg/engine/system_init.go:289–297`. Behavior-preserving extraction is still valuable (Item 5) but the *length* is not urgent debt. |
| 344 "oversized files" reported by burden metric | Most are data-table files (`anatomy_template.go`, `creature_frame_offsets.go`, `book/grammar.go`, `skills/templates.go`) intentional to the zero-asset philosophy — moving data into more files would not improve maintainability. |
| `main` package fan-out 93 dependencies | Entry point — by definition imports everything it wires. Not a coupling defect. |
| 7.05 % duplication ratio is "high" | Below the typical "concerning" threshold (10 %) and dominated by `renamed` clones in domain-isomorphic systems (specialization, weather). The actionable subset is captured in Items 1, 4, 5; the rest is acceptable for a single-maintainer project. |
| Underscore-named packages (`companion_housing`, `world_events`, etc.) | Project-wide convention used consistently (10 packages); changing them touches every importer for cosmetic gain. Filed only as LOW. |
| 89 functions with cyclomatic > 10 | Average complexity 4.0 across 17K functions is excellent. The 89 outliers are predominantly behavior-tree `Tick`, particle systems, and protocol handlers — domains where the problem space requires it. Only those above 15 with deep nesting or parameter bloat are listed in Findings. |
| 63 `Deprecated:` annotations | All carry replacement guidance with new helpers. This is correct API stewardship, not rot. |
| `LEGACY-IGAP-1` Trade-quantity contract | Implementation completeness gap, not maintainability debt. Tracked separately in `.github/copilot-instructions.md:337` and the prior `GAPS.md` Gap 1. |
| `LEGACY-IGAP-2` Material impact particle wiring | Implementation completeness gap; not maintainability. |
| `LEGACY-IGAP-3` VR OpenXR adapter | Implementation completeness gap (kept as LOW reminder above). |
| `LEGACY-IGAP-4` Unused exported sentinel errors | Surface API debt, low maintainability impact; tracked in prior `GAPS.md`. |
| `LEGACY-IGAP-5` Stale REM-### TODOs | Captured as LOW Item 12. |
| `TestCreateEntityConcurrentSafety` race | Pre-existing, tracked in repo memory; impacts test-infra (Item 9), not maintainability of production code. |
| Naming "stuttering" warnings (e.g. `FeatureRegistry` in `pkg/audit/features`) | Conventional Go naming for nested registries; readability cost would exceed gain. |

## Validation Commands

Every remediation above can be verified with the standard repo toolchain:

```bash
# Build / vet
go build ./...
go vet ./...

# Tests (display required)
xvfb-run -a go test ./pkg/...

# Race subset (CI parity)
xvfb-run -a go test -race ./pkg/world/housing/... ./pkg/integration/...

# Re-run metrics after each refactor
go-stats-generator analyze . --skip-tests --format json \
  --sections functions,duplication,packages > /tmp/post.json
go-stats-generator analyze . --skip-tests | head -60
```

— end of audit —
