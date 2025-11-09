# AUTONOMOUS CODEBASE MAINTENANCE

## CONTEXT
Autonomous agent for Venture (Go 1.24+, Ebiten 2.9, ECS, procedural action-RPG). 6 phases: (1) fix build/tests, (2) **complete V2.0 (Phase 10-14)**, (3) align docs, (4) roadmap, (5) refactor, (6) enhance. Use xvfb for tests. Ref: ROADMAP_V2.md, GAPS.md.

**V2.0 Features:**
P10: Rotation, Aim, Projectiles, ScreenShake | P11: Multi-layer, Diagonals, Puzzles, Destructibles, Carry, Context, Hazards | P12: L-System, Narrative, Adaptive music | P13: Behavior trees, Squads, Factions | P14: Shadows, Animation LOD, Enhanced particles/audio

**ALL** in pkg/engine/, cmd/client/main.go loop, entities spawn with V2.0.

## CONSTRAINTS
- Seed RNG only (no time.Now()), ECS: entities=IDs, components=data, systems=logic
- ≥65% coverage (exclude Ebiten), table tests, verify determinism
- Targets: 60 FPS, <500MB, <2s gen | gofmt -w -s, Go stdlib, stubs for tests

## PHASE 1: BUILD/TEST
Install deps, `go build ./...` and `go test ./...`. Fix: (1) compile, (2) imports, (3) tests, (4) races. Skip if pass.

## PHASE 2: V2.0 COMPLETION
**Audit:** All P10-14 components/systems in pkg/engine/, cmd/client/main.go, entity spawning.
**Implement:** Missing items (ECS), integrate loop, spawn logic, test ≥65%, doc.
**Remove:** Deprecated, legacy, pre-V2.0, shims.
**Validate:** Determinism, targets, tests, run client.
**Skip:** Only if ALL done.

## PHASE 3: DOCS
Compare README.md vs code. Classify gaps. Fix top 3 with tests. Skip if aligned.

## PHASE 4: ROADMAP
**Priority:** (1) V2.0, (2) ROADMAP_V2.md, (3) GAPS.md, (4) TODOs.
ECS changes, loop integration, test (≥65%, determinism), verify. Skip if done.
Carry out `EXECUTE.md` instructions

## PHASE 5: QUALITY
Scan complexity >15, nesting >4, length >200, violations, coverage <65%, TODOs, duplication. Pick ONE. Refactor, validate. Skip if good.

## PHASE 6: ENHANCE
Pick ONE: graphics/gameplay/perf/QoL. High impact, low risk, <100 LOC. ECS, determinism, tests, validate, doc. Skip if optimal.

## SUCCESS
- Build/test/race pass, ≥65% coverage, gofmt
- **ALL P10-14 in loop, V2.0 entity spawning**
- No deprecated, ECS OK, determinism, targets (60 FPS, <500MB, <2s)
- Docs current, quality up, enhancement done

## V2.0 CHECKLIST (Before Phase 4/completion)
**Phase 10:** RotationComponent, AimComponent, ProjectileComponent, ScreenShakeComponent + systems in pkg/engine/, registered in cmd/client/main.go
**Phase 11:** LayerComponent, layer-aware collision, diagonal walls, PuzzleComponent/System, DestructibleComponent/System, CarryComponent/System, ContextActionComponent/System, HazardComponent/System
**Phase 12:** pkg/procgen/lsys/, NarrativeComponent, pkg/procgen/narrative/, NarrativeSystem, pkg/audio/music/adaptive.go, motif.go
**Phase 13:** BehaviorTreeComponent, pkg/engine/ai/behavior_tree.go, BehaviorTreeSystem, archetypes, SquadComponent/System, FactionComponent/System, pkg/procgen/faction/
**Phase 14:** ShadowComponent, pkg/rendering/lighting/shadows.go, AnimationComponent w/LOD, pkg/rendering/sprites/animation.go, enhanced particles/audio

**Integration:** cmd/client/main.go registers all systems, entities spawn with V2.0 components, world uses L-System layouts, combat uses projectiles/shake, audio plays adaptive music, rendering shows shadows/LOD/particles.

Execute autonomously. Report comprehensive results for all 6 phases.

