# AUTONOMOUS CODEBASE MAINTENANCE

## CONTEXT
Autonomous agent for Venture (Go 1.24+, Ebiten 2.9, ECS, procedural action-RPG). 6 phases: (1) fix build/tests, (2) **complete current roadmap**, (3) align docs, (4) next roadmap, (5) refactor, (6) enhance. Use xvfb for tests. Ref: ROADMAP_V*.md, AUDIT.md files.

**V3.0 Features (Complete):**
P15: Enhanced sprites, anatomical detail, genre variations | P16: Advanced tiles, textures, transitions | P17: Lighting, shadows, bloom, ambient | P18: Particles, weather, fluid simulation | P19: UI polish, dynamic palettes | P20: Environmental detail, post-processing

**ALL** in pkg/rendering/, pkg/procgen/environment/, cmd/client/main.go.

## CONSTRAINTS
- Seed RNG only (no time.Now()), ECS: entities=IDs, components=data, systems=logic
- ≥65% coverage (exclude Ebiten), table tests, verify determinism
- Targets: 60 FPS, <500MB, <2s gen | gofmt -w -s, Go stdlib, stubs for tests

## PHASE 1: BUILD/TEST
Install deps, `go build ./...` and `go test ./...`. Fix: (1) compile, (2) imports, (3) tests, (4) races. Skip if pass.

## PHASE 2: ROADMAP COMPLETION
**Audit:** Current ROADMAP_V*.md for incomplete features in pkg/, cmd/client/main.go.
**Implement:** Missing items (ECS), integrate loop, test ≥65%, doc.
**Remove:** Deprecated, legacy, pre-current version, shims.
**Validate:** Determinism, targets (60 FPS, <500MB, <2s gen), tests, run client.
**Skip:** Only if ALL done.

## PHASE 3: DOCS
Compare README.md vs code. Classify gaps. Fix top 3 with tests. Skip if aligned.

## PHASE 4: ROADMAP
**Priority:** (1) Current ROADMAP_V*.md (V3→V4→V5→V6), (2) TODOs, (3) EXECUTE.md.
ECS changes, loop integration, test (≥65%, determinism), verify. Skip if done.
Pick next incomplete roadmap: V3 (complete) → V4 → V5 → V6.

## PHASE 5: QUALITY
Scan complexity >15, nesting >4, length >200, violations, coverage <65%, TODOs, duplication. Pick ONE. Refactor, validate. Skip if good.

## PHASE 6: ENHANCE
Pick ONE: graphics/gameplay/perf/QoL. High impact, low risk, <100 LOC. ECS, determinism, tests, validate, doc. Skip if optimal.

## SUCCESS
- Build/test/race pass, ≥65% coverage, gofmt
- **Current roadmap complete, all systems integrated**
- No deprecated, ECS OK, determinism, targets (60 FPS, <500MB, <2s)
- Docs current, quality up, enhancement done

## V3.0 CHECKLIST (Current Status - All Complete ✅)
**Phase 15:** Enhanced sprite templates, anatomical detail, facial features, genre variations - pkg/rendering/sprites/
**Phase 16:** Advanced tile textures, smooth transitions, multi-layer depth - pkg/rendering/tiles/
**Phase 17:** Soft shadows, colored lighting, bloom effects, ambient occlusion - pkg/rendering/lighting/
**Phase 18:** Weather systems (rain/snow/fog), fluid simulation, particles - pkg/rendering/particles/, pkg/procgen/environment/
**Phase 19:** Dynamic UI palettes, visual hierarchy, procedural decorations - pkg/rendering/ui/
**Phase 20:** Parallax backgrounds, time-of-day, post-processing, environmental polish - pkg/rendering/

**Next Versions:** V4.0 (Phases 21-30: Gameplay expansion), V5.0 (Phases 31-36: Social systems), V6.0 (Phases 37+: Server federation)

Execute autonomously. Report comprehensive results for all 6 phases.

