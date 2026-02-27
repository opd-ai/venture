# AUTONOMOUS CODEBASE MAINTENANCE

## CONTEXT
Autonomous agent for Venture (Go 1.24+, Ebiten 2.9, ECS, procedural action-RPG). 4 phases: (1) **complete current roadmap phase**, (2) align docs, (3) roadmap, (4) refactor. Use xvfb for tests. Ref: ROADMAP.md, CHANGELOG.md, AUDIT.md files. Work sequentially where possible. Pick a reasonable workload if the roadmap is too long.

**Target: 8.0 Readiness (V4→V5→V6→V7→V8):**
V4 (P21-30): Vehicles, pets, books, expanded magic, classes, expressions, mini-games, reputation, adaptive music, storytelling | V5 (6 phases): Chat, NPC dialog, image sharing, item trading, social systems | V6 (P31-36): Persistent worlds, federation, cross-server travel, post office, politics/trade, territory control | V7 (P37-42): Advanced AI, dynamic events, procedural quests 2.0, economy simulation, seasonal content, world history | V8 (P43-48): Modding API, user content, workshop integration, scripting system, plugin architecture, community features

**ALL** in pkg/engine/, pkg/procgen/, pkg/network/, pkg/world/, pkg/audio/, pkg/modding/, cmd/client/, cmd/server/.

## CONSTRAINTS
- Seed RNG only (no time.Now()), ECS: entities=IDs, components=data, systems=logic
- ≥40% coverage (≥30% for X11/Wayland/Ebiten-dependent packages), table tests, verify determinism
- Targets: 60 FPS, <500MB, <2s gen | gofmt -w -s, Go stdlib, stubs for tests

## PHASE 1: V4-V8 COMPLETION
**Audit:** V4 (P21-30), V5 (social), V6 (P31-36), V7 (P37-42), V8 (P43-48) components/systems in pkg/, cmd/.
**Implement:** Missing features (vehicles, pets, chat, federation, AI, modding, etc.), test ≥40%, doc.
**Remove:** Deprecated, legacy, pre-V4/V5/V6/V7/V8, shims.
**Validate:** Determinism, targets (60 FPS, <500MB, <2s gen), tests, run client/server.
**Skip:** Only if ALL V4+V5+V6+V7+V8 done.

## PHASE 2: DOCS
Compare README.md vs code. Classify gaps. Fix top 3 with tests. Skip if aligned.

## PHASE 3: ROADMAP
**Priority:** (1) ROADMAP.md, (2) TODOs, (3) EXECUTE.md.
ECS changes, loop integration, test (≥40%, determinism), verify. Skip if done.
Complete remaining V10 phases for production readiness.

## PHASE 4: QUALITY
Scan complexity >15, nesting >4, length >200, violations, coverage <40%, TODOs, duplication. Pick ONE. Refactor, validate. Skip if good.

## SUCCESS
- Build/test/race pass, ≥40% coverage (≥30% for X11/Wayland/Ebiten-dependent packages), gofmt
- **V4+V5+V6+V7+V8 complete, 8.0 ready**
- No deprecated, ECS OK, determinism, targets (60 FPS, <500MB, <2s)
- Docs current, quality up

## 8.0 READINESS CHECKLIST (V4→V5→V6→V7→V8)
**V4.0 (P21-30):** VehicleSystem, PetSystem, BookSystem, ExpandedMagicSystem, CharacterClassSystem, ExpressionSystem, MiniGameSystem, ReputationSystem, AdaptiveSoundtrackSystem, EnvironmentalStorytellingSystem in pkg/engine/, pkg/procgen/, pkg/audio/
**V5.0 (Social):** ChatComponent, DialogSystem, ImageSharingSystem, ItemTradingSystem, E2E encryption, social UI in pkg/engine/social_components.go, pkg/network/
**V6.0 (P31-36):** WorldPersistence, FederationProtocol, PortalSystem, MailSystem, PoliticalSystem, TerritorySystem in pkg/world/, pkg/network/federation/, pkg/engine/
**V7.0 (P37-42):** AdvancedAISystem, DynamicEventSystem, ProceduralQuests2System, EconomySimulationSystem, SeasonalContentSystem, WorldHistorySystem in pkg/engine/ai/, pkg/procgen/events/, pkg/world/economy/
**V8.0 (P43-48):** ModdingAPI, UserContentSystem, WorkshopIntegration, ScriptingSystem, PluginArchitecture, CommunityFeaturesSystem in pkg/modding/, pkg/scripting/, pkg/workshop/

**Integration:** All V4-V8 systems registered in cmd/client/main.go and cmd/server/main.go, features accessible in-game, multiplayer functional, tests pass, docs complete.

Execute autonomously. Report comprehensive results for all 4 phases.
