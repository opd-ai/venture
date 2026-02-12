# Audit: github.com/opd-ai/venture/pkg/procgen/item
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The item package provides procedural generation for weapons, armor, consumables, and accessories with deterministic seed-based generation. The package is well-structured with excellent test coverage (91.8%) and follows proper deterministic generation patterns. However, it has critical integration gaps: newly-added Item fields (ClassRestrictions, SpellEffectID, SpellDuration, SpellTargetType, SpellRadius) are defined and tested for validation logic but never populated during generation, breaking integration with engine systems that expect these fields.

## Issues Found
- [ ] **high** stub/incomplete — `ClassRestrictions` field never populated during generation (`generator.go:155-178`)
- [ ] **high** stub/incomplete — `SpellEffectID` field never populated during generation (`generator.go:155-178`)
- [ ] **high** stub/incomplete — `SpellDuration` field never populated during generation (`generator.go:155-178`)
- [ ] **high** stub/incomplete — `SpellTargetType` field never populated during generation (`generator.go:155-178`)
- [ ] **high** stub/incomplete — `SpellRadius` field never populated during generation (`generator.go:155-178`)
- [ ] **med** integration — `ItemTemplate` lacks fields for class restrictions and spell effects (`templates.go:7-33`)
- [ ] **low** documentation — README claims 93.8% coverage but actual is 91.8% (`README.md:184`)

## Test Coverage
91.8% (target: 65%) ✅

## Integration Status
The package integrates with multiple engine systems:
- ✅ **CraftingSystem** (`pkg/engine/crafting_system.go:46-57`) — Uses ItemGenerator
- ✅ **InventorySystem** (`pkg/engine/inventory_system.go:178-179`) — Validates ClassRestrictions via `validateClassRestrictions()`
- ✅ **InventorySystem** (`pkg/engine/inventory_system.go:345,703-793`) — Uses SpellEffectID and related fields for scroll spell activation
- ✅ **ItemSpawning** (`pkg/engine/item_spawning.go:120`) — Uses ItemGenerator
- ✅ **LegendaryQuestSystem** (`pkg/engine/legendary_quest_system.go:18-29`) — Uses ItemGenerator
- ✅ **MiniGameSystem** (`pkg/engine/minigame_system.go:21-42`) — Uses ItemGenerator

**Critical Gap**: Engine systems expect `ClassRestrictions` and spell effect fields to be populated, but the generator leaves them empty. This means:
1. All generated items are usable by all classes (no class restrictions enforced)
2. All generated scrolls have no spell effects (scrolls are non-functional)

## Recommendations
1. **HIGH PRIORITY: Add spell effect generation for consumable scrolls**
   - Add `SpellEffectIDs`, `SpellDurations`, `SpellTargetTypes`, `SpellRadii` fields to `ItemTemplate`
   - Populate spell effect fields in `GetFantasyConsumableTemplates()` for scroll templates
   - Update `generateSingleItem()` to copy spell effect data from template to item for ConsumableScroll types
   - Add test coverage for spell effect field generation

2. **HIGH PRIORITY: Add class restriction generation for equipment**
   - Add `ClassRestrictions` field to `ItemTemplate`
   - Define class restrictions in weapon/armor templates (e.g., staffs for mages, heavy armor for warriors)
   - Update `generateSingleItem()` to copy class restrictions from template to item
   - Add test coverage for class restriction generation

3. **MEDIUM PRIORITY: Extend ItemTemplate structure**
   - Ensure templates can express all Item fields that should be deterministically generated
   - Consider template inheritance or composition for shared properties

4. **LOW PRIORITY: Update README coverage claim**
   - Change line 184 from "**93.8%**" to "**91.8%**" to match actual coverage
