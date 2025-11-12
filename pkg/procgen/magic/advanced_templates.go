package magic

// Advanced spell effect templates for Phase 24 expansion.
// These templates define spells with the 10 new effect types:
// TerrainManipulation, Transmutation, Summoning, Illusion, TimeManipulation,
// GravityControl, ElementalFusion, LifeDrain, Teleportation, Metamagic.

// GetAdvancedOffensiveTemplates returns offensive spell templates with advanced effects.
func GetAdvancedOffensiveTemplates() []SpellTemplate {
	return []SpellTemplate{
		// Elemental Fusion spells
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementFire, // Fire+Ice fusion
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Steam", "Vapor", "Mist", "Thermal", "Boiling"},
			NameSuffixes:  []string{"Explosion", "Blast", "Nova", "Eruption", "Burst"},
			Tags:          []string{"fusion", "fire", "ice", "area"},
			DamageRange:   [2]int{40, 100},
			ManaCostRange: [2]int{40, 80},
			CooldownRange: [2]float64{8.0, 15.0},
			CastTimeRange: [2]float64{1.5, 3.0},
			RangeRange:    [2]float64{10.0, 25.0},
			AreaSizeRange: [2]float64{8.0, 15.0},
		},
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementEarth, // Earth+Lightning fusion
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Glass", "Crystal", "Molten", "Vitrify", "Shard"},
			NameSuffixes:  []string{"Storm", "Rain", "Cascade", "Barrage", "Torrent"},
			Tags:          []string{"fusion", "earth", "lightning", "area"},
			DamageRange:   [2]int{45, 95},
			ManaCostRange: [2]int{35, 75},
			CooldownRange: [2]float64{7.0, 14.0},
			CastTimeRange: [2]float64{1.2, 2.5},
			RangeRange:    [2]float64{12.0, 28.0},
			AreaSizeRange: [2]float64{7.0, 14.0},
		},
		// Life Drain spells
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementDark,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Vampiric", "Soul", "Life", "Essence", "Vital"},
			NameSuffixes:  []string{"Drain", "Siphon", "Leech", "Tap", "Extraction"},
			Tags:          []string{"lifedrain", "dark", "heal"},
			DamageRange:   [2]int{30, 70},
			ManaCostRange: [2]int{25, 50},
			CooldownRange: [2]float64{5.0, 10.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			RangeRange:    [2]float64{8.0, 20.0},
			DurationRange: [2]float64{3.0, 6.0}, // Duration of drain effect
		},
	}
}

// GetAdvancedUtilityTemplates returns utility spell templates with advanced effects.
func GetAdvancedUtilityTemplates() []SpellTemplate {
	return []SpellTemplate{
		// Teleportation spells
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Blink", "Teleport", "Phase", "Warp", "Flash"},
			NameSuffixes:  []string{"Step", "Shift", "Jaunt", "Jump", "Dash"},
			Tags:          []string{"teleportation", "mobility", "instant"},
			ManaCostRange: [2]int{20, 40},
			CooldownRange: [2]float64{3.0, 8.0},
			CastTimeRange: [2]float64{0.1, 0.5},
			RangeRange:    [2]float64{10.0, 30.0}, // Teleport distance
		},
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Portal", "Gateway", "Rift", "Passage", "Dimensional"},
			NameSuffixes:  []string{"Door", "Gate", "Tunnel", "Bridge", "Link"},
			Tags:          []string{"teleportation", "long-range", "area"},
			ManaCostRange: [2]int{50, 100},
			CooldownRange: [2]float64{30.0, 60.0},
			CastTimeRange: [2]float64{2.0, 5.0},
			RangeRange:    [2]float64{50.0, 200.0}, // Long-range portal
			DurationRange: [2]float64{10.0, 30.0},  // Portal duration
		},
		// Illusion spells
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementLight,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Invisible", "Vanish", "Fade", "Ghost", "Shadow"},
			NameSuffixes:  []string{"Cloak", "Veil", "Shroud", "Form", "Walk"},
			Tags:          []string{"illusion", "invisibility", "stealth"},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{15.0, 30.0},
			CastTimeRange: [2]float64{1.0, 2.0},
			DurationRange: [2]float64{10.0, 30.0},
		},
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Mirror", "Phantom", "Decoy", "Clone", "Duplicate"},
			NameSuffixes:  []string{"Image", "Self", "Copy", "Double", "Illusion"},
			Tags:          []string{"illusion", "decoy", "confusion"},
			ManaCostRange: [2]int{25, 50},
			CooldownRange: [2]float64{10.0, 20.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			DurationRange: [2]float64{15.0, 45.0},
		},
		// Terrain Manipulation spells
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementEarth,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Stone", "Earth", "Rock", "Granite", "Wall"},
			NameSuffixes:  []string{"Wall", "Barrier", "Rampart", "Bulwark", "Fortress"},
			Tags:          []string{"terrain", "earth", "defense"},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{10.0, 20.0},
			CastTimeRange: [2]float64{1.5, 3.0},
			RangeRange:    [2]float64{5.0, 15.0},
			AreaSizeRange: [2]float64{3.0, 8.0},
			DurationRange: [2]float64{30.0, 120.0}, // Wall duration
		},
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementEarth,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Earth", "Ground", "Tremor", "Quake", "Fissure"},
			NameSuffixes:  []string{"Shaping", "Mold", "Pit", "Chasm", "Crater"},
			Tags:          []string{"terrain", "earth", "manipulation"},
			ManaCostRange: [2]int{35, 70},
			CooldownRange: [2]float64{12.0, 25.0},
			CastTimeRange: [2]float64{2.0, 4.0},
			RangeRange:    [2]float64{10.0, 25.0},
			AreaSizeRange: [2]float64{5.0, 12.0},
		},
		// Transmutation spells
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Transmute", "Alter", "Transform", "Convert", "Change"},
			NameSuffixes:  []string{"Stone", "Gold", "Metal", "Material", "Form"},
			Tags:          []string{"transmutation", "alchemy", "material"},
			ManaCostRange: [2]int{40, 80},
			CooldownRange: [2]float64{20.0, 40.0},
			CastTimeRange: [2]float64{2.5, 5.0},
			RangeRange:    [2]float64{5.0, 15.0},
			AreaSizeRange: [2]float64{2.0, 6.0},
		},
	}
}

// GetAdvancedSupportTemplates returns support spell templates with advanced effects.
func GetAdvancedSupportTemplates() []SpellTemplate {
	return []SpellTemplate{
		// Time Manipulation spells
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Time", "Temporal", "Chrono", "Haste", "Quicken"},
			NameSuffixes:  []string{"Acceleration", "Warp", "Rush", "Speed", "Boost"},
			Tags:          []string{"time", "haste", "speed"},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{20.0, 40.0},
			CastTimeRange: [2]float64{1.0, 2.0},
			DurationRange: [2]float64{10.0, 30.0},
		},
		{
			BaseType:      TypeDebuff,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Time", "Temporal", "Chrono", "Slow", "Stasis"},
			NameSuffixes:  []string{"Dilation", "Field", "Freeze", "Stop", "Lock"},
			Tags:          []string{"time", "slow", "debuff"},
			ManaCostRange: [2]int{35, 70},
			CooldownRange: [2]float64{15.0, 30.0},
			CastTimeRange: [2]float64{1.2, 2.5},
			RangeRange:    [2]float64{10.0, 25.0},
			AreaSizeRange: [2]float64{5.0, 12.0},
			DurationRange: [2]float64{5.0, 15.0},
		},
		// Gravity Control spells
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementWind,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Levitate", "Float", "Hover", "Fly", "Feather"},
			NameSuffixes:  []string{"Form", "Fall", "Flight", "Grace", "Wings"},
			Tags:          []string{"gravity", "levitation", "mobility"},
			ManaCostRange: [2]int{25, 50},
			CooldownRange: [2]float64{15.0, 30.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			DurationRange: [2]float64{20.0, 60.0},
		},
		{
			BaseType:      TypeDebuff,
			BaseElement:   ElementEarth,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Gravity", "Weight", "Crush", "Heavy", "Burden"},
			NameSuffixes:  []string{"Increase", "Amplification", "Field", "Well", "Sink"},
			Tags:          []string{"gravity", "weight", "slow"},
			DamageRange:   [2]int{10, 30}, // Crushing damage
			ManaCostRange: [2]int{20, 40},
			CooldownRange: [2]float64{8.0, 16.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			RangeRange:    [2]float64{10.0, 25.0},
			DurationRange: [2]float64{5.0, 15.0},
		},
		// Summoning spells
		{
			BaseType:      TypeSummon,
			BaseElement:   ElementFire,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Summon", "Call", "Conjure", "Invoke", "Manifest"},
			NameSuffixes:  []string{"Elemental", "Spirit", "Guardian", "Servant", "Familiar"},
			Tags:          []string{"summoning", "elemental", "ally"},
			ManaCostRange: [2]int{50, 100},
			CooldownRange: [2]float64{30.0, 60.0},
			CastTimeRange: [2]float64{2.0, 4.0},
			RangeRange:    [2]float64{5.0, 15.0},
			DurationRange: [2]float64{30.0, 120.0}, // Summon duration
		},
		// Metamagic spells
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementArcane,
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Empower", "Amplify", "Intensify", "Maximize", "Augment"},
			NameSuffixes:  []string{"Magic", "Spell", "Power", "Force", "Might"},
			Tags:          []string{"metamagic", "amplify", "power"},
			ManaCostRange: [2]int{40, 80},
			CooldownRange: [2]float64{25.0, 50.0},
			CastTimeRange: [2]float64{1.5, 3.0},
			DurationRange: [2]float64{10.0, 30.0},
		},
	}
}

// GetSciFiAdvancedTemplates returns sci-fi genre templates for advanced effects.
func GetSciFiAdvancedTemplates() []SpellTemplate {
	return []SpellTemplate{
		// Sci-Fi Teleportation
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementLightning, // Tech energy
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Quantum", "Phase", "Dimensional", "Warp", "Spatial"},
			NameSuffixes:  []string{"Shift", "Jump", "Transport", "Displacement", "Hop"},
			Tags:          []string{"teleportation", "tech", "instant"},
			ManaCostRange: [2]int{25, 50},
			CooldownRange: [2]float64{4.0, 10.0},
			CastTimeRange: [2]float64{0.2, 0.8},
			RangeRange:    [2]float64{15.0, 40.0},
		},
		// Sci-Fi Time Manipulation
		{
			BaseType:      TypeBuff,
			BaseElement:   ElementArcane, // Advanced tech
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Overclock", "Turbo", "Nano", "Neural", "Reflex"},
			NameSuffixes:  []string{"Boost", "Acceleration", "Enhancement", "Overdrive", "Mode"},
			Tags:          []string{"time", "tech", "speed"},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{20.0, 40.0},
			CastTimeRange: [2]float64{0.8, 1.5},
			DurationRange: [2]float64{8.0, 25.0},
		},
		// Sci-Fi Gravity Control
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementWind, // Gravity tech
			BaseTarget:    TargetSelf,
			NamePrefixes:  []string{"Anti-Gravity", "Grav", "Repulsor", "Hover", "Zero-G"},
			NameSuffixes:  []string{"Field", "Drive", "System", "Pack", "Assist"},
			Tags:          []string{"gravity", "tech", "mobility"},
			ManaCostRange: [2]int{20, 45},
			CooldownRange: [2]float64{12.0, 25.0},
			CastTimeRange: [2]float64{0.5, 1.2},
			DurationRange: [2]float64{15.0, 45.0},
		},
	}
}

// GetHorrorAdvancedTemplates returns horror genre templates for advanced effects.
func GetHorrorAdvancedTemplates() []SpellTemplate {
	return []SpellTemplate{
		// Horror Life Drain
		{
			BaseType:      TypeOffensive,
			BaseElement:   ElementDark,
			BaseTarget:    TargetSingle,
			NamePrefixes:  []string{"Necrotic", "Withering", "Decay", "Blight", "Corruption"},
			NameSuffixes:  []string{"Touch", "Drain", "Grasp", "Embrace", "Curse"},
			Tags:          []string{"lifedrain", "dark", "horror"},
			DamageRange:   [2]int{35, 75},
			ManaCostRange: [2]int{30, 60},
			CooldownRange: [2]float64{6.0, 12.0},
			CastTimeRange: [2]float64{1.0, 2.0},
			RangeRange:    [2]float64{5.0, 15.0},
			DurationRange: [2]float64{4.0, 8.0},
		},
		// Horror Illusion
		{
			BaseType:      TypeUtility,
			BaseElement:   ElementDark,
			BaseTarget:    TargetArea,
			NamePrefixes:  []string{"Nightmare", "Terror", "Madness", "Fear", "Dread"},
			NameSuffixes:  []string{"Visions", "Phantoms", "Spectres", "Hallucination", "Mirage"},
			Tags:          []string{"illusion", "horror", "fear"},
			ManaCostRange: [2]int{35, 70},
			CooldownRange: [2]float64{15.0, 30.0},
			CastTimeRange: [2]float64{1.5, 3.0},
			RangeRange:    [2]float64{10.0, 25.0},
			AreaSizeRange: [2]float64{8.0, 15.0},
			DurationRange: [2]float64{10.0, 30.0},
		},
	}
}
