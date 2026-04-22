# Implementation Gaps — 2026-04-22

## Gap 1: Trade Quantity Contract Not Wired Into Runtime Trading
- **Intended Behavior**: Trade trust mechanics should enforce quantity constraints as documented and as represented by `TradeValidator.ValidateTradeQuantity`.
- **Current State**: Trade proposals store only item ID slices (`OfferedItems`/`RequestedItems`) with no quantity fields, and quantity validator is only exercised in tests.
- **Blocked Goal**: Complete, trust-aware multiplayer trading behavior (including quantity limits) is not enforceable.
- **Implementation Path**: Add quantity-bearing trade line item type to `TradeProposal`; update trade APIs and validation flow to validate quantities; update serialization/network payloads and trust validation to account for quantities.
- **Dependencies**: Requires schema/API update coordination across `pkg/engine`, `pkg/network/trade`, and validation package.
- **Effort**: medium

## Gap 2: Weapon Material Impact Particle System Missing Combat Callback Integration
- **Intended Behavior**: Material-aware melee impact particles should spawn when melee damage events occur.
- **Current State**: System is instantiated and added to world, but `OnMeleeImpact` is never registered to combat damage callbacks.
- **Blocked Goal**: Material-specific melee impact visual feedback never appears in normal combat flow.
- **Implementation Path**: Register `WeaponMaterialImpactParticleSystem.OnMeleeImpact` using combat damage callback wiring in shared initialization paths; add integration test validating particle spawn on melee hit.
- **Dependencies**: Depends on combat system callback ordering with existing lifesteal and durability callback registrations.
- **Effort**: small

## Gap 3: VR OpenXR Adapter Path Is Placeholder-Only and Unreachable
- **Intended Behavior**: When VR hardware is detected (and vr build mode is used), runtime should use hardware-backed adapters.
- **Current State**: OpenXR adapter methods are TODO/placeholder returns, and client VR initialization always installs stub adapters.
- **Blocked Goal**: Real headset/controller data path for VR mode remains unavailable despite adapter scaffolding.
- **Implementation Path**: Implement vr-tagged adapter selection in client initialization, integrate OpenXR method bodies, and keep stub fallback only for failures/no hardware.
- **Dependencies**: Requires OpenXR SDK availability and cgo build-path testing.
- **Effort**: large

## Gap 4: Unused Exported Sentinel Errors in Story and Prestige Packages
- **Intended Behavior**: Exported sentinel errors should be part of actual returned API error contracts.
- **Current State**: Multiple exported sentinel errors are declared but never returned.
- **Blocked Goal**: Public error contract clarity and reliable `errors.Is` usage for consumers.
- **Implementation Path**: Either wire these errors into relevant validation/lookup failures or remove them from exported API surface.
- **Dependencies**: Review call sites and tests that may rely on future `errors.Is` semantics.
- **Effort**: small

## Gap 5: Stale TODO Annotation in Quest Template Migration Comments
- **Intended Behavior**: TODO markers should represent active work only.
- **Current State**: One migration comment still carries unresolved TODO marker while adjacent files mark the same item resolved.
- **Blocked Goal**: Accurate audit signal and low-noise TODO scans.
- **Implementation Path**: Normalize migration comments to resolved wording and remove stale TODO token.
- **Dependencies**: None.
- **Effort**: small
