// Package quest provides quest type definitions.
// This file defines quest types, objectives, rewards, and quest state
// used by the quest generator.
package quest

// QuestType represents the classification of a quest.
type QuestType int

const (
	// TypeKill represents quests to defeat enemies
	TypeKill QuestType = iota
	// TypeCollect represents quests to gather items
	TypeCollect
	// TypeEscort represents quests to protect NPCs
	TypeEscort
	// TypeExplore represents quests to discover locations
	TypeExplore
	// TypeTalk represents quests to interact with NPCs
	TypeTalk
	// TypeBoss represents quests to defeat specific bosses
	TypeBoss
	// TypeMoralChoice represents quests with moral decisions
	TypeMoralChoice
	// TypeFactionConflict represents quests involving faction choices
	TypeFactionConflict
)

// String returns the string representation of a quest type.
func (t QuestType) String() string {
	switch t {
	case TypeKill:
		return "kill"
	case TypeCollect:
		return "collect"
	case TypeEscort:
		return "escort"
	case TypeExplore:
		return "explore"
	case TypeTalk:
		return "talk"
	case TypeBoss:
		return "boss"
	case TypeMoralChoice:
		return "moral_choice"
	case TypeFactionConflict:
		return "faction_conflict"
	default:
		return "unknown"
	}
}

// QuestStatus represents the current state of a quest.
type QuestStatus int

const (
	// StatusNotStarted indicates the quest hasn't been accepted
	StatusNotStarted QuestStatus = iota
	// StatusActive indicates the quest is in progress
	StatusActive
	// StatusComplete indicates objectives are met but quest not turned in
	StatusComplete
	// StatusTurnedIn indicates quest has been completed and rewards claimed
	StatusTurnedIn
	// StatusFailed indicates the quest has failed
	StatusFailed
)

// String returns the string representation of a quest status.
func (s QuestStatus) String() string {
	switch s {
	case StatusNotStarted:
		return "not_started"
	case StatusActive:
		return "active"
	case StatusComplete:
		return "complete"
	case StatusTurnedIn:
		return "turned_in"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Difficulty represents how challenging a quest is.
type Difficulty int

const (
	// DifficultyTrivial represents very easy quests
	DifficultyTrivial Difficulty = iota
	// DifficultyEasy represents easy quests
	DifficultyEasy
	// DifficultyNormal represents standard difficulty
	DifficultyNormal
	// DifficultyHard represents challenging quests
	DifficultyHard
	// DifficultyElite represents very difficult quests
	DifficultyElite
	// DifficultyLegendary represents the hardest quests
	DifficultyLegendary
)

// String returns the string representation of a difficulty level.
func (d Difficulty) String() string {
	switch d {
	case DifficultyTrivial:
		return "trivial"
	case DifficultyEasy:
		return "easy"
	case DifficultyNormal:
		return "normal"
	case DifficultyHard:
		return "hard"
	case DifficultyElite:
		return "elite"
	case DifficultyLegendary:
		return "legendary"
	default:
		return "unknown"
	}
}

// Objective represents a single quest objective.
type Objective struct {
	// Description is human-readable objective text
	Description string
	// Target is what needs to be achieved (entity type, item name, location)
	Target string
	// Required is how many are needed
	Required int
	// Current is progress toward the objective
	Current int
}

// ObjectiveIsComplete returns true if the objective's current progress
// meets or exceeds its required amount.
//
// Deprecated: Use o.IsComplete() instead.
func ObjectiveIsComplete(o *Objective) bool {
	return o.Current >= o.Required
}

// IsComplete returns true if the objective is met.
// Delegates to ObjectiveIsComplete for ECS-compliant usage.
func (o *Objective) IsComplete() bool {
	return ObjectiveIsComplete(o)
}

// ObjectiveProgress returns the objective's completion percentage (0.0-1.0).
func ObjectiveProgress(o *Objective) float64 {
	if o.Required == 0 {
		return 1.0
	}
	progress := float64(o.Current) / float64(o.Required)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// Progress returns completion percentage (0.0-1.0).
// Delegates to ObjectiveProgress for ECS-compliant usage.
func (o *Objective) Progress() float64 {
	return ObjectiveProgress(o)
}

// Reward represents rewards given upon quest completion.
type Reward struct {
	// XP is experience points awarded
	XP int
	// Gold is currency awarded
	Gold int
	// Items are item IDs or types awarded
	Items []string
	// SkillPoints are skill points awarded
	SkillPoints int
}

// Quest represents a generated quest.
type Quest struct {
	// ID is a unique identifier for this quest
	ID string
	// Name is the procedurally generated name
	Name string
	// Type categorizes the quest
	Type QuestType
	// Difficulty indicates how challenging the quest is
	Difficulty Difficulty
	// Description is generated flavor text
	Description string
	// Objectives are what the player must accomplish
	Objectives []Objective
	// Reward is what the player receives upon completion
	Reward Reward
	// RequiredLevel is minimum level to accept the quest
	RequiredLevel int
	// Status tracks quest state
	Status QuestStatus
	// Seed is the generation seed for this quest
	Seed int64
	// Tags are additional descriptive labels
	Tags []string
	// GiverNPC is the NPC who gives the quest (optional)
	GiverNPC string
	// Location is where the quest takes place (optional)
	Location string
	// MoralChoiceID links this quest to a moral choice (optional)
	MoralChoiceID string
	// FactionA is the first faction in a faction conflict quest (optional)
	FactionA string
	// FactionB is the second faction in a faction conflict quest (optional)
	FactionB string
	// HasMoralConsequences indicates if completing this quest triggers moral choices
	HasMoralConsequences bool
}

// QuestIsComplete returns true if all objectives in the quest are met.
//
// Deprecated: Use q.IsComplete() instead.
func QuestIsComplete(q *Quest) bool {
	for i := range q.Objectives {
		if !ObjectiveIsComplete(&q.Objectives[i]) {
			return false
		}
	}
	return len(q.Objectives) > 0
}

// IsComplete returns true if all objectives are met.
// Delegates to QuestIsComplete for ECS-compliant usage.
func (q *Quest) IsComplete() bool {
	return QuestIsComplete(q)
}

// QuestProgress returns the quest's overall completion percentage (0.0-1.0).
func QuestProgress(q *Quest) float64 {
	if len(q.Objectives) == 0 {
		return 1.0
	}

	totalProgress := 0.0
	for i := range q.Objectives {
		totalProgress += ObjectiveProgress(&q.Objectives[i])
	}
	return totalProgress / float64(len(q.Objectives))
}

// Progress returns overall completion percentage (0.0-1.0).
// Delegates to QuestProgress for ECS-compliant usage.
func (q *Quest) Progress() float64 {
	return QuestProgress(q)
}

// QuestRewardValue estimates the total reward value of a quest.
func QuestRewardValue(q *Quest) int {
	value := q.Reward.XP
	value += q.Reward.Gold * 2
	value += len(q.Reward.Items) * 100
	value += q.Reward.SkillPoints * 500
	return value
}

// GetRewardValue estimates total reward value.
// Delegates to QuestRewardValue for ECS-compliant usage.
func (q *Quest) GetRewardValue() int {
	return QuestRewardValue(q)
}

// QuestTemplate defines a template for generating quests.
type QuestTemplate struct {
	BaseType         QuestType
	NamePrefixes     []string
	NameSuffixes     []string
	DescTemplates    []string
	Tags             []string
	TargetTypes      []string
	RequiredRange    [2]int
	XPRewardRange    [2]int
	GoldRewardRange  [2]int
	ItemRewardChance float64
	SkillPointChance float64
}

// GetFantasyKillTemplates returns kill quest templates for fantasy genre.
// TODO(REM-144): These genre template functions (200+ lines each) could be
// replaced with data-driven tables: var questTemplates = map[string][]QuestTemplate{}.
// The current approach requires modifying function bodies to add new templates.
func GetFantasyKillTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetFantasyCollectTemplates returns collect quest templates for fantasy genre.
func GetFantasyCollectTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetFantasyBossTemplates returns boss quest templates for fantasy genre.
func GetFantasyBossTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetFantasyExploreTemplates returns explore quest templates for fantasy genre.
func GetFantasyExploreTemplates() []QuestTemplate {
	return []QuestTemplate{
		{
			BaseType:     TypeExplore,
			NamePrefixes: []string{"Explore", "Discover", "Scout", "Survey", "Map"},
			NameSuffixes: []string{"the Ancient Ruins", "the Dark Forest", "the Forgotten Temple", "the Mountain Pass", "the Lost City"},
			DescTemplates: []string{
				"We need someone to explore %s. Report back what you find.",
				"Strange reports come from %s. Investigate the area.",
				"Ancient maps mention %s. Discover this location's secrets.",
			},
			Tags:             []string{"exploration", "adventure"},
			TargetTypes:      []string{"Ancient Ruins", "Dark Forest", "Forgotten Temple", "Mountain Pass", "Lost City"},
			RequiredRange:    [2]int{1, 1},
			XPRewardRange:    [2]int{40, 180},
			GoldRewardRange:  [2]int{20, 80},
			ItemRewardChance: 0.35,
		},
	}
}

// GetSciFiKillTemplates returns kill quest templates for sci-fi genre.
func GetSciFiKillTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetSciFiCollectTemplates returns collect quest templates for sci-fi genre.
func GetSciFiCollectTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetSciFiExploreTemplates returns explore quest templates for sci-fi genre.
func GetSciFiExploreTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetSciFiBossTemplates returns boss quest templates for sci-fi genre.
func GetSciFiBossTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetHorrorKillTemplates returns kill quest templates for horror genre.
func GetHorrorKillTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetHorrorCollectTemplates returns collect quest templates for horror genre.
func GetHorrorCollectTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetHorrorBossTemplates returns boss quest templates for horror genre.
func GetHorrorBossTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetHorrorExploreTemplates returns explore quest templates for horror genre.
func GetHorrorExploreTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetCyberpunkKillTemplates returns kill quest templates for cyberpunk genre.
func GetCyberpunkKillTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetCyberpunkCollectTemplates returns collect quest templates for cyberpunk genre.
func GetCyberpunkCollectTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetCyberpunkBossTemplates returns boss quest templates for cyberpunk genre.
func GetCyberpunkBossTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetCyberpunkExploreTemplates returns explore quest templates for cyberpunk genre.
func GetCyberpunkExploreTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetPostApocKillTemplates returns kill quest templates for post-apocalyptic genre.
func GetPostApocKillTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetPostApocCollectTemplates returns collect quest templates for post-apocalyptic genre.
func GetPostApocCollectTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetPostApocBossTemplates returns boss quest templates for post-apocalyptic genre.
func GetPostApocBossTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}

// GetPostApocExploreTemplates returns explore quest templates for post-apocalyptic genre.
func GetPostApocExploreTemplates() []QuestTemplate {
	return []QuestTemplate{
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
	}
}
