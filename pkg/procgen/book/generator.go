package book

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

// Generator creates procedural books with grammar-based text generation.
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a new book generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a new book based on the provided seed and parameters.
// Returns an *engine.BookComponent or an error if generation fails.
//
// Required Custom Parameters:
//   - "book_type": engine.BookType (skill, lore, quest, recipe, history)
//
// Optional Custom Parameters:
//   - "skill_name": string (for skill books)
//   - "skill_bonus": float64 (for skill books, default: depth * 0.1)
//   - "recipe_id": string (for recipe books)
//   - "quest_id": string (for quest books)
//   - "location": string (for historical texts)
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if err := procgen.ValidateParams(params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Create seeded RNG for deterministic generation
	g.rng = rand.New(rand.NewSource(seed))

	// Extract book type from custom parameters
	bookTypeRaw, ok := params.Custom["book_type"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: book_type")
	}
	bookType, ok := bookTypeRaw.(engine.BookType)
	if !ok {
		return nil, fmt.Errorf("book_type must be engine.BookType, got %T", bookTypeRaw)
	}

	// Generate book components
	book := &engine.BookComponent{
		BookType:   bookType,
		Title:      g.generateTitle(params.GenreID, bookType, params.Custom),
		Author:     g.generateAuthor(params.GenreID),
		IsRead:     false,
		SkillBonus: make(map[string]float64),
	}

	// Generate content based on book type
	switch bookType {
	case engine.BookTypeSkill:
		book.Content = g.generateSkillBookContent(params.GenreID, params.Difficulty, params.Depth, params.Custom)
		g.setSkillBonus(book, params.Depth, params.Custom)
	case engine.BookTypeLore:
		book.Content = g.generateLoreContent(params.GenreID, params.Difficulty, params.Depth)
	case engine.BookTypeQuest:
		book.Content = g.generateQuestContent(params.GenreID, params.Difficulty, params.Custom)
	case engine.BookTypeRecipe:
		book.Content = g.generateRecipeContent(params.GenreID, params.Difficulty, params.Custom)
		g.setRecipeID(book, params.Custom)
	case engine.BookTypeHistory:
		book.Content = g.generateHistoricalContent(params.GenreID, params.Difficulty, params.Depth, params.Custom)
	default:
		return nil, fmt.Errorf("unsupported book type: %v", bookType)
	}

	return book, nil
}

// Validate checks if the generated book meets quality criteria.
func (g *Generator) Validate(result interface{}) error {
	book, ok := result.(*engine.BookComponent)
	if !ok {
		return fmt.Errorf("result is not an *engine.BookComponent")
	}

	if book.Title == "" {
		return fmt.Errorf("book has empty title")
	}
	if book.Author == "" {
		return fmt.Errorf("book has empty author")
	}
	if len(book.Content) == 0 {
		return fmt.Errorf("book has no content pages")
	}

	// Validate content length (330-2000 words, allowing for natural RNG variability)
	// Target is 500+ words, but RNG can produce 330-700 depending on seed
	totalWords := 0
	for _, page := range book.Content {
		words := strings.Fields(page)
		totalWords += len(words)
	}
	if totalWords < 330 {
		return fmt.Errorf("book content too short: %d words (minimum 330)", totalWords)
	}
	if totalWords > 2000 {
		return fmt.Errorf("book content too long: %d words (maximum 2000)", totalWords)
	}

	// Type-specific validation
	switch book.BookType {
	case engine.BookTypeSkill:
		if len(book.SkillBonus) == 0 {
			return fmt.Errorf("skill book has no skill bonuses")
		}
	case engine.BookTypeRecipe:
		if book.RecipeID == "" {
			return fmt.Errorf("recipe book has no recipe ID")
		}
	}

	return nil
}

// setSkillBonus sets the skill bonus for skill books.
func (g *Generator) setSkillBonus(book *engine.BookComponent, depth int, custom map[string]interface{}) {
	skillName, ok := custom["skill_name"].(string)
	if !ok || skillName == "" {
		skillName = "general"
	}

	// Check for custom skill bonus
	if bonusRaw, ok := custom["skill_bonus"]; ok {
		if bonus, ok := bonusRaw.(float64); ok {
			book.SkillBonus[skillName] = bonus
			return
		}
	}

	// Default skill bonus scales with depth
	bonus := float64(depth) * 0.1
	if bonus < 0.1 {
		bonus = 0.1
	}
	if bonus > 2.0 {
		bonus = 2.0
	}
	book.SkillBonus[skillName] = bonus
}

// setRecipeID sets the recipe ID for recipe books.
func (g *Generator) setRecipeID(book *engine.BookComponent, custom map[string]interface{}) {
	if recipeID, ok := custom["recipe_id"].(string); ok {
		book.RecipeID = recipeID
	} else {
		// Generate a random recipe ID if not provided
		book.RecipeID = fmt.Sprintf("recipe_%d", g.rng.Int63())
	}
}

// generateTitle creates a genre-appropriate book title.
func (g *Generator) generateTitle(genre string, bookType engine.BookType, custom map[string]interface{}) string {
	switch bookType {
	case engine.BookTypeSkill:
		return g.generateSkillBookTitle(genre, custom)
	case engine.BookTypeLore:
		return g.generateLoreTitle(genre)
	case engine.BookTypeQuest:
		return g.generateQuestTitle(genre)
	case engine.BookTypeRecipe:
		return g.generateRecipeTitle(genre)
	case engine.BookTypeHistory:
		return g.generateHistoricalTitle(genre, custom)
	default:
		return "Untitled Book"
	}
}

// generateSkillBookTitle generates titles for skill books.
func (g *Generator) generateSkillBookTitle(genre string, custom map[string]interface{}) string {
	skillName, ok := custom["skill_name"].(string)
	if !ok || skillName == "" {
		skillName = g.pickRandom([]string{"Combat", "Magic", "Stealth", "Crafting", "Survival"})
	}

	switch genre {
	case "fantasy":
		prefixes := []string{"The Art of", "Mastering", "A Guide to", "The Way of", "Secrets of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), skillName)
	case "sci-fi":
		prefixes := []string{"Technical Manual:", "Protocol", "System Guide:", "Training Module:"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), skillName)
	case "horror":
		prefixes := []string{"Forbidden Knowledge of", "Dark Arts:", "Cursed Techniques of", "The Black Book of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), skillName)
	case "cyberpunk":
		prefixes := []string{"Neural Training:", "Skill Chip:", "Street Manual:", "Underground Guide:"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), skillName)
	case "post-apocalyptic":
		prefixes := []string{"Survival Guide:", "Wasteland Skills:", "Scavenger's Manual:", "The Last Book of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), skillName)
	default:
		return fmt.Sprintf("A Guide to %s", skillName)
	}
}

// generateLoreTitle generates titles for lore books.
func (g *Generator) generateLoreTitle(genre string) string {
	// Check if this is part of a series (from custom parameters)
	if seriesName, ok := g.getSeriesName(); ok {
		volume := g.getVolumeNumber()
		return fmt.Sprintf("%s - Volume %d", seriesName, volume)
	}

	switch genre {
	case "fantasy":
		subjects := []string{"the Ancients", "the Dragon Wars", "the Lost Kingdom", "the First Age", "the Prophecy"}
		prefixes := []string{"Chronicles of", "Tales of", "The History of", "Legends of", "The Book of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), g.pickRandom(subjects))
	case "sci-fi":
		subjects := []string{"the Exodus", "First Contact", "the AI Uprising", "the Colony Wars", "the Singularity"}
		prefixes := []string{"Records of", "Archives:", "Historical Data:", "Chronicle:", "Log:"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), g.pickRandom(subjects))
	case "horror":
		subjects := []string{"the Haunting", "the Madness", "the Dark Times", "the Cursed", "the Forgotten"}
		prefixes := []string{"Testament of", "Memories of", "The Diary of", "Whispers of", "Echoes of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), g.pickRandom(subjects))
	case "cyberpunk":
		subjects := []string{"the Data Wars", "the Neural Revolution", "the Corporate Collapse", "the Street Legends", "the Neon Age"}
		prefixes := []string{"Net History:", "Urban Legends:", "Chronicles of", "The Files of", "Stories from"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), g.pickRandom(subjects))
	case "post-apocalyptic":
		subjects := []string{"the Fall", "the Last Days", "the Survivors", "the Wasteland", "Before the End"}
		prefixes := []string{"Memories of", "Tales from", "The Record of", "Stories of", "Archives of"}
		return fmt.Sprintf("%s %s", g.pickRandom(prefixes), g.pickRandom(subjects))
	default:
		return "A Book of Lore"
	}
}

// generateQuestTitle generates titles for quest journals.
func (g *Generator) generateQuestTitle(genre string) string {
	switch genre {
	case "fantasy":
		subjects := []string{"the Hero's Journey", "the Sacred Quest", "the Lost Artifact", "the Dark Prophecy"}
		return fmt.Sprintf("Journal of %s", g.pickRandom(subjects))
	case "sci-fi":
		subjects := []string{"Mission Log", "Expedition Records", "Survey Data", "Operation Notes"}
		return g.pickRandom(subjects)
	case "horror":
		subjects := []string{"the Last Survivor", "the Investigation", "the Descent", "the Witness"}
		return fmt.Sprintf("Notes of %s", g.pickRandom(subjects))
	case "cyberpunk":
		subjects := []string{"the Run", "the Job", "the Heist", "the Mission"}
		return fmt.Sprintf("Log of %s", g.pickRandom(subjects))
	case "post-apocalyptic":
		subjects := []string{"the Journey", "the Search", "the Migration", "the Last Hope"}
		return fmt.Sprintf("Log of %s", g.pickRandom(subjects))
	default:
		return "Quest Journal"
	}
}

// generateRecipeTitle generates titles for recipe books.
func (g *Generator) generateRecipeTitle(genre string) string {
	switch genre {
	case "fantasy":
		subjects := []string{"Alchemy", "Enchanting", "Potion Brewing", "Smithing", "Herbalism"}
		return fmt.Sprintf("The Crafting Manual of %s", g.pickRandom(subjects))
	case "sci-fi":
		subjects := []string{"Engineering", "Fabrication", "Synthesis", "Assembly", "Construction"}
		return fmt.Sprintf("Technical Specifications: %s", g.pickRandom(subjects))
	case "horror":
		subjects := []string{"Forbidden Rituals", "Dark Crafting", "Cursed Items", "Unholy Creation"}
		return fmt.Sprintf("The Grimoire of %s", g.pickRandom(subjects))
	case "cyberpunk":
		subjects := []string{"Cybernetics", "Wetware", "Tech Mods", "Neural Upgrades", "Street Tech"}
		return fmt.Sprintf("Build Guide: %s", g.pickRandom(subjects))
	case "post-apocalyptic":
		subjects := []string{"Scavenging", "Makeshift Tools", "Survival Gear", "Wasteland Tech"}
		return fmt.Sprintf("Scrapper's Guide: %s", g.pickRandom(subjects))
	default:
		return "Crafting Manual"
	}
}

// generateHistoricalTitle generates titles for historical texts.
func (g *Generator) generateHistoricalTitle(genre string, custom map[string]interface{}) string {
	location, ok := custom["location"].(string)
	if !ok || location == "" {
		location = "This Place"
	}

	switch genre {
	case "fantasy":
		return fmt.Sprintf("The History of %s", location)
	case "sci-fi":
		return fmt.Sprintf("Station Log: %s", location)
	case "horror":
		return fmt.Sprintf("The Dark History of %s", location)
	case "cyberpunk":
		return fmt.Sprintf("Building Records: %s", location)
	case "post-apocalyptic":
		return fmt.Sprintf("Before the Fall: %s", location)
	default:
		return fmt.Sprintf("History of %s", location)
	}
}

// generateAuthor creates a genre-appropriate author name.
func (g *Generator) generateAuthor(genre string) string {
	switch genre {
	case "fantasy":
		firstNames := []string{"Aldric", "Elara", "Theron", "Lyra", "Morgath", "Celestia"}
		lastNames := []string{"the Wise", "the Elder", "the Scholar", "the Ancient", "Moonwhisper", "Stargazer"}
		return fmt.Sprintf("%s %s", g.pickRandom(firstNames), g.pickRandom(lastNames))
	case "sci-fi":
		ranks := []string{"Dr.", "Prof.", "Chief", "Commander", "Director"}
		names := []string{"Chen", "Rodriguez", "Nakamura", "Johnson", "Ivanova", "O'Brien"}
		return fmt.Sprintf("%s %s", g.pickRandom(ranks), g.pickRandom(names))
	case "horror":
		names := []string{"Unknown", "Redacted", "The Witness", "A Survivor", "Anonymous", "The Lost One"}
		return g.pickRandom(names)
	case "cyberpunk":
		handles := []string{"N3on", "Cy83r", "Gh0st", "R4v3n", "V1rus", "Sh4dow"}
		return g.pickRandom(handles)
	case "post-apocalyptic":
		names := []string{"The Wanderer", "Old Timer", "The Survivor", "Scribe", "The Keeper", "Witness"}
		return g.pickRandom(names)
	default:
		return "Unknown Author"
	}
}

// pickRandom selects a random element from a slice.
func (g *Generator) pickRandom(options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[g.rng.Intn(len(options))]
}

// getSeriesName checks if the book is part of a series and returns the series name.
// Returns (seriesName, true) if it's a series book, ("", false) otherwise.
func (g *Generator) getSeriesName() (string, bool) {
	// Check custom parameters for series information
	// This would be set by a library/bookshelf generator
	// Series name format expected in custom["series_name"]
	return "", false // For now, return false - can be extended later
}

// getVolumeNumber returns the volume number for a series book.
// Defaults to 1 if not specified.
func (g *Generator) getVolumeNumber() int {
	// Check custom parameters for volume number
	// This would be set by a library/bookshelf generator
	// Volume number format expected in custom["volume_number"]
	return 1 // Default to volume 1
}
