package quest

// genreQuestTemplates is the authoritative data table for all quest templates,
// keyed by genre ID. Adding a new genre requires only adding an entry here;
// no per-function edits are needed. This replaces the previous pattern of
// 20 per-genre/per-type functions.
//
// Genre keys: "fantasy", "scifi", "horror", "cyberpunk", "postapoc".
var genreQuestTemplates = map[string][]QuestTemplate{
	"fantasy": {
		{
			BaseType:     TypeKill,
			NamePrefixes: []string{"Slay", "Hunt", "Cull", "Exterminate", "Eliminate"},
			NameSuffixes: []string{"the Undead", "the Goblins", "the Bandits", "the Monsters", "the Beasts"},
			DescTemplates: []string{
				"%s have been terrorizing the area. Defeat %d of them.",
				"The local settlement is under attack by %s. Eliminate %d to protect the people.",
				"A horde of %s threatens the region. Hunt down %d of these creatures.",
			},
			Tags:             []string{"combat", "kill"},
			TargetTypes:      []string{"Goblin", "Skeleton", "Orc", "Wolf", "Bandit", "Zombie", "Spider"},
			RequiredRange:    [2]int{5, 20},
			XPRewardRange:    [2]int{50, 200},
			GoldRewardRange:  [2]int{10, 50},
			ItemRewardChance: 0.3,
		},
		{
			BaseType:     TypeCollect,
			NamePrefixes: []string{"Gather", "Collect", "Retrieve", "Find", "Acquire"},
			NameSuffixes: []string{"Herbs", "Crystals", "Artifacts", "Resources", "Components"},
			DescTemplates: []string{
				"I need %d %s for my research. Can you gather them?",
				"The town needs %d %s. Search the area and bring them back.",
				"Ancient %s are scattered throughout the region. Collect %d of them.",
			},
			Tags:             []string{"gather", "explore"},
			TargetTypes:      []string{"Moonflower", "Mana Crystal", "Ancient Rune", "Dragon Scale", "Phoenix Feather"},
			RequiredRange:    [2]int{3, 15},
			XPRewardRange:    [2]int{30, 150},
			GoldRewardRange:  [2]int{15, 60},
			ItemRewardChance: 0.4,
		},
		{
			BaseType:     TypeBoss,
			NamePrefixes: []string{"Defeat", "Vanquish", "Slay", "Destroy", "Conquer"},
			NameSuffixes: []string{"the Dragon Lord", "the Lich King", "the Dark Sorcerer", "the Demon Prince", "the Ancient Wyrm"},
			DescTemplates: []string{
				"%s has awakened and threatens the realm. You must defeat this powerful foe.",
				"Legends speak of %s. Only the bravest hero can face this challenge.",
				"The kingdom's survival depends on stopping %s. This will be your greatest battle.",
			},
			Tags:             []string{"boss", "challenge", "epic"},
			TargetTypes:      []string{"Dragon Lord", "Lich King", "Dark Sorcerer", "Demon Prince", "Ancient Wyrm"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{500, 2000},
			GoldRewardRange:  [2]int{200, 1000},
			ItemRewardChance: 0.9,
			SkillPointChance: 0.5,
		},
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Explore", "Discover", "Scout", "Survey", "Map"},
			NameSuffixes: []string{"the Ancient Ruins", "the Forgotten Dungeon", "the Enchanted Forest", "the Lost Temple", "the Mystic Caverns"},
			DescTemplates: []string{
				"Ancient maps point to %s. Explore the area and document your findings.",
				"Legends speak of %s. Be the first to uncover its secrets.",
				"The guild needs a survey of %s. Map the region and return safely.",
			},
			Tags:             []string{"exploration", "discovery"},
			TargetTypes:      []string{"Ancient Ruins", "Forgotten Dungeon", "Enchanted Forest", "Lost Temple", "Mystic Caverns"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{40, 180},
			GoldRewardRange:  [2]int{20, 80},
			ItemRewardChance: 0.35,
		},
	},
	"scifi": {
		{
			BaseType:     TypeKill,
			NamePrefixes: []string{"Terminate", "Eliminate", "Neutralize", "Destroy", "Eradicate"},
			NameSuffixes: []string{"the Rogue Bots", "the Alien Hostiles", "the Mutants", "the Pirates", "the Drones"},
			DescTemplates: []string{
				"Hostile %s detected in sector. Eliminate %d units.",
				"Security breach: %s are compromising the facility. Neutralize %d threats.",
				"Combat protocol initiated. Destroy %d %s to secure the area.",
			},
			Tags:             []string{"combat", "tactical"},
			TargetTypes:      []string{"Combat Drone", "Alien Warrior", "Mutant", "Space Pirate", "Rogue AI"},
			RequiredRange:    [2]int{5, 20},
			XPRewardRange:    [2]int{50, 200},
			GoldRewardRange:  [2]int{10, 50},
			ItemRewardChance: 0.3,
		},
		{
			BaseType:     TypeCollect,
			NamePrefixes: []string{"Salvage", "Recover", "Extract", "Retrieve", "Collect"},
			NameSuffixes: []string{"Data Cores", "Power Cells", "Tech Modules", "Mineral Samples", "Alien Artifacts"},
			DescTemplates: []string{
				"Mission: Acquire %d %s from the field. Return to base for debriefing.",
				"Scanning systems detected %s nearby. Collect %d units.",
				"Research requires %d %s. Locate and extract them from the area.",
			},
			Tags:             []string{"salvage", "exploration"},
			TargetTypes:      []string{"Data Core", "Power Cell", "Tech Module", "Mineral Sample", "Alien Artifact"},
			RequiredRange:    [2]int{3, 15},
			XPRewardRange:    [2]int{30, 150},
			GoldRewardRange:  [2]int{15, 60},
			ItemRewardChance: 0.4,
		},
		{
			BaseType:     TypeBoss,
			NamePrefixes: []string{"Eliminate", "Terminate", "Neutralize", "Destroy", "Defeat"},
			NameSuffixes: []string{"the Titan Mech", "the Alien Queen", "the AI Overlord", "the Warlord", "the Omega Unit"},
			DescTemplates: []string{
				"Priority target identified: %s. Engage with extreme caution.",
				"Threat level maximum. %s must be neutralized immediately.",
				"All units: %s is the primary objective. Eliminate this threat.",
			},
			Tags:             []string{"boss", "critical", "priority"},
			TargetTypes:      []string{"Titan Mech", "Alien Queen", "AI Overlord", "Warlord", "Omega Unit"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{500, 2000},
			GoldRewardRange:  [2]int{200, 1000},
			ItemRewardChance: 0.9,
			SkillPointChance: 0.5,
		},
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Scan", "Survey", "Investigate", "Recon", "Map"},
			NameSuffixes: []string{"the Derelict Station", "the Anomaly Zone", "the Abandoned Colony", "the Nebula Rift", "the Signal Source"},
			DescTemplates: []string{
				"Sensors detected activity near %s. Investigate and report findings.",
				"Command requests a full survey of %s. Proceed with caution.",
				"An uncharted region near %s requires exploration. Map the area.",
			},
			Tags:             []string{"exploration", "recon"},
			TargetTypes:      []string{"Derelict Station", "Anomaly Zone", "Abandoned Colony", "Nebula Rift", "Signal Source"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{40, 180},
			GoldRewardRange:  [2]int{20, 80},
			ItemRewardChance: 0.35,
		},
	},
	"horror": {
		{
			BaseType:     TypeKill,
			NamePrefixes: []string{"Purge", "Banish", "Exorcise", "Cleanse", "Destroy"},
			NameSuffixes: []string{"the Abominations", "the Possessed", "the Revenants", "the Nightmares", "the Cultists"},
			DescTemplates: []string{
				"Twisted %s lurk in the shadows. Destroy %d of them before they spread.",
				"The darkness has spawned %s. Purge %d of these horrors from the land.",
				"Unholy %s have risen. Banish %d of them back to the void.",
			},
			Tags:             []string{"combat", "horror"},
			TargetTypes:      []string{"Wraith", "Ghoul", "Cultist", "Wendigo", "Shadow Fiend", "Possessed Villager", "Flesh Golem"},
			RequiredRange:    [2]int{3, 15},
			XPRewardRange:    [2]int{60, 220},
			GoldRewardRange:  [2]int{10, 40},
			ItemRewardChance: 0.35,
		},
		{
			BaseType:     TypeCollect,
			NamePrefixes: []string{"Recover", "Retrieve", "Salvage", "Secure", "Gather"},
			NameSuffixes: []string{"Cursed Relics", "Forbidden Tomes", "Ritual Components", "Soul Fragments", "Warding Sigils"},
			DescTemplates: []string{
				"We need %d %s to complete the ritual of protection. Find them quickly.",
				"Scattered %s hold the key to stopping the darkness. Recover %d of them.",
				"The ward is failing. Collect %d %s to restore the seal.",
			},
			Tags:             []string{"gather", "occult"},
			TargetTypes:      []string{"Cursed Relic", "Forbidden Tome", "Ritual Component", "Soul Fragment", "Warding Sigil"},
			RequiredRange:    [2]int{3, 10},
			XPRewardRange:    [2]int{40, 160},
			GoldRewardRange:  [2]int{10, 50},
			ItemRewardChance: 0.45,
		},
		{
			BaseType:     TypeBoss,
			NamePrefixes: []string{"Confront", "Banish", "Seal", "Destroy", "Exorcise"},
			NameSuffixes: []string{"the Elder Horror", "the Plague Mother", "the Shadow King", "the Bone Collector", "the Whispering Dread"},
			DescTemplates: []string{
				"The source of the corruption is %s. You must face this horror alone.",
				"All signs point to %s as the architect of this nightmare. End it.",
				"Only by destroying %s can the tormented souls find peace.",
			},
			Tags:             []string{"boss", "horror", "nightmare"},
			TargetTypes:      []string{"Elder Horror", "Plague Mother", "Shadow King", "Bone Collector", "Whispering Dread"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{500, 2000},
			GoldRewardRange:  [2]int{150, 800},
			ItemRewardChance: 0.9,
			SkillPointChance: 0.5,
		},
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Investigate", "Search", "Explore", "Uncover", "Delve into"},
			NameSuffixes: []string{"the Abandoned Asylum", "the Haunted Manor", "the Catacombs", "the Blighted Village", "the Cursed Graveyard"},
			DescTemplates: []string{
				"People have gone missing near %s. Investigate what lurks within.",
				"Strange sounds echo from %s. Someone must explore the depths.",
				"The darkness emanating from %s grows stronger. Uncover its source.",
			},
			Tags:             []string{"exploration", "investigation"},
			TargetTypes:      []string{"Abandoned Asylum", "Haunted Manor", "Catacombs", "Blighted Village", "Cursed Graveyard"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{50, 200},
			GoldRewardRange:  [2]int{15, 70},
			ItemRewardChance: 0.4,
		},
	},
	"cyberpunk": {
		{
			BaseType:     TypeKill,
			NamePrefixes: []string{"Flatline", "Zero", "Brick", "Ice", "Smoke"},
			NameSuffixes: []string{"the Gangers", "the Corpos", "the Netrunners", "the Cyberpsychos", "the Boostergangs"},
			DescTemplates: []string{
				"A crew of %s is causing havoc in the district. Take out %d of them.",
				"Contract posted: Neutralize %d %s operating in the sector.",
				"The %s have overstepped. Flatline %d of their operatives.",
			},
			Tags:             []string{"combat", "street"},
			TargetTypes:      []string{"Ganger", "Corpo Agent", "Rogue Netrunner", "Cyberpsycho", "Boosterganger", "Maelstrom", "Scav"},
			RequiredRange:    [2]int{4, 18},
			XPRewardRange:    [2]int{55, 210},
			GoldRewardRange:  [2]int{15, 60},
			ItemRewardChance: 0.35,
		},
		{
			BaseType:     TypeCollect,
			NamePrefixes: []string{"Jack", "Acquire", "Boost", "Swipe", "Extract"},
			NameSuffixes: []string{"Neural Chips", "Cyberdecks", "Black ICE", "Bioware Samples", "Encrypted Shards"},
			DescTemplates: []string{
				"Client needs %d %s from the black market. No questions asked.",
				"High-value %s detected in the area. Acquire %d units.",
				"A fixer wants %d %s delivered. Payment on completion.",
			},
			Tags:             []string{"gather", "hustle"},
			TargetTypes:      []string{"Neural Chip", "Cyberdeck", "Black ICE Module", "Bioware Sample", "Encrypted Shard"},
			RequiredRange:    [2]int{2, 12},
			XPRewardRange:    [2]int{35, 160},
			GoldRewardRange:  [2]int{20, 75},
			ItemRewardChance: 0.4,
		},
		{
			BaseType:     TypeBoss,
			NamePrefixes: []string{"Flatline", "Terminate", "Breach", "Override", "Takedown"},
			NameSuffixes: []string{"the Corp Director", "the AI Construct", "the Chrome Warlord", "the Ghost in the Net", "the Syndicate Boss"},
			DescTemplates: []string{
				"Top-priority target: %s. Maximum force authorized.",
				"The streets won't be safe until %s is taken down. Make it happen.",
				"Highest-paying gig of the year: eliminate %s.",
			},
			Tags:             []string{"boss", "high-value", "gig"},
			TargetTypes:      []string{"Corp Director", "AI Construct", "Chrome Warlord", "Ghost in the Net", "Syndicate Boss"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{500, 2000},
			GoldRewardRange:  [2]int{250, 1200},
			ItemRewardChance: 0.9,
			SkillPointChance: 0.5,
		},
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Scout", "Infiltrate", "Recon", "Case", "Breach"},
			NameSuffixes: []string{"the Corp Tower", "the Underground Lab", "the Data Haven", "the Blacksite", "the Abandoned Arcology"},
			DescTemplates: []string{
				"Intel suggests something big at %s. Scout the location.",
				"A client needs eyes on %s. Get in, gather intel, get out.",
				"The location known as %s has gone dark. Investigate.",
			},
			Tags:             []string{"exploration", "infiltration"},
			TargetTypes:      []string{"Corp Tower", "Underground Lab", "Data Haven", "Blacksite", "Abandoned Arcology"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{45, 190},
			GoldRewardRange:  [2]int{25, 90},
			ItemRewardChance: 0.4,
		},
	},
	"postapoc": {
		{
			BaseType:     TypeKill,
			NamePrefixes: []string{"Clear", "Purge", "Eliminate", "Scour", "Hunt"},
			NameSuffixes: []string{"the Mutants", "the Raiders", "the Ferals", "the Scavengers", "the Wasteland Horrors"},
			DescTemplates: []string{
				"A pack of %s has been spotted near the settlement. Kill %d of them before they attack.",
				"The %s have grown too bold. Take out %d to secure the perimeter.",
				"Survivors report %s activity in the ruins. Eliminate %d to make the area safe.",
			},
			Tags:             []string{"combat", "wasteland"},
			TargetTypes:      []string{"Mutant", "Raider", "Feral Ghoul", "Scavenger Gang", "Wasteland Beast", "Irradiated Creature"},
			RequiredRange:    [2]int{4, 18},
			XPRewardRange:    [2]int{50, 200},
			GoldRewardRange:  [2]int{10, 50},
			ItemRewardChance: 0.3,
		},
		{
			BaseType:     TypeCollect,
			NamePrefixes: []string{"Scavenge", "Salvage", "Recover", "Retrieve", "Stockpile"},
			NameSuffixes: []string{"Clean Water", "Medical Supplies", "Fuel Canisters", "Pre-War Tech", "Food Rations"},
			DescTemplates: []string{
				"The settlement is running low on %s. Scavenge %d units from the wastes.",
				"A trader needs %d %s for the caravan. Good caps for fast work.",
				"Rumor has it there's %s in the old ruins. Bring back %d for the community.",
			},
			Tags:             []string{"gather", "survival"},
			TargetTypes:      []string{"Clean Water Jug", "Med Kit", "Fuel Canister", "Pre-War Tech Component", "Food Ration", "Radiation Medicine"},
			RequiredRange:    [2]int{3, 14},
			XPRewardRange:    [2]int{35, 150},
			GoldRewardRange:  [2]int{15, 65},
			ItemRewardChance: 0.35,
		},
		{
			BaseType:     TypeBoss,
			NamePrefixes: []string{"Overthrow", "Dethrone", "End", "Destroy", "Topple"},
			NameSuffixes: []string{"the Warlord", "the Mutant Alpha", "the Raider King", "the Wasteland Tyrant", "the Deathclaw Matriarch"},
			DescTemplates: []string{
				"The settlements will never be safe while %s lives. Take them down.",
				"%s has terrorized the wastes long enough. It's time to end their reign.",
				"A coalition of survivors has put a bounty on %s. The reward is substantial.",
			},
			Tags:             []string{"boss", "wasteland", "liberation"},
			TargetTypes:      []string{"Wasteland Warlord", "Mutant Alpha", "Raider King", "Ghoul Overlord", "Deathclaw Matriarch"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{450, 1800},
			GoldRewardRange:  [2]int{200, 1000},
			ItemRewardChance: 0.85,
			SkillPointChance: 0.45,
		},
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Scout", "Survey", "Map", "Investigate", "Reclaim"},
			NameSuffixes: []string{"the Bunker", "the Ruined City", "the Fallout Zone", "the Pre-War Facility", "the Dead Zone"},
			DescTemplates: []string{
				"Scouts report an unexplored %s. Map the area and report any threats.",
				"The %s may contain valuable salvage. Explore and mark points of interest.",
				"A survivor claims to have seen lights in %s. Investigate the location.",
			},
			Tags:             []string{"exploration", "salvage"},
			TargetTypes:      []string{"Abandoned Bunker", "Ruined City Block", "Fallout Zone", "Pre-War Military Base", "Collapsed Vault"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{40, 175},
			GoldRewardRange:  [2]int{20, 80},
			ItemRewardChance: 0.45,
		},
	},
}

// GetQuestTemplates returns all quest templates for the given genre.
// This is the preferred entry point; the per-type functions are kept for
// backwards compatibility. Returns fantasy templates for unknown genre IDs.
func GetQuestTemplates(genreID string) []QuestTemplate {
	if t, ok := genreQuestTemplates[genreID]; ok {
		return t
	}
	return genreQuestTemplates["fantasy"]
}

// getTemplatesByType filters templates from the named genre by quest type.
func getTemplatesByType(genreID string, qt QuestType) []QuestTemplate {
	all := GetQuestTemplates(genreID)
	out := make([]QuestTemplate, 0, 1)
	for _, t := range all {
		if t.BaseType == qt {
			out = append(out, t)
		}
	}
	return out
}
