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
- [x] `cmd/mobile/AUDIT.md` — Needs Work — 13 issues (3 high, 5 med, 5 low) — Coverage: 36.9% (0.0% mobile.go, 73.9% config/)

### Core Packages
- [x] `pkg/engine/AUDIT.md` — Needs Work — 9 issues (1 high, 4 med, 4 low) — Coverage: 94.0-96.3% (subpackages; main requires X11)
- [x] `pkg/engine/qol/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 94.0%
- [x] `pkg/engine/physics/AUDIT.md` — Needs Work — 9 issues (2 high, 5 med, 2 low) — Coverage: 95.4%
- [x] `pkg/engine/physics/fluids/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 95.2%
- [x] `pkg/engine/physics/vehicle/AUDIT.md` — Needs Work — 8 issues (1 high, 4 med, 3 low) — Coverage: 94.8%
- [x] `pkg/engine/physics/destruction/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 96.3%
- [x] `pkg/engine/prestige/AUDIT.md` — Needs Work — 8 issues (1 high, 3 med, 4 low) — Coverage: Unmeasurable (requires X11; 103.5% test-to-source ratio)
- [x] `pkg/engine/performance/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 95.8%
- [x] `pkg/procgen/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 100.0%
- [x] `pkg/procgen/book/AUDIT.md` — Complete — 1 issue (0 high, 0 med, 1 low) — Coverage: Unmeasurable (requires X11; 35.6% test-to-source ratio)
- [x] `pkg/procgen/environment/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 95.5%
- [x] `pkg/procgen/terrain/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 94.0%
- [x] `pkg/procgen/quest/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 92.3%
- [x] `pkg/procgen/item/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 92.2%
- [x] `pkg/procgen/dialog/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 88.0%
- [x] `pkg/procgen/entity/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 92.4%
- [x] `pkg/procgen/legendary/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 86.6%
- [x] `pkg/procgen/companion/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 56% test-to-source ratio)
- [x] `pkg/procgen/magic/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 89.8%
- [x] `pkg/procgen/genre/AUDIT.md` — Complete — 4 issues (0 high, 0 med, 4 low) — Coverage: 94.8%
- [x] `pkg/procgen/skills/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 87.0%
- [x] `pkg/procgen/narrative/AUDIT.md` — Complete — 5 issues (0 high, 1 med, 4 low) — Coverage: 91.9%
- [x] `pkg/procgen/building/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 92.2%
- [x] `pkg/procgen/vehicle/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 62.8% test-to-source ratio)
- [x] `pkg/procgen/faction/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 122% test-to-source ratio)
- [x] `pkg/procgen/minigame/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 127% test-to-source ratio)
- [x] `pkg/procgen/recipe/AUDIT.md` — Complete — 5 issues (0 high, 1 med, 4 low) — Coverage: Unmeasurable (requires X11; 47.9% test-to-source ratio)
- [x] `pkg/procgen/station/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 89.0%
- [x] `pkg/procgen/puzzle/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 94.3%
- [x] `pkg/procgen/class/AUDIT.md` — Needs Work — 7 issues (2 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 87.4% test-to-source ratio)
- [x] `pkg/procgen/furniture/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 92.5%
- [ ] `pkg/rendering/` — Not audited (subdirectories audited separately)
- [x] `pkg/rendering/animation/AUDIT.md` — Complete — 1 issue (0 high, 0 med, 1 low) — Coverage: Unmeasurable (requires X11; 44.8% test-to-source ratio)
- [x] `pkg/rendering/cache/AUDIT.md` — Complete — 4 issues (0 high, 0 med, 4 low) — Coverage: Unmeasurable (requires X11; 245% test-to-source ratio)
- [x] `pkg/rendering/particles/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 91.8%
- [x] `pkg/rendering/pool/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 30% target; 452 lines tests)
- [x] `pkg/rendering/sprites/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 452% test-to-source ratio)
- [x] `pkg/rendering/ui/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 91% test-to-source ratio)
- [x] `pkg/rendering/lighting/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 56.6% test-to-source ratio)
- [x] `pkg/rendering/postprocess/AUDIT.md` — Complete — 2 issues (0 high, 0 med, 2 low) — Coverage: Unmeasurable (requires X11; 66% test-to-source ratio)
- [x] `pkg/network/federation/AUDIT.md` — Complete — 4 issues (1 high, 2 med, 1 low) — Coverage: 85.8% (subpackages; parent requires X11)
- [x] `pkg/network/federation/webrtc/AUDIT.md` — Complete — 4 issues (0 high, 2 med, 2 low) — Coverage: 85.7%
- [x] `pkg/network/federation/guild/AUDIT.md` — Complete — 4 issues (0 high, 0 med, 4 low) — Coverage: 88.2%
- [x] `pkg/network/federation/mobile/AUDIT.md` — Complete — 6 issues (0 high, 3 med, 3 low) — Coverage: 82.4%
- [x] `pkg/network/chat/AUDIT.md` — Complete — 4 issues (1 high, 2 med, 2 low) — Coverage: Unmeasurable (requires X11; 30% target)
- [x] `pkg/network/trade/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 145% test-to-source ratio)
- [x] `pkg/network/resilience/AUDIT.md` — Complete — 4 issues (0 high, 2 med, 2 low) — Coverage: 89.3%
- [x] `pkg/network/AUDIT.md` — Needs Work — 11 issues (1 high, 5 med, 5 low) — Coverage: Unmeasurable (requires X11; 66% test-to-source ratio)
- [x] `pkg/world/AUDIT.md` — Needs Work — 12 issues (1 high, 6 med, 5 low) — Coverage: 88.8% (core), 88.4% (economy), 90.4% (raids), 90.8% (territory), FAIL (housing)
- [x] `pkg/world/housing/AUDIT.md` — Complete — 4 issues (0 high, 2 med, 2 low) — Coverage: Unmeasurable (requires X11; 139% test-to-source ratio)
- [x] `pkg/world/raids/AUDIT.md` — Complete — 6 issues (1 high, 2 med, 3 low) — Coverage: 90.4%
- [x] `pkg/world/territory/AUDIT.md` — Needs Work — 9 issues (3 high, 3 med, 3 low) — Coverage: 90.8%
- [x] `pkg/audio/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 93.5%
- [x] `pkg/audio/synthesis/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 95.1%

### World Management Packages
- [x] `pkg/world/economy/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 88.4%

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
- [x] `pkg/companion/learning/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 92.5%
- [x] `pkg/social/AUDIT.md` — Complete — 4 issues (0 high, 0 med, 4 low) — Coverage: 95.6% (98.0% social, 93.2% persistence)
- [x] `pkg/hostplay/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 70% test-to-source ratio)
- [x] `pkg/mobile/AUDIT.md` — Complete — 5 issues (1 high, 2 med, 2 low) — Coverage: Unmeasurable (requires X11; 135% test-to-source ratio)
- [ ] `pkg/audit/` — Not audited (only contains `features/` subdir which is already audited)
- [x] `pkg/visualtest/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 88.1% (parity); 113% test-to-source ratio
- [x] `pkg/vr/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low) — Coverage: 76.8%
- [x] `pkg/memprofile/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 88.8%
- [x] `pkg/combat/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 98.3%
- [x] `pkg/audit/features/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 99.2%

### Integration Packages
- [x] `pkg/integration/AUDIT.md` — Complete — 3 issues (0 high, 2 med, 1 low) — Coverage: [no statements] (meta-package; subdirs: 89.3-96.3%)
- [x] `pkg/integration/companion_housing/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low) — Coverage: 93.2%
- [x] `pkg/integration/guild_housing/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 123% test-to-source ratio)
- [x] `pkg/integration/guild_vehicle/AUDIT.md` — Complete — 4 issues (0 high, 2 med, 2 low) — Coverage: 94.0%
- [x] `pkg/integration/housing_crafting/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: 96.3%
- [x] `pkg/integration/choice_consequences/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 89.3%
- [x] `pkg/integration/narrative_world/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (86.7% test-to-source ratio)
- [x] `pkg/integration/political_warfare/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 452% test-to-source ratio)
- [x] `pkg/integration/trade_routes/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: Unmeasurable (requires X11; 99% test-to-source ratio)
- [x] `pkg/integration/world_events/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: 92.9%

## Audit Progress
- **Completed**: 89/90+ packages
- **In Progress**: 0
- **Not Started**: 1 (pkg/procgen/minigame/games)

### Supporting Packages (continued)
- [x] `pkg/benchmark/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: N/A (test-only; 513 test lines; fps requires X11)

- [x] `pkg/rendering/tiles/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low) — Coverage: 91.5%

- [x] `pkg/rendering/display/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 64% test-to-source ratio)

- [x] `pkg/rendering/palette/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 97.0%

- [x] `pkg/rendering/quality/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 96.8%

- [x] `pkg/rendering/patterns/AUDIT.md` — Complete — 4 issues (0 high, 1 med, 3 low) — Coverage: 94.1%

- [x] `pkg/rendering/parallel/AUDIT.md` — Complete — 3 issues (0 high, 1 med, 2 low) — Coverage: 96.7%

- [x] `pkg/procgen/story/AUDIT.md` — Needs Work — 10 issues (3 high, 4 med, 3 low) — Coverage: 88.7% — Re-audited 2026-02-26
- [x] `pkg/procgen/audit/AUDIT.md` — Complete — 3 issues (0 high, 0 med, 3 low) — Coverage: Unmeasurable (requires X11; 1,342% test-to-source ratio)

- [x] `pkg/rendering/shapes/AUDIT.md` — Complete — 5 issues (0 high, 2 med, 3 low) — Coverage: Unmeasurable (requires X11; 120.5% test-to-source ratio)

## Next Priority
Packages with high integration surface and platform-specific concerns:
1. `pkg/procgen/` — Procedural generation (all subdirectories audited)
