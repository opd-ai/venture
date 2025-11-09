# AUTONOMOUS CODEBASE MAINTENANCE

## CONTEXT
Autonomous agent for Venture (Go 1.24+, Ebiten 2.9, ECS, procedural action-RPG). 6 phases: (1) fix build/tests, (2) **complete V6.0 (Phase 4-6)**, (3) align docs, (4) roadmap, (5) refactor, (6) enhance. Use xvfb for tests. Ref: ROADMAP_V6.md, AUDIT.md files.

**V6.0 Features (Phase 4-6):**
P4: Post office, mail system, courier NPCs, async delivery | P5: Politics, factions, alliances, trade network, dynamic pricing | P6: Territory control, border zones, bounty board, server rankings

**ALL** in pkg/engine/ (mail, politics, bounty), pkg/network/federation/market.go, pkg/world/territory.go, cmd/server/main.go.

## CONSTRAINTS
- Seed RNG only (no time.Now()), ECS: entities=IDs, components=data, systems=logic
- ≥65% coverage (exclude Ebiten), table tests, verify determinism
- Targets: 60 FPS, <500MB, <2s gen | gofmt -w -s, Go stdlib, stubs for tests

## PHASE 1: BUILD/TEST
Install deps, `go build ./...` and `go test ./...`. Fix: (1) compile, (2) imports, (3) tests, (4) races. Skip if pass.

## PHASE 2: V6.0 COMPLETION
**Audit:** P4-6 components/systems in pkg/engine/, pkg/network/federation/, pkg/world/, cmd/server/main.go.
**Implement:** Missing items (mail, politics, bounty, territory), test ≥65%, doc.
**Remove:** Deprecated, legacy, pre-V6.0, shims.
**Validate:** Determinism (per-server), targets, tests, run server federation.
**Skip:** Only if ALL done.

## PHASE 3: DOCS
Compare README.md vs code. Classify gaps. Fix top 3 with tests. Skip if aligned.

## PHASE 4: ROADMAP
**Priority:** (1) V6.0, (2) ROADMAP_V*.md, (3), (4) TODOs.
ECS changes, loop integration, test (≥65%, determinism), verify. Skip if done.
Carry out `EXECUTE.md` instructions against current ROADMAP_V*.md file.

## PHASE 5: QUALITY
Scan complexity >15, nesting >4, length >200, violations, coverage <65%, TODOs, duplication. Pick ONE. Refactor, validate. Skip if good.

## PHASE 6: ENHANCE
Pick ONE: graphics/gameplay/perf/QoL. High impact, low risk, <100 LOC. ECS, determinism, tests, validate, doc. Skip if optimal.

## SUCCESS
- Build/test/race pass, ≥65% coverage, gofmt
- **P4-6 features complete, federation operational**
- No deprecated, ECS OK, determinism (per-server), targets (60 FPS, <500MB per server, <2s)
- Docs current, quality up, enhancement done

## V6.0 CHECKLIST (Phase 4-6 Target)
**Phase 4:** MailComponent, MailMessage, MailSystem, post office buildings, courier NPCs, delivery tracking in pkg/engine/mail_component.go
**Phase 5:** ServerFaction, PoliticalEvent, PoliticalSystem, FederatedMarket, dynamic pricing, merchant caravans in pkg/engine/politics_component.go, pkg/network/federation/market.go
**Phase 6:** BorderZone, ControlPoint, TerritorySystem, BountyContract, BountySystem, server leaderboards in pkg/world/territory.go, pkg/engine/bounty_component.go

**Integration:** cmd/server/main.go registers P4-6 systems, mail delivery functional, political events trigger, trade network operational, territory control active, bounty board accessible.

Execute autonomously. Report comprehensive results for all 6 phases.

