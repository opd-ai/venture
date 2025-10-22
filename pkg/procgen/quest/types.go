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

// IsComplete returns true if the objective is met.
func (o *Objective) IsComplete() bool {
	return o.Current >= o.Required
}

// Progress returns completion percentage (0.0-1.0).
func (o *Objective) Progress() float64 {
	if o.Required == 0 {
		return 1.0
	}
	progress := float64(o.Current) / float64(o.Required)
	if progress > 1.0 {
		return 1.0
	}
	return progress
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
}

// IsComplete returns true if all objectives are met.
func (q *Quest) IsComplete() bool {
	for _, obj := range q.Objectives {
		if !obj.IsComplete() {
			return false
		}
	}
	return len(q.Objectives) > 0
}

// Progress returns overall completion percentage (0.0-1.0).
func (q *Quest) Progress() float64 {
	if len(q.Objectives) == 0 {
		return 1.0
	}

	totalProgress := 0.0
	for _, obj := range q.Objectives {
		totalProgress += obj.Progress()
	}
	return totalProgress / float64(len(q.Objectives))
}

// GetRewardValue estimates total reward value.
func (q *Quest) GetRewardValue() int {
	value := q.Reward.XP
	value += q.Reward.Gold * 2
	value += len(q.Reward.Items) * 100
	value += q.Reward.SkillPoints * 500
	return value
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
