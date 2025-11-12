package engine

// BookType represents different types of books
type BookType int

const (
	BookTypeSkill BookType = iota
	BookTypeLore
	BookTypeQuest
	BookTypeRecipe
	BookTypeHistory
)

// BookComponent represents an in-game book
type BookComponent struct {
	Title      string
	Author     string
	BookType   BookType
	Content    []string // Pages of text
	IsRead     bool
	SkillBonus map[string]float64 // Skill books grant bonuses
	RecipeID   string             // Recipe books unlock crafting
}

// Type returns the component type
func (b BookComponent) Type() string {
	return "book"
}

// LibraryComponent tracks collected books
type LibraryComponent struct {
	Books       []uint64        // Book entity IDs
	Completions map[string]bool // Series tracking
}

// Type returns the component type
func (l LibraryComponent) Type() string {
	return "library"
}
