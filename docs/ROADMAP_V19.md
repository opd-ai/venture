# Development Roadmap - Version 19.0: Priority 1 Package Integration

## Current Status

**Status:** IN PROGRESS - 0% (0/4 phases done)  
**Prerequisites:** V18.0 Complete (Gathering & Collection Systems)  
**Started:** December 16, 2025  
**Focus:** Integrate Priority 1 dormant packages into client/server

## Overview

**Mission:** Integrate all Priority 1 dormant packages that are complete but not imported into the client/server entry points. These packages are fully implemented with tests but are not being used.

**Packages to Integrate:**
1. `pkg/procgen/entity` - Entity spawning pipeline
2. `pkg/procgen/dialog` - Procedural dialog generation
3. `pkg/procgen/legendary` - Legendary item generation
4. `pkg/world/economy` - Dynamic economy simulation
5. `pkg/integration/choice_consequences` - Choice tracking and consequences
6. `pkg/integration/guild_vehicle` - Guild vehicle fleet combat
7. `pkg/integration/world_events` - World-responsive events

## Phase Summary

### Phase 99: Entity & Dialog Generation Integration
**Status:** ⏳ In Progress

Integrate procedural entity and dialog generators into client.

**Deliverables:**
- [ ] Import `pkg/procgen/entity` in `cmd/client/handlers.go`
- [ ] Create EntityGenerator during game initialization
- [ ] Replace manual enemy creation with EntityGenerator
- [ ] Import `pkg/procgen/dialog` in `cmd/client/handlers.go`
- [ ] Connect dialog generator to NPC dialog system
- [ ] Tests pass with new integrations

**Acceptance Criteria:**
- [ ] Entity spawning uses procedural generator
- [ ] NPC dialogs are procedurally generated
- [ ] Zero test regressions
- [ ] Test coverage ≥65%

### Phase 100: Legendary & Economy Integration
**Status:** ⏳ Pending

Integrate legendary items and dynamic economy.

**Deliverables:**
- [ ] Import `pkg/procgen/legendary` in `cmd/client/handlers.go`
- [ ] Connect legendary generator to quest reward system
- [ ] Import `pkg/world/economy` in `cmd/client/handlers.go`
- [ ] Create economy system and register with World
- [ ] Connect economy to merchant pricing

**Acceptance Criteria:**
- [ ] Legendary items generate for boss kills/quest rewards
- [ ] Economy affects shop prices dynamically
- [ ] Zero test regressions
- [ ] Test coverage ≥65%

### Phase 101: Integration Package Activation
**Status:** ⏳ Pending

Activate choice consequences, guild vehicle, and world events systems.

**Deliverables:**
- [ ] Import `pkg/integration/choice_consequences` 
- [ ] Import `pkg/integration/guild_vehicle`
- [ ] Import `pkg/integration/world_events`
- [ ] Create and register all three systems with World
- [ ] Connect to relevant game systems

**Acceptance Criteria:**
- [ ] Choice consequences track player decisions
- [ ] Guild vehicles provide formation bonuses
- [ ] World events trigger based on player actions
- [ ] Zero test regressions

### Phase 102: Validation & Documentation
**Status:** ⏳ Pending

Final validation and documentation updates.

**Deliverables:**
- [ ] All tests pass
- [ ] Update INTEGRATION_AUDIT.md to mark packages as active
- [ ] Verify 60 FPS performance maintained
- [ ] Update this roadmap to complete

**Acceptance Criteria:**
- [ ] `go test -race ./...` passes
- [ ] No performance regression
- [ ] Documentation updated

---

## Quality Gates

- Zero regressions from V18.0
- Test coverage ≥65% maintained
- Performance: 60 FPS maintained
- All systems follow ECS patterns

---

**Document Status:** In Progress  
**Last Updated:** December 2025  
**Version:** 19.0.0 Development
