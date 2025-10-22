package skills

// SkillType represents the classification of a skill.
type SkillType int

const (
	// TypePassive represents passive bonuses (no activation required)
	TypePassive SkillType = iota
	// TypeActive represents active abilities (player must activate)
	TypeActive
	// TypeUltimate represents powerful ultimate abilities (long cooldown)
	TypeUltimate
	// TypeSynergy represents skills that enhance other skills
	TypeSynergy
)

// String returns the string representation of a skill type.
func (t SkillType) String() string {
	switch t {
	case TypePassive:
		return "passive"
	case TypeActive:
		return "active"
	case TypeUltimate:
		return "ultimate"
	case TypeSynergy:
		return "synergy"
	default:
		return "unknown"
	}
}

// SkillCategory represents the gameplay role of a skill.
type SkillCategory int

const (
	// CategoryCombat represents damage and combat-focused skills
	CategoryCombat SkillCategory = iota
	// CategoryDefense represents defensive and survival skills
	CategoryDefense
	// CategoryUtility represents utility and convenience skills
	CategoryUtility
	// CategoryMagic represents spell and mana-related skills
	CategoryMagic
	// CategoryCrafting represents crafting and item-related skills
	CategoryCrafting
	// CategorySocial represents NPC interaction and trading skills
	CategorySocial
)

// String returns the string representation of a skill category.
func (c SkillCategory) String() string {
	switch c {
	case CategoryCombat:
		return "combat"
	case CategoryDefense:
		return "defense"
	case CategoryUtility:
		return "utility"
	case CategoryMagic:
		return "magic"
	case CategoryCrafting:
		return "crafting"
	case CategorySocial:
		return "social"
	default:
		return "unknown"
	}
}

// Tier represents the power tier of a skill within a tree.
type Tier int

const (
	// TierBasic represents starting skills (tier 1)
	TierBasic Tier = iota
	// TierIntermediate represents mid-tier skills (tier 2-3)
	TierIntermediate
	// TierAdvanced represents high-tier skills (tier 4-5)
	TierAdvanced
	// TierMaster represents master skills (tier 6+)
	TierMaster
)

// String returns the string representation of a tier.
func (t Tier) String() string {
	switch t {
	case TierBasic:
		return "basic"
	case TierIntermediate:
		return "intermediate"
	case TierAdvanced:
		return "advanced"
	case TierMaster:
		return "master"
	default:
		return "unknown"
	}
}

// Skill represents a single skill/ability in the skill tree.
type Skill struct {
	ID           string        // Unique identifier
	Name         string        // Display name
	Description  string        // Description of effects
	Type         SkillType     // Passive, Active, Ultimate, Synergy
	Category     SkillCategory // Combat, Defense, Magic, etc.
	Tier         Tier          // Power tier
	Level        int           // Current level (0 = unlearned)
	MaxLevel     int           // Maximum level
	Requirements Requirements  // What's needed to unlock
	Effects      []Effect      // Stat bonuses and effects
	Tags         []string      // Searchable tags
	Seed         int64         // Generation seed for determinism
}

// Requirements defines what's needed to unlock a skill.
type Requirements struct {
	PlayerLevel       int            // Minimum player level
	SkillPoints       int            // Skill points needed
	PrerequisiteIDs   []string       // Skills that must be learned first
	AttributeMinimums map[string]int // e.g., {"strength": 10}
}

// Effect represents a bonus or modification provided by a skill.
type Effect struct {
	Type        string  // "damage", "defense", "speed", "mana", etc.
	Value       float64 // Numeric value (can be percentage)
	IsPercent   bool    // Whether value is a percentage
	Description string  // Human-readable description
}

// SkillNode represents a node in the skill tree structure.
type SkillNode struct {
	Skill    *Skill       // The skill at this node
	Children []*SkillNode // Skills that require this one
	Position Position     // Visual position in tree
}

// Position represents 2D coordinates for tree visualization.
type Position struct {
	X int // Horizontal position
	Y int // Vertical position (tier)
}

// SkillTree represents a complete skill progression tree.
type SkillTree struct {
	ID          string        // Unique tree identifier
	Name        string        // Display name (e.g., "Warrior", "Mage")
	Description string        // Tree description
	Category    SkillCategory // Primary category
	Genre       string        // fantasy, scifi, etc.
	Nodes       []*SkillNode  // All skills in this tree
	RootNodes   []*SkillNode  // Starting skills (no prerequisites)
	MaxPoints   int           // Maximum total skill points
	Seed        int64         // Generation seed
}

// SkillTemplate defines a template for procedural skill generation.
type SkillTemplate struct {
	BaseType          SkillType
	BaseCategory      SkillCategory
	NamePrefixes      []string
	NameSuffixes      []string
	DescriptionFormat string
	EffectTypes       []string
	ValueRanges       map[string][2]float64 // min, max for each effect
	Tags              []string
	TierRange         [2]int // min, max tier
	MaxLevelRange     [2]int // min, max level
}

// IsUnlocked checks if a skill can be unlocked given current state.
func (s *Skill) IsUnlocked(playerLevel int, skillPoints int, learnedSkills map[string]bool, attributes map[string]int) bool {
	// Check player level
	if playerLevel < s.Requirements.PlayerLevel {
		return false
	}

	// Check skill points
	if skillPoints < s.Requirements.SkillPoints {
		return false
	}

	// Check prerequisites
	for _, prereqID := range s.Requirements.PrerequisiteIDs {
		if !learnedSkills[prereqID] {
			return false
		}
	}

	// Check attribute minimums
	for attr, minValue := range s.Requirements.AttributeMinimums {
		if attributes[attr] < minValue {
			return false
		}
	}

	return true
}

// CanLevelUp checks if skill can be leveled up further.
func (s *Skill) CanLevelUp() bool {
	return s.Level > 0 && s.Level < s.MaxLevel
}

// TotalPoints calculates total skill points in a tree.
func (st *SkillTree) TotalPoints() int {
	total := 0
	for _, node := range st.Nodes {
		if node.Skill.Level > 0 {
			total += node.Skill.Level * node.Skill.Requirements.SkillPoints
		}
	}
	return total
}

// GetSkillByID finds a skill by its ID.
func (st *SkillTree) GetSkillByID(id string) *Skill {
	for _, node := range st.Nodes {
		if node.Skill.ID == id {
			return node.Skill
		}
	}
	return nil
}

// GetTierSkills returns all skills in a specific tier.
func (st *SkillTree) GetTierSkills(tier Tier) []*Skill {
	skills := make([]*Skill, 0)
	for _, node := range st.Nodes {
		if node.Skill.Tier == tier {
			skills = append(skills, node.Skill)
		}
	}
	return skills
}
