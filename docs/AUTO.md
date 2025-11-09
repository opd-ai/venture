# AUTONOMOUS CODEBASE MAINTENANCE

## CONTEXT
Autonomous agent for Venture (Go 1.24+, Ebiten 2.9, ECS, procedural action-RPG). 6 phases: (1) fix build/tests, (2) **complete V4.0, V5.0, V6.0 → 6.0 readiness**, (3) align docs, (4) roadmap, (5) refactor, (6) enhance. Use xvfb for tests. Ref: ROADMAP_V4.md, ROADMAP_V5.md, ROADMAP_V6.md, AUDIT.md files.

**Target: 6.0 Readiness (V4→V5→V6):**
V4 (P21-30): Vehicles, pets, books, expanded magic, classes, expressions, mini-games, reputation, adaptive music, storytelling | V5 (6 phases): Chat, NPC dialog, image sharing, item trading, social systems | V6 (P31-36): Persistent worlds, federation, cross-server travel, post office, politics/trade, territory control

**ALL** in pkg/engine/, pkg/procgen/, pkg/network/, pkg/world/, pkg/audio/, cmd/client/, cmd/server/.

## CONSTRAINTS
- Seed RNG only (no time.Now()), ECS: entities=IDs, components=data, systems=logic
- ≥65% coverage (exclude Ebiten), table tests, verify determinism
- Targets: 60 FPS, <500MB, <2s gen | gofmt -w -s, Go stdlib, stubs for tests

## PHASE 1: BUILD/TEST
Install deps, `go build ./...` and `go test ./...`. Fix: (1) compile, (2) imports, (3) tests, (4) races. Skip if pass.

## PHASE 2: V4-V6 COMPLETION
**Audit:** V4 (P21-30), V5 (social), V6 (P31-36) components/systems in pkg/, cmd/.
**Implement:** Missing features (vehicles, pets, chat, federation, etc.), test ≥65%, doc.
**Remove:** Deprecated, legacy, pre-V4/V5/V6, shims.
**Validate:** Determinism, targets (60 FPS, <500MB, <2s gen), tests, run client/server.
**Skip:** Only if ALL V4+V5+V6 done.

## PHASE 3: DOCS
Compare README.md vs code. Classify gaps. Fix top 3 with tests. Skip if aligned.

## PHASE 4: ROADMAP
**Priority:** (1) ROADMAP_V4.md → V5.md → V6.md (sequential), (2) TODOs, (3) EXECUTE.md.
ECS changes, loop integration, test (≥65%, determinism), verify. Skip if done.
Complete V4 (Phases 21-30), then V5 (social), then V6 (Phases 31-36) for 6.0 readiness.

## PHASE 5: QUALITY
Scan complexity >15, nesting >4, length >200, violations, coverage <65%, TODOs, duplication. Pick ONE. Refactor, validate. Skip if good.

## PHASE 6: ENHANCE
Pick ONE: graphics/gameplay/perf/QoL. High impact, low risk, <100 LOC. ECS, determinism, tests, validate, doc. Skip if optimal.

## SUCCESS
- Build/test/race pass, ≥65% coverage, gofmt
- **V4+V5+V6 complete, 6.0 ready**
- No deprecated, ECS OK, determinism, targets (60 FPS, <500MB, <2s)
- Docs current, quality up, enhancement done

## 6.0 READINESS CHECKLIST (V4→V5→V6)
**V4.0 (P21-30):** VehicleSystem, PetSystem, BookSystem, ExpandedMagicSystem, CharacterClassSystem, ExpressionSystem, MiniGameSystem, ReputationSystem, AdaptiveSoundtrackSystem, EnvironmentalStorytellingSystem in pkg/engine/, pkg/procgen/, pkg/audio/
**V5.0 (Social):** ChatComponent, DialogSystem, ImageSharingSystem, ItemTradingSystem, E2E encryption, social UI in pkg/engine/social_components.go, pkg/network/
**V6.0 (P31-36):** WorldPersistence, FederationProtocol, PortalSystem, MailSystem, PoliticalSystem, TerritorySystem in pkg/world/, pkg/network/federation/, pkg/engine/

**Integration:** All V4-V6 systems registered in cmd/client/main.go and cmd/server/main.go, features accessible in-game, multiplayer functional, tests pass, docs complete.

Execute autonomously. Report comprehensive results for all 6 phases.

