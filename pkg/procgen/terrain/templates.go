// Package terrain provides architectural templates for themed dungeon generation.
// This file defines genre-specific room templates and environmental storytelling elements.
package terrain

import (
	"math/rand"
)

// RoomTemplate defines the architectural style and theme for a room.
type RoomTemplate struct {
	Name        string   // Template name
	Genre       string   // Genre (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
	RoomType    RoomType // Room functional type
	Description string   // Flavor text for environmental storytelling
	TileTheme   string   // Dominant tile theme
	Decorations []string // Environmental elements
	Lighting    string   // Lighting style (bright, dim, flickering, dark)
	Ambience    string   // Audio/atmosphere description
}

// TemplateLibrary stores all available room templates organized by genre.
type TemplateLibrary struct {
	templates map[string]map[RoomType][]RoomTemplate // genre -> room type -> templates
	rng       *rand.Rand
}

// NewTemplateLibrary creates a new template library with predefined templates.
func NewTemplateLibrary(seed int64) *TemplateLibrary {
	lib := &TemplateLibrary{
		templates: make(map[string]map[RoomType][]RoomTemplate),
		rng:       rand.New(rand.NewSource(seed)),
	}
	lib.initializeTemplates()
	return lib
}

// GetTemplate returns a random template for the given genre and room type.
func (lib *TemplateLibrary) GetTemplate(genre string, roomType RoomType) *RoomTemplate {
	genreTemplates, ok := lib.templates[genre]
	if !ok {
		// Fallback to fantasy if genre not found
		genreTemplates = lib.templates["fantasy"]
	}

	templates, ok := genreTemplates[roomType]
	if !ok || len(templates) == 0 {
		// Return a default template
		return &RoomTemplate{
			Name:        "Generic Room",
			Genre:       genre,
			RoomType:    roomType,
			Description: "A nondescript chamber.",
			TileTheme:   "stone",
			Lighting:    "dim",
		}
	}

	// Select random template
	idx := lib.rng.Intn(len(templates))
	template := templates[idx]
	return &template
}

// initializeTemplates populates the library with predefined templates.
func (lib *TemplateLibrary) initializeTemplates() {
	lib.initializeFantasyTemplates()
	lib.initializeSciFiTemplates()
	lib.initializeHorrorTemplates()
	lib.initializeCyberpunkTemplates()
	lib.initializePostApocalypticTemplates()
}

// initializeFantasyTemplates creates fantasy-themed room templates.
func (lib *TemplateLibrary) initializeFantasyTemplates() {
	genre := "fantasy"
	lib.templates[genre] = make(map[RoomType][]RoomTemplate)

	// Start rooms
	lib.templates[genre][RoomSpawn] = []RoomTemplate{
		{
			Name:        "Grand Entrance Hall",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "A majestic entrance hall with vaulted ceilings and ancient banners. Torch brackets line the walls.",
			TileTheme:   "polished_stone",
			Decorations: []string{"banners", "torches", "pillars", "grand_door"},
			Lighting:    "bright",
			Ambience:    "echoing footsteps, distant dripping water",
		},
		{
			Name:        "Ruined Courtyard",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "An overgrown courtyard with crumbling statues. Ivy climbs the walls and moss covers the stones.",
			TileTheme:   "weathered_stone",
			Decorations: []string{"statues", "ivy", "moss", "broken_fountain"},
			Lighting:    "natural",
			Ambience:    "birdsong, rustling leaves",
		},
	}

	// Combat rooms
	lib.templates[genre][RoomCombat] = []RoomTemplate{
		{
			Name:        "Armory Chamber",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Weapon racks line the walls, most empty. Scorch marks suggest recent battles.",
			TileTheme:   "stone_floor",
			Decorations: []string{"weapon_racks", "armor_stands", "scorch_marks", "blood_stains"},
			Lighting:    "dim",
			Ambience:    "clinking metal, distant combat sounds",
		},
		{
			Name:        "Guard Barracks",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Rows of bunks and overturned tables. The guards didn't go down without a fight.",
			TileTheme:   "wooden_planks",
			Decorations: []string{"bunks", "tables", "scattered_weapons", "torn_banners"},
			Lighting:    "dim",
			Ambience:    "creaking wood, wind through broken windows",
		},
	}

	// Treasure rooms
	lib.templates[genre][RoomTreasure] = []RoomTemplate{
		{
			Name:        "Royal Treasury",
			Genre:       genre,
			RoomType:    RoomTreasure,
			Description: "Gilded pillars support a high ceiling. Ancient chests overflow with treasures.",
			TileTheme:   "marble",
			Decorations: []string{"chests", "gold_piles", "jeweled_statues", "treasure_vault"},
			Lighting:    "bright",
			Ambience:    "echoing silence, glittering reflections",
		},
		{
			Name:        "Dragon Hoard",
			Genre:       genre,
			RoomType:    RoomTreasure,
			Description: "Mounds of coins and gems surround a massive nest. The air smells of sulfur.",
			TileTheme:   "volcanic_rock",
			Decorations: []string{"gold_mounds", "dragon_nest", "burnt_bones", "scorched_treasure"},
			Lighting:    "flickering",
			Ambience:    "heat shimmer, distant roaring",
		},
	}

	// Puzzle rooms
	lib.templates[genre][RoomPuzzle] = []RoomTemplate{
		{
			Name:        "Ancient Library",
			Genre:       genre,
			RoomType:    RoomPuzzle,
			Description: "Towering bookshelves hold forgotten knowledge. A mystical sigil glows on the floor.",
			TileTheme:   "stone",
			Decorations: []string{"bookshelves", "reading_desks", "magical_sigil", "ancient_tomes"},
			Lighting:    "magical_glow",
			Ambience:    "rustling pages, arcane whispers",
		},
		{
			Name:        "Alchemist Laboratory",
			Genre:       genre,
			RoomType:    RoomPuzzle,
			Description: "Bubbling potions and complex apparatus fill the room. A riddle is etched on the wall.",
			TileTheme:   "stone",
			Decorations: []string{"potion_stands", "alchemical_equipment", "riddle_wall", "ingredient_jars"},
			Lighting:    "dim",
			Ambience:    "bubbling liquids, chemical smells",
		},
	}

	// Boss rooms
	lib.templates[genre][RoomBoss] = []RoomTemplate{
		{
			Name:        "Throne Room",
			Genre:       genre,
			RoomType:    RoomBoss,
			Description: "A massive throne dominates the chamber. Dark energy pulses from the ceiling.",
			TileTheme:   "obsidian",
			Decorations: []string{"dark_throne", "skulls", "dark_crystals", "evil_banners"},
			Lighting:    "dark",
			Ambience:    "ominous humming, distant screams",
		},
	}

	// Shop rooms
	lib.templates[genre][RoomShop] = []RoomTemplate{
		{
			Name:        "Merchant's Stall",
			Genre:       genre,
			RoomType:    RoomShop,
			Description: "A well-lit stall displays various wares. The merchant smiles invitingly.",
			TileTheme:   "wooden_planks",
			Decorations: []string{"display_tables", "merchant_cart", "price_signs", "trade_goods"},
			Lighting:    "bright",
			Ambience:    "merchant chatter, coin jingling",
		},
	}

	// Rest rooms
	lib.templates[genre][RoomRest] = []RoomTemplate{
		{
			Name:        "Cozy Campfire",
			Genre:       genre,
			RoomType:    RoomRest,
			Description: "A warm campfire crackles in the center. Bedrolls are arranged around it.",
			TileTheme:   "dirt",
			Decorations: []string{"campfire", "bedrolls", "supply_crates", "cooking_pot"},
			Lighting:    "warm",
			Ambience:    "crackling fire, peaceful silence",
		},
	}

	// Secret rooms
	lib.templates[genre][RoomSecret] = []RoomTemplate{
		{
			Name:        "Hidden Vault",
			Genre:       genre,
			RoomType:    RoomSecret,
			Description: "A concealed chamber holds ancient artifacts. Dust covers everything.",
			TileTheme:   "ancient_stone",
			Decorations: []string{"artifact_pedestals", "dust_layers", "cobwebs", "rune_circles"},
			Lighting:    "dim",
			Ambience:    "absolute silence, mystical energy",
		},
	}
}

// initializeSciFiTemplates creates sci-fi-themed room templates.
func (lib *TemplateLibrary) initializeSciFiTemplates() {
	genre := "sci-fi"
	lib.templates[genre] = make(map[RoomType][]RoomTemplate)

	lib.templates[genre][RoomSpawn] = []RoomTemplate{
		{
			Name:        "Docking Bay",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "A large bay with inactive loading equipment. Emergency lights pulse red.",
			TileTheme:   "metal_grating",
			Decorations: []string{"loading_cranes", "cargo_containers", "warning_lights", "airlocks"},
			Lighting:    "emergency_red",
			Ambience:    "alarm klaxons, hissing steam",
		},
	}

	lib.templates[genre][RoomCombat] = []RoomTemplate{
		{
			Name:        "Security Station",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Monitors display static. Blast marks scar the walls.",
			TileTheme:   "reinforced_metal",
			Decorations: []string{"security_monitors", "blast_doors", "weapon_lockers", "energy_shields"},
			Lighting:    "bright",
			Ambience:    "electronic beeps, power surges",
		},
	}

	lib.templates[genre][RoomTreasure] = []RoomTemplate{
		{
			Name:        "Research Archive",
			Genre:       genre,
			RoomType:    RoomTreasure,
			Description: "Data storage units hold valuable information. Rare prototypes sit on shelves.",
			TileTheme:   "clean_metal",
			Decorations: []string{"data_cores", "prototype_displays", "research_notes", "tech_samples"},
			Lighting:    "bright",
			Ambience:    "humming servers, cooling fans",
		},
	}

	lib.templates[genre][RoomBoss] = []RoomTemplate{
		{
			Name:        "Command Center",
			Genre:       genre,
			RoomType:    RoomBoss,
			Description: "A massive holographic display shows tactical data. The AI core pulses ominously.",
			TileTheme:   "pristine_metal",
			Decorations: []string{"holo_displays", "AI_core", "command_throne", "tactical_screens"},
			Lighting:    "holographic_glow",
			Ambience:    "synthetic voice, power thrumming",
		},
	}
}

// initializeHorrorTemplates creates horror-themed room templates.
func (lib *TemplateLibrary) initializeHorrorTemplates() {
	genre := "horror"
	lib.templates[genre] = make(map[RoomType][]RoomTemplate)

	lib.templates[genre][RoomSpawn] = []RoomTemplate{
		{
			Name:        "Abandoned Reception",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "A decaying reception desk sits amid overturned chairs. Blood trails lead deeper inside.",
			TileTheme:   "stained_tiles",
			Decorations: []string{"broken_furniture", "blood_trails", "scattered_papers", "flickering_light"},
			Lighting:    "flickering",
			Ambience:    "distant screams, dripping blood",
		},
	}

	lib.templates[genre][RoomCombat] = []RoomTemplate{
		{
			Name:        "Surgical Theater",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Rusted surgical tools litter blood-stained tables. The smell is overwhelming.",
			TileTheme:   "grimy_tiles",
			Decorations: []string{"operating_tables", "surgical_tools", "blood_stains", "body_parts"},
			Lighting:    "dim",
			Ambience:    "dripping liquids, distant moaning",
		},
	}

	lib.templates[genre][RoomTreasure] = []RoomTemplate{
		{
			Name:        "Cursed Vault",
			Genre:       genre,
			RoomType:    RoomTreasure,
			Description: "Ancient treasures sit behind a sealed barrier. Ghostly whispers echo around.",
			TileTheme:   "ancient_stone",
			Decorations: []string{"cursed_treasures", "ghost_wisps", "ancient_seals", "death_masks"},
			Lighting:    "supernatural_glow",
			Ambience:    "ghostly whispers, wailing spirits",
		},
	}

	lib.templates[genre][RoomBoss] = []RoomTemplate{
		{
			Name:        "Ritual Chamber",
			Genre:       genre,
			RoomType:    RoomBoss,
			Description: "A pentagram carved in the floor glows with unholy light. The air is thick with dread.",
			TileTheme:   "blood_stone",
			Decorations: []string{"pentagram", "sacrificial_altar", "candles", "dark_artifacts"},
			Lighting:    "dark",
			Ambience:    "chanting, demonic presence",
		},
	}
}

// initializeCyberpunkTemplates creates cyberpunk-themed room templates.
func (lib *TemplateLibrary) initializeCyberpunkTemplates() {
	genre := "cyberpunk"
	lib.templates[genre] = make(map[RoomType][]RoomTemplate)

	lib.templates[genre][RoomSpawn] = []RoomTemplate{
		{
			Name:        "Corporate Lobby",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "Neon advertisements flicker on the walls. Security drones patrol overhead.",
			TileTheme:   "polished_concrete",
			Decorations: []string{"neon_signs", "security_drones", "corporate_logos", "holo_ads"},
			Lighting:    "neon_glow",
			Ambience:    "electronic buzz, corporate jingles",
		},
	}

	lib.templates[genre][RoomCombat] = []RoomTemplate{
		{
			Name:        "Server Farm",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Rows of server racks hum with processing power. Security turrets track movement.",
			TileTheme:   "raised_floor",
			Decorations: []string{"server_racks", "security_turrets", "cable_bundles", "cooling_vents"},
			Lighting:    "bright",
			Ambience:    "humming servers, cooling fans",
		},
	}

	lib.templates[genre][RoomShop] = []RoomTemplate{
		{
			Name:        "Black Market Stall",
			Genre:       genre,
			RoomType:    RoomShop,
			Description: "Illegal tech and cybernetic enhancements are on display. The dealer watches carefully.",
			TileTheme:   "dirty_metal",
			Decorations: []string{"tech_displays", "cyber_implants", "stolen_goods", "holo_prices"},
			Lighting:    "dim",
			Ambience:    "electronic whispers, deal-making",
		},
	}
}

// initializePostApocalypticTemplates creates post-apocalyptic-themed room templates.
func (lib *TemplateLibrary) initializePostApocalypticTemplates() {
	genre := "post-apocalyptic"
	lib.templates[genre] = make(map[RoomType][]RoomTemplate)

	lib.templates[genre][RoomSpawn] = []RoomTemplate{
		{
			Name:        "Ruined Entrance",
			Genre:       genre,
			RoomType:    RoomSpawn,
			Description: "Collapsed walls and debris block most paths. Radiation warning signs are faded.",
			TileTheme:   "rubble",
			Decorations: []string{"collapsed_walls", "debris_piles", "warning_signs", "scattered_bones"},
			Lighting:    "natural",
			Ambience:    "wind howling, creaking metal",
		},
	}

	lib.templates[genre][RoomCombat] = []RoomTemplate{
		{
			Name:        "Raider Camp",
			Genre:       genre,
			RoomType:    RoomCombat,
			Description: "Makeshift barricades surround a fire pit. Scavenged weapons lean against walls.",
			TileTheme:   "dirt_and_debris",
			Decorations: []string{"barricades", "fire_pit", "scavenged_weapons", "sleeping_bags"},
			Lighting:    "firelight",
			Ambience:    "crackling fire, hostile voices",
		},
	}

	lib.templates[genre][RoomTreasure] = []RoomTemplate{
		{
			Name:        "Supply Cache",
			Genre:       genre,
			RoomType:    RoomTreasure,
			Description: "Sealed containers hold pre-war supplies. Water and food are precious here.",
			TileTheme:   "concrete",
			Decorations: []string{"supply_containers", "water_barrels", "canned_food", "medical_kits"},
			Lighting:    "dim",
			Ambience:    "silence, dust settling",
		},
	}

	lib.templates[genre][RoomRest] = []RoomTemplate{
		{
			Name:        "Survivor Bunker",
			Genre:       genre,
			RoomType:    RoomRest,
			Description: "A reinforced bunker provides shelter. A small generator powers dim lights.",
			TileTheme:   "reinforced_concrete",
			Decorations: []string{"reinforced_door", "generator", "sleeping_quarters", "radio"},
			Lighting:    "dim",
			Ambience:    "generator hum, radio static",
		},
	}
}

// ApplyTemplateToRoom applies a template's properties to a room node.
func ApplyTemplateToRoom(room *RoomNode, template *RoomTemplate) {
	if room.Properties == nil {
		room.Properties = make(map[string]interface{})
	}
	
	room.Theme = template.Name
	room.Properties["description"] = template.Description
	room.Properties["tileTheme"] = template.TileTheme
	room.Properties["decorations"] = template.Decorations
	room.Properties["lighting"] = template.Lighting
	room.Properties["ambience"] = template.Ambience
}
