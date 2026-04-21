package class

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// ClassPreset defines a character class configuration.
// It contains all the starting stats and abilities for a character class.
type ClassPreset struct {
	// Type is the character class enum value (e.g., ClassWarrior, ClassMage)
	Type engine.CharacterClass
	// Name is the display name of the class (e.g., "Warrior", "Mage")
	Name string
	// Description is a short text describing the class playstyle
	Description string
	// StartingHP is the initial health points for this class
	StartingHP float64
	// StartingMana is the initial mana/energy pool for this class
	StartingMana float64
	// StartingAttack is the base attack power for this class
	StartingAttack float64
	// StartingDefense is the base defense rating for this class
	StartingDefense float64
	// StartingSpeed is the base movement/action speed for this class
	StartingSpeed float64
	// StartingAbilities is the list of ability IDs available at level 1
	StartingAbilities []string
	// Specializations are the advanced class paths available for this class
	Specializations []engine.SpecializationType
}

// genreTheming holds genre-specific class name/description mappings.
type genreTheming struct {
	name        string
	description string
}

// ClassGenerator generates character class configurations.
type ClassGenerator struct {
	// presets stores the base class configurations indexed by class type.
	// Contains all 21 classes (6 base + 15 hybrid) initialized at construction.
	presets map[engine.CharacterClass]ClassPreset
	// genreThemes stores genre-specific name/description overrides.
	// Key format: "genreID:classType" (e.g., "scifi:0" for Warrior in sci-fi).
	genreThemes map[string]genreTheming
	// logger is the logger instance used for error and debug logging.
	// If nil, the package-level logrus logger is used.
	logger *logrus.Entry
}

// NewClassGenerator creates a new class generator with default logging.
func NewClassGenerator() *ClassGenerator {
	gen := &ClassGenerator{
		presets:     make(map[engine.CharacterClass]ClassPreset),
		genreThemes: make(map[string]genreTheming),
		logger:      logrus.WithField("system_name", "class_generator"),
	}
	gen.initializePresets()
	gen.initializeGenreThemes()
	return gen
}

// NewClassGeneratorWithLogger creates a new class generator with a custom logger.
// The provided logger entry is used for all logging operations, enabling
// integration with custom logging pipelines and structured logging contexts.
func NewClassGeneratorWithLogger(logger *logrus.Entry) *ClassGenerator {
	gen := &ClassGenerator{
		presets:     make(map[engine.CharacterClass]ClassPreset),
		genreThemes: make(map[string]genreTheming),
		logger:      logger,
	}
	if gen.logger == nil {
		gen.logger = logrus.WithField("system_name", "class_generator")
	}
	gen.initializePresets()
	gen.initializeGenreThemes()
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

// initializeGenreThemes sets up genre-specific class name/description mappings.
func (g *ClassGenerator) initializeGenreThemes() {
	// Sci-Fi genre mappings
	g.addGenreTheme("scifi", engine.ClassWarrior, "Shock Trooper", "Elite combat specialist with powered armor and energy weapons.")
	g.addGenreTheme("scifi", engine.ClassRogue, "Infiltrator", "Stealth operative using cloaking tech and precision strikes.")
	g.addGenreTheme("scifi", engine.ClassMage, "Psionic", "Psychic warrior wielding telekinetic and telepathic powers.")
	g.addGenreTheme("scifi", engine.ClassCleric, "Medic", "Field surgeon with advanced healing nanobots and support drones.")
	g.addGenreTheme("scifi", engine.ClassRanger, "Scout", "Recon specialist with tactical sensors and long-range weaponry.")
	g.addGenreTheme("scifi", engine.ClassPaladin, "Vanguard", "Frontline defender in powered exo-armor with energy shields.")

	// Horror genre mappings
	g.addGenreTheme("horror", engine.ClassWarrior, "Survivor", "Hardened fighter who endures against overwhelming darkness.")
	g.addGenreTheme("horror", engine.ClassRogue, "Stalker", "Silent predator who hunts from the shadows.")
	g.addGenreTheme("horror", engine.ClassMage, "Occultist", "Dark ritualist channeling forbidden and cursed energies.")
	g.addGenreTheme("horror", engine.ClassCleric, "Exorcist", "Holy warrior banishing undead and cleansing corruption.")
	g.addGenreTheme("horror", engine.ClassRanger, "Hunter", "Monster slayer tracking supernatural threats.")
	g.addGenreTheme("horror", engine.ClassPaladin, "Crusader", "Divine champion standing against the forces of evil.")

	// Cyberpunk genre mappings
	g.addGenreTheme("cyberpunk", engine.ClassWarrior, "Street Samurai", "Augmented mercenary with cybernetic combat implants.")
	g.addGenreTheme("cyberpunk", engine.ClassRogue, "Netrunner", "Elite hacker infiltrating corporate systems and ICE.")
	g.addGenreTheme("cyberpunk", engine.ClassMage, "Technomancer", "Digital sorcerer manipulating cyberspace with neural interface.")
	g.addGenreTheme("cyberpunk", engine.ClassCleric, "Ripperdoc", "Black market surgeon installing illegal cyberware upgrades.")
	g.addGenreTheme("cyberpunk", engine.ClassRanger, "Drone Rigger", "Tactical operator controlling combat drones remotely.")
	g.addGenreTheme("cyberpunk", engine.ClassPaladin, "Corporate Enforcer", "Heavily augmented security operative protecting corporate assets.")

	// Post-Apocalyptic genre mappings
	g.addGenreTheme("postapoc", engine.ClassWarrior, "Raider", "Wasteland warrior scavenging for survival.")
	g.addGenreTheme("postapoc", engine.ClassRogue, "Scavenger", "Resourceful survivor finding treasures in the ruins.")
	g.addGenreTheme("postapoc", engine.ClassMage, "Mutant", "Radiation-touched individual with strange powers.")
	g.addGenreTheme("postapoc", engine.ClassCleric, "Healer", "Medic keeping communities alive in the wasteland.")
	g.addGenreTheme("postapoc", engine.ClassRanger, "Outrider", "Nomadic scout navigating the dangerous wastes.")
	g.addGenreTheme("postapoc", engine.ClassPaladin, "Protector", "Guardian defending settlements from threats.")
}

// addGenreTheme adds a genre-specific name/description override for a class.
func (g *ClassGenerator) addGenreTheme(genreID string, classType engine.CharacterClass, name, description string) {
	key := fmt.Sprintf("%s:%d", genreID, classType)
	g.genreThemes[key] = genreTheming{
		name:        name,
		description: description,
	}
}

// canonicalGenreID normalizes genre ID aliases to the canonical form used in map keys.
// The CLI flag and predefined genres use "postapoc"; some older code used "postapocalyptic".
func canonicalGenreID(genreID string) string {
	if genreID == "postapocalyptic" {
		return "postapoc"
	}
	return genreID
}

// getThemedName returns the genre-specific name for a class, or the default if no theme exists.
func (g *ClassGenerator) getThemedName(genreID string, classType engine.CharacterClass, defaultName string) string {
	if genreID == "" || genreID == "fantasy" {
		return defaultName
	}
	key := fmt.Sprintf("%s:%d", canonicalGenreID(genreID), classType)
	if theme, ok := g.genreThemes[key]; ok {
		return theme.name
	}
	return defaultName
}

// getThemedDescription returns the genre-specific description for a class, or the default if no theme exists.
func (g *ClassGenerator) getThemedDescription(genreID string, classType engine.CharacterClass, defaultDesc string) string {
	if genreID == "" || genreID == "fantasy" {
		return defaultDesc
	}
	key := fmt.Sprintf("%s:%d", canonicalGenreID(genreID), classType)
	if theme, ok := g.genreThemes[key]; ok {
		return theme.description
	}
	return defaultDesc
}

// Generate creates a class configuration with optional variation.
func (g *ClassGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	rng := rand.New(rand.NewSource(seed))

	// Determine class type from params or random selection from available presets
	var classType engine.CharacterClass
	if classTypeRaw, ok := params.Custom["class_type"]; ok {
		if ct, ok := classTypeRaw.(engine.CharacterClass); ok {
			classType = ct
		} else {
			classType = g.randomClassType(rng)
		}
	} else {
		classType = g.randomClassType(rng)
	}

	// Get base preset
	preset, ok := g.presets[classType]
	if !ok {
		g.logger.WithFields(logrus.Fields{
			"seed":       seed,
			"class_type": classType,
			"difficulty": params.Difficulty,
			"genre_id":   params.GenreID,
		}).Error("class generation failed: invalid class type")
		return nil, fmt.Errorf("invalid class type: %d", classType)
	}

	// Apply difficulty-based variation (±10% for stats)
	variation := 1.0 + (rng.Float64()-0.5)*0.2*params.Difficulty

	// Apply genre theming to name and description
	themedName := g.getThemedName(params.GenreID, preset.Type, preset.Name)
	themedDesc := g.getThemedDescription(params.GenreID, preset.Type, preset.Description)

	result := ClassPreset{
		Type:              preset.Type,
		Name:              themedName,
		Description:       themedDesc,
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

	if preset.Name == "" {
		return fmt.Errorf("class name must not be empty")
	}

	if preset.Description == "" {
		return fmt.Errorf("class description must not be empty")
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

	if preset.StartingDefense <= 0 {
		return fmt.Errorf("invalid starting defense: %f", preset.StartingDefense)
	}

	if preset.StartingSpeed <= 0 {
		return fmt.Errorf("invalid starting speed: %f", preset.StartingSpeed)
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

// randomClassType selects a random class from the available presets.
func (g *ClassGenerator) randomClassType(rng *rand.Rand) engine.CharacterClass {
	keys := make([]engine.CharacterClass, 0, len(g.presets))
	for k := range g.presets {
		keys = append(keys, k)
	}
	return keys[rng.Intn(len(keys))]
}

// GetAllPresets returns all available class presets in enum order.
func (g *ClassGenerator) GetAllPresets() []ClassPreset {
	presets := make([]ClassPreset, 0, len(g.presets))

	// Iterate through all possible enum values up to a reasonable maximum
	// This handles potential gaps in the enum sequence
	const maxClassEnum = 100 // Safety limit to prevent infinite loops
	foundCount := 0

	for i := 0; i < maxClassEnum && foundCount < len(g.presets); i++ {
		if preset, ok := g.presets[engine.CharacterClass(i)]; ok {
			presets = append(presets, preset)
			foundCount++
		}
	}

	// Log warning if we didn't find all expected presets
	if foundCount < len(g.presets) {
		g.logger.WithFields(logrus.Fields{
			"expected": len(g.presets),
			"found":    foundCount,
		}).Warn("GetAllPresets did not find all registered presets - enum may have gaps")
	}

	return presets
}
