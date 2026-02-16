# Venture Package Audit Tracker

This file tracks the audit status of all Go sub-packages in the Venture project.

## Audit Status Legend
- [ ] Not audited
- [x] Audited

## Core Packages

### Engine
- [ ] `pkg/engine/` — Core ECS (400+ files, needs scoped sub-audits)

### Procedural Generation (`pkg/procgen/`)
- [x] `pkg/procgen/building/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/dialog/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/entity/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/faction/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low)
- [x] `pkg/procgen/furniture/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/item/AUDIT_2026-02-13_COMPREHENSIVE.md` — Complete — Multiple issues documented
- [x] `pkg/procgen/magic/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/quest/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/skills/AUDIT.md` — Complete — 0 issues
- [x] `pkg/procgen/terrain/AUDIT.md` — Complete — 0 issues
- [ ] `pkg/procgen/book/` — Not audited
- [ ] `pkg/procgen/class/` — Not audited
- [x] `pkg/procgen/companion/AUDIT.md` — Complete — 1 issue (0 high, 0 med, 1 low)
- [ ] `pkg/procgen/environment/` — Not audited
- [x] `pkg/procgen/genre/AUDIT.md` — Complete — 1 issue (0 high, 0 med, 1 low)
- [x] `pkg/procgen/legendary/AUDIT.md` — Complete — 0 issues
- [ ] `pkg/procgen/minigame/` — Not audited
- [ ] `pkg/procgen/narrative/` — Not audited
- [ ] `pkg/procgen/puzzle/` — Not audited
- [ ] `pkg/procgen/recipe/` — Not audited
- [ ] `pkg/procgen/station/` — Not audited
- [ ] `pkg/procgen/story/` — Not audited
- [ ] `pkg/procgen/vehicle/` — Not audited

### Network (`pkg/network/`)
- [x] `pkg/network/AUDIT_COMPLETE.md` — Complete — Multiple issues documented
- [ ] `pkg/network/chat/` — Not audited
- [x] `pkg/network/federation/AUDIT.md` — Needs Work — 7 issues (3 high, 2 med, 2 low)
- [x] `pkg/network/resilience/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low)
- [ ] `pkg/network/trade/` — Not audited

### Rendering (`pkg/rendering/`)
- [ ] `pkg/rendering/sprites/` — Not audited
- [ ] `pkg/rendering/animation/` — Not audited
- [ ] `pkg/rendering/tiles/` — Not audited
- [ ] `pkg/rendering/lighting/` — Not audited
- [ ] `pkg/rendering/postprocess/` — Not audited
- [ ] `pkg/rendering/particles/` — Not audited
- [ ] `pkg/rendering/ui/` — Not audited
- [ ] `pkg/rendering/palette/` — Not audited
- [ ] `pkg/rendering/cache/` — Not audited
- [ ] `pkg/rendering/pool/` — Not audited

### World (`pkg/world/`)
- [ ] `pkg/world/` — Not audited
- [x] `pkg/world/housing/AUDIT.md` — Complete — 2 issues (0 high, 0 med, 2 low)
- [ ] `pkg/world/economy/` — Not audited
- [ ] `pkg/world/territory/` — Not audited
- [ ] `pkg/world/raids/` — Not audited

### Audio (`pkg/audio/`)
- [x] `pkg/audio/music/AUDIT.md` — Complete — 2 issues (0 high, 0 med, 2 low)
- [ ] `pkg/audio/sfx/` — Not audited
- [ ] `pkg/audio/synthesis/` — Not audited

### Integration (`pkg/integration/`)
- [ ] `pkg/integration/companion_housing/` — Not audited
- [ ] `pkg/integration/guild_housing/` — Not audited
- [ ] `pkg/integration/guild_vehicle/` — Not audited
- [ ] `pkg/integration/housing_crafting/` — Not audited
- [ ] `pkg/integration/choice_consequences/` — Not audited
- [ ] `pkg/integration/narrative_world/` — Not audited
- [ ] `pkg/integration/political_warfare/` — Not audited
- [ ] `pkg/integration/trade_routes/` — Not audited
- [ ] `pkg/integration/world_events/` — Not audited

### Supporting Packages
- [x] `pkg/saveload/AUDIT.md` — Complete — Status unknown
- [ ] `pkg/combat/` — Not audited
- [ ] `pkg/config/` — Not audited
- [ ] `pkg/validation/` — Not audited
- [ ] `pkg/errors/` — Not audited
- [ ] `pkg/logging/` — Not audited
- [ ] `pkg/recovery/` — Not audited
- [ ] `pkg/stability/` — Not audited
- [ ] `pkg/observability/` — Not audited
- [ ] `pkg/security/` — Not audited
- [ ] `pkg/version/` — Not audited
- [ ] `pkg/migration/` — Not audited
- [x] `pkg/modding/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low)
- [ ] `pkg/narrative/` — Not audited
- [ ] `pkg/ux/` — Not audited
- [ ] `pkg/balance/` — Not audited
- [ ] `pkg/class/` — Not audited
- [ ] `pkg/companion/` — Not audited
- [ ] `pkg/social/` — Not audited
- [ ] `pkg/hostplay/` — Not audited
- [ ] `pkg/mobile/` — Not audited
- [ ] `pkg/vr/` — Not audited
- [ ] `pkg/visualtest/` — Not audited (test infrastructure)
- [ ] `pkg/audit/` — Not audited (audit infrastructure)

## Command Packages
- [ ] `cmd/client/` — Not audited
- [ ] `cmd/server/` — Not audited
- [ ] `cmd/mobile/` — Not audited

## Audit Statistics
- **Total Packages Identified**: ~70+
- **Audited**: 17
- **Completion Rate**: ~24%

## Priority Queue (High Integration Surface)
1. `pkg/procgen/genre/` — Genre system affects all procgen
2. `pkg/world/housing/` — Core player feature
3. `pkg/rendering/sprites/` — Critical rendering path
4. `pkg/engine/` (scoped sub-audits) — Core ECS systems
5. `pkg/network/federation/` — Cross-server functionality
