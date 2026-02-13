# Audit: github.com/opd-ai/venture/pkg/engine
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The engine package is the 240K+ LOC core of the ECS system containing 382 production Go files and 100+ systems. Build currently fails due to type error in test file (`weather_spell_damage_system_test.go`). Package exhibits extensive ECS component method violations (200+ instances) and time-based non-determinism in 15+ systems. Critical issues prevent production use, though most systems are functionally complete.

## Issues Found
- [x] **high** **stub/incomplete** — Build failure: type mismatch in `weather_spell_damage_system_test.go:72,105,137,158,185,205,231,257,278,309` - cannot use float literal as `particles.WeatherIntensity` type
- [x] **high** **ECS compliance** — 200+ component behavior methods violating ECS purity (`PvPRewardComponent.AddHonor()`, `EventRewardComponent.SpendCurrency()`, `OperatingHoursComponent.IsOpenAt()`, `ModBrowserComponent.SetSearchQuery()`, `DailyChallengeComponent.GenerateDailyChallenges()`, etc.) - components must be pure data, systems own behavior (`pvp_reward_component.go:159-582`, `event_reward_component.go:158-470`, `operating_hours_component.go:97-186`, `mod_browser_component.go:109-437`, `daily_challenge_component.go:228-822`, and 195+ more across 100+ component files)
- [x] **high** **deterministic procgen** — BountySystem uses `time.Now()` for expiration/timestamps without seed-based offset (`bounty_system.go:88,118,151,183,208,222,236`)
- [x] **high** **deterministic procgen** — ChallengeSystem uses `time.Now()` for time-based game state (`challenge_system.go:79,288,392`)
- [x] **high** **deterministic procgen** — ConversationManager generates UUIDs via `crypto/rand.Reader` with `time.Now()` fallback - non-reproducible (`conversation_manager.go:15,18`)
- [x] **high** **error handling** — Enhanced chat system hardcodes placeholder encryption key in production code (`enhanced_chat_system.go:76` - `key := []byte("chat_encryption_key_placeholder_32b")`)
- [x] **med** **stub/incomplete** — WorldEventsSystem has placeholder comment: "This is a placeholder for future integration" (`world_events_system.go:67`)
- [x] **med** **stub/incomplete** — HeadTrackingSystem method is placeholder for integration (`head_tracking_system.go:197`)
- [x] **med** **stub/incomplete** — SpellEffectSystem has 5 placeholder comments for missing integrations: material system, entity spawning, rendering/AI, elemental combo, spell modifier (`spell_effect_system.go:272,291,305,381,465`)
- [x] **med** **stub/incomplete** — MobileFederationSystem returns success placeholder without implementation (`mobile_federation_system.go:50`)
- [x] **med** **stub/incomplete** — CourierSystem uses placeholder logic for portal movement (`courier_system.go:177`)
- [x] **med** **stub/incomplete** — EnhancedChatSystem deliverToParty is placeholder until party system added (`enhanced_chat_system.go:347`)
- [x] **med** **stub/incomplete** — MultipartyAcceptance test is placeholder for trade conflict resolution (`multiparty_acceptance_test.go:128-130`)
- [x] **med** **stub/incomplete** — MailboxUI uses placeholder colored rectangle for text rendering (`mailbox_ui.go:895`)
- [x] **med** **doc coverage** — No package-level `doc.go` file exists (file found but content not validated)
- [x] **low** **test coverage** — Package build fails - unable to measure coverage (target: 65%+; estimated 80%+ based on test file count but unverifiable)
- [x] **low** **doc coverage** — MenuSystem addNoSavesMenuItem uses "placeholder" terminology in doc comment (`menu_system.go:846`)
- [x] **low** **error handling** — PerformanceMonitoring.CacheAndLOD has placeholder comment for hit rate calculation needing hit/miss tracking (`performance/cache_and_lod.go:156`)

## Test Coverage
Build failed - coverage unmeasurable (estimated 80%+ based on 354 `*_test.go` files, but blocked by type error)

## Integration Status
The engine package is the ECS core - all other packages depend on it. Systems are registered via `system_init.go` initialization. Components lacking Serialize/Deserialize for persistence: `AimComponent`, `ScreenShakeComponent`, `HitStopComponent`, `BehaviorTreeComponent`, `ExpressionComponent`, and 20+ others. Most critical systems (Movement, Collision, Combat, Rendering, AI, Network) are integrated. Placeholder integrations noted for: party system, material system, entity spawning coordination, elemental combo system, spell modifier system.

## Recommendations
1. **CRITICAL**: Fix build error in `weather_spell_damage_system_test.go` - cast float literals to `particles.WeatherIntensity` or update test to use correct type
2. **CRITICAL**: Refactor all component behavior methods to systems - move logic from 200+ methods like `PvPRewardComponent.AddHonor()` to corresponding systems; components should only have `Type()`, `Serialize()`, `Deserialize()`
3. **CRITICAL**: Replace `time.Now()` in BountySystem and ChallengeSystem with deterministic game clock based on seed + elapsed ticks
4. **CRITICAL**: Replace crypto/rand UUID generation in ConversationManager with seed-based deterministic ID generation (e.g., `fmt.Sprintf("%d-%d", seed, counter)`)
5. **HIGH**: Replace placeholder encryption key in EnhancedChatSystem with proper key derivation from game seed or configuration
6. **HIGH**: Complete placeholder implementations (SpellEffectSystem integrations, MobileFederationSystem, CourierSystem portal logic, party system for chat, MailboxUI text rendering)
7. **MED**: Add Serialize/Deserialize to components missing persistence support
8. **MED**: Validate `doc.go` content meets godoc standards, add if missing
9. **LOW**: Run tests once build is fixed to verify 65%+ coverage target
10. **LOW**: Remove "placeholder" terminology from production code comments where implementation is actually functional
