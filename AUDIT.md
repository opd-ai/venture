# Venture Code Audit Tracker

This file tracks package-level audits across the Venture codebase. Each entry indicates audit completion status, issue counts, and test coverage.

## Audit Status Legend
- ✅ Complete: All automated checks passed and fewer than 5 non-critical issues identified
- ⚠️ Needs Work: 5 or more issues identified, or any critical/priority-0 failure
- ❌ Incomplete: Audit was stopped early or required checks were not run

## Package Audits

### Command Packages
- [x] `cmd/client/AUDIT_2026-02-16_COMPREHENSIVE.md` — Pass — 7 issues (0 high, 1 med, 6 low) — Coverage: 49.4%
- [x] `cmd/server/AUDIT.md` — Complete — 7 issues (0 high, 3 med, 4 low) — Coverage: Unmeasurable (requires X11; 184% test-to-source ratio)
- [ ] `cmd/mobile/` — Not audited

### Core Packages
- [x] `pkg/engine/AUDIT.md` — Needs Work — 9 issues (1 high, 4 med, 4 low) — Coverage: 94.0-96.3% (subpackages; main requires X11)
- [ ] `pkg/procgen/` — Not audited (subdirectories audited separately)
- [x] `pkg/procgen/terrain/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 94.0%
- [x] `pkg/procgen/quest/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 92.3%
- [x] `pkg/procgen/item/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 92.2%
- [x] `pkg/procgen/dialog/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 88.0%
- [ ] `pkg/rendering/` — Not audited (subdirectories audited separately)
- [x] `pkg/rendering/pool/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 30% target; 452 lines tests)
- [x] `pkg/rendering/sprites/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 452% test-to-source ratio)
- [x] `pkg/network/federation/AUDIT.md` — Complete — 4 issues (1 high, 2 med, 1 low) — Coverage: 85.8% (subpackages; parent requires X11)
- [x] `pkg/network/chat/AUDIT.md` — Complete — 4 issues (1 high, 2 med, 2 low) — Coverage: Unmeasurable (requires X11; 30% target)
- [x] `pkg/network/trade/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 145% test-to-source ratio)
- [ ] `pkg/network/` (other subpackages) — Not audited
- [x] `pkg/world/AUDIT.md` — Needs Work — 12 issues (1 high, 6 med, 5 low) — Coverage: 88.8% (core), 88.4% (economy), 90.4% (raids), 90.8% (territory), FAIL (housing)
- [x] `pkg/audio/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 93.5%

### Supporting Packages
- [x] `pkg/saveload/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 85.5%
- [x] `pkg/config/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 100.0%
- [x] `pkg/validation/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 98.5%
- [x] `pkg/errors/AUDIT.md` — Complete — 1 issue (0 high, 0 med, 1 low) — Coverage: 100.0%
- [x] `pkg/logging/AUDIT.md` — Complete — 2 issues (0 high, 1 med, 1 low) — Coverage: 100.0%
- [x] `pkg/recovery/AUDIT.md` — Complete — 1 issue (0 high, 1 med, 0 low) — Coverage: 100.0%
- [x] `pkg/stability/AUDIT.md` — Complete — 3 issues (0 high, 2 med, 1 low) — Coverage: 94.4%
- [x] `pkg/observability/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 97.4%
- [x] `pkg/security/AUDIT.md` — Complete — 6 issues (0 high, 2 med, 4 low) — Coverage: 90.0%
- [x] `pkg/version/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 100.0%
- [x] `pkg/migration/AUDIT.md` — Complete — 2 issues (0 high, 0 med, 2 low) — Coverage: 91.3%
- [x] `pkg/modding/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 90.6%
- [x] `pkg/narrative/branching/AUDIT.md` — Complete — 3 issues (0 high, 2 med, 4 low) — Coverage: 88.3%
- [x] `pkg/ux/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 96.9%
- [x] `pkg/balance/AUDIT.md` — Complete — 7 issues (1 high, 2 med, 4 low) — Coverage: Unmeasurable (requires X11; 228% test-to-source ratio)
- [x] `pkg/class/advanced/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 91.8%
- [ ] `pkg/companion/` — Not audited
- [ ] `pkg/social/` — Not audited
- [x] `pkg/hostplay/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 70% test-to-source ratio)
- [x] `pkg/mobile/AUDIT.md` — Complete — 5 issues (1 high, 2 med, 2 low) — Coverage: Unmeasurable (requires X11; 135% test-to-source ratio)
- [ ] `pkg/audit/` — Not audited
- [ ] `pkg/visualtest/` — Not audited
- [ ] `pkg/vr/` — Not audited
- [ ] `pkg/memprofile/` — Not audited
- [x] `pkg/combat/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 98.3%

### Integration Packages
- [x] `pkg/integration/companion_housing/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low) — Coverage: 93.2%
- [ ] `pkg/integration/guild_housing/` — Not audited
- [ ] `pkg/integration/guild_vehicle/` — Not audited
- [ ] `pkg/integration/housing_crafting/` — Not audited
- [ ] `pkg/integration/choice_consequences/` — Not audited
- [ ] `pkg/integration/narrative_world/` — Not audited
- [ ] `pkg/integration/political_warfare/` — Not audited
- [ ] `pkg/integration/trade_routes/` — Not audited
- [ ] `pkg/integration/world_events/` — Not audited

## Audit Progress
- **Completed**: 33/90+ packages
- **In Progress**: 0
- **Not Started**: 57+

## Next Priority
Packages with high integration surface and platform-specific concerns:
1. `pkg/engine/` — Core ECS implementation, 400+ files (highest priority: all systems)
2. `pkg/network/` — Remaining subpackages (chat, trade, resilience, core)
3. `pkg/rendering/` — Graphics pipeline
4. `pkg/world/` — World state and persistence
5. `pkg/procgen/` — Procedural generation (terrain, entity, quest, etc.)
