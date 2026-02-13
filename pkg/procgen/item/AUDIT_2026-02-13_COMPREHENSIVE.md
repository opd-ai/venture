# Audit: github.com/opd-ai/venture/pkg/procgen/item
**Date**: 2026-02-13
**Status**: Needs Work

## Summary
The item package provides procedural generation for weapons, armor, consumables, and accessories with seed-based deterministic generation. Test coverage is excellent at 91.7%, deterministic RNG usage is correct throughout, and the package passes all code quality checks. **UPDATED 2026-02-13**: Fixed critical integration gaps - Item struct fields (`ClassRestrictions`, `SpellEffectID`, `SpellDuration`, `SpellTargetType`, `SpellRadius`) are now properly populated during generation. Sci-fi consumables implemented. Remaining issues: 3 high-priority (horror/cyberpunk/postapoc genre templates missing).

## Issues Found
- [x] **high** stub/incomplete — `ClassRestrictions` field never populated during generation; field always empty array (`generator.go:155-178`) — **FIXED 2026-02-13**: Added `ClassRestrictions` field to `ItemTemplate` struct and populate it in `generateSingleItem()`
- [x] **high** stub/incomplete — `SpellEffectID` field never populated during generation; all scrolls non-functional (`generator.go:155-178`) — **FIXED 2026-02-13**: Added spell effect fields to `ItemTemplate` and copy to items during scroll generation
- [x] **high** stub/incomplete — `SpellDuration` field never populated during generation; field always zero (`generator.go:155-178`) — **FIXED 2026-02-13**: Added spell effect fields to `ItemTemplate` and copy to items during scroll generation
- [x] **high** stub/incomplete — `SpellTargetType` field never populated during generation; field always empty string (`generator.go:155-178`) — **FIXED 2026-02-13**: Added spell effect fields to `ItemTemplate` and copy to items during scroll generation
- [x] **high** stub/incomplete — `SpellRadius` field never populated during generation; field always zero (`generator.go:155-178`) — **FIXED 2026-02-13**: Added spell effect fields to `ItemTemplate` and copy to items during scroll generation
- [ ] **high** stub/incomplete — Horror genre templates completely missing; falls back to fantasy (`generator.go:41-54`, `templates.go`)
- [ ] **high** stub/incomplete — Cyberpunk genre templates completely missing; falls back to fantasy (`generator.go:41-54`, `templates.go`)
- [ ] **high** stub/incomplete — Post-apocalyptic genre templates missing; tested in `determinism_test.go:88` but not implemented
- [x] **high** stub/incomplete — Sci-fi consumable templates not implemented; `GetSciFiConsumableTemplates()` does not exist (`generator.go:48`, `templates.go:218-287`) — **FIXED 2026-02-13**: Added `GetSciFiConsumableTemplates()` and registered in generator
- [x] **med** integration — `ItemTemplate` struct lacks `ClassRestrictions` field to define class-restricted equipment (`templates.go:7-33`) — **FIXED 2026-02-13**: Added `ClassRestrictions` field to `ItemTemplate` struct
- [x] **med** integration — `ItemTemplate` struct lacks spell effect arrays (`SpellEffectIDs`, `SpellDurations`, `SpellTargetTypes`, `SpellRadii`) for scroll generation (`templates.go:7-33`) — **FIXED 2026-02-13**: Added spell effect fields to `ItemTemplate` struct
- [x] **med** integration — No registration of sci-fi consumables in generator initialization even though sci-fi weapons/armor exist (`generator.go:47-48`) — **FIXED 2026-02-13**: Registered `GetSciFiConsumableTemplates()` in generator init
- [x] **low** error-handling — Silent fallback to fantasy templates when unknown genre requested; should log warning (`generator.go:203-207, 211-216, 220-225`) — **FIXED 2026-02-13**: Added logging in `getWeaponTemplates()`, `getArmorTemplates()`, `getConsumableTemplates()` for unknown genre fallback

## Test Coverage
91.8% (target: 65%) ✅

**Coverage Breakdown:**
- Core generation: ✅ Excellent
- Determinism: ✅ Tested extensively (`determinism_test.go`)
- Class restrictions validation: ✅ Tested (`class_restrictions_test.go:85-122`)
- Rarity distribution: ✅ Tested (`item_test.go`)
- Projectile stats: ✅ Tested via template tests
- Benchmarks: ✅ Present (`item_bench_test.go`)

**Test Gaps (cannot test non-existent code):**
- ClassRestrictions field population (never assigned in generator)
- SpellEffectID field population (never assigned in generator)
- Horror/cyberpunk/postapoc genre generation (templates don't exist)
- Sci-fi consumable generation (function doesn't exist)

## Integration Status
The package is imported by 20+ engine systems:

**Confirmed Integration Points:**
- ✅ `pkg/engine/item_spawning.go` — Uses `ItemGenerator.Generate()`
- ✅ `pkg/engine/crafting_system.go` — Uses `ItemGenerator` for crafting output
- ✅ `pkg/engine/legendary_quest_system.go` — Uses `ItemGenerator` for quest rewards
- ✅ `pkg/engine/minigame_system.go` — Uses `ItemGenerator` for minigame prizes
- ✅ `pkg/engine/merchant_spawn.go` — Uses items for merchant inventories
- ❌ `pkg/engine/inventory_system.go:178-179` — **BROKEN**: Calls `validateClassRestrictions()` which checks `item.ClassRestrictions`, but field is always empty `[]` so all items usable by all classes
- ❌ `pkg/engine/inventory_system.go:344-353` — **BROKEN**: `applyScrollSpellEffect()` reads `item.SpellEffectID` at line 350, but field is always empty string `""` so scroll consumption does nothing (see `validateScrollEffectSystem:695`)
- ❌ `pkg/engine/inventory_system.go:695-707` — **BROKEN**: `validateScrollEffectSystem()` returns `false` when `itm.SpellEffectID == ""` (line 700), preventing all scroll usage
- ✅ `pkg/engine/combat_system.go` — Uses `item.Stats` for damage calculations
- ✅ `pkg/engine/equipment_visual_system.go` — Uses `item.Type`, `item.WeaponType`, `item.ArmorType`
- ✅ `pkg/engine/inventory_components.go` — Stores `[]*item.Item` in `InventoryComponent.Items`
- ✅ `pkg/engine/hotbar_component.go` — References `*item.Item`
- ✅ `pkg/engine/commerce_system.go` — Uses items for trading
- ✅ `pkg/engine/carryover_system.go` — Persists items across game sessions
- ✅ `pkg/engine/companion_inventory_system.go` — Companions use items

**Critical Functional Breakage:**
1. **All generated items are usable by all classes** — Despite `CanBeUsedByClass()` method existing and being tested, `ClassRestrictions` is never populated, so validation always succeeds (empty array = no restrictions)
2. **All generated scrolls are non-functional** — `applyScrollSpellEffect()` requires `SpellEffectID` to be populated but generator never sets it, causing `validateScrollEffectSystem()` to return false and abort scroll usage
3. **Genre system incomplete** — Tests in `determinism_test.go:88` reference 5 genres (fantasy, scifi, horror, cyberpunk, postapoc) but only 2 are implemented, with silent fallback to fantasy causing incorrect item flavor

## Recommendations

### HIGH PRIORITY: Fix Scroll Spell Effect Generation
**Impact:** All scrolls generated by the system are non-functional  
**Files:** `templates.go`, `generator.go`

1. Add spell effect fields to `ItemTemplate` struct (`templates.go:7-33`):
   ```go
   SpellEffectIDs    []string       // Pool of spell IDs this template can generate
   SpellDurations    []float64      // Corresponding durations (parallel array)
   SpellTargetTypes  []string       // Corresponding target types (parallel array)
   SpellRadii        []float64      // Corresponding radii (parallel array)
   ```

2. Update `GetFantasyConsumableTemplates()` to populate spell data for scrolls (`templates.go:195-217`):
   ```go
   {
       BaseType:       TypeConsumable,
       ConsumableType: ConsumableScroll,
       NamePrefixes:   []string{"Scroll of", "Ancient", "Mystic"},
       NameSuffixes:   []string{"Fireball", "Lightning", "Ice", "Protection"},
       Tags:           []string{"magical", "spell", "consumable"},
       ValueRange:     [2]int{20, 150},
       WeightRange:    [2]float64{0.1, 0.2},
       SpellEffectIDs:   []string{"fireball", "lightning", "ice_nova", "protection_ward"},
       SpellDurations:   []float64{0.0, 0.0, 0.0, 10.0}, // 0.0 = instant
       SpellTargetTypes: []string{"area", "entity", "area", "self"},
       SpellRadii:       []float64{80.0, 0.0, 100.0, 0.0},
   }
   ```

3. Update `generateSingleItem()` to copy spell data from template (`generator.go:155-178`):
   ```go
   // After line 176 (after item.Description = ...)
   if item.ConsumableType == ConsumableScroll && len(template.SpellEffectIDs) > 0 {
       // Select spell effect from template using RNG
       spellIndex := rng.Intn(len(template.SpellEffectIDs))
       item.SpellEffectID = template.SpellEffectIDs[spellIndex]
       item.SpellDuration = template.SpellDurations[spellIndex]
       item.SpellTargetType = template.SpellTargetTypes[spellIndex]
       item.SpellRadius = template.SpellRadii[spellIndex]
   }
   ```

4. Add test: `TestScrollSpellEffectGeneration` to verify spell fields populated for scrolls but not potions

### HIGH PRIORITY: Fix Class Restriction Generation
**Impact:** Class-specific equipment restrictions are non-functional  
**Files:** `templates.go`, `generator.go`

1. Add `ClassRestrictions []string` field to `ItemTemplate` struct (`templates.go:7-33`)

2. Define class restrictions in weapon templates (`templates.go:36-154`):
   - Heavy weapons (axe line 50-61, crossbow line 110-129): `ClassRestrictions: []string{"warrior", "ranger"}`
   - Magic weapons (staff line 84-95, wand line 131-153): `ClassRestrictions: []string{"mage", "cleric", "necromancer"}`
   - Light weapons (dagger line 97-108): `ClassRestrictions: []string{"rogue", "ranger"}`
   - Bows (line 62-83): `ClassRestrictions: []string{"ranger", "rogue"}`
   - Swords (line 37-49): `ClassRestrictions: []string{}` (no restrictions)

3. Define class restrictions in armor templates (`templates.go:156-193`):
   - Plate armor (chest line 159-169, helmet line 170-180): `ClassRestrictions: []string{"warrior", "cleric"}`
   - Shields (line 181-192): `ClassRestrictions: []string{"warrior", "cleric"}`

4. Update `generateSingleItem()` to copy restrictions from template (`generator.go:155-178`):
   ```go
   // After line 164 (after copy(item.Tags, template.Tags))
   if len(template.ClassRestrictions) > 0 {
       item.ClassRestrictions = make([]string, len(template.ClassRestrictions))
       copy(item.ClassRestrictions, template.ClassRestrictions)
   }
   ```

5. Add test: `TestClassRestrictionsGeneration` to verify restrictions match templates

### HIGH PRIORITY: Implement Horror Genre Templates
**Impact:** Horror genre falls back to fantasy, breaking thematic consistency  
**Files:** `templates.go` (new functions), `generator.go:41-54`

1. Create `GetHorrorWeaponTemplates()` in `templates.go`:
   - Ritual Dagger, Bone Club, Cursed Blade, Meat Hook, Barbed Whip
   - Tags: "cursed", "flesh", "bone", "ritual"

2. Create `GetHorrorArmorTemplates()` in `templates.go`:
   - Tattered Robes, Bone Armor, Flesh Suit, Blood-Soaked Cloak
   - Tags: "cursed", "decrepit", "unholy"

3. Create `GetHorrorConsumableTemplates()` in `templates.go`:
   - Blood Vials, Sanity Potions, Cursed Scrolls
   - SpellEffectIDs: "blood_ritual", "madness_ward", "soul_drain"

4. Register in `NewItemGeneratorWithLogger()` (`generator.go:41-54`):
   ```go
   gen.weaponTemplates["horror"] = GetHorrorWeaponTemplates()
   gen.armorTemplates["horror"] = GetHorrorArmorTemplates()
   gen.consumableTemplates["horror"] = GetHorrorConsumableTemplates()
   ```

5. Add test: `TestHorrorGenreGeneration`

### HIGH PRIORITY: Implement Cyberpunk Genre Templates
**Impact:** Cyberpunk genre falls back to fantasy, breaking thematic consistency  
**Files:** `templates.go` (new functions), `generator.go:41-54`

1. Create `GetCyberpunkWeaponTemplates()` in `templates.go`:
   - Monowire, Smart Pistol, Cyberdeck Weapon, EMP Grenade, Mantis Blades
   - Tags: "cyber", "tech", "smart", "neural"

2. Create `GetCyberpunkArmorTemplates()` in `templates.go`:
   - Cyberware, Ballistic Vest, Neural Shielding, Optical Camo, Subdermal Plating
   - Tags: "augment", "tech", "chrome"

3. Create `GetCyberpunkConsumableTemplates()` in `templates.go`:
   - Stim Packs, ICE Breakers, Neural Boosters, Quickhacks
   - SpellEffectIDs: "neural_boost", "ice_break", "system_shock"

4. Register in `NewItemGeneratorWithLogger()` (`generator.go:41-54`)

5. Add test: `TestCyberpunkGenreGeneration`

### HIGH PRIORITY: Implement Sci-Fi Consumables
**Impact:** Sci-fi genre has weapons/armor but no consumables  
**Files:** `templates.go` (new function), `generator.go:48`

1. Create `GetSciFiConsumableTemplates()` in `templates.go`:
   - Med Packs, Energy Cells, Nanobots, Shield Boosters, Oxygen Canisters
   - Tags: "medical", "energy", "tech"

2. Register in `NewItemGeneratorWithLogger()` at line 48:
   ```go
   gen.consumableTemplates["scifi"] = GetSciFiConsumableTemplates()
   ```

3. Add test: `TestSciFiConsumableGeneration`

### MEDIUM PRIORITY: Add Unknown Genre Warning
**Impact:** Silent fallback makes debugging difficult  
**Files:** `generator.go:203-225`

Add logging in fallback paths:
```go
func (g *ItemGenerator) getWeaponTemplates(genreID string) []ItemTemplate {
    templates := g.weaponTemplates[genreID]
    if templates == nil {
        if g.logger != nil && genreID != "" {
            g.logger.WithField("genreID", genreID).Warn("unknown genre, using default weapon templates")
        }
        templates = g.weaponTemplates[""]
    }
    return templates
}
```

### LOW PRIORITY: Implement Post-Apocalyptic Genre
**Impact:** Genre is tested but not implemented  
**Files:** `templates.go` (new functions), `generator.go:41-54`

Referenced in `determinism_test.go:88` but no templates exist. Add templates similar to horror/cyberpunk.

## Validation Checklist

✅ **Stub/incomplete code** — 9 high-priority stubs found (field population, genre templates)  
✅ **ECS compliance** — N/A (not an ECS component package)  
✅ **Deterministic procgen** — ✅ All RNG uses `rand.New(rand.NewSource(seed))`, no global rand  
✅ **Network interfaces** — N/A (no network code)  
✅ **Error handling** — ✅ All errors checked and logged appropriately  
✅ **Test coverage** — ✅ 91.8% exceeds 65% target  
✅ **Doc coverage** — ✅ Has `doc.go` with comprehensive package docs; all exported types documented  
✅ **Integration points** — ❌ Critical integration failures with inventory system due to unpopulated fields

## go vet Status
✅ **PASS** — No issues found
