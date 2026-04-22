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
// Data is sourced from genreQuestTemplates in templates.go.
func GetFantasyKillTemplates() []QuestTemplate { return getTemplatesByType("fantasy", TypeKill) }

// GetFantasyCollectTemplates returns collect quest templates for fantasy genre.
func GetFantasyCollectTemplates() []QuestTemplate { return getTemplatesByType("fantasy", TypeCollect) }

// GetFantasyBossTemplates returns boss quest templates for fantasy genre.
func GetFantasyBossTemplates() []QuestTemplate { return getTemplatesByType("fantasy", TypeBoss) }

// GetFantasyExploreTemplates returns explore quest templates for fantasy genre.
func GetFantasyExploreTemplates() []QuestTemplate { return getTemplatesByType("fantasy", TypeExplore) }

// GetSciFiKillTemplates returns kill quest templates for sci-fi genre.
func GetSciFiKillTemplates() []QuestTemplate { return getTemplatesByType("scifi", TypeKill) }

// GetSciFiCollectTemplates returns collect quest templates for sci-fi genre.
func GetSciFiCollectTemplates() []QuestTemplate { return getTemplatesByType("scifi", TypeCollect) }

// GetSciFiBossTemplates returns boss quest templates for sci-fi genre.
func GetSciFiBossTemplates() []QuestTemplate { return getTemplatesByType("scifi", TypeBoss) }

// GetSciFiExploreTemplates returns explore quest templates for sci-fi genre.
func GetSciFiExploreTemplates() []QuestTemplate { return getTemplatesByType("scifi", TypeExplore) }

// GetHorrorKillTemplates returns kill quest templates for horror genre.
func GetHorrorKillTemplates() []QuestTemplate { return getTemplatesByType("horror", TypeKill) }

// GetHorrorCollectTemplates returns collect quest templates for horror genre.
func GetHorrorCollectTemplates() []QuestTemplate { return getTemplatesByType("horror", TypeCollect) }

// GetHorrorBossTemplates returns boss quest templates for horror genre.
func GetHorrorBossTemplates() []QuestTemplate { return getTemplatesByType("horror", TypeBoss) }

// GetHorrorExploreTemplates returns explore quest templates for horror genre.
func GetHorrorExploreTemplates() []QuestTemplate { return getTemplatesByType("horror", TypeExplore) }

// GetCyberpunkKillTemplates returns kill quest templates for cyberpunk genre.
func GetCyberpunkKillTemplates() []QuestTemplate { return getTemplatesByType("cyberpunk", TypeKill) }

// GetCyberpunkCollectTemplates returns collect quest templates for cyberpunk genre.
func GetCyberpunkCollectTemplates() []QuestTemplate {
	return getTemplatesByType("cyberpunk", TypeCollect)
}

// GetCyberpunkBossTemplates returns boss quest templates for cyberpunk genre.
func GetCyberpunkBossTemplates() []QuestTemplate { return getTemplatesByType("cyberpunk", TypeBoss) }

// GetCyberpunkExploreTemplates returns explore quest templates for cyberpunk genre.
func GetCyberpunkExploreTemplates() []QuestTemplate {
	return getTemplatesByType("cyberpunk", TypeExplore)
}

// GetPostApocKillTemplates returns kill quest templates for post-apocalyptic genre.
func GetPostApocKillTemplates() []QuestTemplate { return getTemplatesByType("postapoc", TypeKill) }

// GetPostApocCollectTemplates returns collect quest templates for post-apocalyptic genre.
func GetPostApocCollectTemplates() []QuestTemplate {
	return getTemplatesByType("postapoc", TypeCollect)
}

// GetPostApocBossTemplates returns boss quest templates for post-apocalyptic genre.
func GetPostApocBossTemplates() []QuestTemplate { return getTemplatesByType("postapoc", TypeBoss) }

// GetPostApocExploreTemplates returns explore quest templates for post-apocalyptic genre.
func GetPostApocExploreTemplates() []QuestTemplate {
	return getTemplatesByType("postapoc", TypeExplore)
}
