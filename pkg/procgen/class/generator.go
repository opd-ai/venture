package class

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

// ClassPreset defines a character class configuration.
type ClassPreset struct {
	Type              engine.CharacterClass
	Name              string
	Description       string
	StartingHP        float64
	StartingMana      float64
	StartingAttack    float64
	StartingDefense   float64
	StartingSpeed     float64
	StartingAbilities []string
	Specializations   []engine.SpecializationType
}

// ClassGenerator generates character class configurations.
type ClassGenerator struct {
	presets map[engine.CharacterClass]ClassPreset
}

// NewClassGenerator creates a new class generator.
func NewClassGenerator() *ClassGenerator {
	gen := &ClassGenerator{
		presets: make(map[engine.CharacterClass]ClassPreset),
	}
	gen.initializePresets()
	return gen
}

// initializePresets sets up the base class presets.
func (g *ClassGenerator) initializePresets() {
	g.presets[engine.ClassWarrior] = ClassPreset{
		Type:              engine.ClassWarrior,
		Name:              "Warrior",
		Description:       "A mighty combatant specializing in melee combat and heavy armor.",
		StartingHP:        100.0,
		StartingMana:      30.0,
		StartingAttack:    15.0,
		StartingDefense:   12.0,
		StartingSpeed:     5.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassWarrior),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassWarrior),
	}

	g.presets[engine.ClassRogue] = ClassPreset{
		Type:              engine.ClassRogue,
		Name:              "Rogue",
		Description:       "A swift and deadly striker who relies on speed and precision.",
		StartingHP:        70.0,
		StartingMana:      50.0,
		StartingAttack:    18.0,
		StartingDefense:   8.0,
		StartingSpeed:     15.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassRogue),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassRogue),
	}

	g.presets[engine.ClassMage] = ClassPreset{
		Type:              engine.ClassMage,
		Name:              "Mage",
		Description:       "A master of arcane magic who wields devastating spells.",
		StartingHP:        60.0,
		StartingMana:      120.0,
		StartingAttack:    10.0,
		StartingDefense:   6.0,
		StartingSpeed:     8.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassMage),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassMage),
	}

	g.presets[engine.ClassRanger] = ClassPreset{
		Type:              engine.ClassRanger,
		Name:              "Ranger",
		Description:       "A skilled archer and beast tamer who excels at ranged combat.",
		StartingHP:        85.0,
		StartingMana:      60.0,
		StartingAttack:    14.0,
		StartingDefense:   10.0,
		StartingSpeed:     12.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassRanger),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassRanger),
	}

	g.presets[engine.ClassCleric] = ClassPreset{
		Type:              engine.ClassCleric,
		Name:              "Cleric",
		Description:       "A divine caster who heals allies and smites enemies.",
		StartingHP:        90.0,
		StartingMana:      100.0,
		StartingAttack:    12.0,
		StartingDefense:   11.0,
		StartingSpeed:     8.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassCleric),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassCleric),
	}

	g.presets[engine.ClassNecromancer] = ClassPreset{
		Type:              engine.ClassNecromancer,
		Name:              "Necromancer",
		Description:       "A dark mage who commands the undead and drains life force.",
		StartingHP:        65.0,
		StartingMana:      110.0,
		StartingAttack:    11.0,
		StartingDefense:   7.0,
		StartingSpeed:     7.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassNecromancer),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassNecromancer),
	}

	// Hybrid classes
	g.presets[engine.ClassBattlemage] = ClassPreset{
		Type:              engine.ClassBattlemage,
		Name:              "Battlemage",
		Description:       "Armored spellcaster combining martial prowess with destructive magic.",
		StartingHP:        80.0,
		StartingMana:      75.0,
		StartingAttack:    12.5,
		StartingDefense:   9.0,
		StartingSpeed:     6.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassBattlemage),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassBattlemage),
	}

	g.presets[engine.ClassSpellblade] = ClassPreset{
		Type:              engine.ClassSpellblade,
		Name:              "Spellblade",
		Description:       "Agile warrior-mage weaving spells between swift strikes.",
		StartingHP:        65.0,
		StartingMana:      85.0,
		StartingAttack:    14.0,
		StartingDefense:   7.0,
		StartingSpeed:     11.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassSpellblade),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassSpellblade),
	}

	g.presets[engine.ClassPaladin] = ClassPreset{
		Type:              engine.ClassPaladin,
		Name:              "Paladin",
		Description:       "Holy warrior blending heavy armor with divine healing.",
		StartingHP:        95.0,
		StartingMana:      65.0,
		StartingAttack:    13.5,
		StartingDefense:   11.5,
		StartingSpeed:     6.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassPaladin),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassPaladin),
	}

	g.presets[engine.ClassMonk] = ClassPreset{
		Type:              engine.ClassMonk,
		Name:              "Monk",
		Description:       "Unarmed combatant using spiritual energy and incredible speed.",
		StartingHP:        80.0,
		StartingMana:      75.0,
		StartingAttack:    13.0,
		StartingDefense:   9.0,
		StartingSpeed:     11.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassMonk),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassMonk),
	}

	g.presets[engine.ClassDeathKnight] = ClassPreset{
		Type:              engine.ClassDeathKnight,
		Name:              "Death Knight",
		Description:       "Fallen warrior wielding dark necromantic powers.",
		StartingHP:        82.5,
		StartingMana:      70.0,
		StartingAttack:    13.0,
		StartingDefense:   9.5,
		StartingSpeed:     6.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassDeathKnight),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassDeathKnight),
	}

	g.presets[engine.ClassWitchHunter] = ClassPreset{
		Type:              engine.ClassWitchHunter,
		Name:              "Witch Hunter",
		Description:       "Divine marksman specializing in hunting supernatural threats.",
		StartingHP:        87.5,
		StartingMana:      80.0,
		StartingAttack:    13.0,
		StartingDefense:   10.5,
		StartingSpeed:     10.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassWitchHunter),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassWitchHunter),
	}

	g.presets[engine.ClassBeastlord] = ClassPreset{
		Type:              engine.ClassBeastlord,
		Name:              "Beastlord",
		Description:       "Savage warrior commanding powerful beasts.",
		StartingHP:        92.5,
		StartingMana:      45.0,
		StartingAttack:    14.5,
		StartingDefense:   11.0,
		StartingSpeed:     8.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassBeastlord),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassBeastlord),
	}

	g.presets[engine.ClassArcaneArcher] = ClassPreset{
		Type:              engine.ClassArcaneArcher,
		Name:              "Arcane Archer",
		Description:       "Ranger infusing arrows with arcane energy.",
		StartingHP:        72.5,
		StartingMana:      82.5,
		StartingAttack:    12.0,
		StartingDefense:   8.0,
		StartingSpeed:     10.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassArcaneArcher),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassArcaneArcher),
	}

	g.presets[engine.ClassShadowPriest] = ClassPreset{
		Type:              engine.ClassShadowPriest,
		Name:              "Shadow Priest",
		Description:       "Stealthy cleric wielding shadow magic and forbidden knowledge.",
		StartingHP:        67.5,
		StartingMana:      80.0,
		StartingAttack:    14.0,
		StartingDefense:   7.0,
		StartingSpeed:     11.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassShadowPriest),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassShadowPriest),
	}

	g.presets[engine.ClassDruid] = ClassPreset{
		Type:              engine.ClassDruid,
		Name:              "Druid",
		Description:       "Nature guardian shapeshifting between forms.",
		StartingHP:        72.5,
		StartingMana:      82.5,
		StartingAttack:    12.0,
		StartingDefense:   8.0,
		StartingSpeed:     10.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassDruid),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassDruid),
	}

	g.presets[engine.ClassInquisitor] = ClassPreset{
		Type:              engine.ClassInquisitor,
		Name:              "Inquisitor",
		Description:       "Holy investigator rooting out corruption with divine judgment.",
		StartingHP:        80.0,
		StartingMana:      75.0,
		StartingAttack:    13.0,
		StartingDefense:   9.0,
		StartingSpeed:     11.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassInquisitor),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassInquisitor),
	}

	g.presets[engine.ClassBloodKnight] = ClassPreset{
		Type:              engine.ClassBloodKnight,
		Name:              "Blood Knight",
		Description:       "Warrior sacrificing health for devastating blood magic attacks.",
		StartingHP:        82.5,
		StartingMana:      70.0,
		StartingAttack:    13.0,
		StartingDefense:   9.5,
		StartingSpeed:     6.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassBloodKnight),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassBloodKnight),
	}

	g.presets[engine.ClassMystic] = ClassPreset{
		Type:              engine.ClassMystic,
		Name:              "Mystic",
		Description:       "Enlightened caster balancing arcane and divine magic.",
		StartingHP:        75.0,
		StartingMana:      110.0,
		StartingAttack:    11.0,
		StartingDefense:   8.0,
		StartingSpeed:     8.0,
		StartingAbilities: engine.GetClassAbilities(engine.ClassMystic),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassMystic),
	}

	g.presets[engine.ClassWarlock] = ClassPreset{
		Type:              engine.ClassWarlock,
		Name:              "Warlock",
		Description:       "Pact-bound mage wielding eldritch powers.",
		StartingHP:        62.5,
		StartingMana:      115.0,
		StartingAttack:    10.5,
		StartingDefense:   6.5,
		StartingSpeed:     7.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassWarlock),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassWarlock),
	}

	g.presets[engine.ClassNinja] = ClassPreset{
		Type:              engine.ClassNinja,
		Name:              "Ninja",
		Description:       "Master assassin combining stealth with precise strikes.",
		StartingHP:        77.5,
		StartingMana:      55.0,
		StartingAttack:    16.0,
		StartingDefense:   9.0,
		StartingSpeed:     13.5,
		StartingAbilities: engine.GetClassAbilities(engine.ClassNinja),
		Specializations:   engine.GetAvailableSpecializations(engine.ClassNinja),
	}
}

// Generate creates a class configuration with optional variation.
func (g *ClassGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Determine class type from params or random
	var classType engine.CharacterClass
	if classTypeRaw, ok := params.Custom["class_type"]; ok {
		if ct, ok := classTypeRaw.(engine.CharacterClass); ok {
			classType = ct
		} else {
			classType = engine.CharacterClass(rng.Intn(21)) // 21 total classes (6 base + 15 hybrid)
		}
	} else {
		classType = engine.CharacterClass(rng.Intn(21)) // 21 total classes
	}

	// Get base preset
	preset, ok := g.presets[classType]
	if !ok {
		return nil, fmt.Errorf("invalid class type: %d", classType)
	}

	// Apply difficulty-based variation (±10% for stats)
	variation := 1.0 + (rng.Float64()-0.5)*0.2*params.Difficulty

	result := ClassPreset{
		Type:              preset.Type,
		Name:              preset.Name,
		Description:       preset.Description,
		StartingHP:        preset.StartingHP * variation,
		StartingMana:      preset.StartingMana * variation,
		StartingAttack:    preset.StartingAttack * variation,
		StartingDefense:   preset.StartingDefense * variation,
		StartingSpeed:     preset.StartingSpeed * variation,
		StartingAbilities: preset.StartingAbilities,
		Specializations:   preset.Specializations,
	}

	return &result, nil
}

// Validate checks if a generated class configuration is valid.
func (g *ClassGenerator) Validate(result interface{}) error {
	preset, ok := result.(*ClassPreset)
	if !ok {
		return fmt.Errorf("invalid result type: expected *ClassPreset")
	}

	if preset.StartingHP <= 0 {
		return fmt.Errorf("invalid starting HP: %f", preset.StartingHP)
	}

	if preset.StartingMana < 0 {
		return fmt.Errorf("invalid starting mana: %f", preset.StartingMana)
	}

	if preset.StartingAttack <= 0 {
		return fmt.Errorf("invalid starting attack: %f", preset.StartingAttack)
	}

	if len(preset.StartingAbilities) == 0 {
		return fmt.Errorf("class must have at least one starting ability")
	}

	if len(preset.Specializations) == 0 {
		return fmt.Errorf("class must have at least one specialization option")
	}

	return nil
}

// GetPreset returns a base class preset by type.
func (g *ClassGenerator) GetPreset(classType engine.CharacterClass) (ClassPreset, bool) {
	preset, ok := g.presets[classType]
	return preset, ok
}

// GetAllPresets returns all available class presets.
func (g *ClassGenerator) GetAllPresets() []ClassPreset {
	presets := make([]ClassPreset, 0, len(g.presets))
	for i := 0; i < 21; i++ { // 21 total classes (6 base + 15 hybrid)
		if preset, ok := g.presets[engine.CharacterClass(i)]; ok {
			presets = append(presets, preset)
		}
	}
	return presets
}
